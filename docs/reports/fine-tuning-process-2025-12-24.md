# Agent-SRE Fine-Tuning Process Report

**Date:** December 24, 2025  
**Report Type:** Fine-Tuning Pipeline Documentation & Status  
**Model:** FunctionGemma 270M → Agent-SRE Fine-Tuned  
**Training Data:** RUNBOOK.md (2923 lines, 100+ alert examples)  
**Cluster:** studio

---

## Executive Summary

This report documents the complete fine-tuning process for Agent-SRE, from dataset preparation through model deployment to Ollama. The process uses MLX-LM with LoRA to fine-tune FunctionGemma 270M on homelab infrastructure runbook procedures.

**Pipeline Status:** ✅ **Configured** | ⚠️ **Execution Status Unknown**

**Key Components:**
- Training Data: RUNBOOK.md (comprehensive incident response procedures)
- Base Model: FunctionGemma 270M (Google)
- Framework: MLX-LM with LoRA
- Deployment: Ollama format export
- Automation: Flyte workflow pipeline

---

## Process Overview

### Complete Pipeline Flow

```
RUNBOOK.md → Dataset Preparation → Model Conversion → LoRA Training → 
Evaluation → MLflow Registration → MinIO Storage → Ollama Export → 
Kubernetes Import Job → Ollama Model Available
```

### Pipeline Components

| Step | Component | Location | Status |
|------|-----------|----------|--------|
| 1. Dataset Prep | `prepare_dataset()` | `agent_training.py` | ✅ Implemented |
| 2. Model Conversion | `convert_model_to_mlx()` | `agent_training.py` | ✅ Implemented |
| 3. LoRA Training | `train_model_lora()` | `agent_training.py` | ✅ Implemented |
| 4. Evaluation | `evaluate_model()` | `agent_training.py` | ✅ Implemented |
| 5. MLflow Registration | `register_model_mlflow()` | `agent_training.py` | ✅ Implemented |
| 6. MinIO Storage | `store_model_minio()` | `agent_training.py` | ✅ Implemented |
| 7. Ollama Export | `export_model_ollama()` | `agent_training.py` | ✅ Implemented |
| 8. Auto-Import | `trigger_ollama_import()` | `agent_training.py` | ✅ Implemented |

---

## Step 1: Dataset Preparation

### Source: RUNBOOK.md

**Location:** `flux/ai/agent-sre/docs/RUNBOOK.md`

**Statistics:**
- Total lines: 2923
- Alert sections: 100+
- Coverage: All homelab services
- Format: Markdown with code blocks

### Conversion Process

**Function:** `prepare_dataset()` in `agent_training.py`

**Process:**
1. Downloads RUNBOOK.md (if URL) or uses local file
2. Parses alert sections using regex patterns
3. Extracts:
   - Alert name (`### Alert: <Name>`)
   - Symptoms (`**Symptoms:**`)
   - Investigation steps (`**Investigation:**` code blocks)
   - Resolution commands (`**Resolution:**` code blocks)
4. Converts to instruction-following format
5. Splits into train/val/test (80/10/10)

### Dataset Format

**Input Format (Alert):**
```
Alert: FluxReconciliationFailure

Symptoms:
- Flux resource failing to reconcile
- Status shows Ready=False

Investigation Steps:
1. Check reconciliation status: flux get kustomizations -A
2. Check specific resource: kubectl get kustomization <name> -n <namespace>

Resolution Actions:
1. flux reconcile kustomization <name> -n <namespace>
2. Check logs: kubectl logs -n flux-system -l app=kustomize-controller
```

**Output Format (Training Example):**
```json
{
  "instruction": "Alert: FluxReconciliationFailure\n\nSymptoms:\n- Flux resource failing to reconcile\n- Status shows Ready=False\n\nBased on the runbook, provide the exact remediation command.",
  "response": "flux reconcile kustomization <name> -n <namespace>"
}
```

### Dataset Statistics

**Expected Output:**
- Training examples: ~80
- Validation examples: ~10
- Test examples: ~10
- Total: ~100 examples

**Split Configuration:**
- Train: 80%
- Validation: 10%
- Test: 10%

---

## Step 2: Model Conversion

### Base Model

**Model:** `google/functiongemma-270m-it`
- Parameters: 270M
- Format: HuggingFace
- License: Apache 2.0
- Purpose: Function calling and instruction following

### MLX Conversion

**Function:** `convert_model_to_mlx()` in `agent_training.py`

**Process:**
1. Downloads model from HuggingFace (~500MB)
2. Converts to MLX format using `mlx_lm.convert`
3. Optimizes for Apple Silicon
4. Stores in `/tmp/{agent_name}/models/mlx-functiongemma-270m`

**Command Equivalent:**
```bash
python -m mlx_lm.convert \
  --hf-path google/functiongemma-270m-it \
  --mlx-path /tmp/agent-sre/models/mlx-functiongemma-270m
```

---

## Step 3: LoRA Fine-Tuning

### Training Configuration

**Function:** `train_model_lora()` in `agent_training.py`

**Parameters:**
```python
learning_rate: 1e-4
batch_size: 4
iters: 1000
lora_layers: 16
lora_rank: 8
lora_alpha: 16
```

**Training Process:**
1. Loads base MLX model
2. Initializes LoRA adapters
3. Trains on training dataset
4. Validates on validation dataset
5. Saves adapters to `/tmp/{agent_name}/models/functiongemma-sre-finetuned/adapters`

### Success Criteria

- **Accuracy**: >80% correct command generation
- **Relevance**: Commands match runbook procedures
- **Completeness**: Commands include necessary parameters
- **Safety**: No dangerous commands

---

## Step 4-8: Model Registration & Export

### MLflow Registration

**Function:** `register_model_mlflow()`

**Registered Information:**
- Model path (base + adapters)
- Training metrics (accuracy, loss)
- Training parameters
- Agent name and runbook version

### MinIO Storage

**Function:** `store_model_minio()`

**Storage Location:**
- Bucket: `ml-models`
- Path: `agent-sre/{model_version}/`
- Format: Tar.gz archive

**Model Version Format:**
- Pattern: `v{YYYYMMDD}-{execution_id}`
- Example: `v20251224-abc12345`

### Ollama Export

**Function:** `export_model_ollama()`

**Process:**
1. Merges LoRA adapters with base model
2. Creates Modelfile with Agent-SRE configuration
3. Packages model directory
4. Generates import instructions

**Model Name:**
- Format: `agent-sre:v{version}`
- Example: `agent-sre:v20251224-abc12345`

### Automatic Ollama Import

**Function:** `trigger_ollama_import()`

**Process:**
1. Creates Kubernetes Job manifest
2. Job downloads model from MinIO
3. Extracts model files
4. Creates model in Ollama via API
5. Verifies model availability

**Job Name:** `import-agent-sre-ollama-{execution_id}`

---

## Flyte Workflow Integration

### Workflow Definition

**Location:** `flux/infrastructure/flyte/workflows/test/workflows/agent_training.py`

**Workflow:** `agent_training_pipeline()`

**Parameters:**
```python
agent_name: str = "agent-sre"
model_name: str = "google/functiongemma-270m-it"
runbook_path: str = "https://raw.githubusercontent.com/brunovlucena/homelab/main/flux/ai/agent-sre/docs/RUNBOOK.md"
learning_rate: float = 1e-4
batch_size: int = 4
iters: int = 1000
lora_layers: int = 16
lora_rank: int = 8
lora_alpha: int = 16
```

### Execution

**Triggering the Workflow:**
```bash
# Via Flyte CLI
flytectl create execution \
  --project homelab \
  --domain production \
  --name agent-sre-training \
  --workflow agent_training_pipeline \
  --inputs '{"agent_name": "agent-sre", "model_name": "google/functiongemma-270m-it"}'
```

### Workflow Output

**Returns:**
```python
{
    "model_version": "v20251224-abc12345",
    "mlflow_uri": "runs:/...",
    "minio_path": "s3://ml-models/agent-sre/v20251224-abc12345/adapters.tar.gz",
    "adapter_path": "/tmp/agent-sre/models/functiongemma-sre-finetuned/adapters",
    "accuracy": "0.85",
    "ollama_model_name": "agent-sre:v20251224-abc12345",
    "ollama_output_path": "/tmp/agent-sre/models/ollama-export",
    "ollama_modelfile_path": "/tmp/agent-sre/models/ollama-export/Modelfile",
    "ollama_minio_path": "s3://ml-models/agent-sre/v20251224-abc12345-ollama/ollama-model.tar.gz",
    "ollama_import_job": "import-agent-sre-ollama-abc12345",
}
```

---

## Verification Checklist

### Pre-Training

- [ ] RUNBOOK.md is up-to-date
- [ ] Flyte workflow is deployed
- [ ] MinIO is accessible
- [ ] MLflow is configured
- [ ] Ollama service is running

### Post-Training

- [ ] Training completed successfully
- [ ] Model accuracy > 80%
- [ ] Model registered in MLflow
- [ ] Model stored in MinIO
- [ ] Ollama export successful
- [ ] Import job completed
- [ ] Model available in Ollama
- [ ] Model can generate responses

### Deployment

- [ ] Model version updated in LambdaAgent
- [ ] Agent-SRE deployed
- [ ] Model accessible from agent
- [ ] Test alert triggers correct response

---

## Current Status on Studio Cluster

### Verification Steps

**1. Check Flyte workflow executions:**
```bash
flytectl get executions --project homelab --domain production
```

**2. Check for training artifacts in MinIO:**
```bash
kubectl exec -n minio deployment/minio -- mc ls minio/ml-models/agent-sre/
```

**3. Check Kubernetes import jobs:**
```bash
kubectl get jobs -n ai -l component=ollama-import
```

**4. Check Ollama for model:**
```bash
# If Ollama is on host
curl http://localhost:11434/api/tags | grep agent-sre

# Or via port-forward
kubectl port-forward -n ollama svc/ollama-native 11434:11434
curl http://localhost:11434/api/tags
```

### Current Findings

- ⚠️ **Flyte**: Status unknown (need to check)
- ⚠️ **MinIO**: Status unknown (need to verify)
- ⚠️ **Import Jobs**: None found in `ai` namespace
- ⚠️ **Ollama Models**: Cannot verify (Ollama using ExternalName to host)

---

## Troubleshooting

### Common Issues

**1. Training Fails - Out of Memory**
- Solution: Reduce batch_size, lora_rank, or lora_layers

**2. Model Not Learning**
- Solution: Check dataset format, increase learning_rate, verify training data quality

**3. Ollama Import Fails**
- Solution: Check MinIO connectivity, verify Modelfile format, check Ollama API

**4. Model Not Found in Ollama**
- Solution: Re-run import job, verify model name, check Ollama logs

**5. Poor Command Generation**
- Solution: Increase training iterations, check prompt format, verify model loading

---

## Next Steps

### Immediate Actions

1. **Execute Fine-Tuning Pipeline:**
   - Trigger Flyte workflow
   - Monitor training progress
   - Verify all steps complete

2. **Verify Model in Ollama:**
   - Check model list (via host or port-forward)
   - Test model responses
   - Verify model version

3. **Update Agent-SRE Configuration:**
   - Replace placeholder model name
   - Update LambdaAgent with actual version
   - Deploy agent-sre

### Future Improvements

1. **Automation:**
   - Auto-trigger on RUNBOOK.md changes
   - Auto-update model version
   - CI/CD integration

2. **Monitoring:**
   - Track model performance
   - Monitor training metrics
   - Alert on failures

3. **Iteration:**
   - Collect real-world examples
   - Retrain with feedback
   - A/B testing

---

## Conclusion

The fine-tuning process is **fully configured and ready for execution**. All components are in place:

✅ **Dataset preparation**: Automated from RUNBOOK.md  
✅ **Training pipeline**: Complete Flyte workflow  
✅ **Model export**: Ollama format export  
✅ **Auto-import**: Kubernetes job automation  
✅ **Documentation**: Comprehensive guides  

**Status:** Ready for execution. Next step: Trigger Flyte workflow and verify model availability in Ollama.

**Note:** On Studio cluster, Ollama is configured as ExternalName service pointing to `host.docker.internal`. Model verification may require checking the host Ollama instance directly.

---

**Report Generated:** 2025-12-24  
**Cluster:** studio  
**Next Review:** After fine-tuning execution

