package common

import (
	"fmt"
	"os"
	"strings"
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
}

func LoadConfig() *Config {
	config := &Config{
		APIKey:         os.Getenv("GOOGLE_API_KEY"),
		ProjectID:      os.Getenv("GOOGLE_PROJECT_ID"),
		Location:       getEnvOrDefault("GOOGLE_LOCATION", "us-central1"),
		Port:           getEnvOrDefault("PORT", "8080"),
		Transport:      getEnvOrDefault("TRANSPORT", "stdio"),
		OutputDir:      getEnvOrDefault("OUTPUT_DIR", "./output"),
		GenmediaBucket: os.Getenv("GENMEDIA_BUCKET"),
		ServiceTokens:  parseServiceTokens(os.Getenv("SERVICE_TOKENS")),
	}

	// Enable auth if tokens are configured
	config.AuthEnabled = len(config.ServiceTokens) > 0

	// Create output directory if it doesn't exist
	if config.OutputDir != "" {
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

func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("GOOGLE_API_KEY environment variable is required")
	}
	return nil
}
