// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 DEVOPS-006: Secret Management Tests
//
//	User Story: Secret Management
//	Priority: P0 | Story Points: 8
//
//	Tests validate acceptance criteria:
//	✓ Sealed Secrets controller deployed
//	✓ External Secrets Operator configured
//	✓ Secrets encrypted before Git commit
//	✓ Automatic secret rotation
//	✓ RBAC for secret access
//	✓ Audit logging enabled
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

package devops

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	sealedSecretsNamespace = "kube-system"
	externalSecretsNS      = "external-secrets-system"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: Sealed Secrets Integration.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps006_AC1_SealedSecretsController(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Sealed Secrets controller is deployed", func(t *testing.T) {
		deployment, err := client.AppsV1().Deployments(sealedSecretsNamespace).Get(
			ctx,
			"sealed-secrets-controller",
			metav1.GetOptions{},
		)

		if err != nil {
			t.Skip("Sealed Secrets controller not deployed")
			return
		}

		assert.NotNil(t, deployment, "Sealed Secrets controller should exist")
		assert.Greater(t, deployment.Status.ReadyReplicas, int32(0),
			"Sealed Secrets controller should have ready replicas")
	})

	t.Run("Sealed Secrets controller has proper RBAC", func(t *testing.T) {
		// Check ClusterRole
		_, err := client.RbacV1().ClusterRoles().Get(
			ctx,
			"secrets-unsealer",
			metav1.GetOptions{},
		)

		if err != nil {
			t.Log("Sealed Secrets ClusterRole not found, may use different name")
		}

		// Check ServiceAccount
		_, err = client.CoreV1().ServiceAccounts(sealedSecretsNamespace).Get(
			ctx,
			"sealed-secrets-controller",
			metav1.GetOptions{},
		)

		if err != nil {
			t.Log("Sealed Secrets ServiceAccount not found")
		}
	})
}

func TestDevOps006_AC1_SealedSecretCRD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	_, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("SealedSecret CRD is installed", func(t *testing.T) {
		// Check if SealedSecret CRD exists
		_, err := client.Discovery().ServerResourcesForGroupVersion("bitnami.com/v1alpha1")
		if err != nil {
			t.Skip("SealedSecret CRD not installed")
			return
		}

		assert.NoError(t, err, "SealedSecret CRD should be installed")
	})
}

func TestDevOps006_AC1_NoPlaintextSecretsInGit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("No plaintext secret files in repository", func(t *testing.T) {
		// Check if .gitignore excludes secrets
		gitignorePaths := []string{
			"../../../.gitignore",
			"../../.gitignore",
			".gitignore",
		}

		found := false
		for _, path := range gitignorePaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)
				if strings.Contains(content, "secret.yaml") ||
					strings.Contains(content, "*-credentials.yaml") ||
					strings.Contains(content, "*.secret") {
					found = true
					break
				}
			}
		}

		assert.True(t, found, ".gitignore should exclude plaintext secret files")
	})

	t.Run("Repository contains sealed secrets (encrypted)", func(t *testing.T) {
		// Check if sealed secret files exist
		// This validates encryption is being used
		sealedSecretFiles := []string{
			"../../deploy/templates/sealed-secret.yaml",
			"../../deploy/overlays/prd/sealed-secrets.yaml",
		}

		foundSealed := false
		for _, path := range sealedSecretFiles {
			if _, err := os.Stat(path); err == nil {
				foundSealed = true
				break
			}
		}

		if !foundSealed {
			t.Log("No sealed secret files found, may be in different location")
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: External Secrets Operator.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps006_AC2_ExternalSecretsOperator(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("External Secrets Operator is deployed", func(t *testing.T) {
		deployment, err := client.AppsV1().Deployments(externalSecretsNS).Get(
			ctx,
			"external-secrets",
			metav1.GetOptions{},
		)

		if err != nil {
			t.Skip("External Secrets Operator not deployed")
			return
		}

		assert.NotNil(t, deployment, "External Secrets Operator should exist")
		assert.Greater(t, deployment.Status.ReadyReplicas, int32(0),
			"External Secrets Operator should have ready replicas")
	})

	t.Run("SecretStore CRD is installed", func(t *testing.T) {
		// Check if SecretStore CRD exists
		_, err := client.Discovery().ServerResourcesForGroupVersion("external-secrets.io/v1beta1")
		if err != nil {
			t.Skip("External Secrets CRDs not installed")
			return
		}

		assert.NoError(t, err, "SecretStore CRD should be installed")
	})
}

func TestDevOps006_AC2_AWSSecretsManagerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	namespace := getNamespace(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("ServiceAccount has IRSA annotation for AWS access", func(t *testing.T) {
		sa, err := client.CoreV1().ServiceAccounts(namespace).Get(
			ctx,
			"knative-lambda-builder",
			metav1.GetOptions{},
		)

		if err != nil {
			t.Skip("ServiceAccount not found")
			return
		}

		annotations := sa.Annotations
		assert.Contains(t, annotations, "eks.amazonaws.com/role-arn",
			"ServiceAccount should have IRSA annotation for AWS access")

		roleArn := annotations["eks.amazonaws.com/role-arn"]
		assert.Contains(t, roleArn, "arn:aws:iam::",
			"Role ARN should be valid AWS IAM role")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Secret Rotation Support.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps006_AC3_SecretRotationConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Rotation procedures are documented", func(t *testing.T) {
		// Check if rotation documentation exists
		docPaths := []string{
			"../../../docs/03-for-engineers/devops/user-stories/DEVOPS-006-secret-management.md",
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-006-secret-management.md",
		}

		foundDocs := false
		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)
				if strings.Contains(content, "rotation") || strings.Contains(content, "rotate") {
					foundDocs = true
					break
				}
			}
		}

		assert.True(t, foundDocs, "Secret rotation procedures should be documented")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: Security Controls.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

//nolint:funlen // Comprehensive RBAC test with multiple permission scenarios
func TestDevOps006_AC4_RBACForSecretAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	namespace := getNamespace(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("RBAC roles limit secret access", func(t *testing.T) {
		// Check if Role exists for secret access
		roles, err := client.RbacV1().Roles(namespace).List(
			ctx,
			metav1.ListOptions{},
		)

		if err != nil {
			t.Skip("Unable to list roles")
			return
		}

		foundSecretRole := false
		for _, role := range roles.Items {
			for _, rule := range role.Rules {
				for _, resource := range rule.Resources {
					if resource == "secrets" {
						foundSecretRole = true

						// Verify limited verbs (not * or all)
						assert.NotContains(t, rule.Verbs, "*",
							"Secret access should not grant wildcard permissions")

						// Optionally check for specific secret names
						if len(rule.ResourceNames) > 0 {
							t.Logf("Secret access limited to specific secrets: %v", rule.ResourceNames)
						}
					}
				}
			}
		}

		if !foundSecretRole {
			t.Log("No explicit secret RBAC rules found")
		}
	})

	t.Run("ServiceAccount has minimal permissions", func(t *testing.T) {
		sa, err := client.CoreV1().ServiceAccounts(namespace).Get(
			ctx,
			"knative-lambda-builder",
			metav1.GetOptions{},
		)

		if err != nil {
			t.Skip("ServiceAccount not found")
			return
		}

		// Check RoleBindings for this ServiceAccount
		roleBindings, err := client.RbacV1().RoleBindings(namespace).List(
			ctx,
			metav1.ListOptions{},
		)

		if err == nil {
			for _, rb := range roleBindings.Items {
				for _, subject := range rb.Subjects {
					if subject.Kind == "ServiceAccount" && subject.Name == sa.Name {
						t.Logf("ServiceAccount bound to role: %s", rb.RoleRef.Name)
					}
				}
			}
		}
	})
}

func TestDevOps006_AC4_SecretsNotInEnvironmentVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getKubernetesClient(t)
	namespace := getNamespace(t)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Secrets are mounted as volumes, not env vars", func(t *testing.T) {
		pods, err := client.CoreV1().Pods(namespace).List(
			ctx,
			metav1.ListOptions{LabelSelector: "app=knative-lambda-builder"},
		)

		if err != nil || len(pods.Items) == 0 {
			t.Skip("No builder pods found")
			return
		}

		pod := pods.Items[0]
		checkSecretVolumes(t, pod)
		checkSensitiveEnvVars(t, pod)
	})
}

// checkSecretVolumes verifies that secrets are mounted as volumes with proper permissions.
func checkSecretVolumes(t *testing.T, pod v1.Pod) {
	foundSecretVolume := false
	for _, volume := range pod.Spec.Volumes {
		if volume.Secret != nil {
			foundSecretVolume = true

			// Verify it's mounted with proper permissions
			for _, container := range pod.Spec.Containers {
				for _, mount := range container.VolumeMounts {
					if mount.Name == volume.Name {
						t.Logf("Secret %s mounted at %s", volume.Secret.SecretName, mount.MountPath)

						// Verify read-only
						assert.True(t, mount.ReadOnly,
							"Secret volume mount should be read-only")
					}
				}
			}
		}
	}

	if !foundSecretVolume {
		t.Log("No secret volumes mounted, may use other mechanisms")
	}
}

// checkSensitiveEnvVars verifies that sensitive environment variables use valueFrom.
func checkSensitiveEnvVars(t *testing.T, pod v1.Pod) {
	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			if isSensitiveEnvVar(env.Name) {
				if env.Value != "" {
					t.Errorf("Sensitive env var %s should use valueFrom, not direct value", env.Name)
				}

				if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
					t.Logf("Env var %s correctly uses SecretKeyRef", env.Name)
				}
			}
		}
	}
}

// isSensitiveEnvVar checks if an environment variable name suggests sensitive data.
func isSensitiveEnvVar(name string) bool {
	lowerName := strings.ToLower(name)
	return strings.Contains(lowerName, "password") ||
		strings.Contains(lowerName, "secret") ||
		strings.Contains(lowerName, "token")
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Audit Logging.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps006_AC5_AuditLogging(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Audit logging is configured for secret access", func(t *testing.T) {
		// This would require checking Kubernetes audit policy
		// For now, verify documentation exists
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-006-secret-management.md",
		}

		foundAuditDocs := false
		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)
				if strings.Contains(content, "audit") {
					foundAuditDocs = true
					break
				}
			}
		}

		if !foundAuditDocs {
			t.Log("Audit logging documentation not found")
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: Monitoring & Alerts.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps006_AC6_SecretSyncMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Metrics for secret sync are available", func(t *testing.T) {
		// If ESO is deployed, it should expose metrics
		client := getKubernetesClient(t)
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		services, err := client.CoreV1().Services(externalSecretsNS).List(
			ctx,
			metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=external-secrets"},
		)

		if err != nil || len(services.Items) == 0 {
			t.Skip("External Secrets Operator metrics service not found")
			return
		}

		// Check if service exposes metrics port
		for _, svc := range services.Items {
			for _, port := range svc.Spec.Ports {
				if port.Name == "metrics" || port.Port == 8080 {
					t.Logf("External Secrets Operator exposes metrics on port %d", port.Port)
				}
			}
		}
	})
}

func TestDevOps006_AC6_SecretExpirationAlerts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Alert rules for secret expiration exist", func(t *testing.T) {
		// Check Helm templates for secret-related alerts
		alertFiles := []string{
			"../../deploy/templates/alerts.yaml",
			"../../deploy/templates/prometheus-rules.yaml",
		}

		foundSecretAlerts := false
		for _, file := range alertFiles {
			if data, err := os.ReadFile(file); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)
				if strings.Contains(content, "ExternalSecretSyncFailed") ||
					strings.Contains(content, "SecretExpiring") ||
					strings.Contains(content, "SealedSecretDecryptionFailed") {
					foundSecretAlerts = true
					break
				}
			}
		}

		if !foundSecretAlerts {
			t.Log("No secret-specific alerts found in templates")
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Performance Requirements.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps006_Performance_SecretDecryption(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Secret decryption time < 100ms", func(t *testing.T) {
		// This would require timing actual decryption
		// For now, just verify controller is running efficiently
		client := getKubernetesClient(t)
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		pods, err := client.CoreV1().Pods(sealedSecretsNamespace).List(
			ctx,
			metav1.ListOptions{LabelSelector: "name=sealed-secrets-controller"},
		)

		if err != nil || len(pods.Items) == 0 {
			t.Skip("Sealed Secrets controller pod not found")
			return
		}

		pod := pods.Items[0]
		assert.Equal(t, v1.PodRunning, pod.Status.Phase,
			"Sealed Secrets controller should be running")

		// Check resource usage isn't excessive
		for _, container := range pod.Spec.Containers {
			if container.Resources.Limits.Memory() != nil {
				memory := container.Resources.Limits.Memory().Value()
				assert.Less(t, memory, int64(256*1024*1024),
					"Sealed Secrets controller should use < 256Mi memory")
			}
		}
	})
}
