"""
Sales Assistant Agent - FastAPI entry point.

Helps human sales representatives with AI-powered suggestions and escalation handling.
"""
import os
from contextlib import asynccontextmanager
from typing import Optional

from fastapi import FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import Response
from pydantic import BaseModel, Field
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
from cloudevents.http import from_http
import structlog

from .handler import SalesAssistant
from shared.metrics import init_metrics
from shared.events import init_events, shutdown_events

logger = structlog.get_logger()

# Global assistant instance
assistant: SalesAssistant = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global assistant
    
    # Initialize CloudEvents
    pub, sub = init_events(source="/agent-store-multibrands/sales-assistant")
    
    # Initialize assistant
    assistant = SalesAssistant(
        event_publisher=pub,
        event_subscriber=sub,
    )
    assistant.setup_event_handlers()
    
    # Initialize metrics
    version = os.getenv("VERSION", "0.1.0")
    commit = os.getenv("GIT_COMMIT", "unknown")
    init_metrics(version, commit, "sales-assistant")
    
    logger.info("sales_assistant_started", version=version)
    
    yield
    
    # Cleanup
    await shutdown_events()
    logger.info("sales_assistant_shutdown")


app = FastAPI(
    title="Sales Assistant",
    description="AI-powered assistant for human sales representatives",
    version="0.1.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# =============================================================================
# Request/Response Models
# =============================================================================

class SuggestionsRequest(BaseModel):
    """Request for AI suggestions."""
    conversation_context: str = Field(..., description="Conversation history")
    customer_message: str = Field(..., description="Latest customer message")
    brand: str = Field(..., description="Brand being discussed")


class ObjectionRequest(BaseModel):
    """Request for objection analysis."""
    objection: str = Field(..., description="Customer's objection")
    product: str = Field(..., description="Product being discussed")
    brand: str = Field(..., description="Brand")


class AcceptEscalationRequest(BaseModel):
    """Request to accept an escalation."""
    conversation_id: str
    seller_id: str


class ResolveEscalationRequest(BaseModel):
    """Request to resolve an escalation."""
    conversation_id: str
    outcome: str = Field(..., description="purchase|no_purchase|transferred|other")


# =============================================================================
# Health Endpoints
# =============================================================================

@app.get("/health")
async def health():
    """Health check endpoint."""
    ollama_ok = await assistant.health_check() if assistant else False
    return {
        "status": "healthy" if ollama_ok else "degraded",
        "ollama_available": ollama_ok,
    }


@app.get("/ready")
async def ready():
    """Readiness check endpoint."""
    if assistant is None:
        raise HTTPException(status_code=503, detail="Assistant not initialized")
    return {"status": "ready"}


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST
    )


# =============================================================================
# Suggestion Endpoints
# =============================================================================

@app.post("/suggestions")
async def get_suggestions(request: SuggestionsRequest):
    """
    Get AI suggestions for responding to a customer.
    
    Used by human sellers during active conversations.
    """
    if assistant is None:
        raise HTTPException(status_code=503, detail="Assistant not initialized")
    
    result = await assistant.get_suggestions(
        conversation_context=request.conversation_context,
        customer_message=request.customer_message,
        brand=request.brand,
    )
    
    return result


@app.post("/objection")
async def analyze_objection(request: ObjectionRequest):
    """
    Analyze customer objection and get response strategies.
    """
    if assistant is None:
        raise HTTPException(status_code=503, detail="Assistant not initialized")
    
    result = await assistant.analyze_objection(
        objection=request.objection,
        product=request.product,
        brand=request.brand,
    )
    
    return result


# =============================================================================
# Escalation Endpoints
# =============================================================================

@app.get("/escalations")
async def get_escalations(
    brand: Optional[str] = None,
    priority: Optional[str] = None,
    limit: int = 10,
):
    """
    Get pending escalations.
    
    Optionally filter by brand or priority.
    """
    if assistant is None:
        raise HTTPException(status_code=503, detail="Assistant not initialized")
    
    pending = assistant.escalation_queue.get_pending(
        brand=brand,
        priority=priority,
        limit=limit,
    )
    
    return {
        "count": len(pending),
        "escalations": pending,
        "queue_status": assistant.get_queue_status(),
    }


@app.post("/escalations/accept")
async def accept_escalation(request: AcceptEscalationRequest):
    """
    Accept an escalation for handling.
    """
    if assistant is None:
        raise HTTPException(status_code=503, detail="Assistant not initialized")
    
    escalation = assistant.accept_escalation(
        conv_id=request.conversation_id,
        seller_id=request.seller_id,
    )
    
    if escalation is None:
        raise HTTPException(status_code=404, detail="Escalation not found")
    
    return {
        "status": "accepted",
        "escalation": escalation,
    }


@app.post("/escalations/resolve")
async def resolve_escalation(request: ResolveEscalationRequest):
    """
    Resolve an escalation.
    """
    if assistant is None:
        raise HTTPException(status_code=503, detail="Assistant not initialized")
    
    success = assistant.resolve_escalation(
        conv_id=request.conversation_id,
        outcome=request.outcome,
    )
    
    if not success:
        raise HTTPException(status_code=404, detail="Active conversation not found")
    
    return {"status": "resolved", "outcome": request.outcome}


@app.get("/escalations/status")
async def escalation_status():
    """Get escalation queue status."""
    if assistant is None:
        raise HTTPException(status_code=503, detail="Assistant not initialized")
    
    return assistant.get_queue_status()


# =============================================================================
# CloudEvents Endpoint
# =============================================================================

@app.post("/events")
async def receive_cloudevent(request: Request):
    """
    Receive CloudEvents from Knative triggers.
    
    Subscribed events:
    - store.sales.escalate: Escalation from AI seller
    - store.chat.message.new: New message in active conversation
    """
    try:
        headers = dict(request.headers)
        body = await request.body()
        event = from_http(headers, body)
        
        logger.info(
            "cloudevent_received",
            event_type=event["type"],
            source=event["source"],
        )
        
        if assistant and assistant.event_subscriber:
            handled = await assistant.event_subscriber.handle(event)
            return {"status": "handled" if handled else "no_handler"}
        
        return {"status": "subscriber_not_initialized"}
        
    except Exception as e:
        logger.error("cloudevent_processing_failed", error=str(e))
        raise HTTPException(status_code=400, detail=f"Invalid CloudEvent: {str(e)}")


# =============================================================================
# Root Endpoint
# =============================================================================

@app.get("/")
async def root():
    """Root endpoint with service info."""
    return {
        "service": "sales-assistant",
        "description": "AI-powered assistant for human sales representatives",
        "version": os.getenv("VERSION", "0.1.0"),
        "endpoints": {
            "suggestions": "POST /suggestions",
            "objection": "POST /objection",
            "escalations": "GET /escalations",
            "accept": "POST /escalations/accept",
            "resolve": "POST /escalations/resolve",
            "status": "GET /escalations/status",
            "events": "POST /events (CloudEvents)",
            "health": "GET /health",
            "ready": "GET /ready",
            "metrics": "GET /metrics",
        },
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
