package storage

import (
	"fmt"
	"log"

	"gemini-mcp/internal/common"
)

// NewStorage creates the appropriate storage backend based on configuration
func NewStorage(config *common.Config) (Storage, error) {
	// Use S3 only in HTTP mode when S3 is configured
	if config.S3Enabled {
		log.Printf("Initializing S3 storage (endpoint: %s, bucket: %s)", config.S3Endpoint, config.S3Bucket)
		return NewS3Storage(S3Config{
			Endpoint:        config.S3Endpoint,
			AccessKeyID:     config.S3AccessKeyID,
			SecretAccessKey: config.S3SecretAccessKey,
			Region:          config.S3Region,
			Bucket:          config.S3Bucket,
			UseSSL:          config.S3UseSSL,
			PresignTTL:      config.S3PresignTTL,
			ObjectTTL:       config.S3ObjectTTL,
			CleanupInterval: config.S3CleanupInterval,
		})
	}

	// Default to local storage
	log.Printf("Initializing local storage (directory: %s)", config.OutputDir)
	stor, err := NewLocalStorage(config.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create local storage: %w", err)
	}
	return stor, nil
}
