# 📚 Homepage Documentation

Welcome to the Homepage application documentation. This directory contains comprehensive guides for versioning, branching, and release management.

## 📖 Documentation Index

### Quick Start
- **[Quick Start Guide](./QUICK_START_VERSIONING.md)** - Get started with versioning in 5 minutes
  - Common commands and workflows
  - Daily development tasks
  - Quick reference cheat sheet

### Core Documentation
- **[Versioning Strategy](./VERSIONING_STRATEGY.md)** - Complete versioning guide
  - Semantic versioning explained
  - Version tagging strategy
  - CI/CD pipeline details
  - Agent management strategy
  
- **[Branching Guide](./BRANCHING_GUIDE.md)** - Branch management workflows
  - Branch types and purposes
  - Detailed workflow examples
  - Pull request guidelines
  - Rollback procedures

- **[Git Helpers Guide](./GIT_HELPERS_GUIDE.md)** - Shell function reference
  - Complete function documentation
  - Installation and setup
  - Usage examples
  - Workflow automation

### Reference
- **[CHANGELOG](../CHANGELOG.md)** - Release history
- **[Makefile](../Makefile)** - Build and deployment commands

---

## 🎯 What You'll Learn

### For Developers
- How to create feature branches
- Semantic versioning best practices
- Commit message conventions
- Local development workflow
- Testing and deployment

### For SREs/DevOps
- CI/CD pipeline architecture
- Deployment automation
- Rollback strategies
- Monitoring and observability
- Version tracking in Kubernetes

### For Release Managers
- Release process workflow
- Version bumping procedures
- Changelog management
- GitHub release creation
- Tagging strategy

---

## 🚀 Quick Links

### Common Tasks

#### I want to...

**...develop a new feature**
1. Read: [Quick Start → Working on a Feature](./QUICK_START_VERSIONING.md#-working-on-a-feature)
2. Create branch: `gfeature my feature` (auto-adds date)
3. Follow conventional commits: `gcommit feat add my feature`
4. Complete and push: `gfeature-complete`
5. Create PR to `develop` on GitHub

**...release to production**
1. Read: [Quick Start → Releasing to Production](./QUICK_START_VERSIONING.md#-releasing-to-production)
2. Complete release: `grelease-complete 1.1.0`
3. Update CHANGELOG.md
4. Push: `grelease-push`
5. Create PR to `main` on GitHub

**...fix a critical bug in production**
1. Read: [Quick Start → Emergency Hotfix](./QUICK_START_VERSIONING.md#-emergency-hotfix)
2. Create hotfix: `ghotfix critical fix`
3. Fix and commit: `gcommit fix resolve critical bug`
4. Bump version: `gbump patch`
5. Push: `ghotfix-complete`
6. Create PR to `main` and remember to back-merge to `develop`

**...understand the CI/CD pipeline**
1. Read: [Versioning Strategy → CI/CD Pipeline](./VERSIONING_STRATEGY.md#-cicd-pipeline)
2. Check: `.github/workflows/` directory

**...manage versions**
1. Read: [Versioning Strategy → Version Management](./VERSIONING_STRATEGY.md#-semantic-versioning)
2. Use: `gversion`, `gbump <version>`, or `grelnotes`
3. See: [Git Helpers Guide](./GIT_HELPERS_GUIDE.md)

---

## 📋 Before You Start

### Prerequisites
- Git installed and configured
- Docker and Docker Compose
- kubectl and Flux CLI (for deployments)
- Node.js 24+ (for frontend)
- Go 1.23+ (for API)

### Required Knowledge
- Basic Git workflows
- Semantic versioning concepts
- Kubernetes basics (for deployments)
- Docker fundamentals

---

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     GitHub Repository                        │
│  ┌────────────┬────────────┬────────────┬──────────────┐   │
│  │   main     │  develop   │ feature/*  │  hotfix/*    │   │
│  │(production)│ (staging)  │            │              │   │
│  └─────┬──────┴─────┬──────┴─────┬──────┴──────┬───────┘   │
└────────┼────────────┼────────────┼─────────────┼───────────┘
         │            │            │             │
         │            │            │             │
    ┌────▼────┐  ┌────▼────┐  ┌────▼─────┐  ┌───▼──────┐
    │ CD Prod │  │CD Staging│ │   CI     │  │ CD Prod  │
    │Workflow │  │ Workflow │ │ Workflow │  │ Workflow │
    └────┬────┘  └────┬────┘  └────┬─────┘  └────┬─────┘
         │            │            │             │
         ▼            ▼            ▼             ▼
    ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐
    │  GHCR   │  │  GHCR   │  │  Tests  │  │  GHCR   │
    │ v1.0.0  │  │v1.0-beta│  │         │  │ v1.0.1  │
    └────┬────┘  └────┬────┘  └─────────┘  └────┬────┘
         │            │                          │
         ▼            ▼                          ▼
    ┌─────────────────────────────────────────────────┐
    │         Kubernetes Cluster (Homelab)            │
    │  ┌───────────────┐      ┌──────────────────┐   │
    │  │  Production   │      │     Staging      │   │
    │  │  Namespace    │      │    Namespace     │   │
    │  │  (homepage)   │      │   (homepage-*)   │   │
    │  └───────────────┘      └──────────────────┘   │
    │         Flux GitOps                             │
    └─────────────────────────────────────────────────┘
```

---

## 📊 Branching Model

```
main (production)
  ├─── v1.0.0 ────────────────────── v1.1.0 ────── v1.1.1
  │                                      │           │
develop (staging)                        │           │
  ├─── feature/auth ────┬─── (merge)   │           │
  ├─── feature/ui ──────┘               │           │
  └─── bugfix/api ──────────────────────┘           │
                                                     │
hotfix/security ───────────────────────────────────┘
```

---

## 🔄 Release Workflow

### Standard Release (Feature → Staging → Production)
```
1. Feature Development
   └─> feature/* branches → develop

2. Integration Testing
   └─> develop → staging environment (auto-deploy)

3. Release Preparation
   └─> release/* branch from develop
       • Bump version
       • Update CHANGELOG
       • Final testing

4. Production Release
   └─> release/* → main → production (auto-deploy)
       • Create GitHub release
       • Tag version
       • Back-merge to develop
```

### Hotfix (Emergency Production Fix)
```
1. Create hotfix from main
   └─> hotfix/* from main

2. Fix and test
   └─> Apply fix, bump patch version

3. Deploy to production
   └─> hotfix/* → main → production (auto-deploy)

4. Back-merge to develop
   └─> main → develop (keep develop in sync)
```

---

## 🛠️ Tools & Scripts

### Makefile Commands
```bash
make version              # Show current version
make version-bump         # Bump version
make build-push-version   # Build and push with version
make release-notes        # Generate release notes
make deploy              # Deploy to environment
make rollback            # Rollback deployment
```

### Version Manager Script
```bash
./scripts/version-manager.sh show         # Current version
./scripts/version-manager.sh bump 1.0.0   # Manual bump
./scripts/version-manager.sh bump-major   # Auto bump major
./scripts/version-manager.sh bump-minor   # Auto bump minor
./scripts/version-manager.sh bump-patch   # Auto bump patch
```

---

## 📝 Commit Message Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(scope): add new feature
fix(scope): resolve bug
docs: update documentation
chore: bump version
refactor: restructure code
test: add tests
ci: update CI/CD
```

Examples:
```bash
git commit -m "feat(api): add user authentication endpoint"
git commit -m "fix(frontend): resolve memory leak in chatbot"
git commit -m "chore: bump version to 1.1.0"
```

---

## 🔢 Version Examples

| Version | Type | Description |
|---------|------|-------------|
| `1.0.0` | Stable | Production release |
| `1.1.0` | Minor | New features (backwards-compatible) |
| `1.1.1` | Patch | Bug fixes |
| `2.0.0` | Major | Breaking changes |
| `1.2.0-beta.1` | Pre-release | Beta release (staging) |
| `1.2.0-rc.1` | Pre-release | Release candidate |
| `1.2.0-dev.abc123` | Development | Development build |

---

## 🚨 Common Pitfalls

### ❌ Don't
- Push directly to `main` or `develop`
- Use vague commit messages ("fix", "update", "wip")
- Skip version bumps on releases
- Forget to back-merge hotfixes to develop
- Use `:latest` tags without semantic versions

### ✅ Do
- Always work on feature branches
- Use conventional commit messages
- Test in staging before production
- Update CHANGELOG for releases
- Tag all production releases
- Back-merge hotfixes to develop

---

## 📚 Additional Resources

### External Links
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Trunk-Based Development](https://trunkbaseddevelopment.com/)
- [Google SRE Book - Release Engineering](https://sre.google/sre-book/release-engineering/)

### Internal Links
- [GitHub Repository](https://github.com/brunovlucena/homelab)
- [GitHub Actions Workflows](../.github/workflows/)
- [Helm Chart](../chart/)
- [API Source](../api/)
- [Frontend Source](../frontend/)

---

## 🤝 Contributing

When contributing to this documentation:

1. Keep it concise and actionable
2. Use examples and code snippets
3. Update the index when adding new docs
4. Follow markdown best practices
5. Test all commands before documenting

---

## 📧 Support

- **Issues**: [GitHub Issues](https://github.com/brunovlucena/homelab/issues)
- **Questions**: Create a discussion on GitHub
- **Email**: bruno@lucena.cloud

---

## 📄 License

This documentation is part of the Homepage project and follows the same license.

---

**Last Updated**: 2025-10-16  
**Version**: 1.0.0

