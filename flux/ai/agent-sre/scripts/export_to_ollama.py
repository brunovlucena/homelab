#!/usr/bin/env python3
"""
Export Fine-Tuned Model to Ollama - Phase 4

Converts fine-tuned MLX model to Ollama Modelfile format for deployment.
"""
import argparse
import json
import sys
from pathlib import Path
import structlog

logger = structlog.get_logger()


def create_ollama_modelfile(
    model_path: str,
    model_name: str,
    base_model: str = "functiongemma:270m",
    output_path: Optional[str] = None
) -> str:
    """
    Create Ollama Modelfile from fine-tuned model.
    
    Args:
        model_path: Path to fine-tuned MLX model directory
        model_name: Name for the Ollama model
        base_model: Base model name in Ollama
        output_path: Output path for Modelfile (default: model_path/Modelfile)
    
    Returns:
        Path to created Modelfile
    """
    model_dir = Path(model_path)
    
    if not model_dir.exists():
        raise ValueError(f"Model directory not found: {model_path}")
    
    # Create Modelfile content
    modelfile_content = f"""FROM {base_model}

# Fine-tuned Agent-SRE model
# Trained on SRE remediation dataset
# Model path: {model_path}

PARAMETER temperature 0.7
PARAMETER top_p 0.9
PARAMETER top_k 40
PARAMETER num_predict 1024

# System prompt for SRE remediation
SYSTEM \"\"\"You are an expert SRE agent specialized in analyzing Prometheus alerts
and selecting appropriate remediation actions using Lambda functions.

Your task is to:
1. Analyze the alert context (name, labels, annotations)
2. Select the most appropriate Lambda function
3. Extract and format the required parameters
4. Provide clear reasoning for your selection

Available Lambda functions:
- flux-reconcile-kustomization: For Flux Kustomization reconciliation
- flux-reconcile-gitrepository: For Flux GitRepository reconciliation
- flux-reconcile-helmrelease: For Flux HelmRelease reconciliation
- pod-restart: For restarting pods or deployments
- pod-check-status: For checking pod health status
- scale-deployment: For scaling deployments
- check-pvc-status: For checking PVC status

Always respond with valid JSON in this format:
{{
  "lambda_function": "<function-name>",
  "parameters": {{
    "name": "<resource-name>",
    "namespace": "<namespace>",
    ...
  }},
  "reasoning": "<explanation>"
}}
\"\"\"
"""
    
    # Determine output path
    if output_path:
        output_file = Path(output_path)
    else:
        output_file = model_dir / "Modelfile"
    
    output_file.parent.mkdir(parents=True, exist_ok=True)
    
    # Write Modelfile
    with open(output_file, "w") as f:
        f.write(modelfile_content)
    
    logger.info("modelfile_created", path=str(output_file), model_name=model_name)
    
    return str(output_file)


def create_ollama_import_script(model_path: str, model_name: str) -> str:
    """Create script to import model into Ollama."""
    script = f"""#!/bin/bash
# Import fine-tuned model into Ollama

MODEL_PATH="{model_path}"
MODEL_NAME="{model_name}"

echo "📦 Importing model into Ollama..."
echo "   Model: $MODEL_NAME"
echo "   Path: $MODEL_PATH"

# Check if Ollama is available
if ! command -v ollama &> /dev/null; then
    echo "❌ Ollama not found. Please install Ollama first."
    exit 1
fi

# Create model using Modelfile
cd "$MODEL_PATH"
ollama create "$MODEL_NAME" -f Modelfile

if [ $? -eq 0 ]; then
    echo "✅ Model imported successfully!"
    echo "   Use it with: ollama run $MODEL_NAME"
else
    echo "❌ Failed to import model"
    exit 1
fi
"""
    return script


def main():
    parser = argparse.ArgumentParser(description="Export fine-tuned model to Ollama")
    parser.add_argument(
        "--model",
        required=True,
        help="Path to fine-tuned MLX model directory"
    )
    parser.add_argument(
        "--name",
        required=True,
        help="Name for the Ollama model (e.g., agent-sre-functiongemma-270m)"
    )
    parser.add_argument(
        "--base-model",
        default="functiongemma:270m",
        help="Base model name in Ollama"
    )
    parser.add_argument(
        "--output",
        help="Output path for Modelfile (default: model_path/Modelfile)"
    )
    parser.add_argument(
        "--create-import-script",
        action="store_true",
        help="Create import script for Ollama"
    )
    
    args = parser.parse_args()
    
    try:
        # Create Modelfile
        modelfile_path = create_ollama_modelfile(
            model_path=args.model,
            model_name=args.name,
            base_model=args.base_model,
            output_path=args.output
        )
        
        print(f"✅ Created Modelfile: {modelfile_path}")
        
        # Create import script if requested
        if args.create_import_script:
            script_content = create_ollama_import_script(args.model, args.name)
            script_path = Path(args.model) / "import_to_ollama.sh"
            
            with open(script_path, "w") as f:
                f.write(script_content)
            
            script_path.chmod(0o755)
            
            print(f"✅ Created import script: {script_path}")
            print(f"\n📝 To import into Ollama, run:")
            print(f"   {script_path}")
        
        print(f"\n📋 Next steps:")
        print(f"   1. Review Modelfile: {modelfile_path}")
        print(f"   2. Import to Ollama: ollama create {args.name} -f {modelfile_path}")
        print(f"   3. Test: ollama run {args.name}")
        
    except Exception as e:
        logger.error("export_failed", error=str(e), exc_info=True)
        print(f"❌ Export failed: {e}")
        sys.exit(1)


if __name__ == "__main__":
    from typing import Optional
    main()

