# 📊 AI-001: Data Formulator Integration for Observability Visualization

**Linear URL**: https://linear.app/bvlucena/issue/BVL-61/ai-001-data-formulator-visualization  

---

## 📋 User Story

**As an** SRE Engineer  
**I want** agent-sre to use Data Formulator to create rich visualizations from metrics, logs, and traces  
**So that** I can understand system behavior patterns and make informed decisions about remediation


---


## 🎯 Acceptance Criteria

> **Note**: Features are already implemented. This ticket focuses on **validation** to ensure correctness, reliability, and production readiness.


- [ ] Data Formulator integrated into agent-sre workflow
- [ ] Agent-sre can query Prometheus metrics and visualize them using Data Formulator
- [ ] Agent-sre can query Loki logs and create visualizations
- [ ] Agent-sre can query Tempo traces and visualize trace patterns
- [ ] Visualizations are embedded in Linear issue comments as markdown/images
- [ ] Visualizations help identify root causes of incidents
- [ ] Visualizations are generated automatically when agent-sre updates Linear issues
- [ ] Support for goal-driven exploration with AI agent recommendations

---

## 🔐 Security Acceptance Criteria

- [ ] Data Formulator API access requires authentication
- [ ] Sensitive data redacted from visualizations
- [ ] Access control for Data Formulator service
- [ ] Audit logging for visualization generation
- [ ] Rate limiting on Data Formulator queries
- [ ] Input validation for all visualization queries
- [ ] Secrets management for Data Formulator credentials
- [ ] TLS/HTTPS enforced for Data Formulator communications
- [ ] Security testing included in CI/CD pipeline
- [ ] Threat model reviewed and documented

## 🔄 Complete Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│              DATA FORMULATOR INTEGRATION WORKFLOW                    │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ⏱️  t=0s: PROMETHEUS ALERT FIRES                                    │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Prometheus Alert: PodCPUHigh                        │            │
│  │  Severity: warning                                   │            │
│  │  Labels: {pod: "app-xyz", namespace: "production"}   │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=1s: AGENT-SRE RECEIVES CLOUDEVENT                             │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE extracts alert information:               │            │
│  │  - alertname: PodCPUHigh                             │            │
│  │  - labels: {pod, namespace}                          │            │
│  │  - annotations: {lambda_function: "scale-pod"}       │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=2s: CREATE LINEAR ISSUE                                       │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE creates Linear issue:                     │            │
│  │  Title: "[Alert] PodCPUHigh - app-xyz"               │            │
│  │  Description: Initial alert details                  │            │
│  │  Team: SRE                                           │            │
│  │  Priority: High                                      │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=5s: DATA FORMULATOR QUERY METRICS                             │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE queries Prometheus via Data Formulator:   │            │
│  │                                                      │            │
│  │  1. Query CPU metrics for pod "app-xyz":            │            │
│  │     rate(container_cpu_usage_seconds_total{          │            │
│  │       pod="app-xyz", namespace="production"          │            │
│  │     }[5m])                                           │            │
│  │                                                      │            │
│  │  2. Query memory metrics:                           │            │
│  │     container_memory_usage_bytes{                    │            │
│  │       pod="app-xyz", namespace="production"          │            │
│  │     }                                                │            │
│  │                                                      │            │
│  │  3. Query request rate:                             │            │
│  │     rate(http_requests_total{                        │            │
│  │       pod="app-xyz"                                  │            │
│  │     }[5m])                                           │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=8s: DATA FORMULATOR GENERATES VISUALIZATIONS                  │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Data Formulator creates visualizations:             │            │
│  │                                                      │            │
│  │  📊 Chart 1: CPU Usage Over Time                    │            │
│  │     - Line chart showing CPU spike at 10:45 AM      │            │
│  │     - Annotation: "Alert fired at 10:45 AM"         │            │
│  │                                                      │            │
│  │  📊 Chart 2: Memory vs CPU Correlation              │            │
│  │     - Scatter plot showing correlation              │            │
│  │     - Insight: "Memory usage correlated with CPU"   │            │
│  │                                                      │            │
│  │  📊 Chart 3: Request Rate Over Time                 │            │
│  │     - Bar chart showing request surge               │            │
│  │     - Insight: "Traffic spike at 10:44 AM"          │            │
│  │                                                      │            │
│  │  📊 Chart 4: Multi-Metric Dashboard                 │            │
│  │     - Combined view of CPU, Memory, Requests        │            │
│  │     - Timeline shows correlation                    │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=10s: QUERY LOGS FOR ADDITIONAL CONTEXT                        │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE queries Loki logs via Data Formulator:    │            │
│  │                                                      │            │
│  │  LogQL Query:                                        │            │
│  │  {pod="app-xyz", namespace="production"}            │            │
│  │    | json                                           │            │
│  │    | line_format "{{.timestamp}} {{.level}} {{.msg}}" │            │
│  │                                                      │            │
│  │  Data Formulator extracts:                           │            │
│  │  - Error patterns in logs                           │            │
│  │  - Warning frequency                                │            │
│  │  - Log volume over time                             │            │
│  │                                                      │            │
│  │  📊 Chart 5: Log Volume Over Time                   │            │
│  │     - Shows spike in error logs at 10:44 AM         │            │
│  │                                                      │            │
│  │  📊 Chart 6: Error Pattern Analysis                 │            │
│  │     - Pie chart: Error types distribution           │            │
│  │     - Most common: "OutOfMemoryError"               │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=12s: QUERY TRACES FOR DISTRIBUTED TRACING                     │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE queries Tempo traces via Data Formulator: │            │
│  │                                                      │            │
│  │  Trace Query:                                        │            │
│  │  {service.name="app-xyz"}                           │            │
│  │    AND {duration>1s}                                │            │
│  │                                                      │            │
│  │  Data Formulator analyzes:                           │            │
│  │  - Slow request traces                              │            │
│  │  - Service dependencies                             │            │
│  │  - Latency percentiles                              │            │
│  │                                                      │            │
│  │  📊 Chart 7: Latency Distribution                   │            │
│  │     - Histogram showing p50, p95, p99              │            │
│  │     - Shows latency degradation at 10:44 AM         │            │
│  │                                                      │            │
│  │  📊 Chart 8: Service Dependency Map                 │            │
│  │     - Sankey diagram showing request flow           │            │
│  │     - Highlights bottleneck: "database-service"     │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=15s: AI AGENT ANALYSIS                                        │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Data Formulator AI Agent analyzes all visualizations│            │
│  │                                                      │            │
│  │  AI Insights:                                        │            │
│  │  1. "CPU spike correlates with traffic spike"        │            │
│  │  2. "Error logs show OutOfMemoryError pattern"       │            │
│  │  3. "Latency increase suggests resource exhaustion"  │            │
│  │  4. "Root cause: Insufficient memory limits"         │            │
│  │                                                      │            │
│  │  Recommended Actions:                                │            │
│  │  - Scale pod memory limits                           │            │
│  │  - Investigate memory leak in application           │            │
│  │  - Consider horizontal pod autoscaling              │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=18s: UPDATE LINEAR ISSUE WITH VISUALIZATIONS                  │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE adds enriched comment to Linear issue:    │            │
│  │                                                      │            │
│  │  📊 **Observability Analysis**                       │            │
│  │                                                      │            │
│  │  Generated using Data Formulator:                    │            │
│  │                                                      │            │
│  │  ### Metrics Analysis                                │            │
│  │  [Embedded Chart 1: CPU Usage Over Time]            │            │
│  │  [Embedded Chart 2: Memory vs CPU Correlation]      │            │
│  │  [Embedded Chart 3: Request Rate Over Time]         │            │
│  │  [Embedded Chart 4: Multi-Metric Dashboard]         │            │
│  │                                                      │            │
│  │  ### Log Analysis                                    │            │
│  │  [Embedded Chart 5: Log Volume Over Time]           │            │
│  │  [Embedded Chart 6: Error Pattern Analysis]         │            │
│  │                                                      │            │
│  │  ### Trace Analysis                                  │            │
│  │  [Embedded Chart 7: Latency Distribution]           │            │
│  │  [Embedded Chart 8: Service Dependency Map]         │            │
│  │                                                      │            │
│  │  ### AI Agent Insights                               │            │
│  │  {AI-generated insights from Data Formulator}       │            │
│  │                                                      │            │
│  │  ### Recommended Remediation                        │            │
│  │  Based on analysis, agent-sre recommends:           │            │
│  │  1. Scale pod memory limits (immediate)             │            │
│  │  2. Investigate memory leak (follow-up)             │            │
│  │                                                      │            │
│  │  *Analysis generated at 2026-01-15T10:45:18Z*       │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=20s: EXECUTE REMEDIATION                                      │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE calls LambdaFunction: "scale-pod"         │            │
│  │  Parameters: {pod: "app-xyz", memory: "2Gi"}         │            │
│  │                                                      │            │
│  │  Remediation executed successfully                   │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=25s: VERIFY REMEDIATION                                       │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE queries metrics again:                    │            │
│  │  - CPU usage normalized                             │            │
│  │  - Memory errors decreased                          │            │
│  │                                                      │            │
│  │  Data Formulator creates verification visualization: │            │
│  │  📊 Chart 9: Before/After Comparison                │            │
│  │     - Shows improvement after remediation            │            │
│  │                                                      │            │
│  │  Agent-SRE updates Linear issue with verification    │            │
│  └──────────────────────────────────────────────────────┘            │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 🏗️ Architecture Integration

### Data Formulator Components

1. **Data Loaders**
   - Prometheus data loader for metrics
   - Loki data loader for logs
   - Tempo data loader for traces
   - Support for custom data sources

2. **Visualization Engine**
   - Chart generation (line, bar, scatter, pie, etc.)
   - Multi-metric dashboards
   - Time-series analysis
   - Correlation analysis

3. **AI Agent**
   - Goal-driven exploration
   - Automatic chart recommendations
   - Insight generation
   - Root cause analysis

### Integration Points

```
┌─────────────────────────────────────────────────────────────┐
│                    AGENT-SRE ARCHITECTURE                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Prometheus Alert                                           │
│       ↓                                                     │
│  Agent-SRE (receives CloudEvent)                           │
│       ↓                                                     │
│  ├─→ Create Linear Issue                                    │
│  ├─→ Query Observability Data                               │
│  │   ├─→ Prometheus (metrics)                               │
│  │   ├─→ Loki (logs)                                        │
│  │   └─→ Tempo (traces)                                     │
│  ├─→ Data Formulator (visualization)                        │
│  │   ├─→ Load data from observability stack                 │
│  │   ├─→ Generate visualizations                            │
│  │   ├─→ AI agent analysis                                  │
│  │   └─→ Export charts/images                               │
│  ├─→ Update Linear Issue (with visualizations)              │
│  ├─→ Select LambdaFunction (via annotations)                │
│  ├─→ Execute Remediation                                    │
│  └─→ Verify Remediation (with Data Formulator)              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔧 Implementation Details

### 1. Data Formulator Service Integration

```python
# src/sre_agent/data_formulator_client.py
from typing import Dict, List, Any, Optional
import httpx

class DataFormulatorClient:
    """Client for Data Formulator visualization service."""
    
    def __init__(self, base_url: str = "http://data-formulator:5000"):
        self.base_url = base_url
        self.client = httpx.AsyncClient()
    
    async def visualize_metrics(
        self,
        promql_queries: List[str],
        time_range: Dict[str, str],
        goal: Optional[str] = None
    ) -> List[Dict[str, Any]]:
        """
        Query Prometheus metrics and generate visualizations.
        
        Args:
            promql_queries: List of PromQL queries
            time_range: Time range dict with 'start' and 'end'
            goal: Optional analysis goal for AI agent
            
        Returns:
            List of chart definitions with data and metadata
        """
        # Load data from Prometheus
        data = await self._load_prometheus_data(promql_queries, time_range)
        
        # Use Data Formulator AI agent for goal-driven exploration
        if goal:
            charts = await self._ai_agent_explore(data, goal)
        else:
            charts = await self._recommend_charts(data)
        
        return charts
    
    async def visualize_logs(
        self,
        logql_query: str,
        time_range: Dict[str, str]
    ) -> List[Dict[str, Any]]:
        """Query Loki logs and generate visualizations."""
        # Load data from Loki
        data = await self._load_loki_data(logql_query, time_range)
        
        # Generate log visualizations
        charts = await self._analyze_logs(data)
        
        return charts
    
    async def visualize_traces(
        self,
        trace_query: str,
        time_range: Dict[str, str]
    ) -> List[Dict[str, Any]]:
        """Query Tempo traces and generate visualizations."""
        # Load data from Tempo
        data = await self._load_tempo_data(trace_query, time_range)
        
        # Generate trace visualizations
        charts = await self._analyze_traces(data)
        
        return charts
    
    async def export_charts(
        self,
        charts: List[Dict[str, Any]],
        format: str = "png"
    ) -> List[str]:
        """Export charts as images (PNG, SVG, etc.)."""
        # Export charts and return URLs or base64 encoded images
        pass
```

### 2. Agent-SRE Integration

```python
# src/sre_agent/main.py
from sre_agent.data_formulator_client import DataFormulatorClient

async def enrich_issue_with_visualizations(
    issue_id: str,
    alert_data: Dict[str, Any],
    linear_client: LinearClient
):
    """Enrich Linear issue with observability visualizations."""
    
    data_formulator = DataFormulatorClient()
    
    # Extract relevant labels for querying
    pod = alert_data.get("labels", {}).get("pod")
    namespace = alert_data.get("labels", {}).get("namespace")
    alertname = alert_data.get("alertname")
    
    # Determine time range (last 1 hour)
    time_range = {
        "start": "now-1h",
        "end": "now"
    }
    
    # Query metrics
    promql_queries = [
        f'rate(container_cpu_usage_seconds_total{{pod="{pod}",namespace="{namespace}"}}[5m])',
        f'container_memory_usage_bytes{{pod="{pod}",namespace="{namespace}"}}',
        f'rate(http_requests_total{{pod="{pod}"}}[5m])'
    ]
    
    goal = f"Analyze {alertname} alert for pod {pod} and identify root cause"
    metric_charts = await data_formulator.visualize_metrics(
        promql_queries,
        time_range,
        goal=goal
    )
    
    # Query logs
    logql_query = f'{{pod="{pod}",namespace="{namespace}"}} | json'
    log_charts = await data_formulator.visualize_logs(
        logql_query,
        time_range
    )
    
    # Query traces
    trace_query = f'{{service.name="{pod}"}} AND {{duration>1s}}'
    trace_charts = await data_formulator.visualize_traces(
        trace_query,
        time_range
    )
    
    # Export all charts
    all_charts = metric_charts + log_charts + trace_charts
    chart_images = await data_formulator.export_charts(all_charts, format="png")
    
    # Generate markdown comment with embedded images
    comment_body = generate_visualization_comment(
        metric_charts,
        log_charts,
        trace_charts,
        chart_images
    )
    
    # Update Linear issue
    await linear_client.create_comment(
        issue_id=issue_id,
        body=comment_body
    )
```

### 3. Kubernetes Deployment

```yaml
# k8s/data-formulator-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: data-formulator
  namespace: ai
spec:
  selector:
    app: data-formulator
  ports:
    - port: 5000
      targetPort: 5000

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: data-formulator
  namespace: ai
spec:
  replicas: 1
  selector:
    matchLabels:
      app: data-formulator
  template:
    metadata:
      labels:
        app: data-formulator
    spec:
      containers:
        - name: data-formulator
          image: ghcr.io/microsoft/data-formulator:latest
          ports:
            - containerPort: 5000
          env:
            - name: PROMETHEUS_URL
              value: "http://prometheus:9090"
            - name: LOKI_URL
              value: "http://loki:3100"
            - name: TEMPO_URL
              value: "http://tempo:3200"
          resources:
            requests:
              memory: "512Mi"
              cpu: "500m"
            limits:
              memory: "2Gi"
              cpu: "2000m"
```

---

## 📚 References

- [Data Formulator GitHub](https://github.com/microsoft/data-formulator)
- [Data Formulator Paper](https://arxiv.org/abs/2408.16119)
- [Agent-SRE Observability Documentation](../../docs/observability.md)

---

## ✅ Definition of Done

- [ ] Data Formulator service deployed in Kubernetes
- [ ] Prometheus data loader integrated
- [ ] Loki data loader integrated
- [ ] Tempo data loader integrated
- [ ] Agent-SRE can query Data Formulator API
- [ ] Visualizations generated automatically for alerts
- [ ] Visualizations embedded in Linear issue comments
- [ ] AI agent recommendations working
- [ ] Verification visualizations after remediation
- [ ] Documentation updated
- [ ] Integration tests passing

---

**Related Stories**:
- [SRE-007: Observability Enhancement](./BVL-51-SRE-007-observability-enhancement.md)
- [AI-002: LLaMA Factory Integration](./BVL-62-AI-002-llama-factory-finetuning.md)
- [AI-003: TinyRecursiveModels Integration](./BVL-63-AI-003-tiny-recursive-models.md)


## 🧪 Test Scenarios

### Scenario 1: Metrics Visualization Generation
1. Trigger alert for pod CPU high
2. Verify Data Formulator queries Prometheus metrics
3. Verify visualizations generated (CPU usage chart, correlation charts)
4. Verify visualizations include relevant context (alert firing time)
5. Verify charts exported as images (PNG)
6. Verify visualizations embedded in Linear issue
7. Verify visualizations help identify root cause

### Scenario 2: Log Visualization Generation
1. Trigger alert with error logs
2. Verify Data Formulator queries Loki logs
3. Verify log visualizations generated (log volume, error patterns)
4. Verify error patterns identified correctly
5. Verify visualizations show error trends over time
6. Verify visualizations embedded in Linear issue
7. Verify visualizations help diagnose issue

### Scenario 3: Trace Visualization Generation
1. Trigger alert with slow requests
2. Verify Data Formulator queries Tempo traces
3. Verify trace visualizations generated (latency distribution, service map)
4. Verify slow requests identified correctly
5. Verify service dependencies mapped accurately
6. Verify visualizations embedded in Linear issue
7. Verify visualizations help identify bottlenecks

### Scenario 4: AI Agent Analysis and Insights
1. Provide goal to Data Formulator ("identify root cause of CPU spike")
2. Verify AI agent queries relevant metrics/logs/traces
3. Verify AI agent generates relevant visualizations
4. Verify AI agent provides insights (root cause analysis)
5. Verify AI agent provides recommendations
6. Verify insights accuracy > 85%
7. Verify recommendations actionable

### Scenario 5: Multi-Source Visualization
1. Trigger complex alert requiring multiple data sources
2. Verify Data Formulator queries metrics, logs, and traces
3. Verify comprehensive visualizations generated from all sources
4. Verify visualizations show correlations between sources
5. Verify AI agent analyzes all data sources together
6. Verify unified insights generated
7. Verify visualizations embedded in Linear issue

### Scenario 6: Data Formulator Performance
1. Query large time range (7 days of data)
2. Verify query performance acceptable (< 30 seconds)
3. Verify visualization generation performance acceptable (< 10 seconds)
4. Verify no timeout issues
5. Verify memory usage acceptable (< 2GB)
6. Verify concurrent requests handled correctly
7. Verify metrics recorded for performance

### Scenario 7: Data Formulator Failure Handling
1. Simulate Data Formulator service unavailable
2. Trigger enriched issue update
3. Verify failure handled gracefully
4. Verify fallback behavior works (issue still updated without visualizations)
5. Verify error logged with context
6. Verify retry logic works when service recovers
7. Verify alerts fire for repeated failures

### Scenario 8: Visualization Export and Embedding
1. Generate visualizations using Data Formulator
2. Export charts as PNG images
3. Verify images generated correctly
4. Verify images embedded in Linear issue markdown
5. Verify images accessible and viewable
6. Verify image sizes optimized (< 1MB per image)
7. Verify multiple images handled correctly

## 📊 Success Metrics

- **Visualization Generation Success Rate**: > 95%
- **Query Performance**: < 30 seconds for large time ranges (P95)
- **Visualization Generation**: < 10 seconds (P95)
- **AI Insight Accuracy**: > 85% (actionable and relevant)
- **Image Export Performance**: < 5 seconds per image (P95)
- **Image Size**: < 1MB per image
- **Test Pass Rate**: 100%

---

**Last Updated**: January 08, 2026
**Owner**: SRE Team
**Status**: Validation Required