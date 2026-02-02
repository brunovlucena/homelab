# Agent Refactoring Summary - Shared Business Logic & CloudEvents Tracing

**Date:** December 10, 2025  
**Purpose:** Refactor all agents to use shared business logic between API endpoints and CloudEvent handlers, with OpenTelemetry tracing for observability.

---

## 🎯 Refactoring Goals

1. **Eliminate Code Duplication**: API endpoints and CloudEvent handlers now use the same business logic functions
2. **Add Tracing**: OpenTelemetry spans for all CloudEvent processing
3. **Consistent Architecture**: All agents follow the same pattern

---

## ✅ Refactored Agents

### 1. agent-bruno ✅

**Changes:**
- CloudEvent handler (`/events`) now uses same `chatbot.chat()` method as API endpoint (`/chat`)
- Added OpenTelemetry tracing spans for CloudEvent processing
- Added `POST /` endpoint that routes to CloudEvent handler (for Knative compatibility)
- Tracing attributes: `cloudevent.type`, `cloudevent.source`, `cloudevent.id`, `cloudevent.duration_ms`

**Shared Business Logic:**
- `ChatBot.chat()` - Used by both `/chat` API and `/events` CloudEvent handler

**Files Modified:**
- `src/chatbot/main.py`

---

### 2. agent-redteam ✅

**Changes:**
- CloudEvent handler (`POST /`) now uses same `runner.run_exploit()` method as API endpoint (`/exploit/run`)
- CloudEvent handler uses same `runner.cleanup()` method as API endpoint
- CloudEvent handler uses same `runner.create_k6_testrun()` method
- Added OpenTelemetry tracing spans with detailed attributes:
  - `exploit.id`, `exploit.status`, `exploit.matched`, `exploit.random`
  - `vulnerability.id`, `vulnerability.type`
  - `k6.test_type`, `k6.success`
  - `cloudevent.duration_ms`

**Shared Business Logic:**
- `ExploitRunner.run_exploit()` - Used by both API and CloudEvent handlers
- `ExploitRunner.cleanup()` - Used by both API and CloudEvent handlers
- `ExploitRunner.create_k6_testrun()` - Used by both API and CloudEvent handlers

**Files Modified:**
- `src/exploit_runner/main.py`

---

### 3. agent-blueteam ✅

**Changes:**
- CloudEvent handler (`POST /`) now uses same business logic methods as API endpoints:
  - `defense_runner.handle_game_event()` - For game/MAG7 events
  - `defense_runner.analyze_threat()` - For exploit events
  - `defense_runner.execute_defense()` - For defense execution
  - `defense_runner.attack_mag7()` - For MAG7 damage
- Added OpenTelemetry tracing spans with detailed attributes:
  - `event.category` (game, mag7, exploit)
  - `threat.level`, `threat.confidence`, `threat.signature`
  - `defense.action`, `defense.success`
  - `mag7.damage`
  - `cloudevent.duration_ms`

**Shared Business Logic:**
- `DefenseRunner.handle_game_event()` - Used by both API and CloudEvent handlers
- `DefenseRunner.analyze_threat()` - Used by both API and CloudEvent handlers
- `DefenseRunner.execute_defense()` - Used by both API and CloudEvent handlers
- `DefenseRunner.attack_mag7()` - Used by both API and CloudEvent handlers

**Files Modified:**
- `src/defense_runner/main.py`

---

### 4. agent-contracts/vuln-scanner ✅

**Changes:**
- CloudEvent handler (`POST /`) now uses same `scanner.scan()` method as API endpoint (`/scan`)
- Extracted `handle_contract_scan()` as shared function used by both
- Added OpenTelemetry tracing spans with detailed attributes:
  - `contract.chain`, `contract.address`, `contract.name`
  - `scan.vulnerabilities_found`, `scan.max_severity`
  - `scan.duration_seconds`, `scan.analyzers`
  - `cloudevent.sent_to_sink`, `cloudevent.sink`
  - `cloudevent.duration_ms`

**Shared Business Logic:**
- `handle_contract_scan()` - Used by both API endpoint and CloudEvent handler
- `VulnerabilityScanner.scan()` - Core business logic

**Files Modified:**
- `src/vuln_scanner/main.py`

---

### 5. agent-contracts/exploit-generator ✅

**Changes:**
- CloudEvent handler (`POST /`) now uses same `generator.generate()` method as would be used by API endpoint
- Extracted `handle_exploit_generation()` as shared function
- Added OpenTelemetry tracing spans with detailed attributes:
  - `contract.chain`, `contract.address`
  - `vulnerabilities.count`, `vulnerabilities.high_severity_count`
  - `exploit.{n}.status`, `exploit.{n}.vuln_type`
  - `exploits.generated`, `exploits.validated`
  - `cloudevent.sent_to_sink`, `cloudevent.sink`
  - `cloudevent.duration_ms`

**Shared Business Logic:**
- `handle_exploit_generation()` - Used by both API endpoint and CloudEvent handler
- `ExploitGenerator.generate()` - Core business logic

**Files Modified:**
- `src/exploit_generator/main.py`

---

### 6. agent-tools ✅

**Changes:**
- CloudEvent handler (`POST /`) already uses shared `handle()` function
- Added OpenTelemetry tracing spans with detailed attributes:
  - `operation.status`, `operation.name`, `operation.resource`
  - `operation.duration_ms`
  - `cloudevent.duration_ms`

**Shared Business Logic:**
- `handle()` - Already shared between API and CloudEvent handlers

**Files Modified:**
- `src/k8s_tools/main.py`

---

### 7. agent-restaurant ✅

**Changes:**
- Extracted `process_restaurant_request()` as shared function
- CloudEvent handler and API endpoint both use this shared function
- Added OpenTelemetry tracing spans with detailed attributes:
  - `agent.name`, `agent.role`
  - `cloudevent.duration_ms`

**Shared Business Logic:**
- `process_restaurant_request()` - Used by both CloudEvent handler and API endpoint
- `call_ollama()` - Core LLM interaction

**Files Modified:**
- `src/restaurant_agent/main.py`

---

## 📊 Tracing Architecture

### OpenTelemetry Integration

All CloudEvent handlers now create tracing spans with:

**Standard Attributes:**
- `cloudevent.type` - Event type (e.g., `io.homelab.vuln.found`)
- `cloudevent.source` - Event source
- `cloudevent.id` - Event ID
- `cloudevent.duration_ms` - Processing duration

**Agent-Specific Attributes:**
- **agent-redteam**: `exploit.id`, `exploit.status`, `vulnerability.type`
- **agent-blueteam**: `threat.level`, `defense.action`, `mag7.damage`
- **agent-contracts**: `contract.address`, `scan.vulnerabilities_found`, `exploits.validated`
- **agent-tools**: `operation.name`, `operation.resource`, `operation.status`
- **agent-restaurant**: `agent.name`, `agent.role`

### Span Naming Convention

Spans are named using the pattern:
```
cloudevent.{event_type}
```

Where `event_type` has dots and colons replaced with underscores:
- `io.homelab.vuln.found` → `cloudevent.io_homelab_vuln_found`
- `io.knative.lambda.lifecycle.function.ready` → `cloudevent.io_knative_lambda_lifecycle_function_ready`

---

## 🔄 Architecture Pattern

### Before Refactoring

```
API Endpoint → Business Logic Function A
CloudEvent Handler → Business Logic Function B (duplicate!)
```

### After Refactoring

```
API Endpoint → Shared Business Logic Function
CloudEvent Handler → Parse Event → Shared Business Logic Function
                      ↓
                  Add Tracing Span
```

### Example: agent-redteam

**Before:**
```python
# API endpoint
@app.post("/exploit/run")
async def run_exploit(request):
    result = await runner.run_exploit(...)  # Business logic

# CloudEvent handler
@app.post("/")
async def handle_cloudevent(request):
    result = await runner.run_exploit(...)  # Same logic, but separate
```

**After:**
```python
# API endpoint
@app.post("/exploit/run")
async def run_exploit(request):
    result = await runner.run_exploit(...)  # Shared business logic

# CloudEvent handler
@app.post("/")
async def handle_cloudevent(request):
    with tracer.start_as_current_span(...) as span:  # Tracing
        result = await runner.run_exploit(...)  # Same shared business logic
        span.set_attribute("exploit.status", result.status.value)
```

---

## 📝 Benefits

1. **Code Reusability**: No duplication between API and CloudEvent handlers
2. **Consistency**: Same behavior regardless of entry point
3. **Observability**: Full tracing of CloudEvent processing
4. **Maintainability**: Single source of truth for business logic
5. **Testing**: Easier to test shared functions

---

## 🔧 Dependencies

All agents now require:
- `opentelemetry` package for tracing
- OpenTelemetry SDK configured in the cluster

**Installation:**
```python
# Add to requirements.txt or pyproject.toml
opentelemetry-api>=1.20.0
opentelemetry-sdk>=1.20.0
opentelemetry-instrumentation-fastapi>=0.42b0  # For FastAPI agents
```

---

## 🚀 Next Steps

1. **Add OpenTelemetry SDK initialization** to each agent's startup
2. **Configure trace exporters** (Jaeger, Tempo, etc.)
3. **Add span context propagation** for distributed tracing
4. **Create Grafana dashboards** for CloudEvent tracing metrics
5. **Add correlation IDs** to link API requests with CloudEvents

---

## 📚 References

- [OpenTelemetry Python SDK](https://opentelemetry.io/docs/instrumentation/python/)
- [CloudEvents Specification](https://cloudevents.io/)
- [Knative Eventing](https://knative.dev/docs/eventing/)

---

*Refactoring completed: December 10, 2025*
