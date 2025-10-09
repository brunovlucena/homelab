#!/usr/bin/env python3
"""
🤖 Jamie Slack Bot
A sophisticated SRE assistant with LLM brain that communicates via MCP (Model Context Protocol)
Connects to Ollama for intelligence and agent-sre service for tool execution
"""

import os
import json
import asyncio
import logging
from typing import Dict, List, Optional
from datetime import datetime

import aiohttp
from slack_bolt.async_app import AsyncApp
from slack_bolt.adapter.socket_mode.async_handler import AsyncSocketModeHandler
from slack_sdk.errors import SlackApiError

# LangChain imports
from langchain_ollama import ChatOllama
from langchain_core.messages import HumanMessage, SystemMessage, AIMessage

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Configuration
OLLAMA_URL = os.environ.get("OLLAMA_URL", "http://192.168.0.16:11434")
MODEL_NAME = os.environ.get("MODEL_NAME", "bruno-sre:latest")
SERVICE_NAME = "jamie-slack-bot"

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
    """Client for interacting with Agent-SRE via MCP"""
    
    def __init__(self, base_url: str = "http://sre-agent-mcp-server-service.agent-sre:30120"):
        self.base_url = base_url
        self.mcp_url = f"{base_url}/mcp"
        self.health_url = f"{base_url}/health"
        self.session = None
    
    async def __aenter__(self):
        self.session = aiohttp.ClientSession()
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.session:
            await self.session.close()
    
    async def chat(self, message: str, context: Optional[Dict] = None) -> Dict:
        """Send chat message via MCP to Agent-SRE"""
        if not self.session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        # MCP protocol payload - use tools/call method
        payload = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": "sre_chat",
                "arguments": {
                    "message": message,
                    "timestamp": datetime.utcnow().isoformat(),
                    "context": context or {}
                }
            }
        }
        
        try:
            async with self.session.post(
                self.mcp_url,
                json=payload,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                if response.status == 200:
                    result = await response.json()
                    # Extract text from MCP response format
                    if "result" in result and "content" in result["result"]:
                        content = result["result"]["content"]
                        if content and len(content) > 0 and "text" in content[0]:
                            return {"response": content[0]["text"]}
                    return {"response": "No response from SRE agent"}
                else:
                    error_text = await response.text()
                    logger.error(f"Agent-SRE API error: {response.status} - {error_text}")
                    return {
                        "response": "I'm having trouble connecting to the SRE agent. The service might be temporarily unavailable.",
                        "error": True
                    }
        
        except asyncio.TimeoutError:
            logger.error("Agent-SRE API timeout")
            return {
                "response": "The SRE agent is taking longer than usual to respond. Please try again in a moment.",
                "error": True
            }
        except Exception as e:
            logger.error(f"Agent-SRE API error: {e}")
            return {
                "response": "I'm experiencing technical difficulties connecting to the SRE agent. Please try again later.",
                "error": True
            }
    
    async def analyze_logs(self, logs: str, context: Optional[str] = None) -> Dict:
        """Analyze logs via MCP"""
        if not self.session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        payload = {
            "logs": logs,
            "context": context,
            "timestamp": datetime.utcnow().isoformat()
        }
        
        # MCP protocol payload - use tools/call method
        mcp_payload = {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/call",
            "params": {
                "name": "analyze_logs",
                "arguments": payload
            }
        }
        
        try:
            async with self.session.post(
                self.mcp_url,
                json=mcp_payload,
                timeout=aiohttp.ClientTimeout(total=45)
            ) as response:
                if response.status == 200:
                    result = await response.json()
                    # Extract text from MCP response format
                    if "result" in result and "content" in result["result"]:
                        content = result["result"]["content"]
                        if content and len(content) > 0 and "text" in content[0]:
                            return {"analysis": content[0]["text"]}
                    return {"analysis": "No analysis from SRE agent"}
                else:
                    logger.error(f"Log analysis API error: {response.status}")
                    return {
                        "analysis": "Unable to analyze logs at this time.",
                        "error": True
                    }
        except Exception as e:
            logger.error(f"Log analysis error: {e}")
            return {
                "analysis": "Error analyzing logs.",
                "error": True
            }
    
    async def get_status(self) -> Dict:
        """Get Agent-SRE status"""
        if not self.session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        try:
            async with self.session.get(
                self.health_url,
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
    """🤖 Jamie - Your SRE Companion on Slack with LLM Brain"""
    
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
        self.agent_url = os.environ.get("AGENT_SRE_URL", "http://sre-agent-mcp-server-service.agent-sre:30120")
        
        # System prompt for Jamie
        self.system_prompt = """You are Jamie, an expert SRE (Site Reliability Engineering) assistant.
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
When you need to execute infrastructure tasks, you have access to agent-sre MCP tools."""
        
        # User context storage (in production, use Redis or proper database)
        self.user_contexts: Dict[str, Dict] = {}
        
        # Set up event handlers
        self._setup_handlers()
        
        logger.info(f"🤖 Jamie Slack Bot initialized")
        logger.info(f"   🧠 LLM: {MODEL_NAME} @ {OLLAMA_URL}")
        logger.info(f"   🔧 Agent-SRE: {self.agent_url}")
    
    async def _process_with_brain(self, message: str, context: Optional[Dict] = None) -> str:
        """🧠 Process message using Jamie's LLM brain with MCP tool calling"""
        if not self.llm:
            return "⚠️ My brain is temporarily offline. Please try again later."
        
        try:
            # 🔍 First, check if message requires MCP tool execution
            tool_result = await self._detect_and_execute_tool(message)
            
            if tool_result:
                # Tool was executed, now ask LLM to format the result
                format_prompt = f"""The user asked: "{message}"

I executed a tool and got this result:
{tool_result}

Please provide a friendly, concise response to the user based on this data. Use emojis and be conversational."""
                
                messages = [
                    SystemMessage(content=self.system_prompt),
                    HumanMessage(content=format_prompt)
                ]
                
                response = await self.llm.ainvoke(messages)
                return response.content
            
            # No tool needed, just regular LLM response
            messages = [SystemMessage(content=self.system_prompt)]
            
            # Add conversation history if available
            if context and context.get("conversation_history"):
                for msg in context["conversation_history"]:
                    if msg["role"] == "user":
                        messages.append(HumanMessage(content=msg["content"]))
                    elif msg["role"] == "assistant":
                        messages.append(AIMessage(content=msg["content"]))
            
            # Add current message
            messages.append(HumanMessage(content=message))
            
            # Invoke LLM
            logger.info(f"🧠 Processing with LLM: {message[:100]}...")
            response = await self.llm.ainvoke(messages)
            
            return response.content
        
        except Exception as e:
            logger.error(f"❌ Error processing with LLM: {e}")
            return f"⚠️ I encountered an error processing that request: {str(e)}"
    
    async def _detect_and_execute_tool(self, message: str) -> Optional[str]:
        """🔍 Detect if message requires MCP tool and execute it"""
        message_lower = message.lower()
        
        try:
            # 📊 Golden Signals Detection
            if any(keyword in message_lower for keyword in ["golden signal", "golden signals", "check signals", "service health", "service status"]):
                # Extract service name
                service_name = self._extract_service_name(message)
                if service_name:
                    logger.info(f"📊 Detected golden signals request for: {service_name}")
                    async with AgentSREClient(self.agent_url) as agent:
                        result = await self._call_mcp_tool(agent, "check_golden_signals", {
                            "service_name": service_name,
                            "namespace": "default"
                        })
                        return result
            
            # 🔍 Prometheus Query Detection
            if "query prometheus" in message_lower or "promql" in message_lower:
                # Extract query (simple heuristic - text after "query:" or in quotes)
                import re
                query_match = re.search(r'query[:\s]+(.+)', message, re.IGNORECASE)
                if query_match:
                    query = query_match.group(1).strip()
                    logger.info(f"🔍 Detected Prometheus query: {query}")
                    async with AgentSREClient(self.agent_url) as agent:
                        result = await self._call_mcp_tool(agent, "query_prometheus", {"query": query})
                        return result
            
            # 📜 Pod Logs Detection
            if "pod log" in message_lower or "logs for pod" in message_lower or "get logs" in message_lower:
                # Extract pod name
                pod_name = self._extract_pod_name(message)
                if pod_name:
                    logger.info(f"📜 Detected pod logs request for: {pod_name}")
                    async with AgentSREClient(self.agent_url) as agent:
                        result = await self._call_mcp_tool(agent, "get_pod_logs", {
                            "pod_name": pod_name,
                            "namespace": "default"
                        })
                        return result
            
            # No tool detected
            return None
        
        except Exception as e:
            logger.error(f"❌ Error detecting/executing tool: {e}")
            return None
    
    async def _call_mcp_tool(self, agent: 'AgentSREClient', tool_name: str, arguments: Dict) -> str:
        """Call MCP tool via agent-sre"""
        try:
            payload = {
                "jsonrpc": "2.0",
                "id": 1,
                "method": "tools/call",
                "params": {
                    "name": tool_name,
                    "arguments": arguments
                }
            }
            
            async with agent.session.post(
                agent.mcp_url,
                json=payload,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                if response.status == 200:
                    result = await response.json()
                    if "result" in result and "content" in result["result"]:
                        content = result["result"]["content"]
                        if content and len(content) > 0 and "text" in content[0]:
                            return content[0]["text"]
                    return "No response from agent"
                else:
                    error_text = await response.text()
                    return f"Error: {error_text}"
        except Exception as e:
            logger.error(f"❌ Error calling MCP tool {tool_name}: {e}")
            return f"Error: {str(e)}"
    
    def _extract_service_name(self, message: str) -> Optional[str]:
        """Extract service name from message"""
        import re
        # Look for patterns like "for <service>" or "of <service>" or "@<service>"
        patterns = [
            r'for\s+(\w+[-\w]*)',
            r'of\s+(\w+[-\w]*)',
            r'@(\w+[-\w]*)',
            r'service[:\s]+(\w+[-\w]*)'
        ]
        
        for pattern in patterns:
            match = re.search(pattern, message, re.IGNORECASE)
            if match:
                return match.group(1)
        
        # Default to common services
        common_services = ["homepage", "api", "frontend", "backend", "web"]
        for service in common_services:
            if service in message.lower():
                return service
        
        return None
    
    def _extract_pod_name(self, message: str) -> Optional[str]:
        """Extract pod name from message"""
        import re
        # Look for patterns like "pod <name>" or "pod: <name>"
        patterns = [
            r'pod[:\s]+([a-z0-9-]+)',
            r'for pod ([a-z0-9-]+)',
            r'from ([a-z0-9-]+)'
        ]
        
        for pattern in patterns:
            match = re.search(pattern, message, re.IGNORECASE)
            if match:
                return match.group(1)
        
        return None
    
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
• `/jamie-analyze-logs` - Analyze logs from a file or paste

*What I can help with:*
• 📊 Golden Signals monitoring (latency, traffic, errors, saturation)
• ☸️ Kubernetes operations (pods, deployments, logs, status)
• 📈 Grafana operations (dashboards, incidents, alerts)
• 🔍 Log analysis and error pattern detection
• 🚨 Incident investigation and root cause analysis
• 📉 Performance metrics and optimization
• 🎯 Service health monitoring

*Example questions:*
• "Check the golden signals for bruno site"
• "What's the error rate for the API?"
• "List all pods in the default namespace"
• "Show me the dashboard for homepage service"
• "Analyze these logs for errors"
• "What alerts are currently firing?"
• "Investigate high latency in the API"

*Powered by:*
• 🧠 Agent-SRE with MCP (Model Context Protocol)
• 🔧 Dynamic tool discovery and execution
• 📡 Real-time infrastructure monitoring

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
        
        @self.app.command("/jamie-analyze-logs")
        async def handle_analyze_logs_command(ack, respond, command):
            """Handle /jamie-analyze-logs slash command"""
            await ack()
            
            logs = command.get("text", "").strip()
            if not logs:
                await respond("Please provide logs to analyze. Example: `/jamie-analyze-logs ERROR: Connection failed`")
                return
            
            # Analyze logs via MCP
            async with AgentSREClient(self.agent_url) as agent:
                analysis = await agent.analyze_logs(logs)
            
            analysis_text = analysis.get("analysis", "Unable to analyze logs.")
            severity = analysis.get("severity", "unknown")
            recommendations = analysis.get("recommendations", [])
            
            response_text = f"""
{self.bot_emoji} *Log Analysis Results*

📝 *Analysis:*
{analysis_text}

⚠️ *Severity:* {severity}
"""
            
            if recommendations:
                response_text += "\n🔧 *Recommendations:*\n"
                for i, rec in enumerate(recommendations, 1):
                    response_text += f"{i}. {rec}\n"
            
            await respond(response_text)
    
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
    
    async def run(self):
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


async def main():
    """Main function"""
    # Check required environment variables
    required_vars = ["SLACK_BOT_TOKEN", "SLACK_SIGNING_SECRET", "SLACK_APP_TOKEN"]
    missing_vars = [var for var in required_vars if not os.environ.get(var)]
    
    if missing_vars:
        logger.error(f"Missing required environment variables: {missing_vars}")
        logger.error("Please set the following environment variables:")
        for var in missing_vars:
            logger.error(f"  export {var}=your_value_here")
        return
    
    # Create and run the bot
    bot = JamieSlackBot()
    await bot.run()


if __name__ == "__main__":
    asyncio.run(main())

