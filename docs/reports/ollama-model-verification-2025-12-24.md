# Ollama Model Verification Guide

**Date:** December 24, 2025  
**Report Type:** Model Verification & Troubleshooting  
**Cluster:** studio  
**Ollama Service:** ExternalName → host.docker.internal

---

## Executive Summary

This guide provides comprehensive instructions for verifying the Agent-SRE fine-tuned model in Ollama on the Studio cluster. The cluster uses an ExternalName service pointing to `host.docker.internal`, which requires specific verification approaches.

**Current Status:**
- ✅ Ollama namespace exists
- ✅ Service configured (ExternalName)
- ⚠️ Pods not running in cluster (using host Ollama)
- ⚠️ Model availability unknown

---

## Ollama Service Configuration

### Current Setup

**Service Details:**
```yaml
Name: ollama-native
Namespace: ollama
Type: ExternalName
Target: host.docker.internal
Age: 25 minutes
```

**Finding:** Ollama is configured to use the host machine's Ollama instance, not a Kubernetes deployment.

### Service Endpoints

**From Kubernetes pods:**
- `http://ollama-native.ollama.svc.cluster.local:11434`
- May resolve to `host.docker.internal:11434`

**From host:**
- `http://localhost:11434`

---

## Verification Methods

### Method 1: Direct Host Access

**If Ollama is running on the host machine:**

```bash
# List all models
curl http://localhost:11434/api/tags

# Check for agent-sre model
curl http://localhost:11434/api/tags | jq '.models[] | select(.name | contains("agent-sre"))'

# Show model details
curl http://localhost:11434/api/show -d '{"name": "agent-sre:latest"}'

# Test model
curl http://localhost:11434/api/generate -d '{
  "model": "agent-sre:latest",
  "prompt": "Alert: FluxReconciliationFailure. How should I resolve this?",
  "stream": false
}'
```

### Method 2: Port-Forward

**Access Ollama via port-forward:**

```bash
# Forward service to localhost
kubectl port-forward -n ollama svc/ollama-native 11434:11434

# In another terminal, test
curl http://localhost:11434/api/tags | grep agent-sre
```

### Method 3: From Kubernetes Pod

**Test connectivity from a pod:**

```bash
# Create test pod
kubectl run -it --rm test-ollama --image=curlimages/curl --restart=Never -- \
  curl http://ollama-native.ollama.svc.cluster.local:11434/api/tags

# Or use busybox
kubectl run -it --rm test-ollama --image=busybox --restart=Never -- \
  wget -qO- http://ollama-native.ollama.svc.cluster.local:11434/api/tags
```

### Method 4: Check Import Jobs

**Verify if import job ran:**

```bash
# List import jobs
kubectl get jobs -n ai -l component=ollama-import

# Check job status
kubectl get jobs -n ai -l component=ollama-import -o wide

# View job logs
kubectl logs -n ai job/import-agent-sre-ollama-{execution-id}
```

---

## Expected Model Names

### Model Naming Convention

**Format:** `agent-sre:v{YYYYMMDD}-{execution_id}`

**Examples:**
- `agent-sre:v20251224-abc12345`
- `agent-sre:v20251224-def67890`
- `agent-sre:latest` (if tagged)

### Verification Commands

**List all agent-sre models:**
```bash
# Via host
curl http://localhost:11434/api/tags | jq '.models[] | select(.name | startswith("agent-sre"))'

# Via port-forward
kubectl port-forward -n ollama svc/ollama-native 11434:11434 &
curl http://localhost:11434/api/tags | jq '.models[] | select(.name | startswith("agent-sre"))'
```

---

## Model Testing

### Test 1: Basic Availability

**Check if model exists:**
```bash
curl http://localhost:11434/api/tags | grep -i "agent-sre"
```

**Expected Output:**
```json
{
  "models": [
    {
      "name": "agent-sre:v20251224-abc12345",
      "modified_at": "2025-12-24T...",
      "size": 1234567890,
      "digest": "sha256:..."
    }
  ]
}
```

### Test 2: Model Details

**Get model information:**
```bash
curl http://localhost:11434/api/show -d '{
  "name": "agent-sre:v20251224-abc12345"
}'
```

**Expected Output:**
```json
{
  "modelfile": "...",
  "parameters": "...",
  "template": "...",
  "details": {
    "parent_model": "",
    "format": "gguf",
    "family": "gemma",
    "families": ["gemma"],
    "parameter_size": "270M",
    "quantization_level": "Q4_0"
  }
}
```

### Test 3: Generate Response

**Test model with alert:**
```bash
curl http://localhost:11434/api/generate -d '{
  "model": "agent-sre:v20251224-abc12345",
  "prompt": "Alert: FluxReconciliationFailure\n\nSymptoms:\n- Flux Kustomization failing to reconcile\n- Status shows Ready=False\n\nBased on the runbook, provide the exact remediation command.",
  "stream": false,
  "options": {
    "temperature": 0.1,
    "top_p": 0.9
  }
}'
```

**Expected Output:**
```json
{
  "model": "agent-sre:v20251224-abc12345",
  "created_at": "2025-12-24T...",
  "response": "flux reconcile kustomization <name> -n <namespace>",
  "done": true
}
```

### Test 4: Interactive Test

**Use Ollama CLI (if installed on host):**
```bash
ollama run agent-sre:v20251224-abc12345 "Alert: FluxReconciliationFailure. How should I resolve this?"
```

---

## Troubleshooting

### Issue 1: Model Not Found

**Symptoms:** Model doesn't appear in `ollama list`

**Solutions:**
1. **Check if import job completed:**
   ```bash
   kubectl get jobs -n ai -l component=ollama-import
   kubectl logs -n ai job/import-agent-sre-ollama-{id}
   ```

2. **Manually import model:**
   ```bash
   # Download from MinIO
   kubectl exec -n minio deployment/minio -- mc cp minio/ml-models/agent-sre/{version}/ollama-model.tar.gz /tmp/
   
   # Extract and import
   tar -xzf /tmp/ollama-model.tar.gz
   ollama create agent-sre:{version} -f Modelfile
   ```

3. **Check MinIO for model:**
   ```bash
   kubectl exec -n minio deployment/minio -- mc ls minio/ml-models/agent-sre/
   ```

### Issue 2: Cannot Connect to Ollama

**Symptoms:** Connection refused or timeout

**Solutions:**
1. **Verify Ollama is running on host:**
   ```bash
   # On host
   curl http://localhost:11434/api/tags
   ```

2. **Check service configuration:**
   ```bash
   kubectl get svc -n ollama ollama-native -o yaml
   ```

3. **Test from pod:**
   ```bash
   kubectl run -it --rm test --image=curlimages/curl --restart=Never -- \
     curl -v http://ollama-native.ollama.svc.cluster.local:11434/api/tags
   ```

4. **Check network policies:**
   ```bash
   kubectl get networkpolicies -n ollama
   ```

### Issue 3: Model Returns Errors

**Symptoms:** Model generates incorrect or error responses

**Solutions:**
1. **Verify model was trained correctly:**
   - Check training logs
   - Verify model accuracy metrics
   - Review evaluation results

2. **Check model format:**
   ```bash
   curl http://localhost:11434/api/show -d '{"name": "agent-sre:latest"}'
   ```

3. **Test with different prompts:**
   ```bash
   curl http://localhost:11434/api/generate -d '{
     "model": "agent-sre:latest",
     "prompt": "Alert: BrunoSiteAPIDown. Symptoms: Homepage API unavailable.",
     "stream": false
   }'
   ```

### Issue 4: Import Job Failed

**Symptoms:** Import job shows Failed or Error status

**Solutions:**
1. **Check job logs:**
   ```bash
   kubectl logs -n ai job/import-agent-sre-ollama-{id}
   ```

2. **Check job events:**
   ```bash
   kubectl describe job -n ai import-agent-sre-ollama-{id}
   ```

3. **Verify MinIO access:**
   ```bash
   kubectl exec -n minio deployment/minio -- mc ls minio/ml-models/agent-sre/
   ```

4. **Check service account permissions:**
   ```bash
   kubectl get sa -n ai ollama-import-sa
   kubectl describe sa -n ai ollama-import-sa
   ```

---

## Verification Checklist

### Pre-Deployment

- [ ] Ollama service exists and is accessible
- [ ] Model exists in Ollama (check via host or port-forward)
- [ ] Model can generate responses
- [ ] Model name matches LambdaAgent configuration
- [ ] Import job completed successfully (if applicable)

### Post-Deployment

- [ ] Agent-SRE can connect to Ollama
- [ ] Agent-SRE can load model
- [ ] Model generates correct responses
- [ ] Test alert triggers correct remediation
- [ ] Flux reconciliation executes successfully

---

## Next Steps

### Immediate Actions

1. **Verify Ollama on Host:**
   ```bash
   curl http://localhost:11434/api/tags
   ```

2. **Check for Agent-SRE Model:**
   ```bash
   curl http://localhost:11434/api/tags | grep agent-sre
   ```

3. **Test Model (if exists):**
   ```bash
   curl http://localhost:11434/api/generate -d '{
     "model": "agent-sre:latest",
     "prompt": "Alert: FluxReconciliationFailure. How should I resolve this?",
     "stream": false
   }'
   ```

### If Model Not Found

1. **Check Flyte Workflow:**
   - Verify training completed
   - Check workflow output for model version

2. **Check MinIO:**
   - Verify model artifacts stored
   - Check model version

3. **Run Import Job:**
   - Manually trigger import if needed
   - Or re-run Flyte workflow

---

## Conclusion

Ollama is configured as an ExternalName service pointing to the host machine. Verification requires:

✅ **Service exists**: ollama-native in ollama namespace  
⚠️ **Model status**: Unknown - needs verification  
⚠️ **Connectivity**: Needs testing from pods  

**Recommended Approach:**
1. Verify Ollama on host first
2. Check for agent-sre model
3. Test model responses
4. Verify connectivity from Kubernetes pods
5. Update Agent-SRE configuration with correct model name

---

**Report Generated:** 2025-12-24  
**Cluster:** studio  
**Next Review:** After model verification

