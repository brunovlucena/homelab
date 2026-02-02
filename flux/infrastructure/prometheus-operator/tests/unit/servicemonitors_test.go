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

// TestServiceMonitorsYAMLSyntax validates that all ServiceMonitor files have valid YAML syntax.
func TestServiceMonitorsYAMLSyntax(t *testing.T) {
	monitorsDir := filepath.Join(testutils.GetK8sDir(), "servicemonitors")
	if monitorsDir == "" {
		t.Skip("Could not find k8s/servicemonitors directory")
	}

	files, err := testutils.FindYAMLFiles(monitorsDir, "*.yaml")
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

// TestServiceMonitorsStructure validates the structure of ServiceMonitor files.
func TestServiceMonitorsStructure(t *testing.T) {
	monitorsDir := filepath.Join(testutils.GetK8sDir(), "servicemonitors")
	if monitorsDir == "" {
		t.Skip("Could not find k8s/servicemonitors directory")
	}

	files, err := testutils.FindYAMLFiles(monitorsDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		// Skip kustomization.yaml and README files
		if strings.Contains(filepath.Base(file), "kustomization") ||
			strings.Contains(filepath.Base(file), "README") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			var monitor testutils.ServiceMonitor
			err := testutils.LoadYAMLFile(file, &monitor)
			require.NoError(t, err, "Failed to parse ServiceMonitor")

			// Validate apiVersion
			assert.Equal(t, "monitoring.coreos.com/v1", monitor.APIVersion,
				"ServiceMonitor should have correct apiVersion")

			// Validate kind
			assert.Equal(t, "ServiceMonitor", monitor.Kind,
				"Kind should be ServiceMonitor")

			// Validate metadata
			assert.NotEmpty(t, monitor.Metadata.Name,
				"ServiceMonitor must have a name")
			assert.NotEmpty(t, monitor.Metadata.Namespace,
				"ServiceMonitor must have a namespace")

			// Validate selector
			assert.NotNil(t, monitor.Spec.Selector.MatchLabels,
				"ServiceMonitor must have selector.matchLabels")
			assert.NotEmpty(t, monitor.Spec.Selector.MatchLabels,
				"ServiceMonitor selector.matchLabels cannot be empty")

			// Validate endpoints
			assert.NotEmpty(t, monitor.Spec.Endpoints,
				"ServiceMonitor must have at least one endpoint")
		})
	}
}

// TestServiceMonitorsEndpoints validates endpoint configurations in ServiceMonitors.
func TestServiceMonitorsEndpoints(t *testing.T) {
	monitorsDir := filepath.Join(testutils.GetK8sDir(), "servicemonitors")
	if monitorsDir == "" {
		t.Skip("Could not find k8s/servicemonitors directory")
	}

	files, err := testutils.FindYAMLFiles(monitorsDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		// Skip kustomization.yaml and README files
		if strings.Contains(filepath.Base(file), "kustomization") ||
			strings.Contains(filepath.Base(file), "README") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			var monitor testutils.ServiceMonitor
			err := testutils.LoadYAMLFile(file, &monitor)
			require.NoError(t, err, "Failed to parse ServiceMonitor")

			for i, endpoint := range monitor.Spec.Endpoints {
				// Port is required
				assert.NotEmpty(t, endpoint.Port,
					"Endpoint %d must have a port defined", i)

				// Validate interval if specified
				if endpoint.Interval != "" {
					err := testutils.ValidateDuration(endpoint.Interval)
					assert.NoError(t, err,
						"Endpoint %d has invalid interval: %s", i, endpoint.Interval)
				}

				// Validate scrapeTimeout if specified
				if endpoint.ScrapeTimeout != "" {
					err := testutils.ValidateDuration(endpoint.ScrapeTimeout)
					assert.NoError(t, err,
						"Endpoint %d has invalid scrapeTimeout: %s", i, endpoint.ScrapeTimeout)
				}

				// Path should typically be /metrics or start with /
				if endpoint.Path != "" && !strings.HasPrefix(endpoint.Path, "/") {
					t.Errorf("Endpoint %d path should start with '/': %s", i, endpoint.Path)
				}

				// Validate scheme if specified
				if endpoint.Scheme != "" {
					assert.True(t,
						endpoint.Scheme == "http" || endpoint.Scheme == "https",
						"Endpoint %d has invalid scheme: %s (expected http or https)",
						i, endpoint.Scheme)
				}
			}
		})
	}
}

// TestServiceMonitorsNamespaceSelector validates namespace selector configurations.
func TestServiceMonitorsNamespaceSelector(t *testing.T) {
	monitorsDir := filepath.Join(testutils.GetK8sDir(), "servicemonitors")
	if monitorsDir == "" {
		t.Skip("Could not find k8s/servicemonitors directory")
	}

	files, err := testutils.FindYAMLFiles(monitorsDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		// Skip kustomization.yaml and README files
		if strings.Contains(filepath.Base(file), "kustomization") ||
			strings.Contains(filepath.Base(file), "README") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			var monitor testutils.ServiceMonitor
			err := testutils.LoadYAMLFile(file, &monitor)
			require.NoError(t, err, "Failed to parse ServiceMonitor")

			nsSelector := monitor.Spec.NamespaceSelector

			// Should have either matchNames or any (not both empty unless same namespace)
			hasMatchNames := len(nsSelector.MatchNames) > 0
			hasAny := nsSelector.Any

			// If both are empty, it defaults to monitoring same namespace
			// which is usually fine, but let's check if matchNames has valid values when set
			if hasMatchNames {
				for _, ns := range nsSelector.MatchNames {
					assert.NotEmpty(t, ns,
						"ServiceMonitor '%s' has empty namespace in matchNames",
						monitor.Metadata.Name)
				}
			}

			// Can't have both any=true and matchNames
			if hasAny {
				assert.Empty(t, nsSelector.MatchNames,
					"ServiceMonitor '%s' cannot have both any=true and matchNames",
					monitor.Metadata.Name)
			}
		})
	}
}

// TestServiceMonitorsLabels validates that ServiceMonitors have appropriate labels.
func TestServiceMonitorsLabels(t *testing.T) {
	monitorsDir := filepath.Join(testutils.GetK8sDir(), "servicemonitors")
	if monitorsDir == "" {
		t.Skip("Could not find k8s/servicemonitors directory")
	}

	files, err := testutils.FindYAMLFiles(monitorsDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		// Skip kustomization.yaml and README files
		if strings.Contains(filepath.Base(file), "kustomization") ||
			strings.Contains(filepath.Base(file), "README") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			var monitor testutils.ServiceMonitor
			err := testutils.LoadYAMLFile(file, &monitor)
			require.NoError(t, err, "Failed to parse ServiceMonitor")

			// ServiceMonitors should have labels for identification
			assert.NotNil(t, monitor.Metadata.Labels,
				"ServiceMonitor '%s' should have labels", monitor.Metadata.Name)

			// Check for standard Kubernetes labels
			labels := monitor.Metadata.Labels
			_, hasNameLabel := labels["app.kubernetes.io/name"]
			_, hasLegacyAppLabel := labels["app"]

			assert.True(t, hasNameLabel || hasLegacyAppLabel,
				"ServiceMonitor '%s' should have 'app.kubernetes.io/name' or 'app' label",
				monitor.Metadata.Name)
		})
	}
}

// TestServiceMonitorsSelectorMatchesTarget validates that selector labels look reasonable.
func TestServiceMonitorsSelectorMatchesTarget(t *testing.T) {
	monitorsDir := filepath.Join(testutils.GetK8sDir(), "servicemonitors")
	if monitorsDir == "" {
		t.Skip("Could not find k8s/servicemonitors directory")
	}

	files, err := testutils.FindYAMLFiles(monitorsDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		// Skip kustomization.yaml and README files
		if strings.Contains(filepath.Base(file), "kustomization") ||
			strings.Contains(filepath.Base(file), "README") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			var monitor testutils.ServiceMonitor
			err := testutils.LoadYAMLFile(file, &monitor)
			require.NoError(t, err, "Failed to parse ServiceMonitor")

			matchLabels := monitor.Spec.Selector.MatchLabels

			// Selector should have at least one label
			assert.NotEmpty(t, matchLabels,
				"ServiceMonitor '%s' selector.matchLabels should not be empty",
				monitor.Metadata.Name)

			// Check that labels don't have empty values
			for key, value := range matchLabels {
				assert.NotEmpty(t, key,
					"ServiceMonitor '%s' has empty label key", monitor.Metadata.Name)
				assert.NotEmpty(t, value,
					"ServiceMonitor '%s' has empty value for label '%s'",
					monitor.Metadata.Name, key)
			}
		})
	}
}

// TestPodMonitorsYAMLSyntax validates that all PodMonitor files have valid YAML syntax.
func TestPodMonitorsYAMLSyntax(t *testing.T) {
	monitorsDir := filepath.Join(testutils.GetK8sDir(), "podmonitors")
	if monitorsDir == "" {
		t.Skip("Could not find k8s/podmonitors directory")
	}

	files, err := testutils.FindYAMLFiles(monitorsDir, "*.yaml")
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

// TestPodMonitorsStructure validates the structure of PodMonitor files.
func TestPodMonitorsStructure(t *testing.T) {
	monitorsDir := filepath.Join(testutils.GetK8sDir(), "podmonitors")
	if monitorsDir == "" {
		t.Skip("Could not find k8s/podmonitors directory")
	}

	files, err := testutils.FindYAMLFiles(monitorsDir, "*.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range files {
		// Skip kustomization.yaml and README files
		if strings.Contains(filepath.Base(file), "kustomization") ||
			strings.Contains(filepath.Base(file), "README") {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			var monitor testutils.PodMonitor
			err := testutils.LoadYAMLFile(file, &monitor)
			require.NoError(t, err, "Failed to parse PodMonitor")

			// Validate apiVersion
			assert.Equal(t, "monitoring.coreos.com/v1", monitor.APIVersion,
				"PodMonitor should have correct apiVersion")

			// Validate kind
			assert.Equal(t, "PodMonitor", monitor.Kind,
				"Kind should be PodMonitor")

			// Validate metadata
			assert.NotEmpty(t, monitor.Metadata.Name,
				"PodMonitor must have a name")
			assert.NotEmpty(t, monitor.Metadata.Namespace,
				"PodMonitor must have a namespace")

			// Validate selector - must have either matchLabels or matchExpressions
			hasMatchLabels := len(monitor.Spec.Selector.MatchLabels) > 0
			hasMatchExpressions := len(monitor.Spec.Selector.MatchExpressions) > 0
			assert.True(t, hasMatchLabels || hasMatchExpressions,
				"PodMonitor must have selector.matchLabels or selector.matchExpressions")

			// Validate endpoints
			assert.NotEmpty(t, monitor.Spec.PodMetricsEndpoints,
				"PodMonitor must have at least one podMetricsEndpoint")
		})
	}
}
