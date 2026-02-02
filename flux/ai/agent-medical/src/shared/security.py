"""
Security utilities for agent-medical.
HIPAA-compliant access control and encryption.
"""
import hashlib
import os
from typing import Optional
from cryptography.fernet import Fernet
import structlog

from .types import User, UserRole

logger = structlog.get_logger()


class AccessControl:
    """Access control for medical records."""
    
    @staticmethod
    def hash_patient_id(patient_id: str) -> str:
        """Hash patient ID for audit logs (HIPAA compliance)."""
        return hashlib.sha256(patient_id.encode()).hexdigest()
    
    @staticmethod
    def verify_access(user: User, patient_id: str) -> bool:
        """Verify if user has access to patient records."""
        return user.has_patient_access(patient_id)
    
    @staticmethod
    def sanitize_query(query: str) -> str:
        """Sanitize query for audit logs (remove PII if needed)."""
        # Basic sanitization - can be enhanced with PII detection
        # For now, just truncate long queries
        if len(query) > 500:
            return query[:500] + "... [truncated]"
        return query


class Encryption:
    """Encryption utilities for sensitive data."""
    
    def __init__(self):
        # Get encryption key from environment or Vault
        key = os.getenv("ENCRYPTION_KEY")
        if not key:
            # Generate key for development (NOT for production!)
            key = Fernet.generate_key().decode()
            logger.warning("encryption_key_generated", note="Using generated key - not for production!")
        
        if isinstance(key, str):
            key = key.encode()
        
        self.cipher = Fernet(key)
    
    def encrypt(self, data: str) -> str:
        """Encrypt sensitive data."""
        return self.cipher.encrypt(data.encode()).decode()
    
    def decrypt(self, encrypted_data: str) -> str:
        """Decrypt sensitive data."""
        return self.cipher.decrypt(encrypted_data.encode()).decode()


def get_user_from_token(token: str) -> Optional[User]:
    """
    Get user from authentication token.
    
    TODO: Implement proper JWT/OAuth2 token validation
    For now, returns a mock user for development.
    """
    # In production, this should:
    # 1. Validate JWT token
    # 2. Extract user info from token
    # 3. Query database for user permissions
    # 4. Return User object
    
    # Mock implementation for development
    # Support both test token formats and legacy formats
    if token == "doctor-token" or token == "test-token-doctor-12345":
        return User(
            id="doctor-123",
            email="doctor@hospital.com",
            role=UserRole.DOCTOR,
            name="Dr. Silva",
            patient_access=[]  # Doctors have access to all
        )
    elif token == "nurse-token" or token == "test-token-nurse-12345":
        return User(
            id="nurse-456",
            email="nurse@hospital.com",
            role=UserRole.NURSE,
            name="Nurse Santos",
            patient_access=["patient-001", "patient-002"]
        )
    elif token == "patient-token" or token == "test-token-patient-001" or token == "test-token-patient-12345":
        return User(
            id="patient-001",
            email="patient@example.com",
            role=UserRole.PATIENT,
            name="Patient Example",
            patient_access=["patient-001"]
        )
    elif token == "test-token-patient-002":
        return User(
            id="patient-002",
            email="patient002@example.com",
            role=UserRole.PATIENT,
            name="Patient Two",
            patient_access=["patient-002"]
        )
    elif token == "test-token-admin-12345":
        return User(
            id="admin-001",
            email="admin@hospital.com",
            role=UserRole.ADMIN,
            name="Admin User",
            patient_access=[]  # Admins have access to all
        )
    
    return None
