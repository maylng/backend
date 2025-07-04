package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateAPIKey generates a new API key
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return "maylng_" + hex.EncodeToString(bytes), nil
}

// HashAPIKey hashes an API key for storage
func HashAPIKey(apiKey, salt string) string {
	hash := sha256.Sum256([]byte(apiKey + salt))
	return hex.EncodeToString(hash[:])
}

// ValidateAPIKey checks if the provided API key matches the hash
func ValidateAPIKey(apiKey, hash, salt string) bool {
	return HashAPIKey(apiKey, salt) == hash
}

// ExtractAPIKey extracts API key from Authorization header
func ExtractAPIKey(authHeader string) (string, error) {
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return authHeader[7:], nil
}
