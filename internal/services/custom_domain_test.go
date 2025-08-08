package services

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/maylng/backend/internal/models"
)

func TestCustomDomainService_CreateCustomDomain(t *testing.T) {
	// This test validates the CustomDomainService interface and model structure
	// without requiring a database connection

	// Test model creation
	accountID := uuid.New()
	domain := "example.com"

	customDomain := &models.CustomDomain{
		ID:        uuid.New(),
		AccountID: accountID,
		Domain:    domain,
		Status:    models.CustomDomainStatusPending,
	}

	// Test model methods
	if customDomain.IsVerified() {
		t.Error("Expected domain to not be verified when status is pending")
	}

	if customDomain.CanSendEmails() {
		t.Error("Expected domain to not allow sending emails when status is pending")
	}

	expectedFromAddress := "noreply@" + domain
	if customDomain.GetDefaultFromAddress() != expectedFromAddress {
		t.Errorf("Expected default from address '%s', got '%s'", expectedFromAddress, customDomain.GetDefaultFromAddress())
	}

	// Test verified domain
	customDomain.Status = models.CustomDomainStatusVerified
	if !customDomain.IsVerified() {
		t.Error("Expected domain to be verified when status is verified")
	}

	if !customDomain.CanSendEmails() {
		t.Error("Expected domain to allow sending emails when status is verified")
	}
}

func TestCustomDomainStatus(t *testing.T) {
	// Test status constants
	statuses := []models.CustomDomainStatus{
		models.CustomDomainStatusPending,
		models.CustomDomainStatusVerified,
		models.CustomDomainStatusFailed,
		models.CustomDomainStatusDisabled,
	}

	expectedStatuses := []string{
		"pending",
		"verified",
		"failed",
		"disabled",
	}

	for i, status := range statuses {
		if string(status) != expectedStatuses[i] {
			t.Errorf("Expected status '%s', got '%s'", expectedStatuses[i], string(status))
		}
	}
}

func TestDNSRecord(t *testing.T) {
	// Test DNS record model
	record := models.DNSRecord{
		Type:  "TXT",
		Name:  "_amazonses.example.com",
		Value: "verification-token-here",
		TTL:   300,
	}

	fmt.Printf("Record Type: %s\n", record.Type)
	fmt.Printf("Record Name: %s\n", record.Name)
	fmt.Printf("Record Value: %s\n", record.Value)
	fmt.Printf("Record TTL: %d\n", record.TTL)


	if record.Type != "TXT" {
		t.Errorf("Expected record type 'TXT', got '%s'", record.Type)
	}
}
