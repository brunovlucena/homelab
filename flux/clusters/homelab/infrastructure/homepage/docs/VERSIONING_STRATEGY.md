# 🏷️ Homepage Versioning & Branching Strategy

## Overview
This document outlines the versioning and branching strategy for the Homepage application (frontend and API) following SRE best practices.

## Table of Contents
- [Semantic Versioning](#semantic-versioning)
- [Branching Strategy](#branching-strategy)
- [CI/CD Pipeline](#cicd-pipeline)
- [Release Process](#release-process)
- [Rollback Strategy](#rollback-strategy)
- [Agent Management](#agent-management)

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

#### API Version
- Location: `api/VERSION`
- Current: `0.1.0` → `1.0.0` (first stable release)
- Container image: `ghcr.io/brunovlucena/homelab/homepage-api:1.0.0`

#### Frontend Version
- Location: `frontend/package.json` (version field)
- Current: `0.1.0` → `1.0.0` (first stable release)
- Container image: `ghcr.io/brunovlucena/homelab/homepage-frontend:1.0.0`

#### Helm Chart Version
- Location: `chart/Chart.yaml`
- Current: `0.1.24` → `1.0.0` (first stable release)
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
- **Deployment**: Automatically deploys to production
- **Tagging**: All releases tagged from this branch
- **Version**: Full semantic versions (e.g., `v1.0.0`, `v1.1.0`)

#### 2. `develop` - Integration Branch
- **Purpose**: Integration and staging environment
- **Protection**: CI/CD checks required
- **Deployment**: Automatically deploys to staging
- **Version**: Pre-release versions (e.g., `v1.1.0-beta.1`)

#### 3. `feature/*` - Feature Branches
- **Naming**: `feature/description` (e.g., `feature/add-user-auth`)
- **Source**: Branch from `develop`
- **Merge to**: `develop` via PR
- **Lifespan**: Short-lived (1-3 days max)
- **Version**: Development versions (e.g., `v1.1.0-dev.{sha}`)

#### 4. `bugfix/*` - Bug Fix Branches
- **Naming**: `bugfix/issue-description` (e.g., `bugfix/fix-login-error`)
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
│    Push     │ (Push to GHCR with semantic version)
└──────┬──────┘
       │
       v
┌─────────────┐
│   Deploy    │ (Update Helm chart, Flux reconcile)
└─────────────┘
```

### Automated Version Tagging

#### On `feature/*` branches:
- Tag format: `v{VERSION}-dev.{SHORT_SHA}`
- Example: `v1.1.0-dev.abc1234`
- Push images: `homepage-api:1.1.0-dev.abc1234`

#### On `develop` branch:
- Tag format: `v{VERSION}-beta.{BUILD_NUMBER}`
- Example: `v1.1.0-beta.5`
- Push images: `homepage-api:1.1.0-beta.5`

#### On `release/*` branches:
- Tag format: `v{VERSION}-rc.{BUILD_NUMBER}`
- Example: `v1.1.0-rc.1`
- Push images: `homepage-api:1.1.0-rc.1`

#### On `main` branch (after merge):
- Tag format: `v{VERSION}`
- Example: `v1.1.0`
- Push images: `homepage-api:1.1.0`
- Also push: `homepage-api:latest` (for compatibility)

### GitHub Actions Workflows

1. **`ci.yml`** - Continuous Integration
   - Runs on: All PRs, all branches
   - Jobs: Lint, test, build

2. **`cd-staging.yml`** - Continuous Deployment (Staging)
   - Runs on: Push to `develop`
   - Jobs: Build, push beta images, deploy to staging

3. **`cd-production.yml`** - Continuous Deployment (Production)
   - Runs on: Push to `main`, tags
   - Jobs: Build, push release images, deploy to production

4. **`release.yml`** - Release Management
   - Runs on: Manual trigger, release branch creation
   - Jobs: Version bump, changelog, create GitHub release

---

## 🚀 Release Process

### Standard Release (Feature)

```bash
# 1. Create feature branch from develop
git checkout develop
git pull origin develop
git checkout -b feature/my-new-feature

# 2. Make changes and commit
git add .
git commit -m "feat: add new feature"

# 3. Push and create PR to develop
git push origin feature/my-new-feature
# Create PR: feature/my-new-feature → develop

# 4. After PR approval and merge to develop
# CI/CD automatically builds and deploys to staging

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
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0

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
git commit -m "fix: critical security vulnerability"

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
git tag -a v1.0.1 -m "Hotfix v1.0.1"
git push origin v1.0.1

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
kubectl rollout history deployment/homepage-api -n homepage
kubectl rollout history deployment/homepage-frontend -n homepage

# 2. Rollback to previous version
kubectl rollout undo deployment/homepage-api -n homepage
kubectl rollout undo deployment/homepage-frontend -n homepage

# 3. Rollback to specific version
kubectl rollout undo deployment/homepage-api -n homepage --to-revision=3
```

### Helm Rollback

```bash
# 1. View release history
helm history homepage -n homepage

# 2. Rollback to previous release
helm rollback homepage -n homepage

# 3. Rollback to specific release
helm rollback homepage 5 -n homepage
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
helm upgrade homepage ./chart -n homepage \
  --set api.image.tag=1.0.1 \
  --set frontend.image.tag=1.0.1

# 2. Or update values.yaml and let Flux reconcile
```

---

## 🤖 Agent Management

### Agent Versioning Strategy

For AI agents (like Agent Bruno), we maintain separate versioning:

#### Agent Components
1. **Model Version**: Tracks the fine-tuned model version
2. **API Version**: Tracks the agent API/service version
3. **Configuration Version**: Tracks prompt templates and configs

#### Agent Branches

```
main (production agent)
  ├── agent/develop (integration testing)
  ├── agent/train/* (model training experiments)
  ├── agent/prompt/* (prompt engineering)
  └── agent/release/* (agent releases)
```

#### Agent Version Format

```
{AGENT_NAME}-v{MODEL_VERSION}.{API_VERSION}.{CONFIG_VERSION}

Example: agent-bruno-v2.1.0
  - Model version: 2 (major model update)
  - API version: 1 (API changes)
  - Config version: 0 (config updates)
```

#### Agent Release Process

```bash
# 1. Train new model
cd repos/flyte-test/_vault/sre-chatbot-finetune
make train

# 2. Evaluate model performance
make benchmark

# 3. If performance improves, create agent release branch
git checkout -b agent/release/v2.1.0

# 4. Update agent version
echo "2.1.0" > ../agent-bruno/VERSION

# 5. Export model to Ollama format
make export-ollama

# 6. Deploy to staging
make deploy-staging

# 7. Test agent in staging
make test-agent

# 8. Create PR to main
# After approval, deploy to production
make deploy-production

# 9. Tag agent release
git tag -a agent-bruno-v2.1.0 -m "Agent Bruno v2.1.0"
git push origin agent-bruno-v2.1.0
```

### Agent A/B Testing

For testing new agent versions:

```yaml
# Canary deployment (10% traffic to new version)
apiVersion: v1
kind: Service
metadata:
  name: agent-bruno
spec:
  selector:
    app: agent-bruno
    version: v2.1.0  # 10% weight
  selector:
    app: agent-bruno
    version: v2.0.0  # 90% weight
```

---

## 📊 Monitoring & Observability

### Version Tracking

All deployments include version labels:

```yaml
labels:
  app.kubernetes.io/version: "1.0.0"
  app.kubernetes.io/component: "api"
  git.commit.sha: "abc1234"
  git.branch: "main"
```

### Prometheus Metrics

```promql
# Track version deployments
up{job="homepage-api", version="1.0.0"}

# Track rollbacks
rate(kube_deployment_status_replicas_unavailable[5m]) > 0
```

### Alerts

```yaml
# Version mismatch alert
- alert: VersionMismatch
  expr: count(up{job="homepage-api"} by (version)) > 1
  annotations:
    summary: "Multiple versions of homepage-api are running"
```

---

## 📝 Changelog Management

We maintain a `CHANGELOG.md` file following [Keep a Changelog](https://keepachangelog.com/).

### Format

```markdown
# Changelog

## [Unreleased]

## [1.1.0] - 2025-10-20
### Added
- New user authentication feature
- Redis caching for API responses

### Changed
- Updated API response format for projects endpoint

### Fixed
- Fixed memory leak in frontend chatbot

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
make deploy ENV=production VERSION=1.1.0
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
make build VERSION=1.1.0

# Deploy to staging
make deploy ENV=staging VERSION=1.1.0-beta.1

# Deploy to production
make deploy ENV=production VERSION=1.1.0

# Rollback
make rollback ENV=production VERSION=1.0.9

# View current version
make version
```

### Version Decision Tree

```
Is it a bug fix? → Patch (1.0.1)
Is it a new feature? → Minor (1.1.0)
Is it a breaking change? → Major (2.0.0)
Is it urgent? → Hotfix branch from main
Is it experimental? → Feature branch with dev tag
```

