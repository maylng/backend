package services

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/maylng/backend/internal/auth"
	"github.com/maylng/backend/internal/models"
)

type AccountService struct {
	db   *sql.DB
	salt string
}

func NewAccountService(db *sql.DB, salt string) *AccountService {
	return &AccountService{
		db:   db,
		salt: salt,
	}
}

func (s *AccountService) CreateAccount(req *models.CreateAccountRequest) (*models.AccountResponse, error) {
	// Generate API key
	apiKey, err := auth.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash API key
	hashedKey := auth.HashAPIKey(apiKey, s.salt)

	// Set default plan if not provided
	plan := "free"
	if req.Plan != "" {
		plan = req.Plan
	}

	emailLimitPerMonth, emailAddressLimit, err := getPlanLimits(plan)
	if err != nil {
		return nil, err
	}

	// Insert account into database
	query := `
		INSERT INTO accounts (api_key_hash, plan, email_limit_per_month, email_address_limit)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	var account models.Account
	err = s.db.QueryRow(query, hashedKey, plan, emailLimitPerMonth, emailAddressLimit).Scan(
		&account.ID,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return &models.AccountResponse{
		ID:                 account.ID,
		Plan:               plan,
		EmailLimitPerMonth: emailLimitPerMonth,
		EmailAddressLimit:  emailAddressLimit,
		CreatedAt:          account.CreatedAt,
		UpdatedAt:          account.UpdatedAt,
		APIKey:             apiKey, // Only returned on creation
	}, nil
}

func getPlanLimits(plan string) (emailLimitPerMonth int, emailAddressLimit int, err error) {
	switch plan {
	case "free":
		emailLimitPerMonth = 1000
		emailAddressLimit = 5
	case "pro":
		emailLimitPerMonth = 50000
		emailAddressLimit = 50
	case "enterprise":
		emailLimitPerMonth = 1000000
		emailAddressLimit = 500
	default:
		return 0, 0, fmt.Errorf("invalid plan: %s", plan)
	}
	return emailLimitPerMonth, emailAddressLimit, nil
}

func (s *AccountService) GetAccount(accountID uuid.UUID) (*models.AccountResponse, error) {
	query := `
		SELECT id, plan, email_limit_per_month, email_address_limit, created_at, updated_at
		FROM accounts WHERE id = $1
	`

	var account models.Account
	err := s.db.QueryRow(query, accountID).Scan(
		&account.ID,
		&account.Plan,
		&account.EmailLimitPerMonth,
		&account.EmailAddressLimit,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Get email addresses count
	emailAddressesCount, err := s.getEmailAddressesCount(accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get email addresses count: %w", err)
	}

	return &models.AccountResponse{
		ID:                  account.ID,
		Plan:                account.Plan,
		EmailLimitPerMonth:  account.EmailLimitPerMonth,
		EmailAddressLimit:   account.EmailAddressLimit,
		EmailAddressesCount: emailAddressesCount,
		CreatedAt:           account.CreatedAt,
		UpdatedAt:           account.UpdatedAt,
	}, nil
}

func (s *AccountService) UpdateAccount(accountID uuid.UUID, req *models.UpdateAccountRequest) (*models.AccountResponse, error) {
	// Get current account details
	currentAccount, err := s.GetAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account for update: %w", err)
	}

	plan := currentAccount.Plan
	if req.Plan != nil {
		plan = *req.Plan
	}

	emailLimitPerMonth, emailAddressLimit, err := getPlanLimits(plan)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE accounts
		SET plan = $1, email_limit_per_month = $2, email_address_limit = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, plan, email_limit_per_month, email_address_limit, created_at, updated_at
	`
	var account models.Account
	err = s.db.QueryRow(query, plan, emailLimitPerMonth, emailAddressLimit, accountID).Scan(
		&account.ID,
		&account.Plan,
		&account.EmailLimitPerMonth,
		&account.EmailAddressLimit,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	// Get email addresses count
	emailAddressesCount, err := s.getEmailAddressesCount(accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get email addresses count: %w", err)
	}

	return &models.AccountResponse{
		ID:                  account.ID,
		Plan:                account.Plan,
		EmailLimitPerMonth:  account.EmailLimitPerMonth,
		EmailAddressLimit:   account.EmailAddressLimit,
		EmailAddressesCount: emailAddressesCount,
		CreatedAt:           account.CreatedAt,
		UpdatedAt:           account.UpdatedAt,
	}, nil
}

func (s *AccountService) DeleteAccount(accountID uuid.UUID) error {
	// Ensure deletion is scoped by account ID to prevent mass deletion
	query := `DELETE FROM accounts WHERE id = $1`
	result, err := s.db.Exec(query, accountID)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found")
	}

	return nil
}

func (s *AccountService) GenerateNewAPIKey(accountID uuid.UUID) (string, error) {
	// Generate a new API key
	newAPIKey, err := auth.GenerateAPIKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate new API key: %w", err)
	}

	// Hash the new API key
	hashedKey := auth.HashAPIKey(newAPIKey, s.salt)

	// Update the account with the new hashed API key
	query := `
		UPDATE accounts
		SET api_key_hash = $1, updated_at = NOW()
		WHERE id = $2
	`
	result, err := s.db.Exec(query, hashedKey, accountID)
	if err != nil {
		return "", fmt.Errorf("failed to update API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return "", fmt.Errorf("account not found")
	}

	return newAPIKey, nil
}

// getEmailAddressesCount returns the count of email addresses for a given account
func (s *AccountService) getEmailAddressesCount(accountID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM email_addresses 
		WHERE account_id = $1 AND status != 'disabled'
	`

	var count int
	err := s.db.QueryRow(query, accountID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count email addresses: %w", err)
	}

	return count, nil
}
