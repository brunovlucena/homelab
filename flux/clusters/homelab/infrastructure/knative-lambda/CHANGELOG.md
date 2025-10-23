# Changelog

All notable changes to the Knative Lambda Builder project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-10-23
### Added
- Initial stable release with semantic versioning
- Comprehensive versioning strategy documentation
- Version management script (`scripts/version-manager.sh`)
- Makefile targets for version management
  - `make version-info` - Show current version information
  - `make version-bump VERSION_NEW=X.Y.Z` - Bump version to specific version
  - `make version-bump-major` - Auto-bump major version
  - `make version-bump-minor` - Auto-bump minor version
  - `make version-bump-patch` - Auto-bump patch version
  - `make release-notes` - Generate release notes from git commits
- VERSION files for all components (builder, sidecar, metrics-pusher)
- Branching guide and quick start documentation
- CHANGELOG.md for tracking release history

### Changed
- Updated Chart.yaml version from 0.1.0 to 1.0.0
- Updated all image tags from `:latest` to semantic versions (`:1.0.0`)
- Configured automatic version tagging based on git branch:
  - `main` branch: `v1.0.0` (production)
  - `develop` branch: `v1.0.0-beta.N` (staging)
  - `feature/*` branches: `v1.0.0-dev.{sha}` (development)
- Synchronized versioning across all three components (builder, sidecar, metrics-pusher)

### Fixed
- Removed hardcoded `:latest` tags in favor of semantic versions
- Improved build reproducibility with explicit version tags

### Documentation
- Added `docs/VERSIONING_STRATEGY.md` - Complete versioning strategy
- Added `docs/QUICK_START_VERSIONING.md` - Quick start guide for versioning
- Added `docs/BRANCHING_GUIDE.md` - Branching and PR guidelines

## [0.1.0] - 2025-10-15
### Added
- Initial development release
- Knative Lambda Builder service
- Build monitor sidecar
- Metrics pusher component
- Helm chart for deployment
- RabbitMQ event integration
- Kaniko-based containerization
- CloudEvents support
- Job management with cleanup
- Rate limiting and resilience features
- Comprehensive observability with Prometheus metrics
- Security enhancements and validation
- Multi-environment support (dev, prd)

[Unreleased]: https://github.com/notifi-network/infra/compare/knative-lambda-v1.0.0...HEAD
[1.0.0]: https://github.com/notifi-network/infra/compare/knative-lambda-v0.1.0...knative-lambda-v1.0.0
[0.1.0]: https://github.com/notifi-network/infra/releases/tag/knative-lambda-v0.1.0

