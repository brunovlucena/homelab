# 📨 BACKEND-001: CloudEvents Processing

**Linear URL**: https://linear.app/bvlucena/issue/BVL-196/backend-001-cloudevents-processing  

---

## 📋 User Story

**As an** SRE Engineer  
**I want** agent-sre to receive and process CloudEvents from prometheus-events  
**So that** alerts are automatically converted to remediation actions


---


## 🎯 Acceptance Criteria

> **Note**: Features are already implemented. This ticket focuses on **validation** to ensure correctness, reliability, and production readiness.


- [ ] Agent-sre receives CloudEvents via HTTP POST endpoint
- [ ] Supports both structured and binary CloudEvents content modes
- [ ] Extracts alert information from CloudEvent data
- [ ] Validates CloudEvent format (specversion, type, source, id)
- [ ] Handles correlation IDs for request tracing
- [ ] Logs all received CloudEvents with full context
- [ ] Rate limiting to prevent overload
- [ ] Error handling for malformed events
- [ ] OpenTelemetry tracing integration

---

## 🔄 Complete Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│              CLOUDEVENTS PROCESSING WORKFLOW                          │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ⏱️  t=0s: PROMETHEUS-EVENTS SENDS CLOUDEVENT                        │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  HTTP POST: http://agent-sre.ai.svc.cluster.local/   │            │
│  │  Headers:                                             │            │
│  │    Content-Type: application/cloudevents+json         │            │
│  │    Ce-Specversion: 1.0                               │            │
│  │    Ce-Type: io.homelab.prometheus.alert.fired        │            │
│  │    Ce-Source: prometheus-events                       │            │
│  │    Ce-Id: alert-12345                                 │            │
│  │    X-Correlation-ID: abc-123                         │            │
│  │  Body: {                                              │            │
│  │    "type": "io.homelab.prometheus.alert.fired",       │            │
│  │    "source": "prometheus-events",                     │            │
│  │    "id": "alert-12345",                               │            │
│  │    "time": "2026-01-15T10:45:00Z",                    │            │
│  │    "data": {                                          │            │
│  │      "alertname": "PodCPUHigh",                       │            │
│  │      "status": "firing",                              │            │
│  │      "labels": {...},                                 │            │
│  │      "annotations": {...}                             │            │
│  │    }                                                   │            │
│  │  }                                                    │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=1ms: AGENT-SRE RECEIVES REQUEST                                │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  @app.post("/")                                      │            │
│  │  async def handle_cloudevent(request: Request):      │            │
│  │      # Extract correlation ID                        │            │
│  │      correlation_id = get_correlation_id(request)    │            │
│  │                                                      │            │
│  │      # Start OpenTelemetry span                      │            │
│  │      with tracer.start_as_current_span("cloudevent"): │            │
│  │          # Parse CloudEvent                          │            │
│  │          event = parse_cloudevent(request)           │            │
│  │                                                      │            │
│  │          # Process event                             │            │
│  │          await process_alert(event)                   │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=2ms: PARSE CLOUDEVENT                                         │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Parse CloudEvent:                                   │            │
│  │  - Check Content-Type                                │            │
│  │  - Handle structured mode (application/cloudevents+json)│            │
│  │  - Handle binary mode (Ce-* headers)                 │            │
│  │  - Extract: id, type, source, time, data             │            │
│  │  - Validate: specversion, required fields            │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=5ms: EXTRACT ALERT DATA                                       │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Extract alert information:                          │            │
│  │  - alertname: "PodCPUHigh"                           │            │
│  │  - labels: {pod, namespace, severity}               │            │
│  │  - annotations: {summary, description, lambda_function}│            │
│  │  - status: "firing"                                  │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=10ms: PROCESS ALERT                                            │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Process alert via LangGraph workflow:               │            │
│  │  1. Create Linear issue                               │            │
│  │  2. Select remediation                                │            │
│  │  3. Execute remediation                               │            │
│  │  4. Verify remediation                                │            │
│  └──────────────────────────────────────────────────────┘            │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Implementation Details

### CloudEvent Handler

```python
# src/sre_agent/main.py
from fastapi import FastAPI, Request, JSONResponse
from cloudevents.http import from_http
import json

app = FastAPI()

@app.post("/")
async def handle_cloudevent(request: Request):
    """Handle CloudEvents from prometheus-events."""
    global agent, flux_reconciler, lambda_caller
    
    if agent is None:
        return JSONResponse(
            status_code=503,
            content={"error": "Agent not initialized"}
        )
    
    # Extract correlation ID
    headers = dict(request.headers)
    correlation_id = get_correlation_id(headers=headers)
    
    # Start OpenTelemetry span
    tracer = get_tracer()
    with tracer.start_as_current_span("cloudevent.handle") as span:
        try:
            return await _process_cloudevent(request, headers, correlation_id, span)
        except Exception as e:
            span.record_exception(e)
            logger.error(
                "cloudevent_processing_failed",
                error=str(e),
                correlation_id=correlation_id,
                exc_info=True
            )
            return JSONResponse(
                status_code=500,
                content={"error": str(e)}
            )

async def _process_cloudevent(
    request: Request,
    headers: Dict[str, str],
    correlation_id: str,
    span: Optional[Any]
) -> JSONResponse:
    """Process CloudEvent with optional OpenTelemetry span."""
    
    # Parse CloudEvent
    body = await request.body()
    content_type = headers.get("content-type", "").lower()
    
    # Handle structured content mode
    if "application/cloudevents+json" in content_type:
        try:
            event_dict = json.loads(body)
            event_id = event_dict.get("id")
            event_type = event_dict.get("type")
            event_source = event_dict.get("source")
            event_data = event_dict.get("data", {})
        except json.JSONDecodeError as e:
            logger.error("failed_to_parse_cloudevent_json", error=str(e))
            return JSONResponse(
                status_code=400,
                content={"error": "Invalid CloudEvent JSON"}
            )
    else:
        # Handle binary content mode
        event = from_http(headers, body)
        event_id = event.get("id")
        event_type = event.get("type")
        event_source = event.get("source")
        event_data = event.get("data", {})
    
    # Log received CloudEvent
    logger.info(
        "cloudevent_received",
        event_id=event_id,
        event_type=event_type,
        event_source=event_source,
        correlation_id=correlation_id
    )
    
    # Process alert via LangGraph workflow
    await agent.process_alert(event_data, correlation_id)
    
    return JSONResponse(
        status_code=200,
        content={"status": "processed", "event_id": event_id}
    )
```

---

## 📚 Related Documentation

- [CloudEvents Specification](https://cloudevents.io/)
- [Agent-SRE Architecture](../../docs/architecture/agent-sre-architecture.md)

---

**Related Stories**:
- [SRE-013: Schema Evolution](./BVL-57-SRE-013-schema-evolution-compatibility.md)
- [BACKEND-002: Build Context Management](./BVL-60-BACKEND-002-build-context-management.md)



---

**Last Updated**: January 08, 2026
**Owner**: SRE Team
**Status**: Validation Required