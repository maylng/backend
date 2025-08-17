package middleware

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminMiddlewareDB returns middleware that checks the accounts table's is_admin flag.
func AdminMiddlewareDB(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountIDVal, exists := c.Get("account_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		accountID, ok := accountIDVal.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		var isAdmin bool
		query := `SELECT is_admin FROM accounts WHERE id = $1`
		err := db.QueryRow(query, accountID).Scan(&isAdmin)
		if err == sql.ErrNoRows || err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
