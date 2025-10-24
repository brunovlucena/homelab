// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 TESTING HELPERS - Shared test utilities for storage package
//
//	🎯 Purpose: Common test helpers and utilities
//	💡 Features: Observability setup for tests
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package storage

import (
	"testing"

	"knative-lambda-new/internal/observability"
	testhelpers "knative-lambda-new/internal/testing"
)

// createTestObservability creates a minimal observability instance for testing
// Delegates to shared test helper
func createTestObservability(t *testing.T) *observability.Observability {
	t.Helper()
	return testhelpers.CreateTestObservability(t)
}
