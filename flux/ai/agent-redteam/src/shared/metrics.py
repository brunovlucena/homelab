"""
Prometheus metrics for agent-redteam.

Enhanced metrics for attack visibility, security monitoring, and operational health.
"""
from prometheus_client import Counter, Histogram, Gauge, Info, Summary

# =============================================================================
# EXPLOIT EXECUTION METRICS (Attack Visibility)
# =============================================================================

EXPLOITS_EXECUTED = Counter(
    "agent_redteam_exploits_executed_total",
    "Total exploits executed",
    ["exploit_id", "category", "severity", "status"]
)

EXPLOITS_SUCCESSFUL = Counter(
    "agent_redteam_exploits_successful_total",
    "Total successful exploits (unmitigated vulnerabilities)",
    ["exploit_id", "category", "severity"]
)

EXPLOITS_BLOCKED = Counter(
    "agent_redteam_exploits_blocked_total",
    "Total blocked exploits (mitigated vulnerabilities)",
    ["exploit_id", "category", "severity", "mitigated_by"]
)

ACTIVE_EXPLOITS = Gauge(
    "agent_redteam_active_exploits",
    "Currently running exploits"
)

# Attack metrics by severity (for quick dashboard views)
ATTACKS_BY_SEVERITY = Counter(
    "agent_redteam_attacks_by_severity_total",
    "Attack attempts by severity level",
    ["severity", "status"]  # status: success, blocked, failed, error
)

# Attack metrics by category (for attack pattern analysis)
ATTACKS_BY_CATEGORY = Counter(
    "agent_redteam_attacks_by_category_total",
    "Attack attempts by exploit category",
    ["category", "status"]
)

# Target component attacks (for understanding attack surface)
ATTACKS_BY_TARGET = Counter(
    "agent_redteam_attacks_by_target_total",
    "Attack attempts by target component",
    ["target_component", "status"]
)

# =============================================================================
# ATTACK PATTERN METRICS (Security Visibility)
# =============================================================================

# Random exploit executions (chaos testing visibility)
RANDOM_EXPLOITS_EXECUTED = Counter(
    "agent_redteam_random_exploits_total",
    "Random exploits executed (chaos testing)",
    ["exploit_id", "severity", "trigger"]  # trigger: vuln_found, manual, event
)

# Mitigation effectiveness
MITIGATIONS_OBSERVED = Counter(
    "agent_redteam_mitigations_total",
    "Security mitigations observed blocking exploits",
    ["mitigation_type", "exploit_category"]  # mitigation_type: admission_webhook, policy, rbac, etc.
)

# Attack success rate gauge (for SLO monitoring)
ATTACK_SUCCESS_RATE = Gauge(
    "agent_redteam_attack_success_rate",
    "Rolling attack success rate (lower is better security)",
    ["window"]  # window: 5m, 15m, 1h
)

# Critical vulnerabilities gauge
CRITICAL_VULNS_EXPOSED = Gauge(
    "agent_redteam_critical_vulnerabilities_exposed",
    "Critical severity vulnerabilities currently exploitable",
    ["target_namespace"]
)

# =============================================================================
# TEST RUN METRICS
# =============================================================================

TEST_RUNS_TOTAL = Counter(
    "agent_redteam_test_runs_total",
    "Total test runs executed",
    ["status", "test_type"]  # Added test_type: full-suite, category, severity
)

TEST_RUNS_ACTIVE = Gauge(
    "agent_redteam_test_runs_active",
    "Currently active test runs"
)

VULNERABILITIES_FOUND = Gauge(
    "agent_redteam_vulnerabilities_found",
    "Number of exploitable vulnerabilities found in last run",
    ["target_namespace"]
)

# K6 test runs triggered
K6_TESTS_TRIGGERED = Counter(
    "agent_redteam_k6_tests_triggered_total",
    "K6 test runs triggered by agent",
    ["test_type", "trigger_event", "status"]  # test_type: smoke, attack-sequential, etc.
)

# =============================================================================
# PERFORMANCE METRICS
# =============================================================================

EXPLOIT_DURATION = Histogram(
    "agent_redteam_exploit_duration_seconds",
    "Time to execute exploit",
    ["exploit_id", "category"],
    buckets=[1, 5, 10, 30, 60, 120, 300, 600]
)

EXPLOIT_DURATION_BY_SEVERITY = Histogram(
    "agent_redteam_exploit_duration_by_severity_seconds",
    "Exploit duration by severity level",
    ["severity"],
    buckets=[1, 5, 10, 30, 60, 120, 300, 600]
)

TEST_RUN_DURATION = Histogram(
    "agent_redteam_test_run_duration_seconds",
    "Total time for test run",
    buckets=[30, 60, 120, 300, 600, 1200, 1800, 3600]
)

# =============================================================================
# KUBERNETES OPERATIONS
# =============================================================================

K8S_OPERATIONS = Counter(
    "agent_redteam_k8s_operations_total",
    "Kubernetes operations executed",
    ["operation", "resource", "status"]  # operation: apply, delete, get; status: success, error
)

K8S_OPERATION_DURATION = Histogram(
    "agent_redteam_k8s_operation_duration_seconds",
    "Kubernetes operation duration",
    ["operation", "resource"],
    buckets=[0.1, 0.5, 1, 2, 5, 10, 30, 60]
)

K8S_ERRORS = Counter(
    "agent_redteam_k8s_errors_total",
    "Kubernetes API errors",
    ["operation", "resource", "error_type"]  # error_type: not_found, forbidden, timeout, etc.
)

# =============================================================================
# API/HTTP METRICS (Health Monitoring)
# =============================================================================

HTTP_REQUESTS = Counter(
    "agent_redteam_http_requests_total",
    "HTTP requests received",
    ["method", "endpoint", "status_code"]
)

HTTP_REQUEST_DURATION = Histogram(
    "agent_redteam_http_request_duration_seconds",
    "HTTP request duration",
    ["method", "endpoint"],
    buckets=[0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
)

HTTP_ERRORS = Counter(
    "agent_redteam_http_errors_total",
    "HTTP request errors",
    ["method", "endpoint", "error_type"]
)

# =============================================================================
# CLOUDEVENTS METRICS
# =============================================================================

CLOUDEVENTS_RECEIVED = Counter(
    "agent_redteam_cloudevents_received_total",
    "CloudEvents received by type",
    ["event_type", "source"]
)

CLOUDEVENTS_PROCESSED = Counter(
    "agent_redteam_cloudevents_processed_total",
    "CloudEvents processed",
    ["event_type", "status"]  # status: success, error, unknown_type
)

CLOUDEVENTS_PROCESSING_DURATION = Histogram(
    "agent_redteam_cloudevents_processing_seconds",
    "CloudEvent processing duration",
    ["event_type"],
    buckets=[0.1, 0.5, 1, 2, 5, 10, 30, 60]
)

# =============================================================================
# INFO METRICS
# =============================================================================

BUILD_INFO = Info(
    "agent_redteam_build",
    "Build information"
)

CATALOG_INFO = Info(
    "agent_redteam_catalog",
    "Exploit catalog information"
)

# =============================================================================
# GAUGE METRICS (Current State)
# =============================================================================

CATALOG_SIZE = Gauge(
    "agent_redteam_catalog_size",
    "Number of exploits in catalog",
    ["category", "severity"]
)

AGENT_UP = Gauge(
    "agent_redteam_up",
    "Agent is up and running (1=up, 0=down)"
)

AGENT_READY = Gauge(
    "agent_redteam_ready",
    "Agent is ready to process requests (1=ready, 0=not ready)"
)

DRY_RUN_MODE = Gauge(
    "agent_redteam_dry_run_mode",
    "Agent is running in dry-run mode (1=dry-run, 0=live)"
)


# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

def init_build_info(version: str, commit: str = "unknown"):
    """Initialize build info metric."""
    BUILD_INFO.info({
        "version": version,
        "commit": commit,
    })
    AGENT_UP.set(1)


def init_catalog_info(catalog_version: str, exploit_count: int):
    """Initialize catalog info metric."""
    CATALOG_INFO.info({
        "catalog_version": catalog_version,
        "exploit_count": str(exploit_count),
    })


def init_catalog_gauges(exploits: list):
    """Initialize catalog size gauges from exploit list."""
    from collections import defaultdict
    counts = defaultdict(lambda: defaultdict(int))
    
    for exploit in exploits:
        category = exploit.category.value if hasattr(exploit.category, 'value') else str(exploit.category)
        severity = exploit.severity.value if hasattr(exploit.severity, 'value') else str(exploit.severity)
        counts[category][severity] += 1
    
    for category, severities in counts.items():
        for severity, count in severities.items():
            CATALOG_SIZE.labels(category=category, severity=severity).set(count)


def record_attack_metrics(
    exploit_id: str,
    category: str,
    severity: str,
    status: str,
    target_component: str = "unknown",
    was_random: bool = False,
    trigger: str = "manual",
):
    """Record comprehensive attack metrics for a single exploit execution."""
    # Main execution counter
    EXPLOITS_EXECUTED.labels(
        exploit_id=exploit_id,
        category=category,
        severity=severity,
        status=status
    ).inc()
    
    # Severity breakdown
    ATTACKS_BY_SEVERITY.labels(severity=severity, status=status).inc()
    
    # Category breakdown
    ATTACKS_BY_CATEGORY.labels(category=category, status=status).inc()
    
    # Target component
    ATTACKS_BY_TARGET.labels(target_component=target_component, status=status).inc()
    
    # Random exploit tracking
    if was_random:
        RANDOM_EXPLOITS_EXECUTED.labels(
            exploit_id=exploit_id,
            severity=severity,
            trigger=trigger
        ).inc()
    
    # Status-specific counters
    if status == "success":
        EXPLOITS_SUCCESSFUL.labels(
            exploit_id=exploit_id,
            category=category,
            severity=severity
        ).inc()


def record_mitigation(mitigation_type: str, exploit_category: str):
    """Record a mitigation that blocked an exploit."""
    MITIGATIONS_OBSERVED.labels(
        mitigation_type=mitigation_type,
        exploit_category=exploit_category
    ).inc()


def record_cloudevent(event_type: str, source: str, status: str, duration: float = None):
    """Record CloudEvent processing metrics."""
    CLOUDEVENTS_RECEIVED.labels(event_type=event_type, source=source).inc()
    CLOUDEVENTS_PROCESSED.labels(event_type=event_type, status=status).inc()
    
    if duration is not None:
        CLOUDEVENTS_PROCESSING_DURATION.labels(event_type=event_type).observe(duration)
