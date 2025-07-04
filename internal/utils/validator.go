package utils

import (
	"regexp"
	"strings"
)

var (
	// Email regex pattern (simplified)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Domain regex pattern
	domainRegex = regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// ValidateEmail validates an email address format
func ValidateEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

// ValidateDomain validates a domain name format
func ValidateDomain(domain string) bool {
	if len(domain) > 253 {
		return false
	}
	return domainRegex.MatchString(domain)
}

// ValidateEmailPrefix validates an email prefix (local part)
func ValidateEmailPrefix(prefix string) bool {
	if len(prefix) == 0 || len(prefix) > 64 {
		return false
	}

	// Basic validation - no spaces, starts and ends with alphanumeric
	if strings.Contains(prefix, " ") {
		return false
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9]`).MatchString(prefix) {
		return false
	}

	if !regexp.MustCompile(`[a-zA-Z0-9]$`).MatchString(prefix) {
		return false
	}

	return true
}

// SanitizeString removes potentially harmful characters from strings
func SanitizeString(input string) string {
	// Remove null bytes and control characters
	input = strings.ReplaceAll(input, "\x00", "")
	input = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(input, "")

	// Trim whitespace
	return strings.TrimSpace(input)
}

// TruncateString truncates a string to a maximum length
func TruncateString(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}
	return input[:maxLength]
}
