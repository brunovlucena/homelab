# 🔍 WORKFLOW-002: Lambda Function Annotation Discovery

**Linear URL**: https://linear.app/bvlucena/issue/BVL-228/backend-008-error-handling-and-logging
**Linear URL**: https://linear.app/bvlucena/issue/BVL-199/workflow-002-lambda-function-annotation-discovery  

---

## 📋 User Story

**As an** SRE Engineer  
**I want** agent-sre to automatically discover LambdaFunction annotations from PrometheusRules  
**So that** remediation actions are automatically mapped to alerts without manual configuration


---


## 🎯 Acceptance Criteria

> **Note**: Features are already implemented. This ticket focuses on **validation** to ensure correctness, reliability, and production readiness.


- [ ] Agent-sre scans PrometheusRules for lambda_function annotations
- [ ] Discovers LambdaFunction resources in Kubernetes
- [ ] Validates LambdaFunction exists and is accessible
- [ ] Creates mapping between alerts and LambdaFunctions
- [ ] Updates PrometheusRules with missing annotations
- [ ] Logs discovered mappings for audit
- [ ] Handles missing or invalid LambdaFunctions gracefully
- [ ] Periodic re-scan for new PrometheusRules

---

## 🔄 Complete Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│        LAMBDA FUNCTION ANNOTATION DISCOVERY WORKFLOW                   │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ⏱️  t=0s: PERIODIC SCAN TRIGGERED                                    │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE scheduled job runs every 5 minutes         │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=1s: SCAN PROMETHEUSRULES                                      │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Query Kubernetes for PrometheusRules:               │            │
│  │  kubectl get prometheusrules -A                      │            │
│  │                                                      │            │
│  │  Found: 25 PrometheusRules                            │            │
│  │  - 15 with lambda_function annotation                │            │
│  │  - 10 without annotation                              │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=2s: DISCOVER LAMBDAFUNCTIONS                                   │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Query Kubernetes for LambdaFunctions:               │            │
│  │  kubectl get lambdafunctions -A                      │            │
│  │                                                      │            │
│  │  Found: 20 LambdaFunctions                           │            │
│  │  - scale-pod                                         │            │
│  │  - check-pvc-status                                   │            │
│  │  - flux-reconcile-kustomization                       │            │
│  │  - ...                                                │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=3s: MATCH ALERTS TO LAMBDAFUNCTIONS                           │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  For each PrometheusRule without annotation:         │            │
│  │                                                      │            │
│  │  1. Extract alertname                                │            │
│  │  2. Match to LambdaFunction by name pattern          │            │
│  │  3. Validate LambdaFunction exists                   │            │
│  │  4. Add lambda_function annotation                    │            │
│  │                                                      │            │
│  │  Example:                                             │            │
│  │  Alert: PodCPUHigh                                    │            │
│  │  → Match: scale-pod LambdaFunction                    │            │
│  │  → Add annotation: lambda_function: "scale-pod"       │            │
│  └──────────────────────────────────────────────────────┘            │
│                           ↓                                          │
│  ⏱️  t=4s: UPDATE PROMETHEUSRULES                                     │
│  ┌──────────────────────────────────────────────────────┐            │
│  │  Agent-SRE updates PrometheusRules:                  │            │
│  │  - Adds lambda_function annotation                    │            │
│  │  - Adds lambda_parameters if needed                   │            │
│  │  - Commits changes via GitOps (Flux)                  │            │
│  └──────────────────────────────────────────────────────┘            │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Implementation Details

### Annotation Discovery Service

```python
# src/sre_agent/annotation_discovery.py
from typing import List, Dict, Any
from kubernetes import client

class AnnotationDiscovery:
    """Discover and add LambdaFunction annotations to PrometheusRules."""
    
    def __init__(self, k8s_client):
        self.k8s_client = k8s_client
        self.custom_api = client.CustomObjectsApi()
    
    async def discover_annotations(self):
        """Discover and add missing annotations."""
        # Get all PrometheusRules
        prometheus_rules = await self._get_prometheus_rules()
        
        # Get all LambdaFunctions
        lambda_functions = await self._get_lambda_functions()
        
        # Match and update
        for rule in prometheus_rules:
            if not self._has_lambda_annotation(rule):
                matched_lambda = self._match_lambda_function(rule, lambda_functions)
                if matched_lambda:
                    await self._add_annotation(rule, matched_lambda)
    
    def _match_lambda_function(
        self,
        rule: Dict[str, Any],
        lambda_functions: List[Dict[str, Any]]
    ) -> Optional[Dict[str, Any]]:
        """Match PrometheusRule to LambdaFunction."""
        alertname = self._extract_alertname(rule)
        
        # Pattern matching
        patterns = {
            "PodCPUHigh": "scale-pod",
            "PodMemoryHigh": "scale-pod",
            "PersistentVolumeFillingUp": "check-pvc-status",
            "FluxReconciliationFailure": "flux-reconcile-kustomization",
        }
        
        lambda_name = patterns.get(alertname)
        if lambda_name:
            return next(
                (lf for lf in lambda_functions if lf["metadata"]["name"] == lambda_name),
                None
            )
        
        return None
```

---

## 📚 References

- [PrometheusRule CRD](https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.PrometheusRule)
- [LambdaFunction CRD](../../docs/knative/03-for-engineers/backend/README.md)

---

## ✅ Definition of Done

- [ ] PrometheusRule scanning implemented
- [ ] LambdaFunction discovery working
- [ ] Annotation matching logic implemented
- [ ] PrometheusRule updates working
- [ ] Periodic scanning operational
- [ ] Documentation updated

---

**Related Stories**:
- [WORKFLOW-001: PrometheusRule → Linear Issue](./BVL-65-WORKFLOW-001-prometheus-to-linear-with-slm.md)



---

**Last Updated**: January 08, 2026
**Owner**: SRE Team
**Status**: Validation Required