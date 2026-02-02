// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 BACKEND-008: Error Handling & Logging Tests
//
//	User Story: Error Handling & Logging
//	Priority: P1 | Story Points: 5
//
//	Tests validate:
//	- Error type creation and structure
//	- Error messages and formatting
//	- Error wrapping and context
//	- Error classification
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package backend

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalerrors "knative-lambda/internal/errors"
)

// TestBackend008_ValidationError validates ValidationError creation and formatting.
func TestBackend008_ValidationError(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		value       interface{}
		reason      string
		wantContain []string
	}{
		{
			name:   "String value validation",
			field:  "third_party_id",
			value:  "invalid-value",
			reason: "must be alphanumeric",
			wantContain: []string{
				"third_party_id",
				"invalid-value",
				"must be alphanumeric",
			},
		},
		{
			name:   "Empty field validation",
			field:  "parser_id",
			value:  "",
			reason: "cannot be empty",
			wantContain: []string{
				"parser_id",
				"cannot be empty",
			},
		},
		{
			name:   "Numeric value validation",
			field:  "timeout",
			value:  -1,
			reason: "must be positive",
			wantContain: []string{
				"timeout",
				"-1",
				"must be positive",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			err := internalerrors.NewValidationError(tt.field, tt.value, tt.reason)

			// Assert.
			require.Error(t, err)
			for _, want := range tt.wantContain {
				assert.Contains(t, err.Error(), want, "Error message should contain expected text")
			}

			// Verify type.
			validationErr, ok := err.(*internalerrors.ValidationError)
			require.True(t, ok, "Should be ValidationError type")
			assert.Equal(t, tt.field, validationErr.Field)
			assert.Equal(t, tt.value, validationErr.Value)
			assert.Equal(t, tt.reason, validationErr.Reason)
		})
	}
}

// TestBackend008_ConfigurationError validates ConfigurationError creation.
func TestBackend008_ConfigurationError(t *testing.T) {
	tests := []struct {
		name        string
		component   string
		setting     string
		reason      string
		wantContain []string
	}{
		{
			name:      "AWS configuration error",
			component: "aws",
			setting:   "credentials",
			reason:    "missing AWS credentials",
			wantContain: []string{
				"aws",
				"credentials",
				"missing AWS credentials",
			},
		},
		{
			name:      "Kubernetes configuration error",
			component: "kubernetes",
			setting:   "namespace",
			reason:    "invalid namespace format",
			wantContain: []string{
				"kubernetes",
				"namespace",
				"invalid namespace format",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			err := internalerrors.NewConfigurationError(tt.component, tt.setting, tt.reason)

			// Assert.
			require.Error(t, err)
			for _, want := range tt.wantContain {
				assert.Contains(t, err.Error(), want)
			}

			// Verify type.
			configErr, ok := err.(*internalerrors.ConfigurationError)
			require.True(t, ok, "Should be ConfigurationError type")
			assert.Equal(t, tt.component, configErr.Component)
			assert.Equal(t, tt.setting, configErr.Setting)
			assert.Equal(t, tt.reason, configErr.Reason)
		})
	}
}

// TestBackend008_SystemError validates SystemError creation.
func TestBackend008_SystemError(t *testing.T) {
	tests := []struct {
		name        string
		component   string
		operation   string
		wantContain []string
	}{
		{
			name:      "Kubernetes system error",
			component: "kubernetes",
			operation: "create_job",
			wantContain: []string{
				"kubernetes",
				"create_job",
			},
		},
		{
			name:      "AWS system error",
			component: "aws",
			operation: "upload_s3",
			wantContain: []string{
				"aws",
				"upload_s3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			err := internalerrors.NewSystemError(tt.component, tt.operation)

			// Assert.
			require.Error(t, err)
			for _, want := range tt.wantContain {
				assert.Contains(t, err.Error(), want)
			}

			// Verify type.
			sysErr, ok := err.(*internalerrors.SystemError)
			require.True(t, ok, "Should be SystemError type")
			assert.Equal(t, tt.component, sysErr.Component)
			assert.Equal(t, tt.operation, sysErr.Operation)
		})
	}
}

// TestBackend008_ErrorWrapping validates error wrapping with context.
func TestBackend008_ErrorWrapping(t *testing.T) {
	tests := []struct {
		name     string
		original error
		context  string
	}{
		{
			name:     "Wrap ValidationError",
			original: internalerrors.NewValidationError("test_field", "invalid", "test reason"),
			context:  "bucket=test-bucket, key=test-key",
		},
		{
			name:     "Wrap standard error",
			original: stderrors.New("original error"),
			context:  "operation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			wrapped := internalerrors.WrapWithContext(tt.original, tt.context)

			// Assert.
			require.Error(t, wrapped)
			assert.Contains(t, wrapped.Error(), tt.context)
		})
	}
}

// TestBackend008_ErrorTypeChecking validates error type checking functions.
func TestBackend008_ErrorTypeChecking(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		isNotFound bool
	}{
		{
			name:       "NotFoundError detection",
			err:        &internalerrors.NotFoundError{Resource: "job", Identifier: "test-job"},
			isNotFound: true,
		},
		{
			name:       "Non-NotFoundError detection",
			err:        internalerrors.NewValidationError("field", "value", "reason"),
			isNotFound: false,
		},
		{
			name:       "Nil error detection",
			err:        nil,
			isNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			result := internalerrors.IsNotFoundError(tt.err)

			// Assert.
			assert.Equal(t, tt.isNotFound, result)
		})
	}
}

// TestBackend008_ErrorConstants validates error constants are defined.
func TestBackend008_ErrorConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		notEmpty bool
	}{
		{"Config validation failed constant", internalerrors.ErrConfigValidationFailed, true},
		{"AWS region required constant", internalerrors.ErrAWSRegionRequired, true},
		{"K8s config failed constant", internalerrors.ErrK8sConfigFailed, true},
		{"Failed to create job constant", internalerrors.ErrFailedToCreateJob, true},
		{"S3 object not found constant", internalerrors.ErrS3ObjectNotFound, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.notEmpty {
				assert.NotEmpty(t, tt.constant)
			} else {
				assert.Empty(t, tt.constant)
			}
		})
	}
}

// TestBackend008_ValidationErrorWithContext validates validation errors with context.
func TestBackend008_ValidationErrorWithContext(t *testing.T) {
	// Arrange.
	err := &internalerrors.ValidationError{
		Field:   "test_field",
		Value:   "test_value",
		Reason:  "test reason",
		Context: "test context",
	}

	// Act.
	errMsg := err.Error()

	// Assert.
	assert.Contains(t, errMsg, "test_field")
	assert.Contains(t, errMsg, "test_value")
	assert.Contains(t, errMsg, "test reason")
	assert.Contains(t, errMsg, "test context")
}

// TestBackend008_WrapError validates WrapError utility function.
func TestBackend008_WrapError(t *testing.T) {
	// Arrange.
	originalErr := stderrors.New("original error")

	// Act.
	wrappedErr := internalerrors.WrapError(originalErr, "operation failed", "component", "test", "operation", "create")

	// Assert.
	require.Error(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), "operation failed")
	assert.Contains(t, wrappedErr.Error(), "original error")
	assert.Contains(t, wrappedErr.Error(), "component=test")
	assert.Contains(t, wrappedErr.Error(), "operation=create")
}

// TestBackend008_ValidateRequired validates field validation utility.
func TestBackend008_ValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     interface{}
		wantError bool
	}{
		{
			name:      "Valid string",
			fieldName: "test_field",
			value:     "valid_value",
			wantError: false,
		},
		{
			name:      "Empty string",
			fieldName: "test_field",
			value:     "",
			wantError: true,
		},
		{
			name:      "Nil value",
			fieldName: "test_field",
			value:     nil,
			wantError: true,
		},
		{
			name:      "Empty slice",
			fieldName: "test_field",
			value:     []string{},
			wantError: true,
		},
		{
			name:      "Valid slice",
			fieldName: "test_field",
			value:     []string{"item"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			err := internalerrors.ValidateRequired(tt.fieldName, tt.value)

			// Assert.
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.fieldName)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBackend008_ErrorIs validates error.Is compatibility.
func TestBackend008_ErrorIs(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "ValidationError matches itself",
			err:    &internalerrors.ValidationError{},
			target: &internalerrors.ValidationError{},
			want:   true,
		},
		{
			name:   "ConfigurationError matches itself",
			err:    &internalerrors.ConfigurationError{},
			target: &internalerrors.ConfigurationError{},
			want:   true,
		},
		{
			name:   "ValidationError does not match ConfigurationError",
			err:    &internalerrors.ValidationError{},
			target: &internalerrors.ConfigurationError{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			result := stderrors.Is(tt.err, tt.target)

			// Assert.
			assert.Equal(t, tt.want, result)
		})
	}
}
