# WhatsApp Gateway Agent Dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy shared modules
COPY shared/ ./shared/

# Copy agent code
COPY whatsapp_gateway/ ./whatsapp_gateway/

# Environment variables
ENV PYTHONPATH=/app
ENV PORT=8080

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["uvicorn", "whatsapp_gateway.main:app", "--host", "0.0.0.0", "--port", "8080"]
