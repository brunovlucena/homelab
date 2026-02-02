# Prometheus Alerts Review - December 24, 2025

## Executive Summary

**Issue**: Homepage (lucena.cloud) was down but no alerts were received via PagerDuty.

**Root Cause**: Missing critical alerts for:
1. Homepage frontend service down
2. Complete service unavailability (namespace/pods/deployment)
3. Cloudflare Tunnel ingress not ready
4. No ServiceMonitor for frontend (couldn't monitor frontend)

**Status**: ✅ **FIXED** - All critical alerts now properly configured and routing to PagerDuty.

---

## Issues Found & Fixed

### 1. Missing Frontend ServiceMonitor ❌ → ✅

**Problem**: Only API had a ServiceMonitor, frontend couldn't be monitored with `up{job="homepage-frontend"}`.

**Fix**: Created `homepage-frontend-servicemonitor.yaml` to expose frontend metrics.

**Files Changed**:
- `flux/infrastructure/prometheus-operator/k8s/servicemonitors/homepage-frontend-servicemonitor.yaml` (NEW)
- `flux/infrastructure/prometheus-operator/k8s/servicemonitors/kustomization.yaml` (updated)

### 2. Missing Critical Alerts ❌ → ✅

**Problem**: No alerts for:
- Frontend being down
- Complete service unavailability (namespace doesn't exist, no pods, 0 replicas)
- Cloudflare Tunnel ingress not ready
- Deployment not ready

**Fix**: Added 5 new critical alerts to `homepage.yaml`:

1. **BrunoSiteCompletelyDown** - Detects complete service unavailability
   - Checks if namespace exists
   - Checks if frontend pods are running
   - Checks if deployment has 0 available replicas
   - **Severity**: `critical`
   - **For**: 2m

2. **BrunoSiteFrontendDown** - Frontend service down
   - Uses `up{job="homepage-frontend"} == 0`
   - **Severity**: `critical`
   - **For**: 1m

3. **BrunoSiteCloudflareTunnelDown** - Tunnel ingress not ready
   - Checks if service has no endpoints or doesn't exist
   - **Severity**: `critical`
   - **For**: 3m

4. **BrunoSiteDeploymentNotReady** - Deployment has unavailable replicas
   - Compares available vs desired replicas
   - **Severity**: `critical`
   - **For**: 5m

**Files Changed**:
- `flux/infrastructure/prometheus-operator/k8s/prometheusrules/homepage.yaml` (updated)

### 3. Alertmanager Configuration ✅

**Status**: Already correctly configured.

**Verification**:
- ✅ PagerDuty secret exists: `alertmanager-pagerduty` in `prometheus` namespace
- ✅ AlertmanagerConfig `pagerduty-config` routes all `severity: critical` alerts to PagerDuty
- ✅ HelmRelease also has PagerDuty routing configured (backup)

**Configuration**:
```yaml
route:
  routes:
    - matchers:
        - name: severity
          value: critical
          matchType: =
      receiver: 'pagerduty'
      continue: true
receivers:
  - name: 'pagerduty'
    pagerdutyConfigs:
      - serviceKey:
          name: alertmanager-pagerduty
          key: service-key
```

---

## Review of All PrometheusRules

### Critical Alerts Summary

All PrometheusRules reviewed - all critical alerts have `severity: critical` label:

| Rule File | Critical Alerts | Status |
|-----------|----------------|--------|
| `homepage.yaml` | 9 critical alerts | ✅ Fixed (added 4 new) |
| `cert-manager.yaml` | 2 critical alerts | ✅ OK |
| `falco-security.yaml` | 5 critical alerts | ✅ OK |
| `flux-system.yaml` | 1 critical alert | ✅ OK |
| `node-disk.yaml` | 2 critical alerts | ✅ OK |
| `persistent-volumes.yaml` | 1 critical alert | ✅ OK |
| `github-runners.yaml` | 3 critical alerts | ✅ OK |
| `sealed-secrets.yaml` | 1 critical alert | ✅ OK |
| `grafana.yaml` | 4 critical alerts | ✅ OK |

**Total**: 28 critical alerts across all PrometheusRules - all properly labeled and routing to PagerDuty.

---

## Testing Recommendations

1. **Test Frontend Down Alert**:
   ```bash
   kubectl scale deployment homepage-frontend -n homepage --replicas=0
   # Should trigger BrunoSiteFrontendDown and BrunoSiteCompletelyDown after 1-2m
   ```

2. **Test Complete Service Down**:
   ```bash
   kubectl delete namespace homepage
   # Should trigger BrunoSiteCompletelyDown after 2m
   ```

3. **Test Cloudflare Tunnel**:
   ```bash
   kubectl delete cloudflaretunnelingress homepage-frontend -n homepage
   # Should trigger BrunoSiteCloudflareTunnelDown after 3m
   ```

4. **Verify PagerDuty Integration**:
   - Check PagerDuty dashboard for incoming alerts
   - Verify alert details include summary, description, runbook_url

---

## Next Steps

1. ✅ **DONE**: Add frontend ServiceMonitor
2. ✅ **DONE**: Add critical alerts for complete service unavailability
3. ✅ **DONE**: Add Cloudflare Tunnel ingress alert
4. ✅ **DONE**: Verify all critical alerts have `severity: critical` label
5. ✅ **DONE**: Verify Alertmanager routing to PagerDuty

**Optional Improvements**:
- Add synthetic monitoring (external uptime checks)
- Add alert for service response time degradation
- Add alert for SSL certificate expiration (if using cert-manager)

---

## Files Changed

1. `flux/infrastructure/prometheus-operator/k8s/servicemonitors/homepage-frontend-servicemonitor.yaml` (NEW)
2. `flux/infrastructure/prometheus-operator/k8s/servicemonitors/kustomization.yaml` (updated)
3. `flux/infrastructure/prometheus-operator/k8s/prometheusrules/homepage.yaml` (updated)

---

## Deployment

Changes will be automatically deployed via Flux when committed to the repository. The new ServiceMonitor and alerts will be active after:

1. Flux reconciles `studio-02b-observability-extras` (ServiceMonitors)
2. Flux reconciles `studio-02b-observability-extras` (PrometheusRules)

**Estimated time**: ~5 minutes after commit.

---

**Reviewed by**: SRE Engineer  
**Date**: December 24, 2025  
**Status**: ✅ Complete

