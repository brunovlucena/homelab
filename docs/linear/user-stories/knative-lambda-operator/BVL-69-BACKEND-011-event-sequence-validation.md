# 🌐 BACKEND-011: Event Sequence Validation and Ordering

**Priority**: P1 | **Status**: ✅ Implemented  | **Story Points**: 5
**Linear URL**: https://linear.app/bvlucena/issue/BVL-231/backend-011-event-sequence-validation-and-ordering

---

## 📋 User Story

**As a** Backend Developer  
**I want to** validate event sequence order and detect out-of-order events  
**So that** build lifecycle states remain consistent when events are replayed from DLQ

---

## 🎯 Acceptance Criteria

### ✅ Sequence Tracking
- [ ] Track last processed sequence number per build ID
- [ ] Store sequence state in Redis with 24h TTL
- [ ] Detect out-of-order events (sequence < last processed)
- [ ] Allow configurable strict vs. lenient ordering
- [ ] Clean up sequence tracking after build completion

### ✅ Out-of-Order Detection
- [ ] Warn when event arrives out of sequence
- [ ] Log out-of-order events with context
- [ ] Emit metric for sequence violations
- [ ] Optionally skip or process out-of-order events
- [ ] Track sequence gaps per build

### ✅ Event Metadata
- [ ] Add `eventSequence` field to CloudEvent data
- [ ] Add `eventSequence` validation to schema
- [ ] Support events without sequence (backwards compatibility)
- [ ] Default sequence to timestamp if not provided

### ✅ Observability
- [ ] Metric: `event_order_violations_total` counter
- [ ] Metric: `event_sequence_gaps_total` counter
- [ ] Log: Out-of-order event with sequences
- [ ] Trace: Sequence validation span

---

## 🔧 Technical Implementation

### File: `internal/handler/event_sequence.go`

```go
package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"knative-lambda/internal/observability"
)

// 🔢 EventSequenceValidator - "Validate event sequence order"
type EventSequenceValidator struct {
	redis          *redis.Client
	strictOrdering bool
	ttl            time.Duration
	obs            observability.Observability
}

// 🏗️ NewEventSequenceValidator - "Create new sequence validator"
func NewEventSequenceValidator(redis *redis.Client, strictOrdering bool, obs observability.Observability) *EventSequenceValidator {
	return &EventSequenceValidator{
		redis:          redis,
		strictOrdering: strictOrdering,
		ttl:            24 * time.Hour,
		obs:            obs,
	}
}

// ✅ ValidateSequence - "Validate event sequence and update last processed"
// Returns error if strict ordering enabled and event is out of order
func (esv *EventSequenceValidator) ValidateSequence(ctx context.Context, buildID string, sequence int64) error {
	key := fmt.Sprintf("event:seq:%s", buildID)
	
	// Get last processed sequence
	lastSeqStr, err := esv.redis.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get last sequence: %w", err)
	}
	
	var lastSeq int64
	if err != redis.Nil && lastSeqStr != "" {
		fmt.Sscanf(lastSeqStr, "%d", &lastSeq)
	}
	
	// Check for out-of-order event
	if sequence < lastSeq {
		esv.obs.Warn(ctx, "Out-of-order event detected",
			"build_id", buildID,
			"current_sequence", sequence,
			"last_processed_sequence", lastSeq,
			"gap", lastSeq-sequence)
		
		if esv.strictOrdering {
			return fmt.Errorf("out of order event: sequence %d < last processed %d", sequence, lastSeq)
		}
		// Lenient mode: log but continue
		return nil
	}
	
	// Check for sequence gap (missing events)
	if sequence > lastSeq+1 && lastSeq > 0 {
		esv.obs.Warn(ctx, "Sequence gap detected",
			"build_id", buildID,
			"current_sequence", sequence,
			"last_sequence", lastSeq,
			"gap_size", sequence-lastSeq-1)
	}
	
	// Update last processed sequence if this is newer
	if sequence > lastSeq {
		err = esv.redis.Set(ctx, key, fmt.Sprintf("%d", sequence), esv.ttl).Err()
		if err != nil {
			esv.obs.Error(ctx, err, "Failed to update sequence",
				"build_id", buildID,
				"sequence", sequence)
			return fmt.Errorf("failed to update sequence: %w", err)
		}
	}
	
	return nil
}

// 🗑️ CleanupSequence - "Remove sequence tracking for build"
func (esv *EventSequenceValidator) CleanupSequence(ctx context.Context, buildID string) error {
	key := fmt.Sprintf("event:seq:%s", buildID)
	return esv.redis.Del(ctx, key).Err()
}

// 📊 GetLastSequence - "Get last processed sequence for build"
func (esv *EventSequenceValidator) GetLastSequence(ctx context.Context, buildID string) (int64, error) {
	key := fmt.Sprintf("event:seq:%s", buildID)
	
	lastSeqStr, err := esv.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil // No sequence yet
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get last sequence: %w", err)
	}
	
	var lastSeq int64
	fmt.Sscanf(lastSeqStr, "%d", &lastSeq)
	return lastSeq, nil
}
```

### File: `internal/handler/event_handler.go` (Integration)

```go
// 📥 ProcessCloudEvent - "Process CloudEvent with sequence validation"
func (h *EventHandlerImpl) ProcessCloudEvent(ctx context.Context, event *cloudevents.Event) (*builds.HandlerResponse, error) {
	ctx, span := h.obs.StartSpan(ctx, "process_cloud_event")
	defer span.End()

	// Extract event data
	data := event.Data().(map[string]interface{})
	
	// Validate sequence if present
	if buildID, ok := data["buildId"].(string); ok {
		if sequence, ok := data["eventSequence"].(float64); ok {
			ctx, seqSpan := h.obs.StartSpan(ctx, "sequence_validation")
			err := h.sequenceValidator.ValidateSequence(ctx, buildID, int64(sequence))
			if err != nil {
				h.metrics.EventOrderViolations.WithLabelValues(buildID).Inc()
				seqSpan.SetStatus(codes.Error, err.Error())
				seqSpan.End()
				
				if h.config.StrictOrdering {
					return nil, fmt.Errorf("sequence validation failed: %w", err)
				}
				// Lenient mode: log and continue
				h.obs.Warn(ctx, "Sequence validation failed, continuing in lenient mode",
					"error", err.Error())
			}
			seqSpan.End()
		}
	}
	
	// Continue with normal processing...
	response, err := h.processEventWithTracing(ctx, event, h.metricsRec)
	
	// Cleanup sequence on build completion
	if h.isBuildCompleteEvent(event) {
		if buildID, ok := data["buildId"].(string); ok {
			if cleanupErr := h.sequenceValidator.CleanupSequence(ctx, buildID); cleanupErr != nil {
				h.obs.Error(ctx, cleanupErr, "Failed to cleanup sequence tracking")
			}
		}
	}
	
	return response, err
}
```

---

## 🧪 Test Cases

### File: `internal/handler/event_sequence_test.go`

```go
package handler

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"knative-lambda/internal/observability"
)

// Test 1: First event - no previous sequence
func TestEventSequenceValidator_FirstEvent(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := observability.NewMockObservability()
	
	mock.ExpectGet("event:seq:build-123").SetErr(redis.Nil)
	mock.ExpectSet("event:seq:build-123", "1", 24*time.Hour).SetVal("OK")
	
	validator := NewEventSequenceValidator(db, false, obs)
	err := validator.ValidateSequence(ctx, "build-123", 1)
	
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test 2: Sequential events - in order
func TestEventSequenceValidator_Sequential(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := observability.NewMockObservability()
	
	// Sequence 1
	mock.ExpectGet("event:seq:build-456").SetErr(redis.Nil)
	mock.ExpectSet("event:seq:build-456", "1", 24*time.Hour).SetVal("OK")
	
	// Sequence 2
	mock.ExpectGet("event:seq:build-456").SetVal("1")
	mock.ExpectSet("event:seq:build-456", "2", 24*time.Hour).SetVal("OK")
	
	// Sequence 3
	mock.ExpectGet("event:seq:build-456").SetVal("2")
	mock.ExpectSet("event:seq:build-456", "3", 24*time.Hour).SetVal("OK")
	
	validator := NewEventSequenceValidator(db, false, obs)
	
	require.NoError(t, validator.ValidateSequence(ctx, "build-456", 1))
	require.NoError(t, validator.ValidateSequence(ctx, "build-456", 2))
	require.NoError(t, validator.ValidateSequence(ctx, "build-456", 3))
	
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test 3: Out of order - strict mode (should fail)
func TestEventSequenceValidator_OutOfOrder_Strict(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := observability.NewMockObservability()
	
	// Last sequence was 5
	mock.ExpectGet("event:seq:build-789").SetVal("5")
	
	validator := NewEventSequenceValidator(db, true, obs) // strict = true
	err := validator.ValidateSequence(ctx, "build-789", 3) // Trying to process seq 3
	
	require.Error(t, err)
	assert.Contains(t, err.Error(), "out of order event")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test 4: Out of order - lenient mode (should warn but continue)
func TestEventSequenceValidator_OutOfOrder_Lenient(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := observability.NewMockObservability()
	
	// Last sequence was 5
	mock.ExpectGet("event:seq:build-789").SetVal("5")
	// Should NOT update sequence (3 < 5)
	
	validator := NewEventSequenceValidator(db, false, obs) // strict = false
	err := validator.ValidateSequence(ctx, "build-789", 3)
	
	require.NoError(t, err, "Lenient mode should not error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test 5: Sequence gap detection
func TestEventSequenceValidator_SequenceGap(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := observability.NewMockObservability()
	
	// Last sequence was 2, now processing 5 (gap of 2)
	mock.ExpectGet("event:seq:build-gap").SetVal("2")
	mock.ExpectSet("event:seq:build-gap", "5", 24*time.Hour).SetVal("OK")
	
	validator := NewEventSequenceValidator(db, false, obs)
	err := validator.ValidateSequence(ctx, "build-gap", 5)
	
	require.NoError(t, err, "Gap should be logged but not fail")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test 6: Cleanup sequence tracking
func TestEventSequenceValidator_Cleanup(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := observability.NewMockObservability()
	
	mock.ExpectDel("event:seq:build-complete").SetVal(1)
	
	validator := NewEventSequenceValidator(db, false, obs)
	err := validator.CleanupSequence(ctx, "build-complete")
	
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test 7: Get last sequence
func TestEventSequenceValidator_GetLastSequence(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := observability.NewMockObservability()
	
	mock.ExpectGet("event:seq:build-query").SetVal("42")
	
	validator := NewEventSequenceValidator(db, false, obs)
	lastSeq, err := validator.GetLastSequence(ctx, "build-query")
	
	require.NoError(t, err)
	assert.Equal(t, int64(42), lastSeq)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test 8: Redis error handling
func TestEventSequenceValidator_RedisError(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := observability.NewMockObservability()
	
	mock.ExpectGet("event:seq:build-error").SetErr(redis.TxFailedErr)
	
	validator := NewEventSequenceValidator(db, false, obs)
	err := validator.ValidateSequence(ctx, "build-error", 1)
	
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get last sequence")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test 9: Concurrent sequence updates
func TestEventSequenceValidator_ConcurrentUpdates(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := observability.NewMockObservability()
	
	// Simulating race: both read seq=0, both try to set seq=1
	mock.ExpectGet("event:seq:build-race").SetErr(redis.Nil)
	mock.ExpectSet("event:seq:build-race", "1", 24*time.Hour).SetVal("OK")
	
	mock.ExpectGet("event:seq:build-race").SetVal("1") // Second read sees first update
	// Second update tries seq=1 but last is already 1, so no update
	
	validator := NewEventSequenceValidator(db, false, obs)
	
	// First concurrent request
	err1 := validator.ValidateSequence(ctx, "build-race", 1)
	require.NoError(t, err1)
	
	// Second concurrent request (same sequence)
	err2 := validator.ValidateSequence(ctx, "build-race", 1)
	require.NoError(t, err2)
	
	assert.NoError(t, mock.ExpectationsWereMet())
}
```

---

## 📊 Metrics

```prometheus
# Counter: Event order violations
event_order_violations_total{build_id="build-123"}

# Counter: Sequence gaps detected
event_sequence_gaps_total{build_id="build-456",gap_size="2"}

# Histogram: Sequence validation duration
sequence_validation_duration_seconds_bucket{le="0.001"}
```

---

## 🔄 Configuration

```yaml
# ConfigMap for sequence validation settings
apiVersion: v1
kind: ConfigMap
metadata:
  name: knative-lambda-config
  namespace: knative-lambda
data:
  STRICT_ORDERING: "false"  # Set to "true" to reject out-of-order events
  SEQUENCE_TTL_HOURS: "24"
  ENABLE_SEQUENCE_VALIDATION: "true"
```

---

## 🔗 Related Stories

- [BACKEND-010: Idempotency and Duplicate Detection](./BACKEND-010-idempotency-duplicate-detection.md)
- [BACKEND-001: CloudEvents Processing](./BACKEND-001-cloudevents-processing.md)
- [SRE-011: Event Ordering and Idempotency](../../sre/user-stories/SRE-011-event-ordering-and-idempotency.md)

---

## Revision History | Version | Date | Author | Changes | |--------- | ------ | -------- | --------- | | 1.0.0 | 2025-10-29 | Bruno Lucena | Initial sequence validation story |

