"""
Integration tests for CloudEvents handling.
"""
import pytest
import json
from datetime import datetime
from cloudevents.http import CloudEvent, to_structured


def test_medical_query_cloudevent(client):
    """Test medical query via CloudEvent."""
    event = CloudEvent({
        "type": "io.homelab.medical.query",
        "source": "/test",
        "id": "test-123",
        "time": datetime.utcnow().isoformat() + "Z",
        "datacontenttype": "application/json",
    }, {
        "query": "Show me patient-001's lab results",
        "patient_id": "patient-001",
        "token": "doctor-token"
    })
    
    headers, body = to_structured(event)
    
    response = client.post(
        "/",
        content=body,
        headers=dict(headers)
    )
    
    # 200 = success, 503 = service unavailable (no Ollama in test env)
    assert response.status_code in [200, 503]
    if response.status_code == 200:
        data = response.json()
        assert "agent" in data
        assert data["agent"] == "agent-medical"


def test_access_denied_cloudevent(client):
    """Test access denied scenario."""
    event = CloudEvent({
        "type": "io.homelab.medical.query",
        "source": "/test",
        "id": "test-456",
        "time": datetime.utcnow().isoformat() + "Z",
        "datacontenttype": "application/json",
    }, {
        "query": "Show me patient-999's lab results",
        "patient_id": "patient-999",
        "token": "patient-token"  # Patient trying to access another patient
    })
    
    headers, body = to_structured(event)
    
    response = client.post(
        "/",
        content=body,
        headers=dict(headers)
    )
    
    # Should return 403 or error
    assert response.status_code in [403, 401, 500]
