// Package build provides comprehensive tests for the generic storage abstraction.
// These tests validate BACKEND-002: Build Context Management storage backends.
package build

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

// =============================================================================
// StorageSelector Tests
// =============================================================================

func TestBackend002_StorageSelector_SelectStore(t *testing.T) {
	tests := []struct {
		name          string
		annotation    string
		contextSize   int
		s3Configured  bool
		gcsConfigured bool
		expectedStore string
		expectError   bool
		errorContains string
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
		{
			name:          "Select S3 via annotation (uppercase)",
			annotation:    "S3",
			contextSize:   100 * 1024,
			s3Configured:  true,
			expectedStore: StorageBackendS3,
		},
		{
			name:          "Auto-select S3 for large context",
			contextSize:   1024 * 1024, // 1MB > 768KB limit
			s3Configured:  true,
			expectedStore: StorageBackendS3,
		},
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
		{
			name:          "Exact ConfigMap size limit boundary - under",
			contextSize:   ConfigMapSizeLimit - 1,
			expectedStore: StorageBackendConfigMap,
		},
		{
			name:          "Exact ConfigMap size limit boundary - at limit with annotation",
			annotation:    StorageBackendConfigMap,
			contextSize:   ConfigMapSizeLimit,
			expectedStore: StorageBackendConfigMap, // At limit is allowed (>= not used)
		},
		{
			name:          "Zero size context",
			contextSize:   0,
			expectedStore: StorageBackendConfigMap,
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
				// GCS requires real credentials, so we'll skip in tests
			}

			// Create selector
			selector := NewStorageSelector(fakeClient, scheme, s3Config, gcsConfig)
			defer selector.Close() // Test Close() doesn't panic

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

func TestBackend002_StorageSelector_Concurrent(t *testing.T) {
	// Test concurrent access to SelectStore
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = lambdav1alpha1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	s3Config := &S3Config{
		Bucket:          "test-bucket",
		Endpoint:        "minio.test.svc:9000",
		Region:          "us-east-1",
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
		UseSSL:          false,
	}

	selector := NewStorageSelector(fakeClient, scheme, s3Config, nil)
	defer selector.Close()

	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-lambda",
			Namespace: "default",
		},
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(size int) {
			defer wg.Done()
			_, err := selector.SelectStore(lambda, size)
			if err != nil {
				errCh <- err
			}
		}(i * 10000) // Varying sizes
	}

	wg.Wait()
	close(errCh)

	// Collect errors (some should fail for sizes > ConfigMapSizeLimit without S3)
	// This test mainly ensures no race conditions
	for range errCh {
		// Expected errors for large contexts
	}
}

func TestBackend002_StorageSelector_Getters(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	s3Config := &S3Config{
		Bucket:          "test-bucket",
		Endpoint:        "minio.test.svc:9000",
		Region:          "us-east-1",
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
	}

	selector := NewStorageSelector(fakeClient, scheme, s3Config, nil)
	defer selector.Close()

	// Test getters
	assert.NotNil(t, selector.GetConfigMapStore())
	assert.NotNil(t, selector.GetS3Store())
	assert.Nil(t, selector.GetGCSStore()) // GCS not configured
}

// =============================================================================
// ConfigMapStore Tests
// =============================================================================

func TestBackend002_ConfigMapStore_Save(t *testing.T) {
	tests := []struct {
		name          string
		contextData   []byte
		meta          BuildContextMeta
		expectError   bool
		errorContains string
	}{
		{
			name:        "Save small context",
			contextData: bytes.Repeat([]byte("x"), 100*1024), // 100KB
			meta: BuildContextMeta{
				LambdaName:      "test-lambda",
				LambdaNamespace: "default",
				ContentHash:     "abc123def456789012345678901234567890",
				CreatedAt:       time.Now(),
			},
		},
		{
			name:        "Save context at limit minus one",
			contextData: bytes.Repeat([]byte("x"), ConfigMapSizeLimit-1),
			meta: BuildContextMeta{
				LambdaName:      "test-lambda",
				LambdaNamespace: "default",
				ContentHash:     "hash123456789012",
				CreatedAt:       time.Now(),
			},
		},
		{
			name:        "Reject context over limit",
			contextData: bytes.Repeat([]byte("x"), ConfigMapSizeLimit+1),
			meta: BuildContextMeta{
				LambdaName:      "test-lambda",
				LambdaNamespace: "default",
				ContentHash:     "hash123456789012",
				CreatedAt:       time.Now(),
			},
			expectError:   true,
			errorContains: "exceeds ConfigMap limit",
		},
		{
			name:        "Save empty context",
			contextData: []byte{},
			meta: BuildContextMeta{
				LambdaName:      "empty-lambda",
				LambdaNamespace: "default",
				ContentHash:     "emptyhash1234567",
				CreatedAt:       time.Now(),
			},
		},
		{
			name:        "Save with short hash (less than 12 chars)",
			contextData: []byte("test"),
			meta: BuildContextMeta{
				LambdaName:      "short-hash",
				LambdaNamespace: "default",
				ContentHash:     "abc", // Less than 12 chars
				CreatedAt:       time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client
			scheme := runtime.NewScheme()
			_ = corev1.AddToScheme(scheme)
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			store := NewConfigMapStore(fakeClient, scheme)

			location, err := store.Save(context.Background(), "test-key", tt.contextData, tt.meta)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, location)
			assert.Equal(t, StorageBackendConfigMap, location.Backend)
			assert.Equal(t, tt.meta.LambdaName+BuildContextConfigMapSuffix, location.ConfigMapName)
			assert.Equal(t, tt.meta.ContentHash, location.ContentHash)

			// Verify image tag is computed correctly
			if len(tt.meta.ContentHash) > 12 {
				assert.Equal(t, tt.meta.ContentHash[:12], location.ImageTag)
			} else {
				assert.Equal(t, tt.meta.ContentHash, location.ImageTag)
			}
		})
	}
}

func TestBackend002_ConfigMapStore_SaveAndUpdate(t *testing.T) {
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
		LambdaName:      "update-test-lambda",
		LambdaNamespace: "default",
		ContentHash:     "firsthash123456789012345678901234",
		CreatedAt:       time.Now(),
	}

	// First save - creates ConfigMap
	location1, err := store.Save(context.Background(), "key1", []byte("data1"), meta)
	require.NoError(t, err)
	assert.Equal(t, "update-test-lambda-build-context", location1.ConfigMapName)
	assert.Equal(t, "firsthash123", location1.ImageTag)

	// Verify ConfigMap exists
	cm := &corev1.ConfigMap{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      location1.ConfigMapName,
		Namespace: "default",
	}, cm)
	require.NoError(t, err)
	assert.Equal(t, []byte("data1"), cm.BinaryData["context.tar.gz"])

	// Update metadata
	meta.ContentHash = "secondhash12345678901234567890123"

	// Second save - updates ConfigMap
	location2, err := store.Save(context.Background(), "key2", []byte("data2-updated"), meta)
	require.NoError(t, err)
	assert.Equal(t, location1.ConfigMapName, location2.ConfigMapName)
	assert.Equal(t, "secondhash12", location2.ImageTag)

	// Verify ConfigMap was updated
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      location2.ConfigMapName,
		Namespace: "default",
	}, cm)
	require.NoError(t, err)
	assert.Equal(t, []byte("data2-updated"), cm.BinaryData["context.tar.gz"])
}

func TestBackend002_ConfigMapStore_Cleanup(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	store := NewConfigMapStore(fakeClient, scheme)

	// Cleanup should be a no-op for ConfigMap (returns 0, nil)
	cleaned, err := store.Cleanup(context.Background(), 24*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 0, cleaned)

	// Test with different durations
	cleaned, err = store.Cleanup(context.Background(), time.Second)
	require.NoError(t, err)
	assert.Equal(t, 0, cleaned)

	cleaned, err = store.Cleanup(context.Background(), 0)
	require.NoError(t, err)
	assert.Equal(t, 0, cleaned)
}

func TestBackend002_ConfigMapStore_Name(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	store := NewConfigMapStore(fakeClient, scheme)
	assert.Equal(t, StorageBackendConfigMap, store.Name())
}

func TestBackend002_ConfigMapStore_SetOwnerReference(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = lambdav1alpha1.AddToScheme(scheme)

	// Create a ConfigMap that exists
	existingCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-lambda-build-context",
			Namespace: "default",
		},
		BinaryData: map[string][]byte{
			"context.tar.gz": []byte("test"),
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(existingCM).
		Build()

	store := NewConfigMapStore(fakeClient, scheme)

	lambda := &lambdav1alpha1.LambdaFunction{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "lambda.knative.io/v1alpha1",
			Kind:       "LambdaFunction",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-lambda",
			Namespace: "default",
			UID:       "test-uid-12345",
		},
	}

	// Set owner reference
	err := store.SetOwnerReference(context.Background(), lambda, "test-lambda-build-context", scheme)
	require.NoError(t, err)

	// Verify owner reference was set
	cm := &corev1.ConfigMap{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-lambda-build-context",
		Namespace: "default",
	}, cm)
	require.NoError(t, err)
	require.Len(t, cm.OwnerReferences, 1)
	assert.Equal(t, lambda.Name, cm.OwnerReferences[0].Name)
	assert.Equal(t, lambda.UID, cm.OwnerReferences[0].UID)
	assert.True(t, *cm.OwnerReferences[0].Controller)
	assert.True(t, *cm.OwnerReferences[0].BlockOwnerDeletion)
}

func TestBackend002_ConfigMapStore_SetOwnerReference_NotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = lambdav1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	store := NewConfigMapStore(fakeClient, scheme)

	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-lambda",
			Namespace: "default",
		},
	}

	// ConfigMap doesn't exist - should error
	err := store.SetOwnerReference(context.Background(), lambda, "non-existent-cm", scheme)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get ConfigMap")
}

func TestBackend002_ConfigMapStore_SetOwnerReference_Idempotent(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = lambdav1alpha1.AddToScheme(scheme)

	existingCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-lambda-build-context",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(existingCM).
		Build()

	store := NewConfigMapStore(fakeClient, scheme)

	lambda := &lambdav1alpha1.LambdaFunction{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "lambda.knative.io/v1alpha1",
			Kind:       "LambdaFunction",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-lambda",
			Namespace: "default",
			UID:       "test-uid-12345",
		},
	}

	// Set owner reference twice
	err := store.SetOwnerReference(context.Background(), lambda, "test-lambda-build-context", scheme)
	require.NoError(t, err)

	err = store.SetOwnerReference(context.Background(), lambda, "test-lambda-build-context", scheme)
	require.NoError(t, err)

	// Should still have only one owner reference
	cm := &corev1.ConfigMap{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-lambda-build-context",
		Namespace: "default",
	}, cm)
	require.NoError(t, err)
	assert.Len(t, cm.OwnerReferences, 1)
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
			name: "Create with valid config - MinIO endpoint",
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
			name: "Create with endpoint having https prefix",
			config: &S3Config{
				Bucket:          "test-bucket",
				Endpoint:        "https://s3.example.com",
				Region:          "us-east-1",
				AccessKeyID:     "access-key",
				SecretAccessKey: "secret-key",
				UseSSL:          true,
			},
		},
		{
			name: "Create with path prefix",
			config: &S3Config{
				Bucket:          "test-bucket",
				Endpoint:        "minio.test.svc:9000",
				Region:          "us-east-1",
				AccessKeyID:     "access-key",
				SecretAccessKey: "secret-key",
				PathPrefix:      "build-context/prod",
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
			assert.NotNil(t, store.GetClient())
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

func TestBackend002_GCSStore_Close(t *testing.T) {
	// Test that Close on nil client doesn't panic
	store := &GCSStore{client: nil, bucket: "test"}
	err := store.Close()
	assert.NoError(t, err)
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
	assert.Empty(t, location.Endpoint)
	assert.Empty(t, location.Region)
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
	assert.Equal(t, "s3.us-east-1.amazonaws.com", location.Endpoint)
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
	assert.Empty(t, location.Region)
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

func TestBackend002_BuildContextMeta_EmptyValues(t *testing.T) {
	meta := BuildContextMeta{}

	assert.Empty(t, meta.LambdaName)
	assert.Empty(t, meta.LambdaNamespace)
	assert.Empty(t, meta.ContentHash)
	assert.True(t, meta.CreatedAt.IsZero())
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
// Helper Function Tests
// =============================================================================

func TestBackend002_BoolPtr(t *testing.T) {
	truePtr := boolPtr(true)
	falsePtr := boolPtr(false)

	require.NotNil(t, truePtr)
	require.NotNil(t, falsePtr)
	assert.True(t, *truePtr)
	assert.False(t, *falsePtr)
}

// =============================================================================
// Missing Coverage Tests - S3Store
// =============================================================================

func TestBackend002_S3Store_Save(t *testing.T) {
	// Note: This tests the structure but can't actually save without a real S3/MinIO
	// In a real environment, use testcontainers or LocalStack
	config := &S3Config{
		Bucket:          "test-bucket",
		Endpoint:        "localhost:9000",
		Region:          "us-east-1",
		AccessKeyID:     "test",
		SecretAccessKey: "test",
		PathPrefix:      "build-context",
	}

	store, err := NewS3Store(config)
	require.NoError(t, err)
	require.NotNil(t, store)

	// Verify store configuration
	assert.Equal(t, "test-bucket", store.GetBucket())
	assert.Equal(t, "localhost:9000", store.endpoint)
	assert.Equal(t, "build-context", store.pathPrefix)
}

func TestBackend002_S3Store_PathPrefix(t *testing.T) {
	tests := []struct {
		name           string
		pathPrefix     string
		expectedPrefix string
	}{
		{"No prefix", "", ""},
		{"Simple prefix", "build-context", "build-context"},
		{"Prefix with trailing slash", "build-context/", "build-context/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &S3Config{
				Bucket:          "test-bucket",
				Endpoint:        "localhost:9000",
				Region:          "us-east-1",
				AccessKeyID:     "test",
				SecretAccessKey: "test",
				PathPrefix:      tt.pathPrefix,
			}

			store, err := NewS3Store(config)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPrefix, store.pathPrefix)
		})
	}
}

// =============================================================================
// Missing Coverage Tests - Error Cases
// =============================================================================

func TestBackend002_ConfigMapStore_Save_ContextDeadlineExceeded(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	store := NewConfigMapStore(fakeClient, scheme)

	// Note: The fake client doesn't respect context cancellation
	// In a real scenario with a real K8s client, this would fail
	// This test documents that the fake client doesn't propagate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	meta := BuildContextMeta{
		LambdaName:      "test-lambda",
		LambdaNamespace: "default",
		ContentHash:     "abc123def456789012",
		CreatedAt:       time.Now(),
	}

	// Fake client ignores context cancellation, so this succeeds
	// In production with real K8s client, this would error
	location, err := store.Save(ctx, "key", []byte("data"), meta)
	// Document actual behavior: fake client doesn't propagate context
	require.NoError(t, err)
	require.NotNil(t, location)
}

func TestBackend002_StorageSelector_NilLambda(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = lambdav1alpha1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	selector := NewStorageSelector(fakeClient, scheme, nil, nil)
	defer selector.Close()

	// Test with nil lambda - should panic or handle gracefully
	// This documents the expected behavior
	assert.Panics(t, func() {
		_, _ = selector.SelectStore(nil, 100)
	})
}

func TestBackend002_StorageSelector_NegativeSize(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = lambdav1alpha1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	selector := NewStorageSelector(fakeClient, scheme, nil, nil)
	defer selector.Close()

	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-lambda",
			Namespace: "default",
		},
	}

	// Negative size should default to ConfigMap (no error)
	store, err := selector.SelectStore(lambda, -1)
	require.NoError(t, err)
	assert.Equal(t, StorageBackendConfigMap, store.Name())
}

func TestBackend002_ConfigMapStore_Save_VerifyContent(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	store := NewConfigMapStore(fakeClient, scheme)

	testData := []byte("test build context data")
	meta := BuildContextMeta{
		LambdaName:      "verify-lambda",
		LambdaNamespace: "default",
		ContentHash:     "abc123def456789012345678901234567890",
		CreatedAt:       time.Now(),
	}

	location, err := store.Save(context.Background(), "key", testData, meta)
	require.NoError(t, err)

	// Verify the ConfigMap was created with correct content
	cm := &corev1.ConfigMap{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      location.ConfigMapName,
		Namespace: "default",
	}, cm)
	require.NoError(t, err)

	// Verify data
	assert.Equal(t, testData, cm.BinaryData["context.tar.gz"])

	// Verify labels
	assert.Equal(t, "knative-lambda-operator", cm.Labels["app.kubernetes.io/managed-by"])
	assert.Equal(t, "verify-lambda", cm.Labels["lambda.knative.io/name"])
	assert.Equal(t, "build-context", cm.Labels["lambda.knative.io/component"])

	// Verify annotations
	assert.Equal(t, meta.ContentHash, cm.Annotations["lambda.knative.io/content-hash"])
	assert.NotEmpty(t, cm.Annotations["lambda.knative.io/created-at"])
}

// =============================================================================
// Missing Coverage Tests - GCSStore Download
// =============================================================================

func TestBackend002_GCSStore_GetBucket(t *testing.T) {
	store := &GCSStore{bucket: "my-test-bucket"}
	assert.Equal(t, "my-test-bucket", store.GetBucket())
}

func TestBackend002_GCSStore_GetClient_Nil(t *testing.T) {
	store := &GCSStore{client: nil, bucket: "test"}
	assert.Nil(t, store.GetClient())
}

// =============================================================================
// Missing Coverage Tests - RecordMetrics Functions
// =============================================================================

func TestBackend002_RecordSourceFetch(t *testing.T) {
	// These don't return errors, just verify they don't panic
	assert.NotPanics(t, func() {
		RecordSourceFetch("github", "success")
	})
	assert.NotPanics(t, func() {
		RecordSourceFetch("gcs", "validation_error")
	})
	assert.NotPanics(t, func() {
		RecordSourceFetch("unknown", "unknown_error")
	})
}

func TestBackend002_RecordStorageSave(t *testing.T) {
	assert.NotPanics(t, func() {
		RecordStorageSave("configmap", "inline")
	})
	assert.NotPanics(t, func() {
		RecordStorageSave("s3", "github")
	})
	assert.NotPanics(t, func() {
		RecordStorageSave("gcs", "git")
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkBackend002_ConfigMapStore_Save_Small(b *testing.B) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	store := NewConfigMapStore(fakeClient, scheme)

	data := bytes.Repeat([]byte("x"), 10*1024) // 10KB
	meta := BuildContextMeta{
		LambdaName:      "bench-lambda",
		LambdaNamespace: "default",
		ContentHash:     "benchhash123456",
		CreatedAt:       time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Save(context.Background(), "key", data, meta)
	}
}

func BenchmarkBackend002_ConfigMapStore_Save_Large(b *testing.B) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	store := NewConfigMapStore(fakeClient, scheme)

	data := bytes.Repeat([]byte("x"), 500*1024) // 500KB
	meta := BuildContextMeta{
		LambdaName:      "bench-lambda",
		LambdaNamespace: "default",
		ContentHash:     "benchhash123456",
		CreatedAt:       time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Save(context.Background(), "key", data, meta)
	}
}

func BenchmarkBackend002_StorageSelector_SelectStore(b *testing.B) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = lambdav1alpha1.AddToScheme(scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	s3Config := &S3Config{
		Bucket:          "test-bucket",
		Endpoint:        "minio.test.svc:9000",
		Region:          "us-east-1",
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
	}

	selector := NewStorageSelector(fakeClient, scheme, s3Config, nil)
	defer selector.Close()

	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bench-lambda",
			Namespace: "default",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = selector.SelectStore(lambda, 100*1024)
	}
}
