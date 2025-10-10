# 🤖 Jamie Slack Bot Dockerfile
FROM python:3.12-slim

WORKDIR /app

# Copy requirements from parent directory
COPY requirements.txt ./requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

# Copy core module and slack bot code
COPY src/slack-bot/core.py ./
COPY src/slack-bot/jamie_slack_bot.py ./

# Run as non-root user
RUN useradd -m -u 1000 jamie && chown -R jamie:jamie /app
USER jamie

# Health check (basic Python execution)
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD python -c "import sys; sys.exit(0)"

# Start Slack bot
CMD ["python", "-u", "jamie_slack_bot.py"]
