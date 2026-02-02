"""
Universal conftest.py template for homelab agents.

Copy this file to your agent's tests/ directory and customize as needed.
This template imports all shared fixtures and provides a base configuration
for testing any agent in the homelab infrastructure.

Usage:
1. Copy this file to your agent's tests/conftest.py
2. Update AGENT_NAME and AGENT_MODULE
3. Add agent-specific fixtures as needed
4. Import shared test fixtures and base classes in your tests
"""

import os
import sys
import pytest
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch

# =============================================================================
# CONFIGURATION - Update these for your agent
# =============================================================================

AGENT_NAME = "agent-example"  # Change to your agent name
AGENT_MODULE = "example_handler"  # Change to your handler module name
AGENT_NAMESPACE = f"{AGENT_NAME}"  # Kubernetes namespace

# =============================================================================
# PATH SETUP
# =============================================================================

# Add src to path for imports
src_path = Path(__file__).parent.parent / "src"
sys.path.insert(0, str(src_path))

# Add shared-tests to path
shared_tests_path = Path(__file__).parent.parent.parent / "shared-tests"
sys.path.insert(0, str(shared_tests_path))

# =============================================================================
# ENVIRONMENT VARIABLES - Set test defaults
# =============================================================================

os.environ.setdefault("TESTING", "true")
os.environ.setdefault("LOG_LEVEL", "DEBUG")
os.environ.setdefault("OLLAMA_URL", "http://localhost:11434")
os.environ.setdefault("REDIS_URL", "redis://localhost:6379")

# =============================================================================
# IMPORT SHARED FIXTURES
# =============================================================================

# Import all shared fixtures - these will be available in all tests
from shared_tests.fixtures.cloudevents import (
    cloudevent_factory,
    sample_cloudevent,
    sample_cloudevent_batch,
    sample_chat_event,
    sample_exploit_event,
    sample_defense_event,
    sample_contract_event,
    sample_mag7_event,
)

from shared_tests.fixtures.kubernetes import (
    mock_k8s_client,
    mock_k8s_pod,
    mock_k8s_deployment,
    mock_k8s_namespace,
    k8s_resource_factory,
)

from shared_tests.fixtures.http import (
    mock_httpx_client,
    mock_httpx_response,
    respx_mock,
    http_success_response,
    http_error_response,
    http_not_found_response,
)

from shared_tests.fixtures.ollama import (
    mock_ollama_response,
    mock_ollama_client,
    mock_ollama_chat_response,
    mock_ollama_embedding_response,
    ollama_response_factory,
    mock_security_analysis_response,
    mock_code_review_response,
)

from shared_tests.fixtures.metrics import (
    metrics_registry,
    metrics_helper,
    assert_metric_value,
    sample_counter,
    sample_gauge,
    sample_histogram,
)

from shared_tests.fixtures.redis import (
    mock_redis_client,
    mock_redis_with_data,
    redis_cache_factory,
)

# =============================================================================
# ASYNC CONFIGURATION
# =============================================================================

@pytest.fixture(scope="session")
def event_loop():
    """Create event loop for async tests."""
    import asyncio
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()


# =============================================================================
# AGENT-SPECIFIC FIXTURES - Add your custom fixtures below
# =============================================================================

@pytest.fixture
def agent_config():
    """Agent-specific configuration."""
    return {
        "name": AGENT_NAME,
        "namespace": AGENT_NAMESPACE,
        "log_level": "DEBUG",
    }


@pytest.fixture
def mock_handler():
    """
    Create a mock handler for your agent.
    
    Override this fixture in your agent's conftest.py to return
    your actual handler instance.
    """
    handler = MagicMock()
    handler.process = AsyncMock(return_value={"status": "success"})
    handler.health_check = AsyncMock(return_value=True)
    return handler


@pytest.fixture
def mock_dependencies(mock_k8s_client, mock_redis_client, mock_httpx_client):
    """Bundle all common mocked dependencies."""
    return {
        "k8s": mock_k8s_client,
        "redis": mock_redis_client,
        "http": mock_httpx_client,
    }


# =============================================================================
# PYTEST CONFIGURATION
# =============================================================================

def pytest_configure(config):
    """Configure pytest."""
    # Add custom markers
    config.addinivalue_line(
        "markers", "integration: mark test as integration test"
    )
    config.addinivalue_line(
        "markers", "slow: mark test as slow running"
    )
    config.addinivalue_line(
        "markers", "requires_ollama: mark test as requiring Ollama"
    )
    config.addinivalue_line(
        "markers", "requires_k8s: mark test as requiring Kubernetes"
    )


def pytest_collection_modifyitems(config, items):
    """Modify test collection based on markers."""
    # Skip integration tests unless explicitly requested
    if not config.getoption("--integration", default=False):
        skip_integration = pytest.mark.skip(
            reason="Need --integration option to run"
        )
        for item in items:
            if "integration" in item.keywords:
                item.add_marker(skip_integration)


def pytest_addoption(parser):
    """Add custom command line options."""
    parser.addoption(
        "--integration",
        action="store_true",
        default=False,
        help="Run integration tests",
    )
    parser.addoption(
        "--slow",
        action="store_true",
        default=False,
        help="Run slow tests",
    )
