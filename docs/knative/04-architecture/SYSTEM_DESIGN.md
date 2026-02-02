# 🏛️ System Design

**Comprehensive architecture overview of Knative Lambda**

---

## 📖 Table of Contents

- [High-Level Architecture](#high-level-architecture)
- [Component Details](#component-details)
- [Data Flow](#data-flow)
- [Scaling Architecture](#scaling-architecture)
- [Security Architecture](#security-architecture)
- [Network Architecture](#network-architecture)
- [Storage Architecture](#storage-architecture)

---

## 🏗️ High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      KNATIVE LAMBDA PLATFORM                    │
│                         (Kubernetes Cluster)                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  📤 INGRESS LAYER                                        │   │
│  │  ├─ S3/MinIO: Source code storage                        │   │
│  │  ├─ CloudEvents: Standard event format                   │   │
│  │  └─ RabbitMQ: Event routing and delivery                 │   │
│  └──────────────────────────────────────────────────────────┘   │
│                            ↓                                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  🔨 BUILD ORCHESTRATION LAYER                            │   │
│  │                                                          │   │
│  │  Builder Service (Go)                                    │   │
│  │  ├─ Event Processing (CloudEvents SDK)                   │   │
│  │  ├─ Job Management (Kubernetes Client)                   │   │
│  │  ├─ S3 Integration (AWS SDK)                             │   │
│  │  ├─ Rate Limiting (Multi-level)                          │   │
│  │  └─ Observability (Prometheus + OTel)                    │   │
│  └──────────────────────────────────────────────────────────┘   │
│                            ↓                                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  🏗️ BUILD EXECUTION LAYER                                │   │
│  │                                                          │   │
│  │  Kaniko Jobs (Ephemeral)                                 │   │
│  │  ├─ Fetch code from S3                                   │   │
│  │  ├─ Generate Dockerfile (dynamic)                        │   │
│  │  ├─ Build container image (no Docker daemon)             │   │
│  │  ├─ Push to ECR registry                                 │   │
│  │  └─ Cleanup (self-delete after completion)               │   │
│  │                                                          │   │
│  │  Job Sidecar (Monitoring)                                │   │
│  │  ├─ Monitor job status                                   │   │
│  │  ├─ Emit metrics                                         │   │
│  │  └─ Cleanup stale jobs                                   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                            ↓                                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  ☁️ SERVERLESS RUNTIME LAYER                             │   │
│  │                                                          │   │
│  │  Knative Serving                                         │   │
│  │  ├─ Service: Function definition                         │   │
│  │  ├─ Route: Traffic routing                               │   │
│  │  ├─ Configuration: Immutable config                      │   │
│  │  └─ Revision: Versioned deployments                      │   │
│  │                                                          │   │
│  │  Knative Autoscaler                                      │   │
│  │  ├─ KPA: Request-based scaling                           │   │
│  │  ├─ Scale-to-zero: Idle functions                        │   │
│  │  ├─ Activator: Cold start handler                        │   │
│  │  └─ Queue Proxy: Request buffering                       │   │
│  │                                                          │   │
│  │  Internal Networking                                     │   │
│  │  ├─ Service Discovery (ClusterIP)                        │   │
│  │  ├─ Internal Load Balancing                              │   │
│  │  └─ No External Exposure                                 │   │
│  └──────────────────────────────────────────────────────────┘   │
│                            ↓                                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  📊 OBSERVABILITY LAYER                                  │   │
│  │                                                          │   │
│  │  Metrics (Prometheus)                                    │   │
│  │  ├─ Builder metrics (build duration, success rate)       │   │
│  │  ├─ Kaniko metrics (job status, resource usage)          │   │
│  │  ├─ Knative metrics (request rate, latency)              │   │
│  │  └─ Custom metrics (business logic)                      │   │
│  │                                                          │   │
│  │  Traces (Tempo + OpenTelemetry)                          │   │
│  │  ├─ Build traces (S3 → Kaniko → ECR)                     │   │
│  │  ├─ Request traces (Ingress → Function → Response)       │   │
│  │  └─ Event traces (RabbitMQ → Builder → Knative)          │   │
│  │                                                          │   │
│  │  Logs (Loki + Fluent Bit)                                │   │
│  │  ├─ Structured JSON logs                                 │   │
│  │  ├─ Contextual enrichment (trace ID, parser ID)          │   │
│  │  └─ Log aggregation and querying                         │   │
│  │                                                          │   │
│  │  Dashboards (Grafana)                                    │   │
│  │  ├─ Comprehensive Dashboard (all metrics)                │   │
│  │  ├─ Flagger Dashboard (canary deployments)               │   │
│  │  └─ Custom dashboards (per-function)                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

External Dependencies:
├─ AWS S3: Source code storage
├─ AWS ECR: Container registry (339954290315.dkr.ecr.us-west-2.amazonaws.com)
├─ MinIO: Optional S3-compatible local storage
└─ Notifi Platform: External notification services
   ├─ Subscription Manager: User subscription management
   ├─ Storage Manager: Data persistence and module storage
   ├─ Fetch Proxy: External API access
   └─ Blockchain Manager: Multi-chain RPC access
```

---

## 🔧 Component Details

### 1. Builder Service

**Purpose**: Orchestrates the entire build-to-deployment pipeline.

**Technology**: Go 1.24

**Responsibilities**:
- ✅ Consume CloudEvents from RabbitMQ
- ✅ Validate event payloads
- ✅ Fetch parser metadata from S3
- ✅ Create Kaniko build jobs
- ✅ Monitor job status
- ✅ Create/update/delete Knative Services
- ✅ Emit Prometheus metrics
- ✅ Handle errors and retries

**Configuration**:
```yaml
builderService:
  replicas: 1  # Active-passive (only 1 processes events)
  resources:
    requests:
      memory: "256Mi"
      cpu: "100m"
    limits:
      memory: "512Mi"
      cpu: "500m"
  env:
    - name: AWS_REGION
      value: "us-west-2"
    - name: RABBITMQ_URL
      value: "amqp://rabbitmq-cluster-dev:5672"
    - name: LOG_LEVEL
      value: "info"
```

**Key Features**:
- **Rate Limiting**: Multi-level (build context, K8s jobs, S3, client)
- **Resilience**: Exponential backoff, circuit breakers
- **Observability**: Structured logging, distributed tracing
- **Security**: Non-root, read-only filesystem

---

### 2. Kaniko Jobs

**Purpose**: Build container images securely without Docker daemon.

**Technology**: Kaniko (Google)

**Workflow**:
```
1. Job created by Builder Service
   ↓
2. Init container: Fetch code from S3 → /workspace
   ↓
3. Kaniko container: Build image from /workspace
   ↓
4. Push image to ECR (with retries)
   ↓
5. Job completes → Auto-cleanup (TTL: 3600s)
```

**Job Spec**:
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: kaniko-{{parser-id}}
  labels:
    app: kaniko
    parser-id: {{parser-id}}
spec:
  backoffLimit: 2
  ttlSecondsAfterFinished: 3600  # Auto-cleanup after 1 hour
  template:
    spec:
      restartPolicy: Never
      initContainers:
        - name: fetch-code
          image: amazon/aws-cli:2.15.0
          command: ["/bin/sh", "-c"]
          args:
            - |
              aws s3 sync s3://{{bucket}}/{{prefix}} /workspace/
          volumeMounts:
            - name: workspace
              mountPath: /workspace
      containers:
        - name: kaniko
          image: gcr.io/kaniko-project/executor:v1.19.0
          args:
            - "--dockerfile=/workspace/Dockerfile"
            - "--context=/workspace"
            - "--destination={{ecr-repo}}:{{tag}}"
            - "--cache=true"
            - "--compressed-caching=false"
          volumeMounts:
            - name: workspace
              mountPath: /workspace
          resources:
            requests:
              memory: "1Gi"
              cpu: "500m"
            limits:
              memory: "4Gi"
              cpu: "2000m"
```

**Security**:
- ❌ No Docker daemon (eliminates privileged containers)
- ✅ Non-root execution
- ✅ IRSA for AWS credentials (no static keys)
- ✅ Read-only root filesystem

---

### 3. Knative Serving

**Purpose**: Run auto-scaling serverless functions.

**Components**:

**Service** - Top-level resource:
```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: {{parser-id}}
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/min-scale: "0"
        autoscaling.knative.dev/max-scale: "10"
        autoscaling.knative.dev/target: "100"
        autoscaling.knative.dev/scale-down-delay: "30s"
    spec:
      containers:
        - image: {{ecr-repo}}:{{tag}}
          ports:
            - containerPort: 8080
          env:
            - name: PARSER_ID
              value: {{parser-id}}
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "512Mi"
              cpu: "500m"
```

**Autoscaling Modes**:

| Mode | Trigger | Use Case |
|------|---------|----------|
| **KPA** (default) | Request concurrency | Bursty traffic |
| **HPA** | CPU/Memory | Steady load |

**Scale-to-Zero**:
```
Request arrives → Activator buffers request
                ↓
             Autoscaler spins up pod
                ↓
             Request forwarded to pod
                ↓
             Pod processes request
                ↓
             30s idle → Pod terminated
```

---

### 4. RabbitMQ

**Purpose**: CloudEvents routing and delivery.

**Topology**:
```
┌──────────────┐
│  Publisher   │ (Builder Service, External)
└──────┬───────┘
       │ CloudEvent
       ↓
┌────────────────────────────────────┐
│  Exchange: knative-broker          │
│  Type: topic                       │
└────────┬───────────────────────────┘
         │
    ┌────┴─────────┬──────────────┐
    ↓              ↓              ↓
┌─────────┐  ┌──────────┐  ┌──────────────┐
│  Queue  │  │  Queue   │  │    Queue     │
│  build  │  │   job    │  │   service    │
└────┬────┘  └────┬─────┘  └──────┬───────┘
     │            │               │
     ↓            ↓               ↓
  Consumer     Consumer        Consumer
  (Builder)    (Builder)       (Builder)
```

**Configuration**:
```yaml
# RabbitMQ Cluster (3 nodes for HA)
apiVersion: rabbitmq.com/v1beta1
kind: RabbitmqCluster
metadata:
  name: rabbitmq-cluster-dev
spec:
  replicas: 3
  resources:
    requests:
      memory: "2Gi"
      cpu: "500m"
    limits:
      memory: "4Gi"
      cpu: "1000m"
  rabbitmq:
    additionalConfig: |
      consumer_timeout = 3600000
      heartbeat = 60
```

---

## 🔄 Complete Data Flow

### Build Flow (Detailed)

```
┌─────────────────────────────────────────────────────────────────┐
│  PHASE 1: Code Upload                                           │
└─────────────────────────────────────────────────────────────────┘

Developer
   │
   │ aws s3 cp parser.py s3://bucket/global/parser/${PARSER_ID}/
   ↓
┌─────────────────────┐
│  S3 Bucket          │
│  ├─ parser.py       │
│  └─ requirements.txt│
└─────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  PHASE 2: Build Trigger                                         │
└─────────────────────────────────────────────────────────────────┘

Developer/System
   │
   │ Publish CloudEvent (type: build.start)
   ↓
┌──────────────────────┐
│  RabbitMQ            │
│  Exchange: broker    │
│  ├─ Routing: build.*│
│  └─ Queue: build    │
└──────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  PHASE 3: Build Orchestration                                   │
└─────────────────────────────────────────────────────────────────┘

Builder Service (Go)
   │
   ├─ 1. Consume CloudEvent from RabbitMQ
   │     ├─ Validate event schema
   │     ├─ Extract parser_id, s3_prefix
   │     └─ Rate limit check (10 concurrent builds)
   │
   ├─ 2. Fetch S3 metadata
   │     ├─ aws s3 ls s3://bucket/${s3_prefix}
   │     ├─ Detect language (parser.py → Python)
   │     └─ Detect dependencies (requirements.txt)
   │
   ├─ 3. Generate Kaniko Job spec
   │     ├─ Init container: S3 sync
   │     ├─ Kaniko container: Build + Push
   │     └─ Sidecar: Metrics exporter
   │
   └─ 4. Create Kubernetes Job
         └─ kubectl create -f kaniko-job.yaml

┌─────────────────────────────────────────────────────────────────┐
│  PHASE 4: Container Build                                       │
└─────────────────────────────────────────────────────────────────┘

Kaniko Job
   │
   ├─ Init Container: fetch-code
   │     ├─ aws s3 sync s3://bucket/${s3_prefix} /workspace/
   │     ├─ Downloaded: parser.py, requirements.txt
   │     └─ Exit 0
   │
   └─ Main Container: kaniko
         │
         ├─ 1. Generate Dockerfile (dynamic)
         │      FROM python:3.9-slim
         │      WORKDIR /app
         │      COPY requirements.txt .
         │      RUN pip install -r requirements.txt
         │      COPY parser.py .
         │      CMD ["python", "parser.py"]
         │
         ├─ 2. Build image layers
         │      ├─ Layer 1: Base image (python:3.9-slim)
         │      ├─ Layer 2: Dependencies (pip install)
         │      └─ Layer 3: Application code (parser.py)
         │
         ├─ 3. Tag image
         │      └─ {{ecr-repo}}:{{parser-id}}
         │
         ├─ 4. Push to ECR (with retries)
         │      └─ docker push {{ecr-repo}}:{{parser-id}}
         │
         └─ 5. Exit 0 (Success)

Builder Service (monitoring)
   │
   └─ Poll job status every 10s
         ├─ Running → Continue polling
         ├─ Succeeded → Proceed to deployment
         └─ Failed → Emit alert, retry logic

┌─────────────────────────────────────────────────────────────────┐
│  PHASE 5: Service Deployment                                    │
└─────────────────────────────────────────────────────────────────┘

Builder Service
   │
   ├─ 1. Create Knative Service
   │     └─ kubectl apply -f knative-service.yaml
   │
   ├─ 2. Wait for Service ready
   │     └─ kubectl wait --for=condition=Ready ksvc/{{parser-id}}
   │
   └─ 3. Emit success CloudEvent
         └─ Publish (type: build.complete)

Knative Serving
   │
   ├─ Create Revision (immutable)
   │     └─ {{parser-id}}-00001
   │
   ├─ Create Route (traffic routing)
   │     └─ 100% → Revision-00001
   │
   └─ Scale to zero (initially idle)

┌─────────────────────────────────────────────────────────────────┐
│  PHASE 6: Function Execution (Internal Only)                    │
└─────────────────────────────────────────────────────────────────┘

Internal Cluster Request
   │
   │ POST http://{{parser-id}}.{{namespace}}.svc.cluster.local
   │ Headers:
   │   ce-type: user.event
   │   ce-source: internal-system
   ↓
Knative Serving (Internal)
   │
   ├─ Route to Activator (if scaled to zero)
   │     ├─ Buffer request (max 30s)
   │     ├─ Trigger scale-up
   │     └─ Wait for pod ready
   │
   └─ Route to Pod (if running)
         ├─ Queue Proxy (sidecar)
         │     ├─ Metrics collection
         │     └─ Request forwarding
         └─ Function Container
               ├─ handler(event)
               └─ Return response

Internal Response
   │
   └─ Return to internal client (200 OK)

**Note**: Functions are only accessible within the Kubernetes cluster
```

---

## 📊 Scaling Architecture

### Horizontal Pod Autoscaling (HPA)

**Builder Service**:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: knative-lambda-builder
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: knative-lambda-builder
  minReplicas: 1
  maxReplicas: 5
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

### Knative Pod Autoscaler (KPA)

**Functions**:
```yaml
autoscaling.knative.dev/class: kpa.autoscaling.knative.dev
autoscaling.knative.dev/metric: concurrency
autoscaling.knative.dev/target: "100"  # 100 concurrent requests per pod
autoscaling.knative.dev/min-scale: "0"  # Scale to zero
autoscaling.knative.dev/max-scale: "10"  # Max 10 pods
```

**Scaling Algorithm**:
```
desired_pods = ceil(total_requests / target_concurrency)

Example:
- 250 concurrent requests
- Target: 100 requests/pod
- Desired: ceil(250/100) = 3 pods
```

---

## 🔐 Security Architecture

### Network Security

```
┌─────────────────────────────────────────────────────────┐
│  Network Policies                                       │
│  ├─ Default Deny All                                    │
│  ├─ Allow Ingress: Internal Services → Functions        │
│  ├─ Allow Egress: Functions → External (443)            │
│  └─ Allow Egress: Builder → K8s API (6443)              │
└─────────────────────────────────────────────────────────┘
```

### RBAC

**Builder Service**:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: builder-service
rules:
  - apiGroups: ["batch"]
    resources: ["jobs"]
    verbs: ["create", "list", "watch", "delete"]
  - apiGroups: ["serving.knative.dev"]
    resources: ["services"]
    verbs: ["create", "update", "delete", "get", "list"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list"]
```

---

## 🌐 Network Architecture

```
Internal Kubernetes Network
     │
     ↓
┌─────────────────────────────────────────────────────────┐
│  Knative Serving (Internal Only)                        │
│  ├─ Service Discovery (ClusterIP)                       │
│  ├─ Internal Load Balancing                             │
│  └─ No External Exposure                                │
└────────┬────────────────────────────────────────────────┘
         │
    ┌────┴─────┬──────────┐
    ↓          ↓          ↓
┌─────────┐ ┌─────────┐ ┌─────────┐
│Function │ │Function │ │Function │
│   Pod   │ │   Pod   │ │   Pod   │
│(Internal)│ │(Internal)│ │(Internal)│
└─────────┘ └─────────┘ └─────────┘
```

**Key Points**:
- ❌ **No Kourier Gateway**: Functions are not exposed externally
- ❌ **No Load Balancer**: No external traffic routing
- ❌ **No TLS Termination**: No external HTTPS endpoints
- ✅ **Internal Only**: Functions accessible only within the cluster
- ✅ **Service Discovery**: Functions discoverable via Kubernetes DNS
- ✅ **ClusterIP Services**: Internal networking only

---

**Last Updated**: October 29, 2025  
**Version**: 1.0.0

