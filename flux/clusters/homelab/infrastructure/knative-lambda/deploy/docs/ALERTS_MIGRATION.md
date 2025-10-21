# Knative Lambda Alerts Migration Guide

## Overview
The large `alerts.yaml` file has been broken down into smaller, more manageable files based on functionality. This improves maintainability, readability, and allows for better organization of alerts by component.

## New Alert Files Structure

### 1. Core System Alerts
- **`alerts-golden-signals.yaml`** - Golden signals (availability, error rate, latency, saturation)
- **`alerts-build-metrics.yaml`** - Build success rates, durations, queue depths
- **`alerts-service-creation.yaml`** - Service creation metrics and performance
- **`alerts-infrastructure.yaml`** - Infrastructure health (restarts, readiness, GC pressure)
- **`alerts-external-dependencies.yaml`** - External service dependencies (ECR, S3, K8s)
- **`alerts-business.yaml`** - Business logic metrics (build activity, eventing)

### 2. Component-Specific Alerts
- **`alerts-build-context-manager.yaml`** - Build context creation, S3 operations, validation
- **`alerts-event-handler.yaml`** - Event processing, validation, job management
- **`alerts-job-manager.yaml`** - Kubernetes job lifecycle management
- **`alerts-service-manager.yaml`** - Knative service management and resource creation
- **`alerts-cloud-event-handler.yaml`** - Cloud event processing and HTTP handling

### 3. Security Alerts
- **`alerts-security.yaml`** - Security threats, validation failures, attack detection

## Migration Status

### ✅ Completed Files
- `alerts-golden-signals.yaml` - Complete with all golden signal alerts
- `alerts-build-metrics.yaml` - Complete with all build metrics alerts
- `alerts-service-creation.yaml` - Complete with all service creation alerts
- `alerts-infrastructure.yaml` - Complete with all infrastructure alerts
- `alerts-external-dependencies.yaml` - Complete with all external dependency alerts
- `alerts-business.yaml` - Complete with all business logic alerts
- `alerts-security.yaml` - Complete with all security alerts

### 🔄 Partially Complete Files (Placeholders Created)
- `alerts-build-context-manager.yaml` - Contains 1 alert, needs full content migration
- `alerts-event-handler.yaml` - Contains 1 alert, needs full content migration
- `alerts-job-manager.yaml` - Contains 1 alert, needs full content migration
- `alerts-service-manager.yaml` - Contains 1 alert, needs full content migration
- `alerts-cloud-event-handler.yaml` - Contains 1 alert, needs full content migration

## Next Steps

### 1. Complete Content Migration
Move the remaining alert groups from the original `alerts.yaml` to their respective new files:

#### Build Context Manager Alerts (lines 451-755)
```bash
# Extract build context manager alerts from original file
# and add to alerts-build-context-manager.yaml
```

#### Event Handler Alerts (lines 756-1112)
```bash
# Extract event handler alerts from original file
# and add to alerts-event-handler.yaml
```

#### Job Manager Alerts (lines 1113-1499)
```bash
# Extract job manager alerts from original file
# and add to alerts-job-manager.yaml
```

#### Service Manager Alerts (lines 1500-1860)
```bash
# Extract service manager alerts from original file
# and add to alerts-service-manager.yaml
```

#### Cloud Event Handler Alerts (lines 1861-2189)
```bash
# Extract cloud event handler alerts from original file
# and add to alerts-cloud-event-handler.yaml
```

### 2. Update Helm Templates
Once all content is migrated, update the Helm chart to include all the new alert files:

```yaml
# In the main Helm template or values.yaml, ensure all files are included:
# - alerts-golden-signals.yaml
# - alerts-build-metrics.yaml
# - alerts-service-creation.yaml
# - alerts-infrastructure.yaml
# - alerts-external-dependencies.yaml
# - alerts-business.yaml
# - alerts-build-context-manager.yaml
# - alerts-event-handler.yaml
# - alerts-job-manager.yaml
# - alerts-service-manager.yaml
# - alerts-cloud-event-handler.yaml
# - alerts-security.yaml
```

### 3. Remove Original File
After successful migration and testing:
```bash
# Remove the original large alerts.yaml file
rm alerts.yaml
```

## Benefits of This Structure

1. **Maintainability** - Easier to find and modify specific alert types
2. **Readability** - Smaller files are easier to read and understand
3. **Team Ownership** - Different teams can own different alert files
4. **Version Control** - Better diff tracking and conflict resolution
5. **Testing** - Easier to test specific alert groups in isolation
6. **Deployment** - Can selectively deploy specific alert groups

## File Naming Convention
- All files follow the pattern: `alerts-{component-name}.yaml`
- Component names use kebab-case for consistency
- Each file contains a single PrometheusRule with one or more alert groups

## Validation
After migration, validate that:
1. All alerts are properly templated with `{{ .Values.environment }}`
2. All `$value` variables use the correct syntax: `{{ `{{ $value }}` }}`
3. All alert names are unique across all files
4. All required labels and annotations are present
5. Helm template rendering works without errors 