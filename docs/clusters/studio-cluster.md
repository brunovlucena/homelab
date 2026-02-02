# Studio Cluster ⭐

> **Part of**: [Homelab Documentation](../README.md) → Clusters  
> **Last Updated**: November 7, 2025

---

## Overview

**Platform**: Kind  
**Hardware**: Mac Studio (M2 Ultra, 192GB RAM)  
**Architecture**: ARM64  
**Purpose**: Production-like environment for AI agents ⭐  
**Nodes**: 6 (1 control-plane + 5 workers)  
**Network**: 10.245.0.0/16

---

## Node Architecture (Optimized 6-Node Design)

| Node | Role | Zone | Workloads |
|------|------|------|-----------|
| 1 | control-plane | - | Kubernetes API, etcd, scheduler |
| 2 | platform | us-east-1a | Flux, cert-manager, Linkerd, Flagger, operators |
| 3 | ai-agents | us-east-1b | All AI agents (Knative scale-to-zero) |
| 4 | observability | us-east-1a | Prometheus, Grafana, Loki, Tempo, Alloy |
| 5 | data | us-east-1b | PostgreSQL, Redis, MinIO, RabbitMQ |
| 6 | serverless | us-east-1c | Knative, Cloudflare tunnel, pihole, GitHub runners |

### Why 6 Nodes vs 12 Nodes?

| Metric | 12 Nodes | 6 Nodes | Savings |
|--------|----------|---------|---------|
| DaemonSet overhead | ~5.1 GB | ~2.5 GB | **2.6 GB** |
| Control plane overhead | Higher | Lower | ~1 GB |
| Complexity | High | Medium | Simpler ops |
| Pod capacity | 1320 | 660 | Still sufficient |
| Estimated pods used | ~100 | ~100 | Same workload |

**Pod Limits**: 110 pods/node × 6 nodes = 660 total capacity (vs ~100 estimated usage)

---

## Key Features

- **High Availability**: Multi-zone deployment with zone awareness
- **Production-grade Security**: External Secrets Operator + mTLS via Linkerd
- **Comprehensive Observability**: Full metrics, logs, traces
- **AI Agent Hosting**: Production AI agents (agent-bruno, agent-auditor, etc.)
- **Multi-instance Reliability**: Critical services distributed across zones

---

## AI Agents Deployed (Knative Services)

Studio hosts all production AI agents **as Knative services** for optimal resource utilization and event-driven integration:

| Agent | Type | Purpose | Scale-to-Zero |
|-------|------|---------|---------------|
| agent-bruno | Knative Service | Main AI assistant | ✅ Yes |
| agent-auditor | Knative Service | Event-driven infrastructure response | ✅ Yes |
| agent-jamie | Knative Service | DevOps automation | ✅ Yes |
| agent-mary-kay | Knative Service | Sales assistant | ✅ Yes |
| aigoat | Knative Service | Goal tracking | ✅ Yes |

**Benefits of Knative Deployment**:

- **Scale-to-zero**: Inactive agents consume zero resources (cost savings)
- **Auto-scaling**: 0→N in <30s based on request load
- **CloudEvents Native**: Direct integration with eventing platform (RabbitMQ)
- **Event-Driven**: Respond to Prometheus alerts, GitHub webhooks, Slack events
- **Resource Optimization**: Only consume resources when actively processing

---

## Use Cases

### 1. Production AI Agent Hosting

**Purpose**: Host AI agents that connect teams, clusters, and edge devices.

#### Agent Layer Architecture

```
┌─────────────────────────────────────────────────────┐
│  Studio: AI Agent Layer for Teams & Edge            │
├─────────────────────────────────────────────────────┤
│                                                     │
│  Teams ←→ AI Agents ←→ Edge Devices                 │
│                                                     │
│  Example Flow:                                      │
│  1. Team: "Show me warehouse sensor status"        │
│  2. agent-bruno queries Pi cluster via Linkerd     │
│  3. Agent analyzes data with Knowledge Graph       │
│  4. Agent presents insights to team                 │
│                                                     │
│  Edge Data Collection:                              │
│  1. Pi sensors → agent-bruno (anomalies)           │
│  2. Agent stores in Knowledge Graph                │
│  3. Agent alerts teams on critical issues          │
│  4. Agent coordinates with Forge for deep analysis │
│                                                     │
└─────────────────────────────────────────────────────┘
```

**→ [See Pi Cluster](pi-cluster.md#edge-data-pipeline-for-teams) for edge data collection**  
**→ [See AI Connectivity](../architecture/ai-connectivity.md) for full pattern**

#### Deploy AI Agent (Knative Service)

```bash
# Deploy via Flux (GitOps)
flux reconcile helmrelease agent-bruno -n ai-agents

# Verify
kubectl --context=studio get ksvc -n ai-agents
kubectl --context=studio get pods -n ai-agents  # Shows 0 when idle
```

**→ [See Complete Deployment Guide](../architecture/ai-agent-architecture.md#0-deployment-layer-knative-services)** - Full YAML, CloudEvents, triggers, operations

### 2. Multi-zone Failover Testing

Test HA capabilities:

```bash
# Cordon zone
kubectl --context=studio cordon -l topology.kubernetes.io/zone=us-east-1a

# Verify pods migrate
kubectl --context=studio get pods -n ai-agents -o wide

# Uncordon
kubectl --context=studio uncordon -l topology.kubernetes.io/zone=us-east-1a
```

### 3. Production Workload Monitoring

Comprehensive observability:

```bash
# Access Grafana
open http://studio.cluster:30040

# Check metrics
curl http://prometheus.studio.cluster:30041/api/v1/query?query=up

# Check logs
curl -G -d 'query={cluster="studio"}' http://loki.studio.cluster:30042/loki/api/v1/query
```

### 4. Cross-cluster Service Consumption

Agents consume services from Forge and Pi clusters:

```python
# agent-bruno calling VLLM on Forge for deep analysis
import openai

client = openai.OpenAI(
    api_key="EMPTY",
    base_url="http://vllm.ml-inference.svc.forge.remote:8000/v1"
)

response = client.chat.completions.create(
    model="meta-llama/Meta-Llama-3.1-70B-Instruct",
    messages=[{"role": "user", "content": "Analyze this sensor anomaly pattern..."}]
)
```

### 5. Edge Data Access for Teams (Event-Driven)

**Purpose**: Teams query edge device data through AI agents (Knative services) on Studio.

#### Team Interaction Patterns

```python
# Example 1: Developer querying edge sensors (Knative Service)
from homelab import Agent

# Agent automatically scales from 0→1 on first call
agent = Agent("agent-bruno", cluster="studio")

# Natural language query triggers Knative Service
response = agent.query(
    "Show me temperature readings from Pi cluster warehouse sensors in the last 24 hours"
)

print(response)
# Output:
# "Warehouse sensors (3 nodes):
# - Average temp: 22.3°C
# - Min: 19.8°C at 03:15 AM
# - Max: 24.7°C at 2:30 PM
# - Anomalies: 2 detected (both resolved)
# - Status: All sensors healthy ✅"

# Agent scales back to 0 after 60s of inactivity
# Resource usage: 0 CPU, 0 Memory when idle
```

```bash
# Example 2: SRE checking edge health (via Knative Service)
# First request wakes up agent (cold start ~5s)
curl http://agent-bruno.ai-agents.svc.cluster.local/query \
  -H "Content-Type: application/json" \
  -d '{"query": "Are there any unhealthy Pi nodes?"}'

# Agent scales from 0→1 automatically
# Processes request and returns response

# Agent response:
# {
#   "query": "Are there any unhealthy Pi nodes?",
#   "response": "All 6 Pi nodes are healthy. Node utilization:
#   - pi-node-1: 35% CPU, 42% Memory ✅
#   - pi-node-2: 28% CPU, 38% Memory ✅
#   - pi-node-3: 82% CPU, 91% Memory ⚠️
#   
#   Recommendation: pi-node-3 has high load due to video processing.
#   Consider moving workload to Forge GPU cluster.",
#   "cold_start": true,
#   "latency_ms": 5234
# }

# Subsequent requests are faster (warm start ~200ms)
# Agent scales to 0 after 60s idle period
```

```python
# Example 3: Data scientist accessing edge data for ML
from homelab import EdgeData
import pandas as pd

# Get edge data through agent
edge = EdgeData(agent="agent-bruno", cluster="studio")

# Query historical data
df = edge.get_timeseries(
    source="pi",
    sensors=["temperature", "humidity", "pressure"],
    time_range="30d",
    include_anomalies=True
)

# Data includes:
# - Raw sensor readings
# - Anomaly labels (from edge SLM)
# - Context from Knowledge Graph
# - AI analysis notes

print(f"Collected {len(df)} data points from edge")
# Output: Collected 432,000 data points from edge
```

#### Agent Capabilities for Edge Data

**1. Data Aggregation**:

- Collect from multiple Pi nodes
- Normalize different sensor types
- Filter by location, time, type

**2. Anomaly Analysis**:

- Correlate with Knowledge Graph
- Identify patterns and trends
- Predict future issues

**3. Recommendations**:

- Suggest workload optimization
- Alert on critical conditions
- Guide troubleshooting

**4. Visualization**:

- Generate Grafana dashboards
- Create reports for teams
- Export data in various formats

**→ [See Pi Cluster](pi-cluster.md#edge-data-pipeline-for-teams) for complete edge pipeline**

---

## Resource Limits

### Per Node

- **CPU**: 16 cores per specialized node
- **Memory**: 32GB per node
- **Disk**: 200GB per node

### Cluster Total

- **CPU**: 176 cores
- **Memory**: 352GB
- **Disk**: 2.2TB

---

## Production Readiness

Studio requires the highest operational standards:

- ✅ **Backups**: Daily automated backups via Velero
- ✅ **Monitoring**: Comprehensive metrics, logs, traces
- ✅ **Alerting**: AlertManager + PagerDuty
- ✅ **Security**: External Secrets Operator, mTLS, network policies
- ✅ **HA**: Multi-zone deployment
- ⚠️ **CI/CD**: Manual approval required for deployments

---

## Deployment Approval Process

Studio deployments require manual approval:

```yaml
# .github/workflows/deploy-studio.yml
jobs:
  deploy:
    runs-on: [self-hosted]
    environment: production  # Requires approval
    steps:
      - name: Deploy to Studio
        run: |
          kubectl --context=studio apply -f manifests/
```

---

## Related Documentation

- [Air Cluster](air-cluster.md)
- [Pro Cluster](pro-cluster.md)
- [Forge Cluster](forge-cluster.md) - GPU services
- [AI Agent Architecture](../architecture/ai-agent-architecture.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

