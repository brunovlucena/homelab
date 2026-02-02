# Homelab Current State Report

**Date:** December 24, 2025  
**Report Type:** Infrastructure Status Assessment  
**Environment:** Studio Cluster (New Deployment)  
**Cluster:** studio (4 nodes: 1 control-plane + 3 workers)

---

## Executive Summary

This report provides a comprehensive assessment of the current state of the homelab infrastructure on the newly created Studio cluster, focusing on:
- Cluster status and configuration
- Agent-SRE deployment readiness
- Fine-tuning pipeline status
- Model availability in Ollama

**Key Findings:**
- ✅ Studio cluster is running (4 nodes, 33 minutes old)
- ✅ Flux CD is operational and reconciling
- ✅ Core namespaces exist (flux-system, ollama, prometheus)
- ⚠️ Agent-SRE is not currently deployed
- ⚠️ Ollama service exists but pods not running
- ⚠️ Fine-tuning pipeline exists but execution status unknown
- ⚠️ Fine-tuned model availability in Ollama needs verification

---

## Cluster Status

### Cluster Overview

**Cluster Name:** studio  
**Age:** 33 minutes  
**Kubernetes Version:** v1.34.0

| Node | Role | Status | Age |
|------|------|--------|-----|
| studio-control-plane | control-plane | Ready | 33m |
| studio-worker | worker | Ready | 33m |
| studio-worker2 | worker | Ready | 33m |
| studio-worker3 | worker | Ready | 33m |

**Total Nodes:** 4 (1 control-plane + 3 workers)

### Namespace Status

| Namespace | Status | Age | Purpose |
|-----------|--------|-----|---------|
| flux-system | Active | 32m | GitOps (Flux CD) |
| ollama | Active | 25m | LLM service |
| prometheus | Active | 9m32s | Monitoring stack |

---

## Flux CD Status

### Git Repository

**Repository:** homelab  
**Revision:** main@sha1:4b470e93  
**Status:** ✅ Ready  
**Message:** stored artifact for revision 'main@sha1:4b470e93'

### Kustomizations

| Kustomization | Namespace | Status | Ready | Age |
|---------------|-----------|--------|-------|-----|
| studio-01-core | flux-system | ✅ True | Applied | 31m |
| studio-02-observability | flux-system | ✅ True | Applied | 31m |
| studio-02b-observability-extras | flux-system | ✅ True | Applied | 31m |
| studio-03-knative-deps | flux-system | ✅ True | Applied | 31m |
| studio-04-knative-instances | flux-system | ✅ True | Applied | 31m |
| studio-05-testing | flux-system | ✅ True | Applied | 31m |
| studio-06-ci | flux-system | ✅ True | Applied | 31m |
| studio-07-apps | flux-system | ⚠️ Unknown | In progress | 31m |
| studio-08-ai | flux-system | ❌ False | Waiting | 19m |

**Finding:** 
- Most kustomizations are ready
- `studio-07-apps` is still reconciling
- `studio-08-ai` is waiting for `studio-07-apps` dependency

---

## Agent-SRE Deployment Status

### Current State

| Component | Status | Notes |
|-----------|--------|-------|
| LambdaAgent CRD | ⚠️ Unknown | Need to verify if operator is installed |
| Kustomization | ❌ Not deployed | Not found in cluster |
| Pods | ❌ Not running | No pods in `ai` namespace |
| Service | ❌ Not created | Deployment required first |
| Ollama Model | ⚠️ Unknown | Needs verification |

### Deployment Configuration

**Location:** `flux/ai/agent-sre/k8s/kustomize/`

**Configuration Files:**
- ✅ `base/lambdaagent.yaml` - Main LambdaAgent manifest
- ✅ `base/kustomization.yaml` - Base kustomization
- ✅ `pro/kustomization.yaml` - Pro cluster overrides
- ✅ `dev/kustomization.yaml` - Dev cluster overrides

**Model Configuration:**
- Model name: `agent-sre:v20251224-abc12345` (placeholder)
- Provider: ollama
- Endpoint: `http://ollama:11434`

### Required Actions

1. **Add to Cluster Kustomization:**
   - Edit `flux/clusters/studio/deploy/07-apps/kustomization.yaml`
   - Add: `- ../../../../ai/agent-sre/k8s/kustomize/studio` (or pro/dev)
   - Commit and push

2. **Reconcile Flux:**
   ```bash
   flux reconcile kustomization studio-07-apps -n flux-system
   ```

3. **Verify Deployment:**
   ```bash
   kubectl get lambdaagent -n ai agent-sre
   kubectl get pods -n ai -l app=agent-sre
   ```

---

## Ollama Service Status

### Service Configuration

**Service:** ollama-native  
**Type:** ExternalName  
**Target:** host.docker.internal  
**Namespace:** ollama  
**Age:** 25 minutes

**Finding:** Service exists but points to external host, not a Kubernetes deployment.

### Pod Status

**Pods in ollama namespace:** None found

**Possible Issues:**
- Ollama may be running on host (via ExternalName service)
- No Kubernetes deployment for Ollama
- Service may need to be updated to use a proper deployment

### Model Verification

**Cannot verify models** - Ollama pods not accessible via kubectl exec.

**Alternative Verification:**
```bash
# If Ollama is on host
curl http://localhost:11434/api/tags | grep agent-sre

# Or via port-forward if service exists
kubectl port-forward -n ollama svc/ollama-native 11434:11434
curl http://localhost:11434/api/tags
```

---

## Fine-Tuning Pipeline Status

### Pipeline Configuration

**Location:** `flux/infrastructure/flyte/workflows/test/workflows/agent_training.py`

**Status:** ✅ **Configured** | ⚠️ **Execution Status Unknown**

### Pipeline Components

| Step | Component | Status |
|------|-----------|--------|
| 1. Dataset Prep | `prepare_dataset()` | ✅ Implemented |
| 2. Model Conversion | `convert_model_to_mlx()` | ✅ Implemented |
| 3. LoRA Training | `train_model_lora()` | ✅ Implemented |
| 4. Evaluation | `evaluate_model()` | ✅ Implemented |
| 5. MLflow Registration | `register_model_mlflow()` | ✅ Implemented |
| 6. MinIO Storage | `store_model_minio()` | ✅ Implemented |
| 7. Ollama Export | `export_model_ollama()` | ✅ Implemented |
| 8. Auto-Import | `trigger_ollama_import()` | ✅ Implemented |

### Training Data Source

**Source:** `flux/ai/agent-sre/docs/RUNBOOK.md`
- Total lines: 2923
- Alert sections: 100+
- Coverage: All homelab services

### Required Verification

1. **Check Flyte workflow executions:**
   ```bash
   flytectl get executions --project homelab --domain production
   ```

2. **Check for training artifacts in MinIO:**
   ```bash
   kubectl exec -n minio deployment/minio -- mc ls minio/ml-models/agent-sre/
   ```

3. **Check Kubernetes import jobs:**
   ```bash
   kubectl get jobs -n ai -l component=ollama-import
   ```

---

## Integration Points

### Prometheus Integration

**Configuration:** `infrastructure/prometheus-operator/k8s/prometheusrules/agent-sre-triggers.yaml`

**Flow:**
```
Prometheus → Alertmanager → prometheus-events → CloudEvents → agent-sre → Flux Reconciliation
```

**Status:** ⚠️ Configuration exists but agent-sre not deployed to receive events

### CloudEvents Integration

**Event Types:**
- `io.homelab.prometheus.alert.fired`
- `io.homelab.prometheus.alert.resolved`

**Status:** ⚠️ Configured but agent-sre not deployed

---

## Dependencies

### Required Services

| Service | Namespace | Status | Notes |
|---------|-----------|--------|-------|
| Ollama | ollama | ⚠️ ExternalName | Points to host.docker.internal |
| Prometheus | prometheus | ✅ Active | Namespace exists |
| Knative Serving | knative-serving | ⚠️ Unknown | Need to verify |
| Knative Eventing | knative-eventing | ⚠️ Unknown | Need to verify |
| RabbitMQ | knative-lambda | ⚠️ Unknown | Need to verify |

### Required Secrets

| Secret | Namespace | Status |
|--------|-----------|--------|
| ghcr-secret | ai | ⚠️ Unknown |

---

## Recommendations

### Immediate Actions

1. **Deploy Agent-SRE:**
   - Add to `studio-07-apps` kustomization
   - Wait for reconciliation to complete
   - Verify LambdaAgent creation

2. **Verify Ollama:**
   - Check if Ollama is running on host
   - Or deploy Ollama as Kubernetes deployment
   - Verify model availability

3. **Update Model Reference:**
   - Replace placeholder model name with actual version
   - Or use `agent-sre:latest` if available

### Short-Term (1-2 weeks)

1. **Run Fine-Tuning Pipeline:**
   - Execute Flyte workflow
   - Monitor training progress
   - Verify model export to Ollama

2. **Test Agent-SRE:**
   - Trigger test Prometheus alert
   - Verify CloudEvents flow
   - Test Flux reconciliation

3. **Fix Ollama Deployment:**
   - Deploy Ollama as proper Kubernetes service
   - Or verify host Ollama connectivity
   - Test model access

### Long-Term (1 month+)

1. **Automation:**
   - Auto-trigger training on RUNBOOK.md changes
   - Auto-update model version in LambdaAgent
   - CI/CD integration

2. **Monitoring:**
   - Track model performance
   - Monitor reconciliation success rate
   - Alert on model failures

---

## Conclusion

The Studio cluster is **newly deployed and operational**:

✅ **Cluster**: Running (4 nodes)  
✅ **Flux CD**: Operational and reconciling  
✅ **Core Infrastructure**: Namespaces created  
✅ **Agent-SRE Configuration**: Ready for deployment  

**Missing:**
- ❌ Agent-SRE deployment
- ❌ Ollama pods (using ExternalName to host)
- ❌ Fine-tuning execution verification
- ❌ Model availability confirmation

**Next Steps:**
1. Wait for `studio-07-apps` reconciliation to complete
2. Add agent-sre to kustomization
3. Verify Ollama connectivity
4. Deploy agent-sre
5. Test end-to-end flow

---

**Report Generated:** 2025-12-24  
**Cluster:** studio  
**Next Review:** After agent-sre deployment

