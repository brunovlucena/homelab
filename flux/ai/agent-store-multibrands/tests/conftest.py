"""
Pytest configuration and fixtures for Agent Store MultiBrands tests.
"""
import pytest
import asyncio
from typing import Generator


@pytest.fixture(scope="session")
def event_loop() -> Generator:
    """Create event loop for async tests."""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()


@pytest.fixture
def sample_product():
    """Sample product for testing."""
    return {
        "id": "test-001",
        "name": "Test Product",
        "brand": "fashion",
        "description": "A test product",
        "price": 99.90,
        "category": "test",
        "tags": ["test"],
        "stock": 10,
    }


@pytest.fixture
def sample_customer():
    """Sample customer for testing."""
    return {
        "id": "cust-001",
        "phone": "5511999999999",
        "name": "Test Customer",
    }


@pytest.fixture
def sample_order_items():
    """Sample order items for testing."""
    return [
        {
            "product_id": "fashion-001",
            "product_name": "Vestido Floral",
            "quantity": 1,
            "unit_price": 199.90,
            "brand": "fashion",
        },
        {
            "product_id": "fashion-002",
            "product_name": "Jaqueta Jeans",
            "quantity": 2,
            "unit_price": 289.90,
            "brand": "fashion",
        },
    ]
