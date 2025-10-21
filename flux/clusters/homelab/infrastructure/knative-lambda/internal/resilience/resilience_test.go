package resilience

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMultiLevelRateLimiter(t *testing.T) {
	config := MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 10,
		BuildContextBurstSize:      5,
		K8sJobRequestsPerMin:       20,
		K8sJobBurstSize:            10,
		ClientRequestsPerMin:       30,
		ClientBurstSize:            15,
		S3UploadRequestsPerMin:     5,
		S3UploadBurstSize:          2,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
		CleanupInterval:            5 * time.Minute,
		ClientTTL:                  10 * time.Minute,
	}

	limiter, err := NewMultiLevelRateLimiter(config)

	assert.NoError(t, err)
	assert.NotNil(t, limiter)
}

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(10, 5)

	assert.NotNil(t, limiter)
	assert.Equal(t, 10, limiter.requestsPerMin)
	assert.Equal(t, 5, limiter.burstSize)
	assert.NotNil(t, limiter.tokens)
}

func TestNewMemoryMonitor(t *testing.T) {
	monitor := NewMemoryMonitor(80.0, 30*time.Second)

	assert.NotNil(t, monitor)
	assert.Equal(t, 80.0, monitor.maxUsagePercent)
	assert.Equal(t, 30*time.Second, monitor.checkInterval)
}

func TestMultiLevelRateLimiter_Allow(t *testing.T) {
	config := MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 10,
		BuildContextBurstSize:      5,
		K8sJobRequestsPerMin:       20,
		K8sJobBurstSize:            10,
		ClientRequestsPerMin:       30,
		ClientBurstSize:            15,
		S3UploadRequestsPerMin:     5,
		S3UploadBurstSize:          2,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
		CleanupInterval:            5 * time.Minute,
		ClientTTL:                  10 * time.Minute,
	}

	limiter, err := NewMultiLevelRateLimiter(config)
	require.NoError(t, err)

	// Test different operation types
	assert.True(t, limiter.Allow("build_context"))
	assert.True(t, limiter.Allow("k8s_job"))
	assert.True(t, limiter.Allow("client"))
	assert.True(t, limiter.Allow("s3_upload"))

	// Test unknown operation type
	assert.False(t, limiter.Allow("unknown_operation"))
}

func TestMultiLevelRateLimiter_AllowBuildContext(t *testing.T) {
	config := MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 1, // Very low limit for testing
		BuildContextBurstSize:      1,
		K8sJobRequestsPerMin:       20,
		K8sJobBurstSize:            10,
		ClientRequestsPerMin:       30,
		ClientBurstSize:            15,
		S3UploadRequestsPerMin:     5,
		S3UploadBurstSize:          2,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
		CleanupInterval:            5 * time.Minute,
		ClientTTL:                  10 * time.Minute,
	}

	limiter, err := NewMultiLevelRateLimiter(config)
	require.NoError(t, err)

	// First request should be allowed
	assert.True(t, limiter.Allow("build_context"))

	// Second request should be rate limited
	assert.False(t, limiter.Allow("build_context"))
}

func TestMultiLevelRateLimiter_AllowK8sJob(t *testing.T) {
	config := MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 10,
		BuildContextBurstSize:      5,
		K8sJobRequestsPerMin:       1, // Very low limit for testing
		K8sJobBurstSize:            1,
		ClientRequestsPerMin:       30,
		ClientBurstSize:            15,
		S3UploadRequestsPerMin:     5,
		S3UploadBurstSize:          2,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
		CleanupInterval:            5 * time.Minute,
		ClientTTL:                  10 * time.Minute,
	}

	limiter, err := NewMultiLevelRateLimiter(config)
	require.NoError(t, err)

	// First request should be allowed
	assert.True(t, limiter.Allow("k8s_job"))

	// Second request should be rate limited
	assert.False(t, limiter.Allow("k8s_job"))
}

func TestMultiLevelRateLimiter_AllowClient(t *testing.T) {
	config := MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 10,
		BuildContextBurstSize:      5,
		K8sJobRequestsPerMin:       20,
		K8sJobBurstSize:            10,
		ClientRequestsPerMin:       1, // Very low limit for testing
		ClientBurstSize:            1,
		S3UploadRequestsPerMin:     5,
		S3UploadBurstSize:          2,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
		CleanupInterval:            5 * time.Minute,
		ClientTTL:                  10 * time.Minute,
	}

	limiter, err := NewMultiLevelRateLimiter(config)
	require.NoError(t, err)

	// First request should be allowed
	assert.True(t, limiter.Allow("client"))

	// Second request should be rate limited
	assert.False(t, limiter.Allow("client"))
}

func TestMultiLevelRateLimiter_AllowS3Upload(t *testing.T) {
	config := MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 10,
		BuildContextBurstSize:      5,
		K8sJobRequestsPerMin:       20,
		K8sJobBurstSize:            10,
		ClientRequestsPerMin:       30,
		ClientBurstSize:            15,
		S3UploadRequestsPerMin:     1, // Very low limit for testing
		S3UploadBurstSize:          1,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
		CleanupInterval:            5 * time.Minute,
		ClientTTL:                  10 * time.Minute,
	}

	limiter, err := NewMultiLevelRateLimiter(config)
	require.NoError(t, err)

	// First request should be allowed
	assert.True(t, limiter.Allow("s3_upload"))

	// Second request should be rate limited
	assert.False(t, limiter.Allow("s3_upload"))
}

func TestMultiLevelRateLimiter_IsMemoryUsageHigh(t *testing.T) {
	config := MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 10,
		BuildContextBurstSize:      5,
		K8sJobRequestsPerMin:       20,
		K8sJobBurstSize:            10,
		ClientRequestsPerMin:       30,
		ClientBurstSize:            15,
		S3UploadRequestsPerMin:     5,
		S3UploadBurstSize:          2,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
		CleanupInterval:            5 * time.Minute,
		ClientTTL:                  10 * time.Minute,
	}

	limiter, err := NewMultiLevelRateLimiter(config)
	require.NoError(t, err)

	// This is a basic test - actual memory usage depends on the system
	// We're just testing that the function doesn't panic
	_ = limiter.IsMemoryUsageHigh()
}

func TestMultiLevelRateLimiter_StartCleanup(t *testing.T) {
	config := MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 10,
		BuildContextBurstSize:      5,
		K8sJobRequestsPerMin:       20,
		K8sJobBurstSize:            10,
		ClientRequestsPerMin:       30,
		ClientBurstSize:            15,
		S3UploadRequestsPerMin:     5,
		S3UploadBurstSize:          2,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
		CleanupInterval:            5 * time.Minute,
		ClientTTL:                  10 * time.Minute,
	}

	limiter, err := NewMultiLevelRateLimiter(config)
	require.NoError(t, err)

	// Test that cleanup can be started
	limiter.StartCleanup()

	// Test that cleanup can be stopped
	limiter.StopCleanup()
}
