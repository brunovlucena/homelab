# 🚀 Knative Lambda Operator

**Kubernetes operator for building and deploying containerized serverless functions on Knative**

[![Version](https://img.shields.io/badge/version-1.0.4-blue.svg)](https://github.com/brunovlucena/homelab)
[![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

---

## 📋 Table of Contents

- [Overview](#-overview)
- [Key Features](#-key-features)
- [Architecture](#-architecture)
- [GitOps Deployment](#-gitops-deployment)
- [Testing in pro before merging to main](#-testing-in-pro-before-merging-to-main)
- [Quick Start](#-quick-start)
- [Testing](#-testing)
- [Monitoring & Observability](#-monitoring--observability)

---

## 🎯 Overview

**Knative Lambda Operator** is a Kubernetes operator that enables dynamic function-as-a-service (FaaS) and AI agent deployments on Knative Serving. The operator manages two custom resource types:

### 🚀 LambdaFunction
For serverless functions that need to be built from source code. The operator automatically:
- Builds container images with Kaniko from code stored in S3/MinIO
- Deploys Knative Services with auto-scaling
- Configures CloudEvents triggers and RabbitMQ integration
- Monitors function health and metrics

### 🤖 LambdaAgent
For AI agents with pre-built Docker images. The operator manages:
- Direct deployment of pre-built container images (no build pipeline)
- First-class AI/LLM configuration (model, endpoint, temperature, system prompts)
- Intent-based event routing for agent-to-agent communication
- Agent-specific scaling (keeps agents warm, optimized for conversational workloads)
- AI-specific observability (token usage, inference latency, model metrics)
- CloudEvents integration with dedicated brokers and triggers per agent

Both resource types share the operator's eventing infrastructure (RabbitMQ brokers, triggers, DLQ) and observability stack (Prometheus, Grafana, distributed tracing).

### What Problems Does It Solve?

- ✅ **Eliminates manual container builds** - Upload code, get a running function automatically (LambdaFunction)
- ✅ **AI agent deployment simplified** - Deploy pre-built AI agents with first-class LLM configuration (LambdaAgent)
- ✅ **Zero infrastructure management** - Functions and agents scale automatically (0→N for functions, warm for agents)
- ✅ **Cost optimization** - Only pay for compute when functions/agents are actively processing requests
- ✅ **Event-driven architecture** - Native CloudEvents integration with RabbitMQ for both functions and agents
- ✅ **Intent-based agent communication** - Agents automatically route events based on detected intents
- ✅ **Enterprise-grade observability** - Built-in metrics, logging, and distributed tracing for both resource types

---

## ✨ Key Features

### 🚀 LambdaFunction Features

- **Kaniko-based container builds** - Secure, in-cluster builds without Docker daemon
- **S3/MinIO source code storage** - Upload Python/Node.js/Go code, get a running function
- **Multi-language support** - Python, Node.js, Go, with extensible template system
- **Inline code support** - Deploy functions directly from YAML without external storage

### 🤖 LambdaAgent Features

- **Pre-built image deployment** - Deploy AI agents from existing Docker images (no build pipeline)
- **First-class AI configuration** - LLM provider, model, temperature, system prompts
- **Intent-based routing** - Automatic event routing based on detected intents
- **Agent-specific scaling** - Keep agents warm with minimum replicas, optimized for conversations
- **AI observability** - Token usage, inference latency, model health metrics

### ⚡ Auto-Scaling & Performance

- **Scale-to-zero** - Inactive functions consume zero resources
- **Rapid scale-up** - 0→N in <30s with optimized cold start (<5s)
- **Resource optimization** - Configurable CPU/memory limits per function

### 🔄 GitOps & Progressive Delivery

- **Flux CD integration** - Automated GitOps deployments
- **Flagger canary deployments** - Progressive delivery with metrics validation
- **Linkerd service mesh** - mTLS, traffic splitting, observability
- **A/B testing** - Header-based traffic routing for testing

### 📊 Full Observability

- **Prometheus metrics** - Build time, success rate, queue depth, resource usage
- **Grafana dashboards** - Pre-built dashboards for monitoring
- **Alerting** - Comprehensive alert rules for SRE teams

### 🌐 CloudEvents-Native

- **Standardized event format** - All communication via CloudEvents v1.0
- **RabbitMQ broker** - Event routing with triggers and filters
- **Lambda Runtime Wrapper** - Automatic CloudEvents request/response handling
- **DLQ support** - Failed events routed to Dead Letter Queue for retry

---

## 🚀 GitOps Deployment

### Directory Structure

```
k8s/
├── base/                    # 📦 Base resources (all environments)
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── crd.yaml              # LambdaFunction CRD
│   ├── crd-lambdaagent.yaml  # LambdaAgent CRD
│   ├── rbac.yaml
│   ├── agent-rbac.yaml       # LambdaAgent-specific RBAC
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── servicemonitor.yaml
│   ├── cloudevents-receiver.yaml  # CloudEvents webhook receiver
│   └── lambda-samples-*.yaml      # Sample functions
├── overlays/
│   ├── pro/                 # 🛠️ Development (pro cluster)
│   │   ├── kustomization.yaml
│   │   ├── lambdafunctions.yaml   # Sample functions
│   │   └── k6-configmap.yaml      # k6 test scripts
│   └── studio/              # 🚀 Production (studio cluster)
│       ├── kustomization.yaml
│       ├── canary.yaml            # Flagger canary config
│       ├── ab-testing.yaml        # A/B testing rules
│       ├── metrictemplates.yaml   # Prometheus metric templates
│       └── alertrules.yaml        # PrometheusRule alerts
└── tests/                   # 🧪 k6 load tests
    ├── k6-configmap.yaml
    ├── k6-testrun-smoke.yaml
    ├── k6-testrun-load.yaml
    └── k6-testrun-stress.yaml
```

### Environment Differences

| Feature | Pro / Pro-1-Node | Studio / Studio-1-Node |
|---------|------------------|------------------------|
| Replicas | 1 | 2+ (HA) |
| Canary Deployment | ❌ | ✅ (conservative) |
| A/B Testing | ❌ | ✅ |
| Linkerd Injection | ✅ | ✅ |
| Alert Rules | ❌ | ✅ |
| Sample Functions | ✅ | ❌ |
| k6 Load Tests | Available | On-demand |
| Port Range | 31000-31999 | 32000-32999 |

### Deployment Commands

```bash
# Deploy to development (pro cluster)
make deploy-pro

# Deploy to production (studio cluster - triggers canary)
make deploy-studio

# Show diff before deploying
make deploy-diff ENV=pro
make deploy-diff ENV=studio
```

---

## 🧪 Testing in pro before merging to main

Validate changes on **pro** (from your branch) before merging to **main**; **studio** stays on main until you merge.

1. **Point pro at your branch** (pro context):

   ```bash
   kubectl config use-context pro
   make flux-test-branch
   ```

2. **Ensure the operator image exists** for that branch (push so [CI](.github/workflows/operator-knative-lambda.yml) builds/pushes, or run the workflow manually).

3. **Test** on pro.

4. **Point pro back at main** when done:

   ```bash
   kubectl patch gitrepository homelab -n flux-system --type=merge -p '{"spec":{"ref":{"branch":"main"}}}'
   flux reconcile source git homelab
   flux reconcile kustomization pro-04-knative-instances --with-source
   ```

5. **Merge** to main so studio gets the operator from main.

---

## 🐦 Canary Deployments

### Studio (Production) Strategy (Conservative)

- **Step weight:** 5% increments
- **Max weight:** 30%
- **Interval:** 2 minutes between steps
- **Success threshold:** 99.5% success rate
- **Iterations:** 6 (full promotion)
- **Auto-rollback:** On failure

### Canary Commands

```bash
# Check canary status
make canary-status

# Manually promote canary (skip analysis)
make canary-promote

# Rollback canary
make canary-rollback

# Describe canary details
make canary-describe
```

---

## 🔀 A/B Testing (Production Only)

A/B testing is enabled via HTTP headers in production:

```bash
# Route to canary version
curl -H "x-ab-test: canary" http://knative-lambda-operator.knative-lambda/healthz

# Route to primary version
curl -H "x-ab-test: primary" http://knative-lambda-operator.knative-lambda/healthz

# Check traffic split
make ab-split-status
```

---

## 📊 k6 Load Testing

### Available Tests

| Test | Duration | Purpose |
|------|----------|---------|
| Smoke | 30s | Quick validation |
| Load | 13min | Normal traffic simulation |
| Stress | 24min | Find breaking points |

### Running Tests

```bash
# Apply test resources
make k6-apply-tests

# Run smoke test
make k6-smoke

# Run load test
make k6-load

# Run stress test (⚠️ high load)
make k6-stress

# Check test status
make k6-status

# Clean up tests
make k6-clean
```

---

## 📈 Monitoring

### Metrics Endpoint

```bash
# Port-forward and fetch metrics
make metrics
```

### Alert Rules

Production environment includes PrometheusRules for:
- `KnativeLambdaCanaryFailed` - Canary deployment failed
- `KnativeLambdaOperatorDown` - All replicas unavailable
- `KnativeLambdaHighErrorRate` - Error rate > 1%
- `KnativeLambdaBuildFailed` - Multiple build failures

### View Alerts

```bash
make alerts
```

---

## 🛠️ Quick Start

### Prerequisites

- Kubernetes cluster with:
  - Flux CD
  - Flagger
  - Linkerd
  - Prometheus + Grafana
  - k6-operator
  - Knative Serving + Eventing

### Deploy to Development

```bash
# Build and push operator image
make build-images-local

# Deploy to dev
make deploy-dev

# Verify operator deployment
kubectl get deployment -n knative-lambda knative-lambda-operator

# Verify CRDs are installed
kubectl get crd lambdafunctions.lambda.knative.io
kubectl get crd lambdaagents.lambda.knative.io

# Check LambdaFunctions
kubectl get lambdafunction -n knative-lambda

# Check LambdaAgents
kubectl get lambdaagent -A
```

### Deploy Sample LambdaFunction

```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaFunction
metadata:
  name: hello-python
  namespace: knative-lambda
spec:
  source:
    type: inline
    inline:
      code: |
        def handler(event, context):
            return {"message": "Hello from Lambda!"}
  runtime:
    language: python
    version: "3.11"
```

### Deploy Sample LambdaAgent

```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaAgent
metadata:
  name: agent-bruno
  namespace: agent-bruno
spec:
  image:
    repository: localhost:5001/agent-bruno/chatbot
    tag: "0.1.0"
  ai:
    provider: ollama
    endpoint: http://ollama.ollama.svc.cluster.local:11434
    model: llama3.2:3b
    temperature: 0.7
    systemPrompt: |
      You are Bruno's AI assistant.
      Be helpful, concise, and friendly.
  eventing:
    enabled: true
    subscriptions:
      - eventType: io.homelab.chat.message
  scaling:
    minReplicas: 1
    maxReplicas: 5
```

---

## 📁 Project Structure

```
knative-lambda-operator/
├── docs/                         # 📚 Documentation (79 files)
│   ├── 01-getting-started/      # Installation & first steps
│   ├── 02-for-executives/       # Business overview
│   ├── 03-for-engineers/        # Technical deep-dives
│   │   ├── backend/             # Backend user stories (12)
│   │   ├── devops/              # DevOps user stories (10)
│   │   ├── security/            # Security user stories (10)
│   │   ├── sre/                 # SRE user stories (14)
│   │   └── qa/                  # QA testing guides
│   ├── 04-architecture/         # System design docs
│   │   ├── CLOUDEVENTS_SPECIFICATION.md
│   │   ├── NOTIFI_INTEGRATION.md
│   │   ├── DLQ_FLOWS.md
│   │   └── OBSERVABILITY_SPECIFICATION.md
│   ├── 05-operations/           # Troubleshooting guides
│   ├── 06-development/          # Testing strategy
│   └── 07-decisions/            # Architecture decisions
├── k8s/                          # ☸️ Kubernetes manifests (GitOps)
│   ├── base/                    # Base resources
│   ├── overlays/                # Environment overlays
│   │   ├── pro/                # Development (pro cluster)
│   │   └── studio/             # Production (studio cluster)
│   └── tests/                   # k6 load tests
├── src/
│   ├── operator/                # 🎮 Kubernetes operator (Go)
│   │   ├── api/v1alpha1/       # CRD types (LambdaFunction, LambdaAgent)
│   │   ├── controllers/        # Reconciliation logic
│   │   │   ├── lambdafunction_controller.go  # LambdaFunction controller
│   │   │   └── lambdaagent_controller.go      # LambdaAgent controller
│   │   └── internal/           # Build, deploy, events, metrics
│   └── tests/                   # 🧪 Comprehensive test suites
│       ├── unit/               # Unit tests (backend, devops, security, sre)
│       ├── integration/        # Integration tests
│       ├── e2e/                # End-to-end tests
│       └── load/               # k6 + Python load tests
├── Makefile                     # 🔧 Build & deploy automation (1000+ lines)
├── README.md                    # 📖 This file
└── VERSION                      # 📌 Version file (current: 1.0.4)
```

---

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [CloudEvents Specification](docs/04-architecture/CLOUDEVENTS_SPECIFICATION.md) | Event types, schemas, and flows |
| [Notifi Integration](docs/04-architecture/NOTIFI_INTEGRATION.md) | Notifi Fusion platform integration |

---

## 🤝 Contributing

1. Make changes to operator code in `src/operator/`
2. Update version in `VERSION`
3. Build and push operator image: `make build-images-local`
4. Deploy operator to dev: `make deploy-dev`
5. Run tests: `make test && make k6-smoke`
6. Commit and push - Flux will reconcile the operator

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.
