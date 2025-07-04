package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maylng/backend/internal/api/middleware"
	"github.com/maylng/backend/internal/models"
	"github.com/maylng/backend/internal/services"
)

type EmailAddressHandler struct {
	emailAddressService *services.EmailAddressService
}

func NewEmailAddressHandler(emailAddressService *services.EmailAddressService) *EmailAddressHandler {
	return &EmailAddressHandler{
		emailAddressService: emailAddressService,
	}
}

func (h *EmailAddressHandler) CreateEmailAddress(c *gin.Context) {
	accountID, exists := middleware.GetAccountIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not found in context"})
		return
	}

	var req models.CreateEmailAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emailAddress, err := h.emailAddressService.CreateEmailAddress(accountID, &req)
	if err != nil {
		if err.Error() == "email address limit reached" ||
			err.Error() == "email address already exists" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, emailAddress)
}

func (h *EmailAddressHandler) GetEmailAddresses(c *gin.Context) {
	accountID, exists := middleware.GetAccountIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not found in context"})
		return
	}

	emailAddresses, err := h.emailAddressService.GetEmailAddresses(accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"email_addresses": emailAddresses})
}

func (h *EmailAddressHandler) GetEmailAddress(c *gin.Context) {
	accountID, exists := middleware.GetAccountIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not found in context"})
		return
	}

	emailAddressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address ID"})
		return
	}

	emailAddress, err := h.emailAddressService.GetEmailAddress(accountID, emailAddressID)
	if err != nil {
		if err.Error() == "email address not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, emailAddress)
}

func (h *EmailAddressHandler) UpdateEmailAddress(c *gin.Context) {
	accountID, exists := middleware.GetAccountIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not found in context"})
		return
	}

	emailAddressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address ID"})
		return
	}

	var req models.UpdateEmailAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emailAddress, err := h.emailAddressService.UpdateEmailAddress(accountID, emailAddressID, &req)
	if err != nil {
		if err.Error() == "email address not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, emailAddress)
}

func (h *EmailAddressHandler) DeleteEmailAddress(c *gin.Context) {
	accountID, exists := middleware.GetAccountIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not found in context"})
		return
	}

	emailAddressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address ID"})
		return
	}

	err = h.emailAddressService.DeleteEmailAddress(accountID, emailAddressID)
	if err != nil {
		if err.Error() == "email address not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
