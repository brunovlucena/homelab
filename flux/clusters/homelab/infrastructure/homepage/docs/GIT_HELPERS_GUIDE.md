# 🛠️ Git Helper Functions Guide

## Overview

This guide covers the comprehensive set of shell functions designed to streamline your Git workflow with the Homepage versioning strategy.

## Installation

### 1. Source the Functions

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
# Git Helper Functions for Homepage
source ~/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/homepage/scripts/git-helpers.sh
```

### 2. Reload Your Shell

```bash
# For zsh
source ~/.zshrc

# For bash
source ~/.bashrc
```

### 3. Verify Installation

```bash
ghelp
# Should display available commands
```

---

## Branch Creation Functions

### `gfeature` - Create Feature Branch

Creates a feature branch from `develop` with automatic date prefix.

**Usage:**
```bash
gfeature <description>
```

**Examples:**
```bash
gfeature add user authentication
# Creates: feature/2025-10-16/add-user-authentication

gfeature dark mode toggle
# Creates: feature/2025-10-16/dark-mode-toggle
```

**What it does:**
1. Switches to `develop` branch
2. Pulls latest changes
3. Creates `feature/YYYY-MM-DD/<description>` branch
4. Provides next step guidance

---

### `gbugfix` - Create Bugfix Branch

Creates a bugfix branch from `develop`.

**Usage:**
```bash
gbugfix <description>
```

**Examples:**
```bash
gbugfix api timeout issue
# Creates: bugfix/2025-10-16/api-timeout-issue

gbugfix memory leak chatbot
# Creates: bugfix/2025-10-16/memory-leak-chatbot
```

---

### `ghotfix` - Create Hotfix Branch

Creates a critical hotfix branch from `main` (production).

**Usage:**
```bash
ghotfix <description>
```

**Examples:**
```bash
ghotfix security vulnerability
# Creates: hotfix/2025-10-16/security-vulnerability

ghotfix production crash
# Creates: hotfix/2025-10-16/production-crash
```

**Important:**
- ⚠️ Hotfixes must be merged to BOTH `main` AND `develop`
- Function provides reminders about back-merge

---

### `grelease` - Create Release Branch

Creates a release branch from `develop`.

**Usage:**
```bash
grelease <version>
```

**Examples:**
```bash
grelease 1.1.0
# Creates: release/v1.1.0

grelease 2.0.0
# Creates: release/v2.0.0
```

**Provides guidance for:**
1. Running `make version-bump`
2. Updating CHANGELOG.md
3. Creating PR to main

---

### `gtask` - Create Task Branch

Creates a task branch for refactoring, chores, or other non-feature work.

**Usage:**
```bash
gtask <description>
```

**Examples:**
```bash
gtask refactor api handlers
# Creates: task/2025-10-16/refactor-api-handlers

gtask update dependencies
# Creates: task/2025-10-16/update-dependencies
```

---

## Commit Functions

### `gcommit` - Conventional Commit

Quick commit with conventional commit format.

**Usage:**
```bash
gcommit <type> <message>
```

**Valid Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `style` - Code style (no logic change)
- `refactor` - Code refactoring
- `perf` - Performance improvement
- `test` - Tests
- `chore` - Build/tooling
- `ci` - CI/CD changes

**Examples:**
```bash
gcommit feat add user authentication endpoint
# Commits: "feat: add user authentication endpoint"

gcommit fix resolve api timeout issue
# Commits: "fix: resolve api timeout issue"

gcommit docs update readme installation
# Commits: "docs: update readme installation"
```

---

### `gcommits` - Commit with Scope

Conventional commit with scope.

**Usage:**
```bash
gcommits <type> <scope> <message>
```

**Examples:**
```bash
gcommits feat api add user authentication endpoint
# Commits: "feat(api): add user authentication endpoint"

gcommits fix frontend resolve memory leak in chatbot
# Commits: "fix(frontend): resolve memory leak in chatbot"

gcommits docs readme update installation steps
# Commits: "docs(readme): update installation steps"
```

---

## Version Management Functions

### `gversion` - Show Current Version

Displays current version information.

**Usage:**
```bash
gversion
```

**Output:**
```
📋 Current Version Information:
Chart Version: 0.1.24
Image Tag: 0.1.24-dev.abc1234
Git Commit: abc1234
Git Branch: feature/2025-10-16/add-feature
```

---

### `gbump` - Bump Version

Bump version manually or automatically.

**Usage:**
```bash
# Specific version
gbump <version>

# Auto-bump
gbump major|minor|patch
```

**Examples:**
```bash
# Set specific version
gbump 1.1.0

# Auto-bump versions
gbump major   # 1.0.0 → 2.0.0
gbump minor   # 1.0.0 → 1.1.0
gbump patch   # 1.0.0 → 1.0.1
```

**What it updates:**
- `chart/Chart.yaml` - Chart version and appVersion
- `chart/values.yaml` - Image tags
- `frontend/package.json` - NPM version
- `api/VERSION` - API version file
- `CHANGELOG.md` - Adds version entry

---

### `grelnotes` - Generate Release Notes

Generates release notes from git history.

**Usage:**
```bash
grelnotes
```

**Output:**
```markdown
## Release 1.1.0

### Changes
- feat: add user authentication
- fix: resolve API timeout
- docs: update readme

### Docker Images
- ghcr.io/brunovlucena/homelab/homepage-api:1.1.0
- ghcr.io/brunovlucena/homelab/homepage-frontend:1.1.0
```

---

## Workflow Functions

### `gfeature-complete` - Complete Feature Workflow

Complete and push a feature/bugfix/task branch.

**Usage:**
```bash
gfeature-complete
```

**What it does:**
1. Checks you're on correct branch type
2. Pulls latest changes from develop
3. Runs tests
4. Pushes to remote
5. Reminds you to create PR

**Example workflow:**
```bash
# Create feature
gfeature add search

# Make changes
# ... code ...

# Commit
gcommit feat add search functionality

# Complete and push
gfeature-complete
# ✅ All done! Create PR on GitHub
```

---

### `grelease-complete` - Complete Release Workflow

Complete release preparation in one command.

**Usage:**
```bash
grelease-complete <version>
```

**Example:**
```bash
grelease-complete 1.1.0

# This does:
# 1. Creates release/v1.1.0 branch
# 2. Bumps version to 1.1.0
# 3. Prompts to update CHANGELOG.md
```

**After updating CHANGELOG:**
```bash
grelease-push
# Commits and pushes release branch
```

---

### `grelease-push` - Push Release Branch

Push prepared release branch.

**Usage:**
```bash
grelease-push
```

**Requirements:**
- Must be on a `release/*` branch
- CHANGELOG.md should be updated

**What it does:**
1. Commits changes with release message
2. Pushes to remote
3. Reminds you to create PR to main

---

### `ghotfix-complete` - Complete Hotfix Workflow

Complete and push a hotfix.

**Usage:**
```bash
ghotfix-complete
```

**Example workflow:**
```bash
# Create hotfix
ghotfix critical bug

# Fix the issue
# ... fix code ...

# Commit
gcommit fix resolve critical production bug

# Bump patch version
gbump patch

# Complete and push
ghotfix-complete
# ⚠️ REMEMBER: Back-merge to develop after main merge!
```

---

## Build & Deploy Functions

### `gbuild` - Build and Push

Build and push images with semantic versioning.

**Usage:**
```bash
gbuild
```

**What it does:**
- Builds API and frontend images
- Tags with semantic version
- Pushes to GHCR
- Also pushes `:latest` tag

---

### `greconcile` - Trigger Flux Reconciliation

Manually trigger Flux to reconcile the homepage deployment.

**Usage:**
```bash
greconcile
```

**Equivalent to:**
```bash
flux reconcile source git homelab -n flux-system
flux reconcile helmrelease homepage -n homepage
```

---

## Info Functions

### `ginfo` - Show Repository Status

Comprehensive repository status display.

**Usage:**
```bash
ginfo
```

**Output:**
```
📊 Current Repository Status

Branch: feature/2025-10-16/add-feature
Commit: abc1234

Version:
Chart Version: 0.1.24
Image Tag: 0.1.24-dev.abc1234
Git Commit: abc1234
Git Branch: feature/2025-10-16/add-feature

Status:
## feature/2025-10-16/add-feature
 M docs/README.md

Recent commits:
abc1234 feat: add feature
def5678 fix: resolve bug
...
```

---

### `ghelp` - Show Help

Display all available commands with examples.

**Usage:**
```bash
ghelp
```

---

## Complete Workflow Examples

### Example 1: Feature Development

```bash
# 1. Create feature
gfeature add user profile

# 2. Make changes
# ... edit files ...

# 3. Commit incrementally
gcommit feat add profile model
gcommit feat add profile API endpoint
gcommit feat add profile UI component

# 4. Complete feature
gfeature-complete

# 5. Create PR on GitHub
# feature/2025-10-16/add-user-profile → develop
```

---

### Example 2: Production Release

```bash
# 1. Start release
grelease-complete 1.2.0

# 2. Update CHANGELOG.md manually
# Add release notes

# 3. Push release
grelease-push

# 4. Create PR on GitHub
# release/v1.2.0 → main

# 5. After merge, tag the release
git checkout main
git pull
git tag -a homepage-v1.2.0 -m "Release v1.2.0"
git push origin homepage-v1.2.0

# 6. Back-merge to develop
git checkout develop
git merge main
git push origin develop
```

---

### Example 3: Emergency Hotfix

```bash
# 1. Create hotfix
ghotfix security vulnerability CVE-2024-12345

# 2. Apply fix
# ... edit files ...

# 3. Commit
gcommit fix patch security vulnerability CVE-2024-12345

# 4. Bump patch version
gbump patch
# 1.1.0 → 1.1.1

# 5. Complete hotfix
ghotfix-complete

# 6. Create PR to main
# hotfix/2025-10-16/security-vulnerability → main

# 7. After merge, BACK-MERGE TO DEVELOP
git checkout develop
git merge main
git push origin develop

# 8. Tag hotfix
git checkout main
git tag -a homepage-v1.1.1 -m "Hotfix v1.1.1"
git push origin homepage-v1.1.1
```

---

### Example 4: Bug Fix

```bash
# 1. Create bugfix
gbugfix api timeout

# 2. Fix the bug
# ... edit files ...

# 3. Commit with scope
gcommits fix api increase timeout to 30 seconds

# 4. Complete
gfeature-complete

# 5. Create PR
# bugfix/2025-10-16/api-timeout → develop
```

---

## Tips & Best Practices

### 1. Always Use Helper Functions
```bash
# ✅ Good
gfeature add authentication
gcommit feat add OAuth support

# ❌ Avoid manual commands (error-prone)
git checkout -b feature/add-authentication
git commit -m "added auth"
```

### 2. Check Version Before Release
```bash
gversion  # Check current version
grelnotes  # Preview release notes
```

### 3. Use Descriptive Branch Names
```bash
# ✅ Good
gfeature add user authentication with OAuth
gbugfix resolve memory leak in chatbot

# ❌ Too vague
gfeature new feature
gbugfix fix bug
```

### 4. Complete Workflows
```bash
# Use workflow functions for consistency
gfeature-complete  # For features/bugfixes
grelease-complete  # For releases
ghotfix-complete   # For hotfixes
```

### 5. Check Status Regularly
```bash
ginfo  # Quick status check
```

---

## Troubleshooting

### Function Not Found
```bash
# Ensure functions are sourced
source ~/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/homepage/scripts/git-helpers.sh

# Or add to ~/.zshrc and reload
source ~/.zshrc
```

### Wrong Branch Type
```bash
# Functions check branch type
# Error: "Not on a feature/bugfix/task branch"

# Solution: Only use workflow functions on correct branches
gfeature-complete  # Only on feature/bugfix/task branches
grelease-push      # Only on release branches
ghotfix-complete   # Only on hotfix branches
```

### Tests Failing
```bash
# gfeature-complete runs tests before pushing
# Fix tests before running again
cd flux/clusters/homelab/infrastructure/homepage
make test
```

---

## Reference Card

```bash
# Quick Reference
gfeature <desc>        # Create feature
gbugfix <desc>         # Create bugfix
ghotfix <desc>         # Create hotfix (from main!)
grelease <v>           # Create release
gtask <desc>           # Create task

gcommit <type> <msg>   # Conventional commit
gbump <v>              # Bump version
gversion               # Show version
ginfo                  # Show status

gfeature-complete      # Push feature
grelease-complete <v>  # Complete release
ghotfix-complete       # Push hotfix

ghelp                  # Full help
```

---

## Additional Resources

- [Branching Guide](./BRANCHING_GUIDE.md)
- [Versioning Strategy](./VERSIONING_STRATEGY.md)
- [Quick Start](./QUICK_START_VERSIONING.md)
- [CHANGELOG](../CHANGELOG.md)

