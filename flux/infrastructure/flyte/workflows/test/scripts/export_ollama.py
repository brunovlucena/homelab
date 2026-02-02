#!/usr/bin/env python3
"""
Export fine-tuned FunctionGemma model to Ollama format

This script:
1. Merges LoRA adapters with base MLX model using mlx_lm.fuse
2. Creates a Modelfile for Ollama
3. Prepares model for Ollama import

Usage:
    python export_ollama.py \
        --base-model /path/to/mlx-model \
        --adapter-path /path/to/adapters \
        --output /path/to/output \
        --model-name agent-sre:latest
"""
import argparse
import os
import sys
import shutil
from pathlib import Path


def create_modelfile(output_path: str, model_name: str, base_model_path: str) -> str:
    """
    Create Modelfile for Ollama.
    
    Args:
        output_path: Output directory for Modelfile
        model_name: Name for the Ollama model (e.g., agent-sre:latest)
        base_model_path: Path to merged model directory
    
    Returns:
        Path to created Modelfile
    """
    modelfile_path = os.path.join(output_path, "Modelfile")
    
    # Get absolute path to model
    absolute_model_path = os.path.abspath(base_model_path)
    
    # Create Modelfile content for FunctionGemma/agent-sre
    modelfile_content = f"""FROM {absolute_model_path}

# Model metadata
PARAMETER temperature 0.7
PARAMETER top_p 0.9
PARAMETER top_k 40
PARAMETER num_ctx 4096

# System prompt for Agent-SRE
SYSTEM \"\"\"You are an AI-powered Site Reliability Engineering (SRE) assistant fine-tuned on infrastructure runbooks and incident response procedures.

Your capabilities:
- Analyze Prometheus alerts and provide remediation commands
- Understand Flux/GitOps reconciliation issues
- Generate kubectl commands for Kubernetes troubleshooting
- Map alerts to runbook procedures
- Provide context-aware incident response guidance

Based on the training data from RUNBOOK.md, you should:
1. Recognize alert patterns and symptoms
2. Provide investigation steps
3. Generate exact remediation commands (flux reconcile, kubectl, etc.)
4. Follow the runbook procedures for each alert type

Always provide accurate, actionable commands based on your fine-tuning.\"\"\"

# Template for instruction-following format
TEMPLATE \"\"\"<|im_start|>user
{{{{ .Prompt }}}}<|im_end|>
<|im_start|>assistant
{{{{ .Response }}}}<|im_end|>
\"\"\"
"""
    
    with open(modelfile_path, 'w', encoding='utf-8') as f:
        f.write(modelfile_content)
    
    print(f"✅ Modelfile created: {modelfile_path}")
    return modelfile_path


def merge_adapters(base_model_path: str, adapter_path: str, output_path: str) -> str:
    """
    Merge LoRA adapters with base MLX model.
    
    Args:
        base_model_path: Path to base MLX model
        adapter_path: Path to LoRA adapters
        output_path: Output path for merged model
    
    Returns:
        Path to merged model
    """
    import subprocess
    
    print(f"🔄 Merging LoRA adapters with base model...")
    print(f"   Base model: {base_model_path}")
    print(f"   Adapter path: {adapter_path}")
    print(f"   Output: {output_path}")
    
    # Create output directory
    Path(output_path).parent.mkdir(parents=True, exist_ok=True)
    
    # Use mlx_lm.fuse to merge adapters
    cmd = [
        sys.executable, "-m", "mlx_lm.fuse",
        "--model", base_model_path,
        "--adapter-path", adapter_path,
        "--save-path", output_path,
    ]
    
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"❌ Error merging adapters: {result.stderr}")
        if result.stdout:
            print(f"   stdout: {result.stdout}")
        raise RuntimeError(f"Failed to merge adapters: {result.stderr}")
    
    print(f"✅ Merged model saved to: {output_path}")
    return output_path


def export_to_ollama(
    base_model_path: str,
    adapter_path: str,
    output_path: str,
    model_name: str = "agent-sre:latest"
) -> dict:
    """
    Export fine-tuned model to Ollama format.
    
    Args:
        base_model_path: Path to base MLX model
        adapter_path: Path to LoRA adapters
        output_path: Output directory for Ollama model
        model_name: Name for Ollama model (e.g., agent-sre:latest)
    
    Returns:
        Dictionary with paths and instructions
    """
    print("📦 Exporting Agent-SRE model to Ollama format...")
    
    # Clean up existing output directory
    output_dir = Path(output_path)
    if output_dir.exists():
        print(f"🗑️  Cleaning up existing directory: {output_path}")
        shutil.rmtree(output_path)
    
    output_dir.mkdir(parents=True, exist_ok=True)
    
    # Step 1: Merge adapters with base model
    merged_model_path = os.path.join(output_path, "model")
    merge_adapters(base_model_path, adapter_path, merged_model_path)
    
    # Step 2: Create Modelfile
    modelfile_path = create_modelfile(output_path, model_name, merged_model_path)
    
    # Step 3: Create instructions
    instructions = f"""📋 Ollama Import Instructions:

1. Copy the model directory to your Ollama server:
   scp -r {output_path} user@your-ollama-server:/path/to/models/

2. On the Ollama server, create the model:
   ollama create {model_name} -f {modelfile_path}

3. Test the model:
   ollama run {model_name} "Alert: FluxReconciliationFailure. How should I resolve this?"

4. The model will be available via Ollama API at your server's endpoint.

📋 Local Ollama Commands (if running locally):
   ollama create {model_name} -f {modelfile_path}
   ollama run {model_name} "What should I do when Flux reconciliation fails?"

📋 Using in Agent-SRE LambdaAgent:
   Update lambdaagent.yaml:
   ai:
     provider: ollama
     endpoint: "http://ollama:11434"
     model: "{model_name}"
"""
    
    instructions_path = os.path.join(output_path, "OLLAMA_INSTRUCTIONS.txt")
    with open(instructions_path, 'w', encoding='utf-8') as f:
        f.write(instructions)
    
    print(f"✅ Model exported successfully!")
    print(f"   Output directory: {output_path}")
    print(f"   Modelfile: {modelfile_path}")
    print(f"   Instructions: {instructions_path}")
    print()
    print(instructions)
    
    return {
        "output_path": output_path,
        "modelfile_path": modelfile_path,
        "merged_model_path": merged_model_path,
        "model_name": model_name,
        "instructions_path": instructions_path,
    }


def main():
    parser = argparse.ArgumentParser(
        description="Export fine-tuned FunctionGemma model to Ollama format"
    )
    parser.add_argument(
        "--base-model",
        required=True,
        help="Path to base MLX model (e.g., /path/to/mlx-functiongemma-270m)"
    )
    parser.add_argument(
        "--adapter-path",
        required=True,
        help="Path to LoRA adapters directory"
    )
    parser.add_argument(
        "--output",
        required=True,
        help="Output directory for Ollama model"
    )
    parser.add_argument(
        "--model-name",
        default="agent-sre:latest",
        help="Name for Ollama model (default: agent-sre:latest)"
    )
    
    args = parser.parse_args()
    
    # Validate paths
    if not os.path.exists(args.base_model):
        print(f"❌ Base model path does not exist: {args.base_model}")
        sys.exit(1)
    
    if not os.path.exists(args.adapter_path):
        print(f"❌ Adapter path does not exist: {args.adapter_path}")
        sys.exit(1)
    
    try:
        result = export_to_ollama(
            base_model_path=args.base_model,
            adapter_path=args.adapter_path,
            output_path=args.output,
            model_name=args.model_name
        )
        print("\n✅ Export completed successfully!")
        return 0
    except Exception as e:
        print(f"\n❌ Export failed: {e}")
        import traceback
        traceback.print_exc()
        return 1


if __name__ == "__main__":
    import subprocess
    sys.exit(main())

