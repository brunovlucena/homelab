# 🧪 Testing agent-medical

## Quick Test

```bash
# Run local test script
./test-local.sh
```

## Manual Testing

### 1. Install Dependencies

```bash
cd flux/ai/agent-medical
python3 -m venv venv
source venv/bin/activate
pip install -r src/requirements.txt
pip install -r tests/requirements.txt
```

### 2. Run Unit Tests

```bash
# Test security module
pytest tests/unit/test_security.py -v

# Test health endpoints
pytest tests/test_health.py -v
```

### 3. Test CloudEvents (Manual)

```bash
# Start the agent locally
cd src/medical_agent
python -m uvicorn main:app --port 8080

# In another terminal, send a CloudEvent
curl -X POST http://localhost:8080/ \
  -H "ce-type: io.homelab.medical.query" \
  -H "ce-source: /test" \
  -H "ce-id: test-123" \
  -H "Content-Type: application/cloudevents+json" \
  -H "Authorization: Bearer doctor-token" \
  -d '{
    "query": "Show me patient-001 lab results",
    "patient_id": "patient-001"
  }'
```

### 4. Test Access Control

```bash
# Test access denied (patient trying to access another patient)
curl -X POST http://localhost:8080/ \
  -H "ce-type: io.homelab.medical.query" \
  -H "ce-source: /test" \
  -H "Authorization: Bearer patient-token" \
  -d '{
    "query": "Show me patient-999 lab results",
    "patient_id": "patient-999"
  }'
# Should return 403 Forbidden
```

## Integration Testing

### Prerequisites

- MongoDB running (or mock)
- Ollama running externally on host machine (NOT in container)
- Vault running (optional)

### Test with Docker Compose

```yaml
# docker-compose.test.yml
version: '3.8'
services:
  mongodb:
    image: mongo:7
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_DATABASE: medical_db
```

```bash
# Start MongoDB only - Ollama runs externally on your host
docker-compose -f docker-compose.test.yml up -d

export MONGODB_URL="mongodb://localhost:27017/medical_db"
# Use your external Ollama (running on host machine)
export OLLAMA_URL="http://localhost:11434"
pytest tests/ -v
```

## Kubernetes Testing

### Deploy to Cluster

```bash
# Build and push image
make build
make push-local  # or push-remote

# Deploy
make deploy-studio
```

### Test via CloudEvent

```bash
# Get service URL
kubectl get ksvc -n agent-medical agent-medical

# Send CloudEvent
curl -X POST http://agent-medical.agent-medical.svc.cluster.local/ \
  -H "ce-type: io.homelab.medical.query" \
  -H "ce-source: /test" \
  -H "Authorization: Bearer doctor-token" \
  -d '{"query": "Show patient-001 lab results", "patient_id": "patient-001"}'
```

## Test Coverage

Current test coverage:
- ✅ Security module (RBAC, hashing, sanitization)
- ✅ Health endpoints
- ⚠️ CloudEvents (requires running service)
- ⚠️ Database operations (requires MongoDB)
- ⚠️ LLM integration (requires Ollama)

## Notes

- Tests use mocks for external dependencies
- HIPAA mode is disabled in tests (`HIPAA_MODE=false`)
- MongoDB connection is optional (tests will skip if unavailable)
- **Ollama runs externally on host machine** - all agents share the same external Ollama instance
- In K8s: `http://ollama-native.ollama.svc.cluster.local:11434`
- Locally: `http://localhost:11434` (your host Ollama)
