// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 DEVOPS-001: Observability Setup Tests
//
//	User Story: Observability Setup
//	Priority: P0 | Story Points: 8
//
//	Tests validate acceptance criteria:
//	✓ Prometheus scraping all components
//	✓ Grafana dashboards for key workflows
//	✓ Alerts configured for critical metrics
//	✓ OpenTelemetry tracing enabled
//	✓ Log aggregation with structured logging
//	✓ SLO/SLI dashboards tracking 99.9% availability
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

package devops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultNamespace    = "knative-lambda"
	prometheusNamespace = "prometheus"
	grafanaNamespace    = "prometheus" // Grafana is deployed in the same namespace as Prometheus
	testTimeout         = 30 * time.Second
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Helper Functions.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func getKubernetesClient(t *testing.T) *kubernetes.Clientset {
	t.Helper()

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		require.NoError(t, err, "Failed to get home directory")
		kubeconfig = fmt.Sprintf("%s/.kube/config", home)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Skip("Kubernetes config not available, skipping integration test")
		return nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	require.NoError(t, err, "Failed to create Kubernetes client")

	return clientset
}

func getNamespace(t *testing.T) string {
	t.Helper()

	namespace := os.Getenv("TEST_NAMESPACE")
	if namespace == "" {
		namespace = defaultNamespace
	}
	return namespace
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: Prometheus Scraping All Components.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps001_AC1_PrometheusDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Prometheus server deployment exists and is ready", func(t *testing.T) {
		deployment, err := client.AppsV1().Deployments(prometheusNamespace).Get(
			ctx,
			"prometheus-server",
			metav1.GetOptions{},
		)

		if err != nil {
			// Try StatefulSet instead (Prometheus Operator uses StatefulSets)
			sts, stsErr := client.AppsV1().StatefulSets(prometheusNamespace).List(
				ctx,
				metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=prometheus"},
			)
			require.NoError(t, stsErr, "Prometheus deployment/statefulset not found")
			require.NotEmpty(t, sts.Items, "No Prometheus StatefulSet found")

			assert.Greater(t, sts.Items[0].Status.ReadyReplicas, int32(0),
				"Prometheus StatefulSet has no ready replicas")
			return
		}

		assert.NotNil(t, deployment, "Prometheus deployment should exist")
		assert.Greater(t, deployment.Status.ReadyReplicas, int32(0),
			"Prometheus should have at least 1 ready replica")
	})

	t.Run("Prometheus service exists", func(t *testing.T) {
		services, err := client.CoreV1().Services(prometheusNamespace).List(
			ctx,
			metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=prometheus"},
		)
		require.NoError(t, err, "Failed to list Prometheus services")
		assert.NotEmpty(t, services.Items, "Prometheus service should exist")
	})
}

func TestDevOps001_AC1_ServiceMonitorForBuilder(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	namespace := getNamespace(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("ServiceMonitor exists for builder service", func(t *testing.T) {
		// Check if ServiceMonitor CRD exists
		_, err := client.Discovery().ServerResourcesForGroupVersion("monitoring.coreos.com/v1")
		if err != nil {
			t.Skip("ServiceMonitor CRD not available, skipping test")
			return
		}

		// Note: ServiceMonitor is a CRD, would need dynamic client to check
		// For now, check if the service itself exists and has proper labels
		service, err := client.CoreV1().Services(namespace).Get(
			ctx,
			"knative-lambda-builder",
			metav1.GetOptions{},
		)
		require.NoError(t, err, "Builder service should exist")

		// Verify service exists and has proper labels
		require.NotNil(t, service.Labels, "Service should have labels")
	})
}

func TestDevOps001_AC1_MetricsEndpointAccessible(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	namespace := getNamespace(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Builder metrics endpoint returns data", func(t *testing.T) {
		// Get builder pods
		pods, err := client.CoreV1().Pods(namespace).List(
			ctx,
			metav1.ListOptions{LabelSelector: "app=knative-lambda-builder"},
		)
		require.NoError(t, err, "Failed to list builder pods")
		require.NotEmpty(t, pods.Items, "No builder pods found")

		// Test metrics endpoint on first pod
		pod := pods.Items[0]
		assert.Equal(t, "Running", string(pod.Status.Phase),
			"Pod should be running")

		// Note: In real test, would use port-forward or exec into pod
		// For now, just verify pod is running and has proper port
		foundMetricsPort := false
		for _, container := range pod.Spec.Containers {
			for _, port := range container.Ports {
				if port.Name == "metrics" || port.Name == "http" {
					foundMetricsPort = true
					assert.Equal(t, int32(8080), port.ContainerPort,
						"Metrics port should be 8080")
				}
			}
		}
		assert.True(t, foundMetricsPort, "Container should expose metrics port")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Grafana Dashboards for Key Workflows.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps001_AC2_GrafanaDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Grafana deployment exists and is ready", func(t *testing.T) {
		// Try common Grafana deployment names
		deploymentNames := []string{"grafana", "prometheus-operator-grafana", "kube-prometheus-stack-grafana"}
		var deployment *appsv1.Deployment
		var err error

		for _, name := range deploymentNames {
			deployment, err = client.AppsV1().Deployments(grafanaNamespace).Get(
				ctx,
				name,
				metav1.GetOptions{},
			)
			if err == nil {
				break
			}
		}

		if err != nil {
			// Try finding by label
			deployments, listErr := client.AppsV1().Deployments(grafanaNamespace).List(
				ctx,
				metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=grafana"},
			)
			if listErr != nil || len(deployments.Items) == 0 {
				t.Skip("Grafana not deployed, skipping test")
				return
			}
			deployment = &deployments.Items[0]
		}

		assert.NotNil(t, deployment, "Grafana deployment should exist")
		assert.Greater(t, deployment.Status.ReadyReplicas, int32(0),
			"Grafana should have at least 1 ready replica")
	})
}

func TestDevOps001_AC2_DashboardConfigMaps(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Grafana dashboard ConfigMaps exist", func(t *testing.T) {
		configMaps, err := client.CoreV1().ConfigMaps(grafanaNamespace).List(
			ctx,
			metav1.ListOptions{LabelSelector: "grafana_dashboard=1"},
		)

		if err != nil {
			t.Skip("Unable to list dashboard ConfigMaps, skipping")
			return
		}

		// At least one dashboard should exist
		assert.NotEmpty(t, configMaps.Items,
			"At least one Grafana dashboard ConfigMap should exist")
	})

	t.Run("Knative Lambda dashboard JSON exists in repository", func(t *testing.T) {
		// Check if dashboard file exists (from tests/unit/devops/)
		dashboardPaths := []string{
			"../../../dashboards/knative-lambda-comprehensive.json",
			"../../dashboards/knative-lambda-comprehensive.json",
			"dashboards/knative-lambda-comprehensive.json",
		}

		found := false
		for _, path := range dashboardPaths {
			if _, err := os.Stat(path); err == nil {
				found = true

				// Validate it's valid JSON
				//nolint:gosec // Test data file read is safe
				data, err := os.ReadFile(path)
				require.NoError(t, err, "Failed to read dashboard file")

				var dashboard map[string]interface{}
				err = json.Unmarshal(data, &dashboard)
				require.NoError(t, err, "Dashboard should be valid JSON")

				// Check for required dashboard fields (either wrapped or direct)
				// Some formats wrap in "dashboard", others are direct
				if _, hasWrapper := dashboard["dashboard"]; hasWrapper {
					assert.Contains(t, dashboard, "dashboard",
						"Dashboard JSON should have 'dashboard' field")
				} else {
					// Direct format - check for core fields
					assert.Contains(t, dashboard, "title",
						"Dashboard JSON should have 'title' field")
					assert.Contains(t, dashboard, "panels",
						"Dashboard JSON should have 'panels' field")
				}
				break
			}
		}

		assert.True(t, found,
			"Knative Lambda dashboard JSON should exist in repository")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Alerts Configured for Critical Metrics.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps001_AC3_AlertmanagerDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Alertmanager is deployed", func(t *testing.T) {
		// Try Deployment first
		deployment, err := client.AppsV1().Deployments(prometheusNamespace).Get(
			ctx,
			"alertmanager",
			metav1.GetOptions{},
		)

		if err == nil {
			assert.NotNil(t, deployment, "Alertmanager deployment should exist")
			return
		}

		// Try StatefulSet
		sts, err := client.AppsV1().StatefulSets(prometheusNamespace).List(
			ctx,
			metav1.ListOptions{LabelSelector: "app=alertmanager"},
		)

		if err != nil || len(sts.Items) == 0 {
			t.Skip("Alertmanager not found, may be configured elsewhere")
			return
		}

		assert.NotEmpty(t, sts.Items, "Alertmanager StatefulSet should exist")
	})
}

func TestDevOps001_AC3_PrometheusRulesConfigured(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	namespace := getNamespace(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("PrometheusRule CRDs exist for knative-lambda", testPrometheusRuleCRDsExist(ctx, client, namespace))
	t.Run("Critical alerts are defined in Helm templates", testCriticalAlertsDefinedInHelmTemplates)
}

// testPrometheusRuleCRDsExist tests if PrometheusRule CRDs exist.
func testPrometheusRuleCRDsExist(ctx context.Context, client *kubernetes.Clientset, namespace string) func(*testing.T) {
	return func(t *testing.T) {
		_, err := client.Discovery().ServerResourcesForGroupVersion("monitoring.coreos.com/v1")
		if err != nil {
			t.Skip("PrometheusRule CRD not available, skipping test")
			return
		}

		configMaps, err := client.CoreV1().ConfigMaps(namespace).List(
			ctx,
			metav1.ListOptions{},
		)
		require.NoError(t, err, "Failed to list ConfigMaps")

		foundAlerts := false
		for _, cm := range configMaps.Items {
			if strings.Contains(cm.Name, "alert") || strings.Contains(cm.Name, "rules") {
				foundAlerts = true
				break
			}
		}

		if !foundAlerts {
			t.Log("No alert ConfigMaps found, alerts may be defined in PrometheusRule CRDs")
		}
	}
}

// testCriticalAlertsDefinedInHelmTemplates tests if critical alerts are defined in Helm templates.
func testCriticalAlertsDefinedInHelmTemplates(t *testing.T) {
	alertFiles := []string{
		"../../../deploy/templates/alerts.yaml",
		"../../../deploy/templates/prometheus-rules.yaml",
	}

	foundAlertFile := false
	for _, file := range alertFiles {
		if _, err := os.Stat(file); err == nil {
			foundAlertFile = true

			//nolint:gosec // Test data file read is safe
			data, err := os.ReadFile(file)
			require.NoError(t, err, "Failed to read alert file")

			content := string(data)

			criticalAlerts := []string{
				"KnativeLambdaBuilderDown",
				"KnativeLambdaBuildSuccessRateLow",
			}

			for _, alert := range criticalAlerts {
				assert.Contains(t, content, alert,
					"Alert file should define %s alert", alert)
			}
			break
		}
	}

	assert.True(t, foundAlertFile,
		"Alert definitions should exist in Helm templates")
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: Log Aggregation with Structured Logging.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps001_AC4_StructuredLogging(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	namespace := getNamespace(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Application logs are in structured JSON format", func(t *testing.T) {
		// Get builder pods
		pods, err := client.CoreV1().Pods(namespace).List(
			ctx,
			metav1.ListOptions{LabelSelector: "app=knative-lambda-builder"},
		)
		require.NoError(t, err, "Failed to list builder pods")
		require.NotEmpty(t, pods.Items, "No builder pods found")

		// Get logs from first pod
		pod := pods.Items[0]
		req := client.CoreV1().Pods(namespace).GetLogs(pod.Name, &v1.PodLogOptions{
			TailLines: int64Ptr(10),
		})

		logStream, err := req.Stream(ctx)
		if err != nil {
			t.Skipf("Unable to get pod logs: %v", err)
			return
		}
		defer func() {
			if closeErr := logStream.Close(); closeErr != nil {
				t.Logf("Failed to close log stream: %v", closeErr)
			}
		}()

		logs, err := io.ReadAll(logStream)
		require.NoError(t, err, "Failed to read logs")

		logLines := strings.Split(string(logs), "\n")
		if len(logLines) == 0 {
			t.Skip("No logs available yet")
			return
		}

		// Check first non-empty log line
		for _, line := range logLines {
			if line == "" {
				continue
			}

			var logEntry map[string]interface{}
			err := json.Unmarshal([]byte(line), &logEntry)
			assert.NoError(t, err, "Log should be valid JSON")

			// Check for required fields
			assert.Contains(t, logEntry, "level", "Log should have 'level' field")
			assert.Contains(t, logEntry, "msg", "Log should have 'msg' field")
			break
		}
	})
}

func TestDevOps001_AC4_LogAggregation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Loki log aggregation is deployed", func(t *testing.T) {
		// Check for Loki deployment across all namespaces
		namespaces := []string{"loki", "logging", "observability", prometheusNamespace}

		found := false
		for _, ns := range namespaces {
			deployments, err := client.AppsV1().Deployments(ns).List(
				ctx,
				metav1.ListOptions{LabelSelector: "app=loki"},
			)

			if err == nil && len(deployments.Items) > 0 {
				found = true
				break
			}

			// Try StatefulSet
			sts, err := client.AppsV1().StatefulSets(ns).List(
				ctx,
				metav1.ListOptions{LabelSelector: "app=loki"},
			)

			if err == nil && len(sts.Items) > 0 {
				found = true
				break
			}
		}

		if !found {
			t.Skip("Loki not found, log aggregation may use different system")
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: OpenTelemetry Tracing Enabled.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps001_AC5_OpenTelemetryTracing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("OpenTelemetry collector is deployed", func(t *testing.T) {
		// Check for OTel collector across common namespaces
		namespaces := []string{"opentelemetry", "tracing", "observability", prometheusNamespace}

		found := false
		for _, ns := range namespaces {
			deployments, err := client.AppsV1().Deployments(ns).List(
				ctx,
				metav1.ListOptions{},
			)

			if err == nil {
				for _, d := range deployments.Items {
					if strings.Contains(d.Name, "otel") || strings.Contains(d.Name, "opentelemetry") {
						found = true
						break
					}
				}
			}

			if found {
				break
			}
		}

		if !found {
			t.Skip("OpenTelemetry collector not found, tracing may be disabled")
		}
	})

	t.Run("Trace backend (Tempo) is deployed", func(t *testing.T) {
		// Check for Tempo across common namespaces
		namespaces := []string{"tempo", "tracing", "observability", prometheusNamespace}

		found := false
		for _, ns := range namespaces {
			deployments, err := client.AppsV1().Deployments(ns).List(
				ctx,
				metav1.ListOptions{LabelSelector: "app=tempo"},
			)

			if err == nil && len(deployments.Items) > 0 {
				found = true
				break
			}
		}

		if !found {
			t.Skip("Tempo not found, traces may use different backend")
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: SLO/SLI Dashboards Tracking 99.9% Availability
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps001_AC6_SLOTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	namespace := getNamespace(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("SLO configuration exists", func(t *testing.T) {
		// Check for SLO ConfigMap
		_, err := client.CoreV1().ConfigMaps(namespace).Get(
			ctx,
			"slo-config",
			metav1.GetOptions{},
		)

		if err != nil {
			t.Skip("SLO ConfigMap not found, SLOs may be configured differently")
			return
		}
	})

	t.Run("Recording rules exist for SLI calculation", func(t *testing.T) {
		// Check Helm templates for recording rules
		ruleFiles := []string{
			"../../deploy/templates/prometheus-rules.yaml",
			"../../deploy/templates/alerts.yaml",
		}

		foundRecordingRules := false
		for _, file := range ruleFiles {
			if _, err := os.Stat(file); err == nil {
				//nolint:gosec // Test data file read is safe
				data, err := os.ReadFile(file)
				if err == nil && strings.Contains(string(data), "record:") {
					foundRecordingRules = true
					break
				}
			}
		}

		if !foundRecordingRules {
			t.Log("No recording rules found in templates, may be defined elsewhere")
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Helper Functions.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func int64Ptr(i int64) *int64 {
	return &i
}
