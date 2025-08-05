package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/maylng/backend/internal/email"
	"github.com/maylng/backend/internal/models"
)

type EmailService struct {
	db           *sql.DB
	emailService *email.Service
}

func NewEmailService(db *sql.DB, emailService *email.Service) *EmailService {
	return &EmailService{
		db:           db,
		emailService: emailService,
	}
}

func (s *EmailService) SendEmail(accountID uuid.UUID, req *models.SendEmailRequest) (*models.EmailResponse, error) {
	// Validate that the from_email_id belongs to the account
	var fromEmailAddress string
	var customDomainID *uuid.UUID
	err := s.db.QueryRow(
		"SELECT email, custom_domain_id FROM email_addresses WHERE id = $1 AND account_id = $2 AND status = 'active'",
		req.FromEmailID, accountID,
	).Scan(&fromEmailAddress, &customDomainID)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("from email address not found or not active")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to validate from email address: %w", err)
	}

	// If the email address uses a custom domain, verify that the domain is verified
	if customDomainID != nil {
		var domainStatus string
		err = s.db.QueryRow(
			"SELECT status FROM custom_domains WHERE id = $1 AND account_id = $2",
			customDomainID, accountID,
		).Scan(&domainStatus)

		if err != nil {
			return nil, fmt.Errorf("failed to validate custom domain: %w", err)
		}

		if domainStatus != string(models.CustomDomainStatusVerified) {
			return nil, fmt.Errorf("custom domain must be verified before sending emails (current status: %s)", domainStatus)
		}
	}

	// Check rate limits (simplified for MVP)
	// TODO: Implement proper rate limiting with Redis

	// Convert recipients to JSON
	toRecipientsJSON, _ := json.Marshal(req.ToRecipients)
	ccRecipientsJSON, _ := json.Marshal(req.CcRecipients)
	bccRecipientsJSON, _ := json.Marshal(req.BccRecipients)

	// Insert email record
	query := `
		INSERT INTO sent_emails (
			account_id, from_email_id, to_recipients, cc_recipients, bcc_recipients,
			subject, text_content, html_content, attachments, headers, thread_id,
			scheduled_at, status, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`

	var sentEmail models.SentEmail
	var status models.EmailStatus = models.EmailStatusQueued
	if req.ScheduledAt != nil && req.ScheduledAt.After(time.Now()) {
		status = models.EmailStatusScheduled
	}

	err = s.db.QueryRow(
		query,
		accountID,
		req.FromEmailID,
		toRecipientsJSON,
		ccRecipientsJSON,
		bccRecipientsJSON,
		req.Subject,
		req.TextContent,
		req.HTMLContent,
		req.Attachments,
		req.Headers,
		req.ThreadID,
		req.ScheduledAt,
		status,
		req.Metadata,
	).Scan(&sentEmail.ID, &sentEmail.CreatedAt, &sentEmail.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create email record: %w", err)
	}

	// If not scheduled, send immediately
	if status == models.EmailStatusQueued {
		go s.sendEmailAsync(sentEmail.ID, fromEmailAddress, req)
	}

	return &models.EmailResponse{
		ID:            sentEmail.ID,
		FromEmailID:   req.FromEmailID,
		ToRecipients:  req.ToRecipients,
		CcRecipients:  req.CcRecipients,
		BccRecipients: req.BccRecipients,
		Subject:       req.Subject,
		TextContent:   req.TextContent,
		HTMLContent:   req.HTMLContent,
		ThreadID:      req.ThreadID,
		ScheduledAt:   req.ScheduledAt,
		Status:        status,
		CreatedAt:     sentEmail.CreatedAt,
		UpdatedAt:     sentEmail.UpdatedAt,
	}, nil
}

func (s *EmailService) GetEmails(accountID uuid.UUID, limit, offset int) ([]*models.EmailResponse, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, from_email_id, to_recipients, cc_recipients, bcc_recipients,
			   subject, text_content, html_content, thread_id, scheduled_at, sent_at,
			   status, provider_message_id, failure_reason, created_at, updated_at
		FROM sent_emails 
		WHERE account_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get emails: %w", err)
	}
	defer rows.Close()

	var emails []*models.EmailResponse
	for rows.Next() {
		var email models.SentEmail
		var toRecipientsJSON, ccRecipientsJSON, bccRecipientsJSON []byte

		err := rows.Scan(
			&email.ID,
			&email.FromEmailID,
			&toRecipientsJSON,
			&ccRecipientsJSON,
			&bccRecipientsJSON,
			&email.Subject,
			&email.TextContent,
			&email.HTMLContent,
			&email.ThreadID,
			&email.ScheduledAt,
			&email.SentAt,
			&email.Status,
			&email.ProviderMessageID,
			&email.FailureReason,
			&email.CreatedAt,
			&email.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email: %w", err)
		}

		// Parse JSON recipients
		json.Unmarshal(toRecipientsJSON, &email.ToRecipients)
		json.Unmarshal(ccRecipientsJSON, &email.CcRecipients)
		json.Unmarshal(bccRecipientsJSON, &email.BccRecipients)

		emails = append(emails, &models.EmailResponse{
			ID:                email.ID,
			FromEmailID:       email.FromEmailID,
			ToRecipients:      email.ToRecipients,
			CcRecipients:      email.CcRecipients,
			BccRecipients:     email.BccRecipients,
			Subject:           email.Subject,
			TextContent:       email.TextContent,
			HTMLContent:       email.HTMLContent,
			ThreadID:          email.ThreadID,
			ScheduledAt:       email.ScheduledAt,
			SentAt:            email.SentAt,
			Status:            email.Status,
			ProviderMessageID: email.ProviderMessageID,
			FailureReason:     email.FailureReason,
			CreatedAt:         email.CreatedAt,
			UpdatedAt:         email.UpdatedAt,
		})
	}

	return emails, nil
}

func (s *EmailService) GetEmail(accountID, emailID uuid.UUID) (*models.EmailResponse, error) {
	query := `
		SELECT id, from_email_id, to_recipients, cc_recipients, bcc_recipients,
			   subject, text_content, html_content, thread_id, scheduled_at, sent_at,
			   status, provider_message_id, failure_reason, created_at, updated_at
		FROM sent_emails 
		WHERE id = $1 AND account_id = $2
	`

	var email models.SentEmail
	var toRecipientsJSON, ccRecipientsJSON, bccRecipientsJSON []byte

	err := s.db.QueryRow(query, emailID, accountID).Scan(
		&email.ID,
		&email.FromEmailID,
		&toRecipientsJSON,
		&ccRecipientsJSON,
		&bccRecipientsJSON,
		&email.Subject,
		&email.TextContent,
		&email.HTMLContent,
		&email.ThreadID,
		&email.ScheduledAt,
		&email.SentAt,
		&email.Status,
		&email.ProviderMessageID,
		&email.FailureReason,
		&email.CreatedAt,
		&email.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("email not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get email: %w", err)
	}

	// Parse JSON recipients
	json.Unmarshal(toRecipientsJSON, &email.ToRecipients)
	json.Unmarshal(ccRecipientsJSON, &email.CcRecipients)
	json.Unmarshal(bccRecipientsJSON, &email.BccRecipients)

	return &models.EmailResponse{
		ID:                email.ID,
		FromEmailID:       email.FromEmailID,
		ToRecipients:      email.ToRecipients,
		CcRecipients:      email.CcRecipients,
		BccRecipients:     email.BccRecipients,
		Subject:           email.Subject,
		TextContent:       email.TextContent,
		HTMLContent:       email.HTMLContent,
		ThreadID:          email.ThreadID,
		ScheduledAt:       email.ScheduledAt,
		SentAt:            email.SentAt,
		Status:            email.Status,
		ProviderMessageID: email.ProviderMessageID,
		FailureReason:     email.FailureReason,
		CreatedAt:         email.CreatedAt,
		UpdatedAt:         email.UpdatedAt,
	}, nil
}

func (s *EmailService) sendEmailAsync(emailID uuid.UUID, fromEmailAddress string, req *models.SendEmailRequest) {
	// Convert to email format
	emailToSend := &email.Email{
		FromEmail:     fromEmailAddress,
		ToRecipients:  req.ToRecipients,
		CcRecipients:  req.CcRecipients,
		BccRecipients: req.BccRecipients,
		Subject:       req.Subject,
		Headers:       make(map[string]string),
	}

	if req.TextContent != nil {
		emailToSend.TextContent = *req.TextContent
	}
	if req.HTMLContent != nil {
		emailToSend.HTMLContent = *req.HTMLContent
	}

	// Convert headers
	if req.Headers != nil {
		for key, value := range req.Headers {
			if strValue, ok := value.(string); ok {
				emailToSend.Headers[key] = strValue
			}
		}
	}

	// Send email
	result, err := s.emailService.SendEmail(emailToSend)

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
	_, updateErr := s.db.Exec(
		"UPDATE sent_emails SET status = $1, provider_message_id = $2, failure_reason = $3, sent_at = $4 WHERE id = $5",
		status, providerMessageID, failureReason, sentAt, emailID,
	)
	if updateErr != nil {
		fmt.Printf("Failed to update email status: %v\n", updateErr)
	}
}
