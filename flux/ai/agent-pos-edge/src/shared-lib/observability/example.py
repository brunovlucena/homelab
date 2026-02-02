"""
Example usage of the observability module.

This file demonstrates how to use the shared observability module
in different scenarios.
"""

import os
from contextlib import asynccontextmanager
from fastapi import FastAPI
from observability import (
    initialize_observability,
    ObservabilitySettings,
    get_tracer,
    get_meter,
    is_observability_enabled,
)


# Example 1: Simple initialization with environment variables
def example_simple_init():
    """Initialize with environment variables."""
    # Set environment variables
    os.environ["OTEL_SERVICE_NAME"] = "agent-bruno"
    os.environ["OTEL_SERVICE_NAMESPACE"] = "agent-bruno"
    os.environ["OTEL_EXPORTER_OTLP_ENDPOINT"] = "alloy.observability.svc:4317"
    
    # Initialize
    initialize_observability()


# Example 2: Explicit settings
def example_explicit_settings():
    """Initialize with explicit settings."""
    settings = ObservabilitySettings(
        otel_service_name="agent-bruno",
        otel_service_namespace="agent-bruno",
        otel_service_version="1.0.0",
        otel_exporter_otlp_endpoint="alloy.observability.svc:4317",
        otel_tracing_enabled=True,
        otel_metrics_enabled=True,
    )
    
    initialize_observability(settings)


# Example 3: FastAPI integration
def example_fastapi():
    """FastAPI application with observability."""
    
    @asynccontextmanager
    async def lifespan(app: FastAPI):
        # Initialize observability at startup
        initialize_observability(
            service_name="agent-bruno",
            service_namespace="agent-bruno",
        )
        
        yield
        
        # Cleanup is automatic
    
    app = FastAPI(lifespan=lifespan)
    
    @app.get("/health")
    async def health():
        tracer = get_tracer()
        if tracer:
            with tracer.start_as_current_span("health_check") as span:
                span.set_attribute("endpoint", "/health")
                return {"status": "ok", "observability": is_observability_enabled()}
        return {"status": "ok", "observability": False}


# Example 4: Manual tracing
def example_manual_tracing():
    """Manual tracing in a function."""
    tracer = get_tracer()
    
    if not tracer:
        print("Tracer not available")
        return
    
    with tracer.start_as_current_span("process_data") as span:
        span.set_attribute("operation", "data_processing")
        span.set_attribute("input.size", 1000)
        
        # Do work
        result = process_data()
        
        span.set_attribute("output.size", len(result))
        span.set_attribute("success", True)
        
        return result


def process_data():
    """Example function that processes data."""
    return list(range(100))


# Example 5: Metrics
def example_metrics():
    """Using OpenTelemetry metrics."""
    meter = get_meter()
    
    if not meter:
        print("Meter not available")
        return
    
    # Create a counter
    counter = meter.create_counter(
        "requests_total",
        description="Total number of requests",
    )
    
    # Increment counter
    counter.add(1, {"endpoint": "/api/data"})


# Example 6: Sampling configuration
def example_sampling():
    """Configure trace sampling."""
    settings = ObservabilitySettings(
        otel_service_name="agent-bruno",
        otel_traces_sampler="traceidratio",
        otel_traces_sampler_arg=0.1,  # 10% sampling
    )
    
    initialize_observability(settings)


# Example 7: Resource attributes
def example_resource_attributes():
    """Add custom resource attributes."""
    settings = ObservabilitySettings(
        otel_service_name="agent-bruno",
        otel_resource_attributes="team=ai,component=chatbot,version=1.0.0",
    )
    
    initialize_observability(settings)


# Example 8: Check if initialized
def example_check_initialization():
    """Check if observability is initialized."""
    if is_observability_enabled():
        tracer = get_tracer()
        print(f"Observability enabled, tracer: {tracer}")
    else:
        print("Observability not initialized")


if __name__ == "__main__":
    # Run examples
    print("Example 1: Simple initialization")
    example_simple_init()
    
    print("\nExample 2: Explicit settings")
    example_explicit_settings()
    
    print("\nExample 8: Check initialization")
    example_check_initialization()


