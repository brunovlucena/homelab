package eventing

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//  📦 EVENTING CONFIGURATION TYPES
//
//  These types are used to render Knative Eventing templates for:
//  - Brokers (RabbitMQ-backed)
//  - Triggers (CloudEvents routing)
//  - Dead Letter Queues (DLQ)
//  - ApiServerSource (K8s resource watching)
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🐇 RABBITMQ CONFIGURATION                                              │
// └─────────────────────────────────────────────────────────────────────────┘

// RabbitMQConfig holds RabbitMQ cluster configuration
type RabbitMQConfig struct {
	ClusterName   string
	Namespace     string
	QueueType     string
	Parallelism   int // Concurrent event deliveries per trigger dispatcher
	PrefetchCount int // Messages prefetched from RabbitMQ before acks
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ☠️ DEAD LETTER QUEUE CONFIGURATION                                     │
// └─────────────────────────────────────────────────────────────────────────┘

// DLQConfig holds Dead Letter Queue configuration
type DLQConfig struct {
	Enabled                bool
	ExchangeName           string
	QueueName              string
	RoutingKeyPrefix       string
	RetryPolicy            string
	RetryMaxAttempts       int
	RetryBackoffDelay      string
	RetryBackoffMultiplier int
	BackoffPolicy          string
	MessageTTL             int
	MaxLength              int
	OverflowPolicy         string
	AlertThreshold         int
	DepthThreshold         int
	AgeThreshold           string
	Cleanup                CleanupConfig
}

// CleanupConfig holds DLQ cleanup configuration
type CleanupConfig struct {
	Enabled   bool
	Interval  string
	Retention string
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 MONITORING CONFIGURATION                                            │
// └─────────────────────────────────────────────────────────────────────────┘

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Enabled bool
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🌐 EVENT TYPES CONFIGURATION                                           │
// │                                                                          │
// │  CloudEvents type mapping following the specification:                   │
// │  Format: io.knative.lambda.<category>.<entity>.<action>                 │
// │                                                                          │
// │  Categories:                                                             │
// │  - command: Requests for actions (present tense)                        │
// │  - lifecycle: State changes (past tense)                                │
// │  - invoke: Lambda invocations                                           │
// │  - response: Lambda responses                                           │
// │  - notification: Alerts and audit                                       │
// └─────────────────────────────────────────────────────────────────────────┘

// Default event types following CloudEvents specification
const (
	// Command events (present tense - requests)
	DefaultEventTypeCommandBuildStart    = "io.knative.lambda.command.build.start"
	DefaultEventTypeCommandBuildCancel   = "io.knative.lambda.command.build.cancel"
	DefaultEventTypeCommandBuildRetry    = "io.knative.lambda.command.build.retry"
	DefaultEventTypeCommandServiceCreate = "io.knative.lambda.command.service.create"
	DefaultEventTypeCommandServiceDelete = "io.knative.lambda.command.service.delete"

	// Lifecycle events (past tense - state changes)
	DefaultEventTypeLifecycleBuildStarted   = "io.knative.lambda.lifecycle.build.started"
	DefaultEventTypeLifecycleBuildCompleted = "io.knative.lambda.lifecycle.build.completed"
	DefaultEventTypeLifecycleBuildFailed    = "io.knative.lambda.lifecycle.build.failed"
	DefaultEventTypeLifecycleBuildTimeout   = "io.knative.lambda.lifecycle.build.timeout"
	DefaultEventTypeLifecycleBuildCancelled = "io.knative.lambda.lifecycle.build.cancelled"

	// Invoke events
	DefaultEventTypeInvokeSync  = "io.knative.lambda.invoke.sync"
	DefaultEventTypeInvokeAsync = "io.knative.lambda.invoke.async"

	// Response events
	DefaultEventTypeResponseSuccess = "io.knative.lambda.response.success"
	DefaultEventTypeResponseError   = "io.knative.lambda.response.error"
	DefaultEventTypeResponseTimeout = "io.knative.lambda.response.timeout"
)

// EventTypesConfig holds event type mappings for triggers
type EventTypesConfig struct {
	// Command events (triggers that START actions)
	CommandBuildStart    string
	CommandBuildCancel   string
	CommandBuildRetry    string
	CommandServiceCreate string
	CommandServiceDelete string

	// Lifecycle events (triggers for state change NOTIFICATIONS)
	LifecycleBuildStarted   string
	LifecycleBuildCompleted string
	LifecycleBuildFailed    string
	LifecycleBuildTimeout   string
	LifecycleBuildCancelled string

	// Invoke events (triggers to INVOKE lambdas)
	InvokeSync  string
	InvokeAsync string

	// Response events (triggers for lambda RESPONSES)
	ResponseSuccess string
	ResponseError   string
	ResponseTimeout string

	// Legacy fields (deprecated - for backward compatibility)
	// Deprecated: Use CommandBuildStart
	BuildStart string
	// Deprecated: Use LifecycleBuildCompleted
	BuildComplete string
	// Deprecated: Use LifecycleBuildFailed
	BuildFailed string
	// Deprecated: Use LifecycleBuildTimeout
	BuildTimeout string
	// Deprecated: Use CommandServiceDelete
	ServiceDelete string
}

// NewDefaultEventTypesConfig returns EventTypesConfig with default values
func NewDefaultEventTypesConfig() EventTypesConfig {
	return EventTypesConfig{
		// Command events
		CommandBuildStart:    DefaultEventTypeCommandBuildStart,
		CommandBuildCancel:   DefaultEventTypeCommandBuildCancel,
		CommandBuildRetry:    DefaultEventTypeCommandBuildRetry,
		CommandServiceCreate: DefaultEventTypeCommandServiceCreate,
		CommandServiceDelete: DefaultEventTypeCommandServiceDelete,

		// Lifecycle events
		LifecycleBuildStarted:   DefaultEventTypeLifecycleBuildStarted,
		LifecycleBuildCompleted: DefaultEventTypeLifecycleBuildCompleted,
		LifecycleBuildFailed:    DefaultEventTypeLifecycleBuildFailed,
		LifecycleBuildTimeout:   DefaultEventTypeLifecycleBuildTimeout,
		LifecycleBuildCancelled: DefaultEventTypeLifecycleBuildCancelled,

		// Invoke events
		InvokeSync:  DefaultEventTypeInvokeSync,
		InvokeAsync: DefaultEventTypeInvokeAsync,

		// Response events
		ResponseSuccess: DefaultEventTypeResponseSuccess,
		ResponseError:   DefaultEventTypeResponseError,
		ResponseTimeout: DefaultEventTypeResponseTimeout,

		// Legacy (backward compatibility)
		BuildStart:    DefaultEventTypeCommandBuildStart,
		BuildComplete: DefaultEventTypeLifecycleBuildCompleted,
		BuildFailed:   DefaultEventTypeLifecycleBuildFailed,
		BuildTimeout:  DefaultEventTypeLifecycleBuildTimeout,
		ServiceDelete: DefaultEventTypeCommandServiceDelete,
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔌 API SERVER SOURCE CONFIGURATION                                     │
// └─────────────────────────────────────────────────────────────────────────┘

// ApiSourceConfig holds ApiServerSource configuration
type ApiSourceConfig struct {
	Enabled   bool
	Mode      string
	Resources []ApiSourceResource
}

// ApiSourceResource defines a resource to watch
type ApiSourceResource struct {
	APIVersion    string
	Kind          string
	LabelSelector map[string]string
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📄 TEMPLATE DATA STRUCTURES                                            │
// └─────────────────────────────────────────────────────────────────────────┘

// BrokerData holds data for rendering broker templates
type BrokerData struct {
	Name       string
	Namespace  string
	LambdaName string
	RabbitMQ   RabbitMQConfig
	DLQ        DLQConfig
}

// TriggerData holds data for rendering trigger templates
type TriggerData struct {
	Name                  string
	Namespace             string
	LambdaName            string
	BrokerName            string
	SubscriberServiceName string
	EventSource           string
	RabbitMQ              RabbitMQConfig
	EventTypes            EventTypesConfig
}

// DLQData holds data for rendering DLQ templates
type DLQData struct {
	Name       string
	Namespace  string
	LambdaName string
	RabbitMQ   RabbitMQConfig
	DLQ        DLQConfig
	Monitoring MonitoringConfig
}

// RBACData holds data for rendering RBAC templates
type RBACData struct {
	Name       string
	Namespace  string
	LambdaName string
}

// ApiSourceData holds data for rendering ApiServerSource templates
type ApiSourceData struct {
	Name                  string
	Namespace             string
	LambdaName            string
	SubscriberServiceName string
	ApiSource             ApiSourceConfig
}
