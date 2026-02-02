"""
Health check tests.
"""
import pytest
from fastapi.testclient import TestClient


def test_health_endpoint(client):
    """Test health endpoint."""
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert "status" in data
    assert data["agent"] == "agent-medical"


def test_ready_endpoint(client):
    """Test readiness endpoint."""
    response = client.get("/ready")
    # May return 200 or 503 depending on dependencies
    assert response.status_code in [200, 503]
    data = response.json()
    assert "agent" in data


def test_info_endpoint(client):
    """Test info endpoint."""
    response = client.get("/info")
    assert response.status_code == 200
    data = response.json()
    assert data["name"] == "agent-medical"
    assert "hipaa_mode" in data
