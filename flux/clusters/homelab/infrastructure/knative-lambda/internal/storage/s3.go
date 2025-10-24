// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	☁️ S3 STORAGE - AWS S3 storage implementation
//
//	🎯 Purpose: AWS S3 implementation of ObjectStorage interface
//	💡 Features: S3 object management with AWS SDK
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
	"github.com/aws/aws-sdk-go-v2/service/s3"

	apperrors "knative-lambda-new/internal/errors"
	"knative-lambda-new/internal/observability"
)

// ☁️ S3Storage - "AWS S3 storage implementation"
type S3Storage struct {
	client   *s3.Client
	region   string
	endpoint string
	obs      *observability.Observability
}

// 🏗️ S3StorageConfig - "Configuration for S3 storage"
type S3StorageConfig struct {
	Region        string
	Endpoint      string // Optional: custom endpoint for S3
	Observability *observability.Observability
}

// 🏗️ NewS3Storage - "Create new S3 storage client"
func NewS3Storage(ctx context.Context, cfg S3StorageConfig) (*S3Storage, error) {
	if cfg.Observability == nil {
		return nil, apperrors.NewConfigurationError("s3_storage", "observability", "observability cannot be nil")
	}

	ctx, span := cfg.Observability.StartSpan(ctx, "create_s3_storage")
	defer span.End()

	// Load AWS configuration
	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithRetryMode(aws.RetryModeAdaptive),
		config.WithRetryMaxAttempts(3),
	)
	if err != nil {
		return nil, apperrors.NewConfigurationError("s3_storage", "aws_config", fmt.Sprintf("failed to load AWS config: %v", err))
	}

	// Create S3 client with optional custom endpoint
	var s3Client *s3.Client
	if cfg.Endpoint != "" {
		s3Client = s3.NewFromConfig(awsConfig, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	} else {
		s3Client = s3.NewFromConfig(awsConfig)
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://s3.%s.amazonaws.com", cfg.Region)
	}

	cfg.Observability.Info(ctx, "Created S3 storage client",
		"region", cfg.Region,
		"endpoint", endpoint,
		"provider", ProviderS3)

	return &S3Storage{
		client:   s3Client,
		region:   cfg.Region,
		endpoint: endpoint,
		obs:      cfg.Observability,
	}, nil
}

// 📦 UploadObject - "Upload object to S3"
func (s *S3Storage) UploadObject(ctx context.Context, bucket, key string, reader io.Reader, contentType string, size int64) error {
	ctx, span := s.obs.StartSpanWithAttributes(ctx, "s3_upload_object", map[string]string{
		"storage.provider": string(ProviderS3),
		"s3.bucket":        bucket,
		"s3.key":           key,
		"s3.content_type":  contentType,
	})
	defer span.End()

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})

	if err != nil {
		return apperrors.WrapWithContext(err, fmt.Sprintf("bucket=%s, key=%s, size=%d", bucket, key, size))
	}

	s.obs.Info(ctx, "Successfully uploaded object to S3",
		"bucket", bucket,
		"key", key,
		"size", size,
		"content_type", contentType)

	return nil
}

// 📥 GetObject - "Get object from S3"
//
// ⚠️ IMPORTANT: Caller MUST close the returned io.ReadCloser to avoid resource leaks
// Example:
//
//	reader, meta, err := storage.GetObject(ctx, bucket, key)
//	if err != nil { return err }
//	defer reader.Close()
//	// ... use reader
func (s *S3Storage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, ObjectMetadata, error) {
	ctx, span := s.obs.StartSpanWithAttributes(ctx, "s3_get_object", map[string]string{
		"storage.provider": string(ProviderS3),
		"s3.bucket":        bucket,
		"s3.key":           key,
	})
	defer span.End()

	// First get object info
	headResult, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, ObjectMetadata{}, apperrors.WrapWithContext(err, fmt.Sprintf("bucket=%s, key=%s", bucket, key))
	}

	// Then get the object
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
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

	s.obs.Info(ctx, "Successfully retrieved object from S3",
		"bucket", bucket,
		"key", key,
		"size", metadata.Size)

	return result.Body, metadata, nil
}

// 🔍 ObjectExists - "Check if object exists in S3"
func (s *S3Storage) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	ctx, span := s.obs.StartSpanWithAttributes(ctx, "s3_object_exists", map[string]string{
		"storage.provider": string(ProviderS3),
		"s3.bucket":        bucket,
		"s3.key":           key,
	})
	defer span.End()

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
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

// 🗑️ DeleteObject - "Delete object from S3"
func (s *S3Storage) DeleteObject(ctx context.Context, bucket, key string) error {
	ctx, span := s.obs.StartSpanWithAttributes(ctx, "s3_delete_object", map[string]string{
		"storage.provider": string(ProviderS3),
		"s3.bucket":        bucket,
		"s3.key":           key,
	})
	defer span.End()

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return apperrors.WrapWithContext(err, fmt.Sprintf("bucket=%s, key=%s", bucket, key))
	}

	s.obs.Info(ctx, "Successfully deleted object from S3",
		"bucket", bucket,
		"key", key)

	return nil
}

// 🔧 GetProvider - "Get storage provider type"
func (s *S3Storage) GetProvider() StorageProvider {
	return ProviderS3
}

// 🔗 GetEndpoint - "Get S3 endpoint"
func (s *S3Storage) GetEndpoint() string {
	return s.endpoint
}

// 🪣 GetBucketURL - "Get S3 bucket URL for Kaniko context"
func (s *S3Storage) GetBucketURL(bucket, key string) string {
	return fmt.Sprintf("s3://%s/%s", bucket, key)
}

// 💚 HealthCheck - "Perform health check on S3 backend"
func (s *S3Storage) HealthCheck(ctx context.Context) error {
	ctx, span := s.obs.StartSpanWithAttributes(ctx, "s3_health_check", map[string]string{
		"storage.provider": string(ProviderS3),
		"s3.endpoint":      s.endpoint,
		"s3.region":        s.region,
	})
	defer span.End()

	// Use ListBuckets as a health check - it's a lightweight operation
	// that validates credentials and connectivity
	_, err := s.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		s.obs.Error(ctx, err, "S3 health check failed",
			"provider", ProviderS3,
			"endpoint", s.endpoint,
			"region", s.region)
		return apperrors.WrapWithContext(err, fmt.Sprintf("endpoint=%s, region=%s", s.endpoint, s.region))
	}

	s.obs.Info(ctx, "S3 health check passed",
		"provider", ProviderS3,
		"endpoint", s.endpoint,
		"region", s.region)

	return nil
}
