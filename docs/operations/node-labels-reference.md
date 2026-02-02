# Node Labels Reference

> **Part of**: [Homelab Documentation](../README.md) → Operations  
> **Last Updated**: November 7, 2025

---

## Overview

Comprehensive reference of all node labels used across clusters for workload placement, resource management, and operational purposes.

## Standard Kubernetes Labels

```yaml
kubernetes.io/hostname: <node-name>
kubernetes.io/arch: arm64 | amd64
kubernetes.io/os: linux
topology.kubernetes.io/zone: us-east-1a | us-east-1b | us-east-1c
```

## Custom Homelab Labels

### Role Labels

```yaml
role: control-plane | platform | ai-agents | observability | data | serverless | ingress | edge-worker | training | inference | ml-platform
```

**Usage**: Primary function of the node

**Examples**:
- `role=ai-agents`: Runs agent-bruno, agent-auditor, etc.
- `role=training`: GPU nodes for model training
- `role=inference`: GPU nodes for model serving (VLLM, Ollama)

### Tier Labels

```yaml
tier: control-plane | infrastructure | platform | application | data
```

**Usage**: Hierarchical layer in the architecture

**Examples**:
- `tier=infrastructure`: Core services (Flux, cert-manager)
- `tier=application`: User-facing services

### Resource Profile Labels

```yaml
resource-profile: minimal | low | medium | high | gpu-intensive
```

**Usage**: Resource availability indication

**Examples**:
- `resource-profile=minimal`: Pi nodes (2-4GB RAM)
- `resource-profile=gpu-intensive`: Forge GPU nodes

### Criticality Labels

```yaml
criticality: critical | high | medium | low
```

**Usage**: Service importance for SLA/SLO targeting

**Examples**:
- `criticality=critical`: Production agents, databases
- `criticality=low`: Experimental workloads

### GPU Labels (Forge Cluster)

```yaml
gpu-type: nvidia
nvidia.com/gpu: "true"
nvidia.com/gpu.product: NVIDIA-A100-SXM4-40GB
```

**Usage**: GPU resource identification

## Label Combinations

### Studio Cluster Example

```yaml
# High-priority AI agent node
role: ai-agents
tier: application
resource-profile: high
criticality: critical
topology.kubernetes.io/zone: us-east-1a
```

### Forge Cluster Example

```yaml
# GPU training node
role: training
tier: application
resource-profile: gpu-intensive
gpu-type: nvidia
nvidia.com/gpu: "true"
nvidia.com/gpu.product: NVIDIA-A100-SXM4-40GB
```

### Pi Cluster Example

```yaml
# Edge worker node
role: edge-worker
tier: application
resource-profile: minimal
criticality: low
```

## Best Practices

1. **Always use role + tier** for every node
2. **Add resource-profile** for scheduling optimization
3. **Set criticality** for monitoring/alerting thresholds
4. **Use zone labels** for HA deployments (Studio)
5. **Apply GPU labels** automatically via device plugins

## Related Documentation

- [Pod Affinity Examples](pod-affinity-examples.md)
- [Migration Guide](migration-guide.md)
- [Cluster Documentation](../clusters/README.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

