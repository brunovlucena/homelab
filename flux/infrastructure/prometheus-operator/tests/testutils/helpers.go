// Package testutils provides utility functions for prometheus-operator unit tests.
package testutils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// PrometheusRule represents the structure of a Prometheus Rule YAML file.
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

// RuleGroup represents a group of Prometheus rules.
type RuleGroup struct {
	Name  string `yaml:"name"`
	Rules []Rule `yaml:"rules"`
}

// Rule represents a single Prometheus alert or recording rule.
type Rule struct {
	Alert       string            `yaml:"alert,omitempty"`
	Record      string            `yaml:"record,omitempty"`
	Expr        string            `yaml:"expr"`
	For         string            `yaml:"for,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// ServiceMonitor represents the structure of a ServiceMonitor YAML file.
type ServiceMonitor struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string            `yaml:"name"`
		Namespace string            `yaml:"namespace"`
		Labels    map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
	Spec struct {
		Selector struct {
			MatchLabels map[string]string `yaml:"matchLabels"`
		} `yaml:"selector"`
		Endpoints []Endpoint `yaml:"endpoints"`
		NamespaceSelector struct {
			MatchNames []string `yaml:"matchNames,omitempty"`
			Any        bool     `yaml:"any,omitempty"`
		} `yaml:"namespaceSelector"`
	} `yaml:"spec"`
}

// Endpoint represents a ServiceMonitor endpoint configuration.
type Endpoint struct {
	Port          string `yaml:"port"`
	Path          string `yaml:"path,omitempty"`
	Interval      string `yaml:"interval,omitempty"`
	ScrapeTimeout string `yaml:"scrapeTimeout,omitempty"`
	Scheme        string `yaml:"scheme,omitempty"`
}

// PodMonitor represents the structure of a PodMonitor YAML file.
type PodMonitor struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string            `yaml:"name"`
		Namespace string            `yaml:"namespace"`
		Labels    map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
	Spec struct {
		Selector struct {
			MatchLabels      map[string]string  `yaml:"matchLabels,omitempty"`
			MatchExpressions []MatchExpression  `yaml:"matchExpressions,omitempty"`
		} `yaml:"selector"`
		PodMetricsEndpoints []PodEndpoint `yaml:"podMetricsEndpoints"`
		NamespaceSelector   struct {
			MatchNames []string `yaml:"matchNames,omitempty"`
			Any        bool     `yaml:"any,omitempty"`
		} `yaml:"namespaceSelector"`
	} `yaml:"spec"`
}

// MatchExpression represents a label selector expression.
type MatchExpression struct {
	Key      string   `yaml:"key"`
	Operator string   `yaml:"operator"`
	Values   []string `yaml:"values,omitempty"`
}

// PodEndpoint represents a PodMonitor endpoint configuration.
type PodEndpoint struct {
	Port          string `yaml:"port"`
	Path          string `yaml:"path,omitempty"`
	Interval      string `yaml:"interval,omitempty"`
	ScrapeTimeout string `yaml:"scrapeTimeout,omitempty"`
	Scheme        string `yaml:"scheme,omitempty"`
}

// GrafanaDashboard represents the basic structure of a Grafana dashboard JSON.
type GrafanaDashboard struct {
	Annotations   interface{}   `json:"annotations,omitempty"`
	Editable      bool          `json:"editable,omitempty"`
	ID            interface{}   `json:"id"`
	Links         []interface{} `json:"links,omitempty"`
	Panels        []Panel       `json:"panels"`
	Refresh       string        `json:"refresh,omitempty"`
	SchemaVersion int           `json:"schemaVersion,omitempty"`
	Style         string        `json:"style,omitempty"`
	Tags          []string      `json:"tags,omitempty"`
	Templating    interface{}   `json:"templating,omitempty"`
	Time          interface{}   `json:"time,omitempty"`
	Timepicker    interface{}   `json:"timepicker,omitempty"`
	Timezone      string        `json:"timezone,omitempty"`
	Title         string        `json:"title"`
	UID           string        `json:"uid,omitempty"`
	Version       int           `json:"version,omitempty"`
}

// Panel represents a Grafana dashboard panel.
type Panel struct {
	ID         int         `json:"id,omitempty"`
	Type       string      `json:"type"`
	Title      string      `json:"title,omitempty"`
	Targets    []Target    `json:"targets,omitempty"`
	Datasource interface{} `json:"datasource,omitempty"`
	GridPos    interface{} `json:"gridPos,omitempty"`
	Panels     []Panel     `json:"panels,omitempty"` // For collapsed rows
}

// Target represents a panel target/query.
type Target struct {
	Datasource   interface{} `json:"datasource,omitempty"`
	Expr         string      `json:"expr,omitempty"`
	RefID        string      `json:"refId,omitempty"`
	LegendFormat string      `json:"legendFormat,omitempty"`
}

// ConfigMap represents a Kubernetes ConfigMap structure.
type ConfigMap struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string            `yaml:"name"`
		Namespace string            `yaml:"namespace"`
		Labels    map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
	Data map[string]string `yaml:"data"`
}

// FindYAMLFiles finds all YAML files in a directory matching the given pattern.
func FindYAMLFiles(dir string, pattern string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, _ := filepath.Match(pattern, info.Name()); matched {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// FindJSONFiles finds all JSON files in a directory matching the given pattern.
func FindJSONFiles(dir string, pattern string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, _ := filepath.Match(pattern, info.Name()); matched {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// LoadYAMLFile loads and unmarshals a YAML file into the provided interface.
func LoadYAMLFile(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal YAML from %s: %w", path, err)
	}
	return nil
}

// LoadJSONFile loads and unmarshals a JSON file into the provided interface.
func LoadJSONFile(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", path, err)
	}
	return nil
}

// ContainsHelmTemplate checks if a string contains Helm template syntax ({{ }}).
func ContainsHelmTemplate(s string) bool {
	// Check for {{ and }} patterns that indicate Helm template syntax
	helmTemplatePattern := regexp.MustCompile("\\{\\{.*?\\}\\}")
	return helmTemplatePattern.MatchString(s)
}

// ContainsHelmTemplateExcludingGrafana checks for Helm templates excluding Grafana template variables.
// Grafana uses {{` `}} for escaping, and {{ $labels.xxx }} for templating, which is valid in dashboard JSON.
// Also excludes comments (lines starting with #).
func ContainsHelmTemplateExcludingGrafana(s string) bool {
	// Remove comment lines first (YAML comments)
	lines := strings.Split(s, "\n")
	var nonCommentLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			nonCommentLines = append(nonCommentLines, line)
		}
	}
	cleaned := strings.Join(nonCommentLines, "\n")
	
	// Remove Grafana escape patterns: {{` ... `}}
	grafanaEscapePattern := regexp.MustCompile("\\{\\{`.*?`\\}\\}")
	cleaned = grafanaEscapePattern.ReplaceAllString(cleaned, "")
	
	// Remove Grafana label/value templating: {{ $labels.xxx }} or {{ $value }}
	grafanaLabelPattern := regexp.MustCompile("\\{\\{\\s*\\$[a-zA-Z_\\.]+\\s*\\}\\}")
	cleaned = grafanaLabelPattern.ReplaceAllString(cleaned, "")
	
	// Helm template patterns to detect:
	// - {{ .Values.xxx }}
	// - {{ .Release.Namespace }}
	// - {{ include "xxx" }}
	// - {{ template "xxx" }}
	helmPatterns := []string{
		"\\.Values\\.",
		"\\.Release\\.",
		"\\.Chart\\.",
		"include\\s+\"",
		"template\\s+\"",
		"define\\s+\"",
		"with\\s+\\.",
		"range\\s+\\.",
	}
	
	for _, pattern := range helmPatterns {
		helmPattern := regexp.MustCompile("\\{\\{.*?" + pattern + ".*?\\}\\}")
		if helmPattern.MatchString(cleaned) {
			return true
		}
	}
	
	return false
}

// ValidatePromQLSyntax performs basic validation on a PromQL expression.
// This is a simplified validator - for full validation, use promtool.
func ValidatePromQLSyntax(expr string) error {
	if strings.TrimSpace(expr) == "" {
		return fmt.Errorf("expression is empty")
	}
	
	// Check for balanced parentheses
	parenCount := 0
	braceCount := 0
	bracketCount := 0
	
	for _, ch := range expr {
		switch ch {
		case '(':
			parenCount++
		case ')':
			parenCount--
		case '{':
			braceCount++
		case '}':
			braceCount--
		case '[':
			bracketCount++
		case ']':
			bracketCount--
		}
		
		if parenCount < 0 || braceCount < 0 || bracketCount < 0 {
			return fmt.Errorf("unbalanced brackets in expression: %s", expr)
		}
	}
	
	if parenCount != 0 {
		return fmt.Errorf("unbalanced parentheses in expression: %s", expr)
	}
	if braceCount != 0 {
		return fmt.Errorf("unbalanced braces in expression: %s", expr)
	}
	if bracketCount != 0 {
		return fmt.Errorf("unbalanced brackets in expression: %s", expr)
	}
	
	return nil
}

// GetProjectRoot returns the root directory of the prometheus-operator tests.
func GetProjectRoot() string {
	// This assumes tests are run from within the tests directory or its subdirectories
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	
	// Navigate to find the prometheus-operator directory
	for {
		if _, err := os.Stat(filepath.Join(wd, "k8s")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	
	return ""
}

// GetK8sDir returns the path to the k8s directory containing Kubernetes resources.
func GetK8sDir() string {
	root := GetProjectRoot()
	if root == "" {
		return ""
	}
	return filepath.Join(root, "k8s")
}

// ValidateDuration validates a Prometheus duration string (e.g., "5m", "1h", "30s").
func ValidateDuration(d string) error {
	if d == "" {
		return nil // Empty duration is valid (uses default)
	}
	
	// Prometheus duration format: [0-9]+[smhdwy]
	durationPattern := regexp.MustCompile("^([0-9]+[smhdwy])+$")
	if !durationPattern.MatchString(d) {
		return fmt.Errorf("invalid duration format: %s", d)
	}
	return nil
}

// ValidateSeverity validates that severity is one of the standard values.
func ValidateSeverity(severity string) error {
	validSeverities := map[string]bool{
		"critical": true,
		"warning":  true,
		"info":     true,
	}
	
	if !validSeverities[severity] {
		return fmt.Errorf("invalid severity: %s (expected: critical, warning, or info)", severity)
	}
	return nil
}
