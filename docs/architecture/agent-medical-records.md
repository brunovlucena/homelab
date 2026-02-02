# 🏥 Agent Medical Records (Prontuário Médico)

> **Part of**: [AI Agent Architecture](ai-agent-architecture.md)  
> **Related**: [Agent Orchestration](agent-orchestration.md) | [AI Components](ai-components.md) | [Studio Cluster](../clusters/studio-cluster.md)  
> **Status**: 🟡 Architecture Discussion  
> **Last Updated**: January 2025

---

## 🎯 Overview

Agent Medical Records is a HIPAA-compliant AI agent for managing and accessing electronic medical records (prontuário médico) in a secure, privacy-first architecture. The agent provides natural language access to patient records while maintaining strict access controls and audit trails.

**Key Question**: Should each patient be an agent, or should we have a single medical agent with access to all patient data?

**Recommendation**: **Single Medical Agent** with role-based access control (RBAC) and patient data isolation.

---

## 🏗️ Architecture Decision: Single Agent vs. Per-Patient Agents

### Option 1: One Agent Per Patient ❌ (Not Recommended)

**Approach**: Each patient has their own dedicated agent instance.

**Pros**:
- Complete data isolation per patient
- Simple access model (agent = patient)
- Natural scaling per patient

**Cons**:
- ❌ **Resource Overhead**: 1000 patients = 1000 agent instances (even with scale-to-zero, metadata overhead)
- ❌ **Management Complexity**: Updating agent logic requires updating 1000+ instances
- ❌ **Knowledge Sharing**: Medical knowledge can't be shared across patients (contraindications, drug interactions)
- ❌ **Cost**: Higher operational cost (even with scale-to-zero, there's metadata overhead)
- ❌ **Compliance**: Harder to audit (1000 agents vs. 1 agent with audit logs)
- ❌ **Cross-Patient Analysis**: Can't analyze patterns across patients (epidemiology, research)

### Option 2: Single Medical Agent with RBAC ✅ (Recommended)

**Approach**: One agent (`agent-medical`) with role-based access control and patient data isolation at the data layer.

**Pros**:
- ✅ **Efficient**: Single agent instance, scale-to-zero when idle
- ✅ **Centralized Management**: Update once, applies to all patients
- ✅ **Knowledge Sharing**: Medical knowledge graph shared across all interactions
- ✅ **Cost Effective**: Lower operational overhead
- ✅ **Compliance**: Centralized audit logging, easier HIPAA compliance
- ✅ **Cross-Patient Analysis**: Can analyze patterns (with proper anonymization)
- ✅ **Role-Based Access**: Doctor, nurse, patient, admin roles with different permissions
- ✅ **Data Isolation**: Patient data isolated at database/storage layer (not agent layer)

**Cons**:
- ⚠️ Requires careful RBAC implementation
- ⚠️ Need to ensure data isolation in queries

**Decision**: **Option 2 - Single Medical Agent with RBAC** ✅

---

## 🏥 Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Medical Records Agent Architecture                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │              Agent Medical (Knative Service)                  │ │
│  │                                                               │ │
│  │  ┌───────────────────────────────────────────────────────┐ │ │
│  │  │         Access Control Layer (RBAC)                    │ │ │
│  │  │  • Role verification (doctor, nurse, patient, admin)   │ │ │
│  │  │  • Patient data isolation                             │ │ │
│  │  │  • Audit logging                                       │ │ │
│  │  └───────────────────────────────────────────────────────┘ │ │
│  │                      │                                      │ │
│  │  ┌───────────────────────────────────────────────────────┐ │ │
│  │  │         Intent Classifier (SLM)                       │ │ │
│  │  │  • Classify medical queries                            │ │ │
│  │  │  • Extract patient ID, date ranges                    │ │ │
│  │  │  • Detect sensitive operations                         │ │ │
│  │  └───────────────────────────────────────────────────────┘ │ │
│  │                      │                                      │ │
│  │  ┌───────────────────────────────────────────────────────┐ │ │
│  │  │         Knowledge Graph (LanceDB)                      │ │ │
│  │  │  • Medical protocols                                    │ │ │
│  │  │  • Drug interactions                                   │ │ │
│  │  │  • Treatment guidelines                                │ │ │
│  │  │  • Historical cases (anonymized)                       │ │ │
│  │  └───────────────────────────────────────────────────────┘ │ │
│  │                      │                                      │ │
│  │  ┌───────────────────────────────────────────────────────┐ │ │
│  │  │         Medical Records Database                       │ │ │
│  │  │  • Patient records (encrypted)                          │ │ │
│  │  │  • Lab results                                          │ │ │
│  │  │  • Prescriptions                                        │ │ │
│  │  │  • Medical history                                      │ │ │
│  │  └───────────────────────────────────────────────────────┘ │ │
│  │                      │                                      │ │
│  │  ┌───────────────────────────────────────────────────────┐ │ │
│  │  │         LLM (VLLM) - Medical Reasoning                  │ │ │
│  │  │  • Complex medical analysis                             │ │ │
│  │  │  • Treatment recommendations                            │ │ │
│  │  │  • Risk assessment                                      │ │ │
│  │  └───────────────────────────────────────────────────────┘ │ │
│  │                                                               │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                      │                                              │
│  ┌───────────────────┴──────────────────────────────────────────┐ │
│  │              Data Storage Layer                              │ │
│  │                                                              │ │
│  │  ┌──────────────────┐  ┌──────────────────┐                │ │
│  │  │  PostgreSQL      │  │  MinIO (S3)      │                │ │
│  │  │  (Structured)    │  │  (Documents)     │                │ │
│  │  │  • Patient data  │  │  • Scans         │                │ │
│  │  │  • Lab results   │  │  • Images        │                │ │
│  │  │  • Prescriptions │  │  • Reports       │                │ │
│  │  └──────────────────┘  └──────────────────┘                │ │
│  │                                                              │ │
│  │  ┌──────────────────┐  ┌──────────────────┐                │ │
│  │  │  Vault           │  │  Audit Logs     │                │ │
│  │  │  (Secrets)       │  │  (Immutable)    │                │ │
│  │  │  • DB credentials│  │  • All access   │                │ │
│  │  │  • API keys      │  │  • HIPAA audit  │                │ │
│  │  └──────────────────┘  └──────────────────┘                │ │
│  └──────────────────────────────────────────────────────────────┘ │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 🔐 Security & Compliance

### HIPAA Compliance Requirements

| Requirement | Implementation | Status |
|------------|----------------|--------|
| **Access Controls** | RBAC with role-based permissions | ✅ Design |
| **Encryption at Rest** | PostgreSQL encryption + MinIO encryption | ✅ Design |
| **Encryption in Transit** | TLS 1.3 + Linkerd mTLS | ✅ Design |
| **Audit Logging** | Immutable audit logs (all access) | ✅ Design |
| **Data Minimization** | Query-level patient isolation | ✅ Design |
| **Breach Notification** | Automated alerting on unauthorized access | ⚠️ Planned |
| **Business Associate Agreement (BAA)** | Required for cloud providers | ⚠️ Required |

### Access Control Model

```yaml
Roles:
  doctor:
    permissions:
      - read:all_patients
      - write:own_patients
      - read:lab_results
      - write:prescriptions
      - read:medical_history
    
  nurse:
    permissions:
      - read:assigned_patients
      - write:vitals
      - read:lab_results
      - read:prescriptions
    
  patient:
    permissions:
      - read:own_records
      - read:own_lab_results
      - read:own_prescriptions
      - write:own_symptoms
    
  admin:
    permissions:
      - read:all_patients (audit only)
      - manage:users
      - manage:roles
      - read:audit_logs
```

### Data Isolation Strategy

**Patient Data Isolation**:
- Database: Row-level security (PostgreSQL RLS) by patient_id
- Queries: Always filtered by authenticated user's patient access list
- Storage: MinIO bucket per patient (or encrypted with patient-specific keys)
- Audit: All queries logged with patient_id, user_id, timestamp

**Example Query Flow**:
```python
# User: doctor@hospital.com requests patient 12345
# Agent automatically adds access control

query = "Show me patient 12345's lab results"

# Agent adds RBAC check
if not user.has_access(patient_id=12345):
    return "Access denied: You don't have permission to view this patient"

# Query with patient isolation
results = db.query(
    "SELECT * FROM lab_results WHERE patient_id = $1",
    params=[12345],
    # PostgreSQL RLS automatically enforces access
)
```

---

## 📊 Data Model

### Patient Record Structure

```yaml
Patient:
  id: uuid (encrypted)
  name: string (encrypted)
  cpf: string (encrypted, hashed for search)
  birth_date: date
  created_at: timestamp
  updated_at: timestamp

MedicalRecord:
  id: uuid
  patient_id: uuid (FK, encrypted)
  doctor_id: uuid (FK)
  date: timestamp
  type: enum (consultation, lab_result, prescription, imaging)
  content: jsonb (encrypted)
  attachments: []string (MinIO object keys)
  created_at: timestamp
  updated_at: timestamp

LabResult:
  id: uuid
  patient_id: uuid (FK, encrypted)
  test_name: string
  test_date: timestamp
  results: jsonb (encrypted)
  reference_ranges: jsonb
  status: enum (normal, abnormal, critical)
  created_at: timestamp

Prescription:
  id: uuid
  patient_id: uuid (FK, encrypted)
  doctor_id: uuid (FK)
  medication: string
  dosage: string
  frequency: string
  start_date: date
  end_date: date
  status: enum (active, completed, cancelled)
  created_at: timestamp
```

### Knowledge Graph Collections

```yaml
Collections:
  medical_protocols:
    - Treatment guidelines
    - Clinical pathways
    - Best practices
    
  drug_interactions:
    - Drug-drug interactions
    - Drug-food interactions
    - Contraindications
    
  medical_literature:
    - Research papers (anonymized)
    - Case studies (anonymized)
    - Medical guidelines
    
  historical_cases:
    - Anonymized patient cases
    - Treatment outcomes
    - Patterns and trends
```

---

## 🤖 Agent Implementation

### Agent Structure

```python
# agent-medical/src/medical_agent/
from agent_base import BaseAgent
from typing import Optional
import vault

class MedicalAgent(BaseAgent):
    def __init__(self):
        super().__init__(name="agent-medical")
        
        # Medical-specific tools
        self.tools = [
            Tool(name="get_patient_record", func=self.get_patient_record),
            Tool(name="search_patients", func=self.search_patients),
            Tool(name="get_lab_results", func=self.get_lab_results),
            Tool(name="get_prescriptions", func=self.get_prescriptions),
            Tool(name="create_prescription", func=self.create_prescription),
            Tool(name="check_drug_interactions", func=self.check_drug_interactions),
            Tool(name="get_medical_history", func=self.get_medical_history),
        ]
        
        # Vault for secrets
        self.vault = vault.Client()
        self.db_credentials = self.vault.get_secret("medical-db")
        
        # Medical knowledge graph
        self.medical_kg = LanceDBClient(
            collection="medical_protocols",
            endpoint="lancedb.ml-storage.svc.cluster.local:8000"
        )
    
    async def handle_request(self, query: str, user: User) -> dict:
        """
        Main request handler with RBAC enforcement
        """
        # 1. Verify user access
        if not await self.verify_access(user, query):
            return {"error": "Access denied", "audit_log": True}
        
        # 2. Classify intent with SLM
        intent = await self.classify_medical_intent(query, user.role)
        
        # 3. Extract patient ID (if mentioned)
        patient_id = await self.extract_patient_id(query, user)
        
        # 4. Verify patient access
        if patient_id and not user.has_patient_access(patient_id):
            await self.audit_log(user, "access_denied", patient_id)
            return {"error": "Access denied to patient"}
        
        # 5. Retrieve medical context
        context = await self.retrieve_medical_context(query, patient_id)
        
        # 6. Generate response with LLM
        if intent.complexity == "high":
            response = await self.generate_with_llm(query, context, user.role)
        else:
            response = await self.generate_with_slm(query, context)
        
        # 7. Audit log
        await self.audit_log(user, intent.action, patient_id, query)
        
        return {
            "response": response,
            "patient_id": patient_id,
            "model": "llm" if intent.complexity == "high" else "slm",
            "audit_id": self.audit_id
        }
    
    async def get_patient_record(self, patient_id: str, user: User) -> dict:
        """Get patient record with access control"""
        # Verify access
        if not user.has_patient_access(patient_id):
            await self.audit_log(user, "access_denied", patient_id)
            raise AccessDenied("No access to patient")
        
        # Query with RLS (Row-Level Security)
        record = await self.db.query(
            "SELECT * FROM medical_records WHERE patient_id = $1",
            params=[patient_id],
            user_id=user.id  # For RLS
        )
        
        # Audit
        await self.audit_log(user, "read_record", patient_id)
        
        return record
    
    async def check_drug_interactions(self, medications: list[str]) -> dict:
        """Check drug interactions using knowledge graph"""
        # Search knowledge graph
        interactions = await self.medical_kg.search(
            collection="drug_interactions",
            query=f"interactions between {', '.join(medications)}",
            top_k=10
        )
        
        return {
            "medications": medications,
            "interactions": interactions,
            "severity": self.assess_severity(interactions)
        }
```

---

## 🔄 Integration with SUS Cloud

### SUS (Sistema Único de Saúde) Integration

**Status**: ⚠️ **Architecture Placeholder** - Requires SUS Cloud API specifications

**Note**: The user mentioned checking `@vault` for SUS cloud reference. After searching, no specific SUS cloud architecture was found in the vault. This section provides a **proposed integration pattern** that should be adapted based on actual SUS Cloud API specifications.

**Integration Points** (Proposed):
1. **Patient Data Sync**: Sync patient records from SUS cloud
2. **Lab Results**: Import lab results from SUS systems
3. **Prescription Validation**: Validate prescriptions against SUS formulary
4. **Referral System**: Create referrals to SUS facilities
5. **Vaccination Records**: Sync vaccination history
6. **Medical History**: Import historical medical records

**Architecture** (Proposed):
```
Agent Medical → SUS Cloud API (TLS 1.3)
              ↓
         Vault (API Keys, Certificates)
              ↓
         Audit Logs (HIPAA compliant)
              ↓
         PostgreSQL (Encrypted Storage)
```

**SUS Cloud API Integration** (Proposed Pattern):
```python
class SUSCloudClient:
    def __init__(self):
        self.vault = vault.Client()
        self.api_key = self.vault.get_secret("sus-cloud-api-key")
        self.certificate = self.vault.get_secret("sus-cloud-certificate")
        # TODO: Update endpoint based on actual SUS Cloud API
        self.endpoint = os.getenv("SUS_CLOUD_ENDPOINT", "https://api.sus.gov.br/v1")
    
    async def sync_patient(self, cpf: str) -> dict:
        """
        Sync patient data from SUS cloud
        
        TODO: Adapt based on actual SUS Cloud API:
        - Authentication method (OAuth2, mTLS, API key)
        - Endpoint structure
        - Data format
        - Rate limits
        """
        # Authenticate (method depends on SUS API)
        token = await self.authenticate()
        
        # Fetch patient (endpoint depends on SUS API)
        patient = await self.get(
            f"{self.endpoint}/patients/{cpf}",
            headers={"Authorization": f"Bearer {token}"}
        )
        
        # Store locally (encrypted)
        await self.store_patient(patient)
        
        # Audit log
        await self.audit_log("sus_sync", patient_id=cpf)
        
        return patient
    
    async def validate_prescription(self, prescription: dict) -> dict:
        """
        Validate prescription against SUS formulary
        
        TODO: Adapt based on actual SUS formulary API
        """
        result = await self.post(
            f"{self.endpoint}/prescriptions/validate",
            data=prescription,
            headers={"Authorization": f"Bearer {token}"}
        )
        return result
    
    async def create_referral(self, referral: dict) -> dict:
        """
        Create referral to SUS facility
        
        TODO: Adapt based on actual SUS referral system
        """
        result = await self.post(
            f"{self.endpoint}/referrals",
            data=referral,
            headers={"Authorization": f"Bearer {token}"}
        )
        return result
```

**Next Steps**:
1. [ ] Obtain SUS Cloud API documentation
2. [ ] Define authentication mechanism (OAuth2, mTLS, etc.)
3. [ ] Map data models (SUS → Agent Medical)
4. [ ] Implement sync strategy (real-time vs. batch)
5. [ ] Test integration with SUS sandbox (if available)
6. [ ] Document API rate limits and quotas
7. [ ] Implement retry logic and error handling

---

## 📈 Deployment

### Knative Service Configuration

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: agent-medical
  namespace: ai-agents
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "0"
        autoscaling.knative.dev/maxScale: "10"
        autoscaling.knative.dev/target: "50"
    spec:
      serviceAccountName: agent-medical-sa
      containers:
      - name: agent
        image: ghcr.io/brunovlucena/agent-medical:v1.0.0
        env:
        - name: OLLAMA_URL
          value: "http://ollama.ml-inference.svc.forge.remote:11434"
        - name: VLLM_URL
          value: "http://vllm.ml-inference.svc.forge.remote:8000"
        - name: LANCEDB_URL
          value: "http://lancedb.ml-storage.svc.cluster.local:8000"
        - name: POSTGRES_URL
          valueFrom:
            secretKeyRef:
              name: medical-db-credentials
              key: url
        - name: VAULT_ADDR
          value: "http://vault.vault.svc.cluster.local:8200"
        - name: HIPAA_MODE
          value: "true"
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
```

### RBAC Configuration

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: agent-medical-sa
  namespace: ai-agents
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: agent-medical-role
  namespace: ai-agents
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["medical-db-credentials"]
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: agent-medical-binding
  namespace: ai-agents
subjects:
- kind: ServiceAccount
  name: agent-medical-sa
  namespace: ai-agents
roleRef:
  kind: Role
  name: agent-medical-role
  apiGroup: rbac.authorization.k8s.io
```

---

## 📊 Monitoring & Observability

### Metrics

```yaml
Metrics:
  agent_medical_requests_total{role, status}
  agent_medical_access_denied_total{reason}
  agent_medical_patient_queries_total{patient_id_hash}
  agent_medical_response_duration_seconds{model}
  agent_medical_audit_logs_total{action}
  agent_medical_sus_sync_total{status}
```

### Audit Logging

**All access must be logged**:
- User ID
- Patient ID (hashed for privacy)
- Action (read, write, search)
- Timestamp
- IP address
- Query text (sanitized)
- Response status

**Audit Log Format**:
```json
{
  "timestamp": "2025-01-20T10:30:00Z",
  "user_id": "doctor-123",
  "user_role": "doctor",
  "patient_id_hash": "sha256:abc123...",
  "action": "read_record",
  "query": "Show lab results for patient",
  "status": "success",
  "ip_address": "10.0.0.5",
  "audit_id": "audit-789"
}
```

---

## 🚀 Implementation Roadmap

### Phase 1: Foundation (Weeks 1-4)
- [ ] Agent structure (FastAPI + BaseAgent)
- [ ] RBAC implementation
- [ ] PostgreSQL database with RLS
- [ ] Basic patient record CRUD
- [ ] Audit logging

### Phase 2: Medical Features (Weeks 5-8)
- [ ] Medical knowledge graph (LanceDB)
- [ ] Drug interaction checking
- [ ] Lab results integration
- [ ] Prescription management
- [ ] Medical history tracking

### Phase 3: SUS Integration (Weeks 9-12)
- [ ] SUS Cloud API client
- [ ] Patient data sync
- [ ] Prescription validation
- [ ] Referral system

### Phase 4: Advanced Features (Weeks 13-16)
- [ ] Medical reasoning (LLM)
- [ ] Treatment recommendations
- [ ] Risk assessment
- [ ] Anomaly detection

---

## 🔗 Related Documentation

- [🤖 AI Agent Architecture](ai-agent-architecture.md)
- [🎯 Agent Orchestration](agent-orchestration.md)
- [🔧 AI Components](ai-components.md)
- [🎯 Studio Cluster](../clusters/studio-cluster.md)

---

## 📝 Notes

**Key Architectural Decisions**:
1. ✅ **Single Agent**: One agent with RBAC (not per-patient agents)
2. ✅ **Data Isolation**: At database/storage layer (not agent layer)
3. ✅ **HIPAA Compliance**: Encryption, audit logs, access controls
4. ✅ **Knowledge Graph**: Shared medical knowledge (protocols, interactions)
5. ✅ **SUS Integration**: Cloud API integration for data sync

**Open Questions**:
- [ ] SUS Cloud API specifications (check @vault)
- [ ] Specific HIPAA requirements for Brazil (LGPD compliance)
- [ ] Medical model fine-tuning requirements
- [ ] Integration with existing hospital systems

---

**Last Updated**: January 2025  
**Maintained by**: SRE Team (Bruno Lucena)  
**Status**: 🟡 Architecture Discussion - Awaiting SUS Cloud reference review
