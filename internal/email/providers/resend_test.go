package providers

import (
	"strings"
	"testing"

	"github.com/maylng/backend/internal/email"
)

func TestResendProvider_NewResendProvider(t *testing.T) {
	apiKey := "test-api-key"
	provider := NewResendProvider(apiKey)

	if provider == nil {
		t.Fatal("NewResendProvider returned nil")
	}

	if provider.client == nil {
		t.Error("Resend client is nil")
	}
}

func TestResendProvider_SendEmail_NoAPIKey(t *testing.T) {
	provider := NewResendProvider("")

	testEmail := &email.Email{
		FromEmail:    "test@example.com",
		ToRecipients: []string{"recipient@example.com"},
		Subject:      "Test Email",
		TextContent:  "This is a test email",
	}

	result, err := provider.SendEmail(testEmail)

	if err == nil {
		t.Error("Expected error when API key is empty")
	}

	if result.Status != "failed" {
		t.Errorf("Expected status 'failed', got %s", result.Status)
	}

	// The Resend SDK validates API key during the call, so the error message will be about invalid API key
	if !strings.Contains(result.ErrorMessage, "API key is invalid") {
		t.Errorf("Expected error message to contain 'API key is invalid', got %s", result.ErrorMessage)
	}
}

func TestResendProvider_SendEmail_NoRecipients(t *testing.T) {
	provider := NewResendProvider("test-api-key")

	testEmail := &email.Email{
		FromEmail:   "test@example.com",
		Subject:     "Test Email",
		TextContent: "This is a test email",
		// No recipients
	}

	result, err := provider.SendEmail(testEmail)

	if err == nil {
		t.Error("Expected error when no recipients are specified")
	}

	if result.Status != "failed" {
		t.Errorf("Expected status 'failed', got %s", result.Status)
	}

	if result.ErrorMessage != "no recipients specified" {
		t.Errorf("Expected error message 'no recipients specified', got %s", result.ErrorMessage)
	}
}

func TestResendProvider_GetDeliveryStatus_NoAPIKey(t *testing.T) {
	provider := NewResendProvider("")

	_, err := provider.GetDeliveryStatus("test-message-id")

	if err == nil {
		t.Error("Expected error when API key is empty")
	}

	// The Resend SDK validates API key during the call, so check for API key error
	expectedSubstring := "API key is invalid"
	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedSubstring, err.Error())
	}
}
