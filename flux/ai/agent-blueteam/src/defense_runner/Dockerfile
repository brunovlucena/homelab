FROM python:3.11-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy and install shared-lib (agent optimization)
COPY shared-lib/ /app/shared-lib/
RUN pip install --no-cache-dir -e /app/shared-lib

# Copy source code
COPY defense_runner/ ./defense_runner/
COPY shared/ ./shared/

# Add /app to Python path so shared module can be found
ENV PYTHONPATH="/app:${PYTHONPATH}"

# Set working directory to handler
WORKDIR /app/defense_runner

# Run the service
CMD ["python", "main.py"]
