"""
Shared Testing Library for Homelab Agents.

Provides reusable fixtures, base classes, assertions, and utilities
for testing all agents in the homelab infrastructure.
"""

__version__ = "1.0.0"

# Re-export main components for convenience
from shared_tests.fixtures import (
    cloudevent_factory,
    mock_httpx_client,
    mock_k8s_client,
    mock_ollama_response,
    mock_redis_client,
    metrics_registry,
)
from shared_tests.base import (
    BaseAgentTest,
    BaseCloudEventTest,
    BaseMetricsTest,
    BaseHealthCheckTest,
)
from shared_tests.assertions import (
    assert_cloudevent_valid,
    assert_metric_incremented,
    assert_health_check_passes,
)

__all__ = [
    # Version
    "__version__",
    # Fixtures
    "cloudevent_factory",
    "mock_httpx_client",
    "mock_k8s_client",
    "mock_ollama_response",
    "mock_redis_client",
    "metrics_registry",
    # Base classes
    "BaseAgentTest",
    "BaseCloudEventTest",
    "BaseMetricsTest",
    "BaseHealthCheckTest",
    # Assertions
    "assert_cloudevent_valid",
    "assert_metric_incremented",
    "assert_health_check_passes",
]
