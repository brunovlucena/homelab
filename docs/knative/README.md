# Knative Lambda Architecture Documentation

## 📊 Architecture Diagrams

### Knative Lambda Platform Architecture

**File:** `knative-lambda-architecture.png`

**Description:** Comprehensive architecture diagram showing the complete Knative Lambda serverless function platform, including:

- **Edge Layer:** Nginx reverse proxy, Traefik ingress
- **Builder Service:** Go-based Knative service for orchestrating function builds
- **Kaniko Build System:** Secure, in-cluster container builds without Docker daemon
- **Event-Driven Architecture:** RabbitMQ-backed Knative Eventing with brokers, triggers, and sources
- **Lambda Functions:** Dynamically created Knative services with scale-to-zero capabilities
- **Dead Letter Queue:** RabbitMQ-based DLQ with retry logic and error categorization
- **Rate Limiting:** Multi-level rate limiting (build context, K8s jobs, client, S3)
- **Observability Stack:** Prometheus, Grafana, Loki, Tempo, and Alloy (OTel Collector)
- **Notifi Backend Integration:** Scheduler, Subscription Manager, Storage Manager, Blockchain Manager
- **Storage & Registry:** S3 source/temp buckets, ECR production, Harbor local registry
- **GitOps:** Flux CD for automated deployments
- **Security:** RBAC, Pod Security, TLS/mTLS, Sealed Secrets

### Component Overview

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Builder Service** | Go 1.24, Knative Serving | Main orchestrator for builds and deployments |
| **Kaniko Jobs** | Kaniko v1.19.2 | Secure container image builds |
| **Metrics Pusher** | Go sidecar | Prometheus remote write for metrics |
| **Sidecar Monitor** | Go sidecar | Build progress tracking |
| **RabbitMQ Cluster** | RabbitMQ Cluster Operator | Event bus for CloudEvents (3-node cluster) |
| **Knative Eventing** | Knative + RabbitMQ Backend | Event routing and filtering |
| **Lambda Functions** | Python/Node.js/Go | User-defined serverless functions |
| **DLQ Handler** | Deployment | Dead letter queue processing with retry logic |
| **S3 Source Bucket** | AWS S3 | Parser code storage (notifi-uw2-dev-fusion-modules) |
| **S3 Temp Bucket** | AWS S3 | Build context cache (knative-lambda-dev-context-tmp) |
| **ECR Production** | AWS ECR | Lambda container images (339954290315.dkr.ecr.us-west-2) |
| **Harbor Registry** | Harbor | Dev/base images (localhost:5001) |
| **Prometheus** | Prometheus Operator | Metrics collection and alerting |
| **Grafana** | Grafana | Dashboards and visualization |
| **Loki** | Grafana Loki | Log aggregation |
| **Tempo** | Grafana Tempo | Distributed tracing |
| **Alloy** | OpenTelemetry Collector | Unified telemetry collection |
| **Flux CD** | Flux | GitOps continuous delivery |
| **Linkerd2** | Linkerd | Service mesh with mTLS |

## 🚀 Key Features

### 1. Dynamic Function Building

- **Kaniko-based builds:** Secure, in-cluster container builds without Docker daemon
- **Multi-language support:** Python, Node.js, Go with extensible template system
- **Automatic dependency resolution:** `requirements.txt`, `package.json`, `go.mod`
- **Build context caching:** S3 temp bucket for faster builds
- **Sidecar monitoring:** Real-time build progress tracking

### 2. Auto-Scaling & Performance

- **Scale-to-zero:** Inactive functions consume zero resources
- **Rapid scale-up:** 0→N in <30s with optimized cold start (<5s)
- **Concurrency-based:** Knative KPA (Knative Pod Autoscaler)
- **Burst handling:** Max 50 replicas per function
- **Resource optimization:** Configurable CPU/memory per function

### 3. Event-Driven Architecture

- **CloudEvents native:** Standards-based event processing
- **RabbitMQ backend:** Quorum Queues for HA
- **Knative Brokers:** Event routing and filtering
- **Knative Triggers:** Function-specific event subscriptions
- **Multiple event types:** `build.start`, `job.start`, `service.delete`, custom
- **Event sources:** APIServerSource (K8s events), RabbitMQSource (external)

### 4. Full Observability

- **Prometheus metrics:** `build_duration_seconds`, `build_success_rate`, `queue_depth`
- **OpenTelemetry tracing:** End-to-end distributed tracing
- **Structured logging:** JSON logs with context propagation
- **Grafana dashboards:** Pre-built monitoring dashboards
- **Metrics Pusher sidecar:** Prometheus remote write
- **Alloy (OTel Collector):** Unified telemetry collection

### 5. Enterprise Security

- **RBAC:** ClusterRole with fine-grained permissions
- **Pod Security:** `runAsNonRoot: true`, `readOnlyRootFilesystem: true`
- **TLS/mTLS:** Linkerd2 service mesh with automatic mTLS
- **Cert Manager:** Automated TLS certificate management
- **Sealed Secrets:** Encrypted secrets in Git
- **Rate limiting:** Multi-level (build context, K8s jobs, client, S3)
- **EKS Pod Identity:** AWS IAM role for service account

### 6. Dead Letter Queue (DLQ)

- **RabbitMQ-based:** Exponential backoff retry logic
- **7-day retention:** 604800000ms message TTL
- **Max 50,000 messages:** Drop-head overflow policy
- **5 retry attempts:** Exponential backoff
- **Error categorization:** Transient vs permanent failures
- **Automated cleanup:** 24h interval CronJob

## 📈 Key Metrics & Scaling

| Component | Replicas | CPU | Memory | Scaling |
|-----------|----------|-----|--------|---------|
| **Builder Service** | 0-10 | 250m | 256Mi | Concurrency-based |
| **Lambda Functions** | 0-50 | 50m | 64Mi | Scale-to-zero |
| **Kaniko Jobs** | Dynamic | 500m-1000m | 1Gi-2Gi | On-demand |
| **DLQ Handler** | 1 | 100m | 128Mi | Static |
| **Metrics Pusher** | Sidecar | 50m | 64Mi | Sidecar |
| **Sidecar Monitor** | Sidecar | 100m | 128Mi | Sidecar |

### Performance Characteristics

- **Cold Start:** <5s (optimized base images)
- **Build Time:** ~30-90s (cached dependencies)
- **Scale-to-Zero Grace Period:** 30s
- **Event Processing:** 50 parallel RabbitMQ consumers
- **Max Concurrent Builds:** 10
- **Max Concurrent Jobs:** 5

## 🎯 Use Cases

1. **⚡ Webhook Processors**
   - Scale from 0 when webhooks arrive
   - Handle burst traffic automatically
   - Cost-effective for sporadic traffic

2. **🔄 Data Processing Pipelines**
   - Transform and enrich data from multiple sources
   - Parallel processing with auto-scaling
   - S3-triggered data transformations

3. **🌐 API Integrations**
   - Connect to 3rd-party services dynamically
   - No dedicated servers required
   - Pay only for actual processing time

4. **📊 Background Jobs**
   - Scheduled or triggered tasks
   - Auto-scaling based on queue depth
   - Efficient resource utilization

5. **🔨 Event Handlers**
   - Process events from RabbitMQ/Kafka
   - CloudEvents-native event processing
   - Async processing with DLQ support

## 🔧 Generating the Diagram

### Prerequisites

```bash
# Install Python 3 and pip
brew install python3  # macOS
apt-get install python3 python3-pip  # Ubuntu/Debian

# Install diagrams package
pip3 install diagrams
```

### Generate Diagram

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/infrastructure/knative-lambda/docs
python3 knative-lambda-architecture.py
```

The diagram will be saved as `knative-lambda-architecture.png` in the current directory.

## 📚 Related Documentation

- **Main README:** `../README.md` - Platform overview and quick start guide
- **Helm Chart:** `../k8s/chart/` - Kubernetes deployment manifests
- **Source Code:** `../src/builder/` - Builder service implementation
- **Tests:** `../src/tests/` - Unit, integration, and e2e tests

## 🔗 Integration with Notifi Backend

The Knative Lambda platform integrates with the following Notifi backend services:

- **Scheduler:** `notifi-scheduler.notifi.svc.cluster.local` - Fusion execution callbacks
- **Subscription Manager:** `notifi-subscription-manager.notifi.svc.cluster.local:4000` - gRPC
- **Storage Manager:** `notifi-storage-manager.notifi.svc.cluster.local:4000` - gRPC
- **Fetch Proxy:** `notifi-fetch-proxy.notifi.svc.cluster.local:4000` - HTTP proxy
- **Blockchain Manager:** `notifi-blockchain-manager.notifi.svc.cluster.local:4000` - Multi-chain RPC

## 🚀 Deployment Guide

### 1. Build Images

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/infrastructure/knative-lambda
make build-images-local  # Builds builder, sidecar, metrics-pusher images
```

### 2. Deploy to Cluster

```bash
# Deploy via Flux GitOps
kubectl apply -k k8s/kustomize/studio/
flux reconcile helmrelease knative-lambda-studio -n knative-lambda-dev

# Or deploy directly with Helm
helm upgrade --install knative-lambda-dev k8s/chart/ \
  --namespace knative-lambda-dev \
  --create-namespace \
  --values k8s/kustomize/studio/values.yaml
```

### 3. Test with CloudEvent

```bash
curl -X POST http://knative-lambda-builder-dev.knative-lambda-dev.svc.cluster.local:8080/cloudevents \
  -H "Content-Type: application/json" \
  -H "Ce-Id: test-123" \
  -H "Ce-Source: test-cli" \
  -H "Ce-Type: network.notifi.lambda.parser.start" \
  -H "Ce-Specversion: 1.0" \
  -d '{
    "third_party_id": "test",
    "parser_id": "parser-123",
    "s3_key": "parsers/test.js"
  }'
```

## 📊 Monitoring & Observability

### Access Dashboards

```bash
# Port-forward to Grafana
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-grafana 3000:80

# Open in browser
open http://localhost:3000
```

### Key Metrics

- **Build Metrics:** `knative_lambda_build_duration_seconds`, `knative_lambda_build_success_rate`
- **Queue Metrics:** `knative_lambda_queue_depth`, `knative_lambda_event_processing_time_seconds`
- **Lambda Metrics:** `knative_lambda_cold_start_duration_seconds`, `knative_lambda_request_rate`
- **DLQ Metrics:** `knative_lambda_dlq_depth`, `knative_lambda_dlq_message_age_seconds`

### Logs

```bash
# Builder service logs
kubectl logs -n knative-lambda-dev -l serving.knative.dev/service=knative-lambda-builder-dev

# Kaniko job logs
kubectl logs -n knative-lambda-dev <kaniko-job-pod-name> -c kaniko

# DLQ handler logs
kubectl logs -n knative-lambda-dev -l app=knative-lambda-dlq-handler
```

## 🔍 Troubleshooting

### Build Failures

1. Check builder service logs
2. Verify S3 bucket access (IAM role, pod identity)
3. Check Kaniko job status
4. Review build context size limits
5. Verify base image availability

### Scaling Issues

1. Check Knative autoscaler metrics
2. Verify resource quotas
3. Review concurrency settings
4. Check for pod evictions
5. Verify metrics server availability

### Event Processing Issues

1. Check RabbitMQ cluster health
2. Verify broker and trigger configuration
3. Review DLQ for failed events
4. Check event source connectivity
5. Verify CloudEvents format

## 📝 License

This documentation and architecture diagram are part of the Notifi Homelab project.

---

**Last Updated:** 2025-11-27  
**Diagram Version:** 1.0.0  
**Author:** Bruno Lucena
