# 🔍 Observability Review: Pydantic & Grafana Tempo Tracing

**Review Date:** 2025-01-15  
**Reviewer:** ML Engineer  
**Scope:** All agents in homelab for pydantic observability usage and Tempo tracing configuration

---

## Executive Summary

### ✅ Findings

1. **Pydantic Usage**: Most agents use `pydantic` for data models (`BaseModel`), but **NOT** for observability configuration
2. **Tempo Tracing**: Most agents are configured to send traces to Tempo via Grafana Alloy (`alloy.observability.svc:4317`)
3. **OpenTelemetry**: Most agents have OpenTelemetry libraries installed, but initialization patterns vary

### ⚠️ Issues Identified

1. **No Pydantic Settings for Observability**: Agents are not using `pydantic-settings` (`BaseSettings`) for type-safe observability configuration
2. **Inconsistent OpenTelemetry Initialization**: Some agents manually initialize exporters, others rely on auto-instrumentation
3. **Missing Exporter Initialization**: Shared observability libraries provide utilities but don't initialize exporters

---

## Detailed Agent Review

### ✅ Agents with OpenTelemetry & Tempo Configuration

| Agent | Pydantic (Models) | Pydantic Settings | OTEL Libraries | Tempo Endpoint | Manual OTEL Init |
|-------|------------------|-------------------|----------------|----------------|------------------|
| **agent-bruno** | ✅ | ❌ | ✅ | ✅ `alloy.observability.svc:4317` | ❌ |
| **agent-medical** | ✅ | ✅ (general config) | ✅ | ✅ `alloy.observability.svc:4317` | ❌ |
| **agent-contracts** | ✅ | ❌ | ✅ | ✅ `alloy.observability.svc:4317` | ❌ |
| **agent-tools** | ✅ | ❌ | ✅ | ✅ `alloy.observability.svc:4317` | ❌ |
| **agent-restaurant** | ✅ | ❌ | ✅ | ✅ Manual via `OTEL_EXPORTER_OTLP_ENDPOINT` | ✅ |
| **agent-store-multibrands** | ✅ | ❌ | ✅ | ✅ `alloy.observability.svc:4317` | ❌ |
| **agent-pos-edge** | ✅ | ❌ | ✅ | ✅ `alloy.observability.svc:4317` | ❌ |
| **agent-speech-coach** | ✅ | ✅ (general config) | ✅ | ✅ `alloy.observability.svc:4317` | ❌ |
| **agent-redteam** | ✅ | ❌ | ✅ | ✅ `alloy.observability.svc:4317` | ❌ |
| **agent-blueteam** | ✅ | ❌ | ✅ | ✅ `alloy.observability.svc:4317` | ❌ |
| **agent-devsecops** | ✅ | ❌ | ✅ | ✅ `alloy.observability.svc:4317` | ❌ |
| **agent-reasoning** | ✅ | ❌ | ✅ | ❓ Not configured | ❌ |
| **futboss-ai** | ✅ | ✅ (general config) | ❌ | ❌ | ❌ |

---

## Architecture Analysis

### 1. Pydantic Usage

**Current State:**
- ✅ All agents use `pydantic.BaseModel` for data validation
- ❌ **No agents use `pydantic-settings.BaseSettings` for observability configuration**
- Only `futboss-ai`, `agent-medical`, and `agent-speech-coach` use `pydantic-settings` for general app config

**Example from `futboss-ai/apps/api/app/config.py`:**
```python
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    api_host: str = "0.0.0.0"
    api_port: int = 8000
    # ... but no observability settings
```

**Recommendation:**
Agents should use `pydantic-settings` for type-safe observability configuration:

```python
from pydantic_settings import BaseSettings

class ObservabilitySettings(BaseSettings):
    otel_exporter_otlp_endpoint: str = "alloy.observability.svc:4317"
    otel_service_name: str
    otel_service_version: str = "0.1.0"
    tracing_enabled: bool = True
    metrics_enabled: bool = True
    logging_level: str = "info"
    
    class Config:
        env_prefix = "OTEL_"
```

### 2. Tempo Tracing Configuration

**Current State:**
- ✅ Most agents configured with `observability.tracing.endpoint: alloy.observability.svc:4317`
- ✅ LambdaAgent operator sets `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable (see `lambdaagent_controller.go:460`)
- ✅ Traces route: Agent → Alloy → Tempo

**LambdaAgent Operator Configuration:**
```go
// From lambdaagent_controller.go:458-462
if agent.Spec.Observability != nil && agent.Spec.Observability.Tracing != nil {
    if agent.Spec.Observability.Tracing.Enabled {
        env = append(env, corev1.EnvVar{
            Name: "OTEL_EXPORTER_OTLP_ENDPOINT", 
            Value: agent.Spec.Observability.Tracing.Endpoint,
        })
    }
}
```

**Agent Configuration Example:**
```yaml
# From agent-bruno/k8s/kustomize/base/lambdaagent.yaml
observability:
  tracing:
    enabled: true
    endpoint: alloy.observability.svc:4317
  metrics:
    enabled: true
    exemplars: true
  logging:
    level: info
    format: json
    traceContext: true
```

### 3. OpenTelemetry Initialization Patterns

**Pattern 1: Manual Initialization (agent-restaurant)**
```python
# agent-restaurant/src/restaurant_agent/main.py:48-69
OTEL_EXPORTER_OTLP_ENDPOINT = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
OTEL_SERVICE_NAME = os.getenv("OTEL_SERVICE_NAME", f"agent-restaurant-{os.getenv('AGENT_ROLE', 'unknown')}")

if OTEL_EXPORTER_OTLP_ENDPOINT:
    try:
        from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
        resource = Resource.create({
            ResourceAttributes.SERVICE_NAME: OTEL_SERVICE_NAME,
            ResourceAttributes.SERVICE_VERSION: os.getenv("VERSION", "0.1.0"),
        })
        provider = TracerProvider(resource=resource)
        processor = BatchSpanProcessor(OTLPSpanExporter(endpoint=OTEL_EXPORTER_OTLP_ENDPOINT))
        provider.add_span_processor(processor)
        trace.set_tracer_provider(provider)
    except ImportError:
        pass
```

**Pattern 2: Shared Library Utilities (most agents)**
```python
# shared-lib/agent_communication/observability.py
# Provides tracing utilities but doesn't initialize exporters
def get_tracer(name: str = "agent_communication"):
    """Get OpenTelemetry tracer."""
    if OTEL_AVAILABLE:
        return trace.get_tracer(name, "1.0.0")
    return None
```

**Pattern 3: Auto-instrumentation (some agents)**
- Agents with `opentelemetry-instrumentation-fastapi` rely on auto-instrumentation
- Requires proper environment variables to be set

**Issue:** Most agents use Pattern 2 (shared utilities) but don't initialize exporters, so traces may not be sent to Tempo.

---

## Recommendations

### 🔴 High Priority

1. **Standardize OpenTelemetry Initialization**
   - Create a shared initialization module that all agents import
   - Initialize OTLP exporters on startup if `OTEL_EXPORTER_OTLP_ENDPOINT` is set
   - Ensure all agents properly export traces to Tempo

2. **Add Pydantic Settings for Observability**
   - Create `ObservabilitySettings` class using `pydantic-settings`
   - Use type-safe configuration with validation
   - Replace environment variable parsing with pydantic models

### 🟡 Medium Priority

3. **Verify Trace Export**
   - Add health checks to verify traces are reaching Tempo
   - Add metrics for trace export success/failure
   - Monitor trace volume in Grafana

4. **Documentation**
   - Document the observability setup pattern
   - Create examples for new agents
   - Add troubleshooting guide

### 🟢 Low Priority

5. **Consolidate Observability Libraries**
   - Review shared-lib observability modules
   - Ensure consistent patterns across all agents
   - Consider creating a unified observability package

---

## Implementation Plan

### Phase 1: Create Shared Observability Initialization

**File:** `shared-lib/observability/__init__.py`

```python
"""
Shared OpenTelemetry initialization for all agents.
"""
from pydantic_settings import BaseSettings
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource
from opentelemetry.semconv.resource import ResourceAttributes
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter


class ObservabilitySettings(BaseSettings):
    """Type-safe observability configuration using pydantic."""
    otel_exporter_otlp_endpoint: str | None = None
    otel_service_name: str
    otel_service_version: str = "0.1.0"
    otel_service_namespace: str = "default"
    tracing_enabled: bool = True
    metrics_enabled: bool = True
    
    class Config:
        env_prefix = "OTEL_"
        case_sensitive = False


def initialize_observability(settings: ObservabilitySettings | None = None):
    """
    Initialize OpenTelemetry tracing and metrics.
    
    Should be called at application startup.
    """
    if settings is None:
        settings = ObservabilitySettings()
    
    if not settings.tracing_enabled or not settings.otel_exporter_otlp_endpoint:
        return
    
    try:
        resource = Resource.create({
            ResourceAttributes.SERVICE_NAME: settings.otel_service_name,
            ResourceAttributes.SERVICE_VERSION: settings.otel_service_version,
            ResourceAttributes.SERVICE_NAMESPACE: settings.otel_service_namespace,
        })
        
        provider = TracerProvider(resource=resource)
        processor = BatchSpanProcessor(
            OTLPSpanExporter(endpoint=settings.otel_exporter_otlp_endpoint)
        )
        provider.add_span_processor(processor)
        trace.set_tracer_provider(provider)
        
        logger.info(
            "observability_initialized",
            endpoint=settings.otel_exporter_otlp_endpoint,
            service_name=settings.otel_service_name,
        )
    except Exception as e:
        logger.error("observability_init_failed", error=str(e))
```

### Phase 2: Update Agents to Use Shared Initialization

**Example for agent-bruno:**

```python
# agent-bruno/src/chatbot/main.py
from shared.observability import initialize_observability, ObservabilitySettings

@asynccontextmanager
async def lifespan(app: FastAPI):
    global chatbot
    
    # Initialize observability first
    obs_settings = ObservabilitySettings(
        otel_service_name="agent-bruno",
        otel_service_namespace="agent-bruno",
    )
    initialize_observability(obs_settings)
    
    # ... rest of initialization
```

### Phase 3: Update LambdaAgent Operator

**Enhancement:** Set additional OTEL environment variables:

```go
// In lambdaagent_controller.go
if agent.Spec.Observability != nil && agent.Spec.Observability.Tracing != nil {
    if agent.Spec.Observability.Tracing.Enabled {
        env = append(env, corev1.EnvVar{
            Name: "OTEL_EXPORTER_OTLP_ENDPOINT", 
            Value: agent.Spec.Observability.Tracing.Endpoint,
        })
        env = append(env, corev1.EnvVar{
            Name: "OTEL_SERVICE_NAME", 
            Value: agent.Name,
        })
        env = append(env, corev1.EnvVar{
            Name: "OTEL_SERVICE_NAMESPACE", 
            Value: agent.Namespace,
        })
        env = append(env, corev1.EnvVar{
            Name: "OTEL_TRACING_ENABLED", 
            Value: "true",
        })
    }
}
```

---

## Verification Checklist

For each agent, verify:

- [ ] Uses `pydantic-settings` for observability configuration
- [ ] Initializes OpenTelemetry exporters on startup
- [ ] Has `observability.tracing.endpoint` configured in LambdaAgent YAML
- [ ] Traces appear in Grafana Tempo
- [ ] Service name and version are correctly set in traces
- [ ] Trace context propagates across agent boundaries

---

## Summary

**Current State:**
- ✅ Most agents configured for Tempo tracing via Alloy
- ✅ OpenTelemetry libraries installed
- ❌ No pydantic-settings for observability config
- ❌ Inconsistent OpenTelemetry initialization
- ⚠️ Some agents may not be exporting traces

**Next Steps:**
1. Create shared observability initialization module with pydantic-settings
2. Update all agents to use shared initialization
3. Verify traces in Grafana Tempo
4. Add monitoring for trace export health

---

**Review Completed:** 2025-01-15


