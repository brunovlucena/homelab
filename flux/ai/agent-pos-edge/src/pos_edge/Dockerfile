FROM python:3.12-slim

WORKDIR /app

# Install system dependencies for psutil
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt psutil

# Copy and install shared-lib (agent optimization)
COPY shared-lib/ /app/shared-lib/
RUN pip install --no-cache-dir -e /app/shared-lib

# Copy source code
COPY shared/ ./shared/
COPY pos_edge/ ./pos_edge/

ENV PYTHONUNBUFFERED=1
ENV PYTHONDONTWRITEBYTECODE=1

EXPOSE 8080

CMD ["python", "-m", "pos_edge.main"]
