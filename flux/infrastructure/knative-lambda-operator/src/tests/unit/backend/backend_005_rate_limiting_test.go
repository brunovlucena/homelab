// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 BACKEND-005: Rate Limiting & Resilience Tests
//
//	User Story: Rate Limiting & Resilience
//	Priority: P1 | Story Points: 8
//
//	Tests validate:
//	- Rate limiter creation and configuration
//	- Multi-level rate limiting functionality
//	- Token bucket algorithm behavior
//	- Memory monitoring
//	- Concurrent request handling
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package backend

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"knative-lambda/internal/resilience"
)

// TestBackend005_RateLimiterCreation validates rate limiter creation.
func TestBackend005_RateLimiterCreation(t *testing.T) {
	tests := []struct {
		name           string
		requestsPerMin int
		burstSize      int
	}{
		{
			name:           "Standard rate limit",
			requestsPerMin: 60,
			burstSize:      10,
		},
		{
			name:           "High throughput",
			requestsPerMin: 300,
			burstSize:      50,
		},
		{
			name:           "Low rate limit",
			requestsPerMin: 10,
			burstSize:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			limiter := resilience.NewRateLimiter(tt.requestsPerMin, tt.burstSize)

			// Assert
			require.NotNil(t, limiter)
			// Test behavior: should allow requests up to burst size
			for i := 0; i < tt.burstSize; i++ {
				assert.True(t, limiter.Allow(), "Should allow request within burst size")
			}
		})
	}
}

// TestBackend005_BasicRateLimiting validates basic rate limiting functionality.
func TestBackend005_BasicRateLimiting(t *testing.T) {
	// Arrange
	requestsPerMin := 60
	burstSize := 5
	limiter := resilience.NewRateLimiter(requestsPerMin, burstSize)

	// Act - Try to consume burst tokens
	successCount := 0
	for i := 0; i < burstSize+5; i++ {
		if limiter.Allow() {
			successCount++
		}
	}

	// Assert
	// Should allow at most burstSize requests initially
	assert.LessOrEqual(t, successCount, burstSize, "Should not exceed burst limit")
}

// TestBackend005_MultiLevelRateLimiterCreation validates multi-level rate limiter creation.
func TestBackend005_MultiLevelRateLimiterCreation(t *testing.T) {
	// Arrange
	config := resilience.MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 5,
		BuildContextBurstSize:      2,
		K8sJobRequestsPerMin:       10,
		K8sJobBurstSize:            3,
		ClientRequestsPerMin:       5,
		ClientBurstSize:            2,
		S3UploadRequestsPerMin:     50,
		S3UploadBurstSize:          10,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
		CleanupInterval:            5 * time.Minute,
		ClientTTL:                  1 * time.Hour,
	}

	// Act
	limiter, err := resilience.NewMultiLevelRateLimiter(config, nil, nil) // Using in-memory (no Redis)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, limiter)
	// Test behavior: should allow requests for different operation types
	assert.True(t, limiter.Allow("build_context"), "Should allow build_context operations")
	assert.True(t, limiter.Allow("k8s_job"), "Should allow k8s_job operations")
	assert.True(t, limiter.Allow("client"), "Should allow client operations")
	assert.True(t, limiter.Allow("s3_upload"), "Should allow s3_upload operations")
}

// TestBackend005_MultiLevelAllowOperation validates operation-specific rate limiting.
func TestBackend005_MultiLevelAllowOperation(t *testing.T) {
	// Arrange
	config := resilience.MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 5,
		BuildContextBurstSize:      2,
		K8sJobRequestsPerMin:       10,
		K8sJobBurstSize:            3,
		ClientRequestsPerMin:       5,
		ClientBurstSize:            2,
		S3UploadRequestsPerMin:     50,
		S3UploadBurstSize:          10,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
	}
	limiter, err := resilience.NewMultiLevelRateLimiter(config, nil, nil) // Using in-memory (no Redis)
	require.NoError(t, err)

	tests := []struct {
		name          string
		operationType string
		maxRequests   int
	}{
		{
			name:          "Build context rate limiting",
			operationType: "build_context",
			maxRequests:   2,
		},
		{
			name:          "K8s job rate limiting",
			operationType: "k8s_job",
			maxRequests:   3,
		},
		{
			name:          "Client rate limiting",
			operationType: "client",
			maxRequests:   2,
		},
		{
			name:          "S3 upload rate limiting",
			operationType: "s3_upload",
			maxRequests:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			successCount := 0
			for i := 0; i < tt.maxRequests+5; i++ {
				if limiter.Allow(tt.operationType) {
					successCount++
				}
			}

			// Assert
			assert.LessOrEqual(t, successCount, tt.maxRequests, "Should respect burst limit for %s", tt.operationType)
		})
	}
}

// TestBackend005_UnknownOperationType validates handling of unknown operation types.
func TestBackend005_UnknownOperationType(t *testing.T) {
	// Arrange
	config := resilience.MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 5,
		BuildContextBurstSize:      2,
		K8sJobRequestsPerMin:       10,
		K8sJobBurstSize:            3,
		ClientRequestsPerMin:       5,
		ClientBurstSize:            2,
		S3UploadRequestsPerMin:     50,
		S3UploadBurstSize:          10,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
	}
	limiter, err := resilience.NewMultiLevelRateLimiter(config, nil, nil) // Using in-memory (no Redis)
	require.NoError(t, err)

	// Act
	allowed := limiter.Allow("unknown_operation")

	// Assert
	assert.True(t, allowed, "Unknown operation types should be allowed by default")
}

// TestBackend005_MemoryMonitorCreation validates memory monitor creation.
func TestBackend005_MemoryMonitorCreation(t *testing.T) {
	tests := []struct {
		name            string
		maxUsagePercent float64
		checkInterval   time.Duration
	}{
		{
			name:            "Standard memory limit",
			maxUsagePercent: 80.0,
			checkInterval:   30 * time.Second,
		},
		{
			name:            "Conservative memory limit",
			maxUsagePercent: 70.0,
			checkInterval:   10 * time.Second,
		},
		{
			name:            "High memory threshold",
			maxUsagePercent: 90.0,
			checkInterval:   60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			monitor := resilience.NewMemoryMonitor(tt.maxUsagePercent, tt.checkInterval)

			// Assert
			require.NotNil(t, monitor)
			// Test behavior - monitor is created successfully (can't test unexported fields)
		})
	}
}

// TestBackend005_MemoryUsageCheck validates memory usage checking.
func TestBackend005_MemoryUsageCheck(t *testing.T) {
	// Arrange
	config := resilience.MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 5,
		BuildContextBurstSize:      2,
		K8sJobRequestsPerMin:       10,
		K8sJobBurstSize:            3,
		ClientRequestsPerMin:       5,
		ClientBurstSize:            2,
		S3UploadRequestsPerMin:     50,
		S3UploadBurstSize:          10,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
	}
	limiter, err := resilience.NewMultiLevelRateLimiter(config, nil, nil) // Using in-memory (no Redis)
	require.NoError(t, err)

	// Act
	isHigh := limiter.IsMemoryUsageHigh()

	// Assert
	// For now, this should return false as per the implementation
	assert.False(t, isHigh, "Memory usage check should work")
}

// TestBackend005_ConcurrentRateLimiting validates concurrent access to rate limiter.
func TestBackend005_ConcurrentRateLimiting(t *testing.T) {
	// Arrange
	limiter := resilience.NewRateLimiter(60, 10)
	numGoroutines := 20
	requestsPerGoroutine := 5

	// Act
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				if limiter.Allow() {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// Assert
	// Should allow at most burst size (10) initially
	assert.LessOrEqual(t, successCount, 10, "Should respect burst limit under concurrent access")
}

// TestBackend005_RateLimiterThreadSafety validates thread safety of rate limiter.
func TestBackend005_RateLimiterThreadSafety(t *testing.T) {
	// Arrange
	config := resilience.MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 60,
		BuildContextBurstSize:      10,
		K8sJobRequestsPerMin:       60,
		K8sJobBurstSize:            10,
		ClientRequestsPerMin:       60,
		ClientBurstSize:            10,
		S3UploadRequestsPerMin:     60,
		S3UploadBurstSize:          10,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
	}
	limiter, err := resilience.NewMultiLevelRateLimiter(config, nil, nil) // Using in-memory (no Redis)
	require.NoError(t, err)

	// Act - Concurrent access to different operation types
	var wg sync.WaitGroup
	operationTypes := []string{"build_context", "k8s_job", "client", "s3_upload"}

	for _, opType := range operationTypes {
		wg.Add(1)
		go func(op string) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				limiter.Allow(op)
			}
		}(opType)
	}

	// Assert - Should not panic
	wg.Wait()
	// If we get here without panicking, the test passes
}

// TestBackend005_TokenRefillBehavior validates token refill over time.
func TestBackend005_TokenRefillBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping time-dependent test in short mode")
	}

	// Arrange
	requestsPerMin := 60 // 1 per second
	burstSize := 2
	limiter := resilience.NewRateLimiter(requestsPerMin, burstSize)

	// Exhaust the bucket
	for i := 0; i < burstSize; i++ {
		limiter.Allow()
	}

	// Verify exhausted
	assert.False(t, limiter.Allow(), "Bucket should be exhausted")

	// Wait for partial refill
	time.Sleep(1500 * time.Millisecond) // Should add ~1-2 tokens

	// Act
	allowed := limiter.Allow()

	// Assert
	assert.True(t, allowed, "Token bucket should refill over time")
}

// TestBackend005_MultiLevelConfiguration validates full configuration.
func TestBackend005_MultiLevelConfiguration(t *testing.T) {
	// Arrange - Realistic configuration
	config := resilience.MultiLevelRateLimiterConfig{
		BuildContextRequestsPerMin: 5,
		BuildContextBurstSize:      2,
		K8sJobRequestsPerMin:       10,
		K8sJobBurstSize:            3,
		ClientRequestsPerMin:       5,
		ClientBurstSize:            2,
		S3UploadRequestsPerMin:     50,
		S3UploadBurstSize:          10,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
		CleanupInterval:            5 * time.Minute,
		ClientTTL:                  1 * time.Hour,
	}

	// Act
	limiter, err := resilience.NewMultiLevelRateLimiter(config, nil, nil) // Using in-memory (no Redis)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, limiter)

	// Verify all limiters are initialized by testing behavior
	// All operation types should work (testing that limiters are properly initialized)
	assert.True(t, limiter.Allow("build_context"), "build_context limiter should work")
	assert.True(t, limiter.Allow("k8s_job"), "k8s_job limiter should work")
	assert.True(t, limiter.Allow("client"), "client limiter should work")
	assert.True(t, limiter.Allow("s3_upload"), "s3_upload limiter should work")
}
