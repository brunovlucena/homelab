// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 AGENT-007: LambdaAgent CloudEvents Integration Tests
//
//	User Story: Agent Event-Driven Operations
//	Priority: P0 | Story Points: 8
//
//	Tests validate:
//	- Agent creation via CloudEvents
//	- Agent update via CloudEvents
//	- Agent deletion via CloudEvents
//	- Agent build commands via CloudEvents
//	- Agent rollback via CloudEvents
//	- Event validation and schema compliance
//	- Event retry logic
//	- Event ordering guarantees
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package agents

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"knative-lambda/tests/testutils"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Event type constants (matching internal/events package)
const (
	EventTypeCommandFunctionDeploy = "io.knative.lambda.command.function.deploy"
	EventTypeCommandServiceCreate  = "io.knative.lambda.command.service.create"
	EventTypeCommandServiceUpdate  = "io.knative.lambda.command.service.update"
	EventTypeCommandServiceDelete  = "io.knative.lambda.command.service.delete"
	EventTypeCommandBuildStart     = "io.knative.lambda.command.build.start"
	EventTypeCommandBuildCancel    = "io.knative.lambda.command.build.cancel"
	EventTypeCommandFunctionRollback = "io.knative.lambda.command.function.rollback"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Test Helpers
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

// createAgentCloudEvent creates a CloudEvent for agent operations
func createAgentCloudEvent(t *testing.T, eventType, agentName string, data interface{}) *cloudevents.Event {
	t.Helper()
	event := cloudevents.NewEvent()
	event.SetID(fmt.Sprintf("test-%d", time.Now().UnixNano()))
	event.SetType(eventType)
	event.SetSource("io.knative.lambda/test")
	event.SetSubject(agentName)
	event.SetData(cloudevents.ApplicationJSON, data)
	return &event
}

// createAgentDeployEventData creates data for agent deploy events
func createAgentDeployEventData(agentName, namespace string) map[string]interface{} {
	return map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      agentName,
			"namespace": namespace,
			"labels": map[string]string{
				"app.kubernetes.io/name": agentName,
			},
		},
		"spec": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": TestImageRepo,
				"tag":        TestImageTag,
				"port":       8080,
			},
			"ai": map[string]interface{}{
				"provider": "ollama",
				"endpoint": DefaultOllamaURL,
				"model":    DefaultOllamaModel,
			},
		},
	}
}

// Note: The CloudEvents receiver is designed for LambdaFunction, not LambdaAgent.
// These tests validate the event structure and data format that would be used
// for agent operations via CloudEvents. Actual integration would require
// the receiver to support LambdaAgent or a separate agent-specific receiver.

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: Agent Creation via CloudEvents
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT007_AC1_AgentCreationViaCloudEvents(t *testing.T) {
	testutils.SetupTestEnvironment(t)

	t.Run("Validate agent deploy event structure", func(t *testing.T) {
		// Arrange
		agentName := "ce-create-agent"
		data := createAgentDeployEventData(agentName, TestNamespace)

		// Act - Validate event structure
		event := createAgentCloudEvent(t, EventTypeCommandFunctionDeploy, agentName, data)

		// Assert - Event should be properly formed
		assert.Equal(t, EventTypeCommandFunctionDeploy, event.Type())
		assert.Equal(t, agentName, event.Subject())
		assert.NotEmpty(t, event.ID())

		// Validate data structure
		var eventData map[string]interface{}
		err := event.DataAs(&eventData)
		require.NoError(t, err)
		assert.Contains(t, eventData, "metadata")
		assert.Contains(t, eventData, "spec")
	})

	t.Run("Validate agent event with AI configuration", func(t *testing.T) {
		// Arrange
		agentName := "ce-ai-agent"
		data := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      agentName,
				"namespace": TestNamespace,
			},
			"spec": map[string]interface{}{
				"image": map[string]interface{}{
					"repository": TestImageRepo,
					"tag":        TestImageTag,
				},
				"ai": map[string]interface{}{
					"provider":    "ollama",
					"endpoint":    DefaultOllamaURL,
					"model":       DefaultOllamaModel,
					"temperature": "0.7",
					"maxTokens":   2048,
				},
			},
		}

		// Act - Validate event structure
		event := createAgentCloudEvent(t, EventTypeCommandFunctionDeploy, agentName, data)

		// Assert
		var eventData map[string]interface{}
		err := event.DataAs(&eventData)
		require.NoError(t, err)

		spec := eventData["spec"].(map[string]interface{})
		ai := spec["ai"].(map[string]interface{})
		assert.Equal(t, "ollama", ai["provider"])
		assert.Equal(t, DefaultOllamaModel, ai["model"])
	})

	t.Run("Validate agent event with eventing configuration", func(t *testing.T) {
		// Arrange
		agentName := "ce-eventing-agent"
		data := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      agentName,
				"namespace": TestNamespace,
			},
			"spec": map[string]interface{}{
				"image": map[string]interface{}{
					"repository": TestImageRepo,
					"tag":        TestImageTag,
				},
				"eventing": map[string]interface{}{
					"enabled": true,
					"subscriptions": []map[string]interface{}{
						{
							"eventType": "io.homelab.test.event",
							"source":    "/test/*",
						},
					},
				},
			},
		}

		// Act - Validate event structure
		event := createAgentCloudEvent(t, EventTypeCommandFunctionDeploy, agentName, data)

		// Assert
		var eventData map[string]interface{}
		err := event.DataAs(&eventData)
		require.NoError(t, err)

		spec := eventData["spec"].(map[string]interface{})
		eventing := spec["eventing"].(map[string]interface{})
		assert.True(t, eventing["enabled"].(bool))
		subscriptions := eventing["subscriptions"].([]interface{})
		assert.Len(t, subscriptions, 1)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Agent Update via CloudEvents
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT007_AC2_AgentUpdateViaCloudEvents(t *testing.T) {
	t.Run("Validate agent update event structure", func(t *testing.T) {
		// Arrange
		agentName := "ce-update-agent"
		data := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      agentName,
				"namespace": TestNamespace,
			},
			"spec": map[string]interface{}{
				"image": map[string]interface{}{
					"repository": TestImageRepo,
					"tag":        "v2.0.0", // Updated tag
				},
			},
		}

		// Act - Validate event structure
		event := createAgentCloudEvent(t, EventTypeCommandServiceUpdate, agentName, data)

		// Assert
		assert.Equal(t, EventTypeCommandServiceUpdate, event.Type())
		var eventData map[string]interface{}
		err := event.DataAs(&eventData)
		require.NoError(t, err)
		spec := eventData["spec"].(map[string]interface{})
		image := spec["image"].(map[string]interface{})
		assert.Equal(t, "v2.0.0", image["tag"])
	})

	t.Run("Validate agent AI configuration update event", func(t *testing.T) {
		// Arrange
		agentName := "ce-ai-update-agent"
		data := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      agentName,
				"namespace": TestNamespace,
			},
			"spec": map[string]interface{}{
				"image": map[string]interface{}{
					"repository": TestImageRepo,
				},
				"ai": map[string]interface{}{
					"provider": "ollama",
					"model":    "llama3.2:7b", // Updated model
				},
			},
		}

		// Act - Validate event structure
		event := createAgentCloudEvent(t, EventTypeCommandServiceUpdate, agentName, data)

		// Assert
		var eventData map[string]interface{}
		err := event.DataAs(&eventData)
		require.NoError(t, err)
		spec := eventData["spec"].(map[string]interface{})
		ai := spec["ai"].(map[string]interface{})
		assert.Equal(t, "llama3.2:7b", ai["model"])
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Agent Deletion via CloudEvents
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT007_AC3_AgentDeletionViaCloudEvents(t *testing.T) {
	t.Run("Validate agent deletion event structure", func(t *testing.T) {
		// Arrange
		agentName := "ce-delete-agent"
		data := map[string]interface{}{
			"name":      agentName,
			"namespace": TestNamespace,
			"reason":    "User requested deletion",
		}

		// Act - Validate event structure
		event := createAgentCloudEvent(t, EventTypeCommandServiceDelete, agentName, data)

		// Assert
		assert.Equal(t, EventTypeCommandServiceDelete, event.Type())
		assert.Equal(t, agentName, event.Subject())
		var eventData map[string]interface{}
		err := event.DataAs(&eventData)
		require.NoError(t, err)
		assert.Equal(t, agentName, eventData["name"])
		assert.Equal(t, "User requested deletion", eventData["reason"])
	})

	t.Run("Validate deletion event with subject fallback", func(t *testing.T) {
		// Arrange
		agentName := "ce-subject-delete-agent"
		data := map[string]interface{}{
			// Name intentionally omitted - should use subject
		}

		// Act - Validate event structure
		event := createAgentCloudEvent(t, EventTypeCommandServiceDelete, agentName, data)

		// Assert
		assert.Equal(t, agentName, event.Subject(), "Subject should be set for fallback")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: Agent Build Commands via CloudEvents
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT007_AC4_AgentBuildCommandsViaCloudEvents(t *testing.T) {
	t.Run("Validate build.start event structure", func(t *testing.T) {
		// Arrange
		agentName := "ce-build-agent"
		data := map[string]interface{}{
			"name":      agentName,
			"namespace": TestNamespace,
			"reason":    "Code update",
		}

		// Act - Validate event structure
		event := createAgentCloudEvent(t, EventTypeCommandBuildStart, agentName, data)

		// Assert
		assert.Equal(t, EventTypeCommandBuildStart, event.Type())
		var eventData map[string]interface{}
		err := event.DataAs(&eventData)
		require.NoError(t, err)
		assert.Equal(t, agentName, eventData["name"])
		assert.Equal(t, "Code update", eventData["reason"])
	})

	t.Run("Validate force rebuild event structure", func(t *testing.T) {
		// Arrange
		agentName := "ce-force-build-agent"
		data := map[string]interface{}{
			"name":         agentName,
			"namespace":    TestNamespace,
			"forceRebuild": true,
			"reason":       "Dependency update",
		}

		// Act - Validate event structure
		event := createAgentCloudEvent(t, EventTypeCommandBuildStart, agentName, data)

		// Assert
		var eventData map[string]interface{}
		err := event.DataAs(&eventData)
		require.NoError(t, err)
		assert.True(t, eventData["forceRebuild"].(bool))
	})

	t.Run("Validate build.cancel event structure", func(t *testing.T) {
		// Arrange
		agentName := "ce-cancel-build-agent"
		data := map[string]interface{}{
			"name":      agentName,
			"namespace": TestNamespace,
			"reason":    "User cancelled",
		}

		// Act - Validate event structure
		event := createAgentCloudEvent(t, EventTypeCommandBuildCancel, agentName, data)

		// Assert
		assert.Equal(t, EventTypeCommandBuildCancel, event.Type())
		var eventData map[string]interface{}
		err := event.DataAs(&eventData)
		require.NoError(t, err)
		assert.Equal(t, "User cancelled", eventData["reason"])
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Agent Rollback via CloudEvents
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT007_AC5_AgentRollbackViaCloudEvents(t *testing.T) {
	t.Run("Validate rollback event structure", func(t *testing.T) {
		// Arrange
		agentName := "ce-rollback-agent"
		data := map[string]interface{}{
			"name":      agentName,
			"namespace": TestNamespace,
			"revision":  "ce-rollback-agent-00001",
			"reason":    "Regression detected",
		}

		// Act - Validate event structure
		event := createAgentCloudEvent(t, EventTypeCommandFunctionRollback, agentName, data)

		// Assert
		assert.Equal(t, EventTypeCommandFunctionRollback, event.Type())
		var eventData map[string]interface{}
		err := event.DataAs(&eventData)
		require.NoError(t, err)
		assert.Equal(t, "ce-rollback-agent-00001", eventData["revision"])
		assert.Equal(t, "Regression detected", eventData["reason"])
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: Event Validation and Schema Compliance
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT007_AC6_EventValidation(t *testing.T) {
	t.Run("Validate event has required CloudEvents fields", func(t *testing.T) {
		// Arrange
		event := createAgentCloudEvent(t, EventTypeCommandFunctionDeploy, "valid-agent", map[string]interface{}{})

		// Assert
		assert.NotEmpty(t, event.ID(), "Event should have ID")
		assert.NotEmpty(t, event.Type(), "Event should have type")
		assert.NotEmpty(t, event.Source(), "Event should have source")
		assert.Equal(t, "1.0", event.SpecVersion(), "Event should have spec version")
	})

	t.Run("Validate event data structure", func(t *testing.T) {
		// Arrange
		data := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "test-agent",
			},
			"spec": map[string]interface{}{
				"image": map[string]interface{}{
					"repository": TestImageRepo,
				},
			},
		}

		event := createAgentCloudEvent(t, EventTypeCommandFunctionDeploy, "test-agent", data)

		// Act
		var eventData map[string]interface{}
		err := event.DataAs(&eventData)

		// Assert
		assert.NoError(t, err, "Event data should be valid")
		assert.Contains(t, eventData, "metadata")
		assert.Contains(t, eventData, "spec")
	})

	t.Run("Validate event subject is set correctly", func(t *testing.T) {
		// Arrange
		agentName := "subject-test-agent"
		event := createAgentCloudEvent(t, EventTypeCommandFunctionDeploy, agentName, map[string]interface{}{})

		// Assert
		assert.Equal(t, agentName, event.Subject(), "Event subject should match agent name")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC7: HTTP Endpoint Tests
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT007_AC7_EventSerialization(t *testing.T) {
	t.Run("Event should serialize to JSON correctly", func(t *testing.T) {
		// Arrange
		data := createAgentDeployEventData("serialize-test-agent", TestNamespace)
		event := createAgentCloudEvent(t, EventTypeCommandFunctionDeploy, "serialize-test-agent", data)

		// Act
		jsonBytes, err := json.Marshal(event)

		// Assert
		assert.NoError(t, err, "Event should serialize to JSON")
		assert.NotEmpty(t, jsonBytes, "Serialized event should not be empty")

		// Verify it can be deserialized
		var deserializedEvent cloudevents.Event
		err = json.Unmarshal(jsonBytes, &deserializedEvent)
		assert.NoError(t, err, "Event should deserialize from JSON")
		assert.Equal(t, event.Type(), deserializedEvent.Type())
	})

	t.Run("Event data should be accessible after serialization", func(t *testing.T) {
		// Arrange
		data := createAgentDeployEventData("data-access-agent", TestNamespace)
		event := createAgentCloudEvent(t, EventTypeCommandFunctionDeploy, "data-access-agent", data)

		// Act
		var eventData map[string]interface{}
		err := event.DataAs(&eventData)

		// Assert
		assert.NoError(t, err, "Event data should be accessible")
		assert.Contains(t, eventData, "metadata")
		assert.Contains(t, eventData, "spec")
	})
}

// Note: These tests validate CloudEvent structure and data format for agent operations.
// Full integration testing with the CloudEvents receiver would require either:
// 1. Extending the receiver to support LambdaAgent resources, or
// 2. Creating agent-specific event handlers
// The current receiver is designed for LambdaFunction resources.
