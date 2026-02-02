# Alert Dispatcher - Lightweight notification service
FROM python:3.12-slim

WORKDIR /app

# Install Python dependencies (minimal set for alerting)
COPY requirements.txt .
RUN pip install --no-cache-dir \
    fastapi \
    uvicorn \
    httpx \
    cloudevents \
    prometheus-client \
    structlog \
    pydantic

# Copy application
COPY alert_dispatcher/ ./alert_dispatcher/

# Create non-root user
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD python -c "import httpx; httpx.get('http://localhost:8080/health').raise_for_status()"

# Run
CMD ["python", "-m", "uvicorn", "alert_dispatcher.main:app", "--host", "0.0.0.0", "--port", "8080"]

