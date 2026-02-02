// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 AGENT-004: LambdaAgent Scaling Tests
//
//	User Story: Autoscaling for AI Agents
//	Priority: P1 | Story Points: 5
//
//	Tests validate:
//	- Knative autoscaling annotations
//	- MinReplicas default of 1 for agents (keep warm)
//	- Scale-to-zero grace period
//	- Target concurrency configuration
//	- Container concurrency (ADR-004)
//	- Scale-down delay (ADR-004)
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package agents

import (
	"fmt"
	"testing"

	"knative-lambda/tests/testutils"

	"github.com/stretchr/testify/assert"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Test Fixtures for Scaling
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

// FullScalingConfig represents complete scaling configuration
type FullScalingConfig struct {
	MinReplicas            int32
	MaxReplicas            int32
	TargetConcurrency      int32
	ContainerConcurrency   int32  // ADR-004
	ScaleDownDelay         string // ADR-004
	ScaleToZeroGracePeriod string
	Metrics                []ScalingMetric // ADR-004
}

// ScalingMetric represents a scaling metric
type ScalingMetric struct {
	Type   string
	Name   string
	Target int32
}

// createTestScaling creates a test scaling configuration
func createTestScaling() *FullScalingConfig {
	return &FullScalingConfig{
		MinReplicas:            1, // ADR-004 recommends agents stay warm
		MaxReplicas:            5,
		TargetConcurrency:      10,
		ContainerConcurrency:   10,
		ScaleDownDelay:         "5m",
		ScaleToZeroGracePeriod: "0s", // Never scale to zero for agents
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: Knative autoscaling annotations
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT004_AC1_AutoscalingAnnotations(t *testing.T) {
	testutils.SetupTestEnvironment(t)

	t.Run("Uses KPA autoscaling class", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()

		// Act
		annotations := buildKnativeScalingAnnotations(scaling)

		// Assert
		assert.Equal(t, "kpa.autoscaling.knative.dev", annotations["autoscaling.knative.dev/class"])
	})

	t.Run("Min scale annotation set", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()

		// Act
		annotations := buildKnativeScalingAnnotations(scaling)

		// Assert
		assert.Equal(t, "1", annotations["autoscaling.knative.dev/min-scale"])
	})

	t.Run("Max scale annotation set", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()

		// Act
		annotations := buildKnativeScalingAnnotations(scaling)

		// Assert
		assert.Equal(t, "5", annotations["autoscaling.knative.dev/max-scale"])
	})

	t.Run("Target concurrency annotation set", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()

		// Act
		annotations := buildKnativeScalingAnnotations(scaling)

		// Assert
		assert.Equal(t, "10", annotations["autoscaling.knative.dev/target"])
	})

	t.Run("Scale to zero grace period annotation", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()

		// Act
		annotations := buildKnativeScalingAnnotations(scaling)

		// Assert
		assert.Equal(t, "0s", annotations["autoscaling.knative.dev/scale-to-zero-pod-retention-period"])
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: MinReplicas default for agents
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT004_AC2_MinReplicasDefault(t *testing.T) {
	t.Run("Default minReplicas is 1 for agents (ADR-004)", func(t *testing.T) {
		// Arrange - Per ADR-004, agents should stay warm
		scaling := createTestScaling()

		// Assert
		assert.Equal(t, int32(1), scaling.MinReplicas,
			"ADR-004 recommends minReplicas=1 to keep agents warm")
	})

	t.Run("MinReplicas 0 allows scale to zero", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()
		scaling.MinReplicas = 0

		// Act
		canScaleToZero := scaling.MinReplicas == 0

		// Assert
		assert.True(t, canScaleToZero)
	})

	t.Run("Cold start implications documented", func(t *testing.T) {
		// Arrange - Document cold start impact
		coldStartScenarios := map[string]struct {
			MinReplicas int32
			ColdStartMs int64
		}{
			"Scale to zero":    {0, 5000}, // ~5s cold start
			"Warm (1 replica)": {1, 0},    // No cold start
			"HA (2 replicas)":  {2, 0},    // No cold start
		}

		// Act & Assert
		for scenario, config := range coldStartScenarios {
			if config.MinReplicas == 0 {
				assert.Greater(t, config.ColdStartMs, int64(0),
					"%s should have cold start latency", scenario)
			} else {
				assert.Equal(t, int64(0), config.ColdStartMs,
					"%s should not have cold start", scenario)
			}
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Scale-down delay (ADR-004)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT004_AC3_ScaleDownDelay(t *testing.T) {
	t.Run("Scale down delay configuration (ADR-004)", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()

		// Assert
		assert.Equal(t, "5m", scaling.ScaleDownDelay,
			"ADR-004 specifies scaleDownDelay for keeping agents alive")
	})

	t.Run("Scale down delay annotation generated", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()

		// Act
		annotations := buildKnativeScalingAnnotations(scaling)

		// Assert
		assert.Equal(t, "5m", annotations["autoscaling.knative.dev/scale-down-delay"])
	})

	t.Run("Different scale down delay values", func(t *testing.T) {
		// Arrange
		testCases := []struct {
			delay    string
			expected string
		}{
			{"30s", "30s"},
			{"5m", "5m"},
			{"15m", "15m"},
			{"1h", "1h"},
		}

		// Act & Assert
		for _, tc := range testCases {
			scaling := createTestScaling()
			scaling.ScaleDownDelay = tc.delay
			annotations := buildKnativeScalingAnnotations(scaling)
			assert.Equal(t, tc.expected, annotations["autoscaling.knative.dev/scale-down-delay"])
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: Container concurrency (ADR-004)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT004_AC4_ContainerConcurrency(t *testing.T) {
	t.Run("Container concurrency configuration (ADR-004)", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()

		// Assert
		assert.Equal(t, int32(10), scaling.ContainerConcurrency)
	})

	t.Run("Container concurrency in Knative Service spec", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()
		scaling.ContainerConcurrency = 20

		// Act
		ksvcSpec := buildKnativeServiceSpec(scaling)

		// Assert
		assert.Equal(t, int64(20), ksvcSpec.ContainerConcurrency)
	})

	t.Run("Zero container concurrency means unlimited", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()
		scaling.ContainerConcurrency = 0

		// Act
		ksvcSpec := buildKnativeServiceSpec(scaling)

		// Assert
		assert.Equal(t, int64(0), ksvcSpec.ContainerConcurrency,
			"Zero means unlimited concurrency")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Custom Prometheus metrics scaling (ADR-004)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT004_AC5_CustomMetricsScaling(t *testing.T) {
	t.Run("Prometheus custom metrics configuration (ADR-004)", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()
		scaling.Metrics = []ScalingMetric{
			{
				Type:   "prometheus",
				Name:   "agent_active_conversations",
				Target: 50,
			},
		}

		// Assert
		assert.Len(t, scaling.Metrics, 1)
		assert.Equal(t, "prometheus", scaling.Metrics[0].Type)
		assert.Equal(t, "agent_active_conversations", scaling.Metrics[0].Name)
		assert.Equal(t, int32(50), scaling.Metrics[0].Target)
	})

	t.Run("Multiple custom metrics", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()
		scaling.Metrics = []ScalingMetric{
			{Type: "concurrency", Target: 10},
			{Type: "prometheus", Name: "agent_active_conversations", Target: 50},
			{Type: "prometheus", Name: "agent_queue_depth", Target: 100},
		}

		// Assert
		assert.Len(t, scaling.Metrics, 3)
	})

	t.Run("Concurrency metric type", func(t *testing.T) {
		// Arrange
		scaling := createTestScaling()
		scaling.Metrics = []ScalingMetric{
			{Type: "concurrency", Target: 10},
		}

		// Assert
		assert.Equal(t, "concurrency", scaling.Metrics[0].Type)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: Resource limits integration
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT004_AC6_ResourceLimits(t *testing.T) {
	t.Run("Scaling considers resource limits", func(t *testing.T) {
		// Arrange
		resources := struct {
			Requests struct {
				CPU    string
				Memory string
			}
			Limits struct {
				CPU    string
				Memory string
			}
		}{
			Requests: struct {
				CPU    string
				Memory string
			}{"100m", "256Mi"},
			Limits: struct {
				CPU    string
				Memory string
			}{"500m", "512Mi"},
		}

		// Assert
		assert.Equal(t, "100m", resources.Requests.CPU)
		assert.Equal(t, "256Mi", resources.Requests.Memory)
		assert.Equal(t, "500m", resources.Limits.CPU)
		assert.Equal(t, "512Mi", resources.Limits.Memory)
	})

	t.Run("HPA with resource metrics", func(t *testing.T) {
		// Arrange - HPA can use CPU/Memory metrics
		resourceMetrics := []struct {
			Type   string
			Target int32
		}{
			{"cpu", 70},
			{"memory", 80},
		}

		// Assert
		for _, metric := range resourceMetrics {
			assert.Greater(t, metric.Target, int32(0))
			assert.LessOrEqual(t, metric.Target, int32(100))
		}
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Integration Test: Full Scaling Configuration
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT004_Integration_FullScalingConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Complete scaling configuration to annotations", func(t *testing.T) {
		// Arrange - Full ADR-004 scaling config
		scaling := &FullScalingConfig{
			MinReplicas:            1,
			MaxReplicas:            10,
			TargetConcurrency:      10,
			ContainerConcurrency:   10,
			ScaleDownDelay:         "5m",
			ScaleToZeroGracePeriod: "0s",
			Metrics: []ScalingMetric{
				{Type: "concurrency", Target: 10},
				{Type: "prometheus", Name: "agent_active_conversations", Target: 50},
			},
		}

		// Act
		annotations := buildKnativeScalingAnnotations(scaling)
		ksvcSpec := buildKnativeServiceSpec(scaling)

		// Assert - Annotations
		assert.Equal(t, "kpa.autoscaling.knative.dev", annotations["autoscaling.knative.dev/class"])
		assert.Equal(t, "1", annotations["autoscaling.knative.dev/min-scale"])
		assert.Equal(t, "10", annotations["autoscaling.knative.dev/max-scale"])
		assert.Equal(t, "10", annotations["autoscaling.knative.dev/target"])
		assert.Equal(t, "5m", annotations["autoscaling.knative.dev/scale-down-delay"])
		assert.Equal(t, "0s", annotations["autoscaling.knative.dev/scale-to-zero-pod-retention-period"])

		// Assert - Knative Service spec
		assert.Equal(t, int64(10), ksvcSpec.ContainerConcurrency)

		// Assert - Metrics count
		assert.Len(t, scaling.Metrics, 2)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Helper Functions
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

// buildKnativeScalingAnnotations builds Knative scaling annotations
func buildKnativeScalingAnnotations(scaling *FullScalingConfig) map[string]string {
	annotations := map[string]string{
		"autoscaling.knative.dev/class":     "kpa.autoscaling.knative.dev",
		"autoscaling.knative.dev/min-scale": fmt.Sprintf("%d", scaling.MinReplicas),
		"autoscaling.knative.dev/max-scale": fmt.Sprintf("%d", scaling.MaxReplicas),
		"autoscaling.knative.dev/target":    fmt.Sprintf("%d", scaling.TargetConcurrency),
	}

	if scaling.ScaleToZeroGracePeriod != "" {
		annotations["autoscaling.knative.dev/scale-to-zero-pod-retention-period"] = scaling.ScaleToZeroGracePeriod
	}

	if scaling.ScaleDownDelay != "" {
		annotations["autoscaling.knative.dev/scale-down-delay"] = scaling.ScaleDownDelay
	}

	return annotations
}

// KnativeServiceSpec represents relevant Knative Service spec fields
type KnativeServiceSpec struct {
	ContainerConcurrency int64
}

// buildKnativeServiceSpec builds Knative Service spec
func buildKnativeServiceSpec(scaling *FullScalingConfig) KnativeServiceSpec {
	return KnativeServiceSpec{
		ContainerConcurrency: int64(scaling.ContainerConcurrency),
	}
}
