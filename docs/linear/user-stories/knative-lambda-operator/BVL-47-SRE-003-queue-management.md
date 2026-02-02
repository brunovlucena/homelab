# ⚡ SRE-003: Queue Management

**Status**: Done
**Linear URL**: https://linear.app/bvlucena/issue/BVL-220/sre-003-queue-management
**Priority**: P0
**Story Points**: 5  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-172/sre-003-queue-management  
**Created**: 2025-10-29  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---

## 📋 User Story

**As an** SRE Engineer  
**I want** to monitor and manage RabbitMQ queues effectively  
**So that** build events are processed reliably without backlogs or data loss

---


## 🎯 Acceptance Criteria

- [ ] [ ] Queue depth <100 messages (steady state)
- [ ] [ ] Message processing latency <5s (p95)
- [ ] [ ] Dead letter queue processed within 1hr
- [ ] [ ] Zero message loss during failures
- [ ] [ ] Auto-scaling triggers when queue >1000
- [ ] [ ] Alerts fire when queue >500 for >5min
- [ ] [ ] Queue metrics visible in Grafana
- [ ] --

---


## 📊 Acceptance Criteria

- [ ] Queue depth <100 messages (steady state)
- [ ] Message processing latency <5s (p95)
- [ ] Dead letter queue processed within 1hr
- [ ] Zero message loss during failures
- [ ] Auto-scaling triggers when queue >1000
- [ ] Alerts fire when queue >500 for >5min
- [ ] Queue metrics visible in Grafana

---

## 📬 Queue Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                    RABBITMQ QUEUES                             │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  1. lambda-build-events-${ENV}                                 │
│     ├─ Source: External API, scheduled jobs                    │
│     ├─ Consumer: Builder Service (2 replicas)                  │
│     ├─ Purpose: Trigger function builds                        │
│     ├─ Typical depth: 10-50 messages                           │
│     └─ DLQ: lambda-build-events-${ENV}-dlq                     │
│                                                                │
│  2. lambda-service-events-${ENV}                               │
│     ├─ Source: Builder Service (build completions)             │
│     ├─ Consumer: Builder Service (service creator)             │
│     ├─ Purpose: Create Knative services after build            │
│     ├─ Typical depth: 5-20 messages                            │
│     └─ DLQ: lambda-service-events-${ENV}-dlq                   │
│                                                                │
│  3. parser-results-${ENV}                                      │
│     ├─ Source: Functions (parser execution results)            │
│     ├─ Consumer: Results aggregator                            │
│     ├─ Purpose: Collect function outputs                       │
│     ├─ Typical depth: 50-200 messages                          │
│     └─ DLQ: parser-results-${ENV}-dlq                          │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## 🔍 Monitoring & Alerting

### Key Metrics

```bash
# Check queue status
make rabbitmq-status ENV=prd

# Output:
Queue Name | Messages | Consumers | Rate In | Rate Out
----------------------------- | ---------- | ----------- | --------- | ----------
lambda-build-events-prd | 45 | 2         | 10/s | 8/s
lambda-service-events-prd | 12 | 2         | 5/s | 5/s
parser-results-prd | 156 | 1         | 20/s | 15/s
```

### Prometheus Metrics

```promql
# Queue depth
rabbitmq_queue_messages{queue="lambda-build-events-prd"}

# Message rate
rate(rabbitmq_queue_messages_published_total[5m])
rate(rabbitmq_queue_messages_delivered_total[5m])

# Consumer lag
rabbitmq_queue_messages - 
on (queue) rabbitmq_queue_messages_unacknowledged
```

### Alerts

```yaml
# Alert: Queue Backlog
- alert: RabbitMQQueueBacklog
  expr: rabbitmq_queue_messages > 1000
  for: 5m
  severity: warning
  annotations:
    summary: "RabbitMQ queue {{ $labels.queue }} has {{ $value }} messages"

# Alert: No Consumers
- alert: RabbitMQNoConsumers
  expr: rabbitmq_queue_consumers == 0
  for: 2m
  severity: critical
  annotations:
    summary: "Queue {{ $labels.queue }} has no active consumers"
```

---

## 🔧 Common Issues & Resolutions

### Issue 1: Queue Backlog (>1000 messages)

**Symptoms**:
- Queue depth increasing steadily
- Build latency increasing
- Consumers can't keep up with producers

**Root Causes**:
- Insufficient builder replicas
- Slow builds (Kaniko resource starved)
- Consumers crashed/unhealthy

**Resolution**:
```bash
# 1. Check consumer status
kubectl get pods -n knative-lambda -l app=knative-lambda-builder

# 2. Check consumer logs
kubectl logs deployment/knative-lambda-builder -n knative-lambda --tail=100

# 3. Scale up temporarily
kubectl scale deployment/knative-lambda-builder --replicas=5 -n knative-lambda

# 4. Monitor queue drain
watch -n 5 'make rabbitmq-status ENV=prd'

# 5. Scale back when queue <100
kubectl scale deployment/knative-lambda-builder --replicas=2 -n knative-lambda
```

---

### Issue 2: Messages Stuck in DLQ

**Symptoms**:
- Dead letter queue has messages
- Failed builds not retrying
- Alert: DeadLetterQueueNotEmpty

**Root Causes**:
- Persistent failures (S3 access, invalid parser)
- Consumer crashes
- Message format incompatible

**Resolution**:
```bash
# 1. Inspect DLQ messages
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl list_queues name messages | grep dlq

# 2. Peek at DLQ message
# (requires RabbitMQ management UI or rabbitmqadmin)
open http://localhost:15672  # after port-forward

# 3. Replay DLQ messages (if fixable)
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmq-plugins enable rabbitmq_shovel
  
# Configure shovel to replay DLQ → main queue

# 4. Purge DLQ (if unrecoverable)
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl purge_queue lambda-build-events-prd-dlq
```

---

### Issue 3: No Consumers

**Symptoms**:
- Queue has messages but not processing
- Consumer count = 0
- Builder pods CrashLoopBackOff

**Root Causes**:
- Builder service down
- RabbitMQ connection failure
- ConfigMap misconfiguration

**Resolution**:
```bash
# 1. Check pod status
kubectl get pods -n knative-lambda -l app=knative-lambda-builder

# 2. Check logs
kubectl logs deployment/knative-lambda-builder -n knative-lambda --tail=50

# 3. Check RabbitMQ connectivity
kubectl exec -n knative-lambda <builder-pod> -- \
  curl -v telnet://rabbitmq-cluster-prd.rabbitmq-prd:5672

# 4. Restart deployment if needed
kubectl rollout restart deployment/knative-lambda-builder -n knative-lambda
```

---

## ⚙️ Auto-Scaling Configuration

### HPA Based on Queue Depth

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: knative-lambda-builder
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: knative-lambda-builder
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: External
    external:
      metric:
        name: rabbitmq_queue_messages
        selector:
          matchLabels:
            queue: lambda-build-events-prd
      target:
        type: Value
        value: "100"  # Scale up if queue >100
```

---

## 💡 Pro Tips

### Queue Health
- **Target queue depth: <100**: Steady state with healthy consumers
- **Warning level: >500**: Scale up consumers
- **Critical level: >1000**: Immediate action required
- **DLQ threshold: >10**: Investigate failures

### Message Durability
- All queues are **durable** (survive RabbitMQ restarts)
- Messages are **persistent** (written to disk)
- **Prefetch count: 10** (balance throughput vs. rebalancing)
- **Ack mode: manual** (only ack after successful processing)

### Performance
- Use multiple consumers (replicas) for parallel processing
- Monitor consumer lag (messages - unacked)
- Purge test queues regularly (`make rabbitmq-purge-lambda-queues-dev`)
- Enable lazy queues for large backlogs (disk-backed)

---

