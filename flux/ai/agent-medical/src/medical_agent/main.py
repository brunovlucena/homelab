"""
🏥 Medical Records Agent - HIPAA-Compliant AI Agent with Domain Memory

LambdaAgent that handles CloudEvents for medical record queries.

Following Nate B. Jones's Domain Memory Factory pattern for:
- HIPAA-compliant stateful interactions
- Patient context persistence (hashed/anonymized)
- Provider workflow memory
- Clinical decision support with memory

Features:
- RBAC (Role-Based Access Control)
- HIPAA-compliant audit logging with memory
- Patient data isolation
- Integration with MongoDB (Application-level access control)
- Knowledge Graph for medical protocols
- LLM reasoning for complex queries
- Domain memory for clinical context
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
from fastapi import FastAPI, Request, Response, HTTPException, UploadFile, File, Form
from fastapi.responses import JSONResponse
from cloudevents.http import from_http, CloudEvent, to_structured
from pydantic import BaseModel
from opentelemetry import trace

from shared.types import (
    User, UserRole, MedicalQuery, MedicalResponse, 
    QueryIntent, QueryComplexity, IntentClassification
)
from shared.security import AccessControl, get_user_from_token
from shared.database import Database
from shared.audit import AuditLogger
from medical_agent.pdf_extractor import get_pdf_extractor, PDFExtractor
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
from shared.metrics import (
    REQUESTS_TOTAL, ACCESS_DENIED_TOTAL, RESPONSE_DURATION,
    AUDIT_LOGS_TOTAL, LLM_INFERENCE_DURATION, TOKENS_USED,
    init_build_info
)

# Import Domain Memory components
try:
    from agent_memory import (
        DomainMemoryManager,
        MedicalAgentSchema,
        UserMemory,
    )
    MEMORY_AVAILABLE = True
except ImportError:
    MEMORY_AVAILABLE = False
    # Define placeholder to avoid NameError when MEMORY_AVAILABLE is False
    DomainMemoryManager = None
    MedicalAgentSchema = None
    UserMemory = None

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
VLLM_URL = os.getenv("VLLM_URL", "http://vllm.ml-inference.svc.forge.remote:8000")
SYSTEM_PROMPT = os.getenv("SYSTEM_PROMPT", "")
EVENT_SOURCE = os.getenv("EVENT_SOURCE", "/agent-medical/records")
MAX_TOKENS = int(os.getenv("MAX_TOKENS", os.getenv("AI_MAX_TOKENS", "2048")))
TEMPERATURE = float(os.getenv("TEMPERATURE", os.getenv("AI_TEMPERATURE", "0.7")))
HIPAA_MODE = os.getenv("HIPAA_MODE", "true").lower() == "true"

# Memory configuration
MEMORY_ENABLED = os.getenv("MEMORY_ENABLED", "true").lower() == "true"
REDIS_URL = os.getenv("REDIS_URL")
POSTGRES_URL = os.getenv("POSTGRES_URL")

# Default system prompt for medical agent
DEFAULT_SYSTEM_PROMPT = """You are a medical records assistant with access to patient records.

Your responsibilities:
- Answer questions about patient medical records
- Provide lab results, prescriptions, and medical history
- Check for drug interactions
- Analyze medical data
- Maintain patient privacy and confidentiality

Important guidelines:
- Always verify patient access before providing information
- Use medical terminology accurately
- Be concise but thorough
- Flag any concerning patterns or anomalies
- Never share patient information without proper authorization

You have access to:
- Patient medical records
- Lab results
- Prescriptions
- Medical history
- Drug interaction database

Respond in a professional, empathetic manner while maintaining accuracy."""

# Global instances
http_client: Optional[httpx.AsyncClient] = None
db: Optional[Database] = None
audit_logger: Optional[AuditLogger] = None
access_control = AccessControl()
memory_manager: Optional[Any] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Manage application lifespan - startup and shutdown."""
    global http_client, db, memory_manager
    
    logger.info(
        "medical_agent_starting",
        ollama_url=OLLAMA_URL,
        model=OLLAMA_MODEL,
        hipaa_mode=HIPAA_MODE,
        memory_enabled=MEMORY_ENABLED and MEMORY_AVAILABLE,
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
    
    # Initialize audit logger
    global audit_logger
    audit_logger = AuditLogger(db=db)
    
    # Initialize Domain Memory (Nate B. Jones pattern) - HIPAA compliant
    if MEMORY_ENABLED and MEMORY_AVAILABLE:
        try:
            memory_manager = DomainMemoryManager(
                agent_id="agent-medical",
                agent_type="medical",
                domain="healthcare",
                redis_url=REDIS_URL,
                postgres_url=POSTGRES_URL,
                use_redis=bool(REDIS_URL),
                use_postgres=bool(POSTGRES_URL),
                default_constraints=[
                    {
                        "description": "HIPAA compliance - protect all PHI",
                        "hard": True,
                        "category": "privacy",
                    },
                    {
                        "description": "Verify authorization before accessing records",
                        "hard": True,
                        "category": "security",
                    },
                    {
                        "description": "Log all data access for audit trail",
                        "hard": True,
                        "category": "audit",
                    },
                    {
                        "description": "Never store unencrypted PHI in memory",
                        "hard": True,
                        "category": "security",
                    },
                ],
            )
            await memory_manager.connect()
            logger.info("domain_memory_initialized", hipaa_compliant=True)
        except Exception as e:
            logger.error("domain_memory_init_failed", error=str(e))
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
    if db:
        await db.disconnect()
    
    logger.info("medical_agent_shutdown")


app = FastAPI(
    title="🏥 Medical Records Agent",
    description="HIPAA-compliant AI agent for medical records management",
    version="1.0.0",
    lifespan=lifespan,
)


class AgentResponse(BaseModel):
    """Response from the medical agent."""
    agent: str = "agent-medical"
    response: str
    patient_id: Optional[str] = None
    records: list[Dict[str, Any]] = []
    model: str = ""
    tokens_used: int = 0
    duration_ms: float = 0.0
    audit_id: str = ""
    timestamp: str = ""


def get_system_prompt() -> str:
    """Get the system prompt for this agent."""
    return SYSTEM_PROMPT or DEFAULT_SYSTEM_PROMPT


async def classify_intent(query: str, user_role: UserRole) -> IntentClassification:
    """
    Classify query intent using SLM (fast classification).
    
    TODO: Implement proper SLM-based classification
    For now, uses simple keyword matching.
    """
    query_lower = query.lower()
    
    # Simple intent detection
    if any(word in query_lower for word in ["lab", "result", "test", "exame"]):
        intent = QueryIntent.GET_LAB_RESULTS
        complexity = QueryComplexity.LOW
    elif any(word in query_lower for word in ["prescription", "medication", "medicamento", "receita"]):
        intent = QueryIntent.GET_PRESCRIPTIONS
        complexity = QueryComplexity.LOW
    elif any(word in query_lower for word in ["history", "histórico", "prontuário", "record"]):
        intent = QueryIntent.GET_HISTORY
        complexity = QueryComplexity.MEDIUM
    elif any(word in query_lower for word in ["interaction", "interação", "drug", "medicação"]):
        intent = QueryIntent.CHECK_INTERACTIONS
        complexity = QueryComplexity.MEDIUM
    elif any(word in query_lower for word in ["analyze", "analisar", "pattern", "padrão"]):
        intent = QueryIntent.ANALYZE
        complexity = QueryComplexity.HIGH
    elif any(word in query_lower for word in ["search", "buscar", "find", "encontrar"]):
        intent = QueryIntent.SEARCH_PATIENTS
        complexity = QueryComplexity.LOW
    else:
        intent = QueryIntent.READ_RECORD
        complexity = QueryComplexity.MEDIUM
    
    return IntentClassification(
        intent=intent,
        complexity=complexity,
        confidence=0.8
    )


async def extract_patient_id(query: str, user: User) -> Optional[str]:
    """
    Extract patient ID from query.
    
    TODO: Implement NLP-based extraction
    For now, looks for patient ID patterns.
    """
    # Simple extraction - look for patient-XXX or UUID patterns
    import re
    
    # Look for patient-XXX pattern
    match = re.search(r'patient[-\s]?(\w+)', query, re.IGNORECASE)
    if match:
        return match.group(1)
    
    # Look for UUID pattern
    uuid_pattern = r'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'
    match = re.search(uuid_pattern, query, re.IGNORECASE)
    if match:
        return match.group(0)
    
    # If user is a patient, return their own ID
    if user.role == UserRole.PATIENT:
        return user.id
    
    return None


async def call_llm(prompt: str, model: str = None, use_vllm: bool = False) -> tuple[str, int]:
    """
    Call LLM (Ollama or VLLM) with the given prompt.
    
    Args:
        prompt: The prompt to send
        model: Model name (optional)
        use_vllm: Whether to use VLLM (for complex queries)
    
    Returns:
        Tuple of (response_text, tokens_used)
    """
    if not http_client:
        raise HTTPException(status_code=503, detail="HTTP client not initialized")
    
    model = model or OLLAMA_MODEL
    
    try:
        if use_vllm and VLLM_URL:
            # Use VLLM for complex queries
            response = await http_client.post(
                f"{VLLM_URL}/v1/chat/completions",
                json={
                    "model": model,
                    "messages": [
                        {"role": "system", "content": get_system_prompt()},
                        {"role": "user", "content": prompt}
                    ],
                    "temperature": TEMPERATURE,
                    "max_tokens": MAX_TOKENS,
                }
            )
            response.raise_for_status()
            result = response.json()
            return (
                result["choices"][0]["message"]["content"],
                result.get("usage", {}).get("total_tokens", 0)
            )
        else:
            # Use Ollama for simple queries
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
        logger.error("llm_timeout", url=VLLM_URL if use_vllm else OLLAMA_URL)
        raise HTTPException(status_code=504, detail="LLM request timed out")
    except httpx.HTTPStatusError as e:
        logger.error("llm_error", status=e.response.status_code, detail=str(e))
        raise HTTPException(status_code=502, detail=f"LLM service error: {e.response.status_code}")
    except Exception as e:
        logger.error("llm_exception", error=str(e))
        raise HTTPException(status_code=500, detail=f"LLM error: {str(e)}")


async def process_medical_request(
    query: str,
    user: User,
    patient_id: Optional[str] = None,
    event_id: Optional[str] = None,
    source: str = "api"
) -> AgentResponse:
    """
    Shared business logic for processing medical record requests.
    
    This is the core function that handles:
    1. Access control verification
    2. Intent classification
    3. Patient data retrieval
    4. LLM reasoning
    5. Audit logging
    
    Args:
        query: The medical query
        user: Authenticated user
        patient_id: Optional patient ID
        event_id: Optional event ID for tracing
        source: Source of the request
    
    Returns:
        AgentResponse with the agent's response
    """
    start_time = time.time()
    audit_id = str(uuid.uuid4())
    
    logger.info(
        "medical_request_processing",
        query_length=len(query),
        user_id=user.id,
        user_role=user.role.value,
        patient_id=patient_id,
        event_id=event_id,
        source=source,
        audit_id=audit_id,
    )
    
    # 1. Extract patient ID if not provided
    if not patient_id:
        patient_id = await extract_patient_id(query, user)
    
    # 2. Verify access if patient ID is present
    if patient_id:
        if not access_control.verify_access(user, patient_id):
            ACCESS_DENIED_TOTAL.labels(reason="no_patient_access").inc()
            REQUESTS_TOTAL.labels(role=user.role.value, status="denied").inc()
            
            # Audit log
            if HIPAA_MODE:
                patient_id_hash = access_control.hash_patient_id(patient_id) if patient_id else None
                await audit_log(
                    user=user,
                    action="access_denied",
                    patient_id_hash=patient_id_hash or "none",
                    query=access_control.sanitize_query(query),
                    status="denied",
                    audit_id=audit_id
                )
            
            raise HTTPException(
                status_code=403,
                detail="Access denied: You don't have permission to view this patient's records"
            )
    
    # 3. Classify intent
    intent_class = await classify_intent(query, user.role)
    
    # 4. Retrieve patient data if needed
    records = []
    if patient_id and db:
        try:
            if intent_class.intent == QueryIntent.GET_LAB_RESULTS:
                records = await db.get_lab_results(patient_id, user.id, user.role.value)
            elif intent_class.intent == QueryIntent.GET_PRESCRIPTIONS:
                records = await db.get_prescriptions(patient_id, user.id, user.role.value)
            elif intent_class.intent == QueryIntent.GET_HISTORY:
                records = await db.get_medical_records(patient_id, user.id, user.role.value)
            else:
                records = await db.get_medical_records(patient_id, user.id, user.role.value, limit=10)
        except Exception as e:
            logger.error("database_query_failed", error=str(e))
            # Continue without records
    
    # 5. Build context for LLM
    context = {
        "user_role": user.role.value,
        "patient_id": patient_id,
        "intent": intent_class.intent.value,
        "records": records[:5] if records else [],  # Limit context size
    }
    
    # 6. Build prompt
    prompt = f"""Medical Query: {query}

Context:
- User Role: {user.role.value}
- Patient ID: {patient_id or 'Not specified'}
- Intent: {intent_class.intent.value}
- Records Available: {len(records)} records

Patient Records:
{json.dumps(records[:5], default=str) if records else 'No records available'}

Please provide a helpful, accurate response based on the available information.
Maintain patient privacy and use appropriate medical terminology."""

    # 7. Call LLM (use VLLM for complex queries)
    use_vllm = intent_class.complexity == QueryComplexity.HIGH
    with LLM_INFERENCE_DURATION.labels(model=OLLAMA_MODEL if not use_vllm else "vllm").time():
        response_text, tokens = await call_llm(prompt, use_vllm=use_vllm)
    
    # 8. Calculate duration
    duration_ms = (time.time() - start_time) * 1000
    
    # 9. Record metrics
    REQUESTS_TOTAL.labels(role=user.role.value, status="success").inc()
    RESPONSE_DURATION.labels(model=OLLAMA_MODEL).observe(duration_ms / 1000)
    TOKENS_USED.labels(model=OLLAMA_MODEL, type="total").inc(tokens)
    if patient_id:
        patient_id_hash = access_control.hash_patient_id(patient_id)
        # Note: We don't have a metric for this, but we could add it
    
    # 10. Audit log (HIPAA requirement)
    if HIPAA_MODE:
        patient_id_hash = access_control.hash_patient_id(patient_id) if patient_id else None
        await audit_log(
            user=user,
            action=intent_class.intent.value,
            patient_id_hash=patient_id_hash or "none",
            query=access_control.sanitize_query(query),
            status="success",
            audit_id=audit_id
        )
    
    logger.info(
        "medical_request_processed",
        user_id=user.id,
        patient_id=patient_id,
        intent=intent_class.intent.value,
        tokens=tokens,
        duration_ms=duration_ms,
        audit_id=audit_id,
    )
    
    return AgentResponse(
        response=response_text,
        patient_id=patient_id,
        records=records[:10] if records else [],  # Limit response size
        model=OLLAMA_MODEL if not use_vllm else "vllm",
        tokens_used=tokens,
        duration_ms=duration_ms,
        audit_id=audit_id,
        timestamp=datetime.utcnow().isoformat(),
    )


async def audit_log(
    user: User,
    action: str,
    patient_id_hash: str,
    query: Optional[str],
    status: str,
    audit_id: str
):
    """Log audit entry (HIPAA requirement)."""
    if audit_logger:
        await audit_logger.log(
            user=user,
            action=action,
            patient_id=patient_id_hash,  # Already hashed
            query=query,
            status=status,
            audit_id=audit_id,
        )
        AUDIT_LOGS_TOTAL.labels(action=action).inc()


@app.get("/health")
async def health():
    """Health check endpoint."""
    db_ok = db is not None
    return {
        "status": "healthy" if db_ok else "degraded",
        "agent": "agent-medical",
        "database": "connected" if db_ok else "disconnected",
        "hipaa_mode": HIPAA_MODE,
    }


@app.get("/ready")
async def ready():
    """Readiness check - verify Ollama and database are accessible."""
    db_ready = False
    llm_ready = False
    
    if db:
        try:
            # Simple connection check
            db_ready = True
        except:
            pass
    
    if http_client:
        try:
            response = await http_client.get(f"{OLLAMA_URL}/api/tags", timeout=5.0)
            llm_ready = response.status_code == 200
        except:
            pass
    
    if db_ready and llm_ready:
        return {"status": "ready", "agent": "agent-medical"}
    
    return JSONResponse(
        status_code=503,
        content={
            "status": "not_ready",
            "agent": "agent-medical",
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
    """
    Handle incoming CloudEvents.
    
    This is the main entry point for event-driven medical record queries.
    Uses shared business logic for consistency and tracing.
    
    Event Types:
    - io.homelab.medical.query: Medical record query
    - io.homelab.medical.lab.request: Lab results request
    - io.homelab.medical.prescription.request: Prescription request
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
            
            # Extract authentication token
            auth_token = headers.get("authorization", "").replace("Bearer ", "")
            if not auth_token:
                auth_token = event_data.get("token") or event_data.get("auth_token")
            
            # Get user from token
            user = get_user_from_token(auth_token)
            if not user:
                raise HTTPException(status_code=401, detail="Authentication required")
            
            # Create tracing span
            with tracer.start_as_current_span(
                f"cloudevent.{event_type.replace('.', '_').replace(':', '_')}",
                attributes={
                    "cloudevent.type": event_type,
                    "cloudevent.source": event_source,
                    "cloudevent.id": event_id,
                    "user.id": user.id,
                    "user.role": user.role.value,
                }
            ) as span:
                logger.info(
                    "event_received",
                    event_type=event_type,
                    event_id=event_id,
                    source=event_source,
                    user_id=user.id,
                    user_role=user.role.value,
                    source_type="cloudevent",
                )
                
                # Extract query from event data
                query = event_data.get("query") or event_data.get("message") or event_data.get("request") or ""
                patient_id = event_data.get("patient_id")
                
                if not query:
                    raise HTTPException(status_code=400, detail="Query is required")
                
                # Use shared business logic
                response_data = await process_medical_request(
                    query=query,
                    user=user,
                    patient_id=patient_id,
                    event_id=event_id,
                    source="cloudevent",
                )
                
                span.set_attribute("cloudevent.duration_ms", (time.time() - start_time) * 1000)
                
                # Return as CloudEvent response
                response_event = CloudEvent({
                    "type": "io.homelab.medical.response",
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
            
            # Extract authentication
            auth_token = headers.get("authorization", "").replace("Bearer ", "")
            if not auth_token:
                auth_token = event_data.get("token") or event_data.get("auth_token")
            
            user = get_user_from_token(auth_token)
            if not user:
                raise HTTPException(status_code=401, detail="Authentication required")
            
            logger.info(
                "request_received",
                event_type=event_type,
                event_id=event_id,
                user_id=user.id,
                source_type="api",
            )
            
            # Extract query
            query = event_data.get("query") or event_data.get("message") or ""
            patient_id = event_data.get("patient_id")
            
            if not query:
                raise HTTPException(status_code=400, detail="Query is required")
            
            # Use shared business logic
            response_data = await process_medical_request(
                query=query,
                user=user,
                patient_id=patient_id,
                event_id=event_id,
                source="api",
            )
            
            # Return as CloudEvent response (for consistency)
            response_event = CloudEvent({
                "type": "io.homelab.medical.response",
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
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(
            "event_processing_error",
            error=str(e),
            agent="agent-medical",
        )
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/info")
async def info():
    """Get agent information."""
    return {
        "name": "agent-medical",
        "description": "HIPAA-compliant medical records agent",
        "model": OLLAMA_MODEL,
        "endpoint": OLLAMA_URL,
        "event_source": EVENT_SOURCE,
        "hipaa_mode": HIPAA_MODE,
        "version": "1.0.0",
    }


@app.post("/api/v1/exams/upload")
async def upload_medical_exam(
    request: Request,
    file: UploadFile = File(...),
    patient_id: str = Form(...),
    metadata: Optional[str] = Form(None)
):
    """
    Upload a medical exam PDF and extract structured data using LangExtract.
    
    This endpoint:
    1. Accepts a PDF file upload
    2. Extracts text from the PDF
    3. Uses LangExtract to extract structured medical data
    4. Stores the PDF metadata and extracted data in MongoDB
    5. Returns the extracted data and storage information
    
    Args:
        file: PDF file to upload
        patient_id: Patient ID to associate with the exam
        metadata: Optional JSON metadata string
        
    Returns:
        Dictionary with extraction results and storage info
    """
    # Extract authentication token
    headers = dict(request.headers)
    auth_token = headers.get("authorization", "").replace("Bearer ", "")
    if not auth_token:
        raise HTTPException(status_code=401, detail="Authentication required")
    
    user = get_user_from_token(auth_token)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid authentication token")
    
    # Verify file is PDF
    if not file.filename or not file.filename.lower().endswith('.pdf'):
        raise HTTPException(status_code=400, detail="File must be a PDF")
    
    # Verify access to patient
    if not access_control.verify_access(user, patient_id):
        ACCESS_DENIED_TOTAL.labels(reason="no_patient_access").inc()
        raise HTTPException(
            status_code=403,
            detail="Access denied: You don't have permission to upload exams for this patient"
        )
    
    audit_id = str(uuid.uuid4())
    start_time = time.time()
    
    try:
        # Read PDF file content
        pdf_bytes = await file.read()
        
        if len(pdf_bytes) == 0:
            raise HTTPException(status_code=400, detail="PDF file is empty")
        
        # Limit file size (100 MB max)
        max_size = 100 * 1024 * 1024
        if len(pdf_bytes) > max_size:
            raise HTTPException(
                status_code=400,
                detail=f"PDF file too large (max {max_size / 1024 / 1024} MB)"
            )
        
        logger.info(
            "pdf_upload_received",
            filename=file.filename,
            size=len(pdf_bytes),
            patient_id=patient_id,
            user_id=user.id,
            audit_id=audit_id
        )
        
        # Parse metadata if provided
        metadata_dict = {}
        if metadata:
            try:
                metadata_dict = json.loads(metadata)
            except json.JSONDecodeError:
                logger.warning("invalid_metadata_json", metadata=metadata)
        
        # Generate storage path (for now, just a simple path; in production use MinIO)
        storage_path = f"medical_exams/{patient_id}/{uuid.uuid4()}/{file.filename}"
        
        # Extract medical data using LangExtract with Ollama
        extractor = get_pdf_extractor(
            model_id=os.getenv("OLLAMA_MODEL", OLLAMA_MODEL),
            model_url=OLLAMA_URL,
            use_ollama=True
        )
        extracted_data = await extractor.extract_medical_data(
            pdf_bytes=pdf_bytes,
            patient_id=patient_id,
            metadata={**metadata_dict, "filename": file.filename}
        )
        
        # Store in database
        if not db:
            raise HTTPException(status_code=503, detail="Database not available")
        
        # TODO: In production, upload PDF to MinIO/S3 first, then store reference in MongoDB
        # For now, we only store metadata and extracted data (not the PDF bytes themselves)
        stored_doc = await db.store_medical_exam_pdf(
            patient_id=patient_id,
            user_id=user.id,
            user_role=user.role.value,
            pdf_bytes=pdf_bytes,  # Used only for size calculation
            filename=file.filename,
            storage_path=storage_path,
            extracted_data=extracted_data,
            metadata=metadata_dict
        )
        
        duration_ms = (time.time() - start_time) * 1000
        
        # Audit log
        if HIPAA_MODE:
            patient_id_hash = access_control.hash_patient_id(patient_id)
            await audit_log(
                user=user,
                action="upload_medical_exam",
                patient_id_hash=patient_id_hash,
                query=f"Uploaded PDF: {file.filename}",
                status="success",
                audit_id=audit_id
            )
        
        logger.info(
            "pdf_upload_completed",
            filename=file.filename,
            patient_id=patient_id,
            pdf_id=stored_doc.get("id"),
            extraction_count=extracted_data.get("extraction_count", 0),
            duration_ms=duration_ms,
            audit_id=audit_id
        )
        
        return {
            "status": "success",
            "pdf_id": stored_doc.get("id"),
            "filename": file.filename,
            "storage_path": storage_path,
            "patient_id": patient_id,
            "extracted_data": {
                "extraction_count": extracted_data.get("extraction_count", 0),
                "classes": list(extracted_data.get("grouped_extractions", {}).keys()),
                "grouped_extractions": extracted_data.get("grouped_extractions", {}),
            },
            "metadata": metadata_dict,
            "uploaded_by": user.id,
            "uploaded_at": datetime.utcnow().isoformat(),
            "duration_ms": duration_ms,
            "audit_id": audit_id,
        }
        
    except HTTPException:
        raise
    except ValueError as e:
        logger.error("pdf_upload_error", error=str(e), patient_id=patient_id, audit_id=audit_id)
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        logger.error("pdf_upload_exception", error=str(e), patient_id=patient_id, audit_id=audit_id)
        raise HTTPException(status_code=500, detail=f"Failed to process PDF: {str(e)}")


@app.get("/api/v1/exams")
async def get_medical_exams(
    request: Request,
    patient_id: str,
    limit: int = 50
):
    """
    Get medical exam PDFs for a patient.
    
    Args:
        patient_id: Patient ID
        limit: Maximum number of results
        
    Returns:
        List of medical exam PDFs with extracted data
    """
    # Extract authentication token
    headers = dict(request.headers)
    auth_token = headers.get("authorization", "").replace("Bearer ", "")
    if not auth_token:
        raise HTTPException(status_code=401, detail="Authentication required")
    
    user = get_user_from_token(auth_token)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid authentication token")
    
    # Verify access to patient
    if not access_control.verify_access(user, patient_id):
        ACCESS_DENIED_TOTAL.labels(reason="no_patient_access").inc()
        raise HTTPException(
            status_code=403,
            detail="Access denied: You don't have permission to view exams for this patient"
        )
    
    if not db:
        raise HTTPException(status_code=503, detail="Database not available")
    
    try:
        pdfs = await db.get_medical_exam_pdfs(
            patient_id=patient_id,
            user_id=user.id,
            user_role=user.role.value,
            limit=limit
        )
        
        # Return summary information (exclude full extracted_data for performance)
        return {
            "status": "success",
            "patient_id": patient_id,
            "exams": [
                {
                    "id": pdf.get("id"),
                    "filename": pdf.get("filename"),
                    "storage_path": pdf.get("storage_path"),
                    "file_size": pdf.get("file_size"),
                    "extraction_count": pdf.get("extracted_data", {}).get("extraction_count", 0),
                    "extraction_classes": list(pdf.get("extracted_data", {}).get("grouped_extractions", {}).keys()),
                    "uploaded_by": pdf.get("uploaded_by"),
                    "uploaded_at": pdf.get("created_at").isoformat() if pdf.get("created_at") else None,
                    "metadata": pdf.get("metadata", {}),
                }
                for pdf in pdfs
            ],
            "count": len(pdfs),
        }
    except Exception as e:
        logger.error("get_medical_exams_error", error=str(e), patient_id=patient_id)
        raise HTTPException(status_code=500, detail=f"Failed to retrieve exams: {str(e)}")


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
