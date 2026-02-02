#!/usr/bin/env python3
"""
🎓 TRM Fine-Tuning Trainer

Fine-tunes Tiny Recursive Model (TRM) on collected data.
Based on: https://github.com/SamsungSAILMontreal/TinyRecursiveModels
"""

import os
import json
import sys
from pathlib import Path
from typing import List, Dict, Any
from dataclasses import dataclass
import subprocess


@dataclass
class TRMTrainingConfig:
    """TRM training configuration."""
    # Model architecture
    arch: str = "trm"  # Architecture type
    L_layers: int = 2  # Number of layers
    H_cycles: int = 3  # High-level cycles
    L_cycles: int = 6  # Low-level cycles
    mlp_t: bool = False  # Use MLP transformer
    pos_encodings: str = "none"  # Position encodings
    
    # Training hyperparameters
    epochs: int = 50000
    eval_interval: int = 5000
    lr: float = 1e-4
    puzzle_emb_lr: float = 1e-4
    weight_decay: float = 1.0
    puzzle_emb_weight_decay: float = 1.0
    global_batch_size: int = 128
    
    # Data
    data_paths: List[str] = None  # Will be set from training data
    
    # Output
    output_dir: str = "./models/trm-finetuned"
    run_name: str = "trm-homelab-finetune"
    ema: bool = True  # Exponential moving average


class TRMTrainer:
    """TRM model trainer."""
    
    def __init__(
        self,
        config: TRMTrainingConfig,
        trm_repo_path: str = None
    ):
        self.config = config
        self.trm_repo_path = Path(trm_repo_path or os.getenv("TRM_REPO_PATH", "./TinyRecursiveModels"))
        self.pretrain_script = self.trm_repo_path / "pretrain.py"
    
    def prepare_dataset(self, training_data_path: str, output_dir: str) -> str:
        """Prepare training data in TRM format."""
        print(f"📊 Preparing dataset from {training_data_path}...")
        
        # Use the dataset builder to convert to TRM format
        trm_data_dir = Path(output_dir) / "trm_data"
        trm_data_dir.mkdir(parents=True, exist_ok=True)
        
        # Import and run the dataset builder
        import subprocess
        builder_script = Path(__file__).parent / "build_runbook_dataset.py"
        
        result = subprocess.run(
            [sys.executable, str(builder_script), "--input", training_data_path, "--output", str(trm_data_dir)],
            capture_output=True,
            text=True
        )
        
        if result.returncode != 0:
            print(f"❌ Dataset preparation failed:")
            print(result.stderr)
            raise RuntimeError(f"Dataset preparation failed: {result.stderr}")
        
        print(result.stdout)
        print(f"✅ Prepared dataset at {trm_data_dir}")
        return str(trm_data_dir)
    
    def train(self, data_dir: str) -> str:
        """Train TRM model."""
        print(f"🎓 Starting TRM training...")
        print(f"   Data: {data_dir}")
        print(f"   Output: {self.config.output_dir}")
        
        # Convert to absolute path
        data_dir_abs = str(Path(data_dir).absolute())
        output_dir_abs = str(Path(self.config.output_dir).absolute())
        
        # Build training command
        cmd = [
            sys.executable,
            str(self.pretrain_script),
            f"arch={self.config.arch}",
            f"data_paths=[{data_dir_abs}]",
            'evaluators=[]',  # No evaluators for now
            f"epochs={self.config.epochs}",
            f"eval_interval={self.config.eval_interval}",
            f"lr={self.config.lr}",
            f"puzzle_emb_lr={self.config.puzzle_emb_lr}",
            f"weight_decay={self.config.weight_decay}",
            f"puzzle_emb_weight_decay={self.config.puzzle_emb_weight_decay}",
            f"global_batch_size={self.config.global_batch_size}",
            f"arch.L_layers={self.config.L_layers}",
            f"arch.H_cycles={self.config.H_cycles}",
            f"arch.L_cycles={self.config.L_cycles}",
            f"arch.pos_encodings={self.config.pos_encodings}",
            f"+run_name={self.config.run_name}",
        ]
        
        if self.config.mlp_t:
            cmd.append("arch.mlp_t=True")
        
        if self.config.ema:
            cmd.append("ema=True")
        
        print(f"🚀 Running: {' '.join(cmd)}")
        
        # Change to TRM repo directory
        original_cwd = os.getcwd()
        os.chdir(self.trm_repo_path)
        
        try:
            # Run training
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=86400  # 24 hour timeout
            )
            
            if result.returncode != 0:
                print(f"❌ Training failed:")
                print(result.stderr)
                raise RuntimeError(f"Training failed: {result.stderr}")
            
            print("✅ Training completed successfully")
            print(result.stdout[-1000:])  # Print last 1000 chars
            
        finally:
            os.chdir(original_cwd)
        
        return self.config.output_dir
    
    def export_model(self, model_dir: str, export_path: str) -> str:
        """Export trained model for deployment."""
        print(f"📦 Exporting model from {model_dir} to {export_path}...")
        
        export_path = Path(export_path)
        export_path.mkdir(parents=True, exist_ok=True)
        
        # Copy model files
        model_path = Path(model_dir)
        if model_path.exists():
            import shutil
            shutil.copytree(model_path, export_path, dirs_exist_ok=True)
            print(f"✅ Model exported to {export_path}")
        else:
            raise FileNotFoundError(f"Model directory not found: {model_dir}")
        
        return str(export_path)


def main():
    """Main entry point for training."""
    import argparse
    
    parser = argparse.ArgumentParser(description="Train TRM model on runbook data")
    parser.add_argument(
        "--training-data",
        type=str,
        default=os.getenv("TRAINING_DATA", "./data/runbook_test_dataset.jsonl"),
        help="Path to training data JSONL file"
    )
    parser.add_argument(
        "--output-dir",
        type=str,
        default=os.getenv("OUTPUT_DIR", "./models/trm-runbook-only"),
        help="Output directory for trained model"
    )
    parser.add_argument(
        "--trm-repo-path",
        type=str,
        default=os.getenv("TRM_REPO_PATH", "./TinyRecursiveModels"),
        help="Path to TinyRecursiveModels repository"
    )
    parser.add_argument(
        "--epochs",
        type=int,
        default=int(os.getenv("EPOCHS", "10000")),
        help="Number of training epochs"
    )
    parser.add_argument(
        "--eval-interval",
        type=int,
        default=int(os.getenv("EVAL_INTERVAL", "1000")),
        help="Evaluation interval"
    )
    parser.add_argument(
        "--run-name",
        type=str,
        default=os.getenv("RUN_NAME", "trm-runbook-finetune"),
        help="Run name for training"
    )
    
    args = parser.parse_args()
    
    # Check if TRM repo exists
    trm_repo_path = Path(args.trm_repo_path)
    if not trm_repo_path.exists():
        print(f"❌ TRM repository not found at: {trm_repo_path}")
        print()
        print("Please clone the TRM repository first:")
        print(f"  git clone https://github.com/SamsungSAILMontreal/TinyRecursiveModels.git {trm_repo_path}")
        print()
        print("Or set TRM_REPO_PATH environment variable:")
        print(f"  export TRM_REPO_PATH=/path/to/TinyRecursiveModels")
        sys.exit(1)
    
    pretrain_script = trm_repo_path / "pretrain.py"
    if not pretrain_script.exists():
        print(f"❌ TRM pretrain.py not found at: {pretrain_script}")
        print("   Make sure you've cloned the complete TRM repository")
        sys.exit(1)
    
    # Check if training data exists
    if not Path(args.training_data).exists():
        print(f"❌ Training data not found: {args.training_data}")
        print("   Run: python src/test_trm_runbook.py first")
        sys.exit(1)
    
    config = TRMTrainingConfig(
        epochs=args.epochs,
        eval_interval=args.eval_interval,
        output_dir=args.output_dir,
        run_name=args.run_name
    )
    
    trainer = TRMTrainer(config, trm_repo_path=str(trm_repo_path))
    
    print("=" * 60)
    print("🎓 TRM Fine-Tuning on Runbook Data")
    print("=" * 60)
    print(f"Training Data: {args.training_data}")
    print(f"Output Dir: {args.output_dir}")
    print(f"TRM Repo: {trm_repo_path}")
    print(f"Epochs: {args.epochs}")
    print(f"Eval Interval: {args.eval_interval}")
    print(f"Run Name: {args.run_name}")
    print()
    
    # Prepare dataset
    data_dir = trainer.prepare_dataset(args.training_data, args.output_dir)
    
    # Train model
    model_dir = trainer.train(data_dir)
    
    # Export model
    export_path = trainer.export_model(model_dir, f"{args.output_dir}/export")
    
    print()
    print("=" * 60)
    print(f"🎉 Training complete! Model at: {export_path}")
    print("=" * 60)


if __name__ == "__main__":
    main()

