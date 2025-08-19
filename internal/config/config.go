package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL            string
	RedisURL               string
	GinMode                string
	SendGridAPIKey         string
	ResendAPIKey           string
	AWSRegion              string
	JWTSecret              string
	APIKeyHashSalt         string
	Environment            string
	LogLevel               string
	Port                   string
	MaxEmailsPerHour       int
	DefaultDomain          string
	EmailProvider          string // "resend", "sendgrid", or "ses"
	TPSEncryptionKey       string // For encrypting TPS API keys and passwords
	BROWSERBASE_API_KEY    string
	BROWSERBASE_PROJECT_ID string
	// PlatformCreationToken is a shared secret used to allow platform-origin requests to create accounts
	PlatformCreationToken string
}

func Load() *Config {
	return &Config{
		DatabaseURL:            getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/maylng?sslmode=disable"),
		RedisURL:               getEnv("REDIS_URL", "redis://localhost:6379"),
		GinMode:                getEnv("GIN_MODE", "debug"),
		SendGridAPIKey:         getEnv("SENDGRID_API_KEY", ""),
		ResendAPIKey:           getEnv("RESEND_API_KEY", ""),
		AWSRegion:              getEnv("AWS_REGION", "us-east-1"),
		JWTSecret:              getEnv("JWT_SECRET", "your-secret-key"),
		APIKeyHashSalt:         getEnv("API_KEY_HASH_SALT", "your-salt"),
		Environment:            getEnv("ENVIRONMENT", "development"),
		LogLevel:               getEnv("LOG_LEVEL", "info"),
		Port:                   getEnv("PORT", "8080"),
		MaxEmailsPerHour:       getEnvAsInt("MAX_EMAILS_PER_HOUR", 100),
		DefaultDomain:          getEnv("DEFAULT_DOMAIN", "mayl.ng"),
		EmailProvider:          getEnv("EMAIL_PROVIDER", "resend"), // Default to resend
		TPSEncryptionKey:       getEnv("TPS_ENCRYPTION_KEY", "default-key-change-in-production"),
		BROWSERBASE_API_KEY:    getEnv("BROWSERBASE_API_KEY", "your_browserbase_api_key_here"),
		BROWSERBASE_PROJECT_ID: getEnv("BROWSERBASE_PROJECT_ID", "your_browserbase_project_id_here"),
		PlatformCreationToken:  getEnv("PLATFORM_CREATION_TOKEN", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
