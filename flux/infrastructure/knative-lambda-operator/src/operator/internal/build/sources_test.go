// Package build provides comprehensive tests for source fetching capabilities.
// These tests validate BACKEND-002: Build Context Management source backends.
package build

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

// =============================================================================
// GitHubFetcher Tests
// =============================================================================

func TestBackend002_GitHubFetcher_ValidateSource(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	fetcher := NewGitHubFetcher(fakeClient, "default")

	tests := []struct {
		name          string
		spec          *GitHubSource
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid owner and repo",
			spec: &GitHubSource{
				Owner: "brunovlucena",
				Repo:  "homelab",
			},
		},
		{
			name: "Valid with ref and path",
			spec: &GitHubSource{
				Owner: "brunovlucena",
				Repo:  "homelab",
				Ref:   "main",
				Path:  "src/lambda",
			},
		},
		{
			name: "Valid archive URL",
			spec: &GitHubSource{
				ArchiveURL: "https://github.com/owner/repo/archive/main.zip",
			},
		},
		{
			name: "Error when owner missing",
			spec: &GitHubSource{
				Repo: "homelab",
			},
			expectError:   true,
			errorContains: "owner is required",
		},
		{
			name: "Error when repo missing",
			spec: &GitHubSource{
				Owner: "brunovlucena",
			},
			expectError:   true,
			errorContains: "repo is required",
		},
		{
			name: "Error for path traversal in path",
			spec: &GitHubSource{
				Owner: "brunovlucena",
				Repo:  "homelab",
				Path:  "../../../etc/passwd",
			},
			expectError:   true,
			errorContains: "path traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fetcher.validateSource(tt.spec)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestBackend002_GitHubFetcher_BuildArchiveURL(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	fetcher := NewGitHubFetcher(fakeClient, "default")

	tests := []struct {
		name        string
		spec        *GitHubSource
		expectedURL string
	}{
		{
			name: "Default ref (HEAD)",
			spec: &GitHubSource{
				Owner: "brunovlucena",
				Repo:  "homelab",
			},
			expectedURL: "https://api.github.com/repos/brunovlucena/homelab/zipball/HEAD",
		},
		{
			name: "With branch ref",
			spec: &GitHubSource{
				Owner: "brunovlucena",
				Repo:  "homelab",
				Ref:   "main",
			},
			expectedURL: "https://api.github.com/repos/brunovlucena/homelab/zipball/main",
		},
		{
			name: "With tag ref",
			spec: &GitHubSource{
				Owner: "brunovlucena",
				Repo:  "homelab",
				Ref:   "v1.0.0",
			},
			expectedURL: "https://api.github.com/repos/brunovlucena/homelab/zipball/v1.0.0",
		},
		{
			name: "With commit SHA",
			spec: &GitHubSource{
				Owner: "brunovlucena",
				Repo:  "homelab",
				Ref:   "abc123def456",
			},
			expectedURL: "https://api.github.com/repos/brunovlucena/homelab/zipball/abc123def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fetcher.buildArchiveURL(tt.spec)
			assert.Equal(t, tt.expectedURL, url)
		})
	}
}

func TestBackend002_GitHubFetcher_GetToken(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	tests := []struct {
		name          string
		secretData    map[string][]byte
		expectError   bool
		errorContains string
		expectedToken string
	}{
		{
			name: "Token from 'token' key",
			secretData: map[string][]byte{
				"token": []byte("ghp_test123"),
			},
			expectedToken: "ghp_test123",
		},
		{
			name: "Token from 'github-token' key",
			secretData: map[string][]byte{
				"github-token": []byte("ghp_test456"),
			},
			expectedToken: "ghp_test456",
		},
		{
			name: "Token from 'GITHUB_TOKEN' key",
			secretData: map[string][]byte{
				"GITHUB_TOKEN": []byte("ghp_test789"),
			},
			expectedToken: "ghp_test789",
		},
		{
			name: "Token from 'password' key",
			secretData: map[string][]byte{
				"password": []byte("ghp_password"),
			},
			expectedToken: "ghp_password",
		},
		{
			name:          "Error when secret has no token",
			secretData:    map[string][]byte{"other": []byte("value")},
			expectError:   true,
			errorContains: "does not contain a GitHub token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "github-secret",
					Namespace: "default",
				},
				Data: tt.secretData,
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(secret).
				Build()

			fetcher := NewGitHubFetcher(fakeClient, "default")

			token, err := fetcher.getToken(context.Background(), "github-secret")

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

func TestBackend002_GitHubFetcher_ExtractFromArchive(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	fetcher := NewGitHubFetcher(fakeClient, "default")

	// Create a test zip archive similar to GitHub's format
	createTestArchive := func(files map[string]string) []byte {
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)

		for name, content := range files {
			f, _ := w.Create(name)
			io.WriteString(f, content)
		}

		w.Close()
		return buf.Bytes()
	}

	tests := []struct {
		name          string
		archiveFiles  map[string]string
		path          string
		language      string
		expectedCode  string
		expectError   bool
		errorContains string
	}{
		{
			name: "Extract Python file from root",
			archiveFiles: map[string]string{
				"owner-repo-abc123/main.py":   "def handler(): pass",
				"owner-repo-abc123/README.md": "# Test",
			},
			language:     "python",
			expectedCode: "def handler(): pass",
		},
		{
			name: "Extract Node.js file from root",
			archiveFiles: map[string]string{
				"owner-repo-abc123/index.js": "module.exports = {}",
			},
			language:     "nodejs",
			expectedCode: "module.exports = {}",
		},
		{
			name: "Extract Go file from root",
			archiveFiles: map[string]string{
				"owner-repo-abc123/main.go": "package main",
			},
			language:     "go",
			expectedCode: "package main",
		},
		{
			name: "Extract from subdirectory",
			archiveFiles: map[string]string{
				"owner-repo-abc123/src/lambda/main.py": "def handler(): pass",
			},
			path:         "src/lambda",
			language:     "python",
			expectedCode: "def handler(): pass",
		},
		{
			name: "Error when file not found",
			archiveFiles: map[string]string{
				"owner-repo-abc123/other.txt": "content",
			},
			language:      "python",
			expectError:   true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive := createTestArchive(tt.archiveFiles)

			code, filename, err := fetcher.extractFromArchive(archive, tt.path, tt.language)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCode, string(code))
			assert.Equal(t, getSourceFilename(tt.language), filename)
		})
	}
}

func TestBackend002_GitHubFetcher_DownloadArchive(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	fetcher := NewGitHubFetcher(fakeClient, "default")

	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		token          string
		expectError    bool
		errorContains  string
	}{
		{
			name: "Successful download",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("archive-data"))
			},
		},
		{
			name: "Download with auth token",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("archive-data"))
			},
			token: "test-token",
		},
		{
			name: "Handle 404 error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message": "Not Found"}`))
			},
			expectError:   true,
			errorContains: "404",
		},
		{
			name: "Handle 401 error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"message": "Bad credentials"}`))
			},
			expectError:   true,
			errorContains: "401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			data, err := fetcher.downloadArchive(context.Background(), server.URL, tt.token)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, data)
		})
	}
}

// =============================================================================
// GCSFetcher Tests
// =============================================================================

func TestBackend002_GCSFetcher_ValidateSource(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	fetcher := NewGCSFetcher(fakeClient, "default")

	tests := []struct {
		name          string
		spec          *lambdav1alpha1.GCSSource
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid bucket and key",
			spec: &lambdav1alpha1.GCSSource{
				Bucket: "my-bucket",
				Key:    "path/to/source.py",
			},
		},
		{
			name: "Valid with project",
			spec: &lambdav1alpha1.GCSSource{
				Bucket:  "my-bucket",
				Key:     "path/to/source.py",
				Project: "my-project",
			},
		},
		{
			name: "Error when bucket is empty",
			spec: &lambdav1alpha1.GCSSource{
				Key: "path/to/source.py",
			},
			expectError:   true,
			errorContains: "bucket is required",
		},
		{
			name: "Error when key is empty",
			spec: &lambdav1alpha1.GCSSource{
				Bucket: "my-bucket",
			},
			expectError:   true,
			errorContains: "key is required",
		},
		{
			name: "Error for path traversal in key",
			spec: &lambdav1alpha1.GCSSource{
				Bucket: "my-bucket",
				Key:    "../../../etc/passwd",
			},
			expectError:   true,
			errorContains: "path traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fetcher.validateSource(tt.spec)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			require.NoError(t, err)
		})
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestBackend002_GetSourceFilename(t *testing.T) {
	tests := []struct {
		language string
		expected string
	}{
		{"python", "main.py"},
		{"python3", "main.py"},
		{"nodejs", "index.js"},
		{"node", "index.js"},
		{"javascript", "index.js"},
		{"go", "main.go"},
		{"golang", "main.go"},
		{"unknown", "main.py"}, // Default to Python
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			result := getSourceFilename(tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBackend002_TruncateList(t *testing.T) {
	tests := []struct {
		name     string
		list     []string
		max      int
		expected []string
	}{
		{
			name:     "List shorter than max",
			list:     []string{"a", "b"},
			max:      5,
			expected: []string{"a", "b"},
		},
		{
			name:     "List equal to max",
			list:     []string{"a", "b", "c"},
			max:      3,
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "List longer than max",
			list:     []string{"a", "b", "c", "d", "e"},
			max:      3,
			expected: []string{"a", "b", "c", "... and 2 more"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateList(tt.list, tt.max)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// parseGitHubURL Tests
// =============================================================================

func TestBackend002_ParseGitHubURL(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedOwner string
		expectedRepo  string
	}{
		{
			name:          "HTTPS URL",
			url:           "https://github.com/brunovlucena/homelab",
			expectedOwner: "brunovlucena",
			expectedRepo:  "homelab",
		},
		{
			name:          "HTTPS URL with .git",
			url:           "https://github.com/brunovlucena/homelab.git",
			expectedOwner: "brunovlucena",
			expectedRepo:  "homelab",
		},
		{
			name:          "SSH URL",
			url:           "git@github.com:brunovlucena/homelab.git",
			expectedOwner: "brunovlucena",
			expectedRepo:  "homelab",
		},
		{
			name:          "Short URL",
			url:           "github.com/brunovlucena/homelab",
			expectedOwner: "brunovlucena",
			expectedRepo:  "homelab",
		},
		{
			name:          "Invalid URL",
			url:           "not-a-github-url",
			expectedOwner: "",
			expectedRepo:  "",
		},
		{
			name:          "Incomplete URL",
			url:           "https://github.com/brunovlucena",
			expectedOwner: "",
			expectedRepo:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo := parseGitHubURL(tt.url)
			assert.Equal(t, tt.expectedOwner, owner)
			assert.Equal(t, tt.expectedRepo, repo)
		})
	}
}

// =============================================================================
// GitHubSource Tests
// =============================================================================

func TestBackend002_GitHubSource_Struct(t *testing.T) {
	source := &GitHubSource{
		Owner:      "brunovlucena",
		Repo:       "homelab",
		Ref:        "main",
		Path:       "src/lambda",
		ArchiveURL: "",
		SecretRef:  &corev1.LocalObjectReference{Name: "github-secret"},
	}

	assert.Equal(t, "brunovlucena", source.Owner)
	assert.Equal(t, "homelab", source.Repo)
	assert.Equal(t, "main", source.Ref)
	assert.Equal(t, "src/lambda", source.Path)
	assert.Empty(t, source.ArchiveURL)
	assert.Equal(t, "github-secret", source.SecretRef.Name)
}

// =============================================================================
// Source Type Constants Tests
// =============================================================================

func TestBackend002_SourceTypeConstants(t *testing.T) {
	assert.Equal(t, "github", SourceTypeGitHub)
	assert.Equal(t, "gcs", SourceTypeGCS)
}
