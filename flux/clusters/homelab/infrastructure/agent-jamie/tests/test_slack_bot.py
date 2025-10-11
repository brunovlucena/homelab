#!/usr/bin/env python3
"""
Unit tests for Jamie Slack Bot
"""

import json
import os
import sys
from unittest.mock import AsyncMock, Mock, patch

import pytest

# Import the module to test
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src", "slack-bot"))

from jamie_slack_bot import AgentSREClient, JamieSlackBot


class TestAgentSREClient:
    """Test suite for AgentSREClient"""

    @pytest.fixture
    async def sre_client(self):
        """Create an AgentSREClient instance"""
        client = AgentSREClient(base_url="http://test-agent-sre:8080")
        async with client:
            yield client

    @pytest.mark.asyncio
    async def test_client_initialization(self):
        """Test client initialization"""
        client = AgentSREClient(base_url="http://test:8080")
        async with client:
            assert client.base_url == "http://test:8080"
            assert client.http_session is not None

    @pytest.mark.asyncio
    async def test_list_tools_success(self, sre_client):
        """Test successful tool discovery"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(
            return_value={
                "jsonrpc": "2.0",
                "id": 1,
                "result": {
                    "tools": [
                        {"name": "prometheus_query", "description": "Query Prometheus", "inputSchema": {}},
                        {"name": "grafana_query", "description": "Query Grafana", "inputSchema": {}},
                    ]
                },
            }
        )

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        sre_client.http_session.post = Mock(return_value=mock_context)

        tools = await sre_client.list_tools()

        assert len(tools) == 2
        assert tools[0]["name"] == "prometheus_query"
        assert tools[1]["name"] == "grafana_query"

    @pytest.mark.asyncio
    async def test_list_tools_cache(self, sre_client):
        """Test that tools are cached after first call"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(
            return_value={"jsonrpc": "2.0", "id": 1, "result": {"tools": [{"name": "test_tool"}]}}
        )

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        sre_client.http_session.post = Mock(return_value=mock_context)

        # First call
        tools1 = await sre_client.list_tools()
        # Second call should use cache
        tools2 = await sre_client.list_tools()

        assert tools1 == tools2
        # post should only be called once due to caching
        assert sre_client.http_session.post.call_count == 1

    @pytest.mark.asyncio
    async def test_call_tool_success(self, sre_client):
        """Test successful tool call"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(
            return_value={
                "jsonrpc": "2.0",
                "id": 1,
                "result": {"content": [{"type": "text", "text": "Query result"}]},
            }
        )

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        sre_client.http_session.post = Mock(return_value=mock_context)

        result = await sre_client.call_tool("prometheus_query", {"query": "up"})

        assert result == "Query result"

    @pytest.mark.asyncio
    async def test_call_tool_error(self, sre_client):
        """Test tool call with error response"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(
            return_value={"jsonrpc": "2.0", "id": 1, "error": {"code": -32601, "message": "Tool not found"}}
        )

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        sre_client.http_session.post = Mock(return_value=mock_context)

        result = await sre_client.call_tool("unknown_tool", {})

        assert "Error" in result
        assert "Tool not found" in result

    @pytest.mark.asyncio
    async def test_call_tool_http_error(self, sre_client):
        """Test tool call with HTTP error"""
        mock_response = AsyncMock()
        mock_response.status = 500
        mock_response.text = AsyncMock(return_value="Internal Server Error")

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        sre_client.http_session.post = Mock(return_value=mock_context)

        result = await sre_client.call_tool("test_tool", {})

        assert "Error" in result
        assert "500" in result

    @pytest.mark.asyncio
    async def test_query_prometheus(self, sre_client):
        """Test Prometheus query helper method"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(
            return_value={
                "jsonrpc": "2.0",
                "id": 1,
                "result": {"content": [{"type": "text", "text": "Prometheus result"}]},
            }
        )

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        sre_client.http_session.post = Mock(return_value=mock_context)

        result = await sre_client.query_prometheus('up{job="test"}')

        assert result == "Prometheus result"

    @pytest.mark.asyncio
    async def test_query_grafana(self, sre_client):
        """Test Grafana query helper method"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(
            return_value={
                "jsonrpc": "2.0",
                "id": 1,
                "result": {"content": [{"type": "text", "text": "Grafana result"}]},
            }
        )

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        sre_client.http_session.post = Mock(return_value=mock_context)

        result = await sre_client.query_grafana("test-dashboard", "test-query")

        assert result == "Grafana result"


class TestJamieSlackBot:
    """Test suite for JamieSlackBot"""

    @pytest.fixture
    def slack_bot(self):
        """Create a JamieSlackBot instance"""
        with patch("jamie_slack_bot.AsyncApp"):
            with patch("jamie_slack_bot.ChatOllama"):
                bot = JamieSlackBot()
                return bot

    def test_bot_initialization(self, slack_bot):
        """Test bot initialization"""
        assert slack_bot.app is not None
        assert slack_bot.llm is not None

    @pytest.mark.asyncio
    async def test_health_endpoint(self, slack_bot):
        """Test health check endpoint"""
        request = Mock()
        response = await slack_bot.health_handler(request)

        assert response.status == 200
        body = json.loads(response.body.decode())
        assert body["status"] == "healthy"
        assert body["service"] == "jamie-slack-bot"

    @pytest.mark.asyncio
    async def test_chat_api_endpoint(self, slack_bot):
        """Test chat API endpoint"""
        with patch.object(slack_bot, "sre_client") as mock_sre:
            mock_sre.list_tools = AsyncMock(return_value=[])

            with patch.object(slack_bot.llm, "ainvoke") as mock_invoke:
                mock_invoke.return_value = Mock(content="Test response")

                request = Mock()
                request.json = AsyncMock(return_value={"message": "Hello"})

                response = await slack_bot.chat_handler(request)

                assert response.status == 200
                body = json.loads(response.body.decode())
                assert "response" in body

    @pytest.mark.asyncio
    async def test_chat_api_missing_message(self, slack_bot):
        """Test chat API with missing message"""
        request = Mock()
        request.json = AsyncMock(return_value={})

        response = await slack_bot.chat_handler(request)

        assert response.status == 400
        body = json.loads(response.body.decode())
        assert "error" in body

    @pytest.mark.asyncio
    async def test_prometheus_query_endpoint(self, slack_bot):
        """Test Prometheus query endpoint"""
        with patch.object(slack_bot, "sre_client") as mock_sre:
            mock_sre.query_prometheus = AsyncMock(return_value="Query result")

            request = Mock()
            request.json = AsyncMock(return_value={"query": "up"})

            response = await slack_bot.prometheus_query_handler(request)

            assert response.status == 200
            body = json.loads(response.body.decode())
            assert "result" in body

    @pytest.mark.asyncio
    async def test_prometheus_query_missing_query(self, slack_bot):
        """Test Prometheus query with missing query"""
        request = Mock()
        request.json = AsyncMock(return_value={})

        response = await slack_bot.prometheus_query_handler(request)

        assert response.status == 400
        body = json.loads(response.body.decode())
        assert "error" in body

    @pytest.mark.asyncio
    async def test_golden_signals_endpoint(self, slack_bot):
        """Test golden signals endpoint"""
        with patch.object(slack_bot, "sre_client") as mock_sre:
            mock_sre.query_prometheus = AsyncMock(return_value="Metric data")

            request = Mock()
            request.json = AsyncMock(return_value={"service": "test-service", "namespace": "default"})

            response = await slack_bot.golden_signals_handler(request)

            assert response.status == 200
            body = json.loads(response.body.decode())
            assert "latency" in body or "traffic" in body or "errors" in body or "saturation" in body

    @pytest.mark.asyncio
    async def test_golden_signals_missing_service(self, slack_bot):
        """Test golden signals with missing service"""
        request = Mock()
        request.json = AsyncMock(return_value={})

        response = await slack_bot.golden_signals_handler(request)

        assert response.status == 400
        body = json.loads(response.body.decode())
        assert "error" in body

    @pytest.mark.asyncio
    async def test_pod_logs_endpoint(self, slack_bot):
        """Test pod logs endpoint"""
        request = Mock()
        request.json = AsyncMock(return_value={"pod_name": "test-pod", "namespace": "default", "lines": 100})

        # Mock subprocess call for kubectl
        with patch("asyncio.create_subprocess_exec") as mock_subprocess:
            mock_process = Mock()
            mock_process.communicate = AsyncMock(return_value=(b"log line 1\nlog line 2", b""))
            mock_process.returncode = 0
            mock_subprocess.return_value = mock_process

            response = await slack_bot.pod_logs_handler(request)

            assert response.status == 200
            body = json.loads(response.body.decode())
            assert "logs" in body

    @pytest.mark.asyncio
    async def test_pod_logs_missing_pod_name(self, slack_bot):
        """Test pod logs with missing pod name"""
        request = Mock()
        request.json = AsyncMock(return_value={})

        response = await slack_bot.pod_logs_handler(request)

        assert response.status == 400
        body = json.loads(response.body.decode())
        assert "error" in body

    @pytest.mark.asyncio
    async def test_analyze_logs_endpoint(self, slack_bot):
        """Test analyze logs endpoint"""
        with patch.object(slack_bot.llm, "ainvoke") as mock_invoke:
            mock_invoke.return_value = Mock(content="Log analysis result")

            request = Mock()
            request.json = AsyncMock(return_value={"logs": "ERROR: test error", "context": "Test context"})

            response = await slack_bot.analyze_logs_handler(request)

            assert response.status == 200
            body = json.loads(response.body.decode())
            assert "analysis" in body

    @pytest.mark.asyncio
    async def test_analyze_logs_missing_logs(self, slack_bot):
        """Test analyze logs with missing logs"""
        request = Mock()
        request.json = AsyncMock(return_value={})

        response = await slack_bot.analyze_logs_handler(request)

        assert response.status == 400
        body = json.loads(response.body.decode())
        assert "error" in body

    def test_format_message_for_slack(self, slack_bot):
        """Test message formatting for Slack"""
        # Test basic text
        result = slack_bot._format_message_for_slack("Hello world")
        assert "Hello world" in result

        # Test with markdown
        result = slack_bot._format_message_for_slack("**Bold** and *italic*")
        assert "Bold" in result
        assert "italic" in result

        # Test with code blocks
        result = slack_bot._format_message_for_slack("```python\nprint('hello')\n```")
        assert "print" in result

    @pytest.mark.asyncio
    async def test_slack_message_event(self, slack_bot):
        """Test Slack message event handling"""
        with patch.object(slack_bot, "sre_client") as mock_sre:
            mock_sre.list_tools = AsyncMock(return_value=[])

            with patch.object(slack_bot.llm, "ainvoke") as mock_invoke:
                mock_invoke.return_value = Mock(content="Test response")

                mock_say = AsyncMock()
                mock_event = {"text": "Hello Jamie", "user": "U123", "channel": "C123"}

                await slack_bot.handle_message(mock_event, mock_say)

                mock_say.assert_called()

    @pytest.mark.asyncio
    async def test_slack_app_mention(self, slack_bot):
        """Test Slack app mention event handling"""
        with patch.object(slack_bot, "sre_client") as mock_sre:
            mock_sre.list_tools = AsyncMock(return_value=[])

            with patch.object(slack_bot.llm, "ainvoke") as mock_invoke:
                mock_invoke.return_value = Mock(content="Test response")

                mock_say = AsyncMock()
                mock_event = {"text": "<@U123> help", "user": "U456", "channel": "C123"}

                await slack_bot.handle_app_mention(mock_event, mock_say)

                mock_say.assert_called()

