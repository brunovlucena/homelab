# Agent-Bruno Observability Setup

## ✅ Implementation Complete

Agent-Bruno has been updated to use the shared observability module with automatic Tempo tracing support.

## Changes Made

### 1. Updated `requirements.txt`
- Removed direct OpenTelemetry dependencies (now via shared-lib)
- Added `-e /app/shared-lib[otel]` to include observability module with OpenTelemetry

### 2. Updated `chatbot/main.py`
- Added observability initialization in `lifespan()` function
- Replaced manual `trace.get_tracer()` with `get_tracer()` from shared module
- Added trace context to structured logs
- Added tracing spans to:
  - `/health` endpoint
  - `/chat` endpoint
  - `/events` CloudEvent handler

### 3. LambdaAgent Configuration
The YAML already has observability configured:
```yaml
observability:
  tracing:
    enabled: true
    endpoint: alloy.observability.svc:4317
```

The LambdaAgent operator will automatically set:
- `OTEL_EXPORTER_OTLP_ENDPOINT=alloy.observability.svc:4317`

## How It Works

1. **Initialization**: On startup, `initialize_observability()` is called
   - Reads `OTEL_EXPORTER_OTLP_ENDPOINT` from environment (set by operator)
   - Initializes OpenTelemetry tracer with service name "agent-bruno"
   - Sets up OTLP exporter to send traces to Grafana Alloy

2. **Tracing**: All endpoints create spans
   - `/health` → `health_check` span
   - `/chat` → `chat_request` span
   - `/events` → `cloudevent.{event_type}` span

3. **Trace Flow**:
   ```
   Agent-Bruno → OpenTelemetry SDK → OTLP Exporter → Grafana Alloy → Tempo
   ```

4. **Logging**: Structured logs include trace context
   - `trace_id`: Full trace ID
   - `span_id`: Current span ID
   - `trace_flags`: Trace flags

## Verification

### 1. Check Traces in Tempo

1. Access Grafana: `http://grafana.prometheus.svc.cluster.local:3000`
2. Go to Explore → Select Tempo data source
3. Search for service name: `agent-bruno`
4. Look for traces from:
   - `health_check` spans
   - `chat_request` spans
   - `cloudevent.*` spans

### 2. Check Logs

Logs should include trace context:
```json
{
  "event": "agent_bruno_initialized",
  "version": "0.1.0",
  "observability_enabled": true,
  "trace_id": "0123456789abcdef0123456789abcdef",
  "span_id": "0123456789abcdef"
}
```

### 3. Check Metrics

Prometheus metrics should show observability status:
```promql
# Check if observability is enabled
agent_build_info{agent_id="agent-bruno"}
```

### 4. Test Endpoints

```bash
# Health check (creates health_check span)
curl http://agent-bruno.agent-bruno.svc.cluster.local:8080/health

# Chat request (creates chat_request span)
curl -X POST http://agent-bruno.agent-bruno.svc.cluster.local:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello"}'

# CloudEvent (creates cloudevent span)
curl -X POST http://agent-bruno.agent-bruno.svc.cluster.local:8080/events \
  -H "Content-Type: application/json" \
  -H "Ce-Specversion: 1.0" \
  -H "Ce-Type: io.homelab.chat.message" \
  -H "Ce-Source: test" \
  -H "Ce-Id: test-123" \
  -d '{"message": "test"}'
```

## Troubleshooting

### Traces Not Appearing in Tempo

1. **Check if observability initialized:**
   ```bash
   kubectl logs -n agent-bruno deployment/agent-bruno | grep observability
   ```
   Should see: `"observability_initialized": true`

2. **Check environment variables:**
   ```bash
   kubectl exec -n agent-bruno deployment/agent-bruno -- env | grep OTEL
   ```
   Should see: `OTEL_EXPORTER_OTLP_ENDPOINT=alloy.observability.svc:4317`

3. **Check Alloy is running:**
   ```bash
   kubectl get pods -n observability | grep alloy
   ```

4. **Check Alloy logs:**
   ```bash
   kubectl logs -n observability deployment/alloy | grep -i tempo
   ```

5. **Check Tempo is receiving traces:**
   ```bash
   kubectl logs -n tempo deployment/tempo | grep -i trace
   ```

### OpenTelemetry Not Available

If you see warnings about OpenTelemetry not being available:

1. **Check shared-lib installation:**
   ```bash
   kubectl exec -n agent-bruno deployment/agent-bruno -- pip list | grep opentelemetry
   ```

2. **Rebuild Docker image:**
   The Dockerfile should install shared-lib with otel dependencies:
   ```dockerfile
   RUN pip install -e /app/shared-lib[otel]
   ```

### Trace Context Not in Logs

If trace context is missing from logs:

1. **Check if observability is initialized:**
   ```python
   from observability import is_observability_enabled
   print(is_observability_enabled())  # Should be True
   ```

2. **Check if tracer is available:**
   ```python
   from observability import get_tracer
   tracer = get_tracer()
   print(tracer is not None)  # Should be True
   ```

## Next Steps

1. **Deploy the updated agent:**
   ```bash
   kubectl apply -f k8s/kustomize/base/lambdaagent.yaml
   ```

2. **Monitor traces in Grafana:**
   - Set up dashboards for agent-bruno traces
   - Create alerts for trace errors

3. **Add more instrumentation:**
   - Add spans to handler.py methods
   - Add spans to memory operations
   - Add spans to LLM calls

## References

- [Shared Observability Module README](../../shared-lib/observability/README.md)
- [Grafana Tempo Documentation](https://grafana.com/docs/tempo/latest/)
- [OpenTelemetry Python Documentation](https://opentelemetry.io/docs/instrumentation/python/)


