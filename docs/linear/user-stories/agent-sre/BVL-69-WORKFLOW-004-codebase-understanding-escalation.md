# 🔍 WORKFLOW-004: Codebase Understanding & Escalation

**Linear URL**: https://linear.app/bvlucena/issue/BVL-231/backend-011-event-sequence-validation-and-ordering
**Linear URL**: https://linear.app/bvlucena/issue/BVL-203/workflow-004-codebase-understanding-and-escalation  

---

## 📋 User Story

**As an** SRE Engineer  
**I want** agent-sre to understand the codebase and escalate to humans when remediation is complex  
**So that** complex issues are handled appropriately with human-in-the-loop when needed


---


## 🎯 Acceptance Criteria

> **Note**: Features are already implemented. This ticket focuses on **validation** to ensure correctness, reliability, and production readiness.


- [ ] Agent-sre can analyze codebase for remediation context
- [ ] RAG-based codebase search implemented
- [ ] Code understanding for remediation selection
- [ ] Human escalation when confidence is low
- [ ] Escalation includes full context and analysis
- [ ] Human feedback loop for learning
- [ ] Escalation metrics tracked

---

## 🔄 Complete Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│        CODEBASE UNDERSTANDING & ESCALATION WORKFLOW                   │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ⏱️  t=0s: COMPLEX ALERT RECEIVED                                    │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Alert: CustomApplicationError                       │            │
│  │  - No standard remediation available                 │            │
│  │  - Requires codebase understanding                   │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=1s: SEARCH CODEBASE                                           │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE searches codebase:                        │            │
│  │  - RAG search for error patterns                      │            │
│  │  - Find related code files                            │            │
│  │  - Understand application logic                        │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=3s: ANALYZE AND DECIDE                                        │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE analyzes:                                  │            │
│  │  - Code complexity: High                              │            │
│  │  - Remediation confidence: 0.35 (low)                  │            │
│  │  - Risk: High (production impact)                      │            │
│  │                                                      │            │
│  │  → Decision: Escalate to human                        │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=4s: ESCALATE TO HUMAN                                         │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE creates escalation:                        │            │
│  │  - Linear issue updated with "needs-human-review"     │            │
│  │  - Full context included                              │            │
│  │  - Code analysis summary                              │            │
│  │  - Suggested remediation (low confidence)             │            │
│  │  - On-call engineer notified                          │            │
│  └──────────────────────────────────────────────────────┘            │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Implementation Details

### Codebase Understanding Service

```python
# src/sre_agent/codebase_understanding.py
from typing import Dict, Any, List
from sre_agent.rag_client import RAGClient

class CodebaseUnderstanding:
    """Understand codebase for remediation context."""
    
    def __init__(self):
        self.rag_client = RAGClient()
    
    async def understand_and_escalate(
        self,
        alert_data: Dict[str, Any],
        issue_id: str
    ) -> Dict[str, Any]:
        """Understand codebase and escalate if needed."""
        # Search codebase
        code_context = await self._search_codebase(alert_data)
        
        # Analyze complexity
        complexity = self._analyze_complexity(code_context)
        
        # Calculate confidence
        confidence = self._calculate_confidence(code_context, alert_data)
        
        # Decide on escalation
        if confidence < 0.5 or complexity > 0.7:
            return await self._escalate_to_human(
                issue_id,
                alert_data,
                code_context,
                confidence
            )
        
        return {
            "escalated": False,
            "confidence": confidence,
            "remediation": self._suggest_remediation(code_context)
        }
    
    async def _search_codebase(
        self,
        alert_data: Dict[str, Any]
    ) -> List[Dict[str, Any]]:
        """Search codebase using RAG."""
        query = f"Error: {alert_data.get('alertname')} remediation"
        results = await self.rag_client.search(query, limit=10)
        return results
```

---

## 📚 References

- [RAG Implementation](../../docs/rag-implementation.md)
- [Human-in-the-Loop Documentation](../../docs/human-escalation.md)

---

## ✅ Definition of Done

- [ ] Codebase search implemented
- [ ] Complexity analysis working
- [ ] Escalation logic operational
- [ ] Human feedback loop implemented
- [ ] Documentation updated

---

**Related Stories**:
- [WORKFLOW-003: Enriched Issue Updates](./BVL-67-WORKFLOW-003-enriched-issue-updates.md)
- [AI-003: TinyRecursiveModels Integration](./BVL-63-AI-003-tiny-recursive-models.md)


## 🧪 Test Scenarios

### Scenario 1: Low Complexity Alert (No Escalation)
1. Receive alert with simple remediation pattern
2. Verify codebase search finds relevant code
3. Verify complexity analysis shows low complexity (< 0.5)
4. Verify confidence calculated > 0.7
5. Verify no escalation (agent handles autonomously)
6. Verify remediation executed successfully
7. Verify escalation metrics updated (no escalation)

### Scenario 2: High Complexity Alert (Escalation Required)
1. Receive alert with complex remediation pattern
2. Verify codebase search finds multiple related code files
3. Verify complexity analysis shows high complexity (> 0.7)
4. Verify confidence calculated < 0.5
5. Verify escalation triggered (human review required)
6. Verify Linear issue updated with escalation context
7. Verify on-call engineer notified
8. Verify escalation includes full analysis and suggested remediation

### Scenario 3: Codebase Search and Understanding
1. Receive alert requiring codebase understanding
2. Trigger codebase search using RAG
3. Verify relevant code files found (top 10 results)
4. Verify code context extracted correctly
5. Verify code understanding generates summary
6. Verify summary includes remediation suggestions
7. Verify code analysis logged for audit

### Scenario 4: Escalation Context and Analysis
1. Trigger escalation for complex alert
2. Verify escalation includes alert context (full alert data)
3. Verify escalation includes codebase analysis summary
4. Verify escalation includes suggested remediation (low confidence)
5. Verify escalation includes risk assessment
6. Verify escalation includes relevant code snippets
7. Verify escalation formatted clearly for human review

### Scenario 5: Human Feedback Loop
1. Escalate issue to human engineer
2. Engineer reviews and provides feedback
3. Verify feedback collected and stored
4. Verify feedback used for learning (few-shot examples)
5. Verify feedback improves future remediation selection
6. Verify feedback metrics tracked (feedback rate, improvement)
7. Verify feedback loop continuous improvement

### Scenario 6: Escalation Performance
1. Trigger escalation for complex alert
2. Verify escalation created within 5 seconds
3. Verify notification delivered to on-call engineer
4. Verify escalation workflow completes within 10 seconds total
5. Verify no timeout issues
6. Verify metrics recorded for escalation duration
7. Verify escalation works under high load

### Scenario 7: Escalation Failure Handling
1. Simulate escalation service unavailable
2. Verify failure handled gracefully
3. Verify fallback behavior works (create Linear issue manually)
4. Verify error logged with context
5. Verify retry logic works when service recovers
6. Verify alerts fire for repeated failures

## 📊 Success Metrics

- **Escalation Detection Time**: < 5 seconds (P95)
- **Escalation Creation Time**: < 10 seconds (P95)
- **Codebase Search Performance**: < 3 seconds (P95)
- **Escalation Accuracy**: > 90% (correctly identifies complex issues)
- **False Escalation Rate**: < 10% (unnecessary escalations)
- **Human Feedback Rate**: > 80% (engineers provide feedback)
- **Test Pass Rate**: 100%

---

**Last Updated**: January 08, 2026
**Owner**: SRE Team
**Status**: Validation Required