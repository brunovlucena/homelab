# 📊 MCP Observability

> **Part of**: [AI Agent Architecture](ai-agent-architecture.md)  
> **Related**: [Agent Orchestration](agent-orchestration.md) | [Studio Cluster](../clusters/studio-cluster.md)  
> **Last Updated**: November 7, 2025

---

## Overview

This document describes the **Model Context Protocol (MCP) Server** as the foundation layer for observability and monitoring in the AI Agent architecture.

**Key Concept**: MCP provides structured access to observability data, enabling both AI agents (via natural language) and advanced users (via direct API) to interact with the monitoring stack.

---

## Architecture Pattern

### MCP as Foundation Layer

```
┌─────────────────────────────────────────────────────────────┐
│              MCP Observability Architecture                 │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Teams (Users)                          │    │
│  │  • Developers                                       │    │
│  │  • SREs                                             │    │
│  │  • Data Scientists                                  │    │
│  └──────────────────┬──────────────────────────────────┘    │
│                     │                                       │
│           ┌─────────┴─────────┐                             │
│           │                   │                             │
│  ┌────────▼────────┐  ┌───────▼────────┐                    │
│  │   AI Agents     │  │  Direct MCP    │                    │
│  │  (Natural Lang) │  │  (Structured)  │                    │
│  │                 │  │                │                    │
│  │ • agent-bruno   │  │ • CLI tools    │                    │
│  │ • agent-auditor │  │ • Scripts      │                    │
│  └────────┬────────┘  └───────┬────────┘                    │
│           │                   │                             │
│           └─────────┬─────────┘                             │
│                     │                                       │
│  ┌──────────────────▼──────────────────────────────────┐    │
│  │           MCP Server (Foundation)                   │    │
│  │  Service: mcp-observability.observability:8080      │    │
│  │                                                     │    │
│  │  Tools:                                             │    │
│  │  • query_prometheus  • query_loki                   │    │
│  │  • get_traces        • list_alerts                  │    │
│  │  • get_slo_status    • query_metrics                │    │
│  └──────────────────┬──────────────────────────────────┘    │
│                     │                                       │
│  ┌──────────────────▼──────────────────────────────────┐    │
│  │         Observability Stack                         │    │
│  │                                                     │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │    │
│  │  │ Prometheus  │  │    Loki     │  │    Tempo    │  │    │
│  │  │  (Metrics)  │  │   (Logs)    │  │  (Traces)   │  │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  │    │
│  │                                                     │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │    │
│  │  │AlertManager │  │   Grafana   │  │    Sloth    │  │    │
│  │  │  (Alerts)   │  │    (UI)     │  │   (SLOs)    │  │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Access Patterns

```yaml
Natural Language (via Agents):
  User: "Show me api-service errors in last hour"
  Agent: Calls MCP → query_loki('{app="api-service"} |= "ERROR"')
  Agent: Analyzes with LLM → Returns human-friendly response

Structured API (Direct MCP):
  SRE: mcp call query_prometheus 'rate(http_requests_total[5m])'
  MCP: Returns structured JSON data
  SRE: Pipes to scripts, dashboards, automation
```

---

## MCP Server Deployment

### Location & Service

**Location**: [Studio Cluster](../clusters/studio-cluster.md) (observability nodes)

**Configuration**:
```yaml
Service: mcp-observability.observability.svc.cluster.local:8080
Technology: Python + FastAPI + MCP SDK
Replicas: 3 (HA)
Resources:
  cpu: 500m
  memory: 1Gi
Auth: Kubernetes ServiceAccount + RBAC
```

### Server Implementation

```python
# mcp_server.py
from mcp import Server, Tool
from typing import Dict, List

class ObservabilityMCPServer:
    def __init__(self):
        self.prometheus = PrometheusClient("prometheus.observability:9090")
        self.loki = LokiClient("loki.observability:3100")
        self.tempo = TempoClient("tempo.observability:3200")
        self.alertmanager = AlertManagerClient("alertmanager.observability:9093")
        self.sloth = SlothClient("sloth.observability:8080")
        
        # Initialize MCP server
        self.server = Server("mcp-observability")
        self.register_tools()
    
    def register_tools(self):
        """Register all observability tools"""
        self.server.add_tool(self.query_prometheus)
        self.server.add_tool(self.query_loki)
        self.server.add_tool(self.get_traces)
        self.server.add_tool(self.list_active_alerts)
        self.server.add_tool(self.get_slo_status)
        self.server.add_tool(self.query_metrics)
    
    @Tool(
        name="query_prometheus",
        description="Query Prometheus metrics using PromQL",
        parameters={
            "query": "PromQL query string",
            "start": "Start time (RFC3339 or relative like 'now-1h')",
            "end": "End time (RFC3339 or relative like 'now')",
            "step": "Query resolution step in seconds"
        }
    )
    async def query_prometheus(self, query: str, start: str = "now-1h", 
                               end: str = "now", step: int = 60) -> Dict:
        """Query Prometheus metrics with PromQL"""
        result = await self.prometheus.query_range(
            query=query,
            start=start,
            end=end,
            step=f"{step}s"
        )
        return result
    
    @Tool(
        name="query_loki",
        description="Query Loki logs using LogQL",
        parameters={
            "logql": "LogQL query string",
            "start": "Start time",
            "end": "End time",
            "limit": "Maximum number of logs to return"
        }
    )
    async def query_loki(self, logql: str, start: str = "now-1h",
                        end: str = "now", limit: int = 100) -> Dict:
        """Query Loki logs with LogQL"""
        result = await self.loki.query_range(
            query=logql,
            start=start,
            end=end,
            limit=limit
        )
        return result
    
    @Tool(
        name="get_traces",
        description="Retrieve distributed traces from Tempo",
        parameters={
            "service": "Service name to query traces for",
            "operation": "Operation name (optional)",
            "start": "Start time",
            "end": "End time",
            "limit": "Maximum traces to return"
        }
    )
    async def get_traces(self, service: str, operation: str = None,
                        start: str = "now-1h", end: str = "now",
                        limit: int = 20) -> Dict:
        """Retrieve distributed traces from Tempo"""
        result = await self.tempo.search(
            service_name=service,
            operation=operation,
            start=start,
            end=end,
            limit=limit
        )
        return result
    
    @Tool(
        name="list_active_alerts",
        description="List active alerts from AlertManager",
        parameters={
            "severity": "Filter by severity (critical, warning, info)",
            "cluster": "Filter by cluster name"
        }
    )
    async def list_active_alerts(self, severity: str = None,
                                 cluster: str = None) -> List[Dict]:
        """List active alerts from AlertManager"""
        alerts = await self.alertmanager.get_alerts()
        
        # Filter by severity
        if severity:
            alerts = [a for a in alerts if a["severity"] == severity]
        
        # Filter by cluster
        if cluster:
            alerts = [a for a in alerts if a["cluster"] == cluster]
        
        return alerts
    
    @Tool(
        name="get_slo_status",
        description="Get SLO compliance status from Sloth",
        parameters={
            "service": "Service name to check SLO for"
        }
    )
    async def get_slo_status(self, service: str) -> Dict:
        """Get SLO compliance status from Sloth"""
        result = await self.sloth.get_slo_status(service)
        return result
    
    @Tool(
        name="query_metrics",
        description="Quick helper to query common metrics",
        parameters={
            "service": "Service name",
            "metric_type": "Type of metric (latency, errors, requests, cpu, memory)"
        }
    )
    async def query_metrics(self, service: str, metric_type: str) -> Dict:
        """Query common metrics for a service"""
        queries = {
            "latency": f'histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{{service="{service}"}}[5m]))',
            "errors": f'rate(http_requests_total{{service="{service}", status=~"5.."}}[5m])',
            "requests": f'rate(http_requests_total{{service="{service}"}}[5m])',
            "cpu": f'avg(rate(container_cpu_usage_seconds_total{{pod=~"{service}.*"}}[5m]))',
            "memory": f'avg(container_memory_working_set_bytes{{pod=~"{service}.*"}})'
        }
        
        query = queries.get(metric_type)
        if not query:
            return {"error": f"Unknown metric type: {metric_type}"}
        
        return await self.query_prometheus(query)

# Start server
if __name__ == "__main__":
    server = ObservabilityMCPServer()
    server.run(host="0.0.0.0", port=8080)
```

---

## Agent Integration

### Agent Usage Pattern

```python
# agent-auditor using MCP for observability
class SREAgent(BaseAgent):
    def __init__(self):
        super().__init__(name="agent-auditor")
        self.mcp = MCPClient("mcp-observability.observability:8080")
    
    async def handle_incident(self, query: str):
        """
        Handle incident investigation using MCP tools
        """
        # User: "Why is api-service slow?"
        
        # 1. Get latency metrics via MCP
        latency = await self.mcp.call_tool(
            tool="query_metrics",
            params={
                "service": "api-service",
                "metric_type": "latency"
            }
        )
        
        # 2. Get error logs via MCP
        logs = await self.mcp.call_tool(
            tool="query_loki",
            params={
                "logql": '{app="api-service", level="error"}',
                "start": "now-1h",
                "limit": 50
            }
        )
        
        # 3. Get distributed traces via MCP
        traces = await self.mcp.call_tool(
            tool="get_traces",
            params={
                "service": "api-service",
                "start": "now-1h",
                "limit": 10
            }
        )
        
        # 4. Analyze with LLM
        analysis = await self.llm.analyze(
            prompt=f"""
            Analyze this API slowness incident:
            
            Latency P95: {latency}
            Error Logs: {logs}
            Traces: {traces}
            
            Provide:
            1. Root cause
            2. Impact assessment
            3. Remediation steps
            """
        )
        
        return analysis
```

### Example Interactions

#### Example 1: Error Investigation

```python
# Natural language query
user: "Show me api-service errors in last hour"

# Agent workflow
agent_auditor:
  1. Parse intent: "error investigation"
  2. Call MCP:
     logs = mcp.query_loki(
       logql='{app="api-service", level="error"}',
       start="now-1h"
     )
  3. Analyze with SLM:
     summary = slm.summarize(logs)
  4. Respond:
     "Found 23 errors, mostly database timeouts:
      - 18 connection timeout to postgres
      - 5 query timeout (>5s)
      
      Recommendation: Check database connection pool"
```

#### Example 2: Latency Analysis

```python
# Natural language query
user: "Is the API slow right now?"

# Agent workflow
agent_auditor:
  1. Parse intent: "performance check"
  2. Call MCP:
     latency = mcp.query_metrics(
       service="api-service",
       metric_type="latency"
     )
  3. Compare with SLO:
     slo = mcp.get_slo_status(service="api-service")
  4. Analyze:
     current_p95 = latency["p95"]
     slo_target = slo["latency_target"]
     
     if current_p95 > slo_target:
       return "⚠️ Yes, API is slower than normal"
     else:
       return "✅ No, API is within SLO"
```

---

## Direct MCP Usage (Advanced)

### CLI Access

```bash
# SREs can use MCP CLI directly for efficiency
mcp connect mcp-observability.observability:8080

# Query Prometheus
mcp call query_prometheus \
  'rate(http_requests_total{app="api-service"}[5m])' \
  --start="now-1h"

# Query Loki
mcp call query_loki \
  '{app="api-service"} |= "ERROR"' \
  --limit=50

# Get traces
mcp call get_traces \
  --service="api-service" \
  --start="now-1h"

# List alerts
mcp call list_active_alerts \
  --severity="critical" \
  --cluster="pro"
```

### Script Integration

```python
# Python script using MCP directly
from mcp_client import MCPClient

# Connect to MCP server
mcp = MCPClient("mcp-observability.observability:8080")

# Get metrics
latency = await mcp.query_metrics(
    service="api-service",
    metric_type="latency"
)

# Get logs
logs = await mcp.query_loki(
    logql='{app="api-service", level="error"}',
    start="now-1h"
)

# Process results
if latency["p95"] > 500:
    send_alert("API latency high", logs)
```

### Dashboard Integration

```yaml
# Grafana dashboard using MCP data source
apiVersion: 1
datasources:
  - name: MCP-Observability
    type: mcp
    url: http://mcp-observability.observability:8080
    access: proxy

# Panels query MCP instead of Prometheus directly
panels:
  - title: API Latency
    datasource: MCP-Observability
    targets:
      - tool: query_metrics
        params:
          service: api-service
          metric_type: latency
```

---

## Metrics & Observability

### MCP Server Metrics

```yaml
# Agent-specific metrics
agent_requests_total{agent="bruno", intent="sre_help"}
agent_latency_seconds{agent="bruno", model="ollama"}
agent_llm_tokens_total{agent="bruno", model="vllm"}

# MCP Server metrics
mcp_tool_calls_total{tool="query_prometheus"}
mcp_tool_duration_seconds{tool="query_loki"}
mcp_errors_total{tool="get_traces"}
mcp_cache_hits_total{tool="query_prometheus"}
mcp_cache_miss_total{tool="query_prometheus"}

# Knowledge Graph metrics
kg_search_latency_seconds{collection="homelab-docs"}
kg_embedding_cache_hits_total
kg_rag_context_tokens_total

# LLM metrics
vllm_request_duration_seconds{model="llama-3.1-70b"}
vllm_queue_size{model="llama-3.1-70b"}
vllm_gpu_utilization{node="inference-1"}
```

### Performance Characteristics

| Tool | Latency | Caching | Throughput |
|------|---------|---------|------------|
| query_prometheus | 20-50ms | ✅ 5min | 100 req/s |
| query_loki | 50-200ms | ✅ 1min | 50 req/s |
| get_traces | 100-500ms | ✅ 5min | 20 req/s |
| list_alerts | 10-30ms | ✅ 30s | 200 req/s |
| get_slo_status | 30-100ms | ✅ 1min | 50 req/s |

---

## Benefits Summary

### For Teams
- **Natural Language**: Ask questions in plain English
- **No Tool Learning**: Don't need to learn PromQL, LogQL
- **Context-Aware**: Agents provide relevant insights

### For SREs
- **Direct Access**: Bypass agents for efficiency
- **Structured API**: Reliable, typed responses
- **Scriptable**: Integrate with automation

### For Agents
- **Reliable Tools**: Well-defined interfaces
- **Cached Results**: Fast repeated queries
- **Error Handling**: Graceful failures

### For System
- **Single Source**: One interface to observability
- **Consistent**: All queries go through MCP
- **Observable**: MCP itself is monitored

---

## Related Documentation

- [🤖 AI Architecture Overview](ai-agent-architecture.md)
- [🔧 AI Components](ai-components.md)
- [🎯 Agent Orchestration](agent-orchestration.md)
- [🌐 AI Connectivity](ai-connectivity.md)
- [📊 Observability Stack](../implementation/observability-stack.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

