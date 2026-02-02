"""
Prometheus metrics for agent-bruno chatbot.

Principal SRE Engineering Standards:
- RED metrics (Rate, Errors, Duration) for all operations
- Memory-specific metrics for domain memory tracking
- Tracing integration for distributed observability
"""
from prometheus_client import Counter, Histogram, Gauge, Info, Summary

# =============================================================================
# MESSAGE PROCESSING METRICS
# =============================================================================

MESSAGES_PROCESSED = Counter(
    "agent_bruno_messages_total",
    "Total messages processed",
    ["status"]  # status: success, error
)

CONVERSATIONS_ACTIVE = Gauge(
    "agent_bruno_active_conversations",
    "Currently active conversations"
)

CONVERSATION_LENGTH = Histogram(
    "agent_bruno_conversation_length_messages",
    "Number of messages in conversations",
    buckets=[1, 3, 5, 10, 20, 50, 100]
)

# =============================================================================
# PERFORMANCE METRICS
# =============================================================================

RESPONSE_DURATION = Histogram(
    "agent_bruno_response_duration_seconds",
    "Time to generate response",
    ["model"],
    buckets=[0.1, 0.5, 1, 2, 5, 10, 30, 60]
)

LLM_INFERENCE_DURATION = Histogram(
    "agent_bruno_llm_inference_seconds",
    "LLM inference time",
    ["model"],
    buckets=[0.1, 0.5, 1, 2, 5, 10, 30, 60, 120]
)

# =============================================================================
# RESOURCE METRICS
# =============================================================================

TOKENS_USED = Counter(
    "agent_bruno_tokens_total",
    "Total LLM tokens consumed",
    ["model", "type"]  # type: input, output
)

API_CALLS = Counter(
    "agent_bruno_api_calls_total",
    "API calls to LLM service",
    ["service", "status"]  # service: ollama; status: success, error
)

# =============================================================================
# DOMAIN MEMORY METRICS (Agent-specific)
# =============================================================================

MEMORY_OPERATIONS = Counter(
    "agent_bruno_memory_operations_total",
    "Memory operations performed",
    ["operation", "memory_type", "status"]  # operation: read, write; memory_type: conversation, user, entity
)

MEMORY_CONTEXT_BUILD_DURATION = Histogram(
    "agent_bruno_memory_context_build_seconds",
    "Time to build memory context for LLM",
    buckets=[0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0]
)

MEMORY_CONTEXT_SIZE = Histogram(
    "agent_bruno_memory_context_size_chars",
    "Size of memory context in characters",
    buckets=[100, 500, 1000, 2500, 5000, 10000, 25000]
)

USER_FACTS_RECORDED = Counter(
    "agent_bruno_user_facts_total",
    "User facts recorded to memory",
    ["source"]  # source: conversation, explicit
)

USER_PREFERENCES_UPDATED = Counter(
    "agent_bruno_user_preferences_total",
    "User preferences updated",
    ["explicit"]  # explicit: true, false
)

LEARNINGS_RECORDED = Counter(
    "agent_bruno_learnings_total",
    "Learnings recorded to long-term memory",
    ["category"]
)

# =============================================================================
# CLOUDEVENTS METRICS
# =============================================================================

EVENTS_PUBLISHED = Counter(
    "agent_bruno_events_published_total",
    "CloudEvents published",
    ["event_type", "status"]
)

EVENTS_RECEIVED = Counter(
    "agent_bruno_events_received_total",
    "CloudEvents received and processed",
    ["event_type", "status"]
)

EVENT_PROCESSING_DURATION = Histogram(
    "agent_bruno_event_processing_seconds",
    "Time to process received CloudEvents",
    ["event_type"],
    buckets=[0.01, 0.05, 0.1, 0.5, 1, 5, 10]
)

# =============================================================================
# HEALTH & AVAILABILITY METRICS
# =============================================================================

HEALTH_CHECK_DURATION = Summary(
    "agent_bruno_health_check_seconds",
    "Health check response time"
)

MEMORY_STORE_CONNECTED = Gauge(
    "agent_bruno_memory_store_connected",
    "Memory store connection status (1=connected, 0=disconnected)",
    ["store_type"]  # store_type: redis, postgres, in_memory
)

OLLAMA_AVAILABLE = Gauge(
    "agent_bruno_ollama_available",
    "Ollama LLM service availability (1=available, 0=unavailable)"
)

# =============================================================================
# INFO METRICS
# =============================================================================

BUILD_INFO = Info(
    "agent_bruno_build",
    "Build information"
)


def init_build_info(version: str, commit: str = "unknown", memory_enabled: bool = False):
    """Initialize build info metric."""
    BUILD_INFO.info({
        "version": version,
        "commit": commit,
        "memory_enabled": str(memory_enabled).lower(),
    })


def record_memory_operation(operation: str, memory_type: str, status: str = "success"):
    """Record a memory operation."""
    MEMORY_OPERATIONS.labels(
        operation=operation,
        memory_type=memory_type,
        status=status,
    ).inc()


def record_event_published(event_type: str, status: str = "success"):
    """Record a published CloudEvent."""
    EVENTS_PUBLISHED.labels(event_type=event_type, status=status).inc()


def record_event_received(event_type: str, status: str = "success"):
    """Record a received CloudEvent."""
    EVENTS_RECEIVED.labels(event_type=event_type, status=status).inc()


def set_memory_store_connected(store_type: str, connected: bool):
    """Set memory store connection status."""
    MEMORY_STORE_CONNECTED.labels(store_type=store_type).set(1 if connected else 0)


def set_ollama_available(available: bool):
    """Set Ollama availability status."""
    OLLAMA_AVAILABLE.set(1 if available else 0)


def init_metrics(models: list[str] = None):
    """
    Initialize all metrics with their label combinations.
    
    This pre-creates the metric series so they appear in Prometheus
    immediately, even before any actual traffic is recorded.
    
    Args:
        models: List of model names to initialize. Defaults to common models.
    """
    if models is None:
        models = ["llama3.2:3b", "llama3.2:1b", "mistral:7b", "unknown"]
    
    # Initialize message processing metrics
    for status in ["success", "error"]:
        MESSAGES_PROCESSED.labels(status=status)
    
    # Initialize performance metrics (histograms with model labels)
    for model in models:
        RESPONSE_DURATION.labels(model=model)
        LLM_INFERENCE_DURATION.labels(model=model)
    
    # Initialize token metrics
    for model in models:
        for token_type in ["input", "output", "total"]:
            TOKENS_USED.labels(model=model, type=token_type)
    
    # Initialize API call metrics
    for service in ["ollama"]:
        for status in ["success", "error", "timeout"]:
            API_CALLS.labels(service=service, status=status)
    
    # Initialize memory operation metrics
    for operation in ["read", "write"]:
        for memory_type in ["conversation", "user", "entity", "working"]:
            for status in ["success", "error"]:
                MEMORY_OPERATIONS.labels(
                    operation=operation,
                    memory_type=memory_type,
                    status=status,
                )
    
    # Initialize event metrics
    for event_type in [
        "io.homelab.chat.message",
        "io.homelab.chat.intent.security",
        "io.homelab.chat.intent.status",
        "io.homelab.vuln.found",
        "io.homelab.exploit.validated",
    ]:
        for status in ["success", "error"]:
            EVENTS_PUBLISHED.labels(event_type=event_type, status=status)
            EVENTS_RECEIVED.labels(event_type=event_type, status=status)
        EVENT_PROCESSING_DURATION.labels(event_type=event_type)
    
    # Initialize memory store connection metrics
    for store_type in ["redis", "postgres", "in_memory"]:
        MEMORY_STORE_CONNECTED.labels(store_type=store_type)
