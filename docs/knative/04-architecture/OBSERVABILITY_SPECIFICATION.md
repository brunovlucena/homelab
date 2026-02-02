# 🔭 Observability Specification

**Version**: 1.0.0  
**Last Updated**: December 4, 2025  
**Status**: Living Document

---

## 📖 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [OpenTelemetry Standard](#opentelemetry-standard)
- [RED Metrics](#red-metrics)
- [Distributed Tracing](#distributed-tracing)
- [Exemplars](#exemplars)
- [Log Aggregation](#log-aggregation)
- [Logfire Compatibility](#logfire-compatibility)
- [Grafana Stack Integration](#grafana-stack-integration)
- [Configuration](#configuration)
- [Runtime Instrumentation](#runtime-instrumentation)

---

## 🎯 Overview

The Knative Lambda platform implements a comprehensive observability stack using the **OpenTelemetry (OTEL)** standard. This ensures:

1. **Vendor-neutral instrumentation** - Same SDK works with Tempo, Jaeger, Zipkin, Logfire
2. **Automatic correlation** - Traces, metrics, and logs linked via trace context
3. **Exemplar support** - Jump from metrics to specific traces
4. **RED metrics** - Rate, Errors, Duration for every operation
5. **Logfire compatibility** - Native OTEL export to Pydantic Logfire

### Design Principles

| Principle | Implementation |
|-----------|----------------|
| **OTEL Native** | All components use OTEL SDK/API |
| **Zero-config tracing** | Automatic instrumentation via runtime wrappers |
| **Context propagation** | W3C Trace Context + Baggage standards |
| **Exemplar linking** | Prometheus metrics with trace ID exemplars |
| **Structured logging** | JSON logs with trace/span IDs |

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                         OBSERVABILITY ARCHITECTURE                                   │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│   ┌───────────────────────────────────────────────────────────────────────────────┐  │
│   │                           INSTRUMENTATION LAYER                               │  │
│   │                                                                               │  │
│   │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐   │  │
│   │  │ Lambda Runtime  │  │    Operator     │  │   Knative/RabbitMQ/K8s      │   │  │
│   │  │ (OTEL SDK)      │  │   (OTEL SDK)    │  │   (OTEL Instrumentation)    │   │  │
│   │  │                 │  │                 │  │                             │   │  │
│   │  │ • Python/Node/Go│  │ • Go OTEL SDK   │  │ • Knative Serving metrics   │   │  │
│   │  │ • Auto-instr.   │  │ • Exemplars     │  │ • RabbitMQ metrics          │   │  │
│   │  │ • Trace context │  │ • RED metrics   │  │ • K8s events                │   │  │
│   │  └────────┬────────┘  └────────┬────────┘  └─────────────┬───────────────┘   │  │
│   │           │                    │                         │                    │  │
│   └───────────┼────────────────────┼─────────────────────────┼────────────────────┘  │
│               │                    │                         │                       │
│               ▼                    ▼                         ▼                       │
│   ┌───────────────────────────────────────────────────────────────────────────────┐  │
│   │                           GRAFANA ALLOY (COLLECTOR)                           │  │
│   │                                                                               │  │
│   │   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐   │  │
│   │   │   OTLP      │    │  Prometheus │    │    Loki     │    │   Remote    │   │  │
│   │   │  Receiver   │    │  Remote     │    │   Push      │    │   Write     │   │  │
│   │   │  (traces,   │    │  Write      │    │   (logs)    │    │  (Logfire)  │   │  │
│   │   │   metrics)  │    │             │    │             │    │             │   │  │
│   │   └──────┬──────┘    └──────┬──────┘    └──────┬──────┘    └──────┬──────┘   │  │
│   │          │                  │                  │                  │          │  │
│   │          ├──────────────────┴──────────────────┴──────────────────┤          │  │
│   │          │                    PROCESSORS                          │          │  │
│   │          │  • Batch processing   • Attribute enrichment           │          │  │
│   │          │  • Tail sampling      • Resource detection             │          │  │
│   │          │  • Span metrics       • K8s metadata                   │          │  │
│   │          └────────────────────────────────────────────────────────┘          │  │
│   └───────────┬────────────────────┬─────────────────────┬────────────────────────┘  │
│               │                    │                     │                           │
│               ▼                    ▼                     ▼                           │
│   ┌───────────────────┐  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│   │   GRAFANA TEMPO   │  │   PROMETHEUS    │  │   GRAFANA LOKI  │  │   LOGFIRE   │  │
│   │                   │  │                 │  │                 │  │   (Cloud)   │  │
│   │   Distributed     │  │   Metrics       │  │   Logs          │  │             │  │
│   │   Traces          │  │   + Exemplars   │  │   + TraceID     │  │   OTEL      │  │
│   │                   │  │                 │  │                 │  │   Backend   │  │
│   └─────────┬─────────┘  └────────┬────────┘  └────────┬────────┘  └──────┬──────┘  │
│             │                     │                    │                  │         │
│             └─────────────────────┴────────────────────┴──────────────────┘         │
│                                          │                                          │
│                                          ▼                                          │
│                              ┌───────────────────────┐                              │
│                              │       GRAFANA         │                              │
│                              │                       │                              │
│                              │  • Unified Dashboards │                              │
│                              │  • Trace Exploration  │                              │
│                              │  • Log Correlation    │                              │
│                              │  • Exemplar Navigation│                              │
│                              │  • Alerting           │                              │
│                              └───────────────────────┘                              │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 📡 OpenTelemetry Standard

### W3C Trace Context

All components propagate trace context using W3C standards:

```http
traceparent: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01
tracestate: knative-lambda=build-12345
baggage: function.name=hello-python,function.namespace=knative-lambda
```

### OTEL Semantic Conventions

Lambda functions follow OTEL semantic conventions:

| Attribute | Description | Example |
|-----------|-------------|---------|
| `service.name` | Lambda function name | `hello-python` |
| `service.namespace` | Kubernetes namespace | `knative-lambda` |
| `service.version` | Image tag/version | `v1.0.0` |
| `deployment.environment` | Environment | `production` |
| `faas.trigger` | Trigger type | `http`, `cloudevent` |
| `faas.invocation_id` | Unique invocation ID | `inv-abc123` |
| `faas.coldstart` | Cold start indicator | `true`, `false` |

### Resource Attributes

Automatically enriched by Grafana Alloy:

```yaml
resource:
  k8s.namespace.name: knative-lambda
  k8s.pod.name: hello-python-00001-deployment-xyz
  k8s.deployment.name: hello-python-00001-deployment
  k8s.node.name: worker-1
  cloud.provider: kubernetes
  container.runtime: containerd
```

---

## 📊 RED Metrics

### Rate, Errors, Duration

Every operation exposes RED metrics with consistent labels:

```promql
# RATE - Requests per second
knative_lambda_invocations_total{function="hello-python", namespace="knative-lambda"}

# ERRORS - Error rate
knative_lambda_invocations_total{function="hello-python", status="error"} /
knative_lambda_invocations_total{function="hello-python"}

# DURATION - Latency percentiles
histogram_quantile(0.95, rate(knative_lambda_invocation_duration_seconds_bucket[5m]))
```

### Operator RED Metrics

```promql
# Build rate
rate(knative_lambda_operator_build_duration_seconds_count[5m])

# Build error rate
rate(knative_lambda_operator_build_duration_seconds_count{result="failed"}[5m]) /
rate(knative_lambda_operator_build_duration_seconds_count[5m])

# Build duration p95
histogram_quantile(0.95, rate(knative_lambda_operator_build_duration_seconds_bucket[5m]))
```

### Lambda Function RED Metrics

```promql
# Invocation rate
rate(knative_lambda_function_invocations_total[5m])

# Error rate by error type
rate(knative_lambda_function_errors_total{error_type="runtime"}[5m])

# Cold start rate
rate(knative_lambda_function_cold_starts_total[5m]) /
rate(knative_lambda_function_invocations_total[5m])

# Duration percentiles
histogram_quantile(0.50, rate(knative_lambda_function_duration_seconds_bucket[5m]))
histogram_quantile(0.95, rate(knative_lambda_function_duration_seconds_bucket[5m]))
histogram_quantile(0.99, rate(knative_lambda_function_duration_seconds_bucket[5m]))
```

### Metric Labels (Cardinality Optimized)

```yaml
# Low cardinality - safe for all metrics
function: "hello-python"          # Lambda function name
namespace: "knative-lambda"       # K8s namespace
runtime: "python"                 # Language runtime
status: "success|error|timeout"   # Invocation result
error_type: "runtime|validation|timeout|transient"  # Error classification

# Medium cardinality - use with caution
version: "v1.0.0"                 # Function version
trigger_type: "http|cloudevent"   # How function was invoked

# High cardinality - NEVER use as labels
# invocation_id, trace_id, user_id, request_id
# Use exemplars instead!
```

---

## 🔍 Distributed Tracing

### Trace Hierarchy

```
Trace: Lambda Invocation (hello-python)
│
├─ Span: receive_cloudevent
│  ├─ Attributes: ce.type, ce.source, ce.id
│  └─ Duration: 1ms
│
├─ Span: parse_event
│  ├─ Attributes: payload_size_bytes
│  └─ Duration: 2ms
│
├─ Span: handler_execution
│  ├─ Attributes: function.handler, cold_start
│  ├─ Events: handler.start, handler.complete
│  ├─ Duration: 150ms
│  │
│  └─ Child Spans (User Code):
│     ├─ Span: database_query
│     │  ├─ Attributes: db.system, db.statement
│     │  └─ Duration: 45ms
│     │
│     └─ Span: external_api_call
│        ├─ Attributes: http.url, http.method, http.status_code
│        └─ Duration: 80ms
│
└─ Span: build_response
   ├─ Attributes: response.status_code, response.size_bytes
   └─ Duration: 2ms
```

### Trace Context Propagation

Lambda runtime automatically propagates context:

```python
# Python - Automatic propagation via OTEL SDK
from opentelemetry import trace
from opentelemetry.propagate import inject, extract

tracer = trace.get_tracer(__name__)

def handler(event: dict, context: dict) -> dict:
    # Context automatically extracted from CloudEvent headers
    with tracer.start_as_current_span("user_operation") as span:
        span.set_attribute("custom.attribute", "value")
        # Outgoing requests automatically get trace context injected
        response = requests.get("https://api.example.com")
    return {"status": "ok"}
```

### Tempo Configuration

```yaml
# tempo-config.yaml
stream_over_http_enabled: true
server:
  http_listen_port: 3200

distributor:
  receivers:
    otlp:
      protocols:
        http:
        grpc:

storage:
  trace:
    backend: s3
    s3:
      bucket: tempo-traces
      endpoint: minio:9000
      
metrics_generator:
  processor:
    span_metrics:
      dimensions:
        - service.name
        - faas.trigger
        - faas.coldstart
        - status
```

---

## 🔗 Exemplars

### What are Exemplars?

Exemplars link metrics to specific traces, enabling:
- Click from a metric spike → see actual trace causing it
- Root cause analysis without searching logs
- Correlation between aggregate metrics and individual requests

### Exemplar Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                         EXEMPLAR FLOW                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   Lambda Function                     Prometheus                     │
│   ┌───────────────┐                  ┌───────────────┐              │
│   │ duration: 2.5s│ ────────────────►│ Duration=2.5s │              │
│   │ trace_id: abc │   exemplar       │ exemplar:     │              │
│   │               │                  │   trace_id=abc│              │
│   └───────────────┘                  └───────┬───────┘              │
│                                              │                       │
│                                              │ Grafana Query         │
│                                              ▼                       │
│                                      ┌───────────────┐              │
│                                      │   Dashboard   │              │
│                                      │               │              │
│                                      │  ○─────────○  │              │
│                                      │  │ p95     │  │              │
│                                      │  │ ● ←─────┼──┼── Exemplar   │
│                                      │  │         │  │   (click to  │
│                                      │  ○─────────○  │    view trace│
│                                      └───────────────┘              │
│                                              │                       │
│                                              │ Click exemplar        │
│                                              ▼                       │
│                                      ┌───────────────┐              │
│                                      │  Tempo Trace  │              │
│                                      │               │              │
│                                      │ trace_id: abc │              │
│                                      │ spans: [...]  │              │
│                                      └───────────────┘              │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Prometheus Exemplar Configuration

```yaml
# prometheus.yaml
global:
  scrape_interval: 15s
  
  # Enable exemplar storage
  exemplars:
    enabled: true
    max_exemplars: 100000

scrape_configs:
  - job_name: 'knative-lambda-operator'
    static_configs:
      - targets: ['knative-lambda-operator:8080']
    # Scrape exemplars from metrics endpoint
    enable_exemplars: true
```

### Go Implementation with Exemplars

```go
// metrics/metrics.go
package metrics

import (
    "context"
    "github.com/prometheus/client_golang/prometheus"
    "go.opentelemetry.io/otel/trace"
)

var InvocationDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Namespace: "knative_lambda",
        Subsystem: "function",
        Name:      "duration_seconds",
        Help:      "Duration of lambda invocations with exemplars",
        Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
    },
    []string{"function", "namespace", "status"},
)

// ObserveWithExemplar records duration with trace exemplar
func ObserveWithExemplar(ctx context.Context, function, namespace, status string, duration float64) {
    span := trace.SpanFromContext(ctx)
    exemplar := prometheus.Labels{}
    
    if span.SpanContext().IsValid() {
        exemplar["trace_id"] = span.SpanContext().TraceID().String()
        exemplar["span_id"] = span.SpanContext().SpanID().String()
    }
    
    InvocationDuration.WithLabelValues(function, namespace, status).
        (prometheus.ExemplarObserver).ObserveWithExemplar(duration, exemplar)
}
```

---

## 📝 Log Aggregation

### Structured Logging Format

All logs are JSON structured with trace correlation:

```json
{
  "timestamp": "2025-12-04T10:30:00.123Z",
  "level": "info",
  "message": "Handler execution completed",
  "logger": "lambda.runtime",
  
  "trace_id": "0af7651916cd43dd8448eb211c80319c",
  "span_id": "b7ad6b7169203331",
  "trace_flags": "01",
  
  "service": {
    "name": "hello-python",
    "namespace": "knative-lambda",
    "version": "v1.0.0"
  },
  
  "invocation": {
    "id": "inv-abc123",
    "cold_start": false,
    "duration_ms": 150
  },
  
  "cloudevent": {
    "id": "ce-xyz789",
    "type": "io.knative.lambda.invoke.sync",
    "source": "io.knative.lambda/trigger/knative-lambda/hello-python"
  }
}
```

### Loki Query Examples

```logql
# Find logs for a specific trace
{namespace="knative-lambda"} |= `trace_id":"0af7651916cd43dd8448eb211c80319c`

# Error logs with stack traces
{namespace="knative-lambda", app="hello-python"} | json | level="error"

# Slow invocations (>1s)
{namespace="knative-lambda"} | json | duration_ms > 1000

# Cold starts
{namespace="knative-lambda"} | json | cold_start="true"

# Correlate with metrics (via trace_id)
{namespace="knative-lambda"} | json | line_format "{{.trace_id}}"
```

### Log → Trace Correlation in Grafana

```yaml
# datasources.yaml
datasources:
  - name: Loki
    type: loki
    url: http://loki:3100
    jsonData:
      derivedFields:
        - name: TraceID
          matcherRegex: '"trace_id":"([a-f0-9]+)"'
          url: '$${__value.raw}'
          datasourceUid: tempo
          urlDisplayLabel: View Trace
```

---

## 🔥 Logfire Compatibility

### Pydantic Logfire Integration

Knative Lambda is fully compatible with [Pydantic Logfire](https://logfire.pydantic.dev/) through native OTEL export:

```
┌─────────────────────────────────────────────────────────────────────┐
│                      LOGFIRE INTEGRATION                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   Lambda Function                    Grafana Alloy                   │
│   ┌───────────────┐                  ┌───────────────┐              │
│   │ OTEL SDK      │ ─────────────────│  OTLP        │              │
│   │               │   OTLP/gRPC      │  Receiver    │              │
│   └───────────────┘                  └───────┬───────┘              │
│                                              │                       │
│                                              │ Fan-out               │
│                              ┌───────────────┼───────────────┐       │
│                              ▼               ▼               ▼       │
│                      ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│                      │   Tempo     │ │ Prometheus  │ │   Logfire   ││
│                      │   (traces)  │ │  (metrics)  │ │   (OTLP)    ││
│                      └─────────────┘ └─────────────┘ └─────────────┘│
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Alloy Configuration for Logfire

```river
// alloy-config.river

// Receive OTLP from Lambda functions
otelcol.receiver.otlp "default" {
  grpc {
    endpoint = "0.0.0.0:4317"
  }
  http {
    endpoint = "0.0.0.0:4318"
  }
  output {
    metrics = [otelcol.processor.batch.default.input]
    traces  = [otelcol.processor.batch.default.input]
    logs    = [otelcol.processor.batch.default.input]
  }
}

// Batch processing
otelcol.processor.batch "default" {
  timeout = "5s"
  send_batch_size = 1000
  output {
    metrics = [otelcol.exporter.prometheus.default.input, otelcol.exporter.otlphttp.logfire.input]
    traces  = [otelcol.exporter.otlp.tempo.input, otelcol.exporter.otlphttp.logfire.input]
    logs    = [otelcol.exporter.loki.default.input, otelcol.exporter.otlphttp.logfire.input]
  }
}

// Export to Logfire
otelcol.exporter.otlphttp "logfire" {
  client {
    endpoint = "https://logfire-api.pydantic.dev/v1/traces"
    headers = {
      "Authorization" = env("LOGFIRE_TOKEN"),
    }
  }
}

// Export to Tempo
otelcol.exporter.otlp "tempo" {
  client {
    endpoint = "tempo:4317"
    tls {
      insecure = true
    }
  }
}

// Export to Prometheus (with exemplars)
otelcol.exporter.prometheus "default" {
  endpoint {
    path = "/metrics"
    listen_address = "0.0.0.0:8889"
  }
}

// Export to Loki
otelcol.exporter.loki "default" {
  forward_to = [loki.write.default.receiver]
}

loki.write "default" {
  endpoint {
    url = "http://loki:3100/loki/api/v1/push"
  }
}
```

### Python Runtime with Logfire

```python
# runtime.py - Lambda runtime with Logfire support

import os
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource, SERVICE_NAME, SERVICE_NAMESPACE, SERVICE_VERSION
from opentelemetry.instrumentation.requests import RequestsInstrumentor

# Optional: Native Logfire support
LOGFIRE_ENABLED = os.environ.get("LOGFIRE_TOKEN", "") != ""

if LOGFIRE_ENABLED:
    import logfire
    logfire.configure(
        service_name=os.environ.get("FUNCTION_NAME", "lambda"),
        service_version=os.environ.get("FUNCTION_VERSION", "latest"),
    )
    # Logfire auto-instruments when configured
else:
    # Standard OTEL configuration (sends to Alloy → Tempo/Logfire)
    resource = Resource.create({
        SERVICE_NAME: os.environ.get("FUNCTION_NAME", "lambda"),
        SERVICE_NAMESPACE: os.environ.get("FUNCTION_NAMESPACE", "default"),
        SERVICE_VERSION: os.environ.get("FUNCTION_VERSION", "latest"),
        "faas.runtime": "python",
    })
    
    provider = TracerProvider(resource=resource)
    exporter = OTLPSpanExporter(
        endpoint=os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT", "http://alloy:4317"),
    )
    provider.add_span_processor(BatchSpanProcessor(exporter))
    trace.set_tracer_provider(provider)
    
    # Auto-instrument common libraries
    RequestsInstrumentor().instrument()
```

---

## 📊 Grafana Stack Integration

### Dashboard Templates

#### 1. Lambda Overview Dashboard

```json
{
  "panels": [
    {
      "title": "Invocation Rate",
      "type": "timeseries",
      "targets": [{
        "expr": "sum(rate(knative_lambda_function_invocations_total[5m])) by (function)",
        "legendFormat": "{{function}}"
      }]
    },
    {
      "title": "Error Rate",
      "type": "stat",
      "targets": [{
        "expr": "sum(rate(knative_lambda_function_errors_total[5m])) / sum(rate(knative_lambda_function_invocations_total[5m])) * 100"
      }],
      "fieldConfig": {
        "thresholds": [
          {"value": 0, "color": "green"},
          {"value": 1, "color": "yellow"},
          {"value": 5, "color": "red"}
        ]
      }
    },
    {
      "title": "Duration Heatmap",
      "type": "heatmap",
      "targets": [{
        "expr": "sum(rate(knative_lambda_function_duration_seconds_bucket[5m])) by (le)"
      }],
      "options": {
        "exemplars": {
          "enabled": true,
          "datasource": "tempo"
        }
      }
    },
    {
      "title": "Cold Start Rate",
      "type": "gauge",
      "targets": [{
        "expr": "sum(rate(knative_lambda_function_cold_starts_total[5m])) / sum(rate(knative_lambda_function_invocations_total[5m])) * 100"
      }]
    }
  ]
}
```

#### 2. Trace Exploration Panel

```json
{
  "title": "Recent Traces",
  "type": "traces",
  "datasource": "tempo",
  "targets": [{
    "query": "{resource.service.name=\"hello-python\"} | duration > 1s",
    "queryType": "search"
  }],
  "options": {
    "showLogsButton": true,
    "showExemplarLinks": true
  }
}
```

### Grafana Data Source Configuration

```yaml
# datasources.yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    url: http://prometheus:9090
    jsonData:
      exemplarTraceIdDestinations:
        - name: trace_id
          datasourceUid: tempo
    
  - name: Tempo
    type: tempo
    uid: tempo
    url: http://tempo:3200
    jsonData:
      httpMethod: GET
      tracesToLogs:
        datasourceUid: loki
        tags: ['service.name', 'function.name']
        filterByTraceID: true
        filterBySpanID: false
      tracesToMetrics:
        datasourceUid: prometheus
        tags:
          - key: 'service.name'
            value: 'function'
    
  - name: Loki
    type: loki
    uid: loki
    url: http://loki:3100
    jsonData:
      derivedFields:
        - name: TraceID
          matcherRegex: '"trace_id":"([a-f0-9]+)"'
          url: '$${__value.raw}'
          datasourceUid: tempo
```

---

## ⚙️ Configuration

### LambdaFunction CRD Observability Config

```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaFunction
metadata:
  name: hello-python
  namespace: knative-lambda
spec:
  # ... other spec fields ...
  
  observability:
    # Enable/disable tracing (default: true)
    tracing:
      enabled: true
      # Sampling rate (0.0 - 1.0, default: 1.0 for all traces)
      samplingRate: 1.0
      # OTLP endpoint (default: alloy.observability.svc:4317)
      endpoint: "alloy.observability.svc:4317"
      # Propagation format (default: w3c)
      propagation: "w3c"
    
    # Metrics configuration
    metrics:
      enabled: true
      # Enable exemplars linking metrics to traces
      exemplars: true
      # Custom metric labels
      labels:
        team: "platform"
        tier: "production"
    
    # Logging configuration  
    logging:
      # Log level (debug, info, warn, error)
      level: "info"
      # Include trace context in logs
      traceContext: true
      # JSON structured logging
      format: "json"
    
    # Logfire direct integration
    logfire:
      enabled: false
      # Set via secret: LOGFIRE_TOKEN
      tokenSecretRef:
        name: logfire-credentials
        key: token
```

### Environment Variables

Runtime containers receive these environment variables:

```yaml
env:
  # OTEL Configuration
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: "http://alloy.observability.svc:4317"
  - name: OTEL_SERVICE_NAME
    valueFrom:
      fieldRef:
        fieldPath: metadata.labels['app']
  - name: OTEL_RESOURCE_ATTRIBUTES
    value: "service.namespace=$(POD_NAMESPACE),k8s.pod.name=$(POD_NAME)"
  
  # Tracing
  - name: OTEL_TRACES_SAMPLER
    value: "parentbased_traceidratio"
  - name: OTEL_TRACES_SAMPLER_ARG
    value: "1.0"
  
  # Logging
  - name: OTEL_LOG_LEVEL
    value: "info"
  
  # Logfire (optional)
  - name: LOGFIRE_TOKEN
    valueFrom:
      secretKeyRef:
        name: logfire-credentials
        key: token
        optional: true
```

---

## 🎭 Runtime Instrumentation

### Python Auto-Instrumentation

The Python runtime wrapper includes OTEL auto-instrumentation:

```python
#!/usr/bin/env python3
"""
Lambda Runtime Wrapper with OpenTelemetry
"""

import os
from opentelemetry import trace, metrics
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.instrumentation.urllib3 import URLLib3Instrumentor
from opentelemetry.instrumentation.logging import LoggingInstrumentor
from opentelemetry.propagate import extract
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator

# Resource configuration
resource = Resource.create({
    "service.name": os.environ.get("FUNCTION_NAME", "lambda"),
    "service.namespace": os.environ.get("FUNCTION_NAMESPACE", "default"),
    "service.version": os.environ.get("FUNCTION_VERSION", "latest"),
    "deployment.environment": os.environ.get("ENVIRONMENT", "production"),
    "faas.runtime": "python",
})

# Tracer configuration
tracer_provider = TracerProvider(resource=resource)
otlp_exporter = OTLPSpanExporter(
    endpoint=os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT", "http://alloy:4317"),
)
tracer_provider.add_span_processor(BatchSpanProcessor(otlp_exporter))
trace.set_tracer_provider(tracer_provider)

# Meter configuration (for RED metrics)
meter_provider = MeterProvider(
    resource=resource,
    metric_readers=[
        PeriodicExportingMetricReader(
            OTLPMetricExporter(
                endpoint=os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT", "http://alloy:4317"),
            )
        )
    ]
)
metrics.set_meter_provider(meter_provider)

# Auto-instrumentation
RequestsInstrumentor().instrument()
URLLib3Instrumentor().instrument()
LoggingInstrumentor().instrument(set_logging_format=True)

# Create tracer and meter
tracer = trace.get_tracer(__name__)
meter = metrics.get_meter(__name__)

# RED Metrics
invocation_counter = meter.create_counter(
    "knative_lambda_function_invocations",
    description="Number of function invocations",
)
invocation_duration = meter.create_histogram(
    "knative_lambda_function_duration_seconds",
    description="Duration of function invocations",
)
error_counter = meter.create_counter(
    "knative_lambda_function_errors",
    description="Number of function errors",
)
cold_start_counter = meter.create_counter(
    "knative_lambda_function_cold_starts",
    description="Number of cold starts",
)


class LambdaRuntime:
    """OTEL-instrumented Lambda runtime."""
    
    def __init__(self):
        self.cold_start = True
        self.function_name = os.environ.get("FUNCTION_NAME", "lambda")
        self.namespace = os.environ.get("FUNCTION_NAMESPACE", "default")
        
    def handle_request(self, headers: dict, body: dict) -> dict:
        """Handle incoming CloudEvent with full OTEL instrumentation."""
        
        # Extract trace context from CloudEvent headers
        ctx = extract(headers, getter=TraceContextTextMapPropagator())
        
        attributes = {
            "function": self.function_name,
            "namespace": self.namespace,
            "faas.trigger": "cloudevent",
        }
        
        with tracer.start_as_current_span(
            "handler_execution",
            context=ctx,
            attributes=attributes,
        ) as span:
            start_time = time.time()
            status = "success"
            
            try:
                # Record cold start
                if self.cold_start:
                    cold_start_counter.add(1, attributes)
                    span.set_attribute("faas.coldstart", True)
                    self.cold_start = False
                else:
                    span.set_attribute("faas.coldstart", False)
                
                # Execute handler
                result = self.user_handler(body)
                
                return result
                
            except Exception as e:
                status = "error"
                error_counter.add(1, {**attributes, "error_type": type(e).__name__})
                span.record_exception(e)
                span.set_status(trace.StatusCode.ERROR, str(e))
                raise
                
            finally:
                duration = time.time() - start_time
                invocation_counter.add(1, {**attributes, "status": status})
                invocation_duration.record(duration, attributes)
                span.set_attribute("duration_ms", duration * 1000)
```

### Node.js Auto-Instrumentation

```javascript
// runtime.js - Node.js Lambda runtime with OTEL

const { NodeTracerProvider } = require('@opentelemetry/sdk-trace-node');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-grpc');
const { BatchSpanProcessor } = require('@opentelemetry/sdk-trace-base');
const { Resource } = require('@opentelemetry/resources');
const { SemanticResourceAttributes } = require('@opentelemetry/semantic-conventions');
const { MeterProvider, PeriodicExportingMetricReader } = require('@opentelemetry/sdk-metrics');
const { OTLPMetricExporter } = require('@opentelemetry/exporter-metrics-otlp-grpc');
const { registerInstrumentations } = require('@opentelemetry/instrumentation');
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http');
const { ExpressInstrumentation } = require('@opentelemetry/instrumentation-express');
const { WinstonInstrumentation } = require('@opentelemetry/instrumentation-winston');

// Resource
const resource = new Resource({
  [SemanticResourceAttributes.SERVICE_NAME]: process.env.FUNCTION_NAME || 'lambda',
  [SemanticResourceAttributes.SERVICE_NAMESPACE]: process.env.FUNCTION_NAMESPACE || 'default',
  [SemanticResourceAttributes.SERVICE_VERSION]: process.env.FUNCTION_VERSION || 'latest',
  'faas.runtime': 'nodejs',
});

// Tracer
const traceExporter = new OTLPTraceExporter({
  url: process.env.OTEL_EXPORTER_OTLP_ENDPOINT || 'http://alloy:4317',
});

const tracerProvider = new NodeTracerProvider({ resource });
tracerProvider.addSpanProcessor(new BatchSpanProcessor(traceExporter));
tracerProvider.register();

// Metrics
const metricExporter = new OTLPMetricExporter({
  url: process.env.OTEL_EXPORTER_OTLP_ENDPOINT || 'http://alloy:4317',
});

const meterProvider = new MeterProvider({
  resource,
  readers: [new PeriodicExportingMetricReader({ exporter: metricExporter })],
});

// Auto-instrumentation
registerInstrumentations({
  instrumentations: [
    new HttpInstrumentation(),
    new ExpressInstrumentation(),
    new WinstonInstrumentation(),
  ],
});

// RED Metrics
const meter = meterProvider.getMeter('knative-lambda');

const invocationCounter = meter.createCounter('knative_lambda_function_invocations', {
  description: 'Number of function invocations',
});

const invocationDuration = meter.createHistogram('knative_lambda_function_duration_seconds', {
  description: 'Duration of function invocations',
});

const errorCounter = meter.createCounter('knative_lambda_function_errors', {
  description: 'Number of function errors',
});

const coldStartCounter = meter.createCounter('knative_lambda_function_cold_starts', {
  description: 'Number of cold starts',
});

module.exports = {
  tracer: tracerProvider.getTracer('lambda-runtime'),
  meter,
  invocationCounter,
  invocationDuration,
  errorCounter,
  coldStartCounter,
};
```

---

## 📚 References

- [OpenTelemetry Specification](https://opentelemetry.io/docs/specs/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
- [Grafana Tempo Documentation](https://grafana.com/docs/tempo/)
- [Grafana Alloy Documentation](https://grafana.com/docs/alloy/)
- [Pydantic Logfire](https://logfire.pydantic.dev/)
- [Prometheus Exemplars](https://prometheus.io/docs/prometheus/latest/feature_flags/#exemplars-storage)
- [RED Method](https://grafana.com/blog/2018/08/02/the-red-method-how-to-instrument-your-services/)

---

**Maintainer**: Platform Team  
**Review Cycle**: Quarterly

