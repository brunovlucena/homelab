# 🚨 Runbook: MinIO High Error Rate

## Alert Information

**Alert Name:** `MinIOHighErrorRate`  
**Severity:** High  
**Component:** minio  
**Service:** object-storage

## Symptom

MinIO is returning a high rate of error responses (4xx or 5xx HTTP codes), indicating issues with operations.

## Impact

- **User Impact:** HIGH - Object storage operations failing
- **Business Impact:** HIGH - Applications cannot store/retrieve data reliably
- **Data Impact:** LOW to MEDIUM - Depends on error type

## Diagnosis

### 1. Check Error Metrics

```bash
# Port forward to MinIO
kubectl port-forward -n minio svc/minio 9000:9000

# Query Prometheus for error rates
# Or check MinIO metrics endpoint
curl http://localhost:9000/minio/v2/metrics/cluster
```

### 2. Check MinIO Logs for Errors

```bash
kubectl logs -n minio -l app=minio --tail=200 | grep -E "ERROR|WARN|fail"
```

**Common error patterns:**
- "503 Service Unavailable" - Backend overload
- "403 Forbidden" - Permission issues
- "500 Internal Server Error" - Application errors
- "409 Conflict" - Concurrent write issues
- "429 Too Many Requests" - Rate limiting

### 3. Check Pod Status

```bash
kubectl get pods -n minio
kubectl top pods -n minio
```

### 4. Check Recent API Errors

```bash
# Get error breakdown by type
kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- mc admin trace -v --errors local
```

### 5. Check for Storage Issues

```bash
kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- df -h /data
```

### 6. Check Network Connectivity

```bash
# Check if clients can reach MinIO
kubectl run -it --rm debug --image=nicolaka/netshoot --restart=Never -- sh
# Inside pod:
curl -v http://minio.minio.svc.cluster.local:9000/minio/health/live
```

## Resolution

### Scenario A: 503 Service Unavailable

**Likely Cause:** MinIO overloaded or unhealthy

**Steps:**
1. Check resource utilization:
   ```bash
   kubectl top pods -n minio
   ```

2. Scale up replicas if using distributed mode:
   ```bash
   kubectl scale deployment/minio -n minio --replicas=4
   ```

3. Increase resource limits:
   ```yaml
   # Edit flux/clusters/homelab/infrastructure/minio/k8s/helmrelease.yaml
   resources:
     limits:
       cpu: 4000m
       memory: 8Gi
     requests:
       cpu: 2000m
       memory: 4Gi
   ```

4. Restart MinIO:
   ```bash
   kubectl rollout restart deployment/minio -n minio
   ```

### Scenario B: 403 Forbidden Errors

**Likely Cause:** Permission/authentication issues

**Steps:**
1. Verify credentials are correct:
   ```bash
   kubectl get secret -n minio minio -o jsonpath='{.data.rootUser}' | base64 -d
   kubectl get secret -n minio minio -o jsonpath='{.data.rootPassword}' | base64 -d
   ```

2. Check bucket policies:
   ```bash
   mc policy get local/<bucket-name>
   ```

3. Review IAM policies:
   ```bash
   mc admin user list local
   mc admin policy list local
   mc admin policy info local <policy-name>
   ```

4. Fix bucket policy if needed:
   ```bash
   # Set bucket to private
   mc policy set private local/<bucket-name>
   
   # Or set to public read
   mc policy set download local/<bucket-name>
   ```

5. Verify service account credentials:
   ```bash
   mc admin user info local <username>
   ```

### Scenario C: 500 Internal Server Errors

**Likely Cause:** Disk errors, corruption, or application bugs

**Steps:**
1. Check for disk errors in logs:
   ```bash
   kubectl logs -n minio -l app=minio --tail=500 | grep -i "disk\|corrupt\|read error\|write error"
   ```

2. Run disk health check:
   ```bash
   kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- sh
   cd /data
   ls -la
   # Check for .minio.sys/format.json
   cat .minio.sys/format.json
   ```

3. Check MinIO health:
   ```bash
   mc admin info local
   mc admin heal local
   ```

4. If corruption suspected, run healing:
   ```bash
   mc admin heal -r local/<bucket-name>
   ```

### Scenario D: 429 Too Many Requests

**Likely Cause:** Rate limiting or too many concurrent requests

**Steps:**
1. Check current API request rate:
   ```bash
   kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- mc admin trace local --call GET
   ```

2. Increase API request limits:
   ```yaml
   # In HelmRelease environment variables
   environment:
     MINIO_API_REQUESTS_MAX: "10000"
     MINIO_API_REQUESTS_DEADLINE: "10s"
   ```

3. Enable request queuing:
   ```yaml
   environment:
     MINIO_API_REQUESTS_MAX: "10000"
     MINIO_API_READY_DEADLINE: "10s"
   ```

4. Scale horizontally if needed

### Scenario E: 409 Conflict Errors

**Likely Cause:** Concurrent writes to same object or bucket

**Steps:**
1. Check for concurrent write patterns in application:
   ```bash
   kubectl logs -n <app-namespace> -l app=<app-name> | grep -i "minio\|s3"
   ```

2. Enable versioning to handle conflicts:
   ```bash
   mc version enable local/<bucket-name>
   ```

3. Implement retry logic with exponential backoff in application

4. Use unique object keys to avoid conflicts

### Scenario F: Connection Timeouts

**Likely Cause:** Network issues or slow backend

**Steps:**
1. Check network latency:
   ```bash
   kubectl run -it --rm debug --image=nicolaka/netshoot --restart=Never -- sh
   # Inside pod:
   ping minio.minio.svc.cluster.local
   ```

2. Increase timeout settings:
   ```yaml
   # In client configuration
   timeout: 60s
   ```

3. Check if disk I/O is slow:
   ```bash
   kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- sh
   # Run simple I/O test
   dd if=/dev/zero of=/data/test.img bs=1M count=1000 oflag=direct
   rm /data/test.img
   ```

## Verification

### 1. Check Error Rate

```bash
# Should drop below threshold
kubectl logs -n minio -l app=minio --tail=100 | grep -c ERROR
```

### 2. Test Basic Operations

```bash
# Create test object
echo "test" > test.txt
mc cp test.txt local/test-bucket/test.txt

# Read test object
mc cat local/test-bucket/test.txt

# Delete test object
mc rm local/test-bucket/test.txt
rm test.txt
```

### 3. Check Metrics

```bash
curl http://localhost:9000/minio/v2/metrics/cluster | grep minio_http_requests_error_total
```

### 4. Monitor Application Logs

```bash
kubectl logs -n <app-namespace> -l app=<app-name> --tail=50
# Should not show MinIO errors
```

## Prevention

1. **Set up proper monitoring:**
   ```yaml
   # Prometheus alerts
   - alert: MinIOHighErrorRate
     expr: |
       rate(minio_http_requests_error_total[5m]) 
       / rate(minio_http_requests_total[5m]) > 0.05
     for: 5m
     annotations:
       summary: "MinIO error rate above 5%"
   
   - alert: MinIOHigh5xxRate
     expr: |
       rate(minio_s3_requests_5xx_errors_total[5m]) > 10
     for: 2m
     annotations:
       summary: "MinIO returning many 5xx errors"
   ```

2. **Configure appropriate resource limits:**
   - CPU: 2-4 cores minimum
   - Memory: 4-8Gi minimum
   - Storage: Monitor and scale proactively

3. **Enable health checks:**
   ```yaml
   livenessProbe:
     httpGet:
       path: /minio/health/live
       port: 9000
     initialDelaySeconds: 30
     periodSeconds: 30
   
   readinessProbe:
     httpGet:
       path: /minio/health/ready
       port: 9000
     initialDelaySeconds: 15
     periodSeconds: 15
   ```

4. **Implement retry logic in clients:**
   - Exponential backoff
   - Circuit breaker pattern
   - Proper error handling

5. **Regular maintenance:**
   - Run healing regularly: `mc admin heal local`
   - Monitor disk health
   - Update to latest stable version

6. **Enable detailed logging:**
   ```yaml
   environment:
     MINIO_LOG_LEVEL: "DEBUG"
     MINIO_AUDIT_LOGGER_ENABLED: "on"
   ```

## Metrics to Monitor

```promql
# Overall error rate
rate(minio_http_requests_error_total[5m]) / rate(minio_http_requests_total[5m])

# 4xx error rate
rate(minio_s3_requests_4xx_errors_total[5m])

# 5xx error rate
rate(minio_s3_requests_5xx_errors_total[5m])

# Request latency
histogram_quantile(0.99, rate(minio_s3_ttfb_seconds_bucket[5m]))

# Failed operations
rate(minio_s3_requests_errors_total[5m])
```

## Related Alerts

- `MinIODown`
- `MinIOStorageSpaceLow`
- `MinIOSlowOperations`
- `MinIOHighMemoryUsage`
- `MinIODiskErrors`

## Escalation

**When to escalate:**
- Error rate >10% for >15 minutes
- 5xx errors indicating server issues
- Data corruption suspected
- Unable to identify root cause

**Escalation Path:**
1. Senior SRE Team
2. Application Team (if client-side issues)
3. Storage Infrastructure Team
4. MinIO Vendor Support

## Additional Resources

- [MinIO Error Codes](https://min.io/docs/minio/linux/reference/minio-cli/minio-mc.html#error-codes)
- [MinIO Troubleshooting](https://min.io/docs/minio/linux/operations/troubleshooting.html)
- [MinIO Monitoring Guide](https://min.io/docs/minio/linux/operations/monitoring.html)
- Internal Wiki: Object Storage Architecture
- Slack: #sre-storage

