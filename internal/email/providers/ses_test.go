package providers

import (
	"testing"

	"github.com/maylng/backend/internal/email"
)

func TestSESProviderImplementsInterface(t *testing.T) {
	// Test that SESProvider implements the Provider interface
	var _ email.Provider = (*SESProvider)(nil)
}

func TestNewSESProvider(t *testing.T) {
	// Test creating a new SES provider (without actual AWS credentials)
	provider, err := NewSESProvider("us-east-1")
	if err != nil {
		t.Fatalf("Expected to create SES provider, got error: %v", err)
	}

	if provider == nil {
		t.Fatal("Expected non-nil provider")
	}

	if provider.region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got '%s'", provider.region)
	}
}

func TestNewSESProviderDefaultRegion(t *testing.T) {
	// Test creating a new SES provider with empty region (should default to us-east-1)
	provider, err := NewSESProvider("")
	if err != nil {
		t.Fatalf("Expected to create SES provider, got error: %v", err)
	}

	if provider.region != "us-east-1" {
		t.Errorf("Expected default region 'us-east-1', got '%s'", provider.region)
	}
}
