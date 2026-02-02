// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 Unit Tests: LambdaAgent Controller
//
//	Tests for controller logic:
//	- Spec validation
//	- Phase transitions
//	- Condition management
//	- Knative Service building
//	- Eventing integration
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📝 CONSTANTS TESTS                                                     │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAgentFinalizerName(t *testing.T) {
	assert.NotEmpty(t, agentFinalizer, "Agent finalizer should be defined")
	assert.Contains(t, agentFinalizer, "lambdaagent", "Finalizer should reference lambdaagent")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🖼️ IMAGE URI GENERATION TESTS                                          │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAgentImageURIGeneration(t *testing.T) {
	tests := []struct {
		name        string
		imageSpec   lambdav1alpha1.AgentImageSpec
		expectedURI string
		description string
	}{
		{
			name: "Repository with digest",
			imageSpec: lambdav1alpha1.AgentImageSpec{
				Repository: "gcr.io/project/agent",
				Digest:     "sha256:abc123def456",
			},
			expectedURI: "gcr.io/project/agent@sha256:abc123def456",
			description: "Should use digest when provided",
		},
		{
			name: "Repository with tag",
			imageSpec: lambdav1alpha1.AgentImageSpec{
				Repository: "gcr.io/project/agent",
				Tag:        "v2.0.0",
			},
			expectedURI: "gcr.io/project/agent:v2.0.0",
			description: "Should use tag when no digest",
		},
		{
			name: "Repository only defaults to latest",
			imageSpec: lambdav1alpha1.AgentImageSpec{
				Repository: "gcr.io/project/agent",
			},
			expectedURI: "gcr.io/project/agent:latest",
			description: "Should default to :latest",
		},
		{
			name: "Digest takes precedence over tag",
			imageSpec: lambdav1alpha1.AgentImageSpec{
				Repository: "gcr.io/project/agent",
				Tag:        "v1.0.0",
				Digest:     "sha256:abc123",
			},
			expectedURI: "gcr.io/project/agent@sha256:abc123",
			description: "Digest should override tag",
		},
		{
			name: "Local registry",
			imageSpec: lambdav1alpha1.AgentImageSpec{
				Repository: "localhost:5001/my-agent",
				Tag:        "dev",
			},
			expectedURI: "localhost:5001/my-agent:dev",
			description: "Should work with local registry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the logic from reconcileDeploying
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
// │  🔍 SPEC VALIDATION TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAgentSpecValidation(t *testing.T) {
	tests := []struct {
		name        string
		agent       *lambdav1alpha1.LambdaAgent
		expectValid bool
		description string
	}{
		{
			name: "Valid agent with repository",
			agent: &lambdav1alpha1.LambdaAgent{
				Spec: lambdav1alpha1.LambdaAgentSpec{
					Image: lambdav1alpha1.AgentImageSpec{
						Repository: "gcr.io/project/my-agent",
						Tag:        "v1.0.0",
					},
				},
			},
			expectValid: true,
			description: "Agent with repository should be valid",
		},
		{
			name: "Invalid agent without repository",
			agent: &lambdav1alpha1.LambdaAgent{
				Spec: lambdav1alpha1.LambdaAgentSpec{
					Image: lambdav1alpha1.AgentImageSpec{
						Repository: "",
					},
				},
			},
			expectValid: false,
			description: "Agent without repository should be invalid",
		},
		{
			name: "Valid agent with AI configuration",
			agent: &lambdav1alpha1.LambdaAgent{
				Spec: lambdav1alpha1.LambdaAgentSpec{
					Image: lambdav1alpha1.AgentImageSpec{
						Repository: "localhost:5001/ai-agent",
					},
					AI: &lambdav1alpha1.AgentAISpec{
						Model:    "llama3.2",
						Endpoint: "http://ollama.ai.svc:11434",
					},
				},
			},
			expectValid: true,
			description: "Agent with AI config should be valid",
		},
		{
			name: "Valid agent with eventing",
			agent: &lambdav1alpha1.LambdaAgent{
				Spec: lambdav1alpha1.LambdaAgentSpec{
					Image: lambdav1alpha1.AgentImageSpec{
						Repository: "localhost:5001/event-agent",
					},
					Eventing: &lambdav1alpha1.AgentEventingSpec{
						Enabled: true,
						Subscriptions: []lambdav1alpha1.AgentSubscription{
							{EventType: "io.knative.lambda.invoke.sync"},
						},
					},
				},
			},
			expectValid: true,
			description: "Agent with eventing should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the validation logic from reconcilePending
			isValid := tt.agent.Spec.Image.Repository != ""
			assert.Equal(t, tt.expectValid, isValid, tt.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏗️ KNATIVE SERVICE BUILDING TESTS                                      │
// └─────────────────────────────────────────────────────────────────────────┘

func TestBuildKnativeServiceForAgent(t *testing.T) {
	r := &LambdaAgentReconciler{}

	t.Run("Basic agent service", func(t *testing.T) {
		agent := &lambdav1alpha1.LambdaAgent{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-agent",
				Namespace: "default",
			},
			Spec: lambdav1alpha1.LambdaAgentSpec{
				Image: lambdav1alpha1.AgentImageSpec{
					Repository: "gcr.io/project/agent",
					Tag:        "v1.0.0",
					Port:       8080,
				},
			},
		}

		ksvc := r.buildKnativeService(agent, "gcr.io/project/agent:v1.0.0")

		require.NotNil(t, ksvc)
		assert.Equal(t, "test-agent", ksvc.Name)
		assert.Equal(t, "default", ksvc.Namespace)
		assert.Equal(t, "knative-lambda-operator", ksvc.Labels["app.kubernetes.io/managed-by"])
		assert.Equal(t, "agent", ksvc.Labels["app.kubernetes.io/component"])

		// Verify container
		containers := ksvc.Spec.Template.Spec.Containers
		require.Len(t, containers, 1)
		assert.Equal(t, "gcr.io/project/agent:v1.0.0", containers[0].Image)
		assert.Equal(t, int32(8080), containers[0].Ports[0].ContainerPort)
	})

	t.Run("Agent with AI configuration", func(t *testing.T) {
		agent := &lambdav1alpha1.LambdaAgent{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ai-agent",
				Namespace: "default",
			},
			Spec: lambdav1alpha1.LambdaAgentSpec{
				Image: lambdav1alpha1.AgentImageSpec{
					Repository: "localhost:5001/ai-agent",
				},
				AI: &lambdav1alpha1.AgentAISpec{
					Model:    "llama3.2",
					Endpoint: "http://ollama.ai.svc:11434",
				},
			},
		}

		ksvc := r.buildKnativeService(agent, "localhost:5001/ai-agent:latest")

		// Verify AI env vars
		containers := ksvc.Spec.Template.Spec.Containers
		envMap := make(map[string]string)
		for _, env := range containers[0].Env {
			envMap[env.Name] = env.Value
		}

		assert.Equal(t, "http://ollama.ai.svc:11434", envMap["OLLAMA_URL"])
		assert.Equal(t, "llama3.2", envMap["OLLAMA_MODEL"])
	})

	t.Run("Agent with custom resources", func(t *testing.T) {
		agent := &lambdav1alpha1.LambdaAgent{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "resource-agent",
				Namespace: "default",
			},
			Spec: lambdav1alpha1.LambdaAgentSpec{
				Image: lambdav1alpha1.AgentImageSpec{
					Repository: "localhost:5001/agent",
				},
				Resources: &lambdav1alpha1.AgentResourcesSpec{
					Requests: &lambdav1alpha1.AgentResourceQuantity{
						Memory: "512Mi",
						CPU:    "250m",
					},
					Limits: &lambdav1alpha1.AgentResourceQuantity{
						Memory: "2Gi",
						CPU:    "1000m",
					},
				},
			},
		}

		ksvc := r.buildKnativeService(agent, "localhost:5001/agent:latest")

		containers := ksvc.Spec.Template.Spec.Containers
		require.Len(t, containers, 1)
		assert.NotNil(t, containers[0].Resources.Requests)
		assert.NotNil(t, containers[0].Resources.Limits)
	})

	t.Run("Agent with scaling configuration", func(t *testing.T) {
		agent := &lambdav1alpha1.LambdaAgent{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "scaled-agent",
				Namespace: "default",
			},
			Spec: lambdav1alpha1.LambdaAgentSpec{
				Image: lambdav1alpha1.AgentImageSpec{
					Repository: "localhost:5001/agent",
				},
				Scaling: &lambdav1alpha1.AgentScalingSpec{
					MinReplicas:       1,
					MaxReplicas:       10,
					TargetConcurrency: 20,
				},
			},
		}

		ksvc := r.buildKnativeService(agent, "localhost:5001/agent:latest")

		annotations := ksvc.Spec.Template.Annotations
		assert.Equal(t, "1", annotations["autoscaling.knative.dev/min-scale"])
		assert.Equal(t, "10", annotations["autoscaling.knative.dev/max-scale"])
		assert.Equal(t, "20", annotations["autoscaling.knative.dev/target"])
	})

	t.Run("Agent with behavior configuration", func(t *testing.T) {
		agent := &lambdav1alpha1.LambdaAgent{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "behavior-agent",
				Namespace: "default",
			},
			Spec: lambdav1alpha1.LambdaAgentSpec{
				Image: lambdav1alpha1.AgentImageSpec{
					Repository: "localhost:5001/agent",
				},
				Behavior: &lambdav1alpha1.AgentBehaviorSpec{
					EmitEvents:         true,
					MaxContextMessages: 50,
				},
			},
		}

		ksvc := r.buildKnativeService(agent, "localhost:5001/agent:latest")

		containers := ksvc.Spec.Template.Spec.Containers
		envMap := make(map[string]string)
		for _, env := range containers[0].Env {
			envMap[env.Name] = env.Value
		}

		assert.Equal(t, "true", envMap["EMIT_EVENTS"])
		assert.Equal(t, "50", envMap["MAX_CONTEXT_MESSAGES"])
	})

	t.Run("Agent with custom command and args", func(t *testing.T) {
		agent := &lambdav1alpha1.LambdaAgent{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "custom-agent",
				Namespace: "default",
			},
			Spec: lambdav1alpha1.LambdaAgentSpec{
				Image: lambdav1alpha1.AgentImageSpec{
					Repository: "localhost:5001/agent",
					Command:    []string{"python", "-m", "uvicorn"},
					Args:       []string{"main:app", "--host", "0.0.0.0"},
				},
			},
		}

		ksvc := r.buildKnativeService(agent, "localhost:5001/agent:latest")

		containers := ksvc.Spec.Template.Spec.Containers
		assert.Equal(t, []string{"python", "-m", "uvicorn"}, containers[0].Command)
		assert.Equal(t, []string{"main:app", "--host", "0.0.0.0"}, containers[0].Args)
	})

	t.Run("Agent with custom port", func(t *testing.T) {
		agent := &lambdav1alpha1.LambdaAgent{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "port-agent",
				Namespace: "default",
			},
			Spec: lambdav1alpha1.LambdaAgentSpec{
				Image: lambdav1alpha1.AgentImageSpec{
					Repository: "localhost:5001/agent",
					Port:       3000,
				},
			},
		}

		ksvc := r.buildKnativeService(agent, "localhost:5001/agent:latest")

		containers := ksvc.Spec.Template.Spec.Containers
		assert.Equal(t, int32(3000), containers[0].Ports[0].ContainerPort)

		// Verify probes use correct port
		assert.Equal(t, int32(3000), containers[0].ReadinessProbe.HTTPGet.Port.IntVal)
		assert.Equal(t, int32(3000), containers[0].LivenessProbe.HTTPGet.Port.IntVal)
	})
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📋 CONDITION MANAGEMENT TESTS                                          │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAgentSetCondition(t *testing.T) {
	r := &LambdaAgentReconciler{}

	tests := []struct {
		name            string
		initialAgent    *lambdav1alpha1.LambdaAgent
		conditionType   string
		status          metav1.ConditionStatus
		reason          string
		message         string
		expectedCount   int
		expectedStatus  metav1.ConditionStatus
		expectedReason  string
		expectedMessage string
	}{
		{
			name: "Set first condition",
			initialAgent: &lambdav1alpha1.LambdaAgent{
				Status: lambdav1alpha1.LambdaAgentStatus{},
			},
			conditionType:   "Ready",
			status:          metav1.ConditionFalse,
			reason:          "Deploying",
			message:         "Creating Knative Service",
			expectedCount:   1,
			expectedStatus:  metav1.ConditionFalse,
			expectedReason:  "Deploying",
			expectedMessage: "Creating Knative Service",
		},
		{
			name: "Update existing condition",
			initialAgent: &lambdav1alpha1.LambdaAgent{
				Status: lambdav1alpha1.LambdaAgentStatus{
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: metav1.ConditionFalse,
							Reason: "Deploying",
						},
					},
				},
			},
			conditionType:   "Ready",
			status:          metav1.ConditionTrue,
			reason:          "Ready",
			message:         "Agent is ready",
			expectedCount:   1,
			expectedStatus:  metav1.ConditionTrue,
			expectedReason:  "Ready",
			expectedMessage: "Agent is ready",
		},
		{
			name: "Add eventing condition",
			initialAgent: &lambdav1alpha1.LambdaAgent{
				Status: lambdav1alpha1.LambdaAgentStatus{
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: metav1.ConditionFalse,
						},
					},
				},
			},
			conditionType:   "Eventing",
			status:          metav1.ConditionTrue,
			reason:          "EventingReady",
			message:         "Eventing infrastructure ready",
			expectedCount:   2,
			expectedStatus:  metav1.ConditionTrue,
			expectedReason:  "EventingReady",
			expectedMessage: "Eventing infrastructure ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.setCondition(tt.initialAgent, tt.conditionType, tt.status, tt.reason, tt.message)

			assert.Len(t, tt.initialAgent.Status.Conditions, tt.expectedCount)

			cond := tt.initialAgent.Status.GetCondition(tt.conditionType)
			require.NotNil(t, cond, "Condition should exist")
			assert.Equal(t, tt.expectedStatus, cond.Status)
			assert.Equal(t, tt.expectedReason, cond.Reason)
			assert.Equal(t, tt.expectedMessage, cond.Message)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 PHASE TRANSITION TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAgentPhaseTransitions(t *testing.T) {
	// Define valid phase transitions
	validTransitions := map[lambdav1alpha1.LambdaAgentPhase][]lambdav1alpha1.LambdaAgentPhase{
		lambdav1alpha1.AgentPhasePending:   {lambdav1alpha1.AgentPhaseDeploying, lambdav1alpha1.AgentPhaseFailed},
		lambdav1alpha1.AgentPhaseDeploying: {lambdav1alpha1.AgentPhaseReady, lambdav1alpha1.AgentPhaseFailed},
		lambdav1alpha1.AgentPhaseReady:     {lambdav1alpha1.AgentPhaseFailed, lambdav1alpha1.AgentPhaseDeploying},
		lambdav1alpha1.AgentPhaseFailed:    {lambdav1alpha1.AgentPhasePending},
	}

	for fromPhase, toPhases := range validTransitions {
		for _, toPhase := range toPhases {
			t.Run(string(fromPhase)+"_to_"+string(toPhase), func(t *testing.T) {
				assert.NotEmpty(t, string(fromPhase))
				assert.NotEmpty(t, string(toPhase))
			})
		}
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🧪 ENV VAR HANDLING TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAgentEnvVarHandling(t *testing.T) {
	agent := &lambdav1alpha1.LambdaAgent{
		Spec: lambdav1alpha1.LambdaAgentSpec{
			Env: []corev1.EnvVar{
				{Name: "DEBUG", Value: "true"},
				{Name: "LOG_LEVEL", Value: "debug"},
				{
					Name: "API_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "api-secrets",
							},
							Key: "key",
						},
					},
				},
			},
		},
	}

	assert.Len(t, agent.Spec.Env, 3)
	assert.Equal(t, "DEBUG", agent.Spec.Env[0].Name)
	assert.Equal(t, "true", agent.Spec.Env[0].Value)
	assert.NotNil(t, agent.Spec.Env[2].ValueFrom)
	assert.NotNil(t, agent.Spec.Env[2].ValueFrom.SecretKeyRef)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🌐 EVENTING ENABLED LOGIC TESTS                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAgentEventingEnabled(t *testing.T) {
	tests := []struct {
		name            string
		eventing        *lambdav1alpha1.AgentEventingSpec
		expectedEnabled bool
		description     string
	}{
		{
			name:            "Nil eventing (disabled by default)",
			eventing:        nil,
			expectedEnabled: false,
			description:     "Agent with nil eventing should be disabled",
		},
		{
			name:            "Eventing explicitly enabled",
			eventing:        &lambdav1alpha1.AgentEventingSpec{Enabled: true},
			expectedEnabled: true,
			description:     "Agent with enabled eventing should be true",
		},
		{
			name:            "Eventing explicitly disabled",
			eventing:        &lambdav1alpha1.AgentEventingSpec{Enabled: false},
			expectedEnabled: false,
			description:     "Agent with disabled eventing should be false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &lambdav1alpha1.LambdaAgent{
				Spec: lambdav1alpha1.LambdaAgentSpec{
					Eventing: tt.eventing,
				},
			}

			eventingEnabled := agent.Spec.Eventing != nil && agent.Spec.Eventing.Enabled
			assert.Equal(t, tt.expectedEnabled, eventingEnabled, tt.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 TEST HELPER FUNCTIONS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func newTestLambdaAgent(name, namespace string) *lambdav1alpha1.LambdaAgent {
	return &lambdav1alpha1.LambdaAgent{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Generation: 1,
		},
		Spec: lambdav1alpha1.LambdaAgentSpec{
			Image: lambdav1alpha1.AgentImageSpec{
				Repository: "localhost:5001/test-agent",
				Tag:        "v1.0.0",
				Port:       8080,
			},
		},
		Status: lambdav1alpha1.LambdaAgentStatus{
			Phase: lambdav1alpha1.AgentPhasePending,
		},
	}
}

func TestNewTestLambdaAgent(t *testing.T) {
	agent := newTestLambdaAgent("test-agent", "default")

	assert.Equal(t, "test-agent", agent.Name)
	assert.Equal(t, "default", agent.Namespace)
	assert.Equal(t, "localhost:5001/test-agent", agent.Spec.Image.Repository)
	assert.Equal(t, "v1.0.0", agent.Spec.Image.Tag)
	assert.Equal(t, lambdav1alpha1.AgentPhasePending, agent.Status.Phase)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🤖 AI CONFIGURATION TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAgentAIConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		ai       *lambdav1alpha1.AgentAISpec
		expected map[string]string
	}{
		{
			name: "Full AI config",
			ai: &lambdav1alpha1.AgentAISpec{
				Model:    "llama3.2",
				Endpoint: "http://ollama.ai.svc:11434",
			},
			expected: map[string]string{
				"OLLAMA_URL":   "http://ollama.ai.svc:11434",
				"OLLAMA_MODEL": "llama3.2",
			},
		},
		{
			name: "Model only",
			ai: &lambdav1alpha1.AgentAISpec{
				Model: "mistral",
			},
			expected: map[string]string{
				"OLLAMA_MODEL": "mistral",
			},
		},
		{
			name:     "Nil AI config",
			ai:       nil,
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &lambdav1alpha1.LambdaAgent{
				ObjectMeta: metav1.ObjectMeta{Name: "ai-test", Namespace: "default"},
				Spec: lambdav1alpha1.LambdaAgentSpec{
					Image: lambdav1alpha1.AgentImageSpec{Repository: "test"},
					AI:    tt.ai,
				},
			}

			r := &LambdaAgentReconciler{}
			ksvc := r.buildKnativeService(agent, "test:latest")

			envMap := make(map[string]string)
			for _, env := range ksvc.Spec.Template.Spec.Containers[0].Env {
				envMap[env.Name] = env.Value
			}

			for key, expectedVal := range tt.expected {
				assert.Equal(t, expectedVal, envMap[key], "Expected env var %s=%s", key, expectedVal)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 OBSERVABILITY TESTS                                                 │
// └─────────────────────────────────────────────────────────────────────────┘

func TestAgentObservabilityConfiguration(t *testing.T) {
	agent := &lambdav1alpha1.LambdaAgent{
		ObjectMeta: metav1.ObjectMeta{Name: "otel-agent", Namespace: "default"},
		Spec: lambdav1alpha1.LambdaAgentSpec{
			Image: lambdav1alpha1.AgentImageSpec{Repository: "test"},
			Observability: &lambdav1alpha1.AgentObservabilitySpec{
				Tracing: &lambdav1alpha1.AgentTracingSpec{
					Enabled:  true,
					Endpoint: "alloy.observability.svc:4317",
				},
			},
		},
	}

	r := &LambdaAgentReconciler{}
	ksvc := r.buildKnativeService(agent, "test:latest")

	envMap := make(map[string]string)
	for _, env := range ksvc.Spec.Template.Spec.Containers[0].Env {
		envMap[env.Name] = env.Value
	}

	assert.Equal(t, "alloy.observability.svc:4317", envMap["OTEL_EXPORTER_OTLP_ENDPOINT"])
}
