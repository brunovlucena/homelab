"""
Prometheus metrics for agent-tools (k8s_tools).
"""
from prometheus_client import Counter, Histogram, Gauge, Info

# =============================================================================
# BUILD INFO
# =============================================================================

BUILD_INFO = Info(
    "agent_tools_build",
    "Build information"
)


def init_build_info(version: str, commit: str = "unknown"):
    """Initialize build info metric."""
    BUILD_INFO.info({
        "version": version,
        "commit": commit,
    })


# =============================================================================
# OPERATION METRICS
# =============================================================================

OPERATIONS_TOTAL = Counter(
    "agent_tools_operations_total",
    "Total Kubernetes operations executed",
    ["operation", "resource_type", "status"]
)

OPERATION_DURATION = Histogram(
    "agent_tools_operation_duration_seconds",
    "Kubernetes operation duration",
    ["operation", "resource_type"],
    buckets=[0.1, 0.5, 1, 2, 5, 10, 30, 60]
)

OPERATION_ERRORS = Counter(
    "agent_tools_operation_errors_total",
    "Kubernetes operation errors",
    ["operation", "resource_type", "error_type"]
)

# =============================================================================
# CLOUDEVENTS METRICS
# =============================================================================

CLOUDEVENTS_RECEIVED = Counter(
    "agent_tools_cloudevents_received_total",
    "Total CloudEvents received",
    ["event_type", "source"]
)

CLOUDEVENTS_PROCESSED = Counter(
    "agent_tools_cloudevents_processed_total",
    "Total CloudEvents processed",
    ["event_type", "status"]
)
