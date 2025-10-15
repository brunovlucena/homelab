# MinIO Runbooks

This directory contains operational runbooks for troubleshooting and maintaining MinIO object storage in the homelab infrastructure.

## 📚 Available Runbooks

### Critical Issues

1. **[MinIO Down](minio-down.md)** ⚠️ Critical
   - **When to use:** MinIO service is completely unavailable
   - **Common causes:** Pod crashes, PV issues, OOM kills, configuration errors
   - **Impact:** All object storage operations fail

2. **[Authentication Issues](authentication-issues.md)** ⚠️ High
   - **When to use:** 403 Forbidden or 401 Unauthorized errors
   - **Common causes:** Invalid credentials, policy issues, signature mismatch
   - **Impact:** Users/applications cannot access storage

3. **[High Error Rate](high-error-rate.md)** ⚠️ High
   - **When to use:** Elevated 4xx or 5xx error responses
   - **Common causes:** Overload, permission issues, disk errors, rate limiting
   - **Impact:** Operations failing frequently

### Capacity & Performance

4. **[Storage Space Low](storage-space-low.md)** ⚠️ High
   - **When to use:** Disk usage above 80%
   - **Common causes:** Data growth, lack of lifecycle policies, old backups
   - **Impact:** Risk of storage full, uploads may fail

5. **[Slow Operations](slow-operations.md)** ⚠️ Medium
   - **When to use:** High latency on GET/PUT/LIST operations
   - **Common causes:** Disk I/O bottleneck, insufficient resources, network latency
   - **Impact:** Degraded application performance

6. **[High Memory Usage](high-memory-usage.md)** ⚠️ Medium
   - **When to use:** Memory usage approaching limits
   - **Common causes:** Excessive cache, too many connections, memory leak
   - **Impact:** Risk of OOM kill and service interruption

### Operational Issues

7. **[Pod Not Ready](pod-not-ready.md)** ⚠️ High
   - **When to use:** Pod exists but readiness probe failing
   - **Common causes:** Slow startup, PV issues, resource constraints, config errors
   - **Impact:** Service unavailable or degraded

## 🚨 Quick Triage Guide

### Step 1: Identify the Symptom

```bash
# Check pod status
kubectl get pods -n minio

# Check recent events
kubectl get events -n minio --sort-by='.lastTimestamp' | head -20

# Check logs
kubectl logs -n minio -l app=minio --tail=50
```

### Step 2: Match to Runbook

| Symptom | Runbook | Priority |
|---------|---------|----------|
| Pod not running or CrashLoopBackOff | [MinIO Down](minio-down.md) | P0 |
| Pod running but not ready | [Pod Not Ready](pod-not-ready.md) | P1 |
| 403/401 errors in logs | [Authentication Issues](authentication-issues.md) | P1 |
| High error rate in metrics | [High Error Rate](high-error-rate.md) | P1 |
| Disk usage >80% | [Storage Space Low](storage-space-low.md) | P2 |
| Slow requests (P99 latency high) | [Slow Operations](slow-operations.md) | P2 |
| Memory usage >80% | [High Memory Usage](high-memory-usage.md) | P2 |

### Step 3: Follow Runbook

Each runbook follows a consistent structure:
1. **Alert Information** - Alert details and severity
2. **Symptom** - What you're observing
3. **Impact** - User, business, and data impact
4. **Diagnosis** - How to investigate
5. **Resolution** - Step-by-step fixes for different scenarios
6. **Verification** - How to confirm the issue is resolved
7. **Prevention** - How to avoid the issue in the future

## 🔧 Common Commands

### Health Checks

```bash
# Check overall health
kubectl get pods -n minio
mc admin info local

# Test readiness endpoint
kubectl port-forward -n minio svc/minio 9000:9000
curl http://localhost:9000/minio/health/ready

# Test liveness endpoint
curl http://localhost:9000/minio/health/live

# Check metrics
curl http://localhost:9000/minio/v2/metrics/cluster
```

### Resource Monitoring

```bash
# Pod resource usage
kubectl top pods -n minio

# Storage usage
kubectl exec -it -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- df -h /data

# Bucket sizes
mc du local/ --depth 1

# Object count
mc ls --recursive local/ | wc -l
```

### Log Analysis

```bash
# Recent errors
kubectl logs -n minio -l app=minio --tail=100 | grep -E "ERROR|WARN|fail"

# Authentication errors
kubectl logs -n minio -l app=minio --tail=200 | grep -i "auth\|forbidden\|403"

# Performance issues
kubectl logs -n minio -l app=minio --tail=100 | grep -i "slow\|timeout"

# Follow logs in real-time
kubectl logs -n minio -l app=minio -f
```

### Configuration

```bash
# Get current configuration
mc admin config export local

# View specific config section
mc admin config get local <section>

# Examples:
mc admin config get local cache
mc admin config get local api
mc admin config get local identity_ldap
```

### User & Policy Management

```bash
# List users
mc admin user list local

# Check user info
mc admin user info local <username>

# List policies
mc admin policy list local

# Check policy details
mc admin policy info local <policy-name>

# List service accounts
mc admin user svcacct list local <username>
```

### Bucket Operations

```bash
# List buckets
mc ls local/

# Get bucket policy
mc policy get local/<bucket-name>

# Check versioning
mc version info local/<bucket-name>

# Check lifecycle policies
mc ilm ls local/<bucket-name>

# Bucket statistics
mc stat local/<bucket-name>
```

### Troubleshooting

```bash
# Run healing (for distributed setup)
mc admin heal local

# Run speedtest
mc admin speedtest local --size 64MiB --duration 30s

# Trace API calls
mc admin trace local

# Trace only errors
mc admin trace --errors local

# Console logs
mc admin logs local

# Server info
mc admin info local --json
```

## 📊 Monitoring & Alerts

### Key Metrics to Monitor

1. **Availability**
   - Pod ready status
   - Health endpoint response
   - Service endpoint reachability

2. **Performance**
   - Request latency (P50, P95, P99)
   - Throughput (requests/sec)
   - Error rate (4xx, 5xx)

3. **Resource Usage**
   - CPU utilization
   - Memory utilization
   - Disk usage
   - Network throughput

4. **Operations**
   - GET operation latency
   - PUT operation latency
   - DELETE operation latency
   - LIST operation latency

### Recommended Alerts

```yaml
# Service Down
- alert: MinIODown
  expr: up{job="minio"} == 0
  for: 2m

# High Error Rate
- alert: MinIOHighErrorRate
  expr: rate(minio_s3_requests_errors_total[5m]) > 10
  for: 5m

# Storage Space Low
- alert: MinIOStorageSpaceLow
  expr: (minio_cluster_disk_total_bytes - minio_cluster_disk_free_bytes) / minio_cluster_disk_total_bytes > 0.80
  for: 5m

# High Memory
- alert: MinIOHighMemoryUsage
  expr: container_memory_usage_bytes{namespace="minio"} / container_spec_memory_limit_bytes{namespace="minio"} > 0.80
  for: 10m

# Slow Operations
- alert: MinIOSlowOperations
  expr: histogram_quantile(0.99, rate(minio_s3_ttfb_seconds_bucket[5m])) > 5
  for: 10m

# Pod Not Ready
- alert: MinIOPodNotReady
  expr: kube_pod_status_ready{namespace="minio",condition="true"} == 0
  for: 5m
```

## 🔗 Related Documentation

- [MinIO Official Documentation](https://min.io/docs/minio/kubernetes/upstream/)
- [MinIO Troubleshooting Guide](https://min.io/docs/minio/linux/operations/troubleshooting.html)
- [MinIO Monitoring Guide](https://min.io/docs/minio/linux/operations/monitoring.html)
- [Kubernetes Documentation](https://kubernetes.io/docs/home/)

## 📞 Escalation

### When to Escalate

Escalate to the next level if:
- Issue persists >15 minutes after following runbook
- Root cause is unclear or outside scope
- Requires infrastructure changes
- Data loss or corruption suspected
- Security incident suspected

### Escalation Path

1. **Level 1: Senior SRE Team**
   - Complex technical issues
   - Multiple system involvement
   - Slack: #sre-oncall

2. **Level 2: Specialized Teams**
   - Storage Infrastructure Team (storage backend issues)
   - Security Team (authentication/security issues)
   - Network Team (connectivity issues)
   - Slack: #infrastructure

3. **Level 3: Vendor Support**
   - MinIO vendor support
   - Storage vendor support
   - Requires architectural changes

## 🤝 Contributing

When adding new runbooks:

1. Follow the standard template structure
2. Include practical, tested commands
3. Provide clear diagnosis steps
4. Offer multiple resolution scenarios
5. Add verification steps
6. Include prevention measures
7. Link to related runbooks and alerts

## 📝 Maintenance

These runbooks should be reviewed and updated:
- After each incident (capture new scenarios)
- Quarterly (ensure commands/procedures still valid)
- After MinIO upgrades (verify compatibility)
- When infrastructure changes

---

**Last Updated:** 2025-10-15  
**Maintained By:** SRE Team  
**Questions:** #sre-support

