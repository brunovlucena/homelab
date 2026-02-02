FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY src/requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY src/ /app/

# Set Python path
ENV PYTHONPATH=/app
ENV PYTHONUNBUFFERED=1

# Default environment variables
ENV MODEL_PATH=/models/trm-checkpoint.pth
ENV DEVICE=cuda
ENV H_CYCLES=3
ENV L_CYCLES=6
ENV VERSION=0.1.0

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD python -c "import httpx; httpx.get('http://localhost:8080/health', timeout=5)" || exit 1

# Run application
CMD ["uvicorn", "reasoning.main:app", "--host", "0.0.0.0", "--port", "8080"]

