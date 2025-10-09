#!/usr/bin/env python3
"""
SRE Agent Core with LangGraph
State-managed agent for SRE tasks using LangGraph and Ollama
"""

import os
import json
import asyncio
import logging
from typing import Dict, Any, List, Optional, TypedDict, Annotated, Sequence
from datetime import datetime
from operator import add

# LangChain imports
from langchain_ollama import ChatOllama
from langchain_core.messages import BaseMessage, HumanMessage, SystemMessage, AIMessage
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder

# LangGraph imports
from langgraph.graph import StateGraph, END, START
from langgraph.graph.message import add_messages
from langgraph.checkpoint.memory import MemorySaver

# LangSmith imports for tracing
from langsmith import traceable

# Logfire imports
import logfire

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Configuration
OLLAMA_URL = os.environ.get("OLLAMA_URL", "http://192.168.0.3:11434")
MODEL_NAME = os.environ.get("MODEL_NAME", "bruno-sre:latest")
SERVICE_NAME = os.environ.get("SERVICE_NAME", "sre-agent")

# Configure Logfire
sre_agent_token = os.getenv('LOGFIRE_TOKEN_SRE_AGENT')
if sre_agent_token:
    try:
        logfire.configure(service_name=SERVICE_NAME, token=sre_agent_token)
        logger.info("✅ Logfire configured successfully")
    except Exception as e:
        logger.warning(f"⚠️  Logfire configuration failed: {e}")
        logger.warning("⚠️  Continuing without Logfire...")
        os.environ.pop('LOGFIRE_TOKEN_SRE_AGENT', None)
else:
    logger.warning("⚠️  LOGFIRE_TOKEN_SRE_AGENT not set, skipping Logfire configuration")

# Configure LangChain API key
langsmith_api_key = os.getenv('LANGSMITH_API_KEY')
if langsmith_api_key:
    os.environ['LANGCHAIN_API_KEY'] = langsmith_api_key
    logger.info("✅ LangSmith API key configured from environment")
else:
    logger.warning("⚠️  LANGSMITH_API_KEY not set, LangSmith features will be limited")

# Initialize Ollama LLM with LangGraph compatibility
try:
    llm = ChatOllama(
        model=MODEL_NAME,
        base_url=OLLAMA_URL,
        temperature=0.7,
        num_ctx=8192,
    )
    logger.info(f"✅ Ollama connection established: {OLLAMA_URL}")
except Exception as e:
    logger.error(f"❌ Error connecting to Ollama: {e}")
    llm = None


# Define the state for the graph
class AgentState(TypedDict):
    """🔧 State for the SRE Agent"""
    messages: Annotated[Sequence[BaseMessage], add_messages]
    task_type: str
    context: Dict[str, Any]
    analysis_result: Optional[str]
    recommendations: List[str]
    next_action: Optional[str]


class SREAgentGraph:
    """🤖 SRE Agent with LangGraph state management"""
    
    def __init__(self):
        self.llm = llm
        self.service_name = SERVICE_NAME
        self.checkpointer = MemorySaver()
        self.graph = self._build_graph()
        
    def _build_graph(self) -> StateGraph:
        """🏗️ Build the LangGraph workflow"""
        
        # Create the graph
        workflow = StateGraph(AgentState)
        
        # Add nodes
        workflow.add_node("analyze", self._analyze_node)
        workflow.add_node("generate_recommendations", self._generate_recommendations_node)
        workflow.add_node("format_response", self._format_response_node)
        
        # Define the flow
        workflow.add_edge(START, "analyze")
        workflow.add_edge("analyze", "generate_recommendations")
        workflow.add_edge("generate_recommendations", "format_response")
        workflow.add_edge("format_response", END)
        
        return workflow.compile(checkpointer=self.checkpointer)
    
    @traceable(name="sre_analyze_node", run_type="llm")
    @logfire.instrument("analyze_node")
    async def _analyze_node(self, state: AgentState) -> AgentState:
        """🔍 Analyze the input and extract key insights"""
        if not self.llm:
            state["analysis_result"] = "Error: Ollama connection not available"
            return state
        
        task_type = state.get("task_type", "general")
        messages = state.get("messages", [])
        
        # Create analysis prompt based on task type
        system_prompt = self._get_system_prompt(task_type)
        
        # Invoke LLM for analysis
        try:
            analysis_messages = [
                SystemMessage(content=system_prompt),
                *messages
            ]
            
            response = await self.llm.ainvoke(analysis_messages)
            state["analysis_result"] = response.content
            state["messages"] = state["messages"] + [response]
            
        except Exception as e:
            logger.error(f"❌ Error in analysis node: {e}")
            state["analysis_result"] = f"Error during analysis: {str(e)}"
        
        return state
    
    @traceable(name="sre_generate_recommendations", run_type="llm")
    @logfire.instrument("generate_recommendations")
    async def _generate_recommendations_node(self, state: AgentState) -> AgentState:
        """💡 Generate actionable recommendations"""
        if not self.llm:
            state["recommendations"] = ["Error: Ollama connection not available"]
            return state
        
        analysis_result = state.get("analysis_result", "")
        task_type = state.get("task_type", "general")
        
        # Create recommendations prompt
        rec_prompt = f"""
        Based on the following analysis for a {task_type} task, provide 3-5 specific, actionable recommendations:
        
        Analysis:
        {analysis_result}
        
        Format your recommendations as a numbered list with clear action items.
        """
        
        try:
            response = await self.llm.ainvoke([HumanMessage(content=rec_prompt)])
            
            # Parse recommendations from response
            recommendations = []
            for line in response.content.split('\n'):
                line = line.strip()
                if line and (line[0].isdigit() or line.startswith('-') or line.startswith('•')):
                    recommendations.append(line)
            
            state["recommendations"] = recommendations if recommendations else [response.content]
            state["messages"] = state["messages"] + [response]
            
        except Exception as e:
            logger.error(f"❌ Error generating recommendations: {e}")
            state["recommendations"] = [f"Error generating recommendations: {str(e)}"]
        
        return state
    
    @logfire.instrument("format_response")
    async def _format_response_node(self, state: AgentState) -> AgentState:
        """📝 Format the final response"""
        analysis = state.get("analysis_result", "No analysis available")
        recommendations = state.get("recommendations", [])
        
        formatted_response = f"""
## 🔍 Analysis

{analysis}

## 💡 Recommendations

{chr(10).join(recommendations) if recommendations else 'No recommendations available'}
        """.strip()
        
        state["messages"] = state["messages"] + [
            AIMessage(content=formatted_response)
        ]
        state["next_action"] = "complete"
        
        return state
    
    def _get_system_prompt(self, task_type: str) -> str:
        """📋 Get system prompt based on task type"""
        
        base_prompt = """You are an expert SRE (Site Reliability Engineering) AI assistant.
You help with monitoring, troubleshooting, incident response, and maintaining system reliability.
You provide clear, actionable insights based on SRE best practices and observability principles."""
        
        prompts = {
            "logs": f"""{base_prompt}
            
Your current task is to analyze logs and identify:
1. Critical errors and warnings
2. Patterns and anomalies
3. Potential root causes
4. Performance issues
            """,
            
            "incident": f"""{base_prompt}
            
Your current task is incident response. Focus on:
1. Immediate impact assessment
2. Quick mitigation steps
3. Root cause investigation
4. Communication strategy
5. Post-incident actions
            """,
            
            "monitoring": f"""{base_prompt}
            
Your current task is to provide monitoring advice. Focus on:
1. Key metrics to track
2. Alert thresholds and conditions
3. Dashboard design
4. Observability strategy
5. SLIs and SLOs
            """,
            
            "performance": f"""{base_prompt}
            
Your current task is performance analysis. Focus on:
1. Bottleneck identification
2. Resource utilization
3. Optimization opportunities
4. Scalability concerns
5. Cost optimization
            """,
        }
        
        return prompts.get(task_type, base_prompt)
    
    @traceable(name="sre_execute", run_type="chain")
    @logfire.instrument("execute")
    async def execute(
        self, 
        message: str, 
        task_type: str = "general",
        context: Optional[Dict[str, Any]] = None,
        thread_id: str = "default"
    ) -> Dict[str, Any]:
        """🚀 Execute the SRE agent workflow"""
        
        initial_state: AgentState = {
            "messages": [HumanMessage(content=message)],
            "task_type": task_type,
            "context": context or {},
            "analysis_result": None,
            "recommendations": [],
            "next_action": None,
        }
        
        # Execute the graph
        config = {"configurable": {"thread_id": thread_id}}
        final_state = await self.graph.ainvoke(initial_state, config)
        
        return {
            "analysis": final_state.get("analysis_result"),
            "recommendations": final_state.get("recommendations"),
            "full_response": final_state["messages"][-1].content if final_state.get("messages") else None,
            "task_type": task_type,
            "timestamp": datetime.now().isoformat(),
        }
    
    @logfire.instrument("health_check")
    async def health_check(self) -> Dict[str, Any]:
        """❤️ Check agent health status"""
        return {
            "status": "healthy",
            "service": self.service_name,
            "timestamp": datetime.now().isoformat(),
            "ollama_url": OLLAMA_URL,
            "model_name": MODEL_NAME,
            "llm_connected": self.llm is not None,
            "graph_compiled": self.graph is not None,
            "langgraph_enabled": True,
        }
    
    # Legacy methods for backward compatibility
    @traceable(name="sre_chat_legacy", run_type="chain")
    @logfire.instrument("chat_legacy")
    async def chat(self, message: str) -> str:
        """💬 Legacy chat method - forwards to LangGraph"""
        result = await self.execute(message=message, task_type="general")
        return result.get("full_response", "No response")
    
    @traceable(name="sre_analyze_logs_legacy", run_type="chain")
    @logfire.instrument("analyze_logs_legacy")
    async def analyze_logs(self, logs: str) -> str:
        """📊 Legacy log analysis method - forwards to LangGraph"""
        message = f"Analyze these logs and provide insights:\n\n{logs}"
        result = await self.execute(message=message, task_type="logs")
        return result.get("full_response", "No analysis")
    
    @traceable(name="sre_incident_response_legacy", run_type="chain")
    @logfire.instrument("incident_response_legacy")
    async def incident_response(self, incident: str) -> str:
        """🚨 Legacy incident response method - forwards to LangGraph"""
        result = await self.execute(message=incident, task_type="incident")
        return result.get("full_response", "No response")
    
    @traceable(name="sre_monitoring_advice_legacy", run_type="chain")
    @logfire.instrument("monitoring_advice_legacy")
    async def monitoring_advice(self, system: str) -> str:
        """📈 Legacy monitoring advice method - forwards to LangGraph"""
        message = f"Provide monitoring and observability advice for: {system}"
        result = await self.execute(message=message, task_type="monitoring")
        return result.get("full_response", "No advice")


# Global agent instance
agent = SREAgentGraph()

# Export for use in other modules
__all__ = ['agent', 'logger', 'logfire', 'SREAgentGraph']
