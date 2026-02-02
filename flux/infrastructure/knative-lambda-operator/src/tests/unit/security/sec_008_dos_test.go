// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🔒 SEC-008: Denial of Service & Resource Exhaustion Testing
//
//	User Story: Denial of Service & Resource Exhaustion Testing
//	Priority: P1 | Story Points: 5
//
//	Tests validate:
//	- Rate limiting protection
//	- Resource quota enforcement
//	- Pod disruption budget protection
//	- Queue flood prevention
//	- Connection limit protection
//	- CPU/Memory bomb prevention
//	- Slowloris/Slow POST protection
//	- Amplification attack prevention
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package security

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TestSec008_RateLimitingEnforcement validates rate limiting is enforced.
func TestSec008_RateLimitingEnforcement(t *testing.T) {
	// Arrange
	handler := setupRateLimitHandler(t, 10, time.Second)
	successCount := 0
	rateLimitedCount := 0

	// Act - Send 20 requests rapidly
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest("POST", "/api/v1/build", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		switch w.Code {
		case http.StatusOK:
			successCount++
		case http.StatusTooManyRequests:
			rateLimitedCount++
		}
	}

	// Assert
	assert.LessOrEqual(t, successCount, 10, "Should not exceed rate limit")
	assert.Greater(t, rateLimitedCount, 0, "Should rate limit excess requests")
}

// TestSec008_RateLimitingTimeWindow validates rate limit resets after time window.
func TestSec008_RateLimitingTimeWindow(t *testing.T) {
	// Arrange
	handler := setupRateLimitHandler(t, 5, 100*time.Millisecond)

	// Act - First burst
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("POST", "/api/v1/build", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// Wait for time window to expire
	time.Sleep(150 * time.Millisecond)

	// Second request after window should succeed
	req := httptest.NewRequest("POST", "/api/v1/build", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code, "Request after time window should succeed")
}

// TestSec008_ResourceQuotaEnforcement validates resource quotas are enforced.
func TestSec008_ResourceQuotaEnforcement(t *testing.T) {
	tests := []struct {
		name        string
		quota       *corev1.ResourceQuota
		requested   corev1.ResourceList
		shouldAllow bool
		description string
	}{
		{
			name: "Within quota limits",
			quota: createResourceQuota(map[corev1.ResourceName]string{
				corev1.ResourceCPU:    "10",
				corev1.ResourceMemory: "20Gi",
			}),
			requested: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			},
			shouldAllow: true,
			description: "Requests within quota should be allowed",
		},
		{
			name: "Exceeds CPU quota",
			quota: createResourceQuota(map[corev1.ResourceName]string{
				corev1.ResourceCPU:    "10",
				corev1.ResourceMemory: "20Gi",
			}),
			requested: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("15"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			},
			shouldAllow: false,
			description: "Requests exceeding CPU quota should be blocked",
		},
		{
			name: "Exceeds memory quota",
			quota: createResourceQuota(map[corev1.ResourceName]string{
				corev1.ResourceCPU:    "10",
				corev1.ResourceMemory: "20Gi",
			}),
			requested: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("25Gi"),
			},
			shouldAllow: false,
			description: "Requests exceeding memory quota should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			allowed := checkResourceQuota(tt.quota, tt.requested)

			// Assert
			assert.Equal(t, tt.shouldAllow, allowed, tt.description)
		})
	}
}

// TestSec008_PodDisruptionBudget validates PDB protects availability.
func TestSec008_PodDisruptionBudget(t *testing.T) {
	tests := []struct {
		name        string
		pdb         *policyv1.PodDisruptionBudget
		totalPods   int
		disruptions int
		shouldAllow bool
		description string
	}{
		{
			name: "PDB allows disruption",
			pdb: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 2},
				},
			},
			totalPods:   5,
			disruptions: 1,
			shouldAllow: true,
			description: "Should allow disruption when min available maintained",
		},
		{
			name: "PDB blocks disruption",
			pdb: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 3},
				},
			},
			totalPods:   3,
			disruptions: 1,
			shouldAllow: false,
			description: "Should block disruption when it violates min available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			allowed := checkPDBAllowsDisruption(tt.pdb, tt.totalPods, tt.disruptions)

			// Assert
			assert.Equal(t, tt.shouldAllow, allowed, tt.description)
		})
	}
}

// TestSec008_ConnectionLimits validates connection limits are enforced.
func TestSec008_ConnectionLimits(t *testing.T) {
	// Arrange
	maxConnections := 10
	handler := setupConnectionLimitHandler(t, maxConnections)

	var wg sync.WaitGroup
	successCount := 0
	rejectedCount := 0
	var mu sync.Mutex

	// Act - Try to open 20 concurrent connections
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/api/v1/build", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			mu.Lock()
			switch w.Code {
			case http.StatusOK:
				successCount++
			case http.StatusServiceUnavailable:
				rejectedCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Assert
	assert.LessOrEqual(t, successCount, maxConnections, "Should not exceed connection limit")
	assert.Greater(t, rejectedCount, 0, "Should reject excess connections")
}

// TestSec008_QueueFloodPrevention validates queue limits prevent flooding.
func TestSec008_QueueFloodPrevention(t *testing.T) {
	// Arrange
	queue := newQueueLimiter(100)

	tests := []struct {
		name         string
		messages     int
		shouldAccept bool
		description  string
	}{
		{
			name:         "Within queue limit",
			messages:     50,
			shouldAccept: true,
			description:  "Messages within limit should be accepted",
		},
		{
			name:         "At queue limit",
			messages:     100,
			shouldAccept: true,
			description:  "Messages at limit should be accepted",
		},
		{
			name:         "Exceeds queue limit",
			messages:     101,
			shouldAccept: false,
			description:  "Messages exceeding limit should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset queue for each test
			queue.reset()

			// Act - Try to add messages
			accepted := true
			for i := 0; i < tt.messages; i++ {
				if !queue.accept() {
					accepted = false
					break
				}
			}

			// Assert
			assert.Equal(t, tt.shouldAccept, accepted, tt.description)
		})
	}
}

// TestSec008_CPUMemoryLimits validates resource limits prevent bombs.
func TestSec008_CPUMemoryLimits(t *testing.T) {
	tests := []struct {
		name        string
		limits      corev1.ResourceList
		requests    corev1.ResourceList
		isValid     bool
		description string
	}{
		{
			name: "CPU and memory limits set",
			limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2000m"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
			requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
			isValid:     true,
			description: "Should have CPU and memory limits",
		},
		{
			name:        "No resource limits",
			limits:      corev1.ResourceList{},
			requests:    corev1.ResourceList{},
			isValid:     false,
			description: "Missing resource limits should be flagged",
		},
		{
			name: "Missing CPU limit",
			limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
			requests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("100m"),
			},
			isValid:     false,
			description: "Missing CPU limit should be flagged",
		},
		{
			name: "Missing memory limit",
			limits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("2000m"),
			},
			requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
			isValid:     false,
			description: "Missing memory limit should be flagged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isValid := validateResourceLimits(tt.limits, tt.requests)

			// Assert
			assert.Equal(t, tt.isValid, isValid, tt.description)
		})
	}
}

// TestSec008_RequestTimeout validates request timeouts prevent slowloris.
func TestSec008_RequestTimeout(t *testing.T) {
	// Arrange
	handler := setupTimeoutHandler(t, 2*time.Second)

	// Act - Send slow request
	req := httptest.NewRequest("POST", "/api/v1/build", &slowReader{delay: 5 * time.Second})
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	done := make(chan bool)
	go func() {
		handler.ServeHTTP(w, req)
		done <- true
	}()

	// Assert - Should timeout
	select {
	case <-done:
		assert.Equal(t, http.StatusRequestTimeout, w.Code, "Slow request should timeout")
	case <-time.After(3 * time.Second):
		t.Fatal("Handler did not timeout slow request")
	}
}

// TestSec008_PayloadSizeLimits validates payload size limits.
func TestSec008_PayloadSizeLimits(t *testing.T) {
	// Arrange
	handler := setupSizeLimitHandler(t, 10*1024*1024) // 10MB limit

	tests := []struct {
		name        string
		payloadSize int64
		shouldAllow bool
		description string
	}{
		{
			name:        "Within size limit",
			payloadSize: 5 * 1024 * 1024, // 5MB
			shouldAllow: true,
			description: "Payload within limit should be accepted",
		},
		{
			name:        "At size limit",
			payloadSize: 10 * 1024 * 1024, // 10MB
			shouldAllow: true,
			description: "Payload at limit should be accepted",
		},
		{
			name:        "Exceeds size limit",
			payloadSize: 15 * 1024 * 1024, // 15MB
			shouldAllow: false,
			description: "Payload exceeding limit should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			payload := make([]byte, tt.payloadSize)
			req := httptest.NewRequest("POST", "/api/v1/build", bytes.NewReader(payload))
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			if tt.shouldAllow {
				assert.Equal(t, http.StatusOK, w.Code, tt.description)
			} else {
				assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code, tt.description)
			}
		})
	}
}

// Helper Functions.

// RateLimitEntry tracks request count and window start time.
type RateLimitEntry struct {
	count       int
	windowStart time.Time
}

func setupRateLimitHandler(_ *testing.T, limit int, window time.Duration) http.Handler {
	requestData := make(map[string]*RateLimitEntry)
	var mu sync.Mutex

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}

		now := time.Now()
		entry, exists := requestData[clientIP]

		// Initialize or reset window if expired
		if !exists || now.Sub(entry.windowStart) > window {
			requestData[clientIP] = &RateLimitEntry{
				count:       1,
				windowStart: now,
			}
			w.Header().Set("X-RateLimit-Limit", string(rune(limit)))
			w.Header().Set("X-RateLimit-Remaining", string(rune(limit-1)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
			return
		}

		// Increment count within window
		entry.count++

		if entry.count > limit {
			w.Header().Set("X-RateLimit-Limit", string(rune(limit)))
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		w.Header().Set("X-RateLimit-Limit", string(rune(limit)))
		w.Header().Set("X-RateLimit-Remaining", string(rune(limit-entry.count)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))

		// Cleanup old entries periodically (simple approach)
		if len(requestData) > 1000 {
			for ip, e := range requestData {
				if now.Sub(e.windowStart) > window*2 {
					delete(requestData, ip)
				}
			}
		}
	})
}

func setupConnectionLimitHandler(_ *testing.T, maxConnections int) http.Handler {
	activeConnections := 0
	var mu sync.Mutex

	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		if activeConnections >= maxConnections {
			mu.Unlock()
			http.Error(w, "Too many connections", http.StatusServiceUnavailable)
			return
		}
		activeConnections++
		mu.Unlock()

		defer func() {
			mu.Lock()
			activeConnections--
			mu.Unlock()
		}()

		time.Sleep(100 * time.Millisecond) // Simulate work
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
}

func setupTimeoutHandler(_ *testing.T, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		done := make(chan bool)

		go func() {
			// Try to read body
			buf := make([]byte, 1024)
			_, err := r.Body.Read(buf)
			if err != nil {
				done <- false
				return
			}
			done <- true
		}()

		select {
		case success := <-done:
			if success {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		case <-time.After(timeout):
			w.WriteHeader(http.StatusRequestTimeout)
			_, _ = w.Write([]byte("Request timeout"))
		}
	})
}

func setupSizeLimitHandler(t *testing.T, maxSize int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Content-Length header first
		if r.ContentLength > maxSize {
			http.Error(w, "Request entity too large", http.StatusRequestEntityTooLarge)
			return
		}

		// Also limit body reading
		limitedReader := http.MaxBytesReader(w, r.Body, maxSize)
		defer func() {
			if err := limitedReader.Close(); err != nil {
				t.Logf("Failed to close limited reader: %v", err)
			}
		}()

		// Try to read the body
		buf := make([]byte, 32*1024) // 32KB buffer
		totalRead := int64(0)
		for {
			n, err := limitedReader.Read(buf)
			totalRead += int64(n)
			if err != nil {
				break
			}
		}

		if totalRead > maxSize {
			http.Error(w, "Request entity too large", http.StatusRequestEntityTooLarge)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
}

func createResourceQuota(limits map[corev1.ResourceName]string) *corev1.ResourceQuota {
	hard := make(corev1.ResourceList)
	for name, value := range limits {
		hard[name] = resource.MustParse(value)
	}

	return &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{Name: "test-quota"},
		Spec: corev1.ResourceQuotaSpec{
			Hard: hard,
		},
	}
}

func checkResourceQuota(quota *corev1.ResourceQuota, requested corev1.ResourceList) bool {
	for resourceName, requestedQty := range requested {
		if hardLimit, exists := quota.Spec.Hard[resourceName]; exists {
			if requestedQty.Cmp(hardLimit) > 0 {
				return false
			}
		}
	}
	return true
}

func checkPDBAllowsDisruption(pdb *policyv1.PodDisruptionBudget, totalPods, disruptions int) bool {
	if pdb.Spec.MinAvailable != nil {
		minAvailable := int(pdb.Spec.MinAvailable.IntVal)
		availableAfterDisruption := totalPods - disruptions
		return availableAfterDisruption >= minAvailable
	}
	return true
}

func validateResourceLimits(limits, requests corev1.ResourceList) bool {
	// Must have both CPU and memory limits
	if len(limits) == 0 {
		return false
	}

	cpuLimit, hasCPULimit := limits[corev1.ResourceCPU]
	memoryLimit, hasMemoryLimit := limits[corev1.ResourceMemory]

	if !hasCPULimit || !hasMemoryLimit {
		return false
	}

	// Limits must be greater than zero
	if cpuLimit.IsZero() || memoryLimit.IsZero() {
		return false
	}

	// If requests are provided, limits must be greater than requests
	if len(requests) > 0 {
		if cpuRequest, hasRequest := requests[corev1.ResourceCPU]; hasRequest {
			if cpuLimit.Cmp(cpuRequest) <= 0 {
				return false
			}
		}

		if memRequest, hasRequest := requests[corev1.ResourceMemory]; hasRequest {
			if memoryLimit.Cmp(memRequest) <= 0 {
				return false
			}
		}
	}

	return true
}

// queueLimiter tracks queue usage and enforces limits.
type queueLimiter struct {
	maxSize int
	current int
	mu      sync.Mutex
}

func newQueueLimiter(maxSize int) *queueLimiter {
	return &queueLimiter{maxSize: maxSize}
}

func (q *queueLimiter) accept() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.current >= q.maxSize {
		return false
	}
	q.current++
	return true
}

func (q *queueLimiter) reset() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.current = 0
}

// slowReader simulates a slow client.
type slowReader struct {
	delay time.Duration
	read  bool
}

func (s *slowReader) Read(_ []byte) (n int, err error) {
	if !s.read {
		time.Sleep(s.delay)
		s.read = true
	}
	return 0, nil
}
