# 🌐 BACKEND-008: Error Handling and Logging

**Priority**: P1 | **Status**: ✅ Implemented  | **Story Points**: 5
**Linear URL**: https://linear.app/bvlucena/issue/BVL-228/backend-008-error-handling-and-logging

---

## 📋 User Story

**As a** Backend Developer  
**I want to** implement comprehensive error handling and logging  
**So that** errors are caught gracefully, logged with context, and provide actionable debugging information

---

## 🎯 Acceptance Criteria

### ✅ Error Types and Classification
- [ ] Configuration errors (invalid config)
- [ ] Validation errors (bad input)
- [ ] System errors (K8s API failures)
- [ ] Not found errors (resource missing)
- [ ] Conflict errors (race conditions)
- [ ] Each error type has specific HTTP status code

### ✅ Error Context
- [ ] Include correlation ID in all errors
- [ ] Include trace ID for debugging
- [ ] Include operation context
- [ ] Include resource identifiers
- [ ] Include error chain (wrapped errors)
- [ ] Include timestamp

### ✅ Error Responses
- [ ] Structured JSON error responses
- [ ] User-friendly error messages
- [ ] Technical error details (in dev mode)
- [ ] Error codes for programmatic handling
- [ ] Retry hints where applicable
- [ ] Include support correlation ID

### ✅ Error Logging
- [ ] Log all errors at ERROR level
- [ ] Include full error context
- [ ] Include stack traces for critical errors
- [ ] Rate limit repetitive errors
- [ ] Alert on critical error patterns
- [ ] Structured log format for parsing

### ✅ Error Recovery
- [ ] Graceful degradation
- [ ] Automatic retries with backoff
- [ ] Circuit breaker for failing dependencies
- [ ] Resource cleanup on errors
- [ ] Transaction rollback where applicable

---

## 🔧 Technical Implementation

### File: `internal/errors/errors.go`

```go
// Error Types
type ErrorType string

const (
    ErrorTypeConfiguration ErrorType = "configuration_error"
    ErrorTypeValidation    ErrorType = "validation_error"
    ErrorTypeSystem        ErrorType = "system_error"
    ErrorTypeNotFound      ErrorType = "not_found_error"
    ErrorTypeConflict      ErrorType = "conflict_error"
)

// AppError represents application errors with context
type AppError struct {
    Type        ErrorType
    Component   string
    Field       string
    Value       interface{}
    Message     string
    Cause       error
    Timestamp   time.Time
    TraceID     string
    CorrelationID string
}

func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s: %v", e.Component, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Component, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Cause
}

// Error Constructors
func NewConfigurationError(component, field, message string) error {
    return &AppError{
        Type:      ErrorTypeConfiguration,
        Component: component,
        Field:     field,
        Message:   message,
        Timestamp: time.Now(),
    }
}

func NewValidationError(field string, value interface{}, message string) error {
    return &AppError{
        Type:      ErrorTypeValidation,
        Component: "validation",
        Field:     field,
        Value:     value,
        Message:   message,
        Timestamp: time.Now(),
    }
}

func NewSystemError(component, operation string) error {
    return &AppError{
        Type:      ErrorTypeSystem,
        Component: component,
        Message:   fmt.Sprintf("%s operation failed", operation),
        Timestamp: time.Now(),
    }
}

func NewNotFoundError(component, resourceType, resourceID string) error {
    return &AppError{
        Type:      ErrorTypeNotFound,
        Component: component,
        Message:   fmt.Sprintf("%s not found: %s", resourceType, resourceID),
        Timestamp: time.Now(),
    }
}

// Error Type Checking
func IsConfigurationError(err error) bool {
    var appErr *AppError
    return errors.As(err, &appErr) && appErr.Type == ErrorTypeConfiguration
}

func IsValidationError(err error) bool {
    var appErr *AppError
    return errors.As(err, &appErr) && appErr.Type == ErrorTypeValidation
}

func IsNotFoundError(err error) bool {
    var appErr *AppError
    return errors.As(err, &appErr) && appErr.Type == ErrorTypeNotFound
}
```

### Error Response Handling

```go
// HTTP Error Response
type ErrorResponse struct {
    Error       string    `json:"error"`
    ErrorType   string    `json:"error_type"`
    Details     string    `json:"details,omitempty"`
    CorrelationID string  `json:"correlation_id"`
    TraceID     string    `json:"trace_id,omitempty"`
    Timestamp   time.Time `json:"timestamp"`
    RetryAfter  int       `json:"retry_after,omitempty"`
}

func HandleError(ctx context.Context, w http.ResponseWriter, err error, obs *observability.Observability) {
    // Extract context
    correlationID := observability.GetCorrelationID(ctx)
    traceID := ""
    if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
        traceID = span.SpanContext().TraceID().String()
    }
    
    // Determine status code and response
    var statusCode int
    var errorType string
    var message string
    var retryAfter int
    
    var appErr *AppError
    if errors.As(err, &appErr) {
        switch appErr.Type {
        case ErrorTypeConfiguration:
            statusCode = http.StatusInternalServerError
            errorType = "configuration_error"
            message = "Service configuration error"
        case ErrorTypeValidation:
            statusCode = http.StatusBadRequest
            errorType = "validation_error"
            message = appErr.Message
        case ErrorTypeNotFound:
            statusCode = http.StatusNotFound
            errorType = "not_found"
            message = appErr.Message
        case ErrorTypeConflict:
            statusCode = http.StatusConflict
            errorType = "conflict"
            message = "Resource conflict detected"
            retryAfter = 5
        default:
            statusCode = http.StatusInternalServerError
            errorType = "system_error"
            message = "Internal server error"
        }
    } else {
        statusCode = http.StatusInternalServerError
        errorType = "unknown_error"
        message = "An unexpected error occurred"
    }
    
    // Log error with full context
    obs.Error(ctx, err, message,
        "status_code", statusCode,
        "error_type", errorType,
        "correlation_id", correlationID,
        "trace_id", traceID)
    
    // Create error response
    errorResponse := ErrorResponse{
        Error:         message,
        ErrorType:     errorType,
        Details:       err.Error(),
        CorrelationID: correlationID,
        TraceID:       traceID,
        Timestamp:     time.Now(),
        RetryAfter:    retryAfter,
    }
    
    // Send response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(errorResponse)
}
```

### Error Recovery with Retry

```go
// Retry operation with exponential backoff
func WithRetry(ctx context.Context, operation string, fn func() error, maxRetries int, obs *observability.Observability) error {
    backoff := 100 * time.Millisecond
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        // Log attempt
        obs.Info(ctx, "Operation failed, retrying",
            "operation", operation,
            "attempt", attempt+1,
            "max_retries", maxRetries,
            "error", err.Error())
        
        // Check if error is retryable
        if !isRetryableError(err) {
            obs.Info(ctx, "Error not retryable, giving up",
                "operation", operation,
                "error", err.Error())
            return err
        }
        
        // Last attempt, return error
        if attempt == maxRetries {
            return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
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

func isRetryableError(err error) bool {
    // Network errors are retryable
    if errors.Is(err, context.DeadlineExceeded) {
        return true
    }
    
    // K8s API errors
    if apierrors.IsServerTimeout(err) | | apierrors.IsServiceUnavailable(err) {
        return true
    }
    
    // Configuration and validation errors are not retryable
    if IsConfigurationError(err) | | IsValidationError(err) {
        return false
    }
    
    return true
}
```

---

## 📊 Error Response Examples

### Configuration Error (500)
```json
{
  "error": "Service configuration error",
  "error_type": "configuration_error",
  "details": "kubernetes config cannot be nil",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "trace_id": "abc123def456",
  "timestamp": "2025-10-29T12:34:56Z"
}
```

### Validation Error (400)
```json
{
  "error": "Invalid build request",
  "error_type": "validation_error",
  "details": "parser_id is required",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "trace_id": "abc123def456",
  "timestamp": "2025-10-29T12:34:56Z"
}
```

### Conflict Error (409)
```json
{
  "error": "Resource conflict detected",
  "error_type": "conflict",
  "details": "Job already exists and is running",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "trace_id": "abc123def456",
  "timestamp": "2025-10-29T12:34:56Z",
  "retry_after": 5
}
```

---

## 🧪 Testing Scenarios

### 1. Validation Error
```bash
curl -X POST http://localhost:8080/events \
  -H "Ce-Type: network.notifi.lambda.build.start" \
  -d '{}'
```

**Expected**: 400 with validation error

### 2. Not Found Error
```bash
curl http://localhost:8080/builds/non-existent-build
```

**Expected**: 404 with not found error

### 3. Conflict Error
```bash
# Start same build twice
curl -X POST http://localhost:8080/events \
  -H "Ce-Type: network.notifi.lambda.build.start" \
  -d '{"parser_id":"test"}' &
curl -X POST http://localhost:8080/events \
  -H "Ce-Type: network.notifi.lambda.build.start" \
  -d '{"parser_id":"test"}' &
```

**Expected**: One succeeds, one returns 409 conflict

---

## 🔍 Monitoring & Alerts

### Metrics
```promql
# Error rate by type
rate(errors_total{error_type="validation_error"}[5m])

# 5xx error rate
rate(http_requests_total{status=~"5.."}[5m])

# Error rate percentage
rate(errors_total[5m]) / rate(http_requests_total[5m]) * 100
```

### Alerts
- **High Error Rate**: Alert if error rate > 5% for 5 minutes
- **Critical Errors**: Alert on any system/configuration errors
- **Error Spike**: Alert on sudden increase in error rate

---

## 🏗️ Code References

**Main Files**:
- `internal/errors/errors.go` - Error types and constructors
- `internal/handler/middleware.go` - Error handling middleware
- `internal/observability/observability.go` - Error logging

---

**Last Updated**: October 29, 2025  
**Owner**: Backend Team  
**Status**: Production Ready

