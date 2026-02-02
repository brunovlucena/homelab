"""Tests for security module."""
import pytest
from shared.security import AccessControl, get_user_from_token
from shared.types import UserRole


def test_hash_patient_id():
    """Test patient ID hashing."""
    access_control = AccessControl()
    
    patient_id = "patient-123"
    hash1 = access_control.hash_patient_id(patient_id)
    hash2 = access_control.hash_patient_id(patient_id)
    
    # Should be consistent
    assert hash1 == hash2
    
    # Should be different for different IDs
    hash3 = access_control.hash_patient_id("patient-456")
    assert hash1 != hash3


def test_verify_access_doctor():
    """Test doctor access (should have access to all patients)."""
    access_control = AccessControl()
    
    from shared.types import User
    doctor = User(
        id="doctor-123",
        email="doctor@hospital.com",
        role=UserRole.DOCTOR,
        name="Dr. Silva"
    )
    
    assert access_control.verify_access(doctor, "patient-001") is True
    assert access_control.verify_access(doctor, "patient-999") is True


def test_verify_access_patient():
    """Test patient access (only own records)."""
    access_control = AccessControl()
    
    from shared.types import User
    patient = User(
        id="patient-001",
        email="patient@example.com",
        role=UserRole.PATIENT,
        name="Patient"
    )
    
    assert access_control.verify_access(patient, "patient-001") is True
    assert access_control.verify_access(patient, "patient-002") is False


def test_sanitize_query():
    """Test query sanitization."""
    access_control = AccessControl()
    
    short_query = "Show lab results"
    assert access_control.sanitize_query(short_query) == short_query
    
    long_query = "A" * 1000
    sanitized = access_control.sanitize_query(long_query)
    assert len(sanitized) == 515  # 500 + "... [truncated]" (15 chars)
    assert "[truncated]" in sanitized


def test_get_user_from_token():
    """Test user token validation."""
    doctor = get_user_from_token("doctor-token")
    assert doctor is not None
    assert doctor.role == UserRole.DOCTOR
    
    nurse = get_user_from_token("nurse-token")
    assert nurse is not None
    assert nurse.role == UserRole.NURSE
    
    patient = get_user_from_token("patient-token")
    assert patient is not None
    assert patient.role == UserRole.PATIENT
    
    invalid = get_user_from_token("invalid-token")
    assert invalid is None
