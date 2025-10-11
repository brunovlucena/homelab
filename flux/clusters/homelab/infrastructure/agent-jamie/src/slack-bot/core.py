#!/usr/bin/env python3
"""
🤖 Jamie Core Module
Shared functionality for Jamie with Logfire integration
"""

import logging
import os

# Configure logging
logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s")
logger = logging.getLogger(__name__)

# Try to import logfire, but continue if not available
try:
    import logfire

    logger.info("✅ Logfire imported successfully")
except ImportError:
    logger.warning("⚠️  Logfire not available, creating mock")

    # Create a mock logfire module
    class MockLogfire:
        @staticmethod
        def configure(*args, **kwargs):
            pass

        @staticmethod
        def instrument(name):
            def decorator(func):
                return func

            return decorator

    logfire = MockLogfire()

# Configuration
OLLAMA_URL = os.environ.get("OLLAMA_URL", "http://192.168.0.16:11434")
MODEL_NAME = os.environ.get("MODEL_NAME", "llama3.2:3b")
SERVICE_NAME = os.environ.get("SERVICE_NAME", "jamie")
AGENT_SRE_URL = os.environ.get("AGENT_SRE_URL", "http://sre-agent-service.agent-sre:8080")

# Configure OpenTelemetry to export to Alloy only (no Logfire Cloud)
logfire_token = os.getenv("LOGFIRE_TOKEN")
alloy_endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://alloy.alloy.svc.cluster.local:4317")
alloy_protocol = os.getenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
alloy_insecure = os.getenv("OTEL_EXPORTER_OTLP_INSECURE", "true").lower() == "true"

# Disable Logfire Cloud by removing the token - we only want local OTEL export to Alloy
if logfire_token:
    logger.info("⚠️  LOGFIRE_TOKEN detected but removing it to avoid HTTP exporter creation")
    logger.info("⚠️  Using direct OpenTelemetry export to Alloy instead")
    # Remove the token so Logfire doesn't auto-configure
    os.environ.pop("LOGFIRE_TOKEN", None)
    logfire_token = None

# Configure OpenTelemetry directly without Logfire SDK
try:
    from opentelemetry import trace
    from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor
    from opentelemetry.sdk.resources import Resource

    # Remove http:// prefix from endpoint for gRPC exporter
    # gRPC exporter expects host:port format, not http://host:port
    grpc_endpoint = alloy_endpoint.replace("http://", "").replace("https://", "")

    # Create resource with service name
    resource = Resource.create(
        {
            "service.name": SERVICE_NAME,
            "service.version": "1.0.0",
        }
    )

    # Set up tracer provider with Alloy OTLP exporter
    provider = TracerProvider(resource=resource)
    otlp_exporter = OTLPSpanExporter(endpoint=grpc_endpoint, insecure=alloy_insecure)
    provider.add_span_processor(BatchSpanProcessor(otlp_exporter))
    trace.set_tracer_provider(provider)

    logger.info(f"✅ OpenTelemetry configured successfully - exporting to Alloy at {grpc_endpoint}")
    logger.info("✅ Using @logfire.instrument decorators for tracing (no Logfire cloud, no HTTP exporter)")

    # NOTE: We do NOT call logfire.configure() to avoid it creating an HTTP exporter
    # The @logfire.instrument decorators will use our OpenTelemetry TracerProvider

except Exception as e:
    logger.warning(f"⚠️  OpenTelemetry configuration failed: {e}")
    logger.warning("⚠️  Continuing without OpenTelemetry tracing...")

__all__ = ["logger", "logfire", "OLLAMA_URL", "MODEL_NAME", "SERVICE_NAME", "AGENT_SRE_URL"]
