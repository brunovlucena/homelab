"""
🍽️ Restaurant Agent - Generic AI-Powered Restaurant Service Agent with Domain Memory

This is a generic agent that can be configured via environment variables
to act as different restaurant staff (Host, Waiter, Sommelier, Chef).

Following Nate B. Jones's Domain Memory Factory pattern:
- Persistent memory for guest preferences and service history
- Working memory for active tables, orders, and service state
- Entity memory for menu items, tables, and reservations
- Agents that remember regular guests and their preferences

The agent:
1. Receives CloudEvents (HTTP POST)
2. Processes them using Ollama LLM with role-specific prompts
3. Maintains domain memory for stateful service
4. Emits response events

Environment Variables:
- AGENT_NAME: Name of the agent (e.g., "Pierre", "Maximilian")
- AGENT_ROLE: Role (waiter, host, sommelier, chef)
- OLLAMA_URL: Ollama endpoint
- OLLAMA_MODEL: Model to use
- SYSTEM_PROMPT: Role-specific system prompt (optional, uses default if not set)
- MEMORY_ENABLED: Enable domain memory (default: true)
- REDIS_URL: Redis URL for short-term memory
- POSTGRES_URL: PostgreSQL URL for long-term memory
"""
import json
import os
import uuid
from contextlib import asynccontextmanager
from datetime import datetime
from typing import Any

import httpx
import structlog
from cloudevents.http import CloudEvent, from_http, to_structured
from fastapi import FastAPI, HTTPException, Request, Response
from fastapi.responses import JSONResponse
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource
from opentelemetry.semconv.resource import ResourceAttributes
from pydantic import BaseModel

# Configure OpenTelemetry tracing
OTEL_EXPORTER_OTLP_ENDPOINT = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
OTEL_SERVICE_NAME = os.getenv("OTEL_SERVICE_NAME", f"agent-restaurant-{os.getenv('AGENT_ROLE', 'unknown')}")

# Set up tracing if OTLP endpoint is configured
if OTEL_EXPORTER_OTLP_ENDPOINT:
    try:
        from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
        
        resource = Resource.create({
            ResourceAttributes.SERVICE_NAME: OTEL_SERVICE_NAME,
            ResourceAttributes.SERVICE_VERSION: os.getenv("VERSION", "0.1.0"),
            "agent.name": os.getenv("AGENT_NAME", "unknown"),
            "agent.role": os.getenv("AGENT_ROLE", "unknown"),
        })
        
        provider = TracerProvider(resource=resource)
        processor = BatchSpanProcessor(OTLPSpanExporter(endpoint=OTEL_EXPORTER_OTLP_ENDPOINT))
        provider.add_span_processor(processor)
        trace.set_tracer_provider(provider)
    except ImportError:
        pass  # OTLP exporter not installed

from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
from restaurant_agent.metrics import (
    init_build_info,
    REQUESTS_TOTAL,
    REQUEST_DURATION,
    CLOUDEVENTS_RECEIVED,
    CLOUDEVENTS_PROCESSED,
    EVENT_PROCESSING_DURATION,
    LLM_CALLS,
    LLM_DURATION,
    TOKENS_USED,
    MEMORY_OPERATIONS,
    MEMORY_CONTEXT_BUILD_DURATION,
    GUESTS_SERVED,
    ACTIVE_TABLES,
    record_memory_operation,
    record_guest_served,
    record_guest_preference,
    record_guest_fact,
    set_memory_store_connected,
)

# Import Domain Memory components
try:
    from agent_memory import DomainMemoryManager
    MEMORY_AVAILABLE = True
except ImportError:
    MEMORY_AVAILABLE = False

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
        structlog.processors.JSONRenderer()
    ],
    wrapper_class=structlog.stdlib.BoundLogger,
    context_class=dict,
    logger_factory=structlog.stdlib.LoggerFactory(),
    cache_logger_on_first_use=True,
)

logger = structlog.get_logger()

# Configuration from environment
AGENT_NAME = os.getenv("AGENT_NAME", "Agent")
AGENT_ROLE = os.getenv("AGENT_ROLE", "assistant")
OLLAMA_URL = os.getenv("OLLAMA_URL", os.getenv("AI_ENDPOINT", "http://ollama-native.ollama.svc.cluster.local:11434"))
OLLAMA_MODEL = os.getenv("OLLAMA_MODEL", os.getenv("AI_MODEL", "llama3.2:3b"))
SYSTEM_PROMPT = os.getenv("SYSTEM_PROMPT", "")
EVENT_SOURCE = os.getenv("EVENT_SOURCE", f"/agent-restaurant/{AGENT_ROLE}/{AGENT_NAME.lower()}")
MAX_TOKENS = int(os.getenv("MAX_TOKENS", os.getenv("AI_MAX_TOKENS", "1024")))
TEMPERATURE = float(os.getenv("TEMPERATURE", os.getenv("AI_TEMPERATURE", "0.7")))

# Memory configuration
MEMORY_ENABLED = os.getenv("MEMORY_ENABLED", "true").lower() == "true"
REDIS_URL = os.getenv("REDIS_URL")
POSTGRES_URL = os.getenv("POSTGRES_URL")

# Default system prompts for each role
DEFAULT_PROMPTS = {
    "host": """You are {name}, an elegant host at an upscale restaurant.
Be warm, professional, and excellent at managing guest arrivals and seating.
Remember returning guests and their preferences.
Respond in JSON format with: action, greeting, tableId, emotion.""",

    "waiter": """You are {name}, a knowledgeable waiter at an upscale restaurant.
Present dishes with theatrical flair and tell ingredient stories.
Remember guest preferences and dietary restrictions.
Respond in JSON format with: action, message, presentation, emotion.""",

    "sommelier": """You are {name}, a sophisticated sommelier at an upscale restaurant.
Provide wine pairing recommendations with passion and expertise.
Remember guests' wine preferences and past selections.
Respond in JSON format with: action, recommendation, pairing, notes.""",

    "chef": """You are {name}, a passionate chef managing the kitchen.
Coordinate orders, ensure quality, and manage timing perfectly.
Track order progress and kitchen capacity.
Respond in JSON format with: action, status, timing, notes.""",
}

# Global instances
http_client: httpx.AsyncClient | None = None
memory_manager: DomainMemoryManager | None = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Manage application lifespan - startup and shutdown."""
    global http_client, memory_manager

    # Initialize build info for Agent Versions dashboard
    version = os.getenv("VERSION", "0.1.0")
    commit = os.getenv("GIT_COMMIT", "unknown")
    init_build_info(version, commit)

    logger.info(
        "agent_starting",
        agent_name=AGENT_NAME,
        agent_role=AGENT_ROLE,
        version=version,
        ollama_url=OLLAMA_URL,
        model=OLLAMA_MODEL,
        memory_enabled=MEMORY_ENABLED and MEMORY_AVAILABLE,
    )

    # Create HTTP client
    http_client = httpx.AsyncClient(timeout=120.0)

    # Initialize Domain Memory (Nate B. Jones pattern)
    if MEMORY_ENABLED and MEMORY_AVAILABLE:
        try:
            memory_manager = DomainMemoryManager(
                agent_id=f"agent-restaurant-{AGENT_ROLE}-{AGENT_NAME.lower()}",
                agent_type="restaurant",
                domain="hospitality",
                redis_url=REDIS_URL,
                postgres_url=POSTGRES_URL,
                use_redis=bool(REDIS_URL),
                use_postgres=bool(POSTGRES_URL),
                default_constraints=[
                    {
                        "description": "Maintain professional hospitality demeanor at all times",
                        "hard": True,
                        "category": "behavior",
                    },
                    {
                        "description": "Respect guest dietary restrictions and allergies",
                        "hard": True,
                        "category": "safety",
                    },
                    {
                        "description": "Protect guest privacy and payment information",
                        "hard": True,
                        "category": "privacy",
                    },
                ],
            )
            await memory_manager.connect()
            set_memory_store_connected("redis", bool(REDIS_URL))
            set_memory_store_connected("postgres", bool(POSTGRES_URL))
            logger.info(
                "domain_memory_initialized",
                role=AGENT_ROLE,
                agent=AGENT_NAME,
                redis_enabled=bool(REDIS_URL),
                postgres_enabled=bool(POSTGRES_URL),
            )
        except Exception as e:
            set_memory_store_connected("redis", False)
            set_memory_store_connected("postgres", False)
            logger.error("domain_memory_init_failed", error=str(e), agent=AGENT_NAME)
            memory_manager = None

    yield

    # Cleanup
    if memory_manager:
        try:
            await memory_manager.disconnect()
        except Exception as e:
            logger.error("domain_memory_disconnect_failed", error=str(e))

    if http_client:
        await http_client.aclose()

    logger.info("agent_shutdown", agent_name=AGENT_NAME)


app = FastAPI(
    title=f"🍽️ Restaurant Agent - {AGENT_NAME}",
    description=f"AI-powered {AGENT_ROLE} agent with domain memory for restaurant operations",
    version="0.1.0",
    lifespan=lifespan,
)


class AgentResponse(BaseModel):
    """Response from the agent."""
    agent: str
    role: str
    response: Any
    event_id: str
    timestamp: str
    memory_context: dict | None = None


def get_system_prompt() -> str:
    """Get the system prompt for this agent."""
    if SYSTEM_PROMPT:
        return SYSTEM_PROMPT

    default = DEFAULT_PROMPTS.get(AGENT_ROLE.lower(), DEFAULT_PROMPTS["waiter"])
    return default.format(name=AGENT_NAME)


async def get_guest_context(guest_id: str) -> dict:
    """
    Get context about a guest from domain memory.

    This enables personalized service for returning guests.
    """
    if not memory_manager:
        return {}

    try:
        # Get user memory (preferences, history)
        user_mem = await memory_manager.get_user_memory(guest_id)
        if user_mem:
            return {
                "is_returning_guest": True,
                "preferences": user_mem.preferences,
                "facts": [f["fact"] for f in user_mem.facts[-5:]],
                "visit_count": user_mem.interaction_stats.get("total_interactions", 0),
            }
    except Exception as e:
        logger.warning("guest_context_fetch_failed", error=str(e))

    return {"is_returning_guest": False}


async def remember_guest_preference(guest_id: str, preference_type: str, value: Any):
    """Remember a guest's preference for future visits."""
    if not memory_manager:
        return

    try:
        await memory_manager.update_user_preference(guest_id, preference_type, value)
        logger.info(
            "guest_preference_saved",
            guest_id=guest_id,
            preference_type=preference_type,
        )
    except Exception as e:
        logger.warning("guest_preference_save_failed", error=str(e))


async def remember_guest_fact(guest_id: str, fact: str):
    """Remember a fact about a guest (e.g., dietary restriction, special occasion)."""
    if not memory_manager:
        return

    try:
        await memory_manager.add_user_fact(guest_id, fact, source="service_interaction")
        logger.info("guest_fact_saved", guest_id=guest_id)
    except Exception as e:
        logger.warning("guest_fact_save_failed", error=str(e))


async def track_table_service(table_id: str, event_type: str, details: dict):
    """Track service events for a table."""
    if not memory_manager:
        return

    try:
        # Create or update table entity
        await memory_manager.create_or_update_entity(
            entity_type="table",
            entity_id=table_id,
            attributes={
                "last_event": event_type,
                "last_event_time": datetime.utcnow().isoformat(),
                **details,
            },
        )
    except Exception as e:
        logger.warning("table_tracking_failed", error=str(e))


async def call_ollama(prompt: str, context: dict = None, guest_context: dict = None) -> str:
    """Call Ollama API with the given prompt and memory context."""
    import time
    start_time = time.time()
    
    if not http_client:
        LLM_CALLS.labels(model=OLLAMA_MODEL, status="error").inc()
        raise HTTPException(status_code=503, detail="HTTP client not initialized")

    system_prompt = get_system_prompt()

    # Enhance system prompt with guest context if available
    if guest_context and guest_context.get("is_returning_guest"):
        system_prompt += f"""

RETURNING GUEST CONTEXT:
- Visit count: {guest_context.get('visit_count', 'unknown')}
- Known preferences: {guest_context.get('preferences', {})}
- Notes about guest: {', '.join(guest_context.get('facts', []))}

Use this information to provide personalized service!"""

    # Build the full prompt with context
    full_prompt = f"""Context: {json.dumps(context or {})}

User Request: {prompt}

Respond as {AGENT_NAME} the {AGENT_ROLE}. Use JSON format."""

    try:
        logger.info(
            "llm_call_started",
            model=OLLAMA_MODEL,
            prompt_length=len(full_prompt),
            agent=AGENT_NAME,
            role=AGENT_ROLE,
        )
        
        response = await http_client.post(
            f"{OLLAMA_URL}/api/generate",
            json={
                "model": OLLAMA_MODEL,
                "prompt": full_prompt,
                "system": system_prompt,
                "stream": False,
                "options": {
                    "temperature": TEMPERATURE,
                    "num_predict": MAX_TOKENS,
                }
            }
        )
        response.raise_for_status()
        result = response.json()
        
        # Record successful metrics
        duration = time.time() - start_time
        LLM_CALLS.labels(model=OLLAMA_MODEL, status="success").inc()
        LLM_DURATION.labels(model=OLLAMA_MODEL).observe(duration)
        
        # Record token usage if available
        if "prompt_eval_count" in result:
            TOKENS_USED.labels(model=OLLAMA_MODEL, type="input").inc(result["prompt_eval_count"])
        if "eval_count" in result:
            TOKENS_USED.labels(model=OLLAMA_MODEL, type="output").inc(result["eval_count"])
        
        logger.info(
            "llm_call_completed",
            model=OLLAMA_MODEL,
            duration_seconds=duration,
            response_length=len(result.get("response", "")),
            agent=AGENT_NAME,
        )
        
        return result.get("response", "")

    except httpx.TimeoutException:
        LLM_CALLS.labels(model=OLLAMA_MODEL, status="timeout").inc()
        logger.error("ollama_timeout", url=OLLAMA_URL, duration_seconds=time.time() - start_time)
        raise HTTPException(status_code=504, detail="LLM request timed out")
    except httpx.HTTPStatusError as e:
        LLM_CALLS.labels(model=OLLAMA_MODEL, status="error").inc()
        logger.error("ollama_error", status=e.response.status_code, detail=str(e))
        raise HTTPException(status_code=502, detail=f"LLM service error: {e.response.status_code}")
    except Exception as e:
        LLM_CALLS.labels(model=OLLAMA_MODEL, status="error").inc()
        logger.error("ollama_exception", error=str(e), duration_seconds=time.time() - start_time)
        raise HTTPException(status_code=500, detail=f"LLM error: {str(e)}")


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST,
    )


@app.get("/health")
async def health():
    """Health check endpoint."""
    return {
        "status": "healthy",
        "agent": AGENT_NAME,
        "role": AGENT_ROLE,
        "memory_enabled": memory_manager is not None,
    }


@app.get("/ready")
async def ready():
    """Readiness check - verify Ollama is accessible."""
    try:
        if http_client:
            response = await http_client.get(f"{OLLAMA_URL}/api/tags", timeout=5.0)
            if response.status_code == 200:
                return {
                    "status": "ready",
                    "agent": AGENT_NAME,
                    "llm": "connected",
                    "memory": "connected" if memory_manager else "disabled",
                }
    except Exception as e:
        logger.warning("readiness_check_failed", error=str(e))

    return JSONResponse(
        status_code=503,
        content={"status": "not_ready", "agent": AGENT_NAME, "llm": "disconnected"}
    )


async def process_restaurant_request(
    prompt: str,
    context: dict = None,
    event_id: str = None,
    source: str = "api",
    guest_id: str = None,
    table_id: str = None,
):
    """
    Shared business logic for processing restaurant agent requests.

    Now enhanced with domain memory for:
    - Guest preference tracking
    - Service history
    - Table state management

    Args:
        prompt: The request/prompt to process
        context: Optional context data
        event_id: Optional event ID for tracing
        source: Source of the request ("api" or "cloudevent")
        guest_id: Optional guest identifier for personalization
        table_id: Optional table identifier for tracking

    Returns:
        AgentResponse with the agent's response
    """
    import time
    start_time = time.time()
    request_type = f"{AGENT_ROLE}_request"
    
    logger.info(
        "request_processing",
        prompt_length=len(prompt),
        event_id=event_id,
        source=source,
        agent=AGENT_NAME,
        role=AGENT_ROLE,
        guest_id=guest_id,
        table_id=table_id,
        memory_enabled=memory_manager is not None,
    )

    try:
        # Get guest context from memory if available
        guest_context = {}
        if guest_id:
            context_start = time.time()
            guest_context = await get_guest_context(guest_id)
            MEMORY_CONTEXT_BUILD_DURATION.observe(time.time() - context_start)
            
            if guest_context.get("is_returning_guest"):
                logger.info(
                    "returning_guest_recognized",
                    guest_id=guest_id,
                    visit_count=guest_context.get("visit_count"),
                    agent=AGENT_NAME,
                )
                record_memory_operation("read", "user_memory", "success")

        # Call LLM with memory-enhanced context
        llm_response = await call_ollama(prompt, context=context, guest_context=guest_context)

        # Try to parse JSON response
        try:
            parsed_response = json.loads(llm_response)
        except json.JSONDecodeError:
            parsed_response = {"message": llm_response}

        # Extract and remember any preferences mentioned
        if guest_id and memory_manager:
            # Check for dietary restrictions, preferences in the response
            if "dietary" in prompt.lower() or "allergy" in prompt.lower():
                await remember_guest_fact(guest_id, f"Mentioned dietary concern: {prompt[:100]}")
                record_guest_fact("conversation")

            if "prefer" in prompt.lower():
                await remember_guest_fact(guest_id, f"Expressed preference: {prompt[:100]}")
                record_guest_preference("general")

        # Track table service event
        if table_id:
            await track_table_service(
                table_id,
                event_type=f"{AGENT_ROLE}_interaction",
                details={
                    "agent": AGENT_NAME,
                    "action": parsed_response.get("action", "unknown"),
                },
            )
            ACTIVE_TABLES.inc()

        # Record interaction for user memory
        if guest_id and memory_manager:
            await memory_manager.record_user_interaction(guest_id, f"{AGENT_ROLE}_service")
            record_memory_operation("write", "user_interaction", "success")

        # Record successful request metrics
        duration = time.time() - start_time
        REQUESTS_TOTAL.labels(request_type=request_type, status="success").inc()
        REQUEST_DURATION.labels(request_type=request_type).observe(duration)
        record_guest_served(AGENT_ROLE)

        # Build response
        response_data = AgentResponse(
            agent=AGENT_NAME,
            role=AGENT_ROLE,
            response=parsed_response,
            event_id=event_id or str(uuid.uuid4()),
            timestamp=datetime.utcnow().isoformat(),
            memory_context={
                "guest_recognized": guest_context.get("is_returning_guest", False),
                "memory_enabled": memory_manager is not None,
            } if memory_manager else None,
        )

        logger.info(
            "request_processed",
            event_id=event_id,
            source=source,
            agent=AGENT_NAME,
            role=AGENT_ROLE,
            guest_recognized=guest_context.get("is_returning_guest", False),
            duration_seconds=duration,
            status="success",
        )

        return response_data
        
    except Exception as e:
        duration = time.time() - start_time
        REQUESTS_TOTAL.labels(request_type=request_type, status="error").inc()
        REQUEST_DURATION.labels(request_type=request_type).observe(duration)
        logger.error(
            "request_failed",
            event_id=event_id,
            source=source,
            agent=AGENT_NAME,
            error=str(e),
            duration_seconds=duration,
        )
        raise


@app.post("/")
async def handle_event(request: Request):
    """
    Handle incoming CloudEvents.

    This is the main entry point for event-driven interactions.
    Now with domain memory for stateful service.
    """
    import time
    start_time = time.time()

    tracer = trace.get_tracer(__name__)

    # Parse CloudEvent
    try:
        headers = dict(request.headers)
        body = await request.body()

        # Check if it's a CloudEvent
        if headers.get("ce-type") or headers.get("content-type") == "application/cloudevents+json":
            event = from_http(headers, body)
            event_type = event["type"]
            event_data = event.data or {}
            event_id = event["id"]
            event_source = event["source"]

            # Extract identifiers for memory
            guest_id = event_data.get("guest_id") or event_data.get("customer_id")
            table_id = event_data.get("table_id")

            # Record CloudEvent received
            CLOUDEVENTS_RECEIVED.labels(event_type=event_type, source=event_source).inc()
            
            # Create tracing span for CloudEvent processing
            with tracer.start_as_current_span(
                f"cloudevent.{event_type.replace('.', '_').replace(':', '_')}",
                attributes={
                    "cloudevent.type": event_type,
                    "cloudevent.source": event_source,
                    "cloudevent.id": event_id,
                    "agent.name": AGENT_NAME,
                    "agent.role": AGENT_ROLE,
                    "guest.id": guest_id or "anonymous",
                }
            ) as span:
                logger.info(
                    "cloudevent_received",
                    event_type=event_type,
                    event_id=event_id,
                    source=event_source,
                    agent=AGENT_NAME,
                    role=AGENT_ROLE,
                    source_type="cloudevent",
                    guest_id=guest_id,
                )

                # Extract request/prompt from event data
                prompt = event_data.get("request") or event_data.get("message") or json.dumps(event_data)

                # Use shared business logic with memory
                response_data = await process_restaurant_request(
                    prompt=prompt,
                    context=event_data,
                    event_id=event_id,
                    source="cloudevent",
                    guest_id=guest_id,
                    table_id=table_id,
                )

                # Record successful event processing
                event_duration = time.time() - start_time
                span.set_attribute("cloudevent.duration_ms", event_duration * 1000)
                CLOUDEVENTS_PROCESSED.labels(event_type=event_type, status="success").inc()
                EVENT_PROCESSING_DURATION.labels(event_type=event_type).observe(event_duration)
                
                logger.info(
                    "cloudevent_processed",
                    event_type=event_type,
                    event_id=event_id,
                    agent=AGENT_NAME,
                    role=AGENT_ROLE,
                    duration_seconds=event_duration,
                    status="success",
                )

                # Return as CloudEvent response
                response_event = CloudEvent({
                    "type": f"restaurant.{AGENT_ROLE}.response",
                    "source": EVENT_SOURCE,
                    "id": str(uuid.uuid4()),
                    "time": datetime.utcnow().isoformat() + "Z",
                    "datacontenttype": "application/json",
                }, response_data.model_dump())

                headers, body = to_structured(response_event)

                return Response(
                    content=body,
                    media_type="application/cloudevents+json",
                    headers=dict(headers),
                )
        else:
            # Plain JSON request (API endpoint)
            event_data = json.loads(body) if body else {}
            event_type = "direct.request"
            event_id = str(uuid.uuid4())
            event_source = "direct"

            # Extract identifiers for memory
            guest_id = event_data.get("guest_id") or event_data.get("customer_id")
            table_id = event_data.get("table_id")

            logger.info(
                "request_received",
                event_type=event_type,
                event_id=event_id,
                source=event_source,
                agent=AGENT_NAME,
                source_type="api",
            )

            # Extract request/prompt from event data
            prompt = event_data.get("request") or event_data.get("message") or json.dumps(event_data)

            # Use shared business logic
            response_data = await process_restaurant_request(
                prompt=prompt,
                context=event_data,
                event_id=event_id,
                source="api",
                guest_id=guest_id,
                table_id=table_id,
            )

            # Return as CloudEvent response (for consistency)
            response_event = CloudEvent({
                "type": f"restaurant.{AGENT_ROLE}.response",
                "source": EVENT_SOURCE,
                "id": str(uuid.uuid4()),
                "time": datetime.utcnow().isoformat() + "Z",
                "datacontenttype": "application/json",
            }, response_data.model_dump())

            headers, body = to_structured(response_event)

            return Response(
                content=body,
                media_type="application/cloudevents+json",
                headers=dict(headers),
            )

    except Exception as e:
        logger.error(
            "event_processing_error",
            error=str(e),
            agent=AGENT_NAME,
        )
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/info")
async def info():
    """Get agent information."""
    return {
        "name": AGENT_NAME,
        "role": AGENT_ROLE,
        "model": OLLAMA_MODEL,
        "endpoint": OLLAMA_URL,
        "event_source": EVENT_SOURCE,
        "version": "0.1.0",
        "memory_enabled": memory_manager is not None,
        "memory_type": "domain_memory" if memory_manager else "none",
    }


@app.get("/guest/{guest_id}")
async def get_guest_info(guest_id: str):
    """
    Get information about a guest from memory.

    Useful for hosts to prepare for returning guests.
    """
    if not memory_manager:
        return {"error": "Memory not enabled"}

    context = await get_guest_context(guest_id)
    return {
        "guest_id": guest_id,
        **context,
    }


@app.post("/guest/{guest_id}/preference")
async def save_guest_preference(guest_id: str, request: Request):
    """Save a preference for a guest."""
    if not memory_manager:
        raise HTTPException(status_code=503, detail="Memory not enabled")

    body = await request.json()
    preference_type = body.get("type")
    value = body.get("value")

    if not preference_type or value is None:
        raise HTTPException(status_code=400, detail="type and value required")

    await remember_guest_preference(guest_id, preference_type, value)
    return {"status": "saved", "guest_id": guest_id, "type": preference_type}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
