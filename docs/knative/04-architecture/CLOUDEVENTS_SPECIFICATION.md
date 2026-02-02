# 🌐 CloudEvents Specification

**Version**: 1.0.1  
**Last Updated**: December 4, 2025  
**Status**: Living Document

---

## 📖 Table of Contents

- [Overview](#overview)
- [Event Naming Convention](#event-naming-convention)
- [Event Categories](#event-categories)
- [Event Type Catalog](#event-type-catalog)
- [Event Data Schemas](#event-data-schemas)
- [Event Flow Diagrams](#event-flow-diagrams)
- [Best Practices](#best-practices)
- [Migration Notes](#migration-notes)
- [Flux CD Integration (CDEvents)](#flux-cd-integration-cdevents)

---

## 🎯 Overview

The Knative Lambda platform uses CloudEvents as the standard format for all event-driven communication. This specification defines:

1. **Control Plane Events** - Events that manage lambda lifecycle (operator → system)
2. **Data Plane Events** - Events that invoke lambda functions (external → lambda)
3. **Notification Events** - Events emitted for observability (system → external)

### CloudEvents Version

All events comply with **CloudEvents Specification v1.0**.

### Event Flow Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         KNATIVE LAMBDA PLATFORM                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │  1️⃣ CONTROL PLANE (Commands → Operator)                                │  │
│  │                                                                         │  │
│  │  External Systems                   Operator                            │  │
│  │  ┌──────────────┐    CloudEvents    ┌───────────────────────┐          │  │
│  │  │  CI/CD       │ ───────────────── │  knative-lambda-      │          │  │
│  │  │  API         │   .command.*      │  operator             │          │  │
│  │  │  GitOps      │                   │                       │          │  │
│  │  └──────────────┘                   └───────────────────────┘          │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                  │                                           │
│                                  │ Creates/Manages                           │
│                                  ▼                                           │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │  2️⃣ DATA PLANE (Events → Lambda Functions)                             │  │
│  │                                                                         │  │
│  │  Event Sources                      Lambda Functions                    │  │
│  │  ┌──────────────┐    CloudEvents    ┌───────────────────────┐          │  │
│  │  │  RabbitMQ    │ ───────────────── │  hello-python         │          │  │
│  │  │  Broker      │   .invoke.*       │  hello-nodejs         │          │  │
│  │  │  Triggers    │   .event.*        │  ...                  │          │  │
│  │  └──────────────┘                   └───────────────────────┘          │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                  │                                           │
│                                  │ Emits                                     │
│                                  ▼                                           │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │  3️⃣ NOTIFICATION PLANE (Operator → Observers)                          │  │
│  │                                                                         │  │
│  │  Operator                           Observers                           │  │
│  │  ┌──────────────┐    CloudEvents    ┌───────────────────────┐          │  │
│  │  │  knative-    │ ───────────────── │  agent-auditor        │          │  │
│  │  │  lambda-     │   .notification.* │  Loki (audit log)     │          │  │
│  │  │  operator    │   .lifecycle.*    │  agent-contracts      │          │  │
│  │  └──────────────┘                   │  External Systems     │          │  │
│  │                                     └───────────────────────┘          │  │
│  │                                                                         │  │
│  │  ⚠️ NOTE: Alertmanager uses WEBHOOKS, not CloudEvents.                  │  │
│  │     Use notifi-adapter as bridge: CloudEvents → Alertmanager webhook   │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 📛 Event Naming Convention

### Type Format

```
io.knative.lambda.<category>.<entity>.<action>
```

| Component | Description | Examples |
|-----------|-------------|----------|
| `io.knative.lambda` | Reverse DNS prefix | Fixed for this platform |
| `<category>` | Event category | `command`, `lifecycle`, `invoke`, `notification` |
| `<entity>` | Target entity | `function`, `build`, `service` |
| `<action>` | Action (past tense for events) | `created`, `started`, `completed` |

### Source Format

```
io.knative.lambda/<component>/<namespace>/<name>
```

Example: `io.knative.lambda/operator/knative-lambda/hello-python`

### Subject Format

```
<namespace>/<resource-name>
```

Example: `knative-lambda/hello-python`

---

## 📂 Event Categories

### 1. Command Events (Imperative)

Commands are **requests** for the system to perform an action. They use present tense.

| Type Pattern | Direction | Purpose |
|--------------|-----------|---------|
| `*.command.build.*` | External → Operator | Request build operations |
| `*.command.service.*` | External → Operator | Request service operations |
| `*.command.function.*` | External → Operator | Request function operations |

### 2. Lifecycle Events (Declarative)

Lifecycle events notify about **state changes**. They use past tense.

| Type Pattern | Direction | Purpose |
|--------------|-----------|---------|
| `*.lifecycle.function.*` | Operator → Observers | Function state changes |
| `*.lifecycle.build.*` | Operator → Observers | Build state changes |
| `*.lifecycle.service.*` | Operator → Observers | Service state changes |

### 3. Invoke Events (Data)

Invoke events **trigger lambda function execution**.

| Type Pattern | Direction | Purpose |
|--------------|-----------|---------|
| `*.invoke.sync` | External → Lambda | Synchronous invocation |
| `*.invoke.async` | External → Lambda | Asynchronous invocation |
| `*.invoke.scheduled` | Scheduler → Lambda | Scheduled invocation |

### 4. Notification Events (Observability)

Notifications inform about **operational events**.

| Type Pattern | Direction | Purpose |
|--------------|-----------|---------|
| `*.notification.alert.*` | System → Alerting | Alert conditions |
| `*.notification.audit.*` | System → Audit | Audit trail |
| `*.notification.metric.*` | System → Metrics | Metric events |

---

## 📋 Event Type Catalog

### Control Plane Commands

| Event Type | Description | Idempotent | Trigger |
|------------|-------------|------------|---------|
| `io.knative.lambda.command.build.start` | Request to start a build | Yes | Creates build job |
| `io.knative.lambda.command.build.cancel` | Request to cancel a build | Yes | Cancels running job |
| `io.knative.lambda.command.build.retry` | Request to retry failed build | Yes | Retries build job |
| `io.knative.lambda.command.service.create` | Request to create service | Yes | Creates Knative Service |
| `io.knative.lambda.command.service.update` | Request to update service | Yes | Updates Knative Service |
| `io.knative.lambda.command.service.delete` | Request to delete service | Yes | Deletes Knative Service |
| `io.knative.lambda.command.function.deploy` | Request to deploy function | Yes | Full deploy workflow |
| `io.knative.lambda.command.function.rollback` | Request to rollback function | Yes | Rollback to previous |

### Lifecycle Events (Emitted by Operator)

| Event Type | Description | When Emitted |
|------------|-------------|--------------|
| **Function Lifecycle** | | |
| `io.knative.lambda.lifecycle.function.created` | Function CR created | After CR validation |
| `io.knative.lambda.lifecycle.function.updated` | Function CR updated | After spec change |
| `io.knative.lambda.lifecycle.function.deleted` | Function CR deleted | Before finalizer runs |
| `io.knative.lambda.lifecycle.function.ready` | Function is ready | All conditions met |
| `io.knative.lambda.lifecycle.function.degraded` | Function is degraded | Partial failure |
| **Build Lifecycle** | | |
| `io.knative.lambda.lifecycle.build.started` | Build job created | Job submitted to K8s |
| `io.knative.lambda.lifecycle.build.progressing` | Build in progress | Job running |
| `io.knative.lambda.lifecycle.build.completed` | Build succeeded | Job completed (exit 0) |
| `io.knative.lambda.lifecycle.build.failed` | Build failed | Job failed (exit != 0) |
| `io.knative.lambda.lifecycle.build.timeout` | Build timed out | Deadline exceeded |
| `io.knative.lambda.lifecycle.build.cancelled` | Build cancelled | User/system cancelled |
| **Service Lifecycle** | | |
| `io.knative.lambda.lifecycle.service.created` | Knative Service created | After apply |
| `io.knative.lambda.lifecycle.service.updated` | Knative Service updated | After update |
| `io.knative.lambda.lifecycle.service.deleted` | Knative Service deleted | After delete |
| `io.knative.lambda.lifecycle.service.ready` | Service ready | All revisions ready |
| `io.knative.lambda.lifecycle.service.scaled` | Service scaled | Replicas changed |

### Invoke Events (Trigger Lambda Functions)

| Event Type | Description | Response |
|------------|-------------|----------|
| `io.knative.lambda.invoke.sync` | Synchronous invocation | Wait for response |
| `io.knative.lambda.invoke.async` | Asynchronous invocation | 202 Accepted |
| `io.knative.lambda.invoke.scheduled` | Scheduled invocation | 202 Accepted |
| `io.knative.lambda.invoke.retry` | Retry invocation (from DLQ) | 202 Accepted |

### Lambda Response Events

| Event Type | Description | When Emitted |
|------------|-------------|--------------|
| `io.knative.lambda.response.success` | Function executed successfully | Successful execution |
| `io.knative.lambda.response.error` | Function execution failed | Runtime error |
| `io.knative.lambda.response.timeout` | Function timed out | Deadline exceeded |

### Notification Events

| Event Type | Description | Severity |
|------------|-------------|----------|
| `io.knative.lambda.notification.alert.critical` | Critical system alert | Critical |
| `io.knative.lambda.notification.alert.warning` | Warning alert | Warning |
| `io.knative.lambda.notification.alert.info` | Informational alert | Info |
| `io.knative.lambda.notification.audit.access` | Access audit event | - |
| `io.knative.lambda.notification.audit.change` | Change audit event | - |

---

## 📊 Event Data Schemas

### Base Event Structure

```json
{
  "specversion": "1.0",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "source": "io.knative.lambda/operator/knative-lambda/hello-python",
  "type": "io.knative.lambda.lifecycle.build.completed",
  "subject": "knative-lambda/hello-python",
  "time": "2025-12-04T10:30:00Z",
  "datacontenttype": "application/json",
  "data": { ... }
}
```

### Function Event Data

```json
{
  "name": "hello-python",
  "namespace": "knative-lambda",
  "runtime": {
    "language": "python",
    "version": "3.11",
    "handler": "main.handler"
  },
  "phase": "Ready",
  "conditions": [
    {
      "type": "Ready",
      "status": "True",
      "reason": "AllConditionsMet",
      "lastTransitionTime": "2025-12-04T10:30:00Z"
    }
  ],
  "generation": 1,
  "observedGeneration": 1
}
```

### Build Event Data

```json
{
  "name": "hello-python",
  "namespace": "knative-lambda",
  "jobName": "kaniko-hello-python-abc123",
  "buildId": "build-abc123",
  "imageUri": "localhost:5001/hello-python:v1.0.0",
  "digest": "sha256:abc123...",
  "startedAt": "2025-12-04T10:25:00Z",
  "completedAt": "2025-12-04T10:30:00Z",
  "duration": "5m0s",
  "phase": "Completed",
  "error": null,
  "logs": "https://grafana.example.com/explore?..."
}
```

### Service Event Data

```json
{
  "name": "hello-python",
  "namespace": "knative-lambda",
  "serviceName": "hello-python",
  "url": "http://hello-python.knative-lambda.svc.cluster.local",
  "latestRevision": "hello-python-00003",
  "latestReadyRevision": "hello-python-00003",
  "ready": true,
  "replicas": {
    "desired": 2,
    "ready": 2,
    "available": 2
  },
  "traffic": [
    {
      "revisionName": "hello-python-00003",
      "percent": 100,
      "latestRevision": true
    }
  ]
}
```

### Invoke Event Data

```json
{
  "functionName": "hello-python",
  "namespace": "knative-lambda",
  "invocationId": "inv-abc123",
  "correlationId": "corr-xyz789",
  "payload": {
    "message": "Hello, World!",
    "timestamp": "2025-12-04T10:30:00Z"
  },
  "metadata": {
    "traceId": "abc123...",
    "spanId": "def456...",
    "retryCount": 0,
    "deadlineAt": "2025-12-04T10:31:00Z"
  }
}
```

### Response Event Data

```json
{
  "functionName": "hello-python",
  "namespace": "knative-lambda",
  "invocationId": "inv-abc123",
  "correlationId": "corr-xyz789",
  "result": {
    "statusCode": 200,
    "body": {
      "message": "Hello from Python!",
      "timestamp": "2025-12-04T10:30:01Z"
    }
  },
  "metrics": {
    "durationMs": 150,
    "coldStart": false,
    "memoryUsedMb": 64
  }
}
```

### Error Event Data

```json
{
  "functionName": "hello-python",
  "namespace": "knative-lambda",
  "invocationId": "inv-abc123",
  "correlationId": "corr-xyz789",
  "error": {
    "type": "RuntimeError",
    "message": "Division by zero",
    "code": "RUNTIME_ERROR",
    "retryable": false,
    "stackTrace": "..."
  },
  "dlq": {
    "routed": true,
    "queueName": "hello-python-dlq",
    "routedAt": "2025-12-04T10:30:01Z"
  }
}
```

---

## 🔄 Event Flow Diagrams

### Build Flow

```
┌──────────┐    command.build.start     ┌──────────┐
│ External │ ─────────────────────────► │ Operator │
└──────────┘                            └────┬─────┘
                                             │
      ┌──────────────────────────────────────┤
      │                                      │
      ▼                                      ▼
┌─────────────────┐                   ┌─────────────────┐
│ lifecycle.build │                   │  Create Kaniko  │
│    .started     │◄──────────────────│     Job         │
└─────────────────┘                   └────────┬────────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │  Kaniko Job │
                                        │   Running   │
                                        └──────┬──────┘
                                               │
                    ┌──────────────────────────┼──────────────────────────┐
                    ▼                          ▼                          ▼
          ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
          │ lifecycle.build │       │ lifecycle.build │       │ lifecycle.build │
          │   .completed    │       │     .failed     │       │    .timeout     │
          └─────────────────┘       └─────────────────┘       └─────────────────┘
```

### Invocation Flow

```
┌──────────┐      invoke.sync         ┌──────────────┐
│ External │ ────────────────────────►│   RabbitMQ   │
└──────────┘                          │    Broker    │
                                      └──────┬───────┘
                                             │ Trigger (CloudEvent)
                                             ▼
                                      ┌──────────────────────────────────┐
                                      │        Lambda Function           │
                                      │  ┌────────────────────────────┐  │
                                      │  │    Lambda Runtime Wrapper   │  │
                                      │  │  ┌──────────────────────┐  │  │
                                      │  │  │   User Handler Code   │  │  │
                                      │  │  └──────────────────────┘  │  │
                                      │  │  • Parse CloudEvent        │  │
                                      │  │  • Execute handler         │  │
                                      │  │  • Measure duration        │  │
                                      │  │  • Catch errors/timeouts   │  │
                                      │  │  • Emit response CloudEvent│  │
                                      │  └────────────────────────────┘  │
                                      └──────────────┬───────────────────┘
                                                     │ Reply CloudEvent
                    ┌────────────────────────────────┼────────────────────────────────┐
                    ▼                                ▼                                ▼
          ┌─────────────────┐              ┌─────────────────┐              ┌─────────────────┐
          │ response.success│              │  response.error │              │ response.timeout│
          └─────────────────┘              └────────┬────────┘              └─────────────────┘
                                                    │
                                                    ▼
                                           ┌─────────────────┐
                                           │      DLQ        │
                                           │  (if retryable) │
                                           └─────────────────┘
```

#### Response Event Generation

Response CloudEvents (`response.success`, `response.error`, `response.timeout`) are generated by the **Lambda Runtime Wrapper** - a framework-injected layer that wraps user handler code. This wrapper:

1. **Receives** incoming CloudEvents via HTTP POST
2. **Parses** the CloudEvent headers and body
3. **Executes** the user's handler function with the event data
4. **Measures** execution duration and detects cold starts
5. **Catches** errors and timeouts
6. **Emits** a response CloudEvent as the HTTP reply

The response CloudEvent is returned as the HTTP response body with appropriate `Ce-*` headers. Knative Eventing automatically routes this reply back to the broker for downstream processing.

```
                       HTTP Request                         HTTP Response
                       ┌─────────────────┐                 ┌─────────────────┐
                       │ Ce-Type: invoke │                 │ Ce-Type: resp.  │
     Broker ──────────►│ Ce-Source: ...  │     Runtime    │   success       │──────────► Broker
                       │ Content-Type:   │  ───────────►  │ Ce-Source: ...  │
                       │ application/json│                 │ {"result": ...} │
                       └─────────────────┘                 └─────────────────┘
```

---

## 🔧 Lambda Runtime Wrapper

The Lambda Runtime Wrapper is injected into every Lambda function container during the build process. It provides CloudEvents-compliant request/response handling while allowing developers to write simple handler functions.

### Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         LAMBDA CONTAINER                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │                      LAMBDA RUNTIME WRAPPER                            │  │
│  │                                                                         │  │
│  │  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐               │  │
│  │  │ HTTP Server  │   │ CloudEvent   │   │   Response   │               │  │
│  │  │   :8080      │──►│   Parser     │──►│   Builder    │               │  │
│  │  └──────────────┘   └──────────────┘   └──────────────┘               │  │
│  │         │                  │                   │                       │  │
│  │         │                  ▼                   │                       │  │
│  │         │           ┌──────────────┐          │                       │  │
│  │         │           │   Handler    │          │                       │  │
│  │         │           │   Invoker    │──────────┤                       │  │
│  │         │           └──────────────┘          │                       │  │
│  │         │                  │                   │                       │  │
│  │         │                  ▼                   │                       │  │
│  │         │    ┌───────────────────────────┐    │                       │  │
│  │         │    │  USER HANDLER FUNCTION    │    │                       │  │
│  │         │    │  handler(event) → result  │    │                       │  │
│  │         │    └───────────────────────────┘    │                       │  │
│  │         │                                      │                       │  │
│  │         ▼                                      ▼                       │  │
│  │   ┌─────────────────────────────────────────────────┐                 │  │
│  │   │              METRICS & OBSERVABILITY             │                 │  │
│  │   │  • Duration tracking    • Cold start detection   │                 │  │
│  │   │  • Memory usage         • Error classification   │                 │  │
│  │   │  • Trace propagation    • Correlation IDs        │                 │  │
│  │   └─────────────────────────────────────────────────┘                 │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Supported Runtimes

| Runtime | Wrapper Entry Point | Handler Interface |
|---------|--------------------|--------------------|
| Python 3.x | `runtime.py` | `def handler(event: dict) -> dict` |
| Node.js 18+ | `runtime.js` | `exports.handler = async (event) => result` |
| Go 1.21+ | `runtime.go` | `func Handler(event map[string]interface{}) (map[string]interface{}, error)` |

### Handler Contract

User handlers receive the CloudEvent `data` payload and return a result:

**Python:**
```python
def handler(event: dict) -> dict:
    """
    Args:
        event: The CloudEvent data payload (parsed JSON)
    
    Returns:
        dict: Result to be included in response.success CloudEvent
    
    Raises:
        Exception: Any exception triggers response.error CloudEvent
    """
    return {"message": "Hello!", "processed": True}
```

**Node.js:**
```javascript
exports.handler = async (event) => {
    // event: CloudEvent data payload (parsed JSON)
    // return: Result for response.success CloudEvent
    // throw: Triggers response.error CloudEvent
    return { message: "Hello!", processed: true };
};
```

**Go:**
```go
func Handler(event map[string]interface{}) (map[string]interface{}, error) {
    // event: CloudEvent data payload (parsed JSON)
    // return result: For response.success CloudEvent
    // return error: Triggers response.error CloudEvent
    return map[string]interface{}{"message": "Hello!", "processed": true}, nil
}
```

### Environment Variables

The runtime wrapper uses these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port (set by Knative) | `8080` |
| `HANDLER` | Handler function path | `handler` |
| `FUNCTION_NAME` | Lambda function name | Required |
| `FUNCTION_NAMESPACE` | Kubernetes namespace | Required |
| `BROKER_URL` | CloudEvents broker URL (optional) | - |
| `TIMEOUT_SECONDS` | Handler execution timeout | `60` |
| `LOG_LEVEL` | Logging verbosity | `info` |

### Timeout Handling

The runtime wrapper enforces execution timeouts:

1. **Handler Timeout**: If handler execution exceeds `TIMEOUT_SECONDS`, the wrapper:
   - Terminates the handler (where possible)
   - Emits `response.timeout` CloudEvent
   - Returns HTTP 504 Gateway Timeout

2. **Graceful Shutdown**: On SIGTERM/SIGINT:
   - Completes in-flight requests (up to 30s)
   - Emits any pending response events
   - Exits cleanly

### Error Classification

Errors are classified to determine retry behavior:

| Error Type | Code | Retryable | DLQ Routing |
|------------|------|-----------|-------------|
| `RuntimeError` | `RUNTIME_ERROR` | No | Yes |
| `TimeoutError` | `TIMEOUT_ERROR` | Yes | After max retries |
| `ValidationError` | `VALIDATION_ERROR` | No | Yes |
| `TransientError` | `TRANSIENT_ERROR` | Yes | After max retries |
| `SystemError` | `SYSTEM_ERROR` | Yes | After max retries |

---

## ✅ Best Practices

### 1. Idempotency

All commands MUST be idempotent. Include an `idempotencyKey` extension:

```json
{
  "specversion": "1.0",
  "type": "io.knative.lambda.command.build.start",
  "idempotencykey": "build-hello-python-gen-5"
}
```

### 2. Correlation

Always propagate correlation IDs across the event chain:

```json
{
  "correlationid": "corr-xyz789",
  "causationid": "evt-abc123"
}
```

### 3. Ordering

For events requiring ordering, use a sequence number:

```json
{
  "sequencetype": "Integer",
  "sequencevalue": "42"
}
```

### 4. Error Handling

- **Transient errors**: Retry with exponential backoff
- **Permanent errors**: Route to DLQ
- **Poison messages**: Log and alert, do not retry

### 5. Schema Evolution

- Add new fields as optional
- Never remove required fields
- Use `dataschema` extension for versioning:

```json
{
  "dataschema": "https://knative-lambda.io/schemas/build-event/v1.0.0"
}
```

### 6. Security

- Sanitize all event data (no secrets, credentials)
- Use HMAC signatures for external events
- Validate event sources

---

## 🔗 Notifi Fusion Integration

Knative Lambda integrates with the **Notifi Fusion Platform** for blockchain transaction parsing. The Scheduler service orchestrates Lambda function lifecycle and execution.

### Event Flow with Notifi

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    NOTIFI FUSION → KNATIVE LAMBDA FLOW                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1️⃣ BUILD TRIGGER (Scheduler → Operator)                                    │
│     ┌──────────────┐    CloudEvent    ┌──────────────┐    ┌──────────────┐  │
│     │   Scheduler  │ ───────────────► │   RabbitMQ   │ ──►│   Operator   │  │
│     │              │  build.start     │  Exchange:   │    │              │  │
│     │              │                  │  cloud-events│    │              │  │
│     └──────────────┘                  └──────────────┘    └──────────────┘  │
│                                                                              │
│  2️⃣ PARSER EXECUTION (Scheduler → Lambda)                                   │
│     ┌──────────────┐    CloudEvent    ┌──────────────┐    ┌──────────────┐  │
│     │   Scheduler  │ ───────────────► │   RabbitMQ   │ ──►│   Lambda     │  │
│     │              │  parser.start    │   Broker     │    │   Function   │  │
│     └──────────────┘                  └──────────────┘    └──────────────┘  │
│                                                                              │
│  3️⃣ EXECUTION CALLBACK (Lambda → Scheduler)                                 │
│     ┌──────────────┐      HTTP POST       ┌──────────────┐                  │
│     │   Lambda     │ ────────────────────►│   Scheduler  │                  │
│     │   Function   │  /fusion/execution/  │   :5000      │                  │
│     │              │      response        │              │                  │
│     └──────────────┘                      └──────────────┘                  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Notifi Event Types (Legacy)

Events published by Notifi Scheduler to RabbitMQ (exchange: `cloud-events`):

| Notifi Event Type | Description | Consumer |
|-------------------|-------------|----------|
| `network.notifi.lambda.build.start` | Trigger parser build | Operator |
| `network.notifi.lambda.parser.start` | Trigger parser execution | Lambda Function |

### Scheduler Callback

Lambda functions send execution results back to Scheduler via HTTP:

```
POST http://notifi-scheduler.notifi.svc.cluster.local:5000/fusion/execution/response
Content-Type: application/json

{
  "contextId": "exec-123",
  "succeeded": true,
  "eventEntries": [...],
  "errorMessage": null
}
```

### Service Dependencies

Lambda functions can access these Notifi services via gRPC (port 4000):

| Service | Address | Purpose |
|---------|---------|---------|
| **Storage Manager** | `notifi-storage-manager.notifi:4000` | Module retrieval |
| **Blockchain Manager** | `notifi-blockchain-manager.notifi:4000` | RPC access |
| **Fetch Proxy** | `notifi-fetch-proxy.notifi:4000` | External API calls |
| **Subscription Manager** | `notifi-subscription-manager.notifi:4000` | Alert processing |

For full Notifi integration details, see [NOTIFI_INTEGRATION.md](./NOTIFI_INTEGRATION.md).

---

## 🔄 Migration Notes

### Current Events → New Format

| Current Type | New Type |
|--------------|----------|
| `network.notifi.lambda.build.start` | `io.knative.lambda.command.build.start` |
| `network.notifi.lambda.parser.start` | `io.knative.lambda.invoke.async` |
| `lambda.build.start` | `io.knative.lambda.command.build.start` |
| `lambda.build.started` | `io.knative.lambda.lifecycle.build.started` |
| `lambda.build.complete` | `io.knative.lambda.lifecycle.build.completed` |
| `lambda.build.completed` | `io.knative.lambda.lifecycle.build.completed` |
| `lambda.build.failed` | `io.knative.lambda.lifecycle.build.failed` |
| `lambda.build.timeout` | `io.knative.lambda.lifecycle.build.timeout` |
| `lambda.build.cancelled` | `io.knative.lambda.lifecycle.build.cancelled` |
| `lambda.function.created` | `io.knative.lambda.lifecycle.function.created` |
| `lambda.function.updated` | `io.knative.lambda.lifecycle.function.updated` |
| `lambda.function.deleted` | `io.knative.lambda.lifecycle.function.deleted` |
| `lambda.service.created` | `io.knative.lambda.lifecycle.service.created` |
| `lambda.service.updated` | `io.knative.lambda.lifecycle.service.updated` |
| `lambda.service.delete` | `io.knative.lambda.command.service.delete` |
| `lambda.service.deleted` | `io.knative.lambda.lifecycle.service.deleted` |
| `lambda.parser.started` | `io.knative.lambda.invoke.async` |
| `lambda.parser.completed` | `io.knative.lambda.response.success` |
| `lambda.parser.failed` | `io.knative.lambda.response.error` |

### Tense Consistency

- **Commands**: Present tense (`.start`, `.create`, `.delete`)
- **Lifecycle**: Past tense (`.started`, `.created`, `.deleted`)
- **Response**: Past/Present (`.success`, `.error`)

---

## 👁️ Observers

CloudEvents observers are systems that subscribe to notification and lifecycle events.

### Supported Observers

| Observer | Event Types | Description |
|----------|-------------|-------------|
| **agent-auditor** | `*.notification.*`, `*.lifecycle.*` | LLM security testing and audit |
| **Loki** | `*.notification.audit.*` | Centralized audit logging |
| **agent-contracts** | `*.lifecycle.function.*` | Smart contract vulnerability scanning (4 components) |
| **Grafana** | `*.notification.alert.*` | Dashboard annotations |
| **notifi-adapter** | `*.notification.alert.*` | CloudEvents → Alertmanager webhook bridge |

### agent-contracts Integration

The `agent-contracts` pipeline consists of 4 serverless components communicating via CloudEvents:

```
┌────────────────┐  io.homelab.    ┌────────────────┐  io.homelab.    ┌────────────────┐
│ contract-fetch │ ─────────────── │  vuln-scanner  │ ─────────────── │exploit-generator│
│                │  contract.      │                │  vuln.found     │                 │
└────────────────┘  created        └────────────────┘                 └────────┬────────┘
                                                                               │
                                                         io.homelab.exploit.validated
                                                                               │
                                                                               ▼
                                                                     ┌────────────────┐
                                                                     │alert-dispatcher│
                                                                     │  (Grafana,     │
                                                                     │  Telegram, etc)│
                                                                     └────────────────┘
```

**notifi-adapter** acts as a bridge converting CloudEvents to Alertmanager webhook format.

### agent-auditor Integration

The `agent-auditor` service receives CloudEvents for security auditing of Lambda functions:

```yaml
# Trigger for agent-auditor
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: lambda-audit-trigger
  namespace: agent-auditor
spec:
  broker: lambda-broker
  filter:
    attributes:
      type: io.knative.lambda.notification.audit.change
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: agent-auditor
```

**Audited Events**:
- `io.knative.lambda.lifecycle.function.created` - New function deployed
- `io.knative.lambda.lifecycle.function.updated` - Function code/config changed
- `io.knative.lambda.notification.audit.access` - Function invocation audit
- `io.knative.lambda.notification.audit.change` - Configuration changes

### Alertmanager Integration

⚠️ **Alertmanager does NOT natively support CloudEvents**. It uses its own webhook format.

**Integration Pattern**:

```
┌──────────────────┐    CloudEvents    ┌──────────────────┐    Webhook    ┌──────────────────┐
│  knative-lambda  │ ───────────────── │  notifi-adapter  │ ────────────► │   Alertmanager   │
│  operator        │  .notification.*  │  (bridge)        │               │                  │
└──────────────────┘                   └──────────────────┘               └──────────────────┘
```

**Option 1: CloudEvents → notifi-adapter → Alertmanager**

```python
# notifi-adapter receives CloudEvents and forwards to Alertmanager
@app.route('/cloudevents', methods=['POST'])
def receive_cloudevent():
    event = from_http(request.headers, request.data)
    
    if event['type'].startswith('io.knative.lambda.notification.alert'):
        # Transform to Alertmanager format
        alert = {
            'labels': {
                'alertname': event.data['alertName'],
                'severity': event.data['severity'],
            },
            'annotations': {
                'summary': event.data['summary'],
                'description': event.data['description'],
            },
            'startsAt': event.data['startsAt'],
        }
        requests.post(f'{ALERTMANAGER_URL}/api/v1/alerts', json=[alert])
```

**Option 2: PrometheusRule (Recommended)**

Instead of routing CloudEvents to Alertmanager, emit Prometheus metrics and use `PrometheusRule`:

```yaml
# PrometheusRule evaluates metrics → Alertmanager routes alerts
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: knative-lambda-alerts
spec:
  groups:
    - name: lambda.alerts
      rules:
        - alert: LambdaBuildFailed
          expr: increase(knative_lambda_build_errors_total[5m]) > 0
          labels:
            severity: warning
```

---

## 🔄 Flux CD Integration (CDEvents)

Flux CD's Notification Controller supports CloudEvents through its **Receiver** resource with `type: cdevents`. This enables GitOps-triggered responses to event-driven workflows from the Knative Lambda platform.

### Overview

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    FLUX CD ↔ KNATIVE LAMBDA INTEGRATION                          │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │                    KNATIVE LAMBDA OPERATOR                                  │  │
│  │                                                                             │  │
│  │   Emits CloudEvents:                                                        │  │
│  │   • io.knative.lambda.lifecycle.function.ready                             │  │
│  │   • io.knative.lambda.lifecycle.build.completed                            │  │
│  │   • io.knative.lambda.lifecycle.service.ready                              │  │
│  └────────────────────┬───────────────────────────────────────────────────────┘  │
│                       │                                                          │
│                       │ CloudEvents (HTTP POST)                                  │
│                       ▼                                                          │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │                    FLUX NOTIFICATION CONTROLLER                             │  │
│  │                                                                             │  │
│  │  Receiver (type: cdevents)                                                  │  │
│  │  ├─ Validates CDEvents payload using CDEvent Go-SDK                        │  │
│  │  ├─ Filters by event type header                                           │  │
│  │  ├─ CEL expressions for advanced filtering                                 │  │
│  │  └─ Triggers reconciliation of Flux resources                              │  │
│  └────────────────────┬───────────────────────────────────────────────────────┘  │
│                       │                                                          │
│                       │ Reconcile                                                │
│                       ▼                                                          │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │                         FLUX GITOPS RESOURCES                               │  │
│  │                                                                             │  │
│  │  • GitRepository (pull latest configs)                                      │  │
│  │  • Kustomization (apply manifests)                                          │  │
│  │  • HelmRelease (upgrade charts)                                             │  │
│  │  • ImagePolicy (scan for new tags)                                          │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Flux Receiver Configuration

The Flux Notification API v1 `Receiver` resource with `type: cdevents` can receive CloudEvents from the Knative Lambda platform:

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1
kind: Receiver
metadata:
  name: lambda-lifecycle-receiver
  namespace: flux-system
spec:
  # CDEvents receiver type validates CloudEvents payloads
  type: cdevents
  
  # Filter specific event types
  events:
    - "io.knative.lambda.lifecycle.function.ready"
    - "io.knative.lambda.lifecycle.build.completed"
    - "io.knative.lambda.lifecycle.service.ready"
  
  # Authentication secret
  secretRef:
    name: lambda-webhook-token
  
  # Flux resources to trigger reconciliation
  resources:
    - kind: GitRepository
      name: homelab-config
      namespace: flux-system
    - kind: Kustomization
      name: knative-lambda
      namespace: flux-system
    - kind: Kustomization
      name: knative-lambda-configs
      namespace: flux-system
```

### Use Cases

| Lambda Event | Flux Action | Use Case |
|--------------|-------------|----------|
| `lifecycle.function.ready` | Reconcile Kustomization | Update ConfigMaps, Secrets, NetworkPolicies |
| `lifecycle.build.completed` | Trigger ImageRepository scan | Auto-detect new image tags for GitOps |
| `lifecycle.service.created` | Reconcile dependent configs | Apply service-specific configurations |
| `lifecycle.service.deleted` | Cleanup orphan resources | Remove associated ConfigMaps, PVCs |
| `command.build.start` | (Audit only) | Log build initiation to Git history |

### Security Alert → GitOps Response

Integration with `agent-contracts` for automated security responses:

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1
kind: Receiver
metadata:
  name: security-alert-receiver
  namespace: flux-system
spec:
  type: cdevents
  events:
    - "io.homelab.exploit.validated"
    - "io.homelab.vuln.found"
  secretRef:
    name: agent-contracts-webhook-token
  resources:
    - kind: Kustomization
      name: security-policies
      namespace: flux-system
  # CEL filter for critical severity only
  resourceFilter: |
    request.body.data.severity == "critical"
```

**Automated Response Flow**:

```
agent-contracts detects exploit
        │
        │ io.homelab.exploit.validated
        ▼
┌─────────────────────────────────────────────────────────────────┐
│  Flux Receiver (security-alert-receiver)                        │
│  ├─ Validates CloudEvent payload                                │
│  ├─ CEL filter: severity == "critical"                          │
│  └─ Triggers reconciliation                                     │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│  Kustomization: security-policies                               │
│  ├─ Deploys NetworkPolicy (quarantine affected pods)            │
│  ├─ Updates WAF rules                                           │
│  └─ Applies security patches from Git                           │
└─────────────────────────────────────────────────────────────────┘
```

### Build Complete → Image Automation

Trigger Flux Image Automation when Kaniko builds complete:

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1
kind: Receiver
metadata:
  name: build-complete-receiver
  namespace: flux-system
spec:
  type: cdevents
  events:
    - "io.knative.lambda.lifecycle.build.completed"
  secretRef:
    name: build-webhook-token
  resources:
    - kind: ImageRepository
      name: lambda-images
      namespace: flux-system
```

This enables:
1. **Build completes** → Kaniko pushes image to registry
2. **CloudEvent emitted** → `lifecycle.build.completed`
3. **Flux Receiver** → Triggers ImageRepository reconciliation
4. **ImagePolicy** → Detects new tag, updates manifest
5. **GitOps** → Commits change, deploys new version

### Publishing Events to Flux

To send CloudEvents to Flux Receiver, configure a Knative Trigger:

```yaml
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: lifecycle-to-flux
  namespace: knative-lambda
spec:
  broker: lambda-broker
  filter:
    attributes:
      type: io.knative.lambda.lifecycle.function.ready
  subscriber:
    # Flux Receiver webhook URL format:
    # /hook/sha256sum(token+name+namespace)
    uri: http://notification-controller.flux-system.svc.cluster.local/hook/<webhook-path>
```

Alternatively, use a SinkBinding for direct HTTP publishing:

```yaml
apiVersion: sources.knative.dev/v1
kind: SinkBinding
metadata:
  name: operator-to-flux
  namespace: knative-lambda
spec:
  subject:
    apiVersion: apps/v1
    kind: Deployment
    name: knative-lambda-operator
  sink:
    uri: http://notification-controller.flux-system.svc.cluster.local/hook/<webhook-path>
```

### Webhook Secret Setup

```bash
# Generate secure token
TOKEN=$(head -c 12 /dev/urandom | shasum | head -c 32)

# Create secret in flux-system namespace
kubectl create secret generic lambda-webhook-token \
  --from-literal=token=$TOKEN \
  -n flux-system

# Get the webhook URL path (after Receiver is created)
kubectl get receiver lambda-lifecycle-receiver -n flux-system \
  -o jsonpath='{.status.webhookPath}'
```

### Supported CDEvent Types

Flux Notification Controller validates events using the [CDEvents Go-SDK](https://github.com/cdevents/sdk-go). The following CDEvents types are commonly used:

| CDEvent Type | Description | Knative Lambda Equivalent |
|--------------|-------------|---------------------------|
| `dev.cdevents.change.merged` | Code change merged | `io.knative.lambda.command.build.start` |
| `dev.cdevents.artifact.published` | Artifact published | `io.knative.lambda.lifecycle.build.completed` |
| `dev.cdevents.service.deployed` | Service deployed | `io.knative.lambda.lifecycle.service.ready` |
| `dev.cdevents.service.removed` | Service removed | `io.knative.lambda.lifecycle.service.deleted` |

### Architecture: Complete Integration

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           HOMELAB CLUSTER                                        │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  KNATIVE LAMBDA OPERATOR              AGENT-CONTRACTS                   │    │
│  │  ├─ lifecycle.function.ready          ├─ io.homelab.contract.created   │    │
│  │  ├─ lifecycle.build.completed         ├─ io.homelab.vuln.found         │    │
│  │  └─ lifecycle.service.ready           └─ io.homelab.exploit.validated  │    │
│  └────────────────────┬───────────────────────────┬────────────────────────┘    │
│                       │                           │                              │
│                       │ CloudEvents               │                              │
│                       ▼                           ▼                              │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │                       RABBITMQ BROKER                                    │    │
│  │                   (Knative Eventing Exchange)                            │    │
│  └────────────────────┬───────────────────────────┬────────────────────────┘    │
│                       │                           │                              │
│           ┌───────────┴───────────┐               │                              │
│           │                       │               │                              │
│           ▼                       ▼               ▼                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐      │
│  │  Knative        │  │  Alertmanager   │  │  FLUX NOTIFICATION          │      │
│  │  Triggers       │  │  (via notifi-   │  │  CONTROLLER                 │      │
│  │  (Functions)    │  │  adapter)       │  │                             │      │
│  └─────────────────┘  └─────────────────┘  │  ┌─────────────────────────┐│      │
│                                            │  │  Receiver: cdevents     ││      │
│                                            │  │  ├─ lambda-lifecycle    ││      │
│                                            │  │  ├─ security-alerts     ││      │
│                                            │  │  └─ build-complete      ││      │
│                                            │  └─────────────────────────┘│      │
│                                            └──────────────┬──────────────┘      │
│                                                           │                      │
│                                                           │ Reconcile            │
│                                                           ▼                      │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │                         FLUX GITOPS                                      │    │
│  │                                                                          │    │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐         │    │
│  │  │  GitRepository  │  │  Kustomization  │  │   ImagePolicy   │         │    │
│  │  │  (homelab)      │  │  (knative-      │  │  (lambda-       │         │    │
│  │  │                 │  │   lambda)       │  │   images)       │         │    │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘         │    │
│  │                                                                          │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### References

- [Flux Notification API v1 - Receivers](https://fluxcd.io/flux/components/notification/receivers/)
- [Flux Notification API Reference](https://fluxcd.io/flux/components/notification/api/v1/)
- [CDEvents Specification](https://cdevents.dev/)
- [CDEvents Go-SDK](https://github.com/cdevents/sdk-go)

---

## 📚 References

- [CloudEvents Specification v1.0](https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/spec.md)
- [CloudEvents JSON Format](https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/formats/json-format.md)
- [CloudEvents HTTP Protocol Binding](https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/bindings/http-protocol-binding.md)
- [Event Sourcing Pattern](https://microservices.io/patterns/data/event-sourcing.html)
- [CQRS Pattern](https://microservices.io/patterns/data/cqrs.html)
- [Flux CD Notification Controller](https://fluxcd.io/flux/components/notification/)
- [CDEvents - Continuous Delivery Events](https://cdevents.dev/)

---

**Maintainer**: Platform Team  
**Review Cycle**: Quarterly

