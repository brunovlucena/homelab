#!/usr/bin/env python3
"""
Fine-tune FunctionGemma 270M using MLX-LM on Mac Studio M3 Ultra.

This script uses MLX-LM's LoRA fine-tuning for efficient training on Apple Silicon.
"""
import argparse
import subprocess
import sys
from pathlib import Path


def check_mlx_installation():
    """Check if MLX and MLX-LM are installed."""
    try:
        import mlx
        import mlx_lm
        print(f"✅ MLX version: {mlx.__version__}")
        print(f"✅ MLX-LM available")
        return True
    except ImportError as e:
        print(f"❌ MLX not installed: {e}")
        print("\nInstall with: pip install mlx mlx-lm")
        return False


def convert_model_if_needed(model_name: str, output_path: Path):
    """Convert HuggingFace model to MLX format if needed."""
    if output_path.exists():
        print(f"✅ Model already converted: {output_path}")
        return str(output_path)
    
    print(f"Converting model {model_name} to MLX format...")
    cmd = [
        sys.executable, "-m", "mlx_lm.convert",
        "--hf-path", model_name,
        "--mlx-path", str(output_path)
    ]
    
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"❌ Conversion failed: {result.stderr}")
        sys.exit(1)
    
    print(f"✅ Model converted: {output_path}")
    return str(output_path)


def train_with_lora(
    model_path: str,
    train_data: Path,
    val_data: Path,
    output_dir: Path,
    learning_rate: float = 1e-4,
    batch_size: int = 4,
    iters: int = 1000,
    val_batches: int = 20,
    lora_layers: int = 16,
    lora_rank: int = 8,
    lora_alpha: int = 16,
    lora_dropout: float = 0.05
):
    """Fine-tune model using LoRA."""
    output_dir.mkdir(parents=True, exist_ok=True)
    
    print(f"\n🚀 Starting LoRA fine-tuning...")
    print(f"  Model: {model_path}")
    print(f"  Training data: {train_data}")
    print(f"  Validation data: {val_data}")
    print(f"  Output: {output_dir}")
    print(f"  Learning rate: {learning_rate}")
    print(f"  Batch size: {batch_size}")
    print(f"  Iterations: {iters}")
    print(f"  LoRA layers: {lora_layers}")
    print(f"  LoRA rank: {lora_rank}")
    
    cmd = [
        sys.executable, "-m", "mlx_lm.lora",
        "--model", model_path,
        "--train",
        "--data", str(train_data),
        "--val-data", str(val_data),
        "--iters", str(iters),
        "--val-batches", str(val_batches),
        "--learning-rate", str(learning_rate),
        "--batch-size", str(batch_size),
        "--lora-layers", str(lora_layers),
        "--rank", str(lora_rank),
        "--alpha", str(lora_alpha),
        "--dropout", str(lora_dropout),
        "--adapter-path", str(output_dir / "adapters")
    ]
    
    print(f"\nExecuting: {' '.join(cmd)}\n")
    result = subprocess.run(cmd)
    
    if result.returncode != 0:
        print(f"❌ Training failed with exit code {result.returncode}")
        sys.exit(1)
    
    print(f"\n✅ Training completed!")
    print(f"  Adapters saved to: {output_dir / 'adapters'}")


def main():
    parser = argparse.ArgumentParser(
        description="Fine-tune FunctionGemma 270M with MLX-LM on Mac Studio M3 Ultra"
    )
    parser.add_argument(
        "--model",
        default="google/functiongemma-270m-it",
        help="Model name (HuggingFace) or path"
    )
    parser.add_argument(
        "--train-data",
        type=Path,
        default=Path(__file__).parent / "data" / "train.jsonl",
        help="Training data JSONL file"
    )
    parser.add_argument(
        "--val-data",
        type=Path,
        default=Path(__file__).parent / "data" / "val.jsonl",
        help="Validation data JSONL file"
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path(__file__).parent / "models" / "functiongemma-sre-finetuned",
        help="Output directory for fine-tuned model"
    )
    parser.add_argument(
        "--mlx-model-dir",
        type=Path,
        default=Path(__file__).parent / "models" / "mlx-functiongemma-270m",
        help="Directory for MLX-converted model"
    )
    parser.add_argument(
        "--learning-rate",
        type=float,
        default=1e-4,
        help="Learning rate"
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=4,
        help="Batch size"
    )
    parser.add_argument(
        "--iters",
        type=int,
        default=1000,
        help="Number of training iterations"
    )
    parser.add_argument(
        "--lora-layers",
        type=int,
        default=16,
        help="Number of LoRA layers"
    )
    parser.add_argument(
        "--lora-rank",
        type=int,
        default=8,
        help="LoRA rank"
    )
    parser.add_argument(
        "--lora-alpha",
        type=int,
        default=16,
        help="LoRA alpha"
    )
    parser.add_argument(
        "--skip-convert",
        action="store_true",
        help="Skip model conversion (use existing MLX model)"
    )
    
    args = parser.parse_args()
    
    # Check MLX installation
    if not check_mlx_installation():
        sys.exit(1)
    
    # Convert model if needed
    if not args.skip_convert:
        model_path = convert_model_if_needed(args.model, args.mlx_model_dir)
    else:
        model_path = str(args.mlx_model_dir)
    
    # Check training data exists
    if not args.train_data.exists():
        print(f"❌ Training data not found: {args.train_data}")
        print(f"   Run: python training/prepare_dataset.py first")
        sys.exit(1)
    
    if not args.val_data.exists():
        print(f"❌ Validation data not found: {args.val_data}")
        print(f"   Run: python training/prepare_dataset.py first")
        sys.exit(1)
    
    # Train
    train_with_lora(
        model_path=model_path,
        train_data=args.train_data,
        val_data=args.val_data,
        output_dir=args.output_dir,
        learning_rate=args.learning_rate,
        batch_size=args.batch_size,
        iters=args.iters,
        lora_layers=args.lora_layers,
        lora_rank=args.lora_rank,
        lora_alpha=args.lora_alpha
    )


if __name__ == "__main__":
    main()

