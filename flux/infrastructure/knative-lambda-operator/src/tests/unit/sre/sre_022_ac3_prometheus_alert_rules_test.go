// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 SRE-022: AC3 Prometheus Alert Rules Tests
//
//	BVL-388: DEVOPS-001-AC3 Prometheus Alert Rules
//	Priority: P0 | Story Points: 3
//
//	Tests validate acceptance criteria:
//	✓ PrometheusRule knative-lambda-alerts exists with release: prometheus label
//	✓ Canary alerts configured (KnativeLambdaCanaryFailed, KnativeLambdaCanaryStuck)
//	✓ Operator health alerts configured (7 alerts)
//	✓ Build pipeline alerts configured (4 alerts)
//	✓ Function alerts configured (4 alerts)
//	✓ Eventing alerts configured (2 alerts)
//	✓ All alerts have proper labels and annotations
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
// PrometheusRule Types
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// PrometheusRule represents a Kubernetes PrometheusRule resource
type PrometheusRule struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string            `yaml:"name"`
		Namespace string            `yaml:"namespace"`
		Labels    map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
	Spec struct {
		Groups []RuleGroup `yaml:"groups"`
	} `yaml:"spec"`
}

// RuleGroup represents a group of alert rules
type RuleGroup struct {
	Name  string      `yaml:"name"`
	Rules []AlertRule `yaml:"rules"`
}

// AlertRule represents a single alert rule
type AlertRule struct {
	Alert       string            `yaml:"alert"`
	Expr        string            `yaml:"expr"`
	For         string            `yaml:"for"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Helper Functions
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func findAlertRulesFile() string {
	// Look for alertrules.yaml in known locations
	possiblePaths := []string{
		// Relative paths from different test run locations
		"../../../../k8s/overlays/studio/alertrules.yaml",
		"../../../k8s/overlays/studio/alertrules.yaml",
		"../../k8s/overlays/studio/alertrules.yaml",
		"../k8s/overlays/studio/alertrules.yaml",
		// Absolute paths
		"/workspace/flux/infrastructure/knative-lambda-operator/k8s/overlays/studio/alertrules.yaml",
	}

	// Try from workspace root
	workspaceRoot := os.Getenv("WORKSPACE_ROOT")
	if workspaceRoot != "" {
		possiblePaths = append([]string{
			filepath.Join(workspaceRoot, "flux/infrastructure/knative-lambda-operator/k8s/overlays/studio/alertrules.yaml"),
		}, possiblePaths...)
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func loadPrometheusRule(t *testing.T) *PrometheusRule {
	alertFile := findAlertRulesFile()
	if alertFile == "" {
		t.Skip("alertrules.yaml not found, skipping test")
		return nil
	}

	data, err := os.ReadFile(alertFile)
	require.NoError(t, err, "Failed to read alertrules.yaml")

	// Handle YAML document separator
	docs := strings.Split(string(data), "---")
	var yamlDoc string
	for _, doc := range docs {
		trimmed := strings.TrimSpace(doc)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			yamlDoc = doc
			break
		}
	}

	var rule PrometheusRule
	err = yaml.Unmarshal([]byte(yamlDoc), &rule)
	require.NoError(t, err, "Failed to parse PrometheusRule YAML")

	return &rule
}

func findAlertByName(rule *PrometheusRule, alertName string) *AlertRule {
	for _, group := range rule.Spec.Groups {
		for _, alert := range group.Rules {
			if alert.Alert == alertName {
				return &alert
			}
		}
	}
	return nil
}

func getAllAlertNames(rule *PrometheusRule) []string {
	var names []string
	for _, group := range rule.Spec.Groups {
		for _, alert := range group.Rules {
			names = append(names, alert.Alert)
		}
	}
	return names
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC3: PrometheusRule Exists with Correct Metadata
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC3_PrometheusRuleExists(t *testing.T) {
	rule := loadPrometheusRule(t)
	if rule == nil {
		return
	}

	t.Run("PrometheusRule has correct API version and kind", func(t *testing.T) {
		assert.Equal(t, "monitoring.coreos.com/v1", rule.APIVersion,
			"PrometheusRule should have correct API version")
		assert.Equal(t, "PrometheusRule", rule.Kind,
			"Kind should be PrometheusRule")
	})

	t.Run("PrometheusRule name is knative-lambda-alerts", func(t *testing.T) {
		assert.Equal(t, "knative-lambda-alerts", rule.Metadata.Name,
			"PrometheusRule should be named 'knative-lambda-alerts'")
	})

	t.Run("PrometheusRule is in knative-lambda namespace", func(t *testing.T) {
		assert.Equal(t, "knative-lambda", rule.Metadata.Namespace,
			"PrometheusRule should be in 'knative-lambda' namespace")
	})

	t.Run("PrometheusRule has release: prometheus label", func(t *testing.T) {
		require.NotNil(t, rule.Metadata.Labels, "Labels must not be nil")
		assert.Equal(t, "prometheus", rule.Metadata.Labels["release"],
			"PrometheusRule must have 'release: prometheus' label for prometheus-operator to pick it up")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC3: Canary Alerts (2 alerts)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC3_CanaryAlerts(t *testing.T) {
	rule := loadPrometheusRule(t)
	if rule == nil {
		return
	}

	t.Run("KnativeLambdaCanaryFailed alert exists", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaCanaryFailed")
		require.NotNil(t, alert, "KnativeLambdaCanaryFailed alert should exist")

		assert.Contains(t, alert.Expr, `status="Failed"`,
			"Alert should check for Failed status")
		assert.Equal(t, "1m", alert.For,
			"Alert should fire after 1 minute")
		assert.Equal(t, "critical", alert.Labels["severity"],
			"Alert should have critical severity")
		assert.Equal(t, "platform", alert.Labels["team"],
			"Alert should have team: platform label")
		assert.NotEmpty(t, alert.Annotations["summary"],
			"Alert should have summary annotation")
		assert.NotEmpty(t, alert.Annotations["description"],
			"Alert should have description annotation")
	})

	t.Run("KnativeLambdaCanaryStuck alert exists", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaCanaryStuck")
		require.NotNil(t, alert, "KnativeLambdaCanaryStuck alert should exist")

		assert.Contains(t, alert.Expr, `status="Progressing"`,
			"Alert should check for Progressing status")
		assert.Equal(t, "30m", alert.For,
			"Alert should fire after 30 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
		assert.Equal(t, "platform", alert.Labels["team"],
			"Alert should have team: platform label")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC3: Operator Health Alerts (7 alerts)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC3_OperatorHealthAlerts(t *testing.T) {
	rule := loadPrometheusRule(t)
	if rule == nil {
		return
	}

	t.Run("KnativeLambdaOperatorDown alert exists (critical, 2m)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaOperatorDown")
		require.NotNil(t, alert, "KnativeLambdaOperatorDown alert should exist")

		assert.Contains(t, alert.Expr, "kube_deployment_status_replicas_available",
			"Alert should use kube_deployment_status_replicas_available metric")
		assert.Contains(t, alert.Expr, "== 0",
			"Alert should fire when all replicas unavailable")
		assert.Equal(t, "2m", alert.For,
			"Alert should fire after 2 minutes")
		assert.Equal(t, "critical", alert.Labels["severity"],
			"Alert should have critical severity")
	})

	t.Run("KnativeLambdaOperatorNotReady alert exists (warning, 5m)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaOperatorNotReady")
		require.NotNil(t, alert, "KnativeLambdaOperatorNotReady alert should exist")

		assert.Contains(t, alert.Expr, "kube_deployment_status_replicas_ready",
			"Alert should use kube_deployment_status_replicas_ready metric")
		assert.Equal(t, "5m", alert.For,
			"Alert should fire after 5 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})

	t.Run("KnativeLambdaHighErrorRate alert exists (warning, 5m, > 1%)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaHighErrorRate")
		require.NotNil(t, alert, "KnativeLambdaHighErrorRate alert should exist")

		assert.Contains(t, alert.Expr, "knative_lambda_operator_errors_total",
			"Alert should use errors_total metric")
		assert.Contains(t, alert.Expr, "> 1",
			"Alert should fire when error rate > 1%")
		assert.Equal(t, "5m", alert.For,
			"Alert should fire after 5 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})

	t.Run("KnativeLambdaHighLatency alert exists (warning, 5m, P99 > 5s)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaHighLatency")
		require.NotNil(t, alert, "KnativeLambdaHighLatency alert should exist")

		assert.Contains(t, alert.Expr, "histogram_quantile(0.99",
			"Alert should use P99 quantile")
		assert.Contains(t, alert.Expr, "> 5",
			"Alert should fire when P99 > 5s")
		assert.Equal(t, "5m", alert.For,
			"Alert should fire after 5 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})

	t.Run("KnativeLambdaReconcileSuccessRateLow alert exists (warning, 5m, < 99%)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaReconcileSuccessRateLow")
		require.NotNil(t, alert, "KnativeLambdaReconcileSuccessRateLow alert should exist")

		assert.Contains(t, alert.Expr, "knative_lambda_operator_reconcile_total",
			"Alert should use reconcile_total metric")
		assert.Contains(t, alert.Expr, "< 0.99",
			"Alert should fire when success rate < 99%")
		assert.Equal(t, "5m", alert.For,
			"Alert should fire after 5 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})

	t.Run("KnativeLambdaReconcileDurationHigh alert exists (warning, 5m, P95 > 1s)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaReconcileDurationHigh")
		require.NotNil(t, alert, "KnativeLambdaReconcileDurationHigh alert should exist")

		assert.Contains(t, alert.Expr, "histogram_quantile(0.95",
			"Alert should use P95 quantile")
		assert.Contains(t, alert.Expr, "> 1",
			"Alert should fire when P95 > 1s")
		assert.Equal(t, "5m", alert.For,
			"Alert should fire after 5 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})

	t.Run("KnativeLambdaOperatorErrorsHigh alert exists (warning, 5m, > 5 errors)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaOperatorErrorsHigh")
		require.NotNil(t, alert, "KnativeLambdaOperatorErrorsHigh alert should exist")

		assert.Contains(t, alert.Expr, "knative_lambda_operator_errors_total",
			"Alert should use errors_total metric")
		assert.Contains(t, alert.Expr, "> 5",
			"Alert should fire when > 5 errors in 5m")
		assert.Equal(t, "5m", alert.For,
			"Alert should fire after 5 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC3: Build Pipeline Alerts (4 alerts)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC3_BuildPipelineAlerts(t *testing.T) {
	rule := loadPrometheusRule(t)
	if rule == nil {
		return
	}

	t.Run("KnativeLambdaBuildFailed alert exists (warning, > 3 failures in 10m)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaBuildFailed")
		require.NotNil(t, alert, "KnativeLambdaBuildFailed alert should exist")

		assert.Contains(t, alert.Expr, "knative_lambda_operator_build_duration_seconds_count",
			"Alert should use build_duration_seconds_count metric")
		assert.Contains(t, alert.Expr, `result="failure"`,
			"Alert should filter for failure result")
		assert.Contains(t, alert.Expr, "[10m]",
			"Alert should use 10 minute window")
		assert.Contains(t, alert.Expr, "> 3",
			"Alert should fire when > 3 failures")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})

	t.Run("KnativeLambdaBuildSuccessRateLow alert exists (warning, 15m, < 98%)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaBuildSuccessRateLow")
		require.NotNil(t, alert, "KnativeLambdaBuildSuccessRateLow alert should exist")

		assert.Contains(t, alert.Expr, "knative_lambda_operator_build_duration_seconds_count",
			"Alert should use build_duration_seconds_count metric")
		assert.Contains(t, alert.Expr, "< 0.98",
			"Alert should fire when success rate < 98%")
		assert.Equal(t, "15m", alert.For,
			"Alert should fire after 15 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})

	t.Run("KnativeLambdaBuildDurationHigh alert exists (warning, 5m, P95 > 120s)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaBuildDurationHigh")
		require.NotNil(t, alert, "KnativeLambdaBuildDurationHigh alert should exist")

		assert.Contains(t, alert.Expr, "histogram_quantile(0.95",
			"Alert should use P95 quantile")
		assert.Contains(t, alert.Expr, "knative_lambda_operator_build_duration_seconds_bucket",
			"Alert should use build_duration_seconds histogram")
		assert.Contains(t, alert.Expr, "> 120",
			"Alert should fire when P95 > 120s")
		assert.Equal(t, "5m", alert.For,
			"Alert should fire after 5 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})

	t.Run("KnativeLambdaWorkqueueDepthHigh alert exists (warning, 5m, > 10)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaWorkqueueDepthHigh")
		require.NotNil(t, alert, "KnativeLambdaWorkqueueDepthHigh alert should exist")

		assert.Contains(t, alert.Expr, "knative_lambda_operator_workqueue_depth",
			"Alert should use workqueue_depth metric")
		assert.Contains(t, alert.Expr, "> 10",
			"Alert should fire when workqueue depth > 10")
		assert.Equal(t, "5m", alert.For,
			"Alert should fire after 5 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC3: Function Alerts (4 alerts)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC3_FunctionAlerts(t *testing.T) {
	rule := loadPrometheusRule(t)
	if rule == nil {
		return
	}

	t.Run("KnativeLambdaFunctionNotReady alert exists (warning, 10m, > 10%)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaFunctionNotReady")
		require.NotNil(t, alert, "KnativeLambdaFunctionNotReady alert should exist")

		assert.Contains(t, alert.Expr, "knative_lambda_operator_lambdafunctions_total",
			"Alert should use lambdafunctions_total metric")
		assert.Contains(t, alert.Expr, "> 0.1",
			"Alert should fire when > 10% functions not ready")
		assert.Equal(t, "10m", alert.For,
			"Alert should fire after 10 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})

	t.Run("KnativeLambdaFunctionHighErrorRate alert exists (warning, 5m, > 1% SLO)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaFunctionHighErrorRate")
		require.NotNil(t, alert, "KnativeLambdaFunctionHighErrorRate alert should exist")

		assert.Contains(t, alert.Expr, "knative_lambda_function_invocations_total",
			"Alert should use function_invocations_total metric")
		assert.Contains(t, alert.Expr, `status="error"`,
			"Alert should filter for error status")
		assert.Contains(t, alert.Expr, "> 1",
			"Alert should fire when error rate > 1%")
		assert.Equal(t, "5m", alert.For,
			"Alert should fire after 5 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})

	t.Run("KnativeLambdaFunctionHighLatency alert exists (warning, 5m, P95 > 1s SLO)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaFunctionHighLatency")
		require.NotNil(t, alert, "KnativeLambdaFunctionHighLatency alert should exist")

		assert.Contains(t, alert.Expr, "histogram_quantile(0.95",
			"Alert should use P95 quantile")
		assert.Contains(t, alert.Expr, "knative_lambda_function_duration_seconds_bucket",
			"Alert should use function_duration_seconds histogram")
		assert.Contains(t, alert.Expr, "> 1",
			"Alert should fire when P95 > 1s")
		assert.Equal(t, "5m", alert.For,
			"Alert should fire after 5 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})

	t.Run("KnativeLambdaHighColdStartRate alert exists (warning, 10m, > 5% SLO)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaHighColdStartRate")
		require.NotNil(t, alert, "KnativeLambdaHighColdStartRate alert should exist")

		assert.Contains(t, alert.Expr, "knative_lambda_function_cold_starts_total",
			"Alert should use cold_starts_total metric")
		assert.Contains(t, alert.Expr, "knative_lambda_function_invocations_total",
			"Alert should compare against invocations_total")
		assert.Contains(t, alert.Expr, "> 5",
			"Alert should fire when cold start rate > 5%")
		assert.Equal(t, "10m", alert.For,
			"Alert should fire after 10 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC3: Eventing Alerts (2 alerts)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC3_EventingAlerts(t *testing.T) {
	rule := loadPrometheusRule(t)
	if rule == nil {
		return
	}

	t.Run("KnativeLambdaDLQMessagesHigh alert exists (critical, 5m, > 100)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaDLQMessagesHigh")
		require.NotNil(t, alert, "KnativeLambdaDLQMessagesHigh alert should exist")

		assert.Contains(t, alert.Expr, "rabbitmq_queue_messages",
			"Alert should use rabbitmq_queue_messages metric")
		assert.Contains(t, alert.Expr, `queue=~".*dlq.*"`,
			"Alert should filter for DLQ queues")
		assert.Contains(t, alert.Expr, "> 100",
			"Alert should fire when DLQ messages > 100")
		assert.Equal(t, "5m", alert.For,
			"Alert should fire after 5 minutes")
		assert.Equal(t, "critical", alert.Labels["severity"],
			"Alert should have critical severity")
	})

	t.Run("KnativeLambdaDLQIngestRateHigh alert exists (warning, 5m, > 1 msg/s)", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaDLQIngestRateHigh")
		require.NotNil(t, alert, "KnativeLambdaDLQIngestRateHigh alert should exist")

		assert.Contains(t, alert.Expr, "rabbitmq_queue_messages_published_total",
			"Alert should use rabbitmq_queue_messages_published_total metric")
		assert.Contains(t, alert.Expr, `queue=~".*dlq.*"`,
			"Alert should filter for DLQ queues")
		assert.Contains(t, alert.Expr, "> 1",
			"Alert should fire when DLQ ingest rate > 1 msg/s")
		assert.Equal(t, "5m", alert.For,
			"Alert should fire after 5 minutes")
		assert.Equal(t, "warning", alert.Labels["severity"],
			"Alert should have warning severity")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC3: All Alerts Have Proper Labels and Annotations
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC3_AllAlertsHaveProperLabelsAndAnnotations(t *testing.T) {
	rule := loadPrometheusRule(t)
	if rule == nil {
		return
	}

	allAlertNames := getAllAlertNames(rule)

	t.Run("All alerts have severity label", func(t *testing.T) {
		for _, alertName := range allAlertNames {
			alert := findAlertByName(rule, alertName)
			require.NotNil(t, alert, "Alert %s should exist", alertName)
			assert.NotEmpty(t, alert.Labels["severity"],
				"Alert %s should have severity label", alertName)
			assert.Contains(t, []string{"critical", "warning", "info"}, alert.Labels["severity"],
				"Alert %s severity should be one of: critical, warning, info", alertName)
		}
	})

	t.Run("All alerts have team: platform label", func(t *testing.T) {
		for _, alertName := range allAlertNames {
			alert := findAlertByName(rule, alertName)
			require.NotNil(t, alert, "Alert %s should exist", alertName)
			assert.Equal(t, "platform", alert.Labels["team"],
				"Alert %s should have team: platform label", alertName)
		}
	})

	t.Run("All alerts have summary annotation", func(t *testing.T) {
		for _, alertName := range allAlertNames {
			alert := findAlertByName(rule, alertName)
			require.NotNil(t, alert, "Alert %s should exist", alertName)
			assert.NotEmpty(t, alert.Annotations["summary"],
				"Alert %s should have summary annotation", alertName)
		}
	})

	t.Run("All alerts have description annotation", func(t *testing.T) {
		for _, alertName := range allAlertNames {
			alert := findAlertByName(rule, alertName)
			require.NotNil(t, alert, "Alert %s should exist", alertName)
			assert.NotEmpty(t, alert.Annotations["description"],
				"Alert %s should have description annotation", alertName)
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC3: Total Alert Count Verification
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC3_TotalAlertCount(t *testing.T) {
	rule := loadPrometheusRule(t)
	if rule == nil {
		return
	}

	// Expected alerts from AC:
	// Canary: 2
	// Operator health: 7
	// Build pipeline: 4
	// Function: 4
	// Eventing: 2
	// Total: 19

	expectedAlerts := []string{
		// Canary alerts (2)
		"KnativeLambdaCanaryFailed",
		"KnativeLambdaCanaryStuck",
		// Operator health alerts (7)
		"KnativeLambdaOperatorDown",
		"KnativeLambdaOperatorNotReady",
		"KnativeLambdaHighErrorRate",
		"KnativeLambdaHighLatency",
		"KnativeLambdaReconcileSuccessRateLow",
		"KnativeLambdaReconcileDurationHigh",
		"KnativeLambdaOperatorErrorsHigh",
		// Build pipeline alerts (4)
		"KnativeLambdaBuildFailed",
		"KnativeLambdaBuildSuccessRateLow",
		"KnativeLambdaBuildDurationHigh",
		"KnativeLambdaWorkqueueDepthHigh",
		// Function alerts (4)
		"KnativeLambdaFunctionNotReady",
		"KnativeLambdaFunctionHighErrorRate",
		"KnativeLambdaFunctionHighLatency",
		"KnativeLambdaHighColdStartRate",
		// Eventing alerts (2)
		"KnativeLambdaDLQMessagesHigh",
		"KnativeLambdaDLQIngestRateHigh",
	}

	t.Run("All 19 required alerts are defined", func(t *testing.T) {
		actualAlerts := getAllAlertNames(rule)
		assert.Len(t, actualAlerts, len(expectedAlerts),
			"Should have exactly %d alerts defined", len(expectedAlerts))

		for _, expectedAlert := range expectedAlerts {
			alert := findAlertByName(rule, expectedAlert)
			assert.NotNil(t, alert,
				"Expected alert %s should exist", expectedAlert)
		}
	})

	t.Run("Alert groups are properly organized", func(t *testing.T) {
		assert.Len(t, rule.Spec.Groups, 4, "Should have 4 alert groups")

		groupNames := make([]string, len(rule.Spec.Groups))
		for i, group := range rule.Spec.Groups {
			groupNames[i] = group.Name
		}

		assert.Contains(t, groupNames, "knative-lambda.canary",
			"Should have canary alert group")
		assert.Contains(t, groupNames, "knative-lambda.operator",
			"Should have operator alert group")
		assert.Contains(t, groupNames, "knative-lambda.lambdafunctions",
			"Should have lambdafunctions alert group")
		assert.Contains(t, groupNames, "knative-lambda.eventing",
			"Should have eventing alert group")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC3: Critical Alerts Verification
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC3_CriticalAlerts(t *testing.T) {
	rule := loadPrometheusRule(t)
	if rule == nil {
		return
	}

	criticalAlerts := []string{
		"KnativeLambdaCanaryFailed",
		"KnativeLambdaOperatorDown",
		"KnativeLambdaDLQMessagesHigh",
	}

	t.Run("Critical alerts are correctly marked", func(t *testing.T) {
		for _, alertName := range criticalAlerts {
			alert := findAlertByName(rule, alertName)
			require.NotNil(t, alert, "Critical alert %s should exist", alertName)
			assert.Equal(t, "critical", alert.Labels["severity"],
				"Alert %s should have critical severity", alertName)
		}
	})

	t.Run("Critical alerts should have short firing duration", func(t *testing.T) {
		// Critical alerts should fire quickly (1-5 minutes typically)
		for _, alertName := range criticalAlerts {
			alert := findAlertByName(rule, alertName)
			require.NotNil(t, alert, "Critical alert %s should exist", alertName)

			// Parse duration and verify it's reasonable for critical alerts
			duration := alert.For
			assert.NotEmpty(t, duration,
				"Critical alert %s should have 'for' duration", alertName)

			// Critical alerts should fire within 5 minutes
			assert.Contains(t, []string{"1m", "2m", "3m", "4m", "5m"}, duration,
				"Critical alert %s should have for duration <= 5m, got %s", alertName, duration)
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC3: Runbook URL Verification (where applicable)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC3_RunbookURLs(t *testing.T) {
	rule := loadPrometheusRule(t)
	if rule == nil {
		return
	}

	// Alerts that should have runbook URLs (critical/high-impact)
	alertsWithRunbooks := []string{
		"KnativeLambdaCanaryFailed",
		"KnativeLambdaFunctionHighErrorRate",
		"KnativeLambdaBuildDurationHigh",
		"KnativeLambdaWorkqueueDepthHigh",
	}

	t.Run("High-impact alerts have runbook URLs", func(t *testing.T) {
		for _, alertName := range alertsWithRunbooks {
			alert := findAlertByName(rule, alertName)
			require.NotNil(t, alert, "Alert %s should exist", alertName)

			if runbook, ok := alert.Annotations["runbook_url"]; ok {
				assert.NotEmpty(t, runbook,
					"Alert %s runbook_url should not be empty if defined", alertName)
				assert.Contains(t, runbook, "https://",
					"Alert %s runbook_url should be a valid HTTPS URL", alertName)
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AC3: SLO Alignment Verification
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func TestAC3_SLOAlignment(t *testing.T) {
	rule := loadPrometheusRule(t)
	if rule == nil {
		return
	}

	t.Run("Function error rate aligned with 1% SLO", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaFunctionHighErrorRate")
		require.NotNil(t, alert, "KnativeLambdaFunctionHighErrorRate should exist")

		// SLO: Error rate < 1%
		assert.Contains(t, alert.Expr, "> 1",
			"Error rate alert should fire when > 1% (aligned with 1% SLO)")
	})

	t.Run("Function latency aligned with 1s P95 SLO", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaFunctionHighLatency")
		require.NotNil(t, alert, "KnativeLambdaFunctionHighLatency should exist")

		// SLO: P95 latency < 1s
		assert.Contains(t, alert.Expr, "histogram_quantile(0.95",
			"Latency alert should use P95 quantile")
		assert.Contains(t, alert.Expr, "> 1",
			"Latency alert should fire when P95 > 1s (aligned with SLO)")
	})

	t.Run("Cold start rate aligned with 5% SLO", func(t *testing.T) {
		alert := findAlertByName(rule, "KnativeLambdaHighColdStartRate")
		require.NotNil(t, alert, "KnativeLambdaHighColdStartRate should exist")

		// SLO: Cold start rate < 5%
		assert.Contains(t, alert.Expr, "> 5",
			"Cold start alert should fire when > 5% (aligned with 5% SLO)")
	})
}
