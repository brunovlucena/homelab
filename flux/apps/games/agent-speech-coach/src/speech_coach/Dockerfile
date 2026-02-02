FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY speech_coach/ ./speech_coach/
COPY shared/ ./shared/

# Set Python path
ENV PYTHONPATH=/app

# Expose port
EXPOSE 8080

# Run the application
CMD ["python", "-m", "uvicorn", "speech_coach.main:app", "--host", "0.0.0.0", "--port", "8080"]
