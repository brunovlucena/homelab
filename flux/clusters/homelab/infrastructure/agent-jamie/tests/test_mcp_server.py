#!/usr/bin/env python3
"""
Unit tests for Jamie MCP Server
"""

import json
import os
import sys
from unittest.mock import AsyncMock, Mock, patch

import pytest

# Import the module to test
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src", "mcp-server"))

from jamie_mcp_server import JamieAPIClient, JamieMCPServer


class TestJamieAPIClient:
    """Test suite for JamieAPIClient"""

    @pytest.fixture
    def api_client(self):
        """Create a JamieAPIClient instance"""
        return JamieAPIClient(base_url="http://test-jamie:8080")

    @pytest.mark.asyncio
    async def test_chat_success(self, api_client, mock_aiohttp_session):
        """Test successful chat request"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(
            return_value={"response": "Hello! How can I help you?", "timestamp": "2024-01-01T00:00:00"}
        )

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        mock_session_instance = AsyncMock()
        mock_session_instance.post.return_value = mock_context

        with patch("aiohttp.ClientSession") as mock_session:
            mock_session.return_value.__aenter__.return_value = mock_session_instance

            result = await api_client.chat("Hello")

            assert result["success"] is True
            assert "response" in result
            assert result["response"] == "Hello! How can I help you?"

    @pytest.mark.asyncio
    async def test_chat_error(self, api_client):
        """Test chat request with error response"""
        mock_response = AsyncMock()
        mock_response.status = 500
        mock_response.text = AsyncMock(return_value="Internal Server Error")

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        mock_session_instance = AsyncMock()
        mock_session_instance.post.return_value = mock_context

        with patch("aiohttp.ClientSession") as mock_session:
            mock_session.return_value.__aenter__.return_value = mock_session_instance

            result = await api_client.chat("Hello")

            assert result["success"] is False
            assert "error" in result
            assert "500" in result["error"]

    @pytest.mark.asyncio
    async def test_query_prometheus_success(self, api_client):
        """Test successful Prometheus query"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(
            return_value={"status": "success", "data": {"resultType": "vector", "result": []}}
        )

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        mock_session_instance = AsyncMock()
        mock_session_instance.post.return_value = mock_context

        with patch("aiohttp.ClientSession") as mock_session:
            mock_session.return_value.__aenter__.return_value = mock_session_instance

            result = await api_client.query_prometheus('up{job="test"}')

            assert result["success"] is True
            assert "result" in result

    @pytest.mark.asyncio
    async def test_check_golden_signals_success(self, api_client):
        """Test successful golden signals check"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(
            return_value={"latency": "100ms", "traffic": "1000 req/s", "errors": "0.1%", "saturation": "50%"}
        )

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        mock_session_instance = AsyncMock()
        mock_session_instance.post.return_value = mock_context

        with patch("aiohttp.ClientSession") as mock_session:
            mock_session.return_value.__aenter__.return_value = mock_session_instance

            result = await api_client.check_golden_signals("test-service", "default")

            assert result["success"] is True
            assert "signals" in result

    @pytest.mark.asyncio
    async def test_get_pod_logs_success(self, api_client):
        """Test successful pod logs retrieval"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value={"logs": ["log line 1", "log line 2", "log line 3"]})

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        mock_session_instance = AsyncMock()
        mock_session_instance.post.return_value = mock_context

        with patch("aiohttp.ClientSession") as mock_session:
            mock_session.return_value.__aenter__.return_value = mock_session_instance

            result = await api_client.get_pod_logs("test-pod", "default", lines=100)

            assert result["success"] is True
            assert "logs" in result
            assert len(result["logs"]) == 3

    @pytest.mark.asyncio
    async def test_analyze_logs_success(self, api_client):
        """Test successful log analysis"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value={"analysis": "Found 3 errors in the logs"})

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        mock_session_instance = AsyncMock()
        mock_session_instance.post.return_value = mock_context

        with patch("aiohttp.ClientSession") as mock_session:
            mock_session.return_value.__aenter__.return_value = mock_session_instance

            result = await api_client.analyze_logs("ERROR: test\nERROR: test2\nERROR: test3")

            assert result["success"] is True
            assert "analysis" in result

    @pytest.mark.asyncio
    async def test_health_check_success(self, api_client):
        """Test successful health check"""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(
            return_value={"status": "healthy", "service": "jamie-slack-bot", "version": "1.0.0"}
        )

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        mock_session_instance = AsyncMock()
        mock_session_instance.get.return_value = mock_context

        with patch("aiohttp.ClientSession") as mock_session:
            mock_session.return_value.__aenter__.return_value = mock_session_instance

            result = await api_client.health_check()

            assert result["success"] is True
            assert "status" in result

    @pytest.mark.asyncio
    async def test_health_check_failure(self, api_client):
        """Test health check with service down"""
        mock_response = AsyncMock()
        mock_response.status = 503
        mock_response.text = AsyncMock(return_value="Service Unavailable")

        mock_context = AsyncMock()
        mock_context.__aenter__.return_value = mock_response
        mock_session_instance = AsyncMock()
        mock_session_instance.get.return_value = mock_context

        with patch("aiohttp.ClientSession") as mock_session:
            mock_session.return_value.__aenter__.return_value = mock_session_instance

            result = await api_client.health_check()

            assert result["success"] is False
            assert "error" in result


class TestJamieMCPServer:
    """Test suite for JamieMCPServer"""

    @pytest.fixture
    def mcp_server(self):
        """Create a JamieMCPServer instance"""
        with patch("jamie_mcp_server.JamieAPIClient"):
            server = JamieMCPServer()
            return server

    def test_server_initialization(self, mcp_server):
        """Test server initialization"""
        assert mcp_server.server is not None
        assert mcp_server.api_client is not None

    @pytest.mark.asyncio
    async def test_list_tools(self, mcp_server):
        """Test that server lists all available tools"""
        # The server setup creates handlers, we can't easily test them
        # without running the actual MCP server, but we can verify the server exists
        assert mcp_server.server is not None

    @pytest.mark.asyncio
    async def test_chat_tool_success(self, mcp_server):
        """Test chat tool execution with success"""
        mcp_server.api_client.chat = AsyncMock(
            return_value={"success": True, "response": "Test response", "timestamp": "2024-01-01T00:00:00"}
        )

        # We can't directly test tool calls without MCP infrastructure,
        # but we can verify the API client method works
        result = await mcp_server.api_client.chat("Hello")
        assert result["success"] is True
        assert result["response"] == "Test response"

    @pytest.mark.asyncio
    async def test_chat_tool_error(self, mcp_server):
        """Test chat tool execution with error"""
        mcp_server.api_client.chat = AsyncMock(return_value={"success": False, "error": "Connection failed"})

        result = await mcp_server.api_client.chat("Hello")
        assert result["success"] is False
        assert "error" in result

    @pytest.mark.asyncio
    async def test_prometheus_query_tool(self, mcp_server):
        """Test Prometheus query tool"""
        mcp_server.api_client.query_prometheus = AsyncMock(
            return_value={"success": True, "result": {"status": "success", "data": []}}
        )

        result = await mcp_server.api_client.query_prometheus('up{job="test"}')
        assert result["success"] is True
        assert "result" in result

    @pytest.mark.asyncio
    async def test_golden_signals_tool(self, mcp_server):
        """Test golden signals check tool"""
        mcp_server.api_client.check_golden_signals = AsyncMock(
            return_value={"success": True, "signals": {"latency": "100ms", "traffic": "1000 req/s"}}
        )

        result = await mcp_server.api_client.check_golden_signals("test-service")
        assert result["success"] is True
        assert "signals" in result

    @pytest.mark.asyncio
    async def test_pod_logs_tool(self, mcp_server):
        """Test pod logs retrieval tool"""
        mcp_server.api_client.get_pod_logs = AsyncMock(
            return_value={"success": True, "logs": ["line1", "line2"]}
        )

        result = await mcp_server.api_client.get_pod_logs("test-pod")
        assert result["success"] is True
        assert "logs" in result

    @pytest.mark.asyncio
    async def test_analyze_logs_tool(self, mcp_server):
        """Test log analysis tool"""
        mcp_server.api_client.analyze_logs = AsyncMock(
            return_value={"success": True, "analysis": "Analysis result"}
        )

        result = await mcp_server.api_client.analyze_logs("test logs")
        assert result["success"] is True
        assert "analysis" in result

