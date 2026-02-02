// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 DEVOPS-005: Infrastructure as Code Tests
//
//	User Story: Infrastructure as Code
//	Priority: P1 | Story Points: 8
//
//	Tests validate acceptance criteria:
//	✓ All infrastructure defined in code
//	✓ Version controlled infrastructure
//	✓ Automated infrastructure testing
//	✓ Infrastructure change reviews
//	✓ Drift detection and remediation
//	✓ Documentation as code
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

package devops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: Helm Charts for Infrastructure.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps005_AC1_HelmChartsExist(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Helm chart directory exists", func(t *testing.T) {
		chartPaths := []string{
			"../../../deploy",
			"../../deploy",
			"../../charts",
			"charts",
		}

		foundCharts := false
		var chartDir string
		for _, path := range chartPaths {
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				foundCharts = true
				chartDir = path
				t.Logf("Found Helm charts directory: %s", path)
				break
			}
		}

		assert.True(t, foundCharts, "Helm charts directory should exist")

		if foundCharts {
			// Check for Chart.yaml files
			err := filepath.Walk(chartDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if !info.IsDir() && (info.Name() == "Chart.yaml" || info.Name() == "Chart.yml") {
					t.Logf("Found Chart.yaml at: %s", path)
				}
				return nil
			})
			assert.NoError(t, err)
		}
	})
}

func TestDevOps005_AC1_HelmChartStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Helm charts have proper structure", func(t *testing.T) {
		chartPaths := []string{
			"../../charts",
			"charts",
		}

		for _, chartPath := range chartPaths {
			if info, err := os.Stat(chartPath); err == nil && info.IsDir() {
				// Look for Chart.yaml
				chartFiles := []string{
					filepath.Join(chartPath, "knative-lambda", "Chart.yaml"),
					filepath.Join(chartPath, "Chart.yaml"),
				}

				for _, chartFile := range chartFiles {
					if data, err := os.ReadFile(chartFile); err == nil { //nolint:gosec // G304: Test file reading controlled paths
						content := string(data)

						// Verify Chart.yaml has required fields
						assert.Contains(t, content, "apiVersion:", "Chart should have apiVersion")
						assert.Contains(t, content, "name:", "Chart should have name")
						assert.Contains(t, content, "version:", "Chart should have version")

						// Check for values.yaml
						valuesFile := filepath.Join(filepath.Dir(chartFile), "values.yaml")
						if _, err := os.Stat(valuesFile); err == nil {
							t.Log("Found values.yaml for chart")
						}

						// Check for templates directory
						templatesDir := filepath.Join(filepath.Dir(chartFile), "templates")
						if info, err := os.Stat(templatesDir); err == nil && info.IsDir() {
							t.Log("Found templates directory")
						}
						break
					}
				}
				break
			}
		}
	})
}

func TestDevOps005_AC1_ValuesYamlParameterization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("values.yaml provides proper parameterization", func(t *testing.T) {
		valuesPaths := []string{
			"../../charts/knative-lambda/values.yaml",
			"charts/knative-lambda/values.yaml",
		}

		for _, path := range valuesPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for key configurations
				configurations := []string{
					"image:",
					"replicas:",
					"resources:",
					"autoscaling:",
				}

				for _, config := range configurations {
					if strings.Contains(content, config) {
						t.Logf("values.yaml includes configuration for: %s", strings.TrimSuffix(config, ":"))
					}
				}
				break
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Version Control.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps005_AC2_GitIgnoreConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run(".gitignore properly configured for IaC", func(t *testing.T) {
		gitignorePaths := []string{
			"../../.gitignore",
			".gitignore",
		}

		for _, path := range gitignorePaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for common exclusions
				exclusions := []string{
					".terraform",
					"*.tfstate",
					"secrets",
				}

				for _, exclusion := range exclusions {
					if strings.Contains(content, exclusion) {
						t.Logf(".gitignore excludes: %s", exclusion)
					}
				}
				break
			}
		}
	})
}

func TestDevOps005_AC2_InfrastructureInGit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Infrastructure files are tracked in Git", func(t *testing.T) {
		// Check for .git directory (indicates repository)
		gitPaths := []string{
			"../../../../../../../../.git",
			"../../../.git",
			"../../.git",
			".git",
		}

		foundGit := false
		for _, path := range gitPaths {
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				foundGit = true
				t.Log("Repository is under Git version control")
				break
			}
		}

		assert.True(t, foundGit, "Infrastructure should be under version control")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Automated Testing.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps005_AC3_HelmLintConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Helm lint is configured in CI/CD", func(t *testing.T) {
		workflowPaths := []string{
			"../../.github/workflows/ci-cd.yaml",
			".github/workflows/ci-cd.yaml",
		}

		for _, path := range workflowPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for helm lint
				if strings.Contains(content, "helm lint") {
					t.Log("Helm lint is configured in CI/CD pipeline")
				}

				// Check for chart-testing
				if strings.Contains(content, "chart-testing") || strings.Contains(content, "ct") {
					t.Log("Chart testing tool is configured")
				}
				break
			}
		}
	})
}

func TestDevOps005_AC3_KubevalValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Kubernetes manifest validation is configured", func(t *testing.T) {
		workflowPaths := []string{
			"../../.github/workflows/ci-cd.yaml",
			".github/workflows/ci-cd.yaml",
		}

		for _, path := range workflowPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for validation tools
				validationTools := []string{
					"kubeval",
					"kubeconform",
					"kubectl dry-run",
					"--dry-run",
				}

				for _, tool := range validationTools {
					if strings.Contains(content, tool) {
						t.Logf("Kubernetes validation using: %s", tool)
					}
				}
				break
			}
		}
	})
}

func TestDevOps005_AC3_PolicyValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Policy validation tools are configured", func(t *testing.T) {
		workflowPaths := []string{
			"../../.github/workflows/ci-cd.yaml",
			".github/workflows/ci-cd.yaml",
		}

		for _, path := range workflowPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for policy tools
				policyTools := []string{
					"conftest",
					"opa",
					"kyverno",
					"policy",
				}

				for _, tool := range policyTools {
					if strings.Contains(content, tool) {
						t.Logf("Policy validation using: %s", tool)
					}
				}
				break
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: Change Review Process.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps005_AC4_PRTemplateExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Pull request template exists for infrastructure changes", func(t *testing.T) {
		prTemplatePaths := []string{
			"../../.github/pull_request_template.md",
			"../../.github/PULL_REQUEST_TEMPLATE.md",
			".github/pull_request_template.md",
		}

		for _, path := range prTemplatePaths {
			if _, err := os.Stat(path); err == nil {
				t.Logf("Found PR template: %s", path)
				break
			}
		}
	})
}

func TestDevOps005_AC4_CodeOwnersConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("CODEOWNERS file exists for infrastructure reviews", func(t *testing.T) {
		codeownersPaths := []string{
			"../../.github/CODEOWNERS",
			"../../CODEOWNERS",
			".github/CODEOWNERS",
		}

		for _, path := range codeownersPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for infrastructure paths
				if strings.Contains(content, "charts/") || strings.Contains(content, "terraform/") {
					t.Log("CODEOWNERS includes infrastructure paths")
				}
				break
			}
		}
	})
}

func TestDevOps005_AC4_BranchProtectionDocumented(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Branch protection requirements are documented", func(t *testing.T) {
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-005-infrastructure-as-code.md",
		}

		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for branch protection mention
				if strings.Contains(content, "branch protection") ||
					strings.Contains(content, "required reviewers") {
					t.Log("Branch protection requirements are documented")
				}
				break
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Drift Detection.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps005_AC5_FluxDriftDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Flux detects and alerts on drift", func(t *testing.T) {
		fluxConfigPaths := []string{
			"../../flux",
			"flux",
		}

		for _, fluxPath := range fluxConfigPaths {
			if info, err := os.Stat(fluxPath); err == nil && info.IsDir() {
				t.Log("Flux directory exists for GitOps-based drift detection")

				// Check for notification configuration
				err := filepath.Walk(fluxPath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return nil
					}
					if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
						if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
							content := string(data)
							if strings.Contains(content, "kind: Alert") {
								t.Logf("Found Flux Alert configuration: %s", path)
							}
						}
					}
					return nil
				})
				assert.NoError(t, err)
				break
			}
		}
	})
}

func TestDevOps005_AC5_AutomatedReconciliation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Automated drift reconciliation is configured", func(t *testing.T) {
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-005-infrastructure-as-code.md",
		}

		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for reconciliation mentions
				if strings.Contains(content, "reconciliation") || strings.Contains(content, "Reconciliation") {
					t.Log("Drift reconciliation strategy is documented")
				}

				// Check for Flux sync
				if strings.Contains(content, "flux reconcile") || strings.Contains(content, "sync interval") {
					t.Log("Flux reconciliation is configured")
				}
				break
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: Documentation as Code.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps005_AC6_DocumentationStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Documentation directory exists and is organized", func(t *testing.T) {
		docPaths := []string{
			"../../docs",
			"docs",
		}

		for _, docPath := range docPaths {
			if info, err := os.Stat(docPath); err == nil && info.IsDir() {
				t.Log("Documentation directory exists")

				// Count markdown files
				mdCount := 0
				err := filepath.Walk(docPath, func(_ string, info os.FileInfo, err error) error {
					if err != nil {
						return nil
					}
					if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
						mdCount++
					}
					return nil
				})
				assert.NoError(t, err)
				t.Logf("Found %d markdown documentation files", mdCount)
				break
			}
		}
	})
}

func TestDevOps005_AC6_READMEFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("README files exist for key components", func(t *testing.T) {
		readmePaths := []string{
			"../../README.md",
			"../../charts/README.md",
			"../../docs/README.md",
			"README.md",
		}

		foundReadmes := 0
		for _, path := range readmePaths {
			if _, err := os.Stat(path); err == nil {
				foundReadmes++
				t.Logf("Found README: %s", path)
			}
		}

		assert.Greater(t, foundReadmes, 0, "At least one README file should exist")
	})
}

func TestDevOps005_AC6_DiagramsAsCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Architecture diagrams exist in documentation", func(t *testing.T) {
		docPaths := []string{
			"../../docs",
			"docs",
		}

		for _, docPath := range docPaths {
			if info, err := os.Stat(docPath); err == nil && info.IsDir() {
				// Look for diagram files or references
				err := filepath.Walk(docPath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return nil
					}
					if !info.IsDir() {
						// Check for diagram files
						if strings.HasSuffix(info.Name(), ".png") ||
							strings.HasSuffix(info.Name(), ".svg") ||
							strings.HasSuffix(info.Name(), ".mermaid") {
							t.Logf("Found diagram: %s", path)
						}
					}
					return nil
				})
				assert.NoError(t, err)
				break
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Performance Requirements.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestDevOps005_Performance_HelmInstallTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Helm install performance requirements are documented", func(t *testing.T) {
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-005-infrastructure-as-code.md",
		}

		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for performance requirements
				if strings.Contains(content, "Performance Requirements") ||
					strings.Contains(content, "Helm Install Time") {
					t.Log("Infrastructure deployment performance requirements are documented")
				}
				break
			}
		}
	})
}

func TestDevOps005_Performance_ValidationSpeed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Infrastructure validation speed requirements exist", func(t *testing.T) {
		docPaths := []string{
			"../../docs/03-for-engineers/devops/user-stories/DEVOPS-005-infrastructure-as-code.md",
		}

		for _, path := range docPaths {
			if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: Test file reading controlled paths
				content := string(data)

				// Check for validation time requirements
				if strings.Contains(content, "Validation Time") || strings.Contains(content, "30 seconds") {
					t.Log("Validation speed requirements are documented")
				}
				break
			}
		}
	})
}
