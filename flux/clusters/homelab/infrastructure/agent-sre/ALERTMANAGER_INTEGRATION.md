# 🚨 Alertmanager Integration Guide

This guide explains how to configure Alertmanager to automatically trigger investigations when alerts fire.

## 📋 Overview

When an alert fires in Alertmanager, it automatically:
1. ✅ **Creates GitHub issue** with full alert context
2. 🔬 **Runs Sift investigation** analyzing logs and traces
3. 🧠 **Performs LLM analysis** to determine root cause
4. 💡 **Generates recommendations** for resolution
5. 🔄 **Updates GitHub issue** with complete findings

**No manual intervention required!**

## 🏗️ Architecture

```
┌─────────────┐
│ Prometheus  │
│   Alerts    │
└──────┬──────┘
       │
       v
┌─────────────┐      Webhook       ┌─────────────┐
│Alertmanager │ ───────────────►   │ Agent-SRE   │
│   (firing)  │                    │  Webhook    │
└─────────────┘                    └──────┬──────┘
                                          │
                                          v
                                   ┌──────────────┐
                                   │ LangGraph    │
                                   │ Investigation│
                                   │  Workflow    │
                                   └──────┬───────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
                    v                     v                     v
              ┌──────────┐         ┌──────────┐          ┌──────────┐
              │ GitHub   │         │ Grafana  │          │  Ollama  │
              │  Issue   │         │   Sift   │          │   LLM    │
              └──────────┘         └──────────┘          └──────────┘
                                         │
                                         │
                                   ┌─────┴─────┐
                                   │           │
                                   v           v
                              ┌────────┐  ┌────────┐
                              │  Loki  │  │ Tempo  │
                              │  Logs  │  │ Traces │
                              └────────┘  └────────┘
```

## ⚙️ Configuration

### Step 1: Update Alertmanager Configuration

Add Agent-SRE as a webhook receiver in your Alertmanager config:

```yaml
# alertmanager.yaml
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname', 'cluster', 'namespace']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'agent-sre-webhook'
  routes:
    # Critical alerts trigger immediate investigation
    - match:
        severity: critical
      receiver: 'agent-sre-webhook'
      continue: true  # Also send to other receivers
    
    # Warning alerts also trigger investigation
    - match:
        severity: warning
      receiver: 'agent-sre-webhook'
      continue: true
    
    # Info alerts go to different receiver (optional)
    - match:
        severity: info
      receiver: 'default-receiver'

receivers:
  - name: 'agent-sre-webhook'
    webhook_configs:
      - url: 'http://sre-agent-service.agent-sre.svc.cluster.local:8080/webhook/alert'
        send_resolved: false  # Only send firing alerts
        max_alerts: 10  # Limit batch size
        http_config:
          follow_redirects: true
        
  - name: 'default-receiver'
    webhook_configs:
      - url: 'http://your-other-webhook:9000/alerts'
```

### Step 2: Apply Configuration

If you're using the Prometheus Operator:

```bash
# Update the Alertmanager secret
kubectl create secret generic alertmanager-prometheus-operator-kube-p-alertmanager \
  --from-file=alertmanager.yaml=alertmanager.yaml \
  --namespace=prometheus \
  --dry-run=client -o yaml | kubectl apply -f -

# Reload Alertmanager
kubectl exec -n prometheus alertmanager-prometheus-operator-kube-p-alertmanager-0 \
  -- curl -X POST http://localhost:9093/-/reload
```

Or if using HelmRelease with inline config:

```yaml
# prometheus-operator/helmrelease.yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: prometheus-operator
  namespace: prometheus
spec:
  values:
    alertmanager:
      config:
        global:
          resolve_timeout: 5m
        route:
          group_by: ['alertname', 'namespace']
          receiver: 'agent-sre-webhook'
        receivers:
          - name: 'agent-sre-webhook'
            webhook_configs:
              - url: 'http://sre-agent-service.agent-sre.svc.cluster.local:8080/webhook/alert'
                send_resolved: false
```

### Step 3: Verify Agent-SRE Service

Ensure the Agent-SRE service is accessible from Alertmanager:

```bash
# Check service exists
kubectl get svc -n agent-sre sre-agent-service

# Test from within cluster
kubectl run -n prometheus test-curl --rm -it --image=curlimages/curl -- \
  curl http://sre-agent-service.agent-sre.svc.cluster.local:8080/health

# Check webhook endpoint
kubectl run -n prometheus test-webhook --rm -it --image=curlimages/curl -- \
  curl -X POST http://sre-agent-service.agent-sre.svc.cluster.local:8080/webhook/alert \
  -H "Content-Type: application/json" \
  -d '{"alerts":[]}'
```

## 🧪 Testing

### Test 1: Manual Alert

Send a test alert to verify the integration:

```bash
# Create test alert payload
cat > test-alert.json <<'EOF'
{
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "TestAlert",
        "severity": "warning",
        "namespace": "homepage",
        "pod": "homepage-api-12345"
      },
      "annotations": {
        "summary": "This is a test alert",
        "description": "Testing the Alertmanager webhook integration"
      },
      "startsAt": "2025-10-17T10:00:00Z",
      "generatorURL": "http://prometheus:9090/graph?g0.expr=up",
      "fingerprint": "test123"
    }
  ]
}
EOF

# Send to Agent-SRE
kubectl port-forward -n agent-sre svc/sre-agent-service 8080:8080 &
curl -X POST http://localhost:8080/webhook/alert \
  -H "Content-Type: application/json" \
  -d @test-alert.json

# Check response and logs
kubectl logs -n agent-sre -l app=sre-agent --tail=50
```

### Test 2: Trigger Real Alert

Create a simple alert rule to test end-to-end:

```yaml
# test-alert-rule.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: test-investigation
  namespace: prometheus
spec:
  groups:
    - name: test.rules
      interval: 30s
      rules:
        - alert: AgentSRETestAlert
          expr: vector(1)  # Always fires
          for: 1m
          labels:
            severity: warning
            namespace: test
          annotations:
            summary: "Test alert for Agent-SRE investigation"
            description: "This alert tests the automatic investigation workflow"
```

Apply and wait for it to fire:

```bash
kubectl apply -f test-alert-rule.yaml

# Wait ~2 minutes for alert to fire and investigation to complete
sleep 120

# Check for created GitHub issue
gh issue list --label "automated"

# Check Agent-SRE logs
kubectl logs -n agent-sre -l app=sre-agent --tail=100 | grep "Investigation"
```

## 📊 Example Alert Scenarios

### Scenario 1: High Error Rate

**Alert Rule:**
```yaml
- alert: HighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
  for: 5m
  labels:
    severity: critical
    namespace: homepage
  annotations:
    summary: "High error rate detected"
    description: "Error rate is {{ $value | humanizePercentage }}"
```

**What Happens:**
1. Alert fires in Prometheus
2. Alertmanager sends webhook to Agent-SRE
3. Agent-SRE creates investigation:
   - **GitHub Issue**: "Alert: HighErrorRate"
   - **Sift Analysis**: Searches logs for error patterns in last 30 minutes
   - **Root Cause**: LLM analyzes elevated error patterns
   - **Recommendations**: Specific fixes based on error types
4. GitHub issue updated with complete findings

### Scenario 2: Pod CrashLooping

**Alert Rule:**
```yaml
- alert: PodCrashLooping
  expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
  for: 5m
  labels:
    severity: warning
    namespace: "{{ $labels.namespace }}"
  annotations:
    summary: "Pod {{ $labels.pod }} is crash looping"
    description: "Pod has restarted {{ $value }} times"
```

**Investigation Includes:**
- Pod logs from crashes
- Resource utilization patterns
- Recent deployments/changes
- Related errors in other services

### Scenario 3: High Memory Usage

**Alert Rule:**
```yaml
- alert: HighMemoryUsage
  expr: container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.9
  for: 10m
  labels:
    severity: warning
    namespace: "{{ $labels.namespace }}"
  annotations:
    summary: "High memory usage in {{ $labels.pod }}"
    description: "Memory usage is {{ $value | humanizePercentage }}"
```

**Investigation Includes:**
- Memory growth patterns over time
- Potential memory leaks
- Related slow operations
- Recommendations for limits/scaling

## 🎛️ Severity Mapping

Agent-SRE maps Alertmanager severities to investigation priorities:

| Alertmanager Severity | Investigation Severity | Behavior |
|----------------------|------------------------|----------|
| `critical` | `critical` | Immediate investigation, high priority |
| `warning` | `high` | Standard investigation workflow |
| `info` | `medium` | Lower priority investigation |
| (none/other) | `medium` | Default investigation level |

## 🔧 Advanced Configuration

### Filtering Alerts

Only investigate specific alert types:

```python
# In agent.py handle_alertmanager_webhook
INVESTIGATE_ALERTS = [
    "HighErrorRate",
    "PodCrashLooping",
    "HighMemoryUsage",
    "DatabaseConnectionTimeout",
]

# Only process alerts in the list
if alert_name in INVESTIGATE_ALERTS:
    investigation_result = await investigation_workflow.investigate(...)
```

### Custom Component Mapping

Map alert labels to specific components:

```python
# In agent.py handle_alertmanager_webhook
COMPONENT_MAP = {
    "homepage": "homepage",
    "agent-sre": "agent-sre",
    "agent-bruno": "agent-bruno",
    "postgres": "database",
    "redis": "cache",
}

component = COMPONENT_MAP.get(namespace, namespace)
```

### Rate Limiting

Prevent investigation overload during alert storms:

```python
from collections import defaultdict
import time

class AlertThrottler:
    def __init__(self, max_per_hour=5):
        self.max_per_hour = max_per_hour
        self.alert_times = defaultdict(list)
    
    def should_investigate(self, alert_name: str) -> bool:
        now = time.time()
        # Clean old timestamps
        self.alert_times[alert_name] = [
            t for t in self.alert_times[alert_name]
            if now - t < 3600  # 1 hour
        ]
        
        if len(self.alert_times[alert_name]) < self.max_per_hour:
            self.alert_times[alert_name].append(now)
            return True
        return False

# Use in webhook handler
throttler = AlertThrottler(max_per_hour=3)

if throttler.should_investigate(alert_name):
    await investigation_workflow.investigate(...)
else:
    logger.warning(f"⚠️  Throttled investigation for {alert_name}")
```

### Deduplicate by Fingerprint

Avoid investigating the same alert multiple times:

```python
# Store fingerprints of investigated alerts
investigated_alerts = {}

fingerprint = alert.get("fingerprint", "")
if fingerprint in investigated_alerts:
    logger.info(f"⏭️  Skipping duplicate alert: {fingerprint}")
    continue

investigated_alerts[fingerprint] = datetime.now()
```

## 📈 Monitoring

### Metrics to Track

Add custom metrics to Agent-SRE:

```python
from prometheus_client import Counter, Histogram

investigations_total = Counter(
    'agent_sre_investigations_total',
    'Total investigations triggered',
    ['source', 'severity', 'status']
)

investigation_duration = Histogram(
    'agent_sre_investigation_duration_seconds',
    'Investigation duration in seconds'
)

# Use in webhook handler
investigations_total.labels(
    source='alertmanager',
    severity=severity,
    status='completed'
).inc()
```

### Dashboard Queries

**Investigation Rate:**
```promql
rate(agent_sre_investigations_total[5m])
```

**Average Duration:**
```promql
rate(agent_sre_investigation_duration_seconds_sum[5m]) /
rate(agent_sre_investigation_duration_seconds_count[5m])
```

**Investigations by Severity:**
```promql
sum by (severity) (agent_sre_investigations_total)
```

## 🔍 Troubleshooting

### Webhook Not Receiving Alerts

**Check Alertmanager logs:**
```bash
kubectl logs -n prometheus alertmanager-prometheus-operator-kube-p-alertmanager-0 | grep webhook
```

**Common issues:**
- Service name typo in webhook URL
- Network policy blocking traffic
- Alertmanager config syntax error

**Test connectivity:**
```bash
kubectl exec -n prometheus alertmanager-prometheus-operator-kube-p-alertmanager-0 -- \
  wget -O- http://sre-agent-service.agent-sre.svc.cluster.local:8080/health
```

### Investigations Not Created

**Check Agent-SRE logs:**
```bash
kubectl logs -n agent-sre -l app=sre-agent --tail=100
```

**Look for:**
- Webhook received log: `🔔 Received alert:`
- Investigation trigger: `🔍 Triggering investigation workflow`
- Completion: `✅ Investigation completed`

**Common issues:**
- Alert status is "resolved" (only "firing" triggers investigations)
- GitHub token not configured
- MCP server unavailable
- Ollama LLM unreachable

### GitHub Issues Not Created

**Verify GitHub token:**
```bash
kubectl get secret agent-sre-secrets -n agent-sre -o jsonpath='{.data.github-token}' | base64 -d
```

**Test manually:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://api.github.com/repos/brunovlucena/homelab/issues
```

### Sift Analysis Fails

**Check MCP server:**
```bash
kubectl logs -n agent-sre -l app=mcp-server --tail=50
```

**Verify Loki/Tempo:**
```bash
kubectl get pods -n loki
kubectl get pods -n tempo
```

## 🎉 Example Workflow

Here's a complete example of what happens when an alert fires:

### 1. Alert Fires
```
Prometheus detects: HTTP error rate > 5%
Alert: HighHTTPErrorRate
Severity: critical
Namespace: homepage
```

### 2. Alertmanager Webhook
```json
POST http://sre-agent-service.agent-sre.svc.cluster.local:8080/webhook/alert
{
  "alerts": [{
    "status": "firing",
    "labels": {
      "alertname": "HighHTTPErrorRate",
      "severity": "critical",
      "namespace": "homepage"
    },
    "annotations": {
      "summary": "HTTP error rate is 8.5%",
      "description": "Error rate exceeded threshold"
    }
  }]
}
```

### 3. Investigation Triggered
```
Agent-SRE receives webhook
→ Creates GitHub issue #31
→ Runs Sift investigation:
  - Searches Loki for error patterns in homepage logs
  - Analyzes Tempo traces for slow requests
→ LLM analyzes findings:
  - Root cause: Database connection timeout
  - Evidence: 450 "connection timeout" errors in 30min
  - Related: homepage-api pods showing high latency
```

### 4. Results Posted
```
GitHub Issue #31 updated with:
- Error patterns table (10 elevated patterns)
- Slow operations table (5 operations)
- Root cause analysis
- 5 actionable recommendations:
  1. Increase database connection pool size
  2. Add database connection timeout alerts
  3. Review recent database changes
  4. Scale database replica
  5. Add circuit breaker to API
```

### 5. Team Notified
```
Slack notification (via Jamie):
"🚨 New investigation: HighHTTPErrorRate
GitHub: https://github.com/brunovlucena/homelab/issues/31
Root cause: Database connection timeout
Recommendations: 5 actions identified"
```

## 📚 Related Documentation

- [Trigger Investigation Guide](TRIGGER_INVESTIGATION_GUIDE.md) - Manual API triggers
- [Automated Investigation Guide](AUTOMATED_INVESTIGATION_GUIDE.md) - Architecture details
- [Sift Implementation](SIFT_IMPLEMENTATION.md) - How Sift works

---

**Questions?** Check the logs or open an issue!



