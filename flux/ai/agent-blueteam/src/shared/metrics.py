"""
Prometheus metrics for agent-blueteam defense runner.

🛡️ Blue Team Metrics - Measure your defenses!
"""
from prometheus_client import Counter, Gauge, Histogram, Info

# =============================================================================
# Request Metrics
# =============================================================================

REQUEST_COUNT = Counter(
    "blueteam_http_requests_total",
    "Total HTTP requests",
    ["endpoint", "method", "status"],
)

REQUEST_LATENCY = Histogram(
    "blueteam_http_request_duration_seconds",
    "HTTP request latency",
    ["endpoint", "method"],
    buckets=[0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0],
)

# =============================================================================
# CloudEvents Metrics
# =============================================================================

CLOUDEVENTS_RECEIVED = Counter(
    "blueteam_cloudevents_received_total",
    "Total CloudEvents received",
    ["event_type", "source"],
)

CLOUDEVENTS_PROCESSED = Counter(
    "blueteam_cloudevents_processed_total",
    "Total CloudEvents processed",
    ["event_type", "status"],
)

# =============================================================================
# Threat Detection Metrics
# =============================================================================

THREATS_DETECTED = Counter(
    "blueteam_threats_detected_total",
    "Total threats detected",
    ["threat_level", "exploit_id"],
)

THREATS_BLOCKED = Counter(
    "blueteam_threats_blocked_total",
    "Total threats blocked",
    ["action", "exploit_id"],
)

THREATS_MITIGATED = Counter(
    "blueteam_threats_mitigated_total",
    "Total threats mitigated",
    ["action", "exploit_id"],
)

# =============================================================================
# Defense Metrics
# =============================================================================

DEFENSE_ACTIVATIONS = Counter(
    "blueteam_defense_activations_total",
    "Total defense activations",
    ["action", "success"],
)

ACTIVE_DEFENSES = Gauge(
    "blueteam_active_defenses",
    "Number of active defense operations",
)

DEFENSE_DURATION = Histogram(
    "blueteam_defense_duration_seconds",
    "Defense execution duration",
    buckets=[0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0],
)

# =============================================================================
# MAG7 Boss Metrics
# =============================================================================

MAG7_HEALTH = Gauge(
    "blueteam_mag7_health",
    "MAG7 dragon boss health",
)

MAG7_DAMAGE_DEALT = Counter(
    "blueteam_mag7_damage_dealt_total",
    "Total damage dealt to MAG7",
    ["attack_type"],
)

MAG7_PHASE = Info(
    "blueteam_mag7_phase",
    "Current MAG7 boss phase",
)

# =============================================================================
# Game Metrics
# =============================================================================

GAME_SCORE = Gauge(
    "blueteam_game_score",
    "Current game score",
)

GAME_WAVE = Gauge(
    "blueteam_game_wave",
    "Current game wave",
)

EXPLOITS_BLOCKED_GAME = Counter(
    "blueteam_game_exploits_blocked_total",
    "Total exploits blocked in game",
    ["exploit_type"],
)

EXPLOITS_MISSED_GAME = Counter(
    "blueteam_game_exploits_missed_total",
    "Total exploits missed in game",
    ["exploit_type"],
)


# =============================================================================
# Build Info (for Agent Versions Dashboard)
# =============================================================================

BUILD_INFO = Info(
    "agent_blueteam_build",
    "Build information"
)


def init_build_info(version: str, commit: str = "unknown"):
    """Initialize build info metric."""
    BUILD_INFO.info({
        "version": version,
        "commit": commit,
    })


def init_mag7_health(health: int = 1000):
    """Initialize MAG7 health gauge."""
    MAG7_HEALTH.set(health)
    MAG7_PHASE.info({"phase": "normal"})
