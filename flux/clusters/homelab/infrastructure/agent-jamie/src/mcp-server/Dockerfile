FROM python:3.11-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application files
COPY src/mcp-server/core.py .
COPY src/mcp-server/jamie_mcp_server.py .

# Environment variables
ENV JAMIE_SLACK_BOT_URL="http://jamie-slack-bot-service.jamie.svc.cluster.local:8080"
ENV PYTHONUNBUFFERED=1

# Run the MCP server
CMD ["python", "jamie_mcp_server.py"]

