# Mac Studio Requirements for TRM Fine-Tuning

## ✅ Your Setup: Mac Studio with 256GB RAM

**Your Mac Studio is EXCELLENT for TRM fine-tuning!** Here's why:

---

## 🎯 TRM Model Specifications

### Model Size
- **7M parameters** (Tiny Recursive Model)
- **Memory requirement**: ~50-100MB for model weights
- **Training memory**: ~2-4GB with batch size 128
- **Your RAM**: 256GB ✅ (massive headroom)

### Comparison
| Model | Parameters | Training RAM | Your Setup |
|-------|-----------|--------------|------------|
| TRM 7M | 7M | ~4GB | ✅ 256GB (64x headroom) |
| FunctionGemma 270M | 270M | ~8GB | ✅ 256GB (32x headroom) |
| Llama 3.2 3B | 3B | ~12GB | ✅ 256GB (21x headroom) |

---

## 💻 Mac Studio Capabilities

### Your Hardware
- **CPU**: M1/M2/M3 Ultra (excellent for MLX)
- **RAM**: 256GB unified memory (perfect for training)
- **GPU**: Integrated GPU (MLX optimized)
- **Neural Engine**: 32-core (accelerates inference)

### Why Mac Studio is Perfect

1. **MLX Framework**: Optimized for Apple Silicon
   - Uses Metal Performance Shaders
   - Unified memory architecture
   - No GPU memory transfers needed

2. **256GB RAM**: Massive advantage
   - Can train with large batch sizes
   - No memory constraints
   - Can run multiple experiments simultaneously

3. **Energy Efficient**: 
   - Lower power consumption than GPU servers
   - Quiet operation
   - No cooling issues

---

## 📊 Training Requirements

### TRM Fine-Tuning

```python
# Estimated resources for TRM 7M
Model Size: 7M parameters
Training Memory: ~4GB (with batch_size=128)
Training Time: ~2-4 hours (on Mac Studio)
Storage: ~500MB for model + data
```

### Training Configuration

```yaml
# Recommended settings for Mac Studio
batch_size: 128  # Can go higher with 256GB RAM
learning_rate: 1e-4
epochs: 50000
gradient_accumulation: 1  # Not needed with your RAM
mixed_precision: true  # MLX handles this automatically
```

### What You Can Do

✅ **Train TRM locally** - No cloud needed  
✅ **Large batch sizes** - Up to 512+ with your RAM  
✅ **Multiple experiments** - Run several in parallel  
✅ **Fast iteration** - No cloud upload/download delays  
✅ **Privacy** - All data stays local  

---

## 🚀 Setup Instructions

### 1. Install MLX

```bash
# MLX is optimized for Apple Silicon
pip install mlx mlx-lm

# Verify installation
python -c "import mlx; import mlx_lm; print('✅ MLX ready')"
```

### 2. Clone TRM Repository

```bash
cd ~/workspace
git clone https://github.com/SamsungSAILMontreal/TinyRecursiveModels.git
cd TinyRecursiveModels
```

### 3. Install Dependencies

```bash
# Install TRM requirements
pip install -r requirements.txt

# Install your fine-tuning dependencies
cd ../bruno/repos/homelab/flux/ai/trm-finetune
pip install -r requirements.txt
```

### 4. Prepare Training Data

```bash
# Generate runbook dataset
python src/test_trm_runbook.py

# Collect observability data (if you have Prometheus/Loki/Tempo)
python src/data_collector.py
```

### 5. Start Training

```bash
# Train TRM on your data
python src/trm_trainer.py \
  --training-data data/merged_training_data.jsonl \
  --output-dir models/trm-homelab-finetuned \
  --batch-size 128 \
  --epochs 50000 \
  --learning-rate 1e-4
```

---

## ⏱️ Expected Training Times

### TRM 7M Fine-Tuning

| Dataset Size | Batch Size | Training Time (Mac Studio) |
|--------------|------------|----------------------------|
| 1,000 examples | 128 | ~1-2 hours |
| 5,000 examples | 128 | ~3-4 hours |
| 10,000 examples | 128 | ~6-8 hours |
| 10,000 examples | 256 | ~4-6 hours (faster with larger batch) |

**Note**: With 256GB RAM, you can use batch_size=256 or even 512 for faster training!

---

## 🔥 Performance Tips

### 1. Use Larger Batch Sizes

```python
# With 256GB RAM, you can use:
batch_size = 256  # or even 512
# This will speed up training significantly
```

### 2. Parallel Data Loading

```python
# Use multiple workers for data loading
num_workers = 8  # Mac Studio has many CPU cores
```

### 3. Mixed Precision (Automatic in MLX)

MLX automatically uses mixed precision, so you don't need to configure it.

### 4. Monitor Training

```bash
# Watch memory usage
htop  # or Activity Monitor

# Watch GPU usage (if available)
sudo powermetrics --samplers gpu_power -i 1000
```

---

## 🆚 Mac Studio vs Cloud Training

### Mac Studio Advantages

✅ **No cloud costs** - Train for free  
✅ **Privacy** - Data never leaves your machine  
✅ **Fast iteration** - No upload/download delays  
✅ **256GB RAM** - More than most cloud instances  
✅ **MLX optimized** - Native Apple Silicon support  

### Cloud Advantages

✅ **Multi-GPU** - If you need distributed training  
✅ **Scalability** - Can scale to larger models  
✅ **24/7 availability** - Don't need your Mac running  

### Recommendation

**Use your Mac Studio!** TRM is only 7M parameters, so:
- Training is fast (hours, not days)
- Your RAM is more than enough
- MLX is optimized for Apple Silicon
- No cloud costs needed

---

## 📦 Storage Requirements

### Model Storage

```
TRM Base Model: ~50MB
Fine-tuned Model: ~100MB
Training Data: ~100-500MB
Checkpoints: ~500MB-1GB
Total: ~2GB
```

**Your Mac Studio**: Likely has 1TB+ SSD ✅ (plenty of space)

---

## 🧪 Testing Your Setup

### Quick Test

```bash
# Test MLX installation
python -c "
import mlx.core as mx
import mlx.nn as nn
x = mx.random.normal((1000, 1000))
y = nn.Linear(1000, 1000)(x)
print('✅ MLX working on Apple Silicon')
"

# Test TRM model loading
python -c "
from models.recursive_reasoning.trm import TRMModel
print('✅ TRM can be imported')
"
```

### Memory Test

```bash
# Monitor memory during training
# Open Activity Monitor and watch:
# - Memory Pressure (should stay green)
# - GPU Memory (if available)
# - CPU Usage
```

---

## 🎯 Recommended Workflow

### 1. Local Development (Mac Studio)

```bash
# Develop and test on Mac Studio
python src/test_trm_runbook.py
python src/trm_trainer.py --epochs 1000  # Quick test
```

### 2. Full Training (Mac Studio)

```bash
# Full training run
python src/trm_trainer.py \
  --training-data data/merged_training_data.jsonl \
  --output-dir models/trm-homelab-finetuned \
  --epochs 50000 \
  --batch-size 256  # Use large batch with your RAM
```

### 3. Deploy to Homelab

```bash
# Export model
python src/trm_trainer.py --export models/trm-homelab-finetuned/export

# Deploy to Kubernetes
kubectl apply -f k8s/trm-reasoning-service.yaml
```

---

## 💡 Pro Tips

1. **Use Terminal with iTerm2** - Better for long-running training
2. **Enable Do Not Disturb** - Prevent interruptions during training
3. **Monitor Temperature** - Mac Studio should stay cool, but monitor if training for hours
4. **Save Checkpoints** - Save every 1000 epochs
5. **Use Screen/Tmux** - Keep training running if you disconnect

---

## 🚨 Troubleshooting

### Out of Memory (Unlikely with 256GB)

```python
# Reduce batch size
batch_size = 64  # Instead of 128

# Or use gradient accumulation
gradient_accumulation_steps = 2
```

### Slow Training

```python
# Increase batch size (you have the RAM!)
batch_size = 256  # or 512

# Use more data loading workers
num_workers = 8
```

### MLX Not Using GPU

```bash
# Check if MLX detects GPU
python -c "import mlx.core as mx; print(mx.metal.is_available())"
# Should print: True
```

---

## ✅ Conclusion

**Your Mac Studio with 256GB RAM is PERFECT for TRM fine-tuning!**

- ✅ More than enough RAM
- ✅ MLX optimized for Apple Silicon
- ✅ Fast training (hours, not days)
- ✅ No cloud costs
- ✅ Privacy (data stays local)

**You can train TRM locally without any issues!**

---

## 📚 Additional Resources

- [MLX Documentation](https://ml-explore.github.io/mlx/)
- [MLX-LM Examples](https://github.com/ml-explore/mlx-examples)
- [TRM Repository](https://github.com/SamsungSAILMontreal/TinyRecursiveModels)
- [Agent-SRE Fine-Tuning Guide](../../agent-sre/docs/FINE_TUNING.md)

