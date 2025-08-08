package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maylng/backend/internal/models"
	"github.com/maylng/backend/internal/services"
)

type TPSHandler struct {
	tpsService          *services.TPSService
	emailAddressService *services.EmailAddressService
	accountService      *services.AccountService
}

func NewTPSHandler(tpsService *services.TPSService, emailAddressService *services.EmailAddressService, accountService *services.AccountService) *TPSHandler {
	return &TPSHandler{
		tpsService:          tpsService,
		emailAddressService: emailAddressService,
		accountService:      accountService,
	}
}

// CreateTPS creates a new TPS integration for an agent email
func (h *TPSHandler) CreateTPS(c *gin.Context) {
	// Extract account ID from authentication middleware
	accountID, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account ID not found"})
		return
	}

	// Parse email address ID from URL
	emailAddressIDStr := c.Param("email_id")
	emailAddressID, err := uuid.Parse(emailAddressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address ID"})
		return
	}

	// Bind request body
	var req models.CreateTPSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the email address ID from URL
	req.EmailAddressID = emailAddressID

	// Verify email address belongs to the authenticated account and is an agent
	emailAddress, err := h.emailAddressService.GetEmailAddress(accountID.(uuid.UUID), emailAddressID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Email address not found"})
		return
	}

	if emailAddress.AccessType != "agent" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "TPS can only be created for agent email addresses"})
		return
	}

	// Get account plan for limit enforcement
	account, err := h.accountService.GetAccount(accountID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get account information"})
		return
	}

	// Create TPS
	tps, err := h.tpsService.CreateTPS(account.Plan, &req)
	if err != nil {
		// Check if it's a plan limit error
		if err.Error() == "TPS limit reached for this agent email address" {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tps)
}

// GetTPS retrieves a specific TPS integration
func (h *TPSHandler) GetTPS(c *gin.Context) {
	// Extract account ID from authentication middleware
	accountID, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account ID not found"})
		return
	}

	// Parse TPS ID from URL
	tpsIDStr := c.Param("tps_id")
	tpsID, err := uuid.Parse(tpsIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid TPS ID"})
		return
	}

	// Get TPS
	tps, err := h.tpsService.GetTPS(tpsID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "TPS not found"})
		return
	}

	// Verify TPS belongs to the authenticated account
	_, err = h.emailAddressService.GetEmailAddress(accountID.(uuid.UUID), tps.EmailAddressID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "TPS does not belong to your account"})
		return
	}

	c.JSON(http.StatusOK, tps)
}

// ListTPSByEmail lists all TPS integrations for a specific agent email
func (h *TPSHandler) ListTPSByEmail(c *gin.Context) {
	// Extract account ID from authentication middleware
	accountID, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account ID not found"})
		return
	}

	// Parse email address ID from URL
	emailAddressIDStr := c.Param("email_id")
	emailAddressID, err := uuid.Parse(emailAddressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address ID"})
		return
	}

	// Verify email address belongs to the authenticated account
	_, err = h.emailAddressService.GetEmailAddress(accountID.(uuid.UUID), emailAddressID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Email address does not belong to your account"})
		return
	}

	// List TPS integrations
	tpsList, err := h.tpsService.ListTPSByEmailAddress(emailAddressID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tpsList})
}

// UpdateTPS updates an existing TPS integration
func (h *TPSHandler) UpdateTPS(c *gin.Context) {
	// Extract account ID from authentication middleware
	accountID, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account ID not found"})
		return
	}

	// Parse TPS ID from URL
	tpsIDStr := c.Param("tps_id")
	tpsID, err := uuid.Parse(tpsIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid TPS ID"})
		return
	}

	// Bind request body
	var req models.UpdateTPSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify TPS belongs to the authenticated account
	existingTPS, err := h.tpsService.GetTPS(tpsID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "TPS not found"})
		return
	}

	_, err = h.emailAddressService.GetEmailAddress(accountID.(uuid.UUID), existingTPS.EmailAddressID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "TPS does not belong to your account"})
		return
	}

	// Update TPS
	updatedTPS, err := h.tpsService.UpdateTPS(tpsID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedTPS)
}

// DeleteTPS deletes a TPS integration
func (h *TPSHandler) DeleteTPS(c *gin.Context) {
	// Extract account ID from authentication middleware
	accountID, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account ID not found"})
		return
	}

	// Parse TPS ID from URL
	tpsIDStr := c.Param("tps_id")
	tpsID, err := uuid.Parse(tpsIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid TPS ID"})
		return
	}

	// Verify TPS belongs to the authenticated account
	existingTPS, err := h.tpsService.GetTPS(tpsID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "TPS not found"})
		return
	}

	_, err = h.emailAddressService.GetEmailAddress(accountID.(uuid.UUID), existingTPS.EmailAddressID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "TPS does not belong to your account"})
		return
	}

	// Delete TPS
	if err := h.tpsService.DeleteTPS(tpsID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
