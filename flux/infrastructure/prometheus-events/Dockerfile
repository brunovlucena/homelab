FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY src/requirements.txt /app/requirements.txt

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code
COPY src/ /app/src/

# Set Python path
ENV PYTHONPATH=/app

# Expose port
EXPOSE 8080

# Run server
CMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8080"]

