"""
Prometheus metrics for agent-medical.
"""
from prometheus_client import Counter, Histogram, Gauge, Info

# Request metrics
REQUESTS_TOTAL = Counter(
    "agent_medical_requests_total",
    "Total number of medical requests",
    ["role", "status"]
)

ACCESS_DENIED_TOTAL = Counter(
    "agent_medical_access_denied_total",
    "Total number of access denied requests",
    ["reason"]
)

PATIENT_QUERIES_TOTAL = Counter(
    "agent_medical_patient_queries_total",
    "Total number of patient queries",
    ["patient_id_hash"]  # Hashed for privacy
)

RESPONSE_DURATION = Histogram(
    "agent_medical_response_duration_seconds",
    "Response duration in seconds",
    ["model"],
    buckets=[0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0]
)

AUDIT_LOGS_TOTAL = Counter(
    "agent_medical_audit_logs_total",
    "Total number of audit log entries",
    ["action"]
)

SUS_SYNC_TOTAL = Counter(
    "agent_medical_sus_sync_total",
    "Total number of SUS cloud sync operations",
    ["status"]
)

LLM_INFERENCE_DURATION = Histogram(
    "agent_medical_llm_inference_duration_seconds",
    "LLM inference duration",
    ["model"],
    buckets=[0.5, 1.0, 2.0, 5.0, 10.0, 30.0]
)

TOKENS_USED = Counter(
    "agent_medical_tokens_used_total",
    "Total tokens used",
    ["model", "type"]
)

ACTIVE_CONVERSATIONS = Gauge(
    "agent_medical_active_conversations",
    "Number of active conversations"
)

DB_QUERY_DURATION = Histogram(
    "agent_medical_db_query_duration_seconds",
    "Database query duration",
    ["operation"],
    buckets=[0.01, 0.05, 0.1, 0.5, 1.0, 2.0]
)


# =============================================================================
# BUILD INFO (for Agent Versions Dashboard)
# =============================================================================

BUILD_INFO = Info(
    "agent_medical_build",  # Will become agent_medical_build_info in Prometheus
    "Build information"
)


def init_build_info(version: str, commit: str):
    """Initialize build info metric."""
    BUILD_INFO.info({"version": version, "commit": commit})
