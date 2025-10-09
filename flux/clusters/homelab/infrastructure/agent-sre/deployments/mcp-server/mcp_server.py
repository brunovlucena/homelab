#!/usr/bin/env python3
"""
MCP Server - Thin Protocol Layer for SRE Agent
Handles MCP protocol communication and forwards requests to agent service
"""

import os
import json
import asyncio
import logging
from typing import Dict, Any, List, Optional
from datetime import datetime
from aiohttp import web, ClientSession
from aiohttp.web import Request, Response

# Import the core module for shared functionality
from core import logger, logfire

class MCPServer:
    """Thin MCP Server that forwards requests to agent service."""
    
    def __init__(self):
        self.app = web.Application()
        self.agent_service_url = os.getenv("AGENT_SERVICE_URL", "http://sre-agent-service:8080")
        self._setup_routes()
    
    def _setup_routes(self):
        """Setup HTTP routes."""
        # MCP protocol endpoints
        self.app.router.add_post('/mcp', self.handle_mcp_request)
        self.app.router.add_get('/mcp', self.handle_mcp_info)
        self.app.router.add_post('/mcp/', self.handle_mcp_request)
        self.app.router.add_get('/mcp/', self.handle_mcp_info)
        
        # Health and readiness endpoints
        self.app.router.add_get('/health', self.handle_health)
        self.app.router.add_get('/ready', self.handle_readiness)
        
        # SSE endpoint for real-time communication
        self.app.router.add_get('/sse', self.handle_sse)
    
    @logfire.instrument("mcp_info")
    async def handle_mcp_info(self, request: Request) -> Response:
        """Handle GET requests - MCP server information."""
        return web.json_response({
            "name": "sre-agent-mcp-server",
            "version": "1.0.0",
            "description": "SRE Agent MCP Server - Thin protocol layer",
            "protocol": "mcp",
            "capabilities": {
                "tools": True,
                "resources": False,
                "prompts": False
            },
            "endpoints": {
                "mcp": "/mcp",
                "health": "/health",
                "ready": "/ready",
                "sse": "/sse"
            },
            "agent_service": self.agent_service_url
        })
    
    @logfire.instrument("mcp_request")
    async def handle_mcp_request(self, request: Request) -> Response:
        """Handle MCP JSON-RPC 2.0 requests."""
        try:
            data = await request.json()
            
            # Handle notifications (no id field)
            if 'method' in data and 'id' not in data:
                method = data.get('method')
                if method == 'notifications/initialized':
                    return web.json_response({})  # Empty response for notifications
                else:
                    return web.json_response({})  # Empty response for other notifications
            
            if 'method' in data and 'id' in data:
                method = data.get('method')
                params = data.get('params', {})
                request_id = data.get('id')
                
                if method == 'initialize':
                    return web.json_response({
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "result": {
                            "protocolVersion": "2024-11-05",
                            "capabilities": {
                                "tools": {}
                            },
                            "serverInfo": {
                                "name": "sre-agent-mcp-server",
                                "version": "1.0.0"
                            }
                        }
                    })
                
                elif method == 'tools/list':
                    tools = [
                        {
                            "name": "sre_chat",
                            "description": "General SRE chat and consultation",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "message": {
                                        "type": "string",
                                        "description": "Your SRE question or request"
                                    }
                                },
                                "required": ["message"]
                            }
                        },
                        {
                            "name": "analyze_logs",
                            "description": "Analyze logs for SRE insights",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "logs": {
                                        "type": "string",
                                        "description": "Log data to analyze"
                                    }
                                },
                                "required": ["logs"]
                            }
                        },
                        {
                            "name": "incident_response",
                            "description": "Get incident response guidance",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "incident": {
                                        "type": "string",
                                        "description": "Incident description"
                                    }
                                },
                                "required": ["incident"]
                            }
                        },
                        {
                            "name": "monitoring_advice",
                            "description": "Get monitoring and alerting advice",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "system": {
                                        "type": "string",
                                        "description": "System description"
                                    }
                                },
                                "required": ["system"]
                            }
                        },
                        {
                            "name": "health_check",
                            "description": "Check the health status",
                            "inputSchema": {
                                "type": "object",
                                "properties": {}
                            }
                        }
                    ]
                    
                    return web.json_response({
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "result": {
                            "tools": tools
                        }
                    })
                
                elif method == 'tools/call':
                    tool_name = params.get('name')
                    arguments = params.get('arguments', {})
                    
                    if not tool_name:
                        return web.json_response({
                            "jsonrpc": "2.0",
                            "id": request_id,
                            "error": {
                                "code": -32602,
                                "message": "Tool name is required"
                            }
                        })
                    
                    result = await self._forward_to_agent(tool_name, arguments)
                    return web.json_response({
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "result": {
                            "content": [
                                {
                                    "type": "text",
                                    "text": result
                                }
                            ]
                        }
                    })
                
                else:
                    return web.json_response({
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "error": {
                            "code": -32601,
                            "message": f"Unknown method: {method}"
                        }
                    })
            
            else:
                return web.json_response({
                    "jsonrpc": "2.0",
                    "error": {
                        "code": -32700,
                        "message": "Parse error"
                    }
                }, status=400)
                
        except Exception as e:
            logger.error(f"Error handling MCP request: {e}")
            return web.json_response({
                "jsonrpc": "2.0",
                "error": {
                    "code": -32603,
                    "message": str(e)
                }
            }, status=500)
    
    @logfire.instrument("forward_to_agent")
    async def _forward_to_agent(self, tool_name: str, arguments: Dict[str, Any]) -> str:
        """Forward tool execution to the agent service."""
        try:
            async with ClientSession() as session:
                # Map MCP tool names to agent service endpoints
                endpoint_map = {
                    "sre_chat": "/chat",
                    "analyze_logs": "/analyze-logs", 
                    "incident_response": "/incident-response",
                    "monitoring_advice": "/monitoring-advice",
                    "health_check": "/health"
                }
                
                endpoint = endpoint_map.get(tool_name)
                if not endpoint:
                    return f"❌ Unknown tool: {tool_name}"
                
                # Prepare request data
                if tool_name == "sre_chat":
                    request_data = {"message": arguments.get("message", "")}
                elif tool_name == "analyze_logs":
                    request_data = {"logs": arguments.get("logs", "")}
                elif tool_name == "incident_response":
                    request_data = {"incident": arguments.get("incident", "")}
                elif tool_name == "monitoring_advice":
                    request_data = {"system": arguments.get("system", "")}
                elif tool_name == "health_check":
                    request_data = {}
                else:
                    return f"❌ Unknown tool: {tool_name}"
                
                # Make request to agent service
                url = f"{self.agent_service_url}{endpoint}"
                
                if tool_name == "health_check":
                    # Health check is a GET request
                    async with session.get(url, timeout=10) as response:
                        if response.status == 200:
                            data = await response.json()
                            return json.dumps(data, indent=2)
                        else:
                            return f"Error: HTTP {response.status}"
                else:
                    # Other tools are POST requests
                    async with session.post(url, json=request_data, timeout=30) as response:
                        if response.status == 200:
                            data = await response.json()
                            # Extract the response based on the endpoint
                            if tool_name == "sre_chat":
                                return data.get("response", "No response")
                            elif tool_name == "analyze_logs":
                                return data.get("analysis", "No analysis")
                            elif tool_name == "incident_response":
                                return data.get("response", "No response")
                            elif tool_name == "monitoring_advice":
                                return data.get("advice", "No advice")
                            else:
                                return json.dumps(data, indent=2)
                        else:
                            error_text = await response.text()
                            return f"Error: HTTP {response.status} - {error_text}"
        
        except Exception as e:
            logger.error(f"Error forwarding to agent service: {e}")
            return f"Error forwarding to agent service: {str(e)}"
    
    @logfire.instrument("mcp_health")
    async def handle_health(self, request: Request) -> Response:
        """Liveness probe endpoint - checks if the service is alive."""
        try:
            health_status = {
                "status": "healthy",
                "service": "sre-agent-mcp-server",
                "timestamp": datetime.now().isoformat(),
                "uptime": "running",
                "version": "1.0.0",
                "deployment": "thin-mcp-server",
                "agent_service_url": self.agent_service_url
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
    
    @logfire.instrument("mcp_readiness")
    async def handle_readiness(self, request: Request) -> Response:
        """Readiness probe endpoint - checks if the service is ready to serve traffic."""
        try:
            # Check if agent service is reachable
            agent_status = await self._check_agent_service()
            
            if agent_status.get("status") != "connected":
                return web.json_response(
                    {"status": "not_ready", "reason": "Agent service not available", "agent_status": agent_status}, 
                    status=503
                )
            
            readiness_status = {
                "status": "ready",
                "service": "sre-agent-mcp-server",
                "timestamp": datetime.now().isoformat(),
                "agent_service_status": agent_status,
                "mcp_endpoints": ["/mcp", "/health", "/ready", "/sse"],
                "deployment": "thin-mcp-server"
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
    
    @logfire.instrument("check_agent_service")
    async def _check_agent_service(self) -> Dict[str, Any]:
        """Check agent service connectivity."""
        try:
            async with ClientSession() as session:
                async with session.get(f"{self.agent_service_url}/health", timeout=5) as response:
                    if response.status == 200:
                        data = await response.json()
                        return {
                            "status": "connected",
                            "url": self.agent_service_url,
                            "health": data
                        }
                    else:
                        return {
                            "status": "error",
                            "url": self.agent_service_url,
                            "error": f"HTTP {response.status}"
                        }
        except Exception as e:
            return {
                "status": "disconnected",
                "url": self.agent_service_url,
                "error": str(e)
            }
    
    @logfire.instrument("mcp_sse")
    async def handle_sse(self, request: Request) -> Response:
        """Server-Sent Events endpoint for real-time communication."""
        response = web.StreamResponse()
        response.headers['Content-Type'] = 'text/event-stream'
        response.headers['Cache-Control'] = 'no-cache'
        response.headers['Connection'] = 'keep-alive'
        response.headers['Access-Control-Allow-Origin'] = '*'
        
        await response.prepare(request)
        
        try:
            # Send initial connection event
            await response.write(b"data: {\"type\": \"connected\", \"service\": \"sre-agent-mcp-server\", \"timestamp\": \"" + 
                               datetime.now().isoformat().encode() + b"\"}\n\n")
            
            # Send heartbeat every 10 seconds
            for i in range(100):  # Send heartbeats for ~16 minutes
                await asyncio.sleep(10)
                await response.write(b"data: {\"type\": \"heartbeat\", \"count\": " + 
                                   str(i).encode() + b", \"timestamp\": \"" + 
                                   datetime.now().isoformat().encode() + b"\"}\n\n")
                
        except Exception as e:
            logger.error(f"SSE error: {e}")
        finally:
            await response.write_eof()
        
        return response
    
    async def start_server(self, host: str = "0.0.0.0", port: int = 30120):
        """Start the MCP server."""
        runner = web.AppRunner(self.app)
        await runner.setup()
        site = web.TCPSite(runner, host, port)
        await site.start()
        
        logger.info(f"🌐 MCP Server started on {host}:{port}")
        logger.info(f"📋 MCP endpoint: http://localhost:{port}/mcp")
        logger.info(f"🏥 Health endpoint: http://localhost:{port}/health")
        logger.info(f"✅ Readiness endpoint: http://localhost:{port}/ready")
        logger.info(f"📡 SSE endpoint: http://localhost:{port}/sse")
        logger.info(f"🔗 Agent service: {self.agent_service_url}")
        
        return runner

async def main():
    """Main entry point for MCP Server."""
    logger.info("🚀 Starting SRE Agent MCP Server (Thin Layer)")
    
    # Configure server options
    host = os.getenv("MCP_HOST", "0.0.0.0")
    port = int(os.getenv("MCP_PORT", "30120"))
    
    server = MCPServer()
    runner = await server.start_server(host, port)
    
    try:
        logger.info("🏁 MCP Server is running...")
        await asyncio.Event().wait()  # Run forever
    except KeyboardInterrupt:
        logger.info("🛑 Shutting down MCP Server...")
    finally:
        await runner.cleanup()

if __name__ == "__main__":
    asyncio.run(main())