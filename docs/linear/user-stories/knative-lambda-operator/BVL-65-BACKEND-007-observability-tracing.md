# 🌐 BACKEND-007: Observability and Distributed Tracing

**Priority**: P1 | **Status**: ✅ Implemented  | **Story Points**: 8
**Linear URL**: https://linear.app/bvlucena/issue/BVL-227/backend-007-observability-and-distributed-tracing

---

## 📋 User Story

**As a** Backend Developer  
**I want to** implement comprehensive observability with OpenTelemetry  
**So that** I can monitor, debug, and optimize the system through traces, metrics, and structured logs

---

## 🎯 Acceptance Criteria

### ✅ Distributed Tracing
- [ ] OpenTelemetry SDK integration
- [ ] Trace context propagation across services
- [ ] Parent-child span relationships
- [ ] Span attributes for all operations
- [ ] Trace sampling configuration
- [ ] Export traces to Tempo backend

### ✅ Metrics Collection
- [ ] Prometheus metrics exposition
- [ ] Counter metrics for operations
- [ ] Histogram metrics for latency
- [ ] Gauge metrics for resource usage
- [ ] Custom metrics for business logic
- [ ] Exemplars linking metrics to traces

### ✅ Structured Logging
- [ ] JSON structured logs
- [ ] Log levels: DEBUG, INFO, WARN, ERROR
- [ ] Correlation IDs in all logs
- [ ] Trace IDs linking logs to traces
- [ ] Context-aware logging
- [ ] Log aggregation ready

### ✅ Correlation and Context
- [ ] Generate correlation ID per request
- [ ] Propagate correlation ID through system
- [ ] Link traces, metrics, and logs
- [ ] Support W3C Trace Context
- [ ] Include correlation ID in responses

### ✅ Performance
- [ ] Low overhead (< 5% CPU)
- [ ] Async trace export
- [ ] Buffered metric collection
- [ ] Configurable sampling rates
- [ ] Graceful degradation on errors

---

## 🔧 Technical Implementation

### File: `internal/observability/observability.go`

```go
// Observability provides comprehensive monitoring capabilities
type Observability struct {
    logger   *logrus.Logger
    tracer   trace.Tracer
    metrics  *Metrics
    config   *ObservabilityConfig
}

type ObservabilityConfig struct {
    ServiceName     string
    ServiceVersion  string
    Environment     string
    OTLPEndpoint    string
    SamplingRate    float64
    LogLevel        string
}

// Initialize OpenTelemetry
func NewObservability(config *ObservabilityConfig) (*Observability, error) {
    // 1. Setup Trace Provider
    traceProvider, err := setupTraceProvider(config)
    if err != nil {
        return nil, fmt.Errorf("failed to setup trace provider: %w", err)
    }
    
    // 2. Setup Metric Provider
    metricProvider, err := setupMetricProvider(config)
    if err != nil {
        return nil, fmt.Errorf("failed to setup metric provider: %w", err)
    }
    
    // 3. Setup Structured Logger
    logger := setupLogger(config)
    
    // 4. Create Tracer
    tracer := traceProvider.Tracer(
        config.ServiceName,
        trace.WithInstrumentationVersion(config.ServiceVersion),
    )
    
    // 5. Create Metrics
    metrics := NewMetrics(metricProvider)
    
    return &Observability{
        logger:  logger,
        tracer:  tracer,
        metrics: metrics,
        config:  config,
    }, nil
}

// Start Span with automatic context propagation
func (o *Observability) StartSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
    // Extract correlation ID from context
    correlationID := GetCorrelationID(ctx)
    
    // Start span with attributes
    ctx, span := o.tracer.Start(ctx, operationName,
        trace.WithSpanKind(trace.SpanKindServer),
        trace.WithAttributes(
            attribute.String("service.name", o.config.ServiceName),
            attribute.String("correlation.id", correlationID),
        ),
    )
    
    return ctx, span
}

// StartSpanWithAttributes creates span with custom attributes
func (o *Observability) StartSpanWithAttributes(ctx context.Context, operationName string, attributes map[string]string) (context.Context, trace.Span) {
    ctx, span := o.StartSpan(ctx, operationName)
    
    // Add custom attributes
    for key, value := range attributes {
        span.SetAttributes(attribute.String(key, value))
    }
    
    return ctx, span
}

// Structured Logging with Trace Context
func (o *Observability) Info(ctx context.Context, message string, fields ...interface{}) {
    entry := o.logger.WithContext(ctx)
    
    // Add correlation ID
    if correlationID := GetCorrelationID(ctx); correlationID != "" {
        entry = entry.WithField("correlation_id", correlationID)
    }
    
    // Add trace context
    if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
        entry = entry.WithFields(logrus.Fields{
            "trace_id": span.SpanContext().TraceID().String(),
            "span_id":  span.SpanContext().SpanID().String(),
        })
    }
    
    // Add custom fields
    if len(fields) > 0 {
        entry = entry.WithFields(buildFields(fields...))
    }
    
    entry.Info(message)
}

func (o *Observability) Error(ctx context.Context, err error, message string, fields ...interface{}) {
    entry := o.logger.WithContext(ctx).WithError(err)
    
    // Add correlation ID
    if correlationID := GetCorrelationID(ctx); correlationID != "" {
        entry = entry.WithField("correlation_id", correlationID)
    }
    
    // Add trace context and record error in span
    if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
        span.RecordError(err)
        span.SetStatus(codes.Error, message)
        
        entry = entry.WithFields(logrus.Fields{
            "trace_id": span.SpanContext().TraceID().String(),
            "span_id":  span.SpanContext().SpanID().String(),
        })
    }
    
    // Add custom fields
    if len(fields) > 0 {
        entry = entry.WithFields(buildFields(fields...))
    }
    
    entry.Error(message)
}
```

### Metrics with Exemplars

```go
type MetricsRecorder struct {
    obs *Observability
}

// Record build request with exemplar linking to trace
func (m *MetricsRecorder) RecordBuildRequest(ctx context.Context, thirdPartyID, parserID, status string) {
    // Get trace context for exemplar
    span := trace.SpanFromContext(ctx)
    traceID := span.SpanContext().TraceID().String()
    
    // Record metric with exemplar
    m.obs.metrics.BuildRequestsTotal.Add(ctx, 1,
        metric.WithAttributes(
            attribute.String("third_party_id", thirdPartyID),
            attribute.String("parser_id", parserID),
            attribute.String("status", status),
        ),
        // Exemplar links metric to trace
        metric.WithExemplar(traceID),
    )
}

// Record build duration with histogram
func (m *MetricsRecorder) RecordBuildDuration(ctx context.Context, thirdPartyID, parserID string, duration time.Duration) {
    m.obs.metrics.BuildDurationSeconds.Record(ctx, duration.Seconds(),
        metric.WithAttributes(
            attribute.String("third_party_id", thirdPartyID),
            attribute.String("parser_id", parserID),
        ),
    )
}
```

### Correlation ID Management

```go
type correlationIDKey struct{}

var CorrelationIDKey = correlationIDKey{}

// Generate or extract correlation ID
func EnsureCorrelationID(ctx context.Context) context.Context {
    if correlationID := GetCorrelationID(ctx); correlationID != "" {
        return ctx
    }
    
    // Generate new correlation ID
    correlationID := uuid.New().String()
    return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

func GetCorrelationID(ctx context.Context) string {
    if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok {
        return correlationID
    }
    return ""
}
```

---

## 📊 Observability Flow

```
┌─────────────────┐
│  HTTP Request   │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│ Generate        │ ← Correlation ID
│ Correlation ID  │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│  Start Span     │ ← Root span with correlation ID
└────────┬────────┘
         │
         ├──────────────────────────────┐
         │                              │
         ↓                              ↓
┌─────────────────┐          ┌─────────────────┐
│  Child Spans    │          │  Structured Logs│
│  (nested)       │          │  (with trace)   │
└────────┬────────┘          └────────┬────────┘
         │                            │
         ↓                            ↓
┌─────────────────┐          ┌─────────────────┐
│  Metrics        │          │  Exemplars      │
│  (counters)     │◄─────────┤  (link to trace)│
└────────┬────────┘          └─────────────────┘
         │
         ↓
┌─────────────────┐
│  Export to      │
│  - Tempo        │
│  - Prometheus   │
│  - Loki         │
└─────────────────┘
```

---

## 🧪 Testing Scenarios

### 1. End-to-End Tracing
```bash
# Send request with custom correlation ID
curl -X POST http://localhost:8080/events \
  -H "X-Correlation-ID: test-correlation-123" \
  -H "Ce-Type: network.notifi.lambda.build.start" \
  -d '{"parser_id":"test-parser"}'

# Query trace in Tempo
curl "http://tempo:3200/api/search?q={correlation_id=\"test-correlation-123\"}"
```

**Expected**:
- Trace created with correlation ID
- All spans linked in parent-child hierarchy
- Logs include trace ID
- Metrics have exemplars pointing to trace

### 2. Metrics to Traces Navigation
```bash
# Query Prometheus metric
curl 'http://prometheus:9090/api/v1/query?query=build_requests_total{status="success"}'

# Get exemplar trace ID
# Use trace ID to query Tempo
curl "http://tempo:3200/api/traces/{trace_id}"
```

**Expected**:
- Metric has exemplar with trace ID
- Can navigate from metric to trace
- Trace shows full request lifecycle

### 3. Log to Trace Correlation
```bash
# Query logs in Loki
curl 'http://loki:3100/loki/api/v1/query?query={job="knative-lambda-builder"} | json | trace_id="abc123"'

# Extract trace ID from logs
# Query trace in Tempo
```

**Expected**:
- Logs contain trace_id field
- Can navigate from log entry to trace
- Full context available

---

## 📈 Performance Requirements

- **Tracing Overhead**: < 5% CPU
- **Metric Collection**: < 1% CPU
- **Logging Overhead**: < 2% CPU
- **Export Latency**: < 1s for traces
- **Memory Usage**: < 100MB for buffers

---

## 🔍 Monitoring & Alerts

### Key Metrics
```promql
# Request rate
rate(http_requests_total[5m])

# Error rate
rate(http_requests_total{status=~"5.."}[5m])
/ rate(http_requests_total[5m])

# Request latency
histogram_quantile(0.99,
  rate(http_request_duration_seconds_bucket[5m])
)

# Build duration
histogram_quantile(0.95,
  rate(build_duration_seconds_bucket[5m])
)
```

### Trace Queries (Tempo/Jaeger)
```
# Find slow requests
{ duration > 1s }

# Find errors
{ status.code = error }

# Find by service
{ service.name = "knative-lambda-builder" }

# Find by correlation ID
{ correlation.id = "test-123" }
```

---

## 🏗️ Code References

**Main Files**:
- `internal/observability/observability.go` - Core observability
- `internal/observability/metrics.go` - Metrics definitions
- `internal/observability/exemplars.go` - Exemplar configuration
- `internal/config/observability.go` - Observability config

**Configuration**:
```yaml
observability:
  service_name: knative-lambda-builder
  service_version: 0.1.0
  environment: dev
  otlp_endpoint: tempo:4317
  sampling_rate: 1.0
  log_level: info
```

---

## 📚 Related Documentation

- OpenTelemetry Go: https://opentelemetry.io/docs/languages/go/
- Tempo: https://grafana.com/docs/tempo/
- Prometheus: https://prometheus.io/docs/
- W3C Trace Context: https://www.w3.org/TR/trace-context/

---

**Last Updated**: October 29, 2025  
**Owner**: Backend Team  
**Status**: Production Ready

