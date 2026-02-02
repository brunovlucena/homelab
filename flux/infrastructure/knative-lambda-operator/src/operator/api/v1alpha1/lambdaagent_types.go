/*
Copyright 2024 Bruno Lucena.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LambdaAgentSpec defines the desired state of LambdaAgent
type LambdaAgentSpec struct {
	// Image configuration (pre-built Docker image)
	// +kubebuilder:validation:Required
	Image AgentImageSpec `json:"image"`

	// ServiceAccountName for this agent (for K8s API access)
	// The operator will configure the Knative Service to use this SA
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Permission controls (works with K8s RBAC)
	// +optional
	Permissions *AgentPermissionsSpec `json:"permissions,omitempty"`

	// AI/LLM configuration
	// +optional
	AI *AgentAISpec `json:"ai,omitempty"`

	// Agent behavior configuration
	// +optional
	Behavior *AgentBehaviorSpec `json:"behavior,omitempty"`

	// Operation mode: agentic (autonomous) or supervised (requires approval)
	// +kubebuilder:default="agentic"
	// +kubebuilder:validation:Enum=agentic;supervised
	// +optional
	OperationMode string `json:"operationMode,omitempty"`

	// Approval configuration for supervised mode
	// +optional
	Approval *AgentApprovalSpec `json:"approval,omitempty"`

	// Environment variables
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Scaling configuration
	// +optional
	Scaling *AgentScalingSpec `json:"scaling,omitempty"`

	// Resource requirements
	// +optional
	Resources *AgentResourcesSpec `json:"resources,omitempty"`

	// Eventing configuration
	// +optional
	Eventing *AgentEventingSpec `json:"eventing,omitempty"`

	// Observability configuration
	// +optional
	Observability *AgentObservabilitySpec `json:"observability,omitempty"`
}

// AgentImageSpec defines the Docker image configuration
type AgentImageSpec struct {
	// Docker image repository
	// +kubebuilder:validation:Required
	Repository string `json:"repository"`

	// Image tag
	// +kubebuilder:default="latest"
	Tag string `json:"tag,omitempty"`

	// Image digest (takes precedence over tag)
	// +optional
	Digest string `json:"digest,omitempty"`

	// Container port
	// +kubebuilder:default=8080
	Port int32 `json:"port,omitempty"`

	// Image pull policy
	// +kubebuilder:default="IfNotPresent"
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	PullPolicy string `json:"pullPolicy,omitempty"`

	// Image pull secrets
	// +optional
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	// Container command
	// +optional
	Command []string `json:"command,omitempty"`

	// Container arguments
	// +optional
	Args []string `json:"args,omitempty"`
}

// AgentAISpec defines AI/LLM configuration
type AgentAISpec struct {
	// AI provider
	// +kubebuilder:default="ollama"
	// +kubebuilder:validation:Enum=ollama;openai;anthropic;none
	Provider string `json:"provider,omitempty"`

	// AI endpoint URL
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// Model name
	// +optional
	Model string `json:"model,omitempty"`

	// Maximum tokens for response
	// +kubebuilder:default=2048
	MaxTokens int32 `json:"maxTokens,omitempty"`

	// Temperature for generation
	// +kubebuilder:default="0.7"
	Temperature string `json:"temperature,omitempty"`

	// Secret reference for API key
	// +optional
	APIKeySecretRef *corev1.SecretKeySelector `json:"apiKeySecretRef,omitempty"`
}

// AgentBehaviorSpec defines agent behavior
type AgentBehaviorSpec struct {
	// Maximum context messages to maintain
	// +kubebuilder:default=10
	MaxContextMessages int32 `json:"maxContextMessages,omitempty"`

	// Whether to emit CloudEvents
	// +kubebuilder:default=true
	EmitEvents bool `json:"emitEvents,omitempty"`

	// System prompt for the agent
	// +optional
	SystemPrompt string `json:"systemPrompt,omitempty"`
}

// AgentApprovalSpec defines approval configuration for supervised mode
type AgentApprovalSpec struct {
	// Approval provider: slack, custom, or both
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:Enum=slack;custom
	Providers []string `json:"providers"`

	// Slack approval configuration
	// +optional
	Slack *SlackApprovalSpec `json:"slack,omitempty"`

	// Custom app approval configuration
	// +optional
	Custom *CustomApprovalSpec `json:"custom,omitempty"`

	// Timeout for approval requests (default: 1h)
	// +kubebuilder:default="1h"
	Timeout string `json:"timeout,omitempty"`

	// Default action if approval times out (approve, reject, or pending)
	// +kubebuilder:default="pending"
	// +kubebuilder:validation:Enum=approve;reject;pending
	TimeoutAction string `json:"timeoutAction,omitempty"`

	// Require approval from all providers or any provider
	// +kubebuilder:default=false
	RequireAll bool `json:"requireAll,omitempty"`
}

// SlackApprovalSpec defines Slack approval configuration
type SlackApprovalSpec struct {
	// Slack webhook URL for sending approval requests
	// +optional
	WebhookURL string `json:"webhookUrl,omitempty"`

	// Secret reference containing Slack webhook URL
	// +optional
	WebhookURLSecretRef *corev1.SecretKeySelector `json:"webhookUrlSecretRef,omitempty"`

	// Slack bot token for interactive messages
	// +optional
	BotTokenSecretRef *corev1.SecretKeySelector `json:"botTokenSecretRef,omitempty"`

	// Channel to send approval requests to
	// +optional
	Channel string `json:"channel,omitempty"`

	// User IDs or groups that can approve (empty = any user in channel)
	// +optional
	Approvers []string `json:"approvers,omitempty"`

	// Approval callback URL (where Slack sends approval responses)
	// +optional
	CallbackURL string `json:"callbackUrl,omitempty"`
}

// CustomApprovalSpec defines custom app approval configuration
type CustomApprovalSpec struct {
	// Custom app endpoint URL for approval requests
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// HTTP method for approval requests
	// +kubebuilder:default="POST"
	Method string `json:"method,omitempty"`

	// Headers to include in approval requests
	// +optional
	Headers map[string]string `json:"headers,omitempty"`

	// Secret reference for authentication (e.g., API key)
	// +optional
	AuthSecretRef *corev1.SecretKeySelector `json:"authSecretRef,omitempty"`

	// Polling interval for checking approval status (if using polling)
	// +kubebuilder:default="10s"
	PollInterval string `json:"pollInterval,omitempty"`

	// Whether to use webhook callback or polling
	// +kubebuilder:default=true
	UseWebhook bool `json:"useWebhook,omitempty"`

	// Webhook callback URL (where custom app sends approval responses)
	// +optional
	CallbackURL string `json:"callbackUrl,omitempty"`
}

// AgentScalingSpec defines scaling configuration
type AgentScalingSpec struct {
	// Minimum replicas
	// +kubebuilder:default=0
	MinReplicas int32 `json:"minReplicas,omitempty"`

	// Maximum replicas
	// +kubebuilder:default=10
	MaxReplicas int32 `json:"maxReplicas,omitempty"`

	// Target concurrent requests per instance
	// +kubebuilder:default=10
	TargetConcurrency int32 `json:"targetConcurrency,omitempty"`

	// Grace period before scaling to zero
	// +kubebuilder:default="30s"
	ScaleToZeroGracePeriod string `json:"scaleToZeroGracePeriod,omitempty"`
}

// AgentResourcesSpec defines resource requirements
type AgentResourcesSpec struct {
	// Resource requests
	// +optional
	Requests *AgentResourceQuantity `json:"requests,omitempty"`

	// Resource limits
	// +optional
	Limits *AgentResourceQuantity `json:"limits,omitempty"`
}

// AgentResourceQuantity defines CPU and memory for agents
type AgentResourceQuantity struct {
	// CPU
	// +optional
	CPU string `json:"cpu,omitempty"`

	// Memory
	// +optional
	Memory string `json:"memory,omitempty"`
}

// AgentEventingSpec defines eventing configuration
type AgentEventingSpec struct {
	// Enable eventing
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// Event source identifier for emitted events
	// +optional
	EventSource string `json:"eventSource,omitempty"`

	// Intent-based event types this agent EMITS
	// These are the CloudEvent types this agent produces
	// +optional
	Intents []string `json:"intents,omitempty"`

	// Subscriptions define events this agent RECEIVES from other agents
	// The operator will create Triggers for each subscription
	// +optional
	Subscriptions []AgentSubscription `json:"subscriptions,omitempty"`

	// Forwarding rules for cross-namespace event routing
	// The operator will create Channels and Subscriptions for forwarding
	// +optional
	Forwards []AgentForward `json:"forwards,omitempty"`

	// Dead letter queue configuration
	// +optional
	DLQ *AgentDLQSpec `json:"dlq,omitempty"`

	// RabbitMQ configuration for the broker
	// +optional
	RabbitMQ *AgentRabbitMQSpec `json:"rabbitmq,omitempty"`
}

// AgentSubscription defines an event type to receive
type AgentSubscription struct {
	// Event type to subscribe to (CloudEvents type attribute)
	// +kubebuilder:validation:Required
	EventType string `json:"eventType"`

	// Optional filter on source attribute
	// +optional
	Source string `json:"source,omitempty"`

	// Description for documentation
	// +optional
	Description string `json:"description,omitempty"`
}

// AgentForward defines a forwarding rule to another namespace
type AgentForward struct {
	// Event types to forward (CloudEvents type attributes)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	EventTypes []string `json:"eventTypes"`

	// Target agent name in another namespace
	// +kubebuilder:validation:Required
	TargetAgent string `json:"targetAgent"`

	// Target namespace
	// +kubebuilder:validation:Required
	TargetNamespace string `json:"targetNamespace"`

	// Description for documentation
	// +optional
	Description string `json:"description,omitempty"`
}

// AgentDLQSpec defines DLQ configuration
type AgentDLQSpec struct {
	// Enable DLQ
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// Maximum retry attempts
	// +kubebuilder:default=3
	RetryMaxAttempts int32 `json:"retryMaxAttempts,omitempty"`
}

// AgentRabbitMQSpec defines RabbitMQ configuration
type AgentRabbitMQSpec struct {
	// RabbitMQ cluster name
	// +kubebuilder:default="rabbitmq-cluster-knative-lambda"
	ClusterName string `json:"clusterName,omitempty"`

	// RabbitMQ namespace
	// +kubebuilder:default="knative-lambda"
	Namespace string `json:"namespace,omitempty"`
}

// AgentObservabilitySpec defines observability configuration
type AgentObservabilitySpec struct {
	// Tracing configuration
	// +optional
	Tracing *AgentTracingSpec `json:"tracing,omitempty"`

	// Metrics configuration
	// +optional
	Metrics *AgentMetricsSpec `json:"metrics,omitempty"`

	// Logging configuration
	// +optional
	Logging *AgentLoggingSpec `json:"logging,omitempty"`
}

// AgentTracingSpec defines tracing configuration
type AgentTracingSpec struct {
	// Enable tracing
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// OTLP endpoint
	// +kubebuilder:default="alloy.observability.svc:4317"
	Endpoint string `json:"endpoint,omitempty"`
}

// AgentMetricsSpec defines metrics configuration
type AgentMetricsSpec struct {
	// Enable metrics
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// Enable exemplars
	// +kubebuilder:default=true
	Exemplars bool `json:"exemplars,omitempty"`

	// Enable ServiceMonitor creation for Prometheus scraping
	// +kubebuilder:default=true
	ServiceMonitor bool `json:"serviceMonitor,omitempty"`

	// Metrics path
	// +kubebuilder:default="/metrics"
	Path string `json:"path,omitempty"`

	// Scrape interval
	// +kubebuilder:default="30s"
	Interval string `json:"interval,omitempty"`

	// Scrape timeout
	// +kubebuilder:default="10s"
	Timeout string `json:"timeout,omitempty"`
}

// AgentLoggingSpec defines logging configuration
type AgentLoggingSpec struct {
	// Log level
	// +kubebuilder:default="info"
	// +kubebuilder:validation:Enum=debug;info;warn;error
	Level string `json:"level,omitempty"`

	// Log format
	// +kubebuilder:default="json"
	// +kubebuilder:validation:Enum=json;text
	Format string `json:"format,omitempty"`

	// Include trace context
	// +kubebuilder:default=true
	TraceContext bool `json:"traceContext,omitempty"`
}

// AgentPermissionsSpec defines permission controls (works with K8s RBAC)
type AgentPermissionsSpec struct {
	// Disable broker creation even if ServiceAccount has permissions
	// +kubebuilder:default=false
	DisableBrokerCreation bool `json:"disableBrokerCreation,omitempty"`

	// Disable trigger creation even if ServiceAccount has permissions
	// +kubebuilder:default=false
	DisableTriggerCreation bool `json:"disableTriggerCreation,omitempty"`

	// Disable LambdaFunction creation even if ServiceAccount has permissions
	// +kubebuilder:default=false
	DisableFunctionCreation bool `json:"disableFunctionCreation,omitempty"`

	// Namespaces where this agent can create resources (empty = same namespace only)
	// +optional
	AllowedTargetNamespaces []string `json:"allowedTargetNamespaces,omitempty"`

	// Event-based permission control
	// +optional
	EventControls *AgentEventControlsSpec `json:"eventControls,omitempty"`
}

// AgentEventControlsSpec defines event-based permission control
type AgentEventControlsSpec struct {
	// Allow capabilities to be disabled via CloudEvents
	// +kubebuilder:default=false
	AllowEventDisable bool `json:"allowEventDisable,omitempty"`

	// Event types that can control permissions
	// +kubebuilder:default={"io.homelab.rbac.disable","io.homelab.rbac.enable"}
	// +optional
	ControlEventTypes []string `json:"controlEventTypes,omitempty"`

	// Allowed sources for permission control events (empty = any)
	// +optional
	AllowedControlSources []string `json:"allowedControlSources,omitempty"`
}

// LambdaAgentStatus defines the observed state of LambdaAgent
type LambdaAgentStatus struct {
	// Current phase
	// +kubebuilder:validation:Enum=Pending;Deploying;Ready;Failed;Deleting
	Phase LambdaAgentPhase `json:"phase,omitempty"`

	// Conditions
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Service status
	// +optional
	ServiceStatus *AgentServiceStatus `json:"serviceStatus,omitempty"`

	// Eventing status
	// +optional
	EventingStatus *AgentEventingStatus `json:"eventingStatus,omitempty"`

	// AI backend status (ADR-004)
	// +optional
	AIStatus *AgentAIStatus `json:"aiStatus,omitempty"`

	// Permission status
	// +optional
	PermissionStatus *AgentPermissionStatus `json:"permissionStatus,omitempty"`

	// Observed generation
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// AgentPermissionStatus defines the current permission state
type AgentPermissionStatus struct {
	// Whether broker creation is currently disabled
	BrokerCreationDisabled bool `json:"brokerCreationDisabled,omitempty"`

	// Whether trigger creation is currently disabled
	TriggerCreationDisabled bool `json:"triggerCreationDisabled,omitempty"`

	// Whether function creation is currently disabled
	FunctionCreationDisabled bool `json:"functionCreationDisabled,omitempty"`

	// Capabilities disabled via CloudEvents
	// +optional
	DisabledByEvent []DisabledCapability `json:"disabledByEvent,omitempty"`

	// Last time permissions were evaluated
	// +optional
	LastEvaluated *metav1.Time `json:"lastEvaluated,omitempty"`
}

// DisabledCapability tracks a capability disabled via CloudEvent
type DisabledCapability struct {
	// The capability that was disabled
	Capability string `json:"capability,omitempty"`

	// When it was disabled
	DisabledAt metav1.Time `json:"disabledAt,omitempty"`

	// Event source that disabled this capability
	DisabledBy string `json:"disabledBy,omitempty"`

	// When this disable expires (nil = permanent until re-enabled)
	// +optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`
}

// AgentAIStatus defines AI backend status (ADR-004 compliance)
type AgentAIStatus struct {
	// Whether the AI model is loaded and available
	ModelAvailable bool `json:"modelAvailable,omitempty"`

	// Model currently in use
	ActiveModel string `json:"activeModel,omitempty"`

	// AI provider being used
	Provider string `json:"provider,omitempty"`

	// AI endpoint being used
	Endpoint string `json:"endpoint,omitempty"`

	// Last health check timestamp
	// +optional
	LastHealthCheck *metav1.Time `json:"lastHealthCheck,omitempty"`

	// P99 inference latency (from metrics if available)
	// +optional
	InferenceLatencyP99 string `json:"inferenceLatencyP99,omitempty"`

	// Number of active conversations (from metrics if available)
	// +optional
	ActiveConversations int32 `json:"activeConversations,omitempty"`

	// Error message if AI backend is unhealthy
	// +optional
	Error string `json:"error,omitempty"`
}

// AgentEventingStatus defines eventing resource status
type AgentEventingStatus struct {
	// Broker name
	BrokerName string `json:"brokerName,omitempty"`

	// Broker ready status
	BrokerReady bool `json:"brokerReady,omitempty"`

	// Number of active triggers
	TriggerCount int `json:"triggerCount,omitempty"`

	// Number of active forwards (channels)
	ForwardCount int `json:"forwardCount,omitempty"`

	// Broker ingress URL for publishing events
	BrokerURL string `json:"brokerUrl,omitempty"`
}

// LambdaAgentPhase represents the phase of the agent
// +kubebuilder:validation:Enum=Pending;Deploying;Ready;Failed;Deleting
type LambdaAgentPhase string

const (
	AgentPhasePending   LambdaAgentPhase = "Pending"
	AgentPhaseDeploying LambdaAgentPhase = "Deploying"
	AgentPhaseReady     LambdaAgentPhase = "Ready"
	AgentPhaseFailed    LambdaAgentPhase = "Failed"
	AgentPhaseDeleting  LambdaAgentPhase = "Deleting"
)

// AgentServiceStatus defines service status
type AgentServiceStatus struct {
	// Service name
	ServiceName string `json:"serviceName,omitempty"`

	// Service URL
	URL string `json:"url,omitempty"`

	// Ready status
	Ready bool `json:"ready,omitempty"`

	// Latest revision
	LatestRevision string `json:"latestRevision,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.image.repository"
// +kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.serviceStatus.url"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.serviceStatus.ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// LambdaAgent is the Schema for AI agents with pre-built images
type LambdaAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LambdaAgentSpec   `json:"spec,omitempty"`
	Status LambdaAgentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LambdaAgentList contains a list of LambdaAgent
type LambdaAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LambdaAgent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LambdaAgent{}, &LambdaAgentList{})
}

// SetCondition sets a condition on the status
func (s *LambdaAgentStatus) SetCondition(condition metav1.Condition) {
	for i, c := range s.Conditions {
		if c.Type == condition.Type {
			s.Conditions[i] = condition
			return
		}
	}
	s.Conditions = append(s.Conditions, condition)
}

// GetCondition gets a condition by type
func (s *LambdaAgentStatus) GetCondition(conditionType string) *metav1.Condition {
	for _, c := range s.Conditions {
		if c.Type == conditionType {
			return &c
		}
	}
	return nil
}
