# 🌐 CloudEvents Integration for AI Agents

**Version**: 1.0.1  
**Last Updated**: December 4, 2025

---

## 📖 Overview

This document describes how AI agents in the homelab integrate with the CloudEvents ecosystem. The platform uses CloudEvents (v1.0) as the standard event format for event-driven communication.

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         HOMELAB AI EVENT ECOSYSTEM                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                      EVENT PRODUCERS                                  │   │
│  │                                                                       │   │
│  │  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐         │   │
│  │  │ knative-lambda │  │ agent-contracts│  │ External APIs  │         │   │
│  │  │    operator    │  │ contract-fetch │  │  (webhooks)    │         │   │
│  │  └───────┬────────┘  └───────┬────────┘  └───────┬────────┘         │   │
│  │          │                   │                   │                   │   │
│  └──────────┼───────────────────┼───────────────────┼───────────────────┘   │
│             │                   │                   │                       │
│             ▼                   ▼                   ▼                       │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                      RABBITMQ BROKER                                  │   │
│  │                                                                       │   │
│  │  Event Types:                                                         │   │
│  │  • io.knative.lambda.*      (Lambda lifecycle)                       │   │
│  │  • io.homelab.contract.*    (Smart contracts)                        │   │
│  │  • io.homelab.vuln.*        (Vulnerabilities)                        │   │
│  │  • io.homelab.audit.*       (Security audit)                         │   │
│  │  • io.homelab.alert.*       (Alert dispatch)                         │   │
│  │                                                                       │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│             │                   │                   │                       │
│             ▼                   ▼                   ▼                       │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                      EVENT CONSUMERS (AI Agents)                      │   │
│  │                                                                       │   │
│  │  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐         │   │
│  │  │  vuln-scanner  │  │exploit-generator│ │alert-dispatcher│         │   │
│  │  │ (LLM analysis) │  │  (LLM exploit)  │  │ (multi-channel)│         │   │
│  │  └────────────────┘  └────────────────┘  └────────────────┘         │   │
│  │                                                                       │   │
│  │  ┌────────────────┐                                                   │   │
│  │  │ notifi-adapter │  CloudEvents → Alertmanager webhook bridge       │   │
│  │  └────────────────┘                                                   │   │
│  │                                                                       │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 🤖 AI Agent Event Types

### agent-auditor

LLM security testing agent that monitors Lambda functions and AI systems for vulnerabilities.

**Consumed Events**:

| Event Type | Source | Description |
|------------|--------|-------------|
| `io.knative.lambda.lifecycle.function.created` | knative-lambda-operator | New function deployed |
| `io.knative.lambda.lifecycle.function.updated` | knative-lambda-operator | Function code changed |
| `io.knative.lambda.notification.audit.change` | knative-lambda-operator | Config changes |
| `io.knative.lambda.invoke.async` | External | Function invocations |

**Produced Events**:

| Event Type | Destination | Description |
|------------|-------------|-------------|
| `io.homelab.audit.security.passed` | Loki | Security test passed |
| `io.homelab.audit.security.failed` | Alertmanager (via notifi-adapter) | Security test failed |
| `io.homelab.audit.anomaly.detected` | Alertmanager (via notifi-adapter) | Model anomaly |

**Prometheus Metrics** (for Alertmanager):

```python
# Alertable metrics
agent_auditor_attacks_successful_total{attack_type, target}  # → CRITICAL
agent_auditor_prompt_injections_blocked_total{source}        # → WARNING
agent_auditor_anomaly_score{model, namespace}                # → WARNING if > 0.9
```

---

### agent-contracts

Smart contract vulnerability scanner and exploit generator pipeline with 4 serverless components:

```
contract-fetcher → vuln-scanner → exploit-generator → alert-dispatcher
```

**Pipeline Components**:

| Component | Consumes | Produces | Description |
|-----------|----------|----------|-------------|
| **contract-fetcher** | `io.homelab.block.new` | `io.homelab.contract.created` | Fetches contract source from block explorers |
| **vuln-scanner** | `io.homelab.contract.created` | `io.homelab.vuln.found` | Slither + LLM vulnerability analysis |
| **exploit-generator** | `io.homelab.vuln.found` | `io.homelab.exploit.validated` | Generates PoC exploits on Anvil forks |
| **alert-dispatcher** | `io.homelab.exploit.validated` | `io.homelab.alert.sent` | Multi-channel alerting (Grafana, Telegram, Discord) |

**notifi-adapter**: Bridge component that converts CloudEvents to Alertmanager webhook format for integration with Prometheus alerting.

---

## 🔔 Alertmanager Integration

**⚠️ Alertmanager does NOT support CloudEvents natively.**

### Recommended Pattern: Prometheus Metrics + PrometheusRule

Instead of sending CloudEvents directly to Alertmanager, emit Prometheus metrics and let PrometheusRule evaluate alerting conditions:

```
┌──────────────┐     metrics      ┌──────────────┐    alerts    ┌──────────────┐
│  AI Agent    │ ───────────────► │  Prometheus  │ ────────────►│ Alertmanager │
│              │  (scrape)        │              │ (rules)      │              │
└──────────────┘                  └──────────────┘              └──────────────┘
```

**Example PrometheusRule for agent-auditor**:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: llm-security-alerts
  namespace: agent-auditor
spec:
  groups:
    - name: agent-auditor
      rules:
        - alert: LLMSecurityDefenseBypassDetected
          expr: |
            (agent_auditor_attacks_successful_total / agent_auditor_attacks_total) * 100 > 0
          for: 1m
          labels:
            severity: critical
          annotations:
            summary: "CRITICAL: LLM security defense bypassed"

        - alert: HighPromptInjectionRate
          expr: rate(agent_auditor_prompt_injections_blocked_total[5m]) > 10
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "High rate of prompt injection attempts"
```

### Alternative Pattern: CloudEvents → notifi-adapter → Alertmanager

If you need to trigger alerts from CloudEvents:

```python
# notifi-adapter CloudEvents handler
from cloudevents.http import from_http
import requests

ALERTMANAGER_URL = "http://alertmanager.prometheus:9093"

@app.route('/cloudevents', methods=['POST'])
def receive_cloudevent():
    event = from_http(request.headers, request.data)
    
    if event['type'] == 'io.homelab.audit.security.failed':
        alert = {
            'labels': {
                'alertname': 'AISecurityTestFailed',
                'severity': 'critical',
                'namespace': event.data.get('namespace'),
            },
            'annotations': {
                'summary': event.data.get('summary'),
                'description': event.data.get('description'),
            },
            'startsAt': event['time'],
        }
        requests.post(f'{ALERTMANAGER_URL}/api/v1/alerts', json=[alert])
    
    return '', 202
```

---

## 📝 Event Schema Examples

### agent-auditor Security Event

```json
{
  "specversion": "1.0",
  "id": "audit-123",
  "source": "io.homelab/agent-auditor",
  "type": "io.homelab.audit.security.failed",
  "subject": "knative-lambda/hello-python",
  "time": "2025-12-04T10:30:00Z",
  "datacontenttype": "application/json",
  "data": {
    "testName": "prompt-injection-test",
    "attackType": "jailbreak",
    "target": "hello-python",
    "namespace": "knative-lambda",
    "severity": "critical",
    "summary": "Prompt injection attack succeeded",
    "description": "Model responded to jailbreak attempt",
    "evidence": {
      "prompt": "[REDACTED]",
      "response": "[REDACTED]",
      "confidenceScore": 0.95
    }
  }
}
```

### Lambda Lifecycle Event (for auditing)

```json
{
  "specversion": "1.0",
  "id": "lifecycle-456",
  "source": "io.knative.lambda/operator/knative-lambda/hello-python",
  "type": "io.knative.lambda.lifecycle.function.updated",
  "subject": "knative-lambda/hello-python",
  "time": "2025-12-04T10:30:00Z",
  "datacontenttype": "application/json",
  "data": {
    "name": "hello-python",
    "namespace": "knative-lambda",
    "runtime": {
      "language": "python",
      "version": "3.11",
      "handler": "main.handler"
    },
    "phase": "Ready",
    "generation": 2,
    "observedGeneration": 2
  }
}
```

---

## 🔗 Related Documentation

- [CloudEvents Specification](../flux/infrastructure/knative-lambda/docs/04-architecture/CLOUDEVENTS_SPECIFICATION.md)
- [agent-contracts Architecture](./agent-contracts/docs/ARCHITECTURE.md)
- [agent-auditor Prometheus Rules](../vault/_homelab/flux/infrastructure/alerts/apps/agent-auditor/prometheus-rules.yaml)

---

## 📊 Monitoring

All AI agents should expose the following standard metrics:

```python
# Standard metrics for all AI agents
from prometheus_client import Counter, Histogram, Gauge

# CloudEvents received
cloudevents_received_total = Counter(
    'cloudevents_received_total',
    'Total CloudEvents received',
    ['type', 'source']
)

# CloudEvents processing time
cloudevents_processing_seconds = Histogram(
    'cloudevents_processing_seconds',
    'CloudEvent processing duration',
    ['type']
)

# LLM inference metrics
llm_inference_total = Counter(
    'llm_inference_total',
    'Total LLM inference calls',
    ['model', 'operation', 'status']
)

llm_inference_seconds = Histogram(
    'llm_inference_seconds',
    'LLM inference duration',
    ['model', 'operation']
)
```

---

**Maintainer**: Platform Team  
**Review Cycle**: Quarterly

