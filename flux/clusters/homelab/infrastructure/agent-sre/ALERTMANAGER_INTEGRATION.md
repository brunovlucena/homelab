# 🚨 Alertmanager → Agent-SRE Integration

## Overview
This integration enables **intelligent, AI-powered alert analysis** using the agent-sre LLM brain.

## 🔄 Workflow

```
┌─────────────────┐
│  Prometheus     │ Evaluates rules
│  Alertmanager   │
└────────┬────────┘
         │ Webhook POST
         │ /webhook/alert
         ▼
┌─────────────────┐
│   Agent-SRE     │ Receives alert payload
│   Webhook       │
└────────┬────────┘
         │ Extracts context
         │ (labels, annotations, severity)
         ▼
┌─────────────────┐
│   LangGraph     │ Runs investigation workflow
│   Workflow      │ - Analyze
└────────┬────────┘ - Generate Recommendations
         │         - Format Response
         │
         ▼
┌─────────────────┐
│   LLM Brain     │ Ollama: bruno-sre:latest
│   (Ollama)      │ Provides intelligent analysis
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Structured    │ JSON response with:
│   Analysis      │ - Root cause analysis
└─────────────────┘ - Impact assessment
                    - Mitigation steps
                    - Investigation queries
                    - Prevention recommendations
```

## 📋 Configuration

### 1. Alertmanager Configuration

Location: `flux/clusters/homelab/infrastructure/prometheus-operator/helmrelease.yaml`

```yaml
route:
  routes:
  # 🤖 Send all alerts to AI agent for analysis
  - receiver: 'ai-agent-webhook'
    matchers:
    - severity =~ "warning|critical"
    continue: true  # Continue to other receivers (Slack, PagerDuty)

receivers:
  - name: 'ai-agent-webhook'
    webhook_configs:
    - url: 'http://sre-agent-service.agent-sre.svc.cluster.local:8080/webhook/alert'
      send_resolved: true
      http_config:
        follow_redirects: true
      max_alerts: 0  # No limit
```

**Key Points:**
- ✅ `continue: true` - Alert continues to Slack/PagerDuty
- ✅ `send_resolved: true` - Also notified when alerts resolve
- ✅ Matches only `warning` and `critical` severity

### 2. Agent-SRE Webhook Handler

Location: `_vault/agent-sre/main.py`

```python
async def alertmanager_webhook_handler(request: Request) -> Response:
    """🚨 Alertmanager webhook handler"""
    data = await request.json()
    alerts = data.get("alerts", [])
    
    for alert in alerts:
        if alert.get("status") == "firing":
            # Extract context
            alert_name = alert["labels"]["alertname"]
            severity = alert["labels"]["severity"]
            annotations = alert["annotations"]
            
            # Execute LLM investigation
            result = await agent.execute(
                message=investigation_message,
                task_type="incident",
                context={...}
            )
```

**Webhook receives:**
- Alert metadata (name, severity, fingerprint)
- Labels (pod, namespace, container, etc.)
- Annotations (summary, description, runbook)
- Timestamps (startsAt, endsAt)
- Generator URL (link to Prometheus query)

### 3. LLM Investigation

The agent uses **LangGraph** with a state machine workflow:

```python
# Workflow: analyze → generate_recommendations → format_response

1. analyze_node()
   - Extracts key insights from alert
   - Uses task-specific system prompt for "incident"
   - Returns analysis

2. generate_recommendations_node()
   - Based on analysis, generates 3-5 actionable recommendations
   - Formats as numbered list

3. format_response_node()
   - Structures the final output
   - Returns markdown-formatted response
```

**System Prompt (Incident Mode):**
```
You are an expert SRE assistant specializing in incident response.
Focus on:
1. Immediate impact assessment
2. Quick mitigation steps
3. Root cause investigation
4. Communication strategy
5. Post-incident actions
```

## 🧪 Testing

### Option 1: Using Test Script

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/agent-sre

# Test with NodePort (direct to service)
./test-alert-webhook.sh

# Or specify custom URL
SERVICE_URL=http://localhost:31081 ./test-alert-webhook.sh
```

The script tests:
1. ⚠️ Firing warning alert (High Memory)
2. 🔥 Firing critical alert (Pod CrashLooping)
3. ✅ Resolved alert

### Option 2: Manual curl Test

```bash
curl -X POST http://192.168.0.16:31081/webhook/alert \
  -H "Content-Type: application/json" \
  -d '{
    "alerts": [{
      "status": "firing",
      "labels": {
        "alertname": "HighCPU",
        "severity": "warning",
        "pod": "test-pod"
      },
      "annotations": {
        "summary": "High CPU usage",
        "description": "CPU usage above 80%"
      }
    }]
  }'
```

### Option 3: Real Alertmanager Test

1. **Port-forward to Alertmanager:**
   ```bash
   kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-alertmanager 9093:9093
   ```

2. **Trigger a test alert** via Prometheus rules or create a manual firing alert

3. **Watch agent-sre logs:**
   ```bash
   kubectl logs -n agent-sre -l app=sre-agent -f
   ```

## 📊 Expected Output

### Webhook Response Example

```json
{
  "message": "Processed 1 alerts",
  "results": [
    {
      "alert_name": "HighMemoryUsage",
      "severity": "warning",
      "analysis": "## 🔍 Analysis\n\nThe alert indicates high memory usage...",
      "thread_id": "alert-abc123def456"
    }
  ],
  "service": "sre-agent"
}
```

### LLM Analysis Example

```markdown
## 🔍 Analysis

The alert indicates high memory usage on pod bruno-site-api-xyz123 at 85% (1.7Gi of 2Gi).
This is concerning as it approaches OOM kill threshold.

**Key Observations:**
- Memory has been trending upward
- Pod may have a memory leak
- No recent deployments that would explain the increase

**Potential Root Causes:**
1. Memory leak in application code
2. Inefficient caching strategy
3. Increased traffic without scaling
4. Resource limits too low for workload

## 💡 Recommendations

1. **Immediate**: Check if pod is under heavy load - query traffic metrics
   ```promql
   rate(http_requests_total{pod="bruno-site-api-xyz123"}[5m])
   ```

2. **Short-term**: Restart pod to clear memory and monitor
   ```bash
   kubectl delete pod bruno-site-api-xyz123 -n homepage
   ```

3. **Investigation**: Analyze heap dumps for memory leaks
   
4. **Long-term**: Consider increasing memory limits or implementing memory profiling

5. **Prevention**: Set up memory profiling and regular leak detection
```

## 🔍 Monitoring

### Check Agent Health

```bash
# Health check
curl http://192.168.0.16:31081/health

# Expected output
{
  "status": "healthy",
  "service": "sre-agent",
  "ollama_url": "http://192.168.0.16:11434",
  "model_name": "bruno-sre:latest",
  "llm_connected": true,
  "graph_compiled": true
}
```

### View Logs

```bash
# Agent logs
kubectl logs -n agent-sre -l app=sre-agent --tail=100 -f

# Look for:
# ✅ "Ollama connection established"
# 🔔 "Received alert: AlertName (severity=warning, status=firing)"
# ✅ "Completed investigation for alert: AlertName"
```

### Alertmanager Logs

```bash
kubectl logs -n prometheus -l app.kubernetes.io/name=alertmanager --tail=50 -f

# Look for:
# "Notify for alerts" with receiver="ai-agent-webhook"
```

## 🚀 Deployment Steps

1. **Ensure agent-sre is running:**
   ```bash
   kubectl get pods -n agent-sre
   kubectl get svc -n agent-sre sre-agent-service
   ```

2. **Apply Alertmanager configuration:**
   ```bash
   flux reconcile kustomization infrastructure
   # Or
   kubectl rollout restart statefulset -n prometheus prometheus-kube-prometheus-prometheus
   ```

3. **Verify webhook is configured:**
   ```bash
   kubectl get secret -n prometheus prometheus-kube-prometheus-alertmanager-generated -o yaml
   # Check for ai-agent-webhook receiver
   ```

4. **Test the integration:**
   ```bash
   ./test-alert-webhook.sh
   ```

## 🐛 Troubleshooting

### Webhook not receiving alerts

1. Check Alertmanager config:
   ```bash
   kubectl exec -n prometheus prometheus-kube-prometheus-alertmanager-0 -- \
     cat /etc/alertmanager/config/alertmanager.yaml.gz | gunzip
   ```

2. Check Alertmanager logs for errors:
   ```bash
   kubectl logs -n prometheus -l app.kubernetes.io/name=alertmanager | grep "ai-agent-webhook"
   ```

3. Verify service is reachable:
   ```bash
   kubectl run test-curl --image=curlimages/curl -it --rm -- \
     curl -v http://sre-agent-service.agent-sre.svc.cluster.local:8080/health
   ```

### Agent-SRE not processing alerts

1. Check agent logs:
   ```bash
   kubectl logs -n agent-sre -l app=sre-agent --tail=100
   ```

2. Verify Ollama connection:
   ```bash
   kubectl logs -n agent-sre -l app=sre-agent | grep "Ollama"
   ```

3. Test webhook directly:
   ```bash
   kubectl port-forward -n agent-sre svc/sre-agent-service 8080:8080
   curl -X POST http://localhost:8080/webhook/alert -d '{"alerts":[]}'
   ```

### LLM not generating analysis

1. Check Ollama availability:
   ```bash
   curl http://192.168.0.16:11434/api/tags
   ```

2. Verify model is pulled:
   ```bash
   # Should see "bruno-sre:latest" or "gemma3n:e4b"
   ```

3. Check LLM invocation logs:
   ```bash
   kubectl logs -n agent-sre -l app=sre-agent | grep "Processing with LLM"
   ```

## 📈 Benefits

### Before
- Alerts go to Slack
- Engineers manually investigate
- Time to understand context: 5-15 minutes

### After  
- ✅ AI analyzes alert immediately
- ✅ Provides root cause analysis
- ✅ Suggests investigation queries
- ✅ Recommends mitigation steps
- ✅ Engineers get head start
- ⏱️ Time to actionable insights: < 30 seconds

## 🔮 Future Enhancements

1. **Auto-remediation**: Agent executes safe mitigation steps automatically
2. **Slack Integration**: Send analysis to Slack alongside alert
3. **Historical Context**: Link to similar past incidents
4. **Metrics Correlation**: Auto-fetch related metrics
5. **Log Analysis**: Fetch and analyze relevant logs automatically
6. **Runbook Generation**: Create runbooks from incident patterns

---

**Status**: ✅ Configured and Ready to Test  
**Last Updated**: 2025-10-09

