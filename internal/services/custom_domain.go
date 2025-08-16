package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/maylng/backend/internal/models"
)

type CustomDomainService struct {
	db *sql.DB
}

func NewCustomDomainService(db *sql.DB) *CustomDomainService {
	return &CustomDomainService{
		db: db,
	}
}

// CreateCustomDomain creates a new custom domain for an account
func (s *CustomDomainService) CreateCustomDomain(accountID uuid.UUID, domain, verificationProvider string) (*models.CustomDomain, error) {
	if verificationProvider == "" {
		verificationProvider = "ses" // Default to SES for backward compatibility
	}

	customDomain := &models.CustomDomain{
		ID:                   uuid.New(),
		AccountID:            accountID,
		Domain:               domain,
		Status:               models.CustomDomainStatusPending,
		VerificationProvider: verificationProvider,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	dnsRecordsJSON, _ := json.Marshal([]models.DNSRecord{})
	dkimTokensJSON, _ := json.Marshal(map[string]interface{}{})
	metadataJSON, _ := json.Marshal(map[string]interface{}{})

	query := `
		INSERT INTO custom_domains (
			id, account_id, domain, status, verification_provider, dns_records, dkim_tokens, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := s.db.Exec(query,
		customDomain.ID,
		customDomain.AccountID,
		customDomain.Domain,
		customDomain.Status,
		customDomain.VerificationProvider,
		dnsRecordsJSON,
		dkimTokensJSON,
		metadataJSON,
		customDomain.CreatedAt,
		customDomain.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create custom domain: %w", err)
	}

	return customDomain, nil
}

// GetCustomDomainByID retrieves a custom domain by ID
func (s *CustomDomainService) GetCustomDomainByID(id uuid.UUID) (*models.CustomDomain, error) {
	var customDomain models.CustomDomain
	var dnsRecordsJSON, dkimTokensJSON, metadataJSON []byte

	query := `
		SELECT id, account_id, domain, status, verification_provider, provider_verification_status, provider_domain_id,
			   verification_token, dkim_tokens, dns_records,
			   ses_verification_status, ses_dkim_verification_status, verification_attempted_at,
			   verified_at, failure_reason, metadata, created_at, updated_at
		FROM custom_domains WHERE id = $1
	`

	err := s.db.QueryRow(query, id).Scan(
		&customDomain.ID,
		&customDomain.AccountID,
		&customDomain.Domain,
		&customDomain.Status,
		&customDomain.VerificationProvider,
		&customDomain.ProviderVerificationStatus,
		&customDomain.ProviderDomainID,
		&customDomain.VerificationToken,
		&dkimTokensJSON,
		&dnsRecordsJSON,
		&customDomain.SESVerificationStatus,
		&customDomain.SESDKIMVerificationStatus,
		&customDomain.VerificationAttemptedAt,
		&customDomain.VerifiedAt,
		&customDomain.FailureReason,
		&metadataJSON,
		&customDomain.CreatedAt,
		&customDomain.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("custom domain not found")
		}
		return nil, fmt.Errorf("failed to get custom domain: %w", err)
	}

	// Parse JSON fields
	if len(dnsRecordsJSON) > 0 {
		json.Unmarshal(dnsRecordsJSON, &customDomain.DNSRecords)
	}
	if len(dkimTokensJSON) > 0 {
		json.Unmarshal(dkimTokensJSON, &customDomain.DKIMTokens)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &customDomain.Metadata)
	}

	return &customDomain, nil
}

// GetCustomDomainsByAccountID retrieves all custom domains for an account
func (s *CustomDomainService) GetCustomDomainsByAccountID(accountID uuid.UUID) ([]*models.CustomDomain, error) {
	query := `
		SELECT id, account_id, domain, status, verification_provider, provider_verification_status, provider_domain_id,
			   verification_token, dkim_tokens, dns_records,
			   ses_verification_status, ses_dkim_verification_status, verification_attempted_at,
			   verified_at, failure_reason, metadata, created_at, updated_at
		FROM custom_domains WHERE account_id = $1 ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom domains: %w", err)
	}
	defer rows.Close()

	var domains []*models.CustomDomain
	for rows.Next() {
		var customDomain models.CustomDomain
		var dnsRecordsJSON, dkimTokensJSON, metadataJSON []byte

		err := rows.Scan(
			&customDomain.ID,
			&customDomain.AccountID,
			&customDomain.Domain,
			&customDomain.Status,
			&customDomain.VerificationProvider,
			&customDomain.ProviderVerificationStatus,
			&customDomain.ProviderDomainID,
			&customDomain.VerificationToken,
			&dkimTokensJSON,
			&dnsRecordsJSON,
			&customDomain.SESVerificationStatus,
			&customDomain.SESDKIMVerificationStatus,
			&customDomain.VerificationAttemptedAt,
			&customDomain.VerifiedAt,
			&customDomain.FailureReason,
			&metadataJSON,
			&customDomain.CreatedAt,
			&customDomain.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan custom domain: %w", err)
		}

		// Parse JSON fields
		if len(dnsRecordsJSON) > 0 {
			json.Unmarshal(dnsRecordsJSON, &customDomain.DNSRecords)
		}
		if len(dkimTokensJSON) > 0 {
			json.Unmarshal(dkimTokensJSON, &customDomain.DKIMTokens)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &customDomain.Metadata)
		}

		domains = append(domains, &customDomain)
	}

	return domains, nil
}

// UpdateCustomDomain updates a custom domain
func (s *CustomDomainService) UpdateCustomDomain(customDomain *models.CustomDomain) error {
	dnsRecordsJSON, _ := json.Marshal(customDomain.DNSRecords)
	dkimTokensJSON, _ := json.Marshal(customDomain.DKIMTokens)
	metadataJSON, _ := json.Marshal(customDomain.Metadata)

	query := `
		UPDATE custom_domains SET
			status = $1,
			verification_provider = $2,
			provider_verification_status = $3,
			provider_domain_id = $4,
			verification_token = $5,
			dkim_tokens = $6,
			dns_records = $7,
			ses_verification_status = $8,
			ses_dkim_verification_status = $9,
			verification_attempted_at = $10,
			verified_at = $11,
			failure_reason = $12,
			metadata = $13,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $14
	`

	_, err := s.db.Exec(query,
		customDomain.Status,
		customDomain.VerificationProvider,
		customDomain.ProviderVerificationStatus,
		customDomain.ProviderDomainID,
		customDomain.VerificationToken,
		dkimTokensJSON,
		dnsRecordsJSON,
		customDomain.SESVerificationStatus,
		customDomain.SESDKIMVerificationStatus,
		customDomain.VerificationAttemptedAt,
		customDomain.VerifiedAt,
		customDomain.FailureReason,
		metadataJSON,
		customDomain.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update custom domain: %w", err)
	}

	return nil
}

// DeleteCustomDomain deletes a custom domain
func (s *CustomDomainService) DeleteCustomDomain(id uuid.UUID, accountID uuid.UUID) error {
	query := `DELETE FROM custom_domains WHERE id = $1 AND account_id = $2`
	result, err := s.db.Exec(query, id, accountID)
	if err != nil {
		return fmt.Errorf("failed to delete custom domain: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("custom domain not found or access denied")
	}

	return nil
}

// GetCustomDomainByDomain retrieves a custom domain by domain name
func (s *CustomDomainService) GetCustomDomainByDomain(domain string) (*models.CustomDomain, error) {
	var customDomain models.CustomDomain
	var dnsRecordsJSON, dkimTokensJSON, metadataJSON []byte

	query := `
		SELECT id, account_id, domain, status, verification_provider, provider_verification_status, provider_domain_id,
			   verification_token, dkim_tokens, dns_records,
			   ses_verification_status, ses_dkim_verification_status, verification_attempted_at,
			   verified_at, failure_reason, metadata, created_at, updated_at
		FROM custom_domains WHERE domain = $1
	`

	err := s.db.QueryRow(query, domain).Scan(
		&customDomain.ID,
		&customDomain.AccountID,
		&customDomain.Domain,
		&customDomain.Status,
		&customDomain.VerificationProvider,
		&customDomain.ProviderVerificationStatus,
		&customDomain.ProviderDomainID,
		&customDomain.VerificationToken,
		&dkimTokensJSON,
		&dnsRecordsJSON,
		&customDomain.SESVerificationStatus,
		&customDomain.SESDKIMVerificationStatus,
		&customDomain.VerificationAttemptedAt,
		&customDomain.VerifiedAt,
		&customDomain.FailureReason,
		&metadataJSON,
		&customDomain.CreatedAt,
		&customDomain.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("custom domain not found")
		}
		return nil, fmt.Errorf("failed to get custom domain: %w", err)
	}

	// Parse JSON fields
	if len(dnsRecordsJSON) > 0 {
		json.Unmarshal(dnsRecordsJSON, &customDomain.DNSRecords)
	}
	if len(dkimTokensJSON) > 0 {
		json.Unmarshal(dkimTokensJSON, &customDomain.DKIMTokens)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &customDomain.Metadata)
	}

	return &customDomain, nil
}
