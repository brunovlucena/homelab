# 🌿 Homepage Branching & Release Guide

## Quick Reference

```
main (production) ────────┬─────────────── v1.0.0 ────┬────── v1.1.0 ────
                          │                           │
develop (staging) ────────┼───────┬───────────────────┼──────────
                          │       │                   │
feature/add-auth ─────────┤       └─── (merged)       │
                          │                           │
bugfix/fix-typo ──────────┘                           └─── (merged)
```

## Branch Types

### 🔵 `main` - Production Branch
- **Protected**: ✅ Requires 2 PR approvals
- **Deployment**: Automatically deploys to production
- **Version Tags**: All releases are tagged here (e.g., `homepage-v1.0.0`)
- **CI/CD**: `.github/workflows/homepage-cd-production.yml`

### 🟢 `develop` - Integration Branch
- **Protected**: ✅ Requires 1 PR approval
- **Deployment**: Automatically deploys to staging
- **Version Tags**: Pre-release versions (e.g., `v1.1.0-beta.1`)
- **CI/CD**: `.github/workflows/homepage-cd-staging.yml`

### 🟡 `feature/*` - Feature Branches
- **Naming**: `feature/YYYY-MM-DD/<description>`
- **Examples**: 
  - `feature/2025-10-16/user-authentication`
  - `feature/2025-10-17/dark-mode`
  - `feature/2025-10-18/api-versioning`
- **Branch from**: `develop`
- **Merge to**: `develop`
- **Lifespan**: 1-3 days (short-lived)
- **Helper**: `gfeature <description>` (auto-adds date)

### 🟠 `bugfix/*` - Bug Fix Branches
- **Naming**: `bugfix/YYYY-MM-DD/<description>`
- **Examples**:
  - `bugfix/2025-10-16/login-error`
  - `bugfix/2025-10-17/api-timeout`
- **Branch from**: `develop`
- **Merge to**: `develop`
- **Helper**: `gbugfix <description>` (auto-adds date)

### 🔴 `hotfix/*` - Hotfix Branches
- **Naming**: `hotfix/YYYY-MM-DD/<description>`
- **Examples**:
  - `hotfix/2025-10-16/security-vulnerability`
  - `hotfix/2025-10-17/critical-crash`
- **Branch from**: `main`
- **Merge to**: `main` AND `develop`
- **Urgency**: Critical production issues only
- **Helper**: `ghotfix <description>` (auto-adds date)

### 🟣 `release/*` - Release Branches
- **Naming**: `release/v<version>`
- **Examples**: `release/v1.1.0`
- **Branch from**: `develop`
- **Merge to**: `main` (then back to `develop`)
- **Purpose**: Final testing, version bumps, changelog updates

---

## Workflow Examples

### 1️⃣ Feature Development (Standard Flow)

```bash
# 1. Create feature branch (using helper function)
gfeature add user profile
# This creates: feature/2025-10-16/add-user-profile

# Or manually:
# git checkout develop && git pull origin develop
# git checkout -b feature/2025-10-16/add-user-profile

# 2. Make changes
# ... code changes ...

# 3. Commit with conventional commits (using helper)
gcommit feat add user profile page
# Or: git add . && git commit -m "feat: add user profile page"

# 4. Push to remote (using helper)
gfeature-complete
# Or: git push origin feature/2025-10-16/add-user-profile

# 5. Create Pull Request on GitHub
# feature/2025-10-16/add-user-profile → develop

# 6. After PR approval and merge
# CI/CD automatically builds and deploys to staging
# Images tagged as: v0.1.24-beta.{BUILD_NUMBER}
```

**PR Template:**
```markdown
## Description
Add user profile page with avatar upload

## Type of Change
- [x] New feature
- [ ] Bug fix
- [ ] Breaking change

## Testing
- [x] Unit tests added
- [x] Integration tests passed
- [x] Tested in local environment

## Screenshots
[Add screenshots if UI change]
```

---

### 2️⃣ Production Release (Staging → Production)

```bash
# 1. Ensure develop is stable and tested in staging
# All features merged to develop and tested

# 2. Create and prepare release (using helper)
grelease-complete 1.1.0
# This creates branch, bumps version, and prompts for CHANGELOG

# Or manually:
# grelease 1.1.0  # Creates release/v1.1.0
# gbump 1.1.0     # Bumps version

# 3. Update CHANGELOG.md (MANUAL STEP)
# Add release notes for v1.1.0

# 4. Push release branch (using helper)
grelease-push

# Or manually:
# git add .
# git commit -m "chore: release v1.1.0"
# git push origin release/v1.1.0

# 7. Create Pull Request to main
# Go to GitHub and create PR: release/v1.1.0 → main

# 8. After PR approval and merge to main
# CI/CD automatically:
#   - Builds and pushes images with tag v1.1.0
#   - Deploys to production
#   - Creates GitHub release

# 9. Tag the release
git checkout main
git pull origin main
git tag -a homepage-v1.1.0 -m "Release v1.1.0"
git push origin homepage-v1.1.0

# 10. Back-merge to develop
git checkout develop
git merge main
git push origin develop
```

---

### 3️⃣ Hotfix (Urgent Production Fix)

```bash
# 1. Create hotfix branch (using helper)
ghotfix security patch
# This creates: hotfix/2025-10-16/security-patch from main

# Or manually:
# git checkout main && git pull origin main
# git checkout -b hotfix/2025-10-16/security-patch

# 2. Fix the critical issue
# ... fix code ...
gcommit fix patch security vulnerability CVE-2024-XXXX
# Or: git add . && git commit -m "fix: patch security vulnerability CVE-2024-XXXX"

# 3. Bump patch version (using helper)
gbump patch
# This will bump from v1.1.0 → v1.1.1

# 4. Push hotfix branch (using helper)
ghotfix-complete
# Or: git push origin hotfix/2025-10-16/security-patch

# 5. Create PR to main
# Go to GitHub and create PR: hotfix/security-patch → main

# 6. After approval and merge
# CI/CD automatically deploys to production

# 7. Tag the hotfix
git checkout main
git pull origin main
git tag -a homepage-v1.1.1 -m "Hotfix v1.1.1 - Security patch"
git push origin homepage-v1.1.1

# 8. IMPORTANT: Back-merge to develop
git checkout develop
git merge main
git push origin develop
```

---

### 4️⃣ Bug Fix (Non-Critical)

```bash
# 1. Create bugfix branch (using helper)
gbugfix fix api timeout
# This creates: bugfix/2025-10-16/fix-api-timeout from develop

# 2. Fix the bug
gcommit fix increase API timeout to 30s

# 3. Push and create PR (using helper)
gfeature-complete  # Works for bugfix branches too
# Create PR: bugfix/2025-10-16/fix-api-timeout → develop

# 4. After merge, test in staging
# Then include in next production release
```

---

## Commit Message Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/).

### Format
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, no logic change)
- `refactor`: Code refactoring
- `perf`: Performance improvement
- `test`: Add or update tests
- `chore`: Build/tooling changes
- `ci`: CI/CD changes

### Examples
```bash
# Good commits
git commit -m "feat(api): add user authentication endpoint"
git commit -m "fix(frontend): resolve memory leak in chatbot"
git commit -m "docs: update README with deployment instructions"
git commit -m "refactor(api): simplify database connection logic"
git commit -m "chore: bump version to 1.1.0"

# Bad commits (too vague)
git commit -m "fixed bug"
git commit -m "updates"
git commit -m "wip"
```

---

## Version Numbering

### Semantic Versioning (MAJOR.MINOR.PATCH)

```
v1.2.3
│ │ │
│ │ └─ PATCH: Bug fixes, security patches (backwards-compatible)
│ └─── MINOR: New features (backwards-compatible)
└───── MAJOR: Breaking changes
```

### Examples

| Current | Change | New Version | Reason |
|---------|--------|-------------|--------|
| 1.0.0 | Add new API endpoint | 1.1.0 | New feature (minor) |
| 1.1.0 | Fix API bug | 1.1.1 | Bug fix (patch) |
| 1.1.1 | Change API response format | 2.0.0 | Breaking change (major) |
| 1.2.0 | Add dark mode | 1.3.0 | New feature (minor) |

### Pre-release Versions

```
1.1.0-beta.1    # Beta release (staging)
1.1.0-rc.1      # Release candidate
1.1.0-dev.abc123 # Development build
```

---

## CI/CD Pipeline

### On Feature Branch Push
```
1. Lint & Test (homepage-ci.yml)
2. Build images (no push)
3. Security scan
4. Report status on PR
```

### On Merge to Develop
```
1. Run CI tests
2. Build images with tag: v{VERSION}-beta.{BUILD_NUMBER}
3. Push to GHCR
4. Deploy to staging
5. Run smoke tests
```

### On Merge to Main
```
1. Run CI tests
2. Build images with tag: v{VERSION} and latest
3. Push to GHCR with attestations
4. Deploy to production
5. Run smoke tests
6. Run performance tests
7. Create GitHub release
```

---

## Pull Request Guidelines

### Required Checks
- ✅ All CI tests pass
- ✅ Code coverage maintained
- ✅ No linting errors
- ✅ Security scan passes
- ✅ Conventional commit format

### For `develop` → `main` PR
- ✅ All staging tests passed
- ✅ Version bumped
- ✅ CHANGELOG.md updated
- ✅ Release notes prepared
- ✅ 2 approvals required

---

## Rollback Procedures

### Kubernetes Rollback
```bash
# Quick rollback to previous version
kubectl rollout undo deployment/homepage-api -n homepage
kubectl rollout undo deployment/homepage-frontend -n homepage

# Rollback to specific version
kubectl rollout undo deployment/homepage-api -n homepage --to-revision=3
```

### Helm Rollback
```bash
# Rollback to previous release
helm rollback homepage -n homepage

# Rollback to specific release number
helm rollback homepage 5 -n homepage
```

### Git Revert
```bash
# Revert a commit (creates new commit)
git revert <commit-hash>
git push origin main

# This triggers CI/CD to deploy the reverted version
```

---

## FAQ

### Q: How do I test my changes locally before pushing?
```bash
# Start local development environment
make up

# Run tests
make test

# Build images locally
make build-all
```

### Q: Can I push directly to main or develop?
No. All changes must go through Pull Requests with required approvals.

### Q: When should I create a hotfix vs. a bugfix?
- **Hotfix**: Critical production issues (security, crashes, data loss)
- **Bugfix**: Non-critical bugs that can wait for next release

### Q: How do I check the current version?
```bash
# Using Makefile
make version

# Using script
./scripts/version-manager.sh show

# Manual check
grep '^version:' chart/Chart.yaml
```

### Q: What if my PR conflicts with develop?
```bash
# Update your branch with latest develop
git checkout your-feature-branch
git fetch origin
git rebase origin/develop

# Resolve conflicts
# ... fix conflicts ...
git add .
git rebase --continue

# Force push (only on feature branches)
git push origin your-feature-branch --force-with-lease
```

---

## Tools & Scripts

### Version Management
```bash
# Show current version
make version
./scripts/version-manager.sh show

# Bump version
make version-bump VERSION=1.1.0
./scripts/version-manager.sh bump 1.1.0

# Auto-bump
./scripts/version-manager.sh bump-major  # 1.0.0 → 2.0.0
./scripts/version-manager.sh bump-minor  # 1.0.0 → 1.1.0
./scripts/version-manager.sh bump-patch  # 1.0.0 → 1.0.1

# Generate release notes
make release-notes
./scripts/version-manager.sh release-notes
```

### Build & Deploy
```bash
# Build with version
make build-push-version

# Deploy to staging (via Flux)
flux reconcile helmrelease homepage -n homepage

# View deployment status
kubectl rollout status deployment/homepage-api -n homepage
```

---

## References

- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Trunk-Based Development](https://trunkbaseddevelopment.com/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)
- [Keep a Changelog](https://keepachangelog.com/)

