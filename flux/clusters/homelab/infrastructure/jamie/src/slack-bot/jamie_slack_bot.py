#!/usr/bin/env python3
"""
🤖 Jamie Slack Bot
A sophisticated SRE assistant with LLM brain and REST API
Calls Agent-SRE Agent Service REST API (port 8080) for tool execution
"""

import os
import json
import asyncio
from typing import Dict, List, Optional, Any
from datetime import datetime

import aiohttp
from aiohttp import web
from slack_bolt.async_app import AsyncApp
from slack_bolt.adapter.socket_mode.async_handler import AsyncSocketModeHandler
from slack_sdk.errors import SlackApiError

# LangChain imports
from langchain_ollama import ChatOllama
from langchain_core.messages import HumanMessage, SystemMessage, AIMessage, ToolMessage
from langchain_core.tools import tool
from langchain.agents import AgentExecutor, create_tool_calling_agent
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder

# Import from core
from core import logger, logfire, OLLAMA_URL, MODEL_NAME, SERVICE_NAME, AGENT_SRE_URL

# Initialize Ollama LLM
try:
    llm = ChatOllama(
        model=MODEL_NAME,
        base_url=OLLAMA_URL,
        temperature=0.7,
        num_ctx=8192,
    )
    logger.info(f"✅ Ollama connection established: {OLLAMA_URL} using model {MODEL_NAME}")
except Exception as e:
    logger.error(f"❌ Error connecting to Ollama: {e}")
    llm = None


class AgentSREClient:
    """Client for interacting with Agent-SRE Agent Service REST API"""
    
    def __init__(self, base_url: str = None):
        # Agent-SRE Agent Service REST API (port 8080)
        self.base_url = base_url or AGENT_SRE_URL
        self.session = None
    
    async def __aenter__(self):
        self.session = aiohttp.ClientSession()
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.session:
            await self.session.close()
    
    @logfire.instrument("agent_sre_chat")
    async def chat(self, message: str, context: Optional[Dict] = None) -> Dict:
        """Send chat message to Agent-SRE REST API"""
        if not self.session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        payload = {
            "message": message,
            "timestamp": datetime.utcnow().isoformat(),
            "context": context or {}
        }
        
        try:
            async with self.session.post(
                f"{self.base_url}/chat",
                json=payload,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                if response.status == 200:
                    result = await response.json()
                    return {"response": result.get("response", "No response")}
                else:
                    error_text = await response.text()
                    logger.error(f"Agent-SRE API error: {response.status} - {error_text}")
                    return {
                        "response": "I'm having trouble connecting to the SRE agent.",
                        "error": True
                    }
        
        except Exception as e:
            logger.error(f"Agent-SRE API error: {e}")
            return {
                "response": "I'm experiencing technical difficulties connecting to the SRE agent.",
                "error": True
            }
    
    @logfire.instrument("agent_sre_golden_signals")
    async def check_golden_signals(self, service_name: str, namespace: str = "default") -> Dict:
        """Check golden signals via REST API"""
        if not self.session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        payload = {
            "service_name": service_name,
            "namespace": namespace
        }
        
        try:
            async with self.session.post(
                f"{self.base_url}/golden-signals",
                json=payload,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                if response.status == 200:
                    return await response.json()
                else:
                    return {"error": f"HTTP {response.status}"}
        except Exception as e:
            logger.error(f"Golden signals error: {e}")
            return {"error": str(e)}
    
    @logfire.instrument("agent_sre_prometheus")
    async def query_prometheus(self, query: str) -> Dict:
        """Query Prometheus via REST API"""
        if not self.session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        payload = {"query": query}
        
        try:
            async with self.session.post(
                f"{self.base_url}/prometheus/query",
                json=payload,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                if response.status == 200:
                    return await response.json()
                else:
                    return {"error": f"HTTP {response.status}"}
        except Exception as e:
            logger.error(f"Prometheus query error: {e}")
            return {"error": str(e)}
    
    @logfire.instrument("agent_sre_pod_logs")
    async def get_pod_logs(self, pod_name: str, namespace: str = "default", tail_lines: int = 100) -> Dict:
        """Get pod logs via REST API"""
        if not self.session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        payload = {
            "pod_name": pod_name,
            "namespace": namespace,
            "tail_lines": tail_lines
        }
        
        try:
            async with self.session.post(
                f"{self.base_url}/kubernetes/logs",
                json=payload,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                if response.status == 200:
                    return await response.json()
                else:
                    return {"error": f"HTTP {response.status}"}
        except Exception as e:
            logger.error(f"Pod logs error: {e}")
            return {"error": str(e)}
    
    @logfire.instrument("agent_sre_analyze_logs")
    async def analyze_logs(self, logs: str, context: Optional[str] = None) -> Dict:
        """Analyze logs via REST API"""
        if not self.session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        payload = {
            "logs": logs,
            "context": context or "",
            "timestamp": datetime.utcnow().isoformat()
        }
        
        try:
            async with self.session.post(
                f"{self.base_url}/analyze-logs",
                json=payload,
                timeout=aiohttp.ClientTimeout(total=45)
            ) as response:
                if response.status == 200:
                    result = await response.json()
                    return {"analysis": result.get("analysis", "No analysis")}
                else:
                    return {"analysis": "Unable to analyze logs.", "error": True}
        except Exception as e:
            logger.error(f"Log analysis error: {e}")
            return {"analysis": "Error analyzing logs.", "error": True}
    
    @logfire.instrument("agent_sre_status")
    async def get_status(self) -> Dict:
        """Get Agent-SRE status"""
        if not self.session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        try:
            async with self.session.get(
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


class JamieSlackBot:
    """🤖 Jamie - Your SRE Companion on Slack with LLM Brain and REST API"""
    
    def __init__(self):
        # Initialize Slack app
        self.app = AsyncApp(
            token=os.environ.get("SLACK_BOT_TOKEN"),
            signing_secret=os.environ.get("SLACK_SIGNING_SECRET")
        )
        
        # Bot configuration
        self.bot_name = "Jamie"
        self.bot_emoji = "🤖"
        self.llm = llm  # LLM brain
        self.agent_url = AGENT_SRE_URL
        
        # User context storage (in production, use Redis or proper database)
        self.user_contexts: Dict[str, Dict] = {}
        
        # 🔧 Create LangChain tools that call Agent-SRE REST API
        self.tools = self._create_tools()
        
        # 🤖 Create LangChain agent with tools
        self.agent_executor = self._create_agent()
        
        # 🌐 Create aiohttp web app for REST API
        self.web_app = web.Application()
        self._setup_rest_api_routes()
        
        # Set up Slack event handlers
        self._setup_handlers()
        
        logger.info(f"🤖 Jamie Slack Bot initialized")
        logger.info(f"   🧠 LLM: {MODEL_NAME} @ {OLLAMA_URL}")
        logger.info(f"   🔧 Agent-SRE: {self.agent_url}")
        logger.info(f"   🛠️  Loaded {len(self.tools)} tools")
    
    def _setup_rest_api_routes(self):
        """Setup REST API routes for Jamie API"""
        self.web_app.router.add_post('/api/chat', self.handle_api_chat)
        self.web_app.router.add_post('/api/golden-signals', self.handle_api_golden_signals)
        self.web_app.router.add_post('/api/prometheus/query', self.handle_api_prometheus_query)
        self.web_app.router.add_post('/api/pod-logs', self.handle_api_pod_logs)
        self.web_app.router.add_post('/api/analyze-logs', self.handle_api_analyze_logs)
        self.web_app.router.add_get('/health', self.handle_api_health)
        self.web_app.router.add_get('/ready', self.handle_api_ready)
    
    def _create_tools(self) -> List:
        """🔧 Create LangChain tools that call Agent-SRE REST API (port 8080)"""
        
        agent_url = self.agent_url
        
        @tool
        async def check_golden_signals(service_name: str, namespace: str = "default") -> str:
            """Check golden signals (latency, traffic, errors, saturation) for a service.
            Use this when users ask about service health, status, or golden signals.
            
            Args:
                service_name: The name of the service to check (e.g., 'homepage', 'api', 'frontend')
                namespace: The Kubernetes namespace (default: 'default')
            
            Returns:
                JSON string with golden signals data
            """
            async with AgentSREClient(agent_url) as agent:
                result = await agent.check_golden_signals(service_name, namespace)
                return json.dumps(result, indent=2)
        
        @tool
        async def query_prometheus(query: str) -> str:
            """Execute a PromQL query against Prometheus.
            Use this when users want to query metrics or run custom PromQL queries.
            
            Args:
                query: The PromQL query to execute (e.g., 'up', 'rate(http_requests_total[5m])')
            
            Returns:
                JSON string with query results from Prometheus
            """
            async with AgentSREClient(agent_url) as agent:
                result = await agent.query_prometheus(query)
                return json.dumps(result, indent=2)
        
        @tool
        async def get_pod_logs(pod_name: str, namespace: str = "default", tail_lines: int = 100) -> str:
            """Get logs from a Kubernetes pod.
            Use this when users ask for pod logs or want to see what's happening in a pod.
            
            Args:
                pod_name: The name of the pod
                namespace: The Kubernetes namespace (default: 'default')
                tail_lines: Number of log lines to return (default: 100)
            
            Returns:
                Pod logs as a string
            """
            async with AgentSREClient(agent_url) as agent:
                result = await agent.get_pod_logs(pod_name, namespace, tail_lines)
                return result.get("logs", "No logs")
        
        @tool
        async def sre_chat(message: str) -> str:
            """General SRE consultation and advice using AI.
            Use this for general SRE questions, best practices, or when no specific tool applies.
            
            Args:
                message: The SRE question or request
            
            Returns:
                AI-generated response with SRE insights
            """
            async with AgentSREClient(agent_url) as agent:
                result = await agent.chat(message)
                return result.get("response", "No response")
        
        @tool
        async def analyze_logs(logs: str, context: str = None) -> str:
            """Analyze logs for errors, patterns, and insights.
            Use this when users provide logs and want them analyzed.
            
            Args:
                logs: The log data to analyze
                context: Optional context about the logs
            
            Returns:
                Analysis of the logs with insights and recommendations
            """
            async with AgentSREClient(agent_url) as agent:
                result = await agent.analyze_logs(logs, context)
                return result.get("analysis", "No analysis available")
        
        @tool
        async def get_agent_status() -> str:
            """Check the status of the SRE agent service.
            Use this when users ask about the agent's health or availability.
            
            Returns:
                JSON string with agent status information
            """
            async with AgentSREClient(agent_url) as agent:
                status = await agent.get_status()
                return json.dumps(status, indent=2)
        
        return [
            check_golden_signals,
            query_prometheus,
            get_pod_logs,
            sre_chat,
            analyze_logs,
            get_agent_status
        ]
    
    def _create_agent(self) -> AgentExecutor:
        """🤖 Create LangChain agent with tools"""
        if not self.llm:
            logger.error("Cannot create agent: LLM not initialized")
            return None
        
        # Create prompt template for the agent
        prompt = ChatPromptTemplate.from_messages([
            ("system", """You are Jamie, an expert SRE (Site Reliability Engineering) assistant.
You help with monitoring, troubleshooting, incident response, and maintaining system reliability.
You provide clear, actionable insights based on SRE best practices and observability principles.

Your expertise includes:
- 📊 Golden Signals monitoring (latency, traffic, errors, saturation)
- ☸️ Kubernetes operations and troubleshooting
- 📈 Grafana dashboards and alerts
- 🔍 Log analysis and error pattern detection
- 🚨 Incident investigation and root cause analysis
- 📉 Performance metrics and optimization
- 🎯 Service health monitoring

You are friendly, helpful, and concise. Use emojis to make responses engaging.

You have access to powerful SRE tools. Use them when needed:
- Use check_golden_signals when users ask about service health or status
- Use query_prometheus for custom metric queries
- Use get_pod_logs to retrieve pod logs
- Use sre_chat for general SRE advice
- Use analyze_logs when logs are provided
- Use get_agent_status to check the agent's health

Always choose the most appropriate tool based on the user's request."""),
            MessagesPlaceholder(variable_name="chat_history", optional=True),
            ("human", "{input}"),
            MessagesPlaceholder(variable_name="agent_scratchpad"),
        ])
        
        # Create the agent
        agent = create_tool_calling_agent(self.llm, self.tools, prompt)
        
        # Create agent executor
        agent_executor = AgentExecutor(
            agent=agent,
            tools=self.tools,
            verbose=True,
            handle_parsing_errors=True,
            max_iterations=5,
            return_intermediate_steps=False
        )
        
        return agent_executor
    
    @logfire.instrument("process_with_brain")
    async def _process_with_brain(self, message: str, context: Optional[Dict] = None) -> str:
        """🧠 Process message using Jamie's LLM brain with LangChain agent"""
        if not self.agent_executor:
            return "⚠️ My brain is temporarily offline. Please try again later."
        
        try:
            # Build chat history for context
            chat_history = []
            if context and context.get("conversation_history"):
                for msg in context["conversation_history"]:
                    if msg["role"] == "user":
                        chat_history.append(HumanMessage(content=msg["content"]))
                    elif msg["role"] == "assistant":
                        chat_history.append(AIMessage(content=msg["content"]))
            
            # Invoke the agent with the message
            logger.info(f"🧠 Processing with agent: {message[:100]}...")
            
            result = await self.agent_executor.ainvoke({
                "input": message,
                "chat_history": chat_history
            })
            
            # Extract the output
            output = result.get("output", "I'm not sure how to respond to that.")
            
            return output
        
        except Exception as e:
            logger.error(f"❌ Error processing with agent: {e}", exc_info=True)
            return f"⚠️ I encountered an error processing that request: {str(e)}"
    
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
            
            # Process with Jamie's brain
            response = await self._process_with_brain(message)
            
            return web.json_response({
                "response": response,
                "timestamp": datetime.now().isoformat()
            })
        
        except Exception as e:
            logger.error(f"❌ Error in API chat: {e}")
            return web.json_response({"error": str(e)}, status=500)
    
    @logfire.instrument("api_golden_signals")
    async def handle_api_golden_signals(self, request):
        """Handle POST /api/golden-signals"""
        try:
            data = await request.json()
            service_name = data.get("service_name", "")
            namespace = data.get("namespace", "default")
            
            if not service_name:
                return web.json_response(
                    {"error": "service_name is required"},
                    status=400
                )
            
            async with AgentSREClient(self.agent_url) as agent:
                result = await agent.check_golden_signals(service_name, namespace)
            
            return web.json_response(result)
        
        except Exception as e:
            logger.error(f"❌ Error in golden signals: {e}")
            return web.json_response({"error": str(e)}, status=500)
    
    @logfire.instrument("api_prometheus_query")
    async def handle_api_prometheus_query(self, request):
        """Handle POST /api/prometheus/query"""
        try:
            data = await request.json()
            query = data.get("query", "")
            
            if not query:
                return web.json_response(
                    {"error": "query is required"},
                    status=400
                )
            
            async with AgentSREClient(self.agent_url) as agent:
                result = await agent.query_prometheus(query)
            
            return web.json_response(result)
        
        except Exception as e:
            logger.error(f"❌ Error in prometheus query: {e}")
            return web.json_response({"error": str(e)}, status=500)
    
    @logfire.instrument("api_pod_logs")
    async def handle_api_pod_logs(self, request):
        """Handle POST /api/pod-logs"""
        try:
            data = await request.json()
            pod_name = data.get("pod_name", "")
            namespace = data.get("namespace", "default")
            tail_lines = data.get("tail_lines", 100)
            
            if not pod_name:
                return web.json_response(
                    {"error": "pod_name is required"},
                    status=400
                )
            
            async with AgentSREClient(self.agent_url) as agent:
                result = await agent.get_pod_logs(pod_name, namespace, tail_lines)
            
            return web.json_response(result)
        
        except Exception as e:
            logger.error(f"❌ Error in pod logs: {e}")
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
                    {"error": "logs are required"},
                    status=400
                )
            
            async with AgentSREClient(self.agent_url) as agent:
                result = await agent.analyze_logs(logs, context)
            
            return web.json_response(result)
        
        except Exception as e:
            logger.error(f"❌ Error in analyze logs: {e}")
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
            
            # Process with Jamie's brain (LLM)
            response_text = await self._process_with_brain(question, context)
            
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
            
            # Process with Jamie's brain (LLM)
            response_text = await self._process_with_brain(text, context)
            
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
• 📊 Golden Signals monitoring (latency, traffic, errors, saturation)
• ☸️ Kubernetes operations (pods, deployments, logs, status)
• 📈 Prometheus queries and metrics
• 🔍 Log analysis and error pattern detection
• 🚨 Incident investigation and root cause analysis
• 📉 Performance metrics and optimization
• 🎯 Service health monitoring

*Example questions:*
• "Check the golden signals for homepage"
• "What's the error rate for the API?"
• "Show me logs from pod homepage-xyz"
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
                "last_updated": datetime.utcnow().isoformat()
            }
        
        # Add to history
        self.user_contexts[user_id]["history"].extend([
            {"role": "user", "content": question},
            {"role": "assistant", "content": response}
        ])
        
        # Keep only recent history (last 20 exchanges = 10 Q&A pairs)
        if len(self.user_contexts[user_id]["history"]) > 20:
            self.user_contexts[user_id]["history"] = self.user_contexts[user_id]["history"][-20:]
        
        self.user_contexts[user_id]["last_updated"] = datetime.utcnow().isoformat()
    
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
        logger.info(f"📊 Golden Signals: http://localhost:{port}/api/golden-signals")
        logger.info(f"📈 Prometheus: http://localhost:{port}/api/prometheus/query")
        logger.info(f"📝 Pod Logs: http://localhost:{port}/api/pod-logs")
        logger.info(f"🔍 Analyze Logs: http://localhost:{port}/api/analyze-logs")
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
