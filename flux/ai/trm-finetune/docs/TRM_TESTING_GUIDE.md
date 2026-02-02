# TRM Testing Guide - Agent-SRE Integration

## Confirmation: TRM Does NOT Support Tool Calling

**✅ Confirmed**: TRM (Tiny Recursive Model) is a **recursive reasoning model**, not a function calling model. It does NOT have native tool calling support like FunctionGemma.

**Solution**: We use TRM to output **structured text/JSON** that we parse to select Lambda functions.

---

## Testing TRM on Runbook

### Step 1: Generate Test Dataset

```bash
cd flux/ai/trm-finetune

# Generate test examples from runbook
python src/test_trm_runbook.py

# Output: data/runbook_test_dataset.jsonl
# Contains: Alert → Lambda function mappings from RUNBOOK.md
```

**Expected Output**:
```
📚 Parsing runbook...
✅ Extracted 15 examples from runbook
📊 Formatting for TRM...
✅ Formatted 15 TRM examples
💾 Saving dataset...
💾 Saved 15 test examples to data/runbook_test_dataset.jsonl

🎉 Test dataset ready!
   Total examples: 15
   Output: data/runbook_test_dataset.jsonl

📋 Example alerts covered:
   - FluxReconciliationFailure → flux-reconcile-kustomization
   - FluxGitRepositoryOutOfSync → flux-reconcile-gitrepository
   - PodCrashLoopBackOff → pod-restart
   - PrometheusServiceDown → pod-check-status
   - PersistentVolumeFillingUpCritical → check-pvc-status
```

### Step 2: Test TRM Remediation Selector

```bash
# Test with a sample alert
python src/trm_remediation_selector.py '{
  "labels": {
    "alertname": "FluxReconciliationFailure",
    "name": "homepage",
    "namespace": "flux-system",
    "kind": "Kustomization"
  },
  "annotations": {}
}'
```

**Expected Output**:
```json
{
  "lambda_function": "flux-reconcile-kustomization",
  "parameters": {
    "name": "homepage",
    "namespace": "flux-system"
  },
  "reasoning": "Alert indicates Kustomization reconciliation failure, so reconcile it.",
  "confidence": 0.85,
  "method": "trm_recursive_reasoning",
  "iterations": 6
}
```

### Step 3: Test End-to-End Flow

```bash
# Simulate agent-sre receiving alert
curl -X POST http://agent-sre.ai.svc.cluster.local/ \
  -H "Content-Type: application/cloudevents+json" \
  -H "Ce-Type: io.homelab.prometheus.alert.fired" \
  -H "Ce-Source: /prometheus/alerts" \
  -d '{
    "data": {
      "labels": {
        "alertname": "FluxReconciliationFailure",
        "name": "homepage",
        "namespace": "flux-system"
      }
    }
  }'
```

---

## Training TRM on Runbook

### Option 1: Add Runbook to Fine-Tuning Pipeline

```python
# In data_collector.py, add runbook parsing
from test_trm_runbook import RunbookDatasetGenerator

# Collect runbook examples
runbook_gen = RunbookDatasetGenerator("/path/to/RUNBOOK.md")
runbook_examples = runbook_gen.parse_runbook()
trm_runbook_examples = runbook_gen.format_for_trm(runbook_examples)

# Merge with other training data
all_examples = code_examples + obs_examples + trm_runbook_examples
```

### Option 2: Fine-Tune Specifically on Runbook

```bash
# Fine-tune TRM only on runbook data
python src/trm_trainer.py \
  --training-data data/runbook_test_dataset.jsonl \
  --output-dir models/trm-runbook-only \
  --epochs 10000 \
  --eval-interval 1000
```

---

## Integration Testing

### Test 1: TRM API Endpoint

```bash
# Deploy TRM inference service
kubectl apply -f k8s/trm-reasoning-service.yaml

# Test inference
curl -X POST http://trm-reasoning.ml-platform.svc:8080/reason \
  -H "Content-Type: application/json" \
  -d '{
    "problem": "Analyze alert: FluxReconciliationFailure...",
    "max_iterations": 10
  }'
```

### Test 2: Agent-SRE with TRM

```bash
# Enable TRM in agent-sre
kubectl set env deployment/agent-sre \
  USE_TRM=true \
  TRM_API_URL=http://trm-reasoning.ml-platform.svc:8080

# Send test alert
# (See Step 3 above)
```

### Test 3: CloudEvent Flow

```bash
# Verify CloudEvent sent to broker
kubectl logs -n knative-lambda -l app=rabbitmq-broker | grep "lambda.trigger"

# Verify Lambda function received event
kubectl logs -n ai -l app=flux-reconcile-kustomization
```

---

## Validation Checklist

- [ ] TRM can parse runbook examples
- [ ] TRM outputs structured JSON
- [ ] JSON parsing extracts lambda_function correctly
- [ ] Parameters extracted from labels
- [ ] CloudEvent sent to broker
- [ ] Lambda function receives event
- [ ] Remediation executes successfully
- [ ] End-to-end flow works

---

## Troubleshooting

### TRM Output Not Parsable

**Problem**: TRM outputs free-form text, not JSON

**Solution**: 
1. Improve prompt to emphasize JSON output
2. Add JSON schema examples in prompt
3. Use regex fallback parsing
4. Fine-tune on JSON-structured examples

### Wrong Lambda Function Selected

**Problem**: TRM selects incorrect function

**Solution**:
1. Add more runbook examples to training
2. Include reasoning steps in training data
3. Use few-shot examples in prompt
4. Increase training epochs

### CloudEvent Not Received

**Problem**: Lambda function doesn't receive event

**Solution**:
1. Check broker connectivity
2. Verify Trigger configuration
3. Check Lambda function subscription
4. Review broker logs

---

## Performance Metrics

Track these metrics:

- **TRM Selection Accuracy**: % of correct function selections
- **Parsing Success Rate**: % of successfully parsed TRM outputs
- **CloudEvent Delivery Rate**: % of events successfully delivered
- **Remediation Success Rate**: % of successful remediations
- **End-to-End Latency**: Time from alert to remediation

---

## Next Steps

1. ✅ Generate test dataset
2. ✅ Create TRM selector
3. ✅ Integrate with agent-sre
4. ⏳ Fine-tune TRM on runbook
5. ⏳ Deploy TRM inference service
6. ⏳ Test end-to-end
7. ⏳ Monitor and iterate

