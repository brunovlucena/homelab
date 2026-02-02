"""
Notifi Adapter - FastAPI entry point.

Receives Alertmanager webhooks and forwards to notifi-services.
"""
import os
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request, Response, HTTPException
from fastapi.responses import JSONResponse
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
import structlog

from .handler import AlertmanagerAdapter
# Import metrics to ensure they're registered with Prometheus
from shared import metrics

logger = structlog.get_logger()

# Global adapter
adapter: AlertmanagerAdapter = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global adapter
    
    adapter = AlertmanagerAdapter()
    logger.info("notifi_adapter_initialized",
               notifi_enabled=adapter.notifi.enabled)
    
    yield
    
    logger.info("notifi_adapter_shutdown")


app = FastAPI(
    title="Notifi Webhook Adapter",
    description="Transforms Alertmanager webhooks to notifi-services format",
    version="1.0.0",
    lifespan=lifespan,
)


@app.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "healthy"}


@app.get("/ready")
async def ready():
    """Readiness check endpoint."""
    if adapter is None:
        raise HTTPException(status_code=503, detail="Adapter not initialized")
    
    return {
        "status": "ready",
        "notifi_enabled": adapter.notifi.enabled,
    }


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST
    )


@app.post("/webhook/alertmanager")
async def alertmanager_webhook(request: Request):
    """
    Receive Alertmanager webhook and forward to notifi-services.
    
    Expected Alertmanager payload format:
    {
        "receiver": "notifi",
        "status": "firing",
        "alerts": [
            {
                "status": "firing",
                "labels": {"alertname": "...", "severity": "critical", ...},
                "annotations": {"summary": "...", "description": "..."},
                "startsAt": "...",
                "endsAt": "...",
                "fingerprint": "..."
            }
        ],
        "groupLabels": {...},
        "commonLabels": {...},
        "commonAnnotations": {...}
    }
    """
    try:
        payload = await request.json()
    except Exception as e:
        logger.error("json_parse_failed", error=str(e))
        raise HTTPException(status_code=400, detail=f"Invalid JSON: {e}")
    
    result = await adapter.process_webhook(payload)
    return JSONResponse(content=result)


@app.post("/")
async def root_webhook(request: Request):
    """Alternative endpoint for Alertmanager webhook."""
    return await alertmanager_webhook(request)


# Direct test endpoint
@app.post("/test")
async def test_alert(
    severity: str = "high",
    alertname: str = "TestAlert",
    chain: str = "ethereum",
    contract_address: str = "0x0000000000000000000000000000000000000000",
    message: str = "Test alert from notifi-adapter",
):
    """Send a test alert through the adapter."""
    test_payload = {
        "receiver": "test",
        "status": "firing",
        "alerts": [
            {
                "status": "firing",
                "labels": {
                    "alertname": alertname,
                    "severity": severity,
                    "chain": chain,
                    "contract_address": contract_address,
                },
                "annotations": {
                    "summary": alertname,
                    "description": message,
                },
                "startsAt": "2025-01-01T00:00:00Z",
                "fingerprint": "test-fingerprint",
            }
        ],
    }
    
    result = await adapter.process_webhook(test_payload)
    return JSONResponse(content=result)


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)

