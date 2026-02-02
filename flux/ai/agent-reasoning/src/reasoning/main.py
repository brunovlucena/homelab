"""
Agent-Reasoning - FastAPI entry point.

Provides recursive reasoning capabilities using TinyRecursiveModels (TRM).
"""
import os
from contextlib import asynccontextmanager
from typing import Optional

from fastapi import FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import Response
from pydantic import BaseModel
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
from cloudevents.http import from_http
import structlog

from .handler import TRMHandler
from shared.types import ReasoningRequest, ReasoningResponse, HealthResponse
from shared.metrics import init_build_info, init_metrics

logger = structlog.get_logger()

# Global handler instance
trm_handler: TRMHandler = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global trm_handler
    
    # Initialize TRM handler
    trm_handler = TRMHandler(
        model_path=os.getenv("MODEL_PATH", "/models/trm-checkpoint.pth"),
        device=os.getenv("DEVICE", "cuda"),
        h_cycles=int(os.getenv("H_CYCLES", "3")),
        l_cycles=int(os.getenv("L_CYCLES", "6")),
    )
    
    # Load model (async)
    try:
        await trm_handler.load_model()
    except Exception as e:
        logger.error("model_load_failed", error=str(e))
        # Continue without model for health checks
    
    # Initialize build info
    version = os.getenv("VERSION", "0.1.0")
    commit = os.getenv("GIT_COMMIT", "unknown")
    init_build_info(version, commit)
    init_metrics()
    
    logger.info(
        "agent_reasoning_initialized",
        version=version,
        model_loaded=trm_handler._model_loaded,
    )
    
    yield
    
    logger.info("agent_reasoning_shutdown")


app = FastAPI(
    title="Agent-Reasoning",
    description="Recursive reasoning service using TinyRecursiveModels",
    version="0.1.0",
    lifespan=lifespan,
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=os.getenv("CORS_ORIGINS", "*").split(","),
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# =============================================================================
# Endpoints
# =============================================================================

@app.get("/health", response_model=HealthResponse)
async def health():
    """Health check endpoint."""
    if trm_handler is None:
        return HealthResponse(
            status="degraded",
            model_loaded=False,
            device="unknown",
            gpu_available=False,
        )
    
    device_info = trm_handler.get_device_info()
    model_ok = await trm_handler.health_check()
    
    return HealthResponse(
        status="healthy" if model_ok else "degraded",
        model_loaded=device_info["model_loaded"],
        device=device_info["device"],
        gpu_available=device_info["gpu_available"],
    )


@app.get("/ready")
async def ready():
    """Readiness check endpoint."""
    if trm_handler is None:
        raise HTTPException(status_code=503, detail="Handler not initialized")
    
    if not await trm_handler.health_check():
        raise HTTPException(status_code=503, detail="Model not loaded")
    
    return {"status": "ready"}


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST
    )


@app.post("/reason", response_model=ReasoningResponse)
async def reason(request: ReasoningRequest):
    """
    Process a reasoning task with TRM.
    
    This endpoint performs recursive reasoning on the given question,
    iteratively improving the answer over multiple steps.
    """
    if trm_handler is None:
        raise HTTPException(status_code=503, detail="Handler not initialized")
    
    logger.info(
        "reasoning_request_received",
        question_length=len(request.question),
        task_type=request.task_type.value,
        max_steps=request.max_steps,
    )
    
    try:
        result = await trm_handler.reason(request)
        return result
    except Exception as e:
        logger.error("reasoning_request_failed", error=str(e))
        raise HTTPException(status_code=500, detail=f"Reasoning failed: {str(e)}")


@app.post("/events")
async def receive_cloudevent(request: Request):
    """
    Receive CloudEvents for reasoning tasks.
    
    Event types:
    - io.homelab.reasoning.requested: Process reasoning task
    """
    try:
        headers = dict(request.headers)
        body = await request.body()
        event = from_http(headers, body)
        
        event_type = event["type"]
        event_source = event["source"]
        
        logger.info(
            "cloudevent_received",
            event_type=event_type,
            source=event_source,
        )
        
        # Handle reasoning.requested events
        if event_type == "io.homelab.reasoning.requested":
            data = event.get("data", {})
            
            # Convert to ReasoningRequest
            reasoning_request = ReasoningRequest(
                question=data.get("question", ""),
                context=data.get("context", {}),
                max_steps=data.get("max_steps", 6),
                task_type=data.get("task_type", "general"),
                conversation_id=data.get("conversation_id"),
            )
            
            # Process reasoning
            result = await trm_handler.reason(reasoning_request)
            
            return {
                "status": "completed",
                "event_type": event_type,
                "result": result.dict(),
            }
        
        return {"status": "no_handler", "event_type": event_type}
        
    except Exception as e:
        logger.error("cloudevent_processing_failed", error=str(e))
        raise HTTPException(status_code=400, detail=f"Invalid CloudEvent: {str(e)}")


@app.get("/")
async def root():
    """Root endpoint with service info."""
    return {
        "service": "agent-reasoning",
        "description": "Recursive reasoning service using TinyRecursiveModels",
        "version": os.getenv("VERSION", "0.1.0"),
        "endpoints": {
            "reason": "POST /reason",
            "events": "POST /events (CloudEvents)",
            "health": "GET /health",
            "ready": "GET /ready",
            "metrics": "GET /metrics",
        }
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)

