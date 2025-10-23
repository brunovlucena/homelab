# 🌿 Knative Lambda Branching Guide

## Overview
This guide provides detailed branching strategies and workflows for the Knative Lambda Builder project.

## Table of Contents
- [Branch Types](#branch-types)
- [Branching Workflows](#branching-workflows)
- [Pull Request Guidelines](#pull-request-guidelines)
- [Commit Message Convention](#commit-message-convention)
- [Code Review Process](#code-review-process)
- [Release Management](#release-management)

---

## 🌳 Branch Types

### Main Branches

#### `main` - Production Branch
```
Purpose:      Production-ready code
Protection:   ✅ Required reviews (2), ✅ CI checks, ✅ Signed commits
Deployment:   Automatically deploys to production (prd environment)
Version Tag:  v1.0.0, v1.1.0, v2.0.0
Image Tags:   :1.0.0, :prd, :latest
```

**Rules:**
- Never commit directly to `main`
- All changes via PR from `release/*` or `hotfix/*`
- Must pass all CI/CD checks
- Requires 2 approvals
- Automatically tagged and deployed

#### `develop` - Integration Branch
```
Purpose:      Integration and continuous development
Protection:   ✅ Required reviews (1), ✅ CI checks
Deployment:   Automatically deploys to dev environment
Version Tag:  v1.1.0-beta.1, v1.1.0-beta.2
Image Tags:   :1.1.0-beta.1, :dev
```

**Rules:**
- Primary integration branch
- All feature branches merge here
- Continuous deployment to dev
- Must pass CI/CD checks
- Requires 1 approval

---

### Feature Branches

#### `feature/*` - Feature Development
```
Naming:       feature/short-description
Example:      feature/add-build-cache
              feature/arm64-support
              feature/improve-metrics
Source:       Branch from develop
Merge to:     develop (via PR)
Lifespan:     1-3 days (short-lived)
Version Tag:  v1.1.0-dev.abc1234
```

**Workflow:**
```bash
# Create feature branch
git checkout develop
git pull origin develop
git checkout -b feature/add-build-cache

# Make changes
git add .
git commit -m "feat(builder): implement kaniko cache"

# Push and create PR
git push origin feature/add-build-cache
# Create PR: feature/add-build-cache → develop
```

**Naming Convention:**
- Use lowercase with hyphens
- Be descriptive but concise
- Start with what you're doing
  - `feature/add-*` - Adding new functionality
  - `feature/improve-*` - Improving existing functionality
  - `feature/refactor-*` - Code refactoring

---

#### `bugfix/*` - Bug Fixes
```
Naming:       bugfix/issue-description
Example:      bugfix/fix-job-timeout
              bugfix/resolve-race-condition
Source:       Branch from develop
Merge to:     develop (via PR)
Lifespan:     1-2 days
```

**Workflow:**
```bash
# Create bugfix branch
git checkout develop
git pull origin develop
git checkout -b bugfix/fix-job-timeout

# Fix the bug
git add .
git commit -m "fix(job-manager): increase job timeout to 30m"

# Push and create PR
git push origin bugfix/fix-job-timeout
# Create PR: bugfix/fix-job-timeout → develop
```

---

#### `hotfix/*` - Critical Production Fixes
```
Naming:       hotfix/critical-issue
Example:      hotfix/security-patch
              hotfix/memory-leak
Source:       Branch from main
Merge to:     main AND develop (via PR)
Lifespan:     Hours to 1 day
Version:      Patch bump (1.0.0 → 1.0.1)
```

**Workflow:**
```bash
# Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b hotfix/memory-leak

# Fix the critical issue
git add .
git commit -m "fix(builder): resolve memory leak in job cleanup"

# Bump version
make version-bump VERSION=1.0.1

# Commit version bump
git add .
git commit -m "chore: bump version to v1.0.1"

# Push and create PR to main
git push origin hotfix/memory-leak
# Create PR: hotfix/memory-leak → main

# After merge to main, back-merge to develop
git checkout develop
git pull origin develop
git merge main
git push origin develop
```

**Critical Rules:**
- Only for production emergencies
- Branch from `main`, not `develop`
- Must merge to both `main` AND `develop`
- Bump patch version
- Deploy immediately to production

---

#### `release/*` - Release Preparation
```
Naming:       release/vX.Y.Z
Example:      release/v1.1.0
              release/v2.0.0
Source:       Branch from develop
Merge to:     main (then back-merge to develop)
Lifespan:     1-2 days
Purpose:      Final testing, version bump, changelog
```

**Workflow:**
```bash
# Create release branch
git checkout develop
git pull origin develop
git checkout -b release/v1.1.0

# Bump version
make version-bump VERSION=1.1.0

# Update CHANGELOG.md
# - Add release notes
# - Document breaking changes
# - List new features and fixes

# Commit changes
git add .
git commit -m "chore: prepare release v1.1.0"

# Push and create PR to main
git push origin release/v1.1.0
# Create PR: release/v1.1.0 → main

# After merge and deployment
git checkout main
git pull origin main
git tag -a knative-lambda-v1.1.0 -m "Release v1.1.0"
git push origin knative-lambda-v1.1.0

# Back-merge to develop
git checkout develop
git pull origin develop
git merge main
git push origin develop
```

---

## 🔄 Branching Workflows

### Standard Feature Development

```
develop
  │
  ├─ feature/add-cache
  │   │
  │   ├─ Develop feature
  │   ├─ Test locally
  │   ├─ Commit changes
  │   └─ PR to develop
  │
  └─ (merge) ──> develop
       │
       └─ Auto-deploy to dev environment
```

### Release Process

```
develop
  │
  ├─ release/v1.1.0
  │   │
  │   ├─ Bump version
  │   ├─ Update changelog
  │   ├─ Final testing
  │   └─ PR to main
  │
main
  │
  ├─ (merge from release/v1.1.0)
  ├─ Tag: knative-lambda-v1.1.0
  ├─ Auto-deploy to production
  │
  └─ Back-merge to develop
```

### Hotfix Process

```
main (v1.0.0)
  │
  ├─ hotfix/critical-fix
  │   │
  │   ├─ Fix issue
  │   ├─ Bump to v1.0.1
  │   └─ PR to main
  │
main (v1.0.1)
  │
  ├─ Tag: knative-lambda-v1.0.1
  ├─ Auto-deploy to production
  │
  └─ Back-merge to develop
       │
       └─ Develop also has the fix
```

---

## 📋 Pull Request Guidelines

### PR Title Format

Use conventional commit format:

```
<type>(<scope>): <description>

Examples:
feat(builder): add kaniko build caching
fix(job-manager): resolve race condition in job cleanup
docs(readme): update installation instructions
chore(deps): upgrade golang to 1.21
```

### PR Description Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## How Has This Been Tested?
- [ ] Unit tests
- [ ] Integration tests
- [ ] Manual testing in dev environment

## Checklist
- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings or errors
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes

## Related Issues
Closes #123

## Screenshots (if applicable)

## Additional Notes
```

### PR Size Guidelines

- **Small PR**: < 200 lines changed ✅ Preferred
- **Medium PR**: 200-500 lines changed ⚠️ Acceptable
- **Large PR**: > 500 lines changed ❌ Should be split

**Tips for keeping PRs small:**
- One feature/fix per PR
- Split large features into multiple PRs
- Use feature flags for incremental rollout

---

## ✍️ Commit Message Convention

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
- `docs`: Documentation changes
- `style`: Code style changes (formatting, no logic change)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Build process, dependencies, tooling
- `ci`: CI/CD changes

### Scopes

- `builder`: Main builder service
- `sidecar`: Build monitor sidecar
- `metrics-pusher`: Metrics pusher component
- `job-manager`: Job management logic
- `service-manager`: Service management logic
- `helm`: Helm chart changes
- `docs`: Documentation

### Examples

```bash
# Feature
git commit -m "feat(builder): add kaniko build caching support"

# Bug fix
git commit -m "fix(job-manager): resolve race condition in cleanup"

# Breaking change
git commit -m "feat(builder)!: change build event schema

BREAKING CHANGE: build events now require 'version' field"

# Documentation
git commit -m "docs(readme): update installation instructions"

# Chore
git commit -m "chore(deps): upgrade golang to 1.21"
```

---

## 👀 Code Review Process

### Review Checklist

#### Functionality
- [ ] Code does what it's supposed to do
- [ ] Edge cases are handled
- [ ] Error handling is appropriate
- [ ] No obvious bugs

#### Code Quality
- [ ] Code is readable and maintainable
- [ ] Follows project conventions
- [ ] No code duplication (DRY principle)
- [ ] Proper use of abstractions

#### Testing
- [ ] Unit tests are present and meaningful
- [ ] Integration tests cover key scenarios
- [ ] Tests are maintainable
- [ ] Test coverage hasn't decreased

#### Performance
- [ ] No obvious performance issues
- [ ] Appropriate algorithms and data structures
- [ ] Resource usage is reasonable

#### Security
- [ ] No security vulnerabilities introduced
- [ ] Sensitive data is properly handled
- [ ] Input validation is present

#### Documentation
- [ ] Code is properly commented
- [ ] Complex logic is explained
- [ ] Public APIs are documented
- [ ] README/docs are updated if needed

### Review Response Time

- **Priority**: Hotfixes - Same day
- **High**: Features blocking others - 1 day
- **Normal**: Features and fixes - 2 days
- **Low**: Documentation, refactoring - 3 days

---

## 🚀 Release Management

### Release Cadence

- **Major releases**: Every 6 months (or when breaking changes are needed)
- **Minor releases**: Monthly (new features)
- **Patch releases**: As needed (bug fixes)
- **Hotfixes**: Immediately (critical issues)

### Release Checklist

#### Pre-Release
- [ ] All tests passing
- [ ] No critical bugs
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Version bumped

#### Release
- [ ] Create release branch
- [ ] Final testing in dev
- [ ] Create PR to main
- [ ] Get approvals
- [ ] Merge to main
- [ ] Tag release
- [ ] Deploy to production

#### Post-Release
- [ ] Monitor deployment
- [ ] Check metrics and logs
- [ ] Back-merge to develop
- [ ] Announce release
- [ ] Update GitHub release notes

---

## 📊 Branch Lifecycle

```
┌────────────────┐
│  Create Branch │
│  from develop  │
└───────┬────────┘
        │
        v
┌────────────────┐
│   Develop      │
│   Feature      │
└───────┬────────┘
        │
        v
┌────────────────┐
│  Test Locally  │
└───────┬────────┘
        │
        v
┌────────────────┐
│  Push & Create │
│      PR        │
└───────┬────────┘
        │
        v
┌────────────────┐
│  Code Review   │
└───────┬────────┘
        │
        v
┌────────────────┐
│     Merge      │
└───────┬────────┘
        │
        v
┌────────────────┐
│ Delete Branch  │
└────────────────┘
```

---

## 🎯 Best Practices

### DO ✅

- Keep branches short-lived (1-3 days)
- Make small, focused commits
- Write meaningful commit messages
- Test before pushing
- Rebase on develop frequently
- Delete branches after merge
- Update documentation
- Use conventional commits

### DON'T ❌

- Don't commit directly to main or develop
- Don't create long-lived feature branches
- Don't mix multiple features in one branch
- Don't push broken code
- Don't skip code reviews
- Don't forget to back-merge hotfixes
- Don't use generic commit messages
- Don't push sensitive data

---

## 🔧 Git Commands Cheat Sheet

```bash
# Create and switch to feature branch
git checkout -b feature/my-feature develop

# Keep feature branch updated
git checkout feature/my-feature
git fetch origin
git rebase origin/develop

# Interactive rebase to clean up commits
git rebase -i origin/develop

# Amend last commit
git commit --amend

# Stash changes
git stash
git stash pop

# View branch history
git log --oneline --graph --all

# Delete local branch
git branch -d feature/my-feature

# Delete remote branch
git push origin --delete feature/my-feature
```

---

## 📚 References

- [Trunk-Based Development](https://trunkbaseddevelopment.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [GitFlow](https://nvie.com/posts/a-successful-git-branching-model/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)

---

## 🆘 Common Issues

### Issue: Merge Conflicts

```bash
# Update your branch
git checkout feature/my-feature
git fetch origin
git rebase origin/develop

# Resolve conflicts
# Edit conflicting files
git add <resolved-files>
git rebase --continue
```

### Issue: Accidentally Committed to Wrong Branch

```bash
# If not pushed yet
git reset HEAD~1
git stash
git checkout correct-branch
git stash pop
git add .
git commit
```

### Issue: Need to Update PR After Review

```bash
# Make changes
git add .
git commit -m "fix: address review comments"
git push origin feature/my-feature
# PR automatically updates
```

---

**Remember**: Good branching practices lead to a clean git history, easier code reviews, and fewer deployment issues! 🚀


