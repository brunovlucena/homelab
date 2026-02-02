# PrometheusRule Resources

This directory contains PrometheusRule custom resources for alerting rules.

## ⚠️ CRITICAL: Kustomize vs Helm Template Syntax

**These files are processed by KUSTOMIZE, NOT HELM!**

### ❌ DO NOT USE Helm Template Syntax

**NEVER** use Helm template syntax in PrometheusRule files:
- ❌ `{{ .Release.Namespace }}`
- ❌ `{{ .Release.Name }}`
- ❌ `{{ .Values.something }}`
- ❌ Any other `{{ }}` template syntax

### ✅ DO USE Actual Values

Always use literal values:
- ✅ `namespace: prometheus`
- ✅ `app.kubernetes.io/instance: cert-manager-alerts`
- ✅ Hard-coded strings and values

### Why?

These PrometheusRule files are:
1. Included in `prometheus-operator/kustomization.yaml`
2. Processed by Flux's `kustomize-controller`
3. Applied directly as Kubernetes manifests

Kustomize does **NOT** understand Helm template syntax. Using Helm templates will cause:
```
kustomize build failed: map[string]interface {}(nil): yaml: invalid map key: map[string]interface {}{".Release.Namespace":""}
```

### Example

**❌ WRONG:**
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-rules
  namespace: {{ .Release.Namespace }}  # ❌ This will fail!
```

**✅ CORRECT:**
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-rules
  namespace: prometheus  # ✅ Use actual value
```

## File Structure

All PrometheusRule files should:
- Use `namespace: prometheus` (unless targeting a different namespace)
- Include appropriate labels for Prometheus discovery
- Follow the alerting rule structure defined in Prometheus documentation

## Alert Rules Overview

| File | Description | Severity |
|------|-------------|----------|
| `cert-manager.yaml` | Certificate expiry and renewal alerts | warning/critical |
| `cloudflare-argo.yaml` | Argo Smart Routing usage & cost alerts (BVL-23) | info/warning |
| `falco-security.yaml` | Runtime security alerts from Falco | warning/critical |
| `flux-system.yaml` | GitOps reconciliation alerts | warning/critical |
| `github-runners.yaml` | Self-hosted runner health alerts | warning |
| `grafana.yaml` | Grafana service health alerts | warning/critical |
| `homepage.yaml` | Homepage service alerts | warning/critical |
| `homepage-security.yaml` | Security monitoring alerts (SEC-013) | warning/critical |
| `node-disk.yaml` | Node disk space alerts | warning/critical |
| `persistent-volumes.yaml` | PV/PVC usage alerts | warning/critical |
| `sealed-secrets.yaml` | Sealed Secrets controller alerts | warning |
| `agent-sre-triggers.yaml` | SRE automation triggers | info/warning |
| `grafana-service-accounts.yaml` | Grafana SA expiry alerts | warning |

### Cloudflare Argo Smart Routing Alerts (BVL-23)

The `cloudflare-argo.yaml` file monitors Argo Smart Routing bandwidth usage:

**Free Tier Alerts:**
- `CloudflareArgoFreeTierWarning` - 80% of 1GB free tier consumed (~800MB)
- `CloudflareArgoFreeTierExceeded` - Free tier exceeded, now paying $5/10GB

**Cost Tracking Alerts (per 10GB):**
- `CloudflareArgo10GBConsumed` - $5 cost
- `CloudflareArgo20GBConsumed` - $10 cost
- `CloudflareArgo30GBConsumed` - $15 cost
- `CloudflareArgo50GBConsumed` - $25 cost
- `CloudflareArgo100GBConsumed` - $50 cost

**Performance Alerts:**
- `CloudflareArgoUnusualBandwidthSpike` - >3x normal daily bandwidth
- `CloudflareArgoMetricsMissing` - Cloudflare metrics not being collected

## Related Resources

- PrometheusRule CRD: https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.PrometheusRule
- Alerting Rules: https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/
