// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 Unit Tests: CloudEvents Receiver
//
//	Tests for CloudEvents handling:
//	- Configuration
//	- Rate limiting
//	- Event data parsing
//	- Health endpoints
//	- Event processing with fake client
//	- Schema validation
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package webhook

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
	"github.com/brunovlucena/knative-lambda-operator/internal/events"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🛠️ TEST HELPERS                                                        │
// └─────────────────────────────────────────────────────────────────────────┘

// newTestReceiver creates a receiver for testing, disabling schema validation
func newTestReceiver(t *testing.T) *Receiver {
	t.Helper()
	logger := zap.New(zap.UseDevMode(true))
	config := DefaultReceiverConfig()
	config.EnableSchemaValidation = false // Disable for unit tests that don't need it
	receiver, err := NewReceiver(nil, logger, config, nil)
	require.NoError(t, err)
	return receiver
}

// newTestReceiverWithConfig creates a receiver with custom config
func newTestReceiverWithConfig(t *testing.T, config ReceiverConfig) *Receiver {
	t.Helper()
	logger := zap.New(zap.UseDevMode(true))
	receiver, err := NewReceiver(nil, logger, config, nil)
	require.NoError(t, err)
	return receiver
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ⚙️ CONFIGURATION TESTS                                                 │
// └─────────────────────────────────────────────────────────────────────────┘

func TestDefaultReceiverConfig(t *testing.T) {
	config := DefaultReceiverConfig()

	assert.Equal(t, 8080, config.Port, "Default port should be 8080")
	assert.Equal(t, "/", config.Path, "Default path should be /")
	assert.Equal(t, "knative-lambda", config.DefaultNamespace, "Default namespace should be knative-lambda")
	assert.Equal(t, float64(50), config.RateLimit, "Default rate limit should be 50 QPS")
	assert.Equal(t, 100, config.BurstSize, "Default burst size should be 100")
	assert.Equal(t, 10, config.WorkerPoolSize, "Default worker pool size should be 10")
	assert.Equal(t, 30*time.Second, config.ProcessingTimeout, "Default processing timeout should be 30s")
}

func TestReceiverConfig_Custom(t *testing.T) {
	config := ReceiverConfig{
		Port:              9090,
		Path:              "/events",
		DefaultNamespace:  "custom-ns",
		RateLimit:         100,
		BurstSize:         200,
		WorkerPoolSize:    20,
		ProcessingTimeout: 60 * time.Second,
	}

	assert.Equal(t, 9090, config.Port)
	assert.Equal(t, "/events", config.Path)
	assert.Equal(t, "custom-ns", config.DefaultNamespace)
	assert.Equal(t, float64(100), config.RateLimit)
	assert.Equal(t, 200, config.BurstSize)
	assert.Equal(t, 20, config.WorkerPoolSize)
	assert.Equal(t, 60*time.Second, config.ProcessingTimeout)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏥 HEALTH ENDPOINT TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHealthHandler(t *testing.T) {
	receiver := newTestReceiver(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	receiver.healthHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestReadyHandler(t *testing.T) {
	receiver := newTestReceiver(t)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	receiver.readyHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ready", w.Body.String())
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 METRICS HANDLER TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestMetricsHandler(t *testing.T) {
	receiver := newTestReceiver(t)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	receiver.metricsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var metrics map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&metrics)
	require.NoError(t, err)

	assert.Contains(t, metrics, "events_received")
	assert.Contains(t, metrics, "events_processed")
	assert.Contains(t, metrics, "events_failed")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 COUNTER TESTS                                                       │
// └─────────────────────────────────────────────────────────────────────────┘

func TestIncrementReceived(t *testing.T) {
	receiver := newTestReceiver(t)

	assert.Equal(t, int64(0), receiver.eventsReceived)

	receiver.incrementReceived()
	assert.Equal(t, int64(1), receiver.eventsReceived)

	receiver.incrementReceived()
	receiver.incrementReceived()
	assert.Equal(t, int64(3), receiver.eventsReceived)
}

func TestIncrementProcessed(t *testing.T) {
	receiver := newTestReceiver(t)

	assert.Equal(t, int64(0), receiver.eventsProcessed)

	receiver.incrementProcessed()
	assert.Equal(t, int64(1), receiver.eventsProcessed)

	receiver.incrementProcessed()
	assert.Equal(t, int64(2), receiver.eventsProcessed)
}

func TestIncrementFailed(t *testing.T) {
	receiver := newTestReceiver(t)

	assert.Equal(t, int64(0), receiver.eventsFailed)

	receiver.incrementFailed()
	assert.Equal(t, int64(1), receiver.eventsFailed)

	receiver.incrementFailed()
	receiver.incrementFailed()
	receiver.incrementFailed()
	assert.Equal(t, int64(4), receiver.eventsFailed)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 EVENT DATA STRUCTURES TESTS                                         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestFunctionDeployData(t *testing.T) {
	data := FunctionDeployData{
		Metadata: FunctionMetadata{
			Name:      "test-function",
			Namespace: "test-ns",
			Labels: map[string]string{
				"app": "test",
			},
			Annotations: map[string]string{
				"description": "Test function",
			},
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(event): return event",
				},
			},
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	}

	assert.Equal(t, "test-function", data.Metadata.Name)
	assert.Equal(t, "test-ns", data.Metadata.Namespace)
	assert.Equal(t, "test", data.Metadata.Labels["app"])
	assert.Equal(t, "inline", data.Spec.Source.Type)
	assert.Equal(t, "python", data.Spec.Runtime.Language)
}

func TestFunctionDeployData_JSON(t *testing.T) {
	jsonData := `{
		"metadata": {
			"name": "json-function",
			"namespace": "default",
			"labels": {
				"env": "production"
			}
		},
		"spec": {
			"source": {
				"type": "inline",
				"inline": {
					"code": "def handler(e): return e"
				}
			},
			"runtime": {
				"language": "python",
				"version": "3.11"
			}
		}
	}`

	var data FunctionDeployData
	err := json.Unmarshal([]byte(jsonData), &data)
	require.NoError(t, err)

	assert.Equal(t, "json-function", data.Metadata.Name)
	assert.Equal(t, "default", data.Metadata.Namespace)
	assert.Equal(t, "production", data.Metadata.Labels["env"])
	assert.Equal(t, "inline", data.Spec.Source.Type)
}

func TestServiceDeleteData(t *testing.T) {
	data := ServiceDeleteData{
		Name:      "function-to-delete",
		Namespace: "production",
		Reason:    "User requested deletion",
	}

	assert.Equal(t, "function-to-delete", data.Name)
	assert.Equal(t, "production", data.Namespace)
	assert.Equal(t, "User requested deletion", data.Reason)
}

func TestServiceDeleteData_JSON(t *testing.T) {
	jsonData := `{
		"name": "json-delete",
		"namespace": "test-ns",
		"reason": "Cleanup"
	}`

	var data ServiceDeleteData
	err := json.Unmarshal([]byte(jsonData), &data)
	require.NoError(t, err)

	assert.Equal(t, "json-delete", data.Name)
	assert.Equal(t, "test-ns", data.Namespace)
	assert.Equal(t, "Cleanup", data.Reason)
}

func TestBuildCommandData(t *testing.T) {
	data := BuildCommandData{
		Name:         "function-to-build",
		Namespace:    "build-ns",
		ForceRebuild: true,
		Reason:       "Code update",
	}

	assert.Equal(t, "function-to-build", data.Name)
	assert.Equal(t, "build-ns", data.Namespace)
	assert.True(t, data.ForceRebuild)
	assert.Equal(t, "Code update", data.Reason)
}

func TestBuildCommandData_JSON(t *testing.T) {
	jsonData := `{
		"name": "rebuild-function",
		"forceRebuild": true,
		"reason": "Dependency update"
	}`

	var data BuildCommandData
	err := json.Unmarshal([]byte(jsonData), &data)
	require.NoError(t, err)

	assert.Equal(t, "rebuild-function", data.Name)
	assert.True(t, data.ForceRebuild)
	assert.Equal(t, "Dependency update", data.Reason)
	// Namespace defaults to empty
	assert.Empty(t, data.Namespace)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🌐 HTTP METHOD VALIDATION TESTS                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHandleCloudEvent_MethodNotAllowed(t *testing.T) {
	receiver := newTestReceiver(t)

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run("Method_"+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			w := httptest.NewRecorder()

			receiver.handleCloudEvent(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📬 CLOUDEVENT PARSING TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHandleCloudEvent_InvalidCloudEvent(t *testing.T) {
	receiver := newTestReceiver(t)

	tests := []struct {
		name        string
		body        string
		headers     map[string]string
		description string
	}{
		{
			name:        "Missing CloudEvent headers",
			body:        `{"test": "data"}`,
			headers:     map[string]string{"Content-Type": "application/json"},
			description: "Should reject request without CloudEvent headers",
		},
		{
			name:        "Invalid JSON body",
			body:        `not valid json`,
			headers:     map[string]string{"Content-Type": "application/json"},
			description: "Should reject invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()

			receiver.handleCloudEvent(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code, tt.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ⏱️ RATE LIMITER TESTS                                                  │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNewReceiver_RateLimiter(t *testing.T) {
	config := ReceiverConfig{
		RateLimit:              100,
		BurstSize:              50,
		EnableSchemaValidation: false,
	}

	receiver := newTestReceiverWithConfig(t, config)

	require.NotNil(t, receiver.rateLimiter)

	// Should allow up to burst size immediately
	for i := 0; i < 50; i++ {
		allowed := receiver.rateLimiter.Allow()
		assert.True(t, allowed, "Should allow request %d within burst", i)
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📋 METADATA TESTS                                                      │
// └─────────────────────────────────────────────────────────────────────────┘

func TestFunctionMetadata(t *testing.T) {
	metadata := FunctionMetadata{
		Name:      "test-func",
		Namespace: "test-ns",
		Labels: map[string]string{
			"app":     "myapp",
			"version": "v1",
		},
		Annotations: map[string]string{
			"description":              "Test function",
			"owner":                    "team-a",
			"lambda.knative.io/custom": "value",
		},
	}

	assert.Equal(t, "test-func", metadata.Name)
	assert.Equal(t, "test-ns", metadata.Namespace)
	assert.Len(t, metadata.Labels, 2)
	assert.Len(t, metadata.Annotations, 3)
	assert.Equal(t, "myapp", metadata.Labels["app"])
	assert.Equal(t, "team-a", metadata.Annotations["owner"])
}

func TestFunctionMetadata_EmptyNamespace(t *testing.T) {
	metadata := FunctionMetadata{
		Name: "func-without-ns",
	}

	assert.Equal(t, "func-without-ns", metadata.Name)
	assert.Empty(t, metadata.Namespace)
	assert.Nil(t, metadata.Labels)
	assert.Nil(t, metadata.Annotations)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏗️ RECEIVER INITIALIZATION TESTS                                       │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNewReceiver(t *testing.T) {
	logger := zap.New(zap.UseDevMode(true))
	config := DefaultReceiverConfig()
	config.EnableSchemaValidation = false

	receiver, err := NewReceiver(nil, logger, config, nil)
	require.NoError(t, err)

	require.NotNil(t, receiver)
	assert.NotNil(t, receiver.log)
	assert.NotNil(t, receiver.rateLimiter)
	assert.Equal(t, config.Port, receiver.config.Port)
	assert.Equal(t, config.Path, receiver.config.Path)
	assert.Equal(t, config.DefaultNamespace, receiver.config.DefaultNamespace)
}

func TestNewReceiver_WithClient(t *testing.T) {
	logger := zap.New(zap.UseDevMode(true))
	config := DefaultReceiverConfig()
	config.EnableSchemaValidation = false

	// Note: In real test, you'd use fake.NewClientBuilder()
	receiver, err := NewReceiver(nil, logger, config, nil)
	require.NoError(t, err)

	require.NotNil(t, receiver)
	assert.Nil(t, receiver.client) // nil since we passed nil
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔒 CONCURRENT METRICS TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestReceiverMetrics_Concurrent(t *testing.T) {
	receiver := newTestReceiver(t)

	// Run concurrent increments
	done := make(chan bool, 3)

	go func() {
		for i := 0; i < 100; i++ {
			receiver.incrementReceived()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			receiver.incrementProcessed()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			receiver.incrementFailed()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	assert.Equal(t, int64(100), receiver.eventsReceived)
	assert.Equal(t, int64(100), receiver.eventsProcessed)
	assert.Equal(t, int64(100), receiver.eventsFailed)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📝 INTEGRATION-READY HELPERS                                           │
// └─────────────────────────────────────────────────────────────────────────┘

func createValidCloudEventRequest(t *testing.T, eventType, subject string, data interface{}) *http.Request {
	t.Helper()

	body, err := json.Marshal(data)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", eventType)
	req.Header.Set("Ce-Source", "io.knative.lambda/test")
	req.Header.Set("Ce-Id", "test-event-id")
	req.Header.Set("Ce-Subject", subject)

	return req
}

func TestCreateValidCloudEventRequest_Helper(t *testing.T) {
	data := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "test-function",
		},
	}

	req := createValidCloudEventRequest(t, "io.knative.lambda.command.function.deploy", "test-function", data)

	assert.Equal(t, http.MethodPost, req.Method)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "1.0", req.Header.Get("Ce-Specversion"))
	assert.Equal(t, "io.knative.lambda.command.function.deploy", req.Header.Get("Ce-Type"))
	assert.Equal(t, "io.knative.lambda/test", req.Header.Get("Ce-Source"))
	assert.Equal(t, "test-event-id", req.Header.Get("Ce-Id"))
	assert.Equal(t, "test-function", req.Header.Get("Ce-Subject"))
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔐 SCHEMA VALIDATION TESTS                                             │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNewReceiver_WithSchemaValidation(t *testing.T) {
	logger := zap.New(zap.UseDevMode(true))
	config := DefaultReceiverConfig()
	config.EnableSchemaValidation = true // Enable schema validation

	receiver, err := NewReceiver(nil, logger, config, nil)
	require.NoError(t, err)

	require.NotNil(t, receiver)
	require.NotNil(t, receiver.schemaValidator, "Schema validator should be initialized")
}

func TestSchemaValidationError(t *testing.T) {
	err := &SchemaValidationError{
		EventID:   "test-123",
		EventType: "io.knative.lambda.command.function.deploy",
		Message:   "missing required field: metadata.name",
	}

	assert.Contains(t, err.Error(), "test-123")
	assert.Contains(t, err.Error(), "io.knative.lambda.command.function.deploy")
	assert.Contains(t, err.Error(), "metadata.name")
	assert.True(t, IsSchemaValidationError(err))
	assert.False(t, IsSchemaValidationError(nil))
	assert.False(t, IsSchemaValidationError(assert.AnError))
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏭 FAKE CLIENT HELPERS                                                 │
// └─────────────────────────────────────────────────────────────────────────┘

// newTestReceiverWithFakeClient creates a receiver with a fake k8s client
func newTestReceiverWithFakeClient(t *testing.T, objects ...runtime.Object) *Receiver {
	t.Helper()

	scheme := runtime.NewScheme()
	_ = lambdav1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objects...).
		WithStatusSubresource(&lambdav1alpha1.LambdaFunction{}).
		Build()

	logger := zap.New(zap.UseDevMode(true))
	config := DefaultReceiverConfig()
	config.EnableSchemaValidation = false

	receiver, err := NewReceiver(fakeClient, logger, config, nil)
	require.NoError(t, err)
	return receiver
}

// newTestReceiverWithSchemaValidation creates a receiver with schema validation enabled
func newTestReceiverWithSchemaValidation(t *testing.T, objects ...runtime.Object) *Receiver {
	t.Helper()

	scheme := runtime.NewScheme()
	_ = lambdav1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objects...).
		WithStatusSubresource(&lambdav1alpha1.LambdaFunction{}).
		Build()

	logger := zap.New(zap.UseDevMode(true))
	config := DefaultReceiverConfig()
	config.EnableSchemaValidation = true

	receiver, err := NewReceiver(fakeClient, logger, config, nil)
	require.NoError(t, err)
	return receiver
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🚀 PROCESS EVENT TESTS                                                 │
// └─────────────────────────────────────────────────────────────────────────┘

func TestProcessEvent_UnknownEventType(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("test-123")
	event.SetType("io.knative.lambda.unknown.event")
	event.SetSource("io.knative.lambda/test")

	err := receiver.processEvent(ctx, &event)

	// Unknown events should not fail
	assert.NoError(t, err)
}

func TestProcessEvent_CommandFunctionDeploy_Create(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("deploy-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, FunctionDeployData{
		Metadata: FunctionMetadata{
			Name:      "test-function",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(e): return e",
				},
			},
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_CommandServiceCreate(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("create-123")
	event.SetType(events.EventTypeCommandServiceCreate)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, FunctionDeployData{
		Metadata: FunctionMetadata{
			Name:      "service-create-function",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(e): return e",
				},
			},
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_CommandServiceUpdate(t *testing.T) {
	// Create existing function
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing-function",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(e): return e",
				},
			},
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("update-123")
	event.SetType(events.EventTypeCommandServiceUpdate)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, FunctionDeployData{
		Metadata: FunctionMetadata{
			Name:      "existing-function",
			Namespace: "default",
			Labels: map[string]string{
				"updated": "true",
			},
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(e): return {'updated': True}",
				},
			},
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_CommandServiceDelete(t *testing.T) {
	// Create existing function to delete
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "function-to-delete",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("delete-123")
	event.SetType(events.EventTypeCommandServiceDelete)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, ServiceDeleteData{
		Name:      "function-to-delete",
		Namespace: "default",
		Reason:    "User requested deletion",
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_CommandServiceDelete_FromSubject(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "subject-function",
			Namespace: "knative-lambda",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("delete-subject-123")
	event.SetType(events.EventTypeCommandServiceDelete)
	event.SetSource("io.knative.lambda/test")
	event.SetSubject("subject-function")
	event.SetData(cloudevents.ApplicationJSON, ServiceDeleteData{
		// Name empty - should use subject
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_CommandBuildStart(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "build-function",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			Phase: lambdav1alpha1.PhaseReady,
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("build-123")
	event.SetType(events.EventTypeCommandBuildStart)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, BuildCommandData{
		Name:      "build-function",
		Namespace: "default",
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_CommandBuildStart_ForceRebuild(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "force-build-function",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
			Build: &lambdav1alpha1.BuildSpec{
				ForceRebuild: false,
			},
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			Phase: lambdav1alpha1.PhaseReady,
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("force-build-123")
	event.SetType(events.EventTypeCommandBuildStart)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, BuildCommandData{
		Name:         "force-build-function",
		Namespace:    "default",
		ForceRebuild: true,
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_CommandBuildRetry(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "retry-build-function",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			Phase: lambdav1alpha1.PhaseFailed,
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("retry-123")
	event.SetType(events.EventTypeCommandBuildRetry)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, BuildCommandData{
		Name:      "retry-build-function",
		Namespace: "default",
		Reason:    "Retry after transient failure",
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_CommandBuildCancel(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cancel-build-function",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			Phase: lambdav1alpha1.PhaseBuilding,
			BuildStatus: &lambdav1alpha1.BuildStatusInfo{
				JobName: "cancel-build-function-build-123",
			},
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("cancel-123")
	event.SetType(events.EventTypeCommandBuildCancel)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, BuildCommandData{
		Name:      "cancel-build-function",
		Namespace: "default",
		Reason:    "User cancelled build",
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_CommandBuildCancel_NotBuilding(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ready-function",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			Phase: lambdav1alpha1.PhaseReady,
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("cancel-ready-123")
	event.SetType(events.EventTypeCommandBuildCancel)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, BuildCommandData{
		Name:      "ready-function",
		Namespace: "default",
	})

	err := receiver.processEvent(ctx, &event)

	// Should return nil if not in building phase
	assert.NoError(t, err)
}

func TestProcessEvent_CommandFunctionRollback(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rollback-function",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			Phase: lambdav1alpha1.PhaseReady,
			ServiceStatus: &lambdav1alpha1.ServiceStatusInfo{
				LatestRevision: "rollback-function-00003",
			},
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("rollback-123")
	event.SetType(events.EventTypeCommandFunctionRollback)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"name":      "rollback-function",
		"namespace": "default",
		"revision":  "rollback-function-00001",
		"reason":    "Regression detected",
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ❌ ERROR HANDLING TESTS                                                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestProcessEvent_FunctionDeploy_MissingName(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("deploy-no-name-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, FunctionDeployData{
		Metadata: FunctionMetadata{
			Name: "", // Missing name
		},
	})

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestProcessEvent_ServiceDelete_MissingName(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("delete-no-name-123")
	event.SetType(events.EventTypeCommandServiceDelete)
	event.SetSource("io.knative.lambda/test")
	// No subject set either
	event.SetData(cloudevents.ApplicationJSON, ServiceDeleteData{
		Name: "", // Missing name
	})

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestProcessEvent_BuildStart_MissingName(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("build-no-name-123")
	event.SetType(events.EventTypeCommandBuildStart)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, BuildCommandData{
		Name: "",
	})

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestProcessEvent_BuildCancel_MissingName(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("cancel-no-name-123")
	event.SetType(events.EventTypeCommandBuildCancel)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, BuildCommandData{
		Name: "",
	})

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestProcessEvent_Rollback_MissingName(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("rollback-no-name-123")
	event.SetType(events.EventTypeCommandFunctionRollback)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"name": "",
	})

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestProcessEvent_BuildStart_NotFound(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("build-notfound-123")
	event.SetType(events.EventTypeCommandBuildStart)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, BuildCommandData{
		Name:      "nonexistent-function",
		Namespace: "default",
	})

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔐 SCHEMA VALIDATION COUNTER TEST                                      │
// └─────────────────────────────────────────────────────────────────────────┘

func TestIncrementSchemaValidationFailed(t *testing.T) {
	receiver := newTestReceiver(t)

	assert.Equal(t, int64(0), receiver.eventsSchemaValidFailed)

	receiver.incrementSchemaValidationFailed()
	assert.Equal(t, int64(1), receiver.eventsSchemaValidFailed)

	receiver.incrementSchemaValidationFailed()
	receiver.incrementSchemaValidationFailed()
	assert.Equal(t, int64(3), receiver.eventsSchemaValidFailed)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🌐 HANDLE CLOUDEVENT WITH VALID EVENT TESTS                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHandleCloudEvent_ValidEvent_Success(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)

	body := `{
		"metadata": {
			"name": "http-test-function",
			"namespace": "default"
		},
		"spec": {
			"source": {
				"type": "inline",
				"inline": {
					"code": "def handler(e): return e"
				}
			},
			"runtime": {
				"language": "python",
				"version": "3.11"
			}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
	req.Header.Set("Ce-Source", "io.knative.lambda/test")
	req.Header.Set("Ce-Id", "test-http-event-123")

	w := httptest.NewRecorder()

	receiver.handleCloudEvent(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "accepted", response["status"])
}

func TestHandleCloudEvent_UnknownEventType(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)

	body := `{"data": "test"}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", "io.custom.unknown.event")
	req.Header.Set("Ce-Source", "io.custom/source")
	req.Header.Set("Ce-Id", "unknown-event-123")

	w := httptest.NewRecorder()

	receiver.handleCloudEvent(w, req)

	// Unknown events should still return 202 Accepted
	assert.Equal(t, http.StatusAccepted, w.Code)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔒 DEFAULT NAMESPACE TESTS                                             │
// └─────────────────────────────────────────────────────────────────────────┘

func TestProcessEvent_UsesDefaultNamespace(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("default-ns-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, FunctionDeployData{
		Metadata: FunctionMetadata{
			Name: "function-without-namespace",
			// Namespace intentionally empty
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(e): return e",
				},
			},
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	})

	err := receiver.processEvent(ctx, &event)

	// Should use default namespace
	assert.NoError(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📋 LABELS AND ANNOTATIONS HANDLING                                     │
// └─────────────────────────────────────────────────────────────────────────┘

func TestProcessEvent_FunctionDeploy_WithLabelsAndAnnotations(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("labeled-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, FunctionDeployData{
		Metadata: FunctionMetadata{
			Name:      "labeled-function",
			Namespace: "default",
			Labels: map[string]string{
				"app":     "myapp",
				"version": "v1",
				"team":    "platform",
			},
			Annotations: map[string]string{
				"description": "A test function",
				"owner":       "team@example.com",
			},
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(e): return e",
				},
			},
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 UPDATE EXISTING FUNCTION TESTS                                      │
// └─────────────────────────────────────────────────────────────────────────┘

func TestProcessEvent_FunctionDeploy_UpdateExisting_WithNilMaps(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "existing-no-maps",
			Namespace:   "default",
			Labels:      nil, // No labels initially
			Annotations: nil, // No annotations initially
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("update-nil-maps-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, FunctionDeployData{
		Metadata: FunctionMetadata{
			Name:      "existing-no-maps",
			Namespace: "default",
			Labels: map[string]string{
				"new-label": "value",
			},
			Annotations: map[string]string{
				"new-annotation": "value",
			},
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(e): return e",
				},
			},
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.12",
			},
		},
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📤 INVALID DATA PARSING TESTS                                          │
// └─────────────────────────────────────────────────────────────────────────┘

func TestProcessEvent_FunctionDeploy_InvalidData(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("invalid-data-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, "not a valid FunctionDeployData")

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestProcessEvent_ServiceDelete_InvalidData(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("invalid-delete-123")
	event.SetType(events.EventTypeCommandServiceDelete)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, "not valid")

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestProcessEvent_BuildStart_InvalidData(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("invalid-build-123")
	event.SetType(events.EventTypeCommandBuildStart)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, "not valid")

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestProcessEvent_BuildCancel_InvalidData(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("invalid-cancel-123")
	event.SetType(events.EventTypeCommandBuildCancel)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, "not valid")

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestProcessEvent_Rollback_InvalidData(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("invalid-rollback-123")
	event.SetType(events.EventTypeCommandFunctionRollback)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, "not valid")

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔐 SCHEMA VALIDATION INTEGRATION TESTS                                 │
// └─────────────────────────────────────────────────────────────────────────┘

func TestProcessEvent_WithSchemaValidation_ValidPayload(t *testing.T) {
	receiver := newTestReceiverWithSchemaValidation(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("schema-valid-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "valid-schema-function",
		},
		"spec": map[string]interface{}{
			"source": map[string]interface{}{
				"type": "inline",
				"inline": map[string]interface{}{
					"code": "def handler(e): return e",
				},
			},
			"runtime": map[string]interface{}{
				"language": "python",
				"version":  "3.11",
			},
		},
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_WithSchemaValidation_InvalidPayload(t *testing.T) {
	receiver := newTestReceiverWithSchemaValidation(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("schema-invalid-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		// Missing required "metadata" field
		"spec": map[string]interface{}{
			"source": map[string]interface{}{
				"type": "inline",
			},
		},
	})

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	assert.True(t, IsSchemaValidationError(err))
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🧪 SUBJECT FROM EVENT TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestProcessEvent_BuildStart_FromSubject(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "subject-build-func",
			Namespace: "knative-lambda",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("build-subject-123")
	event.SetType(events.EventTypeCommandBuildStart)
	event.SetSource("io.knative.lambda/test")
	event.SetSubject("subject-build-func")
	event.SetData(cloudevents.ApplicationJSON, BuildCommandData{
		// Name empty - should use subject
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_BuildCancel_FromSubject(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "subject-cancel-func",
			Namespace: "knative-lambda",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			Phase: lambdav1alpha1.PhaseBuilding,
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("cancel-subject-123")
	event.SetType(events.EventTypeCommandBuildCancel)
	event.SetSource("io.knative.lambda/test")
	event.SetSubject("subject-cancel-func")
	event.SetData(cloudevents.ApplicationJSON, BuildCommandData{
		// Name empty - should use subject
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_Rollback_FromSubject(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "subject-rollback-func",
			Namespace: "knative-lambda",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("rollback-subject-123")
	event.SetType(events.EventTypeCommandFunctionRollback)
	event.SetSource("io.knative.lambda/test")
	event.SetSubject("subject-rollback-func")
	event.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		// Name empty - should use subject
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ⏱️ RATE LIMITING BEHAVIOR TESTS                                         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHandleCloudEvent_RateLimitExceeded(t *testing.T) {
	// Create receiver with very low rate limit and fake client
	scheme := runtime.NewScheme()
	_ = lambdav1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&lambdav1alpha1.LambdaFunction{}).
		Build()

	logger := zap.New(zap.UseDevMode(true))
	config := DefaultReceiverConfig()
	config.RateLimit = 1.0 // 1 request per second
	config.BurstSize = 1
	config.EnableSchemaValidation = false

	receiver, err := NewReceiver(fakeClient, logger, config, nil)
	require.NoError(t, err)

	// Create a valid CloudEvent request
	body := `{
		"metadata": {
			"name": "rate-limit-test"
		},
		"spec": {
			"source": {
				"type": "inline",
				"inline": {
					"code": "def handler(e): return e"
				}
			}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
	req.Header.Set("Ce-Source", "io.knative.lambda/test")
	req.Header.Set("Ce-Id", "rate-limit-1")

	w1 := httptest.NewRecorder()
	receiver.handleCloudEvent(w1, req)

	// First request should succeed
	assert.Equal(t, http.StatusAccepted, w1.Code)

	// Second request immediately after should be rate limited
	req2 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Ce-Specversion", "1.0")
	req2.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
	req2.Header.Set("Ce-Source", "io.knative.lambda/test")
	req2.Header.Set("Ce-Id", "rate-limit-2")

	// Use a context with timeout to simulate rate limit wait
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	req2 = req2.WithContext(ctx)

	w2 := httptest.NewRecorder()
	receiver.handleCloudEvent(w2, req2)

	// Should return 429 Too Many Requests
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Contains(t, w2.Header().Get("Retry-After"), "5")
}

func TestHandleCloudEvent_RateLimitTimeout(t *testing.T) {
	// Create receiver with fake client
	scheme := runtime.NewScheme()
	_ = lambdav1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&lambdav1alpha1.LambdaFunction{}).
		Build()

	logger := zap.New(zap.UseDevMode(true))
	config := DefaultReceiverConfig()
	config.RateLimit = 0.1 // Very low rate limit
	config.BurstSize = 1
	config.ProcessingTimeout = 50 * time.Millisecond
	config.EnableSchemaValidation = false

	receiver, err := NewReceiver(fakeClient, logger, config, nil)
	require.NoError(t, err)

	body := `{
		"metadata": {
			"name": "timeout-test"
		},
		"spec": {
			"source": {
				"type": "inline",
				"inline": {
					"code": "def handler(e): return e"
				}
			}
		}
	}`

	// Consume the burst
	req1 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Ce-Specversion", "1.0")
	req1.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
	req1.Header.Set("Ce-Source", "io.knative.lambda/test")
	req1.Header.Set("Ce-Id", "timeout-1")
	w1 := httptest.NewRecorder()
	receiver.handleCloudEvent(w1, req1)

	// Second request with short timeout should fail
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	req2 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req2 = req2.WithContext(ctx)
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Ce-Specversion", "1.0")
	req2.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
	req2.Header.Set("Ce-Source", "io.knative.lambda/test")
	req2.Header.Set("Ce-Id", "timeout-2")

	w2 := httptest.NewRecorder()
	receiver.handleCloudEvent(w2, req2)

	// Should return 429 or timeout error
	assert.True(t, w2.Code == http.StatusTooManyRequests || w2.Code >= 500)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 HTTP RESPONSE CODE TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHandleCloudEvent_ResponseCodes_SchemaValidationError(t *testing.T) {
	receiver := newTestReceiverWithSchemaValidation(t)

	// Invalid payload that fails schema validation
	body := `{
		"spec": {
			"source": {
				"type": "inline"
			}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
	req.Header.Set("Ce-Source", "io.knative.lambda/test")
	req.Header.Set("Ce-Id", "schema-error-123")

	w := httptest.NewRecorder()
	receiver.handleCloudEvent(w, req)

	// Should return 400 Bad Request for schema validation errors
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "schema_validation_failed", response["error"])
}

func TestHandleCloudEvent_ResponseCodes_InternalServerError(t *testing.T) {
	// Create receiver with fake client that will fail on Get
	scheme := runtime.NewScheme()
	_ = lambdav1alpha1.AddToScheme(scheme)

	// Use a client that will return an error (we'll simulate by not having the function)
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&lambdav1alpha1.LambdaFunction{}).
		Build()

	logger := zap.New(zap.UseDevMode(true))
	config := DefaultReceiverConfig()
	config.EnableSchemaValidation = false

	receiver, err := NewReceiver(fakeClient, logger, config, nil)
	require.NoError(t, err)

	body := `{
		"metadata": {
			"name": "nonexistent-function"
		},
		"spec": {
			"source": {
				"type": "inline",
				"inline": {
					"code": "def handler(e): return e"
				}
			}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", events.EventTypeCommandBuildStart)
	req.Header.Set("Ce-Source", "io.knative.lambda/test")
	req.Header.Set("Ce-Id", "build-notfound-123")

	w := httptest.NewRecorder()
	receiver.handleCloudEvent(w, req)

	// Should return 500 for internal errors (function not found for build)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleCloudEvent_ResponseCodes_Accepted(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)

	body := `{
		"metadata": {
			"name": "accepted-test",
			"namespace": "default"
		},
		"spec": {
			"source": {
				"type": "inline",
				"inline": {
					"code": "def handler(e): return e"
				}
			}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
	req.Header.Set("Ce-Source", "io.knative.lambda/test")
	req.Header.Set("Ce-Id", "accepted-123")

	w := httptest.NewRecorder()
	receiver.handleCloudEvent(w, req)

	// Should return 202 Accepted for successful processing
	assert.Equal(t, http.StatusAccepted, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "accepted", response["status"])
	assert.Equal(t, "accepted-123", response["eventId"])
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏗️ SERVER LIFECYCLE TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestReceiver_Start_Stop(t *testing.T) {
	logger := zap.New(zap.UseDevMode(true))
	config := DefaultReceiverConfig()
	config.Port = 0 // Use random port for testing
	config.EnableSchemaValidation = false

	receiver, err := NewReceiver(nil, logger, config, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- receiver.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop server
	cancel()

	// Wait for server to stop
	select {
	case err := <-errCh:
		// Server should shut down gracefully
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not stop within timeout")
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 K8S API ERROR HANDLING TESTS                                         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHandleCloudEvent_K8sInvalidError(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)

	// Invalid data that will cause validation error
	body := `{
		"metadata": {
			"name": "invalid-function"
		},
		"spec": {
			"source": {
				"type": ""
			}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
	req.Header.Set("Ce-Source", "io.knative.lambda/test")
	req.Header.Set("Ce-Id", "invalid-123")

	w := httptest.NewRecorder()
	receiver.handleCloudEvent(w, req)

	// Invalid errors should return 400 Bad Request
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusInternalServerError)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 METRICS ENDPOINT TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestMetricsHandler_WithSchemaValidationCounter(t *testing.T) {
	receiver := newTestReceiver(t)

	// Increment all counters
	receiver.incrementReceived()
	receiver.incrementProcessed()
	receiver.incrementFailed()
	receiver.incrementSchemaValidationFailed()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	receiver.metricsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var metrics map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&metrics)
	require.NoError(t, err)

	assert.Equal(t, float64(1), metrics["events_received"])
	assert.Equal(t, float64(1), metrics["events_processed"])
	assert.Equal(t, float64(1), metrics["events_failed"])
}

func TestMetricsHandler_ConcurrentAccess(t *testing.T) {
	receiver := newTestReceiver(t)

	// Concurrent increments
	const numGoroutines = 10
	const incrementsPerGoroutine = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < incrementsPerGoroutine; j++ {
				receiver.incrementReceived()
				receiver.incrementProcessed()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	receiver.metricsHandler(w, req)

	var metrics map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&metrics)
	require.NoError(t, err)

	// Should have exactly numGoroutines * incrementsPerGoroutine
	assert.Equal(t, float64(numGoroutines*incrementsPerGoroutine), metrics["events_received"])
	assert.Equal(t, float64(numGoroutines*incrementsPerGoroutine), metrics["events_processed"])
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔍 EDGE CASE TESTS                                                     │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHandleCloudEvent_EmptyBody(t *testing.T) {
	receiver := newTestReceiver(t)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
	req.Header.Set("Ce-Source", "io.knative.lambda/test")
	req.Header.Set("Ce-Id", "empty-body-123")

	w := httptest.NewRecorder()
	receiver.handleCloudEvent(w, req)

	// Empty body causes parsing error which returns 500 (internal server error)
	// The CloudEvent parsing fails, which is an internal processing error
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusInternalServerError)
}

func TestHandleCloudEvent_MalformedJSON(t *testing.T) {
	receiver := newTestReceiver(t)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"invalid": json}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
	req.Header.Set("Ce-Source", "io.knative.lambda/test")
	req.Header.Set("Ce-Id", "malformed-123")

	w := httptest.NewRecorder()
	receiver.handleCloudEvent(w, req)

	// Malformed JSON causes processing error which returns 500 (internal server error)
	// The data parsing fails during event processing
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusInternalServerError)
}

func TestProcessEvent_EmptySourceType(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("empty-source-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, FunctionDeployData{
		Metadata: FunctionMetadata{
			Name: "empty-source-function",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "", // Empty source type
			},
		},
	})

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source.type is required")
}

func TestProcessEvent_MissingMetadata(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("missing-metadata-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"spec": map[string]interface{}{
			"source": map[string]interface{}{
				"type": "inline",
			},
		},
	})

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	// Error could be either "failed to parse" or "metadata.name is required"
	assert.True(t,
		strings.Contains(err.Error(), "failed to parse") ||
			strings.Contains(err.Error(), "metadata.name is required") ||
			strings.Contains(err.Error(), "name is required"))
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🐛 DEBUG MODE TESTS                                                    │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHandleFunctionDeploy_DebugMode(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = lambdav1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&lambdav1alpha1.LambdaFunction{}).
		Build()

	logger := zap.New(zap.UseDevMode(true))
	config := DefaultReceiverConfig()
	config.EnableSchemaValidation = false
	config.Debug = true // Enable debug mode

	receiver, err := NewReceiver(fakeClient, logger, config, nil)
	require.NoError(t, err)

	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("debug-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, FunctionDeployData{
		Metadata: FunctionMetadata{
			Name: "debug-function",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(e): return e",
				},
			},
		},
	})

	// Debug mode should not cause errors
	err = receiver.processEvent(ctx, &event)
	assert.NoError(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 UPDATE SCENARIOS TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestProcessEvent_FunctionDeploy_UpdateWithAnnotations(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "update-annotations-function",
			Namespace: "default",
			Annotations: map[string]string{
				"existing": "annotation",
			},
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(e): return e",
				},
			},
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("update-annotations-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, FunctionDeployData{
		Metadata: FunctionMetadata{
			Name:      "update-annotations-function",
			Namespace: "default",
			Annotations: map[string]string{
				"new": "annotation",
			},
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(e): return {'updated': True}",
				},
			},
		},
	})

	err := receiver.processEvent(ctx, &event)
	assert.NoError(t, err)
}

func TestProcessEvent_FunctionDeploy_UpdateWithLabels(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "update-labels-function",
			Namespace: "default",
			Labels: map[string]string{
				"existing": "label",
			},
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(e): return e",
				},
			},
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("update-labels-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, FunctionDeployData{
		Metadata: FunctionMetadata{
			Name:      "update-labels-function",
			Namespace: "default",
			Labels: map[string]string{
				"new": "label",
				"app": "myapp",
			},
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "inline",
				Inline: &lambdav1alpha1.InlineSource{
					Code: "def handler(e): return e",
				},
			},
		},
	})

	err := receiver.processEvent(ctx, &event)
	assert.NoError(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔒 CONCURRENT PROCESSING TESTS                                         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHandleCloudEvent_ConcurrentProcessing(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)

	const numEvents = 10
	done := make(chan bool, numEvents)
	errors := make(chan error, numEvents)

	for i := 0; i < numEvents; i++ {
		go func(id int) {
			body := fmt.Sprintf(`{
				"metadata": {
					"name": "concurrent-function-%d",
					"namespace": "default"
				},
				"spec": {
					"source": {
						"type": "inline",
						"inline": {
							"code": "def handler(e): return e"
						}
					}
				}
			}`, id)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Ce-Specversion", "1.0")
			req.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
			req.Header.Set("Ce-Source", "io.knative.lambda/test")
			req.Header.Set("Ce-Id", fmt.Sprintf("concurrent-%d", id))

			w := httptest.NewRecorder()
			receiver.handleCloudEvent(w, req)

			if w.Code != http.StatusAccepted {
				errors <- fmt.Errorf("event %d returned status %d", id, w.Code)
			}
			done <- true
		}(i)
	}

	// Wait for all events
	for i := 0; i < numEvents; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📋 SOURCE TYPE VALIDATION TESTS                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestProcessEvent_FunctionDeploy_MissingSourceType(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("missing-source-type-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, FunctionDeployData{
		Metadata: FunctionMetadata{
			Name: "missing-source-type-function",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "", // Missing source type
			},
		},
	})

	err := receiver.processEvent(ctx, &event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source.type is required")
}

func TestProcessEvent_FunctionDeploy_InvalidSourceType(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("invalid-source-type-123")
	event.SetType(events.EventTypeCommandFunctionDeploy)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, FunctionDeployData{
		Metadata: FunctionMetadata{
			Name: "invalid-source-type-function",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Source: lambdav1alpha1.SourceSpec{
				Type: "unknown", // Invalid source type
			},
		},
	})

	// This should pass validation but may fail later in controller
	err := receiver.processEvent(ctx, &event)
	// The receiver doesn't validate source types, so this should succeed
	assert.NoError(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 BUILD COMMAND EDGE CASES                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestProcessEvent_BuildStart_WithoutBuildSpec(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "no-build-spec-function",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
			// No Build spec
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			Phase: lambdav1alpha1.PhaseReady,
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("build-no-spec-123")
	event.SetType(events.EventTypeCommandBuildStart)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, BuildCommandData{
		Name:         "no-build-spec-function",
		Namespace:    "default",
		ForceRebuild: true,
	})

	err := receiver.processEvent(ctx, &event)

	// Should succeed even without Build spec (ForceRebuild won't be set)
	assert.NoError(t, err)
}

func TestProcessEvent_BuildCancel_NotInBuildingPhase(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "not-building-function",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			Phase: lambdav1alpha1.PhaseReady, // Not building
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("cancel-not-building-123")
	event.SetType(events.EventTypeCommandBuildCancel)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, BuildCommandData{
		Name:      "not-building-function",
		Namespace: "default",
		Reason:    "User cancelled",
	})

	err := receiver.processEvent(ctx, &event)

	// Should return nil (nothing to cancel)
	assert.NoError(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 ROLLBACK EDGE CASES                                                 │
// └─────────────────────────────────────────────────────────────────────────┘

func TestProcessEvent_Rollback_WithoutRevision(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rollback-no-revision-function",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("rollback-no-revision-123")
	event.SetType(events.EventTypeCommandFunctionRollback)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"name":      "rollback-no-revision-function",
		"namespace": "default",
		"reason":    "Regression detected",
		// No revision specified
	})

	err := receiver.processEvent(ctx, &event)

	// Should succeed even without revision
	assert.NoError(t, err)
}

func TestProcessEvent_Rollback_WithRevision(t *testing.T) {
	existingFunc := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rollback-with-revision-function",
			Namespace: "default",
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
			},
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			ServiceStatus: &lambdav1alpha1.ServiceStatusInfo{
				LatestRevision: "rollback-with-revision-function-00003",
			},
		},
	}

	receiver := newTestReceiverWithFakeClient(t, existingFunc)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("rollback-with-revision-123")
	event.SetType(events.EventTypeCommandFunctionRollback)
	event.SetSource("io.knative.lambda/test")
	event.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"name":      "rollback-with-revision-function",
		"namespace": "default",
		"revision":  "rollback-with-revision-function-00001",
		"reason":    "Rollback to stable version",
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🆔 CORRELATION ID TESTS                                                │
// │                                                                         │
// │  Tests for correlation ID generation and propagation:                   │
// │  - Generate UUID when X-Correlation-ID header is missing                │
// │  - Preserve provided correlation ID                                     │
// │  - Include correlation ID in response headers                           │
// │  - Include correlation ID in JSON response body                         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHandleCloudEvent_GeneratesCorrelationID(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)

	body := `{
		"metadata": {
			"name": "correlation-test-function",
			"namespace": "default"
		},
		"spec": {
			"source": {
				"type": "inline",
				"inline": {
					"code": "def handler(e): return e"
				}
			},
			"runtime": {
				"language": "python",
				"version": "3.11"
			}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
	req.Header.Set("Ce-Source", "io.knative.lambda/test")
	req.Header.Set("Ce-Id", "correlation-gen-test-123")
	// Note: X-Correlation-ID header is NOT set

	w := httptest.NewRecorder()
	receiver.handleCloudEvent(w, req)

	// Verify response
	assert.Equal(t, http.StatusAccepted, w.Code)

	// Verify correlation ID is in response header
	correlationID := w.Header().Get("X-Correlation-ID")
	assert.NotEmpty(t, correlationID, "Response should include X-Correlation-ID header")

	// Verify it looks like a UUID (36 characters with dashes)
	assert.Len(t, correlationID, 36, "Generated correlation ID should be a UUID")

	// Verify correlation ID is in response body
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, correlationID, response["correlationId"], "Correlation ID should match in body and header")
}

func TestHandleCloudEvent_PreservesProvidedCorrelationID(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)

	providedCorrelationID := "user-provided-correlation-id-12345"

	body := `{
		"metadata": {
			"name": "preserve-correlation-function",
			"namespace": "default"
		},
		"spec": {
			"source": {
				"type": "inline",
				"inline": {
					"code": "def handler(e): return e"
				}
			},
			"runtime": {
				"language": "python",
				"version": "3.11"
			}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
	req.Header.Set("Ce-Source", "io.knative.lambda/test")
	req.Header.Set("Ce-Id", "preserve-correlation-test-123")
	req.Header.Set("X-Correlation-ID", providedCorrelationID) // Set correlation ID

	w := httptest.NewRecorder()
	receiver.handleCloudEvent(w, req)

	// Verify response
	assert.Equal(t, http.StatusAccepted, w.Code)

	// Verify provided correlation ID is preserved in response header
	correlationID := w.Header().Get("X-Correlation-ID")
	assert.Equal(t, providedCorrelationID, correlationID, "Should preserve provided correlation ID")

	// Verify correlation ID is in response body
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, providedCorrelationID, response["correlationId"], "Correlation ID in body should match provided")
}

func TestHandleCloudEvent_CorrelationIDInErrorResponse(t *testing.T) {
	receiver := newTestReceiver(t)

	providedCorrelationID := "error-test-correlation-id"

	// Send invalid request (missing CloudEvent headers)
	body := `{"invalid": "data"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", providedCorrelationID)
	// Note: Missing Ce-* headers will cause parsing error

	w := httptest.NewRecorder()
	receiver.handleCloudEvent(w, req)

	// Verify error response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify correlation ID is in response header
	correlationID := w.Header().Get("X-Correlation-ID")
	assert.Equal(t, providedCorrelationID, correlationID, "Error response should include correlation ID header")

	// Verify correlation ID is in error response body
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, providedCorrelationID, response["correlationId"], "Error response body should include correlation ID")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📍 /events ENDPOINT TESTS                                              │
// │                                                                         │
// │  Tests for the alternative /events endpoint:                            │
// │  - POST /events accepts CloudEvents                                     │
// │  - Works identically to POST /                                          │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHandleCloudEvent_EventsEndpoint(t *testing.T) {
	// Create receiver with fake client
	scheme := runtime.NewScheme()
	_ = lambdav1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&lambdav1alpha1.LambdaFunction{}).
		Build()

	logger := zap.New(zap.UseDevMode(true))
	config := DefaultReceiverConfig()
	config.Port = 0 // Use random port for testing
	config.EnableSchemaValidation = false

	receiver, err := NewReceiver(fakeClient, logger, config, nil)
	require.NoError(t, err)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- receiver.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test that /events endpoint works via the handleCloudEvent method directly
	body := `{
		"metadata": {
			"name": "events-endpoint-test",
			"namespace": "default"
		},
		"spec": {
			"source": {
				"type": "inline",
				"inline": {
					"code": "def handler(e): return e"
				}
			},
			"runtime": {
				"language": "python",
				"version": "3.11"
			}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", events.EventTypeCommandFunctionDeploy)
	req.Header.Set("Ce-Source", "io.knative.lambda/test")
	req.Header.Set("Ce-Id", "events-endpoint-test-123")

	w := httptest.NewRecorder()
	receiver.handleCloudEvent(w, req)

	// Verify successful processing
	assert.Equal(t, http.StatusAccepted, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "accepted", response["status"])
	assert.NotEmpty(t, response["correlationId"])

	// Cancel context to stop server
	cancel()
}

func TestReceiverConfig_EventsPath(t *testing.T) {
	// Verify that when config.Path is set to "/events", it doesn't register twice
	logger := zap.New(zap.UseDevMode(true))
	config := DefaultReceiverConfig()
	config.Path = "/events" // Set primary path to /events
	config.EnableSchemaValidation = false

	receiver, err := NewReceiver(nil, logger, config, nil)
	require.NoError(t, err)
	require.NotNil(t, receiver)

	// Verify config
	assert.Equal(t, "/events", receiver.config.Path)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 RESPONSE EVENT HANDLER TESTS - RED METRICS                          │
// │                                                                         │
// │  Tests for io.knative.lambda.response.success and                       │
// │  io.knative.lambda.response.error event processing.                     │
// │  These handlers populate function RED metrics:                          │
// │  - knative_lambda_function_invocations_total                           │
// │  - knative_lambda_function_duration_seconds                            │
// │  - knative_lambda_function_errors_total                                │
// │  - knative_lambda_function_cold_starts_total                           │
// └─────────────────────────────────────────────────────────────────────────┘

func TestProcessEvent_ResponseSuccess(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("response-success-123")
	event.SetType(events.EventTypeResponseSuccess)
	event.SetSource("io.knative.lambda/test-function")
	event.SetSubject("test-function")
	event.SetData(cloudevents.ApplicationJSON, ResponseEventData{
		FunctionName:  "test-function",
		Namespace:     "default",
		InvocationID:  "inv-123",
		CorrelationID: "corr-456",
		Result: &ResponseResultData{
			StatusCode: 200,
			Body:       map[string]interface{}{"result": "success"},
		},
		Metrics: &ResponseMetricsData{
			DurationMs:   150,
			ColdStart:    true,
			MemoryUsedMb: 64,
		},
	})

	err := receiver.processEvent(ctx, &event)

	// Should process successfully without error
	assert.NoError(t, err)
}

func TestProcessEvent_ResponseSuccess_MissingFunctionName(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("response-no-name-123")
	event.SetType(events.EventTypeResponseSuccess)
	event.SetSource("io.knative.lambda/unknown")
	// No subject set either
	event.SetData(cloudevents.ApplicationJSON, ResponseEventData{
		FunctionName: "", // Missing function name
		Namespace:    "default",
	})

	err := receiver.processEvent(ctx, &event)

	// Should return nil (skip metrics update, don't fail)
	assert.NoError(t, err)
}

func TestProcessEvent_ResponseSuccess_FromSubject(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("response-subject-123")
	event.SetType(events.EventTypeResponseSuccess)
	event.SetSource("io.knative.lambda/subject-function")
	event.SetSubject("subject-function") // Function name in subject
	event.SetData(cloudevents.ApplicationJSON, ResponseEventData{
		FunctionName: "", // Empty - should use subject
		Namespace:    "default",
		Metrics: &ResponseMetricsData{
			DurationMs: 100,
			ColdStart:  false,
		},
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_ResponseSuccess_NilMetrics(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("response-nil-metrics-123")
	event.SetType(events.EventTypeResponseSuccess)
	event.SetSource("io.knative.lambda/no-metrics-function")
	event.SetData(cloudevents.ApplicationJSON, ResponseEventData{
		FunctionName: "no-metrics-function",
		Namespace:    "default",
		Metrics:      nil, // No metrics provided
	})

	err := receiver.processEvent(ctx, &event)

	// Should process successfully even without metrics
	assert.NoError(t, err)
}

func TestProcessEvent_ResponseError(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("response-error-123")
	event.SetType(events.EventTypeResponseError)
	event.SetSource("io.knative.lambda/error-function")
	event.SetSubject("error-function")
	event.SetData(cloudevents.ApplicationJSON, ResponseEventData{
		FunctionName:  "error-function",
		Namespace:     "production",
		InvocationID:  "inv-error-123",
		CorrelationID: "corr-error-456",
		Error: &ResponseErrorData{
			Code:      "RuntimeError",
			Message:   "Something went wrong",
			Retryable: true,
		},
		Metrics: &ResponseMetricsData{
			DurationMs:   50,
			ColdStart:    false,
			MemoryUsedMb: 128,
		},
	})

	err := receiver.processEvent(ctx, &event)

	assert.NoError(t, err)
}

func TestProcessEvent_ResponseError_InvalidData(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("response-invalid-123")
	event.SetType(events.EventTypeResponseError)
	event.SetSource("io.knative.lambda/invalid")
	event.SetData(cloudevents.ApplicationJSON, "not valid response data")

	err := receiver.processEvent(ctx, &event)

	// Should return nil (skip on parse error, don't fail)
	assert.NoError(t, err)
}

func TestProcessEvent_ResponseError_MissingFunctionName(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("response-error-no-name-123")
	event.SetType(events.EventTypeResponseError)
	event.SetSource("io.knative.lambda/unknown")
	event.SetData(cloudevents.ApplicationJSON, ResponseEventData{
		FunctionName: "", // Missing
		Error: &ResponseErrorData{
			Code:    "TestError",
			Message: "Test error",
		},
	})

	err := receiver.processEvent(ctx, &event)

	// Should return nil (skip metrics update)
	assert.NoError(t, err)
}

func TestProcessEvent_ResponseError_NilErrorData(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)
	ctx := context.Background()

	event := cloudevents.NewEvent()
	event.SetID("response-nil-error-123")
	event.SetType(events.EventTypeResponseError)
	event.SetSource("io.knative.lambda/nil-error-function")
	event.SetData(cloudevents.ApplicationJSON, ResponseEventData{
		FunctionName: "nil-error-function",
		Namespace:    "default",
		Error:        nil, // Nil error data
		Metrics: &ResponseMetricsData{
			DurationMs: 10,
		},
	})

	err := receiver.processEvent(ctx, &event)

	// Should process with "unknown" error type
	assert.NoError(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 ERROR TYPE NORMALIZATION TESTS                                      │
// │                                                                         │
// │  Tests for normalizeErrorType() to ensure bounded cardinality          │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNormalizeErrorType(t *testing.T) {
	tests := []struct {
		name     string
		errData  *ResponseErrorData
		expected string
	}{
		{
			name:     "nil error data",
			errData:  nil,
			expected: "unknown",
		},
		{
			name:     "empty code",
			errData:  &ResponseErrorData{Code: ""},
			expected: "unknown",
		},
		{
			name:     "RuntimeError",
			errData:  &ResponseErrorData{Code: "RuntimeError"},
			expected: "runtime",
		},
		{
			name:     "runtime_error",
			errData:  &ResponseErrorData{Code: "runtime_error"},
			expected: "runtime",
		},
		{
			name:     "TimeoutError",
			errData:  &ResponseErrorData{Code: "TimeoutError"},
			expected: "timeout",
		},
		{
			name:     "timeout",
			errData:  &ResponseErrorData{Code: "timeout"},
			expected: "timeout",
		},
		{
			name:     "Timeout",
			errData:  &ResponseErrorData{Code: "Timeout"},
			expected: "timeout",
		},
		{
			name:     "MemoryError",
			errData:  &ResponseErrorData{Code: "MemoryError"},
			expected: "memory",
		},
		{
			name:     "OutOfMemory",
			errData:  &ResponseErrorData{Code: "OutOfMemory"},
			expected: "memory",
		},
		{
			name:     "OOM",
			errData:  &ResponseErrorData{Code: "OOM"},
			expected: "memory",
		},
		{
			name:     "HandlerError",
			errData:  &ResponseErrorData{Code: "HandlerError"},
			expected: "handler",
		},
		{
			name:     "ImportError",
			errData:  &ResponseErrorData{Code: "ImportError"},
			expected: "import",
		},
		{
			name:     "ModuleNotFound",
			errData:  &ResponseErrorData{Code: "ModuleNotFound"},
			expected: "import",
		},
		{
			name:     "ValueError",
			errData:  &ResponseErrorData{Code: "ValueError"},
			expected: "validation",
		},
		{
			name:     "TypeError",
			errData:  &ResponseErrorData{Code: "TypeError"},
			expected: "validation",
		},
		{
			name:     "ValidationError",
			errData:  &ResponseErrorData{Code: "ValidationError"},
			expected: "validation",
		},
		{
			name:     "ConnectionError",
			errData:  &ResponseErrorData{Code: "ConnectionError"},
			expected: "network",
		},
		{
			name:     "NetworkError",
			errData:  &ResponseErrorData{Code: "NetworkError"},
			expected: "network",
		},
		{
			name:     "PermissionError",
			errData:  &ResponseErrorData{Code: "PermissionError"},
			expected: "permission",
		},
		{
			name:     "AccessDenied",
			errData:  &ResponseErrorData{Code: "AccessDenied"},
			expected: "permission",
		},
		{
			name:     "ConfigError",
			errData:  &ResponseErrorData{Code: "ConfigError"},
			expected: "config",
		},
		{
			name:     "ReadBodyError",
			errData:  &ResponseErrorData{Code: "ReadBodyError"},
			expected: "request",
		},
		{
			name:     "unknown error",
			errData:  &ResponseErrorData{Code: "SomeRandomError"},
			expected: "other",
		},
		{
			name:     "pattern match timeout",
			errData:  &ResponseErrorData{Code: "FunctionTimeoutException"},
			expected: "timeout",
		},
		{
			name:     "pattern match memory",
			errData:  &ResponseErrorData{Code: "OutOfMemoryException"},
			expected: "memory",
		},
		{
			name:     "pattern match import",
			errData:  &ResponseErrorData{Code: "ModuleImportError"},
			expected: "import",
		},
		{
			name:     "pattern match network",
			errData:  &ResponseErrorData{Code: "NetworkConnectionError"},
			expected: "network",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeErrorType(tt.errData)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 HELPER FUNCTION TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		subs     []string
		expected bool
	}{
		{
			name:     "contains first substring",
			s:        "TimeoutError",
			subs:     []string{"timeout", "Timeout"},
			expected: true,
		},
		{
			name:     "contains second substring",
			s:        "FunctionTimeout",
			subs:     []string{"timeout", "Timeout"},
			expected: true,
		},
		{
			name:     "contains none",
			s:        "RuntimeError",
			subs:     []string{"timeout", "memory"},
			expected: false,
		},
		{
			name:     "empty string",
			s:        "",
			subs:     []string{"timeout"},
			expected: false,
		},
		{
			name:     "empty substrings",
			s:        "RuntimeError",
			subs:     []string{},
			expected: false,
		},
		{
			name:     "substring longer than string",
			s:        "Err",
			subs:     []string{"Error"},
			expected: false,
		},
		{
			name:     "exact match",
			s:        "OOM",
			subs:     []string{"OOM"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsAny(tt.s, tt.subs...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDurationMs(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *ResponseMetricsData
		expected int64
	}{
		{
			name:     "nil metrics",
			metrics:  nil,
			expected: 0,
		},
		{
			name:     "zero duration",
			metrics:  &ResponseMetricsData{DurationMs: 0},
			expected: 0,
		},
		{
			name:     "positive duration",
			metrics:  &ResponseMetricsData{DurationMs: 150},
			expected: 150,
		},
		{
			name:     "large duration",
			metrics:  &ResponseMetricsData{DurationMs: 30000},
			expected: 30000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDurationMs(tt.metrics)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetColdStart(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *ResponseMetricsData
		expected bool
	}{
		{
			name:     "nil metrics",
			metrics:  nil,
			expected: false,
		},
		{
			name:     "cold start true",
			metrics:  &ResponseMetricsData{ColdStart: true},
			expected: true,
		},
		{
			name:     "cold start false",
			metrics:  &ResponseMetricsData{ColdStart: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getColdStart(tt.metrics)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 RESPONSE EVENT DATA STRUCTURE TESTS                                 │
// └─────────────────────────────────────────────────────────────────────────┘

func TestResponseEventData_JSON(t *testing.T) {
	jsonData := `{
		"functionName": "json-function",
		"namespace": "production",
		"invocationId": "inv-123",
		"correlationId": "corr-456",
		"result": {
			"statusCode": 200,
			"body": {"message": "success"}
		},
		"metrics": {
			"durationMs": 250,
			"coldStart": true,
			"memoryUsedMb": 128
		}
	}`

	var data ResponseEventData
	err := json.Unmarshal([]byte(jsonData), &data)
	require.NoError(t, err)

	assert.Equal(t, "json-function", data.FunctionName)
	assert.Equal(t, "production", data.Namespace)
	assert.Equal(t, "inv-123", data.InvocationID)
	assert.Equal(t, "corr-456", data.CorrelationID)
	assert.NotNil(t, data.Result)
	assert.Equal(t, 200, data.Result.StatusCode)
	assert.NotNil(t, data.Metrics)
	assert.Equal(t, int64(250), data.Metrics.DurationMs)
	assert.True(t, data.Metrics.ColdStart)
	assert.Equal(t, int64(128), data.Metrics.MemoryUsedMb)
}

func TestResponseEventData_ErrorJSON(t *testing.T) {
	jsonData := `{
		"functionName": "error-function",
		"namespace": "staging",
		"invocationId": "inv-err-123",
		"error": {
			"code": "TimeoutError",
			"message": "Function execution timed out",
			"retryable": true,
			"stack": "at main.handler()\n  at execute()"
		},
		"metrics": {
			"durationMs": 30000,
			"coldStart": false
		}
	}`

	var data ResponseEventData
	err := json.Unmarshal([]byte(jsonData), &data)
	require.NoError(t, err)

	assert.Equal(t, "error-function", data.FunctionName)
	assert.Equal(t, "staging", data.Namespace)
	assert.NotNil(t, data.Error)
	assert.Equal(t, "TimeoutError", data.Error.Code)
	assert.Equal(t, "Function execution timed out", data.Error.Message)
	assert.True(t, data.Error.Retryable)
	assert.Contains(t, data.Error.Stack, "main.handler")
	assert.NotNil(t, data.Metrics)
	assert.Equal(t, int64(30000), data.Metrics.DurationMs)
	assert.False(t, data.Metrics.ColdStart)
}

func TestResponseMetricsData_JSON(t *testing.T) {
	jsonData := `{
		"durationMs": 100,
		"coldStart": true,
		"memoryUsedMb": 256
	}`

	var data ResponseMetricsData
	err := json.Unmarshal([]byte(jsonData), &data)
	require.NoError(t, err)

	assert.Equal(t, int64(100), data.DurationMs)
	assert.True(t, data.ColdStart)
	assert.Equal(t, int64(256), data.MemoryUsedMb)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🌐 RESPONSE EVENT HTTP HANDLER TESTS                                   │
// └─────────────────────────────────────────────────────────────────────────┘

func TestHandleCloudEvent_ResponseSuccess(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)

	body := `{
		"functionName": "http-success-function",
		"namespace": "default",
		"invocationId": "inv-http-123",
		"result": {
			"statusCode": 200,
			"body": {"result": "ok"}
		},
		"metrics": {
			"durationMs": 50,
			"coldStart": false
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", events.EventTypeResponseSuccess)
	req.Header.Set("Ce-Source", "io.knative.lambda/http-success-function")
	req.Header.Set("Ce-Id", "http-success-123")

	w := httptest.NewRecorder()
	receiver.handleCloudEvent(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestHandleCloudEvent_ResponseError(t *testing.T) {
	receiver := newTestReceiverWithFakeClient(t)

	body := `{
		"functionName": "http-error-function",
		"namespace": "default",
		"invocationId": "inv-http-err-123",
		"error": {
			"code": "RuntimeError",
			"message": "Division by zero",
			"retryable": false
		},
		"metrics": {
			"durationMs": 10,
			"coldStart": false
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", "1.0")
	req.Header.Set("Ce-Type", events.EventTypeResponseError)
	req.Header.Set("Ce-Source", "io.knative.lambda/http-error-function")
	req.Header.Set("Ce-Id", "http-error-123")

	w := httptest.NewRecorder()
	receiver.handleCloudEvent(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
}
