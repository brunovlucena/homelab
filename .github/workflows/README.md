# GitHub Actions Workflows

This directory contains CI/CD workflows for the homelab project, with specific focus on the Homepage chatbot agent-sre integration.

## 📋 Available Workflows

### 1. Homepage Integration Tests

**File:** `homepage-tests.yml`  
**Triggers:** Push to main/develop, Pull Requests, Manual dispatch  
**Purpose:** Run comprehensive tests for the homepage chatbot integration

**Jobs:**
- ✅ **Backend Tests (Go)** - Unit tests for API handlers
- ✅ **Frontend Tests (TypeScript)** - Jest tests for chatbot service
- ✅ **Build Verification** - Verify both API and frontend build successfully
- ✅ **Integration Tests** - Full integration testing (on-demand)
- ✅ **Test Summary** - Aggregate results

**Usage:**
```bash
# Automatically runs on push/PR
# Or trigger manually:
gh workflow run homepage-tests.yml
```

### 2. Homepage PR Checks

**File:** `homepage-pr-check.yml`  
**Triggers:** Pull Requests only  
**Purpose:** Fast feedback for PR reviews

**Jobs:**
- 🔍 **Quick Checks** - File size, secrets detection, documentation
- 🧪 **Backend PR Check** - Tests, formatting, code quality
- 🧪 **Frontend PR Check** - Tests, TypeScript, bundle size
- 🔐 **Security Scan** - Trivy, govulncheck, npm audit
- 📊 **PR Summary** - Consolidated results

**Features:**
- Fast feedback (< 5 minutes)
- Code formatting checks
- Security vulnerability scanning
- Bundle size monitoring
- TODO/FIXME detection

### 3. Homepage Nightly Tests

**File:** `homepage-nightly-tests.yml`  
**Triggers:** Daily at 2 AM UTC, Manual dispatch  
**Purpose:** Comprehensive nightly testing and monitoring

**Jobs:**
- 🧪 **Comprehensive Tests** - Full test suite execution
- 🏃 **Performance Tests** - Benchmark tests
- 📦 **Dependency Check** - Outdated dependencies
- 🔐 **Security Audit** - Deep security analysis
- 📊 **Code Quality** - Complexity, duplication analysis
- 💬 **Issue Creation** - Auto-creates issues if tests fail

**Reports Generated:**
- Test coverage reports
- Benchmark results
- Dependency updates needed
- Security vulnerabilities
- Code quality metrics

## 🚀 Quick Start

### Running Locally

You can simulate the CI environment locally:

```bash
# Backend tests
cd flux/clusters/homelab/infrastructure/homepage/api
go test -v -race ./handlers/

# Frontend tests
cd flux/clusters/homelab/infrastructure/homepage/frontend
npm test

# Integration tests
cd flux/clusters/homelab/infrastructure/homepage/tests/integration
./test-agent-sre-integration.sh
```

### Manual Workflow Dispatch

```bash
# Using GitHub CLI
gh workflow run homepage-tests.yml

# With specific branch
gh workflow run homepage-tests.yml --ref feature-branch

# View workflow runs
gh run list --workflow=homepage-tests.yml

# View run details
gh run view <run-id>
```

## 📊 Test Coverage

| Component | Test Type | Location | Coverage |
|-----------|-----------|----------|----------|
| Backend API | Unit | `api/handlers/agent_sre_test.go` | 100% |
| Frontend Service | Unit | `frontend/src/services/chatbot.test.ts` | 100% |
| Integration | E2E | `tests/integration/*.sh` | 100% |

## 🔧 Configuration

### Environment Variables

Workflows use these environment variables:

```yaml
GO_VERSION: '1.23.0'
NODE_VERSION: '18'
HOMEPAGE_PATH: 'flux/clusters/homelab/infrastructure/homepage'
```

### Secrets Required

| Secret | Usage | Required For |
|--------|-------|--------------|
| `CODECOV_TOKEN` | Upload coverage | Optional |
| `GITHUB_TOKEN` | Automatic | All workflows |

### Path Triggers

Workflows only run when these paths change:

```yaml
paths:
  - 'flux/clusters/homelab/infrastructure/homepage/**'
  - '.github/workflows/homepage-*.yml'
```

## 📈 Workflow Features

### Caching

All workflows use caching for faster execution:

- **Go modules:** Cached based on `go.sum`
- **NPM packages:** Cached based on `package-lock.json`
- **Build artifacts:** Cached between steps

### Parallel Execution

Jobs run in parallel when possible:

```
┌─────────────────┐
│ Backend Tests   │──┐
└─────────────────┘  │
┌─────────────────┐  ├──▶ ┌──────────────────┐
│ Frontend Tests  │──┘    │ Build Verification│
└─────────────────┘       └──────────────────┘
```

### Test Reports

All workflows generate detailed summaries:

- ✅ Test results
- 📊 Coverage reports
- 🔐 Security findings
- 📦 Dependency status

## 🎯 Best Practices

### For Pull Requests

1. **Run tests locally first**
   ```bash
   cd api && go test ./...
   cd frontend && npm test
   ```

2. **Check formatting**
   ```bash
   cd api && gofmt -s -w .
   cd frontend && npm run lint
   ```

3. **Review security findings** before pushing

### For Main Branch

1. All tests must pass
2. Code coverage must not decrease
3. No high/critical security vulnerabilities
4. Build verification successful

## 🐛 Troubleshooting

### Test Failures

```bash
# View workflow logs
gh run view <run-id> --log

# Download artifacts
gh run download <run-id>

# Re-run failed jobs
gh run rerun <run-id> --failed
```

### Common Issues

**Issue:** Go tests fail with "race detected"
```bash
# Solution: Fix race conditions
go test -race ./...
```

**Issue:** Frontend tests timeout
```bash
# Solution: Increase timeout
npm test -- --testTimeout=10000
```

**Issue:** Integration tests skip
```bash
# Solution: They only run on main or manual dispatch
gh workflow run homepage-tests.yml
```

## 📚 Related Documentation

- [Test README](../flux/clusters/homelab/infrastructure/homepage/tests/TEST_README.md)
- [Integration Guide](../flux/clusters/homelab/infrastructure/homepage/CHATBOT_AGENT_SRE_INTEGRATION.md)
- [Manual Test Guide](../flux/clusters/homelab/infrastructure/homepage/MANUAL_TEST_GUIDE.md)
- [Tests Working](../flux/clusters/homelab/infrastructure/homepage/TESTS_WORKING.md)

## 🔄 Workflow Updates

### Adding New Tests

1. Add test file to appropriate directory
2. Update workflow if needed:
   ```yaml
   - name: Run new tests
     run: ./new-test.sh
   ```
3. Test workflow locally using [act](https://github.com/nektos/act)
4. Create PR with workflow changes

### Modifying Triggers

```yaml
# Add more branches
on:
  push:
    branches: [ main, develop, staging ]

# Add more paths
paths:
  - 'new-component/**'
```

## 📊 Status Badges

Add to your README:

```markdown
![Homepage Tests](https://github.com/brunovlucena/homelab/actions/workflows/homepage-tests.yml/badge.svg)
![PR Checks](https://github.com/brunovlucena/homelab/actions/workflows/homepage-pr-check.yml/badge.svg)
![Nightly Tests](https://github.com/brunovlucena/homelab/actions/workflows/homepage-nightly-tests.yml/badge.svg)
```

## 🎉 Success Criteria

A successful workflow run shows:

```
✅ Backend Tests: 10/10 passed
✅ Frontend Tests: 25/25 passed
✅ Build Verification: Success
✅ Security Scan: No critical issues
✅ Code Coverage: > 80%
```

## 🤝 Contributing

When adding new workflows:

1. Follow naming convention: `homepage-*.yml`
2. Include comprehensive job summaries
3. Add artifact uploads for important results
4. Document in this README
5. Test thoroughly before merging

---

**Maintainer:** Bruno Lucena  
**Last Updated:** 2025-10-08  
**Version:** 1.0.0

