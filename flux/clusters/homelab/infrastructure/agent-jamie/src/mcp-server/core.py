#!/usr/bin/env python3
"""
🔧 Core Module for Jamie MCP Server
Shared configuration, logging, and logfire setup
"""

import os
import logging
import sys

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout)
    ]
)

logger = logging.getLogger("jamie-mcp")

# 🔥 Logfire setup (optional, gracefully degrades if not configured)
try:
    import logfire as logfire_module
    
    # Configure logfire if token is available
    logfire_token = os.getenv("LOGFIRE_TOKEN")
    
    if logfire_token:
        logfire_module.configure(
            token=logfire_token,
            service_name="jamie-mcp-server"
        )
        logger.info("🔥 Logfire configured successfully")
        logfire = logfire_module
    else:
        logger.warning("⚠️  LOGFIRE_TOKEN not set - running without logfire observability")
        # Create a no-op logfire for graceful degradation
        class NoOpLogfire:
            def instrument(self, *args, **kwargs):
                def decorator(func):
                    return func
                return decorator
            
            def __getattr__(self, name):
                return lambda *args, **kwargs: None
        
        logfire = NoOpLogfire()

except ImportError:
    logger.warning("⚠️  logfire not installed - running without observability")
    # Create a no-op logfire for graceful degradation
    class NoOpLogfire:
        def instrument(self, *args, **kwargs):
            def decorator(func):
                return func
            return decorator
        
        def __getattr__(self, name):
            return lambda *args, **kwargs: None
    
    logfire = NoOpLogfire()

# Service configuration
SERVICE_NAME = "jamie-mcp-server"
JAMIE_SLACK_BOT_URL = os.getenv(
    "JAMIE_SLACK_BOT_URL", 
    "http://jamie-slack-bot-service.jamie.svc.cluster.local:8080"
)

logger.info(f"🤖 {SERVICE_NAME} core module loaded")
logger.info(f"🔗 Jamie Slack Bot URL: {JAMIE_SLACK_BOT_URL}")

