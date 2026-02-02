FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy and install shared-lib (agent optimization)
COPY shared-lib/ /app/shared-lib/
RUN pip install --no-cache-dir -e /app/shared-lib

COPY k8s_tools/ .

ENV PYTHONUNBUFFERED=1
EXPOSE 8080

CMD ["python", "main.py"]
