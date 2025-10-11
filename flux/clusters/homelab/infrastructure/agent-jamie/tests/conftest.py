#!/usr/bin/env python3
"""
Pytest configuration and shared fixtures for Jamie tests
"""

import os
import sys
from unittest.mock import AsyncMock, Mock, patch

import pytest

# Add the src directories to the Python path
mcp_server_path = os.path.join(os.path.dirname(__file__), "..", "src", "mcp-server")
slack_bot_path = os.path.join(os.path.dirname(__file__), "..", "src", "slack-bot")
sys.path.insert(0, mcp_server_path)
sys.path.insert(0, slack_bot_path)


@pytest.fixture(scope="session", autouse=True)
def setup_environment():
    """Set up environment variables for testing"""
    os.environ["JAMIE_SLACK_BOT_URL"] = "http://test-jamie-slack-bot:8080"
    os.environ["SLACK_BOT_TOKEN"] = "xoxb-test-token"
    os.environ["SLACK_APP_TOKEN"] = "xapp-test-token"
    os.environ["AGENT_SRE_URL"] = "http://test-agent-sre:8080"
    os.environ["SERVICE_NAME"] = "test-jamie"
    os.environ["DEBUG"] = "false"

    # Don't configure Logfire or LangSmith in tests
    os.environ.pop("LOGFIRE_TOKEN", None)
    os.environ.pop("LANGSMITH_API_KEY", None)

    yield

    # Cleanup
    for key in [
        "JAMIE_SLACK_BOT_URL",
        "SLACK_BOT_TOKEN",
        "SLACK_APP_TOKEN",
        "AGENT_SRE_URL",
        "SERVICE_NAME",
        "DEBUG",
    ]:
        os.environ.pop(key, None)


@pytest.fixture
def mock_aiohttp_session():
    """Mock aiohttp ClientSession for API calls"""
    with patch("aiohttp.ClientSession") as mock_session:
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value={"status": "ok"})
        mock_response.text = AsyncMock(return_value="OK")

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        mock_context.__aexit__.return_value = None

        mock_session.return_value.get.return_value = mock_context
        mock_session.return_value.post.return_value = mock_context
        mock_session.return_value.__aenter__.return_value = mock_session.return_value
        mock_session.return_value.__aexit__.return_value = None

        yield mock_session


@pytest.fixture
def mock_slack_client():
    """Mock Slack WebClient"""
    with patch("slack_sdk.WebClient") as mock_client:
        mock_client.return_value.auth_test.return_value = {"ok": True, "user": "jamie", "user_id": "U123"}
        mock_client.return_value.chat_postMessage.return_value = {"ok": True, "ts": "1234567890.123456"}
        yield mock_client


@pytest.fixture
def mock_logfire():
    """Mock Logfire for testing"""
    with patch("logfire.configure"):
        with patch("logfire.instrument", lambda name: lambda f: f):
            yield


@pytest.fixture
def disable_logfire():
    """Disable Logfire instrumentation for tests"""
    with patch("logfire.configure"):
        with patch("logfire.instrument", lambda name: lambda f: f):
            yield

