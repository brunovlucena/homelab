# 🔍 Automated Investigation & Issue Management Guide

This guide explains how Agent-SRE automatically creates and investigates GitHub issues using LangGraph workflows and Grafana Sift.

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Setup](#setup)
- [Usage](#usage)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

## Overview

Agent-SRE provides **end-to-end automated investigation** for incidents, failures, and anomalies:

1. **🚨 Detection**: Receives alerts from various sources (Alertmanager, GitHub Actions, manual triggers)
2. **📝 Issue Creation**: Automatically creates a GitHub issue with relevant labels and assignees
3. **🔬 Investigation**: Runs Grafana Sift analysis to detect error patterns and slow requests
4. **🧠 Analysis**: Uses LLM to analyze findings and determine root cause
5. **💡 Recommendations**: Generates specific, actionable recommendations
6. **🔄 Updates**: Posts all findings back to the GitHub issue

## Architecture

```
┌─────────────────┐
│  Alert/Failure  │
│   Detection     │
└────────┬────────┘
         │
         v
┌─────────────────┐
│   LangGraph     │ ──────► State Management
│   Workflow      │         (MemorySaver)
└────────┬────────┘
         │
         ├──► 1. Detect & Classify
         │
         ├──► 2. Create GitHub Issue
         │         │
         │         v
         │    ┌──────────────┐
         │    │ GitHub API   │
         │    └──────────────┘
         │
         ├──► 3. Run Sift Investigation
         │         │
         │         v
         │    ┌──────────────┐
         │    │ Grafana Sift │
         │    │ - Loki logs  │
         │    │ - Tempo traces│
         │    └──────────────┘
         │
         ├──► 4. LLM Analysis
         │         │
         │         v
         │    ┌──────────────┐
         │    │ Ollama LLM   │
         │    │ Root Cause   │
         │    └──────────────┘
         │
         ├──► 5. Generate Recommendations
         │
         └──► 6. Update GitHub Issue
                   │
                   v
              ✅ Complete
```

### Components

| Component | Purpose | Technology |
|-----------|---------|------------|
| `investigation_workflow.py` | LangGraph workflow orchestration | LangGraph, LangChain |
| `github_integration.py` | GitHub API integration | httpx, Pydantic |
| `agent.py` | HTTP API endpoints | aiohttp |
| `deployments/sift/` | Investigation engine | Grafana Sift |
| `deployments/mcp-server/` | Observability tools | MCP protocol |

## Setup

### 1. Environment Variables

Add these to your Agent-SRE deployment:

```yaml
# deployments/agent/k8s-agent.yaml
env:
  # GitHub Integration
  - name: GITHUB_TOKEN
    valueFrom:
      secretKeyRef:
        name: agent-sre-secrets
        key: github-token
  - name: GITHUB_OWNER
    value: "brunovlucena"
  - name: GITHUB_REPO
    value: "homelab"
  - name: GITHUB_DEFAULT_ASSIGNEE
    value: "brunovlucena"
  
  # Ollama LLM
  - name: OLLAMA_URL
    value: "http://192.168.0.16:11434"
  - name: MODEL_NAME
    value: "llama3.2:3b"
  
  # MCP Server
  - name: MCP_SERVER_URL
    value: "http://sre-agent-mcp-server-service:30120"
  
  # Observability
  - name: PROMETHEUS_URL
    value: "http://prometheus-k8s.prometheus.svc.cluster.local:9090"
  - name: GRAFANA_URL
    value: "http://grafana.grafana.svc.cluster.local:3000"
  - name: LOKI_URL
    value: "http://loki-gateway.loki.svc.cluster.local:80"
  - name: TEMPO_URL
    value: "http://tempo.tempo.svc.cluster.local:3100"
```

### 2. Create GitHub Token Secret

```bash
# Create GitHub Personal Access Token with permissions:
# - repo (full access)
# - workflow (if triggering workflows)

kubectl create secret generic agent-sre-secrets \
  --from-literal=github-token="ghp_your_token_here" \
  -n agent-sre
```

### 3. Update Dependencies

Add to `pyproject.toml`:

```toml
dependencies = [
    # ... existing dependencies ...
    "httpx>=0.25.0",
    "pydantic>=2.0.0",
]
```

Then rebuild and redeploy:

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/agent-sre
make build-agent
make push-agent
make deploy-agent
```

### 4. Verify Installation

Test the investigation endpoints:

```bash
# Port-forward the agent service
kubectl port-forward -n agent-sre svc/sre-agent-service 8080:8080

# Test health endpoint
curl http://localhost:8080/health

# Test a simple investigation
curl -X POST http://localhost:8080/investigation/create \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Investigation","description":"Testing the system","severity":"low","component":"test"}'
```

## Usage

### API Endpoints

Agent-SRE exposes these endpoints for investigations:

#### 1. Create Investigation (Manual)

```bash
POST http://sre-agent-service:8080/investigation/create
Content-Type: application/json

{
  "title": "Navigation menu not visible in mobile - homepage",
  "description": "The navigation menu is not visible when viewing the homepage on mobile devices.",
  "severity": "medium",
  "component": "homepage"
}
```

**Response:**
```json
{
  "investigation": {
    "investigation_id": "abc-123-xyz",
    "issue_number": 28,
    "issue_url": "https://github.com/brunovlucena/homelab/issues/28",
    "root_cause": "...",
    "recommendations": [...],
    "error_patterns": 3,
    "slow_requests": 1,
    "completed": true
  }
}
```

#### 2. Workflow Failure Investigation

```bash
POST http://sre-agent-service:8080/investigation/workflow-failure
Content-Type: application/json

{
  "workflow_name": "Homepage CI",
  "run_id": "18595106881",
  "job_id": "53019314213",
  "run_url": "https://github.com/brunovlucena/homelab/actions/runs/18595106881",
  "job_url": "https://github.com/brunovlucena/homelab/actions/runs/18595106881/job/53019314213",
  "failure_details": "Build failed with error: ..."
}
```

#### 3. Get Investigation Status

```bash
GET http://sre-agent-service:8080/investigation/{investigation_id}
```

### CLI Usage

Using `curl`:

```bash
# Create investigation
curl -X POST http://sre-agent-service.agent-sre.svc.cluster.local:8080/investigation/create \
  -H "Content-Type: application/json" \
  -d '{
    "title": "High error rate in homepage API",
    "description": "Error rate increased from 0.1% to 5% in the last hour",
    "severity": "high",
    "component": "homepage"
  }'

# Check investigation
curl http://sre-agent-service.agent-sre.svc.cluster.local:8080/investigation/abc-123-xyz
```

### Kubernetes Port-Forward

```bash
# Port-forward the agent service
kubectl port-forward -n agent-sre svc/sre-agent-service 8080:8080

# Then use localhost
curl http://localhost:8080/investigation/create ...
```

## Examples

### Example 1: Mobile UI Issue (Issue #28)

**Input:**
```json
{
  "title": "Navigation menu not visible in mobile - homepage",
  "description": "The navigation menu is not visible when viewing the homepage on mobile devices.\n\nLocation: /repos/homelab/flux/clusters/homelab/infrastructure/homepage\n\nExpected: Menu should be visible on mobile\nActual: Menu is hidden",
  "severity": "medium",
  "component": "homepage"
}
```

**What Happens:**
1. ✅ GitHub Issue #28 created with labels: `automated`, `investigation`, `severity:medium`, `component:homepage`
2. 🔬 Sift investigation runs checking for:
   - Frontend error patterns in Loki logs
   - Slow page loads in Tempo traces
   - CSS/JavaScript errors
3. 🧠 LLM analyzes findings and determines:
   - Root cause: Missing mobile CSS breakpoint
   - Affected files: `frontend/src/components/Navigation.tsx`
4. 💡 Recommendations generated:
   - Add responsive CSS media queries
   - Test on multiple viewport sizes
   - Add mobile navigation toggle button
5. 🔄 Issue updated with complete investigation report

### Example 2: CI/CD Workflow Failure (Issue #30)

**Triggered by:** GitHub Actions workflow failure

**Input:**
```json
{
  "workflow_name": "Homepage CI",
  "run_id": "18595106881",
  "job_id": "53019314213",
  "run_url": "https://github.com/brunovlucena/homelab/actions/runs/18595106881",
  "job_url": "https://github.com/brunovlucena/homelab/actions/runs/18595106881/job/53019314213",
  "failure_details": "Build failed: TypeScript compilation error..."
}
```

**What Happens:**
1. ✅ GitHub Issue #30 created automatically
2. 🔬 Sift investigates related services for cascading failures
3. 🧠 LLM analyzes build logs and identifies:
   - TypeScript type error in specific file
   - Missing type definition
4. 💡 Recommendations:
   - Fix type definitions
   - Add pre-commit hooks
   - Update CI to catch errors earlier

### Example 3: Alertmanager Integration

Agent-SRE already has Alertmanager webhook support. Enhance it to auto-investigate:

```python
# In agent.py handle_alertmanager_webhook
# After processing alert, trigger investigation:

if status == "firing" and severity in ["critical", "high"]:
    await investigation_workflow.investigate(
        title=f"Alert: {alert_name}",
        description=investigation_message,
        severity=severity,
        component=labels.get("namespace", "unknown")
    )
```

## Workflow States

The LangGraph workflow maintains these states:

| State | Description |
|-------|-------------|
| `investigation_id` | Sift investigation ID |
| `issue_number` | GitHub issue number |
| `title` | Investigation title |
| `description` | Detailed description |
| `severity` | critical, high, medium, low |
| `component` | Affected component/service |
| `sift_investigation` | Sift results |
| `error_patterns` | Elevated error patterns |
| `slow_requests` | Slow operations detected |
| `root_cause` | LLM-determined root cause |
| `recommendations` | Actionable recommendations |
| `github_issue` | Created GitHub issue |
| `completed` | Workflow completion status |

## Troubleshooting

### Issue: GitHub issues not being created

**Check:**
```bash
# Verify GitHub token
kubectl get secret agent-sre-secrets -n agent-sre -o jsonpath='{.data.github-token}' | base64 -d

# Check agent logs
kubectl logs -n agent-sre -l app=sre-agent --tail=100

# Test GitHub API manually
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://api.github.com/repos/brunovlucena/homelab/issues
```

**Solution:**
- Ensure GitHub token has `repo` scope
- Verify token is not expired
- Check agent has network access to github.com

### Issue: Sift investigations fail

**Check:**
```bash
# Verify MCP server is running
kubectl get pods -n agent-sre -l app=mcp-server

# Check MCP server logs
kubectl logs -n agent-sre -l app=mcp-server --tail=100

# Test MCP server manually
curl http://sre-agent-mcp-server-service.agent-sre.svc.cluster.local:30120/health
```

**Solution:**
- Ensure Loki and Tempo are accessible
- Verify Sift database is writable
- Check labels match your services

### Issue: LLM analysis is slow or fails

**Check:**
```bash
# Verify Ollama is accessible
curl http://192.168.0.16:11434/api/tags

# Check model is loaded
curl http://192.168.0.16:11434/api/show -d '{"name":"llama3.2:3b"}'
```

**Solution:**
- Pre-load the model on Ollama server
- Increase timeout in workflow (currently 120s)
- Use a smaller, faster model for quicker results

### Issue: Workflow state not persisting

**Cause:** LangGraph uses in-memory `MemorySaver` by default

**Solution:** For production, use persistent checkpointer:

```python
# investigation_workflow.py
from langgraph.checkpoint.sqlite import SqliteSaver

# Replace MemorySaver with SqliteSaver
self.checkpointer = SqliteSaver.from_conn_string("/data/checkpoints.db")
```

## Advanced Usage

### Custom Investigation Workflows

Create your own investigation nodes:

```python
from investigation_workflow import AutomatedInvestigationWorkflow, InvestigationState

class CustomInvestigationWorkflow(AutomatedInvestigationWorkflow):
    def _build_graph(self):
        workflow = super()._build_graph()
        
        # Add custom node
        workflow.add_node("check_external_services", self._check_external_node)
        
        # Modify flow
        workflow.add_edge("run_sift_investigation", "check_external_services")
        workflow.add_edge("check_external_services", "analyze_findings")
        
        return workflow.compile(checkpointer=self.checkpointer)
    
    async def _check_external_node(self, state: InvestigationState):
        # Your custom logic
        pass
```

### Scheduled Investigations

Run proactive investigations on a schedule:

```yaml
# Create a CronJob
apiVersion: batch/v1
kind: CronJob
metadata:
  name: proactive-investigation
  namespace: agent-sre
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: investigate
            image: curlimages/curl:latest
            command:
            - /bin/sh
            - -c
            - |
              curl -X POST http://sre-agent-service:8080/investigation/create \
                -H "Content-Type: application/json" \
                -d '{"title":"Scheduled Health Check","description":"Proactive investigation","severity":"low","component":"platform"}'
          restartPolicy: OnFailure
```

## Next Steps

1. ✅ **Deploy**: Follow the [Setup](#setup) instructions
2. 🧪 **Test**: Create a manual investigation via API
3. 🔄 **Integrate**: Add to your GitHub Actions workflows
4. 📊 **Monitor**: Check Grafana dashboards for investigation metrics
5. 🎯 **Customize**: Add custom investigation nodes for your use cases

## Related Documentation

- [Agent-SRE README](README.md)
- [Sift Implementation Guide](SIFT_IMPLEMENTATION.md)
- [Sift Quick Start](SIFT_QUICKSTART.md)
- [GitHub Actions Auto-Investigate Workflow](../../.github/workflows/auto-investigate-failure.yml)

---

**Questions or issues?** Create an issue or check the logs!

