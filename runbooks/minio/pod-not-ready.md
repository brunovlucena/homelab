# 🚨 Runbook: MinIO Pod Not Ready

## Alert Information

**Alert Name:** `MinIOPodNotReady`  
**Severity:** High  
**Component:** minio  
**Service:** object-storage

## Symptom

MinIO pod exists but is not in Ready state. Readiness probe is failing.

## Impact

- **User Impact:** HIGH - Service unavailable or degraded
- **Business Impact:** HIGH - Applications cannot access object storage
- **Data Impact:** NONE - Data is safe, service is temporarily unavailable

## Diagnosis

### 1. Check Pod Status

```bash
kubectl get pods -n minio
kubectl describe pod -n minio -l app=minio
```

**Look for:**
- Pod phase (Pending, Running, Failed)
- Ready status (0/1, 1/1)
- Conditions (PodScheduled, Initialized, Ready, ContainersReady)
- Events (recent errors or warnings)

### 2. Check Readiness Probe Configuration

```bash
kubectl get deployment -n minio minio -o yaml | grep -A 10 readinessProbe
```

### 3. Test Readiness Endpoint

```bash
# Port forward to pod
POD_NAME=$(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward -n minio $POD_NAME 9000:9000

# Test readiness endpoint
curl -v http://localhost:9000/minio/health/ready
```

### 4. Check Pod Logs

```bash
kubectl logs -n minio -l app=minio --tail=100
kubectl logs -n minio -l app=minio --previous  # If pod has restarted
```

### 5. Check PersistentVolume Status

```bash
kubectl get pvc -n minio
kubectl describe pvc -n minio
```

### 6. Check Resource Constraints

```bash
kubectl describe node $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].spec.nodeName}')
kubectl top nodes
kubectl top pods -n minio
```

## Resolution

### Scenario A: Readiness Probe Timing Issues

**Likely Cause:** Pod needs more time to initialize

**Steps:**
1. Check current probe settings:
   ```bash
   kubectl get deployment -n minio minio -o yaml | grep -A 15 readinessProbe
   ```

2. Increase initialDelaySeconds and timeouts:
   ```yaml
   # Edit flux/clusters/homelab/infrastructure/minio/k8s/helmrelease.yaml
   readinessProbe:
     httpGet:
       path: /minio/health/ready
       port: 9000
       scheme: HTTP
     initialDelaySeconds: 60  # Increase from 15
     periodSeconds: 20        # Increase from 15
     timeoutSeconds: 10       # Increase from 5
     successThreshold: 1
     failureThreshold: 5      # Increase from 3
   ```

3. Apply changes:
   ```bash
   flux reconcile helmrelease -n minio minio
   ```

### Scenario B: PersistentVolume Not Bound

**Likely Cause:** PVC waiting for PV or storage provisioner issue

**Steps:**
1. Check PVC status:
   ```bash
   kubectl get pvc -n minio
   ```

2. If status is "Pending":
   ```bash
   kubectl describe pvc -n minio <pvc-name>
   # Look for errors in Events section
   ```

3. Check if storage class exists:
   ```bash
   kubectl get storageclass
   ```

4. Check if PV exists:
   ```bash
   kubectl get pv | grep minio
   ```

5. If PV is "Released", clear claimRef:
   ```bash
   kubectl patch pv <pv-name> -p '{"spec":{"claimRef": null}}'
   ```

6. If storage provisioner issue, check provisioner logs:
   ```bash
   kubectl logs -n kube-system -l app=<storage-provisioner>
   ```

### Scenario C: Disk Not Available

**Likely Cause:** Underlying storage is offline or corrupted

**Steps:**
1. Exec into pod:
   ```bash
   kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- sh
   ```

2. Check if /data is mounted:
   ```bash
   df -h /data
   ls -la /data
   ```

3. Check for disk errors:
   ```bash
   dmesg | grep -i error
   cat /proc/mounts | grep data
   ```

4. If mount issue, check node:
   ```bash
   NODE=$(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].spec.nodeName}')
   kubectl describe node $NODE
   # SSH to node and check storage
   ```

### Scenario D: Insufficient Resources

**Likely Cause:** Node has insufficient CPU/memory to run pod

**Steps:**
1. Check pod resource requests:
   ```bash
   kubectl describe pod -n minio -l app=minio | grep -A 10 "Requests:"
   ```

2. Check node capacity:
   ```bash
   kubectl describe node $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].spec.nodeName}') | grep -A 10 "Allocated resources:"
   ```

3. If node is full, either:
   - Add more nodes
   - Reduce resource requests
   - Move other workloads

4. Reduce resource requests if possible:
   ```yaml
   resources:
     limits:
       memory: 4Gi
       cpu: 2000m
     requests:
       memory: 2Gi    # Reduce if too high
       cpu: 500m      # Reduce if too high
   ```

### Scenario E: Configuration Error

**Likely Cause:** Invalid MinIO configuration preventing startup

**Steps:**
1. Check logs for config errors:
   ```bash
   kubectl logs -n minio -l app=minio --tail=200 | grep -i "config\|error\|fatal"
   ```

2. Check environment variables:
   ```bash
   kubectl get deployment -n minio minio -o yaml | grep -A 20 env:
   ```

3. Common config issues:
   - Invalid MINIO_ROOT_USER or MINIO_ROOT_PASSWORD
   - Incorrect domain/endpoint configuration
   - Bad MINIO_CACHE or storage settings

4. Verify secrets:
   ```bash
   kubectl get secret -n minio minio -o yaml
   ```

5. Fix configuration and restart:
   ```bash
   kubectl rollout restart deployment/minio -n minio
   ```

### Scenario F: Port Already in Use

**Likely Cause:** Another process using port 9000

**Steps:**
1. Check if port is in use:
   ```bash
   kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- sh
   netstat -tlnp | grep 9000
   ```

2. Check for multiple MinIO processes:
   ```bash
   ps aux | grep minio
   ```

3. If orphaned process, kill it:
   ```bash
   pkill -9 minio
   ```

4. Restart pod:
   ```bash
   kubectl delete pod -n minio -l app=minio
   ```

### Scenario G: Network Policy Blocking Health Checks

**Likely Cause:** Network policy preventing kubelet from reaching pod

**Steps:**
1. Check network policies:
   ```bash
   kubectl get networkpolicies -n minio
   kubectl describe networkpolicies -n minio
   ```

2. Temporarily disable to test:
   ```bash
   kubectl delete networkpolicy -n minio <policy-name>
   ```

3. If that fixes it, update network policy to allow kubelet:
   ```yaml
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: minio-network-policy
     namespace: minio
   spec:
     podSelector:
       matchLabels:
         app: minio
     ingress:
     - from:
       - namespaceSelector: {}  # Allow from all namespaces for health checks
       ports:
       - protocol: TCP
         port: 9000
   ```

### Scenario H: DNS Resolution Issues

**Likely Cause:** Pod cannot resolve required DNS names

**Steps:**
1. Test DNS from pod:
   ```bash
   kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- sh
   nslookup kubernetes.default
   nslookup minio.minio.svc.cluster.local
   ```

2. Check CoreDNS is healthy:
   ```bash
   kubectl get pods -n kube-system -l k8s-app=kube-dns
   kubectl logs -n kube-system -l k8s-app=kube-dns --tail=50
   ```

3. Restart CoreDNS if needed:
   ```bash
   kubectl rollout restart deployment/coredns -n kube-system
   ```

## Verification

### 1. Check Pod is Ready

```bash
kubectl get pods -n minio
# Should show 1/1 READY
```

### 2. Test Readiness Endpoint

```bash
kubectl port-forward -n minio svc/minio 9000:9000
curl http://localhost:9000/minio/health/ready
# Should return 200 OK
```

### 3. Test Liveness Endpoint

```bash
curl http://localhost:9000/minio/health/live
# Should return 200 OK
```

### 4. Test Service Endpoint

```bash
mc alias set local http://localhost:9000 <access-key> <secret-key>
mc admin info local
```

### 5. No Recent Events

```bash
kubectl get events -n minio --sort-by='.lastTimestamp' | head -20
# Should not show errors
```

## Prevention

1. **Configure appropriate probe settings:**
   ```yaml
   livenessProbe:
     httpGet:
       path: /minio/health/live
       port: 9000
     initialDelaySeconds: 30
     periodSeconds: 30
     timeoutSeconds: 10
     failureThreshold: 3
   
   readinessProbe:
     httpGet:
       path: /minio/health/ready
       port: 9000
     initialDelaySeconds: 15
     periodSeconds: 15
     timeoutSeconds: 5
     failureThreshold: 3
   ```

2. **Set up monitoring:**
   ```yaml
   - alert: MinIOPodNotReady
     expr: |
       kube_pod_status_ready{namespace="minio",condition="true"} == 0
     for: 5m
     annotations:
       summary: "MinIO pod not ready"
   
   - alert: MinIOReadinessProbeFailure
     expr: |
       rate(prober_probe_total{probe_type="readiness",namespace="minio"}[5m]) 
       - rate(prober_probe_success_total{probe_type="readiness",namespace="minio"}[5m]) 
       > 0
     for: 5m
     annotations:
       summary: "MinIO readiness probe failing"
   ```

3. **Use PodDisruptionBudget:**
   ```yaml
   apiVersion: policy/v1
   kind: PodDisruptionBudget
   metadata:
     name: minio-pdb
     namespace: minio
   spec:
     minAvailable: 1
     selector:
       matchLabels:
         app: minio
   ```

4. **Ensure adequate resources:**
   - Size CPU/memory appropriately
   - Monitor resource usage trends
   - Set requests to actual usage, not arbitrary values

5. **Pre-pull images:**
   ```yaml
   apiVersion: apps/v1
   kind: DaemonSet
   metadata:
     name: image-prepuller
   spec:
     selector:
       matchLabels:
         app: image-prepuller
     template:
       spec:
         initContainers:
         - name: prepull-minio
           image: minio/minio:latest
           command: ["sh", "-c", "echo Image pulled"]
   ```

6. **Storage monitoring:**
   - Monitor PV/PVC status
   - Alert on storage provisioner issues
   - Test disaster recovery procedures

7. **Regular health checks:**
   - Automated testing of health endpoints
   - Monitor probe success rates
   - Alert on probe failures

## Metrics to Monitor

```promql
# Pod ready status
kube_pod_status_ready{namespace="minio",condition="true"}

# Readiness probe failures
rate(prober_probe_total{probe_type="readiness",namespace="minio"}[5m]) 
- rate(prober_probe_success_total{probe_type="readiness",namespace="minio"}[5m])

# Pod restarts
rate(kube_pod_container_status_restarts_total{namespace="minio"}[15m])

# Container status
kube_pod_container_status_ready{namespace="minio"}
```

## Related Alerts

- `MinIODown`
- `MinIOPodCrashLoop`
- `MinIOPVCPending`
- `MinIOHighMemoryUsage`
- `MinIOStorageSpaceLow`

## Escalation

**When to escalate:**
- Pod not ready >15 minutes with no clear cause
- Storage infrastructure issues
- Cluster-wide problems affecting multiple services
- Need infrastructure changes (more nodes, storage, etc.)

**Escalation Path:**
1. Senior SRE Team
2. Storage Infrastructure Team
3. Kubernetes Platform Team
4. Vendor Support (for storage or MinIO issues)

## Additional Resources

- [Kubernetes Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [MinIO Health Endpoints](https://min.io/docs/minio/linux/operations/monitoring/healthcheck-probe.html)
- [Troubleshooting Pods](https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/)
- Internal Wiki: Pod Troubleshooting Guide
- Slack: #sre-kubernetes

