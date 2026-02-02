"""
🔧 Observability Configuration - Type-Safe Settings with Pydantic

Provides type-safe observability configuration using pydantic-settings.
All settings can be configured via environment variables with OTEL_ prefix.
"""

from typing import Optional
from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class ObservabilitySettings(BaseSettings):
    """
    Type-safe observability configuration using pydantic-settings.
    
    All settings can be configured via environment variables with OTEL_ prefix.
    For example, OTEL_EXPORTER_OTLP_ENDPOINT maps to otel_exporter_otlp_endpoint.
    
    Attributes:
        otel_exporter_otlp_endpoint: OTLP collector endpoint (default: alloy.observability.svc:4317)
        otel_service_name: Service name for traces (required)
        otel_service_version: Service version (default: "0.1.0")
        otel_service_namespace: Kubernetes namespace (default: "default")
        otel_tracing_enabled: Enable distributed tracing (default: True)
        otel_metrics_enabled: Enable metrics export (default: True)
        otel_logging_level: Logging level (default: "info")
        otel_traces_sampler: Trace sampling strategy (default: "always_on")
        otel_traces_sampler_arg: Sampling rate for ratio-based sampling (default: 1.0)
        otel_resource_attributes: Additional resource attributes as comma-separated key=value pairs
    """
    
    # OTLP Configuration
    otel_exporter_otlp_endpoint: Optional[str] = Field(
        default="alloy.observability.svc:4317",
        description="OTLP collector endpoint (Grafana Alloy or Tempo)",
    )
    otel_exporter_otlp_protocol: str = Field(
        default="grpc",
        description="OTLP protocol: grpc or http/protobuf",
    )
    otel_exporter_otlp_insecure: bool = Field(
        default=True,
        description="Use insecure connection (for local development)",
    )
    
    # Service Identification
    otel_service_name: str = Field(
        ...,
        description="Service name for traces and metrics (required)",
    )
    otel_service_version: str = Field(
        default="0.1.0",
        description="Service version",
    )
    otel_service_namespace: str = Field(
        default="default",
        description="Kubernetes namespace",
    )
    otel_deployment_environment: str = Field(
        default="production",
        description="Deployment environment (production, staging, development)",
    )
    
    # Feature Flags
    otel_tracing_enabled: bool = Field(
        default=True,
        description="Enable distributed tracing",
    )
    otel_metrics_enabled: bool = Field(
        default=True,
        description="Enable metrics export",
    )
    otel_logging_enabled: bool = Field(
        default=True,
        description="Enable structured logging",
    )
    
    # Tracing Configuration
    otel_traces_sampler: str = Field(
        default="always_on",
        description="Trace sampling strategy: always_on, always_off, traceidratio, parentbased_traceidratio",
    )
    otel_traces_sampler_arg: float = Field(
        default=1.0,
        description="Sampling rate for ratio-based samplers (0.0 - 1.0)",
        ge=0.0,
        le=1.0,
    )
    
    # Logging Configuration
    otel_logging_level: str = Field(
        default="info",
        description="Logging level: debug, info, warning, error",
    )
    otel_logging_format: str = Field(
        default="json",
        description="Logging format: json, text",
    )
    
    # Resource Attributes (comma-separated key=value pairs)
    otel_resource_attributes: Optional[str] = Field(
        default=None,
        description="Additional resource attributes as comma-separated key=value pairs",
    )
    
    model_config = SettingsConfigDict(
        env_prefix="OTEL_",
        case_sensitive=False,
        extra="ignore",
        validate_assignment=True,
    )
    
    def get_resource_attributes(self) -> dict[str, str]:
        """
        Parse resource attributes from comma-separated string.
        
        Returns:
            Dictionary of resource attributes
        """
        attrs = {
            "service.name": self.otel_service_name,
            "service.version": self.otel_service_version,
            "service.namespace": self.otel_service_namespace,
            "deployment.environment": self.otel_deployment_environment,
        }
        
        if self.otel_resource_attributes:
            for attr in self.otel_resource_attributes.split(","):
                if "=" in attr:
                    key, value = attr.split("=", 1)
                    attrs[key.strip()] = value.strip()
        
        return attrs
    
    def is_tracing_configured(self) -> bool:
        """Check if tracing is enabled and endpoint is configured."""
        return (
            self.otel_tracing_enabled
            and self.otel_exporter_otlp_endpoint is not None
            and self.otel_exporter_otlp_endpoint != ""
        )
    
    def is_metrics_configured(self) -> bool:
        """Check if metrics are enabled and endpoint is configured."""
        return (
            self.otel_metrics_enabled
            and self.otel_exporter_otlp_endpoint is not None
            and self.otel_exporter_otlp_endpoint != ""
        )


