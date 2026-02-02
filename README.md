# 🏗️ Homelab

> **Run AI Agents Locally on Kubernetes**  
> **Last Updated**: December 4, 2025  
> **SRE**: Bruno Lucena | **IaC**: Pulumi

---

## ⚡ TL;DR

| Project | One Liner |
|---------|-----------|
| **🏠 Homelab** | Your own mini-AWS running on a Mac Studio |
| **⚡ Knative-Lambda** | Upload code → it just runs (like AWS Lambda but yours) |
| **🛡️ Agent-Contracts** | AI that hunts crypto hacks before hackers do |

### How They Connect

```
┌─────────────────────────────────────────────┐
│              🏠 HOMELAB                      │
│         (the infrastructure)                │
│                                             │
│   ┌─────────────────────────────────────┐   │
│   │        ⚡ Knative-Lambda             │   │
│   │     (runs serverless functions)     │   │
│   └─────────────────────────────────────┘   │
│                    │                        │
│           runs on top of                    │
│                    ▼                        │
│   ┌─────────────────────────────────────┐   │
│   │       🛡️ Agent-Contracts             │   │
│   │   (AI security agent = 4 functions) │   │
│   └─────────────────────────────────────┘   │
│                                             │
└─────────────────────────────────────────────┘
```

### Why It's Cool

- ✅ **No AWS bill** — runs on hardware you own
- ✅ **Scale to zero** — functions sleep, no wasted compute
- ✅ **GitOps** — push to Git, stuff auto-deploys
- ✅ **Full observability** — dashboards, logs, alerts
- ✅ **Multi-cluster** — 5 clusters talking to each other

---

## 🎯 What is This?

A Kubernetes-based platform for building and running AI agents locally. Instead of deploying to cloud FaaS providers, this homelab lets you run serverless AI workloads on your own hardware with enterprise-grade features: auto-scaling, observability, and GitOps deployments.

**Two main projects power this platform:**

| Project | Purpose |
|---------|---------|
| **🚀 Knative Lambda** | Serverless FaaS platform — upload code, get a running function |
| **🛡️ Agent-Contracts** | AI-powered smart contract security agent |

Everything else (Kubernetes clusters, service mesh, observability stack) exists to support these two projects.

---

## 🚀 Knative Lambda

**The core serverless engine.** A Function-as-a-Service platform that automatically builds, deploys, and scales containerized functions from user code.

```
flux/infrastructure/knative-lambda/
├── src/operator/     # Go-based Kubernetes operator
├── k8s/              # Kustomize manifests
├── docs/             # Architecture & design docs
└── tests/            # Unit, integration, e2e, load tests
```

### How It Works

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         KNATIVE LAMBDA PLATFORM                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1️⃣ UPLOAD CODE                                                            │
│     └─ User uploads Python/Node.js/Go code to S3 (MinIO)                   │
│                                                                             │
│  2️⃣ AUTOMATIC BUILD                                                        │
│     └─ Operator creates Kaniko job → builds container → pushes to registry │
│                                                                             │
│  3️⃣ DEPLOY AS KNATIVE SERVICE                                              │
│     └─ Auto-scaling 0→N, CloudEvents triggers, health checks               │
│                                                                             │
│  4️⃣ SCALE TO ZERO                                                          │
│     └─ No traffic? Zero resources consumed. Traffic arrives? Scale up.     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Key Features

| Feature | Description |
|---------|-------------|
| 🏗️ **Dynamic Builds** | Kaniko-based in-cluster builds, no Docker daemon required |
| ⚡ **Scale-to-Zero** | Functions consume zero resources when idle |
| 🔄 **Event-Driven** | CloudEvents + RabbitMQ for async processing |
| 📊 **Full Observability** | Prometheus metrics, OpenTelemetry tracing, structured logging |
| 🔒 **Enterprise Security** | RBAC, mTLS, rate limiting, resource quotas |
| 🎯 **GitOps** | Flux CD integration, Helm-based deployments |

### Supported Languages

| Language | Dependency File | Runtime |
|----------|-----------------|---------|
| Python | `requirements.txt` | Python 3.11+ |
| Node.js | `package.json` | Node 20+ |
| Go | `go.mod` | Go 1.24+ |

### Quick Start

```bash
# Deploy the operator
kubectl apply -k flux/infrastructure/knative-lambda/k8s/

# Create a Lambda function
cat <<EOF | kubectl apply -f -
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaFunction
metadata:
  name: my-function
spec:
  runtime: python
  sourceS3Key: functions/my-function.zip
  handler: main.handler
EOF
```

**→ [Full Documentation](flux/infrastructure/knative-lambda/README.md)**

---

## 🛡️ Agent-Contracts

**AI-powered smart contract security.** A defensive AI agent that scans DeFi smart contracts for vulnerabilities, generates exploit proofs-of-concept, and alerts before attackers can exploit.

```
ai/agent-contracts/
├── src/
│   ├── contract_fetcher/    # Fetch contracts from block explorers
│   ├── vuln_scanner/        # Static analysis + LLM vulnerability detection
│   ├── exploit_generator/   # Generate defensive exploit PoCs
│   └── alert_dispatcher/    # Multi-channel alerting
├── k8s/kustomize/           # Kubernetes manifests
└── tests/                   # Unit & integration tests
```

### Architecture

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Contract   │───▶│    Vuln      │───▶│   Exploit    │───▶│    Alert     │
│   Fetcher    │    │   Scanner    │    │  Generator   │    │  Dispatcher  │
└──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
       │                   │                   │                   │
       └───────────────────┴───────────────────┴───────────────────┘
                                    │
                         RabbitMQ (CloudEvents)
```

### How It Works

| Step | Component | What It Does |
|------|-----------|--------------|
| 1 | **Contract Fetcher** | Monitors chains for newly deployed contracts, fetches source from Etherscan/BSCScan |
| 2 | **Vulnerability Scanner** | Runs Slither + LLM analysis (local Ollama or Claude fallback) |
| 3 | **Exploit Generator** | Creates PoC exploits to validate severity (**runs ONLY on local Anvil forks**) |
| 4 | **Alert Dispatcher** | Sends alerts via Grafana, Telegram, Discord |

### Key Features

| Feature | Description |
|---------|-------------|
| 🔍 **Multi-Chain Support** | Ethereum, BNB Chain, Polygon, Arbitrum, Base, Optimism |
| 🤖 **LLM-Powered Analysis** | Local inference (Ollama/DeepSeek-Coder) + cloud fallback |
| ⚡ **Serverless Execution** | Each component runs as a Knative Lambda function |
| 🛡️ **Safety First** | Exploits run ONLY against local Anvil forks, never mainnet |

### Vulnerability Detection

| Category | Examples |
|----------|----------|
| **Critical** | Reentrancy, arbitrary external calls, delegatecall injection |
| **High** | Access control issues, flash loan vectors, price oracle manipulation |
| **Medium** | Integer overflow/underflow, storage collision |
| **Low** | Missing view/pure modifiers, code smells |

### Quick Start

```bash
cd ai/agent-contracts

# Install dependencies
make install

# Run locally
make run-scanner

# Scan a specific contract
make scan-contract CHAIN=ethereum ADDR=0x1234...
```

**→ [Full Documentation](ai/agent-contracts/README.md)** | **[Requirements](ai/agent-contracts/REQUIREMENTS.md)**

---

## 🧱 Supporting Infrastructure

All components below exist to support running Knative Lambda and Agent-Contracts:

### Platform Layer

| Component | Purpose |
|-----------|---------|
| **Kubernetes** | 6 clusters (Pro, Pro-1-Node, Studio, Studio-1-Node, Pi, Forge) managed via Kind/k3s |
| **Linkerd** | Service mesh with multi-cluster mTLS communication |
| **Flux** | GitOps continuous delivery — push to Git, auto-deploy |
| **Pulumi** | Infrastructure as code — destroy and recreate from code |

### Runtime Layer

| Component | Purpose |
|-----------|---------|
| **Knative Serving** | Auto-scaling serverless runtime (0→N) |
| **Knative Eventing** | CloudEvents routing and triggers |
| **RabbitMQ** | Event bus for async function-to-function communication |
| **MinIO** | S3-compatible storage for function source code |

### AI/ML Layer

| Component | Purpose |
|-----------|---------|
| **Ollama** | Local LLM inference for AI agents |
| **VLLM** | High-performance LLM serving (Llama 3.1 70B on Forge GPU) |
| **Redis** | Caching layer for API responses |

### Observability Layer

| Component | Purpose |
|-----------|---------|
| **Prometheus** | Metrics collection and alerting |
| **Grafana** | Dashboards, logs, traces visualization |
| **Loki** | Log aggregation |
| **Tempo** | Distributed tracing |

### Security Layer

| Component | Purpose |
|-----------|---------|
| **Sealed Secrets** | Git-encrypted secrets management |
| **cert-manager** | Automatic TLS certificate provisioning |
| **Falco** | Runtime security monitoring |

---

## 🖥️ Cluster Topology

```
┌─────────────────────────────────────────────────────────────────┐
│  Mac Studio (M2 Ultra, 192GB)                                   │
│  ├─ Pro (Kind)        — 7 nodes  — Development, Testing        │
│  ├─ Pro-1-Node (Kind) — 1 node   — Lightweight Dev ⚡          │
│  ├─ Studio (Kind)     — 12 nodes — Production AI Agents ⭐     │
│  └─ Studio-1-Node     — 1 node   — Lightweight Prod 🎯         │
└─────────────────────────────────────────────────────────────────┘
                              │
                    Linkerd Multi-Cluster
                              │
       ┌──────────────────────┴──────────────────────┐
       │                                             │
┌──────┴──────┐                             ┌────────┴────────┐
│ Raspberry Pi │                             │   GPU Server    │
│ Pi (k3s)     │                             │ Forge (k3s)     │
│ 3-6 nodes    │                             │ 8 nodes         │
│ Edge/IoT 📡  │                             │ AI Training 🤖  │
└──────────────┘                             └─────────────────┘
```

### Single-Node Clusters

The `*-1-node` clusters are optimized for local development with minimal resource usage:

| Cluster | Purpose | Port Range |
|---------|---------|------------|
| **Pro-1-Node** | Dev environment with all services on one node | 31000-31999 |
| **Studio-1-Node** | Prod-like environment for testing deployments | 32000-32999 |

---

## 🚀 Quick Start

### Prerequisites

- Docker & kind (for local clusters)
- kubectl configured
- Pulumi CLI (for infrastructure)

### Deploy Everything

```bash
# Deploy to Pro cluster (development)
make up ENV=pro

# Check cluster status
kubectl get nodes

# Verify Knative Lambda operator
kubectl get pods -n knative-lambda

# Check multi-cluster connectivity
linkerd multicluster gateways
```

### Deploy Individual Components

```bash
# Knative Lambda only
kubectl apply -k flux/infrastructure/knative-lambda/k8s/

# Agent-Contracts only
kubectl apply -k ai/agent-contracts/k8s/kustomize/pro/

# Full Flux reconciliation
make reconcile-all
```

---

## 📁 Project Structure

```
homelab/
├── flux/
│   ├── clusters/                    # Cluster configurations (Kind/k3s)
│   └── infrastructure/
│       ├── knative-lambda/         # 🚀 MAIN: Serverless FaaS platform
│       ├── knative-operator/       # Knative operator (dependency)
│       ├── knative-instances/      # Knative Serving/Eventing instances
│       ├── rabbitmq-*/             # Event bus infrastructure
│       ├── minio/                  # S3-compatible storage
│       ├── prometheus-operator/    # Metrics & alerting
│       ├── loki/                   # Log aggregation
│       ├── tempo/                  # Distributed tracing
│       ├── linkerd/                # Service mesh
│       └── ...                     # Other supporting components
├── ai/
│   ├── agent-contracts/            # 🛡️ MAIN: Smart contract security agent
│   └── deeplearning/               # ML workloads
├── pulumi/                         # Infrastructure as Code
├── docs/                           # Full documentation
├── scripts/                        # Operational scripts
└── Makefile                        # Primary operational interface
```

---

## 📊 Observability

All functions expose metrics at `/metrics`:

```bash
# Knative Lambda metrics
knative_lambda_builds_total{status, runtime}
knative_lambda_build_duration_seconds{runtime}
knative_lambda_functions_active{namespace}

# Agent-Contracts metrics
contracts_fetched_total{chain, status}
vulnerabilities_found_total{chain, severity, type}
scan_duration_seconds{chain, analyzer}
exploits_validated_total{chain, success}
```

Access dashboards:

```bash
# Port-forward Grafana
kubectl port-forward svc/grafana 3000:3000 -n monitoring

# Open http://localhost:3000
```

---

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [📖 Full Documentation](docs/README.md) | Complete documentation hub |
| [🚀 Knative Lambda Docs](flux/infrastructure/knative-lambda/README.md) | FaaS platform details |
| [🛡️ Agent-Contracts Docs](ai/agent-contracts/README.md) | Security agent details |
| [🎯 Architecture](docs/ARCHITECTURE.md) | System architecture overview |
| [☸️ Cluster Guides](docs/clusters/) | Per-cluster documentation |

---

## 🎯 Why This Architecture?

**The goal:** Run AI agents locally with the same capabilities as cloud providers — but on hardware you own.

| Cloud FaaS | This Homelab |
|------------|--------------|
| Vendor lock-in | Run anywhere with Kubernetes |
| Pay-per-invocation | Zero cost when idle |
| Limited observability | Full metrics, logs, traces |
| Black-box scaling | Transparent auto-scaling |
| No local development | Same stack locally and in prod |

**Orchestration benefits:** Kubernetes handles the hard parts — service discovery, health checks, rolling updates, resource scheduling. You focus on writing functions.

---

**Maintained by**: Bruno Lucena  
**License**: MIT
