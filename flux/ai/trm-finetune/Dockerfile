FROM python:3.10-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Install CUDA toolkit (if needed for GPU training)
# RUN apt-get update && apt-get install -y \
#     cuda-toolkit-12-6 \
#     && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /workspace

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

# Install PyTorch with CUDA support (adjust for your CUDA version)
RUN pip install --pre --upgrade torch torchvision torchaudio \
    --index-url https://download.pytorch.org/whl/nightly/cu126

# Clone TRM repository
RUN git clone https://github.com/SamsungSAILMontreal/TinyRecursiveModels.git /workspace/TinyRecursiveModels

# Install TRM dependencies
RUN cd /workspace/TinyRecursiveModels && \
    pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY src/ /workspace/src/

# Set Python path
ENV PYTHONPATH=/workspace:/workspace/src:$PYTHONPATH

# Default command
CMD ["python", "-m", "src.flyte_workflow"]

