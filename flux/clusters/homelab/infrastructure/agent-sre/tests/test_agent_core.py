#!/usr/bin/env python3
"""
Unit tests for SRE Agent Core (LangGraph)
"""

import pytest
import asyncio
from unittest.mock import Mock, patch, AsyncMock
from datetime import datetime

# Import the module to test
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'deployments', 'agent'))

from core import SREAgentGraph, AgentState


class TestSREAgentGraph:
    """Test suite for SREAgentGraph"""

    @pytest.fixture
    def mock_llm(self):
        """Mock LLM for testing"""
        mock = AsyncMock()
        mock.ainvoke = AsyncMock(return_value=Mock(content="Test response"))
        return mock

    @pytest.fixture
    def agent_graph(self, mock_llm):
        """Create an SREAgentGraph instance with mocked LLM"""
        with patch('core.llm', mock_llm):
            agent = SREAgentGraph()
            agent.llm = mock_llm
            return agent

    def test_initialization(self, agent_graph):
        """Test that SREAgentGraph initializes correctly"""
        assert agent_graph is not None
        assert agent_graph.service_name is not None
        assert agent_graph.graph is not None
        assert agent_graph.checkpointer is not None

    def test_get_system_prompt_logs(self, agent_graph):
        """Test system prompt generation for logs task"""
        prompt = agent_graph._get_system_prompt("logs")
        assert "logs" in prompt.lower()
        assert "errors" in prompt.lower() or "warnings" in prompt.lower()

    def test_get_system_prompt_incident(self, agent_graph):
        """Test system prompt generation for incident task"""
        prompt = agent_graph._get_system_prompt("incident")
        assert "incident" in prompt.lower()
        assert "response" in prompt.lower()

    def test_get_system_prompt_monitoring(self, agent_graph):
        """Test system prompt generation for monitoring task"""
        prompt = agent_graph._get_system_prompt("monitoring")
        assert "monitoring" in prompt.lower()
        assert "metrics" in prompt.lower() or "alert" in prompt.lower()

    def test_get_system_prompt_default(self, agent_graph):
        """Test system prompt generation for unknown task type"""
        prompt = agent_graph._get_system_prompt("unknown_task")
        assert "SRE" in prompt
        assert "reliability" in prompt.lower()

    @pytest.mark.asyncio
    async def test_analyze_node_success(self, agent_graph, mock_llm):
        """Test analyze node with successful LLM response"""
        from langchain_core.messages import HumanMessage, AIMessage
        
        mock_llm.ainvoke.return_value = Mock(content="Analysis result")
        
        state: AgentState = {
            "messages": [HumanMessage(content="Test message")],
            "task_type": "logs",
            "context": {},
            "analysis_result": None,
            "recommendations": [],
            "next_action": None,
        }
        
        result = await agent_graph._analyze_node(state)
        
        assert result["analysis_result"] == "Analysis result"
        assert len(result["messages"]) == 2  # Original + AI response

    @pytest.mark.asyncio
    async def test_analyze_node_no_llm(self, agent_graph):
        """Test analyze node when LLM is not available"""
        from langchain_core.messages import HumanMessage
        
        agent_graph.llm = None
        
        state: AgentState = {
            "messages": [HumanMessage(content="Test message")],
            "task_type": "logs",
            "context": {},
            "analysis_result": None,
            "recommendations": [],
            "next_action": None,
        }
        
        result = await agent_graph._analyze_node(state)
        
        assert "Error" in result["analysis_result"]
        assert "Ollama" in result["analysis_result"]

    @pytest.mark.asyncio
    async def test_generate_recommendations_node(self, agent_graph, mock_llm):
        """Test recommendations generation"""
        mock_llm.ainvoke.return_value = Mock(
            content="1. First recommendation\n2. Second recommendation\n3. Third recommendation"
        )
        
        state: AgentState = {
            "messages": [],
            "task_type": "logs",
            "context": {},
            "analysis_result": "Some analysis",
            "recommendations": [],
            "next_action": None,
        }
        
        result = await agent_graph._generate_recommendations_node(state)
        
        assert len(result["recommendations"]) > 0
        assert any("recommendation" in rec.lower() for rec in result["recommendations"])

    @pytest.mark.asyncio
    async def test_format_response_node(self, agent_graph):
        """Test response formatting"""
        state: AgentState = {
            "messages": [],
            "task_type": "logs",
            "context": {},
            "analysis_result": "Test analysis",
            "recommendations": ["Rec 1", "Rec 2"],
            "next_action": None,
        }
        
        result = await agent_graph._format_response_node(state)
        
        assert result["next_action"] == "complete"
        assert len(result["messages"]) == 1
        assert "Analysis" in result["messages"][-1].content
        assert "Recommendations" in result["messages"][-1].content

    @pytest.mark.asyncio
    async def test_health_check(self, agent_graph):
        """Test health check returns valid status"""
        health = await agent_graph.health_check()
        
        assert "status" in health
        assert health["status"] == "healthy"
        assert "service" in health
        assert "timestamp" in health
        assert "llm_connected" in health
        assert "langgraph_enabled" in health
        assert health["langgraph_enabled"] is True

    @pytest.mark.asyncio
    async def test_execute_with_context(self, agent_graph, mock_llm):
        """Test execute method with context"""
        # Mock the graph execution
        mock_result = {
            "messages": [Mock(content="Final response")],
            "analysis_result": "Test analysis",
            "recommendations": ["Rec 1"],
            "task_type": "logs",
            "next_action": "complete"
        }
        
        with patch.object(agent_graph.graph, 'ainvoke', AsyncMock(return_value=mock_result)):
            result = await agent_graph.execute(
                message="Test message",
                task_type="logs",
                context={"key": "value"}
            )
            
            assert "analysis" in result
            assert "recommendations" in result
            assert "full_response" in result
            assert "task_type" in result
            assert result["task_type"] == "logs"

    @pytest.mark.asyncio
    async def test_chat_method(self, agent_graph):
        """Test chat method forwards to execute"""
        with patch.object(agent_graph, 'execute', AsyncMock(return_value={"full_response": "Chat response"})):
            result = await agent_graph.chat("Hello")
            
            assert result == "Chat response"

    @pytest.mark.asyncio
    async def test_analyze_logs_method(self, agent_graph):
        """Test analyze_logs method forwards to execute"""
        with patch.object(agent_graph, 'execute', AsyncMock(return_value={"full_response": "Log analysis"})):
            result = await agent_graph.analyze_logs("Test logs")
            
            assert result == "Log analysis"

    @pytest.mark.asyncio
    async def test_incident_response_method(self, agent_graph):
        """Test incident_response method forwards to execute"""
        with patch.object(agent_graph, 'execute', AsyncMock(return_value={"full_response": "Incident response"})):
            result = await agent_graph.incident_response("Critical incident")
            
            assert result == "Incident response"

    @pytest.mark.asyncio
    async def test_monitoring_advice_method(self, agent_graph):
        """Test monitoring_advice method forwards to execute"""
        with patch.object(agent_graph, 'execute', AsyncMock(return_value={"full_response": "Monitoring advice"})):
            result = await agent_graph.monitoring_advice("Web application")
            
            assert result == "Monitoring advice"


class TestAgentState:
    """Test AgentState TypedDict"""

    def test_agent_state_structure(self):
        """Test that AgentState has expected keys"""
        from typing import get_type_hints
        
        hints = get_type_hints(AgentState)
        
        assert "messages" in hints
        assert "task_type" in hints
        assert "context" in hints
        assert "analysis_result" in hints
        assert "recommendations" in hints
        assert "next_action" in hints

