# Vulnerability Scanner - Static analysis with Slither + LLM
FROM python:3.12-slim

WORKDIR /app

# Install system dependencies for Slither
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install solc-select and solc
RUN pip install --no-cache-dir solc-select && \
    solc-select install 0.8.20 && \
    solc-select use 0.8.20

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Install Slither
RUN pip install --no-cache-dir slither-analyzer

# Copy and install shared-lib (agent optimization)
COPY shared-lib/ /app/shared-lib/
RUN pip install --no-cache-dir -e /app/shared-lib

# Copy application
COPY shared/ ./shared/
COPY vuln_scanner/ ./vuln_scanner/

# Create non-root user
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD python -c "import httpx; httpx.get('http://localhost:8080/health').raise_for_status()"

# Run
CMD ["python", "-m", "uvicorn", "vuln_scanner.main:app", "--host", "0.0.0.0", "--port", "8080"]

