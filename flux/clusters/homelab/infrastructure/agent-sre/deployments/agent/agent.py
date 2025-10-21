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
from investigation_workflow import investigation_workflow
from github_integration import github_client


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
        self.app.router.add_post("/analyze-logs", self.handle_analyze_logs)
        self.app.router.add_post("/incident-response", self.handle_incident_response)
        self.app.router.add_post("/monitoring-advice", self.handle_monitoring_advice)

        # MCP server communication endpoints
        self.app.router.add_post("/mcp/chat", self.handle_mcp_chat)
        self.app.router.add_get("/mcp/status", self.handle_mcp_status)

        # 🔍 Query endpoints
        self.app.router.add_post("/prometheus/query", self.handle_prometheus_query)
        self.app.router.add_post("/grafana/query", self.handle_grafana_query)
        self.app.router.add_post("/k8s/query", self.handle_k8s_query)

        # 🚨 Alertmanager webhook endpoint
        self.app.router.add_post("/webhook/alert", self.handle_alertmanager_webhook)

        # 🔍 Investigation endpoints
        self.app.router.add_post("/investigation/create", self.handle_create_investigation)
        self.app.router.add_post("/investigation/workflow-failure", self.handle_workflow_failure)
        self.app.router.add_get("/investigation/{investigation_id}", self.handle_get_investigation)

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

    async def handle_analyze_logs(self, request: Request) -> Response:
        """Analyze logs endpoint."""
        try:
            data = await request.json()
            logs = data.get("logs", "")

            if not logs:
                return web.json_response({"error": "Logs are required"}, status=400)

            analysis = await self.sre_agent.analyze_logs(logs)
            return web.json_response(
                {
                    "analysis": analysis,
                    "service": "sre-agent",
                    "timestamp": datetime.now().isoformat(),
                    "method": "direct",
                }
            )

        except Exception as e:
            logger.error(f"Error in analyze logs handler: {e}")
            return web.json_response({"error": str(e)}, status=500)

    async def handle_incident_response(self, request: Request) -> Response:
        """Incident response endpoint."""
        try:
            data = await request.json()
            incident = data.get("incident", "")

            if not incident:
                return web.json_response({"error": "Incident description is required"}, status=400)

            response = await self.sre_agent.incident_response(incident)
            return web.json_response(
                {
                    "response": response,
                    "service": "sre-agent",
                    "timestamp": datetime.now().isoformat(),
                }
            )

        except Exception as e:
            logger.error(f"Error in incident response handler: {e}")
            return web.json_response({"error": str(e)}, status=500)

    async def handle_monitoring_advice(self, request: Request) -> Response:
        """Monitoring advice endpoint."""
        try:
            data = await request.json()
            system = data.get("system", "")

            if not system:
                return web.json_response({"error": "System description is required"}, status=400)

            advice = await self.sre_agent.monitoring_advice(system)
            return web.json_response(
                {
                    "advice": advice,
                    "service": "sre-agent",
                    "timestamp": datetime.now().isoformat(),
                }
            )

        except Exception as e:
            logger.error(f"Error in monitoring advice handler: {e}")
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

    async def handle_mcp_status(self, request: Request) -> Response:
        """MCP server status endpoint."""
        try:
            mcp_status = await self._check_mcp_server()
            return web.json_response(
                {
                    "status": mcp_status,
                    "service": "sre-agent",
                    "timestamp": datetime.now().isoformat(),
                }
            )

        except Exception as e:
            logger.error(f"Error in MCP status handler: {e}")
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
        """🚨 Alertmanager webhook handler - receives alerts and triggers investigations."""
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
                severity = alert.get("labels", {}).get("severity", "warning")
                status = alert.get("status", "unknown")  # firing or resolved
                namespace = alert.get("labels", {}).get("namespace", "unknown")
                pod = alert.get("labels", {}).get("pod", "")
                job = alert.get("labels", {}).get("job", "")

                logger.info(f"🔔 Received alert: {alert_name} (severity={severity}, status={status})")

                # Only process firing alerts (skip resolved)
                if status == "firing":
                    # Build investigation context
                    annotations = alert.get("annotations", {})
                    labels = alert.get("labels", {})

                    # Determine component from labels
                    component = namespace
                    if not component or component == "unknown":
                        # Try to extract from job or pod name
                        if job:
                            component = job.split("/")[0] if "/" in job else job
                        elif pod:
                            component = pod.split("-")[0]

                    # Build detailed description with all context
                    description = f"""## 🚨 Alert Triggered

**Alert Name**: {alert_name}
**Status**: {status}
**Severity**: {severity}
**Starts At**: {alert.get("startsAt", "N/A")}
**Generator URL**: {alert.get("generatorURL", "N/A")}
**Fingerprint**: {alert.get("fingerprint", "N/A")}

### Labels
"""
                    for key, value in labels.items():
                        description += f"- **{key}**: `{value}`\n"

                    description += "\n### Annotations\n"
                    for key, value in annotations.items():
                        description += f"- **{key}**: {value}\n"

                    description += """

### Investigation Tasks

This alert has triggered an automated investigation:

1. ✅ Analyzing error patterns in logs (via Grafana Loki)
2. ✅ Detecting slow requests in traces (via Grafana Tempo)
3. ✅ Running LLM-powered root cause analysis
4. ✅ Generating actionable recommendations

---

_This investigation was automatically triggered by Alertmanager._
"""

                    # Map Alertmanager severity to investigation severity
                    severity_map = {
                        "critical": "critical",
                        "warning": "high",
                        "info": "medium",
                    }
                    investigation_severity = severity_map.get(severity.lower(), "medium")

                    # Trigger automated investigation workflow
                    logger.info(f"🔍 Triggering investigation workflow for alert: {alert_name}")
                    
                    investigation_result = await investigation_workflow.investigate(
                        title=f"Alert: {alert_name}",
                        description=description,
                        severity=investigation_severity,
                        component=component,
                    )

                    results.append(
                        {
                            "alert_name": alert_name,
                            "severity": severity,
                            "investigation_id": investigation_result.get("investigation_id"),
                            "issue_number": investigation_result.get("issue_number"),
                            "issue_url": investigation_result.get("issue_url"),
                            "completed": investigation_result.get("completed", False),
                            "fingerprint": alert.get("fingerprint", ""),
                        }
                    )

                    logger.info(
                        f"✅ Investigation completed for alert: {alert_name} "
                        f"(Issue: {investigation_result.get('issue_url', 'N/A')})"
                    )

            return web.json_response(
                {
                    "message": f"Processed {len(results)} alerts",
                    "investigations": results,
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

    async def handle_create_investigation(self, request: Request) -> Response:
        """🔍 Create a new investigation with GitHub issue"""
        try:
            data = await request.json()
            title = data.get("title", "")
            description = data.get("description", "")
            severity = data.get("severity", "medium")
            component = data.get("component", "unknown")

            if not title:
                return web.json_response({"error": "Title is required"}, status=400)

            logger.info(f"🔍 Creating investigation: {title}")

            # Run the investigation workflow
            result = await investigation_workflow.investigate(
                title=title, description=description, severity=severity, component=component
            )

            return web.json_response(
                {
                    "investigation": result,
                    "service": "sre-agent",
                    "timestamp": datetime.now().isoformat(),
                }
            )

        except Exception as e:
            logger.error(f"❌ Error creating investigation: {e}", exc_info=True)
            return web.json_response({"error": str(e)}, status=500)

    async def handle_workflow_failure(self, request: Request) -> Response:
        """🚨 Handle workflow failure and auto-investigate"""
        try:
            data = await request.json()
            workflow_name = data.get("workflow_name", "Unknown Workflow")
            run_id = data.get("run_id")
            job_id = data.get("job_id")
            run_url = data.get("run_url", "")
            job_url = data.get("job_url", "")
            failure_details = data.get("failure_details", "")

            logger.info(f"🚨 Workflow failure detected: {workflow_name} (run: {run_id})")

            # Build investigation title and description
            title = f"CI/CD Workflow Failure - Run #{run_id}"
            description = f"""## Workflow Failure

**Workflow**: {workflow_name}
**Run ID**: {run_id}
**Job ID**: {job_id}

### Details

{failure_details}

### Links

* **Workflow Run**: {run_url}
* **Failed Job**: {job_url}

### Next Steps

1. Review the workflow logs at the link above
2. Identify the root cause of the failure
3. Implement a fix
4. Re-run the workflow to verify the fix

---

_Note: Please review the job logs and update this issue with specific error details._
"""

            # Run investigation workflow
            result = await investigation_workflow.investigate(
                title=title, description=description, severity="high", component="ci-cd"
            )

            logger.info(f"✅ Workflow failure investigation completed: {result.get('issue_url', 'N/A')}")

            return web.json_response(
                {
                    "investigation": result,
                    "service": "sre-agent",
                    "timestamp": datetime.now().isoformat(),
                }
            )

        except Exception as e:
            logger.error(f"❌ Error handling workflow failure: {e}", exc_info=True)
            return web.json_response({"error": str(e)}, status=500)

    async def handle_get_investigation(self, request: Request) -> Response:
        """🔍 Get investigation details"""
        try:
            investigation_id = request.match_info.get("investigation_id")

            if not investigation_id:
                return web.json_response({"error": "Investigation ID is required"}, status=400)

            logger.info(f"🔍 Getting investigation: {investigation_id}")

            # Call MCP server to get investigation details
            result = await self._call_mcp_tool(
                "sift_get_investigation", {"investigation_id": investigation_id}
            )

            return web.json_response(
                {
                    "investigation": result,
                    "service": "sre-agent",
                    "timestamp": datetime.now().isoformat(),
                }
            )

        except Exception as e:
            logger.error(f"❌ Error getting investigation: {e}", exc_info=True)
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
        logger.info(f"🔍 Investigation endpoint: http://localhost:{port}/investigation/create")
        logger.info("🔇 Health/ready check logs filtered for cleaner output")
        logger.info("✨ Alertmanager integration: Auto-investigations enabled!")

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
