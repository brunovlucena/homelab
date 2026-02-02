"""
🔗 Agent Communication Observability

Inter-agent communication tracing, metrics, and logging.
Provides proof of communication through:
- Distributed tracing with OpenTelemetry
- Prometheus metrics for communication events
- Structured logging with correlation IDs
- CloudEvent context propagation
"""

from .observability import (
    # Core observability
    AgentCommunicationLogger,
    get_tracer,
    get_current_trace_context,
    
    # Metrics
    record_event_sent,
    record_event_received,
    record_event_processed,
    record_event_error,
    record_inter_agent_call,
    
    # Decorators
    trace_cloudevent_handler,
    trace_inter_agent_call,
    
    # Context managers
    trace_event_processing,
    
    # Build info
    init_agent_build_info,
    
    # Constants
    PROMETHEUS_AVAILABLE,
    OTEL_AVAILABLE,
)

__version__ = "1.0.0"
__all__ = [
    "AgentCommunicationLogger",
    "get_tracer",
    "get_current_trace_context",
    "record_event_sent",
    "record_event_received",
    "record_event_processed",
    "record_event_error",
    "record_inter_agent_call",
    "trace_cloudevent_handler",
    "trace_inter_agent_call",
    "trace_event_processing",
    "init_agent_build_info",
    "PROMETHEUS_AVAILABLE",
    "OTEL_AVAILABLE",
]
