// TPS (3rd Party Software) Business Logic
package services

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/maylng/backend/internal/auth"
	"github.com/maylng/backend/internal/models"
)

type TPSService struct {
	db            *sql.DB
	encryptionKey string
}

func NewTPSService(db *sql.DB, encryptionKey string) *TPSService {
	return &TPSService{
		db:            db,
		encryptionKey: encryptionKey,
	}
}

// GetTPSLimitByPlan returns the 3rd Party Software limit per agent email address for a given plan
func GetTPSLimitByPlan(plan string) int {
	switch plan {
		case "free":
			return 1
		case "pro":
			return 3
		case "enterprise":
			return 20
		default:
			return 1
	}
}

// CountTPSForEmail returns the number of 3rd Party Software records for a given agent email address
func (s *TPSService) CountTPSForEmail(emailAddressID uuid.UUID) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM tps WHERE email_address_id = $1", emailAddressID).Scan(&count)
	return count, err
}

// CreateTPS creates a new 3rd Party Software record, enforcing the plan limit
func (s *TPSService) CreateTPS(accountPlan string, req *models.CreateTPSRequest) (*models.TPSResponse, error) {
	// Check TPS limit for this email address
	count, err := s.CountTPSForEmail(req.EmailAddressID)
	if err != nil {
		return nil, fmt.Errorf("failed to count TPS: %w", err)
	}
	limit := GetTPSLimitByPlan(accountPlan)
	if count >= limit {
		return nil, fmt.Errorf("TPS limit reached for this agent email address (%d/%d)", count, limit)
	}

	// Encrypt sensitive data before storing
	var encryptedAPIKey *string
	if req.APIKey != nil && *req.APIKey != "" {
		encrypted, err := auth.EncryptString(*req.APIKey, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %w", err)
		}
		encryptedAPIKey = &encrypted
	}

	var encryptedPassword *string
	if req.Password != nil && *req.Password != "" {
		encrypted, err := auth.EncryptString(*req.Password, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		encryptedPassword = &encrypted
	}

	query := `
		INSERT INTO tps (email_address_id, service_name, service_type, service_url, has_premium, is_premium, description, api_key, username, password, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`
	var tps models.TPS
	err = s.db.QueryRow(query,
		req.EmailAddressID,
		req.ServiceName,
		req.ServiceType,
		req.ServiceURL,
		req.HasPremium,
		req.IsPremium,
		req.Description,
		encryptedAPIKey,
		req.Username,
		encryptedPassword,
		models.TPSStatusActive, // Default status
		req.Metadata).Scan(
		&tps.ID, &tps.CreatedAt, &tps.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TPS: %w", err)
	}

	// Return response without sensitive data for security
	return &models.TPSResponse{
		ID:             tps.ID,
		EmailAddressID: req.EmailAddressID,
		ServiceName:    req.ServiceName,
		ServiceType:    req.ServiceType,
		ServiceURL:     req.ServiceURL,
		HasPremium:     req.HasPremium,
		IsPremium:      req.IsPremium,
		Description:    req.Description,
		HasAPIKey:      encryptedAPIKey != nil,
		Username:       req.Username,
		HasPassword:    encryptedPassword != nil,
		Status:         string(models.TPSStatusActive),
		Metadata:       req.Metadata,
		CreatedAt:      tps.CreatedAt,
		UpdatedAt:      tps.UpdatedAt,
	}, nil
}

// GetTPS retrieves a 3rd Party Software record by ID
func (s *TPSService) GetTPS(tpsID uuid.UUID) (*models.TPSResponse, error) {
	var tps models.TPS
	query := `
		SELECT id, email_address_id, service_name, service_type, service_url, 
		       has_premium, is_premium, description, api_key, username, password, 
		       status, metadata, created_at, updated_at
		FROM tps WHERE id = $1
	`
	err := s.db.QueryRow(query, tpsID).Scan(
		&tps.ID, &tps.EmailAddressID, &tps.ServiceName, &tps.ServiceType,
		&tps.ServiceURL, &tps.HasPremium, &tps.IsPremium, &tps.Description,
		&tps.APIKey, &tps.Username, &tps.Password, &tps.Status, &tps.Metadata,
		&tps.CreatedAt, &tps.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get TPS: %w", err)
	}

	return &models.TPSResponse{
		ID:             tps.ID,
		EmailAddressID: tps.EmailAddressID,
		ServiceName:    tps.ServiceName,
		ServiceType:    tps.ServiceType,
		ServiceURL:     tps.ServiceURL,
		HasPremium:     tps.HasPremium,
		IsPremium:      tps.IsPremium,
		Description:    tps.Description,
		HasAPIKey:      tps.APIKey != nil && *tps.APIKey != "",
		Username:       tps.Username,
		HasPassword:    tps.Password != nil && *tps.Password != "",
		Status:         tps.Status,
		Metadata:       tps.Metadata,
		CreatedAt:      tps.CreatedAt,
		UpdatedAt:      tps.UpdatedAt,
	}, nil
}

// GetDecryptedAPIKey retrieves and decrypts the API key for a 3rd Party Software record
func (s *TPSService) GetDecryptedAPIKey(tpsID uuid.UUID) (*string, error) {
	var encryptedAPIKey *string
	err := s.db.QueryRow("SELECT api_key FROM tps WHERE id = $1", tpsID).Scan(&encryptedAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	if encryptedAPIKey == nil || *encryptedAPIKey == "" {
		return nil, nil
	}

	decrypted, err := auth.DecryptString(*encryptedAPIKey, s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}

	return &decrypted, nil
}

// GetDecryptedPassword retrieves and decrypts the password for a 3rd Party Software record
func (s *TPSService) GetDecryptedPassword(tpsID uuid.UUID) (*string, error) {
	var encryptedPassword *string
	err := s.db.QueryRow("SELECT password FROM tps WHERE id = $1", tpsID).Scan(&encryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to get password: %w", err)
	}

	if encryptedPassword == nil || *encryptedPassword == "" {
		return nil, nil
	}

	decrypted, err := auth.DecryptString(*encryptedPassword, s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}

	return &decrypted, nil
}

// UpdateTPS updates a 3rd Party Software record
func (s *TPSService) UpdateTPS(tpsID uuid.UUID, req *models.UpdateTPSRequest) (*models.TPSResponse, error) {
	// Build dynamic query based on provided fields
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.ServiceName != nil {
		setParts = append(setParts, fmt.Sprintf("service_name = $%d", argIndex))
		args = append(args, *req.ServiceName)
		argIndex++
	}

	if req.ServiceType != nil {
		setParts = append(setParts, fmt.Sprintf("service_type = $%d", argIndex))
		args = append(args, *req.ServiceType)
		argIndex++
	}

	if req.ServiceURL != nil {
		setParts = append(setParts, fmt.Sprintf("service_url = $%d", argIndex))
		args = append(args, *req.ServiceURL)
		argIndex++
	}

	if req.HasPremium != nil {
		setParts = append(setParts, fmt.Sprintf("has_premium = $%d", argIndex))
		args = append(args, *req.HasPremium)
		argIndex++
	}

	if req.IsPremium != nil {
		setParts = append(setParts, fmt.Sprintf("is_premium = $%d", argIndex))
		args = append(args, *req.IsPremium)
		argIndex++
	}

	if req.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, req.Description)
		argIndex++
	}

	if req.APIKey != nil {
		var encryptedAPIKey *string
		if *req.APIKey != "" {
			encrypted, err := auth.EncryptString(*req.APIKey, s.encryptionKey)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt API key: %w", err)
			}
			encryptedAPIKey = &encrypted
		}
		setParts = append(setParts, fmt.Sprintf("api_key = $%d", argIndex))
		args = append(args, encryptedAPIKey)
		argIndex++
	}

	if req.Username != nil {
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, req.Username)
		argIndex++
	}

	if req.Password != nil {
		var encryptedPassword *string
		if *req.Password != "" {
			encrypted, err := auth.EncryptString(*req.Password, s.encryptionKey)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt password: %w", err)
			}
			encryptedPassword = &encrypted
		}
		setParts = append(setParts, fmt.Sprintf("password = $%d", argIndex))
		args = append(args, encryptedPassword)
		argIndex++
	}

	if req.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, string(*req.Status))
		argIndex++
	}

	if req.Metadata != nil {
		setParts = append(setParts, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, req.Metadata)
		argIndex++
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Add updated_at
	setParts = append(setParts, "updated_at = NOW()")

	// Add ID to args
	args = append(args, tpsID)
	whereClause := fmt.Sprintf("id = $%d", argIndex)

	query := fmt.Sprintf(`
		UPDATE tps 
		SET %s 
		WHERE %s
		RETURNING id, email_address_id, service_name, service_type, service_url, 
		          has_premium, is_premium, description, api_key, username, password, 
		          status, metadata, created_at, updated_at
	`, strings.Join(setParts, ", "), whereClause)

	var tps models.TPS
	err := s.db.QueryRow(query, args...).Scan(
		&tps.ID, &tps.EmailAddressID, &tps.ServiceName, &tps.ServiceType,
		&tps.ServiceURL, &tps.HasPremium, &tps.IsPremium, &tps.Description,
		&tps.APIKey, &tps.Username, &tps.Password, &tps.Status, &tps.Metadata,
		&tps.CreatedAt, &tps.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update TPS: %w", err)
	}

	return &models.TPSResponse{
		ID:             tps.ID,
		EmailAddressID: tps.EmailAddressID,
		ServiceName:    tps.ServiceName,
		ServiceType:    tps.ServiceType,
		ServiceURL:     tps.ServiceURL,
		HasPremium:     tps.HasPremium,
		IsPremium:      tps.IsPremium,
		Description:    tps.Description,
		HasAPIKey:      tps.APIKey != nil && *tps.APIKey != "",
		Username:       tps.Username,
		HasPassword:    tps.Password != nil && *tps.Password != "",
		Status:         tps.Status,
		Metadata:       tps.Metadata,
		CreatedAt:      tps.CreatedAt,
		UpdatedAt:      tps.UpdatedAt,
	}, nil
}

// DeleteTPS deletes a 3rd Party Software record
func (s *TPSService) DeleteTPS(tpsID uuid.UUID) error {
	result, err := s.db.Exec("DELETE FROM tps WHERE id = $1", tpsID)
	if err != nil {
		return fmt.Errorf("failed to delete TPS: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("TPS not found")
	}

	return nil
}

// ListTPSByEmailAddress lists all 3rd Party Software records for a given email address
func (s *TPSService) ListTPSByEmailAddress(emailAddressID uuid.UUID) ([]*models.TPSResponse, error) {
	query := `
		SELECT id, email_address_id, service_name, service_type, service_url, 
		       has_premium, is_premium, description, api_key, username, password, 
		       status, metadata, created_at, updated_at
		FROM tps WHERE email_address_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query, emailAddressID)
	if err != nil {
		return nil, fmt.Errorf("failed to list TPS: %w", err)
	}
	defer rows.Close()

	var tpsList []*models.TPSResponse
	for rows.Next() {
		var tps models.TPS
		err := rows.Scan(
			&tps.ID, &tps.EmailAddressID, &tps.ServiceName, &tps.ServiceType,
			&tps.ServiceURL, &tps.HasPremium, &tps.IsPremium, &tps.Description,
			&tps.APIKey, &tps.Username, &tps.Password, &tps.Status, &tps.Metadata,
			&tps.CreatedAt, &tps.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan TPS: %w", err)
		}

		tpsList = append(tpsList, &models.TPSResponse{
			ID:             tps.ID,
			EmailAddressID: tps.EmailAddressID,
			ServiceName:    tps.ServiceName,
			ServiceType:    tps.ServiceType,
			ServiceURL:     tps.ServiceURL,
			HasPremium:     tps.HasPremium,
			IsPremium:      tps.IsPremium,
			Description:    tps.Description,
			HasAPIKey:      tps.APIKey != nil && *tps.APIKey != "",
			Username:       tps.Username,
			HasPassword:    tps.Password != nil && *tps.Password != "",
			Status:         tps.Status,
			Metadata:       tps.Metadata,
			CreatedAt:      tps.CreatedAt,
			UpdatedAt:      tps.UpdatedAt,
		})
	}

	return tpsList, nil
}
