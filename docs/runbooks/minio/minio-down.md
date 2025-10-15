# 🚨 Runbook: MinIO Service Down

## Alert Information

**Alert Name:** `MinIODown`  
**Severity:** Critical  
**Component:** minio  
**Service:** object-storage

## Symptom

MinIO service is completely unavailable. API requests are failing and no buckets are accessible.

## Impact

- **User Impact:** CRITICAL - All object storage operations fail
- **Business Impact:** HIGH - Applications dependent on object storage cannot function
- **Data Impact:** NONE (if temporary) - Data preserved, but inaccessible until service restored

## Diagnosis

### 1. Check MinIO Pod Status

```bash
kubectl get pods -n minio
kubectl describe pod -n minio -l app=minio
```

**Look for:**
- Pod status (Running, CrashLoopBackOff, Pending)
- Recent events (OOMKilled, ImagePullBackOff, etc.)
- Restart count

### 2. Check MinIO Logs

```bash
kubectl logs -n minio -l app=minio --tail=100
```

**Common error patterns:**
- "unable to write to backend disks"
- "disk not found"
- "permission denied"
- "exceeded maximum allowed requests"

### 3. Check MinIO Service and Endpoints

```bash
kubectl get svc -n minio
kubectl get endpoints -n minio
```

### 4. Check PersistentVolumeClaims

```bash
kubectl get pvc -n minio
kubectl describe pvc -n minio
```

### 5. Test MinIO Connectivity

```bash
# Port forward to MinIO
kubectl port-forward -n minio svc/minio 9000:9000

# In another terminal, test the endpoint
curl -I http://localhost:9000/minio/health/live
```

### 6. Check Resource Usage

```bash
kubectl top pods -n minio
kubectl describe node $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].spec.nodeName}')
```

## Resolution

### Scenario A: Pod CrashLoopBackOff

**Likely Cause:** Application error, corrupted data, or configuration issue

**Steps:**
1. Check logs for root cause:
   ```bash
   kubectl logs -n minio -l app=minio --previous
   ```

2. If disk corruption suspected:
   ```bash
   # Access pod shell
   kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- sh
   
   # Check disk health
   df -h
   ls -la /data
   ```

3. Restart the pod:
   ```bash
   kubectl rollout restart deployment/minio -n minio
   ```

### Scenario B: PersistentVolume Issues

**Likely Cause:** Storage backend unavailable, PV not bound

**Steps:**
1. Check PV status:
   ```bash
   kubectl get pv | grep minio
   kubectl describe pv <pv-name>
   ```

2. If PV is in "Released" state:
   ```bash
   kubectl patch pv <pv-name> -p '{"spec":{"claimRef": null}}'
   ```

3. Delete and recreate PVC if necessary:
   ```bash
   kubectl delete pvc -n minio <pvc-name>
   # Wait for Flux to recreate or manually apply
   ```

### Scenario C: Out of Memory (OOMKilled)

**Likely Cause:** Insufficient memory allocation

**Steps:**
1. Check memory limits in HelmRelease:
   ```bash
   kubectl get helmrelease -n minio minio -o yaml | grep -A 10 resources
   ```

2. Increase memory limits:
   ```yaml
   # Edit flux/clusters/homelab/infrastructure/minio/k8s/helmrelease.yaml
   resources:
     limits:
       memory: 4Gi  # Increase from current value
     requests:
       memory: 2Gi
   ```

3. Commit and push changes, wait for Flux reconciliation:
   ```bash
   flux reconcile helmrelease -n minio minio
   ```

### Scenario D: ImagePullBackOff

**Likely Cause:** Registry issue or image doesn't exist

**Steps:**
1. Check image name and tag:
   ```bash
   kubectl get pods -n minio -l app=minio -o jsonpath='{.items[0].spec.containers[0].image}'
   ```

2. Verify image exists:
   ```bash
   docker pull <image-name>
   ```

3. Update image in HelmRelease if needed

### Scenario E: Node Resource Exhaustion

**Likely Cause:** Node has insufficient CPU/memory

**Steps:**
1. Check node resources:
   ```bash
   kubectl describe node <node-name> | grep -A 10 "Allocated resources"
   ```

2. If node is full, cordon and drain to move workloads:
   ```bash
   kubectl cordon <node-name>
   kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data
   ```

3. Or add node taint/toleration to spread load

## Verification

### 1. Check Pod Health

```bash
kubectl get pods -n minio
# All pods should be Running with 1/1 ready
```

### 2. Test API Endpoint

```bash
kubectl port-forward -n minio svc/minio 9000:9000

# Should return 200 OK
curl -I http://localhost:9000/minio/health/live
```

### 3. Test MinIO Client

```bash
mc alias set local http://localhost:9000 <access-key> <secret-key>
mc ls local/
```

### 4. Check Application Logs

```bash
kubectl logs -n minio -l app=minio --tail=50
# Should not show errors
```

## Prevention

1. **Set appropriate resource limits:**
   - Memory: 2Gi request, 4Gi limit
   - CPU: 500m request, 2000m limit

2. **Configure PersistentVolume with sufficient storage:**
   - Monitor storage usage
   - Set up alerts for >80% usage

3. **Enable Pod Disruption Budgets:**
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

4. **Set up health check alerts:**
   - Liveness probe failures
   - Readiness probe failures

5. **Regular backup verification:**
   - Test restore procedures
   - Verify bucket replication

## Related Alerts

- `MinIOHighErrorRate`
- `MinIOStorageSpaceLow`
- `MinIOPodNotReady`
- `MinIOHighMemoryUsage`

## Escalation

**When to escalate:**
- Service down > 15 minutes with no clear resolution
- Data corruption suspected
- Multiple restart attempts failed
- PersistentVolume data loss

**Escalation Path:**
1. Senior SRE Team
2. Storage Infrastructure Team
3. Vendor Support (MinIO)

## Additional Resources

- [MinIO Documentation](https://min.io/docs/minio/kubernetes/upstream/)
- [MinIO Troubleshooting Guide](https://min.io/docs/minio/linux/operations/troubleshooting.html)
- Internal Wiki: Storage Architecture
- Slack: #sre-alerts

