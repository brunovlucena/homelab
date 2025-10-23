# DevOps Unblocking - Implementation Checklist

**Status**: 🟡 IN PROGRESS  
**Goal**: Change DevOps review from 🔴 BLOCKED → 🟢 UNBLOCKED  
**Timeline**: 2-4 weeks  
**Last Updated**: October 23, 2025

---

## 📊 Progress Tracker

### Overall Progress: 15% Complete

| Phase | Status | Progress | Timeline |
|-------|--------|----------|----------|
| **Phase 0: Foundation** | 🟡 In Progress | 20% | Week 1 |
| **Phase 1: Implementation** | ⏳ Not Started | 0% | Week 2-3 |
| **Phase 2: Testing** | ⏳ Not Started | 0% | Week 3-4 |
| **Phase 3: Documentation** | 🟡 In Progress | 40% | Ongoing |

---

## 🎯 Phase 0: Foundation (Week 1) - UNBLOCKS PROJECT

### Day 1-2: Container Foundation ⏳ 0/5

**Goal**: Create Dockerfile and container build infrastructure

- [ ] **Task 1.1**: Create `Dockerfile` in project root
  - Reference: [DOCKERFILE_TEMPLATE.md](./DOCKERFILE_TEMPLATE.md)
  - Location: `/Dockerfile`
  - Validation: `docker build -t agent-bruno:test .`
  
- [ ] **Task 1.2**: Create `.dockerignore`
  - Reference: [DOCKERFILE_TEMPLATE.md](./DOCKERFILE_TEMPLATE.md)
  - Location: `/.dockerignore`
  - Validation: Verify excluded files
  
- [ ] **Task 1.3**: Create `docker-compose.yml` for local dev
  - Reference: [DOCKERFILE_TEMPLATE.md](./DOCKERFILE_TEMPLATE.md)
  - Location: `/docker-compose.yml`
  - Validation: `docker-compose up -d`
  
- [ ] **Task 1.4**: Test container build
  - Command: `docker build -t agent-bruno:test .`
  - Expected: Build succeeds, image < 1GB
  - Validation: `docker images agent-bruno:test`
  
- [ ] **Task 1.5**: Test container run
  - Command: `docker run -d -p 8080:8080 agent-bruno:test`
  - Expected: Container starts, health check passes
  - Validation: `curl http://localhost:8080/health`

**Success Criteria**:
- ✅ Dockerfile exists and builds successfully
- ✅ Container image < 1GB
- ✅ Non-root user (UID 1000)
- ✅ Health check works
- ✅ docker-compose.yml starts all services

---

### Day 3-5: CI/CD Pipeline ⏳ 0/8

**Goal**: Implement automated testing and deployment

- [ ] **Task 2.1**: Create `.github/workflows/` directory
  - Location: `/.github/workflows/`
  - Command: `mkdir -p .github/workflows`
  
- [ ] **Task 2.2**: Create CI pipeline
  - Reference: [GITHUB_ACTIONS_TEMPLATES.md](./GITHUB_ACTIONS_TEMPLATES.md)
  - Location: `/.github/workflows/ci.yml`
  - Includes: Test → Security → Build → Push
  
- [ ] **Task 2.3**: Create CD pipeline
  - Reference: [GITHUB_ACTIONS_TEMPLATES.md](./GITHUB_ACTIONS_TEMPLATES.md)
  - Location: `/.github/workflows/cd.yml`
  - Includes: GitOps manifest updates
  
- [ ] **Task 2.4**: Create security pipeline
  - Reference: [GITHUB_ACTIONS_TEMPLATES.md](./GITHUB_ACTIONS_TEMPLATES.md)
  - Location: `/.github/workflows/security.yml`
  - Includes: Trivy, Semgrep, secret scanning
  
- [ ] **Task 2.5**: Create GitHub PAT
  - Purpose: Flux GitOps manifest updates
  - Permissions: `repo`, `packages:write`
  - Secret: `FLUX_GITHUB_TOKEN`
  
- [ ] **Task 2.6**: Configure GitHub secrets
  - `FLUX_GITHUB_TOKEN`: For GitOps updates
  - `SLACK_WEBHOOK`: For notifications (optional)
  - `SNYK_TOKEN`: For Snyk scanning (optional)
  
- [ ] **Task 2.7**: Test CI pipeline
  - Create test PR
  - Verify: Tests run, build succeeds, security scan runs
  - Expected: All jobs pass
  
- [ ] **Task 2.8**: Test CD pipeline
  - Merge to `develop` branch
  - Verify: Image pushed, GitOps manifest updated
  - Expected: Deployment to dev environment

**Success Criteria**:
- ✅ CI runs on every PR/push
- ✅ Tests execute automatically
- ✅ Security scanning works
- ✅ Container images build and push
- ✅ GitOps manifests update automatically
- ✅ Notifications sent (if configured)

---

### Day 6-7: Source Code Foundation ⏳ 0/6

**Goal**: Create minimal FastAPI application

- [ ] **Task 3.1**: Create `src/` directory structure
  - Reference: [SOURCE_CODE_TEMPLATE.md](./SOURCE_CODE_TEMPLATE.md)
  - Command: `mkdir -p src/{api/{routes,middleware},core,models,rag,services,tests/{unit,integration}}`
  
- [ ] **Task 3.2**: Implement `main.py`
  - Reference: [SOURCE_CODE_TEMPLATE.md](./SOURCE_CODE_TEMPLATE.md)
  - Location: `/src/main.py`
  - Features: FastAPI app, Logfire, middleware
  
- [ ] **Task 3.3**: Implement `core/config.py`
  - Reference: [SOURCE_CODE_TEMPLATE.md](./SOURCE_CODE_TEMPLATE.md)
  - Location: `/src/core/config.py`
  - Features: Pydantic settings, env vars
  
- [ ] **Task 3.4**: Implement health endpoints
  - Reference: [SOURCE_CODE_TEMPLATE.md](./SOURCE_CODE_TEMPLATE.md)
  - Location: `/src/api/routes/health.py`
  - Endpoints: `/health`, `/ready`
  
- [ ] **Task 3.5**: Create requirements files
  - Reference: [SOURCE_CODE_TEMPLATE.md](./SOURCE_CODE_TEMPLATE.md)
  - Files: `requirements.txt`, `requirements-dev.txt`
  - Includes: FastAPI, Logfire, Ollama, LanceDB
  
- [ ] **Task 3.6**: Create basic tests
  - Reference: [SOURCE_CODE_TEMPLATE.md](./SOURCE_CODE_TEMPLATE.md)
  - Files: `pytest.ini`, `conftest.py`, `test_health.py`
  - Validation: `pytest tests/unit/ -v`

**Success Criteria**:
- ✅ `src/` directory structure created
- ✅ FastAPI app runs: `python src/main.py`
- ✅ Health endpoint responds: `curl http://localhost:8080/api/v1/health`
- ✅ Tests pass: `pytest tests/ -v`
- ✅ Docker build includes source code

---

## 🎯 Phase 1: Core Implementation (Week 2-3)

### Week 2: Kubernetes Manifests ⏳ 0/7

**Goal**: Organize deployment manifests

- [ ] **Task 4.1**: Create `k8s/base/` directory
  - Command: `mkdir -p k8s/{base,overlays/{dev,staging,prod},backup}`
  - Structure: Kustomize base + overlays
  
- [ ] **Task 4.2**: Move StatefulSet manifests
  - From: `production-fixes/p0-statefulset/lancedb-statefulset.yaml`
  - To: `k8s/base/statefulset.yaml`
  - Update: Image references, labels
  
- [ ] **Task 4.3**: Create `k8s/base/service.yaml`
  - Type: ClusterIP
  - Ports: 8080 (HTTP), 9090 (metrics)
  
- [ ] **Task 4.4**: Create `k8s/base/configmap.yaml`
  - Config: Ollama URL, Redis host, log level
  - Source: From `core/config.py` defaults
  
- [ ] **Task 4.5**: Create `k8s/base/kustomization.yaml`
  - Resources: All base manifests
  - Common labels: app, version, managed-by
  
- [ ] **Task 4.6**: Create environment overlays
  - `k8s/overlays/dev/kustomization.yaml`
  - `k8s/overlays/staging/kustomization.yaml`
  - `k8s/overlays/prod/kustomization.yaml`
  
- [ ] **Task 4.7**: Move Velero manifests
  - From: `production-fixes/p0-velero/`
  - To: `k8s/backup/`
  - Files: `velero-install.yaml`, `backup-schedules.yaml`

**Success Criteria**:
- ✅ K8s manifests organized in Kustomize structure
- ✅ Base + overlays validate: `kustomize build k8s/overlays/dev`
- ✅ StatefulSet includes PVC for LanceDB
- ✅ Backup automation configured

---

### Week 3: Application Implementation ✅ 1/6

**Goal**: Implement core application features

- [x] **Task 5.1**: Implement chat endpoint (basic echo)
  - Reference: [SOURCE_CODE_TEMPLATE.md](./SOURCE_CODE_TEMPLATE.md)
  - Location: `/src/api/routes/chat.py`
  - Status: ✅ Basic template created
  
- [ ] **Task 5.2**: Implement Ollama client
  - Location: `/src/core/ollama.py`
  - Features: Chat completion, streaming
  - Validation: Unit tests
  
- [ ] **Task 5.3**: Implement LanceDB client
  - Location: `/src/core/lancedb.py`
  - Features: Vector search, hybrid search
  - Validation: Integration tests
  
- [ ] **Task 5.4**: Implement Redis client
  - Location: `/src/core/redis.py`
  - Features: Session storage, caching
  - Validation: Connection test
  
- [ ] **Task 5.5**: Implement RAG retriever
  - Location: `/src/rag/retriever.py`
  - Features: Context retrieval, reranking
  - Validation: Unit tests with mock data
  
- [ ] **Task 5.6**: Integrate all components
  - Update: `/src/api/routes/chat.py`
  - Flow: Request → Retrieve → Generate → Response
  - Validation: Integration tests

**Success Criteria**:
- ✅ Chat endpoint works end-to-end
- ✅ RAG retrieval functional
- ✅ Ollama generates responses
- ✅ Redis stores sessions
- ✅ LanceDB vector search works

---

## 🎯 Phase 2: Testing & Automation (Week 3-4)

### Week 3: Testing Suite ⏳ 0/5

**Goal**: Comprehensive automated testing

- [ ] **Task 6.1**: Expand unit tests
  - Coverage target: 70%
  - Files: All `src/` modules
  - Command: `pytest tests/unit/ --cov=src --cov-report=html`
  
- [ ] **Task 6.2**: Create integration tests
  - Location: `src/tests/integration/`
  - Scenarios: Chat flow, RAG retrieval, error handling
  - Command: `pytest tests/integration/ -v`
  
- [ ] **Task 6.3**: Add E2E tests (optional)
  - Location: `src/tests/e2e/`
  - Tools: pytest + httpx
  - Scenarios: Full user workflows
  
- [ ] **Task 6.4**: Configure pytest
  - File: `pytest.ini`
  - Settings: Coverage ≥70%, slow test marking
  - Validation: All tests pass in CI
  
- [ ] **Task 6.5**: Update CI to run tests
  - Already in: `.github/workflows/ci.yml`
  - Verify: Tests run on every PR
  - Expected: Green checkmark

**Success Criteria**:
- ✅ Unit test coverage ≥70%
- ✅ Integration tests pass
- ✅ Tests run automatically in CI
- ✅ Coverage reports generated

---

### Week 4: Backup Automation ✅ 2/4

**Goal**: Automated backup and restore

- [x] **Task 7.1**: Velero installation configured
  - File: `production-fixes/p0-velero/velero-install.yaml`
  - Status: ✅ Already created
  
- [x] **Task 7.2**: Backup schedules configured
  - File: `production-fixes/p0-velero/backup-schedules.yaml`
  - Status: ✅ Already created (hourly + daily)
  
- [ ] **Task 7.3**: Deploy Velero to cluster
  - Command: `kubectl apply -k k8s/backup/`
  - Verify: Velero pods running
  - Expected: Backup CronJobs created
  
- [ ] **Task 7.4**: Test backup and restore
  - Backup: `velero backup create test-backup --include-namespaces agent-bruno`
  - Restore: `velero restore create --from-backup test-backup`
  - Verify: Data restored successfully

**Success Criteria**:
- ✅ Velero deployed and running
- ✅ Hourly backups running
- ✅ Daily backups running
- ✅ Restore procedure tested and documented

---

## 🎯 Phase 3: Documentation Updates (Ongoing)

### Documentation Updates 🟡 3/5

**Goal**: Keep documentation in sync with implementation

- [x] **Task 8.1**: Create unblock plan
  - File: `docs/DEVOPS_UNBLOCK_PLAN.md`
  - Status: ✅ Complete
  
- [x] **Task 8.2**: Update REVIEW_INDEX.md
  - Change status: 🔴 BLOCKED → 🟡 UNBLOCKING
  - Status: ✅ Complete
  
- [x] **Task 8.3**: Create implementation templates
  - Files: `docs/fixes/*.md`
  - Status: ✅ Complete (Dockerfile, GitHub Actions, Source Code)
  
- [ ] **Task 8.4**: Update DEVOPS_ENGINEER_REVIEW.md
  - Add: Implementation status section
  - Update: Completed items checklist
  - Status: ⏳ Pending
  
- [ ] **Task 8.5**: Update CICD_SETUP.md
  - Add: Actual workflow files (not just examples)
  - Update: Setup instructions
  - Status: ⏳ Pending

**Success Criteria**:
- ✅ All documentation reflects current state
- ✅ Templates available for implementation
- ✅ Review status updated

---

## 🎬 Quick Start Commands

### Day 1: Container Foundation
```bash
# Navigate to project root
cd repos/homelab/flux/clusters/homelab/infrastructure/agent-bruno

# Create Dockerfile (copy from template)
cat docs/fixes/DOCKERFILE_TEMPLATE.md

# Build container
docker build -t agent-bruno:test .

# Test locally
docker run -d -p 8080:8080 agent-bruno:test
curl http://localhost:8080/health
```

### Day 3: CI/CD Setup
```bash
# Create workflows directory
mkdir -p .github/workflows

# Copy workflow templates
# (Copy from docs/fixes/GITHUB_ACTIONS_TEMPLATES.md)

# Configure secrets in GitHub UI
# Settings → Secrets → Actions → New repository secret
# - FLUX_GITHUB_TOKEN
# - SLACK_WEBHOOK (optional)
```

### Day 6: Source Code
```bash
# Create source directory
mkdir -p src/{api/{routes,middleware},core,models,rag,services,tests/{unit,integration}}

# Copy templates
# (Copy from docs/fixes/SOURCE_CODE_TEMPLATE.md)

# Install dependencies
pip install -r requirements.txt

# Run tests
pytest tests/ -v

# Start app locally
python src/main.py
```

---

## 📈 Success Metrics

### DevOps Unblocking Criteria (Week 1)
- [x] ✅ Documentation complete
- [ ] ✅ Dockerfile created
- [ ] ✅ CI/CD pipeline implemented
- [ ] ✅ Basic tests running
- [ ] ✅ Container image building automatically
- [ ] ✅ GitOps deployment working

**When all complete**: Change status 🔴 BLOCKED → 🟢 UNBLOCKED

### Production-Ready Criteria (Week 4)
- [ ] ✅ Full source code implementation
- [ ] ✅ Unit test coverage ≥70%
- [ ] ✅ Integration tests passing
- [ ] ✅ Multi-environment deployment
- [ ] ✅ Automated backups working
- [ ] ✅ Security scanning in CI

**When all complete**: Change status 🟢 UNBLOCKED → ✅ PRODUCTION-READY

---

## 🔗 Reference Documentation

### Templates (Use These)
- [DOCKERFILE_TEMPLATE.md](./DOCKERFILE_TEMPLATE.md) - Container image
- [GITHUB_ACTIONS_TEMPLATES.md](./GITHUB_ACTIONS_TEMPLATES.md) - CI/CD workflows
- [SOURCE_CODE_TEMPLATE.md](./SOURCE_CODE_TEMPLATE.md) - FastAPI application

### Plans & Reviews
- [DEVOPS_UNBLOCK_PLAN.md](../DEVOPS_UNBLOCK_PLAN.md) - Overall strategy
- [DEVOPS_ENGINEER_REVIEW.md](../DEVOPS_ENGINEER_REVIEW.md) - Full DevOps review
- [REVIEW_INDEX.md](../REVIEW_INDEX.md) - All reviews summary

### Setup Guides
- [CICD_SETUP.md](../CICD_SETUP.md) - CI/CD configuration
- [BACKUP_SETUP.md](../BACKUP_SETUP.md) - Velero backup setup
- [LANCEDB_PERSISTENCE.md](../LANCEDB_PERSISTENCE.md) - Data persistence

---

## 🎯 This Week's Priorities

### Top 3 Critical Tasks
1. **Create Dockerfile** (Day 1-2) - BLOCKS EVERYTHING
2. **Setup CI/CD Pipeline** (Day 3-5) - ENABLES AUTOMATION
3. **Implement Basic FastAPI App** (Day 6-7) - ENABLES TESTING

### Next Week's Priorities
1. Organize Kubernetes manifests
2. Implement chat endpoint with RAG
3. Deploy to dev environment
4. Test automated backups

---

**Status**: 🟡 15% COMPLETE  
**Next Milestone**: Complete Phase 0 (Week 1)  
**Owner**: Bruno Lucena  
**Last Updated**: October 23, 2025

---

