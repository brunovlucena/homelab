# Kubernetes Operator Architecture for Knative Lambda

## 🎯 Overview

This document proposes a Kubernetes Operator architecture that **integrates with the existing event-driven architecture**. The operator receives CloudEvents via RabbitMQ Broker and Triggers, while also supporting declarative CRD-based management. The operator manages RabbitMQ Brokers, Triggers, and DLQ resources as part of the eventing infrastructure.

### Supported Event Types

The operator supports **17 CloudEvent types** organized into 5 categories:

- **Function Management** (3): `created`, `updated`, `deleted`
- **Build Events** (6): `started`, `completed`, `failed`, `timeout`, `cancelled`, `stopped`
- **Service Events** (3): `created`, `updated`, `deleted`
- **Status Events** (2): `updated`, `health.check`
- **Parser Events** (3): `started`, `completed`, `failed`

See [Complete Event Type Reference](#complete-event-type-reference) for full details.

## 📊 Current Architecture vs Operator Architecture

### Current (Event-Driven)
```
CloudEvent → RabbitMQ Broker → Trigger → Builder Service → Kaniko Job → Knative Service
```

### Proposed (Hybrid: CRD + Event-Driven)
```
┌─────────────────────────────────────────────────────────────────────┐
│              Hybrid Architecture: CRD + Events                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Option 1: Declarative (CRD)                                        │
│  LambdaFunction CRD → KOperator → Kaniko Job → KService             │
│                                                                     │
│  Option 2: Event-Driven                                             │
│  CloudEvent → Broker → Trigger → KOperator → Kaniko Job → KService  │
│                                                                     │
│  Both paths create/manage:                                          │
│  - RabbitMQ Broker                                                  │
│  - Knative Triggers                                                 │
│  - DLQ Resources                                                    │
│  - LambdaFunction Status                                            │
└─────────────────────────────────────────────────────────────────────┘
```

## 🏗️ Architecture Design

### 1. Custom Resource Definition (CRD)

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: lambdafunctions.lambda.knative.io
spec:
  group: lambda.knative.io
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              # Source code configuration
              source:
                type: object
                properties:
                  type: # enum: minio, s3, gcs, git, inline
                    type: string
                    enum: ["minio", "s3", "gcs", "git", "inline"]
                    default: "minio"
                    description: |
                      Storage type for source code:
                      - minio: Local MinIO (default, for Kind clusters)
                      - s3: AWS S3 (optional, for AWS deployments)
                      - gcs: Google Cloud Storage (optional, for GCP deployments)
                      - git: Git repository
                      - inline: Inline code
                  minio:
                    type: object
                    properties:
                      endpoint:
                        type: string
                        default: "minio.minio.svc.cluster.local:9000"
                      bucket:
                        type: string
                      key:
                        type: string
                      accessKey:
                        type: string
                      secretKey:
                        type: string
                  s3:
                    type: object
                    properties:
                      bucket:
                        type: string
                      key:
                        type: string
                      region:
                        type: string
                        default: "us-east-1"
                  gcs:
                    type: object
                    properties:
                      bucket:
                        type: string
                      key:
                        type: string
                      project:
                        type: string
                  git:
                    type: object
                    properties:
                      url:
                        type: string
                      ref:
                        type: string
                      path:
                        type: string
                  inline:
                    type: object
                    properties:
                      code:
                        type: string
                      dependencies:
                        type: string
              
              # Runtime configuration
              runtime:
                type: object
                properties:
                  language:
                    type: string
                    enum: ["nodejs", "python", "go"]
                  version:
                    type: string
                  handler:
                    type: string
                    default: "index.handler"
              
              # Scaling configuration
              scaling:
                type: object
                properties:
                  minReplicas:
                    type: integer
                    default: 0
                  maxReplicas:
                    type: integer
                    default: 50
                  targetConcurrency:
                    type: integer
                    default: 10
                  scaleToZeroGracePeriod:
                    type: string
                    default: "30s"
              
              # Resource limits
              resources:
                type: object
                properties:
                  requests:
                    type: object
                    properties:
                      memory:
                        type: string
                        default: "32Mi"
                      cpu:
                        type: string
                        default: "25m"
                  limits:
                    type: object
                    properties:
                      memory:
                        type: string
                        default: "64Mi"
                      cpu:
                        type: string
                        default: "50m"
              
              # Environment variables
              env:
                type: array
                items:
                  type: object
                  properties:
                    name:
                      type: string
                    value:
                      type: string
                    valueFrom:
                      type: object
              
              # Event triggers
              triggers:
                type: array
                items:
                  type: object
                  properties:
                    broker:
                      type: string
                    filter:
                      type: object
                      properties:
                        type:
                          type: string
                        source:
                          type: string
              
              # Build configuration
              build:
                type: object
                properties:
                  timeout:
                    type: string
                    default: "30m"
                  registry:
                    type: string
                    description: |
                      Container registry URL for built images. Supports:
                      - ECR: {account-id}.dkr.ecr.{region}.amazonaws.com/{repo}:{tag}
                      - GHCR: ghcr.io/{owner}/{repo}:{tag}
                      - GCR: gcr.io/{project-id}/{repo}:{tag} or {region}.gcr.io/{project-id}/{repo}:{tag}
                      - Local: localhost:5001/{repo}:{tag} or {service}.{ns}.svc.cluster.local:5000/{repo}:{tag}
                      If not specified, uses default from ConfigMap or falls back to localhost:5001
                  imagePullSecret:
                    type: string
                    description: |
                      Kubernetes secret name containing registry credentials.
                      Required for private registries (ECR, private GHCR, GCR).
                      Optional for public GHCR and local registries.
          
          status:
            type: object
            properties:
              phase:
                type: string
                enum: ["Pending", "Building", "Deploying", "Ready", "Failed", "Deleting"]
              buildStatus:
                type: object
                properties:
                  jobName:
                    type: string
                  imageURI:
                    type: string
                  startedAt:
                    type: string
                    format: date-time
                  completedAt:
                    type: string
                    format: date-time
                  error:
                    type: string
              serviceStatus:
                type: object
                properties:
                  serviceName:
                    type: string
                  url:
                    type: string
                  ready:
                    type: boolean
                  replicas:
                    type: integer
              conditions:
                type: array
                items:
                  type: object
                  properties:
                    type:
                      type: string
                    status:
                      type: string
                    reason:
                      type: string
                    message:
                      type: string
                    lastTransitionTime:
                      type: string
                      format: date-time
```

### 2. Controller Architecture (Event-Driven)

```
┌────────────────────────────────────────────────────────────┐
│         LambdaFunction Controller (Event-Driven)           │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌────────────────────────────────────────────────────┐    │
│  │  API Endpoints (HTTP/CloudEvents)                  │    │
│  │  ─────────────────────                             │    │
│  │  POST /api/v1/lambda/functions                     │    │
│  │    → lambda.function.created                       │    │
│  │  PUT  /api/v1/lambda/functions/{name}              │    │
│  │    → lambda.function.updated                       │    │
│  │  DELETE /api/v1/lambda/functions/{name}            │    │
│  │    → lambda.function.deleted                       │    │
│  │  POST /api/v1/lambda/functions/{name}/rebuild      │    │
│  │    → lambda.build.started                          │    │
│  │  POST /api/v1/events (CloudEvents endpoint)        │    │
│  │    → All event types                               │    │
│  └────────────────────────────────────────────────────┘    │
│                          │                                 │
│                          ▼                                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Event Ingestion                                   │    │
│  │  ─────────────────────                             │    │
│  │  CloudEvents → RabbitMQ Broker → Triggers          │    │
│  │                                                    │    │
│  │  Supports ALL 17 Event Types:                      │    │
│  │                                                    │    │
│  │  Function Management (3):                          │    │
│  │    • lambda.function.created                       │    │
│  │    • lambda.function.updated                       │    │
│  │    • lambda.function.deleted                       │    │
│  │                                                    │    │
│  │  Build Events (6):                                 │    │
│  │    • lambda.build.started                          │    │
│  │    • lambda.build.completed                        │    │
│  │    • lambda.build.failed                           │    │
│  │    • lambda.build.timeout                          │    │
│  │    • lambda.build.cancelled                        │    │
│  │    • lambda.build.stopped                          │    │
│  │                                                    │    │
│  │  Service Events (3):                               │    │
│  │    • lambda.service.created                        │    │
│  │    • lambda.service.updated                        │    │
│  │    • lambda.service.deleted                        │    │
│  │                                                    │    │
│  │  Status Events (2):                                │    │
│  │    • lambda.status.updated                         │    │
│  │    • lambda.health.check                           │    │
│  │                                                    │    │
│  │  Parser Events (3):                                │    │
│  │    • lambda.parser.started                         │    │
│  │    • lambda.parser.completed                       │    │
│  │    • lambda.parser.failed                          │    │
│  └────────────────────────────────────────────────────┘    │
│                          │                                 │
│                          ▼                                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Reconciler                                        │    │
│  │  - Receives CloudEvents (HTTP + Broker)            │    │
│  │  - Watches LambdaFunction CRDs                     │    │
│  │  - Manages reconciliation loop                     │    │
│  └────────────────────────────────────────────────────┘    │
│                          │                                 │
│                          ▼                                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │  State Machine                                     │    │
│  │  Pending → Building → Deploying → Ready            │    │
│  └────────────────────────────────────────────────────┘    │
│                          │                                 │
│        ┌─────────────────┼─────────────────┐               │
│        ▼                 ▼                 ▼               │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │ Build    │    │ Deploy       │    │ Monitor      │      │
│  │ Manager  │    │ Manager      │    │ Manager      │      │
│  └──────────┘    └──────────────┘    └──────────────┘      │
│        │                 │                 │               │
│        └─────────────────┼─────────────────┘               │
│                          ▼                                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Event Manager                                     │    │
│  │  - Emits CloudEvents (build.complete, etc)         │    │
│  │  - Manages Broker/Triggers                         │    │
│  │  - Handles DLQ                                     │    │
│  └────────────────────────────────────────────────────┘    │
│                          │                                 │
│                          ▼                                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Status Updater                                    │    │
│  │  - Updates CRD status                              │    │
│  │  - Manages conditions                              │    │
│  └────────────────────────────────────────────────────┘    │
└────────────────────────────────────────────────────────────┘
```

### 3. Component Breakdown

#### 3.1 API Server
- **Purpose**: Expose HTTP API for LambdaFunction operations
- **Responsibilities**:
  - Handle REST API endpoints
  - Receive CloudEvents via HTTP
  - Validate request payloads
  - Convert API requests to CloudEvents
  - Emit events to Broker

#### 3.2 Reconciler
- **Purpose**: Main reconciliation loop
- **Responsibilities**:
  - Watch LambdaFunction CRDs
  - Receive CloudEvents via HTTP API and Knative Triggers
  - Determine current state
  - Trigger appropriate action based on state
  - Update status

#### 3.3 Build Manager
- **Purpose**: Handle container image builds
- **Responsibilities**:
  - Create build context from source
  - Create Kaniko Job
  - Monitor build progress via Kubernetes Job API polling
  - Detect job completion (success/failure) during reconciliation
  - Handle build failures
  - Store build artifacts

#### 3.4 Deploy Manager
- **Purpose**: Deploy built images as Knative Services
- **Responsibilities**:
  - Create Knative Service from built image
  - Create Knative Triggers
  - Configure autoscaling
  - Set up environment variables
  - Handle deployment failures

#### 3.5 Monitor Manager
- **Purpose**: Monitor deployed services
- **Responsibilities**:
  - Watch Knative Service status
  - Update LambdaFunction status
  - Handle service failures
  - Manage health checks

#### 3.6 Event Manager (RabbitMQ Integration)
- **Purpose**: Manage event-driven components and receive CloudEvents
- **Responsibilities**:
  - Create and manage RabbitMQ Broker
  - Create and manage Knative Triggers (for operator itself)
  - Receive CloudEvents for LambdaFunction operations
  - Configure DLQ routing
  - Handle event delivery failures
  - Process build.start, build.complete, service.delete events

#### 3.7 DLQ Manager
- **Purpose**: Handle dead letter queue operations
- **Responsibilities**:
  - Create DLQ Exchange, Queue, and Bindings
  - Monitor DLQ depth and age
  - Process DLQ messages for retry
  - Alert on DLQ thresholds

#### 3.8 AI Agent Manager
- **Purpose**: Manage AI Agent integration
- **Responsibilities**:
  - Create AI Agent Trigger (receives all 17 event types)
  - Configure AI Agent Service
  - Monitor AI Agent health
  - Route events to AI Agent

## 🌐 Operator API

### HTTP API Endpoints

The operator exposes a REST API for LambdaFunction operations. All endpoints accept and return CloudEvents format.

**Note on Event Naming Convention**: Following CloudEvents best practices, we use **past tense** for events that represent something that has already happened (notifications). This makes it clear that the event is a fact, not a command.

**Complete Event Type List**:

| Category | Event Type | Description |
|----------|-----------|-------------|
| **Function Management** | `lambda.function.created` | Function CRD was created |
| | `lambda.function.updated` | Function CRD was updated |
| | `lambda.function.deleted` | Function CRD was deleted |
| **Build Events** | `lambda.build.started` | Build process has started |
| | `lambda.build.completed` | Build completed successfully |
| | `lambda.build.failed` | Build failed with error |
| | `lambda.build.timeout` | Build exceeded timeout |
| | `lambda.build.cancelled` | Build was cancelled |
| | `lambda.build.stopped` | Build was stopped |
| **Service Events** | `lambda.service.created` | Knative Service was created |
| | `lambda.service.updated` | Knative Service was updated |
| | `lambda.service.deleted` | Knative Service was deleted |
| **Status Events** | `lambda.status.updated` | Function status changed |
| | `lambda.health.check` | Health check performed |
| **Parser Events** | `lambda.parser.started` | Parser execution started |
| | `lambda.parser.completed` | Parser execution completed |
| | `lambda.parser.failed` | Parser execution failed |

```
┌────────────────────────────────────────────────────────────┐
│              Operator HTTP API                             │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  Base URL: http://operator.lambda.svc.cluster.local        │
│  Content-Type: application/cloudevents+json                │
│                                                            │
│  ┌────────────────────────────────────────────────────┐    │
│  │  POST /api/v1/lambda/functions                     │    │
│  │  ─────────────────────                             │    │
│  │  Creates a new LambdaFunction                      │    │
│  │  Emits: lambda.function.created                    │    │
│  │                                                    │    │
│  │  Request Body:                                     │    │
│  │  {                                                 │    │
│  │    "spec": {                                       │    │
│  │      "source": {...},                              │    │
│  │      "runtime": {...},                             │    │
│  │      "scaling": {...}                              │    │
│  │    }                                               │    │
│  │  }                                                 │    │
│  │                                                    │    │
│  │  Response: 202 Accepted                            │    │
│  │  {                                                 │    │
│  │    "name": "my-function",                          │    │
│  │    "status": "Pending"                             │    │
│  │  }                                                 │    │
│  └────────────────────────────────────────────────────┘    │
│                                                            │
│  ┌────────────────────────────────────────────────────┐    │
│  │  PUT /api/v1/lambda/functions/{name}               │    │
│  │  ─────────────────────                             │    │
│  │  Updates an existing LambdaFunction                │    │
│  │  Emits: lambda.function.updated                    │    │
│  │                                                    │    │
│  │  Request Body:                                     │    │
│  │  {                                                 │    │
│  │    "spec": {                                       │    │
│  │      "scaling": { "maxReplicas": 20 }              │    │
│  │    }                                               │    │
│  │  }                                                 │    │
│  │                                                    │    │
│  │  Response: 200 OK                                  │    │
│  │  {                                                 │    │
│  │    "name": "my-function",                          │    │
│  │    "status": "Ready"                               │    │
│  │  }                                                 │    │
│  └────────────────────────────────────────────────────┘    │
│                                                            │
│  ┌────────────────────────────────────────────────────┐    │
│  │  DELETE /api/v1/lambda/functions/{name}            │    │
│  │  ─────────────────────                             │    │
│  │  Deletes a LambdaFunction                          │    │
│  │  Emits: lambda.function.deleted                    │    │
│  │                                                    │    │
│  │  Response: 202 Accepted                            │    │
│  │  {                                                 │    │
│  │    "name": "my-function",                          │    │
│  │    "status": "Deleting"                            │    │
│  │  }                                                 │    │
│  └────────────────────────────────────────────────────┘    │
│                                                            │
│  ┌────────────────────────────────────────────────────┐    │
│  │  GET /api/v1/lambda/functions                      │    │
│  │  ─────────────────────                             │    │
│  │  Lists all LambdaFunctions                         │    │
│  │                                                    │    │
│  │  Response: 200 OK                                  │    │
│  │  {                                                 │    │
│  │    "items": [...]                                  │    │
│  │  }                                                 │    │
│  └────────────────────────────────────────────────────┘    │
│                                                            │
│  ┌────────────────────────────────────────────────────┐    │
│  │  GET /api/v1/lambda/functions/{name}               │    │
│  │  ─────────────────────                             │    │
│  │  Gets a specific LambdaFunction                    │    │
│  │                                                    │    │
│  │  Response: 200 OK                                  │    │
│  │  {                                                 │    │
│  │    "spec": {...},                                  │    │
│  │    "status": {...}                                 │    │
│  │  }                                                 │    │
│  └────────────────────────────────────────────────────┘    │
│                                                            │
│  ┌────────────────────────────────────────────────────┐    │
│  │  POST /api/v1/events                               │    │
│  │  ─────────────────────                             │    │
│  │  Generic CloudEvents endpoint                      │    │
│  │  Accepts any CloudEvent type                       │    │
│  │                                                    │    │
│  │  Request: CloudEvent format                        │    │
│  │  {                                                 │    │
│  │    "type": "lambda.function.created",              │    │
│  │    "source": "...",                                │    │
│  │    "data": {...}                                   │    │
│  │  }                                                 │    │
│  └────────────────────────────────────────────────────┘    │
└────────────────────────────────────────────────────────────┘
```

### API Request Flow

```
┌─────────────────────────────────────────────────────────────┐
│              API Request Processing Flow                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Client / External System                                   │
│         │                                                   │
│         │ POST /api/v1/lambda/functions                     │
│         │ { "spec": {...} }                                 │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ Operator API Server  │                                   │
│  │ (HTTP Handler)       │                                   │
│  │ ──────────────────── │                                   │
│  │ 1. Validate request  │                                   │
│  │ 2. Parse payload     │                                   │
│  │ 3. Create CloudEvent │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         │ Create CloudEvent                                 │
│         │ type: lambda.function.created                     │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ Event Router         │                                   │
│  │ ──────────────────── │                                   │
│  │ Route to:            │                                   │
│  │ - Internal handler   │                                   │
│  │ - RabbitMQ Broker    │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│    ┌────┴────┐                                              │
│    │        │                                               │
│    ▼        ▼                                               │
│  ┌──────────┐  ┌──────────────────┐                         │
│  │ Internal │  │ RabbitMQ Broker  │                         │
│  │ Handler  │  │ (for downstream) │                         │
│  │ (direct) │  └──────────────────┘                         │
│  └────┬─────┘                                               │
│       │                                                     │
│       │ Process immediately                                 │
│       ▼                                                     │
│  ┌──────────────────────┐                                   │
│  │ Reconciler           │                                   │
│  │ - Create/Update CRD  │                                   │
│  │ - Trigger build      │                                   │
│  │ - Update status      │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         │ Return response                                   │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ HTTP Response        │                                   │
│  │ 202 Accepted         │                                   │
│  │ { "name": "...",      │                                  │
│  │   "status": "..." }   │                                  │
│  └──────────────────────┘                                   │
└─────────────────────────────────────────────────────────────┘
```

### API Event Types

| Endpoint | Method | Event Type Emitted | Description |
|----------|--------|-------------------|-------------|
| `/api/v1/lambda/functions` | POST | `lambda.function.created` | Create new function |
| `/api/v1/lambda/functions/{name}` | PUT | `lambda.function.updated` | Update existing function |
| `/api/v1/lambda/functions/{name}` | DELETE | `lambda.function.deleted` | Delete function |
| `/api/v1/lambda/functions/{name}/rebuild` | POST | `lambda.build.started` | Trigger rebuild |
| `/api/v1/events` | POST | (as specified in request) | Generic CloudEvents endpoint (all event types) |

### Complete Event Type Reference {#complete-event-type-reference}

The operator supports the following CloudEvent types:

#### Function Management Events
- **`lambda.function.created`**: Emitted when a LambdaFunction CRD is created
- **`lambda.function.updated`**: Emitted when a LambdaFunction CRD is updated
- **`lambda.function.deleted`**: Emitted when a LambdaFunction CRD is deleted

#### Build Events
- **`lambda.build.started`**: Emitted when a build process starts
- **`lambda.build.completed`**: Emitted when a build completes successfully
- **`lambda.build.failed`**: Emitted when a build fails with an error
- **`lambda.build.timeout`**: Emitted when a build exceeds its timeout
- **`lambda.build.cancelled`**: Emitted when a build is cancelled
- **`lambda.build.stopped`**: Emitted when a build is stopped

#### Service Events
- **`lambda.service.created`**: Emitted when a Knative Service is created
- **`lambda.service.updated`**: Emitted when a Knative Service is updated
- **`lambda.service.deleted`**: Emitted when a Knative Service is deleted

#### Status Events
- **`lambda.status.updated`**: Emitted when function status changes
- **`lambda.health.check`**: Emitted during health checks

#### Parser Events (Optional)
- **`lambda.parser.started`**: Emitted when parser execution starts
- **`lambda.parser.completed`**: Emitted when parser execution completes
- **`lambda.parser.failed`**: Emitted when parser execution fails

## 🤖 AI Agent Integration

### AI Agent Overview

The AI Agent is a Knative Service that receives **all 17 event types** from the operator broker. It provides intelligent monitoring, analysis, and automated responses to LambdaFunction lifecycle events. As a **first-class observability citizen**, it also receives alerts from Alertmanager, investigates PrometheusRules, queries Prometheus for context, and provides intelligent insights and recommendations.

```
┌────────────────────────────────────────────────────────────┐
│              AI Agent Architecture                         │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌────────────────────────────────────────────────────┐    │
│  │  AI Agent Service                                  │    │
│  │  (Knative Service)                                 │    │
│  │  ─────────────────────                             │    │
│  │  Receives ALL 17 Event Types                       │    │
│  │                                                    │    │
│  │  Capabilities:                                     │    │
│  │  - Event Analysis & Pattern Detection              │    │
│  │  - Alert Investigation (Alertmanager)              │    │
│  │  - PrometheusRule Analysis                         │    │
│  │  - Prometheus Query & Context Gathering            │    │
│  │  - Intelligent Recommendations                     │    │
│  │  - First-Class Observability                       │    │
│  │  - Anomaly Detection                               │    │
│  │  - Predictive Scaling                              │    │
│  │  - Automated Remediation                           │    │
│  │  - Performance Insights                            │    │
│  │  - Cost Optimization                               │    │
│  └────────────────────────────────────────────────────┘    │
│                          │                                 │
│                          ▼                                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Event Processing Pipeline                         │    │
│  │  ─────────────────────                             │    │
│  │  1. Receive CloudEvent                             │    │
│  │  2. Analyze event data                             │    │
│  │  3. Generate insights                              │    │
│  │  4. Take actions (if needed)                       │    │
│  │  5. Emit analysis events                           │    │
│  └────────────────────────────────────────────────────┘    │
└────────────────────────────────────────────────────────────┘
```

### AI Agent Trigger Configuration

The AI Agent has a dedicated trigger that receives all event types:

```yaml
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: ai-agent-all-events
  namespace: knative-lambda
  annotations:
    rabbitmq.eventing.knative.dev/parallelism: "50"
    description: "Routes all LambdaFunction events to AI Agent"
spec:
  broker: knative-lambda-operator-broker
  filter:
    attributes:
      # Function Management Events
      type: lambda.function.created
      type: lambda.function.updated
      type: lambda.function.deleted
      # Build Events
      type: lambda.build.started
      type: lambda.build.completed
      type: lambda.build.failed
      type: lambda.build.timeout
      type: lambda.build.cancelled
      type: lambda.build.stopped
      # Service Events
      type: lambda.service.created
      type: lambda.service.updated
      type: lambda.service.deleted
      # Status Events
      type: lambda.status.updated
      type: lambda.health.check
      # Parser Events
      type: lambda.parser.started
      type: lambda.parser.completed
      type: lambda.parser.failed
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: knative-lambda-ai-agent
      namespace: knative-lambda
  delivery:
    retry: 5
    backoffPolicy: exponential
    backoffDelay: PT1S
```

### AI Agent Event Flow

```
┌─────────────────────────────────────────────────────────────┐
│         AI Agent Event Processing Flow                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Event Source (Operator / Builder / Service)                │
│         │                                                   │
│         │ Emit CloudEvent                                   │
│         │ (any of 17 types)                                 │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ RabbitMQ Broker      │                                   │
│  │ (Operator Broker)    │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         │ Event routing                                     │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ AI Agent Trigger     │                                   │
│  │ (all-events)         │                                   │
│  │ ──────────────────── │                                   │
│  │ Filters: ALL types   │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         │ CloudEvent delivery                               │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ AI Agent Service     │                                   │
│  │ (Knative Service)    │                                   │
│  │ ──────────────────── │                                   │
│  │ 1. Receive event     │                                   │
│  │ 2. Parse & validate  │                                   │
│  │ 3. Analyze context   │                                   │
│  │ 4. Generate insights │                                   │
│  │ 5. Detect anomalies  │                                   │
│  │ 6. Take actions      │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│    ┌────┴────┐                                              │
│    │         │                                              │
│    ▼         ▼                                              │
│  Insights  Actions                                          │
│    │         │                                              │
│    │         │ Emit: ai.analysis.completed                  │
│    │         │ Emit: ai.recommendation.generated            │
│    │         │ Emit: ai.action.taken                        │
│    │         │                                              │
│    └─────────┘                                              │
│         │                                                   │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ Broker (for AI)      │                                   │
│  │ (downstream events)  │                                   │
│  └──────────────────────┘                                   │
└─────────────────────────────────────────────────────────────┘
```

## 🐰 RabbitMQ & Event Components

### RabbitMQ Broker Architecture

```
┌────────────────────────────────────────────────────────────┐
│              RabbitMQ Event Infrastructure                 │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌────────────────────────────────────────────────────┐    │
│  │  RabbitMQ Cluster                                  │    │
│  │  ─────────────────────                             │    │
│  │  Namespace: rabbitmq-cluster-knative-lambda        │    │
│  │  Type: Quorum queues                               │    │
│  └────────────────────────────────────────────────────┘    │
│                          │                                 │
│        ┌─────────────────┼─────────────────┐               │
│        ▼                 ▼                 ▼               │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │ Broker   │    │ DLQ Exchange │    │ DLQ Queue    │      │
│  │ Exchange │    │              │    │              │      │
│  └──────────┘    └──────────────┘    └──────────────┘      │
│        │                 │                 │               │
│        └─────────────────┼─────────────────┘               │
│                          ▼                                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Knative Broker                                    │    │
│  │  ─────────────────────                             │    │
│  │  Class: RabbitMQBroker                             │    │
│  │  DLQ Enabled: true                                 │    │
│  │  Retry Policy: exponential                         │    │
│  │  Max Retries: 5                                    │    │
│  └────────────────────────────────────────────────────┘    │
│                          │                                 │
│                          ▼                                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Knative Triggers                                  │    │
│  │  ─────────────────────                             │    │
│  │  - Filter by event type                            │    │
│  │  - Route to LambdaFunction services                │    │
│  │  - Parallelism: 50                                 │    │
│  └────────────────────────────────────────────────────┘    │
└────────────────────────────────────────────────────────────┘
```

### DLQ Flow Architecture

**What are DLQ Exchange and DLQ Queue?**

In your knative-lambda context, the **DLQ (Dead Letter Queue)** system is a safety mechanism that prevents **event loss** when CloudEvents fail to be processed after all retry attempts are exhausted.

#### DLQ Exchange (`knative-lambda-dlq-exchange`)
- **Purpose**: A RabbitMQ **Exchange** (type: `topic`) that acts as a routing hub for failed events
- **What it does**: Receives events that failed after 5 retry attempts from the Knative Broker
- **How it works**: Routes failed events to the DLQ Queue using routing keys like `dlq.*`
- **Why needed**: Provides a centralized point to collect all failed events before they're stored

#### DLQ Queue (`knative-lambda-dlq`)
- **Purpose**: A RabbitMQ **Queue** (type: `quorum`) that **stores** failed events persistently
- **What it does**: Holds failed CloudEvents that couldn't be processed
- **Configuration**:
  - **TTL**: 7 days (events auto-expire after 7 days)
  - **Max Length**: 50,000 messages (drops oldest when full)
  - **Overflow Policy**: `drop-head` (removes oldest messages when limit reached)
- **Why needed**: Prevents data loss, enables investigation, and allows manual replay

#### When Events Go to DLQ

Events are routed to DLQ when:
1. **All retries exhausted** (5 attempts with exponential backoff)
2. **Poison messages** (malformed events that crash consumers)
3. **Service unavailable** (LambdaFunction service down/crashing)
4. **Timeout exceeded** (processing takes longer than 30s)
5. **Validation failures** (events don't match expected schema)

#### Benefits in Your Context

1. **No Event Loss**: Failed `lambda.build.started`, `lambda.function.created`, etc. are preserved
2. **Debugging**: Investigate why events failed (check DLQ for patterns)
3. **Recovery**: Replay events after fixing underlying issues
4. **Observability**: Monitor DLQ depth/age to detect systemic issues
5. **Poison Message Isolation**: Bad events don't crash all consumers

```
┌────────────────────────────────────────────────────────────┐
│                    DLQ Processing Flow                     │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  Event Processing Failure                                  │
│         │                                                  │
│         ▼                                                  │
│  ┌──────────────────────┐                                  │
│  │ Knative Broker       │                                  │
│  │ Retry Attempt 1      │                                  │
│  └──────┬───────────────┘                                  │
│         │                                                  │
│         ▼ (Failure)                                        │
│  ┌──────────────────────┐                                  │
│  │ Retry Attempt 2      │                                  │
│  └──────┬───────────────┘                                  │
│         │                                                  │
│         ▼ (Failure)                                        │
│  ┌──────────────────────┐                                  │
│  │ Retry Attempt 3-5    │                                  │
│  │ (Exponential backoff)│                                  │
│  └──────┬───────────────┘                                  │
│         │                                                  │
│         ▼ (All Retries Exhausted)                          │
│  ┌──────────────────────┐                                  │
│  │ DLQ Exchange         │                                  │
│  │ (knative-lambda-dlq) │                                  │
│  └──────┬───────────────┘                                  │
│         │                                                  │
│         │ Routing Key: dlq.*                               │
│         ▼                                                  │
│  ┌──────────────────────┐                                  │
│  │ DLQ Binding          │                                  │
│  │ Source: DLQ Exchange │                                  │
│  │ Dest: DLQ Queue      │                                  │
│  └──────┬───────────────┘                                  │
│         │                                                  │
│         ▼                                                  │
│  ┌──────────────────────┐                                  │
│  │ DLQ Queue            │                                  │
│  │ ───────────────────  │                                  │
│  │ TTL: 7 days          │                                  │
│  │ Max Length: 50,000   │                                  │
│  │ Overflow: drop-head  │                                  │
│  └──────┬───────────────┘                                  │
│         │                                                  │
│         ▼                                                  │
│  ┌──────────────────────┐                                  │
│  │ DLQ Handler Service  │                                  │
│  │ ───────────────────  │                                  │
│  │ - Process messages   │                                  │
│  │ - Alert on depth     │                                  │
│  │ - Manual retry       │                                  │
│  └──────────────────────┘                                  │
└────────────────────────────────────────────────────────────┘
```

### Event-Driven Lambda Function Flow

```
┌─────────────────────────────────────────────────────────────┐
│         Event-Driven Lambda Function Execution              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Notifi Scheduler Service (CloudEvent Source)               │
│         │                                                   │
│         │ POST CloudEvent                                   │
│         │ Types:                                            │
│         │ - lambda.build.started (start)                    │
│         │ - lambda.build.stopped (stop)                     │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ Knative Broker       │                                   │
│  │ (RabbitMQ Broker)    │                                   │
│  │ ───────────────────  │                                   │
│  │ Exchange: broker     │                                   │
│  │ Routes by event type │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         │ Event routing                                     │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ Knative Trigger      │                                   │
│  │ ───────────────────  │                                   │
│  │ Filter:              │                                   │
│  │   type: my.event.type│                                   │
│  │ Subscriber:          │                                   │
│  │   my-lambda-function │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         │ CloudEvent delivery                               │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ LambdaFunction       │                                   │
│  │ (Knative Service)    │                                   │
│  │ ───────────────────  │                                   │
│  │ Auto-scales 0→N      │                                   │
│  │ Processes event      │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│    ┌────┴────┐                                              │
│    │         │                                              │
│    ▼         ▼                                              │
│  Success   Failure                                          │
│    │         │                                              │
│    │         └──► Retry via Broker                          │
│    │              (up to 5 attempts)                        │
│    │                                                        │
│    └──► Response / Result                                   │
└─────────────────────────────────────────────────────────────┘
```

## 🔄 Event-Driven Reconciliation Flow

> **Note**: The event source for CloudEvents is the **Notifi Scheduler Service**. The Scheduler can publish events to:
> - **Create** a lambda: `lambda.function.created`
> - **Delete** a lambda: `lambda.function.deleted`
> - **Update** a lambda: `lambda.function.updated`
> - **Start** a lambda: `lambda.build.started` (or `lambda.parser.started`)
> - **Stop** a lambda: `lambda.build.stopped`
>
> See [`NOTIFI_INTEGRATION.md`](../../04-architecture/NOTIFI_INTEGRATION.md) for complete integration details.

### Operator Receives CloudEvents via Broker/Triggers

```
┌─────────────────────────────────────────────────────────────┐
│         Operator Event-Driven Architecture                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. EVENT SOURCE                                            │
│     ┌──────────────────┐                                    │
│     │ Notifi Scheduler │                                    │
│     │ Service          │                                    │
│     │ (External System)│                                    │
│     └────────┬─────────┘                                    │
│              │                                              │
│              │ POST CloudEvent                              │
│              │ Types:                                       │
│              │ - lambda.function.created (create)           │
│              │ - lambda.function.updated (update)           │
│              │ - lambda.function.deleted (delete)           │
│              │ - lambda.build.started (start)               │
│              │ - lambda.build.stopped (stop)                │
│              ▼                                              │
│     ┌──────────────────┐                                    │
│     │ RabbitMQ Broker  │                                    │
│     │ (Knative Broker) │                                    │
│     │ ──────────────── │                                    │
│     │ Exchange: broker │                                    │
│     │ Routes by type   │                                    │
│     └────────┬─────────┘                                    │
│              │                                              │
│  2. EVENT ROUTING                                           │
│              │                                              │
│              │ Filter by event type                         │
│              ▼                                              │
│     ┌──────────────────┐                                    │
│     │ Knative Trigger  │                                    │
│     │ ──────────────── │                                    │
│     │ Filter:          │                                    │
│     │   type: lambda.* │                                    │
│     │ Subscriber:      │                                    │
│     │   operator-ksvc  │                                    │
│     └────────┬─────────┘                                    │
│              │                                              │
│  3. EVENT DELIVERY TO OPERATOR                              │
│              │                                              │
│              │ CloudEvent POST                              │
│              ▼                                              │
│     ┌───────────────────┐                                   │
│     │ Operator          │                                   │
│     │ Controller        │                                   │
│     │ (Knative Service) │                                   │
│     │ ────────────────  │                                   │
│     │ Receives:         │                                   │
│     │ - function.created│                                   │
│     │ - function.updated│                                   │
│     │ - function.deleted│                                   │
│     │ - build.started   │                                   │
│     │ - build.stopped   │                                   │
│     └────────┬──────────┘                                   │
│              │                                              │
│  4. EVENT PROCESSING                                        │
│              │                                              │
│              ▼                                              │
│     ┌───────────────────┐                                   │
│     │ Event Handler     │                                   │
│     │ ────────────────  │                                   │
│     │ - Parse CloudEvent│                                   │
│     │ - Extract data    │                                   │
│     │ - Get/Create CRD  │                                   │
│     │ - Determine action│                                   │
│     └────────┬──────────┘                                   │
│              │                                              │
│              ▼                                              │
│     ┌──────────────────┐                                    │
│     │ Reconciliation   │                                    │
│     │ Logic            │                                    │
│     │ ──────────────── │                                    │
│     │ - Build Manager  │                                    │
│     │ - Deploy Manager │                                    │
│     │ - Event Manager  │                                    │
│     └────────┬─────────┘                                    │
│              │                                              │
│  5. STATUS UPDATE                                           │
│              │                                              │
│              ▼                                              │
│     ┌──────────────────┐                                    │
│     │ Update CRD       │                                    │
│     │ Status           │                                    │
│     │ ──────────────── │                                    │
│     │ - Phase          │                                    │
│     │ - Conditions     │                                    │
│     │ - Build Status   │                                    │
│     │ - Service Status │                                    │
│     └────────┬─────────┘                                    │
│              │                                              │
│  6. EVENT EMISSION                                          │
│              │                                              │
│              │ Emit CloudEvents:                            │
│              │ - lambda.build.completed                     │
│              │ - lambda.build.failed                        │
│              │ (ALL 17 event types)                         │
│              ▼                                              │
│     ┌──────────────────┐                                    │
│     │ RabbitMQ Broker  │                                    │
│     │ (for downstream) │                                    │
│     └──────┬───────────┘                                    │
│            │                                                │
│            │ Events routed to:                              │
│            │ - Operator Controller                          │
│            │ - LambdaFunction Services                      │
│            │ - AI Agent Service                             │
│            ▼                                                │
│     ┌──────────────────┐                                    │
│     │ AI Agent         │                                    │
│     │ (receives all)   │                                    │
│     └──────────────────┘                                    │
└─────────────────────────────────────────────────────────────┘
```

### Operator Triggers Configuration

```
┌────────────────────────────────────────────────────────────┐
│         Operator Knative Triggers                          │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  Trigger 1: Function Create/Update                         │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Filter:                                            │    │
│  │   type: lambda.function.created                    │    │
│  │   type: lambda.function.updated                    │    │
│  │ Subscriber: operator-controller                    │    │
│  └────────────────────────────────────────────────────┘    │
│                                                            │
│  Trigger 2: Build Events                                   │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Filter:                                            │    │
│  │   type: lambda.build.started                       │    │
│  │   type: lambda.build.completed                     │    │
│  │   type: lambda.build.failed                        │    │
│  │   type: lambda.build.timeout                       │    │
│  │   type: lambda.build.cancelled                     │    │
│  │   type: lambda.build.stopped                       │    │
│  │ Subscriber: operator-controller                    │    │
│  └────────────────────────────────────────────────────┘    │
│                                                            │
│  Trigger 3: Service Lifecycle                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Filter:                                            │    │
│  │   type: lambda.service.created                     │    │
│  │   type: lambda.service.updated                     │    │
│  │   type: lambda.service.deleted                     │    │
│  │ Subscriber: operator-controller                    │    │
│  └────────────────────────────────────────────────────┘    │
│                                                            │
│  Trigger 4: Status Events                                  │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Filter:                                            │    │
│  │   type: lambda.status.updated                      │    │
│  │   type: lambda.health.check                        │    │
│  │ Subscriber: operator-controller                    │    │
│  └────────────────────────────────────────────────────┘    │
│                                                            │
│  All Triggers:                                             │
│  - Broker: knative-lambda-operator-broker                  │
│  - Parallelism: 50                                         │
│  - DLQ: Enabled                                            │
│  - Retry: 5 attempts (exponential backoff)                 │
└────────────────────────────────────────────────────────────┘
```

## 🔄 Reconciliation Flow

```
┌─────────────────────────┐
│ LambdaFunction Created  │
└───────────┬─────────────┘
            │
            ▼
    ┌───────────────┐
    │  Check Phase  │
    └───────┬───────┘
            │
    ┌───────┴────────┐
    │                │
    ▼                ▼
┌─────────┐    ┌──────────────┐
│ Pending │    │ Other Phase  │
└────┬────┘    └──────────────┘
     │
     ▼
┌─────────────────────┐
│ Create Build Context│
└──────────┬──────────┘
           │
           ▼
┌──────────────────┐
│ Create Kaniko Job│
└────────┬─────────┘
         │
         ▼
   ┌─────────────┐
   │Build Status │
   └───┬───┬─────┘
       │   │
   ┌───┘   └───┐
   │           │
   ▼           ▼
┌─────────┐ ┌──────────┐
│Building │ │  Failed  │
└────┬────┘ └────┬─────┘
     │           │
     │           ▼
     │    ┌──────────────┐
     │    │Set Error     │
     │    │Condition     │
     │    └──────────────┘
     │
     ▼
┌──────────────┐
│Build Success │
└──────┬───────┘
       │
       ▼
┌──────────────────────┐
│ Update Status:       │
│ Deploying            │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ Create Knative       │
│ Service              │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ Create RabbitMQ      │
│ Broker (if needed)   │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ Create DLQ Resources │
│ (Exchange/Queue/     │
│  Binding)            │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ Create Knative       │
│ Triggers (from spec) │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ Create AI Agent      │
│ Trigger (all events) │
└──────────┬───────────┘
           │
           ▼
    ┌──────────────┐
    │Service Ready?│
    └───┬──────┬───┘
        │      │
    ┌───┘      └───┐
    │              │
    ▼              ▼
┌─────────┐   ┌──────────────┐
│  Yes    │   │     No       │
└────┬────┘   └──────┬───────┘
     │               │
     │               ▼
     │        ┌──────────────┐
     │        │Update Status:│
     │        │Deploying     │
     │        └──────┬───────┘
     │               │
     │               └───┐
     │                   │
     ▼                   │
┌──────────────────┐     │
│Update Status:    │     │
│Ready             │     │
│                  │     │
│ Components:      │     │
│ ✓ Service        │     │
│ ✓ Broker         │     │
│ ✓ Triggers       │     │
│ ✓ DLQ            │     │
│ ✓ AI Agent       │     │
└────────┬─────────┘     │
         │               │
         ▼               │
┌──────────────────┐     │
│ Monitor Service  │◄────┘
│ & Event Flow     │
└────────┬─────────┘
         │
         ▼
   ┌───────────────┐
   │Service        │
   │Healthy?       │
   └───┬──────┬────┘
       │      │
   ┌───┘      └───┐
   │              │
   ▼              ▼
┌─────────┐   ┌──────────────┐
│  Yes    │   │     No       │
└────┬────┘   └──────┬───────┘
     │               │
     │               ▼
     │        ┌──────────────┐
     │        │Update Status:│
     │        │Failed        │
     │        └──────┬───────┘
     │               │
     └───────────────┘
                   │
                   ▼
            ┌──────────────┐
            │Set Error     │
            │Condition     │
            └──────────────┘
```

### Kaniko Job Status Detection

The operator detects kaniko job completion or failure through **polling during reconciliation**. The operator does not rely on job events or webhooks; instead, it actively queries the Kubernetes Job API to check the job status.

#### Detection Mechanism

1. **Reconciliation Trigger**: When a `LambdaFunction` is in the "Building" phase, the reconciler processes it during each reconciliation cycle.

2. **Status Query**: The reconciler calls `BuildManager.GetBuildStatus(ctx, lambda)`, which:
   - Retrieves the Kubernetes Job using the job name stored in `lambda.Status.BuildStatus.JobName`
   - Queries the Job's status from the Kubernetes API
   - Checks the Job's phase: `Pending`, `Running`, `Succeeded`, or `Failed`
   - Extracts completion information and error messages if failed

3. **Status Evaluation**:
   - If `status.Completed == true`:
     - **Success**: If `status.Success == true`, the operator:
       - Updates `LambdaFunction.Status.Phase` to "Deploying"
       - Sets `BuildReady` condition to `True`
       - Stores the built image URI in `lambda.Status.BuildStatus.ImageURI`
       - Records completion timestamp
     - **Failure**: If `status.Success == false`, the operator:
       - Updates `LambdaFunction.Status.Phase` to "Failed"
       - Sets `BuildReady` condition to `False` with error details
       - Stores error message in `lambda.Status.BuildStatus.Error`
       - Records completion timestamp

4. **Polling Interval**: If the job is still running (`status.Completed == false`), the reconciler:
   - Returns `ctrl.Result{RequeueAfter: 10 * time.Second}`
   - The controller-runtime framework automatically requeues the reconciliation after 10 seconds
   - This continues until the job completes or fails

#### Implementation Example

```go
func (r *LambdaFunctionReconciler) reconcileBuilding(ctx context.Context, lambda *lambdaapi.LambdaFunction, log logr.Logger) (ctrl.Result, error) {
    log.Info("Reconciling Building phase")

    // Check build job status via Kubernetes API
    status, err := r.BuildManager.GetBuildStatus(ctx, lambda)
    if err != nil {
        return ctrl.Result{}, err
    }

    if status.Completed {
        if status.Success {
            // Build succeeded, move to Deploying
            lambda.Status.Phase = "Deploying"
            lambda.Status.BuildStatus.ImageURI = status.ImageURI
            lambda.Status.BuildStatus.CompletedAt = &metav1.Time{Time: time.Now()}
            
            if err := r.setCondition(ctx, lambda, "BuildReady", "True", "BuildSucceeded", "Image built successfully", log); err != nil {
                return ctrl.Result{}, err
            }
            
            if err := r.Status().Update(ctx, lambda); err != nil {
                return ctrl.Result{}, err
            }
            
            return ctrl.Result{Requeue: true}, nil
        } else {
            // Build failed
            lambda.Status.Phase = "Failed"
            lambda.Status.BuildStatus.Error = status.Error
            lambda.Status.BuildStatus.CompletedAt = &metav1.Time{Time: time.Now()}
            
            if err := r.setCondition(ctx, lambda, "BuildReady", "False", "BuildFailed", status.Error, log); err != nil {
                return ctrl.Result{}, err
            }
            
            if err := r.Status().Update(ctx, lambda); err != nil {
                return ctrl.Result{}, err
            }
            
            return ctrl.Result{}, nil
        }
    }

    // Still building, requeue after 10 seconds
    return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}
```

#### Key Characteristics

- **Polling-Based**: The operator actively polls the Kubernetes Job API every 10 seconds while the job is running
- **No Event Dependencies**: The detection mechanism does not rely on Kubernetes events or webhooks
- **Controller-Runtime Pattern**: Uses the standard controller-runtime reconciliation pattern with requeue intervals
- **Secondary Event Emission**: While the operator detects completion via polling, it also emits CloudEvents (`lambda.build.completed`, `lambda.build.failed`) to the broker for downstream consumers (AI Agent, monitoring, etc.)

#### Build Manager GetBuildStatus Implementation

The `GetBuildStatus` method queries the Kubernetes Job API:

```go
func (bm *BuildManager) GetBuildStatus(ctx context.Context, lambda *lambdaapi.LambdaFunction) (*BuildStatus, error) {
    // Get job name from LambdaFunction status
    jobName := lambda.Status.BuildStatus.JobName
    
    // Query Kubernetes Job API
    job := &batchv1.Job{}
    err := bm.client.Get(ctx, types.NamespacedName{
        Name:      jobName,
        Namespace: lambda.Namespace,
    }, job)
    
    if err != nil {
        return nil, err
    }
    
    // Check job conditions
    status := &BuildStatus{
        Completed: job.Status.Succeeded > 0 || job.Status.Failed > 0,
        Success:   job.Status.Succeeded > 0,
    }
    
    if !status.Success && job.Status.Failed > 0 {
        // Extract error from job or pod logs
        status.Error = bm.extractJobError(ctx, job)
    }
    
    if status.Success {
        status.ImageURI = bm.getImageURI(lambda, job)
    }
    
    return status, nil
}
```

## 📦 Implementation Structure

The operator follows **Kubebuilder** conventions and Go project layout best practices. The structure separates CRD definitions, controllers, business logic, and HTTP API concerns.

```
operator/
├── api/                                 # CRD API definitions (Kubebuilder convention)
│   └── v1alpha1/
│       ├── lambdafunction_types.go      # CRD types and spec/status
│       ├── lambdafunction_webhook.go    # Validation and mutation webhooks
│       └── groupversion_info.go         # API group version metadata
│
├── controllers/                         # Kubernetes controllers (Kubebuilder convention)
│   └── lambdafunction_controller.go    # Main LambdaFunction reconciler
│
├── internal/                           # Private implementation packages
│   ├── build/                          # Build management logic
│   │   ├── manager.go                   # Build orchestration
│   │   ├── context.go                   # Build context creation
│   │   └── kaniko.go                    # Kaniko job management
│   ├── deploy/                         # Deployment management logic
│   │   ├── manager.go                   # Deployment orchestration
│   │   ├── knative.go                   # Knative service creation
│   │   └── triggers.go                  # Trigger management
│   ├── status/                          # Status management
│   │   ├── updater.go                   # Status updates
│   │   └── conditions.go                # Condition management
│   ├── storage/                         # Storage clients (MinIO/S3/GCS/Git)
│   │   ├── factory.go                   # Storage factory
│   │   ├── minio.go                     # MinIO client
│   │   ├── s3.go                        # AWS S3 client
│   │   ├── gcs.go                       # Google Cloud Storage client
│   │   └── git.go                       # Git client
│   ├── templates/                       # Dockerfile templates
│   │   └── templates.go                 # Template generation
│   ├── events/                          # Event emission helpers
│   │   ├── emitter.go                   # CloudEvents emitter
│   │   └── types.go                      # Event type definitions
│   └── utils/                           # Internal utility functions
│       └── helpers.go                   # Helper functions
│
├── pkg/                                 # Public reusable packages
│   └── server/                          # HTTP API server
│       ├── handlers.go                  # HTTP request handlers
│       ├── routes.go                    # API route definitions
│       ├── cloudevents.go               # CloudEvents processing
│       └── middleware.go                # Auth, validation, logging
│
├── cmd/                                 # Application entry points
│   └── manager/
│       └── main.go                      # Operator main entry point
│
└── config/                             # Configuration manifests
    ├── crd/                             # CRD manifests (Kubebuilder generated)
    │   └── bases/
    │       └── lambda.knative.io_lambdafunctions.yaml
    ├── rbac/                            # RBAC manifests
    │   ├── role.yaml
    │   ├── role_binding.yaml
    │   └── service_account.yaml
    └── samples/                         # Example CRD instances
        └── lambdafunction-sample.yaml
```

### Structure Rationale

**Kubebuilder Conventions**:
- `api/v1alpha1/`: CRD type definitions (standard Kubebuilder location)
- `controllers/`: Controller/reconciler implementations
- `cmd/manager/main.go`: Main entry point (not root `main.go`)
- `config/crd/bases/`: Generated CRD manifests

**Go Project Layout**:
- `internal/`: Private packages (not importable by external projects)
- `pkg/`: Public reusable packages (importable by external projects)
- Separation of concerns: business logic in `internal/`, HTTP API in `pkg/server/`

**Key Improvements**:
- ✅ Removed duplicate `api/` directory (HTTP API moved to `pkg/server/`)
- ✅ Moved managers from `controllers/` to `internal/` (controllers orchestrate, managers implement)
- ✅ Added `pkg/` for public packages
- ✅ Moved `main.go` to `cmd/manager/main.go` (Kubebuilder convention)
- ✅ Better organization of storage clients
- ✅ Clearer separation between CRD definitions and HTTP API

### AI Agent Structure

The AI Agent is a standalone Knative Service (not a Kubernetes operator), so it follows standard Go service layout conventions.

```
ai-agent/
├── internal/                           # Private implementation packages
│   ├── handlers/                      # CloudEvent handlers
│   │   ├── event_handler.go            # CloudEvent handlers
│   │   ├── analysis.go                 # Event analysis
│   │   └── actions.go                   # Automated actions
│   ├── ai/                             # AI/ML models
│   │   ├── anomaly_detector.go         # Anomaly detection
│   │   ├── predictor.go                 # Predictive models
│   │   └── recommender.go               # Recommendations
│   ├── storage/                        # Event storage
│   │   └── event_store.go               # Event persistence
│   └── metrics/                        # Metrics collection
│       └── collector.go                 # Metrics aggregation
│
├── pkg/                                 # Public reusable packages (if needed)
│   └── types/                           # Shared type definitions
│
├── cmd/                                 # Application entry points
│   └── agent/
│       └── main.go                      # AI Agent main entry point
│
└── config/                             # Configuration files
    └── triggers/                        # AI Agent trigger config
        └── ai-agent-trigger.yaml
```

**Structure Notes**:
- `internal/handlers/`: Event handlers (private, not reusable)
- `cmd/agent/main.go`: Main entry point (consistent with operator structure)
- `internal/ai/`: AI/ML logic (private implementation)
- `config/`: Configuration manifests for deployment

## 🛠️ Technology Stack

### Core Framework
- **Kubebuilder**: Operator framework (see [Technical Decision: Kubebuilder Framework](#technical-decision-kubebuilder-framework))
- **Go 1.24+**: Language
- **Kubernetes Client Go**: K8s API client

### Dependencies
- **Knative Serving Client**: For Knative Service management
- **CloudEvents SDK**: For CloudEvents processing
- **HTTP Router** (Gin/Echo): For REST API endpoints
- **Kaniko**: Container builds (existing)
- **MinIO Client**: MinIO access (default, for local storage)
- **AWS SDK**: S3 access (optional, only if using S3 source or ECR registry)
- **Google Cloud Storage SDK**: GCS access (optional, only if using GCS source)
- **Git Client**: For Git-based sources
- **Container Registry Clients**: Support for ECR, GHCR, GCR, and local registries

### AI Agent Dependencies
- **CloudEvents SDK**: For receiving CloudEvents
- **ML/AI Framework**: For analysis and predictions (optional)
- **Vector Database**: For event pattern storage (optional)
- **Metrics Client**: For observability integration

## 📋 Technical Decision: Kubebuilder Framework

### Decision
**Use Kubebuilder as the framework for building the Kubernetes operator.**

### Context
The operator needs to manage Custom Resource Definitions (CRDs) for `LambdaFunction` resources, implement controllers/reconcilers, and integrate with the Kubernetes API. We evaluated several options for building Kubernetes operators in Go.

### Options Considered

#### 1. Kubebuilder ✅ (Selected)
**Pros:**
- **Industry standard**: Widely adopted by major Kubernetes projects (Istio, cert-manager, etc.)
- **Official Kubernetes project**: Maintained by Kubernetes SIG, ensuring long-term support
- **Mature ecosystem**: Stable API, comprehensive tooling, extensive documentation
- **Code generation**: Automatic CRD generation, RBAC manifests, and webhook scaffolding
- **Controller-runtime integration**: Built on `controller-runtime`, the de facto controller library
- **Developer experience**: Excellent CLI tools, testing utilities, and project scaffolding
- **Webhook support**: Built-in support for validation and mutation webhooks
- **Community**: Large community, extensive examples, and best practices

**Cons:**
- Learning curve for developers new to Kubernetes operators
- Some opinionated project structure (though follows Go best practices)

#### 2. Operator SDK
**Pros:**
- Similar feature set to Kubebuilder
- Good documentation

**Cons:**
- Less widely adopted than Kubebuilder
- Red Hat maintained (vs. Kubernetes SIG maintained)
- Slightly less mature ecosystem

#### 3. Raw controller-runtime
**Pros:**
- Maximum flexibility
- No framework overhead

**Cons:**
- Significant boilerplate code
- Manual CRD generation and management
- More maintenance burden
- Slower development velocity

#### 4. Metacontroller
**Pros:**
- Simpler for basic use cases
- Declarative approach

**Cons:**
- Less flexible for complex operators
- Not suitable for our use case (requires custom build logic)

### Decision Rationale

1. **Alignment with Kubernetes ecosystem**: Kubebuilder is the standard tool for building operators, ensuring compatibility and maintainability.

2. **Project structure**: Our architecture already follows Kubebuilder conventions:
   - `api/v1alpha1/` for CRD definitions
   - `controllers/` for reconciler implementations
   - `cmd/manager/main.go` for entry point
   - `config/crd/bases/` for generated manifests

3. **Long-term support**: As an official Kubernetes project, Kubebuilder receives ongoing maintenance and aligns with Kubernetes evolution.

4. **Developer productivity**: Code generation and scaffolding reduce boilerplate and accelerate development.

5. **Best practices**: Kubebuilder enforces Kubernetes operator best practices out of the box.

### Implementation Notes

- **Project initialization**: `kubebuilder init --domain lambda.knative.io`
- **API scaffolding**: `kubebuilder create api --group lambda --version v1alpha1 --kind LambdaFunction`
- **Controller implementation**: Reconcile loop in `controllers/lambdafunction_controller.go`
- **CRD generation**: Automatic via `make manifests`
- **Webhooks**: Support for validation and mutation webhooks if needed

### References
- [Kubebuilder Documentation](https://book.kubebuilder.io/)
- [controller-runtime Documentation](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Kubernetes Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

## 🐳 Container Registry Configuration

The operator supports multiple container registries for storing built function images. The registry is configurable per LambdaFunction or globally.

### Supported Registries

1. **AWS ECR (Elastic Container Registry)**
   - Format: `{account-id}.dkr.ecr.{region}.amazonaws.com/{repository}:{tag}`
   - Authentication: AWS IAM credentials (via imagePullSecret)
   - Use case: AWS-native deployments

2. **GitHub Container Registry (GHCR)**
   - Format: `ghcr.io/{owner}/{repository}:{tag}`
   - Authentication: GitHub Personal Access Token (PAT) or GITHUB_TOKEN
   - Use case: Open source projects, GitHub Actions integration

3. **Google Container Registry (GCR)**
   - Format: `gcr.io/{project-id}/{repository}:{tag}` or `{region}.gcr.io/{project-id}/{repository}:{tag}`
   - Authentication: Google Service Account JSON key
   - Use case: GCP deployments, Google Cloud workloads

4. **Local Docker Registry**
   - Format: `localhost:5001/{repository}:{tag}` or `{registry-service}.{namespace}.svc.cluster.local:5000/{repository}:{tag}`
   - Authentication: Optional (depends on registry configuration)
   - Use case: Development, testing, air-gapped environments, Kind clusters

### Registry Configuration

**Per-Function Configuration** (in LambdaFunction CRD):
```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaFunction
metadata:
  name: my-function
spec:
  build:
    registry: "ghcr.io/myorg/my-function"  # Full registry path
    imagePullSecret: "ghcr-secret"          # Optional: for private registries
```

**Global Default Configuration** (in Operator ConfigMap):
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: knative-lambda-operator-config
  namespace: knative-lambda
data:
  default-registry: "localhost:5001"  # Default for Kind cluster
  registry-type: "local"              # local, ecr, ghcr, gcr
```

### Registry Selection Logic

1. **Function-specific registry** (if specified in `spec.build.registry`)
2. **Namespace default** (if configured via ConfigMap in namespace)
3. **Global default** (if configured in operator ConfigMap)
4. **Fallback**: `localhost:5001` (for local development)

### Image Pull Secrets

All registries (except public GHCR) require authentication via Kubernetes secrets:

```yaml
# Example: GHCR secret
apiVersion: v1
kind: Secret
metadata:
  name: ghcr-secret
  namespace: knative-lambda
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: <base64-encoded-docker-config>
```

**Secret Format**:
```json
{
  "auths": {
    "ghcr.io": {
      "username": "github-username",
      "password": "ghp_xxxxxxxxxxxx",
      "auth": "<base64-encoded-username:password>"
    }
  }
}
```

### Local Registry Setup (Kind Cluster)

For local development with Kind, a local registry can be configured:

```yaml
# Example: Local registry service in Kind cluster
apiVersion: v1
kind: Service
metadata:
  name: local-registry
  namespace: knative-lambda
spec:
  type: NodePort
  ports:
  - port: 5000
    targetPort: 5000
    nodePort: 30050
  selector:
    app: local-registry
```

**Registry URL**: `localhost:5001` (mapped to Kind node port) or `local-registry.knative-lambda.svc.cluster.local:5000`

### Registry-Specific Considerations

**ECR**:
- Requires AWS credentials with ECR push permissions
- Supports lifecycle policies for cost optimization
- Rate limits: 500 requests/second (can be throttled)

**GHCR**:
- Free for public repositories
- Private repositories: $0.10/GB/month
- Rate limits: 5,000 requests/hour (authenticated)

**GCR**:
- Pricing: $0.026/GB/month
- Supports regional registries for lower latency
- Rate limits: 10,000 requests/minute

**Local Registry**:
- No storage costs (uses cluster storage)
- No rate limits (limited by cluster resources)
- Suitable for development and testing
- Can be backed by persistent volumes for data retention

## 🔐 Security Considerations

1. **RBAC**: Fine-grained permissions for controller
2. **Image Pull Secrets**: Secure registry access
3. **Source Code Security**: Encrypted storage for inline code
4. **Network Policies**: Restrict controller network access
5. **Pod Security Standards**: Enforce security contexts

## 📊 Status Management

### Phases
- **Pending**: CRD created, waiting for processing
- **Building**: Kaniko job running
- **Deploying**: Knative service being created
- **Ready**: Service deployed and healthy
- **Failed**: Error occurred
- **Deleting**: Cleanup in progress

### Conditions
- **SourceReady**: Source code available
- **BuildReady**: Image built successfully
- **DeployReady**: Service deployed
- **ServiceReady**: Service healthy and ready
- **BrokerReady**: RabbitMQ Broker configured
- **TriggersReady**: Knative Triggers created
- **DLQReady**: Dead Letter Queue configured
- **AIAgentReady**: AI Agent Trigger and Service configured

## 🔄 Complete Event Flow with DLQ and AI Agent

```
┌─────────────────────────────────────────────────────────────┐
│     Complete Event Processing Flow (with AI Agent)          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. EVENT INGESTION                                         │
│     ┌──────────────┐                                        │
│     │ CloudEvent   │                                        │
│     │ Source       │                                        │
│     └──────┬───────┘                                        │
│            │                                                │
│            │ POST /api/events                               │
│            │ (any of 17 event types)                        │
│            ▼                                                │
│     ┌──────────────┐                                        │
│     │ Knative      │                                        │
│     │ Broker       │                                        │
│     │ (RabbitMQ)   │                                        │
│     └──────┬───────┘                                        │
│            │                                                │
│  2. EVENT ROUTING (Multi-Subscriber)                        │
│            │                                                │
│            │ Events routed in parallel to:                  │
│            │                                                │
│    ┌───────┼───────┬──────────┐                             │
│    │       │       │          │                             │
│    ▼       ▼       ▼          ▼                             │
│  ┌────┐ ┌────┐ ┌────┐ ┌──────────┐                          │
│  │Trig│ │Trig│ │Trig│ │ AI Agent │                          │
│  │(Op)│ │(LF)│ │(Op)│ │ Trigger  │                          │
│  │Ctrl│ │Svc │ │Stat│ │ (ALL 17) │                          │
│  └────┘ └────┘ └────┘ └────┬─────┘                          │
│    │       │       │        │                               │
│    │       │       │        │                               │
│    ▼       ▼       ▼        ▼                               │
│  ┌────┐ ┌────┐ ┌────┐ ┌──────────┐                          │
│  │Op  │ │LF  │ │Op  │ │ AI Agent │                          │
│  │Ctrl│ │Svc │ │Ctrl│ │ Service  │                          │
│  └────┘ └────┘ └────┘ └────┬─────┘                          │
│                              │                              │
│  3. AI AGENT PROCESSING (Parallel)                          │
│                              │                              │
│                              │ Receives ALL events          │
│                              │ Analyzes patterns            │
│                              │ Detects anomalies            │
│                              │ Generates insights           │
│                              ▼                              │
│     ┌──────────────┐                                        │
│     │ AI Agent     │                                        │
│     │ Analysis     │                                        │
│     │ ──────────── │                                        │
│     │ - Patterns   │                                        │
│     │ - Anomalies  │                                        │
│     │ - Insights   │                                        │
│     │ - Actions    │                                        │
│     └──────────────┘                                        │
│            │                                                │
│  4. EVENT DELIVERY TO LAMBDA                                │
│            │                                                │
│            │ CloudEvent to LambdaFunction                   │
│            ▼                                                │
│     ┌───────────────┐                                       │
│     │ LambdaFunction│                                       │
│     │ Service       │                                       │
│     │ (Knative)     │                                       │
│     └──────┬────────┘                                       │
│            │                                                │
│     ┌──────┴──────┐                                         │
│     │             │                                         │
│     ▼             ▼                                         │
│  Success      Failure                                       │
│     │             │                                         │
│     │             │                                         │
│     │     5. RETRY LOGIC                                    │
│     │             │                                         │
│     │             │ Retry 1-5 (exponential backoff)         │
│     │             │                                         │
│     │             ▼                                         │
│     │     ┌──────────────┐                                  │
│     │     │ All Retries  │                                  │
│     │     │ Exhausted?   │                                  │
│     │     └───┬──────┬───┘                                  │
│     │         │      │                                      │
│     │     ┌───┘      └───┐                                  │
│     │     │              │                                  │
│     │     ▼              ▼                                  │
│     │  Success     6. DLQ ROUTING                           │
│     │                  │                                    │
│     │                  ▼                                    │
│     │          ┌──────────────┐                             │
│     │          │ DLQ Exchange │                             │
│     │          │ (knative-    │                             │
│     │          │  lambda-dlq) │                             │
│     │          └──────┬───────┘                             │
│     │                 │                                     │
│     │                 │ Routing Key: dlq.*                  │
│     │                 ▼                                     │
│     │          ┌──────────────┐                             │
│     │          │ DLQ Binding │                              │
│     │          └──────┬───────┘                             │
│     │                 │                                     │
│     │                 ▼                                     │
│     │          ┌──────────────┐                             │
│     │          │ DLQ Queue    │                             │
│     │          │ ──────────── │                             │
│     │          │ TTL: 7 days  │                             │
│     │          │ Max: 50,000  │                             │
│     │          └──────┬───────┘                             │
│     │                 │                                     │
│     │         7. DLQ PROCESSING                             │
│     │                 │                                     │
│     │                 ▼                                     │
│     │          ┌──────────────┐                             │
│     │          │ DLQ Handler  │                             │
│     │          │ Service      │                             │
│     │          │ ──────────── │                             │
│     │          │ - Alert      │                             │
│     │          │ - Monitor    │                             │
│     │          │ - Manual     │                             │
│     │          │   retry      │                             │
│     │          └──────────────┘                             │
│     │                                                       │
│     └───────────────────────────────────────────────────────┘
│                                                             │
│  Note: AI Agent receives ALL 17 event types in parallel     │
│        with Operator Controller and LambdaFunction services │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Migration Path

### Phase 1: Operator as Event Receiver
- Deploy operator as Knative Service
- Operator receives CloudEvents via Broker/Triggers
- Operator creates/manages LambdaFunction CRDs from events
- Existing builder service continues to work
- Both paths create CRDs for state management

### Phase 2: Hybrid Operation
- Operator handles CloudEvents AND CRD reconciliation
- CRD creation triggers same build/deploy flow
- Event-driven creation creates CRD automatically
- Status synchronized between events and CRD

### Phase 3: Full Event-Driven Operator
- Operator is primary event processor
- All operations go through CloudEvents
- CRD serves as state representation
- Builder service becomes optional/adapter layer

## ✅ Benefits of Operator Pattern

1. **Declarative**: Desired state management
2. **Self-Healing**: Automatic reconciliation
3. **GitOps Friendly**: CRDs work with GitOps tools
4. **Better Observability**: Status in CRD
5. **Simpler API**: Kubernetes-native interface
6. **Multi-tenancy**: Namespace isolation
7. **Lifecycle Management**: Built-in finalizers

## 🔄 Comparison: Event-Driven vs Operator

| Aspect | Event-Driven (Current) | Operator (Hybrid) |
|--------|----------------------|-------------------|
| **Interface** | CloudEvents/HTTP | CRD + CloudEvents |
| **Event Reception** | Builder Service | Operator Controller |
| **State Management** | External (RabbitMQ) | Kubernetes API (CRD) |
| **Reconciliation** | Manual/Event-based | Automatic + Event-driven |
| **GitOps** | Requires custom tooling | Native support |
| **Observability** | External metrics | CRD status + Events |
| **Multi-tenancy** | Application-level | Namespace-based |
| **Lifecycle** | Event-driven | Declarative + Event-driven |
| **Broker/Triggers** | Manual creation | Operator-managed |
| **DLQ** | Manual setup | Operator-managed |

## 📝 Example Usage

### Using CRD (Declarative)

```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaFunction
metadata:
  name: my-function
  namespace: lambda-dev
spec:
  source:
    type: minio  # Default: MinIO for local development
    minio:
      endpoint: minio.minio.svc.cluster.local:9000
      bucket: knative-lambda-functions
      key: functions/my-function.zip
    # Alternative: AWS S3
    # type: s3
    # s3:
    #   bucket: my-code-bucket
    #   key: functions/my-function.zip
    #   region: us-west-2
    # Alternative: Google Cloud Storage
    # type: gcs
    # gcs:
    #   bucket: my-code-bucket
    #   key: functions/my-function.zip
    #   project: my-gcp-project
  runtime:
    language: nodejs
    version: "22"
    handler: index.handler
  scaling:
    minReplicas: 0
    maxReplicas: 10
    targetConcurrency: 5
  resources:
    requests:
      memory: "64Mi"
      cpu: "50m"
    limits:
      memory: "128Mi"
      cpu: "100m"
  env:
  - name: API_KEY
    valueFrom:
      secretKeyRef:
        name: api-secrets
        key: api-key
  triggers:
  - broker: knative-lambda-broker-dev
    filter:
      type: my.event.type
      source: my.event.source
status:
  phase: Ready
  buildStatus:
    imageURI: registry.example.com/my-function:abc123
  serviceStatus:
    serviceName: my-function
    url: http://my-function.lambda-dev.svc.cluster.local
    ready: true
    replicas: 2
```

### Using HTTP API

#### Create Function

```bash
curl -X POST http://operator.lambda.svc.cluster.local/api/v1/lambda/functions \
  -H "Content-Type: application/json" \
  -d '{
    "spec": {
      "source": {
        "type": "minio",
        "minio": {
          "endpoint": "minio.minio.svc.cluster.local:9000",
          "bucket": "knative-lambda-functions",
          "key": "functions/my-function.zip"
        }
      },
      "runtime": {
        "language": "nodejs",
        "version": "22",
        "handler": "index.handler"
      },
      "scaling": {
        "minReplicas": 0,
        "maxReplicas": 10,
        "targetConcurrency": 5
      }
    }
  }'
```

**Response:**
```json
{
  "name": "my-function",
  "status": "Pending",
  "message": "LambdaFunction created, build starting"
}
```

#### Update Function

```bash
curl -X PUT http://operator.lambda.svc.cluster.local/api/v1/lambda/functions/my-function \
  -H "Content-Type: application/json" \
  -d '{
    "spec": {
      "scaling": {
        "maxReplicas": 20
      }
    }
  }'
```

**Response:**
```json
{
  "name": "my-function",
  "status": "Ready",
  "message": "LambdaFunction updated successfully"
}
```

#### Delete Function

```bash
curl -X DELETE http://operator.lambda.svc.cluster.local/api/v1/lambda/functions/my-function
```

**Response:**
```json
{
  "name": "my-function",
  "status": "Deleting",
  "message": "LambdaFunction deletion initiated"
}
```

#### Using CloudEvents Endpoint

```bash
curl -X POST http://operator.lambda.svc.cluster.local/api/v1/events \
  -H "Content-Type: application/cloudevents+json" \
  -d '{
    "type": "lambda.function.created",
    "source": "api.gateway",
    "id": "12345",
    "time": "2025-01-20T10:00:00Z",
    "data": {
      "spec": {
        "source": {
          "type": "minio",
          "minio": {
            "endpoint": "minio.minio.svc.cluster.local:9000",
            "bucket": "knative-lambda-functions",
            "key": "functions/my-function.zip"
          }
        },
        "runtime": {
          "language": "nodejs",
          "version": "22"
        }
      }
    }
  }'
```

## 🎯 Next Steps

1. **Proof of Concept**: Build minimal operator with basic CRD
2. **API Server**: Implement HTTP API endpoints (create/update/delete)
3. **Event Reception**: Implement CloudEvent handler (HTTP + Broker)
4. **Build Manager**: Implement Kaniko job creation
5. **Deploy Manager**: Implement Knative service creation
6. **Event Manager**: Implement Broker/Trigger/DLQ management
7. **Status Management**: Implement status updates
8. **Testing**: Unit and integration tests
9. **Documentation**: User guides and API reference
10. **Migration Guide**: From events to CRDs

## 📡 Operator as Knative Service (Event-Driven)

### Operator Deployment

The operator itself is deployed as a **Knative Service** that receives CloudEvents:

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: knative-lambda-operator
  namespace: knative-lambda
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"  # Always running
        autoscaling.knative.dev/maxScale: "10"
    spec:
      containers:
      - image: localhost:5001/knative-lambda-operator:latest
        env:
        - name: KUBERNETES_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
```

### Operator Triggers Setup

The operator needs its own triggers to receive events. Additionally, the AI Agent has a dedicated trigger to receive all events.

#### Operator Triggers

```yaml
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: operator-function-create
spec:
  broker: knative-lambda-operator-broker
  filter:
    attributes:
      type: lambda.function.created
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: knative-lambda-operator
---
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: operator-build-events
spec:
  broker: knative-lambda-operator-broker
  filter:
    attributes:
      # Build Events
      type: lambda.build.started
      type: lambda.build.completed
      type: lambda.build.failed
      type: lambda.build.timeout
      type: lambda.build.cancelled
      type: lambda.build.stopped
      # Service Events
      type: lambda.service.created
      type: lambda.service.updated
      type: lambda.service.deleted
      # Status Events
      type: lambda.status.updated
      type: lambda.health.check
      # Parser Events (optional)
      type: lambda.parser.started
      type: lambda.parser.completed
      type: lambda.parser.failed
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: knative-lambda-operator
---
# AI Agent Trigger - Receives ALL 17 event types
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: ai-agent-all-events
  namespace: knative-lambda
  annotations:
    rabbitmq.eventing.knative.dev/parallelism: "50"
    description: "Routes all LambdaFunction events to AI Agent for analysis"
spec:
  broker: knative-lambda-operator-broker
  filter:
    attributes:
      # Function Management Events
      type: lambda.function.created
      type: lambda.function.updated
      type: lambda.function.deleted
      # Build Events
      type: lambda.build.started
      type: lambda.build.completed
      type: lambda.build.failed
      type: lambda.build.timeout
      type: lambda.build.cancelled
      type: lambda.build.stopped
      # Service Events
      type: lambda.service.created
      type: lambda.service.updated
      type: lambda.service.deleted
      # Status Events
      type: lambda.status.updated
      type: lambda.health.check
      # Parser Events
      type: lambda.parser.started
      type: lambda.parser.completed
      type: lambda.parser.failed
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: knative-lambda-ai-agent
      namespace: knative-lambda
  delivery:
    retry: 5
    backoffPolicy: exponential
    backoffDelay: PT1S
    timeout: PT30S
```

### Complete Event Flow: External → Operator → LambdaFunction

```
┌─────────────────────────────────────────────────────────────┐
│         Complete Event-Driven Operator Flow                 │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Notifi Scheduler Service (CloudEvent Source)               │
│         │                                                   │
│         │ POST CloudEvent                                   │
│         │ Types:                                            │
│         │ - lambda.function.created (create)                │
│         │ - lambda.function.updated (update)                │
│         │ - lambda.function.deleted (delete)                │
│         │ - lambda.build.started (start)                    │
│         │ - lambda.build.stopped (stop)                     │
│         │ data: { source: {...}, runtime: {...} }           │
│         │                                                   │
│         │ See: NOTIFI_INTEGRATION.md for details            │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ Operator Broker      │                                   │
│  │ (RabbitMQ Broker)    │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         │ Event routing                                     │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ Operator Trigger     │                                   │
│  │ (function.create)    │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         │ CloudEvent delivery                               │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ Operator Controller  │                                   │
│  │ (Knative Service)    │                                   │
│  │ ──────────────────── │                                   │
│  │ 1. Receive CloudEvent│                                   │
│  │ 2. Parse event data  │                                   │
│  │ 3. Create/Update CRD │                                   │
│  │ 4. Trigger build     │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         │ Create LambdaFunction CRD                         │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ LambdaFunction CRD   │                                   │
│  │ Status: Pending      │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         │ Operator reconciliation                           │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ Build Manager        │                                   │
│  │ - Create Kaniko Job  │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         │ Emit: lambda.build.started                        │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ Broker (for builds)  │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│    ┌────┴────┐                                              │
│    │         │                                              │
│    ▼         ▼                                              │
│  Success   Failure                                          │
│    │         │                                              │
│    │         │ Emit: lambda.build.failed                    │
│    │         │ Emit: lambda.build.timeout                   │
│    │         │ Emit: lambda.build.cancelled                 │
│    │                                                        │
│    │ Emit: lambda.build.completed                           │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ Operator (receives)  │                                   │
│  │ - Update CRD status  │                                   │
│  │ - Create Knative Svc │                                   │
│  │ - Create Triggers    │                                   │
│  │ - Create DLQ         │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         │ Emit: lambda.service.created                      │
│         │ Emit: lambda.status.updated                       │
│         │                                                   │
│         │ Events also routed to:                            │
│         │ ▼                                                 │
│  ┌──────────────────────┐                                   │
│  │ AI Agent Service     │                                   │
│  │ ──────────────────── │                                   │
│  │ - Receives all events│                                   │
│  │ - Analyzes patterns  │                                   │
│  │ - Generates insights │                                   │
│  │ - Takes actions      │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         │                                                   │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ LambdaFunction Ready │                                   │
│  │ - Service running    │                                   │
│  │ - Triggers active    │                                   │
│  │ - DLQ configured     │                                   │
│  │ - AI Agent monitoring│                                   │
│  └──────────────────────┘                                   │
└─────────────────────────────────────────────────────────────┘
```

## 📊 Observability & Monitoring: First-Class Citizen

Observability is a **first-class citizen** in the Knative Lambda architecture. Every component exposes comprehensive metrics, traces, and logs, with automatic integration to Prometheus, Grafana, Alertmanager, and distributed tracing systems. The AI Agent actively monitors, investigates alerts, and provides intelligent insights.

### Observability Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│         Observability Stack: First-Class Integration        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. METRICS COLLECTION                                      │
│     ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│     │ Operator     │  │ Lambda Funcs │  │ AI Agent     │    │
│     │ Metrics      │  │ Metrics      │  │ Metrics      │    │
│     └──────┬───────┘  └──────┬───────┘  └──────┬───────┘    │
│            │                  │                │            │
│            └──────────────────┼────────────────┘            │
│                               │                             │
│                               ▼                             │
│     ┌──────────────────────────────────────┐                │
│     │ Prometheus                           │                │
│     │ - Scrapes all metrics                │                │
│     │ - Stores time-series data            │                │
│     │ - Evaluates PrometheusRules          │                │
│     └──────────────┬───────────────────────┘                │
│                    │                                        │
│                    ├──────────────────┐                     │
│                    │                  │                     │
│                    ▼                  ▼                     │
│     ┌──────────────────┐  ┌──────────────────┐              │
│     │ Alertmanager     │  │ Grafana          │              │
│     │ - Receives alerts│  │ - Dashboards     │              │
│     │ - Routes alerts  │  │ - Visualization  │              │
│     └──────────┬───────┘  └──────────────────┘              │
│                │                                            │
│                ▼                                            │
│     ┌──────────────────────────────────────┐                │
│     │ AI Agent (Alert Investigation)       │                │
│     │ - Receives alerts from Alertmanager  │                │
│     │ - Investigates PrometheusRules       │                │
│     │ - Queries Prometheus for context     │                │
│     │ - Provides insights & recommendations│                │
│     └──────────────────────────────────────┘                │
│                                                             │
│  2. DISTRIBUTED TRACING (Fully Integrated)                  │
│     ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│     │ Operator     │  │ AI Agent     │  │ Lambda Funcs │    │
│     │ (Go/OTel)    │  │ (Python/OTel)│  │ (Knative)    │    │
│     └──────┬───────┘  └──────┬───────┘  └──────┬───────┘    │
│            │                 │                  │           │
│     ┌──────┴───────┐  ┌──────┴───────┐  ┌──────┴───────┐    │
│     │ Notifi       │  │ Linkerd      │  │ W3C Trace    │    │
│     │ Services     │  │ Proxy Spans  │  │ Context      │    │
│     │ (OTel)       │  │ (Automatic)  │  │ Propagation  │    │
│     └──────┬───────┘  └──────┬───────┘  └──────┬───────┘    │
│            │                 │                 │            │
│            └─────────┬───────┴─────────────────┘            │
│                      │                                      │
│                      ▼                                      │
│     ┌──────────────────────────────────────┐                │
│     │ Tempo (Unified Backend)              │                │
│     │ - All traces: Operator, AI Agent,    │                │
│     │   Lambda Functions, Notifi Services, │                │
│     │   Linkerd Proxies                    │                │
│     │ - W3C Trace Context correlation      │                │
│     │ - End-to-end request flows           │                │
│     └──────────────────────────────────────┘                │
│                      │                                      │
│                      │ (Optional Fallback)                  │
│                      ▼                                      │
│     ┌──────────────────────────────────────┐                │
│     │ Logfire (AI Agent Only - Optional)   │                │
│     │ - Specialized Python/LLM observability│               │
│     │ - Only if LOGFIRE_ENABLED=true       │                │
│     └──────────────────────────────────────┘                │
│                                                             │
│  3. LOGS                                                    │
│     ┌──────────────┐  ┌──────────────┐                      │
│     │ Operator     │  │ Lambda Funcs │                      │
│     │ Logs         │  │ Logs         │                      │
│     └──────┬───────┘  └──────┬───────┘                      │
│            │                 │                              │
│            └─────────┬───────┘                              │
│                      │                                      │
│                      ▼                                      │
│     ┌──────────────────────────────────────┐                │
│     │ Grafana Alloy                        │                │
│     │ - Log collection                     │                │
│     │ - Kubernetes discovery               │                │
│     │ - Label extraction                   │                │
│     └──────┬───────────────────────────────┘                │
│            │                                                │
│            ▼                                                │
│     ┌──────────────────────────────────────┐                │
│     │ Loki                                 │                │
│     │ - Log aggregation                    │                │
│     │ - LogQL queries                      │                │
│     │ - Object storage backend             │                │
│     └──────────────────────────────────────┘                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Metrics Exposition

The operator exposes Prometheus metrics at the standard endpoint `:8080/metrics`. All metrics follow the naming convention: `operator_<component>_<metric_type>_<unit>`.

#### Key Metrics

**Reconciliation Metrics**:
- `operator_reconcile_total` (Counter) - Total reconciliation attempts
  - Labels: `function_name`, `namespace`, `phase`, `result` (success/failure)
- `operator_reconcile_duration_seconds` (Histogram) - Reconciliation latency
  - Labels: `function_name`, `namespace`, `phase`
  - Buckets: `[0.1, 0.5, 1, 2, 5, 10, 30, 60]`

**CRD Operations Metrics**:
- `operator_crd_operations_total` (Counter) - CRD create/update/delete counts
  - Labels: `operation` (create/update/delete), `function_name`, `namespace`
- `operator_crd_operation_duration_seconds` (Histogram) - CRD operation latency
  - Labels: `operation`, `function_name`

**Event Processing Metrics**:
- `operator_event_processing_total` (Counter) - Total CloudEvents processed
  - Labels: `event_type`, `source`, `result` (success/failure)
- `operator_event_processing_duration_seconds` (Histogram) - CloudEvent processing time
  - Labels: `event_type`, `source`
- `operator_event_queue_depth` (Gauge) - Pending events in queue
  - Labels: `event_type`

**Build Metrics**:
- `operator_build_jobs_active` (Gauge) - Active Kaniko jobs
  - Labels: `function_name`, `namespace`
- `operator_build_jobs_total` (Counter) - Total build jobs created
  - Labels: `function_name`, `namespace`, `result` (success/failure)
- `operator_build_duration_seconds` (Histogram) - Build duration
  - Labels: `function_name`, `namespace`

**Deployment Metrics**:
- `operator_knative_services_managed` (Gauge) - Number of managed Knative Services
  - Labels: `namespace`, `status` (ready/pending/failed)
- `operator_service_creation_duration_seconds` (Histogram) - Service creation time
  - Labels: `function_name`, `namespace`

**DLQ Metrics**:
- `operator_dlq_depth` (Gauge) - DLQ queue depth
  - Labels: `function_name`, `namespace`
- `operator_dlq_messages_total` (Counter) - Total messages sent to DLQ
  - Labels: `function_name`, `namespace`, `reason`

**API Metrics**:
- `operator_api_requests_total` (Counter) - HTTP API request counts
  - Labels: `method`, `endpoint`, `status_code`
- `operator_api_request_duration_seconds` (Histogram) - API latency
  - Labels: `method`, `endpoint`
- `operator_api_request_size_bytes` (Histogram) - Request size
  - Labels: `method`, `endpoint`

**System Metrics**:
- `operator_goroutines` (Gauge) - Number of goroutines
- `operator_memory_usage_bytes` (Gauge) - Memory usage
- `operator_cpu_usage_seconds` (Counter) - CPU usage

**Linkerd Metrics** (Automatic for all meshed services):
- `linkerd_response_total` (Counter) - Total responses per route
  - Labels: `deployment`, `namespace`, `route`, `classification` (success/failure)
- `linkerd_response_latency_ms` (Histogram) - Response latency
  - Labels: `deployment`, `namespace`, `route`
  - Percentiles: P50, P95, P99
- `linkerd_retry_total` (Counter) - Retry attempts and successes
  - Labels: `deployment`, `namespace`, `route`, `classification`
- `linkerd_circuit_breaker_state` (Gauge) - Circuit breaker state (0=closed, 1=open, 2=half-open)
  - Labels: `deployment`, `namespace`

**AI Agent Metrics**:
- `ai_agent_events_processed_total` (Counter) - Total events processed
  - Labels: `event_type`, `result` (success/failure)
- `ai_agent_embedding_generation_duration_seconds` (Histogram) - Embedding generation time
- `ai_agent_similarity_search_duration_seconds` (Histogram) - Similarity search time
- `ai_agent_patterns_detected_total` (Counter) - Patterns detected
  - Labels: `pattern_type` (success/failure/anomaly)
- `ai_agent_alerts_investigated_total` (Counter) - Alerts investigated
  - Labels: `alert_name`, `severity`, `result` (resolved/unresolved)
- `ai_agent_investigation_duration_seconds` (Histogram) - Investigation time
- `ai_agent_investigation_queue_depth` (Gauge) - Pending investigations
- `ai_agent_prometheus_queries_total` (Counter) - Prometheus queries executed
  - Labels: `query_type` (instant/range)

**Lambda Function Metrics** (via Linkerd):
- `linkerd_response_total{deployment=~"lambda-.*"}` - Function invocation metrics
- `linkerd_response_latency_ms{deployment=~"lambda-.*"}` - Function execution latency
- `linkerd_circuit_breaker_state{deployment=~"lambda-.*"}` - Function circuit breaker state

### Example Prometheus Queries

```promql
# Reconciliation success rate
rate(operator_reconcile_total{result="success"}[5m]) / 
rate(operator_reconcile_total[5m]) * 100

# P95 reconciliation latency
histogram_quantile(0.95, 
  rate(operator_reconcile_duration_seconds_bucket[5m])
)

# Active build jobs
sum(operator_build_jobs_active)

# DLQ depth alert
operator_dlq_depth > 100

# API error rate
rate(operator_api_requests_total{status_code=~"5.."}[5m]) / 
rate(operator_api_requests_total[5m]) * 100

# Lambda function error rate (via Linkerd)
sum(rate(linkerd_response_total{deployment=~"lambda-.*",classification="failure"}[5m])) by (deployment) /
sum(rate(linkerd_response_total{deployment=~"lambda-.*"}[5m])) by (deployment) * 100

# Lambda function P95 latency
histogram_quantile(0.95,
  sum(rate(linkerd_response_latency_ms_bucket{deployment=~"lambda-.*"}[5m])) by (deployment, le)
)

# Circuit breaker state for Lambda functions
linkerd_circuit_breaker_state{deployment=~"lambda-.*"}

# AI Agent events processed rate
rate(ai_agent_events_processed_total[5m])

# AI Agent investigation queue depth
ai_agent_investigation_queue_depth

# AI Agent investigation duration (P95)
histogram_quantile(0.95,
  rate(ai_agent_investigation_duration_seconds_bucket[5m])
)
```

### Distributed Tracing

The operator integrates with **OpenTelemetry** for fully integrated distributed tracing across all components: **Knative services, AI Agent, Linkerd proxies, and Notifi services**. All traces are sent to **Tempo (primary backend)** with **Logfire as optional fallback** for AI Agent specialized observability.

#### Linkerd Distributed Tracing Integration

Linkerd automatically emits trace spans from proxies when it sees tracing headers in proxied HTTP requests. This provides visibility into request latency within the Linkerd proxy layer without requiring application changes.

**Linkerd Configuration**:
- Linkerd proxies automatically detect W3C Trace Context headers (`traceparent`, `tracestate`)
- When headers are present, Linkerd emits spans showing time spent in the proxy
- No code changes required for Linkerd tracing - works automatically for meshed services

**How It Works**:
1. Ingress layer (or first service) injects trace headers to start a trace
2. Linkerd proxy sees headers and emits spans for proxy processing time
3. Headers are propagated to downstream services
4. All services (operator, AI Agent, Lambda functions) participate in the trace

**Reference**: [Linkerd Distributed Tracing Documentation](https://linkerd.io/2-edge/features/distributed-tracing/)

#### Operator Tracing (Go - OpenTelemetry)

**Configuration**:
```yaml
observability:
  tracing:
    enabled: true
    otlpEndpoint: "tempo-distributor.tempo.svc.cluster.local:4317"
    sampleRate: 0.1  # 10% sampling
    attributes:
      service.name: "knative-lambda-operator"
      service.version: "v1.0.0"
      deployment.environment: "production"
```

**Trace Context Propagation**:
- W3C Trace Context headers propagated through CloudEvents
- Correlation IDs included in all spans
- Parent-child span relationships for reconciliation flow
- Linkerd proxy spans automatically included when requests flow through mesh

**Key Spans**:
- `operator.reconcile` - Main reconciliation span
- `operator.build.create` - Build job creation
- `operator.service.create` - Knative Service creation
- `operator.event.process` - CloudEvent processing
- `operator.api.handle` - HTTP API request handling
- `linkerd.proxy` - Automatic Linkerd proxy spans (when headers present)

**Span Attributes**:
- `function.name` - LambdaFunction name
- `function.namespace` - Namespace
- `function.phase` - Current phase
- `event.type` - CloudEvent type
- `correlation.id` - Correlation ID

#### AI Agent Tracing (Python - OpenTelemetry/Tempo Primary, Logfire Fallback)

The AI Agent uses **OpenTelemetry with Tempo as the primary tracing backend** for full integration with the rest of the system. **Logfire is configured as an optional fallback** for specialized Python/LLM observability when needed.

**Why OpenTelemetry/Tempo Primary**:
- **Full Integration**: All traces (Operator, AI Agent, Linkerd, Notifi services) in one unified backend
- **End-to-End Correlation**: Complete request flows visible across all components
- **Consistent Observability**: Same tracing infrastructure for all services
- **Grafana Integration**: Unified trace visualization in Grafana
- **Cost-Effective**: Self-hosted Tempo in Kubernetes cluster

**Why Logfire as Fallback**:
- **Python-Native**: Purpose-built for Python applications with excellent Pydantic integration
- **LLM Observability**: Specialized features for tracking LLM calls, embeddings, and AI operations
- **Cloud-Managed**: No infrastructure to manage, reliable cloud service
- **Optional**: Used only when specialized Python/LLM observability is needed

**Configuration (OpenTelemetry/Tempo Primary)**:
```python
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.httpx import HTTPXClientInstrumentor
from opentelemetry.propagators.composite import CompositeHTTPPropagator
from opentelemetry.propagate import set_global_propagator

# Configure resource attributes
resource = Resource.create({
    "service.name": "knative-lambda-ai-agent",
    "service.version": os.getenv("VERSION", "v1.0.0"),
    "deployment.environment": os.getenv("ENVIRONMENT", "production"),
    "service.namespace": "knative-lambda"
})

# Configure TracerProvider with Tempo (primary)
tracer_provider = TracerProvider(resource=resource)
tempo_exporter = OTLPSpanExporter(
    endpoint="tempo-distributor.tempo.svc.cluster.local:4317",
    insecure=True  # For internal cluster communication
)
tracer_provider.add_span_processor(BatchSpanProcessor(tempo_exporter))
trace.set_tracer_provider(tracer_provider)

# Configure W3C Trace Context propagation
from opentelemetry.propagators.tracecontext import TraceContextTextMapPropagator
set_global_propagator(TraceContextTextMapPropagator())

# Instrument FastAPI app
FastAPIInstrumentor.instrument_app(app)
HTTPXClientInstrumentor().instrument()

# Automatic tracing for async functions
tracer = trace.get_tracer(__name__)

@tracer.start_as_current_span("process_event")
async def process_event(event: CloudEvent):
    span = trace.get_current_span()
    span.set_attribute("event.type", event.type)
    span.set_attribute("event.source", event.source)
    # Automatic span creation with timing
    pass

# Custom spans with attributes
@tracer.start_as_current_span("investigate_alert")
async def investigate_alert(alert: Alert):
    span = trace.get_current_span()
    span.set_attribute("alert.name", alert.name)
    span.set_attribute("alert.severity", alert.severity)
    span.set_attribute("function.name", alert.labels.get("function_name"))
    # Span includes custom attributes
    pass
```

**Optional Logfire Configuration (Fallback)**:
```python
import logfire
from logfire import with_attributes

# Initialize Logfire (optional, only if needed)
if os.getenv("LOGFIRE_ENABLED", "false").lower() == "true":
    logfire.configure(
        token=os.getenv("LOGFIRE_TOKEN"),
        service_name="knative-lambda-ai-agent",
        environment=os.getenv("ENVIRONMENT", "production")
    )
    # Instrument FastAPI app (optional)
    logfire.instrument_fastapi(app)
```

**OpenTelemetry Integration Points**:
- **Event Processing**: Automatic spans for CloudEvent handlers with W3C Trace Context
- **Alert Investigation**: Detailed spans for Prometheus queries, analysis, recommendations
- **Embedding Generation**: Track embedding model calls and latency
- **Similarity Search**: Trace LanceDB vector searches
- **LLM Operations**: If using LLMs for recommendations, automatic LLM call tracking
- **Notifi Service Calls**: All gRPC/HTTP calls to Notifi services include trace context
- **Linkerd Integration**: Automatic proxy spans when W3C Trace Context headers are present

**Parallel Export to Logfire** (Optional Fallback):
```python
# Configure OpenTelemetry to export to both Tempo (primary) and Logfire (optional fallback)
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

# Tempo (primary backend)
tempo_exporter = OTLPSpanExporter(
    endpoint="tempo-distributor.tempo.svc.cluster.local:4317",
    insecure=True
)

# Logfire (optional fallback, only if enabled)
logfire_exporter = None
if os.getenv("LOGFIRE_ENABLED", "false").lower() == "true":
    logfire_exporter = OTLPSpanExporter(
        endpoint="https://logfire-api.pydantic.dev:443",
        headers={"Authorization": f"Bearer {os.getenv('LOGFIRE_TOKEN')}"}
    )

# Multi-export: send to Tempo (always) and Logfire (optional)
tracer_provider = TracerProvider(resource=resource)
tracer_provider.add_span_processor(BatchSpanProcessor(tempo_exporter))
if logfire_exporter:
    tracer_provider.add_span_processor(BatchSpanProcessor(logfire_exporter))
trace.set_tracer_provider(tracer_provider)
```

**Trace Correlation**:
- All traces use W3C Trace Context headers (`traceparent`, `tracestate`) for end-to-end correlation
- AI Agent traces include `trace_id` that matches Linkerd/operator/Notifi service traces
- W3C Trace Context headers propagated across: Ingress → Linkerd → Operator → AI Agent → Lambda Functions → Notifi Services
- Full request flow visible across all components in unified Tempo backend

#### Notifi Services Tracing Integration

**Notifi Services OpenTelemetry Configuration**:
All Notifi services (Scheduler, Storage Manager, Subscription Manager, Fetch Proxy, Blockchain Manager) are instrumented with OpenTelemetry and export traces to Tempo:

```yaml
# Example: Scheduler Service OpenTelemetry configuration
observability:
  tracing:
    enabled: true
    otlpEndpoint: "tempo-distributor.tempo.svc.cluster.local:4317"
    sampleRate: 0.1
    attributes:
      service.name: "notifi-scheduler"
      service.namespace: "notifi"
```

**Trace Context Propagation**:
- Lambda functions include W3C Trace Context headers in all gRPC/HTTP calls to Notifi services
- Notifi services extract and propagate trace context through their internal processing
- Linkerd proxies automatically emit spans for all meshed traffic
- Complete end-to-end traces: Lambda Function → Linkerd → Notifi Service → Linkerd → Response

#### Knative Services Tracing Integration

**Knative Service Tracing**:
- Knative Services (Lambda functions) automatically propagate W3C Trace Context headers
- Linkerd proxies inject trace context if not present
- All Lambda function invocations include trace context
- Traces show: CloudEvent → Operator → Knative Service → Lambda Function → Notifi Services

#### Trace Backend Architecture

**Primary Backend: Tempo (Unified)**
- Stores traces from all components: Operator (Go), AI Agent (Python), Linkerd proxies, Notifi services, Knative services
- Self-hosted in Kubernetes cluster
- Integrated with Grafana for visualization
- OTLP gRPC endpoint: `tempo-distributor.tempo.svc.cluster.local:4317`
- **Single source of truth** for all distributed traces

**Optional Fallback: Logfire (AI Agent Only)**
- Optional cloud-managed service for specialized Python/LLM observability
- Only enabled when `LOGFIRE_ENABLED=true` environment variable is set
- OTLP HTTPS endpoint: `https://logfire-api.pydantic.dev:443`
- Authentication via API token
- **Not required** - Tempo provides full tracing capabilities

**Linkerd Proxy Spans**
- Automatically included when W3C Trace Context headers present
- Shows proxy processing time (mTLS, load balancing, retries)
- Exported to Tempo via OTLP collector
- No code changes required - works automatically

### Structured Logging

**Log Format**: JSON structured logs with correlation IDs and trace context.

**Log Levels**:
- `DEBUG`: Detailed diagnostic information
- `INFO`: General informational messages
- `WARN`: Warning messages for potential issues
- `ERROR`: Error messages for failures

**Log Fields**:
```json
{
  "level": "info",
  "ts": "2025-01-20T10:00:00Z",
  "caller": "controller/reconciler.go:123",
  "msg": "Reconciling LambdaFunction",
  "correlation_id": "abc123-def456",
  "trace_id": "a1b2c3d4e5f6",
  "span_id": "b2c3d4e5f6",
  "function_name": "my-function",
  "namespace": "lambda-dev",
  "phase": "Building"
}
```

#### Linkerd HTTP Access Logging

Linkerd provides HTTP access logging that captures detailed information about every HTTP request that passes through the proxy. This complements application-level structured logging with proxy-level request/response details.

**Benefits**:
- **Audit Trail**: Complete record of all HTTP traffic through the mesh
- **Security Analysis**: Detect suspicious patterns, unauthorized access attempts
- **Performance Debugging**: Identify slow requests, high latency endpoints
- **Compliance**: Meet regulatory requirements for request logging

**Configuration**:

Linkerd HTTP access logging is configured via the `HTTPAccessLogPolicy` resource:

```yaml
apiVersion: policy.linkerd.io/v1beta3
kind: HTTPAccessLogPolicy
metadata:
  name: operator-access-logs
  namespace: knative-lambda
spec:
  targetRef:
    group: policy.linkerd.io
    kind: Server
    name: operator-server
  conditions:
    # Log all requests
    - methods: ["*"]
      paths: ["*"]
    # Or log specific routes
    - methods: ["POST", "PUT", "DELETE"]
      paths: ["/api/v1/lambda/functions/*"]
```

**Log Format**:

Linkerd access logs are emitted in JSON format with the following fields:

```json
{
  "timestamp": "2025-01-20T10:00:00.123Z",
  "method": "POST",
  "path": "/api/v1/lambda/functions",
  "authority": "operator.knative-lambda.svc.cluster.local:8080",
  "status": 201,
  "latency_ms": 45,
  "request_size": 1024,
  "response_size": 512,
  "client": {
    "identity": "lambda-fn-1.knative-lambda.svc.cluster.local",
    "ip": "10.0.1.5"
  },
  "server": {
    "identity": "operator.knative-lambda.svc.cluster.local",
    "ip": "10.0.2.10"
  },
  "route": "/api/v1/lambda/functions",
  "classification": "success"
}
```

**Log Collection**:

Access logs are written to stdout/stderr of the Linkerd proxy container and are collected by **Grafana Alloy** and sent to **Loki** for storage and querying.

**Grafana Alloy Configuration**:

[Grafana Alloy](https://grafana.com/docs/alloy/latest/) is the OpenTelemetry Collector distribution with Prometheus pipelines. It collects logs from Kubernetes containers and forwards them to Loki.

```yaml
# Grafana Alloy configuration to collect Linkerd proxy logs
apiVersion: v1
kind: ConfigMap
metadata:
  name: alloy-config
  namespace: monitoring
data:
  config.alloy: |
    // Discover Kubernetes pods with Linkerd proxy containers
    discovery.kubernetes "pods" {
      role = "pod"
    }

    // Collect logs from Linkerd proxy containers
    loki.source.kubernetes "linkerd_proxy_logs" {
      targets = discovery.kubernetes.pods.targets
      forward_to = [loki.write.loki.receiver]
      
      // Filter only Linkerd proxy containers
      selector {
        match_labels = {
          "app.kubernetes.io/name" = "linkerd-proxy"
        }
      }
    }

    // Write logs to Loki
    loki.write "loki" {
      endpoint {
        url = "http://loki-distributor.loki.svc.cluster.local:3100"
      }
      external_labels = {
        job = "linkerd-access-logs"
        component = "linkerd-proxy"
      }
    }
```

**Alloy Deployment** (DaemonSet):

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: alloy
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: alloy
  template:
    metadata:
      labels:
        app: alloy
    spec:
      containers:
      - name: alloy
        image: grafana/alloy:latest
        args:
          - run
          - /etc/alloy/config.alloy
        volumeMounts:
        - name: config
          mountPath: /etc/alloy
        - name: varlog
          mountPath: /var/log
          readOnly: true
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: alloy-config
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
```

**Loki Integration**:

[Grafana Loki](https://grafana.com/oss/loki/) is a horizontally scalable, highly available, multi-tenant log aggregation system. It indexes only metadata (labels) rather than log content, making it cost-effective and easy to operate.

- **Loki Distributor**: `loki-distributor.loki.svc.cluster.local:3100`
- **Log Labels**: Automatically extracted from Kubernetes metadata (namespace, pod, container)
- **Storage**: Object storage backend (MinIO/S3/GCS) for durable, cost-effective storage

**LogQL Queries**:

[LogQL](https://grafana.com/docs/loki/latest/logql/) is Loki's powerful query language for exploring logs:
```logql
# All access logs for operator
{job="linkerd-access-logs", namespace="knative-lambda", pod=~"operator-.*"}

# Failed requests (4xx, 5xx)
{job="linkerd-access-logs"} | json | status >= 400

# Slow requests (>1s)
{job="linkerd-access-logs"} | json | latency_ms > 1000

# Requests by method
{job="linkerd-access-logs"} | json | method="POST"

# Requests from specific client
{job="linkerd-access-logs"} | json | client_identity="lambda-fn-1.knative-lambda.svc.cluster.local"
```

**Use Cases**:
- **Security Auditing**: Track all API access, detect unauthorized clients
- **Performance Analysis**: Identify slow endpoints, high latency routes
- **Debugging**: Correlate access logs with application logs using trace IDs
- **Compliance**: Maintain audit trail of all HTTP requests

**Reference**: [Linkerd HTTP Access Logging Documentation](https://linkerd.io/2-edge/features/access-logging/)

### Alertmanager Integration & AI Agent Alert Investigation

The AI Agent receives alerts from Alertmanager and actively investigates them by querying Prometheus, analyzing PrometheusRules, and providing intelligent insights and recommendations.

#### Alertmanager Webhook Configuration

**Alertmanager Route to AI Agent**:

```yaml
# Alertmanager configuration
apiVersion: v1
kind: ConfigMap
metadata:
  name: alertmanager-config
  namespace: monitoring
data:
  alertmanager.yml: |
    global:
      resolve_timeout: 5m
    
    route:
      group_by: ['alertname', 'cluster', 'service']
      group_wait: 10s
      group_interval: 10s
      repeat_interval: 12h
      receiver: 'ai-agent'
      routes:
      # Route Knative Lambda alerts to AI Agent
      - match:
          namespace: knative-lambda
        receiver: 'ai-agent'
        continue: true
      
      # Route critical alerts to AI Agent + PagerDuty
      - match:
          severity: critical
        receiver: 'ai-agent-critical'
        continue: true
    
    receivers:
    - name: 'ai-agent'
      webhook_configs:
      - url: 'http://knative-lambda-ai-agent.knative-lambda.svc.cluster.local:8080/api/alerts'
        http_config:
          bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
        send_resolved: true
    
    - name: 'ai-agent-critical'
      webhook_configs:
      - url: 'http://knative-lambda-ai-agent.knative-lambda.svc.cluster.local:8080/api/alerts'
        http_config:
          bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      - url: 'https://events.pagerduty.com/v2/enqueue'
        # PagerDuty integration
```

#### PrometheusRules for Knative Lambda

**Example PrometheusRules**:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: knative-lambda-alerts
  namespace: knative-lambda
  labels:
    prometheus: kube-prometheus
    role: alert-rules
spec:
  groups:
  - name: knative-lambda-operator
    interval: 30s
    rules:
    # High reconciliation failure rate
    - alert: HighReconciliationFailureRate
      expr: |
        rate(operator_reconcile_total{result="failure"}[5m]) /
        rate(operator_reconcile_total[5m]) > 0.1
      for: 5m
      labels:
        severity: warning
        component: operator
      annotations:
        summary: "High reconciliation failure rate"
        description: "{{ $value | humanizePercentage }} of reconciliations are failing"
        runbook_url: "https://runbooks.example.com/reconciliation-failures"
    
    # DLQ depth too high
    - alert: HighDLQDepth
      expr: operator_dlq_depth > 100
      for: 2m
      labels:
        severity: warning
        component: operator
      annotations:
        summary: "DLQ depth is high"
        description: "DLQ has {{ $value }} messages for function {{ $labels.function_name }}"
    
    # Build job failures
    - alert: BuildJobFailures
      expr: |
        rate(operator_build_jobs_total{result="failure"}[10m]) > 0.1
      for: 5m
      labels:
        severity: warning
        component: operator
      annotations:
        summary: "High build failure rate"
        description: "{{ $value | humanize }} builds failing per second"
    
    # Circuit breaker open
    - alert: CircuitBreakerOpen
      expr: |
        linkerd_response_total{deployment="knative-lambda-operator",classification="failure"} /
        linkerd_response_total{deployment="knative-lambda-operator"} > 0.5
      for: 1m
      labels:
        severity: critical
        component: operator
      annotations:
        summary: "Circuit breaker opened for operator"
        description: "More than 50% of requests are failing"
    
    # Lambda function high error rate
    - alert: LambdaFunctionHighErrorRate
      expr: |
        sum(rate(linkerd_response_total{deployment=~"lambda-.*",classification="failure"}[5m])) by (deployment) /
        sum(rate(linkerd_response_total{deployment=~"lambda-.*"}[5m])) by (deployment) > 0.1
      for: 5m
      labels:
        severity: warning
        component: lambda-function
      annotations:
        summary: "High error rate for Lambda function"
        description: "Function {{ $labels.deployment }} has {{ $value | humanizePercentage }} error rate"
    
    # Lambda function high latency
    - alert: LambdaFunctionHighLatency
      expr: |
        histogram_quantile(0.95,
          sum(rate(linkerd_response_latency_ms_bucket{deployment=~"lambda-.*"}[5m])) by (deployment, le)
        ) > 5000
      for: 5m
      labels:
        severity: warning
        component: lambda-function
      annotations:
        summary: "High latency for Lambda function"
        description: "P95 latency for {{ $labels.deployment }} is {{ $value }}ms"
    
    # AI Agent investigation queue depth
    - alert: AIAgentInvestigationQueueDepth
      expr: ai_agent_investigation_queue_depth > 50
      for: 2m
      labels:
        severity: warning
        component: ai-agent
      annotations:
        summary: "AI Agent investigation queue is backing up"
        description: "{{ $value }} alerts waiting for investigation"
```

#### AI Agent Alert Investigation Implementation

**Alert Reception & Investigation Flow**:

```python
from fastapi import FastAPI, HTTPException, Request
from pydantic import BaseModel, Field
from typing import List, Dict, Optional, Any
from datetime import datetime
import httpx
import json

# Alertmanager webhook payload model
class AlertManagerAlert(BaseModel):
    status: str  # "firing" or "resolved"
    labels: Dict[str, str]
    annotations: Dict[str, str]
    startsAt: datetime
    endsAt: Optional[datetime] = None
    generatorURL: Optional[str] = None

class AlertManagerWebhook(BaseModel):
    version: str = "4"
    groupKey: str
    status: str
    receiver: str
    groupLabels: Dict[str, str]
    commonLabels: Dict[str, str]
    commonAnnotations: Dict[str, str]
    externalURL: Optional[str] = None
    alerts: List[AlertManagerAlert]

# Prometheus query client
class PrometheusClient:
    def __init__(self, base_url: str = "http://prometheus.monitoring.svc.cluster.local:9090"):
        self.base_url = base_url
        self.client = httpx.AsyncClient(base_url=base_url)
    
    async def query(self, query: str, time: Optional[datetime] = None) -> Dict:
        """Execute Prometheus query"""
        params = {"query": query}
        if time:
            params["time"] = time.isoformat()
        
        response = await self.client.get("/api/v1/query", params=params)
        response.raise_for_status()
        return response.json()
    
    async def query_range(self, query: str, start: datetime, end: datetime, step: str = "15s") -> Dict:
        """Execute Prometheus range query"""
        params = {
            "query": query,
            "start": start.timestamp(),
            "end": end.timestamp(),
            "step": step
        }
        response = await self.client.get("/api/v1/query_range", params=params)
        response.raise_for_status()
        return response.json()
    
    async def get_prometheus_rule(self, rule_name: str) -> Optional[Dict]:
        """Get PrometheusRule by name"""
        # Query Prometheus rules API
        response = await self.client.get("/api/v1/rules")
        response.raise_for_status()
        rules = response.json()
        
        # Find matching rule
        for group in rules.get("data", {}).get("groups", []):
            for rule in group.get("rules", []):
                if rule.get("name") == rule_name:
                    return rule
        return None

# Alert investigation service
class AlertInvestigationService:
    def __init__(self, prometheus_client: PrometheusClient, lancedb_store: EventVectorStore):
        self.prometheus = prometheus_client
        self.lancedb = lancedb_store
    
    async def investigate_alert(self, alert: AlertManagerAlert) -> Dict[str, Any]:
        """Investigate an alert by querying Prometheus and analyzing context"""
        investigation = {
            "alert": alert.dict(),
            "investigation_start": datetime.utcnow().isoformat(),
            "findings": [],
            "recommendations": [],
            "related_events": [],
            "metrics_context": {}
        }
        
        # 1. Get PrometheusRule that triggered this alert
        rule_name = alert.labels.get("alertname")
        if rule_name:
            rule = await self.prometheus.get_prometheus_rule(rule_name)
            if rule:
                investigation["prometheus_rule"] = rule
                investigation["findings"].append({
                    "type": "rule_analysis",
                    "message": f"Alert triggered by rule: {rule_name}",
                    "rule_expr": rule.get("expr", ""),
                    "rule_for": rule.get("for", "")
                })
        
        # 2. Query Prometheus for context metrics
        if "function_name" in alert.labels:
            function_name = alert.labels["function_name"]
            
            # Query function-specific metrics
            metrics_queries = {
                "reconciliation_rate": f"rate(operator_reconcile_total{{function_name=\"{function_name}\"}}[5m])",
                "error_rate": f"rate(operator_reconcile_total{{function_name=\"{function_name}\",result=\"failure\"}}[5m])",
                "build_duration": f"histogram_quantile(0.95, operator_build_duration_seconds_bucket{{function_name=\"{function_name}\"}})",
                "dlq_depth": f"operator_dlq_depth{{function_name=\"{function_name}\"}}"
            }
            
            for metric_name, query in metrics_queries.items():
                try:
                    result = await self.prometheus.query(query)
                    investigation["metrics_context"][metric_name] = result
                except Exception as e:
                    investigation["findings"].append({
                        "type": "error",
                        "message": f"Failed to query {metric_name}: {str(e)}"
                    })
        
        # 3. Query Linkerd metrics for service health
        if "deployment" in alert.labels:
            deployment = alert.labels["deployment"]
            
            linkerd_queries = {
                "request_rate": f"sum(rate(linkerd_response_total{{deployment=\"{deployment}\"}}[5m]))",
                "error_rate": f"sum(rate(linkerd_response_total{{deployment=\"{deployment}\",classification=\"failure\"}}[5m]))",
                "p95_latency": f"histogram_quantile(0.95, sum(rate(linkerd_response_latency_ms_bucket{{deployment=\"{deployment}\"}}[5m])) by (le))",
                "circuit_breaker_state": f"linkerd_circuit_breaker_state{{deployment=\"{deployment}\"}}"
            }
            
            for metric_name, query in linkerd_queries.items():
                try:
                    result = await self.prometheus.query(query)
                    investigation["metrics_context"][f"linkerd_{metric_name}"] = result
                except Exception as e:
                    investigation["findings"].append({
                        "type": "error",
                        "message": f"Failed to query Linkerd {metric_name}: {str(e)}"
                    })
        
        # 4. Search for similar events in LanceDB
        if "function_name" in alert.labels:
            similar_events = self.lancedb.search_similar_events_by_function(
                function_name=alert.labels["function_name"],
                event_type="lambda.build.failed",  # Example
                limit=10
            )
            investigation["related_events"] = similar_events
        
        # 5. Generate recommendations based on findings
        investigation["recommendations"] = self._generate_recommendations(
            alert, investigation["metrics_context"], investigation["related_events"]
        )
        
        investigation["investigation_end"] = datetime.utcnow().isoformat()
        return investigation
    
    def _generate_recommendations(self, alert: AlertManagerAlert, metrics: Dict, events: List[Dict]) -> List[str]:
        """Generate intelligent recommendations based on investigation"""
        recommendations = []
        
        alert_name = alert.labels.get("alertname", "")
        
        if alert_name == "HighDLQDepth":
            recommendations.append("Check RabbitMQ consumer health and processing rate")
            recommendations.append("Review DLQ messages for error patterns")
            recommendations.append("Consider scaling up consumers if queue depth persists")
        
        elif alert_name == "BuildJobFailures":
            recommendations.append("Check container registry connectivity")
            recommendations.append("Review Kaniko job logs for build errors")
            recommendations.append("Verify source code availability in MinIO/S3")
            if metrics.get("build_duration"):
                recommendations.append("Consider increasing build timeout if builds are timing out")
        
        elif alert_name == "CircuitBreakerOpen":
            recommendations.append("Service is experiencing high failure rate - circuit breaker opened")
            recommendations.append("Check downstream service health (Notifi services)")
            recommendations.append("Review Linkerd metrics for root cause")
            recommendations.append("Consider scaling up service if load is high")
        
        elif alert_name == "LambdaFunctionHighErrorRate":
            function_name = alert.labels.get("deployment", "unknown")
            recommendations.append(f"Review function logs for {function_name}")
            recommendations.append("Check function resource limits (CPU/memory)")
            recommendations.append("Verify function code for runtime errors")
            if events:
                recommendations.append(f"Found {len(events)} similar past events - review patterns")
        
        return recommendations

# FastAPI endpoint for Alertmanager webhooks
@app.post("/api/alerts")
async def handle_alertmanager_webhook(webhook: AlertManagerWebhook):
    """Receive alerts from Alertmanager and trigger investigation"""
    investigation_service = AlertInvestigationService(
        prometheus_client=PrometheusClient(),
        lancedb_store=vector_store
    )
    
    investigations = []
    
    for alert in webhook.alerts:
        if alert.status == "firing":  # Only investigate firing alerts
            investigation = await investigation_service.investigate_alert(alert)
            investigations.append(investigation)
            
            # Store investigation in LanceDB for pattern analysis
            vector_store.store_investigation(investigation)
            
            # Log investigation
            logger.info(f"Investigated alert: {alert.labels.get('alertname')}", 
                       extra={"investigation": investigation})
    
    return {
        "received": len(webhook.alerts),
        "investigated": len(investigations),
        "investigations": investigations
    }

# Endpoint to manually trigger investigation
@app.post("/api/alerts/investigate")
async def investigate_alert(alert_name: str, labels: Dict[str, str]):
    """Manually trigger alert investigation"""
    alert = AlertManagerAlert(
        status="firing",
        labels=labels,
        annotations={},
        startsAt=datetime.utcnow()
    )
    
    investigation_service = AlertInvestigationService(
        prometheus_client=PrometheusClient(),
        lancedb_store=vector_store
    )
    
    investigation = await investigation_service.investigate_alert(alert)
    return investigation
```

**Investigation Storage in LanceDB**:

```python
# Add investigation storage to EventVectorStore
def store_investigation(self, investigation: Dict):
    """Store alert investigation in LanceDB"""
    investigations_table = self.db.open_table("investigations")
    
    investigation_record = {
        "id": generate_uuid(),
        "alert_name": investigation["alert"]["labels"].get("alertname"),
        "function_name": investigation["alert"]["labels"].get("function_name"),
        "status": investigation["alert"]["status"],
        "investigation_data": json.dumps(investigation),
        "timestamp": datetime.utcnow(),
        "recommendations": json.dumps(investigation["recommendations"]),
        "findings_count": len(investigation["findings"])
    }
    
    investigations_table.add([investigation_record])
```

**PrometheusRule Investigation**:

```python
async def investigate_prometheus_rule(self, rule_name: str) -> Dict:
    """Investigate a PrometheusRule by analyzing its expression and current state"""
    # Get rule definition
    rule = await self.prometheus.get_prometheus_rule(rule_name)
    if not rule:
        return {"error": f"Rule {rule_name} not found"}
    
    investigation = {
        "rule_name": rule_name,
        "rule_expression": rule.get("expr", ""),
        "rule_for": rule.get("for", ""),
        "current_state": "unknown",
        "evaluation_results": {}
    }
    
    # Evaluate rule expression
    expr = rule.get("expr", "")
    if expr:
        try:
            # Query current value
            result = await self.prometheus.query(expr)
            investigation["evaluation_results"]["current"] = result
            
            # Query historical values (last hour)
            end_time = datetime.utcnow()
            start_time = end_time - timedelta(hours=1)
            range_result = await self.prometheus.query_range(expr, start_time, end_time)
            investigation["evaluation_results"]["historical"] = range_result
            
            # Determine if rule would fire
            if result.get("data", {}).get("result"):
                values = [float(r.get("value", [None, 0])[1]) for r in result["data"]["result"]]
                if values and any(v > 0 for v in values):
                    investigation["current_state"] = "firing"
                else:
                    investigation["current_state"] = "normal"
        except Exception as e:
            investigation["error"] = str(e)
    
    return investigation

# Endpoint to investigate PrometheusRule
@app.post("/api/rules/investigate")
async def investigate_prometheus_rule(rule_name: str):
    """Investigate a PrometheusRule by analyzing its expression and current state"""
    with investigation_duration.time():
        investigation = await alert_investigation_service.investigate_prometheus_rule(rule_name)
        prometheus_queries.labels(query_type="rule_investigation").inc()
        return investigation
```

### Observability Integration Summary

**Metrics Exposed**:
- **Operator**: `:8080/metrics` - All operator metrics
- **AI Agent**: `:8080/metrics` - All AI Agent metrics
- **Lambda Functions**: Via Linkerd automatic metrics
- **Linkerd**: Automatic metrics for all meshed services

**Prometheus Scraping Configuration**:

```yaml
# Prometheus ServiceMonitor for Operator
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: knative-lambda-operator
  namespace: knative-lambda
spec:
  selector:
    matchLabels:
      app: knative-lambda-operator
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics

# Prometheus ServiceMonitor for AI Agent
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: knative-lambda-ai-agent
  namespace: knative-lambda
spec:
  selector:
    matchLabels:
      app: knative-lambda-ai-agent
  endpoints:
  - port: http
    interval: 30s
    path: /metrics
```

**Grafana Dashboards**:

1. **Operator Dashboard**: Reconciliation rates, build metrics, DLQ depth, API metrics
2. **Lambda Functions Dashboard**: Invocation rates, error rates, latency (P50/P95/P99), circuit breaker states
3. **AI Agent Dashboard**: Events processed, investigations, pattern detection, embedding generation time
4. **Linkerd Dashboard**: Service topology, request rates, latency, circuit breaker states, retry budgets
5. **Alert Investigation Dashboard**: Alert investigation queue, investigation duration, recommendations generated

**Log Aggregation**: 
- **Application Logs**: Structured JSON logs with correlation IDs and trace context forwarded to Loki
- **Linkerd HTTP Access Logs**: Detailed HTTP request/response logs via `HTTPAccessLogPolicy`, collected by Grafana Alloy and forwarded to Loki
- **LogQL Queries**: Query both application and access logs for security auditing, performance analysis, and debugging

**Distributed Tracing** (Fully Integrated): 
- **Operator (Go)**: Emits traces to Tempo via OpenTelemetry
- **AI Agent (Python)**: Emits traces to Tempo (primary) via OpenTelemetry, with Logfire as optional fallback
- **Linkerd Proxies**: Automatically emit spans when W3C Trace Context headers are present
- **Notifi Services**: All services (Scheduler, Storage Manager, Subscription Manager, Fetch Proxy, Blockchain Manager) export traces to Tempo
- **Knative Services**: Lambda functions propagate W3C Trace Context headers for end-to-end correlation
- **Unified Backend**: Tempo as single source of truth for all distributed traces
- **Correlation**: All traces use W3C Trace Context headers for complete end-to-end request correlation across all components

### Grafana Dashboards

**Recommended Dashboards**:
1. **Operator Overview**: Reconciliation rate, success rate, latency
2. **Build Pipeline**: Build jobs, duration, success rate
3. **API Performance**: Request rate, latency, error rate
4. **DLQ Monitoring**: Queue depth, message rate, replay success
5. **System Health**: CPU, memory, goroutines

## 🚀 Performance & Scalability

### Performance Characteristics

**Target Performance**:
- Reconciliation latency (p95): < 5 seconds
- Reconciliation latency (p99): < 10 seconds
- API response time (p95): < 500ms
- Event processing rate: > 100 events/second
- Throughput: > 50 reconciliations/second per operator instance

**Bottlenecks**:
1. **Kubernetes API Rate Limiting**: 400 requests/second default
   - Mitigation: Batch operations, client-side rate limiting
2. **RabbitMQ Broker Throughput**: ~1000 messages/second
   - Mitigation: Multiple brokers, sharding
3. **Kaniko Job Creation Rate**: Limited by cluster resources
   - Mitigation: Queue management, resource quotas
4. **Container Registry Push Rate Limits**: Registry throttling (ECR/GHCR/GCR/local)
   - Mitigation: Exponential backoff, retry logic
   - Note: Registry is configurable (ECR, GHCR, GCR, or local Docker registry)

### Scalability Limits

**Per Operator Instance**:
- Maximum LambdaFunctions: 1000 (recommended: 500)
- Maximum concurrent reconciliations: 50
- Resource requirements per function: ~1m CPU, ~5Mi memory

**Horizontal Scaling**:
- Operator can be scaled horizontally (multiple replicas)
- Leader election ensures only one active reconciler per namespace
- Event distribution via RabbitMQ ensures load balancing

**Resource Sizing Guidelines**:
```yaml
resources:
  requests:
    cpu: "500m"
    memory: "512Mi"
  limits:
    cpu: "2000m"
    memory: "2Gi"
```

### Optimization Strategies

**Batch Reconciliation**:
- Group multiple CRD updates into single reconciliation cycle
- Reduce Kubernetes API calls by 60-80%

**Event Batching**:
- Batch CloudEvents processing (10-50 events per batch)
- Reduce RabbitMQ connection overhead

**Caching**:
- Cache CRD state (TTL: 30 seconds)
- Cache Knative Service status
- Reduce API server load

**Connection Pooling**:
- Reuse Kubernetes client connections
- Reuse RabbitMQ connections
- Connection pool size: 10-20 connections

## 💰 Cost Optimization

### Cost Model

**Cost per LambdaFunction** (monthly):
- Build cost: $0.05 per build (Kaniko job)
- Runtime cost: $0.10 per function (Knative Service, scale-to-zero)
- Storage cost: Variable based on registry:
  - ECR: $0.01 per function (AWS pricing)
  - GHCR: Free for public, $0.10/GB for private
  - GCR: $0.026/GB per month
  - Local registry: No storage cost (uses cluster storage)
- **Total**: ~$0.16 per function/month (idle, using ECR as example)

**Operator Infrastructure Costs**:
- Operator pod: $15/month (1 CPU, 1Gi memory, always-on)
- RabbitMQ broker: $20/month (shared)
- Monitoring: $5/month (Prometheus storage)
- **Total**: ~$40/month

### Cost Optimization Strategies

**1. Spot Instances for Build Jobs**:
- Use Spot instances for Kaniko jobs (60% savings)
- Fallback to on-demand if spot unavailable
- Estimated savings: $0.03 per build

**2. Container Registry Lifecycle Policies**:
- Delete images older than 90 days
- Keep only last 10 images per function
- Applies to: ECR, GHCR, GCR (local registry uses cluster storage quotas)
- Estimated savings: 50% storage cost (for cloud registries)

**3. Storage Lifecycle Policies**:
- **MinIO** (default): No lifecycle policies needed (uses cluster storage)
- **AWS S3**: Transition to Intelligent-Tiering after 30 days, Archive to Glacier after 90 days
- **GCS**: Transition to Nearline after 30 days, Coldline after 90 days
- Estimated savings: 40% storage cost (for cloud storage)

**4. Resource Rightsizing**:
- Use Vertical Pod Autoscaler (VPA) recommendations
- Right-size operator resources based on actual usage
- Estimated savings: 20-30% compute cost

**5. Scale-to-Zero Benefits**:
- Knative Services scale to zero when idle
- No cost for idle functions
- Significant savings for low-traffic functions

### Cost Monitoring

**Cost Allocation Tags**:
```yaml
tags:
  Environment: production
  Project: knative-lambda
  Team: platform
  CostCenter: engineering
  ManagedBy: operator
```

**Cost Tracking Metrics**:
- `operator_cost_per_function_total` - Cumulative cost per function
- `operator_build_cost_total` - Total build costs
- `operator_storage_cost_total` - Total storage costs

**Cost Anomaly Detection**:
- Alert when daily cost exceeds baseline by 25%
- Alert when per-function cost exceeds threshold
- Budget alerts at 80%, 90%, 100% of monthly budget

## 🧪 Testing Strategy

### Testing Pyramid

```
        ┌─────────────┐
        │  E2E Tests  │  ← 10% (Critical paths)
        ├─────────────┤
        │ Integration │  ← 30% (K8s API, RabbitMQ)
        │    Tests    │
        ├─────────────┤
        │ Unit Tests  │  ← 60% (Business logic)
        └─────────────┘
```

### Unit Testing

**Coverage Goal**: 80%

**What to Test**:
- Event parsing and validation
- Reconciliation logic
- Status updates
- Error handling
- Metrics emission

**Testing Framework**: Go `testing` package with `testify`

**Example**:
```go
func TestReconcileLambdaFunction(t *testing.T) {
    tests := []struct {
        name    string
        crd     *LambdaFunction
        wantErr bool
    }{
        {
            name: "valid function",
            crd: &LambdaFunction{
                Spec: LambdaFunctionSpec{
                    Source: SourceSpec{Type: "minio"},  // Default: MinIO
                    Runtime: RuntimeSpec{Language: "nodejs"},
                },
            },
            wantErr: false,
        },
    }
    // ... test implementation
}
```

### Integration Testing

**Coverage Goal**: 70%

**Test Environment**:
- Kind (Kubernetes in Docker) cluster
- Local RabbitMQ instance
- Mock AWS services (LocalStack)

**What to Test**:
- CRD creation/update/delete
- Build job creation
- Knative Service creation
- Event processing
- DLQ handling

**Test Utilities**:
- `testenv` package for test environment setup
- `testfixtures` for test data
- `testhelpers` for common operations

### E2E Testing

**Coverage**: Critical paths only

**Test Scenarios**:
1. Create function → Build → Deploy → Ready
2. Update function → Rebuild → Redeploy
3. Delete function → Cleanup
4. Build failure → DLQ → Retry
5. Service failure → Health check → Recovery

**Chaos Engineering**:
- Random pod failures
- Network partitions
- API server throttling
- Resource exhaustion

**Load Testing**:
- 100 concurrent function creations
- 1000 events/second processing
- Stress test reconciliation loop

## 🔐 Security Architecture

### RBAC Configuration

**Service Account**:
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: knative-lambda-operator
  namespace: knative-lambda
```

**ClusterRole** (Least Privilege):
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: knative-lambda-operator
rules:
# LambdaFunction CRD
- apiGroups: ["lambda.knative.io"]
  resources: ["lambdafunctions"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
# Knative Services
- apiGroups: ["serving.knative.dev"]
  resources: ["services", "revisions", "routes"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
# Build Jobs
- apiGroups: ["batch"]
  resources: ["jobs"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
# Events
- apiGroups: ["eventing.knative.dev"]
  resources: ["brokers", "triggers"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
# RabbitMQ
- apiGroups: ["rabbitmq.com"]
  resources: ["exchanges", "queues", "bindings"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

### Network Security

**Network Policies**:
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: operator-network-policy
spec:
  podSelector:
    matchLabels:
      app: knative-lambda-operator
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: knative-lambda
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: knative-lambda
    ports:
    - protocol: TCP
      port: 8080
  - to:
    - namespaceSelector:
        matchLabels:
          name: rabbitmq-cluster
    ports:
    - protocol: TCP
      port: 5672
```

**Service Mesh Integration** (Linkerd):
- Automatic mTLS for all pod-to-pod communication
- Traffic policies for rate limiting via Policy API (Server + ServerAuthorization)
- Circuit breakers via Linkerd retries and timeouts
- **Observability First-Class**: Automatic metrics, traces, service topology, and real-time request inspection
- See [Service Mesh Integration (Linkerd)](#service-mesh-integration-linkerd) for detailed configuration
- See [Observability & Monitoring](#observability--monitoring-first-class-citizen) for comprehensive observability strategy

### Secret Management

**Secret Storage**: Vault or Sealed Secrets

**Secret Rotation**:
- AWS credentials rotated every 90 days
- Image pull secrets rotated on security events
- Automatic secret refresh in operator

**Secret Access**:
```yaml
env:
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: aws-credentials
      key: access-key-id
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: aws-credentials
      key: secret-access-key
```

### Container Security

**Image Scanning**:
- Trivy scanning in CI/CD pipeline
- Vulnerability scanning on image pull
- Block deployment if critical vulnerabilities found

**Pod Security Standards**:
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000
  seccompProfile:
    type: RuntimeDefault
  capabilities:
    drop:
    - ALL
```

### API Security

**Authentication**: Kubernetes Service Account tokens

**Authorization**: RBAC-based access control

**Rate Limiting**: 
- 100 requests/second per client
- Exponential backoff on rate limit

**Input Validation**:
- Schema validation for all API requests
- Sanitization of user inputs
- Reject malformed requests

## 📈 Capacity Planning

### Resource Forecasting

**Growth Projections**:
```
Month    | Functions | Builds/Day | Peak Concurrent
---------|-----------|------------|-----------------
Jan 2025 | 100       | 500        | 10
Mar 2025 | 200       | 1,000      | 20
Jun 2025 | 500       | 2,500      | 50
Dec 2025 | 1,000     | 5,000      | 100
```

**Resource Requirements Calculation**:
```python
# Per function resource requirements
cpu_per_function = 0.001  # 1m CPU
memory_per_function = 0.005  # 5Mi memory

# Operator base resources
operator_base_cpu = 0.5  # 500m CPU
operator_base_memory = 0.5  # 512Mi memory

# Total resources needed
total_cpu = operator_base_cpu + (num_functions * cpu_per_function)
total_memory = operator_base_memory + (num_functions * memory_per_function)
```

**Headroom Recommendations**:
- 30% headroom for unexpected spikes
- 20% headroom for growth
- Total: 50% headroom recommended

### Scaling Triggers

**Horizontal Pod Autoscaler (HPA)**:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: knative-lambda-operator
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: knative-lambda-operator
  minReplicas: 1
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

**Vertical Pod Autoscaler (VPA)**:
- Recommendations only (updateMode: "Off")
- Provides resource sizing recommendations
- Review weekly and apply manually

### Load Testing

**Test Scenarios**:
1. **Normal Load**: 100 functions, 500 builds/day
2. **Peak Load**: 200 functions, 1,000 builds/day
3. **Stress Test**: 500 functions, 2,500 builds/day

**Performance Baselines**:
- Reconciliation latency: < 5s (p95)
- API response time: < 500ms (p95)
- Event processing: > 100 events/s
- Build job creation: > 10 jobs/s

## 🔄 Disaster Recovery & High Availability

### High Availability

**Operator Replication**:
- Multiple operator replicas (3 recommended)
- Leader election ensures only one active reconciler
- Automatic failover on leader failure

**Leader Election**:
```yaml
leaderElection:
  enabled: true
  resourceName: knative-lambda-operator-leader
  resourceNamespace: knative-lambda
  leaseDuration: 15s
  renewDeadline: 10s
  retryPeriod: 2s
```

**Multi-Zone Deployment**:
- Operator pods distributed across availability zones
- RabbitMQ cluster across zones
- Kubernetes API server multi-zone

### Disaster Recovery

**Backup Strategies**:
1. **CRD State**: Backed up via etcd snapshots
2. **RabbitMQ State**: Persistent volumes with snapshots
3. **Container Images**: Replicated to secondary region (if using cloud registries with replication support: ECR, GCR)
4. **Source Code Storage**:
   - **MinIO** (default): Backed up via persistent volume snapshots
   - **AWS S3**: Cross-region replication (optional)
   - **GCS**: Cross-region replication (optional)

**Recovery Procedures**:
1. Restore etcd from snapshot
2. Restore RabbitMQ from volume snapshot
3. Verify operator connectivity
4. Trigger reconciliation for all functions
5. Validate service health

**RTO/RPO Targets**:
- RTO (Recovery Time Objective): < 1 hour
- RPO (Recovery Point Objective): < 15 minutes

### Failure Scenarios

**Operator Failure**:
- Automatic restart via Kubernetes
- Leader election ensures continuity
- No data loss (state in CRDs)

**RabbitMQ Failure**:
- Events buffered in DLQ
- Automatic reconnection
- Event replay on recovery

**Kubernetes API Failure**:
- Exponential backoff retry
- Queue operations for later
- Graceful degradation

## 🤖 AI Agent ML Integration Details

### ML/AI Framework Integration

**Model Serving Architecture**:
```
┌─────────────────────────────────────────────────────────────┐
│              AI Agent ML Pipeline                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  CloudEvent Input                                           │
│         │                                                   │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ Feature Engineering  │                                   │
│  │ - Event type         │                                   │
│  │ - Timestamp          │                                   │
│  │ - Function metadata  │                                   │
│  │ - Historical patterns│                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         ▼                                                   │
│  ┌──────────────────────┐                                   │
│  │ Model Inference      │                                   │
│  │ - Anomaly detection  │                                   │
│  │ - Prediction models  │                                   │
│  │ - Recommendation     │                                   │
│  └──────┬───────────────┘                                   │
│         │                                                   │
│         ▼                                                   │
│  ┌──────────────────────────┐                               │
│  │ Action Generation        │                               │
│  │ - Scaling recommendations│                               │
│  │ - Cost optimizations     │                               │
│  │ - Alert triggers         │                               │
│  └──────────────────────────┘                               │
└─────────────────────────────────────────────────────────────┘
```

**Model Versioning**:
- Models stored in model registry
- A/B testing for model updates
- Gradual rollout (10% → 50% → 100%)
- Automatic rollback on performance degradation

**A/B Testing**:
- Split traffic between model versions
- Compare metrics (accuracy, latency)
- Select best performing model

### Data Pipeline

**Event Data Preprocessing**:
- Normalize event schemas
- Extract features (event type, timestamp, metadata)
- Enrich with historical context
- Handle missing values

**Feature Engineering**:
- Time-based features (hour, day, week)
- Aggregated features (event rate, success rate)
- Function-specific features (runtime, memory)
- Cross-function features (correlation)

**Data Storage**:
- **LanceDB**: Vector database for event embeddings and similarity search
- Time-series database (InfluxDB) for metrics
- **Object storage**: MinIO (default, local), AWS S3 (optional), or Google Cloud Storage (optional) for raw events and training datasets

**Training Data Collection**:
- Collect labeled events (success/failure)
- Store in LanceDB with metadata for fine-tuning
- Export to training datasets (JSONL format)
- Periodic retraining (weekly/monthly)

### Model Operations

**Model Deployment**:
- Containerized models (ONNX, TensorFlow Serving)
- Knative Service for model serving
- Auto-scaling based on inference load
- Blue-green deployments

**Model Monitoring**:
- Prediction accuracy tracking
- Latency monitoring
- Drift detection (data/model)
- Performance degradation alerts

**Drift Detection**:
- Statistical tests (KS test, PSI)
- Alert when drift exceeds threshold
- Automatic retraining trigger

**Retraining Triggers**:
- Scheduled (weekly/monthly)
- On drift detection
- On performance degradation
- On new data availability

### Performance Metrics

**Inference Latency**:
- Target: < 100ms (p95)
- Measurement: End-to-end inference time
- Alert: > 200ms (p95)

**Model Accuracy**:
- Classification accuracy: > 90%
- Regression R²: > 0.85
- Alert: Accuracy drops > 5%

**Prediction Confidence**:
- Track confidence scores
- Low confidence alerts
- Human review for low confidence predictions

**Action Success Rate**:
- Track action outcomes
- Measure success rate
- Alert: Success rate < 80%

## 🤖 AI Agent Architecture: Pydantic & LanceDB Integration

### Overview

The AI Agent is implemented as a **Python-based Knative Service** that processes all 17 event types using:
- **Pydantic**: For event validation and data modeling
- **LanceDB**: For vector storage and similarity search of event patterns
- **Event Collection Pipeline**: For gathering training data for fine-tuning

### AI Agent Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│              AI Agent Service Architecture                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. EVENT INGESTION                                         │
│     ┌──────────────────┐                                    │
│     │ CloudEvent       │                                    │
│     │ (17 event types) │                                    │
│     └────────┬─────────┘                                    │
│              │                                              │
│              │ POST /api/events                             │
│              ▼                                              │
│     ┌──────────────────┐                                    │
│     │ Event Handler    │                                    │
│     │ (FastAPI)        │                                    │
│     └────────┬─────────┘                                    │
│              │                                              │
│  2. EVENT VALIDATION (Pydantic)                             │
│              │                                              │
│              ▼                                              │
│     ┌──────────────────┐                                    │
│     │ Pydantic Models  │                                    │
│     │ ──────────────── │                                    │
│     │ - BaseEvent      │                                    │
│     │ - FunctionEvent  │                                    │
│     │ - BuildEvent     │                                    │
│     │ - ServiceEvent   │                                    │
│     │ - StatusEvent    │                                    │
│     │ - ParserEvent    │                                    │
│     └────────┬─────────┘                                    │
│              │                                              │
│              │ Validated Event                              │
│              ▼                                              │
│     ┌──────────────────────┐                                │
│     │ Event Processor      │                                │
│     │ ────────────────     │                                │
│     │ - Extract features   │                                │
│     │ - Generate embeddings│                                │
│     │ - Analyze patterns   │                                │
│     └────────┬─────────────┘                                │
│              │                                              │
│  3. VECTOR STORAGE (LanceDB)                                │
│              │                                              │
│              ├──────────────┬──────────────┐                │
│              │              │              │                │
│              ▼              ▼              ▼                │
│     ┌─────────────┐ ┌──────────────┐ ┌─────────────┐        │
│     │ LanceDB     │ │ LanceDB      │ │ LanceDB     │        │
│     │ Table:      │ │ Table:       │ │ Table:      │        │
│     │ events      │ │ embeddings   │ │ patterns    │        │
│     │ ─────────── │ │ ───────────  │ │ ─────────── │        │
│     │ - Event data│ │ - Vector     │ │ - Patterns  │        │
│     │ - Metadata  │ │   embeddings │ │ - Anomalies │        │
│     │ - Timestamp │ │ - Similarity │ │ - Insights  │        │
│     └─────────────┘ └──────────────┘ └─────────────┘        │
│              │                                              │
│  4. ANALYSIS & INSIGHTS                                     │
│              │                                              │
│              ▼                                              │
│     ┌──────────────────┐                                    │
│     │ AI Analysis      │                                    │
│     │ ──────────────── │                                    │
│     │ - Anomaly detect │                                    │
│     │ - Pattern match  │                                    │
│     │ - Predictions    │                                    │
│     │ - Recommendations│                                    │
│     └────────┬─────────┘                                    │
│              │                                              │
│  5. TRAINING DATA COLLECTION                                │
│              │                                              │
│              ▼                                              │
│     ┌──────────────────┐                                    │
│     │ Training Data    │                                    │
│     │ Pipeline         │                                    │
│     │ ──────────────── │                                    │
│     │ - Label events   │                                    │
│     │ - Export JSONL   │                                    │
│     │ - Store in
|     |    MinIO/S3/GCS  │                                     
│     │ - Prepare for    │                                    │
│     │   fine-tuning    │                                    │
│     └──────────────────┘                                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Pydantic Event Models

The AI Agent uses Pydantic models for type-safe event validation and serialization:

```python
from pydantic import BaseModel, Field, validator, ValidationError
from typing import Optional, Dict, Any, Literal, Union
from datetime import datetime
from enum import Enum

class EventType(str, Enum):
    """All 17 supported event types"""
    # Function Management
    FUNCTION_CREATED = "lambda.function.created"
    FUNCTION_UPDATED = "lambda.function.updated"
    FUNCTION_DELETED = "lambda.function.deleted"
    
    # Build Events
    BUILD_STARTED = "lambda.build.started"
    BUILD_COMPLETED = "lambda.build.completed"
    BUILD_FAILED = "lambda.build.failed"
    BUILD_TIMEOUT = "lambda.build.timeout"
    BUILD_CANCELLED = "lambda.build.cancelled"
    BUILD_STOPPED = "lambda.build.stopped"
    
    # Service Events
    SERVICE_CREATED = "lambda.service.created"
    SERVICE_UPDATED = "lambda.service.updated"
    SERVICE_DELETED = "lambda.service.deleted"
    
    # Status Events
    STATUS_UPDATED = "lambda.status.updated"
    HEALTH_CHECK = "lambda.health.check"
    
    # Parser Events
    PARSER_STARTED = "lambda.parser.started"
    PARSER_COMPLETED = "lambda.parser.completed"
    PARSER_FAILED = "lambda.parser.failed"

class BaseEvent(BaseModel):
    """Base CloudEvent model with Pydantic validation"""
    specversion: str = Field(default="1.0", alias="specversion")
    type: EventType
    source: str
    id: str
    time: datetime
    datacontenttype: Optional[str] = Field(default="application/json", alias="datacontenttype")
    data: Dict[str, Any]
    
    class Config:
        allow_population_by_field_name = True
        use_enum_values = True

class FunctionEventData(BaseModel):
    """Data payload for function management events"""
    name: str
    namespace: str = "knative-lambda"
    spec: Dict[str, Any]
    status: Optional[Dict[str, Any]] = None

class BuildEventData(BaseModel):
    """Data payload for build events"""
    function_name: str
    build_id: str
    job_name: Optional[str] = None
    image_uri: Optional[str] = None
    status: Literal["started", "completed", "failed", "timeout", "cancelled", "stopped"]
    error: Optional[str] = None
    duration_seconds: Optional[float] = None

class ServiceEventData(BaseModel):
    """Data payload for service events"""
    service_name: str
    namespace: str = "knative-lambda"
    url: Optional[str] = None
    status: Optional[Dict[str, Any]] = None

class StatusEventData(BaseModel):
    """Data payload for status events"""
    function_name: str
    phase: Literal["Pending", "Building", "Deploying", "Ready", "Failed", "Deleting"]
    conditions: Optional[list] = None
    health_status: Optional[Literal["healthy", "unhealthy", "unknown"]] = None

class ParserEventData(BaseModel):
    """Data payload for parser events"""
    parser_id: str
    function_name: str
    execution_id: str
    status: Literal["started", "completed", "failed"]
    error: Optional[str] = None
    duration_ms: Optional[float] = None

class LambdaEvent(BaseEvent):
    """Typed Lambda event with validated data"""
    data: Union[FunctionEventData, BuildEventData, ServiceEventData, 
                StatusEventData, ParserEventData]
    
    @validator('data', pre=True)
    def validate_data_by_type(cls, v, values):
        """Validate data based on event type"""
        event_type = values.get('type')
        
        if event_type in [EventType.FUNCTION_CREATED, EventType.FUNCTION_UPDATED, 
                          EventType.FUNCTION_DELETED]:
            return FunctionEventData(**v)
        elif event_type in [EventType.BUILD_STARTED, EventType.BUILD_COMPLETED,
                            EventType.BUILD_FAILED, EventType.BUILD_TIMEOUT,
                            EventType.BUILD_CANCELLED, EventType.BUILD_STOPPED]:
            return BuildEventData(**v)
        elif event_type in [EventType.SERVICE_CREATED, EventType.SERVICE_UPDATED,
                           EventType.SERVICE_DELETED]:
            return ServiceEventData(**v)
        elif event_type in [EventType.STATUS_UPDATED, EventType.HEALTH_CHECK]:
            return StatusEventData(**v)
        elif event_type in [EventType.PARSER_STARTED, EventType.PARSER_COMPLETED,
                           EventType.PARSER_FAILED]:
            return ParserEventData(**v)
        return v
```

### LanceDB Integration

LanceDB is used for storing event embeddings and enabling similarity search:

```python
import lancedb
import pyarrow as pa
from typing import List, Dict, Any
import numpy as np

class EventVectorStore:
    """LanceDB integration for event vector storage"""
    
    def __init__(self, db_path: str = "/data/lancedb"):
        self.db = lancedb.connect(db_path)
        self._initialize_tables()
    
    def _initialize_tables(self):
        """Initialize LanceDB tables for events, embeddings, and patterns"""
        
        # Events table schema
        events_schema = pa.schema([
            pa.field("id", pa.string()),
            pa.field("event_id", pa.string()),
            pa.field("event_type", pa.string()),
            pa.field("timestamp", pa.timestamp("ns")),
            pa.field("function_name", pa.string()),
            pa.field("data", pa.string()),  # JSON string
            pa.field("metadata", pa.string()),  # JSON string
        ])
        
        # Embeddings table schema
        embeddings_schema = pa.schema([
            pa.field("id", pa.string()),
            pa.field("event_id", pa.string()),
            pa.field("embedding", pa.list_(pa.float32())),  # Vector embedding
            pa.field("event_type", pa.string()),
            pa.field("timestamp", pa.timestamp("ns")),
        ])
        
        # Patterns table schema
        patterns_schema = pa.schema([
            pa.field("id", pa.string()),
            pa.field("pattern_type", pa.string()),  # anomaly, success, failure
            pa.field("event_sequence", pa.list_(pa.string())),  # List of event IDs
            pa.field("embedding", pa.list_(pa.float32())),
            pa.field("metadata", pa.string()),
        ])
        
        # Create tables if they don't exist
        if "events" not in self.db.table_names():
            self.db.create_table("events", schema=events_schema)
        if "embeddings" not in self.db.table_names():
            self.db.create_table("embeddings", schema=embeddings_schema)
        if "patterns" not in self.db.table_names():
            self.db.create_table("patterns", schema=patterns_schema)
    
    def store_event(self, event: LambdaEvent, embedding: List[float]):
        """Store event and its embedding in LanceDB"""
        events_table = self.db.open_table("events")
        embeddings_table = self.db.open_table("embeddings")
        
        # Store event data
        event_record = {
            "id": event.id,
            "event_id": event.id,
            "event_type": event.type.value,
            "timestamp": event.time,
            "function_name": event.data.function_name if hasattr(event.data, 'function_name') else None,
            "data": event.data.json(),
            "metadata": event.dict(exclude={"data"}).json(),
        }
        events_table.add([event_record])
        
        # Store embedding
        embedding_record = {
            "id": event.id,
            "event_id": event.id,
            "embedding": embedding,
            "event_type": event.type.value,
            "timestamp": event.time,
        }
        embeddings_table.add([embedding_record])
    
    def search_similar_events(self, embedding: List[float], limit: int = 10) -> List[Dict]:
        """Search for similar events using vector similarity"""
        embeddings_table = self.db.open_table("embeddings")
        
        # Vector similarity search
        results = embeddings_table.search(embedding).limit(limit).to_pandas()
        return results.to_dict("records")
    
    def find_patterns(self, event_sequence: List[str]) -> List[Dict]:
        """Find similar patterns in event sequences"""
        patterns_table = self.db.open_table("patterns")
        
        # Search for similar patterns
        # This would use sequence embeddings or pattern matching
        # Implementation depends on pattern representation
        pass
```

### LanceDB Table Schemas & Examples

This section provides detailed schemas, data examples, and usage patterns for each LanceDB table used by the AI Agent.

#### 1. Events Table

**Purpose**: Stores raw event data with metadata for historical analysis and pattern detection.

**Schema**:
```python
events_schema = pa.schema([
    pa.field("id", pa.string()),                    # Unique event ID (UUID)
    pa.field("event_id", pa.string()),             # CloudEvent ID (same as id)
    pa.field("event_type", pa.string()),           # Event type (e.g., "lambda.function.created")
    pa.field("timestamp", pa.timestamp("ns")),      # Event timestamp (nanosecond precision)
    pa.field("function_name", pa.string()),        # Lambda function name (nullable)
    pa.field("data", pa.string()),                 # JSON string of event data
    pa.field("metadata", pa.string()),             # JSON string of CloudEvent metadata
])
```

**Example Records**:

```json
// Record 1: Function Created Event
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_type": "lambda.function.created",
  "timestamp": "2025-01-20T10:00:00.123456789Z",
  "function_name": "my-lambda-function",
  "data": "{\"function_name\":\"my-lambda-function\",\"runtime\":{\"language\":\"nodejs\",\"version\":\"22\"},\"source\":{\"type\":\"minio\",\"bucket\":\"knative-lambda-functions\",\"key\":\"functions/my-function.zip\"}}",
  "metadata": "{\"source\":\"notifi-scheduler\",\"specversion\":\"1.0\",\"type\":\"lambda.function.created\"}"
}

// Record 2: Build Started Event
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "event_id": "660e8400-e29b-41d4-a716-446655440001",
  "event_type": "lambda.build.started",
  "timestamp": "2025-01-20T10:00:05.234567890Z",
  "function_name": "my-lambda-function",
  "data": "{\"function_name\":\"my-lambda-function\",\"build_id\":\"build-123\",\"image\":\"registry.example.com/my-function:abc123\"}",
  "metadata": "{\"source\":\"knative-lambda-operator\",\"specversion\":\"1.0\",\"type\":\"lambda.build.started\"}"
}

// Record 3: Build Completed Event
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "event_id": "770e8400-e29b-41d4-a716-446655440002",
  "event_type": "lambda.build.completed",
  "timestamp": "2025-01-20T10:05:30.345678901Z",
  "function_name": "my-lambda-function",
  "data": "{\"function_name\":\"my-lambda-function\",\"build_id\":\"build-123\",\"image\":\"registry.example.com/my-function:abc123\",\"duration_seconds\":325,\"status\":\"success\"}",
  "metadata": "{\"source\":\"knative-lambda-operator\",\"specversion\":\"1.0\",\"type\":\"lambda.build.completed\"}"
}

// Record 4: Build Failed Event
{
  "id": "880e8400-e29b-41d4-a716-446655440003",
  "event_id": "880e8400-e29b-41d4-a716-446655440003",
  "event_type": "lambda.build.failed",
  "timestamp": "2025-01-20T10:10:15.456789012Z",
  "function_name": "another-function",
  "data": "{\"function_name\":\"another-function\",\"build_id\":\"build-124\",\"error\":\"Dockerfile parse error: invalid syntax\",\"duration_seconds\":45}",
  "metadata": "{\"source\":\"knative-lambda-operator\",\"specversion\":\"1.0\",\"type\":\"lambda.build.failed\"}"
}
```

**Query Examples**:

```python
# Query events by function name
events_table = db.open_table("events")
function_events = events_table.search().where(
    "function_name = 'my-lambda-function'"
).to_pandas()

# Query events by time range
recent_events = events_table.search().where(
    "timestamp >= '2025-01-20T10:00:00Z' AND timestamp <= '2025-01-20T11:00:00Z'"
).to_pandas()

# Query events by type
build_events = events_table.search().where(
    "event_type LIKE 'lambda.build.%'"
).to_pandas()

# Query failed builds
failed_builds = events_table.search().where(
    "event_type = 'lambda.build.failed'"
).to_pandas()
```

**Use Cases**:
- Historical event analysis
- Function lifecycle tracking
- Error pattern identification
- Training data collection
- Audit logging

---

#### 2. Embeddings Table

**Purpose**: Stores vector embeddings of events for similarity search and pattern matching.

**Schema**:
```python
embeddings_schema = pa.schema([
    pa.field("id", pa.string()),                    # Unique embedding ID (UUID)
    pa.field("event_id", pa.string()),              # Reference to events table
    pa.field("embedding", pa.list_(pa.float32())), # Vector embedding (e.g., 384-dim from sentence-transformers)
    pa.field("event_type", pa.string()),            # Event type for filtering
    pa.field("timestamp", pa.timestamp("ns")),     # Event timestamp
])
```

**Example Records**:

```json
// Record 1: Embedding for Function Created Event
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "embedding": [0.123, -0.456, 0.789, ..., 0.234],  // 384-dimensional vector
  "event_type": "lambda.function.created",
  "timestamp": "2025-01-20T10:00:00.123456789Z"
}

// Record 2: Embedding for Build Started Event
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "event_id": "660e8400-e29b-41d4-a716-446655440001",
  "embedding": [0.234, -0.567, 0.890, ..., 0.345],  // 384-dimensional vector
  "event_type": "lambda.build.started",
  "timestamp": "2025-01-20T10:00:05.234567890Z"
}

// Record 3: Embedding for Build Failed Event
{
  "id": "880e8400-e29b-41d4-a716-446655440003",
  "event_id": "880e8400-e29b-41d4-a716-446655440003",
  "embedding": [-0.123, 0.456, -0.789, ..., -0.234],  // 384-dimensional vector (different pattern)
  "event_type": "lambda.build.failed",
  "timestamp": "2025-01-20T10:10:15.456789012Z"
}
```

**Vector Embedding Generation**:

```python
from sentence_transformers import SentenceTransformer

# Initialize model (384-dimensional embeddings)
model = SentenceTransformer('all-MiniLM-L6-v2')

def generate_embedding(event: LambdaEvent) -> List[float]:
    """Generate vector embedding from event"""
    # Create text representation of event
    event_text = f"""
    Event Type: {event.type.value}
    Function: {event.data.function_name if hasattr(event.data, 'function_name') else 'N/A'}
    Timestamp: {event.time}
    Data: {event.data.json()}
    """
    
    # Generate embedding
    embedding = model.encode(event_text, convert_to_numpy=True).tolist()
    return embedding  # Returns 384-dimensional vector
```

**Similarity Search Examples**:

```python
# Find similar events to a given event
def find_similar_events(event_id: str, limit: int = 10) -> List[Dict]:
    embeddings_table = db.open_table("embeddings")
    
    # Get embedding for the query event
    query_event = embeddings_table.search().where(
        f"event_id = '{event_id}'"
    ).to_pandas()
    
    if query_event.empty:
        return []
    
    query_embedding = query_event.iloc[0]['embedding']
    
    # Vector similarity search (cosine similarity)
    results = embeddings_table.search(query_embedding).limit(limit).to_pandas()
    
    return results.to_dict("records")

# Example: Find events similar to a build failure
similar_failures = find_similar_events(
    event_id="880e8400-e29b-41d4-a716-446655440003",
    limit=5
)
# Returns: List of similar build failure events with similarity scores
```

**Use Cases**:
- Finding similar events (e.g., similar build failures)
- Anomaly detection (events with low similarity to normal patterns)
- Pattern clustering (grouping similar events)
- Recommendation system (suggest actions based on similar past events)

---

#### 3. Patterns Table

**Purpose**: Stores detected patterns, anomalies, and insights from event sequences.

**Schema**:
```python
patterns_schema = pa.schema([
    pa.field("id", pa.string()),                    # Unique pattern ID (UUID)
    pa.field("pattern_type", pa.string()),         # Type: "anomaly", "success", "failure", "sequence"
    pa.field("event_sequence", pa.list_(pa.string())), # List of event IDs forming the pattern
    pa.field("embedding", pa.list_(pa.float32())),   # Vector embedding of the pattern
    pa.field("metadata", pa.string()),             # JSON string with pattern details
])
```

**Example Records**:

```json
// Record 1: Success Pattern (Normal Build Flow)
{
  "id": "pattern-001",
  "pattern_type": "success",
  "event_sequence": [
    "550e8400-e29b-41d4-a716-446655440000",  // function.created
    "660e8400-e29b-41d4-a716-446655440001",  // build.started
    "770e8400-e29b-41d4-a716-446655440002"   // build.completed
  ],
  "embedding": [0.1, -0.2, 0.3, ..., 0.4],  // Pattern embedding
  "metadata": "{\"pattern_name\":\"normal_build_flow\",\"frequency\":150,\"success_rate\":0.95,\"avg_duration_seconds\":320,\"description\":\"Normal function creation and build completion flow\"}"
}

// Record 2: Failure Pattern (Build Timeout)
{
  "id": "pattern-002",
  "pattern_type": "failure",
  "event_sequence": [
    "550e8400-e29b-41d4-a716-446655440000",  // function.created
    "660e8400-e29b-41d4-a716-446655440001",  // build.started
    "990e8400-e29b-41d4-a716-446655440004"   // build.timeout
  ],
  "embedding": [-0.1, 0.2, -0.3, ..., -0.4],  // Pattern embedding
  "metadata": "{\"pattern_name\":\"build_timeout\",\"frequency\":5,\"failure_rate\":1.0,\"avg_duration_seconds\":1800,\"description\":\"Build process times out after 30 minutes\",\"common_causes\":[\"large_dependencies\",\"slow_registry\",\"resource_constraints\"]}"
}

// Record 3: Anomaly Pattern (Rapid Build Failures)
{
  "id": "pattern-003",
  "pattern_type": "anomaly",
  "event_sequence": [
    "880e8400-e29b-41d4-a716-446655440003",  // build.failed
    "880e8400-e29b-41d4-a716-446655440005",  // build.failed
    "880e8400-e29b-41d4-a716-446655440006",  // build.failed
    "880e8400-e29b-41d4-a716-446655440007"   // build.failed
  ],
  "embedding": [0.5, -0.5, 0.5, ..., -0.5],  // Pattern embedding
  "metadata": "{\"pattern_name\":\"rapid_build_failures\",\"frequency\":2,\"severity\":\"high\",\"time_window_seconds\":300,\"description\":\"Multiple build failures within 5 minutes\",\"alert_threshold\":3,\"recommended_action\":\"Check registry connectivity and resource availability\"}"
}

// Record 4: Sequence Pattern (Function Lifecycle)
{
  "id": "pattern-004",
  "pattern_type": "sequence",
  "event_sequence": [
    "550e8400-e29b-41d4-a716-446655440000",  // function.created
    "660e8400-e29b-41d4-a716-446655440001",  // build.started
    "770e8400-e29b-41d4-a716-446655440002",  // build.completed
    "aa0e8400-e29b-41d4-a716-446655440008",  // service.created
    "bb0e8400-e29b-41d4-a716-446655440009"   // status.updated (Ready)
  ],
  "embedding": [0.2, -0.3, 0.4, ..., 0.5],  // Pattern embedding
  "metadata": "{\"pattern_name\":\"complete_function_lifecycle\",\"frequency\":120,\"avg_total_duration_seconds\":450,\"stages\":[\"creation\",\"build\",\"deployment\",\"ready\"],\"description\":\"Complete function lifecycle from creation to ready state\"}"
}
```

**Pattern Detection Logic**:

```python
def detect_patterns(event_sequence: List[LambdaEvent]) -> List[Dict]:
    """Detect patterns in event sequences"""
    patterns = []
    
    # Pattern 1: Success pattern (created -> started -> completed)
    if len(event_sequence) >= 3:
        types = [e.type.value for e in event_sequence]
        if (types[0] == "lambda.function.created" and
            types[1] == "lambda.build.started" and
            types[2] == "lambda.build.completed"):
            pattern = {
                "pattern_type": "success",
                "event_sequence": [e.id for e in event_sequence[:3]],
                "metadata": {
                    "pattern_name": "normal_build_flow",
                    "duration_seconds": (event_sequence[2].time - event_sequence[0].time).total_seconds()
                }
            }
            patterns.append(pattern)
    
    # Pattern 2: Failure pattern (started -> failed)
    if len(event_sequence) >= 2:
        types = [e.type.value for e in event_sequence]
        if (types[0] == "lambda.build.started" and
            types[1] == "lambda.build.failed"):
            pattern = {
                "pattern_type": "failure",
                "event_sequence": [e.id for e in event_sequence[:2]],
                "metadata": {
                    "pattern_name": "build_failure",
                    "error": event_sequence[1].data.error if hasattr(event_sequence[1].data, 'error') else None
                }
            }
            patterns.append(pattern)
    
    # Pattern 3: Anomaly detection (rapid failures)
    failure_events = [e for e in event_sequence if "failed" in e.type.value]
    if len(failure_events) >= 3:
        time_window = (failure_events[-1].time - failure_events[0].time).total_seconds()
        if time_window < 300:  # 3+ failures in 5 minutes
            pattern = {
                "pattern_type": "anomaly",
                "event_sequence": [e.id for e in failure_events],
                "metadata": {
                    "pattern_name": "rapid_failures",
                    "severity": "high",
                    "time_window_seconds": time_window,
                    "count": len(failure_events)
                }
            }
            patterns.append(pattern)
    
    return patterns
```

**Pattern Query Examples**:

```python
# Query patterns by type
patterns_table = db.open_table("patterns")
success_patterns = patterns_table.search().where(
    "pattern_type = 'success'"
).to_pandas()

# Query anomaly patterns
anomalies = patterns_table.search().where(
    "pattern_type = 'anomaly'"
).to_pandas()

# Query patterns containing specific event
patterns_with_event = patterns_table.search().where(
    "event_sequence CONTAINS '550e8400-e29b-41d4-a716-446655440000'"
).to_pandas()

# Find similar patterns using vector search
def find_similar_patterns(pattern_id: str, limit: int = 5) -> List[Dict]:
    patterns_table = db.open_table("patterns")
    
    # Get pattern embedding
    query_pattern = patterns_table.search().where(
        f"id = '{pattern_id}'"
    ).to_pandas()
    
    if query_pattern.empty:
        return []
    
    query_embedding = query_pattern.iloc[0]['embedding']
    
    # Vector similarity search
    results = patterns_table.search(query_embedding).limit(limit).to_pandas()
    return results.to_dict("records")
```

**Use Cases**:
- Anomaly detection (identify unusual event sequences)
- Success pattern recognition (learn from successful deployments)
- Failure pattern analysis (identify common failure modes)
- Predictive insights (predict outcomes based on patterns)
- Alert generation (trigger alerts for anomaly patterns)

---

#### Table Relationships & Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│              LanceDB Table Relationships                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Events Table (Source of Truth)                             │
│  ┌──────────────────────────────────────┐                   │
│  │ id (PK)                              │                   │
│  │ event_id                             │                   │
│  │ event_type                           │                   │
│  │ timestamp                            │                   │
│  │ function_name                        │                   │
│  │ data (JSON)                          │                   │
│  │ metadata (JSON)                      │                   │
│  └──────────────┬───────────────────────┘                   │
│                 │                                           │
│                 │ 1:1 relationship                          │
│                 ▼                                           │
│  Embeddings Table (Vector Search)                           │
│  ┌──────────────────────────────────────┐                   │
│  │ id (PK)                              │                   │
│  │ event_id (FK → events.id)            │                   │
│  │ embedding (384-dim vector)           │                   │
│  │ event_type                           │                   │
│  │ timestamp                            │                   │
│  └──────────────┬───────────────────────┘                   │
│                 │                                           │
│                 │ Many:1 relationship                       │
│                 ▼                                           │
│  Patterns Table (Aggregated Insights)                       │
│  ┌──────────────────────────────────────┐                   │
│  │ id (PK)                              │                   │
│  │ pattern_type                         │                   │
│  │ event_sequence (FK[] → events.id)    │                   │
│  │ embedding (pattern vector)           │                   │
│  │ metadata (JSON)                      │                   │
│  └──────────────────────────────────────┘                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Data Flow Example**:

```python
# 1. Event arrives
event = LambdaEvent(...)

# 2. Store in events table
events_table.add([{
    "id": event.id,
    "event_id": event.id,
    "event_type": event.type.value,
    "timestamp": event.time,
    "function_name": event.data.function_name,
    "data": event.data.json(),
    "metadata": event.dict(exclude={"data"}).json()
}])

# 3. Generate embedding
embedding = generate_embedding(event)  # 384-dim vector

# 4. Store in embeddings table
embeddings_table.add([{
    "id": event.id,
    "event_id": event.id,
    "embedding": embedding,
    "event_type": event.type.value,
    "timestamp": event.time
}])

# 5. Detect patterns (periodic batch job)
recent_events = get_recent_events(time_window="1h")
patterns = detect_patterns(recent_events)

# 6. Store patterns
for pattern in patterns:
    pattern_embedding = generate_pattern_embedding(pattern)
    patterns_table.add([{
        "id": generate_uuid(),
        "pattern_type": pattern["pattern_type"],
        "event_sequence": pattern["event_sequence"],
        "embedding": pattern_embedding,
        "metadata": json.dumps(pattern["metadata"])
    }])
```

**Performance Considerations**:

- **Indexing**: LanceDB automatically creates vector indexes for similarity search
- **Partitioning**: Consider partitioning by `event_type` or `timestamp` for large datasets
- **Retention**: Implement data retention policies (e.g., keep events for 90 days, patterns indefinitely)
- **Batch Operations**: Use batch inserts for better performance (100-1000 records per batch)

---

### Event Collection Pipeline for Fine-tuning

The AI Agent collects events for fine-tuning with proper labeling and export:

```python
from typing import List, Dict, Any, Optional
import json
from datetime import datetime, timedelta
import pandas as pd
from minio import Minio
from minio.error import S3Error
import boto3  # Optional: only if using AWS S3
from google.cloud import storage  # Optional: only if using GCS

class TrainingDataCollector:
    """Collects and prepares events for fine-tuning"""
    
    def __init__(self, 
                 storage_type: str = "minio",  # minio, s3, gcs
                 bucket: str = "knative-lambda-training-data",
                 lancedb_store: EventVectorStore = None,
                 # MinIO config (default)
                 minio_endpoint: str = "minio.minio.svc.cluster.local:9000",
                 minio_access_key: Optional[str] = None,
                 minio_secret_key: Optional[str] = None,
                 # AWS S3 config (optional)
                 s3_region: Optional[str] = None,
                 # GCS config (optional)
                 gcs_project: Optional[str] = None):
        self.storage_type = storage_type
        self.bucket = bucket
        self.lancedb = lancedb_store
        
        # Initialize storage client based on type
        if storage_type == "minio":
            self.storage_client = Minio(
                minio_endpoint,
                access_key=minio_access_key or "minioadmin",
                secret_key=minio_secret_key or "minioadmin",
                secure=False  # Use TLS in production
            )
        elif storage_type == "s3":
            self.storage_client = boto3.client('s3', region_name=s3_region)
        elif storage_type == "gcs":
            self.storage_client = storage.Client(project=gcs_project)
        else:
            raise ValueError(f"Unsupported storage type: {storage_type}")
    
    def collect_labeled_events(self, 
                               start_time: datetime,
                               end_time: datetime,
                               labels: Dict[str, str]) -> List[Dict]:
        """Collect events with labels for training"""
        events_table = self.lancedb.db.open_table("events")
        
        # Query events in time range
        query = events_table.search().where(
            f"timestamp >= '{start_time.isoformat()}' AND "
            f"timestamp <= '{end_time.isoformat()}'"
        )
        
        events = query.to_pandas()
        
        # Add labels to events
        labeled_events = []
        for _, event in events.iterrows():
            event_dict = json.loads(event['data'])
            event_dict['label'] = labels.get(event['event_type'], 'unknown')
            event_dict['event_id'] = event['event_id']
            event_dict['timestamp'] = event['timestamp']
            labeled_events.append(event_dict)
        
        return labeled_events
    
    def export_to_jsonl(self, events: List[Dict], output_path: str):
        """Export events to JSONL format for fine-tuning"""
        with open(output_path, 'w') as f:
            for event in events:
                f.write(json.dumps(event) + '\n')
    
    def upload_to_storage(self, local_path: str, object_key: str):
        """Upload training dataset to storage (MinIO/S3/GCS)"""
        if self.storage_type == "minio":
            self.storage_client.fput_object(
                self.bucket,
                object_key,
                local_path
            )
        elif self.storage_type == "s3":
            self.storage_client.upload_file(
                local_path,
                self.bucket,
                object_key
            )
        elif self.storage_type == "gcs":
            bucket = self.storage_client.bucket(self.bucket)
            blob = bucket.blob(object_key)
            blob.upload_from_filename(local_path)
    
    def collect_event(self, event: LambdaEvent):
        """Collect a single event for training (async-friendly)"""
        # Events are already stored in LanceDB by the main handler
        # This method can be used for additional processing or labeling
        pass
    
    def prepare_training_dataset(self, 
                                 time_range_days: int = 30,
                                 output_key: str = None):
        """Prepare complete training dataset from collected events"""
        end_time = datetime.utcnow()
        start_time = end_time - timedelta(days=time_range_days)
        
        # Define labels based on event outcomes
        labels = {
            "lambda.build.completed": "success",
            "lambda.build.failed": "failure",
            "lambda.parser.completed": "success",
            "lambda.parser.failed": "failure",
            "lambda.status.updated": "info",
            # ... more labels
        }
        
        # Collect labeled events
        labeled_events = self.collect_labeled_events(start_time, end_time, labels)
        
        # Export to JSONL
        local_path = f"/tmp/training_data_{datetime.utcnow().isoformat()}.jsonl"
        self.export_to_jsonl(labeled_events, local_path)
        
        # Upload to storage (MinIO/S3/GCS)
        if output_key is None:
            output_key = f"training-data/{datetime.utcnow().strftime('%Y/%m/%d')}/dataset.jsonl"
        
        self.upload_to_storage(local_path, output_key)
        
        # Generate storage URL based on type
        if self.storage_type == "minio":
            storage_url = f"minio://{self.bucket}/{output_key}"
        elif self.storage_type == "s3":
            storage_url = f"s3://{self.bucket}/{output_key}"
        elif self.storage_type == "gcs":
            storage_url = f"gs://{self.bucket}/{output_key}"
        
        return {
            "event_count": len(labeled_events),
            "storage_location": storage_url,
            "storage_type": self.storage_type,
            "time_range": {
                "start": start_time.isoformat(),
                "end": end_time.isoformat()
            }
        }
```

### Complete AI Agent Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│         AI Agent Event Processing & Training Flow           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. EVENT RECEPTION                                         │
│     CloudEvent → FastAPI Handler                            │
│                                                             │
│  2. VALIDATION (Pydantic)                                   │
│     Raw Event → LambdaEvent (validated)                     │
│                                                             │
│  3. FEATURE EXTRACTION                                      │
│     LambdaEvent → Features:                                 │
│     - Event type                                            │
│     - Timestamp                                             │
│     - Function metadata                                     │
│     - Status information                                    │
│                                                             │
│  4. EMBEDDING GENERATION                                    │
│     Features → Vector Embedding (via ML model)              │
│                                                             │
│  5. STORAGE (LanceDB)                                       │
│     ├─ Events Table: Raw event data                         │
│     ├─ Embeddings Table: Vector embeddings                  │
│     └─ Patterns Table: Detected patterns                    │
│                                                             │
│  6. ANALYSIS                                                │
│     ├─ Similarity Search (LanceDB)                          │
│     ├─ Anomaly Detection                                    │
│     ├─ Pattern Matching                                     │
│     └─ Prediction Generation                                │
│                                                             │
│  7. TRAINING DATA COLLECTION                                │
│     ├─ Label Events (success/failure)                       │
│     ├─ Export to JSONL                                      │
│     └─ Upload to MinIO/S3/GCS for Fine-tuning               │
│                                                             │
│  8. FINE-TUNING PREPARATION                                 │
│     Storage Dataset → Training Pipeline                     │
│     (MinIO default, S3/GCS optional)                        │
│     - Format conversion                                     │
│     - Data augmentation                                     │
│     - Model fine-tuning                                     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### AI Agent Service Implementation

The AI Agent is a **first-class observability citizen** that processes events, investigates alerts from Alertmanager, queries Prometheus, and provides intelligent insights. It exposes comprehensive metrics and integrates with the full observability stack.

```python
from fastapi import FastAPI, HTTPException, Request
from cloudevents.http import from_http
from cloudevents.pydantic import CloudEvent
from pydantic import ValidationError
from typing import List, Dict, Optional
from datetime import datetime
import uvicorn
import asyncio
import os
from sentence_transformers import SentenceTransformer
from prometheus_client import Counter, Histogram, Gauge, generate_latest, CONTENT_TYPE_LATEST
from starlette.responses import Response
import httpx
import json
import logfire
from logfire import with_attributes, instrument

# Initialize Logfire (optional, only if enabled)
if os.getenv("LOGFIRE_ENABLED", "false").lower() == "true":
    logfire.configure(
        token=os.getenv("LOGFIRE_TOKEN"),
        service_name="knative-lambda-ai-agent",
        environment=os.getenv("ENVIRONMENT", "production")
    )

app = FastAPI(title="Knative Lambda AI Agent")

# Instrument FastAPI with Logfire (optional, only if enabled)
if os.getenv("LOGFIRE_ENABLED", "false").lower() == "true":
    logfire.instrument_fastapi(app)

# Prometheus metrics (first-class observability)
events_processed = Counter(
    'ai_agent_events_processed_total',
    'Total events processed',
    ['event_type', 'result']
)

embedding_duration = Histogram(
    'ai_agent_embedding_generation_duration_seconds',
    'Embedding generation time',
    buckets=[0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0]
)

similarity_search_duration = Histogram(
    'ai_agent_similarity_search_duration_seconds',
    'Similarity search time',
    buckets=[0.01, 0.05, 0.1, 0.5, 1.0, 2.0]
)

patterns_detected = Counter(
    'ai_agent_patterns_detected_total',
    'Patterns detected',
    ['pattern_type']
)

alerts_investigated = Counter(
    'ai_agent_alerts_investigated_total',
    'Alerts investigated',
    ['alert_name', 'severity', 'result']
)

investigation_duration = Histogram(
    'ai_agent_investigation_duration_seconds',
    'Investigation time',
    buckets=[1.0, 5.0, 10.0, 30.0, 60.0, 120.0]
)

investigation_queue_depth = Gauge(
    'ai_agent_investigation_queue_depth',
    'Pending investigations'
)

prometheus_queries = Counter(
    'ai_agent_prometheus_queries_total',
    'Prometheus queries executed',
    ['query_type']
)

# Initialize components
vector_store = EventVectorStore(db_path="/data/lancedb")
training_collector = TrainingDataCollector(
    storage_type="minio",  # Default: MinIO for local development
    bucket="knative-lambda-training-data",
    lancedb_store=vector_store,
    minio_endpoint="minio.minio.svc.cluster.local:9000"
    # For AWS S3: storage_type="s3", s3_region="us-east-1"
    # For GCS: storage_type="gcs", gcs_project="my-project"
)

# Prometheus client for alert investigation
prometheus_client = PrometheusClient(
    base_url="http://prometheus.monitoring.svc.cluster.local:9090"
)

# Alert investigation service
alert_investigation_service = AlertInvestigationService(
    prometheus_client=prometheus_client,
    lancedb_store=vector_store
)

# Metrics endpoint
@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint"""
    return Response(content=generate_latest(), media_type=CONTENT_TYPE_LATEST)

# Initialize embedding model
embedding_model = SentenceTransformer('all-MiniLM-L6-v2')

@logfire.instrument
def generate_embedding(event: LambdaEvent) -> List[float]:
    """Generate vector embedding from event"""
    with embedding_duration.time():
        event_text = f"""
        Event Type: {event.type.value}
        Function: {event.data.function_name if hasattr(event.data, 'function_name') else 'N/A'}
        Timestamp: {event.time}
        Data: {event.data.json()}
        """
        embedding = embedding_model.encode(event_text, convert_to_numpy=True).tolist()
        return embedding  # Returns 384-dimensional vector

@logfire.instrument
def analyze_event(event: LambdaEvent, embedding: List[float]) -> Dict:
    """Analyze event and generate insights"""
    # Similarity search in LanceDB
    with similarity_search_duration.time():
        similar_events = vector_store.search_similar_events(embedding, limit=5)
    
    # Anomaly detection
    is_anomaly = detect_anomaly(event, similar_events)
    
    # Pattern matching
    patterns = vector_store.find_patterns([event.id])
    
    # Track patterns detected
    for pattern in patterns:
        patterns_detected.labels(pattern_type=pattern.get("pattern_type", "unknown")).inc()
    
    return {
        "similar_events": len(similar_events),
        "is_anomaly": is_anomaly,
        "patterns": patterns,
        "timestamp": datetime.utcnow().isoformat()
    }

def detect_anomaly(event: LambdaEvent, similar_events: List[Dict]) -> bool:
    """Detect if event is anomalous based on similar events"""
    # Simple anomaly detection logic
    # More sophisticated implementation would use statistical methods
    if len(similar_events) < 2:
        return True  # Rare event pattern
    return False

def should_collect_for_training(event: LambdaEvent) -> bool:
    """Determine if event should be collected for training"""
    # Collect events that have clear outcomes (success/failure)
    training_events = [
        "lambda.build.completed",
        "lambda.build.failed",
        "lambda.parser.completed",
        "lambda.parser.failed"
    ]
    return event.type.value in training_events

@app.post("/api/events")
# Note: Logfire decorators are only effective if LOGFIRE_ENABLED=true
@logfire.with_attributes(
    event_type=lambda req: from_http(req.headers, req.body()).get("type", "unknown")
)
async def handle_event(request: Request):
    """Handle incoming CloudEvent with observability and Logfire tracing (optional)"""
    try:
        # Parse CloudEvent
        event = from_http(request.headers, await request.body())
        
        # Validate with Pydantic
        lambda_event = LambdaEvent(**event)
        
        # Generate embedding (automatically traced by Logfire)
        embedding = generate_embedding(lambda_event)
        
        # Store in LanceDB
        vector_store.store_event(lambda_event, embedding)
        
        # Analyze event (automatically traced by Logfire)
        analysis = analyze_event(lambda_event, embedding)
        
        # Collect for training (async)
        if should_collect_for_training(lambda_event):
            # Run async to not block response
            asyncio.create_task(
                training_collector.collect_event(lambda_event)
            )
        
        # Track metrics
        events_processed.labels(
            event_type=lambda_event.type.value,
            result="success"
        ).inc()
        
        return {"status": "processed", "analysis": analysis}
        
    except ValidationError as e:
        # Track validation errors
        events_processed.labels(
            event_type="unknown",
            result="validation_error"
        ).inc()
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        # Track processing errors
        events_processed.labels(
            event_type="unknown",
            result="error"
        ).inc()
        logger.error(f"Error processing event: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

# Alertmanager webhook endpoint (already defined above in Alert Investigation section)
# This endpoint receives alerts and triggers investigation

### Training Data Export Schedule

The AI Agent periodically exports training data:

```python
from apscheduler.schedulers.background import BackgroundScheduler

def scheduled_training_export():
    """Scheduled job to export training data"""
    collector = TrainingDataCollector(
        storage_type="minio",  # Default: MinIO
        bucket="knative-lambda-training-data"
    )
    result = collector.prepare_training_dataset(
        time_range_days=30,
        output_key=None  # Auto-generate path
    )
    logger.info(f"Exported {result['event_count']} events to {result['storage_location']}")

# Schedule weekly exports
scheduler = BackgroundScheduler()
scheduler.add_job(
    scheduled_training_export,
    trigger='cron',
    day_of_week='sunday',
    hour=2
)
scheduler.start()
```

## 🔌 Integration Patterns

### Service Mesh Integration (Linkerd)

Linkerd provides automatic mTLS, traffic policies, circuit breakers, **topology aware routing**, and **first-class observability** for all services in the mesh. This section details Policy API configurations (Server + ServerAuthorization), circuit breaker patterns, rate limiting, HTTP access logging, and topology aware routing.

> **Note**: ServiceProfile (`linkerd.io/v1alpha2`) is deprecated. The modern approach uses the Policy API (`Server` + `ServerAuthorization`) for authorization and port configuration. For retries and timeouts, use HTTPRoute (Gateway API) or ServiceProfile (deprecated but still functional).

#### Linkerd Policy API: Comprehensive Configuration

**1. Operator Server & Authorization** (REST API):

```yaml
# Server: Defines port and protocol
apiVersion: policy.linkerd.io/v1beta3
kind: Server
metadata:
  name: operator-server
  namespace: knative-lambda
spec:
  podSelector:
    matchLabels:
      app: knative-lambda-operator
  port: 8080
  proxyProtocol: HTTP/1

---
# ServerAuthorization: Defines who can access
apiVersion: policy.linkerd.io/v1beta1
kind: ServerAuthorization
metadata:
  name: operator-auth
  namespace: knative-lambda
spec:
  server:
    name: operator-server
  client:
    networks:
    - cidr: 10.0.0.0/8
    - cidr: 172.16.0.0/12
    - cidr: 192.168.0.0/16
    meshTLS:
      identities:
      - "*.knative-lambda.svc.cluster.local"
      - "*.knative-serving.svc.cluster.local"
      - "*.linkerd-viz.svc.cluster.local"

---
# HTTPRoute: Retries and timeouts (Gateway API)
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: operator-routes
  namespace: knative-lambda
spec:
  parentRefs:
  - name: operator-server
    kind: Server
    group: policy.linkerd.io
  rules:
  # High-priority: Function creation
  - matches:
    - path:
        type: PathPrefix
        value: /api/v1/lambda/functions
      method: POST
    backendRefs:
    - name: knative-lambda-operator
      port: 8080
    timeouts:
      request: 30s
    filters:
    - type: RequestMirror
    retries:
      attempts: 3
      perTryTimeout: 10s
  
  # Read operations: Lower timeout, no retries
  - matches:
    - path:
        type: PathPrefix
        value: /api/v1/lambda/functions
      method: GET
    backendRefs:
    - name: knative-lambda-operator
      port: 8080
    timeouts:
      request: 10s
  
  # Function updates
  - matches:
    - path:
        type: PathPrefix
        value: /api/v1/lambda/functions
      method: PUT
    backendRefs:
    - name: knative-lambda-operator
      port: 8080
    timeouts:
      request: 30s
    retries:
      attempts: 2
      perTryTimeout: 15s
  
  # Function deletion: Fast timeout, no retries
  - matches:
    - path:
        type: PathPrefix
        value: /api/v1/lambda/functions
      method: DELETE
    backendRefs:
    - name: knative-lambda-operator
      port: 8080
    timeouts:
      request: 5s
  
  # Health check endpoint
  - matches:
    - path:
        type: PathPrefix
        value: /health
      method: GET
    backendRefs:
    - name: knative-lambda-operator
      port: 8080
    timeouts:
      request: 2s
```

**2. Lambda Function Servers** (Auto-generated per function):

```yaml
# Server for Lambda function
apiVersion: policy.linkerd.io/v1beta3
kind: Server
metadata:
  name: ${function_name}-server  # e.g., my-lambda-function-server
  namespace: knative-lambda
spec:
  podSelector:
    matchLabels:
      serving.knative.dev/service: ${function_name}
  port: 8080
  proxyProtocol: HTTP/1

---
# ServerAuthorization for Lambda function
apiVersion: policy.linkerd.io/v1beta1
kind: ServerAuthorization
metadata:
  name: ${function_name}-auth
  namespace: knative-lambda
spec:
  server:
    name: ${function_name}-server
  client:
    meshTLS:
      identities:
      - "*.knative-lambda.svc.cluster.local"
      - "*.knative-serving.svc.cluster.local"
```

**3. Notifi Services Servers** (gRPC):

```yaml
# Server for Storage Manager
apiVersion: policy.linkerd.io/v1beta3
kind: Server
metadata:
  name: storage-manager-server
  namespace: notifi
spec:
  podSelector:
    matchLabels:
      app: storage-manager
  port: 4000
  proxyProtocol: HTTP/2  # gRPC uses HTTP/2

---
# ServerAuthorization for Storage Manager
apiVersion: policy.linkerd.io/v1beta1
kind: ServerAuthorization
metadata:
  name: storage-manager-auth
  namespace: notifi
spec:
  server:
    name: storage-manager-server
  client:
    networks:
    - cidr: 10.0.0.0/8
    - cidr: 172.16.0.0/12
    - cidr: 192.168.0.0/16
    meshTLS:
      identities:
      - "*.knative-lambda.svc.cluster.local"
      - "*.notifi.svc.cluster.local"
```

**4. AI Agent Server & Authorization**:

```yaml
# Server for AI Agent
apiVersion: policy.linkerd.io/v1beta3
kind: Server
metadata:
  name: ai-agent-server
  namespace: knative-lambda
spec:
  podSelector:
    matchLabels:
      app: knative-lambda-ai-agent
  port: 8080
  proxyProtocol: HTTP/1

---
# ServerAuthorization for AI Agent
apiVersion: policy.linkerd.io/v1beta1
kind: ServerAuthorization
metadata:
  name: ai-agent-auth
  namespace: knative-lambda
spec:
  server:
    name: ai-agent-server
  client:
    meshTLS:
      identities:
      - "*.knative-lambda.svc.cluster.local"
      - "*.linkerd-viz.svc.cluster.local"
      - "*.prometheus.svc.cluster.local"
```

#### Linkerd Circuit Breakers: Detailed Configuration

Linkerd implements circuit breakers through **retry budgets** and **failure accrual**. When a service experiences failures, Linkerd automatically opens the circuit to prevent cascading failures.

**Circuit Breaker Behavior**:

```yaml
# Circuit breaker opens when:
# 1. Retry budget exhausted (retryRatio exceeded)
# 2. Failure rate > 50% over 10s window
# 3. Consecutive failures >= 3

# Example: Operator API circuit breaker
retries:
  budget:
    retryRatio: 0.2      # Max 20% of requests can be retries
    minRetriesPerSecond: 10  # Minimum retry rate
    ttl: 10s             # Budget refresh window
  isRetryable:
    statusCodes: [500, 502, 503, 504]  # Only retry these errors
```

**Circuit Breaker States**:

1. **Closed** (Normal): All requests pass through
2. **Open** (Failing): Circuit opened, requests fail fast
3. **Half-Open** (Recovering): Testing if service recovered

**Automatic Circuit Breaker Configuration**:

Linkerd automatically manages circuit breakers based on:
- **Failure accrual**: Tracks consecutive failures
- **Success rate**: Monitors success rate over time window
- **Latency**: Tracks P50, P95, P99 latencies
- **Retry budget**: Limits retry attempts

**Manual Circuit Breaker Configuration** (via Policy API):

```yaml
# Server with connection limits
apiVersion: policy.linkerd.io/v1beta3
kind: Server
metadata:
  name: operator-server
  namespace: knative-lambda
spec:
  podSelector:
    matchLabels:
      app: knative-lambda-operator
  port: 8080
  proxyProtocol: HTTP/1
  # Connection limits act as circuit breaker
  clientConnectionPolicy:
    maxConnections: 100
    maxPendingRequests: 50
```

#### Linkerd mTLS Configuration

**Automatic mTLS**:
- **Enabled by default**: All meshed services automatically use mTLS
- **Certificate rotation**: Automatic, no manual intervention
- **Zero-trust**: All traffic encrypted and authenticated
- **Policy enforcement**: Via Linkerd policy resources

**mTLS Verification**:

```bash
# Verify mTLS is enabled
linkerd viz stat deploy -n knative-lambda

# Check mTLS status per service
linkerd viz tap deploy/knative-lambda-operator -n knative-lambda
```

#### Linkerd Observability Features

**1. Automatic Metrics** (Prometheus):
- `request_total`: Total requests per route
- `response_latency_ms`: P50, P95, P99 latencies
- `response_total`: Success/failure counts
- `retry_total`: Retry attempts and successes
- `circuit_breaker_state`: Open/closed/half-open state

**2. Service Topology**:
- Visual service dependency graph
- Real-time traffic flow visualization
- Service health indicators

**3. Distributed Tracing** (Fully Integrated):
- **Linkerd Automatic Spans**: Proxy automatically emits spans when W3C Trace Context headers present
- **Operator Tracing**: OpenTelemetry integration with Tempo backend
- **AI Agent Tracing**: OpenTelemetry/Tempo (primary) with Logfire (optional fallback)
- **Notifi Services Tracing**: All Notifi services export traces to Tempo via OpenTelemetry
- **Knative Services Tracing**: Lambda functions propagate W3C Trace Context headers
- **Request Correlation**: W3C Trace Context headers propagate across all services (Ingress → Linkerd → Operator → AI Agent → Lambda Functions → Notifi Services)
- **Unified Backend**: Tempo as single source of truth for all distributed traces

**4. Tap (Real-time Request Inspection)**:
```bash
# Tap into operator traffic
linkerd viz tap deploy/knative-lambda-operator -n knative-lambda

# Tap into specific route
linkerd viz tap deploy/knative-lambda-operator -n knative-lambda --path /api/v1/lambda/functions

# Tap with filters
linkerd viz tap deploy/knative-lambda-operator -n knative-lambda --method POST
```

**5. Linkerd Dashboard**:
- Real-time metrics per service
- Circuit breaker state visualization
- Retry budget usage
- Service dependency graph
- Request/response inspection

**6. HTTP Access Logging**:
- Detailed HTTP request/response logging via `HTTPAccessLogPolicy`
- JSON format with method, path, status, latency, client/server identity
- Integrated with Loki for log aggregation and analysis
- Security auditing and performance debugging
- See [Linkerd HTTP Access Logging](#linkerd-http-access-logging) section for details

#### Topology Aware Routing

Topology Aware Routing enables Linkerd to prefer routing traffic to endpoints in the same zone/region, reducing cross-zone network costs and improving latency. This is particularly valuable in multi-AZ Kubernetes clusters.

**Benefits**:
- **Cost Reduction**: Minimize cross-zone data transfer costs (AWS, GCP, Azure)
- **Lower Latency**: Prefer same-zone endpoints for faster response times
- **Bandwidth Optimization**: Reduce inter-zone bandwidth usage
- **Automatic Failover**: Fallback to other zones if same-zone endpoints unavailable

**How It Works**:

1. Linkerd destination controller discovers endpoint topology (zone labels)
2. When selecting endpoints, Linkerd prefers endpoints in the same zone as the client
3. If same-zone endpoints unavailable, falls back to other zones
4. Works automatically for all meshed services - no code changes required

**Configuration**:

Topology Aware Routing is enabled via Linkerd configuration:

```yaml
# Linkerd configuration
apiVersion: v1
kind: ConfigMap
metadata:
  name: linkerd-config
  namespace: linkerd
data:
  config.yaml: |
    # Enable topology aware routing
    enableTopologyAwareRouting: true
    
    # Zone label (default: topology.kubernetes.io/zone)
    # Can be customized for different cloud providers
    zoneLabel: "topology.kubernetes.io/zone"
```

**Zone Labels**:

Kubernetes nodes must have zone labels for topology aware routing to work:

```yaml
# Node labels (automatically set by cloud providers)
apiVersion: v1
kind: Node
metadata:
  name: worker-node-1
  labels:
    topology.kubernetes.io/zone: "us-east-1a"  # AWS
    # topology.kubernetes.io/region: "us-east-1"
    # failure-domain.beta.kubernetes.io/zone: "us-east-1a"  # Legacy
```

**Verification**:

```bash
# Check if topology aware routing is enabled
linkerd check --proxy

# View endpoint distribution by zone
linkerd viz stat deploy -n knative-lambda --all-namespaces

# Check endpoint selection in logs
linkerd viz tap deploy/operator -n knative-lambda | grep zone
```

**Use Cases for Knative Lambda**:

1. **Lambda Function Invocations**:
   - Prefer same-zone Lambda function pods for faster execution
   - Reduce cross-zone latency for synchronous invocations
   - Lower data transfer costs for high-throughput workloads

2. **Operator API Calls**:
   - Operator prefers same-zone Knative Service endpoints
   - Faster reconciliation when operator and services in same zone
   - Reduced cross-zone API call latency

3. **AI Agent Queries**:
   - AI Agent prefers same-zone Prometheus/Tempo endpoints
   - Faster alert investigation and metric queries
   - Lower latency for real-time observability queries

4. **Notifi Service Integration**:
   - Lambda functions prefer same-zone Notifi service pods
   - Faster gRPC calls to storage-manager, event-processor
   - Reduced cross-zone gRPC latency

**Example: Lambda Function → Notifi Service**

```yaml
# Lambda function pod in us-east-1a
# Notifi service has pods in us-east-1a, us-east-1b, us-east-1c

# Without Topology Aware Routing:
# - 33% chance of routing to us-east-1a (same zone)
# - 33% chance of routing to us-east-1b (cross-zone)
# - 33% chance of routing to us-east-1c (cross-zone)

# With Topology Aware Routing:
# - 100% preference for us-east-1a (same zone)
# - Fallback to us-east-1b/1c only if us-east-1a unavailable
```

**Monitoring**:

Track topology aware routing effectiveness:

```promql
# Cross-zone vs same-zone request ratio
sum(rate(linkerd_response_total{zone_match="false"}[5m])) by (deployment) /
sum(rate(linkerd_response_total[5m])) by (deployment)

# Same-zone request percentage (should be high)
sum(rate(linkerd_response_total{zone_match="true"}[5m])) by (deployment) /
sum(rate(linkerd_response_total[5m])) by (deployment) * 100

# Latency difference: same-zone vs cross-zone
histogram_quantile(0.95,
  sum(rate(linkerd_response_latency_ms_bucket{zone_match="true"}[5m])) by (le)
) -
histogram_quantile(0.95,
  sum(rate(linkerd_response_latency_ms_bucket{zone_match="false"}[5m])) by (le)
)
```

**Reference**: [Linkerd Topology Aware Routing Documentation](https://linkerd.io/2-edge/features/topology-aware-routing/)

### API Gateway Integration

**External API Exposure**:
- Ingress controller (NGINX/Ambassador)
- TLS termination
- Rate limiting
- Authentication (OAuth2/JWT)

**Rate Limiting**:
- 100 requests/second per client
- Burst: 200 requests
- Exponential backoff

**Authentication**:
- Service Account tokens
- OAuth2 for external clients
- JWT validation

### GitOps Integration

**ArgoCD/Flux Integration**:
- CRD sync from Git repository
- Automatic deployment on changes
- Rollback on failure
- Multi-environment support

**CRD Sync**:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: lambda-functions
spec:
  source:
    repoURL: https://github.com/org/lambda-functions
    path: functions/
    targetRevision: main
  destination:
    server: https://kubernetes.default.svc
    namespace: knative-lambda
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

## 📝 API Versioning & Compatibility

### API Versioning Strategy

**Version Scheme**: Semantic versioning (v1, v2, etc.)

**Version Endpoints**:
- `/api/v1/lambda/functions` - v1 API
- `/api/v2/lambda/functions` - v2 API (future)

**Backward Compatibility**:
- Maintain v1 API for 12 months after v2 release
- Deprecation warnings in responses
- Migration guides provided

**Deprecation Policy**:
1. Announce deprecation (3 months notice)
2. Mark as deprecated in documentation
3. Remove deprecated endpoints (after 12 months)

### CRD Versioning

**CRD Version Management**:
- Current version: `v1alpha1`
- Storage version: `v1alpha1`
- Served versions: `v1alpha1`, `v1beta1` (future)

**Migration Strategies**:
- Conversion webhooks for version upgrades
- Automatic migration on CRD update
- Validation webhooks for schema enforcement

**Conversion Webhooks**:
```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  versions:
  - name: v1alpha1
    served: true
    storage: true
  - name: v1beta1
    served: true
    storage: false
    schema:
      # v1beta1 schema
  conversion:
    strategy: Webhook
    webhook:
      clientConfig:
        service:
          name: operator-webhook
          namespace: knative-lambda
          path: /convert
```

## 📊 Multi-Tenancy Details

### Namespace Isolation

**Namespace Strategy**:
- One namespace per tenant
- Resource quotas per namespace
- Network policies per namespace
- RBAC per namespace

**Resource Quotas**:
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: tenant-quota
  namespace: tenant-1
spec:
  hard:
    requests.cpu: "10"
    requests.memory: "20Gi"
    limits.cpu: "20"
    limits.memory: "40Gi"
    lambdafunctions.lambda.knative.io: "100"
    persistentvolumeclaims: "10"
```

**Network Policies per Namespace**:
- Isolate tenant namespaces
- Allow only necessary cross-namespace communication
- Block external access except via ingress

**RBAC per Namespace**:
- Tenant-specific ServiceAccounts
- RoleBindings per namespace
- Least privilege access

### Tenant Management

**Tenant Onboarding**:
1. Create namespace
2. Apply resource quotas
3. Configure network policies
4. Create ServiceAccount
5. Set up RBAC
6. Configure monitoring

**Resource Allocation**:
- Fair share allocation
- Burst capacity for spikes
- Priority classes for important tenants

**Cost Tracking per Tenant**:
- Cost allocation tags per tenant
- Per-tenant cost dashboards
- Budget alerts per tenant

## 🎯 Rate Limiting & Throttling

### API Rate Limiting

**Rate Limit Configuration**:
```yaml
rateLimiting:
  enabled: true
  requestsPerSecond: 100
  burstSize: 200
  perClient: true
  perEndpoint: true
```

**Throttling Strategies**:
- Token bucket algorithm
- Sliding window rate limiting
- Per-client rate limits
- Per-endpoint rate limits

**Rate Limit Headers**:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642684800
Retry-After: 5
```

### Reconciliation Rate Limiting

**Max Reconciliations per Second**:
- Default: 10 reconciliations/second
- Configurable per operator instance
- Queue-based throttling

**Backoff Strategies**:
- Exponential backoff on errors
- Jitter to prevent thundering herd
- Maximum backoff: 5 minutes

**Queue Management**:
- Priority queue for critical functions
- FIFO queue for normal functions
- Queue depth monitoring
- Alert on queue depth > 1000

### Rate Limiting & Circuit Breakers: Comprehensive Strategy

This section outlines rate limiting and circuit breakers specifically for **Knative Lambda functions**, **AI Agent**, and **Notifi services** integration, leveraging both **RabbitMQ native features** and **Linkerd service mesh capabilities**.

#### Architecture Overview: Rate Limiting & Circuit Breakers

This section details rate limiting and circuit breakers specifically for **Knative Lambda functions**, **AI Agent**, and **Notifi services** integration, based on the complete Notifi Fusion Platform architecture.

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│              NOTIFI FUSION PLATFORM - COMPLETE RATE LIMITING ARCHITECTURE                           │
│              (Knative Lambda Operator + AI Agent + Notifi Services Integration)                    │
├─────────────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                                     │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────────┐   │
│  │                    NOTIFI SERVICES (Kubernetes Cluster - namespace: notifi)                │   │
│  ├─────────────────────────────────────────────────────────────────────────────────────────────┤   │
│  │                                                                                             │   │
│  │  ┌─────────────────────────────────────────────────────────────────────────────────────┐   │   │
│  │  │                    CORE ORCHESTRATION SERVICES                                      │   │   │
│  │  ├─────────────────────────────────────────────────────────────────────────────────────┤   │   │
│  │  │                                                                                     │   │   │
│  │  │  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐     │   │   │
│  │  │  │  Scheduler   │◄───│ Event        │    │   Monitor    │    │ Management   │     │   │   │
│  │  │  │  Service     │    │ Processor    │    │   Service    │    │ Gateway      │     │   │   │
│  │  │  │              │    │              │    │              │    │              │     │   │   │
│  │  │  │ HTTP: 5000   │    │ gRPC: 4030   │    │ gRPC: 4055   │    │ HTTP: 5001   │     │   │   │
│  │  │  │ gRPC: 4000   │    │              │    │              │    │ gRPC: 4001   │     │   │   │
│  │  │  │ Health: 6000 │    │              │    │              │    │              │     │   │   │
│  │  │  │ Metrics: 7000│    │              │    │              │    │              │     │   │   │
│  │  │  │              │    │              │    │              │    │              │     │   │   │
│  │  │  │ Endpoints:   │    │              │    │              │    │              │     │   │   │
│  │  │  │ /fusion/     │    │              │    │              │    │              │     │   │   │
│  │  │  │ execution/   │    │              │    │              │    │              │     │   │   │
│  │  │  │ response     │    │              │    │              │    │              │     │   │   │
│  │  │  │              │    │              │    │              │    │              │     │   │   │
│  │  │  │ Rate: 100/s  │    │              │    │              │    │              │     │   │   │
│  │  │  │ Per pod      │    │              │    │              │    │              │     │   │   │
│  │  │  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘    └──────┬───────┘     │   │   │
│  │  │         │                   │                    │                    │             │   │   │
│  │  │         │                   │                    │                    │             │   │   │
│  │  │         │                   └────────────────────┴────────────────────┘             │   │   │
│  │  │         │                                    │                                        │   │   │
│  │  │         │         ┌──────────────────────────▼──────────────────────────┐           │   │   │
│  │  │         │         │         RabbitMQ (CloudEvents Broker)                │           │   │   │
│  │  │         │         │  ┌──────────────────────────────────────────────┐   │           │   │   │
│  │  │         │         │  │ Exchange: "cloud-events"                    │   │           │   │   │
│  │  │         │         │  │ Routing Keys:                                │   │           │   │   │
│  │  │         │         │  │ - network.notifi.lambda.build.start         │   │           │   │   │
│  │  │         │         │  │ - network.notifi.lambda.parser.start        │   │           │   │   │
│  │  │         │         │  │                                              │   │           │   │   │
│  │  │         │         │  │ Rate Limit: 1000 msg/s per consumer          │   │           │   │   │
│  │  │         │         │  │ Prefetch: 10 messages (circuit breaker)      │   │           │   │   │
│  │  │         │         │  │ TTL: 7 days                                  │   │           │   │   │
│  │  │         │         │  │ Max Queue Size: 50k messages                 │   │           │   │   │
│  │  │         │         │  └──────────────────────────────────────────────┘   │           │   │   │
│  │  │         │         └──────────────────────────┬──────────────────────────┘           │   │   │
│  │  │         │                                    │                                        │   │   │
│  │  │         │                                    ▼                                        │   │   │
│  │  │         │         ┌──────────────────────────────────────────────┐                  │   │   │
│  │  │         │         │   Knative Lambda Operator (Consumer)          │                  │   │   │
│  │  │         │         │   Rate Limit: 1000 msg/s                      │                  │   │   │
│  │  │         │         │   Circuit Breaker: Prefetch 10                │                  │   │   │
│  │  │         │         └──────────────────────────────────────────────┘                  │   │   │
│  │  │         │                                    │                                        │   │   │
│  │  │         │                                    │                                        │   │   │
│  │  │         ▼                                    ▼                                        │   │   │
│  │  │  ┌──────────────┐                    ┌──────────────────────────────────┐            │   │   │
│  │  │  │  RateLimit   │                    │  Knative Lambda Functions        │            │   │   │
│  │  │  │  Broker      │                    │  (Built & Deployed by Operator)   │            │   │   │
│  │  │  │              │                    │                                  │            │   │   │
│  │  │  │ gRPC: 4060   │                    │  ┌────────────────────────────┐ │            │   │   │
│  │  │  │              │                    │  │ Parser Functions           │ │            │   │   │
│  │  │  │ Purpose:     │                    │  │ Filter Functions            │ │            │   │   │
│  │  │  │ Capacity     │                    │  │ Alert Functions            │ │            │   │   │
│  │  │  │ Management   │                    │  │                            │ │            │   │   │
│  │  │  └──────────────┘                    │  │ Rate Limits (per function):│ │            │   │   │
│  │  │                                      │  │ → Scheduler: 10 req/s      │ │            │   │   │
│  │  │  ┌──────────────┐                    │  │   (HTTP: 5000)            │ │            │   │   │
│  │  │  │  Blockchain  │                    │  │ → Storage: 10 req/s        │ │            │   │   │
│  │  │  │   Manager    │                    │  │   (gRPC: 4000)             │ │            │   │   │
│  │  │  │              │                    │  │ → Subscription: 10 req/s   │ │            │   │   │
│  │  │  │ gRPC: 4000   │◄───────────────────┤  │   (gRPC: 4000)            │ │            │   │   │
│  │  │  │ Health: 6000 │                    │  │ → Fetch Proxy: 10 req/s    │ │            │   │   │
│  │  │  │ Metrics: 7000│                    │  │   (gRPC: 4000)             │ │            │   │   │
│  │  │  │              │                    │  │ → Blockchain: 10 req/s      │ │            │   │   │
│  │  │  │ Rate: 100/s  │                    │  │   (gRPC: 4000)             │ │            │   │   │
│  │  │  │ Per pod      │                    │  │                            │ │            │   │   │
│  │  │  │              │                    │  │ Circuit Breaker:            │ │            │   │   │
│  │  │  │ Supports:    │                    │  │ - Opens: 3 failures OR 50%  │ │            │   │   │
│  │  │  │ - EVM        │                    │  │ - Half-open: 30s           │ │            │   │   │
│  │  │  │ - Solana     │                    │  │ - Closes: 1 success        │ │            │   │   │
│  │  │  │ - Sui        │                    │  └────────────────────────────┘ │            │   │   │
│  │  │  │ - Cosmos     │                    │                                  │            │   │   │
│  │  │  └──────────────┘                    │  ┌────────────────────────────┐ │            │   │   │
│  │  │                                      │  │ AI Agent (Knative Service)  │ │            │   │   │
│  │  │  ┌──────────────┐                    │  │                            │ │            │   │   │
│  │  │  │   Storage    │                    │  │ Rate Limits (higher quota): │ │            │   │   │
│  │  │  │   Manager    │                    │  │ → Scheduler: 20 req/s       │ │            │   │   │
│  │  │  │              │                    │  │   (HTTP: 5000)             │ │            │   │   │
│  │  │  │ gRPC: 4000   │◄───────────────────┤  │ → Storage: 20 req/s        │ │            │   │   │
│  │  │  │ Health: 6000 │                    │  │   (gRPC: 4000)             │ │            │   │   │
│  │  │  │ Metrics: 7000│                    │  │ → Subscription: 20 req/s   │ │            │   │   │
│  │  │  │              │                    │  │   (gRPC: 4000)             │ │            │   │   │
│  │  │  │ Rate: 100/s  │                    │  │ → Fetch Proxy: 20 req/s     │ │            │   │   │
│  │  │  │ Per pod      │                    │  │   (gRPC: 4000)             │ │            │   │   │
│  │  │  │              │                    │  │ → Blockchain: 20 req/s     │ │            │   │   │
│  │  │  │ Storage:     │                    │  │   (gRPC: 4000)              │ │            │   │   │
│  │  │  │ - MinIO      │                    │  │                            │ │            │   │   │
│  │  │  │ - S3/GCS     │                    │  │ Circuit Breaker: Same as     │ │            │   │   │
│  │  │  │ (optional)   │                    │  │ Lambda functions           │ │            │   │   │
│  │  │  └──────────────┘                    │  └────────────────────────────┘ │            │   │   │
│  │  │                                      │                                  │            │   │   │
│  │  │  ┌──────────────┐                    │  ┌────────────────────────────┐ │            │   │   │
│  │  │  │ Subscription │                    │  │ AI Agent → Storage         │ │            │   │   │
│  │  │  │   Manager    │                    │  │ → MinIO (default)          │ │            │   │   │
│  │  │  │              │                    │  │ → AWS S3 (optional)        │ │            │   │   │
│  │  │  │ gRPC: 4000   │◄───────────────────┤  │ → GCS (optional)           │ │            │   │   │
│  │  │  │ Health: 6000 │                    │  │ → LanceDB (vector DB)      │ │            │   │   │
│  │  │  │ Metrics: 7000│                    │  │ Retry: 3 attempts          │ │            │   │   │
│  │  │  │              │                    │  │ Backoff: 1s, 2s, 4s        │ │            │   │   │
│  │  │  │ Rate: 100/s  │                    │  └────────────────────────────┘ │            │   │   │
│  │  │  │ Per pod      │                    │                                  │            │   │   │
│  │  │  └──────────────┘                    └──────────────────────────────────┘            │   │   │
│  │  │                                                                                     │   │   │
│  │  │  ┌──────────────┐                                                                  │   │   │
│  │  │  │  Fetch Proxy │                                                                  │   │   │
│  │  │  │              │                                                                  │   │   │
│  │  │  │ gRPC: 4000   │◄───────────────────────────────────────────────────────────────  │   │   │
│  │  │  │ Health: 6000 │                                                                  │   │   │
│  │  │  │ Metrics: 7000│                                                                  │   │   │
│  │  │  │              │                                                                  │   │   │
│  │  │  │ Rate: 100/s  │                                                                  │   │   │
│  │  │  │ Per pod      │                                                                  │   │   │
│  │  │  │              │                                                                  │   │   │
│  │  │  │ Purpose:     │                                                                  │   │   │
│  │  │  │ External API │                                                                  │   │   │
│  │  │  │ calls         │                                                                  │   │   │
│  │  │  └──────────────┘                                                                  │   │   │
│  │  │                                                                                     │   │   │
│  │  └─────────────────────────────────────────────────────────────────────────────────────┘   │   │
│  │                                                                                             │   │
│  └─────────────────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                                     │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────────┐   │
│  │                    KUBERNETES API (Operator → K8s API)                                       │   │
│  │                    Rate Limit: 400 req/s, Batch operations                                   │   │
│  └─────────────────────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘

KEY INTERACTIONS & RATE LIMITS:
────────────────────────────────
1. Scheduler → RabbitMQ → Operator:
   - Exchange: "cloud-events"
   - Routing Keys: network.notifi.lambda.build.start, network.notifi.lambda.parser.start
   - Format: application/cloudevents+json
   - Rate Limit: 1000 msg/s per consumer (RabbitMQ)
   - Circuit Breaker: Prefetch 10 messages
   - TTL: 7 days, Max Queue Size: 50k messages

2. Scheduler → RateLimitBroker → EventProcessor:
   - Scheduler requests capacity from RateLimitBroker (gRPC:4060) before queuing events
   - If capacity granted, Scheduler queues events to EventProcessor (gRPC:4030)
   - If capacity denied, request fails with ResourceExhausted
   - Capacity refunded if event queuing fails

3. Lambda Functions → Notifi Services (gRPC/HTTP):
   - Scheduler (HTTP:5000): 10 req/s per function, 100 req/s per pod
     Endpoint: POST /fusion/execution/response
   - Storage Manager (gRPC:4000): 10 req/s per function, 100 req/s per pod
   - Subscription Manager (gRPC:4000): 10 req/s per function, 100 req/s per pod
   - Fetch Proxy (gRPC:4000): 10 req/s per function, 100 req/s per pod
   - Blockchain Manager (gRPC:4000): 10 req/s per function, 100 req/s per pod
   - Circuit Breaker: Opens after 3 failures OR 50% failure rate over 10s

4. AI Agent → Notifi Services (gRPC/HTTP):
   - Same services as Lambda functions, but 20 req/s per AI Agent (higher quota)
   - Circuit Breaker: Same as Lambda functions

5. AI Agent → Storage (MinIO/S3/GCS/LanceDB):
   - Retry: 3 attempts with exponential backoff (1s, 2s, 4s)
   - No explicit rate limiting (limited by cluster resources)

6. Operator → K8s API:
   - Rate Limit: 400 req/s (client-side)
   - Batch operations: 10-50 operations per batch

NOTIFI SERVICE PORT STANDARD:
──────────────────────────────
All Notifi services follow the same port pattern:
- HTTP: 5000 (Http1) - REST API endpoints
- gRPC: 4000 (Http2) - gRPC service calls
- Health: 6000 (Http2) - Health checks (readiness/liveness)
- Metrics: 7000 (Http1) - Prometheus metrics
```

#### Critical Points for Rate Limiting

**1. Operator HTTP API** (Priority: HIGH)
- **Location**: Operator REST API endpoints
- **Implementation**: Linkerd Policy API (Server + ServerAuthorization) + HTTPRoute for retries/timeouts
- **Configuration**:
  ```yaml
  # Linkerd Policy API
  apiVersion: policy.linkerd.io/v1beta3
  kind: Server
  spec:
    podSelector:
      matchLabels:
        app: knative-lambda-operator
    port: 8080
    proxyProtocol: HTTP/1
  
  # HTTPRoute for retries and timeouts
  apiVersion: gateway.networking.k8s.io/v1
  kind: HTTPRoute
  spec:
    rules:
    - matches:
      - method: POST
        path:
          type: PathPrefix
          value: /api/v1/lambda/functions
      timeouts:
        request: 30s
      retries:
        attempts: 3
        perTryTimeout: 10s
  ```
- **Rate Limits**:
  - Create/Update: 10 req/s per client
  - List/Get: 50 req/s per client
  - Delete: 5 req/s per client
- **Rationale**: Prevents operator overload, protects against DDoS

**2. RabbitMQ Consumer Rate Limiting** (Priority: HIGH)
- **Location**: RabbitMQ consumer connections
- **Implementation**: RabbitMQ policies + consumer prefetch
- **Context**: The Operator consumes CloudEvents from RabbitMQ published by the **Notifi Scheduler Service**. The Scheduler publishes events to the `cloud-events` exchange with routing keys:
  - `network.notifi.lambda.build.start` - Triggers build process for Lambda functions
  - `network.notifi.lambda.parser.start` - Triggers parser execution
- **Scheduler Integration**: The Scheduler uses `IRabbitMqClient` to publish CloudEvents in `application/cloudevents+json` format with persistent delivery mode
- **Configuration**:
  ```yaml
  # RabbitMQ Policy
  apiVersion: rabbitmq.com/v1beta1
  kind: Policy
  metadata:
    name: knative-lambda-rate-limit
    namespace: rabbitmq-cluster-knative-lambda
  spec:
    name: rate-limit-policy
    pattern: "^knative-lambda"
    definition:
      max-rate: 1000  # messages per second per consumer
      consumer-timeout: 30000  # 30 seconds
      prefetch-count: 10  # Circuit breaker: max unacked messages
  ```
- **Rate Limits**:
  - Per consumer: 1000 msg/s
  - Per queue: 5000 msg/s
  - Prefetch: 10 messages (circuit breaker)
- **Rationale**: Prevents consumer overload, ensures fair distribution

**3. Lambda Functions → Notifi Services (gRPC/HTTP)** (Priority: HIGH)
- **Location**: Knative Lambda functions → Notifi services (gRPC/HTTP)
- **Notifi Services**: Scheduler (HTTP), Storage Manager (gRPC), Subscription Manager (gRPC), Fetch Proxy (gRPC), Blockchain Manager (gRPC)
- **Implementation**: Linkerd Policy API with `HTTPLocalRateLimitPolicy` for local rate limiting
- **Scope**: Local rate limiting (per-pod inbound proxy), not global across replicas
- **Algorithm**: Generic Cell Rate Algorithm (GCRA) - more performant than token bucket/leaky bucket
  - Cell rate: Derived from `requestsPerSecond` in policy
  - Tolerance: 1 second (accommodates small variations/bursts while maintaining long-term rate limits)
- **Service Ports** (Standard Notifi Service Pattern):
  - **HTTP**: Port 5000 (Http1) - REST API endpoints
  - **gRPC**: Port 4000 (Http2) - gRPC service calls
  - **Health**: Port 6000 (Http2) - Health checks (readiness/liveness)
  - **Metrics**: Port 7000 (Http1) - Prometheus metrics
- **Context**: Lambda functions call Notifi services to:
  - **Scheduler (HTTP:5000)**: Send execution response callbacks (`POST /fusion/execution/response`)
  - **Storage Manager (gRPC:4000)**: Retrieve parser modules and store execution results
  - **Subscription Manager (gRPC:4000)**: Trigger alert processing
  - **Fetch Proxy (gRPC:4000)**: Make external API calls
  - **Blockchain Manager (gRPC:4000)**: Query blockchain data (EVM, Solana, Sui, Cosmos)
- **Configuration**:
  ```yaml
  # Linkerd Server for Scheduler (HTTP:5000)
  apiVersion: policy.linkerd.io/v1beta3
  kind: Server
  metadata:
    name: scheduler-server
    namespace: notifi
  spec:
    podSelector:
      matchLabels:
        app: scheduler
    port: 5000  # HTTP port (standard Notifi service pattern)
    proxyProtocol: HTTP/1
  
  # Linkerd Server for Storage Manager (gRPC)
  apiVersion: policy.linkerd.io/v1beta3
  kind: Server
  metadata:
    name: storage-manager-server
    namespace: notifi
  spec:
    podSelector:
      matchLabels:
        app: storage-manager
    port: 4000
    proxyProtocol: HTTP/2  # gRPC uses HTTP/2
  
  # Linkerd ServerAuthorization for mTLS
  apiVersion: policy.linkerd.io/v1beta1
  kind: ServerAuthorization
  metadata:
    name: storage-manager-auth
    namespace: notifi
  spec:
    server:
      name: storage-manager-server
    client:
      meshTLS:
        identities:
        - "*.knative-lambda.svc.cluster.local"
  
  # Linkerd HTTPLocalRateLimitPolicy for rate limiting
  apiVersion: policy.linkerd.io/v1beta3
  kind: HTTPLocalRateLimitPolicy
  metadata:
    name: storage-manager-rate-limit
    namespace: notifi
  spec:
    targetRef:
      group: policy.linkerd.io
      kind: Server
      name: storage-manager-server
    conditions:
      # Global rate limit for all inbound traffic
      - requestsPerSecond: 100
      # Per-identity rate limits (fairness - prevents single client from consuming all quota)
      - requestsPerSecond: 10
        client:
          meshTLS:
            identities:
            - "*.knative-lambda.svc.cluster.local"
      # Override for specific high-priority clients
      - requestsPerSecond: 50
        client:
          meshTLS:
            identities:
            - "admin.knative-lambda.svc.cluster.local"
  ```
- **Rate Limits**:
  - Global: 100 req/s per pod (all inbound traffic across all services)
  - Per Lambda function identity: 10 req/s per service (fairness)
  - Per admin identity: 50 req/s (override)
  - Unmeshed sources: Treated as single source
- **Fairness**: Per-identity limits prevent specific clients from consuming all quota
- **Service Addresses** (Kubernetes DNS):
  - Scheduler: `notifi-scheduler.notifi.svc.cluster.local:5000` (HTTP), `:4000` (gRPC)
  - Storage Manager: `notifi-storage-manager.notifi.svc.cluster.local:4000` (gRPC)
  - Subscription Manager: `notifi-subscription-manager.notifi.svc.cluster.local:4000` (gRPC)
  - Fetch Proxy: `notifi-fetch-proxy.notifi.svc.cluster.local:4000` (gRPC)
  - Blockchain Manager: `notifi-blockchain-manager.notifi.svc.cluster.local:4000` (gRPC)
- **Rationale**: Protects Notifi services from overload, ensures fair resource distribution, respects quotas

**4. AI Agent → Notifi Services (gRPC/HTTP)** (Priority: HIGH)
- **Location**: AI Agent (Knative Service) → Notifi services (gRPC/HTTP)
- **Notifi Services**: Scheduler (HTTP), Storage Manager (gRPC), Subscription Manager (gRPC), Fetch Proxy (gRPC), Blockchain Manager (gRPC)
- **Implementation**: Linkerd Policy API with `HTTPLocalRateLimitPolicy` for local rate limiting
- **Scope**: Local rate limiting (per-pod inbound proxy), not global across replicas
- **Algorithm**: Generic Cell Rate Algorithm (GCRA) - same as Lambda functions
- **Service Ports** (Standard Notifi Service Pattern):
  - **HTTP**: Port 5000 (Http1) - REST API endpoints
  - **gRPC**: Port 4000 (Http2) - gRPC service calls
  - **Health**: Port 6000 (Http2) - Health checks
  - **Metrics**: Port 7000 (Http1) - Prometheus metrics
- **Context**: AI Agent calls Notifi services for:
  - **Scheduler (HTTP:5000)**: Event investigation and execution status queries
  - **Storage Manager (gRPC:4000)**: Retrieve event data and execution logs
  - **Subscription Manager (gRPC:4000)**: Alert processing and user subscription data
  - **Fetch Proxy (gRPC:4000)**: External API calls during investigations
  - **Blockchain Manager (gRPC:4000)**: Blockchain data queries for context
- **Configuration**:
  ```yaml
  # Linkerd HTTPLocalRateLimitPolicy for AI Agent
  apiVersion: policy.linkerd.io/v1beta3
  kind: HTTPLocalRateLimitPolicy
  metadata:
    name: storage-manager-ai-agent-rate-limit
    namespace: notifi
  spec:
    targetRef:
      group: policy.linkerd.io
      kind: Server
      name: storage-manager-server
    conditions:
      # Per-identity rate limit for AI Agent (higher quota than Lambda functions)
      - requestsPerSecond: 20
        client:
          meshTLS:
            identities:
            - "knative-lambda-ai-agent.knative-lambda.svc.cluster.local"
  ```
- **Rate Limits**:
  - Per AI Agent identity: 20 req/s per service (higher quota than Lambda functions for investigation workloads)
  - Global: 100 req/s per pod (shared with Lambda functions)
- **Service Addresses** (Kubernetes DNS):
  - Same as Lambda functions (see section 3)
- **Rationale**: AI Agent needs higher quota for alert investigations and event analysis, but still respects global limits

**5. Scheduler → RateLimitBroker (gRPC)** (Priority: MEDIUM)
- **Location**: Scheduler → RateLimitBroker service
- **Service**: RateLimitBroker (gRPC:4060)
- **Purpose**: Capacity management for event queuing to prevent EventProcessor overload
- **Implementation**: Scheduler calls RateLimitBroker before queuing events to EventProcessor
- **Context**: The Scheduler uses `IRateLimitBrokerClient` to request capacity before queuing parsed events. If capacity is not granted, the request fails with `ResourceExhausted` status. Capacity is refunded if event queuing fails.
- **Configuration**:
  ```yaml
  # Linkerd Server for RateLimitBroker
  apiVersion: policy.linkerd.io/v1beta3
  kind: Server
  metadata:
    name: rate-limit-broker-server
    namespace: notifi
  spec:
    podSelector:
      matchLabels:
        app: rate-limit-broker
    port: 4060
    proxyProtocol: HTTP/2  # gRPC uses HTTP/2
  ```
- **Rate Limits**: Managed by RateLimitBroker service (internal quota management)
- **Service Address**: `notifi-rate-limit-broker.notifi.svc.cluster.local:4060` (gRPC)
- **Rationale**: Prevents EventProcessor from being overwhelmed by too many events, ensures fair capacity distribution

**5.1. Scheduler → EventProcessor & Monitor (gRPC)** (Priority: MEDIUM)
- **Location**: Scheduler → EventProcessor and Monitor services
- **Services**: 
  - EventProcessor (gRPC:4030) - Processes parsed events from Lambda functions
  - Monitor (gRPC:4055) - Monitors event processing and handles failures
- **Purpose**: Event processing pipeline and monitoring
- **Context**: 
  - Scheduler calls EventProcessor to queue parsed events after capacity is granted by RateLimitBroker
  - Monitor is used to mark events as failed and activate cursors on internal errors
  - These services are part of the core orchestration but not directly called by Lambda functions
- **Service Addresses**:
  - EventProcessor: `notifi-event-processor.notifi.svc.cluster.local:4030` (gRPC)
  - Monitor: `notifi-monitor.notifi.svc.cluster.local:4055` (gRPC)
- **Rationale**: These are internal Scheduler dependencies, not directly exposed to Lambda functions

**6. AI Agent → Storage (MinIO/S3/GCS/LanceDB)** (Priority: MEDIUM)
- **Location**: AI Agent → Object storage (MinIO default, S3/GCS optional) and LanceDB vector database
- **Implementation**: Application-level retry with exponential backoff
- **Context**: AI Agent stores event embeddings, training data, and vector search results
- **Configuration**:
  ```python
  # MinIO client (default) with retry
  from minio import Minio
  from minio.error import S3Error
  
  minio_client = Minio(
      "minio.minio.svc.cluster.local:9000",
      access_key="minioadmin",
      secret_key="minioadmin",
      secure=False
  )
  
  # For AWS S3 (optional)
  s3_client = boto3.client('s3', config=Config(
      retries={'max_attempts': 3, 'mode': 'adaptive'}
  ))
  
  # For GCS (optional)
  from google.cloud import storage
  gcs_client = storage.Client()
  
  # LanceDB client with retry
  import lancedb
  db = lancedb.connect("/tmp/lancedb")
  ```
- **Rate Limits**:
  - MinIO: No explicit rate limit (limited by cluster resources)
  - AWS S3: 5,500 PUT/COPY/POST/DELETE requests/second per prefix
  - GCS: 10,000 requests/minute per bucket
  - LanceDB: No explicit rate limit (local/embedded database)
- **Retry Logic**:
  - Max retries: 3 attempts
  - Exponential backoff: 1s, 2s, 4s
  - After 3 failures → log error, continue processing (non-blocking)
- **Rationale**: Handles transient storage issues, doesn't block event processing or alert investigations
- **Storage Selection**: MinIO (default), AWS S3 or GCS (optional)

**7. Kubernetes API Calls** (Priority: MEDIUM)
- **Location**: Operator → Kubernetes API
- **Implementation**: Client-side rate limiting in operator code
- **Configuration**:
  ```go
  // Client-side rate limiter
  rateLimiter := rate.NewLimiter(rate.Limit(400), 50) // 400 req/s, burst 50
  
  // Batch operations
  batchSize := 10
  reconcileBatch(funcs []LambdaFunction) {
    // Process in batches to respect rate limits
  }
  ```
- **Rate Limits**:
  - Default: 400 req/s (Kubernetes default)
  - Batch operations: 10-50 operations per batch
- **Rationale**: Prevents API server throttling, ensures fair resource usage

**5. Container Registry Push Operations** (Priority: MEDIUM)
- **Location**: Kaniko jobs → Container registries
- **Implementation**: Exponential backoff + retry logic
- **Configuration**:
  ```yaml
  # Kaniko job with retry
  spec:
    backoffLimit: 3
    template:
      spec:
        containers:
        - name: kaniko
          env:
          - name: REGISTRY_RETRY_ATTEMPTS
            value: "3"
          - name: REGISTRY_RETRY_DELAY
            value: "5s"
  ```
- **Rate Limits**:
  - ECR: 500 req/s
  - GHCR: 5000 req/hour
  - GCR: 10000 req/minute
- **Rationale**: Respects registry rate limits, prevents build failures

#### Critical Points for Circuit Breakers

**1. Lambda Functions & AI Agent → Notifi Services (gRPC/HTTP)** (Priority: CRITICAL)
- **Location**: Knative Lambda functions and AI Agent → Notifi services (gRPC/HTTP)
- **Notifi Services**: Scheduler (HTTP:5000), Storage Manager (gRPC:4000), Subscription Manager (gRPC:4000), Fetch Proxy (gRPC:4000), Blockchain Manager (gRPC:4000)
- **Implementation**: Linkerd automatic circuit breaking via Policy API
- **Context**: Both Lambda functions and AI Agent call Notifi services via gRPC/HTTP. Circuit breakers protect Notifi services from overload when services are experiencing failures.
- **Note**: The Scheduler also publishes CloudEvents to the Operator via RabbitMQ exchange `cloud-events` with routing keys `network.notifi.lambda.build.start` and `network.notifi.lambda.parser.start` (covered in RabbitMQ rate limiting section). This section covers Lambda functions and AI Agent calling back to Scheduler.
- **Configuration**:
  ```yaml
  # Linkerd Policy API with connection limits (circuit breaker)
  apiVersion: policy.linkerd.io/v1beta3
  kind: Server
  spec:
    podSelector:
      matchLabels:
        app: storage-manager
    port: 4000
    proxyProtocol: HTTP/2
    clientConnectionPolicy:
      maxConnections: 100
      maxPendingRequests: 50
    # Linkerd automatically opens circuit after:
    # - 3 consecutive failures
    # - 50% failure rate over 10s window
    # - Connection limits exceeded
  ```
- **Circuit Breaker Logic**:
  - Open after: 3 consecutive failures OR 50% failure rate
  - Half-open after: 30s
  - Close after: 1 successful request
- **Rationale**: Prevents cascading failures, protects Notifi services

**2. RabbitMQ Consumer Circuit Breaker** (Priority: HIGH)
- **Location**: RabbitMQ consumers (via prefetch)
- **Implementation**: RabbitMQ consumer prefetch + DLQ
- **Configuration**:
  ```yaml
  # RabbitMQ consumer prefetch acts as circuit breaker
  # If consumer can't process messages, prefetch buffer fills
  # Once prefetch is full, RabbitMQ stops delivering to that consumer
  # Failed messages go to DLQ after retries exhausted
  ```
- **Circuit Breaker Logic**:
  - Prefetch: 10 messages (max unacked)
  - If all 10 messages are unacked → consumer paused
  - After 5 retries → message to DLQ
  - Consumer resumes when messages acked
- **Rationale**: Prevents consumer overload, ensures message delivery

**3. Container Registry Operations** (Priority: MEDIUM)
- **Location**: Kaniko jobs → Container registries
- **Implementation**: Job-level retry + exponential backoff
- **Configuration**:
  ```yaml
  # Kubernetes Job with circuit breaker pattern
  spec:
    backoffLimit: 3  # Max 3 retries
    activeDeadlineSeconds: 1800  # 30 min timeout
    template:
      spec:
        restartPolicy: Never
  ```
- **Circuit Breaker Logic**:
  - Max retries: 3
  - Timeout: 30 minutes
  - After 3 failures → job marked as failed
  - Event sent to DLQ for manual intervention
- **Rationale**: Prevents infinite retry loops, fails fast on persistent issues

**4. Kubernetes API Operations** (Priority: MEDIUM)
- **Location**: Operator → Kubernetes API
- **Implementation**: Client-side circuit breaker pattern
- **Configuration**:
  ```go
  // Circuit breaker for K8s API
  cb := circuit.NewBreaker(
    circuit.WithMaxRequests(100),
    circuit.WithTimeout(30*time.Second),
    circuit.WithReadyToTrip(func(counts circuit.Counts) bool {
      return counts.ConsecutiveFailures >= 5
    }),
  )
  ```
- **Circuit Breaker Logic**:
  - Open after: 5 consecutive failures
  - Half-open after: 60s
  - Close after: 1 successful request
- **Rationale**: Prevents API server overload, handles transient failures

#### RabbitMQ + Linkerd Integration Strategy

**RabbitMQ Native Rate Limiting**:
```yaml
# RabbitMQ Policy for rate limiting
apiVersion: rabbitmq.com/v1beta1
kind: Policy
metadata:
  name: knative-lambda-rate-limit
spec:
  name: rate-limit
  pattern: "^knative-lambda"
  definition:
    max-rate: 1000  # messages/second per consumer
    consumer-timeout: 30000  # 30s timeout
    prefetch-count: 10  # Circuit breaker: max unacked
    message-ttl: 604800000  # 7 days
    max-length: 50000  # Max queue size
```

**Linkerd Policy API for RabbitMQ Consumers**:
```yaml
# Linkerd can't directly mesh RabbitMQ (AMQP protocol)
# But can mesh HTTP/gRPC services that consume from RabbitMQ
apiVersion: policy.linkerd.io/v1beta3
kind: Server
metadata:
  name: operator-server
  namespace: knative-lambda
spec:
  podSelector:
    matchLabels:
      app: knative-lambda-operator
  port: 8080
  proxyProtocol: HTTP/1
  clientConnectionPolicy:
    maxConnections: 100
    maxPendingRequests: 50
```

**Hybrid Approach**:
1. **RabbitMQ**: Handles message-level rate limiting and circuit breaking via prefetch
2. **Linkerd**: Handles service-to-service (HTTP/gRPC) rate limiting and circuit breaking
3. **Application**: Handles business logic retries and backoff

#### Implementation Priority

**Phase 1 (Critical - Implement First)**:
1. ✅ RabbitMQ consumer prefetch (circuit breaker)
2. ✅ Notifi service gRPC circuit breakers (Linkerd)
3. ✅ Operator API rate limiting (Linkerd Policy API)

**Phase 2 (High Priority)**:
4. ✅ RabbitMQ rate limiting policies
5. ✅ Kubernetes API client-side rate limiting
6. ✅ Container registry retry logic

**Phase 3 (Medium Priority)**:
7. ✅ Storage operation retries
8. ✅ Enhanced observability for rate limits/circuit breakers

#### Observability for Rate Limiting & Circuit Breakers

**Linkerd Metrics** (Automatic):
- `request_total`: Total requests per route
- `response_latency_ms`: P50, P95, P99 latencies
- `response_total`: Success/failure counts
- `retry_total`: Retry attempts and successes
- Circuit breaker state (open/closed/half-open)

**RabbitMQ Metrics**:
- `rabbitmq_queue_messages`: Queue depth
- `rabbitmq_queue_consumers`: Active consumers
- `rabbitmq_queue_message_rate`: Messages/second
- `rabbitmq_consumer_prefetch_count`: Unacked messages (circuit breaker indicator)

**Custom Metrics** (Prometheus):
```go
// Rate limiting metrics
rateLimitHits := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "operator_rate_limit_hits_total",
        Help: "Total rate limit hits",
    },
    []string{"endpoint", "client"},
)

// Circuit breaker metrics
circuitBreakerState := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "operator_circuit_breaker_state",
        Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
    },
    []string{"service", "route"},
)
```

**Alerting Rules**:
```yaml
# Prometheus alerting rules
groups:
- name: rate_limiting
  rules:
  - alert: HighRateLimitHits
    expr: rate(operator_rate_limit_hits_total[5m]) > 10
    annotations:
      summary: "High rate limit hits detected"
  
  - alert: CircuitBreakerOpen
    expr: operator_circuit_breaker_state == 1
    annotations:
      summary: "Circuit breaker opened for {{ $labels.service }}"
  
  - alert: RabbitMQConsumerBacklog
    expr: rabbitmq_queue_messages > 1000
    annotations:
      summary: "RabbitMQ queue backlog high: {{ $value }} messages"
```

**Linkerd Dashboard**:
- Real-time service topology showing circuit breaker states
- Request success rates per route
- Latency percentiles
- Retry budgets and usage

**RabbitMQ Management UI**:
- Queue depth monitoring
- Consumer connection status
- Message rate graphs
- Prefetch utilization (circuit breaker indicator)

---

## 🔍 Missing Topics & Recommendations (ML Engineer Review)

This section identifies topics that should be added to complete the operator architecture documentation, based on a comprehensive review of the knative-lambda documentation ecosystem.

### 📊 Observability & Metrics (HIGH PRIORITY)

**Status**: Mentioned but not detailed

**Missing Details**:
1. **Prometheus Metrics Endpoints**
   - Operator metrics endpoint (`:8080/metrics`)
   - Specific metrics exposed (counters, histograms, gauges)
   - Metric naming conventions
   - Label dimensions for metrics

2. **Key Metrics to Expose**
   - `operator_reconcile_total` - Total reconciliation attempts
   - `operator_reconcile_duration_seconds` - Reconciliation latency
   - `operator_crd_operations_total` - CRD create/update/delete counts
   - `operator_event_processing_duration_seconds` - CloudEvent processing time
   - `operator_build_jobs_active` - Active Kaniko jobs
   - `operator_knative_services_managed` - Number of managed services
   - `operator_dlq_depth` - DLQ queue depth
   - `operator_api_requests_total` - HTTP API request counts
   - `operator_api_request_duration_seconds` - API latency

3. **Distributed Tracing**
   - OpenTelemetry integration details (Go operator)
   - Pydantic Logfire integration (Python AI Agent)
   - Linkerd automatic proxy span generation
   - Trace context propagation across components (W3C Trace Context)
   - Span attributes for operator operations
   - Trace sampling configuration
   - Multi-backend architecture: Tempo (primary) + Logfire (AI Agent)

4. **Structured Logging**
   - Log format (JSON structured)
   - Correlation ID propagation
   - Log levels and when to use each
   - Log aggregation strategy

**Recommendation**: Add section "Observability & Monitoring" with:
- Metrics exposition configuration
- Example Prometheus queries
- Grafana dashboard recommendations
- Tracing setup and best practices

### 🚀 Performance & Scalability (HIGH PRIORITY)

**Status**: Not covered

**Missing Details**:
1. **Performance Characteristics**
   - Expected reconciliation latency (p50, p95, p99)
   - Throughput (reconciliations/second)
   - API response time targets
   - Event processing rate

2. **Scalability Limits**
   - Maximum LambdaFunctions per operator instance
   - Maximum concurrent reconciliations
   - Resource requirements (CPU/Memory) per function
   - Horizontal scaling strategy

3. **Bottleneck Analysis**
   - Kubernetes API rate limiting
   - RabbitMQ broker throughput
   - Kaniko job creation rate
   - ECR push rate limits

4. **Optimization Strategies**
   - Batch reconciliation
   - Event batching
   - Caching strategies
   - Connection pooling

**Recommendation**: Add section "Performance & Scalability" with:
- Performance benchmarks
- Scalability testing results
- Resource sizing guidelines
- Optimization recommendations

### 💰 Cost Optimization (MEDIUM PRIORITY)

**Status**: Mentioned in AI Agent context but not detailed

**Missing Details**:
1. **Cost Model**
   - Cost per LambdaFunction (build + runtime)
   - Operator infrastructure costs
   - Storage costs (Registry, MinIO/S3/GCS)
   - Network costs

2. **Cost Optimization Strategies**
   - Spot instances for build jobs
   - ECR lifecycle policies
   - Storage lifecycle policies (MinIO/S3/GCS)
   - Resource rightsizing
   - Scale-to-zero benefits

3. **Cost Monitoring**
   - Cost allocation tags
   - Per-function cost tracking
   - Cost anomaly detection
   - Budget alerts

**Recommendation**: Add section "Cost Optimization" with:
- Cost breakdown analysis
- Optimization strategies
- Cost monitoring setup
- ROI calculations

### 🧪 Testing Strategy (MEDIUM PRIORITY)

**Status**: Mentioned in "Next Steps" but not detailed

**Missing Details**:
1. **Unit Testing**
   - Test coverage goals
   - Mocking strategies
   - Test utilities

2. **Integration Testing**
   - Test environment setup
   - Kubernetes test clusters
   - Test fixtures and helpers

3. **E2E Testing**
   - Critical path testing
   - Chaos engineering
   - Load testing scenarios

4. **Testing Tools**
   - Testing frameworks
   - Mock Kubernetes clients
   - Test data generators

**Recommendation**: Add section "Testing Strategy" with:
- Testing pyramid
- Test coverage goals
- Example test cases
- CI/CD integration

### 🔐 Security Deep Dive (MEDIUM PRIORITY)

**Status**: Superficial coverage

**Missing Details**:
1. **RBAC Details**
   - Required permissions breakdown
   - Least privilege principles
   - Service account configuration
   - ClusterRole/Role definitions

2. **Network Security**
   - Network policies
   - Service mesh integration (Linkerd)
   - Ingress/egress rules
   - Pod-to-pod communication

3. **Secret Management**
   - Secret storage (Vault/Sealed Secrets)
   - Secret rotation
   - Image pull secrets
   - AWS credentials management

4. **Container Security**
   - Image scanning
   - Vulnerability management
   - Pod security standards
   - Security contexts

5. **API Security**
   - Authentication mechanisms
   - Authorization policies
   - Rate limiting
   - Input validation

**Recommendation**: Add section "Security Architecture" with:
- RBAC configuration examples
- Network policy examples
- Secret management strategy
- Security best practices

### 📈 Capacity Planning (LOW PRIORITY)

**Status**: Not covered

**Missing Details**:
1. **Resource Forecasting**
   - Growth projections
   - Resource requirements calculation
   - Headroom recommendations

2. **Scaling Triggers**
   - When to scale operator
   - HPA configuration
   - VPA recommendations

3. **Load Testing**
   - Test scenarios
   - Performance baselines
   - Stress testing

**Recommendation**: Add section "Capacity Planning" with:
- Forecasting model
- Scaling guidelines
- Load testing results

### 🔄 Disaster Recovery & High Availability (LOW PRIORITY)

**Status**: Not covered

**Missing Details**:
1. **High Availability**
   - Operator replication strategy
   - Leader election
   - Multi-zone deployment

2. **Disaster Recovery**
   - Backup strategies
   - Recovery procedures
   - RTO/RPO targets

3. **Failure Scenarios**
   - Operator failure handling
   - Data loss prevention
   - State recovery

**Recommendation**: Add section "Disaster Recovery" with:
- HA architecture
- Backup procedures
- Recovery runbooks

### 🤖 AI Agent Integration Details (MEDIUM PRIORITY)

**Status**: Overview provided but technical details missing

**Missing Details**:
1. **ML/AI Framework Integration**
   - Model serving architecture
   - Inference pipeline
   - Model versioning
   - A/B testing

2. **Data Pipeline**
   - Event data preprocessing
   - Feature engineering
   - Data storage (vector DB)
   - Training data collection

3. **Model Operations**
   - Model deployment
   - Model monitoring
   - Drift detection
   - Retraining triggers

4. **Performance Metrics**
   - Inference latency
   - Model accuracy
   - Prediction confidence
   - Action success rate

**Recommendation**: Expand "AI Agent Integration" section with:
- ML architecture diagram
- Model serving details
- Data pipeline flow
- MLOps practices

### 🔌 Integration Patterns (LOW PRIORITY)

**Status**: Not covered

**Missing Details**:
1. **Service Mesh Integration**
   - Linkerd integration with Policy API (Server + ServerAuthorization)
   - mTLS configuration
   - Traffic policies

2. **API Gateway Integration**
   - External API exposure
   - Rate limiting
   - Authentication

3. **GitOps Integration**
   - ArgoCD/Flux integration
   - CRD sync
   - Deployment automation

**Recommendation**: Add section "Integration Patterns" with:
- Service mesh setup
- API gateway configuration
- GitOps workflows

### 📝 API Versioning & Compatibility (LOW PRIORITY)

**Status**: Not covered

**Missing Details**:
1. **API Versioning Strategy**
   - Version scheme (v1, v2)
   - Backward compatibility
   - Deprecation policy

2. **CRD Versioning**
   - CRD version management
   - Migration strategies
   - Conversion webhooks

**Recommendation**: Add section "API Versioning" with:
- Versioning strategy
- Migration guides
- Compatibility matrix

### 📊 Multi-Tenancy Details (LOW PRIORITY)

**Status**: Mentioned but not detailed

**Missing Details**:
1. **Namespace Isolation**
   - Resource quotas
   - Network policies per namespace
   - RBAC per namespace

2. **Tenant Management**
   - Tenant onboarding
   - Resource allocation
   - Cost tracking per tenant

**Recommendation**: Expand multi-tenancy section with:
- Namespace strategy
- Resource quotas
- Tenant isolation

### 🎯 Rate Limiting & Throttling (LOW PRIORITY)

**Status**: Not covered in operator context

**Missing Details**:
1. **API Rate Limiting**
   - Rate limit configuration
   - Throttling strategies
   - Rate limit headers

2. **Reconciliation Rate Limiting**
   - Max reconciliations per second
   - Backoff strategies
   - Queue management

**Recommendation**: Add section "Rate Limiting" with:
- Rate limit configuration
- Throttling strategies
- Best practices

---

## 📋 Documentation Completeness Checklist

### High Priority (Should be added)
- [x] Observability & Metrics (detailed) - ✅ Added (Section: "📊 Observability & Monitoring")
- [x] Performance & Scalability - ✅ Added (Section: "🚀 Performance & Scalability")
- [x] Security Deep Dive - ✅ Added (Section: "🔐 Security Architecture")
- [x] AI Agent ML Integration Details - ✅ Added (Section: "🤖 AI Agent ML Integration Details")

### Medium Priority (Nice to have)
- [x] Cost Optimization - ✅ Added (Section: "💰 Cost Optimization")
- [x] Testing Strategy - ✅ Added (Section: "🧪 Testing Strategy")
- [x] Disaster Recovery - ✅ Added (Section: "🔄 Disaster Recovery & High Availability")

### Low Priority (Future enhancements)
- [x] Capacity Planning - ✅ Added (Section: "📈 Capacity Planning")
- [x] Integration Patterns - ✅ Added (Section: "🔌 Integration Patterns")
- [x] API Versioning - ✅ Added (Section: "📝 API Versioning & Compatibility")
- [x] Multi-Tenancy Details - ✅ Added (Section: "📊 Multi-Tenancy Details")
- [x] Rate Limiting - ✅ Added (Section: "🎯 Rate Limiting & Throttling")

---

**Review Date**: 2025-01-20  
**Reviewer**: ML Engineer  
**Documentation Coverage**: ~100% ✅  
**Status**: All identified topics have been added to the documentation  
**Last Updated**: 2025-01-20

