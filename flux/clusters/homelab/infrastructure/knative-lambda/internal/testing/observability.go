// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 TESTING HELPERS - Shared test utilities for all packages
//
//	🎯 Purpose: Common test helpers and utilities
//	💡 Features: Observability setup, test fixtures, mock objects
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package testing

import (
	"testing"

	"knative-lambda-new/internal/config"
	"knative-lambda-new/internal/observability"
)

// 🧪 CreateTestObservability - "Create a properly initialized observability instance for testing"
//
// This helper creates an observability instance with minimal but complete configuration
// suitable for testing. It ensures all internal components (tracer, metrics, etc.) are
// properly initialized to avoid nil pointer dereferences.
//
// Parameters:
//   - t: testing.T instance for error reporting
//
// Returns:
//   - *observability.Observability: fully initialized observability instance
func CreateTestObservability(t *testing.T) *observability.Observability {
	t.Helper()

	obsConfig := observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "error", // Minimal logging to reduce test noise
		MetricsEnabled: true,    // Enable for proper initialization
		TracingEnabled: true,    // Enable for proper initialization
		OTLPEndpoint:   "",      // No actual endpoint needed for tests
		SampleRate:     1.0,     // Sample all traces in tests
	}

	obs, err := observability.New(obsConfig)
	if err != nil {
		t.Fatalf("Failed to create test observability: %v", err)
	}

	return obs
}

// 🧪 CreateTestObservabilityWithConfig - "Create observability with custom config for advanced testing"
//
// This helper allows tests to customize the observability configuration while still
// ensuring proper initialization.
//
// Parameters:
//   - t: testing.T instance for error reporting
//   - customize: function to customize the default config
//
// Returns:
//   - *observability.Observability: fully initialized observability instance
func CreateTestObservabilityWithConfig(t *testing.T, customize func(*observability.Config)) *observability.Observability {
	t.Helper()

	obsConfig := observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "error",
		MetricsEnabled: true,
		TracingEnabled: true,
		OTLPEndpoint:   "",
		SampleRate:     1.0,
	}

	// Allow test to customize config
	if customize != nil {
		customize(&obsConfig)
	}

	obs, err := observability.New(obsConfig)
	if err != nil {
		t.Fatalf("Failed to create test observability: %v", err)
	}

	return obs
}

// 🧪 CreateTestConfig - "Create a properly initialized config instance for testing"
//
// This helper creates a configuration instance with minimal but valid configuration
// suitable for testing. It ensures all internal components are properly initialized
// with test-safe defaults.
//
// Parameters:
//   - t: testing.T instance for error reporting
//
// Returns:
//   - *config.Config: fully initialized config instance
func CreateTestConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg, err := config.NewConfigBuilder().
		WithEnvironment("test").
		Build()

	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	return cfg
}

// 🧪 CreateTestConfigWithCustomizer - "Create config with custom modifications for advanced testing"
//
// This helper allows tests to customize the configuration while still
// ensuring proper initialization.
//
// Parameters:
//   - t: testing.T instance for error reporting
//   - customize: function to customize the builder before build
//
// Returns:
//   - *config.Config: fully initialized config instance
func CreateTestConfigWithCustomizer(t *testing.T, customize func(*config.ConfigBuilder) *config.ConfigBuilder) *config.Config {
	t.Helper()

	builder := config.NewConfigBuilder().
		WithEnvironment("test")

	// Allow test to customize builder
	if customize != nil {
		builder = customize(builder)
	}

	cfg, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	return cfg
}
