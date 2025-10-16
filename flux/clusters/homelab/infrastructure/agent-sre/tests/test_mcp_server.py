#!/usr/bin/env python3
"""
Unit tests for MCP Server
"""

import pytest
import json
from unittest.mock import Mock, patch, AsyncMock
from aiohttp.test_utils import AioHTTPTestCase, unittest_run_loop

# Import the module to test
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'deployments', 'mcp-server'))

from mcp_http_wrapper import MCPHTTPWrapper


class TestMCPServer(AioHTTPTestCase):
    """Test suite for MCP Server HTTP endpoints"""

    async def get_application(self):
        """Create the aiohttp application for testing"""
        server = MCPHTTPWrapper()
        return server.app

    @unittest_run_loop
    async def test_health_endpoint(self):
        """Test /health endpoint returns 200"""
        resp = await self.client.request("GET", "/health")
        assert resp.status == 200
        
        data = await resp.json()
        assert data["status"] == "healthy"
        assert data["service"] == "agent-sre-mcp-http-wrapper"
        assert "timestamp" in data

    @unittest_run_loop
    async def test_mcp_info_endpoint(self):
        """Test GET /tools endpoint returns available tools"""
        resp = await self.client.request("GET", "/tools")
        assert resp.status == 200
        
        data = await resp.json()
        assert "tools" in data
        assert "count" in data
        assert len(data["tools"]) > 0

    @unittest_run_loop
    async def test_mcp_initialize(self):
        """Test MCP initialize method"""
        payload = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {}
        }
        resp = await self.client.request("POST", "/mcp", json=payload)
        
        assert resp.status == 200
        data = await resp.json()
        assert data["jsonrpc"] == "2.0"
        assert data["id"] == 1
        assert "result" in data
        assert data["result"]["protocolVersion"] == "2024-11-05"
        assert "capabilities" in data["result"]

    @unittest_run_loop
    async def test_mcp_tools_list(self):
        """Test MCP tools/list method"""
        payload = {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/list",
            "params": {}
        }
        resp = await self.client.request("POST", "/mcp", json=payload)
        
        assert resp.status == 200
        data = await resp.json()
        assert data["jsonrpc"] == "2.0"
        assert "result" in data
        assert "tools" in data["result"]
        
        tools = data["result"]["tools"]
        tool_names = [t["name"] for t in tools]
        assert "sre_chat" in tool_names
        assert "analyze_logs" in tool_names
        assert "incident_response" in tool_names
        assert "monitoring_advice" in tool_names
        assert "health_check" in tool_names

    @unittest_run_loop
    async def test_mcp_notification(self):
        """Test MCP notification (no id field)"""
        payload = {
            "jsonrpc": "2.0",
            "method": "notifications/initialized"
        }
        resp = await self.client.request("POST", "/mcp", json=payload)
        
        assert resp.status == 200
        # Notifications return empty response

    @unittest_run_loop
    async def test_mcp_unknown_method(self):
        """Test MCP with unknown method returns error"""
        payload = {
            "jsonrpc": "2.0",
            "id": 99,
            "method": "unknown/method",
            "params": {}
        }
        resp = await self.client.request("POST", "/mcp", json=payload)
        
        assert resp.status == 200
        data = await resp.json()
        assert "error" in data
        assert data["error"]["code"] == -32601

    @unittest_run_loop
    async def test_mcp_malformed_request(self):
        """Test MCP with malformed request"""
        payload = {
            "not": "valid"
        }
        resp = await self.client.request("POST", "/mcp", json=payload)
        
        assert resp.status == 400
        data = await resp.json()
        assert "error" in data

    @unittest_run_loop
    async def test_readiness_endpoint(self):
        """Test /ready endpoint"""
        resp = await self.client.request("GET", "/ready")
        
        # Might return 503 if agent service is not available
        assert resp.status in [200, 503]
        data = await resp.json()
        assert "status" in data


class TestMCPServerUnit:
    """Unit tests for MCPServer methods"""

    @pytest.fixture
    def server(self):
        """Create an MCPHTTPWrapper instance"""
        return MCPHTTPWrapper()

    @pytest.mark.asyncio
    async def test_check_agent_service_connected(self, server):
        """Test agent service check when connected"""
        mock_response = Mock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value={"status": "healthy"})
        
        with patch('aiohttp.ClientSession.get') as mock_get:
            mock_get.return_value.__aenter__.return_value = mock_response
            
            result = await server._check_agent_service()
            
            assert result["status"] == "connected"
            assert "url" in result
            assert "health" in result

    @pytest.mark.asyncio
    async def test_check_agent_service_disconnected(self, server):
        """Test agent service check when disconnected"""
        with patch('aiohttp.ClientSession.get', side_effect=Exception("Connection failed")):
            result = await server._check_agent_service()
            
            assert result["status"] == "disconnected"
            assert "error" in result

    @pytest.mark.asyncio
    async def test_forward_to_agent_chat(self, server):
        """Test forwarding chat tool to agent"""
        mock_response = Mock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value={"response": "Test response"})
        
        with patch('aiohttp.ClientSession.post') as mock_post:
            mock_post.return_value.__aenter__.return_value = mock_response
            
            result = await server._forward_to_agent("sre_chat", {"message": "Hello"})
            
            assert result == "Test response"

    @pytest.mark.asyncio
    async def test_forward_to_agent_analyze_logs(self, server):
        """Test forwarding analyze_logs tool to agent"""
        mock_response = Mock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value={"analysis": "Log analysis result"})
        
        with patch('aiohttp.ClientSession.post') as mock_post:
            mock_post.return_value.__aenter__.return_value = mock_response
            
            result = await server._forward_to_agent("analyze_logs", {"logs": "ERROR: test"})
            
            assert result == "Log analysis result"

    @pytest.mark.asyncio
    async def test_forward_to_agent_health_check(self, server):
        """Test forwarding health_check tool to agent"""
        mock_response = Mock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value={"status": "healthy"})
        
        with patch('aiohttp.ClientSession.get') as mock_get:
            mock_get.return_value.__aenter__.return_value = mock_response
            
            result = await server._forward_to_agent("health_check", {})
            
            assert "status" in result
            assert "healthy" in result

    @pytest.mark.asyncio
    async def test_forward_to_agent_unknown_tool(self, server):
        """Test forwarding unknown tool returns error"""
        result = await server._forward_to_agent("unknown_tool", {})
        
        assert "Unknown tool" in result

    @pytest.mark.asyncio
    async def test_forward_to_agent_network_error(self, server):
        """Test forwarding with network error"""
        with patch('aiohttp.ClientSession.post', side_effect=Exception("Network error")):
            result = await server._forward_to_agent("sre_chat", {"message": "Test"})
            
            assert "Error" in result
            assert "Network error" in result

