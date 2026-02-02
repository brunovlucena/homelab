# 🌐 BACKEND-009: API Management and Build Operations

**Priority**: P1 | **Status**: ✅ Implemented  | **Story Points**: 5
**Linear URL**: https://linear.app/bvlucena/issue/BVL-229/backend-009-api-management-and-build-operations

---

## 📋 User Story

**As a** Backend Developer  
**I want to** provide RESTful API endpoints for build management  
**So that** users can list, retrieve, and cancel builds programmatically

---

## 🎯 Acceptance Criteria

### ✅ Build Listing API
- [ ] GET `/builds` endpoint with pagination
- [ ] Filter by `third_party_id`
- [ ] Filter by `parser_id`
- [ ] Filter by `status` (pending, running, succeeded, failed)
- [ ] Configurable page size (default: 30, max: 100)
- [ ] Return total count and build list
- [ ] Include timestamps and duration

### ✅ Build Retrieval API
- [ ] GET `/builds/{id}` endpoint
- [ ] Return detailed build information
- [ ] Include job status and events
- [ ] Include build logs (last N lines)
- [ ] Include error messages if failed
- [ ] Return 404 if build not found

### ✅ Build Cancellation API
- [ ] POST `/builds/{id}/cancel` endpoint
- [ ] Cancel running Kubernetes job
- [ ] Clean up resources
- [ ] Emit cancellation event
- [ ] Return 409 if already completed
- [ ] Return 404 if build not found

### ✅ Stats and Monitoring API
- [ ] GET `/async-jobs/stats` endpoint
- [ ] Return worker pool statistics
- [ ] Return queue depth
- [ ] Return processing metrics
- [ ] Real-time status information

### ✅ Response Format
- [ ] Consistent JSON structure
- [ ] Include pagination metadata
- [ ] Include timestamps (ISO 8601)
- [ ] Include correlation IDs
- [ ] Support CORS headers

---

## 🔧 Technical Implementation

### File: `internal/handler/http_handler.go`

```go
// Build Listing Endpoint
func (h *HTTPHandlerImpl) HandleListBuilds(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse query parameters
    thirdPartyID := r.URL.Query().Get("third_party_id")
    parserID := r.URL.Query().Get("parser_id")
    status := r.URL.Query().Get("status")
    limit := h.config.DefaultListLimit
    
    if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
        if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
            if parsedLimit > 0 && parsedLimit <= h.config.MaxListLimit {
                limit = parsedLimit
            }
        }
    }
    
    // Start tracing
    ctx, span := h.obs.StartSpanWithAttributes(ctx, "list_builds", map[string]string{
        "third_party_id": thirdPartyID,
        "parser_id":      parserID,
        "status":         status,
        "limit":          strconv.Itoa(limit),
    })
    defer span.End()
    
    start := time.Now()
    
    // Get job manager
    jobManager := h.container.GetJobManager()
    if jobManager == nil {
        h.handleError(ctx, w, "Job manager not available", 
            fmt.Errorf("job manager not available"), 
            http.StatusServiceUnavailable)
        return
    }
    
    // List builds with filters
    builds, err := h.listBuildsWithFilters(ctx, jobManager, thirdPartyID, parserID, status, limit)
    if err != nil {
        h.handleError(ctx, w, "Failed to list builds", err, http.StatusInternalServerError)
        return
    }
    
    // Create response
    response := map[string]interface{}{
        "builds":    builds,
        "count":     len(builds),
        "limit":     limit,
        "success":   true,
        "timestamp": time.Now().UTC(),
        "duration":  time.Since(start).Seconds(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

// Build Retrieval Endpoint
func (h *HTTPHandlerImpl) HandleGetBuild(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    buildID := chi.URLParam(r, "id")
    
    if buildID == "" {
        h.handleError(ctx, w, "Build ID is required", 
            fmt.Errorf("build ID is required"), 
            http.StatusBadRequest)
        return
    }
    
    // Start tracing
    ctx, span := h.obs.StartSpanWithAttributes(ctx, "get_build", map[string]string{
        "build_id": buildID,
    })
    defer span.End()
    
    start := time.Now()
    
    // Get job manager
    jobManager := h.container.GetJobManager()
    if jobManager == nil {
        h.handleError(ctx, w, "Job manager not available",
            fmt.Errorf("job manager not available"),
            http.StatusServiceUnavailable)
        return
    }
    
    // Get build details
    build, err := h.getBuildDetails(ctx, jobManager, buildID)
    if err != nil {
        if errors.IsNotFoundError(err) {
            h.handleError(ctx, w, "Build not found", err, http.StatusNotFound)
        } else {
            h.handleError(ctx, w, "Failed to get build", err, http.StatusInternalServerError)
        }
        return
    }
    
    // Create response
    response := map[string]interface{}{
        "build":     build,
        "success":   true,
        "timestamp": time.Now().UTC(),
        "duration":  time.Since(start).Seconds(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

// Build Cancellation Endpoint
func (h *HTTPHandlerImpl) HandleCancelBuild(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    buildID := chi.URLParam(r, "id")
    
    if buildID == "" {
        h.handleError(ctx, w, "Build ID is required",
            fmt.Errorf("build ID is required"),
            http.StatusBadRequest)
        return
    }
    
    // Start tracing
    ctx, span := h.obs.StartSpanWithAttributes(ctx, "cancel_build", map[string]string{
        "build_id": buildID,
    })
    defer span.End()
    
    start := time.Now()
    
    // Get job manager
    jobManager := h.container.GetJobManager()
    if jobManager == nil {
        h.handleError(ctx, w, "Job manager not available",
            fmt.Errorf("job manager not available"),
            http.StatusServiceUnavailable)
        return
    }
    
    // Cancel build
    err := h.cancelBuild(ctx, jobManager, buildID)
    if err != nil {
        if errors.IsNotFoundError(err) {
            h.handleError(ctx, w, "Build not found", err, http.StatusNotFound)
        } else if err.Error() == "invalid state" {
            h.handleError(ctx, w, "Build cannot be cancelled", err, http.StatusConflict)
        } else {
            h.handleError(ctx, w, "Failed to cancel build", err, http.StatusInternalServerError)
        }
        return
    }
    
    // Create response
    response := map[string]interface{}{
        "build_id":  buildID,
        "status":    "cancelled",
        "message":   "Build cancelled successfully",
        "success":   true,
        "timestamp": time.Now().UTC(),
        "duration":  time.Since(start).Seconds(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

// Async Job Stats Endpoint
func (h *HTTPHandlerImpl) HandleAsyncJobStats(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Get async job creator
    asyncJobCreator := h.container.GetAsyncJobCreator()
    if asyncJobCreator == nil {
        h.handleError(ctx, w, "Async job creator not available",
            fmt.Errorf("async job creator not available"),
            http.StatusServiceUnavailable)
        return
    }
    
    // Get statistics
    stats := asyncJobCreator.GetStats()
    
    // Add timestamp
    stats["timestamp"] = time.Now().UTC()
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(stats)
}
```

---

## 📊 API Specification

### List Builds

**Request**:
```http
GET /builds?third_party_id=customer-123&status=running&limit=50
Authorization: Bearer {token}
```

**Response**:
```json
{
  "builds": [
    {
      "id": "build-customer-123-parser-abc-123def",
      "name": "build-customer-123-parser-abc-123def",
      "third_party_id": "customer-123",
      "parser_id": "parser-abc",
      "status": "running",
      "created_at": "2025-10-29T12:00:00Z",
      "started_at": "2025-10-29T12:00:05Z",
      "image_uri": "ecr.../knative-lambdas:customer-123-parser-abc-12345678"
    }
  ],
  "count": 1,
  "limit": 50,
  "success": true,
  "timestamp": "2025-10-29T12:34:56Z",
  "duration": 0.123
}
```

### Get Build

**Request**:
```http
GET /builds/build-customer-123-parser-abc-123def
Authorization: Bearer {token}
```

**Response**:
```json
{
  "build": {
    "id": "build-customer-123-parser-abc-123def",
    "name": "build-customer-123-parser-abc-123def",
    "third_party_id": "customer-123",
    "parser_id": "parser-abc",
    "status": "succeeded",
    "created_at": "2025-10-29T12:00:00Z",
    "started_at": "2025-10-29T12:00:05Z",
    "completed_at": "2025-10-29T12:03:42Z",
    "duration": 217.5,
    "image_uri": "ecr.../knative-lambdas:customer-123-parser-abc-12345678",
    "logs_preview": "Step 1/5 : FROM node:20-alpine\n..."
  },
  "success": true,
  "timestamp": "2025-10-29T12:34:56Z",
  "duration": 0.045
}
```

### Cancel Build

**Request**:
```http
POST /builds/build-customer-123-parser-abc-123def/cancel
Authorization: Bearer {token}
```

**Response**:
```json
{
  "build_id": "build-customer-123-parser-abc-123def",
  "status": "cancelled",
  "message": "Build cancelled successfully",
  "success": true,
  "timestamp": "2025-10-29T12:34:56Z",
  "duration": 0.234
}
```

### Async Job Stats

**Request**:
```http
GET /async-jobs/stats
```

**Response**:
```json
{
  "worker_count": 5,
  "queue_size": 23,
  "max_queue_size": 100,
  "results_count": 150,
  "is_shutdown": false,
  "timestamp": "2025-10-29T12:34:56Z"
}
```

---

## 🧪 Testing Scenarios

### 1. List All Builds
```bash
curl -X GET "http://localhost:8080/builds"
```

### 2. Filter by Third Party
```bash
curl -X GET "http://localhost:8080/builds?third_party_id=customer-123"
```

### 3. Get Specific Build
```bash
curl -X GET "http://localhost:8080/builds/build-customer-123-parser-abc-123def"
```

### 4. Cancel Running Build
```bash
curl -X POST "http://localhost:8080/builds/build-customer-123-parser-abc-123def/cancel"
```

### 5. Check Worker Stats
```bash
curl -X GET "http://localhost:8080/async-jobs/stats"
```

---

## 📈 Performance Requirements

- **List Builds**: < 500ms for 100 builds
- **Get Build**: < 100ms
- **Cancel Build**: < 500ms
- **Stats Endpoint**: < 50ms

---

## 🔍 Monitoring & Alerts

### Metrics
```promql
# API request rate
rate(api_requests_total{endpoint="/builds"}[5m])

# API latency
histogram_quantile(0.95,
  rate(api_request_duration_seconds_bucket{endpoint="/builds"}[5m])
)
```

---

## 🏗️ Code References

**Main Files**:
- `internal/handler/http_handler.go` - API endpoints
- `internal/handler/job_manager.go` - Build operations

---

**Last Updated**: October 29, 2025  
**Owner**: Backend Team  
**Status**: Production Ready

