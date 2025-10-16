# 🚀 Release v0.1.25 - Dashboard Integration

## Release Date
October 16, 2025

## Summary
This release integrates the Grafana Golden Signals dashboard into the homepage Helm chart, enabling it to follow the same versioning and branching strategy as the main application.

## 🎯 What Changed

### Dashboard Integration
- **Moved**: Dashboard from `infrastructure/dashboards/homepage.yaml` → `chart/templates/monitoring/dashboard.yaml`
- **Templated**: Converted static YAML to Helm template with dynamic configuration
- **Versioned**: Dashboard now includes version annotations from Chart.yaml
- **Configurable**: Added `monitoring.dashboard` section in values.yaml

### Version Updates
- Chart version: 0.1.24 → **0.1.25**
- AppVersion: 1.0.0 → **1.0.0**

## 📋 Changes Made

### 1. Dashboard Template (`chart/templates/monitoring/dashboard.yaml`)
```yaml
{{- if .Values.monitoring.dashboard.enabled }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "homepage.fullname" . }}-golden-signals-dashboard
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "homepage.labels" . | nindent 4 }}
    app.kubernetes.io/component: monitoring
    grafana_dashboard: "1"
  annotations:
    app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
    helm.sh/chart: {{ include "homepage.chart" . }}
```

### 2. Values Configuration (`chart/values.yaml`)
```yaml
monitoring:
  # Grafana Dashboard Configuration
  dashboard:
    enabled: true
    namespace: homepage
    labels:
      grafana_dashboard: "1"
```

### 3. Updated Files
- ✅ `chart/Chart.yaml` - Version bumped to 0.1.25
- ✅ `chart/values.yaml` - Added dashboard configuration
- ✅ `chart/templates/monitoring/dashboard.yaml` - New Helm template
- ✅ `docs/CHANGELOG.md` - Added release notes
- ✅ `api/VERSION` - Created version file
- ✅ `frontend/package.json` - Updated version
- 🗑️ `infrastructure/dashboards/homepage.yaml` - Removed (now in Helm chart)

## 🎨 Benefits

### For Versioning
- ✅ Dashboard now follows same semantic versioning as application
- ✅ Dashboard version tracked in Git with application code
- ✅ Single source of truth for all homepage components

### For Deployment
- ✅ Dashboard deployed automatically with Helm chart
- ✅ Dashboard configuration managed through values.yaml
- ✅ Dashboard can be enabled/disabled per environment
- ✅ Dashboard metadata includes proper version annotations

### For Branching
- ✅ Dashboard changes go through same PR process
- ✅ Dashboard tested in staging before production
- ✅ Dashboard versioned with each release

## 🔄 Migration Impact

### Before (Old Structure)
```
infrastructure/
├── dashboards/
│   ├── homepage.yaml          ❌ Separate versioning
│   └── kustomization.yaml
└── homepage/
    └── chart/                 ❌ Dashboard not included
```

### After (New Structure)
```
infrastructure/
└── homepage/
    └── chart/
        └── templates/
            └── monitoring/
                └── dashboard.yaml  ✅ Versioned together
```

## 📦 Deployment Notes

### Automatic Deployment
The dashboard will be deployed automatically when the homepage Helm chart is deployed or upgraded.

### Manual Dashboard Disable
If you need to disable the dashboard in a specific environment:

```yaml
# values-production.yaml
monitoring:
  dashboard:
    enabled: false
```

### Verify Dashboard
After deployment, check that the dashboard ConfigMap was created:

```bash
kubectl get configmap -n homepage -l app.kubernetes.io/component=monitoring
kubectl get configmap -n homepage -l grafana_dashboard=1
```

## 🚀 Next Steps

### 1. Review Changes
```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab
git diff
```

### 2. Stage and Commit
```bash
git add .
git commit -m "feat: integrate dashboard into homepage Helm chart

- Move dashboard from infrastructure/dashboards to chart/templates/monitoring
- Add Helm templating for dynamic configuration
- Add version annotations to dashboard metadata
- Add monitoring.dashboard configuration in values.yaml
- Update CHANGELOG for v0.1.25

BREAKING CHANGE: Dashboard moved from separate kustomization to Helm chart"
```

### 3. Tag Release
```bash
git tag -a homepage-v0.1.25 -m "Release v0.1.25 - Dashboard Integration"
```

### 4. Push to Remote
```bash
# Push commits and tags
git push origin main
git push origin homepage-v0.1.25
```

### 5. Verify Deployment
```bash
# Force Flux reconciliation
flux reconcile source git homelab -n flux-system
flux reconcile helmrelease homepage -n homepage

# Check dashboard deployment
kubectl get configmap -n homepage | grep dashboard
kubectl describe configmap -n homepage -l grafana_dashboard=1
```

## 🔍 Testing Checklist

- [ ] Dashboard ConfigMap created in correct namespace
- [ ] Dashboard has proper version annotations
- [ ] Dashboard visible in Grafana
- [ ] Dashboard panels loading data correctly
- [ ] Dashboard can be disabled via values.yaml
- [ ] Old dashboard location removed from cluster

## 📚 Documentation Updates

- ✅ CHANGELOG.md updated with v0.1.25 changes
- ✅ This release document created
- ✅ Version bumped in all required files

## 🔗 Related Files

- `chart/Chart.yaml` - Chart version and app version
- `chart/values.yaml` - Dashboard configuration
- `chart/templates/monitoring/dashboard.yaml` - Dashboard template
- `docs/CHANGELOG.md` - Change history
- `api/VERSION` - API version file
- `frontend/package.json` - Frontend version

## 🎉 Summary

This release successfully integrates the Grafana dashboard into the homepage Helm chart, enabling unified versioning and deployment. The dashboard now follows the same branching strategy and release process as the rest of the application.

**Key Achievement**: Single versioned artifact containing application + monitoring configuration! 🎯

