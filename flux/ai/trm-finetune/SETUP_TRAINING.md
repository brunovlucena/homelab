# TRM Training Setup Guide

## ✅ Step 1: Clone TRM Repository - COMPLETE

The TRM repository has been cloned to:
```
/Users/brunolucena/workspace/bruno/repos/homelab/flux/ai/trm
```

## ⚠️ Step 2: Install Dependencies

Training requires PyTorch and other dependencies. Since your Python environment is managed by `uv`, you have a few options:

### Option A: Use uv to manage dependencies (Recommended)

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/ai/trm-finetune

# Create a uv project (if not already)
uv init --no-readme

# Add TRM dependencies
uv add torch torchvision torchaudio
uv add einops tqdm coolname pydantic argdantic wandb omegaconf hydra-core huggingface_hub packaging ninja wheel setuptools setuptools-scm pydantic-core numba triton

# Install adam-atan2 separately (requires special handling)
uv pip install --no-cache-dir --no-build-isolation adam-atan2
```

### Option B: Use conda/mamba environment

```bash
# Create conda environment
conda create -n trm-training python=3.10
conda activate trm-training

# Install PyTorch (adjust for your CUDA version)
conda install pytorch torchvision torchaudio pytorch-cuda=12.6 -c pytorch -c nvidia

# Install other dependencies
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/ai/trm
pip install -r requirements.txt
pip install --no-cache-dir --no-build-isolation adam-atan2
```

### Option C: Use Docker (Recommended for GPU training)

Create a Dockerfile or use the existing one in the TRM repository.

## 🚀 Step 3: Run Training

Once dependencies are installed:

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/ai/trm-finetune

# Set TRM repository path
export TRM_REPO_PATH=../trm

# Run training
python src/trm_trainer.py \
  --training-data ./data/runbook_test_dataset.jsonl \
  --output-dir models/trm-runbook-only \
  --epochs 10000 \
  --eval-interval 1000 \
  --run-name trm-runbook-finetune
```

## 📝 Current Status

- ✅ TRM repository cloned
- ✅ Dataset prepared and validated
- ✅ Training script ready
- ⏳ Dependencies need to be installed
- ⏳ Training ready to run (after dependencies)

## ⚠️ Notes

1. **GPU Recommended**: Training will be much faster on GPU. CPU training is possible but slow.

2. **Small Dataset**: With only 7 examples, training may:
   - Overfit quickly
   - Require careful monitoring
   - Benefit from data augmentation

3. **Training Time**:
   - GPU (L40S): ~30-60 minutes for 10000 epochs
   - CPU: ~2-4 hours for 10000 epochs

4. **Environment**: Your system uses `uv` for Python management. Consider:
   - Using `uv` to manage project dependencies
   - Creating a separate conda environment
   - Using Docker for isolated training environment

## 🔍 Verify Setup

After installing dependencies, verify:

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/ai/trm
python -c "import torch; print(f'PyTorch: {torch.__version__}')"
python -c "import torch; print(f'CUDA available: {torch.cuda.is_available()}')"
```

## 📚 References

- [TRM Repository](https://github.com/SamsungSAILMontreal/TinyRecursiveModels)
- [TRM README](../trm/README.md)
- [Training Guide](TRAINING_GUIDE.md)
