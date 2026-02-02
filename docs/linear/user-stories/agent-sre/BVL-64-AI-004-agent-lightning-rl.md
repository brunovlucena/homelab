# ⚡ AI-004: Agent-Lightning RL Training

**Linear URL**: https://linear.app/bvlucena/issue/BVL-226/backend-006-knative-service-management
**Linear URL**: https://linear.app/bvlucena/issue/BVL-198/ai-004-agent-lightning-rl-training  

---

## 📋 User Story

**As an** SRE Engineer  
**I want** agent-sre to use Agent-Lightning for reinforcement learning training  
**So that** agent-sre can optimize remediation selection and execution through continuous learning


---


## 🎯 Acceptance Criteria

> **Note**: Features are already implemented. This ticket focuses on **validation** to ensure correctness, reliability, and production readiness.


- [ ] Agent-Lightning integrated into agent-sre training pipeline
- [ ] Reward function based on remediation success rate
- [ ] RL training loop operational
- [ ] Model optimization through trial and error
- [ ] Performance metrics tracked (success rate, latency, false positives)
- [ ] Continuous learning from production feedback
- [ ] A/B testing framework for RL models
- [ ] Model versioning and rollback capability

---

## 🔄 Complete Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│          AGENT-LIGHTNING RL TRAINING WORKFLOW                          │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ⏱️  Phase 1: COLLECT TRAINING DATA                                  │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE collects remediation outcomes:             │            │
│  │  - Successful remediations                            │            │
│  │  - Failed remediations                                │            │
│  │  - Remediation latency                                │            │
│  │  - False positive rate                                │            │
│  │  - Alert resolution time                              │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  Phase 2: CALCULATE REWARDS                                       │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Reward Function:                                    │            │
│  │  - +10: Successful remediation                       │            │
│  │  - -5: Failed remediation                             │            │
│  │  - -1: False positive                                 │            │
│  │  - +2: Fast remediation (<30s)                        │            │
│  │  - -1: Slow remediation (>5min)                       │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  Phase 3: RL TRAINING                                             │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-Lightning trains agent-sre:                    │            │
│  │  - State: Alert context + system state                │            │
│  │  - Action: Remediation selection                      │            │
│  │  - Reward: Calculated from outcomes                   │            │
│  │  - Policy: Optimized through RL                       │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  Phase 4: DEPLOY OPTIMIZED MODEL                                  │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Deploy RL-optimized agent-sre:                       │            │
│  │  - A/B testing (50/50 split)                          │            │
│  │  - Monitor performance                                │            │
│  │  - Gradual rollout if successful                      │            │
│  └──────────────────────────────────────────────────────┘            │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Implementation Details

### Reward Function

```python
# src/sre_agent/rl_training.py
from typing import Dict, Any

class RewardFunction:
    """Calculate rewards for RL training."""
    
    def calculate_reward(
        self,
        outcome: Dict[str, Any]
    ) -> float:
        """
        Calculate reward based on remediation outcome.
        
        Args:
            outcome: Remediation outcome with metrics
            
        Returns:
            Reward value (higher is better)
        """
        reward = 0.0
        
        # Success reward
        if outcome.get("success", False):
            reward += 10.0
        else:
            reward -= 5.0
        
        # Latency reward
        latency = outcome.get("latency_seconds", 0)
        if latency < 30:
            reward += 2.0
        elif latency > 300:
            reward -= 1.0
        
        # False positive penalty
        if outcome.get("false_positive", False):
            reward -= 1.0
        
        # Alert resolution reward
        if outcome.get("alert_resolved", False):
            reward += 5.0
        
        return reward
```

---

## 📚 References

- [Agent-Lightning GitHub](https://github.com/lightning-ai/agent-lightning)
- [RL Training Documentation](../../docs/training.md)

---

## ✅ Definition of Done

- [ ] Agent-Lightning integrated
- [ ] Reward function implemented
- [ ] RL training loop operational
- [ ] Performance metrics tracked
- [ ] A/B testing framework working
- [ ] Documentation updated

---

**Related Stories**:
- [AI-002: LLaMA Factory Integration](./BVL-62-AI-002-llama-factory-finetuning.md)
- [AI-003: TinyRecursiveModels Integration](./BVL-63-AI-003-tiny-recursive-models.md)


## 🧪 Test Scenarios

### Scenario 1: Training Data Collection
1. Configure RL training data collection
2. Verify successful remediations collected
3. Verify failed remediations collected
4. Verify remediation latency recorded
5. Verify false positive rate recorded
6. Verify alert resolution time recorded
7. Verify training data formatted correctly
8. Verify training data validated (no duplicates, valid format)

### Scenario 2: Reward Function Calculation
1. Provide successful remediation outcome
2. Verify reward calculated correctly (+10 for success)
3. Provide failed remediation outcome
4. Verify penalty calculated correctly (-5 for failure)
5. Provide fast remediation (< 30s)
6. Verify latency reward added (+2)
7. Provide slow remediation (> 5min)
8. Verify latency penalty added (-1)
9. Provide false positive outcome
10. Verify false positive penalty added (-1)
11. Verify total reward calculated correctly

### Scenario 3: RL Training Loop
1. Configure RL training with Agent-Lightning
2. Load training dataset (state, action, reward tuples)
3. Execute RL training loop
4. Verify agent policy updated based on rewards
5. Verify training loss decreases over episodes
6. Verify model checkpoints saved correctly
7. Verify training metrics recorded (episode reward, loss)
8. Verify training completes successfully

### Scenario 4: RL Model Optimization
1. Train RL model on remediation dataset
2. Evaluate model on test dataset
3. Verify remediation success rate improved (> baseline)
4. Verify remediation latency reduced (< baseline)
5. Verify false positive rate reduced (< baseline)
6. Verify model optimization measurable
7. Verify optimization metrics tracked

### Scenario 5: RL Model Deployment and A/B Testing
1. Deploy RL-optimized agent-sre model
2. Configure A/B testing (50/50 split)
3. Verify 50% traffic routed to RL model
4. Verify 50% traffic routed to baseline model
5. Monitor performance for 7 days
6. Verify RL model performance comparable or better
7. Verify remediation success rate improved
8. Verify A/B test results analyzed

### Scenario 6: RL Model Gradual Rollout
1. Complete A/B test successfully
2. Start gradual rollout (10% traffic to RL model)
3. Monitor metrics for 1 week
4. Increase to 50% traffic
5. Monitor metrics for 1 week
6. Increase to 100% traffic
7. Verify no issues during rollout
8. Verify rollout metrics tracked

### Scenario 7: RL Model Rollback
1. Deploy RL-optimized model
2. Detect performance degradation (simulated)
3. Trigger automatic rollback
4. Verify rollback to previous model version
5. Verify traffic routed back to previous model
6. Verify system recovers after rollback
7. Verify rollback metrics recorded
8. Verify alert fires for rollback event

### Scenario 8: Continuous Learning Loop
1. Monitor production metrics for RL model
2. Collect remediation outcomes (success/failure)
3. Verify outcomes used for reward calculation
4. Trigger periodic RL training with new data
5. Verify training completes successfully
6. Verify model improved with new data
7. Verify continuous learning metrics tracked
8. Verify model versioning and rollback working

## 📊 Success Metrics

- **RL Training Success Rate**: > 95%
- **Remediation Success Rate Improvement**: > 5% (vs baseline)
- **Remediation Latency Reduction**: > 10% (vs baseline)
- **False Positive Rate Reduction**: > 20% (vs baseline)
- **RL Training Episode Reward**: Increasing trend over episodes
- **A/B Test Success Rate**: > 90% (RL model comparable or better)
- **Model Deployment Success Rate**: > 95%
- **Test Pass Rate**: 100%

---

**Last Updated**: January 08, 2026
**Owner**: SRE Team
**Status**: Validation Required