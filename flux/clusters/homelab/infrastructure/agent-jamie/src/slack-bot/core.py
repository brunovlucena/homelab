#!/usr/bin/env python3
"""
🤖 Jamie Core Module
Shared functionality for Jamie with Logfire integration
"""

import os
import logging
from typing import Optional

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
OLLAMA_URL = os.environ.get("OLLAMA_URL", "http://192.168.0.16:11434")
MODEL_NAME = os.environ.get("MODEL_NAME", "llama3.2:3b")
SERVICE_NAME = os.environ.get("SERVICE_NAME", "jamie-slack-bot")
AGENT_SRE_URL = os.environ.get("AGENT_SRE_URL", "http://sre-agent-service.agent-sre:8080")

# Configure Logfire with dual export (Alloy + Logfire Cloud)
jamie_token = os.getenv('LOGFIRE_TOKEN_JAMIE')
alloy_endpoint = os.getenv('OTEL_EXPORTER_OTLP_ENDPOINT', 'http://alloy.alloy.svc.cluster.local:4317')
alloy_protocol = os.getenv('OTEL_EXPORTER_OTLP_PROTOCOL', 'grpc')

if jamie_token:
    try:
        # Configure to send to both Logfire cloud AND Alloy collector
        logfire.configure(
            service_name=SERVICE_NAME,
            token=jamie_token,
            send_to_logfire=True,  # ✅ Send to Logfire cloud
            console=False,
        )
        logger.info(f"✅ Logfire configured successfully (dual export: Logfire cloud + Alloy at {alloy_endpoint})")
    except Exception as e:
        logger.warning(f"⚠️  Logfire configuration failed: {e}")
        logger.warning("⚠️  Continuing without Logfire...")
        os.environ.pop('LOGFIRE_TOKEN_JAMIE', None)
else:
    logger.warning("⚠️  LOGFIRE_TOKEN_JAMIE not set, skipping Logfire configuration")

__all__ = ['logger', 'logfire', 'OLLAMA_URL', 'MODEL_NAME', 'SERVICE_NAME', 'AGENT_SRE_URL']

