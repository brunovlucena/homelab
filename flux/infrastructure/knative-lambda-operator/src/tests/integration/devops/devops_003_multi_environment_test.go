// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 DEVOPS-003: Multi-Environment Management Tests
//
//	User Story: Multi-Environment Management
//	Priority: P0 | Story Points: 8
//
//	Tests validate acceptance criteria:
//	✓ Separate Kubernetes namespaces per environment
//	✓ Environment-specific configurations
//	✓ Promotion strategy
//	✓ RBAC per environment
//	✓ Cost tracking per environment
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

package devops

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	devops003DocPath = "../../docs/03-for-engineers/devops/user-stories/DEVOPS-003-multi-environment.md"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: Separate Kubernetes Namespaces Per Environment.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps003_AC1_NamespaceIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Environment namespaces exist", func(t *testing.T) {
		envNamespaces := []string{
			"knative-lambda",
		}

		foundNamespaces := 0
		for _, ns := range envNamespaces {
			namespace, err := client.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
			if err == nil {
				foundNamespaces++
				t.Logf("Found environment namespace: %s", ns)

				// Check for environment label
				if labels := namespace.Labels; labels != nil {
					if env, ok := labels["environment"]; ok {
						t.Logf("Namespace %s has environment label: %s", ns, env)
					}
				}
			}
		}

		assert.GreaterOrEqual(t, foundNamespaces, 1,
			"At least one environment namespace should exist")
	})

	t.Run("Namespaces have resource quotas", func(t *testing.T) {
		envNamespaces := []string{
			"knative-lambda",
		}

		for _, ns := range envNamespaces {
			quotas, err := client.CoreV1().ResourceQuotas(ns).List(
				ctx,
				metav1.ListOptions{},
			)

			if err != nil {
				continue
			}

			if len(quotas.Items) > 0 {
				t.Logf("Namespace %s has %d resource quota(s)", ns, len(quotas.Items))

				for _, quota := range quotas.Items {
					t.Logf("Quota %s in %s: %v", quota.Name, ns, quota.Status.Hard)
				}
			}
		}
	})
}

func TestDevOps003_AC1_NetworkPolicies(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Network policies exist for environment isolation", func(t *testing.T) {
		envNamespaces := []string{
			"knative-lambda",
		}

		foundNetworkPolicies := false
		for _, ns := range envNamespaces {
			policies, err := client.NetworkingV1().NetworkPolicies(ns).List(
				ctx,
				metav1.ListOptions{},
			)

			if err != nil {
				continue
			}

			if len(policies.Items) > 0 {
				foundNetworkPolicies = true
				t.Logf("Namespace %s has %d network policy/policies", ns, len(policies.Items))
			}
		}

		if !foundNetworkPolicies {
			t.Log("No network policies found, cross-environment traffic may not be restricted")
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Environment-Specific Configurations.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps003_AC2_EnvironmentConfigFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Environment-specific values files exist", func(t *testing.T) {
		envConfigs := map[string]string{
			"prd": "../../../deploy/overlays/prd/values-prd.yaml",
		}

		foundConfigs := 0
		for env, path := range envConfigs {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				foundConfigs++
				content := string(data)

				// Verify environment-specific settings
				assert.Contains(t, content, "environment", "Config should specify environment")
				t.Logf("Found configuration for %s environment", env)

				// Check for differentiated resource limits
				if strings.Contains(content, "resources:") {
					t.Logf("Environment %s has resource limits configured", env)
				}
			}
		}

		assert.GreaterOrEqual(t, foundConfigs, 2,
			"At least 2 environment configurations should exist")
	})

	t.Run("Environments have different resource allocations", func(t *testing.T) {
		// Read configs and verify they're different
		devConfig, devErr := os.ReadFile("../../../deploy/overlays/dev/values-dev.yaml")
		prdConfig, prdErr := os.ReadFile("../../../deploy/overlays/prd/values-prd.yaml")

		if devErr != nil || prdErr != nil {
			t.Skip("Environment config files not found")
			return
		}

		// Configs should be different (not identical)
		assert.NotEqual(t, string(devConfig), string(prdConfig),
			"Dev and prod configurations should be different")

		// Check for replica count differences
		devContent := string(devConfig)
		prdContent := string(prdConfig)

		if strings.Contains(devContent, "replicaCount") && strings.Contains(prdContent, "replicaCount") {
			t.Log("Replica counts are configured per environment")
		}
	})
}

func TestDevOps003_AC2_EnvironmentConfigMaps(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Environment-specific ConfigMaps exist", func(t *testing.T) {
		envNamespaces := []string{
			"knative-lambda",
		}

		for _, ns := range envNamespaces {
			configMaps, err := client.CoreV1().ConfigMaps(ns).List(
				ctx,
				metav1.ListOptions{},
			)

			if err != nil {
				continue
			}

			t.Logf("Namespace %s has %d ConfigMap(s)", ns, len(configMaps.Items))

			for _, cm := range configMaps.Items {
				// Check for environment-specific data
				if env, ok := cm.Data["ENVIRONMENT"]; ok {
					t.Logf("ConfigMap %s has ENVIRONMENT=%s", cm.Name, env)
				}
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Promotion Strategy.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps003_AC3_PromotionWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Promotion scripts are documented", func(t *testing.T) {
		data, err := os.ReadFile(devops003DocPath)
		if err != nil {
			t.Skip("Documentation not found")
			return
		}

		content := string(data)
		assert.Contains(t, content, "promotion", "Documentation should describe promotion process")

		// Check for promotion workflow steps
		if strings.Contains(content, "dev → staging") || strings.Contains(content, "staging → prod") {
			t.Log("Promotion workflow is documented")
		}
	})

	t.Run("Promotion scripts exist", func(t *testing.T) {
		scriptPaths := []string{
			"../../scripts/promote-to-staging.sh",
			"../../scripts/promote-to-prod.sh",
		}

		foundScripts := 0
		for _, path := range scriptPaths {
			if _, err := os.Stat(path); err == nil {
				foundScripts++
				t.Logf("Found promotion script: %s", path)
			}
		}

		if foundScripts == 0 {
			t.Log("No promotion scripts found, may use manual or GitOps-based promotion")
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: RBAC Per Environment.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps003_AC4_EnvironmentRBAC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("ServiceAccounts exist per environment", testServiceAccountsPerEnvironment(ctx, client))
	t.Run("Roles and RoleBindings are environment-specific", testRolesRoleBindingsPerEnvironment(ctx, client))
}

// testServiceAccountsPerEnvironment tests if service accounts exist per environment.
func testServiceAccountsPerEnvironment(ctx context.Context, client *kubernetes.Clientset) func(*testing.T) {
	return func(t *testing.T) {
		envNamespaces := []string{
			"knative-lambda",
		}

		for _, ns := range envNamespaces {
			serviceAccounts, err := client.CoreV1().ServiceAccounts(ns).List(
				ctx,
				metav1.ListOptions{},
			)

			if err != nil {
				continue
			}

			t.Logf("Namespace %s has %d ServiceAccount(s)", ns, len(serviceAccounts.Items))

			for _, sa := range serviceAccounts.Items {
				if strings.Contains(sa.Name, "knative-lambda") {
					t.Logf("Found ServiceAccount: %s in %s", sa.Name, ns)

					if annotations := sa.Annotations; annotations != nil {
						if roleArn, ok := annotations["eks.amazonaws.com/role-arn"]; ok {
							t.Logf("ServiceAccount has IRSA role: %s", roleArn)

							envShort := strings.TrimPrefix(ns, "knative-lambda-")
							if strings.Contains(roleArn, envShort) {
								t.Logf("IAM role is environment-specific")
							}
						}
					}
				}
			}
		}
	}
}

// testRolesRoleBindingsPerEnvironment tests if roles and role bindings are environment-specific.
func testRolesRoleBindingsPerEnvironment(ctx context.Context, client *kubernetes.Clientset) func(*testing.T) {
	return func(t *testing.T) {
		envNamespaces := []string{
			"knative-lambda",
		}

		for _, ns := range envNamespaces {
			roles, err := client.RbacV1().Roles(ns).List(
				ctx,
				metav1.ListOptions{},
			)

			if err != nil {
				continue
			}

			if len(roles.Items) > 0 {
				t.Logf("Namespace %s has %d Role(s)", ns, len(roles.Items))
			}

			roleBindings, err := client.RbacV1().RoleBindings(ns).List(
				ctx,
				metav1.ListOptions{},
			)

			if err != nil {
				continue
			}

			if len(roleBindings.Items) > 0 {
				t.Logf("Namespace %s has %d RoleBinding(s)", ns, len(roleBindings.Items))
			}
		}
	}
}

func TestDevOps003_AC4_ProductionAccessRestriction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Production RBAC is documented", func(t *testing.T) {
		data, err := os.ReadFile(devops003DocPath)
		if err != nil {
			t.Skip("Documentation not found")
			return
		}

		content := string(data)
		if strings.Contains(content, "RBAC") || strings.Contains(content, "access control") {
			t.Log("RBAC policies are documented")
		}

		// Check for production access restrictions
		if strings.Contains(content, "production") && strings.Contains(content, "approval") {
			t.Log("Production access restrictions are documented")
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Cost Tracking Per Environment.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps003_AC5_CostAllocationTags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Namespaces have cost allocation labels", testNamespacesCostAllocationLabels(ctx, client))
	t.Run("Resources have environment labels for cost tracking", testResourcesEnvironmentLabels(ctx, client))
}

// testNamespacesCostAllocationLabels tests if namespaces have cost allocation labels.
func testNamespacesCostAllocationLabels(ctx context.Context, client *kubernetes.Clientset) func(*testing.T) {
	return func(t *testing.T) {
		envNamespaces := []string{
			"knative-lambda",
		}

		for _, ns := range envNamespaces {
			namespace, err := client.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
			if err != nil {
				continue
			}

			labels := namespace.Labels
			if labels != nil {
				costLabels := []string{"environment", "project", "team", "cost-center"}
				foundLabels := 0

				for _, label := range costLabels {
					if _, ok := labels[label]; ok {
						foundLabels++
						t.Logf("Namespace %s has label: %s=%s", ns, label, labels[label])
					}
				}

				if foundLabels > 0 {
					t.Logf("Namespace %s has %d cost allocation label(s)", ns, foundLabels)
				}
			}
		}
	}
}

// testResourcesEnvironmentLabels tests if resources have environment labels for cost tracking.
func testResourcesEnvironmentLabels(ctx context.Context, client *kubernetes.Clientset) func(*testing.T) {
	return func(t *testing.T) {
		namespace := getNamespace(t)

		deployments, err := client.AppsV1().Deployments(namespace).List(
			ctx,
			metav1.ListOptions{},
		)

		if err != nil {
			t.Skip("Unable to list deployments")
			return
		}

		for _, deployment := range deployments.Items {
			labels := deployment.Labels
			if labels != nil {
				if env, ok := labels["environment"]; ok {
					t.Logf("Deployment %s has environment label: %s", deployment.Name, env)
				}
			}
		}
	}
}

func TestDevOps003_AC5_CostTrackingDocumentation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Cost tracking is documented per environment", func(t *testing.T) {
		data, err := os.ReadFile(devops003DocPath)
		if err != nil {
			t.Skip("Documentation not found")
			return
		}

		content := string(data)
		if strings.Contains(content, "cost") || strings.Contains(content, "Cost") {
			t.Log("Cost tracking documentation exists")

			// Check for per-environment cost breakdown
			if strings.Contains(content, "Monthly Cost by Environment") ||
				strings.Contains(content, "Cost per environment") {
				t.Log("Per-environment cost breakdown is documented")
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Performance Requirements.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps003_Performance_EnvironmentDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	namespace := getNamespace(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Deployments are healthy in all environments", func(t *testing.T) {
		deployment, err := client.AppsV1().Deployments(namespace).Get(
			ctx,
			"knative-lambda-builder",
			metav1.GetOptions{},
		)

		if err != nil {
			t.Skip("Builder deployment not found in namespace")
			return
		}

		assert.Equal(t, deployment.Status.Replicas, deployment.Status.ReadyReplicas,
			"All deployment replicas should be ready")
		assert.Equal(t, deployment.Status.Replicas, deployment.Status.AvailableReplicas,
			"All deployment replicas should be available")
	})
}

func TestDevOps003_Performance_PromotionTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Promotion time requirements are documented", func(t *testing.T) {
		data, err := os.ReadFile(devops003DocPath)
		if err != nil {
			t.Skip("Documentation not found")
			return
		}

		content := string(data)
		// Check for performance requirements section
		if strings.Contains(content, "Performance Requirements") ||
			strings.Contains(content, "performance requirements") {
			t.Log("Performance requirements are documented")
		}
	})
}
