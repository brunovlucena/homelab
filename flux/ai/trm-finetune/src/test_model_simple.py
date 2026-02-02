#!/usr/bin/env python3
"""
🧪 Simple Test for Trained TRM Model

Quick test to verify the model loads and can run inference.
"""

import os
import sys
import json
import torch
from pathlib import Path

# Add TRM repo to path
TRM_REPO = Path(__file__).parent.parent.parent / "trm"
sys.path.insert(0, str(TRM_REPO))

print("=" * 70)
print("🧪 Testing Trained TRM Model")
print("=" * 70)

# Paths
checkpoint_path = Path("../trm/checkpoints/Trm_data-ACT-torch/trm-runbook-mac/step_0")
config_path = checkpoint_path.parent / "all_config.yaml"
dataset_path = Path("./data/runbook_test_dataset.jsonl")

print(f"\n📦 Checkpoint: {checkpoint_path}")
print(f"📋 Config: {config_path}")
print(f"📚 Dataset: {dataset_path}")

# Check files exist
if not checkpoint_path.exists():
    print(f"\n❌ Checkpoint not found: {checkpoint_path}")
    sys.exit(1)

if not config_path.exists():
    print(f"\n❌ Config not found: {config_path}")
    sys.exit(1)

if not dataset_path.exists():
    print(f"\n❌ Dataset not found: {dataset_path}")
    sys.exit(1)

print("\n✅ All files found!")

# Load config
print("\n📋 Loading config...")
import yaml
with open(config_path, 'r') as f:
    config = yaml.safe_load(f)

arch_config_dict = config.get("arch", {})
print(f"   Architecture: {arch_config_dict.get('name', 'unknown')}")
print(f"   Hidden size: {arch_config_dict.get('hidden_size', 'unknown')}")
print(f"   L_cycles: {arch_config_dict.get('L_cycles', 'unknown')}")
print(f"   H_cycles: {arch_config_dict.get('H_cycles', 'unknown')}")

# Load dataset metadata
trm_data_path = Path("./models/trm-runbook-only/trm_data")
dataset_json_path = trm_data_path / "train" / "dataset.json"

if not dataset_json_path.exists():
    print(f"\n❌ Dataset metadata not found: {dataset_json_path}")
    sys.exit(1)

print("\n📊 Loading dataset metadata...")
with open(dataset_json_path, 'r') as f:
    dataset_metadata = json.load(f)

vocab_size = dataset_metadata.get("vocab_size", 0)
print(f"   Vocab size: {vocab_size}")

# Load model
print("\n🤖 Loading model...")
from utils.functions import load_model_class

# Get model class from config
model_cls = load_model_class(arch_config_dict.get("name", "recursive_reasoning.trm@TinyRecursiveReasoningModel_ACTV1"))

# Create model config
model_cfg = dict(
    **arch_config_dict,
    batch_size=1,
    vocab_size=vocab_size,
    seq_len=dataset_metadata.get("seq_len", 1024),
    num_puzzle_identifiers=dataset_metadata.get("num_puzzle_identifiers", 1),
    causal=False
)

# Instantiate model and wrap with loss head
base_model = model_cls(model_cfg)
loss_config = arch_config_dict.get("loss", {})
loss_head_cls = load_model_class(loss_config.get("name", "losses@ACTLossHead"))
loss_type = loss_config.get("loss_type", "stablemax_cross_entropy")
model = loss_head_cls(base_model, loss_type=loss_type)

device = "cuda" if torch.cuda.is_available() else "cpu"
print(f"   Device: {device}")

# Load checkpoint
print(f"\n📦 Loading checkpoint...")
state_dict = torch.load(checkpoint_path, map_location=device)

# Handle prefixes: _orig_mod. and model.
# The checkpoint has keys like "inner.*" but the wrapped model expects "model.inner.*"
new_state_dict = {}
for k, v in state_dict.items():
    new_key = k
    # Remove _orig_mod. prefix
    if new_key.startswith("_orig_mod."):
        new_key = new_key[10:]
    # Add model. prefix if not present (wrapped model needs it)
    if not new_key.startswith("model.") and not new_key.startswith("loss_fn"):
        new_key = "model." + new_key
    new_state_dict[new_key] = v

state_dict = new_state_dict
model.load_state_dict(state_dict, assign=True)
model.eval()
model.to(device)
print("✅ Model loaded successfully!")

# Test with a simple batch
print("\n🧪 Testing inference...")
from puzzle_dataset import PuzzleDataset, PuzzleDatasetConfig, PuzzleDatasetMetadata

# Load dataset
dataset_config = PuzzleDatasetConfig(
    dataset_paths=[str(trm_data_path)],
    global_batch_size=1,
    seed=0,
    rank=0,
    num_replicas=1,
    test_set_mode=False,  # Boolean, not string
    epochs_per_iter=1,
)
# Create dataset - it will load metadata internally
dataset = PuzzleDataset(dataset_config, split="train")
dataset_metadata_obj = dataset.metadata

# Get first example using iterator
print("   Getting first example from dataset...")
dataset_iter = iter(dataset)
# PuzzleDataset returns (set_name, batch_dict, global_batch_size)
set_name, batch, global_batch_size = next(dataset_iter)
print(f"   Set name: {set_name}, Global batch size: {global_batch_size}")
print(f"   Batch keys: {list(batch.keys())}")

print(f"   Batch shape - inputs: {batch['inputs'].shape}, labels: {batch['labels'].shape}")

# Run inference
print("   Running inference...")
with torch.no_grad():
    carry = model.initial_carry(batch)
    
    steps = 0
    max_steps = 10
    
    while steps < max_steps:
        # Model with loss head expects return_keys parameter
        carry, loss, metrics, preds, all_finish = model(
            carry=carry,
            batch=batch,
            return_keys=["outputs"]
        )
        steps += 1
        
        loss_val = loss.item() if isinstance(loss, torch.Tensor) else loss
        finished = all_finish.item() if isinstance(all_finish, torch.Tensor) else all_finish
        
        print(f"      Step {steps}: loss={loss_val:.4f}, finished={finished}")
        
        if finished:
            break

print(f"\n✅ Inference completed in {steps} steps!")

# Test loading test dataset
print("\n📚 Testing dataset loading...")
examples = []
with open(dataset_path, 'r') as f:
    for line in f:
        if line.strip():
            examples.append(json.loads(line))

print(f"✅ Loaded {len(examples)} test examples")

# Show first example
if examples:
    first_test = examples[0]
    print(f"\n📋 First test example:")
    print(f"   Problem: {first_test.get('problem', '')[:100]}...")
    solution = json.loads(first_test.get('solution', '{}'))
    print(f"   Expected lambda: {solution.get('lambda_function', 'N/A')}")
    print(f"   Expected params: {solution.get('parameters', {})}")

print("\n" + "=" * 70)
print("🎉 Model test completed successfully!")
print("=" * 70)
print("\nNext steps:")
print("1. Use validate_selector.py to test against full dataset")
print("2. Use test_trained_model.py for detailed inference testing")
print("3. Integrate with trm_remediation_selector.py")
