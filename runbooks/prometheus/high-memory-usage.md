# 🚨 Runbook: Prometheus High Memory Usage

## Alert Information

**Alert Name:** `PrometheusHighMemoryUsage`  
**Severity:** Warning  
**Component:** prometheus  
**Service:** prometheus-server

## Symptom

Prometheus memory usage has exceeded 80% of the configured limit for more than 10 minutes.

## Impact

- **User Impact:** LOW - Service still operational
- **Business Impact:** MEDIUM - Risk of OOMKill and metrics loss
- **Data Impact:** HIGH - Potential metrics gap if OOMKilled

## Diagnosis

### 1. Check Current Memory Usage

```bash
# Check pod memory usage
kubectl top pods -n prometheus -l app.kubernetes.io/name=prometheus

# Check memory limits
kubectl get pods -n prometheus -l app.kubernetes.io/name=prometheus \
  -o jsonpath='{.items[0].spec.containers[0].resources}'
```

### 2. Check Memory Usage Over Time

```promql
# Current memory usage
container_memory_usage_bytes{pod=~"prometheus-prometheus-kube-prometheus-prometheus-.*", namespace="prometheus"}

# Memory usage percentage
100 * container_memory_usage_bytes{pod=~"prometheus-prometheus-kube-prometheus-prometheus-.*", namespace="prometheus"} 
/ 
container_spec_memory_limit_bytes{pod=~"prometheus-prometheus-kube-prometheus-prometheus-.*", namespace="prometheus"}

# Memory usage trend (last 24h)
rate(container_memory_usage_bytes{pod=~"prometheus-prometheus-kube-prometheus-prometheus-.*", namespace="prometheus"}[24h])
```

### 3. Check Prometheus TSDB Stats

```bash
# Access Prometheus UI
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-prometheus 9090:9090

# Open http://localhost:9090/tsdb-status
# Check:
# - Number of series in head
# - Number of chunks in head
# - Head chunk size
```

### 4. Check Series Cardinality

```promql
# Total number of series
prometheus_tsdb_symbol_table_size_bytes / 1024 / 1024

# Series count
count({__name__=~".+"})

# Top 10 metrics by cardinality
topk(10, count by (__name__)({__name__=~".+"}))

# Series per job
count by (job)({__name__=~".+"})
```

### 5. Check Scrape Targets and Metrics

```promql
# Number of active targets
count(up)

# Samples scraped per target
sum by (job) (scrape_samples_scraped)

# Top targets by sample count
topk(10, scrape_samples_scraped)
```

## Resolution Steps

### Step 1: Identify the cause

```bash
# Check if memory is steadily increasing
kubectl top pods -n prometheus -l app.kubernetes.io/name=prometheus

# Watch for a few minutes
watch -n 30 kubectl top pods -n prometheus -l app.kubernetes.io/name=prometheus
```

### Step 2: Analyze High Cardinality Metrics

```promql
# Find metrics with high cardinality
topk(20, count by (__name__)({__name__=~".+"}))

# Check labels with high cardinality
topk(20, count by (__name__, job)({__name__=~".+"}))

# Find specific high-cardinality metrics
count({__name__=~"<metric-name>"}) by (job, instance)
```

### Step 3: Common Issues and Fixes

#### Issue: High Series Cardinality
**Cause:** Too many unique time series (high cardinality labels)  
**Fix:**
```bash
# Identify high-cardinality metrics
# Use queries from Step 2

# Options:
# 1. Drop high-cardinality metrics
# Edit Prometheus config to add metric_relabel_configs

# 2. Reduce label cardinality in exporters
# Configure exporters to reduce labels

# 3. Use recording rules to aggregate high-cardinality metrics
kubectl edit prometheusrule -n prometheus
```

#### Issue: Too Many Targets
**Cause:** Scraping too many endpoints  
**Fix:**
```bash
# Count targets
# In Prometheus UI -> Status -> Targets

# Reduce scrape frequency for non-critical targets
kubectl edit servicemonitor -n <namespace> <servicemonitor-name>
# Increase scrapeInterval: from 30s to 60s or higher

# Disable unnecessary ServiceMonitors
kubectl scale deployment <exporter-deployment> --replicas=0 -n <namespace>
# Or delete ServiceMonitor
```

#### Issue: Long Retention Period
**Cause:** Storing too much historical data  
**Fix:**
```bash
# Check current retention
kubectl get prometheus -n prometheus prometheus-kube-prometheus-prometheus \
  -o jsonpath='{.spec.retention}'

# Reduce retention via Helm values
# retention: 7d  # Reduce from current (e.g., 30d -> 7d)

# Apply changes
flux reconcile helmrelease kube-prometheus-stack -n prometheus
```

#### Issue: Memory Leak
**Cause:** Prometheus bug or configuration issue  
**Fix:**
```bash
# Restart Prometheus
kubectl delete pod -n prometheus -l app.kubernetes.io/name=prometheus

# Monitor memory after restart
watch -n 30 kubectl top pods -n prometheus -l app.kubernetes.io/name=prometheus

# If leak continues, check Prometheus version and upgrade
kubectl get prometheus -n prometheus prometheus-kube-prometheus-prometheus \
  -o jsonpath='{.spec.version}'
```

#### Issue: WAL Size Growing
**Cause:** Write-Ahead Log not being truncated  
**Fix:**
```bash
# Check WAL size
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  du -sh /prometheus/wal

# Force WAL truncation (by restarting)
kubectl delete pod -n prometheus -l app.kubernetes.io/name=prometheus
```

### Step 4: Increase Memory Limits (If Necessary)

```bash
# Edit Prometheus resource limits via Helm values
# resources:
#   limits:
#     memory: 6Gi  # Increase from current
#   requests:
#     memory: 4Gi  # Increase from current

# Apply changes
flux reconcile helmrelease kube-prometheus-stack -n prometheus

# Verify new limits
kubectl get pods -n prometheus -l app.kubernetes.io/name=prometheus \
  -o jsonpath='{.items[0].spec.containers[0].resources}'
```

### Step 5: Implement Metric Dropping/Relabeling

```yaml
# Add to Prometheus config via Helm values
additionalScrapeConfigs:
  - job_name: 'high-cardinality-job'
    metric_relabel_configs:
      # Drop specific high-cardinality metrics
      - source_labels: [__name__]
        regex: 'expensive_metric_.*'
        action: drop
      
      # Drop specific labels
      - source_labels: [high_cardinality_label]
        action: labeldrop
        regex: high_cardinality_label
      
      # Keep only specific metrics
      - source_labels: [__name__]
        regex: '(important_metric_1|important_metric_2)'
        action: keep
```

### Step 6: Enable Remote Write (Long-term Solution)

```yaml
# Configure remote write to offload data
remoteWrite:
  - url: "http://thanos-receive:19291/api/v1/receive"
    queueConfig:
      capacity: 10000
      maxShards: 50
      minShards: 1
```

## Verification

1. Check memory usage decreased:
```bash
kubectl top pods -n prometheus -l app.kubernetes.io/name=prometheus
```

2. Verify series count reduced (if that was the fix):
```promql
count({__name__=~".+"})
```

3. Check TSDB stats:
```bash
# Open http://localhost:9090/tsdb-status
# Verify head series count is reasonable
```

4. Monitor memory trend:
```promql
container_memory_usage_bytes{pod=~"prometheus-prometheus-kube-prometheus-prometheus-.*", namespace="prometheus"}
```

5. Verify all important metrics still being collected:
```bash
# Test critical queries in Prometheus UI
```

## Prevention

1. Set up recording rules for high-cardinality metrics
2. Monitor series count and set alerts
3. Regular cardinality audits
4. Configure appropriate retention period
5. Use remote write for long-term storage
6. Implement metric dropping for unnecessary metrics
7. Set proper memory requests and limits
8. Monitor memory usage trends
9. Educate teams about metric cardinality best practices
10. Use relabeling to reduce label cardinality

## Related Alerts

- `PrometheusDown`
- `PrometheusHighCardinality`
- `PrometheusTSDBCompactionsFailing`
- `PrometheusStorageFull`
- `PrometheusOOMKilled`

## Escalation

If the issue persists after following these steps:
1. Review all ServiceMonitors for inefficient scrape configs
2. Audit all exporters for high-cardinality metrics
3. Consider horizontal scaling (Prometheus federation/sharding)
4. Review Prometheus version for known memory issues
5. Contact Prometheus expert or on-call engineer

## Additional Resources

- [Prometheus Memory Tuning](https://prometheus.io/docs/prometheus/latest/storage/#memory-usage)
- [High Cardinality Guide](https://www.robustperception.io/cardinality-is-key)
- [Metric Relabeling](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config)
- [TSDB Format](https://prometheus.io/docs/prometheus/latest/storage/)
- [Remote Write](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

