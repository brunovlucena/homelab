"""
Flyte Workflow for Agent Model Training Pipeline

This workflow automates the complete training pipeline:
1. Prepare dataset from RUNBOOK.md
2. Convert model to MLX format (if needed)
3. Fine-tune with LoRA
4. Evaluate model
5. Register model in MLflow
6. Store model artifacts in MinIO
7. Update model version

Deployed on Forge cluster for GPU-accelerated training.
"""
from flytekit import task, workflow, Resources, current_context
from flytekit.types.file import FlyteFile
from flytekit.types.directory import FlyteDirectory
from typing import Dict, Any
import os
import json
import urllib.request
from datetime import datetime, timedelta
from pathlib import Path


# Define container image with MLX dependencies AND flytekit
# The image must have both MLX/MLX-LM for training AND flytekit for execution
# Use flyte-sandbox-training image which has all dependencies including flytekit
training_image = "ghcr.io/brunovlucena/flyte-sandbox-training:latest"


@task(
    container_image=training_image,
    requests=Resources(cpu="100m", mem="256Mi"),
    limits=Resources(cpu="500m", mem="512Mi"),
    retries=1,
)
def download_runbook_if_url(
    runbook_path: str,
    agent_name: str,
) -> str:
    """
    Download RUNBOOK.md if it's a URL, otherwise return the path as-is.
    
    Args:
        runbook_path: Path to RUNBOOK.md file (can be URL or local path)
        agent_name: Name of the agent
    
    Returns:
        Local path to RUNBOOK.md file
    """
    import urllib.request
    from pathlib import Path
    
    # If it's a URL, download it
    if runbook_path.startswith("http"):
        local_runbook = f"/tmp/{agent_name}/RUNBOOK.md"
        Path(f"/tmp/{agent_name}").mkdir(parents=True, exist_ok=True)
        urllib.request.urlretrieve(runbook_path, local_runbook)
        return local_runbook
    
    # Otherwise return as-is
    return runbook_path


@task(
    container_image=training_image,
    requests=Resources(cpu="200m", mem="512Mi"),
    limits=Resources(cpu="500m", mem="1Gi"),
    retries=2,
)
def prepare_dataset(
    runbook_path: str,
    output_dir: str,
    train_split: float = 0.8,
    val_split: float = 0.1,
) -> Dict[str, str]:
    """
    Prepare training dataset from RUNBOOK.md.
    
    Args:
        runbook_path: Path to RUNBOOK.md file (must be local path)
        output_dir: Output directory for training data
        train_split: Training set split ratio
        val_split: Validation set split ratio
    
    Returns:
        Dictionary with paths to train/val/test JSONL files
    """
    import subprocess
    import sys
    
    # Run dataset preparation script
    script_path = Path("/app/scripts/prepare_dataset.py")
    
    cmd = [
        sys.executable,
        str(script_path),
        "--runbook", runbook_path,
        "--output-dir", output_dir,
        "--format", "instruction",
        "--train-split", str(train_split),
        "--val-split", str(val_split),
    ]
    
    result = subprocess.run(cmd, capture_output=True, text=True, check=True)
    
    return {
        "train": f"{output_dir}/train.jsonl",
        "val": f"{output_dir}/val.jsonl",
        "test": f"{output_dir}/test.jsonl",
        "stats": result.stdout,
    }


@task(
    container_image=training_image,
    requests=Resources(cpu="200m", mem="512Mi"),
    limits=Resources(cpu="500m", mem="1Gi"),
    retries=1,
)
def convert_model_to_mlx(
    model_name: str,
    output_path: str,
) -> str:
    """
    Convert HuggingFace model to MLX format.
    
    Args:
        model_name: HuggingFace model name (e.g., google/functiongemma-270m-it)
        output_path: Output path for MLX model
    
    Returns:
        Path to converted MLX model
    """
    import subprocess
    import sys
    
    # Check if already converted
    if Path(output_path).exists():
        print(f"✅ Model already converted: {output_path}")
        return output_path
    
    # Convert using mlx_lm
    cmd = [
        sys.executable, "-m", "mlx_lm.convert",
        "--hf-path", model_name,
        "--mlx-path", output_path,
    ]
    
    subprocess.run(cmd, check=True)
    
    return output_path


@task(
    container_image=training_image,
    requests=Resources(cpu="200m", mem="512Mi"),  # Reduced for platform limits
    limits=Resources(cpu="500m", mem="1Gi"),
    retries=1,
    timeout=timedelta(hours=2),  # Training can take time
)
def train_model_lora(
    model_path: str,
    train_data: str,
    val_data: str,
    output_dir: str,
    learning_rate: float = 1e-4,
    batch_size: int = 4,
    iters: int = 1000,
    lora_layers: int = 16,
    lora_rank: int = 8,
    lora_alpha: int = 16,
) -> Dict[str, str]:
    """
    Fine-tune model using LoRA.
    
    Args:
        model_path: Path to MLX model
        train_data: Path to training JSONL
        val_data: Path to validation JSONL
        output_dir: Output directory for adapters
        learning_rate: Learning rate
        batch_size: Batch size
        iters: Number of iterations
        lora_layers: Number of LoRA layers
        lora_rank: LoRA rank
        lora_alpha: LoRA alpha
    
    Returns:
        Dictionary with adapter path and training metrics
    """
    import subprocess
    import sys
    
    # Run training script
    script_path = Path("/app/scripts/train.py")
    
    cmd = [
        sys.executable,
        str(script_path),
        "--model", model_path,
        "--train-data", train_data,
        "--val-data", val_data,
        "--output-dir", output_dir,
        "--learning-rate", str(learning_rate),
        "--batch-size", str(batch_size),
        "--iters", str(iters),
        "--lora-layers", str(lora_layers),
        "--lora-rank", str(lora_rank),
        "--lora-alpha", str(lora_alpha),
        "--skip-convert",  # Model already converted
    ]
    
    result = subprocess.run(cmd, capture_output=True, text=True, check=True)
    
    adapter_path = f"{output_dir}/adapters"
    
    return {
        "adapter_path": adapter_path,
        "output_dir": output_dir,
        "training_log": result.stdout,
    }


@task(
    container_image=training_image,
    requests=Resources(cpu="200m", mem="512Mi"),
    limits=Resources(cpu="500m", mem="1Gi"),
    retries=1,
)
def evaluate_model(
    model_path: str,
    adapter_path: str,
    test_data: str,
    max_tokens: int = 512,
) -> Dict[str, float]:
    """
    Evaluate fine-tuned model on test dataset.
    
    Args:
        model_path: Path to base MLX model
        adapter_path: Path to LoRA adapters
        test_data: Path to test JSONL
        max_tokens: Maximum tokens to generate
    
    Returns:
        Dictionary with evaluation metrics
    """
    import subprocess
    import sys
    
    # Run evaluation script
    script_path = Path("/app/scripts/evaluate.py")
    
    cmd = [
        sys.executable,
        str(script_path),
        "--model", model_path,
        "--adapter", adapter_path,
        "--test-data", test_data,
        "--max-tokens", str(max_tokens),
    ]
    
    result = subprocess.run(cmd, capture_output=True, text=True, check=True)
    
    # Parse accuracy from output (simple extraction)
    accuracy = 0.0
    for line in result.stdout.split('\n'):
        if 'Accuracy:' in line:
            try:
                accuracy = float(line.split('Accuracy:')[1].split('%')[0].strip()) / 100.0
            except:
                pass
    
    return {
        "accuracy": accuracy,
        "evaluation_log": result.stdout,
    }


@task(
    container_image=training_image,
    requests=Resources(cpu="200m", mem="512Mi"),
    limits=Resources(cpu="500m", mem="1Gi"),
    retries=2,
)
def register_model_mlflow(
    model_path: str,
    adapter_path: str,
    metrics: Dict[str, float],
    agent_name: str,
    runbook_version: str,
) -> str:
    """
    Register model in MLflow model registry.
    
    Args:
        model_path: Path to base MLX model
        adapter_path: Path to LoRA adapters
        metrics: Evaluation metrics
        agent_name: Name of the agent (e.g., agent-sre)
        runbook_version: Version of RUNBOOK.md used for training
    
    Returns:
        MLflow model version URI
    """
    import mlflow
    
    # MLflow tracking URI (from Forge cluster)
    mlflow.set_tracking_uri("http://mlflow.ml-platform.svc.forge.remote:5000")
    
    experiment_name = f"{agent_name}-training"
    mlflow.set_experiment(experiment_name)
    
    with mlflow.start_run() as run:
        # Log parameters
        mlflow.log_param("agent_name", agent_name)
        mlflow.log_param("base_model", Path(model_path).name)
        mlflow.log_param("runbook_version", runbook_version)
        mlflow.log_param("training_date", datetime.now().isoformat())
        
        # Log metrics
        for key, value in metrics.items():
            mlflow.log_metric(key, value)
        
        # Log model artifacts
        mlflow.log_artifacts(adapter_path, "adapters")
        mlflow.log_artifact(model_path, "base_model")
        
        # Register model
        model_name = f"{agent_name}-functiongemma"
        mlflow.register_model(
            f"runs:/{run.info.run_id}/adapters",
            model_name
        )
        
        return f"models:/{model_name}/latest"


@task(
    container_image=training_image,
    requests=Resources(cpu="200m", mem="512Mi"),
    limits=Resources(cpu="500m", mem="1Gi"),
    retries=2,
)
def store_model_minio(
    adapter_path: str,
    agent_name: str,
    model_version: str,
) -> str:
    """
    Store model artifacts in MinIO for long-term storage.
    
    Args:
        adapter_path: Path to LoRA adapters
        agent_name: Name of the agent
        model_version: Model version (e.g., v1.0.0)
    
    Returns:
        MinIO object path
    """
    from minio import Minio
    import tarfile
    import tempfile
    
    # MinIO client (use local MinIO service)
    minio_endpoint = os.getenv("AWS_ENDPOINT", "minio.minio.svc.cluster.local:9000").replace("http://", "").replace("https://", "")
    client = Minio(
        minio_endpoint,
        access_key=os.getenv("AWS_ACCESS_KEY_ID", os.getenv("MINIO_ACCESS_KEY", "minioadmin")),
        secret_key=os.getenv("AWS_SECRET_ACCESS_KEY", os.getenv("MINIO_SECRET_KEY", "minioadmin")),
        secure=False
    )
    
    # Create tarball of adapters
    with tempfile.NamedTemporaryFile(suffix=".tar.gz", delete=False) as tmp:
        with tarfile.open(tmp.name, "w:gz") as tar:
            tar.add(adapter_path, arcname="adapters")
        
        # Upload to MinIO
        # Use flyte-data bucket (same as Flyte uses) or create ml-models bucket
        bucket = os.getenv("MINIO_BUCKET", "ml-models")
        object_name = f"{agent_name}/functiongemma-270m/{model_version}/adapters.tar.gz"
        
        # Ensure bucket exists
        if not client.bucket_exists(bucket):
            client.make_bucket(bucket)
        
        client.fput_object(bucket, object_name, tmp.name)
        
        # Cleanup
        Path(tmp.name).unlink()
    
    return f"s3://{bucket}/{object_name}"


@task(
    container_image=training_image,
    requests=Resources(cpu="200m", mem="512Mi"),
    limits=Resources(cpu="500m", mem="1Gi"),
    retries=2,
)
def export_model_ollama(
    base_model_path: str,
    adapter_path: str,
    output_dir: str,
    agent_name: str,
    model_version: str,
) -> Dict[str, str]:
    """
    Export fine-tuned model to Ollama format.
    
    This task:
    1. Merges LoRA adapters with base MLX model
    2. Creates Modelfile for Ollama
    3. Prepares model for Ollama import
    
    Args:
        base_model_path: Path to base MLX model
        adapter_path: Path to LoRA adapters
        output_dir: Output directory for Ollama model
        agent_name: Name of the agent (e.g., agent-sre)
        model_version: Model version (e.g., v1.0.0)
    
    Returns:
        Dictionary with paths to exported model and Modelfile
    """
    import subprocess
    import sys
    
    # Export to Ollama format
    # Model name format: agent-sre:{model_version} (e.g., agent-sre:v20251224-abc12345)
    ollama_model_name = f"{agent_name}:{model_version}"
    ollama_output_path = Path(f"{output_dir}/ollama-{model_version}")
    ollama_output_path.mkdir(parents=True, exist_ok=True)
    
    # Try to find export script in common locations
    script_paths = [
        Path("/app/scripts/export_ollama.py"),
        Path("/workflows/scripts/export_ollama.py"),
        Path("./scripts/export_ollama.py"),
    ]
    
    script_path = None
    for path in script_paths:
        if path.exists():
            script_path = path
            break
    
    if script_path and script_path.exists():
        # Use existing export script
        cmd = [
            sys.executable,
            str(script_path),
            "--base-model", base_model_path,
            "--adapter-path", adapter_path,
            "--output", str(ollama_output_path),
            "--model-name", ollama_model_name,
        ]
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        export_log = result.stdout
        modelfile_path = ollama_output_path / "Modelfile"
    else:
        # Inline export: merge adapters and create Modelfile
        from mlx_lm import fuse as merge_lora
        import shutil
        
        # Merge LoRA adapters with base model
        merged_model_path = ollama_output_path / "model"
        merged_model_path.mkdir(parents=True, exist_ok=True)
        
        # Use mlx_lm.fuse to merge adapters
        cmd = [
            sys.executable, "-m", "mlx_lm.fuse",
            "--model", base_model_path,
            "--adapter-path", adapter_path,
            "--save-path", str(merged_model_path),
        ]
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        
        # Create Modelfile for Ollama with agent-sre specific configuration
        modelfile_path = ollama_output_path / "Modelfile"
        with open(modelfile_path, "w") as f:
            f.write(f"FROM {merged_model_path}\n\n")
            f.write("# Model metadata\n")
            f.write("PARAMETER temperature 0.7\n")
            f.write("PARAMETER top_p 0.9\n")
            f.write("PARAMETER top_k 40\n")
            f.write("PARAMETER num_ctx 4096\n\n")
            f.write("# System prompt for Agent-SRE\n")
            f.write('SYSTEM """You are an AI-powered Site Reliability Engineering (SRE) assistant fine-tuned on infrastructure runbooks and incident response procedures.\n\n')
            f.write("Your capabilities:\n")
            f.write("- Analyze Prometheus alerts and provide remediation commands\n")
            f.write("- Understand Flux/GitOps reconciliation issues\n")
            f.write("- Generate kubectl commands for Kubernetes troubleshooting\n")
            f.write("- Map alerts to runbook procedures\n")
            f.write("- Provide context-aware incident response guidance\n\n")
            f.write("Based on the training data from RUNBOOK.md, you should:\n")
            f.write("1. Recognize alert patterns and symptoms\n")
            f.write("2. Provide investigation steps\n")
            f.write("3. Generate exact remediation commands (flux reconcile, kubectl, etc.)\n")
            f.write('4. Follow the runbook procedures for each alert type\n\nAlways provide accurate, actionable commands based on your fine-tuning."""\n\n')
            f.write("# Template for instruction-following format\n")
            f.write('TEMPLATE """<|im_start|>user\n{{ .Prompt }}<|im_end|>\n<|im_start|>assistant\n{{ .Response }}<|im_end|>\n"""\n')
        
        export_log = f"Merged LoRA adapters and created Modelfile for {ollama_model_name}\nModel saved to: {ollama_output_path}"
    
    return {
        "ollama_output_path": str(ollama_output_path),
        "modelfile_path": str(modelfile_path),
        "model_name": ollama_model_name,  # Format: agent-sre:v20251224-abc12345
        "export_log": export_log,
    }


@task(
    container_image=training_image,
    requests=Resources(cpu="100m", mem="256Mi"),
    limits=Resources(cpu="500m", mem="512Mi"),
    retries=2,
)
def trigger_ollama_import(
    model_name: str,
    model_version: str,
    minio_path: str,
) -> Dict[str, str]:
    """
    Automatically trigger Kubernetes job to import model into Ollama.
    
    Args:
        model_name: Ollama model name (e.g., agent-sre:v20251224-abc12345)
        model_version: Model version string
        minio_path: MinIO path where model is stored
    
    Returns:
        Dictionary with job name and status
    """
    import subprocess
    import sys
    import json
    
    # Extract version from model_name if it's in format agent-sre:v20251224-abc12345
    version_suffix = model_version
    
    # Create unique job name
    execution_id = current_context().execution_id.name[:8]
    job_name = f"import-agent-sre-ollama-{execution_id}"
    
    # Create Kubernetes job manifest
    job_manifest = {
        "apiVersion": "batch/v1",
        "kind": "Job",
        "metadata": {
            "name": job_name,
            "namespace": "ai",
            "labels": {
                "app": "flyte-workflow",
                "component": "ollama-import",
                "agent": "agent-sre",
                "model-version": version_suffix,
            }
        },
        "spec": {
            "backoffLimit": 3,
            "ttlSecondsAfterFinished": 3600,
            "template": {
                "metadata": {
                    "labels": {
                        "app": "flyte-workflow",
                        "component": "ollama-import",
                        "agent": "agent-sre",
                    }
                },
                "spec": {
                    "restartPolicy": "OnFailure",
                    "serviceAccountName": "ollama-import-sa",
                    "containers": [{
                        "name": "import-model",
                        "image": "python:3.11-slim",
                        "command": ["/bin/sh", "-c"],
                        "args": [f"""
set -e
echo "🤖 Triggering Ollama import for {model_name}..."
echo "   MinIO path: {minio_path}"
echo "   Model version: {version_suffix}"

# Install dependencies
pip install -q minio requests

# Download and import script
python3 << 'PYTHON_SCRIPT'
import os
import sys
import requests
import json
from minio import Minio
import tarfile
import tempfile

ollama_host = os.getenv("OLLAMA_HOST", "ollama.ollama.svc.cluster.local:11434")
model_name = "{model_name}"
minio_path = "{minio_path}"

# Parse MinIO path: s3://bucket/path/to/model.tar.gz
if minio_path.startswith("s3://"):
    parts = minio_path[5:].split("/", 1)
    bucket = parts[0]
    object_name = parts[1] if len(parts) > 1 else ""
else:
    bucket = "ml-models"
    object_name = minio_path

print(f"Downloading from MinIO: {{bucket}}/{{object_name}}")

# Download from MinIO
minio_endpoint = os.getenv("MINIO_ENDPOINT", "minio.minio.svc.cluster.local:9000")
client = Minio(
    minio_endpoint,
    access_key=os.getenv("MINIO_ACCESS_KEY", "minioadmin"),
    secret_key=os.getenv("MINIO_SECRET_KEY", "minioadmin"),
    secure=False
)

with tempfile.NamedTemporaryFile(delete=False, suffix=".tar.gz") as tmp:
    client.fget_object(bucket, object_name, tmp.name)
    print(f"Downloaded to {{tmp.name}}")
    
    # Extract
    extract_dir = "/tmp/ollama-model"
    os.makedirs(extract_dir, exist_ok=True)
    with tarfile.open(tmp.name, "r:gz") as tar:
        tar.extractall(extract_dir)
    print(f"Extracted to {{extract_dir}}")

# Find Modelfile
modelfile_path = os.path.join(extract_dir, "Modelfile")
if not os.path.exists(modelfile_path):
    import glob
    modelfiles = glob.glob(os.path.join(extract_dir, "**/Modelfile"), recursive=True)
    if modelfiles:
        modelfile_path = modelfiles[0]

if not os.path.exists(modelfile_path):
    print("ERROR: Modelfile not found")
    sys.exit(1)

# Read Modelfile
with open(modelfile_path, "r") as f:
    modelfile_content = f.read()

# Update FROM path to use extracted model directory
model_dir = os.path.dirname(modelfile_path)
modelfile_content = modelfile_content.replace("FROM {ABSOLUTE_PATH_TO_MERGED_MODEL}", f"FROM {model_dir}/model")
modelfile_content = modelfile_content.replace("FROM ./model", f"FROM {model_dir}/model")

# Create model in Ollama
url = f"http://{{ollama_host}}/api/create"
data = {{
    "name": model_name,
    "modelfile": modelfile_content,
    "stream": False
}}

print(f"Creating model {{model_name}} in Ollama...")
response = requests.post(url, json=data, timeout=600)
response.raise_for_status()
print(f"✅ Model {{model_name}} created successfully!")

# Verify
verify_url = f"http://{{ollama_host}}/api/tags"
verify_response = requests.get(verify_url)
if verify_response.status_code == 200:
    models = verify_response.json().get("models", [])
    found = any(m.get("name") == model_name for m in models)
    if found:
        print(f"✅ Verified: {{model_name}} is available")
    else:
        print(f"⚠️  Warning: Model created but not found in list")
PYTHON_SCRIPT
"""],
                        "env": [
                            {"name": "OLLAMA_HOST", "value": "ollama.ollama.svc.cluster.local:11434"},
                            {"name": "MINIO_ENDPOINT", "value": "minio.minio.svc.cluster.local:9000"},
                            {"name": "MINIO_ACCESS_KEY", "valueFrom": {"secretKeyRef": {"name": "minio-credentials", "key": "access-key", "optional": True}}},
                            {"name": "MINIO_SECRET_KEY", "valueFrom": {"secretKeyRef": {"name": "minio-credentials", "key": "secret-key", "optional": True}}},
                        ],
                        "resources": {
                            "requests": {"memory": "512Mi", "cpu": "200m"},
                            "limits": {"memory": "2Gi", "cpu": "1000m"}
                        }
                    }]
                }
            }
        }
    }
    
    # Apply job using kubectl
    try:
        # Write manifest to temp file
        import tempfile
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            json.dump(job_manifest, f, indent=2)
            manifest_path = f.name
        
        # Apply using kubectl
        cmd = ["kubectl", "apply", "-f", manifest_path]
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        
        if result.returncode == 0:
            print(f"✅ Ollama import job created: {job_name}")
            return {
                "job_name": job_name,
                "status": "created",
                "model_name": model_name,
            }
        else:
            print(f"⚠️  Warning: Could not create import job: {result.stderr}")
            return {
                "job_name": job_name,
                "status": "failed",
                "error": result.stderr,
            }
    except Exception as e:
        print(f"⚠️  Warning: Error creating import job: {e}")
        return {
            "job_name": job_name,
            "status": "error",
            "error": str(e),
        }


@workflow
def agent_training_pipeline(
    agent_name: str = "agent-sre",
    model_name: str = "google/functiongemma-270m-it",
    runbook_path: str = "https://raw.githubusercontent.com/brunovlucena/homelab/main/flux/ai/agent-sre/docs/RUNBOOK.md",
    learning_rate: float = 1e-4,
    batch_size: int = 4,
    iters: int = 1000,
    lora_layers: int = 16,
    lora_rank: int = 8,
    lora_alpha: int = 16,
) -> Dict[str, str]:
    """
    Complete training pipeline for agent model fine-tuning.
    
    This workflow:
    1. Prepares dataset from RUNBOOK.md
    2. Converts model to MLX format
    3. Fine-tunes with LoRA
    4. Evaluates model
    5. Registers in MLflow
    6. Stores in MinIO
    7. Exports to Ollama format
    8. Updates model version
    
    Args:
        agent_name: Name of the agent
        model_name: HuggingFace model name
        runbook_path: Path to RUNBOOK.md
        learning_rate: Learning rate for training
        batch_size: Batch size
        iters: Training iterations
        lora_layers: Number of LoRA layers
        lora_rank: LoRA rank
        lora_alpha: LoRA alpha
    
    Returns:
        Dictionary with all artifact paths and model version
    """
    # Step 1: Download and prepare dataset
    # First download RUNBOOK.md if it's a URL
    local_runbook = download_runbook_if_url(
        runbook_path=runbook_path,
        agent_name=agent_name,
    )
    
    dataset = prepare_dataset(
        runbook_path=local_runbook,
        output_dir=f"/tmp/{agent_name}/data",
        train_split=0.8,
        val_split=0.1,
    )
    
    # Step 2: Convert model to MLX
    mlx_model_path = convert_model_to_mlx(
        model_name=model_name,
        output_path=f"/tmp/{agent_name}/models/mlx-functiongemma-270m",
    )
    
    # Step 3: Train model
    training_result = train_model_lora(
        model_path=mlx_model_path,
        train_data=dataset["train"],
        val_data=dataset["val"],
        output_dir=f"/tmp/{agent_name}/models/functiongemma-sre-finetuned",
        learning_rate=learning_rate,
        batch_size=batch_size,
        iters=iters,
        lora_layers=lora_layers,
        lora_rank=lora_rank,
        lora_alpha=lora_alpha,
    )
    
    # Step 4: Evaluate model
    metrics = evaluate_model(
        model_path=mlx_model_path,
        adapter_path=training_result["adapter_path"],
        test_data=dataset["test"],
    )
    
    # Step 5: Register in MLflow
    mlflow_uri = register_model_mlflow(
        model_path=mlx_model_path,
        adapter_path=training_result["adapter_path"],
        metrics=metrics,
        agent_name=agent_name,
        runbook_version="latest",  # Could extract from Git
    )
    
    # Step 6: Store in MinIO
    model_version = f"v{datetime.now().strftime('%Y%m%d')}-{current_context().execution_id.name[:8]}"
    minio_path = store_model_minio(
        adapter_path=training_result["adapter_path"],
        agent_name=agent_name,
        model_version=model_version,
    )
    
    # Step 7: Export to Ollama format
    ollama_result = export_model_ollama(
        base_model_path=mlx_model_path,
        adapter_path=training_result["adapter_path"],
        output_dir=f"/tmp/{agent_name}/models",
        agent_name=agent_name,
        model_version=model_version,
    )
    
    # Step 8: Upload Ollama model to MinIO for persistence
    ollama_minio_path = store_model_minio(
        adapter_path=ollama_result["ollama_output_path"],
        agent_name=agent_name,
        model_version=f"{model_version}-ollama",
    )
    
    # Step 9: Automatically trigger Ollama import job
    import_job_result = trigger_ollama_import(
        model_name=ollama_result["model_name"],
        model_version=model_version,
        minio_path=ollama_minio_path,
    )
    
    return {
        "model_version": model_version,
        "mlflow_uri": mlflow_uri,
        "minio_path": minio_path,  # Adapters in MinIO
        "adapter_path": training_result["adapter_path"],
        "accuracy": str(metrics["accuracy"]),
        "ollama_model_name": ollama_result["model_name"],  # Format: agent-sre:v20251224-abc12345
        "ollama_output_path": ollama_result["ollama_output_path"],
        "ollama_modelfile_path": ollama_result["modelfile_path"],
        "ollama_minio_path": ollama_minio_path,  # Ollama model in MinIO
        "ollama_import_job": import_job_result["job_name"],  # Kubernetes job name
    }


if __name__ == "__main__":
    # For local testing
    result = agent_training_pipeline(
        agent_name="agent-sre",
        model_name="google/functiongemma-270m-it",
        runbook_path="https://raw.githubusercontent.com/brunovlucena/homelab/main/flux/ai/agent-sre/docs/RUNBOOK.md",
    )
    print(json.dumps(result, indent=2))

