// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	⚡ BENCHMARK TESTS - Performance benchmarks for storage operations
//
//	🎯 Purpose: Measure performance of storage operations
//	💡 Features: Upload/download benchmarks, various file sizes
//
//	⚡ Run with: go test -bench=. -benchmem ./internal/storage
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"knative-lambda-new/internal/config"
	testhelpers "knative-lambda-new/internal/testing"
)

// 🔧 generateTestData - "Generate test data of specified size"
func generateTestData(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}

// ⚡ BenchmarkMinIOStorage_Upload - "Benchmark MinIO upload operations"
func BenchmarkMinIOStorage_Upload(b *testing.B) {
	sizes := []int{
		1024,             // 1 KB
		10 * 1024,        // 10 KB
		100 * 1024,       // 100 KB
		1024 * 1024,      // 1 MB
		10 * 1024 * 1024, // 10 MB
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%dKB", size/1024), func(b *testing.B) {
			benchmarkMinIOUpload(b, size)
		})
	}
}

func benchmarkMinIOUpload(b *testing.B, size int) {
	ctx := context.Background()
	obs := testhelpers.CreateTestObservability(b)

	config := MinIOStorageConfig{
		Endpoint:      "localhost:9000",
		AccessKey:     "minioadmin",
		SecretKey:     "minioadmin",
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	storage, err := NewMinIOStorage(ctx, config)
	if err != nil {
		b.Skip("MinIO not available:", err)
	}

	testData := generateTestData(size)
	bucket := "benchmark-bucket"

	b.ResetTimer()
	b.SetBytes(int64(size))

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark/upload/file_%d.bin", i)
		reader := bytes.NewReader(testData)

		err := storage.UploadObject(ctx, bucket, key, reader, "application/octet-stream", int64(size))
		if err != nil {
			b.Fatalf("Upload failed: %v", err)
		}
	}
}

// ⚡ BenchmarkMinIOStorage_Download - "Benchmark MinIO download operations"
func BenchmarkMinIOStorage_Download(b *testing.B) {
	sizes := []int{
		1024,             // 1 KB
		10 * 1024,        // 10 KB
		100 * 1024,       // 100 KB
		1024 * 1024,      // 1 MB
		10 * 1024 * 1024, // 10 MB
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%dKB", size/1024), func(b *testing.B) {
			benchmarkMinIODownload(b, size)
		})
	}
}

func benchmarkMinIODownload(b *testing.B, size int) {
	ctx := context.Background()
	obs := testhelpers.CreateTestObservability(b)

	config := MinIOStorageConfig{
		Endpoint:      "localhost:9000",
		AccessKey:     "minioadmin",
		SecretKey:     "minioadmin",
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	storage, err := NewMinIOStorage(ctx, config)
	if err != nil {
		b.Skip("MinIO not available:", err)
	}

	// Upload test file once
	testData := generateTestData(size)
	bucket := "benchmark-bucket"
	key := fmt.Sprintf("benchmark/download/file_%d.bin", size)

	reader := bytes.NewReader(testData)
	err = storage.UploadObject(ctx, bucket, key, reader, "application/octet-stream", int64(size))
	if err != nil {
		b.Fatalf("Setup upload failed: %v", err)
	}

	b.ResetTimer()
	b.SetBytes(int64(size))

	for i := 0; i < b.N; i++ {
		reader, _, err := storage.GetObject(ctx, bucket, key)
		if err != nil {
			b.Fatalf("Download failed: %v", err)
		}

		// Read all data
		_, err = io.ReadAll(reader)
		reader.Close()
		if err != nil {
			b.Fatalf("Read failed: %v", err)
		}
	}
}

// ⚡ BenchmarkMinIOStorage_ObjectExists - "Benchmark MinIO exists check"
func BenchmarkMinIOStorage_ObjectExists(b *testing.B) {
	ctx := context.Background()
	obs := testhelpers.CreateTestObservability(b)

	config := MinIOStorageConfig{
		Endpoint:      "localhost:9000",
		AccessKey:     "minioadmin",
		SecretKey:     "minioadmin",
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	storage, err := NewMinIOStorage(ctx, config)
	if err != nil {
		b.Skip("MinIO not available:", err)
	}

	// Upload test file once
	testData := []byte("test content")
	bucket := "benchmark-bucket"
	key := "benchmark/exists/file.txt"

	reader := bytes.NewReader(testData)
	err = storage.UploadObject(ctx, bucket, key, reader, "text/plain", int64(len(testData)))
	if err != nil {
		b.Fatalf("Setup upload failed: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := storage.ObjectExists(ctx, bucket, key)
		if err != nil {
			b.Fatalf("Exists check failed: %v", err)
		}
	}
}

// ⚡ BenchmarkMinIOStorage_Delete - "Benchmark MinIO delete operations"
func BenchmarkMinIOStorage_Delete(b *testing.B) {
	ctx := context.Background()
	obs := testhelpers.CreateTestObservability(b)

	config := MinIOStorageConfig{
		Endpoint:      "localhost:9000",
		AccessKey:     "minioadmin",
		SecretKey:     "minioadmin",
		UseSSL:        false,
		Region:        "us-east-1",
		Observability: obs,
	}

	storage, err := NewMinIOStorage(ctx, config)
	if err != nil {
		b.Skip("MinIO not available:", err)
	}

	testData := []byte("test content")
	bucket := "benchmark-bucket"

	// Upload files for deletion benchmark
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark/delete/file_%d.txt", i)
		reader := bytes.NewReader(testData)
		err := storage.UploadObject(ctx, bucket, key, reader, "text/plain", int64(len(testData)))
		if err != nil {
			b.Fatalf("Setup upload failed: %v", err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark/delete/file_%d.txt", i)
		err := storage.DeleteObject(ctx, bucket, key)
		if err != nil {
			b.Fatalf("Delete failed: %v", err)
		}
	}
}

// ⚡ BenchmarkS3Storage_Upload - "Benchmark S3 upload operations"
func BenchmarkS3Storage_Upload(b *testing.B) {
	sizes := []int{
		1024,        // 1 KB
		10 * 1024,   // 10 KB
		100 * 1024,  // 100 KB
		1024 * 1024, // 1 MB
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%dKB", size/1024), func(b *testing.B) {
			benchmarkS3Upload(b, size)
		})
	}
}

func benchmarkS3Upload(b *testing.B, size int) {
	ctx := context.Background()
	obs := testhelpers.CreateTestObservability(b)

	config := S3StorageConfig{
		Region:        "us-west-2",
		Endpoint:      "",
		Observability: obs,
	}

	storage, err := NewS3Storage(ctx, config)
	if err != nil {
		b.Skip("S3 not available:", err)
	}

	testData := generateTestData(size)
	bucket := "benchmark-bucket"

	b.ResetTimer()
	b.SetBytes(int64(size))

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark/upload/file_%d.bin", i)
		reader := bytes.NewReader(testData)

		err := storage.UploadObject(ctx, bucket, key, reader, "application/octet-stream", int64(size))
		if err != nil {
			b.Fatalf("Upload failed: %v", err)
		}
	}
}

// ⚡ BenchmarkStorageFactory_CreateStorage - "Benchmark factory storage creation"
func BenchmarkStorageFactory_CreateStorage(b *testing.B) {
	obs := testhelpers.CreateTestObservability(b)

	storageConfig := &config.StorageConfig{
		Provider: "minio",
		MinIO: config.MinIOConfig{
			Endpoint:     "localhost:9000",
			AccessKey:    "minioadmin",
			SecretKey:    "minioadmin",
			UseSSL:       false,
			Region:       "us-east-1",
			SourceBucket: "test-source",
			TempBucket:   "test-temp",
		},
	}

	factory, err := NewStorageFactory(storageConfig, obs)
	if err != nil {
		b.Fatalf("Failed to create factory: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := factory.CreateStorage(ctx, ProviderMinIO)
		if err != nil {
			b.Fatalf("Create storage failed: %v", err)
		}
	}
}
