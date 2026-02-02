# TRM Fine-Tuning Guide - Runbook Dataset

## Overview

This guide walks through fine-tuning TRM (Tiny Recursive Models) on the runbook dataset for Agent-SRE remediation selection.

## Prerequisites

1. **TRM Repository**: Clone the TinyRecursiveModels repository
2. **Training Data**: Runbook test dataset (7 examples)
3. **GPU Access**: Recommended for training (can use CPU but slower)

## Step 1: Prepare Training Data

```bash
cd flux/ai/trm-finetune

# Generate test dataset (if not already done)
python src/test_trm_runbook.py

# Validate and prepare for TRM training
python src/prepare_training.py
```

**Output**:
- Validates dataset format
- Analyzes statistics
- Prepares TRM format at `models/trm-runbook-only/trm_data/train.jsonl`

## Step 2: Setup TRM Repository

```bash
# Clone TRM repository (if not already done)
cd ../..
git clone https://github.com/SamsungSAILMontreal/TinyRecursiveModels.git trm
cd trm

# Install dependencies
pip install --upgrade pip wheel setuptools
pip install --pre --upgrade torch torchvision torchaudio --index-url https://download.pytorch.org/whl/nightly/cu126
pip install -r requirements.txt
```

## Step 3: Run Training

```bash
cd flux/ai/trm-finetune

# Set TRM repository path
export TRM_REPO_PATH=../../trm

# Run training
python src/trm_trainer.py \
  --training-data ./data/runbook_test_dataset.jsonl \
  --output-dir models/trm-runbook-only \
  --epochs 10000 \
  --eval-interval 1000 \
  --run-name trm-runbook-finetune
```

### Training Parameters

- **epochs**: Number of training epochs (default: 10000 for small dataset)
- **eval-interval**: Evaluate every N epochs (default: 1000)
- **output-dir**: Where to save the trained model
- **run-name**: Name for this training run

### Expected Training Time

- **CPU**: ~2-4 hours for 10000 epochs (7 examples)
- **GPU**: ~30-60 minutes for 10000 epochs

## Step 4: Use Trained Model

After training, the model will be exported to:
```
models/trm-runbook-only/export/
```

### Test with Selector

```bash
# Set model path
export TRM_MODEL_PATH=./models/trm-runbook-only/export

# Test selector
python src/trm_remediation_selector.py '{
  "labels": {
    "alertname": "FluxReconciliationFailure",
    "name": "homepage",
    "namespace": "flux-system"
  }
}'
```

## Dataset Statistics

Current dataset:
- **Total Examples**: 7
- **Alert Types**: 7 unique alerts
- **Lambda Functions**: 6 unique functions
- **Average Problem Length**: ~862 characters
- **Average Solution Length**: ~215 characters
- **Reasoning Steps**: 6 steps per example

## Limitations & Recommendations

### Current Limitations

1. **Small Dataset**: Only 7 examples
   - May lead to overfitting
   - Limited generalization

2. **No Validation Set**: All examples used for training
   - Hard to measure true performance

### Recommendations

1. **Data Augmentation**:
   - Generate variations of existing alerts
   - Add more alert types from runbook
   - Include edge cases

2. **Collect More Data**:
   - Extract from actual Prometheus alerts
   - Use historical incident data
   - Generate synthetic alerts

3. **Use Pre-trained Model**:
   - Start with pre-trained TRM (if available)
   - Fine-tune on runbook data
   - Better generalization

4. **Cross-Validation**:
   - Split into train/validation sets
   - Monitor overfitting
   - Early stopping

## Troubleshooting

### TRM Repository Not Found

```bash
# Check if repository exists
ls -la ../../trm

# If not, clone it
git clone https://github.com/SamsungSAILMontreal/TinyRecursiveModels.git ../../trm
```

### Training Fails

1. **Check GPU availability**:
   ```bash
   nvidia-smi  # For NVIDIA GPUs
   ```

2. **Reduce batch size** if OOM:
   - Edit `trm_trainer.py` config
   - Set `global_batch_size=64` or lower

3. **Check data format**:
   ```bash
   python src/prepare_training.py
   ```

### Model Not Loading

1. **Verify export path**:
   ```bash
   ls -la models/trm-runbook-only/export/
   ```

2. **Check model checkpoint**:
   - Should contain `.pth` or `.ckpt` files
   - Verify file permissions

## Next Steps

After training:

1. ✅ Validate model on test dataset
2. ✅ Integrate with agent-sre
3. ✅ Deploy TRM inference service
4. ✅ Test end-to-end flow
5. ✅ Monitor performance in production

## References

- [TRM Repository](https://github.com/SamsungSAILMontreal/TinyRecursiveModels)
- [TRM Testing Guide](docs/TRM_TESTING_GUIDE.md)
- [TRM Integration](docs/TRM_AGENT_SRE_INTEGRATION.md)
