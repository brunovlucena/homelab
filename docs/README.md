# 🏗️ Homelab Documentation

> **Purpose**: Connecting Teams, Clusters & Edge Devices in a Unified Platform  
> **Last Updated**: November 19, 2025  
> **SRE**: Bruno Lucena | **IaC**: Pulumi  
> **Production Readiness**: 65% → Target: 94%  
> **ML Readiness**: 48% → Target: 81%

---

## 📚 Documentation Structure

All documentation is organized into five main categories:

### 🎯 Architecture

Core architectural patterns and designs:

- [AI Agent Architecture](architecture/ai-agent-architecture.md) - **Knative services deployment** (scale-to-zero, event-driven) ⭐
- [AI Components](architecture/ai-components.md) - Detailed component breakdown
- [AI Connectivity](architecture/ai-connectivity.md) - Cross-cluster communication patterns
- [MCP Observability](architecture/mcp-observability.md) - Model Context Protocol for observability
- [Agent Orchestration](architecture/agent-orchestration.md) - Agent coordination and management

### ☸️ Clusters

Individual cluster documentation:

- [Air Cluster](clusters/air-cluster.md) - Experimental & CI/CD (4 nodes, Kind)
- [Pro Cluster](clusters/pro-cluster.md) - Development & Testing (7 nodes, Kind)
- [Studio Cluster](clusters/studio-cluster.md) - Production AI Agents (12 nodes, Kind) ⭐
- [Pi Cluster](clusters/pi-cluster.md) - Edge & IoT (3-6 nodes, k3s)
- [Forge Cluster](clusters/forge-cluster.md) - GPU AI Training & Inference (8 nodes, k3s) 🤖

### 🔧 Operations

Day-to-day operational guides:

- [Node Labels Reference](operations/node-labels-reference.md) - All node labels and their purposes
- [Port Mapping Strategy](operations/port-mapping-strategy.md) - Port ranges and allocations
- [Migration Guide](operations/migration-guide.md) - Cluster migration procedures
- [Pod Affinity Examples](operations/pod-affinity-examples.md) - Workload placement patterns
- [Local Registry](operations/local-registry.md) - Local container registry setup

### 📊 Analysis

Technical assessments and reviews:

- [Network Engineering Analysis](analysis/network-engineering-analysis.md) - Network feasibility, CIDR, topology
- [DevOps Engineering Analysis](analysis/devops-engineering-analysis.md) - Operational maturity assessment (65% readiness)
- [ML Engineering Analysis](analysis/ml-engineering-analysis.md) - ML lifecycle and MLOps assessment (48% readiness)
- [Data Engineering Analysis](analysis/data-engineering-analysis.md) - Data pipeline and storage architecture (42% readiness)
- [QA Engineering Analysis](analysis/qa-engineering-analysis.md) - Testing and quality assurance
- [SRE Technical Analysis](analysis/sre-technical-analysis.md) - Site reliability engineering assessment
- [Production Readiness](analysis/production-readiness.md) - Current status and blockers

### 📈 Reports

Progress and status reports:

- [Weekly Progress Report](reports/homelab-report.md) - Development progress and updates
- [Knative Lambda Report](reports/knative-lambda-report.md) - Serverless platform development
- [JIRA](reports/jira.md) - Issue tracking and project management

### 🚀 Implementation

Roadmaps and execution plans:

- [Operational Maturity Roadmap](implementation/operational-maturity-roadmap.md) - Phase 1 (12-16 weeks)
- [Phase 1 Implementation](implementation/phase1-implementation.md) - Week-by-week breakdown
- [Phase 2 Preparation](implementation/phase2-preparation.md) - Brazil regional expansion

---

## 🚀 Quick Start

### For Developers

```bash
# Deploy to Air cluster (experimental)
make up ENV=air

# Deploy to Pro cluster (development)
make up ENV=pro
```

### For SRE

```bash
# Deploy all clusters
make up

# Reconcile Flux
make reconcile-all

# Check multi-cluster connectivity
linkerd multicluster gateways
```

### For AI/ML Engineers

```bash
# Access VLLM on Forge
curl http://vllm.ml-inference.svc.forge.remote:8000/health

# Submit training job to Flyte
flytectl create execution --project homelab --domain production
```

---

## 📌 Key Concepts

### Main Purpose: Connecting Everything

The homelab connects three pillars:

1. **🧑‍💻 Teams** - Developers, SREs, Data Scientists collaborating
2. **☸️ Clusters** - 5 Kubernetes clusters (Air, Pro, Studio, Pi, Forge)
3. **📡 Edge Devices** - Raspberry Pi, IoT sensors, gateways

### Technology Stack

- **Service Mesh**: Linkerd (multi-cluster mTLS)
- **GitOps**: Flux (automated deployments)
- **IaC**: Pulumi (infrastructure as code)
- **Secret Management**: External Secrets Operator with GitHub backend
- **Certificate Management**: cert-manager
- **Networking**: WARP tunnels (Phase 1), WireGuard (Phase 2+)
- **AI/ML**: VLLM (Forge), Ollama, LanceDB (planned)

### Current Infrastructure

```text
5 Clusters Connected via Linkerd Service Mesh:

┌─────────────────────────────────────────────┐
│  Mac Studio (M2 Ultra, 192GB)              │
│  ├─ Air (Kind) - 4 nodes                   │
│  ├─ Pro (Kind) - 7 nodes                   │
│  └─ Studio (Kind) - 12 nodes ⭐ Production │
└─────────────────────────────────────────────┘
         │
         ├── WARP Tunnels (encrypted)
         │
┌────────┴──────────┐         ┌───────────────┐
│ Raspberry Pi      │         │ GPU Server    │
│ Pi (k3s)          │         │ Forge (k3s)   │
│ 3-6 nodes         │         │ 8 nodes       │
│ Edge/IoT 📡       │         │ AI Training 🤖│
└───────────────────┘         └───────────────┘
```

---

## 🎯 Current Status

### Production Readiness: 62%

**→ [See Production Readiness Analysis](analysis/production-readiness.md)** - Complete status, gap analysis, and roadmap

---

## 🔗 Quick Links

### By Role

**Developers**:

- [Air Cluster](clusters/air-cluster.md) - Experimental environment
- [Pro Cluster](clusters/pro-cluster.md) - Development environment
- [Migration Guide](operations/migration-guide.md) - Moving between clusters

**SREs**:

- [Production Readiness](analysis/production-readiness.md) - Current status
- [Weekly Progress Report](reports/homelab-report.md) - Development progress updates
- [Operational Maturity Roadmap](implementation/operational-maturity-roadmap.md) - Phase 1 plan
- [Operations Guides](operations/node-labels-reference.md) - Day-to-day operations

**AI/ML Engineers**:

- [Forge Cluster](clusters/forge-cluster.md) - GPU infrastructure
- [AI Agent Architecture](architecture/ai-agent-architecture.md) - Agent design
- [AI Components](architecture/ai-components.md) - Component details

**Network Engineers**:

- [Network Engineering Analysis](analysis/network-engineering-analysis.md) - Network design
- [CIDR Allocation](analysis/network-engineering-analysis.md#1-cidr-allocation---improved-design) - IP planning

**DevOps Engineers**:

- [DevOps Engineering Analysis](analysis/devops-engineering-analysis.md) - Maturity assessment
- [Data Engineering Analysis](analysis/data-engineering-analysis.md) - Data pipeline architecture
- [Phase 1 Implementation](implementation/phase1-implementation.md) - Week-by-week plan

**QA Engineers**:

- [QA Engineering Analysis](analysis/qa-engineering-analysis.md) - Testing and quality assurance
- [Knative Lambda Report](reports/knative-lambda-report.md) - Serverless platform testing

**Data Engineers**:

- [Data Engineering Analysis](analysis/data-engineering-analysis.md) - Data pipeline and storage
- [Forge Cluster](clusters/forge-cluster.md) - GPU infrastructure for ML workloads

---

## 🤝 Contributing

This homelab is managed by an SRE Engineer (solo) with AI assistance. For questions or suggestions:

1. Check relevant documentation section above
2. Review implementation roadmaps
3. Consult analysis documents for technical details

---

## 📝 Document Conventions

- **📁 Category icons**: 🎯 Architecture, ☸️ Clusters, 🔧 Operations, 📊 Analysis, 🚀 Implementation
- **Status indicators**: ✅ Complete, ⚠️ In Progress, ❌ Blocked, 🚧 Planned
- **Code blocks**: Include cluster context and service endpoints
- **Cross-references**: Use relative links between documents

---

**Last Updated**: November 19, 2025  
**Maintained by**: AI Engineer (Bruno Lucena)  
**Next Review**: Phase 1 Week 4
