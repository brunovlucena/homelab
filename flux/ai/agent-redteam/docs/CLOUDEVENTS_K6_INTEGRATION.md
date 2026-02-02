# 🔴 CloudEvents → K6 Tests Integration

Agent-redteam receives CloudEvents from knative-lambda-operator and automatically triggers k6 attack/vulnerability assessment tests.

## 📡 Event Flow

```
knative-lambda-operator (emits event)
    ↓
Broker (RabbitMQ)
    ↓
Trigger (created by operator for LambdaAgent)
    ↓
agent-redteam (receives CloudEvent)
    ↓
Creates TestRun CRD (k6.io/v1alpha1)
    ↓
k6 operator executes test
    ↓
Metrics exported to Prometheus
```

## 🎯 Event Mappings

| Operator Event | K6 Test Triggered | Purpose |
|----------------|-------------------|---------|
| `io.knative.lambda.lifecycle.function.created` | `attack-sequential` | Test new function with sequential exploits |
| `io.knative.lambda.lifecycle.function.ready` | `attack-parallel` | Test ready function with parallel attacks |
| `io.knative.lambda.lifecycle.build.failed` | `vulnerability-assessment` | Full assessment when build fails |
| `io.knative.lambda.notification.alert.critical` | `random-chaos` | Chaos testing on critical alerts |
| `io.knative.lambda.lifecycle.service.scaled` | `attack-parallel` | Load simulation when service scales |

## 🔧 Manual K6 Test Trigger

You can also explicitly trigger k6 tests via CloudEvent:

**Event Type**: `io.homelab.redteam.test.k6`

**Payload**:
```json
{
  "test_type": "attack-sequential",
  "namespace": "agent-redteam",
  "env_vars": {
    "CUSTOM_VAR": "value"
  }
}
```

**Valid test types**:
- `smoke` - Basic functionality validation
- `attack-sequential` - Sequential exploit attacks
- `attack-parallel` - Parallel exploit attacks
- `vulnerability-assessment` - Full vulnerability assessment
- `random-chaos` - Random exploit chaos testing

## 📋 LambdaAgent Subscriptions

The agent-redteam LambdaAgent is configured to receive these events:

```yaml
subscriptions:
  # Operator lifecycle events
  - eventType: io.knative.lambda.lifecycle.function.created
  - eventType: io.knative.lambda.lifecycle.function.ready
  - eventType: io.knative.lambda.lifecycle.build.failed
  - eventType: io.knative.lambda.notification.alert.critical
  - eventType: io.knative.lambda.lifecycle.service.scaled
  
  # Manual k6 test trigger
  - eventType: io.homelab.redteam.test.k6
```

## 🚀 Example: Trigger K6 Test via CloudEvent

```bash
# Send event to trigger sequential attack test
kubectl run -n agent-redteam ce-k6-test --rm -i --restart=Never \
  --image=curlimages/curl:latest -- \
  curl -X POST http://agent-redteam-broker.agent-redteam.svc.cluster.local \
  -H "Content-Type: application/cloudevents+json" \
  -H "Ce-Type: io.homelab.redteam.test.k6" \
  -H "Ce-Source: /test/k6-trigger" \
  -d '{
    "test_type": "attack-sequential",
    "namespace": "agent-redteam"
  }'
```

## 🔄 Automatic Test Execution

When knative-lambda-operator emits events, agent-redteam automatically:

1. **Receives CloudEvent** via Trigger (created by operator)
2. **Maps event to test type** using event-to-test mapping
3. **Creates TestRun CRD** via K8s API
4. **k6 operator executes test** automatically
5. **Metrics exported** to Prometheus

## 📊 TestRun CRD Structure

The agent-redteam creates TestRun CRDs with:

- **Name**: `agent-redteam-{test-type}-{timestamp}`
- **Namespace**: `agent-redteam` (or specified)
- **ConfigMap**: References existing k6 test script ConfigMap
- **Environment Variables**: Includes trigger event info
- **Resources**: Configured based on test type

## 🎯 Use Cases

### 1. Continuous Security Testing
When a new LambdaFunction is created, automatically test it with sequential exploits.

### 2. Load Testing
When a service scales, trigger parallel attack test to simulate load.

### 3. Incident Response
When a critical alert fires, trigger chaos testing to validate resilience.

### 4. Build Failure Analysis
When a build fails, run full vulnerability assessment to identify issues.

## ⚙️ Configuration

TestRun resources are configured based on test type:

| Test Type | Memory | CPU |
|-----------|--------|-----|
| smoke | 256Mi | 200m |
| attack-sequential | 512Mi | 500m |
| attack-parallel | 1Gi | 1000m |
| vulnerability-assessment | 512Mi | 500m |
| random-chaos | 512Mi | 500m |

## 📝 Notes

- All TestRuns are created in the `agent-redteam` namespace
- ConfigMaps must exist before TestRuns can be created
- TestRuns are automatically cleaned up by k6 operator after completion
- Metrics are exported to Prometheus for monitoring
