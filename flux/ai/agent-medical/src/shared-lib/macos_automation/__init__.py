"""
macOS Automation Client Library
Provides Python client for calling macOS automation service from Kubernetes agents
"""

from .client import MacOSAutomationClient, AutomationError

__all__ = ["MacOSAutomationClient", "AutomationError"]

