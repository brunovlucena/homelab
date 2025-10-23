# 🐍 Python 3.11 slim base image
FROM python:3.11-slim

# 📋 Set metadata
LABEL maintainer="bruno"
LABEL description="Serverless LanceDB with MinIO and Knative"

# 🔧 Set working directory
WORKDIR /app

# 📦 Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# 📥 Copy requirements first (for better layer caching)
COPY requirements.txt .

# 🔨 Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# 📁 Copy application code
COPY app.py .

# 👤 Create non-root user
RUN useradd -m -u 1000 appuser && \
    chown -R appuser:appuser /app

USER appuser

# 🌐 Expose port
EXPOSE 8080

# 🔥 Set environment variables
ENV PORT=8080 \
    PYTHONUNBUFFERED=1 \
    LOG_LEVEL=INFO

# 🚀 Run the application
CMD ["python", "app.py"]

