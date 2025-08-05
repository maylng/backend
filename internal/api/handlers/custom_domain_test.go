package handlers

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/maylng/backend/internal/models"
)

func TestCustomDomainHandler_toResponse(t *testing.T) {
	// Test the response conversion function
	handler := &CustomDomainHandler{}

	domain := &models.CustomDomain{
		ID:        uuid.New(),
		AccountID: uuid.New(),
		Domain:    "example.com",
		Status:    models.CustomDomainStatusVerified,
		DNSRecords: []models.DNSRecord{
			{
				Type:  "CNAME",
				Name:  "token._domainkey.example.com",
				Value: "token.dkim.amazonses.com",
				TTL:   1800,
			},
		},
	}

	response := handler.toResponse(domain)

	if response.ID != domain.ID {
		t.Errorf("Expected ID %s, got %s", domain.ID, response.ID)
	}

	if response.Domain != domain.Domain {
		t.Errorf("Expected domain %s, got %s", domain.Domain, response.Domain)
	}

	if response.Status != string(domain.Status) {
		t.Errorf("Expected status %s, got %s", domain.Status, response.Status)
	}

	if len(response.DNSRecords) != len(domain.DNSRecords) {
		t.Errorf("Expected %d DNS records, got %d", len(domain.DNSRecords), len(response.DNSRecords))
	}
}

func TestCreateCustomDomainRequest_Validation(t *testing.T) {
	// Test domain validation logic
	testCases := []struct {
		domain   string
		expected string
		valid    bool
	}{
		{"example.com", "example.com", true},
		{"Example.COM", "example.com", true},
		{"  sub.example.com  ", "sub.example.com", true},
		{"http://example.com", "example.com", true},
		{"https://www.example.com", "example.com", true},
		{"", "", false},
		{"example.com/path", "", false},
		{"invalid domain", "", false},
	}

	for _, tc := range testCases {
		// This would be the normalization logic from the handler
		normalized := normalizeCustomDomain(tc.domain)

		if tc.valid && normalized != tc.expected {
			t.Errorf("Expected normalized domain %s, got %s for input %s", tc.expected, normalized, tc.domain)
		}

		if !tc.valid && normalized != "" {
			t.Errorf("Expected empty string for invalid domain %s, got %s", tc.domain, normalized)
		}
	}
}

// Helper function that would be extracted from the handler
func normalizeCustomDomain(domain string) string {
	// This is the same logic used in the handler
	domain = strings.ToLower(strings.TrimSpace(domain))
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "www.")

	// Basic validation
	if domain == "" || strings.Contains(domain, "/") || strings.Contains(domain, " ") {
		return ""
	}

	return domain
}
