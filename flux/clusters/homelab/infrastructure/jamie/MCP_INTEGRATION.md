# Jamie <-> Agent-SRE MCP Integration

## Overview

This document explains how Jamie Slack Bot communicates with Agent-SRE via MCP (Model Context Protocol) to execute SRE tools like checking golden signals, querying Prometheus, and analyzing logs.

## Architecture

```
┌──────────────┐
│ User (Slack) │
└──────┬───────┘
       │ "Check golden signals for homepage"
       ▼
┌──────────────────────┐
│  Jamie Slack Bot     │
│  - Detects intent    │
│  - Extracts params   │
└──────┬───────────────┘
       │ MCP tools/call request
       ▼
┌────────────────────────────┐
│ Agent-SRE MCP Server       │
│ (port 30120)               │
│ - Receives MCP requests    │
│ - Routes to agent service  │
└──────┬─────────────────────┘
       │ HTTP POST /golden-signals
       ▼
┌────────────────────────────┐
│ Agent-SRE Service          │
│ (port 8080)                │
│ - Executes PromQL queries  │
│ - Queries Prometheus       │
└──────┬─────────────────────┘
       │ HTTP GET /api/v1/query
       ▼
┌────────────────────────────┐
│ Prometheus                 │
│ (port 9090)                │
│ - Returns metrics          │
└────────────────────────────┘
```

## Flow Example: "Check Golden Signals for Homepage"

### 1. User Message in Slack
```
@Jamie check golden signals for homepage
```

### 2. Jamie Detects Intent
Jamie's `_detect_and_execute_tool()` method:
- Detects keywords: "golden signals"
- Extracts service name: "homepage"
- Decides to call `check_golden_signals` MCP tool

### 3. Jamie Calls MCP Tool
```python
payload = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "check_golden_signals",
        "arguments": {
            "service_name": "homepage",
            "namespace": "default"
        }
    }
}
```

Sends to: `http://sre-agent-mcp-server-service.agent-sre:30120/mcp`

### 4. MCP Server Routes Request
Agent-SRE MCP server (`mcp_server.py`):
- Receives MCP request
- Maps `check_golden_signals` to `/golden-signals` endpoint
- Forwards to agent service: `http://sre-agent-service:8080/golden-signals`

### 5. Agent Service Executes Query
Agent service (`agent.py`):
- Builds PromQL queries for golden signals:
  - **Latency**: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job="homepage"}[5m])) * 1000`
  - **Traffic**: `rate(http_requests_total{job="homepage"}[5m]) * 60`
  - **Errors**: `rate(http_requests_total{job="homepage",code=~"5.."}[5m]) / rate(http_requests_total{job="homepage"}[5m]) * 100`
  - **Saturation**: `avg(rate(container_cpu_usage_seconds_total{namespace="default",pod=~"homepage-.*"}[5m])) * 100`
- Queries Prometheus directly
- Evaluates thresholds
- Returns structured result

### 6. Response Back to Jamie
```json
{
  "service_name": "homepage",
  "namespace": "default",
  "overall_status": "healthy",
  "signals": {
    "latency": {"value": 45.2, "status": "healthy"},
    "traffic": {"value": 123.4, "status": "healthy"},
    "errors": {"value": 0.1, "status": "healthy"},
    "saturation": {"value": 25.3, "status": "healthy"}
  }
}
```

### 7. Jamie Formats Response
Jamie's LLM formats the data into a friendly message:
```
🤖 Hey! I checked the golden signals for homepage:

📊 Overall Status: ✅ Healthy

Signals:
• Latency: 45.2ms ✅
• Traffic: 123.4 req/min ✅
• Errors: 0.1% ✅
• Saturation: 25.3% ✅

Everything looks good! 🎉
```

## Available MCP Tools

### 1. check_golden_signals
**Description**: Check golden signals (latency, traffic, errors, saturation) for a service

**Arguments**:
- `service_name` (required): Service name (e.g., "homepage", "api")
- `namespace` (optional): Kubernetes namespace (default: "default")

**Example Usage**:
- "Check golden signals for homepage"
- "What's the status of the api service?"
- "Show me the health of frontend"

### 2. query_prometheus
**Description**: Execute a PromQL query against Prometheus

**Arguments**:
- `query` (required): PromQL query string

**Example Usage**:
- "Query Prometheus: up"
- "Run PromQL query: rate(http_requests_total[5m])"

### 3. get_pod_logs
**Description**: Get logs from a Kubernetes pod

**Arguments**:
- `pod_name` (required): Pod name
- `namespace` (optional): Namespace (default: "default")
- `tail_lines` (optional): Number of lines to tail (default: 100)

**Example Usage**:
- "Get logs for pod homepage-abc123"
- "Show me logs from homepage-abc123"

### 4. sre_chat
**Description**: General SRE chat and consultation using LLM

**Arguments**:
- `message` (required): Your SRE question or request

**Example Usage**:
- "What are best practices for monitoring?"
- "How do I debug high latency?"

### 5. analyze_logs
**Description**: Analyze logs for SRE insights

**Arguments**:
- `logs` (required): Log data to analyze

### 6. incident_response
**Description**: Get incident response guidance

**Arguments**:
- `incident` (required): Incident description

### 7. monitoring_advice
**Description**: Get monitoring and alerting advice

**Arguments**:
- `system` (required): System description

## Testing

### Local Testing
1. Port-forward the MCP server:
```bash
kubectl port-forward -n agent-sre svc/sre-agent-mcp-server-service 30120:30120
```

2. Run the test script:
```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/jamie
python test_mcp_flow.py
```

### Testing in Slack
1. Ensure Jamie is running in the cluster
2. Send messages to Jamie:
   - `@Jamie check golden signals for homepage`
   - `@Jamie what's the health of the api service?`
   - `@Jamie query prometheus: up`

## Configuration

### Jamie Configuration
Environment variables in `jamie-slack-bot` deployment:
- `AGENT_SRE_URL`: URL to agent-sre MCP server (default: `http://sre-agent-mcp-server-service.agent-sre:30120`)
- `OLLAMA_URL`: URL to Ollama LLM server
- `MODEL_NAME`: LLM model to use (e.g., `bruno-sre:latest`)

### Agent-SRE Configuration
Environment variables in `sre-agent-service` deployment:
- `PROMETHEUS_URL`: URL to Prometheus (default: `http://prometheus-operated.prometheus:9090`)
- `GRAFANA_MCP_URL`: URL to Grafana MCP server (future use)

## Debugging

### Check MCP Server Health
```bash
kubectl exec -n agent-sre -it deploy/sre-agent-mcp-server -- curl http://localhost:30120/health
```

### Check Agent Service Health
```bash
kubectl exec -n agent-sre -it deploy/sre-agent -- curl http://localhost:8080/health
```

### View Jamie Logs
```bash
kubectl logs -n jamie -f deploy/jamie-slack-bot
```

### View Agent-SRE Logs
```bash
kubectl logs -n agent-sre -f deploy/sre-agent-mcp-server
kubectl logs -n agent-sre -f deploy/sre-agent
```

## Troubleshooting

### Issue: "I'm having trouble connecting to the SRE agent"
**Cause**: MCP server is unavailable or unreachable

**Solution**:
1. Check if MCP server is running: `kubectl get pods -n agent-sre`
2. Check MCP server logs for errors
3. Verify service exists: `kubectl get svc -n agent-sre sre-agent-mcp-server-service`

### Issue: "No response from SRE agent"
**Cause**: Agent service is not processing requests correctly

**Solution**:
1. Check agent service logs
2. Verify Prometheus is reachable from agent service
3. Test Prometheus query directly: `curl http://prometheus-operated.prometheus:9090/api/v1/query?query=up`

### Issue: Tool not detected
**Cause**: Message doesn't match detection patterns

**Solution**:
1. Use clearer keywords: "golden signals", "query prometheus", "pod logs"
2. Include service name explicitly: "for homepage", "@homepage"
3. Check Jamie logs to see if detection triggered

## Future Enhancements

1. **Use Grafana MCP Server**: Integrate with grafana-mcp for richer dashboard/alerting capabilities
2. **Add Kubernetes Tools**: Integrate kubernetes-mcp server for pod management
3. **Add Loki Integration**: Query logs via Loki instead of direct kubectl
4. **Tool Suggestions**: Have Jamie proactively suggest tools based on context
5. **Multi-Service Queries**: Support checking golden signals for multiple services at once

## References

- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [LangChain MCP Integration](https://python.langchain.com/docs/integrations/tools/)
- [Prometheus Query API](https://prometheus.io/docs/prometheus/latest/querying/api/)
- [Golden Signals (SRE Book)](https://sre.google/sre-book/monitoring-distributed-systems/)

