#!/usr/bin/env python3
"""
🌐 Agent-SRE MCP Server HTTP Wrapper
Exposes MCP tools via HTTP for Kubernetes deployment
Supports proper JSON-RPC 2.0 protocol for tools/list and tools/call
"""

import asyncio
import json
import logging
import os
import subprocess
from datetime import datetime
from typing import Any, Dict, List

import aiohttp
from aiohttp import web

# Import the actual MCP server functions
from mcp_server import (
    execute_grafana_query,
    execute_prometheus_query,
    execute_prometheus_query_range,
)
from mcp_server import list_tools as mcp_list_tools

# Configure logging
logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s")
logger = logging.getLogger(__name__)


class MCPHTTPWrapper:
    """🌐 HTTP wrapper for MCP server tools"""

    def __init__(self):
        self.app = web.Application()
        self._setup_routes()
        logger.info("🌐 MCP HTTP Wrapper initialized")

    def _setup_routes(self):
        """Setup HTTP routes"""
        # Health endpoints
        self.app.router.add_get("/health", self.handle_health)
        self.app.router.add_get("/ready", self.handle_ready)

        # MCP JSON-RPC 2.0 endpoints (proper MCP protocol)
        self.app.router.add_post("/mcp", self.handle_jsonrpc)

        # Legacy MCP tool endpoints
        self.app.router.add_post("/mcp/tool", self.handle_mcp_tool)

        # Direct tool endpoints for convenience
        self.app.router.add_post("/tools/prometheus_query", self.handle_prometheus_query)
        self.app.router.add_post("/tools/prometheus_query_range", self.handle_prometheus_query_range)
        self.app.router.add_post("/tools/grafana_query", self.handle_grafana_query)

        # List available tools
        self.app.router.add_get("/tools", self.handle_list_tools)

    async def handle_health(self, request: web.Request) -> web.Response:
        """❤️ Health check endpoint"""
        return web.json_response(
            {"status": "healthy", "service": "agent-sre-mcp-http-wrapper", "timestamp": datetime.now().isoformat()}
        )

    async def handle_ready(self, request: web.Request) -> web.Response:
        """✅ Readiness check endpoint"""
        return web.json_response(
            {"status": "ready", "service": "agent-sre-mcp-http-wrapper", "timestamp": datetime.now().isoformat()}
        )

    async def handle_jsonrpc(self, request: web.Request) -> web.Response:
        """🔌 Handle JSON-RPC 2.0 requests (proper MCP protocol)"""
        try:
            data = await request.json()

            # Validate JSON-RPC 2.0 format
            if data.get("jsonrpc") != "2.0":
                return web.json_response(
                    {
                        "jsonrpc": "2.0",
                        "id": data.get("id"),
                        "error": {"code": -32600, "message": "Invalid Request: jsonrpc must be '2.0'"},
                    },
                    status=400,
                )

            method = data.get("method")
            params = data.get("params", {})
            request_id = data.get("id")

            logger.info(f"🔌 JSON-RPC method: {method}")

            # Route to appropriate handler
            if method == "tools/list":
                result = await self._handle_tools_list(params)
            elif method == "tools/call":
                result = await self._handle_tools_call(params)
            else:
                return web.json_response(
                    {
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "error": {"code": -32601, "message": f"Method not found: {method}"},
                    },
                    status=404,
                )

            # Return JSON-RPC 2.0 response
            return web.json_response({"jsonrpc": "2.0", "id": request_id, "result": result})

        except Exception as e:
            logger.error(f"❌ Error handling JSON-RPC request: {e}", exc_info=True)
            return web.json_response(
                {
                    "jsonrpc": "2.0",
                    "id": data.get("id") if "data" in locals() else None,
                    "error": {"code": -32603, "message": f"Internal error: {str(e)}"},
                },
                status=500,
            )

    async def _handle_tools_list(self, params: Dict[str, Any]) -> Dict[str, Any]:
        """📋 Handle tools/list request"""
        logger.info("📋 Listing available MCP tools")

        # Get tools from the MCP server
        tools_list = await mcp_list_tools()

        # Convert MCP Tool objects to dict format
        tools_dict = []
        for tool in tools_list:
            tools_dict.append({"name": tool.name, "description": tool.description, "inputSchema": tool.inputSchema})

        # Support pagination (cursor parameter)
        cursor = params.get("cursor")

        return {
            "tools": tools_dict,
            # In a real implementation, you'd implement pagination
            # For now, we return all tools at once
            "nextCursor": None,
        }

    async def _handle_tools_call(self, params: Dict[str, Any]) -> Dict[str, Any]:
        """🔧 Handle tools/call request"""
        tool_name = params.get("name")
        arguments = params.get("arguments", {})

        if not tool_name:
            raise ValueError("Tool name is required")

        logger.info(f"🔧 Calling tool: {tool_name} with arguments: {arguments}")

        # Route to the appropriate tool handler
        if tool_name == "prometheus_query":
            result = await execute_prometheus_query(arguments)
        elif tool_name == "prometheus_query_range":
            result = await execute_prometheus_query_range(arguments)
        elif tool_name == "grafana_query":
            result = await execute_grafana_query(arguments)
        else:
            return {"content": [{"type": "text", "text": f"Unknown tool: {tool_name}"}], "isError": True}

        # Check if the tool execution had an error
        is_error = result.get("status") == "error"

        # Format result as MCP tool result
        return {"content": [{"type": "text", "text": json.dumps(result, indent=2, default=str)}], "isError": is_error}

    async def handle_list_tools(self, request: web.Request) -> web.Response:
        """📋 List available MCP tools (legacy endpoint)"""
        tools = [
            {
                "name": "prometheus_query",
                "description": "Execute a PromQL query against Prometheus",
                "required_params": ["query"],
                "optional_params": ["time", "timeout"],
            },
            {
                "name": "prometheus_query_range",
                "description": "Execute a range query against Prometheus",
                "required_params": ["query", "start", "end", "step"],
                "optional_params": ["timeout"],
            },
            {
                "name": "grafana_query",
                "description": "Query Grafana dashboards or datasources",
                "required_params": ["query_type", "query"],
                "optional_params": ["dashboard_id", "panel_id", "from_time", "to_time"],
            },
        ]

        return web.json_response({"tools": tools, "count": len(tools), "timestamp": datetime.now().isoformat()})

    async def handle_mcp_tool(self, request: web.Request) -> web.Response:
        """🔧 Generic MCP tool handler"""
        try:
            data = await request.json()
            tool_name = data.get("tool")
            arguments = data.get("arguments", {})

            if not tool_name:
                return web.json_response({"error": "Tool name is required"}, status=400)

            logger.info(f"🔧 Calling tool: {tool_name}")

            # Route to the appropriate tool handler
            if tool_name == "prometheus_query":
                result = await execute_prometheus_query(arguments)
            elif tool_name == "prometheus_query_range":
                result = await execute_prometheus_query_range(arguments)
            elif tool_name == "grafana_query":
                result = await execute_grafana_query(arguments)
            else:
                return web.json_response({"error": f"Unknown tool: {tool_name}"}, status=400)

            return web.json_response(result)

        except Exception as e:
            logger.error(f"❌ Error handling MCP tool: {e}", exc_info=True)
            return web.json_response({"error": str(e), "timestamp": datetime.now().isoformat()}, status=500)

    async def handle_prometheus_query(self, request: web.Request) -> web.Response:
        """🔍 Prometheus query endpoint"""
        try:
            arguments = await request.json()
            result = await execute_prometheus_query(arguments)
            return web.json_response(result)
        except Exception as e:
            logger.error(f"❌ Error in prometheus_query: {e}", exc_info=True)
            return web.json_response({"error": str(e), "timestamp": datetime.now().isoformat()}, status=500)

    async def handle_prometheus_query_range(self, request: web.Request) -> web.Response:
        """📈 Prometheus range query endpoint"""
        try:
            arguments = await request.json()
            result = await execute_prometheus_query_range(arguments)
            return web.json_response(result)
        except Exception as e:
            logger.error(f"❌ Error in prometheus_query_range: {e}", exc_info=True)
            return web.json_response({"error": str(e), "timestamp": datetime.now().isoformat()}, status=500)

    async def handle_grafana_query(self, request: web.Request) -> web.Response:
        """📊 Grafana query endpoint"""
        try:
            arguments = await request.json()
            result = await execute_grafana_query(arguments)
            return web.json_response(result)
        except Exception as e:
            logger.error(f"❌ Error in grafana_query: {e}", exc_info=True)
            return web.json_response({"error": str(e), "timestamp": datetime.now().isoformat()}, status=500)

    async def start_server(self, host: str = "0.0.0.0", port: int = 3000):
        """🚀 Start the HTTP wrapper server"""
        runner = web.AppRunner(self.app)
        await runner.setup()
        site = web.TCPSite(runner, host, port)
        await site.start()

        logger.info(f"🌐 MCP HTTP Wrapper started on {host}:{port}")
        logger.info(f"🏥 Health endpoint: http://localhost:{port}/health")
        logger.info(f"✅ Ready endpoint: http://localhost:{port}/ready")
        logger.info(f"📋 Tools list: http://localhost:{port}/tools")
        logger.info(f"🔧 MCP tool endpoint: http://localhost:{port}/mcp/tool")

        return runner


async def main():
    """🚀 Main entry point"""
    logger.info("🚀 Starting Agent-SRE MCP HTTP Wrapper")

    host = os.getenv("HTTP_HOST", "0.0.0.0")
    port = int(os.getenv("HTTP_PORT", "3000"))

    wrapper = MCPHTTPWrapper()
    runner = await wrapper.start_server(host, port)

    try:
        logger.info("🏁 MCP HTTP Wrapper is running...")
        await asyncio.Event().wait()  # Run forever
    except KeyboardInterrupt:
        logger.info("🛑 Shutting down MCP HTTP Wrapper...")
    finally:
        await runner.cleanup()


if __name__ == "__main__":
    asyncio.run(main())
