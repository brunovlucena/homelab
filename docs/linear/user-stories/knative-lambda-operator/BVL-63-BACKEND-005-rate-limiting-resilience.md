# 🌐 BACKEND-005: Rate Limiting and Resilience

**Priority**: P1 | **Status**: ✅ Implemented  | **Story Points**: 8  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-218/backend-005-rate-limiting-and-resilience

---

## 📋 User Story

**As a** Backend Developer  
**I want to** implement comprehensive rate limiting and resilience patterns  
**So that** the service can handle traffic spikes, protect downstream systems, and maintain stability under load

---

## 🎯 Acceptance Criteria

### ✅ Multi-Level Rate Limiting
- [ ] Global rate limit across all requests
- [ ] Per-endpoint rate limiting
- [ ] Per-third-party rate limiting
- [ ] Token bucket algorithm implementation
- [ ] Configurable limits and burst capacity
- [ ] Return 429 Too Many Requests when limit exceeded

### ✅ Concurrent Job Limits
- [ ] Maximum concurrent Kubernetes jobs (default: 10)
- [ ] Per-namespace job isolation
- [ ] Graceful rejection when limit reached
- [ ] Queue management when at capacity
- [ ] Auto-scaling considerations

### ✅ Failure Handling
- [ ] Exponential backoff for failed operations
- [ ] Circuit breaker for external dependencies
- [ ] Automatic retry with jitter
- [ ] Failure backoff period for failed builds
- [ ] Graceful degradation strategies

### ✅ Resource Protection
- [ ] Memory limits per operation
- [ ] CPU throttling under load
- [ ] Disk I/O rate limiting
- [ ] Network bandwidth management
- [ ] Kubernetes resource quotas enforcement

### ✅ Observability
- [ ] Rate limit metrics and traces
- [ ] Rejection logging with reasons
- [ ] Resource usage tracking
- [ ] Alert on rate limit threshold breaches
- [ ] Dashboard for rate limit status

---

## 🔧 Technical Implementation

### File: `internal/resilience/rate_limiter.go`

```go
// Multi-Level Rate Limiter
type MultiLevelRateLimiter struct {
    // Global rate limiter
    globalLimiter *rate.Limiter
    
    // Per-endpoint rate limiters
    endpointLimiters map[string]*rate.Limiter
    
    // Per-third-party rate limiters
    thirdPartyLimiters sync.Map
    
    // Configuration
    config *RateLimitConfig
    obs    *observability.Observability
}

type RateLimitConfig struct {
    // Global limits
    GlobalRPS   int // Requests per second
    GlobalBurst int // Burst capacity
    
    // Endpoint limits
    BuildStartRPS int
    BuildStartBurst int
    
    // Per-third-party limits
    ThirdPartyRPS int
    ThirdPartyBurst int
    
    // Job limits
    MaxConcurrentJobs int
}

// Check Rate Limit
func (r *MultiLevelRateLimiter) Allow(ctx context.Context, endpoint string, thirdPartyID string) bool {
    // 1. Check global rate limit
    if !r.globalLimiter.Allow() {
        r.obs.Info(ctx, "Global rate limit exceeded",
            "endpoint", endpoint,
            "third_party_id", thirdPartyID)
        return false
    }
    
    // 2. Check endpoint rate limit
    if endpointLimiter, ok := r.endpointLimiters[endpoint]; ok {
        if !endpointLimiter.Allow() {
            r.obs.Info(ctx, "Endpoint rate limit exceeded",
                "endpoint", endpoint,
                "third_party_id", thirdPartyID)
            return false
        }
    }
    
    // 3. Check per-third-party rate limit
    limiter := r.getOrCreateThirdPartyLimiter(thirdPartyID)
    if !limiter.Allow() {
        r.obs.Info(ctx, "Third-party rate limit exceeded",
            "third_party_id", thirdPartyID)
        return false
    }
    
    return true
}

func (r *MultiLevelRateLimiter) getOrCreateThirdPartyLimiter(thirdPartyID string) *rate.Limiter {
    if limiter, ok := r.thirdPartyLimiters.Load(thirdPartyID); ok {
        return limiter.(*rate.Limiter)
    }
    
    limiter := rate.NewLimiter(rate.Limit(r.config.ThirdPartyRPS), r.config.ThirdPartyBurst)
    r.thirdPartyLimiters.Store(thirdPartyID, limiter)
    return limiter
}
```

### File: `internal/handler/middleware.go`

```go
// Rate Limiting Middleware
func RateLimitMiddleware(rateLimiter *resilience.RateLimiter, obs *observability.Observability) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := r.Context()
            
            // Extract endpoint and third-party ID
            endpoint := r.URL.Path
            thirdPartyID := r.Header.Get("X-Third-Party-ID")
            
            // Check rate limit
            if !rateLimiter.Allow(endpoint) {
                obs.Info(ctx, "Request rate limited",
                    "endpoint", endpoint,
                    "third_party_id", thirdPartyID,
                    "remote_addr", r.RemoteAddr)
                
                w.Header().Set("X-RateLimit-Limit", "100")
                w.Header().Set("X-RateLimit-Remaining", "0")
                w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
                w.Header().Set("Retry-After", "60")
                
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Retry Logic with Exponential Backoff

```go
// Retry with exponential backoff
func retryWithBackoff(ctx context.Context, operation func() error, maxRetries int) error {
    backoff := 100 * time.Millisecond
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        // Don't retry on certain errors
        if isNonRetryableError(err) {
            return err
        }
        
        // Wait with exponential backoff and jitter
        jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(backoff + jitter):
            backoff *= 2
        }
    }
    
    return fmt.Errorf("operation failed after %d retries", maxRetries)
}
```

### Failure Backoff for Failed Builds

```go
// Check failure backoff period
func (j *JobManagerImpl) checkFailureBackoff(ctx context.Context, existingJob *batchv1.Job) (bool, error) {
    if existingJob.Status.Failed > 0 {
        jobAge := time.Since(existingJob.CreationTimestamp.Time)
        failureBackoffPeriod := 5 * time.Minute
        
        if jobAge < failureBackoffPeriod {
            j.obs.Info(ctx, "Job failed recently, implementing failure backoff",
                "job_name", existingJob.Name,
                "job_age", jobAge,
                "backoff_period", failureBackoffPeriod)
            
            return true, nil // In backoff period
        }
    }
    
    return false, nil // Not in backoff period
}
```

---

## 📊 Rate Limit Configuration

### Default Limits

```yaml
rate_limiting:
  # Global limits
  global_rps: 100
  global_burst: 200
  
  # Endpoint limits
  build_start_rps: 50
  build_start_burst: 100
  
  # Per-third-party limits
  third_party_rps: 10
  third_party_burst: 20
  
  # Job limits
  max_concurrent_jobs: 10
  
  # Failure handling
  failure_backoff_period: 5m
  max_retries: 3
  retry_base_delay: 100ms
```

---

## 🧪 Testing Scenarios

### 1. Global Rate Limit
```bash
# Send 200 requests rapidly (global limit: 100 RPS)
for i in {1..200}; do
  curl -X POST http://localhost:8080/events \
    -H "Ce-Type: network.notifi.lambda.build.start" \
    -d "{\"parser_id\":\"parser-$i\"}" &
done
```

**Expected**:
- First ~100 requests succeed (200 OK)
- Remaining requests return 429 Too Many Requests
- Headers include rate limit information
- Service remains stable

### 2. Per-Third-Party Rate Limit
```bash
# Send 30 requests for same third-party (limit: 10 RPS)
for i in {1..30}; do
  curl -X POST http://localhost:8080/events \
    -H "Ce-Source: network.notifi.customer-123" \
    -d "{\"parser_id\":\"parser-$i\"}" &
done
```

**Expected**:
- First ~10 requests succeed
- Remaining requests rate limited
- Other third-parties not affected

### 3. Concurrent Job Limit
```bash
# Trigger 15 builds (limit: 10 concurrent)
for i in {1..15}; do
  make trigger-build-dev PARSER_ID=parser-$i &
done
```

**Expected**:
- First 10 jobs created
- Jobs 11-15 return error: "concurrent job limit reached"
- As jobs complete, new jobs can be created

### 4. Failure Backoff
```bash
# Trigger build that will fail
make trigger-build-dev PARSER_ID=invalid-parser

# Wait 2 minutes and retry
sleep 120
make trigger-build-dev PARSER_ID=invalid-parser
```

**Expected**:
- First build fails
- Second attempt within 5 minutes returns: "build failed recently, backoff period active"
- After 5 minutes, new build attempt allowed

---

## 📈 Performance Requirements

- **Rate Limit Check**: < 1ms per request
- **Token Bucket Update**: < 100μs
- **Concurrent Job Count**: < 10ms
- **Retry Delay**: 100ms base, max 3.2s
- **Memory Overhead**: < 10MB for all rate limiters

---

## 🔍 Monitoring & Alerts

### Metrics
- `rate_limit_requests_total{limit_type="global | endpoint | third_party"}` - Total requests
- `rate_limit_rejected_total{limit_type="global | endpoint | third_party"}` - Rejected requests
- `rate_limit_current_usage{limit_type="global | endpoint | third_party"}` - Current usage
- `concurrent_jobs_active` - Active job count
- `concurrent_jobs_rejected_total` - Jobs rejected due to limit
- `retry_attempts_total{success="true | false"}` - Retry attempts

### Alerts
- **High Rate Limit Rejection**: Alert if > 20% requests rejected over 5 minutes
- **Concurrent Job Limit**: Alert when hitting 90% of job limit
- **Retry Exhaustion**: Alert if retry failures > 5% of operations
- **Rate Limiter Error**: Alert on any rate limiter failures

### Response Headers
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 47
X-RateLimit-Reset: 1635523200
Retry-After: 60
```

---

## 🏗️ Code References

**Main Files**:
- `internal/resilience/rate_limiter.go` - Rate limiting implementation
- `internal/resilience/resilience.go` - Retry and backoff logic
- `internal/handler/middleware.go` - Rate limiting middleware
- `internal/config/rate_limiting.go` - Rate limit configuration

**Libraries**:
- `golang.org/x/time/rate` - Token bucket rate limiter

---

## 📚 Related Documentation

- [BACKEND-003: Kubernetes Job Lifecycle](BACKEND-003-kubernetes-job-lifecycle.md)
- [BACKEND-004: Async Job Processing](BACKEND-004-async-job-processing.md)
- Rate Limiting Patterns: https://cloud.google.com/architecture/rate-limiting-strategies-techniques

---

**Last Updated**: October 29, 2025  
**Owner**: Backend Team  
**Status**: Production Ready

