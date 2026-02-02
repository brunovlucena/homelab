"""Tests for metrics collector."""
import pytest
from unittest.mock import AsyncMock, MagicMock, patch
import httpx

from src.metrics_collector import MetricsCollector


@pytest.mark.asyncio
async def test_query_success(mock_httpx_client, mock_prometheus_response):
    """Test successful Prometheus query."""
    collector = MetricsCollector("http://prometheus:9090")
    collector.client = mock_httpx_client
    
    result = await collector.query("test_metric")
    
    assert result == mock_prometheus_response
    mock_httpx_client.get.assert_called_once()


@pytest.mark.asyncio
async def test_query_failure(mock_httpx_client):
    """Test Prometheus query failure."""
    collector = MetricsCollector("http://prometheus:9090")
    collector.client = mock_httpx_client
    
    # Simulate HTTP error
    mock_httpx_client.get.side_effect = httpx.HTTPError("Connection failed")
    
    with pytest.raises(httpx.HTTPError):
        await collector.query("test_metric")


@pytest.mark.asyncio
async def test_collect_loki_metrics(mock_httpx_client, mock_prometheus_response):
    """Test collecting Loki metrics."""
    collector = MetricsCollector("http://prometheus:9090")
    collector.client = mock_httpx_client
    
    metrics = await collector.collect_loki_metrics()
    
    assert isinstance(metrics, dict)
    assert len(metrics) > 0
    # Should have attempted to query multiple record rules
    assert mock_httpx_client.get.call_count > 0


@pytest.mark.asyncio
async def test_collect_prometheus_metrics(mock_httpx_client, mock_prometheus_response):
    """Test collecting Prometheus metrics."""
    collector = MetricsCollector("http://prometheus:9090")
    collector.client = mock_httpx_client
    
    metrics = await collector.collect_prometheus_metrics()
    
    assert isinstance(metrics, dict)
    assert len(metrics) > 0


@pytest.mark.asyncio
async def test_collect_infrastructure_metrics(mock_httpx_client, mock_prometheus_response):
    """Test collecting infrastructure metrics."""
    collector = MetricsCollector("http://prometheus:9090")
    collector.client = mock_httpx_client
    
    metrics = await collector.collect_infrastructure_metrics()
    
    assert isinstance(metrics, dict)
    assert len(metrics) > 0


@pytest.mark.asyncio
async def test_collect_observability_metrics(mock_httpx_client, mock_prometheus_response):
    """Test collecting observability metrics."""
    collector = MetricsCollector("http://prometheus:9090")
    collector.client = mock_httpx_client
    
    metrics = await collector.collect_observability_metrics()
    
    assert isinstance(metrics, dict)
    assert len(metrics) > 0


@pytest.mark.asyncio
async def test_extract_value_success():
    """Test extracting value from Prometheus response."""
    collector = MetricsCollector("http://prometheus:9090")
    
    data = {
        "result": [
            {
                "value": [1234567890, "0.95"]
            }
        ]
    }
    
    value = collector._extract_value(data)
    assert value == 0.95


@pytest.mark.asyncio
async def test_extract_value_empty():
    """Test extracting value from empty response."""
    collector = MetricsCollector("http://prometheus:9090")
    
    data = {"result": []}
    
    value = collector._extract_value(data)
    assert value is None


@pytest.mark.asyncio
async def test_extract_value_invalid():
    """Test extracting value from invalid response."""
    collector = MetricsCollector("http://prometheus:9090")
    
    data = {
        "result": [
            {
                "value": [1234567890, "invalid"]
            }
        ]
    }
    
    value = collector._extract_value(data)
    assert value is None


@pytest.mark.asyncio
async def test_close():
    """Test closing HTTP client."""
    collector = MetricsCollector("http://prometheus:9090")
    mock_client = AsyncMock()
    collector.client = mock_client
    
    await collector.close()
    
    mock_client.aclose.assert_called_once()

