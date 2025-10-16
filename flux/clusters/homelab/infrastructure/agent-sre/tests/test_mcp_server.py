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
    async def test_mcp_tools_call(self):
        """Test MCP tools/call method"""
        payload = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": "prometheus_query",
                "arguments": {
                    "query": "up"
                }
            }
        }
        resp = await self.client.request("POST", "/mcp", json=payload)
        
        assert resp.status == 200
        data = await resp.json()
        assert data["jsonrpc"] == "2.0"
        assert data["id"] == 1
        assert "result" in data

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
        assert "prometheus_query" in tool_names
        assert "grafana_query" in tool_names
        assert "prometheus_query_range" in tool_names

    @unittest_run_loop
    async def test_mcp_invalid_method(self):
        """Test MCP with invalid method field returns 404"""
        payload = {
            "jsonrpc": "2.0",
            "id": 99,
            "method": "invalid/method",
            "params": {}
        }
        resp = await self.client.request("POST", "/mcp", json=payload)
        
        assert resp.status == 404
        data = await resp.json()
        assert "error" in data

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
        
        assert resp.status == 404
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
    """Unit tests for MCPHTTPWrapper methods"""

    @pytest.fixture
    def server(self):
        """Create an MCPHTTPWrapper instance"""
        return MCPHTTPWrapper()

    @pytest.mark.asyncio
    async def test_handle_tools_list(self, server):
        """Test tools list handler"""
        result = await server._handle_tools_list({})
        
        assert "tools" in result
        assert isinstance(result["tools"], list)
        assert len(result["tools"]) > 0
        
        # Check that all required tools are present
        tool_names = [t["name"] for t in result["tools"]]
        assert "prometheus_query" in tool_names
        assert "prometheus_query_range" in tool_names
        assert "grafana_query" in tool_names

    @pytest.mark.asyncio
    async def test_handle_tools_call_unknown_tool(self, server):
        """Test calling unknown tool returns error"""
        result = await server._handle_tools_call({"name": "unknown_tool", "arguments": {}})
        
        assert "content" in result
        assert result["isError"] is True
        assert "Unknown tool" in result["content"][0]["text"]

    @pytest.mark.asyncio
    async def test_handle_tools_call_no_tool_name(self, server):
        """Test calling tool without name raises error"""
        with pytest.raises(ValueError, match="Tool name is required"):
            await server._handle_tools_call({"arguments": {}})

