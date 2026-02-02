# TRM Integration with Agent-SRE

## Overview

TRM (Tiny Recursive Model) is integrated with agent-sre to provide **recursive reasoning** for Lambda function selection. Unlike FunctionGemma which supports tool calling, TRM uses **structured text output** that we parse to trigger Lambda functions via CloudEvents.

## Key Design Decision

**TRM does NOT support tool calling** - it's a recursive reasoning model, not a function calling model.

**Solution**: TRM outputs structured JSON/text that we parse to:
1. Select Lambda function
2. Extract parameters
3. Send CloudEvent to trigger Lambda function

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              Prometheus Alert CloudEvent                    │
│              io.homelab.prometheus.alert.fired             │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    Agent-SRE                                 │
│                                                              │
│  1. Receives CloudEvent                                      │
│  2. Checks static annotations (fast path)                   │
│  3. If no annotation → Use TRM for reasoning                 │
│     └─ TRM reasons about alert                               │
│     └─ Outputs: {"lambda_function": "...", "parameters": {...}} │
│  4. Parse TRM output                                         │
│  5. Send CloudEvent to trigger Lambda function              │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│         CloudEvent: io.homelab.agent-sre.lambda.trigger     │
│         {                                                    │
│           "lambda_function": "flux-reconcile-kustomization",│
│           "parameters": {"name": "...", "namespace": "..."}  │
│         }                                                    │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              Lambda Function (Knative Service)               │
│              Receives CloudEvent and executes remediation    │
└─────────────────────────────────────────────────────────────┘
```

## TRM vs FunctionGemma

| Feature | FunctionGemma 270M | TRM 7M |
|---------|-------------------|--------|
| **Tool Calling** | ✅ Native support | ❌ No support |
| **Recursive Reasoning** | ❌ No | ✅ Yes (core feature) |
| **Structured Output** | ✅ Function schemas | ⚠️ Text → JSON parsing |
| **Model Size** | 270M params | 7M params (38x smaller) |
| **Use Case** | Function calling, structured output | Recursive problem solving |

## Integration Flow

### 1. Alert Received

```python
# In agent-sre main.py
event_type = "io.homelab.prometheus.alert.fired"
alert_data = {
    "labels": {"alertname": "FluxReconciliationFailure", ...},
    "annotations": {}
}
```

### 2. TRM Reasoning

```python
# In trm_remediation.py
trm_result = await select_remediation_with_trm(
    alert_data=alert_data,
    trm_api_url="http://trm-reasoning.ml-platform.svc:8080"
)

# TRM outputs structured text:
# {
#   "lambda_function": "flux-reconcile-kustomization",
#   "parameters": {"name": "homepage", "namespace": "flux-system"},
#   "reasoning": "Alert indicates Kustomization reconciliation failure..."
# }
```

### 3. CloudEvent Trigger

```python
# Send CloudEvent to broker
event = CloudEvent(
    type="io.homelab.agent-sre.lambda.trigger",
    source="/agent-sre/remediation",
    data={
        "lambda_function": "flux-reconcile-kustomization",
        "parameters": {"name": "homepage", "namespace": "flux-system"},
        "alert": alert_data
    }
)
```

### 4. Lambda Function Execution

Lambda function receives CloudEvent and executes remediation.

## Configuration

### Enable TRM in Agent-SRE

```yaml
# In agent-sre deployment
env:
  - name: USE_TRM
    value: "true"  # Enable TRM reasoning
  - name: TRM_API_URL
    value: "http://trm-reasoning.ml-platform.svc:8080"
  - name: BROKER_URL
    value: "http://lambda-broker.knative-lambda.svc.cluster.local"
```

### TRM Inference Service

Deploy TRM as a Knative service:

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: trm-reasoning
  namespace: ml-platform
spec:
  template:
    spec:
      containers:
      - image: localhost:5001/trm-reasoning:latest
        env:
        - name: MODEL_PATH
          value: /models/trm-homelab-finetuned
        resources:
          requests:
            nvidia.com/gpu: "1"
```

## Training TRM on Runbook

### 1. Generate Training Data

```bash
cd flux/ai/trm-finetune
python src/test_trm_runbook.py

# Output: data/runbook_test_dataset.jsonl
```

### 2. Fine-Tune TRM

```bash
# Add runbook data to training
python src/data_collector.py  # Collects notifi-services + observability
# Manually merge runbook_test_dataset.jsonl into training data

# Fine-tune
python src/trm_trainer.py \
  --training-data data/merged_training_data.jsonl \
  --output-dir models/trm-agent-sre-finetuned
```

### 3. Deploy TRM Model

```bash
# Export model
python src/trm_trainer.py --export models/trm-agent-sre-finetuned/export

# Deploy to Ollama or VLLM
# Or deploy as inference service
```

## Testing

### Test TRM on Runbook Examples

```bash
# Generate test dataset
python src/test_trm_runbook.py

# Test TRM selector
python src/trm_remediation_selector.py '{
  "labels": {
    "alertname": "FluxReconciliationFailure",
    "name": "homepage",
    "namespace": "flux-system"
  }
}'
```

### Expected Output

```json
{
  "lambda_function": "flux-reconcile-kustomization",
  "parameters": {
    "name": "homepage",
    "namespace": "flux-system"
  },
  "reasoning": "Alert indicates Kustomization reconciliation failure, so reconcile it.",
  "confidence": 0.85,
  "method": "trm_recursive_reasoning"
}
```

## Lambda Function Trigger Design

### CloudEvent Schema

**Event Type**: `io.homelab.agent-sre.lambda.trigger`

**Payload**:
```json
{
  "lambda_function": "flux-reconcile-kustomization",
  "parameters": {
    "name": "homepage",
    "namespace": "flux-system"
  },
  "alert": {
    "labels": {...},
    "annotations": {...}
  },
  "triggered_by": "trm-reasoning",
  "correlation_id": "abc-123"
}
```

### Lambda Function Handler

Each Lambda function receives this CloudEvent and executes the remediation:

```python
# In lambda function (e.g., flux-reconcile-kustomization)
def handle(event, context):
    data = event.data
    
    lambda_function = data["lambda_function"]
    parameters = data["parameters"]
    alert = data["alert"]
    
    # Execute remediation
    if lambda_function == "flux-reconcile-kustomization":
        flux.reconcile_kustomization(
            name=parameters["name"],
            namespace=parameters["namespace"]
        )
```

## PrometheusRule → Lambda Function Mapping

Based on RUNBOOK.md, here are the mappings:

| Alert Name | Lambda Function | Parameters |
|------------|----------------|------------|
| `FluxReconciliationFailure` | `flux-reconcile-kustomization` | `name`, `namespace` |
| `FluxGitRepositoryOutOfSync` | `flux-reconcile-gitrepository` | `name`, `namespace` |
| `FluxHelmReleaseFailing` | `flux-reconcile-helmrelease` | `name`, `namespace` |
| `PodCrashLoopBackOff` | `pod-restart` | `name`, `namespace`, `type` |
| `PrometheusServiceDown` | `pod-check-status` → `pod-restart` | `namespace`, `selector` |
| `PersistentVolumeFillingUpCritical` | `check-pvc-status` | `name`, `namespace` |

## Advantages of TRM Approach

1. **Recursive Reasoning**: TRM can reason through complex multi-step problems
2. **Smaller Model**: 7M params vs 270M (38x smaller, faster inference)
3. **Domain-Specific**: Fine-tuned on your runbook and observability data
4. **Cost Efficient**: Lower compute requirements

## Limitations

1. **No Native Tool Calling**: Must parse text output (less reliable than function schemas)
2. **Parsing Required**: Need robust JSON extraction from text
3. **Error Handling**: More complex error handling than native function calling

## Fallback Strategy

```
Static Annotation (fastest)
    ↓ (if not found)
TRM Reasoning (if enabled)
    ↓ (if fails)
FunctionGemma (fallback)
    ↓ (if fails)
Rule-Based (last resort)
```

## Next Steps

1. ✅ Generate runbook test dataset
2. ✅ Create TRM remediation selector
3. ✅ Integrate with agent-sre
4. ⏳ Fine-tune TRM on runbook + observability data
5. ⏳ Deploy TRM inference service
6. ⏳ Test end-to-end flow
7. ⏳ Create Lambda function CloudEvent handlers

