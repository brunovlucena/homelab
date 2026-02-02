"""POS Edge agent main entry point."""
import os
import asyncio

from fastapi import FastAPI, Request, Response
from fastapi.responses import JSONResponse
from cloudevents.http import from_http
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
import structlog

from pos_edge.handler import POSEdgeHandler
from shared.types import HealthMetrics, AlertSeverity
# Import metrics to ensure they're registered with Prometheus
from shared import metrics
from shared.metrics import init_build_info

# Configure structured logging
structlog.configure(
    processors=[
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.JSONRenderer(),
    ],
)

logger = structlog.get_logger()

app = FastAPI(
    title="POS Edge Agent",
    description="Lightweight edge agent for POS terminals",
    version="0.1.0",
)

# Initialize build info for Agent Versions dashboard
VERSION = os.getenv("VERSION", "0.1.0")
GIT_COMMIT = os.getenv("GIT_COMMIT", "unknown")
init_build_info(VERSION, GIT_COMMIT)

handler = POSEdgeHandler()

logger.info("agent_pos_edge_initialized", version=VERSION)


@app.post("/")
async def receive_event(request: Request):
    """Receive CloudEvents from command center."""
    try:
        body = await request.body()
        event = from_http(dict(request.headers), body)
        
        result = await handler.handle(event)
        return JSONResponse(content=result)
    
    except Exception as e:
        logger.exception("event_processing_error", error=str(e))
        return JSONResponse(status_code=500, content={"error": str(e)})


@app.get("/health")
async def health():
    """Health check."""
    return {
        "status": "healthy",
        "agent": "pos-edge",
        "location_id": handler.location_id,
        "pos_id": handler.pos_id,
        "is_online": handler.is_online,
        "buffer_size": len(handler.offline_buffer),
    }


@app.get("/ready")
async def ready():
    """Readiness check."""
    return {"status": "ready"}


@app.get("/metrics")
async def metrics_endpoint():
    """Prometheus metrics."""
    return Response(content=generate_latest(), media_type=CONTENT_TYPE_LATEST)


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Simulation endpoints for testing
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

@app.post("/api/simulate/transaction")
async def simulate_transaction(request: Request):
    """Simulate a transaction for testing."""
    from shared.types import TransactionItem
    import uuid
    
    data = await request.json()
    
    # Start transaction
    items = [
        TransactionItem(**item) for item in data.get("items", [])
    ]
    txn_id = data.get("transaction_id", str(uuid.uuid4()))
    
    await handler.start_transaction(txn_id, items)
    
    # Complete or fail based on request
    if data.get("fail"):
        await handler.fail_transaction(data.get("error", "Simulated failure"))
        return {"status": "failed", "transaction_id": txn_id}
    
    await handler.complete_transaction(
        total=data.get("total", 0),
        payment_type=data.get("payment_type", "card"),
    )
    
    return {"status": "completed", "transaction_id": txn_id}


@app.post("/api/simulate/alert")
async def simulate_alert(request: Request):
    """Simulate an alert for testing."""
    data = await request.json()
    
    await handler.raise_alert(
        severity=AlertSeverity(data.get("severity", "medium")),
        alert_type=data.get("alert_type", "test"),
        message=data.get("message", "Test alert"),
    )
    
    return {"status": "alert_sent"}


@app.on_event("startup")
async def startup():
    """Start background tasks."""
    logger.info(
        "pos_edge_starting",
        location_id=handler.location_id,
        pos_id=handler.pos_id,
    )
    
    # Heartbeat loop
    async def heartbeat_loop():
        while True:
            await asyncio.sleep(handler.heartbeat_interval)
            await handler.send_heartbeat()
    
    # Health check loop
    async def health_loop():
        import psutil
        
        while True:
            await asyncio.sleep(handler.health_check_interval)
            
            try:
                health = HealthMetrics(
                    cpu_percent=psutil.cpu_percent(),
                    memory_percent=psutil.virtual_memory().percent,
                    disk_percent=psutil.disk_usage("/").percent,
                    network_up=True,  # Simplified
                )
                await handler.send_health_report(health)
            except Exception as e:
                logger.error("health_check_failed", error=str(e))
    
    # Buffer flush loop
    async def buffer_flush_loop():
        while True:
            await asyncio.sleep(30)
            if handler.offline_buffer and not handler.is_online:
                await handler.flush_buffer()
    
    asyncio.create_task(heartbeat_loop())
    asyncio.create_task(health_loop())
    asyncio.create_task(buffer_flush_loop())


@app.on_event("shutdown")
async def shutdown():
    """Cleanup."""
    await handler.emitter.close()
    logger.info("pos_edge_shutdown")


if __name__ == "__main__":
    import uvicorn
    
    port = int(os.getenv("PORT", "8080"))
    uvicorn.run(app, host="0.0.0.0", port=port)
