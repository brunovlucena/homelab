"""Pytest configuration and fixtures for agent-restaurant tests."""
import pytest
from unittest.mock import MagicMock, AsyncMock, patch


@pytest.fixture
def mock_httpx_client():
    """Mock httpx async client."""
    mock_client = AsyncMock()
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.json.return_value = {
        "message": {
            "content": "Hello! How can I help you today?",
            "role": "assistant",
        }
    }
    mock_client.post.return_value = mock_response
    mock_client.get.return_value = mock_response
    return mock_client


@pytest.fixture
def sample_order():
    """Sample restaurant order."""
    return {
        "id": "order-123",
        "table_number": 5,
        "items": [
            {"name": "Burger", "quantity": 2, "price": 12.99},
            {"name": "Fries", "quantity": 2, "price": 4.99},
            {"name": "Cola", "quantity": 2, "price": 2.99},
        ],
        "status": "pending",
        "total": 41.94,
    }


@pytest.fixture
def sample_menu_item():
    """Sample menu item."""
    return {
        "id": "item-456",
        "name": "Burger",
        "description": "Delicious beef burger with cheese",
        "price": 12.99,
        "category": "main",
        "available": True,
    }


@pytest.fixture
def sample_reservation():
    """Sample reservation."""
    return {
        "id": "res-789",
        "name": "John Doe",
        "party_size": 4,
        "date": "2025-01-15",
        "time": "19:00",
        "status": "confirmed",
    }


@pytest.fixture
def sample_cloudevent():
    """Sample CloudEvent for testing."""
    return {
        "type": "io.homelab.restaurant.order.create",
        "source": "test",
        "data": {
            "table_number": 5,
            "items": [{"name": "Burger", "quantity": 1}],
        }
    }
