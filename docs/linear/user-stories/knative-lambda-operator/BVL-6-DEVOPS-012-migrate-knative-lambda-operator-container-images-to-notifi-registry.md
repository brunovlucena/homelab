# 🔄 DEVOPS-012: Migrate knative-lambda-operator container images to Notifi registry

**Status**: Backlog  | **Priority**: P1**Linear URL**: https://linear.app/bvlucena/issue/BVL-6/migration-migrate-knative-lambda-operator-container-images-to-notifi-registry | **Status**: Backlog  | **Priority**: P1**Linear URL**: https://linear.app/bvlucena/issue/BVL-6/migration-migrate-knative-lambda-operator-container-images-to-notifi-registry | **Story Points**: 8

**Created**: 2025-12-26T14:36:19.386Z  
**Updated**: 2025-12-26T14:36:19.386Z  
**Project**: knative-lambda-operator  

---

# 🎯 Objective

Migrate knative-lambda-operator container images from homelab registry to Notifi container registry.


## 📋 User Story

**As a** DevOps Engineer  
**I want to** migrate knative-lambda-operator container images to notifi registry  
**So that** I can improve system reliability, security, and performance

---


## 📋 Current State

* **Source Registry:** `ghcr.io/brunovlucena/knative-lambda-operator`
* **Local Registry:** `localhost:5001/knative-lambda-operator`
* **Current Version:** 1.13.11

## 🔧 Tasks

### 1\. Registry Setup (if needed)

- [ ] Identify Notifi container registry endpoint
- [ ] Configure registry authentication/credentials
- [ ] Update CI/CD pipeline for new registry (if applicable)

### 2\. Image Migration

- [ ] Build and push current version (1.13.11) to Notifi registry
- [ ] Tag appropriately for environments (dev, prd)
- [ ] Verify image accessibility and permissions
- [ ] Update Helm chart values.yaml with new image registry

### 3\. Update Build Process

- [ ] Update Makefile image build targets
- [ ] Update Dockerfile if needed for multi-arch builds
- [ ] Update versioning strategy to align with Notifi standards
- [ ] Test build process end-to-end

### 4\. Update References

- [ ] Update Helm templates with new image reference
- [ ] Update any hardcoded image references in code
- [ ] Update documentation with new registry information

## ✅ Acceptance Criteria

- [ ] Images available in Notifi registry
- [ ] Helm chart uses new registry references
- [ ] Images pull successfully in target cluster
- [ ] Build process works with new registry
- [ ] Documentation updated

## 🔗 Dependencies

INFRA-MIG-001: Convert knative-lambda-operator from Kustomize to Helm Chart
