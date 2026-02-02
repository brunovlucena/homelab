# 🔄 DEVOPS-016: Verify dependencies and prerequisites for agent-sre migration

**Priority**: P1 | **Status**: Backlog  | **Story Points**: 13  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-10/migration-verify-dependencies-and-prerequisites-for-agent-sre-migration

**Created**: 2025-12-26T14:37:02.485Z  
**Updated**: 2025-12-26T14:37:02.485Z  
**Project**: knative-lambda-operator  

---

# 🎯 Objective

Verify all dependencies and prerequisites for agent-sre migration to Notifi infrastructure.


## 📋 User Story

**As a** DevOps Engineer  
**I want to** verify dependencies and prerequisites for agent-sre migration  
**So that** I can improve system reliability, security, and performance

---


## 📋 Blocking Dependencies

⚠️ **knative-lambda-operator** must be migrated first (agent-sre depends on LambdaAgent CRD)

## 🔧 Tasks

### 1\. Operator Dependency Verification

- [ ] Verify knative-lambda-operator is deployed in Notifi
- [ ] Verify LambdaAgent CRD is available
- [ ] Test LambdaAgent resource creation
- [ ] Verify operator can manage LambdaAgent resources

### 2\. LLM Provider Verification

- [ ] Verify Ollama service availability in Notifi
- [ ] Verify Ollama endpoint and port configuration
- [ ] Test Ollama API connectivity
- [ ] Verify model storage/accessibility in Ollama
- [ ] Identify alternative LLM providers if Ollama unavailable

### 3\. Storage Backend Verification

- [ ] Verify MinIO/S3 service availability
- [ ] Verify storage credentials and access
- [ ] Test storage read/write permissions
- [ ] Verify model artifact storage path structure
- [ ] Test model upload/download process

### 4\. Observability Stack Verification

- [ ] Verify Prometheus access and endpoint
- [ ] Verify Loki access and endpoint
- [ ] Test Prometheus query API
- [ ] Test Loki query API
- [ ] Verify record rules availability (if used)

### 5\. Training Pipeline Assessment

- [ ] Verify Flyte availability (if training pipeline needed)
- [ ] Identify alternative workflow engines if Flyte unavailable
- [ ] Assess need for training pipeline migration
- [ ] Document training pipeline requirements

### 6\. Network and Security

- [ ] Verify network connectivity between services
- [ ] Verify service account permissions
- [ ] Verify RBAC permissions needed
- [ ] Identify any network policies needed

## ✅ Acceptance Criteria

- [ ] All critical dependencies verified available
- [ ] Configuration endpoints documented
- [ ] Access credentials identified
- [ ] Network connectivity verified
- [ ] Migration blockers identified and documented
- [ ] Decision made on optional dependencies (training pipeline)

## 📝 Deliverables

* Dependency verification report
* Configuration requirements document
* Migration blocker list (if any)

## 🔗 Dependencies

INFRA-MIG-004: Deploy and validate knative-lambda-operator in dev environment (MUST BE COMPLETE FIRST)
