# 🚀 Quick Start: Knative Lambda Versioning & Release Management

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
git clone https://github.com/notifi-network/infra.git
cd infra/flux/clusters/homelab/infrastructure/knative-lambda

# Make scripts executable
chmod +x scripts/*.sh
```

### 2. Install Dependencies
```bash
# Go modules
go mod download
go mod tidy

# Install build tools
make setup
```

### 3. Set Up Local Development
```bash
# Build all components
make build-all

# Run locally
make run

# Access locally
# Service: http://localhost:8080
```

---

## Daily Workflows

### 📝 Working on a Feature

```bash
# 1. Start from develop
git checkout develop
git pull origin develop

# 2. Create feature branch
git checkout -b feature/add-build-cache

# 3. Develop and test locally
make build-all
make test

# 4. Commit with conventional commits
git add .
git commit -m "feat(builder): add kaniko build caching"

# 5. Push and create PR
git push origin feature/add-build-cache
# Go to GitHub → Create PR to develop

# 6. After PR approval
# ✅ Automatically built with tag: v1.1.0-dev.abc1234
# ✅ Deployed to dev environment
```

### 🔄 Testing in Dev Environment

```bash
# View dev deployment
kubectl get pods -n knative-lambda-dev

# Check logs
kubectl logs -f deployment/knative-lambda-builder -n knative-lambda-dev

# Trigger test lambda build
make trigger-build-dev

# Check RabbitMQ status
make rabbitmq-status ENV=dev
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
# ✅ Tags: knative-lambda-v1.1.0

# 7. Tag and back-merge
git checkout main
git pull origin main
git tag -a knative-lambda-v1.1.0 -m "Release v1.1.0"
git push origin knative-lambda-v1.1.0

git checkout develop
git merge main
git push origin develop
```

### 🆘 Emergency Hotfix

```bash
# 1. Branch from main
git checkout main
git pull origin main
git checkout -b hotfix/fix-job-cleanup

# 2. Fix the issue
# ... apply fix ...
git add .
git commit -m "fix: resolve job cleanup race condition"

# 3. Bump patch version
make version-bump VERSION=1.0.1

# 4. Push and create PR to main
git push origin hotfix/fix-job-cleanup

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
make version-info
# Output:
# Chart Version: 1.0.0
# Service Tag: 1.0.0
# Sidecar Tag: 1.0.0
# Metrics Pusher Tag: 1.0.0
# Git Commit: abc1234
# Git Branch: main
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
- feat: add build caching
- fix: resolve job timeout
- chore: update dependencies

### Docker Images
- knative-lambda-builder:1.1.0
- knative-lambda-sidecar:1.1.0
- knative-lambda-metrics-pusher:1.1.0
```

---

## Build & Deploy

### Local Development
```bash
# Build all binaries
make build-all

# Run service locally
make run

# Run tests
make test

# Lint code
make lint
```

### Build Images with Version
```bash
# Build all images for production
make build-and-push-all-prd

# This creates:
# - knative-lambda-builder:1.0.0
# - knative-lambda-builder:prd
# - knative-lambda-builder:latest
# - knative-lambda-sidecar:1.0.0
# - knative-lambda-sidecar:prd
# - knative-lambda-sidecar:latest
# - knative-lambda-metrics-pusher:1.0.0
# - knative-lambda-metrics-pusher:prd
# - knative-lambda-metrics-pusher:latest
```

### Manual Deployment
```bash
# Deploy to dev
kubectl apply -k deploy/overlays/dev

# Deploy to prd
kubectl apply -k deploy/overlays/prd

# Or use ArgoCD/Flux reconciliation
argocd app sync knative-lambda-prd
flux reconcile helmrelease knative-lambda -n knative-lambda-prd
```

---

## Debugging

### Check Deployment Status
```bash
# View pods
kubectl get pods -n knative-lambda-prd

# Check rollout status
kubectl rollout status deployment/knative-lambda-builder -n knative-lambda-prd

# View events
kubectl get events -n knative-lambda-prd --sort-by='.lastTimestamp'
```

### View Logs
```bash
# Builder service logs
kubectl logs -f deployment/knative-lambda-builder -n knative-lambda-prd

# Sidecar logs
kubectl logs -f deployment/knative-lambda-builder -c sidecar -n knative-lambda-prd

# Previous pod logs (if crashed)
kubectl logs deployment/knative-lambda-builder -n knative-lambda-prd --previous
```

### Rollback
```bash
# Rollback to previous version
kubectl rollout undo deployment/knative-lambda-builder -n knative-lambda-prd

# Or via Helm
helm rollback knative-lambda -n knative-lambda-prd
```

---

## Common Scenarios

### Scenario 1: Feature Development → Dev → Production
```bash
# Day 1: Develop feature
feature/add-cache → develop (merged)
✅ Auto-deploy to dev environment

# Day 2: Test in dev
# Run integration tests, verify functionality
make trigger-build-dev
make rabbitmq-status ENV=dev

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
✅ Dev environment updated with fix
```

### Scenario 3: Multiple Features in Parallel
```bash
# Multiple developers
Developer A: feature/add-cache → develop
Developer B: feature/arm64-support → develop
Developer C: feature/metrics-improvement → develop

# All merge to develop independently
# Test together in dev environment
# Release all features in v1.2.0
```

---

## Testing Workflows

### Test Lambda Creation
```bash
# Trigger lambda build in dev
make trigger-build-dev

# Trigger lambda service creation
make trigger-lambda-dev

# Check status
kubectl get ksvc -n knative-lambda-dev
```

### Test in Production (Canary)
```bash
# Deploy with canary strategy
# 1. Deploy v1.1.0 alongside v1.0.0
# 2. Route 10% traffic to v1.1.0
# 3. Monitor metrics
# 4. Gradually increase traffic
# 5. Full cutover to v1.1.0
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
git new-feature add-cache

# Create hotfix branch
git new-hotfix security-patch

# Create release branch
git new-release 1.1.0
```

---

## Environment-Specific Commands

### Development Environment
```bash
# Build and push for dev
make build-and-push-all-dev

# Trigger test events
make trigger-build-dev
make trigger-lambda-dev

# View status
make rabbitmq-status ENV=dev
kubectl get pods -n knative-lambda-dev
```

### Production Environment
```bash
# Build and push for production
make build-and-push-all-prd

# Trigger test events (use with caution!)
make trigger-build-prd
make trigger-lambda-prd

# View status
make rabbitmq-status ENV=prd
kubectl get pods -n knative-lambda-prd
```

---

## Troubleshooting

### Problem: Version bump failed
```bash
# Ensure you're on correct branch
git status

# Ensure Chart.yaml exists
ls -la deploy/Chart.yaml

# Try manual version bump
./scripts/version-manager.sh bump 1.1.0
```

### Problem: CI/CD not triggering
```bash
# Check GitHub Actions
# Go to: https://github.com/notifi-network/infra/actions

# Verify path filters match your changes
# knative-lambda: flux/clusters/homelab/infrastructure/knative-lambda/**
```

### Problem: Image not pulling
```bash
# Check ECR login
aws ecr get-login-password --region us-west-2 | \
  docker login --username AWS --password-stdin \
  339954290315.dkr.ecr.us-west-2.amazonaws.com

# Verify image exists
aws ecr describe-images \
  --repository-name knative-lambdas/knative-lambda-builder \
  --region us-west-2
```

### Problem: Helm release not updating
```bash
# Force reconciliation (ArgoCD)
argocd app sync knative-lambda-prd

# Force reconciliation (Flux)
flux reconcile source git homelab -n flux-system
flux reconcile helmrelease knative-lambda -n knative-lambda-prd --timeout 10m
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

## Important Makefile Targets

### Building
- `make build-all` - Build all binaries
- `make docker-build` - Build all Docker images
- `make docker-push` - Push all Docker images

### Testing
- `make test` - Run all tests
- `make lint` - Run linting
- `make trigger-build-dev` - Trigger test build in dev

### Versioning
- `make version-info` - Show current version
- `make version-bump VERSION=X.Y.Z` - Bump version
- `make release-notes` - Generate release notes

### Deployment
- `make build-and-push-all-dev` - Build and push dev images
- `make build-and-push-all-prd` - Build and push production images

### Debugging
- `make rabbitmq-status ENV=dev` - Check RabbitMQ status
- `make rabbitmq-purge-lambda-queues-dev` - Purge dev queues
- `make clean-knative-lambda` - Clean Knative resources

---

**Remember**: 
- ✅ Always work on feature branches
- ✅ Use conventional commits
- ✅ Test in dev before production
- ✅ Update CHANGELOG for releases
- ✅ Tag releases on main branch
- ✅ Never push to main directly
- ✅ All three components (builder, sidecar, metrics-pusher) share the same version


