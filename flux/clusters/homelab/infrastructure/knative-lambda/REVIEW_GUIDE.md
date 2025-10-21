# 🔍 Knative Lambda Builder - File Review Guide

## 📋 Quick Reference

This guide provides a quick overview of the main files to review when improving the Knative Lambda Builder. Each section includes file location, complexity, and specific review points.

---

## 🏗️ Core Business Logic

### 1. Event Handler
**File**: `internal/handler/event_handler.go`  
**Lines**: ~500 | **Complexity**: 🔴 High | **Priority**: P0

**What it does**:
- Main CloudEvent processing orchestration
- Routes events to appropriate handlers
- Manages component composition via dependency injection

**Review Checklist**:
```go
// Key Functions to Review:
- ProcessCloudEvent()           // Line ~XX - Main entry point
- processBuildStartEvent()      // Line ~XX - Build start handling
- processBuildCompleteEvent()   // Line ~XX - Build completion handling
- isBuildStartEvent()          // Line ~XX - Event type checking
- isBuildCompleteEvent()       // Line ~XX - Event type checking
```

**Issues to Look For**:
- [ ] CloudEvent metrics NOT recorded (defined in observability.go but not used here)
- [ ] Functions > 50 lines (violates VALIDATION.md)
- [ ] Error context not fully propagated
- [ ] Missing correlation ID in some paths
- [ ] Magic strings for event types

**Immediate Actions**:
1. Add CloudEvent metrics: `obs.GetMetrics().CloudEventsTotal.WithLabelValues(...).Inc()`
2. Add CloudEvent duration: `obs.GetMetrics().CloudEventDuration.WithLabelValues(...).Observe(duration)`
3. Extract long functions into smaller helpers
4. Replace magic strings with constants from `internal/constants/`

---

### 2. Service Manager
**File**: `internal/handler/service_manager.go`  
**Lines**: ~600 | **Complexity**: 🔴 High | **Priority**: P1

**What it does**:
- Creates Knative services from successful builds
- Creates Knative triggers for event routing
- Manages service lifecycle

**Review Checklist**:
```go
// Key Functions to Review:
- CreateService()              // Line ~XX - Main service creation
- CreateTrigger()              // Line ~XX - Trigger creation
- generateServiceName()        // Line ~XX - Name generation
- serviceExists()              // Line ~XX - Existence check
- updateService()              // Line ~XX - Service updates
```

**Issues to Look For**:
- [ ] Complex service creation logic (extract to templates)
- [ ] Duplicated resource creation code
- [ ] Error handling could be more granular
- [ ] Resource cleanup not verified
- [ ] Missing tests for edge cases

**Immediate Actions**:
1. Extract Knative resource creation into `internal/templates/`
2. Add unit tests for all creation scenarios
3. Add resource cleanup verification
4. Improve error categorization (transient vs permanent)

---

### 3. Job Manager
**File**: `internal/handler/job_manager.go`  
**Lines**: ~400 | **Complexity**: 🟡 Medium-High | **Priority**: P1

**What it does**:
- Creates Kubernetes jobs for Kaniko builds
- Manages job lifecycle
- Handles job cleanup

**Review Checklist**:
```go
// Key Functions to Review:
- CreateBuildJob()             // Line ~XX - Job creation
- waitForJobCompletion()       // Line ~XX - Job monitoring
- cleanupJob()                 // Line ~XX - Job cleanup
- getJobStatus()               // Line ~XX - Status checking
```

**Issues to Look For**:
- [ ] Job cleanup incomplete (check all error paths)
- [ ] Timeout handling edge cases
- [ ] Resource limit validation missing
- [ ] Job failure reasons not categorized
- [ ] Missing tests for timeout scenarios

**Immediate Actions**:
1. Add comprehensive job lifecycle tests
2. Improve error categorization
3. Add job status monitoring metrics
4. Document job failure scenarios

---

### 4. Build Context Manager
**File**: `internal/handler/build_context_manager.go`  
**Lines**: ~350 | **Complexity**: 🟡 Medium | **Priority**: P1

**What it does**:
- Creates build contexts from source code
- Uploads contexts to S3
- Manages context cleanup

**Review Checklist**:
```go
// Key Functions to Review:
- CreateBuildContext()         // Line ~XX - Context creation
- uploadToS3()                 // Line ~XX - S3 upload
- cleanupContext()             // Line ~XX - Context cleanup
- validateContext()            // Line ~XX - Context validation
```

**Issues to Look For**:
- [ ] S3 upload optimization (multipart, compression)
- [ ] Build context size not validated
- [ ] Cleanup not handling all error cases
- [ ] Memory usage for large contexts
- [ ] Missing performance metrics

**Immediate Actions**:
1. Add multipart upload for large contexts
2. Add context size validation
3. Optimize memory usage
4. Add performance metrics (upload time, context size)

---

## 📊 Observability

### 5. Observability Core
**File**: `internal/observability/observability.go`  
**Lines**: ~1200 | **Complexity**: 🔴 Very High | **Priority**: P0

**What it does**:
- Defines ALL metrics (HTTP, Build, K8s, AWS, System, Error)
- Provides metric recording helpers
- Manages tracing and logging
- Collects system metrics

**Review Checklist**:
```go
// Key Sections to Review:
- initializeMetrics()          // Lines 242-540 - Metric definitions
- Metrics struct               // Lines 50-100 - Metric fields
- RecordBuildRequest()         // Lines 745-824 - Build metrics
- RecordK8sJobCreation()       // Lines 843-888 - K8s metrics
- RecordAWSOperation()         // Lines 906-940 - AWS metrics
- CollectSystemMetrics()       // Lines 1075-1148 - System metrics
```

**Issues to Look For**:
- [ ] File TOO LARGE (1200+ lines) - violates maintainability
- [ ] CloudEvent metrics defined but NOT integrated
- [ ] Potential high metric cardinality
- [ ] Duplicated metric recording patterns
- [ ] Missing tests for metric recording

**Immediate Actions** (CRITICAL):
1. **SPLIT THIS FILE** into:
   - `observability.go` - Core setup (200 lines)
   - `metrics.go` - Metric definitions (300 lines)
   - `metrics_helpers.go` - Recording helpers (400 lines)
   - `tracing.go` - Tracing setup (100 lines)
   - `logging.go` - Logging setup (100 lines)
   - `system_metrics.go` - System collection (100 lines)

2. Add CloudEvent metrics to `event_handler.go`
3. Review metric cardinality (especially labels)
4. Add metric collection tests

**Refactoring Plan**:
```go
// Before (1 file, 1200 lines):
internal/observability/
  └── observability.go (1200 lines) ❌

// After (6 files, same total):
internal/observability/
  ├── observability.go      (200 lines) ✅ Core + initialization
  ├── metrics.go            (300 lines) ✅ Metric definitions
  ├── metrics_helpers.go    (400 lines) ✅ Recording helpers
  ├── tracing.go            (100 lines) ✅ Tracing setup
  ├── logging.go            (100 lines) ✅ Logging setup
  └── system_metrics.go     (100 lines) ✅ System collection
```

---

### 6. Middleware
**File**: `internal/handler/middleware.go`  
**Lines**: ~200 | **Complexity**: 🟡 Medium | **Priority**: P2

**What it does**:
- Records HTTP metrics
- Handles request/response
- Adds security headers
- Manages CORS

**Review Checklist**:
```go
// Key Functions to Review:
- ObservabilityMiddleware()    // Line ~XX - Main middleware
- recordMetrics()              // Line ~XX - Metric recording
- addSecurityHeaders()         // Line ~XX - Security headers
```

**Issues to Look For**:
- [ ] Security header tests missing
- [ ] Request validation could be stronger
- [ ] Error handling could be more consistent
- [ ] Missing rate limiting integration

**Immediate Actions**:
1. Add comprehensive security header tests
2. Verify metrics accuracy
3. Add request validation
4. Document middleware chain

---

## ⚙️ Configuration

### 7. Configuration Core
**File**: `internal/config/config.go`  
**Lines**: ~300 | **Complexity**: 🟡 Medium | **Priority**: P2

**What it does**:
- Loads configuration from environment variables
- Validates configuration
- Provides defaults
- Manages config struct

**Review Checklist**:
```go
// Key Functions to Review:
- Load()                       // Line ~XX - Config loading
- Validate()                   // Line ~XX - Validation
- setDefaults()               // Line ~XX - Default values
```

**Issues to Look For**:
- [ ] Validation may be incomplete
- [ ] Default values not optimal
- [ ] Sensitive data may be logged
- [ ] Required fields not enforced
- [ ] Missing configuration examples

**Immediate Actions**:
1. Add comprehensive validation tests
2. Document all configuration options
3. Create configuration examples
4. Ensure no secrets in logs

---

## 🔐 Security

### 8. Security
**File**: `internal/security/security.go`  
**Lines**: ~200 | **Complexity**: 🟡 Medium | **Priority**: P0

**What it does**:
- Input validation
- Sanitization
- Rate limiting integration
- Security checks

**Review Checklist**:
```go
// Key Functions to Review:
- ValidateInput()              // Line ~XX - Input validation
- SanitizeString()            // Line ~XX - Sanitization
- ValidateID()                // Line ~XX - ID validation
```

**Issues to Look For**:
- [ ] Validation may not cover all inputs
- [ ] Error messages may leak information
- [ ] Regex patterns not documented
- [ ] Missing injection vulnerability tests
- [ ] Rate limiting not comprehensive

**Immediate Actions**:
1. Add comprehensive input validation tests
2. Review error messages for leakage
3. Document regex patterns
4. Add security scanning to CI/CD

---

## 📝 Templates

### 9. Templates
**File**: `internal/templates/templates.go`  
**Lines**: ~400 | **Complexity**: 🟡 Medium | **Priority**: P2

**What it does**:
- Defines Dockerfile templates
- Defines function templates
- Manages template rendering

**Review Checklist**:
```go
// Key Functions to Review:
- RenderDockerfile()          // Line ~XX - Dockerfile rendering
- RenderPackageJSON()         // Line ~XX - Package.json rendering
- RenderFunctionCode()        // Line ~XX - Function code rendering
```

**Issues to Look For**:
- [ ] Template injection vulnerabilities
- [ ] Template variables not validated
- [ ] Rendering errors not handled
- [ ] Missing tests for all templates
- [ ] No template security scanning

**Immediate Actions**:
1. Add template validation
2. Add comprehensive rendering tests
3. Document template variables
4. Add security scanning for templates

---

## 🧪 Testing Files

### Current Test Files
```
internal/
├── aws/client_test.go           ✅ Exists
├── config/
│   ├── aws_test.go              ✅ Exists
│   ├── config_test.go           ✅ Exists
│   └── observability_test.go    ✅ Exists
├── errors/errors_test.go        ✅ Exists
├── handler/
│   ├── async_job_creator_test.go       ✅ Exists
│   ├── build_context_manager_test.go   ✅ Exists
│   ├── event_handler_test.go           ✅ Exists (expand)
│   ├── job_manager_test.go             ✅ Exists (expand)
│   └── service_manager_test.go         ✅ Exists (expand)
├── resilience/resilience_test.go       ✅ Exists
├── security/security_test.go           ✅ Exists
└── templates/templates_test.go         ✅ Exists
```

### Missing Test Coverage
```
❌ observability/observability_test.go  - CRITICAL (0% coverage)
❌ handler/cloud_event_handler_test.go  - HIGH (0% coverage)
❌ handler/middleware_test.go           - MEDIUM (0% coverage)
❌ handler/http_handler_test.go         - MEDIUM (0% coverage)
```

---

## 📦 Deployment Files

### Helm Templates to Review

**Priority Order**:

1. **`deploy/templates/builder.yaml`** - Main service deployment
   - Review resource limits
   - Review autoscaling configuration
   - Review environment variables

2. **`deploy/templates/alerts-*.yaml`** - Alert definitions (13 files)
   - Review alert thresholds
   - Add runbook annotations
   - Test alert firing

3. **`deploy/templates/triggers.yaml`** - Event triggers
   - Review filter configuration
   - Review subscriber references

4. **`deploy/values.yaml`** - Configuration values
   - Document all values
   - Add validation
   - Create examples

---

## 🎯 Priority Review Order

### Week 1: Critical Observability
1. ✅ Read `internal/observability/observability.go`
2. 🔧 Split observability.go into multiple files
3. 🔧 Implement CloudEvent metrics in event_handler.go
4. ✅ Add observability tests

### Week 2: Core Logic
1. ✅ Review `internal/handler/event_handler.go`
2. ✅ Review `internal/handler/service_manager.go`
3. ✅ Review `internal/handler/job_manager.go`
4. 🔧 Add missing tests

### Week 3: Security & Config
1. ✅ Review `internal/security/security.go`
2. ✅ Review `internal/config/config.go`
3. 🔧 Add security tests
4. 🔧 Add configuration validation

### Week 4: Templates & Build
1. ✅ Review `internal/templates/templates.go`
2. ✅ Review `internal/handler/build_context_manager.go`
3. 🔧 Optimize build context
4. 🔧 Add template tests

---

## 🔍 Code Quality Tools

### Run Before Review
```bash
# Linting
make lint

# Tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Complexity analysis
gocyclo -over 15 .
gocognit -over 15 .

# Code duplication
dupl -threshold 50 .

# Security scanning
gosec ./...
```

### Metrics to Track
- **Test Coverage**: `go tool cover -func=coverage.out | grep total`
- **Code Complexity**: `gocyclo -avg .`
- **Function Length**: `wc -l` per function
- **Duplicate Code**: `dupl -t 50 .`

---

## 📋 Review Checklist Template

Use this for each file review:

```markdown
## File: internal/handler/example.go

### Basic Info
- Lines: XXX
- Complexity: Low/Medium/High
- Last Modified: YYYY-MM-DD
- Test Coverage: XX%

### Review Findings
- [ ] Functions < 50 lines
- [ ] Single responsibility per function
- [ ] Proper error handling
- [ ] No magic numbers/strings
- [ ] Comprehensive tests
- [ ] Documentation complete
- [ ] No security issues
- [ ] Performance acceptable

### Issues Found
1. Issue description
2. Issue description

### Action Items
- [ ] Action item 1
- [ ] Action item 2

### Priority: P0/P1/P2/P3
```

---

## 🎯 Success Criteria

### Code Quality
- [ ] All functions < 50 lines
- [ ] Test coverage > 80%
- [ ] No critical security issues
- [ ] No high complexity functions
- [ ] No duplicate code > 50 lines

### Documentation
- [ ] All public functions documented
- [ ] All complex logic explained
- [ ] All TODOs moved to TODO.md
- [ ] All edge cases documented

### Testing
- [ ] All paths tested
- [ ] All error cases tested
- [ ] All edge cases tested
- [ ] Integration tests exist

---

**Last Updated**: 2025-01-21  
**Next Review**: Weekly  
**Maintainer**: @brunolucena

