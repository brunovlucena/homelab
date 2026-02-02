"""
🔭 Shared Observability Module - OpenTelemetry Initialization with Pydantic Settings

Provides type-safe observability configuration and automatic OpenTelemetry initialization
for all homelab agents.

Features:
- Type-safe configuration using pydantic-settings
- Automatic OpenTelemetry exporter initialization
- Tempo tracing support via Grafana Alloy
- Prometheus metrics integration
- Structured logging with trace context

Usage:
    from observability import initialize_observability, ObservabilitySettings
    
    # Initialize with defaults from environment variables
    settings = ObservabilitySettings(
        otel_service_name="agent-bruno",
        otel_service_namespace="agent-bruno",
    )
    initialize_observability(settings)
    
    # Or use environment variables directly (OTEL_EXPORTER_OTLP_ENDPOINT, etc.)
    initialize_observability()  # Reads from environment
"""

from .config import ObservabilitySettings
from .init import (
    initialize_observability,
    initialize_tracing,
    initialize_metrics,
    get_tracer,
    get_meter,
    is_observability_enabled,
    get_settings,
    get_current_trace_context,
)

__all__ = [
    "ObservabilitySettings",
    "initialize_observability",
    "initialize_tracing",
    "initialize_metrics",
    "get_tracer",
    "get_meter",
    "is_observability_enabled",
    "get_settings",
    "get_current_trace_context",
]

__version__ = "1.0.0"


