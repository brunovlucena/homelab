// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🔐 AWS CLIENT - AWS Services Integration
//
//	🎯 Purpose: AWS SDK client wrapper for S3 and ECR operations
//	💡 Features: S3 object management, ECR repository management
//
//	🏛️ ARCHITECTURE:
//	📦 S3 Operations - Object upload, download, existence checks
//	🐳 ECR Operations - Repository management, image URI generation

//	📊 Observability - Comprehensive logging and metrics
//
//	🔧 COMPONENTS:
//	🎯 Client Configuration - AWS credentials and region setup
//	📦 S3 Client - Object storage operations
//	🐳 ECR Client - Container registry operations

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package aws

import (
	"context"
	"fmt"
	"io"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"knative-lambda-new/internal/errors"
	"knative-lambda-new/internal/observability"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🎯 CLIENT STRUCTURE - "AWS client with all services"                  │
// └─────────────────────────────────────────────────────────────────────────┘

// 🎯 Client - "AWS client with S3 and ECR services"
type Client struct {
	config            aws.Config
	s3Client          *s3.Client
	ecrClient         *ecr.Client
	region            string
	accountID         string
	ecrRegistry       string
	ecrRepositoryName string
	sourceBucket      string
	tempBucket        string
	obs               *observability.Observability
}

// 🎯 ClientConfig - "Configuration for creating AWS client"
type ClientConfig struct {
	Region            string
	AccountID         string
	ECRRegistry       string
	ECRRepositoryName string
	S3SourceBucket    string
	S3TempBucket      string
	Observability     *observability.Observability
}

// 🎯 ObjectSize - "S3 object size information"
type ObjectSize struct {
	Size int64
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏗️ CLIENT CREATION - "Create and configure AWS client"                │
// └─────────────────────────────────────────────────────────────────────────┘

// 🏗️ NewClient - "Create new AWS client with configuration"
func NewClient(ctx context.Context, cfg ClientConfig) (*Client, error) {
	ctx, span := cfg.Observability.StartSpan(ctx, "aws_client_creation")
	defer span.End()

	// Load AWS configuration
	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithRetryMode(aws.RetryModeAdaptive),
		config.WithRetryMaxAttempts(3),
	)
	if err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("Failed to load AWS config: %v", err))
		return nil, errors.NewConfigurationError("aws", "config", fmt.Sprintf("failed to load AWS config: %v", err))
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(awsConfig)

	// Create ECR client
	ecrClient := ecr.NewFromConfig(awsConfig)

	span.SetAttributes(
		attribute.String("aws.region", cfg.Region),
		attribute.String("aws.account_id", cfg.AccountID),
		attribute.String("aws.ecr_registry", cfg.ECRRegistry),
		attribute.String("aws.ecr_repository_name", cfg.ECRRepositoryName),
		attribute.String("aws.s3_source_bucket", cfg.S3SourceBucket),
		attribute.String("aws.s3_temp_bucket", cfg.S3TempBucket),
	)

	return &Client{
		config:            awsConfig,
		s3Client:          s3Client,
		ecrClient:         ecrClient,
		region:            cfg.Region,
		accountID:         cfg.AccountID,
		ecrRegistry:       cfg.ECRRegistry,
		ecrRepositoryName: cfg.ECRRepositoryName,
		sourceBucket:      cfg.S3SourceBucket,
		tempBucket:        cfg.S3TempBucket,
		obs:               cfg.Observability,
	}, nil
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 S3 OPERATIONS - "S3 object management operations"                  │
// └─────────────────────────────────────────────────────────────────────────┘

// 📦 ObjectExists - "Check if S3 object exists"
func (c *Client) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	ctx, span := c.obs.StartSpanWithAttributes(ctx, "s3_object_exists", map[string]string{
		"aws.service":   "s3",
		"aws.operation": "head_object",
		"s3.bucket":     bucket,
		"s3.key":        key,
	})
	defer span.End()

	_, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if err.Error() == "NotFound" || err.Error() == "NoSuchKey" {
			span.SetAttributes(attribute.Bool("s3.object_exists", false))
			return false, nil
		}
		span.SetStatus(codes.Error, err.Error())
		// Use enhanced error handling with context
		return false, errors.WrapWithContext(err, fmt.Sprintf("bucket=%s, key=%s", bucket, key))
	}

	span.SetAttributes(attribute.Bool("s3.object_exists", true))
	return true, nil
}

// 📦 GetObject - "Get S3 object with size information"
func (c *Client) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, ObjectSize, error) {
	ctx, span := c.obs.StartSpanWithAttributes(ctx, "s3_get_object", map[string]string{
		"aws.service":   "s3",
		"aws.operation": "get_object",
		"s3.bucket":     bucket,
		"s3.key":        key,
	})
	defer span.End()

	// First get object info
	headResult, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, ObjectSize{}, errors.WrapWithContext(err, fmt.Sprintf("bucket=%s, key=%s", bucket, key))
	}

	// Then get the object
	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, ObjectSize{}, errors.WrapWithContext(err, fmt.Sprintf("bucket=%s, key=%s", bucket, key))
	}

	var size int64
	if headResult.ContentLength != nil {
		size = *headResult.ContentLength
	}

	span.SetAttributes(attribute.Int64("s3.object_size", size))
	objectSize := ObjectSize{
		Size: size,
	}

	return result.Body, objectSize, nil
}

// 📦 UploadObjectWithSize - "Upload object to S3 with known size"
func (c *Client) UploadObjectWithSize(ctx context.Context, bucket, key string, reader io.Reader, contentType string, size int64) error {
	startTime := time.Now()

	ctx, span := c.obs.StartSpanWithAttributes(ctx, "s3_upload_object", map[string]string{
		"aws.service":     "s3",
		"aws.operation":   "put_object",
		"s3.bucket":       bucket,
		"s3.key":          key,
		"s3.content_type": contentType,
	})
	defer span.End()

	span.SetAttributes(attribute.Int64("s3.object_size", size))

	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})

	duration := time.Since(startTime)

	// Record S3 upload metrics
	if c.obs.GetMetrics() != nil {
		metricsRec := observability.NewMetricsRecorder(c.obs)
		if err != nil {
			metricsRec.RecordS3Upload(ctx, bucket, "failure")
		} else {
			metricsRec.RecordS3Upload(ctx, bucket, "success")
			metricsRec.RecordS3UploadDuration(ctx, bucket, duration)
			metricsRec.RecordS3UploadSize(ctx, bucket, size)
		}
	}

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return errors.WrapWithContext(err, fmt.Sprintf("bucket=%s, key=%s, size=%d", bucket, key, size))
	}

	return nil
}

// 📦 DeleteObject - "Delete S3 object"
func (c *Client) DeleteObject(ctx context.Context, bucket, key string) error {
	ctx, span := c.obs.StartSpanWithAttributes(ctx, "s3_delete_object", map[string]string{
		"aws.service":   "s3",
		"aws.operation": "delete_object",
		"s3.bucket":     bucket,
		"s3.key":        key,
	})
	defer span.End()

	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return errors.WrapWithContext(err, fmt.Sprintf("bucket=%s, key=%s", bucket, key))
	}

	return nil
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🐳 ECR OPERATIONS - "ECR repository management"                       │
// └─────────────────────────────────────────────────────────────────────────┘

// 🐳 EnsureKnativeLambdasRepository - "Ensure ECR repository exists"
func (c *Client) EnsureKnativeLambdasRepository(ctx context.Context) error {
	startTime := time.Now()

	ctx, span := c.obs.StartSpanWithAttributes(ctx, "ecr_ensure_repository", map[string]string{
		"aws.service":         "ecr",
		"aws.operation":       "ensure_repository",
		"ecr.repository_name": c.ecrRepositoryName,
	})
	defer span.End()

	repositoryName := c.ecrRepositoryName

	// Check if repository exists
	_, err := c.ecrClient.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{repositoryName},
	})

	if err != nil {
		// Repository doesn't exist, create it
		ctx, createSpan := c.obs.StartSpan(ctx, "ecr_create_repository")
		_, err := c.ecrClient.CreateRepository(ctx, &ecr.CreateRepositoryInput{
			RepositoryName: aws.String(repositoryName),
		})
		if err != nil {
			createSpan.SetStatus(codes.Error, err.Error())
			span.SetStatus(codes.Error, err.Error())

			// Record ECR push failure metric
			if c.obs.GetMetrics() != nil {
				metricsRec := observability.NewMetricsRecorder(c.obs)
				metricsRec.RecordECRPush(ctx, repositoryName, "failure")
				metricsRec.RecordECRPushDuration(ctx, repositoryName, time.Since(startTime))
			}

			return errors.WrapWithContext(err, fmt.Sprintf("repository=%s", repositoryName))
		}
		createSpan.End()
		span.SetAttributes(attribute.Bool("ecr.repository_created", true))

		// Record ECR push success metric
		if c.obs.GetMetrics() != nil {
			metricsRec := observability.NewMetricsRecorder(c.obs)
			metricsRec.RecordECRPush(ctx, repositoryName, "success")
			metricsRec.RecordECRPushDuration(ctx, repositoryName, time.Since(startTime))
		}
	} else {
		span.SetAttributes(attribute.Bool("ecr.repository_exists", true))

		// Record ECR push success metric (repository already exists)
		if c.obs.GetMetrics() != nil {
			metricsRec := observability.NewMetricsRecorder(c.obs)
			metricsRec.RecordECRPush(ctx, repositoryName, "success")
			metricsRec.RecordECRPushDuration(ctx, repositoryName, time.Since(startTime))
		}
	}

	return nil
}

// 🐳 GetImageURI - "Generate ECR image URI"
func (c *Client) GetImageURI(thirdPartyID, parserID string) string {
	repositoryName := c.ecrRepositoryName
	imageTag := fmt.Sprintf("%s-%s", thirdPartyID, parserID)

	if c.ecrRegistry != "" {
		return fmt.Sprintf("%s/%s:%s", c.ecrRegistry, repositoryName, imageTag)
	}

	// Fallback to standard ECR URI format
	return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s",
		c.accountID, c.region, repositoryName, imageTag)
}
