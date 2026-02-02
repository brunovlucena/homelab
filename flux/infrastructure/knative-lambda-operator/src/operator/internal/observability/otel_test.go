// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 Unit Tests: OpenTelemetry Observability (BVL-389 / AC4)
//
//	Tests for DEVOPS-001-AC4: OpenTelemetry Tracing
//	Acceptance Criteria:
//	- AC4.1: OTEL SDK is integrated (go.opentelemetry.io/otel)
//	- AC4.2: Trace context propagation uses W3C Trace Context (traceparent)
//	- AC4.3: OTLP exporter configured for Tempo endpoint
//	- AC4.4: Span creation covers key operations
//	- AC4.5: Resource attributes are set
//	- AC4.6: Trace sampling is configurable
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package observability

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.1: OTEL SDK INTEGRATION TESTS                                      │
// │  Verify go.opentelemetry.io/otel is integrated in operator code         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_1_OTEL_SDK_ServiceNameConstant(t *testing.T) {
	assert.Equal(t, "knative-lambda-operator", ServiceName)
}

func TestAC4_1_OTEL_SDK_ServiceNamespaceConstant(t *testing.T) {
	assert.Equal(t, "knative-lambda", ServiceNamespace)
}

func TestAC4_1_OTEL_SDK_SpanNameConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"Reconcile", SpanNameReconcile, "reconcile"},
		{"ReconcilePhase", SpanNameReconcilePhase, "reconcile.phase"},
		{"BuildContext", SpanNameBuildContext, "build.create_context"},
		{"BuildJob", SpanNameBuildJob, "build.create_job"},
		{"BuildStatus", SpanNameBuildStatus, "build.get_status"},
		{"DeployService", SpanNameDeployService, "deploy.create_service"},
		{"DeployStatus", SpanNameDeployStatus, "deploy.get_status"},
		{"EventingReconcile", SpanNameEventingReconcile, "eventing.reconcile"},
		{"CloudEventReceive", SpanNameCloudEventReceive, "cloudevents.receive"},
		{"CloudEventProcess", SpanNameCloudEventProcess, "cloudevents.process"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.3: OTLP EXPORTER CONFIGURATION TESTS                               │
// │  Verify OTLP exporter is configured to send traces to Tempo endpoint    │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_3_OTLP_Exporter_DefaultTempoEndpoint(t *testing.T) {
	// Clear env vars first
	os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	os.Unsetenv("OTEL_TRACES_SAMPLER_ARG")
	os.Unsetenv("OTEL_TRACING_ENABLED")
	os.Unsetenv("OTEL_SERVICE_NAME")
	os.Unsetenv("OTEL_SERVICE_NAMESPACE")
	os.Unsetenv("VERSION")
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("POD_NAME")
	os.Unsetenv("POD_NAMESPACE")
	os.Unsetenv("NODE_NAME")

	cfg := DefaultConfig()

	assert.Equal(t, ServiceName, cfg.ServiceName)
	assert.Equal(t, ServiceNamespace, cfg.ServiceNamespace)
	assert.Equal(t, "tempo.tempo.svc:4317", cfg.OTLPEndpoint, "Default endpoint should be Tempo")
	assert.Equal(t, 1.0, cfg.TracingSamplingRate)
	assert.True(t, cfg.MetricsEnabled)
	assert.True(t, cfg.TracingEnabled)
	assert.Equal(t, "production", cfg.Environment)
}

func TestAC4_3_OTLP_Exporter_CustomEndpointFromEnv(t *testing.T) {
	// Set env vars
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "custom-collector:4317")
	os.Setenv("OTEL_TRACES_SAMPLER_ARG", "0.5")
	os.Setenv("OTEL_TRACING_ENABLED", "true")
	os.Setenv("OTEL_SERVICE_NAME", "custom-service")
	os.Setenv("OTEL_SERVICE_NAMESPACE", "custom-namespace")
	os.Setenv("VERSION", "v1.0.0")
	os.Setenv("ENVIRONMENT", "staging")
	os.Setenv("POD_NAME", "operator-pod-xyz")
	os.Setenv("POD_NAMESPACE", "knative-lambda")
	os.Setenv("NODE_NAME", "worker-1")
	defer func() {
		os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		os.Unsetenv("OTEL_TRACES_SAMPLER_ARG")
		os.Unsetenv("OTEL_TRACING_ENABLED")
		os.Unsetenv("OTEL_SERVICE_NAME")
		os.Unsetenv("OTEL_SERVICE_NAMESPACE")
		os.Unsetenv("VERSION")
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("POD_NAME")
		os.Unsetenv("POD_NAMESPACE")
		os.Unsetenv("NODE_NAME")
	}()

	cfg := DefaultConfig()

	assert.Equal(t, "custom-service", cfg.ServiceName)
	assert.Equal(t, "custom-namespace", cfg.ServiceNamespace)
	assert.Equal(t, "custom-collector:4317", cfg.OTLPEndpoint)
	assert.Equal(t, 0.5, cfg.TracingSamplingRate)
	assert.Equal(t, "v1.0.0", cfg.ServiceVersion)
	assert.Equal(t, "staging", cfg.Environment)
	assert.Equal(t, "operator-pod-xyz", cfg.PodName)
	assert.Equal(t, "knative-lambda", cfg.PodNamespace)
	assert.Equal(t, "worker-1", cfg.NodeName)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.6: TRACE SAMPLING CONFIGURATION TESTS                              │
// │  Verify trace sampling is configurable (default 100% for dev)           │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_6_TraceSampling_DisabledViaEnv(t *testing.T) {
	os.Setenv("OTEL_TRACING_ENABLED", "false")
	defer os.Unsetenv("OTEL_TRACING_ENABLED")

	cfg := DefaultConfig()

	assert.False(t, cfg.TracingEnabled, "Tracing should be disabled via env var")
}

func TestAC4_6_TraceSampling_InvalidRateDefaultsTo100Percent(t *testing.T) {
	os.Setenv("OTEL_TRACES_SAMPLER_ARG", "invalid")
	defer os.Unsetenv("OTEL_TRACES_SAMPLER_ARG")

	cfg := DefaultConfig()

	assert.Equal(t, 1.0, cfg.TracingSamplingRate, "Invalid rate should default to 100%")
}

func TestAC4_6_TraceSampling_ConfigurableRate(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected float64
	}{
		{"100_percent", "1.0", 1.0},
		{"50_percent", "0.5", 0.5},
		{"10_percent", "0.1", 0.1},
		{"0_percent", "0", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("OTEL_TRACES_SAMPLER_ARG", tt.envValue)
			defer os.Unsetenv("OTEL_TRACES_SAMPLER_ARG")

			cfg := DefaultConfig()
			assert.Equal(t, tt.expected, cfg.TracingSamplingRate)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.1: OTEL SDK HELPER FUNCTIONS                                       │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_1_OTEL_SDK_GetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "Returns_env_value_when_set",
			key:          "TEST_VAR_1",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "Returns_default_when_env_not_set",
			key:          "TEST_VAR_2",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvOrDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAC4_6_TraceSampling_ParseFloat(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  float64
		expectErr bool
	}{
		{"Valid_float", "0.5", 0.5, false},
		{"Valid_integer", "1", 1.0, false},
		{"Valid_zero", "0", 0.0, false},
		{"Invalid_string", "abc", 0, true},
		{"Empty_string", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFloat(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.1: OTEL SDK PROVIDER INITIALIZATION                                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_1_OTEL_SDK_ProviderCreation_TracingDisabled(t *testing.T) {
	cfg := Config{
		ServiceName:      "test-service",
		ServiceNamespace: "test-namespace",
		TracingEnabled:   false,
		MetricsEnabled:   false,
	}

	provider, err := NewProvider(cfg)

	require.NoError(t, err)
	require.NotNil(t, provider)
	assert.Nil(t, provider.tracerProvider)
	assert.Nil(t, provider.meterProvider)
}

func TestAC4_1_OTEL_SDK_StartSpan_NilTracerSafety(t *testing.T) {
	provider := &Provider{}

	ctx, span := provider.StartSpan(context.Background(), "test-span")

	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	// Should return a no-op span
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.4: SPAN CREATION TESTS                                             │
// │  Verify spans cover: Reconcile, Build, Knative Service, CloudEvents     │
// └─────────────────────────────────────────────────────────────────────────┘

// createTestProvider creates a provider with an in-memory exporter for testing
func createTestProvider(t *testing.T) (*Provider, *tracetest.InMemoryExporter) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	provider := &Provider{
		config: Config{
			ServiceName:      "test-service",
			ServiceNamespace: "test-namespace",
			TracingEnabled:   true,
		},
		tracerProvider: tp,
		Tracer:         tp.Tracer("test-service"),
	}

	return provider, exporter
}

// AC4.4.1: Reconcile operations (per phase)
func TestAC4_4_SpanCreation_ReconcileOperations(t *testing.T) {
	provider, exporter := createTestProvider(t)

	ctx, span := provider.StartReconcileSpan(context.Background(), "test-function", "test-namespace")
	span.End()

	assert.NotNil(t, ctx)
	require.NotNil(t, span)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	recordedSpan := spans[0]
	assert.Equal(t, SpanNameReconcile, recordedSpan.Name)

	// Check attributes
	attrs := getSpanAttributes(recordedSpan)
	assert.Equal(t, "test-function", attrs["lambda.function"])
	assert.Equal(t, "test-namespace", attrs["lambda.namespace"])
	assert.Equal(t, "reconcile", attrs["operation.type"])
}

func TestAC4_4_SpanCreation_ReconcilePhases(t *testing.T) {
	provider, exporter := createTestProvider(t)

	phases := []string{"Pending", "Building", "Deploying", "Ready", "Failed", "Deleting"}

	for _, phase := range phases {
		t.Run("Phase_"+phase, func(t *testing.T) {
			exporter.Reset()

			ctx, span := provider.StartReconcilePhaseSpan(context.Background(), "test-func", "test-ns", phase)
			span.End()

			assert.NotNil(t, ctx)

			spans := exporter.GetSpans()
			require.Len(t, spans, 1)

			attrs := getSpanAttributes(spans[0])
			assert.Equal(t, phase, attrs["lambda.phase"])
			assert.Equal(t, "reconcile.phase", attrs["operation.type"])
		})
	}
}

// AC4.4.2: Build job creation and monitoring
func TestAC4_4_SpanCreation_BuildJobCreation(t *testing.T) {
	provider, exporter := createTestProvider(t)

	runtimes := []string{"python", "nodejs", "go"}

	for _, runtime := range runtimes {
		t.Run("Runtime_"+runtime, func(t *testing.T) {
			exporter.Reset()

			ctx, span := provider.StartBuildContextSpan(context.Background(), "test-func", "test-ns", runtime)
			span.End()

			assert.NotNil(t, ctx)

			spans := exporter.GetSpans()
			require.Len(t, spans, 1)

			attrs := getSpanAttributes(spans[0])
			assert.Equal(t, runtime, attrs["lambda.runtime"])
			assert.Equal(t, "build.context", attrs["operation.type"])
		})
	}
}

func TestAC4_4_SpanCreation_BuildJobSpan(t *testing.T) {
	provider, exporter := createTestProvider(t)

	ctx, span := provider.StartBuildJobSpan(context.Background(), "test-func", "test-ns", "test-job-123")
	span.End()

	assert.NotNil(t, ctx)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := getSpanAttributes(spans[0])
	assert.Equal(t, "test-job-123", attrs["build.job_name"])
	assert.Equal(t, "build.job", attrs["operation.type"])
}

func TestAC4_4_SpanCreation_BuildJobMonitoring(t *testing.T) {
	provider, exporter := createTestProvider(t)

	ctx, span := provider.StartBuildStatusSpan(context.Background(), "test-ns", "test-job-456")
	span.End()

	assert.NotNil(t, ctx)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := getSpanAttributes(spans[0])
	assert.Equal(t, "test-job-456", attrs["build.job_name"])
	assert.Equal(t, "build.status", attrs["operation.type"])
}

// AC4.4.3: Knative Service creation
func TestAC4_4_SpanCreation_KnativeServiceCreation(t *testing.T) {
	provider, exporter := createTestProvider(t)

	ctx, span := provider.StartDeployServiceSpan(context.Background(), "test-func", "test-ns")
	span.End()

	assert.NotNil(t, ctx)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := getSpanAttributes(spans[0])
	assert.Equal(t, "test-func", attrs["lambda.function"])
	assert.Equal(t, "deploy.service", attrs["operation.type"])
}

func TestAC4_4_SpanCreation_KnativeServiceStatus(t *testing.T) {
	provider, exporter := createTestProvider(t)

	ctx, span := provider.StartDeployStatusSpan(context.Background(), "test-func", "test-ns")
	span.End()

	assert.NotNil(t, ctx)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := getSpanAttributes(spans[0])
	assert.Equal(t, "deploy.status", attrs["operation.type"])
}

func TestAC4_4_SpanCreation_EventingReconcile(t *testing.T) {
	provider, exporter := createTestProvider(t)

	ctx, span := provider.StartEventingReconcileSpan(context.Background(), "test-func", "test-ns")
	span.End()

	assert.NotNil(t, ctx)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := getSpanAttributes(spans[0])
	assert.Equal(t, "eventing.reconcile", attrs["operation.type"])
}

// AC4.4.4: CloudEvents processing
func TestAC4_4_SpanCreation_CloudEventsReceive(t *testing.T) {
	provider, exporter := createTestProvider(t)

	ctx, span := provider.StartCloudEventReceiveSpan(
		context.Background(),
		"io.knative.lambda.invoke",
		"test-source",
		"event-123",
	)
	span.End()

	assert.NotNil(t, ctx)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := getSpanAttributes(spans[0])
	assert.Equal(t, "io.knative.lambda.invoke", attrs["cloudevents.type"])
	assert.Equal(t, "test-source", attrs["cloudevents.source"])
	assert.Equal(t, "event-123", attrs["cloudevents.id"])
	assert.Equal(t, "cloudevents.receive", attrs["operation.type"])
}

func TestAC4_4_SpanCreation_CloudEventsProcessing(t *testing.T) {
	provider, exporter := createTestProvider(t)

	ctx, span := provider.StartCloudEventProcessSpan(
		context.Background(),
		"io.knative.lambda.invoke",
		"test-func",
		"test-ns",
	)
	span.End()

	assert.NotNil(t, ctx)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := getSpanAttributes(spans[0])
	assert.Equal(t, "io.knative.lambda.invoke", attrs["cloudevents.type"])
	assert.Equal(t, "cloudevents.process", attrs["operation.type"])
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.1: OTEL SDK SPAN HELPER FUNCTIONS                                  │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_1_OTEL_SDK_RecordError(t *testing.T) {
	provider, exporter := createTestProvider(t)

	ctx, span := provider.StartSpan(context.Background(), "test-span")

	testErr := assert.AnError
	RecordError(span, testErr, "test error message")
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	recordedSpan := spans[0]
	assert.Equal(t, codes.Error, recordedSpan.Status.Code)
	assert.Equal(t, "test error message", recordedSpan.Status.Description)

	// Check that error event was recorded
	require.NotEmpty(t, recordedSpan.Events)

	_ = ctx // use ctx to avoid lint error
}

func TestAC4_1_OTEL_SDK_RecordError_NilSpanSafety(t *testing.T) {
	// Should not panic with nil span
	RecordError(nil, assert.AnError, "test")
}

func TestAC4_1_OTEL_SDK_RecordError_NilErrorSafety(t *testing.T) {
	provider, exporter := createTestProvider(t)

	_, span := provider.StartSpan(context.Background(), "test-span")

	// Should not panic with nil error
	RecordError(span, nil, "test")
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	// Status should not be error
	assert.NotEqual(t, codes.Error, spans[0].Status.Code)
}

func TestAC4_1_OTEL_SDK_SetSpanOK(t *testing.T) {
	provider, exporter := createTestProvider(t)

	_, span := provider.StartSpan(context.Background(), "test-span")
	SetSpanOK(span)
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	assert.Equal(t, codes.Ok, spans[0].Status.Code)
}

func TestAC4_1_OTEL_SDK_SetSpanOK_NilSpanSafety(t *testing.T) {
	// Should not panic with nil span
	SetSpanOK(nil)
}

func TestAC4_1_OTEL_SDK_AddSpanEvent(t *testing.T) {
	provider, exporter := createTestProvider(t)

	_, span := provider.StartSpan(context.Background(), "test-span")
	AddSpanEvent(span, "test_event",
		attribute.String("key1", "value1"),
		attribute.Int("key2", 42),
	)
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	events := spans[0].Events
	require.Len(t, events, 1)
	assert.Equal(t, "test_event", events[0].Name)
}

func TestAC4_1_OTEL_SDK_AddSpanEvent_NilSpanSafety(t *testing.T) {
	// Should not panic with nil span
	AddSpanEvent(nil, "test_event")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.2: W3C TRACE CONTEXT - CONTEXT HELPERS                             │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_2_W3C_TraceContext_SpanFromContext(t *testing.T) {
	provider, _ := createTestProvider(t)

	ctx, span := provider.StartSpan(context.Background(), "test-span")

	retrievedSpan := SpanFromContext(ctx)

	assert.Equal(t, span.SpanContext().TraceID(), retrievedSpan.SpanContext().TraceID())
	assert.Equal(t, span.SpanContext().SpanID(), retrievedSpan.SpanContext().SpanID())

	span.End()
}

func TestAC4_2_W3C_TraceContext_SpanFromContext_NoSpan(t *testing.T) {
	span := SpanFromContext(context.Background())

	// Should return a no-op span, not nil
	assert.NotNil(t, span)
	assert.False(t, span.SpanContext().IsValid())
}

func TestAC4_2_W3C_TraceContext_ContextWithSpan(t *testing.T) {
	provider, _ := createTestProvider(t)

	_, span := provider.StartSpan(context.Background(), "test-span")

	newCtx := ContextWithSpan(context.Background(), span)

	retrievedSpan := trace.SpanFromContext(newCtx)
	assert.Equal(t, span.SpanContext().SpanID(), retrievedSpan.SpanContext().SpanID())

	span.End()
}

func TestAC4_2_W3C_TraceContext_GetTraceID(t *testing.T) {
	provider, _ := createTestProvider(t)

	ctx, span := provider.StartSpan(context.Background(), "test-span")

	traceID := GetTraceID(ctx)

	assert.NotEmpty(t, traceID)
	assert.Len(t, traceID, 32, "Trace ID should be 16 bytes = 32 hex chars")

	span.End()
}

func TestAC4_2_W3C_TraceContext_GetTraceID_NoSpan(t *testing.T) {
	traceID := GetTraceID(context.Background())

	assert.Empty(t, traceID)
}

func TestAC4_2_W3C_TraceContext_GetSpanID(t *testing.T) {
	provider, _ := createTestProvider(t)

	ctx, span := provider.StartSpan(context.Background(), "test-span")

	spanID := GetSpanID(ctx)

	assert.NotEmpty(t, spanID)
	assert.Len(t, spanID, 16, "Span ID should be 8 bytes = 16 hex chars")

	span.End()
}

func TestAC4_2_W3C_TraceContext_GetSpanID_NoSpan(t *testing.T) {
	spanID := GetSpanID(context.Background())

	assert.Empty(t, spanID)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.1: OTEL SDK GLOBAL PROVIDER                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_1_OTEL_SDK_SetGetGlobalProvider(t *testing.T) {
	// Save original
	original := globalProvider
	defer func() { globalProvider = original }()

	provider := &Provider{
		config: Config{ServiceName: "test"},
	}

	SetGlobalProvider(provider)

	retrieved := GetGlobalProvider()
	assert.Equal(t, provider, retrieved)
}

func TestAC4_1_OTEL_SDK_Tracer_WithGlobalProvider(t *testing.T) {
	// Save original
	original := globalProvider
	defer func() { globalProvider = original }()

	provider, _ := createTestProvider(t)
	SetGlobalProvider(provider)

	tracer := Tracer()

	assert.NotNil(t, tracer)
}

func TestAC4_1_OTEL_SDK_Tracer_NoGlobalProvider(t *testing.T) {
	// Save original
	original := globalProvider
	defer func() { globalProvider = original }()

	globalProvider = nil

	tracer := Tracer()

	// Should return a tracer from the global otel provider
	assert.NotNil(t, tracer)
}

func TestAC4_1_OTEL_SDK_StartSpanFromContext(t *testing.T) {
	// Save original
	original := globalProvider
	defer func() { globalProvider = original }()

	provider, exporter := createTestProvider(t)
	SetGlobalProvider(provider)

	ctx, span := StartSpanFromContext(context.Background(), "test-span")
	span.End()

	assert.NotNil(t, ctx)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "test-span", spans[0].Name)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.5: RESOURCE ATTRIBUTES TESTS                                       │
// │  Verify: service.name, service.namespace, k8s.pod.name, k8s.namespace   │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_5_ResourceAttributes_LambdaAttributes(t *testing.T) {
	provider, exporter := createTestProvider(t)

	opts := WithLambdaAttributes("my-function", "my-namespace", "python")

	ctx, span := provider.Tracer.Start(context.Background(), "test-span", opts)
	span.End()

	_ = ctx

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := getSpanAttributes(spans[0])
	assert.Equal(t, "my-function", attrs["lambda.function"])
	assert.Equal(t, "my-namespace", attrs["lambda.namespace"])
	assert.Equal(t, "python", attrs["lambda.runtime"])
}

func TestAC4_5_ResourceAttributes_BuildAttributes(t *testing.T) {
	provider, exporter := createTestProvider(t)

	opts := WithBuildAttributes("job-123", "registry/image:tag")

	ctx, span := provider.Tracer.Start(context.Background(), "test-span", opts)
	span.End()

	_ = ctx

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := getSpanAttributes(spans[0])
	assert.Equal(t, "job-123", attrs["build.job_name"])
	assert.Equal(t, "registry/image:tag", attrs["build.image_uri"])
}

func TestAC4_5_ResourceAttributes_CloudEventAttributes(t *testing.T) {
	provider, exporter := createTestProvider(t)

	opts := WithCloudEventAttributes("io.knative.lambda.invoke", "source", "event-id")

	ctx, span := provider.Tracer.Start(context.Background(), "test-span", opts)
	span.End()

	_ = ctx

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := getSpanAttributes(spans[0])
	assert.Equal(t, "io.knative.lambda.invoke", attrs["cloudevents.type"])
	assert.Equal(t, "source", attrs["cloudevents.source"])
	assert.Equal(t, "event-id", attrs["cloudevents.id"])
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.1: OTEL SDK BACKWARD COMPATIBILITY (DEPRECATED METHODS)            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_1_OTEL_SDK_RecordBuildSpan_BackwardCompat(t *testing.T) {
	provider, exporter := createTestProvider(t)

	ctx, span := provider.RecordBuildSpan(context.Background(), "test-func", "test-ns", "python")
	span.End()

	assert.NotNil(t, ctx)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := getSpanAttributes(spans[0])
	assert.Equal(t, "python", attrs["lambda.runtime"])
}

func TestAC4_1_OTEL_SDK_RecordDeploySpan_BackwardCompat(t *testing.T) {
	provider, exporter := createTestProvider(t)

	ctx, span := provider.RecordDeploySpan(context.Background(), "test-func", "test-ns")
	span.End()

	assert.NotNil(t, ctx)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
}

func TestAC4_1_OTEL_SDK_RecordReconcileSpan_BackwardCompat(t *testing.T) {
	provider, exporter := createTestProvider(t)

	ctx, span := provider.RecordReconcileSpan(context.Background(), "test-func", "test-ns", "Building")
	span.End()

	assert.NotNil(t, ctx)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := getSpanAttributes(spans[0])
	assert.Equal(t, "Building", attrs["lambda.phase"])
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.2: W3C TRACE CONTEXT PROPAGATION (traceparent header)              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_2_W3C_TraceContext_TraceparentHeaderInjection(t *testing.T) {
	// Setup propagator
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(propagator)

	provider, _ := createTestProvider(t)

	// Create a span
	ctx, span := provider.StartSpan(context.Background(), "parent-span")

	// Extract trace context into carrier
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	// Verify traceparent header is present
	traceparent, ok := carrier["traceparent"]
	assert.True(t, ok, "traceparent header should be present")
	assert.NotEmpty(t, traceparent)

	// Verify format: version-traceid-parentid-flags
	// e.g., "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"
	assert.Contains(t, traceparent, "-", "traceparent should use W3C format with dashes")

	span.End()
}

func TestAC4_2_W3C_TraceContext_TraceparentHeaderExtraction(t *testing.T) {
	// Setup propagator
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(propagator)

	// Create carrier with traceparent (W3C format)
	carrier := propagation.MapCarrier{
		"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
	}

	// Extract context
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

	// Verify span context is extracted
	spanCtx := trace.SpanContextFromContext(ctx)
	assert.True(t, spanCtx.IsValid(), "Span context should be valid after extraction")
	assert.Equal(t, "0af7651916cd43dd8448eb211c80319c", spanCtx.TraceID().String())
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.4: SPAN CREATION - NESTED SPAN HIERARCHY                           │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_4_SpanCreation_NestedSpanHierarchy(t *testing.T) {
	provider, exporter := createTestProvider(t)

	// Create parent span
	ctx, parentSpan := provider.StartReconcileSpan(context.Background(), "test-func", "test-ns")

	// Create child span
	ctx, childSpan := provider.StartReconcilePhaseSpan(ctx, "test-func", "test-ns", "Building")

	// Create grandchild span
	_, grandchildSpan := provider.StartBuildContextSpan(ctx, "test-func", "test-ns", "python")

	// End in reverse order
	grandchildSpan.End()
	childSpan.End()
	parentSpan.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 3)

	// Verify parent-child relationships via TraceID
	parentTraceID := spans[2].SpanContext.TraceID()
	for _, span := range spans {
		assert.Equal(t, parentTraceID, span.SpanContext.TraceID(), "All spans should share the same trace ID")
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.1: OTEL SDK PROVIDER LIFECYCLE                                     │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_1_OTEL_SDK_ProviderShutdown(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)

	provider := &Provider{
		tracerProvider: tp,
	}

	err := provider.Shutdown(context.Background())

	assert.NoError(t, err)
}

func TestAC4_1_OTEL_SDK_ProviderShutdown_NilSafety(t *testing.T) {
	provider := &Provider{}

	err := provider.Shutdown(context.Background())

	assert.NoError(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  TEST HELPERS                                                           │
// └─────────────────────────────────────────────────────────────────────────┘

// getSpanAttributes extracts attributes from a recorded span as a map
func getSpanAttributes(span tracetest.SpanStub) map[string]string {
	attrs := make(map[string]string)
	for _, attr := range span.Attributes {
		attrs[string(attr.Key)] = attr.Value.AsString()
	}
	return attrs
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.4: SPAN CREATION - SPAN KIND VERIFICATION                          │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_4_SpanCreation_SpanKinds(t *testing.T) {
	provider, exporter := createTestProvider(t)

	tests := []struct {
		name         string
		spanFunc     func() trace.Span
		expectedKind trace.SpanKind
	}{
		{
			name: "CloudEventReceive_is_Consumer_SpanKind",
			spanFunc: func() trace.Span {
				_, span := provider.StartCloudEventReceiveSpan(context.Background(), "type", "source", "id")
				return span
			},
			expectedKind: trace.SpanKindConsumer,
		},
		{
			name: "Reconcile_is_Internal_SpanKind",
			spanFunc: func() trace.Span {
				_, span := provider.StartReconcileSpan(context.Background(), "func", "ns")
				return span
			},
			expectedKind: trace.SpanKindInternal,
		},
		{
			name: "BuildContext_is_Internal_SpanKind",
			spanFunc: func() trace.Span {
				_, span := provider.StartBuildContextSpan(context.Background(), "func", "ns", "python")
				return span
			},
			expectedKind: trace.SpanKindInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter.Reset()

			span := tt.spanFunc()
			span.End()

			spans := exporter.GetSpans()
			require.Len(t, spans, 1)
			assert.Equal(t, tt.expectedKind, spans[0].SpanKind)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.4: SPAN CREATION - CONCURRENT SAFETY                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_4_SpanCreation_ConcurrentSafety(t *testing.T) {
	provider, exporter := createTestProvider(t)

	const numGoroutines = 50
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			ctx, span := provider.StartReconcileSpan(context.Background(), "func", "ns")
			_, phaseSpan := provider.StartReconcilePhaseSpan(ctx, "func", "ns", "Building")
			phaseSpan.End()
			span.End()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	spans := exporter.GetSpans()
	assert.Equal(t, numGoroutines*2, len(spans))
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  AC4.4: SPAN CREATION - FULL RECONCILE LIFECYCLE TRACE                  │
// │  Integration test verifying all span types work together                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAC4_4_SpanCreation_FullReconcileLifecycleTrace(t *testing.T) {
	provider, exporter := createTestProvider(t)

	// Simulate a full reconcile lifecycle
	ctx := context.Background()

	// 1. Start main reconcile span
	ctx, reconcileSpan := provider.StartReconcileSpan(ctx, "hello-python", "knative-lambda")

	// 2. Pending phase - create build context
	ctx, pendingSpan := provider.StartReconcilePhaseSpan(ctx, "hello-python", "knative-lambda", "Pending")
	ctx, buildCtxSpan := provider.StartBuildContextSpan(ctx, "hello-python", "knative-lambda", "python")
	buildCtxSpan.SetAttributes(attribute.String("build.context_configmap", "hello-python-build-context"))
	SetSpanOK(buildCtxSpan)
	buildCtxSpan.End()

	// 3. Create build job
	ctx, buildJobSpan := provider.StartBuildJobSpan(ctx, "hello-python", "knative-lambda", "hello-python-build-123")
	SetSpanOK(buildJobSpan)
	buildJobSpan.End()
	SetSpanOK(pendingSpan)
	pendingSpan.End()

	// 4. Building phase - monitor build
	ctx, buildingSpan := provider.StartReconcilePhaseSpan(ctx, "hello-python", "knative-lambda", "Building")
	ctx, buildStatusSpan := provider.StartBuildStatusSpan(ctx, "knative-lambda", "hello-python-build-123")
	buildStatusSpan.SetAttributes(
		attribute.Bool("build.completed", true),
		attribute.Bool("build.success", true),
		attribute.String("build.image_uri", "localhost:5001/knative-lambda/hello-python:abc123"),
	)
	SetSpanOK(buildStatusSpan)
	buildStatusSpan.End()
	SetSpanOK(buildingSpan)
	buildingSpan.End()

	// 5. Deploying phase
	ctx, deployingSpan := provider.StartReconcilePhaseSpan(ctx, "hello-python", "knative-lambda", "Deploying")

	// Eventing reconcile
	ctx, eventingSpan := provider.StartEventingReconcileSpan(ctx, "hello-python", "knative-lambda")
	SetSpanOK(eventingSpan)
	eventingSpan.End()

	// Deploy service
	ctx, deploySpan := provider.StartDeployServiceSpan(ctx, "hello-python", "knative-lambda")
	AddSpanEvent(deploySpan, "creating_service")
	AddSpanEvent(deploySpan, "service_created", attribute.String("service.name", "hello-python"))

	// Check status
	_, statusSpan := provider.StartDeployStatusSpan(ctx, "hello-python", "knative-lambda")
	statusSpan.SetAttributes(
		attribute.Bool("service.ready", true),
		attribute.String("service.url", "http://hello-python.knative-lambda.svc.cluster.local"),
	)
	SetSpanOK(statusSpan)
	statusSpan.End()

	SetSpanOK(deploySpan)
	deploySpan.End()
	SetSpanOK(deployingSpan)
	deployingSpan.End()

	// 6. Complete reconcile
	reconcileSpan.SetAttributes(attribute.Float64("reconcile.duration_ms", 5000.0))
	SetSpanOK(reconcileSpan)
	reconcileSpan.End()

	// Verify all spans were created
	spans := exporter.GetSpans()
	assert.GreaterOrEqual(t, len(spans), 10, "Should have at least 10 spans for full lifecycle")

	// All spans should share the same trace ID
	traceID := spans[0].SpanContext.TraceID()
	for _, span := range spans {
		assert.Equal(t, traceID, span.SpanContext.TraceID())
	}

	// Verify all spans have OK status
	for _, span := range spans {
		assert.Equal(t, codes.Ok, span.Status.Code, "Span %s should have OK status", span.Name)
	}
}
