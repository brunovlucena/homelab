FROM python:3.12-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy and install shared-lib (agent optimization)
COPY shared-lib/ /app/shared-lib/
RUN pip install --no-cache-dir -e /app/shared-lib

# Copy source code
COPY . .

# Set Python path
ENV PYTHONPATH=/app
ENV PYTHONUNBUFFERED=1

# Default environment variables
ENV OLLAMA_URL=http://ollama.ai-inference.svc.cluster.local:11434
ENV OLLAMA_MODEL=llama3.2:3b
ENV VERSION=0.1.0

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import httpx; httpx.get('http://localhost:8080/health')" || exit 1

# Run the application
CMD ["uvicorn", "chatbot.main:app", "--host", "0.0.0.0", "--port", "8080"]
