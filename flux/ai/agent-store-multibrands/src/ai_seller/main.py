"""
AI Seller Agent - FastAPI entry point.

Multi-brand AI sales agent powered by LLM.
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

from .handler import AISeller
from shared.types import Brand
from shared.metrics import init_metrics, init_build_info
from shared.events import init_events, shutdown_events

logger = structlog.get_logger()

# Global seller instance
seller: AISeller = None


def get_brand_from_env() -> Brand:
    """Get brand from environment variable."""
    brand_str = os.getenv("AGENT_BRAND", "fashion").lower()
    try:
        return Brand(brand_str)
    except ValueError:
        logger.warning(f"Unknown brand '{brand_str}', defaulting to fashion")
        return Brand.FASHION


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global seller
    
    # Get brand from env
    brand = get_brand_from_env()
    
    # Initialize CloudEvents
    pub, sub = init_events(source=f"/agent-store-multibrands/ai-seller-{brand.value}")
    
    # Initialize seller
    seller = AISeller(
        brand=brand,
        event_publisher=pub,
        event_subscriber=sub,
    )
    seller.setup_event_handlers()
    
    # Initialize metrics
    version = os.getenv("VERSION", "0.1.0")
    commit = os.getenv("GIT_COMMIT", "unknown")
    init_metrics(version, commit, f"ai-seller-{brand.value}")
    init_build_info(version, commit)  # For Agent Versions dashboard
    
    logger.info(
        "ai_seller_started",
        version=version,
        brand=brand.value,
        brand_display=brand.display_name,
    )
    
    yield
    
    # Cleanup
    await shutdown_events()
    logger.info("ai_seller_shutdown", brand=brand.value)


app = FastAPI(
    title="AI Seller Agent",
    description="Multi-brand AI sales agent for Agent Store MultiBrands",
    version="0.1.0",
    lifespan=lifespan,
)

# CORS for testing
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

class ChatRequest(BaseModel):
    """Chat request from frontend or gateway."""
    message: str = Field(..., min_length=1, max_length=4096)
    customer_id: str = Field(..., min_length=1)
    customer_phone: str = Field(..., min_length=1)
    conversation_id: Optional[str] = None


class ChatResponse(BaseModel):
    """Chat response to frontend or gateway."""
    message: str
    conversation_id: str
    products_mentioned: list[str] = []
    recommendations: list[dict] = []
    actions: list[str] = []
    tokens_used: int = 0
    duration_ms: float = 0.0
    escalated: bool = False


class HealthResponse(BaseModel):
    """Health check response."""
    status: str
    brand: str
    ollama_available: bool = False


# =============================================================================
# Endpoints
# =============================================================================

@app.get("/health", response_model=HealthResponse)
async def health():
    """Health check endpoint."""
    brand = get_brand_from_env()
    ollama_ok = await seller.health_check() if seller else False
    return HealthResponse(
        status="healthy" if ollama_ok else "degraded",
        brand=brand.value,
        ollama_available=ollama_ok,
    )


@app.get("/ready")
async def ready():
    """Readiness check endpoint."""
    if seller is None:
        raise HTTPException(status_code=503, detail="Seller not initialized")
    
    if not await seller.health_check():
        raise HTTPException(status_code=503, detail="Ollama not available")
    
    return {"status": "ready", "brand": seller.brand.value}


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST
    )


@app.post("/chat", response_model=ChatResponse)
async def chat(request: ChatRequest):
    """
    Process a chat message from a customer.
    
    This endpoint can be called directly or via CloudEvents.
    """
    if seller is None:
        raise HTTPException(status_code=503, detail="Seller not initialized")
    
    logger.info(
        "chat_request_received",
        customer_phone=request.customer_phone,
        message_length=len(request.message),
    )
    
    result = await seller.handle_message(
        message=request.message,
        customer_id=request.customer_id,
        customer_phone=request.customer_phone,
        conversation_id=request.conversation_id,
    )
    
    # Ensure we have a valid message (never empty)
    message = result.get("message") or ""
    if not message.strip():
        message = "Olá! Como posso ajudá-lo hoje? 👋"
    
    return ChatResponse(
        message=message,
        conversation_id=result.get("conversation_id") or "",
        products_mentioned=result.get("products_mentioned") or [],
        recommendations=result.get("recommendations") or [],
        actions=result.get("actions") or [],
        tokens_used=result.get("tokens_used") or 0,
        duration_ms=result.get("duration_ms") or 0.0,
        escalated=result.get("escalated") or False,
    )


# =============================================================================
# CloudEvents Endpoint
# =============================================================================

@app.post("/events")
async def receive_cloudevent(request: Request):
    """
    Receive CloudEvents from Knative triggers.
    
    Subscribed events:
    - store.chat.message.new: New customer message to process
    - store.product.query.result: Product search results
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
        
        if seller and seller.event_subscriber:
            handled = await seller.event_subscriber.handle(event)
            return {"status": "handled" if handled else "no_handler"}
        
        return {"status": "subscriber_not_initialized"}
        
    except Exception as e:
        logger.error("cloudevent_processing_failed", error=str(e))
        raise HTTPException(status_code=400, detail=f"Invalid CloudEvent: {str(e)}")


# =============================================================================
# Info Endpoints
# =============================================================================

@app.get("/")
async def root():
    """Root endpoint with service info."""
    brand = get_brand_from_env()
    return {
        "service": f"ai-seller-{brand.value}",
        "brand": brand.value,
        "brand_display": brand.display_name,
        "brand_emoji": brand.emoji,
        "description": f"AI Sales Agent for {brand.display_name}",
        "version": os.getenv("VERSION", "0.1.0"),
        "endpoints": {
            "chat": "POST /chat",
            "events": "POST /events (CloudEvents)",
            "health": "GET /health",
            "ready": "GET /ready",
            "metrics": "GET /metrics",
        },
    }


@app.get("/brand")
async def brand_info():
    """Get brand information."""
    if seller:
        return {
            "brand": seller.brand.value,
            "display_name": seller.brand.display_name,
            "emoji": seller.brand.emoji,
        }
    return {"brand": "unknown"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
