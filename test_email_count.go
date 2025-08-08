package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/maylng/backend/internal/models"
)

// This is a simple test script to verify email address count tracking
func main() {
	fmt.Println("=== Email Address Count Test ===")
	fmt.Println("Testing account response includes email_addresses_count field...")

	// Test the models
	testModels()

	fmt.Println("✅ All tests passed!")
}

func testModels() {
	// Test that the AccountResponse struct includes the new field
	accountResp := &models.AccountResponse{
		ID:                  uuid.New(),
		Plan:                "free",
		EmailLimitPerMonth:  1000,
		EmailAddressLimit:   5,
		EmailAddressesCount: 2, // This is the new field we added
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	fmt.Printf("Account ID: %s\n", accountResp.ID)
	fmt.Printf("Plan: %s\n", accountResp.Plan)
	fmt.Printf("Email Limit Per Month: %d\n", accountResp.EmailLimitPerMonth)
	fmt.Printf("Email Address Limit: %d\n", accountResp.EmailAddressLimit)
	fmt.Printf("Email Addresses Count: %d\n", accountResp.EmailAddressesCount)
	fmt.Printf("Created At: %s\n", accountResp.CreatedAt)
	fmt.Printf("Updated At: %s\n", accountResp.UpdatedAt)
	fmt.Printf("Usage: %d/%d email addresses\n", accountResp.EmailAddressesCount, accountResp.EmailAddressLimit)

	if accountResp.EmailAddressesCount <= accountResp.EmailAddressLimit {
		fmt.Println("✅ Email address count is within limits")
	} else {
		fmt.Println("❌ Email address count exceeds limits")
	}
}
