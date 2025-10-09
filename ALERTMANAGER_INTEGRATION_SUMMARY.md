# 🚨 Alertmanager → Agent-SRE Webhook Integration - Complete!

## ✅ What Was Completed

### 1. Added Alertmanager Webhook Configuration
**File**: `flux/clusters/homelab/infrastructure/prometheus-operator/helmrelease.yaml`

✅ Added new `ai-agent-webhook` receiver that sends alerts to agent-sre  
✅ Configured to match `warning|critical` severity alerts  
✅ Set `continue: true` to also send alerts to Slack/PagerDuty  
✅ Webhook URL: `http://sre-agent-service.agent-sre.svc.cluster.local:8080/webhook/alert`

### 2. Implemented Webhook Handler in Agent-SRE
**File**: `flux/clusters/homelab/infrastructure/agent-sre/deployments/agent/agent.py`

✅ Added `/webhook/alert` endpoint handler  
✅ Extracts alert context (labels, annotations, severity, fingerprint)  
✅ Builds detailed investigation message for LLM  
✅ Calls `sre_agent.incident_response()` for intelligent analysis  
✅ Returns structured JSON response with analysis  
✅ Only processes firing alerts (ignores resolved for now)  
✅ Added logging at startup to show webhook endpoint

### 3. Created Comprehensive Documentation
**Files Created**:
- `flux/clusters/homelab/infrastructure/agent-sre/ALERTMANAGER_INTEGRATION.md`
- `flux/clusters/homelab/infrastructure/agent-sre/test-alert-webhook.sh`
- `ALERTMANAGER_INTEGRATION_SUMMARY.md` (this file)

## 🔄 Complete Workflow

```
┌────────────────┐
│  Prometheus    │  Evaluates alert rules
│  Alertmanager  │
└───────┬────────┘
        │ HTTP POST /webhook/alert
        │ {
        │   "alerts": [{
        │     "status": "firing",
        │     "labels": {...},
        │     "annotations": {...}
        │   }]
        │ }
        ▼
┌────────────────┐
│  agent-sre     │  🚨 handle_alertmanager_webhook()
│  webhook       │  - Extracts context
└───────┬────────┘  - Filters firing alerts
        │
        │ sre_agent.incident_response(...)
        ▼
┌────────────────┐
│  SRE Agent     │  🤖 LLM Brain (Ollama)
│  LLM Core      │  - Analyzes alert
└───────┬────────┘  - Assesses impact
        │           - Suggests mitigations
        │           - Generates queries
        ▼
┌────────────────┐
│  Structured    │  📊 JSON Response
│  Analysis      │  {
└────────────────┘    "results": [{
                        "alert_name": "...",
                        "severity": "...",
                        "analysis": "..."
                      }]
                    }
```

## 🧪 Testing

### Option 1: Run Test Script
```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/agent-sre

# After agent-sre is rebuilt and deployed
./test-alert-webhook.sh
```

The script tests:
1. ⚠️ Warning alert (High Memory)
2. 🔥 Critical alert (Pod CrashLooping)
3. ✅ Resolved alert

### Option 2: Trigger Real Alert
1. Port-forward to Alertmanager:
   ```bash
   kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-alertmanager 9093:9093
   ```

2. Create a test alert or wait for a real one

3. Watch agent logs:
   ```bash
   kubectl logs -n agent-sre -l app=sre-agent -f | grep -E "Received alert|Completed investigation"
   ```

## 🚀 Deployment Steps

### 1. Build and Push Updated Agent Image
```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab

# Commit changes
git add flux/clusters/homelab/infrastructure/agent-sre/deployments/agent/agent.py
git commit -m "feat(agent-sre): add Alertmanager webhook handler for intelligent alert analysis"

# Push to trigger CI/CD
git push origin main
```

The GitHub Actions workflow will:
- Run tests
- Build multi-arch Docker image (`ghcr.io/brunovlucena/agent-sre:latest`)
- Push to GitHub Container Registry
- Scan for vulnerabilities

### 2. Deploy Alertmanager Configuration
```bash
# The alertmanager config change is already in the repo
git add flux/clusters/homelab/infrastructure/prometheus-operator/helmrelease.yaml
git commit -m "feat(alertmanager): add AI agent webhook for intelligent alert analysis"
git push origin main

# Trigger Flux reconciliation
flux reconcile kustomization infrastructure

# Or manually restart alertmanager
kubectl rollout restart statefulset -n prometheus prometheus-kube-prometheus-alertmanager
```

### 3. Verify Deployment
```bash
# Check agent-sre is running new version
kubectl get pods -n agent-sre
kubectl logs -n agent-sre -l app=sre-agent | grep "Alertmanager webhook"

# Expected:
# 🚨 Alertmanager webhook: http://localhost:8080/webhook/alert

# Verify alertmanager config
kubectl exec -n prometheus prometheus-kube-prometheus-alertmanager-0 -- \
  cat /etc/alertmanager/config/alertmanager.yaml.gz | gunzip | grep -A 5 "ai-agent-webhook"
```

### 4. Test the Integration
```bash
# Port-forward to agent
kubectl port-forward -n agent-sre svc/sre-agent-service 8080:8080

# Run test
cd flux/clusters/homelab/infrastructure/agent-sre
SERVICE_URL=http://localhost:8080 ./test-alert-webhook.sh
```

## 📊 Expected Output

### Webhook Request (from Alertmanager)
```json
{
  "alerts": [{
    "status": "firing",
    "labels": {
      "alertname": "HighMemoryUsage",
      "severity": "warning",
      "namespace": "homepage",
      "pod": "bruno-site-api-xyz"
    },
    "annotations": {
      "summary": "High memory usage detected",
      "description": "Pod is using 85% memory"
    },
    "startsAt": "2025-10-09T13:20:00Z",
    "fingerprint": "abc123"
  }]
}
```

### LLM Analysis Response
```json
{
  "message": "Processed 1 alerts",
  "results": [{
    "alert_name": "HighMemoryUsage",
    "severity": "warning",
    "analysis": "## Analysis\n\nThe alert indicates high memory usage...\n\n## Recommendations\n\n1. Check current memory usage trends...\n2. Investigate memory leaks...\n3. Consider scaling...\n4. Review resource limits...\n5. Enable memory profiling...",
    "fingerprint": "abc123"
  }],
  "service": "sre-agent",
  "timestamp": "2025-10-09T13:21:15.123Z"
}
```

### Agent Logs
```
INFO:__main__:🔔 Received alert: HighMemoryUsage (severity=warning, status=firing)
INFO:__main__:✅ Completed investigation for alert: HighMemoryUsage
INFO:aiohttp.access:10.96.x.x [09/Oct/2025:13:21:15 +0000] "POST /webhook/alert HTTP/1.1" 200 1234
```

## 📈 Benefits

### Before
```
Alert fires → Slack notification → Engineer investigates manually
Time to actionable insights: 5-15 minutes
```

### After
```
Alert fires → Alertmanager → Agent-SRE LLM → Structured analysis
Time to actionable insights: < 30 seconds
```

**Key Improvements:**
- ✅ Immediate intelligent analysis
- ✅ Root cause suggestions
- ✅ Mitigation steps provided
- ✅ Investigation queries generated  
- ✅ Engineers get head start on incident response
- ✅ Consistent analysis approach
- ✅ Learning from patterns over time

## 🔮 Future Enhancements

1. **Jamie Integration**: Send analysis to Slack via Jamie bot
2. **Auto-remediation**: Execute safe mitigation steps automatically
3. **Historical Context**: Link to similar past incidents
4. **Metrics Fetching**: Auto-fetch related Prometheus metrics
5. **Log Correlation**: Automatically analyze related logs
6. **Runbook Generation**: Create/update runbooks from patterns
7. **Grafana Integration**: Create incident dashboards automatically
8. **Post-mortem**: Generate incident reports

## 📝 Files Modified

### Configuration
- ✅ `flux/clusters/homelab/infrastructure/prometheus-operator/helmrelease.yaml`

### Code
- ✅ `flux/clusters/homelab/infrastructure/agent-sre/deployments/agent/agent.py`

### Documentation
- ✅ `flux/clusters/homelab/infrastructure/agent-sre/ALERTMANAGER_INTEGRATION.md`
- ✅ `flux/clusters/homelab/infrastructure/agent-sre/test-alert-webhook.sh`
- ✅ `ALERTMANAGER_INTEGRATION_SUMMARY.md`

## 🎯 Next Steps

1. **Commit and Push** changes to trigger CI/CD
2. **Wait for image build** (check GitHub Actions)
3. **Verify deployment** in cluster
4. **Run test script** to verify webhook works
5. **Monitor real alerts** and observe AI analysis
6. **Iterate** on system prompts based on analysis quality
7. **Integrate with Jamie** for Slack notifications

## 🐛 Troubleshooting

### Webhook not receiving alerts
```bash
# Check alertmanager config
kubectl exec -n prometheus prometheus-kube-prometheus-alertmanager-0 -- \
  cat /etc/alertmanager/config/alertmanager.yaml.gz | gunzip | grep ai-agent

# Check alertmanager logs
kubectl logs -n prometheus -l app.kubernetes.io/name=alertmanager | grep webhook
```

### Agent not processing alerts
```bash
# Check agent logs
kubectl logs -n agent-sre -l app=sre-agent --tail=100

# Test webhook directly
kubectl port-forward -n agent-sre svc/sre-agent-service 8080:8080
curl -X POST http://localhost:8080/webhook/alert -d '{"alerts":[]}'
```

### LLM not generating analysis
```bash
# Check Ollama connection
curl http://192.168.0.16:11434/api/tags

# Check agent logs for LLM errors
kubectl logs -n agent-sre -l app=sre-agent | grep -E "Ollama|LLM|Error"
```

---

**Status**: ✅ Implementation Complete - Ready for Deployment  
**Date**: 2025-10-09  
**Next Action**: Commit changes and trigger CI/CD pipeline

