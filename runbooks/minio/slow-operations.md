# 🚨 Runbook: MinIO Slow Operations

## Alert Information

**Alert Name:** `MinIOSlowOperations`  
**Severity:** Medium  
**Component:** minio  
**Service:** object-storage

## Symptom

MinIO operations (GET, PUT, LIST) are taking significantly longer than normal. P99 latency exceeds acceptable thresholds.

## Impact

- **User Impact:** MEDIUM - Slow application performance
- **Business Impact:** MEDIUM - Degraded user experience
- **Data Impact:** NONE - No data loss, just performance degradation

## Diagnosis

### 1. Check Current Latency Metrics

```bash
# Port forward to MinIO
kubectl port-forward -n minio svc/minio 9000:9000

# Check metrics
curl http://localhost:9000/minio/v2/metrics/cluster | grep ttfb
```

### 2. Check MinIO Performance

```bash
# Run MinIO's built-in speedtest
kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- sh
mc admin speedtest local --size 64MiB --duration 30s
```

### 3. Check Resource Utilization

```bash
kubectl top pods -n minio
kubectl top nodes
```

### 4. Check Disk I/O

```bash
kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- sh

# Check disk I/O stats
iostat -x 5 3

# Or simple I/O test
dd if=/dev/zero of=/data/test.img bs=1M count=1000 oflag=direct
dd if=/data/test.img of=/dev/null bs=1M iflag=direct
rm /data/test.img
```

### 5. Check Network Latency

```bash
kubectl run -it --rm debug --image=nicolaka/netshoot --restart=Never -- sh
# Inside pod:
ping minio.minio.svc.cluster.local
curl -w "@curl-format.txt" -o /dev/null -s http://minio.minio.svc.cluster.local:9000/minio/health/live
```

### 6. Check for Large Objects

```bash
# List largest objects
mc ls --recursive local/ | sort -k4 -n -r | head -20
```

### 7. Check Concurrent Operations

```bash
kubectl logs -n minio -l app=minio --tail=100 | grep -c "API:"
# High count indicates many concurrent operations
```

## Resolution

### Scenario A: High CPU Usage

**Likely Cause:** Insufficient CPU resources or CPU throttling

**Steps:**
1. Check CPU throttling:
   ```bash
   kubectl describe pod -n minio -l app=minio | grep -A 5 "Limits\|Requests"
   ```

2. Increase CPU limits:
   ```yaml
   # Edit flux/clusters/homelab/infrastructure/minio/k8s/helmrelease.yaml
   resources:
     limits:
       cpu: 4000m  # Increase
       memory: 8Gi
     requests:
       cpu: 2000m  # Increase
       memory: 4Gi
   ```

3. Apply changes:
   ```bash
   flux reconcile helmrelease -n minio minio
   ```

### Scenario B: Disk I/O Bottleneck

**Likely Cause:** Slow storage backend or disk saturation

**Steps:**
1. Check storage class and backend:
   ```bash
   kubectl get pvc -n minio -o yaml | grep storageClassName
   kubectl get storageclass <storage-class-name> -o yaml
   ```

2. If using local storage, check node disk performance:
   ```bash
   kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].spec.nodeName}'
   # SSH to node and run fio benchmark
   ```

3. Migrate to faster storage class if available:
   ```yaml
   persistence:
     storageClass: fast-ssd  # Use SSD-backed storage
     size: 200Gi
   ```

4. Enable SSD/NVMe for hot data

5. Consider using distributed MinIO for better I/O parallelism

### Scenario C: Memory Pressure

**Likely Cause:** Insufficient memory for caching

**Steps:**
1. Check memory usage and cache hit rate:
   ```bash
   kubectl top pods -n minio
   curl http://localhost:9000/minio/v2/metrics/cluster | grep cache
   ```

2. Increase memory allocation:
   ```yaml
   resources:
     limits:
       memory: 16Gi  # Increase for more cache
     requests:
       memory: 8Gi
   ```

3. Configure cache settings:
   ```yaml
   environment:
     MINIO_CACHE: "on"
     MINIO_CACHE_DRIVES: "/cache"
     MINIO_CACHE_QUOTA: "80"
     MINIO_CACHE_AFTER: "3"
     MINIO_CACHE_WATERMARK_LOW: "70"
     MINIO_CACHE_WATERMARK_HIGH: "90"
   ```

### Scenario D: Network Latency

**Likely Cause:** Network congestion or misconfiguration

**Steps:**
1. Check network policies:
   ```bash
   kubectl get networkpolicies -n minio
   ```

2. Verify service mesh/proxy overhead:
   ```bash
   # If using Istio/Linkerd, check sidecar metrics
   kubectl logs -n minio -l app=minio -c istio-proxy --tail=50
   ```

3. Test direct pod connectivity:
   ```bash
   POD_IP=$(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].status.podIP}')
   kubectl run -it --rm debug --image=nicolaka/netshoot --restart=Never -- curl -w "@curl-format.txt" http://$POD_IP:9000/minio/health/live
   ```

4. Consider using hostNetwork for better performance (if security allows):
   ```yaml
   hostNetwork: true
   ```

### Scenario E: Too Many Small Operations

**Likely Cause:** Application making many small requests instead of batching

**Steps:**
1. Analyze request patterns:
   ```bash
   kubectl logs -n minio -l app=minio --tail=500 | grep "API:" | awk '{print $8}' | sort | uniq -c | sort -rn
   ```

2. Enable multipart uploads in client:
   ```python
   # For Python clients
   s3_client.upload_file(
       'large_file.bin',
       'bucket',
       'key',
       Config=TransferConfig(
           multipart_threshold=8388608,  # 8MB
           multipart_chunksize=8388608
       )
   )
   ```

3. Use batch operations where possible:
   ```bash
   # Instead of many single deletes, use batch delete
   mc rm --recursive local/bucket/prefix/
   ```

4. Implement connection pooling in application

### Scenario F: Large Object Transfers

**Likely Cause:** Single large object slowing down operations

**Steps:**
1. Enable multipart upload with optimal part size:
   ```yaml
   environment:
     MINIO_API_REPLICATION_WORKERS: "100"
     MINIO_API_REQUESTS_MAX: "10000"
   ```

2. Tune chunk sizes in client application

3. Use multiple connections for large transfers:
   ```bash
   mc cp --parallel 8 large-file.iso local/bucket/
   ```

### Scenario G: Distributed MinIO Node Issues

**Likely Cause:** One or more nodes in distributed setup are slow

**Steps:**
1. Check all nodes:
   ```bash
   mc admin info local
   ```

2. Check individual node health:
   ```bash
   mc admin trace local --call PUT
   # Look for slow nodes
   ```

3. Isolate slow node:
   ```bash
   kubectl get pods -n minio -o wide
   # Note which node is slow
   kubectl cordon <node-name>
   kubectl delete pod -n minio <slow-pod-name>
   ```

## Verification

### 1. Check Latency Improved

```bash
# Run speedtest
mc admin speedtest local --size 64MiB --duration 30s

# Check metrics
curl http://localhost:9000/minio/v2/metrics/cluster | grep ttfb
```

### 2. Test Operations

```bash
# Upload test
time mc cp /dev/urandom local/test-bucket/random.bin --attr "size=100MB"

# Download test
time mc cp local/test-bucket/random.bin /tmp/random.bin

# List test
time mc ls --recursive local/test-bucket/

# Cleanup
mc rm local/test-bucket/random.bin
rm /tmp/random.bin
```

### 3. Check Application Performance

Monitor application logs for improved response times.

## Prevention

1. **Set up latency monitoring:**
   ```yaml
   # Prometheus alerts
   - alert: MinIOSlowOperations
     expr: |
       histogram_quantile(0.99, 
         rate(minio_s3_ttfb_seconds_bucket[5m])
       ) > 5
     for: 10m
     annotations:
       summary: "MinIO P99 latency above 5 seconds"
   
   - alert: MinIOSlowPutOperations
     expr: |
       histogram_quantile(0.95,
         rate(minio_s3_requests_ttfb_seconds_bucket{api="putobject"}[5m])
       ) > 2
     for: 5m
     annotations:
       summary: "MinIO PUT operations slow"
   ```

2. **Use appropriate storage class:**
   - SSD/NVMe for hot data
   - HDD for cold/archive data

3. **Optimize MinIO configuration:**
   ```yaml
   environment:
     # Increase concurrent operations
     MINIO_API_REQUESTS_MAX: "10000"
     
     # Enable caching
     MINIO_CACHE: "on"
     
     # Optimize healing
     MINIO_HEAL_INTERVAL: "24h"
     
     # Batch operations
     MINIO_API_REPLICATION_WORKERS: "100"
   ```

4. **Client-side optimizations:**
   - Connection pooling
   - Multipart uploads for large files
   - Batch operations
   - Proper retry logic

5. **Resource planning:**
   - Size CPU/memory appropriately
   - Monitor resource utilization
   - Scale horizontally when needed

6. **Regular performance testing:**
   ```bash
   # Run monthly performance benchmarks
   mc admin speedtest local --duration 60s --size 128MiB
   ```

7. **Network optimization:**
   - Use dedicated network for storage if possible
   - Minimize hops between client and MinIO
   - Consider using jumbo frames for large transfers

## Metrics to Monitor

```promql
# P99 latency
histogram_quantile(0.99, rate(minio_s3_ttfb_seconds_bucket[5m]))

# P95 latency
histogram_quantile(0.95, rate(minio_s3_ttfb_seconds_bucket[5m]))

# Average latency by operation
rate(minio_s3_requests_ttfb_seconds_sum[5m]) / rate(minio_s3_requests_ttfb_seconds_count[5m])

# Disk I/O wait
rate(minio_node_disk_io_wait_time_seconds[5m])

# Network throughput
rate(minio_node_network_received_bytes_total[5m])
rate(minio_node_network_sent_bytes_total[5m])

# CPU usage
rate(minio_node_process_cpu_seconds_total[5m])
```

## Related Alerts

- `MinIODown`
- `MinIOHighErrorRate`
- `MinIOHighMemoryUsage`
- `MinIOHighCPUUsage`
- `MinIODiskSlow`

## Escalation

**When to escalate:**
- Latency consistently >10 seconds with no improvement
- Storage backend issues suspected
- Network infrastructure issues
- Requires infrastructure changes

**Escalation Path:**
1. Senior SRE Team
2. Storage Infrastructure Team
3. Network Team
4. MinIO Vendor Support

## Additional Resources

- [MinIO Performance Tuning](https://min.io/docs/minio/linux/operations/performance.html)
- [MinIO Caching](https://min.io/docs/minio/linux/operations/caching.html)
- [MinIO Hardware Requirements](https://min.io/docs/minio/linux/operations/hardware.html)
- Internal Wiki: Performance Optimization Guide
- Slack: #sre-performance

