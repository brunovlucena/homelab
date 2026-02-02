# 🔭 Shared Observability Module

Type-safe OpenTelemetry initialization for homelab agents with automatic Tempo tracing support.

## Features

- ✅ **Type-safe configuration** using `pydantic-settings`
- ✅ **Automatic OpenTelemetry initialization** for tracing and metrics
- ✅ **Grafana Tempo integration** via Grafana Alloy
- ✅ **Environment variable support** with `OTEL_` prefix
- ✅ **Graceful degradation** if OpenTelemetry is not installed
- ✅ **Resource attributes** for service identification
- ✅ **Configurable sampling** strategies

## Installation

The observability module is part of `shared-lib`. Install it in your agent:

```bash
# In agent's requirements.txt
-e ../shared-lib

# Or with OpenTelemetry dependencies
-e ../shared-lib[otel]
```

## Quick Start

### Basic Usage

```python
from observability import initialize_observability

# Initialize with environment variables
# Requires: OTEL_SERVICE_NAME, OTEL_EXPORTER_OTLP_ENDPOINT
initialize_observability()
```

### With Explicit Settings

```python
from observability import initialize_observability, ObservabilitySettings

settings = ObservabilitySettings(
    otel_service_name="agent-bruno",
    otel_service_namespace="agent-bruno",
    otel_service_version="1.0.0",
    otel_exporter_otlp_endpoint="alloy.observability.svc:4317",
)

initialize_observability(settings)
```

### With Service Name Override

```python
from observability import initialize_observability

# Reads other settings from environment, but overrides service name
initialize_observability(
    service_name="agent-bruno",
    service_namespace="agent-bruno",
    service_version="1.0.0",
)
```

## Configuration

### Environment Variables

All settings can be configured via environment variables with `OTEL_` prefix:

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP collector endpoint | `alloy.observability.svc:4317` |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | Protocol: `grpc` or `http` | `grpc` |
| `OTEL_SERVICE_NAME` | Service name (required) | - |
| `OTEL_SERVICE_VERSION` | Service version | `0.1.0` |
| `OTEL_SERVICE_NAMESPACE` | Kubernetes namespace | `default` |
| `OTEL_DEPLOYMENT_ENVIRONMENT` | Environment name | `production` |
| `OTEL_TRACING_ENABLED` | Enable tracing | `true` |
| `OTEL_METRICS_ENABLED` | Enable metrics | `true` |
| `OTEL_TRACES_SAMPLER` | Sampling strategy | `always_on` |
| `OTEL_TRACES_SAMPLER_ARG` | Sampling rate (0.0-1.0) | `1.0` |

### LambdaAgent Configuration

The LambdaAgent operator automatically sets `OTEL_EXPORTER_OTLP_ENDPOINT` from the YAML:

```yaml
observability:
  tracing:
    enabled: true
    endpoint: alloy.observability.svc:4317
  metrics:
    enabled: true
```

## Usage Examples

### FastAPI Application

```python
from contextlib import asynccontextmanager
from fastapi import FastAPI
from observability import initialize_observability, get_tracer
from opentelemetry import trace

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Initialize observability at startup
    initialize_observability(
        service_name="agent-bruno",
        service_namespace="agent-bruno",
    )
    
    yield
    
    # Cleanup (automatic)

app = FastAPI(lifespan=lifespan)

@app.get("/health")
async def health():
    tracer = get_tracer()
    with tracer.start_as_current_span("health_check") as span:
        span.set_attribute("endpoint", "/health")
        return {"status": "ok"}
```

### With Manual Tracing

```python
from observability import get_tracer
from opentelemetry import trace

tracer = get_tracer()

def process_event(event):
    with tracer.start_as_current_span("process_event") as span:
        span.set_attribute("event.type", event.type)
        span.set_attribute("event.id", event.id)
        
        # Process event
        result = do_work(event)
        
        span.set_attribute("result.success", True)
        return result
```

### With Structured Logging

```python
import structlog
from observability import get_current_trace_context

logger = structlog.get_logger()

def log_with_trace(message: str, **kwargs):
    trace_ctx = get_current_trace_context()
    logger.info(message, **trace_ctx, **kwargs)
```

## Sampling Strategies

| Strategy | Description | Use Case |
|----------|-------------|----------|
| `always_on` | Sample all traces | Development, low-traffic |
| `always_off` | Sample no traces | Disable tracing |
| `traceidratio` | Sample based on trace ID hash | Production (e.g., 0.1 = 10%) |
| `parentbased_traceidratio` | Respect parent sampling decision | Distributed systems |

Example with 10% sampling:

```python
settings = ObservabilitySettings(
    otel_service_name="agent-bruno",
    otel_traces_sampler="traceidratio",
    otel_traces_sampler_arg=0.1,  # 10% sampling
)
```

## Integration with Existing Observability

The module integrates with existing observability libraries:

- **agent_communication.observability**: Uses shared tracer
- **agent_memory.observability**: Uses shared tracer
- **Prometheus metrics**: Works alongside OpenTelemetry metrics

## Troubleshooting

### Traces Not Appearing in Tempo

1. **Check endpoint configuration:**
   ```python
   from observability import get_settings
   settings = get_settings()
   print(f"Endpoint: {settings.otel_exporter_otlp_endpoint}")
   ```

2. **Verify OpenTelemetry is installed:**
   ```bash
   pip install opentelemetry-api opentelemetry-sdk opentelemetry-exporter-otlp-proto-grpc
   ```

3. **Check logs for initialization errors:**
   ```python
   import structlog
   structlog.configure(processors=[structlog.processors.JSONRenderer()])
   ```

4. **Verify Alloy is running:**
   ```bash
   kubectl get pods -n observability | grep alloy
   ```

### Graceful Degradation

If OpenTelemetry is not installed, the module will:
- Log warnings but not crash
- Return `None` from `get_tracer()` and `get_meter()`
- Continue operating without tracing

## API Reference

### `ObservabilitySettings`

Type-safe configuration class using pydantic-settings.

### `initialize_observability(settings=None, **kwargs)`

Initialize OpenTelemetry tracing and metrics.

### `get_tracer(name=None)`

Get OpenTelemetry tracer instance.

### `get_meter(name=None)`

Get OpenTelemetry meter instance.

### `is_observability_enabled()`

Check if observability has been initialized.

## See Also

- [OpenTelemetry Python Documentation](https://opentelemetry.io/docs/instrumentation/python/)
- [Grafana Tempo Documentation](https://grafana.com/docs/tempo/latest/)
- [Pydantic Settings Documentation](https://docs.pydantic.dev/latest/concepts/pydantic_settings/)


