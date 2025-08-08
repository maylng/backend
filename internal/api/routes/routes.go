package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/maylng/backend/internal/api/handlers"
	"github.com/maylng/backend/internal/api/middleware"
	"github.com/maylng/backend/internal/config"
	"github.com/maylng/backend/internal/email"
	"github.com/maylng/backend/internal/services"
	"github.com/redis/go-redis/v9"
)

func SetupRoutes(router *gin.Engine, cfg *config.Config, db *sql.DB, redisClient *redis.Client, emailService *email.Service) {
	// Initialize services
	accountService := services.NewAccountService(db, cfg.APIKeyHashSalt)
	emailAddressService := services.NewEmailAddressService(db, cfg)
	emailSvc := services.NewEmailService(db, emailService)
	customDomainService := services.NewCustomDomainService(db)
	tpsService := services.NewTPSService(db, cfg.TPSEncryptionKey)

	// Initialize SES verification service (only if using SES)
	var sesVerificationService *services.SESVerificationService
	if cfg.EmailProvider == "ses" && cfg.AWSRegion != "" {
		var err error
		sesVerificationService, err = services.NewSESVerificationService(cfg.AWSRegion, customDomainService)
		if err != nil {
			// Log error but don't fail startup
			// Custom domains will work without SES verification
			sesVerificationService = nil
		}
	}

	// Initialize DNS validation service
	dnsValidationService := services.NewDNSValidationService()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	accountHandler := handlers.NewAccountHandler(accountService)
	emailAddressHandler := handlers.NewEmailAddressHandler(emailAddressService)
	emailHandler := handlers.NewEmailHandler(emailSvc)
	customDomainHandler := handlers.NewCustomDomainHandler(customDomainService, sesVerificationService, dnsValidationService)
	tpsHandler := handlers.NewTPSHandler(tpsService, emailAddressService, accountService)

	// Middleware
	router.Use(middleware.CORSMiddleware())

	// Health check routes (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/v1/health", healthHandler.HealthV1)

	// Public routes
	public := router.Group("/v1")
	{
		public.POST("/accounts", accountHandler.CreateAccount)
	}

	// Protected routes
	protected := router.Group("/v1")
	protected.Use(middleware.AuthMiddleware(db, cfg.APIKeyHashSalt))
	{
		// Account management
		protected.GET("/account", accountHandler.GetAccount)
		protected.POST("/account/api-key", accountHandler.GenerateNewAPIKey)

		// Email address management
		protected.POST("/email-addresses", emailAddressHandler.CreateEmailAddress)
		protected.GET("/email-addresses", emailAddressHandler.GetEmailAddresses)
		protected.GET("/email-addresses/:id", emailAddressHandler.GetEmailAddress)
		protected.PATCH("/email-addresses/:id", emailAddressHandler.UpdateEmailAddress)
		protected.DELETE("/email-addresses/:id", emailAddressHandler.DeleteEmailAddress)

		// TPS (3rd Party Software) management
		protected.POST("/email-addresses/:email_id/tps", tpsHandler.CreateTPS)
		protected.GET("/email-addresses/:email_id/tps", tpsHandler.ListTPSByEmail)
		protected.GET("/tps/:tps_id", tpsHandler.GetTPS)
		protected.PATCH("/tps/:tps_id", tpsHandler.UpdateTPS)
		protected.DELETE("/tps/:tps_id", tpsHandler.DeleteTPS)

		// Email operations
		protected.POST("/emails/send", emailHandler.SendEmail)
		protected.GET("/emails", emailHandler.GetEmails)
		protected.GET("/emails/:id", emailHandler.GetEmail)
		protected.GET("/emails/:id/status", emailHandler.GetEmailStatus)

		// Custom domain management
		protected.POST("/custom-domains", customDomainHandler.CreateCustomDomain)
		protected.GET("/custom-domains", customDomainHandler.GetCustomDomains)
		protected.GET("/custom-domains/:id", customDomainHandler.GetCustomDomain)
		protected.DELETE("/custom-domains/:id", customDomainHandler.DeleteCustomDomain)
		protected.POST("/custom-domains/:id/verify", customDomainHandler.VerifyCustomDomain)
		protected.GET("/custom-domains/:id/status", customDomainHandler.CheckVerificationStatus)
		protected.GET("/custom-domains/:id/dns", customDomainHandler.ValidateDomainDNS)
	}

	// TODO: Webhook routes for email providers
	// webhooks := router.Group("/webhooks")
	// {
	//     webhooks.POST("/sendgrid", webhookHandler.SendGrid)
	//     webhooks.POST("/ses", webhookHandler.SES)
	//     webhooks.POST("/postmark", webhookHandler.Postmark)
	// }
}
