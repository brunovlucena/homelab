// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	📦 BUILD CONTEXT MANAGER - Focused build context and archive management
//
//	🎯 Purpose: Handle build context creation, validation, and archive operations
//	💡 Features: Context creation, archive generation, validation, S3 integration
//
//	🏛️ ARCHITECTURE:
//	📦 Build Context - Create and manage build contexts for Lambda functions
//	🗜️ Archive Management - Generate and upload build context archives
//	🔍 Validation - Validate build requests and context requirements
//	☁️ S3 Integration - Upload and manage build contexts in S3
//	🛡️ AWS S3 Integration - Build context storage and management
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"time"

	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"knative-lambda-new/internal/aws"
	"knative-lambda-new/internal/config"
	"knative-lambda-new/internal/errors"
	"knative-lambda-new/internal/observability"
	"knative-lambda-new/internal/resilience"
	"knative-lambda-new/internal/templates"
	"knative-lambda-new/pkg/builds"
	"strings"
)

// 📦 BuildContextManagerImpl - "Focused build context and archive management"
type BuildContextManagerImpl struct {
	awsClient *aws.Client
	config    *config.Config
	obs       *observability.Observability
	// 🛡️ Rate Limiting Protection
	rateLimiter *resilience.MultiLevelRateLimiter
	// 📄 Template Processing
	templateProcessor *templates.TemplateProcessor
}

// 📦 BuildContextManagerConfig - "Configuration for creating build context manager"
type BuildContextManagerConfig struct {
	AWSClient     *aws.Client
	Config        *config.Config
	Observability *observability.Observability
	RateLimiter   *resilience.MultiLevelRateLimiter
}

// 🏗️ NewBuildContextManager - "Create new build context manager with dependencies"
func NewBuildContextManager(config BuildContextManagerConfig) (BuildContextManager, error) {
	if config.AWSClient == nil {
		return nil, errors.NewConfigurationError("build_context_manager", "aws_client", "AWS client cannot be nil")
	}

	if config.Config == nil {
		return nil, errors.NewConfigurationError("build_context_manager", "config", "config cannot be nil")
	}

	if config.Observability == nil {
		return nil, errors.NewConfigurationError("build_context_manager", "observability", "observability cannot be nil")
	}

	// Initialize template processor
	templateProcessor := templates.NewTemplateProcessor(config.Observability)

	return &BuildContextManagerImpl{
		awsClient:         config.AWSClient,
		config:            config.Config,
		obs:               config.Observability,
		rateLimiter:       config.RateLimiter,
		templateProcessor: templateProcessor,
	}, nil
}

// 📦 CreateBuildContext - "Create a build context for the build request"
func (b *BuildContextManagerImpl) CreateBuildContext(ctx context.Context, buildRequest *builds.BuildRequest) (string, error) {
	ctx, span := b.obs.StartSpan(ctx, "create_build_context")
	defer span.End()

	b.obs.Info(ctx, "Starting build context creation",
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID,
		"build_type", buildRequest.BuildType,
		"runtime", buildRequest.Runtime,
		"source_url", buildRequest.SourceURL,
		"source_bucket", buildRequest.SourceBucket,
		"source_key", buildRequest.SourceKey)

	// Validate build request
	if err := b.ValidateBuildRequest(buildRequest); err != nil {
		b.obs.Error(ctx, err, "Build request validation failed",
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
		return "", err
	}

	b.obs.Info(ctx, "Build request validation passed",
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	// Generate build context key for the archive
	buildContextKey := b.generateSourceKey(buildRequest.ParserID)

	b.obs.Info(ctx, "Generated build context key",
		"build_context_key", buildContextKey,
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	// 🛡️ Check rate limiting for build context operations
	if b.rateLimiter != nil && !b.rateLimiter.Allow("build_context") {
		b.obs.Error(ctx, errors.NewSystemError("build_context_manager", "rate_limit_exceeded"), "Rate limit exceeded for build context creation",
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
		return "", errors.NewSystemError("build_context_manager", "rate_limit_exceeded")
	}

	b.obs.Info(ctx, "Rate limit check passed",
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	// 📄 Download parser files from S3 (instead of generating them)
	// Parse source URL to extract bucket and key
	sourceBucket, sourceKey, err := b.parseSourceURL(buildRequest.SourceURL)
	if err != nil {
		b.obs.Error(ctx, err, "Failed to parse source URL",
			"source_url", buildRequest.SourceURL,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
		return "", fmt.Errorf("failed to parse source URL: %w", err)
	}

	b.obs.Info(ctx, "Parsed source URL",
		"source_url", buildRequest.SourceURL,
		"source_bucket", sourceBucket,
		"source_key", sourceKey,
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	// Generate content hash from S3 parser file
	contentHash, err := b.generateContentHash(ctx, sourceBucket, sourceKey)
	if err != nil {
		b.obs.Error(ctx, err, "Failed to generate content hash",
			"source_bucket", sourceBucket,
			"source_key", sourceKey,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
		return "", fmt.Errorf("failed to generate content hash: %w", err)
	}

	// Store content hash in build request for later use
	buildRequest.ContentHash = contentHash

	b.obs.Info(ctx, "Generated content hash for build",
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"content_hash", contentHash,
		"source_bucket", sourceBucket,
		"source_key", sourceKey,
		"correlation_id", buildRequest.CorrelationID)

	parserFiles, err := b.downloadParserFiles(ctx, sourceBucket, sourceKey, buildRequest.Runtime)
	if err != nil {
		b.obs.Error(ctx, err, "Failed to download parser files from S3",
			"source_bucket", sourceBucket,
			"source_key", sourceKey,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID,
			"error_details", err.Error())
		return "", fmt.Errorf("failed to download parser files from S3: %w", err)
	}

	b.obs.Info(ctx, "Successfully downloaded parser files",
		"files_count", len(parserFiles),
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	// 📦 Create build context archive in memory
	b.obs.Info(ctx, "Creating build context archive",
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID,
		"files_count", len(parserFiles))

	var archiveBuffer bytes.Buffer
	err = b.CreateBuildContextArchive(ctx, &archiveBuffer, buildRequest, parserFiles)
	if err != nil {
		b.obs.Error(ctx, err, "Failed to create build context archive",
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID,
			"error_details", err.Error())
		return "", fmt.Errorf("failed to create build context archive: %w", err)
	}

	b.obs.Info(ctx, "Successfully created build context archive",
		"archive_size", archiveBuffer.Len(),
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	// 📤 Upload build context archive to S3 temp bucket
	tempBucket := b.config.AWS.GetS3TempBucket()

	// 🗑️ Delete old build context if it exists to ensure fresh upload
	b.obs.Info(ctx, "Checking for existing build context to delete",
		"temp_bucket", tempBucket,
		"build_context_key", buildContextKey,
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	// Check if the object exists before trying to delete it
	objectExists, existsErr := b.awsClient.ObjectExists(ctx, tempBucket, buildContextKey)
	if existsErr != nil {
		b.obs.Error(ctx, existsErr, "Failed to check if old build context exists",
			"temp_bucket", tempBucket,
			"build_context_key", buildContextKey,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
	} else {
		b.obs.Info(ctx, "Old build context existence check",
			"temp_bucket", tempBucket,
			"build_context_key", buildContextKey,
			"object_exists", objectExists,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
	}

	// Try to delete the old context (ignore errors if it doesn't exist)
	b.obs.Info(ctx, "Attempting to delete old build context",
		"temp_bucket", tempBucket,
		"build_context_key", buildContextKey,
		"full_s3_path", fmt.Sprintf("s3://%s/%s", tempBucket, buildContextKey),
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	deleteErr := b.awsClient.DeleteObject(ctx, tempBucket, buildContextKey)
	if deleteErr != nil {
		// Log the error but don't fail - the object might not exist
		b.obs.Error(ctx, deleteErr, "Failed to delete old build context",
			"temp_bucket", tempBucket,
			"build_context_key", buildContextKey,
			"full_s3_path", fmt.Sprintf("s3://%s/%s", tempBucket, buildContextKey),
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID,
			"error_details", deleteErr.Error())

		// Check if it's a "not found" error (which is expected)
		if strings.Contains(deleteErr.Error(), "NotFound") || strings.Contains(deleteErr.Error(), "NoSuchKey") {
			b.obs.Info(ctx, "Old build context not found (expected for first-time builds)",
				"temp_bucket", tempBucket,
				"build_context_key", buildContextKey,
				"third_party_id", buildRequest.ThirdPartyID,
				"parser_id", buildRequest.ParserID,
				"correlation_id", buildRequest.CorrelationID)
		}
	} else {
		b.obs.Info(ctx, "Successfully deleted old build context",
			"temp_bucket", tempBucket,
			"build_context_key", buildContextKey,
			"full_s3_path", fmt.Sprintf("s3://%s/%s", tempBucket, buildContextKey),
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
	}

	b.obs.Info(ctx, "Uploading build context archive to S3 temp bucket",
		"temp_bucket", tempBucket,
		"build_context_key", buildContextKey,
		"archive_size", archiveBuffer.Len(),
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID)

	archiveReader := bytes.NewReader(archiveBuffer.Bytes())
	err = b.awsClient.UploadObjectWithSize(ctx, tempBucket, buildContextKey, archiveReader, "application/gzip", int64(archiveBuffer.Len()))
	if err != nil {
		b.obs.Error(ctx, err, "Failed to upload build context archive to S3",
			"temp_bucket", tempBucket,
			"build_context_key", buildContextKey,
			"archive_size", archiveBuffer.Len(),
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID,
			"error_details", err.Error())
		return "", fmt.Errorf("failed to upload build context archive to S3: %w", err)
	}

	b.obs.Info(ctx, "Build context created and uploaded successfully",
		"temp_bucket", tempBucket,
		"build_context_key", buildContextKey,
		"archive_size", archiveBuffer.Len(),
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID,
		"s3_location", fmt.Sprintf("s3://%s/%s", tempBucket, buildContextKey))

	// Verify the new context was uploaded successfully
	uploadedExists, verifyErr := b.awsClient.ObjectExists(ctx, tempBucket, buildContextKey)
	if verifyErr != nil {
		b.obs.Error(ctx, verifyErr, "Failed to verify uploaded build context",
			"temp_bucket", tempBucket,
			"build_context_key", buildContextKey,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
	} else {
		b.obs.Info(ctx, "Build context upload verification",
			"temp_bucket", tempBucket,
			"build_context_key", buildContextKey,
			"uploaded_exists", uploadedExists,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
	}

	return buildContextKey, nil
}

// 📦 CreateBuildContextArchive - "Create a build context archive"
func (b *BuildContextManagerImpl) CreateBuildContextArchive(ctx context.Context, out io.Writer, buildRequest *builds.BuildRequest, parserFiles map[string][]byte) error {
	ctx, span := b.obs.StartSpan(ctx, "create_build_context_archive")
	defer span.End()

	b.obs.Info(ctx, "Creating build context archive",
		"parser_id", buildRequest.ParserID,
		"files_count", len(parserFiles),
		"build_type", buildRequest.BuildType,
		"runtime", buildRequest.Runtime)

	// Create gzip writer
	gzipWriter := gzip.NewWriter(out)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// 📄 Add build configuration file
	buildConfig := b.createBuildConfig(buildRequest)
	buildConfigBytes := []byte(buildConfig)

	err := b.addFileToArchive(tarWriter, "build-config.json", buildConfigBytes, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add build config to archive: %w", err)
	}

	// 📄 Add parser files
	for filename, content := range parserFiles {
		// For the main parser file, rename it to the parser ID for Node.js runtimes
		if strings.HasPrefix(buildRequest.Runtime, "nodejs") && filename == b.getParserFileName(buildRequest.Runtime) {
			// Rename the main parser file to {parserId}.js
			parserFileName := fmt.Sprintf("%s.js", buildRequest.ParserID)
			b.obs.Info(ctx, "Renaming parser file for Node.js runtime",
				"original_filename", filename,
				"new_filename", parserFileName,
				"parser_id", buildRequest.ParserID,
				"runtime", buildRequest.Runtime)

			// Add the parser file as-is (no conversion needed)
			err = b.addFileToArchive(tarWriter, parserFileName, content, time.Now())
			if err != nil {
				return fmt.Errorf("failed to add renamed parser file %s to archive: %w", parserFileName, err)
			}
		} else {
			// Keep other files (like package.json) with their original names
			err = b.addFileToArchive(tarWriter, filename, content, time.Now())
			if err != nil {
				return fmt.Errorf("failed to add parser file %s to archive: %w", filename, err)
			}
		}
	}

	// 📄 Add Dockerfile if it's a container build
	if buildRequest.BuildType == "container" {
		dockerfileBytes, err := b.createDockerfile(buildRequest)
		if err != nil {
			return fmt.Errorf("failed to create Dockerfile: %w", err)
		}

		err = b.addFileToArchive(tarWriter, "Dockerfile", dockerfileBytes, time.Now())
		if err != nil {
			return fmt.Errorf("failed to add Dockerfile to archive: %w", err)
		}
	}

	// 📄 Add .dockerignore if it's a container build
	if buildRequest.BuildType == "container" {
		dockerignore := b.createDockerignore()
		dockerignoreBytes := []byte(dockerignore)

		err = b.addFileToArchive(tarWriter, ".dockerignore", dockerignoreBytes, time.Now())
		if err != nil {
			return fmt.Errorf("failed to add .dockerignore to archive: %w", err)
		}
	}

	// 📄 Add environment variables file if provided
	if len(buildRequest.Environment) > 0 {
		envContent := b.createEnvironmentFile(buildRequest.Environment)
		envBytes := []byte(envContent)

		err = b.addFileToArchive(tarWriter, ".env", envBytes, time.Now())
		if err != nil {
			return fmt.Errorf("failed to add environment file to archive: %w", err)
		}
	}

	// 📄 Add package.json and index.js for Node.js runtimes to support ES modules
	if strings.HasPrefix(buildRequest.Runtime, "nodejs") {
		packageJSONBytes, err := b.createPackageJSONTemplate(buildRequest)
		if err != nil {
			return fmt.Errorf("failed to create package.json: %w", err)
		}

		err = b.addFileToArchive(tarWriter, "package.json", packageJSONBytes, time.Now())
		if err != nil {
			return fmt.Errorf("failed to add package.json to archive: %w", err)
		}

		// Add index.js template for Node.js runtimes
		indexJSBytes, err := b.createIndexJS(buildRequest)
		if err != nil {
			return fmt.Errorf("failed to create index.js: %w", err)
		}

		err = b.addFileToArchive(tarWriter, "index.js", indexJSBytes, time.Now())
		if err != nil {
			return fmt.Errorf("failed to add index.js to archive: %w", err)
		}
	}

	b.obs.Info(ctx, "Build context archive created successfully",
		"parser_id", buildRequest.ParserID,
		"build_type", buildRequest.BuildType,
		"runtime", buildRequest.Runtime,
		"has_environment", len(buildRequest.Environment) > 0,
		"has_package_json", strings.HasPrefix(buildRequest.Runtime, "nodejs"))

	return nil
}

// 🔍 ValidateBuildRequest - "Validate a build request"
func (b *BuildContextManagerImpl) ValidateBuildRequest(buildRequest *builds.BuildRequest) error {
	if buildRequest == nil {
		return errors.NewValidationError("build_request", nil, "build request cannot be nil")
	}

	// Validate required fields
	if err := errors.ValidateRequired("third_party_id", buildRequest.ThirdPartyID); err != nil {
		return err
	}

	if err := errors.ValidateRequired("parser_id", buildRequest.ParserID); err != nil {
		return err
	}

	if err := errors.ValidateRequired("correlation_id", buildRequest.CorrelationID); err != nil {
		return err
	}

	// Validate field lengths
	if len(buildRequest.ThirdPartyID) > 100 {
		return errors.NewValidationError("third_party_id", buildRequest.ThirdPartyID, "third party ID too long (max 100 characters)")
	}

	if len(buildRequest.ParserID) > 100 {
		return errors.NewValidationError("parser_id", buildRequest.ParserID, "parser ID too long (max 100 characters)")
	}

	// Validate correlation ID format (should be a UUID)
	if err := b.validateUUID("correlation_id", buildRequest.CorrelationID); err != nil {
		return err
	}

	return nil
}

// 🔧 generateSourceKey - "Generate S3 source key for build context"
func (b *BuildContextManagerImpl) generateSourceKey(parserID string) string {
	return fmt.Sprintf("build-context/%s/context.tar.gz", parserID)
}

// 🔧 parseSourceURL - "Parse source URL to extract bucket and key"
func (b *BuildContextManagerImpl) parseSourceURL(sourceURL string) (string, string, error) {
	// Simple regex to extract bucket and key
	re := regexp.MustCompile(`s3://([^/]+)/(.+)`)
	matches := re.FindStringSubmatch(sourceURL)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid source URL format: %s", sourceURL)
	}
	return matches[1], matches[2], nil
}

// 🔧 downloadParserFiles - "Download parser files from S3 directory"
func (b *BuildContextManagerImpl) downloadParserFiles(ctx context.Context, sourceBucket, sourceKey, runtime string) (map[string][]byte, error) {
	ctx, span := b.obs.StartSpan(ctx, "download_parser_files")
	defer span.End()

	parserFiles := make(map[string][]byte)

	// The sourceKey IS the parser file itself (e.g., the JavaScript code)
	// We need to download it and rename it to the appropriate filename for the runtime
	parserFileName := b.getParserFileName(runtime)

	b.obs.Info(ctx, "Starting parser file download",
		"source_bucket", sourceBucket,
		"source_key", sourceKey,
		"parser_filename", parserFileName,
		"runtime", runtime)

	// Download the parser file (sourceKey is the actual file)
	b.obs.Info(ctx, "Attempting to download parser file",
		"source_bucket", sourceBucket,
		"file_key", sourceKey,
		"parser_filename", parserFileName,
		"full_s3_path", fmt.Sprintf("s3://%s/%s", sourceBucket, sourceKey))

	reader, objectSize, err := b.awsClient.GetObject(ctx, sourceBucket, sourceKey)
	if err != nil {
		return nil, fmt.Errorf("failed to download parser file %s: %w", sourceKey, err)
	}
	defer reader.Close()

	// Read file content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read parser file content for %s: %w", sourceKey, err)
	}

	// Store the content with the appropriate filename for the runtime
	parserFiles[parserFileName] = content

	b.obs.Info(ctx, "Successfully downloaded parser file",
		"file_key", sourceKey,
		"parser_filename", parserFileName,
		"size", objectSize.Size)

	return parserFiles, nil
}

// 🔧 generateContentHash - "Generate content hash from S3 parser file"
func (b *BuildContextManagerImpl) generateContentHash(ctx context.Context, sourceBucket, sourceKey string) (string, error) {
	ctx, span := b.obs.StartSpan(ctx, "generate_content_hash")
	defer span.End()

	b.obs.Info(ctx, "Generating content hash from S3 parser",
		"source_bucket", sourceBucket,
		"source_key", sourceKey)

	// Download parser file from S3
	reader, _, err := b.awsClient.GetObject(ctx, sourceBucket, sourceKey)
	if err != nil {
		return "", fmt.Errorf("failed to download parser for hash: %w", err)
	}
	defer reader.Close()

	// Read content
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read parser content for hash: %w", err)
	}

	// Generate SHA256 hash of content
	hash := sha256.Sum256(content)
	contentHash := hex.EncodeToString(hash[:])[:12] // Use first 12 chars for shorter tag

	b.obs.Info(ctx, "Generated content hash",
		"source_bucket", sourceBucket,
		"source_key", sourceKey,
		"content_size", len(content),
		"content_hash", contentHash)

	return contentHash, nil
}

// 🔧 validateUUID - "Validate UUID format"
func (b *BuildContextManagerImpl) validateUUID(fieldName, value string) error {
	if value == "" {
		return errors.NewValidationError(fieldName, value, "UUID cannot be empty")
	}

	// Simple UUID v4 validation regex
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(value) {
		return errors.NewValidationError(fieldName, value, "invalid UUID format")
	}

	return nil
}

// 🔧 addFileToArchive - "Add a file to the tar archive"
func (b *BuildContextManagerImpl) addFileToArchive(tarWriter *tar.Writer, filename string, content []byte, modTime time.Time) error {
	header := &tar.Header{
		Name:    filename,
		Mode:    0644,
		Size:    int64(len(content)),
		ModTime: modTime,
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header for %s: %w", filename, err)
	}

	if _, err := tarWriter.Write(content); err != nil {
		return fmt.Errorf("failed to write content for %s: %w", filename, err)
	}

	return nil
}

// 🔧 createBuildConfig - "Create build configuration JSON"
func (b *BuildContextManagerImpl) createBuildConfig(buildRequest *builds.BuildRequest) string {
	config := map[string]interface{}{
		"build_type":     buildRequest.BuildType,
		"runtime":        buildRequest.Runtime,
		"source_url":     buildRequest.SourceURL,
		"third_party_id": buildRequest.ThirdPartyID,
		"parser_id":      buildRequest.ParserID,
		"block_id":       buildRequest.BlockID,
		"build_timeout":  buildRequest.BuildTimeout,
		"build_args":     buildRequest.BuildArgs,
		"tags":           buildRequest.Tags,
		"created_at":     buildRequest.CreatedAt.Format(time.RFC3339),
		"correlation_id": buildRequest.CorrelationID,
	}

	// Convert to JSON
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		// Fallback to simple string representation
		return fmt.Sprintf(`{"build_type":"%s","runtime":"%s","parser_id":"%s"}`,
			buildRequest.BuildType, buildRequest.Runtime, buildRequest.ParserID)
	}

	return string(configJSON)
}

// 🔧 getParserFileName - "Get appropriate filename for parser code based on runtime"
func (b *BuildContextManagerImpl) getParserFileName(runtime string) string {
	switch runtime {
	case "nodejs22", "nodejs22.x":
		return "index.js"
	// TODO: Implement it
	// case "python3.9", "python3.10", "python3.11", "python3.12":
	// 	return "lambda_function.py"
	// case "go1.x":
	// 	return "main.go"
	// case "java11", "java17", "java21":
	// 	return "LambdaFunction.java"
	// case "dotnet6", "dotnet8":
	// 	return "Function.cs"
	default:
		return "function." + runtime
	}
}

// 🔧 createDockerfile - "Create Dockerfile for container builds using template"
func (b *BuildContextManagerImpl) createDockerfile(buildRequest *builds.BuildRequest) ([]byte, error) {
	// Get base image for the runtime
	baseImage := b.config.AWS.NodeBaseImage

	// Get runtime CMD
	runtimeCMD := b.getRuntimeCMD(buildRequest.Runtime)

	// Create template data
	templateData := templates.CreateTemplateData(buildRequest, baseImage, runtimeCMD)

	// Process the Dockerfile template
	dockerfileBytes, err := b.templateProcessor.ProcessDockerfileTemplate(context.Background(), templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to process Dockerfile template: %w", err)
	}

	return dockerfileBytes, nil
}

// 🔧 createIndexJS - "Create index.js for Node.js runtimes using template"
func (b *BuildContextManagerImpl) createIndexJS(buildRequest *builds.BuildRequest) ([]byte, error) {
	// Get base image for the runtime (needed for template data)
	baseImage := b.config.AWS.NodeBaseImage

	// Get runtime CMD (needed for template data)
	runtimeCMD := b.getRuntimeCMD(buildRequest.Runtime)

	// Create template data
	templateData := templates.CreateTemplateData(buildRequest, baseImage, runtimeCMD)

	// Process the index.js template
	indexJSBytes, err := b.templateProcessor.ProcessIndexJSTemplate(context.Background(), templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to process index.js template: %w", err)
	}

	return indexJSBytes, nil
}

// 🔧 createPackageJSONTemplate - "Create package.json for Node.js runtimes using template"
func (b *BuildContextManagerImpl) createPackageJSONTemplate(buildRequest *builds.BuildRequest) ([]byte, error) {
	// Get base image for the runtime (needed for template data)
	baseImage := b.config.AWS.NodeBaseImage

	// Get runtime CMD (needed for template data)
	runtimeCMD := b.getRuntimeCMD(buildRequest.Runtime)

	// Create template data
	templateData := templates.CreateTemplateData(buildRequest, baseImage, runtimeCMD)

	// Process the package.json template
	packageJSONBytes, err := b.templateProcessor.ProcessPackageJSONTemplate(context.Background(), templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to process package.json template: %w", err)
	}

	return packageJSONBytes, nil
}

// 🔧 getRuntimeCMD - "Get runtime-specific CMD for Dockerfile"
func (b *BuildContextManagerImpl) getRuntimeCMD(runtime string) string {
	parserFileName := b.getParserFileName(runtime)
	switch runtime {
	case "nodejs22", "nodejs22.x":
		return `CMD ["node", "index.js"]`
	case "python3.9", "python3.10", "python3.11", "python3.12":
		return fmt.Sprintf(`CMD ["python", "%s"]`, parserFileName)
	case "go1.x":
		return `CMD ["./main"]`
	default:
		return `CMD ["echo", "Unsupported runtime: ` + runtime + `"]`
	}
}

// 🔧 createDockerignore - "Create .dockerignore file"
func (b *BuildContextManagerImpl) createDockerignore() string {
	return `# Build artifacts
*.log
*.tmp
*.tar.gz

# Development files
.git
.gitignore
README.md
*.md

# IDE files
.vscode
.idea
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db

# Node.js
node_modules
npm-debug.log

# Python
__pycache__
*.pyc
*.pyo
*.pyd
.Python
env
pip-log.txt
pip-delete-this-directory.txt
.tox
.coverage
.coverage.*
.cache
nosetests.xml
coverage.xml
*.cover
*.log
.git
.mypy_cache
.pytest_cache
.hypothesis

# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
go.work
`
}

// 🔧 createEnvironmentFile - "Create environment variables file"
func (b *BuildContextManagerImpl) createEnvironmentFile(environment map[string]string) string {
	var envContent strings.Builder

	for key, value := range environment {
		envContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	return envContent.String()
}
