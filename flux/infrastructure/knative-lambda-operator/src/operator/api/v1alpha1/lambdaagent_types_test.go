// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 Unit Tests: LambdaAgent Types
//
//	Tests for LambdaAgent API type helpers:
//	- SetCondition / GetCondition
//	- Status management
//	- Phase constants
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📋 AGENT STATUS CONDITION TESTS                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestLambdaAgentStatus_SetCondition(t *testing.T) {
	tests := []struct {
		name              string
		initialConditions []metav1.Condition
		newCondition      metav1.Condition
		expectedCount     int
		description       string
	}{
		{
			name:              "Add first condition to agent",
			initialConditions: nil,
			newCondition: metav1.Condition{
				Type:    "Ready",
				Status:  metav1.ConditionFalse,
				Reason:  "Initializing",
				Message: "Agent is initializing",
			},
			expectedCount: 1,
			description:   "Should add condition to empty list",
		},
		{
			name: "Update Ready condition",
			initialConditions: []metav1.Condition{
				{
					Type:    "Ready",
					Status:  metav1.ConditionFalse,
					Reason:  "Deploying",
					Message: "Service is deploying",
				},
			},
			newCondition: metav1.Condition{
				Type:    "Ready",
				Status:  metav1.ConditionTrue,
				Reason:  "Ready",
				Message: "Agent is ready",
			},
			expectedCount: 1,
			description:   "Should update existing condition",
		},
		{
			name: "Add Eventing condition",
			initialConditions: []metav1.Condition{
				{Type: "Ready", Status: metav1.ConditionTrue, Reason: "Ready", Message: "Ready"},
			},
			newCondition: metav1.Condition{
				Type:    "Eventing",
				Status:  metav1.ConditionTrue,
				Reason:  "EventingReady",
				Message: "Eventing infrastructure ready",
			},
			expectedCount: 2,
			description:   "Should add different condition type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			status := &LambdaAgentStatus{
				Conditions: tt.initialConditions,
			}

			// Act
			status.SetCondition(tt.newCondition)

			// Assert
			assert.Len(t, status.Conditions, tt.expectedCount, tt.description)

			found := status.GetCondition(tt.newCondition.Type)
			require.NotNil(t, found, "Condition should exist after SetCondition")
			assert.Equal(t, tt.newCondition.Status, found.Status)
			assert.Equal(t, tt.newCondition.Reason, found.Reason)
			assert.Equal(t, tt.newCondition.Message, found.Message)
		})
	}
}

func TestLambdaAgentStatus_GetCondition(t *testing.T) {
	tests := []struct {
		name           string
		conditions     []metav1.Condition
		conditionType  string
		expectNil      bool
		expectedStatus metav1.ConditionStatus
		description    string
	}{
		{
			name:          "Get from empty conditions",
			conditions:    nil,
			conditionType: "Ready",
			expectNil:     true,
			description:   "Should return nil for empty list",
		},
		{
			name: "Get existing Ready condition",
			conditions: []metav1.Condition{
				{Type: "Ready", Status: metav1.ConditionTrue, Reason: "Ready"},
			},
			conditionType:  "Ready",
			expectNil:      false,
			expectedStatus: metav1.ConditionTrue,
			description:    "Should return existing condition",
		},
		{
			name: "Get non-existing condition",
			conditions: []metav1.Condition{
				{Type: "Ready", Status: metav1.ConditionTrue, Reason: "Ready"},
			},
			conditionType: "Eventing",
			expectNil:     true,
			description:   "Should return nil for missing type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := &LambdaAgentStatus{
				Conditions: tt.conditions,
			}

			result := status.GetCondition(tt.conditionType)

			if tt.expectNil {
				assert.Nil(t, result, tt.description)
			} else {
				require.NotNil(t, result, tt.description)
				assert.Equal(t, tt.expectedStatus, result.Status)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 AGENT PHASE TESTS                                                   │
// └─────────────────────────────────────────────────────────────────────────┘

func TestLambdaAgentPhase_Constants(t *testing.T) {
	phases := map[LambdaAgentPhase]string{
		AgentPhasePending:   "Pending",
		AgentPhaseDeploying: "Deploying",
		AgentPhaseReady:     "Ready",
		AgentPhaseFailed:    "Failed",
		AgentPhaseDeleting:  "Deleting",
	}

	for phase, expected := range phases {
		t.Run(string(phase), func(t *testing.T) {
			assert.Equal(t, expected, string(phase), "Phase constant should match")
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🖼️ AGENT IMAGE SPEC TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAgentImageSpec_URIGeneration(t *testing.T) {
	tests := []struct {
		name        string
		imageSpec   AgentImageSpec
		expectedURI string
		description string
	}{
		{
			name: "Repository with tag",
			imageSpec: AgentImageSpec{
				Repository: "gcr.io/project/agent",
				Tag:        "v1.0.0",
			},
			expectedURI: "gcr.io/project/agent:v1.0.0",
			description: "Should combine repository and tag",
		},
		{
			name: "Repository with digest",
			imageSpec: AgentImageSpec{
				Repository: "gcr.io/project/agent",
				Digest:     "sha256:abc123",
			},
			expectedURI: "gcr.io/project/agent@sha256:abc123",
			description: "Digest should take precedence",
		},
		{
			name: "Repository only (default to latest)",
			imageSpec: AgentImageSpec{
				Repository: "gcr.io/project/agent",
			},
			expectedURI: "gcr.io/project/agent:latest",
			description: "Should default to latest tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build URI based on spec (mimicking controller logic)
			imageURI := tt.imageSpec.Repository
			if tt.imageSpec.Digest != "" {
				imageURI = imageURI + "@" + tt.imageSpec.Digest
			} else if tt.imageSpec.Tag != "" {
				imageURI = imageURI + ":" + tt.imageSpec.Tag
			} else {
				imageURI = imageURI + ":latest"
			}

			assert.Equal(t, tt.expectedURI, imageURI, tt.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🤖 AGENT AI CONFIG TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAgentAISpec_Configuration(t *testing.T) {
	aiSpec := &AgentAISpec{
		Provider:    "ollama",
		Endpoint:    "http://ollama.ai.svc:11434",
		Model:       "llama2",
		MaxTokens:   4096,
		Temperature: "0.7",
	}

	assert.Equal(t, "ollama", aiSpec.Provider)
	assert.Equal(t, "http://ollama.ai.svc:11434", aiSpec.Endpoint)
	assert.Equal(t, "llama2", aiSpec.Model)
	assert.Equal(t, int32(4096), aiSpec.MaxTokens)
	assert.Equal(t, "0.7", aiSpec.Temperature)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📡 AGENT EVENTING TESTS                                                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAgentEventingSpec_Subscriptions(t *testing.T) {
	eventing := &AgentEventingSpec{
		Enabled:     true,
		EventSource: "io.example.agent",
		Intents:     []string{"io.example.intent.query", "io.example.intent.action"},
		Subscriptions: []AgentSubscription{
			{
				EventType:   "io.example.event.created",
				Source:      "io.example.service",
				Description: "Receive created events",
			},
			{
				EventType:   "io.example.event.updated",
				Description: "Receive updated events",
			},
		},
		Forwards: []AgentForward{
			{
				EventTypes:      []string{"io.example.intent.query"},
				TargetAgent:     "query-processor",
				TargetNamespace: "processing",
				Description:     "Forward query intents",
			},
		},
	}

	assert.True(t, eventing.Enabled)
	assert.Equal(t, "io.example.agent", eventing.EventSource)
	assert.Len(t, eventing.Intents, 2)
	assert.Len(t, eventing.Subscriptions, 2)
	assert.Len(t, eventing.Forwards, 1)

	// Test subscription details
	assert.Equal(t, "io.example.event.created", eventing.Subscriptions[0].EventType)
	assert.Equal(t, "io.example.service", eventing.Subscriptions[0].Source)

	// Test forward details
	assert.Equal(t, "query-processor", eventing.Forwards[0].TargetAgent)
	assert.Equal(t, "processing", eventing.Forwards[0].TargetNamespace)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ⚖️ AGENT SCALING TESTS                                                 │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAgentScalingSpec_Defaults(t *testing.T) {
	scaling := &AgentScalingSpec{
		MinReplicas:            0,
		MaxReplicas:            10,
		TargetConcurrency:      10,
		ScaleToZeroGracePeriod: "30s",
	}

	assert.Equal(t, int32(0), scaling.MinReplicas)
	assert.Equal(t, int32(10), scaling.MaxReplicas)
	assert.Equal(t, int32(10), scaling.TargetConcurrency)
	assert.Equal(t, "30s", scaling.ScaleToZeroGracePeriod)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 FULL AGENT OBJECT TESTS                                             │
// └─────────────────────────────────────────────────────────────────────────┘

func TestLambdaAgent_Initialization(t *testing.T) {
	agent := &LambdaAgent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-agent",
			Namespace: "agents",
		},
		Spec: LambdaAgentSpec{
			Image: AgentImageSpec{
				Repository: "localhost:5001/agents/test-agent",
				Tag:        "v1.0.0",
				Port:       8080,
			},
			AI: &AgentAISpec{
				Provider: "ollama",
				Endpoint: "http://ollama.ai.svc:11434",
				Model:    "llama2",
			},
			Behavior: &AgentBehaviorSpec{
				MaxContextMessages: 10,
				EmitEvents:         true,
				SystemPrompt:       "You are a helpful assistant",
			},
		},
	}

	assert.Equal(t, "test-agent", agent.Name)
	assert.Equal(t, "agents", agent.Namespace)
	assert.Equal(t, "localhost:5001/agents/test-agent", agent.Spec.Image.Repository)
	assert.Equal(t, "v1.0.0", agent.Spec.Image.Tag)
	require.NotNil(t, agent.Spec.AI)
	assert.Equal(t, "ollama", agent.Spec.AI.Provider)
	require.NotNil(t, agent.Spec.Behavior)
	assert.Equal(t, int32(10), agent.Spec.Behavior.MaxContextMessages)
	assert.True(t, agent.Spec.Behavior.EmitEvents)
}

func TestAgentServiceStatus_Fields(t *testing.T) {
	status := &AgentServiceStatus{
		ServiceName:    "test-agent",
		URL:            "http://test-agent.agents.svc.cluster.local",
		Ready:          true,
		LatestRevision: "test-agent-00001",
	}

	assert.Equal(t, "test-agent", status.ServiceName)
	assert.Equal(t, "http://test-agent.agents.svc.cluster.local", status.URL)
	assert.True(t, status.Ready)
	assert.Equal(t, "test-agent-00001", status.LatestRevision)
}

func TestAgentEventingStatus_Fields(t *testing.T) {
	status := &AgentEventingStatus{
		BrokerName:   "test-agent-broker",
		BrokerReady:  true,
		BrokerURL:    "http://test-agent-broker.agents.svc.cluster.local",
		TriggerCount: 3,
		ForwardCount: 1,
	}

	assert.Equal(t, "test-agent-broker", status.BrokerName)
	assert.True(t, status.BrokerReady)
	assert.Equal(t, "http://test-agent-broker.agents.svc.cluster.local", status.BrokerURL)
	assert.Equal(t, 3, status.TriggerCount)
	assert.Equal(t, 1, status.ForwardCount)
}
