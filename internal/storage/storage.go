package storage

import (
	"context"
	"time"
)

// StorageResult represents the result of a storage operation
type StorageResult struct {
	// Location is the access URL/path for the stored content
	// For stdio mode: local file path
	// For HTTP mode with S3: presigned URL
	Location string

	// ObjectKey is the storage path (e.g., "2024/12/23/gemini_image_abc123.png")
	ObjectKey string

	// ContentHash is the SHA256 hash of the content (first 16 chars used in filename)
	ContentHash string

	// ExpiresAt is the expiration time for presigned URL (nil for local storage)
	ExpiresAt *time.Time

	// MIMEType is the content type (e.g., "image/png", "video/mp4")
	MIMEType string

	// Size is the content size in bytes
	Size int64
}

// Storage defines the interface for storing generated content
type Storage interface {
	// Store saves content and returns the storage result
	// - data: the raw bytes to store
	// - mimeType: content type (e.g., "image/png", "video/mp4")
	// - prefix: prefix for the filename (e.g., "gemini_image", "veo_video")
	Store(ctx context.Context, data []byte, mimeType string, prefix string) (*StorageResult, error)

	// Delete removes an object by its key
	Delete(ctx context.Context, objectKey string) error

	// Close cleans up any resources (stops cleanup goroutines, etc.)
	Close() error

	// IsRemote returns true if storage is remote (S3), false for local
	IsRemote() bool
}

// extensionFromMIME returns the file extension for a given MIME type
func ExtensionFromMIME(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "application/json":
		return ".json"
	default:
		return ""
	}
}
