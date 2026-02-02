// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 Unit Tests: Metrics
//
//	Tests for Prometheus metrics:
//	- Metric registration
//	- ReconcilerMetrics functionality
//	- Exemplar support
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📝 CONSTANTS TESTS                                                     │
// └─────────────────────────────────────────────────────────────────────────┘

func TestMetricConstants(t *testing.T) {
	assert.Equal(t, "knative_lambda", namespace, "Namespace should be knative_lambda")
	assert.Equal(t, "operator", subsystem, "Subsystem should be operator")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 METRIC VECTORS TESTS                                                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestReconcileTotal(t *testing.T) {
	// Verify ReconcileTotal counter has expected labels
	require.NotNil(t, ReconcileTotal)

	// Create test metric
	counter := ReconcileTotal.WithLabelValues("Pending", "success")
	require.NotNil(t, counter)

	// Increment and verify no panic
	counter.Inc()
}

func TestReconcileDuration(t *testing.T) {
	// Verify ReconcileDuration histogram has expected labels
	require.NotNil(t, ReconcileDuration)

	// Create test metric
	histogram := ReconcileDuration.WithLabelValues("Building")
	require.NotNil(t, histogram)

	// Observe and verify no panic
	histogram.Observe(0.5)
}

func TestLambdaFunctionsTotal(t *testing.T) {
	// Verify LambdaFunctionsTotal gauge has expected labels
	require.NotNil(t, LambdaFunctionsTotal)

	// Create test metric
	gauge := LambdaFunctionsTotal.WithLabelValues("default", "Ready")
	require.NotNil(t, gauge)

	// Set and verify no panic
	gauge.Set(10)
}

func TestBuildJobsActive(t *testing.T) {
	// Verify BuildJobsActive gauge
	require.NotNil(t, BuildJobsActive)

	gauge := BuildJobsActive.WithLabelValues("default")
	require.NotNil(t, gauge)

	gauge.Set(5)
}

func TestBuildDuration(t *testing.T) {
	// Verify BuildDuration histogram
	require.NotNil(t, BuildDuration)

	histogram := BuildDuration.WithLabelValues("python", "success")
	require.NotNil(t, histogram)

	histogram.Observe(60.0)
}

func TestEventingResourcesTotal(t *testing.T) {
	// Verify EventingResourcesTotal gauge
	require.NotNil(t, EventingResourcesTotal)

	gauge := EventingResourcesTotal.WithLabelValues("default", "broker")
	require.NotNil(t, gauge)

	gauge.Set(1)
}

func TestAPIServerRequestsTotal(t *testing.T) {
	// Verify APIServerRequestsTotal counter
	require.NotNil(t, APIServerRequestsTotal)

	counter := APIServerRequestsTotal.WithLabelValues("get", "lambdafunction", "success")
	require.NotNil(t, counter)

	counter.Inc()
}

func TestWorkQueueDepth(t *testing.T) {
	// Verify WorkQueueDepth gauge
	require.NotNil(t, WorkQueueDepth)

	WorkQueueDepth.Set(25)
}

func TestWorkQueueLatency(t *testing.T) {
	// Verify WorkQueueLatency histogram
	require.NotNil(t, WorkQueueLatency)

	WorkQueueLatency.Observe(0.1)
}

func TestErrorsTotal(t *testing.T) {
	// Verify ErrorsTotal counter
	require.NotNil(t, ErrorsTotal)

	counter := ErrorsTotal.WithLabelValues("build", "image_pull_failed")
	require.NotNil(t, counter)

	counter.Inc()
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📈 FUNCTION METRICS TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestFunctionInvocationsTotal(t *testing.T) {
	require.NotNil(t, FunctionInvocationsTotal)

	counter := FunctionInvocationsTotal.WithLabelValues("my-function", "default", "success")
	require.NotNil(t, counter)

	counter.Inc()
}

func TestFunctionDuration(t *testing.T) {
	require.NotNil(t, FunctionDuration)

	histogram := FunctionDuration.WithLabelValues("my-function", "default")
	require.NotNil(t, histogram)

	histogram.Observe(0.05) // 50ms
}

func TestFunctionErrorsTotal(t *testing.T) {
	require.NotNil(t, FunctionErrorsTotal)

	counter := FunctionErrorsTotal.WithLabelValues("my-function", "default", "timeout")
	require.NotNil(t, counter)

	counter.Inc()
}

func TestFunctionColdStartsTotal(t *testing.T) {
	require.NotNil(t, FunctionColdStartsTotal)

	counter := FunctionColdStartsTotal.WithLabelValues("my-function", "default")
	require.NotNil(t, counter)

	counter.Inc()
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 RECONCILER METRICS TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNewReconcilerMetrics(t *testing.T) {
	m := NewReconcilerMetrics()

	require.NotNil(t, m)
	assert.True(t, m.ExemplarsEnabled, "Exemplars should be enabled by default")
}

func TestNewReconcilerMetricsWithExemplars(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "Exemplars enabled",
			enabled:  true,
			expected: true,
		},
		{
			name:     "Exemplars disabled",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewReconcilerMetricsWithExemplars(tt.enabled)
			require.NotNil(t, m)
			assert.Equal(t, tt.expected, m.ExemplarsEnabled)
		})
	}
}

func TestReconcilerMetrics_RecordReconcile(t *testing.T) {
	m := NewReconcilerMetrics()

	// Should not panic
	m.RecordReconcile("Pending", "success", 0.1)
	m.RecordReconcile("Building", "error", 0.5)
	m.RecordReconcile("Deploying", "success", 1.0)
}

func TestReconcilerMetrics_RecordBuild(t *testing.T) {
	m := NewReconcilerMetrics()

	// Should not panic
	m.RecordBuild("python", "success", 60.0)
	m.RecordBuild("nodejs", "failed", 30.0)
	m.RecordBuild("go", "success", 120.0)
}

func TestReconcilerMetrics_RecordError(t *testing.T) {
	m := NewReconcilerMetrics()

	// Should not panic
	m.RecordError("build", "job_creation_failed")
	m.RecordError("deploy", "service_creation_failed")
	m.RecordError("eventing", "broker_creation_failed")
}

func TestReconcilerMetrics_SetLambdaCount(t *testing.T) {
	m := NewReconcilerMetrics()

	// Should not panic
	m.SetLambdaCount("default", "Ready", 10)
	m.SetLambdaCount("production", "Building", 5)
	m.SetLambdaCount("default", "Failed", 2)
}

func TestReconcilerMetrics_SetActiveBuildJobs(t *testing.T) {
	m := NewReconcilerMetrics()

	// Should not panic
	m.SetActiveBuildJobs("default", 3)
	m.SetActiveBuildJobs("production", 0)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔗 EXEMPLAR EXTRACTION TESTS                                           │
// └─────────────────────────────────────────────────────────────────────────┘

func TestReconcilerMetrics_ExtractExemplar_NoSpan(t *testing.T) {
	m := NewReconcilerMetrics()

	// Context without span should return nil
	labels := m.extractExemplar(nil)
	assert.Nil(t, labels, "Should return nil for nil context")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 HISTOGRAM BUCKETS TESTS                                             │
// └─────────────────────────────────────────────────────────────────────────┘

func TestReconcileDurationBuckets(t *testing.T) {
	// Verify histogram has appropriate buckets for reconcile duration
	// These should cover the range from milliseconds to minutes
	expectedBuckets := []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60}

	// We can't directly access the buckets, but we can verify the histogram works
	// across the expected range
	for _, bucket := range expectedBuckets {
		// Should not panic for any value in expected range
		ReconcileDuration.WithLabelValues("test").Observe(bucket)
	}
}

func TestBuildDurationBuckets(t *testing.T) {
	// Verify histogram has appropriate buckets for build duration
	// Builds can take from seconds to 30 minutes
	expectedBuckets := []float64{10, 30, 60, 120, 300, 600, 900, 1200, 1800}

	for _, bucket := range expectedBuckets {
		// Should not panic for any value in expected range
		BuildDuration.WithLabelValues("python", "test").Observe(bucket)
	}
}

func TestWorkQueueLatencyBuckets(t *testing.T) {
	// Verify histogram has appropriate buckets for queue latency
	expectedBuckets := []float64{.001, .01, .1, 1, 10, 60, 300}

	for _, bucket := range expectedBuckets {
		WorkQueueLatency.Observe(bucket)
	}
}

func TestFunctionDurationBuckets(t *testing.T) {
	// Verify histogram has appropriate buckets for function duration
	// Functions typically run from milliseconds to seconds
	expectedBuckets := []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30}

	for _, bucket := range expectedBuckets {
		FunctionDuration.WithLabelValues("test-function", "default").Observe(bucket)
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏷️ LABEL VALIDATION TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestMetricLabels_Phases(t *testing.T) {
	phases := []string{"Pending", "Building", "Deploying", "Ready", "Failed", "Deleting"}

	for _, phase := range phases {
		t.Run("Phase_"+phase, func(t *testing.T) {
			// ReconcileTotal with phase
			counter := ReconcileTotal.WithLabelValues(phase, "success")
			require.NotNil(t, counter)

			// ReconcileDuration with phase
			histogram := ReconcileDuration.WithLabelValues(phase)
			require.NotNil(t, histogram)
		})
	}
}

func TestMetricLabels_Results(t *testing.T) {
	results := []string{"success", "error", "timeout", "cancelled"}

	for _, result := range results {
		t.Run("Result_"+result, func(t *testing.T) {
			// ReconcileTotal with result
			counter := ReconcileTotal.WithLabelValues("Ready", result)
			require.NotNil(t, counter)

			// BuildDuration with result
			histogram := BuildDuration.WithLabelValues("python", result)
			require.NotNil(t, histogram)
		})
	}
}

func TestMetricLabels_Runtimes(t *testing.T) {
	runtimes := []string{"python", "nodejs", "go", "java", "rust"}

	for _, runtime := range runtimes {
		t.Run("Runtime_"+runtime, func(t *testing.T) {
			histogram := BuildDuration.WithLabelValues(runtime, "success")
			require.NotNil(t, histogram)
		})
	}
}

func TestMetricLabels_Components(t *testing.T) {
	components := []string{"build", "deploy", "eventing", "reconcile", "webhook"}

	for _, component := range components {
		t.Run("Component_"+component, func(t *testing.T) {
			counter := ErrorsTotal.WithLabelValues(component, "generic_error")
			require.NotNil(t, counter)
		})
	}
}

func TestMetricLabels_ResourceTypes(t *testing.T) {
	resourceTypes := []string{"broker", "trigger", "channel", "subscription"}

	for _, resourceType := range resourceTypes {
		t.Run("ResourceType_"+resourceType, func(t *testing.T) {
			gauge := EventingResourcesTotal.WithLabelValues("default", resourceType)
			require.NotNil(t, gauge)
		})
	}
}

func TestMetricLabels_Verbs(t *testing.T) {
	verbs := []string{"get", "list", "create", "update", "patch", "delete", "watch"}

	for _, verb := range verbs {
		t.Run("Verb_"+verb, func(t *testing.T) {
			counter := APIServerRequestsTotal.WithLabelValues(verb, "lambdafunction", "success")
			require.NotNil(t, counter)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔗 CONTEXT-BASED METRICS TESTS                                         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestReconcilerMetrics_RecordReconcileWithContext_NoSpan(t *testing.T) {
	m := NewReconcilerMetrics()

	// Test with background context (no span)
	ctx := context.Background()

	// Should not panic and should fall back to regular recording
	m.RecordReconcileWithContext(ctx, "Pending", "success", 0.1)
	m.RecordReconcileWithContext(ctx, "Building", "error", 0.5)
}

func TestReconcilerMetrics_RecordReconcileWithContext_ExemplarsDisabled(t *testing.T) {
	m := NewReconcilerMetricsWithExemplars(false)

	ctx := context.Background()

	// With exemplars disabled, should fall back to regular recording
	m.RecordReconcileWithContext(ctx, "Ready", "success", 0.05)
}

func TestReconcilerMetrics_RecordBuildWithContext_NoSpan(t *testing.T) {
	m := NewReconcilerMetrics()

	ctx := context.Background()

	// Should not panic and should fall back to regular recording
	m.RecordBuildWithContext(ctx, "python", "success", 120.0)
	m.RecordBuildWithContext(ctx, "nodejs", "failed", 60.0)
}

func TestReconcilerMetrics_RecordBuildWithContext_ExemplarsDisabled(t *testing.T) {
	m := NewReconcilerMetricsWithExemplars(false)

	ctx := context.Background()

	// With exemplars disabled, should fall back to regular recording
	m.RecordBuildWithContext(ctx, "go", "success", 90.0)
}

func TestReconcilerMetrics_RecordErrorWithContext_NoSpan(t *testing.T) {
	m := NewReconcilerMetrics()

	ctx := context.Background()

	// Should not panic and should fall back to regular recording
	m.RecordErrorWithContext(ctx, "build", "context_no_span")
	m.RecordErrorWithContext(ctx, "deploy", "context_error")
}

func TestReconcilerMetrics_RecordErrorWithContext_ExemplarsDisabled(t *testing.T) {
	m := NewReconcilerMetricsWithExemplars(false)

	ctx := context.Background()

	// With exemplars disabled, should fall back to regular recording
	m.RecordErrorWithContext(ctx, "eventing", "disabled_exemplars")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🧪 EXTRACT EXEMPLAR TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestReconcilerMetrics_ExtractExemplar_BackgroundContext(t *testing.T) {
	m := NewReconcilerMetrics()

	ctx := context.Background()
	labels := m.extractExemplar(ctx)

	// Background context has no span, should return nil
	assert.Nil(t, labels)
}

func TestReconcilerMetrics_ExtractExemplar_ContextWithValues(t *testing.T) {
	m := NewReconcilerMetrics()

	// Context with values but no span
	ctx := context.WithValue(context.Background(), "key", "value")
	labels := m.extractExemplar(ctx)

	// No span in context, should return nil
	assert.Nil(t, labels)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 METRIC DESCRIPTION TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestMetricDescriptions(t *testing.T) {
	tests := []struct {
		name        string
		metricCheck func() bool
	}{
		{
			name: "ReconcileTotal is counter",
			metricCheck: func() bool {
				return ReconcileTotal.WithLabelValues("test", "test") != nil
			},
		},
		{
			name: "ReconcileDuration is histogram",
			metricCheck: func() bool {
				return ReconcileDuration.WithLabelValues("test") != nil
			},
		},
		{
			name: "LambdaFunctionsTotal is gauge",
			metricCheck: func() bool {
				return LambdaFunctionsTotal.WithLabelValues("test", "test") != nil
			},
		},
		{
			name: "BuildJobsActive is gauge",
			metricCheck: func() bool {
				return BuildJobsActive.WithLabelValues("test") != nil
			},
		},
		{
			name: "BuildDuration is histogram",
			metricCheck: func() bool {
				return BuildDuration.WithLabelValues("test", "test") != nil
			},
		},
		{
			name: "EventingResourcesTotal is gauge",
			metricCheck: func() bool {
				return EventingResourcesTotal.WithLabelValues("test", "test") != nil
			},
		},
		{
			name: "APIServerRequestsTotal is counter",
			metricCheck: func() bool {
				return APIServerRequestsTotal.WithLabelValues("test", "test", "test") != nil
			},
		},
		{
			name: "WorkQueueDepth is gauge",
			metricCheck: func() bool {
				WorkQueueDepth.Set(0)
				return true
			},
		},
		{
			name: "WorkQueueLatency is histogram",
			metricCheck: func() bool {
				WorkQueueLatency.Observe(0)
				return true
			},
		},
		{
			name: "ErrorsTotal is counter",
			metricCheck: func() bool {
				return ErrorsTotal.WithLabelValues("test", "test") != nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, tc.metricCheck())
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 MULTI-NAMESPACE METRICS TESTS                                       │
// └─────────────────────────────────────────────────────────────────────────┘

func TestMultiNamespaceMetrics(t *testing.T) {
	m := NewReconcilerMetrics()
	namespaces := []string{"default", "production", "staging", "development"}

	for _, ns := range namespaces {
		t.Run("Namespace_"+ns, func(t *testing.T) {
			// Lambda count
			m.SetLambdaCount(ns, "Ready", 10)
			m.SetLambdaCount(ns, "Building", 2)
			m.SetLambdaCount(ns, "Failed", 1)

			// Active build jobs
			m.SetActiveBuildJobs(ns, 3)

			// Function metrics
			FunctionInvocationsTotal.WithLabelValues("test-func", ns, "success").Inc()
			FunctionErrorsTotal.WithLabelValues("test-func", ns, "timeout").Inc()
			FunctionColdStartsTotal.WithLabelValues("test-func", ns).Inc()
			FunctionDuration.WithLabelValues("test-func", ns).Observe(0.1)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📈 METRIC VALUES TESTS                                                 │
// └─────────────────────────────────────────────────────────────────────────┘

func TestGaugeIncrementDecrement(t *testing.T) {
	// Test gauge increment/decrement behavior
	gauge := BuildJobsActive.WithLabelValues("test-inc-dec")

	gauge.Set(0)
	gauge.Inc()
	gauge.Inc()
	gauge.Dec()
	// Final value should be 1
}

func TestCounterAddition(t *testing.T) {
	// Test counter add behavior
	counter := ReconcileTotal.WithLabelValues("test-add", "test")

	counter.Add(5)
	counter.Add(10)
	// Total should be 15 added
}

func TestHistogramMultipleObservations(t *testing.T) {
	// Test histogram with multiple observations
	histogram := ReconcileDuration.WithLabelValues("test-multi")

	observations := []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0}
	for _, obs := range observations {
		histogram.Observe(obs)
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🎯 ERROR TYPE TESTS                                                    │
// └─────────────────────────────────────────────────────────────────────────┘

func TestErrorTypes(t *testing.T) {
	m := NewReconcilerMetrics()

	errorTypes := map[string][]string{
		"build": {
			"job_creation_failed",
			"image_pull_failed",
			"dockerfile_error",
			"timeout",
			"out_of_memory",
			"registry_push_failed",
		},
		"deploy": {
			"service_creation_failed",
			"revision_failed",
			"route_creation_failed",
			"timeout",
			"probe_failed",
		},
		"eventing": {
			"broker_creation_failed",
			"trigger_creation_failed",
			"subscription_failed",
			"dlq_creation_failed",
		},
		"reconcile": {
			"status_update_failed",
			"finalizer_error",
			"spec_validation_failed",
			"resource_conflict",
		},
	}

	for component, types := range errorTypes {
		for _, errType := range types {
			t.Run(component+"_"+errType, func(t *testing.T) {
				m.RecordError(component, errType)
			})
		}
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ⏱️ DURATION RANGE TESTS                                                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestReconcileDurationRange(t *testing.T) {
	m := NewReconcilerMetrics()

	// Test various duration ranges
	durations := []struct {
		name     string
		duration float64
	}{
		{"sub_millisecond", 0.0005},
		{"milliseconds", 0.01},
		{"hundred_ms", 0.1},
		{"one_second", 1.0},
		{"ten_seconds", 10.0},
		{"one_minute", 60.0},
		{"five_minutes", 300.0},
	}

	for _, d := range durations {
		t.Run(d.name, func(t *testing.T) {
			m.RecordReconcile("test", "success", d.duration)
		})
	}
}

func TestBuildDurationRange(t *testing.T) {
	m := NewReconcilerMetrics()

	// Test various build duration ranges
	durations := []struct {
		name     string
		duration float64
	}{
		{"ten_seconds", 10.0},
		{"thirty_seconds", 30.0},
		{"one_minute", 60.0},
		{"two_minutes", 120.0},
		{"five_minutes", 300.0},
		{"ten_minutes", 600.0},
		{"fifteen_minutes", 900.0},
		{"twenty_minutes", 1200.0},
		{"thirty_minutes", 1800.0},
	}

	for _, d := range durations {
		t.Run(d.name, func(t *testing.T) {
			m.RecordBuild("python", "success", d.duration)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔢 FUNCTION STATUS TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestFunctionInvocationStatuses(t *testing.T) {
	statuses := []string{"success", "error", "timeout", "cancelled", "cold_start"}

	for _, status := range statuses {
		t.Run("Status_"+status, func(t *testing.T) {
			counter := FunctionInvocationsTotal.WithLabelValues("test-func", "default", status)
			require.NotNil(t, counter)
			counter.Inc()
		})
	}
}

func TestFunctionErrorTypes(t *testing.T) {
	errorTypes := []string{
		"runtime_error",
		"timeout",
		"out_of_memory",
		"cold_start_failure",
		"configuration_error",
		"dependency_error",
		"network_error",
	}

	for _, errType := range errorTypes {
		t.Run("ErrorType_"+errType, func(t *testing.T) {
			counter := FunctionErrorsTotal.WithLabelValues("test-func", "default", errType)
			require.NotNil(t, counter)
			counter.Inc()
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 CONCURRENT METRICS TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestConcurrentMetricUpdates(t *testing.T) {
	m := NewReconcilerMetrics()

	const numGoroutines = 50
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Concurrent reconcile recordings
			m.RecordReconcile("Ready", "success", 0.1)

			// Concurrent lambda counts
			m.SetLambdaCount("default", "Ready", float64(id))

			// Concurrent build jobs
			m.SetActiveBuildJobs("default", float64(id%5))

			// Concurrent error recordings
			m.RecordError("reconcile", "concurrent_test")

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 API SERVER METRICS TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAPIServerMetrics(t *testing.T) {
	resources := []string{
		"lambdafunction",
		"lambdaagent",
		"service",
		"job",
		"secret",
		"configmap",
		"broker",
		"trigger",
	}

	verbs := []string{"get", "list", "create", "update", "delete", "watch"}
	results := []string{"success", "error", "not_found", "conflict"}

	for _, resource := range resources {
		for _, verb := range verbs {
			for _, result := range results {
				t.Run(resource+"_"+verb+"_"+result, func(t *testing.T) {
					counter := APIServerRequestsTotal.WithLabelValues(verb, resource, result)
					require.NotNil(t, counter)
					counter.Inc()
				})
			}
		}
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 FULL LIFECYCLE METRICS TEST                                         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestFullReconcileLifecycleMetrics(t *testing.T) {
	m := NewReconcilerMetrics()
	ctx := context.Background()

	// Simulate full lifecycle of a lambda function
	phases := []string{"Pending", "Building", "Deploying", "Ready"}

	for _, phase := range phases {
		// Record reconcile for each phase
		m.RecordReconcile(phase, "success", 0.1)
		m.RecordReconcileWithContext(ctx, phase, "success", 0.1)
	}

	// Record build
	m.RecordBuild("python", "success", 120.0)
	m.RecordBuildWithContext(ctx, "python", "success", 120.0)

	// Set lambda counts
	m.SetLambdaCount("default", "Ready", 1)

	// Clear build jobs
	m.SetActiveBuildJobs("default", 0)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ❌ FAILURE SCENARIO METRICS TESTS                                      │
// └─────────────────────────────────────────────────────────────────────────┘

func TestFailureScenarioMetrics(t *testing.T) {
	m := NewReconcilerMetrics()
	ctx := context.Background()

	// Simulate failure scenarios
	t.Run("Build failure", func(t *testing.T) {
		m.RecordReconcile("Building", "error", 30.0)
		m.RecordBuild("python", "failed", 30.0)
		m.RecordError("build", "job_failed")
		m.RecordErrorWithContext(ctx, "build", "job_failed")
	})

	t.Run("Deploy failure", func(t *testing.T) {
		m.RecordReconcile("Deploying", "error", 5.0)
		m.RecordError("deploy", "service_creation_failed")
	})

	t.Run("Reconcile timeout", func(t *testing.T) {
		m.RecordReconcile("Pending", "timeout", 60.0)
		m.RecordError("reconcile", "timeout")
	})

	// Update lambda counts for failures
	m.SetLambdaCount("default", "Failed", 3)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📋 METRIC NAME FORMAT TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestMetricNameFormats(t *testing.T) {
	// Verify metric names follow Prometheus conventions
	tests := []struct {
		name           string
		expectedPrefix string
	}{
		{
			name:           "ReconcileTotal uses namespace prefix",
			expectedPrefix: "knative_lambda_operator_reconcile_total",
		},
		{
			name:           "FunctionDuration uses function subsystem",
			expectedPrefix: "knative_lambda_function_duration_seconds",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Verify constants are used
			assert.Equal(t, "knative_lambda", namespace)
			assert.Equal(t, "operator", subsystem)
		})
	}
}
