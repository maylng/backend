package models

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	APIKeyHash         string    `json:"-" db:"api_key_hash"`
	Plan               string    `json:"plan" db:"plan"`
	EmailLimitPerMonth int       `json:"email_limit_per_month" db:"email_limit_per_month"`
	EmailAddressLimit  int       `json:"email_address_limit" db:"email_address_limit"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type CreateAccountRequest struct {
	Plan string `json:"plan" validate:"omitempty,oneof=free pro enterprise"`
}

type AccountResponse struct {
	ID                   uuid.UUID `json:"id"`
	Plan                 string    `json:"plan"`
	EmailLimitPerMonth   int       `json:"email_limit_per_month"`
	EmailAddressLimit    int       `json:"email_address_limit"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	APIKey               string    `json:"api_key,omitempty"` // Only returned on creation
	EmailAddressesCount  int       `json:"email_addresses_count,omitempty"`
	TPSCount             int       `json:"tps_count,omitempty"` // 3rd Party Software
	CustomDomainsCount   int       `json:"custom_domains_count,omitempty"`
	VerifiedDomainsCount int       `json:"verified_domains_count,omitempty"`
}

type UpdateAccountRequest struct {
	Plan *string `json:"plan" validate:"omitempty,oneof=free pro enterprise"`
}
