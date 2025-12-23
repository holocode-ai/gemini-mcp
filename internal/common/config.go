package common

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Gemini API Configuration
	APIKey    string
	ProjectID string
	Location  string

	// Server Configuration
	Port           string
	Transport      string
	OutputDir      string
	GenmediaBucket string

	// Authentication Configuration
	ServiceTokens []string // Comma-separated list of valid Bearer tokens
	AuthEnabled   bool     // Whether authentication is required for HTTP transport

	// S3 Storage Configuration (HTTP mode only)
	S3Endpoint        string        // S3/MinIO endpoint (e.g., "minio:9000" or "s3.amazonaws.com")
	S3Bucket          string        // Bucket name for storing generated files
	S3Region          string        // AWS region (default: us-east-1)
	S3AccessKeyID     string        // Access key ID
	S3SecretAccessKey string        // Secret access key
	S3UseSSL          bool          // Use SSL/TLS for S3 connection (default: true)
	S3PresignTTL      time.Duration // TTL for presigned URLs (default: 24h)
	S3ObjectTTL       time.Duration // TTL for objects before auto-deletion (default: 24h)
	S3CleanupInterval time.Duration // Cleanup task interval (default: 1h)
	S3Enabled         bool          // Auto-enabled when S3 is configured in HTTP mode
}

func LoadConfig() *Config {
	config := &Config{
		APIKey:         os.Getenv("GOOGLE_API_KEY"),
		ProjectID:      os.Getenv("GOOGLE_PROJECT_ID"),
		Location:       getEnvOrDefault("GOOGLE_LOCATION", "us-central1"),
		Port:           getEnvOrDefault("PORT", "8080"),
		Transport:      getEnvOrDefault("TRANSPORT", "stdio"),
		OutputDir:      getEnvOrDefault("OUTPUT_DIR", "/tmp/gemini-mcp"),
		GenmediaBucket: os.Getenv("GENMEDIA_BUCKET"),
		ServiceTokens:  parseServiceTokens(os.Getenv("SERVICE_TOKENS")),

		// S3 configuration
		S3Endpoint:        os.Getenv("S3_ENDPOINT"),
		S3Bucket:          getEnvOrDefault("S3_BUCKET", "gemini-media"),
		S3Region:          getEnvOrDefault("S3_REGION", "us-east-1"),
		S3AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3UseSSL:          getEnvOrDefaultBool("S3_USE_SSL", true),
		S3PresignTTL:      getEnvOrDefaultDuration("S3_PRESIGN_TTL", 24*time.Hour),
		S3ObjectTTL:       getEnvOrDefaultDuration("S3_OBJECT_TTL", 24*time.Hour),
		S3CleanupInterval: getEnvOrDefaultDuration("S3_CLEANUP_INTERVAL", 1*time.Hour),
	}

	// Enable auth if tokens are configured
	config.AuthEnabled = len(config.ServiceTokens) > 0

	// Enable S3 if endpoint is configured and transport is HTTP
	config.S3Enabled = config.S3Endpoint != "" &&
		config.S3AccessKeyID != "" &&
		config.S3SecretAccessKey != "" &&
		(config.Transport == "http" || config.Transport == "sse")

	// Create output directory if it doesn't exist (for stdio mode or S3 disabled)
	if config.OutputDir != "" && !config.S3Enabled {
		if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
			fmt.Printf("Warning: Failed to create output directory: %v\n", err)
		}
	}

	return config
}

// parseServiceTokens parses a comma-separated list of tokens
func parseServiceTokens(tokensStr string) []string {
	if tokensStr == "" {
		return nil
	}
	tokens := strings.Split(tokensStr, ",")
	var result []string
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return parsed
	}
	return defaultValue
}

func getEnvOrDefaultDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			return defaultValue
		}
		return parsed
	}
	return defaultValue
}

func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("GOOGLE_API_KEY environment variable is required")
	}
	return nil
}

// UploadConfig is a minimal config for the upload_media CLI
type UploadConfig struct {
	S3Endpoint        string
	S3Bucket          string
	S3Region          string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3UseSSL          bool
	S3PresignTTL      time.Duration
	S3ObjectTTL       time.Duration
}

// LoadUploadConfig loads configuration for the upload_media CLI
func LoadUploadConfig() *UploadConfig {
	return &UploadConfig{
		S3Endpoint:        os.Getenv("S3_ENDPOINT"),
		S3Bucket:          getEnvOrDefault("S3_BUCKET", "gemini-media"),
		S3Region:          getEnvOrDefault("S3_REGION", "us-east-1"),
		S3AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3UseSSL:          getEnvOrDefaultBool("S3_USE_SSL", true),
		S3PresignTTL:      getEnvOrDefaultDuration("S3_PRESIGN_TTL", 24*time.Hour),
		S3ObjectTTL:       getEnvOrDefaultDuration("S3_OBJECT_TTL", 24*time.Hour),
	}
}

// Validate validates configuration for the upload CLI
func (c *UploadConfig) Validate() error {
	if c.S3Endpoint == "" {
		return fmt.Errorf("S3_ENDPOINT environment variable is required")
	}
	if c.S3AccessKeyID == "" {
		return fmt.Errorf("S3_ACCESS_KEY_ID environment variable is required")
	}
	if c.S3SecretAccessKey == "" {
		return fmt.Errorf("S3_SECRET_ACCESS_KEY environment variable is required")
	}
	return nil
}
