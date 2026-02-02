"""
🚀 OpenTelemetry Initialization - Automatic Setup for Agents

Initializes OpenTelemetry tracing and metrics based on ObservabilitySettings.
Handles exporter setup, resource configuration, and sampler configuration.
"""

import os
import logging
from typing import Optional

import structlog

# OpenTelemetry imports (optional - fail gracefully if not installed)
try:
    from opentelemetry import trace, metrics
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor
    from opentelemetry.sdk.metrics import MeterProvider
    from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
    from opentelemetry.sdk.resources import Resource
    from opentelemetry.semconv.resource import ResourceAttributes
    from opentelemetry.trace import TraceIdRatioBased, AlwaysOn, AlwaysOff
    from opentelemetry.sdk.trace.sampling import ParentBased
    
    # OTLP exporters
    try:
        from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
        from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter
        OTLP_GRPC_AVAILABLE = True
    except ImportError:
        OTLP_GRPC_AVAILABLE = False
    
    try:
        from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter as OTLPSpanExporterHTTP
        from opentelemetry.exporter.otlp.proto.http.metric_exporter import OTLPMetricExporter as OTLPMetricExporterHTTP
        OTLP_HTTP_AVAILABLE = True
    except ImportError:
        OTLP_HTTP_AVAILABLE = False
    
    OTEL_AVAILABLE = True
except ImportError:
    OTEL_AVAILABLE = False
    OTLP_GRPC_AVAILABLE = False
    OTLP_HTTP_AVAILABLE = False
    # Create mock modules to avoid NameError
    class MockTrace:
        Tracer = None
    class MockMetrics:
        Meter = None
    trace = MockTrace()
    metrics = MockMetrics()
    TracerProvider = None
    MeterProvider = None

from .config import ObservabilitySettings

logger = structlog.get_logger()

# Global state
# Use string annotations to avoid NameError if OpenTelemetry is not available
_tracer_provider: Optional["TracerProvider"] = None
_meter_provider: Optional["MeterProvider"] = None
_initialized = False
_settings: Optional[ObservabilitySettings] = None


def _create_sampler(settings: ObservabilitySettings):
    """Create trace sampler based on configuration."""
    sampler_type = settings.otel_traces_sampler.lower()
    sampler_arg = settings.otel_traces_sampler_arg
    
    if sampler_type == "always_on":
        return AlwaysOn()
    elif sampler_type == "always_off":
        return AlwaysOff()
    elif sampler_type == "traceidratio":
        return TraceIdRatioBased(sampler_arg)
    elif sampler_type == "parentbased_traceidratio":
        return ParentBased(TraceIdRatioBased(sampler_arg))
    else:
        logger.warning(
            "unknown_sampler_type",
            sampler_type=sampler_type,
            defaulting_to="always_on",
        )
        return AlwaysOn()


def initialize_tracing(settings: ObservabilitySettings) -> bool:
    """
    Initialize OpenTelemetry tracing.
    
    Args:
        settings: ObservabilitySettings instance
        
    Returns:
        True if tracing was initialized, False otherwise
    """
    global _tracer_provider
    
    if not OTEL_AVAILABLE:
        logger.warning("opentelemetry_not_available", message="OpenTelemetry SDK not installed")
        return False
    
    if not settings.is_tracing_configured():
        logger.info("tracing_disabled", message="Tracing is disabled or endpoint not configured")
        return False
    
    try:
        # Create resource with service information
        resource_attrs = settings.get_resource_attributes()
        resource = Resource.create(resource_attrs)
        
        # Create OTLP exporter
        endpoint = settings.otel_exporter_otlp_endpoint
        protocol = settings.otel_exporter_otlp_protocol.lower()
        
        if protocol == "grpc":
            if not OTLP_GRPC_AVAILABLE:
                logger.error(
                    "otlp_grpc_not_available",
                    message="OTLP gRPC exporter not installed. Install: opentelemetry-exporter-otlp-proto-grpc",
                )
                return False
            
            exporter = OTLPSpanExporter(
                endpoint=endpoint,
                insecure=settings.otel_exporter_otlp_insecure,
            )
        elif protocol in ("http", "http/protobuf"):
            if not OTLP_HTTP_AVAILABLE:
                logger.error(
                    "otlp_http_not_available",
                    message="OTLP HTTP exporter not installed. Install: opentelemetry-exporter-otlp-proto-http",
                )
                return False
            
            # HTTP endpoint should include protocol
            if not endpoint.startswith("http://") and not endpoint.startswith("https://"):
                endpoint = f"http://{endpoint}"
            
            exporter = OTLPSpanExporterHTTP(
                endpoint=endpoint,
            )
        else:
            logger.error(
                "unknown_otlp_protocol",
                protocol=protocol,
                supported=["grpc", "http", "http/protobuf"],
            )
            return False
        
        # Create sampler
        sampler = _create_sampler(settings)
        
        # Create tracer provider
        _tracer_provider = TracerProvider(
            resource=resource,
            sampler=sampler,
        )
        
        # Add batch span processor
        processor = BatchSpanProcessor(exporter)
        _tracer_provider.add_span_processor(processor)
        
        # Set global tracer provider
        trace.set_tracer_provider(_tracer_provider)
        
        logger.info(
            "tracing_initialized",
            endpoint=endpoint,
            protocol=protocol,
            service_name=settings.otel_service_name,
            sampler=settings.otel_traces_sampler,
        )
        
        return True
        
    except Exception as e:
        logger.error(
            "tracing_init_failed",
            error=str(e),
            error_type=type(e).__name__,
            exc_info=True,
        )
        return False


def initialize_metrics(settings: ObservabilitySettings) -> bool:
    """
    Initialize OpenTelemetry metrics.
    
    Args:
        settings: ObservabilitySettings instance
        
    Returns:
        True if metrics were initialized, False otherwise
    """
    global _meter_provider
    
    if not OTEL_AVAILABLE:
        logger.warning("opentelemetry_not_available", message="OpenTelemetry SDK not installed")
        return False
    
    if not settings.is_metrics_configured():
        logger.info("metrics_disabled", message="Metrics are disabled or endpoint not configured")
        return False
    
    try:
        # Create resource with service information
        resource_attrs = settings.get_resource_attributes()
        resource = Resource.create(resource_attrs)
        
        # Create OTLP exporter
        endpoint = settings.otel_exporter_otlp_endpoint
        protocol = settings.otel_exporter_otlp_protocol.lower()
        
        if protocol == "grpc":
            if not OTLP_GRPC_AVAILABLE:
                logger.error(
                    "otlp_grpc_not_available",
                    message="OTLP gRPC exporter not installed. Install: opentelemetry-exporter-otlp-proto-grpc",
                )
                return False
            
            exporter = OTLPMetricExporter(
                endpoint=endpoint,
                insecure=settings.otel_exporter_otlp_insecure,
            )
        elif protocol in ("http", "http/protobuf"):
            if not OTLP_HTTP_AVAILABLE:
                logger.error(
                    "otlp_http_not_available",
                    message="OTLP HTTP exporter not installed. Install: opentelemetry-exporter-otlp-proto-http",
                )
                return False
            
            # HTTP endpoint should include protocol
            if not endpoint.startswith("http://") and not endpoint.startswith("https://"):
                endpoint = f"http://{endpoint}"
            
            exporter = OTLPMetricExporterHTTP(
                endpoint=endpoint,
            )
        else:
            logger.error(
                "unknown_otlp_protocol",
                protocol=protocol,
                supported=["grpc", "http", "http/protobuf"],
            )
            return False
        
        # Create metric reader
        reader = PeriodicExportingMetricReader(
            exporter,
            export_interval_millis=60000,  # Export every 60 seconds
        )
        
        # Create meter provider
        _meter_provider = MeterProvider(
            resource=resource,
            metric_readers=[reader],
        )
        
        # Set global meter provider
        metrics.set_meter_provider(_meter_provider)
        
        logger.info(
            "metrics_initialized",
            endpoint=endpoint,
            protocol=protocol,
            service_name=settings.otel_service_name,
        )
        
        return True
        
    except Exception as e:
        logger.error(
            "metrics_init_failed",
            error=str(e),
            error_type=type(e).__name__,
            exc_info=True,
        )
        return False


def initialize_observability(
    settings: Optional[ObservabilitySettings] = None,
    service_name: Optional[str] = None,
    service_namespace: Optional[str] = None,
    service_version: Optional[str] = None,
) -> bool:
    """
    Initialize OpenTelemetry observability (tracing and metrics).
    
    This is the main entry point for agents to initialize observability.
    It should be called at application startup.
    
    Args:
        settings: Optional ObservabilitySettings instance. If None, will be created from environment.
        service_name: Optional service name (overrides settings)
        service_namespace: Optional service namespace (overrides settings)
        service_version: Optional service version (overrides settings)
        
    Returns:
        True if observability was initialized successfully, False otherwise
        
    Example:
        # Simple usage with environment variables
        initialize_observability()
        
        # With explicit settings
        settings = ObservabilitySettings(
            otel_service_name="agent-bruno",
            otel_service_namespace="agent-bruno",
        )
        initialize_observability(settings)
        
        # With overrides
        initialize_observability(
            service_name="agent-bruno",
            service_namespace="agent-bruno",
        )
    """
    global _initialized, _settings
    
    if _initialized:
        logger.warning("observability_already_initialized")
        return True
    
    try:
        # Create or use provided settings
        if settings is None:
            # Try to create from environment, but allow service_name override
            if service_name:
                # Create minimal settings with required service_name
                settings = ObservabilitySettings(
                    otel_service_name=service_name,
                    otel_service_namespace=service_namespace or os.getenv("OTEL_SERVICE_NAMESPACE", "default"),
                    otel_service_version=service_version or os.getenv("OTEL_SERVICE_VERSION", "0.1.0"),
                )
            else:
                # Read entirely from environment
                settings = ObservabilitySettings()
        else:
            # Override with explicit parameters if provided
            if service_name:
                settings.otel_service_name = service_name
            if service_namespace:
                settings.otel_service_namespace = service_namespace
            if service_version:
                settings.otel_service_version = service_version
        
        _settings = settings
        
        # Initialize tracing
        tracing_initialized = initialize_tracing(settings)
        
        # Initialize metrics
        metrics_initialized = initialize_metrics(settings)
        
        if tracing_initialized or metrics_initialized:
            _initialized = True
            logger.info(
                "observability_initialized",
                tracing=tracing_initialized,
                metrics=metrics_initialized,
                service_name=settings.otel_service_name,
                endpoint=settings.otel_exporter_otlp_endpoint,
            )
            return True
        else:
            logger.warning(
                "observability_not_initialized",
                message="Neither tracing nor metrics were initialized. Check configuration.",
            )
            return False
            
    except Exception as e:
        logger.error(
            "observability_init_failed",
            error=str(e),
            error_type=type(e).__name__,
            exc_info=True,
        )
        return False


def get_tracer(name: Optional[str] = None) -> Optional["trace.Tracer"]:
    """
    Get OpenTelemetry tracer.
    
    Args:
        name: Optional tracer name (defaults to service name from settings)
        
    Returns:
        Tracer instance or None if not initialized
    """
    if not OTEL_AVAILABLE or not _initialized:
        return None
    
    if _tracer_provider is None:
        return None
    
    tracer_name = name or (_settings.otel_service_name if _settings else "unknown")
    return _tracer_provider.get_tracer(tracer_name)


def get_meter(name: Optional[str] = None) -> Optional["metrics.Meter"]:
    """
    Get OpenTelemetry meter.
    
    Args:
        name: Optional meter name (defaults to service name from settings)
        
    Returns:
        Meter instance or None if not initialized
    """
    if not OTEL_AVAILABLE or not _initialized:
        return None
    
    if _meter_provider is None:
        return None
    
    meter_name = name or (_settings.otel_service_name if _settings else "unknown")
    return _meter_provider.get_meter(meter_name)


def is_observability_enabled() -> bool:
    """Check if observability has been initialized."""
    return _initialized


def get_settings() -> Optional[ObservabilitySettings]:
    """Get current observability settings."""
    return _settings


def get_current_trace_context() -> dict[str, str]:
    """
    Get current trace context for logging.
    
    Returns:
        Dictionary with trace_id, span_id, and trace_flags if available
    """
    if not OTEL_AVAILABLE or not _initialized:
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


