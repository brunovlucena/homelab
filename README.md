# 🏠 Homelab Infrastructure

> **Production-grade Kubernetes infrastructure with AI agents, observability, and serverless capabilities**

This repository contains the complete infrastructure-as-code setup for a sophisticated homelab environment, built with **Pulumi**, **Flux GitOps**, **Linkerd service mesh**, and comprehensive AI-powered observability.

## 🎯 Overview

The homelab project provides an enterprise-grade Kubernetes infrastructure featuring:
- **Kind** multi-node cluster with specialized worker nodes
- **Pulumi (Go)** for declarative infrastructure provisioning
- **Flux** for GitOps-based continuous deployment
- **Linkerd** service mesh for zero-trust networking
- **AI Agents** for SRE automation and intelligent monitoring
- **Knative** for serverless workloads
- **Full observability stack** (Prometheus, Grafana, Loki, Tempo, Alloy)

## 🏗️ Architecture

```
┌────────────────────────────────────────────────────────────────────────────────────┐
│                          🏠 Homelab Kubernetes Cluster                             │
│                        Kind Multi-Node (1 Control + 4 Workers)                     │
└────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                    ┌───────────────────┴───────────────────┐
                    │                                       │
             ┌──────▼──────┐                         ┌──────▼──────┐
             │   Control   │                         │   Workers   │
             │    Plane    │                         │    (x4)     │
             │  (Linkerd)  │                         └─────────────┘
             └─────────────┘                                │
                    │                    ┌──────────────────┼──────────────────┐
                    │                    │                  │                  │
                    │             ┌──────▼──────┐    ┌──────▼──────┐   ┌──────▼──────┐
                    │             │ AI Worker   │    │ Serverless  │   │Observability│
                    │             │:30120-30124 │    │   Worker    │   │  Workers    │
                    │             └─────────────┘    │:30130-30132 │   │(x2 for HA)  │
                    │                    │           └─────────────┘   │:30040-30053 │
                    │                    │                  │           └─────────────┘
                    └────────────────────┼──────────────────┤                  │
                                         │                  │                  │
                    ┌────────────────────┼──────────────────┤                  │
                    │                    │                  │                  │
         ┌──────────▼─────────┐  ┌──────▼─────────┐  ┌─────▼──────┐  ┌────────▼────────┐
         │   🤖 AI Agents     │  │  ⚡ Serverless  │  │ 🏠 Apps    │  │ 📊 Observability│
         ├────────────────────┤  ├────────────────┤  ├────────────┤  ├─────────────────┤
         │ • Agent Bruno      │  │ Knative Stack: │  │ • Homepage │  │ Prometheus      │
         │   (Memory + KB)    │  │ • Operator     │  │   API (Go) │  │ Operator        │
         │ • Agent Jamie      │  │ • Serving      │  │   Frontend │  │ • Grafana:30040 │
         │   (Slack Bot)      │  │ • Eventing     │  │   (React)  │  │ • Prom:30041    │
         │ • Agent SRE        │  │                │  │ • Notifi   │  │                 │
         │   (Automation)     │  │ Event Brokers  │  │   Test     │  │ Loki:30042      │
         │ • MCP Server       │  │ & Triggers     │  │            │  │ Tempo:30043     │
         │   :30123           │  │                │  │            │  │ Alloy           │
         └────────────────────┘  └────────────────┘  └────────────┘  └─────────────────┘
                    │                    │                  │                  │
                    └────────────────────┴──────────────────┴──────────────────┘
                                         │
                    ┌────────────────────┼────────────────────────────┐
                    │                    │                            │
         ┌──────────▼─────────┐  ┌───────▼───────┐        ┌──────────▼─────────┐
         │  🔄 GitOps & IaC   │  │  🔗 Service   │        │  💾 Data Stores    │
         ├────────────────────┤  │     Mesh      │        ├────────────────────┤
         │ Flux GitOps        │  ├───────────────┤        │ • PostgreSQL       │
         │ • HelmRelease      │  │ Linkerd       │        │ • Redis            │
         │ • Kustomize        │  │ • mTLS        │        │ • MongoDB          │
         │ • GitRepository    │  │ • Viz :8084   │        │ • MinIO (S3)       │
         │                    │  │ • Traffic     │        │                    │
         │ Pulumi (Go)        │  │   Control     │        │ Backends:          │
         │ • IaC              │  │ • Policies    │        │ • Loki storage     │
         │ • Stack Mgmt       │  │               │        │ • Tempo storage    │
         └────────────────────┘  └───────────────┘        └────────────────────┘
                    │                    │                            │
                    └────────────────────┴────────────────────────────┘
                                         │
                    ┌────────────────────┼────────────────────────────┐
                    │                    │                            │
         ┌──────────▼─────────┐  ┌───────▼────────┐       ┌──────────▼─────────┐
         │  🔒 Security &     │  │  🛠️ Platform   │       │  🌐 External       │
         │     Networking     │  │    Services    │       │     Access         │
         ├────────────────────┤  ├────────────────┤       ├────────────────────┤
         │ Cert Manager       │  │ Metrics Server │       │ Cloudflare Tunnel  │
         │ • ClusterIssuer    │  │ Headlamp       │       │ • Zero Trust       │
         │ • Auto TLS         │  │ • K8s UI       │       │ • Auto HTTPS       │
         │                    │  │   :30000       │       │ • HA Canary        │
         │ Sealed Secrets     │  │                │       │ • ServiceMonitor   │
         │ • Encrypted Mgmt   │  │ K6 Operator    │       │                    │
         │                    │  │ • Load Tests   │       │ NodePorts          │
         │ Network Policies   │  │                │       │ • External access  │
         │ • Isolation        │  │ Statuspage     │       │ • Local dev        │
         │ • Zero Trust       │  │ Exporter       │       │                    │
         └────────────────────┘  └────────────────┘       └────────────────────┘
                                         │
                    ┌────────────────────┴────────────────────┐
                    │                                         │
         ┌──────────▼─────────┐                   ┌──────────▼─────────┐
         │  📦 Helm Repos     │                   │  🎯 Monitoring     │
         ├────────────────────┤                   ├────────────────────┤
         │ • Grafana          │                   │ ServiceMonitors    │
         │ • Prometheus       │                   │ • All services     │
         │ • Bitnami          │                   │                    │
         │ • Linkerd          │                   │ PrometheusRules    │
         │ • Knative          │                   │ • Alerts           │
         │ • Git Sources      │                   │                    │
         └────────────────────┘                   │ Grafana Dashboards │
                                                  │ • Homepage         │
                                                  │ • Strava           │
                                                  │ • Golden Signals   │
                                                  └────────────────────┘

┌────────────────────────────────────────────────────────────────────────────────────┐
│                               🔄 Data Flow Overview                                │
├────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                    │
│  User → Slack → Jamie Bot → Agent SRE → Prometheus/Loki/Tempo → Response           │
│                              ↓                                                     │
│                         MCP Server → Tools → Observability Stack                   │
│                                                                                    │
│  Apps → Alloy → Prometheus/Loki/Tempo → Grafana → Dashboards                       │
│                                                                                    │
│  Git Push → Flux → Reconcile → Apply → Kubernetes → Linkerd Mesh → Services        │
│                                                                                    │
│  Pulumi → Kind Cluster → Flux Bootstrap → GitOps → Infrastructure Deployment       │
│                                                                                    │
└────────────────────────────────────────────────────────────────────────────────────┘
```

## 📁 Project Structure

```
homelab/
├── 📄 Makefile                          # 🛠️  Automation commands
├── 📄 README.md                         # 📖 This file
├── 📄 ARCHITECTURE.md                   # 🏗️  Detailed architecture docs
├── 📄 .gitignore                        # 🚫 Git ignore rules
├── 📁 .github/                          # 🤖 GitHub Actions workflows
│   ├── 📁 workflows/                    # CI/CD pipelines
│   │   ├── agent-bruno.yml             # Agent Bruno image builds
│   │   ├── agent-sre-images.yml        # Agent SRE image builds
│   │   ├── homepage-images.yml         # Homepage image builds
│   │   ├── jamie-images.yml            # Jamie bot image builds
│   │   └── dependabot.yml              # Automated dependency updates
│   └── 📄 README.md                     # Workflows documentation
├── 📁 pulumi/                           # 🏗️  Infrastructure as Code
│   ├── 📄 main.go                       # 🐹 Pulumi Go program
│   ├── 📄 Pulumi.yaml                   # ⚙️  Pulumi configuration
│   ├── 📄 go.mod                        # 📦 Go dependencies
│   └── 📄 go.sum                        # 🔒 Go checksums
├── 📁 scripts/                          # 🔧 Bootstrap & utility scripts
│   ├── 📄 create-kind-cluster.sh       # Kind cluster creation
│   ├── 📄 create-secrets.sh            # Kubernetes secrets setup
│   └── 📄 install-*.sh                 # Legacy installation scripts (optional)
└── 📁 flux/                             # 🔄 GitOps configuration
    └── 📁 clusters/                     # 🎯 Cluster-specific configs
        └── 📁 homelab/                  # 🏠 Homelab cluster
            ├── 📄 kind.yaml             # 🐳 Multi-node Kind config
            ├── 📄 kustomization.yaml
            └── 📁 infrastructure/       # 🏗️  K8s applications
                ├── 📁 flux-bootstrap/   # 🔄 Flux self-installation
                ├── 📁 linkerd/          # 🔗 Linkerd via GitOps Jobs
                ├── 📁 agent-bruno/      # 🤖 Bruno AI assistant
                ├── 📁 agent-jamie/      # 💬 Jamie Slack bot
                ├── 📁 agent-sre/        # 🚨 SRE automation agent
                ├── 📁 homepage/         # 🏠 Personal homepage with chatbot
                ├── 📁 prometheus-operator/ # 📊 Monitoring stack
                ├── 📁 loki/             # 📝 Log aggregation
                ├── 📁 tempo/            # 🔍 Distributed tracing
                ├── 📁 alloy/            # 🔄 Telemetry collector
                ├── 📁 knative-operator/ # ⚡ Serverless platform
                ├── 📁 knative-serving/  # 🚀 Serving components
                ├── 📁 knative-eventing/ # 📬 Event-driven architecture
                ├── 📁 postgres/         # 🐘 PostgreSQL database
                ├── 📁 redis/            # 🔴 Redis cache
                ├── 📁 mongodb/          # 🍃 MongoDB database
                ├── 📁 minio/            # 📦 Object storage
                ├── 📁 cloudflare-tunnel/ # 🌐 Secure tunnel
                ├── 📁 cert-manager/     # 🔒 Certificate management
                ├── 📁 metrics-server/   # 📈 Resource metrics
                ├── 📁 headlamp/         # 🎛️  K8s dashboard
                └── 📁 repositories/     # 📚 Helm & Git repos
```

## 🚀 Quick Start

### Prerequisites

- **Docker Desktop** with Kind support (minimum 8 CPU, 16GB RAM recommended)
- **kubectl** (v1.34+) for Kubernetes management
- **Pulumi CLI** (v3.0+) for infrastructure provisioning
- **Flux CLI** (v2.0+) for GitOps operations
- **Go** (v1.23+) for Pulumi program compilation
- **Linkerd CLI** for service mesh operations

### Environment Setup

1. **Set required environment variables:**

   Add the following to your `~/.zshrc`:

```bash
# GitHub Container Registry
export GITHUB_TOKEN="ghp_your_github_token"
export GHCR_USERNAME="your_github_username"
export GHCR_TOKEN="ghp_your_github_token"

# Cloudflare
export CLOUDFLARE_TOKEN="your_cloudflare_tunnel_token"

# Grafana
export GRAFANA_PASSWORD="your_grafana_password"

# Slack (for Jamie bot and Alertmanager)
export SLACK_APP_JAMIE_APP_TOKEN="xapp-..."
export SLACK_BOT_JAMIE_OAUTH_TOKEN="xoxb-..."
export SLACK_SIGNING_SECRET="..."
export SLACK_WEBHOOK_URL="https://hooks.slack.com/..."

# AI/Observability
export LANGSMITH_API_KEY="lsv2_..."
export LOGFIRE_TOKEN="..."

# Databases
export POSTGRES_PASSWORD="your_postgres_password"
export REDIS_PASSWORD="your_redis_password"
export MINIO_ROOT_USER="admin"
export MINIO_ROOT_PASSWORD="your_minio_password"

# See scripts/create-secrets.sh for complete list
```

2. **Initialize the Pulumi stack (first time only):**

```bash
cd pulumi
pulumi stack init homelab
cd ..
```

3. **Deploy the infrastructure:**

```bash
make up
```

This single command will:
- ✅ Create a multi-node Kind cluster with specialized workers
- ✅ Install Flux GitOps controllers via Job-based bootstrap
- ✅ **Automatically create all Kubernetes secrets** (via `scripts/create-secrets.sh`)
- ✅ Install Linkerd service mesh + Viz via GitOps Jobs
- ✅ Deploy all infrastructure components via Kustomize
- ✅ Set up observability stack (Prometheus, Grafana, Loki, Tempo)
- ✅ Deploy AI agents (Bruno, Jamie, SRE)
- ✅ Configure Knative serverless platform

**Secrets automatically created:**
- `ghcr-secret` - GitHub Container Registry credentials (jamie, bruno, agent-sre)
- `prometheus-secrets` - Grafana, Slack, PagerDuty, Strava
- `agent-sre-secrets` - AI services (LangSmith, Logfire, GitHub, HuggingFace)
- `jamie-secrets` - Slack bot credentials
- `bruno-site-secret` - Homepage app credentials (PostgreSQL, Redis, MinIO)
- `alloy-secrets` - Telemetry pipeline credentials
- `loki-minio-secret` - Loki storage backend
- `minio-secret` - MinIO root credentials
- `cloudflare-tunnel-credentials` - Tunnel token

**First deployment takes ~10-15 minutes** ⏱️

## 🛠️ Available Commands

### Infrastructure Management

| Command | Description |
|---------|-------------|
| `make init` | 🏗️ Initialize Pulumi homelab stack |
| `make up` | 🚀 Deploy homelab stack (with validation) |
| `make destroy` | 💥 Destroy homelab stack |
| `make cancel` | ⏹️ Cancel ongoing Pulumi operation |

### Flux Operations

| Command | Description |
|---------|-------------|
| `make flux-refresh` | 🔄 Refresh all HelmRepositories, GitRepositories, and HelmReleases |

### Linkerd Service Mesh

| Command | Description |
|---------|-------------|
| `make linkerd-install` | 🔗 Install Linkerd service mesh |
| `make linkerd-status` | 📊 Check Linkerd health status |
| `make linkerd-viz-install` | 📈 Install Linkerd Viz dashboard |
| `make linkerd-viz-status` | 📊 Check Linkerd Viz status |
| `make linkerd-dashboard` | 🌐 Open Linkerd dashboard (localhost:8084) |

## 🔧 Infrastructure Components

### 🎯 Core Infrastructure

- **Kind Cluster**: Multi-node local Kubernetes cluster (1 control-plane + 4 workers)
  - **AI Worker**: Dedicated node for AI agents (ports 30120-30124)
  - **Serverless Worker**: Knative workloads (ports 30130-30132)
  - **Observability Workers (x2)**: Monitoring stack (ports 30040-30053)
- **Flux GitOps**: Automated deployment and reconciliation via Job-based bootstrap
- **Linkerd Service Mesh**: Zero-trust networking, mTLS, observability (installed via GitOps Jobs)
- **Cert Manager**: Automatic SSL/TLS certificate provisioning
- **Metrics Server**: Kubernetes resource metrics API

**🚀 GitOps-First Approach:**
Both Flux and Linkerd are installed using Kubernetes Jobs that run their respective CLIs. This ensures:
- ✅ Fully declarative and version-controlled
- ✅ No blocking Pulumi operations (runs asynchronously)
- ✅ Idempotent and reliable installation
- ✅ Easy version updates via image tags

### 🤖 AI & Automation Agents

- **Agent Bruno** 🧑‍💻: Personal AI assistant with memory and knowledge management
  - Persistent memory with MongoDB
  - Code snippets, runbooks, incident tracking
  - REST API + MCP protocol support
  
- **Agent Jamie** 💬: Slack bot with LLM brain (Ollama)
  - LangChain-powered tool calling
  - Integration with Agent SRE
  - Interactive SRE assistance
  
- **Agent SRE** 🚨: SRE automation and observability agent
  - LangGraph state management
  - Direct queries to Prometheus, Loki, Tempo, Grafana
  - Alertmanager webhook receiver
  - Golden signals monitoring
  - MCP server for tool exposure

### ⚡ Serverless Platform

- **Knative Operator**: Manages Knative installation
- **Knative Serving**: Request-driven autoscaling workloads
- **Knative Eventing**: Event-driven architecture with brokers and triggers

### 📊 Observability Stack (LGTM)

- **Prometheus Operator**: Metrics collection with ServiceMonitors
  - Grafana: Visualization and dashboards (NodePort 30040)
  - Alertmanager: Alert routing and management
  - Node Exporter: System metrics
  - Kube State Metrics: Kubernetes state
- **Loki**: Log aggregation and querying (backed by MinIO)
- **Tempo**: Distributed tracing (backed by MinIO)
- **Alloy**: OpenTelemetry collector and telemetry pipeline
- **Custom Dashboards**: Homepage dashboard, Strava integration

### 💾 Data Stores

- **PostgreSQL**: Relational database (Homepage backend)
- **Redis**: In-memory cache and session store
- **MongoDB**: Document store (Agent memory persistence)
- **MinIO**: S3-compatible object storage (Loki, Tempo)

### 🌐 Networking & Access

- **Cloudflare Tunnel**: Secure external access without port forwarding
  - Zero Trust security
  - Automatic HTTPS
  - High availability with canary monitoring
- **Headlamp**: Modern Kubernetes dashboard UI

### 🏠 Applications

- **Homepage**: Personal homepage with integrated chatbot
  - Go API backend with OpenTelemetry
  - React frontend with real-time metrics
  - Agent SRE integration for intelligent assistance
  - PostgreSQL persistence

### 🛡️ Security & Monitoring

- **Sealed Secrets**: Encrypted secret management (commented out)
- **ServiceMonitors**: Prometheus scraping for all services
- **Grafana Dashboards**: Pre-configured monitoring views
- **Alertmanager Rules**: Proactive alerting
  - Cloudflare Tunnel health monitoring
  - Custom service alerts

## 🌐 Network Architecture

### Node Port Mappings

The cluster exposes services via NodePorts on specialized workers:

**AI Worker** (role: ai):
- `30120`: Jamie Slack Bot
- `30121`: Agent Bruno
- `30122`: Agent SRE API
- `30123`: Agent SRE MCP Server
- `30124`: Reserved

**Serverless Worker** (role: serverless):
- `30130-30132`: Knative services

**Observability Workers** (role: observability):
- **Node 1**: `30040` (Grafana), `30041` (Prometheus)
- **Node 2**: `30050` (Grafana HA), `30051` (Prometheus HA)
- Loki & Tempo also exposed

### External Access

- **Local Development**: Direct access via NodePorts
- **External Access**: Cloudflare Tunnel (secure, no port forwarding)
- **Container Registry**: GitHub Container Registry (ghcr.io/brunovlucena/)
- **Service Mesh**: Linkerd provides encrypted inter-service communication

### Cloudflare Tunnel

Secure external access without exposing ports:
- ✅ Zero Trust security model
- ✅ Automatic HTTPS certificates
- ✅ High availability with canary deployment
- ✅ Prometheus metrics and health monitoring
- ✅ Integrated with observability stack

**Setup:**
1. Create tunnel: [Cloudflare Zero Trust Dashboard](https://one.dash.cloudflare.com/)
2. Configure routes in Cloudflare dashboard
3. Tunnel auto-deploys via Flux

See [`cloudflare-tunnel/TROUBLESHOOTING.md`](flux/clusters/homelab/infrastructure/cloudflare-tunnel/TROUBLESHOOTING.md) for details.

## 🔄 GitOps Workflow

The repository follows a fully automated GitOps approach:

### 1️⃣ Infrastructure Changes
```bash
# Modify Pulumi infrastructure
vim pulumi/main.go

# Apply changes
make up
```

### 2️⃣ Application Changes
```bash
# Update Kubernetes manifests
vim flux/clusters/homelab/infrastructure/<component>/

# Commit and push
git add . && git commit -m "Update component" && git push

# Flux auto-syncs within 1 minute
flux reconcile kustomization infrastructure --with-source
```

### 3️⃣ Image Updates

**Automated via GitHub Actions:**
- `agent-bruno.yml`: Builds Agent Bruno images
- `agent-sre-images.yml`: Builds Agent SRE + MCP server images
- `jamie-images.yml`: Builds Jamie Slack bot images  
- `homepage-images.yml`: Builds Homepage API + Frontend images
- All images tagged as `:latest` and pushed to ghcr.io

**Trigger builds:**
```bash
# Automatically on push to main
git push origin main

# Or manually dispatch
gh workflow run agent-sre-images.yml
```

### 4️⃣ Dependency Updates

**Dependabot** automatically creates PRs for:
- Go dependencies (Pulumi, Homepage API)
- Docker base images
- GitHub Actions versions
- Helm chart updates

### 5️⃣ Monitoring Deployments

```bash
# Check Flux status
flux get all -A

# Watch HelmRelease progress
watch kubectl get helmrelease -A

# Check application pods
kubectl get pods -A

# View Linkerd mesh status
make linkerd-status
```

## 🤖 CI/CD & Automation

### GitHub Actions Workflows

The repository includes comprehensive CI/CD pipelines:

#### 🔨 Image Build Workflows

| Workflow | Trigger | Images Built |
|----------|---------|--------------|
| `agent-bruno.yml` | Push to `agent-bruno/` | Agent Bruno API |
| `agent-sre-images.yml` | Push to `agent-sre/` | Agent SRE API + MCP Server |
| `jamie-images.yml` | Push to `agent-jamie/` | Jamie Slack Bot + MCP Server |
| `homepage-images.yml` | Push to `homepage/` | Homepage API (Go) + Frontend (React) |

**Features:**
- ✅ Multi-stage Docker builds
- ✅ Layer caching for speed
- ✅ Multi-platform (amd64/arm64)
- ✅ Automatic tagging (`:latest`)
- ✅ Push to ghcr.io

#### 🔄 Dependabot Configuration

**Automated dependency updates for:**
- 📦 Go modules (Pulumi, Homepage API)
- 🐳 Docker base images
- 📋 GitHub Actions versions
- 🐍 Python packages (AI agents)
- 📦 npm packages (Homepage frontend)

**Auto-merge enabled for:**
- Patch version updates
- Minor version updates (non-breaking)
- Security patches

See [`.github/workflows/README.md`](.github/workflows/README.md) for complete documentation.

## 🐛 Troubleshooting

### Common Issues

#### 1. Missing Environment Variables
```bash
Error: GITHUB_TOKEN environment variable is required
```
**Solution:**
```bash
export GITHUB_TOKEN="ghp_your_token_here"
make up
```

#### 2. Kind Cluster Creation Fails
```bash
Error: failed to create cluster: port is already allocated
```
**Solution:**
```bash
# Check for existing clusters
kind get clusters

# Delete old cluster
kind delete cluster --name homelab

# Recreate
make up
```

#### 3. Insufficient Docker Resources
```bash
Error: failed to create container
```
**Solution:**
- Increase Docker Desktop resources to at least:
  - CPUs: 8
  - Memory: 16 GB
  - Swap: 4 GB

#### 4. Flux Not Syncing
```bash
# Check Flux status
flux get all -A

# Force reconciliation
make flux-refresh

# Check specific resource
flux reconcile helmrelease <name> -n <namespace>
```

#### 5. Linkerd Issues
```bash
# Check Linkerd health
make linkerd-status

# View Linkerd logs
kubectl logs -n linkerd -l linkerd.io/control-plane-component=controller

# Reinstall if needed
make linkerd-install
```

#### 6. AI Agents Not Responding
```bash
# Check agent pods
kubectl get pods -n agent-bruno
kubectl get pods -n agent-jamie  
kubectl get pods -n agent-sre

# View logs
kubectl logs -n agent-sre -l app=sre-agent -f

# Check Ollama connectivity
kubectl exec -it -n agent-sre <pod> -- curl http://192.168.0.16:11434/api/tags
```

### Useful Debugging Commands

```bash
# Cluster health
kubectl get nodes
kubectl cluster-info

# All pods status
kubectl get pods -A

# Flux status
flux get all -A
kubectl get gitrepository -A
kubectl get helmrelease -A

# Linkerd mesh
linkerd viz stat deployment -A
linkerd viz tap deployment/<name> -n <namespace>

# Service mesh connectivity
linkerd viz top deployment/<name> -n <namespace>

# Observability stack
kubectl get servicemonitor -A
kubectl get prometheusrule -A

# Check specific service
kubectl describe pod <pod-name> -n <namespace>
kubectl logs <pod-name> -n <namespace> -f

# Port forward for local access
kubectl port-forward -n prometheus svc/prometheus-operated 9090:9090
kubectl port-forward -n prometheus svc/grafana 3000:80

# Force image pull
kubectl delete pod -n <namespace> -l app=<app-name>
```

### Quick Reset

If everything breaks:
```bash
# Nuclear option: destroy and recreate
make destroy
make up

# This will rebuild the entire cluster from scratch
# Takes ~10-15 minutes
```

## 🔒 Security

### Security Features

- **🔗 Linkerd Service Mesh**: Automatic mTLS between all services
- **🔐 Sealed Secrets**: Encrypted secret management (optional)
- **📛 Network Policies**: Service-to-service access control
- **🎯 Namespace Isolation**: Each component in dedicated namespace
- **🔑 RBAC**: Role-based access control
- **🛡️ Cloudflare Zero Trust**: Secure external access
- **📊 Security Monitoring**: Grafana dashboards for security metrics

### Secrets Management

**Current secrets stored as Kubernetes Secrets:**
```bash
# Agent SRE secrets
- GRAFANA_API_KEY
- LOGFIRE_TOKEN_SRE_AGENT  
- LANGSMITH_API_KEY

# Jamie Slack bot secrets
- SLACK_BOT_TOKEN
- SLACK_SIGNING_SECRET
- SLACK_APP_TOKEN

# Cloudflare secrets
- CLOUDFLARE_TUNNEL_TOKEN
- CLOUDFLARE_API_TOKEN
```

**Best practices:**
- 🔄 Rotate tokens quarterly
- 🔒 Use Sealed Secrets for Git storage (optional)
- 📝 Never commit plaintext secrets
- 🎯 Principle of least privilege

## 📈 Monitoring & Observability

### LGTM Stack (Loki, Grafana, Tempo, Mimir)

**Full observability pipeline:**

```
Application Metrics/Logs/Traces
            │
            ▼
    Alloy (Collector)
            │
    ┌───────┼───────┐
    │       │       │
    ▼       ▼       ▼
Prometheus  Loki   Tempo
    │       │       │
    └───────┼───────┘
            ▼
         Grafana
      (Visualization)
```

### Access Points

| Service | URL | Purpose |
|---------|-----|---------|
| Grafana | http://localhost:30040 | Dashboards and visualization |
| Prometheus | http://localhost:30041 | Raw metrics queries |
| Linkerd Viz | http://localhost:8084 | Service mesh observability |
| Headlamp | http://localhost:30000 | Kubernetes dashboard |

### Pre-configured Dashboards

- **Homepage Dashboard**: Full-stack application monitoring
- **Strava Dashboard**: Personal fitness tracking integration
- **Golden Signals**: Latency, traffic, errors, saturation
- **Linkerd Dashboards**: Service mesh traffic and health
- **Cloudflare Tunnel**: Tunnel health and connectivity

### Metrics Collection

All services instrumented with:
- **Prometheus ServiceMonitors**: Automatic scraping
- **OpenTelemetry**: Distributed tracing
- **Structured Logging**: JSON logs to Loki
- **Custom Metrics**: Application-specific metrics

### Alerting

**Alertmanager configured for:**
- 📊 Resource utilization alerts
- 🚨 Service health monitoring
- 🌐 Cloudflare Tunnel connectivity
- 🤖 AI agent responsiveness
- 📈 Performance degradation

**Alert channels:**
- Slack notifications (optional)
- PagerDuty integration (optional)
- Grafana dashboard alerts

## 🎓 Learning Resources

This homelab demonstrates:

- ✅ **GitOps**: Flux-based continuous deployment
- ✅ **Infrastructure as Code**: Pulumi with Go
- ✅ **Service Mesh**: Linkerd for zero-trust networking
- ✅ **Observability**: Complete LGTM stack implementation
- ✅ **Serverless**: Knative for event-driven workloads
- ✅ **AI/ML Ops**: Production LLM agent deployment
- ✅ **CI/CD**: GitHub Actions with automated builds
- ✅ **Multi-tenancy**: Namespace isolation patterns
- ✅ **Security**: mTLS, RBAC, secret management

### Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md): Detailed technical architecture
- [Agent Bruno README](flux/clusters/homelab/infrastructure/agent-bruno/README.md): AI assistant details
- [Agent SRE README](flux/clusters/homelab/infrastructure/agent-sre/README.md): SRE automation
- [Homepage README](flux/clusters/homelab/infrastructure/homepage/README.md): Chatbot integration
- [Workflows README](.github/workflows/README.md): CI/CD pipelines

## 🤝 Contributing

This is a personal infrastructure project, but contributions are welcome:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Test thoroughly in your own cluster
5. Commit (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## 📊 Project Stats

- **Lines of Code**: ~50,000+
- **Components**: 25+ services
- **Namespaces**: 15+
- **Custom Dashboards**: 5+
- **AI Agents**: 3
- **Languages**: Go, Python, TypeScript, YAML
- **Monthly Cost**: $0 (fully local)

## 🙏 Acknowledgments

- **Pulumi** for excellent IaC tooling
- **Flux** for robust GitOps automation
- **Linkerd** for lightweight service mesh
- **CNCF** for amazing cloud-native projects
- **Kind** for local Kubernetes development
- **Grafana Labs** for observability stack
- **Knative** for serverless platform
- **Ollama** for local LLM inference

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Built with ❤️ for learning, automation, and curiosity**

🏠 **Homelab** | 🤖 **AI-Powered** | 📊 **Fully Observable** | ⚡ **Cloud-Native**
