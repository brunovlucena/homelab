// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	💾 STORAGE INTERFACE - Storage abstraction for S3 and MinIO
//
//	🎯 Purpose: Define common interface for object storage operations
//	💡 Features: Support for both AWS S3 and MinIO with seamless switching
//
//	🏛️ ARCHITECTURE:
//	📦 Storage Operations - Upload, download, delete, existence checks
//	🔄 Provider Abstraction - S3-compatible API for both providers
//	⚙️ Configuration - Flexible provider selection
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package storage

import (
	"context"
	"io"
)

// 💾 StorageProvider - "Storage provider type"
type StorageProvider string

const (
	// ☁️ AWS S3 - "AWS S3 object storage"
	ProviderS3 StorageProvider = "aws-s3"
	// 🏠 MinIO - "MinIO S3-compatible object storage"
	ProviderMinIO StorageProvider = "minio"
)

// 📊 ObjectMetadata - "Object metadata information"
type ObjectMetadata struct {
	Size        int64
	ContentType string
	ETag        string
}

// 💾 ObjectStorage - "Common interface for object storage operations"
type ObjectStorage interface {
	// 📦 Upload Operations
	// UploadObject uploads an object to storage
	UploadObject(ctx context.Context, bucket, key string, reader io.Reader, contentType string, size int64) error

	// 📥 Download Operations
	// GetObject retrieves an object from storage
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, ObjectMetadata, error)

	// 🔍 Existence Checks
	// ObjectExists checks if an object exists in storage
	ObjectExists(ctx context.Context, bucket, key string) (bool, error)

	// 🗑️ Delete Operations
	// DeleteObject deletes an object from storage
	DeleteObject(ctx context.Context, bucket, key string) error

	// 🔧 Utility Operations
	// GetProvider returns the storage provider type
	GetProvider() StorageProvider

	// 🔗 GetEndpoint returns the storage endpoint URL
	GetEndpoint() string

	// 🪣 GetBucketURL returns the full bucket URL (for Kaniko context)
	GetBucketURL(bucket, key string) string
}

// 🏭 StorageFactory - "Factory for creating storage clients"
type StorageFactory interface {
	// CreateStorage creates a storage client based on provider
	CreateStorage(ctx context.Context, provider StorageProvider) (ObjectStorage, error)
}
