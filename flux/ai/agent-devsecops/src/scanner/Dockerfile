# Agent DevSecOps - Security Scanner
FROM python:3.12-slim

# Security: Run as non-root (use nobody user)
RUN useradd -r -s /bin/false scanner || true

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code
COPY *.py ./

# Security: Set ownership and make read-only
RUN chown -R scanner:scanner /app

USER scanner

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8080/health')" || exit 1

ENV PORT=8080
EXPOSE 8080

CMD ["python", "main.py"]
