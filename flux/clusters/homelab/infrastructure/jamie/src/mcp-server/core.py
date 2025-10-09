#!/usr/bin/env python3
"""
🤖 Jamie MCP Server Core Module
Shared functionality for Jamie MCP Server with Logfire integration
MCP Server is just a protocol wrapper - NO AI
"""

import os
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
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
SERVICE_NAME = os.environ.get("SERVICE_NAME", "jamie-mcp-server")

# Configure Logfire
jamie_mcp_token = os.getenv('LOGFIRE_TOKEN_JAMIE_MCP')
if jamie_mcp_token:
    try:
        logfire.configure(service_name=SERVICE_NAME, token=jamie_mcp_token)
        logger.info("✅ Logfire configured successfully")
    except Exception as e:
        logger.warning(f"⚠️  Logfire configuration failed: {e}")
        logger.warning("⚠️  Continuing without Logfire...")
        os.environ.pop('LOGFIRE_TOKEN_JAMIE_MCP', None)
else:
    logger.warning("⚠️  LOGFIRE_TOKEN_JAMIE_MCP not set, skipping Logfire configuration")

__all__ = ['logger', 'logfire', 'SERVICE_NAME']
