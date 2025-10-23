// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🏭 STORAGE FACTORY TESTS - Test storage factory implementation
//
//	🎯 Purpose: Unit tests for StorageFactory
//	💡 Features: Provider selection, client creation, validation
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
)

// TestStorageFactory_NewStorageFactory tests factory creation
func TestStorageFactory_NewStorageFactory(t *testing.T) {
	tests := []struct {
		name          string
		storageConfig *config.StorageConfig
		obs           *observability.Observability
		expectError   bool
		errorMsg      string
	}{
		{
			name: "valid configuration",
			storageConfig: &config.StorageConfig{
				Provider: "aws-s3",
				S3: config.S3Config{
					Region:       "us-west-2",
					SourceBucket: "test-source",
					TempBucket:   "test-tmp",
				},
			},
			obs:         createTestObservability(t),
			expectError: false,
		},
		{
			name:          "nil storage config",
			storageConfig: nil,
			obs:           createTestObservability(t),
			expectError:   true,
			errorMsg:      "storage config cannot be nil",
		},
		{
			name: "nil observability",
			storageConfig: &config.StorageConfig{
				Provider: "aws-s3",
			},
			obs:         nil,
			expectError: true,
			errorMsg:    "observability cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, err := NewStorageFactory(tt.storageConfig, tt.obs)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
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
		name             string
		provider         StorageProvider
		storageConfig    *config.StorageConfig
		expectError      bool
		errorMsg         string
		expectedProvider StorageProvider
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
					TempBucket:   "test-tmp",
				},
			},
			expectError:      false,
			expectedProvider: ProviderS3,
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
					TempBucket:   "test-tmp",
				},
			},
			expectError:      false,
			expectedProvider: ProviderMinIO,
		},
		{
			name:     "unsupported provider",
			provider: "invalid-provider",
			storageConfig: &config.StorageConfig{
				Provider: "invalid",
			},
			expectError: true,
			errorMsg:    "unsupported storage provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			obs := createTestObservability(t)

			factory, err := NewStorageFactory(tt.storageConfig, obs)
			require.NoError(t, err)

			client, err := factory.CreateStorage(ctx, tt.provider)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.expectedProvider, client.GetProvider())
			}
		})
	}
}

// TestStorageFactory_GetDefaultStorage tests default storage retrieval
func TestStorageFactory_GetDefaultStorage(t *testing.T) {
	tests := []struct {
		name             string
		storageConfig    *config.StorageConfig
		expectedProvider StorageProvider
		expectError      bool
	}{
		{
			name: "default S3 storage",
			storageConfig: &config.StorageConfig{
				Provider: "aws-s3",
				S3: config.S3Config{
					Region:       "us-west-2",
					SourceBucket: "test-source",
					TempBucket:   "test-tmp",
				},
			},
			expectedProvider: ProviderS3,
			expectError:      false,
		},
		{
			name: "default MinIO storage",
			storageConfig: &config.StorageConfig{
				Provider: "minio",
				MinIO: config.MinIOConfig{
					Endpoint:     "minio.minio.svc.cluster.local:9000",
					AccessKey:    "minioadmin",
					SecretKey:    "minioadmin",
					UseSSL:       false,
					Region:       "us-east-1",
					SourceBucket: "test-source",
					TempBucket:   "test-tmp",
				},
			},
			expectedProvider: ProviderMinIO,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Properly initialize observability
			obs, err := observability.New(observability.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
			})
			require.NoError(t, err)
			defer obs.Shutdown(ctx)

			factory, err := NewStorageFactory(tt.storageConfig, obs)
			require.NoError(t, err)

			client, err := factory.GetDefaultStorage(ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.expectedProvider, client.GetProvider())
			}
		})
	}
}

// TestStorageFactory_ProviderSwitching tests switching between providers
func TestStorageFactory_ProviderSwitching(t *testing.T) {
	ctx := context.Background()

	// Properly initialize observability
	obs, err := observability.New(observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	})
	require.NoError(t, err)
	defer obs.Shutdown(ctx)

	// Create config that supports both providers
	storageConfig := &config.StorageConfig{
		Provider: "aws-s3",
		S3: config.S3Config{
			Region:       "us-west-2",
			SourceBucket: "s3-source",
			TempBucket:   "s3-tmp",
		},
		MinIO: config.MinIOConfig{
			Endpoint:     "minio.minio.svc.cluster.local:9000",
			AccessKey:    "minioadmin",
			SecretKey:    "minioadmin",
			UseSSL:       false,
			Region:       "us-east-1",
			SourceBucket: "minio-source",
			TempBucket:   "minio-tmp",
		},
	}

	factory, err := NewStorageFactory(storageConfig, obs)
	require.NoError(t, err)

	// Create S3 client
	s3Client, err := factory.CreateStorage(ctx, ProviderS3)
	require.NoError(t, err)
	assert.NotNil(t, s3Client)
	assert.Equal(t, ProviderS3, s3Client.GetProvider())

	// Create MinIO client
	minioClient, err := factory.CreateStorage(ctx, ProviderMinIO)
	require.NoError(t, err)
	assert.NotNil(t, minioClient)
	assert.Equal(t, ProviderMinIO, minioClient.GetProvider())

	// Verify they are different instances
	assert.NotEqual(t, s3Client, minioClient)
}
