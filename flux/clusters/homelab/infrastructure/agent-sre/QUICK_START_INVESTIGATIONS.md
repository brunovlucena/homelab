# 🚀 Quick Start: Automated Investigations

Get started with automated GitHub issue creation and investigation in 5 minutes.

## 🎯 What You Get

When issues occur, Agent-SRE automatically:
- ✅ Creates detailed GitHub issues
- 🔬 Analyzes logs and traces via Grafana Sift
- 🧠 Determines root cause using AI
- 💡 Generates actionable recommendations
- 🔄 Updates issue with complete findings

## ⚡ Quick Setup

### 1. Add GitHub Token Secret

```bash
# Create a GitHub Personal Access Token with 'repo' scope at:
# https://github.com/settings/tokens

kubectl create secret generic agent-sre-secrets \
  --from-literal=github-token="ghp_your_token_here" \
  -n agent-sre
```

### 2. Update Agent-SRE Deployment

Add environment variables:

```yaml
# deployments/agent/k8s-agent.yaml
env:
  - name: GITHUB_TOKEN
    valueFrom:
      secretKeyRef:
        name: agent-sre-secrets
        key: github-token
  - name: GITHUB_OWNER
    value: "brunovlucena"
  - name: GITHUB_REPO
    value: "homelab"
```

### 3. Deploy

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/agent-sre

# Build and push
make build-agent
make push-agent

# Deploy
kubectl apply -k deployments/agent/
```

### 4. Test

```bash
# Port-forward the service
kubectl port-forward -n agent-sre svc/sre-agent-service 8080:8080

# Trigger a test investigation
curl -X POST http://localhost:8080/investigation/create \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Investigation",
    "description": "Testing automated investigation system",
    "severity": "low",
    "component": "test"
  }'

# Check response for issue URL
# Check GitHub for new issue
```

## 🔥 Usage Patterns

### Pattern 1: Alertmanager (Automatic)

**Setup once, investigations happen automatically!**

Add webhook to Alertmanager config:

```yaml
receivers:
  - name: 'agent-sre'
    webhook_configs:
      - url: 'http://sre-agent-service.agent-sre.svc.cluster.local:8080/webhook/alert'
```

**Result:** Every critical/warning alert automatically creates an investigated issue.

📖 [Full Guide](ALERTMANAGER_INTEGRATION.md)

### Pattern 2: Slack Command (Manual)

Trigger from Slack via Jamie:

```
@jamie investigate "Homepage API latency spike" severity=high component=homepage
```

📖 [Integration Code](TRIGGER_INVESTIGATION_GUIDE.md#-slack-integration-jamie)

### Pattern 3: API Call (Manual)

From anywhere in your cluster or via port-forward:

```bash
curl -X POST http://sre-agent-service.agent-sre.svc.cluster.local:8080/investigation/create \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Database connection timeouts",
    "description": "Seeing increased timeouts in production",
    "severity": "high",
    "component": "database"
  }'
```

📖 [API Reference](TRIGGER_INVESTIGATION_GUIDE.md#-api-reference)

### Pattern 4: Python/Go Integration

Call from your applications:

```python
import httpx

async def investigate(title: str, desc: str):
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            "http://sre-agent-service.agent-sre.svc.cluster.local:8080/investigation/create",
            json={"title": title, "description": desc, "severity": "high"}
        )
        return resp.json()
```

📖 [Code Examples](TRIGGER_INVESTIGATION_GUIDE.md#-advanced-usage)

## 📊 What Gets Investigated?

### Error Patterns (via Loki)
- Elevated error rates vs 24h baseline
- New error types
- Error pattern clustering
- Affected services

### Performance Issues (via Tempo)
- Slow request detection vs baseline
- P95 latency degradation
- Slow operations by service
- Trace analysis

### Root Cause Analysis (via LLM)
- Correlates logs, traces, and metrics
- Identifies most likely causes
- Provides evidence and reasoning
- Suggests investigation queries

### Recommendations
- Immediate mitigation steps
- Short-term fixes with code snippets
- Long-term improvements
- Prevention strategies

## 🎯 Example Investigation Results

### Input
```json
{
  "title": "Homepage API errors increased",
  "description": "Error rate jumped from 0.1% to 5%",
  "severity": "high",
  "component": "homepage"
}
```

### Output (GitHub Issue)

```markdown
## 🔍 Investigation Results

**Investigation ID**: `abc-123-xyz`
**Completed**: 2025-10-17T10:35:00Z

### 📊 Findings

#### Error Patterns
| Pattern | Current | Baseline | Factor | Severity |
| --- | --- | --- | --- | --- |
| `ERROR: Database connection timeout` | 450 | 5 | 90.0x | critical |
| `ERROR: API gateway timeout` | 120 | 2 | 60.0x | high |
| `WARN: Slow query detected` | 80 | 10 | 8.0x | medium |

#### Slow Operations
| Operation | Current P95 | Baseline P95 | Factor | Severity |
| --- | --- | --- | --- | --- |
| `GET /api/projects` | 2500ms | 200ms | 12.5x | critical |
| `GET /api/users` | 1800ms | 150ms | 12.0x | critical |

### 🧠 Root Cause Analysis

Analysis indicates the issue is caused by **database connection pool exhaustion**:

1. **Evidence**: 450 "connection timeout" errors (90x increase)
2. **Timing**: Started at 09:45 AM, correlates with deployment
3. **Impact**: API latency increased 12x on database-dependent endpoints
4. **Related**: Connection pool metrics show 100% utilization

### 💡 Recommendations

#### Immediate Actions (next 5 minutes)
1. Scale database connection pool from 20 to 50 connections
2. Restart homepage-api pods to reset connections
3. Enable query timeout at 5 seconds

#### Short-term Fixes (next 1 hour)
4. Update `homepage/api/config.yaml`:
   ```yaml
   database:
     max_connections: 50
     connection_timeout: 5s
     idle_timeout: 60s
   ```
5. Add connection pool monitoring dashboard

#### Long-term Prevention
6. Implement connection pool auto-scaling
7. Add database connection alerts at 80% utilization
8. Review and optimize slow queries
```

## 🚀 Next Steps

1. ✅ Complete [Quick Setup](#-quick-setup)
2. 🔗 Configure [Alertmanager Integration](ALERTMANAGER_INTEGRATION.md)
3. 💬 Add [Slack Integration](TRIGGER_INVESTIGATION_GUIDE.md#-slack-integration-jamie)
4. 📊 Create monitoring dashboards
5. 🎯 Customize investigation workflow

## 🆘 Troubleshooting

### Issue not created?
```bash
# Check Agent-SRE logs
kubectl logs -n agent-sre -l app=sre-agent --tail=50

# Verify GitHub token
kubectl get secret agent-sre-secrets -n agent-sre
```

### Sift analysis failed?
```bash
# Check MCP server
kubectl logs -n agent-sre -l app=mcp-server --tail=50

# Verify Loki/Tempo
kubectl get pods -n loki
kubectl get pods -n tempo
```

### LLM analysis slow/failed?
```bash
# Check Ollama
curl http://192.168.0.16:11434/api/tags

# Use faster model
kubectl set env deployment/sre-agent -n agent-sre MODEL_NAME=llama3.2:3b
```

## 📚 Full Documentation

- 📘 [Automated Investigation Guide](AUTOMATED_INVESTIGATION_GUIDE.md) - Complete architecture
- 🚀 [Trigger Guide](TRIGGER_INVESTIGATION_GUIDE.md) - API, Slack, MCP methods
- 🚨 [Alertmanager Integration](ALERTMANAGER_INTEGRATION.md) - Automatic alerts
- 🔬 [Sift Implementation](SIFT_IMPLEMENTATION.md) - How Sift works

---

**Ready to investigate!** 🔍



