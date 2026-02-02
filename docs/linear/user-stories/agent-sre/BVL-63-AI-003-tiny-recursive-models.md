# 🧠 AI-003: TinyRecursiveModels Integration for Recursive Reasoning

**Linear URL**: https://linear.app/bvlucena/issue/BVL-63/ai-003-tiny-recursive-models  

---

## 📋 User Story

**As an** SRE Engineer  
**I want** agent-sre to use TinyRecursiveModels (TRM) for recursive reasoning to select proper lambda functions  
**So that** agent-sre can solve complex remediation problems with a tiny 7M parameter model, achieving lowest cost and most private possible reasoning


---


## 🎯 Acceptance Criteria

> **Note**: Features are already implemented. This ticket focuses on **validation** to ensure correctness, reliability, and production readiness.


- [ ] TinyRecursiveModels (TRM) integrated into agent-sre remediation selection pipeline
- [ ] TRM used as Phase 1 in multi-phase remediation selection
- [ ] TRM model (7M parameters) deployed locally for privacy
- [ ] Recursive reasoning improves remediation selection accuracy
- [ ] TRM achieves >45% accuracy on ARC-AGI-like remediation tasks
- [ ] Inference latency <500ms per remediation decision
- [ ] Model fine-tuned on remediation dataset
- [ ] Integration with existing TRM infrastructure verified
- [ ] Fallback to RAG/Few-shot/AI if TRM confidence low
- [ ] TRM metrics tracked (inference count, confidence, fallback rate)

---

## 🔐 Security Acceptance Criteria

- [ ] TRM input validation and sanitization
- [ ] Protection against prompt injection attacks
- [ ] Confidence threshold validation (prevent low-confidence actions)
- [ ] TRM model output sanitization
- [ ] Access control for TRM inference service
- [ ] Audit logging for TRM inference operations
- [ ] Rate limiting on TRM inference requests
- [ ] Secrets management for TRM credentials
- [ ] Security testing included in CI/CD pipeline
- [ ] Threat model reviewed and documented

## 🔄 Complete Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│        TINY RECURSIVE MODELS REMEDIATION SELECTION                    │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ⏱️  t=0s: PROMETHEUS ALERT FIRES                                    │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Prometheus Alert: PodCPUHigh                         │            │
│  │  Labels: {pod: "app-xyz", namespace: "production"}     │            │
│  │  Annotations: {} (no lambda_function annotation)       │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=1s: AGENT-SRE RECEIVES CLOUDEVENT                             │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE extracts alert:                            │            │
│  │  - alertname: PodCPUHigh                               │            │
│  │  - labels: {pod, namespace}                            │            │
│  │  - annotations: {} (missing lambda_function)           │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  Phase 0: STATIC ANNOTATIONS CHECK (FAST PATH)                    │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Check if lambda_function annotation exists:         │            │
│  │                                                      │            │
│  │  if alert.annotations.get("lambda_function"):        │            │
│  │      return alert.annotations["lambda_function"]     │            │
│  │                                                      │            │
│  │  Result: None (annotation missing)                   │            │
│  │  → Proceed to Phase 1: TRM Reasoning                 │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  Phase 1: TRM RECURSIVE REASONING                                │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  TRM Model (7M parameters) performs recursive reasoning│            │
│  │                                                      │            │
│  │  Input:                                               │            │
│  │  - Question: "What remediation action for PodCPUHigh?" │            │
│  │  - Context:                                           │            │
│  │    - Alert: PodCPUHigh                                 │            │
│  │    - Labels: {pod: "app-xyz", namespace: "production"} │            │
│  │    - Current CPU: 85%                                  │            │
│  │    - Memory: 2.5Gi / 4Gi                               │            │
│  │    - Request rate: 1000 req/s                          │            │
│  │                                                      │            │
│  │  TRM Recursive Reasoning:                              │            │
│  │  ┌────────────────────────────────────────────────┐   │            │
│  │  │  Iteration 1: Initial reasoning                 │   │            │
│  │  │  - CPU is high (85%)                            │   │            │
│  │  │  - Memory is normal (62.5%)                      │   │            │
│  │  │  - Request rate is high (1000 req/s)            │   │            │
│  │  │  → Hypothesis: Traffic surge                     │   │            │
│  │  └────────────────────────────────────────────────┘   │            │
│  │  ┌────────────────────────────────────────────────┐   │            │
│  │  │  Iteration 2: Refined reasoning                 │   │            │
│  │  │  - If traffic surge, should scale horizontally  │   │            │
│  │  │  - Check if HPA is configured                    │   │            │
│  │  │  - Check if CPU limits are appropriate           │   │            │
│  │  │  → Refined: Scale pod or adjust CPU limits       │   │            │
│  │  └────────────────────────────────────────────────┘   │            │
│  │  ┌────────────────────────────────────────────────┐   │            │
│  │  │  Iteration 3: Final decision                     │   │            │
│  │  │  - Memory is normal, so not memory issue         │   │            │
│  │  │  - CPU is high, likely compute-bound             │   │            │
│  │  │  - Best action: Scale pod horizontally           │   │            │
│  │  │  → LambdaFunction: "scale-pod"                    │   │            │
│  │  │  → Parameters: {pod: "app-xyz", replicas: 3}      │   │            │
│  │  └────────────────────────────────────────────────┘   │            │
│  │                                                      │            │
│  │  TRM Output (structured JSON):                        │            │
│  │  {                                                    │            │
│  │    "lambda_function": "scale-pod",                    │            │
│  │    "parameters": {                                     │            │
│  │      "pod": "app-xyz",                                 │            │
│  │      "namespace": "production",                        │            │
│  │      "replicas": 3                                     │            │
│  │    },                                                  │            │
│  │    "confidence": 0.85,                                 │            │
│  │    "reasoning": "High CPU with normal memory suggests compute-bound workload. Scaling horizontally is appropriate."│            │
│  │  }                                                    │            │
│  │                                                      │            │
│  │  Confidence: 0.85 (>0.7 threshold)                    │            │
│  │  → Accept TRM recommendation                          │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=500ms: TRM RECOMMENDATION ACCEPTED                            │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE accepts TRM recommendation:               │            │
│  │  - LambdaFunction: "scale-pod"                        │            │
│  │  - Parameters: {pod: "app-xyz", replicas: 3}          │            │
│  │  - Confidence: 0.85                                   │            │
│  │  - Reasoning: "High CPU with normal memory..."        │            │
│  │                                                      │            │
│  │  Skip Phase 2 (RAG), Phase 3 (Few-shot), Phase 4 (AI)│            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=600ms: EXECUTE REMEDIATION                                    │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE calls LambdaFunction: "scale-pod"         │            │
│  │  Parameters: {pod: "app-xyz", replicas: 3}            │            │
│  │                                                      │            │
│  │  Remediation executed successfully                   │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  Alternative Path: LOW CONFIDENCE (<0.7)                         │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  If TRM confidence < 0.7:                             │            │
│  │  → Proceed to Phase 2: RAG-based selection            │            │
│  │  → Or Phase 3: Few-shot learning                      │            │
│  │  → Or Phase 4: AI function calling                    │            │
│  │                                                      │            │
│  │  Example:                                             │            │
│  │  TRM confidence: 0.45 (low)                           │            │
│  │  → Fallback to RAG                                    │            │
│  │  → RAG searches RUNBOOK.md                            │            │
│  │  → Returns: "scale-pod" (confidence: 0.80)            │            │
│  │  → Use RAG recommendation                             │            │
│  └──────────────────────────────────────────────────────┘            │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 🏗️ Architecture Integration

### TRM Components

1. **TRM Model**
   - 7M parameter neural network
   - 2-layer architecture
   - Recursive reasoning with H_cycles=3, L_cycles=6
   - Trained on remediation dataset

2. **Inference Service**
   - Local deployment (privacy-preserving)
   - Fast inference (<500ms)
   - Structured JSON output
   - Confidence scoring

3. **Integration Points**
   - Phase 1 in multi-phase remediation selection
   - Fallback to RAG/Few-shot/AI if needed
   - Metrics tracking (inference count, confidence, fallback)

### Integration Flow

```
┌─────────────────────────────────────────────────────────────┐
│          TRM REMEDIATION SELECTION PIPELINE                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Prometheus Alert                                           │
│       ↓                                                     │
│  Agent-SRE (receives CloudEvent)                           │
│       ↓                                                     │
│  Phase 0: Static Annotations (fast path)                    │
│       ↓ (if not found)                                      │
│  Phase 1: TRM Recursive Reasoning                           │
│       ├─→ TRM Model (7M params)                             │
│       ├─→ Recursive reasoning (H_cycles=3, L_cycles=6)      │
│       ├─→ Structured output (JSON)                          │
│       └─→ Confidence score                                  │
│       ↓                                                     │
│  Decision:                                                  │
│  ├─→ High confidence (>=0.7): Use TRM recommendation        │
│  └─→ Low confidence (<0.7): Fallback to Phase 2/3/4        │
│       ↓                                                     │
│  Execute Remediation (LambdaFunction)                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔧 Implementation Details

### 1. TRM Remediation Selector

```python
# src/sre_agent/trm_remediation.py (already exists, enhance it)
from typing import Dict, Any, Optional
import torch
from trm_model import TinyRecursiveModel

class TRMRemediationSelector:
    """TRM-based remediation selector for agent-sre."""
    
    def __init__(self, model_path: str):
        """
        Initialize TRM model.
        
        Args:
            model_path: Path to trained TRM model
        """
        self.model = TinyRecursiveModel.from_pretrained(model_path)
        self.model.eval()
        self.confidence_threshold = 0.7
        
    async def select_remediation(
        self,
        alert_data: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Select remediation using TRM recursive reasoning.
        
        Args:
            alert_data: Alert data from CloudEvent
            
        Returns:
            Remediation recommendation with confidence
        """
        # Format input for TRM
        question = f"What remediation action for {alert_data['alertname']}?"
        context = self._format_context(alert_data)
        
        # Initial embeddings
        question_emb = self._embed(question)
        context_emb = self._embed(context)
        answer_emb = self._embed("")  # Initial empty answer
        latent_z = torch.zeros(self.model.latent_dim)
        
        # Recursive reasoning loop (H_cycles=3, L_cycles=6)
        for h_cycle in range(3):  # High-level cycles
            for l_cycle in range(6):  # Low-level cycles
                # Update latent z (recursive reasoning)
                latent_z = self.model.update_latent(
                    question_emb,
                    answer_emb,
                    latent_z
                )
                
                # Update answer
                answer_emb = self.model.update_answer(
                    answer_emb,
                    latent_z
                )
        
        # Decode answer
        answer = self._decode_answer(answer_emb)
        
        # Parse structured output
        remediation = self._parse_output(answer)
        
        return remediation
    
    def _format_context(self, alert_data: Dict[str, Any]) -> str:
        """Format alert data as context for TRM."""
        context_parts = [
            f"Alert: {alert_data['alertname']}",
            f"Labels: {json.dumps(alert_data.get('labels', {}))}",
            f"Annotations: {json.dumps(alert_data.get('annotations', {}))}"
        ]
        
        # Add metrics if available
        if 'metrics' in alert_data:
            context_parts.append(f"Metrics: {json.dumps(alert_data['metrics'])}")
        
        return "\n".join(context_parts)
    
    def _parse_output(self, answer: str) -> Dict[str, Any]:
        """Parse TRM output into structured remediation."""
        try:
            # Try to parse as JSON
            remediation = json.loads(answer)
            
            # Validate structure
            if 'lambda_function' in remediation:
                return {
                    'lambda_function': remediation['lambda_function'],
                    'parameters': remediation.get('parameters', {}),
                    'confidence': remediation.get('confidence', 0.5),
                    'reasoning': remediation.get('reasoning', ''),
                    'method': 'trm'
                }
        except json.JSONDecodeError:
            pass
        
        # Fallback: Extract from text
        # Try to find lambda_function name in answer
        # (Simplified for brevity)
        return {
            'lambda_function': None,
            'parameters': {},
            'confidence': 0.3,
            'reasoning': answer,
            'method': 'trm'
        }
```

### 2. Integration with Intelligent Remediation

```python
# src/sre_agent/intelligent_remediation.py
from sre_agent.trm_remediation import TRMRemediationSelector

class IntelligentRemediation:
    """Multi-phase remediation selection."""
    
    def __init__(self):
        self.trm_selector = TRMRemediationSelector(
            model_path=os.getenv("TRM_MODEL_PATH", "/models/trm-sre-remediation.pt")
        )
        self.rag_selector = RAGRemediationSelector()
        self.fewshot_selector = FewShotRemediationSelector()
        self.ai_selector = AIRemediationSelector()
    
    async def select_remediation(
        self,
        alert_data: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Multi-phase remediation selection.
        
        Args:
            alert_data: Alert data from CloudEvent
            
        Returns:
            Selected remediation
        """
        # Phase 0: Static annotations (fast path)
        if 'annotations' in alert_data:
            lambda_function = alert_data['annotations'].get('lambda_function')
            if lambda_function:
                return {
                    'lambda_function': lambda_function,
                    'parameters': json.loads(
                        alert_data['annotations'].get('lambda_parameters', '{}')
                    ),
                    'confidence': 1.0,
                    'method': 'annotation'
                }
        
        # Phase 1: TRM recursive reasoning
        trm_result = await self.trm_selector.select_remediation(alert_data)
        
        if trm_result['confidence'] >= 0.7:
            # High confidence: use TRM recommendation
            logger.info(
                "trm_recommendation_accepted",
                lambda_function=trm_result['lambda_function'],
                confidence=trm_result['confidence'],
                method='trm'
            )
            return trm_result
        
        # Low confidence: fallback to Phase 2
        logger.info(
            "trm_recommendation_low_confidence",
            confidence=trm_result['confidence'],
            fallback='rag'
        )
        
        # Phase 2: RAG-based selection
        rag_result = await self.rag_selector.select_remediation(alert_data)
        
        if rag_result['confidence'] >= 0.7:
            return rag_result
        
        # Phase 3: Few-shot learning
        fewshot_result = await self.fewshot_selector.select_remediation(alert_data)
        
        if fewshot_result['confidence'] >= 0.7:
            return fewshot_result
        
        # Phase 4: AI function calling (fallback)
        return await self.ai_selector.select_remediation(alert_data)
```

### 3. TRM Model Training

```python
# scripts/train_trm_remediation.py
from trm_model import TinyRecursiveModel
import torch
from torch.utils.data import DataLoader

def train_trm_remediation_model(
    dataset_path: str,
    output_path: str,
    epochs: int = 50
):
    """
    Train TRM model on remediation dataset.
    
    Args:
        dataset_path: Path to training dataset
        output_path: Path to save trained model
        epochs: Number of training epochs
    """
    # Load dataset
    dataset = RemediationDataset(dataset_path)
    dataloader = DataLoader(dataset, batch_size=32, shuffle=True)
    
    # Initialize model (7M parameters, 2 layers)
    model = TinyRecursiveModel(
        input_dim=512,
        latent_dim=256,
        output_dim=512,
        num_layers=2,
        H_cycles=3,
        L_cycles=6
    )
    
    # Training configuration
    optimizer = torch.optim.AdamW(model.parameters(), lr=1e-4)
    criterion = torch.nn.CrossEntropyLoss()
    
    # Training loop
    for epoch in range(epochs):
        total_loss = 0.0
        
        for batch in dataloader:
            # Forward pass
            question_emb, context_emb, target_answer_emb = batch
            
            # Initial embeddings
            answer_emb = torch.zeros_like(target_answer_emb)
            latent_z = torch.zeros(batch_size, model.latent_dim)
            
            # Recursive reasoning
            for h_cycle in range(model.H_cycles):
                for l_cycle in range(model.L_cycles):
                    latent_z = model.update_latent(
                        question_emb,
                        answer_emb,
                        latent_z
                    )
                    answer_emb = model.update_answer(answer_emb, latent_z)
            
            # Compute loss
            loss = criterion(answer_emb, target_answer_emb)
            
            # Backward pass
            optimizer.zero_grad()
            loss.backward()
            optimizer.step()
            
            total_loss += loss.item()
        
        print(f"Epoch {epoch+1}/{epochs}, Loss: {total_loss/len(dataloader):.4f}")
    
    # Save model
    torch.save(model.state_dict(), output_path)
    print(f"Model saved to {output_path}")
```

---

## 📚 References

- [TinyRecursiveModels GitHub](https://github.com/SamsungSAILMontreal/TinyRecursiveModels)
- [TRM Paper](https://arxiv.org/abs/2510.04871)
- [Agent-SRE TRM Integration](../../docs/TRM_AGENT_SRE_INTEGRATION.md)

---

## ✅ Definition of Done

- [ ] TRM model (7M params) trained on remediation dataset
- [ ] TRM inference service deployed locally
- [ ] TRM integration in Phase 1 of remediation selection
- [ ] Recursive reasoning working (H_cycles=3, L_cycles=6)
- [ ] Structured output parsing implemented
- [ ] Confidence scoring and thresholding working
- [ ] Fallback to RAG/Few-shot/AI implemented
- [ ] TRM metrics tracked (inference count, confidence, fallback rate)
- [ ] Integration tests passing
- [ ] Documentation updated

---

**Related Stories**:
- [AI-001: Data Formulator Integration](./BVL-61-AI-001-data-formulator-visualization.md)
- [AI-002: LLaMA Factory Integration](./BVL-62-AI-002-llama-factory-finetuning.md)
- [AI-004: Agent-Lightning Integration](./BVL-64-AI-004-agent-lightning-rl.md)
- [SRE-001: Build Failure Investigation](./BVL-45-SRE-001-build-failure-investigation.md)


## 🧪 Test Scenarios

### Scenario 1: TRM Remediation Selection (High Confidence)
1. Receive alert without lambda_function annotation
2. Trigger TRM recursive reasoning
3. Verify TRM model loaded and ready
4. Verify recursive reasoning executes (H_cycles=3, L_cycles=6)
5. Verify structured output generated (JSON format)
6. Verify confidence score calculated (> 0.7)
7. Verify remediation selected correctly
8. Verify TRM recommendation accepted (no fallback)
9. Verify LambdaFunction executed with TRM parameters

### Scenario 2: TRM Remediation Selection (Low Confidence, Fallback)
1. Receive complex alert without lambda_function annotation
2. Trigger TRM recursive reasoning
3. Verify TRM model executes correctly
4. Verify confidence score calculated (< 0.7)
5. Verify fallback to Phase 2 (RAG) triggered
6. Verify RAG selects remediation
7. Verify TRM metrics recorded (low confidence, fallback)
8. Verify remediation executed with RAG recommendation

### Scenario 3: TRM Recursive Reasoning Validation
1. Provide alert context to TRM model
2. Verify recursive reasoning loop executes (3 H-cycles, 6 L-cycles per H-cycle)
3. Verify latent z updated at each iteration
4. Verify answer embedding updated at each iteration
5. Verify final output includes structured remediation
6. Verify reasoning trace logged (if available)
7. Verify reasoning time < 500ms

### Scenario 4: TRM Output Parsing and Validation
1. Receive TRM output (structured JSON)
2. Verify output parsed correctly
3. Verify lambda_function extracted correctly
4. Verify parameters extracted correctly
5. Verify confidence score extracted and validated
6. Verify reasoning text extracted
7. Verify invalid output handled gracefully (fallback)

### Scenario 5: TRM Performance and Scalability
1. Send 100+ alerts simultaneously to TRM
2. Verify TRM handles concurrent requests
3. Verify inference latency < 500ms per request (P95)
4. Verify no memory leaks during high load
5. Verify TRM model serves requests correctly under load
6. Verify metrics recorded for all requests
7. Verify system recovers after load decreases

### Scenario 6: TRM Model Loading and Initialization
1. Start agent-sre service
2. Verify TRM model loaded on startup
3. Verify model loaded from configured path
4. Verify model initialized correctly (eval mode)
5. Verify model ready for inference
6. Verify initialization time acceptable (< 5 seconds)
7. Verify model version/logged

### Scenario 7: TRM Fallback Chain Validation
1. Configure multi-phase fallback (TRM → RAG → Few-shot → AI)
2. Trigger remediation selection
3. Verify TRM executes first (Phase 1)
4. If TRM confidence low, verify RAG executes (Phase 2)
5. If RAG confidence low, verify Few-shot executes (Phase 3)
6. If Few-shot confidence low, verify AI executes (Phase 4)
7. Verify fallback chain metrics recorded (which phase selected)

### Scenario 8: TRM Confidence Threshold Validation
1. Configure confidence threshold (default 0.7)
2. Test with high confidence output (0.85)
3. Verify TRM recommendation accepted
4. Test with medium confidence output (0.65)
5. Verify TRM recommendation rejected, fallback triggered
6. Verify confidence threshold configurable
7. Verify threshold changes take effect

## 📊 Success Metrics

- **TRM Inference Success Rate**: > 95%
- **TRM Inference Latency**: < 500ms (P95)
- **TRM Remediation Selection Accuracy**: > 85% (on test dataset)
- **TRM Confidence Accuracy**: > 90% (confidence correlates with success)
- **TRM Fallback Rate**: < 30% (70% handled by TRM)
- **TRM Model Load Time**: < 5 seconds
- **Test Pass Rate**: 100%

---

**Last Updated**: January 08, 2026
**Owner**: SRE Team
**Status**: Validation Required