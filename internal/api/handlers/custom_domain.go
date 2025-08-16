package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maylng/backend/internal/models"
	"github.com/maylng/backend/internal/services"
)

type CreateCustomDomainRequest struct {
	Domain               string `json:"domain" binding:"required" validate:"required,fqdn"`
	VerificationProvider string `json:"verification_provider,omitempty" validate:"omitempty,oneof=ses resend"`
}

type CustomDomainResponse struct {
	ID                         uuid.UUID                 `json:"id"`
	Domain                     string                    `json:"domain"`
	Status                     string                    `json:"status"`
	VerificationProvider       string                    `json:"verification_provider"`
	ProviderVerificationStatus *string                   `json:"provider_verification_status,omitempty"`
	DNSRecords                 []models.DNSRecord        `json:"dns_records"`
	SESVerificationStatus      *string                   `json:"ses_verification_status,omitempty"`
	SESDKIMVerificationStatus  *string                   `json:"ses_dkim_verification_status,omitempty"`
	FailureReason              *string                   `json:"failure_reason,omitempty"`
	VerificationAttemptedAt    *string                   `json:"verification_attempted_at,omitempty"`
	VerifiedAt                 *string                   `json:"verified_at,omitempty"`
	CreatedAt                  string                    `json:"created_at"`
	UpdatedAt                  string                    `json:"updated_at"`
	DNSStatus                  *services.DomainDNSStatus `json:"dns_status,omitempty"`
}

type CustomDomainHandler struct {
	customDomainService         *services.CustomDomainService
	domainVerificationService   *services.DomainVerificationService
	sesVerificationService      *services.SESVerificationService
	resendVerificationService   *services.ResendVerificationService
	dnsValidationService        *services.DNSValidationService
	defaultVerificationProvider string
}

func NewCustomDomainHandler(
	customDomainService *services.CustomDomainService,
	domainVerificationService *services.DomainVerificationService,
	sesVerificationService *services.SESVerificationService,
	resendVerificationService *services.ResendVerificationService,
	dnsValidationService *services.DNSValidationService,
	defaultVerificationProvider string,
) *CustomDomainHandler {
	if defaultVerificationProvider == "" {
		defaultVerificationProvider = "ses" // Default to SES for backward compatibility
	}

	return &CustomDomainHandler{
		customDomainService:         customDomainService,
		domainVerificationService:   domainVerificationService,
		sesVerificationService:      sesVerificationService,
		resendVerificationService:   resendVerificationService,
		dnsValidationService:        dnsValidationService,
		defaultVerificationProvider: defaultVerificationProvider,
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

	// Determine verification provider
	verificationProvider := req.VerificationProvider
	if verificationProvider == "" {
		verificationProvider = h.defaultVerificationProvider
	}

	// Validate verification provider
	if verificationProvider != "ses" && verificationProvider != "resend" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification provider. Must be 'ses' or 'resend'"})
		return
	}

	// Create custom domain
	customDomain, err := h.customDomainService.CreateCustomDomain(accountID.(uuid.UUID), domain, verificationProvider)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "Domain already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom domain"})
		return
	}

	// Get the appropriate verification service
	var verificationService services.DomainVerificationProvider
	switch verificationProvider {
	case "resend":
		if h.resendVerificationService != nil {
			verificationService = h.resendVerificationService
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Resend verification service not available"})
			return
		}
	case "ses":
		if h.sesVerificationService != nil {
			verificationService = h.sesVerificationService
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SES verification service not available"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification provider"})
		return
	}

	// Initiate verification using the selected provider
	if verificationService != nil {
		err = verificationService.InitiateDomainVerification(customDomain)
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

	// Delete from verification provider if verification service is available
	var verificationService services.DomainVerificationProvider
	switch domain.VerificationProvider {
	case "resend":
		verificationService = h.resendVerificationService
	case "ses":
		verificationService = h.sesVerificationService
	}

	if verificationService != nil {
		// Best effort - don't fail if provider deletion fails
		verificationService.DeleteDomainIdentity(domain.Domain)
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

	// Trigger verification using the appropriate provider
	var verificationService services.DomainVerificationProvider
	switch domain.VerificationProvider {
	case "resend":
		verificationService = h.resendVerificationService
	case "ses":
		verificationService = h.sesVerificationService
	}

	if verificationService != nil {
		err = verificationService.InitiateDomainVerification(domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate verification: " + err.Error()})
			return
		}
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": fmt.Sprintf("Verification service for provider '%s' not available", domain.VerificationProvider),
		})
		return
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

	// Check status using the appropriate provider
	var verificationService services.DomainVerificationProvider
	switch domain.VerificationProvider {
	case "resend":
		verificationService = h.resendVerificationService
	case "ses":
		verificationService = h.sesVerificationService
	}

	if verificationService != nil {
		err = verificationService.CheckVerificationStatus(domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check verification status: " + err.Error()})
			return
		}
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": fmt.Sprintf("Verification service for provider '%s' not available", domain.VerificationProvider),
		})
		return
	}

	// Return status
	response := gin.H{
		"status":                       domain.Status,
		"verification_provider":        domain.VerificationProvider,
		"provider_verification_status": domain.ProviderVerificationStatus,
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
		ID:                         domain.ID,
		Domain:                     domain.Domain,
		Status:                     string(domain.Status),
		VerificationProvider:       domain.VerificationProvider,
		ProviderVerificationStatus: domain.ProviderVerificationStatus,
		DNSRecords:                 domain.DNSRecords,
		SESVerificationStatus:      domain.SESVerificationStatus,
		SESDKIMVerificationStatus:  domain.SESDKIMVerificationStatus,
		FailureReason:              domain.FailureReason,
		CreatedAt:                  domain.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:                  domain.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
