#!/usr/bin/env python3
"""
🤖 Jamie MCP Server
MCP protocol wrapper that exposes Jamie Slack Bot API to Cursor IDE
NO AI - just protocol adapter (KISS principle)
"""

import os
import json
import asyncio
from typing import Dict, Any, List
from datetime import datetime
from aiohttp import web, ClientSession, ClientTimeout
from aiohttp.web import Request, Response

# Import from core
from core import logger, logfire

# Jamie Slack Bot API URL (the actual AI service)
JAMIE_API_URL = os.getenv("JAMIE_API_URL", "http://jamie-slack-bot-service:8080")

class JamieMCPServer:
    """MCP Server that exposes Jamie Slack Bot via MCP protocol."""
    
    def __init__(self):
        self.app = web.Application()
        self.jamie_api_url = JAMIE_API_URL
        self._setup_routes()
        logger.info(f"🤖 Jamie MCP Server initialized")
        logger.info(f"   🔗 Jamie API: {self.jamie_api_url}")
    
    def _setup_routes(self):
        """Setup HTTP routes for MCP protocol and REST API."""
        # MCP protocol endpoints
        self.app.router.add_post('/mcp', self.handle_mcp_request)
        self.app.router.add_get('/mcp', self.handle_mcp_info)
        
        # REST API endpoints (proxy to Jamie Slack Bot)
        self.app.router.add_post('/api/chat', self.handle_rest_chat)
        self.app.router.add_post('/api/golden-signals', self.handle_rest_golden_signals)
        self.app.router.add_post('/api/prometheus/query', self.handle_rest_prometheus_query)
        self.app.router.add_post('/api/pod-logs', self.handle_rest_pod_logs)
        self.app.router.add_post('/api/analyze-logs', self.handle_rest_analyze_logs)
        
        # Health and readiness endpoints
        self.app.router.add_get('/health', self.handle_health)
        self.app.router.add_get('/ready', self.handle_readiness)
    
    @logfire.instrument("mcp_info")
    async def handle_mcp_info(self, request: Request) -> Response:
        """Handle GET requests - MCP server information."""
        return web.json_response({
            "name": "jamie-mcp-server",
            "version": "1.0.0",
            "description": "🤖 Jamie - MCP wrapper for Jamie Slack Bot API",
            "protocol": "mcp",
            "capabilities": {
                "tools": True,
                "resources": False,
                "prompts": False
            },
            "endpoints": {
                "mcp": "/mcp",
                "health": "/health",
                "ready": "/ready"
            },
            "jamie_api_url": self.jamie_api_url
        })
    
    @logfire.instrument("mcp_request")
    async def handle_mcp_request(self, request: Request) -> Response:
        """Handle MCP JSON-RPC 2.0 requests - forward to Jamie API."""
        try:
            data = await request.json()
            
            # Handle notifications (no id field)
            if 'method' in data and 'id' not in data:
                return web.json_response({})
            
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
                                "name": "jamie-mcp-server",
                                "version": "1.0.0"
                            }
                        }
                    })
                
                elif method == 'tools/list':
                    tools = [
                        {
                            "name": "ask_jamie",
                            "description": "🤖 Ask Jamie anything about SRE. Jamie is an AI-powered SRE assistant.",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "question": {
                                        "type": "string",
                                        "description": "Your SRE question"
                                    }
                                },
                                "required": ["question"]
                            }
                        },
                        {
                            "name": "check_golden_signals",
                            "description": "📊 Check golden signals for a service",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "service_name": {
                                        "type": "string",
                                        "description": "Service name"
                                    },
                                    "namespace": {
                                        "type": "string",
                                        "description": "Kubernetes namespace",
                                        "default": "default"
                                    }
                                },
                                "required": ["service_name"]
                            }
                        },
                        {
                            "name": "query_prometheus",
                            "description": "📈 Execute PromQL query",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "query": {
                                        "type": "string",
                                        "description": "PromQL query"
                                    }
                                },
                                "required": ["query"]
                            }
                        },
                        {
                            "name": "get_pod_logs",
                            "description": "📝 Get Kubernetes pod logs",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "pod_name": {
                                        "type": "string",
                                        "description": "Pod name"
                                    },
                                    "namespace": {
                                        "type": "string",
                                        "description": "Namespace",
                                        "default": "default"
                                    },
                                    "tail_lines": {
                                        "type": "number",
                                        "description": "Number of lines",
                                        "default": 100
                                    }
                                },
                                "required": ["pod_name"]
                            }
                        },
                        {
                            "name": "analyze_logs",
                            "description": "🔍 Analyze logs with AI",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "logs": {
                                        "type": "string",
                                        "description": "Log data"
                                    },
                                    "context": {
                                        "type": "string",
                                        "description": "Optional context"
                                    }
                                },
                                "required": ["logs"]
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
                    
                    # Forward to Jamie Slack Bot API
                    result = await self._call_jamie_api(tool_name, arguments)
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
            logger.error(f"❌ Error handling MCP request: {e}")
            return web.json_response({
                "jsonrpc": "2.0",
                "error": {
                    "code": -32603,
                    "message": str(e)
                }
            }, status=500)
    
    @logfire.instrument("call_jamie_api")
    async def _call_jamie_api(self, tool_name: str, arguments: Dict[str, Any]) -> str:
        """Forward tool calls to Jamie Slack Bot REST API."""
        try:
            async with ClientSession() as session:
                # Map MCP tool names to Jamie API endpoints
                endpoint_map = {
                    "ask_jamie": "/api/chat",
                    "check_golden_signals": "/api/golden-signals",
                    "query_prometheus": "/api/prometheus/query",
                    "get_pod_logs": "/api/pod-logs",
                    "analyze_logs": "/api/analyze-logs"
                }
                
                endpoint = endpoint_map.get(tool_name)
                if not endpoint:
                    return f"❌ Unknown tool: {tool_name}"
                
                # Prepare request data
                if tool_name == "ask_jamie":
                    request_data = {"message": arguments.get("question", "")}
                else:
                    request_data = arguments
                
                # Call Jamie Slack Bot API
                url = f"{self.jamie_api_url}{endpoint}"
                
                async with session.post(url, json=request_data, timeout=ClientTimeout(total=60)) as response:
                    if response.status == 200:
                        data = await response.json()
                        # Extract response based on endpoint
                        if tool_name == "ask_jamie":
                            return data.get("response", "No response")
                        else:
                            return json.dumps(data, indent=2)
                    else:
                        error_text = await response.text()
                        logger.error(f"❌ Jamie API error: {response.status} - {error_text}")
                        return f"❌ Error calling Jamie API: {error_text}"
        
        except Exception as e:
            logger.error(f"❌ Error calling Jamie API {tool_name}: {e}")
            return f"❌ Error: {str(e)}"
    
    # REST API Handlers - Proxy to Jamie Slack Bot
    
    @logfire.instrument("rest_chat")
    async def handle_rest_chat(self, request: Request) -> Response:
        """Proxy chat requests to Jamie Slack Bot API."""
        try:
            data = await request.json()
            
            async with ClientSession() as session:
                async with session.post(
                    f"{self.jamie_api_url}/api/chat",
                    json=data,
                    timeout=ClientTimeout(total=60)
                ) as response:
                    result = await response.json()
                    return web.json_response(result, status=response.status)
        
        except Exception as e:
            logger.error(f"❌ Error in REST chat: {e}")
            return web.json_response({"error": str(e)}, status=500)
    
    @logfire.instrument("rest_golden_signals")
    async def handle_rest_golden_signals(self, request: Request) -> Response:
        """Proxy golden signals requests to Jamie Slack Bot API."""
        try:
            data = await request.json()
            
            async with ClientSession() as session:
                async with session.post(
                    f"{self.jamie_api_url}/api/golden-signals",
                    json=data,
                    timeout=ClientTimeout(total=30)
                ) as response:
                    result = await response.json()
                    return web.json_response(result, status=response.status)
        
        except Exception as e:
            logger.error(f"❌ Error in golden signals: {e}")
            return web.json_response({"error": str(e)}, status=500)
    
    @logfire.instrument("rest_prometheus_query")
    async def handle_rest_prometheus_query(self, request: Request) -> Response:
        """Proxy Prometheus query requests to Jamie Slack Bot API."""
        try:
            data = await request.json()
            
            async with ClientSession() as session:
                async with session.post(
                    f"{self.jamie_api_url}/api/prometheus/query",
                    json=data,
                    timeout=ClientTimeout(total=30)
                ) as response:
                    result = await response.json()
                    return web.json_response(result, status=response.status)
        
        except Exception as e:
            logger.error(f"❌ Error in prometheus query: {e}")
            return web.json_response({"error": str(e)}, status=500)
    
    @logfire.instrument("rest_pod_logs")
    async def handle_rest_pod_logs(self, request: Request) -> Response:
        """Proxy pod logs requests to Jamie Slack Bot API."""
        try:
            data = await request.json()
            
            async with ClientSession() as session:
                async with session.post(
                    f"{self.jamie_api_url}/api/pod-logs",
                    json=data,
                    timeout=ClientTimeout(total=30)
                ) as response:
                    result = await response.json()
                    return web.json_response(result, status=response.status)
        
        except Exception as e:
            logger.error(f"❌ Error in pod logs: {e}")
            return web.json_response({"error": str(e)}, status=500)
    
    @logfire.instrument("rest_analyze_logs")
    async def handle_rest_analyze_logs(self, request: Request) -> Response:
        """Proxy analyze logs requests to Jamie Slack Bot API."""
        try:
            data = await request.json()
            
            async with ClientSession() as session:
                async with session.post(
                    f"{self.jamie_api_url}/api/analyze-logs",
                    json=data,
                    timeout=ClientTimeout(total=45)
                ) as response:
                    result = await response.json()
                    return web.json_response(result, status=response.status)
        
        except Exception as e:
            logger.error(f"❌ Error in analyze logs: {e}")
            return web.json_response({"error": str(e)}, status=500)
    
    @logfire.instrument("health_check")
    async def handle_health(self, request: Request) -> Response:
        """Health check endpoint."""
        try:
            health_status = {
                "status": "healthy",
                "service": "jamie-mcp-server",
                "timestamp": datetime.now().isoformat(),
                "version": "1.0.0",
                "jamie_api_url": self.jamie_api_url
            }
            return web.json_response(health_status)
        except Exception as e:
            logger.error(f"❌ Health check failed: {e}")
            return web.json_response(
                {"status": "unhealthy", "error": str(e)}, 
                status=503
            )
    
    @logfire.instrument("readiness_check")
    async def handle_readiness(self, request: Request) -> Response:
        """Readiness check endpoint."""
        try:
            # Check if Jamie Slack Bot API is reachable
            async with ClientSession() as session:
                try:
                    async with session.get(
                        f"{self.jamie_api_url}/health",
                        timeout=ClientTimeout(total=5)
                    ) as response:
                        jamie_status = response.status == 200
                except:
                    jamie_status = False
            
            if not jamie_status:
                return web.json_response(
                    {"status": "not_ready", "reason": "Jamie API not available"}, 
                    status=503
                )
            
            readiness_status = {
                "status": "ready",
                "service": "jamie-mcp-server",
                "timestamp": datetime.now().isoformat(),
                "jamie_api_status": "connected"
            }
            return web.json_response(readiness_status)
            
        except Exception as e:
            logger.error(f"❌ Readiness check failed: {e}")
            return web.json_response(
                {"status": "not_ready", "error": str(e)}, 
                status=503
            )
    
    async def start_server(self, host: str = "0.0.0.0", port: int = 30121):
        """Start the MCP server."""
        runner = web.AppRunner(self.app)
        await runner.setup()
        site = web.TCPSite(runner, host, port)
        await site.start()
        
        logger.info("=" * 60)
        logger.info("🤖 Jamie MCP Server Started!")
        logger.info("=" * 60)
        logger.info(f"🌐 Server: http://{host}:{port}")
        logger.info(f"📋 MCP endpoint: http://localhost:{port}/mcp")
        logger.info("")
        logger.info("🔌 REST API endpoints (proxy to Jamie API):")
        logger.info(f"   💬 Chat: http://localhost:{port}/api/chat")
        logger.info(f"   📊 Golden Signals: http://localhost:{port}/api/golden-signals")
        logger.info(f"   📈 Prometheus: http://localhost:{port}/api/prometheus/query")
        logger.info(f"   📝 Pod Logs: http://localhost:{port}/api/pod-logs")
        logger.info(f"   🔍 Analyze Logs: http://localhost:{port}/api/analyze-logs")
        logger.info("")
        logger.info(f"🏥 Health: http://localhost:{port}/health")
        logger.info(f"✅ Ready: http://localhost:{port}/ready")
        logger.info(f"🔗 Jamie API: {self.jamie_api_url}")
        logger.info("=" * 60)
        
        return runner

async def main():
    """Main entry point for Jamie MCP Server."""
    logger.info("🚀 Starting Jamie MCP Server (MCP wrapper for Jamie API)")
    
    # Configure server options
    host = os.getenv("MCP_HOST", "0.0.0.0")
    port = int(os.getenv("MCP_PORT", "30121"))
    
    server = JamieMCPServer()
    runner = await server.start_server(host, port)
    
    try:
        logger.info("🏁 Jamie MCP Server is running...")
        await asyncio.Event().wait()  # Run forever
    except KeyboardInterrupt:
        logger.info("🛑 Shutting down Jamie MCP Server...")
    finally:
        await runner.cleanup()

if __name__ == "__main__":
    asyncio.run(main())
