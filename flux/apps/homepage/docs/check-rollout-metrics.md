# Homepage Rollout Status Check via Prometheus Metrics

## Current Status

**Namespace:** `homepage` - **NOT FOUND**
- The homepage namespace does not exist in the cluster
- Deployments cannot be queried until the namespace and resources are created

## Prometheus Queries for Rollout Monitoring

Once the homepage is deployed, use these Prometheus queries to monitor the rollout:

### 1. Deployment Replica Status

```promql
# Total desired replicas
kube_deployment_spec_replicas{namespace="homepage", deployment=~"homepage-.*"}

# Current replicas
kube_deployment_status_replicas{namespace="homepage", deployment=~"homepage-.*"}

# Ready replicas
kube_deployment_status_replicas_ready{namespace="homepage", deployment=~"homepage-.*"}

# Available replicas
kube_deployment_status_replicas_available{namespace="homepage", deployment=~"homepage-.*"}

# Updated replicas (new version)
kube_deployment_status_replicas_updated{namespace="homepage", deployment=~"homepage-.*"}

# Unavailable replicas
kube_deployment_status_replicas_unavailable{namespace="homepage", deployment=~"homepage-.*"}
```

### 2. Deployment Conditions

```promql
# Deployment conditions (Available, Progressing, ReplicaFailure)
kube_deployment_status_condition{namespace="homepage", deployment=~"homepage-.*"}
```

### 3. Pod Status

```promql
# Pod phase (Pending, Running, Succeeded, Failed, Unknown)
kube_pod_status_phase{namespace="homepage", pod=~"homepage-.*"}

# Container ready status
kube_pod_container_status_ready{namespace="homepage", pod=~"homepage-.*"}

# Container restart count
kube_pod_container_status_restarts_total{namespace="homepage", pod=~"homepage-.*"}
```

### 4. Rollout Progress

```promql
# Rollout percentage (updated / desired)
(
  kube_deployment_status_replicas_updated{namespace="homepage", deployment=~"homepage-.*"}
  /
  kube_deployment_spec_replicas{namespace="homepage", deployment=~"homepage-.*"}
) * 100

# Replica availability percentage
(
  kube_deployment_status_replicas_available{namespace="homepage", deployment=~"homepage-.*"}
  /
  kube_deployment_spec_replicas{namespace="homepage", deployment=~"homepage-.*"}
) * 100
```

### 5. Resource Usage During Rollout

```promql
# CPU usage per pod
rate(container_cpu_usage_seconds_total{namespace="homepage", pod=~"homepage-.*"}[5m])

# Memory usage per pod
container_memory_working_set_bytes{namespace="homepage", pod=~"homepage-.*"}

# Network traffic
rate(container_network_receive_bytes_total{namespace="homepage", pod=~"homepage-.*"}[5m])
rate(container_network_transmit_bytes_total{namespace="homepage", pod=~"homepage-.*"}[5m])
```

## Expected Deployment Configuration

Based on the deployment YAML:
- **Frontend Deployment:** `homepage-frontend`
- **API Deployment:** `homepage-api`
- **Namespace:** `homepage`
- **Replicas:** 2 (frontend)
- **Strategy:** RollingUpdate
  - MaxSurge: 1
  - MaxUnavailable: 0

## Next Steps

1. **Create namespace** (if using kustomize, it should create it):
   ```bash
   kubectl apply -k flux/infrastructure/homepage/k8s/kustomize/base
   ```

2. **Apply deployments**:
   ```bash
   kubectl apply -k flux/infrastructure/homepage/k8s/kustomize/pro
   ```

3. **Monitor rollout**:
   ```bash
   kubectl rollout status deployment/homepage-frontend -n homepage
   kubectl rollout status deployment/homepage-api -n homepage
   ```

4. **Query Prometheus** using the queries above once deployed

