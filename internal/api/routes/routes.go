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

	// Initialize SES verification service
	var sesVerificationService *services.SESVerificationService
	if cfg.AWSRegion != "" {
		var err error
		sesVerificationService, err = services.NewSESVerificationService(cfg.AWSRegion, customDomainService)
		if err != nil {
			// Log error but don't fail startup
			sesVerificationService = nil
		}
	}

	// Initialize Resend verification service
	var resendVerificationService *services.ResendVerificationService
	if cfg.ResendAPIKey != "" {
		var err error
		resendVerificationService, err = services.NewResendVerificationService(cfg.ResendAPIKey, cfg.AWSRegion, customDomainService)
		if err != nil {
			// Log error but don't fail startup
			resendVerificationService = nil
		}
	}

	// Determine default verification provider based on email provider configuration
	defaultVerificationProvider := "ses"
	if cfg.EmailProvider == "resend" && resendVerificationService != nil {
		defaultVerificationProvider = "resend"
	}

	// Initialize DNS validation service
	dnsValidationService := services.NewDNSValidationService()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	accountHandler := handlers.NewAccountHandler(accountService)
	emailAddressHandler := handlers.NewEmailAddressHandler(emailAddressService)
	emailHandler := handlers.NewEmailHandler(emailSvc)
	customDomainHandler := handlers.NewCustomDomainHandler(
		customDomainService,
		nil, // domain verification service - we'll implement later
		sesVerificationService,
		resendVerificationService,
		dnsValidationService,
		defaultVerificationProvider,
	)
	tpsHandler := handlers.NewTPSHandler(tpsService, emailAddressService, accountService)

	// Middleware
	router.Use(middleware.CORSMiddleware())

	// Health check routes (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/v1/health", healthHandler.HealthV1)

	// Public routes: open signups are disabled. Use admin or platform routes below.

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
		protected.POST("/tps", tpsHandler.CreateTPS)
		protected.GET("/tps", tpsHandler.ListTPSByEmail)
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
		protected.PATCH("/custom-domains/:id", customDomainHandler.UpdateCustomDomain)
		protected.DELETE("/custom-domains/:id", customDomainHandler.DeleteCustomDomain)
		protected.POST("/custom-domains/:id/verify", customDomainHandler.VerifyCustomDomain)
		protected.GET("/custom-domains/:id/status", customDomainHandler.CheckVerificationStatus)
		protected.GET("/custom-domains/:id/dns", customDomainHandler.ValidateDomainDNS)

		// Admin-only routes (also allow admins to create accounts)
		adminHandler := handlers.NewAdminHandler(accountService, emailAddressService)
		admin := protected.Group("/admin")
		admin.Use(middleware.AdminMiddlewareDB(db))
		{
			// Admins can create accounts via POST /v1/admin/users
			admin.POST("/users", accountHandler.CreateAccount)

			admin.GET("/users", adminHandler.ListUsers)
			admin.GET("/users/:id", adminHandler.GetUser)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
			admin.POST("/users/:id/revoke-key", adminHandler.RevokeKey)
			admin.GET("/users/:id/email-addresses", adminHandler.ListEmailAddresses)
			admin.GET("/stats", adminHandler.Stats)
		}

		// Platform-origin account creation route (requires X-Platform-Token header)
		if cfg.PlatformCreationToken != "" {
			platform := router.Group("/v1/platform")
			platform.Use(middleware.PlatformTokenMiddleware(cfg.PlatformCreationToken))
			{
				platform.POST("/accounts", accountHandler.CreateAccount)
			}
		}
	}

	// TODO: Webhook routes for email providers
	// webhooks := router.Group("/webhooks")
	// {
	//     webhooks.POST("/sendgrid", webhookHandler.SendGrid)
	//     webhooks.POST("/ses", webhookHandler.SES)
	//     webhooks.POST("/postmark", webhookHandler.Postmark)
	// }
}
