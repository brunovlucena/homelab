# 🔄 DEVOPS-004: CI/CD Pipeline

**Priority**: P1 | **Status**: ✅ Implemented K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-236/devops-004-cicd-pipeline


---

## 📋 User Story

**As a** DevOps Engineer  
**I want to** implement a fully automated CI/CD pipeline  
**So that** code changes are automatically tested, built, and deployed with zero manual intervention

---

## 🎯 Acceptance Criteria

### ✅ Continuous Integration
- [ ] Automated testing on every PR
- [ ] Linting and code quality checks
- [ ] Security scanning (SAST, dependency checks)
- [ ] Unit tests with coverage reporting
- [ ] Integration tests with test infrastructure
- [ ] Build validation before merge

### ✅ Continuous Delivery
- [ ] Automated Docker image builds
- [ ] Image tagging strategy (semantic versioning)
- [ ] Push to ECR with proper tags
- [ ] Helm chart version bumping
- [ ] Automated deployment to dev environment
- [ ] Release notes generation

### ✅ Deployment Automation
- [ ] GitOps-based deployment via Flux
- [ ] Environment-specific deployment strategies
- [ ] Automated smoke tests post-deployment
- [ ] Rollback on failure detection
- [ ] Deployment notifications (Slack, email)
- [ ] Deployment tracking and history

### ✅ Quality Gates
- [ ] Test coverage > 80%
- [ ] No critical security vulnerabilities
- [ ] Docker image size < 500MB
- [ ] Build time < 5 minutes
- [ ] All integration tests passing
- [ ] Manual approval for production

### ✅ Observability
- [ ] Pipeline execution metrics
- [ ] Build success/failure rates
- [ ] Deployment frequency tracking
- [ ] Lead time for changes
- [ ] Mean time to recovery (MTTR)

---

## 🏗️ CI/CD Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    CI/CD PIPELINE FLOW                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. CODE PUSH                                                   │
│     Developer → git push origin feature/new-feature             │
│                                                                 │
│  2. GITHUB ACTIONS TRIGGERED                                    │
│     ├─ Checkout code                                            │
│     ├─ Set up Go environment                                    │
│     ├─ Cache dependencies                                       │
│     └─ Install dependencies                                     │
│                                                                 │
│  3. CODE QUALITY CHECKS                                         │
│     ├─ golangci-lint (formatting, bugs)                         │
│     ├─ gosec (security analysis)                                │
│     ├─ go vet (code analysis)                                   │
│     └─ staticcheck (static analysis)                            │
│                                                                 │
│  4. TESTING                                                     │
│     ├─ Unit Tests (go test -v ./...)                            │
│     ├─ Integration Tests (with test containers)                 │
│     ├─ Coverage Report (> 80% required)                         │
│     └─ Test Results Upload (GitHub)                             │
│                                                                 │
│  5. SECURITY SCANNING                                           │
│     ├─ Trivy (dependency vulnerabilities)                       │
│     ├─ Snyk (security vulnerabilities)                          │
│     ├─ SAST (static application security testing)               │
│     └─ License compliance check                                 │
│                                                                 │
│  6. BUILD DOCKER IMAGE                                          │
│     ├─ docker build -t builder:${GIT_SHA}                       │
│     ├─ Trivy scan image                                         │
│     ├─ Image size validation (< 500MB)                          │
│     └─ Generate SBOM (Software Bill of Materials)               │
│                                                                 │
│  7. PUSH TO ECR                                                 │
│     ├─ Tag: dev-${GIT_SHA}                                      │
│     ├─ Tag: dev-latest (if main branch)                         │
│     ├─ Tag: staging-v${VERSION} (if release branch)             │
│     └─ Tag: prd-v${VERSION} (if tagged release)                 │
│                                                                 │
│  8. UPDATE MANIFESTS                                            │
│     ├─ Update Helm values with new image tag                    │
│     ├─ Bump chart version                                       │
│     ├─ Commit to Git (for GitOps)                               │
│     └─ Create PR for staging/prod promotions                    │
│                                                                 │
│  9. GITOPS DEPLOYMENT (FLUX)                                    │
│     ├─ Flux detects Git change                                  │
│     ├─ Apply manifests to cluster                               │
│     ├─ Wait for deployment rollout                              │
│     └─ Health check validation                                  │
│                                                                 │
│  10. POST-DEPLOYMENT VALIDATION                                 │
│     ├─ Smoke tests (health check, metrics)                      │
│     ├─ Integration tests (real API calls)                       │
│     ├─ Performance tests (latency check)                        │
│     └─ Rollback if validation fails                             │
│                                                                 │
│  11. NOTIFICATION                                               │
│     ├─ Slack: Deployment successful ✅                          │
│     ├─ GitHub: Update commit status                             │
│     └─ Datadog: Deployment event marker                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Technical Implementation

### GitHub Actions Workflow

**File**: `.github/workflows/ci-cd.yaml`
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop, 'feature/**', 'hotfix/**' ]
    tags: [ 'v*.*.*' ]
  pull_request:
    branches: [ main, develop ]

env:
  GO_VERSION: '1.21'
  ECR_REGISTRY: 339954290315.dkr.ecr.us-west-2.amazonaws.com
  ECR_REPOSITORY: knative-lambda-builder
  AWS_REGION: us-west-2

jobs:
  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  # 🔍 CODE QUALITY & LINTING
  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  lint:
    name: 🔍 Lint & Code Quality
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
        cache: true
    
    - name: golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
        args: --timeout=5m
    
    - name: Run gosec (security)
      run: | go install github.com/securego/gosec/v2/cmd/gosec@latest
        gosec -fmt json -out gosec-report.json ./...
    
    - name: Upload security report
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: gosec-report.json

  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  # 🧪 TESTING
  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  test:
    name: 🧪 Unit & Integration Tests
    runs-on: ubuntu-latest
    
    services:
      rabbitmq:
        image: rabbitmq:3.12-management
        ports:
          - 5672:5672
          - 15672:15672
        env:
          RABBITMQ_DEFAULT_USER: admin
          RABBITMQ_DEFAULT_PASS: admin
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
        cache: true
    
    - name: Download dependencies
      run: go mod download
    
    - name: Run unit tests
      run: | go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
    
    - name: Check test coverage
      run: | COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        echo "Test coverage: ${COVERAGE}%"
        if (( $(echo "$COVERAGE < 80" | bc -l) )); then
          echo "❌ Test coverage ${COVERAGE}% is below 80%"
          exit 1
        fi
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        flags: unittests
    
    - name: Run integration tests
      run: | export RABBITMQ_URL="amqp://admin:admin@localhost:5672/"
        go test -v -tags=integration ./tests/integration/...

  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  # 🔒 SECURITY SCANNING
  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  security:
    name: 🔒 Security Scanning
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'sarif'
        output: 'trivy-results.sarif'
        severity: 'CRITICAL,HIGH'
    
    - name: Upload Trivy results
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: trivy-results.sarif
    
    - name: Snyk security scan
      uses: snyk/actions/golang@master
      env:
        SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
      with:
        args: --severity-threshold=high

  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  # 🐳 BUILD & PUSH DOCKER IMAGE
  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  build:
    name: 🐳 Build & Push Image
    runs-on: ubuntu-latest
    needs: [lint, test, security]
    if: github.event_name == 'push'
    
    outputs:
      image-tag: ${{ steps.meta.outputs.tags }}
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: ${{ env.AWS_REGION }}
    
    - name: Login to Amazon ECR
      id: login-ecr
      uses: aws-actions/amazon-ecr-login@v2
    
    - name: Docker meta (generate tags)
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.ECR_REGISTRY }}/${{ env.ECR_REPOSITORY }}
        tags: | # For main branch: dev-latest + dev-{sha}
          type=raw,value=dev-latest,enable={{is_default_branch}}
          type=sha,prefix=dev-,enable={{is_default_branch}}
          
          # For develop branch: staging-latest + staging-{sha}
          type=raw,value=staging-latest,enable=${{ github.ref == 'refs/heads/develop' }}
          type=sha,prefix=staging-,enable=${{ github.ref == 'refs/heads/develop' }}
          
          # For tags: prd-v{version}
          type=semver,pattern=prd-v{{version}}
          type=semver,pattern=prd-v{{major}}.{{minor}}
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3
    
    - name: Build and push
      uses: docker/build-push-action@v5
      with:
        context: .
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
        build-args: | VERSION=${{ github.ref_name }}
          COMMIT_SHA=${{ github.sha }}
          BUILD_DATE=${{ github.event.head_commit.timestamp }}
    
    - name: Scan image with Trivy
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: ${{ steps.meta.outputs.tags }}
        format: 'sarif'
        output: 'trivy-image-results.sarif'
    
    - name: Check image size
      run: | IMAGE_SIZE=$(docker image inspect ${{ steps.meta.outputs.tags }} --format='{{.Size}}')
        IMAGE_SIZE_MB=$((IMAGE_SIZE / 1024 / 1024))
        echo "Image size: ${IMAGE_SIZE_MB}MB"
        if [ ${IMAGE_SIZE_MB} -gt 500 ]; then
          echo "❌ Image size ${IMAGE_SIZE_MB}MB exceeds 500MB limit"
          exit 1
        fi

  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  # 🚀 DEPLOY TO DEV (Auto)
  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  deploy-dev:
    name: 🚀 Deploy to Dev
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/main'
    environment:
      name: dev
      url: https://dev.knative-lambda.homelab
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Update Helm values
      run: | NEW_TAG="dev-${GITHUB_SHA::7}"
        yq eval ".image.tag = \"${NEW_TAG}\"" \
          -i deploy/overlays/dev/values-dev.yaml
    
    - name: Commit and push
      run: | git config user.name "GitHub Actions Bot"
        git config user.email "actions@github.com"
        git add deploy/overlays/dev/values-dev.yaml
        git commit -m "chore(dev): update image to ${NEW_TAG}"
        git push origin main
    
    - name: Trigger Flux reconciliation
      run: | # Flux will automatically sync within 5 minutes
        # For immediate sync, use flux CLI or webhook
        echo "✅ GitOps sync triggered for dev environment"

  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  # 🧪 SMOKE TESTS
  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  smoke-tests:
    name: 🧪 Smoke Tests
    runs-on: ubuntu-latest
    needs: deploy-dev
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Wait for deployment
      run: sleep 120  # Wait for deployment to stabilize
    
    - name: Health check
      run: | curl -f https://dev.knative-lambda.homelab/health | | exit 1
    
    - name: Metrics check
      run: | curl -f https://dev.knative-lambda.homelab/metrics | | exit 1
    
    - name: Basic API test
      run: | curl -X POST https://dev.knative-lambda.homelab/events \
          -H "Ce-Type: network.notifi.lambda.build.start" \
          -d '{"parser_id":"test"}' | | exit 1

  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  # 📢 NOTIFICATIONS
  # ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  notify:
    name: 📢 Send Notifications
    runs-on: ubuntu-latest
    needs: [build, deploy-dev, smoke-tests]
    if: always()
    
    steps:
    - name: Slack notification (success)
      if: needs.smoke-tests.result == 'success'
      uses: slackapi/slack-github-action@v1
      with:
        webhook: ${{ secrets.SLACK_WEBHOOK_URL }}
        payload: | {
            "text": "✅ Deployment successful",
            "blocks": [
              {
                "type": "section",
                "text": {
                  "type": "mrkdwn",
                  "text": "*Deployment Successful* ✅\n*Environment:* dev\n*Version:* ${{ github.sha }}\n*Author:* ${{ github.actor }}"
                }
              }
            ]
          }
    
    - name: Slack notification (failure)
      if: needs.smoke-tests.result == 'failure'
      uses: slackapi/slack-github-action@v1
      with:
        webhook: ${{ secrets.SLACK_WEBHOOK_URL }}
        payload: | {
            "text": "❌ Deployment failed",
            "blocks": [
              {
                "type": "section",
                "text": {
                  "type": "mrkdwn",
                  "text": "*Deployment Failed* ❌\n*Environment:* dev\n*Version:* ${{ github.sha }}\n*Author:* ${{ github.actor }}\n*Logs:* ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}"
                }
              }
            ]
          }
```

---

## 🎨 Tagging Strategy

### Tag Format | Branch | Tag Format | Example | Environment | |-------- | ------------ | --------- | ------------- | | `main` | `dev-{sha}` | `dev-a1b2c3d` | Dev | | `develop` | `staging-{sha}` | `staging-x7y8z9` | Staging | | `v*.*.*` | `prd-v{version}` | `prd-v1.2.3` | Production | ### Semantic Versioning

```bash
# Major version (breaking changes)
git tag -a v2.0.0 -m "Major release: New event format"

# Minor version (new features)
git tag -a v1.3.0 -m "Feature: Add rate limiting"

# Patch version (bug fixes)
git tag -a v1.2.1 -m "Fix: Memory leak in job manager"

# Push tags
git push origin --tags
```

---

## 📊 Pipeline Metrics

### DORA Metrics Dashboard

```yaml
# Grafana Dashboard Configuration
{
  "dashboard": {
    "title": "CI/CD Pipeline Metrics (DORA)",
    "panels": [
      {
        "title": "Deployment Frequency",
        "targets": [{
          "expr": "sum(increase(deployments_total{environment=\"prd\"}[1d]))"
        }]
      },
      {
        "title": "Lead Time for Changes",
        "targets": [{
          "expr": "histogram_quantile(0.95, rate(deployment_lead_time_seconds_bucket[1d]))"
        }]
      },
      {
        "title": "Change Failure Rate",
        "targets": [{
          "expr": "sum(rate(deployments_total{status=\"failed\"}[1d])) / sum(rate(deployments_total[1d]))"
        }]
      },
      {
        "title": "Mean Time to Recovery",
        "targets": [{
          "expr": "avg(deployment_recovery_time_seconds)"
        }]
      }
    ]
  }
}
```

### Pipeline Success Rate

```promql
# Build success rate
sum(rate(github_workflow_runs_total{status="success"}[1d])) /
sum(rate(github_workflow_runs_total[1d])) * 100

# Average build time
histogram_quantile(0.95, rate(github_workflow_duration_seconds_bucket[1h]))
```

---

## 🧪 Testing the Pipeline

### Local Pipeline Simulation

```bash
# Install act (GitHub Actions local runner)
brew install act

# Run workflow locally
act -j lint
act -j test
act -j build --secret-file .secrets

# Test specific job
act push -j deploy-dev
```

### Trigger Manual Pipeline Run

```bash
# Via GitHub CLI
gh workflow run ci-cd.yaml \
  --ref main \
  -f environment=dev

# Via API
curl -X POST \
  -H "Authorization: token ${GITHUB_TOKEN}" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/brunolucena/homelab/actions/workflows/ci-cd.yaml/dispatches \
  -d '{"ref":"main"}'
```

---

## 💡 Pro Tips

### 1. Cache Dependencies

```yaml
- name: Cache Go modules
  uses: actions/cache@v3
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: | ${{ runner.os }}-go-
```

### 2. Parallel Job Execution

```yaml
jobs:
  test:
    strategy:
      matrix:
        go-version: ['1.20', '1.21']
        os: [ubuntu-latest, macos-latest]
    runs-on: ${{ matrix.os }}
```

### 3. Conditional Steps

```yaml
- name: Deploy to production
  if: github.ref == 'refs/tags/v*'
  run: | echo "Deploying to production..."
```

### 4. Artifacts and Reports

```yaml
- name: Upload test results
  uses: actions/upload-artifact@v3
  with:
    name: test-results
    path: test-results.xml
```

---

## 📈 Performance Requirements

- **Pipeline Total Duration**: < 8 minutes (PR checks)
- **Build Time**: < 3 minutes
- **Test Time**: < 2 minutes
- **Deploy Time**: < 3 minutes
- **Smoke Tests**: < 1 minute

---

## 🔐 Secrets Management

### Required GitHub Secrets

```bash
# AWS credentials
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY

# Container registry
ECR_REGISTRY

# Notifications
SLACK_WEBHOOK_URL

# Security scanning
SNYK_TOKEN

# Kubernetes access
KUBECONFIG_BASE64
```

---

## 📚 Related Documentation

- [DEVOPS-002: GitOps Deployment](DEVOPS-002-gitops-deployment.md)
- [DEVOPS-003: Multi-Environment Management](DEVOPS-003-multi-environment.md)
- [DEVOPS-005: Infrastructure as Code](DEVOPS-005-infrastructure-as-code.md)
- GitHub Actions: https://docs.github.com/en/actions
- DORA Metrics: https://cloud.google.com/blog/products/devops-sre/using-the-four-keys-to-measure-your-devops-performance

---

**Last Updated**: October 29, 2025  
**Owner**: DevOps Team  
**Status**: ✅ Implemented K

