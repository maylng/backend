package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EmailStatus string

const (
	EmailStatusQueued    EmailStatus = "queued"
	EmailStatusSent      EmailStatus = "sent"
	EmailStatusDelivered EmailStatus = "delivered"
	EmailStatusFailed    EmailStatus = "failed"
	EmailStatusScheduled EmailStatus = "scheduled"
)

type Recipients []string

func (r Recipients) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *Recipients) Scan(value interface{}) error {
	if value == nil {
		*r = Recipients{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, r)
}

type SentEmail struct {
	ID                uuid.UUID   `json:"id" db:"id"`
	AccountID         uuid.UUID   `json:"account_id" db:"account_id"`
	FromEmailID       uuid.UUID   `json:"from_email_id" db:"from_email_id"`
	ToRecipients      Recipients  `json:"to_recipients" db:"to_recipients"`
	CcRecipients      Recipients  `json:"cc_recipients" db:"cc_recipients"`
	BccRecipients     Recipients  `json:"bcc_recipients" db:"bcc_recipients"`
	Subject           string      `json:"subject" db:"subject"`
	TextContent       *string     `json:"text_content" db:"text_content"`
	HTMLContent       *string     `json:"html_content" db:"html_content"`
	Attachments       Metadata    `json:"attachments" db:"attachments"`
	Headers           Metadata    `json:"headers" db:"headers"`
	ThreadID          *uuid.UUID  `json:"thread_id" db:"thread_id"`
	ScheduledAt       *time.Time  `json:"scheduled_at" db:"scheduled_at"`
	SentAt            *time.Time  `json:"sent_at" db:"sent_at"`
	Status            EmailStatus `json:"status" db:"status"`
	ProviderMessageID *string     `json:"provider_message_id" db:"provider_message_id"`
	FailureReason     *string     `json:"failure_reason" db:"failure_reason"`
	Metadata          Metadata    `json:"metadata" db:"metadata"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at" db:"updated_at"`
}

type SendEmailRequest struct {
	FromEmailID   uuid.UUID  `json:"from_email_id" validate:"required"`
	ToRecipients  Recipients `json:"to_recipients" validate:"required,min=1,dive,email"`
	CcRecipients  Recipients `json:"cc_recipients" validate:"omitempty,dive,email"`
	BccRecipients Recipients `json:"bcc_recipients" validate:"omitempty,dive,email"`
	Subject       string     `json:"subject" validate:"required,max=998"`
	TextContent   *string    `json:"text_content"`
	HTMLContent   *string    `json:"html_content"`
	Attachments   Metadata   `json:"attachments"`
	Headers       Metadata   `json:"headers"`
	ThreadID      *uuid.UUID `json:"thread_id"`
	ScheduledAt   *time.Time `json:"scheduled_at"`
	Metadata      Metadata   `json:"metadata"`
}

type EmailResponse struct {
	ID                uuid.UUID   `json:"id"`
	FromEmailID       uuid.UUID   `json:"from_email_id"`
	ToRecipients      Recipients  `json:"to_recipients"`
	CcRecipients      Recipients  `json:"cc_recipients"`
	BccRecipients     Recipients  `json:"bcc_recipients"`
	Subject           string      `json:"subject"`
	TextContent       *string     `json:"text_content"`
	HTMLContent       *string     `json:"html_content"`
	ThreadID          *uuid.UUID  `json:"thread_id"`
	ScheduledAt       *time.Time  `json:"scheduled_at"`
	SentAt            *time.Time  `json:"sent_at"`
	Status            EmailStatus `json:"status"`
	ProviderMessageID *string     `json:"provider_message_id"`
	FailureReason     *string     `json:"failure_reason"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}
