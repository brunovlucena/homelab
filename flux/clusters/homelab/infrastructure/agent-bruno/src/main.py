"""
🌐 Agent Bruno - FastAPI Server

Main FastAPI application with chat endpoints, memory management, and observability.
"""

import logging
import os
from contextlib import asynccontextmanager
from fastapi import FastAPI, Request, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel
from typing import Optional, Dict, Any
from prometheus_client import Counter, Histogram, Gauge, generate_latest
from fastapi.responses import Response

from .agent.core import AgentBruno
from .memory.manager import MemoryManager

# Configure logging
logging.basicConfig(
    level=os.getenv("LOG_LEVEL", "INFO"),
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

# Prometheus metrics
requests_total = Counter(
    "bruno_requests_total",
    "Total requests",
    ["method", "endpoint", "status"]
)
request_duration = Histogram(
    "bruno_request_duration_seconds",
    "Request duration",
    ["method", "endpoint"]
)
memory_operations = Counter(
    "bruno_memory_operations_total",
    "Memory operations",
    ["operation", "status"]
)
active_sessions = Gauge(
    "bruno_active_sessions",
    "Active sessions"
)

# Global agent instance
agent: Optional[AgentBruno] = None
memory_manager: Optional[MemoryManager] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifespan context manager for startup and shutdown"""
    global agent, memory_manager
    
    # Startup
    logger.info("🚀 Starting Agent Bruno...")
    
    # Load configuration
    redis_url = os.getenv("REDIS_URL", "redis://redis:6379")
    mongodb_url = os.getenv("MONGODB_URL", "mongodb://mongodb:27017")
    mongodb_db = os.getenv("MONGODB_DB", "agent_bruno")
    session_ttl = int(os.getenv("SESSION_TTL", "86400"))
    ollama_url = os.getenv("OLLAMA_URL", "http://192.168.0.16:11434")
    
    # Initialize memory manager
    memory_manager = MemoryManager(
        redis_url=redis_url,
        mongodb_url=mongodb_url,
        mongodb_db=mongodb_db,
        session_ttl=session_ttl
    )
    await memory_manager.connect()
    
    # Initialize agent
    agent = AgentBruno(
        memory_manager=memory_manager,
        ollama_url=ollama_url
    )
    
    logger.info("✅ Agent Bruno started successfully")
    
    yield
    
    # Shutdown
    logger.info("👋 Shutting down Agent Bruno...")
    if memory_manager:
        await memory_manager.disconnect()
    logger.info("✅ Shutdown complete")


# Create FastAPI app
app = FastAPI(
    title="🤖 Agent Bruno",
    description="AI assistant with homepage knowledge and IP-based memory",
    version="0.1.0",
    lifespan=lifespan
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configured for internal service
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# Pydantic models
class ChatRequest(BaseModel):
    message: str
    context: Optional[Dict[str, Any]] = None


class ChatResponse(BaseModel):
    success: bool
    response: str
    model: Optional[str] = None
    context_used: Optional[Dict[str, Any]] = None
    error: Optional[str] = None


class MemoryStats(BaseModel):
    ip: str
    recent_messages: int
    total_conversations: int
    has_history: bool


class SystemStats(BaseModel):
    active_sessions: int
    total_conversations: int
    unique_ips: int


# Middleware for metrics
@app.middleware("http")
async def metrics_middleware(request: Request, call_next):
    """Collect metrics for all requests"""
    import time
    
    start_time = time.time()
    response = await call_next(request)
    duration = time.time() - start_time
    
    # Record metrics
    requests_total.labels(
        method=request.method,
        endpoint=request.url.path,
        status=response.status_code
    ).inc()
    
    request_duration.labels(
        method=request.method,
        endpoint=request.url.path
    ).observe(duration)
    
    return response


def get_client_ip(request: Request) -> str:
    """Extract client IP from request"""
    # Check for X-Forwarded-For header (proxy)
    forwarded_for = request.headers.get("X-Forwarded-For")
    if forwarded_for:
        return forwarded_for.split(",")[0].strip()
    
    # Fall back to client host
    if request.client:
        return request.client.host
    
    return "unknown"


# Health check endpoints
@app.get("/health")
async def health():
    """Health check endpoint"""
    return {"status": "healthy", "service": "agent-bruno"}


@app.get("/ready")
async def ready():
    """Readiness check endpoint - allows service to run with degraded functionality"""
    if not memory_manager:
        raise HTTPException(status_code=503, detail="Memory manager not initialized")
    
    health = await memory_manager.health_check()
    
    # Service is ready even if stores are unavailable (degraded mode)
    status = "ready"
    if not health["overall"]:
        status = "degraded"
        logger.warning(f"⚠️  Service running in degraded mode: Redis={health['redis']}, MongoDB={health['mongodb']}")
    
    return {
        "status": status,
        "service": "agent-bruno",
        "health": health,
        "degraded": not health["overall"]
    }


@app.get("/status")
async def status():
    """Status endpoint - comprehensive service status"""
    if not memory_manager or not agent:
        return {
            "status": "initializing",
            "service": "agent-bruno",
            "ready": False
        }
    
    try:
        health = await memory_manager.health_check()
        
        return {
            "status": "healthy" if health["overall"] else "degraded",
            "service": "agent-bruno",
            "ready": True,
            "health": {
                "redis": health["redis"],
                "mongodb": health["mongodb"]
            },
            "degraded": not health["overall"]
        }
    except Exception as e:
        logger.error(f"❌ Status check error: {e}")
        return {
            "status": "error",
            "service": "agent-bruno",
            "ready": False,
            "error": str(e)
        }


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint"""
    return Response(content=generate_latest(), media_type="text/plain")


# Chat endpoints
@app.post("/chat", response_model=ChatResponse)
async def chat(request: ChatRequest, req: Request):
    """Chat with Agent Bruno"""
    if not agent:
        raise HTTPException(status_code=503, detail="Agent not initialized")
    
    ip = get_client_ip(req)
    
    try:
        result = await agent.chat(
            message=request.message,
            ip=ip,
            context=request.context
        )
        
        memory_operations.labels(operation="chat", status="success").inc()
        
        return ChatResponse(**result)
    
    except Exception as e:
        logger.error(f"❌ Chat error: {e}")
        memory_operations.labels(operation="chat", status="error").inc()
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/mcp/chat", response_model=ChatResponse)
async def mcp_chat(request: ChatRequest, req: Request):
    """Chat with Agent Bruno via MCP protocol (alias for compatibility)"""
    return await chat(request, req)


# Memory endpoints
@app.get("/memory/{ip}", response_model=MemoryStats)
async def get_memory(ip: str):
    """Get memory statistics for IP"""
    if not agent:
        raise HTTPException(status_code=503, detail="Agent not initialized")
    
    try:
        stats = await agent.get_memory_stats(ip)
        return MemoryStats(**stats)
    
    except Exception as e:
        logger.error(f"❌ Get memory error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/memory/{ip}/history")
async def get_memory_history(ip: str, limit: int = 50, skip: int = 0):
    """Get full conversation history for IP"""
    if not memory_manager:
        raise HTTPException(status_code=503, detail="Memory manager not initialized")
    
    try:
        history = await memory_manager.get_full_history(ip, limit, skip)
        return {"ip": ip, "history": history, "count": len(history)}
    
    except Exception as e:
        logger.error(f"❌ Get history error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.delete("/memory/{ip}")
async def clear_memory(ip: str):
    """Clear memory for IP"""
    if not agent:
        raise HTTPException(status_code=503, detail="Agent not initialized")
    
    try:
        await agent.clear_memory(ip)
        memory_operations.labels(operation="clear", status="success").inc()
        return {"status": "success", "message": f"Memory cleared for IP: {ip}"}
    
    except Exception as e:
        logger.error(f"❌ Clear memory error: {e}")
        memory_operations.labels(operation="clear", status="error").inc()
        raise HTTPException(status_code=500, detail=str(e))


# Knowledge endpoints
@app.get("/knowledge/summary")
async def knowledge_summary():
    """Get knowledge base summary"""
    if not agent:
        raise HTTPException(status_code=503, detail="Agent not initialized")
    
    return {"summary": agent.get_knowledge_summary()}


@app.get("/knowledge/search")
async def knowledge_search(q: str):
    """Search knowledge base"""
    if not agent:
        raise HTTPException(status_code=503, detail="Agent not initialized")
    
    results = agent.search_knowledge(q)
    return {"query": q, "results": results}


# System endpoints
@app.get("/stats", response_model=SystemStats)
async def system_stats():
    """Get system statistics"""
    if not memory_manager:
        raise HTTPException(status_code=503, detail="Memory manager not initialized")
    
    try:
        stats = await memory_manager.get_stats()
        
        # Update Prometheus gauge
        active_sessions.set(stats["active_sessions"])
        
        return SystemStats(
            active_sessions=stats["active_sessions"],
            total_conversations=stats["total_conversations"],
            unique_ips=stats["unique_ips"]
        )
    
    except Exception as e:
        logger.error(f"❌ Stats error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "service": "agent-bruno",
        "version": "0.1.0",
        "status": "running",
        "endpoints": {
            "chat": "/chat",
            "mcp_chat": "/mcp/chat",
            "health": "/health",
            "ready": "/ready",
            "status": "/status",
            "metrics": "/metrics",
            "memory": "/memory/{ip}",
            "knowledge": "/knowledge/summary"
        }
    }


if __name__ == "__main__":
    import uvicorn
    
    port = int(os.getenv("PORT", "8080"))
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=port,
        log_level="info"
    )

