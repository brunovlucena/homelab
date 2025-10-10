# 🐳 Agent-SRE MCP Server Dockerfile
FROM python:3.12-slim

# 📦 Set working directory
WORKDIR /app

# 📥 Copy dependency files
COPY pyproject.toml ./

# 🔧 Install uv for fast package management
RUN pip install --no-cache-dir uv

# 📦 Install dependencies using uv
RUN uv pip install --system --no-cache -e .

# 📄 Copy application code
COPY deployments/mcp-server/mcp_server.py ./
COPY deployments/mcp-server/mcp_http_wrapper.py ./

# 🔐 Run as non-root user
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

# 🌐 Expose port for HTTP wrapper
EXPOSE 3000

# 🚀 Run MCP HTTP wrapper (wraps MCP tools in HTTP API)
CMD ["python", "-u", "mcp_http_wrapper.py"]

