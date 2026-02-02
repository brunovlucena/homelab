# 🔀 WORKFLOW-005: PR Generation & Automated Merging

**Linear URL**: https://linear.app/bvlucena/issue/BVL-232/backend-012-schema-validation-and-registry
**Linear URL**: https://linear.app/bvlucena/issue/BVL-202/workflow-005-pr-generation-and-automated-merging  

---

## 📋 User Story

**As an** SRE Engineer  
**I want** agent-sre to generate pull requests for infrastructure changes and automatically merge them when safe  
**So that** remediation actions can be applied as code changes with proper review and automation


---


## 🎯 Acceptance Criteria

> **Note**: Features are already implemented. This ticket focuses on **validation** to ensure correctness, reliability, and production readiness.


- [ ] Agent-sre can generate PRs for infrastructure changes
- [ ] PRs include proper descriptions and context
- [ ] Automated merging for low-risk changes
- [ ] Human review required for high-risk changes
- [ ] PR templates and standards followed
- [ ] CI/CD validation before merging
- [ ] Rollback capability if issues detected
- [ ] PR metrics tracked (generation rate, merge rate, rollback rate)

---

## 🔄 Complete Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│          PR GENERATION & AUTOMATED MERGING WORKFLOW                    │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ⏱️  t=0s: REMEDIATION REQUIRES INFRASTRUCTURE CHANGE                │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Alert: PersistentVolumeFillingUp                    │            │
│  │  Remediation: Increase PVC size                       │            │
│  │  Change: Update PVC spec in GitOps repo               │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=1s: GENERATE PR                                                │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE:                                           │            │
│  │  1. Creates branch: agent-sre/pvc-increase-{id}      │            │
│  │  2. Updates PVC YAML                                  │            │
│  │  3. Commits changes                                   │            │
│  │  4. Creates PR with description                       │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=2s: ASSESS RISK                                                │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE assesses:                                  │            │
│  │  - Change type: PVC size increase                     │            │
│  │  - Risk level: Low                                    │            │
│  │  - Impact: Single namespace                           │            │
│  │  - Rollback: Easy                                     │            │
│  │                                                      │            │
│  │  → Decision: Auto-merge after CI passes               │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=5s: CI/CD VALIDATION                                           │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  CI/CD pipeline runs:                                 │            │
│  │  - YAML validation                                    │            │
│  │  - Dry-run apply                                      │            │
│  │  - Security scan                                      │            │
│  │                                                      │            │
│  │  → All checks pass                                    │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=6s: AUTO-MERGE                                                │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE merges PR:                                 │            │
│  │  - Squash merge                                       │            │
│  │  - Updates Linear issue                               │            │
│  │  - Flux applies changes                               │            │
│  └──────────────────────────────────────────────────────┘            │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Implementation Details

### PR Generation Service

```python
# src/sre_agent/pr_generation.py
from typing import Dict, Any, Optional
from github import Github

class PRGenerator:
    """Generate and manage pull requests for infrastructure changes."""
    
    def __init__(self, github_token: str, repo_name: str):
        self.github = Github(github_token)
        self.repo = self.github.get_repo(repo_name)
    
    async def generate_and_merge_pr(
        self,
        change: Dict[str, Any],
        issue_id: str
    ) -> Dict[str, Any]:
        """Generate PR and merge if safe."""
        # Create branch
        branch_name = f"agent-sre/{change['type']}-{issue_id}"
        branch = await self._create_branch(branch_name)
        
        # Make changes
        await self._apply_changes(branch, change)
        
        # Create PR
        pr = await self._create_pr(branch, change, issue_id)
        
        # Assess risk
        risk_level = self._assess_risk(change)
        
        # Auto-merge if low risk
        if risk_level == "low":
            await self._wait_for_ci(pr)
            await self._merge_pr(pr)
            return {"merged": True, "pr_url": pr.html_url}
        else:
            return {"merged": False, "pr_url": pr.html_url, "requires_review": True}
    
    def _assess_risk(self, change: Dict[str, Any]) -> str:
        """Assess risk level of change."""
        change_type = change.get("type")
        
        low_risk = ["pvc-size", "resource-limits", "replica-count"]
        medium_risk = ["network-policy", "service-account"]
        high_risk = ["cluster-config", "security-policy"]
        
        if change_type in low_risk:
            return "low"
        elif change_type in medium_risk:
            return "medium"
        elif change_type in high_risk:
            return "high"
        else:
            return "medium"  # Default to medium
```

---

## 📚 References

- [GitOps Best Practices](../../docs/gitops-best-practices.md)
- [PR Automation Documentation](../../docs/pr-automation.md)

---

## ✅ Definition of Done

- [ ] PR generation implemented
- [ ] Risk assessment working
- [ ] Auto-merge logic operational
- [ ] CI/CD integration working
- [ ] Documentation updated

---

**Related Stories**:
- [WORKFLOW-004: Codebase Understanding](./BVL-69-WORKFLOW-004-codebase-understanding-escalation.md)
- [SRE-001: Build Failure Investigation](./BVL-45-SRE-001-build-failure-investigation.md)


## 🧪 Test Scenarios

### Scenario 1: Low-Risk PR Generation and Auto-Merge
1. Remediation requires low-risk change (PVC size increase)
2. Verify PR branch created successfully
3. Verify changes committed correctly
4. Verify PR created with proper description and context
5. Verify risk assessment determines low risk
6. Verify CI/CD pipeline runs and passes
7. Verify PR auto-merged after CI passes
8. Verify Flux applies changes automatically
9. Verify Linear issue updated with PR link and merge status

### Scenario 2: Medium-Risk PR Generation (Requires Review)
1. Remediation requires medium-risk change (network policy update)
2. Verify PR created successfully
3. Verify risk assessment determines medium risk
4. Verify PR marked as requiring review (not auto-merged)
5. Verify reviewers notified
6. Verify PR reviewed by human engineer
7. Verify PR merged manually after approval
8. Verify Linear issue updated with review status

### Scenario 3: High-Risk PR Generation (Requires Approval)
1. Remediation requires high-risk change (cluster config update)
2. Verify PR created successfully
3. Verify risk assessment determines high risk
4. Verify PR marked as requiring approval (not auto-merged)
5. Verify approval required from senior engineer
6. Verify PR blocked until approval received
7. Verify PR merged after approval
8. Verify Linear issue updated with approval status

### Scenario 4: PR CI/CD Validation
1. Generate PR for infrastructure change
2. Verify CI/CD pipeline runs automatically
3. Verify YAML validation passes
4. Verify dry-run apply succeeds
5. Verify security scan passes
6. Verify all CI checks pass
7. Verify PR ready for merge after CI passes
8. Verify CI failures block auto-merge

### Scenario 5: PR Rollback Capability
1. Generate and merge PR with infrastructure change
2. Verify change applied by Flux
3. Verify change causes issues (simulated)
4. Trigger rollback process
5. Verify rollback PR created automatically
6. Verify rollback PR auto-merged
7. Verify rollback applied successfully
8. Verify system restored to previous state
9. Verify rollback metrics recorded

### Scenario 6: PR Templates and Standards
1. Generate PR for infrastructure change
2. Verify PR follows template (description format)
3. Verify PR includes required context (alert, remediation, impact)
4. Verify PR includes links to Linear issue
5. Verify PR includes test results
6. Verify PR follows code standards
7. Verify PR description clear and actionable

### Scenario 7: PR Performance and Scalability
1. Generate 10+ PRs simultaneously
2. Verify all PRs created successfully
3. Verify CI/CD handles concurrent PRs
4. Verify no race conditions in Git operations
5. Verify PR creation performance acceptable (< 30 seconds per PR)
6. Verify merge performance acceptable (< 60 seconds per PR)
7. Verify system handles load without degradation

### Scenario 8: PR Failure Handling
1. Simulate Git repository unavailable
2. Verify PR generation failure handled gracefully
3. Verify error logged with context
4. Verify retry logic works when repository recovers
5. Verify alerts fire for repeated failures
6. Verify fallback behavior works (manual PR creation)
7. Verify Linear issue updated with failure status

## 📊 Success Metrics

- **PR Generation Success Rate**: > 95%
- **PR Generation Time**: < 30 seconds (P95)
- **CI/CD Validation Time**: < 5 minutes (P95)
- **PR Merge Time**: < 60 seconds after CI passes (P95)
- **Auto-Merge Success Rate**: > 90% for low-risk PRs
- **Rollback Success Rate**: > 95%
- **PR Template Compliance**: 100%
- **Test Pass Rate**: 100%

---

**Last Updated**: January 08, 2026
**Owner**: SRE Team
**Status**: Validation Required