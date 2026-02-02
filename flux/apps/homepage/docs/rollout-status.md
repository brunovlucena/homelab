# Homepage Rollout Status Check

## Current Status: **NOT DEPLOYED**

**Date:** 2025-12-21 15:36

### Deployment Status
- ❌ **Namespace `homepage` does not exist**
- ❌ **No deployments found** (`homepage-frontend`, `homepage-api`)
- ❌ **No pods running** with label `app.kubernetes.io/name=homepage`

### Version Information
- **Current Version in VERSION file:** `0.1.35`
- **Latest Commit:** `267b1672` - "fix: Use public directory for three-scales-framework image"
- **Previous Commit:** `11faff45` - "chore(release): homepage v0.1.35"

### Image Status
- ✅ **Image path fixed** in markdown: `/blog-posts/graphs/three-scales-framework.png`
- ✅ **Image file exists** in public directory: `blog-posts/graphs/three-scales-framework.png` (438KB)
- ⚠️ **Browser still showing old path** (`/storage/homepage-blog/...`) - indicates old version is cached/served

## Prometheus Metrics Queries (Once Deployed)

When the homepage is deployed, use these Prometheus queries to monitor rollout:

### Deployment Replica Status
```promql
# Frontend deployment status
kube_deployment_status_replicas{namespace="homepage", deployment="homepage-frontend"}
kube_deployment_status_replicas_ready{namespace="homepage", deployment="homepage-frontend"}
kube_deployment_status_replicas_available{namespace="homepage", deployment="homepage-frontend"}
kube_deployment_status_replicas_updated{namespace="homepage", deployment="homepage-frontend"}

# API deployment status
kube_deployment_status_replicas{namespace="homepage", deployment="homepage-api"}
kube_deployment_status_replicas_ready{namespace="homepage", deployment="homepage-api"}
```

### Deployment Conditions
```promql
# Check deployment conditions (Available, Progressing, ReplicaFailure)
kube_deployment_status_condition{namespace="homepage", deployment=~"homepage-.*"}
```

### Pod Status
```promql
# Pod phase status
kube_pod_status_phase{namespace="homepage", pod=~"homepage-.*"}

# Container ready status
kube_pod_container_status_ready{namespace="homepage", pod=~"homepage-.*"}

# Container restart count
kube_pod_container_status_restarts_total{namespace="homepage", pod=~"homepage-.*"}
```

### Rollout Progress
```promql
# Rollout completion percentage
(
  kube_deployment_status_replicas_updated{namespace="homepage", deployment="homepage-frontend"}
  /
  kube_deployment_spec_replicas{namespace="homepage", deployment="homepage-frontend"}
) * 100
```

## Next Steps to Deploy

1. **Apply the kustomization** (creates namespace and deployments):
   ```bash
   cd flux/infrastructure/homepage
   kubectl apply -k k8s/kustomize/pro
   ```

2. **Or use Makefile**:
   ```bash
   cd flux/infrastructure/homepage
   make build-images-local
   make rollout
   ```

3. **Monitor rollout**:
   ```bash
   kubectl rollout status deployment/homepage-frontend -n homepage
   kubectl rollout status deployment/homepage-api -n homepage
   ```

4. **After deployment, query Prometheus** using the queries above

## Expected Configuration

- **Namespace:** `homepage`
- **Frontend Deployment:** `homepage-frontend`
  - Replicas: 2
  - Image: `localhost:5001/homepage-frontend:v0.1.35`
  - Strategy: RollingUpdate (maxSurge: 1, maxUnavailable: 0)
- **API Deployment:** `homepage-api`
  - Image: `localhost:5001/homepage-api:v0.1.35`

