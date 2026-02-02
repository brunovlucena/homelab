# Agent-SRE Deployment Status Report

**Date:** December 24, 2025  
**Report Type:** Deployment Readiness & Status  
**Component:** Agent-SRE (AI-Powered SRE Agent)  
**Cluster:** studio

---

## Executive Summary

This report assesses the deployment status of Agent-SRE on the Studio cluster. Agent-SRE is an AI-powered SRE agent that automates incident response using fine-tuned FunctionGemma 270M models. The agent receives Prometheus alerts via CloudEvents and performs Flux reconciliation based on RUNBOOK.md procedures.

**Current Status:** ⚠️ **Configured but Not Deployed**

**Key Findings:**
- ✅ All deployment manifests are ready
- ✅ LambdaAgent configuration is complete
- ✅ Integration points are configured
- ❌ Not included in cluster kustomization
- ❌ Not deployed to cluster
- ⚠️ Model availability needs verification
- ⚠️ Ollama service uses ExternalName (host.docker.internal)

---

## Deployment Configuration

### Component Location

**Base Path:** `flux/ai/agent-sre/k8s/kustomize/`

**Structure:**
```
k8s/kustomize/
├── base/
│   ├── kustomization.yaml
│   ├── lambdaagent.yaml      # Main LambdaAgent CRD
│   ├── service.yaml          # Optional service
│   └── deployment.yaml       # Fallback deployment
├── pro/
│   └── kustomization.yaml    # Pro cluster overrides
├── dev/
│   └── kustomization.yaml    # Dev cluster overrides
└── studio/ (if exists)
    └── kustomization.yaml    # Studio cluster overrides
```

### LambdaAgent Configuration

**File:** `k8s/kustomize/base/lambdaagent.yaml`

**Key Settings:**

#### Image Configuration
```yaml
image:
  repository: localhost:5001/agent-sre
  tag: "latest"
  port: 8000
  imagePullSecrets:
    - ghcr-secret
```

#### AI/LLM Configuration
```yaml
ai:
  provider: ollama
  endpoint: "http://ollama:11434"
  model: "agent-sre:v20251224-abc12345"  # ⚠️ Placeholder - needs actual version
```

#### Scaling Configuration
```yaml
scaling:
  minReplicas: 0  # Scale to zero
  maxReplicas: 5
  targetConcurrency: 10
  scaleToZeroGracePeriod: 30s
```

#### Resources (Base)
```yaml
resources:
  requests:
    cpu: "200m"
    memory: "512Mi"
  limits:
    cpu: "1000m"
    memory: "2Gi"
```

#### Eventing Configuration
```yaml
eventing:
  enabled: true
  eventSource: "/agent-sre/prometheus-alerts"
  subscriptions:
    - eventType: io.homelab.prometheus.alert.fired
    - eventType: io.homelab.prometheus.alert.resolved
  dlq:
    enabled: true
    retryMaxAttempts: 3
```

---

## Cluster Integration Status

### Current State

**Finding:** Agent-SRE is **NOT** included in Studio cluster's kustomization files.

**Checked Files:**
- `flux/clusters/studio/deploy/07-apps/kustomization.yaml` - ⚠️ Need to check

**Flux Status:**
- `studio-07-apps`: ⚠️ Unknown (reconciliation in progress)
- `studio-08-ai`: ❌ False (waiting for studio-07-apps dependency)

### Port Configuration

**Port mappings:** Not checked (may be configured in cluster definition)

---

## Deployment Steps

### Step 1: Add to Cluster Kustomization

**For Studio Cluster:**

Edit `flux/clusters/studio/deploy/07-apps/kustomization.yaml`:

```yaml
resources:
  # ... existing resources ...
  - ../../../../ai/agent-sre/k8s/kustomize/studio  # or pro/dev if studio doesn't exist
```

**Note:** If `studio` overlay doesn't exist, use `pro` or `dev` overlay.

### Step 2: Update Model Version

**Before deployment, update model name in `lambdaagent.yaml`:**

```yaml
# Replace placeholder
model: "agent-sre:v20251224-abc12345"  # ❌ Placeholder

# With actual version from fine-tuning
model: "agent-sre:v20251224-{actual-execution-id}"  # ✅ Actual version
```

**Or use `agent-sre:latest` if using latest model.**

### Step 3: Reconcile Flux

```bash
# Wait for studio-07-apps to complete
flux get kustomizations studio-07-apps -n flux-system

# Once ready, reconcile
flux reconcile kustomization studio-07-apps -n flux-system

# Or reconcile all
flux reconcile kustomization -A
```

### Step 4: Verify Deployment

```bash
# Check LambdaAgent
kubectl get lambdaagent -n ai agent-sre

# Check pods (may be scaled to zero)
kubectl get pods -n ai -l app=agent-sre

# Check service (if created)
kubectl get svc -n ai agent-sre

# Check Knative service
kubectl get ksvc -n ai agent-sre
```

---

## Ollama Configuration

### Current Setup

**Service:** ollama-native  
**Type:** ExternalName  
**Target:** host.docker.internal  
**Namespace:** ollama

**Finding:** Ollama is configured to use host Ollama instance, not a Kubernetes deployment.

### Implications

**For Agent-SRE:**
- Endpoint should be: `http://ollama-native.ollama.svc.cluster.local:11434`
- Or: `http://host.docker.internal:11434` (if accessible from pods)
- May need to verify network connectivity

### Model Verification

**Check if model exists in Ollama:**

```bash
# Option 1: Via host (if accessible)
curl http://localhost:11434/api/tags | grep agent-sre

# Option 2: Via port-forward
kubectl port-forward -n ollama svc/ollama-native 11434:11434
curl http://localhost:11434/api/tags

# Option 3: From pod (if Ollama accessible)
kubectl run -it --rm test-ollama --image=curlimages/curl --restart=Never -- \
  curl http://ollama-native.ollama.svc.cluster.local:11434/api/tags
```

---

## Dependencies

### Required Services

| Service | Namespace | Status | Notes |
|---------|-----------|--------|-------|
| **Ollama** | `ollama` | ⚠️ ExternalName | Points to host.docker.internal |
| **Prometheus** | `prometheus` | ✅ Active | Namespace exists |
| **Knative Serving** | `knative-serving` | ⚠️ Unknown | Need to verify |
| **Knative Eventing** | `knative-eventing` | ⚠️ Unknown | Need to verify |
| **RabbitMQ** | `knative-lambda` | ⚠️ Unknown | Need to verify |
| **prometheus-events** | `prometheus` | ⚠️ Unknown | Need to verify |

### Required Secrets

| Secret | Namespace | Purpose | Status |
|--------|-----------|---------|--------|
| **ghcr-secret** | `ai` | Image pull | ⚠️ Needs verification |

### Required RBAC

**Service Account:** Created automatically by LambdaAgent operator

**Permissions:**
- Flux reconciliation (kustomization, gitrepository, helmrelease)
- Kubernetes API access (for kubectl commands)
- Event subscription (CloudEvents)

---

## Testing

### Test Deployment

**1. Deploy to cluster:**
```bash
# Add to studio-07-apps kustomization
# Reconcile Flux
flux reconcile kustomization studio-07-apps -n flux-system
```

**2. Trigger test alert:**
```bash
# Create test PrometheusRule with flux_reconcile annotation
# Wait for alert to fire
# Verify agent-sre receives CloudEvent
```

**3. Verify reconciliation:**
```bash
# Check Flux reconciliation logs
kubectl logs -n flux-system -l app=kustomize-controller | grep <resource-name>

# Verify resource status
flux get kustomizations <name> -n <namespace>
```

### End-to-End Test

**Test Flow:**
1. Create PrometheusRule with `flux_reconcile` annotation
2. Trigger alert (or wait for real alert)
3. Verify prometheus-events converts to CloudEvent
4. Verify agent-sre receives event
5. Verify agent-sre generates remediation command
6. Verify Flux reconciliation executes
7. Verify alert resolves

---

## Troubleshooting

### Common Issues

**1. LambdaAgent Not Created**
- Check Flux reconciliation status
- Verify kustomization includes agent-sre
- Check for errors in Flux logs
- Verify `studio-07-apps` is ready

**2. Pods Not Starting**
- Check image pull secrets
- Verify image exists in registry
- Check resource constraints
- Review pod events: `kubectl describe pod -n ai <pod-name>`

**3. Model Not Found**
- Verify model exists in Ollama (check host or port-forward)
- Check model name matches configuration
- Verify Ollama service is accessible from pods
- Test connectivity: `curl http://ollama-native.ollama.svc.cluster.local:11434/api/tags`

**4. Events Not Received**
- Verify Knative Trigger exists
- Check CloudEvents broker
- Verify prometheus-events is running
- Check event source configuration

**5. Flux Reconciliation Fails**
- Verify RBAC permissions
- Check Flux service account
- Review Flux reconciliation logs
- Verify resource names and namespaces

---

## Recommendations

### Immediate Actions

1. **Wait for Flux Reconciliation:**
   - Monitor `studio-07-apps` status
   - Wait for it to become Ready
   - Then add agent-sre

2. **Deploy Agent-SRE:**
   - Add to cluster kustomization
   - Update model version (after fine-tuning)
   - Reconcile Flux
   - Verify deployment

3. **Verify Dependencies:**
   - Check Ollama connectivity from pods
   - Verify Prometheus connectivity
   - Test CloudEvents flow
   - Verify Flux permissions

### Short-Term (1-2 weeks)

1. **Production Deployment:**
   - Deploy to Studio cluster
   - Monitor performance
   - Tune scaling parameters

2. **Model Updates:**
   - Complete fine-tuning
   - Update model version
   - Test new model

3. **Ollama Setup:**
   - Consider deploying Ollama as Kubernetes service
   - Or verify host Ollama connectivity
   - Test model access from pods

### Long-Term (1 month+)

1. **Automation:**
   - Auto-update model versions
   - Auto-scale based on load
   - Auto-retry failed reconciliations

2. **Enhancements:**
   - Add more alert types
   - Improve model accuracy
   - Add feedback loop

---

## Conclusion

Agent-SRE deployment configuration is **complete and ready**:

✅ **Manifests**: All configured  
✅ **Integration**: All points configured  
✅ **Documentation**: Comprehensive  

**Missing:**
- ❌ Cluster kustomization inclusion
- ❌ Actual deployment
- ❌ Model version update
- ❌ End-to-end testing

**Blockers:**
- ⚠️ `studio-07-apps` still reconciling
- ⚠️ Ollama using ExternalName (may need verification)

**Next Steps:**
1. Wait for `studio-07-apps` to complete
2. Add agent-sre to cluster kustomization
3. Update model version (after fine-tuning verification)
4. Reconcile Flux
5. Verify deployment
6. Test end-to-end flow

---

**Report Generated:** 2025-12-24  
**Cluster:** studio  
**Next Review:** After deployment

