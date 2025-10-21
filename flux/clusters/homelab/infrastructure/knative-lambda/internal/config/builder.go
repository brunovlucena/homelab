// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🔧 CONFIG BUILDER - Fluent interface for centralized config assembly
//
//	🎯 Purpose: Centralized configuration assembly with fluent interface
//	💡 Features: Fluent interface, validation, environment variable mapping, hot-reloading
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
)

// 🔧 ConfigBuilder - "Fluent interface for configuration building"
type ConfigBuilder struct {
	config *Config
	err    error
}

// 🔧 NewConfigBuilder - "Create a new configuration builder"
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &Config{
			HTTP:       NewHTTPConfig(),
			Kubernetes: NewKubernetesConfig(),
			AWS:        NewAWSConfig(),

			Observability:  NewObservabilityConfig(),
			Build:          NewBuildConfig(),
			Lambda:         NewLambdaConfig(),
			LambdaServices: NewLambdaServicesConfig(),
			Knative:        &KnativeConfig{},
			Security:       NewSecurityConfig(),
			RateLimiting:   NewRateLimitingConfig(),
			Notifi:         NewNotifiConfig(),
			MetricsPusher:  NewMetricsPusherConfig(),
		},
	}
}

// 🔧 WithEnvironment - "Set environment"
func (b *ConfigBuilder) WithEnvironment(env string) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	b.config.Environment = env
	return b
}

// 🔧 WithHTTPConfig - "Set HTTP configuration"
func (b *ConfigBuilder) WithHTTPConfig(httpConfig *HTTPConfig) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	b.config.HTTP = httpConfig
	return b
}

// 🔧 WithKubernetesConfig - "Set Kubernetes configuration"
func (b *ConfigBuilder) WithKubernetesConfig(k8sConfig *KubernetesConfig) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	b.config.Kubernetes = k8sConfig
	return b
}

// 🔧 WithAWSConfig - "Set AWS configuration"
func (b *ConfigBuilder) WithAWSConfig(awsConfig *AWSConfig) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	b.config.AWS = awsConfig
	return b
}

// 🔧 WithObservabilityConfig - "Set observability configuration"
func (b *ConfigBuilder) WithObservabilityConfig(obsConfig *ObservabilityConfig) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	b.config.Observability = obsConfig
	return b
}

// 🔧 WithBuildConfig - "Set build configuration"
func (b *ConfigBuilder) WithBuildConfig(buildConfig *BuildConfig) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	b.config.Build = buildConfig
	return b
}

// 🔧 WithLambdaConfig - "Set Lambda configuration"
func (b *ConfigBuilder) WithLambdaConfig(lambdaConfig *LambdaConfig) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	b.config.Lambda = lambdaConfig
	return b
}

// 🔧 WithKnativeConfig - "Set Knative configuration"
func (b *ConfigBuilder) WithKnativeConfig(knativeConfig *KnativeConfig) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	b.config.Knative = knativeConfig
	return b
}

// 🔧 WithSecurityConfig - "Set security configuration"
func (b *ConfigBuilder) WithSecurityConfig(securityConfig *SecurityConfig) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	b.config.Security = securityConfig
	return b
}

// 🔧 WithNotifiConfig - "Set Notifi configuration"
func (b *ConfigBuilder) WithNotifiConfig(notifiConfig *NotifiConfig) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	b.config.Notifi = notifiConfig
	return b
}

// 🔧 WithMetricsPusherConfig - "Set MetricsPusher configuration"
func (b *ConfigBuilder) WithMetricsPusherConfig(metricsPusherConfig *MetricsPusherConfig) *ConfigBuilder {
	if b.err != nil {
		return b
	}
	b.config.MetricsPusher = metricsPusherConfig
	return b
}

// 🔧 LoadFromEnvironment - "Load configuration from environment variables"
func (b *ConfigBuilder) LoadFromEnvironment() *ConfigBuilder {
	if b.err != nil {
		return b
	}

	// Load HTTP configuration
	b.loadHTTPFromEnvironment()
	b.loadKubernetesFromEnvironment()
	b.loadAWSFromEnvironment()
	b.loadKnativeFromEnvironment()
	b.loadObservabilityFromEnvironment()
	b.loadBuildFromEnvironment()
	b.loadLambdaFromEnvironment()
	b.loadLambdaServicesFromEnvironment() // 🚨 ADDED: Load lambda services configuration
	b.loadSecurityFromEnvironment()
	b.loadRateLimitingFromEnvironment()
	b.loadNotifiFromEnvironment()
	b.loadMetricsPusherFromEnvironment()
	b.loadPerformanceFromEnvironment()

	return b
}

// 🔧 Validate - "Validate all configuration components"
func (b *ConfigBuilder) Validate() *ConfigBuilder {
	if b.err != nil {
		return b
	}

	// Validate each component
	b.validateHTTPConfig()
	b.validateKubernetesConfig()
	b.validateAWSConfig()
	b.validateObservabilityConfig()
	b.validateBuildConfig()
	b.validateLambdaConfig()
	b.validateLambdaServicesConfig() // 🚨 ADDED: Validate lambda services configuration
	b.validateKnativeConfig()
	b.validateSecurityConfig()
	b.validateRateLimitingConfig()
	b.validateNotifiConfig()
	b.validateMetricsPusherConfig()

	return b
}

// 🔧 validateHTTPConfig - "Validate HTTP configuration"
func (b *ConfigBuilder) validateHTTPConfig() {
	if b.err != nil {
		return
	}
	if err := b.config.HTTP.Validate(); err != nil {
		b.err = errors.NewConfigurationError("http", "validation", fmt.Sprintf("HTTP config validation failed: %v", err))
	}
}

// 🔧 validateKubernetesConfig - "Validate Kubernetes configuration"
func (b *ConfigBuilder) validateKubernetesConfig() {
	if b.err != nil {
		return
	}
	if err := b.config.Kubernetes.Validate(); err != nil {
		b.err = errors.NewConfigurationError("kubernetes", "validation", fmt.Sprintf("Kubernetes config validation failed: %v", err))
	}
}

// 🔧 validateAWSConfig - "Validate AWS configuration"
func (b *ConfigBuilder) validateAWSConfig() {
	if b.err != nil {
		return
	}
	if err := b.config.AWS.Validate(); err != nil {
		b.err = errors.NewConfigurationError("aws", "validation", fmt.Sprintf("AWS config validation failed: %v", err))
	}
}

// 🔧 validateObservabilityConfig - "Validate observability configuration"
func (b *ConfigBuilder) validateObservabilityConfig() {
	if b.err != nil {
		return
	}
	if err := b.config.Observability.Validate(); err != nil {
		b.err = errors.NewConfigurationError("observability", "validation", fmt.Sprintf("Observability config validation failed: %v", err))
	}
}

// 🔧 validateBuildConfig - "Validate build configuration"
func (b *ConfigBuilder) validateBuildConfig() {
	if b.err != nil {
		return
	}
	if err := b.config.Build.Validate(); err != nil {
		b.err = errors.NewConfigurationError("build", "validation", fmt.Sprintf("Build config validation failed: %v", err))
	}
}

// 🔧 validateLambdaConfig - "Validate Lambda configuration"
func (b *ConfigBuilder) validateLambdaConfig() {
	if b.err != nil {
		return
	}
	if err := b.config.Lambda.Validate(); err != nil {
		b.err = errors.NewConfigurationError("lambda", "validation", fmt.Sprintf("Lambda config validation failed: %v", err))
	}
}

// 🔧 validateKnativeConfig - "Validate Knative configuration"
func (b *ConfigBuilder) validateKnativeConfig() {
	if b.err != nil {
		return
	}
	if err := b.config.Knative.Validate(); err != nil {
		b.err = errors.NewConfigurationError("knative", "validation", fmt.Sprintf("Knative config validation failed: %v", err))
	}
}

// 🔧 validateSecurityConfig - "Validate security configuration"
func (b *ConfigBuilder) validateSecurityConfig() {
	if b.err != nil {
		return
	}
	if err := b.config.Security.Validate(); err != nil {
		b.err = errors.NewConfigurationError("security", "validation", fmt.Sprintf("Security config validation failed: %v", err))
	}
}

// 🔧 validateRateLimitingConfig - "Validate rate limiting configuration"
func (b *ConfigBuilder) validateRateLimitingConfig() {
	if b.err != nil {
		return
	}
	if err := b.config.RateLimiting.Validate(); err != nil {
		b.err = errors.NewConfigurationError("rate_limiting", "validation", fmt.Sprintf("Rate limiting config validation failed: %v", err))
	}
}

// 🔧 validateNotifiConfig - "Validate Notifi configuration"
func (b *ConfigBuilder) validateNotifiConfig() {
	if b.err != nil {
		return
	}
	if err := b.config.Notifi.Validate(); err != nil {
		b.err = errors.NewConfigurationError("notifi", "validation", fmt.Sprintf("Notifi config validation failed: %v", err))
	}
}

// 🔧 validateMetricsPusherConfig - "Validate MetricsPusher configuration"
func (b *ConfigBuilder) validateMetricsPusherConfig() {
	if b.err != nil {
		return
	}
	if err := b.config.MetricsPusher.Validate(); err != nil {
		b.err = errors.NewConfigurationError("metrics_pusher", "validation", fmt.Sprintf("MetricsPusher config validation failed: %v", err))
	}
}

// 🔧 Build - "Build the final configuration"
func (b *ConfigBuilder) Build() (*Config, error) {
	if b.err != nil {
		return nil, b.err
	}

	return b.config, nil
}

// 🔧 loadHTTPFromEnvironment - "Load HTTP configuration from environment"
func (b *ConfigBuilder) loadHTTPFromEnvironment() {
	b.config.HTTP.Port = getEnvInt("HTTP_PORT", constants.PortDefault)
	b.config.HTTP.MetricsPort = getEnvInt("METRICS_PORT", constants.MetricsPortDefault)
	b.config.HTTP.Timeout = getEnvDuration("TIMEOUT", constants.RequestTimeoutDefault)
	b.config.HTTP.APITimeout = getEnvDuration("API_TIMEOUT", constants.APITimeoutDefault)
	b.config.HTTP.MaxRequestSize = int64(getEnvInt("MAX_REQUEST_SIZE", int(constants.MaxRequestSizeDefault)))
	b.config.HTTP.DefaultListLimit = getEnvInt("DEFAULT_LIST_LIMIT", constants.DefaultListLimit)
	b.config.HTTP.MaxListLimit = getEnvInt("MAX_LIST_LIMIT", constants.MaxListLimit)
	b.config.HTTP.ValidateInput = true // Default: always validate input
}

// 🔧 loadKubernetesFromEnvironment - "Load Kubernetes configuration from environment"
func (b *ConfigBuilder) loadKubernetesFromEnvironment() {
	environment := b.config.Environment
	defaultNamespace := fmt.Sprintf("knative-lambda-%s", environment)
	b.config.Kubernetes.Namespace = getEnv("NAMESPACE", defaultNamespace)
	b.config.Kubernetes.InCluster = getEnvBool("IN_CLUSTER", constants.InClusterDefault)
	b.config.Kubernetes.KubeConfig = getEnv("KUBECONFIG", "")
	b.config.Kubernetes.ServiceAccount = getEnv("SERVICE_ACCOUNT", constants.ServiceAccountDefault)
	b.config.Kubernetes.RunAsUser = int64(getEnvInt("RUN_AS_USER", int(constants.K8sRunAsUserDefault)))
	b.config.Kubernetes.JobTTLSeconds = int32(getEnvInt("JOB_TTL_SECONDS", constants.K8sJobTTLSecondsDefault))
	b.config.Kubernetes.JobDeletionWaitTimeout = getEnvDuration("JOB_DELETION_WAIT_TIMEOUT", constants.JobDeletionWaitTimeoutDefault)
	b.config.Kubernetes.JobDeletionCheckInterval = getEnvDuration("JOB_DELETION_CHECK_INTERVAL", constants.JobDeletionCheckIntervalDefault)
}

// 🔧 loadAWSFromEnvironment - "Load AWS configuration from environment"
func (b *ConfigBuilder) loadAWSFromEnvironment() {
	b.config.AWS.AWSRegion = getEnv("AWS_REGION", constants.AWSRegionDefault)
	b.config.AWS.AWSAccountID = getEnv("AWS_ACCOUNT_ID", "")
	b.config.AWS.ECRRegistry = getEnv("ECR_REGISTRY", "")
	b.config.AWS.ECRRepositoryName = getEnv("ECR_REPOSITORY_NAME", constants.AWSECRRepositoryNameDefault)
	b.config.AWS.S3SourceBucket = getEnv("S3_SOURCE_BUCKET", "")
	b.config.AWS.S3TempBucket = getEnv("S3_TEMP_BUCKET", "")
	b.config.AWS.RegistryMirror = getEnv("REGISTRY_MIRROR", constants.RegistryMirrorDefault)
	b.config.AWS.SkipTLSVerifyRegistry = getEnv("SKIP_TLS_VERIFY_REGISTRY", constants.SkipTLSVerifyRegistryDefault)

	b.config.AWS.NodeBaseImage = getEnv("NODE_BASE_IMAGE", constants.NodeBaseImageDefault)
	b.config.AWS.PythonBaseImage = getEnv("PYTHON_BASE_IMAGE", constants.PythonBaseImageDefault)
	b.config.AWS.GoBaseImage = getEnv("GO_BASE_IMAGE", constants.GoBaseImageDefault)
	// EKS Pod Identity Configuration
	b.config.AWS.UseEKSPodIdentity = getEnvBool("USE_EKS_POD_IDENTITY", constants.UseEKSPodIdentityDefault)
	b.config.AWS.PodIdentityRole = getEnv("POD_IDENTITY_ROLE", "")
}

// 🔧 loadObservabilityFromEnvironment - "Load observability configuration from environment"
func (b *ConfigBuilder) loadObservabilityFromEnvironment() {
	// Use K_SERVICE (automatically set by Knative) if available, otherwise fall back to SERVICE_NAME
	knativeServiceName := getEnv("K_SERVICE", "")
	if knativeServiceName != "" {
		b.config.Observability.ServiceName = knativeServiceName
	} else {
		b.config.Observability.ServiceName = getEnv("SERVICE_NAME", constants.ServiceNameDefault)
	}
	b.config.Observability.ServiceVersion = getEnv("SERVICE_VERSION", constants.ServiceVersionDefault)
	b.config.Observability.LogLevel = getEnv("LOG_LEVEL", constants.LogLevelInfo)
	b.config.Observability.OTLPEndpoint = getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	b.config.Observability.TracingEnabled = getEnvBool("TRACING_ENABLED", constants.TracingEnabledDefault)
	b.config.Observability.SampleRate = getEnvFloat("SAMPLE_RATE", constants.SampleRateDefault)
	b.config.Observability.MetricsEnabled = getEnvBool("METRICS_ENABLED", constants.MetricsEnabledDefault)

	// Load exemplars configuration from environment
	b.config.Observability.ExemplarsEnabled = getEnvBool("EXEMPLARS_ENABLED", constants.ExemplarsEnabledDefault)
	b.config.Observability.ExemplarsMaxPerMetric = getEnvInt("EXEMPLARS_MAX_PER_METRIC", constants.ExemplarsMaxPerMetricDefault)
	b.config.Observability.ExemplarsSampleRate = getEnvFloat("EXEMPLARS_SAMPLE_RATE", constants.ExemplarsSampleRateDefault)
	b.config.Observability.ExemplarsTraceIDLabel = getEnv("EXEMPLARS_TRACE_ID_LABEL", constants.ExemplarsTraceIDLabelDefault)
	b.config.Observability.ExemplarsSpanIDLabel = getEnv("EXEMPLARS_SPAN_ID_LABEL", constants.ExemplarsSpanIDLabelDefault)
	b.config.Observability.ExemplarsIncludeLabels = getEnv("EXEMPLARS_INCLUDE_LABELS", constants.ExemplarsIncludeLabelsDefault)
}

// 🔧 loadBuildFromEnvironment - "Load build configuration from environment"
func (b *ConfigBuilder) loadBuildFromEnvironment() {
	b.config.Build.KanikoImage = getEnv("KANIKO_IMAGE", constants.KanikoImageDefault)
	b.config.Build.SidecarImage = getEnv("SIDECAR_IMAGE", "")
	b.config.Build.BuildTimeout = getEnvDuration("BUILD_TIMEOUT", constants.BuildTimeoutDefault)
	b.config.Build.CPURequest = getEnv("CPU_REQUEST", constants.CPURequestDefault)
	b.config.Build.CPULimit = getEnv("CPU_LIMIT", constants.CPULimitDefault)
	b.config.Build.MemoryRequest = getEnv("MEMORY_REQUEST", constants.MemoryRequestDefault)
	b.config.Build.MemoryLimit = getEnv("MEMORY_LIMIT", constants.MemoryLimitDefault)
	b.config.Build.MaxParserSize = int64(getEnvInt("MAX_PARSER_SIZE", int(constants.K8sMaxParserSizeDefault)))
}

// 🔧 loadLambdaFromEnvironment - "Load Lambda configuration from environment"
func (b *ConfigBuilder) loadLambdaFromEnvironment() {
	b.config.Lambda.DefaultRuntime = getEnv("LAMBDA_RUNTIME", constants.RuntimeDefault)
	b.config.Lambda.DefaultHandler = getEnv("LAMBDA_HANDLER", constants.HandlerDefault)
	b.config.Lambda.DefaultTrigger = getEnv("LAMBDA_TRIGGER", constants.TriggerDefault)
	b.config.Lambda.FunctionMemoryLimit = getEnv("LAMBDA_FUNCTION_MEMORY_LIMIT", constants.FunctionMemoryLimitDefault)
	b.config.Lambda.FunctionCPULimit = getEnv("LAMBDA_FUNCTION_CPU_LIMIT", constants.FunctionCPULimitDefault)
	b.config.Lambda.FunctionMemoryRequest = getEnv("LAMBDA_FUNCTION_MEMORY_REQUEST", constants.FunctionMemoryRequestDefault)
	b.config.Lambda.FunctionCPURequest = getEnv("LAMBDA_FUNCTION_CPU_REQUEST", constants.FunctionCPURequestDefault)
	b.config.Lambda.FunctionMemoryLimitMi = getEnv("FUNCTION_MEMORY_LIMIT_MI", constants.FunctionMemoryLimitMiDefault)
	b.config.Lambda.FunctionCPULimitM = getEnv("FUNCTION_CPU_LIMIT_M", constants.FunctionCPULimitMDefault)
}

// 🔧 loadKnativeFromEnvironment - "Load Knative configuration from environment"
func (b *ConfigBuilder) loadKnativeFromEnvironment() {
	b.config.Knative.KnativeTargetConcurrency = getEnv("KUBERNETES_TARGET_CONCURRENCY", constants.BuilderServiceTargetConcurrencyDefault)
	b.config.Knative.KnativeTargetUtilization = getEnv("KUBERNETES_TARGET_UTILIZATION", constants.BuilderServiceTargetUtilizationDefault)
	b.config.Knative.KnativeTarget = getEnv("KUBERNETES_TARGET", constants.BuilderServiceTargetDefault)
	b.config.Knative.KnativeContainerConcurrency = getEnv("KUBERNETES_CONTAINER_CONCURRENCY", constants.BuilderServiceContainerConcurrencyDefault)
	b.config.Knative.KnativeMinScale = getEnv("KUBERNETES_MIN_SCALE", constants.BuilderServiceMinScaleDefault)
	b.config.Knative.KnativeMaxScale = getEnv("KUBERNETES_MAX_SCALE", constants.BuilderServiceMaxScaleDefault)
	b.config.Knative.KnativeScaleToZeroGracePeriod = getEnv("KUBERNETES_SCALE_TO_ZERO_GRACE_PERIOD", constants.BuilderServiceScaleToZeroGracePeriodDefault)
	b.config.Knative.KnativeScaleDownDelay = getEnv("KUBERNETES_SCALE_DOWN_DELAY", constants.BuilderServiceScaleDownDelayDefault)
	b.config.Knative.KnativeStableWindow = getEnv("KUBERNETES_STABLE_WINDOW", constants.BuilderServiceStableWindowDefault)
	b.config.Knative.DefaultEventType = getEnv("DEFAULT_EVENT_TYPE", constants.EventTypeDefault)
	b.config.Knative.DefaultBrokerName = getEnv("BROKER_NAME", constants.BrokerNameDefault)
	b.config.Knative.DefaultTriggerNamespace = getEnv("TRIGGER_NAMESPACE", constants.TriggerNamespaceDefault)
	b.config.Knative.DefaultDeliveryRetries = getEnv("DEFAULT_DELIVERY_RETRIES", constants.DeliveryRetriesDefault)
	b.config.Knative.DefaultDeliveryBackoffPolicy = getEnv("DEFAULT_DELIVERY_BACKOFF_POLICY", constants.DeliveryBackoffPolicyDefault)
	b.config.Knative.DefaultDeliveryBackoffDelay = getEnv("DEFAULT_DELIVERY_BACKOFF_DELAY", constants.DeliveryBackoffDelayDefault)

}

// 🔧 loadSecurityFromEnvironment - "Load security configuration from environment"
func (b *ConfigBuilder) loadSecurityFromEnvironment() {
	b.config.Security.ValidateInput = true // Default: always validate input

	b.config.Security.SecurityEnabled = getEnvBool("SECURITY_ENABLED", constants.SecurityEnabledDefault)
	b.config.Security.DebugMode = getEnvBool("DEBUG_MODE", constants.DebugModeDefault)
	b.config.Security.DryRun = getEnvBool("DRY_RUN", constants.DryRunDefault)
}

// 🔧 loadRateLimitingFromEnvironment - "Load rate limiting configuration from environment variables"
func (b *ConfigBuilder) loadRateLimitingFromEnvironment() {
	if b.err != nil {
		return
	}
	b.config.RateLimiting = &RateLimitingConfig{
		Enabled:                    getEnvBool("RATE_LIMITING_ENABLED", constants.RateLimitingEnabledDefault),
		BuildContextRequestsPerMin: getEnvInt("BUILD_CONTEXT_REQUESTS_PER_MIN", constants.BuildContextRequestsPerMinDefault),
		BuildContextBurstSize:      getEnvInt("BUILD_CONTEXT_BURST_SIZE", constants.BuildContextBurstSizeDefault),
		K8sJobRequestsPerMin:       getEnvInt("K8S_JOB_REQUESTS_PER_MIN", constants.K8sJobRequestsPerMinDefault),
		K8sJobBurstSize:            getEnvInt("K8S_JOB_BURST_SIZE", constants.K8sJobBurstSizeDefault),
		ClientRequestsPerMin:       getEnvInt("CLIENT_REQUESTS_PER_MIN", constants.ClientRequestsPerMinDefault),
		ClientBurstSize:            getEnvInt("CLIENT_BURST_SIZE", constants.ClientBurstSizeDefault),
		S3UploadRequestsPerMin:     getEnvInt("S3_UPLOAD_REQUESTS_PER_MIN", constants.S3UploadRequestsPerMinDefault),
		S3UploadBurstSize:          getEnvInt("S3_UPLOAD_BURST_SIZE", constants.S3UploadBurstSizeDefault),
		MaxMemoryUsagePercent:      getEnvFloat("MAX_MEMORY_USAGE_PERCENT", constants.MaxMemoryUsagePercentDefault),
		MemoryCheckInterval:        getEnvDuration("MEMORY_CHECK_INTERVAL", constants.MemoryCheckIntervalDefault),
		CleanupInterval:            getEnvDuration("CLEANUP_INTERVAL", constants.CleanupIntervalDefault),
		ClientTTL:                  getEnvDuration("CLIENT_TTL", constants.ClientTTLDefault),
	}
}

// 🔧 loadNotifiFromEnvironment - "Load Notifi configuration from environment variables"
func (b *ConfigBuilder) loadNotifiFromEnvironment() {
	if b.err != nil {
		return
	}
	b.config.Notifi.SubscriptionManagerAddress = getEnv("SUBSCRIPTION_MANAGER_ADDRESS", constants.NotifiSubscriptionManagerAddressDefault)
	b.config.Notifi.EphemeralStorageAddress = getEnv("EPHEMERAL_STORAGE_ADDRESS", constants.NotifiEphemeralStorageAddressDefault)
	b.config.Notifi.PersistentStorageAddress = getEnv("PERSISTENT_STORAGE_ADDRESS", constants.NotifiPersistentStorageAddressDefault)
	b.config.Notifi.FusionFetchProxyAddress = getEnv("FUSION_FETCH_PROXY_ADDRESS", constants.NotifiFusionFetchProxyAddressDefault)
	b.config.Notifi.EvmRpcAddress = getEnv("EVM_RPC_ADDRESS", constants.NotifiEvmRpcAddressDefault)
	b.config.Notifi.SolanaRpcAddress = getEnv("SOLANA_RPC_ADDRESS", constants.NotifiSolanaRpcAddressDefault)
	b.config.Notifi.SuiRpcAddress = getEnv("SUI_RPC_ADDRESS", constants.NotifiSuiRpcAddressDefault)
	b.config.Notifi.GrpcInsecure = getEnvBool("GRPC_INSECURE", constants.NotifiGrpcInsecureDefault)
}

// 🔧 loadMetricsPusherFromEnvironment - "Load MetricsPusher configuration from environment variables"
func (b *ConfigBuilder) loadMetricsPusherFromEnvironment() {
	if b.err != nil {
		return
	}
	b.config.MetricsPusher.Enabled = getEnvBool("METRICS_PUSHER_ENABLED", constants.MetricsPusherEnabledDefault)
	b.config.MetricsPusher.ImageRegistry = getEnv("METRICS_PUSHER_IMAGE_REGISTRY", constants.MetricsPusherImageRegistryDefault)
	b.config.MetricsPusher.ImageRepository = getEnv("METRICS_PUSHER_IMAGE_REPOSITORY", constants.MetricsPusherImageRepositoryDefault)
	b.config.MetricsPusher.ImageTag = getEnv("METRICS_PUSHER_IMAGE_TAG", constants.MetricsPusherImageTagDefault)
	b.config.MetricsPusher.ImagePullPolicy = getEnv("METRICS_PUSHER_IMAGE_PULL_POLICY", constants.MetricsPusherImagePullPolicyDefault)
	b.config.MetricsPusher.RemoteWriteURL = getEnv("METRICS_PUSHER_REMOTE_WRITE_URL", constants.MetricsPusherRemoteWriteURLDefault)
	b.config.MetricsPusher.PushInterval = getEnvDuration("METRICS_PUSHER_PUSH_INTERVAL", constants.MetricsPusherPushIntervalDefault)
	b.config.MetricsPusher.Timeout = getEnvDuration("METRICS_PUSHER_TIMEOUT", constants.MetricsPusherTimeoutDefault)
	b.config.MetricsPusher.LogLevel = getEnv("METRICS_PUSHER_LOG_LEVEL", constants.MetricsPusherLogLevelDefault)
	b.config.MetricsPusher.LogFormat = getEnv("METRICS_PUSHER_LOG_FORMAT", constants.MetricsPusherLogFormatDefault)
	b.config.MetricsPusher.QueueProxyMetricsPort = getEnv("METRICS_PUSHER_QUEUE_PROXY_PORT", constants.MetricsPusherQueueProxyPortDefault)
	b.config.MetricsPusher.QueueProxyMetricsPath = getEnv("METRICS_PUSHER_QUEUE_PROXY_PATH", constants.MetricsPusherQueueProxyPathDefault)
	b.config.MetricsPusher.BuilderMetricsPort = getEnv("METRICS_PUSHER_BUILDER_PORT", constants.MetricsPusherBuilderMetricsPortDefault)
	b.config.MetricsPusher.BuilderMetricsPath = getEnv("METRICS_PUSHER_BUILDER_PATH", constants.MetricsPusherBuilderPathDefault)
	b.config.MetricsPusher.FailureTolerance = getEnvInt("METRICS_PUSHER_FAILURE_TOLERANCE", constants.MetricsPusherFailureToleranceDefault)
	b.config.MetricsPusher.ResourceCPURequest = getEnv("METRICS_PUSHER_CPU_REQUEST", constants.MetricsPusherCPURequestDefault)
	b.config.MetricsPusher.ResourceMemoryRequest = getEnv("METRICS_PUSHER_MEMORY_REQUEST", constants.MetricsPusherMemoryRequestDefault)
	b.config.MetricsPusher.ResourceCPULimit = getEnv("METRICS_PUSHER_CPU_LIMIT", constants.MetricsPusherCPULimitDefault)
	b.config.MetricsPusher.ResourceMemoryLimit = getEnv("METRICS_PUSHER_MEMORY_LIMIT", constants.MetricsPusherMemoryLimitDefault)
}

// 🔧 loadPerformanceFromEnvironment - "Load performance configuration from environment variables"
func (b *ConfigBuilder) loadPerformanceFromEnvironment() {
	if b.err != nil {
		return
	}
	// Update the existing RateLimiting config with performance settings
	b.config.RateLimiting.MaxConcurrentBuilds = getEnvInt("MAX_CONCURRENT_BUILDS", constants.MaxConcurrentBuildsDefault)
	b.config.RateLimiting.MaxConcurrentJobs = getEnvInt("MAX_CONCURRENT_JOBS", constants.MaxConcurrentJobsDefault)
	b.config.RateLimiting.BuildTimeout = getEnvDuration("BUILD_TIMEOUT", constants.RateLimitBuildTimeoutDefault)
	b.config.RateLimiting.JobTimeout = getEnvDuration("JOB_TIMEOUT", constants.RateLimitJobTimeoutDefault)
	b.config.RateLimiting.RequestTimeout = getEnvDuration("REQUEST_TIMEOUT", constants.RateLimitRequestTimeoutDefault)
}

// 🔧 loadLambdaServicesFromEnvironment - "Load Lambda services configuration from environment"
func (b *ConfigBuilder) loadLambdaServicesFromEnvironment() {
	if b.err != nil {
		return
	}
	// 🔧 AUTOSCALING: Configuration for lambda service autoscaling
	b.config.LambdaServices.MinScale = getEnv("LAMBDA_SERVICES_MIN_SCALE", constants.LambdaServicesMinScaleDefault)
	b.config.LambdaServices.MaxScale = getEnv("LAMBDA_SERVICES_MAX_SCALE", constants.LambdaServicesMaxScaleDefault)
	b.config.LambdaServices.TargetConcurrency = getEnv("LAMBDA_SERVICES_TARGET_CONCURRENCY", constants.LambdaServicesTargetConcurrencyDefault)
	b.config.LambdaServices.TargetUtilization = getEnv("LAMBDA_SERVICES_TARGET_UTILIZATION", constants.LambdaServicesTargetUtilizationDefault)
	b.config.LambdaServices.Target = getEnv("LAMBDA_SERVICES_TARGET", constants.LambdaServicesTargetDefault)
	b.config.LambdaServices.ContainerConcurrency = getEnv("LAMBDA_SERVICES_CONTAINER_CONCURRENCY", constants.LambdaServicesContainerConcurrencyDefault)
	b.config.LambdaServices.ScaleToZeroGracePeriod = getEnv("LAMBDA_SERVICES_SCALE_TO_ZERO_GRACE_PERIOD", constants.LambdaServicesScaleToZeroGracePeriodDefault)
	b.config.LambdaServices.ScaleDownDelay = getEnv("LAMBDA_SERVICES_SCALE_DOWN_DELAY", constants.LambdaServicesScaleDownDelayDefault)
	b.config.LambdaServices.StableWindow = getEnv("LAMBDA_SERVICES_STABLE_WINDOW", constants.LambdaServicesStableWindowDefault)

	// 🚀 PANIC MODE: For rapid scaling during traffic spikes
	b.config.LambdaServices.PanicWindowPercentage = getEnv("LAMBDA_SERVICES_PANIC_WINDOW_PERCENTAGE", constants.LambdaServicesPanicWindowPercentageDefault)
	b.config.LambdaServices.PanicThresholdPercentage = getEnv("LAMBDA_SERVICES_PANIC_THRESHOLD_PERCENTAGE", constants.LambdaServicesPanicThresholdPercentageDefault)

	// 📦 RESOURCE CONFIGURATION - Specific to lambda services
	b.config.LambdaServices.ResourceMemoryRequest = getEnv("LAMBDA_SERVICES_RESOURCE_MEMORY_REQUEST", constants.LambdaServicesResourceMemoryRequestDefault)
	b.config.LambdaServices.ResourceCPURequest = getEnv("LAMBDA_SERVICES_RESOURCE_CPU_REQUEST", constants.LambdaServicesResourceCPURequestDefault)
	b.config.LambdaServices.ResourceMemoryLimit = getEnv("LAMBDA_SERVICES_RESOURCE_MEMORY_LIMIT", constants.LambdaServicesResourceMemoryLimitDefault)
	b.config.LambdaServices.ResourceCPULimit = getEnv("LAMBDA_SERVICES_RESOURCE_CPU_LIMIT", constants.LambdaServicesResourceCPULimitDefault)

	// ⏰ TIMEOUT CONFIGURATION - Specific to lambda services
	b.config.LambdaServices.TimeoutResponse = getEnv("LAMBDA_SERVICES_TIMEOUT_RESPONSE", constants.LambdaServicesTimeoutResponseDefault)
	b.config.LambdaServices.TimeoutIdle = getEnv("LAMBDA_SERVICES_TIMEOUT_IDLE", constants.LambdaServicesTimeoutIdleDefault)
}

// 🔧 validateLambdaServicesConfig - "Validate Lambda services configuration"
func (b *ConfigBuilder) validateLambdaServicesConfig() {
	if b.err != nil {
		return
	}
	if err := b.config.LambdaServices.Validate(); err != nil {
		b.err = errors.NewConfigurationError("lambda_services", "validation", fmt.Sprintf("Lambda services config validation failed: %v", err))
	}
}

// 🔄 ReloadFromEnvironment - "Reload configuration from environment variables at runtime"
func (b *ConfigBuilder) ReloadFromEnvironment() error {
	b.LoadFromEnvironment()
	b.Validate()
	if b.err != nil {
		return b.err
	}
	return nil
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
