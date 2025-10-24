# 🏷️ Knative Lambda Versioning & Branching Strategy

## Overview
This document outlines the versioning and branching strategy for the Knative Lambda Builder service following SRE best practices.

## Table of Contents
- [Semantic Versioning](#semantic-versioning)
- [Branching Strategy](#branching-strategy)
- [CI/CD Pipeline](#cicd-pipeline)
- [Release Process](#release-process)
- [Rollback Strategy](#rollback-strategy)
- [Multi-Component Versioning](#multi-component-versioning)

---

## 🔢 Semantic Versioning

We follow **Semantic Versioning 2.0.0** (https://semver.org/) for all components.

### Version Format
```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

- **MAJOR**: Breaking changes or incompatible API changes
- **MINOR**: New features (backwards-compatible)
- **PATCH**: Bug fixes (backwards-compatible)
- **PRERELEASE**: Optional (alpha, beta, rc.1, etc.)
- **BUILD**: Optional build metadata

### Examples
- `1.0.0` - Production release
- `1.1.0` - New feature added
- `1.1.1` - Bug fix
- `2.0.0` - Breaking change
- `1.2.0-beta.1` - Beta release
- `1.2.0-rc.1` - Release candidate

### Component Versioning

#### Service (Main Builder Service)
- Location: `VERSION`
- Container image: `339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-knative-lambda-builder:1.0.0`

#### Sidecar (Build Monitor Sidecar)
- Location: `sidecar/VERSION`
- Container image: `339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-knative-lambda-sidecar:1.0.0`

#### Metrics Pusher
- Location: `metrics-pusher/VERSION`
- Container image: `339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-knative-lambda-metrics-pusher:1.0.0`

#### Helm Chart Version
- Location: `deploy/Chart.yaml`
- Current: `0.1.0` → `1.0.0` (first stable release)
- Follows application version

---

## 🌿 Branching Strategy

We use **Trunk-Based Development** with short-lived feature branches.

### Branch Structure

```
main (production-ready)
  ├── develop (integration/staging)
  ├── feature/* (short-lived feature branches)
  ├── bugfix/* (bug fixes)
  ├── hotfix/* (urgent production fixes)
  └── release/* (release preparation)
```

### Branch Types

#### 1. `main` - Production Branch
- **Purpose**: Production-ready code
- **Protection**: Required PR reviews, CI/CD checks
- **Deployment**: Automatically deploys to production (prd environment)
- **Tagging**: All releases tagged from this branch
- **Version**: Full semantic versions (e.g., `v1.0.0`, `v1.1.0`)

#### 2. `develop` - Integration Branch
- **Purpose**: Integration and staging environment
- **Protection**: CI/CD checks required
- **Deployment**: Automatically deploys to dev environment
- **Version**: Pre-release versions (e.g., `v1.1.0-beta.1`)

#### 3. `feature/*` - Feature Branches
- **Naming**: `feature/description` (e.g., `feature/add-kaniko-cache`)
- **Source**: Branch from `develop`
- **Merge to**: `develop` via PR
- **Lifespan**: Short-lived (1-3 days max)
- **Version**: Development versions (e.g., `v1.1.0-dev.{sha}`)

#### 4. `bugfix/*` - Bug Fix Branches
- **Naming**: `bugfix/issue-description` (e.g., `bugfix/fix-job-timeout`)
- **Source**: Branch from `develop`
- **Merge to**: `develop` via PR
- **Lifespan**: Short-lived

#### 5. `hotfix/*` - Hotfix Branches
- **Naming**: `hotfix/critical-issue` (e.g., `hotfix/security-patch`)
- **Source**: Branch from `main`
- **Merge to**: `main` AND `develop` via PR
- **Version**: Patch version bump (e.g., `v1.0.0` → `v1.0.1`)
- **Urgency**: Critical production issues only

#### 6. `release/*` - Release Branches
- **Naming**: `release/v1.1.0`
- **Source**: Branch from `develop`
- **Purpose**: Final testing, version bumps, changelog updates
- **Merge to**: `main` (then tag) and back-merge to `develop`
- **Lifespan**: Short (1-2 days for final testing)

### Branch Protection Rules

#### `main` Branch
- ✅ Require pull request reviews (2 approvers)
- ✅ Require status checks to pass
- ✅ Require branches to be up to date
- ✅ Require signed commits
- ✅ Include administrators
- ✅ Restrict force pushes
- ✅ Restrict deletions

#### `develop` Branch
- ✅ Require pull request reviews (1 approver)
- ✅ Require status checks to pass
- ✅ Restrict force pushes

---

## 🔄 CI/CD Pipeline

### Pipeline Stages

```
┌─────────────┐
│   Trigger   │ (Push/PR)
└──────┬──────┘
       │
       v
┌─────────────┐
│    Test     │ (Lint, Unit, Integration)
└──────┬──────┘
       │
       v
┌─────────────┐
│    Build    │ (Docker images with version tags)
└──────┬──────┘
       │
       v
┌─────────────┐
│    Push     │ (Push to ECR with semantic version)
└──────┬──────┘
       │
       v
┌─────────────┐
│   Deploy    │ (Update Helm chart, ArgoCD/Flux reconcile)
└─────────────┘
```

### Automated Version Tagging

#### On `feature/*` branches:
- Tag format: `v{VERSION}-dev.{SHORT_SHA}`
- Example: `v1.1.0-dev.abc1234`
- Push images: 
  - `knative-lambda-builder:1.1.0-dev.abc1234`
  - `knative-lambda-sidecar:1.1.0-dev.abc1234`
  - `knative-lambda-metrics-pusher:1.1.0-dev.abc1234`

#### On `develop` branch:
- Tag format: `v{VERSION}-beta.{BUILD_NUMBER}`
- Example: `v1.1.0-beta.5`
- Push images: 
  - `knative-lambda-builder:1.1.0-beta.5`
  - `knative-lambda-sidecar:1.1.0-beta.5`
  - `knative-lambda-metrics-pusher:1.1.0-beta.5`

#### On `release/*` branches:
- Tag format: `v{VERSION}-rc.{BUILD_NUMBER}`
- Example: `v1.1.0-rc.1`
- Push images: 
  - `knative-lambda-builder:1.1.0-rc.1`
  - `knative-lambda-sidecar:1.1.0-rc.1`
  - `knative-lambda-metrics-pusher:1.1.0-rc.1`

#### On `main` branch (after merge):
- Tag format: `v{VERSION}`
- Example: `v1.1.0`
- Push images: 
  - `knative-lambda-builder:1.1.0`
  - `knative-lambda-sidecar:1.1.0`
  - `knative-lambda-metrics-pusher:1.1.0`
- Also push: `:prd` and `:latest` tags (for compatibility)

---

## 🚀 Release Process

### Standard Release (Feature)

```bash
# 1. Create feature branch from develop
git checkout develop
git pull origin develop
git checkout -b feature/add-build-cache

# 2. Make changes and commit
git add .
git commit -m "feat: add kaniko build caching"

# 3. Push and create PR to develop
git push origin feature/add-build-cache
# Create PR: feature/add-build-cache → develop

# 4. After PR approval and merge to develop
# CI/CD automatically builds and deploys to dev environment

# 5. When ready for production, create release branch
git checkout develop
git pull origin develop
git checkout -b release/v1.1.0

# 6. Bump version in all locations
make version-bump VERSION=1.1.0

# 7. Update CHANGELOG.md
# Add release notes

# 8. Commit and push release branch
git add .
git commit -m "chore: bump version to v1.1.0"
git push origin release/v1.1.0

# 9. Create PR: release/v1.1.0 → main
# After approval and merge, CI/CD deploys to production

# 10. Tag the release
git checkout main
git pull origin main
git tag -a knative-lambda-v1.1.0 -m "Release v1.1.0"
git push origin knative-lambda-v1.1.0

# 11. Back-merge to develop
git checkout develop
git merge main
git push origin develop
```

### Hotfix Release (Urgent)

```bash
# 1. Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b hotfix/critical-bug

# 2. Fix the issue
git add .
git commit -m "fix: critical job cleanup issue"

# 3. Bump patch version
make version-bump VERSION=1.0.1

# 4. Push and create PR to main
git push origin hotfix/critical-bug
# Create PR: hotfix/critical-bug → main

# 5. After approval and merge
# CI/CD automatically deploys to production

# 6. Tag the hotfix
git checkout main
git pull origin main
git tag -a knative-lambda-v1.0.1 -m "Hotfix v1.0.1"
git push origin knative-lambda-v1.0.1

# 7. Back-merge to develop
git checkout develop
git merge main
git push origin develop
```

---

## ⏮️ Rollback Strategy

### Kubernetes Rollback

```bash
# 1. View deployment history
kubectl rollout history deployment/knative-lambda-builder -n knative-lambda-prd

# 2. Rollback to previous version
kubectl rollout undo deployment/knative-lambda-builder -n knative-lambda-prd

# 3. Rollback to specific version
kubectl rollout undo deployment/knative-lambda-builder -n knative-lambda-prd --to-revision=3
```

### Helm Rollback

```bash
# 1. View release history
helm history knative-lambda -n knative-lambda-prd

# 2. Rollback to previous release
helm rollback knative-lambda -n knative-lambda-prd

# 3. Rollback to specific release
helm rollback knative-lambda 5 -n knative-lambda-prd
```

### Git Revert

```bash
# Revert a commit (creates new commit)
git revert <commit-hash>
git push origin main
```

### Image Tag Rollback

```bash
# 1. Update Helm values to previous version
helm upgrade knative-lambda ./deploy -n knative-lambda-prd \
  --set builder.image.tag=1.0.1 \
  --set sidecar.image.tag=1.0.1 \
  --set metricsPusher.image.tag=1.0.1

# 2. Or update values.yaml and let ArgoCD/Flux reconcile
```

---

## 🏗️ Multi-Component Versioning

Knative Lambda consists of three main components that are versioned together:

### Version Synchronization

All components share the same version number for consistency:

```yaml
# deploy/Chart.yaml
version: 1.0.0
appVersion: "1.0.0"

# deploy/values.yaml
builder:
  image:
    tag: "1.0.0"

sidecar:
  image:
    tag: "1.0.0"

metricsPusher:
  image:
    tag: "1.0.0"
```

### Independent Component Versions (Advanced)

For advanced scenarios, components can have independent versions:

```yaml
# deploy/values.yaml
builder:
  image:
    tag: "1.2.0"  # Major feature update

sidecar:
  image:
    tag: "1.1.1"  # Bug fix only

metricsPusher:
  image:
    tag: "1.0.0"  # No changes
```

**Note**: Use synchronized versioning by default for simplicity.

---

## 📊 Monitoring & Observability

### Version Tracking

All deployments include version labels:

```yaml
labels:
  app.kubernetes.io/version: "1.0.0"
  app.kubernetes.io/component: "builder"
  git.commit.sha: "abc1234"
  git.branch: "main"
```

### Prometheus Metrics

```promql
# Track version deployments
up{job="knative-lambda-builder", version="1.0.0"}

# Track rollbacks
rate(kube_deployment_status_replicas_unavailable[5m]) > 0
```

### Alerts

```yaml
# Version mismatch alert
- alert: VersionMismatch
  expr: count(up{job="knative-lambda-builder"} by (version)) > 1
  annotations:
    summary: "Multiple versions of knative-lambda-builder are running"
```

---

## 📝 Changelog Management

We maintain a `CHANGELOG.md` file following [Keep a Changelog](https://keepachangelog.com/).

### Format

```markdown
# Changelog

## [Unreleased]

## [1.1.0] - 2025-10-23
### Added
- Kaniko build caching for faster builds
- Support for ARM64 architecture

### Changed
- Updated job cleanup logic to use batch operations

### Fixed
- Fixed race condition in service creation

## [1.0.0] - 2025-10-01
### Added
- Initial stable release
```

---

## 🔧 Tools & Scripts

### Version Bump Script
```bash
make version-bump VERSION=1.1.0
```

### Release Script
```bash
make release VERSION=1.1.0
```

### Deploy Script
```bash
make deploy ENV=prd VERSION=1.1.0
```

---

## 📚 References

- [Semantic Versioning 2.0.0](https://semver.org/)
- [Trunk-Based Development](https://trunkbaseddevelopment.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [GitFlow](https://nvie.com/posts/a-successful-git-branching-model/)
- [SRE Book - Release Engineering](https://sre.google/sre-book/release-engineering/)

---

## 🎯 Quick Reference

### Common Commands

```bash
# Create feature branch
git checkout -b feature/my-feature develop

# Bump version
make version-bump VERSION=1.1.0

# Build with version
make build-and-push-all-prd VERSION=1.1.0

# Deploy to dev
make deploy ENV=dev VERSION=1.1.0-beta.1

# Deploy to production
make deploy ENV=prd VERSION=1.1.0

# View current version
make version-info
```

### Version Decision Tree

```
Is it a bug fix? → Patch (1.0.1)
Is it a new feature? → Minor (1.1.0)
Is it a breaking change? → Major (2.0.0)
Is it urgent? → Hotfix branch from main
Is it experimental? → Feature branch with dev tag
```

### Environment-Specific Tags

```
Development (dev):    v1.1.0-beta.N
Production (prd):     v1.1.0
Feature branches:     v1.1.0-dev.{sha}
Release candidates:   v1.1.0-rc.N
```

---

## 🚨 Important Notes

1. **Never use `:latest` in production** - Always use semantic versions
2. **All three components must be deployed together** - Use the same version tag
3. **Test in dev environment first** - Never deploy directly to production
4. **Always back-merge hotfixes** - Ensure develop has all production fixes
5. **Use conventional commits** - Helps with automated changelog generation
6. **Tag all production releases** - Format: `knative-lambda-v{VERSION}`


