#!/usr/bin/env python3
"""
SRE Agent - Standalone Agent Service
Handles HTTP API requests and communicates with MCP server
"""

import os
import json
import asyncio
import logging
from typing import Dict, Any, List, Optional
from datetime import datetime
from aiohttp import web, ClientSession
from aiohttp.web import Request, Response

# Import the SRE agent from local core
from core import agent, logger

class SREAgentService:
    """Standalone SRE Agent Service."""
    
    def __init__(self):
        self.sre_agent = agent
        self.app = web.Application()
        self.mcp_server_url = os.getenv("MCP_SERVER_URL", "http://sre-agent-mcp-server-service:30120")
        self._setup_routes()
    
    def _setup_routes(self):
        """Setup HTTP routes."""
        # Health and readiness endpoints
        self.app.router.add_get('/health', self.handle_health)
        self.app.router.add_get('/ready', self.handle_readiness)
        
        # Agent API endpoints
        self.app.router.add_post('/chat', self.handle_chat)
        self.app.router.add_post('/analyze-logs', self.handle_analyze_logs)
        self.app.router.add_post('/incident-response', self.handle_incident_response)
        self.app.router.add_post('/monitoring-advice', self.handle_monitoring_advice)
        
        # MCP server communication endpoints
        self.app.router.add_post('/mcp/chat', self.handle_mcp_chat)
        self.app.router.add_post('/mcp/analyze-logs', self.handle_mcp_analyze_logs)
        self.app.router.add_post('/mcp/incident-response', self.handle_mcp_incident_response)
        self.app.router.add_post('/mcp/monitoring-advice', self.handle_mcp_monitoring_advice)
        
        # Status and info endpoints
        self.app.router.add_get('/status', self.handle_status)
        self.app.router.add_get('/mcp/status', self.handle_mcp_status)
    
    async def handle_health(self, request: Request) -> Response:
        """Liveness probe endpoint."""
        try:
            health_status = {
                "status": "healthy",
                "service": "sre-agent",
                "timestamp": datetime.now().isoformat(),
                "uptime": "running",
                "version": "1.0.0",
                "deployment": "standalone-agent"
            }
            # Only log health checks in debug mode
            if os.getenv("DEBUG", "false").lower() == "true":
                logger.debug(f"Health check: {health_status}")
            return web.json_response(health_status)
        except Exception as e:
            logger.error(f"Health check failed: {e}")
            return web.json_response(
                {"status": "unhealthy", "error": str(e)}, 
                status=503
            )
    
    async def handle_readiness(self, request: Request) -> Response:
        """Readiness probe endpoint."""
        try:
            # Check if the SRE agent is properly initialized
            if not self.sre_agent or not self.sre_agent.llm:
                return web.json_response(
                    {"status": "not_ready", "reason": "SRE agent not initialized"}, 
                    status=503
                )
            
            # Check MCP server connectivity
            mcp_status = await self._check_mcp_server()
            
            readiness_status = {
                "status": "ready",
                "service": "sre-agent",
                "timestamp": datetime.now().isoformat(),
                "mcp_server_status": mcp_status,
                "deployment": "standalone-agent"
            }
            # Only log readiness checks in debug mode
            if os.getenv("DEBUG", "false").lower() == "true":
                logger.debug(f"Readiness check: {readiness_status}")
            return web.json_response(readiness_status)
            
        except Exception as e:
            logger.error(f"Readiness check failed: {e}")
            return web.json_response(
                {"status": "not_ready", "error": str(e)}, 
                status=503
            )
    
    async def handle_chat(self, request: Request) -> Response:
        """Direct chat endpoint using local agent."""
        try:
            data = await request.json()
            message = data.get("message", "")
            
            if not message:
                return web.json_response(
                    {"error": "Message is required"},
                    status=400
                )
            
            response = await self.sre_agent.chat(message)
            return web.json_response({
                "response": response,
                "service": "sre-agent",
                "timestamp": datetime.now().isoformat(),
                "method": "direct"
            })
        
        except Exception as e:
            logger.error(f"Error in chat handler: {e}")
            return web.json_response(
                {"error": str(e)},
                status=500
            )
    
    async def handle_analyze_logs(self, request: Request) -> Response:
        """Direct log analysis endpoint using local agent."""
        try:
            data = await request.json()
            logs = data.get("logs", "")
            
            if not logs:
                return web.json_response(
                    {"error": "Logs are required"},
                    status=400
                )
            
            analysis = await self.sre_agent.analyze_logs(logs)
            return web.json_response({
                "analysis": analysis,
                "service": "sre-agent",
                "timestamp": datetime.now().isoformat(),
                "method": "direct"
            })
        
        except Exception as e:
            logger.error(f"Error in analyze_logs handler: {e}")
            return web.json_response(
                {"error": str(e)},
                status=500
            )
    
    async def handle_incident_response(self, request: Request) -> Response:
        """Direct incident response endpoint using local agent."""
        try:
            data = await request.json()
            incident = data.get("incident", "")
            
            if not incident:
                return web.json_response(
                    {"error": "Incident description is required"},
                    status=400
                )
            
            response = await self.sre_agent.incident_response(incident)
            return web.json_response({
                "response": response,
                "service": "sre-agent",
                "timestamp": datetime.now().isoformat(),
                "method": "direct"
            })
        
        except Exception as e:
            logger.error(f"Error in incident_response handler: {e}")
            return web.json_response(
                {"error": str(e)},
                status=500
            )
    
    async def handle_monitoring_advice(self, request: Request) -> Response:
        """Direct monitoring advice endpoint using local agent."""
        try:
            data = await request.json()
            system = data.get("system", "")
            
            if not system:
                return web.json_response(
                    {"error": "System description is required"},
                    status=400
                )
            
            advice = await self.sre_agent.monitoring_advice(system)
            return web.json_response({
                "advice": advice,
                "service": "sre-agent",
                "timestamp": datetime.now().isoformat(),
                "method": "direct"
            })
        
        except Exception as e:
            logger.error(f"Error in monitoring_advice handler: {e}")
            return web.json_response(
                {"error": str(e)},
                status=500
            )
    
    async def handle_mcp_chat(self, request: Request) -> Response:
        """Chat endpoint via MCP server."""
        try:
            data = await request.json()
            message = data.get("message", "")
            
            if not message:
                return web.json_response(
                    {"error": "Message is required"},
                    status=400
                )
            
            result = await self._call_mcp_tool("sre_chat", {"message": message})
            return web.json_response({
                "response": result,
                "service": "sre-agent",
                "timestamp": datetime.now().isoformat(),
                "method": "mcp"
            })
        
        except Exception as e:
            logger.error(f"Error in MCP chat handler: {e}")
            return web.json_response(
                {"error": str(e)},
                status=500
            )
    
    async def handle_mcp_analyze_logs(self, request: Request) -> Response:
        """Log analysis endpoint via MCP server."""
        try:
            data = await request.json()
            logs = data.get("logs", "")
            
            if not logs:
                return web.json_response(
                    {"error": "Logs are required"},
                    status=400
                )
            
            result = await self._call_mcp_tool("analyze_logs", {"logs": logs})
            return web.json_response({
                "analysis": result,
                "service": "sre-agent",
                "timestamp": datetime.now().isoformat(),
                "method": "mcp"
            })
        
        except Exception as e:
            logger.error(f"Error in MCP analyze_logs handler: {e}")
            return web.json_response(
                {"error": str(e)},
                status=500
            )
    
    async def handle_mcp_incident_response(self, request: Request) -> Response:
        """Incident response endpoint via MCP server."""
        try:
            data = await request.json()
            incident = data.get("incident", "")
            
            if not incident:
                return web.json_response(
                    {"error": "Incident description is required"},
                    status=400
                )
            
            result = await self._call_mcp_tool("incident_response", {"incident": incident})
            return web.json_response({
                "response": result,
                "service": "sre-agent",
                "timestamp": datetime.now().isoformat(),
                "method": "mcp"
            })
        
        except Exception as e:
            logger.error(f"Error in MCP incident_response handler: {e}")
            return web.json_response(
                {"error": str(e)},
                status=500
            )
    
    async def handle_mcp_monitoring_advice(self, request: Request) -> Response:
        """Monitoring advice endpoint via MCP server."""
        try:
            data = await request.json()
            system = data.get("system", "")
            
            if not system:
                return web.json_response(
                    {"error": "System description is required"},
                    status=400
                )
            
            result = await self._call_mcp_tool("monitoring_advice", {"system": system})
            return web.json_response({
                "advice": result,
                "service": "sre-agent",
                "timestamp": datetime.now().isoformat(),
                "method": "mcp"
            })
        
        except Exception as e:
            logger.error(f"Error in MCP monitoring_advice handler: {e}")
            return web.json_response(
                {"error": str(e)},
                status=500
            )
    
    async def handle_status(self, request: Request) -> Response:
        """Agent status endpoint."""
        try:
            health = await self.sre_agent.health_check()
            mcp_status = await self._check_mcp_server()
            
            status = {
                "agent": health,
                "mcp_server": mcp_status,
                "service": "sre-agent",
                "timestamp": datetime.now().isoformat(),
                "deployment": "standalone-agent"
            }
            return web.json_response(status)
        
        except Exception as e:
            logger.error(f"Error in status handler: {e}")
            return web.json_response(
                {"error": str(e)},
                status=500
            )
    
    async def handle_mcp_status(self, request: Request) -> Response:
        """MCP server status endpoint."""
        try:
            mcp_status = await self._check_mcp_server()
            return web.json_response(mcp_status)
        
        except Exception as e:
            logger.error(f"Error in MCP status handler: {e}")
            return web.json_response(
                {"error": str(e)},
                status=500
            )
    
    async def _check_mcp_server(self) -> Dict[str, Any]:
        """Check MCP server connectivity."""
        try:
            async with ClientSession() as session:
                async with session.get(f"{self.mcp_server_url}/health", timeout=5) as response:
                    if response.status == 200:
                        data = await response.json()
                        return {
                            "status": "connected",
                            "url": self.mcp_server_url,
                            "health": data
                        }
                    else:
                        return {
                            "status": "error",
                            "url": self.mcp_server_url,
                            "error": f"HTTP {response.status}"
                        }
        except Exception as e:
            return {
                "status": "disconnected",
                "url": self.mcp_server_url,
                "error": str(e)
            }
    
    async def _call_mcp_tool(self, tool_name: str, arguments: Dict[str, Any]) -> str:
        """Call MCP server tool."""
        try:
            async with ClientSession() as session:
                mcp_request = {
                    "jsonrpc": "2.0",
                    "id": 1,
                    "method": "tools/call",
                    "params": {
                        "name": tool_name,
                        "arguments": arguments
                    }
                }
                
                async with session.post(
                    f"{self.mcp_server_url}/mcp",
                    json=mcp_request,
                    timeout=30
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        if "result" in data and "content" in data["result"]:
                            return data["result"]["content"][0]["text"]
                        else:
                            return f"Error: {data.get('error', 'Unknown error')}"
                    else:
                        return f"Error: HTTP {response.status}"
        
        except Exception as e:
            logger.error(f"Error calling MCP tool {tool_name}: {e}")
            return f"Error calling MCP tool: {str(e)}"
    
    async def start_server(self, host: str = "0.0.0.0", port: int = 8080):
        """Start the agent server."""
        runner = web.AppRunner(self.app)
        await runner.setup()
        site = web.TCPSite(runner, host, port)
        await site.start()
        
        logger.info(f"🌐 SRE Agent started on {host}:{port}")
        logger.info(f"🏥 Health endpoint: http://localhost:{port}/health")
        logger.info(f"✅ Readiness endpoint: http://localhost:{port}/ready")
        logger.info(f"💬 Chat endpoint: http://localhost:{port}/chat")
        logger.info(f"📊 MCP Chat endpoint: http://localhost:{port}/mcp/chat")
        logger.info(f"📈 Status endpoint: http://localhost:{port}/status")
        
        return runner

async def main():
    """Main entry point for SRE Agent."""
    logger.info("🚀 Starting SRE Agent (Standalone)")
    
    # Configure server options
    host = os.getenv("AGENT_HOST", "0.0.0.0")
    port = int(os.getenv("AGENT_PORT", "8080"))
    
    service = SREAgentService()
    runner = await service.start_server(host, port)
    
    try:
        logger.info("🏁 SRE Agent is running...")
        await asyncio.Event().wait()  # Run forever
    except KeyboardInterrupt:
        logger.info("🛑 Shutting down SRE Agent...")
    finally:
        await runner.cleanup()

if __name__ == "__main__":
    asyncio.run(main())
