"""
Shared library for integrating Agent-Reasoning (TRM) into homelab agents.

This module provides a client for calling the reasoning service from other agents.
"""
from .client import ReasoningClient
from .types import ReasoningRequest, ReasoningResponse, TaskType

__all__ = [
    "ReasoningClient",
    "ReasoningRequest",
    "ReasoningResponse",
    "TaskType",
]

