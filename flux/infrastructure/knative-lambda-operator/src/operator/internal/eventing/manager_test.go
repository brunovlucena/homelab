// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 Unit Tests: Eventing Manager
//
//	Tests for eventing operations:
//	- Broker/Trigger naming conventions
//	- Configuration building
//	- RabbitMQ configuration
//	- DLQ configuration
//	- Event type sanitization
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package eventing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📝 CONSTANTS TESTS                                                     │
// └─────────────────────────────────────────────────────────────────────────┘

func TestEventingConstants(t *testing.T) {
	assert.Equal(t, "lambda-broker", SharedBrokerName, "Shared broker name should be lambda-broker")
	assert.Equal(t, "lambda-dlq", SharedDLQPrefix, "DLQ prefix should be lambda-dlq")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ⚙️ DEFAULT CONFIG TESTS                                                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// RabbitMQ defaults
	assert.Equal(t, "rabbitmq", config.DefaultRabbitMQCluster)
	assert.Equal(t, "rabbitmq-system", config.DefaultRabbitMQNamespace)
	assert.Equal(t, "quorum", config.DefaultRabbitMQQueueType)
	assert.Equal(t, 50, config.DefaultRabbitMQParallelism)

	// DLQ defaults
	assert.True(t, config.DefaultDLQEnabled)
	assert.Equal(t, "lambda-dlq-exchange", config.DefaultDLQExchangeName)
	assert.Equal(t, "lambda-dlq-queue", config.DefaultDLQQueueName)
	assert.Equal(t, "io.knative.lambda.dlq", config.DefaultDLQRoutingKeyPrefix)
	assert.Equal(t, 5, config.DefaultDLQRetryMaxAttempts)
	assert.Equal(t, "PT1S", config.DefaultDLQRetryBackoffDelay)
	assert.Equal(t, 604800000, config.DefaultDLQMessageTTL) // 7 days
	assert.Equal(t, 100000, config.DefaultDLQMaxLength)

	// Event source
	assert.Equal(t, "io.knative.lambda/operator", config.DefaultEventSource)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏷️ NAMING CONVENTION TESTS                                             │
// └─────────────────────────────────────────────────────────────────────────┘

func TestGetSharedBrokerName(t *testing.T) {
	m := &Manager{config: DefaultConfig()}

	tests := []struct {
		name         string
		lambda       *lambdav1alpha1.LambdaFunction
		expectedName string
		description  string
	}{
		{
			name: "No custom broker name",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{},
			},
			expectedName: SharedBrokerName,
			description:  "Should use shared broker name by default",
		},
		{
			name: "Nil eventing spec",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: nil,
				},
			},
			expectedName: SharedBrokerName,
			description:  "Should use shared broker name when eventing is nil",
		},
		{
			name: "Custom broker name",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: &lambdav1alpha1.EventingSpec{
						BrokerName: "my-custom-broker",
					},
				},
			},
			expectedName: "my-custom-broker",
			description:  "Should use custom broker name when specified",
		},
		{
			name: "Empty broker name in spec",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: &lambdav1alpha1.EventingSpec{
						BrokerName: "",
					},
				},
			},
			expectedName: SharedBrokerName,
			description:  "Should use shared broker name when custom is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.getSharedBrokerName(tt.lambda)
			assert.Equal(t, tt.expectedName, result, tt.description)
		})
	}
}

func TestGetLambdaTriggerName(t *testing.T) {
	m := &Manager{}

	tests := []struct {
		name         string
		lambda       *lambdav1alpha1.LambdaFunction
		expectedName string
	}{
		{
			name: "Simple function name",
			lambda: &lambdav1alpha1.LambdaFunction{
				ObjectMeta: metav1.ObjectMeta{Name: "my-function"},
			},
			expectedName: "my-function-trigger",
		},
		{
			name: "Function with hyphens",
			lambda: &lambdav1alpha1.LambdaFunction{
				ObjectMeta: metav1.ObjectMeta{Name: "my-python-function"},
			},
			expectedName: "my-python-function-trigger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.getLambdaTriggerName(tt.lambda)
			assert.Equal(t, tt.expectedName, result)
		})
	}
}

func TestGetAgentBrokerName(t *testing.T) {
	m := &Manager{}

	tests := []struct {
		name         string
		agent        *lambdav1alpha1.LambdaAgent
		expectedName string
	}{
		{
			name: "Simple agent name",
			agent: &lambdav1alpha1.LambdaAgent{
				ObjectMeta: metav1.ObjectMeta{Name: "query-agent"},
			},
			expectedName: "query-agent-broker",
		},
		{
			name: "Agent with complex name",
			agent: &lambdav1alpha1.LambdaAgent{
				ObjectMeta: metav1.ObjectMeta{Name: "multi-purpose-ai-agent"},
			},
			expectedName: "multi-purpose-ai-agent-broker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.getAgentBrokerName(tt.agent)
			assert.Equal(t, tt.expectedName, result)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔤 EVENT TYPE SANITIZATION TESTS                                       │
// └─────────────────────────────────────────────────────────────────────────┘

func TestSanitizeEventType(t *testing.T) {
	tests := []struct {
		name           string
		eventType      string
		expectedResult string
		description    string
	}{
		{
			name:           "Dots converted to dashes",
			eventType:      "io.knative.lambda.invoke",
			expectedResult: "io-knative-lambda-invoke",
			description:    "Should convert dots to dashes",
		},
		{
			name:           "Slashes converted to dashes",
			eventType:      "io/knative/lambda",
			expectedResult: "io-knative-lambda",
			description:    "Should convert slashes to dashes",
		},
		{
			name:           "Underscores converted to dashes",
			eventType:      "event_type_name",
			expectedResult: "event-type-name",
			description:    "Should convert underscores to dashes",
		},
		{
			name:           "Uppercase converted to lowercase",
			eventType:      "MyEventType",
			expectedResult: "myeventtype",
			description:    "Should convert to lowercase",
		},
		{
			name:           "Long name truncated",
			eventType:      "this.is.a.very.long.event.type.name.that.exceeds.the.maximum.allowed.length",
			expectedResult: "this-is-a-very-long-event-type-name-that", // 40 chars (truncated at 40)
			description:    "Should truncate long names",
		},
		{
			name:           "Simple event type",
			eventType:      "created",
			expectedResult: "created",
			description:    "Simple names should pass through",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeEventType(tt.eventType)
			assert.Equal(t, tt.expectedResult, result, tt.description)
			assert.LessOrEqual(t, len(result), 40, "Result should be at most 40 characters")
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🐰 RABBITMQ CONFIG TESTS                                               │
// └─────────────────────────────────────────────────────────────────────────┘

func TestBuildRabbitMQConfig(t *testing.T) {
	m := &Manager{config: DefaultConfig()}

	tests := []struct {
		name           string
		lambda         *lambdav1alpha1.LambdaFunction
		expectedConfig RabbitMQConfig
		description    string
	}{
		{
			name: "Default RabbitMQ config",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{},
			},
			expectedConfig: RabbitMQConfig{
				ClusterName: "rabbitmq",
				Namespace:   "rabbitmq-system",
				QueueType:   "quorum",
				Parallelism: 50,
			},
			description: "Should use defaults when no config specified",
		},
		{
			name: "Custom RabbitMQ cluster",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: &lambdav1alpha1.EventingSpec{
						RabbitMQ: &lambdav1alpha1.RabbitMQSpec{
							ClusterName: "my-rabbitmq",
							Namespace:   "my-namespace",
						},
					},
				},
			},
			expectedConfig: RabbitMQConfig{
				ClusterName: "my-rabbitmq",
				Namespace:   "my-namespace",
				QueueType:   "quorum", // Default
				Parallelism: 50,       // Default
			},
			description: "Should use custom cluster when specified",
		},
		{
			name: "Custom queue type",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: &lambdav1alpha1.EventingSpec{
						RabbitMQ: &lambdav1alpha1.RabbitMQSpec{
							QueueType: "classic",
						},
					},
				},
			},
			expectedConfig: RabbitMQConfig{
				ClusterName: "rabbitmq",
				Namespace:   "rabbitmq-system",
				QueueType:   "classic",
				Parallelism: 50,
			},
			description: "Should use custom queue type",
		},
		{
			name: "Custom parallelism",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: &lambdav1alpha1.EventingSpec{
						RabbitMQ: &lambdav1alpha1.RabbitMQSpec{
							Parallelism: 50,
						},
					},
				},
			},
			expectedConfig: RabbitMQConfig{
				ClusterName: "rabbitmq",
				Namespace:   "rabbitmq-system",
				QueueType:   "quorum",
				Parallelism: 50,
			},
			description: "Should use custom parallelism",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.buildRabbitMQConfig(tt.lambda)
			assert.Equal(t, tt.expectedConfig.ClusterName, result.ClusterName, tt.description)
			assert.Equal(t, tt.expectedConfig.Namespace, result.Namespace, tt.description)
			assert.Equal(t, tt.expectedConfig.QueueType, result.QueueType, tt.description)
			assert.Equal(t, tt.expectedConfig.Parallelism, result.Parallelism, tt.description)
		})
	}
}

func TestGetParallelism(t *testing.T) {
	m := &Manager{config: DefaultConfig()}

	tests := []struct {
		name        string
		lambda      *lambdav1alpha1.LambdaFunction
		expected    int
		description string
	}{
		{
			name:        "Default parallelism",
			lambda:      &lambdav1alpha1.LambdaFunction{},
			expected:    50,
			description: "Should return default parallelism",
		},
		{
			name: "Custom parallelism",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: &lambdav1alpha1.EventingSpec{
						RabbitMQ: &lambdav1alpha1.RabbitMQSpec{
							Parallelism: 25,
						},
					},
				},
			},
			expected:    25,
			description: "Should return custom parallelism",
		},
		{
			name: "Zero parallelism uses default",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: &lambdav1alpha1.EventingSpec{
						RabbitMQ: &lambdav1alpha1.RabbitMQSpec{
							Parallelism: 0,
						},
					},
				},
			},
			expected:    50,
			description: "Zero parallelism should use default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.getParallelism(tt.lambda)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📭 DLQ CONFIG TESTS                                                    │
// └─────────────────────────────────────────────────────────────────────────┘

func TestIsDLQEnabled(t *testing.T) {
	m := &Manager{config: DefaultConfig()}

	tests := []struct {
		name        string
		lambda      *lambdav1alpha1.LambdaFunction
		expected    bool
		description string
	}{
		{
			name:        "Default DLQ enabled",
			lambda:      &lambdav1alpha1.LambdaFunction{},
			expected:    true,
			description: "DLQ should be enabled by default",
		},
		{
			name: "Nil DLQ spec uses default",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: &lambdav1alpha1.EventingSpec{
						DLQ: nil,
					},
				},
			},
			expected:    true,
			description: "Nil DLQ spec should use default (enabled)",
		},
		{
			name: "DLQ explicitly enabled",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: &lambdav1alpha1.EventingSpec{
						DLQ: &lambdav1alpha1.DLQSpec{
							Enabled: true,
						},
					},
				},
			},
			expected:    true,
			description: "DLQ should be enabled when explicitly set",
		},
		{
			name: "DLQ explicitly disabled",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: &lambdav1alpha1.EventingSpec{
						DLQ: &lambdav1alpha1.DLQSpec{
							Enabled: false,
						},
					},
				},
			},
			expected:    false,
			description: "DLQ should be disabled when explicitly set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.isDLQEnabled(tt.lambda)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestBuildDLQConfig(t *testing.T) {
	m := &Manager{config: DefaultConfig()}

	tests := []struct {
		name        string
		lambda      *lambdav1alpha1.LambdaFunction
		checkFields func(t *testing.T, config DLQConfig)
		description string
	}{
		{
			name:   "Default DLQ config",
			lambda: &lambdav1alpha1.LambdaFunction{},
			checkFields: func(t *testing.T, config DLQConfig) {
				assert.True(t, config.Enabled)
				assert.Equal(t, "lambda-dlq-exchange", config.ExchangeName)
				assert.Equal(t, "lambda-dlq-queue", config.QueueName)
				assert.Equal(t, 5, config.RetryMaxAttempts)
				assert.Equal(t, "PT1S", config.RetryBackoffDelay)
				assert.Equal(t, 604800000, config.MessageTTL)
				assert.Equal(t, 100000, config.MaxLength)
				assert.Equal(t, "reject-publish", config.OverflowPolicy)
			},
			description: "Should use defaults when no DLQ config specified",
		},
		{
			name: "Custom DLQ exchange and queue names",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: &lambdav1alpha1.EventingSpec{
						DLQ: &lambdav1alpha1.DLQSpec{
							Enabled:      true,
							ExchangeName: "my-dlq-exchange",
							QueueName:    "my-dlq-queue",
						},
					},
				},
			},
			checkFields: func(t *testing.T, config DLQConfig) {
				assert.Equal(t, "my-dlq-exchange", config.ExchangeName)
				assert.Equal(t, "my-dlq-queue", config.QueueName)
			},
			description: "Should use custom exchange and queue names",
		},
		{
			name: "Custom retry configuration",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: &lambdav1alpha1.EventingSpec{
						DLQ: &lambdav1alpha1.DLQSpec{
							Enabled:           true,
							RetryMaxAttempts:  10,
							RetryBackoffDelay: "PT5S",
						},
					},
				},
			},
			checkFields: func(t *testing.T, config DLQConfig) {
				assert.Equal(t, 10, config.RetryMaxAttempts)
				assert.Equal(t, "PT5S", config.RetryBackoffDelay)
			},
			description: "Should use custom retry configuration",
		},
		{
			name: "Custom overflow policy",
			lambda: &lambdav1alpha1.LambdaFunction{
				Spec: lambdav1alpha1.LambdaFunctionSpec{
					Eventing: &lambdav1alpha1.EventingSpec{
						DLQ: &lambdav1alpha1.DLQSpec{
							Enabled:        true,
							MaxLength:      10000,
							OverflowPolicy: "drop-head",
						},
					},
				},
			},
			checkFields: func(t *testing.T, config DLQConfig) {
				assert.Equal(t, 10000, config.MaxLength)
				assert.Equal(t, "drop-head", config.OverflowPolicy)
			},
			description: "Should use custom overflow policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.buildDLQConfig(tt.lambda)
			tt.checkFields(t, result)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏗️ SHARED BROKER DATA TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestBuildSharedBrokerData(t *testing.T) {
	m := &Manager{config: DefaultConfig()}

	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-function",
			Namespace: "test-namespace",
		},
	}

	data := m.buildSharedBrokerData(lambda)

	assert.Equal(t, SharedBrokerName, data.Name)
	assert.Equal(t, "test-namespace", data.Namespace)
	assert.Equal(t, "shared", data.LambdaName)
	assert.Equal(t, "rabbitmq", data.RabbitMQ.ClusterName)
}

func TestBuildSharedDLQData(t *testing.T) {
	m := &Manager{config: DefaultConfig()}

	lambda := &lambdav1alpha1.LambdaFunction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-function",
			Namespace: "test-namespace",
		},
	}

	data := m.buildSharedDLQData(lambda)

	assert.Equal(t, SharedDLQPrefix, data.Name)
	assert.Equal(t, "test-namespace", data.Namespace)
	assert.Equal(t, "shared", data.LambdaName)
	assert.False(t, data.Monitoring.Enabled) // Default is disabled
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🤖 AGENT RABBITMQ CONFIG TESTS                                         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestBuildAgentRabbitMQConfig(t *testing.T) {
	m := &Manager{config: DefaultConfig()}

	tests := []struct {
		name           string
		agent          *lambdav1alpha1.LambdaAgent
		expectedConfig RabbitMQConfig
		description    string
	}{
		{
			name: "Default agent RabbitMQ config",
			agent: &lambdav1alpha1.LambdaAgent{
				Spec: lambdav1alpha1.LambdaAgentSpec{},
			},
			expectedConfig: RabbitMQConfig{
				ClusterName: "rabbitmq",
				Namespace:   "rabbitmq-system",
				QueueType:   "quorum",
				Parallelism: 50,
			},
			description: "Should use defaults for agents",
		},
		{
			name: "Custom agent RabbitMQ config",
			agent: &lambdav1alpha1.LambdaAgent{
				Spec: lambdav1alpha1.LambdaAgentSpec{
					Eventing: &lambdav1alpha1.AgentEventingSpec{
						RabbitMQ: &lambdav1alpha1.AgentRabbitMQSpec{
							ClusterName: "agent-rabbitmq",
							Namespace:   "agent-mq-ns",
						},
					},
				},
			},
			expectedConfig: RabbitMQConfig{
				ClusterName: "agent-rabbitmq",
				Namespace:   "agent-mq-ns",
				QueueType:   "quorum",
				Parallelism: 50,
			},
			description: "Should use custom config for agents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.buildAgentRabbitMQConfig(tt.agent)
			assert.Equal(t, tt.expectedConfig.ClusterName, result.ClusterName, tt.description)
			assert.Equal(t, tt.expectedConfig.Namespace, result.Namespace, tt.description)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 MANAGER INITIALIZATION TESTS                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNewManager(t *testing.T) {
	m, err := NewManager(nil, nil)

	require.NoError(t, err)
	require.NotNil(t, m)
	assert.NotNil(t, m.config)
	assert.NotNil(t, m.renderer)
}

func TestNewManagerWithConfig(t *testing.T) {
	customConfig := &Config{
		DefaultRabbitMQCluster:   "custom-cluster",
		DefaultRabbitMQNamespace: "custom-ns",
		DefaultDLQEnabled:        false,
	}

	m, err := NewManagerWithConfig(nil, nil, customConfig)

	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Equal(t, "custom-cluster", m.config.DefaultRabbitMQCluster)
	assert.Equal(t, "custom-ns", m.config.DefaultRabbitMQNamespace)
	assert.False(t, m.config.DefaultDLQEnabled)
}

func TestNewManagerWithConfig_NilConfig(t *testing.T) {
	m, err := NewManagerWithConfig(nil, nil, nil)

	require.NoError(t, err)
	require.NotNil(t, m)
	// Should use defaults
	assert.Equal(t, "rabbitmq", m.config.DefaultRabbitMQCluster)
}
