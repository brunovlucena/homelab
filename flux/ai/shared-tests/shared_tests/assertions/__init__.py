"""
Custom assertions for homelab agent testing.

Provides semantic assertions for common testing patterns
across all agents.
"""

from shared_tests.assertions.cloudevent import (
    assert_cloudevent_valid,
    assert_cloudevent_type,
    assert_cloudevent_source,
    assert_cloudevent_data_contains,
)
from shared_tests.assertions.metrics import (
    assert_metric_incremented,
    assert_metric_value,
    assert_metric_exists,
)
from shared_tests.assertions.health import (
    assert_health_check_passes,
    assert_health_check_fails,
    assert_dependency_healthy,
)
from shared_tests.assertions.kubernetes import (
    assert_k8s_resource_created,
    assert_k8s_resource_deleted,
    assert_k8s_resource_exists,
)

__all__ = [
    # CloudEvent assertions
    "assert_cloudevent_valid",
    "assert_cloudevent_type",
    "assert_cloudevent_source",
    "assert_cloudevent_data_contains",
    # Metrics assertions
    "assert_metric_incremented",
    "assert_metric_value",
    "assert_metric_exists",
    # Health assertions
    "assert_health_check_passes",
    "assert_health_check_fails",
    "assert_dependency_healthy",
    # Kubernetes assertions
    "assert_k8s_resource_created",
    "assert_k8s_resource_deleted",
    "assert_k8s_resource_exists",
]
