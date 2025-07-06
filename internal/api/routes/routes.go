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

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	accountHandler := handlers.NewAccountHandler(accountService)
	emailAddressHandler := handlers.NewEmailAddressHandler(emailAddressService)
	emailHandler := handlers.NewEmailHandler(emailSvc)

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

		// Email address management
		protected.POST("/email-addresses", emailAddressHandler.CreateEmailAddress)
		protected.GET("/email-addresses", emailAddressHandler.GetEmailAddresses)
		protected.GET("/email-addresses/:id", emailAddressHandler.GetEmailAddress)
		protected.PATCH("/email-addresses/:id", emailAddressHandler.UpdateEmailAddress)
		protected.DELETE("/email-addresses/:id", emailAddressHandler.DeleteEmailAddress)

		// Email operations
		protected.POST("/emails/send", emailHandler.SendEmail)
		protected.GET("/emails", emailHandler.GetEmails)
		protected.GET("/emails/:id", emailHandler.GetEmail)
		protected.GET("/emails/:id/status", emailHandler.GetEmailStatus)
	}

	// TODO: Webhook routes for email providers
	// webhooks := router.Group("/webhooks")
	// {
	//     webhooks.POST("/sendgrid", webhookHandler.SendGrid)
	//     webhooks.POST("/ses", webhookHandler.SES)
	//     webhooks.POST("/postmark", webhookHandler.Postmark)
	// }
}
