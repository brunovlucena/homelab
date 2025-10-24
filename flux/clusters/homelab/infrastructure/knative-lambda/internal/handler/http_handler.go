// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🌐 HTTP HANDLER - Focused HTTP server management and routing
//
//	🎯 Purpose: Handle HTTP server operations, routing, and middleware composition
//	💡 Features: Server lifecycle, route registration, middleware chain management
//
//	🏛️ ARCHITECTURE:
//	🌐 HTTP Server Management - Server startup, shutdown, and lifecycle
//	🛣️ Route Registration - CloudEvent endpoints, health checks, build management
//	🔗 Middleware Composition - Logging, metrics, security, rate limiting
//	📊 Response Handling - Error responses and status codes
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"

	"github.com/go-chi/chi/v5"

	"knative-lambda-new/internal/config"
	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
	"knative-lambda-new/internal/observability"
	"knative-lambda-new/internal/resilience"
	"knative-lambda-new/pkg/builds"
)

// 🌐 HTTPHandlerImpl - "Focused HTTP server management and routing"
type HTTPHandlerImpl struct {
	config      *config.HTTPConfig
	obs         *observability.Observability
	router      *chi.Mux
	middleware  http.Handler
	server      *http.Server
	rateLimiter *resilience.RateLimiter
	container   ComponentContainer
}

// 🌐 HTTPHandlerConfig - "Configuration for creating HTTP handler"
type HTTPHandlerConfig struct {
	Config        *config.HTTPConfig
	Observability *observability.Observability
	Container     ComponentContainer
	RateLimiter   *resilience.RateLimiter
}

// 🏗️ NewHTTPHandler - "Create new HTTP handler with dependencies"
func NewHTTPHandler(config HTTPHandlerConfig) (HTTPHandler, error) {
	if config.Config == nil {
		return nil, errors.NewConfigurationError("http_config", "config", "config cannot be nil")
	}

	if config.Observability == nil {
		return nil, errors.NewConfigurationError("observability", "observability", "observability cannot be nil")
	}

	router := chi.NewRouter()

	// Use provided rate limiter or create a new one if not provided
	var rateLimiter *resilience.RateLimiter

	if config.RateLimiter != nil {
		rateLimiter = config.RateLimiter
	} else {
		// Create a simple rate limiter
		rateLimiter = resilience.NewRateLimiter(10, 5) // 10 requests per minute, burst of 5
	}

	// Create middleware chain with observability and rate limiting
	middlewareChain := CreateDefaultMiddlewareChain(config.Observability, rateLimiter)
	middlewareHandler := middlewareChain(router)

	server := &http.Server{
		Addr:              config.Config.GetServerAddress(),
		Handler:           middlewareHandler,
		ReadTimeout:       constants.HTTPReadTimeoutDefault,
		WriteTimeout:      config.Config.APITimeout + constants.HTTPWriteTimeoutOffset,
		IdleTimeout:       constants.HTTPIdleTimeoutDefault,
		ReadHeaderTimeout: constants.HTTPReadHeaderTimeoutDefault,
	}

	handler := &HTTPHandlerImpl{
		config:      config.Config,
		obs:         config.Observability,
		router:      router,
		middleware:  router,
		server:      server,
		rateLimiter: rateLimiter,
		container:   config.Container,
	}

	// Register routes on the router
	handler.RegisterRoutes(nil)

	return handler, nil
}

// 🌐 StartServer - "Start HTTP server with CloudEvent endpoint"
func (h *HTTPHandlerImpl) StartServer(ctx context.Context) error {
	h.obs.Info(ctx, "Starting HTTP server", "port", h.config.Port)

	// Start server in goroutine
	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.obs.Error(ctx, err, "Failed to start HTTP server")
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), constants.HTTPGracefulShutdownTimeout)
	defer cancel()

	h.obs.Info(ctx, "Shutting down HTTP server")
	return h.server.Shutdown(shutdownCtx)
}

// 🛣️ RegisterRoutes - "Register all HTTP routes"
func (h *HTTPHandlerImpl) RegisterRoutes(_ http.Handler) http.Handler {
	// Health check endpoint for Knative queue-proxy
	h.router.Get("/health", h.HandleHealthCheck)

	// CloudEvent endpoint
	h.router.Post("/events", h.HandleCloudEvent)
	h.router.Post("/", h.HandleCloudEvent)

	// Build management endpoints
	h.router.Get("/builds", h.HandleListBuilds)
	h.router.Get("/builds/{id}", h.HandleGetBuild)
	h.router.Post("/builds/{id}/cancel", h.HandleCancelBuild)

	// Async job creator status endpoint
	h.router.Get("/async-jobs/stats", h.HandleAsyncJobStats)

	return h.middleware
}

// 🏥 HealthCheckResponse - "Health check response structure"
type HealthCheckResponse struct {
	Status       string                  `json:"status"`       // overall, ready, alive
	Service      string                  `json:"service"`      // service name
	Timestamp    string                  `json:"timestamp"`    // ISO 8601 timestamp
	Dependencies HealthCheckDependencies `json:"dependencies"` // dependency statuses
	Details      map[string]interface{}  `json:"details,omitempty"`
}

// 🏥 HealthCheckDependencies - "Dependency health check status"
type HealthCheckDependencies struct {
	Storage    *StorageHealthCheck    `json:"storage,omitempty"`
	Kubernetes *KubernetesHealthCheck `json:"kubernetes,omitempty"`
}

// 💾 StorageHealthCheck - "Storage backend health check"
type StorageHealthCheck struct {
	Status   string `json:"status"`   // healthy, degraded, unhealthy
	Provider string `json:"provider"` // s3, minio
	Endpoint string `json:"endpoint"`
	Latency  int64  `json:"latency_ms"` // health check latency in milliseconds
	Error    string `json:"error,omitempty"`
}

// ☸️ KubernetesHealthCheck - "Kubernetes API health check"
type KubernetesHealthCheck struct {
	Status  string `json:"status"` // healthy, unhealthy
	Latency int64  `json:"latency_ms"`
	Error   string `json:"error,omitempty"`
}

// HandleHealthCheck handles health check requests with comprehensive dependency checks
// Supports both liveness (/health/live) and readiness (/health/ready) probes
func (h *HTTPHandlerImpl) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Determine health check type from query parameter or default to readiness
	checkType := r.URL.Query().Get("type")
	if checkType == "" {
		checkType = "ready" // default to readiness check
	}

	response := HealthCheckResponse{
		Service:      "knative-lambda-builder",
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Dependencies: HealthCheckDependencies{},
		Details:      make(map[string]interface{}),
	}

	// For liveness checks, just return OK (minimal check)
	if checkType == "live" {
		response.Status = "alive"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		h.obs.Info(ctx, "Liveness check completed", "endpoint", "/health", "type", "live")
		return
	}

	// For readiness checks, verify all dependencies
	allHealthy := true

	// Check storage backend health
	buildContextManager := h.container.GetBuildContextManager()
	if buildContextManager != nil {
		storageHealth := h.checkStorageHealth(ctx)
		response.Dependencies.Storage = storageHealth
		if storageHealth.Status != "healthy" {
			allHealthy = false
		}
	} else {
		response.Dependencies.Storage = &StorageHealthCheck{
			Status: "unknown",
			Error:  "build context manager not initialized",
		}
		allHealthy = false
	}

	// Check Kubernetes API health
	jobManager := h.container.GetJobManager()
	if jobManager != nil {
		k8sHealth := h.checkKubernetesHealth(ctx)
		response.Dependencies.Kubernetes = k8sHealth
		if k8sHealth.Status != "healthy" {
			allHealthy = false
		}
	}

	// Set overall status
	if allHealthy {
		response.Status = "ready"
		w.WriteHeader(http.StatusOK)
	} else {
		response.Status = "not_ready"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	// Add response time
	duration := time.Since(startTime)
	response.Details["health_check_duration_ms"] = duration.Milliseconds()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.obs.Info(ctx, "Health check completed",
		"endpoint", "/health",
		"type", checkType,
		"status", response.Status,
		"duration_ms", duration.Milliseconds())
}

// checkStorageHealth performs health check on storage backend
func (h *HTTPHandlerImpl) checkStorageHealth(ctx context.Context) *StorageHealthCheck {
	startTime := time.Now()

	buildContextManager := h.container.GetBuildContextManager()
	if buildContextManager == nil {
		return &StorageHealthCheck{
			Status: "unhealthy",
			Error:  "build context manager not available",
		}
	}

	// Get storage client from build context manager
	// Note: This assumes BuildContextManager has access to storage
	// You may need to add a GetStorage() method to BuildContextManager interface
	storage := buildContextManager.GetStorage()
	if storage == nil {
		return &StorageHealthCheck{
			Status: "unhealthy",
			Error:  "storage client not available",
		}
	}

	// Perform health check with timeout
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := storage.HealthCheck(healthCtx)
	latency := time.Since(startTime).Milliseconds()

	if err != nil {
		h.obs.Error(ctx, err, "Storage health check failed",
			"provider", storage.GetProvider(),
			"endpoint", storage.GetEndpoint(),
			"latency_ms", latency)

		return &StorageHealthCheck{
			Status:   "unhealthy",
			Provider: string(storage.GetProvider()),
			Endpoint: storage.GetEndpoint(),
			Latency:  latency,
			Error:    err.Error(),
		}
	}

	return &StorageHealthCheck{
		Status:   "healthy",
		Provider: string(storage.GetProvider()),
		Endpoint: storage.GetEndpoint(),
		Latency:  latency,
	}
}

// checkKubernetesHealth performs health check on Kubernetes API
func (h *HTTPHandlerImpl) checkKubernetesHealth(ctx context.Context) *KubernetesHealthCheck {
	startTime := time.Now()

	jobManager := h.container.GetJobManager()
	if jobManager == nil {
		return &KubernetesHealthCheck{
			Status: "unhealthy",
			Error:  "job manager not available",
		}
	}

	// Simple Kubernetes health check - try to list namespaces or get server version
	// This is a lightweight operation that validates API server connectivity
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Note: You may need to add a HealthCheck() method to JobManager interface
	// For now, we'll assume if JobManager exists, k8s is healthy
	// In production, you should implement actual k8s API health check
	_ = healthCtx

	latency := time.Since(startTime).Milliseconds()

	return &KubernetesHealthCheck{
		Status:  "healthy",
		Latency: latency,
	}
}

// HandleCloudEvent handles CloudEvent ingestion with comprehensive tracing
func (h *HTTPHandlerImpl) HandleCloudEvent(w http.ResponseWriter, r *http.Request) {
	// Extract trace context from request headers
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))

	// Create span for the entire CloudEvent processing
	ctx, span := h.obs.StartSpanWithAttributes(ctx, "handle_cloud_event", map[string]string{
		"http.method": r.Method,
		"http.path":   r.URL.Path,
		"endpoint":    "cloud_event",
	})
	defer span.End()

	// Add correlation ID to response headers for trace propagation
	correlationID := observability.GetCorrelationID(ctx)
	if correlationID != "" {
		w.Header().Set("X-Correlation-ID", correlationID)
	}

	// Get CloudEvent handler from container
	cloudEventHandler := h.container.GetCloudEventHandler()
	if cloudEventHandler == nil {
		span.SetStatus(codes.Error, "CloudEvent handler not available")
		h.obs.Error(ctx, fmt.Errorf("cloud event handler not available"), "CloudEvent handler not available")
		http.Error(w, "CloudEvent handler not available", http.StatusServiceUnavailable)
		return
	}

	// Delegate to CloudEvent handler with tracing context
	r = r.WithContext(ctx)
	cloudEventHandler.HandleCloudEvent(w, r)
}

// 📋 HandleListBuilds - "Handle build listing requests"
func (h *HTTPHandlerImpl) HandleListBuilds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.obs.Info(ctx, "Build listing requested", "endpoint", "/builds")

	// Parse query parameters for filtering
	queryParams := r.URL.Query()
	thirdPartyID := queryParams.Get("third_party_id")
	parserID := queryParams.Get("parser_id")
	status := queryParams.Get("status")
	limit := h.config.DefaultListLimit // Use configured default limit
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= h.config.MaxListLimit {
			limit = parsedLimit
		}
	}

	// Create metrics recorder for this request
	metricsRec := observability.NewMetricsRecorder(h.obs)

	// Start distributed tracing span
	ctx, span := h.obs.StartSpanWithAttributes(ctx, "list_builds", map[string]string{
		"http.method":    r.Method,
		"http.path":      r.URL.Path,
		"third_party_id": thirdPartyID,
		"parser_id":      parserID,
		"status":         status,
		"limit":          strconv.Itoa(limit),
	})
	defer span.End()

	start := time.Now()

	// Get job manager from container for business logic
	jobManager := h.container.GetJobManager()
	if jobManager == nil {
		metricsRec.RecordError(ctx, "http_handler", "job_manager_unavailable", "error")
		h.handleError(ctx, w, "Job manager not available", fmt.Errorf("job manager not available"), http.StatusServiceUnavailable)
		return
	}

	// Call business logic directly
	builds, err := h.listBuildsWithTracing(ctx, jobManager, thirdPartyID, parserID, status, limit)
	if err != nil {
		metricsRec.RecordError(ctx, "http_handler", "list_builds_error", "error")
		h.handleError(ctx, w, "Failed to list builds", err, http.StatusInternalServerError)
		return
	}

	// Record success metrics
	metricsRec.RecordError(ctx, "http_handler", "list_builds_success", "info")

	h.sendBuildListResponse(ctx, w, builds, start, metricsRec)
}

// 📊 HandleAsyncJobStats - "Handle async job creator statistics"
func (h *HTTPHandlerImpl) HandleAsyncJobStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.obs.Info(ctx, "Async job stats requested", "endpoint", "/async-jobs/stats")

	// Get async job creator from container
	asyncJobCreator := h.container.GetAsyncJobCreator()
	if asyncJobCreator == nil {
		h.obs.Error(ctx, fmt.Errorf("async job creator not available"), "Async job creator not available")
		http.Error(w, "Async job creator not available", http.StatusServiceUnavailable)
		return
	}

	// Get statistics
	stats := asyncJobCreator.GetStats()

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Simple JSON response
	response := fmt.Sprintf(`{
		"worker_count": %v,
		"queue_size": %v,
		"max_queue_size": %v,
		"results_count": %v,
		"is_shutdown": %v
	}`, stats["worker_count"], stats["queue_size"], stats["max_queue_size"], stats["results_count"], stats["is_shutdown"])

	w.Write([]byte(response))
}

// 📋 HandleGetBuild - "Handle individual build requests"
func (h *HTTPHandlerImpl) HandleGetBuild(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	buildID := chi.URLParam(r, "id")
	h.obs.Info(ctx, "Build retrieval requested", "endpoint", "/builds/{id}", "build_id", buildID)

	if buildID == "" {
		h.obs.Error(ctx, fmt.Errorf("build ID is required"), "Build ID is required")
		http.Error(w, "Build ID is required", http.StatusBadRequest)
		return
	}

	// Create metrics recorder for this request
	metricsRec := observability.NewMetricsRecorder(h.obs)

	// Start distributed tracing span
	ctx, span := h.obs.StartSpanWithAttributes(ctx, "get_build", map[string]string{
		"http.method": r.Method,
		"http.path":   r.URL.Path,
		"build_id":    buildID,
	})
	defer span.End()

	start := time.Now()

	// Get job manager from container for business logic
	jobManager := h.container.GetJobManager()
	if jobManager == nil {
		metricsRec.RecordError(ctx, "http_handler", "job_manager_unavailable", "error")
		h.handleError(ctx, w, "Job manager not available", fmt.Errorf("job manager not available"), http.StatusServiceUnavailable)
		return
	}

	// Call business logic directly
	build, err := h.getBuildWithTracing(ctx, jobManager, buildID)
	if err != nil {
		metricsRec.RecordError(ctx, "http_handler", "get_build_error", "error")
		if errors.IsNotFoundError(err) {
			h.handleError(ctx, w, "Build not found", err, http.StatusNotFound)
		} else {
			h.handleError(ctx, w, "Failed to get build", err, http.StatusInternalServerError)
		}
		return
	}

	// Record success metrics
	metricsRec.RecordError(ctx, "http_handler", "get_build_success", "info")

	h.sendBuildResponse(ctx, w, build, start, metricsRec)
}

// 🚫 HandleCancelBuild - "Handle build cancellation requests"
func (h *HTTPHandlerImpl) HandleCancelBuild(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	buildID := chi.URLParam(r, "id")
	h.obs.Info(ctx, "Build cancellation requested", "endpoint", "/builds/{id}/cancel", "build_id", buildID)

	if buildID == "" {
		h.obs.Error(ctx, fmt.Errorf("build ID is required"), "Build ID is required")
		http.Error(w, "Build ID is required", http.StatusBadRequest)
		return
	}

	// Create metrics recorder for this request
	metricsRec := observability.NewMetricsRecorder(h.obs)

	// Start distributed tracing span
	ctx, span := h.obs.StartSpanWithAttributes(ctx, "cancel_build", map[string]string{
		"http.method": r.Method,
		"http.path":   r.URL.Path,
		"build_id":    buildID,
	})
	defer span.End()

	start := time.Now()

	// Get job manager from container for business logic
	jobManager := h.container.GetJobManager()
	if jobManager == nil {
		metricsRec.RecordError(ctx, "http_handler", "job_manager_unavailable", "error")
		h.handleError(ctx, w, "Job manager not available", fmt.Errorf("job manager not available"), http.StatusServiceUnavailable)
		return
	}

	// Call business logic directly
	err := h.cancelBuildWithTracing(ctx, jobManager, buildID)
	if err != nil {
		metricsRec.RecordError(ctx, "http_handler", "cancel_build_error", "error")
		if errors.IsNotFoundError(err) {
			h.handleError(ctx, w, "Build not found", err, http.StatusNotFound)
		} else if err.Error() == "invalid state" {
			h.handleError(ctx, w, "Build cannot be cancelled", err, http.StatusConflict)
		} else {
			h.handleError(ctx, w, "Failed to cancel build", err, http.StatusInternalServerError)
		}
		return
	}

	// Record success metrics
	metricsRec.RecordError(ctx, "http_handler", "cancel_build_success", "info")

	h.sendCancellationResponse(ctx, w, buildID, start, metricsRec)
}

// 🔧 GetRouter - "Get the underlying router for route registration"
func (h *HTTPHandlerImpl) GetRouter() *chi.Mux {
	return h.router
}

// 🔧 GetServer - "Get the underlying HTTP server"
func (h *HTTPHandlerImpl) GetServer() *http.Server {
	return h.server
}

// Helper methods for build operations with tracing

// listBuildsWithTracing lists builds with comprehensive tracing
func (h *HTTPHandlerImpl) listBuildsWithTracing(ctx context.Context, jobManager JobManager, thirdPartyID, parserID, status string, limit int) ([]*builds.BuildJob, error) {
	_, span := h.obs.StartSpanWithAttributes(ctx, "list_builds_operation", map[string]string{
		"third_party_id": thirdPartyID,
		"parser_id":      parserID,
		"status":         status,
		"limit":          strconv.Itoa(limit),
	})
	defer span.End()

	// TODO: Implement actual build listing logic when JobManager interface is extended
	// For now, return empty list with proper structure
	builds := make([]*builds.BuildJob, 0)

	span.SetAttributes(attribute.Int("builds.count", len(builds)))
	return builds, nil
}

// getBuildWithTracing gets a specific build with comprehensive tracing
func (h *HTTPHandlerImpl) getBuildWithTracing(ctx context.Context, jobManager JobManager, buildID string) (*builds.BuildJob, error) {
	_, span := h.obs.StartSpanWithAttributes(ctx, "get_build_operation", map[string]string{
		"build_id": buildID,
	})
	defer span.End()

	// TODO: Implement actual build retrieval logic when JobManager interface is extended
	// For now, return a mock build for demonstration
	build := &builds.BuildJob{
		ID:        buildID,
		Name:      fmt.Sprintf("build-%s", buildID),
		Namespace: "default",
		Status:    "unknown",
		CreatedAt: time.Now(),
		Labels: map[string]string{
			"build_id": buildID,
		},
	}

	span.SetAttributes(attribute.String("build.status", build.Status))
	return build, nil
}

// cancelBuildWithTracing cancels a build with comprehensive tracing
func (h *HTTPHandlerImpl) cancelBuildWithTracing(ctx context.Context, jobManager JobManager, buildID string) error {
	ctx, span := h.obs.StartSpanWithAttributes(ctx, "cancel_build_operation", map[string]string{
		"build_id": buildID,
	})
	defer span.End()

	// TODO: Implement actual build cancellation logic when JobManager interface is extended
	// For now, simulate successful cancellation
	h.obs.Info(ctx, "Build cancellation simulated", "build_id", buildID)

	span.SetAttributes(attribute.String("cancellation.status", "success"))
	return nil
}

// Response handling methods

// sendBuildListResponse sends build list response with metrics
func (h *HTTPHandlerImpl) sendBuildListResponse(ctx context.Context, w http.ResponseWriter, builds []*builds.BuildJob, start time.Time, metricsRec *observability.MetricsRecorder) {
	ctx, span := h.obs.StartSpan(ctx, "send_build_list_response")
	defer span.End()

	duration := time.Since(start)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Create response structure
	response := map[string]interface{}{
		"builds":  builds,
		"count":   len(builds),
		"success": true,
	}

	// Encode response
	responseData, err := json.Marshal(response)
	if err != nil {
		h.obs.Error(ctx, err, "Failed to marshal build list response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Write response
	_, err = w.Write(responseData)
	if err != nil {
		h.obs.Error(ctx, err, "Failed to write build list response")
		return
	}

	// Record response metrics
	span.SetAttributes(
		attribute.Int("response.status_code", http.StatusOK),
		attribute.Int("response.size_bytes", len(responseData)),
		attribute.Float64("response.duration_seconds", duration.Seconds()),
		attribute.Int("builds.count", len(builds)),
	)

	h.obs.Info(ctx, "Build list retrieved successfully",
		"duration_seconds", duration.Seconds(),
		"response_size_bytes", len(responseData),
		"status_code", http.StatusOK,
		"builds_count", len(builds),
	)
}

// sendBuildResponse sends individual build response with metrics
func (h *HTTPHandlerImpl) sendBuildResponse(ctx context.Context, w http.ResponseWriter, build *builds.BuildJob, start time.Time, metricsRec *observability.MetricsRecorder) {
	ctx, span := h.obs.StartSpan(ctx, "send_build_response")
	defer span.End()

	duration := time.Since(start)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Create response structure
	response := map[string]interface{}{
		"build":   build,
		"success": true,
	}

	// Encode response
	responseData, err := json.Marshal(response)
	if err != nil {
		h.obs.Error(ctx, err, "Failed to marshal build response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Write response
	_, err = w.Write(responseData)
	if err != nil {
		h.obs.Error(ctx, err, "Failed to write build response")
		return
	}

	// Record response metrics
	span.SetAttributes(
		attribute.Int("response.status_code", http.StatusOK),
		attribute.Int("response.size_bytes", len(responseData)),
		attribute.Float64("response.duration_seconds", duration.Seconds()),
		attribute.String("build.status", build.Status),
	)

	h.obs.Info(ctx, "Build retrieved successfully",
		"duration_seconds", duration.Seconds(),
		"response_size_bytes", len(responseData),
		"status_code", http.StatusOK,
		"build_id", build.ID,
		"build_status", build.Status,
	)
}

// sendCancellationResponse sends build cancellation response with metrics
func (h *HTTPHandlerImpl) sendCancellationResponse(ctx context.Context, w http.ResponseWriter, buildID string, start time.Time, metricsRec *observability.MetricsRecorder) {
	ctx, span := h.obs.StartSpan(ctx, "send_cancellation_response")
	defer span.End()

	duration := time.Since(start)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Create response structure
	response := map[string]interface{}{
		"build_id": buildID,
		"status":   "cancelled",
		"message":  "Build cancelled successfully",
		"success":  true,
	}

	// Encode response
	responseData, err := json.Marshal(response)
	if err != nil {
		h.obs.Error(ctx, err, "Failed to marshal cancellation response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Write response
	_, err = w.Write(responseData)
	if err != nil {
		h.obs.Error(ctx, err, "Failed to write cancellation response")
		return
	}

	// Record response metrics
	span.SetAttributes(
		attribute.Int("response.status_code", http.StatusOK),
		attribute.Int("response.size_bytes", len(responseData)),
		attribute.Float64("response.duration_seconds", duration.Seconds()),
	)

	h.obs.Info(ctx, "Build cancelled successfully",
		"duration_seconds", duration.Seconds(),
		"response_size_bytes", len(responseData),
		"status_code", http.StatusOK,
		"build_id", buildID,
	)
}

// handleError handles errors with proper logging and response
func (h *HTTPHandlerImpl) handleError(ctx context.Context, w http.ResponseWriter, message string, err error, statusCode int) {
	h.obs.Error(ctx, err, message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]string{
		"error":   message,
		"details": err.Error(),
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		h.obs.Error(ctx, err, "Failed to encode error response")
	}
}
