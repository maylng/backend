// TPS (3rd Party Software) Data Model
package models

import (
	"time"

	"github.com/google/uuid"
)

type TPSStatus string

const (
	TPSStatusActive    TPSStatus = "active"
	TPSStatusInactive  TPSStatus = "inactive"
	TPSStatusPending   TPSStatus = "pending"
	TPSStatusFailed    TPSStatus = "failed"
	TPSStatusSuspended TPSStatus = "suspended"
)

type TPS struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	EmailAddressID uuid.UUID      `json:"email_address_id" db:"email_address_id"`
	ServiceName    string         `json:"service_name" db:"service_name"`
	ServiceType    string         `json:"service_type" db:"service_type"`
	ServiceURL     string         `json:"service_url" db:"service_url"`
	HasPremium     bool           `json:"has_premium" db:"has_premium"`
	IsPremium      bool           `json:"is_premium" db:"is_premium"`
	Description    *string        `json:"description,omitempty" db:"description"`
	APIKey         *string        `json:"api_key,omitempty" db:"api_key"`
	Username       *string        `json:"username,omitempty" db:"username"`
	Password       *string        `json:"password,omitempty" db:"password"`
	Status         string         `json:"status" db:"status"`
	Metadata       map[string]any `json:"metadata" db:"metadata"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

type CreateTPSRequest struct {
	EmailAddressID uuid.UUID      `json:"email_address_id" validate:"required"`
	ServiceName    string         `json:"service_name" validate:"required,min=1,max=100"`
	ServiceType    string         `json:"service_type" validate:"required,min=1,max=50"`
	ServiceURL     string         `json:"service_url" validate:"required,url"`
	HasPremium     bool           `json:"has_premium"`
	IsPremium      bool           `json:"is_premium"`
	Description    *string        `json:"description,omitempty" validate:"omitempty,max=500"`
	APIKey         *string        `json:"api_key,omitempty"`
	Username       *string        `json:"username,omitempty" validate:"omitempty,min=1,max=100"`
	Password       *string        `json:"password,omitempty"`
	Metadata       map[string]any `json:"metadata"`
}

type UpdateTPSRequest struct {
	ServiceName *string        `json:"service_name,omitempty" validate:"omitempty,min=1,max=100"`
	ServiceType *string        `json:"service_type,omitempty" validate:"omitempty,min=1,max=50"`
	ServiceURL  *string        `json:"service_url,omitempty" validate:"omitempty,url"`
	HasPremium  *bool          `json:"has_premium,omitempty"`
	IsPremium   *bool          `json:"is_premium,omitempty"`
	Description *string        `json:"description,omitempty" validate:"omitempty,max=500"`
	APIKey      *string        `json:"api_key,omitempty"`
	Username    *string        `json:"username,omitempty" validate:"omitempty,min=1,max=100"`
	Password    *string        `json:"password,omitempty"`
	Status      *TPSStatus     `json:"status,omitempty" validate:"omitempty,oneof=active inactive pending failed suspended"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type TPSResponse struct {
	ID             uuid.UUID      `json:"id"`
	EmailAddressID uuid.UUID      `json:"email_address_id"`
	ServiceName    string         `json:"service_name"`
	ServiceType    string         `json:"service_type"`
	ServiceURL     string         `json:"service_url"`
	HasPremium     bool           `json:"has_premium"`
	IsPremium      bool           `json:"is_premium"`
	Description    *string        `json:"description,omitempty"`
	HasAPIKey      bool           `json:"has_api_key"`
	Username       *string        `json:"username,omitempty"`
	HasPassword    bool           `json:"has_password"`
	Status         string         `json:"status"`
	Metadata       map[string]any `json:"metadata"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}
