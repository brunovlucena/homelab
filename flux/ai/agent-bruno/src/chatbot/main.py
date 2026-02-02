"""
Agent-Bruno Chatbot - FastAPI entry point.

Provides a conversational AI assistant for the homelab homepage.

CloudEvents Integration:
- Receives events from agent-contracts (vuln.found, exploit.validated)
- Emits events for analytics (chat.message, chat.intent.*)
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

# Shared observability module
from observability import (
    initialize_observability,
    get_tracer,
    get_current_trace_context,
    is_observability_enabled,
)

from .handler import ChatBot
from shared.metrics import init_build_info, init_metrics
from shared.events import (
    init_events,
    shutdown_events,
    publisher as event_publisher,
    subscriber as event_subscriber,
)

logger = structlog.get_logger()

# Global chatbot instance
chatbot: ChatBot = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global chatbot
    
    # Initialize OpenTelemetry observability first (before other components)
    # This sets up tracing to Tempo via Grafana Alloy
    version = os.getenv("VERSION", "0.1.0")
    service_name = os.getenv("OTEL_SERVICE_NAME", "agent-bruno")
    service_namespace = os.getenv("OTEL_SERVICE_NAMESPACE", "agent-bruno")
    
    observability_initialized = initialize_observability(
        service_name=service_name,
        service_namespace=service_namespace,
        service_version=version,
    )
    
    # Initialize CloudEvents publisher and subscriber
    pub, sub = init_events()
    
    # Initialize chatbot with event integration, domain memory, and TRM
    chatbot = ChatBot(
        event_publisher=pub,
        event_subscriber=sub,
        memory_enabled=os.getenv("MEMORY_ENABLED", "true").lower() == "true",
        redis_url=os.getenv("REDIS_URL"),
        postgres_url=os.getenv("POSTGRES_URL"),
        trm_model_name=os.getenv("TRM_MODEL_NAME", "ainz/tiny-recursive-model"),
        trm_use_hf_api=os.getenv("TRM_USE_HF_API", "false").lower() == "true",
    )
    
    # Initialize domain memory (stateful agent - Nate B. Jones pattern)
    await chatbot.initialize_memory()
    
    # Initialize build info
    commit = os.getenv("GIT_COMMIT", "unknown")
    init_build_info(version, commit)
    
    # Initialize metrics with model labels so they appear in Prometheus immediately
    model = os.getenv("OLLAMA_MODEL", "llama3.2:3b")
    init_metrics(models=[model, "unknown"])
    
    memory_status = "enabled" if chatbot.memory_manager else "disabled"
    logger.info(
        "agent_bruno_initialized",
        version=version,
        events_enabled=True,
        domain_memory=memory_status,
        observability_enabled=observability_initialized,
        **get_current_trace_context(),  # Include trace context in logs
    )
    
    yield
    
    # Cleanup domain memory
    await chatbot.shutdown_memory()
    
    # Cleanup events
    await shutdown_events()
    logger.info("agent_bruno_shutdown", **get_current_trace_context())


app = FastAPI(
    title="Agent-Bruno",
    description="AI-powered chatbot assistant for homelab homepage",
    version="0.1.0",
    lifespan=lifespan,
)

# CORS middleware for frontend access
app.add_middleware(
    CORSMiddleware,
    allow_origins=os.getenv("CORS_ORIGINS", "*").split(","),
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# =============================================================================
# Request/Response Models
# =============================================================================

class ChatRequest(BaseModel):
    """Chat request from frontend."""
    message: str = Field(..., min_length=1, max_length=4096)
    conversation_id: Optional[str] = Field(default=None, max_length=64)


class ChatResponseModel(BaseModel):
    """Chat response to frontend."""
    response: str
    conversation_id: str
    tokens_used: int = 0
    model: str = ""
    duration_ms: float = 0.0


class HealthResponse(BaseModel):
    """Health check response."""
    status: str
    ollama_available: bool = False


# =============================================================================
# Endpoints
# =============================================================================

@app.get("/health", response_model=HealthResponse)
async def health():
    """Health check endpoint."""
    tracer = get_tracer()
    
    if tracer:
        with tracer.start_as_current_span("health_check") as span:
            span.set_attribute("endpoint", "/health")
            
            ollama_ok = await chatbot.health_check() if chatbot else False
            
            span.set_attribute("ollama_available", ollama_ok)
            span.set_attribute("observability_enabled", is_observability_enabled())
            
            return HealthResponse(
                status="healthy" if ollama_ok else "degraded",
                ollama_available=ollama_ok,
            )
    else:
        ollama_ok = await chatbot.health_check() if chatbot else False
        return HealthResponse(
            status="healthy" if ollama_ok else "degraded",
            ollama_available=ollama_ok,
        )


@app.get("/ready")
async def ready():
    """Readiness check endpoint."""
    if chatbot is None:
        raise HTTPException(status_code=503, detail="Chatbot not initialized")
    
    # Check Ollama connectivity
    if not await chatbot.health_check():
        raise HTTPException(status_code=503, detail="Ollama not available")
    
    return {"status": "ready"}


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST
    )


@app.post("/chat", response_model=ChatResponseModel)
async def chat(request: ChatRequest):
    """
    Process a chat message and return AI response.
    
    This is the main endpoint for the homepage chatbot widget.
    """
    if chatbot is None:
        raise HTTPException(status_code=503, detail="Chatbot not initialized")
    
    tracer = get_tracer()
    
    if tracer:
        with tracer.start_as_current_span("chat_request") as span:
            span.set_attribute("endpoint", "/chat")
            span.set_attribute("message_length", len(request.message))
            span.set_attribute("has_conversation_id", request.conversation_id is not None)
            
            logger.info(
                "chat_request_received",
                conversation_id=request.conversation_id,
                message_length=len(request.message),
                source="api",
                **get_current_trace_context(),  # Include trace context
            )
            
            # Use shared business logic
            result = await chatbot.chat(
                message=request.message,
                conversation_id=request.conversation_id,
            )
            
            span.set_attribute("tokens_used", result.tokens_used)
            span.set_attribute("model", result.model)
            span.set_attribute("duration_ms", result.duration_ms)
            span.set_attribute("response_length", len(result.response))
            
            return ChatResponseModel(
                response=result.response,
                conversation_id=result.conversation_id,
                tokens_used=result.tokens_used,
                model=result.model,
                duration_ms=result.duration_ms,
            )
    else:
        logger.info(
            "chat_request_received",
            conversation_id=request.conversation_id,
            message_length=len(request.message),
            source="api",
            **get_current_trace_context(),  # Include trace context
        )
        
        result = await chatbot.chat(
            message=request.message,
            conversation_id=request.conversation_id,
        )
        
        return ChatResponseModel(
            response=result.response,
            conversation_id=result.conversation_id,
            tokens_used=result.tokens_used,
            model=result.model,
            duration_ms=result.duration_ms,
        )


@app.post("/")
async def handle_cloudevent_root(request: Request):
    """
    Root CloudEvent handler - receives events at root path.
    
    This allows CloudEvents to be sent to both /events and / endpoints.
    Uses the same shared business logic as /events.
    """
    return await receive_cloudevent(request)


@app.get("/")
async def root():
    """Root endpoint with service info."""
    return {
        "service": "agent-bruno",
        "description": "AI-powered chatbot for homelab homepage",
        "version": os.getenv("VERSION", "0.1.0"),
        "endpoints": {
            "chat": "POST /chat",
            "events": "POST /events (CloudEvents)",
            "cloudevent": "POST / (CloudEvents)",
            "notifications": "GET /notifications",
            "health": "GET /health",
            "ready": "GET /ready",
            "metrics": "GET /metrics",
        }
    }


# =============================================================================
# CloudEvents Endpoints
# =============================================================================

@app.post("/events")
async def receive_cloudevent(request: Request):
    """
    Receive CloudEvents from Knative triggers.
    
    Subscribed events:
    - io.homelab.vuln.found: Vulnerability found by agent-contracts
    - io.homelab.exploit.validated: Exploit validated by agent-contracts
    - io.homelab.alert.fired: System alert fired
    - io.homelab.chat.message: Chat message (for tracing/analytics)
    
    These events are stored and can influence chatbot responses.
    """
    import time
    
    tracer = get_tracer()
    start_time = time.monotonic()
    
    try:
        # Parse CloudEvent from HTTP request
        headers = dict(request.headers)
        body = await request.body()
        event = from_http(headers, body)
        
        event_type = event["type"]
        event_source = event["source"]
        event_id = event.get("id", "unknown")
        
        # Create tracing span for CloudEvent processing
        if tracer:
            with tracer.start_as_current_span(
                f"cloudevent.{event_type.replace('.', '_')}",
            ) as span:
                span.set_attribute("cloudevent.type", event_type)
                span.set_attribute("cloudevent.source", event_source)
                span.set_attribute("cloudevent.id", event_id)
                span.set_attribute("cloudevent.specversion", event.get("specversion", "1.0"))
                span.set_attribute("endpoint", "/events")
                
                logger.info(
                    "cloudevent_received",
                    event_type=event_type,
                    source=event_source,
                    subject=event.get("subject"),
                    event_id=event_id,
                    source_type="cloudevent",
                    **get_current_trace_context(),  # Include trace context
                )
                
                # Handle the event using shared business logic
                if chatbot and chatbot.event_subscriber:
                    handled = await chatbot.event_subscriber.handle(event)
                    duration_ms = (time.monotonic() - start_time) * 1000
                    
                    span.set_attribute("cloudevent.handled", handled)
                    span.set_attribute("cloudevent.duration_ms", duration_ms)
                    
                    if handled:
                        return {
                            "status": "handled",
                            "event_type": event_type,
                            "event_id": event_id,
                            "duration_ms": duration_ms,
                        }
                    else:
                        return {
                            "status": "no_handler",
                            "event_type": event_type,
                            "event_id": event_id,
                        }
                
                return {"status": "subscriber_not_initialized"}
        else:
            # No tracer available, process without tracing
            logger.info(
                "cloudevent_received",
                event_type=event_type,
                source=event_source,
                subject=event.get("subject"),
                event_id=event_id,
                source_type="cloudevent",
                **get_current_trace_context(),
            )
            
            if chatbot and chatbot.event_subscriber:
                handled = await chatbot.event_subscriber.handle(event)
                duration_ms = (time.monotonic() - start_time) * 1000
                
                if handled:
                    return {
                        "status": "handled",
                        "event_type": event_type,
                        "event_id": event_id,
                        "duration_ms": duration_ms,
                    }
                else:
                    return {
                        "status": "no_handler",
                        "event_type": event_type,
                        "event_id": event_id,
                    }
            
            return {"status": "subscriber_not_initialized"}
        
    except Exception as e:
        duration_ms = (time.monotonic() - start_time) * 1000
        logger.error(
            "cloudevent_processing_failed",
            error=str(e),
            duration_ms=duration_ms,
            **get_current_trace_context(),  # Include trace context even on error
        )
        raise HTTPException(status_code=400, detail=f"Invalid CloudEvent: {str(e)}")


@app.get("/notifications")
async def get_notifications(limit: int = 10):
    """
    Get recent notifications received from other agents.
    
    This endpoint is useful for debugging and for the frontend
    to show security alerts.
    """
    if chatbot and chatbot.event_subscriber:
        notifications = chatbot.event_subscriber.get_recent_notifications(limit=limit)
        return {
            "count": len(notifications),
            "notifications": notifications,
        }
    return {"count": 0, "notifications": []}


@app.delete("/notifications")
async def clear_notifications():
    """Clear all stored notifications."""
    if chatbot and chatbot.event_subscriber:
        chatbot.event_subscriber.clear_notifications()
        return {"status": "cleared"}
    return {"status": "no_subscriber"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
