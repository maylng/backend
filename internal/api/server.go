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
	if cfg.SendGridAPIKey != "" {
		sendGridProvider := providers.NewSendGridProvider(cfg.SendGridAPIKey)
		emailService = email.NewService(sendGridProvider, nil)
	} else {
		// For development, create a mock service
		emailService = email.NewService(nil, nil)
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
