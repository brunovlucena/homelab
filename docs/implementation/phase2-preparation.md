# Phase 2 Preparation - Brazil Regional Expansion

> **Part of**: [Homelab Documentation](../README.md) → Implementation  
> **Last Updated**: November 7, 2025

---

## Status

🚧 **Blocked until Phase 1 complete**

## Overview

Planning and preparation for Brazil regional expansion with 4-5 Studio clusters across Brazilian regions.

## Prerequisites

Before starting Phase 2, all of the following must be complete:

1. ✅ Phase 1 complete (94% production readiness)
2. ✅ Operational maturity demonstrated
3. ✅ 30 days of stable operations
4. ✅ All runbooks tested and validated
5. ✅ DevOps Engineer hired (recommended)
6. ✅ Successful DR drill executed
7. ✅ Security audit passed

## Target Regions

### 1. Studio-Sul (Porto Alegre)

**Role**: Regional hub for Southern Brazil

**Hardware**:
- Mac Studio (M2 Ultra, 192GB RAM)
- 12 nodes (similar to current Studio)

**Network**: 10.16.0.0/16

**Purpose**:
- Primary hub for Brazil region
- AI agent hosting
- Observability hub
- Development environment

**Latency Targets**:
- Studio-Primary → Studio-Sul: 25-35ms
- Studio-Sul → Other Brazil regions: 50-100ms

---

### 2. Studio-NE (Recife)

**Role**: Northeast Brazil coverage

**Hardware**:
- Mac Mini or Mac Studio (depending on scale)
- 6-8 nodes

**Network**: 10.17.0.0/16

**Purpose**:
- Regional AI inference
- Edge gateway for NE sensors
- Local data processing (LGPD compliance)

**Latency Targets**:
- Studio-Sul → Studio-NE: 60-80ms
- Studio-NE → Users in NE: <20ms

---

### 3. Studio-CO (Brasília)

**Role**: Central-West Brazil coverage

**Hardware**:
- Mac Mini or Mac Studio
- 6-8 nodes

**Network**: 10.18.0.0/16

**Purpose**:
- Government/enterprise workloads
- Regional inference
- Backup site for critical services

**Latency Targets**:
- Studio-Sul → Studio-CO: 50-70ms
- Studio-CO → Users in CO: <20ms

---

### 4. Studio-N (Manaus)

**Role**: North Brazil coverage (optional)

**Hardware**:
- Mac Mini
- 4-6 nodes

**Network**: 10.19.0.0/16

**Purpose**:
- Remote region coverage
- IoT hub for Amazon sensors
- Edge inference

**Latency Targets**:
- Studio-Sul → Studio-N: 80-100ms
- Studio-N → Users in N: <20ms

---

## Architecture Changes Required

### 1. Hub-and-Spoke Topology

**Current**: WARP tunnels (full mesh)

**Target**: Hub-and-spoke with 3 tiers

```
Tier 1 (Global Hubs):
  Studio-Primary (São Paulo) ← Primary Hub
  Studio-Sul (Porto Alegre)  ← Brazil Hub

Tier 2 (Regional Hubs):
  Studio-NE (Recife)
  Studio-CO (Brasília)
  Studio-N (Manaus)

Tier 3 (Edge):
  Pi clusters, IoT gateways
```

**Benefits**:
- O(n) scalability instead of O(n²)
- Reduced complexity
- Lower bandwidth costs

### 2. Revised CIDR Allocation

**Current**: Non-contiguous blocks

**Target**: Contiguous /12 blocks for route summarization

```yaml
Local (Phase 1):
  Base: 10.0.0.0/12  # 16× /16 subnets
  Air: 10.0.0.0/16
  Pro: 10.1.0.0/16
  Studio: 10.2.0.0/16
  Pi: 10.3.0.0/16
  Forge: 10.4.0.0/16
  Reserved: 10.5-15.0.0/16

Brazil (Phase 2):
  Base: 10.16.0.0/12  # 16× /16 subnets
  Studio-Sul: 10.16.0.0/16
  Studio-NE: 10.17.0.0/16
  Studio-CO: 10.18.0.0/16
  Studio-N: 10.19.0.0/16
  Reserved: 10.20-31.0.0/16
```

### 3. WireGuard/Tailscale

**Current**: Cloudflare WARP tunnels (dev/test)

**Target**: WireGuard or Tailscale (production)

**Rationale**:
- Better performance
- Lower latency
- More control
- Production-grade SLA

### 4. Thanos

**Purpose**: Global Prometheus federation

**Architecture**:
```
Studio-Primary (Thanos Query Frontend)
    ↓
    ├── Thanos Sidecar (Studio-Sul)
    ├── Thanos Sidecar (Studio-NE)
    ├── Thanos Sidecar (Studio-CO)
    └── Thanos Sidecar (Studio-N)
    ↓
Thanos Store (Long-term storage in MinIO)
```

**Benefits**:
- Global metrics view
- Long-term retention
- Reduced Prometheus overhead

### 5. Centralized Forge

**Current**: Single Forge cluster (local)

**Target**: 1-2 Forge GPU hubs

```yaml
Forge-BR (São Paulo):
  Hardware: 8× A100 GPUs
  Purpose: Training + Inference for Brazil/Americas
  Cost: ~$9k/month

Regional Inference:
  Hardware: CPU or small GPU (T4)
  Purpose: Low-latency inference per region
  Cost: ~$4k/month

Total: $13k/month (vs $25k/month for per-region Forge)
```

---

## Additional Tooling Required

### Rancher

**Purpose**: Multi-cluster management UI

**Features**:
- Centralized cluster view
- RBAC management
- Policy enforcement
- Application catalog

### Backstage

**Purpose**: Developer portal

**Features**:
- Service catalog
- Documentation hub
- API reference
- Deployment tracking

### OpenCost

**Purpose**: Multi-cluster cost visibility

**Features**:
- Per-cluster cost breakdown
- Resource utilization
- Cost allocation

### Crossplane

**Purpose**: Infrastructure provisioning

**Features**:
- Declarative infrastructure
- Cloud provider abstraction
- GitOps-driven

### Teleport

**Purpose**: Secure remote access

**Features**:
- Zero-trust access
- Session recording
- MFA enforcement
- Audit logging

---

## Timeline & Cost

### Timeline

**Estimated**: 16-24 weeks after Phase 1 complete

**Breakdown**:
- Week 1-4: Infrastructure provisioning
- Week 5-8: Network setup (WireGuard, hub-and-spoke)
- Week 9-12: Tooling deployment (Rancher, Thanos, Backstage)
- Week 13-16: Service migration and testing
- Week 17-20: Production cutover
- Week 21-24: Optimization and tuning

### Cost

**Estimated**: $3k-6k/month

**Breakdown**:
- Hardware (Mac Studios/Minis): $15k-30k upfront
- Colocation/Hosting: $500-1000/month per region
- Network (WireGuard/Tailscale): $100-300/month
- Monitoring (Grafana Cloud, PagerDuty): $200-500/month
- Backups (S3, GCS): $200-500/month
- Total: $3k-6k/month operational + $15k-30k upfront

---

## Next Steps

1. ✅ Complete Phase 1 (12-16 weeks)
2. ⏸️ Evaluate results and lessons learned
3. ⏸️ Hire DevOps Engineer (if proceeding to Phase 2)
4. ⏸️ Secure hardware (Mac Studios/Minis)
5. ⏸️ Setup colocation/hosting in target regions
6. ⏸️ Begin Phase 2 deployment

---

## Risk Assessment

### Technical Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Network latency >200ms | High | Low | Use WireGuard, optimize routing |
| Hub-and-spoke single point of failure | Critical | Medium | Deploy backup hubs |
| CIDR exhaustion | Medium | Low | Use /12 blocks, proper allocation |
| GPU shortage | High | Medium | Order hardware early, consider cloud |
| Thanos complexity | Medium | High | Hire expertise, extensive testing |

### Operational Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Solo SRE overload | Critical | High | Hire DevOps Engineer |
| Regional outages | High | Medium | Multi-region failover |
| LGPD compliance issues | Critical | Low | Regional data residency |
| Cost overrun | Medium | Medium | Monthly budget reviews |

---

## Success Criteria

Phase 2 is successful when:

1. ✅ All 4 regions operational (Studio-Sul, NE, CO, N)
2. ✅ Hub-and-spoke topology stable
3. ✅ Latency targets met (<100ms intra-Brazil)
4. ✅ Thanos operational with 90-day retention
5. ✅ LGPD compliance validated
6. ✅ Production readiness >90% across all regions
7. ✅ Operational cost within budget ($3k-6k/month)
8. ✅ 30 days stable operations

---

## Related Documentation

- [Operational Maturity Roadmap](operational-maturity-roadmap.md)
- [Phase 1 Implementation](phase1-implementation.md)
- [Network Engineering Analysis](../analysis/network-engineering-analysis.md)
- [DevOps Engineering Analysis](../analysis/devops-engineering-analysis.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

