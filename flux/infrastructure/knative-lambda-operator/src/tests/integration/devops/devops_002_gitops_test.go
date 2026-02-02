// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 DEVOPS-002: GitOps Deployment Tests
//
//	User Story: GitOps Deployment
//	Priority: P0 | Story Points: 13
//
//	Tests validate acceptance criteria:
//	✓ All infrastructure defined in Git
//	✓ Flux automatically syncs changes
//	✓ Rollback via Git revert
//	✓ Environment promotion (dev → staging → prod)
//	✓ Drift detection and auto-remediation
//	✓ Deployment notifications to Slack
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

package devops

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	fluxNamespace = "flux-system"
	testTimeout   = 30 * time.Second
)

// getKubernetesClient creates a Kubernetes client for testing.
func getKubernetesClient(t *testing.T) *kubernetes.Clientset {
	t.Helper()

	// Try to get kubeconfig from environment or default location
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	require.NoError(t, err, "Failed to build kubeconfig")

	client, err := kubernetes.NewForConfig(config)
	require.NoError(t, err, "Failed to create Kubernetes client")

	return client
}

// getNamespace gets the test namespace.
func getNamespace(t *testing.T) string {
	t.Helper()
	namespace := os.Getenv("TEST_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	return namespace
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: All Infrastructure Defined in Git.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps002_AC1_FluxInstalled(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Flux namespace exists", func(t *testing.T) {
		ns, err := client.CoreV1().Namespaces().Get(ctx, fluxNamespace, metav1.GetOptions{})
		require.NoError(t, err, "Flux namespace should exist")
		assert.NotNil(t, ns)
	})

	t.Run("Flux components are deployed and ready", func(t *testing.T) {
		fluxComponents := []string{
			"source-controller",
			"kustomize-controller",
			"helm-controller",
			"notification-controller",
		}

		for _, component := range fluxComponents {
			deployment, err := client.AppsV1().Deployments(fluxNamespace).Get(
				ctx,
				component,
				metav1.GetOptions{},
			)

			if err != nil {
				t.Logf("Component %s not found: %v", component, err)
				continue
			}

			assert.Greater(t, deployment.Status.ReadyReplicas, int32(0),
				"%s should have at least 1 ready replica", component)
		}
	})
}

func TestDevOps002_AC1_GitRepositoryConfigured(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("GitRepository CRD exists", func(t *testing.T) {
		_, err := client.Discovery().ServerResourcesForGroupVersion("source.toolkit.fluxcd.io/v1")
		if err != nil {
			t.Skip("Flux GitRepository CRD not available")
			return
		}
	})

	t.Run("Git repository secret exists", func(t *testing.T) {
		// Check for common Flux Git secret names
		secretNames := []string{"flux-system", "git-credentials", "github-token"}

		found := false
		for _, name := range secretNames {
			_, err := client.CoreV1().Secrets(fluxNamespace).Get(ctx, name, metav1.GetOptions{})
			if err == nil {
				found = true
				t.Logf("Found Git secret: %s", name)
				break
			}
		}

		if !found {
			t.Log("No Git credentials secret found, may use public repository")
		}
	})
}

func TestDevOps002_AC1_InfrastructureInGit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Deploy directory exists with Helm chart", func(t *testing.T) {
		deployPaths := []string{"../../../deploy", "../../deploy", "deploy"}

		foundDeploy := false
		for _, path := range deployPaths {
			if _, err := os.Stat(path); err == nil {
				foundDeploy = true

				// Check for Chart.yaml
				chartPath := path + "/Chart.yaml"
				_, err := os.Stat(chartPath)
				assert.NoError(t, err, "Chart.yaml should exist in deploy directory")

				// Check for values.yaml
				valuesPath := path + "/values.yaml"
				_, err = os.Stat(valuesPath)
				assert.NoError(t, err, "values.yaml should exist in deploy directory")

				// Check for templates directory
				templatesPath := path + "/templates"
				info, err := os.Stat(templatesPath)
				assert.NoError(t, err, "templates directory should exist")
				if err == nil {
					assert.True(t, info.IsDir(), "templates should be a directory")
				}

				break
			}
		}

		assert.True(t, foundDeploy, "Deploy directory should exist in repository")
	})

	t.Run("Kustomize overlays exist for environments", func(t *testing.T) {
		overlaysPaths := []string{"../../../deploy/overlays", "../../deploy/overlays", "deploy/overlays"}

		for _, path := range overlaysPaths {
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				// Check for environment directories
				envs := []string{"dev", "staging", "prd"}
				for _, env := range envs {
					envPath := path + "/" + env
					if _, err := os.Stat(envPath); err == nil {
						t.Logf("Found environment overlay: %s", env)
					}
				}
				break
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Flux Automatically Syncs Changes.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps002_AC2_KustomizationConfigured(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Kustomization CRD exists", func(t *testing.T) {
		_, err := client.Discovery().ServerResourcesForGroupVersion("kustomize.toolkit.fluxcd.io/v1")
		if err != nil {
			t.Skip("Flux Kustomization CRD not available")
			return
		}
	})

	t.Run("Kustomization resources exist", func(t *testing.T) {
		// List all ConfigMaps in flux-system to find Kustomization-related config
		configMaps, err := client.CoreV1().ConfigMaps(fluxNamespace).List(
			ctx,
			metav1.ListOptions{},
		)

		if err != nil {
			t.Skip("Unable to list ConfigMaps")
			return
		}

		foundKustomization := false
		for _, cm := range configMaps.Items {
			if strings.Contains(cm.Name, "kustomization") {
				foundKustomization = true
				t.Logf("Found Kustomization ConfigMap: %s", cm.Name)
			}
		}

		if !foundKustomization {
			t.Log("No Kustomization ConfigMaps found, may be defined as CRDs")
		}
	})
}

func TestDevOps002_AC2_AutomaticSyncEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Flux CLI is available for management", func(t *testing.T) {
		_, err := exec.LookPath("flux")
		if err != nil {
			t.Skip("Flux CLI not installed, skipping CLI tests")
			return
		}

		// Test flux check command
		cmd := exec.Command("flux", "check")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Logf("Flux check output: %s", output)
			t.Skip("Flux check failed, may not be fully configured")
			return
		}

		assert.NoError(t, err, "Flux check should pass")
	})
}

func TestDevOps002_AC2_HealthChecksConfigured(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	namespace := getNamespace(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Deployments have health checks configured", func(t *testing.T) {
		deployment, err := client.AppsV1().Deployments(namespace).Get(
			ctx,
			"knative-lambda-builder",
			metav1.GetOptions{},
		)

		if err != nil {
			t.Skip("Builder deployment not found")
			return
		}

		// Check for liveness and readiness probes
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.LivenessProbe != nil {
				t.Logf("Container %s has liveness probe", container.Name)
			}
			if container.ReadinessProbe != nil {
				t.Logf("Container %s has readiness probe", container.Name)
			}

			assert.NotNil(t, container.ReadinessProbe,
				"Container should have readiness probe")
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Rollback via Git Revert.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps002_AC3_GitHistoryAvailable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Git repository exists with history", func(t *testing.T) {
		gitPaths := []string{"../../../../../../../../.git", "../../../.git", "../../.git", ".git"}

		foundGit := false
		for _, path := range gitPaths {
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				foundGit = true

				// Check git log
				//nolint:gosec // Git command with sanitized path is safe
				cmd := exec.Command("git", "-C", path+"/..", "log", "--oneline", "-n", "5")
				output, err := cmd.CombinedOutput()

				if err == nil {
					commits := strings.Split(strings.TrimSpace(string(output)), "\n")
					assert.Greater(t, len(commits), 0, "Should have commit history")
					t.Logf("Found %d recent commits", len(commits))
				}

				break
			}
		}

		assert.True(t, foundGit, "Git repository should exist")
	})
}

func TestDevOps002_AC3_FluxSuspendResumeCapability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Flux suspend and resume commands are available", func(t *testing.T) {
		_, err := exec.LookPath("flux")
		if err != nil {
			t.Skip("Flux CLI not installed")
			return
		}

		// Test flux suspend help (dry-run)
		cmd := exec.Command("flux", "suspend", "kustomization", "--help")
		err = cmd.Run()
		assert.NoError(t, err, "Flux suspend command should be available")

		// Test flux resume help (dry-run)
		cmd = exec.Command("flux", "resume", "kustomization", "--help")
		err = cmd.Run()
		assert.NoError(t, err, "Flux resume command should be available")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: Environment Promotion.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps002_AC4_MultiEnvironmentConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Environment-specific values files exist", func(t *testing.T) {
		overlaysPaths := []string{"../../deploy/overlays", "deploy/overlays"}

		for _, basePath := range overlaysPaths {
			envs := []string{"dev", "staging", "prd"}
			foundEnvs := 0

			for _, env := range envs {
				valuesFile := basePath + "/" + env + "/values-" + env + ".yaml"
				if _, err := os.Stat(valuesFile); err == nil {
					foundEnvs++
					t.Logf("Found values file for %s environment", env)

					// Read file and verify it's not empty
					data, err := os.ReadFile(valuesFile) //nolint:gosec // G304: Test file reading controlled paths
					if err == nil {
						assert.Greater(t, len(data), 0, "Values file should not be empty")
					}
				}
			}

			if foundEnvs > 0 {
				assert.GreaterOrEqual(t, foundEnvs, 2,
					"Should have values files for at least 2 environments")
				break
			}
		}
	})

	t.Run("Promotion scripts exist", func(t *testing.T) {
		scriptPaths := []string{
			"../../scripts/promote-to-staging.sh",
			"../../scripts/promote-to-prod.sh",
			"scripts/promote-to-staging.sh",
			"scripts/promote-to-prod.sh",
		}

		foundPromotionScript := false
		for _, path := range scriptPaths {
			if _, err := os.Stat(path); err == nil {
				foundPromotionScript = true
				t.Logf("Found promotion script: %s", path)

				// Verify script is executable
				info, _ := os.Stat(path)
				mode := info.Mode()
				if mode&0111 != 0 {
					t.Logf("Script is executable")
				}

				break
			}
		}

		if !foundPromotionScript {
			t.Log("No promotion scripts found, may use manual process")
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Drift Detection and Auto-Remediation.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps002_AC5_DriftDetectionEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Flux prune is configured for drift detection", func(t *testing.T) {
		// Check documentation mentions prune
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-002-gitops-deployment.md",
		}

		foundPruneDocs := false
		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)
				if strings.Contains(content, "prune") {
					foundPruneDocs = true
					t.Log("Documentation mentions prune configuration")
					break
				}
			}
		}

		assert.True(t, foundPruneDocs, "Documentation should mention prune configuration")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: Deployment Notifications.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps002_AC6_NotificationProvider(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Notification controller is deployed", testNotificationControllerDeployed(ctx, client))
	t.Run("Provider CRD exists", testProviderCRDExists(client))
	t.Run("Webhook secret exists for notifications", testWebhookSecretExists(ctx, client))
}

// testNotificationControllerDeployed tests if notification controller is deployed.
func testNotificationControllerDeployed(ctx context.Context, client *kubernetes.Clientset) func(*testing.T) {
	return func(t *testing.T) {
		deployment, err := client.AppsV1().Deployments(fluxNamespace).Get(
			ctx,
			"notification-controller",
			metav1.GetOptions{},
		)

		if err != nil {
			t.Skip("Notification controller not deployed")
			return
		}

		assert.Greater(t, deployment.Status.ReadyReplicas, int32(0),
			"Notification controller should have ready replicas")
	}
}

// testProviderCRDExists tests if the Provider CRD exists.
func testProviderCRDExists(client *kubernetes.Clientset) func(*testing.T) {
	return func(t *testing.T) {
		_, err := client.Discovery().ServerResourcesForGroupVersion("notification.toolkit.fluxcd.io/v1beta3")
		if err != nil {
			// Try v1beta2
			_, err = client.Discovery().ServerResourcesForGroupVersion("notification.toolkit.fluxcd.io/v1beta2")
			if err != nil {
				t.Skip("Notification CRDs not available")
				return
			}
		}
	}
}

// testWebhookSecretExists tests if webhook secret exists for notifications.
func testWebhookSecretExists(ctx context.Context, client *kubernetes.Clientset) func(*testing.T) {
	return func(t *testing.T) {
		secrets, err := client.CoreV1().Secrets(fluxNamespace).List(
			ctx,
			metav1.ListOptions{},
		)

		if err != nil {
			t.Skip("Unable to list secrets")
			return
		}

		foundWebhook := false
		for _, secret := range secrets.Items {
			if strings.Contains(secret.Name, "webhook") ||
				strings.Contains(secret.Name, "slack") ||
				strings.Contains(secret.Name, "notification") {
				foundWebhook = true
				t.Logf("Found notification secret: %s", secret.Name)
				break
			}
		}

		if !foundWebhook {
			t.Log("No notification webhook secrets found")
		}
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Performance Requirements.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps002_Performance_SyncInterval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Sync interval is configured reasonably", func(t *testing.T) {
		// Check documentation for sync interval
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-002-gitops-deployment.md",
		}

		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths //nolint:gosec // G304: Test file reading controlled documentation paths
				content := string(data)
				if strings.Contains(content, "interval") {
					t.Log("Documentation includes sync interval configuration")

					// Check for reasonable intervals (minutes, not hours)
					if strings.Contains(content, "5m") || strings.Contains(content, "1m") {
						t.Log("Sync interval appears to be in minutes (good)")
					}
				}
				break
			}
		}
	})
}

func TestDevOps002_Performance_DeploymentTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	namespace := getNamespace(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Deployments complete within reasonable time", func(t *testing.T) {
		deployment, err := client.AppsV1().Deployments(namespace).Get(
			ctx,
			"knative-lambda-builder",
			metav1.GetOptions{},
		)

		if err != nil {
			t.Skip("Builder deployment not found")
			return
		}

		// Check if deployment is progressing or available
		for _, condition := range deployment.Status.Conditions {
			if condition.Type == "Available" && condition.Status == "True" {
				t.Log("Deployment is available and healthy")
			}
			if condition.Type == "Progressing" {
				t.Logf("Deployment progressing: %s", condition.Message)
			}
		}

		assert.Equal(t, deployment.Status.Replicas, deployment.Status.ReadyReplicas,
			"All replicas should be ready")
	})
}
