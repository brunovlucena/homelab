# DevOps Unblock Plan - Agent Bruno Infrastructure

**Status**: 🔴 BLOCKED → 🟢 UNBLOCKING  
**Review Score**: 8.5/10 (B+)  
**Blocker**: No implementation (documentation-only project)  
**Timeline**: 2-4 weeks to production-ready  
**Last Updated**: October 23, 2025

---

## 🎯 Executive Summary

The Agent Bruno project has **world-class documentation and design** but is currently BLOCKED because:

1. ✅ **Documentation**: Comprehensive, industry-leading (10/10)
2. ✅ **Architecture**: Excellent GitOps design with Flux + Flagger (9/10)
3. ✅ **Observability**: Best-in-class LGTM stack (10/10)
4. 🔴 **Implementation**: Missing (empty `src/` and `k8s/` directories) (2/10)
5. 🔴 **CI/CD**: No automation pipelines (0/10)
6. 🔴 **Testing**: Not running automatically (2/10)

**Verdict**: 🔴 BLOCKED - "World-class docs, no implementation"

---

## 🚨 Critical Blockers

### Blocker 1: No Source Code Implementation
**Current State**: Empty `src/` directory  
**Impact**: Cannot deploy anything  
**Priority**: P0 - CRITICAL

```bash
repos/homelab/flux/clusters/homelab/infrastructure/agent-bruno/
├── src/          # ❌ EMPTY
├── k8s/          # ❌ EMPTY (some in production-fixes/)
└── docs/         # ✅ COMPLETE (comprehensive)
```

### Blocker 2: No CI/CD Pipeline
**Current State**: No GitHub Actions workflows  
**Impact**: Manual builds, no automation  
**Priority**: P0 - CRITICAL

**Missing**:
- `.github/workflows/` directory
- Automated testing
- Image building
- Security scanning
- GitOps deployment

### Blocker 3: No Dockerfile
**Current State**: No container image definition  
**Impact**: Cannot build images  
**Priority**: P0 - CRITICAL

### Blocker 4: No Deployment Manifests in k8s/
**Current State**: Some manifests in `production-fixes/` but not in main `k8s/`  
**Impact**: Incomplete deployment structure  
**Priority**: P0 - CRITICAL

---

## 📋 Unblocking Roadmap

### Phase 0: Foundation (Week 1) - UNBLOCKS PROJECT ✅

#### Day 1-2: Create Dockerfile & Container Image
**Status**: ⏳ TODO  
**Files**:
- `Dockerfile` (multi-stage build)
- `.dockerignore`
- `docker-compose.yml` (local dev)

**Implementation**:
```dockerfile
# Dockerfile
FROM python:3.11-slim as builder
WORKDIR /app
RUN apt-get update && apt-get install -y build-essential
COPY requirements.txt .
RUN pip install --user --no-cache-dir -r requirements.txt

FROM python:3.11-slim
RUN useradd -m -u 1000 agent && \
    mkdir -p /app /data/lancedb && \
    chown -R agent:agent /app /data
WORKDIR /app
COPY --from=builder /root/.local /home/agent/.local
COPY --chown=agent:agent src/ .
USER agent
ENV PATH=/home/agent/.local/bin:$PATH
HEALTHCHECK --interval=30s --timeout=3s \
  CMD python -c "import requests; requests.get('http://localhost:8080/health')"
EXPOSE 8080
CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8080"]
```

**Deliverables**:
- ✅ Multi-stage build (smaller image)
- ✅ Non-root user (security)
- ✅ Health check (K8s readiness)
- ✅ Layer caching optimization

---

#### Day 3-5: GitHub Actions CI/CD Pipeline
**Status**: ⏳ TODO  
**Files**:
- `.github/workflows/ci.yml`
- `.github/workflows/cd.yml`
- `.github/workflows/security.yml`

**Implementation**:
```yaml
# .github/workflows/ci.yml
name: CI Pipeline

on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [main, develop]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}/agent-bruno

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      
      - name: Install dependencies
        run: |
          pip install -r requirements.txt
          pip install -r requirements-dev.txt
      
      - name: Run tests
        run: pytest tests/ -v --cov --cov-report=xml
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.xml

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          severity: 'CRITICAL,HIGH'
          exit-code: '1'
      
      - name: Run Semgrep
        uses: returntocorp/semgrep-action@v1
        with:
          config: p/security-audit p/python

  build:
    needs: [test, security]
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4
      
      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=sha,prefix={{branch}}-
            type=semver,pattern={{version}}
      
      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
      
      - name: Generate SBOM
        uses: anchore/sbom-action@v0
        with:
          image: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
```

**Deliverables**:
- ✅ Automated testing
- ✅ Security scanning (Trivy, Semgrep)
- ✅ Container image building
- ✅ SBOM generation
- ✅ Coverage reporting

---

#### Day 6-7: GitOps Deployment Automation
**Status**: ⏳ TODO  
**Files**:
- `.github/workflows/deploy.yml`

**Implementation**:
```yaml
# .github/workflows/deploy.yml
name: CD - Deploy to Kubernetes

on:
  push:
    branches: [main, develop]
    tags: ['v*']

jobs:
  deploy:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
        with:
          token: ${{ secrets.FLUX_GITHUB_TOKEN }}
      
      - name: Update image tag
        run: |
          NEW_TAG="main-${{ github.sha }}"
          sed -i "s|image: .*agent-bruno:.*|image: ghcr.io/${{ github.repository }}/agent-bruno:$NEW_TAG|" \
            flux/clusters/homelab/infrastructure/agent-bruno/k8s/deployment.yaml
      
      - name: Commit and push
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add flux/clusters/homelab/infrastructure/agent-bruno/k8s/
          git commit -m "chore: update agent-bruno image to ${{ github.sha }}"
          git push
      
      - name: Notify deployment
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: 'Deployment started for agent-bruno:${{ github.sha }}'
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

**Deliverables**:
- ✅ Automated deployment to Kubernetes
- ✅ GitOps manifest updates
- ✅ Slack notifications
- ✅ Multi-environment support

---

### Phase 1: Core Implementation (Week 2-3)

#### Organize Kubernetes Manifests
**Status**: 🟡 PARTIAL (exists in `production-fixes/`)  
**Action**: Move to proper structure

**Current**:
```
production-fixes/
├── p0-statefulset/
│   ├── lancedb-statefulset.yaml  ✅
│   └── migration-job.yaml        ✅
└── p0-velero/
    ├── velero-install.yaml       ✅
    └── backup-schedules.yaml     ✅
```

**Target Structure**:
```
k8s/
├── base/
│   ├── namespace.yaml
│   ├── statefulset.yaml          # From production-fixes/p0-statefulset/
│   ├── service.yaml
│   ├── configmap.yaml
│   └── kustomization.yaml
├── overlays/
│   ├── dev/
│   │   └── kustomization.yaml
│   ├── staging/
│   │   └── kustomization.yaml
│   └── prod/
│       └── kustomization.yaml
└── backup/
    ├── velero-install.yaml       # From production-fixes/p0-velero/
    └── backup-schedules.yaml     # From production-fixes/p0-velero/
```

**Action Items**:
1. ✅ Create `k8s/base/` directory structure
2. ✅ Move StatefulSet from `production-fixes/p0-statefulset/`
3. ✅ Move Velero from `production-fixes/p0-velero/`
4. ✅ Create Kustomize overlays for environments
5. ✅ Update Flux to point to new structure

---

#### Create Source Code Structure
**Status**: 🔴 MISSING  
**Action**: Implement basic FastAPI application

**Target Structure**:
```
src/
├── main.py                    # FastAPI app entry point
├── api/
│   ├── __init__.py
│   ├── routes/
│   │   ├── __init__.py
│   │   ├── health.py
│   │   ├── chat.py
│   │   └── mcp.py
│   └── middleware/
│       ├── __init__.py
│       ├── logging.py
│       └── tracing.py
├── core/
│   ├── __init__.py
│   ├── config.py              # Environment config
│   ├── ollama.py              # Ollama client
│   └── lancedb.py             # LanceDB client
├── rag/
│   ├── __init__.py
│   ├── retriever.py
│   ├── embeddings.py
│   └── reranker.py
└── tests/
    ├── __init__.py
    ├── unit/
    │   └── test_health.py
    └── integration/
        └── test_chat.py
```

**Minimum Viable Implementation**:
```python
# src/main.py
from fastapi import FastAPI
from logfire import configure_logfire
import uvicorn

configure_logfire()
app = FastAPI(title="Agent Bruno API")

@app.get("/health")
async def health():
    return {"status": "healthy"}

@app.get("/ready")
async def ready():
    return {"status": "ready"}

@app.post("/api/v1/chat")
async def chat(message: str):
    # TODO: Implement actual chat logic
    return {"response": f"Echo: {message}"}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8080)
```

---

### Phase 2: Testing & Automation (Week 3-4)

#### Automated Testing Suite
**Files**:
- `src/tests/unit/test_*.py`
- `src/tests/integration/test_*.py`
- `pytest.ini`
- `conftest.py`

**Implementation**:
```python
# src/tests/unit/test_health.py
import pytest
from fastapi.testclient import TestClient
from main import app

client = TestClient(app)

def test_health_endpoint():
    response = client.get("/health")
    assert response.status_code == 200
    assert response.json() == {"status": "healthy"}

def test_ready_endpoint():
    response = client.get("/ready")
    assert response.status_code == 200
    assert response.json() == {"status": "ready"}
```

```yaml
# pytest.ini
[pytest]
testpaths = tests
python_files = test_*.py
python_classes = Test*
python_functions = test_*
markers =
    unit: Unit tests
    integration: Integration tests
    slow: Slow tests
addopts = 
    --verbose
    --cov=src
    --cov-report=html
    --cov-report=term
    -m "not slow"
```

---

#### Backup Automation
**Status**: ✅ DOCUMENTED (exists in `production-fixes/p0-velero/`)  
**Action**: Integrate into main k8s structure

**Already Implemented**:
- ✅ Velero HelmRelease
- ✅ Backup schedules (hourly)
- ✅ S3/Minio configuration

**Additional Requirements**:
```yaml
# k8s/backup/backup-schedules.yaml
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: agent-bruno-hourly
  namespace: velero
spec:
  schedule: "0 * * * *"  # Every hour
  template:
    includedNamespaces:
      - agent-bruno
    ttl: 720h  # 30 days
    storageLocation: default
    volumeSnapshotLocations:
      - default
    defaultVolumesToRestic: true  # Backup PVCs
    
---
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: agent-bruno-daily
  namespace: velero
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  template:
    includedNamespaces:
      - agent-bruno
    ttl: 2160h  # 90 days
    storageLocation: default
```

---

### Phase 3: Documentation Updates (Ongoing)

#### Update Review Documentation
**Files to Update**:

1. **REVIEW_INDEX.md** - Update DevOps status
```markdown
| **AI Senior DevOps Engineer** | 8.5/10 | B+ | 🟢 UNBLOCKED | CI/CD implemented, automation complete |
```

2. **DEVOPS_ENGINEER_REVIEW.md** - Add implementation section
```markdown
## Implementation Status Update (October 23, 2025)

### ✅ Completed
- [x] Dockerfile created (multi-stage build)
- [x] GitHub Actions CI/CD pipeline
- [x] Automated testing in CI
- [x] Security scanning (Trivy, Semgrep)
- [x] GitOps deployment automation
- [x] Kubernetes manifest organization
- [x] Backup automation (Velero)

### 🟡 In Progress
- [ ] Source code implementation (FastAPI)
- [ ] Multi-environment setup (dev/staging/prod)
- [ ] DORA metrics dashboard

### Production Readiness: 🟢 UNBLOCKED
The project now has full CI/CD automation and is ready for implementation phase.
```

3. **CICD_SETUP.md** - Update with actual workflows
4. **BACKUP_SETUP.md** - Reference production-fixes implementation

---

## 📊 Progress Tracking

### Completion Checklist

#### P0 - Critical (Week 1) ✅ UNBLOCKS PROJECT
- [ ] Create Dockerfile
- [ ] Create `.dockerignore`
- [ ] Create GitHub Actions CI workflow
- [ ] Create GitHub Actions CD workflow
- [ ] Create GitHub Actions security workflow
- [ ] Organize k8s/ directory structure
- [ ] Move StatefulSet manifests
- [ ] Move Velero manifests
- [ ] Create minimal FastAPI app (health endpoints)
- [ ] Create basic tests

#### P1 - High Priority (Week 2-3)
- [ ] Implement chat API endpoint
- [ ] Implement LanceDB integration
- [ ] Implement Ollama integration
- [ ] Create multi-environment overlays
- [ ] Setup DORA metrics
- [ ] Add deployment notifications
- [ ] Create operational scripts

#### P2 - Medium Priority (Week 3-4)
- [ ] Implement RAG pipeline
- [ ] Add integration tests
- [ ] Setup load testing
- [ ] Create Tilt configuration (local dev)
- [ ] Add CONTRIBUTING.md
- [ ] Add CHANGELOG.md

---

## 🎯 Success Criteria

### DevOps Unblocking Criteria
To change status from 🔴 BLOCKED → 🟢 UNBLOCKED:

1. ✅ Dockerfile exists and builds successfully
2. ✅ GitHub Actions CI/CD pipeline runs on every PR/push
3. ✅ Tests run automatically in CI
4. ✅ Container images build and push automatically
5. ✅ Security scanning runs automatically
6. ✅ GitOps deployment automation works
7. ✅ K8s manifests organized properly
8. ✅ Basic FastAPI app deployable

### Production-Ready Criteria
To achieve full production readiness:

1. ✅ All P0 items complete
2. ✅ Source code implementation (chat API)
3. ✅ Multi-environment deployment (dev/staging/prod)
4. ✅ Automated backups working
5. ✅ DORA metrics tracking
6. ✅ Integration tests passing
7. ✅ Load testing completed
8. ✅ Security scanning passing

---

## 📈 Timeline

```
Week 1: Foundation & Automation
├── Day 1-2: Dockerfile + Docker Compose
├── Day 3-5: GitHub Actions CI/CD
└── Day 6-7: GitOps Deployment

Week 2: Implementation
├── Organize k8s/ structure
├── Implement FastAPI app
├── Create tests
└── Setup monitoring

Week 3: Multi-Environment
├── Dev environment
├── Staging environment
├── Prod environment
└── Environment promotion

Week 4: Production Hardening
├── DORA metrics
├── Load testing
├── Documentation updates
└── Production deployment

Total: 4 weeks to production-ready ✅
```

---

## 🔗 Related Documentation

### Implementation Guides
- [CICD_SETUP.md](./CICD_SETUP.md) - CI/CD pipeline setup
- [BACKUP_SETUP.md](./BACKUP_SETUP.md) - Backup automation
- [LANCEDB_PERSISTENCE.md](./LANCEDB_PERSISTENCE.md) - Data persistence

### Review Documentation
- [DEVOPS_ENGINEER_REVIEW.md](./DEVOPS_ENGINEER_REVIEW.md) - Full DevOps review
- [REVIEW_INDEX.md](./REVIEW_INDEX.md) - All reviews summary
- [SRE_REVIEW.md](./SRE_REVIEW.md) - SRE assessment

### Architecture
- [ARCHITECTURE.md](./ARCHITECTURE.md) - System architecture
- [OBSERVABILITY.md](./OBSERVABILITY.md) - Monitoring stack
- [TESTING.md](./TESTING.md) - Testing strategy

---

## 🎬 Quick Start

### For DevOps Engineers
```bash
# 1. Create foundation
cd repos/homelab/flux/clusters/homelab/infrastructure/agent-bruno
touch Dockerfile .dockerignore docker-compose.yml

# 2. Create CI/CD workflows
mkdir -p .github/workflows
# Copy workflows from this document

# 3. Organize k8s manifests
mkdir -p k8s/{base,overlays/{dev,staging,prod},backup}
mv production-fixes/p0-statefulset/* k8s/base/
mv production-fixes/p0-velero/* k8s/backup/

# 4. Create source code
mkdir -p src/{api,core,rag,tests}
# Implement minimal FastAPI app

# 5. Commit and push
git add .
git commit -m "feat: implement DevOps automation (unblock project)"
git push
```

---

**Status**: 🟡 IN PROGRESS  
**Next Update**: After Phase 0 completion  
**Owner**: Bruno Lucena  
**Last Updated**: October 23, 2025

---

