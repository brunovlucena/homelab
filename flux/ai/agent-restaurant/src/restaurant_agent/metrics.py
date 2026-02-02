"""
Prometheus metrics for agent-restaurant.

Principal SRE Engineering Standards:
- RED metrics (Rate, Errors, Duration) for all operations
- Memory-specific metrics for domain memory tracking
- Restaurant service metrics for hospitality operations
"""
from prometheus_client import Counter, Histogram, Gauge, Info, Summary

# =============================================================================
# BUILD INFO
# =============================================================================

BUILD_INFO = Info(
    "agent_restaurant_build",
    "Build information"
)


def init_build_info(version: str, commit: str = "unknown", memory_enabled: bool = False, role: str = "unknown"):
    """Initialize build info metric."""
    BUILD_INFO.info({
        "version": version,
        "commit": commit,
        "memory_enabled": str(memory_enabled).lower(),
        "role": role,
    })


# =============================================================================
# REQUEST METRICS
# =============================================================================

REQUESTS_TOTAL = Counter(
    "agent_restaurant_requests_total",
    "Total requests processed",
    ["request_type", "status"]
)

REQUEST_DURATION = Histogram(
    "agent_restaurant_request_duration_seconds",
    "Request processing duration",
    ["request_type"],
    buckets=[0.1, 0.5, 1, 2, 5, 10, 30, 60]
)

# =============================================================================
# CLOUDEVENTS METRICS
# =============================================================================

CLOUDEVENTS_RECEIVED = Counter(
    "agent_restaurant_cloudevents_received_total",
    "Total CloudEvents received",
    ["event_type", "source"]
)

CLOUDEVENTS_PROCESSED = Counter(
    "agent_restaurant_cloudevents_processed_total",
    "Total CloudEvents processed",
    ["event_type", "status"]
)

EVENT_PROCESSING_DURATION = Histogram(
    "agent_restaurant_event_processing_seconds",
    "Time to process CloudEvents",
    ["event_type"],
    buckets=[0.01, 0.05, 0.1, 0.5, 1, 5, 10]
)

# =============================================================================
# LLM METRICS
# =============================================================================

LLM_CALLS = Counter(
    "agent_restaurant_llm_calls_total",
    "Total LLM API calls",
    ["model", "status"]
)

LLM_DURATION = Histogram(
    "agent_restaurant_llm_duration_seconds",
    "LLM inference duration",
    ["model"],
    buckets=[0.5, 1, 2, 5, 10, 30, 60]
)

TOKENS_USED = Counter(
    "agent_restaurant_tokens_total",
    "Total tokens used",
    ["model", "type"]  # type: input, output
)

# =============================================================================
# DOMAIN MEMORY METRICS
# =============================================================================

MEMORY_OPERATIONS = Counter(
    "agent_restaurant_memory_operations_total",
    "Memory operations performed",
    ["operation", "memory_type", "status"]
)

MEMORY_CONTEXT_BUILD_DURATION = Histogram(
    "agent_restaurant_memory_context_build_seconds",
    "Time to build memory context",
    buckets=[0.01, 0.025, 0.05, 0.1, 0.25, 0.5]
)

MEMORY_STORE_CONNECTED = Gauge(
    "agent_restaurant_memory_store_connected",
    "Memory store connection status",
    ["store_type"]
)

# =============================================================================
# RESTAURANT SERVICE METRICS
# =============================================================================

GUESTS_SERVED = Counter(
    "agent_restaurant_guests_served_total",
    "Total guests served",
    ["service_type"]  # type: greeting, order, recommendation, checkout
)

GUEST_PREFERENCES_RECORDED = Counter(
    "agent_restaurant_guest_preferences_total",
    "Guest preferences recorded",
    ["preference_type"]  # type: dietary, seating, drink
)

GUEST_FACTS_RECORDED = Counter(
    "agent_restaurant_guest_facts_total",
    "Guest facts recorded",
    ["source"]  # source: conversation, explicit
)

ACTIVE_TABLES = Gauge(
    "agent_restaurant_active_tables",
    "Currently active tables being served"
)

SERVICE_QUALITY_SCORE = Summary(
    "agent_restaurant_service_quality",
    "Service quality scores from feedback"
)

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

def record_memory_operation(operation: str, memory_type: str, status: str = "success"):
    """Record a memory operation."""
    MEMORY_OPERATIONS.labels(
        operation=operation,
        memory_type=memory_type,
        status=status,
    ).inc()


def record_guest_served(service_type: str):
    """Record a guest service event."""
    GUESTS_SERVED.labels(service_type=service_type).inc()


def record_guest_preference(preference_type: str):
    """Record a guest preference."""
    GUEST_PREFERENCES_RECORDED.labels(preference_type=preference_type).inc()


def record_guest_fact(source: str):
    """Record a guest fact."""
    GUEST_FACTS_RECORDED.labels(source=source).inc()


def set_memory_store_connected(store_type: str, connected: bool):
    """Set memory store connection status."""
    MEMORY_STORE_CONNECTED.labels(store_type=store_type).set(1 if connected else 0)
