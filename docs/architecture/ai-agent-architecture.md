# 🤖 AI Agent Architecture

> **Part of**: [Homelab Cluster Architecture](../CLUSTER_ARCHITECTURE.md)  
> **Related**: [Multi-Cluster Mesh](multi-cluster-mesh.md) | [Forge Cluster](../clusters/forge-cluster.md) | [Studio Cluster](../clusters/studio-cluster.md)  
> **Last Updated**: November 7, 2025

---

## Overview

The homelab implements a **production-grade AI Agent architecture** that combines Small Language Models (SLMs), Knowledge Graphs, and Large Language Models (LLMs) to create intelligent, context-aware agents.

**All agents deployed as Knative services** for scale-to-zero and event-driven architecture.

This document provides an overview of the architecture. For detailed information, see:

- [🔧 AI Components](ai-components.md) - Technical specifications for SLMs, LLMs, and Knowledge Graph
- [🎯 Agent Orchestration](agent-orchestration.md) - Agent logic, workflows, and examples
- [🌐 AI Connectivity](ai-connectivity.md) - How AI connects teams, clusters, and edge devices
- [📊 MCP Observability](mcp-observability.md) - Model Context Protocol and observability integration

---

## Design Pattern: SLM + Knowledge Graph + LLM

```
┌────────────────────────────────────────────────────────────────────┐
│                    AI Agent Architecture Pattern                   │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  ┌────────────────┐                                                │
│  │ Small Language │         ┌─────────────────┐                    │
│  │    Models      │────────►│   Knowledge     │                    │
│  │    (SLMs)      │         │     Graph       │                    │
│  │                │         │                 │                    │
│  │ • Ollama       │         │ • LanceDB       │                    │
│  │ • Llama 3      │◄────────│ • Vector Store  │                    │
│  │ • CodeLlama    │         │ • RAG Context   │                    │
│  │ • Mistral      │         │ • Metadata      │                    │
│  └────────┬───────┘         └────────┬────────┘                    │
│           │                          │                             │
│           │         ┌────────────────┘                             │
│           │         │                                              │
│           │         │                                              │
│  ┌────────▼─────────▼──────┐         ┌─────────────────┐           │
│  │                         │         │   Workflow      │           │
│  │        Agent            │◄────────│   Orchestration │           │
│  │     Orchestrator        │         │   (Flyte)       │           │
│  │                         │         │                 │           │
│  │ • agent-bruno           │         │ • Task Graphs   │           │
│  │ • agent-auditor         │         │ • Scheduling    │           │
│  │ • agent-jamie           │         │ • Pipelines     │           │
│  │ • agent-mary-kay        │─────────►                 │           │
│  └────────┬────────────────┘         └─────────────────┘           │
│           │                                                        │
│           │                                                        │
│  ┌────────▼─────────────────┐         ┌─────────────────┐          │
│  │                          │         │   Security      │          │
│  │    Large Language        │         │   & Auth        │          │
│  │       Model              │◄────────│                 │          │
│  │       (LLM)              │         │ • ESO (GitHub)  │          │
│  │                          │         │ • mTLS          │          │
│  │ • VLLM                   │         │ • RBAC          │          │
│  │ • Llama 3.1 70B          │         │                 │          │
│  │ • Tensor Parallel        │─────────►                 │          │
│  │ • GPU-accelerated        │         └─────────────────┘          │
│  └──────────────────────────┘                                      │
│           │                              ┌─────────────────┐       │
│           │                              │ Observability   │       │
│           └──────────────────────────────►                 │       │
│                                          │ • Prometheus    │       │
│                                          │ • Grafana       │       │
│                                          │ • Loki          │       │
│                                          │ • Tempo         │       │
│                                          └─────────────────┘       │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

---

## Architecture Layers

### 0. Deployment Layer: Knative Services

**All AI agents are deployed as Knative services** for scale-to-zero and event-driven integration.

#### Architecture

```text
CloudEvent → RabbitMQ → Knative Trigger → Agent (0→1) → Process → Return → Scale to 0

Benefits:
• 80% resource savings (scale-to-zero)
• Cold start: ~5s, Warm start: ~200ms  
• Native CloudEvents integration
• Auto-scaling based on load
```

#### Deployment Example

```yaml
# Knative Service
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: agent-bruno
  namespace: ai-agents
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "0"
        autoscaling.knative.dev/maxScale: "10"
    spec:
      containers:
      - name: agent
        image: agent-bruno:v1.0.0
        env:
        - name: OLLAMA_URL
          value: "http://ollama.ml-inference.svc.forge.remote:11434"
        - name: VLLM_URL
          value: "http://vllm.ml-inference.svc.forge.remote:8000"
```

```yaml
# Knative Trigger (Event Routing)
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: agent-bruno-alerts
  namespace: ai-agents
spec:
  broker: default
  filter:
    attributes:
      type: io.prometheus.alert
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: agent-bruno
```

#### Operations

```bash
# Deploy
flux reconcile helmrelease agent-bruno -n ai-agents

# Monitor
kubectl get ksvc -n ai-agents
kubectl get pods -n ai-agents -w  # Watch scaling

# Trigger via CloudEvent
curl -X POST http://agent-bruno.ai-agents.svc.cluster.local/events \
  -H "Content-Type: application/cloudevents+json" \
  -d '{"specversion": "1.0", "type": "query.request", "source": "cli", "id": "001", "data": {"query": "status"}}'
```

**→ [See Studio Cluster](../clusters/studio-cluster.md)** for production deployment details

---

### 1. Foundation Layer: MCP Server

The **Model Context Protocol (MCP) Server** acts as the foundation layer, providing structured access to observability data.

```
Teams → AI Agents → MCP Server → Observability Stack
         ↓              ↑
    (Natural Lang) (Structured API)
         
Advanced Users can bypass agents and use MCP directly
```

**Benefits**:
- **For Teams**: Natural language via agents
- **For SREs**: Direct structured access via MCP
- **For Agents**: Reliable tool access
- **For System**: Single source of truth

See: [📊 MCP Observability](mcp-observability.md)

### 2. Data Layer: Knowledge Graph

**LanceDB** provides centralized knowledge storage, relationships, and context for AI agents.

**Collections**:
- `homelab-docs` - Documentation and best practices
- `incident-history` - Operational incidents and resolutions
- `code-snippets` - Tested code examples
- `team-knowledge` - Collaborative knowledge base

See: [🔧 AI Components - Knowledge Graph](ai-components.md#knowledge-graph)

### 3. Intelligence Layer: SLMs + LLMs

**Small Language Models (SLMs)**:
- Fast, specialized tasks (classification, extraction)
- Deployed on Forge cluster via Ollama
- Response time: <100ms

**Large Language Models (LLMs)**:
- Complex reasoning and analysis
- Deployed on Forge cluster via VLLM
- GPU-accelerated (Llama 3.1 70B)

See: [🔧 AI Components](ai-components.md)

### 4. Orchestration Layer: Agents

**Agent Orchestrator** coordinates between SLMs, Knowledge Graph, and LLMs.

**Deployed Agents**:
- `agent-bruno` (30120) - General purpose assistant
- `agent-auditor` (30121) - SRE/DevOps automation
- `agent-jamie` (30122) - Data science workflows
- `agent-mary-kay` (30127) - Customer interaction

See: [🎯 Agent Orchestration](agent-orchestration.md)

### 5. Security Layer

- **Vault**: Secret management
- **mTLS**: Linkerd service mesh (cluster-wide)
- **RBAC**: Kubernetes ServiceAccounts per agent

---

## Key Features

### Efficiency
- **Fast SLMs** for simple tasks (classification, extraction)
- **Powerful LLMs** only when needed (complex reasoning)
- **Result**: 10x lower cost, 5x lower latency

### Intelligence
- **Knowledge Graph** accumulates institutional knowledge
- **Continuous Learning** from all interactions
- **Result**: Agents get smarter over time

### Connectivity
- **Teams** access via natural language
- **Clusters** orchestrated intelligently
- **Edge** integrated seamlessly
- **Result**: Unified platform for all stakeholders

See: [🌐 AI Connectivity](ai-connectivity.md)

### Observability
- **All interactions** logged and traced
- **Metrics** for every component
- **Result**: Full visibility into AI decision-making

See: [📊 MCP Observability](mcp-observability.md)

### Security
- **Zero-trust** via Linkerd mTLS
- **Secrets** managed by Vault
- **RBAC** per agent
- **Result**: Production-grade security

---

## Implementation Status

```yaml
Current (Phase 1):
├─ ✅ Forge Cluster with GPU (VLLM deployed)
├─ ✅ Studio Cluster (agents deployed)
├─ ✅ Linkerd multi-cluster (service mesh)
├─ ⚠️ Ollama (planned for Forge)
├─ ⚠️ LanceDB (planned for Studio)
└─ ⚠️ Knowledge Graph integration (Phase 1 Week 8-12)

Planned (Phase 1 - Next 12 weeks):
├─ Week 1-4: Deploy Ollama on Forge
├─ Week 4-8: Deploy LanceDB + embeddings
├─ Week 8-12: Agent-Knowledge Graph integration
└─ Week 12: Full RAG pipeline operational

Future (Phase 2+):
├─ Edge SLMs on Pi cluster
├─ Regional Knowledge Graph replication
└─ Multi-region agent coordination
```

---

## Documentation Index

### Core Architecture
- [🔧 AI Components](ai-components.md) - SLMs, LLMs, Knowledge Graph, Flyte
- [🎯 Agent Orchestration](agent-orchestration.md) - Agent logic and workflows
- [🌐 AI Connectivity](ai-connectivity.md) - Teams, clusters, edge integration
- [📊 MCP Observability](mcp-observability.md) - Monitoring and MCP server

### Related Infrastructure
- [🏗️ Main Architecture](../CLUSTER_ARCHITECTURE.md)
- [🔗 Multi-Cluster Mesh](multi-cluster-mesh.md)
- [⚙️ Forge Cluster](../clusters/forge-cluster.md)
- [🎯 Studio Cluster](../clusters/studio-cluster.md)
- [🚀 Implementation Roadmap](../implementation/operational-maturity-roadmap.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

