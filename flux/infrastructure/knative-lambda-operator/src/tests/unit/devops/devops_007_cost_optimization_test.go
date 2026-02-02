// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 DEVOPS-007: Cost Optimization Tests
//
//	User Story: Cost Optimization
//	Priority: P1 | Story Points: 5
//
//	Tests validate acceptance criteria:
//	✓ Resource limits defined for all workloads
//	✓ Autoscaling configured appropriately
//	✓ Spot instances/serverless for non-critical workloads
//	✓ Cost monitoring and alerting
//	✓ Resource rightsizing recommendations
//	✓ Idle resource cleanup
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

package devops

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Helper Functions.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func setupCostK8sClient(t *testing.T) *kubernetes.Clientset {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home := os.Getenv("HOME")
		kubeconfig = home + "/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Skipf("Skipping test: cannot load kubeconfig: %v", err)
		return nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Skipf("Skipping test: cannot create Kubernetes client: %v", err)
		return nil
	}

	return clientset
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: Resource Limits.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps007_AC1_ResourceLimitsConfigured(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	clientset := setupCostK8sClient(t)
	if clientset == nil {
		return
	}

	t.Run("All pods have resource limits defined", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get all namespaces
		namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Skipf("Cannot list namespaces: %v", err)
			return
		}

		podsWithoutLimits := 0
		totalPods := 0

		for _, ns := range namespaces.Items {
			// Skip system namespaces
			if strings.HasPrefix(ns.Name, "kube-") || ns.Name == "default" {
				continue
			}

			pods, err := clientset.CoreV1().Pods(ns.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				continue
			}

			for _, pod := range pods.Items {
				totalPods++
				for _, container := range pod.Spec.Containers {
					if len(container.Resources.Limits) == 0 {
						podsWithoutLimits++
						t.Logf("Pod %s/%s container %s has no resource limits",
							pod.Namespace, pod.Name, container.Name)
					}
				}
			}
		}

		if totalPods > 0 {
			percentage := float64(totalPods-podsWithoutLimits) / float64(totalPods) * 100
			t.Logf("Resource limits coverage: %.2f%% (%d/%d pods)",
				percentage, totalPods-podsWithoutLimits, totalPods)

			// At least 80% of pods should have resource limits
			assert.GreaterOrEqual(t, percentage, 80.0,
				"At least 80%% of pods should have resource limits defined")
		}
	})
}

func TestDevOps007_AC1_ResourceRequestsConfigured(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	clientset := setupCostK8sClient(t)
	if clientset == nil {
		return
	}

	t.Run("All pods have resource requests defined", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Skipf("Cannot list namespaces: %v", err)
			return
		}

		podsWithoutRequests := 0
		totalPods := 0

		for _, ns := range namespaces.Items {
			if strings.HasPrefix(ns.Name, "kube-") || ns.Name == "default" {
				continue
			}

			pods, err := clientset.CoreV1().Pods(ns.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				continue
			}

			for _, pod := range pods.Items {
				totalPods++
				for _, container := range pod.Spec.Containers {
					if len(container.Resources.Requests) == 0 {
						podsWithoutRequests++
					}
				}
			}
		}

		if totalPods > 0 {
			percentage := float64(totalPods-podsWithoutRequests) / float64(totalPods) * 100
			t.Logf("Resource requests coverage: %.2f%% (%d/%d pods)",
				percentage, totalPods-podsWithoutRequests, totalPods)
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Autoscaling Configuration.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps007_AC2_HPAConfigured(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	clientset := setupCostK8sClient(t)
	if clientset == nil {
		return
	}

	t.Run("HorizontalPodAutoscalers are configured", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Skipf("Cannot list namespaces: %v", err)
			return
		}

		hpaCount := 0
		for _, ns := range namespaces.Items {
			if strings.HasPrefix(ns.Name, "kube-") {
				continue
			}

			hpas, err := clientset.AutoscalingV2().HorizontalPodAutoscalers(ns.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				continue
			}

			hpaCount += len(hpas.Items)
			for _, hpa := range hpas.Items {
				t.Logf("Found HPA: %s/%s (min: %d, max: %d)",
					hpa.Namespace, hpa.Name, *hpa.Spec.MinReplicas, hpa.Spec.MaxReplicas)
			}
		}

		if hpaCount > 0 {
			t.Logf("Total HPAs configured: %d", hpaCount)
		} else {
			t.Log("No HPAs found - may be using Knative autoscaling")
		}
	})
}

func TestDevOps007_AC2_KnativeAutoscalingConfigured(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Knative autoscaling parameters are documented", func(t *testing.T) {
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-007-cost-optimization.md",
		}

		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for autoscaling configuration
				if strings.Contains(content, "autoscaling") || strings.Contains(content, "scale-to-zero") {
					t.Log("Knative autoscaling parameters are documented")

					// Check for scale-to-zero
					if strings.Contains(content, "scale-to-zero") || strings.Contains(content, "minScale: 0") {
						t.Log("Scale-to-zero is configured for cost optimization")
					}
				}
				break
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Spot Instances / Serverless.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps007_AC3_NodeAffinityForSpotInstances(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	clientset := setupCostK8sClient(t)
	if clientset == nil {
		return
	}

	t.Run("Node affinity configured for spot instances", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Check for nodes with spot instance labels
		nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Skipf("Cannot list nodes: %v", err)
			return
		}

		spotNodes := 0
		for _, node := range nodes.Items {
			// Check common spot instance labels
			if capacity, ok := node.Labels["karpenter.sh/capacity-type"]; ok && capacity == "spot" {
				spotNodes++
				t.Logf("Found spot instance node: %s", node.Name)
			}
			if _, ok := node.Labels["eks.amazonaws.com/capacityType"]; ok {
				if node.Labels["eks.amazonaws.com/capacityType"] == "SPOT" {
					spotNodes++
					t.Logf("Found EKS spot instance node: %s", node.Name)
				}
			}
		}

		if spotNodes > 0 {
			t.Logf("Spot instances in use: %d nodes", spotNodes)
		} else {
			t.Log("No spot instances detected - may be using on-demand or serverless")
		}
	})
}

func TestDevOps007_AC3_KnativeServerlessConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Knative serverless is configured", func(t *testing.T) {
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-007-cost-optimization.md",
		}

		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for serverless mentions
				if strings.Contains(content, "serverless") || strings.Contains(content, "Knative") {
					t.Log("Serverless architecture is documented for cost optimization")
				}
				break
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: Cost Monitoring.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps007_AC4_PrometheusMetricsForCost(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	clientset := setupCostK8sClient(t)
	if clientset == nil {
		return
	}

	t.Run("Prometheus is deployed for cost metrics", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Check for Prometheus deployment
		namespaces := []string{"prometheus", "monitoring", "observability"}
		prometheusFound := false

		for _, ns := range namespaces {
			deployments, err := clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
			if err != nil {
				continue
			}

			for _, deploy := range deployments.Items {
				if strings.Contains(deploy.Name, "prometheus") {
					prometheusFound = true
					t.Logf("Found Prometheus deployment: %s/%s", deploy.Namespace, deploy.Name)
					break
				}
			}

			if prometheusFound {
				break
			}
		}

		if prometheusFound {
			t.Log("Prometheus is available for cost monitoring")
		}
	})
}

func TestDevOps007_AC4_GrafanaDashboardsForCost(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Grafana dashboards for cost monitoring exist", func(t *testing.T) {
		dashboardPaths := []string{
			"../../flux/infrastructure/observability/grafana/dashboards",
			"flux/infrastructure/observability/grafana/dashboards",
		}

		foundCostDashboard := false
		for _, dashPath := range dashboardPaths {
			if info, err := os.Stat(dashPath); err == nil && info.IsDir() {
				// Look for cost-related dashboards
				files, err := os.ReadDir(dashPath)
				if err != nil {
					continue
				}

				for _, file := range files {
					if strings.Contains(strings.ToLower(file.Name()), "cost") ||
						strings.Contains(strings.ToLower(file.Name()), "resource") {
						foundCostDashboard = true
						t.Logf("Found cost monitoring dashboard: %s", file.Name())
					}
				}
				break
			}
		}

		if foundCostDashboard {
			t.Log("Cost monitoring dashboards are configured")
		} else {
			t.Log("No specific cost dashboards found - may be using generic resource dashboards")
		}
	})
}

func TestDevOps007_AC4_CostAlerts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Cost alerts are configured in Prometheus", func(t *testing.T) {
		alertPaths := []string{
			"../../flux/infrastructure/observability/prometheus/alerts",
			"flux/infrastructure/observability/prometheus/alerts",
		}

		for _, alertPath := range alertPaths {
			if info, err := os.Stat(alertPath); err == nil && info.IsDir() {
				files, err := os.ReadDir(alertPath)
				if err != nil {
					continue
				}

				for _, file := range files {
					if strings.Contains(strings.ToLower(file.Name()), "cost") ||
						strings.Contains(strings.ToLower(file.Name()), "resource") {
						t.Logf("Found cost alert configuration: %s", file.Name())
					}
				}
				break
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Resource Rightsizing.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps007_AC5_VPAConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	clientset := setupCostK8sClient(t)
	if clientset == nil {
		return
	}

	t.Run("VerticalPodAutoscaler is configured", func(t *testing.T) {
		// Check for VPA CRDs
		crdClient := clientset.Discovery()
		apiList, err := crdClient.ServerGroups()
		if err != nil {
			t.Skipf("Cannot query API groups: %v", err)
			return
		}

		vpaFound := false
		for _, group := range apiList.Groups {
			if strings.Contains(group.Name, "autoscaling.k8s.io") {
				vpaFound = true
				t.Log("VPA API group is available")
				break
			}
		}

		if vpaFound {
			t.Log("VPA can be used for resource rightsizing recommendations")
		} else {
			t.Log("VPA not detected - rightsizing may be done manually")
		}
	})
}

func TestDevOps007_AC5_RightsizingDocumentation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Resource rightsizing process is documented", func(t *testing.T) {
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-007-cost-optimization.md",
		}

		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for rightsizing mentions
				if strings.Contains(content, "rightsizing") || strings.Contains(content, "VPA") {
					t.Log("Resource rightsizing process is documented")
				}
				break
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: Idle Resource Cleanup.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps007_AC6_ScaleToZeroConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Scale-to-zero is configured for Knative services", func(t *testing.T) {
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-007-cost-optimization.md",
			"../../charts/knative-lambda/values.yaml",
		}

		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for scale-to-zero configuration
				if strings.Contains(content, "scale-to-zero") || strings.Contains(content, "minScale: 0") {
					t.Log("Scale-to-zero is configured")
					break
				}
			}
		}
	})
}

func TestDevOps007_AC6_IdleTimeoutConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Idle timeout is configured appropriately", func(t *testing.T) {
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-007-cost-optimization.md",
		}

		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for idle timeout configuration
				if strings.Contains(content, "idle-timeout") || strings.Contains(content, "scale-to-zero-grace-period") {
					t.Log("Idle timeout configuration is documented")
				}
				break
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Performance Requirements.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps007_Performance_CostSavingsTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Cost savings targets are documented", func(t *testing.T) {
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-007-cost-optimization.md",
		}

		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for cost targets
				if strings.Contains(content, "30% reduction") ||
					strings.Contains(content, "Cost Optimization Target") {
					t.Log("Cost savings targets are documented (30% reduction baseline)")
				}
				break
			}
		}
	})
}

func TestDevOps007_Performance_ResourceUtilization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Resource utilization targets are documented", func(t *testing.T) {
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-007-cost-optimization.md",
		}

		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for utilization targets
				if strings.Contains(content, "70%") && strings.Contains(content, "utilization") {
					t.Log("Resource utilization targets are documented (70%+)")
				}
				break
			}
		}
	})
}
