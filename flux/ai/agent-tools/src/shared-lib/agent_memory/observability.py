"""
🔭 Agent Memory Observability - Metrics, Tracing, and Structured Logging

Principal SRE Engineering best practices for memory system observability:
- RED metrics (Rate, Errors, Duration) for all memory operations
- OpenTelemetry distributed tracing
- Structured logs with trace context propagation
- Memory health and capacity metrics

Metric Naming Convention:
    agent_memory_{subsystem}_{metric}_{unit}
    
Examples:
    agent_memory_store_operations_total
    agent_memory_store_latency_seconds
    agent_memory_conversation_messages_total
"""

import functools
import time
from contextlib import asynccontextmanager
from datetime import datetime, timezone
from typing import Any, Callable, Optional

import structlog

# Prometheus metrics
try:
    from prometheus_client import Counter, Histogram, Gauge, Info, Summary
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
# PROMETHEUS METRICS - Memory Operations
# =============================================================================

if PROMETHEUS_AVAILABLE:
    # Store Operations
    MEMORY_STORE_OPERATIONS = Counter(
        "agent_memory_store_operations_total",
        "Total memory store operations",
        ["agent_id", "store_type", "operation", "status"]
    )
    
    MEMORY_STORE_LATENCY = Histogram(
        "agent_memory_store_latency_seconds",
        "Memory store operation latency",
        ["agent_id", "store_type", "operation"],
        buckets=[0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0]
    )
    
    MEMORY_STORE_SIZE = Gauge(
        "agent_memory_store_entries_count",
        "Number of entries in memory store",
        ["agent_id", "store_type", "memory_type"]
    )
    
    MEMORY_STORE_BYTES = Gauge(
        "agent_memory_store_bytes",
        "Estimated memory size in bytes",
        ["agent_id", "store_type"]
    )
    
    # Connection Health
    MEMORY_STORE_CONNECTED = Gauge(
        "agent_memory_store_connected",
        "Memory store connection status (1=connected, 0=disconnected)",
        ["agent_id", "store_type"]
    )
    
    MEMORY_STORE_ERRORS = Counter(
        "agent_memory_store_errors_total",
        "Memory store error count",
        ["agent_id", "store_type", "error_type"]
    )
    
    # Conversation Metrics
    MEMORY_CONVERSATIONS_ACTIVE = Gauge(
        "agent_memory_conversations_active",
        "Currently active conversations",
        ["agent_id"]
    )
    
    MEMORY_CONVERSATION_MESSAGES = Counter(
        "agent_memory_conversation_messages_total",
        "Total messages stored in conversations",
        ["agent_id", "role"]  # role: user, assistant, system
    )
    
    MEMORY_CONVERSATION_LENGTH = Histogram(
        "agent_memory_conversation_length_messages",
        "Number of messages per conversation",
        ["agent_id"],
        buckets=[1, 5, 10, 20, 50, 100, 200, 500]
    )
    
    # User Memory Metrics
    MEMORY_USER_FACTS = Counter(
        "agent_memory_user_facts_total",
        "Total user facts recorded",
        ["agent_id", "source"]
    )
    
    MEMORY_USER_PREFERENCES = Counter(
        "agent_memory_user_preferences_total",
        "Total user preferences updated",
        ["agent_id", "explicit"]  # explicit: true, false
    )
    
    # Entity Memory Metrics
    MEMORY_ENTITIES = Gauge(
        "agent_memory_entities_count",
        "Number of entities in memory",
        ["agent_id", "entity_type"]
    )
    
    MEMORY_ENTITY_RELATIONSHIPS = Counter(
        "agent_memory_entity_relationships_total",
        "Total entity relationships created",
        ["agent_id", "relation_type"]
    )
    
    # Long-term Memory Metrics
    MEMORY_LEARNINGS = Counter(
        "agent_memory_learnings_total",
        "Total learnings recorded",
        ["agent_id", "category"]
    )
    
    MEMORY_PATTERNS = Gauge(
        "agent_memory_patterns_count",
        "Number of patterns discovered",
        ["agent_id"]
    )
    
    MEMORY_ERRORS_RECORDED = Counter(
        "agent_memory_errors_recorded_total",
        "Total error patterns recorded",
        ["agent_id", "severity"]
    )
    
    # Task/Schema Metrics
    MEMORY_TASKS_CREATED = Counter(
        "agent_memory_tasks_created_total",
        "Total tasks created",
        ["agent_id", "schema_type"]
    )
    
    MEMORY_TASKS_COMPLETED = Counter(
        "agent_memory_tasks_completed_total",
        "Total tasks completed",
        ["agent_id", "success"]  # success: true, false
    )
    
    MEMORY_TASK_DURATION = Histogram(
        "agent_memory_task_duration_seconds",
        "Task execution duration",
        ["agent_id", "schema_type"],
        buckets=[0.1, 0.5, 1, 5, 10, 30, 60, 120, 300, 600]
    )
    
    # Cache/Hit Metrics
    MEMORY_CACHE_HITS = Counter(
        "agent_memory_cache_hits_total",
        "Memory cache hits",
        ["agent_id", "memory_type"]
    )
    
    MEMORY_CACHE_MISSES = Counter(
        "agent_memory_cache_misses_total",
        "Memory cache misses",
        ["agent_id", "memory_type"]
    )
    
    # Context Building Metrics
    MEMORY_CONTEXT_BUILD_DURATION = Histogram(
        "agent_memory_context_build_seconds",
        "Time to build memory context for LLM",
        ["agent_id"],
        buckets=[0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0]
    )
    
    MEMORY_CONTEXT_SIZE = Histogram(
        "agent_memory_context_size_chars",
        "Size of built context in characters",
        ["agent_id"],
        buckets=[100, 500, 1000, 2500, 5000, 10000, 25000, 50000]
    )
    
    # Build Info
    MEMORY_BUILD_INFO = Info(
        "agent_memory_build",
        "Agent memory library build information"
    )


# =============================================================================
# OPENTELEMETRY TRACING
# =============================================================================

def get_tracer(name: str = "agent_memory"):
    """Get OpenTelemetry tracer."""
    if OTEL_AVAILABLE:
        return trace.get_tracer(name, "1.0.0")
    return None


def get_current_trace_context() -> dict:
    """Get current trace context for logging."""
    if not OTEL_AVAILABLE:
        return {}
    
    span = trace.get_current_span()
    if span and span.is_recording():
        ctx = span.get_span_context()
        return {
            "trace_id": format(ctx.trace_id, "032x"),
            "span_id": format(ctx.span_id, "016x"),
            "trace_flags": ctx.trace_flags,
        }
    return {}


@asynccontextmanager
async def trace_memory_operation(
    operation: str,
    agent_id: str,
    store_type: str = "unknown",
    attributes: dict = None,
):
    """
    Context manager for tracing memory operations.
    
    Usage:
        async with trace_memory_operation("save", "agent-bruno", "redis") as span:
            await store.save(entry)
            span.set_attribute("entry.id", entry.id)
    """
    tracer = get_tracer()
    
    if tracer:
        with tracer.start_as_current_span(
            f"memory.{operation}",
            kind=SpanKind.INTERNAL,
            attributes={
                "memory.agent_id": agent_id,
                "memory.store_type": store_type,
                "memory.operation": operation,
                **(attributes or {}),
            }
        ) as span:
            start_time = time.perf_counter()
            try:
                yield span
                span.set_status(Status(StatusCode.OK))
                status = "success"
            except Exception as e:
                span.set_status(Status(StatusCode.ERROR, str(e)))
                span.record_exception(e)
                status = "error"
                raise
            finally:
                duration = time.perf_counter() - start_time
                span.set_attribute("memory.duration_ms", duration * 1000)
                
                # Record metrics
                if PROMETHEUS_AVAILABLE:
                    MEMORY_STORE_OPERATIONS.labels(
                        agent_id=agent_id,
                        store_type=store_type,
                        operation=operation,
                        status=status,
                    ).inc()
                    MEMORY_STORE_LATENCY.labels(
                        agent_id=agent_id,
                        store_type=store_type,
                        operation=operation,
                    ).observe(duration)
    else:
        # No tracing available, just record metrics
        start_time = time.perf_counter()
        try:
            yield None
            status = "success"
        except Exception:
            status = "error"
            raise
        finally:
            duration = time.perf_counter() - start_time
            if PROMETHEUS_AVAILABLE:
                MEMORY_STORE_OPERATIONS.labels(
                    agent_id=agent_id,
                    store_type=store_type,
                    operation=operation,
                    status=status,
                ).inc()
                MEMORY_STORE_LATENCY.labels(
                    agent_id=agent_id,
                    store_type=store_type,
                    operation=operation,
                ).observe(duration)


def trace_async(
    operation: str = None,
    record_args: list[str] = None,
):
    """
    Decorator for tracing async memory operations.
    
    Usage:
        @trace_async("conversation.start", record_args=["user_id"])
        async def start_conversation(self, user_id: str): ...
    """
    def decorator(func: Callable):
        @functools.wraps(func)
        async def wrapper(*args, **kwargs):
            tracer = get_tracer()
            op_name = operation or f"{func.__module__}.{func.__name__}"
            
            if tracer:
                attributes = {}
                if record_args:
                    for arg_name in record_args:
                        if arg_name in kwargs:
                            value = kwargs[arg_name]
                            if isinstance(value, (str, int, float, bool)):
                                attributes[f"memory.{arg_name}"] = value
                
                with tracer.start_as_current_span(
                    op_name,
                    kind=SpanKind.INTERNAL,
                    attributes=attributes,
                ) as span:
                    try:
                        result = await func(*args, **kwargs)
                        span.set_status(Status(StatusCode.OK))
                        return result
                    except Exception as e:
                        span.set_status(Status(StatusCode.ERROR, str(e)))
                        span.record_exception(e)
                        raise
            else:
                return await func(*args, **kwargs)
        return wrapper
    return decorator


# =============================================================================
# STRUCTURED LOGGING WITH TRACE CONTEXT
# =============================================================================

class MemoryLogger:
    """
    Structured logger for memory operations with automatic trace context.
    
    Ensures all memory-related logs have:
    - Consistent field names
    - Trace context (when available)
    - Agent identification
    - Operation timing
    """
    
    def __init__(self, agent_id: str, component: str = "memory"):
        self.agent_id = agent_id
        self.component = component
        self._logger = structlog.get_logger()
    
    def _enrich(self, **kwargs) -> dict:
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
    
    # Memory-specific log methods
    def store_operation(
        self,
        operation: str,
        store_type: str,
        duration_ms: float,
        status: str = "success",
        entry_id: str = None,
        memory_type: str = None,
        **kwargs,
    ):
        """Log a memory store operation."""
        self.info(
            "memory_store_operation",
            operation=operation,
            store_type=store_type,
            duration_ms=round(duration_ms, 3),
            status=status,
            entry_id=entry_id,
            memory_type=memory_type,
            **kwargs,
        )
    
    def conversation_event(
        self,
        event_type: str,
        conversation_id: str,
        user_id: str = None,
        message_count: int = None,
        **kwargs,
    ):
        """Log a conversation event."""
        self.info(
            f"memory_conversation_{event_type}",
            conversation_id=conversation_id,
            user_id=user_id,
            message_count=message_count,
            **kwargs,
        )
    
    def user_memory_event(
        self,
        event_type: str,
        user_id: str,
        **kwargs,
    ):
        """Log a user memory event."""
        self.info(
            f"memory_user_{event_type}",
            user_id=user_id,
            **kwargs,
        )
    
    def entity_event(
        self,
        event_type: str,
        entity_type: str,
        entity_id: str,
        **kwargs,
    ):
        """Log an entity memory event."""
        self.info(
            f"memory_entity_{event_type}",
            entity_type=entity_type,
            entity_id=entity_id,
            **kwargs,
        )
    
    def task_event(
        self,
        event_type: str,
        task_id: str = None,
        schema_type: str = None,
        duration_ms: float = None,
        success: bool = None,
        **kwargs,
    ):
        """Log a task/schema event."""
        self.info(
            f"memory_task_{event_type}",
            task_id=task_id,
            schema_type=schema_type,
            duration_ms=duration_ms,
            success=success,
            **kwargs,
        )
    
    def learning_event(
        self,
        category: str,
        content_preview: str,
        source: str = None,
        **kwargs,
    ):
        """Log a learning event."""
        self.info(
            "memory_learning_recorded",
            category=category,
            content_preview=content_preview[:100] if content_preview else None,
            source=source,
            **kwargs,
        )
    
    def context_built(
        self,
        user_id: str = None,
        conversation_id: str = None,
        context_size_chars: int = None,
        duration_ms: float = None,
        **kwargs,
    ):
        """Log context building for LLM."""
        self.info(
            "memory_context_built",
            user_id=user_id,
            conversation_id=conversation_id,
            context_size_chars=context_size_chars,
            duration_ms=round(duration_ms, 3) if duration_ms else None,
            **kwargs,
        )


# =============================================================================
# METRICS HELPER FUNCTIONS
# =============================================================================

def record_store_connected(agent_id: str, store_type: str, connected: bool):
    """Record store connection status."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_STORE_CONNECTED.labels(
            agent_id=agent_id,
            store_type=store_type,
        ).set(1 if connected else 0)


def record_store_error(agent_id: str, store_type: str, error_type: str):
    """Record a store error."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_STORE_ERRORS.labels(
            agent_id=agent_id,
            store_type=store_type,
            error_type=error_type,
        ).inc()


def record_conversation_message(agent_id: str, role: str):
    """Record a conversation message."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_CONVERSATION_MESSAGES.labels(
            agent_id=agent_id,
            role=role,
        ).inc()


def record_conversation_length(agent_id: str, length: int):
    """Record conversation length."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_CONVERSATION_LENGTH.labels(agent_id=agent_id).observe(length)


def record_user_fact(agent_id: str, source: str):
    """Record a user fact."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_USER_FACTS.labels(agent_id=agent_id, source=source).inc()


def record_user_preference(agent_id: str, explicit: bool):
    """Record a user preference update."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_USER_PREFERENCES.labels(
            agent_id=agent_id,
            explicit=str(explicit).lower(),
        ).inc()


def record_entity_relationship(agent_id: str, relation_type: str):
    """Record an entity relationship."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_ENTITY_RELATIONSHIPS.labels(
            agent_id=agent_id,
            relation_type=relation_type,
        ).inc()


def record_learning(agent_id: str, category: str):
    """Record a learning."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_LEARNINGS.labels(agent_id=agent_id, category=category).inc()


def record_error_pattern(agent_id: str, severity: str):
    """Record an error pattern."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_ERRORS_RECORDED.labels(agent_id=agent_id, severity=severity).inc()


def record_task_created(agent_id: str, schema_type: str):
    """Record task creation."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_TASKS_CREATED.labels(
            agent_id=agent_id,
            schema_type=schema_type,
        ).inc()


def record_task_completed(agent_id: str, success: bool, duration_seconds: float, schema_type: str = "default"):
    """Record task completion."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_TASKS_COMPLETED.labels(
            agent_id=agent_id,
            success=str(success).lower(),
        ).inc()
        MEMORY_TASK_DURATION.labels(
            agent_id=agent_id,
            schema_type=schema_type,
        ).observe(duration_seconds)


def record_cache_hit(agent_id: str, memory_type: str):
    """Record a cache hit."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_CACHE_HITS.labels(agent_id=agent_id, memory_type=memory_type).inc()


def record_cache_miss(agent_id: str, memory_type: str):
    """Record a cache miss."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_CACHE_MISSES.labels(agent_id=agent_id, memory_type=memory_type).inc()


def record_context_build(agent_id: str, duration_seconds: float, size_chars: int):
    """Record context building metrics."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_CONTEXT_BUILD_DURATION.labels(agent_id=agent_id).observe(duration_seconds)
        MEMORY_CONTEXT_SIZE.labels(agent_id=agent_id).observe(size_chars)


def set_conversations_active(agent_id: str, count: int):
    """Set active conversations gauge."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_CONVERSATIONS_ACTIVE.labels(agent_id=agent_id).set(count)


def set_entities_count(agent_id: str, entity_type: str, count: int):
    """Set entities count gauge."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_ENTITIES.labels(agent_id=agent_id, entity_type=entity_type).set(count)


def set_patterns_count(agent_id: str, count: int):
    """Set patterns count gauge."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_PATTERNS.labels(agent_id=agent_id).set(count)


def set_store_entries(agent_id: str, store_type: str, memory_type: str, count: int):
    """Set store entries gauge."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_STORE_SIZE.labels(
            agent_id=agent_id,
            store_type=store_type,
            memory_type=memory_type,
        ).set(count)


def init_memory_build_info(version: str):
    """Initialize memory build info."""
    if PROMETHEUS_AVAILABLE:
        MEMORY_BUILD_INFO.info({
            "version": version,
            "prometheus_enabled": "true",
            "otel_enabled": str(OTEL_AVAILABLE).lower(),
        })


# Initialize build info on import
init_memory_build_info("1.0.0")
