package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maylng/backend/internal/models"
	"github.com/maylng/backend/internal/services"
)

type CreateCustomDomainRequest struct {
	Domain string `json:"domain" binding:"required" validate:"required,fqdn"`
}

type CustomDomainResponse struct {
	ID                        uuid.UUID                 `json:"id"`
	Domain                    string                    `json:"domain"`
	Status                    string                    `json:"status"`
	DNSRecords                []models.DNSRecord        `json:"dns_records"`
	SESVerificationStatus     *string                   `json:"ses_verification_status,omitempty"`
	SESDKIMVerificationStatus *string                   `json:"ses_dkim_verification_status,omitempty"`
	FailureReason             *string                   `json:"failure_reason,omitempty"`
	VerificationAttemptedAt   *string                   `json:"verification_attempted_at,omitempty"`
	VerifiedAt                *string                   `json:"verified_at,omitempty"`
	CreatedAt                 string                    `json:"created_at"`
	UpdatedAt                 string                    `json:"updated_at"`
	DNSStatus                 *services.DomainDNSStatus `json:"dns_status,omitempty"`
}

type CustomDomainHandler struct {
	customDomainService    *services.CustomDomainService
	sesVerificationService *services.SESVerificationService
	dnsValidationService   *services.DNSValidationService
}

func NewCustomDomainHandler(customDomainService *services.CustomDomainService, sesVerificationService *services.SESVerificationService, dnsValidationService *services.DNSValidationService) *CustomDomainHandler {
	return &CustomDomainHandler{
		customDomainService:    customDomainService,
		sesVerificationService: sesVerificationService,
		dnsValidationService:   dnsValidationService,
	}
}

// CreateCustomDomain adds a new custom domain for the account
func (h *CustomDomainHandler) CreateCustomDomain(c *gin.Context) {
	accountID, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account ID not found"})
		return
	}

	var req CreateCustomDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Normalize domain (remove protocol, lowercase)
	domain := strings.ToLower(strings.TrimSpace(req.Domain))
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "www.")

	// Validate domain format
	if domain == "" || strings.Contains(domain, "/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain format"})
		return
	}

	// Create custom domain
	customDomain, err := h.customDomainService.CreateCustomDomain(accountID.(uuid.UUID), domain)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "Domain already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom domain"})
		return
	}

	// Initiate SES verification
	if h.sesVerificationService != nil {
		err = h.sesVerificationService.InitiateDomainVerification(customDomain)
		if err != nil {
			// Log error but don't fail the request
			// The domain is created, verification can be retried later
			errMsg := err.Error()
			customDomain.FailureReason = &errMsg
			h.customDomainService.UpdateCustomDomain(customDomain)
		}
	}

	// Return response
	response := h.toResponse(customDomain)
	c.JSON(http.StatusCreated, response)
}

// GetCustomDomains lists all custom domains for the account
func (h *CustomDomainHandler) GetCustomDomains(c *gin.Context) {
	accountID, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account ID not found"})
		return
	}

	domains, err := h.customDomainService.GetCustomDomainsByAccountID(accountID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get custom domains"})
		return
	}

	var responses []CustomDomainResponse
	for _, domain := range domains {
		responses = append(responses, h.toResponse(domain))
	}

	c.JSON(http.StatusOK, responses)
}

// GetCustomDomain gets a specific custom domain by ID
func (h *CustomDomainHandler) GetCustomDomain(c *gin.Context) {
	accountID, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account ID not found"})
		return
	}

	domainID := c.Param("id")
	id, err := uuid.Parse(domainID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	domain, err := h.customDomainService.GetCustomDomainByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Custom domain not found"})
		return
	}

	// Verify ownership
	if domain.AccountID != accountID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	response := h.toResponse(domain)
	c.JSON(http.StatusOK, response)
}

// DeleteCustomDomain deletes a custom domain
func (h *CustomDomainHandler) DeleteCustomDomain(c *gin.Context) {
	accountID, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account ID not found"})
		return
	}

	domainID := c.Param("id")
	id, err := uuid.Parse(domainID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	// Get domain to verify ownership and get domain name
	domain, err := h.customDomainService.GetCustomDomainByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Custom domain not found"})
		return
	}

	// Verify ownership
	if domain.AccountID != accountID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Delete from SES if verification service is available
	if h.sesVerificationService != nil {
		// Best effort - don't fail if SES deletion fails
		h.sesVerificationService.DeleteDomainIdentity(domain.Domain)
	}

	// Delete from database
	err = h.customDomainService.DeleteCustomDomain(id, accountID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete custom domain"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// VerifyCustomDomain triggers domain verification
func (h *CustomDomainHandler) VerifyCustomDomain(c *gin.Context) {
	accountID, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account ID not found"})
		return
	}

	domainID := c.Param("id")
	id, err := uuid.Parse(domainID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	domain, err := h.customDomainService.GetCustomDomainByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Custom domain not found"})
		return
	}

	// Verify ownership
	if domain.AccountID != accountID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Trigger verification
	if h.sesVerificationService != nil {
		err = h.sesVerificationService.InitiateDomainVerification(domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate verification: " + err.Error()})
			return
		}
	}

	// Return updated status
	response := h.toResponse(domain)
	c.JSON(http.StatusOK, response)
}

// CheckVerificationStatus checks the current verification status of a domain
func (h *CustomDomainHandler) CheckVerificationStatus(c *gin.Context) {
	accountID, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account ID not found"})
		return
	}

	domainID := c.Param("id")
	id, err := uuid.Parse(domainID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	domain, err := h.customDomainService.GetCustomDomainByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Custom domain not found"})
		return
	}

	// Verify ownership
	if domain.AccountID != accountID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Check status with SES
	if h.sesVerificationService != nil {
		err = h.sesVerificationService.CheckVerificationStatus(domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check verification status: " + err.Error()})
			return
		}
	}

	// Return status
	response := gin.H{
		"status":                       domain.Status,
		"ses_verification_status":      domain.SESVerificationStatus,
		"ses_dkim_verification_status": domain.SESDKIMVerificationStatus,
		"verified_at":                  domain.VerifiedAt,
		"failure_reason":               domain.FailureReason,
	}

	c.JSON(http.StatusOK, response)
}

// ValidateDomainDNS checks the DNS configuration for a custom domain
func (h *CustomDomainHandler) ValidateDomainDNS(c *gin.Context) {
	accountID, exists := c.Get("account_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account ID not found"})
		return
	}

	domainID := c.Param("id")
	id, err := uuid.Parse(domainID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	domain, err := h.customDomainService.GetCustomDomainByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Custom domain not found"})
		return
	}

	// Verify ownership
	if domain.AccountID != accountID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Validate DNS
	if h.dnsValidationService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "DNS validation service not available"})
		return
	}

	dnsStatus, err := h.dnsValidationService.ValidateDomainDNS(domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate DNS: " + err.Error()})
		return
	}

	// Return DNS validation results
	c.JSON(http.StatusOK, dnsStatus)
}

// Helper function to convert domain to response
func (h *CustomDomainHandler) toResponse(domain *models.CustomDomain) CustomDomainResponse {
	response := CustomDomainResponse{
		ID:                        domain.ID,
		Domain:                    domain.Domain,
		Status:                    string(domain.Status),
		DNSRecords:                domain.DNSRecords,
		SESVerificationStatus:     domain.SESVerificationStatus,
		SESDKIMVerificationStatus: domain.SESDKIMVerificationStatus,
		FailureReason:             domain.FailureReason,
		CreatedAt:                 domain.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:                 domain.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if domain.VerificationAttemptedAt != nil {
		formatted := domain.VerificationAttemptedAt.Format("2006-01-02T15:04:05Z")
		response.VerificationAttemptedAt = &formatted
	}

	if domain.VerifiedAt != nil {
		formatted := domain.VerifiedAt.Format("2006-01-02T15:04:05Z")
		response.VerifiedAt = &formatted
	}

	// Add DNS validation status if service is available
	if h.dnsValidationService != nil {
		if dnsStatus, err := h.dnsValidationService.ValidateDomainDNS(domain); err == nil {
			response.DNSStatus = dnsStatus
		}
	}

	return response
}
