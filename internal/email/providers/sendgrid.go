package providers

import (
	"encoding/base64"
	"fmt"

	"github.com/maylng/backend/internal/email"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridProvider struct {
	client *sendgrid.Client
	apiKey string
}

func NewSendGridProvider(apiKey string) *SendGridProvider {
	return &SendGridProvider{
		client: sendgrid.NewSendClient(apiKey),
		apiKey: apiKey,
	}
}

func (p *SendGridProvider) SendEmail(emailMsg *email.Email) (*email.SendResult, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("SendGrid API key not configured")
	}

	// Create SendGrid message
	from := mail.NewEmail(emailMsg.FromName, emailMsg.FromEmail)

	// Create message
	var message *mail.SGMailV3

	if len(emailMsg.ToRecipients) == 1 {
		// Single recipient
		to := mail.NewEmail("", emailMsg.ToRecipients[0])
		message = mail.NewSingleEmail(from, emailMsg.Subject, to, emailMsg.TextContent, emailMsg.HTMLContent)
	} else {
		// Multiple recipients
		message = mail.NewV3Mail()
		message.SetFrom(from)
		message.Subject = emailMsg.Subject

		// Add content
		if emailMsg.TextContent != "" {
			message.AddContent(mail.NewContent("text/plain", emailMsg.TextContent))
		}
		if emailMsg.HTMLContent != "" {
			message.AddContent(mail.NewContent("text/html", emailMsg.HTMLContent))
		}

		// Add personalization for multiple recipients
		personalization := mail.NewPersonalization()
		for _, recipient := range emailMsg.ToRecipients {
			personalization.AddTos(mail.NewEmail("", recipient))
		}
		for _, recipient := range emailMsg.CcRecipients {
			personalization.AddCCs(mail.NewEmail("", recipient))
		}
		for _, recipient := range emailMsg.BccRecipients {
			personalization.AddBCCs(mail.NewEmail("", recipient))
		}
		message.AddPersonalizations(personalization)
	}

	// Add custom headers
	for key, value := range emailMsg.Headers {
		message.SetHeader(key, value)
	}

	// Add attachments
	for _, attachment := range emailMsg.Attachments {
		att := mail.NewAttachment()
		att.SetFilename(attachment.Filename)
		att.SetContent(base64.StdEncoding.EncodeToString(attachment.Content))
		att.SetType(attachment.ContentType)
		att.SetDisposition("attachment")
		message.AddAttachment(att)
	}

	// Send email
	response, err := p.client.Send(message)
	if err != nil {
		return &email.SendResult{
			Status:       "failed",
			ErrorMessage: err.Error(),
		}, err
	}

	// Check response status
	if response.StatusCode >= 400 {
		return &email.SendResult{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("SendGrid error: %d - %s", response.StatusCode, response.Body),
		}, fmt.Errorf("SendGrid error: %d", response.StatusCode)
	}

	// Extract message ID from headers
	messageID := ""
	if response.Headers != nil {
		if xMessageIds, exists := response.Headers["X-Message-Id"]; exists && len(xMessageIds) > 0 {
			messageID = xMessageIds[0]
		}
	}

	return &email.SendResult{
		MessageID:  messageID,
		ProviderID: "sendgrid",
		Status:     "sent",
	}, nil
}

func (p *SendGridProvider) GetDeliveryStatus(messageID string) (*email.DeliveryStatus, error) {
	// SendGrid doesn't provide a direct API to get delivery status by message ID
	// This would typically be handled through webhooks
	return &email.DeliveryStatus{
		MessageID: messageID,
		Status:    "unknown",
	}, nil
}
