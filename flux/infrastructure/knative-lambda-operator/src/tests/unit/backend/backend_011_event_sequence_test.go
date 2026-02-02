package backend

import (
	"context"
	"testing"
	"time"

	"knative-lambda/internal/handler"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 🔢 TestEventSequenceValidator tests event sequence validation and ordering.
// Related Story: BACKEND-011-event-sequence-validation.md
//
// Note: Redis is used for sequence tracking to ensure event ordering.

// 🧪 Test 1: First event - no previous sequence.
func TestBackend011_EventSequenceValidator_FirstEvent(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := handler.NewMockObservability()

	mock.ExpectGet("event:seq:build-123").SetErr(redis.Nil)
	mock.ExpectSet("event:seq:build-123", uint64(1), 24*time.Hour).SetVal("OK")

	validator := handler.NewEventSequenceValidator(db, false, 24*time.Hour, obs)

	// Execute
	err := validator.ValidateSequence(ctx, "build-123", 1)

	// Assert
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 🧪 Test 2: Sequential events - in order.
func TestBackend011_EventSequenceValidator_Sequential(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := handler.NewMockObservability()

	// Sequence 1 (first)
	mock.ExpectGet("event:seq:build-456").SetErr(redis.Nil)
	mock.ExpectSet("event:seq:build-456", uint64(1), 24*time.Hour).SetVal("OK")

	// Sequence 2 (second)
	mock.ExpectGet("event:seq:build-456").SetVal("1")
	mock.ExpectSet("event:seq:build-456", uint64(2), 24*time.Hour).SetVal("OK")

	// Sequence 3 (third)
	mock.ExpectGet("event:seq:build-456").SetVal("2")
	mock.ExpectSet("event:seq:build-456", uint64(3), 24*time.Hour).SetVal("OK")

	validator := handler.NewEventSequenceValidator(db, false, 24*time.Hour, obs)

	// Execute
	err1 := validator.ValidateSequence(ctx, "build-456", 1)
	err2 := validator.ValidateSequence(ctx, "build-456", 2)
	err3 := validator.ValidateSequence(ctx, "build-456", 3)

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 🧪 Test 3: Out of order event - should fail in strict mode.
func TestBackend011_EventSequenceValidator_OutOfOrder(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := handler.NewMockObservability()

	// Last sequence was 5, trying to send sequence 3 (out of order)
	mock.ExpectGet("event:seq:build-789").SetVal("5")

	validator := handler.NewEventSequenceValidator(db, false, 24*time.Hour, obs)

	// Execute
	err := validator.ValidateSequence(ctx, "build-789", 3)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "out of order")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 🧪 Test 4: GetLastSequence.
func TestBackend011_EventSequenceValidator_GetLastSequence(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := handler.NewMockObservability()

	mock.ExpectGet("event:seq:build-999").SetVal("42")

	validator := handler.NewEventSequenceValidator(db, false, 24*time.Hour, obs)

	// Execute
	lastSeq, exists, err := validator.GetLastSequence(ctx, "build-999")

	// Assert
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, uint64(42), lastSeq)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 🧪 Test 5: Reset sequence.
func TestBackend011_EventSequenceValidator_Reset(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	obs := handler.NewMockObservability()

	mock.ExpectDel("event:seq:build-reset").SetVal(1)

	validator := handler.NewEventSequenceValidator(db, false, 24*time.Hour, obs)

	// Execute
	err := validator.Reset(ctx, "build-reset")

	// Assert
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
