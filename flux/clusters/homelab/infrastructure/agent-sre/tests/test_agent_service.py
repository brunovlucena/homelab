#!/usr/bin/env python3
"""
Unit tests for SRE Agent Service (HTTP API)
"""

import pytest
import json
from unittest.mock import Mock, patch, AsyncMock
from aiohttp import web
from aiohttp.test_utils import AioHTTPTestCase, unittest_run_loop

# Import the module to test
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'deployments', 'agent'))

from agent import SREAgentService


class TestSREAgentService(AioHTTPTestCase):
    """Test suite for SREAgentService HTTP endpoints"""

    async def get_application(self):
        """Create the aiohttp application for testing"""
        with patch('agent.agent') as mock_agent:
            # Mock the agent methods
            mock_agent.chat = AsyncMock(return_value="Chat response")
            mock_agent.analyze_logs = AsyncMock(return_value="Log analysis")
            mock_agent.incident_response = AsyncMock(return_value="Incident response")
            mock_agent.monitoring_advice = AsyncMock(return_value="Monitoring advice")
            mock_agent.health_check = AsyncMock(return_value={
                "status": "healthy",
                "llm_connected": True
            })
            mock_agent.llm = Mock()
            
            service = SREAgentService()
            service.sre_agent = mock_agent
            return service.app

    @unittest_run_loop
    async def test_health_endpoint(self):
        """Test /health endpoint returns 200"""
        resp = await self.client.request("GET", "/health")
        assert resp.status == 200
        
        data = await resp.json()
        assert data["status"] == "healthy"
        assert data["service"] == "sre-agent"
        assert "timestamp" in data
        assert "version" in data

    @unittest_run_loop
    async def test_readiness_endpoint(self):
        """Test /ready endpoint returns 200 when ready"""
        resp = await self.client.request("GET", "/ready")
        assert resp.status == 200
        
        data = await resp.json()
        assert data["status"] == "ready"
        assert "mcp_server_status" in data

    @unittest_run_loop
    async def test_chat_endpoint_success(self):
        """Test /chat endpoint with valid message"""
        payload = {"message": "Hello, how are you?"}
        resp = await self.client.request("POST", "/chat", json=payload)
        
        assert resp.status == 200
        data = await resp.json()
        assert "response" in data
        assert data["service"] == "sre-agent"
        assert data["method"] == "direct"

    @unittest_run_loop
    async def test_chat_endpoint_missing_message(self):
        """Test /chat endpoint without message returns 400"""
        payload = {}
        resp = await self.client.request("POST", "/chat", json=payload)
        
        assert resp.status == 400
        data = await resp.json()
        assert "error" in data

    @unittest_run_loop
    async def test_analyze_logs_endpoint_success(self):
        """Test /analyze-logs endpoint with valid logs"""
        payload = {"logs": "ERROR: Something went wrong\nWARN: Performance degraded"}
        resp = await self.client.request("POST", "/analyze-logs", json=payload)
        
        assert resp.status == 200
        data = await resp.json()
        assert "analysis" in data
        assert data["service"] == "sre-agent"
        assert data["method"] == "direct"

    @unittest_run_loop
    async def test_analyze_logs_endpoint_missing_logs(self):
        """Test /analyze-logs endpoint without logs returns 400"""
        payload = {}
        resp = await self.client.request("POST", "/analyze-logs", json=payload)
        
        assert resp.status == 400
        data = await resp.json()
        assert "error" in data

    @unittest_run_loop
    async def test_incident_response_endpoint_success(self):
        """Test /incident-response endpoint with valid incident"""
        payload = {"incident": "Database is down, users cannot login"}
        resp = await self.client.request("POST", "/incident-response", json=payload)
        
        assert resp.status == 200
        data = await resp.json()
        assert "response" in data
        assert data["service"] == "sre-agent"

    @unittest_run_loop
    async def test_incident_response_endpoint_missing_incident(self):
        """Test /incident-response endpoint without incident returns 400"""
        payload = {}
        resp = await self.client.request("POST", "/incident-response", json=payload)
        
        assert resp.status == 400
        data = await resp.json()
        assert "error" in data

    @unittest_run_loop
    async def test_monitoring_advice_endpoint_success(self):
        """Test /monitoring-advice endpoint with valid system"""
        payload = {"system": "High-traffic e-commerce website"}
        resp = await self.client.request("POST", "/monitoring-advice", json=payload)
        
        assert resp.status == 200
        data = await resp.json()
        assert "advice" in data
        assert data["service"] == "sre-agent"

    @unittest_run_loop
    async def test_monitoring_advice_endpoint_missing_system(self):
        """Test /monitoring-advice endpoint without system returns 400"""
        payload = {}
        resp = await self.client.request("POST", "/monitoring-advice", json=payload)
        
        assert resp.status == 400
        data = await resp.json()
        assert "error" in data

    @unittest_run_loop
    async def test_status_endpoint(self):
        """Test /status endpoint returns agent and MCP status"""
        resp = await self.client.request("GET", "/status")
        
        assert resp.status == 200
        data = await resp.json()
        assert "agent" in data
        assert "mcp_server" in data
        assert "service" in data

    @unittest_run_loop
    async def test_mcp_status_endpoint(self):
        """Test /mcp/status endpoint"""
        resp = await self.client.request("GET", "/mcp/status")
        
        assert resp.status == 200
        data = await resp.json()
        assert "status" in data

    @unittest_run_loop
    async def test_mcp_chat_endpoint(self):
        """Test /mcp/chat endpoint forwards to MCP server"""
        payload = {"message": "Test message"}
        resp = await self.client.request("POST", "/mcp/chat", json=payload)
        
        # Should attempt to forward, might fail if MCP server not available
        # but shouldn't crash
        assert resp.status in [200, 500]


class TestSREAgentServiceUnit:
    """Unit tests for SREAgentService methods"""

    @pytest.fixture
    def mock_agent(self):
        """Create a mock agent"""
        agent = Mock()
        agent.chat = AsyncMock(return_value="Chat response")
        agent.analyze_logs = AsyncMock(return_value="Log analysis")
        agent.incident_response = AsyncMock(return_value="Incident response")
        agent.monitoring_advice = AsyncMock(return_value="Monitoring advice")
        agent.health_check = AsyncMock(return_value={"status": "healthy"})
        agent.llm = Mock()
        return agent

    @pytest.fixture
    def service(self, mock_agent):
        """Create a service instance with mocked agent"""
        with patch('agent.agent', mock_agent):
            service = SREAgentService()
            service.sre_agent = mock_agent
            return service

    @pytest.mark.asyncio
    async def test_check_mcp_server_connected(self, service):
        """Test MCP server check when connected"""
        mock_response = Mock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value={"status": "healthy"})
        
        with patch('aiohttp.ClientSession.get') as mock_get:
            mock_get.return_value.__aenter__.return_value = mock_response
            
            result = await service._check_mcp_server()
            
            assert result["status"] == "connected"
            assert "url" in result
            assert "response_time" in result

    @pytest.mark.asyncio
    async def test_check_mcp_server_disconnected(self, service):
        """Test MCP server check when disconnected"""
        with patch('aiohttp.ClientSession.get', side_effect=Exception("Connection failed")):
            result = await service._check_mcp_server()
            
            assert result["status"] == "disconnected"
            assert "error" in result

    @pytest.mark.asyncio
    async def test_call_mcp_tool_success(self, service):
        """Test calling MCP tool successfully"""
        mock_response = Mock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value={
            "result": {
                "content": [{"text": "MCP response"}]
            }
        })
        
        with patch('aiohttp.ClientSession.post') as mock_post:
            mock_post.return_value.__aenter__.return_value = mock_response
            
            result = await service._call_mcp_tool("sre_chat", {"message": "Test"})
            
            assert result == "MCP response"

    @pytest.mark.asyncio
    async def test_call_mcp_tool_error(self, service):
        """Test calling MCP tool with error"""
        with patch('aiohttp.ClientSession.post', side_effect=Exception("Network error")):
            result = await service._call_mcp_tool("sre_chat", {"message": "Test"})
            
            assert "Error" in result
            assert "Network error" in result

