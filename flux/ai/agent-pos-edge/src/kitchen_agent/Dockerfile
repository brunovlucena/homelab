FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY shared/ ./shared/
COPY kitchen_agent/ ./kitchen_agent/

ENV PYTHONUNBUFFERED=1
EXPOSE 8080

CMD ["python", "-m", "kitchen_agent.main"]
