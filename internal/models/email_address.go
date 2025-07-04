package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EmailAddressType string

const (
	EmailAddressTypeTemporary  EmailAddressType = "temporary"
	EmailAddressTypePersistent EmailAddressType = "persistent"
)

type EmailAddressStatus string

const (
	EmailAddressStatusActive   EmailAddressStatus = "active"
	EmailAddressStatusExpired  EmailAddressStatus = "expired"
	EmailAddressStatusDisabled EmailAddressStatus = "disabled"
)

type Metadata map[string]interface{}

func (m Metadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(Metadata)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

type EmailAddress struct {
	ID        uuid.UUID          `json:"id" db:"id"`
	AccountID uuid.UUID          `json:"account_id" db:"account_id"`
	Email     string             `json:"email" db:"email"`
	Type      EmailAddressType   `json:"type" db:"type"`
	Prefix    string             `json:"prefix" db:"prefix"`
	Domain    string             `json:"domain" db:"domain"`
	Status    EmailAddressStatus `json:"status" db:"status"`
	ExpiresAt *time.Time         `json:"expires_at" db:"expires_at"`
	Metadata  Metadata           `json:"metadata" db:"metadata"`
	CreatedAt time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" db:"updated_at"`
}

type CreateEmailAddressRequest struct {
	Type      EmailAddressType `json:"type" validate:"required,oneof=temporary persistent"`
	Prefix    string           `json:"prefix" validate:"omitempty,min=1,max=100"`
	Domain    string           `json:"domain" validate:"omitempty,hostname"`
	ExpiresAt *time.Time       `json:"expires_at"`
	Metadata  Metadata         `json:"metadata"`
}

type UpdateEmailAddressRequest struct {
	Status    *EmailAddressStatus `json:"status" validate:"omitempty,oneof=active expired disabled"`
	ExpiresAt *time.Time          `json:"expires_at"`
	Metadata  Metadata            `json:"metadata"`
}

type EmailAddressResponse struct {
	ID        uuid.UUID          `json:"id"`
	Email     string             `json:"email"`
	Type      EmailAddressType   `json:"type"`
	Prefix    string             `json:"prefix"`
	Domain    string             `json:"domain"`
	Status    EmailAddressStatus `json:"status"`
	ExpiresAt *time.Time         `json:"expires_at"`
	Metadata  Metadata           `json:"metadata"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}
