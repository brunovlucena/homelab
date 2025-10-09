FROM python:3.11-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application
COPY jamie_slack_bot.py .

# Run as non-root user
RUN useradd -m -u 1000 jamie && \
    chown -R jamie:jamie /app

USER jamie

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD python -c "import sys; sys.exit(0)"

CMD ["python", "-u", "jamie_slack_bot.py"]

