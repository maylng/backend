package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maylng/backend/internal/api/middleware"
	"github.com/maylng/backend/internal/models"
	"github.com/maylng/backend/internal/services"
)

type EmailHandler struct {
	emailService *services.EmailService
}

func NewEmailHandler(emailService *services.EmailService) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
	}
}

func (h *EmailHandler) SendEmail(c *gin.Context) {
	accountID, exists := middleware.GetAccountIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not found in context"})
		return
	}

	var req models.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that at least one content type is provided
	if (req.TextContent == nil || *req.TextContent == "") &&
		(req.HTMLContent == nil || *req.HTMLContent == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one of text_content or html_content must be provided"})
		return
	}

	email, err := h.emailService.SendEmail(accountID, &req)
	if err != nil {
		if err.Error() == "from email address not found or not active" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, email)
}

func (h *EmailHandler) GetEmails(c *gin.Context) {
	accountID, exists := middleware.GetAccountIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not found in context"})
		return
	}

	// Parse query parameters
	limit := 50
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	emails, err := h.emailService.GetEmails(accountID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"emails": emails,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *EmailHandler) GetEmail(c *gin.Context) {
	accountID, exists := middleware.GetAccountIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not found in context"})
		return
	}

	emailID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}

	email, err := h.emailService.GetEmail(accountID, emailID)
	if err != nil {
		if err.Error() == "email not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, email)
}

func (h *EmailHandler) GetEmailStatus(c *gin.Context) {
	accountID, exists := middleware.GetAccountIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not found in context"})
		return
	}

	emailID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}

	email, err := h.emailService.GetEmail(accountID, emailID)
	if err != nil {
		if err.Error() == "email not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                  email.ID,
		"status":              email.Status,
		"sent_at":             email.SentAt,
		"provider_message_id": email.ProviderMessageID,
		"failure_reason":      email.FailureReason,
	})
}
