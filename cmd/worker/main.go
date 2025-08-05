package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/maylng/backend/internal/config"
	"github.com/maylng/backend/internal/database"
	"github.com/maylng/backend/internal/email"
	"github.com/maylng/backend/internal/email/providers"
	"github.com/maylng/backend/internal/models"
	"github.com/maylng/backend/internal/services"
	"github.com/redis/go-redis/v9"
)

type Worker struct {
	db                     *sql.DB
	redisClient            *redis.Client
	emailService           *email.Service
	emailSvc               *services.EmailService
	customDomainService    *services.CustomDomainService
	sesVerificationService *services.SESVerificationService
	dnsValidationService   *services.DNSValidationService
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database connections
	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	redisClient := database.NewRedisClient(cfg.RedisURL)
	defer redisClient.Close()

	// Initialize email service
	var emailService *email.Service

	switch cfg.EmailProvider {
	case "ses":
		if cfg.AWSRegion != "" {
			sesProvider, err := providers.NewSESProvider(cfg.AWSRegion)
			if err != nil {
				log.Printf("Failed to initialize SES provider: %v", err)
				emailService = email.NewService(nil, nil)
			} else {
				emailService = email.NewService(sesProvider, nil)
				log.Println("Using SES email provider")
			}
		} else {
			log.Println("Warning: SES provider selected but AWS_REGION not configured")
			emailService = email.NewService(nil, nil)
		}
	case "sendgrid":
		if cfg.SendGridAPIKey != "" {
			sendGridProvider := providers.NewSendGridProvider(cfg.SendGridAPIKey)
			emailService = email.NewService(sendGridProvider, nil)
			log.Println("Using SendGrid email provider")
		} else {
			log.Println("Warning: SendGrid provider selected but SENDGRID_API_KEY not configured")
			emailService = email.NewService(nil, nil)
		}
	default:
		log.Printf("Warning: Unknown email provider '%s', trying to initialize available providers", cfg.EmailProvider)
		var primary, fallback email.Provider

		// Try SES first
		if cfg.AWSRegion != "" {
			if sesProvider, err := providers.NewSESProvider(cfg.AWSRegion); err == nil {
				primary = sesProvider
				log.Println("Initialized SES as primary provider")
			}
		}

		// Try SendGrid as fallback
		if cfg.SendGridAPIKey != "" {
			sendGridProvider := providers.NewSendGridProvider(cfg.SendGridAPIKey)
			if primary == nil {
				primary = sendGridProvider
				log.Println("Initialized SendGrid as primary provider")
			} else {
				fallback = sendGridProvider
				log.Println("Initialized SendGrid as fallback provider")
			}
		}

		emailService = email.NewService(primary, fallback)
		if primary == nil {
			log.Println("Warning: No email providers configured")
		}
	}

	emailSvc := services.NewEmailService(db, emailService)

	// Initialize custom domain services
	customDomainService := services.NewCustomDomainService(db)
	dnsValidationService := services.NewDNSValidationService()

	// Initialize SES verification service if using SES
	var sesVerificationService *services.SESVerificationService
	if cfg.EmailProvider == "ses" && cfg.AWSRegion != "" {
		var err error
		sesVerificationService, err = services.NewSESVerificationService(cfg.AWSRegion, customDomainService)
		if err != nil {
			log.Printf("Warning: Failed to initialize SES verification service: %v", err)
		} else {
			log.Println("Initialized SES verification service for domain polling")
		}
	}

	// Initialize worker
	worker := &Worker{
		db:                     db,
		redisClient:            redisClient,
		emailService:           emailService,
		emailSvc:               emailSvc,
		customDomainService:    customDomainService,
		sesVerificationService: sesVerificationService,
		dnsValidationService:   dnsValidationService,
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, shutting down gracefully...")
		cancel()
	}()

	log.Println("Starting email worker...")

	// Start worker loops
	go worker.processScheduledEmails(ctx)
	go worker.processQueuedEmails(ctx)
	go worker.cleanupExpiredEmails(ctx)
	go worker.processDomainVerification(ctx)

	// Wait for shutdown
	<-ctx.Done()
	log.Println("Worker shutdown complete")
}

func (w *Worker) processScheduledEmails(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processScheduledEmailsBatch()
		}
	}
}

func (w *Worker) processScheduledEmailsBatch() {
	// Get scheduled emails that are ready to send
	query := `
		SELECT id, account_id, from_email_id, to_recipients, cc_recipients, bcc_recipients,
			   subject, text_content, html_content, attachments, headers, thread_id, metadata
		FROM sent_emails 
		WHERE status = 'scheduled' AND scheduled_at <= $1
		LIMIT 100
	`

	rows, err := w.db.Query(query, time.Now())
	if err != nil {
		log.Printf("Failed to query scheduled emails: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var sentEmail models.SentEmail
		var toRecipientsJSON, ccRecipientsJSON, bccRecipientsJSON []byte

		err := rows.Scan(
			&sentEmail.ID,
			&sentEmail.AccountID,
			&sentEmail.FromEmailID,
			&toRecipientsJSON,
			&ccRecipientsJSON,
			&bccRecipientsJSON,
			&sentEmail.Subject,
			&sentEmail.TextContent,
			&sentEmail.HTMLContent,
			&sentEmail.Attachments,
			&sentEmail.Headers,
			&sentEmail.ThreadID,
			&sentEmail.Metadata,
		)
		if err != nil {
			log.Printf("Failed to scan scheduled email: %v", err)
			continue
		}

		// Parse JSON recipients
		json.Unmarshal(toRecipientsJSON, &sentEmail.ToRecipients)
		json.Unmarshal(ccRecipientsJSON, &sentEmail.CcRecipients)
		json.Unmarshal(bccRecipientsJSON, &sentEmail.BccRecipients)

		// Get from email address
		var fromEmailAddress string
		err = w.db.QueryRow(
			"SELECT email FROM email_addresses WHERE id = $1",
			sentEmail.FromEmailID,
		).Scan(&fromEmailAddress)

		if err != nil {
			log.Printf("Failed to get from email address for email %s: %v", sentEmail.ID, err)
			continue
		}

		// Send email
		w.sendScheduledEmail(&sentEmail, fromEmailAddress)
	}
}

func (w *Worker) sendScheduledEmail(sentEmail *models.SentEmail, fromEmailAddress string) {
	// Convert to email format
	emailToSend := &email.Email{
		FromEmail:     fromEmailAddress,
		ToRecipients:  sentEmail.ToRecipients,
		CcRecipients:  sentEmail.CcRecipients,
		BccRecipients: sentEmail.BccRecipients,
		Subject:       sentEmail.Subject,
		Headers:       make(map[string]string),
	}

	if sentEmail.TextContent != nil {
		emailToSend.TextContent = *sentEmail.TextContent
	}
	if sentEmail.HTMLContent != nil {
		emailToSend.HTMLContent = *sentEmail.HTMLContent
	}

	// Convert headers from metadata
	if sentEmail.Headers != nil {
		for key, value := range sentEmail.Headers {
			if strValue, ok := value.(string); ok {
				emailToSend.Headers[key] = strValue
			}
		}
	}

	// Send email
	result, err := w.emailService.SendEmail(emailToSend)

	// Update email status
	var status models.EmailStatus
	var providerMessageID *string
	var failureReason *string
	var sentAt *time.Time

	if err != nil {
		status = models.EmailStatusFailed
		errMsg := err.Error()
		failureReason = &errMsg
	} else {
		status = models.EmailStatusSent
		if result.MessageID != "" {
			providerMessageID = &result.MessageID
		}
		now := time.Now()
		sentAt = &now
	}

	// Update database
	_, updateErr := w.db.Exec(
		"UPDATE sent_emails SET status = $1, provider_message_id = $2, failure_reason = $3, sent_at = $4 WHERE id = $5",
		status, providerMessageID, failureReason, sentAt, sentEmail.ID,
	)
	if updateErr != nil {
		log.Printf("Failed to update email status for %s: %v", sentEmail.ID, updateErr)
	}
}

func (w *Worker) processQueuedEmails(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Check more frequently for queued emails
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processQueuedEmailsBatch()
		}
	}
}

func (w *Worker) processQueuedEmailsBatch() {
	// Get queued emails that are ready to send
	query := `
		SELECT id, account_id, from_email_id, to_recipients, cc_recipients, bcc_recipients,
			   subject, text_content, html_content, attachments, headers, thread_id, metadata
		FROM sent_emails 
		WHERE status = 'queued'
		LIMIT 100
	`

	rows, err := w.db.Query(query)
	if err != nil {
		log.Printf("Failed to query queued emails: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var sentEmail models.SentEmail
		var toRecipientsJSON, ccRecipientsJSON, bccRecipientsJSON []byte

		err := rows.Scan(
			&sentEmail.ID,
			&sentEmail.AccountID,
			&sentEmail.FromEmailID,
			&toRecipientsJSON,
			&ccRecipientsJSON,
			&bccRecipientsJSON,
			&sentEmail.Subject,
			&sentEmail.TextContent,
			&sentEmail.HTMLContent,
			&sentEmail.Attachments,
			&sentEmail.Headers,
			&sentEmail.ThreadID,
			&sentEmail.Metadata,
		)
		if err != nil {
			log.Printf("Failed to scan queued email: %v", err)
			continue
		}

		// Parse JSON recipients
		json.Unmarshal(toRecipientsJSON, &sentEmail.ToRecipients)
		json.Unmarshal(ccRecipientsJSON, &sentEmail.CcRecipients)
		json.Unmarshal(bccRecipientsJSON, &sentEmail.BccRecipients)

		// Get from email address
		var fromEmailAddress string
		err = w.db.QueryRow(
			"SELECT email FROM email_addresses WHERE id = $1",
			sentEmail.FromEmailID,
		).Scan(&fromEmailAddress)

		if err != nil {
			log.Printf("Failed to get from email address for email %s: %v", sentEmail.ID, err)
			continue
		}

		log.Printf("Processing queued email %s from %s", sentEmail.ID, fromEmailAddress)
		w.sendScheduledEmail(&sentEmail, fromEmailAddress)
	}
}

func (w *Worker) cleanupExpiredEmails(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.cleanupExpiredEmailsBatch()
		}
	}
}

func (w *Worker) cleanupExpiredEmailsBatch() {
	// Update expired email addresses
	_, err := w.db.Exec(`
		UPDATE email_addresses 
		SET status = 'expired' 
		WHERE type = 'temporary' 
		AND status = 'active' 
		AND expires_at IS NOT NULL 
		AND expires_at <= $1
	`, time.Now())

	if err != nil {
		log.Printf("Failed to update expired email addresses: %v", err)
	}

	// Clean up old rate limit entries
	_, err = w.db.Exec("DELETE FROM rate_limits WHERE expires_at <= $1", time.Now())
	if err != nil {
		log.Printf("Failed to clean up rate limits: %v", err)
	}
}

func (w *Worker) processDomainVerification(ctx context.Context) {
	// Check domain verification every 15 minutes
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	// Run once immediately on startup
	w.processDomainVerificationBatch()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processDomainVerificationBatch()
		}
	}
}

func (w *Worker) processDomainVerificationBatch() {
	// Only process if SES verification service is available
	if w.sesVerificationService == nil {
		return
	}

	log.Println("Starting domain verification batch process...")

	// Get all pending domains that haven't been checked in the last 10 minutes
	query := `
		SELECT id, account_id, domain, status, verification_attempted_at, created_at
		FROM custom_domains 
		WHERE status IN ('pending', 'failed') 
		AND (verification_attempted_at IS NULL OR verification_attempted_at <= $1)
		ORDER BY created_at ASC
		LIMIT 50
	`

	// Check domains that haven't been verified in the last 10 minutes
	cutoffTime := time.Now().Add(-10 * time.Minute)
	rows, err := w.db.Query(query, cutoffTime)
	if err != nil {
		log.Printf("Failed to query pending domains: %v", err)
		return
	}
	defer rows.Close()

	processedCount := 0
	for rows.Next() {
		var domainID, accountID uuid.UUID
		var domain, status string
		var verificationAttemptedAt, createdAt time.Time

		err := rows.Scan(&domainID, &accountID, &domain, &status, &verificationAttemptedAt, &createdAt)
		if err != nil {
			log.Printf("Failed to scan domain row: %v", err)
			continue
		}

		// Get full domain details
		customDomain, err := w.customDomainService.GetCustomDomainByID(domainID)
		if err != nil {
			log.Printf("Failed to get domain %s: %v", domain, err)
			continue
		}

		// Check DNS first if DNS validation service is available
		var dnsOK bool = true
		var dnsMessage string
		if w.dnsValidationService != nil {
			dnsMessage, err = w.dnsValidationService.GetDNSPropagationStatus(customDomain)
			if err != nil {
				log.Printf("DNS check failed for domain %s: %v", domain, err)
			}

			dnsStatus, _ := w.dnsValidationService.ValidateDomainDNS(customDomain)
			dnsOK = dnsStatus != nil && dnsStatus.AllRecordsPresent
		}

		// Only check SES if DNS is properly configured
		if dnsOK {
			// Check verification status with SES
			err = w.sesVerificationService.CheckVerificationStatus(customDomain)
			if err != nil {
				log.Printf("Failed to check SES verification for domain %s: %v", domain, err)
				continue
			}

			log.Printf("Checked domain %s: status=%s, ses_status=%s",
				domain, customDomain.Status, safeString(customDomain.SESVerificationStatus))
		} else {
			// Update domain with DNS status message
			if dnsMessage != "" {
				customDomain.FailureReason = &dnsMessage
				w.customDomainService.UpdateCustomDomain(customDomain)
			}
			log.Printf("Skipped SES check for domain %s: DNS not ready (%s)", domain, dnsMessage)
		}

		processedCount++

		// Add small delay between checks to avoid overwhelming SES API
		time.Sleep(1 * time.Second)
	}

	if processedCount > 0 {
		log.Printf("Completed domain verification batch: processed %d domains", processedCount)
	}
}

// Helper function to safely dereference string pointers
func safeString(s *string) string {
	if s == nil {
		return "nil"
	}
	return *s
}
