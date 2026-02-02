"""
WhatsApp Gateway - FastAPI entry point.

Handles WhatsApp Business API webhooks and message routing.
"""
import os
from contextlib import asynccontextmanager
from typing import Optional

from fastapi import FastAPI, HTTPException, Request, Query
from fastapi.responses import Response, PlainTextResponse
from pydantic import BaseModel, Field
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
from cloudevents.http import from_http
import structlog

from .handler import WhatsAppGateway
from shared.metrics import init_metrics
from shared.events import init_events, shutdown_events

logger = structlog.get_logger()

# Global gateway instance
gateway: WhatsAppGateway = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global gateway
    
    # Initialize CloudEvents
    pub, sub = init_events(source="/agent-store-multibrands/whatsapp-gateway")
    
    # Initialize gateway
    gateway = WhatsAppGateway(
        event_publisher=pub,
        event_subscriber=sub,
    )
    gateway.setup_event_handlers()
    
    # Initialize metrics
    version = os.getenv("VERSION", "0.1.0")
    commit = os.getenv("GIT_COMMIT", "unknown")
    init_metrics(version, commit, "whatsapp-gateway")
    
    logger.info("whatsapp_gateway_started", version=version)
    
    yield
    
    # Cleanup
    await shutdown_events()
    logger.info("whatsapp_gateway_shutdown")


app = FastAPI(
    title="WhatsApp Gateway",
    description="WhatsApp Business API integration for Agent Store MultiBrands",
    version="0.1.0",
    lifespan=lifespan,
)


# =============================================================================
# Health Endpoints
# =============================================================================

@app.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "healthy", "service": "whatsapp-gateway"}


@app.get("/ready")
async def ready():
    """Readiness check endpoint."""
    if gateway is None:
        raise HTTPException(status_code=503, detail="Gateway not initialized")
    return {"status": "ready"}


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST
    )


# =============================================================================
# WhatsApp Webhook Endpoints
# =============================================================================

@app.get("/webhook")
async def verify_webhook(
    mode: str = Query(None, alias="hub.mode"),
    token: str = Query(None, alias="hub.verify_token"),
    challenge: str = Query(None, alias="hub.challenge"),
):
    """
    Webhook verification endpoint for Meta.
    
    Meta sends a GET request with these parameters during webhook setup.
    """
    if not all([mode, token, challenge]):
        raise HTTPException(status_code=400, detail="Missing parameters")
    
    result = gateway.verify_webhook(mode, token, challenge)
    
    if result:
        return PlainTextResponse(content=result)
    else:
        raise HTTPException(status_code=403, detail="Verification failed")


@app.post("/webhook")
async def receive_webhook(request: Request):
    """
    Webhook endpoint for incoming WhatsApp messages.
    
    Meta sends POST requests when messages arrive.
    """
    # Verify signature
    signature = request.headers.get("X-Hub-Signature-256", "")
    body = await request.body()
    
    if not gateway.verify_signature(body, signature):
        logger.warning("invalid_webhook_signature")
        raise HTTPException(status_code=401, detail="Invalid signature")
    
    # Parse payload
    try:
        payload = await request.json()
    except Exception as e:
        logger.error("webhook_parse_failed", error=str(e))
        raise HTTPException(status_code=400, detail="Invalid JSON")
    
    # Process messages
    messages = await gateway.process_webhook(payload)
    
    # Always return 200 to acknowledge receipt
    return {
        "status": "ok",
        "messages_processed": len(messages),
    }


# =============================================================================
# CloudEvents Endpoint
# =============================================================================

@app.post("/events")
async def receive_cloudevent(request: Request):
    """
    Receive CloudEvents from Knative triggers.
    
    Subscribed events:
    - store.chat.response: AI seller response to send to customer
    - store.order.status.update: Order status changes to notify customer
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
        
        if gateway and gateway.event_subscriber:
            handled = await gateway.event_subscriber.handle(event)
            return {"status": "handled" if handled else "no_handler"}
        
        return {"status": "subscriber_not_initialized"}
        
    except Exception as e:
        logger.error("cloudevent_processing_failed", error=str(e))
        raise HTTPException(status_code=400, detail=f"Invalid CloudEvent: {str(e)}")


# =============================================================================
# Manual Send Endpoints (for testing)
# =============================================================================

class SendMessageRequest(BaseModel):
    """Request to send a message."""
    phone: str = Field(..., description="Recipient phone number")
    message: str = Field(..., description="Message text")
    reply_to: Optional[str] = Field(None, description="Message ID to reply to")


@app.post("/send")
async def send_message(request: SendMessageRequest):
    """
    Manually send a WhatsApp message.
    
    For testing and admin purposes.
    """
    if gateway is None:
        raise HTTPException(status_code=503, detail="Gateway not initialized")
    
    msg_id = await gateway.send_text_message(
        to=request.phone,
        text=request.message,
        reply_to=request.reply_to,
    )
    
    if msg_id:
        return {"status": "sent", "message_id": msg_id}
    else:
        raise HTTPException(status_code=500, detail="Failed to send message")


class SendButtonsRequest(BaseModel):
    """Request to send interactive buttons."""
    phone: str
    body_text: str
    buttons: list[dict]  # [{"id": "btn1", "title": "Button 1"}]
    header: Optional[str] = None
    footer: Optional[str] = None


@app.post("/send/buttons")
async def send_buttons(request: SendButtonsRequest):
    """Send interactive button message."""
    if gateway is None:
        raise HTTPException(status_code=503, detail="Gateway not initialized")
    
    msg_id = await gateway.send_interactive_buttons(
        to=request.phone,
        body_text=request.body_text,
        buttons=request.buttons,
        header=request.header,
        footer=request.footer,
    )
    
    if msg_id:
        return {"status": "sent", "message_id": msg_id}
    else:
        raise HTTPException(status_code=500, detail="Failed to send message")


# =============================================================================
# Root Endpoint
# =============================================================================

@app.get("/")
async def root():
    """Root endpoint with service info."""
    return {
        "service": "whatsapp-gateway",
        "description": "WhatsApp Business API Gateway for Multi-Brand Store",
        "version": os.getenv("VERSION", "0.1.0"),
        "endpoints": {
            "webhook": "GET/POST /webhook",
            "events": "POST /events (CloudEvents)",
            "send": "POST /send",
            "send_buttons": "POST /send/buttons",
            "health": "GET /health",
            "ready": "GET /ready",
            "metrics": "GET /metrics",
        },
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
