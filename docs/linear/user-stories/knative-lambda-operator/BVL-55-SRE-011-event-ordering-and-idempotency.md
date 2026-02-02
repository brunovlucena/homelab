# 🔄 SRE-011: Event Ordering and Idempotency in DLQ Scenarios

**Status**: Backlog
**Priority**: P1
**Story Points**: 8  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-179/sre-011-event-ordering-and-idempotency-in-dlq-scenarios  
**Created**: 2026-01-19  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---


## 📋 User Story

**As a** SRE Engineer  
**I want to** event ordering and idempotency in dlq scenarios  
**So that** I can improve system reliability, security, and performance

---



## 🎯 Acceptance Criteria

- [ ] [ ] Event order preserved within same partition/context
- [ ] [ ] DLQ replay respects original event timestamps
- [ ] [ ] Out-of-order events detected and flagged
- [ ] [ ] Idempotency keys prevent duplicate processing
- [ ] [ ] Alert fires: "EventOrderViolation"
- [ ] [ ] Dashboard shows event sequence gaps
- [ ] [ ] Replay strategy documented per event type
- [ ] --

---


## Overview

This runbook addresses event ordering guarantees and idempotency concerns when events are processed, retried, and replayed from Dead Letter Queues. Ensures data consistency and prevents duplicate processing in distributed event-driven systems.

---

## 🎯 User Story: Maintain Event Order During DLQ Replay

### Story

**As an** SRE Engineer  
**I want** to maintain event ordering guarantees when replaying from DLQ  
**So that** business logic that depends on event sequence remains correct

### Acceptance Criteria

- [ ] Event order preserved within same partition/context
- [ ] DLQ replay respects original event timestamps
- [ ] Out-of-order events detected and flagged
- [ ] Idempotency keys prevent duplicate processing
- [ ] Alert fires: "EventOrderViolation"
- [ ] Dashboard shows event sequence gaps
- [ ] Replay strategy documented per event type

---

## 📊 Event Ordering Challenges

### Scenario 1: Build Lifecycle Events Out of Order

```yaml
Problem: Build lifecycle events must be processed in order

Expected Order:
  1. build.start    (id: build-123, seq: 1)
  2. build.progress (id: build-123, seq: 2)
  3. build.complete (id: build-123, seq: 3)
  4. service.create (id: build-123, seq: 4)

Actual Order After DLQ Replay:
  1. build.start    (seq: 1) → Success
  2. build.progress (seq: 2) → Failed → DLQ
  3. build.complete (seq: 3) → Success (processed before seq: 2!)
  4. service.create (seq: 4) → Success
  5. build.progress (seq: 2) → Replayed from DLQ (out of order!)

Impact:
  - Service created before build fully validated
  - Progress metrics incorrect
  - State machine confused
```

### Detection

```bash
# Query for out-of-order events
kubectl logs -n knative-lambda -l app=knative-lambda-builder | \
  jq -r 'select(.event_sequence) | [.build_id, .event_sequence, .event_type, .timestamp] | @csv' | \
  sort -t',' -k1,1 -k2,2n > events_by_sequence.csv

# Identify sequence gaps
awk -F',' '
  {
    if ($1 == prev_build && $2 != prev_seq + 1) {
      print "Gap detected: Build " $1 " jumped from seq " prev_seq " to " $2
    }
    prev_build = $1
    prev_seq = $2
  }
' events_by_sequence.csv

# Check DLQ for events with earlier sequences than processed events
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin get queue=lambda-build-events-prd-dlq count=100 | \
  jq -r '.[].payload.data | [.buildId, .eventSequence, .eventType] | @csv'
```

### Remediation

```go
// Add event sequencing validation in event handler
type EventSequence struct {
    BuildID       string
    LastSequence  int64
    LastTimestamp time.Time
    ProcessedIDs  map[string]bool  // Idempotency tracking
}

func (h *EventHandlerImpl) ValidateEventOrder(ctx context.Context, event *cloudevents.Event) error {
    buildID := event.Data().(map[string]interface{})["buildId"].(string)
    sequence := event.Data().(map[string]interface{})["eventSequence"].(int64)
    
    // Get last processed sequence for this build
    lastSeq, err := h.redis.Get(ctx, fmt.Sprintf("build:seq:%s", buildID)).Int64()
    if err != nil && err != redis.Nil {
        return fmt.Errorf("failed to get last sequence: %w", err)
    }
    
    // Check for out-of-order event
    if sequence < lastSeq {
        h.obs.Warn(ctx, "Out-of-order event detected",
            "build_id", buildID,
            "current_sequence", sequence,
            "last_processed_sequence", lastSeq)
        
        // Record metric
        h.metrics.EventOrderViolations.WithLabelValues(buildID).Inc()
        
        // Decide: skip or process anyway
        if h.config.StrictOrdering {
            return errors.NewValidationError("event_sequence", sequence, 
                fmt.Sprintf("out of order (last=%d)", lastSeq))
        }
    }
    
    // Update last processed sequence
    if sequence > lastSeq {
        err = h.redis.Set(ctx, fmt.Sprintf("build:seq:%s", buildID), sequence, 24*time.Hour).Err()
        if err != nil {
            h.obs.Error(ctx, err, "Failed to update sequence",
                "build_id", buildID,
                "sequence", sequence)
        }
    }
    
    return nil
}
```

### Prevention Strategy

```yaml
# Add sequence metadata to all events
apiVersion: v1
kind: ConfigMap
metadata:
  name: event-sequence-config
data:
  enable_sequence_validation: "true"
  strict_ordering: "false"  # Warn but don't reject
  sequence_tracking_ttl: "24h"
  
---
# Redis for sequence tracking
apiVersion: v1
kind: Service
metadata:
  name: event-sequence-redis
spec:
  selector:
    app: redis-sequence
  ports:
  - port: 6379
    targetPort: 6379
```

---

## 🔑 Idempotency Guarantees

### Scenario 2: Duplicate Event Processing

```yaml
Problem: Event replayed from DLQ causes duplicate processing

Event Timeline:
  T+0s: Event "build-456-start" published
  T+1s: Consumer A receives event
  T+2s: Consumer A processes event → Database write
  T+3s: Consumer A crashes before ACK
  T+4s: Event re-queued (not ACK'd)
  T+5s: Consumer B receives same event
  T+6s: Consumer B processes event → Duplicate database write!
  
  T+1h: Original event in DLQ (timeout)
  T+1h+5m: DLQ replay → Event processed AGAIN → Third duplicate!

Impact:
  - Duplicate builds created
  - Duplicate service deployments
  - Wasted resources
  - Incorrect metrics
```

### Detection

```bash
# Find duplicate build IDs in database
kubectl exec -n database postgres-0 -- psql -U postgres -d lambdadb -c "
  SELECT build_id, COUNT(*) as count
  FROM builds
  GROUP BY build_id
  HAVING COUNT(*) > 1
  ORDER BY count DESC
  LIMIT 20;
"

# Find duplicate event processing in logs
kubectl logs -n knative-lambda -l app=knative-lambda-builder --since=24h | \
  jq -r 'select(.event_id) | .event_id' | \
  sort | uniq -c | \
  awk '$1 > 1 {print $2, "processed", $1, "times"}'

# Check for duplicate CloudEvent IDs
kubectl logs -n knative-lambda -l app=knative-lambda-builder --since=24h | \
  jq -r 'select(.cloud_event_id) | [.cloud_event_id, .timestamp, .event_type] | @csv' | \
  sort -t',' -k1,1 | \
  awk -F',' '{if ($1 == prev) print "Duplicate:", $0; prev=$1}'
```

### Remediation - Idempotency Key Implementation

```go
// Add idempotency layer to event processing
type IdempotencyChecker struct {
    redis  *redis.Client
    ttl    time.Duration
}

func NewIdempotencyChecker(redis *redis.Client) *IdempotencyChecker {
    return &IdempotencyChecker{
        redis: redis,
        ttl:   24 * time.Hour,  // Keep idempotency keys for 24h
    }
}

func (ic *IdempotencyChecker) CheckAndMark(ctx context.Context, eventID string) (bool, error) {
    key := fmt.Sprintf("idempotency:event:%s", eventID)
    
    // Try to set key if not exists (NX flag)
    result, err := ic.redis.SetNX(ctx, key, time.Now().Unix(), ic.ttl).Result()
    if err != nil {
        return false, fmt.Errorf("failed to check idempotency: %w", err)
    }
    
    // If SetNX returned false, key already exists → duplicate
    if !result {
        return true, nil  // true = is duplicate
    }
    
    return false, nil  // false = first time processing
}

func (h *EventHandlerImpl) ProcessCloudEvent(ctx context.Context, event *cloudevents.Event) (*builds.HandlerResponse, error) {
    // Check idempotency before processing
    isDuplicate, err := h.idempotency.CheckAndMark(ctx, event.ID())
    if err != nil {
        h.obs.Error(ctx, err, "Idempotency check failed", "event_id", event.ID())
        // Continue processing - better to risk duplicate than lose event
    }
    
    if isDuplicate {
        h.obs.Info(ctx, "Duplicate event detected, skipping processing",
            "event_id", event.ID(),
            "event_type", event.Type(),
            "event_source", event.Source())
        
        // Record metric
        h.metrics.DuplicateEventsSkipped.WithLabelValues(event.Type()).Inc()
        
        // Return success without processing
        return &builds.HandlerResponse{
            Status:  "skipped",
            Message: "Duplicate event",
            EventID: event.ID(),
        }, nil
    }
    
    // Process event normally
    response, err := h.processEventWithTracing(ctx, event, h.metricsRec)
    if err != nil {
        // On failure, remove idempotency key to allow retry
        h.idempotency.redis.Del(ctx, fmt.Sprintf("idempotency:event:%s", event.ID()))
        return nil, err
    }
    
    return response, nil
}
```

### Database-Level Idempotency

```sql
-- Add unique constraint on build_id to prevent duplicates
ALTER TABLE builds
ADD CONSTRAINT builds_build_id_unique UNIQUE (build_id);

-- Use UPSERT pattern for idempotent writes
INSERT INTO builds (
    build_id,
    third_party_id,
    parser_id,
    status,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, 'pending', NOW(), NOW()
)
ON CONFLICT (build_id) DO UPDATE SET
    status = EXCLUDED.status,
    updated_at = NOW()
RETURNING *;

-- Create idempotency tracking table
CREATE TABLE event_idempotency (
    event_id VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processor_pod VARCHAR(255),
    result JSONB,
    INDEX idx_event_type_processed_at (event_type, processed_at)
);

-- Set TTL on idempotency records (PostgreSQL + pg_cron)
CREATE EXTENSION IF NOT EXISTS pg_cron;

SELECT cron.schedule(
    'cleanup-old-idempotency-records',
    '0 * * * *',  -- Every hour
    'DELETE FROM event_idempotency WHERE processed_at < NOW() - INTERVAL ''24 hours'''
);
```

---

## 🔄 Ordered DLQ Replay Strategy

### FIFO Queue Configuration

```yaml
# RabbitMQ Quorum Queue with FIFO guarantees
apiVersion: eventing.knative.dev/v1alpha1
kind: RabbitmqBrokerConfig
metadata:
  name: lambda-broker-config-prd
  namespace: knative-lambda
spec:
  rabbitmqClusterReference:
    name: rabbitmq-cluster-prd
    namespace: rabbitmq-prd
  queueType: quorum
  delivery:
    # Single active consumer ensures ordering
    singleActiveConsumer: true
    # Limit prefetch to 1 for strict ordering
    prefetchCount: 1
  arguments:
    # Dead letter exchange
    x-dead-letter-exchange: lambda-dlx-prd
    x-dead-letter-routing-key: lambda-build-events-dlq
    # Preserve message order
    x-single-active-consumer: true
    # Quorum queue for HA
    x-queue-type: quorum
```

### Ordered Replay Script

```bash
#!/bin/bash
# Replay DLQ events in strict order

DLQ_NAME="lambda-build-events-prd-dlq"
TARGET_EXCHANGE="knative-lambda-broker-prd"
ROUTING_KEY="lambda-build-events"

echo "Starting ordered DLQ replay..."

# Get all messages from DLQ (oldest first)
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin get queue=$DLQ_NAME count=1000 | \
  jq -r '.[] | @base64' > dlq_messages_b64.txt

# Sort by original timestamp (preserved in headers)
while IFS= read -r msg_b64; do
  msg=$(echo "$msg_b64" | base64 -d)
  timestamp=$(echo "$msg" | jq -r '.properties.timestamp')
  sequence=$(echo "$msg" | jq -r '.payload.data.eventSequence // 0')
  build_id=$(echo "$msg" | jq -r '.payload.data.buildId // "unknown"')
  
  echo "$timestamp | $sequence | $build_id | $msg_b64"
done < dlq_messages_b64.txt | \
  sort -t' | ' -k1,1 -k2,2n > dlq_messages_sorted.txt

# Replay in order
echo "Replaying messages in timestamp order..."
REPLAYED=0

while IFS=' | ' read -r timestamp sequence build_id msg_b64; do
  msg=$(echo "$msg_b64" | base64 -d)
  
  echo "Replaying: build=$build_id, seq=$sequence, time=$timestamp"
  
  # Publish to main queue
  echo "$msg" | jq -r '.payload' | \
    kubectl exec -i -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
    rabbitmqadmin publish \
    exchange=$TARGET_EXCHANGE \
    routing_key=$ROUTING_KEY \
    payload=-
  
  REPLAYED=$((REPLAYED + 1))
  
  # Wait for processing to complete before next message (strict ordering)
  sleep 1
  
done < dlq_messages_sorted.txt

echo "Ordered replay complete. Replayed $REPLAYED messages."
```

---

## 📊 Monitoring & Alerts

```prometheus
# Alert: Event Order Violation
- alert: EventOrderViolation
  expr: | rate(event_order_violations_total[5m]) > 0
  for: 1m
  severity: warning
  annotations:
    summary: "Event order violations detected"
    description: "{{ $value }} out-of-order events detected. Check DLQ replay strategy."

# Alert: Duplicate Event Processing
- alert: DuplicateEventProcessing
  expr: | rate(duplicate_events_skipped_total[5m]) > 0.5
  for: 5m
  severity: warning
  annotations:
    summary: "High rate of duplicate events"
    description: "{{ $value }} duplicate events/sec. Check idempotency implementation."

# Metric: Sequence Gap Detection
- metric: event_sequence_gaps_total
  expr: | increase(event_sequence_gaps_total[1h])

# Dashboard: Idempotency Stats
- panel: "Idempotency Hit Rate"
  expr: | rate(duplicate_events_skipped_total[5m]) / rate(cloudevents_received_total[5m])
```

---

## 🔧 Operational Guidelines

### When to Use Strict Ordering

**Use strict ordering (single active consumer) when:**
- Events form a state machine (build lifecycle)
- Later events depend on earlier events
- Order violations cause data corruption
- Business logic requires sequential processing

**Use parallel processing when:**
- Events are independent
- High throughput required
- Idempotency handles duplicates
- Order doesn't matter for correctness

### Idempotency Best Practices

1. **Always use CloudEvent ID as idempotency key**
2. **Store idempotency state in Redis with TTL**
3. **Add unique constraints at database level**
4. **Use UPSERT patterns for database writes**
5. **Return 200 OK for duplicate events (not 409)**
6. **Log duplicate events for monitoring**
7. **Clean up idempotency keys after 24h**

### DLQ Replay Checklist

- [ ] Check for related events in main queue
- [ ] Sort DLQ messages by timestamp
- [ ] Group by build_id or context_id
- [ ] Replay in batches by group
- [ ] Monitor for duplicates during replay
- [ ] Verify state consistency after replay
- [ ] Clean up successfully replayed messages

---

## 📚 Related Documentation

- [SRE-010: Dead Letter Queue Management](./SRE-010-dead-letter-queue-management.md)
- [SRE-003: Queue Management](./SRE-003-queue-management.md)
- [CloudEvents Specification](https://cloudevents.io/)
- [RabbitMQ Quorum Queues](https://www.rabbitmq.com/quorum-queues.html)

---

## Revision History | Version | Date | Author | Changes | |--------- | ------ | -------- | --------- | | 1.0.0 | 2025-10-29 | Bruno Lucena (Principal SRE) | Initial event ordering and idempotency runbook |

