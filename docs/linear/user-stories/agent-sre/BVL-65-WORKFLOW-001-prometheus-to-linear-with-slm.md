# 🔄 WORKFLOW-001: PrometheusRule → Linear Issue Creation with SLM

**Linear URL**: https://linear.app/bvlucena/issue/BVL-266/workflow-001-prometheusrule-→-linear-issue-creation-with-slm

**Linear URL**: https://linear.app/bvlucena/issue/BVL-266/workflow-001-prometheusrule-→-linear-issue-creation-with-slm  

---

## 📋 User Story

**As an** SRE Engineer  
**I want** agent-sre to automatically create Linear issues from PrometheusRule alerts using Service Level Management (SLM) data  
**So that** incidents are tracked and resolved efficiently with proper priority based on SLM targets


---


## 🎯 Acceptance Criteria

> **Note**: Features are already implemented. This ticket focuses on **validation** to ensure correctness, reliability, and production readiness.


- [ ] PrometheusRule triggers agent-sre via prometheus-events
- [ ] Agent-sre extracts SLM data (SLOs, SLIs, error budgets)
- [ ] Agent-sre creates Linear issue with SLM context
- [ ] Issue priority determined by SLM violation severity
- [ ] Issue includes alert details, labels, annotations
- [ ] Issue linked to relevant SLO/SLI
- [ ] Issue automatically assigned to on-call engineer
- [ ] Issue updated as remediation progresses
- [ ] Issue closed when alert resolves

---

## 🔐 Security Acceptance Criteria

- [ ] Linear API authentication required (API tokens)
- [ ] Rate limiting on Linear API calls (prevent DoS)
- [ ] Input validation for all alert data before creating issues
- [ ] Secrets management for Linear API keys
- [ ] Audit logging for all Linear issue operations
- [ ] Error messages don't leak sensitive information
- [ ] TLS/HTTPS enforced for Linear API communications
- [ ] Access control for Linear issue creation
- [ ] Security testing included in CI/CD pipeline
- [ ] Threat model reviewed and documented

## 🔄 Complete Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│      PROMETHEUSRULE → LINEAR ISSUE WORKFLOW WITH SLM                   │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ⏱️  t=0s: PROMETHEUSRULE FIRES                                      │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  PrometheusRule: PodCPUHigh                          │            │
│  │                                                      │            │
│  │  apiVersion: monitoring.coreos.com/v1                 │            │
│  │  kind: PrometheusRule                                 │            │
│  │  metadata:                                            │            │
│  │    name: pod-cpu-high                                  │            │
│  │  spec:                                                │            │
│  │    groups:                                            │            │
│  │      - name: pod-alerts                                │            │
│  │        rules:                                         │            │
│  │          - alert: PodCPUHigh                           │            │
│  │            expr: |                                     │            │
│  │              rate(container_cpu_usage_seconds_total{  │            │
│  │                pod=~".+",                             │            │
│  │                namespace="production"                 │            │
│  │              }[5m]) > 0.8                             │            │
│  │            for: 5m                                     │            │
│  │            labels:                                     │            │
│  │              severity: warning                         │            │
│  │              slo: availability                         │            │
│  │            annotations:                                │            │
│  │              summary: "Pod CPU usage high"             │            │
│  │              description: "Pod {{ $labels.pod }} CPU > 80%"│            │
│  │              lambda_function: "scale-pod"              │            │
│  │              lambda_parameters: |                       │            │
│  │                {"pod": "{{ $labels.pod }}", "namespace": "{{ $labels.namespace }}"}│            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=1s: PROMETHEUS FIRES ALERT                                     │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Prometheus evaluates rule:                          │            │
│  │  - Condition: CPU > 80% for 5 minutes                 │            │
│  │  - Status: FIRING                                     │            │
│  │  - Labels: {pod: "app-xyz", namespace: "production"}   │            │
│  │  - Annotations: {summary, description, lambda_function}│            │
│  │                                                      │            │
│  │  Alert sent to:                                       │            │
│  │  - Alertmanager                                       │            │
│  │  - prometheus-events (CloudEvent source)              │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=2s: PROMETHEUS-EVENTS CONVERTS TO CLOUDEVENT                  │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  prometheus-events converts alert to CloudEvent:     │            │
│  │                                                      │            │
│  │  CloudEvent:                                          │            │
│  │  {                                                    │            │
│  │    "type": "io.homelab.prometheus.alert.fired",       │            │
│  │    "source": "prometheus-events",                     │            │
│  │    "subject": "PodCPUHigh",                           │            │
│  │    "id": "alert-12345",                                │            │
│  │    "time": "2026-01-15T10:45:00Z",                    │            │
│  │    "data": {                                           │            │
│  │      "alertname": "PodCPUHigh",                        │            │
│  │      "status": "firing",                               │            │
│  │      "labels": {                                       │            │
│  │        "alertname": "PodCPUHigh",                      │            │
│  │        "pod": "app-xyz",                               │            │
│  │        "namespace": "production",                      │            │
│  │        "severity": "warning",                          │            │
│  │        "slo": "availability"                           │            │
│  │      },                                                │            │
│  │      "annotations": {                                  │            │
│  │        "summary": "Pod CPU usage high",                │            │
│  │        "description": "Pod app-xyz CPU > 80%",         │            │
│  │        "lambda_function": "scale-pod",                 │            │
│  │        "lambda_parameters": '{"pod": "app-xyz", "namespace": "production"}'│            │
│  │      },                                                │            │
│  │      "startsAt": "2026-01-15T10:40:00Z",               │            │
│  │      "endsAt": null                                    │            │
│  │    }                                                   │            │
│  │  }                                                    │            │
│  │                                                      │            │
│  │  CloudEvent sent to:                                   │            │
│  │  - agent-sre service (CloudEvent sink)                │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=3s: AGENT-SRE RECEIVES CLOUDEVENT                             │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE receives CloudEvent:                      │            │
│  │                                                      │            │
│  │  @app.post("/")                                      │            │
│  │  async def handle_cloudevent(request: Request):      │            │
│  │      event = await request.json()                    │            │
│  │      alert_data = event["data"]                       │            │
│  │                                                      │            │
│  │  Extract alert information:                           │            │
│  │  - alertname: "PodCPUHigh"                            │            │
│  │  - labels: {pod, namespace, severity, slo}            │            │
│  │  - annotations: {summary, description, lambda_function}│            │
│  │  - slo: "availability" (from labels)                  │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=4s: QUERY SLM DATA                                            │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE queries SLM data for SLO:                 │            │
│  │                                                      │            │
│  │  SLM Query:                                           │            │
│  │  - SLO: "availability"                                │            │
│  │  - SLI: "uptime"                                      │            │
│  │  - Target: 99.9%                                      │            │
│  │  - Current: 99.85%                                    │            │
│  │  - Error Budget: 0.1% (remaining)                     │            │
│  │  - Error Budget Burn Rate: 0.05% (this alert)         │            │
│  │                                                      │            │
│  │  SLM Context:                                         │            │
│  │  {                                                    │            │
│  │    "slo": "availability",                             │            │
│  │    "sli": "uptime",                                   │            │
│  │    "target": 0.999,                                   │            │
│  │    "current": 0.9985,                                 │            │
│  │    "error_budget_remaining": 0.001,                   │            │
│  │    "error_budget_burn_rate": 0.0005,                  │            │
│  │    "violation_severity": "high",                      │            │
│  │    "on_call_engineer": "alice@example.com"            │            │
│  │  }                                                    │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=5s: DETERMINE PRIORITY                                        │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE determines issue priority:                │            │
│  │                                                      │            │
│  │  Priority Calculation:                                │            │
│  │  - Base priority: From alert severity ("warning" = High)│            │
│  │  - SLM adjustment: Error budget burn rate             │            │
│  │    - High burn rate (>0.05%) → Increase priority      │            │
│  │    - Low burn rate (<0.01%) → Decrease priority       │            │
│  │                                                      │            │
│  │  Priority Mapping:                                    │            │
│  │  - Urgent (1): Critical + High SLM violation          │            │
│  │  - High (2): Warning + High SLM violation             │            │
│  │  - Normal (3): Warning + Low SLM violation            │            │
│  │  - Low (4): Info + Any SLM violation                  │            │
│  │                                                      │            │
│  │  Result:                                              │            │
│  │  - Priority: High (2)                                 │            │
│  │  - Reason: "Warning severity + 0.05% error budget burn"│            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=6s: CREATE LINEAR ISSUE                                       │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE creates Linear issue:                     │            │
│  │                                                      │            │
│  │  Issue Details:                                       │            │
│  │  - Title: "[Alert] PodCPUHigh - app-xyz (SLO: availability)"│            │
│  │  - Description: |                                      │            │
│  │      **Alert Details**                                │            │
│  │      - Alert: PodCPUHigh                              │            │
│  │      - Severity: warning                              │            │
│  │      - Started: 2026-01-15T10:40:00Z                  │            │
│  │      - Pod: app-xyz                                   │            │
│  │      - Namespace: production                          │            │
│  │                                                      │            │
│  │      **SLM Context**                                  │            │
│  │      - SLO: availability                              │            │
│  │      - SLI: uptime                                    │            │
│  │      - Target: 99.9%                                  │            │
│  │      - Current: 99.85%                                │            │
│  │      - Error Budget Remaining: 0.1%                   │            │
│  │      - Error Budget Burn Rate: 0.05%                  │            │
│  │      - Violation Severity: high                       │            │
│  │                                                      │            │
│  │      **Description**                                  │            │
│  │      Pod app-xyz CPU > 80%                            │            │
│  │                                                      │            │
│  │      **Labels**                                       │            │
│  │      - pod: app-xyz                                   │            │
│  │      - namespace: production                          │            │
│  │      - severity: warning                              │            │
│  │      - slo: availability                              │            │
│  │                                                      │            │
│  │      **Remediation**                                  │            │
│  │      - LambdaFunction: scale-pod                      │            │
│  │      - Parameters: {pod: "app-xyz", namespace: "production"}│            │
│  │                                                      │            │
│  │      **Correlation ID**: alert-12345                  │            │
│  │                                                      │            │
│  │      ---                                              │            │
│  │      *Created by agent-sre*                           │            │
│  │                                                      │            │
│  │  - Team: SRE (from configuration)                     │            │
│  │  - Priority: High (2)                                 │            │
│  │  - Assignee: alice@example.com (on-call engineer)     │            │
│  │  - Labels: ["alert", "prometheus", "slo-violation"]   │            │
│  │  - State: Open                                        │            │
│  │                                                      │            │
│  │  Issue Created:                                       │            │
│  │  - Issue ID: BVL-66                                   │            │
│  │  - Issue URL: https://linear.app/bvlucena/issue/BVL-66│            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=7s: LINK TO SLO                                               │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE links issue to SLO:                      │            │
│  │                                                      │            │
│  │  Add Link:                                            │            │
│  │  - URL: https://grafana.example.com/d/slo-availability│            │
│  │  - Title: "SLO: Availability (99.9% target)"          │            │
│  │                                                      │            │
│  │  Add Comment:                                         │            │
│  │  "This alert impacts SLO: availability. Current SLI: 99.85%, Target: 99.9%, Error Budget Burn Rate: 0.05%."│            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=8s: NOTIFY ON-CALL                                            │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE notifies on-call engineer:               │            │
│  │                                                      │            │
│  │  Notification:                                        │            │
│  │  - Channel: Slack #sre-alerts                         │            │
│  │  - Message:                                           │            │
│  │    "[Alert] PodCPUHigh - app-xyz                      │            │
│  │     Priority: High                                    │            │
│  │     SLO Impact: availability (0.05% error budget burn)│            │
│  │     Linear Issue: https://linear.app/bvlucena/issue/BVL-66"│            │
│  │                                                      │            │
│  │  - Channel: Email                                     │            │
│  │  - Recipient: alice@example.com                       │            │
│  │  - Subject: "[High Priority] PodCPUHigh - app-xyz"   │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=10s: SELECT REMEDIATION                                       │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE selects remediation:                     │            │
│  │                                                      │            │
│  │  Remediation Selection:                               │            │
│  │  - Method: Static annotation (fast path)             │            │
│  │  - LambdaFunction: "scale-pod"                        │            │
│  │  - Parameters: {pod: "app-xyz", namespace: "production"}│            │
│  │  - Confidence: 1.0                                    │            │
│  │                                                      │            │
│  │  (Alternative: Use TRM/RAG/Few-shot/AI if no annotation)│            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=11s: UPDATE ISSUE WITH REMEDIATION                            │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE updates Linear issue:                    │            │
│  │                                                      │            │
│  │  Add Comment:                                         │            │
│  │  "**Remediation Selected**                            │            │
│  │                                                      │            │
│  │  - LambdaFunction: scale-pod                          │            │
│  │  - Parameters: {pod: "app-xyz", namespace: "production"}│            │
│  │  - Method: Static annotation                          │            │
│  │  - Confidence: 1.0                                    │            │
│  │                                                      │            │
│  │  Executing remediation..."                            │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=12s: EXECUTE REMEDIATION                                      │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE calls LambdaFunction:                     │            │
│  │                                                      │            │
│  │  HTTP POST:                                           │            │
│  │  - URL: http://scale-pod.ai.svc.cluster.local/       │            │
│  │  - Body: {                                            │            │
│  │      "pod": "app-xyz",                                │            │
│  │      "namespace": "production"                        │            │
│  │    }                                                  │            │
│  │                                                      │            │
│  │  Remediation executed successfully                   │            │
│  │  - Pod scaled from 2 to 3 replicas                   │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=17s: VERIFY REMEDIATION                                       │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE verifies remediation:                    │            │
│  │                                                      │            │
│  │  Verification:                                        │            │
│  │  - Query Prometheus: CPU usage < 80%                  │            │
│  │  - Check pod status: Running                         │            │
│  │  - Check replica count: 3                            │            │
│  │                                                      │            │
│  │  Result:                                              │            │
│  │  - CPU usage: 45% (< 80%) ✅                          │            │
│  │  - Pod status: Running ✅                             │            │
│  │  - Replica count: 3 ✅                                │            │
│  │  - Remediation successful ✅                          │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=18s: UPDATE ISSUE WITH VERIFICATION                           │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE updates Linear issue:                    │            │
│  │                                                      │            │
│  │  Add Comment:                                         │            │
│  │  "**Remediation Verified**                            │            │
│  │                                                      │            │
│  │  ✅ CPU usage normalized: 45% (< 80%)                 │            │
│  │  ✅ Pod status: Running                               │            │
│  │  ✅ Replica count: 3                                  │            │
│  │                                                      │            │
│  │  Alert should resolve within 5 minutes..."            │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=25s: ALERT RESOLVES                                           │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Prometheus alert resolves:                          │            │
│  │                                                      │            │
│  │  Alert Status:                                        │            │
│  │  - Status: RESOLVED                                   │            │
│  │  - EndsAt: 2026-01-15T10:45:00Z                      │            │
│  │                                                      │            │
│  │  CloudEvent: io.homelab.prometheus.alert.resolved     │            │
│  │  → Agent-SRE receives resolution event                │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=26s: CLOSE LINEAR ISSUE                                       │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE closes Linear issue:                     │            │
│  │                                                      │            │
│  │  Update Issue:                                        │            │
│  │  - State: Completed                                   │            │
│  │  - Resolution: "Remediation successful: Pod scaled from 2 to 3 replicas"│            │
│  │                                                      │            │
│  │  Add Comment:                                         │            │
│  │  "**Alert Resolved**                                  │            │
│  │                                                      │            │
│  │  Alert resolved at 2026-01-15T10:45:00Z               │            │
│  │  Remediation: Pod scaled from 2 to 3 replicas         │            │
│  │  Total time to resolution: 5 minutes                  │            │
│  │  SLO impact: 0.05% error budget restored              │            │
│  │                                                      │            │
│  │  Issue closed."                                       │            │
│  └──────────────────────────────────────────────────────┘            │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Implementation Details

### 1. SLM Data Service

```python
# src/sre_agent/slm_service.py
from typing import Dict, Any, Optional
import httpx

class SLMService:
    """Service for querying Service Level Management data."""
    
    def __init__(self, prometheus_url: str):
        self.prometheus_url = prometheus_url
        self.client = httpx.AsyncClient()
    
    async def get_slm_context(
        self,
        slo_name: str
    ) -> Dict[str, Any]:
        """
        Get SLM context for a given SLO.
        
        Args:
            slo_name: Name of the SLO (e.g., "availability")
            
        Returns:
            SLM context with SLO, SLI, target, current, error budget
        """
        # Query Prometheus for SLM data
        slm_queries = {
            "slo_target": f'slo:target{{{slo_name=~"{slo_name}"}}}}',
            "sli_current": f'sli:current{{{slo_name=~"{slo_name}"}}}}',
            "error_budget_remaining": f'slo:error_budget_remaining{{{slo_name=~"{slo_name}"}}}}',
            "error_budget_burn_rate": f'slo:error_budget_burn_rate{{{slo_name=~"{slo_name}"}}}}'
        }
        
        slm_data = {}
        for key, query in slm_queries.items():
            response = await self.client.get(
                f"{self.prometheus_url}/api/v1/query",
                params={"query": query}
            )
            result = response.json()
            if result["status"] == "success" and result["data"]["result"]:
                slm_data[key] = float(result["data"]["result"][0]["value"][1])
            else:
                slm_data[key] = None
        
        # Calculate violation severity
        violation_severity = self._calculate_violation_severity(slm_data)
        
        # Get on-call engineer
        on_call_engineer = await self._get_on_call_engineer()
        
        return {
            "slo": slo_name,
            "sli": self._get_sli_name(slo_name),
            "target": slm_data.get("slo_target", 0.999),
            "current": slm_data.get("sli_current", 1.0),
            "error_budget_remaining": slm_data.get("error_budget_remaining", 0.001),
            "error_budget_burn_rate": slm_data.get("error_budget_burn_rate", 0.0),
            "violation_severity": violation_severity,
            "on_call_engineer": on_call_engineer
        }
    
    def _calculate_violation_severity(
        self,
        slm_data: Dict[str, Any]
    ) -> str:
        """Calculate violation severity from SLM data."""
        burn_rate = slm_data.get("error_budget_burn_rate", 0.0)
        
        if burn_rate > 0.05:
            return "critical"
        elif burn_rate > 0.01:
            return "high"
        elif burn_rate > 0.001:
            return "medium"
        else:
            return "low"
```

### 2. Linear Issue Creation

```python
# src/sre_agent/linear_handler.py (enhance existing)
async def create_alert_ticket_with_slm(
    self,
    alert: Dict[str, Any],
    slm_context: Dict[str, Any],
    correlation_id: Optional[str] = None
) -> Optional[str]:
    """
    Create Linear issue with SLM context.
    
    Args:
        alert: Alert data from CloudEvent
        slm_context: SLM context from SLMService
        correlation_id: Correlation ID for tracing
        
    Returns:
        Linear issue URL or None
    """
    # Determine priority based on alert severity and SLM violation
    priority = self._calculate_priority(
        alert.get("labels", {}).get("severity", "info"),
        slm_context.get("violation_severity", "low")
    )
    
    # Build issue title
    alertname = alert.get("labels", {}).get("alertname", "Unknown")
    pod = alert.get("labels", {}).get("pod", "unknown")
    slo = slm_context.get("slo", "")
    title = f"[Alert] {alertname} - {pod}"
    if slo:
        title += f" (SLO: {slo})"
    
    # Build issue description
    description_parts = [
        "**Alert Details**",
        f"- Alert: {alertname}",
        f"- Severity: {alert.get('labels', {}).get('severity', 'unknown')}",
        f"- Started: {alert.get('startsAt', 'unknown')}",
    ]
    
    # Add labels
    labels = alert.get("labels", {})
    for key, value in labels.items():
        if key != "alertname":
            description_parts.append(f"- {key}: {value}")
    
    # Add SLM context
    description_parts.extend([
        "",
        "**SLM Context**",
        f"- SLO: {slm_context.get('slo', 'unknown')}",
        f"- SLI: {slm_context.get('sli', 'unknown')}",
        f"- Target: {slm_context.get('target', 0.0) * 100:.2f}%",
        f"- Current: {slm_context.get('current', 0.0) * 100:.2f}%",
        f"- Error Budget Remaining: {slm_context.get('error_budget_remaining', 0.0) * 100:.2f}%",
        f"- Error Budget Burn Rate: {slm_context.get('error_budget_burn_rate', 0.0) * 100:.2f}%",
        f"- Violation Severity: {slm_context.get('violation_severity', 'low')}",
    ])
    
    # Add annotations
    annotations = alert.get("annotations", {})
    if annotations.get("description"):
        description_parts.extend([
            "",
            "**Description**",
            annotations["description"]
        ])
    
    # Add remediation
    if annotations.get("lambda_function"):
        description_parts.extend([
            "",
            "**Remediation**",
            f"- LambdaFunction: {annotations['lambda_function']}",
            f"- Parameters: {annotations.get('lambda_parameters', '{}')}"
        ])
    
    # Add correlation ID
    if correlation_id:
        description_parts.append(f"\n**Correlation ID**: `{correlation_id}`")
    
    description_parts.append("\n---\n*Created by agent-sre*")
    description = "\n".join(description_parts)
    
    # Create issue
    issue = await self.client.create_issue(
        title=title,
        description=description,
        team_id=self.team_id,
        priority=priority,
        assignee_id=slm_context.get("on_call_engineer_id")
    )
    
    # Link to SLO dashboard
    if slm_context.get("slo"):
        await self.client.create_link(
            issue_id=issue["id"],
            url=f"https://grafana.example.com/d/slo-{slm_context['slo']}",
            title=f"SLO: {slm_context['slo']} ({slm_context.get('target', 0.0) * 100:.2f}% target)"
        )
    
    return issue.get("url")
```

---

## 📚 References

- [Agent-SRE Linear Integration](../../docs/linear-agent-integration.md)
- [SLM Best Practices](../../docs/slm-best-practices.md)

---

## ✅ Definition of Done

- [ ] PrometheusRule triggers agent-sre via prometheus-events
- [ ] SLM data querying service implemented
- [ ] Linear issue creation with SLM context working
- [ ] Priority calculation based on SLM violation severity
- [ ] Issue linking to SLO dashboards working
- [ ] On-call engineer assignment implemented
- [ ] Issue updates as remediation progresses
- [ ] Issue closure when alert resolves
- [ ] Integration tests passing
- [ ] Documentation updated

---

**Related Stories**:
- [WORKFLOW-002: Lambda Function Annotation Discovery](./BVL-66-WORKFLOW-002-lambda-annotation-discovery.md)
- [WORKFLOW-003: Enriched Issue Updates](./BVL-67-WORKFLOW-003-enriched-issue-updates.md)
- [AI-001: Data Formulator Integration](./BVL-61-AI-001-data-formulator-visualization.md)



---

**Last Updated**: January 08, 2026
**Owner**: SRE Team
**Status**: Validation Required