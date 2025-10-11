#!/usr/bin/env python3
"""
SRE Agent - Standalone Agent Service
Handles HTTP API requests and communicates with MCP server
"""

import asyncio
import logging
import os
from datetime import datetime
from typing import Any, Dict

from aiohttp import ClientSession, web
from aiohttp.web import Request, Response

# Import the SRE agent from local core
from core import agent, logger


class SREAgentService:
    """Standalone SRE Agent Service."""

    def __init__(self):
        self.sre_agent = agent
        # Create app with custom logging to filter out health/ready check noise
        self.app = web.Application(middlewares=[self._logging_middleware])

        # Configure service URLs from environment variables
        self.prometheus_url = os.getenv(
            "PROMETHEUS_URL", "http://prometheus-operator-kube-p-prometheus.prometheus.svc.cluster.local:9090"
        )
        self.mcp_server_url = os.getenv("MCP_SERVER_URL", "http://sre-agent-mcp-server-service:30120")

        # Log configuration
        logger.info("🔧 SRE Agent Configuration:")
        logger.info(f"  📊 Prometheus URL: {self.prometheus_url}")
        logger.info(f"  🔌 MCP Server URL: {self.mcp_server_url}")

        self._setup_routes()

    @web.middleware
    async def _logging_middleware(self, request: Request, handler):
        """Middleware to log all requests except health checks"""
        path = request.path
        method = request.method

        # Skip logging for health and ready endpoints
        if path not in ["/health", "/ready"]:
            logger.info(f"📥 {method} {path} - Client: {request.remote}")
            logger.debug(f"📥 Headers: {dict(request.headers)}")

        try:
            response = await handler(request)

            # Skip logging for health and ready endpoints
            if path not in ["/health", "/ready"]:
                logger.info(f"📤 {method} {path} - Status: {response.status}")

            return response
        except Exception as e:
            logger.error(f"❌ {method} {path} - Error: {e}", exc_info=True)
            raise

    def _setup_routes(self):
        """Setup HTTP routes."""
        # Health and readiness endpoints
        self.app.router.add_get("/health", self.handle_health)
        self.app.router.add_get("/ready", self.handle_readiness)

        # Status and info endpoints
        self.app.router.add_get("/status", self.handle_status)

        # Agent API endpoints
        self.app.router.add_post("/ping", self.handle_ping)
        self.app.router.add_post("/chat", self.handle_chat)

        # MCP server communication endpoints
        self.app.router.add_post("/mcp/chat", self.handle_mcp_chat)

        # 🔍 Query endpoints
        self.app.router.add_post("/prometheus/query", self.handle_prometheus_query)
        self.app.router.add_post("/grafana/query", self.handle_grafana_query)
        self.app.router.add_post("/k8s/query", self.handle_k8s_query)

        # 🚨 Alertmanager webhook endpoint
        self.app.router.add_post("/webhook/alert", self.handle_alertmanager_webhook)

    async def handle_health(self, request: Request) -> Response:
        """Liveness probe endpoint."""
        try:
            health_status = {
                "status": "healthy",
                "service": "sre-agent",
                "timestamp": datetime.now().isoformat(),
                "uptime": "running",
                "version": "1.0.0",
                "deployment": "standalone-agent",
            }
            # Only log health checks in debug mode
            if os.getenv("DEBUG", "false").lower() == "true":
                logger.debug(f"Health check: {health_status}")
            return web.json_response(health_status)
        except Exception as e:
            logger.error(f"Health check failed: {e}")
            return web.json_response({"status": "unhealthy", "error": str(e)}, status=503)

    async def handle_readiness(self, request: Request) -> Response:
        """Readiness probe endpoint."""
        try:
            # Check if the SRE agent is properly initialized
            if not self.sre_agent or not self.sre_agent.llm:
                return web.json_response({"status": "not_ready", "reason": "SRE agent not initialized"}, status=503)

            # Check MCP server connectivity
            mcp_status = await self._check_mcp_server()

            readiness_status = {
                "status": "ready",
                "service": "sre-agent",
                "timestamp": datetime.now().isoformat(),
                "mcp_server_status": mcp_status,
                "deployment": "standalone-agent",
            }
            # Only log readiness checks in debug mode
            if os.getenv("DEBUG", "false").lower() == "true":
                logger.debug(f"Readiness check: {readiness_status}")
            return web.json_response(readiness_status)

        except Exception as e:
            logger.error(f"Readiness check failed: {e}")
            return web.json_response({"status": "not_ready", "error": str(e)}, status=503)

    async def handle_chat(self, request: Request) -> Response:
        """Direct chat endpoint using local agent."""
        try:
            data = await request.json()
            message = data.get("message", "")

            if not message:
                return web.json_response({"error": "Message is required"}, status=400)

            response = await self.sre_agent.chat(message)
            return web.json_response(
                {
                    "response": response,
                    "service": "sre-agent",
                    "timestamp": datetime.now().isoformat(),
                    "method": "direct",
                }
            )

        except Exception as e:
            logger.error(f"Error in chat handler: {e}")
            return web.json_response({"error": str(e)}, status=500)

    async def handle_mcp_chat(self, request: Request) -> Response:
        """Chat endpoint via MCP server."""
        try:
            data = await request.json()
            message = data.get("message", "")

            if not message:
                return web.json_response({"error": "Message is required"}, status=400)

            result = await self._call_mcp_tool("sre_chat", {"message": message})
            return web.json_response(
                {"response": result, "service": "sre-agent", "timestamp": datetime.now().isoformat(), "method": "mcp"}
            )

        except Exception as e:
            logger.error(f"Error in MCP chat handler: {e}")
            return web.json_response({"error": str(e)}, status=500)

    async def handle_ping(self, request: Request) -> Response:
        """Ping endpoint for service connectivity check."""
        try:
            return web.json_response(
                {"message": "pong", "service": "sre-agent", "timestamp": datetime.now().isoformat(), "status": "alive"}
            )

        except Exception as e:
            logger.error(f"Error in ping handler: {e}")
            return web.json_response({"error": str(e)}, status=500)

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
                "deployment": "standalone-agent",
            }
            return web.json_response(status)

        except Exception as e:
            logger.error(f"Error in status handler: {e}")
            return web.json_response({"error": str(e)}, status=500)

    async def handle_alertmanager_webhook(self, request: Request) -> Response:
        """🚨 Alertmanager webhook handler - receives alerts from Alertmanager."""
        try:
            data = await request.json()

            # Alertmanager sends alerts in this format
            alerts = data.get("alerts", [])

            if not alerts:
                logger.warning("⚠️  No alerts in webhook payload")
                return web.json_response({"message": "No alerts to process"}, status=200)

            # Process each alert
            results = []
            for alert in alerts:
                alert_name = alert.get("labels", {}).get("alertname", "Unknown")
                severity = alert.get("labels", {}).get("severity", "unknown")
                status = alert.get("status", "unknown")  # firing or resolved

                logger.info(f"🔔 Received alert: {alert_name} (severity={severity}, status={status})")

                # Only process firing alerts (skip resolved for now)
                if status == "firing":
                    # Build investigation context
                    annotations = alert.get("annotations", {})
                    labels = alert.get("labels", {})

                    investigation_message = f"""
🚨 ALERT INVESTIGATION REQUIRED

Alert: {alert_name}
Severity: {severity}
Status: {status}

Labels:
{chr(10).join([f"  - {k}: {v}" for k, v in labels.items()])}

Annotations:
{chr(10).join([f"  - {k}: {v}" for k, v in annotations.items()])}

Starts At: {alert.get("startsAt", "N/A")}
Generator URL: {alert.get("generatorURL", "N/A")}

Please provide:
1. Root cause analysis
2. Impact assessment  
3. Immediate mitigation steps
4. Recommended investigation queries (PromQL, LogQL)
5. Prevention recommendations
                    """.strip()

                    # Execute incident analysis using the local agent
                    analysis = await self.sre_agent.incident_response(investigation_message)

                    results.append(
                        {
                            "alert_name": alert_name,
                            "severity": severity,
                            "analysis": analysis,
                            "fingerprint": alert.get("fingerprint", ""),
                        }
                    )

                    logger.info(f"✅ Completed investigation for alert: {alert_name}")

            return web.json_response(
                {
                    "message": f"Processed {len(results)} alerts",
                    "results": results,
                    "service": "sre-agent",
                    "timestamp": datetime.now().isoformat(),
                }
            )

        except Exception as e:
            logger.error(f"❌ Error in alertmanager webhook handler: {e}", exc_info=True)
            return web.json_response({"error": str(e)}, status=500)

    async def handle_prometheus_query(self, request: Request) -> Response:
        """🔍 Execute a PromQL query"""
        try:
            data = await request.json()
            query = data.get("query", "")

            if not query:
                return web.json_response({"error": "Query is required"}, status=400)

            logger.info(f"🔍 Executing Prometheus query: {query}")
            prom_url = f"{self.prometheus_url}/api/v1/query"
            logger.debug(f"  📡 Prometheus URL: {prom_url}")

            async with ClientSession() as session:
                async with session.get(prom_url, params={"query": query}, timeout=10) as response:
                    logger.info(f"  📊 Prometheus responded: {response.status}")

                    if response.status == 200:
                        result = await response.json()
                        logger.info(
                            f"  ✅ Query successful - returned {len(result.get('data', {}).get('result', []))} results"
                        )
                        return web.json_response(
                            {"query": query, "result": result, "timestamp": datetime.now().isoformat()}
                        )
                    else:
                        error_text = await response.text()
                        logger.error(f"  ❌ Prometheus error ({response.status}): {error_text}")
                        return web.json_response({"error": f"Prometheus error: {error_text}"}, status=response.status)

        except Exception as e:
            logger.error(f"❌ Error executing Prometheus query: {e}", exc_info=True)
            logger.error(f"  🔗 Prometheus URL was: {self.prometheus_url}")
            return web.json_response({"error": str(e), "prometheus_url": self.prometheus_url}, status=500)

    async def handle_grafana_query(self, request: Request) -> Response:
        """📊 Execute a Grafana query"""
        try:
            data = await request.json()
            query = data.get("query", "")
            dashboard_id = data.get("dashboard_id")
            panel_id = data.get("panel_id")

            if not query:
                return web.json_response({"error": "Query is required"}, status=400)

            logger.info(f"📊 Executing Grafana query: {query}")

            # TODO: Implement actual Grafana API call
            # For now, return a placeholder
            return web.json_response(
                {
                    "query": query,
                    "dashboard_id": dashboard_id,
                    "panel_id": panel_id,
                    "result": "Grafana query not yet implemented. Use Grafana UI directly.",
                    "timestamp": datetime.now().isoformat(),
                }
            )

        except Exception as e:
            logger.error(f"❌ Error executing Grafana query: {e}")
            return web.json_response({"error": str(e)}, status=500)

    async def handle_k8s_query(self, request: Request) -> Response:
        """☸️ Execute a Kubernetes query"""
        try:
            data = await request.json()
            query = data.get("query", "")

            if not query:
                return web.json_response({"error": "Query is required"}, status=400)

            logger.info(f"☸️ Executing Kubernetes query: {query}")

            # Use the SRE agent to handle the Kubernetes query
            response = await self.sre_agent.chat(query)

            return web.json_response({"query": query, "result": response, "timestamp": datetime.now().isoformat()})

        except Exception as e:
            logger.error(f"❌ Error executing Kubernetes query: {e}")
            return web.json_response({"error": str(e)}, status=500)

    async def _check_mcp_server(self) -> Dict[str, Any]:
        """Check MCP server connectivity."""
        try:
            async with ClientSession() as session:
                async with session.get(f"{self.mcp_server_url}/health", timeout=5) as response:
                    if response.status == 200:
                        return {"status": "connected", "url": self.mcp_server_url, "response_time": "ok"}
                    else:
                        return {"status": "error", "url": self.mcp_server_url, "error": f"HTTP {response.status}"}

        except Exception as e:
            return {"status": "disconnected", "url": self.mcp_server_url, "error": str(e)}

    async def _call_mcp_tool(self, tool_name: str, arguments: Dict[str, Any]) -> str:
        """Call MCP server tool."""
        try:
            async with ClientSession() as session:
                mcp_request = {
                    "jsonrpc": "2.0",
                    "id": 1,
                    "method": "tools/call",
                    "params": {"name": tool_name, "arguments": arguments},
                }

                async with session.post(f"{self.mcp_server_url}/mcp", json=mcp_request, timeout=30) as response:
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
        # Start server with default access logging
        runner = web.AppRunner(self.app)
        await runner.setup()
        site = web.TCPSite(runner, host, port)
        await site.start()

        logger.info(f"🌐 SRE Agent started on {host}:{port}")

        # Apply health check filter to aiohttp access logger AFTER server starts
        try:

            class HealthCheckFilter(logging.Filter):
                """Filter out health and ready endpoint access logs"""

                def filter(self, record):
                    message = record.getMessage()
                    should_keep = not ("/health" in message or "/ready" in message)
                    return should_keep

            access_logger = logging.getLogger("aiohttp.access")
            health_filter = HealthCheckFilter()
            access_logger.addFilter(health_filter)
            # Also add to all handlers
            handler_count = 0
            for handler in access_logger.handlers:
                handler.addFilter(health_filter)
                handler_count += 1

            logger.info(f"🔇 Applied HealthCheckFilter to aiohttp.access logger ({handler_count} handlers)")
        except Exception as e:
            logger.error(f"❌ Failed to apply HealthCheckFilter: {e}", exc_info=True)
        logger.info(f"🏥 Health endpoint: http://localhost:{port}/health")
        logger.info(f"✅ Readiness endpoint: http://localhost:{port}/ready")
        logger.info(f"💬 Chat endpoint: http://localhost:{port}/chat")
        logger.info(f"📊 MCP Chat endpoint: http://localhost:{port}/mcp/chat")
        logger.info(f"📈 Status endpoint: http://localhost:{port}/status")
        logger.info(f"🚨 Alertmanager webhook: http://localhost:{port}/webhook/alert")
        logger.info(f"🔇 Health/ready check logs filtered for cleaner output")

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
