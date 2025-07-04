package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maylng/backend/internal/models"
)

type EmailAddressService struct {
	db *sql.DB
}

func NewEmailAddressService(db *sql.DB) *EmailAddressService {
	return &EmailAddressService{db: db}
}

func (s *EmailAddressService) CreateEmailAddress(accountID uuid.UUID, req *models.CreateEmailAddressRequest) (*models.EmailAddressResponse, error) {
	// Check account limits
	count, err := s.countEmailAddresses(accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to check email address count: %w", err)
	}

	var limit int
	err = s.db.QueryRow("SELECT email_address_limit FROM accounts WHERE id = $1", accountID).Scan(&limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get account limits: %w", err)
	}

	if count >= limit {
		return nil, fmt.Errorf("email address limit reached (%d/%d)", count, limit)
	}

	// Generate email address
	email, prefix, domain := s.generateEmailAddress(req)

	// Set expiration for temporary emails
	var expiresAt *time.Time
	if req.Type == models.EmailAddressTypeTemporary {
		if req.ExpiresAt != nil {
			expiresAt = req.ExpiresAt
		} else {
			// Default to 24 hours for temporary emails
			expiry := time.Now().Add(24 * time.Hour)
			expiresAt = &expiry
		}
	}

	// Insert email address
	query := `
		INSERT INTO email_addresses (account_id, email, type, prefix, domain, expires_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, status, created_at, updated_at
	`

	var emailAddr models.EmailAddress
	err = s.db.QueryRow(query, accountID, email, req.Type, prefix, domain, expiresAt, req.Metadata).Scan(
		&emailAddr.ID,
		&emailAddr.Status,
		&emailAddr.CreatedAt,
		&emailAddr.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("email address already exists")
		}
		return nil, fmt.Errorf("failed to create email address: %w", err)
	}

	return &models.EmailAddressResponse{
		ID:        emailAddr.ID,
		Email:     email,
		Type:      req.Type,
		Prefix:    prefix,
		Domain:    domain,
		Status:    emailAddr.Status,
		ExpiresAt: expiresAt,
		Metadata:  req.Metadata,
		CreatedAt: emailAddr.CreatedAt,
		UpdatedAt: emailAddr.UpdatedAt,
	}, nil
}

func (s *EmailAddressService) GetEmailAddresses(accountID uuid.UUID) ([]*models.EmailAddressResponse, error) {
	query := `
		SELECT id, email, type, prefix, domain, status, expires_at, metadata, created_at, updated_at
		FROM email_addresses 
		WHERE account_id = $1 
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get email addresses: %w", err)
	}
	defer rows.Close()

	var emailAddresses []*models.EmailAddressResponse
	for rows.Next() {
		var addr models.EmailAddress
		err := rows.Scan(
			&addr.ID,
			&addr.Email,
			&addr.Type,
			&addr.Prefix,
			&addr.Domain,
			&addr.Status,
			&addr.ExpiresAt,
			&addr.Metadata,
			&addr.CreatedAt,
			&addr.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email address: %w", err)
		}

		emailAddresses = append(emailAddresses, &models.EmailAddressResponse{
			ID:        addr.ID,
			Email:     addr.Email,
			Type:      addr.Type,
			Prefix:    addr.Prefix,
			Domain:    addr.Domain,
			Status:    addr.Status,
			ExpiresAt: addr.ExpiresAt,
			Metadata:  addr.Metadata,
			CreatedAt: addr.CreatedAt,
			UpdatedAt: addr.UpdatedAt,
		})
	}

	return emailAddresses, nil
}

func (s *EmailAddressService) GetEmailAddress(accountID, emailAddressID uuid.UUID) (*models.EmailAddressResponse, error) {
	query := `
		SELECT id, email, type, prefix, domain, status, expires_at, metadata, created_at, updated_at
		FROM email_addresses 
		WHERE id = $1 AND account_id = $2
	`

	var addr models.EmailAddress
	err := s.db.QueryRow(query, emailAddressID, accountID).Scan(
		&addr.ID,
		&addr.Email,
		&addr.Type,
		&addr.Prefix,
		&addr.Domain,
		&addr.Status,
		&addr.ExpiresAt,
		&addr.Metadata,
		&addr.CreatedAt,
		&addr.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("email address not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get email address: %w", err)
	}

	return &models.EmailAddressResponse{
		ID:        addr.ID,
		Email:     addr.Email,
		Type:      addr.Type,
		Prefix:    addr.Prefix,
		Domain:    addr.Domain,
		Status:    addr.Status,
		ExpiresAt: addr.ExpiresAt,
		Metadata:  addr.Metadata,
		CreatedAt: addr.CreatedAt,
		UpdatedAt: addr.UpdatedAt,
	}, nil
}

func (s *EmailAddressService) UpdateEmailAddress(accountID, emailAddressID uuid.UUID, req *models.UpdateEmailAddressRequest) (*models.EmailAddressResponse, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.ExpiresAt != nil {
		setParts = append(setParts, fmt.Sprintf("expires_at = $%d", argIndex))
		args = append(args, *req.ExpiresAt)
		argIndex++
	}

	if req.Metadata != nil {
		setParts = append(setParts, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, req.Metadata)
		argIndex++
	}

	if len(setParts) == 0 {
		return s.GetEmailAddress(accountID, emailAddressID)
	}

	query := fmt.Sprintf(`
		UPDATE email_addresses 
		SET %s 
		WHERE id = $%d AND account_id = $%d
	`, strings.Join(setParts, ", "), argIndex, argIndex+1)

	args = append(args, emailAddressID, accountID)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update email address: %w", err)
	}

	return s.GetEmailAddress(accountID, emailAddressID)
}

func (s *EmailAddressService) DeleteEmailAddress(accountID, emailAddressID uuid.UUID) error {
	result, err := s.db.Exec("DELETE FROM email_addresses WHERE id = $1 AND account_id = $2", emailAddressID, accountID)
	if err != nil {
		return fmt.Errorf("failed to delete email address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("email address not found")
	}

	return nil
}

func (s *EmailAddressService) countEmailAddresses(accountID uuid.UUID) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM email_addresses WHERE account_id = $1", accountID).Scan(&count)
	return count, err
}

func (s *EmailAddressService) generateEmailAddress(req *models.CreateEmailAddressRequest) (email, prefix, domain string) {
	// Use provided domain or default
	domain = "maylng.dev" // Default domain
	if req.Domain != "" {
		domain = req.Domain
	}

	// Use provided prefix or generate one
	prefix = req.Prefix
	if prefix == "" {
		// Generate random prefix
		prefix = uuid.New().String()[:8]
	}

	email = fmt.Sprintf("%s@%s", prefix, domain)
	return email, prefix, domain
}
