package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/maylng/backend/internal/api/routes"
	"github.com/maylng/backend/internal/config"
	"github.com/maylng/backend/internal/email"
	"github.com/maylng/backend/internal/email/providers"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	router *gin.Engine
	config *config.Config
}

func NewServer(cfg *config.Config, db *sql.DB, redisClient *redis.Client) *Server {
	router := gin.New()

	// Add built-in middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Initialize email service
	var emailService *email.Service

	switch cfg.EmailProvider {
	case "resend":
		// Use Resend as primary provider
		if cfg.ResendAPIKey != "" {
			resendProvider := providers.NewResendProvider(cfg.ResendAPIKey)
			var fallbackProvider email.Provider
			if cfg.SendGridAPIKey != "" {
				fallbackProvider = providers.NewSendGridProvider(cfg.SendGridAPIKey)
			}
			emailService = email.NewService(resendProvider, fallbackProvider)
		} else {
			// No Resend API key provided, fall back to other providers
			if cfg.SendGridAPIKey != "" {
				sendGridProvider := providers.NewSendGridProvider(cfg.SendGridAPIKey)
				emailService = email.NewService(sendGridProvider, nil)
			} else {
				emailService = email.NewService(nil, nil)
			}
		}
	case "ses":
		// Use SES as primary provider
		sesProvider, err := providers.NewSESProvider(cfg.AWSRegion)
		if err != nil {
			// If SES fails to initialize, fall back to other providers
			if cfg.ResendAPIKey != "" {
				resendProvider := providers.NewResendProvider(cfg.ResendAPIKey)
				emailService = email.NewService(resendProvider, nil)
			} else if cfg.SendGridAPIKey != "" {
				sendGridProvider := providers.NewSendGridProvider(cfg.SendGridAPIKey)
				emailService = email.NewService(sendGridProvider, nil)
			} else {
				emailService = email.NewService(nil, nil)
			}
		} else {
			var fallbackProvider email.Provider
			if cfg.ResendAPIKey != "" {
				fallbackProvider = providers.NewResendProvider(cfg.ResendAPIKey)
			} else if cfg.SendGridAPIKey != "" {
				fallbackProvider = providers.NewSendGridProvider(cfg.SendGridAPIKey)
			}
			emailService = email.NewService(sesProvider, fallbackProvider)
		}
	case "sendgrid":
		// Use SendGrid as primary provider
		if cfg.SendGridAPIKey != "" {
			sendGridProvider := providers.NewSendGridProvider(cfg.SendGridAPIKey)
			var fallbackProvider email.Provider
			if cfg.ResendAPIKey != "" {
				fallbackProvider = providers.NewResendProvider(cfg.ResendAPIKey)
			}
			emailService = email.NewService(sendGridProvider, fallbackProvider)
		} else {
			// No SendGrid API key provided, fall back to other providers
			if cfg.ResendAPIKey != "" {
				resendProvider := providers.NewResendProvider(cfg.ResendAPIKey)
				emailService = email.NewService(resendProvider, nil)
			} else {
				emailService = email.NewService(nil, nil)
			}
		}
	default:
		// Default case - prefer Resend, then SendGrid for backward compatibility
		if cfg.ResendAPIKey != "" {
			resendProvider := providers.NewResendProvider(cfg.ResendAPIKey)
			var fallbackProvider email.Provider
			if cfg.SendGridAPIKey != "" {
				fallbackProvider = providers.NewSendGridProvider(cfg.SendGridAPIKey)
			}
			emailService = email.NewService(resendProvider, fallbackProvider)
		} else if cfg.SendGridAPIKey != "" {
			sendGridProvider := providers.NewSendGridProvider(cfg.SendGridAPIKey)
			emailService = email.NewService(sendGridProvider, nil)
		} else {
			// For development, create a service with no providers
			emailService = email.NewService(nil, nil)
		}
	}

	// Setup routes
	routes.SetupRoutes(router, cfg, db, redisClient, emailService)

	return &Server{
		router: router,
		config: cfg,
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
