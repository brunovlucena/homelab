# 🏗️ Homelab Architecture Overview

> **Purpose**: Connecting Teams, Clusters & Edge Devices in a Unified Platform  
> **Last Updated**: November 7, 2025  
> **Production Readiness**: 62% → Target: 94%  
> **Architecture**: 5 Clusters + Edge Devices via Linkerd Service Mesh

---

## 📚 Complete Documentation

**→ [Documentation Hub](README.md)** - Start here for all documentation

---

## Overview

The homelab is a **production-grade infrastructure** that connects:

1. **🧑‍💻 Teams** - Developers, SREs, Data Scientists
2. **☸️ Clusters** - 5 Kubernetes clusters (Air, Pro, Studio, Pi, Forge)
3. **📡 Edge Devices** - Raspberry Pi, IoT sensors, gateways
4. **🤖 AI/ML** - GPU inference, training, intelligent agents

### Five Clusters

- **Air** (Kind, 4 nodes) - Experimental & CI/CD
- **Pro** (Kind, 7 nodes) - Development & Testing
- **Studio** (Kind, 12 nodes) - Production AI Agents ⭐
- **Pi** (k3s, 3-6 nodes) - Edge & IoT
- **Forge** (k3s, 8 nodes) - GPU Training & Inference 🤖

**→ See [Cluster Comparison](#cluster-comparison) below**  
**→ See individual cluster docs**: [Air](clusters/air-cluster.md) | [Pro](clusters/pro-cluster.md) | [Studio](clusters/studio-cluster.md) | [Pi](clusters/pi-cluster.md) | [Forge](clusters/forge-cluster.md)

### Key Technologies

- **Service Mesh**: Linkerd multi-cluster with mTLS
- **GitOps**: Flux continuous delivery
- **IaC**: Pulumi (all infrastructure as code)
- **AI/ML**: VLLM (Forge), Ollama, AI agents (Studio)
- **Secrets**: External Secrets Operator with GitHub backend
- **Observability**: Prometheus, Grafana, Loki, Tempo

### Design Principles

1. **Connectivity First** - Teams, clusters, edge as first-class citizens
2. **Role-Based Segregation** - Each node has specific workload types
3. **Multi-Tier Architecture** - Infrastructure → Platform → Application → Data
4. **HA Simulation** - Multi-zone deployment (Studio)
5. **Multi-Cluster Mesh** - Linkerd + WARP tunnels

### Connection Architecture

```
┌─────────────────────────────────────────────────────┐
│          5 Clusters + Edge via Linkerd Mesh         │
├─────────────────────────────────────────────────────┤
│                                                     │
│  Mac Studio (192GB)         GPU Server              │
│  ┌─────────────────┐        ┌─────────────┐         │
│  │ Air    (4 nodes)│        │   Forge     │         │
│  │ Pro    (7 nodes)│◄──────►│  (8 nodes)  │         │
│  │ Studio (12 nodes)│ Linkerd│  2×A100 GPU│          │
│  └────────┬────────┘        └─────────────┘         │
│           │                                         │
│           │ WARP Tunnels                            │
│           │                                         │
│  ┌────────▼────────┐                                │
│  │  Pi (3-6 nodes) │                                │
│  │  Edge & IoT     │                                │
│  └─────────────────┘                                │
│                                                     │
│  ✅ mTLS everywhere                                 │
│  ✅ Cross-cluster service discovery                 │
│  ✅ GitOps deployments (Flux)                       │
│  ✅ Unified observability (Prometheus/Grafana)      │
│                                                     │
└─────────────────────────────────────────────────────┘
```

**How It Connects:**

- **Teams**: Access all clusters via AI agents and GitOps workflows
- **Clusters**: Service mesh enables `service.namespace.svc.cluster.remote` pattern
- **Edge**: Pi nodes integrated into same service mesh as cloud clusters

**→ [See AI Connectivity](architecture/ai-connectivity.md) for detailed connectivity patterns**

---

## Cluster Comparison

| Aspect | Air | Pro | Studio | Pi | Forge |
|--------|-----|-----|--------|-----|-------|
| **Platform** | Kind | Kind | Kind | k3s | k3s |
| **Hardware** | Mac Studio | Mac Studio | Mac Studio | Raspberry Pi | NVIDIA GPU Server |
| **Architecture** | ARM64 | ARM64 | ARM64 | ARM64 | x86_64 |
| **Purpose** | Quick experiments | Dev/Test | Production-like | Edge/IoT | AI Training |
| **Total Nodes** | 4 | 7 | 12 | 3-6 | 8 |
| **Node Strategy** | Multi-purpose | Specialized | Highly specialized + HA | Lightweight | GPU-optimized |
| **HA Simulation** | No | Partial | Yes (zone-aware) | Yes (distributed) | Yes (GPU pools) |
| **Resource Usage** | Minimal | Moderate | High | Very Low | Extreme |
| **Network** | 10.246.0.0/16 | 10.244.0.0/16 | 10.245.0.0/16 | 10.247.0.0/16 | 10.248.0.0/16 |
| **Multi-Cluster** | Gateway enabled | Gateway enabled | Primary gateway | Edge gateway | ML gateway |
| **WARP Tunnel** | Yes | Yes | Yes | Yes | Yes |
| **Best For** | CI/CD, testing | Full-stack dev | AI agents, production | IoT, edge compute | Model training, fine-tuning |

### Cluster Roles

**Air Cluster** - Experimental, CI/CD, rapid iteration  
→ [Full Documentation](clusters/air-cluster.md)

**Pro Cluster** - Development, testing, full-stack applications  
→ [Full Documentation](clusters/pro-cluster.md)

**Studio Cluster** - Production AI agents, high availability, observability  
→ [Full Documentation](clusters/studio-cluster.md)

**Pi Cluster** - Edge computing, IoT integration, distributed sensors  
→ [Full Documentation](clusters/pi-cluster.md)

**Forge Cluster** - GPU training, ML inference, VLLM, PyTorch, Flyte  
→ [Full Documentation](clusters/forge-cluster.md)

---

## AI Agent Architecture 🤖 (Serverless)

The homelab implements a **production-grade AI Agent architecture** as **Knative services** for optimal resource utilization and event-driven integration:

### Architecture Components

- **Small Language Models (SLMs)** - Ollama on Forge for fast, specialized tasks
- **Knowledge Graph** - LanceDB (planned) for RAG context and team knowledge
- **Large Language Models (LLMs)** - VLLM serving Llama 3.1 70B on Forge
- **Agent Orchestration** - Flyte workflows for complex multi-step tasks
- **Event Platform** - Knative Eventing + RabbitMQ for CloudEvents

### AI Agents (Knative Services)

All agents are deployed as Knative services with **scale-to-zero** capability:

| Agent | Type | Purpose | Scale Behavior |
|-------|------|---------|----------------|
| `agent-bruno` | Knative Service | Main AI assistant | 0→N auto-scale |
| `agent-auditor` | Knative Service | Event-driven infra response | 0→N auto-scale |
| `agent-jamie` | Knative Service | DevOps automation | 0→N auto-scale |
| `agent-mary-kay` | Knative Service | Sales assistant | 0→N auto-scale |

**Key Benefits:**
- **Zero Cost When Idle**: Agents consume 0 CPU/Memory when not in use (80% resource savings)
- **Event-Driven**: Native CloudEvents integration with Prometheus, GitHub, Slack
- **Auto-Scaling**: 0→N in <30s, cold start ~5s, warm start ~200ms
- **Event Flow**: CloudEvent → RabbitMQ → Knative Trigger → Agent (0→1) → Process → Return → Scale to 0

**Deployment:**
```bash
# Deploy via Flux (GitOps)
flux reconcile helmrelease agent-bruno -n ai-agents

# Verify
kubectl get ksvc -n ai-agents
kubectl get pods -n ai-agents  # Shows 0 when idle
```

**→ [See AI Agent Architecture](architecture/ai-agent-architecture.md)** - Complete deployment patterns, YAML, operations  
**→ [See AI Components](architecture/ai-components.md)** - Technical specs  
**→ [See AI Connectivity](architecture/ai-connectivity.md)** - How AI connects teams/clusters/edge  
**→ [See Studio Cluster](clusters/studio-cluster.md)** - Production deployment details

---

## Technology Stack

**Infrastructure**:
- Pulumi (IaC), Flux (GitOps), External Secrets Operator (secrets via GitHub), cert-manager (certificates)

**Service Mesh**:
- Linkerd multi-cluster, WARP tunnels (Phase 1), WireGuard (Phase 2+)

**Serverless Platform**:
- Knative Serving (auto-scaling), Knative Eventing (CloudEvents)
- RabbitMQ (event bus), Kaniko (secure builds)
- All AI agents deployed as Knative services

**AI/ML**:
- VLLM (Forge), Ollama (Forge/Pi), LanceDB (planned), Flyte, PyTorch
- AI agents: Knative services with scale-to-zero

**Observability**:
- Prometheus, Grafana, Loki, Tempo, AlertManager (Phase 1)

**CI/CD** (Phase 1):
- GitHub Actions, Trivy, Velero, Chaos Mesh

**→ [See Production Readiness](analysis/production-readiness.md) for current status**

---

## Current Status

**Production Readiness: 62% → Target: 94%**

**→ [See Production Readiness Analysis](analysis/production-readiness.md)** - Complete scorecard, gap analysis, and Phase 1 roadmap

---

## Documentation Index

**Architecture**:
- [AI Agent Architecture](architecture/ai-agent-architecture.md) - SLM + Knowledge Graph + LLM pattern
- [AI Components](architecture/ai-components.md) - Technical specifications
- [Agent Orchestration](architecture/agent-orchestration.md) - Workflows and coordination
- [AI Connectivity](architecture/ai-connectivity.md) - Teams, clusters, edge connectivity
- [MCP Observability](architecture/mcp-observability.md) - Model Context Protocol integration

**Clusters**:
- [Air](clusters/air-cluster.md) | [Pro](clusters/pro-cluster.md) | [Studio](clusters/studio-cluster.md) | [Pi](clusters/pi-cluster.md) | [Forge](clusters/forge-cluster.md)

**Operations**:
- [Node Labels](operations/node-labels-reference.md) | [Port Mapping](operations/port-mapping-strategy.md) | [Migration](operations/migration-guide.md) | [Affinity](operations/pod-affinity-examples.md)

**Analysis**:
- [Network Engineering](analysis/network-engineering-analysis.md) | [DevOps Engineering](analysis/devops-engineering-analysis.md) | [Production Readiness](analysis/production-readiness.md)

**Implementation**:
- [Roadmap](implementation/operational-maturity-roadmap.md) | [Phase 1](implementation/phase1-implementation.md) | [Phase 2](implementation/phase2-preparation.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

