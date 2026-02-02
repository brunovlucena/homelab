# 🔄 SRE-013: Schema Evolution and Event Compatibility

**Linear URL**: https://linear.app/bvlucena/issue/BVL-195/sre-013-schema-evolution-and-event-compatibility  

---

## 📋 User Story

**As an** SRE Engineer  
**I want** to handle event schema changes gracefully  
**So that** version upgrades don't cause event processing failures or data loss


---


## 🎯 Acceptance Criteria

> **Note**: Features are already implemented. This ticket focuses on **validation** to ensure correctness, reliability, and production readiness.


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
  "type": "io.homelab.prometheus.alert.fired",
  "specversion": "1.0",
  "source": "prometheus-events",
  "data": {
    "alertname": "PodCPUHigh",
    "labels": {
      "pod": "app-xyz",
      "namespace": "production"
    }
  }
}

# Version 1.1 - Added Optional Field (SAFE)
{
  "type": "io.homelab.prometheus.alert.fired",
  "specversion": "1.0",
  "source": "prometheus-events",
  "dataschema": "https://schemas.homelab.com/alert/v1.1",
  "data": {
    "alertname": "PodCPUHigh",
    "labels": {
      "pod": "app-xyz",
      "namespace": "production"
    },
    "slo": "availability"  # NEW: Optional field
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
    "alertname": "PodCPUHigh",
    "labels": {
      "pod": "app-xyz"
    }
  }
}

# Version 2.0 - Renamed Field (BREAKING)
{
  "dataschema": "https://schemas.homelab.com/alert/v2.0",
  "data": {
    "alert_name": "PodCPUHigh",  # RENAMED from alertname
    "labels": {
      "pod": "app-xyz"
    }
  }
}

Compatibility: ❌ BREAKING CHANGE
- Old consumers: Cannot find "alertname" → FAIL
- New consumers: Cannot find "alert_name" in v1 events → FAIL
- Result: Events move to DLQ

Required Action:
1. Deploy schema adapter/transformer
2. Support both versions during transition
3. Migrate existing events
4. Deprecate old version after migration complete
```

---

## 🔧 Implementation Details

### Schema Transformer for Agent-SRE

Agent-sre should implement schema transformation middleware to handle version mismatches:

```python
# src/sre_agent/schema_transformer.py
from typing import Dict, Any, Optional
from cloudevents.http import CloudEvent

class SchemaTransformer:
    """Transform CloudEvents between schema versions."""
    
    def __init__(self):
        self.transforms = {
            "v1.0->v2.0": self._transform_v1_to_v2,
            "v2.0->v1.0": self._transform_v2_to_v1,
        }
    
    def transform(
        self,
        event: CloudEvent,
        target_version: str
    ) -> CloudEvent:
        """Transform event to target schema version."""
        current_version = self._get_schema_version(event)
        
        if current_version == target_version:
            return event
        
        transform_key = f"{current_version}->{target_version}"
        transform_fn = self.transforms.get(transform_key)
        
        if not transform_fn:
            raise ValueError(f"No transformation defined for {transform_key}")
        
        return transform_fn(event)
    
    def _get_schema_version(self, event: CloudEvent) -> str:
        """Extract schema version from event."""
        schema = event.get("dataschema", "")
        if not schema:
            return "v1.0"
        
        # Extract version from schema URL
        # e.g., "https://schemas.homelab.com/alert/v2.0" → "v2.0"
        parts = schema.split("/")
        return parts[-1] if parts else "v1.0"
```

---

## 📚 Related Documentation

- [SRE-010: Dead Letter Queue Management](./BVL-54-SRE-010-dead-letter-queue-management.md)
- [SRE-011: Event Ordering and Idempotency](./BVL-55-SRE-011-event-ordering-and-idempotency.md)
- [CloudEvents Specification](https://cloudevents.io/)

---

**Related Stories**:
- [BACKEND-001: CloudEvents Processing](./BVL-59-BACKEND-001-cloudevents-processing.md)


## 🧪 Test Scenarios

### Scenario 1: Additive Schema Change (Safe)
1. Deploy schema v1.1 with new optional field
2. Send events with v1.0 schema (no new field)
3. Verify events processed successfully
4. Send events with v1.1 schema (with new field)
5. Verify events processed successfully
6. Verify no DLQ events created
7. Verify backward compatibility maintained

### Scenario 2: Breaking Schema Change (Unsafe)
1. Deploy schema v2.0 with breaking change (renamed field)
2. Send events with v1.0 schema (old field name)
3. Verify events fail validation
4. Verify failed events moved to DLQ with schema metadata
5. Verify "SchemaCompatibilityFailure" alert fires
6. Deploy schema transformer/adapter
7. Verify v1.0 events transformed and processed successfully

### Scenario 3: Schema Validation Error Handling
1. Send event with invalid schema (missing required fields)
2. Verify event fails validation
3. Verify validation error logged with version info
4. Verify event moved to DLQ with schema metadata
5. Verify alert fires for schema validation failure
6. Verify schema registry tracks validation errors

### Scenario 4: Multiple Schema Versions Support
1. Deploy support for both v1.0 and v1.1 schemas
2. Send events with v1.0 schema
3. Verify v1.0 events processed correctly
4. Send events with v1.1 schema
5. Verify v1.1 events processed correctly
6. Verify both versions work simultaneously
7. Verify schema registry tracks both versions

### Scenario 5: Schema Migration Path
1. Deploy schema transformer for v1.0 -> v2.0 migration
2. Send events with v1.0 schema
3. Verify events transformed to v2.0 schema
4. Verify transformed events processed successfully
5. Verify migration path documented
6. Verify both versions supported during migration
7. Deprecate v1.0 after migration complete

### Scenario 6: Schema Registry Tracking
1. Deploy new schema version
2. Verify schema registry tracks new version
3. Send events with new version
4. Verify schema registry records version usage
5. Verify schema registry shows all versions
6. Verify schema registry queries work correctly
7. Verify schema registry alerts configured

### Scenario 7: High Load Schema Validation
1. Send 1000+ events with various schema versions
2. Verify schema validation handles load (< 50ms per event)
3. Verify no events lost during validation
4. Verify schema compatibility checks work under load
5. Verify metrics and alerts work under load
6. Verify DLQ handling works under load

## 📊 Success Metrics

- **Schema Validation Time**: < 50ms per event (P95)
- **Backward Compatibility**: 100% (90 days)
- **Schema Validation Success Rate**: > 99%
- **DLQ Events from Schema Issues**: < 1% of total events
- **Schema Migration Success Rate**: > 95%
- **Test Pass Rate**: 100%

---

**Last Updated**: January 08, 2026
**Owner**: SRE Team
**Status**: Validation Required