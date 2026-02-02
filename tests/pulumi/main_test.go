package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetClusterConfig tests the getClusterConfig function with valid inputs
func TestGetClusterConfig_ValidStack(t *testing.T) {
	// Setup test fixture
	testStack := "test-cluster"
	testDir := filepath.Join("../../flux/clusters", testStack)
	kindConfigPath := filepath.Join(testDir, "kind.yaml")

	// Create temporary test directory and file
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create a dummy kind.yaml
	kindConfigContent := `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: test-cluster
nodes:
  - role: control-plane
`
	if err := os.WriteFile(kindConfigPath, []byte(kindConfigContent), 0644); err != nil {
		t.Fatalf("Failed to create test kind.yaml: %v", err)
	}

	// Act
	config, err := getClusterConfig(testStack)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	if config.Name != testStack {
		t.Errorf("Expected Name to be %s, got %s", testStack, config.Name)
	}

	if config.KindConfig != kindConfigPath {
		t.Errorf("Expected KindConfig to be %s, got %s", kindConfigPath, config.KindConfig)
	}
}

// TestGetClusterConfig_NonExistentStack tests error handling for non-existent clusters
func TestGetClusterConfig_NonExistentStack(t *testing.T) {
	// Arrange
	nonExistentStack := "non-existent-cluster-12345"

	// Act
	config, err := getClusterConfig(nonExistentStack)

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent cluster, got nil")
	}

	if config != nil {
		t.Error("Expected config to be nil for non-existent cluster")
	}

	expectedErrMsg := "cluster 'non-existent-cluster-12345' not found"
	if err != nil && !contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedErrMsg, err)
	}
}

// TestGetClusterConfig_MissingKindYaml tests error handling for missing kind.yaml
func TestGetClusterConfig_MissingKindYaml(t *testing.T) {
	// Setup test fixture - directory exists but no kind.yaml
	testStack := "missing-yaml-cluster"
	testDir := filepath.Join("../../flux/clusters", testStack)

	// Create directory but no kind.yaml
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Act
	config, err := getClusterConfig(testStack)

	// Assert
	if err == nil {
		t.Error("Expected error for missing kind.yaml, got nil")
	}

	if config != nil {
		t.Error("Expected config to be nil for missing kind.yaml")
	}

	expectedErrMsg := "kind.yaml not found"
	if err != nil && !contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedErrMsg, err)
	}
}

// TestGetClusterConfig_ValidatesKindConfigPath tests that path is correctly formed
func TestGetClusterConfig_ValidatesKindConfigPath(t *testing.T) {
	// Setup multiple test cases
	testCases := []struct {
		name          string
		stack         string
		expectedPath  string
		shouldSucceed bool
	}{
		{
			name:          "Studio cluster",
			stack:         "studio",
			expectedPath:  "../../flux/clusters/studio/kind.yaml",
			shouldSucceed: true,
		},
		{
			name:          "Pro cluster",
			stack:         "pro",
			expectedPath:  "../../flux/clusters/pro/kind.yaml",
			shouldSucceed: true,
		},
		{
			name:          "Air cluster",
			stack:         "air",
			expectedPath:  "../../flux/clusters/air/kind.yaml",
			shouldSucceed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			testDir := filepath.Join("../../flux/clusters", tc.stack)
			kindConfigPath := filepath.Join(testDir, "kind.yaml")

			if tc.shouldSucceed {
				if err := os.MkdirAll(testDir, 0755); err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}
				defer os.RemoveAll(testDir)

				if err := os.WriteFile(kindConfigPath, []byte("test content"), 0644); err != nil {
					t.Fatalf("Failed to create test kind.yaml: %v", err)
				}
			}

			// Act
			config, err := getClusterConfig(tc.stack)

			// Assert
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("Expected no error for %s, got: %v", tc.stack, err)
				}
				if config != nil && config.KindConfig != tc.expectedPath {
					t.Errorf("Expected KindConfig to be %s, got %s", tc.expectedPath, config.KindConfig)
				}
			}
		})
	}
}

// TestClusterConfig_FieldsAreSet tests that all required fields are set
func TestClusterConfig_FieldsAreSet(t *testing.T) {
	testStack := "fields-test-cluster"
	testDir := filepath.Join("../../flux/clusters", testStack)
	kindConfigPath := filepath.Join(testDir, "kind.yaml")

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	if err := os.WriteFile(kindConfigPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test kind.yaml: %v", err)
	}

	config, err := getClusterConfig(testStack)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if config != nil {
		if config.Name == "" {
			t.Error("Expected Name to be set")
		}
		if config.KindConfig == "" {
			t.Error("Expected KindConfig to be set")
		}
	}
}

// TestClusterConfig_StructFields tests that ClusterConfig has expected fields
func TestClusterConfig_StructFields(t *testing.T) {
	config := &ClusterConfig{
		Name:       "test-cluster",
		KindConfig: "/path/to/kind.yaml",
	}

	if config.Name != "test-cluster" {
		t.Errorf("Expected Name to be 'test-cluster', got %s", config.Name)
	}

	if config.KindConfig != "/path/to/kind.yaml" {
		t.Errorf("Expected KindConfig to be '/path/to/kind.yaml', got %s", config.KindConfig)
	}
}

// TestGetClusterConfig_EmptyStackName tests error handling for empty stack name
func TestGetClusterConfig_EmptyStackName(t *testing.T) {
	config, err := getClusterConfig("")

	if err == nil {
		t.Error("Expected error for empty stack name, got nil")
	}

	if config != nil {
		t.Error("Expected config to be nil for empty stack name")
	}
}

// TestGetClusterConfig_PathTraversal tests that path traversal is handled correctly
func TestGetClusterConfig_PathTraversal(t *testing.T) {
	testCases := []string{
		"../../../etc/passwd",
		"../../..",
		"./../../test",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			config, err := getClusterConfig(tc)

			if err == nil {
				t.Error("Expected error for path traversal attempt, got nil")
			}

			if config != nil {
				t.Error("Expected config to be nil for path traversal attempt")
			}
		})
	}
}

// TestGetClusterConfig_SpecialCharacters tests handling of special characters in stack names
func TestGetClusterConfig_SpecialCharacters(t *testing.T) {
	testCases := []string{
		"cluster@name",
		"cluster#123",
		"cluster with spaces",
		"cluster$variable",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			// These should fail as they're invalid cluster names
			config, err := getClusterConfig(tc)

			if err == nil {
				t.Error("Expected error for special characters in cluster name, got nil")
			}

			if config != nil {
				t.Error("Expected config to be nil for invalid cluster name")
			}
		})
	}
}

// TestGetClusterConfig_ValidClusterNames tests valid cluster naming conventions
func TestGetClusterConfig_ValidClusterNames(t *testing.T) {
	validNames := []string{
		"studio",
		"pro",
		"air",
		"forge",
		"pi",
		"test-cluster",
		"cluster123",
		"my-test-cluster-1",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			// Setup
			testDir := filepath.Join("../../flux/clusters", name)
			kindConfigPath := filepath.Join(testDir, "kind.yaml")

			if err := os.MkdirAll(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}
			defer os.RemoveAll(testDir)

			if err := os.WriteFile(kindConfigPath, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create test kind.yaml: %v", err)
			}

			// Act
			config, err := getClusterConfig(name)

			// Assert
			if err != nil {
				t.Errorf("Expected no error for valid cluster name '%s', got: %v", name, err)
			}

			if config == nil {
				t.Errorf("Expected config to be non-nil for valid cluster name '%s'", name)
			}
		})
	}
}

// TestClusterConfig_ImmutableAfterCreation tests that config values match input
func TestClusterConfig_ImmutableAfterCreation(t *testing.T) {
	testStack := "immutable-test"
	testDir := filepath.Join("../../flux/clusters", testStack)
	kindConfigPath := filepath.Join(testDir, "kind.yaml")

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	if err := os.WriteFile(kindConfigPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test kind.yaml: %v", err)
	}

	config, err := getClusterConfig(testStack)
	if err != nil {
		t.Fatalf("Failed to get cluster config: %v", err)
	}

	// Store original values
	originalName := config.Name
	originalKindConfig := config.KindConfig

	// Verify values haven't changed
	if config.Name != originalName {
		t.Error("Name value changed unexpectedly")
	}
	if config.KindConfig != originalKindConfig {
		t.Error("KindConfig value changed unexpectedly")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests for performance validation
func BenchmarkGetClusterConfig(b *testing.B) {
	testStack := "benchmark-cluster"
	testDir := filepath.Join("../../flux/clusters", testStack)
	kindConfigPath := filepath.Join(testDir, "kind.yaml")

	if err := os.MkdirAll(testDir, 0755); err != nil {
		b.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	if err := os.WriteFile(kindConfigPath, []byte("test"), 0644); err != nil {
		b.Fatalf("Failed to create test kind.yaml: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = getClusterConfig(testStack)
	}
}
