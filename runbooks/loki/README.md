# 📚 Loki Runbooks

Comprehensive operational runbooks for troubleshooting and resolving Loki issues in the homelab Kubernetes cluster.

## Overview

Loki is our log aggregation system, running in a distributed mode with:
- **Storage Backend**: MinIO (S3-compatible) with 50Gi storage
- **Retention Period**: 744 hours (31 days)
- **Architecture**: 
  - 1x Backend (schema operations, compaction)
  - 2x Read replicas (query processing)
  - 2x Write replicas (log ingestion)
  - 1x MinIO instance (storage backend)

## Quick Reference

| Alert | Severity | Impact | Runbook |
|-------|----------|--------|---------|
| LokiDown | Critical | Complete service outage | [loki-down.md](./loki-down.md) |
| LokiWritePathDown | Critical | No new logs being stored | [loki-write-path-down.md](./loki-write-path-down.md) |
| LokiReadPathSlow | Warning | Slow log queries | [loki-read-path-slow.md](./loki-read-path-slow.md) |
| LokiIngestionErrors | Warning | Logs being rejected/dropped | [loki-ingestion-errors.md](./loki-ingestion-errors.md) |
| LokiStorageIssues | Critical | Storage backend problems | [loki-storage-issues.md](./loki-storage-issues.md) |
| LokiHighMemory | Warning | Memory pressure/OOMKills | [loki-high-memory.md](./loki-high-memory.md) |
| LokiQueryTimeout | Warning | Queries timing out | [loki-query-timeout.md](./loki-query-timeout.md) |
| LokiRetentionIssues | Warning | Data not being cleaned up | [loki-retention-compaction.md](./loki-retention-compaction.md) |

## Runbooks

### 🚨 Critical Issues

#### [Loki Service Down](./loki-down.md)
Complete Loki outage - no logs can be queried or ingested.

**Quick Check:**
```bash
kubectl get pods -n loki
```

**Quick Fix:**
```bash
# Restart all components
kubectl rollout restart statefulset -n loki loki-backend
kubectl rollout restart deployment -n loki loki-write
kubectl rollout restart deployment -n loki loki-read
```

---

#### [Write Path Down](./loki-write-path-down.md)
Log ingestion failing - new logs not being stored.

**Quick Check:**
```bash
kubectl get pods -n loki -l app.kubernetes.io/component=write
kubectl logs -n loki -l app.kubernetes.io/component=write --tail=50
```

**Quick Fix:**
```bash
# Restart write pods
kubectl rollout restart deployment -n loki loki-write
```

---

#### [Storage Issues](./loki-storage-issues.md)
MinIO/S3 storage backend problems - reads or writes failing.

**Quick Check:**
```bash
kubectl get pods -n loki -l app.kubernetes.io/name=minio
kubectl exec -n loki <minio-pod> -- df -h /export
```

**Quick Fix:**
```bash
# Restart MinIO
kubectl rollout restart statefulset -n loki loki-minio
# Wait for ready, then restart Loki
kubectl rollout restart statefulset -n loki loki-backend
```

---

### ⚠️ Warning Issues

#### [Read Path Slow](./loki-read-path-slow.md)
Queries taking longer than expected - slow dashboard and searches.

**Quick Check:**
```bash
kubectl top pods -n loki -l app.kubernetes.io/component=read
kubectl port-forward -n loki svc/loki-gateway 3100:80
time curl -G "http://localhost:3100/loki/api/v1/query" --data-urlencode 'query={namespace="loki"}'
```

**Quick Fix:**
```bash
# Scale read replicas
kubectl scale deployment -n loki loki-read --replicas=3
```

---

#### [Ingestion Errors](./loki-ingestion-errors.md)
Logs being rejected or dropped - partial log loss.

**Quick Check:**
```bash
kubectl logs -n loki -l app.kubernetes.io/component=write --tail=100 | grep -i "error\|reject"
kubectl port-forward -n loki svc/loki-gateway 3100:80
curl http://localhost:3100/metrics | grep loki_discarded
```

**Quick Fix:**
```bash
# Increase rate limits
kubectl edit helmrelease -n loki loki
# Update: loki.limits_config.ingestion_rate_mb: 10
```

---

#### [High Memory Usage](./loki-high-memory.md)
Memory pressure or OOMKills affecting performance.

**Quick Check:**
```bash
kubectl top pods -n loki
kubectl get pods -n loki -o json | jq -r '.items[] | select(.status.containerStatuses[].lastState.terminated.reason == "OOMKilled") | .metadata.name'
```

**Quick Fix:**
```bash
# Increase memory limits
kubectl edit helmrelease -n loki loki
# Update component memory limits
```

---

#### [Query Timeout](./loki-query-timeout.md)
Queries timing out before completion.

**Quick Check:**
```bash
kubectl logs -n loki -l app.kubernetes.io/component=read --tail=100 | grep -i "timeout"
kubectl port-forward -n loki svc/loki-gateway 3100:80
time curl -G "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={namespace="loki"}' \
  --data-urlencode 'start=now-24h' --data-urlencode 'end=now'
```

**Quick Fix:**
```bash
# Increase timeout
kubectl edit helmrelease -n loki loki
# Add: loki.limits_config.query_timeout: 10m
```

---

#### [Retention & Compaction Issues](./loki-retention-compaction.md)
Old logs not being cleaned up - storage growing unexpectedly.

**Quick Check:**
```bash
kubectl exec -n loki <minio-pod> -- df -h /export
kubectl logs -n loki -l app.kubernetes.io/component=backend --tail=200 | grep -i compact
```

**Quick Fix:**
```bash
# Verify retention config and restart backend
kubectl get helmrelease -n loki loki -o yaml | grep retention
kubectl rollout restart statefulset -n loki loki-backend
```

---

## Common Troubleshooting Commands

### Check Overall Health
```bash
# All pods status
kubectl get pods -n loki

# Resource usage
kubectl top pods -n loki

# Recent events
kubectl get events -n loki --sort-by='.lastTimestamp' | head -20
```

### Check Loki Metrics
```bash
# Port forward to Loki
kubectl port-forward -n loki svc/loki-gateway 3100:80

# Health endpoints
curl http://localhost:3100/ready
curl http://localhost:3100/metrics

# Ingestion rate
curl http://localhost:3100/metrics | grep loki_distributor_bytes_received_total

# Query performance
curl http://localhost:3100/metrics | grep loki_query_duration_seconds
```

### Check Storage
```bash
# MinIO pod status
kubectl get pods -n loki -l app.kubernetes.io/name=minio

# Disk usage
kubectl exec -n loki <minio-pod> -- df -h /export

# Object count
kubectl exec -n loki <minio-pod> -- mc ls --recursive local/loki | wc -l
```

### Test Functionality
```bash
# Test ingestion
kubectl run test-logger --image=busybox --restart=Never -- sh -c "echo 'Test log'"

# Test query
kubectl port-forward -n loki svc/loki-gateway 3100:80
curl -G "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={namespace="default"}' | jq .

# Cleanup
kubectl delete pod test-logger
```

## Architecture

```
┌─────────────┐
│   Alloy     │ Log collectors
│  (agents)   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Loki Write │◄── Ingestion
│  (2 replicas│
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Backend   │◄── Schema, Compaction
│  (1 replica)│
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    MinIO    │◄── S3 Storage
│   (50Gi)    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Loki Read  │◄── Queries
│  (2 replicas│
└─────────────┘
```

## Configuration

**Location**: `flux/clusters/homelab/infrastructure/loki/helmrelease.yaml`

**Key Settings**:
- Retention: 744h (31 days)
- Storage: MinIO S3-compatible
- Schema: TSDB (v13)
- Auth: Disabled (single-tenant)

## Labels and Cardinality

**Recommended labels** (keep under 10 per stream):
- ✅ `namespace` - Kubernetes namespace
- ✅ `app` - Application name
- ✅ `component` - Component within app
- ✅ `environment` - prod/dev/staging
- ✅ `cluster` - Cluster identifier

**Avoid high-cardinality labels**:
- ❌ `pod_name` - Changes with every deployment
- ❌ `container_id` - Unique per container
- ❌ `request_id` - Unique per request
- ❌ `timestamp` - Always unique

## Query Optimization

### Efficient Queries ✅
```logql
# Specific labels + line filter
{namespace="production", app="api"} |= "error"

# Structured parsing
{namespace="production"} | json | level="error"

# Metric aggregation
rate({namespace="production"}[5m])
```

### Inefficient Queries ❌
```logql
# No label selector (scans everything!)
{} |= "error"

# Regex on labels (very slow!)
{namespace=~"prod.*"}

# Large time range (expensive!)
{namespace="production"}[30d]
```

## Escalation Matrix

| Issue | First Response | Escalation Time | Escalate To |
|-------|---------------|-----------------|-------------|
| Complete outage | Restart components | 15 minutes | Platform team |
| Storage full | Increase PVC/cleanup | 30 minutes | Storage admin |
| Write path down | Check MinIO, restart | 15 minutes | Platform team |
| Query performance | Scale replicas | 1 hour | Performance team |
| Memory issues | Increase limits | 30 minutes | Capacity planning |

## Related Documentation

- [Loki Configuration](../../../flux/clusters/homelab/infrastructure/loki/helmrelease.yaml)
- [Architecture Overview](../../../ARCHITECTURE.md)
- [Alloy Configuration](../../../flux/clusters/homelab/infrastructure/alloy/helmrelease.yaml)
- [Grafana Official Docs](https://grafana.com/docs/loki/latest/)

## Monitoring & Alerts

Loki metrics are scraped by Prometheus and visualized in Grafana. Key dashboards:
- Loki / Overview
- Loki / Reads
- Loki / Writes
- Loki / Operational

Access Grafana at your configured URL and search for "Loki" dashboards.

## Support

For issues not covered by these runbooks:
1. Check Loki logs: `kubectl logs -n loki -l app.kubernetes.io/name=loki`
2. Review recent changes: `flux get helmreleases -n loki`
3. Consult [Grafana Loki documentation](https://grafana.com/docs/loki/latest/)
4. Check [GitHub issues](https://github.com/grafana/loki/issues)

---

**Last Updated**: 2025-10-15  
**Loki Version**: 6.25.0  
**Maintainer**: Homelab Platform Team

