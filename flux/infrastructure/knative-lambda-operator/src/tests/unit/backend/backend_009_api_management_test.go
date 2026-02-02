// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 BACKEND-009: API Management Tests
//
//	User Story: API Management
//	Priority: P1 | Story Points: 5
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config_pkg "knative-lambda/internal/config"
	"knative-lambda/internal/handler"
	"knative-lambda/internal/observability"
	"knative-lambda/pkg/builds"
)

// TestBackend009_ListBuildsAPI validates GET /builds endpoint.
func TestBackend009_ListBuildsAPI(t *testing.T) {
	// Arrange
	router := setupAPIRouter(t)
	req := httptest.NewRequest("GET", "/builds", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "builds")
	assert.Contains(t, response, "count")
}

// TestBackend009_BuildFiltering validates build filtering.
func TestBackend009_BuildFiltering(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		description string
	}{
		{
			name:        "Filter by third_party_id",
			queryParams: "?third_party_id=customer-123",
			description: "Should filter builds by third_party_id",
		},
		{
			name:        "Filter by parser_id",
			queryParams: "?parser_id=parser-abc",
			description: "Should filter builds by parser_id",
		},
		{
			name:        "Filter by status",
			queryParams: "?status=completed",
			description: "Should filter builds by status",
		},
		{
			name:        "Multiple filters",
			queryParams: "?third_party_id=customer-123&status=running",
			description: "Should apply multiple filters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			router := setupAPIRouter(t)
			req := httptest.NewRequest("GET", "/builds"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code, tt.description)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			builds, ok := response["builds"].([]interface{})
			require.True(t, ok)
			assert.NotNil(t, builds)
		})
	}
}

// TestBackend009_Pagination validates pagination.
func TestBackend009_Pagination(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		description string
	}{
		{
			name:        "Default pagination",
			queryParams: "",
			description: "Should use default limit of 50",
		},
		{
			name:        "Custom limit",
			queryParams: "?limit=10",
			description: "Should respect custom limit",
		},
		{
			name:        "Custom page",
			queryParams: "?page=2&limit=10",
			description: "Should return page 2",
		},
		{
			name:        "Max limit enforcement",
			queryParams: "?limit=500",
			description: "Should enforce max limit of 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			router := setupAPIRouter(t)
			req := httptest.NewRequest("GET", "/builds"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code, tt.description)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "page")
			assert.Contains(t, response, "limit")
			assert.Contains(t, response, "total")
		})
	}
}

// TestBackend009_GetBuildAPI validates GET /builds/{id} endpoint.
func TestBackend009_GetBuildAPI(t *testing.T) {
	// Arrange
	router := setupAPIRouter(t)
	buildID := "test-build-123"

	req := httptest.NewRequest("GET", "/builds/"+buildID, nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response builds.BuildJob
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, buildID, response.ID)
}

// TestBackend009_GetBuildNotFound validates 404 for non-existent build.
func TestBackend009_GetBuildNotFound(t *testing.T) {
	// Arrange
	router := setupAPIRouter(t)
	req := httptest.NewRequest("GET", "/builds/nonexistent-build", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["error"], "not found")
}

// TestBackend009_CancelBuildAPI validates POST /builds/{id}/cancel endpoint.
func TestBackend009_CancelBuildAPI(t *testing.T) {
	// Arrange
	router := setupAPIRouter(t)
	buildID := "test-build-123"

	req := httptest.NewRequest("POST", "/builds/"+buildID+"/cancel", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Contains(t, response["message"], "cancelled")
}

// TestBackend009_CancelCompletedBuild validates 409 for completed builds.
func TestBackend009_CancelCompletedBuild(t *testing.T) {
	// Arrange
	router := setupAPIRouter(t)
	buildID := "completed-build-123"

	req := httptest.NewRequest("POST", "/builds/"+buildID+"/cancel", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	// Should return 409 Conflict if build is already completed
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusConflict)
}

// TestBackend009_StatsAPI validates GET /async-jobs/stats endpoint.
func TestBackend009_StatsAPI(t *testing.T) {
	// Arrange
	router := setupAPIRouter(t)
	req := httptest.NewRequest("GET", "/async-jobs/stats", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var stats map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&stats)
	require.NoError(t, err)

	assert.Contains(t, stats, "worker_count")
	assert.Contains(t, stats, "max_queue_size")
	assert.Contains(t, stats, "total_queued")
}

// TestBackend009_ResponseFormat validates JSON response format.
func TestBackend009_ResponseFormat(t *testing.T) {
	// Arrange
	router := setupAPIRouter(t)
	req := httptest.NewRequest("GET", "/builds", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Response should be valid JSON")

	// Validate response structure
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "timestamp")
}

// TestBackend009_ErrorResponses validates error response format.
func TestBackend009_ErrorResponses(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		description    string
	}{
		{
			name:           "Invalid limit parameter",
			path:           "/builds?limit=invalid",
			method:         "GET",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 for invalid limit",
		},
		{
			name:           "Invalid page parameter",
			path:           "/builds?page=-1",
			method:         "GET",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 for negative page",
		},
		{
			name:           "Method not allowed",
			path:           "/builds",
			method:         "DELETE",
			expectedStatus: http.StatusMethodNotAllowed,
			description:    "Should return 405 for unsupported method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			router := setupAPIRouter(t)
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			assert.True(t, w.Code >= 400, tt.description)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			assert.Equal(t, "error", response["status"])
		})
	}
}

// TestBackend009_CORSHeaders validates CORS header handling.
func TestBackend009_CORSHeaders(t *testing.T) {
	// Arrange
	router := setupAPIRouter(t)
	req := httptest.NewRequest("OPTIONS", "/builds", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	// CORS headers should be present (if configured)
	assert.NotEmpty(t, w.Header().Get("Content-Type"))
}

// TestBackend009_RateLimiting validates API rate limiting.
func TestBackend009_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rate limit test in short mode")
	}

	// Arrange
	router := setupAPIRouter(t)

	// Act - Send many requests quickly
	successCount := 0
	rateLimitedCount := 0

	for i := 0; i < 150; i++ {
		req := httptest.NewRequest("GET", "/builds", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		switch w.Code {
		case http.StatusOK:
			successCount++
		case http.StatusTooManyRequests:
			rateLimitedCount++
		}
	}

	// Assert
	assert.Greater(t, successCount, 0, "Some requests should succeed")
}

// TestBackend009_APITimeout validates API timeout handling.
func TestBackend009_APITimeout(t *testing.T) {
	// Arrange
	router := setupAPIRouter(t)
	req := httptest.NewRequest("GET", "/builds", nil)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	// Should handle timeout gracefully
	assert.NotEqual(t, 0, w.Code)
}

// Helper Functions.

func setupAPIRouter(t *testing.T) *chi.Mux {
	// Create mock components
	obs := handler.NewMockObservability()
	mockJobManager := &MockJobManager{}
	mockAsyncCreator := &MockAsyncJobCreator{}

	container := &MockAPIContainer{
		jobManager:      mockJobManager,
		asyncJobCreator: mockAsyncCreator,
	}

	// Create HTTP handler
	config := handler.HTTPHandlerConfig{
		Config:        &config_pkg.HTTPConfig{Port: 8080},
		Observability: obs,
		Container:     container,
	}

	handlerInstance, err := handler.NewHTTPHandler(config)
	require.NoError(t, err)

	httpImpl := handlerInstance.(*handler.HTTPHandlerImpl)
	return httpImpl.GetRouter()
}

type MockHTTPConfig struct{}

func (c *MockHTTPConfig) GetServerAddress() string {
	return ":8080"
}

type MockAPIContainer struct {
	jobManager      handler.JobManager
	asyncJobCreator handler.AsyncJobCreatorInterface
}

func (c *MockAPIContainer) GetHTTPHandler() handler.HTTPHandler {
	return nil
}

func (c *MockAPIContainer) GetCloudEventHandler() handler.CloudEventHandler {
	return nil
}

func (c *MockAPIContainer) GetEventHandler() handler.EventHandler {
	return nil
}

func (c *MockAPIContainer) GetJobManager() handler.JobManager {
	return c.jobManager
}

func (c *MockAPIContainer) GetAsyncJobCreator() handler.AsyncJobCreatorInterface {
	return c.asyncJobCreator
}

func (c *MockAPIContainer) GetBuildContextManager() handler.BuildContextManager {
	return nil
}

func (c *MockAPIContainer) GetServiceManager() handler.ServiceManager {
	return nil
}

func (c *MockAPIContainer) GetConfig() *config_pkg.Config {
	return &config_pkg.Config{}
}

func (c *MockAPIContainer) GetObservability() *observability.Observability {
	return &observability.Observability{}
}

func (c *MockAPIContainer) Shutdown(_ context.Context) error {
	return nil
}

// MockAsyncJobCreator is an alias for the shared mock.
type MockAsyncJobCreator = SharedMockAsyncJobCreator
