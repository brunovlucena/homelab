#!/usr/bin/env python3
"""
Shared Core Module for MCP Server
Imports from agent core
"""

import os
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
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

__all__ = ['logger', 'logfire']
