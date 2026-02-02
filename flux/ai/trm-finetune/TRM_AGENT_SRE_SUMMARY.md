# TRM Integration with Agent-SRE - Summary

## ✅ Confirmation

**TRM does NOT support tool calling** - it's a recursive reasoning model, not a function calling model.

**Solution**: Use TRM to output structured JSON/text that we parse to select Lambda functions and send CloudEvents.

---

## 📋 What Was Created

### 1. Test Framework
- **`src/test_trm_runbook.py`**: Generates test dataset from RUNBOOK.md
  - Parses runbook alerts → Lambda function mappings
  - Formats for TRM training
  - Output: `data/runbook_test_dataset.jsonl`

### 2. TRM Remediation Selector
- **`src/trm_remediation_selector.py`**: Standalone TRM selector
  - Uses TRM for recursive reasoning
  - Parses structured output
  - Returns Lambda function + parameters

### 3. Agent-SRE Integration
- **`agent-sre/src/sre_agent/trm_remediation.py`**: Integration module
  - `TRMRemediationSelector`: Selects Lambda functions using TRM
  - `select_remediation_with_trm()`: Main entry point
  - Sends CloudEvents to trigger Lambda functions

### 4. Lambda Function Handler Template
- **`src/lambda_trigger_handler.py`**: Template for Lambda functions
  - Receives CloudEvents from agent-sre
  - Executes remediation actions
  - Handles: Flux reconciliation, pod restart, scaling, etc.

### 5. Documentation
- **`docs/TRM_AGENT_SRE_INTEGRATION.md`**: Architecture and design
- **`docs/TRM_TESTING_GUIDE.md`**: Testing procedures
- **`docs/DESIGN_DECISION.md`**: Design rationale

---

## 🔄 Architecture Flow

```
Prometheus Alert (CloudEvent)
    ↓
Agent-SRE receives alert
    ↓
Check static annotations (fast path)
    ↓ (if not found)
TRM Reasoning (if enabled)
    ├─ TRM recursively reasons about alert
    ├─ Outputs structured JSON: {"lambda_function": "...", "parameters": {...}}
    └─ Parse output
    ↓
Send CloudEvent: io.homelab.agent-sre.lambda.trigger
    ↓
Lambda Function receives CloudEvent
    ↓
Execute remediation (Flux reconcile, pod restart, etc.)
```

---

## 🧪 Testing

### Generate Test Dataset
```bash
cd flux/ai/trm-finetune
python src/test_trm_runbook.py
```

### Test TRM Selector
```bash
python src/trm_remediation_selector.py '{
  "labels": {
    "alertname": "FluxReconciliationFailure",
    "name": "homepage",
    "namespace": "flux-system"
  }
}'
```

### Enable TRM in Agent-SRE
```yaml
env:
  - name: USE_TRM
    value: "true"
  - name: TRM_API_URL
    value: "http://trm-reasoning.ml-platform.svc:8080"
```

---

## 📊 Alert → Lambda Function Mappings

Based on RUNBOOK.md:

| Alert | Lambda Function | Parameters |
|-------|----------------|------------|
| `FluxReconciliationFailure` | `flux-reconcile-kustomization` | `name`, `namespace` |
| `FluxGitRepositoryOutOfSync` | `flux-reconcile-gitrepository` | `name`, `namespace` |
| `FluxHelmReleaseFailing` | `flux-reconcile-helmrelease` | `name`, `namespace` |
| `PodCrashLoopBackOff` | `pod-restart` | `name`, `namespace`, `type` |
| `PrometheusServiceDown` | `pod-check-status` | `namespace`, `selector` |
| `PersistentVolumeFillingUpCritical` | `check-pvc-status` | `name`, `namespace` |

---

## 🎯 Next Steps

1. ✅ Generate runbook test dataset
2. ✅ Create TRM remediation selector
3. ✅ Integrate with agent-sre
4. ⏳ Fine-tune TRM on runbook + observability data
5. ⏳ Deploy TRM inference service
6. ⏳ Test end-to-end flow
7. ⏳ Create Lambda function CloudEvent handlers

---

## 📝 Key Files

- **Test**: `trm-finetune/src/test_trm_runbook.py`
- **Selector**: `trm-finetune/src/trm_remediation_selector.py`
- **Integration**: `agent-sre/src/sre_agent/trm_remediation.py`
- **Handler**: `trm-finetune/src/lambda_trigger_handler.py`
- **Docs**: `trm-finetune/docs/TRM_AGENT_SRE_INTEGRATION.md`

---

## ⚠️ Important Notes

1. **TRM doesn't support tool calling** - we parse text output
2. **Parsing can fail** - need robust error handling
3. **Fallback chain**: Static → TRM → FunctionGemma → Rule-based
4. **CloudEvents are used** to trigger Lambda functions (not direct calls)

---

## 🔗 Related Documentation

- [TRM Fine-Tuning Project](../README.md)
- [Agent-SRE Documentation](../../agent-sre/docs/)
- [Runbook](../../agent-sre/docs/RUNBOOK.md)

