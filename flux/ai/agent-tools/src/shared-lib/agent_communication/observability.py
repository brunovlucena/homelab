"""
🔗 Agent Communication Observability - Inter-Agent Tracing, Metrics, and Logging

Provides PROOF OF COMMUNICATION through:
1. Distributed tracing with OpenTelemetry (trace IDs propagate across agents)
2. Prometheus metrics for all communication events
3. Structured logging with correlation IDs
4. CloudEvent context propagation

Usage:
    from agent_communication import (
        AgentCommunicationLogger,
        record_event_received,
        trace_cloudevent_handler,
    )
    
    logger = AgentCommunicationLogger("agent-bruno")
    
    @trace_cloudevent_handler("io.homelab.chat.message")
    async def handle_chat(event):
        logger.event_received(event)
        # ... process event ...
        logger.event_processed(event, success=True)
"""

import functools
import time
import uuid
from contextlib import asynccontextmanager, contextmanager
from datetime import datetime, timezone
from typing import Any, Callable, Optional, Dict

import structlog

# Prometheus metrics
try:
    from prometheus_client import Counter, Histogram, Gauge, Info
    PROMETHEUS_AVAILABLE = True
except ImportError:
    PROMETHEUS_AVAILABLE = False

# OpenTelemetry tracing
try:
    from opentelemetry import trace
    from opentelemetry.trace import Status, StatusCode, SpanKind
    from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator
    OTEL_AVAILABLE = True
except ImportError:
    OTEL_AVAILABLE = False

logger = structlog.get_logger()

# =============================================================================
# PROMETHEUS METRICS - Inter-Agent Communication
# =============================================================================

if PROMETHEUS_AVAILABLE:
    # CloudEvent metrics
    EVENTS_RECEIVED = Counter(
        "agent_events_received_total",
        "Total CloudEvents received by this agent",
        ["agent_id", "event_type", "source_agent", "status"]
    )
    
    EVENTS_SENT = Counter(
        "agent_events_sent_total",
        "Total CloudEvents sent by this agent",
        ["agent_id", "event_type", "target_agent", "status"]
    )
    
    EVENT_PROCESSING_DURATION = Histogram(
        "agent_event_processing_seconds",
        "CloudEvent processing duration",
        ["agent_id", "event_type"],
        buckets=[0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0]
    )
    
    INTER_AGENT_CALLS = Counter(
        "agent_inter_agent_calls_total",
        "Total inter-agent HTTP/RPC calls",
        ["source_agent", "target_agent", "operation", "status"]
    )
    
    INTER_AGENT_LATENCY = Histogram(
        "agent_inter_agent_latency_seconds",
        "Inter-agent call latency",
        ["source_agent", "target_agent", "operation"],
        buckets=[0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0]
    )
    
    # Communication health
    AGENT_COMMUNICATION_ERRORS = Counter(
        "agent_communication_errors_total",
        "Total communication errors",
        ["agent_id", "error_type", "target"]
    )
    
    AGENT_ACTIVE_CONNECTIONS = Gauge(
        "agent_active_connections",
        "Number of active connections to other agents",
        ["agent_id", "target_agent"]
    )
    
    # Build info
    AGENT_BUILD_INFO = Info(
        "agent_build",
        "Agent build information"
    )


# =============================================================================
# OPENTELEMETRY TRACING
# =============================================================================

def get_tracer(name: str = "agent_communication"):
    """Get OpenTelemetry tracer."""
    if OTEL_AVAILABLE:
        return trace.get_tracer(name, "1.0.0")
    return None


def get_current_trace_context() -> Dict[str, Any]:
    """Get current trace context for logging and propagation."""
    if not OTEL_AVAILABLE:
        return {"correlation_id": str(uuid.uuid4())[:8]}
    
    span = trace.get_current_span()
    if span and span.is_recording():
        ctx = span.get_span_context()
        return {
            "trace_id": format(ctx.trace_id, "032x"),
            "span_id": format(ctx.span_id, "016x"),
            "trace_flags": ctx.trace_flags,
        }
    return {"correlation_id": str(uuid.uuid4())[:8]}


def extract_trace_from_cloudevent(event) -> Dict[str, str]:
    """Extract trace context from CloudEvent extensions."""
    trace_ctx = {}
    
    # Check for W3C trace context in CloudEvent extensions
    if hasattr(event, 'get'):
        trace_ctx['traceparent'] = event.get('traceparent', '')
        trace_ctx['tracestate'] = event.get('tracestate', '')
    
    # Also check for custom correlation ID
    if hasattr(event, '__getitem__'):
        try:
            trace_ctx['correlation_id'] = event.get('correlationid', event.get('id', ''))
        except Exception:
            pass
    
    return trace_ctx


def inject_trace_into_cloudevent(event_attributes: Dict) -> Dict:
    """Inject current trace context into CloudEvent attributes."""
    if not OTEL_AVAILABLE:
        event_attributes['correlationid'] = str(uuid.uuid4())[:8]
        return event_attributes
    
    carrier = {}
    propagator = TraceContextTextMapPropagator()
    propagator.inject(carrier)
    
    if 'traceparent' in carrier:
        event_attributes['traceparent'] = carrier['traceparent']
    if 'tracestate' in carrier:
        event_attributes['tracestate'] = carrier['tracestate']
    
    return event_attributes


# =============================================================================
# CONTEXT MANAGERS FOR TRACING
# =============================================================================

@contextmanager
def trace_event_processing(
    agent_id: str,
    event_type: str,
    event_id: str,
    source: str = "unknown",
):
    """
    Context manager for tracing CloudEvent processing.
    
    Usage:
        with trace_event_processing("agent-bruno", "io.homelab.chat.message", event.id, event.source):
            # Process event
            pass
    """
    tracer = get_tracer()
    start_time = time.perf_counter()
    status = "success"
    
    span_context = None
    if tracer:
        span_context = tracer.start_as_current_span(
            f"process.{event_type}",
            kind=SpanKind.CONSUMER,
            attributes={
                "agent.id": agent_id,
                "cloudevent.type": event_type,
                "cloudevent.id": event_id,
                "cloudevent.source": source,
            }
        )
        span_context.__enter__()
    
    try:
        yield
    except Exception as e:
        status = "error"
        if span_context:
            span = trace.get_current_span()
            span.set_status(Status(StatusCode.ERROR, str(e)))
            span.record_exception(e)
        raise
    finally:
        duration = time.perf_counter() - start_time
        
        if span_context:
            span = trace.get_current_span()
            span.set_attribute("processing.duration_ms", duration * 1000)
            if status == "success":
                span.set_status(Status(StatusCode.OK))
            span_context.__exit__(None, None, None)
        
        # Record metrics
        if PROMETHEUS_AVAILABLE:
            source_agent = _extract_agent_from_source(source)
            EVENTS_RECEIVED.labels(
                agent_id=agent_id,
                event_type=event_type,
                source_agent=source_agent,
                status=status,
            ).inc()
            EVENT_PROCESSING_DURATION.labels(
                agent_id=agent_id,
                event_type=event_type,
            ).observe(duration)


@asynccontextmanager
async def trace_inter_agent_request(
    source_agent: str,
    target_agent: str,
    operation: str,
    attributes: Dict = None,
):
    """
    Async context manager for tracing inter-agent HTTP/RPC calls.
    
    Usage:
        async with trace_inter_agent_request("agent-bruno", "agent-contracts", "scan"):
            response = await http_client.post(...)
    """
    tracer = get_tracer()
    start_time = time.perf_counter()
    status = "success"
    
    if tracer:
        with tracer.start_as_current_span(
            f"call.{target_agent}.{operation}",
            kind=SpanKind.CLIENT,
            attributes={
                "agent.source": source_agent,
                "agent.target": target_agent,
                "operation": operation,
                **(attributes or {}),
            }
        ) as span:
            try:
                yield span
                span.set_status(Status(StatusCode.OK))
            except Exception as e:
                status = "error"
                span.set_status(Status(StatusCode.ERROR, str(e)))
                span.record_exception(e)
                raise
            finally:
                duration = time.perf_counter() - start_time
                span.set_attribute("duration_ms", duration * 1000)
                
                if PROMETHEUS_AVAILABLE:
                    INTER_AGENT_CALLS.labels(
                        source_agent=source_agent,
                        target_agent=target_agent,
                        operation=operation,
                        status=status,
                    ).inc()
                    INTER_AGENT_LATENCY.labels(
                        source_agent=source_agent,
                        target_agent=target_agent,
                        operation=operation,
                    ).observe(duration)
    else:
        try:
            yield None
            status = "success"
        except Exception:
            status = "error"
            raise
        finally:
            duration = time.perf_counter() - start_time
            if PROMETHEUS_AVAILABLE:
                INTER_AGENT_CALLS.labels(
                    source_agent=source_agent,
                    target_agent=target_agent,
                    operation=operation,
                    status=status,
                ).inc()
                INTER_AGENT_LATENCY.labels(
                    source_agent=source_agent,
                    target_agent=target_agent,
                    operation=operation,
                ).observe(duration)


# =============================================================================
# DECORATORS
# =============================================================================

def trace_cloudevent_handler(event_type: str, agent_id: str = None):
    """
    Decorator for tracing CloudEvent handlers.
    
    Usage:
        @trace_cloudevent_handler("io.homelab.chat.message", "agent-bruno")
        async def handle_chat(event):
            pass
    """
    def decorator(func: Callable):
        @functools.wraps(func)
        async def async_wrapper(event, *args, **kwargs):
            aid = agent_id or kwargs.get('agent_id', 'unknown')
            eid = getattr(event, 'id', str(uuid.uuid4())[:8])
            src = getattr(event, 'source', 'unknown')
            
            with trace_event_processing(aid, event_type, eid, src):
                return await func(event, *args, **kwargs)
        
        @functools.wraps(func)
        def sync_wrapper(event, *args, **kwargs):
            aid = agent_id or kwargs.get('agent_id', 'unknown')
            eid = getattr(event, 'id', str(uuid.uuid4())[:8])
            src = getattr(event, 'source', 'unknown')
            
            with trace_event_processing(aid, event_type, eid, src):
                return func(event, *args, **kwargs)
        
        # Return appropriate wrapper based on function type
        import asyncio
        if asyncio.iscoroutinefunction(func):
            return async_wrapper
        return sync_wrapper
    return decorator


def trace_inter_agent_call(target_agent: str, operation: str):
    """
    Decorator for tracing inter-agent calls.
    
    Usage:
        @trace_inter_agent_call("agent-contracts", "scan")
        async def call_contracts_scan(data):
            pass
    """
    def decorator(func: Callable):
        @functools.wraps(func)
        async def wrapper(self, *args, **kwargs):
            source_agent = getattr(self, 'agent_id', 'unknown')
            async with trace_inter_agent_request(source_agent, target_agent, operation):
                return await func(self, *args, **kwargs)
        return wrapper
    return decorator


# =============================================================================
# METRIC RECORDING FUNCTIONS
# =============================================================================

def record_event_received(
    agent_id: str,
    event_type: str,
    source: str,
    status: str = "success"
):
    """Record a received CloudEvent."""
    if PROMETHEUS_AVAILABLE:
        source_agent = _extract_agent_from_source(source)
        EVENTS_RECEIVED.labels(
            agent_id=agent_id,
            event_type=event_type,
            source_agent=source_agent,
            status=status,
        ).inc()


def record_event_sent(
    agent_id: str,
    event_type: str,
    target: str,
    status: str = "success"
):
    """Record a sent CloudEvent."""
    if PROMETHEUS_AVAILABLE:
        target_agent = _extract_agent_from_source(target)
        EVENTS_SENT.labels(
            agent_id=agent_id,
            event_type=event_type,
            target_agent=target_agent,
            status=status,
        ).inc()


def record_event_processed(
    agent_id: str,
    event_type: str,
    duration_seconds: float,
    success: bool = True
):
    """Record event processing completion."""
    if PROMETHEUS_AVAILABLE:
        EVENT_PROCESSING_DURATION.labels(
            agent_id=agent_id,
            event_type=event_type,
        ).observe(duration_seconds)


def record_event_error(
    agent_id: str,
    error_type: str,
    target: str = "unknown"
):
    """Record a communication error."""
    if PROMETHEUS_AVAILABLE:
        AGENT_COMMUNICATION_ERRORS.labels(
            agent_id=agent_id,
            error_type=error_type,
            target=target,
        ).inc()


def record_inter_agent_call(
    source_agent: str,
    target_agent: str,
    operation: str,
    duration_seconds: float,
    status: str = "success"
):
    """Record an inter-agent call."""
    if PROMETHEUS_AVAILABLE:
        INTER_AGENT_CALLS.labels(
            source_agent=source_agent,
            target_agent=target_agent,
            operation=operation,
            status=status,
        ).inc()
        INTER_AGENT_LATENCY.labels(
            source_agent=source_agent,
            target_agent=target_agent,
            operation=operation,
        ).observe(duration_seconds)


def init_agent_build_info(
    agent_id: str,
    version: str,
    commit: str = "unknown",
    **extra
):
    """Initialize agent build info metric."""
    if PROMETHEUS_AVAILABLE:
        info = {
            "agent_id": agent_id,
            "version": version,
            "commit": commit,
            **extra
        }
        AGENT_BUILD_INFO.info(info)


# =============================================================================
# STRUCTURED LOGGING
# =============================================================================

class AgentCommunicationLogger:
    """
    Structured logger for inter-agent communication.
    
    Ensures all communication logs have:
    - Trace context (when available)
    - Correlation IDs
    - Consistent field names
    - CloudEvent metadata
    
    Usage:
        logger = AgentCommunicationLogger("agent-bruno")
        logger.event_received(event)
        logger.event_sent("io.homelab.alert", target="agent-contracts")
        logger.inter_agent_call_start("agent-contracts", "scan")
    """
    
    def __init__(self, agent_id: str, component: str = "communication"):
        self.agent_id = agent_id
        self.component = component
        self._logger = structlog.get_logger()
    
    def _enrich(self, **kwargs) -> Dict:
        """Enrich log with standard fields and trace context."""
        enriched = {
            "agent_id": self.agent_id,
            "component": self.component,
            "timestamp": datetime.now(timezone.utc).isoformat(),
            **get_current_trace_context(),
            **kwargs,
        }
        return enriched
    
    def debug(self, event: str, **kwargs):
        self._logger.debug(event, **self._enrich(**kwargs))
    
    def info(self, event: str, **kwargs):
        self._logger.info(event, **self._enrich(**kwargs))
    
    def warning(self, event: str, **kwargs):
        self._logger.warning(event, **self._enrich(**kwargs))
    
    def error(self, event: str, **kwargs):
        self._logger.error(event, **self._enrich(**kwargs))
    
    # Communication-specific log methods
    def event_received(
        self,
        event,
        source_type: str = "cloudevent",
    ):
        """Log a received CloudEvent - PROOF OF COMMUNICATION."""
        event_type = getattr(event, 'type', event.get('type', 'unknown') if hasattr(event, 'get') else 'unknown')
        event_id = getattr(event, 'id', event.get('id', 'unknown') if hasattr(event, 'get') else 'unknown')
        source = getattr(event, 'source', event.get('source', 'unknown') if hasattr(event, 'get') else 'unknown')
        
        self.info(
            "cloudevent_received",
            event_type=event_type,
            event_id=event_id,
            source=source,
            source_type=source_type,
            subject=getattr(event, 'subject', None),
        )
    
    def event_processed(
        self,
        event,
        success: bool,
        duration_ms: float,
        result: str = None,
    ):
        """Log CloudEvent processing completion - PROOF OF PROCESSING."""
        event_type = getattr(event, 'type', event.get('type', 'unknown') if hasattr(event, 'get') else 'unknown')
        event_id = getattr(event, 'id', event.get('id', 'unknown') if hasattr(event, 'get') else 'unknown')
        
        self.info(
            "cloudevent_processed",
            event_type=event_type,
            event_id=event_id,
            success=success,
            duration_ms=round(duration_ms, 3),
            result=result,
        )
    
    def event_sent(
        self,
        event_type: str,
        event_id: str,
        target: str,
        subject: str = None,
    ):
        """Log a sent CloudEvent - PROOF OF OUTBOUND COMMUNICATION."""
        self.info(
            "cloudevent_sent",
            event_type=event_type,
            event_id=event_id,
            target=target,
            subject=subject,
        )
    
    def event_error(
        self,
        event,
        error: str,
        duration_ms: float = 0,
    ):
        """Log CloudEvent processing error."""
        event_type = getattr(event, 'type', 'unknown')
        event_id = getattr(event, 'id', 'unknown')
        
        self.error(
            "cloudevent_processing_failed",
            event_type=event_type,
            event_id=event_id,
            error=error,
            duration_ms=round(duration_ms, 3),
        )
    
    def inter_agent_call_start(
        self,
        target_agent: str,
        operation: str,
        **kwargs
    ):
        """Log start of inter-agent call."""
        self.info(
            "inter_agent_call_start",
            target_agent=target_agent,
            operation=operation,
            **kwargs,
        )
    
    def inter_agent_call_complete(
        self,
        target_agent: str,
        operation: str,
        duration_ms: float,
        success: bool,
        **kwargs
    ):
        """Log completion of inter-agent call - PROOF OF INTER-AGENT COMMUNICATION."""
        self.info(
            "inter_agent_call_complete",
            target_agent=target_agent,
            operation=operation,
            duration_ms=round(duration_ms, 3),
            success=success,
            **kwargs,
        )
    
    def broker_event_published(
        self,
        event_type: str,
        broker: str,
        event_id: str = None,
    ):
        """Log event published to broker."""
        self.info(
            "broker_event_published",
            event_type=event_type,
            broker=broker,
            event_id=event_id,
        )


# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

def _extract_agent_from_source(source: str) -> str:
    """Extract agent name from CloudEvent source."""
    if not source:
        return "unknown"
    
    # Handle formats like:
    # /agent-bruno/chatbot
    # /test/proof-audit
    # agent-contracts/contract-fetcher
    parts = source.strip('/').split('/')
    if parts:
        first_part = parts[0]
        if first_part.startswith('agent-'):
            return first_part
        # Try second part
        if len(parts) > 1 and parts[1].startswith('agent-'):
            return parts[1]
    return source.strip('/')[:30]  # Return truncated source
