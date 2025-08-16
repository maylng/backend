package models

import (
	"time"

	"github.com/google/uuid"
)

type CustomDomainStatus string

const (
	CustomDomainStatusPending  CustomDomainStatus = "pending"
	CustomDomainStatusVerified CustomDomainStatus = "verified"
	CustomDomainStatusFailed   CustomDomainStatus = "failed"
	CustomDomainStatusDisabled CustomDomainStatus = "disabled"
)

type SESVerificationStatus string

const (
	SESVerificationStatusPending          SESVerificationStatus = "Pending"
	SESVerificationStatusSuccess          SESVerificationStatus = "Success"
	SESVerificationStatusFailed           SESVerificationStatus = "Failed"
	SESVerificationStatusTemporaryFailure SESVerificationStatus = "TemporaryFailure"
	SESVerificationStatusNotStarted       SESVerificationStatus = "NotStarted"
)

type DNSRecord struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

type CustomDomain struct {
	ID                         uuid.UUID              `json:"id" db:"id"`
	AccountID                  uuid.UUID              `json:"account_id" db:"account_id"`
	Domain                     string                 `json:"domain" db:"domain"`
	Status                     CustomDomainStatus     `json:"status" db:"status"`
	VerificationProvider       string                 `json:"verification_provider" db:"verification_provider"`
	ProviderVerificationStatus *string                `json:"provider_verification_status,omitempty" db:"provider_verification_status"`
	ProviderDomainID           *string                `json:"provider_domain_id,omitempty" db:"provider_domain_id"`
	VerificationToken          *string                `json:"verification_token,omitempty" db:"verification_token"`
	DKIMTokens                 map[string]interface{} `json:"dkim_tokens,omitempty" db:"dkim_tokens"`
	DNSRecords                 []DNSRecord            `json:"dns_records,omitempty" db:"dns_records"`
	SESVerificationStatus      *string                `json:"ses_verification_status,omitempty" db:"ses_verification_status"`
	SESDKIMVerificationStatus  *string                `json:"ses_dkim_verification_status,omitempty" db:"ses_dkim_verification_status"`
	VerificationAttemptedAt    *time.Time             `json:"verification_attempted_at,omitempty" db:"verification_attempted_at"`
	VerifiedAt                 *time.Time             `json:"verified_at,omitempty" db:"verified_at"`
	FailureReason              *string                `json:"failure_reason,omitempty" db:"failure_reason"`
	Metadata                   map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt                  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt                  time.Time              `json:"updated_at" db:"updated_at"`
}

// IsVerified returns true if the domain is verified
func (cd *CustomDomain) IsVerified() bool {
	return cd.Status == CustomDomainStatusVerified
}

// CanSendEmails returns true if the domain can be used for sending emails
func (cd *CustomDomain) CanSendEmails() bool {
	return cd.Status == CustomDomainStatusVerified
}

// GetDefaultFromAddress returns a default from address for this domain
func (cd *CustomDomain) GetDefaultFromAddress() string {
	return "noreply@" + cd.Domain
}
