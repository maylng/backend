package providers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/maylng/backend/internal/email"
)

type SESProvider struct {
	client *sesv2.Client
	region string
}

func NewSESProvider(region string) (*SESProvider, error) {
	if region == "" {
		region = "us-east-1" // Default region
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sesv2.NewFromConfig(cfg)

	return &SESProvider{
		client: client,
		region: region,
	}, nil
}

func (p *SESProvider) SendEmail(emailMsg *email.Email) (*email.SendResult, error) {
	// Prepare recipients
	var destinations []string
	destinations = append(destinations, emailMsg.ToRecipients...)
	destinations = append(destinations, emailMsg.CcRecipients...)
	destinations = append(destinations, emailMsg.BccRecipients...)

	if len(destinations) == 0 {
		return &email.SendResult{
			Status:       "failed",
			ErrorMessage: "no recipients specified",
		}, fmt.Errorf("no recipients specified")
	}

	// Build email content
	content := &types.EmailContent{
		Simple: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(emailMsg.Subject),
				Charset: aws.String("UTF-8"),
			},
		},
	}

	// Add body content
	if emailMsg.TextContent != "" || emailMsg.HTMLContent != "" {
		body := &types.Body{}

		if emailMsg.TextContent != "" {
			body.Text = &types.Content{
				Data:    aws.String(emailMsg.TextContent),
				Charset: aws.String("UTF-8"),
			}
		}

		if emailMsg.HTMLContent != "" {
			body.Html = &types.Content{
				Data:    aws.String(emailMsg.HTMLContent),
				Charset: aws.String("UTF-8"),
			}
		}

		content.Simple.Body = body
	}

	// Handle attachments (if any)
	if len(emailMsg.Attachments) > 0 {
		// For attachments, we need to use Raw content instead of Simple
		rawContent, err := p.buildRawEmailContent(emailMsg)
		if err != nil {
			return &email.SendResult{
				Status:       "failed",
				ErrorMessage: fmt.Sprintf("failed to build raw email content: %v", err),
			}, err
		}

		content = &types.EmailContent{
			Raw: &types.RawMessage{
				Data: rawContent,
			},
		}
	}

	// Prepare destination
	destination := &types.Destination{
		ToAddresses: emailMsg.ToRecipients,
	}

	if len(emailMsg.CcRecipients) > 0 {
		destination.CcAddresses = emailMsg.CcRecipients
	}

	if len(emailMsg.BccRecipients) > 0 {
		destination.BccAddresses = emailMsg.BccRecipients
	}

	// Build from address
	fromAddress := emailMsg.FromEmail
	if emailMsg.FromName != "" {
		fromAddress = fmt.Sprintf("%s <%s>", emailMsg.FromName, emailMsg.FromEmail)
	}

	// Send email
	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(fromAddress),
		Destination:      destination,
		Content:          content,
	}

	// Add custom headers if no attachments (only supported in Simple content)
	if len(emailMsg.Attachments) == 0 && len(emailMsg.Headers) > 0 {
		// Note: SES v2 doesn't directly support custom headers in Simple content
		// Custom headers would need to be included in Raw content
		// For now, we'll log a warning and continue
		fmt.Printf("Warning: Custom headers not supported in SES Simple content mode\n")
	}

	resp, err := p.client.SendEmail(context.TODO(), input)
	if err != nil {
		return &email.SendResult{
			Status:       "failed",
			ErrorMessage: err.Error(),
		}, err
	}

	return &email.SendResult{
		MessageID:  aws.ToString(resp.MessageId),
		ProviderID: "ses",
		Status:     "sent",
	}, nil
}

func (p *SESProvider) GetDeliveryStatus(messageID string) (*email.DeliveryStatus, error) {
	// SES doesn't provide a direct API to get delivery status by message ID
	// This would typically be handled through SNS notifications or CloudWatch events
	return &email.DeliveryStatus{
		MessageID: messageID,
		Status:    "unknown",
	}, nil
}

// buildRawEmailContent builds raw email content for emails with attachments
func (p *SESProvider) buildRawEmailContent(emailMsg *email.Email) ([]byte, error) {
	boundary := "----=_NextPart_000_0000_01234567.89ABCDEF"

	var rawEmail strings.Builder

	// Email headers
	fromAddress := emailMsg.FromEmail
	if emailMsg.FromName != "" {
		fromAddress = fmt.Sprintf("%s <%s>", emailMsg.FromName, emailMsg.FromEmail)
	}

	rawEmail.WriteString(fmt.Sprintf("From: %s\r\n", fromAddress))
	rawEmail.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(emailMsg.ToRecipients, ", ")))

	if len(emailMsg.CcRecipients) > 0 {
		rawEmail.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(emailMsg.CcRecipients, ", ")))
	}

	rawEmail.WriteString(fmt.Sprintf("Subject: %s\r\n", emailMsg.Subject))
	rawEmail.WriteString("MIME-Version: 1.0\r\n")
	rawEmail.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))

	// Add custom headers
	for key, value := range emailMsg.Headers {
		rawEmail.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	rawEmail.WriteString("\r\n")

	// Email body
	rawEmail.WriteString(fmt.Sprintf("--%s\r\n", boundary))

	if emailMsg.HTMLContent != "" && emailMsg.TextContent != "" {
		// Multipart alternative for both text and HTML
		altBoundary := "----=_NextPart_Alt_000_0001_01234567.89ABCDEF"
		rawEmail.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", altBoundary))

		// Text part
		rawEmail.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
		rawEmail.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		rawEmail.WriteString(emailMsg.TextContent)
		rawEmail.WriteString("\r\n\r\n")

		// HTML part
		rawEmail.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
		rawEmail.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		rawEmail.WriteString(emailMsg.HTMLContent)
		rawEmail.WriteString("\r\n\r\n")

		rawEmail.WriteString(fmt.Sprintf("--%s--\r\n", altBoundary))
	} else if emailMsg.HTMLContent != "" {
		// HTML only
		rawEmail.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		rawEmail.WriteString(emailMsg.HTMLContent)
		rawEmail.WriteString("\r\n")
	} else {
		// Text only
		rawEmail.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		rawEmail.WriteString(emailMsg.TextContent)
		rawEmail.WriteString("\r\n")
	}

	// Attachments
	for _, attachment := range emailMsg.Attachments {
		rawEmail.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
		rawEmail.WriteString(fmt.Sprintf("Content-Type: %s\r\n", attachment.ContentType))
		rawEmail.WriteString("Content-Transfer-Encoding: base64\r\n")
		rawEmail.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", attachment.Filename))

		// Encode attachment content
		encoded := base64.StdEncoding.EncodeToString(attachment.Content)
		// Split into lines of 76 characters (RFC 2045)
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			rawEmail.WriteString(encoded[i:end])
			rawEmail.WriteString("\r\n")
		}
	}

	rawEmail.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	return []byte(rawEmail.String()), nil
}
