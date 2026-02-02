# 🔄 DEVOPS-013: Create ArgoCD ApplicationSet for knative-lambda-operator

**Status**: Backlog  | **Priority**: P1**Linear URL**: https://linear.app/bvlucena/issue/BVL-7/migration-create-argocd-applicationset-for-knative-lambda-operator | **Status**: Backlog  | **Priority**: P1**Linear URL**: https://linear.app/bvlucena/issue/BVL-7/migration-create-argocd-applicationset-for-knative-lambda-operator | **Story Points**: 13

**Created**: 2025-12-26T14:36:28.945Z  
**Updated**: 2025-12-26T14:36:28.945Z  
**Project**: knative-lambda-operator  

---

# 🎯 Objective

Create ArgoCD ApplicationSet configuration for automated GitOps deployment of knative-lambda-operator across environments.


## 📋 User Story

**As a** DevOps Engineer  
**I want to** create argocd applicationset for knative-lambda-operator  
**So that** I can improve system reliability, security, and performance

---


## 📋 Current State

* **Deployment:** Currently using Flux CD in homelab
* **Target:** ArgoCD ApplicationSet in Notifi infrastructure

## 🔧 Tasks

### 1\. Review Existing ApplicationSet Patterns

- [ ] Review `20-platform/argocd/config/{local,dev,prd}/applicationsets/` for patterns
- [ ] Identify appropriate ApplicationSet type (list, cluster, git generator)
- [ ] Review existing service ApplicationSets for reference

### 2\. Create ApplicationSet Manifest

- [ ] Create ApplicationSet YAML following Notifi patterns
- [ ] Configure for multiple environments (local, dev, prd)
- [ ] Set up git generator or list generator as appropriate
- [ ] Configure destination clusters/namespaces

### 3\. Configure Helm Values

- [ ] Ensure environment-specific values files exist
- [ ] Configure sync policies (automated/manual)
- [ ] Configure sync windows if needed
- [ ] Set up health checks and sync options

### 4\. RBAC and Permissions

- [ ] Verify ArgoCD has permissions to deploy operator
- [ ] Configure ApplicationSet RBAC if needed
- [ ] Test deployment permissions

### 5\. Testing and Validation

- [ ] Test ApplicationSet creation
- [ ] Verify applications are created correctly
- [ ] Test sync process
- [ ] Verify operator deployment in target environments

## ✅ Acceptance Criteria

- [ ] ApplicationSet creates applications for all target environments
- [ ] Applications sync successfully
- [ ] Operator deploys correctly via ArgoCD
- [ ] Follows Notifi ApplicationSet patterns
- [ ] Documentation updated

## 📚 References

* Existing ApplicationSets: `20-platform/argocd/config/{env}/applicationsets/`
* ArgoCD documentation for ApplicationSets

## 🔗 Dependencies

INFRA-MIG-001: Convert knative-lambda-operator from Kustomize to Helm Chart
INFRA-MIG-002: Migrate knative-lambda-operator container images to Notifi registry
