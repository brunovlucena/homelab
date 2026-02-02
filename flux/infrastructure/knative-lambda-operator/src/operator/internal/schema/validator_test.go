// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 SCHEMA VALIDATOR TESTS
//
//	Comprehensive tests for CloudEvent JSON Schema validation
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 VALIDATOR INITIALIZATION TESTS                                      │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNewValidator(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)
	require.NotNil(t, v)

	// Verify schemas are registered
	types := v.RegisteredEventTypes()
	assert.NotEmpty(t, types, "Should have registered event types")
	assert.Contains(t, types, "io.knative.lambda.command.function.deploy")
	assert.Contains(t, types, "io.knative.lambda.command.service.delete")
	assert.Contains(t, types, "io.knative.lambda.command.build.start")
	assert.Contains(t, types, "io.knative.lambda.invoke.sync")
}

func TestHasSchema(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		eventType string
		expected  bool
	}{
		{"io.knative.lambda.command.function.deploy", true},
		{"io.knative.lambda.command.service.delete", true},
		{"io.knative.lambda.invoke.async", true},
		{"io.knative.lambda.unknown.event", false},
		{"some.random.event", false},
	}

	for _, tc := range tests {
		t.Run(tc.eventType, func(t *testing.T) {
			assert.Equal(t, tc.expected, v.HasSchema(tc.eventType))
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🚀 FUNCTION DEPLOY VALIDATION TESTS                                    │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateFunctionDeploy_Valid(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name string
		data map[string]interface{}
	}{
		{
			name: "MinIO source",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      "test-lambda",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "minio",
						"minio": map[string]interface{}{
							"bucket": "lambda-functions",
							"key":    "hello-python/",
						},
					},
					"runtime": map[string]interface{}{
						"language": "python",
						"version":  "3.11",
						"handler":  "main.handler",
					},
				},
			},
		},
		{
			name: "Git source",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "git-lambda",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "git",
						"git": map[string]interface{}{
							"url": "https://github.com/example/repo.git",
							"ref": "main",
						},
					},
					"runtime": map[string]interface{}{
						"language": "nodejs",
						"version":  "20",
					},
				},
			},
		},
		{
			name: "Inline source",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "inline-lambda",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "inline",
						"inline": map[string]interface{}{
							"code": "def handler(event): return {'status': 'ok'}",
						},
					},
					"runtime": map[string]interface{}{
						"language": "python",
						"version":  "3.11",
					},
				},
			},
		},
		{
			name: "Image source",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "image-lambda",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "image",
						"image": map[string]interface{}{
							"repository": "localhost:5001/my-app",
							"tag":        "v1.0.0",
						},
					},
					"runtime": map[string]interface{}{
						"language": "python",
						"version":  "3.11",
					},
				},
			},
		},
		{
			name: "With scaling and resources",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "full-lambda",
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "inline",
						"inline": map[string]interface{}{
							"code": "handler code",
						},
					},
					"runtime": map[string]interface{}{
						"language": "go",
						"version":  "1.21",
					},
					"scaling": map[string]interface{}{
						"minReplicas": 0,
						"maxReplicas": 10,
					},
					"resources": map[string]interface{}{
						"limits": map[string]interface{}{
							"memory": "256Mi",
							"cpu":    "100m",
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate("io.knative.lambda.command.function.deploy", tc.data)
			assert.NoError(t, err, "Valid payload should pass validation")
		})
	}
}

func TestValidateFunctionDeploy_Invalid(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name          string
		data          map[string]interface{}
		expectedError string
	}{
		{
			name: "Missing metadata",
			data: map[string]interface{}{
				"spec": map[string]interface{}{
					"source":  map[string]interface{}{"type": "inline"},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			},
			expectedError: "metadata",
		},
		{
			name: "Missing metadata.name",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"source":  map[string]interface{}{"type": "inline", "inline": map[string]interface{}{"code": "x"}},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			},
			expectedError: "name",
		},
		{
			name: "Invalid source type",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "test"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "invalid-source-type",
					},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			},
			expectedError: "type",
		},
		{
			name: "Invalid runtime language",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "test"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type":   "inline",
						"inline": map[string]interface{}{"code": "x"},
					},
					"runtime": map[string]interface{}{
						"language": "rust", // Not supported
						"version":  "1.70",
					},
				},
			},
			expectedError: "language",
		},
		{
			name: "MinIO source without bucket",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "test"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "minio",
						"minio": map[string]interface{}{
							"key": "some-key",
							// Missing bucket
						},
					},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			},
			expectedError: "bucket",
		},
		{
			name: "Git source without URL",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "test"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "git",
						"git": map[string]interface{}{
							"ref": "main",
							// Missing url
						},
					},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			},
			expectedError: "url",
		},
		{
			name: "Invalid name format (uppercase)",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "INVALID-NAME", // Must be lowercase RFC 1123
				},
				"spec": map[string]interface{}{
					"source":  map[string]interface{}{"type": "inline", "inline": map[string]interface{}{"code": "x"}},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			},
			expectedError: "name",
		},
		{
			name: "Type mismatch - source config doesn't match type",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "test"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "minio", // Says minio but provides git config
						"git": map[string]interface{}{
							"url": "https://github.com/test/repo",
						},
					},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			},
			expectedError: "minio",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate("io.knative.lambda.command.function.deploy", tc.data)
			require.Error(t, err, "Invalid payload should fail validation")
			assert.Contains(t, err.Error(), tc.expectedError,
				"Error should mention: %s", tc.expectedError)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔨 BUILD COMMAND VALIDATION TESTS                                      │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateBuildCommand(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	eventTypes := []string{
		"io.knative.lambda.command.build.start",
		"io.knative.lambda.command.build.cancel",
		"io.knative.lambda.command.build.retry",
	}

	for _, eventType := range eventTypes {
		t.Run(eventType, func(t *testing.T) {
			// Valid
			err := v.Validate(eventType, map[string]interface{}{
				"name":      "test-lambda",
				"namespace": "default",
			})
			assert.NoError(t, err)

			// Invalid - missing name
			err = v.Validate(eventType, map[string]interface{}{
				"namespace": "default",
			})
			assert.Error(t, err)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🚀 INVOKE VALIDATION TESTS                                             │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateInvoke(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	eventTypes := []string{
		"io.knative.lambda.invoke.sync",
		"io.knative.lambda.invoke.async",
		"io.knative.lambda.invoke.scheduled",
	}

	for _, eventType := range eventTypes {
		t.Run(eventType, func(t *testing.T) {
			// Valid with payload
			err := v.Validate(eventType, map[string]interface{}{
				"payload": map[string]interface{}{
					"key": "value",
				},
				"correlationId": "abc123",
			})
			assert.NoError(t, err)

			// Valid empty (invoke can have any or no payload)
			err = v.Validate(eventType, map[string]interface{}{})
			assert.NoError(t, err)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🗑️ SERVICE DELETE VALIDATION TESTS                                     │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateServiceDelete(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	// Valid
	err = v.Validate("io.knative.lambda.command.service.delete", map[string]interface{}{
		"name":      "test-lambda",
		"namespace": "default",
	})
	assert.NoError(t, err)

	// Valid - minimal (name can come from Ce-Subject header)
	err = v.Validate("io.knative.lambda.command.service.delete", map[string]interface{}{})
	assert.NoError(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📨 VALIDATION ERROR TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidationError(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	err = v.Validate("io.knative.lambda.command.function.deploy", map[string]interface{}{
		"invalid": "data",
	})

	require.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	require.True(t, ok, "Should return ValidationError")

	assert.Equal(t, "io.knative.lambda.command.function.deploy", validationErr.EventType)
	assert.NotEmpty(t, validationErr.Errors)
	assert.NotEmpty(t, validationErr.Error())
}

func TestValidateJSON(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	// Valid JSON
	validJSON := []byte(`{
		"metadata": {"name": "test-lambda"},
		"spec": {
			"source": {"type": "inline", "inline": {"code": "handler code"}},
			"runtime": {"language": "python", "version": "3.11"}
		}
	}`)

	err = v.ValidateJSON("io.knative.lambda.command.function.deploy", validJSON)
	assert.NoError(t, err)

	// Invalid JSON syntax
	err = v.ValidateJSON("io.knative.lambda.command.function.deploy", []byte(`{invalid json`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")

	// Valid JSON but invalid schema
	err = v.ValidateJSON("io.knative.lambda.command.function.deploy", []byte(`{"invalid": "schema"}`))
	assert.Error(t, err)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔍 UNKNOWN EVENT TYPE TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateUnknownEventType(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	err = v.Validate("io.knative.lambda.unknown.event.type", map[string]interface{}{
		"any": "data",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no schema registered")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📋 SCHEMA VERSION TEST                                                 │
// └─────────────────────────────────────────────────────────────────────────┘

func TestSchemaVersion(t *testing.T) {
	assert.NotEmpty(t, SchemaVersion)
	assert.Equal(t, "1.0.0", SchemaVersion)
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 LIFECYCLE BUILD VALIDATION TESTS                                    │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateLifecycleBuild_Valid(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name      string
		eventType string
		data      map[string]interface{}
	}{
		{
			name:      "Build started minimal",
			eventType: "io.knative.lambda.lifecycle.build.started",
			data: map[string]interface{}{
				"name":      "test-lambda",
				"namespace": "default",
			},
		},
		{
			name:      "Build completed with image",
			eventType: "io.knative.lambda.lifecycle.build.completed",
			data: map[string]interface{}{
				"name":        "test-lambda",
				"namespace":   "default",
				"jobName":     "test-lambda-build-abc123",
				"imageURI":    "localhost:5001/default/test-lambda:v1",
				"startedAt":   "2024-01-01T12:00:00Z",
				"completedAt": "2024-01-01T12:05:00Z",
				"duration":    "5m0s",
			},
		},
		{
			name:      "Build failed with error",
			eventType: "io.knative.lambda.lifecycle.build.failed",
			data: map[string]interface{}{
				"name":      "test-lambda",
				"namespace": "production",
				"jobName":   "test-lambda-build-xyz789",
				"error":     "Failed to pull base image: timeout",
				"attempt":   3,
			},
		},
		{
			name:      "Build timeout event",
			eventType: "io.knative.lambda.lifecycle.build.timeout",
			data: map[string]interface{}{
				"name":      "slow-function",
				"namespace": "default",
				"duration":  "30m0s",
				"attempt":   1,
			},
		},
		{
			name:      "Build cancelled event",
			eventType: "io.knative.lambda.lifecycle.build.cancelled",
			data: map[string]interface{}{
				"name":      "cancelled-func",
				"namespace": "test",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate(tc.eventType, tc.data)
			assert.NoError(t, err, "Valid lifecycle build payload should pass")
		})
	}
}

func TestValidateLifecycleBuild_Invalid(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name          string
		data          map[string]interface{}
		expectedError string
	}{
		{
			name: "Missing name",
			data: map[string]interface{}{
				"namespace": "default",
			},
			expectedError: "name",
		},
		{
			name: "Missing namespace",
			data: map[string]interface{}{
				"name": "test-lambda",
			},
			expectedError: "namespace",
		},
		{
			name:          "Empty object",
			data:          map[string]interface{}{},
			expectedError: "missing properties",
		},
		{
			name: "Invalid attempt number",
			data: map[string]interface{}{
				"name":      "test",
				"namespace": "default",
				"attempt":   0, // Must be >= 1
			},
			expectedError: "attempt",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate("io.knative.lambda.lifecycle.build.started", tc.data)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📨 RESPONSE VALIDATION TESTS                                           │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateResponse_Valid(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name      string
		eventType string
		data      map[string]interface{}
	}{
		{
			name:      "Success response",
			eventType: "io.knative.lambda.response.success",
			data: map[string]interface{}{
				"statusCode": 200,
				"body":       map[string]interface{}{"message": "Hello"},
				"duration":   "150ms",
			},
		},
		{
			name:      "Success with headers",
			eventType: "io.knative.lambda.response.success",
			data: map[string]interface{}{
				"statusCode": 201,
				"body":       "created",
				"headers": map[string]interface{}{
					"Content-Type": "application/json",
					"X-Request-ID": "abc123",
				},
			},
		},
		{
			name:      "Error response",
			eventType: "io.knative.lambda.response.error",
			data: map[string]interface{}{
				"statusCode": 500,
				"error":      "Internal server error",
				"errorType":  "RuntimeError",
				"stackTrace": []interface{}{
					"at handler (/app/main.py:10)",
					"at process (/app/utils.py:25)",
				},
			},
		},
		{
			name:      "Timeout response",
			eventType: "io.knative.lambda.response.timeout",
			data: map[string]interface{}{
				"statusCode":    504,
				"error":         "Function execution timed out",
				"errorType":     "TimeoutError",
				"duration":      "30s",
				"correlationId": "corr-xyz-123",
			},
		},
		{
			name:      "Empty response (valid)",
			eventType: "io.knative.lambda.response.success",
			data:      map[string]interface{}{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate(tc.eventType, tc.data)
			assert.NoError(t, err)
		})
	}
}

func TestValidateResponse_Invalid(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name          string
		data          map[string]interface{}
		expectedError string
	}{
		{
			name: "Invalid status code - too low",
			data: map[string]interface{}{
				"statusCode": 50, // Must be >= 100
			},
			expectedError: "statusCode",
		},
		{
			name: "Invalid status code - too high",
			data: map[string]interface{}{
				"statusCode": 600, // Must be <= 599
			},
			expectedError: "statusCode",
		},
		{
			name: "Invalid headers type",
			data: map[string]interface{}{
				"headers": "should-be-object",
			},
			expectedError: "headers",
		},
		{
			name: "Invalid stackTrace type",
			data: map[string]interface{}{
				"stackTrace": "should-be-array",
			},
			expectedError: "stackTrace",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate("io.knative.lambda.response.success", tc.data)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ☁️ S3/GCS SOURCE VALIDATION TESTS                                      │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateS3Source(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name        string
		data        map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid S3 source",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "s3-lambda"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "s3",
						"s3": map[string]interface{}{
							"bucket": "my-bucket",
							"key":    "functions/hello.zip",
							"region": "us-east-1",
						},
					},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			},
			expectError: false,
		},
		{
			name: "Valid S3 source with secretRef",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "s3-secret-lambda"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "s3",
						"s3": map[string]interface{}{
							"bucket":    "private-bucket",
							"key":       "functions/secure.zip",
							"region":    "eu-west-1",
							"secretRef": map[string]interface{}{"name": "aws-credentials"},
						},
					},
					"runtime": map[string]interface{}{"language": "nodejs", "version": "20"},
				},
			},
			expectError: false,
		},
		{
			name: "S3 source missing bucket",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "s3-invalid"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "s3",
						"s3": map[string]interface{}{
							"key": "functions/hello.zip",
						},
					},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			},
			expectError: true,
		},
		{
			name: "S3 source missing key",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "s3-invalid"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "s3",
						"s3": map[string]interface{}{
							"bucket": "my-bucket",
						},
					},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate("io.knative.lambda.command.function.deploy", tc.data)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateGCSSource(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name        string
		data        map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid GCS source",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "gcs-lambda"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "gcs",
						"gcs": map[string]interface{}{
							"bucket":  "gcs-bucket",
							"key":     "functions/handler.zip",
							"project": "my-gcp-project",
						},
					},
					"runtime": map[string]interface{}{"language": "go", "version": "1.21"},
				},
			},
			expectError: false,
		},
		{
			name: "Valid GCS source with secretRef",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "gcs-secret-lambda"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "gcs",
						"gcs": map[string]interface{}{
							"bucket":    "private-gcs-bucket",
							"key":       "functions/secure.zip",
							"secretRef": map[string]interface{}{"name": "gcp-credentials"},
						},
					},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			},
			expectError: false,
		},
		{
			name: "GCS source missing bucket",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "gcs-invalid"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "gcs",
						"gcs": map[string]interface{}{
							"key": "functions/hello.zip",
						},
					},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate("io.knative.lambda.command.function.deploy", tc.data)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔧 BUILD SPEC VALIDATION TESTS                                         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateBuildSpec(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name        string
		buildSpec   map[string]interface{}
		expectError bool
	}{
		{
			name: "Full build spec",
			buildSpec: map[string]interface{}{
				"timeout":      "10m",
				"registry":     "gcr.io",
				"registryType": "gcr",
				"repository":   "my-project/functions",
				"tag":          "v1.0.0",
				"insecure":     false,
				"forceRebuild": true,
			},
			expectError: false,
		},
		{
			name: "ECR registry type",
			buildSpec: map[string]interface{}{
				"registryType": "ecr",
				"registry":     "123456789.dkr.ecr.us-east-1.amazonaws.com",
			},
			expectError: false,
		},
		{
			name: "GHCR registry type",
			buildSpec: map[string]interface{}{
				"registryType": "ghcr",
				"registry":     "ghcr.io",
			},
			expectError: false,
		},
		{
			name: "Local insecure registry",
			buildSpec: map[string]interface{}{
				"registryType": "local",
				"registry":     "localhost:5001",
				"insecure":     true,
			},
			expectError: false,
		},
		{
			name: "Invalid registry type",
			buildSpec: map[string]interface{}{
				"registryType": "invalid-registry",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]interface{}{
				"metadata": map[string]interface{}{"name": "build-test"},
				"spec": map[string]interface{}{
					"source":  map[string]interface{}{"type": "inline", "inline": map[string]interface{}{"code": "x"}},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
					"build":   tc.buildSpec,
				},
			}
			err := v.Validate("io.knative.lambda.command.function.deploy", data)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔑 ENV VAR VALIDATION TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateEnvVars(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name        string
		envVars     []interface{}
		expectError bool
	}{
		{
			name: "Simple env vars",
			envVars: []interface{}{
				map[string]interface{}{"name": "DEBUG", "value": "true"},
				map[string]interface{}{"name": "LOG_LEVEL", "value": "debug"},
			},
			expectError: false,
		},
		{
			name: "Env var with valueFrom",
			envVars: []interface{}{
				map[string]interface{}{
					"name": "API_KEY",
					"valueFrom": map[string]interface{}{
						"secretKeyRef": map[string]interface{}{
							"name": "api-secrets",
							"key":  "api-key",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Mixed env vars",
			envVars: []interface{}{
				map[string]interface{}{"name": "NODE_ENV", "value": "production"},
				map[string]interface{}{
					"name": "DB_PASSWORD",
					"valueFrom": map[string]interface{}{
						"secretKeyRef": map[string]interface{}{
							"name": "db-secrets",
							"key":  "password",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Env var missing name",
			envVars: []interface{}{
				map[string]interface{}{"value": "no-name"},
			},
			expectError: true,
		},
		{
			name: "Empty env var name",
			envVars: []interface{}{
				map[string]interface{}{"name": "", "value": "empty-name"},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]interface{}{
				"metadata": map[string]interface{}{"name": "env-test"},
				"spec": map[string]interface{}{
					"source":  map[string]interface{}{"type": "inline", "inline": map[string]interface{}{"code": "x"}},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
					"env":     tc.envVars,
				},
			}
			err := v.Validate("io.knative.lambda.command.function.deploy", data)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🖼️ IMAGE SOURCE ADVANCED TESTS                                         │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateImageSourceAdvanced(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name        string
		imageSpec   map[string]interface{}
		expectError bool
	}{
		{
			name: "Image with custom port",
			imageSpec: map[string]interface{}{
				"repository": "localhost:5001/my-app",
				"tag":        "v1.0.0",
				"port":       3000,
			},
			expectError: false,
		},
		{
			name: "Image with pullPolicy Always",
			imageSpec: map[string]interface{}{
				"repository": "gcr.io/project/app",
				"pullPolicy": "Always",
			},
			expectError: false,
		},
		{
			name: "Image with pullPolicy IfNotPresent",
			imageSpec: map[string]interface{}{
				"repository": "docker.io/library/nginx",
				"tag":        "latest",
				"pullPolicy": "IfNotPresent",
			},
			expectError: false,
		},
		{
			name: "Image with pullPolicy Never",
			imageSpec: map[string]interface{}{
				"repository": "localhost/local-image",
				"pullPolicy": "Never",
			},
			expectError: false,
		},
		{
			name: "Image with digest",
			imageSpec: map[string]interface{}{
				"repository": "gcr.io/project/app",
				"digest":     "sha256:abc123def456",
			},
			expectError: false,
		},
		{
			name: "Invalid port - too low",
			imageSpec: map[string]interface{}{
				"repository": "localhost:5001/app",
				"port":       0,
			},
			expectError: true,
		},
		{
			name: "Invalid port - too high",
			imageSpec: map[string]interface{}{
				"repository": "localhost:5001/app",
				"port":       70000,
			},
			expectError: true,
		},
		{
			name: "Invalid pullPolicy",
			imageSpec: map[string]interface{}{
				"repository": "localhost:5001/app",
				"pullPolicy": "InvalidPolicy",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]interface{}{
				"metadata": map[string]interface{}{"name": "image-test"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type":  "image",
						"image": tc.imageSpec,
					},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			}
			err := v.Validate("io.knative.lambda.command.function.deploy", data)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📏 RESOURCE VALIDATION TESTS                                           │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateResources(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name        string
		resources   map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid memory and CPU",
			resources: map[string]interface{}{
				"requests": map[string]interface{}{
					"memory": "128Mi",
					"cpu":    "100m",
				},
				"limits": map[string]interface{}{
					"memory": "512Mi",
					"cpu":    "500m",
				},
			},
			expectError: false,
		},
		{
			name: "Memory in Gi",
			resources: map[string]interface{}{
				"limits": map[string]interface{}{
					"memory": "2Gi",
				},
			},
			expectError: false,
		},
		{
			name: "CPU as whole number",
			resources: map[string]interface{}{
				"requests": map[string]interface{}{
					"cpu": "1",
				},
			},
			expectError: false,
		},
		{
			name: "Memory in Ki",
			resources: map[string]interface{}{
				"requests": map[string]interface{}{
					"memory": "65536Ki",
				},
			},
			expectError: false,
		},
		{
			name: "Invalid memory format",
			resources: map[string]interface{}{
				"requests": map[string]interface{}{
					"memory": "invalid",
				},
			},
			expectError: true,
		},
		{
			name: "Invalid CPU format",
			resources: map[string]interface{}{
				"requests": map[string]interface{}{
					"cpu": "invalid-cpu",
				},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]interface{}{
				"metadata": map[string]interface{}{"name": "resource-test"},
				"spec": map[string]interface{}{
					"source":    map[string]interface{}{"type": "inline", "inline": map[string]interface{}{"code": "x"}},
					"runtime":   map[string]interface{}{"language": "python", "version": "3.11"},
					"resources": tc.resources,
				},
			}
			err := v.Validate("io.knative.lambda.command.function.deploy", data)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 SCALING VALIDATION TESTS                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateScaling(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name        string
		scaling     map[string]interface{}
		expectError bool
	}{
		{
			name: "Scale to zero",
			scaling: map[string]interface{}{
				"minReplicas": 0,
				"maxReplicas": 10,
			},
			expectError: false,
		},
		{
			name: "Always on (minReplicas > 0)",
			scaling: map[string]interface{}{
				"minReplicas": 2,
				"maxReplicas": 20,
			},
			expectError: false,
		},
		{
			name: "With target concurrency",
			scaling: map[string]interface{}{
				"minReplicas":       1,
				"maxReplicas":       5,
				"targetConcurrency": 50,
			},
			expectError: false,
		},
		{
			name: "With scale to zero grace period",
			scaling: map[string]interface{}{
				"minReplicas":            0,
				"maxReplicas":            10,
				"scaleToZeroGracePeriod": "30s",
			},
			expectError: false,
		},
		{
			name: "Invalid minReplicas - negative",
			scaling: map[string]interface{}{
				"minReplicas": -1,
			},
			expectError: true,
		},
		{
			name: "Invalid maxReplicas - zero",
			scaling: map[string]interface{}{
				"maxReplicas": 0,
			},
			expectError: true,
		},
		{
			name: "Invalid targetConcurrency - zero",
			scaling: map[string]interface{}{
				"targetConcurrency": 0,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]interface{}{
				"metadata": map[string]interface{}{"name": "scaling-test"},
				"spec": map[string]interface{}{
					"source":  map[string]interface{}{"type": "inline", "inline": map[string]interface{}{"code": "x"}},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
					"scaling": tc.scaling,
				},
			}
			err := v.Validate("io.knative.lambda.command.function.deploy", data)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔗 INVOKE SCHEMA ADVANCED TESTS                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateInvokeAdvanced(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name        string
		data        map[string]interface{}
		expectError bool
	}{
		{
			name: "With timeout in seconds",
			data: map[string]interface{}{
				"payload": map[string]interface{}{"data": "test"},
				"timeout": "30s",
			},
			expectError: false,
		},
		{
			name: "With timeout in minutes",
			data: map[string]interface{}{
				"timeout": "5m",
			},
			expectError: false,
		},
		{
			name: "With timeout in hours",
			data: map[string]interface{}{
				"timeout": "1h",
			},
			expectError: false,
		},
		{
			name: "With contextId and correlationId",
			data: map[string]interface{}{
				"correlationId": "corr-123",
				"contextId":     "ctx-456",
			},
			expectError: false,
		},
		{
			name: "Complex payload",
			data: map[string]interface{}{
				"payload": map[string]interface{}{
					"users": []interface{}{
						map[string]interface{}{"id": 1, "name": "Alice"},
						map[string]interface{}{"id": 2, "name": "Bob"},
					},
					"config": map[string]interface{}{
						"debug":   true,
						"retries": 3,
					},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid timeout format",
			data: map[string]interface{}{
				"timeout": "invalid",
			},
			expectError: true,
		},
		{
			name: "Invalid timeout - no unit",
			data: map[string]interface{}{
				"timeout": "30",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate("io.knative.lambda.invoke.async", tc.data)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🃏 WILDCARD SCHEMA MATCHING TESTS                                      │
// └─────────────────────────────────────────────────────────────────────────┘

func TestWildcardSchemaMatching(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	// The current implementation doesn't use wildcards directly,
	// but we test the findWildcardSchema behavior
	t.Run("No wildcard match for unknown category", func(t *testing.T) {
		hasSchema := v.HasSchema("io.knative.lambda.unknown.category.event")
		assert.False(t, hasSchema)
	})

	t.Run("No wildcard match for short event type", func(t *testing.T) {
		hasSchema := v.HasSchema("io.knative")
		assert.False(t, hasSchema)
	})

	t.Run("No wildcard match for random event", func(t *testing.T) {
		hasSchema := v.HasSchema("com.example.custom.event")
		assert.False(t, hasSchema)
	})
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🧵 CONCURRENCY TESTS                                                   │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidatorConcurrency(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	validData := map[string]interface{}{
		"metadata": map[string]interface{}{"name": "concurrent-test"},
		"spec": map[string]interface{}{
			"source":  map[string]interface{}{"type": "inline", "inline": map[string]interface{}{"code": "x"}},
			"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
		},
	}

	// Run multiple concurrent validations
	const numGoroutines = 50
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			err := v.Validate("io.knative.lambda.command.function.deploy", validData)
			assert.NoError(t, err)

			hasSchema := v.HasSchema("io.knative.lambda.command.function.deploy")
			assert.True(t, hasSchema)

			types := v.RegisteredEventTypes()
			assert.NotEmpty(t, types)

			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📛 NAME VALIDATION EDGE CASES                                          │
// └─────────────────────────────────────────────────────────────────────────┘

func TestNameValidationEdgeCases(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name        string
		funcName    string
		expectError bool
	}{
		{
			name:        "Single character name",
			funcName:    "a",
			expectError: false,
		},
		{
			name:        "Maximum length name (63 chars)",
			funcName:    "abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefgh",
			expectError: false,
		},
		{
			name:        "Name starting with number",
			funcName:    "1-my-function",
			expectError: false,
		},
		{
			name:        "Name with all numbers",
			funcName:    "123456789",
			expectError: false,
		},
		{
			name:        "Name starting with hyphen",
			funcName:    "-invalid",
			expectError: true,
		},
		{
			name:        "Name ending with hyphen",
			funcName:    "invalid-",
			expectError: true,
		},
		{
			name:        "Name with underscore",
			funcName:    "invalid_name",
			expectError: true,
		},
		{
			name:        "Name with uppercase",
			funcName:    "InvalidName",
			expectError: true,
		},
		{
			name:        "Name with spaces",
			funcName:    "invalid name",
			expectError: true,
		},
		{
			name:        "Name exceeding 63 chars",
			funcName:    "abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]interface{}{
				"metadata": map[string]interface{}{"name": tc.funcName},
				"spec": map[string]interface{}{
					"source":  map[string]interface{}{"type": "inline", "inline": map[string]interface{}{"code": "x"}},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			}
			err := v.Validate("io.knative.lambda.command.function.deploy", data)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 GIT SOURCE ADVANCED TESTS                                           │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateGitSourceAdvanced(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name        string
		gitSpec     map[string]interface{}
		expectError bool
	}{
		{
			name: "HTTPS URL",
			gitSpec: map[string]interface{}{
				"url": "https://github.com/example/repo.git",
				"ref": "main",
			},
			expectError: false,
		},
		{
			name: "HTTP URL",
			gitSpec: map[string]interface{}{
				"url": "http://internal-git.company.com/repo.git",
			},
			expectError: false,
		},
		{
			name: "SSH URL with git@",
			gitSpec: map[string]interface{}{
				"url": "git@github.com:example/repo.git",
				"ref": "v1.0.0",
			},
			expectError: false,
		},
		{
			name: "SSH URL with ssh://",
			gitSpec: map[string]interface{}{
				"url": "ssh://git@github.com/example/repo.git",
			},
			expectError: false,
		},
		{
			name: "With path subdirectory",
			gitSpec: map[string]interface{}{
				"url":  "https://github.com/example/monorepo.git",
				"ref":  "main",
				"path": "functions/hello-world",
			},
			expectError: false,
		},
		{
			name: "With secretRef for private repo",
			gitSpec: map[string]interface{}{
				"url":       "https://github.com/private/repo.git",
				"secretRef": map[string]interface{}{"name": "git-credentials"},
			},
			expectError: false,
		},
		{
			name: "Invalid URL - missing protocol",
			gitSpec: map[string]interface{}{
				"url": "github.com/example/repo.git",
			},
			expectError: true,
		},
		{
			name: "Invalid URL - ftp protocol",
			gitSpec: map[string]interface{}{
				"url": "ftp://example.com/repo.git",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]interface{}{
				"metadata": map[string]interface{}{"name": "git-test"},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"type": "git",
						"git":  tc.gitSpec,
					},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			}
			err := v.Validate("io.knative.lambda.command.function.deploy", data)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📜 REGISTERED EVENT TYPES TESTS                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestRegisteredEventTypes(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	types := v.RegisteredEventTypes()

	// Verify all expected event types are registered
	expectedTypes := []string{
		// Command events
		"io.knative.lambda.command.function.deploy",
		"io.knative.lambda.command.service.create",
		"io.knative.lambda.command.service.update",
		"io.knative.lambda.command.service.delete",
		"io.knative.lambda.command.build.start",
		"io.knative.lambda.command.build.cancel",
		"io.knative.lambda.command.build.retry",

		// Invoke events
		"io.knative.lambda.invoke.sync",
		"io.knative.lambda.invoke.async",
		"io.knative.lambda.invoke.scheduled",

		// Lifecycle events
		"io.knative.lambda.lifecycle.build.started",
		"io.knative.lambda.lifecycle.build.completed",
		"io.knative.lambda.lifecycle.build.failed",
		"io.knative.lambda.lifecycle.build.timeout",
		"io.knative.lambda.lifecycle.build.cancelled",

		// Response events
		"io.knative.lambda.response.success",
		"io.knative.lambda.response.error",
		"io.knative.lambda.response.timeout",
	}

	for _, expected := range expectedTypes {
		assert.Contains(t, types, expected, "Should have registered: %s", expected)
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ❌ VALIDATION ERROR STRUCTURE TESTS                                    │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidationErrorStructure(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	// Trigger a validation error
	err = v.Validate("io.knative.lambda.command.function.deploy", map[string]interface{}{})

	require.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)

	// Test error structure
	assert.Equal(t, "io.knative.lambda.command.function.deploy", validationErr.EventType)
	assert.NotEmpty(t, validationErr.Errors)

	// Test Error() method
	errorString := validationErr.Error()
	assert.Contains(t, errorString, "io.knative.lambda.command.function.deploy")
	assert.Contains(t, errorString, "schema validation failed")
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📦 SCHEMA CONSTANTS TESTS                                              │
// └─────────────────────────────────────────────────────────────────────────┘

func TestSchemaConstants(t *testing.T) {
	// Verify all schema constants are valid JSON
	schemas := map[string]string{
		"FunctionDeploySchema": FunctionDeploySchema,
		"ServiceDeleteSchema":  ServiceDeleteSchema,
		"BuildCommandSchema":   BuildCommandSchema,
		"InvokeSchema":         InvokeSchema,
		"LifecycleBuildSchema": LifecycleBuildSchema,
		"ResponseSchema":       ResponseSchema,
	}

	for name, schema := range schemas {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, schema, "%s should not be empty", name)
			assert.Contains(t, schema, `"$schema"`, "%s should have $schema", name)
			assert.Contains(t, schema, "2020-12", "%s should use draft 2020-12", name)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🏷️ LABELS AND ANNOTATIONS TESTS                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateLabelsAndAnnotations(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name        string
		metadata    map[string]interface{}
		expectError bool
	}{
		{
			name: "With valid labels",
			metadata: map[string]interface{}{
				"name": "labeled-func",
				"labels": map[string]interface{}{
					"app":         "my-app",
					"environment": "production",
					"team":        "platform",
				},
			},
			expectError: false,
		},
		{
			name: "With valid annotations",
			metadata: map[string]interface{}{
				"name": "annotated-func",
				"annotations": map[string]interface{}{
					"description":          "This is my function",
					"prometheus.io/scrape": "true",
					"prometheus.io/port":   "9090",
				},
			},
			expectError: false,
		},
		{
			name: "With both labels and annotations",
			metadata: map[string]interface{}{
				"name": "full-metadata-func",
				"labels": map[string]interface{}{
					"app": "test",
				},
				"annotations": map[string]interface{}{
					"note": "test annotation",
				},
			},
			expectError: false,
		},
		{
			name: "Invalid label value type (number)",
			metadata: map[string]interface{}{
				"name": "invalid-labels",
				"labels": map[string]interface{}{
					"count": 123, // Should be string
				},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]interface{}{
				"metadata": tc.metadata,
				"spec": map[string]interface{}{
					"source":  map[string]interface{}{"type": "inline", "inline": map[string]interface{}{"code": "x"}},
					"runtime": map[string]interface{}{"language": "python", "version": "3.11"},
				},
			}
			err := v.Validate("io.knative.lambda.command.function.deploy", data)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🎛️ EVENTING SPEC TESTS                                                 │
// └─────────────────────────────────────────────────────────────────────────┘

func TestValidateEventingSpec(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name        string
		eventing    map[string]interface{}
		expectError bool
	}{
		{
			name: "Eventing enabled",
			eventing: map[string]interface{}{
				"enabled": true,
			},
			expectError: false,
		},
		{
			name: "Eventing disabled",
			eventing: map[string]interface{}{
				"enabled": false,
			},
			expectError: false,
		},
		{
			name:        "Empty eventing spec",
			eventing:    map[string]interface{}{},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]interface{}{
				"metadata": map[string]interface{}{"name": "eventing-test"},
				"spec": map[string]interface{}{
					"source":   map[string]interface{}{"type": "inline", "inline": map[string]interface{}{"code": "x"}},
					"runtime":  map[string]interface{}{"language": "python", "version": "3.11"},
					"eventing": tc.eventing,
				},
			}
			err := v.Validate("io.knative.lambda.command.function.deploy", data)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
