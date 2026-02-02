"""Pytest configuration and fixtures."""
import pytest
import os
import sys
from pathlib import Path
from unittest.mock import Mock, AsyncMock, patch
from fastapi.testclient import TestClient

# Add src to Python path
src_path = Path(__file__).parent.parent / "src"
sys.path.insert(0, str(src_path))

# Set test environment variables
os.environ["OLLAMA_URL"] = "http://localhost:11434"
os.environ["MONGODB_URL"] = "mongodb://localhost:27017/test_db"
os.environ["HIPAA_MODE"] = "false"  # Disable HIPAA for tests
os.environ["VAULT_ADDR"] = "http://localhost:8200"

from shared.types import User, UserRole
from medical_agent.main import app


@pytest.fixture
def client():
    """Test client for FastAPI app."""
    return TestClient(app)


@pytest.fixture
def mock_user_doctor():
    """Mock doctor user."""
    return User(
        id="doctor-123",
        email="doctor@hospital.com",
        role=UserRole.DOCTOR,
        name="Dr. Silva",
        patient_access=[]
    )


@pytest.fixture
def mock_user_nurse():
    """Mock nurse user."""
    return User(
        id="nurse-456",
        email="nurse@hospital.com",
        role=UserRole.NURSE,
        name="Nurse Santos",
        patient_access=["patient-001", "patient-002"]
    )


@pytest.fixture
def mock_user_patient():
    """Mock patient user."""
    return User(
        id="patient-001",
        email="patient@example.com",
        role=UserRole.PATIENT,
        name="Patient Example",
        patient_access=["patient-001"]
    )


@pytest.fixture
def mock_database():
    """Mock database."""
    db = Mock()
    db.db = Mock()
    db.get_patient_record = AsyncMock(return_value=None)
    db.get_medical_records = AsyncMock(return_value=[])
    db.get_lab_results = AsyncMock(return_value=[])
    db.get_prescriptions = AsyncMock(return_value=[])
    return db
