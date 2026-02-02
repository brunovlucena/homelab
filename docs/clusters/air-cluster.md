# Air Cluster

> **Part of**: [Homelab Documentation](../README.md) → Clusters  
> **Last Updated**: November 7, 2025

---

## Overview

**Platform**: Kind  
**Hardware**: Mac Studio (M2 Ultra, 192GB RAM)  
**Architecture**: ARM64  
**Purpose**: Quick experiments, CI/CD, rapid iteration  
**Nodes**: 4 (1 control-plane + 3 workers)  
**Network**: 10.246.0.0/16

---

## Node Architecture

| Node | Role | Workloads |
|------|------|-----------|
| 1 | control-plane | Kubernetes control plane |
| 2 | worker | Experimental workloads |
| 3 | worker | CI/CD testing |
| 4 | worker | Feature validation |

---

## Key Features

- **Minimal Resource Usage**: Lightweight configuration for quick iteration
- **Fast Spin-up/Tear-down**: Can be recreated in minutes
- **CI/CD Testing**: Automated pipeline validation
- **Experimental Workloads**: Safe environment for breaking changes

---

## Use Cases

### 1. Testing New Helm Charts

Before deploying to Pro or Studio, validate Helm charts in Air:

```bash
# Deploy to Air for testing
kubectl --context=air apply -f new-chart.yaml

# Verify
kubectl --context=air get pods -n test-namespace

# If good, promote to Pro
kubectl --context=pro apply -f new-chart.yaml
```

### 2. Flux Configuration Validation

Test Flux configurations without affecting production:

```bash
# Test Flux reconciliation
flux --context=air reconcile kustomization flux-system

# Check for errors
flux --context=air get all
```

### 3. Breaking Changes Testing

Test potentially breaking changes safely:

```bash
# Test new API version
kubectl --context=air apply -f new-api-version.yaml

# Test backwards compatibility
make test-air
```

### 4. Quick Proof-of-Concepts

Rapidly iterate on new ideas:

```bash
# Deploy POC
kubectl --context=air create deployment poc --image=nginx

# Test
curl http://poc.air.cluster

# Delete if not needed
kubectl --context=air delete deployment poc
```

---

## Resource Limits

### Per Node

- **CPU**: 4 cores per worker node
- **Memory**: 16GB per worker node
- **Disk**: 50GB per node

### Cluster Total

- **CPU**: 12 cores (3 workers × 4 cores)
- **Memory**: 48GB
- **Disk**: 150GB

---

## Service Mesh Integration

Air cluster is connected to the multi-cluster service mesh via Linkerd:

```yaml
# Access services in other clusters
apiVersion: v1
kind: Service
metadata:
  name: remote-service
spec:
  type: ExternalName
  externalName: service.namespace.svc.pro.remote
```

---

## CI/CD Integration

### GitHub Actions

Air is used for automated CI/CD testing:

```yaml
# .github/workflows/deploy-air.yml
name: Deploy to Air
on:
  push:
    branches: ['feature/*']
  pull_request:

jobs:
  deploy:
    runs-on: [self-hosted]
    steps:
      - name: Deploy to Air
        run: |
          kubectl --context=air apply -f manifests/
      
      - name: Run smoke tests
        run: |
          make test-smoke CLUSTER=air
```

---

## Best Practices

### DO

- ✅ Use Air for quick experiments
- ✅ Test breaking changes here first
- ✅ Validate Flux configurations
- ✅ Run automated CI/CD tests
- ✅ Tear down and recreate frequently

### DON'T

- ❌ Don't store persistent data
- ❌ Don't run long-running experiments
- ❌ Don't use for production workloads
- ❌ Don't depend on uptime

---

## Common Operations

### Create Air Cluster

```bash
# Via Makefile
make up ENV=air
```

### Delete Air Cluster

```bash
# Via Makefile
make down ENV=air
```

### Recreate Air Cluster

```bash
# Full recreate
make down ENV=air && make up ENV=air
```

### Deploy Workload

```bash
# Deploy
kubectl --context=air apply -f workload.yaml

# Verify
kubectl --context=air get pods -A
```

### Check Logs

```bash
# Pod logs
kubectl --context=air logs pod-name -n namespace

# Follow logs
kubectl --context=air logs -f pod-name -n namespace
```

---

## Migration Path

When an experiment in Air is successful:

1. **Validate in Air** - Test thoroughly
2. **Promote to Pro** - Deploy to development cluster
3. **Test in Pro** - Integration and E2E tests
4. **Promote to Studio** - Production deployment (manual approval)

See [Migration Guide](../operations/migration-guide.md) for detailed procedures.

---

## Troubleshooting

### Cluster Won't Start

```bash
# Check Docker
docker ps

# Check Kind
kind get clusters

# Recreate
make down ENV=air && make up ENV=air
```

### Out of Resources

```bash
# Check resource usage
kubectl --context=air top nodes
kubectl --context=air top pods -A

# Delete old workloads
kubectl --context=air delete namespace old-namespace
```

### Network Issues

```bash
# Check Linkerd
linkerd --context=air check

# Restart gateway
kubectl --context=air rollout restart deployment/linkerd-gateway -n linkerd-multicluster
```

---

## Related Documentation

- [Pro Cluster](pro-cluster.md)
- [Studio Cluster](studio-cluster.md)
- [Migration Guide](../operations/migration-guide.md)
- [CI/CD Pipeline](../implementation/operational-maturity-roadmap.md#cicd-foundation)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

