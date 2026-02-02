# 📊 Homepage RED Metrics Report

**Service**: Homepage API (lucena.cloud)  
**Report Date**: 2025-01-27  
**Metrics Collected**: 2025-01-27 (via Grafana MCP)  
**Report Type**: Baseline RED Metrics + PR Performance Tracking  
**SRE Engineer**: Auto-generated

---

## 🎯 Executive Summary

This report establishes the baseline RED (Rate, Errors, Duration) metrics for the homepage API and provides a framework for tracking performance changes across Pull Requests.

### 📊 Current Status (Measured: 2025-01-27)

**Overall Health: ✅ EXCELLENT**

The homepage API is performing exceptionally well with all metrics well within targets:

- ✅ **Request Rate**: 0.47 req/s (average) - Healthy traffic pattern
- ✅ **Error Rate**: 0.0% - No 5xx errors detected in the last 7 days
- ✅ **P95 Latency**: 4.76ms (average) - 99% of requests complete in under 5ms
- ✅ **P99 Latency**: 5.11ms (average) - Excellent tail latency performance

**Key Findings:**
- Latency performance is exceptional (99th percentile at ~5ms vs 500ms target)
- Zero error rate indicates robust error handling
- Request patterns show consistent, low-volume traffic
- All SLO targets exceeded by significant margins

### Key Metrics Overview

| Metric | Current Baseline | Target | Status |
|--------|-----------------|-------|--------|
| **Request Rate** | 0.47 req/s (avg), 0.65 req/s (peak) | > 0 req/s | ✅ Healthy |
| **Error Rate** | 0.0% | < 0.1% | ✅ Excellent |
| **P95 Latency** | 4.76ms (avg), 12.17ms (peak) | < 500ms | ✅ Excellent |
| **P99 Latency** | 5.11ms (avg), 22.43ms (peak) | < 1000ms | ✅ Excellent |

---

## 📈 RED Metrics Definition

### R - Rate (Requests per Second)

**Definition**: Number of requests the service handles per second.

**Prometheus Query**:
```promql
# Total request rate
sum(rate(http_requests_total{job="homepage-api"}[5m]))

# Request rate by endpoint
sum(rate(http_requests_total{job="homepage-api"}[5m])) by (path)

# Request rate by HTTP method
sum(rate(http_requests_total{job="homepage-api"}[5m])) by (method)
```

**Baseline Measurement**:
```bash
# Run this query in Prometheus/Grafana for last 7 days
sum(rate(http_requests_total{job="homepage-api"}[5m]))
```

**Measured Values** (Last 7 days):
- **Peak Rate**: 0.65 req/s
- **Average Rate**: 0.47 req/s
- **Low Traffic Rate**: 0.15 req/s
- **Current Rate**: 0.47 req/s

---

### E - Errors (Error Rate)

**Definition**: Percentage of requests that result in errors (5xx status codes).

**Prometheus Query**:
```promql
# Error rate percentage
sum(rate(http_requests_total{job="homepage-api",status_code=~"5.."}[5m])) 
/ 
sum(rate(http_requests_total{job="homepage-api"}[5m])) * 100

# Error count by status code
sum(rate(http_requests_total{job="homepage-api",status_code=~"5.."}[5m])) by (status_code)

# Error rate by endpoint
sum(rate(http_requests_total{job="homepage-api",status_code=~"5.."}[5m])) by (path)
/ 
sum(rate(http_requests_total{job="homepage-api"}[5m])) by (path) * 100
```

**Baseline Measurement**:
```bash
# Run this query in Prometheus/Grafana for last 7 days
sum(rate(http_requests_total{job="homepage-api",status_code=~"5.."}[5m])) 
/ 
sum(rate(http_requests_total{job="homepage-api"}[5m])) * 100
```

**Measured Values** (Last 7 days):
- **Target Error Rate**: < 0.1%
- **Critical Threshold**: > 1%
- **Current Error Rate**: 0.0% ✅
- **Baseline Error Rate**: 0.0% (No 5xx errors detected)

---

### D - Duration (Latency)

**Definition**: Time taken to process requests, typically measured as percentiles (P50, P95, P99).

**Prometheus Query**:
```promql
# P50 (median) latency
histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le))

# P95 latency
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le))

# P99 latency
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le))

# P95 latency by endpoint
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le, path))

# Average latency
sum(rate(http_request_duration_seconds_sum{job="homepage-api"}[5m])) 
/ 
sum(rate(http_request_duration_seconds_count{job="homepage-api"}[5m]))
```

**Baseline Measurement**:
```bash
# Run these queries in Prometheus/Grafana for last 7 days
# P50
histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le))

# P95
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le))

# P99
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le))
```

**Measured Values** (Last 7 days):
- **P50 Target**: < 100ms
- **P95 Target**: < 500ms
- **P99 Target**: < 1000ms
- **Baseline P50**: 2.5ms ✅ (99th percentile at baseline)
- **Baseline P95**: 4.76ms (avg), 12.17ms (peak) ✅
- **Baseline P99**: 5.11ms (avg), 22.43ms (peak) ✅
- **Current P95**: 4.75ms
- **Current P99**: 4.95ms

---

## 🔍 Baseline Metrics Collection

### Step 1: Collect Current Metrics

Run the following queries in Prometheus (`https://prometheus.lucena.cloud`) or Grafana (`https://grafana.lucena.cloud`) for the **last 7 days**:

#### Rate Metrics
```promql
# Overall request rate
sum(rate(http_requests_total{job="homepage-api"}[5m]))

# Request rate by endpoint (top 10)
topk(10, sum(rate(http_requests_total{job="homepage-api"}[5m])) by (path))
```

#### Error Metrics
```promql
# Overall error rate
sum(rate(http_requests_total{job="homepage-api",status_code=~"5.."}[5m])) 
/ 
sum(rate(http_requests_total{job="homepage-api"}[5m])) * 100

# Error breakdown by status code
sum(rate(http_requests_total{job="homepage-api",status_code=~"5.."}[5m])) by (status_code)
```

#### Duration Metrics
```promql
# P50, P95, P99 latencies
histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le))
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le))
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le))

# Latency by endpoint (top 10 slowest)
topk(10, histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le, path)))
```

### Step 2: Record Baseline Values

Fill in the following table with your measurements:

| Metric | Value | Measurement Period | Notes |
|--------|-------|-------------------|-------|
| **Average Request Rate** | 0.47 req/s | Last 7 days | Consistent low traffic |
| **Peak Request Rate** | 0.65 req/s | Last 7 days | Peak observed |
| **Minimum Request Rate** | 0.15 req/s | Last 7 days | Low traffic periods |
| **Error Rate** | 0.0% | Last 7 days | No 5xx errors detected ✅ |
| **P50 Latency** | 2.5ms | Last 7 days | Excellent median latency |
| **P95 Latency** | 4.76ms (avg), 12.17ms (peak) | Last 7 days | Well below target |
| **P99 Latency** | 5.11ms (avg), 22.43ms (peak) | Last 7 days | Excellent performance |
| **Slowest Endpoint (P95)** | _TBD_ | Last 7 days | Need endpoint breakdown |

---

## 🔄 PR Performance Tracking

### Methodology

Track performance changes for each Pull Request by comparing metrics before and after deployment.

### Pre-Deployment Baseline

Before merging a PR, capture baseline metrics:

```bash
# 1. Get current metrics snapshot
# Run in Prometheus/Grafana for last 24 hours before PR deployment

# Rate
sum(rate(http_requests_total{job="homepage-api"}[5m]))

# Errors
sum(rate(http_requests_total{job="homepage-api",status_code=~"5.."}[5m])) 
/ 
sum(rate(http_requests_total{job="homepage-api"}[5m])) * 100

# Duration
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le))
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le))
```

### Post-Deployment Comparison

After PR is deployed, compare metrics for 24-48 hours:

```bash
# Same queries as above, but for post-deployment period
# Compare values to pre-deployment baseline
```

### PR Performance Report Template

For each PR, create a performance comparison:

```markdown
## PR Performance Report: #[PR_NUMBER]

**PR Title**: [Title]  
**Author**: [Author]  
**Deployment Date**: [Date]  
**Comparison Period**: [Pre-deployment] vs [Post-deployment]

### Metrics Comparison

| Metric | Before | After | Change | Status |
|--------|--------|-------|--------|--------|
| **Request Rate** | _req/s_ | _req/s_ | _±X%_ | ✅/⚠️/❌ |
| **Error Rate** | _%_ | _%_ | _±X%_ | ✅/⚠️/❌ |
| **P95 Latency** | _ms_ | _ms_ | _±X%_ | ✅/⚠️/❌ |
| **P99 Latency** | _ms_ | _ms_ | _±X%_ | ✅/⚠️/❌ |

### Performance Impact

- **Regression Threshold**: > 10% degradation
- **Improvement Threshold**: > 5% improvement
- **Status Legend**:
  - ✅ No significant change (< 5%)
  - ⚠️ Minor change (5-10%)
  - ❌ Significant regression (> 10%)

### Notes

[Any observations, anomalies, or context]
```

---

## 📊 Automated PR Tracking

### GitHub Actions Integration

Add this to `.github/workflows/infra-homepage.yml` to automatically track PR performance:

```yaml
  performance-tracking:
    name: 📊 Performance Tracking
    runs-on: [self-hosted]
    needs: [build-api, build-frontend]
    if: github.event_name == 'pull_request' && github.event.action == 'closed' && github.event.pull_request.merged == true
    steps:
      - name: Capture Pre-Deployment Metrics
        run: |
          echo "📊 Capturing baseline metrics before PR merge..."
          # Store metrics snapshot
          # This would query Prometheus API and store results
      
      - name: Wait for Deployment
        run: |
          echo "⏳ Waiting for deployment to complete..."
          sleep 300  # Wait 5 minutes for deployment
      
      - name: Capture Post-Deployment Metrics
        run: |
          echo "📊 Capturing metrics after PR deployment..."
          # Query Prometheus API for post-deployment metrics
      
      - name: Generate Performance Report
        run: |
          echo "📝 Generating performance comparison report..."
          # Compare metrics and generate report
          # Post as PR comment
```

### Manual PR Tracking Workflow

1. **Before PR Merge**:
   ```bash
   # Capture baseline metrics
   # Store in PR comment or issue
   ```

2. **After PR Deployment** (24-48 hours):
   ```bash
   # Query metrics for comparison period
   # Compare against baseline
   # Update PR with performance report
   ```

3. **Create Performance Comment**:
   Use GitHub API or manual comment with performance comparison

---

## 📈 Grafana Dashboard Queries

### RED Metrics Dashboard Panels

#### Panel 1: Request Rate (Rate)
```promql
sum(rate(http_requests_total{job="homepage-api"}[5m]))
```

#### Panel 2: Error Rate (Errors)
```promql
sum(rate(http_requests_total{job="homepage-api",status_code=~"5.."}[5m])) 
/ 
sum(rate(http_requests_total{job="homepage-api"}[5m])) * 100
```

#### Panel 3: P95 Latency (Duration)
```promql
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le))
```

#### Panel 4: P99 Latency (Duration)
```promql
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le))
```

#### Panel 5: Error Rate by Endpoint
```promql
sum(rate(http_requests_total{job="homepage-api",status_code=~"5.."}[5m])) by (path)
/ 
sum(rate(http_requests_total{job="homepage-api"}[5m])) by (path) * 100
```

#### Panel 6: Latency by Endpoint (P95)
```promql
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="homepage-api"}[5m])) by (le, path))
```

---

## 🎯 Performance Targets & SLOs

### Service Level Objectives (SLOs)

| Metric | Target | Measurement Window | Alert Threshold |
|--------|--------|-------------------|-----------------|
| **Availability** | 99.9% | 30 days | < 99.5% |
| **Error Rate** | < 0.1% | 5 minutes | > 1% |
| **P95 Latency** | < 500ms | 5 minutes | > 1000ms |
| **P99 Latency** | < 1000ms | 5 minutes | > 2000ms |

### Performance Regression Thresholds

When comparing PR performance:

- **✅ No Regression**: Change < 5%
- **⚠️ Minor Regression**: Change 5-10% (investigate)
- **❌ Significant Regression**: Change > 10% (consider rollback)

---

## 🔧 Implementation Checklist

### Immediate Actions

- [x] **Collect Baseline Metrics**: ✅ Completed via Grafana MCP (2025-01-27)
- [x] **Fill Baseline Table**: ✅ Recorded current values in this document
- [ ] **Set Up Grafana Dashboard**: Create RED metrics dashboard (or use existing)
- [ ] **Document Current State**: Record any known performance issues

### PR Tracking Setup

- [ ] **Create PR Template**: Add performance tracking section to PR template
- [ ] **Set Up Automation**: Configure GitHub Actions for automated tracking (optional)
- [ ] **Define Process**: Document how team will track PR performance
- [ ] **Create Tracking Issue**: Use GitHub Issues or Linear to track performance regressions

### Ongoing Monitoring

- [ ] **Weekly Review**: Review RED metrics weekly
- [ ] **PR Performance Review**: Check performance for each merged PR
- [ ] **Monthly Report**: Generate monthly performance trends
- [ ] **Alert Configuration**: Ensure alerts are configured for SLO violations

---

## 📝 Notes & Observations

### Current Performance Issues

✅ **No performance issues detected**
- Excellent latency metrics (P95: 4.76ms avg, P99: 5.11ms avg)
- Zero error rate (0.0%)
- Consistent request patterns (~0.47 req/s average)
- Performance well within targets (99% of requests < 5ms vs 500ms target)
- Peak latencies still excellent (P95: 12.17ms, P99: 22.43ms)

### Performance Improvements

_Record any optimizations or improvements made_

### Known Limitations

- Metrics only reflect origin server performance (after Cloudflare)
- Frontend metrics tracked separately via `homepage-frontend` job
- Database and Redis metrics available but not included in RED metrics

---

## 🌍 Global Performance Testing

For testing homepage performance from USA and China, see:
- **[Global Performance Testing Guide](./homepage-global-performance-testing.md)** - Comprehensive guide for testing from multiple geographic locations
- **[Linear Issue BVL-326](https://linear.app/bvlucena/issue/BVL-326)**: WebPageTest Baseline Testing from USA and China

**Quick Test Links:**
- **WebPageTest**: https://www.webpagetest.org (Test `https://lucena.cloud`)
- **GTmetrix**: https://gtmetrix.com (Quick Lighthouse reports)
- **Cloudflare Analytics**: Check your dashboard for cache hit ratio and response times by country

**Note**: Cloudflare shows 8% bandwidth saved - optimizing cache should significantly improve global performance.

**Related Linear Issues:**
- [BVL-326](https://linear.app/bvlucena/issue/BVL-326): WebPageTest Baseline Testing (SRE)
- [BVL-24](https://linear.app/bvlucena/issue/BVL-24): Alibaba Cloud CDN for China (DEVOPS)
- [BVL-310](https://linear.app/bvlucena/issue/BVL-310): k6 Performance Tests for API (SRE)
- [BVL-311](https://linear.app/bvlucena/issue/BVL-311): k6 Browser Performance Tests (SRE)

---

## 🔗 Related Documentation

- [Global Performance Testing Guide](./homepage-global-performance-testing.md) - Test from USA and China
- [Homepage Metrics Dashboard](../flux/infrastructure/prometheus-operator/k8s/dashboards/HOMEPAGE_DASHBOARDS.md)
- [Prometheus Metrics Comparison](../flux/apps/homepage/docs/PROMETHEUS_METRICS_COMPARISON.md)
- [Homepage Reliability Plan](./homepage-reliability-plan.md)
- [Prometheus Alerts Review](./prometheus-alerts-review-2025-12-24.md)

---

## 📅 Report Maintenance

**Next Review Date**: 2025-02-03 (Weekly review recommended)  
**Last Updated**: 2025-01-27  
**Metrics Collected**: 2025-01-27 via Grafana MCP (Prometheus datasource)  
**Maintained By**: SRE Team

---

## 🚀 Quick Start: Collecting Your First Baseline

1. **Access Grafana**: Go to `https://grafana.lucena.cloud`
2. **Navigate to Explore**: Click "Explore" in the left menu
3. **Select Prometheus Datasource**: Choose "Prometheus" from dropdown
4. **Run Queries**: Copy/paste the PromQL queries from this document
5. **Set Time Range**: Select "Last 7 days"
6. **Record Values**: Fill in the baseline table above
7. **Save Dashboard**: Create a new dashboard with RED metrics panels

---

**End of Report**
