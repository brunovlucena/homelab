# 🔄 DEVOPS-017: Migrate agent-sre container images to Notifi registry

**Status**: Backlog  | **Priority**: P2**Linear URL**: https://linear.app/bvlucena/issue/BVL-11/migration-migrate-agent-sre-container-images-to-notifi-registry | **Status**: Backlog  | **Priority**: P2**Linear URL**: https://linear.app/bvlucena/issue/BVL-11/migration-migrate-agent-sre-container-images-to-notifi-registry | **Story Points**: 13

**Created**: 2025-12-26T14:37:11.189Z  
**Updated**: 2025-12-26T14:37:11.189Z  
**Project**: knative-lambda-operator  

---

# 🎯 Objective

Migrate agent-sre container images from homelab registry to Notifi container registry.


## 📋 User Story

**As a** DevOps Engineer  
**I want to** migrate agent-sre container images to notifi registry  
**So that** I can improve system reliability, security, and performance

---


## 📋 Current State

* **Source Registry:** `ghcr.io/brunovlucena/agent-sre`
* **Local Registry:** `localhost:5001/agent-sre`
* **Current Version:** 0.2.0

## 🔧 Tasks

### 1\. Registry Setup

- [ ] Verify Notifi container registry endpoint
- [ ] Configure registry authentication/credentials
- [ ] Update CI/CD pipeline for new registry (if applicable)

### 2\. Image Migration

- [ ] Build and push current version (0.2.0) to Notifi registry
- [ ] Tag appropriately for environments (dev, prd)
- [ ] Verify image accessibility and permissions
- [ ] Update LambdaAgent manifest with new image registry

### 3\. Update Build Process

- [ ] Update Makefile image build targets
- [ ] Update Dockerfile if needed for multi-arch builds
- [ ] Update versioning strategy to align with Notifi standards
- [ ] Test build process end-to-end

### 4\. Update References

- [ ] Update LambdaAgent YAML manifests with new image reference
- [ ] Update any hardcoded image references in code
- [ ] Update documentation with new registry information

### 5\. Model Artifacts (if applicable)

- [ ] Verify model artifacts storage location
- [ ] Update model path references if needed
- [ ] Verify model download process works with new storage

## ✅ Acceptance Criteria

- [ ] Images available in Notifi registry
- [ ] LambdaAgent uses new registry references
- [ ] Images pull successfully in target cluster
- [ ] Build process works with new registry
- [ ] Model artifacts accessible (if applicable)
- [ ] Documentation updated

## 🔗 Dependencies

INFRA-MIG-006: Verify dependencies and prerequisites for agent-sre migration
