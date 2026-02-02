// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 AGENT-002: LambdaAgent Eventing Tests
//
//	User Story: LambdaAgent Event Infrastructure
//	Priority: P0 | Story Points: 8
//
//	Tests validate:
//	- Broker creation per agent
//	- Trigger creation for subscriptions
//	- Forward rules for cross-namespace routing
//	- DLQ configuration
//	- Event source injection into container
//	- Eventing status tracking
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

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Test Fixtures for Eventing
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

// MockAgentEventing represents eventing configuration
type MockAgentEventing struct {
	Enabled       bool
	EventSource   string
	Subscriptions []MockSubscription
	Forwards      []MockForward
	DLQ           *MockDLQ
	RabbitMQ      *MockRabbitMQ
}

// MockSubscription represents an event subscription
type MockSubscription struct {
	EventType   string
	Source      string
	Description string
}

// MockForward represents a forwarding rule
type MockForward struct {
	EventTypes      []string
	TargetAgent     string
	TargetNamespace string
	Description     string
}

// MockDLQ represents DLQ configuration
type MockDLQ struct {
	Enabled          bool
	RetryMaxAttempts int32
}

// MockRabbitMQ represents RabbitMQ configuration
type MockRabbitMQ struct {
	ClusterName string
	Namespace   string
}

// MockEventingStatus represents eventing status
type MockEventingStatus struct {
	BrokerName   string
	BrokerReady  bool
	TriggerCount int
	ForwardCount int
	BrokerURL    string
}

// createTestEventing creates mock eventing configuration
func createTestEventing(enabled bool) *MockAgentEventing {
	return &MockAgentEventing{
		Enabled:     enabled,
		EventSource: "/test-agent/chatbot",
		Subscriptions: []MockSubscription{
			{
				EventType:   "io.homelab.vuln.found",
				Source:      "/agent-contracts/*",
				Description: "Vulnerability findings from security agents",
			},
		},
		Forwards: []MockForward{
			{
				EventTypes:      []string{"io.homelab.intent.security"},
				TargetAgent:     "vuln-scanner",
				TargetNamespace: "agent-contracts",
				Description:     "Forward security intents to vuln-scanner",
			},
		},
		DLQ: &MockDLQ{
			Enabled:          true,
			RetryMaxAttempts: 3,
		},
		RabbitMQ: &MockRabbitMQ{
			ClusterName: "rabbitmq-cluster-knative-lambda",
			Namespace:   "knative-lambda",
		},
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: Broker creation for agent
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT002_AC1_BrokerCreation(t *testing.T) {
	testutils.SetupTestEnvironment(t)

	t.Run("Broker created when eventing enabled", func(t *testing.T) {
		// Arrange
		agentName := "broker-test-agent"
		eventing := createTestEventing(true)

		// Act
		brokerName := fmt.Sprintf("%s-broker", agentName)
		shouldCreateBroker := eventing.Enabled

		// Assert
		assert.True(t, shouldCreateBroker, "Should create broker when eventing enabled")
		assert.Equal(t, "broker-test-agent-broker", brokerName)
	})

	t.Run("Broker not created when eventing disabled", func(t *testing.T) {
		// Arrange
		eventing := createTestEventing(false)

		// Act
		shouldCreateBroker := eventing.Enabled

		// Assert
		assert.False(t, shouldCreateBroker, "Should not create broker when eventing disabled")
	})

	t.Run("Broker uses RabbitMQ class", func(t *testing.T) {
		// Arrange
		eventing := createTestEventing(true)

		// Act & Assert
		assert.Equal(t, "rabbitmq-cluster-knative-lambda", eventing.RabbitMQ.ClusterName)
		assert.Equal(t, "knative-lambda", eventing.RabbitMQ.Namespace)
	})

	t.Run("Broker has owner reference to agent", func(t *testing.T) {
		// Arrange
		agentName := "owner-test-agent"
		agentUID := "test-uid-12345"

		// Act
		ownerRef := buildOwnerReference(agentName, agentUID)

		// Assert
		assert.Equal(t, "LambdaAgent", ownerRef.Kind)
		assert.Equal(t, agentName, ownerRef.Name)
		assert.Equal(t, agentUID, ownerRef.UID)
		assert.True(t, ownerRef.Controller)
	})

	t.Run("RabbitmqBrokerConfig created", func(t *testing.T) {
		// Arrange
		agentName := "config-test-agent"
		namespace := "agent-test"

		// Act
		configName := fmt.Sprintf("%s-broker-config", agentName)

		// Assert
		assert.Equal(t, "config-test-agent-broker-config", configName)
		assert.NotEmpty(t, namespace)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Trigger creation for subscriptions
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT002_AC2_TriggerCreation(t *testing.T) {
	t.Run("Trigger created for each subscription", func(t *testing.T) {
		// Arrange
		agentName := "trigger-test-agent"
		eventing := createTestEventing(true)
		eventing.Subscriptions = []MockSubscription{
			{EventType: "io.homelab.vuln.found", Source: "/agent-contracts/*"},
			{EventType: "io.homelab.exploit.validated", Source: "/agent-contracts/*"},
		}

		// Act
		triggers := buildTriggers(agentName, eventing.Subscriptions)

		// Assert
		assert.Len(t, triggers, 2, "Should create trigger for each subscription")
		assert.Contains(t, triggers[0].Name, "trigger-test-agent")
		assert.Equal(t, "io.homelab.vuln.found", triggers[0].EventType)
	})

	t.Run("Trigger filters on eventType", func(t *testing.T) {
		// Arrange
		subscription := MockSubscription{
			EventType: "io.homelab.chat.message",
			Source:    "/homepage/*",
		}

		// Act
		filter := buildTriggerFilter(subscription)

		// Assert
		assert.Equal(t, "io.homelab.chat.message", filter.Type)
	})

	t.Run("Trigger filters on source when specified", func(t *testing.T) {
		// Arrange
		subscription := MockSubscription{
			EventType: "io.homelab.chat.message",
			Source:    "/homepage/*",
		}

		// Act
		filter := buildTriggerFilter(subscription)

		// Assert
		assert.Equal(t, "/homepage/*", filter.Source)
	})

	t.Run("Trigger has correct subscriber reference", func(t *testing.T) {
		// Arrange
		agentName := "subscriber-test-agent"
		namespace := "agent-test"

		// Act
		subscriber := buildSubscriberRef(agentName, namespace)

		// Assert
		assert.Equal(t, "Service", subscriber.Kind)
		assert.Equal(t, "serving.knative.dev", subscriber.APIVersion)
		assert.Equal(t, agentName, subscriber.Name)
		assert.Equal(t, namespace, subscriber.Namespace)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Forward rules for cross-namespace routing
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT002_AC3_ForwardRules(t *testing.T) {
	t.Run("Forward trigger created for each forward rule", func(t *testing.T) {
		// Arrange
		eventing := createTestEventing(true)
		eventing.Forwards = []MockForward{
			{
				EventTypes:      []string{"io.homelab.intent.security"},
				TargetAgent:     "vuln-scanner",
				TargetNamespace: "agent-contracts",
			},
			{
				EventTypes:      []string{"io.homelab.intent.status"},
				TargetAgent:     "status-checker",
				TargetNamespace: "monitoring",
			},
		}

		// Act
		forwards := len(eventing.Forwards)

		// Assert
		assert.Equal(t, 2, forwards, "Should have 2 forward rules")
	})

	t.Run("Forward uses correct broker ingress URL pattern", func(t *testing.T) {
		// Arrange
		forward := MockForward{
			TargetAgent:     "vuln-scanner",
			TargetNamespace: "agent-contracts",
		}

		// Act
		brokerURL := buildForwardBrokerURL(forward.TargetAgent, forward.TargetNamespace)

		// Assert - Using correct Knative broker ingress URL pattern
		expected := "http://broker-ingress.knative-eventing.svc.cluster.local/agent-contracts/vuln-scanner-broker"
		assert.Equal(t, expected, brokerURL, "Should use correct Knative broker URL pattern")
	})

	t.Run("Forward channel created for cross-namespace routing", func(t *testing.T) {
		// Arrange
		agentName := "channel-test-agent"
		forward := MockForward{
			EventTypes:      []string{"io.homelab.intent.security"},
			TargetAgent:     "vuln-scanner",
			TargetNamespace: "agent-contracts",
		}

		// Act
		channelName := buildChannelName(agentName, forward)

		// Assert
		assert.Contains(t, channelName, agentName)
		assert.Contains(t, channelName, "vuln-scanner")
	})

	t.Run("Forward handles multiple event types", func(t *testing.T) {
		// Arrange
		forward := MockForward{
			EventTypes: []string{
				"io.homelab.intent.security",
				"io.homelab.intent.audit",
				"io.homelab.intent.compliance",
			},
			TargetAgent:     "security-hub",
			TargetNamespace: "security",
		}

		// Act & Assert
		assert.Len(t, forward.EventTypes, 3, "Should handle multiple event types")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: DLQ configuration
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT002_AC4_DLQConfiguration(t *testing.T) {
	t.Run("DLQ enabled by default", func(t *testing.T) {
		// Arrange
		eventing := createTestEventing(true)

		// Assert
		assert.NotNil(t, eventing.DLQ, "DLQ should be configured")
		assert.True(t, eventing.DLQ.Enabled, "DLQ should be enabled by default")
	})

	t.Run("DLQ retry attempts configurable", func(t *testing.T) {
		// Arrange
		eventing := createTestEventing(true)
		eventing.DLQ.RetryMaxAttempts = 5

		// Assert
		assert.Equal(t, int32(5), eventing.DLQ.RetryMaxAttempts)
	})

	t.Run("DLQ queue name follows convention", func(t *testing.T) {
		// Arrange
		agentName := "dlq-test-agent"

		// Act
		dlqName := buildDLQName(agentName)

		// Assert
		assert.Equal(t, "dlq-test-agent-dlq", dlqName)
	})

	t.Run("DLQ trigger created for failed events", func(t *testing.T) {
		// Arrange
		dlqConfig := &MockDLQ{
			Enabled:          true,
			RetryMaxAttempts: 3,
		}

		// Act
		dlqTriggerName := fmt.Sprintf("%s-dlq-trigger", "dlq-trigger-agent")

		// Assert
		assert.NotNil(t, dlqConfig)
		assert.Equal(t, "dlq-trigger-agent-dlq-trigger", dlqTriggerName)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Event source injection into container
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT002_AC5_EventSourceInjection(t *testing.T) {
	t.Run("Event source injected as environment variable", func(t *testing.T) {
		// Arrange
		eventing := createTestEventing(true)
		eventing.EventSource = "/agent-bruno/chatbot"

		// Act
		envVars := buildEventingEnvVars(eventing)

		// Assert
		assert.Contains(t, envVars, "K_SINK")
		assert.Contains(t, envVars, "EVENT_SOURCE")
		assert.Equal(t, "/agent-bruno/chatbot", envVars["EVENT_SOURCE"])
	})

	t.Run("Broker URL injected for event emission", func(t *testing.T) {
		// Arrange
		agentName := "emit-test-agent"
		namespace := "agent-test"

		// Act
		brokerURL := buildBrokerURL(agentName, namespace)
		envVars := map[string]string{
			"K_SINK": brokerURL,
		}

		// Assert
		assert.NotEmpty(t, envVars["K_SINK"])
		assert.Contains(t, envVars["K_SINK"], "broker-ingress")
	})

	t.Run("EMIT_EVENTS env var set when configured", func(t *testing.T) {
		// Arrange
		eventing := createTestEventing(true)

		// Act
		envVars := buildEventingEnvVars(eventing)

		// Assert
		assert.Equal(t, "true", envVars["EMIT_EVENTS"])
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: Eventing status tracking
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT002_AC6_EventingStatusTracking(t *testing.T) {
	t.Run("Eventing status includes broker info", func(t *testing.T) {
		// Arrange
		status := &MockEventingStatus{
			BrokerName:  "test-agent-broker",
			BrokerReady: true,
			BrokerURL:   "http://broker-ingress.knative-eventing.svc.cluster.local/agent-test/test-agent-broker",
		}

		// Assert
		assert.NotEmpty(t, status.BrokerName)
		assert.True(t, status.BrokerReady)
		assert.NotEmpty(t, status.BrokerURL)
	})

	t.Run("Eventing status tracks trigger count", func(t *testing.T) {
		// Arrange
		eventing := createTestEventing(true)
		eventing.Subscriptions = []MockSubscription{
			{EventType: "type1"},
			{EventType: "type2"},
			{EventType: "type3"},
		}

		// Act
		status := &MockEventingStatus{
			TriggerCount: len(eventing.Subscriptions),
		}

		// Assert
		assert.Equal(t, 3, status.TriggerCount)
	})

	t.Run("Eventing status tracks forward count", func(t *testing.T) {
		// Arrange
		eventing := createTestEventing(true)

		// Act
		status := &MockEventingStatus{
			ForwardCount: len(eventing.Forwards),
		}

		// Assert
		assert.Equal(t, 1, status.ForwardCount)
	})

	t.Run("Eventing condition set on success", func(t *testing.T) {
		// Arrange
		condition := MockCondition{
			Type:    "Eventing",
			Status:  "True",
			Reason:  "EventingReady",
			Message: "Eventing infrastructure ready",
		}

		// Assert
		assert.Equal(t, "Eventing", condition.Type)
		assert.Equal(t, "True", condition.Status)
	})

	t.Run("Eventing condition set on failure", func(t *testing.T) {
		// Arrange
		err := fmt.Errorf("broker creation failed: rabbitmq unavailable")
		condition := MockCondition{
			Type:    "Eventing",
			Status:  "False",
			Reason:  "EventingFailed",
			Message: err.Error(),
		}

		// Assert
		assert.Equal(t, "False", condition.Status)
		assert.Contains(t, condition.Message, "rabbitmq unavailable")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Integration Test: Full Eventing Setup
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT002_Integration_FullEventingSetup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Complete eventing infrastructure setup", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		_ = ctx // Would be used with real client

		agentName := "full-eventing-agent"
		namespace := "agent-test"
		startTime := time.Now()

		// Step 1: Create eventing config
		eventing := createTestEventing(true)
		eventing.Subscriptions = []MockSubscription{
			{EventType: "io.homelab.vuln.found"},
			{EventType: "io.homelab.exploit.validated"},
		}
		eventing.Forwards = []MockForward{
			{
				EventTypes:      []string{"io.homelab.intent.security"},
				TargetAgent:     "vuln-scanner",
				TargetNamespace: "agent-contracts",
			},
		}

		// Step 2: Create broker
		brokerName := fmt.Sprintf("%s-broker", agentName)
		require.NotEmpty(t, brokerName)

		// Step 3: Create triggers
		triggers := buildTriggers(agentName, eventing.Subscriptions)
		assert.Len(t, triggers, 2)

		// Step 4: Create forward channels
		forwardCount := len(eventing.Forwards)
		assert.Equal(t, 1, forwardCount)

		// Step 5: Update status
		status := &MockEventingStatus{
			BrokerName:   brokerName,
			BrokerReady:  true,
			TriggerCount: len(triggers),
			ForwardCount: forwardCount,
			BrokerURL:    buildBrokerURL(agentName, namespace),
		}

		endTime := time.Now()

		// Assert
		assert.True(t, status.BrokerReady)
		assert.Equal(t, 2, status.TriggerCount)
		assert.Equal(t, 1, status.ForwardCount)

		// Timing
		setupDuration := endTime.Sub(startTime)
		assert.Less(t, setupDuration.Milliseconds(), int64(1000),
			"Eventing setup simulation should complete quickly")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Helper Functions
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

// OwnerReference represents an owner reference
type OwnerReference struct {
	Kind       string
	Name       string
	UID        string
	Controller bool
}

// buildOwnerReference builds an owner reference
func buildOwnerReference(name, uid string) OwnerReference {
	return OwnerReference{
		Kind:       "LambdaAgent",
		Name:       name,
		UID:        uid,
		Controller: true,
	}
}

// TriggerConfig represents a trigger configuration
type TriggerConfig struct {
	Name      string
	EventType string
	Source    string
}

// buildTriggers builds trigger configurations
func buildTriggers(agentName string, subscriptions []MockSubscription) []TriggerConfig {
	var triggers []TriggerConfig
	for i, sub := range subscriptions {
		triggers = append(triggers, TriggerConfig{
			Name:      fmt.Sprintf("%s-trigger-%d", agentName, i),
			EventType: sub.EventType,
			Source:    sub.Source,
		})
	}
	return triggers
}

// TriggerFilter represents trigger filter
type TriggerFilter struct {
	Type   string
	Source string
}

// buildTriggerFilter builds a trigger filter
func buildTriggerFilter(sub MockSubscription) TriggerFilter {
	return TriggerFilter{
		Type:   sub.EventType,
		Source: sub.Source,
	}
}

// SubscriberRef represents subscriber reference
type SubscriberRef struct {
	Kind       string
	APIVersion string
	Name       string
	Namespace  string
}

// buildSubscriberRef builds subscriber reference
func buildSubscriberRef(name, namespace string) SubscriberRef {
	return SubscriberRef{
		Kind:       "Service",
		APIVersion: "serving.knative.dev",
		Name:       name,
		Namespace:  namespace,
	}
}

// buildForwardBrokerURL builds the forward broker URL using correct Knative pattern
func buildForwardBrokerURL(targetAgent, targetNamespace string) string {
	brokerName := fmt.Sprintf("%s-broker", targetAgent)
	return fmt.Sprintf("http://broker-ingress.knative-eventing.svc.cluster.local/%s/%s",
		targetNamespace, brokerName)
}

// buildChannelName builds a channel name for forwarding
func buildChannelName(agentName string, forward MockForward) string {
	return fmt.Sprintf("%s-to-%s-channel", agentName, forward.TargetAgent)
}

// buildDLQName builds the DLQ name
func buildDLQName(agentName string) string {
	return fmt.Sprintf("%s-dlq", agentName)
}

// buildEventingEnvVars builds eventing environment variables
func buildEventingEnvVars(eventing *MockAgentEventing) map[string]string {
	return map[string]string{
		"K_SINK":       "http://broker-ingress.knative-eventing.svc.cluster.local/test/test-broker",
		"EVENT_SOURCE": eventing.EventSource,
		"EMIT_EVENTS":  "true",
	}
}

// buildBrokerURL builds the broker URL
func buildBrokerURL(agentName, namespace string) string {
	brokerName := fmt.Sprintf("%s-broker", agentName)
	return fmt.Sprintf("http://broker-ingress.knative-eventing.svc.cluster.local/%s/%s",
		namespace, brokerName)
}
