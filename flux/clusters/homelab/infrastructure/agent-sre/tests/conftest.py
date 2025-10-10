#!/usr/bin/env python3
"""
Pytest configuration and shared fixtures for agent-sre tests
"""

import pytest
import sys
import os
from unittest.mock import Mock, patch

# Add the deployments directories to the Python path
agent_path = os.path.join(os.path.dirname(__file__), '..', 'deployments', 'agent')
mcp_path = os.path.join(os.path.dirname(__file__), '..', 'deployments', 'mcp-server')
sys.path.insert(0, agent_path)
sys.path.insert(0, mcp_path)


@pytest.fixture(scope="session", autouse=True)
def setup_environment():
    """Set up environment variables for testing"""
    os.environ['OLLAMA_URL'] = 'http://test-ollama:11434'
    os.environ['MODEL_NAME'] = 'test-model'
    os.environ['SERVICE_NAME'] = 'test-sre-agent'
    os.environ['DEBUG'] = 'false'
    
    # Don't configure Logfire or LangSmith in tests
    os.environ.pop('LOGFIRE_TOKEN', None)
    os.environ.pop('LANGSMITH_API_KEY', None)
    
    yield
    
    # Cleanup
    for key in ['OLLAMA_URL', 'MODEL_NAME', 'SERVICE_NAME', 'DEBUG']:
        os.environ.pop(key, None)


@pytest.fixture
def mock_ollama_llm():
    """Mock Ollama LLM for testing"""
    with patch('core.llm') as mock_llm:
        mock_llm.ainvoke = Mock(return_value=Mock(content="Test response"))
        yield mock_llm


@pytest.fixture
def disable_logfire():
    """Disable Logfire instrumentation for tests"""
    with patch('logfire.configure'):
        with patch('logfire.instrument', lambda name: lambda f: f):
            yield

