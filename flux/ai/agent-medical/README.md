# 🏥 Agent-Medical: HIPAA-Compliant Medical Records Agent

**AI-Powered Medical Records Assistant with RBAC and Audit Logging**

A LambdaAgent that provides natural language access to medical records (prontuário médico) while maintaining strict HIPAA compliance, role-based access control, and comprehensive audit trails.

## 🎯 Overview

Agent-Medical is a HIPAA-compliant AI agent that:
- **Manages Medical Records**: Access patient records, lab results, prescriptions
- **RBAC**: Role-based access control (doctor, nurse, patient, admin)
- **HIPAA Compliance**: Encryption, audit logging, patient data isolation
- **CloudEvents**: Event-driven architecture for integration
- **Knowledge Graph**: Medical protocols, drug interactions, treatment guidelines

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│              Medical Records Agent (LambdaAgent)                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  CloudEvent → RBAC Check → Intent Classification →              │
│  Database Query (RLS) → LLM Reasoning → Response → Audit Log    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Key Features

- **Single Agent with RBAC**: One agent instance with role-based permissions (not per-patient agents)
- **Patient Data Isolation**: Application-level access control with MongoDB
- **HIPAA Compliance**: Encryption, audit logs, access controls
- **Event-Driven**: CloudEvents integration for cross-agent communication
- **Scale-to-Zero**: Knative LambdaAgent with scale-to-zero capability

## 📋 Quick Start

### Prerequisites

- MongoDB database (v4.0+ for ACID transactions)
- Ollama or VLLM for LLM inference
- Vault for secrets management (optional)
- RabbitMQ for CloudEvents (via Knative)

### Deploy

```bash
# Build image
make build

# Push to registry
make push-local  # For local development
make push-remote # For production

# Deploy to cluster
make deploy-studio  # Studio cluster
make deploy-pro     # Pro cluster
```

### Verify Deployment

```bash
# Check LambdaAgent
kubectl get lambdaagent -n agent-medical

# Check pods
kubectl get pods -n agent-medical

# Check logs
kubectl logs -n agent-medical -l app.kubernetes.io/name=agent-medical
```

## 🔐 Security & RBAC

### User Roles

| Role | Permissions |
|------|-------------|
| **Doctor** | Read all patients, write prescriptions, full access |
| **Nurse** | Read assigned patients, write vitals, read lab results |
| **Patient** | Read own records only |
| **Admin** | Audit access, user management |

### Access Control

- **Authentication**: JWT tokens (TODO: implement proper validation)
- **Authorization**: Role-based permissions per patient
- **Data Isolation**: Application-level access control with MongoDB
- **Audit Logging**: All access logged with patient ID hash

## 📡 CloudEvents

### Events Received

| Event Type | Description | Payload |
|------------|-------------|---------|
| `io.homelab.medical.query` | General medical query | `{query, patient_id?, token}` |
| `io.homelab.medical.lab.request` | Lab results request | `{patient_id, token}` |
| `io.homelab.medical.prescription.request` | Prescription request | `{patient_id, token}` |
| `io.homelab.medical.history.request` | Medical history request | `{patient_id, token}` |
| `io.homelab.medical.analyze.request` | Medical analysis request | `{query, patient_id, token}` |

### Events Emitted

| Event Type | Description |
|------------|-------------|
| `io.homelab.medical.response` | Medical query response |
| `io.homelab.medical.access.denied` | Access denied event |
| `io.homelab.medical.audit` | Audit log event |

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OLLAMA_URL` | Ollama endpoint | `http://ollama-native.ollama.svc.cluster.local:11434` |
| `OLLAMA_MODEL` | LLM model | `llama3.2:3b` |
| `VLLM_URL` | VLLM endpoint (for complex queries) | `http://vllm.ml-inference.svc.forge.remote:8000` |
| `MONGODB_URL` | MongoDB connection string | From secret |
| `MONGODB_DATABASE` | MongoDB database name | `medical_db` |
| `VAULT_ADDR` | Vault address | `http://vault.vault.svc.cluster.local:8200` |
| `HIPAA_MODE` | Enable HIPAA compliance | `true` |
| `EVENT_SOURCE` | CloudEvents source | `/agent-medical/records` |

## 📊 API Endpoints

### POST /

Main endpoint for CloudEvents and direct API requests.

**Request** (CloudEvent):
```json
{
  "type": "io.homelab.medical.query",
  "source": "/test",
  "data": {
    "query": "Show me patient-123's lab results",
    "patient_id": "patient-123",
    "token": "doctor-token"
  }
}
```

**Response** (CloudEvent):
```json
{
  "type": "io.homelab.medical.response",
  "source": "/agent-medical/records",
  "data": {
    "agent": "agent-medical",
    "response": "Patient-123's lab results...",
    "patient_id": "patient-123",
    "records": [...],
    "model": "llama3.2:3b",
    "tokens_used": 256,
    "duration_ms": 1234.5,
    "audit_id": "audit-789"
  }
}
```

### GET /health

Health check endpoint.

### GET /ready

Readiness check (verifies Ollama and database connectivity).

### GET /info

Agent information endpoint.

## 🗄️ Database Schema

### MongoDB Collections

The agent uses MongoDB with the following collections:

```javascript
// Patients collection
db.patients.createIndex({ "name": "text" })
db.patients.createIndex({ "patient_id": 1 })

// Medical records collection
db.medical_records.createIndex({ "patient_id": 1, "date": -1 })
db.medical_records.createIndex({ "type": 1 })

// Lab results collection
db.lab_results.createIndex({ "patient_id": 1, "test_date": -1 })
db.lab_results.createIndex({ "status": 1 })

// Prescriptions collection
db.prescriptions.createIndex({ "patient_id": 1, "start_date": -1 })
db.prescriptions.createIndex({ "status": 1 })

// Audit logs collection (HIPAA requirement)
db.audit_logs.createIndex({ "timestamp": -1 })
db.audit_logs.createIndex({ "user_id": 1, "timestamp": -1 })
db.audit_logs.createIndex({ "patient_id_hash": 1 })
```

### Document Structure Examples

```javascript
// Patient document
{
  "_id": ObjectId("..."),
  "id": "patient-123",
  "name": "John Doe",
  "cpf": "encrypted_cpf_value",
  "birth_date": ISODate("1990-01-01"),
  "created_at": ISODate("2024-01-01"),
  "updated_at": ISODate("2024-01-01")
}

// Medical record document
{
  "_id": ObjectId("..."),
  "patient_id": "patient-123",
  "doctor_id": "doctor-456",
  "date": ISODate("2024-01-15"),
  "type": "consultation",
  "content": {
    "symptoms": "...",
    "diagnosis": "...",
    "notes": "..."
  },
  "attachments": ["minio://bucket/file.pdf"],
  "created_at": ISODate("2024-01-15"),
  "updated_at": ISODate("2024-01-15")
}

// Lab result document
{
  "_id": ObjectId("..."),
  "patient_id": "patient-123",
  "test_name": "Complete Blood Count",
  "test_date": ISODate("2024-01-10"),
  "results": {
    "hemoglobin": 14.5,
    "hematocrit": 42.0
  },
  "reference_ranges": {
    "hemoglobin": { "min": 12.0, "max": 16.0 }
  },
  "status": "normal",
  "created_at": ISODate("2024-01-10")
}

// Prescription document
{
  "_id": ObjectId("..."),
  "patient_id": "patient-123",
  "doctor_id": "doctor-456",
  "medication": "Amoxicillin",
  "dosage": "500mg",
  "frequency": "3x daily",
  "start_date": ISODate("2024-01-15"),
  "end_date": ISODate("2024-01-22"),
  "status": "active",
  "created_at": ISODate("2024-01-15")
}

// Audit log document
{
  "_id": ObjectId("..."),
  "timestamp": ISODate("2024-01-15T10:30:00Z"),
  "user_id": "doctor-456",
  "user_role": "doctor",
  "patient_id_hash": "sha256_hash_of_patient_id",
  "action": "read_record",
  "query": "Show lab results",
  "status": "success",
  "ip_address": "10.0.0.5",
  "audit_id": "audit-uuid"
}
```

### Access Control

Access control is enforced at the **application level**:
- User permissions are checked before database queries
- Patient data is filtered by `patient_id` in queries
- Only authorized users can access specific patients

## 📈 Monitoring

### Metrics

- `agent_medical_requests_total{role, status}` - Total requests
- `agent_medical_access_denied_total{reason}` - Access denied count
- `agent_medical_response_duration_seconds{model}` - Response latency
- `agent_medical_audit_logs_total{action}` - Audit log entries
- `agent_medical_tokens_used_total{model, type}` - LLM tokens used

### Audit Logs

All access is logged with:
- User ID and role
- Patient ID (hashed)
- Action performed
- Query (sanitized)
- Timestamp
- Audit ID

## 🔗 Related Documentation

- [Architecture Documentation](../../docs/architecture/agent-medical-records.md)
- [AI Agent Architecture](../../docs/architecture/ai-agent-architecture.md)
- [Agent Orchestration](../../docs/architecture/agent-orchestration.md)

## 🚀 Roadmap

### Phase 1: Foundation ✅
- [x] LambdaAgent structure
- [x] RBAC implementation
- [x] Database integration
- [x] CloudEvents support
- [x] Basic audit logging

### Phase 2: Medical Features
- [ ] Knowledge Graph integration (LanceDB)
- [ ] Drug interaction checking
- [ ] Medical protocol retrieval
- [ ] Treatment recommendations

### Phase 3: SUS Integration
- [ ] SUS Cloud API client
- [ ] Patient data sync
- [ ] Prescription validation
- [ ] Referral system

### Phase 4: Advanced Features
- [ ] Medical reasoning (VLLM)
- [ ] Anomaly detection
- [ ] Risk assessment
- [ ] Pattern analysis

## 📝 License

Part of the homelab project. MIT License.

---

**🏥 HIPAA-Compliant Medical Records Management with AI! 🏥**
