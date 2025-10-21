# 📊 Knative Lambda Builder Metrics

## 🎯 Overview

This document provides a comprehensive list of all metrics exposed by the Knative Lambda Builder service, organized by the **Four Golden Signals** methodology and additional business metrics.

## 🏆 Four Golden Signals

The **Four Golden Signals** are the key metrics that indicate the health and performance of our service:

### 1. ✅ **Availability (Service Health)**
**Metric**: `up`
- **Description**: Service availability status
- **Type**: Gauge
- **Labels**: `namespace`, `job`, `service`
- **Alert**: `KnativeLambdaBuilderDown` - Service down for >2m

### 2. ❌ **Error Rate**
**Metric**: `cloudevents_total`
- **Description**: CloudEvent processing error rate percentage
- **Calculation**: `(5xx_events / total_events) * 100`
- **Type**: Counter
- **Labels**: `method`, `endpoint`, `status_code`, `handler`
- **Alerts**:
  - `KnativeLambdaBuilderHighErrorRate` - >5% error rate (warning)
  - `KnativeLambdaBuilderCriticalErrorRate` - >20% error rate (critical)

### 3. ⏱️ **Latency**
**Metric**: `cloudevent_duration_seconds`
- **Description**: CloudEvent processing duration in seconds
- **Type**: Histogram
- **Labels**: `method`, `endpoint`, `handler`
- **Buckets**: Default Prometheus buckets
- **Alerts**:
  - `KnativeLambdaBuilderHighLatency` - P95 >30s (warning)
  - `KnativeLambdaBuilderCriticalLatency` - P95 >120s (critical)

### 4. 💾 **Saturation**
**Metrics**:
- **CPU**: `container_cpu_usage_seconds_total`
- **Memory**: `container_memory_usage_bytes`
- **Type**: Gauge
- **Labels**: `namespace`, `pod`
- **Alerts**:
  - `KnativeLambdaBuilderHighCPUUsage` - CPU >80%
  - `KnativeLambdaBuilderHighMemoryUsage` - Memory >85%

## ☁️ CloudEvent Metrics

### CloudEvent Metrics
| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `cloudevents_total` | Counter | `method`, `endpoint`, `status_code`, `handler` | Total CloudEvents processed |
| `cloudevent_duration_seconds` | Histogram | `method`, `endpoint`, `handler` | CloudEvent processing duration |
| `cloudevent_size_bytes` | Histogram | `method`, `endpoint` | CloudEvent size in bytes |
| `cloudevent_response_size_bytes` | Histogram | `method`, `endpoint`, `status_code` | CloudEvent response size in bytes |

### Sample Queries
```promql
# CloudEvent processing rate by endpoint
rate(cloudevents_total[5m])

# Error rate percentage
(rate(cloudevents_total{status=~"5.."}[5m]) / rate(cloudevents_total[5m])) * 100

# P95 latency by endpoint
histogram_quantile(0.95, sum(rate(cloudevent_duration_seconds_bucket[5m])) by (le, endpoint))
```

## 🏗️ Build Process Metrics

### Build Request Metrics
| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `build_requests_total` | Counter | `third_party_id`, `parser_id`, `status` | Total build requests |
| `build_request_duration_seconds` | Histogram | `third_party_id`, `parser_id` | Build request duration |
| `build_success_total` | Counter | `third_party_id`, `parser_id` | Successful builds |
| `build_failure_total` | Counter | `third_party_id`, `parser_id`, `error_type` | Failed builds |

### Queue Metrics
| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `build_queue_size` | Gauge | `priority` | Current builds in queue |
| `build_queue_duration_seconds` | Histogram | `priority` | Time builds spend in queue |

### Sample Queries
```promql
# Build success rate
rate(build_success_total[5m]) / (rate(build_success_total[5m]) + rate(build_failure_total[5m])) * 100

# Average build duration
rate(build_request_duration_seconds_sum[5m]) / rate(build_request_duration_seconds_count[5m])

# Queue depth
build_queue_size
```

## ☸️ Kubernetes Metrics

### Job Metrics
| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `k8s_job_creation_total` | Counter | `job_type`, `status` | Total jobs created |
| `k8s_job_creation_duration_seconds` | Histogram | `job_type` | Job creation duration |
| `k8s_job_success_total` | Counter | `job_type` | Successful jobs |
| `k8s_job_failure_total` | Counter | `job_type`, `error_type` | Failed jobs |
| `k8s_job_duration_seconds` | Histogram | `job_type`, `status` | Job execution duration |

### Sample Queries
```promql
# Job success rate
rate(k8s_job_success_total[5m]) / (rate(k8s_job_success_total[5m]) + rate(k8s_job_failure_total[5m])) * 100

# Average job duration
rate(k8s_job_duration_seconds_sum[5m]) / rate(k8s_job_duration_seconds_count[5m])
```

## ☁️ AWS Metrics

### S3 Metrics
| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `aws_s3_upload_total` | Counter | `bucket`, `status` | Total S3 uploads |
| `aws_s3_upload_duration_seconds` | Histogram | `bucket` | S3 upload duration |
| `aws_s3_upload_size_bytes` | Histogram | `bucket` | S3 upload size |

### ECR Metrics
| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `aws_ecr_push_total` | Counter | `repository`, `status` | Total ECR pushes |
| `aws_ecr_push_duration_seconds` | Histogram | `repository` | ECR push duration |

### Sample Queries
```promql
# S3 upload success rate
rate(aws_s3_upload_total{status="success"}[5m]) / rate(aws_s3_upload_total[5m]) * 100

# Average ECR push duration
rate(aws_ecr_push_duration_seconds_sum[5m]) / rate(aws_ecr_push_duration_seconds_count[5m])
```

## 💻 System Metrics

### Resource Usage
| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `system_memory_usage_bytes` | Gauge | `type` | Memory usage in bytes |
| `system_cpu_usage_percent` | Gauge | `type` | CPU usage percentage |
| `system_goroutines` | Gauge | None | Number of goroutines |
| `system_heap_alloc_bytes` | Gauge | None | Heap allocation in bytes |

### Sample Queries
```promql
# Memory usage percentage
system_memory_usage_bytes / container_spec_memory_limit_bytes * 100

# CPU usage percentage
rate(container_cpu_usage_seconds_total[5m]) * 100

# Goroutine count
system_goroutines
```

## 🚨 Error Metrics

### Error Tracking
| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `error_total` | Counter | `component`, `error_type`, `severity` | Total errors |
| `error_rate` | Gauge | `component` | Error rate per minute |

### Sample Queries
```promql
# Error rate by component
rate(error_total[5m])

# Error rate percentage
error_rate
```

## 📈 Business Metrics

### Cost Metrics
| Metric | Description | Calculation |
|--------|-------------|-------------|
| `build_cost_per_hour` | Estimated hourly build cost | Based on resource usage and AWS pricing |
| `build_throughput` | Builds per hour | `rate(build_requests_total[1h])` |

### SLA Metrics
| Metric | Description | Target |
|--------|-------------|--------|
| `build_success_rate` | Percentage of successful builds | >99% |
| `build_duration_p95` | 95th percentile build duration | <30 minutes |
| `queue_depth` | Number of builds waiting | <50 |

## 🔍 Prometheus Queries

### Golden Signals Queries
```promql
# 1. Availability
up{namespace="knative-lambda-prd", job=~".*builder.*"}

# 2. Error Rate
(rate(cloudevents_total{status=~"5.."}[5m]) / rate(cloudevents_total[5m])) * 100

# 3. Latency (P95)
histogram_quantile(0.95, sum(rate(cloudevent_duration_seconds_bucket[5m])) by (le))

# 4. Saturation (CPU)
rate(container_cpu_usage_seconds_total[5m]) / container_spec_cpu_quota * 100

# 4. Saturation (Memory)
container_memory_usage_bytes / container_spec_memory_limit_bytes * 100
```

### Business Metrics Queries
```promql
# Build Success Rate
rate(build_success_total[5m]) / (rate(build_success_total[5m]) + rate(build_failure_total[5m])) * 100

# Build Throughput
rate(build_requests_total[5m])

# Queue Depth
build_queue_size

# Average Build Duration
rate(build_request_duration_seconds_sum[5m]) / rate(build_request_duration_seconds_count[5m])
```

## 📊 Grafana Dashboard Variables

### Environment Variables
```yaml
environment: prd|dev|local
namespace: knative-lambda-${environment}
service: knative-lambda-builder
```

### Time Ranges
- **Short-term**: 1h, 6h, 24h
- **Long-term**: 7d, 30d
- **Custom**: Based on alert investigation needs

## 🚨 Alert Thresholds

### Golden Signals Thresholds
| Signal | Warning | Critical | Duration |
|--------|---------|----------|----------|
| **Availability** | N/A | Down | 2m |
| **Error Rate** | >5% | >20% | 5m/2m |
| **Latency (P95)** | >30s | >120s | 5m/2m |
| **CPU Usage** | >80% | >90% | 5m |
| **Memory Usage** | >85% | >95% | 5m |

### Business Metrics Thresholds
| Metric | Warning | Critical | Duration |
|--------|---------|----------|----------|
| **Build Success Rate** | <99% | <95% | 5m |
| **Build Duration (P95)** | >30m | >60m | 5m |
| **Queue Depth** | >50 | >100 | 5m |
| **Job Failure Rate** | >10% | >20% | 5m |

## 🔧 Metric Labels

### Common Labels
All metrics include these common labels:
- `service`: Service name
- `version`: Service version
- `env`: Environment (prd/dev/local)

### HTTP Labels
- `method`: HTTP method (GET, POST, etc.)
- `endpoint`: API endpoint path
- `status_code`: HTTP status code
- `handler`: Handler function name

### Business Labels
- `third_party_id`: Third-party identifier
- `parser_id`: Parser identifier
- `job_type`: Kubernetes job type
- `error_type`: Error classification
- `priority`: Build priority level

### AWS Labels
- `bucket`: S3 bucket name
- `repository`: ECR repository name
- `status`: Operation status (success/failure)

## 📝 Metric Naming Convention

All metrics follow the Prometheus naming convention:
- **Counters**: `_total` suffix
- **Histograms**: `_bucket`, `_sum`, `_count` suffixes
- **Gauges**: No special suffix
- **Names**: Lowercase with underscores
- **Units**: Explicit units in metric names (seconds, bytes, percent)

## 🎯 SLO/SLI Targets

### Service Level Objectives
- **Availability**: 99.9% uptime
- **Error Rate**: <0.1% error rate
- **Latency**: P95 <30s for build requests
- **Throughput**: >100 builds/hour

### Service Level Indicators
- `up` > 0.999
- `error_rate` < 0.001
- `http_request_duration_seconds_p95` < 30
- `build_success_rate` > 0.99

## 🏗️ Metric Implementation Details

### 📍 **Where Metrics Are Defined**

#### 1. **Core Metric Definitions**
**File**: `internal/observability/observability.go`
- **Lines 242-540**: `initializeMetrics()` function
- **Lines 50-100**: `Metrics` struct definition
- **Lines 505-540**: Metric registration with Prometheus registry

#### 2. **Metric Structure**
```go
type Metrics struct {
    // HTTP Metrics
    HTTPRequestsTotal   *prometheus.CounterVec
    HTTPRequestDuration *prometheus.HistogramVec
    HTTPRequestSize     *prometheus.HistogramVec
    HTTPResponseSize    *prometheus.HistogramVec

    // Business Logic Metrics
    buildRequestsTotal   *prometheus.CounterVec
    buildRequestDuration *prometheus.HistogramVec
    buildSuccessTotal    *prometheus.CounterVec
    buildFailureTotal    *prometheus.CounterVec
    buildQueueSize       *prometheus.GaugeVec
    buildQueueDuration   *prometheus.HistogramVec

    // Kubernetes Metrics
    k8sJobCreationTotal    *prometheus.CounterVec
    k8sJobCreationDuration *prometheus.HistogramVec
    k8sJobSuccessTotal     *prometheus.CounterVec
    k8sJobFailureTotal     *prometheus.CounterVec
    k8sJobDuration         *prometheus.HistogramVec

    // AWS Metrics
    awsS3UploadTotal    *prometheus.CounterVec
    awsS3UploadDuration *prometheus.HistogramVec
    awsS3UploadSize     *prometheus.HistogramVec
    awsECRPushTotal     *prometheus.CounterVec
    awsECRPushDuration  *prometheus.HistogramVec

    // System Metrics
    systemMemoryUsage *prometheus.GaugeVec
    systemCPUUsage    *prometheus.GaugeVec
    systemGoroutines  *prometheus.GaugeVec
    systemHeapAlloc   *prometheus.GaugeVec

    // Error Metrics
    errorTotal *prometheus.CounterVec
    errorRate  *prometheus.GaugeVec

    registry *prometheus.Registry
}
```

### 📍 **Where Metrics Are Used**

#### 1. **CloudEvent Metrics**
**File**: `internal/observability/observability.go`
- **Lines 256-290**: CloudEvent metrics definition
```go
CloudEventsTotal: prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name:        "cloudevents_total",
        Help:        "Total number of CloudEvents processed",
        ConstLabels: commonLabels,
    },
    []string{"method", "endpoint", "status_code", "handler"},
),

CloudEventDuration: prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:        "cloudevent_duration_seconds",
        Help:        "CloudEvent processing duration in seconds",
        Buckets:     prometheus.DefBuckets,
        ConstLabels: commonLabels,
    },
    []string{"method", "endpoint", "handler"},
),

CloudEventSize: prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:        "cloudevent_size_bytes",
        Help:        "CloudEvent size in bytes",
        Buckets:     prometheus.ExponentialBuckets(100, 10, 8),
        ConstLabels: commonLabels,
    },
    []string{"method", "endpoint"},
),

CloudEventResponseSize: prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:        "cloudevent_response_size_bytes",
        Help:        "CloudEvent response size in bytes",
        Buckets:     prometheus.ExponentialBuckets(100, 10, 8),
        ConstLabels: commonLabels,
    },
    []string{"method", "endpoint", "status_code"},
),
```

**Note**: These CloudEvent metrics are defined but not yet implemented in the CloudEvent handler. They need to be integrated into the CloudEvent processing pipeline.

#### 2. **HTTP Request Metrics**
**File**: `internal/handler/middleware.go`
- **Lines 72-84**: HTTP request/response metrics recording
```go
obs.GetMetrics().HTTPRequestSize.WithLabelValues(r.Method, r.URL.Path).Observe(float64(requestSize))
obs.GetMetrics().HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, statusCode, "http").Inc()
obs.GetMetrics().HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path, "http").Observe(duration)
obs.GetMetrics().HTTPResponseSize.WithLabelValues(r.Method, r.URL.Path, statusCode).Observe(float64(responseWriter.size))
```

#### 2. **Build Process Metrics**
**File**: `internal/observability/observability.go`
- **Lines 745-824**: Build request tracking
```go
mr.obs.GetMetrics().buildRequestsTotal.WithLabelValues(thirdPartyID, parserID, status).Inc()
mr.obs.GetMetrics().buildRequestDuration.WithLabelValues(thirdPartyID, parserID).Observe(durationSeconds)
mr.obs.GetMetrics().buildSuccessTotal.WithLabelValues(thirdPartyID, parserID).Inc()
mr.obs.GetMetrics().buildFailureTotal.WithLabelValues(thirdPartyID, parserID, errorType).Inc()
```

#### 3. **Kubernetes Job Metrics**
**File**: `internal/observability/observability.go`
- **Lines 843-888**: K8s job tracking
```go
mr.obs.GetMetrics().k8sJobCreationTotal.WithLabelValues(jobType, status).Inc()
mr.obs.GetMetrics().k8sJobSuccessTotal.WithLabelValues(jobType).Inc()
mr.obs.GetMetrics().k8sJobFailureTotal.WithLabelValues(jobType, errorType).Inc()
```

#### 4. **AWS Service Metrics**
**File**: `internal/observability/observability.go`
- **Lines 906-940**: AWS S3 and ECR operations
```go
mr.obs.GetMetrics().awsS3UploadTotal.WithLabelValues(bucket, status).Inc()
mr.obs.GetMetrics().awsS3UploadDuration.WithLabelValues(bucket).Observe(duration.Seconds())
mr.obs.GetMetrics().awsECRPushTotal.WithLabelValues(repository, status).Inc()
```

#### 5. **System Metrics**
**File**: `internal/observability/observability.go`
- **Lines 991-1001**: System resource monitoring
```go
mr.obs.GetMetrics().systemMemoryUsage.WithLabelValues("alloc").Set(float64(m.Alloc))
mr.obs.GetMetrics().systemGoroutines.WithLabelValues().Set(float64(runtime.NumGoroutine()))
mr.obs.GetMetrics().systemHeapAlloc.WithLabelValues().Set(float64(m.HeapAlloc))
```

#### 6. **Error Metrics**
**File**: `internal/observability/observability.go`
- **Lines 592, 632, 679, 957**: Error tracking
```go
o.metrics.errorTotal.WithLabelValues("general", "unknown", "error").Inc()
o.metrics.errorTotal.WithLabelValues(labelValues...).Add(value)
o.metrics.errorTotal.WithLabelValues("security", eventType, "info").Inc()
```

### 📍 **Where Metrics Are Exposed**

#### 1. **HTTP Metrics Endpoint**
**File**: `internal/handler/middleware.go`
- **Lines 167-168**: `/metrics` endpoint handler
```go
if r.URL.Path == "/metrics" {
    obs.GetMetricsHandler().ServeHTTP(w, r)
}
```

#### 2. **Prometheus Handler**
**File**: `internal/observability/observability.go`
- **Lines 637-644**: Prometheus HTTP handler
```go
func (o *Observability) GetMetricsHandler() http.Handler {
    if o.metrics == nil {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            http.Error(w, "Metrics not enabled", http.StatusServiceUnavailable)
        })
    }
    return promhttp.HandlerFor(o.metrics.registry, promhttp.HandlerOpts{})
}
```

#### 3. **Service Initialization**
**File**: `cmd/service/main.go`
- **Lines 70-90**: Observability setup
```go
obs, err := observability.New(observability.Config{
    ServiceName:    cfg.Observability.ServiceName,
    ServiceVersion: cfg.Observability.ServiceVersion,
    Environment:    cfg.Environment,
    LogLevel:       cfg.Observability.LogLevel,
    MetricsEnabled: cfg.Observability.MetricsEnabled,
    TracingEnabled: cfg.Observability.TracingEnabled,
    OTLPEndpoint:   cfg.Observability.OTLPEndpoint,
    SampleRate:     cfg.Observability.SampleRate,
    Exemplars:      cfg.Observability.ToExemplarsConfig(),
})
```

### 📍 **Metric Collection & Scraping**

#### 1. **Prometheus ServiceMonitor**
**File**: `deploy/templates/servicemonitor.yaml`
- **Scraping Configuration**: Every 30s
- **Metrics Path**: `/metrics`
- **Port**: `http1` (8080)

#### 2. **Metrics Pusher**
**File**: `metrics-pusher/pusher.go`
- **Remote Write**: Pushes metrics to Prometheus remote write endpoint
- **Interval**: Every 30s
- **Failure Tolerance**: Configurable retry logic

#### 3. **System Metrics Collector**
**File**: `internal/observability/observability.go`
- **Lines 1075-1148**: Automatic system metrics collection
- **Interval**: Every 30s
- **Metrics**: Memory, CPU, goroutines, heap allocation

### 📍 **Alert Configuration**

#### 1. **Golden Signals Alerts**
**File**: `deploy/templates/alerts-golden-signals.yaml`
- **Lines 1-136**: Complete golden signals alerting
- **Availability**: `up` metric monitoring
- **Error Rate**: HTTP 5xx error percentage
- **Latency**: P95 request duration
- **Saturation**: CPU and memory usage

#### 2. **Business Metrics Alerts**
**File**: `deploy/templates/alerts.yaml`
- **Build Success Rate**: <99% threshold
- **Queue Depth**: >50 items threshold
- **Job Failure Rate**: >10% threshold

### 📍 **Dashboard Integration**

#### 1. **Grafana Dashboards**
- **Metrics Source**: Prometheus
- **Variables**: Environment, namespace, service
- **Panels**: Golden signals, business metrics, system metrics

#### 2. **Tempo Tracing**
- **Trace Context**: Correlation IDs in metrics
- **Exemplars**: Trace IDs linked to metrics
- **Integration**: Metrics + traces correlation

### 📍 **Configuration Files**

#### 1. **Values Configuration**
**File**: `deploy/values.yaml`
- **Lines 60-120**: Monitoring configuration
- **Alert Thresholds**: Configurable via values
- **SLO/SLI**: Service level objectives

#### 2. **Environment Variables**
**File**: `internal/config/config.go`
- **Metrics Enabled**: `METRICS_ENABLED=true`
- **Metrics Port**: `METRICS_PORT=9090`
- **Tracing Enabled**: `TRACING_ENABLED=true`
