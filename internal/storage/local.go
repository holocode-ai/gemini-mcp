package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// LocalStorage implements Storage interface for local filesystem
type LocalStorage struct {
	baseDir string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &LocalStorage{baseDir: baseDir}, nil
}

// Store saves content to the local filesystem
func (s *LocalStorage) Store(ctx context.Context, data []byte, mimeType string, prefix string) (*StorageResult, error) {
	// Generate SHA256 hash of content
	hash := sha256.Sum256(data)
	contentHash := hex.EncodeToString(hash[:])

	// Determine file extension from MIME type
	ext := ExtensionFromMIME(mimeType)

	// Build filename with prefix and hash (first 16 chars)
	filename := fmt.Sprintf("%s_%s%s", prefix, contentHash[:16], ext)

	// Full path in base directory
	outputPath := filepath.Join(s.baseDir, filename)

	// Write file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &StorageResult{
		Location:    outputPath,
		ObjectKey:   filename,
		ContentHash: contentHash,
		MIMEType:    mimeType,
		Size:        int64(len(data)),
		ExpiresAt:   nil, // Local storage doesn't expire
	}, nil
}

// Retrieve returns the local file path for a given object key
// For local storage, no download is needed - just verify the file exists
func (s *LocalStorage) Retrieve(ctx context.Context, objectKey string) (string, func(), error) {
	filePath := filepath.Join(s.baseDir, objectKey)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("file not found: %s", objectKey)
	} else if err != nil {
		return "", nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// No cleanup needed for local storage
	cleanup := func() {}

	return filePath, cleanup, nil
}

// Delete removes a file from local storage
func (s *LocalStorage) Delete(ctx context.Context, objectKey string) error {
	filePath := filepath.Join(s.baseDir, objectKey)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// Close is a no-op for local storage
func (s *LocalStorage) Close() error {
	return nil
}

// IsRemote returns false for local storage
func (s *LocalStorage) IsRemote() bool {
	return false
}
