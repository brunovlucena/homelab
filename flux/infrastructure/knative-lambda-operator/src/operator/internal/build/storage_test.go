// Package build provides comprehensive tests for the generic storage abstraction.
// These tests validate BACKEND-002: Build Context Management storage backends.
package build

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

// =============================================================================
// StorageSelector Tests
// =============================================================================

func TestBackend002_StorageSelector_SelectStore(t *testing.T) {
	tests := []struct {
		name           string
		annotation     string
		contextSize    int
		s3Configured   bool
		gcsConfigured  bool
		expectedStore  string
		expectError    bool
		errorContains  string
	}{
		{
			name:          "Default to ConfigMap for small context",
			contextSize:   100 * 1024, // 100KB
			expectedStore: StorageBackendConfigMap,
		},
		{
			name:          "Select ConfigMap via annotation",
			annotation:    StorageBackendConfigMap,
			contextSize:   100 * 1024,
			expectedStore: StorageBackendConfigMap,
		},
		{
			name:          "Select S3 via annotation",
			annotation:    StorageBackendS3,
			contextSize:   100 * 1024,
			s3Configured:  true,
			expectedStore: StorageBackendS3,
		},
		// GCS tests skipped because GCS client requires real credentials
		// {
		// 	name:          "Select GCS via annotation",
		// 	annotation:    StorageBackendGCS,
		// 	contextSize:   100 * 1024,
		// 	gcsConfigured: true,
		// 	expectedStore: StorageBackendGCS,
		// },
		{
			name:          "Auto-select S3 for large context",
			contextSize:   1024 * 1024, // 1MB > 768KB limit
			s3Configured:  true,
			expectedStore: StorageBackendS3,
		},
		// GCS tests skipped because GCS client requires real credentials
		// {
		// 	name:          "Auto-select GCS when S3 not configured and context is large",
		// 	contextSize:   1024 * 1024,
		// 	gcsConfigured: true,
		// 	expectedStore: StorageBackendGCS,
		// },
		{
			name:          "Error when ConfigMap requested but context too large",
			annotation:    StorageBackendConfigMap,
			contextSize:   1024 * 1024, // 1MB
			expectError:   true,
			errorContains: "exceeds ConfigMap limit",
		},
		{
			name:          "Error when S3 requested but not configured",
			annotation:    StorageBackendS3,
			contextSize:   100 * 1024,
			expectError:   true,
			errorContains: "S3 storage backend not configured",
		},
		{
			name:          "Error when GCS requested but not configured",
			annotation:    StorageBackendGCS,
			contextSize:   100 * 1024,
			expectError:   true,
			errorContains: "GCS storage backend not configured",
		},
		{
			name:          "Error when context too large and no object storage configured",
			contextSize:   1024 * 1024,
			expectError:   true,
			errorContains: "no object storage backend is configured",
		},
		{
			name:          "Error for unsupported storage backend",
			annotation:    "invalid-backend",
			contextSize:   100 * 1024,
			expectError:   true,
			errorContains: "unsupported storage backend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client
			scheme := runtime.NewScheme()
			_ = corev1.AddToScheme(scheme)
			_ = lambdav1alpha1.AddToScheme(scheme)
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			// Configure storage backends
			var s3Config *S3Config
			var gcsConfig *GCSConfig

			if tt.s3Configured {
				s3Config = &S3Config{
					Bucket:          "test-bucket",
					Endpoint:        "minio.test.svc:9000",
					Region:          "us-east-1",
					AccessKeyID:     "test-access-key",
					SecretAccessKey: "test-secret-key",
					UseSSL:          false,
				}
			}

			if tt.gcsConfigured {
				// GCS requires real credentials, so we'll skip creating the store in tests
				// In real scenarios, this would work with workload identity
			}

			// Create selector
			selector := NewStorageSelector(fakeClient, scheme, s3Config, gcsConfig)

			// Create test lambda
			lambda := &lambdav1alpha1.LambdaFunction{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-lambda",
					Namespace: "default",
				},
			}

			if tt.annotation != "" {
				lambda.Annotations = map[string]string{
					StorageAnnotation: tt.annotation,
				}
			}

			// Select store
			store, err := selector.SelectStore(lambda, tt.contextSize)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, store)
			assert.Equal(t, tt.expectedStore, store.Name())
		})
	}
}

// =============================================================================
// ConfigMapStore Tests
// =============================================================================

func TestBackend002_ConfigMapStore_Save(t *testing.T) {
	tests := []struct {
		name          string
		contextData   []byte
		expectError   bool
		errorContains string
	}{
		{
			name:        "Save small context",
			contextData: bytes.Repeat([]byte("x"), 100*1024), // 100KB
		},
		{
			name:        "Save context at limit",
			contextData: bytes.Repeat([]byte("x"), ConfigMapSizeLimit-1), // Just under limit
		},
		{
			name:          "Reject context over limit",
			contextData:   bytes.Repeat([]byte("x"), ConfigMapSizeLimit+1), // Over limit
			expectError:   true,
			errorContains: "exceeds ConfigMap limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client
			scheme := runtime.NewScheme()
			_ = corev1.AddToScheme(scheme)
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			store := NewConfigMapStore(fakeClient, scheme)

			meta := BuildContextMeta{
				LambdaName:      "test-lambda",
				LambdaNamespace: "default",
				ContentHash:     "abc123def456",
				CreatedAt:       time.Now(),
			}

			location, err := store.Save(context.Background(), "test-key", tt.contextData, meta)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, location)
			assert.Equal(t, StorageBackendConfigMap, location.Backend)
			assert.Equal(t, "test-lambda-build-context", location.ConfigMapName)
			assert.Equal(t, "abc123def456", location.ContentHash)
			assert.Equal(t, "abc123def456"[:12], location.ImageTag)
		})
	}
}

func TestBackend002_ConfigMapStore_Cleanup(t *testing.T) {
	// Create fake client
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	store := NewConfigMapStore(fakeClient, scheme)

	// Cleanup should be a no-op for ConfigMap (returns 0, nil)
	cleaned, err := store.Cleanup(context.Background(), 24*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 0, cleaned)
}

func TestBackend002_ConfigMapStore_Name(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	store := NewConfigMapStore(fakeClient, scheme)
	assert.Equal(t, StorageBackendConfigMap, store.Name())
}

// =============================================================================
// S3Store Tests
// =============================================================================

func TestBackend002_S3Store_NewS3Store(t *testing.T) {
	tests := []struct {
		name          string
		config        *S3Config
		expectError   bool
		errorContains string
	}{
		{
			name: "Create with valid config",
			config: &S3Config{
				Bucket:          "test-bucket",
				Endpoint:        "minio.test.svc:9000",
				Region:          "us-east-1",
				AccessKeyID:     "access-key",
				SecretAccessKey: "secret-key",
				UseSSL:          false,
			},
		},
		{
			name: "Create with AWS S3 (no endpoint)",
			config: &S3Config{
				Bucket:          "test-bucket",
				Region:          "us-west-2",
				AccessKeyID:     "access-key",
				SecretAccessKey: "secret-key",
				UseSSL:          true,
			},
		},
		{
			name: "Error when bucket is empty",
			config: &S3Config{
				Bucket: "",
			},
			expectError:   true,
			errorContains: "bucket is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewS3Store(tt.config)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, store)
			assert.Equal(t, StorageBackendS3, store.Name())
			assert.Equal(t, tt.config.Bucket, store.GetBucket())
		})
	}
}

func TestBackend002_S3Store_Name(t *testing.T) {
	store := &S3Store{bucket: "test"}
	assert.Equal(t, StorageBackendS3, store.Name())
}

// =============================================================================
// GCSStore Tests
// =============================================================================

func TestBackend002_GCSStore_Validation(t *testing.T) {
	// Test GCS config validation
	tests := []struct {
		name          string
		config        *GCSConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "Error when bucket is empty",
			config: &GCSConfig{
				Bucket: "",
			},
			expectError:   true,
			errorContains: "bucket is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGCSStore(context.Background(), tt.config)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			}
		})
	}
}

func TestBackend002_GCSStore_Name(t *testing.T) {
	store := &GCSStore{bucket: "test"}
	assert.Equal(t, StorageBackendGCS, store.Name())
}

// =============================================================================
// BuildContextLocation Tests
// =============================================================================

func TestBackend002_BuildContextLocation_ConfigMap(t *testing.T) {
	location := &BuildContextLocation{
		Backend:       StorageBackendConfigMap,
		ConfigMapName: "my-lambda-build-context",
		ContentHash:   "abc123def456789",
		ImageTag:      "abc123def456",
	}

	assert.Equal(t, StorageBackendConfigMap, location.Backend)
	assert.Equal(t, "my-lambda-build-context", location.ConfigMapName)
	assert.Empty(t, location.Bucket)
	assert.Empty(t, location.Key)
}

func TestBackend002_BuildContextLocation_S3(t *testing.T) {
	location := &BuildContextLocation{
		Backend:     StorageBackendS3,
		Bucket:      "my-bucket",
		Key:         "build-context/my-lambda/context.tar.gz",
		Endpoint:    "s3.us-east-1.amazonaws.com",
		Region:      "us-east-1",
		ContentHash: "abc123def456789",
		ImageTag:    "abc123def456",
	}

	assert.Equal(t, StorageBackendS3, location.Backend)
	assert.Equal(t, "my-bucket", location.Bucket)
	assert.Equal(t, "build-context/my-lambda/context.tar.gz", location.Key)
	assert.Equal(t, "us-east-1", location.Region)
	assert.Empty(t, location.ConfigMapName)
}

func TestBackend002_BuildContextLocation_GCS(t *testing.T) {
	location := &BuildContextLocation{
		Backend:     StorageBackendGCS,
		Bucket:      "my-gcs-bucket",
		Key:         "build-context/my-lambda/context.tar.gz",
		ContentHash: "abc123def456789",
		ImageTag:    "abc123def456",
	}

	assert.Equal(t, StorageBackendGCS, location.Backend)
	assert.Equal(t, "my-gcs-bucket", location.Bucket)
	assert.Equal(t, "build-context/my-lambda/context.tar.gz", location.Key)
	assert.Empty(t, location.ConfigMapName)
	assert.Empty(t, location.Endpoint)
}

// =============================================================================
// BuildContextMeta Tests
// =============================================================================

func TestBackend002_BuildContextMeta(t *testing.T) {
	now := time.Now()
	meta := BuildContextMeta{
		LambdaName:      "test-lambda",
		LambdaNamespace: "production",
		ContentHash:     "sha256:abc123",
		CreatedAt:       now,
	}

	assert.Equal(t, "test-lambda", meta.LambdaName)
	assert.Equal(t, "production", meta.LambdaNamespace)
	assert.Equal(t, "sha256:abc123", meta.ContentHash)
	assert.Equal(t, now, meta.CreatedAt)
}

// =============================================================================
// Constants Tests
// =============================================================================

func TestBackend002_Constants(t *testing.T) {
	assert.Equal(t, "lambda.knative.io/build-context-storage", StorageAnnotation)
	assert.Equal(t, 768*1024, ConfigMapSizeLimit) // 768KB
	assert.Equal(t, "configmap", StorageBackendConfigMap)
	assert.Equal(t, "s3", StorageBackendS3)
	assert.Equal(t, "gcs", StorageBackendGCS)
}

// =============================================================================
// Integration-style Tests (using fake clients)
// =============================================================================

func TestBackend002_ConfigMapStore_CreateAndUpdate(t *testing.T) {
	// Create fake client with initial namespace
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "default"},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(ns).
		Build()

	store := NewConfigMapStore(fakeClient, scheme)

	meta := BuildContextMeta{
		LambdaName:      "test-lambda",
		LambdaNamespace: "default",
		ContentHash:     "abc123def456789012345678901234567890", // Full SHA-256 style hash
		CreatedAt:       time.Now(),
	}

	// First save - creates ConfigMap
	location1, err := store.Save(context.Background(), "key1", []byte("data1"), meta)
	require.NoError(t, err)
	assert.Equal(t, "test-lambda-build-context", location1.ConfigMapName)
	assert.Equal(t, "abc123def456", location1.ImageTag) // First 12 chars

	// Update metadata
	meta.ContentHash = "xyz789abc123456789012345678901234567890"

	// Second save - updates ConfigMap
	location2, err := store.Save(context.Background(), "key2", []byte("data2"), meta)
	require.NoError(t, err)
	assert.Equal(t, location1.ConfigMapName, location2.ConfigMapName)
	assert.Equal(t, "xyz789abc123456789012345678901234567890", location2.ContentHash)
	assert.Equal(t, "xyz789abc123", location2.ImageTag) // First 12 chars
}
