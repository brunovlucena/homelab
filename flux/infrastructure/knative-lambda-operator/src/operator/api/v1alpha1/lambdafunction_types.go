package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LambdaFunctionSpec defines the desired state of LambdaFunction
type LambdaFunctionSpec struct {
	// Source configuration for the Lambda function code
	// +kubebuilder:validation:Required
	Source SourceSpec `json:"source"`

	// Runtime configuration
	// +kubebuilder:validation:Required
	Runtime RuntimeSpec `json:"runtime"`

	// Scaling configuration
	// +optional
	Scaling *ScalingSpec `json:"scaling,omitempty"`

	// Resource limits and requests
	// +optional
	Resources *ResourceSpec `json:"resources,omitempty"`

	// Environment variables
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Event triggers configuration (legacy, use Eventing instead)
	// +optional
	Triggers []TriggerSpec `json:"triggers,omitempty"`

	// Build configuration
	// +optional
	Build *BuildSpec `json:"build,omitempty"`

	// Eventing configuration for brokers, triggers, DLQ, and API sources
	// +optional
	Eventing *EventingSpec `json:"eventing,omitempty"`

	// Observability configuration (OTEL tracing, metrics with exemplars, structured logging)
	// +optional
	Observability *ObservabilitySpec `json:"observability,omitempty"`

	// ImagePullPolicy defines when to pull the container image
	// For built images (minio, s3, git, inline sources), defaults to "Always" to avoid stale cache
	// For pre-built images (source.type=image), defaults to "IfNotPresent"
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	// +optional
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
}

// SourceSpec defines the source code location
type SourceSpec struct {
	// Type of source storage
	// +kubebuilder:validation:Enum=minio;s3;gcs;git;inline;image
	// +kubebuilder:validation:Required
	Type string `json:"type"`

	// MinIO source configuration
	// +optional
	MinIO *MinIOSource `json:"minio,omitempty"`

	// S3 source configuration
	// +optional
	S3 *S3Source `json:"s3,omitempty"`

	// GCS source configuration
	// +optional
	GCS *GCSSource `json:"gcs,omitempty"`

	// Git source configuration
	// +optional
	Git *GitSource `json:"git,omitempty"`

	// Inline source configuration
	// +optional
	Inline *InlineSource `json:"inline,omitempty"`

	// Image source configuration (pre-built Docker image, skips build)
	// +optional
	Image *ImageSource `json:"image,omitempty"`
}

// MinIOSource defines MinIO storage configuration
// Security: Endpoint validated to prevent SSRF, bucket/key validated for injection prevention
type MinIOSource struct {
	// MinIO endpoint URL (hostname:port format)
	// Security: Validated to prevent SSRF to internal services
	// +kubebuilder:default="minio.minio.svc.cluster.local:9000"
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9][-a-zA-Z0-9.]*[a-zA-Z0-9](:[0-9]{1,5})?$`
	Endpoint string `json:"endpoint,omitempty"`

	// Bucket name (S3-compatible naming: 3-63 chars, lowercase, dots/hyphens)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`
	Bucket string `json:"bucket"`

	// Object key path (validated to prevent path traversal and injection)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9!_.*'()/-]+$`
	Key string `json:"key"`

	// Secret reference containing access credentials
	// +optional
	SecretRef *corev1.LocalObjectReference `json:"secretRef,omitempty"`
}

// S3Source defines S3 storage configuration
// Security: Bucket/key validated for injection prevention
type S3Source struct {
	// S3 bucket name (3-63 chars, lowercase, dots/hyphens)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`
	Bucket string `json:"bucket"`

	// Object key path (validated to prevent path traversal and injection)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9!_.*'()/-]+$`
	Key string `json:"key"`

	// AWS region (validated format)
	// +kubebuilder:default="us-east-1"
	// +kubebuilder:validation:MaxLength=30
	// +kubebuilder:validation:Pattern=`^[a-z]{2}-[a-z]+-[0-9]+$`
	Region string `json:"region,omitempty"`

	// Secret reference containing AWS credentials
	// +optional
	SecretRef *corev1.LocalObjectReference `json:"secretRef,omitempty"`
}

// GCSSource defines Google Cloud Storage configuration
type GCSSource struct {
	// GCS bucket name
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`

	// Object key path
	// +kubebuilder:validation:Required
	Key string `json:"key"`

	// GCP project ID
	// +optional
	Project string `json:"project,omitempty"`

	// Secret reference containing GCP credentials
	// +optional
	SecretRef *corev1.LocalObjectReference `json:"secretRef,omitempty"`
}

// GitSource defines Git repository configuration
// Security: All fields validated to prevent SSRF (BLUE-001), path traversal (BLUE-005), and injection
type GitSource struct {
	// Git repository URL
	// Security: Must be HTTPS or internal cluster URL (SSRF prevention)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	// +kubebuilder:validation:Pattern=`^(https://|git://|git@)[a-zA-Z0-9][-a-zA-Z0-9._~:/?#\[\]@!$&'()*+,;=%]+$`
	URL string `json:"url"`

	// Git reference (branch, tag, or commit)
	// Security: Only alphanumeric, dash, underscore, dot, forward slash allowed
	// +kubebuilder:default="main"
	// +kubebuilder:validation:MaxLength=256
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9][a-zA-Z0-9._/-]*$`
	Ref string `json:"ref,omitempty"`

	// Path within the repository
	// Security: Cannot contain .. for path traversal prevention
	// +optional
	// +kubebuilder:validation:MaxLength=512
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9][a-zA-Z0-9._/-]*$`
	Path string `json:"path,omitempty"`

	// Secret reference containing Git credentials
	// +optional
	SecretRef *corev1.LocalObjectReference `json:"secretRef,omitempty"`
}

// InlineSource defines inline code configuration
type InlineSource struct {
	// Source code content
	// +kubebuilder:validation:Required
	Code string `json:"code"`

	// Dependencies (e.g., requirements.txt, package.json)
	// +optional
	Dependencies string `json:"dependencies,omitempty"`
}

// ImageSource defines pre-built Docker image configuration (skips build pipeline)
// Use this for FastAPI apps, existing containers, or any pre-built image
type ImageSource struct {
	// Repository is the Docker image repository (e.g., localhost:5001/my-app)
	// +kubebuilder:validation:Required
	Repository string `json:"repository"`

	// Tag is the image tag (defaults to "latest")
	// +kubebuilder:default="latest"
	Tag string `json:"tag,omitempty"`

	// Digest is the image digest (overrides tag if specified)
	// +optional
	Digest string `json:"digest,omitempty"`

	// PullPolicy defines when to pull the image
	// +kubebuilder:default="IfNotPresent"
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	PullPolicy string `json:"pullPolicy,omitempty"`

	// ImagePullSecrets for private registries
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// Port the container listens on (defaults to 8080)
	// +kubebuilder:default=8080
	Port int32 `json:"port,omitempty"`

	// Command overrides the container entrypoint
	// +optional
	Command []string `json:"command,omitempty"`

	// Args overrides the container arguments
	// +optional
	Args []string `json:"args,omitempty"`
}

// RuntimeSpec defines the runtime configuration
type RuntimeSpec struct {
	// Programming language
	// +kubebuilder:validation:Enum=nodejs;python;go
	// +kubebuilder:validation:Required
	Language string `json:"language"`

	// Language version
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=20
	// +kubebuilder:validation:Pattern=`^[0-9]+(\.[0-9]+)*(-[a-zA-Z0-9]+)?$`
	Version string `json:"version"`

	// Handler function name in format module.function (e.g., main.handler, index.process)
	// Security: Validated to prevent template injection (BLUE-002)
	// Only alphanumeric and underscore allowed, must have exactly one dot
	// +kubebuilder:default="index.handler"
	// +kubebuilder:validation:MaxLength=100
	// +kubebuilder:validation:Pattern=`^[a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z_][a-zA-Z0-9_]*$`
	Handler string `json:"handler,omitempty"`
}

// ScalingSpec defines autoscaling configuration
type ScalingSpec struct {
	// Minimum number of replicas
	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// Maximum number of replicas
	// +kubebuilder:default=50
	// +kubebuilder:validation:Minimum=1
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`

	// ContainerConcurrency is the HARD LIMIT of concurrent requests per pod
	// This is the maximum number of requests a single pod can handle simultaneously
	// For I/O-bound workloads, set higher (10-100). For CPU-bound, set lower (1-5)
	// +kubebuilder:default=10
	// +kubebuilder:validation:Minimum=1
	ContainerConcurrency *int32 `json:"containerConcurrency,omitempty"`

	// TargetConcurrency is the autoscaling target (when to add more pods)
	// Scale up when average concurrency reaches this value
	// Set lower than ContainerConcurrency for headroom (e.g., 30% of ContainerConcurrency)
	// +kubebuilder:default=5
	// +kubebuilder:validation:Minimum=1
	TargetConcurrency *int32 `json:"targetConcurrency,omitempty"`

	// Grace period before scaling to zero
	// +kubebuilder:default="30s"
	ScaleToZeroGracePeriod string `json:"scaleToZeroGracePeriod,omitempty"`
}

// ResourceSpec defines resource limits and requests
type ResourceSpec struct {
	// Resource requests
	// +optional
	Requests *ResourceRequirements `json:"requests,omitempty"`

	// Resource limits
	// +optional
	Limits *ResourceRequirements `json:"limits,omitempty"`
}

// ResourceRequirements defines CPU and memory requirements
type ResourceRequirements struct {
	// Memory requirement
	// +kubebuilder:default="64Mi"
	Memory string `json:"memory,omitempty"`

	// CPU requirement
	// +kubebuilder:default="50m"
	CPU string `json:"cpu,omitempty"`
}

// TriggerSpec defines event trigger configuration (legacy)
type TriggerSpec struct {
	// Name of the Knative Broker
	// +kubebuilder:validation:Required
	Broker string `json:"broker"`

	// Event filter attributes
	// +optional
	Filter map[string]string `json:"filter,omitempty"`
}

// EventingSpec defines eventing infrastructure configuration
type EventingSpec struct {
	// Enable eventing infrastructure (brokers, triggers, DLQ)
	// Eventing is ENABLED by default - lambdas receive CloudEvents
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// Resource prefix for all eventing resources (defaults to lambda name)
	// +optional
	ResourcePrefix string `json:"resourcePrefix,omitempty"`

	// Name of an existing broker to use (skips broker creation)
	// +optional
	BrokerName string `json:"brokerName,omitempty"`

	// Name of the subscriber Knative Service (defaults to lambda name)
	// +optional
	SubscriberServiceName string `json:"subscriberServiceName,omitempty"`

	// Event source identifier
	// +kubebuilder:default="knative-lambda-operator"
	EventSource string `json:"eventSource,omitempty"`

	// RabbitMQ cluster configuration
	// +optional
	RabbitMQ *RabbitMQSpec `json:"rabbitmq,omitempty"`

	// Dead Letter Queue configuration
	// +optional
	DLQ *DLQSpec `json:"dlq,omitempty"`

	// Custom event types
	// +optional
	EventTypes *EventTypesSpec `json:"eventTypes,omitempty"`

	// ApiServerSource configuration
	// +optional
	ApiSource *ApiSourceSpec `json:"apiSource,omitempty"`

	// Monitoring configuration
	// +optional
	Monitoring *MonitoringSpec `json:"monitoring,omitempty"`
}

// RabbitMQSpec defines RabbitMQ cluster configuration
type RabbitMQSpec struct {
	// Name of the RabbitMQ cluster
	// +kubebuilder:default="rabbitmq"
	ClusterName string `json:"clusterName,omitempty"`

	// Namespace of the RabbitMQ cluster
	// +kubebuilder:default="rabbitmq-system"
	Namespace string `json:"namespace,omitempty"`

	// Queue type (classic or quorum)
	// +kubebuilder:validation:Enum=classic;quorum
	// +kubebuilder:default="quorum"
	QueueType string `json:"queueType,omitempty"`

	// Parallelism for event processing (concurrent event deliveries per trigger dispatcher)
	// Higher values = more concurrent deliveries = better autoscaling
	// +kubebuilder:default=50
	// +kubebuilder:validation:Minimum=1
	Parallelism int `json:"parallelism,omitempty"`

	// PrefetchCount controls how many messages RabbitMQ sends to a consumer before waiting for acks
	// Higher values = better throughput, lower values = more even distribution
	// +kubebuilder:default=100
	// +kubebuilder:validation:Minimum=1
	PrefetchCount int `json:"prefetchCount,omitempty"`
}

// DLQSpec defines Dead Letter Queue configuration
type DLQSpec struct {
	// Enable DLQ
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// DLQ exchange name
	// +kubebuilder:default="lambda-dlq-exchange"
	ExchangeName string `json:"exchangeName,omitempty"`

	// DLQ queue name
	// +kubebuilder:default="lambda-dlq-queue"
	QueueName string `json:"queueName,omitempty"`

	// Routing key prefix for DLQ
	// +kubebuilder:default="lambda.dlq"
	RoutingKeyPrefix string `json:"routingKeyPrefix,omitempty"`

	// Maximum retry attempts before sending to DLQ
	// +kubebuilder:default=5
	// +kubebuilder:validation:Minimum=1
	RetryMaxAttempts int `json:"retryMaxAttempts,omitempty"`

	// Retry backoff delay (ISO 8601 duration)
	// +kubebuilder:default="PT1S"
	RetryBackoffDelay string `json:"retryBackoffDelay,omitempty"`

	// Message TTL in milliseconds (default 7 days)
	// +kubebuilder:default=604800000
	MessageTTL int `json:"messageTTL,omitempty"`

	// Maximum number of messages in DLQ
	// +kubebuilder:default=50000
	MaxLength int `json:"maxLength,omitempty"`

	// Overflow policy when queue is full
	// +kubebuilder:validation:Enum=drop-head;reject-publish
	// +kubebuilder:default="reject-publish"
	OverflowPolicy string `json:"overflowPolicy,omitempty"`

	// DLQ cleanup configuration
	// +optional
	Cleanup *DLQCleanupSpec `json:"cleanup,omitempty"`
}

// DLQCleanupSpec defines DLQ cleanup configuration
type DLQCleanupSpec struct {
	// Enable automatic cleanup
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// Cleanup interval
	// +kubebuilder:default="1h"
	Interval string `json:"interval,omitempty"`

	// Retention period
	// +kubebuilder:default="168h"
	Retention string `json:"retention,omitempty"`
}

// EventTypesSpec defines custom event type mappings
// CloudEvents format: io.knative.lambda.<category>.<entity>.<action>
// Categories: command (present tense), lifecycle (past tense), invoke, response
type EventTypesSpec struct {
	// ┌─────────────────────────────────────────────────────────────────────┐
	// │  📤 COMMAND EVENTS - Requests for actions (present tense)          │
	// └─────────────────────────────────────────────────────────────────────┘

	// Command event type for build start request
	// +kubebuilder:default="io.knative.lambda.command.build.start"
	CommandBuildStart string `json:"commandBuildStart,omitempty"`

	// Command event type for build cancel request
	// +kubebuilder:default="io.knative.lambda.command.build.cancel"
	CommandBuildCancel string `json:"commandBuildCancel,omitempty"`

	// Command event type for build retry request
	// +kubebuilder:default="io.knative.lambda.command.build.retry"
	CommandBuildRetry string `json:"commandBuildRetry,omitempty"`

	// Command event type for service create request
	// +kubebuilder:default="io.knative.lambda.command.service.create"
	CommandServiceCreate string `json:"commandServiceCreate,omitempty"`

	// Command event type for service delete request
	// +kubebuilder:default="io.knative.lambda.command.service.delete"
	CommandServiceDelete string `json:"commandServiceDelete,omitempty"`

	// ┌─────────────────────────────────────────────────────────────────────┐
	// │  📊 LIFECYCLE EVENTS - State changes (past tense)                  │
	// └─────────────────────────────────────────────────────────────────────┘

	// Lifecycle event type for build started
	// +kubebuilder:default="io.knative.lambda.lifecycle.build.started"
	LifecycleBuildStarted string `json:"lifecycleBuildStarted,omitempty"`

	// Lifecycle event type for build completed
	// +kubebuilder:default="io.knative.lambda.lifecycle.build.completed"
	LifecycleBuildCompleted string `json:"lifecycleBuildCompleted,omitempty"`

	// Lifecycle event type for build failed
	// +kubebuilder:default="io.knative.lambda.lifecycle.build.failed"
	LifecycleBuildFailed string `json:"lifecycleBuildFailed,omitempty"`

	// Lifecycle event type for build timeout
	// +kubebuilder:default="io.knative.lambda.lifecycle.build.timeout"
	LifecycleBuildTimeout string `json:"lifecycleBuildTimeout,omitempty"`

	// Lifecycle event type for build cancelled
	// +kubebuilder:default="io.knative.lambda.lifecycle.build.cancelled"
	LifecycleBuildCancelled string `json:"lifecycleBuildCancelled,omitempty"`

	// ┌─────────────────────────────────────────────────────────────────────┐
	// │  🚀 INVOKE EVENTS - Trigger lambda execution                       │
	// └─────────────────────────────────────────────────────────────────────┘

	// Invoke event type for synchronous invocation
	// +kubebuilder:default="io.knative.lambda.invoke.sync"
	InvokeSync string `json:"invokeSync,omitempty"`

	// Invoke event type for asynchronous invocation
	// +kubebuilder:default="io.knative.lambda.invoke.async"
	InvokeAsync string `json:"invokeAsync,omitempty"`

	// Invoke event type for scheduled invocation
	// +kubebuilder:default="io.knative.lambda.invoke.scheduled"
	InvokeScheduled string `json:"invokeScheduled,omitempty"`

	// ┌─────────────────────────────────────────────────────────────────────┐
	// │  📨 RESPONSE EVENTS - Lambda execution results                     │
	// └─────────────────────────────────────────────────────────────────────┘

	// Response event type for successful execution
	// +kubebuilder:default="io.knative.lambda.response.success"
	ResponseSuccess string `json:"responseSuccess,omitempty"`

	// Response event type for execution error
	// +kubebuilder:default="io.knative.lambda.response.error"
	ResponseError string `json:"responseError,omitempty"`

	// Response event type for execution timeout
	// +kubebuilder:default="io.knative.lambda.response.timeout"
	ResponseTimeout string `json:"responseTimeout,omitempty"`

	// ┌─────────────────────────────────────────────────────────────────────┐
	// │  🔧 LEGACY FIELDS - For backward compatibility                     │
	// └─────────────────────────────────────────────────────────────────────┘

	// Deprecated: Use CommandBuildStart. Legacy event type for build start
	// +kubebuilder:default="io.knative.lambda.command.build.start"
	BuildStart string `json:"buildStart,omitempty"`

	// Deprecated: Use LifecycleBuildCompleted. Legacy event type for build completion
	// +kubebuilder:default="io.knative.lambda.lifecycle.build.completed"
	BuildComplete string `json:"buildComplete,omitempty"`

	// Deprecated: Use LifecycleBuildFailed. Legacy event type for build failure
	// +kubebuilder:default="io.knative.lambda.lifecycle.build.failed"
	BuildFailed string `json:"buildFailed,omitempty"`

	// Deprecated: Use LifecycleBuildTimeout. Legacy event type for build timeout
	// +kubebuilder:default="io.knative.lambda.lifecycle.build.timeout"
	BuildTimeout string `json:"buildTimeout,omitempty"`

	// Deprecated: Use CommandServiceDelete. Legacy event type for service deletion
	// +kubebuilder:default="io.knative.lambda.command.service.delete"
	ServiceDelete string `json:"serviceDelete,omitempty"`
}

// ApiSourceSpec defines ApiServerSource configuration
type ApiSourceSpec struct {
	// Enable ApiServerSource for watching Kubernetes resources
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// Watch mode (Resource or Reference)
	// +kubebuilder:validation:Enum=Resource;Reference
	// +kubebuilder:default="Resource"
	Mode string `json:"mode,omitempty"`

	// Resources to watch
	// +optional
	Resources []ApiSourceResourceSpec `json:"resources,omitempty"`
}

// ApiSourceResourceSpec defines a Kubernetes resource to watch
type ApiSourceResourceSpec struct {
	// API version of the resource
	// +kubebuilder:validation:Required
	APIVersion string `json:"apiVersion"`

	// Kind of the resource
	// +kubebuilder:validation:Required
	Kind string `json:"kind"`

	// Label selector to filter resources
	// +optional
	LabelSelector map[string]string `json:"labelSelector,omitempty"`
}

// MonitoringSpec defines monitoring configuration
type MonitoringSpec struct {
	// Enable ServiceMonitor creation
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`
}

// ObservabilitySpec defines observability configuration (OTEL, tracing, metrics, logging)
type ObservabilitySpec struct {
	// Tracing configuration (OpenTelemetry distributed tracing)
	// +optional
	Tracing *TracingSpec `json:"tracing,omitempty"`

	// Metrics configuration (Prometheus with exemplars)
	// +optional
	Metrics *MetricsSpec `json:"metrics,omitempty"`

	// Logging configuration (structured JSON with trace context)
	// +optional
	Logging *LoggingSpec `json:"logging,omitempty"`

	// Logfire configuration (Pydantic Logfire integration)
	// +optional
	Logfire *LogfireSpec `json:"logfire,omitempty"`
}

// TracingSpec defines distributed tracing configuration
type TracingSpec struct {
	// Enable distributed tracing
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// Sampling rate (0.0 - 1.0, default: 1.0 for all traces)
	// +kubebuilder:default="1.0"
	// +kubebuilder:validation:Pattern=`^(0(\.\d+)?|1(\.0+)?)$`
	SamplingRate string `json:"samplingRate,omitempty"`

	// OTLP endpoint for trace export
	// +kubebuilder:default="alloy.observability.svc:4317"
	Endpoint string `json:"endpoint,omitempty"`

	// Trace context propagation format
	// +kubebuilder:validation:Enum=w3c;b3;jaeger
	// +kubebuilder:default="w3c"
	Propagation string `json:"propagation,omitempty"`

	// Additional resource attributes for traces
	// +optional
	ResourceAttributes map[string]string `json:"resourceAttributes,omitempty"`
}

// MetricsSpec defines metrics configuration with exemplar support
type MetricsSpec struct {
	// Enable metrics collection
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// Enable exemplars (link metrics to traces)
	// +kubebuilder:default=true
	Exemplars bool `json:"exemplars,omitempty"`

	// Custom metric labels (low cardinality only!)
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Metrics endpoint path
	// +kubebuilder:default="/metrics"
	Path string `json:"path,omitempty"`

	// Metrics port
	// +kubebuilder:default=9090
	Port int32 `json:"port,omitempty"`
}

// LoggingSpec defines structured logging configuration
type LoggingSpec struct {
	// Log level
	// +kubebuilder:validation:Enum=debug;info;warn;error
	// +kubebuilder:default="info"
	Level string `json:"level,omitempty"`

	// Include trace context in logs (trace_id, span_id)
	// +kubebuilder:default=true
	TraceContext bool `json:"traceContext,omitempty"`

	// Log output format
	// +kubebuilder:validation:Enum=json;text
	// +kubebuilder:default="json"
	Format string `json:"format,omitempty"`

	// Include CloudEvent metadata in logs
	// +kubebuilder:default=true
	CloudEventMetadata bool `json:"cloudEventMetadata,omitempty"`
}

// LogfireSpec defines Pydantic Logfire integration
type LogfireSpec struct {
	// Enable Logfire integration
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// Secret reference containing Logfire token
	// +optional
	TokenSecretRef *corev1.SecretKeySelector `json:"tokenSecretRef,omitempty"`

	// Send directly to Logfire (bypasses Alloy)
	// +kubebuilder:default=false
	DirectExport bool `json:"directExport,omitempty"`
}

// BuildSpec defines build configuration
type BuildSpec struct {
	// Build timeout duration
	// +kubebuilder:default="30m"
	Timeout string `json:"timeout,omitempty"`

	// Container registry URL for built images
	// Examples: localhost:5001, 123456789.dkr.ecr.us-west-2.amazonaws.com, gcr.io/my-project
	// +optional
	Registry string `json:"registry,omitempty"`

	// Registry type for authentication configuration
	// +kubebuilder:validation:Enum=local;ecr;gcr;ghcr;dockerhub;generic
	// +kubebuilder:default=local
	// +optional
	RegistryType string `json:"registryType,omitempty"`

	// Repository name within the registry (overrides default namespace/name pattern)
	// +optional
	Repository string `json:"repository,omitempty"`

	// Image tag (defaults to ResourceVersion)
	// +optional
	Tag string `json:"tag,omitempty"`

	// Kubernetes secret name containing registry credentials
	// For ECR: not needed if using IRSA/Pod Identity
	// For GCR: secret with key.json
	// For DockerHub/GHCR: secret with .dockerconfigjson
	// +optional
	ImagePullSecret string `json:"imagePullSecret,omitempty"`

	// AWS region for ECR (required when registryType=ecr)
	// +optional
	AWSRegion string `json:"awsRegion,omitempty"`

	// Use insecure registry (no TLS verification)
	// +kubebuilder:default=false
	// +optional
	Insecure bool `json:"insecure,omitempty"`

	// Force rebuild even if image exists
	// +kubebuilder:default=false
	ForceRebuild bool `json:"forceRebuild,omitempty"`
}

// LambdaFunctionStatus defines the observed state of LambdaFunction
type LambdaFunctionStatus struct {
	// Current phase of the LambdaFunction
	// +kubebuilder:validation:Enum=Pending;Building;Deploying;Ready;Failed;Deleting
	Phase LambdaPhase `json:"phase,omitempty"`

	// Build status information
	// +optional
	BuildStatus *BuildStatusInfo `json:"buildStatus,omitempty"`

	// Service status information
	// +optional
	ServiceStatus *ServiceStatusInfo `json:"serviceStatus,omitempty"`

	// Standard Kubernetes conditions
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the most recent generation observed
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// LambdaPhase represents the current phase of the LambdaFunction
// +kubebuilder:validation:Enum=Pending;Building;Deploying;Ready;Failed;Deleting
type LambdaPhase string

const (
	// PhasePending indicates the function is pending processing
	PhasePending LambdaPhase = "Pending"
	// PhaseBuilding indicates the function is being built
	PhaseBuilding LambdaPhase = "Building"
	// PhaseDeploying indicates the function is being deployed
	PhaseDeploying LambdaPhase = "Deploying"
	// PhaseReady indicates the function is ready
	PhaseReady LambdaPhase = "Ready"
	// PhaseFailed indicates the function has failed
	PhaseFailed LambdaPhase = "Failed"
	// PhaseDeleting indicates the function is being deleted
	PhaseDeleting LambdaPhase = "Deleting"
)

// BuildStatusInfo represents build state
type BuildStatusInfo struct {
	// Name of the build Job
	JobName string `json:"jobName,omitempty"`

	// URI of the built container image
	ImageURI string `json:"imageURI,omitempty"`

	// Time when the build started
	StartedAt *metav1.Time `json:"startedAt,omitempty"`

	// Time when the build completed
	CompletedAt *metav1.Time `json:"completedAt,omitempty"`

	// Error message if build failed
	Error string `json:"error,omitempty"`

	// Build attempt number
	Attempt int32 `json:"attempt,omitempty"`
}

// ServiceStatusInfo represents Knative Service state
type ServiceStatusInfo struct {
	// Name of the Knative Service
	ServiceName string `json:"serviceName,omitempty"`

	// External URL of the service
	URL string `json:"url,omitempty"`

	// Whether the service is ready
	Ready bool `json:"ready,omitempty"`

	// Current number of replicas
	Replicas int32 `json:"replicas,omitempty"`

	// Latest revision name
	LatestRevision string `json:"latestRevision,omitempty"`
}

// Condition types for LambdaFunction
const (
	// ConditionSourceReady indicates source code is available
	ConditionSourceReady = "SourceReady"
	// ConditionBuildReady indicates build is complete
	ConditionBuildReady = "BuildReady"
	// ConditionEventingReady indicates eventing infrastructure (Broker/Trigger) is ready
	ConditionEventingReady = "EventingReady"
	// ConditionDeployReady indicates deployment is complete
	ConditionDeployReady = "DeployReady"
	// ConditionServiceReady indicates service is ready
	ConditionServiceReady = "ServiceReady"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=lf;lfunc
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Current phase"
//+kubebuilder:printcolumn:name="Image",type="string",JSONPath=".status.buildStatus.imageURI",description="Built image URI"
//+kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.serviceStatus.url",description="Service URL"
//+kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.serviceStatus.ready",description="Is service ready"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// LambdaFunction is the Schema for the lambdafunctions API
type LambdaFunction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LambdaFunctionSpec   `json:"spec,omitempty"`
	Status LambdaFunctionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LambdaFunctionList contains a list of LambdaFunction
type LambdaFunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LambdaFunction `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LambdaFunction{}, &LambdaFunctionList{})
}

// GetCondition returns the condition with the given type
func (s *LambdaFunctionStatus) GetCondition(condType string) *metav1.Condition {
	for i := range s.Conditions {
		if s.Conditions[i].Type == condType {
			return &s.Conditions[i]
		}
	}
	return nil
}

// SetCondition sets or updates a condition
func (s *LambdaFunctionStatus) SetCondition(condition metav1.Condition) {
	for i := range s.Conditions {
		if s.Conditions[i].Type == condition.Type {
			s.Conditions[i] = condition
			return
		}
	}
	s.Conditions = append(s.Conditions, condition)
}
