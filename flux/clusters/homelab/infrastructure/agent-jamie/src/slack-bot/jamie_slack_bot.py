#!/usr/bin/env python3
"""
🤖 Jamie Slack Bot
A sophisticated SRE assistant
Calls Agent-SRE Agent Service MCP Server for tool execution
"""

import os
import json
import asyncio
from typing import Dict, List, Optional, Any, Annotated
from datetime import datetime, UTC

import aiohttp
from aiohttp import web
from slack_bolt.async_app import AsyncApp
from slack_bolt.adapter.socket_mode.async_handler import AsyncSocketModeHandler
from slack_sdk.errors import SlackApiError

# LangGraph imports for agent workflow
from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages
from typing_extensions import TypedDict

# LangChain imports
from langchain_core.messages import HumanMessage, AIMessage, SystemMessage, ToolMessage
from langchain_ollama import ChatOllama

# Import from core
from core import logger, logfire, OLLAMA_URL, MODEL_NAME, SERVICE_NAME, AGENT_SRE_URL


class AgentSREClient:
    """🔌 Client for interacting with Agent-SRE MCP Server using JSON-RPC 2.0 protocol
    
    This client connects to the Agent-SRE MCP server to call tools like:
    - prometheus_query: Query Prometheus metrics via PromQL
    - grafana_query: Query Grafana dashboards and datasources
    
    Uses proper MCP JSON-RPC 2.0 protocol for tool discovery and calling.
    """
    
    def __init__(self, base_url: str = None):
        # Agent-SRE Service URL (for HTTP fallback)
        self.base_url = base_url or AGENT_SRE_URL
        self.http_session = None
        
        # Configuration for MCP server connection
        self.mcp_server_host = os.getenv("MCP_SERVER_HOST", "agent-sre-mcp-server.agent-sre")
        self.mcp_server_port = int(os.getenv("MCP_SERVER_PORT", "3000"))
        self.mcp_url = f"http://{self.mcp_server_host}:{self.mcp_server_port}/mcp"
        
        # Cache for available tools
        self._available_tools: Optional[List[Dict]] = None
        self._request_id = 0
    
    async def __aenter__(self):
        """Initialize HTTP session"""
        self.http_session = aiohttp.ClientSession()
        
        logger.info(f"🔌 AgentSREClient initialized - MCP endpoint: {self.mcp_url}")
        
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Cleanup HTTP session"""
        if self.http_session:
            await self.http_session.close()
    
    def _get_next_request_id(self) -> int:
        """Get next JSON-RPC request ID"""
        self._request_id += 1
        return self._request_id
    
    @logfire.instrument("mcp_tools_list")
    async def list_tools(self) -> List[Dict]:
        """📋 Discover available MCP tools using JSON-RPC 2.0 protocol"""
        if not self.http_session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        # Return cached tools if available
        if self._available_tools is not None:
            return self._available_tools
        
        try:
            logger.info("📋 Discovering MCP tools...")
            
            # JSON-RPC 2.0 request for tools/list
            payload = {
                "jsonrpc": "2.0",
                "id": self._get_next_request_id(),
                "method": "tools/list",
                "params": {}
            }
            
            async with self.http_session.post(
                self.mcp_url,
                json=payload,
                timeout=aiohttp.ClientTimeout(total=10)
            ) as response:
                if response.status == 200:
                    data = await response.json()
                    
                    if "error" in data:
                        logger.error(f"❌ MCP tools/list error: {data['error']}")
                        return []
                    
                    result = data.get("result", {})
                    tools = result.get("tools", [])
                    
                    logger.info(f"✅ Discovered {len(tools)} MCP tools")
                    for tool in tools:
                        logger.info(f"   🔧 {tool['name']}: {tool['description'][:60]}...")
                    
                    # Cache the tools
                    self._available_tools = tools
                    return tools
                else:
                    error_text = await response.text()
                    logger.error(f"❌ MCP tools/list failed: {response.status} - {error_text}")
                    return []
        
        except Exception as e:
            logger.error(f"❌ Error discovering MCP tools: {e}", exc_info=True)
            return []
    
    @logfire.instrument("mcp_tool_call")
    async def call_tool(self, tool_name: str, arguments: Dict[str, Any]) -> Dict:
        """🔧 Call an MCP tool using JSON-RPC 2.0 protocol"""
        if not self.http_session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        try:
            logger.info(f"🔧 Calling MCP tool: {tool_name}")
            logger.debug(f"🔧 Tool arguments: {arguments}")
            
            # JSON-RPC 2.0 request for tools/call
            payload = {
                "jsonrpc": "2.0",
                "id": self._get_next_request_id(),
                "method": "tools/call",
                "params": {
                    "name": tool_name,
                    "arguments": arguments
                }
            }
            
            async with self.http_session.post(
                self.mcp_url,
                json=payload,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                if response.status == 200:
                    data = await response.json()
                    
                    if "error" in data:
                        logger.error(f"❌ MCP tools/call error: {data['error']}")
                        return {
                            "response": f"MCP error: {data['error']['message']}",
                            "error": True
                        }
                    
                    result = data.get("result", {})
                    content = result.get("content", [])
                    is_error = result.get("isError", False)
                    
                    # Extract text from content
                    response_text = ""
                    for item in content:
                        if item.get("type") == "text":
                            response_text += item.get("text", "")
                    
                    logger.info(f"✅ MCP tool {tool_name} completed - isError: {is_error}")
                    
                    return {
                        "response": response_text,
                        "error": is_error,
                        "raw_result": result
                    }
                else:
                    error_text = await response.text()
                    logger.error(f"❌ MCP tool {tool_name} failed: {response.status} - {error_text}")
                    return {
                        "response": f"MCP tool error: {error_text}",
                        "error": True
                    }
        
        except Exception as e:
            logger.error(f"❌ Error calling MCP tool {tool_name}: {e}", exc_info=True)
            return {
                "response": f"Error calling MCP tool: {str(e)}",
                "error": True
            }
    
    async def _generate_promql_with_llm(self, natural_query: str) -> str:
        """🤖 Use LLM to generate PromQL query from natural language"""
        try:
            # Call Ollama to generate PromQL
            prompt = f"""Convert this natural language query to valid PromQL syntax:

Natural language: "{natural_query}"

Generate a valid PromQL query that would answer this question. 
Focus on common Kubernetes metrics like:
- CPU usage: rate(container_cpu_usage_seconds_total{{namespace="agent-sre"}}[5m])
- Memory usage: container_memory_working_set_bytes{{namespace="agent-sre"}}
- Pod status: kube_pod_info{{namespace="agent-sre"}}
- Service health: up{{namespace="agent-sre"}}

Return ONLY the PromQL query, nothing else."""

            async with self.http_session.post(
                f"{OLLAMA_URL}/api/generate",
                json={
                    "model": MODEL_NAME,
                    "prompt": prompt,
                    "stream": False
                },
                timeout=aiohttp.ClientTimeout(total=10)
            ) as response:
                if response.status == 200:
                    result = await response.json()
                    promql = result.get("response", "").strip()
                    logger.info(f"🤖 LLM generated PromQL: {promql}")
                    return promql
                else:
                    logger.error(f"❌ LLM call failed: {response.status}")
                    return 'up{namespace="agent-sre"}'  # fallback
        except Exception as e:
            logger.error(f"❌ Error calling LLM: {e}")
            return 'up{namespace="agent-sre"}'  # fallback

    @logfire.instrument("sre_prometheus_query")
    async def prometheus_query(self, query: str, context: Optional[Dict] = None) -> Dict:
        """🔍 Query Prometheus via MCP tool"""
        if not self.http_session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        logger.info(f"🔍 Prometheus query: {query}")
        
        # Use LLM to generate PromQL from natural language
        promql_query = await self._generate_promql_with_llm(query)
        logger.info(f"🔧 Generated PromQL: {promql_query}")
        
        # Call prometheus_query MCP tool with properly formatted JSON
        arguments = {"query": promql_query}
        
        # Add optional parameters from context
        if context:
            if "time" in context:
                arguments["time"] = context["time"]
            if "timeout" in context:
                arguments["timeout"] = context["timeout"]
        
        result = await self.call_tool("prometheus_query", arguments)
        
        if result.get("error"):
            return {
                "response": "I'm having trouble querying Prometheus.",
                "error": True,
                "details": result.get("response")
            }
        
        return {
            "response": result.get("response"),
            "error": False
        }
    
    @logfire.instrument("sre_grafana_query")
    async def grafana_query(self, query: str, query_type: str = "search", context: Optional[Dict] = None) -> Dict:
        """📊 Query Grafana via MCP tool"""
        if not self.http_session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        logger.info(f"📊 Grafana query: type={query_type}, query={query}")
        
        # Call grafana_query MCP tool
        arguments = {
            "query_type": query_type,
            "query": query
        }
        
        # Add optional parameters from context
        if context:
            if "dashboard_id" in context:
                arguments["dashboard_id"] = context["dashboard_id"]
            if "panel_id" in context:
                arguments["panel_id"] = context["panel_id"]
            if "from_time" in context:
                arguments["from_time"] = context["from_time"]
            if "to_time" in context:
                arguments["to_time"] = context["to_time"]
        
        result = await self.call_tool("grafana_query", arguments)
        
        if result.get("error"):
            return {
                "response": "I'm having trouble querying Grafana.",
                "error": True,
                "details": result.get("response")
            }
        
        return {
            "response": result.get("response"),
            "error": False
        }
    
    @logfire.instrument("sre_chat")
    async def chat(self, message: str, context: Optional[Dict] = None) -> Dict:
        """💬 Send chat message to Agent-SRE Service (HTTP fallback)"""
        if not self.http_session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        payload = {"message": message}
        
        try:
            async with self.http_session.post(
                f"{self.base_url}/chat",
                json=payload,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                if response.status == 200:
                    result = await response.json()
                    return {"response": result.get("response", "No response")}
                else:
                    error_text = await response.text()
                    logger.error(f"Agent-SRE Service error: {response.status} - {error_text}")
                    return {
                        "response": "I'm having trouble connecting to the SRE agent.",
                        "error": True
                    }
        except Exception as e:
            logger.error(f"Agent-SRE Service error: {e}")
            return {
                "response": "I'm experiencing technical difficulties connecting to the SRE agent.",
                "error": True
            }
    
    @logfire.instrument("sre_k8s_query")
    async def k8s_query(self, message: str, context: Optional[Dict] = None) -> Dict:
        """☸️ Send Kubernetes query to Agent-SRE Service (HTTP fallback)"""
        if not self.http_session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        payload = {"query": message, "context": context or {}}
        
        try:
            async with self.http_session.post(
                f"{self.base_url}/k8s/query",
                json=payload,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                if response.status == 200:
                    result = await response.json()
                    return {"response": result.get("result", "No result")}
                else:
                    error_text = await response.text()
                    logger.error(f"Agent-SRE K8s error: {response.status} - {error_text}")
                    return {
                        "response": "I'm having trouble querying Kubernetes.",
                        "error": True
                    }
        except Exception as e:
            logger.error(f"Agent-SRE K8s error: {e}")
            return {
                "response": "I'm experiencing technical difficulties querying Kubernetes.",
                "error": True
            }
    
    @logfire.instrument("sre_status")
    async def get_status(self) -> Dict:
        """❤️ Get status from Agent-SRE Service"""
        if not self.http_session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        try:
            async with self.http_session.get(
                f"{self.base_url}/health",
                timeout=aiohttp.ClientTimeout(total=5)
            ) as response:
                if response.status == 200:
                    return await response.json()
                else:
                    return {"status": "unavailable"}
        except Exception as e:
            logger.error(f"Status check error: {e}")
            return {"status": "error", "error": str(e)}


# 🧠 LangGraph Agent State
class AgentState(TypedDict):
    """State for the LangGraph agent"""
    messages: Annotated[List, add_messages]
    available_tools: List[Dict]
    tool_results: List[Dict]
    next_action: str
    selected_tool: Optional[str]
    tool_arguments: Optional[Dict]


class JamieLangGraphAgent:
    """🧠 LangGraph-based agent for Jamie with MCP tool discovery and calling"""
    
    def __init__(self, agent_sre_client: AgentSREClient, model_name: str = MODEL_NAME):
        self.agent_client = agent_sre_client
        self.model_name = model_name
        
        # Initialize Ollama LLM
        self.llm = ChatOllama(
            model=self.model_name,
            base_url=OLLAMA_URL,
            temperature=0.0
        )
        
        # Build LangGraph workflow
        self.workflow = self._build_workflow()
        self.app = self.workflow.compile()
        
        logger.info(f"🧠 JamieLangGraphAgent initialized with model: {model_name}")
    
    def _build_workflow(self) -> StateGraph:
        """🏗️ Build the LangGraph workflow"""
        workflow = StateGraph(AgentState)
        
        # Add nodes
        workflow.add_node("discover_tools", self._discover_tools_node)
        workflow.add_node("decide_action", self._decide_action_node)
        workflow.add_node("call_tool", self._call_tool_node)
        workflow.add_node("respond", self._respond_node)
        
        # Define edges
        workflow.set_entry_point("discover_tools")
        workflow.add_edge("discover_tools", "decide_action")
        
        # Conditional routing from decide_action
        workflow.add_conditional_edges(
            "decide_action",
            self._route_decision,
            {
                "call_tool": "call_tool",
                "respond": "respond",
                "end": END
            }
        )
        
        # After calling a tool, go directly to respond (avoid infinite loop)
        workflow.add_edge("call_tool", "respond")
        workflow.add_edge("respond", END)
        
        return workflow
    
    async def _discover_tools_node(self, state: AgentState) -> Dict:
        """📋 Node: Discover available MCP tools"""
        logger.info("📋 Discovering MCP tools...")
        tools = await self.agent_client.list_tools()
        
        # Log if no tools discovered
        if not tools:
            logger.error("❌ No MCP tools discovered - Agent-SRE MCP server not available")
        
        return {
            "available_tools": tools,
            "next_action": "decide"
        }
    
    async def _decide_action_node(self, state: AgentState) -> Dict:
        """🤔 Node: Decide what action to take"""
        messages = state["messages"]
        available_tools = state["available_tools"]
        tool_results = state.get("tool_results", [])
        
        # Get the last user message
        user_message = None
        for msg in reversed(messages):
            if isinstance(msg, HumanMessage):
                user_message = msg.content
                break
        
        if not user_message:
            return {"next_action": "respond"}
        
        # Analyze user message to determine which tool to call
        tool_descriptions = "\n".join([
            f"- {tool['name']}: {tool['description'][:100]}..."
            for tool in available_tools
        ])
        
        prompt = f"""You are an SRE assistant. Analyze the user's request and determine which tool to call.

Available Tools:
{tool_descriptions}

User Request: {user_message}

If you need to call a tool, respond ONLY with:
CALL_TOOL: <tool_name> <arguments_as_json>

Example:
CALL_TOOL: prometheus_query {{"query": "up{{namespace=\\"agent-sre\\"}}"}}

If no tool is needed, respond with:
NO_TOOL_NEEDED
"""
        
        # Enhanced keyword matching for SRE operations
        message_lower = user_message.lower()
        logger.info(f"🤔 Analyzing message: '{user_message}'")
        
        # Prometheus/Metrics queries
        prometheus_keywords = [
            "prometheus", "metrics", "query", "cpu", "memory", "pod", "namespace",
            "container", "rate", "up{", "histogram", "counter", "gauge", "alert",
            "monitoring", "observability", "golden signals", "latency", "traffic",
            "errors", "saturation", "throughput", "response time"
        ]
        
        # Grafana queries
        grafana_keywords = [
            "grafana", "dashboard", "panel", "visualization", "chart", "graph"
        ]
        
        # MCP tools discovery
        mcp_keywords = [
            "mcp", "tools", "list", "discover", "available", "agent-sre"
        ]
        
        # Only proceed with tool calls if we have available tools
        if not available_tools:
            logger.error("❌ No tools available - cannot process tool requests")
            return {"next_action": "respond"}
        
        # Check for Prometheus queries
        matched_prometheus = [kw for kw in prometheus_keywords if kw in message_lower]
        if matched_prometheus:
            logger.info(f"🔍 Detected Prometheus keywords: {matched_prometheus}")
            query = user_message  # Pass the full message as query for now
            return {
                "next_action": "call_tool",
                "selected_tool": "prometheus_query",
                "tool_arguments": {"query": query}
            }
        
        # Check for Grafana queries
        matched_grafana = [kw for kw in grafana_keywords if kw in message_lower]
        if matched_grafana:
            logger.info(f"📊 Detected Grafana keywords: {matched_grafana}")
            return {
                "next_action": "call_tool",
                "selected_tool": "grafana_query",
                "tool_arguments": {"query_type": "search", "query": user_message}
            }
        
        # Check for MCP tools discovery
        matched_mcp = [kw for kw in mcp_keywords if kw in message_lower]
        if matched_mcp:
            logger.info(f"🔧 Detected MCP keywords: {matched_mcp}")
            return {
                "next_action": "call_tool",
                "selected_tool": "list_tools",
                "tool_arguments": {}
            }
        
        logger.info("❌ No matching keywords found, responding directly")
        return {"next_action": "respond"}
    
    async def _call_tool_node(self, state: AgentState) -> Dict:
        """🔧 Node: Call the selected MCP tool"""
        selected_tool = state.get("selected_tool")
        tool_arguments = state.get("tool_arguments", {})
        
        if not selected_tool:
            logger.warning("⚠️  No tool selected, skipping tool call")
            return {}
        
        logger.info(f"🔧 Calling tool: {selected_tool}")
        
        # Handle special cases
        if selected_tool == "list_tools":
            # Call list_tools method directly
            try:
                tools = await self.agent_client.list_tools()
                if tools:
                    result = {
                        "response": f"📋 Available MCP Tools:\n\n" + "\n".join([
                            f"• **{tool['name']}**: {tool['description']}"
                            for tool in tools
                        ]),
                        "error": False
                    }
                else:
                    result = {
                        "response": "⚠️ No MCP tools available - Agent-SRE MCP server is not accessible.",
                        "error": True
                    }
            except Exception as e:
                logger.error(f"❌ Error calling list_tools: {e}")
                result = {
                    "response": f"⚠️ Error connecting to MCP server: {str(e)}",
                    "error": True
                }
        else:
            # Call the MCP tool
            try:
                result = await self.agent_client.call_tool(selected_tool, tool_arguments)
            except Exception as e:
                logger.error(f"❌ Error calling {selected_tool}: {e}")
                result = {
                    "response": f"⚠️ Error calling {selected_tool}: {str(e)}",
                    "error": True
                }
        
        # Store result
        tool_results = state.get("tool_results", [])
        tool_results.append({
            "tool": selected_tool,
            "arguments": tool_arguments,
            "result": result
        })
        
        return {
            "tool_results": tool_results
        }
    
    async def _respond_node(self, state: AgentState) -> Dict:
        """💬 Node: Generate response to user"""
        messages = state["messages"]
        tool_results = state.get("tool_results", [])
        
        # Get the user's question
        user_message = None
        for msg in reversed(messages):
            if isinstance(msg, HumanMessage):
                user_message = msg.content
                break
        
        if tool_results:
            # Format tool results for the response
            results_text = "\n\n".join([
                f"Tool: {r['tool']}\nResult:\n{r['result'].get('response', 'No response')}"
                for r in tool_results
            ])
            
            response_text = f"📊 Here are the results:\n\n{results_text}"
        else:
            # Check if this looks like a tool request but no tools were available
            message_lower = user_message.lower() if user_message else ""
            if any(kw in message_lower for kw in ["prometheus", "grafana", "mcp", "tools", "query", "metrics"]):
                response_text = f"⚠️ I detected a request for SRE tools, but the Agent-SRE MCP server is not available.\n\nPlease check that the Agent-SRE service is running and accessible."
            else:
                response_text = f"🤖 I received your message: {user_message}\n\nI'm still learning how to help with this type of request!"
        
        # Add AI message to state
        ai_message = AIMessage(content=response_text)
        
        return {
            "messages": [ai_message]
        }
    
    def _route_decision(self, state: AgentState) -> str:
        """🔀 Route based on next_action"""
        next_action = state.get("next_action", "respond")
        logger.debug(f"🔀 Routing decision: next_action={next_action}")
        
        if next_action == "call_tool":
            return "call_tool"
        elif next_action == "respond":
            return "respond"
        else:
            return "end"
    
    async def process_message(self, user_message: str) -> str:
        """Process a user message through the LangGraph workflow"""
        logger.info(f"🧠 Processing message with LangGraph: {user_message[:100]}...")
        
        # Initialize state
        initial_state: AgentState = {
            "messages": [HumanMessage(content=user_message)],
            "available_tools": [],
            "tool_results": [],
            "next_action": "discover",
            "selected_tool": None,
            "tool_arguments": None
        }
        
        try:
            # Run the workflow
            final_state = await self.app.ainvoke(initial_state)
            
            # Extract response from final state
            messages = final_state.get("messages", [])
            
            for msg in reversed(messages):
                if isinstance(msg, AIMessage):
                    return msg.content
            
            return "I'm sorry, I couldn't process your request."
        
        except Exception as e:
            logger.error(f"❌ Error in LangGraph workflow: {e}", exc_info=True)
            return f"⚠️ I encountered an error: {str(e)}"


class JamieSlackBot:
    """🤖 Jamie - Your SRE Companion on Slack - forwards to Agent-SRE MCP Server"""
    
    def __init__(self):
        # Initialize Slack app
        self.app = AsyncApp(
            token=os.environ.get("SLACK_BOT_TOKEN"),
            signing_secret=os.environ.get("SLACK_SIGNING_SECRET")
        )
        
        # Bot configuration
        self.bot_name = "Jamie"
        self.bot_emoji = "🤖"
        self.agent_url = AGENT_SRE_URL
        
        # User context storage (in production, use Redis or proper database)
        self.user_contexts: Dict[str, Dict] = {}
        
        # 🌐 Create aiohttp web app for REST API
        self.web_app = web.Application()
        self._setup_rest_api_routes()
        
        # Set up Slack event handlers
        self._setup_handlers()
        
        logger.info(f"🤖 Jamie Slack Bot initialized")
        logger.info(f"   🔧 Agent-SRE MCP: {self.agent_url}")
    
    def _setup_rest_api_routes(self):
        """Setup REST API routes for Jamie API"""
        self.web_app.router.add_post('/api/chat', self.handle_api_chat)
        self.web_app.router.add_post('/api/prometheus/query', self.handle_api_prometheus_query)
        self.web_app.router.add_post('/api/golden-signals', self.handle_api_golden_signals)
        self.web_app.router.add_post('/api/pod-logs', self.handle_api_pod_logs)
        self.web_app.router.add_post('/api/analyze-logs', self.handle_api_analyze_logs)
        self.web_app.router.add_get('/health', self.handle_api_health)
        self.web_app.router.add_get('/ready', self.handle_api_ready)
        self.web_app.router.add_get('/status', self.handle_api_status)
    
    @logfire.instrument("process_message")
    async def _process_message(self, message: str, context: Optional[Dict] = None) -> str:
        """Process message using LangGraph agent with MCP tool discovery"""
        try:
            logger.info(f"🧠 Processing message with LangGraph agent: {message[:100]}...")
            
            # Create agent client and LangGraph agent
            async with AgentSREClient(self.agent_url) as agent_client:
                # Create LangGraph agent
                langgraph_agent = JamieLangGraphAgent(agent_client)
                
                # Process message through LangGraph workflow
                response = await langgraph_agent.process_message(message)
                
                return response
        
        except Exception as e:
            logger.error(f"❌ Error processing message: {e}", exc_info=True)
            return f"⚠️ I encountered an error: {str(e)}"
    
    def _detect_sre_request_type(self, message: str) -> str:
        """Detect what type of SRE request this is and return endpoint type"""
        message_lower = message.lower()
        
        # Prometheus-specific keywords
        prometheus_keywords = [
            "prometheus", "promql", "metrics", "query", "latency", "traffic", 
            "errors", "saturation", "rate", "histogram", "counter", "gauge",
            "up{", "cpu", "memory", "disk", "network", "response_time"
        ]
        
        # Grafana-specific keywords  
        grafana_keywords = [
            "grafana", "dashboard", "panel", "visualization", "chart", "graph"
        ]
        
        # Kubernetes-specific keywords
        k8s_keywords = [
            "kubectl", "pods", "deployments", "logs", "namespace", "cluster",
            "kubernetes", "k8s", "container", "node", "service", "ingress"
        ]
        
        # Check for Prometheus queries
        for keyword in prometheus_keywords:
            if keyword in message_lower:
                return "prometheus"
        
        # Check for Grafana queries
        for keyword in grafana_keywords:
            if keyword in message_lower:
                return "grafana"
        
        # Check for Kubernetes operations
        for keyword in k8s_keywords:
            if keyword in message_lower:
                return "k8s"
        
        # Check for general SRE commands
        sre_keywords = [
            "infrastructure", "monitoring", "observability", "incident",
            "root cause", "performance", "health check", "status",
            "check", "show me", "get", "list", "describe", "analyze",
            "investigate", "debug", "troubleshoot"
        ]
        
        for keyword in sre_keywords:
            if keyword in message_lower:
                return "general"
        
        return "none"
    
    @logfire.instrument("call_agent_sre_endpoint")
    async def _call_agent_sre_endpoint(self, message: str, request_type: str, context: Optional[Dict] = None) -> str:
        """Call specific agent-sre endpoint based on request type"""
        try:
            async with AgentSREClient(self.agent_url) as agent:
                if request_type == "prometheus":
                    logger.info(f"📊 Calling Agent-SRE Prometheus endpoint: {message[:100]}...")
                    result = await agent.prometheus_query(message, context)
                elif request_type == "grafana":
                    logger.info(f"📈 Calling Agent-SRE Grafana endpoint: {message[:100]}...")
                    result = await agent.grafana_query(message, context)
                elif request_type == "k8s":
                    logger.info(f"☸️ Calling Agent-SRE K8s endpoint: {message[:100]}...")
                    result = await agent.k8s_query(message, context)
                else:  # general
                    logger.info(f"🔧 Calling Agent-SRE general chat: {message[:100]}...")
                    result = await agent.chat(message, context)
            
            if result.get("error"):
                return result.get("response", "⚠️ I'm having trouble connecting to the SRE agent.")
            
            return result.get("response", "No response from agent")
        
        except Exception as e:
            logger.error(f"❌ Error calling Agent-SRE {request_type} endpoint: {e}", exc_info=True)
            return f"⚠️ I encountered an error connecting to the SRE agent: {str(e)}"
    
    @logfire.instrument("handle_general_conversation")
    async def _handle_general_conversation(self, message: str, context: Optional[Dict] = None) -> str:
        """Handle general conversation with local LLM (Ollama)"""
        try:
            # TODO: Implement local LLM conversation
            # For now, provide a helpful response indicating Jamie's capabilities
            return f"""🤖 Hi! I'm Jamie, your SRE assistant. 

I can help you with:
• **Infrastructure monitoring** - Ask me to check golden signals, metrics, or service health
• **Kubernetes operations** - Query pods, deployments, logs, and cluster status  
• **Log analysis** - Search logs with Loki/LogQL or analyze error patterns
• **Incident investigation** - Help troubleshoot issues and find root causes
• **General SRE questions** - Ask about best practices, monitoring strategies, etc.

For infrastructure operations, just ask me things like:
• "Check the golden signals for homepage"
• "Show me pods in the production namespace" 
• "Analyze logs from the API service"
• "What's the error rate for our services?"

What would you like to know? 🚀"""
        
        except Exception as e:
            logger.error(f"❌ Error in general conversation: {e}", exc_info=True)
            return f"⚠️ I encountered an error: {str(e)}"
    
    # REST API Handlers
    
    @logfire.instrument("api_chat")
    async def handle_api_chat(self, request):
        """Handle POST /api/chat"""
        try:
            data = await request.json()
            message = data.get("message", "")
            
            if not message:
                return web.json_response(
                    {"error": "Message is required"},
                    status=400
                )
            
            # Forward to Agent-SRE
            response = await self._process_message(message)
            
            return web.json_response({
                "response": response,
                "timestamp": datetime.now().isoformat()
            })
        
        except Exception as e:
            logger.error(f"❌ Error in API chat: {e}")
            return web.json_response({"error": str(e)}, status=500)
    
    @logfire.instrument("api_prometheus_query")
    async def handle_api_prometheus_query(self, request):
        """Handle POST /api/prometheus/query"""
        try:
            data = await request.json()
            query = data.get("query", "")
            time_param = data.get("time")
            
            if not query:
                return web.json_response(
                    {"error": "Query is required"},
                    status=400
                )
            
            # Build context
            context = {}
            if time_param:
                context["time"] = time_param
            
            # Call Agent-SRE Prometheus endpoint
            async with AgentSREClient(self.agent_url) as agent:
                result = await agent.prometheus_query(query, context)
            
            if result.get("error"):
                return web.json_response(
                    {"error": result.get("details", "Failed to query Prometheus")},
                    status=500
                )
            
            return web.json_response({
                "result": result.get("response"),
                "timestamp": datetime.now().isoformat()
            })
        
        except Exception as e:
            logger.error(f"❌ Error in Prometheus query: {e}")
            return web.json_response({"error": str(e)}, status=500)
    
    @logfire.instrument("api_golden_signals")
    async def handle_api_golden_signals(self, request):
        """Handle POST /api/golden-signals"""
        try:
            data = await request.json()
            service = data.get("service", "")
            namespace = data.get("namespace", "default")
            
            if not service:
                return web.json_response(
                    {"error": "Service name is required"},
                    status=400
                )
            
            # Build a message asking for golden signals
            message = f"Check the golden signals for service {service} in namespace {namespace}"
            
            # Process through Jamie's agent
            response = await self._process_message(message)
            
            # Try to parse golden signals from response
            # For now, return the full response
            return web.json_response({
                "service": service,
                "namespace": namespace,
                "latency": "N/A",
                "traffic": "N/A", 
                "errors": "N/A",
                "saturation": "N/A",
                "analysis": response,
                "timestamp": datetime.now().isoformat()
            })
        
        except Exception as e:
            logger.error(f"❌ Error checking golden signals: {e}")
            return web.json_response({"error": str(e)}, status=500)
    
    @logfire.instrument("api_pod_logs")
    async def handle_api_pod_logs(self, request):
        """Handle POST /api/pod-logs"""
        try:
            data = await request.json()
            pod_name = data.get("pod_name", "")
            namespace = data.get("namespace", "default")
            container = data.get("container")
            lines = data.get("lines", 100)
            
            if not pod_name:
                return web.json_response(
                    {"error": "Pod name is required"},
                    status=400
                )
            
            # Build message for Agent-SRE
            message = f"Get logs from pod {pod_name} in namespace {namespace}"
            if container:
                message += f" container {container}"
            message += f" (last {lines} lines)"
            
            # Call Agent-SRE K8s endpoint
            context = {
                "pod_name": pod_name,
                "namespace": namespace,
                "lines": lines
            }
            if container:
                context["container"] = container
            
            async with AgentSREClient(self.agent_url) as agent:
                result = await agent.k8s_query(message, context)
            
            if result.get("error"):
                return web.json_response(
                    {"error": result.get("response", "Failed to get pod logs")},
                    status=500
                )
            
            # Parse logs from response (assuming it's text)
            logs_text = result.get("response", "")
            logs = logs_text.split("\n") if logs_text else []
            
            return web.json_response({
                "pod_name": pod_name,
                "namespace": namespace,
                "container": container,
                "logs": logs,
                "timestamp": datetime.now().isoformat()
            })
        
        except Exception as e:
            logger.error(f"❌ Error getting pod logs: {e}")
            return web.json_response({"error": str(e)}, status=500)
    
    @logfire.instrument("api_analyze_logs")
    async def handle_api_analyze_logs(self, request):
        """Handle POST /api/analyze-logs"""
        try:
            data = await request.json()
            logs = data.get("logs", "")
            context = data.get("context", "")
            
            if not logs:
                return web.json_response(
                    {"error": "Logs are required"},
                    status=400
                )
            
            # Build analysis message
            message = f"Analyze these logs and identify any errors, patterns, or issues:\n\n{logs}"
            if context:
                message = f"Context: {context}\n\n{message}"
            
            # Process through Jamie's AI
            analysis = await self._process_message(message)
            
            return web.json_response({
                "analysis": analysis,
                "timestamp": datetime.now().isoformat()
            })
        
        except Exception as e:
            logger.error(f"❌ Error analyzing logs: {e}")
            return web.json_response({"error": str(e)}, status=500)
    
    @logfire.instrument("api_health")
    async def handle_api_health(self, request):
        """Handle GET /health"""
        return web.json_response({
            "status": "healthy",
            "service": "jamie-slack-bot",
            "timestamp": datetime.now().isoformat(),
            "version": "1.0.0",
            "agent_sre_url": self.agent_url
        })
    
    @logfire.instrument("api_ready")
    async def handle_api_ready(self, request):
        """Handle GET /ready"""
        try:
            async with AgentSREClient(self.agent_url) as agent:
                status = await agent.get_status()
            
            if status.get("status") != "healthy":
                return web.json_response(
                    {"status": "not_ready", "reason": "Agent-SRE unavailable"},
                    status=503
                )
            
            return web.json_response({
                "status": "ready",
                "service": "jamie-slack-bot",
                "timestamp": datetime.now().isoformat()
            })
        
        except Exception as e:
            return web.json_response(
                {"status": "not_ready", "error": str(e)},
                status=503
            )
    
    @logfire.instrument("api_status")
    async def handle_api_status(self, request):
        """Handle GET /status - Detailed status endpoint for Homepage"""
        try:
            # Get Agent-SRE status
            async with AgentSREClient(self.agent_url) as agent:
                agent_status = await agent.get_status()
            
            return web.json_response({
                "status": "healthy",
                "service": "jamie-slack-bot",
                "timestamp": datetime.now().isoformat(),
                "version": "1.0.0",
                "agent_sre_url": self.agent_url,
                "agent_sre_status": agent_status.get("status", "unknown"),
                "components": {
                    "slack_bot": "healthy",
                    "rest_api": "healthy",
                    "agent_sre": agent_status.get("status", "unknown"),
                    "ollama": "available"
                }
            })
        
        except Exception as e:
            logger.error(f"❌ Error in status check: {e}")
            return web.json_response({
                "status": "degraded",
                "service": "jamie-slack-bot",
                "timestamp": datetime.now().isoformat(),
                "error": str(e),
                "components": {
                    "slack_bot": "healthy",
                    "rest_api": "healthy",
                    "agent_sre": "unavailable",
                    "ollama": "unknown"
                }
            }, status=200)  # Return 200 even if degraded, so Homepage knows Jamie is alive
    
    def _setup_handlers(self):
        """Set up Slack event handlers"""
        
        @self.app.event("app_mention")
        async def handle_mention(event, say, client):
            """Handle when Jamie is mentioned"""
            user_id = event["user"]
            channel = event["channel"]
            text = event["text"]
            
            # Extract the actual question (remove bot mention)
            bot_user_id = (await client.auth_test())["user_id"]
            question = text.replace(f"<@{bot_user_id}>", "").strip()
            
            if not question:
                await say(f"{self.bot_emoji} Hey! I'm Jamie, your SRE assistant. How can I help you today?")
                return
            
            # Show typing indicator
            try:
                await client.chat_postMessage(
                    channel=channel,
                    text=f"{self.bot_emoji} _Thinking..._"
                )
            except Exception:
                pass
            
            # Get user context
            context = self._get_user_context(user_id)
            
            # Forward to Agent-SRE MCP Server
            response_text = await self._process_message(question, context)
            
            # Update user context
            self._update_user_context(user_id, question, response_text)
            
            # Format response
            formatted_response = f"{self.bot_emoji} {response_text}"
            
            # Send response
            await say(formatted_response)
        
        @self.app.event("message")
        async def handle_direct_message(event, say, client):
            """Handle direct messages to Jamie"""
            # Only respond to direct messages (not channel messages)
            if event.get("channel_type") != "im":
                return
            
            # Ignore bot messages
            if event.get("bot_id"):
                return
            
            user_id = event["user"]
            text = event.get("text", "")
            
            if not text:
                return
            
            # Get user context
            context = self._get_user_context(user_id)
            
            # Forward to Agent-SRE MCP Server
            response_text = await self._process_message(text, context)
            
            # Update user context
            self._update_user_context(user_id, text, response_text)
            
            # Format response
            formatted_response = f"{self.bot_emoji} {response_text}"
            
            # Send response
            await say(formatted_response)
        
        @self.app.command("/jamie-help")
        async def handle_help_command(ack, respond):
            """Handle /jamie-help slash command"""
            await ack()
            
            help_text = """
🤖 *Jamie - Your SRE Assistant*

*How to use me:*
• Mention me in any channel: `@Jamie check the golden signals`
• Send me a direct message
• Use slash commands for quick actions

*Quick Commands:*
• `/jamie-help` - Show this help message
• `/jamie-status` - Check Agent-SRE status

*What I can help with:*
• � Golden Signals monitoring (latency, traffic, errors, saturation)
• ☸️ Kubernetes operations (pods, deployments, logs, status)
• �📈 Prometheus queries and metrics
• 📝 Loki log queries (LogQL)
• 🔍 Tempo trace queries (distributed tracing)
• 🔎 Log analysis and error pattern detection
• 🚨 Incident investigation and root cause analysis
• 📉 Performance metrics and optimization
• 🎯 Service health monitoring

*Example questions:*
• "Check the golden signals for homepage"
• "What's the error rate for the API?"
• "Show me logs from pod homepage-xyz"
• "Query Loki: {namespace=\"production\"} |= \"error\""
• "Look up trace 1234567890abcdef in Tempo"
• "Analyze these logs for errors"
• "Query Prometheus: up{job=\"homepage\"}"

Just ask me anything about your infrastructure! 🚀
            """
            
            await respond(help_text)
        
        @self.app.command("/jamie-status")
        async def handle_status_command(ack, respond):
            """Handle /jamie-status slash command"""
            await ack()
            
            async with AgentSREClient(self.agent_url) as agent:
                status = await agent.get_status()
            
            if status.get("status") == "error" or status.get("status") == "unavailable":
                await respond(f"{self.bot_emoji} ⚠️ Agent-SRE is currently unavailable. Please try again later.")
            else:
                status_text = f"""
{self.bot_emoji} *Agent-SRE Status*

✅ Status: {status.get('status', 'unknown')}
🕐 Timestamp: {status.get('timestamp', 'N/A')}
🔧 Service: {status.get('service', 'agent-sre')}

_All systems operational!_
                """
                await respond(status_text)
    
    def _get_user_context(self, user_id: str) -> Optional[Dict]:
        """Get user's conversation context"""
        context_data = self.user_contexts.get(user_id, {})
        if not context_data:
            return None
        
        # Return recent conversation history (last 5 exchanges)
        history = context_data.get("history", [])
        if len(history) > 10:  # Keep last 5 Q&A pairs
            history = history[-10:]
        
        return {
            "conversation_history": history,
            "last_updated": context_data.get("last_updated")
        }
    
    def _update_user_context(self, user_id: str, question: str, response: str):
        """Update user's conversation context"""
        if user_id not in self.user_contexts:
            self.user_contexts[user_id] = {
                "history": [],
                "last_updated": datetime.now(UTC).isoformat()
            }
        
        # Add to history
        self.user_contexts[user_id]["history"].extend([
            {"role": "user", "content": question},
            {"role": "assistant", "content": response}
        ])
        
        # Keep only recent history (last 20 exchanges = 10 Q&A pairs)
        if len(self.user_contexts[user_id]["history"]) > 20:
            self.user_contexts[user_id]["history"] = self.user_contexts[user_id]["history"][-20:]
        
        self.user_contexts[user_id]["last_updated"] = datetime.now(UTC).isoformat()
    
    async def run_slack_bot(self):
        """Start the Slack bot"""
        try:
            # Get app-level token for Socket Mode
            app_token = os.environ.get("SLACK_APP_TOKEN")
            if not app_token:
                raise ValueError("SLACK_APP_TOKEN environment variable is required")
            
            # Start the bot
            handler = AsyncSocketModeHandler(self.app, app_token)
            logger.info("🤖 Starting Jamie Slack Bot...")
            await handler.start_async()
            
        except Exception as e:
            logger.error(f"Failed to start bot: {e}")
            raise
    
    async def run_rest_api(self, host: str = "0.0.0.0", port: int = 8080):
        """Start the REST API server"""
        runner = web.AppRunner(self.web_app)
        await runner.setup()
        site = web.TCPSite(runner, host, port)
        await site.start()
        
        logger.info("=" * 60)
        logger.info("🌐 Jamie REST API Started!")
        logger.info("=" * 60)
        logger.info(f"🌐 Server: http://{host}:{port}")
        logger.info(f"💬 Chat: http://localhost:{port}/api/chat")
        logger.info(f"🏥 Health: http://localhost:{port}/health")
        logger.info(f"✅ Ready: http://localhost:{port}/ready")
        logger.info(f"🔗 Agent-SRE: {self.agent_url}")
        logger.info("=" * 60)
        
        return runner


async def main():
    """Main function - runs both Slack bot and REST API"""
    # Check required environment variables
    required_vars = ["SLACK_BOT_TOKEN", "SLACK_SIGNING_SECRET", "SLACK_APP_TOKEN"]
    missing_vars = [var for var in required_vars if not os.environ.get(var)]
    
    if missing_vars:
        logger.error(f"Missing required environment variables: {missing_vars}")
        logger.error("Please set the following environment variables:")
        for var in missing_vars:
            logger.error(f"  export {var}=your_value_here")
        return
    
    # Create Jamie bot
    bot = JamieSlackBot()
    
    # Start REST API server
    api_host = os.getenv("API_HOST", "0.0.0.0")
    api_port = int(os.getenv("API_PORT", "8080"))
    api_runner = await bot.run_rest_api(api_host, api_port)
    
    # Start Slack bot
    await bot.run_slack_bot()


if __name__ == "__main__":
    asyncio.run(main())
