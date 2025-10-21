// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🔐 AWS CONFIGURATION - AWS services configuration
//
//	🎯 Purpose: AWS services settings, ECR registry, S3 buckets, IAM settings
//	💡 Features: Region configuration, ECR registry, S3 bucket settings, EKS pod identity
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
)

// 🔐 AWSConfig - "AWS services configuration"
type AWSConfig struct {
	// Region Configuration
	AWSRegion    string `envconfig:"AWS_REGION"`
	AWSAccountID string `envconfig:"AWS_ACCOUNT_ID"`

	// ECR Configuration
	ECRRegistry       string `envconfig:"ECR_REGISTRY"`
	ECRRepositoryName string `envconfig:"ECR_REPOSITORY_NAME"`

	// S3 Configuration
	S3SourceBucket string `envconfig:"S3_SOURCE_BUCKET"`
	S3TempBucket   string `envconfig:"S3_TEMP_BUCKET"`

	// Registry Configuration
	RegistryMirror        string `envconfig:"REGISTRY_MIRROR"`
	SkipTLSVerifyRegistry string `envconfig:"SKIP_TLS_VERIFY_REGISTRY"`

	// Base Image Configuration
	NodeBaseImage   string `envconfig:"NODE_BASE_IMAGE"`
	PythonBaseImage string `envconfig:"PYTHON_BASE_IMAGE"`
	GoBaseImage     string `envconfig:"GO_BASE_IMAGE"`

	// EKS Pod Identity Configuration
	UseEKSPodIdentity bool   `envconfig:"USE_EKS_POD_IDENTITY"`
	PodIdentityRole   string `envconfig:"POD_IDENTITY_ROLE"`
}

// 🔧 NewAWSConfig - "Create AWS configuration with environment variables first, then fallback to constants"
func NewAWSConfig() *AWSConfig {
	return &AWSConfig{
		AWSRegion:             getEnv("AWS_REGION", constants.AWSRegionDefault),
		AWSAccountID:          getEnv("AWS_ACCOUNT_ID", constants.AWSAccountIDDefault),
		ECRRegistry:           getEnv("ECR_REGISTRY", constants.RegistryDefault),
		ECRRepositoryName:     getEnv("ECR_REPOSITORY_NAME", constants.AWSECRRepositoryNameDefault),
		S3SourceBucket:        getEnv("S3_SOURCE_BUCKET", constants.S3SourceBucketDefault),
		S3TempBucket:          getEnv("S3_TEMP_BUCKET", constants.S3TempBucketDefault),
		RegistryMirror:        getEnv("REGISTRY_MIRROR", constants.RegistryMirrorDefault),
		SkipTLSVerifyRegistry: getEnv("SKIP_TLS_VERIFY_REGISTRY", constants.SkipTLSVerifyRegistryDefault),
		NodeBaseImage:         getEnv("NODE_BASE_IMAGE", constants.NodeBaseImageDefault),
		PythonBaseImage:       getEnv("PYTHON_BASE_IMAGE", constants.PythonBaseImageDefault),
		GoBaseImage:           getEnv("GO_BASE_IMAGE", constants.GoBaseImageDefault),
		UseEKSPodIdentity:     getEnvBool("USE_EKS_POD_IDENTITY", constants.UseEKSPodIdentityDefault),
		PodIdentityRole:       getEnv("POD_IDENTITY_ROLE", constants.PodIdentityRoleDefault),
	}
}

// 🔧 Validate - "Validate AWS configuration"
func (c *AWSConfig) Validate() error {
	if c.AWSRegion == "" {
		return errors.NewValidationError("aws_region", c.AWSRegion, constants.ErrAWSRegionRequired)
	}

	if c.AWSAccountID == "" {
		return errors.NewValidationError("aws_account_id", c.AWSAccountID, constants.ErrAWSAccountIDRequired)
	}

	if len(c.AWSAccountID) != constants.AWSAccountIDLength {
		return errors.NewValidationError("aws_account_id", c.AWSAccountID, constants.ErrAWSAccountIDMustBe12Digits)
	}

	if c.ECRRegistry == "" {
		return errors.NewValidationError("ecr_registry", c.ECRRegistry, constants.ErrECRRegistryRequired)
	}

	if c.ECRRepositoryName == "" {
		return errors.NewValidationError("ecr_repository_name", c.ECRRepositoryName, constants.ErrECRRepositoryNameRequired)
	}

	if c.S3SourceBucket == "" {
		return errors.NewValidationError("s3_source_bucket", c.S3SourceBucket, constants.ErrS3SourceBucketRequired)
	}

	if c.S3TempBucket == "" {
		return errors.NewValidationError("s3_temp_bucket", c.S3TempBucket, constants.ErrS3TempBucketRequired)
	}

	// Validate pod identity configuration
	if c.UseEKSPodIdentity && c.PodIdentityRole == "" {
		return errors.NewValidationError("pod_identity_role", c.PodIdentityRole, constants.ErrPodIdentityRoleRequired)
	}

	return nil
}

// 🔧 GetAWSConfig - "Get AWS SDK configuration"
func (c *AWSConfig) GetAWSConfig(ctx context.Context) (aws.Config, error) {
	// Configure AWS SDK with EKS pod identity support
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(c.AWSRegion),
		config.WithRetryMode(aws.RetryModeAdaptive),
		config.WithRetryMaxAttempts(constants.MaxRetriesDefault),
	}

	// When using EKS pod identity, the AWS SDK automatically detects and uses
	// the pod identity credentials from the EKS pod identity webhook
	// No additional configuration is needed as the default credential provider chain
	// includes the EKS pod identity provider

	return config.LoadDefaultConfig(ctx, opts...)
}

// 🔧 GetRegion - "Get AWS region"
func (c *AWSConfig) GetRegion() string {
	return c.AWSRegion
}

// 🔧 GetAccountID - "Get AWS account ID"
func (c *AWSConfig) GetAccountID() string {
	return c.AWSAccountID
}

// 🔧 GetECRRegistry - "Get ECR registry URL"
func (c *AWSConfig) GetECRRegistry() string {
	return c.ECRRegistry
}

// 🔧 GetECRRepositoryName - "Get ECR repository name"
func (c *AWSConfig) GetECRRepositoryName() string {
	return c.ECRRepositoryName
}

// 🔧 GetS3SourceBucket - "Get S3 source bucket name"
func (c *AWSConfig) GetS3SourceBucket() string {
	return c.S3SourceBucket
}

// 🔧 GetS3TempBucket - "Get S3 temp bucket name"
func (c *AWSConfig) GetS3TempBucket() string {
	return c.S3TempBucket
}

// 🔧 IsEKSPodIdentityEnabled - "Check if EKS pod identity is enabled"
func (c *AWSConfig) IsEKSPodIdentityEnabled() bool {
	return c.UseEKSPodIdentity
}

// 🔧 GetPodIdentityRole - "Get pod identity role name"
func (c *AWSConfig) GetPodIdentityRole() string {
	return c.PodIdentityRole
}
