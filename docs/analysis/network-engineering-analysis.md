# Network Engineering Analysis

> **Part of**: [Homelab Documentation](../README.md) → Analysis  
> **Last Updated**: November 7, 2025

---

## Executive Summary

The geo-distributed architecture is feasible but requires modifications for scalability. Phase 1 (local 5 clusters) is production-ready with 95% confidence. Phase 2 (Brazil regional) is highly feasible with modifications (85% confidence). Phase 3 (global) requires significant investment (70% confidence).

## Key Findings

### Phase 1 (Local 5 Clusters): ✅ Production-Ready (95% confidence)

**Status**: Currently operational

**Infrastructure**:
- 5 clusters: Air, Pro, Studio, Pi, Forge
- WARP tunnels for connectivity
- Linkerd multi-cluster service mesh
- Low latency (<5ms local)

**Recommendation**: ✅ Execute now - already operational

**Risk Level**: Low

---

### Phase 2 (Brazil Regional): ⚠️ Highly Feasible with Modifications (85% confidence)

**Target**: 4-5 Studio clusters across Brazil

**Expected Latency**: 73-195ms cross-region

**Estimated Cost**: $3k-6k/month

**Critical Changes Required**:

1. **Hub-and-Spoke Topology** (not full mesh)
2. **Revised CIDR Allocation** (contiguous /12 blocks)
3. **WireGuard/Tailscale** instead of WARP for production
4. **Optimized GPU Economics** (1-2 Forge hubs, not per-region)

**Recommendation**: ⚠️ Plan carefully, implement modifications

**Risk Level**: Medium (manageable with proper planning)

---

### Phase 3 (Global): 🚧 Requires Significant Investment (70% confidence)

**Target**: 20-25 clusters across continents

**Expected Latency**: 200-365ms inter-continental

**Estimated Cost**: $15k-25k/month

**Timeline**: 24-36 months

**Recommendation**: 🚧 Defer until Phase 2 proves itself

**Risk Level**: High (significant investment required)

---

## Critical Technical Issues Identified

### 1. CIDR Allocation - Improved Design

#### Problem

Original non-contiguous CIDR blocks prevent route summarization, causing:
- Large routing tables
- Inefficient BGP advertisements
- Difficult troubleshooting
- Scalability limitations

#### Solution

Use contiguous /12 blocks for route summarization:

```yaml
# New CIDR Allocation Strategy

Local (Phase 1):
  Base: 10.0.0.0/12  # Summarizes to single /12 route
  Air: 10.0.0.0/16
  Pro: 10.1.0.0/16
  Studio: 10.2.0.0/16
  Pi: 10.3.0.0/16
  Forge: 10.4.0.0/16
  Reserved: 10.5-15.0.0/16  # Future local clusters

Brazil (Phase 2):
  Base: 10.16.0.0/12  # Summarizes to single /12 route
  Studio-Sul: 10.16.0.0/16
  Studio-NE: 10.17.0.0/16
  Studio-CO: 10.18.0.0/16
  Studio-N: 10.19.0.0/16
  Reserved: 10.20-31.0.0/16  # Future Brazil clusters

Americas (Phase 3):
  Base: 10.32.0.0/12  # US, Canada, Mexico
  Reserved: 10.32-47.0.0/16

Europe (Phase 3):
  Base: 10.48.0.0/12  # UK, Germany, France
  Reserved: 10.48-63.0.0/16

Asia-Pacific (Phase 3):
  Base: 10.64.0.0/12  # Japan, Singapore, Australia
  Reserved: 10.64-79.0.0/16

Forge GPU Hubs (Global):
  Base: 10.80.0.0/12
  Forge-BR: 10.80.0.0/16
  Forge-US: 10.81.0.0/16
  Forge-EU: 10.82.0.0/16
  Reserved: 10.83-95.0.0/16
```

#### Benefits

- **Route Summarization**: Each region becomes a single /12 route
- **Scalability**: Support for 16× /16 subnets per region
- **Troubleshooting**: Easier to identify regions by CIDR
- **BGP Efficiency**: Fewer routes to advertise

---

### 2. Topology Optimization - Hub-and-Spoke

#### Problem

Full mesh doesn't scale:
- **Connections**: O(n²) complexity
- **Example**: 20 clusters = 190 tunnels
- **Bandwidth**: Exponential growth
- **Complexity**: Difficult to manage

#### Solution

3-tier hierarchical hub-and-spoke topology:

```
Tier 1 (Global Hubs):
  Studio-Primary (São Paulo) ← Primary Hub
  Studio-US (Miami)         ← Americas Hub
  Studio-EU (London)        ← Europe Hub

Tier 2 (Regional Hubs):
  Brazil: Studio-Sul, Studio-NE, Studio-CO, Studio-N
  Development: Pro, Air
  Edge: Pi

Tier 3 (Edge):
  Pi clusters
  IoT gateways
  Sensors

Connections: O(n) instead of O(n²)
```

#### Routing Rules

```yaml
# Tier 3 → Tier 2 (always direct)
Pi → Pro (direct)
IoT → Studio-Sul (direct)

# Tier 2 → Tier 1 (hub uplink)
Studio-NE → Studio-Sul → Studio-Primary

# Tier 1 → Tier 1 (hub-to-hub)
Studio-Primary ↔ Studio-US ↔ Studio-EU (mesh between hubs only)
```

#### Benefits

- **Scalability**: Linear growth (O(n))
- **Reduced Bandwidth**: Fewer connections
- **Lower Cost**: Less infrastructure
- **Easier Management**: Clearer routing paths

#### Trade-offs

- **Latency**: +10-20ms for multi-hop
- **Hub Criticality**: Single point of failure (mitigated by backup hubs)

---

### 3. GPU Economics - Realistic Cost Model

#### Problem

Multiple Forge clusters = unsustainable costs:
- 5 Forge clusters × $5k/month = $25k/month
- Underutilized GPUs in each region
- Complex management

#### Solution

Centralized training + regional inference:

```yaml
Optimized Architecture:

  Training (Centralized):
    Locations:
      - Forge-BR (São Paulo)
      - Forge-US (Miami) [optional]
    
    Hardware per hub:
      - 8× NVIDIA A100 GPUs (40GB)
      - High-end CPU (64+ cores)
      - 512GB+ RAM
      - 10Gbps network
    
    Purpose:
      - Model training
      - Fine-tuning
      - Batch inference
    
    Cost: ~$9k/month per hub
    Total (2 hubs): $18k/month

  Inference (Regional):
    Locations:
      - Studio-Sul, NE, CO, N (Brazil)
      - Studio-US, EU (Phase 3)
    
    Hardware per region:
      - CPU-based (high-core count)
      - OR small GPU (T4, 16GB)
      - 128GB RAM
      - 1Gbps network
    
    Purpose:
      - Real-time inference
      - Low-latency responses
      - Edge ML
    
    Cost: ~$500-1000/month per region
    Total (4 regions): $2k-4k/month

  Total Cost: $20k-22k/month (vs $25k/month with per-region Forge)
  Savings: 12-20%
```

#### Benefits

- **Lower Cost**: Shared training infrastructure
- **Better Utilization**: GPUs used 24/7 via time-zone sharing
- **Easier Management**: Centralized model updates
- **Regional Performance**: Low-latency inference

#### Trade-offs

- **Training Latency**: Need to send data to central hub
- **Hub Criticality**: Training bottleneck (mitigated by 2 hubs)

---

### 4. Tunnel Technology - Production vs Development

#### Current (Phase 1)

**Technology**: Cloudflare WARP tunnels

**Pros**:
- Easy to set up
- Free for development
- Good for testing

**Cons**:
- Not production-grade
- Limited SLA
- Less control over routing

**Verdict**: ✅ Fine for Phase 1 (local development)

#### Target (Phase 2+)

**Technology**: WireGuard or Tailscale

**WireGuard**:
- Open source
- High performance
- Full control
- Requires management

**Tailscale** (WireGuard-based):
- Managed WireGuard
- Easy to use
- Built-in NAT traversal
- Production SLA available
- Cost: ~$100-300/month

**Verdict**: ⚠️ Required for Phase 2 production

---

### 5. Observability at Scale - Thanos

#### Problem

Prometheus doesn't scale to multiple regions:
- Each cluster has separate Prometheus
- No global metrics view
- Difficult to correlate issues
- Limited retention (30 days)

#### Solution

Deploy Thanos for global metrics federation:

```
Architecture:

Studio-Primary (Query Frontend)
    ↓
    ├── Thanos Query
    │   ├── Thanos Sidecar (Studio-Sul Prometheus)
    │   ├── Thanos Sidecar (Studio-NE Prometheus)
    │   ├── Thanos Sidecar (Studio-CO Prometheus)
    │   └── Thanos Sidecar (Studio-N Prometheus)
    │
    └── Thanos Store Gateway
        └── Object Storage (MinIO/S3)
            └── Long-term metrics (90+ days)
```

#### Benefits

- **Global View**: Query metrics across all regions
- **Long-term Retention**: 90+ days (configurable)
- **Reduced Prometheus Load**: Offload to object storage
- **Cost Effective**: S3/MinIO cheaper than Prometheus storage

#### Cost

- Storage: ~$200-500/month (S3/MinIO for 90-day retention)
- Compute: Minimal (runs on existing nodes)

**Verdict**: ⚠️ Required for Phase 2

---

## Latency Analysis

### Phase 1 (Local)

| From | To | Expected Latency | Actual |
|------|-----|------------------|--------|
| Studio | Forge | <5ms | ✅ 2-3ms |
| Studio | Pro | <5ms | ✅ 1-2ms |
| Studio | Air | <5ms | ✅ 1-2ms |
| Studio | Pi | <10ms | ✅ 5-8ms |

**Verdict**: ✅ Excellent

---

### Phase 2 (Brazil)

| From | To | Expected Latency | Notes |
|------|-----|------------------|-------|
| São Paulo | Porto Alegre | 25-35ms | Acceptable |
| São Paulo | Recife | 60-80ms | Acceptable |
| São Paulo | Brasília | 50-70ms | Acceptable |
| São Paulo | Manaus | 80-100ms | High but workable |
| Porto Alegre | Recife | 100-130ms | Via hub: +hub latency |
| Porto Alegre | Brasília | 80-100ms | Via hub |

**Verdict**: ⚠️ Acceptable with hub-and-spoke

---

### Phase 3 (Global)

| From | To | Expected Latency | Notes |
|------|-----|------------------|-------|
| Brazil | US East | 120-150ms | Good for Americas |
| Brazil | US West | 180-220ms | Higher latency |
| Brazil | Europe | 200-250ms | High latency |
| Brazil | Asia-Pacific | 300-365ms | Very high latency |

**Verdict**: 🚧 Requires careful service placement

---

## Network Bandwidth Requirements

### Phase 1 (Local)

**Total Bandwidth**: <1Gbps

**Breakdown**:
- Linkerd control plane: <50Mbps
- Observability (metrics, logs): <200Mbps
- Application traffic: <500Mbps
- Backup traffic: <250Mbps (off-hours)

**Verdict**: ✅ Easily supported

---

### Phase 2 (Brazil)

**Total Bandwidth**: 2-5Gbps

**Breakdown**:
- Hub-to-hub traffic: 1-2Gbps
- Observability (Thanos): 500Mbps-1Gbps
- Application traffic: 500Mbps-1Gbps
- Backup replication: 500Mbps (off-hours)

**Cost**: ~$500-1000/month for bandwidth

**Verdict**: ⚠️ Requires monitoring and optimization

---

## Risk Assessment

| Risk | Impact | Likelihood | Phase | Mitigation |
|------|--------|------------|-------|------------|
| Latency >200ms | High | Low | 2 | Hub-and-spoke, optimize routing |
| Hub single point of failure | Critical | Medium | 2 | Deploy backup hubs |
| CIDR exhaustion | Medium | Low | 2 | Use /12 blocks, proper allocation |
| GPU shortage | High | Medium | 2-3 | Order early, consider cloud |
| Thanos complexity | Medium | High | 2 | Hire expertise, testing |
| WireGuard management | Medium | Medium | 2 | Use Tailscale managed service |
| Bandwidth costs | High | Medium | 2-3 | Monitor, optimize, compression |
| Security vulnerabilities | Critical | Low | All | Regular audits, Trivy scanning |

---

## Recommendations

### Immediate (Phase 1)

1. ✅ Continue with current setup (already operational)
2. ✅ Document network topology
3. ✅ Monitor latency and bandwidth

### Short-term (Phase 2 Planning)

1. ⚠️ Revise CIDR allocation to contiguous /12 blocks
2. ⚠️ Plan hub-and-spoke topology
3. ⚠️ Evaluate WireGuard vs Tailscale
4. ⚠️ Design Thanos architecture
5. ⚠️ Procure hardware for regional hubs

### Long-term (Phase 3)

1. 🚧 Defer until Phase 2 proves successful
2. 🚧 Evaluate global CDN integration
3. 🚧 Consider edge caching strategies
4. 🚧 Plan for multi-region failover

---

## Conclusion

The network architecture is sound for Phase 1 and feasible for Phase 2 with the recommended modifications. Phase 3 requires significant investment and should be deferred until Phase 2 demonstrates success.

**Confidence Levels**:
- Phase 1: 95% ✅
- Phase 2: 85% ⚠️ (with modifications)
- Phase 3: 70% 🚧 (high investment required)

---

## Related Documentation

- [DevOps Engineering Analysis](devops-engineering-analysis.md)
- [Production Readiness](production-readiness.md)
- [Phase 2 Preparation](../implementation/phase2-preparation.md)
- [Operational Maturity Roadmap](../implementation/operational-maturity-roadmap.md)

---

**Last Updated**: November 7, 2025  
**Analyzed by**: Senior Network Engineer (AI-assisted)  
**Maintained by**: SRE Team (Bruno Lucena)

