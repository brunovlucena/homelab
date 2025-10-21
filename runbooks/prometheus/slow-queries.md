# 🚨 Runbook: Prometheus Slow Queries

## Alert Information

**Alert Name:** `PrometheusSlowQueries`  
**Severity:** Warning  
**Component:** prometheus  
**Service:** query-engine

## Symptom

Prometheus queries are taking longer than expected to execute, impacting dashboard load times and API responsiveness.

## Impact

- **User Impact:** MEDIUM - Slow dashboards, delayed alerts
- **Business Impact:** MEDIUM - Degraded observability experience
- **Data Impact:** NONE - No data loss

## Diagnosis

### 1. Check Query Duration

```promql
# 99th percentile query duration
histogram_quantile(0.99, rate(prometheus_http_request_duration_seconds_bucket{handler="/api/v1/query"}[5m]))

# Average query duration
rate(prometheus_http_request_duration_seconds_sum{handler="/api/v1/query"}[5m]) /
rate(prometheus_http_request_duration_seconds_count{handler="/api/v1/query"}[5m])

# Queries taking > 10 seconds
prometheus_http_request_duration_seconds_bucket{handler="/api/v1/query", le="10"} < 
prometheus_http_request_duration_seconds_count{handler="/api/v1/query"}
```

### 2. Check TSDB Performance

```promql
# Head chunks
prometheus_tsdb_head_chunks

# Number of series in head
prometheus_tsdb_head_series

# Queries in progress
prometheus_engine_queries

# Query samples processed
rate(prometheus_engine_query_samples_total[5m])
```

### 3. Check Resource Usage

```bash
# Check CPU usage
kubectl top pods -n prometheus -l app.kubernetes.io/name=prometheus

# Check memory usage
kubectl top pods -n prometheus -l app.kubernetes.io/name=prometheus

# Check disk I/O
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  iostat -x 1 5
```

### 4. Identify Slow Queries

```bash
# Enable query logging (if not already enabled)
kubectl edit prometheus -n prometheus prometheus-kube-prometheus-prometheus
# Add: enableFeatures: ["promql-at-modifier", "promql-negative-offset"]
# Add: queryLogFile: "/prometheus/queries.log"

# Check query logs
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  tail -100 /prometheus/queries.log

# Look for queries with high execution time
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  cat /prometheus/queries.log | jq 'select(.duration > 10)' | head -20
```

### 5. Check for High Cardinality

```promql
# Total series count
count({__name__=~".+"})

# Series per metric
topk(20, count by (__name__)({__name__=~".+"}))

# Churn rate (new series creation)
rate(prometheus_tsdb_head_series_created_total[5m])
```

## Resolution Steps

### Step 1: Identify problematic queries

```bash
# Check Prometheus logs for slow queries
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus | grep -i "slow query"

# Enable query logging if not enabled
# See diagnosis step 4
```

### Step 2: Common Issues and Fixes

#### Issue: High Cardinality Queries
**Cause:** Queries returning millions of time series  
**Fix:**
```promql
# Bad:  {__name__=~".*"}
# Good: {job="specific-job"}

# Bad:  rate(metric[5m])  # Returns all series
# Good: rate(metric{job="specific-job"}[5m])  # Filter first

# Bad:  count(metric) without (high_cardinality_label)
# Good: count(metric) by (low_cardinality_label)
```

#### Issue: Long Time Ranges
**Cause:** Queries spanning weeks or months of data  
**Fix:**
```promql
# Bad:  rate(metric[30d])  # Very expensive
# Good: rate(metric[5m])   # Use shorter ranges

# For long-term trends, use recording rules
# Or query from long-term storage (Thanos/Cortex)

# Limit dashboard time range
# Use "Last 1 hour" instead of "Last 30 days"
```

#### Issue: Subqueries Without Limits
**Cause:** Nested queries without step parameter  
**Fix:**
```promql
# Bad:  max_over_time(rate(metric[5m])[1h:])
# Good: max_over_time(rate(metric[5m])[1h:30s])  # Add step

# Step should be appropriate for resolution needed
# Smaller step = more computation
```

#### Issue: High-Resolution Queries
**Cause:** Requesting too many data points  
**Fix:**
```bash
# Reduce query resolution
# In Grafana: Settings -> Dashboard -> Time Options
# Set "Min interval" to appropriate value (e.g., 30s, 1m)

# For dashboards, use:
# - Step: 30s for recent data
# - Step: 5m for historical data
```

#### Issue: Regex Matchers
**Cause:** Inefficient regex in label matchers  
**Fix:**
```promql
# Bad:  {job=~".*api.*"}              # Scans all series
# Good: {job=~"api-.+"}               # More specific

# Bad:  {pod=~".+"}                   # Inefficient
# Good: {pod!=""}                     # Faster

# Bad:  {namespace=~"default|kube-.*"}  # Complex regex
# Good: {namespace="default"} or {namespace=~"kube-.+"}  # Separate or simplify
```

#### Issue: Aggregation Without Grouping
**Cause:** Aggregating without proper by/without clauses  
**Fix:**
```promql
# Bad:  sum(rate(metric[5m]))  # Aggregates everything
# Good: sum by (job) (rate(metric[5m]))  # Group by relevant labels

# Use 'without' to exclude high-cardinality labels
sum without (pod, instance) (rate(metric[5m]))
```

#### Issue: Multiple Label Joins
**Cause:** Joining metrics with complex label matching  
**Fix:**
```promql
# Optimize label joins using 'on' or 'ignoring'

# Bad:  metric1 / metric2  # May fail on label mismatch
# Good: metric1 / on(job, instance) metric2
# Good: metric1 / ignoring(pod) metric2

# Pre-aggregate if possible
sum by (job) (metric1) / sum by (job) (metric2)
```

#### Issue: Unnecessary Sorting
**Cause:** Using topk/bottomk with large k values  
**Fix:**
```promql
# Bad:  topk(1000, metric)  # Returns too many series
# Good: topk(10, metric)    # Limit to what you actually need

# Consider if you really need sorting
# Aggregation might be sufficient
```

### Step 3: Optimize with Recording Rules

```yaml
# Create recording rules for expensive queries
# Helm values:
additionalPrometheusRulesMap:
  recording-rules:
    groups:
      - name: expensive_queries
        interval: 30s
        rules:
          - record: job:http_requests_total:rate5m
            expr: sum by (job) (rate(http_requests_total[5m]))
          
          - record: job:http_request_duration_seconds:p95
            expr: histogram_quantile(0.95, sum by (job, le) (rate(http_request_duration_seconds_bucket[5m])))

# Use pre-calculated metrics in dashboards
# Instead of: rate(http_requests_total[5m])
# Use: job:http_requests_total:rate5m
```

### Step 4: Increase Query Resources

```bash
# If queries are legitimately slow due to load
# Increase CPU/memory limits
# Helm values:
# resources:
#   limits:
#     cpu: 4000m      # Increase from current
#     memory: 8Gi     # Increase from current
#   requests:
#     cpu: 2000m
#     memory: 4Gi

flux reconcile helmrelease kube-prometheus-stack -n prometheus
```

### Step 5: Configure Query Limits

```bash
# Set query timeout and max samples
kubectl edit prometheus -n prometheus prometheus-kube-prometheus-prometheus

# Add:
# spec:
#   query:
#     timeout: 2m          # Max query duration
#     maxSamples: 50000000 # Max samples per query
#     maxConcurrent: 20    # Max concurrent queries
```

### Step 6: Enable Query Stats

```bash
# Enable query statistics
kubectl edit prometheus -n prometheus prometheus-kube-prometheus-prometheus

# Add:
# spec:
#   enableFeatures:
#     - promql-at-modifier
#     - promql-negative-offset
#   queryLogFile: /prometheus/queries.log

# Restart to apply
kubectl delete pod -n prometheus -l app.kubernetes.io/name=prometheus
```

## Verification

1. Check query duration improved:
```promql
histogram_quantile(0.99, rate(prometheus_http_request_duration_seconds_bucket{handler="/api/v1/query"}[5m]))
```

2. Verify dashboards load faster:
```bash
# Time dashboard load
time curl -s "http://localhost:9090/api/v1/query?query=up" > /dev/null
```

3. Check query logs:
```bash
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  tail -50 /prometheus/queries.log
```

4. Monitor query samples:
```promql
rate(prometheus_engine_query_samples_total[5m])
```

5. Check resource usage:
```bash
kubectl top pods -n prometheus -l app.kubernetes.io/name=prometheus
```

## Prevention

1. Use recording rules for complex queries
2. Limit dashboard time ranges
3. Use appropriate step intervals
4. Avoid high-cardinality label operations
5. Regular query performance audits
6. Educate users on efficient PromQL
7. Set query timeouts and limits
8. Monitor query performance metrics
9. Use specific label matchers
10. Implement query result caching (Grafana)

## PromQL Best Practices

### Efficient Query Patterns

```promql
# ✅ GOOD: Specific label matchers
sum(rate(http_requests_total{job="api", status="200"}[5m]))

# ❌ BAD: Broad matchers
sum(rate(http_requests_total[5m]))

# ✅ GOOD: Aggregate early
sum by (job) (rate(metric[5m]))

# ❌ BAD: Aggregate late
sum(metric) by (job)

# ✅ GOOD: Use recording rules for dashboard queries
job:http_requests:rate5m

# ❌ BAD: Complex calculation in every dashboard
sum by (job) (rate(http_requests_total[5m])) / sum by (job) (rate(http_requests_duration_seconds_count[5m]))

# ✅ GOOD: Appropriate time ranges
rate(metric[5m])  # For rate calculations

# ❌ BAD: Excessive time ranges
rate(metric[1h])  # Usually unnecessary
```

### Query Optimization Checklist

- [ ] Use specific label matchers (job, namespace)
- [ ] Limit time range to necessary duration
- [ ] Specify step parameter for subqueries
- [ ] Use recording rules for repeated calculations
- [ ] Aggregate with 'by' clause to reduce series
- [ ] Avoid regex when exact match possible
- [ ] Use 'on' or 'ignoring' for label joins
- [ ] Limit topk/bottomk to small values
- [ ] Check query returns reasonable number of series
- [ ] Test query performance before adding to dashboard

## Related Alerts

- `PrometheusDown`
- `PrometheusHighMemoryUsage`
- `PrometheusHighCPUUsage`
- `PrometheusQueryTimeout`
- `PrometheusHighCardinality`

## Escalation

If the issue persists after following these steps:
1. Review all dashboard queries for efficiency
2. Consider Prometheus federation/sharding
3. Implement query frontend caching
4. Review metric retention and cardinality
5. Contact Prometheus expert or on-call engineer

## Additional Resources

- [PromQL Documentation](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Query Performance](https://prometheus.io/docs/prometheus/latest/querying/basics/#performance-considerations)
- [Recording Rules](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/)
- [PromQL Best Practices](https://prometheus.io/docs/practices/naming/)
- [Grafana Query Optimization](https://grafana.com/docs/grafana/latest/datasources/prometheus/#query-optimization)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

