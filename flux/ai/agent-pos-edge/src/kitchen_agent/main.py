"""Kitchen Agent main entry point."""
import os

from fastapi import FastAPI, Request, Response
from fastapi.responses import JSONResponse
from cloudevents.http import from_http
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
import structlog

from kitchen_agent.handler import KitchenAgentHandler
# Import metrics to ensure they're registered with Prometheus
from shared import metrics

structlog.configure(
    processors=[
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.JSONRenderer(),
    ],
)

logger = structlog.get_logger()

app = FastAPI(
    title="Kitchen Agent",
    description="Kitchen operations monitoring for fast-food",
    version="0.1.0",
)

handler = KitchenAgentHandler()


@app.post("/")
async def receive_event(request: Request):
    """Receive CloudEvents."""
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
    return {
        "status": "healthy",
        "agent": "kitchen",
        "location_id": handler.location_id,
        "queue_depth": len(handler.order_queue),
    }


@app.get("/ready")
async def ready():
    return {"status": "ready"}


@app.get("/metrics")
async def metrics_endpoint():
    return Response(content=generate_latest(), media_type=CONTENT_TYPE_LATEST)


@app.get("/api/queue")
async def get_queue():
    """Get current order queue."""
    return {
        "queue_depth": len(handler.order_queue),
        "orders": [
            {
                "order_id": oid,
                "items": handler.orders[oid].items if oid in handler.orders else [],
                "received_at": handler.orders[oid].received_at.isoformat() if oid in handler.orders else None,
            }
            for oid in handler.order_queue
        ],
    }


@app.post("/api/order/{order_id}/start")
async def start_order(order_id: str, station: str = "grill"):
    """Start preparing an order."""
    success = await handler.start_order(order_id, station)
    return {"success": success, "order_id": order_id}


@app.post("/api/order/{order_id}/complete")
async def complete_order(order_id: str):
    """Mark order as ready."""
    success = await handler.complete_order(order_id)
    return {"success": success, "order_id": order_id}


@app.on_event("startup")
async def startup():
    logger.info("kitchen_agent_starting", location_id=handler.location_id)


@app.on_event("shutdown")
async def shutdown():
    await handler.emitter.close()


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=int(os.getenv("PORT", "8080")))
