package providers

import (
	"fmt"

	"github.com/maylng/backend/internal/email"
	"github.com/resend/resend-go/v2"
)

type ResendProvider struct {
	client *resend.Client
}

func NewResendProvider(apiKey string) *ResendProvider {
	client := resend.NewClient(apiKey)
	return &ResendProvider{
		client: client,
	}
}

func (p *ResendProvider) SendEmail(emailMsg *email.Email) (*email.SendResult, error) {
	if p.client == nil {
		return &email.SendResult{
			Status:       "failed",
			ErrorMessage: "Resend client not configured",
		}, fmt.Errorf("resend client not configured")
	}

	// Validate recipients
	if len(emailMsg.ToRecipients) == 0 {
		return &email.SendResult{
			Status:       "failed",
			ErrorMessage: "no recipients specified",
		}, fmt.Errorf("no recipients specified")
	}

	// Build from address
	fromAddress := emailMsg.FromEmail
	if emailMsg.FromName != "" {
		fromAddress = fmt.Sprintf("%s <%s>", emailMsg.FromName, emailMsg.FromEmail)
	}

	// Prepare email request
	params := &resend.SendEmailRequest{
		From:    fromAddress,
		To:      emailMsg.ToRecipients,
		Subject: emailMsg.Subject,
	}

	// Add optional fields
	if emailMsg.HTMLContent != "" {
		params.Html = emailMsg.HTMLContent
	}
	if emailMsg.TextContent != "" {
		params.Text = emailMsg.TextContent
	}
	if len(emailMsg.CcRecipients) > 0 {
		params.Cc = emailMsg.CcRecipients
	}
	if len(emailMsg.BccRecipients) > 0 {
		params.Bcc = emailMsg.BccRecipients
	}
	if len(emailMsg.Headers) > 0 {
		params.Headers = emailMsg.Headers
	}

	// Add attachments if any
	if len(emailMsg.Attachments) > 0 {
		for _, attachment := range emailMsg.Attachments {
			params.Attachments = append(params.Attachments, &resend.Attachment{
				Filename: attachment.Filename,
				Content:  attachment.Content,
			})
		}
	}

	// Send email
	sent, err := p.client.Emails.Send(params)
	if err != nil {
		return &email.SendResult{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to send email: %v", err),
		}, err
	}

	return &email.SendResult{
		MessageID:  sent.Id,
		ProviderID: "resend",
		Status:     "sent",
	}, nil
}

func (p *ResendProvider) GetDeliveryStatus(messageID string) (*email.DeliveryStatus, error) {
	if p.client == nil {
		return nil, fmt.Errorf("resend client not configured")
	}

	// Try to get email details using Resend SDK
	emailDetails, err := p.client.Emails.Get(messageID)
	if err != nil {
		return &email.DeliveryStatus{
			MessageID:    messageID,
			Status:       "unknown",
			ErrorMessage: fmt.Sprintf("failed to get email status: %v", err),
		}, err
	}

	// Map Resend status to our internal status
	status := "unknown"
	if emailDetails.LastEvent != "" {
		switch emailDetails.LastEvent {
		case "sent":
			status = "sent"
		case "delivered":
			status = "delivered"
		case "delivery_delayed":
			status = "delayed"
		case "complained":
			status = "complained"
		case "bounced":
			status = "bounced"
		case "opened", "clicked":
			status = "delivered" // If opened/clicked, it was delivered
		}
	}

	return &email.DeliveryStatus{
		MessageID: messageID,
		Status:    status,
	}, nil
}
