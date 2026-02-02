# 🔄 DEVOPS-019: Deploy and validate agent-sre in Notifi environment

**Status**: Backlog  | **Priority**: P2**Linear URL**: https://linear.app/bvlucena/issue/BVL-13/migration-deploy-and-validate-agent-sre-in-notifi-environment | **Status**: Backlog  | **Priority**: P2**Linear URL**: https://linear.app/bvlucena/issue/BVL-13/migration-deploy-and-validate-agent-sre-in-notifi-environment | **Story Points**: 13

**Created**: 2025-12-26T14:37:29.839Z  
**Updated**: 2025-12-26T14:37:29.839Z  
**Project**: knative-lambda-operator  

---

# 🎯 Objective

Deploy agent-sre as LambdaAgent resource in Notifi and validate all functionality.


## 📋 User Story

**As a** DevOps Engineer  
**I want to** deploy and validate agent-sre in notifi environment  
**So that** I can improve system reliability, security, and performance

---


## 📋 Prerequisites

* knative-lambda-operator deployed and functional
* Dependencies verified (Ollama, storage, observability)
* Container images migrated
* Configuration updated for Notifi

## 🔧 Tasks

### 1\. LambdaAgent Deployment

- [ ] Create LambdaAgent resource manifest for Notifi
- [ ] Apply LambdaAgent resource
- [ ] Verify operator creates Knative Service
- [ ] Verify operator creates CloudEvents triggers
- [ ] Verify agent pod starts successfully

### 2\. Model Loading Verification

- [ ] Verify Ollama model is accessible
- [ ] Test model loading in agent container
- [ ] Verify inference works correctly
- [ ] Test with sample prompts
- [ ] Verify model performance

### 3\. Observability Integration

- [ ] Test Prometheus metrics collection
- [ ] Test Loki log queries
- [ ] Verify metrics are exposed correctly
- [ ] Verify distributed tracing (if configured)
- [ ] Check Grafana dashboards (if exist)

### 4\. Health Report Generation

- [ ] Test SRE health report generation
- [ ] Verify Prometheus queries work
- [ ] Verify Loki queries work
- [ ] Verify report formatting
- [ ] Test report delivery/output

### 5\. LambdaFunction Integration (Supporting Functions)

- [ ] Deploy supporting LambdaFunctions (if any)
- [ ] Test pod-restart LambdaFunction
- [ ] Test flux-reconcile LambdaFunctions
- [ ] Test check-pvc-status LambdaFunction
- [ ] Test scale-deployment LambdaFunction

### 6\. Integration Tests

- [ ] Run unit tests
- [ ] Run integration tests
- [ ] Run E2E tests (if applicable)
- [ ] Test k6 load tests

### 7\. Security Validation

- [ ] Verify RBAC permissions
- [ ] Verify service account permissions
- [ ] Verify secret access
- [ ] Run security scans
- [ ] Verify network policies

## ✅ Acceptance Criteria

- [ ] LambdaAgent deployed successfully
- [ ] Model loading and inference working
- [ ] Prometheus/Loki integration verified
- [ ] Health report generation functional
- [ ] Supporting LambdaFunctions working (if any)
- [ ] All tests passing
- [ ] No critical security issues
- [ ] Documentation updated

## 🔗 Dependencies

INFRA-MIG-004: Deploy and validate knative-lambda-operator in dev environment
INFRA-MIG-008: Update agent-sre configuration for Notifi environment
INFRA-MIG-007: Migrate agent-sre container images to Notifi registry
