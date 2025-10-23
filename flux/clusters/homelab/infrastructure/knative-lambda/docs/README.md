# 📚 Knative Lambda Documentation

Welcome to the Knative Lambda Builder documentation! This directory contains comprehensive guides and documentation for the project.

## 📖 Documentation Index

### 🏷️ Versioning & Release Management
- **[VERSIONING_STRATEGY.md](./VERSIONING_STRATEGY.md)** - Complete versioning strategy including semantic versioning, branching, CI/CD, and rollback procedures
- **[QUICK_START_VERSIONING.md](./QUICK_START_VERSIONING.md)** - Quick start guide with common commands and daily workflows
- **[BRANCHING_GUIDE.md](./BRANCHING_GUIDE.md)** - Detailed branching strategies, PR guidelines, and commit conventions

### 🏗️ System Documentation  
- **[JOB_START_EVENTS.md](./JOB_START_EVENTS.md)** - Documentation about job start events and event handling

## 🚀 Quick Links

### For New Contributors
1. Read [QUICK_START_VERSIONING.md](./QUICK_START_VERSIONING.md) for setup and daily workflows
2. Review [BRANCHING_GUIDE.md](./BRANCHING_GUIDE.md) for branch naming and PR guidelines
3. Check [../CHANGELOG.md](../CHANGELOG.md) for recent changes

### For DevOps Engineers
1. Read [VERSIONING_STRATEGY.md](./VERSIONING_STRATEGY.md) for full versioning details
2. Understand CI/CD pipelines and deployment strategies
3. Review rollback procedures

### For Release Managers
1. Follow release process in [VERSIONING_STRATEGY.md](./VERSIONING_STRATEGY.md)
2. Use `make version-bump VERSION_NEW=X.Y.Z` for version management
3. Update [../CHANGELOG.md](../CHANGELOG.md) with release notes

## 🔧 Common Commands

```bash
# Show current version
make version-info

# Bump version
make version-bump VERSION_NEW=1.0.0

# Auto-bump versions
make version-bump-patch  # 1.0.0 → 1.0.1
make version-bump-minor  # 1.0.0 → 1.1.0
make version-bump-major  # 1.0.0 → 2.0.0

# Generate release notes
make release-notes

# Build and push images
make build-and-push-all-dev   # Dev environment
make build-and-push-all-prd   # Production environment
```

## 📝 Version Format

We follow [Semantic Versioning 2.0.0](https://semver.org/):

```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]

Examples:
- 1.0.0         Production release
- 1.1.0         New features
- 1.0.1         Bug fix
- 2.0.0         Breaking changes
- 1.2.0-beta.1  Beta release
- 1.2.0-dev.a1b2c3d  Development build
```

## 🌿 Branch Strategy

```
main                    Production (v1.0.0)
  ├── develop           Staging (v1.1.0-beta.N)
  ├── feature/*         Development (v1.1.0-dev.{sha})
  ├── bugfix/*          Bug fixes
  ├── hotfix/*          Critical fixes
  └── release/*         Release preparation
```

## 🏗️ Components

The Knative Lambda system consists of three main components, all sharing the same version:

1. **Builder Service** - Main lambda builder service
   - Image: `knative-lambda-builder:1.0.0`
   - VERSION file: `VERSION`

2. **Sidecar** - Build monitoring sidecar
   - Image: `knative-lambda-sidecar:1.0.0`
   - VERSION file: `sidecar/VERSION`

3. **Metrics Pusher** - Metrics collection and pushing
   - Image: `knative-lambda-metrics-pusher:1.0.0`
   - VERSION file: `metrics-pusher/VERSION`

## 📊 Version Tracking

All versions are synchronized and managed via:
- `deploy/Chart.yaml` - Helm chart version
- `deploy/values.yaml` - Image tags
- `VERSION` files - Component-specific versions
- `CHANGELOG.md` - Release history

## 🎯 Quick Workflows

### Feature Development
```bash
git checkout -b feature/my-feature develop
# ... make changes ...
git commit -m "feat: add my feature"
git push origin feature/my-feature
# Create PR to develop
```

### Release to Production
```bash
git checkout -b release/v1.1.0 develop
make version-bump VERSION_NEW=1.1.0
# Update CHANGELOG.md
git commit -m "chore: release v1.1.0"
git push origin release/v1.1.0
# Create PR to main
```

### Emergency Hotfix
```bash
git checkout -b hotfix/critical-fix main
# ... fix issue ...
make version-bump VERSION_NEW=1.0.1
git commit -m "fix: critical issue"
git push origin hotfix/critical-fix
# Create PR to main AND develop
```

## 🔍 Additional Resources

- [Semantic Versioning 2.0.0](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Trunk-Based Development](https://trunkbaseddevelopment.com/)

## 📞 Support

For questions or issues:
1. Check the relevant documentation above
2. Review [../CHANGELOG.md](../CHANGELOG.md) for known issues
3. Contact the Notifi Infrastructure Team

---

**Last Updated**: 2025-10-23
**Documentation Version**: 1.0.0


