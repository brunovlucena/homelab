# ⚡ SRE-010: Dead Letter Queue Management

**Status**: Backlog  
**Priority**: P0  
**Story Points**: 13  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-180/sre-010-dead-letter-queue-management  
**Created**: 2026-01-19  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---


## 📋 User Story

**As a** SRE Engineer  
**I want to** register parser result  
**So that** I can improve system reliability, security, and performance

---



## 🎯 Acceptance Criteria

- [ ] [ ] Poison messages detected within 30 seconds
- [ ] [ ] Consumers survive poison message without crashing
- [ ] [ ] DLQ receives poison message after first retry
- [ ] [ ] Alert fires: "PoisonMessageDetected"
- [ ] [ ] Message metadata includes failure reason
- [ ] [ ] Dashboard shows poison message count
- [ ] [ ] Automated schema validation before processing

---


## Overview

This runbook provides comprehensive procedures for managing Dead Letter Queues (DLQ) in the Knative Lambda event-driven system. DLQs capture events that fail processing after retry exhaustion, preventing data loss and enabling failure investigation.

---

## 📊 Event Failure Scenarios

### Failure Taxonomy | Failure Type | Root Cause | DLQ Impact | Recovery Strategy | |-------------- | ----------- | ------------ | ------------------- | | **Poison Message** | Malformed event schema | Immediate DLQ | Schema validation + manual fix | | **Consumer Crash** | Service OOM/panic | DLQ after retries | Resource tuning + replay | | **Timeout** | Long-running processing | DLQ after timeout | Increase timeout + replay | | **Dependency Failure** | External service down | DLQ after retries | Wait for service + replay | | **Resource Exhaustion** | CPU/Memory limits | DLQ after retries | Scale resources + replay | | **Schema Evolution** | Incompatible version | Immediate DLQ | Deploy compatible version | | **Network Partition** | Broker unreachable | DLQ after retries | Network fix + replay | | **Validation Failure** | Business rule violation | DLQ after retries | Fix validation + replay | | **Partial Processing** | Mid-processing failure | DLQ after retries | Idempotent retry | | **Broker Failure** | RabbitMQ crash | DLQ on recovery | Broker restart + replay | ---

## 🎯 User Story 1: Poison Message Detection and Remediation

### Story

**As an** SRE Engineer  
**I want** to automatically detect and isolate poison messages  
**So that** one bad event doesn't crash all consumers or block the queue

### Acceptance Criteria

- [ ] Poison messages detected within 30 seconds
- [ ] Consumers survive poison message without crashing
- [ ] DLQ receives poison message after first retry
- [ ] Alert fires: "PoisonMessageDetected"
- [ ] Message metadata includes failure reason
- [ ] Dashboard shows poison message count
- [ ] Automated schema validation before processing

### Failure Scenario

```yaml
Event Flow:
  1. Malformed CloudEvent arrives at broker
     ├─ Missing required field: event.data.contextId
     ├─ Invalid JSON in event.data
     └─ Schema version mismatch (v2 event to v1 consumer)
  
  2. Consumer attempts processing
     ├─ JSON parse fails
     ├─ Validation error thrown
     └─ Consumer returns 500 error
  
  3. Knative retry logic
     ├─ Retry 1: Backoff 1s → FAIL
     ├─ Retry 2: Backoff 2s → FAIL
     ├─ Retry 3: Backoff 4s → FAIL
     ├─ Retry 4: Backoff 8s → FAIL
     └─ Retry 5: Backoff 16s → FAIL
  
  4. Event moved to DLQ
     └─ DLQ: lambda-build-events-prd-dlq
```

### Detection

```bash
# Monitor for repeated 500 errors on same event ID
kubectl logs -n knative-lambda -l app=knative-lambda-builder | \
  grep "event_id" | sort | uniq -c | grep -E "^\s+[5-9]"

# Check DLQ depth
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl list_queues name messages | grep dlq

# Query Prometheus for poison message pattern
curl -g 'http://prometheus:9090/api/v1/query' \
  --data-urlencode 'query=rate(http_requests_total{status="500",handler="cloud_event"}[5m]) > 0.1'
```

### Remediation Steps

```bash
# Step 1: Identify poison message in DLQ
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin get queue=lambda-build-events-prd-dlq count=1

# Step 2: Extract message to file for analysis
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin get queue=lambda-build-events-prd-dlq count=1 > poison_message.json

# Step 3: Analyze message structure
cat poison_message.json | jq '.payload' | jq '.data'

# Step 4: Fix message schema
cat poison_message.json | jq '.payload' | jq '.data | = . + {"contextId": "recovered-context"}' > fixed_message.json

# Step 5: Republish to main queue (after validation)
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin publish exchange=knative-lambda-broker-prd \
  routing_key=lambda-build-events \
  payload="$(cat fixed_message.json)"

# Step 6: Remove from DLQ
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin delete queue=lambda-build-events-prd-dlq count=1
```

### Prevention

```yaml
# Add schema validation middleware
apiVersion: v1
kind: ConfigMap
metadata:
  name: event-schema-validation
data:
  schema.json: | {
      "type": "object",
      "required": ["data", "type", "source", "id"],
      "properties": {
        "data": {
          "type": "object",
          "required": ["contextId", "thirdPartyId", "parserId"],
          "properties": {
            "contextId": {"type": "string", "minLength": 1},
            "thirdPartyId": {"type": "string"},
            "parserId": {"type": "string"}
          }
        },
        "type": {"type": "string"},
        "source": {"type": "string"},
        "id": {"type": "string"}
      }
    }
```

### Monitoring

```prometheus
# Alert: Poison Message Detected
- alert: PoisonMessageDetected
  expr: | rate(cloudevents_processing_errors_total{error_type="validation"}[5m]) > 0.1
  for: 1m
  severity: warning
  annotations:
    summary: "Poison message detected in {{ $labels.queue }}"
    description: "Event validation failures spiking. Check DLQ for malformed events."

# Alert: DLQ Not Empty
- alert: DeadLetterQueueNotEmpty
  expr: rabbitmq_queue_messages{queue=~".*-dlq"} > 0
  for: 5m
  severity: warning
  annotations:
    summary: "Dead Letter Queue {{ $labels.queue }} has {{ $value }} messages"
    description: "Failed events require investigation and replay."
```

---

## 🎯 User Story 2: Consumer Service Unavailability

### Story

**As an** SRE Engineer  
**I want** events to be safely queued in DLQ when consumers are unavailable  
**So that** no events are lost during deployment, crashes, or maintenance

### Acceptance Criteria

- [ ] Events preserved during consumer downtime
- [ ] DLQ receives events after retry exhaustion
- [ ] Consumers auto-recover and resume processing
- [ ] Events automatically replayed after recovery
- [ ] Zero event loss during deployments
- [ ] Alert fires: "ConsumerUnavailable"
- [ ] Deployment strategy includes queue drain

### Failure Scenario

```yaml
Scenario: Rolling Deployment with Consumer Downtime
  
  1. Deployment initiated (kubectl rollout)
     ├─ Old pods: 2 replicas running
     ├─ New pods: 0 replicas (pulling image)
     └─ Events arriving at 10/sec
  
  2. Old pods terminating (SIGTERM)
     ├─ Graceful shutdown: 30s grace period
     ├─ In-flight events: 20 events processing
     ├─ Pending events: 50 events in consumer prefetch
     └─ New events: Continue arriving (no consumers ready)
  
  3. New pods starting
     ├─ Image pull: 15s
     ├─ Container startup: 10s
     ├─ Health check: 5s (readiness probe)
     └─ Total downtime: 30s
  
  4. Event delivery attempts during gap
     ├─ Attempt 1: 503 Service Unavailable
     ├─ Retry after 1s: 503
     ├─ Retry after 2s: 503
     ├─ Retry after 4s: 503
     ├─ Retry after 8s: 503
     └─ After 5 retries → DLQ
  
  5. Recovery
     ├─ New pods ready
     ├─ Events in DLQ: 300 events (30s × 10/sec)
     └─ Require manual or automated replay
```

### Detection

```bash
# Check consumer availability
kubectl get pods -n knative-lambda -l app=knative-lambda-builder

# Check event delivery failures
kubectl logs -n knative-eventing -l eventing.knative.dev/broker=lambda-broker-prd | \
  grep -i "connection refused\ | service unavailable"

# Check DLQ accumulation during deployment
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl list_queues name messages | grep dlq | \
  awk '{print $2}' | \
  xargs -I {} echo "DLQ Depth: {}"
```

### Remediation Steps

```bash
# Step 1: Verify consumer recovery
kubectl get pods -n knative-lambda -l app=knative-lambda-builder -o wide
kubectl logs -n knative-lambda -l app=knative-lambda-builder --tail=50

# Step 2: Check DLQ depth
DLQ_DEPTH=$(kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl list_queues name messages | grep lambda-build-events-prd-dlq | awk '{print $2}')
echo "Events to replay: $DLQ_DEPTH"

# Step 3: Automated replay from DLQ to main queue
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- bash <<'EOF'
  rabbitmq-plugins enable rabbitmq_shovel
  
  rabbitmqctl set_parameter shovel dlq-replay \
    '{"src-queue": "lambda-build-events-prd-dlq",
      "src-uri": "amqp://",
      "dest-exchange": "knative-lambda-broker-prd",
      "dest-exchange-key": "lambda-build-events",
      "dest-uri": "amqp://",
      "ack-mode": "on-confirm",
      "delete-after": "never"}'
EOF

# Step 4: Monitor replay progress
watch -n 5 'kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl list_queues name messages | grep lambda-build-events'

# Step 5: Verify consumers processing events
kubectl logs -n knative-lambda -l app=knative-lambda-builder --tail=100 | \
  grep "Processing CloudEvent"

# Step 6: Clean up shovel after replay complete
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl clear_parameter shovel dlq-replay
```

### Prevention - Zero-Downtime Deployment

```yaml
# Updated Deployment Strategy
apiVersion: apps/v1
kind: Deployment
metadata:
  name: knative-lambda-builder
spec:
  replicas: 3  # Increased from 2 for overlap
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1  # Always keep 2 pods running
      maxSurge: 1  # Allow 4 pods during rollout
  template:
    spec:
      containers:
      - name: builder
        lifecycle:
          preStop:
            exec:
              command:
              - /bin/sh
              - -c
              - | # Drain consumers gracefully
                echo "Draining event consumers..."
                # Stop accepting new events
                curl -X POST localhost:8081/admin/drain
                # Wait for in-flight events to complete
                sleep 25
                echo "Drain complete"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          failureThreshold: 2
      terminationGracePeriodSeconds: 30
```

### Monitoring

```prometheus
# Alert: Consumer Unavailable
- alert: ConsumerServiceUnavailable
  expr: | up{job="knative-lambda-builder"} == 0
  for: 1m
  severity: critical
  annotations:
    summary: "All consumer pods are down"
    description: "Events will be sent to DLQ after retry exhaustion."

# Alert: Consumer Scale During Deployment
- alert: ConsumerScaledDownDuringLoad
  expr: | (rabbitmq_queue_messages{queue="lambda-build-events-prd"} > 100) 
    and 
    (kube_deployment_status_replicas_available{deployment="knative-lambda-builder"} < 2)
  for: 30s
  severity: critical
  annotations:
    summary: "Consumers scaled down while queue has backlog"
    description: "Risk of events moving to DLQ. Immediate action required."

# Metric: DLQ Accumulation Rate During Deployments
- record: dlq_accumulation_rate
  expr: | rate(rabbitmq_queue_messages{queue=~".*-dlq"}[1m])
```

---

## 🎯 User Story 3: Processing Timeout Failures

### Story

**As an** SRE Engineer  
**I want** long-running event processing to timeout gracefully  
**So that** slow consumers don't block the queue and events are properly handled

### Acceptance Criteria

- [ ] Processing timeout set to 60 seconds
- [ ] Timed-out events moved to DLQ
- [ ] Timeout metadata captured in DLQ message
- [ ] Alert fires: "EventProcessingTimeout"
- [ ] Grafana shows timeout patterns by parser
- [ ] Long-running parsers identified automatically
- [ ] Timeout values tunable per event type

### Failure Scenario

```yaml
Scenario: Parser Execution Timeout

  1. CloudEvent received: parser.start
     ├─ Parser ID: "complex-data-parser"
     ├─ Expected processing time: 5s
     └─ Actual processing time: 120s (hung)
  
  2. Consumer processing
     ├─ Knative timeout: 60s
     ├─ Parser timeout: Not configured (default: infinite)
     └─ Parser hangs on external API call
  
  3. Timeout enforcement
     ├─ Knative gives up after 60s
     ├─ Consumer returns: 504 Gateway Timeout
     ├─ In-flight work abandoned (no cancellation)
     └─ Consumer pod may need restart if goroutine leaked
  
  4. Retry attempts
     ├─ Retry 1: Timeout after 60s → FAIL
     ├─ Retry 2: Timeout after 60s → FAIL
     ├─ Retry 3: Timeout after 60s → FAIL
     ├─ Retry 4: Timeout after 60s → FAIL
     └─ Retry 5: Timeout after 60s → FAIL
  
  5. DLQ placement
     └─ Event moved to DLQ with timeout metadata
     
  Total time wasted: 300 seconds (5 retries × 60s)
```

### Detection

```bash
# Identify timeout errors in logs
kubectl logs -n knative-lambda -l app=knative-lambda-builder | \
  grep -i "timeout\ | deadline exceeded\ | context deadline"

# Check parser execution times
kubectl logs -n knative-lambda -l app=knative-lambda-builder | \
  jq -r 'select(.parser_duration_ms > 60000) | [.parser_id, .parser_duration_ms] | @csv'

# Query Prometheus for slow parsers
curl -g 'http://prometheus:9090/api/v1/query' \
  --data-urlencode 'query=histogram_quantile(0.95, rate(parser_execution_duration_seconds_bucket[5m])) > 60'

# Check DLQ for timeout events
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin get queue=lambda-build-events-prd-dlq count=10 | \
  jq '.[] | select(.properties.headers."x-death-reason" == "timeout")'
```

### Remediation Steps

```bash
# Step 1: Identify problematic parser
kubectl logs -n knative-lambda -l app=knative-lambda-builder --since=1h | \
  jq -r 'select(.parser_duration_ms > 60000) | .parser_id' | \
  sort | uniq -c | sort -rn

# Step 2: Analyze timed-out events in DLQ
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin get queue=lambda-build-events-prd-dlq count=100 > timeout_events.json

cat timeout_events.json | jq '[.[] | select(.properties.headers."x-death-reason" == "timeout")] | group_by(.payload.data.parserId) | map({parser: .[0].payload.data.parserId, count: length})'

# Step 3: Update parser timeout configuration
kubectl patch configmap knative-lambda-config -n knative-lambda --type merge -p '
{
  "data": {
    "PARSER_TIMEOUT_SECONDS": "120",
    "PARSER_TIMEOUT_complex-data-parser": "300"
  }
}'

# Step 4: Restart consumers to pick up new config
kubectl rollout restart deployment/knative-lambda-builder -n knative-lambda

# Step 5: Replay timed-out events from DLQ
cat timeout_events.json | jq -r '.[] | .payload' | while read event; do
  echo "$event" | kubectl exec -i -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
    rabbitmqadmin publish exchange=knative-lambda-broker-prd \
    routing_key=lambda-build-events \
    payload=-
done
```

### Prevention - Timeout Configuration

```go
// Add context timeout to parser execution
func (h *EventHandlerImpl) executeParserWithTimeout(ctx context.Context, event *cloudevents.Event) error {
    // Get parser-specific timeout or use default
    parserID := event.Data().(map[string]interface{})["parserId"].(string)
    timeout := h.getParserTimeout(parserID) // Default: 60s, configurable per parser
    
    // Create context with timeout
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    // Execute parser with cancellation
    resultChan := make(chan error, 1)
    go func() {
        resultChan <- h.executeParser(ctx, event)
    }()
    
    // Wait for result or timeout
    select {
    case err := <-resultChan:
        return err
    case <-ctx.Done():
        h.obs.Error(ctx, ctx.Err(), "Parser execution timed out",
            "parser_id", parserID,
            "timeout_seconds", timeout.Seconds())
        return fmt.Errorf("parser execution timeout: %w", ctx.Err())
    }
}
```

```yaml
# Knative Service timeout configuration
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: lambda-parser-complex-data
spec:
  template:
    metadata:
      annotations:
        # Request timeout (Knative timeout)
        serving.knative.dev/timeout-seconds: "300"
    spec:
      containers:
      - name: parser
        env:
        - name: PARSER_TIMEOUT
          value: "280"  # 20s buffer for graceful handling
        - name: EXTERNAL_API_TIMEOUT
          value: "60"  # Timeout for external dependencies
```

### Monitoring

```prometheus
# Alert: Event Processing Timeout
- alert: EventProcessingTimeout
  expr: | rate(cloudevents_processing_errors_total{error_type="timeout"}[5m]) > 0.1
  for: 2m
  severity: warning
  annotations:
    summary: "High rate of event processing timeouts"
    description: "{{ $value }} events/sec timing out. Check parser performance."

# Alert: Parser Consistently Slow
- alert: ParserConsistentlySlow
  expr: | histogram_quantile(0.5, rate(parser_execution_duration_seconds_bucket[5m])) > 30
  for: 10m
  severity: warning
  labels:
    parser_id: "{{ $labels.parser_id }}"
  annotations:
    summary: "Parser {{ $labels.parser_id }} is consistently slow"
    description: "Median execution time > 30s. May need timeout increase or optimization."

# Dashboard: Timeout Patterns
- panel: Parser Execution Time (p50, p95, p99)
  expr: | histogram_quantile(0.50, rate(parser_execution_duration_seconds_bucket[5m]))
    histogram_quantile(0.95, rate(parser_execution_duration_seconds_bucket[5m]))
    histogram_quantile(0.99, rate(parser_execution_duration_seconds_bucket[5m]))
```

---

## 🎯 User Story 4: Dependency Service Failures

### Story

**As an** SRE Engineer  
**I want** events to be safely queued when external dependencies fail  
**So that** temporary service outages don't cause permanent event loss

### Acceptance Criteria

- [ ] Circuit breaker for external service calls
- [ ] Events retry with exponential backoff
- [ ] DLQ receives events after dependency timeout
- [ ] Alert fires: "DependencyServiceDown"
- [ ] External service health checks automated
- [ ] Events automatically replay when service recovers
- [ ] Dependency failure patterns tracked

### Failure Scenario

```yaml
Scenario: External API Dependency Failure

  1. CloudEvent processing requires external API
     ├─ Service: Notifi Scheduler (scheduler.notifi.svc.cluster.local)
     ├─ Purpose: Register parser result
     └─ Expected latency: 200ms
  
  2. Scheduler service crashes
     ├─ Root cause: OOMKilled
     ├─ Restart in progress: 45s
     └─ Endpoint returns: Connection Refused
  
  3. Parser completes successfully
     ├─ Event parsed: SUCCESS
     ├─ Result ready: {"succeeded": true, "eventEntries": [...]}
     └─ Attempt to POST to scheduler
  
  4. POST to scheduler fails
     ├─ Error: ECONNREFUSED
     ├─ Consumer returns 500 error
     └─ Knative initiates retry
  
  5. Retry attempts
     ├─ Retry 1 (1s): Scheduler still down → FAIL
     ├─ Retry 2 (2s): Scheduler still down → FAIL
     ├─ Retry 3 (4s): Scheduler still down → FAIL
     ├─ Retry 4 (8s): Scheduler still down → FAIL
     └─ Retry 5 (16s): Scheduler still down → FAIL
  
  6. Event moved to DLQ
     └─ Parser result lost (not idempotent)
```

### Detection

```bash
# Check scheduler service health
kubectl get pods -n notifi -l app=notifi-scheduler
kubectl logs -n notifi -l app=notifi-scheduler --tail=50

# Check for connection refused errors
kubectl logs -n knative-lambda -l app=knative-lambda-builder | \
  grep -i "ECONNREFUSED\ | connection refused\ | no such host"

# Test scheduler endpoint manually
kubectl run curl-test -n knative-lambda --rm -it --restart=Never \
  --image=curlimages/curl:latest -- \
  curl -v http://notifi-scheduler.notifi.svc.cluster.local/fusion/execution/response

# Check DLQ for dependency failures
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin get queue=lambda-build-events-prd-dlq count=10 | \
  jq '.[] | select(.payload.data.errorMessage | contains("scheduler"))'
```

### Remediation Steps

```bash
# Step 1: Verify dependency service health
kubectl get all -n notifi -l app=notifi-scheduler

# Step 2: Check service recovery
kubectl logs -n notifi -l app=notifi-scheduler --tail=100

# Step 3: Wait for service to be ready (or fix it)
kubectl wait --for=condition=ready pod -l app=notifi-scheduler -n notifi --timeout=60s

# Step 4: Test service endpoint
kubectl run curl-test -n knative-lambda --rm -it --restart=Never \
  --image=curlimages/curl:latest -- \
  curl -X POST http://notifi-scheduler.notifi.svc.cluster.local/fusion/execution/response \
  -H "Content-Type: application/json" \
  -d '{"test": true}'

# Step 5: Replay events from DLQ
# First, extract events that failed due to scheduler
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin get queue=lambda-build-events-prd-dlq count=1000 > dlq_events.json

cat dlq_events.json | jq -c '.[] | select(.payload.data.errorMessage | contains("scheduler"))' > scheduler_failures.json

# Step 6: Replay each event
cat scheduler_failures.json | while IFS= read -r event; do
  echo "$event" | jq -c '.payload' | kubectl exec -i -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
    rabbitmqadmin publish exchange=knative-lambda-broker-prd \
    routing_key=lambda-build-events \
    payload=-
  echo "Replayed event"
  sleep 0.1  # Rate limit
done
```

### Prevention - Circuit Breaker Pattern

```javascript
// Add to index.js template - Circuit Breaker for Scheduler
class CircuitBreaker {
  constructor(failureThreshold = 5, timeout = 60000) {
    this.failureThreshold = failureThreshold;
    this.timeout = timeout;
    this.failureCount = 0;
    this.state = 'CLOSED'; // CLOSED, OPEN, HALF_OPEN
    this.nextAttempt = Date.now();
  }

  async call(fn) {
    if (this.state === 'OPEN') {
      if (Date.now() < this.nextAttempt) {
        throw new Error('Circuit breaker is OPEN. Service unavailable.');
      }
      this.state = 'HALF_OPEN';
    }

    try {
      const result = await fn();
      this.onSuccess();
      return result;
    } catch (error) {
      this.onFailure();
      throw error;
    }
  }

  onSuccess() {
    this.failureCount = 0;
    this.state = 'CLOSED';
  }

  onFailure() {
    this.failureCount++;
    if (this.failureCount >= this.failureThreshold) {
      this.state = 'OPEN';
      this.nextAttempt = Date.now() + this.timeout;
      console.error('[CIRCUIT_BREAKER] Circuit opened. Too many failures.');
    }
  }
}

// Initialize circuit breaker for scheduler
const schedulerCircuitBreaker = new CircuitBreaker(5, 60000);

// Use circuit breaker in handleCloudEvent
const sendToScheduler = async (result) => {
  return schedulerCircuitBreaker.call(async () => {
    const response = await fetch(SCHEDULER_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(result),
      timeout: 5000, // 5s timeout per request
    });

    if (!response.ok) {
      throw new Error(`Scheduler error: ${response.status}`);
    }

    return response.json();
  });
};
```

### Monitoring

```prometheus
# Alert: Dependency Service Down
- alert: DependencyServiceDown
  expr: | up{job="notifi-scheduler"} == 0
  for: 1m
  severity: critical
  annotations:
    summary: "Notifi Scheduler is down"
    description: "Parser results cannot be delivered. Events will move to DLQ after 5 retries."

# Alert: High Dependency Error Rate
- alert: HighDependencyErrorRate
  expr: | rate(external_service_errors_total{service="scheduler"}[5m]) > 0.5
  for: 2m
  severity: warning
  annotations:
    summary: "High error rate calling {{ $labels.service }}"
    description: "{{ $value }} errors/sec calling external service."

# Circuit Breaker Metrics
- metric: circuit_breaker_state
  expr: | circuit_breaker_state{service="scheduler"}  # 0=CLOSED, 1=OPEN, 2=HALF_OPEN
```

---

## 🎯 User Story 5: Resource Exhaustion Failures

### Story

**As an** SRE Engineer  
**I want** to prevent resource exhaustion from causing cascading DLQ failures  
**So that** OOM kills and CPU throttling don't degrade the entire event system

### Acceptance Criteria

- [ ] Memory limits enforced per consumer pod
- [ ] CPU requests/limits configured appropriately
- [ ] OOMKilled pods restart and resume processing
- [ ] Resource metrics tracked per parser
- [ ] Alert fires: "ConsumerResourceExhaustion"
- [ ] Auto-scaling based on resource usage
- [ ] Heavy parsers identified and isolated

### Failure Scenario

```yaml
Scenario: Parser Memory Leak Causing OOMKill

  1. Consumer pod resources
     ├─ Memory request: 256Mi
     ├─ Memory limit: 512Mi
     ├─ CPU request: 100m
     └─ CPU limit: 1000m
  
  2. Memory-intensive parser triggered
     ├─ Parser: "large-dataset-parser"
     ├─ Expected memory: 100Mi
     ├─ Actual memory: 800Mi (memory leak)
     └─ Processing large dataset (50MB payload)
  
  3. Memory consumption grows
     ├─ T+0s: 100Mi (baseline)
     ├─ T+5s: 300Mi (processing)
     ├─ T+10s: 500Mi (approaching limit)
     └─ T+15s: 512Mi → OOMKilled
  
  4. Pod killed mid-processing
     ├─ Kubernetes sends SIGKILL
     ├─ Event processing interrupted
     ├─ No graceful shutdown
     └─ Event nack'd back to broker
  
  5. Pod restarts
     ├─ Restart delay: 10s
     ├─ Pod initialization: 5s
     └─ Same event consumed again
  
  6. Retry loop
     ├─ Retry 1: OOMKilled → FAIL
     ├─ Retry 2: OOMKilled → FAIL
     ├─ Retry 3: OOMKilled → FAIL
     ├─ Retry 4: OOMKilled → FAIL
     └─ Retry 5: OOMKilled → FAIL
  
  7. Event moved to DLQ
     └─ Other events also failing (pod keeps crashing)
```

### Detection

```bash
# Check for OOMKilled pods
kubectl get pods -n knative-lambda -o json | \
  jq -r '.items[] | select(.status.containerStatuses[].lastState.terminated.reason == "OOMKilled") | .metadata.name'

# Check memory usage
kubectl top pods -n knative-lambda -l app=knative-lambda-builder

# Check pod restart count
kubectl get pods -n knative-lambda -l app=knative-lambda-builder -o json | \
  jq -r '.items[] | select(.status.containerStatuses[].restartCount > 3) | [.metadata.name, .status.containerStatuses[].restartCount] | @csv'

# Identify memory-intensive parsers
kubectl logs -n knative-lambda -l app=knative-lambda-builder --previous | \
  grep -B5 "OOMKilled\ | out of memory" | \
  jq -r 'select(.parser_id) | .parser_id' | \
  sort | uniq -c | sort -rn

# Check for CPU throttling
kubectl exec -n knative-lambda <pod-name> -- cat /sys/fs/cgroup/cpu/cpu.stat | grep throttled
```

### Remediation Steps

```bash
# Step 1: Identify problematic parser from DLQ
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin get queue=lambda-build-events-prd-dlq count=100 | \
  jq -r '.[].payload.data.parserId' | \
  sort | uniq -c | sort -rn

# Step 2: Increase memory limit for consumers
kubectl patch deployment knative-lambda-builder -n knative-lambda --type json -p='[
  {
    "op": "replace",
    "path": "/spec/template/spec/containers/0/resources/limits/memory",
    "value": "1Gi"
  },
  {
    "op": "replace",
    "path": "/spec/template/spec/containers/0/resources/requests/memory",
    "value": "512Mi"
  }
]'

# Step 3: Wait for rollout to complete
kubectl rollout status deployment/knative-lambda-builder -n knative-lambda

# Step 4: Test problematic parser
kubectl logs -n knative-lambda -l app=knative-lambda-builder --tail=50 -f

# Step 5: Replay events from DLQ (cautiously)
# Test with ONE event first
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin get queue=lambda-build-events-prd-dlq count=1 requeue=false | \
  jq -r '.[] | .payload' | \
  kubectl exec -i -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin publish exchange=knative-lambda-broker-prd \
  routing_key=lambda-build-events \
  payload=-

# Step 6: Monitor memory after replay
watch -n 5 'kubectl top pods -n knative-lambda -l app=knative-lambda-builder'

# If successful, replay remaining events
# If still failing, isolate problematic parser to dedicated high-memory pod
```

### Prevention - Resource Isolation

```yaml
# Isolate memory-intensive parsers to dedicated pods
apiVersion: apps/v1
kind: Deployment
metadata:
  name: knative-lambda-builder-high-memory
  namespace: knative-lambda
spec:
  replicas: 2
  selector:
    matchLabels:
      app: knative-lambda-builder
      profile: high-memory
  template:
    metadata:
      labels:
        app: knative-lambda-builder
        profile: high-memory
    spec:
      containers:
      - name: builder
        image: knative-lambda-builder:v1.0.0
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
        env:
        - name: PARSER_FILTER
          value: "large-dataset-parser,memory-intensive-parser"
        - name: MEMORY_PROFILE
          value: "high"

---
# Standard builder for normal parsers
apiVersion: apps/v1
kind: Deployment
metadata:
  name: knative-lambda-builder-standard
  namespace: knative-lambda
spec:
  replicas: 3
  selector:
    matchLabels:
      app: knative-lambda-builder
      profile: standard
  template:
    metadata:
      labels:
        app: knative-lambda-builder
        profile: standard
    spec:
      containers:
      - name: builder
        image: knative-lambda-builder:v1.0.0
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "1000m"
        env:
        - name: PARSER_FILTER
          value: "!large-dataset-parser,!memory-intensive-parser"
        - name: MEMORY_PROFILE
          value: "standard"
```

### Monitoring

```prometheus
# Alert: Consumer OOMKilled
- alert: ConsumerPodOOMKilled
  expr: | rate(kube_pod_container_status_terminated_reason{reason="OOMKilled",namespace="knative-lambda"}[5m]) > 0
  for: 1m
  severity: critical
  annotations:
    summary: "Consumer pod OOMKilled in {{ $labels.namespace }}"
    description: "Pod {{ $labels.pod }} was killed due to out of memory. Check parser memory usage."

# Alert: High Memory Usage
- alert: ConsumerHighMemoryUsage
  expr: | container_memory_usage_bytes{namespace="knative-lambda",container="builder"} 
    / 
    container_spec_memory_limit_bytes{namespace="knative-lambda",container="builder"} 
    > 0.8
  for: 5m
  severity: warning
  annotations:
    summary: "Consumer pod using >80% memory"
    description: "Pod {{ $labels.pod }} at {{ $value | humanizePercentage }} memory usage."

# Alert: CPU Throttling
- alert: ConsumerCPUThrottling
  expr: | rate(container_cpu_cfs_throttled_seconds_total{namespace="knative-lambda"}[5m]) > 0.5
  for: 5m
  severity: warning
  annotations:
    summary: "Consumer pod CPU throttled"
    description: "Pod {{ $labels.pod }} throttled {{ $value }} seconds in last 5min."
```

---

## 🔄 DLQ Replay Automation

### Automated Replay Strategy

```yaml
# CronJob: Automated DLQ Replay (runs every 5 minutes)
apiVersion: batch/v1
kind: CronJob
metadata:
  name: dlq-replay-automation
  namespace: knative-lambda
spec:
  schedule: "*/5 * * * *"
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: dlq-replay-sa
          containers:
          - name: dlq-replay
            image: brunolucena/dlq-replay:v1.0.0
            env:
            - name: RABBITMQ_HOST
              value: "rabbitmq-cluster-prd.rabbitmq-prd"
            - name: RABBITMQ_PORT
              value: "5672"
            - name: RABBITMQ_USER
              valueFrom:
                secretKeyRef:
                  name: rabbitmq-credentials
                  key: username
            - name: RABBITMQ_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: rabbitmq-credentials
                  key: password
            - name: DLQ_NAME
              value: "lambda-build-events-prd-dlq"
            - name: TARGET_EXCHANGE
              value: "knative-lambda-broker-prd"
            - name: REPLAY_BATCH_SIZE
              value: "10"
            - name: REPLAY_DELAY_MS
              value: "100"
            - name: MAX_AGE_HOURS
              value: "24"
            command:
            - /bin/sh
            - -c
            - | #!/bin/sh
              set -e
              
              echo "DLQ Replay Automation - Starting"
              
              # Check DLQ depth
              DLQ_DEPTH=$(rabbitmqctl list_queues name messages | grep $DLQ_NAME | awk '{print $2}')
              echo "DLQ depth: $DLQ_DEPTH messages"
              
              if [ "$DLQ_DEPTH" -eq 0 ]; then
                echo "DLQ is empty. Nothing to replay."
                exit 0
              fi
              
              # Check if services are healthy before replaying
              if ! kubectl get pods -n knative-lambda -l app=knative-lambda-builder | grep -q "Running"; then
                echo "Consumer pods not healthy. Skipping replay."
                exit 1
              fi
              
              # Replay messages in batches
              REPLAYED=0
              while [ $REPLAYED -lt $REPLAY_BATCH_SIZE ] && [ $REPLAYED -lt $DLQ_DEPTH ]; do
                # Get one message from DLQ
                MESSAGE=$(rabbitmqadmin get queue=$DLQ_NAME count=1 requeue=false)
                
                if [ -z "$MESSAGE" ]; then
                  echo "No more messages in DLQ"
                  break
                fi
                
                # Check message age
                MESSAGE_TIME=$(echo "$MESSAGE" | jq -r '.[].properties.timestamp')
                CURRENT_TIME=$(date +%s)
                AGE_HOURS=$(( ($CURRENT_TIME - $MESSAGE_TIME) / 3600 ))
                
                if [ $AGE_HOURS -gt $MAX_AGE_HOURS ]; then
                  echo "Message too old ($AGE_HOURS hours). Discarding."
                  continue
                fi
                
                # Republish to main queue
                echo "$MESSAGE" | jq -r '.[].payload' | \
                  rabbitmqadmin publish exchange=$TARGET_EXCHANGE \
                  routing_key=lambda-build-events \
                  payload=-
                
                REPLAYED=$((REPLAYED + 1))
                echo "Replayed message $REPLAYED"
                
                # Rate limiting
                sleep $(echo "scale=3; $REPLAY_DELAY_MS / 1000" | bc)
              done
              
              echo "DLQ Replay complete. Replayed $REPLAYED messages."
          restartPolicy: OnFailure
```

---

## 📊 DLQ Monitoring Dashboard

### Grafana Dashboard Configuration

```json
{
  "dashboard": {
    "title": "Dead Letter Queue Monitoring",
    "panels": [
      {
        "title": "DLQ Depth Over Time",
        "targets": [
          {
            "expr": "rabbitmq_queue_messages{queue=~\".*-dlq\"}"
          }
        ]
      },
      {
        "title": "DLQ Accumulation Rate",
        "targets": [
          {
            "expr": "rate(rabbitmq_queue_messages{queue=~\".*-dlq\"}[5m])"
          }
        ]
      },
      {
        "title": "Failed Events by Failure Type",
        "targets": [
          {
            "expr": "sum by (error_type) (rate(cloudevents_processing_errors_total[5m]))"
          }
        ]
      },
      {
        "title": "DLQ Messages by Parser ID",
        "targets": [
          {
            "expr": "sum by (parser_id) (dlq_messages_total)"
          }
        ]
      },
      {
        "title": "Event Processing Success Rate",
        "targets": [
          {
            "expr": "rate(cloudevents_processed_total{status=\"success\"}[5m]) / rate(cloudevents_processed_total[5m])"
          }
        ]
      },
      {
        "title": "Average Time in DLQ Before Replay",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(dlq_message_age_seconds_bucket[5m]))"
          }
        ]
      }
    ]
  }
}
```

---

## 🚨 Critical Alerts

### Alert Rules

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: dlq-alerts
  namespace: knative-lambda
spec:
  groups:
  - name: dead_letter_queue
    interval: 30s
    rules:
    
    # Critical: DLQ Growing Rapidly
    - alert: DLQGrowingRapidly
      expr: | rate(rabbitmq_queue_messages{queue=~".*-dlq"}[5m]) > 1
      for: 5m
      severity: critical
      annotations:
        summary: "DLQ {{ $labels.queue }} growing rapidly"
        description: "{{ $value }} messages/sec entering DLQ. Investigate root cause immediately."
    
    # Critical: DLQ Above Threshold
    - alert: DLQAboveThreshold
      expr: | rabbitmq_queue_messages{queue=~".*-dlq"} > 100
      for: 10m
      severity: warning
      annotations:
        summary: "DLQ {{ $labels.queue }} has {{ $value }} messages"
        description: "DLQ depth exceeds threshold. Manual intervention may be required."
    
    # Warning: Events Failing Consistently
    - alert: HighEventFailureRate
      expr: | rate(cloudevents_processing_errors_total[5m]) / rate(cloudevents_processed_total[5m]) > 0.1
      for: 5m
      severity: warning
      annotations:
        summary: "High event failure rate: {{ $value | humanizePercentage }}"
        description: "More than 10% of events failing. Check consumer health."
    
    # Warning: DLQ Not Being Processed
    - alert: DLQNotProcessed
      expr: | rabbitmq_queue_messages{queue=~".*-dlq"} > 10
        and
        rate(rabbitmq_queue_messages{queue=~".*-dlq"}[10m]) == 0
      for: 30m
      severity: warning
      annotations:
        summary: "DLQ {{ $labels.queue }} has stale messages"
        description: "DLQ has {{ $value }} messages that haven't been processed in 30min."
    
    # Info: DLQ Replay Success
    - alert: DLQReplaySuccessful
      expr: | increase(dlq_replay_success_total[5m]) > 0
      for: 1m
      severity: info
      annotations:
        summary: "DLQ replay succeeded"
        description: "{{ $value }} messages successfully replayed from DLQ."
```

---

## 🔧 Operational Procedures

### Daily DLQ Health Check

```bash
#!/bin/bash
# Daily DLQ health check script

echo "=== DLQ Health Check - $(date) ==="

# Check all DLQs
for ENV in dev stg prd; do
  echo ""
  echo "Environment: $ENV"
  
  # Get DLQ depth
  DLQ_DEPTH=$(kubectl exec -n rabbitmq-$ENV rabbitmq-cluster-$ENV-0 -- \
    rabbitmqctl list_queues name messages | \
    grep "lambda-build-events-$ENV-dlq" | \
    awk '{print $2}')
  
  echo "  DLQ Depth: $DLQ_DEPTH messages"
  
  # Get consumer health
  CONSUMER_COUNT=$(kubectl get pods -n knative-lambda-$ENV -l app=knative-lambda-builder --field-selector=status.phase=Running | wc -l)
  echo "  Consumers: $CONSUMER_COUNT running"
  
  # Get event failure rate
  FAILURE_RATE=$(curl -s "http://prometheus:9090/api/v1/query" \
    --data-urlencode "query=rate(cloudevents_processing_errors_total{environment='$ENV'}[5m])" | \
    jq -r '.data.result[0].value[1]')
  echo "  Failure Rate: $FAILURE_RATE errors/sec"
  
  # Recommendations
  if [ "$DLQ_DEPTH" -gt 100 ]; then
    echo "  ⚠️  WARNING: High DLQ depth. Investigation required."
  elif [ "$DLQ_DEPTH" -gt 0 ]; then
    echo "  ℹ️  INFO: DLQ contains messages. Review recommended."
  else
    echo "  ✅ OK: DLQ is empty."
  fi
done

echo ""
echo "=== End DLQ Health Check ==="
```

### Weekly DLQ Report

```bash
#!/bin/bash
# Weekly DLQ report for SRE review

echo "=== Weekly DLQ Report - $(date) ==="
echo ""

# Total events processed this week
TOTAL_EVENTS=$(curl -s "http://prometheus:9090/api/v1/query" \
  --data-urlencode "query=increase(cloudevents_processed_total[7d])" | \
  jq -r '.data.result[0].value[1]')

echo "Total Events Processed (7d): $TOTAL_EVENTS"

# Total events failed
FAILED_EVENTS=$(curl -s "http://prometheus:9090/api/v1/query" \
  --data-urlencode "query=increase(cloudevents_processing_errors_total[7d])" | \
  jq -r '.data.result[0].value[1]')

echo "Total Events Failed (7d): $FAILED_EVENTS"

# Failure rate
FAILURE_RATE=$(echo "scale=4; $FAILED_EVENTS / $TOTAL_EVENTS * 100" | bc)
echo "Failure Rate: ${FAILURE_RATE}%"

# Top failure types
echo ""
echo "Top Failure Types:"
curl -s "http://prometheus:9090/api/v1/query" \
  --data-urlencode "query=topk(5, increase(cloudevents_processing_errors_total[7d]))" | \
  jq -r '.data.result[] | "  \(.metric.error_type): \(.value[1])"'

# Top problematic parsers
echo ""
echo "Top Problematic Parsers:"
curl -s "http://prometheus:9090/api/v1/query" \
  --data-urlencode "query=topk(5, increase(parser_execution_errors_total[7d]))" | \
  jq -r '.data.result[] | "  \(.metric.parser_id): \(.value[1])"'

# DLQ replay stats
echo ""
echo "DLQ Replay Stats:"
REPLAYED_EVENTS=$(curl -s "http://prometheus:9090/api/v1/query" \
  --data-urlencode "query=increase(dlq_replay_success_total[7d])" | \
  jq -r '.data.result[0].value[1]')
echo "  Successfully Replayed: $REPLAYED_EVENTS events"

# Current DLQ depth
echo ""
echo "Current DLQ Depths:"
for ENV in dev stg prd; do
  DLQ_DEPTH=$(kubectl exec -n rabbitmq-$ENV rabbitmq-cluster-$ENV-0 -- \
    rabbitmqctl list_queues name messages | \
    grep "lambda-build-events-$ENV-dlq" | \
    awk '{print $2}')
  echo "  $ENV: $DLQ_DEPTH messages"
done

echo ""
echo "=== End Weekly DLQ Report ==="
```

---

## 📚 Related Documentation

- [SRE-003: Queue Management](./SRE-003-queue-management.md)
- [SRE-006: Disaster Recovery](./SRE-006-disaster-recovery.md)
- [SRE-009: Backup and Restore Operations](./SRE-009-backup-restore-operations.md)
- [BACKEND-001: CloudEvents Processing](../../backend/user-stories/BACKEND-001-cloudevents-processing.md)
- [Knative Eventing Documentation](https://knative.dev/docs/eventing/)
- [RabbitMQ DLQ Guide](https://www.rabbitmq.com/dlx.html)

---

## Revision History | Version | Date | Author | Changes | |--------- | ------ | -------- | --------- | | 1.0.0 | 2025-10-29 | Bruno Lucena (Principal SRE) | Initial comprehensive DLQ runbook |

