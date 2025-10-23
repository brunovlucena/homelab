// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🏭 STORAGE FACTORY TESTS - Test factory for creating storage clients
//
//	🎯 Purpose: Unit tests for StorageFactoryImpl
//	💡 Features: Test provider selection, validation, configuration
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"knative-lambda-new/internal/config"
	"knative-lambda-new/internal/observability"
	testhelpers "knative-lambda-new/internal/testing"
)

// TestNewStorageFactory tests factory creation
func TestNewStorageFactory(t *testing.T) {
	tests := []struct {
		name          string
		storageConfig *config.StorageConfig
		obs           *observability.Observability
		expectError   bool
		errorContains string
	}{
		{
			name: "valid configuration",
			storageConfig: &config.StorageConfig{
				Provider: "aws-s3",
				S3: config.S3Config{
					Region:       "us-west-2",
					Endpoint:     "",
					SourceBucket: "test-source",
					TempBucket:   "test-temp",
				},
			},
			obs:         testhelpers.CreateTestObservability(t),
			expectError: false,
		},
		{
			name:          "nil storage config",
			storageConfig: nil,
			obs:           testhelpers.CreateTestObservability(t),
			expectError:   true,
			errorContains: "storage config cannot be nil",
		},
		{
			name: "nil observability",
			storageConfig: &config.StorageConfig{
				Provider: "aws-s3",
			},
			obs:           nil,
			expectError:   true,
			errorContains: "observability cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, err := NewStorageFactory(tt.storageConfig, tt.obs)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, factory)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, factory)
			}
		})
	}
}

// TestStorageFactory_CreateStorage tests storage client creation
func TestStorageFactory_CreateStorage(t *testing.T) {
	tests := []struct {
		name          string
		provider      StorageProvider
		storageConfig *config.StorageConfig
		expectError   bool
		expectType    string
		errorContains string
	}{
		{
			name:     "create S3 storage",
			provider: ProviderS3,
			storageConfig: &config.StorageConfig{
				Provider: "aws-s3",
				S3: config.S3Config{
					Region:       "us-west-2",
					Endpoint:     "",
					SourceBucket: "test-source",
					TempBucket:   "test-temp",
				},
			},
			expectError: false,
			expectType:  "*storage.S3Storage",
		},
		{
			name:     "create MinIO storage",
			provider: ProviderMinIO,
			storageConfig: &config.StorageConfig{
				Provider: "minio",
				MinIO: config.MinIOConfig{
					Endpoint:     "minio.minio.svc.cluster.local:9000",
					AccessKey:    "minioadmin",
					SecretKey:    "minioadmin",
					UseSSL:       false,
					Region:       "us-east-1",
					SourceBucket: "test-source",
					TempBucket:   "test-temp",
				},
			},
			expectError: false,
			expectType:  "*storage.MinIOStorage",
		},
		{
			name:     "invalid provider",
			provider: StorageProvider("invalid"),
			storageConfig: &config.StorageConfig{
				Provider: "invalid",
			},
			expectError:   true,
			errorContains: "unsupported storage provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := testhelpers.CreateTestObservability(t)
			factory, err := NewStorageFactory(tt.storageConfig, obs)
			require.NoError(t, err)

			ctx := context.Background()
			storage, err := factory.CreateStorage(ctx, tt.provider)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, storage)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, storage)

				// Verify provider type
				switch tt.provider {
				case ProviderS3:
					_, ok := storage.(*S3Storage)
					assert.True(t, ok, "Expected *S3Storage, got %T", storage)
					assert.Equal(t, ProviderS3, storage.GetProvider())
				case ProviderMinIO:
					_, ok := storage.(*MinIOStorage)
					assert.True(t, ok, "Expected *MinIOStorage, got %T", storage)
					assert.Equal(t, ProviderMinIO, storage.GetProvider())
				}
			}
		})
	}
}

// TestStorageFactory_GetDefaultStorage tests default storage retrieval
func TestStorageFactory_GetDefaultStorage(t *testing.T) {
	tests := []struct {
		name          string
		storageConfig *config.StorageConfig
		expectError   bool
		expectedType  StorageProvider
	}{
		{
			name: "default S3 storage",
			storageConfig: &config.StorageConfig{
				Provider: "aws-s3",
				S3: config.S3Config{
					Region:       "us-west-2",
					SourceBucket: "test-source",
					TempBucket:   "test-temp",
				},
			},
			expectError:  false,
			expectedType: ProviderS3,
		},
		{
			name: "default MinIO storage",
			storageConfig: &config.StorageConfig{
				Provider: "minio",
				MinIO: config.MinIOConfig{
					Endpoint:     "minio.minio.svc.cluster.local:9000",
					AccessKey:    "minioadmin",
					SecretKey:    "minioadmin",
					SourceBucket: "test-source",
					TempBucket:   "test-temp",
				},
			},
			expectError:  false,
			expectedType: ProviderMinIO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := testhelpers.CreateTestObservability(t)
			factory, err := NewStorageFactory(tt.storageConfig, obs)
			require.NoError(t, err)

			ctx := context.Background()
			storage, err := factory.GetDefaultStorage(ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, storage)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, storage)
				assert.Equal(t, tt.expectedType, storage.GetProvider())
			}
		})
	}
}

// TestStorageFactory_S3Configuration tests S3 client configuration
func TestStorageFactory_S3Configuration(t *testing.T) {
	storageConfig := &config.StorageConfig{
		Provider: "aws-s3",
		S3: config.S3Config{
			Region:       "us-west-2",
			Endpoint:     "https://s3.custom.com",
			SourceBucket: "test-source",
			TempBucket:   "test-temp",
		},
	}

	obs := testhelpers.CreateTestObservability(t)
	factory, err := NewStorageFactory(storageConfig, obs)
	require.NoError(t, err)

	ctx := context.Background()
	storage, err := factory.CreateStorage(ctx, ProviderS3)
	require.NoError(t, err)

	s3Storage, ok := storage.(*S3Storage)
	require.True(t, ok)

	// Verify configuration
	assert.Equal(t, ProviderS3, s3Storage.GetProvider())
	endpoint := s3Storage.GetEndpoint()
	assert.NotEmpty(t, endpoint)
}

// TestStorageFactory_MinIOConfiguration tests MinIO client configuration
func TestStorageFactory_MinIOConfiguration(t *testing.T) {
	storageConfig := &config.StorageConfig{
		Provider: "minio",
		MinIO: config.MinIOConfig{
			Endpoint:     "minio.local:9000",
			AccessKey:    "test-access",
			SecretKey:    "test-secret",
			UseSSL:       false,
			Region:       "us-east-1",
			SourceBucket: "test-source",
			TempBucket:   "test-temp",
		},
	}

	obs := testhelpers.CreateTestObservability(t)
	factory, err := NewStorageFactory(storageConfig, obs)
	require.NoError(t, err)

	ctx := context.Background()
	storage, err := factory.CreateStorage(ctx, ProviderMinIO)
	require.NoError(t, err)

	minioStorage, ok := storage.(*MinIOStorage)
	require.True(t, ok)

	// Verify configuration
	assert.Equal(t, ProviderMinIO, minioStorage.GetProvider())
	endpoint := minioStorage.GetEndpoint()
	assert.Contains(t, endpoint, "minio.local:9000")
}

// TestStorageFactory_ProviderSwitching tests switching between providers
func TestStorageFactory_ProviderSwitching(t *testing.T) {
	storageConfig := &config.StorageConfig{
		Provider: "aws-s3",
		S3: config.S3Config{
			Region:       "us-west-2",
			SourceBucket: "test-source",
			TempBucket:   "test-temp",
		},
		MinIO: config.MinIOConfig{
			Endpoint:     "minio.local:9000",
			AccessKey:    "minioadmin",
			SecretKey:    "minioadmin",
			SourceBucket: "test-source",
			TempBucket:   "test-temp",
		},
	}

	obs := testhelpers.CreateTestObservability(t)
	factory, err := NewStorageFactory(storageConfig, obs)
	require.NoError(t, err)

	ctx := context.Background()

	// Create S3 storage
	s3Storage, err := factory.CreateStorage(ctx, ProviderS3)
	require.NoError(t, err)
	assert.Equal(t, ProviderS3, s3Storage.GetProvider())

	// Create MinIO storage with same factory
	minioStorage, err := factory.CreateStorage(ctx, ProviderMinIO)
	require.NoError(t, err)
	assert.Equal(t, ProviderMinIO, minioStorage.GetProvider())

	// Verify they are different instances
	assert.NotEqual(t, s3Storage, minioStorage)
}
