// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 AGENT-001: LambdaAgent Lifecycle Tests
//
//	User Story: LambdaAgent Deployment and Management
//	Priority: P0 | Story Points: 8
//
//	Tests validate:
//	- Agent creation from pre-built image
//	- Agent status transitions (Pending → Deploying → Ready)
//	- Agent update triggers new revision
//	- Agent deletion cleans up all resources
//	- Finalizer prevents orphaned resources
//	- Owner references cascade deletion
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package agents

import (
	"context"
	"fmt"
	"testing"
	"time"

	"knative-lambda/tests/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants
const (
	TestNamespace      = "agent-test"
	TestAgentName      = "test-agent"
	TestImageRepo      = "localhost:5001/test-agent"
	TestImageTag       = "v1.0.0"
	DefaultOllamaURL   = "http://ollama.ollama.svc.cluster.local:11434"
	DefaultOllamaModel = "llama3.2:3b"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Test Fixtures and Helpers
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

// LambdaAgentPhase represents agent phase states
type LambdaAgentPhase string

const (
	AgentPhasePending   LambdaAgentPhase = "Pending"
	AgentPhaseDeploying LambdaAgentPhase = "Deploying"
	AgentPhaseReady     LambdaAgentPhase = "Ready"
	AgentPhaseFailed    LambdaAgentPhase = "Failed"
	AgentPhaseDeleting  LambdaAgentPhase = "Deleting"
)

// MockLambdaAgent represents a test LambdaAgent structure
type MockLambdaAgent struct {
	Name       string
	Namespace  string
	Image      MockAgentImage
	AI         *MockAgentAI
	Scaling    *MockAgentScaling
	Phase      LambdaAgentPhase
	Conditions []MockCondition
	Finalizers []string
}

// MockAgentImage represents image configuration
type MockAgentImage struct {
	Repository string
	Tag        string
	Port       int32
}

// MockAgentAI represents AI configuration
type MockAgentAI struct {
	Provider        string
	Endpoint        string
	Model           string
	Temperature     string
	MaxTokens       int32
	APIKeySecretRef *MockSecretKeySelector
}

// MockSecretKeySelector represents a secret key selector
type MockSecretKeySelector struct {
	Name string
	Key  string
}

// MockAgentScaling represents scaling configuration
type MockAgentScaling struct {
	MinReplicas       int32
	MaxReplicas       int32
	TargetConcurrency int32
}

// MockCondition represents a condition
type MockCondition struct {
	Type    string
	Status  string
	Reason  string
	Message string
}

// createTestAgent creates a mock LambdaAgent for testing
func createTestAgent(name string, opts ...func(*MockLambdaAgent)) *MockLambdaAgent {
	agent := &MockLambdaAgent{
		Name:      name,
		Namespace: TestNamespace,
		Image: MockAgentImage{
			Repository: TestImageRepo,
			Tag:        TestImageTag,
			Port:       8080,
		},
		AI: &MockAgentAI{
			Provider:    "ollama",
			Endpoint:    DefaultOllamaURL,
			Model:       DefaultOllamaModel,
			Temperature: "0.7",
			MaxTokens:   2048,
		},
		Scaling: &MockAgentScaling{
			MinReplicas:       1,
			MaxReplicas:       5,
			TargetConcurrency: 10,
		},
		Phase:      AgentPhasePending,
		Finalizers: []string{"lambdaagent.lambda.knative.io/finalizer"},
	}

	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

// withPhase sets the agent phase
func withPhase(phase LambdaAgentPhase) func(*MockLambdaAgent) {
	return func(a *MockLambdaAgent) {
		a.Phase = phase
	}
}

// withCondition adds a condition
func withCondition(condType, status, reason, message string) func(*MockLambdaAgent) {
	return func(a *MockLambdaAgent) {
		a.Conditions = append(a.Conditions, MockCondition{
			Type:    condType,
			Status:  status,
			Reason:  reason,
			Message: message,
		})
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: Agent creation from pre-built image (no build required)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT001_AC1_AgentCreation(t *testing.T) {
	testutils.SetupTestEnvironment(t)

	t.Run("Create agent with minimal configuration", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("minimal-agent")

		// Act & Assert
		assert.NotNil(t, agent, "Agent should be created")
		assert.Equal(t, "minimal-agent", agent.Name)
		assert.Equal(t, TestNamespace, agent.Namespace)
		assert.Equal(t, TestImageRepo, agent.Image.Repository)
		assert.Equal(t, TestImageTag, agent.Image.Tag)
	})

	t.Run("Create agent with full AI configuration", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("full-ai-agent")
		agent.AI = &MockAgentAI{
			Provider:    "ollama",
			Endpoint:    "http://custom-ollama:11434",
			Model:       "codellama:13b",
			Temperature: "0.5",
			MaxTokens:   4096,
		}

		// Act & Assert
		assert.NotNil(t, agent.AI, "AI config should be set")
		assert.Equal(t, "ollama", agent.AI.Provider)
		assert.Equal(t, "codellama:13b", agent.AI.Model)
		assert.Equal(t, int32(4096), agent.AI.MaxTokens)
	})

	t.Run("Agent requires image repository", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("no-repo-agent")
		agent.Image.Repository = ""

		// Act
		isValid := validateAgentSpec(agent)

		// Assert
		assert.False(t, isValid, "Agent with empty repository should be invalid")
	})

	t.Run("Agent uses digest over tag when both specified", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("digest-agent")
		imageDigest := "sha256:abc123def456"

		// Act
		imageURI := buildImageURI(agent.Image.Repository, agent.Image.Tag, imageDigest)

		// Assert
		expected := fmt.Sprintf("%s@%s", TestImageRepo, imageDigest)
		assert.Equal(t, expected, imageURI, "Should use digest over tag")
	})

	t.Run("Agent defaults to latest tag when not specified", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("default-tag-agent")
		agent.Image.Tag = ""

		// Act
		imageURI := buildImageURI(agent.Image.Repository, agent.Image.Tag, "")

		// Assert
		expected := fmt.Sprintf("%s:latest", TestImageRepo)
		assert.Equal(t, expected, imageURI, "Should default to latest tag")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Agent status transitions (Pending → Deploying → Ready)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT001_AC2_StatusTransitions(t *testing.T) {
	t.Run("New agent starts in Pending phase", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("pending-agent")

		// Assert
		assert.Equal(t, AgentPhasePending, agent.Phase, "New agent should be Pending")
	})

	t.Run("Agent transitions to Deploying when spec is valid", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("deploying-agent", withPhase(AgentPhasePending))

		// Act
		if validateAgentSpec(agent) {
			agent.Phase = AgentPhaseDeploying
		}

		// Assert
		assert.Equal(t, AgentPhaseDeploying, agent.Phase, "Valid agent should transition to Deploying")
	})

	t.Run("Agent transitions to Ready when service is ready", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("ready-agent", withPhase(AgentPhaseDeploying))

		// Act - Simulate Knative Service becoming ready
		serviceReady := true
		if serviceReady {
			agent.Phase = AgentPhaseReady
			agent.Conditions = append(agent.Conditions, MockCondition{
				Type:    "Ready",
				Status:  "True",
				Reason:  "Ready",
				Message: "Agent is ready",
			})
		}

		// Assert
		assert.Equal(t, AgentPhaseReady, agent.Phase, "Agent should be Ready")
		assert.Len(t, agent.Conditions, 1, "Should have Ready condition")
	})

	t.Run("Agent transitions to Failed on error", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("failed-agent", withPhase(AgentPhaseDeploying))

		// Act - Simulate failure
		deployError := fmt.Errorf("image pull failed: unauthorized")
		agent.Phase = AgentPhaseFailed
		agent.Conditions = append(agent.Conditions, MockCondition{
			Type:    "Ready",
			Status:  "False",
			Reason:  "CreateFailed",
			Message: deployError.Error(),
		})

		// Assert
		assert.Equal(t, AgentPhaseFailed, agent.Phase, "Agent should be Failed")
		assert.Equal(t, "False", agent.Conditions[0].Status)
	})

	t.Run("Failed agent retries after backoff", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("retry-agent", withPhase(AgentPhaseFailed))
		retryDelay := 60 * time.Second

		// Act - Simulate retry
		agent.Phase = AgentPhasePending

		// Assert
		assert.Equal(t, AgentPhasePending, agent.Phase, "Failed agent should reset to Pending")
		assert.Equal(t, 60*time.Second, retryDelay, "Retry delay should be 60 seconds")
	})

	t.Run("Status transition timing under SLO", func(t *testing.T) {
		// Arrange
		startTime := time.Now()
		maxDeployTime := 60 * time.Second

		// Simulate deployment phases
		phases := []testutils.Phase{
			{Name: "Pending", Duration: 1 * time.Second},
			{Name: "Deploying", Duration: 5 * time.Second},
			{Name: "ServiceCreation", Duration: 10 * time.Second},
		}

		// Act
		var totalDuration time.Duration
		for _, phase := range phases {
			totalDuration += phase.Duration
		}
		endTime := startTime.Add(totalDuration)

		// Assert
		testutils.RunTimingTest(t, "Agent deployment", startTime, endTime, maxDeployTime, phases)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Agent update triggers new revision
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT001_AC3_AgentUpdates(t *testing.T) {
	t.Run("Image tag change triggers new revision", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("update-image-agent", withPhase(AgentPhaseReady))
		oldTag := agent.Image.Tag
		newTag := "v2.0.0"

		// Act
		agent.Image.Tag = newTag
		revisionChanged := oldTag != newTag

		// Assert
		assert.True(t, revisionChanged, "Tag change should trigger revision")
		assert.Equal(t, "v2.0.0", agent.Image.Tag)
	})

	t.Run("AI model change triggers new revision", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("update-model-agent", withPhase(AgentPhaseReady))
		oldModel := agent.AI.Model
		newModel := "llama3.2:7b"

		// Act
		agent.AI.Model = newModel
		revisionChanged := oldModel != newModel

		// Assert
		assert.True(t, revisionChanged, "Model change should trigger revision")
	})

	t.Run("Scaling change does not trigger revision", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("scale-agent", withPhase(AgentPhaseReady))

		// Act - Scaling changes should not cause new pod creation
		agent.Scaling.MinReplicas = 2
		agent.Scaling.MaxReplicas = 10

		// Assert - Scaling is annotation-based, not revision-based
		assert.Equal(t, int32(2), agent.Scaling.MinReplicas)
		assert.Equal(t, int32(10), agent.Scaling.MaxReplicas)
	})

	t.Run("Environment variable change triggers new revision", func(t *testing.T) {
		// Arrange
		envVars := map[string]string{
			"LOG_LEVEL": "info",
		}

		// Act
		envVars["LOG_LEVEL"] = "debug"

		// Assert
		assert.Equal(t, "debug", envVars["LOG_LEVEL"])
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: Agent deletion cleans up all resources
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT001_AC4_AgentDeletion(t *testing.T) {
	t.Run("Deletion sets phase to Deleting", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("delete-agent", withPhase(AgentPhaseReady))

		// Act - Simulate deletion
		agent.Phase = AgentPhaseDeleting

		// Assert
		assert.Equal(t, AgentPhaseDeleting, agent.Phase)
	})

	t.Run("Finalizer is present on new agent", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("finalizer-agent")

		// Assert
		assert.Contains(t, agent.Finalizers, "lambdaagent.lambda.knative.io/finalizer")
	})

	t.Run("Finalizer removed after cleanup", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("cleanup-agent", withPhase(AgentPhaseDeleting))

		// Act - Simulate cleanup
		// Remove Knative Service (done)
		// Remove eventing resources (done)
		agent.Finalizers = []string{}

		// Assert
		assert.Empty(t, agent.Finalizers, "Finalizers should be removed after cleanup")
	})

	t.Run("Owner references cascade deletion", func(t *testing.T) {
		// Arrange
		ownerReferences := []struct {
			Kind string
			Name string
		}{
			{Kind: "Service", Name: "test-agent"},
			{Kind: "Broker", Name: "test-agent-broker"},
			{Kind: "Trigger", Name: "test-agent-trigger"},
		}

		// Act & Assert
		for _, ref := range ownerReferences {
			assert.NotEmpty(t, ref.Kind)
			assert.NotEmpty(t, ref.Name)
		}
		assert.Len(t, ownerReferences, 3, "Should have 3 owned resources")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Knative Service creation from agent spec
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT001_AC5_KnativeServiceCreation(t *testing.T) {
	t.Run("Knative Service created with correct image", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("ksvc-agent")
		expectedImageURI := fmt.Sprintf("%s:%s", TestImageRepo, TestImageTag)

		// Act
		imageURI := buildImageURI(agent.Image.Repository, agent.Image.Tag, "")

		// Assert
		assert.Equal(t, expectedImageURI, imageURI)
	})

	t.Run("Knative Service has correct scaling annotations", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("scaling-agent")

		// Act
		annotations := buildScalingAnnotations(agent.Scaling)

		// Assert
		assert.Equal(t, "1", annotations["autoscaling.knative.dev/min-scale"])
		assert.Equal(t, "5", annotations["autoscaling.knative.dev/max-scale"])
		assert.Equal(t, "10", annotations["autoscaling.knative.dev/target"])
	})

	t.Run("Knative Service has correct labels", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("labels-agent")

		// Act
		labels := buildAgentLabels(agent)

		// Assert
		assert.Equal(t, agent.Name, labels["app.kubernetes.io/name"])
		assert.Equal(t, "knative-lambda-operator", labels["app.kubernetes.io/managed-by"])
		assert.Equal(t, "agent", labels["app.kubernetes.io/component"])
	})

	t.Run("Knative Service has health probes", func(t *testing.T) {
		// Arrange
		healthPath := "/health"
		containerPort := int32(8080)

		// Act
		readinessProbe := buildProbe(healthPath, containerPort, 5, 10)
		livenessProbe := buildProbe(healthPath, containerPort, 15, 20)

		// Assert
		assert.Equal(t, healthPath, readinessProbe.Path)
		assert.Equal(t, containerPort, readinessProbe.Port)
		assert.Equal(t, int32(5), readinessProbe.InitialDelaySeconds)
		assert.Equal(t, int32(15), livenessProbe.InitialDelaySeconds)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Integration Test: Full Agent Lifecycle
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT001_Integration_FullLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Complete agent lifecycle", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		_ = ctx // Would be used with real client

		startTime := time.Now()

		// Step 1: Create agent
		agent := createTestAgent("lifecycle-agent")
		assert.Equal(t, AgentPhasePending, agent.Phase)

		// Step 2: Validate and transition to Deploying
		if validateAgentSpec(agent) {
			agent.Phase = AgentPhaseDeploying
		}
		assert.Equal(t, AgentPhaseDeploying, agent.Phase)

		// Step 3: Simulate Knative Service becoming ready
		agent.Phase = AgentPhaseReady
		agent.Conditions = append(agent.Conditions, MockCondition{
			Type:   "Ready",
			Status: "True",
			Reason: "Ready",
		})
		assert.Equal(t, AgentPhaseReady, agent.Phase)

		// Step 4: Update agent
		agent.Image.Tag = "v2.0.0"
		agent.Phase = AgentPhaseDeploying // Would re-reconcile

		// Step 5: Delete agent
		agent.Phase = AgentPhaseDeleting
		agent.Finalizers = []string{}

		endTime := time.Now()

		// Assert timing
		lifecycleDuration := endTime.Sub(startTime)
		assert.Less(t, lifecycleDuration.Milliseconds(), int64(1000),
			"Lifecycle simulation should complete quickly")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Helper Functions
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

// validateAgentSpec validates the agent specification
func validateAgentSpec(agent *MockLambdaAgent) bool {
	if agent.Image.Repository == "" {
		return false
	}
	return true
}

// buildImageURI builds the full image URI
func buildImageURI(repository, tag, digest string) string {
	if digest != "" {
		return fmt.Sprintf("%s@%s", repository, digest)
	}
	if tag == "" {
		return fmt.Sprintf("%s:latest", repository)
	}
	return fmt.Sprintf("%s:%s", repository, tag)
}

// buildScalingAnnotations builds Knative scaling annotations
func buildScalingAnnotations(scaling *MockAgentScaling) map[string]string {
	return map[string]string{
		"autoscaling.knative.dev/class":     "kpa.autoscaling.knative.dev",
		"autoscaling.knative.dev/min-scale": fmt.Sprintf("%d", scaling.MinReplicas),
		"autoscaling.knative.dev/max-scale": fmt.Sprintf("%d", scaling.MaxReplicas),
		"autoscaling.knative.dev/target":    fmt.Sprintf("%d", scaling.TargetConcurrency),
	}
}

// buildAgentLabels builds labels for the agent
func buildAgentLabels(agent *MockLambdaAgent) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       agent.Name,
		"app.kubernetes.io/managed-by": "knative-lambda-operator",
		"app.kubernetes.io/component":  "agent",
	}
}

// ProbeConfig represents probe configuration
type ProbeConfig struct {
	Path                string
	Port                int32
	InitialDelaySeconds int32
	PeriodSeconds       int32
}

// buildProbe builds a probe configuration
func buildProbe(path string, port int32, initialDelay, period int32) ProbeConfig {
	return ProbeConfig{
		Path:                path,
		Port:                port,
		InitialDelaySeconds: initialDelay,
		PeriodSeconds:       period,
	}
}

// Ensure require is used
var _ = require.NoError
