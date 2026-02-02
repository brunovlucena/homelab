// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🔐 SCHEMA VALIDATOR - CloudEvents JSON Schema Validation
//
//	This package provides JSON Schema validation for all CloudEvent types
//	used by the Knative Lambda Operator.
//
//	Features:
//	- Strict validation of CloudEvent payloads
//	- Detailed error messages with JSON paths
//	- Schema versioning support
//	- Pre-compiled schemas for performance
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package schema

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📋 VALIDATOR                                                           │
// └─────────────────────────────────────────────────────────────────────────┘

// Validator validates CloudEvent payloads against JSON schemas
type Validator struct {
	compiler *jsonschema.Compiler
	schemas  map[string]*jsonschema.Schema
	mu       sync.RWMutex
}

// ValidationError represents a schema validation error
type ValidationError struct {
	EventType string   `json:"eventType"`
	Errors    []string `json:"errors"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("schema validation failed for event type %s: %s", e.EventType, strings.Join(e.Errors, "; "))
}

// NewValidator creates a new schema validator with pre-compiled schemas
func NewValidator() (*Validator, error) {
	v := &Validator{
		compiler: jsonschema.NewCompiler(),
		schemas:  make(map[string]*jsonschema.Schema),
	}

	// Configure compiler for strict validation
	v.compiler.Draft = jsonschema.Draft2020

	// Register all schemas
	if err := v.registerSchemas(); err != nil {
		return nil, fmt.Errorf("failed to register schemas: %w", err)
	}

	return v, nil
}

// Validate validates the given data against the schema for the specified event type
func (v *Validator) Validate(eventType string, data interface{}) error {
	v.mu.RLock()
	schema, ok := v.schemas[eventType]
	v.mu.RUnlock()

	if !ok {
		// Check for wildcard patterns
		schema = v.findWildcardSchema(eventType)
		if schema == nil {
			return &ValidationError{
				EventType: eventType,
				Errors:    []string{fmt.Sprintf("no schema registered for event type: %s", eventType)},
			}
		}
	}

	// Convert data to JSON and back to ensure proper type handling
	jsonData, err := json.Marshal(data)
	if err != nil {
		return &ValidationError{
			EventType: eventType,
			Errors:    []string{fmt.Sprintf("failed to marshal data: %v", err)},
		}
	}

	var jsonObj interface{}
	if err := json.Unmarshal(jsonData, &jsonObj); err != nil {
		return &ValidationError{
			EventType: eventType,
			Errors:    []string{fmt.Sprintf("failed to unmarshal data: %v", err)},
		}
	}

	// Validate against schema
	if err := schema.Validate(jsonObj); err != nil {
		validationErr, ok := err.(*jsonschema.ValidationError)
		if ok {
			return &ValidationError{
				EventType: eventType,
				Errors:    extractValidationErrors(validationErr),
			}
		}
		return &ValidationError{
			EventType: eventType,
			Errors:    []string{err.Error()},
		}
	}

	return nil
}

// ValidateJSON validates raw JSON bytes against the schema for the specified event type
func (v *Validator) ValidateJSON(eventType string, jsonData []byte) error {
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return &ValidationError{
			EventType: eventType,
			Errors:    []string{fmt.Sprintf("invalid JSON: %v", err)},
		}
	}
	return v.Validate(eventType, data)
}

// HasSchema returns true if a schema is registered for the given event type
func (v *Validator) HasSchema(eventType string) bool {
	v.mu.RLock()
	_, ok := v.schemas[eventType]
	v.mu.RUnlock()
	if ok {
		return true
	}
	return v.findWildcardSchema(eventType) != nil
}

// RegisteredEventTypes returns all registered event types
func (v *Validator) RegisteredEventTypes() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	types := make([]string, 0, len(v.schemas))
	for t := range v.schemas {
		types = append(types, t)
	}
	return types
}

// findWildcardSchema finds a schema using wildcard matching
func (v *Validator) findWildcardSchema(eventType string) *jsonschema.Schema {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// Check for category-level wildcards (e.g., "io.knative.lambda.command.*")
	parts := strings.Split(eventType, ".")
	if len(parts) >= 4 {
		// Try io.knative.lambda.<category>.*
		wildcardType := strings.Join(parts[:4], ".") + ".*"
		if schema, ok := v.schemas[wildcardType]; ok {
			return schema
		}
	}

	return nil
}

// registerSchemas registers all known schemas
func (v *Validator) registerSchemas() error {
	schemas := map[string]string{
		// Command events
		"io.knative.lambda.command.function.deploy": FunctionDeploySchema,
		"io.knative.lambda.command.service.create":  FunctionDeploySchema, // Alias
		"io.knative.lambda.command.service.update":  FunctionDeploySchema, // Alias
		"io.knative.lambda.command.service.delete":  ServiceDeleteSchema,
		"io.knative.lambda.command.build.start":     BuildCommandSchema,
		"io.knative.lambda.command.build.cancel":    BuildCommandSchema,
		"io.knative.lambda.command.build.retry":     BuildCommandSchema,

		// Invoke events
		"io.knative.lambda.invoke.sync":      InvokeSchema,
		"io.knative.lambda.invoke.async":     InvokeSchema,
		"io.knative.lambda.invoke.scheduled": InvokeSchema,

		// Lifecycle events (response schemas)
		"io.knative.lambda.lifecycle.build.started":   LifecycleBuildSchema,
		"io.knative.lambda.lifecycle.build.completed": LifecycleBuildSchema,
		"io.knative.lambda.lifecycle.build.failed":    LifecycleBuildSchema,
		"io.knative.lambda.lifecycle.build.timeout":   LifecycleBuildSchema,
		"io.knative.lambda.lifecycle.build.cancelled": LifecycleBuildSchema,

		// Response events
		"io.knative.lambda.response.success": ResponseSchema,
		"io.knative.lambda.response.error":   ResponseSchema,
		"io.knative.lambda.response.timeout": ResponseSchema,
	}

	for eventType, schemaJSON := range schemas {
		if err := v.compiler.AddResource(eventType, strings.NewReader(schemaJSON)); err != nil {
			return fmt.Errorf("failed to add schema for %s: %w", eventType, err)
		}

		schema, err := v.compiler.Compile(eventType)
		if err != nil {
			return fmt.Errorf("failed to compile schema for %s: %w", eventType, err)
		}

		v.schemas[eventType] = schema
	}

	return nil
}

// extractValidationErrors extracts human-readable errors from validation result
func extractValidationErrors(err *jsonschema.ValidationError) []string {
	var errors []string
	extractErrors(err, &errors, "")
	return errors
}

func extractErrors(err *jsonschema.ValidationError, errors *[]string, path string) {
	currentPath := path
	if err.InstanceLocation != "" {
		currentPath = err.InstanceLocation
	}

	if err.Message != "" {
		errorMsg := err.Message
		if currentPath != "" {
			errorMsg = fmt.Sprintf("%s: %s", currentPath, err.Message)
		}
		*errors = append(*errors, errorMsg)
	}

	for _, cause := range err.Causes {
		extractErrors(cause, errors, currentPath)
	}
}
