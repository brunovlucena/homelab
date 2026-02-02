FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY shared/ ./shared/
COPY pump_agent/ ./pump_agent/

ENV PYTHONUNBUFFERED=1
EXPOSE 8080

CMD ["python", "-m", "pump_agent.main"]
