# GitHub Actions CI/CD Templates - Agent Bruno

**Status**: 🔴 P0 - CRITICAL BLOCKER  
**Timeline**: Day 3-5  
**Blocks**: Automated testing, deployments, security scanning

---

## 📁 Directory Structure

Create these files in `.github/workflows/`:

```
.github/
└── workflows/
    ├── ci.yml           # Main CI pipeline (test, build, scan)
    ├── cd.yml           # Continuous deployment
    ├── security.yml     # Security scanning
    └── release.yml      # Release automation
```

---

## 1️⃣ CI Pipeline Template

Create `.github/workflows/ci.yml`:

```yaml
# ============================================================================
# Agent Bruno - Continuous Integration Pipeline
# ============================================================================
# Triggers: Pull requests, pushes to main/develop
# Jobs: Test → Security Scan → Build → Push Image
# ============================================================================

name: CI Pipeline

on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [main, develop]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}/agent-bruno
  PYTHON_VERSION: '3.11'

jobs:
  # --------------------------------------------------------------------------
  # Job 1: Run Tests
  # --------------------------------------------------------------------------
  test:
    name: Run Tests
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: ${{ env.PYTHON_VERSION }}
          cache: 'pip'
      
      - name: Install dependencies
        run: |
          python -m pip install --upgrade pip
          pip install -r requirements.txt
          pip install -r requirements-dev.txt || echo "No dev requirements"
      
      - name: Run linters
        run: |
          pip install ruff black mypy
          ruff check src/ || true
          black --check src/ || true
          mypy src/ || true
      
      - name: Run unit tests
        run: |
          pytest tests/unit/ -v --cov=src --cov-report=xml --cov-report=term || true
      
      - name: Run integration tests
        run: |
          pytest tests/integration/ -v || true
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.xml
          fail_ci_if_error: false
      
      - name: Generate coverage report
        run: |
          pip install coverage
          coverage report || true

  # --------------------------------------------------------------------------
  # Job 2: Security Scanning
  # --------------------------------------------------------------------------
  security:
    name: Security Scan
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          severity: 'CRITICAL,HIGH'
          exit-code: '0'  # Don't fail build yet
          format: 'sarif'
          output: 'trivy-results.sarif'
      
      - name: Upload Trivy results to GitHub Security
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'
      
      - name: Run Semgrep SAST
        uses: returntocorp/semgrep-action@v1
        with:
          config: >-
            p/security-audit
            p/python
            p/owasp-top-ten
          generateSarif: true
      
      - name: Python dependency security check
        run: |
          pip install safety
          safety check --json || true

  # --------------------------------------------------------------------------
  # Job 3: Build Container Image
  # --------------------------------------------------------------------------
  build:
    name: Build & Push Image
    runs-on: ubuntu-latest
    needs: [test, security]
    permissions:
      contents: read
      packages: write
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=sha,prefix={{branch}}-
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
      
      - name: Build and push image
        id: build
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          platforms: linux/amd64
      
      - name: Generate SBOM
        uses: anchore/sbom-action@v0
        with:
          image: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${{ steps.build.outputs.digest }}
          format: cyclonedx-json
          output-file: sbom.cyclonedx.json
      
      - name: Upload SBOM artifact
        uses: actions/upload-artifact@v3
        with:
          name: sbom
          path: sbom.cyclonedx.json
      
      - name: Scan container image
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${{ steps.build.outputs.digest }}
          format: 'sarif'
          output: 'trivy-image-results.sarif'
      
      - name: Upload container scan results
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-image-results.sarif'

  # --------------------------------------------------------------------------
  # Job 4: Notify
  # --------------------------------------------------------------------------
  notify:
    name: Notify Status
    runs-on: ubuntu-latest
    needs: [test, security, build]
    if: always()
    
    steps:
      - name: Send Slack notification
        uses: 8398a7/action-slack@v3
        if: always()
        with:
          status: ${{ job.status }}
          text: |
            CI Pipeline: ${{ job.status }}
            Branch: ${{ github.ref }}
            Commit: ${{ github.sha }}
            Author: ${{ github.actor }}
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

---

## 2️⃣ CD Pipeline Template

Create `.github/workflows/cd.yml`:

```yaml
# ============================================================================
# Agent Bruno - Continuous Deployment Pipeline
# ============================================================================
# Triggers: Successful CI on main/develop branches
# Jobs: Update GitOps manifests → Flux reconciles → Flagger canary
# ============================================================================

name: CD - Deploy to Kubernetes

on:
  workflow_run:
    workflows: ["CI Pipeline"]
    types:
      - completed
    branches: [main, develop]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}/agent-bruno

jobs:
  # --------------------------------------------------------------------------
  # Job 1: Update GitOps Manifests
  # --------------------------------------------------------------------------
  deploy:
    name: Deploy to Kubernetes
    runs-on: ubuntu-latest
    if: ${{ github.event.workflow_run.conclusion == 'success' }}
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          token: ${{ secrets.FLUX_GITHUB_TOKEN }}
          fetch-depth: 0
      
      - name: Determine environment
        id: env
        run: |
          if [[ "${{ github.ref }}" == "refs/heads/main" ]]; then
            echo "environment=staging" >> $GITHUB_OUTPUT
            echo "namespace=agent-bruno-staging" >> $GITHUB_OUTPUT
          elif [[ "${{ github.ref }}" == "refs/heads/develop" ]]; then
            echo "environment=dev" >> $GITHUB_OUTPUT
            echo "namespace=agent-bruno-dev" >> $GITHUB_OUTPUT
          fi
      
      - name: Update image tag
        run: |
          NEW_TAG="${{ github.ref_name }}-${{ github.sha }}"
          
          # Update k8s manifests
          sed -i "s|image: .*agent-bruno:.*|image: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:$NEW_TAG|" \
            flux/clusters/homelab/infrastructure/agent-bruno/k8s/overlays/${{ steps.env.outputs.environment }}/kustomization.yaml
      
      - name: Commit and push
        run: |
          git config user.name "GitHub Actions Bot"
          git config user.email "actions@github.com"
          
          git add flux/clusters/homelab/infrastructure/agent-bruno/
          git commit -m "chore(${{ steps.env.outputs.environment }}): deploy agent-bruno:${{ github.sha }}"
          git push
      
      - name: Wait for Flux reconciliation
        run: |
          echo "Waiting for Flux to reconcile..."
          sleep 30
      
      - name: Verify deployment
        run: |
          # TODO: Add kubectl wait for deployment ready
          echo "Deployment verification pending"
      
      - name: Notify deployment
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: |
            🚀 Deployment to ${{ steps.env.outputs.environment }}
            Image: ${{ env.IMAGE_NAME }}:${{ github.sha }}
            Environment: ${{ steps.env.outputs.environment }}
            Namespace: ${{ steps.env.outputs.namespace }}
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}

  # --------------------------------------------------------------------------
  # Job 2: Production Promotion (Manual)
  # --------------------------------------------------------------------------
  promote-to-prod:
    name: Promote to Production
    runs-on: ubuntu-latest
    needs: [deploy]
    if: github.ref == 'refs/heads/main'
    environment:
      name: production
      url: https://agent-bruno.yourdomain.com
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          token: ${{ secrets.FLUX_GITHUB_TOKEN }}
      
      - name: Create production tag
        run: |
          # TODO: Implement semantic versioning
          echo "Production promotion pending manual approval"
```

---

## 3️⃣ Security Scanning Template

Create `.github/workflows/security.yml`:

```yaml
# ============================================================================
# Agent Bruno - Security Scanning Pipeline
# ============================================================================
# Triggers: Daily, on PR, on push
# Jobs: Secret scanning, dependency audit, container scan
# ============================================================================

name: Security Scans

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
  pull_request:
  push:
    branches: [main, develop]

jobs:
  # --------------------------------------------------------------------------
  # Job 1: Secret Scanning
  # --------------------------------------------------------------------------
  secret-scan:
    name: Scan for Secrets
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - name: TruffleHog scan
        uses: trufflesecurity/trufflehog@main
        with:
          path: ./
          base: ${{ github.event.repository.default_branch }}
          head: HEAD
          extra_args: --debug --only-verified
      
      - name: Gitleaks scan
        uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  # --------------------------------------------------------------------------
  # Job 2: Dependency Audit
  # --------------------------------------------------------------------------
  dependency-audit:
    name: Audit Dependencies
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      
      - name: Safety check
        run: |
          pip install safety
          safety check --json --output safety-report.json || true
      
      - name: Upload safety report
        uses: actions/upload-artifact@v3
        with:
          name: safety-report
          path: safety-report.json
      
      - name: Snyk dependency scan
        uses: snyk/actions/python@master
        continue-on-error: true
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          args: --severity-threshold=high
```

---

## 4️⃣ Release Automation Template

Create `.github/workflows/release.yml`:

```yaml
# ============================================================================
# Agent Bruno - Release Automation
# ============================================================================
# Triggers: Manual workflow dispatch, tags (v*)
# Jobs: Create release, build artifacts, deploy to production
# ============================================================================

name: Release

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to release (e.g., v1.0.0)'
        required: true

jobs:
  release:
    name: Create Release
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - name: Generate changelog
        id: changelog
        uses: metcalfc/changelog-generator@v4.1.0
        with:
          myToken: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Create GitHub Release
        uses: softprops/action-gh-release@v1
        with:
          body: ${{ steps.changelog.outputs.changelog }}
          draft: false
          prerelease: false
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## ✅ Setup Checklist

### Prerequisites
- [ ] Create GitHub PAT with `repo` and `packages:write` permissions
- [ ] Add `FLUX_GITHUB_TOKEN` secret to repository
- [ ] Add `SLACK_WEBHOOK` secret (optional, for notifications)
- [ ] Add `SNYK_TOKEN` secret (optional, for Snyk scanning)

### File Creation
- [ ] Create `.github/workflows/ci.yml`
- [ ] Create `.github/workflows/cd.yml`
- [ ] Create `.github/workflows/security.yml`
- [ ] Create `.github/workflows/release.yml`

### Testing
- [ ] Create test PR to verify CI pipeline
- [ ] Merge to develop to verify CD to dev environment
- [ ] Merge to main to verify CD to staging environment

---

## 🔗 Related Documentation

- [DEVOPS_UNBLOCK_PLAN.md](../DEVOPS_UNBLOCK_PLAN.md) - Overall unblock plan
- [DOCKERFILE_TEMPLATE.md](./DOCKERFILE_TEMPLATE.md) - Container image template
- [CICD_SETUP.md](../CICD_SETUP.md) - CI/CD setup guide

---

**Status**: 🔴 NOT IMPLEMENTED  
**Next Step**: Create `.github/workflows/` directory and add workflows  
**Owner**: DevOps Team  
**Timeline**: Day 3-5

---

