"""
Shared pytest fixtures for homelab agents.

Usage:
    from shared_tests.fixtures import *
    # or
    from shared_tests.fixtures import cloudevent_factory, mock_k8s_client
"""

from shared_tests.fixtures.cloudevents import (
    cloudevent_factory,
    sample_cloudevent,
    sample_cloudevent_batch,
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
)
from shared_tests.fixtures.ollama import (
    mock_ollama_response,
    mock_ollama_client,
    mock_ollama_chat_response,
    mock_ollama_embedding_response,
)
from shared_tests.fixtures.metrics import (
    metrics_registry,
    assert_metric_value,
    reset_metrics,
)
from shared_tests.fixtures.redis import (
    mock_redis_client,
)

__all__ = [
    # CloudEvents
    "cloudevent_factory",
    "sample_cloudevent",
    "sample_cloudevent_batch",
    # Kubernetes
    "mock_k8s_client",
    "mock_k8s_pod",
    "mock_k8s_deployment",
    "mock_k8s_namespace",
    "k8s_resource_factory",
    # HTTP
    "mock_httpx_client",
    "mock_httpx_response",
    "respx_mock",
    # Ollama
    "mock_ollama_response",
    "mock_ollama_client",
    "mock_ollama_chat_response",
    "mock_ollama_embedding_response",
    # Metrics
    "metrics_registry",
    "assert_metric_value",
    "reset_metrics",
    # Redis
    "mock_redis_client",
]
