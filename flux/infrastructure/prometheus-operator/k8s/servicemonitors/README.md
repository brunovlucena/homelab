# ServiceMonitor Resources

This directory contains ServiceMonitor custom resources for Prometheus metric scraping.

## ⚠️ CRITICAL: Deployment Order

**ServiceMonitors MUST be deployed AFTER Prometheus Operator CRDs are available!**

### Why?

ServiceMonitor is a CustomResourceDefinition (CRD) provided by the Prometheus Operator. If you try to create a ServiceMonitor before the CRD exists, Flux will fail with:

```
ServiceMonitor/.../... dry-run failed: no matches for kind "ServiceMonitor" in version "monitoring.coreos.com/v1"
```

### Deployment Strategy

1. **Level 1 (01-core)**: Infrastructure components (NO ServiceMonitors)
   - Core services deploy first
   - ServiceMonitors are NOT included here

2. **Level 2 (02-observability)**: Prometheus Operator installation
   - Installs `kube-prometheus-stack` HelmRelease
   - Creates ServiceMonitor CRD

3. **Level 2b (02b-observability-extras)**: ServiceMonitors deployment
   - All ServiceMonitors deploy here
   - CRDs are guaranteed to exist

### Moving ServiceMonitors from Level 1

If you need to add a ServiceMonitor for a level 1 component:

1. **DO NOT** add it to the component's kustomization.yaml
2. **DO** add it to this directory (`prometheus-operator/k8s/servicemonitors/`)
3. **DO** update `servicemonitors/kustomization.yaml` to include it
4. **DO** remove any ServiceMonitor from the level 1 component

### Example Migration

**Before (❌ WRONG - in level 1 component):**
```yaml
# infrastructure/my-component/kustomization.yaml
resources:
  - deployment.yaml
  - service.yaml
  - servicemonitor.yaml  # ❌ This will fail!
```

**After (✅ CORRECT):**
```yaml
# infrastructure/my-component/kustomization.yaml
resources:
  - deployment.yaml
  - service.yaml
  # ServiceMonitor moved to prometheus-operator/k8s/servicemonitors/
```

```yaml
# prometheus-operator/k8s/servicemonitors/my-component-servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-component
  namespace: my-component-namespace
  # ...
```

```yaml
# prometheus-operator/k8s/servicemonitors/kustomization.yaml
resources:
  - my-component-servicemonitor.yaml
```

## ⚠️ CRITICAL: Kustomize vs Helm Template Syntax

**These files are processed by KUSTOMIZE, NOT HELM!**

### ❌ DO NOT USE Helm Template Syntax

**NEVER** use Helm template syntax in ServiceMonitor files:
- ❌ `{{ .Release.Namespace }}`
- ❌ `{{ .Release.Name }}`
- ❌ `{{ .Values.something }}`
- ❌ Any other `{{ }}` template syntax

### ✅ DO USE Actual Values

Always use literal values:
- ✅ `namespace: my-namespace`
- ✅ `name: my-service-monitor`
- ✅ Hard-coded strings and values

### Why?

These ServiceMonitor files are:
1. Included in `prometheus-operator/k8s/servicemonitors/kustomization.yaml`
2. Processed by Flux's `kustomize-controller`
3. Applied directly as Kubernetes manifests

Kustomize does **NOT** understand Helm template syntax. Using Helm templates will cause:
```
kustomize build failed: map[string]interface {}(nil): yaml: invalid map key: map[string]interface {}{\".Release.Namespace\":\"\"}
```

## File Naming Convention

Use descriptive names: `<component>-servicemonitor.yaml`

Examples:
- `cloudflare-tunnel-servicemonitor.yaml`
- `homepage-servicemonitor.yaml`
- `knative-lambda-operator-servicemonitor.yaml`

## Related Resources

- ServiceMonitor CRD: https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.ServiceMonitor
- Prometheus Operator: https://prometheus-operator.dev/
