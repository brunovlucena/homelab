// Package testutils provides common test utilities and helpers for the knative-lambda project.
package testutils

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// SetupTestEnvironment prepares common test environment variables.
func SetupTestEnvironment(t *testing.T) {
	t.Helper()

	// Set default test environment variables
	envVars := map[string]string{
		"ENV":                    "test",
		"AWS_REGION":             "us-west-2",
		"NAMESPACE":              "knative-lambda",
		"LOG_LEVEL":              "debug",
		"REDIS_ADDR":             "localhost:6379",
		"PROMETHEUS_PUSHGATEWAY": "localhost:9091",
	}

	for key, value := range envVars {
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				// In test context, we can't return error, so just log it
				// Use logrus for proper structured logging
				logrus.WithError(err).WithField("env_var", key).Warn("Failed to set environment variable")
			}
		}
	}
}

// CleanupTestEnvironment cleans up test resources.
func CleanupTestEnvironment(t *testing.T) {
	t.Helper()
	// Add cleanup logic here if needed
}

// GetTestContext returns a context with timeout for tests.
func GetTestContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return context.WithTimeout(context.Background(), timeout)
}

// SkipIfShort skips a test if running in short mode.
func SkipIfShort(t *testing.T, reason string) {
	t.Helper()
	if testing.Short() {
		t.Skipf("Skipping test in short mode: %s", reason)
	}
}

// RequireEnv fails the test if an environment variable is not set.
func RequireEnv(t *testing.T, key string) string {
	t.Helper()
	value := os.Getenv(key)
	if value == "" {
		t.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

// MockCloudEventData returns a sample CloudEvent JSON payload for testing.
func MockCloudEventData(eventType string) []byte {
	switch eventType {
	case "build":
		return []byte(`{
			"specversion": "1.0",
			"type": "dev.knative.lambda.build",
			"source": "test",
			"id": "test-build-1",
			"datacontenttype": "application/json",
			"data": {
				"parserId": "test-parser-123",
				"environment": "test",
				"imageRegistry": "localhost:5001"
			}
		}`)
	case "service":
		return []byte(`{
			"specversion": "1.0",
			"type": "dev.knative.lambda.service",
			"source": "test",
			"id": "test-service-1",
			"datacontenttype": "application/json",
			"data": {
				"serviceName": "test-service",
				"environment": "test",
				"image": "localhost:5001/test:latest"
			}
		}`)
	default:
		return []byte(`{
			"specversion": "1.0",
			"type": "test.event",
			"source": "test",
			"id": "test-1",
			"datacontenttype": "application/json",
			"data": {}
		}`)
	}
}
