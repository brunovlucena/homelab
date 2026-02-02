# 🌐 BACKEND-001: CloudEvents HTTP Processing

**Priority**: P0 | **Status**: ✅ Implemented  | **Story Points**: 8  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-214/backend-001-cloudevents-http-processing

---

## 📋 User Story

**As a** Backend Developer  
**I want to** process incoming CloudEvents via HTTP endpoints  
**So that** I can trigger builds, services, and other lambda operations through a standardized event interface

---

## 🏷️ Event type naming

- **Canonical (current)**: Event types use the `io.knative.lambda.*` namespace (e.g. `io.knative.lambda.command.build.start`). These are the supported types.
- **Legacy**: Event types under `network.notifi.lambda.*` (e.g. `network.notifi.lambda.build.start`, `network.notifi.lambda.service.delete`) are **legacy** and should not be used for new integrations. The receiver does not route by legacy type names; use `io.knative.lambda.*` instead.

---

## 🎯 Acceptance Criteria

### ✅ HTTP Endpoint Requirements
- [x] HTTP server listens on configurable port (default: 8080)
- [ ] POST `/events` endpoint accepts CloudEvents (optional; Knative Trigger uses `/`)
- [x] POST `/` root endpoint accepts CloudEvents (Knative Service)
- [x] GET `/health` returns 200 OK for Knative queue-proxy
- [x] All endpoints support graceful shutdown
- [ ] Request/response headers include correlation IDs for tracing

### ✅ CloudEvent Validation
- [x] Validate CloudEvents v1.0 specification compliance
- [x] Required attributes: type, source, id (Ce-Type, Ce-Source, Ce-Id or structured body)
- [x] Parse structured CloudEvent JSON body
- [x] Support both binary and structured content modes
- [x] Return 400 Bad Request for invalid events
- [x] Optional schema validation against registered event types

### ✅ Event Type Routing (canonical: `io.knative.lambda.*`)

**Command events**
- [x] Route `io.knative.lambda.command.function.deploy` to function deploy (create/update LambdaFunction)
- [x] Route `io.knative.lambda.command.service.create` / `.update` to deploy (aliases)
- [x] Route `io.knative.lambda.command.service.delete` to service deletion
- [x] Route `io.knative.lambda.command.build.start` (and `.retry`) to build handler
- [x] Route `io.knative.lambda.command.build.cancel` to build cancel
- [x] Route `io.knative.lambda.command.function.rollback` to rollback handler

**Response events (RED metrics)**
- [x] Route `io.knative.lambda.response.success` to success metrics handler
- [x] Route `io.knative.lambda.response.error` to error metrics handler

Unsupported event types are ignored (202); consider 400 for strict validation.

### ✅ Observability
- [ ] Generate correlation ID for each request
- [x] Propagate trace context from CloudEvent headers
- [x] Log all incoming requests with event metadata
- [x] Emit metrics for event processing (success/failure/duration)
- [x] Include OpenTelemetry spans for distributed tracing
- [x] Track event processing latency

---

## 🔧 Technical Implementation

**Implementation**: `internal/webhook/cloudevents_receiver.go` (CloudEvents receiver as Knative Service).

- **ReceiverConfig**: Port, Path `/`, DefaultNamespace, RateLimit, WorkerPoolSize, ProcessingTimeout, Schema validation, Debug.
- **Endpoints**: POST `/` (CloudEvents), GET `/health`, GET `/ready`, GET `/metrics`.
- **Lifecycle**: `Start(ctx)` with graceful shutdown via `server.Shutdown`.
- **Parsing**: `cehttp.NewEventFromHTTPRequest(req)`; process by `event.Type()` with schema validation when enabled.

---

## 📊 API Specification

### Request format (canonical event types)

```http
POST / HTTP/1.1
Host: lambda-command-receiver.example.com
Content-Type: application/cloudevents+json
Ce-Specversion: 1.0
Ce-Type: io.knative.lambda.command.build.start
Ce-Source: io.knative.lambda/operator
Ce-Id: 550e8400-e29b-41d4-a716-446655440000
Ce-Subject: my-function
X-Correlation-ID: correlation-123

{
  "name": "my-function",
  "namespace": "knative-lambda",
  "forceRebuild": false,
  "reason": "manual"
}
```

Function deploy example:

```http
POST / HTTP/1.1
Ce-Type: io.knative.lambda.command.function.deploy
...

{
  "metadata": { "name": "my-fn", "namespace": "knative-lambda" },
  "spec": { ... LambdaFunctionSpec ... }
}
```

### Response format
- **Success**: 202 Accepted, `{"status":"accepted","eventId":"..."}`
- **Invalid event**: 400 Bad Request
- **Rate limit**: 429 Too Many Requests, Retry-After

---

## 🧪 Testing Scenarios

### 1. Valid build start event (canonical type)
```bash
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/cloudevents+json" \
  -H "Ce-Specversion: 1.0" \
  -H "Ce-Type: io.knative.lambda.command.build.start" \
  -H "Ce-Source: io.knative.lambda/operator" \
  -H "Ce-Id: $(uuidgen)" \
  -d '{"name":"my-function","namespace":"knative-lambda"}'
```

**Expected**: 202 Accepted

### 2. Invalid CloudEvent (missing required attributes)
```bash
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{"data":"test"}'
```

**Expected**: 400 Bad Request

### 3. Health check
```bash
curl http://localhost:8080/health
```

**Expected**: 200 OK with body "ok"

---

## 📈 Performance Requirements

- **Latency**: < 100ms for event ingestion (before async processing)
- **Throughput**: Support 1000 events/second (rate limiter protects K8s API)
- **Concurrency**: Worker pool; rate limit 50 QPS, burst 100
- **Timeout**: 30s processing timeout per event
- **Resource**: < 200MB memory per pod

---

## 🔍 Monitoring & Alerts

### Metrics
- Receiver metrics endpoint: `/metrics` (events_received, events_processed, events_failed)
- Prometheus: `knative_lambda_function_*` (invocations, duration, errors, cold_starts) from response events

### Alerts
- **HTTP 5xx Rate**: Alert if > 5% over 5 minutes
- **Request Latency**: Alert if p99 > 1s
- **Event Validation Failures**: Alert if > 10% of requests

---

## 📚 Related Documentation

- CloudEvents Specification: https://cloudevents.io/
- [BACKEND-002: Build Context Management](BACKEND-002-build-context-management.md)
- [BACKEND-003: Kubernetes Job Lifecycle](BACKEND-003-kubernetes-job-lifecycle.md)
- OpenTelemetry Go SDK: https://opentelemetry.io/docs/languages/go/

---

## 🏗️ Code References

**Implementation** (knative-lambda-operator):

- `internal/webhook/cloudevents_receiver.go` – CloudEvents HTTP receiver, routing, handlers
- `internal/events/manager.go` – Event type constants (`io.knative.lambda.*`)
- `internal/schema/validator.go` – Optional schema validation

---

**Last Updated**: February 2026  
**Owner**: Backend Team  
**Status**: Production Ready
