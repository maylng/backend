package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/maylng/backend/internal/models"
)

type SESVerificationService struct {
	client              *sesv2.Client
	region              string
	customDomainService *CustomDomainService
}

func NewSESVerificationService(region string, customDomainService *CustomDomainService) (*SESVerificationService, error) {
	if region == "" {
		region = "us-east-1"
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sesv2.NewFromConfig(cfg)

	return &SESVerificationService{
		client:              client,
		region:              region,
		customDomainService: customDomainService,
	}, nil
}

// InitiateDomainVerification starts the SES domain verification process
func (s *SESVerificationService) InitiateDomainVerification(customDomain *models.CustomDomain) error {
	ctx := context.TODO()

	// 1. Create email identity for the domain
	createInput := &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(customDomain.Domain),
		DkimSigningAttributes: &types.DkimSigningAttributes{
			NextSigningKeyLength: types.DkimSigningKeyLengthRsa1024Bit,
		},
	}

	createResp, err := s.client.CreateEmailIdentity(ctx, createInput)
	if err != nil {
		// Check if domain already exists
		if err.Error() != "AlreadyExistsException" {
			return fmt.Errorf("failed to create email identity: %w", err)
		}
	}

	// 2. Get verification details
	getInput := &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(customDomain.Domain),
	}

	getResp, err := s.client.GetEmailIdentity(ctx, getInput)
	if err != nil {
		return fmt.Errorf("failed to get verification details: %w", err)
	}

	// 3. Update custom domain with verification details
	dnsRecords := []models.DNSRecord{}

	// Add DKIM CNAME records if available
	if createResp != nil && createResp.DkimAttributes != nil && createResp.DkimAttributes.Tokens != nil {
		for i, token := range createResp.DkimAttributes.Tokens {
			dnsRecords = append(dnsRecords, models.DNSRecord{
				Type:  "CNAME",
				Name:  fmt.Sprintf("%s._domainkey.%s", token, customDomain.Domain),
				Value: fmt.Sprintf("%s.dkim.amazonses.com", token),
				TTL:   1800,
			})

			// Store DKIM tokens
			if customDomain.DKIMTokens == nil {
				customDomain.DKIMTokens = make(map[string]interface{})
			}
			customDomain.DKIMTokens[fmt.Sprintf("token_%d", i)] = token
		}
	} else if getResp.DkimAttributes != nil && getResp.DkimAttributes.Tokens != nil {
		// Use tokens from get response if create response doesn't have them
		for i, token := range getResp.DkimAttributes.Tokens {
			dnsRecords = append(dnsRecords, models.DNSRecord{
				Type:  "CNAME",
				Name:  fmt.Sprintf("%s._domainkey.%s", token, customDomain.Domain),
				Value: fmt.Sprintf("%s.dkim.amazonses.com", token),
				TTL:   1800,
			})

			// Store DKIM tokens
			if customDomain.DKIMTokens == nil {
				customDomain.DKIMTokens = make(map[string]interface{})
			}
			customDomain.DKIMTokens[fmt.Sprintf("token_%d", i)] = token
		}
	}

	// Update domain with verification details
	customDomain.DNSRecords = dnsRecords
	customDomain.SESVerificationStatus = aws.String(string(getResp.VerificationStatus))

	if getResp.DkimAttributes != nil {
		customDomain.SESDKIMVerificationStatus = aws.String(string(getResp.DkimAttributes.Status))
	}

	// Mark verification as attempted
	now := time.Now()
	customDomain.VerificationAttemptedAt = &now

	// Save to database
	return s.customDomainService.UpdateCustomDomain(customDomain)
}

// CheckVerificationStatus checks the current verification status from SES
func (s *SESVerificationService) CheckVerificationStatus(customDomain *models.CustomDomain) error {
	ctx := context.TODO()

	input := &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(customDomain.Domain),
	}

	resp, err := s.client.GetEmailIdentity(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to check verification status: %w", err)
	}

	// Update verification status
	previousStatus := customDomain.Status
	customDomain.SESVerificationStatus = aws.String(string(resp.VerificationStatus))

	if resp.DkimAttributes != nil {
		customDomain.SESDKIMVerificationStatus = aws.String(string(resp.DkimAttributes.Status))
	}

	// Update domain status based on SES status
	if resp.VerificationStatus == types.VerificationStatusSuccess {
		if resp.DkimAttributes != nil && resp.DkimAttributes.Status == types.DkimStatusSuccess {
			customDomain.Status = models.CustomDomainStatusVerified
			if customDomain.VerifiedAt == nil {
				now := time.Now()
				customDomain.VerifiedAt = &now
			}
		} else {
			// Domain verified but DKIM pending
			customDomain.Status = models.CustomDomainStatusPending
		}
	} else if resp.VerificationStatus == types.VerificationStatusFailed {
		customDomain.Status = models.CustomDomainStatusFailed
		customDomain.FailureReason = aws.String("Domain verification failed in SES")
	} else {
		customDomain.Status = models.CustomDomainStatusPending
	}

	// Log status change
	if previousStatus != customDomain.Status {
		fmt.Printf("Domain %s status changed from %s to %s\n",
			customDomain.Domain, previousStatus, customDomain.Status)
	}

	// Save to database
	return s.customDomainService.UpdateCustomDomain(customDomain)
}

// DeleteDomainIdentity removes the domain from SES
func (s *SESVerificationService) DeleteDomainIdentity(domain string) error {
	ctx := context.TODO()

	input := &sesv2.DeleteEmailIdentityInput{
		EmailIdentity: aws.String(domain),
	}

	_, err := s.client.DeleteEmailIdentity(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete domain identity from SES: %w", err)
	}

	return nil
}

// GetDomainDKIMAttributes gets DKIM attributes for a domain
func (s *SESVerificationService) GetDomainDKIMAttributes(domain string) (*types.DkimAttributes, error) {
	ctx := context.TODO()

	input := &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(domain),
	}

	resp, err := s.client.GetEmailIdentity(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get DKIM attributes: %w", err)
	}

	return resp.DkimAttributes, nil
}

// RetryVerification retries the verification process for a domain
func (s *SESVerificationService) RetryVerification(customDomain *models.CustomDomain) error {
	// Reset some fields
	customDomain.Status = models.CustomDomainStatusPending
	customDomain.FailureReason = nil
	customDomain.VerifiedAt = nil

	// Re-initiate verification
	return s.InitiateDomainVerification(customDomain)
}
