# 📊 Knative Lambda Builder - Grafana Dashboard

## 🎯 Overview

This comprehensive Grafana dashboard provides complete visibility into the Knative Lambda Builder service, covering all critical metrics across 7 major categories:

## 📈 Dashboard Sections

### 1. ☁️ CloudEvent Metrics (4 metrics)
- **CloudEvents Processing Rate** - Total CloudEvents processed per second
- **CloudEvent Processing Duration (P95)** - 95th percentile processing time
- **Average CloudEvent Size** - Mean CloudEvent size in bytes
- **Average CloudEvent Response Size** - Mean response size in bytes

### 2. 🏗️ Build Process Metrics (6 metrics)
- **Build Requests Rate** - Total build requests per second
- **Build Request Duration (P95)** - 95th percentile build duration
- **Build Success Rate** - Successful builds per second
- **Build Failure Rate** - Failed builds per second
- **Build Queue Size** - Current builds in queue
- **Build Queue Duration (P95)** - 95th percentile queue wait time

### 3. ☸️ Kubernetes Job Metrics (5 metrics)
- **K8s Job Creation Rate** - Jobs created per second
- **K8s Job Creation Duration (P95)** - 95th percentile job creation time
- **K8s Job Success Rate** - Successful jobs per second
- **K8s Job Failure Rate** - Failed jobs per second
- **K8s Job Execution Duration (P95)** - 95th percentile job execution time

### 4. ☁️ AWS Metrics (5 metrics)
- **S3 Upload Rate** - S3 uploads per second
- **S3 Upload Duration (P95)** - 95th percentile upload time
- **Average S3 Upload Size** - Mean upload size in bytes
- **ECR Push Rate** - ECR pushes per second
- **ECR Push Duration (P95)** - 95th percentile push time

### 5. 💻 System Metrics (4 metrics)
- **System Memory Usage** - Memory usage in bytes
- **System CPU Usage** - CPU usage percentage
- **Number of Goroutines** - Active goroutines count
- **Heap Allocation** - Heap allocation in bytes

### 6. 🚨 Error Metrics (2 metrics)
- **Error Rate** - Total errors per second
- **Error Rate per Minute** - Error rate per minute

### 7. 📡 HTTP Metrics (4 metrics)
- **HTTP Requests Rate** - HTTP requests per second
- **HTTP Request Duration (P95)** - 95th percentile request time
- **Average HTTP Request Size** - Mean request size in bytes
- **Average HTTP Response Size** - Mean response size in bytes

## 🔧 Configuration

### Data Source
- **Variable**: `$prometheus_datasource`
- **Type**: Prometheus
- **Default**: Prometheus

### Time Range
- **Default**: Last 1 hour
- **Refresh**: Every 5 seconds
- **Timezone**: System default

## 📊 Metric Details

### CloudEvent Metrics
```promql
# Processing Rate
rate(cloudevents_total[5m])

# Duration P95
histogram_quantile(0.95, sum(rate(cloudevent_duration_seconds_bucket[5m])) by (le))

# Average Size
rate(cloudevent_size_bytes_sum[5m]) / rate(cloudevent_size_bytes_count[5m])

# Average Response Size
rate(cloudevent_response_size_bytes_sum[5m]) / rate(cloudevent_response_size_bytes_count[5m])
```

### Build Process Metrics
```promql
# Request Rate
rate(build_requests_total[5m])

# Duration P95
histogram_quantile(0.95, sum(rate(build_request_duration_seconds_bucket[5m])) by (le))

# Success Rate
rate(build_success_total[5m])

# Failure Rate
rate(build_failure_total[5m])

# Queue Size
build_queue_size

# Queue Duration P95
histogram_quantile(0.95, sum(rate(build_queue_duration_seconds_bucket[5m])) by (le))
```

### Kubernetes Job Metrics
```promql
# Creation Rate
rate(k8s_job_creation_total[5m])

# Creation Duration P95
histogram_quantile(0.95, sum(rate(k8s_job_creation_duration_seconds_bucket[5m])) by (le))

# Success Rate
rate(k8s_job_success_total[5m])

# Failure Rate
rate(k8s_job_failure_total[5m])

# Execution Duration P95
histogram_quantile(0.95, sum(rate(k8s_job_duration_seconds_bucket[5m])) by (le))
```

### AWS Metrics
```promql
# S3 Upload Rate
rate(aws_s3_upload_total[5m])

# S3 Upload Duration P95
histogram_quantile(0.95, sum(rate(aws_s3_upload_duration_seconds_bucket[5m])) by (le))

# S3 Upload Size
rate(aws_s3_upload_size_bytes_sum[5m]) / rate(aws_s3_upload_size_bytes_count[5m])

# ECR Push Rate
rate(aws_ecr_push_total[5m])

# ECR Push Duration P95
histogram_quantile(0.95, sum(rate(aws_ecr_push_duration_seconds_bucket[5m])) by (le))
```

### System Metrics
```promql
# Memory Usage
system_memory_usage_bytes

# CPU Usage
system_cpu_usage_percent

# Goroutines
system_goroutines

# Heap Allocation
system_heap_alloc_bytes
```

### Error Metrics
```promql
# Error Rate
rate(error_total[5m])

# Error Rate per Minute
error_rate
```

### HTTP Metrics
```promql
# Request Rate
rate(http_requests_total[5m])

# Request Duration P95
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))

# Request Size
rate(http_request_size_bytes_sum[5m]) / rate(http_request_size_bytes_count[5m])

# Response Size
rate(http_response_size_bytes_sum[5m]) / rate(http_response_size_bytes_count[5m])
```

## 🚨 Alert Thresholds

### Golden Signals
- **Error Rate**: >5% (warning), >20% (critical)
- **Latency (P95)**: >30s (warning), >120s (critical)
- **CPU Usage**: >80% (warning), >90% (critical)
- **Memory Usage**: >85% (warning), >95% (critical)

### Business Metrics
- **Build Success Rate**: <99% (warning), <95% (critical)
- **Build Duration (P95)**: >30m (warning), >60m (critical)
- **Queue Depth**: >50 (warning), >100 (critical)
- **Job Failure Rate**: >10% (warning), >20% (critical)

## 📁 Files

- **`knative-lambda-comprehensive.json`** - Main comprehensive dashboard
- **`knative-lambda-cloudevent.json`** - CloudEvent-focused dashboard
- **`README.md`** - This documentation file

## 🔍 Usage

1. **Import Dashboard**: Import the JSON file into Grafana
2. **Configure Data Source**: Set the `$prometheus_datasource` variable
3. **Customize Time Range**: Adjust based on monitoring needs
4. **Set Alerts**: Configure alerting rules based on thresholds

## 🎯 SLO/SLI Targets

- **Availability**: 99.9% uptime
- **Error Rate**: <0.1% error rate
- **Latency**: P95 <30s for build requests
- **Throughput**: >100 builds/hour

## 🔧 Troubleshooting

### Common Issues
1. **No Data**: Verify Prometheus data source is configured correctly
2. **Missing Metrics**: Ensure all metrics are exposed by the service
3. **High Latency**: Check system resources and queue depth
4. **High Error Rate**: Review logs and error patterns

### Verification Steps
1. Check `/metrics` endpoint returns data
2. Verify Prometheus is scraping the service
3. Confirm metric names match exactly
4. Check time range and refresh settings

## 📚 References

- [Prometheus Query Language](https://prometheus.io/docs/prometheus/latest/querying/)
- [Grafana Dashboard Documentation](https://grafana.com/docs/grafana/latest/dashboards/)
- [Knative Lambda Builder Metrics](../METRICS.md)
- [Service Architecture](../INTRO.md)
