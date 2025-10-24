// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🔄 RETRY LOGIC - Retry logic for transient storage failures
//
//	🎯 Purpose: Handle transient failures in storage operations
//	💡 Features: Exponential backoff, configurable retries, error classification
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package storage

import (
	"context"
	"time"

	"knative-lambda-new/internal/resilience"
)

// 🔧 RetryConfig - "Configuration for retry logic"
type RetryConfig struct {
	MaxAttempts  int           // Maximum number of retry attempts
	InitialDelay time.Duration // Initial delay between retries
	MaxDelay     time.Duration // Maximum delay between retries
	Multiplier   float64       // Backoff multiplier
}

// 🔧 DefaultRetryConfig - "Default retry configuration for storage operations"
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// 🔄 WithRetry - "Execute operation with retry logic for transient failures"
//
// This function wraps storage operations with retry logic using exponential backoff.
// It automatically retries on transient errors like network issues, timeouts, and
// temporary service unavailability.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - config: retry configuration (use DefaultRetryConfig() for defaults)
//   - operation: the operation to retry
//
// Returns:
//   - error: nil if operation succeeded, error if all retries failed
func WithRetry(ctx context.Context, config RetryConfig, operation func() error) error {
	retryPolicy := resilience.RetryPolicy{
		MaxAttempts:       config.MaxAttempts,
		InitialDelay:      config.InitialDelay,
		MaxDelay:          config.MaxDelay,
		BackoffMultiplier: config.Multiplier,
		ShouldRetry: func(err error) bool {
			// Retry on transient errors
			return resilience.IsTransientError(err)
		},
	}

	return resilience.ExecuteWithRetry(ctx, retryPolicy, operation)
}

// 🔧 IsRetryable - "Check if an error is retryable"
//
// Determines if an error should trigger a retry based on error type and code.
// This includes network errors, timeouts, and temporary service unavailability.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a transient error using resilience package
	return resilience.IsTransientError(err)
}
