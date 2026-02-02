# 🛡️ Garak - LLM Vulnerability Scanner
# NVIDIA's comprehensive LLM security testing tool

FROM python:3.11-slim

WORKDIR /app

# Install system dependencies including build tools for Rust
RUN apt-get update && apt-get install -y \
    curl \
    git \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Install Rust and Cargo (required for garak's base2048 dependency)
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"

# Install Garak (requires Rust for base2048 dependency compilation)
RUN pip install --no-cache-dir garak

# Install FastAPI for API wrapper
RUN pip install --no-cache-dir fastapi uvicorn[standard] pydantic

# Copy API wrapper
COPY src/ ./src/

# Create non-root user
RUN useradd -m -u 1000 garak && \
    chown -R garak:garak /app

USER garak

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run API wrapper
CMD ["python", "-m", "uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8080"]

