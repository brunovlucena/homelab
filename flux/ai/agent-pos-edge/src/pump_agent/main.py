"""Pump Agent main entry point."""
import os

from fastapi import FastAPI, Request, Response
from fastapi.responses import JSONResponse
from cloudevents.http import from_http
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
import structlog

from pump_agent.handler import PumpAgentHandler
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
    title="Pump Agent",
    description="Fuel pump and tank monitoring for gas stations",
    version="0.1.0",
)

handler = PumpAgentHandler()


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
        "agent": "pump",
        "location_id": handler.location_id,
        "pumps_count": len(handler.pumps),
        "tanks_count": len(handler.tanks),
    }


@app.get("/ready")
async def ready():
    return {"status": "ready"}


@app.get("/metrics")
async def metrics_endpoint():
    return Response(content=generate_latest(), media_type=CONTENT_TYPE_LATEST)


# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# API Endpoints
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

@app.get("/api/pumps")
async def list_pumps():
    """List all pumps."""
    return {
        "pumps": [
            {
                "pump_id": p.pump_id,
                "status": p.status.value,
                "fuel_types": p.fuel_types,
                "has_transaction": p.current_transaction is not None,
            }
            for p in handler.pumps.values()
        ]
    }


@app.get("/api/tanks")
async def list_tanks():
    """List all tanks."""
    return {
        "tanks": [
            {
                "tank_id": t.tank_id,
                "fuel_type": t.fuel_type,
                "level_percent": t.level_percent,
                "current_level": t.current_level,
                "capacity": t.capacity,
            }
            for t in handler.tanks.values()
        ]
    }


@app.get("/api/tanks/{tank_id}/prediction")
async def get_tank_prediction(tank_id: str):
    """Get refill prediction for tank."""
    prediction = handler.predict_refill(tank_id)
    if prediction:
        return prediction
    return {"error": "Not enough data for prediction"}


@app.post("/api/simulate/pump/{pump_id}/start")
async def simulate_pump_start(pump_id: str, fuel_type: str = "regular"):
    """Simulate starting pump."""
    success = await handler.start_pumping(pump_id, fuel_type)
    return {"success": success, "pump_id": pump_id}


@app.post("/api/simulate/pump/{pump_id}/end")
async def simulate_pump_end(pump_id: str, liters: float = 30.0):
    """Simulate ending pump transaction."""
    if pump_id in handler.active_transactions:
        handler.active_transactions[pump_id].liters = liters
    
    txn = await handler.end_pumping(pump_id)
    if txn:
        return {
            "success": True,
            "liters": txn.liters,
            "total": txn.total,
        }
    return {"success": False}


@app.post("/api/simulate/tank/{tank_id}")
async def simulate_tank_reading(
    tank_id: str,
    level: float = 10000,
    capacity: float = 20000,
    fuel_type: str = "regular",
):
    """Simulate tank sensor reading."""
    await handler._handle_tank_sensor({
        "tank_id": tank_id,
        "level": level,
        "capacity": capacity,
        "fuel_type": fuel_type,
    })
    return {"success": True, "tank_id": tank_id}


@app.on_event("startup")
async def startup():
    logger.info("pump_agent_starting", location_id=handler.location_id)


@app.on_event("shutdown")
async def shutdown():
    await handler.emitter.close()


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=int(os.getenv("PORT", "8080")))
