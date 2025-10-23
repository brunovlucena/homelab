// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	⏱️ TIMEOUT HANDLING - Context timeout management for storage operations
//
//	🎯 Purpose: Ensure storage operations respect context deadlines
//	💡 Features: Operation-specific timeouts, context validation
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package storage

import (
	"context"
	"fmt"
	"time"
)

// ⏱️ Operation timeouts - "Default timeouts for storage operations"
const (
	DefaultUploadTimeout   = 5 * time.Minute  // Upload timeout
	DefaultDownloadTimeout = 5 * time.Minute  // Download timeout
	DefaultDeleteTimeout   = 30 * time.Second // Delete timeout
	DefaultExistsTimeout   = 10 * time.Second // Exists check timeout
)

// ⏱️ OperationTimeouts - "Configurable timeouts for storage operations"
type OperationTimeouts struct {
	Upload   time.Duration
	Download time.Duration
	Delete   time.Duration
	Exists   time.Duration
}

// 🔧 DefaultOperationTimeouts - "Get default operation timeouts"
func DefaultOperationTimeouts() OperationTimeouts {
	return OperationTimeouts{
		Upload:   DefaultUploadTimeout,
		Download: DefaultDownloadTimeout,
		Delete:   DefaultDeleteTimeout,
		Exists:   DefaultExistsTimeout,
	}
}

// ⏱️ WithUploadTimeout - "Add timeout to context for upload operation"
//
// This function wraps the context with a timeout specific to upload operations.
// If the context already has a deadline, it uses the shorter of the two.
//
// Parameters:
//   - ctx: parent context
//   - timeout: upload timeout duration (0 = use default)
//
// Returns:
//   - context.Context: context with timeout
//   - context.CancelFunc: cancel function to clean up resources
func WithUploadTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		timeout = DefaultUploadTimeout
	}
	return withOperationTimeout(ctx, timeout, "upload")
}

// ⏱️ WithDownloadTimeout - "Add timeout to context for download operation"
func WithDownloadTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		timeout = DefaultDownloadTimeout
	}
	return withOperationTimeout(ctx, timeout, "download")
}

// ⏱️ WithDeleteTimeout - "Add timeout to context for delete operation"
func WithDeleteTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		timeout = DefaultDeleteTimeout
	}
	return withOperationTimeout(ctx, timeout, "delete")
}

// ⏱️ WithExistsTimeout - "Add timeout to context for exists check operation"
func WithExistsTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		timeout = DefaultExistsTimeout
	}
	return withOperationTimeout(ctx, timeout, "exists")
}

// 🔧 withOperationTimeout - "Internal helper to add timeout to context"
func withOperationTimeout(ctx context.Context, timeout time.Duration, operation string) (context.Context, context.CancelFunc) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		// Return context with immediate cancel
		newCtx, cancel := context.WithCancel(context.Background())
		cancel()
		return newCtx, cancel
	}

	// Check if context already has a deadline
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		// Calculate remaining time
		remaining := time.Until(deadline)

		// Use shorter of existing deadline or new timeout
		if remaining < timeout {
			// Existing deadline is shorter, don't add new timeout
			return ctx, func() {} // No-op cancel
		}
	}

	// Add timeout to context
	return context.WithTimeout(ctx, timeout)
}

// ⏱️ ValidateContext - "Validate context is not cancelled or expired"
//
// This function checks if a context is valid for starting a new operation.
// It returns an error if the context is already cancelled or expired.
func ValidateContext(ctx context.Context, operation string) error {
	if ctx == nil {
		return fmt.Errorf("context is nil for %s operation", operation)
	}

	if err := ctx.Err(); err != nil {
		if err == context.Canceled {
			return fmt.Errorf("%s operation cancelled before start: %w", operation, err)
		}
		if err == context.DeadlineExceeded {
			return fmt.Errorf("%s operation deadline exceeded before start: %w", operation, err)
		}
		return fmt.Errorf("%s operation context error: %w", operation, err)
	}

	return nil
}

// ⏱️ RemainingTime - "Get remaining time before context deadline"
//
// Returns the duration until the context deadline, or 0 if no deadline.
func RemainingTime(ctx context.Context) time.Duration {
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		remaining := time.Until(deadline)
		if remaining < 0 {
			return 0
		}
		return remaining
	}
	return 0 // No deadline
}

// ⏱️ HasSufficientTime - "Check if context has sufficient time remaining"
//
// Returns true if the context has at least the specified duration remaining
// before its deadline, or if it has no deadline.
func HasSufficientTime(ctx context.Context, required time.Duration) bool {
	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		return true // No deadline, sufficient time
	}

	remaining := time.Until(deadline)
	return remaining >= required
}

// ⏱️ EstimateOperationTime - "Estimate time required for operation based on size"
//
// Provides a rough estimate of operation time based on file size.
// This can be used to check if there's sufficient time in the context.
//
// Parameters:
//   - operation: type of operation ("upload", "download", etc.)
//   - sizeBytes: size of data in bytes
//
// Returns:
//   - time.Duration: estimated time for operation
func EstimateOperationTime(operation string, sizeBytes int64) time.Duration {
	// Rough estimates assuming network speed of 10 MB/s
	const bytesPerSecond = 10 * 1024 * 1024

	baseTime := 1 * time.Second // Base overhead for API call
	transferTime := time.Duration(float64(sizeBytes)/float64(bytesPerSecond)) * time.Second

	switch operation {
	case "upload", "download":
		return baseTime + transferTime
	case "delete", "exists":
		return baseTime
	default:
		return baseTime + transferTime
	}
}
