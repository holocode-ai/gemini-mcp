package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Storage implements Storage interface for S3/MinIO
type S3Storage struct {
	client          *minio.Client
	bucket          string
	presignTTL      time.Duration
	objectTTL       time.Duration
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// S3Config holds S3 storage configuration
type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Bucket          string
	UseSSL          bool
	PresignTTL      time.Duration
	ObjectTTL       time.Duration
	CleanupInterval time.Duration
}

// parseEndpoint extracts host:port from an endpoint that may include a protocol
func parseEndpoint(endpoint string, defaultUseSSL bool) (host string, useSSL bool) {
	useSSL = defaultUseSSL

	// Check if endpoint has a scheme
	if parsed, err := url.Parse(endpoint); err == nil && parsed.Scheme != "" {
		// Has scheme (http:// or https://)
		useSSL = parsed.Scheme == "https"
		host = parsed.Host
		if host == "" {
			// Fallback if parsing didn't work as expected
			host = endpoint
		}
	} else {
		// No scheme, use as-is
		host = endpoint
	}

	return host, useSSL
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	// Parse endpoint to extract host:port and detect SSL from scheme
	endpoint, useSSL := parseEndpoint(cfg.Endpoint, cfg.UseSSL)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: useSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	ctx := context.Background()

	// Check if bucket exists, create if not
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		log.Printf("Bucket %s does not exist, creating...", cfg.Bucket)
		err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{
			Region: cfg.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket %s: %w", cfg.Bucket, err)
		}
		log.Printf("Bucket %s created successfully", cfg.Bucket)
	}

	s := &S3Storage{
		client:          client,
		bucket:          cfg.Bucket,
		presignTTL:      cfg.PresignTTL,
		objectTTL:       cfg.ObjectTTL,
		cleanupInterval: cfg.CleanupInterval,
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup routine
	go s.startCleanupRoutine()

	return s, nil
}

// Store saves content to S3 and returns a presigned URL
func (s *S3Storage) Store(ctx context.Context, data []byte, mimeType string, prefix string) (*StorageResult, error) {
	// Generate SHA256 hash of content
	hash := sha256.Sum256(data)
	contentHash := hex.EncodeToString(hash[:])

	// Build date-organized path: YYYY/MM/DD/prefix_hash.ext
	now := time.Now().UTC()
	datePath := now.Format("2006/01/02")
	ext := ExtensionFromMIME(mimeType)
	filename := fmt.Sprintf("%s_%s%s", prefix, contentHash[:16], ext)
	objectKey := fmt.Sprintf("%s/%s", datePath, filename)

	// Upload to S3
	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: mimeType,
		UserMetadata: map[string]string{
			"created-at": now.Format(time.RFC3339),
			"expires-at": now.Add(s.objectTTL).Format(time.RFC3339),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Generate presigned URL
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, s.presignTTL, url.Values{})
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	expiresAt := now.Add(s.presignTTL)

	return &StorageResult{
		Location:    presignedURL.String(),
		ObjectKey:   objectKey,
		ContentHash: contentHash,
		MIMEType:    mimeType,
		Size:        int64(len(data)),
		ExpiresAt:   &expiresAt,
	}, nil
}

// Delete removes an object from S3
func (s *S3Storage) Delete(ctx context.Context, objectKey string) error {
	err := s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// Close stops the cleanup routine
func (s *S3Storage) Close() error {
	close(s.stopCleanup)
	return nil
}

// IsRemote returns true for S3 storage
func (s *S3Storage) IsRemote() bool {
	return true
}

// startCleanupRoutine periodically cleans up expired objects
func (s *S3Storage) startCleanupRoutine() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	log.Printf("S3 cleanup routine started (interval: %v, TTL: %v)", s.cleanupInterval, s.objectTTL)

	for {
		select {
		case <-ticker.C:
			s.cleanupExpiredObjects()
		case <-s.stopCleanup:
			log.Println("S3 cleanup routine stopped")
			return
		}
	}
}

// cleanupExpiredObjects removes objects that have exceeded their TTL
func (s *S3Storage) cleanupExpiredObjects() {
	ctx := context.Background()
	now := time.Now().UTC()
	deletedCount := 0
	errorCount := 0

	// List all objects in the bucket
	objectCh := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("Error listing objects: %v", object.Err)
			errorCount++
			continue
		}

		// Check if object is older than TTL based on LastModified
		age := now.Sub(object.LastModified)
		if age > s.objectTTL {
			err := s.client.RemoveObject(ctx, s.bucket, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				log.Printf("Failed to delete expired object %s: %v", object.Key, err)
				errorCount++
			} else {
				deletedCount++
			}
		}
	}

	if deletedCount > 0 || errorCount > 0 {
		log.Printf("S3 cleanup completed: deleted %d objects, %d errors", deletedCount, errorCount)
	}
}
