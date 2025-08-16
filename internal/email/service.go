package email

// TODO: Add mailgun as a provider
import (
	"fmt"
	"github.com/maylng/backend/internal/models"
)

type Provider interface {
	SendEmail(email *Email) (*SendResult, error)
	GetDeliveryStatus(messageID string) (*DeliveryStatus, error)
}

type Email struct {
	FromEmail     string
	FromName      string
	ToRecipients  []string
	CcRecipients  []string
	BccRecipients []string
	Subject       string
	TextContent   string
	HTMLContent   string
	Attachments   []Attachment
	Headers       map[string]string
}

type Attachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

type SendResult struct {
	MessageID    string
	ProviderID   string
	Status       string
	ErrorMessage string
}

type DeliveryStatus struct {
	MessageID    string
	Status       string
	DeliveredAt  *string
	ErrorMessage string
}

type Service struct {
	primary  Provider
	fallback Provider
}

func NewService(primary Provider, fallback Provider) *Service {
	return &Service{
		primary:  primary,
		fallback: fallback,
	}
}

func (s *Service) SendEmail(email *Email) (*SendResult, error) {
	// Try primary provider first
	if s.primary != nil {
		result, err := s.primary.SendEmail(email)
		if err == nil {
			return result, nil
		}
		fmt.Printf("Primary email provider failed: %v, trying fallback\n", err)
	}

	// Try fallback provider if primary fails
	if s.fallback != nil {
		return s.fallback.SendEmail(email)
	}

	return nil, fmt.Errorf("no email providers available")
}

func (s *Service) GetDeliveryStatus(messageID string) (*DeliveryStatus, error) {
	// Try primary provider first
	if s.primary != nil {
		return s.primary.GetDeliveryStatus(messageID)
	}

	// Try fallback provider
	if s.fallback != nil {
		return s.fallback.GetDeliveryStatus(messageID)
	}

	return nil, fmt.Errorf("no email providers available")
}

// ConvertFromSentEmail converts a SentEmail model to Email for sending
func ConvertFromSentEmail(sentEmail *models.SentEmail, fromEmailAddress string) *Email {
	email := &Email{
		FromEmail:     fromEmailAddress,
		ToRecipients:  sentEmail.ToRecipients,
		CcRecipients:  sentEmail.CcRecipients,
		BccRecipients: sentEmail.BccRecipients,
		Subject:       sentEmail.Subject,
		Headers:       make(map[string]string),
	}

	if sentEmail.TextContent != nil {
		email.TextContent = *sentEmail.TextContent
	}

	if sentEmail.HTMLContent != nil {
		email.HTMLContent = *sentEmail.HTMLContent
	}

	// Convert headers from metadata
	if sentEmail.Headers != nil {
		for key, value := range sentEmail.Headers {
			if strValue, ok := value.(string); ok {
				email.Headers[key] = strValue
			}
		}
	}

	return email
}
