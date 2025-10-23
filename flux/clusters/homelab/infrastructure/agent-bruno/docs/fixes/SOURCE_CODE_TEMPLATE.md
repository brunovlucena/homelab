# Source Code Template - Agent Bruno

**Status**: 🔴 P0 - CRITICAL BLOCKER  
**Timeline**: Week 1-2  
**Blocks**: Application deployment, testing

---

## 📁 Directory Structure

Create this structure in `src/`:

```
src/
├── main.py                    # FastAPI application entry point
├── __init__.py
├── api/
│   ├── __init__.py
│   ├── dependencies.py        # Dependency injection
│   ├── middleware/
│   │   ├── __init__.py
│   │   ├── logging.py        # Request/response logging
│   │   ├── tracing.py        # OpenTelemetry tracing
│   │   └── error_handler.py  # Global error handling
│   └── routes/
│       ├── __init__.py
│       ├── health.py         # Health check endpoints
│       ├── chat.py           # Chat API endpoints
│       └── mcp.py            # MCP server endpoints
├── core/
│   ├── __init__.py
│   ├── config.py             # Configuration management
│   ├── logging.py            # Logfire configuration
│   ├── ollama.py             # Ollama client wrapper
│   ├── lancedb.py            # LanceDB client wrapper
│   └── redis.py              # Redis client wrapper
├── models/
│   ├── __init__.py
│   ├── chat.py               # Chat request/response models
│   └── schemas.py            # Pydantic schemas
├── rag/
│   ├── __init__.py
│   ├── retriever.py          # RAG retrieval logic
│   ├── embeddings.py         # Embedding generation
│   └── reranker.py           # Result reranking
├── services/
│   ├── __init__.py
│   └── chat_service.py       # Business logic
└── tests/
    ├── __init__.py
    ├── conftest.py           # Pytest fixtures
    ├── unit/
    │   ├── __init__.py
    │   ├── test_health.py
    │   └── test_config.py
    └── integration/
        ├── __init__.py
        └── test_chat_api.py
```

---

## 1️⃣ FastAPI Application Entry Point

`src/main.py`:

```python
"""
Agent Bruno - AI Assistant API
FastAPI application with Logfire observability
"""

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from logfire import configure_logfire
import uvicorn

from api.routes import health, chat, mcp
from api.middleware.logging import LoggingMiddleware
from api.middleware.tracing import TracingMiddleware
from api.middleware.error_handler import ErrorHandlerMiddleware
from core.config import settings

# Configure Logfire observability
configure_logfire(
    token=settings.logfire_token,
    service_name="agent-bruno",
    environment=settings.environment,
)

# Create FastAPI app
app = FastAPI(
    title="Agent Bruno API",
    description="AI Assistant with RAG and MCP capabilities",
    version="1.0.0",
    docs_url="/docs" if settings.environment != "production" else None,
    redoc_url="/redoc" if settings.environment != "production" else None,
)

# Add middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
app.add_middleware(ErrorHandlerMiddleware)
app.add_middleware(TracingMiddleware)
app.add_middleware(LoggingMiddleware)

# Include routers
app.include_router(health.router, prefix="/api/v1", tags=["health"])
app.include_router(chat.router, prefix="/api/v1", tags=["chat"])
app.include_router(mcp.router, prefix="/api/v1/mcp", tags=["mcp"])

@app.on_event("startup")
async def startup_event():
    """Initialize services on startup"""
    # TODO: Initialize LanceDB, Ollama, Redis connections
    pass

@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on shutdown"""
    # TODO: Close connections
    pass

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8080,
        reload=settings.environment == "development",
        log_level=settings.log_level.lower(),
    )
```

---

## 2️⃣ Configuration Management

`src/core/config.py`:

```python
"""
Configuration management using Pydantic BaseSettings
Loads from environment variables
"""

from pydantic_settings import BaseSettings
from pydantic import Field
from typing import List

class Settings(BaseSettings):
    """Application settings"""
    
    # Application
    environment: str = Field(default="development", env="ENVIRONMENT")
    log_level: str = Field(default="INFO", env="LOG_LEVEL")
    
    # API
    api_host: str = Field(default="0.0.0.0", env="API_HOST")
    api_port: int = Field(default=8080, env="API_PORT")
    cors_origins: List[str] = Field(
        default=["http://localhost:3000"],
        env="CORS_ORIGINS"
    )
    
    # LanceDB
    lancedb_path: str = Field(default="/data/lancedb", env="LANCEDB_PATH")
    lancedb_storage_mode: str = Field(default="local", env="LANCEDB_STORAGE_MODE")
    
    # Ollama
    ollama_base_url: str = Field(
        default="http://192.168.0.16:11434",
        env="OLLAMA_BASE_URL"
    )
    ollama_model: str = Field(default="llama3.1:8b", env="OLLAMA_MODEL")
    
    # Redis
    redis_host: str = Field(default="localhost", env="REDIS_HOST")
    redis_port: int = Field(default=6379, env="REDIS_PORT")
    redis_password: str = Field(default="", env="REDIS_PASSWORD")
    
    # Observability
    logfire_token: str = Field(default="", env="LOGFIRE_TOKEN")
    otel_exporter_otlp_endpoint: str = Field(
        default="http://tempo:4317",
        env="OTEL_EXPORTER_OTLP_ENDPOINT"
    )
    
    class Config:
        env_file = ".env"
        case_sensitive = False

# Create global settings instance
settings = Settings()
```

---

## 3️⃣ Health Check Endpoints

`src/api/routes/health.py`:

```python
"""
Health check endpoints for Kubernetes liveness/readiness probes
"""

from fastapi import APIRouter, status
from pydantic import BaseModel
import time

router = APIRouter()

# Application start time
start_time = time.time()

class HealthResponse(BaseModel):
    """Health check response"""
    status: str
    uptime_seconds: float
    version: str = "1.0.0"

@router.get(
    "/health",
    response_model=HealthResponse,
    status_code=status.HTTP_200_OK,
    summary="Liveness probe",
    description="Returns healthy if the application is running"
)
async def health():
    """Kubernetes liveness probe endpoint"""
    return HealthResponse(
        status="healthy",
        uptime_seconds=time.time() - start_time
    )

@router.get(
    "/ready",
    response_model=HealthResponse,
    status_code=status.HTTP_200_OK,
    summary="Readiness probe",
    description="Returns ready if the application can serve traffic"
)
async def ready():
    """Kubernetes readiness probe endpoint"""
    # TODO: Check LanceDB, Ollama, Redis connections
    return HealthResponse(
        status="ready",
        uptime_seconds=time.time() - start_time
    )
```

---

## 4️⃣ Chat API Endpoints

`src/api/routes/chat.py`:

```python
"""
Chat API endpoints
"""

from fastapi import APIRouter, HTTPException, Depends
from pydantic import BaseModel
from typing import List, Optional
import logfire

router = APIRouter()

class ChatMessage(BaseModel):
    """Chat message model"""
    role: str  # "user" or "assistant"
    content: str

class ChatRequest(BaseModel):
    """Chat request model"""
    message: str
    session_id: Optional[str] = None
    conversation_history: Optional[List[ChatMessage]] = None

class ChatResponse(BaseModel):
    """Chat response model"""
    response: str
    session_id: str
    sources: Optional[List[str]] = None

@router.post(
    "/chat",
    response_model=ChatResponse,
    summary="Send chat message",
    description="Send a message and get AI response with RAG"
)
async def chat(request: ChatRequest):
    """Chat endpoint with RAG"""
    
    with logfire.span("chat_request", message=request.message):
        try:
            # TODO: Implement actual chat logic
            # 1. Retrieve relevant context from LanceDB
            # 2. Generate response with Ollama
            # 3. Store in Redis session
            
            response = ChatResponse(
                response=f"Echo: {request.message}",
                session_id=request.session_id or "test-session",
                sources=[]
            )
            
            logfire.info("chat_response", response=response.response)
            return response
            
        except Exception as e:
            logfire.error("chat_error", error=str(e))
            raise HTTPException(
                status_code=500,
                detail=f"Chat processing failed: {str(e)}"
            )
```

---

## 5️⃣ MCP Server Endpoints

`src/api/routes/mcp.py`:

```python
"""
MCP (Model Context Protocol) server endpoints
"""

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import List, Dict, Any

router = APIRouter()

class MCPToolRequest(BaseModel):
    """MCP tool request"""
    tool_name: str
    parameters: Dict[str, Any]

class MCPToolResponse(BaseModel):
    """MCP tool response"""
    result: Any
    status: str

@router.post(
    "/tools/execute",
    response_model=MCPToolResponse,
    summary="Execute MCP tool",
    description="Execute an MCP tool with parameters"
)
async def execute_tool(request: MCPToolRequest):
    """Execute MCP tool"""
    # TODO: Implement actual MCP tool execution
    return MCPToolResponse(
        result={"message": "MCP tool execution not implemented"},
        status="pending"
    )

@router.get(
    "/tools",
    summary="List available MCP tools",
    description="Get list of available MCP tools"
)
async def list_tools():
    """List available MCP tools"""
    # TODO: Return actual MCP tools
    return {
        "tools": [
            {"name": "search", "description": "Search knowledge base"},
            {"name": "summarize", "description": "Summarize text"},
        ]
    }
```

---

## 6️⃣ Requirements Files

`requirements.txt`:

```
# Core
fastapi==0.109.0
uvicorn[standard]==0.27.0
pydantic==2.5.0
pydantic-settings==2.1.0

# Observability
logfire==0.23.0
opentelemetry-api==1.22.0
opentelemetry-sdk==1.22.0

# AI/ML
ollama==0.1.6
lancedb==0.4.0
sentence-transformers==2.3.0

# Storage
redis==5.0.1
httpx==0.26.0

# Utilities
python-dotenv==1.0.0
python-multipart==0.0.6
```

`requirements-dev.txt`:

```
# Testing
pytest==7.4.3
pytest-asyncio==0.23.2
pytest-cov==4.1.0
httpx==0.26.0

# Linting
ruff==0.1.11
black==23.12.1
mypy==1.8.0

# Type stubs
types-redis==4.6.0
```

---

## 7️⃣ Pytest Configuration

`pytest.ini`:

```ini
[pytest]
testpaths = tests
python_files = test_*.py
python_classes = Test*
python_functions = test_*
markers =
    unit: Unit tests
    integration: Integration tests
    slow: Slow-running tests
    asyncio: Async tests
addopts = 
    --verbose
    --cov=src
    --cov-report=html
    --cov-report=term-missing
    --cov-fail-under=70
    -m "not slow"
asyncio_mode = auto
```

`src/tests/conftest.py`:

```python
"""
Pytest fixtures and configuration
"""

import pytest
from fastapi.testclient import TestClient
from src.main import app

@pytest.fixture
def client():
    """Test client fixture"""
    return TestClient(app)

@pytest.fixture
def mock_ollama_response():
    """Mock Ollama response"""
    return {
        "model": "llama3.1:8b",
        "response": "Test response",
        "done": True
    }
```

---

## 8️⃣ Basic Tests

`src/tests/unit/test_health.py`:

```python
"""
Unit tests for health endpoints
"""

import pytest
from fastapi.testclient import TestClient

def test_health_endpoint(client: TestClient):
    """Test health endpoint returns 200"""
    response = client.get("/api/v1/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "healthy"
    assert "uptime_seconds" in data

def test_ready_endpoint(client: TestClient):
    """Test ready endpoint returns 200"""
    response = client.get("/api/v1/ready")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "ready"
```

`src/tests/integration/test_chat_api.py`:

```python
"""
Integration tests for chat API
"""

import pytest
from fastapi.testclient import TestClient

@pytest.mark.integration
def test_chat_endpoint(client: TestClient):
    """Test chat endpoint"""
    response = client.post(
        "/api/v1/chat",
        json={
            "message": "Hello, world!",
            "session_id": "test-session"
        }
    )
    assert response.status_code == 200
    data = response.json()
    assert "response" in data
    assert data["session_id"] == "test-session"
```

---

## ✅ Implementation Checklist

### Week 1: Foundation
- [ ] Create directory structure
- [ ] Implement `main.py` (FastAPI app)
- [ ] Implement `core/config.py` (configuration)
- [ ] Implement `api/routes/health.py` (health checks)
- [ ] Create `requirements.txt` and `requirements-dev.txt`
- [ ] Create `pytest.ini` and `conftest.py`
- [ ] Write basic unit tests
- [ ] Verify tests pass locally

### Week 2: Core Features
- [ ] Implement `api/routes/chat.py` (chat endpoints)
- [ ] Implement `core/ollama.py` (Ollama client)
- [ ] Implement `core/lancedb.py` (LanceDB client)
- [ ] Implement `core/redis.py` (Redis client)
- [ ] Implement `rag/retriever.py` (RAG logic)
- [ ] Write integration tests
- [ ] Deploy to dev environment

---

## 🔗 Related Documentation

- [DEVOPS_UNBLOCK_PLAN.md](../DEVOPS_UNBLOCK_PLAN.md) - Overall unblock plan
- [DOCKERFILE_TEMPLATE.md](./DOCKERFILE_TEMPLATE.md) - Container template
- [GITHUB_ACTIONS_TEMPLATES.md](./GITHUB_ACTIONS_TEMPLATES.md) - CI/CD templates

---

**Status**: 🔴 NOT IMPLEMENTED  
**Next Step**: Create `src/` directory and implement files  
**Owner**: Development Team  
**Timeline**: Week 1-2

---

