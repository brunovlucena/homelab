# 🔄 SRE-013: Schema Evolution and Event Compatibility

**Status**: Backlog
**Priority**: P1
**Story Points**: 8  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-176/sre-013-schema-evolution-and-event-compatibility  
**Created**: 2026-01-19  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---


## 📋 User Story

**As a** SRE Engineer  
**I want to** schema evolution and event compatibility  
**So that** I can improve system reliability, security, and performance

---



## 🎯 Acceptance Criteria

- [ ] [ ] Support multiple schema versions simultaneously
- [ ] [ ] Backward compatibility maintained for 90 days
- [ ] [ ] Schema validation errors logged with version info
- [ ] [ ] Incompatible events moved to DLQ with schema metadata
- [ ] [ ] Alert fires: "SchemaCompatibilityFailure"
- [ ] [ ] Schema registry tracks all versions
- [ ] [ ] Migration path documented per schema change
- [ ] --

---


## Overview

This runbook addresses schema evolution challenges in event-driven systems, focusing on backward/forward compatibility, version migration strategies, and handling incompatible events that land in Dead Letter Queues due to schema mismatches.

---

## 🎯 User Story: Handle Schema Evolution Without Event Loss

### Story

**As an** SRE Engineer  
**I want** to handle event schema changes gracefully  
**So that** version upgrades don't cause event processing failures or data loss

### Acceptance Criteria

- [ ] Support multiple schema versions simultaneously
- [ ] Backward compatibility maintained for 90 days
- [ ] Schema validation errors logged with version info
- [ ] Incompatible events moved to DLQ with schema metadata
- [ ] Alert fires: "SchemaCompatibilityFailure"
- [ ] Schema registry tracks all versions
- [ ] Migration path documented per schema change

---

## 📋 Schema Evolution Patterns

### Pattern 1: Additive Changes (Safe)

```yaml
# Version 1.0 - Original Schema
{
  "type": "network.notifi.lambda.build.start",
  "specversion": "1.0",
  "source": "api.notifi.com",
  "data": {
    "buildId": "string",
    "thirdPartyId": "string",
    "parserId": "string",
    "contextId": "string"
  }
}

# Version 1.1 - Added Optional Field (SAFE)
{
  "type": "network.notifi.lambda.build.start",
  "specversion": "1.0",
  "source": "api.notifi.com",
  "dataschema": "https://schemas.notifi.com/build/start/v1.1",
  "data": {
    "buildId": "string",
    "thirdPartyId": "string",
    "parserId": "string",
    "contextId": "string",
    "priority": "string"  # NEW: Optional field
  }
}

Compatibility: ✅ BACKWARD COMPATIBLE
- Old consumers: Ignore new field
- New consumers: Handle optional field
- No DLQ events expected
```

### Pattern 2: Breaking Changes (Unsafe)

```yaml
# Version 1.0 - Original Schema
{
  "data": {
    "buildId": "string",
    "thirdPartyId": "string"
  }
}

# Version 2.0 - Renamed Field (BREAKING)
{
  "dataschema": "https://schemas.notifi.com/build/start/v2.0",
  "data": {
    "buildId": "string",
    "organizationId": "string"  # RENAMED from thirdPartyId
  }
}

Compatibility: ❌ BREAKING CHANGE
- Old consumers: Cannot find "thirdPartyId" → FAIL
- New consumers: Cannot find "organizationId" in v1 events → FAIL
- Result: Events move to DLQ

Required Action:
1. Deploy schema adapter/transformer
2. Support both versions during transition
3. Migrate existing events
4. Deprecate old version after migration complete
```

---

## 💥 Failure Scenario: Schema Version Mismatch

```yaml
Timeline:

  T+0d: System running v1.0 schema
    ├─ Producers: Publishing v1.0 events
    ├─ Consumers: Processing v1.0 events
    └─ DLQ: Empty

  T+1d: Deploy v2.0 consumer (breaking change)
    ├─ Consumer v2.0: Expects "organizationId" field
    ├─ Producer: Still publishing v1.0 with "thirdPartyId"
    └─ Schema mismatch!

  T+1d+10m: Events start failing
    ├─ Consumer v2.0: "organizationId is required"
    ├─ Validation: FAIL
    ├─ Retry: 5 attempts (all fail)
    └─ DLQ: Event moved with "schema_validation_error"

  T+1d+1h: DLQ depth = 600 events
    └─ All events from last hour moved to DLQ

  T+1d+2h: Rollback consumer to v1.0
    ├─ Consumer v1.0: Redeployed
    ├─ New events: Processing successfully
    └─ DLQ events: Still need migration

  T+1d+4h: Fix and replay DLQ
    ├─ Schema transformer deployed
    ├─ DLQ events: Transformed v1→v2
    ├─ Events replayed successfully
    └─ DLQ: Empty
```

### Detection

```bash
# Check for schema validation errors in logs
kubectl logs -n knative-lambda -l app=knative-lambda-builder | \
  jq -r 'select(.error_type == "schema_validation") | [.timestamp, .event_type, .schema_version, .error] | @csv'

# Query Prometheus for schema errors
curl -g 'http://prometheus:9090/api/v1/query' \
  --data-urlencode 'query=rate(cloudevents_schema_validation_errors_total[5m])'

# Check DLQ for schema-related failures
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin get queue=lambda-build-events-prd-dlq count=100 | \
  jq -r '.[] | select(.properties.headers."x-death-reason" | contains("schema")) | [.payload.data.buildId, .payload.dataschema, .properties.headers."x-death-reason"] | @csv'

# Compare event schema versions in queue
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin get queue=lambda-build-events-prd count=100 requeue=true | \
  jq -r '.[].payload.dataschema' | sort | uniq -c
```

### Remediation - Schema Transformer

```go
// Schema Transformer Middleware
type SchemaTransformer struct {
    transforms map[string]TransformFunc
    obs        observability.Observability
}

type TransformFunc func(event *cloudevents.Event) (*cloudevents.Event, error)

func NewSchemaTransformer(obs observability.Observability) *SchemaTransformer {
    st := &SchemaTransformer{
        transforms: make(map[string]TransformFunc),
        obs:        obs,
    }
    
    // Register transformations
    st.RegisterTransform("v1.0->v2.0", st.transformV1ToV2)
    st.RegisterTransform("v2.0->v1.0", st.transformV2ToV1)
    
    return st
}

func (st *SchemaTransformer) RegisterTransform(key string, fn TransformFunc) {
    st.transforms[key] = fn
}

func (st *SchemaTransformer) Transform(ctx context.Context, event *cloudevents.Event, targetVersion string) (*cloudevents.Event, error) {
    currentVersion := st.getSchemaVersion(event)
    
    if currentVersion == targetVersion {
        // No transformation needed
        return event, nil
    }
    
    transformKey := fmt.Sprintf("%s->%s", currentVersion, targetVersion)
    transformFn, exists := st.transforms[transformKey]
    if !exists {
        return nil, fmt.Errorf("no transformation defined for %s", transformKey)
    }
    
    st.obs.Info(ctx, "Transforming event schema",
        "event_id", event.ID(),
        "from_version", currentVersion,
        "to_version", targetVersion)
    
    transformedEvent, err := transformFn(event)
    if err != nil {
        st.obs.Error(ctx, err, "Schema transformation failed",
            "event_id", event.ID(),
            "transform", transformKey)
        return nil, err
    }
    
    // Update schema version in event
    transformedEvent.SetExtension("dataschema", 
        fmt.Sprintf("https://schemas.notifi.com/build/start/%s", targetVersion))
    
    return transformedEvent, nil
}

func (st *SchemaTransformer) transformV1ToV2(event *cloudevents.Event) (*cloudevents.Event, error) {
    data := event.Data().(map[string]interface{})
    
    // Rename field: thirdPartyId → organizationId
    if thirdPartyID, ok := data["thirdPartyId"].(string); ok {
        data["organizationId"] = thirdPartyID
        delete(data, "thirdPartyId")
    }
    
    // Add new required field with default
    if _, ok := data["priority"]; !ok {
        data["priority"] = "normal"
    }
    
    // Create new event with transformed data
    newEvent := cloudevents.NewEvent()
    newEvent.SetID(event.ID())
    newEvent.SetType(event.Type())
    newEvent.SetSource(event.Source())
    newEvent.SetTime(event.Time())
    newEvent.SetData(event.DataContentType(), data)
    
    return &newEvent, nil
}

func (st *SchemaTransformer) transformV2ToV1(event *cloudevents.Event) (*cloudevents.Event, error) {
    data := event.Data().(map[string]interface{})
    
    // Rename field: organizationId → thirdPartyId
    if orgID, ok := data["organizationId"].(string); ok {
        data["thirdPartyId"] = orgID
        delete(data, "organizationId")
    }
    
    // Remove fields not in v1 schema
    delete(data, "priority")
    
    newEvent := cloudevents.NewEvent()
    newEvent.SetID(event.ID())
    newEvent.SetType(event.Type())
    newEvent.SetSource(event.Source())
    newEvent.SetTime(event.Time())
    newEvent.SetData(event.DataContentType(), data)
    
    return &newEvent, nil
}

func (st *SchemaTransformer) getSchemaVersion(event *cloudevents.Event) string {
    schema := event.DataSchema()
    if schema == "" {
        return "v1.0"  // Default to v1.0 if not specified
    }
    
    // Extract version from schema URL
    // e.g., "https://schemas.notifi.com/build/start/v2.0" → "v2.0"
    parts := strings.Split(schema, "/")
    if len(parts) > 0 {
        return parts[len(parts)-1]
    }
    
    return "v1.0"
}
```

### DLQ Replay with Schema Transformation

```bash
#!/bin/bash
# Replay DLQ events with schema transformation

DLQ_NAME="lambda-build-events-prd-dlq"
TARGET_EXCHANGE="knative-lambda-broker-prd"
ROUTING_KEY="lambda-build-events"
TRANSFORM_ENDPOINT="http://schema-transformer.knative-lambda:8080/transform"

echo "Replaying DLQ with schema transformation..."

# Get all DLQ messages
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqadmin get queue=$DLQ_NAME count=1000 > dlq_messages.json

# Count by schema version
echo "Schema versions in DLQ:"
jq -r '.[] | .payload.dataschema // "v1.0"' dlq_messages.json | sort | uniq -c

# Transform and replay each message
jq -c '.[]' dlq_messages.json | while IFS= read -r msg; do
  EVENT_ID=$(echo "$msg" | jq -r '.payload.id')
  SCHEMA_VERSION=$(echo "$msg" | jq -r '.payload.dataschema // "v1.0"')
  
  echo "Processing event $EVENT_ID (schema: $SCHEMA_VERSION)"
  
  # Call schema transformer service
  TRANSFORMED=$(echo "$msg" | jq -c '.payload' | curl -s -X POST \
    -H "Content-Type: application/json" \
    -d @- \
    "$TRANSFORM_ENDPOINT?target_version=v2.0")
  
  if [ $? -ne 0 ]; then
    echo "❌ Transformation failed for $EVENT_ID"
    continue
  fi
  
  # Publish transformed event
  echo "$TRANSFORMED" | kubectl exec -i -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
    rabbitmqadmin publish \
    exchange=$TARGET_EXCHANGE \
    routing_key=$ROUTING_KEY \
    payload=-
  
  echo "✅ Replayed $EVENT_ID with schema transformation"
  sleep 0.1
done

echo "Schema transformation replay complete"
```

---

## 🔧 Schema Registry Implementation

### JSON Schema Validation

```go
// Schema Validator with Registry
type SchemaValidator struct {
    registry map[string]*jsonschema.Schema
    obs      observability.Observability
}

func NewSchemaValidator(obs observability.Observability) (*SchemaValidator, error) {
    sv := &SchemaValidator{
        registry: make(map[string]*jsonschema.Schema),
        obs:      obs,
    }
    
    // Load schemas from ConfigMaps or external registry
    if err := sv.loadSchemas(); err != nil {
        return nil, err
    }
    
    return sv, nil
}

func (sv *SchemaValidator) loadSchemas() error {
    // Load v1.0 schema
    schemaV1 := `{
        "$schema": "http://json-schema.org/draft-07/schema#",
        "type": "object",
        "required": ["buildId", "thirdPartyId", "parserId", "contextId"],
        "properties": {
            "buildId": {"type": "string", "minLength": 1},
            "thirdPartyId": {"type": "string", "minLength": 1},
            "parserId": {"type": "string", "minLength": 1},
            "contextId": {"type": "string", "minLength": 1}
        }
    }`
    
    schemaV1Compiled, err := jsonschema.CompileString("v1.0", schemaV1)
    if err != nil {
        return fmt.Errorf("failed to compile v1.0 schema: %w", err)
    }
    sv.registry["v1.0"] = schemaV1Compiled
    
    // Load v2.0 schema
    schemaV2 := `{
        "$schema": "http://json-schema.org/draft-07/schema#",
        "type": "object",
        "required": ["buildId", "organizationId", "parserId", "contextId", "priority"],
        "properties": {
            "buildId": {"type": "string", "minLength": 1},
            "organizationId": {"type": "string", "minLength": 1},
            "parserId": {"type": "string", "minLength": 1},
            "contextId": {"type": "string", "minLength": 1},
            "priority": {"type": "string", "enum": ["low", "normal", "high", "urgent"]}
        }
    }`
    
    schemaV2Compiled, err := jsonschema.CompileString("v2.0", schemaV2)
    if err != nil {
        return fmt.Errorf("failed to compile v2.0 schema: %w", err)
    }
    sv.registry["v2.0"] = schemaV2Compiled
    
    return nil
}

func (sv *SchemaValidator) Validate(ctx context.Context, event *cloudevents.Event) error {
    schemaVersion := sv.getSchemaVersion(event)
    
    schema, exists := sv.registry[schemaVersion]
    if !exists {
        return fmt.Errorf("unknown schema version: %s", schemaVersion)
    }
    
    data := event.Data()
    
    if err := schema.Validate(data); err != nil {
        sv.obs.Error(ctx, err, "Schema validation failed",
            "event_id", event.ID(),
            "schema_version", schemaVersion)
        return fmt.Errorf("schema validation failed: %w", err)
    }
    
    sv.obs.Debug(ctx, "Schema validation passed",
        "event_id", event.ID(),
        "schema_version", schemaVersion)
    
    return nil
}

func (sv *SchemaValidator) getSchemaVersion(event *cloudevents.Event) string {
    schema := event.DataSchema()
    if schema == "" {
        return "v1.0"
    }
    
    parts := strings.Split(schema, "/")
    if len(parts) > 0 {
        return parts[len(parts)-1]
    }
    
    return "v1.0"
}
```

### Schema Registry ConfigMap

```yaml
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
  
  build-start-v2.0.json: | {
      "$schema": "http://json-schema.org/draft-07/schema#",
      "type": "object",
      "required": ["buildId", "organizationId", "parserId", "contextId", "priority"],
      "properties": {
        "buildId": {"type": "string", "minLength": 1},
        "organizationId": {"type": "string", "minLength": 1},
        "parserId": {"type": "string", "minLength": 1},
        "contextId": {"type": "string", "minLength": 1},
        "priority": {
          "type": "string",
          "enum": ["low", "normal", "high", "urgent"],
          "default": "normal"
        }
      }
    }
  
  schema-compatibility.json: | {
      "v1.0": {
        "forward_compatible": ["v1.1"],
        "backward_compatible": [],
        "deprecated": false,
        "deprecation_date": null,
        "end_of_life_date": null
      },
      "v1.1": {
        "forward_compatible": ["v2.0"],
        "backward_compatible": ["v1.0"],
        "deprecated": false
      },
      "v2.0": {
        "forward_compatible": [],
        "backward_compatible": ["v1.1"],
        "deprecated": false,
        "breaking_changes": [
          "Renamed thirdPartyId to organizationId",
          "Added required priority field"
        ]
      }
    }
```

---

## 📊 Schema Version Migration Strategy

### Phase 1: Preparation (Week 0)

```yaml
Actions:
  1. Document schema changes
     └─ Create migration guide: v1.0 → v2.0
  
  2. Deploy schema validator
     └─ Add validation to all consumers
  
  3. Add schema version tracking
     └─ Instrument Prometheus metrics
  
  4. Create transformation service
     └─ Deploy schema-transformer service
  
  5. Test in dev environment
     └─ Verify forward/backward compatibility
```

### Phase 2: Dual-Version Support (Week 1-4)

```yaml
Actions:
  1. Deploy updated consumers (support both v1.0 and v2.0)
     └─ Consumers handle both schemas
  
  2. Update producers to publish v2.0
     └─ Gradual rollout: 10% → 50% → 100%
  
  3. Monitor schema version distribution
     └─ Track v1.0 vs v2.0 events
  
  4. Transform DLQ events
     └─ Replay v1.0 events as v2.0
  
  5. Alert on v1.0 event ingestion
     └─ Identify producers still using v1.0
```

### Phase 3: Deprecation (Week 5-12)

```yaml
Actions:
  1. Mark v1.0 as deprecated
     └─ Add deprecation warnings to logs
  
  2. Contact remaining v1.0 producers
     └─ Request migration to v2.0
  
  3. Set end-of-life date for v1.0
     └─ 90 days from deprecation notice
  
  4. Final v1.0 → v2.0 transformation
     └─ Replay all remaining v1.0 events
  
  5. Remove v1.0 support
     └─ Consumers only support v2.0+
```

---

## 🚨 Monitoring & Alerts

```prometheus
# Alert: Schema Validation Failures
- alert: SchemaValidationFailureRate
  expr: | rate(cloudevents_schema_validation_errors_total[5m]) > 0.1
  for: 5m
  severity: warning
  annotations:
    summary: "High rate of schema validation failures"
    description: "{{ $value }} events/sec failing schema validation. Check for incompatible schema versions."

# Alert: Deprecated Schema Version Used
- alert: DeprecatedSchemaVersionInUse
  expr: | rate(cloudevents_received_total{schema_version="v1.0"}[5m]) > 0
    and ignoring(schema_version)
    schema_deprecated{schema_version="v1.0"} == 1
  for: 1h
  severity: info
  annotations:
    summary: "Deprecated schema version v1.0 still in use"
    description: "{{ $value }} v1.0 events/sec. Producers should migrate to v2.0."

# Metric: Schema Version Distribution
- metric: cloudevents_schema_version_distribution
  expr: | sum by (schema_version) (rate(cloudevents_received_total[5m]))

# Dashboard: Schema Migration Progress
- panel: "Schema Version Distribution Over Time"
  expr: | sum by (schema_version) (
      increase(cloudevents_received_total[1h])
    )
```

---

## 🔧 Best Practices

### Schema Evolution Guidelines

1. **Always Additive** - Prefer adding optional fields over renaming
2. **Version Everything** - Include schema version in all events
3. **Validate at Boundaries** - Validate at producer and consumer
4. **Test Compatibility** - Test forward/backward compatibility
5. **Document Changes** - Maintain schema changelog
6. **Deprecate Gracefully** - 90-day deprecation period minimum
7. **Transform in Transit** - Use middleware for schema transformation
8. **Monitor Adoption** - Track schema version distribution

### Compatibility Checklist

**Backward Compatible (Safe):**
- ✅ Add optional field
- ✅ Add new event type
- ✅ Widen field validation (e.g., longer string)
- ✅ Add default value for new field

**Breaking Changes (Unsafe):**
- ❌ Remove field
- ❌ Rename field
- ❌ Change field type
- ❌ Make optional field required
- ❌ Narrow field validation

---

## 📚 Related Documentation

- [SRE-010: Dead Letter Queue Management](./SRE-010-dead-letter-queue-management.md)
- [SRE-011: Event Ordering and Idempotency](./SRE-011-event-ordering-and-idempotency.md)
- [CloudEvents Specification](https://cloudevents.io/)
- [JSON Schema Documentation](https://json-schema.org/)

---

## Revision History | Version | Date | Author | Changes | |--------- | ------ | -------- | --------- | | 1.0.0 | 2025-10-29 | Bruno Lucena (Principal SRE) | Initial schema evolution and compatibility runbook |

