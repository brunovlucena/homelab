// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 INTEGRATION TESTS - Integration tests with MinIO testcontainers
//
//	🎯 Purpose: Test storage implementations against real MinIO instance
//	💡 Features: Full end-to-end testing with containerized MinIO
//
//	⚠️ Run with: go test -tags=integration ./...
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//go:build integration
// +build integration

package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	testhelpers "knative-lambda-new/internal/testing"
)

const (
	minioImage      = "minio/minio:latest"
	minioAccessKey  = "minioadmin"
	minioSecretKey  = "minioadmin"
	testBucket      = "test-bucket"
	testTimeout     = 2 * time.Minute
)

// 🏗️ setupMinIOContainer - "Start MinIO container for testing"
func setupMinIOContainer(t *testing.T) (string, func()) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        minioImage,
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"MINIO_ROOT_USER":     minioAccessKey,
			"MINIO_ROOT_PASSWORD": minioSecretKey,
		},
		Cmd:        []string{"server", "/data"},
		WaitingFor: wait.ForHTTP("/minio/health/live").WithPort("9000/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "Failed to start MinIO container")

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "9000")
	require.NoError(t, err)

	endpoint := fmt.Sprintf("%s:%s", host, port.Port())

	// Wait for MinIO to be ready
	time.Sleep(2 * time.Second)

	cleanup := func() {
		ctx := context.Background()
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}

	return endpoint, cleanup
}

// 🧪 TestMinIOStorage_Integration_FullCycle - "Test full upload/download/delete cycle"
func TestMinIOStorage_Integration_FullCycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	endpoint, cleanup := setupMinIOContainer(t)
	defer cleanup()

	obs := testhelpers.CreateTestObservability(t)
	config := MinIOStorageConfig{
		Endpoint:      endpoint,
		AccessKey:     minioAccessKey,
		SecretKey:     minioSecretKey,
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	ctx := context.Background()
	storage, err := NewMinIOStorage(ctx, config)
	require.NoError(t, err)

	// Test data
	testKey := "test/file.txt"
	testContent := []byte("Hello, MinIO Integration Test!")
	contentType := "text/plain"

	t.Run("Upload Object", func(t *testing.T) {
		reader := bytes.NewReader(testContent)
		err := storage.UploadObject(ctx, testBucket, testKey, reader, contentType, int64(len(testContent)))
		require.NoError(t, err)
	})

	t.Run("Object Exists", func(t *testing.T) {
		exists, err := storage.ObjectExists(ctx, testBucket, testKey)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Download Object", func(t *testing.T) {
		reader, metadata, err := storage.GetObject(ctx, testBucket, testKey)
		require.NoError(t, err)
		defer reader.Close()

		// Verify metadata
		assert.Equal(t, int64(len(testContent)), metadata.Size)
		assert.Equal(t, contentType, metadata.ContentType)

		// Verify content
		downloadedContent, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, testContent, downloadedContent)
	})

	t.Run("Delete Object", func(t *testing.T) {
		err := storage.DeleteObject(ctx, testBucket, testKey)
		require.NoError(t, err)

		// Verify object is deleted
		exists, err := storage.ObjectExists(ctx, testBucket, testKey)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

// 🧪 TestMinIOStorage_Integration_LargeFile - "Test large file upload/download"
func TestMinIOStorage_Integration_LargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	endpoint, cleanup := setupMinIOContainer(t)
	defer cleanup()

	obs := testhelpers.CreateTestObservability(t)
	config := MinIOStorageConfig{
		Endpoint:      endpoint,
		AccessKey:     minioAccessKey,
		SecretKey:     minioSecretKey,
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	ctx := context.Background()
	storage, err := NewMinIOStorage(ctx, config)
	require.NoError(t, err)

	// Create 5MB test file
	testKey := "test/largefile.bin"
	fileSize := 5 * 1024 * 1024 // 5MB
	testContent := make([]byte, fileSize)
	for i := range testContent {
		testContent[i] = byte(i % 256)
	}

	t.Run("Upload Large File", func(t *testing.T) {
		reader := bytes.NewReader(testContent)
		err := storage.UploadObject(ctx, testBucket, testKey, reader, "application/octet-stream", int64(fileSize))
		require.NoError(t, err)
	})

	t.Run("Download Large File", func(t *testing.T) {
		reader, metadata, err := storage.GetObject(ctx, testBucket, testKey)
		require.NoError(t, err)
		defer reader.Close()

		assert.Equal(t, int64(fileSize), metadata.Size)

		downloadedContent, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, len(testContent), len(downloadedContent))
		assert.Equal(t, testContent, downloadedContent)
	})

	t.Run("Cleanup Large File", func(t *testing.T) {
		err := storage.DeleteObject(ctx, testBucket, testKey)
		require.NoError(t, err)
	})
}

// 🧪 TestMinIOStorage_Integration_MultipleFiles - "Test multiple file operations"
func TestMinIOStorage_Integration_MultipleFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	endpoint, cleanup := setupMinIOContainer(t)
	defer cleanup()

	obs := testhelpers.CreateTestObservability(t)
	config := MinIOStorageConfig{
		Endpoint:      endpoint,
		AccessKey:     minioAccessKey,
		SecretKey:     minioSecretKey,
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	ctx := context.Background()
	storage, err := NewMinIOStorage(ctx, config)
	require.NoError(t, err)

	// Upload multiple files
	numFiles := 10
	files := make(map[string][]byte)

	for i := 0; i < numFiles; i++ {
		key := fmt.Sprintf("test/file%d.txt", i)
		content := []byte(fmt.Sprintf("Content of file %d", i))
		files[key] = content

		reader := bytes.NewReader(content)
		err := storage.UploadObject(ctx, testBucket, key, reader, "text/plain", int64(len(content)))
		require.NoError(t, err)
	}

	// Verify all files exist and have correct content
	for key, expectedContent := range files {
		exists, err := storage.ObjectExists(ctx, testBucket, key)
		require.NoError(t, err)
		assert.True(t, exists, "File %s should exist", key)

		reader, _, err := storage.GetObject(ctx, testBucket, key)
		require.NoError(t, err)

		actualContent, err := io.ReadAll(reader)
		reader.Close()
		require.NoError(t, err)
		assert.Equal(t, expectedContent, actualContent)
	}

	// Delete all files
	for key := range files {
		err := storage.DeleteObject(ctx, testBucket, key)
		require.NoError(t, err)
	}
}

// 🧪 TestMinIOStorage_Integration_ErrorCases - "Test error handling"
func TestMinIOStorage_Integration_ErrorCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	endpoint, cleanup := setupMinIOContainer(t)
	defer cleanup()

	obs := testhelpers.CreateTestObservability(t)
	config := MinIOStorageConfig{
		Endpoint:      endpoint,
		AccessKey:     minioAccessKey,
		SecretKey:     minioSecretKey,
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	ctx := context.Background()
	storage, err := NewMinIOStorage(ctx, config)
	require.NoError(t, err)

	t.Run("Download Non-Existent Object", func(t *testing.T) {
		_, _, err := storage.GetObject(ctx, testBucket, "nonexistent/file.txt")
		assert.Error(t, err)
	})

	t.Run("Delete Non-Existent Object", func(t *testing.T) {
		// MinIO/S3 returns success when deleting non-existent objects
		err := storage.DeleteObject(ctx, testBucket, "nonexistent/file.txt")
		assert.NoError(t, err)
	})

	t.Run("Check Non-Existent Object", func(t *testing.T) {
		exists, err := storage.ObjectExists(ctx, testBucket, "nonexistent/file.txt")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

