"""
Configuration for SRE Agent.
"""
import os
from typing import Optional
from pydantic import BaseModel, Field


class AgentConfig(BaseModel):
    """SRE Agent configuration."""
    
    # Prometheus settings
    prometheus_url: str = Field(
        default=os.getenv("PROMETHEUS_URL", "http://prometheus:9090"),
        description="Prometheus server URL"
    )
    prometheus_timeout: int = Field(
        default=int(os.getenv("PROMETHEUS_TIMEOUT", "30")),
        description="Prometheus query timeout in seconds"
    )
    
    # Model settings
    model_name: str = Field(
        default=os.getenv("MODEL_NAME", "functiongemma-270m-it"),
        description="Model name for inference"
    )
    model_backend: str = Field(
        default=os.getenv("MODEL_BACKEND", "mlx"),
        description="Model backend: mlx, ollama, or anthropic"
    )
    mlx_enabled: bool = Field(
        default=os.getenv("MLX_ENABLED", "true").lower() == "true",
        description="Enable MLX-LM framework (requires Apple Silicon)"
    )
    
    # Ollama settings (fallback)
    ollama_url: str = Field(
        default=os.getenv("OLLAMA_URL", "http://ollama-native.ollama.svc.cluster.local:11434"),
        description="Ollama server URL"
    )
    
    # Anthropic settings (fallback)
    anthropic_api_key: Optional[str] = Field(
        default=os.getenv("ANTHROPIC_API_KEY"),
        description="Anthropic API key"
    )
    
    # Report settings
    report_time_range: str = Field(
        default=os.getenv("REPORT_TIME_RANGE", "1h"),
        description="Time range for metrics (Prometheus format)"
    )
    
    @classmethod
    def from_env(cls) -> "AgentConfig":
        """Create config from environment variables."""
        return cls()

