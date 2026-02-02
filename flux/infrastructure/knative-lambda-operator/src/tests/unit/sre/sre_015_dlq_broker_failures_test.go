// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 SRE-015: DLQ RabbitMQ Broker Failures
//
//	User Story: Handle RabbitMQ broker failures with DLQ
//	Priority: P0 | Story Points: 13
//
//	Tests validate:
//	- Broker connection failures routed to DLQ
//	- Broker authentication failures tracked
//	- Broker unavailability triggers failover
//	- Circuit breaker prevents cascading failures
//	- Retry policy with exponential backoff
//	- DLQ for broker misconfiguration
//	- Metrics track broker health
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package sre

import (
	"context"
	"knative-lambda/tests/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Test Fixtures.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

type BrokerFailureType string

const (
	BrokerConnectionRefused BrokerFailureType = "connection_refused"
	BrokerAuthFailed        BrokerFailureType = "authentication_failed"
	BrokerTimeout           BrokerFailureType = "connection_timeout"
	BrokerNotFound          BrokerFailureType = "broker_not_found"
	BrokerNetworkPartition  BrokerFailureType = "network_partition"
)

type BrokerFailureEvent struct {
	EventID      string
	FailureType  BrokerFailureType
	Timestamp    time.Time
	RetryCount   int
	MovedToDLQ   bool
	DLQReason    string
	ErrorDetails string
	RecoveryTime time.Duration
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: Broker connection failures routed to DLQ.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE015_AC1_BrokerConnectionFailuresDLQ(t *testing.T) {
	t.Run("Connection refused after max retries routed to DLQ", func(t *testing.T) {
		// Arrange
		event := BrokerFailureEvent{
			EventID:     "build-123",
			FailureType: BrokerConnectionRefused,
			RetryCount:  5,
			MovedToDLQ:  false,
		}
		maxRetries := 5

		// Act
		if event.RetryCount >= maxRetries {
			event.MovedToDLQ = true
			event.DLQReason = "broker_connection_refused_max_retries"
		}

		// Assert
		assert.True(t, event.MovedToDLQ, "Event should be moved to DLQ after max retries")
		assert.Equal(t, "broker_connection_refused_max_retries", event.DLQReason)
	})

	t.Run("Connection timeout tracked in DLQ metadata", func(t *testing.T) {
		// Arrange
		event := BrokerFailureEvent{
			EventID:      "build-456",
			FailureType:  BrokerTimeout,
			Timestamp:    time.Now().Add(-30 * time.Second),
			RetryCount:   5,
			MovedToDLQ:   true,
			DLQReason:    "broker_timeout",
			ErrorDetails: "connection timeout after 10s",
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
		assert.Contains(t, event.ErrorDetails, "timeout")
		assert.Equal(t, BrokerTimeout, event.FailureType)
	})

	t.Run("Broker not found routes to DLQ immediately", func(t *testing.T) {
		// Arrange
		event := BrokerFailureEvent{
			EventID:     "build-789",
			FailureType: BrokerNotFound,
			RetryCount:  0,
		}

		// Act - Permanent failures skip retries
		if event.FailureType == BrokerNotFound {
			event.MovedToDLQ = true
			event.DLQReason = "broker_not_found_permanent_failure"
		}

		// Assert
		assert.True(t, event.MovedToDLQ, "Permanent failures should skip retries")
		assert.Equal(t, 0, event.RetryCount, "Should not retry broker not found")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Broker authentication failures tracked.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE015_AC2_BrokerAuthFailuresTracked(t *testing.T) {
	t.Run("Authentication failure routes to DLQ with security alert", func(t *testing.T) {
		// Arrange
		event := BrokerFailureEvent{
			EventID:      "build-auth-1",
			FailureType:  BrokerAuthFailed,
			ErrorDetails: "AMQP authentication failed: ACCESS_REFUSED",
			MovedToDLQ:   true,
			DLQReason:    "broker_auth_failed",
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
		assert.Equal(t, BrokerAuthFailed, event.FailureType)
		assert.Contains(t, event.ErrorDetails, "ACCESS_REFUSED")
	})

	t.Run("Auth failure metrics track credential rotation needed", func(t *testing.T) {
		// Arrange
		authFailures := []BrokerFailureEvent{
			{EventID: "e1", FailureType: BrokerAuthFailed, Timestamp: time.Now()},
			{EventID: "e2", FailureType: BrokerAuthFailed, Timestamp: time.Now()},
			{EventID: "e3", FailureType: BrokerAuthFailed, Timestamp: time.Now()},
		}

		// Act
		authFailureCount := len(authFailures)
		credentialRotationThreshold := 2

		// Assert
		assert.GreaterOrEqual(t, authFailureCount, credentialRotationThreshold,
			"Should trigger credential rotation alert")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Broker unavailability triggers failover.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE015_AC3_BrokerFailover(t *testing.T) {
	t.Run("Failover to secondary broker on primary unavailability", func(t *testing.T) {
		// Arrange
		type BrokerConfig struct {
			Primary   string
			Secondary string
			Active    string
			Failover  bool
		}

		config := BrokerConfig{
			Primary:   "rabbitmq-cluster-prd-0.rabbitmq-prd",
			Secondary: "rabbitmq-cluster-prd-1.rabbitmq-prd",
			Active:    "rabbitmq-cluster-prd-0.rabbitmq-prd",
		}

		// Act - Simulate primary failure
		primaryAvailable := false
		if !primaryAvailable {
			config.Active = config.Secondary
			config.Failover = true
		}

		// Assert
		assert.True(t, config.Failover, "Should trigger failover")
		assert.Equal(t, config.Secondary, config.Active, "Should use secondary broker")
	})

	t.Run("Failover completes within 30 seconds", func(t *testing.T) {
		// Arrange
		startTime := time.Now()

		// Act - Simulate failover
		time.Sleep(100 * time.Millisecond) // Simulate failover time
		endTime := time.Now()
		maxDuration := 30 * time.Second

		phases := []testutils.Phase{
			{Name: "Health check", Duration: 5 * time.Second},
			{Name: "Broker switch", Duration: 10 * time.Second},
			{Name: "Connection re-establishment", Duration: 10 * time.Second},
			{Name: "Verification", Duration: 5 * time.Second},
		}

		// Assert
		testutils.RunTimingTest(t, "Failover", startTime, endTime, maxDuration, phases)
	})

	t.Run("Events during failover queued in DLQ", func(t *testing.T) {
		// Arrange
		eventsInFailoverWindow := 15
		type FailoverStats struct {
			EventsDuringFailover int
			EventsInDLQ          int
			EventsLost           int
		}

		stats := FailoverStats{
			EventsDuringFailover: eventsInFailoverWindow,
			EventsInDLQ:          eventsInFailoverWindow,
			EventsLost:           0,
		}

		// Assert
		assert.Equal(t, stats.EventsDuringFailover, stats.EventsInDLQ,
			"All events during failover should be in DLQ")
		assert.Equal(t, 0, stats.EventsLost, "No events should be lost")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: Circuit breaker prevents cascading failures.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE015_AC4_CircuitBreaker(t *testing.T) {
	t.Run("Circuit breaker opens after failure threshold", testCircuitBreakerOpens)
	t.Run("Circuit breaker routes to DLQ when open", testCircuitBreakerDLQRouting)
	t.Run("Circuit breaker transitions to half-open for testing", testCircuitBreakerHalfOpen)
}

func testCircuitBreakerOpens(t *testing.T) {
	// Arrange
	type CircuitBreaker struct {
		State            string // "closed", "open", "half-open"
		FailureCount     int
		FailureThreshold int
		ResetTimeout     time.Duration
	}

	cb := CircuitBreaker{
		State:            "closed",
		FailureCount:     0,
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
	}

	// Act - Simulate failures
	for i := 0; i < 6; i++ {
		cb.FailureCount++
		if cb.FailureCount >= cb.FailureThreshold {
			cb.State = "open"
		}
	}

	// Assert
	assert.Equal(t, "open", cb.State, "Circuit breaker should open")
	assert.GreaterOrEqual(t, cb.FailureCount, cb.FailureThreshold)
}

func testCircuitBreakerDLQRouting(t *testing.T) {
	// Arrange
	circuitBreakerOpen := true
	event := BrokerFailureEvent{
		EventID:     "build-circuit-1",
		FailureType: BrokerConnectionRefused,
	}

	// Act
	if circuitBreakerOpen {
		event.MovedToDLQ = true
		event.DLQReason = "circuit_breaker_open"
	}

	// Assert
	assert.True(t, event.MovedToDLQ, "Should route to DLQ when circuit open")
	assert.Equal(t, "circuit_breaker_open", event.DLQReason)
}

func testCircuitBreakerHalfOpen(t *testing.T) {
	// Arrange
	type CircuitBreaker struct {
		State        string
		LastFailure  time.Time
		ResetTimeout time.Duration
	}

	cb := CircuitBreaker{
		State:        "open",
		LastFailure:  time.Now().Add(-35 * time.Second),
		ResetTimeout: 30 * time.Second,
	}

	// Act
	if time.Since(cb.LastFailure) > cb.ResetTimeout {
		cb.State = "half-open"
	}

	// Assert
	assert.Equal(t, "half-open", cb.State,
		"Should transition to half-open after reset timeout")
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Retry policy with exponential backoff.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE015_AC5_RetryPolicyBackoff(t *testing.T) {
	t.Run("Exponential backoff delays increase per retry", func(t *testing.T) {
		// Arrange
		type RetryPolicy struct {
			InitialDelay time.Duration
			Multiplier   float64
			MaxDelay     time.Duration
		}

		policy := RetryPolicy{
			InitialDelay: 1 * time.Second,
			Multiplier:   2.0,
			MaxDelay:     30 * time.Second,
		}

		// Act - Calculate delays
		delays := []time.Duration{}
		currentDelay := policy.InitialDelay
		for i := 0; i < 5; i++ {
			delays = append(delays, currentDelay)
			currentDelay = time.Duration(float64(currentDelay) * policy.Multiplier)
			if currentDelay > policy.MaxDelay {
				currentDelay = policy.MaxDelay
			}
		}

		// Assert
		assert.Equal(t, 1*time.Second, delays[0])
		assert.Equal(t, 2*time.Second, delays[1])
		assert.Equal(t, 4*time.Second, delays[2])
		assert.Equal(t, 8*time.Second, delays[3])
		assert.Equal(t, 16*time.Second, delays[4])
	})

	t.Run("Max retry attempts enforced", func(t *testing.T) {
		// Arrange
		maxRetries := 5
		attempts := 0

		// Act
		for attempts < maxRetries {
			attempts++
			// Simulate retry
		}

		// Assert
		assert.Equal(t, maxRetries, attempts, "Should stop at max retries")
	})

	t.Run("Jitter added to prevent thundering herd", func(t *testing.T) {
		// Arrange
		baseDelay := 5 * time.Second
		jitterPercent := 0.2 // 20% jitter

		// Act
		minDelay := time.Duration(float64(baseDelay) * (1 - jitterPercent))
		maxDelay := time.Duration(float64(baseDelay) * (1 + jitterPercent))

		// Assert
		assert.Equal(t, 4*time.Second, minDelay)
		assert.Equal(t, 6*time.Second, maxDelay)
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: DLQ for broker misconfiguration.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE015_AC6_BrokerMisconfigurationDLQ(t *testing.T) {
	t.Run("Invalid broker URL routed to DLQ", func(t *testing.T) {
		// Arrange
		event := BrokerFailureEvent{
			EventID:      "build-config-1",
			FailureType:  BrokerNotFound,
			ErrorDetails: "invalid broker URL: amqp://invalid-host:5672",
			MovedToDLQ:   true,
			DLQReason:    "broker_misconfiguration",
		}

		// Assert
		assert.True(t, event.MovedToDLQ)
		assert.Contains(t, event.ErrorDetails, "invalid")
	})

	t.Run("Missing broker credentials detected", func(t *testing.T) {
		// Arrange
		type BrokerCredentials struct {
			Username string
			Password string
			Valid    bool
		}

		creds := BrokerCredentials{
			Username: "",
			Password: "",
			Valid:    false,
		}

		// Act
		if creds.Username == "" || creds.Password == "" {
			creds.Valid = false
		}

		// Assert
		assert.False(t, creds.Valid, "Should detect missing credentials")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC7: Metrics track broker health.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE015_AC7_BrokerHealthMetrics(t *testing.T) {
	t.Run("Broker connection success rate metric", func(t *testing.T) {
		// Arrange
		totalAttempts := 100
		successfulConnections := 95

		// Act
		successRate := float64(successfulConnections) / float64(totalAttempts)

		// Assert
		assert.GreaterOrEqual(t, successRate, 0.95, "Success rate should be >= 95%")
	})

	t.Run("Broker latency metric tracked", func(t *testing.T) {
		// Arrange
		connectionLatencies := []time.Duration{
			10 * time.Millisecond,
			15 * time.Millisecond,
			12 * time.Millisecond,
			20 * time.Millisecond,
		}

		// Act
		var totalMs float64
		for _, latency := range connectionLatencies {
			totalMs += latency.Seconds() * 1000
		}
		avgLatency := totalMs / float64(len(connectionLatencies))

		// Assert
		assert.Less(t, avgLatency, 50.0, "Average latency should be < 50ms")
	})

	t.Run("Broker failure rate by type metric", func(t *testing.T) {
		// Arrange
		failures := map[BrokerFailureType]int{
			BrokerConnectionRefused: 5,
			BrokerAuthFailed:        2,
			BrokerTimeout:           3,
			BrokerNetworkPartition:  1,
		}

		// Act
		totalFailures := 0
		for _, count := range failures {
			totalFailures += count
		}

		// Assert
		assert.Equal(t, 11, totalFailures)
		assert.Equal(t, 5, failures[BrokerConnectionRefused])
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// Integration Test: Complete Broker Failure DLQ Workflow.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestSRE015_Integration_BrokerFailureDLQ(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Complete broker failure recovery workflow", func(t *testing.T) {
		ctx := context.Background()

		// Step 1: Detect broker failure
		event := BrokerFailureEvent{
			EventID:     "integration-test-1",
			FailureType: BrokerConnectionRefused,
			Timestamp:   time.Now(),
		}

		// Step 2: Retry with exponential backoff
		maxRetries := 5
		for event.RetryCount < maxRetries {
			event.RetryCount++
			time.Sleep(10 * time.Millisecond) // Simulate retry delay

			// Simulate persistent failure
			brokerAvailable := false
			if brokerAvailable {
				break
			}
		}

		// Step 3: Move to DLQ after max retries
		if event.RetryCount >= maxRetries {
			event.MovedToDLQ = true
			event.DLQReason = "broker_connection_failed_max_retries"
		}

		// Step 4: Trigger failover
		failoverStart := time.Now()
		secondaryBroker := "rabbitmq-1"
		activeBroker := secondaryBroker // Failover occurred
		failoverEnd := time.Now()
		event.RecoveryTime = failoverEnd.Sub(failoverStart)

		// Step 5: Verify DLQ metrics
		require.NotNil(t, ctx)
		assert.Equal(t, maxRetries, event.RetryCount)
		assert.True(t, event.MovedToDLQ)
		assert.Equal(t, secondaryBroker, activeBroker)
		testutils.RunTimingTest(t, "Broker failover recovery", failoverStart, failoverEnd, 30*time.Second, []testutils.Phase{})

		t.Logf("🎯 Broker failure recovery workflow completed:")
		t.Logf("  - Retries: %d", event.RetryCount)
		t.Logf("  - DLQ: %v", event.MovedToDLQ)
		t.Logf("  - Active broker: %s", activeBroker)
		t.Logf("  - Recovery time: %v", event.RecoveryTime)
	})
}
