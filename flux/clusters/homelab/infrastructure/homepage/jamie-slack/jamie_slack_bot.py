#!/usr/bin/env python3
"""
🤖 Jamie Slack Bot - Your SRE Companion
A sophisticated SRE assistant that communicates via MCP (Model Context Protocol) 
and Ollama for AI-powered infrastructure management
"""

import os
import json
import asyncio
import logging
import time
from typing import Dict, List, Optional
from datetime import datetime

import aiohttp
import requests
from slack_bolt.async_app import AsyncApp
from slack_bolt.adapter.socket_mode.async_handler import AsyncSocketModeHandler
from slack_sdk.errors import SlackApiError

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class OllamaClient:
    """Client for interacting with Ollama API for AI responses"""
    
    def __init__(self, base_url: str = "http://192.168.0.16:11434"):
        self.base_url = base_url
        self.model_name = "bruno-sre"  # Bruno's fine-tuned SRE model
        self.session = None
    
    async def __aenter__(self):
        self.session = aiohttp.ClientSession()
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.session:
            await self.session.close()
    
    async def generate_response(self, prompt: str, context: Optional[str] = None) -> Dict:
        """Generate AI response from Ollama"""
        if not self.session:
            raise RuntimeError("OllamaClient not properly initialized")
        
        # Prepare the full prompt with context
        full_prompt = f"{context}\n\nUser: {prompt}\nAssistant:" if context else f"User: {prompt}\nAssistant:"
        
        payload = {
            "model": self.model_name,
            "prompt": full_prompt,
            "stream": False,
            "options": {
                "temperature": 0.7,
                "top_p": 0.9,
                "max_tokens": 1000,
                "repeat_penalty": 1.1
            }
        }
        
        try:
            start_time = time.time()
            async with self.session.post(
                f"{self.base_url}/api/generate",
                json=payload,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                duration = time.time() - start_time
                
                if response.status == 200:
                    result = await response.json()
                    response_text = result.get("response", "I'm sorry, I couldn't generate a response.")
                    
                    logger.info(f"Ollama response generated in {duration:.2f}s")
                    return {
                        "response": response_text,
                        "model": self.model_name,
                        "duration": duration,
                        "source": "ollama"
                    }
                else:
                    error_text = await response.text()
                    logger.error(f"Ollama API error: {response.status} - {error_text}")
                    return {
                        "response": "I'm having trouble connecting to my AI model. Please try again in a moment.",
                        "error": True,
                        "source": "ollama"
                    }
        
        except asyncio.TimeoutError:
            logger.error("Ollama API timeout")
            return {
                "response": "The AI model is taking longer than usual to respond. Please try again in a moment.",
                "error": True,
                "source": "ollama"
            }
        except Exception as e:
            logger.error(f"Ollama API error: {e}")
            return {
                "response": "I'm experiencing technical difficulties with my AI model. Please try again later.",
                "error": True,
                "source": "ollama"
            }
    
    async def is_available(self) -> bool:
        """Check if Ollama is available"""
        try:
            async with self.session.get(
                f"{self.base_url}/api/tags",
                timeout=aiohttp.ClientTimeout(total=5)
            ) as response:
                return response.status == 200
        except Exception:
            return False


class AgentSREClient:
    """Client for interacting with Agent-SRE via MCP"""
    
    def __init__(self, base_url: str = "http://homepage-api:8080"):
        self.base_url = base_url
        self.mcp_chat_url = f"{base_url}/api/v1/agent-sre/mcp/chat"
        self.mcp_analyze_logs_url = f"{base_url}/api/v1/agent-sre/mcp/analyze-logs"
        self.status_url = f"{base_url}/api/v1/agent-sre/status"
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
        
        payload = {
            "message": message,
            "timestamp": datetime.utcnow().isoformat(),
            "context": context or {}
        }
        
        try:
            async with self.session.post(
                self.mcp_chat_url,
                json=payload,
                timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                if response.status == 200:
                    result = await response.json()
                    result["source"] = "agent-sre-mcp"
                    return result
                else:
                    error_text = await response.text()
                    logger.error(f"Agent-SRE API error: {response.status} - {error_text}")
                    return {
                        "response": "I'm having trouble connecting to the SRE agent. The service might be temporarily unavailable.",
                        "error": True,
                        "source": "agent-sre-mcp"
                    }
        
        except asyncio.TimeoutError:
            logger.error("Agent-SRE API timeout")
            return {
                "response": "The SRE agent is taking longer than usual to respond. Please try again in a moment.",
                "error": True,
                "source": "agent-sre-mcp"
            }
        except Exception as e:
            logger.error(f"Agent-SRE API error: {e}")
            return {
                "response": "I'm experiencing technical difficulties connecting to the SRE agent. Please try again later.",
                "error": True,
                "source": "agent-sre-mcp"
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
        
        try:
            async with self.session.post(
                self.mcp_analyze_logs_url,
                json=payload,
                timeout=aiohttp.ClientTimeout(total=45)
            ) as response:
                if response.status == 200:
                    result = await response.json()
                    result["source"] = "agent-sre-mcp"
                    return result
                else:
                    logger.error(f"Log analysis API error: {response.status}")
                    return {
                        "analysis": "Unable to analyze logs at this time.",
                        "error": True,
                        "source": "agent-sre-mcp"
                    }
        except Exception as e:
            logger.error(f"Log analysis error: {e}")
            return {
                "analysis": "Error analyzing logs.",
                "error": True,
                "source": "agent-sre-mcp"
            }
    
    async def get_status(self) -> Dict:
        """Get Agent-SRE status"""
        if not self.session:
            raise RuntimeError("AgentSREClient not properly initialized")
        
        try:
            async with self.session.get(
                self.status_url,
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
    """🤖 Jamie - Your SRE Companion on Slack"""
    
    def __init__(self):
        # Initialize Slack app
        self.app = AsyncApp(
            token=os.environ.get("SLACK_BOT_TOKEN"),
            signing_secret=os.environ.get("SLACK_SIGNING_SECRET")
        )
        
        # Bot configuration
        self.bot_name = "Jamie"
        self.bot_emoji = "🤖"
        self.agent_url = os.environ.get("AGENT_SRE_URL", "http://homepage-api:8080")
        self.ollama_url = os.environ.get("OLLAMA_URL", "http://192.168.0.16:11434")
        
        # User context storage (in production, use Redis or proper database)
        self.user_contexts: Dict[str, Dict] = {}
        
        # Set up event handlers
        self._setup_handlers()
        
        logger.info(f"🤖 Jamie Slack Bot initialized with Agent-SRE at {self.agent_url} and Ollama at {self.ollama_url}")
    
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
            
            # Determine which service to use based on the question
            response = await self._route_question(question, context)
            
            # Update user context
            self._update_user_context(user_id, question, response.get("response", ""))
            
            # Format response with sources if available
            response_text = response.get("response", "I couldn't process that request.")
            sources = response.get("sources", [])
            model = response.get("model", "")
            source = response.get("source", "")
            
            formatted_response = f"{self.bot_emoji} {response_text}"
            
            if sources:
                formatted_response += f"\n\n_Sources: {', '.join(sources)}_"
            
            if model and source:
                formatted_response += f"\n_Powered by: {source} ({model})_"
            elif source:
                formatted_response += f"\n_Powered by: {source}_"
            
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
            
            # Route the question
            response = await self._route_question(text, context)
            
            # Update user context
            self._update_user_context(user_id, text, response.get("response", ""))
            
            # Format response
            response_text = response.get("response", "I couldn't process that request.")
            sources = response.get("sources", [])
            source = response.get("source", "")
            
            formatted_response = f"{self.bot_emoji} {response_text}"
            
            if sources:
                formatted_response += f"\n\n_Sources: {', '.join(sources)}_"
            
            if source:
                formatted_response += f"\n_Powered by: {source}_"
            
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
• `/jamie-status` - Check service status
• `/jamie-analyze-logs` - Analyze logs from a file or paste

*What I can help with:*
• 📊 Golden Signals monitoring (latency, traffic, errors, saturation)
• ☸️ Kubernetes operations (pods, deployments, logs, status)
• 📈 Grafana operations (dashboards, incidents, alerts)
• 🔍 Log analysis and error pattern detection
• 🚨 Incident investigation and root cause analysis
• 📉 Performance metrics and optimization
• 🎯 Service health monitoring
• 🧠 General SRE questions and best practices

*Example questions:*
• "Check the golden signals for bruno site"
• "What's the error rate for the API?"
• "List all pods in the default namespace"
• "Show me the dashboard for homepage service"
• "Analyze these logs for errors"
• "What alerts are currently firing?"
• "Investigate high latency in the API"
• "How do I troubleshoot a failing deployment?"

*Powered by:*
• 🧠 Ollama AI (Bruno's fine-tuned SRE model)
• 🔧 Agent-SRE with MCP (Model Context Protocol)
• 📡 Real-time infrastructure monitoring

Just ask me anything about your infrastructure! 🚀
            """
            
            await respond(help_text)
        
        @self.app.command("/jamie-status")
        async def handle_status_command(ack, respond):
            """Handle /jamie-status slash command"""
            await ack()
            
            # Check both services
            ollama_status = "❌ Unavailable"
            agent_status = "❌ Unavailable"
            
            try:
                async with OllamaClient(self.ollama_url) as ollama:
                    if await ollama.is_available():
                        ollama_status = "✅ Available"
            except Exception as e:
                logger.error(f"Ollama status check failed: {e}")
            
            try:
                async with AgentSREClient(self.agent_url) as agent:
                    status = await agent.get_status()
                    if status.get("status") not in ["error", "unavailable"]:
                        agent_status = "✅ Available"
            except Exception as e:
                logger.error(f"Agent-SRE status check failed: {e}")
            
            status_text = f"""
{self.bot_emoji} *Jamie Service Status*

🧠 *Ollama AI:* {ollama_status}
🔧 *Agent-SRE:* {agent_status}

*Endpoints:*
• Ollama: {self.ollama_url}
• Agent-SRE: {self.agent_url}

*All systems operational!* 🚀
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
            
            # Try Agent-SRE first, fallback to Ollama
            analysis = None
            
            try:
                async with AgentSREClient(self.agent_url) as agent:
                    analysis = await agent.analyze_logs(logs)
            except Exception as e:
                logger.error(f"Agent-SRE log analysis failed: {e}")
            
            # Fallback to Ollama if Agent-SRE fails
            if not analysis or analysis.get("error"):
                try:
                    async with OllamaClient(self.ollama_url) as ollama:
                        analysis = await ollama.generate_response(
                            f"Analyze these logs for errors and provide recommendations: {logs}",
                            "You are an SRE expert. Analyze logs for errors, patterns, and provide actionable recommendations."
                        )
                except Exception as e:
                    logger.error(f"Ollama log analysis failed: {e}")
                    analysis = {
                        "analysis": "Unable to analyze logs at this time.",
                        "error": True
                    }
            
            analysis_text = analysis.get("analysis", "Unable to analyze logs.")
            severity = analysis.get("severity", "unknown")
            recommendations = analysis.get("recommendations", [])
            source = analysis.get("source", "ollama")
            
            response_text = f"""
{self.bot_emoji} *Log Analysis Results*

📝 *Analysis:*
{analysis_text}

⚠️ *Severity:* {severity}
🔧 *Powered by:* {source}
"""
            
            if recommendations:
                response_text += "\n🔧 *Recommendations:*\n"
                for i, rec in enumerate(recommendations, 1):
                    response_text += f"{i}. {rec}\n"
            
            await respond(response_text)
    
    async def _route_question(self, question: str, context: Optional[Dict]) -> Dict:
        """Route question to appropriate service based on content"""
        
        # Keywords that suggest Agent-SRE should handle it
        agent_keywords = [
            "golden signals", "latency", "traffic", "error rate", "saturation",
            "pods", "deployment", "namespace", "kubernetes", "k8s",
            "grafana", "dashboard", "incident", "alert", "prometheus",
            "logs", "investigate", "troubleshoot", "analyze"
        ]
        
        # Check if question contains Agent-SRE keywords
        question_lower = question.lower()
        should_use_agent = any(keyword in question_lower for keyword in agent_keywords)
        
        if should_use_agent:
            try:
                async with AgentSREClient(self.agent_url) as agent:
                    response = await agent.chat(question, context)
                    if not response.get("error"):
                        return response
            except Exception as e:
                logger.error(f"Agent-SRE routing failed: {e}")
        
        # Fallback to Ollama for general questions or if Agent-SRE fails
        try:
            async with OllamaClient(self.ollama_url) as ollama:
                return await ollama.generate_response(question, self._format_context(context))
        except Exception as e:
            logger.error(f"Ollama routing failed: {e}")
            return {
                "response": "I'm experiencing technical difficulties. Please try again later.",
                "error": True,
                "source": "error"
            }
    
    def _format_context(self, context: Optional[Dict]) -> Optional[str]:
        """Format context for Ollama"""
        if not context:
            return None
        
        history = context.get("conversation_history", [])
        if not history:
            return None
        
        # Format conversation history for Ollama
        formatted_history = []
        for msg in history[-10:]:  # Last 10 messages
            role = msg.get("role", "")
            content = msg.get("content", "")
            if role and content:
                formatted_history.append(f"{role.title()}: {content}")
        
        return "\n".join(formatted_history)
    
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