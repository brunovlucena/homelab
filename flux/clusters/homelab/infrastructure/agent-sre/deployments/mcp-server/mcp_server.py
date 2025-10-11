#!/usr/bin/env python3
"""
🔌 Agent-SRE MCP Server
Exposes Prometheus and Grafana query tools via MCP protocol
"""

import asyncio
import json
import logging
import os
from datetime import datetime
from typing import Any, Dict, List, Optional

import aiohttp
from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.types import EmbeddedResource, ImageContent, TextContent, Tool

# Configure logging
logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s")
logger = logging.getLogger(__name__)

# Configuration
PROMETHEUS_URL = os.getenv(
    "PROMETHEUS_URL", "http://prometheus-operator-kube-p-prometheus.prometheus.svc.cluster.local:9090"
)
GRAFANA_URL = os.getenv("GRAFANA_URL", "http://prometheus-operator-grafana.prometheus.svc.cluster.local:80")
GRAFANA_API_KEY = os.getenv("GRAFANA_API_KEY", "")

# Create MCP server instance
mcp_server = Server("agent-sre-mcp-server")


@mcp_server.list_tools()
async def list_tools() -> List[Tool]:
    """📋 List available MCP tools"""
    return [
        Tool(
            name="prometheus_query",
            description="""🔍 Execute a PromQL query against Prometheus.
            
Use this tool to query metrics from Prometheus. You can query:
- Current values: rate(http_requests_total[5m])
- Time series data: node_memory_usage_bytes
- Aggregations: sum(rate(container_cpu_usage_seconds_total[5m])) by (namespace)
- Alerts: ALERTS{alertname="HighMemoryUsage"}

The query should be a valid PromQL expression.
            """,
            inputSchema={
                "type": "object",
                "properties": {
                    "query": {
                        "type": "string",
                        "description": "PromQL query to execute (e.g., 'up', 'rate(http_requests_total[5m])')",
                    },
                    "time": {
                        "type": "string",
                        "description": "Optional RFC3339 or Unix timestamp for query evaluation",
                    },
                    "timeout": {"type": "string", "description": "Optional timeout for the query (e.g., '30s')"},
                },
                "required": ["query"],
            },
        ),
        Tool(
            name="grafana_query",
            description="""📊 Query data from Grafana dashboards or datasources.
            
Use this tool to:
- Get dashboard information by UID or ID
- Query datasources directly
- Retrieve panel data
- Search dashboards

This is useful for getting visualization data and dashboard states.
            """,
            inputSchema={
                "type": "object",
                "properties": {
                    "query_type": {
                        "type": "string",
                        "description": "Type of query: 'dashboard', 'datasource', 'search', 'panel'",
                        "enum": ["dashboard", "datasource", "search", "panel"],
                    },
                    "query": {
                        "type": "string",
                        "description": "Query string or identifier (dashboard UID, search term, etc.)",
                    },
                    "dashboard_id": {"type": "string", "description": "Optional dashboard UID for panel queries"},
                    "panel_id": {"type": "integer", "description": "Optional panel ID within the dashboard"},
                    "from_time": {"type": "string", "description": "Optional start time (RFC3339 or Unix timestamp)"},
                    "to_time": {"type": "string", "description": "Optional end time (RFC3339 or Unix timestamp)"},
                },
                "required": ["query_type", "query"],
            },
        ),
        Tool(
            name="prometheus_query_range",
            description="""📈 Execute a range query against Prometheus to get time series data.
            
Use this tool to query metrics over a time range:
- Memory usage over the last hour
- CPU trends for the past day
- Request rate patterns over time

Returns time series data with timestamps and values.
            """,
            inputSchema={
                "type": "object",
                "properties": {
                    "query": {"type": "string", "description": "PromQL query to execute"},
                    "start": {"type": "string", "description": "Start timestamp (RFC3339 or Unix timestamp)"},
                    "end": {"type": "string", "description": "End timestamp (RFC3339 or Unix timestamp)"},
                    "step": {"type": "string", "description": "Query resolution step width (e.g., '15s', '1m', '5m')"},
                    "timeout": {"type": "string", "description": "Optional timeout for the query (e.g., '30s')"},
                },
                "required": ["query", "start", "end", "step"],
            },
        ),
    ]


@mcp_server.call_tool()
async def call_tool(name: str, arguments: Any) -> List[TextContent]:
    """🔧 Execute an MCP tool"""

    logger.info(f"🔧 Tool called: {name} with arguments: {arguments}")

    try:
        if name == "prometheus_query":
            result = await execute_prometheus_query(arguments)
        elif name == "prometheus_query_range":
            result = await execute_prometheus_query_range(arguments)
        elif name == "grafana_query":
            result = await execute_grafana_query(arguments)
        else:
            result = {"error": f"Unknown tool: {name}"}

        # Format the result as MCP TextContent
        return [TextContent(type="text", text=json.dumps(result, indent=2, default=str))]

    except Exception as e:
        logger.error(f"❌ Error executing tool {name}: {e}", exc_info=True)
        return [
            TextContent(
                type="text",
                text=json.dumps({"error": str(e), "tool": name, "timestamp": datetime.now().isoformat()}, indent=2),
            )
        ]


async def execute_prometheus_query(arguments: Dict[str, Any]) -> Dict[str, Any]:
    """🔍 Execute a Prometheus instant query"""

    query = arguments.get("query", "")
    if not query:
        return {"error": "Query parameter is required"}

    params = {"query": query}

    # Optional parameters
    if "time" in arguments:
        params["time"] = arguments["time"]
    if "timeout" in arguments:
        params["timeout"] = arguments["timeout"]

    logger.info(f"🔍 Executing Prometheus query: {query}")

    async with aiohttp.ClientSession() as session:
        try:
            async with session.get(
                f"{PROMETHEUS_URL}/api/v1/query", params=params, timeout=aiohttp.ClientTimeout(total=30)
            ) as response:
                if response.status == 200:
                    data = await response.json()

                    return {
                        "status": "success",
                        "query": query,
                        "result": data.get("data", {}),
                        "timestamp": datetime.now().isoformat(),
                        "prometheus_url": PROMETHEUS_URL,
                    }
                else:
                    error_text = await response.text()
                    logger.error(f"❌ Prometheus error: {response.status} - {error_text}")
                    return {
                        "status": "error",
                        "query": query,
                        "error": f"HTTP {response.status}: {error_text}",
                        "timestamp": datetime.now().isoformat(),
                    }

        except asyncio.TimeoutError:
            logger.error(f"❌ Prometheus query timeout: {query}")
            return {
                "status": "error",
                "query": query,
                "error": "Query timeout",
                "timestamp": datetime.now().isoformat(),
            }
        except Exception as e:
            logger.error(f"❌ Prometheus query error: {e}", exc_info=True)
            return {"status": "error", "query": query, "error": str(e), "timestamp": datetime.now().isoformat()}


async def execute_prometheus_query_range(arguments: Dict[str, Any]) -> Dict[str, Any]:
    """📈 Execute a Prometheus range query"""

    query = arguments.get("query", "")
    start = arguments.get("start", "")
    end = arguments.get("end", "")
    step = arguments.get("step", "15s")

    if not query or not start or not end:
        return {"error": "Query, start, and end parameters are required"}

    params = {"query": query, "start": start, "end": end, "step": step}

    # Optional parameters
    if "timeout" in arguments:
        params["timeout"] = arguments["timeout"]

    logger.info(f"📈 Executing Prometheus range query: {query} from {start} to {end}")

    async with aiohttp.ClientSession() as session:
        try:
            async with session.get(
                f"{PROMETHEUS_URL}/api/v1/query_range", params=params, timeout=aiohttp.ClientTimeout(total=60)
            ) as response:
                if response.status == 200:
                    data = await response.json()

                    return {
                        "status": "success",
                        "query": query,
                        "start": start,
                        "end": end,
                        "step": step,
                        "result": data.get("data", {}),
                        "timestamp": datetime.now().isoformat(),
                        "prometheus_url": PROMETHEUS_URL,
                    }
                else:
                    error_text = await response.text()
                    logger.error(f"❌ Prometheus range query error: {response.status} - {error_text}")
                    return {
                        "status": "error",
                        "query": query,
                        "error": f"HTTP {response.status}: {error_text}",
                        "timestamp": datetime.now().isoformat(),
                    }

        except asyncio.TimeoutError:
            logger.error(f"❌ Prometheus range query timeout: {query}")
            return {
                "status": "error",
                "query": query,
                "error": "Query timeout",
                "timestamp": datetime.now().isoformat(),
            }
        except Exception as e:
            logger.error(f"❌ Prometheus range query error: {e}", exc_info=True)
            return {"status": "error", "query": query, "error": str(e), "timestamp": datetime.now().isoformat()}


async def execute_grafana_query(arguments: Dict[str, Any]) -> Dict[str, Any]:
    """📊 Execute a Grafana query"""

    query_type = arguments.get("query_type", "")
    query = arguments.get("query", "")

    if not query_type or not query:
        return {"error": "query_type and query parameters are required"}

    logger.info(f"📊 Executing Grafana query: type={query_type}, query={query}")

    # Build headers
    headers = {}
    if GRAFANA_API_KEY:
        headers["Authorization"] = f"Bearer {GRAFANA_API_KEY}"

    async with aiohttp.ClientSession() as session:
        try:
            # Handle different query types
            if query_type == "dashboard":
                url = f"{GRAFANA_URL}/api/dashboards/uid/{query}"
            elif query_type == "search":
                url = f"{GRAFANA_URL}/api/search?query={query}"
            elif query_type == "datasource":
                url = f"{GRAFANA_URL}/api/datasources/name/{query}"
            elif query_type == "panel":
                dashboard_id = arguments.get("dashboard_id", "")
                panel_id = arguments.get("panel_id", 0)
                if not dashboard_id:
                    return {"error": "dashboard_id required for panel queries"}
                url = f"{GRAFANA_URL}/api/dashboards/uid/{dashboard_id}"
            else:
                return {"error": f"Unknown query type: {query_type}"}

            logger.info(f"📊 Grafana URL: {url}")

            async with session.get(url, headers=headers, timeout=aiohttp.ClientTimeout(total=30)) as response:
                if response.status == 200:
                    data = await response.json()

                    return {
                        "status": "success",
                        "query_type": query_type,
                        "query": query,
                        "result": data,
                        "timestamp": datetime.now().isoformat(),
                        "grafana_url": GRAFANA_URL,
                    }
                elif response.status == 404:
                    return {
                        "status": "not_found",
                        "query_type": query_type,
                        "query": query,
                        "error": f"Resource not found: {query}",
                        "timestamp": datetime.now().isoformat(),
                    }
                else:
                    error_text = await response.text()
                    logger.error(f"❌ Grafana error: {response.status} - {error_text}")
                    return {
                        "status": "error",
                        "query_type": query_type,
                        "query": query,
                        "error": f"HTTP {response.status}: {error_text}",
                        "timestamp": datetime.now().isoformat(),
                    }

        except asyncio.TimeoutError:
            logger.error(f"❌ Grafana query timeout: {query}")
            return {
                "status": "error",
                "query_type": query_type,
                "query": query,
                "error": "Query timeout",
                "timestamp": datetime.now().isoformat(),
            }
        except Exception as e:
            logger.error(f"❌ Grafana query error: {e}", exc_info=True)
            return {
                "status": "error",
                "query_type": query_type,
                "query": query,
                "error": str(e),
                "timestamp": datetime.now().isoformat(),
            }


async def main():
    """🚀 Main entry point for MCP server"""
    logger.info("🚀 Starting Agent-SRE MCP Server")
    logger.info(f"📊 Prometheus URL: {PROMETHEUS_URL}")
    logger.info(f"📈 Grafana URL: {GRAFANA_URL}")
    logger.info(f"🔐 Grafana API Key: {'configured' if GRAFANA_API_KEY else 'not configured'}")

    async with stdio_server() as (read_stream, write_stream):
        await mcp_server.run(read_stream, write_stream, mcp_server.create_initialization_options())


if __name__ == "__main__":
    asyncio.run(main())
