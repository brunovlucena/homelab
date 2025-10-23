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
	"fmt"
	"regexp"

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

	// Validate source bucket name format
	if err := validateBucketName(c.S3.SourceBucket); err != nil {
		return errors.NewValidationError("s3_source_bucket", c.S3.SourceBucket, err.Error())
	}

	if c.S3.TempBucket == "" {
		return errors.NewValidationError("s3_temp_bucket", c.S3.TempBucket, "S3 temp bucket is required when using AWS S3")
	}

	// Validate temp bucket name format
	if err := validateBucketName(c.S3.TempBucket); err != nil {
		return errors.NewValidationError("s3_temp_bucket", c.S3.TempBucket, err.Error())
	}

	return nil
}

// 🔧 validateMinIOConfig - "Validate MinIO configuration"
func (c *StorageConfig) validateMinIOConfig() error {
	if c.MinIO.Endpoint == "" {
		return errors.NewValidationError("minio_endpoint", c.MinIO.Endpoint, "MinIO endpoint is required when using MinIO")
	}

	// Validate endpoint format includes port
	if !regexp.MustCompile(`:\d+$`).MatchString(c.MinIO.Endpoint) {
		return errors.NewValidationError("minio_endpoint", c.MinIO.Endpoint, "MinIO endpoint must include port (e.g., 'host:9000')")
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

	// Validate source bucket name format
	if err := validateBucketName(c.MinIO.SourceBucket); err != nil {
		return errors.NewValidationError("minio_source_bucket", c.MinIO.SourceBucket, err.Error())
	}

	if c.MinIO.TempBucket == "" {
		return errors.NewValidationError("minio_temp_bucket", c.MinIO.TempBucket, "MinIO temp bucket is required when using MinIO")
	}

	// Validate temp bucket name format
	if err := validateBucketName(c.MinIO.TempBucket); err != nil {
		return errors.NewValidationError("minio_temp_bucket", c.MinIO.TempBucket, err.Error())
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

// 🔧 validateBucketName - "Validate S3/MinIO bucket name follows naming rules"
//
// S3 bucket naming rules:
// - Must be between 3 and 63 characters long
// - Can consist only of lowercase letters, numbers, dots (.), and hyphens (-)
// - Must begin and end with a letter or number
// - Must not be formatted as an IP address (e.g., 192.168.5.4)
// - Must not start with 'xn--' (used for punycode)
// - Must not end with '-s3alias' (reserved)
func validateBucketName(name string) error {
	if len(name) < 3 || len(name) > 63 {
		return fmt.Errorf("bucket name must be between 3 and 63 characters, got %d", len(name))
	}

	// Check valid characters and start/end requirements
	validBucketName := regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*[a-z0-9]$`)
	if !validBucketName.MatchString(name) {
		return fmt.Errorf("bucket name must start and end with lowercase letter or number, and contain only lowercase letters, numbers, dots, and hyphens")
	}

	// Check for consecutive dots
	if regexp.MustCompile(`\.\.`).MatchString(name) {
		return fmt.Errorf("bucket name must not contain consecutive dots")
	}

	// Check for IP address format
	ipAddress := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if ipAddress.MatchString(name) {
		return fmt.Errorf("bucket name must not be formatted as an IP address")
	}

	// Check reserved prefixes/suffixes
	if regexp.MustCompile(`^xn--`).MatchString(name) {
		return fmt.Errorf("bucket name must not start with 'xn--' (reserved for punycode)")
	}

	if regexp.MustCompile(`-s3alias$`).MatchString(name) {
		return fmt.Errorf("bucket name must not end with '-s3alias' (reserved)")
	}

	return nil
}
