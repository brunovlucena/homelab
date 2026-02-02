# 🌐 BACKEND-012: Schema Validation and Registry

**Priority**: P1 | **Status**: ✅ Implemented  | **Story Points**: 8
**Linear URL**: https://linear.app/bvlucena/issue/BVL-232/backend-012-schema-validation-and-registry

---

## 📋 User Story

**As a** Backend Developer  
**I want to** validate CloudEvent schemas and support multiple schema versions  
**So that** incompatible events are rejected early and schema evolution is managed safely

---

## 🎯 Acceptance Criteria

### ✅ Schema Registry
- [ ] Load JSON schemas from ConfigMaps
- [ ] Support multiple schema versions per event type
- [ ] Register schemas at startup
- [ ] Hot-reload schemas on ConfigMap changes
- [ ] Track schema versions in metrics

### ✅ Schema Validation
- [ ] Validate CloudEvent data against JSON schema
- [ ] Return detailed validation errors
- [ ] Log schema violations with event context
- [ ] Emit metrics for validation failures
- [ ] Support schema version from `dataschema` field

### ✅ Compatibility Checking
- [ ] Define forward/backward compatibility rules
- [ ] Warn on deprecated schema usage
- [ ] Block events using unsupported schemas
- [ ] Track schema version distribution

### ✅ Observability
- [ ] Metric: `schema_validation_errors_total` counter
- [ ] Metric: `schema_version_distribution` gauge
- [ ] Log: Validation errors with schema diff
- [ ] Trace: Schema validation span

---

## 🔧 Technical Implementation

### File: `internal/handler/schema_validator.go`

```go
package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"knative-lambda/internal/observability"
)

// 📋 SchemaValidator - "Validate CloudEvents against JSON schemas"
type SchemaValidator struct {
	registry map[string]*jsonschema.Schema
	obs      observability.Observability
}

// 🏗️ NewSchemaValidator - "Create new schema validator with registry"
func NewSchemaValidator(obs observability.Observability) (*SchemaValidator, error) {
	sv := &SchemaValidator{
		registry: make(map[string]*jsonschema.Schema),
		obs:      obs,
	}
	
	// Load schemas from ConfigMaps or embedded
	if err := sv.loadSchemas(); err != nil {
		return nil, err
	}
	
	return sv, nil
}

// 📚 loadSchemas - "Load JSON schemas into registry"
func (sv *SchemaValidator) loadSchemas() error {
	// Load v1.0 schema
	schemaV1 := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["buildId", "thirdPartyId", "parserId", "contextId"],
		"properties": {
			"buildId": {
				"type": "string",
				"minLength": 1,
				"pattern": "^[a-zA-Z0-9-]+$"
			},
			"thirdPartyId": {
				"type": "string",
				"minLength": 1
			},
			"parserId": {
				"type": "string",
				"minLength": 1
			},
			"contextId": {
				"type": "string",
				"minLength": 1
			},
			"eventSequence": {
				"type": "integer",
				"minimum": 1
			}
		},
		"additionalProperties": true
	}`
	
	schemaV1Compiled, err := jsonschema.CompileString("v1.0", schemaV1)
	if err != nil {
		return fmt.Errorf("failed to compile v1.0 schema: %w", err)
	}
	sv.registry["v1.0"] = schemaV1Compiled
	
	// Load v2.0 schema (with breaking changes)
	schemaV2 := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["buildId", "organizationId", "parserId", "contextId", "priority"],
		"properties": {
			"buildId": {
				"type": "string",
				"minLength": 1,
				"pattern": "^[a-zA-Z0-9-]+$"
			},
			"organizationId": {
				"type": "string",
				"minLength": 1
			},
			"parserId": {
				"type": "string",
				"minLength": 1
			},
			"contextId": {
				"type": "string",
				"minLength": 1
			},
			"priority": {
				"type": "string",
				"enum": ["low", "normal", "high", "urgent"]
			},
			"eventSequence": {
				"type": "integer",
				"minimum": 1
			}
		},
		"additionalProperties": true
	}`
	
	schemaV2Compiled, err := jsonschema.CompileString("v2.0", schemaV2)
	if err != nil {
		return fmt.Errorf("failed to compile v2.0 schema: %w", err)
	}
	sv.registry["v2.0"] = schemaV2Compiled
	
	return nil
}

// ✅ Validate - "Validate CloudEvent data against schema"
func (sv *SchemaValidator) Validate(ctx context.Context, eventType string, dataSchema string, data interface{}) error {
	ctx, span := sv.obs.StartSpan(ctx, "schema_validation")
	defer span.End()
	
	// Extract schema version
	schemaVersion := sv.extractVersion(dataSchema)
	if schemaVersion == "" {
		schemaVersion = "v1.0" // Default to v1.0 for backwards compatibility
	}
	
	// Get schema from registry
	schema, exists := sv.registry[schemaVersion]
	if !exists {
		err := fmt.Errorf("unknown schema version: %s", schemaVersion)
		sv.obs.Error(ctx, err, "Schema version not found",
			"event_type", eventType,
			"schema_version", schemaVersion)
		return err
	}
	
	// Validate data
	if err := schema.Validate(data); err != nil {
		// Extract validation details
		validationErr := err.(*jsonschema.ValidationError)
		
		sv.obs.Error(ctx, err, "Schema validation failed",
			"event_type", eventType,
			"schema_version", schemaVersion,
			"validation_error", validationErr.Error(),
			"failed_path", validationErr.InstanceLocation)
		
		return fmt.Errorf("schema validation failed: %w", err)
	}
	
	sv.obs.Debug(ctx, "Schema validation passed",
		"event_type", eventType,
		"schema_version", schemaVersion)
	
	return nil
}

// 🔍 extractVersion - "Extract version from schema URL"
func (sv *SchemaValidator) extractVersion(dataSchema string) string {
	if dataSchema == "" {
		return ""
	}
	
	// Extract version from URL like "https://schemas.notifi.com/build/start/v2.0"
	parts := strings.Split(dataSchema, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	
	return ""
}

// 📊 GetSupportedVersions - "Get list of supported schema versions"
func (sv *SchemaValidator) GetSupportedVersions() []string {
	versions := make([]string, 0, len(sv.registry))
	for version := range sv.registry {
		versions = append(versions, version)
	}
	return versions
}

// 🔄 RegisterSchema - "Dynamically register a new schema version"
func (sv *SchemaValidator) RegisterSchema(version string, schemaJSON string) error {
	schema, err := jsonschema.CompileString(version, schemaJSON)
	if err != nil {
		return fmt.Errorf("failed to compile schema %s: %w", version, err)
	}
	
	sv.registry[version] = schema
	return nil
}
```

### File: `internal/handler/event_handler.go` (Integration)

```go
// 📥 ProcessCloudEvent - "Process CloudEvent with schema validation"
func (h *EventHandlerImpl) ProcessCloudEvent(ctx context.Context, event *cloudevents.Event) (*builds.HandlerResponse, error) {
	ctx, span := h.obs.StartSpan(ctx, "process_cloud_event")
	defer span.End()

	// Validate event schema
	ctx, schemaSpan := h.obs.StartSpan(ctx, "schema_validation")
	err := h.schemaValidator.Validate(
		ctx,
		event.Type(),
		event.DataSchema(),
		event.Data(),
	)
	if err != nil {
		h.metrics.SchemaValidationErrors.WithLabelValues(event.Type(), "validation_failed").Inc()
		schemaSpan.SetStatus(codes.Error, err.Error())
		schemaSpan.End()
		
		// Schema validation failure → DLQ
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}
	schemaSpan.End()
	
	// Track schema version usage
	schemaVersion := h.extractSchemaVersion(event.DataSchema())
	h.metrics.SchemaVersionDistribution.WithLabelValues(event.Type(), schemaVersion).Inc()
	
	// Continue with normal processing...
	return h.processEventWithTracing(ctx, event, h.metricsRec)
}

func (h *EventHandlerImpl) extractSchemaVersion(dataSchema string) string {
	if dataSchema == "" {
		return "v1.0"
	}
	parts := strings.Split(dataSchema, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "v1.0"
}
```

---

## 🧪 Test Cases

### File: `internal/handler/schema_validator_test.go`

```go
package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"knative-lambda/internal/observability"
)

// Test 1: Valid event - v1.0 schema
func TestSchemaValidator_ValidV1(t *testing.T) {
	ctx := context.Background()
	obs := observability.NewMockObservability()
	
	validator, err := NewSchemaValidator(obs)
	require.NoError(t, err)
	
	data := map[string]interface{}{
		"buildId":       "build-123",
		"thirdPartyId":  "tp-456",
		"parserId":      "parser-789",
		"contextId":     "ctx-abc",
		"eventSequence": float64(1),
	}
	
	err = validator.Validate(ctx, "network.notifi.lambda.build.start", "v1.0", data)
	require.NoError(t, err)
}

// Test 2: Valid event - v2.0 schema
func TestSchemaValidator_ValidV2(t *testing.T) {
	ctx := context.Background()
	obs := observability.NewMockObservability()
	
	validator, err := NewSchemaValidator(obs)
	require.NoError(t, err)
	
	data := map[string]interface{}{
		"buildId":        "build-123",
		"organizationId": "org-456",
		"parserId":       "parser-789",
		"contextId":      "ctx-abc",
		"priority":       "high",
		"eventSequence":  float64(1),
	}
	
	err = validator.Validate(ctx, "network.notifi.lambda.build.start", "v2.0", data)
	require.NoError(t, err)
}

// Test 3: Missing required field
func TestSchemaValidator_MissingField(t *testing.T) {
	ctx := context.Background()
	obs := observability.NewMockObservability()
	
	validator, err := NewSchemaValidator(obs)
	require.NoError(t, err)
	
	data := map[string]interface{}{
		"buildId":  "build-123",
		// Missing: thirdPartyId, parserId, contextId
	}
	
	err = validator.Validate(ctx, "network.notifi.lambda.build.start", "v1.0", data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema validation failed")
}

// Test 4: Invalid field type
func TestSchemaValidator_InvalidType(t *testing.T) {
	ctx := context.Background()
	obs := observability.NewMockObservability()
	
	validator, err := NewSchemaValidator(obs)
	require.NoError(t, err)
	
	data := map[string]interface{}{
		"buildId":       123, // Should be string, not number
		"thirdPartyId":  "tp-456",
		"parserId":      "parser-789",
		"contextId":     "ctx-abc",
	}
	
	err = validator.Validate(ctx, "network.notifi.lambda.build.start", "v1.0", data)
	require.Error(t, err)
}

// Test 5: Empty string (violates minLength)
func TestSchemaValidator_EmptyString(t *testing.T) {
	ctx := context.Background()
	obs := observability.NewMockObservability()
	
	validator, err := NewSchemaValidator(obs)
	require.NoError(t, err)
	
	data := map[string]interface{}{
		"buildId":      "", // Empty string violates minLength: 1
		"thirdPartyId": "tp-456",
		"parserId":     "parser-789",
		"contextId":    "ctx-abc",
	}
	
	err = validator.Validate(ctx, "network.notifi.lambda.build.start", "v1.0", data)
	require.Error(t, err)
}

// Test 6: Invalid enum value (v2.0 priority)
func TestSchemaValidator_InvalidEnum(t *testing.T) {
	ctx := context.Background()
	obs := observability.NewMockObservability()
	
	validator, err := NewSchemaValidator(obs)
	require.NoError(t, err)
	
	data := map[string]interface{}{
		"buildId":        "build-123",
		"organizationId": "org-456",
		"parserId":       "parser-789",
		"contextId":      "ctx-abc",
		"priority":       "critical", // Invalid: not in enum [low, normal, high, urgent]
	}
	
	err = validator.Validate(ctx, "network.notifi.lambda.build.start", "v2.0", data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema validation failed")
}

// Test 7: Unknown schema version
func TestSchemaValidator_UnknownVersion(t *testing.T) {
	ctx := context.Background()
	obs := observability.NewMockObservability()
	
	validator, err := NewSchemaValidator(obs)
	require.NoError(t, err)
	
	data := map[string]interface{}{
		"buildId": "build-123",
	}
	
	err = validator.Validate(ctx, "network.notifi.lambda.build.start", "v99.0", data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown schema version")
}

// Test 8: Default to v1.0 when no schema specified
func TestSchemaValidator_DefaultVersion(t *testing.T) {
	ctx := context.Background()
	obs := observability.NewMockObservability()
	
	validator, err := NewSchemaValidator(obs)
	require.NoError(t, err)
	
	data := map[string]interface{}{
		"buildId":      "build-123",
		"thirdPartyId": "tp-456",
		"parserId":     "parser-789",
		"contextId":    "ctx-abc",
	}
	
	// Empty dataSchema should default to v1.0
	err = validator.Validate(ctx, "network.notifi.lambda.build.start", "", data)
	require.NoError(t, err)
}

// Test 9: Extract version from schema URL
func TestSchemaValidator_ExtractVersion(t *testing.T) {
	obs := observability.NewMockObservability()
	validator, _ := NewSchemaValidator(obs)
	
	tests := []struct {
		url      string
		expected string
	}{
		{"https://schemas.notifi.com/build/start/v2.0", "v2.0"},
		{"https://schemas.notifi.com/build/start/v1.0", "v1.0"},
		{"v2.0", "v2.0"},
		{"", ""},
	}
	
	for _, tt := range tests {
		result := validator.extractVersion(tt.url)
		assert.Equal(t, tt.expected, result, "URL: %s", tt.url)
	}
}

// Test 10: Get supported versions
func TestSchemaValidator_GetSupportedVersions(t *testing.T) {
	obs := observability.NewMockObservability()
	validator, err := NewSchemaValidator(obs)
	require.NoError(t, err)
	
	versions := validator.GetSupportedVersions()
	
	assert.Contains(t, versions, "v1.0")
	assert.Contains(t, versions, "v2.0")
	assert.Len(t, versions, 2)
}

// Test 11: Dynamic schema registration
func TestSchemaValidator_RegisterSchema(t *testing.T) {
	obs := observability.NewMockObservability()
	validator, err := NewSchemaValidator(obs)
	require.NoError(t, err)
	
	// Register v3.0 schema dynamically
	schemaV3 := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["buildId"],
		"properties": {
			"buildId": {"type": "string"}
		}
	}`
	
	err = validator.RegisterSchema("v3.0", schemaV3)
	require.NoError(t, err)
	
	// Validate using new schema
	data := map[string]interface{}{
		"buildId": "build-v3",
	}
	
	err = validator.Validate(context.Background(), "test", "v3.0", data)
	require.NoError(t, err)
}

// Test 12: Pattern validation (buildId format)
func TestSchemaValidator_PatternValidation(t *testing.T) {
	ctx := context.Background()
	obs := observability.NewMockObservability()
	
	validator, err := NewSchemaValidator(obs)
	require.NoError(t, err)
	
	tests := []struct {
		buildID  string
		valid    bool
	}{
		{"build-123", true},
		{"BUILD-ABC", true},
		{"build_123", false},  // Underscore not allowed
		{"build@123", false},  // Special char not allowed
		{"build 123", false},  // Space not allowed
	}
	
	for _, tt := range tests {
		data := map[string]interface{}{
			"buildId":      tt.buildID,
			"thirdPartyId": "tp-456",
			"parserId":     "parser-789",
			"contextId":    "ctx-abc",
		}
		
		err = validator.Validate(ctx, "test", "v1.0", data)
		if tt.valid {
			assert.NoError(t, err, "buildID: %s should be valid", tt.buildID)
		} else {
			assert.Error(t, err, "buildID: %s should be invalid", tt.buildID)
		}
	}
}

// Test 13: Optional fields validation
func TestSchemaValidator_OptionalFields(t *testing.T) {
	ctx := context.Background()
	obs := observability.NewMockObservability()
	
	validator, err := NewSchemaValidator(obs)
	require.NoError(t, err)
	
	// eventSequence is optional
	data := map[string]interface{}{
		"buildId":      "build-123",
		"thirdPartyId": "tp-456",
		"parserId":     "parser-789",
		"contextId":    "ctx-abc",
		// eventSequence omitted
	}
	
	err = validator.Validate(ctx, "test", "v1.0", data)
	require.NoError(t, err, "Optional fields should be allowed to be missing")
}
```

---

## 📊 Metrics

```prometheus
# Counter: Schema validation errors
schema_validation_errors_total{event_type="network.notifi.lambda.build.start",reason="missing_field"}

# Gauge: Schema version distribution
schema_version_distribution{event_type="network.notifi.lambda.build.start",version="v1.0"}

# Histogram: Schema validation duration
schema_validation_duration_seconds_bucket{le="0.001"}
```

---

## 🔄 Configuration

```yaml
# ConfigMap with JSON schemas
apiVersion: v1
kind: ConfigMap
metadata:
  name: event-schemas
  namespace: knative-lambda
data:
  build-start-v1.0.json: | {
      "$schema": "http://json-schema.org/draft-07/schema#",
      "type": "object",
      "required": ["buildId", "thirdPartyId", "parserId", "contextId"],
      "properties": {
        "buildId": {"type": "string", "minLength": 1},
        "thirdPartyId": {"type": "string", "minLength": 1},
        "parserId": {"type": "string", "minLength": 1},
        "contextId": {"type": "string", "minLength": 1}
      }
    }
  
  schema-compatibility.json: | {
      "v1.0": {
        "deprecated": false,
        "end_of_life": null
      },
      "v2.0": {
        "deprecated": false,
        "breaking_changes": ["Renamed thirdPartyId to organizationId"]
      }
    }
```

---

## 🔗 Related Stories

- [BACKEND-010: Idempotency and Duplicate Detection](./BACKEND-010-idempotency-duplicate-detection.md)
- [BACKEND-001: CloudEvents Processing](./BACKEND-001-cloudevents-processing.md)
- [SRE-013: Schema Evolution and Compatibility](../../sre/user-stories/SRE-013-schema-evolution-compatibility.md)

---

## Revision History | Version | Date | Author | Changes | |--------- | ------ | -------- | --------- | | 1.0.0 | 2025-10-29 | Bruno Lucena | Initial schema validation story |

