# Migration Guide

> **Part of**: [Homelab Documentation](../README.md) → Operations  
> **Last Updated**: November 7, 2025

---

## Overview

Step-by-step procedures for migrating workloads between clusters in the homelab environment.

## Migration Scenarios

### 1. Air → Pro

**When**: After successful experiments in Air cluster

**Use Case**: Promoting validated configurations to development

**Considerations**:
- More nodes available in Pro (7 vs 4)
- More specialized node roles
- Additional services (databases, monitoring)

### 2. Pro → Studio

**When**: Promoting to production

**Use Case**: Production-ready applications and services

**Considerations**:
- High availability (12 nodes, multi-zone)
- Production-grade monitoring
- Manual approval required
- Backup procedures mandatory

### 3. Studio → Regional

**When**: Phase 2 Brazil expansion (future)

**Use Case**: Geographic distribution for latency reduction

**Considerations**:
- LGPD compliance (Brazil data residency)
- Regional Forge hubs for GPU workloads
- Thanos for federated metrics

### 4. Forge Model Updates

**When**: Deploying new models or updating existing ones

**Use Case**: LLM updates, new model versions

**Considerations**:
- GPU resource allocation
- Model download size and time
- Rolling updates to avoid downtime

## General Migration Steps

### Step 1: Export Configuration

```bash
# Export deployment from source cluster
kubectl --context=air get deployment myapp -o yaml > myapp.yaml

# Export service
kubectl --context=air get service myapp -o yaml > myapp-svc.yaml

# Export ConfigMap (if exists)
kubectl --context=air get configmap myapp-config -o yaml > myapp-config.yaml
```

### Step 2: Update Cluster-Specific Values

Edit the exported YAML files:

```yaml
# Update namespaces if needed
metadata:
  namespace: production  # was: development

# Update node selectors
spec:
  template:
    spec:
      nodeSelector:
        role: ai-agents     # Target Pro/Studio specific nodes
        criticality: high   # Raise criticality for production

# Update resource limits
      resources:
        requests:
          cpu: "500m"       # Increase for production
          memory: "1Gi"
        limits:
          cpu: "2000m"
          memory: "4Gi"

# Update service endpoints for cross-cluster
# Example: Database connection string
env:
- name: DATABASE_URL
  value: "postgresql://postgres.data.svc.studio.remote:5432/myapp"
```

### Step 3: Apply to Target Cluster

```bash
# Apply to target cluster
kubectl --context=pro apply -f myapp-config.yaml
kubectl --context=pro apply -f myapp-svc.yaml
kubectl --context=pro apply -f myapp.yaml
```

### Step 4: Verify Deployment

```bash
# Check pods are running
kubectl --context=pro get pods -l app=myapp

# Check logs
kubectl --context=pro logs deployment/myapp --tail=50

# Check service endpoints
kubectl --context=pro get endpoints myapp

# Verify resource allocation
kubectl --context=pro top pods -l app=myapp
```

### Step 5: Test Functionality

```bash
# Test internal connectivity
kubectl --context=pro exec -it deployment/myapp -- curl http://localhost:8080/health

# Test cross-cluster connectivity (if applicable)
kubectl --context=pro exec -it deployment/myapp -- curl http://service.namespace.svc.forge.remote:8080/health

# Run smoke tests
make test-smoke CLUSTER=pro APP=myapp
```

### Step 6: Clean Up Source

⚠️ **CRITICAL**: Only after validation is complete!

```bash
# Verify target is healthy first
kubectl --context=pro get pods -l app=myapp
kubectl --context=pro logs deployment/myapp --tail=20

# Then delete from source
# NOTE: This requires user confirmation per memory rules
kubectl --context=air delete deployment myapp
kubectl --context=air delete service myapp
kubectl --context=air delete configmap myapp-config
```

## Cross-Cluster Service Migration

For services that need to remain accessible across clusters during migration:

### Step 1: Export Service

```yaml
# Enable cross-cluster export on source
apiVersion: v1
kind: Service
metadata:
  name: myapp
  namespace: default
  annotations:
    multicluster.linkerd.io/export: "true"  # Enable cross-cluster access
spec:
  ports:
  - port: 8080
    targetPort: 8080
  selector:
    app: myapp
```

### Step 2: Update Client Configurations

```yaml
# Clients should use multi-cluster endpoint
# From Studio cluster accessing service in Pro
env:
- name: SERVICE_URL
  value: "http://myapp.default.svc.pro.remote:8080"
```

### Step 3: Deploy to Target

Deploy the service to target cluster with export enabled.

### Step 4: Verify Cross-Cluster

```bash
# From another cluster, test connectivity
kubectl --context=studio exec -it test-pod -- curl http://myapp.default.svc.pro.remote:8080/health
```

### Step 5: Decomission Source

Once all clients are using the new location, remove the export annotation from the source cluster and delete the deployment.

## Migration Checklist

Before starting any migration:

- [ ] Backup current configuration (`kubectl get ... -o yaml`)
- [ ] Document dependencies (databases, APIs, services)
- [ ] Review resource requirements
- [ ] Update node selectors/affinity
- [ ] Update cross-cluster service references
- [ ] Test in target cluster
- [ ] Update monitoring dashboards
- [ ] Update documentation
- [ ] Clean up source cluster (after validation)

## Common Issues

### Issue: Pods Stuck in Pending

**Cause**: Node selectors don't match target cluster nodes

**Solution**: Update node selectors to match target cluster labels

```yaml
spec:
  template:
    spec:
      nodeSelector:
        role: ai-agents  # Ensure this label exists on target nodes
```

### Issue: Service Not Reachable

**Cause**: Service not exported for cross-cluster access

**Solution**: Add export annotation

```yaml
metadata:
  annotations:
    multicluster.linkerd.io/export: "true"
```

### Issue: Image Pull Errors

**Cause**: Image not available in target cluster registry

**Solution**: Ensure image is pushed to shared registry or pull from Docker Hub

## Related Documentation

- [Node Labels Reference](node-labels-reference.md)
- [Pod Affinity Examples](pod-affinity-examples.md)
- [Cluster Documentation](../clusters/README.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

