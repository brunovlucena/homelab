// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	💾 STORAGE CONFIGURATION - Storage provider configuration
//
//	🎯 Purpose: Configure storage provider (S3 or MinIO) for object storage
//	💡 Features: Multi-provider support, flexible switching, validation
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package config

import (
	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
)

// 💾 StorageConfig - "Storage provider configuration"
type StorageConfig struct {
	// Provider - Storage provider type ("aws-s3" or "minio")
	Provider string `envconfig:"STORAGE_PROVIDER"`

	// S3 Configuration
	S3 S3Config

	// MinIO Configuration
	MinIO MinIOConfig
}

// ☁️ S3Config - "AWS S3 storage configuration"
type S3Config struct {
	// Region - AWS region
	Region string `envconfig:"AWS_REGION"`

	// Endpoint - Custom S3 endpoint (optional)
	Endpoint string `envconfig:"S3_ENDPOINT"`

	// SourceBucket - Source code bucket
	SourceBucket string `envconfig:"S3_SOURCE_BUCKET"`

	// TempBucket - Temporary build context bucket
	TempBucket string `envconfig:"S3_TEMP_BUCKET"`
}

// 🏠 MinIOConfig - "MinIO storage configuration"
type MinIOConfig struct {
	// Endpoint - MinIO server endpoint (e.g., "minio.minio.svc.cluster.local:9000")
	Endpoint string `envconfig:"MINIO_ENDPOINT"`

	// AccessKey - MinIO access key
	AccessKey string `envconfig:"MINIO_ACCESS_KEY"`

	// SecretKey - MinIO secret key
	SecretKey string `envconfig:"MINIO_SECRET_KEY"`

	// UseSSL - Use HTTPS for MinIO connection
	UseSSL bool `envconfig:"MINIO_USE_SSL"`

	// Region - Region for S3 API compatibility (default: "us-east-1")
	Region string `envconfig:"MINIO_REGION"`

	// SourceBucket - Source code bucket
	SourceBucket string `envconfig:"MINIO_SOURCE_BUCKET"`

	// TempBucket - Temporary build context bucket
	TempBucket string `envconfig:"MINIO_TEMP_BUCKET"`
}

// 🔧 NewStorageConfig - "Create storage configuration with environment variables first, then fallback to constants"
func NewStorageConfig() *StorageConfig {
	provider := getEnv("STORAGE_PROVIDER", constants.StorageProviderDefault)

	return &StorageConfig{
		Provider: provider,
		S3: S3Config{
			Region:       getEnv("AWS_REGION", constants.AWSRegionDefault),
			Endpoint:     getEnv("S3_ENDPOINT", constants.S3EndpointDefault),
			SourceBucket: getEnv("S3_SOURCE_BUCKET", constants.S3SourceBucketDefault),
			TempBucket:   getEnv("S3_TEMP_BUCKET", constants.S3TempBucketDefault),
		},
		MinIO: MinIOConfig{
			Endpoint:     getEnv("MINIO_ENDPOINT", constants.MinIOEndpointDefault),
			AccessKey:    getEnv("MINIO_ACCESS_KEY", constants.MinIOAccessKeyDefault),
			SecretKey:    getEnv("MINIO_SECRET_KEY", constants.MinIOSecretKeyDefault),
			UseSSL:       getEnvBool("MINIO_USE_SSL", constants.MinIOUseSSLDefault),
			Region:       getEnv("MINIO_REGION", constants.MinIORegionDefault),
			SourceBucket: getEnv("MINIO_SOURCE_BUCKET", constants.MinIOSourceBucketDefault),
			TempBucket:   getEnv("MINIO_TEMP_BUCKET", constants.MinIOTempBucketDefault),
		},
	}
}

// 🔧 Validate - "Validate storage configuration"
func (c *StorageConfig) Validate() error {
	// Validate provider
	if c.Provider != "aws-s3" && c.Provider != "minio" {
		return errors.NewValidationError("storage_provider", c.Provider, "storage provider must be 'aws-s3' or 'minio'")
	}

	// Validate based on provider
	switch c.Provider {
	case "aws-s3":
		return c.validateS3Config()
	case "minio":
		return c.validateMinIOConfig()
	default:
		return errors.NewValidationError("storage_provider", c.Provider, "unsupported storage provider")
	}
}

// 🔧 validateS3Config - "Validate S3 configuration"
func (c *StorageConfig) validateS3Config() error {
	if c.S3.Region == "" {
		return errors.NewValidationError("s3_region", c.S3.Region, "S3 region is required when using AWS S3")
	}

	if c.S3.SourceBucket == "" {
		return errors.NewValidationError("s3_source_bucket", c.S3.SourceBucket, "S3 source bucket is required when using AWS S3")
	}

	if c.S3.TempBucket == "" {
		return errors.NewValidationError("s3_temp_bucket", c.S3.TempBucket, "S3 temp bucket is required when using AWS S3")
	}

	return nil
}

// 🔧 validateMinIOConfig - "Validate MinIO configuration"
func (c *StorageConfig) validateMinIOConfig() error {
	if c.MinIO.Endpoint == "" {
		return errors.NewValidationError("minio_endpoint", c.MinIO.Endpoint, "MinIO endpoint is required when using MinIO")
	}

	if c.MinIO.AccessKey == "" {
		return errors.NewValidationError("minio_access_key", c.MinIO.AccessKey, "MinIO access key is required when using MinIO")
	}

	if c.MinIO.SecretKey == "" {
		return errors.NewValidationError("minio_secret_key", c.MinIO.SecretKey, "MinIO secret key is required when using MinIO")
	}

	if c.MinIO.SourceBucket == "" {
		return errors.NewValidationError("minio_source_bucket", c.MinIO.SourceBucket, "MinIO source bucket is required when using MinIO")
	}

	if c.MinIO.TempBucket == "" {
		return errors.NewValidationError("minio_temp_bucket", c.MinIO.TempBucket, "MinIO temp bucket is required when using MinIO")
	}

	return nil
}

// 🔧 GetProvider - "Get storage provider"
func (c *StorageConfig) GetProvider() string {
	return c.Provider
}

// 🔧 GetSourceBucket - "Get source bucket name based on provider"
func (c *StorageConfig) GetSourceBucket() string {
	if c.Provider == "minio" {
		return c.MinIO.SourceBucket
	}
	return c.S3.SourceBucket
}

// 🔧 GetTempBucket - "Get temp bucket name based on provider"
func (c *StorageConfig) GetTempBucket() string {
	if c.Provider == "minio" {
		return c.MinIO.TempBucket
	}
	return c.S3.TempBucket
}

// 🔧 IsMinIO - "Check if MinIO is configured as provider"
func (c *StorageConfig) IsMinIO() bool {
	return c.Provider == "minio"
}

// 🔧 IsS3 - "Check if S3 is configured as provider"
func (c *StorageConfig) IsS3() bool {
	return c.Provider == "aws-s3"
}
