// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🏠 MINIO STORAGE - MinIO S3-compatible storage implementation
//
//	🎯 Purpose: MinIO implementation of ObjectStorage interface
//	💡 Features: S3-compatible API for on-premises object storage
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	apperrors "knative-lambda-new/internal/errors"
	"knative-lambda-new/internal/observability"
)

// 🏠 MinIOStorage - "MinIO S3-compatible storage implementation"
type MinIOStorage struct {
	client    *s3.Client
	endpoint  string
	accessKey string
	secretKey string
	useSSL    bool
	obs       *observability.Observability
}

// 🏗️ MinIOStorageConfig - "Configuration for MinIO storage"
type MinIOStorageConfig struct {
	Endpoint      string // MinIO endpoint (e.g., "minio.minio.svc.cluster.local:9000")
	AccessKey     string // MinIO access key
	SecretKey     string // MinIO secret key
	UseSSL        bool   // Use HTTPS (default: false for internal cluster access)
	Region        string // Region for S3 API (default: "us-east-1" for MinIO)
	Observability *observability.Observability
}

// 🏗️ NewMinIOStorage - "Create new MinIO storage client"
func NewMinIOStorage(ctx context.Context, cfg MinIOStorageConfig) (*MinIOStorage, error) {
	if cfg.Observability == nil {
		return nil, apperrors.NewConfigurationError("minio_storage", "observability", "observability cannot be nil")
	}

	ctx, span := cfg.Observability.StartSpan(ctx, "create_minio_storage")
	defer span.End()

	// Validate configuration
	if cfg.Endpoint == "" {
		return nil, apperrors.NewConfigurationError("minio_storage", "endpoint", "MinIO endpoint is required")
	}

	if cfg.AccessKey == "" {
		return nil, apperrors.NewConfigurationError("minio_storage", "access_key", "MinIO access key is required")
	}

	if cfg.SecretKey == "" {
		return nil, apperrors.NewConfigurationError("minio_storage", "secret_key", "MinIO secret key is required")
	}

	// Default region for MinIO
	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	// Create AWS config with MinIO credentials
	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"", // session token (not used for MinIO)
		)),
		config.WithRetryMode(aws.RetryModeAdaptive),
		config.WithRetryMaxAttempts(3),
	)
	if err != nil {
		return nil, apperrors.NewConfigurationError("minio_storage", "aws_config", fmt.Sprintf("failed to load AWS config for MinIO: %v", err))
	}

	// Create endpoint URL with proper scheme
	endpointScheme := "http"
	if cfg.UseSSL {
		endpointScheme = "https"
	}
	endpointURL := fmt.Sprintf("%s://%s", endpointScheme, cfg.Endpoint)

	// Create S3 client with MinIO endpoint
	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpointURL)
		o.UsePathStyle = true // MinIO requires path-style addressing
	})

	cfg.Observability.Info(ctx, "Created MinIO storage client",
		"endpoint", endpointURL,
		"region", region,
		"use_ssl", cfg.UseSSL,
		"provider", ProviderMinIO)

	return &MinIOStorage{
		client:    s3Client,
		endpoint:  endpointURL,
		accessKey: cfg.AccessKey,
		secretKey: cfg.SecretKey,
		useSSL:    cfg.UseSSL,
		obs:       cfg.Observability,
	}, nil
}

// 📦 UploadObject - "Upload object to MinIO"
func (m *MinIOStorage) UploadObject(ctx context.Context, bucket, key string, reader io.Reader, contentType string, size int64) error {
	ctx, span := m.obs.StartSpanWithAttributes(ctx, "minio_upload_object", map[string]string{
		"storage.provider":   string(ProviderMinIO),
		"minio.bucket":       bucket,
		"minio.key":          key,
		"minio.content_type": contentType,
	})
	defer span.End()

	_, err := m.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})

	if err != nil {
		return apperrors.WrapWithContext(err, fmt.Sprintf("bucket=%s, key=%s, size=%d", bucket, key, size))
	}

	m.obs.Info(ctx, "Successfully uploaded object to MinIO",
		"bucket", bucket,
		"key", key,
		"size", size,
		"content_type", contentType)

	return nil
}

// 📥 GetObject - "Get object from MinIO"
//
// ⚠️ IMPORTANT: Caller MUST close the returned io.ReadCloser to avoid resource leaks
// Example:
//
//	reader, meta, err := storage.GetObject(ctx, bucket, key)
//	if err != nil { return err }
//	defer reader.Close()
//	// ... use reader
func (m *MinIOStorage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, ObjectMetadata, error) {
	ctx, span := m.obs.StartSpanWithAttributes(ctx, "minio_get_object", map[string]string{
		"storage.provider": string(ProviderMinIO),
		"minio.bucket":     bucket,
		"minio.key":        key,
	})
	defer span.End()

	// First get object info
	headResult, err := m.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, ObjectMetadata{}, apperrors.WrapWithContext(err, fmt.Sprintf("bucket=%s, key=%s", bucket, key))
	}

	// Then get the object
	result, err := m.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, ObjectMetadata{}, apperrors.WrapWithContext(err, fmt.Sprintf("bucket=%s, key=%s", bucket, key))
	}

	metadata := ObjectMetadata{
		Size:        aws.ToInt64(headResult.ContentLength),
		ContentType: aws.ToString(headResult.ContentType),
		ETag:        aws.ToString(headResult.ETag),
	}

	m.obs.Info(ctx, "Successfully retrieved object from MinIO",
		"bucket", bucket,
		"key", key,
		"size", metadata.Size)

	return result.Body, metadata, nil
}

// 🔍 ObjectExists - "Check if object exists in MinIO"
func (m *MinIOStorage) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	ctx, span := m.obs.StartSpanWithAttributes(ctx, "minio_object_exists", map[string]string{
		"storage.provider": string(ProviderMinIO),
		"minio.bucket":     bucket,
		"minio.key":        key,
	})
	defer span.End()

	_, err := m.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		// Use proper error type checking instead of string matching
		var apiErr interface{ ErrorCode() string }
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "NotFound", "NoSuchKey", "NoSuchBucket":
				return false, nil
			}
		}
		return false, apperrors.WrapWithContext(err, fmt.Sprintf("bucket=%s, key=%s", bucket, key))
	}

	return true, nil
}

// 🗑️ DeleteObject - "Delete object from MinIO"
func (m *MinIOStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	ctx, span := m.obs.StartSpanWithAttributes(ctx, "minio_delete_object", map[string]string{
		"storage.provider": string(ProviderMinIO),
		"minio.bucket":     bucket,
		"minio.key":        key,
	})
	defer span.End()

	_, err := m.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return apperrors.WrapWithContext(err, fmt.Sprintf("bucket=%s, key=%s", bucket, key))
	}

	m.obs.Info(ctx, "Successfully deleted object from MinIO",
		"bucket", bucket,
		"key", key)

	return nil
}

// 🔧 GetProvider - "Get storage provider type"
func (m *MinIOStorage) GetProvider() StorageProvider {
	return ProviderMinIO
}

// 🔗 GetEndpoint - "Get MinIO endpoint"
func (m *MinIOStorage) GetEndpoint() string {
	return m.endpoint
}

// 🪣 GetBucketURL - "Get MinIO bucket URL for Kaniko context"
// For MinIO, we need to use a custom S3 endpoint format that Kaniko understands
func (m *MinIOStorage) GetBucketURL(bucket, key string) string {
	// For MinIO with Kaniko, we still use s3:// prefix
	// but Kaniko needs to be configured with custom endpoint via environment variables
	return fmt.Sprintf("s3://%s/%s", bucket, key)
}

// 💚 HealthCheck - "Perform health check on MinIO backend"
func (m *MinIOStorage) HealthCheck(ctx context.Context) error {
	ctx, span := m.obs.StartSpanWithAttributes(ctx, "minio_health_check", map[string]string{
		"storage.provider": string(ProviderMinIO),
		"minio.endpoint":   m.endpoint,
		"minio.use_ssl":    fmt.Sprintf("%t", m.useSSL),
	})
	defer span.End()

	// Use ListBuckets as a health check - it's a lightweight operation
	// that validates credentials and connectivity
	_, err := m.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		m.obs.Error(ctx, err, "MinIO health check failed",
			"provider", ProviderMinIO,
			"endpoint", m.endpoint,
			"use_ssl", m.useSSL)
		return apperrors.WrapWithContext(err, fmt.Sprintf("MinIO health check failed: endpoint=%s, use_ssl=%t", m.endpoint, m.useSSL))
	}

	m.obs.Info(ctx, "MinIO health check passed",
		"provider", ProviderMinIO,
		"endpoint", m.endpoint,
		"use_ssl", m.useSSL)

	return nil
}
