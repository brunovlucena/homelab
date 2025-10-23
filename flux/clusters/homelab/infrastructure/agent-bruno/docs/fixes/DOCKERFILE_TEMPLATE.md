# Dockerfile Template - Agent Bruno

**Status**: 🔴 P0 - CRITICAL BLOCKER  
**Timeline**: Day 1-2  
**Blocks**: CI/CD pipeline, container deployment

---

## 📋 Template

Create this file as `/Dockerfile` in the project root:

```dockerfile
# ============================================================================
# Agent Bruno - Multi-Stage Container Build
# ============================================================================
# Stage 1: Builder - Install dependencies
# Stage 2: Runtime - Minimal production image
# 
# Security: Non-root user, minimal base image, health checks
# Performance: Layer caching, multi-stage build
# ============================================================================

# ----------------------------------------------------------------------------
# Stage 1: Builder
# ----------------------------------------------------------------------------
FROM python:3.11-slim as builder

LABEL maintainer="Bruno Lucena <bruno@example.com>" \
      description="Agent Bruno AI Assistant - Builder Stage" \
      version="1.0.0"

WORKDIR /app

# Install build dependencies (compilers, headers)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        build-essential \
        gcc \
        g++ \
        git \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first (layer caching)
COPY requirements.txt .

# Install Python dependencies to user directory
RUN pip install --user --no-cache-dir --upgrade pip && \
    pip install --user --no-cache-dir -r requirements.txt

# ----------------------------------------------------------------------------
# Stage 2: Runtime
# ----------------------------------------------------------------------------
FROM python:3.11-slim

LABEL maintainer="Bruno Lucena <bruno@example.com>" \
      description="Agent Bruno AI Assistant" \
      version="1.0.0"

# Install runtime dependencies only (minimal)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        curl \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN useradd -m -u 1000 -s /bin/bash agent && \
    mkdir -p /app /data/lancedb /tmp && \
    chown -R agent:agent /app /data /tmp

WORKDIR /app

# Copy Python dependencies from builder stage
COPY --from=builder --chown=agent:agent /root/.local /home/agent/.local

# Copy application code
COPY --chown=agent:agent src/ ./src/
COPY --chown=agent:agent requirements.txt .

# Switch to non-root user
USER agent

# Add user Python packages to PATH
ENV PATH=/home/agent/.local/bin:$PATH \
    PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    LANCEDB_PATH=/data/lancedb

# Health check (Kubernetes liveness/readiness)
HEALTHCHECK --interval=30s \
            --timeout=5s \
            --start-period=10s \
            --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Expose application port
EXPOSE 8080

# Expose metrics port (Prometheus)
EXPOSE 9090

# Run application
CMD ["python", "-m", "uvicorn", "src.main:app", \
     "--host", "0.0.0.0", \
     "--port", "8080", \
     "--workers", "1", \
     "--log-level", "info"]
```

---

## 📋 .dockerignore Template

Create this file as `/.dockerignore`:

```gitignore
# Version control
.git
.gitignore
.github

# Python
__pycache__
*.py[cod]
*$py.class
*.so
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
*.egg-info/
.installed.cfg
*.egg
MANIFEST

# Virtual environments
venv/
ENV/
env/

# Testing
.pytest_cache/
.coverage
htmlcov/
.tox/
.hypothesis/

# IDEs
.vscode/
.idea/
*.swp
*.swo
*~

# Documentation
docs/
*.md
!README.md

# Kubernetes
k8s/
flux/
production-fixes/

# CI/CD
.github/

# Local development
.env
.env.local
docker-compose.yml
Makefile

# Data
data/
*.db
*.sqlite
```

---

## 📋 docker-compose.yml Template

Create this file as `/docker-compose.yml` for local development:

```yaml
# ============================================================================
# Agent Bruno - Local Development Environment
# ============================================================================
# Services: agent-bruno, redis, minio
# Usage: docker-compose up -d
# ============================================================================

version: '3.8'

services:
  # --------------------------------------------------------------------------
  # Agent Bruno API
  # --------------------------------------------------------------------------
  agent-bruno:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: agent-bruno-dev
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      # LanceDB
      - LANCEDB_PATH=/data/lancedb
      - LANCEDB_STORAGE_MODE=local
      
      # Redis
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      
      # MinIO (S3-compatible)
      - MINIO_ENDPOINT=minio:9000
      - MINIO_ACCESS_KEY=minioadmin
      - MINIO_SECRET_KEY=minioadmin
      
      # Ollama
      - OLLAMA_BASE_URL=http://192.168.0.16:11434
      
      # Observability
      - LOG_LEVEL=debug
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://tempo:4317
    volumes:
      - ./src:/app/src
      - lancedb-data:/data/lancedb
    depends_on:
      - redis
      - minio
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
    networks:
      - agent-bruno

  # --------------------------------------------------------------------------
  # Redis (Session & Cache)
  # --------------------------------------------------------------------------
  redis:
    image: redis:7-alpine
    container_name: agent-bruno-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3
    networks:
      - agent-bruno

  # --------------------------------------------------------------------------
  # MinIO (S3-compatible storage)
  # --------------------------------------------------------------------------
  minio:
    image: minio/minio:latest
    container_name: agent-bruno-minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=minioadmin
      - MINIO_ROOT_PASSWORD=minioadmin
    volumes:
      - minio-data:/data
    command: server /data --console-address ":9001"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 5s
      retries: 3
    networks:
      - agent-bruno

# ----------------------------------------------------------------------------
# Volumes
# ----------------------------------------------------------------------------
volumes:
  lancedb-data:
    driver: local
  redis-data:
    driver: local
  minio-data:
    driver: local

# ----------------------------------------------------------------------------
# Networks
# ----------------------------------------------------------------------------
networks:
  agent-bruno:
    driver: bridge
```

---

## ✅ Validation

### Build Test
```bash
# Build the image
docker build -t agent-bruno:test .

# Verify image size (should be <1GB)
docker images agent-bruno:test

# Check layers
docker history agent-bruno:test

# Run security scan
docker scout cves agent-bruno:test
```

### Run Test
```bash
# Run container
docker run -d \
  --name agent-bruno-test \
  -p 8080:8080 \
  -e LANCEDB_PATH=/data/lancedb \
  agent-bruno:test

# Check health
curl http://localhost:8080/health

# Check logs
docker logs agent-bruno-test

# Cleanup
docker stop agent-bruno-test
docker rm agent-bruno-test
```

### Docker Compose Test
```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f agent-bruno

# Test endpoints
curl http://localhost:8080/health
curl http://localhost:8080/ready

# Cleanup
docker-compose down -v
```

---

## 🔒 Security Best Practices

### ✅ Implemented
- [x] Multi-stage build (minimal runtime image)
- [x] Non-root user (UID 1000)
- [x] Minimal base image (python:3.11-slim)
- [x] No secrets in image
- [x] Health checks
- [x] .dockerignore (exclude sensitive files)
- [x] Layer caching optimization

### 📋 Additional Recommendations
- [ ] Sign images with Cosign (CI/CD)
- [ ] Generate SBOM (CI/CD)
- [ ] Scan for vulnerabilities (CI/CD)
- [ ] Use specific image tags (not :latest)

---

## 🔗 Related Files

- [CICD_SETUP.md](../CICD_SETUP.md) - CI/CD pipeline (uses this Dockerfile)
- [DEVOPS_UNBLOCK_PLAN.md](../DEVOPS_UNBLOCK_PLAN.md) - Overall unblock plan
- [BACKUP_SETUP.md](../BACKUP_SETUP.md) - Backup strategy

---

**Status**: 🔴 NOT IMPLEMENTED  
**Next Step**: Create Dockerfile in project root  
**Owner**: DevOps Team  
**Timeline**: Day 1-2

---

