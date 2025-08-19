package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PlatformTokenMiddleware returns middleware that validates the X-Platform-Token header
// against the provided token. If token is empty, middleware rejects requests.
func PlatformTokenMiddleware(expectedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expectedToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "platform token not configured"})
			c.Abort()
			return
		}

		token := c.GetHeader("X-Platform-Token")
		if token == "" || token != expectedToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid platform token"})
			c.Abort()
			return
		}

		c.Next()
	}
}
