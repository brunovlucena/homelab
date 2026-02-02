# 📊 TRM Model Monitoring Setup

This document describes how to monitor the TRM model performance in production.

## Metrics to Track

### 1. Model Performance Metrics

- **Accuracy**: Percentage of correct lambda function selections
- **Confidence Scores**: Distribution of confidence scores from TRM
- **Inference Latency**: Time taken for TRM to reason and select remediation
- **Fallback Rate**: How often TRM falls back to rule-based selection

### 2. Remediation Success Metrics

- **Remediation Success Rate**: Percentage of successful remediations
- **Lambda Function Usage**: Which functions are called most often
- **Error Rates**: Failed remediations by type

### 3. Model Health Metrics

- **Model Load Status**: Whether model loads successfully
- **Memory Usage**: Model memory consumption
- **Inference Errors**: Errors during model inference

## Implementation

### Prometheus Metrics

Add these metrics to `agent-sre/src/sre_agent/observability.py`:

```python
from prometheus_client import Counter, Histogram, Gauge

# TRM Model Metrics
trm_inference_total = Counter(
    'trm_inference_total',
    'Total number of TRM inference calls',
    ['method', 'status']
)

trm_inference_duration = Histogram(
    'trm_inference_duration_seconds',
    'Time spent in TRM inference',
    ['method']
)

trm_confidence_score = Histogram(
    'trm_confidence_score',
    'TRM confidence scores',
    ['alertname']
)

trm_fallback_total = Counter(
    'trm_fallback_total',
    'Number of times TRM falls back to rule-based',
    ['reason']
)

trm_model_loaded = Gauge(
    'trm_model_loaded',
    'Whether TRM model is loaded (1=loaded, 0=not loaded)'
)

remediation_success_total = Counter(
    'remediation_success_total',
    'Successful remediations',
    ['lambda_function', 'method']
)

remediation_failure_total = Counter(
    'remediation_failure_total',
    'Failed remediations',
    ['lambda_function', 'method', 'error_type']
)
```

### Logging

Structured logging is already in place via `structlog`. Key events to log:

```python
logger.info(
    "trm_inference_start",
    alertname=alertname,
    correlation_id=correlation_id
)

logger.info(
    "trm_inference_complete",
    alertname=alertname,
    lambda_function=result["lambda_function"],
    confidence=result["confidence"],
    method=result["method"],
    duration_ms=duration_ms,
    correlation_id=correlation_id
)

logger.warning(
    "trm_fallback",
    alertname=alertname,
    reason=reason,
    correlation_id=correlation_id
)
```

### Grafana Dashboard

Create a dashboard with:

1. **TRM Performance Panel**
   - Inference latency (p50, p95, p99)
   - Success rate over time
   - Confidence score distribution

2. **Remediation Success Panel**
   - Success rate by lambda function
   - Success rate by alert type
   - Error breakdown

3. **Model Health Panel**
   - Model load status
   - Memory usage
   - Error rate

## Example Queries

### TRM Inference Rate
```promql
rate(trm_inference_total[5m])
```

### Average Confidence Score
```promql
avg(trm_confidence_score)
```

### Remediation Success Rate
```promql
rate(remediation_success_total[5m]) / 
(rate(remediation_success_total[5m]) + rate(remediation_failure_total[5m]))
```

### TRM Fallback Rate
```promql
rate(trm_fallback_total[5m]) / rate(trm_inference_total[5m])
```

## Alerting Rules

### High Fallback Rate
```yaml
- alert: TRMHighFallbackRate
  expr: rate(trm_fallback_total[5m]) / rate(trm_inference_total[5m]) > 0.5
  for: 10m
  annotations:
    summary: "TRM falling back to rule-based > 50% of the time"
```

### Low Confidence Scores
```yaml
- alert: TRMLowConfidence
  expr: avg(trm_confidence_score) < 0.5
  for: 15m
  annotations:
    summary: "TRM average confidence score below 0.5"
```

### Model Not Loaded
```yaml
- alert: TRMModelNotLoaded
  expr: trm_model_loaded == 0
  for: 5m
  annotations:
    summary: "TRM model failed to load"
```

## Testing Monitoring

Test the monitoring setup:

```bash
# Send test alert
curl -X POST http://agent-sre:8080/cloudevents \
  -H "Content-Type: application/json" \
  -d '{
    "type": "io.homelab.prometheus.alert.fired",
    "data": {
      "labels": {
        "alertname": "FluxReconciliationFailure",
        "name": "test-app",
        "namespace": "flux-system"
      }
    }
  }'

# Check metrics
curl http://agent-sre:8080/metrics | grep trm_
```

## Next Steps

1. ✅ Add metrics to observability.py
2. ✅ Update TRM selector to emit metrics
3. ✅ Create Grafana dashboard
4. ✅ Set up alerting rules
5. ✅ Document monitoring runbook
