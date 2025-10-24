// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🔧 CONSTANTS - Application-wide immutable constants
//
//	🎯 Purpose: True constants that never change - error messages, validation patterns, etc.
//	💡 Features: Error messages, validation patterns, immutable configuration values
//
//	🏛️ ARCHITECTURE:
//	❌ Error Constants - Error message templates
//	🏷️ Label Constants - Kubernetes and application labels
//	📏 Limit Constants - Hard limits and validation rules
//	⏱️ Timeout Constants - Fixed timeout values
//	🔒 Security Constants - Security-related immutable values
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package constants

import "time"

// AWS Account ID length
const AWSAccountIDLength = 12

// Kubernetes component constants
const (
	Kubernetes    = "kubernetes"
	Kubeconfig    = "kubeconfig"
	Config        = "config"
	Client        = "client"
	DynamicClient = "dynamic_client"
)

// Error constants
const (
	ErrJobIsNil                        = "job is nil"
	ErrJobNameEmpty                    = "job name is empty"
	ErrJobNamespaceEmpty               = "job namespace is empty"
	ErrNameEmpty                       = "name is empty"
	ErrNameTooLong                     = "name is too long"
	ErrNameInvalidStart                = "name starts with invalid character"
	ErrNameInvalidEnd                  = "name ends with invalid character"
	ErrNameInvalidCharacters           = "name contains invalid characters"
	ErrLabelKeyEmpty                   = "label key is empty"
	ErrLabelKeyTooLong                 = "label key is too long"
	ErrLabelKeyInvalidStart            = "label key starts with invalid character"
	ErrLabelKeyInvalidEnd              = "label key ends with invalid character"
	ErrLabelKeyInvalidCharacters       = "label key contains invalid characters"
	ErrLabelValueEmpty                 = "label value is empty"
	ErrLabelValueTooLong               = "label value is too long"
	ErrLabelValueInvalidStart          = "label value starts with invalid character"
	ErrLabelValueInvalidEnd            = "label value ends with invalid character"
	ErrLabelValueInvalidCharacters     = "label value contains invalid characters"
	ErrBuildRequestNil                 = "build request is nil"
	ErrCloudEventNil                   = "cloud event is nil"
	ErrInvalidNamespaceFormat          = "invalid namespace format"
	ErrInvalidServiceAccountFormat     = "invalid service account format"
	ErrInvalidAWSRegionFormat          = "invalid AWS region format"
	ErrInvalidAWSAccountIDFormat       = "invalid AWS account ID format"
	ErrInvalidECRRegistryFormat        = "invalid ECR registry format"
	ErrThirdPartyIDEmpty               = "third party ID is empty"
	ErrParserIDEmpty                   = "parser ID is empty"
	ErrSourceBucketEmpty               = "source bucket is empty"
	ErrSourceKeyEmpty                  = "source key is empty"
	ErrCorrelationIDEmpty              = "correlation ID is empty"
	ErrEventIDEmpty                    = "event ID is empty"
	ErrEventTypeEmpty                  = "event type is empty"
	ErrEventSourceEmpty                = "event source is empty"
	ErrEventTimeEmpty                  = "event time is empty"
	ErrUnsupportedEventType            = "unsupported event type: %s"
	ErrInvalidEventSource              = "invalid event source: %s"
	ErrEventDataEmpty                  = "event data is empty"
	ErrIDTooLong                       = "ID is too long"
	ErrIDContainsInvalidChars          = "ID contains invalid characters"
	ErrInvalidS3BucketLength           = "invalid S3 bucket length"
	ErrInvalidS3BucketChars            = "invalid S3 bucket characters"
	ErrS3KeyTooLong                    = "S3 key is too long"
	ErrInvalidS3KeyFormat              = "invalid S3 key format"
	ErrAWSRegionEmpty                  = "AWS region is empty"
	ErrInvalidAWSRegionLength          = "invalid AWS region length"
	ErrAWSAccountIDEmpty               = "AWS account ID is empty"
	ErrInvalidAWSAccountIDLength       = "invalid AWS account ID length"
	ErrECRRegistryEmpty                = "ECR registry is empty"
	ErrInvalidECRRegistrySuffix        = "invalid ECR registry suffix"
	ErrInvalidEventBody                = "invalid event body"
	ErrInvalidBuildCompletionEventBody = "invalid build completion event body"
	ErrImageURIEmpty                   = "image URI is empty"

	// Sidecar-specific error constants
	ErrKanikoNamespaceRequired         = "kaniko namespace is required"
	ErrKanikoPodNameRequired           = "kaniko pod name is required"
	ErrBuildJobNameRequired            = "build job name is required"
	ErrImageURIRequired                = "image URI is required"
	ErrThirdPartyIDRequired            = "third party ID is required"
	ErrParserIDRequired                = "parser ID is required"
	ErrCorrelationIDRequired           = "correlation ID is required"
	ErrKnativeBrokerURLRequired        = "knative broker URL is required"
	ErrInvalidMonitorInterval          = "invalid monitor interval: %v"
	ErrInvalidBuildTimeout             = "invalid build timeout: %v"
	ErrInvalidTLSEnabled               = "invalid TLS enabled: %v"
	ErrInvalidMetricsEnabled           = "invalid metrics enabled: %v"
	ErrInvalidRunAsUser                = "invalid run as user: %v"
	ErrInvalidRunAsGroup               = "invalid run as group: %v"
	ErrK8sConfigFailed                 = "failed to create Kubernetes config"
	ErrK8sClientFailed                 = "failed to create Kubernetes client"
	ErrFailedToGetService              = "failed to get service %s: %w"
	ErrKanikoContainerNotFound         = "kaniko container not found"
	ErrFailedToSetEventData            = "failed to set event data"
	ErrFailedToCreateCloudEventsClient = "failed to create CloudEvents client"
	ErrBrokerURLNotConfigured          = "broker URL not configured"
	ErrFailedToSendEventToBroker       = "failed to send event to broker"
	ErrInvalidMetricsPort              = "invalid metrics port: %v"
	ErrInvalidPort                     = "invalid port: %v"
	ErrTLSCertAndKeyRequired           = "TLS certificate and key are required when TLS is enabled"
)

// Resource and size limits
const (
	// 🔧 K8sMaxParserSizeDefault - Maximum size of parser files in bytes (100MB)
	//    Can be overridden via environment variable: MAX_PARSER_SIZE
	//    Example: export MAX_PARSER_SIZE=104857600 (100MB)
	//    Code location: internal/config/builder.go:369
	K8sMaxParserSizeDefault = 104857600 // 100MB (matches values.yaml maxParserSize: "104857600")

	// Request size default
	MaxRequestSizeDefault = 10485760 // 10MB
)

// CPU/Memory resource defaults
const (
	CPULimitDefault      = "2000m"
	MemoryRequestDefault = "512Mi"
	MemoryLimitDefault   = "2Gi"
)

// Port and timeout defaults
const (
	PortDefault        = 8080
	MetricsPortDefault = 8080 // 🔧 FIXED: Metrics are served on the same HTTP port via /metrics endpoint

	RequestTimeoutDefault = 30 * time.Second
	APITimeoutDefault     = 400 * time.Millisecond // Updated to match values.yaml apiTimeout: "400ms"
)

// Service and job defaults
const (
	ServiceAccountDefault   = "knative-lambda-builder"
	K8sJobTTLSecondsDefault = 3600
)

// Service identification
const (
	ServiceNameDefault    = "knative-lambda-new" // Updated to match values.yaml serviceName: "knative-lambda-new"
	ServiceVersionDefault = "1.0.0"
)

// Runtime and handler defaults
const (
	RuntimeDefault = "nodejs22"
	HandlerDefault = "index.handler"
	TriggerDefault = "http"
)

// Function resource defaults
const (
	FunctionMemoryLimitDefault   = "256Mi" // Updated to match values.yaml lambdaFunctionMemoryLimit: "256Mi"
	FunctionCPULimitDefault      = "100m"  // Updated to match values.yaml lambdaFunctionCpuLimit: "100m"
	FunctionMemoryRequestDefault = "64Mi"  // Updated to match values.yaml lambdaFunctionMemoryRequest: "64Mi"
	FunctionCPURequestDefault    = "50m"   // Updated to match values.yaml lambdaFunctionCpuRequest: "50m"
	FunctionMemoryLimitMiDefault = "256"   // Updated to match values.yaml functionMemoryLimitMi: "256"
	FunctionCPULimitMDefault     = "100m"  // Updated to match values.yaml lambdaFunctionCpuLimit: "100m"
)

// Event and broker configuration - These are truly constant values
const (
	// Event types and routing
	EventTypeDefault        = "network.notifi.lambda.parser.start"
	BrokerNameDefault       = "knative-lambda-service-broker-dev" // Updated to match values.yaml brokerName: "knative-lambda-service-broker-dev"
	TriggerNamespaceDefault = "knative-lambda-dev"                // Updated to match values.yaml triggerNamespace: "knative-lambda-dev"
	BrokerURLDefault        = "broker-ingress.knative-eventing.svc.cluster.local"
	BrokerPortDefault       = "80"

	// Delivery retry configuration
	DeliveryRetriesDefault       = "5"
	DeliveryBackoffPolicyDefault = "exponential"
	DeliveryBackoffDelayDefault  = "PT1S"

	// RabbitMQ eventing configuration
	RabbitMQEventingParallelismDefault = 50 // Number of parallel consumers for RabbitMQ events
)

// Builder service defaults - These are configurable but have sensible defaults
// 🔧 These constants control how your builder service scales up and down
const (
	BuilderServiceTargetConcurrencyDefault = "5" // Updated to match values.yaml builder.targetConcurrency: 5

	BuilderServiceTargetUtilizationDefault = "50"

	BuilderServiceTargetDefault = "5" // Updated to match values.yaml builder.targetConcurrency: 5

	BuilderServiceContainerConcurrencyDefault = "5" // Updated to match values.yaml builder.targetConcurrency: 5

	BuilderServiceMinScaleDefault = "0" // Can scale to zero (cost savings)

	BuilderServiceMaxScaleDefault = "10" // Updated to match values.yaml builder.maxScale: 10

	BuilderServiceScaleToZeroGracePeriodDefault = "30s" // Wait 30s before scaling to zero

	BuilderServiceScaleDownDelayDefault = "0s" // Scale down immediately

	BuilderServiceStableWindowDefault = "10s" // Average 10s of metrics
)

// Lambda Services Configuration Defaults - For dynamically created lambda services
// 🚨 CRITICAL: These control scaling for lambda services (not builder service)
const (
	// 🔧 AUTOSCALING: Configuration for lambda service autoscaling
	// 🚨 CRITICAL FIX: Updated for proper concurrency handling
	LambdaServicesMinScaleDefault               = "0"   // ✅ KnativeMinScaleDefault = "0"
	LambdaServicesMaxScaleDefault               = "50"  // ✅ KnativeMaxScaleDefault = "50"
	LambdaServicesTargetConcurrencyDefault      = "10"  // 🚨 FIXED: 10 request per pod for aggressive scaling
	LambdaServicesTargetUtilizationDefault      = "50"  // 🚨 FIXED: 50% of target concurrency triggers new pod
	LambdaServicesTargetDefault                 = "10"  // 🚨 FIXED: Autoscaling target of 10 concurrent requests
	LambdaServicesContainerConcurrencyDefault   = "10"  // 🚨 FIXED: 10 request per container for maximum distribution
	LambdaServicesScaleToZeroGracePeriodDefault = "30s" // ✅ KnativeScaleToZeroGracePeriodDefault = "30s"
	LambdaServicesScaleDownDelayDefault         = "0s"  // ✅ KnativeScaleDownDelayDefault = "0s"
	LambdaServicesStableWindowDefault           = "10s" // ✅ KnativeStableWindowDefault = "10s"
	// 🚀 PANIC MODE: For rapid scaling during traffic spikes
	LambdaServicesPanicWindowPercentageDefault    = "10.0"  // 10% of stable window for panic mode
	LambdaServicesPanicThresholdPercentageDefault = "200.0" // 200% of target triggers panic mode

	// 📦 RESOURCE CONFIGURATION - Specific to lambda services
	LambdaServicesResourceMemoryRequestDefault = "64Mi"  // Updated to match values.yaml lambdaServices.defaults.resources.requests.memory: "64Mi"
	LambdaServicesResourceCPURequestDefault    = "50m"   // Updated to match values.yaml lambdaServices.defaults.resources.requests.cpu: "50m"
	LambdaServicesResourceMemoryLimitDefault   = "256Mi" // Updated to match values.yaml lambdaServices.defaults.resources.limits.memory: "256Mi"
	LambdaServicesResourceCPULimitDefault      = "100m"  // Updated to match values.yaml lambdaServices.defaults.resources.limits.cpu: "100m"

	// ⏰ TIMEOUT CONFIGURATION - Specific to lambda services
	LambdaServicesTimeoutResponseDefault = "60s"
	LambdaServicesTimeoutIdleDefault     = "30s"

	// 🔧 NPM CONFIGURATION - For reliable npm installs in Kaniko builds
	NpmConfigTimeoutDefault              = "60000"
	NpmConfigFetchRetriesDefault         = "5"
	NpmConfigFetchRetryMinTimeoutDefault = "10000"
	NpmConfigFetchRetryMaxTimeoutDefault = "60000"
	NpmConfigFetchRetryFactorDefault     = "2"
	NpmConfigMaxSocketsDefault           = "50"
)

// Main configuration defaults
const (
	ConfigEnvironmentDefault = "dev"
)

// Metrics pusher defaults
const (
	MetricsPusherQueueProxyPortDefault     = "9091"
	MetricsPusherFailureToleranceDefault   = 5
	MetricsPusherBuilderMetricsPortDefault = "9092"
	MetricsPusherEnabledDefault            = true
	MetricsPusherImageRegistryDefault      = "ghcr.io/brunovlucena"
	MetricsPusherImageRepositoryDefault    = "knative-lambda-knative-lambda-metrics-pusher"
	MetricsPusherImageTagDefault           = "latest"
	MetricsPusherImagePullPolicyDefault    = "Always"
	MetricsPusherRemoteWriteURLDefault     = "http://prometheus-kube-prometheus-prometheus.prometheus:9090/api/v1/write"
	MetricsPusherPushIntervalDefault       = 30 * time.Second
	MetricsPusherTimeoutDefault            = 10 * time.Second
	MetricsPusherLogLevelDefault           = "info"
	MetricsPusherLogFormatDefault          = "json"
	MetricsPusherQueueProxyPathDefault     = "/metrics"
	MetricsPusherBuilderPathDefault        = "/metrics"
	MetricsPusherCPURequestDefault         = "50m"
	MetricsPusherMemoryRequestDefault      = "64Mi"
	MetricsPusherCPULimitDefault           = "100m"
	MetricsPusherMemoryLimitDefault        = "128Mi"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// TIMEOUT CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Kubernetes operation timeouts
const (
	K8sOperationTimeout = 30 * time.Second
	K8sWaitTimeout      = 300 * time.Second
	K8sPodTimeout       = 600 * time.Second
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// LIMIT CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Kubernetes job limits
const (
	JobBackoffLimit       = 2
	JobMaxRetries         = 3
	JobMaxConcurrentJobs  = 10
	JobMaxNameLength      = 63
	JobMaxLabelLength     = 63
	K8sMaxNamespaceLength = 63
)

// Validation and AWS limits
const (
	// ID and name length limits
	MaxThirdPartyIDLength  = 100
	MaxParserIDLength      = 100
	MaxCorrelationIDLength = 100
	MaxBucketNameLength    = 63
	MaxObjectKeyLength     = 1024
	MaxS3KeyLength         = 1024

	// AWS-specific limits
	AWSS3MaxKeyLength        = 1024
	AWSS3MaxBucketNameLength = 63
	AWSS3MinBucketNameLength = 3
	MinAWSRegionLength       = 8
	MaxAWSRegionLength       = 20
	MaxRetriesDefault        = 3
	MinS3BucketNameLength    = 3
	MaxS3BucketNameLength    = 63
)

// Resource limits
const (
	MaxCPURequest     = "1000m"
	MaxMemoryRequest  = "2Gi"
	MinCPURequest     = "100m"
	MinMemoryRequest  = "128Mi"
	CPURequestDefault = "500m"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// LABEL CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Application labels
const (
	AppLabelName         = "app"
	AppLabelValue        = "knative-lambda-builder"
	EnvironmentLabelName = "environment"
	ComponentLabelName   = "component"
	ComponentLabelValue  = "service"
)

// Build labels
const (
	BuildThirdPartyIDLabel = "build.notifi.network/third-party-id"
	BuildParserIDLabel     = "build.notifi.network/parser-id"
	BuildJobNameLabel      = "build.notifi.network/job-name"
	BuildCorrelationLabel  = "build.notifi.network/correlation-id"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// ERROR MESSAGE CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Kubernetes error messages
const (
	K8sJobCreationFailed     = "failed to create Kubernetes job"
	K8sJobDeletionFailed     = "failed to delete Kubernetes job"
	K8sJobRetrievalFailed    = "failed to retrieve Kubernetes job"
	K8sJobListingFailed      = "failed to list Kubernetes jobs"
	K8sResourceNotFound      = "Kubernetes resource not found"
	K8sResourceAlreadyExists = "Kubernetes resource already exists"
)

// AWS error messages
const (
	AWSS3UploadFailed      = "failed to upload object to S3"
	AWSS3DownloadFailed    = "failed to download object from S3"
	AWSS3DeleteFailed      = "failed to delete object from S3"
	AWSEcrPushFailed       = "failed to push image to ECR"
	AWSEcrPullFailed       = "failed to pull image from ECR"
	AWSEcrRepositoryFailed = "failed to create ECR repository"

	// AWS validation error messages - CRITICAL for AWS integration
	ErrAWSRegionRequired          = "AWS region is required"
	ErrAWSAccountIDRequired       = "AWS account ID is required"
	ErrAWSAccountIDMustBe12Digits = "AWS account ID must be 12 digits"
	ErrECRRegistryRequired        = "ECR registry URL is required"
	ErrECRRepositoryNameRequired  = "ECR repository name is required"
	ErrS3SourceBucketRequired     = "S3 source bucket is required"
	ErrS3TempBucketRequired       = "S3 temp bucket is required"
	ErrPodIdentityRoleRequired    = "Pod identity role is required when EKS pod identity is enabled"

	// Build validation error messages
	ErrKanikoImageRequired     = "Kaniko image is required"
	ErrSidecarImageRequired    = "Sidecar image is required"
	ErrBuildTimeoutMin5Minutes = "must be at least 5 minutes"
	ErrCPURequestRequired      = "CPU request is required"
	ErrCPULimitRequired        = "CPU limit is required"
	ErrMemoryRequestRequired   = "Memory request is required"
	ErrMemoryLimitRequired     = "Memory limit is required"
	ErrMaxParserSizePositive   = "must be positive"

	// HTTP validation error messages
	ErrPortRange1024To65535           = "must be between 1024-65535"
	ErrMetricsPortRange1024To65535    = "must be between 1024-65535"
	ErrTimeoutMin1Second              = "must be at least 1 second"
	ErrAPITimeoutRange50msTo1s        = "must be between 50ms and 1s"
	ErrMaxRequestSizePositive         = "must be positive"
	ErrDefaultListLimitRange1To1000   = "must be between 1 and 1000"
	ErrMaxListLimitRange1To1000       = "must be between 1 and 1000"
	ErrDefaultListLimitGreaterThanMax = "cannot be greater than max_list_limit"

	// Knative validation error messages
	ErrTargetConcurrencyRequired       = "target concurrency is required"
	ErrTargetUtilizationRequired       = "target utilization is required"
	ErrTargetRequired                  = "target is required"
	ErrContainerConcurrencyRequired    = "container concurrency is required"
	ErrMinScaleRequired                = "min scale is required"
	ErrMaxScaleRequired                = "max scale is required"
	ErrDefaultEventTypeRequired        = "default event type is required"
	ErrDefaultBrokerNameRequired       = "default broker name is required"
	ErrDefaultTriggerNamespaceRequired = "default trigger namespace is required"

	// Lambda validation error messages
	ErrDefaultRuntimeRequired        = "default runtime is required"
	ErrDefaultHandlerRequired        = "default handler is required"
	ErrDefaultTriggerRequired        = "default trigger is required"
	ErrFunctionMemoryLimitRequired   = "function memory limit is required"
	ErrFunctionCPULimitRequired      = "function CPU limit is required"
	ErrFunctionMemoryRequestRequired = "function memory request is required"
	ErrFunctionCPURequestRequired    = "function CPU request is required"

	// Rate limiting validation error messages
	ErrRateLimitPositive                = "must be positive"
	ErrMaxMemoryUsagePercentRange0To100 = "must be between 0 and 100"
	ErrMemoryCheckIntervalPositive      = "must be positive"
	ErrCleanupIntervalPositive          = "must be positive"
	ErrClientTTLPositive                = "must be positive"
	ErrMaxConcurrentBuildsPositive      = "must be positive"
	ErrMaxConcurrentJobsPositive        = "must be positive"
	ErrBuildTimeoutPositive             = "must be positive"
	ErrJobTimeoutPositive               = "must be positive"
	ErrRequestTimeoutPositive           = "must be positive"

	// Kubernetes validation error messages
	ErrK8sNamespaceValid            = "must be a valid Kubernetes namespace"
	ErrK8sServiceAccountValid       = "must be a valid Kubernetes name"
	ErrK8sRunAsUserValid            = "must be a valid user ID"
	ErrK8sJobTTLMin60Seconds        = "must be at least 60 seconds"
	ErrK8sJobDeletionWaitMin1Second = "must be at least 1 second"
	ErrK8sJobDeletionCheckMin100ms  = "must be at least 100ms"

	// Observability validation error messages
	ErrLogLevelValid                 = "must be one of: debug, info, warn, error"
	ErrSampleRateRange0To1           = "must be between 0 and 1"
	ErrServiceNameRequired           = "service name is required"
	ErrServiceVersionRequired        = "service version is required"
	ErrExemplarsSampleRateRange0To1  = "must be between 0 and 1"
	ErrExemplarsMaxPerMetricPositive = "must be greater than 0"
)

// Configuration error messages
const (
	ErrConfigRequired = "configuration is required"
	ErrConfigInvalid  = "invalid configuration"
	ErrConfigMissing  = "missing required configuration"
)

// Security error messages
const (
	ErrSecurityForbidden    = "forbidden"
	ErrSecurityUnauthorized = "unauthorized"
)

// Kubernetes error messages
const (
	ErrK8sConnection = "kubernetes connection failed"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// CONFIGURATION CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Default configuration values
const (
	DefaultNamespace      = "knative-lambda"
	DefaultServiceAccount = "knative-lambda-builder"
	DefaultKanikoImage    = "gcr.io/kaniko-project/executor:latest"
	DefaultSidecarImage   = "knative-lambda-sidecar:latest"
	DefaultDockerfilePath = "Dockerfile"
)

// AWS configuration defaults - MATCHING values.yaml
const (
	AWSRegionDefault             = "us-west-2"                                    // values.yaml awsRegion: "us-west-2"
	AWSAccountIDDefault          = "339954290315"                                 // values.yaml awsAccountId: "339954290315"
	RegistryDefault              = "339954290315.dkr.ecr.us-west-2.amazonaws.com" // 🔧 ONE registry for everything
	S3SourceBucketDefault        = "notifi-uw2-dev-fusion-modules"                // values.yaml s3SourceBucket: "notifi-uw2-dev-fusion-modules"
	S3TempBucketDefault          = "knative-lambda-dev-context-tmp"               // values.yaml s3TmpBucket: "knative-lambda-dev-context-tmp"
	RegistryMirrorDefault        = "docker.io"                                    // 🔧 Empty = pull base images directly from Docker Hub
	SkipTLSVerifyRegistryDefault = "docker.io"                                    // 🔧 Default to docker.io for TLS verification skip
	NodeBaseImageDefault         = "docker.io/library/node:22-alpine"             // values.yaml nodeBaseImage: "docker.io/library/node:22-alpine"
	PythonBaseImageDefault       = "docker.io/library/python:3.11-alpine"         // values.yaml pythonBaseImage: "docker.io/library/python:3.11-alpine"
	GoBaseImageDefault           = "docker.io/library/golang:1.21-alpine"         // values.yaml goBaseImage: "docker.io/library/golang:1.21-alpine"
)

// 💾 Storage configuration defaults - MATCHING values.yaml
const (
	StorageProviderDefault      = "aws-s3"                                 // values.yaml storageProvider: "aws-s3" (or "minio")
	S3EndpointDefault           = "https://s3.us-west-2.amazonaws.com"     // values.yaml s3Endpoint (optional for AWS S3)
	MinIOEndpointDefault        = "minio.minio.svc.cluster.local:9000"     // values.yaml minioEndpoint
	MinIOAccessKeyDefault       = ""                                       // values.yaml minioAccessKey (from secret)
	MinIOSecretKeyDefault       = ""                                       // values.yaml minioSecretKey (from secret)
	MinIOUseSSLDefault          = false                                    // values.yaml minioUseSSL: false (HTTP for internal cluster)
	MinIORegionDefault          = "us-east-1"                              // values.yaml minioRegion: "us-east-1" (MinIO default)
	MinIOSourceBucketDefault    = "knative-lambda-source"                  // values.yaml minioSourceBucket
	MinIOTempBucketDefault      = "knative-lambda-tmp"                     // values.yaml minioTempBucket
	AWSECRRepositoryNameDefault = "knative-lambda"                         // values.yaml ecrRepositoryName: "knative-lambda"
	PodIdentityRoleDefault      = "knative-lambda-builder"                 // values.yaml podIdentityRole: "knative-lambda-builder"
	KanikoImageDefault          = "gcr.io/kaniko-project/executor:v1.19.2" // values.yaml kanikoImage: "gcr.io/kaniko-project/executor:v1.19.2"
)

// Container names
const (
	ContainerNameKaniko  = "kaniko"
	ContainerNameSidecar = "sidecar"
)

// Timeout, monitoring, and environment constants
const (
	// Build and operation timeouts
	BuildTimeoutDefault = 30 * time.Minute

	// Kubernetes security and networking
	K8sRunAsUserDefault   = 65534
	K8sMinPort            = 1024
	K8sMaxPort            = 65535
	K8sMetricsPortDefault = 9091

	// Monitoring paths
	MetricsPath = "/metrics"

	// Environment variable names
	EnvEnvironment    = "ENVIRONMENT"
	EnvLogLevel       = "LOG_LEVEL"
	EnvMetricsEnabled = "METRICS_ENABLED"
	EnvTracingEnabled = "TRACING_ENABLED"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// MONITORING CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Monitoring intervals
const (
	MonitorIntervalDefault = "10s"
	MonitorIntervalShort   = "5s"
	MonitorIntervalLong    = "30s"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// VALIDATION CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Validation patterns
const (
	ValidThirdPartyIDPattern  = `^[a-zA-Z0-9_-]+$`
	ValidParserIDPattern      = `^[a-zA-Z0-9_-]+$`
	ValidCorrelationIDPattern = `^[a-zA-Z0-9_-]+$`
	ValidBucketNamePattern    = `^[a-z0-9][a-z0-9.-]*[a-z0-9]$`
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// RATE LIMITING CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Rate limiting configuration defaults - Critical for system stability and performance
const (
	// 🔧 RateLimitingEnabledDefault - Master switch for rate limiting
	//    - Default: true (rate limiting enabled)
	//    - IMPACT: Disabling can lead to resource exhaustion and system instability
	RateLimitingEnabledDefault = true

	// 🚀 Build context rate limiting - Controls S3 download frequency
	//    - Default: 5 requests/min, burst: 2 (updated to match values.yaml)
	//    - IMPACT: Lower values = slower builds, higher values = potential S3 throttling
	BuildContextRequestsPerMinDefault = 5 // Updated to match values.yaml buildContextRequestsPerMin: "5"
	BuildContextBurstSizeDefault      = 2 // Updated to match values.yaml buildContextBurstSize: "2"

	// ⚡ Kubernetes job rate limiting - Controls K8s API calls
	//    - Default: 10 requests/min, burst: 3 (updated to match values.yaml)
	//    - IMPACT: Lower values = slower job creation, higher values = potential API throttling
	K8sJobRequestsPerMinDefault = 10 // Updated to match values.yaml k8sJobRequestsPerMin: "10"
	K8sJobBurstSizeDefault      = 3  // Updated to match values.yaml k8sJobBurstSize: "3"

	// 🌐 Client rate limiting - Controls HTTP request frequency
	//    - Default: 5 requests/min, burst: 2 (updated to match values.yaml)
	//    - IMPACT: Lower values = slower client operations, higher values = potential overload
	ClientRequestsPerMinDefault = 5 // Updated to match values.yaml clientRequestsPerMin: "5"
	ClientBurstSizeDefault      = 2 // Updated to match values.yaml clientBurstSize: "2"

	// 📤 S3 upload rate limiting - Controls S3 upload frequency
	//    - Default: 50 requests/min, burst: 10 (updated to match values.yaml)
	//    - IMPACT: Lower values = slower uploads, higher values = potential S3 throttling
	S3UploadRequestsPerMinDefault = 50 // Updated to match values.yaml s3UploadRequestsPerMin: "50"
	S3UploadBurstSizeDefault      = 10 // Updated to match values.yaml s3UploadBurstSize: "10"

	// 🎯 Performance and concurrency limits - CRITICAL for system stability
	//    - MaxMemoryUsagePercentDefault: 80% (prevents OOM)
	//    - MaxConcurrentBuildsDefault: 10 (prevents resource exhaustion)
	//    - MaxConcurrentJobsDefault: 5 (prevents K8s API overload, updated to match values.yaml)
	//    - IMPACT: These values directly affect system stability and performance
	MaxMemoryUsagePercentDefault = 80.0
	MaxConcurrentBuildsDefault   = 10
	MaxConcurrentJobsDefault     = 5 // Updated to match values.yaml maxConcurrentJobs: "5"

	// ⏰ Memory and cleanup intervals - System maintenance
	//    - MemoryCheckIntervalDefault: 30s (memory monitoring frequency)
	//    - CleanupIntervalDefault: 5m (cleanup frequency)
	//    - ClientTTLDefault: 1h (client session timeout, updated to match values.yaml)
	MemoryCheckIntervalDefault = 30 * time.Second
	CleanupIntervalDefault     = 5 * time.Minute
	ClientTTLDefault           = 1 * time.Hour // Updated to match values.yaml clientTTL: "1h"

	// ⏱️ Timeout defaults for rate limiting - CRITICAL for preventing hanging operations
	//    - RateLimitBuildTimeoutDefault: 30m (max build duration, updated to match values.yaml)
	//    - RateLimitJobTimeoutDefault: 1h (max job duration, updated to match values.yaml)
	//    - RateLimitRequestTimeoutDefault: 5m (max request duration, updated to match values.yaml)
	//    - IMPACT: These prevent resource leaks and hanging operations
	RateLimitBuildTimeoutDefault   = 30 * time.Minute // Updated to match values.yaml buildTimeout: "30m"
	RateLimitJobTimeoutDefault     = 1 * time.Hour    // Updated to match values.yaml jobTimeout: "1h"
	RateLimitRequestTimeoutDefault = 5 * time.Minute  // Updated to match values.yaml requestTimeout: "5m"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// OBSERVABILITY CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Observability configuration defaults - Critical for monitoring and debugging
const (
	// 🔍 Tracing and sampling - CRITICAL for distributed tracing
	//    - TracingEnabledDefault: false (disabled by default for performance)
	//    - SampleRateDefault: 1.0 (100% sampling when enabled)
	//    - MetricsEnabledDefault: true (metrics always enabled)
	//    - OTLPEndpointDefault: "" (no endpoint by default)
	//    - IMPACT: Enabling tracing adds overhead but provides visibility
	TracingEnabledDefault = false
	SampleRateDefault     = 1.0
	MetricsEnabledDefault = true
	OTLPEndpointDefault   = ""

	// 📊 Exemplars configuration - Advanced metrics with trace correlation
	//    - ExemplarsEnabledDefault: false (disabled by default)
	//    - ExemplarsMaxPerMetricDefault: 10 (max exemplars per metric)
	//    - ExemplarsSampleRateDefault: 0.1 (10% sampling)
	//    - IMPACT: Exemplars provide trace correlation but increase metric cardinality
	ExemplarsEnabledDefault       = false
	ExemplarsMaxPerMetricDefault  = 10
	ExemplarsSampleRateDefault    = 0.1
	ExemplarsTraceIDLabelDefault  = "trace_id"
	ExemplarsSpanIDLabelDefault   = "span_id"
	ExemplarsIncludeLabelsDefault = "true"

	// 🔄 System metrics collection interval - CRITICAL for monitoring
	//    - ObservabilitySystemMetricsCollectionIntervalDefault: 30s (collect every 30 seconds)
	//    - IMPACT: Affects monitoring granularity and resource usage
	ObservabilitySystemMetricsCollectionIntervalDefault = 30 * time.Second
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// SECURITY CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Security and infrastructure configuration defaults
const (
	// 🔒 Security configuration - CRITICAL for system security
	//    - ValidateInputDefault: true (input validation enabled)
	//    - SecurityEnabledDefault: true (security features enabled)
	//    - DebugModeDefault: false (debug mode disabled for security)
	//    - DryRunDefault: false (dry run disabled)
	//    - IMPACT: These settings directly affect system security posture
	ValidateInputDefault   = true
	SecurityEnabledDefault = true
	DebugModeDefault       = false
	DryRunDefault          = false

	// ☸️ Kubernetes configuration
	//    - InClusterDefault: true (assume running in-cluster)
	InClusterDefault = true

	// ☁️ AWS configuration
	//    - UseEKSPodIdentityDefault: true (use EKS pod identity)
	UseEKSPodIdentityDefault = true
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// SECURITY SETTINGS CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Security settings
const (
	MaxRequestSize  = 10 * 1024 * 1024 // 10MB
	MaxResponseSize = 50 * 1024 * 1024 // 50MB
	MaxHeaderSize   = 1 * 1024 * 1024  // 1MB
	MaxURLLength    = 2048
	MaxHeaderCount  = 100
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// DEPLOYMENT CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Deployment settings
const (
	DefaultReplicas    = 1
	MaxReplicas        = 10
	MinReplicas        = 0
	DefaultPort        = 8081
	DefaultMetricsPort = 9090
	DefaultHealthPort  = 8081
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// LOGGING CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Log field names
const (
	LogFieldJobName       = "job_name"
	LogFieldNamespace     = "namespace"
	LogFieldThirdPartyID  = "third_party_id"
	LogFieldParserID      = "parser_id"
	LogFieldCorrelationID = "correlation_id"
	LogFieldError         = "error"
	LogFieldDuration      = "duration"
	LogFieldState         = "state"
	LogFieldOperation     = "operation"
	LogFieldComponent     = "component"
	LogFieldEnvironment   = "environment"
)

// Log levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// TESTING CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Test configuration
const (
	TestTimeoutShort  = 5 * time.Second
	TestTimeoutMedium = 30 * time.Second
	TestTimeoutLong   = 2 * time.Minute
	TestRetryAttempts = 3
	TestRetryDelay    = 100 * time.Millisecond
)

// Test data
const (
	TestThirdPartyID  = "test-third-party-id-12345"
	TestParserID      = "test-parser-id-67890"
	TestCorrelationID = "test-correlation-id-abcde"
	TestBucketName    = "test-bucket-name"
	TestObjectKey     = "test/object/key"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// HELPER FUNCTIONS - Utility functions for constants
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// IsValidPort checks if a port number is within valid range
func IsValidPort(port int) bool {
	return port >= K8sMinPort && port <= K8sMaxPort
}

// IsValidNamespace checks if a namespace name is valid
func IsValidNamespace(name string) bool {
	return len(name) > 0 && len(name) <= JobMaxNameLength
}

// IsValidName checks if a Kubernetes resource name is valid
func IsValidName(name string) bool {
	return len(name) > 0 && len(name) <= JobMaxNameLength
}

// IsValidUserID checks if a user ID is within valid range
func IsValidUserID(userID int64) bool {
	return userID >= 1 && userID <= 65534
}

// IsValidLogLevel checks if a log level is valid
func IsValidLogLevel(level string) bool {
	return level == LogLevelDebug || level == LogLevelInfo || level == LogLevelWarn || level == LogLevelError
}

// Environment names
const (
	EnvironmentDev   = "dev"
	EnvironmentPrd   = "prd"
	EnvironmentLocal = "local"
)

// HTTP status codes
const (
	HTTPStatusOK                  = 200
	HTTPStatusCreated             = 201
	HTTPStatusAccepted            = 202
	HTTPStatusBadRequest          = 400
	HTTPStatusNotFound            = 404
	HTTPStatusInternalServerError = 500
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// MEMORY AND SIZE CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Memory size constants (in bytes)
const (
	OneKB     = 1024
	OneMB     = 1024 * 1024
	TenMB     = 10 * 1024 * 1024
	HundredMB = 100 * 1024 * 1024
	OneGB     = 1024 * 1024 * 1024
	TenGB     = 10 * 1024 * 1024 * 1024
	HundredGB = 100 * 1024 * 1024 * 1024
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// CONCURRENCY AND LIMIT CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Concurrency limits
const (
	DefaultConcurrency = 100
	MaxConcurrency     = 1000
	MinConcurrency     = 10
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// STRING CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Common string constants
const (
	TestStringDefault = "test"
	SuccessString     = "success"
	ErrorString       = "error"
	DefaultString     = "default"
	UnknownString     = "unknown"
)

// HTTP server timeouts
const (
	HTTPReadTimeoutDefault       = 5 * time.Second
	HTTPWriteTimeoutOffset       = 500 * time.Millisecond
	HTTPIdleTimeoutDefault       = 120 * time.Second
	HTTPReadHeaderTimeoutDefault = 2 * time.Second
	HTTPGracefulShutdownTimeout  = 30 * time.Second
)

// HTTP list configuration defaults
const (
	DefaultListLimit = 50  // Updated to match values.yaml defaultListLimit: "50"
	MaxListLimit     = 100 // Updated to match values.yaml maxListLimit: "100"
)

// Job management timeouts
const (
	JobDeletionWaitTimeoutDefault   = 30 * time.Second
	JobDeletionCheckIntervalDefault = 5 * time.Second // Updated to match values.yaml jobDeletionCheckInterval: "5s"
	JobFailureBackoffPeriodDefault  = 5 * time.Minute
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// NOTIFI SERVICE ADDRESS CONSTANTS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// Notifi service address defaults
const (
	NotifiSubscriptionManagerAddressDefault = "notifi-subscription-manager.notifi.svc.cluster.local:4000"
	NotifiEphemeralStorageAddressDefault    = "notifi-storage-manager.notifi.svc.cluster.local:4000"
	NotifiPersistentStorageAddressDefault   = "notifi-storage-manager.notifi.svc.cluster.local:4000"
	NotifiFusionFetchProxyAddressDefault    = "notifi-fetch-proxy.notifi.svc.cluster.local:4000"
	NotifiEvmRpcAddressDefault              = "notifi-blockchain-manager.notifi.svc.cluster.local:4000"
	NotifiSolanaRpcAddressDefault           = "notifi-blockchain-manager.notifi.svc.cluster.local:4000"
	NotifiSuiRpcAddressDefault              = "notifi-blockchain-manager.notifi.svc.cluster.local:4000"
	NotifiGrpcInsecureDefault               = true
)
