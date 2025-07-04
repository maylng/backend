package middleware

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maylng/backend/internal/auth"
	"github.com/maylng/backend/internal/models"
)

func AuthMiddleware(db *sql.DB, salt string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		apiKey, err := auth.ExtractAPIKey(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Hash the API key to look up in database
		hashedKey := auth.HashAPIKey(apiKey, salt)

		// Look up account by hashed API key
		var account models.Account
		query := `
			SELECT id, api_key_hash, plan, email_limit_per_month, email_address_limit, created_at, updated_at
			FROM accounts WHERE api_key_hash = $1
		`
		err = db.QueryRow(query, hashedKey).Scan(
			&account.ID,
			&account.APIKeyHash,
			&account.Plan,
			&account.EmailLimitPerMonth,
			&account.EmailAddressLimit,
			&account.CreatedAt,
			&account.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			c.Abort()
			return
		}

		// Store account in context
		c.Set("account", account)
		c.Set("account_id", account.ID)
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func GetAccountFromContext(c *gin.Context) (*models.Account, bool) {
	account, exists := c.Get("account")
	if !exists {
		return nil, false
	}
	acc, ok := account.(models.Account)
	return &acc, ok
}

func GetAccountIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	accountID, exists := c.Get("account_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := accountID.(uuid.UUID)
	return id, ok
}
