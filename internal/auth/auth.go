package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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

// EncryptString encrypts a string using AES-GCM
func EncryptString(plaintext, key string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Create a 32-byte key from the provided key using SHA-256
	hasher := sha256.New()
	hasher.Write([]byte(key))
	keyBytes := hasher.Sum(nil)

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// DecryptString decrypts a string using AES-GCM
func DecryptString(ciphertext, key string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Create a 32-byte key from the provided key using SHA-256
	hasher := sha256.New()
	hasher.Write([]byte(key))
	keyBytes := hasher.Sum(nil)

	data, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext_bytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext_bytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
