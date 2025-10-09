package storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"bruno-site/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient wraps the MinIO client
type MinIOClient struct {
	client *minio.Client
	bucket string
}

// NewMinIOClient creates a new MinIO client
func NewMinIOClient(cfg config.MinIOConfig) (*MinIOClient, error) {
	// Initialize MinIO client
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	log.Printf("📦 MinIO client initialized: %s", cfg.Endpoint)

	return &MinIOClient{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// GetObject retrieves an object from MinIO
func (m *MinIOClient) GetObject(ctx context.Context, objectName string) (io.ReadCloser, int64, string, error) {
	// Get object from MinIO
	object, err := m.client.GetObject(ctx, m.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to get object: %w", err)
	}

	// Get object info to retrieve size and content type
	info, err := object.Stat()
	if err != nil {
		object.Close()
		return nil, 0, "", fmt.Errorf("failed to get object info: %w", err)
	}

	return object, info.Size, info.ContentType, nil
}

// ObjectExists checks if an object exists in MinIO
func (m *MinIOClient) ObjectExists(ctx context.Context, objectName string) bool {
	_, err := m.client.StatObject(ctx, m.bucket, objectName, minio.StatObjectOptions{})
	return err == nil
}
