# Pro Cluster

> **Part of**: [Homelab Documentation](../README.md) → Clusters  
> **Last Updated**: November 7, 2025

---

## Overview

**Platform**: Kind  
**Hardware**: Mac Studio (M2 Ultra, 192GB RAM)  
**Architecture**: ARM64  
**Purpose**: Development and testing with full feature set  
**Nodes**: 7 (1 control-plane + 6 workers)  
**Network**: 10.244.0.0/16

---

## Node Architecture

| Node | Role | Workloads |
|------|------|-----------|
| 1 | control-plane | Kubernetes control plane |
| 2 | platform | Flux, cert-manager, Linkerd, Flagger |
| 3 | ai-agents | agent-auditor, agent-bruno, aigoat |
| 4 | serverless | Knative Serving/Eventing |
| 5 | observability | Prometheus, Grafana, Loki, Tempo |
| 6 | data | PostgreSQL, MongoDB, Redis, MinIO |
| 7 | ingress | Cloudflare tunnel, load balancers |

---

## Key Features

- **Full-stack Development**: Complete platform services
- **Integration Testing**: All services available
- **Performance Testing**: k6 load testing
- **Pre-production Validation**: Final testing before Studio

---

## Use Cases

### 1. Full-stack Application Development

Develop applications with all dependencies:

```bash
# Deploy full stack
kubectl --context=pro apply -f app/
kubectl --context=pro apply -f database/
kubectl --context=pro apply -f cache/

# Test end-to-end
make test-e2e CLUSTER=pro
```

### 2. Integration Testing

Test service interactions:

```bash
# Deploy services
kubectl --context=pro apply -f services/

# Run integration tests
make test-integration CLUSTER=pro
```

### 3. Performance Testing

Run load tests with k6:

```bash
# Deploy application
kubectl --context=pro apply -f app/

# Run load test
k6 run --vus 100 --duration 5m load-test.js
```

### 4. Pre-production Validation

Final validation before Studio:

```bash
# Deploy to Pro
kubectl --context=pro apply -f manifests/

# Run all tests
make test-all CLUSTER=pro

# If passing, promote to Studio
kubectl --context=studio apply -f manifests/
```

---

## Resource Limits

### Per Node

- **CPU**: 8 cores per specialized node
- **Memory**: 32GB per node
- **Disk**: 100GB per node

### Cluster Total

- **CPU**: 48 cores
- **Memory**: 192GB
- **Disk**: 700GB

---

## Service Mesh Integration

Pro cluster exports services to other clusters:

```yaml
# Export service for cross-cluster access
apiVersion: v1
kind: Service
metadata:
  name: api-service
  annotations:
    multicluster.linkerd.io/export: "true"
spec:
  ports:
  - port: 8080
  selector:
    app: api-service
```

---

## Related Documentation

- [Air Cluster](air-cluster.md)
- [Studio Cluster](studio-cluster.md)
- [Migration Guide](../operations/migration-guide.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

