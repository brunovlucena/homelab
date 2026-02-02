"""Command Center agent main entry point."""
import os
import asyncio

from fastapi import FastAPI, Request, Response
from fastapi.responses import JSONResponse
from cloudevents.http import from_http
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
import structlog

from command_center.handler import CommandCenterHandler
# Import metrics to ensure they're registered with Prometheus
from shared import metrics

# Configure structured logging
structlog.configure(
    processors=[
        structlog.stdlib.filter_by_level,
        structlog.stdlib.add_logger_name,
        structlog.stdlib.add_log_level,
        structlog.stdlib.PositionalArgumentsFormatter(),
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.UnicodeDecoder(),
        structlog.processors.JSONRenderer(),
    ],
    wrapper_class=structlog.stdlib.BoundLogger,
    context_class=dict,
    logger_factory=structlog.stdlib.LoggerFactory(),
    cache_logger_on_first_use=True,
)

logger = structlog.get_logger()

app = FastAPI(
    title="POS Command Center",
    description="Central monitoring hub for POS edge agents",
    version="0.1.0",
)

handler = CommandCenterHandler()


@app.post("/")
async def receive_event(request: Request):
    """Receive and process CloudEvents."""
    try:
        body = await request.body()
        event = from_http(dict(request.headers), body)
        
        result = await handler.handle(event)
        
        return JSONResponse(content=result)
    
    except Exception as e:
        logger.exception("event_processing_error", error=str(e))
        return JSONResponse(
            status_code=500,
            content={"error": str(e)},
        )


@app.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "healthy", "agent": "command-center"}


@app.get("/ready")
async def ready():
    """Readiness check endpoint."""
    return {"status": "ready"}


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST,
    )


@app.get("/api/locations")
async def list_locations():
    """List all known locations."""
    return {"locations": handler.locations}


@app.get("/api/alerts")
async def list_alerts():
    """List active alerts."""
    return {"alerts": handler.alerts}


@app.on_event("startup")
async def startup():
    """Start background tasks."""
    logger.info("command_center_starting")
    
    # Start stale location checker
    async def check_stale_loop():
        while True:
            await asyncio.sleep(60)
            stale = await handler.check_stale_locations()
            if stale:
                logger.warning("stale_locations_detected", locations=stale)
    
    asyncio.create_task(check_stale_loop())


@app.on_event("shutdown")
async def shutdown():
    """Cleanup on shutdown."""
    await handler.emitter.close()
    logger.info("command_center_shutdown")


if __name__ == "__main__":
    import uvicorn
    
    port = int(os.getenv("PORT", "8080"))
    uvicorn.run(app, host="0.0.0.0", port=port)
