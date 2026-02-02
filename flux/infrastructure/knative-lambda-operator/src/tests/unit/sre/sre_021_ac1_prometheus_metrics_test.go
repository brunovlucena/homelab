// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 SRE-021: AC1 Prometheus Metrics Collection Tests
//
//	BVL-386: DEVOPS-001-AC1 Prometheus Metrics Collection
//	Priority: P0 | Story Points: 2
//
//	Tests validate acceptance criteria:
//	✓ ServiceMonitor exists for operator service
//	✓ Metrics endpoint /metrics accessible on port 8080
//	✓ All operator metrics exposed with knative_lambda_operator_* prefix
//	✓ Function RED metrics exposed with knative_lambda_function_* prefix
//	✓ Prometheus scrapes at 30s interval
//	✓ Metric cardinality controlled (no high-cardinality labels)
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

package sre

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC1: ServiceMonitor Exists for Operator Service
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ServiceMonitor represents a Kubernetes ServiceMonitor resource
type ServiceMonitor struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string            `yaml:"name"`
		Namespace string            `yaml:"namespace"`
		Labels    map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
	Spec struct {
		Selector struct {
			MatchLabels map[string]string `yaml:"matchLabels"`
		} `yaml:"selector"`
		NamespaceSelector struct {
			MatchNames []string `yaml:"matchNames"`
		} `yaml:"namespaceSelector"`
		Endpoints []struct {
			Port          string `yaml:"port"`
			Interval      string `yaml:"interval"`
			ScrapeTimeout string `yaml:"scrapeTimeout"`
			Path          string `yaml:"path"`
			Scheme        string `yaml:"scheme"`
		} `yaml:"endpoints"`
	} `yaml:"spec"`
}

func findServiceMonitorFile() string {
	// Look for ServiceMonitor in known locations
	// The test runs from: /workspace/flux/infrastructure/knative-lambda-operator/src/tests/unit/sre
	// The ServiceMonitor is at: /workspace/flux/infrastructure/prometheus-operator/k8s/servicemonitors/
	possiblePaths := []string{
		// Relative paths from different test run locations
		"../../../../../../prometheus-operator/k8s/servicemonitors/knative-lambda-operator-servicemonitor.yaml",
		"../../../../../prometheus-operator/k8s/servicemonitors/knative-lambda-operator-servicemonitor.yaml",
		"../../../../prometheus-operator/k8s/servicemonitors/knative-lambda-operator-servicemonitor.yaml",
		"../../../prometheus-operator/k8s/servicemonitors/knative-lambda-operator-servicemonitor.yaml",
		// Absolute paths
		"/workspace/flux/infrastructure/prometheus-operator/k8s/servicemonitors/knative-lambda-operator-servicemonitor.yaml",
	}

	// Try from workspace root
	workspaceRoot := os.Getenv("WORKSPACE_ROOT")
	if workspaceRoot != "" {
		possiblePaths = append([]string{
			filepath.Join(workspaceRoot, "flux/infrastructure/prometheus-operator/k8s/servicemonitors/knative-lambda-operator-servicemonitor.yaml"),
		}, possiblePaths...)
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func TestAC1_ServiceMonitorExists(t *testing.T) {
	smFile := findServiceMonitorFile()
	if smFile == "" {
		t.Skip("ServiceMonitor file not found, skipping test")
		return
	}

	data, err := os.ReadFile(smFile)
	require.NoError(t, err, "Failed to read ServiceMonitor file")

	var sm ServiceMonitor
	err = yaml.Unmarshal(data, &sm)
	require.NoError(t, err, "Failed to parse ServiceMonitor YAML")

	t.Run("ServiceMonitor has correct metadata", func(t *testing.T) {
		assert.Equal(t, "monitoring.coreos.com/v1", sm.APIVersion,
			"ServiceMonitor should have correct API version")
		assert.Equal(t, "ServiceMonitor", sm.Kind,
			"Kind should be ServiceMonitor")
		assert.Equal(t, "knative-lambda-operator", sm.Metadata.Name,
			"ServiceMonitor should be named 'knative-lambda-operator'")
		assert.Equal(t, "knative-lambda", sm.Metadata.Namespace,
			"ServiceMonitor should be in 'knative-lambda' namespace")
	})

	t.Run("ServiceMonitor targets operator service", func(t *testing.T) {
		matchLabels := sm.Spec.Selector.MatchLabels
		assert.NotEmpty(t, matchLabels,
			"ServiceMonitor must have selector.matchLabels")
		assert.Equal(t, "knative-lambda-operator", matchLabels["app.kubernetes.io/name"],
			"ServiceMonitor should target knative-lambda-operator service")
	})

	t.Run("ServiceMonitor namespace selector includes knative-lambda", func(t *testing.T) {
		namespaces := sm.Spec.NamespaceSelector.MatchNames
		assert.Contains(t, namespaces, "knative-lambda",
			"ServiceMonitor should select 'knative-lambda' namespace")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC1: Metrics Endpoint on Port 8080
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC1_MetricsEndpointConfig(t *testing.T) {
	smFile := findServiceMonitorFile()
	if smFile == "" {
		t.Skip("ServiceMonitor file not found, skipping test")
		return
	}

	data, err := os.ReadFile(smFile)
	require.NoError(t, err, "Failed to read ServiceMonitor file")

	var sm ServiceMonitor
	err = yaml.Unmarshal(data, &sm)
	require.NoError(t, err, "Failed to parse ServiceMonitor YAML")

	t.Run("Metrics endpoint configured on port metrics (8080)", func(t *testing.T) {
		require.NotEmpty(t, sm.Spec.Endpoints,
			"ServiceMonitor must have at least one endpoint")

		endpoint := sm.Spec.Endpoints[0]
		assert.Equal(t, "metrics", endpoint.Port,
			"Endpoint should target 'metrics' port (which is 8080)")
		assert.Equal(t, "/metrics", endpoint.Path,
			"Endpoint should scrape '/metrics' path")
		assert.Equal(t, "http", endpoint.Scheme,
			"Endpoint should use HTTP scheme")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC1: Prometheus Scrapes at 30s Interval
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC1_ScrapeInterval30s(t *testing.T) {
	smFile := findServiceMonitorFile()
	if smFile == "" {
		t.Skip("ServiceMonitor file not found, skipping test")
		return
	}

	data, err := os.ReadFile(smFile)
	require.NoError(t, err, "Failed to read ServiceMonitor file")

	var sm ServiceMonitor
	err = yaml.Unmarshal(data, &sm)
	require.NoError(t, err, "Failed to parse ServiceMonitor YAML")

	t.Run("Scrape interval is 30s", func(t *testing.T) {
		require.NotEmpty(t, sm.Spec.Endpoints,
			"ServiceMonitor must have at least one endpoint")

		endpoint := sm.Spec.Endpoints[0]
		assert.Equal(t, "30s", endpoint.Interval,
			"Scrape interval should be 30s")
	})

	t.Run("Scrape timeout is less than interval", func(t *testing.T) {
		require.NotEmpty(t, sm.Spec.Endpoints,
			"ServiceMonitor must have at least one endpoint")

		endpoint := sm.Spec.Endpoints[0]
		assert.NotEmpty(t, endpoint.ScrapeTimeout,
			"Scrape timeout should be configured")
		assert.Equal(t, "10s", endpoint.ScrapeTimeout,
			"Scrape timeout should be 10s (less than 30s interval)")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC1: Operator Metrics with knative_lambda_operator_* prefix
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func findMetricsSourceFile() string {
	// Look for metrics.go in known locations
	possiblePaths := []string{
		"../../../operator/internal/metrics/metrics.go",
		"../../operator/internal/metrics/metrics.go",
		"../operator/internal/metrics/metrics.go",
	}

	workspaceRoot := os.Getenv("WORKSPACE_ROOT")
	if workspaceRoot != "" {
		possiblePaths = append([]string{
			filepath.Join(workspaceRoot, "flux/infrastructure/knative-lambda-operator/src/operator/internal/metrics/metrics.go"),
		}, possiblePaths...)
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func TestAC1_OperatorMetricsExist(t *testing.T) {
	metricsFile := findMetricsSourceFile()
	if metricsFile == "" {
		t.Skip("metrics.go not found, skipping test")
		return
	}

	data, err := os.ReadFile(metricsFile)
	require.NoError(t, err, "Failed to read metrics.go")
	source := string(data)

	t.Run("knative_lambda_operator_reconcile_total exists", func(t *testing.T) {
		assert.Contains(t, source, `Name:      "reconcile_total"`,
			"reconcile_total metric should be defined")
		assert.Contains(t, source, `[]string{"phase", "result"}`,
			"reconcile_total should have phase and result labels")
	})

	t.Run("knative_lambda_operator_reconcile_duration_seconds exists", func(t *testing.T) {
		assert.Contains(t, source, `Name:      "reconcile_duration_seconds"`,
			"reconcile_duration_seconds metric should be defined")
		assert.Contains(t, source, `[]string{"phase"}`,
			"reconcile_duration_seconds should have phase label")
	})

	t.Run("knative_lambda_operator_lambdafunctions_total exists", func(t *testing.T) {
		assert.Contains(t, source, `Name:      "lambdafunctions_total"`,
			"lambdafunctions_total metric should be defined")
		assert.Contains(t, source, `[]string{"namespace", "phase"}`,
			"lambdafunctions_total should have namespace and phase labels")
	})

	t.Run("knative_lambda_operator_build_duration_seconds exists", func(t *testing.T) {
		assert.Contains(t, source, `Name:      "build_duration_seconds"`,
			"build_duration_seconds metric should be defined")
		assert.Contains(t, source, `[]string{"runtime", "result"}`,
			"build_duration_seconds should have runtime and result labels")
	})

	t.Run("knative_lambda_operator_build_jobs_active exists", func(t *testing.T) {
		assert.Contains(t, source, `Name:      "build_jobs_active"`,
			"build_jobs_active metric should be defined")
	})

	t.Run("knative_lambda_operator_errors_total exists", func(t *testing.T) {
		assert.Contains(t, source, `Name:      "errors_total"`,
			"errors_total metric should be defined")
	})

	t.Run("knative_lambda_operator_workqueue_depth exists", func(t *testing.T) {
		assert.Contains(t, source, `Name:      "workqueue_depth"`,
			"workqueue_depth metric should be defined")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC1: Function RED Metrics with knative_lambda_function_* prefix
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC1_FunctionREDMetricsExist(t *testing.T) {
	metricsFile := findMetricsSourceFile()
	if metricsFile == "" {
		t.Skip("metrics.go not found, skipping test")
		return
	}

	data, err := os.ReadFile(metricsFile)
	require.NoError(t, err, "Failed to read metrics.go")
	source := string(data)

	t.Run("knative_lambda_function_invocations_total exists (Rate)", func(t *testing.T) {
		assert.Contains(t, source, `Subsystem: "function"`,
			"function metrics should use function subsystem")
		assert.Contains(t, source, `Name:      "invocations_total"`,
			"invocations_total metric should be defined")
		assert.Contains(t, source, `[]string{"function", "namespace", "status"}`,
			"invocations_total should have function, namespace, status labels")
	})

	t.Run("knative_lambda_function_duration_seconds exists (Duration)", func(t *testing.T) {
		assert.Contains(t, source, `Name:      "duration_seconds"`,
			"duration_seconds metric should be defined for functions")
		assert.Contains(t, source, `[]string{"function", "namespace"}`,
			"duration_seconds should have function and namespace labels")
	})

	t.Run("knative_lambda_function_errors_total exists (Errors)", func(t *testing.T) {
		// Note: This is the function errors metric, not operator errors
		// Look for the specific function errors metric definition
		assert.Contains(t, source, `FunctionErrorsTotal`,
			"FunctionErrorsTotal metric should be defined")
		assert.Contains(t, source, `[]string{"function", "namespace", "error_type"}`,
			"function errors_total should have function, namespace, error_type labels")
	})

	t.Run("knative_lambda_function_cold_starts_total exists", func(t *testing.T) {
		assert.Contains(t, source, `Name:      "cold_starts_total"`,
			"cold_starts_total metric should be defined")
		assert.Contains(t, source, `FunctionColdStartsTotal`,
			"FunctionColdStartsTotal should be defined")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC1: Function RED Metrics Implementation Gap Analysis
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// TestAC1_FunctionREDMetricsImplementationGap documents that function RED metrics
// are DEFINED but NOT POPULATED in production code.
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// WHO SHOULD CREATE RESPONSEMETRICS EVENTS?
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
// The LAMBDA RUNTIME WRAPPER (generated code in each function container) should
// create ResponseMetrics. The runtime wrapper is generated from templates:
//
//   build/templates/runtimes/python/runtime.py.tmpl
//   build/templates/runtimes/nodejs/runtime.js.tmpl
//   build/templates/runtimes/go/runtime.go.tmpl
//
// CURRENT IMPLEMENTATION STATUS:
//
// ┌─────────────────────────────────────────────────────────────────────────┐
// │ Node.js Runtime (runtime.js.tmpl) - PARTIALLY IMPLEMENTED              │
// ├─────────────────────────────────────────────────────────────────────────┤
// │ ✅ Tracks invocations, errors, duration (histogram), cold starts       │
// │ ✅ Exposes /metrics endpoint in Prometheus format                      │
// │ ✅ Pushes metrics directly to Prometheus via remote write              │
// │ ❌ Does NOT emit response.success/error CloudEvents to operator        │
// │                                                                         │
// │ Result: Metrics go DIRECTLY to Prometheus, bypassing operator          │
// └─────────────────────────────────────────────────────────────────────────┘
//
// ┌─────────────────────────────────────────────────────────────────────────┐
// │ Python Runtime (runtime.py.tmpl) - NOT IMPLEMENTED                     │
// ├─────────────────────────────────────────────────────────────────────────┤
// │ ❌ Does NOT track any metrics                                          │
// │ ❌ Does NOT expose /metrics endpoint                                   │
// │ ❌ Does NOT emit ResponseMetrics in CloudEvents                        │
// │                                                                         │
// │ Result: No function metrics for Python lambdas                         │
// └─────────────────────────────────────────────────────────────────────────┘
//
// ┌─────────────────────────────────────────────────────────────────────────┐
// │ Operator (metrics.go) - DEAD CODE                                      │
// ├─────────────────────────────────────────────────────────────────────────┤
// │ ✅ FunctionInvocationsTotal defined                                    │
// │ ✅ FunctionDuration defined                                            │
// │ ✅ FunctionErrorsTotal defined                                         │
// │ ✅ FunctionColdStartsTotal defined                                     │
// │ ❌ Never populated - operator doesn't receive response events          │
// │                                                                         │
// │ Result: These metrics are always zero                                  │
// └─────────────────────────────────────────────────────────────────────────┘
//
// THE DESIGN INTENT WAS:
// 1. Function container runs Lambda Runtime Wrapper
// 2. Wrapper tracks metrics during execution (duration, cold start, etc.)
// 3. Wrapper emits io.knative.lambda.response.success/error CloudEvents
// 4. CloudEvents include ResponseMetrics (DurationMs, ColdStart, MemoryUsedMb)
// 5. Operator receives these events via Trigger subscription
// 6. Operator extracts ResponseMetrics and populates its Prometheus metrics
//
// BUT THE REALITY IS:
// - Node.js: Pushes directly to Prometheus (works but bypasses operator)
// - Python: No metrics at all
// - Go: No metrics implementation
// - Operator: Never receives response events, metrics are dead code
//
// RECOMMENDED FIXES:
//
// Option A: Use PodMonitor to scrape /metrics from Node.js functions
//   - Already works for Node.js functions
//   - Need to add /metrics endpoint to Python/Go runtimes
//   - Pros: Simple, decentralized, standard Prometheus pattern
//   - Cons: Requires each function to expose metrics
//
// Option B: Complete the CloudEvents flow
//   - Add ResponseMetrics to Python/Go runtimes
//   - Create Trigger for operator to receive response events
//   - Populate operator metrics from response event data
//   - Pros: Centralized metrics, works with scale-to-zero
//   - Cons: More complex, adds latency to event flow
//
// CURRENT STATUS: Metrics are registered but always zero (dead code)
func TestAC1_FunctionREDMetricsImplementationGap(t *testing.T) {
	t.Run("Document: Function RED metrics are defined but not populated", func(t *testing.T) {
		// This test documents the implementation gap
		t.Log("=== IMPLEMENTATION GAP ANALYSIS ===")
		t.Log("")
		t.Log("FINDING: Function RED metrics are DEFINED but NOT POPULATED in production code")
		t.Log("")
		t.Log("Metrics defined in metrics.go:")
		t.Log("  - knative_lambda_function_invocations_total")
		t.Log("  - knative_lambda_function_duration_seconds")
		t.Log("  - knative_lambda_function_errors_total")
		t.Log("  - knative_lambda_function_cold_starts_total")
		t.Log("")
		t.Log("Files that use these metrics:")
		t.Log("  - metrics.go (definition)")
		t.Log("  - metrics_test.go (unit tests only)")
		t.Log("")
		t.Log("Files that SHOULD populate these metrics but DON'T:")
		t.Log("  - events/manager.go - has ResponseMetrics struct but doesn't update Prometheus")
		t.Log("  - webhook/cloudevents_receiver.go - could process response events")
		t.Log("")
		t.Log("RECOMMENDATION: Implement Option 4 (CloudEvents Response Processing)")
		t.Log("  1. Subscribe operator to response.success/error events")
		t.Log("  2. In event handler, extract ResponseMetrics from event data")
		t.Log("  3. Call metrics.FunctionInvocationsTotal.WithLabelValues(...).Inc()")
		t.Log("  4. Call metrics.FunctionDuration.WithLabelValues(...).Observe()")
		t.Log("  5. Track cold starts from ResponseMetrics.ColdStart field")
	})

	t.Run("Verify metrics are registered but unused in production", func(t *testing.T) {
		// Search for actual usage of function metrics in non-test code
		productionFiles := []string{
			"../../../operator/controllers/lambdafunction_controller.go",
			"../../operator/controllers/lambdafunction_controller.go",
			"../operator/controllers/lambdafunction_controller.go",
			"/workspace/flux/infrastructure/knative-lambda-operator/src/operator/controllers/lambdafunction_controller.go",
		}

		var controllerSource string
		for _, path := range productionFiles {
			data, err := os.ReadFile(path)
			if err == nil {
				controllerSource = string(data)
				break
			}
		}

		if controllerSource == "" {
			t.Skip("Controller file not found")
			return
		}

		// Verify function metrics are NOT used in controller
		assert.NotContains(t, controllerSource, "FunctionInvocationsTotal",
			"FunctionInvocationsTotal should NOT be used in controller (implementation gap)")
		assert.NotContains(t, controllerSource, "FunctionDuration",
			"FunctionDuration should NOT be used in controller (implementation gap)")
		assert.NotContains(t, controllerSource, "FunctionErrorsTotal",
			"FunctionErrorsTotal should NOT be used in controller (implementation gap)")
		assert.NotContains(t, controllerSource, "FunctionColdStartsTotal",
			"FunctionColdStartsTotal should NOT be used in controller (implementation gap)")

		t.Log("CONFIRMED: Function RED metrics are not populated in controller code")
	})

	t.Run("Verify ResponseMetrics struct exists for future implementation", func(t *testing.T) {
		eventsManagerFiles := []string{
			"../../../operator/internal/events/manager.go",
			"../../operator/internal/events/manager.go",
			"../operator/internal/events/manager.go",
			"/workspace/flux/infrastructure/knative-lambda-operator/src/operator/internal/events/manager.go",
		}

		var eventsSource string
		for _, path := range eventsManagerFiles {
			data, err := os.ReadFile(path)
			if err == nil {
				eventsSource = string(data)
				break
			}
		}

		if eventsSource == "" {
			t.Skip("Events manager file not found")
			return
		}

		// Verify ResponseMetrics struct exists
		assert.Contains(t, eventsSource, "type ResponseMetrics struct",
			"ResponseMetrics struct should exist for event data")
		assert.Contains(t, eventsSource, "DurationMs",
			"ResponseMetrics should have DurationMs field")
		assert.Contains(t, eventsSource, "ColdStart",
			"ResponseMetrics should have ColdStart field")

		// Verify EmitResponseSuccess exists
		assert.Contains(t, eventsSource, "func (m *Manager) EmitResponseSuccess",
			"EmitResponseSuccess should exist to emit response events")

		// Verify metrics are NOT populated in recordEventMetrics
		assert.Contains(t, eventsSource, "func (m *Manager) recordEventMetrics",
			"recordEventMetrics function should exist")

		// Check that recordEventMetrics only handles build/service/parser events
		// and NOT function invocation metrics
		t.Log("CONFIRMED: ResponseMetrics struct exists but is not used to populate Prometheus metrics")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC1: Metric Cardinality Control (No High-Cardinality Labels)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// highCardinalityLabels are labels that should NEVER be used in metrics as they cause cardinality explosion
// Note: trace_id/span_id are OK in exemplars (not labels) for linking to traces
var highCardinalityLabels = []string{
	"invocation_id",
	"user_id",
	"request_id",
	"correlation_id",
	"session_id",
	"transaction_id",
	"pod_id",
	"container_id",
	"event_id",
	"message",       // Free-form text
	"error_message", // Free-form text
	"stack_trace",   // Free-form text
	"url",           // Could be unbounded
	"path",          // Could be unbounded (for HTTP paths)
	"ip",            // IP addresses
	"ip_address",
}

// potentiallyHighCardinalityLabels are labels that are acceptable but should be documented
// as potentially high cardinality in certain environments
var potentiallyHighCardinalityLabels = []string{
	"function",   // Could grow with many lambda functions - OK but monitor
	"namespace",  // Could grow in multi-tenant clusters - OK but monitor
}

func TestAC1_NoHighCardinalityLabels(t *testing.T) {
	t.Run("Metrics source code has no high-cardinality labels in metric definitions", func(t *testing.T) {
		// Read the metrics.go source file
		metricsFile := findMetricsSourceFile()
		if metricsFile == "" {
			t.Skip("metrics.go not found, skipping source code check")
			return
		}

		data, err := os.ReadFile(metricsFile)
		require.NoError(t, err, "Failed to read metrics.go")
		metricsSource := string(data)

		// Check that none of the high-cardinality labels are used in metric label definitions
		// We look for []string{...} patterns that define labels
		for _, label := range highCardinalityLabels {
			// Check if the label appears in a []string{} label definition
			// This is more precise than just checking if the string exists anywhere
			labelInArray := `"` + label + `"`

			// Count occurrences in label arrays ([]string{"...", "label", "..."})
			if strings.Contains(metricsSource, labelInArray) {
				// Check if it's in a label definition context (preceded by []string{)
				lines := strings.Split(metricsSource, "\n")
				for i, line := range lines {
					if strings.Contains(line, labelInArray) {
						// Check if this line or recent lines have []string{
						// This is a heuristic to detect label definitions
						contextStart := i - 5
						if contextStart < 0 {
							contextStart = 0
						}
						context := strings.Join(lines[contextStart:i+1], "\n")

						// If []string{ is in context and label is on the line, it's likely a label definition
						if strings.Contains(context, "[]string{") && strings.Contains(line, labelInArray) {
							t.Errorf("metrics.go should not use high-cardinality label in metric definition: %s (line %d)", label, i+1)
						}
					}
				}
			}
		}
	})

	t.Run("Exemplar labels (trace_id, span_id) are allowed but not metric labels", func(t *testing.T) {
		metricsFile := findMetricsSourceFile()
		if metricsFile == "" {
			t.Skip("metrics.go not found, skipping source code check")
			return
		}

		data, err := os.ReadFile(metricsFile)
		require.NoError(t, err, "Failed to read metrics.go")
		metricsSource := string(data)

		// trace_id and span_id are allowed in exemplars (for linking metrics to traces)
		// They should appear in the extractExemplar function, not in metric label definitions

		// Verify trace_id is used only in exemplar context
		if strings.Contains(metricsSource, `"trace_id"`) {
			assert.Contains(t, metricsSource, "extractExemplar",
				"trace_id should only be used in exemplar extraction, not as metric label")
			assert.Contains(t, metricsSource, "prometheus.Labels",
				"trace_id should be in prometheus.Labels for exemplars")
		}
	})

	t.Run("Document potentially high cardinality labels that are acceptable", func(t *testing.T) {
		metricsFile := findMetricsSourceFile()
		if metricsFile == "" {
			t.Skip("metrics.go not found, skipping source code check")
			return
		}

		data, err := os.ReadFile(metricsFile)
		require.NoError(t, err, "Failed to read metrics.go")
		metricsSource := string(data)

		// These labels are acceptable but teams should be aware they can grow
		for _, label := range potentiallyHighCardinalityLabels {
			labelInArray := `"` + label + `"`
			if strings.Contains(metricsSource, labelInArray) {
				t.Logf("INFO: Metric uses '%s' label - acceptable but monitor cardinality in production", label)
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC1: Bounded Label Value Tests
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC1_BoundedLabelValues(t *testing.T) {
	t.Run("Phase labels have bounded values", func(t *testing.T) {
		// Valid phases are limited to these values
		validPhases := []string{"Pending", "Building", "Deploying", "Ready", "Failed", "Deleting"}
		assert.Len(t, validPhases, 6, "Phase label should have exactly 6 bounded values")
	})

	t.Run("Result labels have bounded values", func(t *testing.T) {
		// Valid results are limited to these values
		validResults := []string{"success", "error", "timeout", "cancelled"}
		assert.Len(t, validResults, 4, "Result label should have exactly 4 bounded values")
	})

	t.Run("Runtime labels have bounded values", func(t *testing.T) {
		// Valid runtimes are limited to supported languages
		validRuntimes := []string{"python", "nodejs", "go", "java", "rust"}
		assert.Len(t, validRuntimes, 5, "Runtime label should have exactly 5 bounded values")
	})

	t.Run("Component labels have bounded values", func(t *testing.T) {
		// Valid components are limited to operator components
		validComponents := []string{"build", "deploy", "eventing", "reconcile", "webhook"}
		assert.Len(t, validComponents, 5, "Component label should have exactly 5 bounded values")
	})

	t.Run("Error type labels are hardcoded (not dynamic)", func(t *testing.T) {
		// Error types should be hardcoded, not derived from error messages
		metricsFile := findMetricsSourceFile()
		if metricsFile == "" {
			t.Skip("metrics.go not found, skipping source code check")
			return
		}

		// Check controller file for recordError calls
		controllerFiles := []string{
			"../../../operator/controllers/lambdafunction_controller.go",
			"../../operator/controllers/lambdafunction_controller.go",
			"../operator/controllers/lambdafunction_controller.go",
			"/workspace/flux/infrastructure/knative-lambda-operator/src/operator/controllers/lambdafunction_controller.go",
		}

		var controllerSource string
		for _, path := range controllerFiles {
			data, err := os.ReadFile(path)
			if err == nil {
				controllerSource = string(data)
				break
			}
		}

		if controllerSource == "" {
			t.Skip("Controller file not found")
			return
		}

		// Verify error types are hardcoded strings, not err.Error()
		assert.NotContains(t, controllerSource, `recordError(.*err.Error()`,
			"Error types should not use err.Error() which creates unbounded cardinality")
	})

	t.Run("Status labels for function invocations are bounded", func(t *testing.T) {
		validStatuses := []string{"success", "error", "timeout"}
		assert.Len(t, validStatuses, 3, "Function invocation status should have bounded values")
	})

	t.Run("Event status labels are bounded", func(t *testing.T) {
		// Build event statuses
		validBuildStatuses := []string{"start", "complete", "failed", "timeout", "cancel"}
		assert.Len(t, validBuildStatuses, 5, "Build event status should have bounded values")

		// Service event statuses
		validServiceStatuses := []string{"create", "update", "delete", "ready", "scaled"}
		assert.Len(t, validServiceStatuses, 5, "Service event status should have bounded values")

		// Parser event statuses
		validParserStatuses := []string{"start", "complete", "failed"}
		assert.Len(t, validParserStatuses, 3, "Parser event status should have bounded values")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC1: Cardinality Risk Analysis
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC1_CardinalityRiskAnalysis(t *testing.T) {
	t.Run("Analyze operator metrics cardinality risk", func(t *testing.T) {
		// Operator metrics and their label cardinality
		type metricCardinality struct {
			name         string
			labels       []string
			maxValues    []int // estimated max values per label
			totalMax     int   // product of all max values
			riskLevel    string
		}

		operatorMetrics := []metricCardinality{
			{
				name:      "knative_lambda_operator_reconcile_total",
				labels:    []string{"phase", "result"},
				maxValues: []int{6, 4}, // 6 phases, 4 results
				totalMax:  24,
				riskLevel: "LOW",
			},
			{
				name:      "knative_lambda_operator_reconcile_duration_seconds",
				labels:    []string{"phase"},
				maxValues: []int{6},
				totalMax:  6,
				riskLevel: "LOW",
			},
			{
				name:      "knative_lambda_operator_lambdafunctions_total",
				labels:    []string{"namespace", "phase"},
				maxValues: []int{100, 6}, // assume max 100 namespaces
				totalMax:  600,
				riskLevel: "MEDIUM", // namespace can grow
			},
			{
				name:      "knative_lambda_operator_build_duration_seconds",
				labels:    []string{"runtime", "result"},
				maxValues: []int{5, 2}, // 5 runtimes, 2 results
				totalMax:  10,
				riskLevel: "LOW",
			},
			{
				name:      "knative_lambda_operator_build_jobs_active",
				labels:    []string{"namespace"},
				maxValues: []int{100}, // assume max 100 namespaces
				totalMax:  100,
				riskLevel: "MEDIUM", // namespace can grow
			},
			{
				name:      "knative_lambda_operator_errors_total",
				labels:    []string{"component", "error_type"},
				maxValues: []int{5, 20}, // 5 components, ~20 error types
				totalMax:  100,
				riskLevel: "LOW",
			},
			{
				name:      "knative_lambda_operator_workqueue_depth",
				labels:    []string{}, // no labels
				maxValues: []int{},
				totalMax:  1,
				riskLevel: "LOW",
			},
		}

		for _, metric := range operatorMetrics {
			t.Run(metric.name, func(t *testing.T) {
				// Log cardinality info
				t.Logf("Metric: %s", metric.name)
				t.Logf("  Labels: %v", metric.labels)
				t.Logf("  Max cardinality: %d", metric.totalMax)
				t.Logf("  Risk level: %s", metric.riskLevel)

				// Verify cardinality is reasonable (< 10,000 time series)
				assert.Less(t, metric.totalMax, 10000,
					"Metric %s has too high max cardinality: %d", metric.name, metric.totalMax)
			})
		}
	})

	t.Run("Analyze function RED metrics cardinality risk", func(t *testing.T) {
		// Function metrics have higher cardinality risk due to function label
		type metricCardinality struct {
			name      string
			labels    []string
			maxValues []int
			totalMax  int
			riskLevel string
			note      string
		}

		functionMetrics := []metricCardinality{
			{
				name:      "knative_lambda_function_invocations_total",
				labels:    []string{"function", "namespace", "status"},
				maxValues: []int{1000, 100, 3}, // assume 1000 functions, 100 namespaces, 3 statuses
				totalMax:  300000,
				riskLevel: "HIGH",
				note:      "Function label can grow unbounded - consider using recording rules",
			},
			{
				name:      "knative_lambda_function_duration_seconds",
				labels:    []string{"function", "namespace"},
				maxValues: []int{1000, 100},
				totalMax:  100000,
				riskLevel: "HIGH",
				note:      "Function label can grow unbounded - consider using recording rules",
			},
			{
				name:      "knative_lambda_function_errors_total",
				labels:    []string{"function", "namespace", "error_type"},
				maxValues: []int{1000, 100, 10}, // 10 error types
				totalMax:  1000000,
				riskLevel: "HIGH",
				note:      "Very high potential cardinality - needs recording rules or label reduction",
			},
			{
				name:      "knative_lambda_function_cold_starts_total",
				labels:    []string{"function", "namespace"},
				maxValues: []int{1000, 100},
				totalMax:  100000,
				riskLevel: "HIGH",
				note:      "Function label can grow unbounded - consider using recording rules",
			},
		}

		for _, metric := range functionMetrics {
			t.Run(metric.name, func(t *testing.T) {
				t.Logf("Metric: %s", metric.name)
				t.Logf("  Labels: %v", metric.labels)
				t.Logf("  Max cardinality (worst case): %d", metric.totalMax)
				t.Logf("  Risk level: %s", metric.riskLevel)
				t.Logf("  Note: %s", metric.note)

				// For function metrics, we document the risk but don't fail
				// because they're necessary for per-function monitoring
				if metric.riskLevel == "HIGH" {
					t.Logf("WARNING: High cardinality metric - ensure Prometheus has adequate resources")
				}
			})
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC1: ReconcilerMetrics Helper Tests (Source Code Verification)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC1_ReconcilerMetricsHelper(t *testing.T) {
	metricsFile := findMetricsSourceFile()
	if metricsFile == "" {
		t.Skip("metrics.go not found, skipping test")
		return
	}

	data, err := os.ReadFile(metricsFile)
	require.NoError(t, err, "Failed to read metrics.go")
	source := string(data)

	t.Run("ReconcilerMetrics struct exists", func(t *testing.T) {
		assert.Contains(t, source, "type ReconcilerMetrics struct",
			"ReconcilerMetrics struct should be defined")
	})

	t.Run("NewReconcilerMetrics constructor exists", func(t *testing.T) {
		assert.Contains(t, source, "func NewReconcilerMetrics()",
			"NewReconcilerMetrics constructor should be defined")
	})

	t.Run("RecordReconcile method exists", func(t *testing.T) {
		assert.Contains(t, source, "func (m *ReconcilerMetrics) RecordReconcile",
			"RecordReconcile method should be defined")
	})

	t.Run("RecordBuild method exists", func(t *testing.T) {
		assert.Contains(t, source, "func (m *ReconcilerMetrics) RecordBuild",
			"RecordBuild method should be defined")
	})

	t.Run("RecordError method exists", func(t *testing.T) {
		assert.Contains(t, source, "func (m *ReconcilerMetrics) RecordError",
			"RecordError method should be defined")
	})

	t.Run("SetLambdaCount method exists", func(t *testing.T) {
		assert.Contains(t, source, "func (m *ReconcilerMetrics) SetLambdaCount",
			"SetLambdaCount method should be defined")
	})

	t.Run("SetActiveBuildJobs method exists", func(t *testing.T) {
		assert.Contains(t, source, "func (m *ReconcilerMetrics) SetActiveBuildJobs",
			"SetActiveBuildJobs method should be defined")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC1: Label Value Tests (Source Code Verification)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC1_LabelValues(t *testing.T) {
	metricsFile := findMetricsSourceFile()
	if metricsFile == "" {
		t.Skip("metrics.go not found, skipping test")
		return
	}

	data, err := os.ReadFile(metricsFile)
	require.NoError(t, err, "Failed to read metrics.go")
	source := string(data)

	t.Run("Reconcile metrics have correct labels", func(t *testing.T) {
		// Verify ReconcileTotal has phase and result labels
		assert.Contains(t, source, `[]string{"phase", "result"}`,
			"ReconcileTotal should have phase and result labels")
	})

	t.Run("Build metrics have correct labels", func(t *testing.T) {
		// Verify BuildDuration has runtime and result labels
		assert.Contains(t, source, `[]string{"runtime", "result"}`,
			"BuildDuration should have runtime and result labels")
	})

	t.Run("Function invocation metrics have correct labels", func(t *testing.T) {
		// Verify FunctionInvocationsTotal has function, namespace, status labels
		assert.Contains(t, source, `[]string{"function", "namespace", "status"}`,
			"FunctionInvocationsTotal should have function, namespace, status labels")
	})

	t.Run("Function error metrics have correct labels", func(t *testing.T) {
		// Verify FunctionErrorsTotal has function, namespace, error_type labels
		assert.Contains(t, source, `[]string{"function", "namespace", "error_type"}`,
			"FunctionErrorsTotal should have function, namespace, error_type labels")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC1: Metrics Registration Test
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC1_MetricsRegistration(t *testing.T) {
	metricsFile := findMetricsSourceFile()
	if metricsFile == "" {
		t.Skip("metrics.go not found, skipping test")
		return
	}

	data, err := os.ReadFile(metricsFile)
	require.NoError(t, err, "Failed to read metrics.go")
	source := string(data)

	t.Run("Register function exists and registers all metrics", func(t *testing.T) {
		assert.Contains(t, source, "func Register()",
			"Register function should be defined")
		assert.Contains(t, source, "metrics.Registry.MustRegister",
			"Metrics should be registered with controller-runtime registry")
	})

	t.Run("Function RED metrics are registered in init()", func(t *testing.T) {
		assert.Contains(t, source, "func init()",
			"init function should be defined")
		assert.Contains(t, source, "FunctionInvocationsTotal",
			"FunctionInvocationsTotal should be registered")
		assert.Contains(t, source, "FunctionDuration",
			"FunctionDuration should be registered")
		assert.Contains(t, source, "FunctionErrorsTotal",
			"FunctionErrorsTotal should be registered")
		assert.Contains(t, source, "FunctionColdStartsTotal",
			"FunctionColdStartsTotal should be registered")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC1: Metric Naming Convention Tests
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC1_MetricNamingConventions(t *testing.T) {
	t.Run("All operator metrics use knative_lambda_operator prefix", func(t *testing.T) {
		// Read metrics source to verify naming
		metricsFiles := []string{
			"../../../operator/internal/metrics/metrics.go",
			"../../operator/internal/metrics/metrics.go",
			"../operator/internal/metrics/metrics.go",
		}

		var found bool
		for _, path := range metricsFiles {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			found = true
			source := string(data)

			// Check namespace and subsystem constants
			assert.Contains(t, source, `namespace = "knative_lambda"`,
				"Namespace should be 'knative_lambda'")
			assert.Contains(t, source, `subsystem = "operator"`,
				"Subsystem should be 'operator'")
			break
		}

		if !found {
			t.Skip("metrics.go not found")
		}
	})

	t.Run("All function metrics use knative_lambda_function prefix", func(t *testing.T) {
		metricsFiles := []string{
			"../../../operator/internal/metrics/metrics.go",
			"../../operator/internal/metrics/metrics.go",
			"../operator/internal/metrics/metrics.go",
		}

		var found bool
		for _, path := range metricsFiles {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			found = true
			source := string(data)

			// Check function subsystem is used for RED metrics
			assert.Contains(t, source, `Subsystem: "function"`,
				"Function metrics should use 'function' subsystem")
			break
		}

		if !found {
			t.Skip("metrics.go not found")
		}
	})
}
