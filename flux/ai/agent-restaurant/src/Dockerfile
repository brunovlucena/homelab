# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🍽️ Restaurant Agent - Generic Dockerfile
#
# This Dockerfile builds a generic restaurant agent that can be configured
# via environment variables to act as different restaurant staff.
#
# Build args:
#   AGENT_NAME: Name of the agent (e.g., Pierre, Maximilian)
#   AGENT_ROLE: Role of the agent (waiter, host, sommelier, chef)
#
# Example:
#   docker build --build-arg AGENT_NAME=Pierre --build-arg AGENT_ROLE=waiter -t waiter-agent .
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
FROM python:3.12-slim

# Build arguments
ARG AGENT_NAME=Agent
ARG AGENT_ROLE=assistant

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy and install shared-lib (agent optimization)
COPY shared-lib/ /app/shared-lib/
RUN pip install --no-cache-dir -e /app/shared-lib

# Copy source code
COPY restaurant_agent/ ./restaurant_agent/

# Set Python path
ENV PYTHONPATH=/app
ENV PYTHONUNBUFFERED=1

# Agent configuration (can be overridden at runtime)
ENV AGENT_NAME=${AGENT_NAME}
ENV AGENT_ROLE=${AGENT_ROLE}
ENV OLLAMA_URL=http://ollama-native.ollama.svc.cluster.local:11434
ENV OLLAMA_MODEL=llama3.2:3b

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import httpx; httpx.get('http://localhost:8080/health')" || exit 1

# Run the application
CMD ["uvicorn", "restaurant_agent.main:app", "--host", "0.0.0.0", "--port", "8080"]
