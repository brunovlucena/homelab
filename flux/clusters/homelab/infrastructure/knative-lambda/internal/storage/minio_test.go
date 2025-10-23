// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🏠 MINIO STORAGE TESTS - Test MinIO storage implementation
//
//	🎯 Purpose: Unit tests for MinIOStorage implementation
//	💡 Features: Upload, download, existence checks, deletion
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package storage

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMinIOStorage_NewMinIOStorage tests MinIO storage client creation
func TestMinIOStorage_NewMinIOStorage(t *testing.T) {
	tests := []struct {
		name        string
		config      MinIOStorageConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			config: MinIOStorageConfig{
				Endpoint:      "minio.minio.svc.cluster.local:9000",
				AccessKey:     "minioadmin",
				SecretKey:     "minioadmin",
				UseSSL:        false,
				Region:        "us-east-1",
				Observability: createTestObservability(t),
			},
			expectError: false,
		},
		{
			name: "valid configuration with SSL",
			config: MinIOStorageConfig{
				Endpoint:      "minio.example.com:9000",
				AccessKey:     "minioadmin",
				SecretKey:     "minioadmin",
				UseSSL:        true,
				Region:        "us-east-1",
				Observability: createTestObservability(t),
			},
			expectError: false,
		},
		{
			name: "missing endpoint",
			config: MinIOStorageConfig{
				Endpoint:      "",
				AccessKey:     "minioadmin",
				SecretKey:     "minioadmin",
				Observability: createTestObservability(t),
			},
			expectError: true,
			errorMsg:    "MinIO endpoint is required",
		},
		{
			name: "missing access key",
			config: MinIOStorageConfig{
				Endpoint:      "minio.minio.svc.cluster.local:9000",
				AccessKey:     "",
				SecretKey:     "minioadmin",
				Observability: createTestObservability(t),
			},
			expectError: true,
			errorMsg:    "MinIO access key is required",
		},
		{
			name: "missing secret key",
			config: MinIOStorageConfig{
				Endpoint:      "minio.minio.svc.cluster.local:9000",
				AccessKey:     "minioadmin",
				SecretKey:     "",
				Observability: createTestObservability(t),
			},
			expectError: true,
			errorMsg:    "MinIO secret key is required",
		},
		{
			name: "nil observability",
			config: MinIOStorageConfig{
				Endpoint:      "minio.minio.svc.cluster.local:9000",
				AccessKey:     "minioadmin",
				SecretKey:     "minioadmin",
				Observability: nil,
			},
			expectError: true,
			errorMsg:    "observability cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			client, err := NewMinIOStorage(ctx, tt.config)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, ProviderMinIO, client.GetProvider())
			}
		})
	}
}

// TestMinIOStorage_GetProvider tests provider type retrieval
func TestMinIOStorage_GetProvider(t *testing.T) {
	ctx := context.Background()
	obs := createTestObservability(t)

	config := MinIOStorageConfig{
		Endpoint:      "minio.minio.svc.cluster.local:9000",
		AccessKey:     "minioadmin",
		SecretKey:     "minioadmin",
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	client, err := NewMinIOStorage(ctx, config)
	require.NoError(t, err)

	provider := client.GetProvider()
	assert.Equal(t, ProviderMinIO, provider)
}

// TestMinIOStorage_GetEndpoint tests endpoint retrieval
func TestMinIOStorage_GetEndpoint(t *testing.T) {
	tests := []struct {
		name             string
		endpoint         string
		useSSL           bool
		expectedEndpoint string
	}{
		{
			name:             "HTTP endpoint",
			endpoint:         "minio.minio.svc.cluster.local:9000",
			useSSL:           false,
			expectedEndpoint: "http://minio.minio.svc.cluster.local:9000",
		},
		{
			name:             "HTTPS endpoint",
			endpoint:         "minio.example.com:9000",
			useSSL:           true,
			expectedEndpoint: "https://minio.example.com:9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			obs := createTestObservability(t)

			config := MinIOStorageConfig{
				Endpoint:      tt.endpoint,
				AccessKey:     "minioadmin",
				SecretKey:     "minioadmin",
				UseSSL:        tt.useSSL,
				Region:        "us-east-1",
				Observability: obs,
			}

			client, err := NewMinIOStorage(ctx, config)
			require.NoError(t, err)

			endpoint := client.GetEndpoint()
			assert.Equal(t, tt.expectedEndpoint, endpoint)
		})
	}
}

// TestMinIOStorage_GetBucketURL tests bucket URL generation
func TestMinIOStorage_GetBucketURL(t *testing.T) {
	ctx := context.Background()
	obs := createTestObservability(t)

	config := MinIOStorageConfig{
		Endpoint:      "minio.minio.svc.cluster.local:9000",
		AccessKey:     "minioadmin",
		SecretKey:     "minioadmin",
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	client, err := NewMinIOStorage(ctx, config)
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

// TestMinIOStorage_GetAccessKey tests access key retrieval
func TestMinIOStorage_GetAccessKey(t *testing.T) {
	ctx := context.Background()
	obs := createTestObservability(t)

	config := MinIOStorageConfig{
		Endpoint:      "minio.minio.svc.cluster.local:9000",
		AccessKey:     "test-access-key",
		SecretKey:     "test-secret-key",
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	client, err := NewMinIOStorage(ctx, config)
	require.NoError(t, err)

	accessKey := client.GetAccessKey()
	assert.Equal(t, "test-access-key", accessKey)
}

// TestMinIOStorage_GetSecretKey tests secret key retrieval
func TestMinIOStorage_GetSecretKey(t *testing.T) {
	ctx := context.Background()
	obs := createTestObservability(t)

	config := MinIOStorageConfig{
		Endpoint:      "minio.minio.svc.cluster.local:9000",
		AccessKey:     "test-access-key",
		SecretKey:     "test-secret-key",
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	client, err := NewMinIOStorage(ctx, config)
	require.NoError(t, err)

	secretKey := client.GetSecretKey()
	assert.Equal(t, "test-secret-key", secretKey)
}

// Mock tests for MinIO operations would require MinIO client mocking
// For integration tests, use actual MinIO instance

// TestMinIOStorage_UploadObject_Mock demonstrates the expected behavior
func TestMinIOStorage_UploadObject_Mock(t *testing.T) {
	t.Skip("Requires MinIO client mocking or actual MinIO instance")

	ctx := context.Background()
	obs := createTestObservability(t)

	config := MinIOStorageConfig{
		Endpoint:      "localhost:9000",
		AccessKey:     "minioadmin",
		SecretKey:     "minioadmin",
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	client, err := NewMinIOStorage(ctx, config)
	require.NoError(t, err)

	// Test data
	data := []byte("test content")
	reader := bytes.NewReader(data)

	// This would require a running MinIO instance or mocking
	err = client.UploadObject(ctx, "test-bucket", "test-key", reader, "application/octet-stream", int64(len(data)))

	// With proper setup, this should succeed
	assert.NoError(t, err)
}

// TestMinIOStorage_GetObject_Mock demonstrates the expected behavior
func TestMinIOStorage_GetObject_Mock(t *testing.T) {
	t.Skip("Requires MinIO client mocking or actual MinIO instance")

	ctx := context.Background()
	obs := createTestObservability(t)

	config := MinIOStorageConfig{
		Endpoint:      "localhost:9000",
		AccessKey:     "minioadmin",
		SecretKey:     "minioadmin",
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	client, err := NewMinIOStorage(ctx, config)
	require.NoError(t, err)

	// This would require a running MinIO instance or mocking
	reader, metadata, err := client.GetObject(ctx, "test-bucket", "test-key")

	// With proper setup, this should succeed
	assert.NoError(t, err)
	assert.NotNil(t, reader)
	assert.Greater(t, metadata.Size, int64(0))

	if reader != nil {
		defer reader.Close()
		content, _ := io.ReadAll(reader)
		assert.NotEmpty(t, content)
	}
}
