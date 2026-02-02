"""
Alert Dispatcher - FastAPI entry point.

Handles CloudEvents for multi-channel alert delivery.
"""
import os
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request, Response, HTTPException
from fastapi.responses import JSONResponse
from cloudevents.http import from_http, CloudEvent
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
import structlog

from .handler import AlertDispatcher, Alert, Severity, handle_alert_event

logger = structlog.get_logger()

# Global dispatcher
dispatcher: AlertDispatcher = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global dispatcher
    
    dispatcher = AlertDispatcher()
    logger.info("alert_dispatcher_initialized")
    
    yield
    
    logger.info("alert_dispatcher_shutdown")


app = FastAPI(
    title="Alert Dispatcher",
    description="Multi-channel alert delivery for smart contract security findings",
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
    if dispatcher is None:
        raise HTTPException(status_code=503, detail="Dispatcher not initialized")
    
    # Check channel availability
    channels = {
        "telegram": dispatcher.telegram.enabled,
        "discord": dispatcher.discord.enabled,
        "grafana": dispatcher.grafana.enabled,
    }
    
    return {"status": "ready", "channels": channels}


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST
    )


@app.post("/")
async def handle_event(request: Request):
    """
    Handle incoming CloudEvents.
    
    Accepts:
    - io.homelab.alert.sent: Alert to dispatch
    - io.homelab.exploit.validated: Validated exploit (auto-generates alert)
    """
    # Parse CloudEvent
    try:
        event = from_http(dict(request.headers), await request.body())
    except Exception as e:
        logger.error("cloudevent_parse_failed", error=str(e))
        raise HTTPException(status_code=400, detail=f"Invalid CloudEvent: {e}")
    
    log = logger.bind(
        event_type=event["type"],
        event_id=event["id"],
    )
    log.info("event_received")
    
    if event["type"] == "io.homelab.alert.sent":
        results = await handle_alert_event(event)
        return JSONResponse(content={"status": "dispatched", "results": results})
    
    elif event["type"] == "io.homelab.exploit.validated":
        return await handle_exploit_alert(event)
    
    else:
        log.warning("unknown_event_type")
        return JSONResponse(
            status_code=200,
            content={"status": "ignored", "reason": "Unknown event type"}
        )


async def handle_exploit_alert(event: CloudEvent) -> Response:
    """Generate and dispatch alert for validated exploit."""
    data = event.data
    
    # Map exploit status to severity
    severity = Severity.HIGH
    if data.get("vulnerability_type") in ["reentrancy", "arbitrary_call", "delegatecall"]:
        severity = Severity.CRITICAL
    if data.get("profit_potential"):
        severity = Severity.CRITICAL
    
    alert = Alert(
        severity=severity,
        title=f"Validated Exploit: {data.get('vulnerability_type', 'unknown')}",
        chain=data.get("chain", "unknown"),
        contract_address=data.get("contract_address", ""),
        vulnerability_type=data.get("vulnerability_type", "unknown"),
        description=f"Exploit validated on Anvil fork. Status: {data.get('status')}",
        profit_potential=data.get("profit_potential"),
        exploit_validated=data.get("status") == "validated",
    )
    
    results = await dispatcher.dispatch(alert)
    return JSONResponse(content={"status": "dispatched", "results": results})


# Direct API endpoint for testing
@app.post("/alert")
async def send_alert(
    severity: str = "high",
    title: str = "Test Alert",
    chain: str = "ethereum",
    contract_address: str = "0x0000000000000000000000000000000000000000",
    vulnerability_type: str = "test",
    description: str = "",
):
    """Direct API to send an alert (for testing)."""
    alert = Alert(
        severity=Severity(severity),
        title=title,
        chain=chain,
        contract_address=contract_address,
        vulnerability_type=vulnerability_type,
        description=description,
    )
    
    results = await dispatcher.dispatch(alert)
    return {"status": "dispatched", "results": results}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)

