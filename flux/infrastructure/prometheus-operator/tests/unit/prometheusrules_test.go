// Package unit contains unit tests for prometheus-operator configurations.
package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brunovlucena/homelab/flux/infrastructure/prometheus-operator/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestPrometheusRulesYAMLSyntax validates that all PrometheusRule files have valid YAML syntax.
func TestPrometheusRulesYAMLSyntax(t *testing.T) {
	rulesDir := filepath.Join(testutils.GetK8sDir(), "prometheusrules")
	if rulesDir == "" {
		t.Skip("Could not find k8s/prometheusrules directory")
	}

	files, err := testutils.FindYAMLFiles(rulesDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			require.NoError(t, err, "Failed to read file: %s", file)

			var content interface{}
			err = yaml.Unmarshal(data, &content)
			assert.NoError(t, err, "Invalid YAML syntax in %s", file)
		})
	}
}

// TestPrometheusRulesStructure validates the structure of PrometheusRule files.
func TestPrometheusRulesStructure(t *testing.T) {
	rulesDir := filepath.Join(testutils.GetK8sDir(), "prometheusrules")
	if rulesDir == "" {
		t.Skip("Could not find k8s/prometheusrules directory")
	}

	files, err := testutils.FindYAMLFiles(rulesDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		// Skip kustomization.yaml and README files
		if strings.Contains(filepath.Base(file), "kustomization") ||
			strings.Contains(filepath.Base(file), "README") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			var rule testutils.PrometheusRule
			err := testutils.LoadYAMLFile(file, &rule)
			require.NoError(t, err, "Failed to parse PrometheusRule")

			// Validate apiVersion
			assert.Equal(t, "monitoring.coreos.com/v1", rule.APIVersion,
				"PrometheusRule should have correct apiVersion")

			// Validate kind
			assert.Equal(t, "PrometheusRule", rule.Kind,
				"Kind should be PrometheusRule")

			// Validate metadata
			assert.NotEmpty(t, rule.Metadata.Name,
				"PrometheusRule must have a name")
			assert.NotEmpty(t, rule.Metadata.Namespace,
				"PrometheusRule must have a namespace")

			// Validate groups exist
			assert.NotEmpty(t, rule.Spec.Groups,
				"PrometheusRule must have at least one group")

			// Validate each group
			for _, group := range rule.Spec.Groups {
				assert.NotEmpty(t, group.Name, "Rule group must have a name")
				assert.NotEmpty(t, group.Rules, "Rule group '%s' must have at least one rule", group.Name)
			}
		})
	}
}

// TestPrometheusRulesAlertStructure validates the structure of alert rules.
func TestPrometheusRulesAlertStructure(t *testing.T) {
	rulesDir := filepath.Join(testutils.GetK8sDir(), "prometheusrules")
	if rulesDir == "" {
		t.Skip("Could not find k8s/prometheusrules directory")
	}

	files, err := testutils.FindYAMLFiles(rulesDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		// Skip kustomization.yaml and README files
		if strings.Contains(filepath.Base(file), "kustomization") ||
			strings.Contains(filepath.Base(file), "README") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			var rule testutils.PrometheusRule
			err := testutils.LoadYAMLFile(file, &rule)
			require.NoError(t, err, "Failed to parse PrometheusRule")

			for _, group := range rule.Spec.Groups {
				for i, r := range group.Rules {
					// Each rule must have either alert or record name
					hasAlert := r.Alert != ""
					hasRecord := r.Record != ""

					assert.True(t, hasAlert || hasRecord,
						"Rule %d in group '%s' must have either 'alert' or 'record' field",
						i, group.Name)

					// Expression is required
					assert.NotEmpty(t, r.Expr,
						"Rule '%s' in group '%s' must have an expression",
						r.Alert+r.Record, group.Name)

					// For alert rules, validate additional requirements
					if hasAlert {
						// Alert rules should have labels
						assert.NotEmpty(t, r.Labels,
							"Alert '%s' should have labels", r.Alert)

						// Check for severity label
						severity, hasSeverity := r.Labels["severity"]
						assert.True(t, hasSeverity,
							"Alert '%s' should have a severity label", r.Alert)

						if hasSeverity {
							err := testutils.ValidateSeverity(severity)
							assert.NoError(t, err,
								"Alert '%s' has invalid severity", r.Alert)
						}

						// Alert rules should have annotations
						assert.NotEmpty(t, r.Annotations,
							"Alert '%s' should have annotations", r.Alert)

						// Check for summary annotation
						_, hasSummary := r.Annotations["summary"]
						assert.True(t, hasSummary,
							"Alert '%s' should have a summary annotation", r.Alert)

						// Check for description annotation
						_, hasDescription := r.Annotations["description"]
						assert.True(t, hasDescription,
							"Alert '%s' should have a description annotation", r.Alert)

						// Validate 'for' duration if present
						if r.For != "" {
							err := testutils.ValidateDuration(r.For)
							assert.NoError(t, err,
								"Alert '%s' has invalid 'for' duration: %s", r.Alert, r.For)
						}
					}
				}
			}
		})
	}
}

// TestPrometheusRulesPromQLSyntax validates PromQL expressions in PrometheusRules.
func TestPrometheusRulesPromQLSyntax(t *testing.T) {
	rulesDir := filepath.Join(testutils.GetK8sDir(), "prometheusrules")
	if rulesDir == "" {
		t.Skip("Could not find k8s/prometheusrules directory")
	}

	files, err := testutils.FindYAMLFiles(rulesDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		// Skip kustomization.yaml and README files
		if strings.Contains(filepath.Base(file), "kustomization") ||
			strings.Contains(filepath.Base(file), "README") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			var rule testutils.PrometheusRule
			err := testutils.LoadYAMLFile(file, &rule)
			require.NoError(t, err, "Failed to parse PrometheusRule")

			for _, group := range rule.Spec.Groups {
				for _, r := range group.Rules {
					ruleName := r.Alert
					if ruleName == "" {
						ruleName = r.Record
					}

					// Basic PromQL syntax validation
					err := testutils.ValidatePromQLSyntax(r.Expr)
					assert.NoError(t, err,
						"Rule '%s' in group '%s' has invalid PromQL syntax",
						ruleName, group.Name)
				}
			}
		})
	}
}

// TestPrometheusRulesNoHelmTemplates ensures no Helm template syntax exists in Kustomize-processed files.
func TestPrometheusRulesNoHelmTemplates(t *testing.T) {
	rulesDir := filepath.Join(testutils.GetK8sDir(), "prometheusrules")
	if rulesDir == "" {
		t.Skip("Could not find k8s/prometheusrules directory")
	}

	files, err := testutils.FindYAMLFiles(rulesDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		// Skip kustomization.yaml and README files
		if strings.Contains(filepath.Base(file), "kustomization") ||
			strings.Contains(filepath.Base(file), "README") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			require.NoError(t, err, "Failed to read file")

			content := string(data)

			// Check for Helm template syntax (excluding Grafana escaping patterns)
			hasHelmTemplate := testutils.ContainsHelmTemplateExcludingGrafana(content)
			assert.False(t, hasHelmTemplate,
				"File %s contains Helm template syntax ({{ }}). "+
					"Kustomize does not process Helm templates. "+
					"Use actual values instead of {{ .Release.Namespace }} etc.",
				filepath.Base(file))
		})
	}
}

// TestPrometheusRulesNamespaceConsistency validates namespace consistency in PrometheusRules.
func TestPrometheusRulesNamespaceConsistency(t *testing.T) {
	rulesDir := filepath.Join(testutils.GetK8sDir(), "prometheusrules")
	if rulesDir == "" {
		t.Skip("Could not find k8s/prometheusrules directory")
	}

	files, err := testutils.FindYAMLFiles(rulesDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		// Skip kustomization.yaml and README files
		if strings.Contains(filepath.Base(file), "kustomization") ||
			strings.Contains(filepath.Base(file), "README") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			var rule testutils.PrometheusRule
			err := testutils.LoadYAMLFile(file, &rule)
			require.NoError(t, err, "Failed to parse PrometheusRule")

			// Namespace should be 'prometheus' for consistency
			assert.Equal(t, "prometheus", rule.Metadata.Namespace,
				"PrometheusRule namespace should be 'prometheus'")
		})
	}
}

// TestPrometheusRulesLabels validates that PrometheusRules have labels.
func TestPrometheusRulesLabels(t *testing.T) {
	rulesDir := filepath.Join(testutils.GetK8sDir(), "prometheusrules")
	if rulesDir == "" {
		t.Skip("Could not find k8s/prometheusrules directory")
	}

	files, err := testutils.FindYAMLFiles(rulesDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		// Skip kustomization.yaml and README files
		if strings.Contains(filepath.Base(file), "kustomization") ||
			strings.Contains(filepath.Base(file), "README") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			var rule testutils.PrometheusRule
			err := testutils.LoadYAMLFile(file, &rule)
			require.NoError(t, err, "Failed to parse PrometheusRule")

			// PrometheusRules should have labels for identification
			labels := rule.Metadata.Labels
			assert.NotNil(t, labels, "PrometheusRule '%s' should have labels", rule.Metadata.Name)
			assert.NotEmpty(t, labels, "PrometheusRule '%s' should have at least one label", rule.Metadata.Name)
		})
	}
}

// TestPrometheusRulesUniqueAlertNames validates that alert names are unique within a file.
func TestPrometheusRulesUniqueAlertNames(t *testing.T) {
	rulesDir := filepath.Join(testutils.GetK8sDir(), "prometheusrules")
	if rulesDir == "" {
		t.Skip("Could not find k8s/prometheusrules directory")
	}

	files, err := testutils.FindYAMLFiles(rulesDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		// Skip kustomization.yaml and README files
		if strings.Contains(filepath.Base(file), "kustomization") ||
			strings.Contains(filepath.Base(file), "README") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			var rule testutils.PrometheusRule
			err := testutils.LoadYAMLFile(file, &rule)
			require.NoError(t, err, "Failed to parse PrometheusRule")

			alertNames := make(map[string]int)
			for _, group := range rule.Spec.Groups {
				for _, r := range group.Rules {
					if r.Alert != "" {
						alertNames[r.Alert]++
					}
				}
			}

			for name, count := range alertNames {
				assert.Equal(t, 1, count,
					"Alert name '%s' appears %d times in %s (should be unique)",
					name, count, filepath.Base(file))
			}
		})
	}
}
