# 🧪 Testing the Trained TRM Model

This guide shows you how to test your trained TRM model.

## Quick Test

The simplest way to verify your model works:

```bash
cd flux/ai/trm-finetune
.venv/bin/python src/test_model_simple.py
```

This will:
- ✅ Load the checkpoint
- ✅ Verify the model architecture
- ✅ Run a simple inference test
- ✅ Load and display test examples

## Test Against Full Dataset

### Option 1: Use the Validator (Rule-based fallback)

Test the remediation selector (will use rule-based fallback if model not loaded):

```bash
.venv/bin/python src/validate_selector.py ./data/runbook_test_dataset.jsonl
```

### Option 2: Test with Trained Model

For detailed inference testing with the actual trained model:

```bash
.venv/bin/python src/test_trained_model.py \
  --checkpoint ../trm/checkpoints/Trm_data-ACT-torch/trm-runbook-mac/step_0 \
  --dataset ./data/runbook_test_dataset.jsonl \
  --trm-data ./models/trm-runbook-only/trm_data
```

### Option 3: Test Individual Alerts

Test the selector on a single alert:

```bash
# Create a test alert JSON
echo '{
  "labels": {
    "alertname": "FluxReconciliationFailure",
    "name": "my-app",
    "namespace": "flux-system",
    "kind": "Kustomization"
  },
  "annotations": {
    "summary": "Kustomization reconciliation failed"
  }
}' > /tmp/test_alert.json

# Test with selector
.venv/bin/python src/trm_remediation_selector.py "$(cat /tmp/test_alert.json)"
```

## Model Checkpoint Location

Your trained model checkpoint is at:
- **Checkpoint**: `../trm/checkpoints/Trm_data-ACT-torch/trm-runbook-mac/step_0`
- **Config**: `../trm/checkpoints/Trm_data-ACT-torch/trm-runbook-mac/all_config.yaml`

## Expected Results

### Simple Test
- ✅ Model loads without errors
- ✅ Inference runs successfully
- ✅ Loss decreases over steps
- ✅ Model finishes inference

### Full Dataset Test
- ✅ All 7 test examples processed
- ✅ Model predictions match expected outputs
- ✅ Lambda function selection is correct
- ✅ Parameters are correctly extracted

## Troubleshooting

### Model won't load
- Check checkpoint path exists
- Verify config file is present
- Ensure TRM repo is in correct location

### Inference errors
- Check dataset path is correct
- Verify vocab size matches
- Ensure device (CPU/CUDA) is correct

### Import errors
- Make sure TRM_REPO_PATH is set correctly
- Verify all dependencies are installed
- Check Python path includes TRM repo

## Next Steps

After testing:
1. ✅ Verify model works correctly
2. 🔄 Run longer training if needed
3. 🚀 Integrate with agent-sre
4. 📊 Monitor performance in production
