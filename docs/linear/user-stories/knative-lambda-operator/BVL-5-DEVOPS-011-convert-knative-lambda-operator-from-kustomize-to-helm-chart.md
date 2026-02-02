# 🔄 DEVOPS-011: Convert knative-lambda-operator from Kustomize to Helm Chart

**Status**: Backlog  | **Priority**: P0**Linear URL**: https://linear.app/bvlucena/issue/BVL-5/migration-convert-knative-lambda-operator-from-kustomize-to-helm-chart | **Status**: Backlog  | **Priority**: P0**Linear URL**: https://linear.app/bvlucena/issue/BVL-5/migration-convert-knative-lambda-operator-from-kustomize-to-helm-chart | **Story Points**: 13

**Created**: 2025-12-26T14:36:12.538Z  
**Updated**: 2025-12-26T14:36:12.538Z  
**Project**: knative-lambda-operator  

---

# 🎯 Objective

Convert the knative-lambda-operator deployment from Kustomize-based to Helm chart following Notifi infrastructure patterns.


## 📋 User Story

**As a** DevOps Engineer  
**I want to** convert knative-lambda-operator from kustomize to helm chart  
**So that** I can improve system reliability, security, and performance

---


## 📋 Current State

* **Location:** `/bruno/repos/homelab/flux/infrastructure/knative-lambda-operator/`
* **Deployment:** Kustomize (base + overlays for pro/studio)
* **Version:** 1.13.11

## 🔧 Tasks

### 1\. Create Helm Chart Structure

- [ ] Create `deploy/Chart.yaml` following Notifi service patterns
- [ ] Create `deploy/values.yaml` with base configuration
- [ ] Create `deploy/templates/_helpers.tpl` for reusable templates

### 2\. Convert Kustomize Resources to Helm Templates

- [ ] Convert `k8s/base/namespace.yaml` → `deploy/templates/namespace.yaml`
- [ ] Convert `k8s/base/crd.yaml` → `deploy/templates/crd.yaml`
- [ ] Convert `k8s/base/crd-lambdaagent.yaml` → `deploy/templates/crd-lambdaagent.yaml`
- [ ] Convert `k8s/base/rbac.yaml` → `deploy/templates/rbac.yaml`
- [ ] Convert `k8s/base/agent-rbac.yaml` → `deploy/templates/agent-rbac.yaml`
- [ ] Convert `k8s/base/security-rbac.yaml` → `deploy/templates/security-rbac.yaml`
- [ ] Convert `k8s/base/deployment.yaml` → `deploy/templates/deployment.yaml`
- [ ] Convert `k8s/base/service.yaml` → `deploy/templates/service.yaml`
- [ ] Convert `k8s/base/lambda-command-receiver.yaml` → `deploy/templates/lambda-command-receiver.yaml`
- [ ] Convert `k8s/base/minio-secret-init.yaml` → `deploy/templates/minio-secret-init.yaml`
- [ ] Convert `k8s/base/ghcr-secret-init.yaml` → `deploy/templates/ghcr-secret-init.yaml`
- [ ] Convert `k8s/base/knative-serving-config.yaml` → `deploy/templates/knative-serving-config.yaml`

### 3\. Create Environment Overlays

- [ ] Create `deploy/overlays/local/values-local.yaml` for local development
- [ ] Create `deploy/overlays/dev/values-dev.yaml` for development environment
- [ ] Create `deploy/overlays/prd/values-prd.yaml` for production environment
- [ ] Map pro/studio overlays to dev/prd respectively

### 4\. Configuration Migration

- [ ] Extract image version from VERSION file to values.yaml
- [ ] Extract namespace configuration to values.yaml
- [ ] Extract replica counts to values.yaml (environment-specific)
- [ ] Extract resource limits/requests to values.yaml
- [ ] Extract environment variables to values.yaml
- [ ] Map canary/flagger configs to production values

## ✅ Acceptance Criteria

- [ ] Helm chart validates with `helm lint`
- [ ] Chart templates render correctly with `helm template`
- [ ] All resources match existing Kustomize output (diff validation)
- [ ] Chart follows Notifi service patterns (see knative-lambda, loki services for reference)
- [ ] Environment overlays work correctly
- [ ] Documentation updated

## 📚 References

* Notifi service pattern: `20-platform/services/knative-lambda/deploy/`
* Notifi service pattern: `20-platform/services/loki/deploy/`
* Current operator structure: `knative-lambda-operator/k8s/`

## 🔗 Dependencies

None (foundational migration)
