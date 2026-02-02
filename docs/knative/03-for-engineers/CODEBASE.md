# 🔬 Codebase Deep Dive

**Comprehensive guide to the Knative Lambda codebase architecture and internal workings**

---

## 📋 Table of Contents

- [Overview](#overview)
- [Entry Point](#entry-point)
- [Component Architecture](#component-architecture)
- [Core Packages](#core-packages)
- [Data Flow](#data-flow)
- [Key Design Patterns](#key-design-patterns)
- [Testing Strategy](#testing-strategy)

---

## 🎯 Overview

The Knative Lambda codebase is a **Go-based serverless platform** that orchestrates dynamic function builds and deployments on Kubernetes. It follows **clean architecture principles** with clear separation of concerns, dependency injection, and comprehensive observability.

### Architectural Principles

1. **Dependency Injection** - All components are injected via interfaces
2. **Single Responsibility** - Each package has a focused, well-defined purpose
3. **Observability First** - Metrics, logging, and tracing are built into every operation
4. **Error Wrapping** - Consistent error handling with context preservation
5. **Interface Segregation** - Small, focused interfaces over monolithic ones
6. **Graceful Degradation** - Components fail safely without cascading failures

---

## 🚀 Entry Point

### `cmd/service/main.go`

The application entry point orchestrates the entire service lifecycle:

#### Initialization Flow

```12:80:cmd/service/main.go
package main

// ... imports ...

func main() {
	// Create a cancellable context for graceful shutdown handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context is cancelled when main exits

	// Display startup banner with service information
	printStartupBanner()

	// Load and validate all application configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Initialize observability system (metrics, logging, tracing, system metrics)
	obs, err := initializeObservability(ctx, cfg)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize observability")
	}
	defer shutdownObservability(ctx, obs) // Ensure observability is shut down gracefully

	// Initialize and run the main service with all components
	errors.InitializeService(ctx, obs, cfg.Observability.ServiceName, func() error {
		return initializeAndRunService(ctx, obs, cfg, cancel)
	})
}
```

#### Component Initialization Hierarchy

```
main()
  ├─ Initialize Observability (metrics, tracing, logging)
  ├─ Initialize Infrastructure Components
  │   ├─ Kubernetes Client
  │   ├─ AWS Client
  │   ├─ Redis Client
  │   └─ Rate Limiter
  ├─ Initialize Handler Components
  │   ├─ Kubernetes Handlers
  │   │   ├─ Job Manager
  │   │   └─ Service Manager
  │   ├─ AWS Handlers
  │   │   └─ Build Context Manager
  │   └─ Event Handler
  │       └─ HTTP Handler
  └─ Start HTTP Server
```

#### Key Responsibilities

- **Configuration Loading** - Environment variable parsing and validation
- **Component Wiring** - Dependency injection and component initialization
- **Graceful Shutdown** - Signal handling and clean resource cleanup
- **Error Recovery** - Panic recovery and observability integration

---

## 🏗️ Component Architecture

### Component Container Pattern

The system uses a **dependency injection container** to manage all service components:

```18:46:internal/handler/container.go
// 🎯 ComponentContainerImpl - "Dependency injection container implementation"
type ComponentContainerImpl struct {
	// 🔧 Core Dependencies
	config      *config.Config
	obs         *observability.Observability
	redisClient *redisclient.Client

	// 🎯 HTTP Components
	httpHandler       HTTPHandler
	cloudEventHandler CloudEventHandler

	// 🎯 Job Management Components
	jobManager      JobManager
	asyncJobCreator AsyncJobCreatorInterface

	// 🎯 Event Processing Components
	eventHandler EventHandler

	// 🎯 Service Management Components
	serviceManager ServiceManager

	// 🎯 Build Context Components
	buildContextManager BuildContextManager

	// 🚦 Rate Limiting
	rateLimiter *resilience.MultiLevelRateLimiter

	// 🔒 Thread Safety
	mu sync.RWMutex
}
```

**Benefits:**
- ✅ Centralized component management
- ✅ Thread-safe access via `sync.RWMutex`
- ✅ Clear dependency graph
- ✅ Easy testing via interface mocking

---

## 📦 Core Packages

### 1️⃣ `internal/config` - Configuration Management

**Purpose:** Centralized configuration with environment variable loading and validation.

#### Key Files

| File | Purpose |
|------|---------|
| `config.go` | Main config struct and validation |
| `aws.go` | AWS service configuration |
| `http.go` | HTTP server configuration |
| `kubernetes.go` | Kubernetes client configuration |
| `build.go` | Build process configuration |
| `observability.go` | Metrics, logging, tracing configuration |

#### Configuration Loading

```23:87:internal/config/config.go
// Config represents the complete configuration for the Knative Lambda service.
//
// The configuration is organized into logical sections for different service components:
// - HTTP: Server settings, ports, timeouts
// - Kubernetes: Namespace, client configuration, RBAC
// - AWS: ECR registry, S3 buckets, IAM settings
// - RabbitMQ: Connection strings, exchanges, queues
// - Observability: Metrics, tracing, logging
// - Security: Authentication, authorization, validation
// - Build: Kaniko settings, timeouts, resource limits
// - Lambda: Function-specific settings
// - LambdaServices: Configuration for dynamically created lambda services
// - Knative: Eventing configuration
// - MetricsPusher: Configuration for metrics-pusher sidecar
// - Features: Feature flags and toggles
// - Performance: Performance tuning parameters
type Config struct {
	// Environment specifies the deployment environment (dev, prd, local)
	Environment string `envconfig:"ENVIRONMENT" default:"dev" validate:"required,oneof=dev prd local"`

	// Component configurations
	HTTP       *HTTPConfig       `json:"http"`
	Kubernetes *KubernetesConfig `json:"kubernetes"`
	AWS        *AWSConfig        `json:"aws"`
	Storage    *StorageConfig    `json:"storage"`
	Redis      *RedisConfig      `json:"redis"`

	Observability  *ObservabilityConfig  `json:"observability"`
	Build          *BuildConfig          `json:"build"`
	Lambda         *LambdaConfig         `json:"lambda"`
	LambdaServices *LambdaServicesConfig `json:"lambda_services"`
	Knative        *KnativeConfig        `json:"knative"`
	Security       *SecurityConfig       `json:"security"`
	RateLimiting   *RateLimitingConfig   `json:"rate_limiting"`
	Notifi         *NotifiConfig         `json:"notifi"`
	MetricsPusher  *MetricsPusherConfig  `json:"metrics_pusher"`
}

// LoadConfig loads and validates the service configuration using the ConfigBuilder pattern.
// Returns a fully validated Config struct or an error if validation fails.
func LoadConfig() (*Config, error) {
	cfg, err := NewConfigBuilder().
		WithEnvironment(getEnv("ENVIRONMENT", constants.EnvironmentDev)).
		LoadFromEnvironment().
		Validate().
		Build()

	if err != nil {
		return nil, errors.NewConfigurationError("config", "validation", fmt.Sprintf("configuration validation failed: %v", err))
	}

	return cfg, nil
}
```

**Features:**
- 🔧 Builder pattern for configuration construction
- ✅ Comprehensive validation on load
- 🔍 Environment-aware defaults
- 📝 Structured configuration sections

---

### 2️⃣ `internal/handler` - Event Processing

**Purpose:** Core event processing logic and business rules.

#### Component Hierarchy

```
EventHandler (Main Orchestrator)
  ├─ CloudEventHandler (CloudEvents processing)
  ├─ BuildContextManager (Build context creation)
  ├─ JobManager (Kubernetes job lifecycle)
  ├─ ServiceManager (Knative service lifecycle)
  ├─ AsyncJobCreator (Parallel job creation)
  └─ HTTPHandler (HTTP server)
```

#### Event Handler

```25:92:internal/handler/event_handler.go
// 🎯 EventHandlerImpl - "Composed handler with focused components"
type EventHandlerImpl struct {
	// 🎯 Dependency Injection Container - Centralized component management
	container ComponentContainer

	// 🔧 Shared Dependencies
	config *config.Config
	obs    *observability.Observability
}

// 🎯 EventHandlerConfig - "Configuration for creating event handler"
type EventHandlerConfig struct {
	Container ComponentContainer
}

// 🏗️ NewEventHandler - "Create new event handler with composed components"
func NewEventHandler(config EventHandlerConfig) (*EventHandlerImpl, error) {
	if config.Container == nil {
		return nil, errors.NewConfigurationError("event_handler", "container", "container cannot be nil")
	}

	// Validate all components are initialized
	if err := config.Container.(*ComponentContainerImpl).ValidateComponents(); err != nil {
		return nil, errors.NewConfigurationError("event_handler", "container", fmt.Sprintf("container validation failed: %v", err))
	}

	return &EventHandlerImpl{
		container: config.Container,
		config:    config.Container.GetConfig(),
		obs:       config.Container.GetObservability(),
	}, nil
}

// 📥 ProcessCloudEvent - "Process CloudEvent with comprehensive observability"
func (h *EventHandlerImpl) ProcessCloudEvent(ctx context.Context, event *cloudevents.Event) (*builds.HandlerResponse, error) {
	// Create metrics recorder for this request
	metricsRec := observability.NewMetricsRecorder(h.obs)

	// Start distributed tracing span
	ctx, span := h.obs.StartSpanWithAttributes(ctx, "process_cloud_event", map[string]string{
		"event.type":   event.Type(),
		"event.source": event.Source(),
		"event.id":     event.ID(),
	})
	defer span.End()

	// Validate event using internal validation
	if err := h.ValidateEvent(ctx, event); err != nil {
		metricsRec.RecordError(ctx, "event_handler", "validation_error", "error")
		return nil, err
	}

	// Record build request metric for start events
	if h.isBuildStartEvent(event) {
		if buildRequest, err := h.ParseBuildRequest(ctx, event); err == nil {
			metricsRec.RecordBuildRequest(ctx, buildRequest.ThirdPartyID, buildRequest.ParserID, "received")
		}
	}

	// Process event with comprehensive tracing
	response, err := h.processEventWithTracing(ctx, event, metricsRec)
	if err != nil {
		metricsRec.RecordError(ctx, "event_handler", "processing_error", "error")
		return nil, err
	}

	return response, nil
}
```

**Responsibilities:**
- 📥 CloudEvent validation and parsing
- 🔄 Event type routing (`build.start`, `job.start`, `service.delete`)
- 📊 Metrics and tracing integration
- ⚡ Async job creation coordination
- 🛡️ Error handling and recovery

#### Job Manager

```28:120:internal/handler/job_manager.go
// 🔄 JobManagerImpl - "Focused Kubernetes job lifecycle management"
type JobManagerImpl struct {
	k8sClient       kubernetes.Interface
	config          *config.KubernetesConfig
	buildConfig     *config.BuildConfig
	awsConfig       *config.AWSConfig
	rateLimitConfig *config.RateLimitingConfig
	obs             *observability.Observability
	// 🛡️ Rate Limiting Protection
	rateLimiter *resilience.MultiLevelRateLimiter
}

// 🔄 JobManagerConfig - "Configuration for creating job manager"
type JobManagerConfig struct {
	K8sClient       kubernetes.Interface
	K8sConfig       *config.KubernetesConfig
	BuildConfig     *config.BuildConfig
	AWSConfig       *config.AWSConfig
	RateLimitConfig *config.RateLimitingConfig
	Observability   *observability.Observability
	RateLimiter     *resilience.MultiLevelRateLimiter
}

// 🏗️ NewJobManager - "Create new job manager with dependencies"
func NewJobManager(config JobManagerConfig) (JobManager, error) {
	if config.K8sClient == nil {
		return nil, errors.NewConfigurationError("job_manager", "k8s_client", "kubernetes client cannot be nil")
	}

	if config.K8sConfig == nil {
		return nil, errors.NewConfigurationError("job_manager", "k8s_config", "kubernetes config cannot be nil")
	}

	if config.BuildConfig == nil {
		return nil, errors.NewConfigurationError("job_manager", "build_config", "build config cannot be nil")
	}

	if config.Observability == nil {
		return nil, errors.NewConfigurationError("job_manager", "observability", "observability cannot be nil")
	}

	return &JobManagerImpl{
		k8sClient:       config.K8sClient,
		config:          config.K8sConfig,
		buildConfig:     config.BuildConfig,
		awsConfig:       config.AWSConfig,
		rateLimitConfig: config.RateLimitConfig,
		obs:             config.Observability,
		rateLimiter:     config.RateLimiter,
	}, nil
}

// 🔄 CreateJob - "Create a new Kubernetes job (KISS approach)"
func (j *JobManagerImpl) CreateJob(ctx context.Context, jobName string, buildRequest *builds.BuildRequest) (*batchv1.Job, error) {
	ctx, span := j.obs.StartSpan(ctx, "create_job")
	defer span.End()

	j.obs.Info(ctx, "Starting Kubernetes job creation",
		"job_name", jobName,
		"third_party_id", buildRequest.ThirdPartyID,
		"parser_id", buildRequest.ParserID,
		"correlation_id", buildRequest.CorrelationID,
		"build_type", buildRequest.BuildType,
		"runtime", buildRequest.Runtime)

	// 🚀 KISS: Delete existing job if it exists, then create new one
	existingJob, err := j.FindExistingJob(ctx, buildRequest.ThirdPartyID, buildRequest.ParserID)
	if err != nil {
		j.obs.Info(ctx, "Failed to check for existing job, continuing with creation",
			"job_name", jobName,
			"third_party_id", buildRequest.ThirdPartyID,
			"parser_id", buildRequest.ParserID,
			"correlation_id", buildRequest.CorrelationID)
```

**Key Operations:**
- 🚀 Job creation with Kaniko configuration
- 🔍 Job status checking and monitoring
- 🗑️ Job cleanup and failed job handling
- 🛡️ Rate limiting for K8s API calls
- 📊 Metrics for job lifecycle events

#### Service Manager

```30:100:internal/handler/service_manager.go
// 🚀 ServiceManagerImpl - "Focused Knative service lifecycle management"
type ServiceManagerImpl struct {
	k8sClient     kubernetes.Interface
	dynamicClient dynamic.Interface
	config        *config.KubernetesConfig
	obs           *observability.Observability
	// 🛡️ Rate Limiting Protection
	rateLimiter *resilience.MultiLevelRateLimiter
	// 🎯 Knative Configuration (for main builder service)
	knativeConfig *config.KnativeConfig
	// 🚀 Lambda Services Configuration (for dynamically created lambda services)
	lambdaServicesConfig *config.LambdaServicesConfig
	// 🔗 Notifi Configuration
	notifiConfig *config.NotifiConfig
	// 📊 Metrics Pusher Configuration
	metricsPusherConfig *config.MetricsPusherConfig
}

// 🚀 ServiceManagerConfig - "Configuration for creating service manager"
type ServiceManagerConfig struct {
	K8sClient            kubernetes.Interface
	DynamicClient        dynamic.Interface
	K8sConfig            *config.KubernetesConfig
	Observability        *observability.Observability
	RateLimiter          *resilience.MultiLevelRateLimiter
	KnativeConfig        *config.KnativeConfig
	LambdaServicesConfig *config.LambdaServicesConfig
	NotifiConfig         *config.NotifiConfig
	MetricsPusherConfig  *config.MetricsPusherConfig
}

// 🏗️ NewServiceManager - "Create new service manager with dependencies"
func NewServiceManager(config ServiceManagerConfig) (ServiceManager, error) {
	if config.K8sClient == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "k8s_client", "kubernetes client cannot be nil")
	}

	if config.DynamicClient == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "dynamic_client", "dynamic client cannot be nil")
	}

	if config.K8sConfig == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "k8s_config", "kubernetes config cannot be nil")
	}

	if config.Observability == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "observability", "observability cannot be nil")
	}

	if config.KnativeConfig == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "knative_config", "knative config cannot be nil")
	}

	if config.LambdaServicesConfig == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "lambda_services_config", "lambda services config cannot be nil")
	}

	if config.NotifiConfig == nil {
		return nil, internalerrors.NewConfigurationError("service_manager", "notifi_config", "notifi config cannot be nil")
	}

	// MetricsPusher config is optional - can be nil

	return &ServiceManagerImpl{
		k8sClient:            config.K8sClient,
		dynamicClient:        config.DynamicClient,
		config:               config.K8sConfig,
		obs:                  config.Observability,
		rateLimiter:          config.RateLimiter,
		knativeConfig:        config.KnativeConfig,
		lambdaServicesConfig: config.LambdaServicesConfig,
```

**Responsibilities:**
- 🚀 Knative Service creation and management
- 📦 ConfigMap and ServiceAccount management
- 🔗 Trigger creation for event routing
- 🗑️ Service deletion and cleanup
- 📊 Kubernetes resource monitoring

#### Build Context Manager

```29:86:internal/handler/build_context_manager.go
// 📦 BuildContextManagerImpl - "Focused build context and archive management"
type BuildContextManagerImpl struct {
	awsClient *aws.Client
	config    *config.Config
	obs       *observability.Observability
	// 🛡️ Rate Limiting Protection
	rateLimiter *resilience.MultiLevelRateLimiter
	// 📄 Template Processing
	templateProcessor *templates.TemplateProcessor
}

// 📦 BuildContextManagerConfig - "Configuration for creating build context manager"
type BuildContextManagerConfig struct {
	Storage       interface{} // Storage client (implements storage.Storage interface)
	Config        *config.Config
	Observability *observability.Observability
	RateLimiter   *resilience.MultiLevelRateLimiter
}

// 🏗️ NewBuildContextManager - "Create new build context manager with dependencies"
func NewBuildContextManager(config BuildContextManagerConfig) (BuildContextManager, error) {
	if config.Storage == nil {
		return nil, errors.NewConfigurationError("build_context_manager", "storage", "storage client cannot be nil")
	}

	if config.Config == nil {
		return nil, errors.NewConfigurationError("build_context_manager", "config", "config cannot be nil")
	}

	if config.Observability == nil {
		return nil, errors.NewConfigurationError("build_context_manager", "observability", "observability cannot be nil")
	}

	// Create AWS client for ECR operations
	awsClient, err := aws.NewClient(context.Background(), aws.ClientConfig{
		Region:            config.Config.AWS.GetRegion(),
		AccountID:         config.Config.AWS.GetAccountID(),
		ECRRegistry:       config.Config.AWS.GetECRRegistry(),
		ECRRepositoryName: config.Config.AWS.GetECRRepositoryName(),
		S3SourceBucket:    config.Config.AWS.GetS3SourceBucket(),
		S3TempBucket:      config.Config.AWS.GetS3TempBucket(),
		Observability:     config.Observability,
	})
	if err != nil {
		return nil, errors.WrapWithContext(err, "failed to create AWS client")
	}

	// Initialize template processor
	templateProcessor := templates.NewTemplateProcessor(config.Observability)

	return &BuildContextManagerImpl{
		awsClient:         awsClient,
		config:            config.Config,
		obs:               config.Observability,
		rateLimiter:       config.RateLimiter,
		templateProcessor: templateProcessor,
	}, nil
}
```

**Key Operations:**
- 📦 Build context creation (tar.gz archives)
- 📄 Dynamically generated Dockerfiles
- ☁️ S3 upload for build contexts
- ✅ Request validation
- 🧮 Archive checksums for idempotency

---

### 3️⃣ `internal/observability` - Monitoring & Tracing

**Purpose:** Unified observability with metrics, logging, and distributed tracing.

#### Core Components

- **Metrics** - Prometheus metrics integration
- **Logging** - Structured JSON logging with Logrus
- **Tracing** - OpenTelemetry distributed tracing
- **Exemplars** - Trace exemplar linking to metrics

**Features:**
- 📊 Automatic request/response metrics
- 🔍 Distributed tracing with context propagation
- 📝 Structured logging with correlation IDs
- ⏱️ Performance monitoring and profiling

---

### 4️⃣ `internal/aws` - AWS Integration

**Purpose:** AWS service clients (S3, ECR, IAM).

#### Key Operations

- **S3** - Source code upload and download
- **ECR** - Container image registry access
- **IAM** - Credential management
- **CloudWatch** - Metrics and logging integration

---

### 5️⃣ `internal/resilience` - Rate Limiting

**Purpose:** Multi-level rate limiting with Redis support.

**Rate Limiting Levels:**
1. **Build Context** - S3 upload rate limiting
2. **K8s Jobs** - Kubernetes API rate limiting
3. **Client Requests** - HTTP request rate limiting
4. **S3 Uploads** - S3 API rate limiting

**Features:**
- 🛡️ Token bucket algorithm
- 🔴 Redis-backed distributed rate limiting
- 📊 Metrics for rate limit violations
- 🧹 Automatic cleanup of expired entries

---

### 6️⃣ `internal/storage` - Storage Abstraction

**Purpose:** Pluggable storage backend (S3, MinIO).

**Storage Interface:**
- `GetObject()` - Download files
- `PutObject()` - Upload files
- `DeleteObject()` - Delete files
- `ListObjects()` - List directory contents

**Supported Backends:**
- ☁️ AWS S3 (production)
- 🏠 MinIO (local development)

---

### 7️⃣ `internal/redis` - Caching & State

**Purpose:** Redis client for caching, rate limiting, and state management.

**Features:**
- 🔗 Connection pooling
- ❤️ Health checks
- 🔄 Automatic reconnection
- 📊 Connection metrics

---

## 🔄 Data Flow

### Build Start Event Flow

```
CloudEvent (build.start)
  ↓
HTTPHandler (Receive & Parse)
  ↓
CloudEventHandler (Validate)
  ↓
EventHandler (Process)
  ↓
├─→ BuildContextManager (Create build context)
│     ├─ Fetch parser files from S3
│     ├─ Generate Dockerfile
│     ├─ Create tar.gz archive
│     └─ Upload to S3
  ↓
JobManager (Create Kaniko job)
  ↓
Kubernetes API (Job Creation)
  ↓
Kaniko Pod (Container build)
  ├─ Pull source from S3
  ├─ Build container image
  └─ Push to ECR
  ↓
CloudEvent (build.complete)
  ↓
ServiceManager (Create Knative Service)
  ├─ ConfigMap
  ├─ ServiceAccount
  ├─ Knative Service
  └─ Trigger
  ↓
Function Ready
```

### Request Processing Flow

```29:97:internal/handler/http_handler.go
// 🌐 HTTPHandlerImpl - "Focused HTTP server management and routing"
type HTTPHandlerImpl struct {
	config      *config.HTTPConfig
	obs         *observability.Observability
	router      *chi.Mux
	middleware  http.Handler
	server      *http.Server
	rateLimiter *resilience.RateLimiter
	container   ComponentContainer
}

// 🌐 HTTPHandlerConfig - "Configuration for creating HTTP handler"
type HTTPHandlerConfig struct {
	Config        *config.HTTPConfig
	Observability *observability.Observability
	Container     ComponentContainer
	RateLimiter   *resilience.RateLimiter
}

// 🏗️ NewHTTPHandler - "Create new HTTP handler with dependencies"
func NewHTTPHandler(config HTTPHandlerConfig) (HTTPHandler, error) {
	if config.Config == nil {
		return nil, errors.NewConfigurationError("http_config", "config", "config cannot be nil")
	}

	if config.Observability == nil {
		return nil, errors.NewConfigurationError("observability", "observability", "observability cannot be nil")
	}

	router := chi.NewRouter()

	// Use provided rate limiter or create a new one if not provided
	var rateLimiter *resilience.RateLimiter

	if config.RateLimiter != nil {
		rateLimiter = config.RateLimiter
	} else {
		// Create a simple rate limiter
		rateLimiter = resilience.NewRateLimiter(10, 5) // 10 requests per minute, burst of 5
	}

	// Create middleware chain with observability and rate limiting
	middlewareChain := CreateDefaultMiddlewareChain(config.Observability, rateLimiter)
	middlewareHandler := middlewareChain(router)

	server := &http.Server{
		Addr:              config.Config.GetServerAddress(),
		Handler:           middlewareHandler,
		ReadTimeout:       constants.HTTPReadTimeoutDefault,
		WriteTimeout:      config.Config.APITimeout + constants.HTTPWriteTimeoutOffset,
		IdleTimeout:       constants.HTTPIdleTimeoutDefault,
		ReadHeaderTimeout: constants.HTTPReadHeaderTimeoutDefault,
	}

	handler := &HTTPHandlerImpl{
		config:      config.Config,
		obs:         config.Observability,
		router:      router,
		middleware:  router,
		server:      server,
		rateLimiter: rateLimiter,
		container:   config.Container,
	}

	// Register routes on the router
	handler.RegisterRoutes(nil)

	return handler, nil
}
```

---

## 🎨 Key Design Patterns

### 1️⃣ Dependency Injection

All components receive dependencies via constructor parameters:

```go
// Good - Dependencies injected
func NewJobManager(config JobManagerConfig) (JobManager, error) {
    return &JobManagerImpl{
        k8sClient: config.K8sClient,
        obs:       config.Observability,
        ...
    }, nil
}
```

### 2️⃣ Interface Segregation

Small, focused interfaces over large ones:

```go
type JobCreator interface {
    CreateJob(...) (*batchv1.Job, error)
}

type JobFinder interface {
    FindExistingJob(...) (*batchv1.Job, error)
}

type JobManager interface {
    JobCreator
    JobFinder
    JobStatusChecker
    DeleteJob(...) error
}
```

### 3️⃣ Builder Pattern

Used for configuration construction:

```go
cfg, err := NewConfigBuilder().
    WithEnvironment("dev").
    LoadFromEnvironment().
    Validate().
    Build()
```

### 4️⃣ Observability Wrapper

Every operation wrapped with metrics and tracing:

```go
func (h *EventHandlerImpl) ProcessCloudEvent(...) {
    metricsRec := observability.NewMetricsRecorder(h.obs)
    ctx, span := h.obs.StartSpan(...)
    defer span.End()
    
    // ... operation ...
    
    metricsRec.RecordSuccess(...)
}
```

### 5️⃣ Error Wrapping

Consistent error handling with context:

```go
return nil, errors.WrapWithContext(err, "failed to create job", "job_name", jobName)
```

---

## 🧪 Testing Strategy

### Unit Tests

**Location:** `internal/*/*_test.go`

**Coverage:**
- ✅ All public functions
- ✅ Error handling paths
- ✅ Edge cases
- ✅ Interface contracts

**Example:**
```go
func TestJobManager_CreateJob(t *testing.T) {
    // Mock Kubernetes client
    mockClient := &MockK8sClient{}
    
    // Create job manager
    manager := NewJobManager(JobManagerConfig{
        K8sClient: mockClient,
        ...
    })
    
    // Test job creation
    job, err := manager.CreateJob(...)
    assert.NoError(t, err)
    assert.NotNil(t, job)
}
```

### Integration Tests

**Location:** `internal/handler/*_test.go`

**Coverage:**
- ✅ Component interactions
- ✅ End-to-end event processing
- ✅ Error propagation
- ✅ Resource cleanup

### E2E Tests

**Location:** `tests/e2e/`

**Coverage:**
- ✅ Full CloudEvent processing pipeline
- ✅ Kubernetes resource lifecycle
- ✅ Multi-environment deployment
- ✅ Load and performance testing

---

## 📚 Additional Resources

- [Getting Started Guide](README.md) - Quick start and installation
- [Architecture Documentation](../04-architecture/README.md) - System design
- [SRE Guide](sre/README.md) - Operations and troubleshooting
- [Backend Guide](backend/README.md) - Development guide
- [DevOps Guide](devops/README.md) - Deployment guide
- [Security Guide](security/README.md) - Security best practices

---

**Last Updated:** 2025-01-23  
**Maintained By:** Platform Team

