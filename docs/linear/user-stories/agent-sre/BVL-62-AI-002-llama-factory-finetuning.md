# 🎯 AI-002: LLaMA Factory Integration for Local LLM Fine-Tuning

**Linear URL**: https://linear.app/bvlucena/issue/BVL-62/ai-002-llama-factory-finetuning  

---

## 📋 User Story

**As an** SRE Engineer  
**I want** agent-sre to use LLaMA Factory for local LLM fine-tuning at lowest cost and most private possible  
**So that** agent-sre can evolve from predictable scripts to intelligent code generation while maintaining privacy and cost efficiency


---


## 🎯 Acceptance Criteria

> **Note**: Features are already implemented. This ticket focuses on **validation** to ensure correctness, reliability, and production readiness.


- [ ] LLaMA Factory integrated into agent-sre training pipeline
- [ ] Agent-sre can fine-tune LLMs locally using LLaMA Factory
- [ ] Support for multiple LLM architectures (LLaMA, Qwen, Yi, etc.)
- [ ] Fine-tuning datasets generated from historical remediation actions
- [ ] Training pipeline automated for continuous improvement
- [ ] Local deployment for privacy (no data leaves infrastructure)
- [ ] Cost-optimized training (LoRA, QLoRA, quantization)
- [ ] Model evaluation and selection pipeline
- [ ] A/B testing framework for model comparison
- [ ] Model versioning and rollback capability

---

## 🔐 Security Acceptance Criteria

- [ ] Training data access requires authentication
- [ ] Training data encrypted at rest
- [ ] Training data encrypted in transit
- [ ] Access control for training data and models
- [ ] Audit logging for all training operations
- [ ] Model security validation (adversarial testing)
- [ ] Secrets management for training credentials
- [ ] Model artifacts signed with cryptographic signatures
- [ ] Security testing included in CI/CD pipeline
- [ ] Threat model reviewed and documented

## 🔄 Complete Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│          LLaMA FACTORY FINE-TUNING WORKFLOW                           │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ⏱️  Phase 1: DATA COLLECTION                                        │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE collects training data:                   │            │
│  │                                                      │            │
│  │  1. Historical Prometheus alerts                    │            │
│  │     - Alert patterns                                │            │
│  │     - Remediation actions taken                     │            │
│  │     - Success/failure outcomes                      │            │
│  │                                                      │            │
│  │  2. Runbook actions                                 │            │
│  │     - Successful remediation steps                  │            │
│  │     - Failed remediation attempts                   │            │
│  │     - Context and labels                            │            │
│  │                                                      │            │
│  │  3. Linear issue resolutions                        │            │
│  │     - Issue descriptions                            │            │
│  │     - Resolution steps                              │            │
│  │     - Root cause analysis                           │            │
│  │                                                      │            │
│  │  4. Code changes from PRs                           │            │
│  │     - Infrastructure-as-code fixes                  │            │
│  │     - Configuration changes                         │            │
│  │     - Remediation scripts                           │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  Phase 2: DATA PREPARATION                                        │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Format data for LLaMA Factory:                      │            │
│  │                                                      │            │
│  │  Dataset Format (JSONL):                             │            │
│  │  {                                                    │            │
│  │    "instruction": "Remediate FluxReconciliationFailure",│            │
│  │    "input": "Kustomization 'app' in namespace 'production' failed",│            │
│  │    "output": "Call LambdaFunction 'flux-reconcile-kustomization' with parameters: {name: 'app', namespace: 'production'}"│            │
│  │  }                                                    │            │
│  │                                                      │            │
│  │  Dataset Structure:                                  │            │
│  │  - Training set: 80%                                 │            │
│  │  - Validation set: 10%                               │            │
│  │  - Test set: 10%                                     │            │
│  │                                                      │            │
│  │  Dataset Size:                                       │            │
│  │  - Minimum: 1,000 examples                           │            │
│  │  - Target: 10,000 examples                           │            │
│  │  - Maximum: 100,000 examples                         │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  Phase 3: MODEL SELECTION                                         │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Select base model for fine-tuning:                  │            │
│  │                                                      │            │
│  │  Options:                                            │            │
│  │  1. Qwen2-1.5B (smallest, fastest)                  │            │
│  │  2. Qwen2-7B (balanced)                              │            │
│  │  3. LLaMA-3-8B (larger, more capable)                │            │
│  │                                                      │            │
│  │  Selection Criteria:                                 │            │
│  │  - Model size vs capability tradeoff                │            │
│  │  - Training cost                                    │            │
│  │  - Inference latency                                │            │
│  │  - Resource availability                            │            │
│  │                                                      │            │
│  │  Decision: Start with Qwen2-7B (balanced)            │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  Phase 4: TRAINING CONFIGURATION                                  │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Configure LLaMA Factory training:                   │            │
│  │                                                      │            │
│  │  Training Method: LoRA (Low-Rank Adaptation)        │            │
│  │  - Parameter-efficient fine-tuning                   │            │
│  │  - Lower memory requirements                        │            │
│  │  - Faster training                                  │            │
│  │                                                      │            │
│  │  Quantization: 4-bit QLoRA                          │            │
│  │  - 4-bit quantization for memory efficiency          │            │
│  │  - Maintains model quality                          │            │
│  │  - Enables training on consumer GPUs                │            │
│  │                                                      │            │
│  │  Training Parameters:                                │            │
│  │  - Learning rate: 2e-4                              │            │
│  │  - Batch size: 8                                    │            │
│  │  - Gradient accumulation: 4                         │            │
│  │  - Epochs: 3                                        │            │
│  │  - Warmup steps: 100                                │            │
│  │                                                      │            │
│  │  Optimization: BAdam (Bitwise Adam)                 │            │
│  │  - Memory-efficient optimizer                       │            │
│  │  - Suitable for quantization                        │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  Phase 5: TRAINING EXECUTION                                      │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Execute training via LLaMA Factory:                 │            │
│  │                                                      │            │
│  │  Training Command:                                   │            │
│  │  llamafactory-cli train \                            │            │
│  │    --model_name_or_path qwen/Qwen2-7B \              │            │
│  │    --dataset sre_remediation \                       │            │
│  │    --template qwen \                                  │            │
│  │    --finetuning_type lora \                          │            │
│  │    --lora_target all \                               │            │
│  │    --output_dir ./outputs/qwen2-7b-sre-lora \        │            │
│  │    --per_device_train_batch_size 8 \                 │            │
│  │    --gradient_accumulation_steps 4 \                 │            │
│  │    --lr_scheduler_type cosine \                      │            │
│  │    --logging_steps 10 \                              │            │
│  │    --save_steps 500 \                                │            │
│  │    --learning_rate 2e-4 \                            │            │
│  │    --num_train_epochs 3 \                            │            │
│  │    --quantization_bit 4 \                            │            │
│  │    --optim adamw_bits_8bit                           │            │
│  │                                                      │            │
│  │  Training Output:                                    │            │
│  │  - Training loss per step                            │            │
│  │  - Validation loss per epoch                         │            │
│  │  - Model checkpoints                                 │            │
│  │  - Training metrics (WandB/TensorBoard)             │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  Phase 6: MODEL EVALUATION                                        │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Evaluate fine-tuned model:                          │            │
│  │                                                      │            │
│  │  Evaluation Metrics:                                 │            │
│  │  1. Remediation selection accuracy                   │            │
│  │     - Correct LambdaFunction selected                │            │
│  │     - Target: >90% accuracy                          │            │
│  │                                                      │            │
│  │  2. Parameter extraction accuracy                    │            │
│  │     - Correct parameters extracted                   │            │
│  │     - Target: >95% accuracy                          │            │
│  │                                                      │            │
│  │  3. Response latency                                 │            │
│  │     - Inference time <500ms                          │            │
│  │     - Total response time <2s                        │            │
│  │                                                      │            │
│  │  4. False positive rate                              │            │
│  │     - Unnecessary remediations                       │            │
│  │     - Target: <5% false positives                    │            │
│  │                                                      │            │
│  │  Test Dataset:                                       │            │
│  │  - 1,000 held-out examples                           │            │
│  │  - Representative of production alerts               │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  Phase 7: MODEL DEPLOYMENT                                        │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Deploy fine-tuned model to production:              │            │
│  │                                                      │            │
│  │  Deployment Strategy:                                │            │
│  │  1. A/B Testing (50/50 split)                        │            │
│  │     - 50% traffic to old model                       │            │
│  │     - 50% traffic to new model                       │            │
│  │     - Monitor metrics for 1 week                     │            │
│  │                                                      │            │
│  │  2. Gradual Rollout (if A/B test passes)             │            │
│  │     - Week 1: 10% traffic                            │            │
│  │     - Week 2: 50% traffic                            │            │
│  │     - Week 3: 100% traffic                           │            │
│  │                                                      │            │
│  │  3. Rollback Capability                              │            │
│  │     - Automatic rollback on metrics degradation      │            │
│  │     - Manual rollback via configuration              │            │
│  │                                                      │            │
│  │  Model Serving:                                      │            │
│  │  - vLLM for fast inference                           │            │
│  │  - Local deployment (no external API)                │            │
│  │  - Kubernetes deployment via LambdaAgent             │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  Phase 8: CONTINUOUS IMPROVEMENT                                  │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Continuous learning loop:                           │            │
│  │                                                      │            │
│  │  1. Monitor production metrics                       │            │
│  │     - Remediation success rate                       │            │
│  │     - False positive rate                            │            │
│  │     - User feedback (Linear issue resolution)        │            │
│  │                                                      │            │
│  │  2. Collect new training data                        │            │
│  │     - Successful remediations                        │            │
│  │     - Failed remediations (for learning)             │            │
│  │     - Edge cases                                     │            │
│  │                                                      │            │
│  │  3. Retrain model periodically                       │            │
│  │     - Weekly retraining on new data                  │            │
│  │     - Monthly full retraining                        │            │
│  │     - Incremental learning                           │            │
│  │                                                      │            │
│  │  4. Model versioning                                 │            │
│  │     - Semantic versioning (v1.0.0, v1.1.0, etc.)     │            │
│  │     - Model registry (Hugging Face/Artifactory)      │            │
│  │     - Rollback capability                            │            │
│  └──────────────────────────────────────────────────────┘            │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 🏗️ Architecture Integration

### LLaMA Factory Components

1. **Training Pipeline**
   - Data preparation and formatting
   - Model selection and configuration
   - LoRA/QLoRA fine-tuning
   - Model evaluation and selection

2. **Model Serving**
   - vLLM inference engine
   - Local deployment (privacy-preserving)
   - Kubernetes integration via LambdaAgent
   - A/B testing framework

3. **Continuous Learning**
   - Production data collection
   - Periodic retraining
   - Model versioning and rollback
   - Performance monitoring

### Integration Points

```
┌─────────────────────────────────────────────────────────────┐
│          LLaMA FACTORY INTEGRATION ARCHITECTURE                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Agent-SRE Production                                       │
│       ↓                                                     │
│  ├─→ Collect Training Data                                  │
│  │   ├─→ Historical alerts                                  │
│  │   ├─→ Remediation actions                                │
│  │   └─→ Issue resolutions                                  │
│  ├─→ LLaMA Factory Training                                 │
│  │   ├─→ Data preparation                                   │
│  │   ├─→ Model fine-tuning (LoRA/QLoRA)                    │
│  │   └─→ Model evaluation                                   │
│  ├─→ Model Deployment                                       │
│  │   ├─→ A/B testing                                        │
│  │   ├─→ Gradual rollout                                    │
│  │   └─→ vLLM serving                                       │
│  └─→ Continuous Improvement                                 │
│       ├─→ Production monitoring                             │
│       ├─→ Data collection                                   │
│       └─→ Periodic retraining                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔧 Implementation Details

### 1. Data Collection Service

```python
# src/sre_agent/llama_factory/training_data_collector.py
from typing import List, Dict, Any
import json
from datetime import datetime, timedelta

class TrainingDataCollector:
    """Collect training data from agent-sre production runs."""
    
    def __init__(self):
        self.linear_client = LinearClient()
        self.prometheus_client = PrometheusClient()
        self.metrics_collector = MetricsCollector()
    
    async def collect_remediation_data(
        self,
        days_back: int = 30
    ) -> List[Dict[str, Any]]:
        """
        Collect historical remediation data for training.
        
        Args:
            days_back: Number of days to look back
            
        Returns:
            List of training examples in LLaMA Factory format
        """
        training_examples = []
        
        # Query resolved Linear issues
        start_date = datetime.now() - timedelta(days=days_back)
        issues = await self.linear_client.list_issues(
            state="completed",
            created_at_after=start_date.isoformat()
        )
        
        for issue in issues:
            # Extract alert information
            alert_data = self._extract_alert_from_issue(issue)
            
            # Extract remediation action
            remediation_action = self._extract_remediation_action(issue)
            
            # Extract outcome (success/failure)
            outcome = self._extract_outcome(issue)
            
            # Format as training example
            example = {
                "instruction": f"Remediate {alert_data['alertname']} alert",
                "input": self._format_input(alert_data),
                "output": self._format_output(remediation_action, outcome)
            }
            
            training_examples.append(example)
        
        return training_examples
    
    def _format_input(self, alert_data: Dict[str, Any]) -> str:
        """Format alert data as input for training."""
        return f"""
Alert: {alert_data['alertname']}
Labels: {json.dumps(alert_data.get('labels', {}))}
Annotations: {json.dumps(alert_data.get('annotations', {}))}
Description: {alert_data.get('description', '')}
        """.strip()
    
    def _format_output(
        self,
        remediation_action: Dict[str, Any],
        outcome: str
    ) -> str:
        """Format remediation action as output for training."""
        return f"""
LambdaFunction: {remediation_action.get('lambda_function')}
Parameters: {json.dumps(remediation_action.get('parameters', {}))}
Outcome: {outcome}
        """.strip()
```

### 2. Training Service

```python
# src/sre_agent/llama_factory/training_service.py
from llama_factory import TrainArguments, run_sft
import os

class LLaMAFactoryTrainingService:
    """Service for fine-tuning LLMs using LLaMA Factory."""
    
    def __init__(self, model_name: str = "qwen/Qwen2-7B"):
        self.model_name = model_name
        self.output_dir = "./outputs/sre-remediation"
        
    async def fine_tune(
        self,
        dataset_path: str,
        config: Dict[str, Any]
    ) -> str:
        """
        Fine-tune model using LLaMA Factory.
        
        Args:
            dataset_path: Path to training dataset (JSONL)
            config: Training configuration
            
        Returns:
            Path to fine-tuned model
        """
        # Configure training arguments
        training_args = TrainArguments(
            model_name_or_path=self.model_name,
            dataset_dir="./data",
            dataset="sre_remediation",
            template="qwen",
            finetuning_type="lora",
            lora_target="all",
            output_dir=self.output_dir,
            per_device_train_batch_size=config.get("batch_size", 8),
            gradient_accumulation_steps=config.get("gradient_accumulation", 4),
            lr_scheduler_type="cosine",
            logging_steps=10,
            save_steps=500,
            learning_rate=config.get("learning_rate", 2e-4),
            num_train_epochs=config.get("epochs", 3),
            quantization_bit=config.get("quantization_bit", 4),
            optim=config.get("optimizer", "adamw_bits_8bit")
        )
        
        # Run training
        run_sft(training_args)
        
        return self.output_dir
    
    async def evaluate(
        self,
        model_path: str,
        test_dataset_path: str
    ) -> Dict[str, float]:
        """
        Evaluate fine-tuned model.
        
        Args:
            model_path: Path to fine-tuned model
            test_dataset_path: Path to test dataset
            
        Returns:
            Evaluation metrics
        """
        # Load test dataset
        test_data = self._load_test_dataset(test_dataset_path)
        
        # Evaluate model
        metrics = {
            "remediation_accuracy": 0.0,
            "parameter_accuracy": 0.0,
            "false_positive_rate": 0.0,
            "average_latency_ms": 0.0
        }
        
        correct_remediations = 0
        correct_parameters = 0
        false_positives = 0
        total_latency = 0.0
        
        for example in test_data:
            # Run inference
            start_time = time.time()
            prediction = await self._infer(model_path, example["input"])
            latency = (time.time() - start_time) * 1000
            total_latency += latency
            
            # Check remediation selection
            if prediction["lambda_function"] == example["output"]["lambda_function"]:
                correct_remediations += 1
            
            # Check parameter extraction
            if self._parameters_match(prediction["parameters"], example["output"]["parameters"]):
                correct_parameters += 1
            
            # Check false positives
            if prediction["lambda_function"] and not example["output"]["lambda_function"]:
                false_positives += 1
        
        metrics["remediation_accuracy"] = correct_remediations / len(test_data)
        metrics["parameter_accuracy"] = correct_parameters / len(test_data)
        metrics["false_positive_rate"] = false_positives / len(test_data)
        metrics["average_latency_ms"] = total_latency / len(test_data)
        
        return metrics
```

### 3. Model Deployment

```yaml
# k8s/lambdaagent-model-serving.yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaAgent
metadata:
  name: agent-sre-model-serving
  namespace: ai
  annotations:
    lambda.knative.io/model-version: "v1.0.0"
    lambda.knative.io/ab-test-percentage: "50"
spec:
  image:
    repository: ghcr.io/homelab/vllm-sre-remediation
    tag: "v1.0.0"
    pullPolicy: IfNotPresent
  llm:
    model: "qwen/Qwen2-7B"
    quantization: "awq"
    maxTokens: 2048
    temperature: 0.1
  scaling:
    minReplicas: 1
    maxReplicas: 5
    targetConcurrency: 10
  env:
    - name: MODEL_PATH
      value: "/models/sre-remediation-v1.0.0"
    - name: VLLM_HOST
      value: "0.0.0.0"
    - name: VLLM_PORT
      value: "8000"
```

---

## 📚 References

- [LLaMA Factory Documentation](https://llamafactory.readthedocs.io/)
- [LLaMA Factory GitHub](https://github.com/hiyouga/LLaMA-Factory)
- [Agent-SRE Training Documentation](../../docs/training.md)

---

## ✅ Definition of Done

- [ ] LLaMA Factory installed and configured
- [ ] Training data collection pipeline implemented
- [ ] Dataset preparation and formatting working
- [ ] Model fine-tuning pipeline operational
- [ ] Model evaluation framework implemented
- [ ] Model deployment via LambdaAgent working
- [ ] A/B testing framework operational
- [ ] Continuous learning loop implemented
- [ ] Model versioning and rollback working
- [ ] Documentation updated
- [ ] Integration tests passing

---

**Related Stories**:
- [AI-001: Data Formulator Integration](./BVL-61-AI-001-data-formulator-visualization.md)
- [AI-003: TinyRecursiveModels Integration](./BVL-63-AI-003-tiny-recursive-models.md)
- [AI-004: Agent-Lightning Integration](./BVL-64-AI-004-agent-lightning-rl.md)


## 🧪 Test Scenarios

### Scenario 1: Training Data Collection
1. Configure training data collection for last 30 days
2. Verify historical alerts collected from Prometheus
3. Verify remediation actions collected from agent-sre logs
4. Verify Linear issue resolutions collected
5. Verify training dataset formatted correctly (JSONL)
6. Verify dataset split correctly (80/10/10)
7. Verify dataset size sufficient (> 1000 examples)
8. Verify dataset quality validated (no duplicates, valid format)

### Scenario 2: Model Fine-Tuning
1. Select base model (Qwen2-7B)
2. Configure training with LoRA/QLoRA
3. Load training dataset
4. Execute fine-tuning using LLaMA Factory
5. Verify training completes successfully
6. Verify training loss decreases over epochs
7. Verify model checkpoints saved correctly
8. Verify training metrics recorded (loss, learning rate)

### Scenario 3: Model Evaluation
1. Load fine-tuned model
2. Evaluate on test dataset (1000 held-out examples)
3. Verify remediation selection accuracy > 90%
4. Verify parameter extraction accuracy > 95%
5. Verify inference latency < 500ms
6. Verify false positive rate < 5%
7. Verify evaluation metrics recorded
8. Verify model meets acceptance criteria

### Scenario 4: Model Deployment and A/B Testing
1. Deploy fine-tuned model to production
2. Configure A/B testing (50/50 split)
3. Verify 50% traffic routed to new model
4. Verify 50% traffic routed to old model
5. Monitor metrics for 7 days
6. Verify new model performance comparable or better
7. Verify no degradation in remediation success rate
8. Verify A/B test results tracked and analyzed

### Scenario 5: Model Gradual Rollout
1. Complete A/B test successfully
2. Start gradual rollout (10% traffic to new model)
3. Verify 10% traffic routed correctly
4. Monitor metrics for 1 week
5. Increase to 50% traffic
6. Monitor metrics for 1 week
7. Increase to 100% traffic
8. Verify no issues during rollout
9. Verify rollout metrics tracked

### Scenario 6: Model Rollback
1. Deploy new model version
2. Detect performance degradation (simulated)
3. Trigger automatic rollback
4. Verify rollback to previous model version
5. Verify traffic routed back to previous model
6. Verify system recovers after rollback
7. Verify rollback metrics recorded
8. Verify alert fires for rollback event

### Scenario 7: Continuous Learning Loop
1. Monitor production metrics for new model
2. Collect successful remediation examples
3. Collect failed remediation examples
4. Verify training data collected correctly
5. Trigger weekly retraining with new data
6. Verify retraining completes successfully
7. Verify retrained model evaluated
8. Verify retrained model deployed if better
9. Verify continuous learning metrics tracked

### Scenario 8: Model Versioning and Registry
1. Train new model version
2. Verify model versioned correctly (semantic versioning)
3. Verify model stored in model registry
4. Verify model metadata recorded (training date, metrics, dataset size)
5. Verify model retrieval works correctly
6. Verify model rollback using versioning
7. Verify version history tracked

## 📊 Success Metrics

- **Training Data Collection Success Rate**: > 99%
- **Model Fine-Tuning Success Rate**: > 95%
- **Remediation Selection Accuracy**: > 90%
- **Parameter Extraction Accuracy**: > 95%
- **Inference Latency**: < 500ms (P95)
- **False Positive Rate**: < 5%
- **A/B Test Success Rate**: > 90% (new model comparable or better)
- **Model Deployment Success Rate**: > 95%
- **Test Pass Rate**: 100%

---

**Last Updated**: January 08, 2026
**Owner**: SRE Team
**Status**: Validation Required