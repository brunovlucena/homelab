# Design Decision: TRM Integration with Agent-SRE

## Question

Can TRM be used to select Lambda functions for Prometheus alerts, following the runbook?

## Answer

**Yes, but with important limitations.**

## Key Finding

**TRM does NOT support tool calling** - it's a recursive reasoning model, not a function calling model like FunctionGemma.

## Solution Design

### Approach: Structured Text Output → CloudEvents

Since TRM can't call functions directly, we use a **two-step process**:

1. **TRM Reasoning**: TRM recursively reasons about the alert and outputs structured text/JSON
2. **Parse & Trigger**: Agent-sre parses TRM output and sends CloudEvent to trigger Lambda function

### Architecture

```
Alert → Agent-SRE → TRM Reasoning → Parse JSON → CloudEvent → Lambda Function
```

### Flow Details

1. **Alert Received**: `io.homelab.prometheus.alert.fired`
2. **TRM Reasoning**: 
   - Input: Alert details (name, labels, annotations)
   - Process: Recursive reasoning (up to 10 iterations)
   - Output: Structured text containing JSON
3. **Parse Output**: Extract `lambda_function` and `parameters` from TRM output
4. **Send CloudEvent**: `io.homelab.agent-sre.lambda.trigger` with function name and parameters
5. **Lambda Execution**: Lambda function receives CloudEvent and executes remediation

## Comparison: TRM vs FunctionGemma

| Aspect | FunctionGemma 270M | TRM 7M |
|--------|-------------------|--------|
| **Tool Calling** | ✅ Native (function schemas) | ❌ No (text output only) |
| **Recursive Reasoning** | ❌ No | ✅ Yes (core feature) |
| **Structured Output** | ✅ Guaranteed (schemas) | ⚠️ Parsing required |
| **Model Size** | 270M params | 7M params (38x smaller) |
| **Inference Speed** | Slower | Faster (smaller model) |
| **Reliability** | High (schema validation) | Medium (parsing can fail) |
| **Use Case** | Function calling, structured output | Recursive problem solving |

## Advantages of TRM Approach

1. **Recursive Reasoning**: Can reason through complex multi-step problems
2. **Smaller Model**: 7M params vs 270M (38x smaller, faster inference)
3. **Domain-Specific**: Fine-tuned on your runbook and observability data
4. **Cost Efficient**: Lower compute requirements

## Limitations

1. **No Native Tool Calling**: Must parse text output (less reliable)
2. **Parsing Required**: Need robust JSON extraction from text
3. **Error Handling**: More complex than native function calling
4. **Output Format**: Not guaranteed to be valid JSON

## Mitigation Strategies

### 1. Robust Parsing

```python
def _parse_trm_output(output_text: str) -> Dict[str, Any]:
    # Try multiple parsing strategies:
    # 1. Direct JSON extraction
    # 2. JSON in code blocks
    # 3. Regex extraction
    # 4. Rule-based fallback
```

### 2. Training on Structured Output

Fine-tune TRM to output JSON consistently:

```json
{
  "lambda_function": "flux-reconcile-kustomization",
  "parameters": {"name": "...", "namespace": "..."},
  "reasoning": "..."
}
```

### 3. Fallback Chain

```
Static Annotation (fastest, most reliable)
    ↓ (if not found)
TRM Reasoning (if enabled)
    ↓ (if fails or no function selected)
FunctionGemma (fallback)
    ↓ (if fails)
Rule-Based (last resort)
```

## Testing Strategy

### 1. Unit Tests

- Test TRM output parsing
- Test JSON extraction
- Test parameter extraction

### 2. Integration Tests

- Test end-to-end flow: Alert → TRM → CloudEvent → Lambda
- Test fallback mechanisms
- Test error handling

### 3. Validation Tests

- Test on runbook examples
- Verify correct Lambda function selection
- Verify parameter extraction

## Recommendation

**Use TRM as an alternative to FunctionGemma**, not a replacement:

- **Enable TRM** when you want recursive reasoning for complex alerts
- **Keep FunctionGemma** as fallback for reliability
- **Use static annotations** for known, simple cases (fastest path)

## Implementation Status

- ✅ TRM remediation selector created
- ✅ CloudEvent trigger design
- ✅ Integration with agent-sre
- ⏳ TRM fine-tuning on runbook (next step)
- ⏳ TRM inference service deployment
- ⏳ End-to-end testing

## Conclusion

TRM can be used for Lambda function selection, but requires:
1. Structured output training
2. Robust parsing
3. Fallback mechanisms
4. Careful error handling

The recursive reasoning capability is valuable for complex alerts, but the lack of native tool calling makes it less reliable than FunctionGemma for this use case.

