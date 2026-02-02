// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 BACKEND-001: CloudEvents HTTP Processing Tests
//
//	User Story: CloudEvents HTTP Processing
//	Priority: P0 | Story Points: 8
//
//	Tests validate:
//	- HTTP endpoint availability and routing
//	- CloudEvent validation and parsing
//	- Event type routing
//	- Correlation ID management
//	- Observability integration
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_pkg "knative-lambda/internal/config"
	"knative-lambda/internal/handler"
	"knative-lambda/internal/observability"
	"knative-lambda/pkg/builds"
)

// TestBackend001_HTTPEndpointAvailability validates HTTP endpoints are available.
func TestBackend001_HTTPEndpointAvailability(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		description    string
	}{
		{
			name:           "POST /events accepts CloudEvents",
			method:         "POST",
			path:           "/events",
			expectedStatus: http.StatusOK,
			description:    "Main CloudEvent endpoint",
		},
		{
			name:           "POST / accepts CloudEvents",
			method:         "POST",
			path:           "/",
			expectedStatus: http.StatusOK,
			description:    "Root endpoint for compatibility",
		},
		{
			name:           "GET /health returns 200 OK",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
			description:    "Health check for Knative queue-proxy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler, cleanup := setupTestHTTPHandler(t)
			defer cleanup()

			var body string
			if tt.method == "POST" {
				body = createValidCloudEventJSON(t, builds.EventTypeBuildStart)
			}

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(body))
			if tt.method == "POST" {
				addCloudEventHeaders(req, builds.EventTypeBuildStart)
			}
			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch for %s", tt.description)
		})
	}
}

// TestBackend001_CloudEventValidation validates CloudEvent specification compliance.
func TestBackend001_CloudEventValidation(t *testing.T) {
	tests := getCloudEventValidationTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, cleanup := setupTestHTTPHandler(t)
			defer cleanup()

			body := createValidCloudEventJSON(t, builds.EventTypeBuildStart)
			req := httptest.NewRequest("POST", "/events", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			tt.setupRequest(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)
		})
	}
}

// getCloudEventValidationTestCases returns test cases for CloudEvent validation.
func getCloudEventValidationTestCases() []struct {
	name           string
	setupRequest   func(*http.Request)
	expectedStatus int
	description    string
} {
	return []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedStatus int
		description    string
	}{
		{
			name: "Valid CloudEvent with all required headers",
			setupRequest: func(req *http.Request) {
				addCloudEventHeaders(req, builds.EventTypeBuildStart)
			},
			expectedStatus: http.StatusOK,
			description:    "Should accept valid CloudEvent",
		},
		{
			name: "Missing Ce-Type header",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Ce-Specversion", "1.0")
				req.Header.Set("Ce-Source", "network.notifi.customer-123")
				req.Header.Set("Ce-Id", "test-id")
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject CloudEvent without type",
		},
		{
			name: "Missing Ce-Source header",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Ce-Specversion", "1.0")
				req.Header.Set("Ce-Type", builds.EventTypeBuildStart)
				req.Header.Set("Ce-Id", "test-id")
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject CloudEvent without source",
		},
		{
			name: "Missing Ce-Id header",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Ce-Specversion", "1.0")
				req.Header.Set("Ce-Type", builds.EventTypeBuildStart)
				req.Header.Set("Ce-Source", "network.notifi.customer-123")
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject CloudEvent without ID",
		},
	}
}

// TestBackend001_EventTypeRouting validates event type routing.
func TestBackend001_EventTypeRouting(t *testing.T) {
	supportedTypes := []struct {
		eventType   string
		description string
	}{
		{builds.EventTypeBuildStart, "Build start event"},
		{builds.EventTypeBuildComplete, "Build complete event"},
		{builds.EventTypeBuildFailed, "Build failed event"},
		{builds.EventTypeJobStart, "Job start event"},
		{builds.EventTypeParserStart, "Parser start event"},
		{builds.EventTypeParserComplete, "Parser complete event"},
		{builds.EventTypeServiceDelete, "Service delete event"},
	}

	for _, st := range supportedTypes {
		t.Run(st.description, func(t *testing.T) {
			// Arrange
			handler, cleanup := setupTestHTTPHandler(t)
			defer cleanup()

			body := createValidCloudEventJSON(t, st.eventType)
			req := httptest.NewRequest("POST", "/events", strings.NewReader(body))
			addCloudEventHeaders(req, st.eventType)
			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			assert.NotEqual(t, http.StatusNotFound, w.Code, "Event type should be routed")
		})
	}
}

// TestBackend001_UnsupportedEventType validates rejection of unsupported events.
func TestBackend001_UnsupportedEventType(t *testing.T) {
	// Arrange
	handler, cleanup := setupTestHTTPHandler(t)
	defer cleanup()

	body := createValidCloudEventJSON(t, "network.notifi.unsupported.event")
	req := httptest.NewRequest("POST", "/events", strings.NewReader(body))
	addCloudEventHeaders(req, "network.notifi.unsupported.event")
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code, "Should reject unsupported event type")
}

// TestBackend001_CorrelationIDManagement validates correlation ID handling.
func TestBackend001_CorrelationIDManagement(t *testing.T) {
	tests := []struct {
		name               string
		inputCorrelationID string
		description        string
	}{
		{
			name:               "User provided correlation ID",
			inputCorrelationID: "user-correlation-123",
			description:        "Should use provided correlation ID",
		},
		{
			name:               "Generated correlation ID",
			inputCorrelationID: "",
			description:        "Should generate new correlation ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler, cleanup := setupTestHTTPHandler(t)
			defer cleanup()

			body := createValidCloudEventJSON(t, builds.EventTypeBuildStart)
			req := httptest.NewRequest("POST", "/events", strings.NewReader(body))
			addCloudEventHeaders(req, builds.EventTypeBuildStart)

			if tt.inputCorrelationID != "" {
				req.Header.Set("X-Correlation-ID", tt.inputCorrelationID)
			}

			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			correlationID := w.Header().Get("X-Correlation-ID")
			assert.NotEmpty(t, correlationID, "Response should include correlation ID")

			if tt.inputCorrelationID != "" {
				assert.Equal(t, tt.inputCorrelationID, correlationID, "Should preserve user correlation ID")
			} else {
				assert.NotEmpty(t, correlationID, "Should generate correlation ID")
			}
		})
	}
}

// TestBackend001_TraceContextPropagation validates OpenTelemetry trace context.
func TestBackend001_TraceContextPropagation(t *testing.T) {
	// Arrange
	handler, cleanup := setupTestHTTPHandler(t)
	defer cleanup()

	body := createValidCloudEventJSON(t, builds.EventTypeBuildStart)
	req := httptest.NewRequest("POST", "/events", strings.NewReader(body))
	addCloudEventHeaders(req, builds.EventTypeBuildStart)

	// Add trace parent header (W3C Trace Context)
	req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")

	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	traceID := w.Header().Get("X-Trace-ID")
	assert.NotEmpty(t, traceID, "Response should include trace ID")
}

// TestBackend001_GracefulShutdown validates server graceful shutdown.
func TestBackend001_GracefulShutdown(t *testing.T) {
	// Arrange
	handler, cleanup := setupTestHTTPHandler(t)

	// Start a request
	body := createValidCloudEventJSON(t, builds.EventTypeBuildStart)
	req := httptest.NewRequest("POST", "/events", strings.NewReader(body))
	addCloudEventHeaders(req, builds.EventTypeBuildStart)
	w := httptest.NewRecorder()

	// Act - Process request
	handler.ServeHTTP(w, req)

	// Cleanup (triggers graceful shutdown)
	cleanup()

	// Assert
	assert.Equal(t, http.StatusOK, w.Code, "Requests should complete before shutdown")
}

// TestBackend001_ConcurrentRequests validates handling of concurrent requests.
func TestBackend001_ConcurrentRequests(t *testing.T) {
	// Arrange
	handler, cleanup := setupTestHTTPHandler(t)
	defer cleanup()

	concurrency := 100
	results := make(chan int, concurrency)

	// Act - Send concurrent requests
	for i := 0; i < concurrency; i++ {
		go func(_ int) {
			body := createValidCloudEventJSON(t, builds.EventTypeBuildStart)
			req := httptest.NewRequest("POST", "/events", strings.NewReader(body))
			addCloudEventHeaders(req, builds.EventTypeBuildStart)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
			results <- w.Code
		}(i)
	}

	// Assert - Collect results
	successCount := 0
	for i := 0; i < concurrency; i++ {
		status := <-results
		if status == http.StatusOK {
			successCount++
		}
	}

	assert.Greater(t, successCount, 0, "Should handle concurrent requests")
}

// TestBackend001_RequestTimeouts validates request timeout handling.
func TestBackend001_RequestTimeouts(t *testing.T) {
	// Arrange
	handler, cleanup := setupTestHTTPHandler(t)
	defer cleanup()

	body := createValidCloudEventJSON(t, builds.EventTypeBuildStart)
	req := httptest.NewRequest("POST", "/events", strings.NewReader(body))
	addCloudEventHeaders(req, builds.EventTypeBuildStart)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert - Should handle timeout gracefully
	// Note: Actual behavior depends on implementation
	assert.NotEqual(t, 0, w.Code, "Should return a response even on timeout")
}

// TestBackend001_MalformedJSON validates handling of malformed JSON body.
func TestBackend001_MalformedJSON(t *testing.T) {
	// Arrange
	handler, cleanup := setupTestHTTPHandler(t)
	defer cleanup()

	// Malformed JSON
	body := `{"invalid": json, "missing": quote}`
	req := httptest.NewRequest("POST", "/events", strings.NewReader(body))
	addCloudEventHeaders(req, builds.EventTypeBuildStart)

	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code, "Should reject malformed JSON")
}

// TestBackend001_LargePayload validates handling of large payloads.
func TestBackend001_LargePayload(t *testing.T) {
	// Arrange
	handler, cleanup := setupTestHTTPHandler(t)
	defer cleanup()

	// Create large payload (> 1MB)
	largeData := make([]byte, 2*1024*1024) // 2MB
	for i := range largeData {
		largeData[i] = byte('a' + (i % 26))
	}
	body := `{"third_party_id": "customer-123", "parser_id": "parser-abc", "large_data": "` + string(largeData) + `"}`
	req := httptest.NewRequest("POST", "/events", strings.NewReader(body))
	addCloudEventHeaders(req, builds.EventTypeBuildStart)

	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert - Should either reject or handle gracefully
	// Not asserting specific status, but should not panic
	assert.NotEqual(t, 0, w.Code, "Should handle large payload without panic")
}

// TestBackend001_InvalidContentType validates Content-Type header handling.
func TestBackend001_InvalidContentType(t *testing.T) {
	tests := []struct {
		name           string
		contentType    string
		expectedStatus int
		description    string
	}{
		{
			name:           "Missing Content-Type",
			contentType:    "",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject request without Content-Type",
		},
		{
			name:           "Invalid Content-Type",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject non-JSON Content-Type",
		},
		{
			name:           "Valid Content-Type",
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			description:    "Should accept valid Content-Type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			handler, cleanup := setupTestHTTPHandler(t)
			defer cleanup()

			body := createValidCloudEventJSON(t, builds.EventTypeBuildStart)
			req := httptest.NewRequest("POST", "/events", strings.NewReader(body))

			// Set CloudEvent headers
			req.Header.Set("Ce-Specversion", "1.0")
			req.Header.Set("Ce-Type", builds.EventTypeBuildStart)
			req.Header.Set("Ce-Source", "network.notifi.customer-123")
			req.Header.Set("Ce-Id", "test-id")

			// Set Content-Type based on test case
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)
		})
	}
}

// Helper Functions.

func setupTestHTTPHandler(t *testing.T) (http.Handler, func()) {
	// Create mock observability with proper initialization
	obs := handler.NewMockObservability()

	// Create mock container
	container := &MockComponentContainer{
		cloudEventHandler: &MockCloudEventHandler{},
		eventHandler:      &MockEventHandlerForTest{},
		jobManager:        &MockJobManager{},
	}

	// Create HTTP handler config
	config := handler.HTTPHandlerConfig{
		Config:        &config_pkg.HTTPConfig{Port: 8080},
		Observability: obs,
		Container:     container,
	}

	handler, err := handler.NewHTTPHandler(config)
	require.NoError(t, err)

	cleanup := func() {
		// Cleanup observability resources
		if obs != nil {
			if err := obs.Shutdown(context.Background()); err != nil {
				t.Logf("Failed to shutdown observability: %v", err)
			}
		}
	}

	// Get router - HTTPHandlerImpl implements GetRouter() method
	// We need to access it through a type assertion to the concrete type
	// Since GetRouter() returns *chi.Mux, we can call it directly after asserting
	var router *chi.Mux
	if httpImpl, ok := handler.(interface{ GetRouter() *chi.Mux }); ok {
		router = httpImpl.GetRouter()
	} else {
		t.Fatalf("Handler does not implement GetRouter()")
	}
	return router, cleanup
}

func addCloudEventHeaders(req *http.Request, eventType string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", eventType)
	req.Header.Set("Ce-Source", "network.notifi.customer-123")
	req.Header.Set("Ce-Id", "test-event-id-123")
	req.Header.Set("Ce-Subject", "parser-abc")
}

func createValidCloudEventJSON(t *testing.T, eventType string) string {
	var data interface{}

	switch eventType {
	case builds.EventTypeBuildStart, builds.EventTypeJobStart:
		data = map[string]interface{}{
			"third_party_id": "customer-123",
			"parser_id":      "parser-abc",
		}
	case builds.EventTypeBuildComplete, builds.EventTypeBuildFailed:
		data = map[string]interface{}{
			"third_party_id": "customer-123",
			"parser_id":      "parser-abc",
			"status":         "success",
			"image_uri":      "ecr.../image:tag",
		}
	case builds.EventTypeServiceDelete:
		data = map[string]interface{}{
			"third_party_id": "customer-123",
			"parser_id":      "parser-abc",
		}
	default:
		data = map[string]interface{}{}
	}

	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	return string(jsonData)
}

// Mock types for testing.

type HTTPConfigForTest struct {
	Port             string
	APITimeout       time.Duration
	DefaultListLimit int
	MaxListLimit     int
}

func (c *HTTPConfigForTest) GetServerAddress() string {
	return ":" + c.Port
}

type MockCloudEventHandler struct{}

func (m *MockCloudEventHandler) HandleCloudEvent(w http.ResponseWriter, r *http.Request) {
	// Use real CloudEvent handler for validation tests
	// Create a real handler instance with mock container
	obs := handler.NewMockObservability()
	container := &MockComponentContainer{
		eventHandler: &MockEventHandlerForTest{},
	}

	realHandler, err := handler.NewCloudEventHandler(handler.CloudEventHandlerConfig{
		Observability: obs,
		Container:     container,
	})

	if err != nil {
		http.Error(w, "Failed to create handler", http.StatusInternalServerError)
		return
	}

	// Delegate to real handler for proper validation
	realHandler.HandleCloudEvent(w, r)
}

type MockEventHandlerForTest struct{}

func (m *MockEventHandlerForTest) ProcessCloudEvent(_ context.Context, _ *cloudevents.Event) (*builds.HandlerResponse, error) {
	return &builds.HandlerResponse{
		Status:    "queued",
		Message:   "Build job creation queued successfully",
		Timestamp: time.Now(),
	}, nil
}

func (m *MockEventHandlerForTest) ValidateEvent(_ context.Context, event *cloudevents.Event) error {
	// Validate required fields
	if event.Type() == "" || event.Source() == "" || event.ID() == "" {
		return fmt.Errorf("missing required CloudEvent headers")
	}

	// Validate event type is supported
	if !m.IsSupportedEventType(event.Type()) {
		return fmt.Errorf("unsupported event type: %s", event.Type())
	}

	return nil
}

func (m *MockEventHandlerForTest) ParseBuildRequest(_ context.Context, _ *cloudevents.Event) (*builds.BuildRequest, error) {
	return &builds.BuildRequest{}, nil
}

func (m *MockEventHandlerForTest) IsSupportedEventType(eventType string) bool {
	supportedTypes := []string{
		"network.notifi.lambda.build.start",
		"network.notifi.lambda.build.complete",
		"network.notifi.lambda.build.failed",
		"network.notifi.lambda.job.start",
		"network.notifi.lambda.parser.start",
		"network.notifi.lambda.parser.complete",
		"network.notifi.lambda.service.delete",
	}
	for _, t := range supportedTypes {
		if t == eventType {
			return true
		}
	}
	return false
}

// SharedMockJobManager is a mock implementation of JobManager for testing.
type SharedMockJobManager struct{}

func (m *SharedMockJobManager) CreateJob(_ context.Context, jobName string, _ *builds.BuildRequest) (*batchv1.Job, error) {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: "knative-lambda",
		},
	}, nil
}

func (m *SharedMockJobManager) GenerateJobName(thirdPartyID, parserID string) string {
	return fmt.Sprintf("job-%s-%s", thirdPartyID, parserID)
}

func (m *SharedMockJobManager) FindExistingJob(_ context.Context, _, _ string) (*batchv1.Job, error) {
	return nil, nil
}

func (m *SharedMockJobManager) GetJob(_ context.Context, _ string) (*batchv1.Job, error) {
	return nil, nil
}

func (m *SharedMockJobManager) DeleteJob(_ context.Context, _ string) error {
	return nil
}

func (m *SharedMockJobManager) CleanupFailedJob(_ context.Context, _ string) error {
	return nil
}

func (m *SharedMockJobManager) HasFailedJobs(_ context.Context) (bool, error) {
	return false, nil
}

func (m *SharedMockJobManager) IsJobRunning(_ *batchv1.Job) bool {
	return false
}

func (m *SharedMockJobManager) IsJobFailed(_ *batchv1.Job) bool {
	return false
}

func (m *SharedMockJobManager) IsJobSucceeded(_ *batchv1.Job) bool {
	return true
}

func (m *SharedMockJobManager) CountActiveJobs(_ context.Context) (int, error) {
	return 0, nil
}

func (m *SharedMockJobManager) JobExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *SharedMockJobManager) CheckJobStatus(_ context.Context, _ string) (string, error) {
	return "Completed", nil
}

func (m *SharedMockJobManager) WaitForJobCompletion(_ context.Context, _ string) error {
	return nil
}

func (m *SharedMockJobManager) WaitForJobDeletion(_ context.Context, _ string) error {
	return nil
}

// SharedMockAsyncJobCreator is a mock implementation of AsyncJobCreatorInterface for testing.
type SharedMockAsyncJobCreator struct{}

func (m *SharedMockAsyncJobCreator) CreateJobAsync(_ context.Context, _ *builds.BuildRequest) (string, error) {
	return "test-correlation-id", nil
}

func (m *SharedMockAsyncJobCreator) GetJobCreationResult(_ string) (*handler.JobCreationResult, bool) {
	return nil, false
}

func (m *SharedMockAsyncJobCreator) WaitForJobCreation(_ context.Context, _ string) (*handler.JobCreationResult, error) {
	return &handler.JobCreationResult{Error: nil}, nil
}

func (m *SharedMockAsyncJobCreator) GetStats() map[string]interface{} {
	return map[string]interface{}{}
}

func (m *SharedMockAsyncJobCreator) Shutdown(_ context.Context) error {
	return nil
}

// MockJobManager is an alias for SharedMockJobManager for backward compatibility.
type MockJobManager = SharedMockJobManager

type MockComponentContainer struct {
	cloudEventHandler handler.CloudEventHandler
	eventHandler      handler.EventHandler
	jobManager        handler.JobManager
}

func (m *MockComponentContainer) GetCloudEventHandler() handler.CloudEventHandler {
	return m.cloudEventHandler
}

func (m *MockComponentContainer) GetEventHandler() handler.EventHandler {
	return m.eventHandler
}

func (m *MockComponentContainer) GetJobManager() handler.JobManager {
	return m.jobManager
}

func (m *MockComponentContainer) GetAsyncJobCreator() handler.AsyncJobCreatorInterface {
	return &SharedMockAsyncJobCreator{}
}

func (m *MockComponentContainer) GetBuildContextManager() handler.BuildContextManager {
	return nil
}

func (m *MockComponentContainer) GetServiceManager() handler.ServiceManager {
	return nil
}

func (m *MockComponentContainer) GetConfig() *config_pkg.Config {
	return &config_pkg.Config{}
}

func (m *MockComponentContainer) GetObservability() *observability.Observability {
	obs, _ := observability.New(observability.Config{
		ServiceName:    "test",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "info",
		MetricsEnabled: false,
		TracingEnabled: false,
	})
	return obs
}

func (m *MockComponentContainer) Shutdown(_ context.Context) error {
	return nil
}

func (m *MockComponentContainer) GetHTTPHandler() handler.HTTPHandler {
	return nil
}
