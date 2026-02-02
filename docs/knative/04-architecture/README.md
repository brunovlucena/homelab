# 🏗️ Architecture Documentation

**Deep technical documentation for system design, data flow, and technology choices**

---

## 📚 Architecture Documents

| Document | Description | Time |
|----------|-------------|------|
| **[SYSTEM_DESIGN.md](SYSTEM_DESIGN.md)** | High-level architecture overview | 25 min |
| **[TECHNOLOGY_STACK.md](TECHNOLOGY_STACK.md)** | Why we chose each technology | 18 min |
| **[DATA_FLOW.md](DATA_FLOW.md)** | How events and data move through the system | 15 min |
| **[BUILD_PIPELINE.md](BUILD_PIPELINE.md)** | Container build process deep-dive | 20 min |
| **[DEPLOYMENT_MODEL.md](DEPLOYMENT_MODEL.md)** | How functions are deployed and managed | 15 min |
| **[OBSERVABILITY_SPECIFICATION.md](OBSERVABILITY_SPECIFICATION.md)** | Metrics, logging, and tracing architecture | 15 min |
| **[CLOUDEVENTS_SPECIFICATION.md](CLOUDEVENTS_SPECIFICATION.md)** | CloudEvents format, types, and Flux CD integration | 25 min |
| **[NOTIFI_INTEGRATION.md](NOTIFI_INTEGRATION.md)** | Integration with Notifi notification platform | 20 min |
| **[DLQ_FLOWS.md](DLQ_FLOWS.md)** | Dead Letter Queue handling and retry patterns | 15 min |

---

## 🎯 Quick Reference

### System Architecture Overview

```
┌────────────────────────────────────────────────────────────────┐
│                    KNATIVE LAMBDA PLATFORM                     │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  📤 INPUT LAYER                                                │
│     ├─ S3 Storage (parser code + dependencies)                 │
│     ├─ CloudEvent (build.start, job.start, service.delete)     │
│     └─ RabbitMQ Broker (event routing)                         │
│                                                                │
│  🔨 BUILD LAYER                                                │
│     ├─ Builder Service (Go) - Orchestrates builds              │
│     ├─ Kaniko Jobs - Secure container builds                   │
│     ├─ ECR Registry - Container image storage                  │
│     └─ Sidecar Monitor - Job status tracking                   │
│                                                                │
│  ☁️ RUNTIME LAYER                                              │
│     ├─ Knative Serving - Auto-scaling functions                │
│     ├─ Internal Networking - Service discovery only            │
│     ├─ Activator - Scale-from-zero handler                     │
│     └─ Queue Proxy - Request buffering                         │
│                                                                │
│  📊 OBSERVABILITY LAYER                                        │
│     ├─ Prometheus - Metrics collection                         │
│     ├─ Tempo - Distributed tracing                             │
│     ├─ Loki - Log aggregation                                  │
│     └─ Grafana - Visualization                                 │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

→ **[Complete System Design](SYSTEM_DESIGN.md)**

---

## 🔄 Data Flow Overview

### Build Flow

```
Developer                Builder Service          Kaniko Job              Knative Serving
    |                           |                      |                       |
    |-- Upload code to S3 ----->|                      |                       |
    |                           |                      |                       |
    |-- Send CloudEvent ------->|                      |                       |
    |   (build.start)           |                      |                       |
    |                           |                      |                       |
    |                           |-- Create Job ------->|                       |
    |                           |                      |                       |
    |                           |                      |-- Fetch from S3       |
    |                           |                      |                       |
    |                           |                      |-- Build image         |
    |                           |                      |                       |
    |                           |                      |-- Push to ECR         |
    |                           |                      |                       |
    |                           |<-- Job complete -----|                       |
    |                           |                      |                       |
    |                           |-- Create Service -------------------------->|
    |                           |                      |                       |
    |<------------------------- Success -------------------------------------->|
```

→ **[Detailed Data Flow](DATA_FLOW.md)**

---

## 🔄 Detailed Component Flow (ASCII)

### Complete Request Processing Flow

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    KNATIVE LAMBDA REQUEST FLOW                                  │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘

External System
       │
       │ HTTP POST CloudEvent
       ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    HTTP LAYER                                                   │
├─────────────────────────────────────────────────────────────────────────────────────────────────┤
│  HTTP Handler                                                                                   │
│       │                                                                                         │
│       │ Route to CloudEvent Handler                                                             │
│       ▼                                                                                         │
│  CloudEvent Handler                                                                             │
│       │                                                                                         │
│       ├─ Parse CloudEvent (headers + body)                                                     │
│       ├─ Validate Content-Type                                                                  │
│       ├─ Extract Correlation ID                                                                 │
│       └─ Setup Distributed Tracing                                                              │
│       │                                                                                         │
│       │ Get Event Handler from Container                                                        │
│       ▼                                                                                         │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘
       │
       │ ProcessCloudEvent()
       ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                 EVENT PROCESSING LAYER                                          │
├─────────────────────────────────────────────────────────────────────────────────────────────────┤
│  Event Handler (Orchestrator)                                                                  │
│       │                                                                                         │
│       ├─ Validate Event                                                                         │
│       ├─ Record Metrics                                                                         │
│       └─ Route by Event Type                                                                    │
│       │                                                                                         │
│       ┌─ build.start ──────────────────────────────────────────────────────────────────────────┐ │
│       │                                                                                         │ │
│       │  Build Start Flow:                                                                      │ │
│       │       │                                                                                 │ │
│       │       ├─ Parse Build Request                                                            │ │
│       │       ├─ Create Build Context (BuildContextManager)                                    │ │
│       │       │   │                                                                             │ │
│       │       │   ├─ Generate tar.gz archive                                                   │ │
│       │       │   ├─ Upload to S3 temp bucket                                                  │ │
│       │       │   └─ Return build context key                                                  │ │
│       │       │                                                                                 │ │
│       │       ├─ Create Kaniko Job (AsyncJobCreator)                                           │ │
│       │       │   │                                                                             │ │
│       │       │   ├─ Queue job creation request                                                │ │
│       │       │   ├─ Worker Pool processes request                                             │ │
│       │       │   │   │                                                                         │ │
│       │       │   │   ├─ Worker picks up request                                               │ │
│       │       │   │   ├─ Call JobManager.CreateJob()                                           │ │
│       │       │   │   │   │                                                                     │ │
│       │       │   │   │   ├─ Find existing job (KISS: delete if exists)                        │ │
│       │       │   │   │   ├─ Create Kaniko job spec                                            │ │
│       │       │   │   │   ├─ Apply to Kubernetes API                                           │ │
│       │       │   │   │   └─ Return created job                                                │ │
│       │       │   │   │                                                                         │ │
│       │       │   │   └─ Send result to result queue                                           │ │
│       │       │   │                                                                             │ │
│       │       │   └─ Result processor stores result by correlation ID                          │ │
│       │       │                                                                                 │ │
│       │       └─ Return job name immediately (async)                                           │ │
│       │                                                                                         │ │
│       └─ Return "started" response                                                              │ │
│       │                                                                                         │
│       ┌─ build.complete ──────────────────────────────────────────────────────────────────────┐ │
│       │                                                                                         │ │
│       │  Build Complete Flow:                                                                   │ │
│       │       │                                                                                 │ │
│       │       ├─ Parse completion data                                                          │ │
│       │       ├─ Create Knative Service (ServiceManager)                                       │ │
│       │       │   │                                                                             │ │
│       │       │   ├─ Generate service spec                                                     │ │
│       │       │   ├─ Apply to Kubernetes API                                                   │ │
│       │       │   └─ Return service details                                                    │ │
│       │       │                                                                                 │ │
│       │       ├─ Create Knative Trigger (ServiceManager)                                       │ │
│       │       │   │                                                                             │ │
│       │       │   ├─ Generate trigger spec                                                     │ │
│       │       │   ├─ Apply to Kubernetes API                                                   │ │
│       │       │   └─ Return trigger details                                                    │ │
│       │       │                                                                                 │ │
│       │       └─ Return "service_created" response                                             │ │
│       │                                                                                         │ │
│       └─ Return success response                                                                │ │
│       │                                                                                         │
│       ┌─ job.start ───────────────────────────────────────────────────────────────────────────┐ │
│       │                                                                                         │ │
│       │  Job Start Flow:                                                                        │ │
│       │       │                                                                                 │ │
│       │       ├─ Parse job start data                                                           │ │
│       │       ├─ Record job start metrics                                                       │ │
│       │       └─ Return "acknowledged" response                                                 │ │
│       │                                                                                         │ │
│       └─ Return success response                                                                │ │
│       │                                                                                         │
│       ┌─ parser.start/complete ───────────────────────────────────────────────────────────────┐ │
│       │                                                                                         │ │
│       │  Parser Flow:                                                                           │ │
│       │       │                                                                                 │ │
│       │       ├─ Parse parser event data                                                        │ │
│       │       ├─ Record parser metrics                                                          │ │
│       │       └─ Return "acknowledged" response                                                 │ │
│       │                                                                                         │ │
│       └─ Return success response                                                                │ │
│       │                                                                                         │
│       ┌─ service.delete ──────────────────────────────────────────────────────────────────────┐ │
│       │                                                                                         │ │
│       │  Service Delete Flow:                                                                   │ │
│       │       │                                                                                 │ │
│       │       ├─ Parse delete request                                                           │ │
│       │       ├─ Delete Knative Service (ServiceManager)                                       │ │
│       │       ├─ Delete Knative Trigger (ServiceManager)                                       │ │
│       │       └─ Return "deleted" response                                                     │ │
│       │                                                                                         │ │
│       └─ Return success response                                                                │ │
│                                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘
       │
       │ HandlerResponse
       ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    RESPONSE LAYER                                               │
├─────────────────────────────────────────────────────────────────────────────────────────────────┤
│  CloudEvent Handler                                                                             │
│       │                                                                                         │
│       ├─ Set response headers (correlation ID, trace ID)                                       │
│       ├─ Encode JSON response                                                                   │
│       └─ Send HTTP response                                                                     │
│       │                                                                                         │
│       │ HTTP Response                                                                           │
│       ▼                                                                                         │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘
       │
       │ HTTP Response
       ▼
External System
```

### Kubernetes Job Processing Flow

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    KUBERNETES JOB FLOW                                          │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘

JobManager.CreateJob()
       │
       ├─ Find existing job (by third-party-id + parser-id labels)
       ├─ Delete existing job if found (KISS principle)
       ├─ Create Kaniko job spec:
       │   ├─ Container: Kaniko with S3 context
       │   ├─ Sidecar: CloudEvent monitor
       │   ├─ Resources: CPU/Memory limits
       │   └─ Environment: AWS credentials, ECR settings
       ├─ Apply job to Kubernetes API
       └─ Return created job
       │
       ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    KUBERNETES CLUSTER                                           │
├─────────────────────────────────────────────────────────────────────────────────────────────────┤
│  Kaniko Pod                                                                                    │
│       │                                                                                         │
│       ├─ Fetch build context from S3                                                           │
│       ├─ Build container image using Dockerfile                                                │
│       ├─ Push image to ECR registry                                                            │
│       └─ Update job status                                                                     │
│       │                                                                                         │
│       │ Job Status Update                                                                       │
│       ▼                                                                                         │
│  Sidecar Container                                                                             │
│       │                                                                                         │
│       ├─ Monitor Kaniko container status                                                       │
│       ├─ Detect job completion (success/failure)                                               │
│       ├─ Send build.complete CloudEvent to broker                                              │
│       └─ Include image URI and build metadata                                                  │
│       │                                                                                         │
│       │ CloudEvent (build.complete)                                                            │
│       ▼                                                                                         │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘
       │
       │ CloudEvent
       ▼
RabbitMQ Broker
       │
       │ CloudEvent
       ▼
Event Handler (build.complete processing)
```

### Component Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    COMPONENT ARCHITECTURE                                       │
└─────────────────────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  HTTP Handler   │    │CloudEvent Handler│    │ Event Handler   │    │Component Container│
│                 │    │                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │Route requests│ │    │ │Parse events │ │    │ │Route by type│ │    │ │Dependency   │ │
│ │to handlers  │ │    │ │Validate     │ │    │ │Orchestrate  │ │    │ │Injection    │ │
│ │             │ │    │ │Tracing      │ │    │ │Components   │ │    │ │             │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │                       │
         └───────────────────────┼───────────────────────┼───────────────────────┘
                                 │                       │
                                 ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│BuildContext     │    │AsyncJobCreator  │    │ JobManager      │    │ServiceManager   │
│Manager          │    │                 │    │                 │    │                 │
│                 │    │                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │Create tar.gz│ │    │ │Worker Pool  │ │    │ │K8s Job Ops  │ │    │ │Knative Ops  │ │
│ │Upload to S3 │ │    │ │Queue Mgmt   │ │    │ │Conflict Res  │ │    │ │Service Mgmt │ │
│ │S3 Integration│ │    │ │Retry Logic  │ │    │ │Rate Limiting│ │    │ │Trigger Mgmt │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │                       │
         └───────────────────────┼───────────────────────┼───────────────────────┘
                                 │                       │
                                 ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   AWS S3        │    │   Kubernetes    │    │   ECR Registry  │    │  Knative        │
│   Storage       │    │   API Server    │    │   Container     │    │  Serving        │
│                 │    │                 │    │   Images        │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │Build Context│ │    │ │Job Creation │ │    │ │Image Storage│ │    │ │Auto-scaling │ │
│ │Archives     │ │    │ │Pod Management│ │    │ │Image Tags   │ │    │ │Scale-to-zero│ │
│ │Temp Storage │ │    │ │Status Updates│ │    │ │Pull/Push    │ │    │ │Load Balancing│ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

---

## 📊 Technology Stack

| Layer | Component | Technology | Purpose |
|-------|-----------|------------|---------|
| **Event Bus** | Message Broker | RabbitMQ | CloudEvents routing |
| **Build** | Container Builder | Kaniko | Secure image builds |
| **Build** | Orchestrator | Go 1.24 | Build coordination |
| **Runtime** | Serverless Platform | Knative Serving | Auto-scaling functions |
| **Runtime** | Networking | Internal (ClusterIP) | Internal service discovery |
| **Storage** | Object Storage | S3 / MinIO | Source code |
| **Storage** | Container Registry | ECR | Docker images |
| **Observability** | Metrics | Prometheus | Time-series metrics |
| **Observability** | Tracing | Tempo | Distributed tracing |
| **Observability** | Logging | Loki | Log aggregation |
| **Orchestration** | Platform | Kubernetes | Container orchestration |

→ **[Technology Rationale](TECHNOLOGY_STACK.md)**

---

## 🎓 For Different Audiences

**Architects**: Start with [SYSTEM_DESIGN.md](SYSTEM_DESIGN.md)  
**Engineers**: Start with [DATA_FLOW.md](DATA_FLOW.md)  
**DevOps**: Start with [DEPLOYMENT_MODEL.md](DEPLOYMENT_MODEL.md)  
**SRE**: Start with [OBSERVABILITY.md](OBSERVABILITY.md)  
**Product/Planning**: Start with [BUILD_PIPELINE.md](BUILD_PIPELINE.md)  
**Integration Teams**: Start with [NOTIFI_INTEGRATION.md](NOTIFI_INTEGRATION.md)

---

## 🔗 Related Documentation

- **[Engineering Docs](../03-for-engineers/)** - Role-specific implementation guides
- **[Operations](../05-operations/)** - Running in production
- **[Decisions](../07-decisions/)** - ADRs and design rationale
- **[Getting Started](../01-getting-started/)** - Introduction and setup

---

## 📖 Key Architectural Principles

### 1. Event-Driven Architecture

All operations triggered by CloudEvents:
- **build.start** → Initiates container build
- **job.start** → Creates Kubernetes job
- **service.delete** → Removes deployed function

### 2. Separation of Concerns

- **Builder Service**: Orchestration only
- **Kaniko Jobs**: Build execution only
- **Knative Serving**: Runtime only
- **RabbitMQ**: Event routing only

### 3. Security by Default

- Non-root containers
- RBAC with least privilege
- No Docker daemon (Kaniko)
- TLS/mTLS communication
- Resource quotas

### 4. Observability First

- Prometheus metrics (RED method)
- Structured JSON logging
- OpenTelemetry tracing
- Pre-built Grafana dashboards

### 5. Cloud-Native Patterns

- **12-Factor App** compliance
- **GitOps** deployment model
- **Immutable infrastructure**
- **Infrastructure as Code**

### 6. GitOps Integration (Flux CD)

CloudEvents trigger Flux reconciliation:
- **lifecycle.function.ready** → Update dependent configs
- **lifecycle.build.completed** → Trigger ImagePolicy scan
- **Security alerts** → Auto-deploy remediation

→ **[Flux CD Integration Details](CLOUDEVENTS_SPECIFICATION.md#flux-cd-integration-cdevents)**

---

## 🏛️ Design Decisions

Key architectural decisions documented:

| Decision | Rationale |
|----------|-----------|
| **[Why Kaniko?](../07-decisions/WHY_KANIKO.md)** | Secure builds without Docker daemon |
| **[Why Knative?](../07-decisions/WHY_KNATIVE.md)** | Industry-standard serverless on K8s |
| **[Why RabbitMQ?](../07-decisions/WHY_RABBITMQ.md)** | Simple, reliable event routing |
| **[Why CloudEvents?](../07-decisions/WHY_CLOUDEVENTS.md)** | Vendor-neutral event format |

---

## 🔍 Deep Dives

### Build Pipeline

Understand how code becomes a running function:

1. **Code upload** to S3
2. **Event generation** (CloudEvent)
3. **Build orchestration** (Builder Service)
4. **Container build** (Kaniko)
5. **Image push** (ECR)
6. **Service deployment** (Knative)
7. **Auto-scaling** (Knative Autoscaler)

→ **[Build Pipeline Details](BUILD_PIPELINE.md)**

---

### Deployment Model

How functions are deployed and managed:

- **Namespaces**: Environment isolation
- **Helm Charts**: Declarative configuration
- **GitOps**: Flux CD automation
- **Versioning**: Semantic versioning
- **Rollbacks**: Automatic on failure

→ **[Deployment Model](DEPLOYMENT_MODEL.md)**

---

### Observability

Full-stack observability:

```
Application
    ↓
Traces (OpenTelemetry)
    ↓
Metrics (Prometheus)
    ↓
Logs (Loki)
    ↓
Dashboards (Grafana)
    ↓
Alerts (Alertmanager)
```

→ **[Observability Architecture](OBSERVABILITY.md)**

---

## 📐 Architectural Diagrams

### Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│  Developer Workstation                                   │
│  ├─ AWS CLI (S3 upload)                                 │
│  └─ Python/Node/Go code                                 │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  S3 / MinIO                                              │
│  └─ knative-lambda-{env}-fusion-modules-tmp             │
│     └─ global/parser/{parser-id}/                       │
│        ├─ parser.py                                     │
│        └─ requirements.txt                              │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  RabbitMQ Cluster                                        │
│  ├─ Exchange: knative-broker                            │
│  ├─ Queue: build-events                                 │
│  └─ Queue: job-events                                   │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Builder Service (Go)                                    │
│  ├─ CloudEvent Consumer                                 │
│  ├─ Kubernetes Client                                   │
│  ├─ S3 Client                                           │
│  └─ Prometheus Exporter                                 │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        ▼                          ▼
┌─────────────────┐    ┌─────────────────────┐
│  Kaniko Job     │    │  Knative Service    │
│  ├─ Build       │    │  ├─ Auto-scaling    │
│  ├─ Push ECR    │    │  ├─ Health checks   │
│  └─ Metrics     │    │  └─ Event routing   │
└─────────────────┘    └─────────────────────┘
```

---

## 🔐 Security Architecture

### Defense in Depth

```
Layer 1: Network Policies
    ↓
Layer 2: RBAC (Kubernetes)
    ↓
Layer 3: Pod Security Standards
    ↓
Layer 4: Non-root containers
    ↓
Layer 5: Resource quotas
    ↓
Layer 6: Image scanning (Trivy)
    ↓
Layer 7: Secrets encryption
```

---

## 🚀 Scalability Architecture

### Horizontal Scaling

| Component | Scaling Strategy |
|-----------|-----------------|
| **Builder Service** | HPA (CPU/Memory) |
| **Kaniko Jobs** | Parallel jobs (rate-limited) |
| **Knative Services** | KPA (request-based) |
| **RabbitMQ** | Cluster (3+ nodes) |

### Vertical Scaling

| Component | Resource Tuning |
|-----------|----------------|
| **Builder Service** | 256Mi-512Mi, 100m-500m CPU |
| **Kaniko Jobs** | 1Gi-4Gi, 500m-2000m CPU |
| **Knative Services** | User-defined limits |

---

**Last Updated**: October 29, 2025  
**Version**: 1.0.0

