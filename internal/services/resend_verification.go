package services

import (
	"fmt"
	"time"

	"github.com/maylng/backend/internal/models"
	"github.com/resend/resend-go/v2"
)

type ResendVerificationService struct {
	client              *resend.Client
	region              string
	customDomainService *CustomDomainService
}

func NewResendVerificationService(apiKey, region string, customDomainService *CustomDomainService) (*ResendVerificationService, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("resend API key is required")
	}

	if region == "" {
		region = "us-east-1"
	}

	client := resend.NewClient(apiKey)

	return &ResendVerificationService{
		client:              client,
		region:              region,
		customDomainService: customDomainService,
	}, nil
}

// InitiateDomainVerification starts the Resend domain verification process
func (s *ResendVerificationService) InitiateDomainVerification(customDomain *models.CustomDomain) error {
	if s.client == nil {
		return fmt.Errorf("resend client not configured")
	}

	// 1. Create domain in Resend
	createParams := &resend.CreateDomainRequest{
		Name:   customDomain.Domain,
		Region: s.region,
	}

	createResp, err := s.client.Domains.Create(createParams)
	if err != nil {
		return fmt.Errorf("failed to create domain in Resend: %w", err)
	}

	// 2. Store Resend domain ID in both provider_domain_id field and metadata
	customDomain.ProviderDomainID = &createResp.Id
	if customDomain.Metadata == nil {
		customDomain.Metadata = make(map[string]interface{})
	}
	customDomain.Metadata["resend_domain_id"] = createResp.Id
	customDomain.Metadata["resend_region"] = s.region

	// 3. Extract DNS records from the response
	dnsRecords := []models.DNSRecord{}

	for _, record := range createResp.Records {
		var recordType string
		switch record.Type {
		case "TXT":
			recordType = "TXT"
		case "CNAME":
			recordType = "CNAME"
		case "MX":
			recordType = "MX"
		default:
			recordType = record.Type
		}

		dnsRecord := models.DNSRecord{
			Type:  recordType,
			Name:  record.Name,
			Value: record.Value,
			TTL:   3600, // Default TTL
		}

		// Handle MX record priority - Resend API may return it as string
		if record.Priority != "" {
			// Priority is a string in the SDK, try to parse it
			var priority int
			if _, err := fmt.Sscanf(string(record.Priority), "%d", &priority); err == nil {
				dnsRecord.Priority = priority
			}
		}

		dnsRecords = append(dnsRecords, dnsRecord)
	}

	// 4. Update custom domain with verification details
	customDomain.DNSRecords = dnsRecords
	customDomain.Status = models.CustomDomainStatusPending

	// Store Resend-specific verification status in both provider field and metadata
	resendStatus := string(createResp.Status)
	customDomain.ProviderVerificationStatus = &resendStatus
	if customDomain.Metadata == nil {
		customDomain.Metadata = make(map[string]interface{})
	}
	customDomain.Metadata["resend_verification_status"] = resendStatus

	// Mark verification as attempted
	now := time.Now()
	customDomain.VerificationAttemptedAt = &now

	// Save to database
	return s.customDomainService.UpdateCustomDomain(customDomain)
}

// CheckVerificationStatus checks the current verification status from Resend
func (s *ResendVerificationService) CheckVerificationStatus(customDomain *models.CustomDomain) error {
	if s.client == nil {
		return fmt.Errorf("resend client not configured")
	}

	// Get Resend domain ID from provider_domain_id field or fallback to metadata
	var resendDomainID string
	if customDomain.ProviderDomainID != nil {
		resendDomainID = *customDomain.ProviderDomainID
	} else if customDomain.Metadata != nil {
		if id, ok := customDomain.Metadata["resend_domain_id"].(string); ok {
			resendDomainID = id
		}
	}

	if resendDomainID == "" {
		return fmt.Errorf("resend domain ID not found in provider_domain_id or metadata")
	}

	// Get domain details from Resend
	domain, err := s.client.Domains.Get(resendDomainID)
	if err != nil {
		return fmt.Errorf("failed to get domain from Resend: %w", err)
	}

	// Update verification status
	previousStatus := customDomain.Status

	// Store Resend-specific verification status in both provider field and metadata
	resendStatus := string(domain.Status)
	customDomain.ProviderVerificationStatus = &resendStatus
	if customDomain.Metadata == nil {
		customDomain.Metadata = make(map[string]interface{})
	}
	customDomain.Metadata["resend_verification_status"] = resendStatus

	// Map Resend status to our internal status
	switch string(domain.Status) {
	case "verified":
		customDomain.Status = models.CustomDomainStatusVerified
		if customDomain.VerifiedAt == nil {
			now := time.Now()
			customDomain.VerifiedAt = &now
		}
	case "failed":
		customDomain.Status = models.CustomDomainStatusFailed
		customDomain.FailureReason = stringPtr("Domain verification failed in Resend")
	case "pending", "not_started", "temporary_failure":
		customDomain.Status = models.CustomDomainStatusPending
	default:
		customDomain.Status = models.CustomDomainStatusPending
	}

	// Log status change
	if previousStatus != customDomain.Status {
		fmt.Printf("Domain %s status changed from %s to %s (Resend status: %s)\n",
			customDomain.Domain, previousStatus, customDomain.Status, domain.Status)
	}

	// Save to database
	return s.customDomainService.UpdateCustomDomain(customDomain)
}

// DeleteDomainIdentity removes the domain from Resend using domain ID
func (s *ResendVerificationService) DeleteDomainIdentity(domain string) error {
	if s.client == nil {
		return fmt.Errorf("resend client not configured")
	}

	// First, find the domain ID by listing all domains
	domains, err := s.client.Domains.List()
	if err != nil {
		return fmt.Errorf("failed to list domains from Resend: %w", err)
	}

	var domainID string
	for _, d := range domains.Data {
		if d.Name == domain {
			domainID = d.Id
			break
		}
	}

	if domainID == "" {
		// Domain not found, which is fine for deletion
		return nil
	}

	// Delete the domain
	_, err = s.client.Domains.Remove(domainID)
	if err != nil {
		return fmt.Errorf("failed to delete domain from Resend: %w", err)
	}

	return nil
}

// DeleteDomainByID removes the domain from Resend using the stored provider domain ID
func (s *ResendVerificationService) DeleteDomainByID(customDomain *models.CustomDomain) error {
	if s.client == nil {
		return fmt.Errorf("resend client not configured")
	}

	// Get Resend domain ID from provider_domain_id field or fallback to metadata
	var resendDomainID string
	if customDomain.ProviderDomainID != nil {
		resendDomainID = *customDomain.ProviderDomainID
	} else if customDomain.Metadata != nil {
		if id, ok := customDomain.Metadata["resend_domain_id"].(string); ok {
			resendDomainID = id
		}
	}

	if resendDomainID == "" {
		// No domain ID found, try fallback to domain name
		return s.DeleteDomainIdentity(customDomain.Domain)
	}

	// Delete the domain using the stored ID
	_, err := s.client.Domains.Remove(resendDomainID)
	if err != nil {
		return fmt.Errorf("failed to delete domain from Resend: %w", err)
	}

	return nil
}

// UpdateDomainSettings updates domain configuration in Resend (tracking, TLS, etc.)
func (s *ResendVerificationService) UpdateDomainSettings(customDomain *models.CustomDomain, settings map[string]interface{}) error {
	if s.client == nil {
		return fmt.Errorf("resend client not configured")
	}

	// Get Resend domain ID from provider_domain_id field or fallback to metadata
	var resendDomainID string
	if customDomain.ProviderDomainID != nil {
		resendDomainID = *customDomain.ProviderDomainID
	} else if customDomain.Metadata != nil {
		if id, ok := customDomain.Metadata["resend_domain_id"].(string); ok {
			resendDomainID = id
		}
	}

	if resendDomainID == "" {
		return fmt.Errorf("resend domain ID not found")
	}

	// Build update request
	updateParams := &resend.UpdateDomainRequest{}

	// Handle click tracking
	if clickTracking, ok := settings["click_tracking"].(bool); ok {
		updateParams.ClickTracking = clickTracking
	}

	// Handle open tracking
	if openTracking, ok := settings["open_tracking"].(bool); ok {
		updateParams.OpenTracking = openTracking
	}

	// Handle TLS setting
	if tlsStr, ok := settings["tls"].(string); ok {
		switch tlsStr {
		case "enforced":
			updateParams.Tls = resend.Enforced
		case "opportunistic":
			updateParams.Tls = resend.Opportunistic
		}
	}

	// Update domain in Resend
	_, err := s.client.Domains.Update(resendDomainID, updateParams)
	if err != nil {
		return fmt.Errorf("failed to update domain in Resend: %w", err)
	}

	// Store updated settings in metadata
	if customDomain.Metadata == nil {
		customDomain.Metadata = make(map[string]interface{})
	}
	for key, value := range settings {
		customDomain.Metadata["resend_"+key] = value
	}

	// Save to database
	return s.customDomainService.UpdateCustomDomain(customDomain)
}

// RetryVerification retries the verification process for a domain
func (s *ResendVerificationService) RetryVerification(customDomain *models.CustomDomain) error {
	// Reset some fields
	customDomain.Status = models.CustomDomainStatusPending
	customDomain.FailureReason = nil
	customDomain.VerifiedAt = nil

	// Get Resend domain ID from provider_domain_id field or fallback to metadata
	var resendDomainID string
	if customDomain.ProviderDomainID != nil {
		resendDomainID = *customDomain.ProviderDomainID
	} else if customDomain.Metadata != nil {
		if id, ok := customDomain.Metadata["resend_domain_id"].(string); ok {
			resendDomainID = id
		}
	}

	if resendDomainID == "" {
		// If no domain ID exists, re-initiate the whole process
		return s.InitiateDomainVerification(customDomain)
	}

	// Trigger verification in Resend
	_, err := s.client.Domains.Verify(resendDomainID)
	if err != nil {
		return fmt.Errorf("failed to trigger verification in Resend: %w", err)
	}

	// Update verification attempted timestamp
	now := time.Now()
	customDomain.VerificationAttemptedAt = &now

	// Save to database
	return s.customDomainService.UpdateCustomDomain(customDomain)
}

// GetProviderType returns the provider type
func (s *ResendVerificationService) GetProviderType() string {
	return "resend"
}

// Helper function to convert string to string pointer
func stringPtr(s string) *string {
	return &s
}
