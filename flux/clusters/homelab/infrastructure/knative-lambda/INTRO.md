# Knative Lambda Builder Service

## Overview

This service implements a **RabbitMQ CloudEvents Processing Pattern** for building container images using Kaniko and creating Knative services. It processes CloudEvents from RabbitMQ to orchestrate the complete build-to-deploy lifecycle.

## 🔄 How It Works

### 1. RabbitMQ CloudEvent Reception
The service receives CloudEvents from RabbitMQ with two main event types:
- `network.notifi.lambda.build.start` - Initiates a new build
- `network.notifi.lambda.build.complete` - Handles build completion

### 2. Build Start Event Processing
When a `network.notifi.lambda.build.start` event is received:
- **Parse & Validate**: Extract build request data from the CloudEvent
- **Create Build Context**: Generate and upload build context to S3
- **Create Kaniko Job**: Deploy a Kubernetes Job with Kaniko executor
- **Fire-and-Forget**: Return immediately after job creation, allowing the service to scale down

### 3. Build Complete Event Processing
When a `network.notifi.lambda.build.complete` event is received:
- **Parse Completion Data**: Extract build results and metadata
- **Validate Success**: Only proceed for successful builds (`status == "success"`)
- **Create Knative Service**: Deploy the built container as a Knative service
- **Create Knative Trigger**: Set up event triggers for the new service

### 4. Sidecar-Based Build Monitoring
The Kaniko Job runs with a **sidecar container** that:
- Monitors the Kaniko build process completion
- Detects success or failure by watching the Kaniko container's exit code
- Automatically emits `network.notifi.lambda.build.complete` or `network.notifi.lambda.build.failed` events to RabbitMQ
- Handles cleanup and resource management

## ✅ Why This Pattern Is Correct

### **RabbitMQ CloudEvents Integration**
- **Event-Driven Architecture**: Uses RabbitMQ as the event backbone for reliable message delivery
- **Decoupled Processing**: Build requests and completions are handled as separate events
- **Scalable Event Processing**: Can handle high volumes of build requests without blocking

### **Scalability & Efficiency**
- **Stateless Processing**: Each event is processed independently
- **Scale-to-Zero**: Service can scale down when not processing events
- **Resource Optimization**: Build context creation and job deployment are optimized for efficiency

### **Complete Lifecycle Management**
- **Build-to-Deploy Pipeline**: Handles the complete journey from build request to deployed service
- **Automatic Service Creation**: Successful builds automatically create Knative services and triggers
- **Error Handling**: Failed builds are handled gracefully without service creation

### **Kubernetes-Native**
- **Kaniko Integration**: Uses Kaniko for secure, efficient container builds
- **Knative Serving**: Leverages Knative for serverless service deployment
- **Event-Driven Triggers**: Uses Knative eventing for service activation

## 🔄 Event Flow Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   RabbitMQ      │    │  Knative Lambda  │    │   Kubernetes    │
│   CloudEvents   │    │     Service      │    │     Cluster     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         │ 1. build.start        │                       │
         │──────────────────────▶│                       │
         │                       │ 2. Create S3 Context  │
         │                       │──────────────────────▶│
         │                       │ 3. Create Kaniko Job  │
         │                       │──────────────────────▶│
         │                       │                       │
         │                       │ 4. Job Completes      │
         │                       │◀──────────────────────│
         │ 5. build.complete     │                       │
         │◀──────────────────────│                       │
         │                       │ 6. Create Service     │
         │                       │──────────────────────▶│
         │                       │ 7. Create Trigger     │
         │                       │──────────────────────▶│
```

### **Event Types**

#### **Build Start Event** (`network.notifi.lambda.build.start`)
```json
{
  "type": "network.notifi.lambda.build.start",
  "source": "rabbitmq",
  "data": {
    "third_party_id": "example-org",
    "parser_id": "user-parser",
    "build_type": "lambda",
    "runtime": "nodejs22",
    "source_url": "https://github.com/example/parser",
    "build_timeout": 1800
  }
}
```

#### **Build Complete Event** (`network.notifi.lambda.build.complete`)
```json
{
  "type": "network.notifi.lambda.build.complete",
  "source": "rabbitmq",
  "data": {
    "third_party_id": "example-org",
    "parser_id": "user-parser",
    "job_name": "build-example-org-user-parser-abc123",
    "status": "success",
    "image_uri": "123456789.dkr.ecr.us-west-2.amazonaws.com/parser:abc123",
    "duration": "2m30s",
    "correlation_id": "corr-123"
  }
}
```

## 📁 Project Structure

```
├── deploy/
│   └── overlays/
│   └── templates/
│       ├── apisource.yaml
│       ├── brokers.yaml
│       ├── builder.yaml
│       ├── namespace.yaml
│       ├── queues.yaml
│       ├── secrets.yaml
│       ├── serviceaccount.yaml
│       ├── servicemonitor.yaml
│       └── source.yaml
│       └── triggers.yaml
├── cmd/
│   └── service/
│       ├── main.go
│       └── templates/
│           ├── Dockerfile.tpl
│           ├── func.yaml.tpl
│           ├── kaniko-job.yaml.tpl
│           ├── service.yaml.tpl
│           ├── package.json.tpl
│           └── index.js.tpl
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── aws/
│   │   ├── s3.go
│   │   └── ecr.go
│   ├── k8s/
│   │   ├── client.go
│   │   └── job_builder.go
│   ├── handler/
│   │   ├── event_handler.go          # Main event orchestration
│   │   ├── event_processor.go        # Event parsing and validation
│   │   ├── service_manager.go        # Knative service creation
│   │   ├── job_manager.go            # Kaniko job management
│   │   ├── build_context_manager.go  # S3 context creation
│   │   └── interfaces.go             # Component interfaces
│   ├── observability/
│   │   ├── metrics.go
│   │   ├── tracing.go
│   │   └── logging.go
│   ├── model/
│   │   ├── cloudevent.go
│   │   └── build_request.go
│   ├── security/
│   │   └── validation.go
│   └── sidecar/
│       ├── monitor.go
│       ├── publisher.go
│       └── cleanup.go
├── pkg/
│   └── builds/
│       ├── types.go                  # Build data structures
│       └── request.go                # Build request models
├── sidecar/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── watcher/
│   │   │   └── process_watcher.go
│   │   ├── events/
│   │   │   └── event_publisher.go
│   │   └── config/
│   │       └── sidecar_config.go
│   └── Dockerfile
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## 📝 Directory Explanation

### **`cmd/service/main.go`**
Entry point of the Knative service that:
- Initializes configuration and component container
- Sets up the CloudEvents receiver for RabbitMQ events
- Registers the event handler with dependency injection
- Starts the HTTP server with health checks

### **`internal/handler/`**
Contains the core event processing logic:

#### **`event_handler.go`**
Main orchestration component that:
- Processes RabbitMQ CloudEvents based on event type
- Routes build start events to job creation flow
- Routes build complete events to service creation flow
- Manages component composition and dependency injection

#### **`event_processor.go`**
Focused event processing that:
- Parses and validates CloudEvent payloads
- Handles build request parsing for start events
- Handles build completion parsing for complete events
- Performs security validation on event data

#### **`service_manager.go`**
Knative service lifecycle management:
- Creates Knative services from successful builds
- Creates Knative triggers for event handling
- Manages service accounts, config maps, and monitoring
- Handles service existence checks and updates

#### **`job_manager.go`**
Kaniko job management:
- Creates Kubernetes jobs with Kaniko executor
- Manages job lifecycle and cleanup
- Handles job status checking and monitoring
- Generates unique job names and configurations

#### **`build_context_manager.go`**
S3 build context management:
- Creates build contexts from source code
- Uploads contexts to S3 for Kaniko access
- Manages context validation and cleanup
- Handles S3 bucket and key generation

#### **`interfaces.go`**
Component interface definitions:
- Defines contracts for all major components
- Enables dependency injection and testing
- Provides clear separation of concerns
- Supports component composition patterns

### **`pkg/builds/`**
Reusable build-related data structures:

#### **`types.go`**
Defines build event data structures:
- `BuildEventData` for build start events
- `BuildCompletionEventData` for build complete events
- `HandlerResponse` for service responses
- `BuildJob` for job status tracking

#### **`request.go`**
Build request models:
- `BuildRequest` for comprehensive build configuration
- `BuildConfig` for build-specific settings
- Validation helpers for build requests

### **`sidecar/`**
Contains the sidecar container application that monitors Kaniko builds and publishes completion events to RabbitMQ.

## 🚀 Code Flow

### 1. **RabbitMQ Event Reception**
The service receives CloudEvents from RabbitMQ and routes them based on event type.

### 2. **Build Start Event Processing**
```
RabbitMQ Event → Event Handler → Event Processor → Build Context Manager → Job Manager
     ↓              ↓                ↓                    ↓                ↓
build.start    Parse Request    Validate Data    Create S3 Context    Create Kaniko Job
```

### 3. **Build Complete Event Processing**
```
RabbitMQ Event → Event Handler → Event Processor → Service Manager
     ↓              ↓                ↓                ↓
build.complete  Parse Completion  Validate Success  Create Knative Service + Trigger
```

### 4. **Component Composition**
The event handler uses dependency injection to compose focused components:
- **EventProcessor**: Handles event parsing and validation
- **BuildContextManager**: Manages S3 context creation
- **JobManager**: Manages Kaniko job lifecycle
- **ServiceManager**: Manages Knative service creation
- **Observability**: Provides metrics, tracing, and logging

## 🔧 Implementation Details

### **Event Handler Architecture**
```go
type EventHandlerImpl struct {
    container ComponentContainer  // Dependency injection container
    config    *config.Config     // Service configuration
    obs       *observability.Observability  // Observability tools
}
```

### **Component Container Pattern**
```go
type ComponentContainer interface {
    GetEventProcessor() EventProcessor
    GetBuildContextManager() BuildContextManager
    GetJobManager() JobManager
    GetServiceManager() ServiceManager
    GetObservability() *observability.Observability
    // ... other components
}
```

### **Event Processing Flow**
```go
func (h *EventHandlerImpl) ProcessCloudEvent(ctx context.Context, event *cloudevents.Event) (*builds.HandlerResponse, error) {
    // 1. Validate event
    if err := h.container.GetEventValidator().ValidateEvent(ctx, event); err != nil {
        return nil, err
    }
    
    // 2. Route based on event type
    if h.isBuildStartEvent(event) {
        return h.processBuildStartEvent(ctx, event)
    } else if h.isBuildCompleteEvent(event) {
        return h.processBuildCompleteEvent(ctx, event)
    }
    
    return nil, errors.NewValidationError("event_type", event.Type(), "unsupported event type")
}
```

### **Service Creation Process**
```go
func (h *EventHandlerImpl) createServiceAndTrigger(ctx context.Context, completionData *builds.BuildCompletionEventData) (*builds.HandlerResponse, error) {
    serviceManager := h.container.GetServiceManager()
    serviceName := serviceManager.GenerateServiceName(completionData.ThirdPartyID, completionData.ParserID)
    
    // Create Knative service (includes trigger creation)
    if err := serviceManager.CreateService(ctx, serviceName, completionData); err != nil {
        return nil, err
    }
    
    return &builds.HandlerResponse{
        Status:  "service_created",
        Message: "Knative service and trigger created successfully",
        JobName: completionData.JobName,
    }, nil
}
```

## 🔐 Security, Reusability, and Maintainability

### **Security Features**
- **Input Validation**: Comprehensive validation of all RabbitMQ event payloads
- **Security Validation**: ID validation for third-party and parser IDs
- **Rate Limiting**: Multi-level rate limiting for build context and job creation
- **Error Handling**: Secure error handling without information leakage

### **Reusability**
- **Component Interfaces**: All major components are interface-based for easy testing and swapping
- **Dependency Injection**: Clean dependency management for component composition
- **Modular Design**: Each component has a single responsibility and clear contracts

### **Maintainability**
- **Comprehensive Observability**: Distributed tracing, metrics, and structured logging
- **Error Propagation**: Proper error wrapping and context preservation
- **Component Composition**: Clear separation of concerns with focused components
- **Testability**: Interface-based design enables comprehensive unit testing

### **Best Practices**
- **Go 1.18+ Features**: Uses modern Go features like generics and improved error handling
- **Kubernetes Best Practices**: Follows Kubernetes patterns for job and service management
- **CloudEvents Standards**: Implements CloudEvents specification for event interoperability
- **Knative Integration**: Leverages Knative serving and eventing for serverless capabilities