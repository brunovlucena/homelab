# 🚀 Knative Lambda Builder Service

## 📋 Overview

The **Knative Lambda Builder** service is a serverless function builder that mimics the [AWS Lambda](https://aws.amazon.com/lambda/) 🎯. It provides automated build, deployment, and management of serverless functions with comprehensive monitoring and alerting capabilities. 

The service leverages modern observability standards including:
- 📊 **Prometheus metrics format** for real-time monitoring
- 🔍 **OpenTelemetry with Grafana Tempo** for distributed tracing  
- 📝 **Grafana Loki** for centralized log aggregation and analysis

## ✨ Features

- 🏗️ **Serverless Function Building**: Automated build and deployment of Lambda functions (currently JavaScript, with Go and Python support planned)
- ⚡ **Event-Driven Processing**: CloudEvents broker integration for asynchronous processing
- 📊 **Enhanced Monitoring**: Comprehensive metrics collection and observability
- 🚨 **Real-Time Alerting**: PrometheusRules with Slack integration
- 🛡️ **Resilience Patterns**: Circuit breakers, rate limiting, and retry mechanisms
- ⚙️ **Resource Management**: Efficient resource utilization and scaling
- 🏢 **Multi-Tenant Sandboxes**: Support for AWS Spot instances or Bare metal virtual machines as sandboxes using [Kamaji/Clastix](https://github.com/clastix/kamaji) for adding tenants
- 🗄️ **Storage Flexibility**: AWS S3 integration with plans to add [MinIO](https://min.io/) for on-premises S3-compatible storage

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   📡 Broker     │    │   🚀 Knative    │    │   ⚡ AWS Lambda  │
│   (CloudEvents) │───▶│   Lambda        │───▶│   Functions     │
│                 │    │   Builder       │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌────────────────────┐
│   📊 Monitoring │    │   🚨 Alerting   │    │   🔍 Observability │
│   (Prometheus)  │    │   (Slack)       │    │   (Tracing)        │
└─────────────────┘    └─────────────────┘    └────────────────────┘
```

## 🔗 Event Flow Architecture

The Knative Lambda Builder uses a sophisticated event-driven architecture with RabbitMQ as the backing store for the CloudEvents broker. Here's how the components connect and how data flows through the system:

### 📊 Data Flow Diagram

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   📡 Knative    │    │   🎯 Triggers   │    │   🚀 Lambda     │
│      Broker     │◄───│  (Subscription) │───▶│   Functions     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────────┐
│   📋 Trigger    │    │   🔄 Event      │    │   📊 Observability  │
│   Queues        │    │   Processing    │    │   (Monitoring)      │
└─────────────────┘    └─────────────────┘    └─────────────────────┘
```

### 🔄 Event Flow Sequence

```
1. 📨 CloudEvent Source
   ↓
   ↓
2. 📡 Knative Broker (Event Distribution)
   ↓
3. 🎯 Triggers (Event Filtering)
   ↓
4. 🚀 Lambda Functions (Event Processing)
   ↓
5. 📊 Observability Stack (Monitoring/Tracing/Logging)
```

### 📡 Broker → RabbitMQ Connection

The broker uses a **RabbitMQ-based configuration** rather than an in-memory broker:

```yaml
# RabbitMQ Broker Configuration
apiVersion: eventing.knative.dev/v1alpha1
kind: RabbitmqBrokerConfig
spec:
  rabbitmqClusterReference:
    name: rabbitmq-cluster-{environment}
    namespace: rabbitmq
  queueType: quorum

# Broker references the config
apiVersion: eventing.knative.dev/v1
kind: Broker
metadata:
  annotations:
    eventing.knative.dev/broker.class: RabbitMQBroker
spec:
  config:
    apiVersion: eventing.knative.dev/v1alpha1
    kind: RabbitmqBrokerConfig
    name: {app-name}-lambda-broker-config-{environment}
```

**Why RabbitmqBrokerConfig?**
- The broker uses RabbitMQ as its backing store for persistent message storage
- RabbitMQ provides high availability and message durability
- The broker itself doesn't have queues - it's an event distribution mechanism
- Triggers subscribe to the broker and have their own queues for event processing

### 🎯 Trigger → Broker Connection

Triggers subscribe to the broker and have their own queues for event processing:

```yaml
# Static triggers for build events
apiVersion: eventing.knative.dev/v1
kind: Trigger
spec:
  broker: {app-name}-lambda-broker-{environment}  # Subscribe to broker
  filter:
    attributes:
      type: "network.notifi.lambda.build.start"   # Filter events
  subscriber:
    ref:
      kind: Service
      name: {builder-name}-{environment}          # Route to service
```

**How Triggers Work:**
- Triggers subscribe to the broker and create their own queues
- Each trigger filters events based on attributes (type, source, subject)
- Events matching the filter are delivered to the trigger's queue
- The trigger routes events from its queue to the subscriber service

### 🚀 Dynamic Trigger Creation

When a lambda function is built, the system creates a **new trigger** for that specific function:

```go
// Each lambda function gets its own trigger
func CreateTriggerResource(serviceName string, completionData *BuildCompletionEventData) *unstructured.Unstructured {
    return &unstructured.Unstructured{
        Object: map[string]interface{}{
            "apiVersion": "eventing.knative.dev/v1",
            "kind":       "Trigger",
            "spec": map[string]interface{}{
                "broker": "knative-lambda-service-broker-{environment}",
                "filter": map[string]interface{}{
                    "attributes": map[string]interface{}{
                        "type":    "network.notifi.lambda.parser.start",
                        "source":  fmt.Sprintf("network.notifi.%s", completionData.ThirdPartyID),
                        "subject": fmt.Sprintf("%s", completionData.ParserID),
                    },
                },
                "subscriber": map[string]interface{}{
                    "ref": map[string]interface{}{
                        "kind": "Service",
                        "name": serviceName, // Points to the lambda function
                    },
                },
            },
        },
    }
}
```

### 🔄 Complete Event Flow

1. **RabbitMQ Cluster** - Provides persistent message infrastructure
2. **RabbitmqBrokerConfig** - Tells the broker how to connect to RabbitMQ
3. **Broker** - Event distribution mechanism (no queues)
4. **Static Triggers** - Subscribe to broker, have their own queues, handle build events
5. **Dynamic Triggers** - Subscribe to broker, have their own queues, filter by event type/source/subject
6. **Lambda Functions** - Receive events via their individual trigger queues

### 🎯 Event Routing Logic

- **Build Events**: Static triggers route to the builder service
- **Parser Events**: Dynamic triggers route to specific lambda functions
- **Filtering**: Events are filtered by `type`, `source`, and `subject` attributes
- **Scaling**: Each trigger can scale independently based on event volume

### 📊 Data Flow Types

#### 🔨 Build Flow
```
Build Request → RabbitMQ → Broker → Static Triggers → Builder Service → Lambda Creation
```

#### 🚀 Execution Flow
```
Parser Event → RabbitMQ → Broker → Dynamic Triggers → Lambda Function → Response
```

#### 📈 Monitoring Flow
```
All Components → Prometheus/Tempo/Loki → Dashboards & Alerts
```

This architecture ensures:
- ✅ **High Availability**: RabbitMQ provides message persistence
- ✅ **Scalability**: Independent scaling of triggers and functions
- ✅ **Isolation**: Each lambda function has its own event routing
- ✅ **Reliability**: Dead letter queues and retry mechanisms
- ✅ **Observability**: Full event tracing and monitoring

## 🚀 Quick Start

### 📋 Prerequisites

- ☸️ **Kubernetes cluster** with Knative Serving installed
- 🏗️ **Pulumi CLI** installed and configured
- 🔄 **ArgoCD** installed and configured
- ☁️ **AWS credentials** configured
- 📡 **Knative Eventing broker** configured
- 📊 **Prometheus and Alertmanager** deployed

### 🚀 Deployment

1. **📥 Clone the repository**:
   ```bash
   git clone https://github.com/notifi-network/notifi-infra.git
   cd notifi-infra/20-platform/services/knative-lambda
   ```

2. **⚙️ Configure environment**:
   ```bash
   # Set environment variables from values.yaml
   export ENVIRONMENT=prd
   export AWS_REGION=us-west-2
   export AWS_ACCOUNT_ID=339954290315
   export ECR_BASE_REGISTRY=339954290315.dkr.ecr.us-west-2.amazonaws.com
   export BROKER_URL=broker-ingress.knative-eventing.svc.cluster.local
   export BROKER_PORT=80
   export BROKER_TOPIC=parser-results
   export TRACING_ENABLED=true
   export OTEL_EXPORTER_OTLP_ENDPOINT=tempo-distributor.tempo.svc.cluster.local:4317
   export NAMESPACE=knative-lambda-prd
   export LOG_LEVEL=info
   export METRICS_ENABLED=true
   export RATE_LIMITING_ENABLED=true
   ```

3. **🏗️ Deploy the service**:
   ```bash
   # Deploy using Pulumi and ArgoCD
   # The service is deployed via GitOps using ArgoCD applications
   # managed by Pulumi infrastructure as code
   
   # Deploy infrastructure using Pulumi
   cd pulumi/
   pulumi stack select ${ENVIRONMENT}
   pulumi up --yes
   
   # Verify ArgoCD application deployment
   kubectl get applications -n argocd
   kubectl get applicationsets -n argocd
   
   # Check service deployment status
   kubectl get pods -n knative-lambda-${ENVIRONMENT}
   ```

4. **✅ Verify deployment**:
   ```bash
   # Check service status
   kubectl get pods -n knative-lambda-${ENVIRONMENT}
   
   # 📊 Access Grafana Dashboard
   kubectl port-forward -n grafana svc/grafana 3000:80 &
   # Open browser: http://localhost:3000
   # Default credentials: admin/admin
   # Dashboard: Knative Lambda Builder - Overview
   
   # 📈 Access Prometheus Metrics
   kubectl port-forward -n prometheus svc/prometheus-operated 9090:9090 &
   # Open browser: http://localhost:9090
   
   # 🔍 Access Tempo Tracing
   kubectl port-forward -n tempo svc/tempo 3200:3200 &
   # Open browser: http://localhost:3200
   
   # 📝 Access Loki Logs
   kubectl port-forward -n loki svc/loki 3100:3100 &
   # Open browser: http://localhost:3100
   ```

## 📊 Monitoring & Alerting

### 📈 Metrics

The service exposes comprehensive metrics at `/metrics` endpoint:

- 🌐 **HTTP Metrics**: Request duration, status codes, request/response sizes
- 🏗️ **Build Metrics**: Queue depth, build phases, success/failure rates
- 💻 **Resource Metrics**: CPU, memory, goroutines, file descriptors
- 💰 **Business Metrics**: Throughput, costs, SLA compliance
- 🛡️ **Resilience Metrics**: Circuit breaker state, rate limiter status

### 🚨 Alerting

The service includes comprehensive alerting with PrometheusRules and Slack integration:

#### 🎯 Alert Categories

1. **🏆 Golden Signals** (SRE Best Practices)
   - ✅ Service availability
   - ❌ Error rates
   - ⏱️ Latency (95th percentile)
   - 💾 Resource saturation

2. **🏗️ Build Process Alerts**
   - 📋 Queue depth monitoring
   - 🚫 Build failure rates
   - 🐌 Build performance issues

3. **💻 Resource Utilization Alerts**
   - 🔥 CPU and memory usage
   - 🖥️ System resource monitoring
   - 🔗 Connection pool health

4. **💰 Business Metrics Alerts**
   - 📈 Build throughput
   - 💸 Cost monitoring
   - 📊 SLA violations

5. **🛡️ Resilience Pattern Alerts**
   - ⚡ Circuit breaker state
   - 🚦 Rate limiter activity
   - 🔌 Dependency health

#### Slack Integration

Alerts are automatically routed to Slack channels based on severity:

- **Critical Alerts**: Immediate notification to on-call engineers
- **Warning Alerts**: Team awareness notifications
- **Info Alerts**: High activity notifications

Each alert includes:
- Clear description of the issue
- Direct links to runbook, dashboard, and query
- Severity-based color coding
- Quick action buttons

### Testing Alerting Configuration

Use the provided test script to validate the alerting setup:

```bash
# Run the test script
./scripts/test-alerts.sh

# Or with custom environment
ENVIRONMENT=dev ./scripts/test-alerts.sh
```

## Configuration

### Environment Variables

```bash
# Service Configuration
HTTP_PORT=8080
METRICS_PORT=9090
LOG_LEVEL=info
NAMESPACE=knative-lambda-prd
ENVIRONMENT=prd

# AWS Configuration
AWS_REGION=us-west-2
AWS_ACCOUNT_ID=339954290315
ECR_BASE_REGISTRY=339954290315.dkr.ecr.us-west-2.amazonaws.com
S3_SOURCE_BUCKET=notifi-uw2-prd-fusion-modules
S3_TMP_BUCKET=knative-lambda-prd-context-tmp

# EKS Pod Identity Configuration
USE_EKS_POD_IDENTITY=true
POD_IDENTITY_ROLE=knative-lambda-builder

# Storage Configuration
STORAGE_PROVIDER=aws-s3  # Options: aws-s3, minio (planned)
S3_ENDPOINT=https://s3.us-west-2.amazonaws.com  # AWS S3 endpoint
# MINIO_ENDPOINT=http://minio-service.minio.svc.cluster.local:9000  # MinIO endpoint (planned)
# MINIO_ACCESS_KEY=your-access-key  # MinIO access key (planned)
# MINIO_SECRET_KEY=your-secret-key  # MinIO secret key (planned)

# CloudEvents Broker Configuration
BROKER_TOPIC=parser-results
DEFAULT_BROKER_NAME=knative-lambda-lambda-broker-prd

# Monitoring Configuration
TRACING_ENABLED=true
OTEL_EXPORTER_OTLP_ENDPOINT=tempo-distributor.tempo.svc.cluster.local:4317
METRICS_ENABLED=true
RATE_LIMITING_ENABLED=true

# Lambda Function Configuration
# TODO: Add configuration per tenant
LAMBDA_DEFAULT_RUNTIME=nodejs22
LAMBDA_DEFAULT_HANDLER=index.handler
LAMBDA_FUNCTION_MEMORY_LIMIT=512Mi
LAMBDA_FUNCTION_CPU_LIMIT=1000m
LAMBDA_FUNCTION_MEMORY_REQUEST=128Mi
LAMBDA_FUNCTION_CPU_REQUEST=100m
# 🔄 Lambda Worker Pool Configuration
# Think of this as a team of workers handling your function builds!
LAMBDA_WORKER_POOL_SIZE=5          # 🧑‍💻 Number of workers in the pool
LAMBDA_WORKER_POOL_CAPACITY=10     # 📦 How many tasks each worker can handle
LAMBDA_EVENT_QUEUE_SIZE=50         # 📋 Size of the waiting list for new tasks

# Kaniko Build Configuration (for container image building)
KANIKO_ENABLED=true
KANIKO_IMAGE=gcr.io/kaniko-project/executor:v1.19.2
KANIKO_TIMEOUT=30m
KANIKO_CACHE=true
KANIKO_CACHE_TTL=24h
KANIKO_SKIP_TLS_VERIFY=false
KANIKO_INSECURE_REGISTRY=false

# Kaniko NPM Configuration (for JavaScript function builds)
KANIKO_NPM_REGISTRY=https://registry.npmjs.org/
KANIKO_NPM_TIMEOUT=60000
KANIKO_NPM_FETCH_RETRIES=5
KANIKO_NPM_FETCH_RETRY_MINTIMEOUT=10000
KANIKO_NPM_FETCH_RETRY_MAXTIMEOUT=60000
KANIKO_NPM_FETCH_RETRY_FACTOR=2
KANIKO_NPM_PREFER_OFFLINE=true
KANIKO_NPM_AUDIT=false
KANIKO_NPM_FUND=false
KANIKO_NPM_UPDATE_NOTIFIER=false
KANIKO_NPM_LOGLEVEL=warn

# Kaniko Registry Configuration
KANIKO_REGISTRY_MIRROR=""
KANIKO_DOCKER_CONFIG=""
KANIKO_IMAGE_PULL_SECRET=""

# Kaniko Build Context Configuration
KANIKO_BUILD_CONTEXT_SIZE_LIMIT=2Gi
KANIKO_BUILD_CONTEXT_TIMEOUT=10m
KANIKO_BUILD_ARGS=""
KANIKO_BUILD_TARGET=""
KANIKO_DESTINATION=""

# Kaniko Security Configuration
KANIKO_RUN_AS_USER=1000
KANIKO_RUN_AS_GROUP=1000
KANIKO_SECURITY_CONTEXT_RUN_AS_NON_ROOT=true
KANIKO_SECURITY_CONTEXT_READ_ONLY_ROOT_FILESYSTEM=true


```
### Scaling Configuration

```bash
# Knative Autoscaling Configuration
KUBERNETES_TARGET_CONCURRENCY=5      # Reduced for faster scaling
KUBERNETES_TARGET_UTILIZATION=70     # Increased for better resource utilization
KUBERNETES_TARGET=0.01               # Very aggressive scaling - forces multiple pods
KUBERNETES_CONTAINER_CONCURRENCY=0   # Unlimited concurrency - forces load balancing
KUBERNETES_MIN_SCALE=0
KUBERNETES_MAX_SCALE=10              # Production limit
KUBERNETES_SCALE_TO_ZERO_GRACE_PERIOD=2m  # Reduced for faster scale down
KUBERNETES_SCALE_DOWN_DELAY=30s          # Reduced for faster scale down
KUBERNETES_STABLE_WINDOW=30s            # Reduced for faster response

# Builder Service Scaling
LAMBDA_BUILDER_MIN_SCALE=0 
LAMBDA_BUILDER_MAX_SCALE=10
LAMBDA_BUILDER_TARGET_CONCURRENCY=5
LAMBDA_BUILDER_SCALE_TO_ZERO_GRACE_PERIOD=2m
LAMBDA_BUILDER_SCALE_DOWN_DELAY=30s
LAMBDA_BUILDER_STABLE_WINDOW=30s

# HPA Configuration (Production)
LAMBDA_HPA_ENABLED=true
LAMBDA_HPA_MIN_REPLICAS=3
LAMBDA_HPA_MAX_REPLICAS=20
LAMBDA_HPA_TARGET_CPU_UTILIZATION_PERCENTAGE=80
LAMBDA_HPA_TARGET_MEMORY_UTILIZATION_PERCENTAGE=80
```

### Language Support

The Knative Lambda Builder currently supports **JavaScript/Node.js** functions with comprehensive build optimization and dependency management.

#### Current Support
- **JavaScript/Node.js**: Full support with npm dependency resolution, build optimization, and runtime configuration
- **Build Process**: Automated dependency installation, code bundling, and container image creation
- **Runtime**: Optimized Node.js runtime with performance monitoring and resource management

#### Planned Language Support
- **Go**: Native Go function support with module management and binary optimization
- **Python**: Python function support with pip dependency management and virtual environment handling

#### Function Requirements
- **JavaScript**: Requires `package.json` with proper dependencies and entry point
- **Go**: Will require `go.mod` and `go.sum` for dependency management
- **Python**: Will require `requirements.txt` or `pyproject.toml` for dependency management

Each language will include:
- Automated dependency resolution
- Build optimization and caching
- Runtime-specific monitoring and metrics
- Resource allocation optimization

### Storage Options

The Knative Lambda Builder supports multiple storage providers for build artifacts and temporary files:

#### Current Storage Support
- **AWS S3**: Full integration with Amazon S3 for build artifact storage
  - Source code buckets for function source files
  - Temporary buckets for build context and intermediate files
  - ECR integration for container image storage

#### Planned Storage Support
- **MinIO**: On-premises S3-compatible object storage
  - Self-hosted object storage solution
  - S3-compatible API for seamless integration
  - Local/private cloud deployment options
  - Cost-effective alternative to cloud storage

#### Storage Features
- **Multi-provider support**: Configurable storage backends
- **S3-compatible API**: Consistent interface across providers
- **Bucket management**: Automated bucket creation and lifecycle policies
- **Security**: IAM roles, access keys, and encryption support
- **Monitoring**: Storage usage metrics and cost tracking

## Documentation

### 📚 Core Documentation
- **[Alerting](ALERTING.md)**: Comprehensive alerting setup guide
- **[Runbook](RUNBOOK.md)**: Troubleshooting procedures for all alerts
- **[Metrics](METRICS.md)**: Detailed metrics reference
- **[Deployment](DEPLOYMENT.md)**: Step-by-step deployment instructions
- **[Improvement Plan](IMPROVEMENT_PLAN.md)**: Strategic roadmap and priorities
- **[Review Guide](REVIEW_GUIDE.md)**: Comprehensive file-by-file review guide

### 🔍 Code Review Documents
- **[AI SRE Engineer Review](AI_SRE_ENGINEER_REVIEW.md)**: Observability, reliability, and operational excellence
- **[AI Cloud Architect Review](AI_CLOUD_ARCHITECT_REVIEW.md)**: Architecture, design patterns, and scalability
- **[AI DevOps Engineer Review](AI_DEVOPS_ENGINEER_REVIEW.md)**: CI/CD, infrastructure as code, and deployment
- **[AI Senior Golang Engineer Review](AI_SENIOR_GOLANG_ENGINEER_REVIEW.md)**: Code quality, Go best practices, and testing
- **[Bruno's Review](BRUNO_REVIEW.md)**: Strategic direction and final approval

### 🔐 Security Reviews
- **[Senior Pentester Review](SENIOR_PENTESTER_REVIEW.md)**: Defensive security and vulnerability assessment
- **[Blackhat Review](BLACKHAT_REVIEW.md)**: Offensive security and exploit development
- **[Security Review Summary](SECURITY_REVIEW_SUMMARY.md)**: Consolidated security findings and remediation plan

### 📖 Additional Documentation
- **[Introduction](INTRO.md)**: Project introduction and getting started
- **[Validation Guide](VALIDATION.md)**: Code quality and validation standards
- **[Summary](SUMMARY.md)**: Project summary and status
- **[TODO](TODO.md)**: Current tasks and future improvements
- **[Changelog](CHANGELOG.md)**: Version history and changes

### 📂 Documentation in `/docs`
- **[Versioning Strategy](docs/VERSIONING_STRATEGY.md)**: Semantic versioning and release management
- **[Branching Guide](docs/BRANCHING_GUIDE.md)**: Git workflow and branch strategy
- **[Quick Start Versioning](docs/QUICK_START_VERSIONING.md)**: Quick reference for version management
- **[MinIO Setup](docs/MINIO_SETUP.md)**: MinIO configuration guide
- **[Job Start Events](docs/JOB_START_EVENTS.md)**: Build job event documentation

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests and documentation
5. Submit a pull request

## Support

- **On-Call**: Check PagerDuty for current engineer
- **Slack**: #infra-alerts-{environment}
- **Documentation**: This repository
- **Issues**: GitHub issues for bug reports

## License

This project is licensed under the MIT License - see the LICENSE file for details. # Test dev workflow - Thu Aug 14 15:20:53 -03 2025
