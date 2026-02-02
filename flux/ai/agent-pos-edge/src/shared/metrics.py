"""Prometheus metrics for POS Edge agents."""
from prometheus_client import Counter, Histogram, Gauge, Info

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Common Metrics
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

EVENTS_RECEIVED = Counter(
    "pos_events_received_total",
    "Total CloudEvents received",
    ["event_type", "location_id", "agent_role"],
)

EVENTS_EMITTED = Counter(
    "pos_events_emitted_total",
    "Total CloudEvents emitted",
    ["event_type", "location_id", "agent_role"],
)

EVENT_PROCESSING_DURATION = Histogram(
    "pos_event_processing_seconds",
    "Event processing duration",
    ["event_type", "agent_role"],
    buckets=[0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0],
)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Transaction Metrics
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

TRANSACTIONS_TOTAL = Counter(
    "pos_transactions_total",
    "Total POS transactions",
    ["location_id", "status", "payment_type"],
)

TRANSACTION_VALUE = Counter(
    "pos_transaction_value_total",
    "Total transaction value",
    ["location_id", "payment_type"],
)

TRANSACTION_DURATION = Histogram(
    "pos_transaction_duration_seconds",
    "Transaction duration from start to complete",
    ["location_id"],
    buckets=[5, 10, 30, 60, 120, 300, 600],
)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Kitchen Metrics
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

KITCHEN_QUEUE_DEPTH = Gauge(
    "pos_kitchen_queue_depth",
    "Current kitchen queue depth",
    ["location_id"],
)

KITCHEN_AVG_WAIT = Gauge(
    "pos_kitchen_avg_wait_seconds",
    "Average kitchen wait time",
    ["location_id"],
)

KITCHEN_ORDERS_TOTAL = Counter(
    "pos_kitchen_orders_total",
    "Total kitchen orders",
    ["location_id", "status"],
)

KITCHEN_ORDER_DURATION = Histogram(
    "pos_kitchen_order_duration_seconds",
    "Kitchen order preparation time",
    ["location_id", "station"],
    buckets=[60, 120, 180, 240, 300, 420, 600],
)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Pump Metrics
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

PUMP_TRANSACTIONS_TOTAL = Counter(
    "pos_pump_transactions_total",
    "Total pump transactions",
    ["location_id", "pump_id", "fuel_type"],
)

PUMP_LITERS_TOTAL = Counter(
    "pos_pump_liters_total",
    "Total liters dispensed",
    ["location_id", "fuel_type"],
)

TANK_LEVEL_PERCENT = Gauge(
    "pos_tank_level_percent",
    "Current tank level percentage",
    ["location_id", "tank_id", "fuel_type"],
)

PUMP_STATUS_INFO = Gauge(
    "pos_pump_status",
    "Pump status (1=available, 2=in_use, 3=reserved, 0=offline)",
    ["location_id", "pump_id"],
)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Alert Metrics
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ALERTS_TOTAL = Counter(
    "pos_alerts_total",
    "Total alerts raised",
    ["location_id", "severity", "alert_type"],
)

ACTIVE_ALERTS = Gauge(
    "pos_active_alerts",
    "Currently active alerts",
    ["location_id", "severity"],
)

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Health Metrics
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

LOCATION_HEARTBEAT_TIMESTAMP = Gauge(
    "pos_location_heartbeat_timestamp",
    "Last heartbeat timestamp from location",
    ["location_id"],
)

LOCATION_STATUS = Gauge(
    "pos_location_status",
    "Location status (1=healthy, 0=unhealthy)",
    ["location_id", "location_type"],
)

# =============================================================================
# Build Info (for Agent Versions Dashboard)
# =============================================================================

BUILD_INFO = Info(
    "agent_pos_edge_build",
    "Build information"
)


def init_build_info(version: str, commit: str = "unknown"):
    """Initialize build info metric."""
    BUILD_INFO.info({
        "version": version,
        "commit": commit,
    })
