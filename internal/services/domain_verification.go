package services

import (
	"github.com/maylng/backend/internal/models"
)

// DomainVerificationProvider defines the interface for domain verification services
type DomainVerificationProvider interface {
	// InitiateDomainVerification starts the domain verification process
	InitiateDomainVerification(customDomain *models.CustomDomain) error

	// CheckVerificationStatus checks the current verification status
	CheckVerificationStatus(customDomain *models.CustomDomain) error

	// DeleteDomainIdentity removes the domain from the provider
	DeleteDomainIdentity(domain string) error

	// GetProviderType returns the provider type (ses, resend, etc.)
	GetProviderType() string

	// RetryVerification retries the verification process
	RetryVerification(customDomain *models.CustomDomain) error
}

// DomainVerificationService is a wrapper that uses the appropriate provider
type DomainVerificationService struct {
	provider            DomainVerificationProvider
	customDomainService *CustomDomainService
}

// NewDomainVerificationService creates a new domain verification service
func NewDomainVerificationService(provider DomainVerificationProvider, customDomainService *CustomDomainService) *DomainVerificationService {
	return &DomainVerificationService{
		provider:            provider,
		customDomainService: customDomainService,
	}
}

// InitiateDomainVerification starts the domain verification process
func (s *DomainVerificationService) InitiateDomainVerification(customDomain *models.CustomDomain) error {
	// Set the provider type in metadata
	if customDomain.Metadata == nil {
		customDomain.Metadata = make(map[string]interface{})
	}
	customDomain.Metadata["verification_provider"] = s.provider.GetProviderType()

	return s.provider.InitiateDomainVerification(customDomain)
}

// CheckVerificationStatus checks the current verification status
func (s *DomainVerificationService) CheckVerificationStatus(customDomain *models.CustomDomain) error {
	return s.provider.CheckVerificationStatus(customDomain)
}

// DeleteDomainIdentity removes the domain from the provider
func (s *DomainVerificationService) DeleteDomainIdentity(domain string) error {
	return s.provider.DeleteDomainIdentity(domain)
}

// RetryVerification retries the verification process
func (s *DomainVerificationService) RetryVerification(customDomain *models.CustomDomain) error {
	return s.provider.RetryVerification(customDomain)
}

// GetProviderType returns the provider type
func (s *DomainVerificationService) GetProviderType() string {
	return s.provider.GetProviderType()
}
