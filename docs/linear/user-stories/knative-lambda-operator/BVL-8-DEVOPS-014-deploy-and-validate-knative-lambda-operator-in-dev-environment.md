# 🔄 DEVOPS-014: Deploy and validate knative-lambda-operator in dev environment

**Status**: Backlog  | **Priority**: P1**Linear URL**: https://linear.app/bvlucena/issue/BVL-8/migration-deploy-and-validate-knative-lambda-operator-in-dev-environment | **Status**: Backlog  | **Priority**: P1**Linear URL**: https://linear.app/bvlucena/issue/BVL-8/migration-deploy-and-validate-knative-lambda-operator-in-dev-environment | **Story Points**: 13

**Created**: 2025-12-26T14:36:40.027Z  
**Updated**: 2025-12-26T14:36:40.027Z  
**Project**: knative-lambda-operator  

---

# 🎯 Objective

Deploy knative-lambda-operator to Notifi dev environment and validate all functionality works correctly.


## 📋 User Story

**As a** DevOps Engineer  
**I want to** deploy and validate knative-lambda-operator in dev environment  
**So that** I can improve system reliability, security, and performance

---


## 📋 Prerequisites

* Helm chart created and validated
* Container images migrated to Notifi registry
* ArgoCD ApplicationSet created (or manual deployment)

## 🔧 Tasks

### 1\. Pre-Deployment Verification

- [ ] Verify cluster access to dev environment
- [ ] Verify namespace exists or will be created
- [ ] Verify dependencies (Knative, RabbitMQ, Prometheus) are available
- [ ] Verify storage backend (MinIO/S3) accessibility
- [ ] Verify image registry accessibility

### 2\. Initial Deployment

- [ ] Deploy operator via Helm chart or ArgoCD
- [ ] Verify namespace creation
- [ ] Verify CRD installation (LambdaFunction, LambdaAgent)
- [ ] Verify operator pod starts successfully
- [ ] Verify RBAC permissions

### 3\. CRD Validation

- [ ] Verify LambdaFunction CRD is installed and valid
- [ ] Verify LambdaAgent CRD is installed and valid
- [ ] Test CRD schema validation
- [ ] Verify CRD finalizers work correctly

### 4\. Operator Functionality Tests

- [ ] Test LambdaFunction creation and reconciliation
- [ ] Test LambdaFunction build process (Kaniko job creation)
- [ ] Test LambdaFunction deployment (Knative Service creation)
- [ ] Test LambdaAgent creation and reconciliation
- [ ] Test LambdaAgent deployment
- [ ] Test CloudEvents integration
- [ ] Test RabbitMQ broker integration

### 5\. Observability Validation

- [ ] Verify Prometheus metrics are exposed
- [ ] Verify ServiceMonitor is working
- [ ] Verify metrics are scraped correctly
- [ ] Check Grafana dashboard (if exists)
- [ ] Verify distributed tracing (if configured)
- [ ] Verify logs are accessible

### 6\. Integration Tests

- [ ] Run operator unit tests in dev environment
- [ ] Run integration tests
- [ ] Run E2E tests (if applicable)
- [ ] Test k6 load tests (if applicable)

### 7\. Security Validation

- [ ] Verify RBAC permissions are correct
- [ ] Verify service account permissions
- [ ] Verify network policies (if applicable)
- [ ] Verify secret management
- [ ] Run security scans on images

## ✅ Acceptance Criteria

- [ ] Operator deployed successfully in dev
- [ ] All CRDs functional
- [ ] LambdaFunction resource works end-to-end
- [ ] LambdaAgent resource works end-to-end
- [ ] Observability stack integrated
- [ ] All tests passing
- [ ] No critical security issues
- [ ] Documentation updated with dev deployment steps

## 🔗 Dependencies

INFRA-MIG-001: Convert knative-lambda-operator from Kustomize to Helm Chart
INFRA-MIG-002: Migrate knative-lambda-operator container images to Notifi registry
