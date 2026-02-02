// Package validation provides security validation utilities for the knative-lambda-operator.
// This package implements input validation to prevent:
// - SSRF (Server-Side Request Forgery) via git URLs
// - Path traversal attacks via git paths
// - Template injection via handler fields
// - Command injection via shell metacharacters
//
// Security Fixes:
// - BLUE-001: SSRF via Go-Git Library
// - BLUE-002: Go Template Injection
// - BLUE-005: Path Traversal in Git Source Path
// - VULN-001: Command Injection via Git Clone
// - VULN-009: Missing CRD Validation
package validation

import (
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// validHandlerPattern matches safe handler formats: module.function
	// Only allows alphanumeric, underscore, and single dot separator
	// Examples: main.handler, index.handler, app.process_event
	validHandlerPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z_][a-zA-Z0-9_]*$`)

	// validGitRefPattern matches safe git refs (branch, tag, commit)
	// Only allows alphanumeric, dash, underscore, dot, and forward slash
	validGitRefPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*$`)

	// validBucketPattern matches S3/MinIO bucket names
	// Must be 3-63 chars, lowercase alphanumeric, dots and hyphens
	validBucketPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)

	// validObjectKeyPattern matches S3/MinIO object keys
	// Allows alphanumeric, common safe characters, no shell metacharacters
	validObjectKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9!_.*'()/-]+$`)

	// validEndpointPattern matches safe endpoint hostnames
	// hostname:port or hostname format
	validEndpointPattern = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9.]*[a-zA-Z0-9](:[0-9]{1,5})?$`)

	// shellMetacharacters detects dangerous shell injection characters
	shellMetacharacters = regexp.MustCompile("[;&|$`(){}\\[\\]<>!#*?~\n\r\\\\]")

	// blockedHosts contains hosts that should never be accessed via SSRF
	blockedHosts = []string{
		"169.254.169.254",          // AWS metadata
		"169.254.170.2",            // AWS ECS metadata
		"metadata.google.internal", // GCP metadata
		"metadata.goog",            // GCP metadata alternative
		"kubernetes",               // K8s API
		"kubernetes.default",       // K8s API
		"kubernetes.default.svc",   // K8s API
		"localhost",                // Localhost
		"127.0.0.1",                // Loopback
		"0.0.0.0",                  // All interfaces
		"[::1]",                    // IPv6 loopback
		"10.96.0.1",                // Common K8s API ClusterIP
	}

	// blockedIPRanges contains CIDR ranges that should not be accessed
	blockedIPRanges = []string{
		"169.254.0.0/16", // Link-local (metadata endpoints)
		"127.0.0.0/8",    // Loopback
		"10.0.0.0/8",     // Private (could expose internal services)
		"172.16.0.0/12",  // Private
		"192.168.0.0/16", // Private
		"100.64.0.0/10",  // Carrier-grade NAT
	}

	// parsedBlockedRanges holds parsed CIDR networks
	parsedBlockedRanges []*net.IPNet
)

func init() {
	// Parse blocked IP ranges at startup
	for _, cidr := range blockedIPRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			parsedBlockedRanges = append(parsedBlockedRanges, network)
		}
	}
}

// ValidationError represents a security validation error
type ValidationError struct {
	Field   string
	Message string
	Code    string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message, code string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	}
}

// ValidateHandler validates the handler field to prevent template injection (BLUE-002)
// Handler must be in format: module.function (e.g., main.handler, index.process)
func ValidateHandler(handler string) error {
	if handler == "" {
		return nil // Empty handler uses default
	}

	// Check for maximum length to prevent DoS
	if len(handler) > 100 {
		return NewValidationError("handler", "handler exceeds maximum length of 100 characters", "HANDLER_TOO_LONG")
	}

	// Check for shell metacharacters that could escape template context
	if shellMetacharacters.MatchString(handler) {
		return NewValidationError("handler", "handler contains invalid characters (potential injection)", "HANDLER_INJECTION")
	}

	// Check for quotes that could break template escaping
	if strings.ContainsAny(handler, `"'\`+"`") {
		return NewValidationError("handler", "handler contains quote characters (potential injection)", "HANDLER_QUOTES")
	}

	// Validate format: module.function only
	if !validHandlerPattern.MatchString(handler) {
		return NewValidationError("handler",
			"handler must be in format 'module.function' (e.g., main.handler, index.process)",
			"HANDLER_INVALID_FORMAT")
	}

	return nil
}

// SanitizeHandler returns a safe handler value for template rendering
// If the handler is empty or invalid, returns a safe default
func SanitizeHandler(handler string) string {
	// Empty handler should use default
	if handler == "" {
		return "main.handler"
	}
	// Invalid handler should use default
	if err := ValidateHandler(handler); err != nil {
		return "main.handler"
	}
	return handler
}

// ValidateGitURL validates a Git URL to prevent SSRF (BLUE-001)
func ValidateGitURL(gitURL string) error {
	if gitURL == "" {
		return NewValidationError("git.url", "git URL is required", "URL_REQUIRED")
	}

	// Check for maximum length
	if len(gitURL) > 2048 {
		return NewValidationError("git.url", "git URL exceeds maximum length of 2048 characters", "URL_TOO_LONG")
	}

	// Check for shell metacharacters
	if shellMetacharacters.MatchString(gitURL) {
		return NewValidationError("git.url", "git URL contains invalid characters (potential injection)", "URL_INJECTION")
	}

	// Parse the URL
	parsed, err := url.Parse(gitURL)
	if err != nil {
		return NewValidationError("git.url", fmt.Sprintf("invalid git URL format: %v", err), "URL_PARSE_ERROR")
	}

	// Check scheme - only allow https and git protocols for security
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "https" && scheme != "git" && scheme != "ssh" {
		// Allow http only for internal registries (*.svc.cluster.local)
		if scheme == "http" {
			if !strings.HasSuffix(parsed.Host, ".svc.cluster.local") &&
				!strings.HasSuffix(parsed.Host, ".svc") {
				return NewValidationError("git.url",
					"git URL must use HTTPS for external repositories (http only allowed for *.svc.cluster.local)",
					"URL_INSECURE_SCHEME")
			}
		} else {
			return NewValidationError("git.url",
				fmt.Sprintf("git URL scheme '%s' is not allowed (use https, git, or ssh)", scheme),
				"URL_INVALID_SCHEME")
		}
	}

	// Extract host (without port)
	host := parsed.Hostname()

	// Check against blocked hosts (SSRF prevention)
	hostLower := strings.ToLower(host)
	for _, blocked := range blockedHosts {
		if hostLower == blocked || strings.HasSuffix(hostLower, "."+blocked) {
			return NewValidationError("git.url",
				fmt.Sprintf("git URL host '%s' is blocked (potential SSRF)", host),
				"URL_BLOCKED_HOST")
		}
	}

	// Check if host is an IP address and validate against blocked ranges
	if ip := net.ParseIP(host); ip != nil {
		for _, network := range parsedBlockedRanges {
			if network.Contains(ip) {
				return NewValidationError("git.url",
					fmt.Sprintf("git URL IP '%s' is in a blocked range (potential SSRF)", host),
					"URL_BLOCKED_IP")
			}
		}
	}

	// Resolve hostname and check IP (defense in depth against DNS rebinding)
	ips, err := net.LookupIP(host)
	if err == nil {
		for _, ip := range ips {
			for _, network := range parsedBlockedRanges {
				if network.Contains(ip) {
					return NewValidationError("git.url",
						fmt.Sprintf("git URL resolves to blocked IP '%s' (potential SSRF)", ip),
						"URL_BLOCKED_RESOLVED_IP")
				}
			}
		}
	}

	return nil
}

// ValidateGitRef validates a Git ref to prevent command injection
func ValidateGitRef(ref string) error {
	if ref == "" {
		return nil // Empty ref uses default (main)
	}

	// Check for maximum length
	if len(ref) > 256 {
		return NewValidationError("git.ref", "git ref exceeds maximum length of 256 characters", "REF_TOO_LONG")
	}

	// Check for shell metacharacters
	if shellMetacharacters.MatchString(ref) {
		return NewValidationError("git.ref", "git ref contains invalid characters (potential injection)", "REF_INJECTION")
	}

	// Validate format
	if !validGitRefPattern.MatchString(ref) {
		return NewValidationError("git.ref",
			"git ref contains invalid characters (only alphanumeric, dash, underscore, dot, and forward slash allowed)",
			"REF_INVALID_FORMAT")
	}

	// Check for path traversal
	if strings.Contains(ref, "..") {
		return NewValidationError("git.ref", "git ref contains path traversal sequence", "REF_PATH_TRAVERSAL")
	}

	return nil
}

// ValidateGitPath validates a Git path to prevent path traversal (BLUE-005)
func ValidateGitPath(path string) error {
	if path == "" {
		return nil // Empty path means root of repo
	}

	// Check for maximum length
	if len(path) > 512 {
		return NewValidationError("git.path", "git path exceeds maximum length of 512 characters", "PATH_TOO_LONG")
	}

	// Check for shell metacharacters
	if shellMetacharacters.MatchString(path) {
		return NewValidationError("git.path", "git path contains invalid characters (potential injection)", "PATH_INJECTION")
	}

	// Check for path traversal sequences
	if strings.Contains(path, "..") {
		return NewValidationError("git.path", "git path contains path traversal sequence (..)", "PATH_TRAVERSAL")
	}

	// Clean the path and verify it doesn't escape
	cleaned := filepath.Clean(path)
	if strings.HasPrefix(cleaned, "..") || strings.HasPrefix(cleaned, "/") {
		return NewValidationError("git.path", "git path attempts to escape repository root", "PATH_ESCAPE")
	}

	return nil
}

// ValidateSecurePath validates that a path is safe and contained within a base directory
func ValidateSecurePath(basePath, requestedPath string) error {
	if requestedPath == "" {
		return nil
	}

	// Reject absolute paths immediately
	if filepath.IsAbs(requestedPath) {
		return NewValidationError("path",
			fmt.Sprintf("absolute path '%s' not allowed", requestedPath),
			"PATH_ABSOLUTE")
	}

	// Check for path traversal sequences before joining
	if strings.Contains(requestedPath, "..") {
		return NewValidationError("path",
			fmt.Sprintf("path traversal sequence in '%s'", requestedPath),
			"PATH_TRAVERSAL")
	}

	// Clean and resolve the full path
	fullPath := filepath.Clean(filepath.Join(basePath, requestedPath))
	cleanBase := filepath.Clean(basePath)

	// Ensure the resolved path is still within the base path
	if !strings.HasPrefix(fullPath, cleanBase+string(filepath.Separator)) && fullPath != cleanBase {
		return NewValidationError("path",
			fmt.Sprintf("path '%s' escapes base directory", requestedPath),
			"PATH_ESCAPE")
	}

	return nil
}

// ValidateMinIOEndpoint validates a MinIO endpoint to prevent SSRF
func ValidateMinIOEndpoint(endpoint string) error {
	if endpoint == "" {
		return nil // Empty uses default
	}

	// Check for maximum length
	if len(endpoint) > 253 {
		return NewValidationError("minio.endpoint", "endpoint exceeds maximum length of 253 characters", "ENDPOINT_TOO_LONG")
	}

	// Check for shell metacharacters
	if shellMetacharacters.MatchString(endpoint) {
		return NewValidationError("minio.endpoint", "endpoint contains invalid characters (potential injection)", "ENDPOINT_INJECTION")
	}

	// Validate format
	if !validEndpointPattern.MatchString(endpoint) {
		return NewValidationError("minio.endpoint", "endpoint format is invalid", "ENDPOINT_INVALID_FORMAT")
	}

	// Extract host (remove port if present)
	host := endpoint
	if colonIdx := strings.LastIndex(endpoint, ":"); colonIdx != -1 {
		// Check if it's a port (not part of IPv6)
		if !strings.Contains(endpoint, "[") {
			host = endpoint[:colonIdx]
		}
	}

	// Check against blocked hosts
	hostLower := strings.ToLower(host)
	for _, blocked := range blockedHosts {
		if hostLower == blocked || strings.HasSuffix(hostLower, "."+blocked) {
			return NewValidationError("minio.endpoint",
				fmt.Sprintf("endpoint host '%s' is blocked", host),
				"ENDPOINT_BLOCKED_HOST")
		}
	}

	return nil
}

// ValidateBucketName validates an S3/MinIO bucket name
func ValidateBucketName(bucket string) error {
	if bucket == "" {
		return NewValidationError("bucket", "bucket name is required", "BUCKET_REQUIRED")
	}

	// Check for shell metacharacters
	if shellMetacharacters.MatchString(bucket) {
		return NewValidationError("bucket", "bucket name contains invalid characters (potential injection)", "BUCKET_INJECTION")
	}

	// Validate bucket name format (S3 naming rules)
	if !validBucketPattern.MatchString(bucket) {
		return NewValidationError("bucket",
			"bucket name must be 3-63 characters, lowercase alphanumeric, dots and hyphens only",
			"BUCKET_INVALID_FORMAT")
	}

	return nil
}

// ValidateObjectKey validates an S3/MinIO object key
func ValidateObjectKey(key string) error {
	if key == "" {
		return NewValidationError("key", "object key is required", "KEY_REQUIRED")
	}

	// Check for maximum length (S3 limit is 1024)
	if len(key) > 1024 {
		return NewValidationError("key", "object key exceeds maximum length of 1024 characters", "KEY_TOO_LONG")
	}

	// Check for shell metacharacters
	if shellMetacharacters.MatchString(key) {
		return NewValidationError("key", "object key contains invalid characters (potential injection)", "KEY_INJECTION")
	}

	// Check for path traversal
	if strings.Contains(key, "..") {
		return NewValidationError("key", "object key contains path traversal sequence", "KEY_PATH_TRAVERSAL")
	}

	// Validate format
	if !validObjectKeyPattern.MatchString(key) {
		return NewValidationError("key",
			"object key contains invalid characters",
			"KEY_INVALID_FORMAT")
	}

	return nil
}

// ValidateGitSource performs comprehensive validation of a Git source configuration
func ValidateGitSource(url, ref, path string) error {
	if err := ValidateGitURL(url); err != nil {
		return err
	}
	if err := ValidateGitRef(ref); err != nil {
		return err
	}
	if err := ValidateGitPath(path); err != nil {
		return err
	}
	return nil
}

// ValidateMinIOSource performs comprehensive validation of a MinIO source configuration
func ValidateMinIOSource(endpoint, bucket, key string) error {
	if err := ValidateMinIOEndpoint(endpoint); err != nil {
		return err
	}
	if err := ValidateBucketName(bucket); err != nil {
		return err
	}
	if err := ValidateObjectKey(key); err != nil {
		return err
	}
	return nil
}

// ValidateS3Source performs comprehensive validation of an S3 source configuration
func ValidateS3Source(bucket, key, region string) error {
	if err := ValidateBucketName(bucket); err != nil {
		return err
	}
	if err := ValidateObjectKey(key); err != nil {
		return err
	}
	// Region validation - basic check
	if region != "" && !regexp.MustCompile(`^[a-z]{2}-[a-z]+-[0-9]+$`).MatchString(region) {
		return NewValidationError("region", "invalid AWS region format", "REGION_INVALID_FORMAT")
	}
	return nil
}

// ValidateGCSBucket validates a GCS bucket name
// GCS bucket naming rules are similar to S3 but with some differences
func ValidateGCSBucket(bucket string) error {
	if bucket == "" {
		return NewValidationError("gcs.bucket", "GCS bucket name is required", "GCS_BUCKET_REQUIRED")
	}

	// Check for shell metacharacters
	if shellMetacharacters.MatchString(bucket) {
		return NewValidationError("gcs.bucket", "GCS bucket name contains invalid characters (potential injection)", "GCS_BUCKET_INJECTION")
	}

	// GCS bucket naming rules:
	// - 3-63 characters
	// - Start and end with alphanumeric
	// - Only lowercase letters, numbers, dashes, underscores, and dots
	// - Cannot start with "goog" prefix
	// - Cannot contain "google" or close misspellings
	gcsBucketPattern := regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{1,61}[a-z0-9]$`)
	if !gcsBucketPattern.MatchString(bucket) {
		return NewValidationError("gcs.bucket",
			"GCS bucket name must be 3-63 characters, lowercase alphanumeric with dots, dashes, and underscores",
			"GCS_BUCKET_INVALID_FORMAT")
	}

	// Check for reserved prefixes
	if strings.HasPrefix(bucket, "goog") {
		return NewValidationError("gcs.bucket", "GCS bucket name cannot start with 'goog'", "GCS_BUCKET_RESERVED_PREFIX")
	}

	// Check for "google" in the name
	if strings.Contains(bucket, "google") {
		return NewValidationError("gcs.bucket", "GCS bucket name cannot contain 'google'", "GCS_BUCKET_RESERVED_NAME")
	}

	return nil
}

// ValidateGCSKey validates a GCS object key/path
func ValidateGCSKey(key string) error {
	if key == "" {
		return NewValidationError("gcs.key", "GCS object key is required", "GCS_KEY_REQUIRED")
	}

	// Check for maximum length (GCS limit is 1024 bytes)
	if len(key) > 1024 {
		return NewValidationError("gcs.key", "GCS object key exceeds maximum length of 1024 characters", "GCS_KEY_TOO_LONG")
	}

	// Check for shell metacharacters
	if shellMetacharacters.MatchString(key) {
		return NewValidationError("gcs.key", "GCS object key contains invalid characters (potential injection)", "GCS_KEY_INJECTION")
	}

	// Check for path traversal
	if strings.Contains(key, "..") {
		return NewValidationError("gcs.key", "GCS object key contains path traversal sequence", "GCS_KEY_PATH_TRAVERSAL")
	}

	// GCS allows more characters than S3, but we restrict for security
	gcsKeyPattern := regexp.MustCompile(`^[a-zA-Z0-9!_.*'()/-]+$`)
	if !gcsKeyPattern.MatchString(key) {
		return NewValidationError("gcs.key",
			"GCS object key contains invalid characters",
			"GCS_KEY_INVALID_FORMAT")
	}

	return nil
}

// ValidateGCSSource performs comprehensive validation of a GCS source configuration
func ValidateGCSSource(bucket, key, project string) error {
	if err := ValidateGCSBucket(bucket); err != nil {
		return err
	}
	if err := ValidateGCSKey(key); err != nil {
		return err
	}
	// Project ID validation - alphanumeric and hyphens, 6-30 chars
	if project != "" {
		projectPattern := regexp.MustCompile(`^[a-z][a-z0-9-]{4,28}[a-z0-9]$`)
		if !projectPattern.MatchString(project) {
			return NewValidationError("gcs.project", "invalid GCP project ID format", "GCS_PROJECT_INVALID_FORMAT")
		}
	}
	return nil
}

// ValidateGitHubOwner validates a GitHub repository owner
func ValidateGitHubOwner(owner string) error {
	if owner == "" {
		return NewValidationError("github.owner", "GitHub owner is required", "GITHUB_OWNER_REQUIRED")
	}

	// GitHub username/org rules: alphanumeric and hyphens, 1-39 chars, cannot start/end with hyphen
	ownerPattern := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,37}[a-zA-Z0-9])?$`)
	if !ownerPattern.MatchString(owner) {
		return NewValidationError("github.owner",
			"GitHub owner must be 1-39 characters, alphanumeric with hyphens (not at start/end)",
			"GITHUB_OWNER_INVALID_FORMAT")
	}

	return nil
}

// ValidateGitHubRepo validates a GitHub repository name
func ValidateGitHubRepo(repo string) error {
	if repo == "" {
		return NewValidationError("github.repo", "GitHub repository name is required", "GITHUB_REPO_REQUIRED")
	}

	// GitHub repo names: alphanumeric, hyphens, underscores, dots, 1-100 chars
	repoPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]{1,100}$`)
	if !repoPattern.MatchString(repo) {
		return NewValidationError("github.repo",
			"GitHub repository name must be 1-100 characters, alphanumeric with dots, hyphens, and underscores",
			"GITHUB_REPO_INVALID_FORMAT")
	}

	// Cannot be just dots
	if repo == "." || repo == ".." {
		return NewValidationError("github.repo", "GitHub repository name cannot be '.' or '..'", "GITHUB_REPO_INVALID_NAME")
	}

	return nil
}

// ValidateGitHubSource performs comprehensive validation of a GitHub source configuration
func ValidateGitHubSource(owner, repo, ref, path string) error {
	if err := ValidateGitHubOwner(owner); err != nil {
		return err
	}
	if err := ValidateGitHubRepo(repo); err != nil {
		return err
	}
	// Reuse git ref validation
	if err := ValidateGitRef(ref); err != nil {
		return err
	}
	// Reuse git path validation
	if err := ValidateGitPath(path); err != nil {
		return err
	}
	return nil
}
