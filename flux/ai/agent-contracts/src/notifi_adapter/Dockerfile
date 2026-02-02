# Notifi Webhook Adapter
# Transforms Alertmanager webhooks to notifi-services format

FROM python:3.12-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir \
    fastapi \
    uvicorn \
    httpx \
    structlog \
    prometheus-client \
    python-dotenv

# Copy shared library and adapter code
COPY shared/ ./shared/
COPY notifi_adapter/ ./notifi_adapter/

# Create non-root user
RUN useradd -m -u 1000 app && chown -R app:app /app
USER app

# Expose port
EXPOSE 8080

# Run with uvicorn
CMD ["uvicorn", "notifi_adapter.main:app", "--host", "0.0.0.0", "--port", "8080"]

