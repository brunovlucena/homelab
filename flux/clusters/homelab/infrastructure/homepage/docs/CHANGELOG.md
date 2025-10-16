# Changelog

All notable changes to the Homepage application will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.25] - 2025-10-16

### Added
- Integrated Grafana Golden Signals dashboard into Helm chart
- Dashboard now follows the same versioning/branching scheme as the application
- Added `monitoring.dashboard` configuration in values.yaml
- Dashboard is now version-controlled with the homepage application

### Changed
- Moved dashboard from `infrastructure/dashboards/homepage.yaml` to `chart/templates/monitoring/dashboard.yaml`
- Dashboard now uses Helm templating for dynamic configuration
- Dashboard metadata now includes version annotations

### Infrastructure
- Dashboard ConfigMap name now follows Helm naming conventions
- Dashboard namespace is now dynamic based on Helm release namespace
- Added conditional rendering for dashboard (can be disabled via values.yaml)

## [0.1.24] - 2025-10-16

### Added
- Initial versioned release
- API with health checks and metrics
- Frontend with React and TypeScript
- Helm chart for Kubernetes deployment
- Integration with Agent Bruno (chatbot)
- Redis caching
- PostgreSQL database
- MinIO storage
- OpenTelemetry instrumentation

### Infrastructure
- Flux GitOps deployment
- Prometheus monitoring
- Grafana dashboards
- ServiceMonitor for metrics collection

---

## Version History Format

### Added
- New features

### Changed
- Changes in existing functionality

### Deprecated
- Soon-to-be removed features

### Removed
- Removed features

### Fixed
- Bug fixes

### Security
- Security updates

---

## Upgrade Guide

### From 0.x to 1.0.0

```bash
# 1. Update Helm values
helm upgrade homepage ./chart \
  --set api.image.tag=1.0.0 \
  --set frontend.image.tag=1.0.0 \
  --namespace homepage

# 2. Or trigger Flux reconciliation
flux reconcile helmrelease homepage -n homepage

# 3. Verify deployment
kubectl rollout status deployment/homepage-api -n homepage
kubectl rollout status deployment/homepage-frontend -n homepage
```

---

## Links

- [GitHub Repository](https://github.com/brunovlucena/homelab)
- [Documentation](./docs/VERSIONING_STRATEGY.md)
- [Release Notes](https://github.com/brunovlucena/homelab/releases)

