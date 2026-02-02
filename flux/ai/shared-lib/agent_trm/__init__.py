"""
TRM (Tiny Recursive Model) client library for homelab agents.

Provides direct TRM inference capabilities with built-in reflection.
All agents should use this instead of Ollama for reasoning tasks.
"""
from .client import TRMClient
from .types import TRMRequest, TRMResponse, ReflectionStep

__all__ = [
    "TRMClient",
    "TRMRequest",
    "TRMResponse",
    "ReflectionStep",
]
