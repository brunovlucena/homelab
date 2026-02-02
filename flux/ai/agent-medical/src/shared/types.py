"""
Shared types for agent-medical.
"""
from datetime import datetime
from enum import Enum
from typing import Optional, List, Dict, Any
from uuid import UUID, uuid4
from pydantic import BaseModel, Field


class UserRole(str, Enum):
    """User roles for RBAC."""
    DOCTOR = "doctor"
    NURSE = "nurse"
    PATIENT = "patient"
    ADMIN = "admin"


class MedicalRecordType(str, Enum):
    """Types of medical records."""
    CONSULTATION = "consultation"
    LAB_RESULT = "lab_result"
    PRESCRIPTION = "prescription"
    IMAGING = "imaging"
    VITAL_SIGNS = "vital_signs"
    NOTE = "note"


class QueryIntent(str, Enum):
    """Medical query intents."""
    READ_RECORD = "read_record"
    SEARCH_PATIENTS = "search_patients"
    GET_LAB_RESULTS = "get_lab_results"
    GET_PRESCRIPTIONS = "get_prescriptions"
    CREATE_PRESCRIPTION = "create_prescription"
    CHECK_INTERACTIONS = "check_interactions"
    GET_HISTORY = "get_history"
    ANALYZE = "analyze"
    UNKNOWN = "unknown"


class QueryComplexity(str, Enum):
    """Query complexity levels."""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"


class User(BaseModel):
    """User model with role and permissions."""
    id: str
    email: str
    role: UserRole
    name: str
    patient_access: List[str] = Field(default_factory=list)  # List of patient IDs user can access
    
    def has_patient_access(self, patient_id: str) -> bool:
        """Check if user has access to a specific patient."""
        if self.role == UserRole.ADMIN:
            return True
        if self.role == UserRole.DOCTOR:
            return True  # Doctors can access all patients
        if self.role == UserRole.NURSE:
            return patient_id in self.patient_access
        if self.role == UserRole.PATIENT:
            # Patients can only access their own records
            return patient_id == self.id
        return False


class Patient(BaseModel):
    """Patient model."""
    id: str
    name: str
    cpf: str  # Encrypted in storage
    birth_date: datetime
    created_at: datetime
    updated_at: datetime


class MedicalRecord(BaseModel):
    """Medical record model."""
    id: str
    patient_id: str
    doctor_id: Optional[str] = None
    date: datetime
    type: MedicalRecordType
    content: Dict[str, Any]
    attachments: List[str] = Field(default_factory=list)  # MinIO object keys
    created_at: datetime
    updated_at: datetime


class LabResult(BaseModel):
    """Lab result model."""
    id: str
    patient_id: str
    test_name: str
    test_date: datetime
    results: Dict[str, Any]
    reference_ranges: Dict[str, Any]
    status: str  # normal, abnormal, critical
    created_at: datetime


class Prescription(BaseModel):
    """Prescription model."""
    id: str
    patient_id: str
    doctor_id: str
    medication: str
    dosage: str
    frequency: str
    start_date: datetime
    end_date: Optional[datetime] = None
    status: str  # active, completed, cancelled
    created_at: datetime


class MedicalQuery(BaseModel):
    """Medical query request."""
    query: str = Field(..., min_length=1, max_length=4096)
    patient_id: Optional[str] = None
    conversation_id: Optional[str] = None


class MedicalResponse(BaseModel):
    """Medical query response."""
    response: str
    patient_id: Optional[str] = None
    records: List[Dict[str, Any]] = Field(default_factory=list)
    model: str = ""
    tokens_used: int = 0
    duration_ms: float = 0.0
    audit_id: str = ""


class IntentClassification(BaseModel):
    """Intent classification result."""
    intent: QueryIntent
    complexity: QueryComplexity
    patient_id: Optional[str] = None
    confidence: float = 0.0


class AuditLog(BaseModel):
    """Audit log entry."""
    id: str = Field(default_factory=lambda: str(uuid4()))
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    user_id: str
    user_role: UserRole
    patient_id_hash: str  # SHA256 hash of patient ID
    action: str
    query: Optional[str] = None
    status: str  # success, denied, error
    ip_address: Optional[str] = None
    audit_id: str = Field(default_factory=lambda: str(uuid4()))
