# 🤖 Jamie MCP Server Dockerfile
FROM python:3.12-slim

WORKDIR /app

# Copy requirements from parent directory
COPY requirements.txt ./requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

# Copy core module and MCP server code
COPY src/mcp-server/core.py ./
COPY src/mcp-server/mcp_server.py ./

# Expose MCP port
EXPOSE 30121

# Run as non-root user
RUN useradd -m -u 1000 jamie && chown -R jamie:jamie /app
USER jamie

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:30121/health')"

# Start MCP server
CMD ["python", "-u", "mcp_server.py"]
