# 🚀 Knative Lambda Builder - TODO & Improvement Plan

## 📋 Project Overview

**Knative Lambda Builder** is a serverless function builder that mimics AWS Lambda using Knative. It processes CloudEvents from RabbitMQ to orchestrate the complete build-to-deploy lifecycle with comprehensive observability (Prometheus, Tempo, Loki).

### Current State
- ✅ Core build pipeline (Kaniko + Knative)
- ✅ RabbitMQ CloudEvents integration
- ✅ Comprehensive metrics collection
- ✅ Distributed tracing (OpenTelemetry + Tempo)
- ✅ Alerting with PrometheusRules
- ✅ JavaScript/Node.js support
- ⚠️ CloudEvent metrics defined but not implemented
- ⚠️ Testing coverage incomplete
- ⚠️ Documentation needs expansion

---

## 🎯 Priority System

- **P0** (Critical) - Blocking issues, security, production bugs
- **P1** (High) - Important improvements, missing functionality
- **P2** (Medium) - Nice-to-have improvements, optimization
- **P3** (Low) - Future enhancements, technical debt

---

## 📊 Observability Improvements

### P0: Implement CloudEvent Metrics in Handler
**Status**: 🔴 Not Started  
**File**: `internal/handler/cloud_event_handler.go`

CloudEvent metrics are defined in `internal/observability/observability.go` (lines 256-290) but not yet integrated into the event processing pipeline.

**Action Items**:
- [ ] Add CloudEvent metrics recording in `cloud_event_handler.go`
  - [ ] Record `CloudEventsTotal` counter on each event
  - [ ] Record `CloudEventDuration` histogram for processing time
  - [ ] Record `CloudEventSize` histogram for incoming event size
  - [ ] Record `CloudEventResponseSize` histogram for response size
- [ ] Add metric labels: `method`, `endpoint`, `status_code`, `handler`
- [ ] Test metrics are exposed at `/metrics` endpoint
- [ ] Update Grafana dashboards to include CloudEvent metrics
- [ ] Create alerts for CloudEvent error rates

**Files to Modify**:
- `internal/handler/cloud_event_handler.go` - Add metrics recording
- `deploy/templates/alerts.yaml` - Add CloudEvent alerts
- `dashboards/` - Update Grafana dashboards

---

### P1: Enhanced Error Tracking and Correlation
**Status**: 🟡 In Progress

**Action Items**:
- [ ] Implement structured error logging with correlation IDs
- [ ] Add error categorization (transient, permanent, unknown)
- [ ] Link errors to traces using exemplars
- [ ] Create error rate dashboards by component
- [ ] Add error budget tracking for SLOs

**Files to Create/Modify**:
- `internal/errors/categories.go` - Error categorization
- `internal/observability/error_tracking.go` - Enhanced error tracking
- `dashboards/error-analysis.json` - Error dashboard

---

### P2: Add Exemplars to All Metrics
**Status**: 🟡 In Progress

**Action Items**:
- [ ] Review current exemplar implementation in `internal/observability/exemplars.go`
- [ ] Add trace ID exemplars to all histograms
- [ ] Test exemplar display in Grafana
- [ ] Document exemplar usage in METRICS.md

---

## 🧪 Testing Improvements

### P0: Increase Unit Test Coverage
**Status**: 🔴 Not Started  
**Current Coverage**: ~40% (estimate)  
**Target**: 80%

**Action Items**:
- [ ] Add tests for `event_handler.go` (priority handlers)
- [ ] Add tests for `service_manager.go` (Knative service creation)
- [ ] Add tests for `job_manager.go` (Kaniko job management)
- [ ] Add tests for `build_context_manager.go` (S3 context creation)
- [ ] Add mock implementations for external dependencies
- [ ] Set up test coverage reporting in CI/CD

**Files to Create**:
- `internal/handler/event_handler_test.go` (expand)
- `internal/handler/service_manager_test.go` (expand)
- `internal/handler/job_manager_test.go` (expand)
- `internal/handler/build_context_manager_test.go` (expand)
- `internal/handler/mocks/` - Mock implementations

---

### P1: Integration Tests
**Status**: 🔴 Not Started

**Action Items**:
- [ ] Create integration test suite
  - [ ] Test complete build-to-deploy flow
  - [ ] Test RabbitMQ event processing
  - [ ] Test Kaniko job creation and monitoring
  - [ ] Test Knative service creation
- [ ] Set up test environment (kind cluster + RabbitMQ)
- [ ] Add integration tests to CI/CD pipeline
- [ ] Document test setup in `tests/README.md`

**Files to Create**:
- `tests/integration/` - Integration test suite
- `tests/integration/README.md` - Setup instructions
- `tests/integration/setup.sh` - Test environment setup

---

### P2: Load Testing
**Status**: 🟡 In Progress  
**Files**: `tests/k6/`

**Action Items**:
- [ ] Expand k6 load tests for build events
- [ ] Add stress tests for queue depth
- [ ] Test autoscaling behavior under load
- [ ] Document load test results
- [ ] Create performance benchmarks

---

## 🏗️ Code Quality & Architecture

### P1: Refactor God Objects
**Status**: 🟡 In Progress

Based on VALIDATION.md checklist, identify and refactor any "God Objects" or "God Configs".

**Action Items**:
- [ ] Audit `internal/config/config.go` - Split into smaller configs
- [ ] Review `internal/handler/container.go` - Ensure proper DI
- [ ] Ensure all functions are < 50 lines
- [ ] Ensure single responsibility per function/struct
- [ ] Add complexity analysis to CI/CD

**Files to Review**:
- `internal/config/config.go` - Main config (currently fragmented, good)
- `internal/handler/event_handler.go` - Main handler
- `internal/observability/observability.go` - Large file, review complexity

---

### P2: Eliminate Magic Numbers and Strings
**Status**: 🔴 Not Started

**Action Items**:
- [ ] Scan codebase for magic numbers
- [ ] Move all constants to `internal/constants/constants.go`
- [ ] Create typed constants for event types
- [ ] Create typed constants for metric names
- [ ] Document all constants

---

### P2: Improve Error Handling
**Status**: 🟡 In Progress

**Action Items**:
- [ ] Ensure all errors use custom error types from `internal/errors/`
- [ ] Add error wrapping context everywhere
- [ ] Remove any panic() calls, use proper error returns
- [ ] Add error categorization (transient/permanent)
- [ ] Test error propagation paths

---

## 📚 Documentation Improvements

### P1: Expand Technical Documentation
**Status**: 🟡 In Progress

**Action Items**:
- [ ] Create `docs/ARCHITECTURE_DEEP_DIVE.md`
  - [ ] Component interaction diagrams
  - [ ] Sequence diagrams for build flow
  - [ ] Data flow diagrams
- [ ] Create `docs/RUNBOOK.md` (referenced but missing)
  - [ ] Common troubleshooting scenarios
  - [ ] Step-by-step recovery procedures
  - [ ] Links to dashboards and queries
- [ ] Create `docs/DEPLOYMENT.md` (referenced but missing)
  - [ ] Step-by-step deployment guide
  - [ ] Environment setup
  - [ ] Configuration reference
- [ ] Create `docs/ALERTING.md` (referenced but missing)
  - [ ] Alert descriptions and severity
  - [ ] Escalation procedures
  - [ ] Runbook links
- [ ] Expand `docs/JOB_START_EVENTS.md`
  - [ ] Add more examples
  - [ ] Add error scenarios
  - [ ] Add troubleshooting section

**Files to Create**:
- `docs/ARCHITECTURE_DEEP_DIVE.md`
- `docs/RUNBOOK.md`
- `docs/DEPLOYMENT.md`
- `docs/ALERTING.md`
- `docs/DEVELOPMENT_GUIDE.md`
- `docs/CONTRIBUTING.md`

---

### P2: API Documentation
**Status**: 🔴 Not Started

**Action Items**:
- [ ] Document all CloudEvent types
- [ ] Document all HTTP endpoints
- [ ] Create OpenAPI/Swagger spec
- [ ] Add example requests/responses
- [ ] Document error codes

**Files to Create**:
- `docs/API.md` - API reference
- `docs/openapi.yaml` - OpenAPI specification
- `docs/EVENTS.md` - CloudEvent reference

---

### P2: Code Documentation
**Status**: 🟡 In Progress

**Action Items**:
- [ ] Add godoc comments to all exported functions
- [ ] Add package-level documentation
- [ ] Add examples in godoc
- [ ] Generate and publish godoc
- [ ] Add inline comments for complex logic

---

## 🔐 Security Improvements

### P0: Input Validation Enhancement
**Status**: 🟡 In Progress  
**File**: `internal/security/security.go`

**Action Items**:
- [ ] Audit all input validation points
- [ ] Add validation for all CloudEvent fields
- [ ] Add validation for all environment variables
- [ ] Add validation for all Kubernetes resources
- [ ] Test with malicious inputs
- [ ] Document security assumptions

---

### P1: Secrets Management
**Status**: 🟡 In Progress

**Action Items**:
- [ ] Review secret handling in `deploy/templates/secrets.yaml`
- [ ] Ensure no secrets in logs or errors
- [ ] Add secret rotation documentation
- [ ] Test secret injection in all environments
- [ ] Document secret management in README

---

### P2: Security Scanning
**Status**: 🔴 Not Started

**Action Items**:
- [ ] Add dependency scanning to CI/CD (Dependabot/Renovate)
- [ ] Add container image scanning (Trivy)
- [ ] Add SAST scanning (gosec, golangci-lint security rules)
- [ ] Document security scanning process
- [ ] Create security policy document

**Files to Create**:
- `.github/workflows/security-scan.yml` - Security scanning workflow
- `SECURITY.md` - Security policy

---

## ⚡ Performance Optimization

### P1: Build Context Optimization
**Status**: 🔴 Not Started  
**File**: `internal/handler/build_context_manager.go`

**Action Items**:
- [ ] Profile build context creation time
- [ ] Optimize S3 uploads (multipart, compression)
- [ ] Add build context caching
- [ ] Reduce build context size
- [ ] Add metrics for context size and duration

---

### P2: Kaniko Build Optimization
**Status**: 🟡 In Progress

**Action Items**:
- [ ] Review Kaniko cache configuration
- [ ] Optimize NPM install (cache, --prefer-offline)
- [ ] Reduce Docker image layers
- [ ] Add build time metrics
- [ ] Document build optimization techniques

---

### P2: Resource Optimization
**Status**: 🟡 In Progress

**Action Items**:
- [ ] Review CPU/memory allocations
- [ ] Optimize goroutine usage
- [ ] Profile memory allocations
- [ ] Reduce container image size
- [ ] Add resource usage dashboards

---

## 🚀 Feature Enhancements

### P1: Go Language Support
**Status**: 🔴 Not Started

As mentioned in README.md, Go support is planned but not implemented.

**Action Items**:
- [ ] Design Go function build flow
- [ ] Create Go Dockerfile template
- [ ] Add Go module dependency resolution
- [ ] Create Go runtime configuration
- [ ] Add Go examples
- [ ] Test Go function builds
- [ ] Document Go support in README

**Files to Create**:
- `internal/templates/go_dockerfile.tpl` - Go Dockerfile template
- `internal/templates/go_main.tpl` - Go main function template
- `tests/fixtures/go/` - Go test fixtures

---

### P1: Python Language Support
**Status**: 🔴 Not Started

As mentioned in README.md, Python support is planned but not implemented.

**Action Items**:
- [ ] Design Python function build flow
- [ ] Create Python Dockerfile template
- [ ] Add pip dependency resolution
- [ ] Create Python runtime configuration
- [ ] Add Python examples
- [ ] Test Python function builds
- [ ] Document Python support in README

**Files to Create**:
- `internal/templates/python_dockerfile.tpl` - Python Dockerfile template
- `internal/templates/python_handler.tpl` - Python handler template
- `tests/fixtures/python/` - Python test fixtures

---

### P2: MinIO Storage Support
**Status**: 🔴 Not Started

As mentioned in README.md, MinIO support is planned for on-premises S3-compatible storage.

**Action Items**:
- [ ] Design storage abstraction layer
- [ ] Implement MinIO client
- [ ] Add MinIO configuration
- [ ] Test with MinIO
- [ ] Add storage provider selection logic
- [ ] Document MinIO setup in README

**Files to Create/Modify**:
- `internal/storage/interface.go` - Storage abstraction
- `internal/storage/s3.go` - AWS S3 implementation
- `internal/storage/minio.go` - MinIO implementation
- `deploy/templates/minio-config.yaml` - MinIO configuration

---

### P3: Multi-Tenant Support Enhancement
**Status**: 🔴 Not Started

**Action Items**:
- [ ] Design tenant isolation model
- [ ] Add tenant-specific resource quotas
- [ ] Add tenant-specific metrics
- [ ] Test tenant isolation
- [ ] Document multi-tenancy in README

---

## 🔧 DevOps & CI/CD

### P1: GitHub Actions Workflows
**Status**: 🔴 Not Started

**Action Items**:
- [ ] Create CI workflow for tests and linting
- [ ] Create CD workflow for image building and pushing
- [ ] Add workflow for security scanning
- [ ] Add workflow for documentation deployment
- [ ] Add workflow for release automation

**Files to Create**:
- `.github/workflows/ci.yml` - CI pipeline
- `.github/workflows/cd.yml` - CD pipeline
- `.github/workflows/security.yml` - Security scanning
- `.github/workflows/docs.yml` - Documentation deployment

---

### P2: Improve Makefile
**Status**: 🟡 In Progress  
**File**: `Makefile`

**Action Items**:
- [ ] Add `make test-coverage` target with HTML report
- [ ] Add `make benchmark` target
- [ ] Add `make docs` target for godoc generation
- [ ] Add `make security-scan` target
- [ ] Improve help documentation
- [ ] Add validation targets (lint, fmt, vet in one command)

---

### P2: Local Development Experience
**Status**: 🟡 In Progress

**Action Items**:
- [ ] Create Docker Compose for local development
- [ ] Add local RabbitMQ setup
- [ ] Add local Prometheus/Grafana setup
- [ ] Create development setup script
- [ ] Document local development in DEVELOPMENT_GUIDE.md

**Files to Create**:
- `docker-compose.yml` - Local development stack
- `scripts/setup-local-dev.sh` - Development setup
- `docs/DEVELOPMENT_GUIDE.md` - Development guide

---

## 📊 Monitoring & Alerting

### P1: Enhance Grafana Dashboards
**Status**: 🟡 In Progress  
**Directory**: `dashboards/`

**Action Items**:
- [ ] Create comprehensive overview dashboard
- [ ] Create build pipeline dashboard
- [ ] Create error analysis dashboard
- [ ] Create resource usage dashboard
- [ ] Add CloudEvent metrics to dashboards
- [ ] Export dashboards as JSON
- [ ] Document dashboard usage

**Files to Create**:
- `dashboards/overview.json` - Main overview
- `dashboards/build-pipeline.json` - Build metrics
- `dashboards/errors.json` - Error analysis
- `dashboards/resources.json` - Resource usage

---

### P1: Alert Runbook Links
**Status**: 🔴 Not Started

**Action Items**:
- [ ] Add runbook links to all alerts
- [ ] Create runbook pages for each alert
- [ ] Add dashboard links to alerts
- [ ] Test alert routing to Slack
- [ ] Document alert escalation process

**Files to Modify**:
- `deploy/templates/alerts-*.yaml` - Add runbook annotations
- `docs/runbooks/` - Create runbook directory

---

### P2: SLO/SLI Tracking
**Status**: 🔴 Not Started

**Action Items**:
- [ ] Define SLOs for availability, latency, error rate
- [ ] Create SLI recording rules
- [ ] Create error budget dashboards
- [ ] Add error budget alerting
- [ ] Document SLOs in README

**Files to Create**:
- `deploy/templates/slo-rules.yaml` - SLO recording rules
- `docs/SLO.md` - SLO documentation

---

## 🧹 Technical Debt

### P2: Reduce Code Duplication
**Status**: 🔴 Not Started

**Action Items**:
- [ ] Identify duplicated code patterns
- [ ] Extract common functions to shared packages
- [ ] Review DRY principle violations
- [ ] Refactor duplicate Kubernetes resource creation
- [ ] Add linting rules to detect duplication

---

### P2: Dependency Updates
**Status**: 🟡 In Progress  
**File**: `go.mod`

**Action Items**:
- [ ] Update all dependencies to latest versions
- [ ] Add Dependabot/Renovate configuration
- [ ] Test with updated dependencies
- [ ] Document dependency update process
- [ ] Create dependency update schedule

**Files to Create**:
- `.github/dependabot.yml` - Dependabot configuration

---

### P3: Legacy Code Cleanup
**Status**: 🔴 Not Started

**Action Items**:
- [ ] Remove commented-out code
- [ ] Remove unused functions and variables
- [ ] Remove TODO comments (move to this file)
- [ ] Remove debug logging in production code
- [ ] Add linting rules to prevent legacy code

---

## 🎨 Code Formatting & Style

### P2: Enforce Code Style
**Status**: 🟡 In Progress  
**File**: `.golangci.yml`

**Action Items**:
- [ ] Review and enhance `.golangci.yml` configuration
- [ ] Add more linting rules (gocognit, gocyclo, etc.)
- [ ] Enforce consistent error handling
- [ ] Enforce consistent naming conventions
- [ ] Add pre-commit hooks for formatting

**Files to Create/Modify**:
- `.golangci.yml` - Enhance linter configuration
- `.pre-commit-config.yaml` - Pre-commit hooks

---

## 📦 Build & Deployment

### P1: Multi-Architecture Support
**Status**: 🔴 Not Started

**Action Items**:
- [ ] Add ARM64 build support
- [ ] Update Dockerfile for multi-arch builds
- [ ] Test on ARM64 platforms
- [ ] Update CI/CD for multi-arch builds
- [ ] Document architecture support

---

### P2: Helm Chart Improvements
**Status**: 🟡 In Progress  
**Directory**: `deploy/`

**Action Items**:
- [ ] Add Helm chart linting to CI/CD
- [ ] Add Helm chart testing
- [ ] Improve values.yaml documentation
- [ ] Add Helm chart versioning
- [ ] Publish Helm chart to registry

---

## 🔄 Maintenance Tasks

### P1: Regular Metrics Review
**Status**: 🔴 Not Started

**Action Items**:
- [ ] Review metrics collection every quarter
- [ ] Remove unused metrics
- [ ] Add missing business metrics
- [ ] Optimize metric cardinality
- [ ] Update METRICS.md documentation

---

### P2: Regular Alert Review
**Status**: 🔴 Not Started

**Action Items**:
- [ ] Review alerts every quarter
- [ ] Tune alert thresholds based on data
- [ ] Remove noisy alerts
- [ ] Add missing critical alerts
- [ ] Document alert tuning decisions

---

## 📝 Notes

### Recently Completed
- ✅ Initial project structure
- ✅ RabbitMQ CloudEvents integration
- ✅ Kaniko build pipeline
- ✅ Knative service creation
- ✅ Comprehensive metrics definitions
- ✅ Alert definitions
- ✅ Basic documentation

### Known Issues
- ⚠️ CloudEvent metrics defined but not implemented in handler
- ⚠️ Test coverage < 50%
- ⚠️ Some documentation referenced but not created (RUNBOOK.md, DEPLOYMENT.md, ALERTING.md)
- ⚠️ Go and Python language support planned but not implemented
- ⚠️ MinIO storage support planned but not implemented

### Future Considerations
- 🔮 gRPC support for events
- 🔮 WebAssembly function support
- 🔮 Function marketplace
- 🔮 Function versioning and blue/green deployments
- 🔮 Function cold start optimization
- 🔮 Function cost optimization

---

## 📅 Quarterly Goals

### Q1 2025
- [ ] P0: Implement CloudEvent metrics
- [ ] P0: Increase test coverage to 80%
- [ ] P1: Create missing documentation (RUNBOOK, DEPLOYMENT, ALERTING)
- [ ] P1: Enhance Grafana dashboards

### Q2 2025
- [ ] P1: Add Go language support
- [ ] P1: Add Python language support
- [ ] P1: Integration test suite
- [ ] P1: GitHub Actions workflows

### Q3 2025
- [ ] P2: MinIO storage support
- [ ] P2: Performance optimization
- [ ] P2: Multi-architecture support
- [ ] P2: Security scanning automation

### Q4 2025
- [ ] P3: Multi-tenant enhancements
- [ ] P3: Advanced features (versioning, blue/green)
- [ ] P3: Technical debt cleanup

---

## 🤝 Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) (to be created) for guidelines on contributing to this project.

---

**Last Updated**: 2025-01-21  
**Maintainer**: @brunolucena
