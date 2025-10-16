# 🚀 Quick Start: Versioning & Release Management

## TL;DR - Most Common Commands

```bash
# 1. Create a feature
git checkout -b feature/my-feature develop
# ... make changes ...
git commit -m "feat: add my feature"
git push origin feature/my-feature
# Create PR to develop

# 2. Release to production
git checkout -b release/v1.1.0 develop
make version-bump VERSION=1.1.0
git commit -m "chore: bump version to v1.1.0"
git push origin release/v1.1.0
# Create PR to main

# 3. Hotfix production issue
git checkout -b hotfix/critical-fix main
# ... fix issue ...
git commit -m "fix: resolve critical issue"
make version-bump VERSION=1.0.1
git push origin hotfix/critical-fix
# Create PR to main AND develop
```

---

## First Time Setup

### 1. Clone and Configure
```bash
# Clone repository
git clone https://github.com/brunovlucena/homelab.git
cd homelab/flux/clusters/homelab/infrastructure/homepage

# Make scripts executable
chmod +x scripts/*.sh
```

### 2. Install Dependencies
```bash
# API (Go)
cd api && go mod download && cd ..

# Frontend (Node.js)
cd frontend && npm install --legacy-peer-deps && cd ..
```

### 3. Set Up Local Development
```bash
# Start all services
make up

# Access locally
# Frontend: http://localhost:3000
# API: http://localhost:8080
```

---

## Daily Workflows

### 📝 Working on a Feature

```bash
# 1. Start from develop
git checkout develop
git pull origin develop

# 2. Create feature branch
git checkout -b feature/add-dark-mode

# 3. Develop and test locally
make up
make test

# 4. Commit with conventional commits
git add .
git commit -m "feat(frontend): add dark mode toggle"

# 5. Push and create PR
git push origin feature/add-dark-mode
# Go to GitHub → Create PR to develop

# 6. After PR approval
# ✅ Automatically deployed to staging
# ✅ Images: homepage-api:v0.1.24-beta.123
```

### 🔄 Testing in Staging

```bash
# View staging deployment
kubectl get pods -n homepage

# Check logs
kubectl logs -f deployment/homepage-api -n homepage
kubectl logs -f deployment/homepage-frontend -n homepage

# Test endpoints
curl https://staging.lucena.cloud/health
curl https://staging.lucena.cloud/api/v1/projects
```

### 🚀 Releasing to Production

```bash
# 1. Create release branch
git checkout develop
git pull origin develop
git checkout -b release/v1.1.0

# 2. Bump version (auto-updates all files)
make version-bump VERSION=1.1.0

# 3. Update CHANGELOG.md
# Add release notes

# 4. Commit and push
git add .
git commit -m "chore: release v1.1.0"
git push origin release/v1.1.0

# 5. Create PR to main
# Go to GitHub → Create PR: release/v1.1.0 → main
# ✅ Requires 2 approvals

# 6. After merge to main
# ✅ Automatically deploys to production
# ✅ Creates GitHub release
# ✅ Tags: homepage-v1.1.0

# 7. Tag and back-merge
git checkout main
git pull origin main
git tag -a homepage-v1.1.0 -m "Release v1.1.0"
git push origin homepage-v1.1.0

git checkout develop
git merge main
git push origin develop
```

### 🆘 Emergency Hotfix

```bash
# 1. Branch from main
git checkout main
git pull origin main
git checkout -b hotfix/security-patch

# 2. Fix the issue
# ... apply fix ...
git add .
git commit -m "fix: patch security vulnerability"

# 3. Bump patch version
make version-bump VERSION=1.0.1

# 4. Push and create PR to main
git push origin hotfix/security-patch

# 5. After merge
# ✅ Deployed to production immediately

# 6. Back-merge to develop (IMPORTANT!)
git checkout develop
git merge main
git push origin develop
```

---

## Version Management Cheat Sheet

### Show Current Version
```bash
make version
# Output:
# Chart Version: 0.1.24
# Image Tag: 0.1.24
# Git Commit: abc1234
# Git Branch: develop
```

### Bump Version
```bash
# Manual version
make version-bump VERSION=1.1.0

# Auto-bump patch (1.0.0 → 1.0.1)
./scripts/version-manager.sh bump-patch

# Auto-bump minor (1.0.0 → 1.1.0)
./scripts/version-manager.sh bump-minor

# Auto-bump major (1.0.0 → 2.0.0)
./scripts/version-manager.sh bump-major
```

### Generate Release Notes
```bash
make release-notes

# Output:
## Release 1.1.0

### Changes
- feat: add dark mode
- fix: resolve API timeout
- chore: update dependencies

### Docker Images
- ghcr.io/brunovlucena/homelab/homepage-api:1.1.0
- ghcr.io/brunovlucena/homelab/homepage-frontend:1.1.0
```

---

## Build & Deploy

### Local Development
```bash
# Start services
make up

# View logs
make logs

# Test API
make test-api-endpoints

# Run all tests
make test
```

### Build Images with Version
```bash
# Build and push with semantic version
make build-push-version

# This creates:
# - homepage-api:1.0.0
# - homepage-api:latest
# - homepage-frontend:1.0.0
# - homepage-frontend:latest
```

### Manual Deployment
```bash
# Trigger Flux reconciliation
make reconcile

# Or manually
flux reconcile source git homelab -n flux-system
flux reconcile helmrelease homepage -n homepage
```

---

## Debugging

### Check Deployment Status
```bash
# View pods
kubectl get pods -n homepage

# Check rollout status
kubectl rollout status deployment/homepage-api -n homepage
kubectl rollout status deployment/homepage-frontend -n homepage

# View events
kubectl get events -n homepage --sort-by='.lastTimestamp'
```

### View Logs
```bash
# API logs
kubectl logs -f deployment/homepage-api -n homepage

# Frontend logs
kubectl logs -f deployment/homepage-frontend -n homepage

# Previous pod logs (if crashed)
kubectl logs deployment/homepage-api -n homepage --previous
```

### Rollback
```bash
# Rollback to previous version
kubectl rollout undo deployment/homepage-api -n homepage

# Or via Helm
helm rollback homepage -n homepage
```

---

## Common Scenarios

### Scenario 1: Feature Development → Staging → Production
```bash
# Day 1: Develop feature
feature/add-search → develop (merged)
✅ Auto-deploy to staging

# Day 2: Test in staging
# Run tests, verify functionality

# Day 3: Release to production
release/v1.1.0 → main (merged)
✅ Auto-deploy to production
```

### Scenario 2: Critical Production Bug
```bash
# Immediate hotfix
hotfix/fix-crash → main (merged)
✅ Deploy to production (v1.0.1)

# Back-merge to develop
main → develop (merged)
✅ Staging updated with fix
```

### Scenario 3: Multiple Features in Parallel
```bash
# Multiple developers
Developer A: feature/user-auth → develop
Developer B: feature/dark-mode → develop
Developer C: feature/notifications → develop

# All merge to develop independently
# Test together in staging
# Release all features in v1.2.0
```

---

## Git Aliases (Optional)

Add to your `~/.gitconfig`:

```ini
[alias]
    # Branch management
    new-feature = "!f() { git checkout develop && git pull && git checkout -b feature/$1; }; f"
    new-bugfix = "!f() { git checkout develop && git pull && git checkout -b bugfix/$1; }; f"
    new-hotfix = "!f() { git checkout main && git pull && git checkout -b hotfix/$1; }; f"
    new-release = "!f() { git checkout develop && git pull && git checkout -b release/v$1; }; f"
    
    # Quick commands
    co = checkout
    st = status
    cm = commit -m
    pom = push origin main
    pod = push origin develop
    pf = "!git push origin $(git rev-parse --abbrev-ref HEAD)"
```

Usage:
```bash
# Create feature branch
git new-feature add-search

# Create hotfix branch
git new-hotfix security-patch

# Create release branch
git new-release 1.1.0
```

---

## Troubleshooting

### Problem: Version bump failed
```bash
# Ensure you're on correct branch
git status

# Ensure Chart.yaml exists
ls -la chart/Chart.yaml

# Try manual version bump
./scripts/version-manager.sh bump 1.1.0
```

### Problem: CI/CD not triggering
```bash
# Check GitHub Actions
# Go to: https://github.com/brunovlucena/homelab/actions

# Verify path filters match your changes
# API: flux/clusters/homelab/infrastructure/homepage/api/**
# Frontend: flux/clusters/homelab/infrastructure/homepage/frontend/**
```

### Problem: Helm release not updating
```bash
# Force reconciliation
flux reconcile source git homelab -n flux-system
flux reconcile helmrelease homepage -n homepage --timeout 10m

# Check Flux logs
flux logs --all-namespaces --follow
```

---

## Resources

- 📚 [Full Versioning Strategy](./VERSIONING_STRATEGY.md)
- 🌿 [Branching Guide](./BRANCHING_GUIDE.md)
- 📝 [CHANGELOG](../CHANGELOG.md)
- 🔧 [Makefile Commands](../Makefile)

---

## Need Help?

```bash
# Show all make commands
make help

# Show version manager commands
./scripts/version-manager.sh help

# Check current version
make version-info
```

---

**Remember**: 
- ✅ Always work on feature branches
- ✅ Use conventional commits
- ✅ Test in staging before production
- ✅ Update CHANGELOG for releases
- ✅ Tag releases on main branch

