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

	// Set limits based on plan
	var emailLimitPerMonth, emailAddressLimit int
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
		return nil, fmt.Errorf("invalid plan: %s", plan)
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

	return &models.AccountResponse{
		ID:                 account.ID,
		Plan:               account.Plan,
		EmailLimitPerMonth: account.EmailLimitPerMonth,
		EmailAddressLimit:  account.EmailAddressLimit,
		CreatedAt:          account.CreatedAt,
		UpdatedAt:          account.UpdatedAt,
	}, nil
}
