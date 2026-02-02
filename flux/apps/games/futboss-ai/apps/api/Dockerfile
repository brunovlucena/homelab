FROM python:3.12-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code
COPY app/ ./app/

# Set Python path
ENV PYTHONPATH=/app
ENV PYTHONUNBUFFERED=1

# Default environment variables (can be overridden via ConfigMap/Secrets)
ENV API_HOST=0.0.0.0
ENV API_PORT=8000
ENV DEBUG=false
ENV MONGODB_URL=mongodb://futboss-mongodb.futboss-ai.svc.cluster.local:27017
ENV MONGODB_DATABASE=futboss
ENV OLLAMA_BASE_URL=http://ollama-native.ollama.svc.cluster.local:11434
ENV OLLAMA_MODEL=llama3.2:3b

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import httpx; httpx.get('http://localhost:8000/health')" || exit 1

# Run the application
CMD ["uvicorn", "app.main:socket_app", "--host", "0.0.0.0", "--port", "8000"]
