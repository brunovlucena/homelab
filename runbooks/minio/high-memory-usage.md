# 🚨 Runbook: MinIO High Memory Usage

## Alert Information

**Alert Name:** `MinIOHighMemoryUsage`  
**Severity:** Medium  
**Component:** minio  
**Service:** object-storage

## Symptom

MinIO pod is consuming excessive memory, approaching or exceeding allocated limits. May lead to OOMKill.

## Impact

- **User Impact:** MEDIUM - Potential service interruption if OOMKilled
- **Business Impact:** MEDIUM - Risk of service unavailability
- **Data Impact:** NONE - No data loss from high memory usage

## Diagnosis

### 1. Check Current Memory Usage

```bash
kubectl top pods -n minio
kubectl describe pod -n minio -l app=minio | grep -A 10 "Limits\|Requests"
```

### 2. Check Memory Metrics

```bash
# Port forward to MinIO
kubectl port-forward -n minio svc/minio 9000:9000

# Check memory metrics
curl http://localhost:9000/minio/v2/metrics/cluster | grep memory
```

### 3. Check for OOMKill Events

```bash
kubectl get events -n minio --sort-by='.lastTimestamp' | grep -i "oom\|memory"
kubectl describe pod -n minio -l app=minio | grep -i "oom\|terminated"
```

### 4. Check MinIO Process Memory

```bash
kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- sh

# Check process memory
ps aux | grep minio
top -b -n 1 | grep minio

# Check memory details
cat /proc/$(pidof minio)/status | grep -i mem
```

### 5. Check Connection Count

```bash
# High connection count can increase memory usage
kubectl logs -n minio -l app=minio --tail=100 | grep -i "connection\|client"

# Check active connections
kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- sh
netstat -an | grep :9000 | grep ESTABLISHED | wc -l
```

### 6. Check for Memory Leaks

```bash
kubectl logs -n minio -l app=minio --tail=500 | grep -i "leak\|gc\|heap"
```

### 7. Check Cache Configuration

```bash
mc admin config get local cache
```

## Resolution

### Scenario A: Memory Limit Too Low

**Likely Cause:** Insufficient memory allocated for workload

**Steps:**
1. Check current allocation:
   ```bash
   kubectl get deployment -n minio minio -o jsonpath='{.spec.template.spec.containers[0].resources}'
   ```

2. Increase memory limits:
   ```yaml
   # Edit flux/clusters/homelab/infrastructure/minio/k8s/helmrelease.yaml
   resources:
     limits:
       memory: 16Gi  # Increase from current (e.g., 4Gi -> 16Gi)
       cpu: 4000m
     requests:
       memory: 8Gi   # Increase proportionally
       cpu: 2000m
   ```

3. Apply changes:
   ```bash
   git add .
   git commit -m "Increase MinIO memory limits"
   git push
   flux reconcile helmrelease -n minio minio
   ```

4. Monitor new pod:
   ```bash
   kubectl get pods -n minio -w
   kubectl top pods -n minio
   ```

### Scenario B: Excessive Cache Size

**Likely Cause:** Cache configured too large or inefficiently

**Steps:**
1. Check cache settings:
   ```bash
   mc admin config get local cache
   ```

2. Reduce cache quota:
   ```bash
   mc admin config set local cache \
     quota=70 \
     after=3 \
     watermark_low=60 \
     watermark_high=80
   
   mc admin service restart local
   ```

3. Or disable cache if not needed:
   ```bash
   mc admin config set local cache enable=off
   mc admin service restart local
   ```

4. Alternatively, configure via environment variables:
   ```yaml
   # In HelmRelease
   environment:
     MINIO_CACHE: "on"
     MINIO_CACHE_QUOTA: "60"  # Reduce from higher value
     MINIO_CACHE_WATERMARK_LOW: "50"
     MINIO_CACHE_WATERMARK_HIGH: "70"
   ```

### Scenario C: Too Many Concurrent Connections

**Likely Cause:** High number of clients or slow clients holding connections

**Steps:**
1. Check connection count:
   ```bash
   kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- sh
   netstat -an | grep :9000 | wc -l
   ```

2. Limit concurrent connections:
   ```yaml
   # In HelmRelease environment
   environment:
     MINIO_API_REQUESTS_MAX: "5000"  # Reduce from default
     MINIO_API_REQUESTS_DEADLINE: "10s"
   ```

3. Implement connection pooling in clients:
   ```python
   # Python example
   from minio import Minio
   from urllib3.poolmanager import PoolManager
   
   client = Minio(
       "minio:9000",
       access_key="access-key",
       secret_key="secret-key",
       http_client=PoolManager(
           maxsize=10,  # Limit connections per client
           timeout=30
       )
   )
   ```

4. Scale horizontally if many legitimate clients:
   ```bash
   kubectl scale deployment/minio -n minio --replicas=3
   ```

### Scenario D: Memory Leak in MinIO

**Likely Cause:** Bug in MinIO version

**Steps:**
1. Check MinIO version:
   ```bash
   mc admin info local | grep Version
   ```

2. Check for known issues:
   ```bash
   # Search MinIO GitHub issues for memory leak
   # https://github.com/minio/minio/issues?q=memory+leak
   ```

3. Upgrade to latest version:
   ```yaml
   # Edit flux/clusters/homelab/infrastructure/minio/k8s/helmrelease.yaml
   image:
     tag: RELEASE.2024-01-01T00-00-00Z  # Use latest stable
   ```

4. As temporary workaround, schedule periodic restarts:
   ```yaml
   apiVersion: batch/v1
   kind: CronJob
   metadata:
     name: minio-restart
     namespace: minio
   spec:
     schedule: "0 3 * * 0"  # Weekly at 3 AM Sunday
     jobTemplate:
       spec:
         template:
           spec:
             serviceAccountName: minio-restart
             containers:
             - name: kubectl
               image: bitnami/kubectl:latest
               command:
               - /bin/sh
               - -c
               - kubectl rollout restart deployment/minio -n minio
             restartPolicy: OnFailure
   ```

### Scenario E: Large Object Uploads

**Likely Cause:** Large objects being held in memory during multipart uploads

**Steps:**
1. Check ongoing multipart uploads:
   ```bash
   mc ls --recursive --incomplete local/
   ```

2. Tune multipart settings:
   ```yaml
   environment:
     MINIO_API_REPLICATION_WORKERS: "50"  # Reduce if too high
     MINIO_API_REQUESTS_MAX: "5000"
   ```

3. Configure smaller multipart chunk size in clients:
   ```python
   from boto3.s3.transfer import TransferConfig
   
   config = TransferConfig(
       multipart_threshold=8 * 1024 * 1024,  # 8 MB
       multipart_chunksize=8 * 1024 * 1024,  # 8 MB
       max_concurrency=10
   )
   ```

4. Clean up old incomplete uploads:
   ```bash
   mc rm --recursive --incomplete --force local/<bucket-name>/
   ```

### Scenario F: Metadata Overhead

**Likely Cause:** Too many small objects causing metadata bloat

**Steps:**
1. Check object count:
   ```bash
   mc du local/<bucket-name>
   ```

2. If millions of tiny objects, consider:
   - Aggregating small files
   - Using object lifecycle to clean up
   - Moving to a different storage solution for small files

3. Set up lifecycle policy:
   ```bash
   mc ilm add local/<bucket-name> --expiry-days 30 --prefix "temp/"
   ```

## Verification

### 1. Check Memory Usage Stabilized

```bash
kubectl top pods -n minio
# Monitor over 10-15 minutes
watch kubectl top pods -n minio
```

Memory should be well below limits and not continuously increasing.

### 2. Check No OOM Events

```bash
kubectl get events -n minio --sort-by='.lastTimestamp' | grep -i oom
# Should return no recent events
```

### 3. Test MinIO Functionality

```bash
# Upload/download test
echo "test" > test.txt
mc cp test.txt local/test-bucket/test.txt
mc cp local/test-bucket/test.txt test-download.txt
mc rm local/test-bucket/test.txt
rm test.txt test-download.txt
```

### 4. Monitor Metrics

```bash
curl http://localhost:9000/minio/v2/metrics/cluster | grep minio_node_process_resident_memory_bytes
```

## Prevention

1. **Set appropriate memory limits with headroom:**
   ```yaml
   resources:
     limits:
       memory: 16Gi  # Leave 25-30% headroom
     requests:
       memory: 8Gi
   ```

2. **Monitor memory usage proactively:**
   ```yaml
   # Prometheus alerts
   - alert: MinIOHighMemoryUsage
     expr: |
       container_memory_usage_bytes{namespace="minio",pod=~"minio-.*"}
       / container_spec_memory_limit_bytes{namespace="minio",pod=~"minio-.*"}
       > 0.80
     for: 10m
     annotations:
       summary: "MinIO memory usage above 80%"
   
   - alert: MinIOMemoryLeak
     expr: |
       rate(container_memory_usage_bytes{namespace="minio"}[1h]) > 0
     for: 4h
     annotations:
       summary: "MinIO memory usage continuously increasing"
   ```

3. **Configure cache conservatively:**
   ```yaml
   environment:
     MINIO_CACHE_QUOTA: "60"  # Max 60% of disk for cache
     MINIO_CACHE_WATERMARK_LOW: "50"
     MINIO_CACHE_WATERMARK_HIGH: "70"
   ```

4. **Implement connection limits:**
   ```yaml
   environment:
     MINIO_API_REQUESTS_MAX: "5000"
     MINIO_API_REQUESTS_DEADLINE: "10s"
   ```

5. **Regular cleanup:**
   - Delete incomplete multipart uploads
   - Clean up old temporary data
   - Archive old buckets

6. **Keep MinIO updated:**
   - Watch release notes for memory improvements
   - Test upgrades in non-production first
   - Plan regular upgrade schedule

7. **Use horizontal scaling:**
   - Distribute load across multiple pods
   - Reduces memory pressure per pod

## Metrics to Monitor

```promql
# Memory usage percentage
container_memory_usage_bytes{namespace="minio"} 
/ container_spec_memory_limit_bytes{namespace="minio"}

# Memory growth rate
rate(container_memory_usage_bytes{namespace="minio"}[1h])

# MinIO heap usage
minio_node_process_resident_memory_bytes

# Cache usage
minio_cache_usage_percent

# Connection count
minio_s3_requests_current
```

## Related Alerts

- `MinIODown`
- `MinIOPodOOMKilled`
- `MinIOHighCPUUsage`
- `MinIOSlowOperations`

## Escalation

**When to escalate:**
- Memory leak confirmed with no known fix
- Repeated OOMKills despite increase in limits
- Memory requirements exceed infrastructure capacity
- Suspected MinIO bug

**Escalation Path:**
1. Senior SRE Team
2. Application Team (if client-side issue)
3. Infrastructure Team (for capacity)
4. MinIO Vendor Support

## Additional Resources

- [MinIO Memory Requirements](https://min.io/docs/minio/linux/operations/hardware.html#memory)
- [MinIO Caching](https://min.io/docs/minio/linux/operations/caching.html)
- [MinIO Performance Tuning](https://min.io/docs/minio/linux/operations/performance.html)
- Internal Wiki: Memory Sizing Guidelines
- Slack: #sre-alerts

