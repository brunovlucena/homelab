# Dockerfile for Flyte Sandbox Training Pipeline
# Used by Flyte workflows on Forge cluster

FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Upgrade pip first to avoid resolution issues
RUN pip install --upgrade pip setuptools wheel

# Install dependencies in stages to avoid timeout/resolution issues
# Stage 1: Core dependencies (smallest first)
RUN pip install --no-cache-dir \
    boto3>=1.34.0 \
    minio>=7.2.0

# Stage 2: MLflow
RUN pip install --no-cache-dir mlflow>=2.10.0

# Stage 3: MLX and MLX-LM (Apple Silicon optimized)
# These may take longer to install due to compilation
RUN pip install --no-cache-dir \
    mlx>=0.18.0 \
    mlx-lm>=0.1.0

# Stage 4: Flytekit and HuggingFace
RUN pip install --no-cache-dir \
    flytekit>=1.10.0 \
    huggingface_hub>=0.20.0

# Set working directory
WORKDIR /app

# Copy workflow and scripts
COPY workflows/ /app/workflows/
COPY scripts/ /app/scripts/
COPY config/ /app/config/

# Set Python path
ENV PYTHONPATH=/app

# Default command (Flyte will override)
CMD ["python", "-m", "workflows.agent_training"]

