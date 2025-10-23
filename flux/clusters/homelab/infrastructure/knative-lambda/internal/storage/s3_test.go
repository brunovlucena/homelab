// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	☁️ S3 STORAGE TESTS - Test AWS S3 storage implementation
//
//	🎯 Purpose: Unit tests for S3Storage implementation
//	💡 Features: Upload, download, existence checks, deletion
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package storage

import (
	"bytes"
	"context"
	"io"
	testhelpers "knative-lambda-new/internal/testing"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestS3Storage_NewS3Storage tests S3 storage client creation
func TestS3Storage_NewS3Storage(t *testing.T) {
	tests := []struct {
		name        string
		config      S3StorageConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			config: S3StorageConfig{
				Region:        "us-west-2",
				Endpoint:      "",
				Observability: testhelpers.CreateTestObservability(t),
			},
			expectError: false,
		},
		{
			name: "valid configuration with custom endpoint",
			config: S3StorageConfig{
				Region:        "us-west-2",
				Endpoint:      "https://s3.custom.com",
				Observability: testhelpers.CreateTestObservability(t),
			},
			expectError: false,
		},
		{
			name: "nil observability",
			config: S3StorageConfig{
				Region:        "us-west-2",
				Observability: nil,
			},
			expectError: true,
			errorMsg:    "observability cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			client, err := NewS3Storage(ctx, tt.config)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, ProviderS3, client.GetProvider())
			}
		})
	}
}

// TestS3Storage_GetProvider tests provider type retrieval
func TestS3Storage_GetProvider(t *testing.T) {
	ctx := context.Background()
	obs := testhelpers.CreateTestObservability(t)

	config := S3StorageConfig{
		Region:        "us-west-2",
		Observability: obs,
	}

	client, err := NewS3Storage(ctx, config)
	require.NoError(t, err)

	provider := client.GetProvider()
	assert.Equal(t, ProviderS3, provider)
}

// TestS3Storage_GetEndpoint tests endpoint retrieval
func TestS3Storage_GetEndpoint(t *testing.T) {
	tests := []struct {
		name             string
		endpoint         string
		region           string
		expectedEndpoint string
	}{
		{
			name:             "default endpoint",
			endpoint:         "",
			region:           "us-west-2",
			expectedEndpoint: "https://s3.us-west-2.amazonaws.com",
		},
		{
			name:             "custom endpoint",
			endpoint:         "https://s3.custom.com",
			region:           "us-west-2",
			expectedEndpoint: "https://s3.custom.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			obs := testhelpers.CreateTestObservability(t)

			config := S3StorageConfig{
				Region:        tt.region,
				Endpoint:      tt.endpoint,
				Observability: obs,
			}

			client, err := NewS3Storage(ctx, config)
			require.NoError(t, err)

			endpoint := client.GetEndpoint()
			assert.Equal(t, tt.expectedEndpoint, endpoint)
		})
	}
}

// TestS3Storage_GetBucketURL tests bucket URL generation
func TestS3Storage_GetBucketURL(t *testing.T) {
	ctx := context.Background()
	obs := testhelpers.CreateTestObservability(t)

	config := S3StorageConfig{
		Region:        "us-west-2",
		Observability: obs,
	}

	client, err := NewS3Storage(ctx, config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		bucket      string
		key         string
		expectedURL string
	}{
		{
			name:        "simple path",
			bucket:      "test-bucket",
			key:         "test-key",
			expectedURL: "s3://test-bucket/test-key",
		},
		{
			name:        "nested path",
			bucket:      "test-bucket",
			key:         "path/to/object.tar.gz",
			expectedURL: "s3://test-bucket/path/to/object.tar.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := client.GetBucketURL(tt.bucket, tt.key)
			assert.Equal(t, tt.expectedURL, url)
		})
	}
}

// Mock tests for S3 operations would require AWS SDK mocking
// For integration tests, use localstack or actual S3 buckets

// TestS3Storage_UploadObject_Mock demonstrates the expected behavior
func TestS3Storage_UploadObject_Mock(t *testing.T) {
	t.Skip("Requires AWS SDK mocking or localstack")

	ctx := context.Background()
	obs := testhelpers.CreateTestObservability(t)

	config := S3StorageConfig{
		Region:        "us-west-2",
		Observability: obs,
	}

	client, err := NewS3Storage(ctx, config)
	require.NoError(t, err)

	// Test data
	data := []byte("test content")
	reader := bytes.NewReader(data)

	// This would require mocking the S3 client
	err = client.UploadObject(ctx, "test-bucket", "test-key", reader, "application/octet-stream", int64(len(data)))

	// With proper mocking, this should succeed
	assert.NoError(t, err)
}

// TestS3Storage_GetObject_Mock demonstrates the expected behavior
func TestS3Storage_GetObject_Mock(t *testing.T) {
	t.Skip("Requires AWS SDK mocking or localstack")

	ctx := context.Background()
	obs := testhelpers.CreateTestObservability(t)

	config := S3StorageConfig{
		Region:        "us-west-2",
		Observability: obs,
	}

	client, err := NewS3Storage(ctx, config)
	require.NoError(t, err)

	// This would require mocking the S3 client
	reader, metadata, err := client.GetObject(ctx, "test-bucket", "test-key")

	// With proper mocking, this should succeed
	assert.NoError(t, err)
	assert.NotNil(t, reader)
	assert.Greater(t, metadata.Size, int64(0))

	if reader != nil {
		defer reader.Close()
		content, _ := io.ReadAll(reader)
		assert.NotEmpty(t, content)
	}
}

// TestS3Storage_ObjectExists_Mock demonstrates the expected behavior
func TestS3Storage_ObjectExists_Mock(t *testing.T) {
	t.Skip("Requires AWS SDK mocking or localstack")

	ctx := context.Background()
	obs := testhelpers.CreateTestObservability(t)

	config := S3StorageConfig{
		Region:        "us-west-2",
		Observability: obs,
	}

	client, err := NewS3Storage(ctx, config)
	require.NoError(t, err)

	// This would require mocking the S3 client
	exists, err := client.ObjectExists(ctx, "test-bucket", "test-key")

	// With proper mocking, this should succeed
	assert.NoError(t, err)
	assert.True(t, exists)
}

// TestS3Storage_DeleteObject_Mock demonstrates the expected behavior
func TestS3Storage_DeleteObject_Mock(t *testing.T) {
	t.Skip("Requires AWS SDK mocking or localstack")

	ctx := context.Background()
	obs := testhelpers.CreateTestObservability(t)

	config := S3StorageConfig{
		Region:        "us-west-2",
		Observability: obs,
	}

	client, err := NewS3Storage(ctx, config)
	require.NoError(t, err)

	// This would require mocking the S3 client
	err = client.DeleteObject(ctx, "test-bucket", "test-key")

	// With proper mocking, this should succeed
	assert.NoError(t, err)
}
