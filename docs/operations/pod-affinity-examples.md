# Pod Affinity Examples

> **Part of**: [Homelab Documentation](../README.md) → Operations  
> **Last Updated**: November 7, 2025

---

## Overview

Workload placement patterns using node selectors, affinity, and anti-affinity rules for optimal resource utilization and reliability.

## Basic Node Selector

**Use Case**: Place pod on specific node type

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agent-bruno
  namespace: ai-agents
spec:
  replicas: 1
  selector:
    matchLabels:
      app: agent-bruno
  template:
    metadata:
      labels:
        app: agent-bruno
    spec:
      nodeSelector:
        role: ai-agents
        criticality: high
      containers:
      - name: agent-bruno
        image: brunolucena/agent-bruno:latest
        resources:
          requests:
            cpu: "500m"
            memory: "1Gi"
          limits:
            cpu: "2000m"
            memory: "4Gi"
```

## Pod Affinity (Co-location)

**Use Case**: Place pods close to each other for low-latency communication

### Example: Data Processor Near Database

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: data-processor
  namespace: data
spec:
  replicas: 3
  selector:
    matchLabels:
      app: data-processor
  template:
    metadata:
      labels:
        app: data-processor
    spec:
      affinity:
        podAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - postgresql
            topologyKey: kubernetes.io/hostname  # Same node
      containers:
      - name: processor
        image: data-processor:latest
```

### Example: Agent Near LanceDB (Future)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agent-auditor
  namespace: ai-agents
spec:
  template:
    spec:
      affinity:
        podAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 80
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - lancedb
              topologyKey: topology.kubernetes.io/zone  # Same zone (best effort)
      containers:
      - name: agent-auditor
        image: brunolucena/agent-auditor:latest
```

## Pod Anti-Affinity (Distribution)

**Use Case**: Spread pods across nodes/zones for high availability

### Example: Distribute API Servers

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-server
  template:
    metadata:
      labels:
        app: api-server
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - api-server
              topologyKey: topology.kubernetes.io/zone  # Different zones
          - weight: 50
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - api-server
              topologyKey: kubernetes.io/hostname  # Different nodes
      containers:
      - name: api-server
        image: api-server:latest
```

### Example: Distribute AI Agents (Studio)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ai-agents
  namespace: ai-agents
spec:
  replicas: 6  # agent-bruno, agent-auditor, agent-jamie, agent-devops, agent-mary-kay, aigoat
  selector:
    matchLabels:
      tier: ai-agents
  template:
    metadata:
      labels:
        tier: ai-agents
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: tier
                operator: In
                values:
                - ai-agents
            topologyKey: topology.kubernetes.io/zone  # Must be in different zones
      nodeSelector:
        role: ai-agents
      containers:
      - name: agent
        image: agent:latest
```

## GPU Workload Placement (Forge)

**Use Case**: Schedule GPU workloads on nodes with NVIDIA GPUs

### Example: Training Job

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: training-job
  namespace: ml-training
spec:
  replicas: 1
  selector:
    matchLabels:
      app: training-job
  template:
    metadata:
      labels:
        app: training-job
    spec:
      nodeSelector:
        role: training
        gpu-type: nvidia
      containers:
      - name: trainer
        image: pytorch:latest
        resources:
          limits:
            nvidia.com/gpu: "2"  # Request 2 GPUs
          requests:
            cpu: "4000m"
            memory: "16Gi"
        env:
        - name: CUDA_VISIBLE_DEVICES
          value: "0,1"
```

### Example: VLLM Inference

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vllm
  namespace: ml-inference
  annotations:
    multicluster.linkerd.io/export: "true"  # Cross-cluster access
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vllm
  template:
    metadata:
      labels:
        app: vllm
    spec:
      nodeSelector:
        role: inference
        gpu-type: nvidia
      containers:
      - name: vllm
        image: vllm/vllm-openai:latest
        args:
        - "--model=meta-llama/Meta-Llama-3.1-70B-Instruct"
        - "--tensor-parallel-size=2"
        resources:
          limits:
            nvidia.com/gpu: "2"  # 2× A100 GPUs
          requests:
            cpu: "8000m"
            memory: "32Gi"
        ports:
        - containerPort: 8000
```

## Edge Workload Placement (Pi)

**Use Case**: Schedule lightweight workloads on Pi cluster

### Example: IoT Sensor Gateway

```yaml
apiVersion: apps/v1
kind: DaemonSet  # Run on every Pi node
metadata:
  name: sensor-gateway
  namespace: iot
spec:
  selector:
    matchLabels:
      app: sensor-gateway
  template:
    metadata:
      labels:
        app: sensor-gateway
    spec:
      nodeSelector:
        role: edge-worker
      containers:
      - name: gateway
        image: sensor-gateway:arm64
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "200m"
            memory: "256Mi"
```

## Combined Affinity Rules

**Use Case**: Complex placement with multiple constraints

### Example: High Availability Service with Resource Constraints

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: critical-service
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: critical-service
  template:
    metadata:
      labels:
        app: critical-service
    spec:
      nodeSelector:
        criticality: critical  # Only critical nodes
        resource-profile: high  # High-resource nodes
      affinity:
        # Prefer nodes with platform services
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 50
            preference:
              matchExpressions:
              - key: role
                operator: In
                values:
                - platform
        
        # Co-locate with cache
        podAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 70
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - redis
              topologyKey: topology.kubernetes.io/zone
        
        # Distribute across zones
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - critical-service
            topologyKey: topology.kubernetes.io/zone
      
      containers:
      - name: service
        image: critical-service:latest
```

## Taints and Tolerations

**Use Case**: Reserve nodes for specific workloads

### Example: GPU Node Taint (Forge)

```bash
# Taint GPU nodes to reserve for GPU workloads only
kubectl --context=forge taint nodes forge-worker-3 gpu=true:NoSchedule
```

```yaml
# Only pods with this toleration can schedule on GPU nodes
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gpu-workload
spec:
  template:
    spec:
      tolerations:
      - key: "gpu"
        operator: "Equal"
        value: "true"
        effect: "NoSchedule"
      nodeSelector:
        gpu-type: nvidia
      containers:
      - name: workload
        resources:
          limits:
            nvidia.com/gpu: "1"
```

## Best Practices

1. **Use nodeSelector for simple cases** - Easy to understand and maintain
2. **Use affinity for complex placement** - When you need weights and preferences
3. **Use anti-affinity for HA** - Distribute across nodes/zones
4. **Combine rules carefully** - Test with `kubectl describe pod` to see why scheduling fails
5. **Document custom labels** - Keep [Node Labels Reference](node-labels-reference.md) updated
6. **Test in Air/Pro first** - Validate placement before Studio production

## Troubleshooting

### Pod Stuck in Pending

```bash
# Check why pod isn't scheduling
kubectl --context=studio describe pod <pod-name> | grep -A 10 "Events:"
```

Common causes:
- No nodes match nodeSelector
- Affinity rules can't be satisfied
- Resource requests exceed available capacity
- Taints without matching tolerations

### Verify Node Labels

```bash
# Check node labels
kubectl --context=studio get nodes --show-labels

# Check specific label
kubectl --context=studio get nodes -l role=ai-agents
```

## Related Documentation

- [Node Labels Reference](node-labels-reference.md)
- [Migration Guide](migration-guide.md)
- [Cluster Documentation](../clusters/README.md)

---

**Last Updated**: November 7, 2025  
**Maintained by**: SRE Team (Bruno Lucena)

