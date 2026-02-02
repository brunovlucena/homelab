# 🔧 GitHub Actions Workflows

This directory contains GitHub Actions workflows for the homelab repository, organized by type.

## 📁 Naming Convention

| Prefix | Purpose | Example |
|--------|---------|---------|
| `_reusable-` | Shared reusable workflows | `_reusable-python-ci.yml` |
| `agent-` | AI agent CI/CD pipelines | `agent-bruno-ci-cd.yml` |
| `operator-` | Kubernetes operator CI/CD | `operator-knative-lambda.yml` |
| `infra-` | Infrastructure components | `infra-homepage.yml` |
| `ci-` | CI-only validation workflows | `ci-dependabot-validation.yml` |
| `tools-` | Custom tool/image builds | `tools-custom-images.yml` |

## 🔄 Reusable Workflows

These workflows are designed to be called by other workflows, providing standardized CI/CD patterns.

### Core Reusable Workflows

| Workflow | Purpose | Inputs |
|----------|---------|--------|
| `_reusable-detect-changes.yml` | Detects file changes, extracts version | `working-directory`, `version-file` |
| `_reusable-python-ci.yml` | Python linting (Ruff), type checking (MyPy), testing (Pytest) | `working-directory`, `python-version` |
| `_reusable-go-ci.yml` | Go linting, formatting, testing | `working-directory`, `go-version-file` |
| `_reusable-nodejs-ci.yml` | Node.js linting, type checking, building | `working-directory`, `node-version` |
| `_reusable-docker-build.yml` | Multi-platform Docker builds with caching | `image-name`, `dockerfile`, `context` |
| `_reusable-security-scan.yml` | Trivy vulnerability scanning, gitleaks | `scan-path`, `scan-type` |

### Meta Workflows

| Workflow | Purpose |
|----------|---------|
| `_reusable-agent-cicd.yml` | Complete CI/CD for Python AI agents (combines detect + lint + test + build + security) |

## 🤖 AI Agent Workflows

All AI agents use the reusable `_reusable-agent-cicd.yml` meta-workflow for consistent CI/CD.

### Production Agents

| Workflow | Agent | Tech Stack |
|----------|-------|------------|
| `agent-bruno-ci-cd.yml` | Personal AI Assistant | Python |
| `agent-medical-ci-cd.yml` | Medical AI | Python |
| `agent-restaurant-ci-cd.yml` | Restaurant Management | Python |
| `agent-pos-edge-ci-cd.yml` | POS Edge Computing | Python |
| `agent-store-multibrands-ci-cd.yml` | Multi-brand Store | Python |
| `agent-tools-ci-cd.yml` | AI Tool Suite | Python |

### Security Agents

| Workflow | Agent | Purpose |
|----------|-------|---------|
| `agent-blueteam-ci-cd.yml` | Blue Team | Defensive security |
| `agent-redteam-ci-cd.yml` | Red Team | Offensive security |
| `agent-devsecops-ci-cd.yml` | DevSecOps | Security automation |
| `agent-contracts-ci-cd.yml` | Contract Security | Multi-service security scanning |

### Multi-Service Agents

| Workflow | Agent | Services |
|----------|-------|----------|
| `agent-chat-ci-cd.yml` | Chat System | messaging-hub, command-center, voice, media, location |
| `agent-webinterface-ci-cd.yml` | Web Interface | Next.js frontend |

### Auditor Demos

| Workflow | Scenario |
|----------|----------|
| `agent-auditor-ci.yml` | Base auditor |
| `agent-auditor-cost-tracking.yml` | Cost optimization demo |
| `agent-auditor-cross-team-handoff.yml` | Team handoff demo |
| `agent-auditor-incident-response.yml` | Incident response demo |
| `agent-auditor-ml-infrastructure.yml` | ML infrastructure demo |
| `agent-auditor-mobile-dev.yml` | Mobile development demo |
| `agent-auditor-security-sprint.yml` | Security sprint demo |

## ⚙️ Operator Workflows

| Workflow | Operator | Tech Stack |
|----------|----------|------------|
| `operator-knative-lambda.yml` | Knative Lambda | Go + Kubernetes |
| `operator-cloudflare-tunnel.yml` | Cloudflare Tunnel | Go + Kubernetes |

## 🏗️ Infrastructure Workflows

| Workflow | Component | Tech Stack |
|----------|-----------|------------|
| `infra-homepage.yml` | Homepage | Go (API) + Next.js (Frontend) |
| `infra-garak.yml` | Garak | Python + FastAPI |

## 🔧 Tool Workflows

| Workflow | Purpose |
|----------|---------|
| `tools-custom-images.yml` | Build kubectl, linkerd-cli, cloudflare-warp images |
| `ci-dependabot-validation.yml` | Validate Dependabot PRs |

## 🚀 Quick Start

### For New Python AI Agents

```yaml
name: 🤖 My Agent - CI/CD

on:
  push:
    branches: [main, feature/*, bugfix/*]
    paths: ['flux/ai/my-agent/**']
  pull_request:
    paths: ['flux/ai/my-agent/**']
  workflow_dispatch:

permissions:
  contents: write
  packages: write
  security-events: write
  attestations: write
  id-token: write

jobs:
  ci-cd:
    uses: ./.github/workflows/_reusable-agent-cicd.yml
    with:
      agent-name: my-agent
      working-directory: flux/ai/my-agent
      image-name: my-agent/app
      dockerfile: src/app/Dockerfile
    secrets: inherit
```

### For Go Operators

```yaml
jobs:
  lint:
    uses: ./.github/workflows/_reusable-go-ci.yml
    with:
      working-directory: flux/infrastructure/my-operator/src

  build:
    needs: [lint]
    uses: ./.github/workflows/_reusable-docker-build.yml
    with:
      image-name: my-operator
      dockerfile: src/Dockerfile
      context: flux/infrastructure/my-operator/src
```

### For Node.js Projects

```yaml
jobs:
  lint-and-build:
    uses: ./.github/workflows/_reusable-nodejs-ci.yml
    with:
      working-directory: flux/infrastructure/my-app
      node-version: '20'
```

## 📊 Configuration Reference

### `_reusable-agent-cicd.yml` Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `agent-name` | ✅ | - | Agent identifier |
| `working-directory` | ✅ | - | Path to agent directory |
| `image-name` | ✅ | - | Docker image name |
| `dockerfile` | ✅ | - | Path to Dockerfile |
| `python-version` | ❌ | `3.12` | Python version |
| `src-directory` | ❌ | `src` | Source directory |
| `run-tests` | ❌ | `true` | Run pytest |
| `run-lint` | ❌ | `true` | Run ruff/mypy |
| `copy-shared-lib` | ❌ | `true` | Copy shared-lib |
| `platforms` | ❌ | `linux/amd64,linux/arm64` | Build platforms |

### `_reusable-go-ci.yml` Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `working-directory` | ✅ | - | Path to Go module |
| `go-version-file` | ❌ | `go.mod` | Go version file |
| `run-tests` | ❌ | `true` | Run go test |
| `run-lint` | ❌ | `true` | Run golangci-lint |
| `golangci-lint-version` | ❌ | `latest` | Lint version |

### `_reusable-nodejs-ci.yml` Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `working-directory` | ✅ | - | Path to package.json |
| `node-version` | ❌ | `20` | Node.js version |
| `package-manager` | ❌ | `npm` | npm/yarn/pnpm |
| `run-tests` | ❌ | `true` | Run tests |
| `run-lint` | ❌ | `true` | Run ESLint |
| `run-type-check` | ❌ | `true` | Run tsc |
| `run-build` | ❌ | `true` | Run build |

### `_reusable-docker-build.yml` Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `image-name` | ✅ | - | Image name |
| `dockerfile` | ✅ | - | Dockerfile path |
| `context` | ✅ | - | Build context |
| `platforms` | ❌ | `linux/amd64,linux/arm64` | Target platforms |
| `push` | ❌ | `true` | Push to registry |
| `version` | ❌ | - | Version tag |

## 📈 Benefits of Reusable Workflows

1. **Consistency** - All agents follow the same CI/CD patterns
2. **Maintainability** - Update once, apply everywhere
3. **Reduced Duplication** - ~70% less YAML per workflow
4. **Security** - Centralized security scanning configuration
5. **Discoverability** - Organized naming convention
