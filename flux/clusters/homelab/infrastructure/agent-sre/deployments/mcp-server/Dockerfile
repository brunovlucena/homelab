FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install uv
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /usr/local/bin/

# Copy uv configuration files
COPY pyproject.toml uv.lock ./

# Copy MCP server specific code
COPY deployments/mcp-server/mcp_server.py ./
COPY deployments/mcp-server/core.py ./

# Install Python dependencies with uv
RUN uv sync --frozen --no-dev

# Create non-root user
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 30120

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:30120/health || exit 1

# Run the MCP server using uv
CMD ["uv", "run", "python", "mcp_server.py"]
