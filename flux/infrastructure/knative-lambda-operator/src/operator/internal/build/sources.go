// Package build provides source fetching capabilities for build context management.
// This file implements source backends for BACKEND-002: Build Context Management.
//
// Supported source backends:
//   - GitHub: Fetch from GitHub archive API (branch, tag, commit, or archive URL)
//   - GCS: Fetch from Google Cloud Storage bucket
//   - Git: Clone from Git repository (existing in manager.go)
//   - S3/MinIO: Fetch from S3-compatible storage (existing in manager.go)
//   - Inline: Use inline code from spec (existing in manager.go)
package build

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
	"github.com/brunovlucena/knative-lambda-operator/internal/validation"
)

const (
	// SourceTypeGitHub is the source type for GitHub
	SourceTypeGitHub = "github"
	// SourceTypeGCS is the source type for GCS
	SourceTypeGCS = "gcs"

	// GitHubArchiveTimeout is the timeout for GitHub archive downloads
	GitHubArchiveTimeout = 60 * time.Second
	// GCSDownloadTimeout is the timeout for GCS downloads
	GCSDownloadTimeout = 60 * time.Second

	// MaxArchiveSize is the maximum size for downloaded archives (50MB)
	MaxArchiveSize = 50 * 1024 * 1024
)

// GitHubSource defines GitHub source configuration
// This is an extension to the existing SourceSpec to support GitHub-specific features
type GitHubSource struct {
	// Owner is the GitHub repository owner (user or organization)
	Owner string `json:"owner,omitempty"`
	// Repo is the GitHub repository name
	Repo string `json:"repo,omitempty"`
	// Ref is the branch, tag, or commit SHA (defaults to default branch)
	Ref string `json:"ref,omitempty"`
	// Path is the path within the repository (optional)
	Path string `json:"path,omitempty"`
	// ArchiveURL is a direct URL to a GitHub archive (optional, overrides owner/repo)
	ArchiveURL string `json:"archiveUrl,omitempty"`
	// SecretRef contains GitHub token for private repos
	SecretRef *corev1.LocalObjectReference `json:"secretRef,omitempty"`
}

// GitHubFetcher fetches source code from GitHub
type GitHubFetcher struct {
	client    client.Client
	namespace string
}

// NewGitHubFetcher creates a new GitHub source fetcher
func NewGitHubFetcher(k8sClient client.Client, namespace string) *GitHubFetcher {
	return &GitHubFetcher{
		client:    k8sClient,
		namespace: namespace,
	}
}

// Fetch downloads source code from GitHub and returns the source code bytes and filename
func (f *GitHubFetcher) Fetch(ctx context.Context, spec *GitHubSource, language string) ([]byte, string, error) {
	startTime := time.Now()
	defer func() {
		// Record fetch attempt with duration for observability
		duration := time.Since(startTime).Seconds()
		buildContextCreationDuration.WithLabelValues(SourceTypeGitHub).Observe(duration)
	}()

	// Validate GitHub source
	if err := f.validateSource(spec); err != nil {
		RecordSourceFetch(SourceTypeGitHub, "validation_error")
		return nil, "", fmt.Errorf("invalid GitHub source: %w", err)
	}

	// Get GitHub token if secretRef is provided
	token := ""
	if spec.SecretRef != nil {
		var err error
		token, err = f.getToken(ctx, spec.SecretRef.Name)
		if err != nil {
			RecordSourceFetch(SourceTypeGitHub, "auth_error")
			return nil, "", fmt.Errorf("failed to get GitHub token: %w", err)
		}
	}

	// Build archive URL
	archiveURL := spec.ArchiveURL
	if archiveURL == "" {
		archiveURL = f.buildArchiveURL(spec)
	}

	// Download archive
	archiveData, err := f.downloadArchive(ctx, archiveURL, token)
	if err != nil {
		RecordSourceFetch(SourceTypeGitHub, "download_error")
		return nil, "", fmt.Errorf("failed to download GitHub archive: %w", err)
	}

	// Extract source code from archive
	sourceCode, filename, err := f.extractFromArchive(archiveData, spec.Path, language)
	if err != nil {
		RecordSourceFetch(SourceTypeGitHub, "extract_error")
		return nil, "", fmt.Errorf("failed to extract source from GitHub archive: %w", err)
	}

	RecordSourceFetch(SourceTypeGitHub, "success")
	return sourceCode, filename, nil
}

// validateSource validates the GitHub source configuration
func (f *GitHubFetcher) validateSource(spec *GitHubSource) error {
	if spec.ArchiveURL != "" {
		// Validate archive URL
		if err := validation.ValidateGitURL(spec.ArchiveURL); err != nil {
			return fmt.Errorf("invalid archive URL: %w", err)
		}
		return nil
	}

	if spec.Owner == "" {
		return fmt.Errorf("owner is required when archiveUrl is not specified")
	}
	if spec.Repo == "" {
		return fmt.Errorf("repo is required when archiveUrl is not specified")
	}

	// Validate path if provided
	if spec.Path != "" {
		if err := validation.ValidateGitPath(spec.Path); err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}
	}

	return nil
}

// getToken retrieves the GitHub token from a secret
func (f *GitHubFetcher) getToken(ctx context.Context, secretName string) (string, error) {
	secret := &corev1.Secret{}
	if err := f.client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: f.namespace}, secret); err != nil {
		return "", fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	// Try different key names for the token
	for _, key := range []string{"token", "github-token", "GITHUB_TOKEN", "password"} {
		if token, ok := secret.Data[key]; ok && len(token) > 0 {
			return string(token), nil
		}
	}

	return "", fmt.Errorf("secret %s does not contain a GitHub token (expected keys: token, github-token, GITHUB_TOKEN, password)", secretName)
}

// buildArchiveURL constructs the GitHub archive URL
func (f *GitHubFetcher) buildArchiveURL(spec *GitHubSource) string {
	ref := spec.Ref
	if ref == "" {
		ref = "HEAD" // GitHub resolves HEAD to the default branch
	}

	// Use zipball URL for GitHub
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/zipball/%s", spec.Owner, spec.Repo, ref)
}

// httpClient is a shared HTTP client with proper timeouts
var httpClient = &http.Client{
	Timeout: GitHubArchiveTimeout,
	Transport: &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

// downloadArchive downloads the archive from the given URL
func (f *GitHubFetcher) downloadArchive(ctx context.Context, url, token string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, GitHubArchiveTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "knative-lambda-operator")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Execute request with configured client (has timeouts)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read archive with size limit
	data, err := io.ReadAll(io.LimitReader(resp.Body, MaxArchiveSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read archive: %w", err)
	}

	return data, nil
}

// extractFromArchive extracts the source code from a GitHub zip archive
func (f *GitHubFetcher) extractFromArchive(archiveData []byte, path, language string) ([]byte, string, error) {
	// GitHub returns a zip archive
	zipReader, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
	if err != nil {
		return nil, "", fmt.Errorf("failed to open zip archive: %w", err)
	}

	// GitHub archives have a root directory like "owner-repo-sha/"
	// Find the expected source file
	expectedFilename := getSourceFilename(language)

	// Build the target path
	targetSuffix := expectedFilename
	if path != "" {
		targetSuffix = filepath.Join(path, expectedFilename)
	}

	var sourceCode []byte
	var foundFiles []string

	for _, file := range zipReader.File {
		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Get path relative to root directory (remove "owner-repo-sha/" prefix)
		parts := strings.SplitN(file.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		relativePath := parts[1]
		foundFiles = append(foundFiles, relativePath)

		// Check if this is the file we're looking for
		if relativePath == targetSuffix || strings.HasSuffix(relativePath, "/"+expectedFilename) {
			rc, err := file.Open()
			if err != nil {
				return nil, "", fmt.Errorf("failed to open file %s: %w", file.Name, err)
			}

			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, "", fmt.Errorf("failed to read file %s: %w", file.Name, err)
			}

			sourceCode = data
			break
		}
	}

	if sourceCode == nil {
		return nil, "", fmt.Errorf("source file '%s' not found in GitHub archive (path: %s). Found %d files: %v",
			expectedFilename, path, len(foundFiles), truncateList(foundFiles, 10))
	}

	return sourceCode, expectedFilename, nil
}

// GCSFetcher fetches source code from Google Cloud Storage
type GCSFetcher struct {
	k8sClient client.Client
	namespace string
}

// NewGCSFetcher creates a new GCS source fetcher
func NewGCSFetcher(k8sClient client.Client, namespace string) *GCSFetcher {
	return &GCSFetcher{
		k8sClient: k8sClient,
		namespace: namespace,
	}
}

// Fetch downloads source code from GCS and returns the source code bytes and filename
func (f *GCSFetcher) Fetch(ctx context.Context, spec *lambdav1alpha1.GCSSource, language string) ([]byte, string, error) {
	startTime := time.Now()
	defer func() {
		// Record fetch attempt with duration for observability
		duration := time.Since(startTime).Seconds()
		buildContextCreationDuration.WithLabelValues(SourceTypeGCS).Observe(duration)
	}()

	// Validate GCS source
	if err := f.validateSource(spec); err != nil {
		RecordSourceFetch(SourceTypeGCS, "validation_error")
		return nil, "", fmt.Errorf("invalid GCS source: %w", err)
	}

	// Create GCS client
	gcsClient, err := f.createClient(ctx, spec)
	if err != nil {
		RecordSourceFetch(SourceTypeGCS, "client_error")
		return nil, "", fmt.Errorf("failed to create GCS client: %w", err)
	}
	defer gcsClient.Close()

	// Check if key is a directory
	key := spec.Key
	isDirectory := strings.HasSuffix(key, "/")

	if isDirectory {
		// Download all files from the directory and find the main source file
		sourceCode, filename, err := f.downloadDirectory(ctx, gcsClient, spec.Bucket, key, language)
		if err != nil {
			RecordSourceFetch(SourceTypeGCS, "download_error")
			return nil, "", err
		}
		RecordSourceFetch(SourceTypeGCS, "success")
		return sourceCode, filename, nil
	}

	// Download single file
	sourceCode, err := f.downloadFile(ctx, gcsClient, spec.Bucket, key)
	if err != nil {
		RecordSourceFetch(SourceTypeGCS, "download_error")
		return nil, "", err
	}

	filename := getSourceFilename(language)
	RecordSourceFetch(SourceTypeGCS, "success")
	return sourceCode, filename, nil
}

// validateSource validates the GCS source configuration
func (f *GCSFetcher) validateSource(spec *lambdav1alpha1.GCSSource) error {
	if spec.Bucket == "" {
		return fmt.Errorf("bucket is required")
	}
	if spec.Key == "" {
		return fmt.Errorf("key is required")
	}

	// Validate bucket name
	if err := validation.ValidateBucketName(spec.Bucket); err != nil {
		return fmt.Errorf("invalid bucket name: %w", err)
	}

	// Validate object key
	if err := validation.ValidateObjectKey(spec.Key); err != nil {
		return fmt.Errorf("invalid object key: %w", err)
	}

	return nil
}

// createClient creates a GCS client with credentials from secret if provided
func (f *GCSFetcher) createClient(ctx context.Context, spec *lambdav1alpha1.GCSSource) (*storage.Client, error) {
	// If secretRef is provided, get credentials from it
	if spec.SecretRef != nil {
		secret := &corev1.Secret{}
		if err := f.k8sClient.Get(ctx, types.NamespacedName{Name: spec.SecretRef.Name, Namespace: f.namespace}, secret); err != nil {
			return nil, fmt.Errorf("failed to get GCS credentials secret %s: %w", spec.SecretRef.Name, err)
		}

		// Look for credentials JSON
		var credJSON []byte
		for _, key := range []string{"key.json", "credentials.json", "service-account.json", "GOOGLE_APPLICATION_CREDENTIALS"} {
			if data, ok := secret.Data[key]; ok && len(data) > 0 {
				credJSON = data
				break
			}
		}

		if credJSON != nil {
			// Use credentials directly via option - no temp files or env vars (thread-safe)
			client, err := storage.NewClient(ctx, option.WithCredentialsJSON(credJSON))
			if err != nil {
				return nil, fmt.Errorf("failed to create GCS client with credentials: %w", err)
			}
			return client, nil
		}
	}

	// Create client with default credentials (workload identity, etc.)
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return client, nil
}

// downloadFile downloads a single file from GCS
func (f *GCSFetcher) downloadFile(ctx context.Context, client *storage.Client, bucket, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, GCSDownloadTimeout)
	defer cancel()

	obj := client.Bucket(bucket).Object(key)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open GCS object gs://%s/%s: %w", bucket, key, err)
	}
	defer reader.Close()

	data, err := io.ReadAll(io.LimitReader(reader, MaxArchiveSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read GCS object gs://%s/%s: %w", bucket, key, err)
	}

	return data, nil
}

// downloadDirectory downloads files from a GCS directory prefix and finds the main source file
func (f *GCSFetcher) downloadDirectory(ctx context.Context, client *storage.Client, bucket, prefix, language string) ([]byte, string, error) {
	expectedFilename := getSourceFilename(language)

	ctx, cancel := context.WithTimeout(ctx, GCSDownloadTimeout)
	defer cancel()

	bkt := client.Bucket(bucket)
	it := bkt.Objects(ctx, &storage.Query{Prefix: prefix})

	var sourceCode []byte
	var foundFiles []string

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("error listing GCS objects: %w", err)
		}

		// Skip directories
		if strings.HasSuffix(attrs.Name, "/") {
			continue
		}

		// Track found files
		relativePath := strings.TrimPrefix(attrs.Name, prefix)
		foundFiles = append(foundFiles, relativePath)

		// Check if this is the expected source file
		if relativePath == expectedFilename || filepath.Base(attrs.Name) == expectedFilename {
			reader, err := bkt.Object(attrs.Name).NewReader(ctx)
			if err != nil {
				return nil, "", fmt.Errorf("failed to open GCS object %s: %w", attrs.Name, err)
			}

			data, err := io.ReadAll(reader)
			reader.Close()
			if err != nil {
				return nil, "", fmt.Errorf("failed to read GCS object %s: %w", attrs.Name, err)
			}

			sourceCode = data
			break
		}
	}

	if sourceCode == nil {
		return nil, "", fmt.Errorf("source file '%s' not found in GCS bucket '%s' with prefix '%s'. Found %d files: %v",
			expectedFilename, bucket, prefix, len(foundFiles), truncateList(foundFiles, 10))
	}

	return sourceCode, expectedFilename, nil
}

// =============================================================================
// Archive Utilities
// =============================================================================

// CreateTarGzFromDirectory creates a tar.gz archive from a directory
func CreateTarGzFromDirectory(dir string) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == dir {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			// CRITICAL FIX: Use closure to ensure file is closed immediately,
			// not deferred until function returns (which would leak handles in loop)
			if err := addFileToArchive(tarWriter, path); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create tar archive: %w", err)
	}

	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// addFileToArchive adds a single file to the tar writer
// Extracted to fix file handle leak (defer in loop antipattern)
func addFileToArchive(tarWriter *tar.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	if _, err := io.Copy(tarWriter, file); err != nil {
		return fmt.Errorf("failed to copy file %s to archive: %w", path, err)
	}

	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// getSourceFilename returns the appropriate source filename for a language
func getSourceFilename(language string) string {
	switch strings.ToLower(language) {
	case "python", "python3":
		return "main.py"
	case "nodejs", "node", "javascript":
		return "index.js"
	case "go", "golang":
		return "main.go"
	default:
		return "main.py"
	}
}

// truncateList truncates a list to the specified maximum length
func truncateList(list []string, max int) []string {
	if len(list) <= max {
		return list
	}
	result := make([]string, max+1)
	copy(result, list[:max])
	result[max] = fmt.Sprintf("... and %d more", len(list)-max)
	return result
}
