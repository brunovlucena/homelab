// Package unit contains unit tests for prometheus-operator configurations.
package unit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brunovlucena/homelab/flux/infrastructure/prometheus-operator/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestDashboardsJSONSyntax validates that all dashboard JSON files have valid JSON syntax.
func TestDashboardsJSONSyntax(t *testing.T) {
	dashboardsDir := filepath.Join(testutils.GetK8sDir(), "dashboards")
	if dashboardsDir == "" {
		t.Skip("Could not find k8s/dashboards directory")
	}

	files, err := testutils.FindJSONFiles(dashboardsDir, "*.json")
	require.NoError(t, err, "Failed to find JSON files")

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			require.NoError(t, err, "Failed to read file: %s", file)

			var content interface{}
			err = json.Unmarshal(data, &content)
			assert.NoError(t, err, "Invalid JSON syntax in %s", file)
		})
	}
}

// TestDashboardsStructure validates the basic structure of Grafana dashboard JSON files.
func TestDashboardsStructure(t *testing.T) {
	dashboardsDir := filepath.Join(testutils.GetK8sDir(), "dashboards")
	if dashboardsDir == "" {
		t.Skip("Could not find k8s/dashboards directory")
	}

	files, err := testutils.FindJSONFiles(dashboardsDir, "*.json")
	require.NoError(t, err, "Failed to find JSON files")

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			var dashboard testutils.GrafanaDashboard
			err := testutils.LoadJSONFile(file, &dashboard)
			require.NoError(t, err, "Failed to parse dashboard JSON")

			// Dashboard should have a title
			assert.NotEmpty(t, dashboard.Title,
				"Dashboard must have a title")

			// Dashboard should have panels
			assert.NotEmpty(t, dashboard.Panels,
				"Dashboard must have at least one panel")

			// Check schemaVersion (modern dashboards use higher versions)
			assert.Greater(t, dashboard.SchemaVersion, 0,
				"Dashboard should have a valid schemaVersion")
		})
	}
}

// TestDashboardsPanelsValid validates panel configurations in dashboards.
func TestDashboardsPanelsValid(t *testing.T) {
	dashboardsDir := filepath.Join(testutils.GetK8sDir(), "dashboards")
	if dashboardsDir == "" {
		t.Skip("Could not find k8s/dashboards directory")
	}

	files, err := testutils.FindJSONFiles(dashboardsDir, "*.json")
	require.NoError(t, err, "Failed to find JSON files")

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			var dashboard testutils.GrafanaDashboard
			err := testutils.LoadJSONFile(file, &dashboard)
			require.NoError(t, err, "Failed to parse dashboard JSON")

			validatePanels(t, dashboard.Panels, filepath.Base(file))
		})
	}
}

// validatePanels recursively validates panels including nested panels in rows.
func validatePanels(t *testing.T, panels []testutils.Panel, filename string) {
	for i, panel := range panels {
		// Panel type is required
		assert.NotEmpty(t, panel.Type,
			"Panel %d in %s must have a type", i, filename)

		// Check for valid panel types
		validTypes := map[string]bool{
			"row":         true,
			"graph":       true,
			"stat":        true,
			"gauge":       true,
			"timeseries":  true,
			"table":       true,
			"text":        true,
			"singlestat":  true,
			"bargauge":    true,
			"barchart":    true,
			"piechart":    true,
			"heatmap":     true,
			"logs":        true,
			"news":        true,
			"alertlist":   true,
			"dashlist":    true,
			"pluginlist":  true,
			"nodeGraph":   true,
			"histogram":   true,
			"state-timeline": true,
			"status-history": true,
			"canvas":      true,
			"candlestick": true,
			"geomap":      true,
			"flamegraph":  true,
			"traces":      true,
		}

		assert.True(t, validTypes[panel.Type],
			"Panel %d in %s has unknown type: %s", i, filename, panel.Type)

		// Row panels can have nested panels
		if panel.Type == "row" && len(panel.Panels) > 0 {
			validatePanels(t, panel.Panels, filename)
		}

		// For non-row panels with targets, validate targets
		if panel.Type != "row" && len(panel.Targets) > 0 {
			for j, target := range panel.Targets {
				// RefID should be present
				assert.NotEmpty(t, target.RefID,
					"Target %d in panel %d (%s) should have a refId",
					j, i, filename)

				// If expr is present, do basic validation
				if target.Expr != "" {
					err := testutils.ValidatePromQLSyntax(target.Expr)
					assert.NoError(t, err,
						"Target %d in panel %d (%s) has invalid PromQL syntax",
						j, i, filename)
				}
			}
		}
	}
}

// TestDashboardsDataSources validates datasource references in dashboards.
func TestDashboardsDataSources(t *testing.T) {
	dashboardsDir := filepath.Join(testutils.GetK8sDir(), "dashboards")
	if dashboardsDir == "" {
		t.Skip("Could not find k8s/dashboards directory")
	}

	files, err := testutils.FindJSONFiles(dashboardsDir, "*.json")
	require.NoError(t, err, "Failed to find JSON files")

	// Valid datasource types
	validDatasourceTypes := map[string]bool{
		"prometheus": true,
		"loki":       true,
		"tempo":      true,
		"grafana":    true,
		"alertmanager": true,
		"elasticsearch": true,
		"influxdb":   true,
		"graphite":   true,
		"cloudwatch": true,
		"stackdriver": true,
		"google-analytics-data-source": true,
		"blackcowmoo-googleanalytics-datasource": true,
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			var dashboard testutils.GrafanaDashboard
			err := testutils.LoadJSONFile(file, &dashboard)
			require.NoError(t, err, "Failed to parse dashboard JSON")

			for i, panel := range dashboard.Panels {
				if panel.Datasource != nil {
					validateDatasource(t, panel.Datasource, validDatasourceTypes, i, filepath.Base(file))
				}

				for j, target := range panel.Targets {
					if target.Datasource != nil {
						validateDatasource(t, target.Datasource, validDatasourceTypes, j, filepath.Base(file))
					}
				}
			}
		})
	}
}

// validateDatasource validates a datasource reference.
func validateDatasource(t *testing.T, ds interface{}, validTypes map[string]bool, index int, filename string) {
	switch v := ds.(type) {
	case map[string]interface{}:
		if dsType, ok := v["type"].(string); ok {
			assert.True(t, validTypes[dsType],
				"Unknown datasource type '%s' in panel/target %d of %s",
				dsType, index, filename)
		}
	case string:
		// String datasource references are valid (just the name)
		assert.NotEmpty(t, v,
			"Empty datasource reference in panel/target %d of %s",
			index, filename)
	}
}

// TestDashboardConfigMapsYAMLSyntax validates ConfigMap YAML files containing dashboards.
func TestDashboardConfigMapsYAMLSyntax(t *testing.T) {
	dashboardsDir := filepath.Join(testutils.GetK8sDir(), "dashboards")
	if dashboardsDir == "" {
		t.Skip("Could not find k8s/dashboards directory")
	}

	files, err := testutils.FindYAMLFiles(dashboardsDir, "*-configmap.yaml")
	require.NoError(t, err, "Failed to find ConfigMap YAML files")

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

// TestDashboardConfigMapsStructure validates ConfigMap structure for dashboards.
func TestDashboardConfigMapsStructure(t *testing.T) {
	dashboardsDir := filepath.Join(testutils.GetK8sDir(), "dashboards")
	if dashboardsDir == "" {
		t.Skip("Could not find k8s/dashboards directory")
	}

	files, err := testutils.FindYAMLFiles(dashboardsDir, "*-configmap.yaml")
	require.NoError(t, err, "Failed to find ConfigMap YAML files")

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			var cm testutils.ConfigMap
			err := testutils.LoadYAMLFile(file, &cm)
			require.NoError(t, err, "Failed to parse ConfigMap")

			// Validate apiVersion
			assert.Equal(t, "v1", cm.APIVersion,
				"ConfigMap should have apiVersion: v1")

			// Validate kind
			assert.Equal(t, "ConfigMap", cm.Kind,
				"Kind should be ConfigMap")

			// Validate metadata
			assert.NotEmpty(t, cm.Metadata.Name,
				"ConfigMap must have a name")
			assert.NotEmpty(t, cm.Metadata.Namespace,
				"ConfigMap must have a namespace")

			// Dashboard ConfigMaps should have the grafana_dashboard label
			labels := cm.Metadata.Labels
			if labels != nil {
				_, hasGrafanaLabel := labels["grafana_dashboard"]
				assert.True(t, hasGrafanaLabel,
					"Dashboard ConfigMap '%s' should have 'grafana_dashboard' label for Grafana sidecar discovery",
					cm.Metadata.Name)
			}

			// Validate data contains JSON
			assert.NotEmpty(t, cm.Data,
				"ConfigMap must have data")

			for key, value := range cm.Data {
				// Dashboard JSON files should end with .json
				if strings.HasSuffix(key, ".json") {
					var jsonContent interface{}
					err := json.Unmarshal([]byte(value), &jsonContent)
					assert.NoError(t, err,
						"ConfigMap data '%s' should contain valid JSON", key)
				}
			}
		})
	}
}

// TestDashboardConfigMapsEmbeddedJSON validates the embedded JSON in dashboard ConfigMaps.
func TestDashboardConfigMapsEmbeddedJSON(t *testing.T) {
	dashboardsDir := filepath.Join(testutils.GetK8sDir(), "dashboards")
	if dashboardsDir == "" {
		t.Skip("Could not find k8s/dashboards directory")
	}

	files, err := testutils.FindYAMLFiles(dashboardsDir, "*-configmap.yaml")
	require.NoError(t, err, "Failed to find ConfigMap YAML files")

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			var cm testutils.ConfigMap
			err := testutils.LoadYAMLFile(file, &cm)
			require.NoError(t, err, "Failed to parse ConfigMap")

			for key, value := range cm.Data {
				if strings.HasSuffix(key, ".json") {
					t.Run(key, func(t *testing.T) {
						var dashboard testutils.GrafanaDashboard
						err := json.Unmarshal([]byte(value), &dashboard)
						require.NoError(t, err, "Failed to parse embedded dashboard JSON")

						// Validate dashboard structure
						assert.NotEmpty(t, dashboard.Title,
							"Embedded dashboard must have a title")
						assert.NotEmpty(t, dashboard.Panels,
							"Embedded dashboard must have panels")
					})
				}
			}
		})
	}
}

// TestDashboardsNoHelmTemplates ensures no Helm template syntax exists in dashboard files.
func TestDashboardsNoHelmTemplates(t *testing.T) {
	dashboardsDir := filepath.Join(testutils.GetK8sDir(), "dashboards")
	if dashboardsDir == "" {
		t.Skip("Could not find k8s/dashboards directory")
	}

	// Check YAML ConfigMaps
	yamlFiles, err := testutils.FindYAMLFiles(dashboardsDir, "*-configmap.yaml")
	require.NoError(t, err, "Failed to find YAML files")

	for _, file := range yamlFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			require.NoError(t, err, "Failed to read file")

			content := string(data)

			// Check for Helm template syntax (excluding Grafana escaping)
			hasHelmTemplate := testutils.ContainsHelmTemplateExcludingGrafana(content)
			assert.False(t, hasHelmTemplate,
				"File %s contains Helm template syntax. "+
					"Kustomize does not process Helm templates.",
				filepath.Base(file))
		})
	}
}

// TestDashboardsMetricsFormat validates that metrics fields are properly formatted.
// This catches issues like PR #294 where metrics was a string instead of array for Prometheus dashboards.
// Note: Some datasources (like Google Analytics) legitimately use string format for metrics.
func TestDashboardsMetricsFormat(t *testing.T) {
	dashboardsDir := filepath.Join(testutils.GetK8sDir(), "dashboards")
	if dashboardsDir == "" {
		t.Skip("Could not find k8s/dashboards directory")
	}

	files, err := testutils.FindJSONFiles(dashboardsDir, "*.json")
	require.NoError(t, err, "Failed to find JSON files")

	// Skip dashboards that use datasources with string metrics
	skipFiles := map[string]bool{
		"google-analytics-dashboard.json": true, // GA datasource uses string metrics
	}

	for _, file := range files {
		if skipFiles[filepath.Base(file)] {
			continue
		}

		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			require.NoError(t, err, "Failed to read file")

			var rawDashboard map[string]interface{}
			err = json.Unmarshal(data, &rawDashboard)
			require.NoError(t, err, "Failed to parse JSON")

			// Recursively check for metrics fields in Prometheus-style dashboards
			checkMetricsFields(t, rawDashboard, filepath.Base(file), "root")
		})
	}
}

// checkMetricsFields recursively checks metrics fields in a JSON structure.
func checkMetricsFields(t *testing.T, obj map[string]interface{}, filename, path string) {
	for key, value := range obj {
		currentPath := path + "." + key

		if key == "metrics" {
			// metrics should be an array, not a string
			switch v := value.(type) {
			case []interface{}:
				// This is correct - metrics should be an array
			case string:
				t.Errorf("File %s: 'metrics' at %s is a string but should be an array. Value: %s",
					filename, currentPath, v)
			default:
				// Could be nil or other type, which might be okay
			}
		}

		// Recursively check nested objects
		switch v := value.(type) {
		case map[string]interface{}:
			checkMetricsFields(t, v, filename, currentPath)
		case []interface{}:
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					checkMetricsFields(t, itemMap, filename, currentPath+"["+string(rune('0'+i))+"]")
				}
			}
		}
	}
}

// TestDashboardsUIDUniqueness validates that dashboard UIDs are unique across all dashboards.
func TestDashboardsUIDUniqueness(t *testing.T) {
	dashboardsDir := filepath.Join(testutils.GetK8sDir(), "dashboards")
	if dashboardsDir == "" {
		t.Skip("Could not find k8s/dashboards directory")
	}

	files, err := testutils.FindJSONFiles(dashboardsDir, "*.json")
	require.NoError(t, err, "Failed to find JSON files")

	uidMap := make(map[string]string) // UID -> filename

	for _, file := range files {
		var dashboard testutils.GrafanaDashboard
		err := testutils.LoadJSONFile(file, &dashboard)
		require.NoError(t, err, "Failed to parse dashboard: %s", file)

		if dashboard.UID != "" {
			if existingFile, exists := uidMap[dashboard.UID]; exists {
				t.Errorf("Duplicate dashboard UID '%s' found in %s and %s",
					dashboard.UID, filepath.Base(existingFile), filepath.Base(file))
			} else {
				uidMap[dashboard.UID] = file
			}
		}
	}
}

// TestDashboardsTitleUniqueness validates that dashboard titles are unique.
func TestDashboardsTitleUniqueness(t *testing.T) {
	dashboardsDir := filepath.Join(testutils.GetK8sDir(), "dashboards")
	if dashboardsDir == "" {
		t.Skip("Could not find k8s/dashboards directory")
	}

	files, err := testutils.FindJSONFiles(dashboardsDir, "*.json")
	require.NoError(t, err, "Failed to find JSON files")

	titleMap := make(map[string]string) // Title -> filename

	for _, file := range files {
		var dashboard testutils.GrafanaDashboard
		err := testutils.LoadJSONFile(file, &dashboard)
		require.NoError(t, err, "Failed to parse dashboard: %s", file)

		if dashboard.Title != "" {
			if existingFile, exists := titleMap[dashboard.Title]; exists {
				t.Errorf("Duplicate dashboard title '%s' found in %s and %s",
					dashboard.Title, filepath.Base(existingFile), filepath.Base(file))
			} else {
				titleMap[dashboard.Title] = file
			}
		}
	}
}
