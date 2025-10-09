# 🏗️ Architecture - Homepage System

## 📋 Overview

The homepage is built as a modern, cloud-native application with microservices architecture, focusing on scalability, resilience, and observability.

## 🎯 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Cloudflare CDN (Optional)                 │
│                  SSL/TLS, DDoS Protection                    │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Kubernetes Cluster                         │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐ │
│  │   Frontend   │    │   API (Go)   │    │  PostgreSQL  │ │
│  │   (React)    │◄──►│   (Proxy)    │◄──►│  (Database)  │ │
│  │              │    │              │    │              │ │
│  └──────────────┘    └──────┬───────┘    └──────────────┘ │
│                             │                               │
│                             ▼                               │
│              ┌──────────────────────────┐                  │
│              │     Agent-SRE Service    │                  │
│              │  ┌────────────────────┐  │                  │
│              │  │   SRE Agent        │  │                  │
│              │  │   (Port 8080)      │  │                  │
│              │  └─────────┬──────────┘  │                  │
│              │            │              │                  │
│              │            ▼              │                  │
│              │  ┌────────────────────┐  │                  │
│              │  │   MCP Server       │  │                  │
│              │  │   (Port 30120)     │  │                  │
│              │  └─────────┬──────────┘  │                  │
│              └────────────┼─────────────┘                  │
│                           │                                 │
│                           ▼                                 │
│              ┌────────────────────────┐                    │
│              │   Ollama/LLM Service   │                    │
│              │   (External)           │                    │
│              └────────────────────────┘                    │
│                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐│
│  │    Redis     │    │    MinIO     │    │  Prometheus  ││
│  │   (Cache)    │    │  (Storage)   │    │ (Monitoring) ││
│  └──────────────┘    └──────────────┘    └──────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## 🔄 Data Flow

### 1. User Interaction

```
User Browser
    ↓ HTTPS
Frontend (React)
    ↓ /api/v1/*
API Gateway (Go/Gin)
    ↓ PostgreSQL
Database
```

### 2. Chatbot Communication

```
User Message
    ↓
Frontend (chatbot.ts)
    ↓ Try MCP first
API Proxy (/api/v1/agent-sre/mcp/chat)
    ↓
Agent-SRE Service
    ↓
MCP Server
    ↓
Ollama/LLM
    ↓
Response
```

**Fallback Path:**
```
MCP Fails
    ↓
API Proxy (/api/v1/agent-sre/chat)
    ↓
Agent-SRE Service (Direct)
    ↓
Ollama/LLM
    ↓
Response
```

## 🧩 Components

### Frontend (React + TypeScript)

**Technology:**
- React 18
- TypeScript
- Vite (build tool)
- Axios (HTTP client)

**Key Files:**
- `src/services/chatbot.ts` - Chatbot service with MCP/Direct modes
- `src/services/api.ts` - API client
- `src/components/Chatbot.tsx` - Chat UI component

**Deployment:**
- Nginx container
- Static file serving
- API proxy configuration

### Backend API (Go)

**Technology:**
- Go 1.23
- Gin framework
- GORM (ORM)
- Redis (caching)

**Key Components:**
- `handlers/agent_sre.go` - Agent-SRE proxy handler
- `router/router.go` - Route definitions
- `config/config.go` - Configuration management

**Endpoints:**
- `/api/v1/projects` - Project CRUD
- `/api/v1/skills` - Skills management
- `/api/v1/agent-sre/*` - Chatbot proxy

**Features:**
- Request/response logging
- CORS handling
- Compression (gzip)
- Health checks

### Database (PostgreSQL 15)

**Schema:**
- `projects` - Project information
- `skills` - Skills and proficiency
- `experiences` - Work experience
- `content` - Dynamic content

**Features:**
- Migrations via init job
- Performance indexes
- Backup strategy

### Agent-SRE Service

**Technology:**
- Python
- FastAPI
- MCP Protocol
- Ollama client

**Components:**
- SRE Agent - Main service (port 8080)
- MCP Server - Protocol handler (port 30120)

**Communication Modes:**
1. **MCP Mode** - Structured protocol
2. **Direct Mode** - Simple HTTP

### Cache Layer (Redis)

**Usage:**
- Session storage
- API response caching (future)
- Rate limiting (future)

**Configuration:**
- Persistence enabled
- Memory limit: 512MB

### Storage (MinIO)

**Usage:**
- Asset storage (images, files)
- Object storage interface
- S3-compatible API

**Buckets:**
- `homepage-assets` - Public assets

## 🔐 Security Architecture

### Network Security

```
Internet
    ↓ HTTPS (443)
Cloudflare CDN
    ↓ TLS
Ingress Controller
    ↓ HTTP (internal)
Services (ClusterIP)
```

**Security Layers:**
1. Cloudflare - DDoS protection, WAF
2. Kubernetes Network Policies
3. Service mesh (future)

### Authentication & Authorization

**Current:**
- No authentication (public site)

**Future:**
- JWT tokens for admin API
- OAuth2 for user login

### Secrets Management

**Method:** Sealed Secrets

**Secrets:**
- Database passwords
- Redis passwords
- MinIO credentials
- Cloudflare tokens

**Location:** Kubernetes secrets in `homepage` namespace

## 📊 Observability

### Metrics (Prometheus)

**Collected Metrics:**
- HTTP request duration
- Request count by endpoint
- Error rates
- Resource usage (CPU, memory)

**Endpoints:**
- API: `/metrics`
- Frontend: `/metrics`
- Database: postgres-exporter

### Logging

**Stack:**
- Application logs → stdout
- Kubernetes logs → Loki
- Grafana dashboard

**Log Levels:**
- ERROR - Critical issues
- WARN - Warnings
- INFO - Important events
- DEBUG - Detailed debugging

### Tracing (Future)

**Planned:**
- Jaeger for distributed tracing
- OpenTelemetry integration

## 🚀 Deployment Architecture

### Development Environment

```
Docker Compose
├── postgres (local DB)
├── redis (local cache)
├── api (hot reload)
└── frontend (dev server)
```

**Access:**
- Frontend: http://localhost:3000
- API: http://localhost:8080
- Agent-SRE: http://localhost:31081

### Production Environment (Kubernetes)

```
Helm Chart
├── Deployments
│   ├── frontend (2 replicas)
│   ├── api (2 replicas)
│   ├── postgres (1 replica)
│   └── redis (1 replica)
├── Services
│   ├── frontend (ClusterIP)
│   ├── api (ClusterIP)
│   └── postgres (ClusterIP)
├── Ingress
│   └── HTTP/HTTPS routing
└── ConfigMaps/Secrets
    ├── Configuration
    └── Sensitive data
```

**Scaling:**
- HPA enabled for frontend/API
- Min: 1 replica
- Max: 3 replicas (Mac Studio)
- Target: 80% CPU

## 🔄 CI/CD Pipeline

### GitHub Actions Workflows

**1. homepage-tests.yml**
- Trigger: Push/PR
- Tests: Backend + Frontend
- Duration: ~5 minutes

**2. homepage-pr-check.yml**
- Trigger: PR only
- Checks: Code quality, security
- Duration: ~3 minutes

**3. homepage-nightly-tests.yml**
- Trigger: Daily 2 AM UTC
- Tests: Comprehensive suite
- Duration: ~15 minutes

### Build Process

```
Code Push
    ↓
GitHub Actions
    ↓
├─ Run Tests
├─ Build Images
├─ Security Scan
└─ Push to Registry
    ↓
ArgoCD/Flux (future)
    ↓
Deploy to Kubernetes
```

## 📈 Scaling Considerations

### Horizontal Scaling

**Current:**
- Frontend: 1-3 replicas
- API: 1-3 replicas
- Database: 1 replica (primary)

**Future:**
- Database: Read replicas
- Redis: Cluster mode
- Load balancer: Multiple regions

### Vertical Scaling

**Resource Limits:**
- API: 2 CPU, 2Gi memory
- Frontend: 1 CPU, 1Gi memory
- Database: 2 CPU, 4Gi memory

### Performance Optimization

**Implemented:**
- Gzip compression
- Static asset caching
- Database indexes

**Planned:**
- CDN caching
- Redis caching
- Query optimization

## 🔧 Configuration Management

### Environment Variables

**API:**
- `DATABASE_URL` - PostgreSQL connection
- `REDIS_URL` - Redis connection
- `AGENT_SRE_URL` - Agent service URL
- `CORS_ORIGIN` - Allowed origins

**Frontend:**
- `VITE_API_URL` - API base URL

### ConfigMaps

**api-config:**
- Non-sensitive configuration
- Feature flags

**frontend-nginx-config:**
- Nginx configuration
- Proxy settings

## 🎯 Design Principles

1. **Microservices** - Loosely coupled components
2. **Stateless** - Easy to scale horizontally
3. **12-Factor App** - Cloud-native best practices
4. **Observability** - Metrics, logs, traces
5. **Security** - Defense in depth
6. **Resilience** - Graceful degradation
7. **Automation** - CI/CD, auto-scaling

## 📚 Technology Decisions

### Why Go for API?

- High performance
- Great concurrency
- Strong typing
- Excellent tooling

### Why React for Frontend?

- Component reusability
- Large ecosystem
- TypeScript support
- Developer experience

### Why PostgreSQL?

- Reliability
- ACID compliance
- Rich feature set
- Good performance

### Why Agent-SRE Integration?

- AI-powered assistance
- SRE knowledge base
- Kubernetes expertise
- Flexible architecture

---

**Architecture Version:** 1.0.0  
**Last Updated:** 2025-10-08  
**Status:** Production Ready

