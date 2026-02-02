# Pi Cluster

> **Part of**: [Homelab Documentation](../README.md) → Clusters  
> **Last Updated**: November 7, 2025

---

## Overview

**Platform**: k3s (lightweight Kubernetes)  
**Hardware**: Raspberry Pi 4/5 (4-8GB RAM each)  
**Architecture**: ARM64  
**Purpose**: Edge computing, IoT integration, distributed sensors  
**Nodes**: 3-6 (lightweight, can scale)  
**Network**: 10.247.0.0/16

---

## Node Architecture

| Node | Role | Workloads |
|------|------|-----------|
| 1 | control-plane | k3s server (lightweight) |
| 2-3 | edge-worker | IoT hubs, sensors, gateways |
| 4-6 | optional | Additional edge nodes |

---

## Key Features

- **Lightweight k3s**: Optimized for resource-constrained devices
- **Low Power Consumption**: ~15W per Pi node
- **Edge AI Inference**: Run small language models locally
- **IoT Integration**: Connect sensors, cameras, smart devices
- **Distributed Locations**: Can be placed across physical locations

---

## Use Cases

### 1. IoT Sensor Data Collection for Teams

**Purpose**: Collect edge device data and make it accessible to teams via AI agents.

#### Data Flow Architecture

```
┌─────────────────────────────────────────────────────┐
│  Edge → Studio Agents → Teams                       │
├─────────────────────────────────────────────────────┤
│                                                     │
│  1. Pi Cluster (Edge)                               │
│     └─ Sensors collect data                         │
│        └─ Local SLM: Quick analysis (<50ms)         │
│           └─ Anomaly? → Send to Studio              │
│              Normal? → Store locally                │
│                                                     │
│  2. Studio Cluster (Agent Layer)                    │
│     └─ agent-bruno receives edge data               │
│        └─ Query Knowledge Graph for context         │
│           └─ Present insights to teams              │
│                                                     │
│  3. Teams Access                                    │
│     └─ "agent-bruno, show me sensor anomalies"     │
│     └─ "What's the temperature in the warehouse?"  │
│     └─ "Alert me if humidity drops below 30%"      │
│                                                     │
└─────────────────────────────────────────────────────┘
```

**→ [See AI Connectivity](../architecture/ai-connectivity.md#for-edge-devices-intelligent-edge) for full edge intelligence pattern**

#### Sensor Collection Deployment

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: sensor-collector
  namespace: iot
spec:
  selector:
    matchLabels:
      app: sensor-collector
  template:
    spec:
      containers:
      - name: collector
        image: sensor-collector:arm64
        env:
        - name: STUDIO_AGENT_URL
          value: "http://agent-bruno.ai-agents.svc.studio.remote:30120"
        - name: REPORT_INTERVAL
          value: "60"  # seconds
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
```

#### Team Access Examples

**Via AI Agent (Natural Language)**:
```bash
# Team member asks agent-bruno
$ curl http://agent-bruno.studio.cluster:30120/query \
  -d "Show me all sensor readings from warehouse Pi nodes in the last hour"

# Agent queries edge data and responds
Response: "Warehouse sensors (3 nodes):
- Temperature: 22.5°C (normal)
- Humidity: 45% (normal)  
- Anomaly detected at 14:23: Temperature spike to 28°C
- Root cause: HVAC malfunction (resolved)"
```

**Via Direct API**:
```python
# Python SDK for team access
from homelab import EdgeData

# Get sensor data via agent
edge = EdgeData(agent="agent-bruno")

# Query recent anomalies
anomalies = edge.get_anomalies(
    cluster="pi",
    namespace="iot",
    time_range="1h"
)

for anomaly in anomalies:
    print(f"Sensor: {anomaly.sensor_id}")
    print(f"Value: {anomaly.value}")
    print(f"Analysis: {anomaly.ai_analysis}")
```

### 2. Edge AI Inference

Run small language models locally:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollama-edge
  namespace: edge-ai
spec:
  replicas: 1
  template:
    spec:
      nodeSelector:
        role: edge-worker
      containers:
      - name: ollama
        image: ollama/ollama:latest
        args:
        - "run"
        - "llama2:7b"  # Small model for Pi
        resources:
          requests:
            cpu: "2000m"
            memory: "4Gi"
```

### 3. Home Automation

Smart home integration:

```bash
# Deploy home automation hub
kubectl --context=pi apply -f home-automation/

# Connect devices
kubectl --context=pi get pods -n home-automation
```

### 4. Environmental Monitoring

Monitor temperature, humidity, air quality:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: environmental-sensor
spec:
  containers:
  - name: sensor
    image: environmental-sensor:arm64
    env:
    - name: SENSOR_TYPE
      value: "DHT22"
    - name: REPORT_INTERVAL
      value: "60"  # seconds
    - name: SEND_TO_AGENT
      value: "agent-bruno.ai-agents.svc.studio.remote:30120"
```

---

## Edge Data Pipeline for Teams

### Architecture Overview

```
┌────────────────────────────────────────────────────────────┐
│           Complete Edge-to-Team Data Pipeline              │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  Edge Collection (Pi Cluster)                              │
│  ├─ Temperature sensors (DHT22)                            │
│  ├─ Humidity sensors                                       │
│  ├─ Motion detectors                                       │
│  ├─ Camera feeds                                           │
│  └─ Local SLM analysis: Ollama Llama 3 (8B)               │
│     └─ Fast decisions: <50ms                               │
│        └─ Normal: Store locally (95% of data)             │
│        └─ Anomaly: Send to Studio (5% of data)            │
│                                                            │
│  ↓ (mTLS via Linkerd)                                     │
│                                                            │
│  Agent Layer (Studio Cluster)                              │
│  ├─ agent-bruno receives edge events                       │
│  ├─ Query Knowledge Graph for context                      │
│  ├─ Complex analysis via Forge LLM (if needed)            │
│  └─ Present insights to teams                              │
│                                                            │
│  ↓                                                         │
│                                                            │
│  Team Access (Multiple Interfaces)                         │
│  ├─ Natural Language: "Show warehouse sensors"            │
│  ├─ API: edge.get_readings(cluster="pi")                  │
│  ├─ Dashboards: Grafana with edge metrics                 │
│  └─ Alerts: PagerDuty on anomalies                        │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### Team-Specific Use Cases

#### For Developers

```python
# Deploy edge application to Pi cluster
from homelab import EdgeDeploy

deploy = EdgeDeploy(cluster="pi")

# Deploy data collector
deploy.apply(
    name="warehouse-sensors",
    image="sensor-collector:v1.2.0",
    env={
        "LOCATION": "warehouse-b",
        "REPORT_TO": "agent-bruno"
    }
)

# Query via agent
data = deploy.query_agent("Show me warehouse-b sensor status")
print(data)
```

#### For SREs

```bash
# Monitor edge health via agent
curl http://agent-bruno.studio.cluster:30120/query \
  -d "List all Pi nodes and their health status"

# Response:
# Pi nodes: 6 healthy, 0 unhealthy
# - pi-node-1: CPU 35%, Memory 45%, Temp 42°C ✅
# - pi-node-2: CPU 28%, Memory 38%, Temp 39°C ✅
# - pi-node-3: CPU 82%, Memory 91%, Temp 68°C ⚠️ (High load)

# Get recommendations
curl http://agent-bruno.studio.cluster:30120/query \
  -d "Why is pi-node-3 under high load?"

# Response with AI analysis:
# "Analysis: Pi-node-3 is processing video stream from camera-5.
# Recommendation: Move video processing to Forge cluster GPU,
# keep only lightweight preprocessing on Pi."
```

#### For Data Scientists

```python
# Access edge data for ML training
from homelab import EdgeData
import pandas as pd

# Get historical sensor data
edge = EdgeData(agent="agent-bruno")

# Query last 30 days of data
df = edge.get_timeseries(
    cluster="pi",
    sensors=["temperature", "humidity", "pressure"],
    time_range="30d",
    include_anomalies=True
)

# Train model on Forge
from homelab import ForgeTraining

training = ForgeTraining(cluster="forge")
model = training.train(
    data=df,
    model_type="timeseries_forecast",
    target="temperature",
    gpus=2
)

# Deploy model back to edge
edge.deploy_model(model, nodes=["pi-node-1", "pi-node-2"])
```

### Data Retention Strategy

**Edge (Pi)**: 
- Raw data: 7 days local storage
- Anomalies: Immediate send to Studio
- Normal data: Aggregate & batch every 5 minutes

**Studio (Agent Layer)**:
- All events: 30 days (hot storage)
- Aggregated data: 1 year (warm storage)
- Knowledge Graph: Indefinite (insights & patterns)

**Forge (Analysis)**:
- Training datasets: 90 days
- Model artifacts: Versioned, indefinite

### Bandwidth Optimization

```python
# Example: Intelligent edge filtering
# Running on Pi nodes

async def process_sensor_reading(reading: dict):
    """
    Smart filtering at edge - only send what matters
    """
    # Quick local analysis with SLM (Ollama on Pi)
    analysis = await ollama.analyze(reading)
    
    if analysis["severity"] == "critical":
        # Send immediately to Studio
        await send_to_studio(reading, priority="high")
    
    elif analysis["severity"] == "warning":
        # Batch and send every 5 minutes
        await batch_for_studio(reading)
    
    else:
        # Just store locally
        await store_locally(reading)
    
    # Result: 95% of data stays local
    # Bandwidth savings: ~20x reduction
```

---

## Resource Limits

### Per Node (Pi 5, 8GB)

- **CPU**: 4 cores (ARM Cortex-A76)
- **Memory**: 8GB RAM
- **Disk**: 64GB+ microSD or SSD
- **Power**: ~15W per node

### Cluster Total (6 nodes)

- **CPU**: 24 cores
- **Memory**: 48GB
- **Disk**: 384GB+
- **Power**: ~90W total

---

## k3s Configuration

Pi cluster uses k3s instead of full Kubernetes:

```yaml
# k3s optimizations for Pi
k3s-config:
  disable:
    - traefik  # Use Linkerd instead
    - servicelb
  kubelet-arg:
    - "max-pods=30"  # Limit per node
    - "eviction-hard=memory.available<100Mi"
  kube-apiserver-arg:
    - "feature-gates=EphemeralContainers=true"
```

---

## Service Mesh Integration

Pi connects to other clusters via Linkerd:

```yaml
# Access Studio services from Pi
apiVersion: v1
kind: Service
metadata:
  name: studio-api
spec:
  type: ExternalName
  externalName: api.default.svc.studio.remote
```

---

## Best Practices for Pi

### DO

- ✅ Use ARM64 images
- ✅ Set resource limits
- ✅ Use SSDs instead of microSD (better performance)
- ✅ Monitor temperature
- ✅ Use DaemonSets for edge workloads

### DON'T

- ❌ Don't run resource-intensive workloads
- ❌ Don't use swap (kills SD card)
- ❌ Don't overclock (stability issues)
- ❌ Don't store critical data (use Studio/Pro)

---

## Hardware Recommendations

### Raspberry Pi 5 (Recommended)

- **CPU**: 4× ARM Cortex-A76 @ 2.4GHz
- **RAM**: 8GB
- **Storage**: NVMe SSD via PCIe
- **Network**: Gigabit Ethernet
- **Cost**: ~$80 + SSD

### Raspberry Pi 4 (Budget)

- **CPU**: 4× ARM Cortex-A72 @ 1.8GHz
- **RAM**: 4-8GB
- **Storage**: microSD or USB SSD
- **Network**: Gigabit Ethernet
- **Cost**: ~$55 + storage

---

## Common Operations

### Deploy Workload

```bash
# Deploy to Pi
kubectl --context=pi apply -f workload.yaml

# Verify
kubectl --context=pi get pods -A
```

### Check Resource Usage

```bash
# Node resources
kubectl --context=pi top nodes

# Pod resources
kubectl --context=pi top pods -A
```

### Monitor Temperature

```bash
# SSH to Pi node
ssh pi@pi-node-1

# Check temperature
vcgencmd measure_temp
```

---

## Troubleshooting

### High Temperature

```bash
# Check temperature
vcgencmd measure_temp

# If >80°C, improve cooling:
# - Add heatsink
# - Add fan
# - Improve airflow
```

### Out of Memory

```bash
# Check memory
free -h

# Restart high-memory pods
kubectl --context=pi rollout restart deployment/high-memory-pod

# Reduce replicas
kubectl --context=pi scale deployment/app --replicas=1
```

### Slow Performance

```bash
# Use SSD instead of microSD
# Check disk I/O
iostat -x 1

# Upgrade to Pi 5 if using Pi 4
```

---

## Related Documentation

- [Studio Cluster](studio-cluster.md)
- [Forge Cluster](forge-cluster.md)
- [Edge AI Architecture](../architecture/ai-connectivity.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

