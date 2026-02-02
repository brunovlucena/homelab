"""
🎯 Speech Coach Agent - Autism Speech Development

LambdaAgent that helps autistic children develop speech skills through interactive games.
"""
import os
import json
import time
import uuid
from datetime import datetime
from contextlib import asynccontextmanager
from typing import Optional, Any, Dict

import httpx
import structlog
from fastapi import FastAPI, Request, Response, HTTPException
from fastapi.responses import JSONResponse
from cloudevents.http import from_http, CloudEvent, to_structured
from pydantic import BaseModel
from opentelemetry import trace
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST

from shared.types import (
    User, Exercise, GameSession, Progress, SpeechRequest, SpeechResponse,
    ExerciseType, DifficultyLevel, GameStatus
)
from shared.database import Database
from shared.metrics import (
    REQUESTS_TOTAL, RESPONSE_DURATION, LLM_INFERENCE_DURATION, TOKENS_USED,
    EXERCISES_COMPLETED, GAME_SESSIONS_TOTAL, init_build_info
)

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
OLLAMA_URL = os.getenv("OLLAMA_URL", os.getenv("AI_ENDPOINT", "http://ollama-native.ollama.svc.cluster.local:11434"))
OLLAMA_MODEL = os.getenv("OLLAMA_MODEL", os.getenv("AI_MODEL", "llama3.2:3b"))
SYSTEM_PROMPT = os.getenv("SYSTEM_PROMPT", "")
EVENT_SOURCE = os.getenv("EVENT_SOURCE", "/agent-speech-coach/games")
MAX_TOKENS = int(os.getenv("MAX_TOKENS", os.getenv("AI_MAX_TOKENS", "2048")))
TEMPERATURE = float(os.getenv("TEMPERATURE", os.getenv("AI_TEMPERATURE", "0.7")))

# Default system prompt for speech coach
DEFAULT_SYSTEM_PROMPT = """You are a friendly and patient speech coach for autistic children.

Your responsibilities:
- Encourage speech development through fun games and exercises
- Provide positive, encouraging feedback
- Adapt exercises to the child's level and interests
- Suggest appropriate exercises based on progress
- Use simple, clear language
- Be supportive and understanding

Important guidelines:
- Always be positive and encouraging
- Use age-appropriate language
- Break down complex tasks into simple steps
- Celebrate small wins
- Never be critical or push too hard
- Adapt to the child's pace and preferences
- Use visual and game-based approaches when possible

You can help with:
- Word repetition exercises
- Phrase completion games
- Story telling activities
- Conversation practice
- Imitation games
- Question-answer exercises

Remember: The goal is to make speech practice fun and engaging, not stressful."""

# Global instances
http_client: Optional[httpx.AsyncClient] = None
db: Optional[Database] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Manage application lifespan - startup and shutdown."""
    global http_client, db
    
    logger.info(
        "speech_coach_agent_starting",
        ollama_url=OLLAMA_URL,
        model=OLLAMA_MODEL,
    )
    
    # Initialize build info
    version = os.getenv("VERSION", "1.0.0")
    commit = os.getenv("GIT_COMMIT", "unknown")
    init_build_info(version, commit)
    
    # Create HTTP client
    http_client = httpx.AsyncClient(timeout=120.0)
    
    # Initialize database
    db = Database()
    try:
        await db.connect()
        logger.info("database_connected")
    except Exception as e:
        logger.error("database_connection_failed", error=str(e))
        # Continue without DB for development
    
    yield
    
    # Cleanup
    if http_client:
        await http_client.aclose()
    if db:
        await db.disconnect()
    
    logger.info("speech_coach_agent_shutdown")


app = FastAPI(
    title="🎯 Speech Coach Agent",
    description="AI agent for autism speech development",
    version="1.0.0",
    lifespan=lifespan,
)


def get_system_prompt() -> str:
    """Get the system prompt for this agent."""
    return SYSTEM_PROMPT or DEFAULT_SYSTEM_PROMPT


async def call_llm(prompt: str, model: str = None) -> tuple[str, int]:
    """Call LLM (Ollama) with the given prompt."""
    if not http_client:
        raise HTTPException(status_code=503, detail="HTTP client not initialized")
    
    model = model or OLLAMA_MODEL
    
    try:
        response = await http_client.post(
            f"{OLLAMA_URL}/api/generate",
            json={
                "model": model,
                "prompt": prompt,
                "system": get_system_prompt(),
                "stream": False,
                "options": {
                    "temperature": TEMPERATURE,
                    "num_predict": MAX_TOKENS,
                }
            }
        )
        response.raise_for_status()
        result = response.json()
        return (
            result.get("response", "").strip(),
            result.get("eval_count", 0)
        )
    except httpx.TimeoutException:
        logger.error("llm_timeout", url=OLLAMA_URL)
        raise HTTPException(status_code=504, detail="LLM request timed out")
    except httpx.HTTPStatusError as e:
        logger.error("llm_error", status=e.response.status_code, detail=str(e))
        raise HTTPException(status_code=502, detail=f"LLM service error: {e.response.status_code}")
    except Exception as e:
        logger.error("llm_exception", error=str(e))
        raise HTTPException(status_code=500, detail=f"LLM error: {str(e)}")


def get_default_exercises() -> Dict[str, Exercise]:
    """Get default exercises for speech development."""
    return {
        "word_rep_1": Exercise(
            id="word_rep_1",
            type=ExerciseType.WORD_REPETITION,
            title="Say the Word",
            description="Repeat simple words",
            difficulty=DifficultyLevel.BEGINNER,
            instructions="Listen and repeat the word after me",
            target_words=["cat", "dog", "ball", "car"],
            expected_duration_minutes=5,
            points=10
        ),
        "phrase_comp_1": Exercise(
            id="phrase_comp_1",
            type=ExerciseType.PHRASE_COMPLETION,
            title="Complete the Sentence",
            description="Finish the sentence",
            difficulty=DifficultyLevel.INTERMEDIATE,
            instructions="I'll start a sentence, you finish it",
            target_words=["happy", "good", "play"],
            expected_duration_minutes=10,
            points=20
        ),
        "story_1": Exercise(
            id="story_1",
            type=ExerciseType.STORY_TELLING,
            title="Tell a Story",
            description="Create a simple story",
            difficulty=DifficultyLevel.ADVANCED,
            instructions="Tell me about your day or make up a story",
            target_words=[],
            expected_duration_minutes=15,
            points=30
        ),
    }


async def process_speech_request(request: SpeechRequest) -> SpeechResponse:
    """Process speech coaching request."""
    start_time = time.time()
    
    logger.info(
        "speech_request_processing",
        user_id=request.user_id,
        query=request.query,
        exercise_type=request.exercise_type,
        session_id=request.session_id,
    )
    
    # Get or create user
    user = await db.get_user(request.user_id) if db else None
    if not user:
        user = User(id=request.user_id, name="Child")
        if db:
            await db.save_user(user)
    
    # Get progress
    progress = await db.get_progress(request.user_id) if db else None
    if not progress:
        progress = Progress(user_id=request.user_id)
        if db:
            await db.update_progress(progress)
    
    # Determine exercise
    exercise = None
    if request.exercise_type:
        exercises = get_default_exercises()
        # Find matching exercise
        for ex in exercises.values():
            if ex.type == request.exercise_type:
                exercise = ex
                break
    
    # Create or get session
    session = None
    if request.session_id and db:
        # Get existing session
        sessions = await db.get_user_sessions(request.user_id, limit=1)
        if sessions:
            session = sessions[0]
    
    if not session:
        session = GameSession(
            id=str(uuid.uuid4()),
            user_id=request.user_id,
            exercise_id=exercise.id if exercise else "general",
            status=GameStatus.IN_PROGRESS,
            started_at=datetime.utcnow(),
        )
        if db:
            await db.save_session(session)
        GAME_SESSIONS_TOTAL.labels(status="started").inc()
    
    # Build prompt for LLM
    prompt = f"""Child's request: {request.query}

Context:
- User: {user.name} (ID: {user.id})
- Exercise: {exercise.title if exercise else 'General conversation'}
- Progress: {progress.total_sessions} sessions completed, {progress.total_points} points earned

Please provide:
1. A friendly, encouraging response
2. Suggestions for next steps or exercises
3. Positive feedback if they completed something
4. A fun way to continue practicing

Keep your response short, simple, and encouraging. Use age-appropriate language."""

    # Call LLM
    with LLM_INFERENCE_DURATION.labels(model=OLLAMA_MODEL).time():
        response_text, tokens = await call_llm(prompt)
    
    # Update session
    session.attempts += 1
    if db:
        await db.save_session(session)
    
    # Calculate duration
    duration_ms = (time.time() - start_time) * 1000
    
    # Record metrics
    REQUESTS_TOTAL.labels(status="success", exercise_type=exercise.type.value if exercise else "general").inc()
    RESPONSE_DURATION.labels(model=OLLAMA_MODEL).observe(duration_ms / 1000)
    TOKENS_USED.labels(model=OLLAMA_MODEL, type="total").inc(tokens)
    
    logger.info(
        "speech_request_processed",
        user_id=request.user_id,
        tokens=tokens,
        duration_ms=duration_ms,
    )
    
    return SpeechResponse(
        response=response_text,
        exercise=exercise,
        session=session,
        progress=progress,
        suggestions=[],
        encouragement="Great job! Keep practicing!",
        model=OLLAMA_MODEL,
        tokens_used=tokens,
        duration_ms=duration_ms,
    )


@app.get("/health")
async def health():
    """Health check endpoint."""
    db_ok = db is not None
    return {
        "status": "healthy" if db_ok else "degraded",
        "agent": "agent-speech-coach",
        "database": "connected" if db_ok else "disconnected",
    }


@app.get("/ready")
async def ready():
    """Readiness check."""
    db_ready = bool(db)
    llm_ready = False
    
    if http_client:
        try:
            response = await http_client.get(f"{OLLAMA_URL}/api/tags", timeout=5.0)
            llm_ready = response.status_code == 200
        except:
            pass
    
    if db_ready and llm_ready:
        return {"status": "ready", "agent": "agent-speech-coach"}
    
    return JSONResponse(
        status_code=503,
        content={
            "status": "not_ready",
            "agent": "agent-speech-coach",
            "database": "ready" if db_ready else "not_ready",
            "llm": "ready" if llm_ready else "not_ready",
        }
    )


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST
    )


@app.post("/")
async def handle_event(request: Request):
    """Handle incoming CloudEvents."""
    start_time = time.time()
    
    tracer = trace.get_tracer(__name__)
    
    try:
        headers = dict(request.headers)
        body = await request.body()
        
        # Parse CloudEvent
        if headers.get("ce-type") or headers.get("content-type") == "application/cloudevents+json":
            event = from_http(headers, body)
            event_type = event["type"]
            event_data = event.data or {}
            event_id = event["id"]
            
            with tracer.start_as_current_span(
                f"cloudevent.{event_type.replace('.', '_')}",
                attributes={
                    "cloudevent.type": event_type,
                    "cloudevent.id": event_id,
                }
            ) as span:
                logger.info(
                    "event_received",
                    event_type=event_type,
                    event_id=event_id,
                )
                
                # Extract data
                user_id = event_data.get("user_id") or event_data.get("userId") or "default"
                query = event_data.get("query") or event_data.get("message") or event_data.get("content") or ""
                exercise_type = event_data.get("exercise_type")
                
                if not query:
                    raise HTTPException(status_code=400, detail="Query is required")
                
                # Create request
                speech_request = SpeechRequest(
                    user_id=user_id,
                    query=query,
                    exercise_type=ExerciseType(exercise_type) if exercise_type else None,
                    metadata=event_data.get("metadata", {}),
                )
                
                # Process request
                response_data = await process_speech_request(speech_request)
                
                span.set_attribute("cloudevent.duration_ms", (time.time() - start_time) * 1000)
                
                # Return as CloudEvent
                response_event = CloudEvent({
                    "type": "io.homelab.speech-coach.response",
                    "source": EVENT_SOURCE,
                    "id": str(uuid.uuid4()),
                    "time": datetime.utcnow().isoformat() + "Z",
                    "datacontenttype": "application/json",
                }, response_data.model_dump(mode="json"))
                
                headers, body = to_structured(response_event)
                return Response(
                    content=body,
                    media_type="application/cloudevents+json",
                    headers=dict(headers),
                )
        else:
            # Plain JSON request
            event_data = json.loads(body) if body else {}
            user_id = event_data.get("user_id") or event_data.get("userId") or "default"
            query = event_data.get("query") or event_data.get("message") or ""
            
            if not query:
                raise HTTPException(status_code=400, detail="Query is required")
            
            speech_request = SpeechRequest(
                user_id=user_id,
                query=query,
                exercise_type=ExerciseType(event_data.get("exercise_type")) if event_data.get("exercise_type") else None,
            )
            
            response_data = await process_speech_request(speech_request)
            
            # Return as CloudEvent for consistency
            response_event = CloudEvent({
                "type": "io.homelab.speech-coach.response",
                "source": EVENT_SOURCE,
                "id": str(uuid.uuid4()),
                "time": datetime.utcnow().isoformat() + "Z",
                "datacontenttype": "application/json",
            }, response_data.model_dump(mode="json"))
            
            headers, body = to_structured(response_event)
            return Response(
                content=body,
                media_type="application/cloudevents+json",
                headers=dict(headers),
            )
    
    except HTTPException:
        raise
    except Exception as e:
        logger.error("event_processing_error", error=str(e))
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/info")
async def info():
    """Get agent information."""
    return {
        "name": "agent-speech-coach",
        "description": "Speech development coach for autistic children",
        "model": OLLAMA_MODEL,
        "endpoint": OLLAMA_URL,
        "event_source": EVENT_SOURCE,
        "version": "1.0.0",
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
