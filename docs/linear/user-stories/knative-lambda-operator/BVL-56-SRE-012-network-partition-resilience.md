# 🌐 SRE-012: Network Partition Resilience and DLQ Recovery

**Status**: Backlog  
**Priority**: P0  
**Story Points**: 13  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-177/sre-012-network-partition-resilience-and-dlq-recovery  
**Created**: 2026-01-19  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---


## 📋 User Story

**As a** SRE Engineer  
**I want to** network partition resilience and dlq recovery  
**So that** I can improve system reliability, security, and performance

---



## 🎯 Acceptance Criteria

- [ ] [ ] Events buffered during network outages
- [ ] [ ] No event loss during network partitions
- [ ] [ ] Automatic reconnection after network recovery
- [ ] [ ] Circuit breaker prevents cascading failures
- [ ] [ ] Alert fires: "NetworkPartitionDetected"
- [ ] [ ] Recovery time < 5 minutes
- [ ] [ ] Connection pool health monitored
- [ ] --

---


## Overview

This runbook addresses network partition scenarios where communication between Knative components, RabbitMQ broker, and consumer services is disrupted. Covers detection, mitigation, and recovery strategies to prevent event loss during network failures.

---

## 🎯 User Story: Survive Network Partitions Without Event Loss

### Story

**As an** SRE Engineer  
**I want** the event system to survive network partitions gracefully  
**So that** events are preserved and processed after network recovery without data loss

### Acceptance Criteria

- [ ] Events buffered during network outages
- [ ] No event loss during network partitions
- [ ] Automatic reconnection after network recovery
- [ ] Circuit breaker prevents cascading failures
- [ ] Alert fires: "NetworkPartitionDetected"
- [ ] Recovery time < 5 minutes
- [ ] Connection pool health monitored

---

## 💥 Network Partition Scenarios

### Scenario 1: Consumer → RabbitMQ Partition

```yaml
Failure Timeline:

  T+0s: Normal operation
    ├─ RabbitMQ: Healthy, accepting connections
    ├─ Consumers: 3 pods connected, consuming events
    └─ Event rate: 10 events/sec

  T+30s: Network partition occurs
    ├─ Kubernetes CNI issue: Pod network disruption
    ├─ Consumers lose connection to RabbitMQ
    ├─ RabbitMQ still healthy (control plane OK)
    └─ Consumer pods show: "Connection refused" errors

  T+30s to T+5m: Partition active
    ├─ New events: Published to RabbitMQ (via broker)
    ├─ Events queued: 3000 events (5min × 10/sec)
    ├─ Consumers: Attempting reconnection (exponential backoff)
    └─ Consumer logs: Repeated connection failures

  T+5m: Network recovery
    ├─ CNI restored
    ├─ Consumers reconnect to RabbitMQ
    ├─ Event backlog: 3000 events waiting
    └─ Processing resumes

  T+5m to T+10m: Backlog processing
    ├─ Consumers drain queue
    ├─ Auto-scaling triggered (queue depth > 1000)
    ├─ Scaled to 6 consumers
    └─ Normal operation restored at T+10m
```

### Detection

```bash
# Check RabbitMQ connectivity from consumer pods
kubectl exec -n knative-lambda deployment/knative-lambda-builder -- \
  nc -zv rabbitmq-cluster-prd.rabbitmq-prd.svc.cluster.local 5672

# Check consumer connection status
kubectl logs -n knative-lambda -l app=knative-lambda-builder --tail=100 | \
  grep -i "connection\ | disconnect\ | network\ | refused"

# Check RabbitMQ client connections
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl list_connections name peer_host peer_port state | \
  grep knative-lambda

# Check for network partition alerts in RabbitMQ
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl cluster_status | grep -i partition

# Check Kubernetes network policies
kubectl get networkpolicies -n knative-lambda
kubectl get networkpolicies -n rabbitmq-prd

# Test connectivity from test pod
kubectl run network-test -n knative-lambda --rm -it --restart=Never \
  --image=nicolaka/netshoot:latest -- \
  bash -c "
    echo 'Testing RabbitMQ connectivity...'
    nc -zv rabbitmq-cluster-prd.rabbitmq-prd.svc.cluster.local 5672
    echo 'Testing Knative Broker...'
    nc -zv knative-lambda-broker-prd-broker.knative-lambda.svc.cluster.local 80
  "
```

### Remediation

```bash
# Step 1: Verify network partition
kubectl get pods -n knative-lambda -o wide
kubectl get pods -n rabbitmq-prd -o wide

# Step 2: Check CNI plugin health
kubectl get pods -n kube-system -l k8s-app=calico-node
# or for other CNI:
kubectl get pods -n kube-system -l app=cilium

# Step 3: Restart consumer pods to force reconnection (if needed)
kubectl rollout restart deployment/knative-lambda-builder -n knative-lambda

# Step 4: Monitor reconnection
kubectl logs -n knative-lambda -l app=knative-lambda-builder --tail=50 -f | \
  grep -i "connection established\ | connected"

# Step 5: Verify event processing resumed
kubectl logs -n knative-lambda -l app=knative-lambda-builder --tail=20 | \
  grep "Processing CloudEvent"

# Step 6: Check queue backlog
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl list_queues name messages consumers

# Step 7: Scale consumers if backlog is large
BACKLOG=$(kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl list_queues name messages | \
  grep lambda-build-events-prd | awk '{print $2}')

if [ "$BACKLOG" -gt 1000 ]; then
  echo "Large backlog detected. Scaling consumers..."
  kubectl scale deployment/knative-lambda-builder -n knative-lambda --replicas=6
fi
```

### Prevention - Resilient Connection Handling

```go
// Add connection retry logic with exponential backoff
type RabbitMQConnection struct {
    conn          *amqp.Connection
    channel       *amqp.Channel
    connectionURL string
    reconnectMax  int
    reconnectTime time.Duration
    obs           observability.Observability
    mu            sync.RWMutex
    connected     bool
}

func NewRabbitMQConnection(url string, obs observability.Observability) *RabbitMQConnection {
    return &RabbitMQConnection{
        connectionURL: url,
        reconnectMax:  10,
        reconnectTime: 5 * time.Second,
        obs:           obs,
        connected:     false,
    }
}

func (r *RabbitMQConnection) Connect(ctx context.Context) error {
    attempt := 0
    backoff := time.Second

    for attempt < r.reconnectMax {
        attempt++
        
        r.obs.Info(ctx, "Attempting RabbitMQ connection",
            "attempt", attempt,
            "max_attempts", r.reconnectMax,
            "backoff_seconds", backoff.Seconds())

        conn, err := amqp.Dial(r.connectionURL)
        if err != nil {
            r.obs.Warn(ctx, "RabbitMQ connection failed, retrying",
                "attempt", attempt,
                "error", err.Error(),
                "next_retry_in", backoff.String())
            
            // Exponential backoff with jitter
            jitter := time.Duration(rand.Int63n(int64(backoff / 4)))
            time.Sleep(backoff + jitter)
            backoff *= 2
            if backoff > r.reconnectTime {
                backoff = r.reconnectTime
            }
            continue
        }

        // Connection successful
        ch, err := conn.Channel()
        if err != nil {
            conn.Close()
            r.obs.Error(ctx, err, "Failed to open channel")
            time.Sleep(backoff)
            continue
        }

        r.mu.Lock()
        r.conn = conn
        r.channel = ch
        r.connected = true
        r.mu.Unlock()

        r.obs.Info(ctx, "RabbitMQ connection established",
            "attempt", attempt,
            "server", r.connectionURL)

        // Start monitoring connection health
        go r.monitorConnection(ctx)

        return nil
    }

    return fmt.Errorf("failed to connect after %d attempts", r.reconnectMax)
}

func (r *RabbitMQConnection) monitorConnection(ctx context.Context) {
    // Listen for connection close notifications
    notifyClose := make(chan *amqp.Error)
    r.conn.NotifyClose(notifyClose)

    select {
    case err := <-notifyClose:
        r.mu.Lock()
        r.connected = false
        r.mu.Unlock()

        if err != nil {
            r.obs.Error(ctx, err, "RabbitMQ connection lost, attempting reconnect")
        }

        // Attempt reconnection
        time.Sleep(time.Second)
        if err := r.Connect(ctx); err != nil {
            r.obs.Error(ctx, err, "Reconnection failed")
        }
    case <-ctx.Done():
        return
    }
}

func (r *RabbitMQConnection) IsConnected() bool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.connected
}

func (r *RabbitMQConnection) Close() error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if r.channel != nil {
        r.channel.Close()
    }
    if r.conn != nil {
        return r.conn.Close()
    }
    return nil
}
```

---

## 🔀 Scenario 2: Knative Broker → Event Source Partition

```yaml
Failure Timeline:

  T+0s: External API publishes events to Knative Broker
    ├─ External service: api.example.com
    ├─ Knative Broker: Accepting events via HTTP POST
    └─ Event rate: 5 events/sec

  T+60s: Network partition between external API and cluster
    ├─ Ingress controller: Unable to reach broker
    ├─ External API: Receives 503 Service Unavailable
    ├─ Events: Buffered at external API (if implemented)
    └─ Knative Broker: Healthy but unreachable

  T+60s to T+10m: Partition active
    ├─ External API: Retry logic activated
    ├─ Events: Queued at source (300 events)
    ├─ Some events: Lost if no retry at source
    └─ Ingress logs: Connection timeouts

  T+10m: Network recovery
    ├─ Ingress accessible again
    ├─ External API: Replays buffered events
    ├─ Broker: Receives 300 event burst
    └─ Potential: Rate limiting triggered

  Risk: Events lost if external source doesn't buffer
```

### Detection

```bash
# Check Knative Broker ingress health
kubectl get ksvc -n knative-lambda
kubectl get ingress -n knative-lambda

# Check ingress controller logs
kubectl logs -n ingress-nginx -l app.kubernetes.io/component=controller | \
  grep -i "timeout\ | refused\ | unavailable"

# Test broker endpoint from outside cluster
curl -v -X POST https://broker.knative-lambda.example.com \
  -H "Content-Type: application/cloudevents+json" \
  -d '{
    "specversion": "1.0",
    "type": "com.example.test",
    "source": "test",
    "id": "test-123",
    "data": {"test": true}
  }'

# Check for dropped connections at ingress
kubectl exec -n ingress-nginx deployment/ingress-nginx-controller -- \
  cat /var/log/nginx/access.log | \
  grep -E "503 | 504" | tail -50
```

### Prevention - Client-Side Retry Logic

```javascript
// Add retry logic to event publisher (external service)
class KnativeEventPublisher {
  constructor(brokerUrl, options = {}) {
    this.brokerUrl = brokerUrl;
    this.maxRetries = options.maxRetries | | 5;
    this.retryDelay = options.retryDelay | | 1000;
    this.timeout = options.timeout | | 10000;
    this.eventBuffer = [];
    this.maxBufferSize = options.maxBufferSize | | 1000;
  }

  async publishEvent(event) {
    let attempt = 0;
    let backoff = this.retryDelay;

    while (attempt < this.maxRetries) {
      attempt++;

      try {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.timeout);

        const response = await fetch(this.brokerUrl, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/cloudevents+json',
            'Ce-Specversion': '1.0',
            'Ce-Type': event.type,
            'Ce-Source': event.source,
            'Ce-Id': event.id,
          },
          body: JSON.stringify(event.data),
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        if (response.ok) {
          console.log(`Event ${event.id} published successfully`);
          return true;
        }

        // Broker returned error - check if retryable
        if (response.status >= 500 | | response.status === 429) {
          throw new Error(`Server error: ${response.status}`);
        }

        // 4xx error - don't retry
        console.error(`Event ${event.id} rejected: ${response.status}`);
        return false;

      } catch (error) {
        console.warn(`Publish attempt ${attempt} failed:`, error.message);

        if (attempt >= this.maxRetries) {
          // Max retries reached - buffer event
          this.bufferEvent(event);
          return false;
        }

        // Exponential backoff with jitter
        const jitter = Math.random() * 1000;
        await new Promise(resolve => setTimeout(resolve, backoff + jitter));
        backoff *= 2;
      }
    }

    return false;
  }

  bufferEvent(event) {
    if (this.eventBuffer.length >= this.maxBufferSize) {
      console.error('Event buffer full. Dropping oldest event.');
      this.eventBuffer.shift();
    }

    this.eventBuffer.push({
      event: event,
      timestamp: Date.now(),
      attempts: 0,
    });

    console.log(`Event ${event.id} buffered. Buffer size: ${this.eventBuffer.length}`);
  }

  async replayBufferedEvents() {
    console.log(`Replaying ${this.eventBuffer.length} buffered events...`);

    let replayed = 0;
    let failed = 0;

    // Process in batches to avoid overwhelming broker
    const batchSize = 10;
    for (let i = 0; i < this.eventBuffer.length; i += batchSize) {
      const batch = this.eventBuffer.slice(i, i + batchSize);

      await Promise.all(batch.map(async (bufferedEvent) => {
        const success = await this.publishEvent(bufferedEvent.event);
        if (success) {
          replayed++;
        } else {
          failed++;
        }
      }));

      // Rate limiting between batches
      await new Promise(resolve => setTimeout(resolve, 500));
    }

    // Clear successfully replayed events
    this.eventBuffer = this.eventBuffer.filter(
      (_, index) => index >= replayed
    );

    console.log(`Replay complete. Replayed: ${replayed}, Failed: ${failed}`);
  }
}

// Usage
const publisher = new KnativeEventPublisher('https://broker.example.com', {
  maxRetries: 5,
  retryDelay: 1000,
  maxBufferSize: 1000,
});

// Publish with automatic retry and buffering
await publisher.publishEvent({
  type: 'com.example.build.start',
  source: 'external-api',
  id: 'build-789',
  data: { buildId: '789', thirdPartyId: 'tp-123' },
});

// Periodically replay buffered events (e.g., every minute)
setInterval(() => {
  if (publisher.eventBuffer.length > 0) {
    publisher.replayBufferedEvents();
  }
}, 60000);
```

---

## 🔀 Scenario 3: RabbitMQ Cluster Split-Brain

```yaml
Failure Timeline:

  T+0s: RabbitMQ cluster with 3 nodes
    ├─ rabbitmq-0: Leader
    ├─ rabbitmq-1: Follower
    └─ rabbitmq-2: Follower

  T+2m: Network partition between nodes
    ├─ Partition 1: rabbitmq-0, rabbitmq-1 (2 nodes)
    ├─ Partition 2: rabbitmq-2 (1 node)
    └─ Both partitions: Consider themselves active

  T+2m to T+15m: Split-brain active
    ├─ Partition 1: Accepts events (has quorum: 2/3)
    ├─ Partition 2: Rejects events (no quorum: 1/3)
    ├─ Consumers: Some connected to P1, some to P2
    └─ Risk: Data inconsistency

  T+15m: Network recovery
    ├─ RabbitMQ: Detects partition
    ├─ Resolution: Restart minority partition (rabbitmq-2)
    ├─ Data sync: P2 data lost (minority partition)
    └─ Recovery: Normal operation restored

  Impact: Events sent to minority partition are lost!
```

### Detection

```bash
# Check for RabbitMQ cluster partitions
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl cluster_status | grep -A10 "Partitions"

# Check quorum queue status
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl list_queues name type leader members online

# Check node connectivity
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl eval "rabbit_mnesia:status()."

# Monitor RabbitMQ cluster health
kubectl get rabbitmqclusters -n rabbitmq-prd -o yaml | \
  grep -A5 "status:"
```

### Remediation

```bash
# Step 1: Identify partition
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl cluster_status

# Step 2: Stop minority partition node (if split-brain detected)
MINORITY_NODE="rabbitmq-cluster-prd-2"
kubectl exec -n rabbitmq-prd $MINORITY_NODE -- \
  rabbitmqctl stop_app

# Step 3: Reset minority node
kubectl exec -n rabbitmq-prd $MINORITY_NODE -- \
  rabbitmqctl reset

# Step 4: Rejoin cluster
kubectl exec -n rabbitmq-prd $MINORITY_NODE -- \
  rabbitmqctl join_cluster rabbit@rabbitmq-cluster-prd-0

# Step 5: Start application
kubectl exec -n rabbitmq-prd $MINORITY_NODE -- \
  rabbitmqctl start_app

# Step 6: Verify cluster health
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl cluster_status

# Step 7: Check for lost messages (compare queue depths before/after)
kubectl exec -n rabbitmq-prd rabbitmq-cluster-prd-0 -- \
  rabbitmqctl list_queues name messages
```

### Prevention - Quorum Queues

```yaml
# Quorum queues prevent split-brain data loss
apiVersion: eventing.knative.dev/v1alpha1
kind: RabbitmqBrokerConfig
metadata:
  name: lambda-broker-config-prd
spec:
  queueType: quorum  # Requires quorum (majority) for writes
  arguments:
    x-queue-type: quorum
    x-quorum-initial-group-size: 3  # Replicate across 3 nodes
    x-max-in-memory-length: 0  # Store all messages on disk
    x-max-in-memory-bytes: 0

---
# RabbitMQ Cluster configuration
apiVersion: rabbitmq.com/v1beta1
kind: RabbitmqCluster
metadata:
  name: rabbitmq-cluster-prd
  namespace: rabbitmq-prd
spec:
  replicas: 3
  rabbitmq:
    additionalConfig: | # Pause minority during network partition
      cluster_partition_handling = pause_minority
      # Quorum queue settings
      quorum_commands = all
      quorum_ticket_ttl = 5000
```

---

## 📊 Monitoring & Alerts

```prometheus
# Alert: Network Partition Detected
- alert: NetworkPartitionDetected
  expr: | rabbitmq_partitions > 0
  for: 1m
  severity: critical
  annotations:
    summary: "RabbitMQ cluster partition detected"
    description: "Network partition between RabbitMQ nodes. Potential data loss risk."

# Alert: Consumer Disconnections
- alert: ConsumerDisconnectionSpike
  expr: | rate(rabbitmq_connections_closed_total[5m]) > 1
  for: 2m
  severity: warning
  annotations:
    summary: "High rate of consumer disconnections"
    description: "{{ $value }} disconnections/sec. Possible network issues."

# Alert: Broker Unreachable
- alert: KnativeBrokerUnreachable
  expr: | probe_success{job="knative-broker"} == 0
  for: 2m
  severity: critical
  annotations:
    summary: "Knative Broker unreachable from external probes"
    description: "Events cannot be published. Check ingress and network."

# Metric: Connection Uptime
- metric: rabbitmq_connection_uptime_seconds
  expr: | time() - rabbitmq_connection_created_timestamp_seconds
```

---

## 🔧 Network Resilience Best Practices

1. **Use Quorum Queues** - Prevent split-brain data loss
2. **Configure Connection Retries** - Exponential backoff with jitter
3. **Monitor Connection Health** - Track uptime and reconnections
4. **Implement Circuit Breakers** - Prevent cascading failures
5. **Buffer Events at Source** - Client-side event buffering
6. **Set Proper Timeouts** - Fail fast, retry faster
7. **Use Health Checks** - Liveness and readiness probes
8. **Deploy Multi-AZ** - Reduce single point of failure
9. **Rate Limit Reconnections** - Prevent thundering herd
10. **Log All Network Errors** - Debug and post-mortem

---

## 📚 Related Documentation

- [SRE-010: Dead Letter Queue Management](./SRE-010-dead-letter-queue-management.md)
- [SRE-006: Disaster Recovery](./SRE-006-disaster-recovery.md)
- [RabbitMQ Network Partitions](https://www.rabbitmq.com/partitions.html)
- [Knative Eventing Reliability](https://knative.dev/docs/eventing/)

---

## Revision History | Version | Date | Author | Changes | |--------- | ------ | -------- | --------- | | 1.0.0 | 2025-10-29 | Bruno Lucena (Principal SRE) | Initial network partition resilience runbook |

