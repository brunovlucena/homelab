"""
Pytest configuration and fixtures for agent-bruno tests.

This conftest imports shared fixtures from the shared-tests library
and provides agent-specific fixtures.
"""

import os
import sys
import pytest
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock

# =============================================================================
# PATH SETUP
# =============================================================================

# Add src to path for imports
src_path = Path(__file__).parent.parent / "src"
sys.path.insert(0, str(src_path))

# Add shared-tests to path
shared_tests_path = Path(__file__).parent.parent.parent / "shared-tests"
sys.path.insert(0, str(shared_tests_path))

# =============================================================================
# ENVIRONMENT VARIABLES
# =============================================================================

os.environ.setdefault("TESTING", "true")
os.environ.setdefault("LOG_LEVEL", "DEBUG")
os.environ.setdefault("OLLAMA_URL", "http://localhost:11434")

# =============================================================================
# IMPORT SHARED FIXTURES
# =============================================================================

try:
    from shared_tests.fixtures.cloudevents import (
        cloudevent_factory,
        sample_cloudevent,
        sample_chat_event,
    )
    from shared_tests.fixtures.http import (
        mock_httpx_client as shared_mock_httpx_client,
        mock_httpx_response,
    )
    from shared_tests.fixtures.ollama import (
        mock_ollama_response as shared_mock_ollama_response,
        mock_ollama_client,
        mock_ollama_chat_response,
    )
    from shared_tests.fixtures.metrics import (
        metrics_registry,
        metrics_helper,
    )
    from shared_tests.fixtures.redis import (
        mock_redis_client,
    )
    
    SHARED_TESTS_AVAILABLE = True
except ImportError:
    SHARED_TESTS_AVAILABLE = False

# =============================================================================
# AGENT-SPECIFIC FIXTURES
# =============================================================================

@pytest.fixture
def mock_ollama_response():
    """Mock Ollama API response for agent-bruno."""
    return {
        "response": "Hello! I'm Agent-Bruno, your homelab assistant. How can I help you today?",
        "eval_count": 25,  # Output tokens
        "prompt_eval_count": 50,  # Input tokens
        "model": "llama3.2:3b",
    }


@pytest.fixture
def mock_httpx_client(mock_ollama_response):
    """Mock httpx AsyncClient."""
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.json.return_value = mock_ollama_response
    mock_response.raise_for_status = MagicMock()
    
    mock_client = AsyncMock()
    mock_client.post.return_value = mock_response
    mock_client.get.return_value = mock_response
    mock_client.__aenter__.return_value = mock_client
    mock_client.__aexit__.return_value = None
    
    return mock_client


@pytest.fixture
def chat_bot_config():
    """Configuration for ChatBot testing."""
    return {
        "ollama_url": "http://localhost:11434",
        "model": "llama3.2:3b",
        "system_prompt": "You are Agent-Bruno, a helpful homelab assistant.",
    }


@pytest.fixture
def sample_conversation():
    """Sample conversation for testing."""
    from shared.types import Conversation, Message, MessageRole
    
    conv = Conversation(id="test-conv-123")
    conv.add_message(MessageRole.USER, "Hello!")
    conv.add_message(MessageRole.ASSISTANT, "Hi there! How can I help?")
    conv.add_message(MessageRole.USER, "Tell me about the homelab.")
    
    return conv


@pytest.fixture
def sample_notification():
    """Sample notification from another agent."""
    return {
        "type": "io.homelab.exploit.success",
        "source": "/agent-redteam/exploit-runner",
        "data": {
            "exploit_id": "vuln-001",
            "severity": "critical",
            "description": "Command injection detected",
        }
    }
