# 🚨 Runbook: Loki Service Down

## Alert Information

**Alert Name:** `LokiDown`  
**Severity:** Critical  
**Component:** Loki  
**Service:** Log Aggregation

## Symptom

Loki service is completely unavailable. No logs can be queried or ingested.

## Impact

- **User Impact:** CRITICAL - No log access for debugging and troubleshooting
- **Business Impact:** HIGH - Inability to diagnose production issues
- **Data Impact:** POTENTIAL - Logs may be buffered or lost during downtime

## Diagnosis

### 1. Check Loki Pod Status

```bash
kubectl get pods -n loki
kubectl describe pod -n loki -l app.kubernetes.io/name=loki
```

### 2. Check All Loki Components

```bash
# Check backend pods
kubectl get pods -n loki -l app.kubernetes.io/component=backend

# Check read pods
kubectl get pods -n loki -l app.kubernetes.io/component=read

# Check write pods
kubectl get pods -n loki -l app.kubernetes.io/component=write
```

### 3. Check Loki Logs

```bash
# Backend logs
kubectl logs -n loki -l app.kubernetes.io/component=backend --tail=100

# Read logs
kubectl logs -n loki -l app.kubernetes.io/component=read --tail=100

# Write logs
kubectl logs -n loki -l app.kubernetes.io/component=write --tail=100
```

### 4. Check Service and Endpoints

```bash
kubectl get svc -n loki
kubectl get endpoints -n loki loki-gateway
```

### 5. Check Events

```bash
kubectl get events -n loki --sort-by='.lastTimestamp' | head -30
```

### 6. Check Resource Usage

```bash
kubectl top pods -n loki
```

## Resolution Steps

### Step 1: Identify which component is down

```bash
# Check all components
kubectl get pods -n loki -o wide
```

### Step 2: Check for common issues

```bash
# Check for CrashLoopBackOff
kubectl get pods -n loki | grep -i crash

# Check for ImagePullBackOff
kubectl get pods -n loki | grep -i imagepull

# Check for pending pods
kubectl get pods -n loki | grep -i pending
```

### Step 3: Common Issues and Fixes

#### Issue: All pods in CrashLoopBackOff
**Cause:** Configuration error or storage backend unavailable  
**Fix:**
```bash
# Check MinIO (storage backend) status
kubectl get pods -n loki -l app.kubernetes.io/name=minio

# Check Loki configuration
kubectl get secret -n loki loki-minio-secret
kubectl describe helmrelease -n loki loki

# Check logs for specific error
kubectl logs -n loki -l app.kubernetes.io/component=backend --tail=50 | grep -i error
```

#### Issue: MinIO backend unavailable
**Cause:** Storage backend not responding  
**Fix:**
```bash
# Restart MinIO
kubectl rollout restart statefulset -n loki loki-minio

# Wait for MinIO to be ready
kubectl wait --for=condition=ready pod -n loki -l app.kubernetes.io/name=minio --timeout=300s

# Then restart Loki components
kubectl rollout restart statefulset -n loki loki-backend
kubectl rollout restart deployment -n loki loki-read
kubectl rollout restart deployment -n loki loki-write
```

#### Issue: Storage secret missing or incorrect
**Cause:** MinIO credentials not configured properly  
**Fix:**
```bash
# Verify secret exists
kubectl get secret -n loki loki-minio-secret

# If missing, recreate the secret (check scripts/create-secrets.sh)
# Then force Flux to reconcile
flux reconcile helmrelease loki -n loki
```

#### Issue: OOMKilled pods
**Cause:** Out of memory  
**Fix:**
```bash
# Check which component is OOMKilled
kubectl describe pod -n loki <pod-name> | grep -A 10 "Last State"

# Temporarily increase memory limit
kubectl edit helmrelease -n loki loki
# Increase memory limits for the affected component
```

### Step 4: Restart Loki components in order

```bash
# 1. First ensure MinIO is healthy
kubectl get pods -n loki -l app.kubernetes.io/name=minio

# 2. Restart backend (handles schema operations)
kubectl rollout restart statefulset -n loki loki-backend
kubectl rollout status statefulset -n loki loki-backend

# 3. Restart write path
kubectl rollout restart deployment -n loki loki-write
kubectl rollout status deployment -n loki loki-write

# 4. Restart read path
kubectl rollout restart deployment -n loki loki-read
kubectl rollout status deployment -n loki loki-read
```

### Step 5: Force Flux reconciliation if needed

```bash
flux reconcile helmrelease loki -n loki
```

## Verification

1. Check all pods are running:
```bash
kubectl get pods -n loki
```

2. Verify Loki endpoints are healthy:
```bash
# Port forward to Loki gateway
kubectl port-forward -n loki svc/loki-gateway 3100:80

# Check ready endpoint
curl http://localhost:3100/ready

# Check metrics endpoint
curl http://localhost:3100/metrics
```

3. Test log query:
```bash
# Query recent logs
curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={namespace="loki"}' | jq .
```

4. Verify from Grafana:
```bash
# Access Grafana and run a LogQL query
# Navigate to Explore -> Loki datasource
# Query: {namespace="loki"}
```

## Prevention

1. **Monitor storage backend health**
   - Set up MinIO monitoring alerts
   - Monitor disk usage on MinIO PVC

2. **Set appropriate resource limits**
   - Review and adjust memory/CPU based on log volume
   - Implement pod disruption budgets

3. **Enable retention policies**
   - Configure appropriate retention period (currently 744h/31 days)
   - Monitor storage growth

4. **Implement backup strategy**
   - Backup MinIO data regularly
   - Document recovery procedures

5. **Monitor component health**
   - Set up alerts for pod restarts
   - Monitor ingestion rate and query performance

## Related Alerts

- `LokiWritePathDown`
- `LokiReadPathSlow`
- `LokiIngestionErrors`
- `LokiStorageIssues`
- `LokiHighMemory`

## Escalation

If Loki remains down after following these steps:
1. Check MinIO data integrity
2. Review recent Helm chart updates
3. Check cluster-wide resource constraints
4. Contact platform team or Loki maintainers

## Additional Resources

- [Loki Documentation](https://grafana.com/docs/loki/latest/)
- [Loki Operations Guide](https://grafana.com/docs/loki/latest/operations/)
- [Troubleshooting Loki](https://grafana.com/docs/loki/latest/operations/troubleshooting/)
- [MinIO Operations](https://min.io/docs/minio/kubernetes/upstream/)
- [Loki Configuration](../../flux/clusters/homelab/infrastructure/loki/helmrelease.yaml)

