"""Pytest configuration and fixtures."""
import pytest
from unittest.mock import AsyncMock, MagicMock
import httpx


@pytest.fixture
def mock_prometheus_response():
    """Mock Prometheus API response."""
    return {
        "status": "success",
        "data": {
            "resultType": "vector",
            "result": [
                {
                    "metric": {},
                    "value": [1234567890, "0.95"]
                }
            ]
        }
    }


@pytest.fixture
def mock_httpx_client(mock_prometheus_response):
    """Mock httpx client for Prometheus queries."""
    client = AsyncMock(spec=httpx.AsyncClient)
    response = MagicMock()
    response.json.return_value = mock_prometheus_response
    response.raise_for_status = MagicMock()
    client.get = AsyncMock(return_value=response)
    return client


@pytest.fixture
def sample_metrics():
    """Sample metrics data."""
    return {
        "loki:health:score": 0.95,
        "loki:health:availability:ratio": 0.99,
        "loki:health:error_rate:ingestion": 0.001,
        "loki:health:query_latency:p95": 150.5
    }

