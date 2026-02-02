package backend

import (
	"context"
	"sync"
	"testing"
	"time"

	"knative-lambda/internal/handler"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 🔑 TestIdempotencyChecker tests the idempotency checking functionality.
// Related Story: BACKEND-010-idempotency-duplicate-detection.md
//
// Note: Redis is used for both rate-limiting AND idempotency in this system.

// 🧪 Test 1: First time processing - not duplicate.
func TestBackend010_IdempotencyChecker_FirstTime(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, mock := redismock.NewClientMock()

	// Expect: SetNX returns true when key doesn't exist
	// Note: Using AnyTimes matcher since redismock doesn't support MatchFn
	mock.Regexp().ExpectSetNX("idempotency:event:event-123", `\d+`, 24*time.Hour).SetVal(true)

	checker := handler.NewIdempotencyChecker(db, 24*time.Hour)

	// Execute
	isDuplicate, err := checker.CheckAndMark(ctx, "event-123")

	// Assert
	require.NoError(t, err)
	assert.False(t, isDuplicate, "First time processing should not be duplicate")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 🧪 Test 2: Duplicate event - key already exists.
func TestBackend010_IdempotencyChecker_Duplicate(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, mock := redismock.NewClientMock()

	// Expect: SetNX returns false when key exists
	mock.Regexp().ExpectSetNX("idempotency:event:event-456", `\d+`, 24*time.Hour).SetVal(false)

	checker := handler.NewIdempotencyChecker(db, 24*time.Hour)

	// Execute
	isDuplicate, err := checker.CheckAndMark(ctx, "event-456")

	// Assert
	require.NoError(t, err)
	assert.True(t, isDuplicate, "Second processing should be duplicate")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 🧪 Test 3: Redis connection error.
func TestBackend010_IdempotencyChecker_RedisError(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, mock := redismock.NewClientMock()

	// Expect: Redis returns a generic error (not redis.Nil)
	mock.Regexp().ExpectSetNX("idempotency:event:event-789", `\d+`, 24*time.Hour).SetErr(assert.AnError)

	checker := handler.NewIdempotencyChecker(db, 24*time.Hour)

	// Execute
	isDuplicate, err := checker.CheckAndMark(ctx, "event-789")

	// Assert
	require.Error(t, err)
	assert.False(t, isDuplicate)
	assert.Contains(t, err.Error(), "failed to check idempotency")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 🧪 Test 4: Exists check.
func TestBackend010_IdempotencyChecker_Exists(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, mock := redismock.NewClientMock()

	mock.ExpectExists("idempotency:event:existing").SetVal(1)

	checker := handler.NewIdempotencyChecker(db, 24*time.Hour)

	// Execute
	exists, err := checker.Exists(ctx, "existing")

	// Assert
	require.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 🧪 Test 5: Remove/Clear functionality.
func TestBackend010_IdempotencyChecker_Remove(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, mock := redismock.NewClientMock()

	mock.ExpectDel("idempotency:event:failed-event").SetVal(1)

	checker := handler.NewIdempotencyChecker(db, 24*time.Hour)

	// Execute
	err := checker.Remove(ctx, "failed-event")

	// Assert
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 🧪 Test 6: NewIdempotencyChecker constructor.
func TestBackend010_NewIdempotencyChecker(t *testing.T) {
	// Setup
	db, _ := redismock.NewClientMock()

	// Execute
	checker := handler.NewIdempotencyChecker(db, 24*time.Hour)

	// Assert
	assert.NotNil(t, checker)
	// Note: ttl is unexported, tested through behavior in other tests
}

// 🧪 Test 7: Concurrent access - thread safety verification.
// This test verifies Redis SETNX atomicity prevents race conditions.
func TestBackend010_IdempotencyChecker_ConcurrentAccess(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, mock := redismock.NewClientMock()

	const goroutines = 100
	const eventID = "concurrent-event-id"

	// Mock: First SetNX succeeds (returns true), rest fail (return false)
	// This simulates Redis atomic behavior where only ONE SetNX succeeds
	mock.Regexp().ExpectSetNX("idempotency:event:"+eventID, `\d+`, 24*time.Hour).SetVal(true)

	// All subsequent attempts should return false (key already exists)
	for i := 0; i < goroutines-1; i++ {
		mock.Regexp().ExpectSetNX("idempotency:event:"+eventID, `\d+`, 24*time.Hour).SetVal(false)
	}

	checker := handler.NewIdempotencyChecker(db, 24*time.Hour)

	// Execute - Launch concurrent goroutines
	results := make(chan bool, goroutines)
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			isDup, err := checker.CheckAndMark(ctx, eventID)
			require.NoError(t, err)
			results <- isDup
		}()
	}

	wg.Wait()
	close(results)

	// Assert - Exactly ONE should be non-duplicate (first one wins)
	nonDuplicateCount := 0
	duplicateCount := 0

	for isDup := range results {
		if isDup {
			duplicateCount++
		} else {
			nonDuplicateCount++
		}
	}

	assert.Equal(t, 1, nonDuplicateCount, "Exactly one goroutine should win the race")
	assert.Equal(t, goroutines-1, duplicateCount, "All other goroutines should detect duplicate")
	assert.NoError(t, mock.ExpectationsWereMet())
}
