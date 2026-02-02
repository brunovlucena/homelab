// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 Unit Tests: Events Manager
//
//	Tests for CloudEvents emission:
//	- Event type constants
//	- Event data structures
//	- Event type detection helpers
//	- Event data building
//	- Manager configuration
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📝 CONSTANTS TESTS                                                     │
// └─────────────────────────────────────────────────────────────────────────┘

func TestEventTypePrefix(t *testing.T) {
	assert.Equal(t, "io.knative.lambda", EventTypePrefix)
	assert.NotEmpty(t, DefaultSourceFormat)
	assert.NotEmpty(t, DefaultBrokerURL)
}

func TestEventTypeConstants(t *testing.T) {
	// Command events
	assert.Contains(t, EventTypeCommandBuildStart, "command.build")
	assert.Contains(t, EventTypeCommandBuildCancel, "command.build")
	assert.Contains(t, EventTypeCommandServiceCreate, "command.service")
	assert.Contains(t, EventTypeCommandServiceDelete, "command.service")
	assert.Contains(t, EventTypeCommandFunctionDeploy, "command.function")

	// Lifecycle events
	assert.Contains(t, EventTypeLifecycleFunctionCreated, "lifecycle.function")
	assert.Contains(t, EventTypeLifecycleFunctionUpdated, "lifecycle.function")
	assert.Contains(t, EventTypeLifecycleFunctionDeleted, "lifecycle.function")
	assert.Contains(t, EventTypeLifecycleBuildStarted, "lifecycle.build")
	assert.Contains(t, EventTypeLifecycleBuildCompleted, "lifecycle.build")
	assert.Contains(t, EventTypeLifecycleBuildFailed, "lifecycle.build")
	assert.Contains(t, EventTypeLifecycleServiceCreated, "lifecycle.service")

	// Invoke events
	assert.Contains(t, EventTypeInvokeSync, "invoke.sync")
	assert.Contains(t, EventTypeInvokeAsync, "invoke.async")
	assert.Contains(t, EventTypeInvokeScheduled, "invoke.scheduled")

	// Response events
	assert.Contains(t, EventTypeResponseSuccess, "response.success")
	assert.Contains(t, EventTypeResponseError, "response.error")
	assert.Contains(t, EventTypeResponseTimeout, "response.timeout")

	// Notification events
	assert.Contains(t, EventTypeNotificationAlertCritical, "notification.alert")
	assert.Contains(t, EventTypeNotificationAlertWarning, "notification.alert")
	assert.Contains(t, EventTypeNotificationAuditChange, "notification.audit")
}

func TestLegacyEventTypeAliases(t *testing.T) {
	// Legacy aliases should point to new event types
	assert.Equal(t, EventTypeLifecycleFunctionCreated, EventTypeFunctionCreated)
	assert.Equal(t, EventTypeLifecycleFunctionUpdated, EventTypeFunctionUpdated)
	assert.Equal(t, EventTypeLifecycleFunctionDeleted, EventTypeFunctionDeleted)
	assert.Equal(t, EventTypeLifecycleBuildStarted, EventTypeBuildStarted)
	assert.Equal(t, EventTypeLifecycleBuildCompleted, EventTypeBuildCompleted)
	assert.Equal(t, EventTypeLifecycleBuildFailed, EventTypeBuildFailed)
	assert.Equal(t, EventTypeLifecycleBuildCancelled, EventTypeBuildCancelled)
	assert.Equal(t, EventTypeLifecycleServiceCreated, EventTypeServiceCreated)
	assert.Equal(t, EventTypeLifecycleServiceDeleted, EventTypeServiceDeleted)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔍 EVENT TYPE DETECTION TESTS                                          │
// └─────────────────────────────────────────────────────────────────────────┘

func TestIsCommandEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		expected  bool
	}{
		{
			name:      "Build start command",
			eventType: EventTypeCommandBuildStart,
			expected:  true,
		},
		{
			name:      "Service create command",
			eventType: EventTypeCommandServiceCreate,
			expected:  true,
		},
		{
			name:      "Function deploy command",
			eventType: EventTypeCommandFunctionDeploy,
			expected:  true,
		},
		{
			name:      "Lifecycle event not command",
			eventType: EventTypeLifecycleBuildStarted,
			expected:  false,
		},
		{
			name:      "Invoke event not command",
			eventType: EventTypeInvokeAsync,
			expected:  false,
		},
		{
			name:      "Response event not command",
			eventType: EventTypeResponseSuccess,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCommandEvent(tt.eventType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsLifecycleEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		expected  bool
	}{
		{
			name:      "Function created lifecycle",
			eventType: EventTypeLifecycleFunctionCreated,
			expected:  true,
		},
		{
			name:      "Build completed lifecycle",
			eventType: EventTypeLifecycleBuildCompleted,
			expected:  true,
		},
		{
			name:      "Service ready lifecycle",
			eventType: EventTypeLifecycleServiceReady,
			expected:  true,
		},
		{
			name:      "Command event not lifecycle",
			eventType: EventTypeCommandBuildStart,
			expected:  false,
		},
		{
			name:      "Invoke event not lifecycle",
			eventType: EventTypeInvokeSync,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLifecycleEvent(tt.eventType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsInvokeEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		expected  bool
	}{
		{
			name:      "Sync invoke",
			eventType: EventTypeInvokeSync,
			expected:  true,
		},
		{
			name:      "Async invoke",
			eventType: EventTypeInvokeAsync,
			expected:  true,
		},
		{
			name:      "Scheduled invoke",
			eventType: EventTypeInvokeScheduled,
			expected:  true,
		},
		{
			name:      "Retry invoke",
			eventType: EventTypeInvokeRetry,
			expected:  true,
		},
		{
			name:      "Response event not invoke",
			eventType: EventTypeResponseSuccess,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInvokeEvent(tt.eventType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsResponseEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		expected  bool
	}{
		{
			name:      "Success response",
			eventType: EventTypeResponseSuccess,
			expected:  true,
		},
		{
			name:      "Error response",
			eventType: EventTypeResponseError,
			expected:  true,
		},
		{
			name:      "Timeout response",
			eventType: EventTypeResponseTimeout,
			expected:  true,
		},
		{
			name:      "Invoke event not response",
			eventType: EventTypeInvokeAsync,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsResponseEvent(tt.eventType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNotificationEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		expected  bool
	}{
		{
			name:      "Critical alert notification",
			eventType: EventTypeNotificationAlertCritical,
			expected:  true,
		},
		{
			name:      "Warning alert notification",
			eventType: EventTypeNotificationAlertWarning,
			expected:  true,
		},
		{
			name:      "Audit change notification",
			eventType: EventTypeNotificationAuditChange,
			expected:  true,
		},
		{
			name:      "Response event not notification",
			eventType: EventTypeResponseSuccess,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotificationEvent(tt.eventType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏭 MANAGER TESTS                                                       │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNewManager(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		m := NewManager(Config{})

		assert.NotNil(t, m)
		assert.Equal(t, DefaultBrokerURL, m.config.BrokerURL)
		assert.Equal(t, DefaultSourceFormat, m.config.Source)
		assert.NotNil(t, m.httpClient)
	})

	t.Run("Custom configuration", func(t *testing.T) {
		m := NewManager(Config{
			BrokerURL: "http://custom-broker:8080",
			Source:    "custom.source/test",
			Enabled:   true,
		})

		assert.Equal(t, "http://custom-broker:8080", m.config.BrokerURL)
		assert.Equal(t, "custom.source/test", m.config.Source)
		assert.True(t, m.config.Enabled)
	})

	t.Run("HTTP client timeout", func(t *testing.T) {
		m := NewManager(Config{})

		assert.Equal(t, 10*time.Second, m.httpClient.Timeout)
	})
}

func TestGetSource(t *testing.T) {
	m := NewManager(Config{
		Source: "io.knative.lambda/operator",
	})

	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-function",
			Namespace: "default",
		},
	}

	source := m.getSource(lambda)

	assert.Contains(t, source, "io.knative.lambda/operator")
	assert.Contains(t, source, "default")
	assert.Contains(t, source, "test-function")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 EVENT DATA BUILDING TESTS                                           │
// └─────────────────────────────────────────────────────────────────────────┘

func TestBuildFunctionEventData(t *testing.T) {
	m := NewManager(Config{})

	now := metav1.Now()
	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "my-function",
			Namespace:  "production",
			Generation: 5,
		},
		Spec: lambdav1alpha1.LambdaFunctionSpec{
			Runtime: lambdav1alpha1.RuntimeSpec{
				Language: "python",
				Version:  "3.11",
				Handler:  "main.handler",
			},
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			Phase:              lambdav1alpha1.PhaseReady,
			ObservedGeneration: 5,
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					Reason:             "ServiceReady",
					Message:            "Function is ready",
					LastTransitionTime: now,
				},
			},
		},
	}

	data := m.buildFunctionEventData(lambda)

	assert.Equal(t, "my-function", data.Name)
	assert.Equal(t, "production", data.Namespace)
	assert.Equal(t, "Ready", data.Phase)
	assert.Equal(t, int64(5), data.Generation)
	assert.Equal(t, int64(5), data.ObservedGeneration)

	// Check runtime data
	require.NotNil(t, data.Runtime)
	assert.Equal(t, "python", data.Runtime.Language)
	assert.Equal(t, "3.11", data.Runtime.Version)
	assert.Equal(t, "main.handler", data.Runtime.Handler)

	// Check conditions
	require.Len(t, data.Conditions, 1)
	assert.Equal(t, "Ready", data.Conditions[0].Type)
	assert.Equal(t, "True", data.Conditions[0].Status)
	assert.Equal(t, "ServiceReady", data.Conditions[0].Reason)
}

func TestBuildBuildEventData(t *testing.T) {
	m := NewManager(Config{})

	startTime := metav1.Now()
	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "build-test",
			Namespace: "default",
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			BuildStatus: &lambdav1alpha1.BuildStatusInfo{
				JobName:   "build-test-build-12345",
				ImageURI:  "localhost:5001/default/build-test:abc123",
				StartedAt: &startTime,
			},
		},
	}

	data := m.buildBuildEventData(lambda)

	assert.Equal(t, "build-test", data.Name)
	assert.Equal(t, "default", data.Namespace)
	assert.Equal(t, "build-test-build-12345", data.JobName)
	assert.Equal(t, "localhost:5001/default/build-test:abc123", data.ImageURI)
	assert.NotEmpty(t, data.StartedAt)
}

func TestBuildBuildEventData_NilStatus(t *testing.T) {
	m := NewManager(Config{})

	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "no-build",
			Namespace: "default",
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			BuildStatus: nil,
		},
	}

	data := m.buildBuildEventData(lambda)

	assert.Equal(t, "no-build", data.Name)
	assert.Equal(t, "default", data.Namespace)
	assert.Empty(t, data.JobName)
	assert.Empty(t, data.ImageURI)
}

func TestBuildServiceEventData(t *testing.T) {
	m := NewManager(Config{})

	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-test",
			Namespace: "production",
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			ServiceStatus: &lambdav1alpha1.ServiceStatusInfo{
				ServiceName:    "service-test",
				URL:            "http://service-test.production.svc.cluster.local",
				Ready:          true,
				Replicas:       3,
				LatestRevision: "service-test-00001",
			},
		},
	}

	data := m.buildServiceEventData(lambda)

	assert.Equal(t, "service-test", data.Name)
	assert.Equal(t, "production", data.Namespace)
	assert.Equal(t, "service-test", data.ServiceName)
	assert.Equal(t, "http://service-test.production.svc.cluster.local", data.URL)
	assert.True(t, data.Ready)
	assert.Equal(t, "service-test-00001", data.LatestRevision)
}

func TestBuildServiceEventData_NilStatus(t *testing.T) {
	m := NewManager(Config{})

	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "no-service",
			Namespace: "default",
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			ServiceStatus: nil,
		},
	}

	data := m.buildServiceEventData(lambda)

	assert.Equal(t, "no-service", data.Name)
	assert.Equal(t, "default", data.Namespace)
	assert.Empty(t, data.ServiceName)
	assert.Empty(t, data.URL)
	assert.False(t, data.Ready)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 DATA STRUCTURE TESTS                                                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestErrorData(t *testing.T) {
	err := &ErrorData{
		Type:      "BuildError",
		Message:   "Failed to build image",
		Code:      "BUILD_FAILED",
		Retryable: true,
	}

	assert.Equal(t, "BuildError", err.Type)
	assert.Equal(t, "Failed to build image", err.Message)
	assert.Equal(t, "BUILD_FAILED", err.Code)
	assert.True(t, err.Retryable)
}

func TestInvokeEventData(t *testing.T) {
	data := &InvokeEventData{
		FunctionName:  "test-func",
		Namespace:     "default",
		InvocationID:  "inv-123",
		CorrelationID: "corr-456",
		Payload:       map[string]interface{}{"key": "value"},
		Metadata: &InvokeMetadata{
			TraceID:    "trace-789",
			SpanID:     "span-012",
			RetryCount: 2,
			DeadlineAt: "2024-01-01T12:00:00Z",
		},
	}

	assert.Equal(t, "test-func", data.FunctionName)
	assert.Equal(t, "inv-123", data.InvocationID)
	assert.Equal(t, "corr-456", data.CorrelationID)
	assert.Equal(t, "value", data.Payload["key"])
	require.NotNil(t, data.Metadata)
	assert.Equal(t, "trace-789", data.Metadata.TraceID)
	assert.Equal(t, 2, data.Metadata.RetryCount)
}

func TestResponseEventData(t *testing.T) {
	data := &ResponseEventData{
		FunctionName:  "test-func",
		Namespace:     "default",
		InvocationID:  "inv-123",
		CorrelationID: "corr-456",
		Result: &ResultData{
			StatusCode: 200,
			Body:       map[string]interface{}{"status": "ok"},
		},
		Metrics: &ResponseMetrics{
			DurationMs:   150,
			ColdStart:    false,
			MemoryUsedMb: 64,
		},
	}

	assert.Equal(t, "test-func", data.FunctionName)
	require.NotNil(t, data.Result)
	assert.Equal(t, 200, data.Result.StatusCode)
	require.NotNil(t, data.Metrics)
	assert.Equal(t, int64(150), data.Metrics.DurationMs)
	assert.False(t, data.Metrics.ColdStart)
}

func TestDLQData(t *testing.T) {
	data := &DLQData{
		Routed:    true,
		QueueName: "my-function-dlq",
		RoutedAt:  "2024-01-01T12:00:00Z",
	}

	assert.True(t, data.Routed)
	assert.Equal(t, "my-function-dlq", data.QueueName)
	assert.NotEmpty(t, data.RoutedAt)
}

func TestAlertEventData(t *testing.T) {
	data := &AlertEventData{
		AlertName:   "HighErrorRate",
		Severity:    "critical",
		Summary:     "High error rate detected",
		Description: "Error rate is above 10%",
		Labels: map[string]string{
			"function": "my-func",
			"env":      "production",
		},
		StartsAt: "2024-01-01T12:00:00Z",
	}

	assert.Equal(t, "HighErrorRate", data.AlertName)
	assert.Equal(t, "critical", data.Severity)
	assert.Equal(t, "my-func", data.Labels["function"])
}

func TestAuditEventData(t *testing.T) {
	data := &AuditEventData{
		Action:       "UPDATE",
		Resource:     "LambdaFunction",
		ResourceName: "my-function",
		Namespace:    "default",
		User:         "admin",
		Reason:       "Scaling configuration changed",
		Details: map[string]string{
			"oldReplicas": "1",
			"newReplicas": "5",
		},
		Timestamp: "2024-01-01T12:00:00Z",
	}

	assert.Equal(t, "UPDATE", data.Action)
	assert.Equal(t, "LambdaFunction", data.Resource)
	assert.Equal(t, "my-function", data.ResourceName)
	assert.Equal(t, "admin", data.User)
	assert.Equal(t, "5", data.Details["newReplicas"])
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 TEST HELPERS                                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func newTestLambdaFunction(name, namespace string) *lambdav1alpha1.LambdaFunction {
	return &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Generation: 1,
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
				Handler:  "handler",
			},
		},
		Status: lambdav1alpha1.LambdaFunctionStatus{
			Phase: lambdav1alpha1.PhasePending,
		},
	}
}

func TestNewTestLambdaFunction(t *testing.T) {
	lambda := newTestLambdaFunction("test", "default")

	assert.Equal(t, "test", lambda.Name)
	assert.Equal(t, "default", lambda.Namespace)
	assert.Equal(t, "python", lambda.Spec.Runtime.Language)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ⚡ CONFIG TESTS                                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestConfig(t *testing.T) {
	config := Config{
		BrokerURL: "http://broker:8080",
		Source:    "my-source",
		Enabled:   true,
	}

	assert.Equal(t, "http://broker:8080", config.BrokerURL)
	assert.Equal(t, "my-source", config.Source)
	assert.True(t, config.Enabled)
}

func TestConfigDefaults(t *testing.T) {
	config := Config{}

	// Empty config should have empty values
	assert.Empty(t, config.BrokerURL)
	assert.Empty(t, config.Source)
	assert.False(t, config.Enabled)

	// NewManager should set defaults
	m := NewManager(config)
	assert.Equal(t, DefaultBrokerURL, m.config.BrokerURL)
	assert.Equal(t, DefaultSourceFormat, m.config.Source)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 RUNTIME DATA TESTS                                                  │
// └─────────────────────────────────────────────────────────────────────────┘

func TestRuntimeData(t *testing.T) {
	data := &RuntimeData{
		Language: "nodejs",
		Version:  "20",
		Handler:  "index.handler",
	}

	assert.Equal(t, "nodejs", data.Language)
	assert.Equal(t, "20", data.Version)
	assert.Equal(t, "index.handler", data.Handler)
}

func TestConditionData(t *testing.T) {
	data := &ConditionData{
		Type:               "Ready",
		Status:             "True",
		Reason:             "ServiceReady",
		Message:            "All conditions met",
		LastTransitionTime: "2024-01-01T12:00:00Z",
	}

	assert.Equal(t, "Ready", data.Type)
	assert.Equal(t, "True", data.Status)
	assert.Equal(t, "ServiceReady", data.Reason)
}

func TestReplicasData(t *testing.T) {
	data := &ReplicasData{
		Desired:   3,
		Ready:     2,
		Available: 2,
	}

	assert.Equal(t, int32(3), data.Desired)
	assert.Equal(t, int32(2), data.Ready)
	assert.Equal(t, int32(2), data.Available)
}

func TestTrafficData(t *testing.T) {
	data := &TrafficData{
		RevisionName:   "my-func-00001",
		Percent:        100,
		LatestRevision: true,
	}

	assert.Equal(t, "my-func-00001", data.RevisionName)
	assert.Equal(t, int64(100), data.Percent)
	assert.True(t, data.LatestRevision)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 METRICS DATA TESTS                                                  │
// └─────────────────────────────────────────────────────────────────────────┘

func TestResponseMetrics(t *testing.T) {
	metrics := &ResponseMetrics{
		DurationMs:   250,
		ColdStart:    true,
		MemoryUsedMb: 128,
	}

	assert.Equal(t, int64(250), metrics.DurationMs)
	assert.True(t, metrics.ColdStart)
	assert.Equal(t, int64(128), metrics.MemoryUsedMb)
}

func TestInvokeMetadata(t *testing.T) {
	metadata := &InvokeMetadata{
		TraceID:    "abc123",
		SpanID:     "def456",
		RetryCount: 0,
		DeadlineAt: "2024-01-01T12:05:00Z",
	}

	assert.Equal(t, "abc123", metadata.TraceID)
	assert.Equal(t, "def456", metadata.SpanID)
	assert.Equal(t, 0, metadata.RetryCount)
}
