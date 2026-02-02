# 🔄 DEVOPS-018: Update agent-sre configuration for Notifi environment

**Status**: Backlog  | **Priority**: P2**Linear URL**: https://linear.app/bvlucena/issue/BVL-12/migration-update-agent-sre-configuration-for-notifi-environment | **Status**: Backlog  | **Priority**: P2**Linear URL**: https://linear.app/bvlucena/issue/BVL-12/migration-update-agent-sre-configuration-for-notifi-environment | **Story Points**: 13

**Created**: 2025-12-26T14:37:19.939Z  
**Updated**: 2025-12-26T14:37:19.939Z  
**Project**: knative-lambda-operator  

---

# 🎯 Objective

Update agent-sre configuration files and code to work with Notifi infrastructure endpoints and settings.


## 📋 User Story

**As a** DevOps Engineer  
**I want to** update agent-sre configuration for notifi environment  
**So that** I can improve system reliability, security, and performance

---


## 📋 Configuration Areas

* Ollama endpoint and model configuration
* Prometheus/Loki endpoint configuration
* MinIO/S3 storage configuration
* Environment-specific settings

## 🔧 Tasks

### 1\. Ollama Configuration

- [ ] Update Ollama endpoint in agent-sre config
- [ ] Update model name/version references
- [ ] Verify Ollama API compatibility
- [ ] Test model loading and inference
- [ ] Update LambdaAgent AI configuration section

### 2\. Observability Configuration

- [ ] Update Prometheus endpoint configuration
- [ ] Update Loki endpoint configuration
- [ ] Update query timeouts if needed
- [ ] Verify authentication/authorization
- [ ] Test metrics collection
- [ ] Test log query functionality

### 3\. Storage Configuration

- [ ] Update MinIO/S3 endpoint configuration
- [ ] Update bucket names and paths
- [ ] Update credentials/secret references
- [ ] Verify storage access permissions
- [ ] Test model artifact storage/retrieval

### 4\. Environment-Specific Configuration

- [ ] Create dev environment configuration
- [ ] Create production environment configuration
- [ ] Update namespace references
- [ ] Update resource limits/requests
- [ ] Configure scaling parameters

### 5\. LambdaAgent Manifest Updates

- [ ] Update LambdaAgent YAML with new configuration
- [ ] Update image references
- [ ] Update environment variables
- [ ] Update scaling configuration
- [ ] Update CloudEvents subscriptions

### 6\. Code Configuration

- [ ] Review and update Python configuration files
- [ ] Update any hardcoded endpoints
- [ ] Update secret/credential handling
- [ ] Add environment variable support
- [ ] Test configuration loading

## ✅ Acceptance Criteria

- [ ] All endpoints updated to Notifi services
- [ ] Configuration loads correctly
- [ ] Ollama connectivity verified
- [ ] Prometheus/Loki queries work
- [ ] Storage access verified
- [ ] Environment-specific configs tested
- [ ] Documentation updated

## 📚 References

* Current config: `agent-sre/src/sre_agent/config.py`
* LambdaAgent manifest: `agent-sre/k8s/kustomize/base/lambdaagent.yaml`

## 🔗 Dependencies

INFRA-MIG-006: Verify dependencies and prerequisites for agent-sre migration
INFRA-MIG-007: Migrate agent-sre container images to Notifi registry
