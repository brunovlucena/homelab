package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSystemError(t *testing.T) {
	err := NewSystemError("test-component", "test-operation")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "system error")
	assert.Contains(t, err.Error(), "test-component")
	assert.Contains(t, err.Error(), "test-operation")
}

func TestNewConfigurationError(t *testing.T) {
	err := NewConfigurationError("test-component", "test-field", "test-message")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration error")
	assert.Contains(t, err.Error(), "test-component")
	assert.Contains(t, err.Error(), "test-field")
	assert.Contains(t, err.Error(), "test-message")
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("test-field", nil, "test-message")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, err.Error(), "test-field")
	assert.Contains(t, err.Error(), "test-message")
}

func TestNewConfigurationErrorWithValue(t *testing.T) {
	err := NewConfigurationErrorWithValue("test-component", "test-field", "test-value", "test-message")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration error")
	assert.Contains(t, err.Error(), "test-component")
	assert.Contains(t, err.Error(), "test-field")
	assert.Contains(t, err.Error(), "test-value")
	assert.Contains(t, err.Error(), "test-message")
}

func TestWrapWithContext(t *testing.T) {
	originalErr := errors.New("original error")
	err := WrapWithContext(originalErr, "additional context")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "original error")
	assert.Contains(t, err.Error(), "additional context")
}
