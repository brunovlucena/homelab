"""
Observability utilities for agent-sre - OpenTelemetry Tracing, Metrics, and Structured Logging.

REFACTORED: Now uses OpenTelemetry for ALL metrics (not just Prometheus client).
Metrics are exported via OTLP to Alloy/Tempo, which converts them to Prometheus format.
This provides unified observability with proper trace context propagation.
"""
import os
import uuid
from typing import Optional, Dict, Any
from contextlib import contextmanager
from time import time
import structlog

# OpenTelemetry imports
try:
    from opentelemetry import trace, metrics
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor, ConsoleSpanExporter
    from opentelemetry.sdk.metrics import MeterProvider
    from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
    from opentelemetry.sdk.resources import Resource
    from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
    from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter
    from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
    from opentelemetry.instrumentation.httpx import HTTPXClientInstrumentor
    from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator
    OTEL_AVAILABLE = True
except ImportError:
    OTEL_AVAILABLE = False
    trace = None
    metrics = None

# Prometheus client for /metrics endpoint (backward compatibility)
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST

# Global state
_tracer_provider: Optional[TracerProvider] = None
_meter_provider: Optional[MeterProvider] = None
_tracer: Optional[trace.Tracer] = None
_meter: Optional[metrics.Meter] = None
_initialized = False

# OpenTelemetry Metrics (replaces Prometheus client metrics)
_remediation_attempts_counter: Optional[Any] = None
_remediation_duration_histogram: Optional[Any] = None
_active_remediations_gauge: Optional[Any] = None
_cloudevents_received_counter: Optional[Any] = None
_trm_inference_counter: Optional[Any] = None
_trm_inference_duration_histogram: Optional[Any] = None
_trm_confidence_histogram: Optional[Any] = None
_trm_fallback_counter: Optional[Any] = None
_trm_model_loaded_gauge: Optional[Any] = None


def _initialize_opentelemetry():
    """Initialize OpenTelemetry tracing and metrics."""
    global _tracer_provider, _meter_provider, _tracer, _meter
    global _remediation_attempts_counter, _remediation_duration_histogram
    global _active_remediations_gauge, _cloudevents_received_counter
    global _trm_inference_counter, _trm_inference_duration_histogram
    global _trm_confidence_histogram, _trm_fallback_counter, _trm_model_loaded_gauge
    
    if not OTEL_AVAILABLE:
        logger.warning("opentelemetry_not_available", message="OpenTelemetry SDK not installed")
        return False
    
    # Create resource
    resource = Resource.create({
        "service.name": "agent-sre",
        "service.namespace": os.getenv("NAMESPACE", "ai"),
        "service.version": os.getenv("VERSION", "0.5.0"),
        "deployment.environment": os.getenv("ENVIRONMENT", "production"),
    })
    
    # Initialize Tracing - use Tempo instead of Alloy
    otlp_endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "tempo.tempo:4317")
    if otlp_endpoint:
        try:
            otlp_trace_exporter = OTLPSpanExporter(endpoint=otlp_endpoint, insecure=True)
            _tracer_provider = TracerProvider(resource=resource)
            _tracer_provider.add_span_processor(BatchSpanProcessor(otlp_trace_exporter))
            trace.set_tracer_provider(_tracer_provider)
            _tracer = trace.get_tracer("agent-sre")
            logger.info("tracing_initialized", endpoint=otlp_endpoint)
        except Exception as e:
            logger.warning("tracing_init_failed", error=str(e), fallback_to_console=True)
            console_exporter = ConsoleSpanExporter()
            _tracer_provider = TracerProvider(resource=resource)
            _tracer_provider.add_span_processor(BatchSpanProcessor(console_exporter))
            trace.set_tracer_provider(_tracer_provider)
            _tracer = trace.get_tracer("agent-sre")
    else:
        console_exporter = ConsoleSpanExporter()
        _tracer_provider = TracerProvider(resource=resource)
        _tracer_provider.add_span_processor(BatchSpanProcessor(console_exporter))
        trace.set_tracer_provider(_tracer_provider)
        _tracer = trace.get_tracer("agent-sre")
    
    # Initialize Metrics
    if otlp_endpoint:
        try:
            otlp_metric_exporter = OTLPMetricExporter(endpoint=otlp_endpoint, insecure=True)
            metric_reader = PeriodicExportingMetricReader(
                otlp_metric_exporter,
                export_interval_millis=60000,  # Export every 60 seconds
            )
            _meter_provider = MeterProvider(resource=resource, metric_readers=[metric_reader])
            metrics.set_meter_provider(_meter_provider)
            _meter = metrics.get_meter("agent-sre")
            
            # Create OpenTelemetry metrics
            _remediation_attempts_counter = _meter.create_counter(
                "agent_sre_remediation_attempts_total",
                description="Total number of remediation attempts",
                unit="1"
            )
            
            _remediation_duration_histogram = _meter.create_histogram(
                "agent_sre_remediation_duration_seconds",
                description="Duration of remediation operations in seconds",
                unit="s"
            )
            
            _active_remediations_gauge = _meter.create_up_down_counter(
                "agent_sre_active_remediations",
                description="Number of active remediation operations",
                unit="1"
            )
            
            _cloudevents_received_counter = _meter.create_counter(
                "agent_sre_cloudevents_received_total",
                description="Total number of CloudEvents received",
                unit="1"
            )
            
            # TRM Model Metrics
            _trm_inference_counter = _meter.create_counter(
                "agent_sre_trm_inference_total",
                description="Total number of TRM inference calls",
                unit="1"
            )
            
            _trm_inference_duration_histogram = _meter.create_histogram(
                "agent_sre_trm_inference_duration_seconds",
                description="Time spent in TRM inference",
                unit="s"
            )
            
            _trm_confidence_histogram = _meter.create_histogram(
                "agent_sre_trm_confidence_score",
                description="TRM confidence scores",
                unit="1"
            )
            
            _trm_fallback_counter = _meter.create_counter(
                "agent_sre_trm_fallback_total",
                description="Number of times TRM falls back to rule-based",
                unit="1"
            )
            
            _trm_model_loaded_gauge = _meter.create_up_down_counter(
                "agent_sre_trm_model_loaded",
                description="Whether TRM model is loaded (1=loaded, 0=not loaded)",
                unit="1"
            )
            
            logger.info("metrics_initialized", endpoint=otlp_endpoint)
        except Exception as e:
            logger.warning("metrics_init_failed", error=str(e))
    
    return True


# Configure structured logging with trace context
structlog.configure(
    processors=[
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.JSONRenderer()
    ],
    context_class=dict,
    logger_factory=structlog.PrintLoggerFactory(),
    wrapper_class=structlog.make_filtering_bound_logger(20),  # INFO level
    cache_logger_on_first_use=True,
)

logger = structlog.get_logger()


def get_correlation_id(event_id: Optional[str] = None, headers: Optional[Dict[str, str]] = None) -> str:
    """
    Extract or generate correlation ID from event or headers.
    Also extracts trace context from headers for distributed tracing.
    
    Args:
        event_id: CloudEvent ID
        headers: HTTP headers (may contain X-Correlation-ID, ce-id, or traceparent)
    
    Returns:
        Correlation ID for tracing
    """
    if headers:
        # Check for correlation ID in headers
        correlation_id = (
            headers.get("X-Correlation-ID") or
            headers.get("ce-id") or
            headers.get("traceparent", "").split("-")[1] if headers.get("traceparent") else None
        )
        if correlation_id:
            # Extract trace context from headers for distributed tracing
            if OTEL_AVAILABLE and headers.get("traceparent"):
                try:
                    propagator = TraceContextTextMapPropagator()
                    context = propagator.extract(headers)
                    trace.set_tracer_provider(_tracer_provider)
                    trace.set_span_in_context(context)
                except Exception as e:
                    logger.warning("trace_context_extraction_failed", error=str(e))
            return correlation_id
    
    # Use event ID if provided
    if event_id:
        return event_id
    
    # Generate new correlation ID
    return str(uuid.uuid4())


def set_correlation_context(correlation_id: str, event_id: Optional[str] = None, alertname: Optional[str] = None):
    """
    Set correlation context for structured logging and OpenTelemetry.
    
    Args:
        correlation_id: Correlation ID for this request
        event_id: CloudEvent ID
        alertname: Alert name (if applicable)
    """
    structlog.contextvars.clear_contextvars()
    structlog.contextvars.bind_contextvars(
        correlation_id=correlation_id,
        event_id=event_id,
        alertname=alertname,
    )
    
    # Add trace context to logs if available
    if OTEL_AVAILABLE and _tracer:
        try:
            span = trace.get_current_span()
            if span and span.is_recording():
                ctx = span.get_span_context()
                structlog.contextvars.bind_contextvars(
                    trace_id=format(ctx.trace_id, "032x"),
                    span_id=format(ctx.span_id, "016x"),
                )
        except Exception:
            pass


@contextmanager
def trace_remediation(alertname: str, lambda_function: str, correlation_id: str):
    """
    Context manager for tracing remediation operations with OpenTelemetry.
    
    Args:
        alertname: Name of the alert
        lambda_function: Name of the LambdaFunction
        correlation_id: Correlation ID for tracing
    
    Yields:
        OpenTelemetry span for the remediation operation
    """
    if not OTEL_AVAILABLE or not _tracer:
        # Fallback: just yield None if OpenTelemetry not available
        yield None
        return
    
    from opentelemetry import trace as otel_trace
    
    # Increment active remediations gauge
    if _active_remediations_gauge:
        _active_remediations_gauge.add(1, {
            "alertname": alertname,
        })
    
    start_time = time()
    
    with _tracer.start_as_current_span(
        "remediation.execute",
        attributes={
            "alertname": alertname,
            "lambda_function": lambda_function,
            "correlation_id": correlation_id,
            "operation.type": "remediation",
        }
    ) as span:
        try:
            yield span
            span.set_status(otel_trace.Status(otel_trace.StatusCode.OK))
            status_for_metrics = "success"
            
            # Record success metric
            if _remediation_attempts_counter:
                _remediation_attempts_counter.add(1, {
                    "alertname": alertname,
                    "lambda_function": lambda_function,
                    "status": "success",
                })
        except Exception as e:
            span.record_exception(e)
            span.set_status(otel_trace.Status(otel_trace.StatusCode.ERROR, str(e)))
            status_for_metrics = "error"
            
            # Record error metric
            if _remediation_attempts_counter:
                _remediation_attempts_counter.add(1, {
                    "alertname": alertname,
                    "lambda_function": lambda_function,
                    "status": "error",
                })
            raise
        finally:
            # Decrement active remediations
            if _active_remediations_gauge:
                _active_remediations_gauge.add(-1, {
                    "alertname": alertname,
                })
            
            # Record duration
            duration = time() - start_time
            if _remediation_duration_histogram:
                _remediation_duration_histogram.record(duration, {
                    "lambda_function": lambda_function,
                    "status": status_for_metrics,
                })


def log_remediation_step(
    step: str,
    alertname: str,
    lambda_function: Optional[str] = None,
    correlation_id: Optional[str] = None,
    **kwargs
):
    """
    Log a remediation step with full context including OpenTelemetry trace context.
    
    Args:
        step: Step name (e.g., "remediation.started", "remediation.lambda_called")
        alertname: Alert name
        lambda_function: LambdaFunction name (if applicable)
        correlation_id: Correlation ID
        **kwargs: Additional context to log
    """
    # Get current trace context
    trace_context = {}
    if OTEL_AVAILABLE:
        try:
            span = trace.get_current_span()
            if span and span.is_recording():
                ctx = span.get_span_context()
                trace_context = {
                    "trace_id": format(ctx.trace_id, "032x"),
                    "span_id": format(ctx.span_id, "016x"),
                }
        except Exception:
            pass
    
    logger.info(
        step,
        alertname=alertname,
        lambda_function=lambda_function,
        correlation_id=correlation_id,
        **trace_context,
        **kwargs
    )


def record_cloudevent_received(event_type: str, event_source: str):
    """Record CloudEvent received metric."""
    if _cloudevents_received_counter:
        _cloudevents_received_counter.add(1, {
            "event_type": event_type,
            "event_source": event_source,
        })


def record_trm_inference(method: str, status: str, duration: Optional[float] = None):
    """Record TRM inference metric."""
    if _trm_inference_counter:
        _trm_inference_counter.add(1, {
            "method": method,
            "status": status,
        })
    
    if duration is not None and _trm_inference_duration_histogram:
        _trm_inference_duration_histogram.record(duration, {
            "method": method,
        })


def record_trm_confidence(alertname: str, confidence: float):
    """Record TRM confidence score."""
    if _trm_confidence_histogram:
        _trm_confidence_histogram.record(confidence, {
            "alertname": alertname,
        })


def record_trm_fallback(reason: str):
    """Record TRM fallback metric."""
    if _trm_fallback_counter:
        _trm_fallback_counter.add(1, {
            "reason": reason,
        })


def set_trm_model_loaded(loaded: bool):
    """Set TRM model loaded gauge."""
    if _trm_model_loaded_gauge:
        if loaded:
            _trm_model_loaded_gauge.add(1)
        else:
            _trm_model_loaded_gauge.add(-1)


def initialize_observability():
    """
    Initialize observability components with OpenTelemetry.
    
    NOTE: We do NOT start a separate Prometheus metrics server to follow
    Knative best practices. A separate metrics server on a different port
    prevents scale-to-zero because it keeps the pod alive with constant
    health checks. Metrics should be exposed via the main application port
    if needed, or collected via the queue-proxy metrics endpoint.
    
    OpenTelemetry metrics are exported via OTLP to Alloy, which converts
    them to Prometheus format automatically.
    """
    global _initialized
    
    if _initialized:
        logger.warning("observability_already_initialized")
        return
    
    # Initialize OpenTelemetry
    _initialize_opentelemetry()
    
    # Instrument FastAPI and HTTPX
    if OTEL_AVAILABLE:
        try:
            FastAPIInstrumentor().instrument()
            HTTPXClientInstrumentor().instrument()
        except Exception as e:
            logger.warning("instrumentation_failed", error=str(e))
    
    _initialized = True
    logger.info(
        "observability_initialized",
        opentelemetry_available=OTEL_AVAILABLE,
        tracing_enabled=_tracer is not None,
        metrics_enabled=_meter is not None,
    )


def get_prometheus_metrics() -> bytes:
    """
    Get Prometheus-formatted metrics for /metrics endpoint.
    
    NOTE: This is for backward compatibility. OpenTelemetry metrics
    are exported via OTLP to Alloy, which converts them to Prometheus.
    This endpoint can be used for direct Prometheus scraping if needed.
    """
    # For now, return empty metrics since we're using OpenTelemetry
    # In the future, we could use opentelemetry-exporter-prometheus
    # to expose Prometheus format directly
    return b"# Metrics exported via OpenTelemetry OTLP to Alloy\n"


def get_tracer() -> Optional[trace.Tracer]:
    """Get OpenTelemetry tracer."""
    return _tracer


def get_meter() -> Optional[metrics.Meter]:
    """Get OpenTelemetry meter."""
    return _meter


def get_current_trace_context() -> Dict[str, str]:
    """
    Get current trace context for logging.
    
    Returns:
        Dictionary with trace_id, span_id, and trace_flags if available
    """
    if not OTEL_AVAILABLE:
        return {}
    
    try:
        span = trace.get_current_span()
        if span and span.is_recording():
            ctx = span.get_span_context()
            return {
                "trace_id": format(ctx.trace_id, "032x"),
                "span_id": format(ctx.span_id, "016x"),
                "trace_flags": str(ctx.trace_flags),
            }
    except Exception:
        pass
    
    return {}
