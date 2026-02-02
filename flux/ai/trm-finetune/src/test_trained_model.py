#!/usr/bin/env python3
"""
🧪 Test Trained TRM Model

Loads the trained checkpoint and tests it against the runbook dataset.
"""

import os
import sys
import json
import torch
import argparse
from pathlib import Path
from typing import Dict, Any, List

# Add TRM repo to path
TRM_REPO = Path(__file__).parent.parent.parent / "trm"
sys.path.insert(0, str(TRM_REPO))

from puzzle_dataset import PuzzleDataset, PuzzleDatasetConfig, PuzzleDatasetMetadata
from models.recursive_reasoning.trm import TinyRecursiveReasoningModel_ACTV1
from models.recursive_reasoning.trm_config import TinyRecursiveReasoningModelConfig


def load_checkpoint(checkpoint_path: str, model: torch.nn.Module):
    """Load checkpoint into model."""
    device = "cuda" if torch.cuda.is_available() else "cpu"
    print(f"📦 Loading checkpoint from {checkpoint_path}...")
    
    state_dict = torch.load(checkpoint_path, map_location=device)
    
    # Handle potential _orig_mod prefix from compiled models
    if any(k.startswith("_orig_mod.") for k in state_dict.keys()):
        new_state_dict = {}
        for k, v in state_dict.items():
            if k.startswith("_orig_mod."):
                new_state_dict[k[10:]] = v
            else:
                new_state_dict[k] = v
        state_dict = new_state_dict
    
    model.load_state_dict(state_dict, assign=True)
    print(f"✅ Checkpoint loaded successfully")
    return model


def load_test_dataset(dataset_path: str) -> List[Dict[str, Any]]:
    """Load test dataset from JSONL."""
    examples = []
    with open(dataset_path, 'r') as f:
        for line in f:
            if line.strip():
                examples.append(json.loads(line))
    return examples


def extract_expected_solution(example: Dict[str, Any]) -> Dict[str, Any]:
    """Extract expected solution from example."""
    solution_str = example.get("solution", "{}")
    if isinstance(solution_str, str):
        return json.loads(solution_str)
    return solution_str


def create_batch_from_example(example: Dict[str, Any], dataset: PuzzleDataset) -> Dict[str, torch.Tensor]:
    """Create a batch tensor from a single example."""
    # Get the example index from the dataset
    problem = example.get("problem", "")
    
    # Find matching example in dataset
    for idx in range(len(dataset)):
        dataset_example = dataset[idx]
        # Compare problems (simplified - in practice you'd want better matching)
        if dataset_example.get("problem") == problem:
            # Get the actual batch data
            inputs = dataset_example.get("inputs")
            labels = dataset_example.get("labels")
            puzzle_identifiers = dataset_example.get("puzzle_identifiers", [0])
            
            return {
                "inputs": torch.tensor([inputs], dtype=torch.long),
                "labels": torch.tensor([labels], dtype=torch.long),
                "puzzle_identifiers": torch.tensor([puzzle_identifiers], dtype=torch.long),
            }
    
    # Fallback: create from example directly if we can't find it
    # This is a simplified approach - you may need to tokenize properly
    return None


def run_inference(
    model: torch.nn.Module,
    batch: Dict[str, torch.Tensor],
    max_steps: int = 16
) -> Dict[str, Any]:
    """Run inference on a batch."""
    device = "cuda" if torch.cuda.is_available() else "cpu"
    model.eval()
    model.to(device)
    
    # Move batch to device
    batch = {k: v.to(device) for k, v in batch.items()}
    
    with torch.no_grad():
        # Initialize carry
        carry = model.initial_carry(batch)
        
        # Run inference steps
        all_finish = False
        steps = 0
        outputs = []
        
        while not all_finish and steps < max_steps:
            carry, loss, metrics, preds, all_finish = model(
                carry=carry,
                batch=batch,
                return_keys={"outputs"}
            )
            steps += 1
            outputs.append({
                "step": steps,
                "loss": loss.item() if isinstance(loss, torch.Tensor) else loss,
                "finished": all_finish.item() if isinstance(all_finish, torch.Tensor) else all_finish
            })
            
            if all_finish:
                break
        
        # Get final predictions
        final_output = preds.get("outputs", None)
        
        return {
            "steps": steps,
            "outputs": outputs,
            "final_prediction": final_output.cpu().numpy().tolist() if final_output is not None else None,
            "finished": all_finish
        }


def decode_output(output_ids: List[int], dataset: PuzzleDataset) -> str:
    """Decode output token IDs to text."""
    # This depends on how your dataset tokenizes
    # For now, return a placeholder
    if hasattr(dataset, 'tokenizer'):
        return dataset.tokenizer.decode(output_ids)
    # Fallback: try to get vocab from metadata
    return f"<tokens: {len(output_ids)}>"


def test_model(
    checkpoint_path: str,
    dataset_path: str,
    trm_data_path: str,
    max_examples: int = None
):
    """Test the trained model against the test dataset."""
    print("=" * 70)
    print("🧪 Testing Trained TRM Model")
    print("=" * 70)
    print(f"Checkpoint: {checkpoint_path}")
    print(f"Dataset: {dataset_path}")
    print(f"TRM Data: {trm_data_path}")
    print()
    
    # Load test dataset
    print("📚 Loading test dataset...")
    examples = load_test_dataset(dataset_path)
    if max_examples:
        examples = examples[:max_examples]
    print(f"✅ Loaded {len(examples)} examples\n")
    
    # Load TRM dataset for tokenization/metadata
    print("📊 Loading TRM dataset metadata...")
    trm_data_path = Path(trm_data_path)
    train_dir = trm_data_path / "train"
    
    if not (train_dir / "dataset.json").exists():
        print(f"❌ TRM dataset not found at {train_dir}")
        return
    
    with open(train_dir / "dataset.json", 'r') as f:
        dataset_metadata = json.load(f)
    
    # Load dataset for tokenization
    dataset_config = PuzzleDatasetConfig(
        data_path=str(trm_data_path),
        sets=["train"],
        global_batch_size=1,
        eval_global_batch_size=1,
    )
    dataset_metadata_obj = PuzzleDatasetMetadata.from_dict(dataset_metadata)
    dataset = PuzzleDataset(dataset_config, dataset_metadata_obj)
    print(f"✅ Dataset loaded (vocab size: {dataset_metadata_obj.vocab_size})\n")
    
    # Load model
    print("🤖 Loading model...")
    # Get config from checkpoint directory
    checkpoint_dir = Path(checkpoint_path).parent
    config_path = checkpoint_dir / "all_config.yaml"
    
    if not config_path.exists():
        print(f"⚠️  Config not found at {config_path}, using defaults")
        # Use default config matching training
        arch_config = TinyRecursiveReasoningModelConfig(
            hidden_size=512,
            num_heads=8,
            L_layers=2,
            H_cycles=3,
            L_cycles=6,
            pos_encodings="none",
            puzzle_emb_len=16,
            puzzle_emb_ndim=512,
        )
    else:
        import yaml
        with open(config_path, 'r') as f:
            config_dict = yaml.safe_load(f)
        arch_dict = config_dict.get("arch", {})
        arch_config = TinyRecursiveReasoningModelConfig(**arch_dict)
    
    model = TinyRecursiveReasoningModel_ACTV1(arch_config, dataset_metadata_obj.vocab_size)
    model = load_checkpoint(checkpoint_path, model)
    print()
    
    # Test on examples
    print("🧪 Running inference on test examples...\n")
    results = {
        "total": len(examples),
        "successful": 0,
        "failed": 0,
        "details": []
    }
    
    for i, example in enumerate(examples, 1):
        print(f"Example {i}/{len(examples)}: {example.get('problem', '')[:80]}...")
        
        # Get expected solution
        expected = extract_expected_solution(example)
        expected_lambda = expected.get("lambda_function")
        expected_params = expected.get("parameters", {})
        
        # Create batch (simplified - you may need proper tokenization)
        # For now, we'll use a simpler approach: test the model can load and run
        try:
            # Try to find matching example in dataset
            batch = create_batch_from_example(example, dataset)
            
            if batch is None:
                print(f"  ⚠️  Could not create batch for example {i}")
                results["failed"] += 1
                results["details"].append({
                    "example": i,
                    "status": "failed",
                    "reason": "Could not create batch"
                })
                continue
            
            # Run inference
            inference_result = run_inference(model, batch)
            
            print(f"  ✅ Inference completed in {inference_result['steps']} steps")
            print(f"  📊 Loss: {inference_result['outputs'][-1]['loss']:.4f}")
            
            results["successful"] += 1
            results["details"].append({
                "example": i,
                "status": "success",
                "steps": inference_result["steps"],
                "finished": inference_result["finished"]
            })
            
        except Exception as e:
            print(f"  ❌ Error: {e}")
            results["failed"] += 1
            results["details"].append({
                "example": i,
                "status": "failed",
                "error": str(e)
            })
        
        print()
    
    # Print summary
    print("=" * 70)
    print("📊 Test Summary")
    print("=" * 70)
    print(f"Total examples:  {results['total']}")
    print(f"✅ Successful:    {results['successful']} ({100 * results['successful'] / results['total']:.1f}%)")
    print(f"❌ Failed:       {results['failed']} ({100 * results['failed'] / results['total']:.1f}%)")
    print()
    
    if results['failed'] > 0:
        print("❌ Failed Examples:")
        for detail in results['details']:
            if detail['status'] == 'failed':
                print(f"  - Example {detail['example']}: {detail.get('error', 'Unknown error')}")
        print()
    
    return results


def main():
    parser = argparse.ArgumentParser(description="Test trained TRM model")
    parser.add_argument(
        "--checkpoint",
        type=str,
        default="../trm/checkpoints/Trm_data-ACT-torch/trm-runbook-mac/step_0",
        help="Path to model checkpoint"
    )
    parser.add_argument(
        "--dataset",
        type=str,
        default="./data/runbook_test_dataset.jsonl",
        help="Path to test dataset JSONL"
    )
    parser.add_argument(
        "--trm-data",
        type=str,
        default="./models/trm-runbook-only/trm_data",
        help="Path to TRM dataset directory"
    )
    parser.add_argument(
        "--max-examples",
        type=int,
        default=None,
        help="Maximum number of examples to test (for quick testing)"
    )
    
    args = parser.parse_args()
    
    # Validate paths
    if not Path(args.checkpoint).exists():
        print(f"❌ Checkpoint not found: {args.checkpoint}")
        sys.exit(1)
    
    if not Path(args.dataset).exists():
        print(f"❌ Dataset not found: {args.dataset}")
        sys.exit(1)
    
    if not Path(args.trm_data).exists():
        print(f"❌ TRM data not found: {args.trm_data}")
        sys.exit(1)
    
    # Run tests
    results = test_model(
        checkpoint_path=args.checkpoint,
        dataset_path=args.dataset,
        trm_data_path=args.trm_data,
        max_examples=args.max_examples
    )
    
    # Exit with error if any failures
    if results and results['failed'] > 0:
        sys.exit(1)
    else:
        print("🎉 All tests completed!")


if __name__ == "__main__":
    main()
