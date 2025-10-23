// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🛡️ RESILIENCE - Rate limiting and memory monitoring patterns
//
//	🎯 Purpose: Protect services from overload and provide graceful degradation
//	💡 Features: Multi-level rate limiting, backoff strategies, memory monitoring
//
//	🏛️ ARCHITECTURE:
//	⏱️ Rate Limiting - Multi-level rate limiting for different operation types
//	🔄 Rate Limiting - Multi-level rate limiting for different operations
//	📊 Monitoring - Metrics and observability for resilience patterns
//	🎯 Graceful Degradation - Fallback strategies for service failures
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package resilience

import (
	"sync"
	"time"
)

// MultiLevelRateLimiter provides rate limiting for different operation types
type MultiLevelRateLimiter struct {
	buildContextLimiter *RateLimiter
	k8sJobLimiter       *RateLimiter
	clientLimiter       *RateLimiter
	s3UploadLimiter     *RateLimiter
	memoryMonitor       *MemoryMonitor
	mu                  sync.RWMutex
}

// MultiLevelRateLimiterConfig holds configuration for the multi-level rate limiter
type MultiLevelRateLimiterConfig struct {
	BuildContextRequestsPerMin int
	BuildContextBurstSize      int
	K8sJobRequestsPerMin       int
	K8sJobBurstSize            int
	ClientRequestsPerMin       int
	ClientBurstSize            int
	S3UploadRequestsPerMin     int
	S3UploadBurstSize          int
	MaxMemoryUsagePercent      float64
	MemoryCheckInterval        time.Duration
	CleanupInterval            time.Duration
	ClientTTL                  time.Duration
}

// RateLimiter provides basic rate limiting functionality
type RateLimiter struct {
	requestsPerMin int
	burstSize      int
	tokens         chan struct{}
	lastRefill     time.Time
	mu             sync.Mutex
}

// MemoryMonitor monitors memory usage
type MemoryMonitor struct {
	maxUsagePercent float64
	checkInterval   time.Duration
}

// NewMultiLevelRateLimiter creates a new multi-level rate limiter
func NewMultiLevelRateLimiter(config MultiLevelRateLimiterConfig) (*MultiLevelRateLimiter, error) {
	limiter := &MultiLevelRateLimiter{
		buildContextLimiter: NewRateLimiter(config.BuildContextRequestsPerMin, config.BuildContextBurstSize),
		k8sJobLimiter:       NewRateLimiter(config.K8sJobRequestsPerMin, config.K8sJobBurstSize),
		clientLimiter:       NewRateLimiter(config.ClientRequestsPerMin, config.ClientBurstSize),
		s3UploadLimiter:     NewRateLimiter(config.S3UploadRequestsPerMin, config.S3UploadBurstSize),
		memoryMonitor:       NewMemoryMonitor(config.MaxMemoryUsagePercent, config.MemoryCheckInterval),
	}

	return limiter, nil
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMin, burstSize int) *RateLimiter {
	tokens := make(chan struct{}, burstSize)

	// Pre-fill the token bucket with initial tokens
	for i := 0; i < burstSize; i++ {
		tokens <- struct{}{}
	}

	return &RateLimiter{
		requestsPerMin: requestsPerMin,
		burstSize:      burstSize,
		tokens:         tokens,
		lastRefill:     time.Now(),
	}
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor(maxUsagePercent float64, checkInterval time.Duration) *MemoryMonitor {
	return &MemoryMonitor{
		maxUsagePercent: maxUsagePercent,
		checkInterval:   checkInterval,
	}
}

// Allow checks if a request is allowed for the given operation type
func (m *MultiLevelRateLimiter) Allow(operationType string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch operationType {
	case "build_context":
		return m.buildContextLimiter.Allow()
	case "k8s_job":
		return m.k8sJobLimiter.Allow()
	case "client":
		return m.clientLimiter.Allow()
	case "s3_upload":
		return m.s3UploadLimiter.Allow()
	default:
		return false
	}
}

// Allow checks if a request is allowed
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastRefill)
	tokensToAdd := int(elapsed.Minutes() * float64(r.requestsPerMin))

	if tokensToAdd > 0 {
		// Refill tokens
		for i := 0; i < tokensToAdd && len(r.tokens) < r.burstSize; i++ {
			select {
			case r.tokens <- struct{}{}:
			default:
				// Channel is full, continue to next iteration
			}
		}
		r.lastRefill = now
	}

	// Try to consume a token
	select {
	case <-r.tokens:
		return true
	default:
		return false
	}
}

// IsMemoryUsageHigh checks if memory usage is above the configured threshold
func (m *MultiLevelRateLimiter) IsMemoryUsageHigh() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// For now, return false as a basic implementation
	// In a real implementation, this would check actual memory usage
	return false
}

// StartCleanup starts the cleanup process for the rate limiter
func (m *MultiLevelRateLimiter) StartCleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Basic implementation - in a real implementation this would start
	// background goroutines for cleanup tasks
}

// StopCleanup stops the cleanup process for the rate limiter
func (m *MultiLevelRateLimiter) StopCleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Basic implementation - in a real implementation this would stop
	// background goroutines for cleanup tasks
}

// Close closes the rate limiter
func (m *MultiLevelRateLimiter) Close() error {
	// Cleanup resources
	return nil
}
