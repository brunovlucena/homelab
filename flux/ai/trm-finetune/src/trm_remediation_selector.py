#!/usr/bin/env python3
"""
🎯 TRM-Based Remediation Selector

Uses TRM model for recursive reasoning to select Lambda functions.
TRM does NOT support tool calling, so we use it to output structured JSON decisions.
"""

import os
import json
import re
from typing import Dict, Any, Optional
from pathlib import Path


class TRMRemediationSelector:
    """
    Selects Lambda functions using TRM recursive reasoning.
    
    TRM outputs structured text/JSON that we parse to get:
    - lambda_function: Name of Lambda function to call
    - parameters: Parameters for the Lambda function
    - reasoning: Why this function was selected
    """
    
    def __init__(self, model_path: str, trm_repo_path: str = None):
        self.model_path = Path(model_path)
        self.trm_repo_path = Path(trm_repo_path or os.getenv("TRM_REPO_PATH", "./TinyRecursiveModels"))
        self.model = None  # Will be loaded on first use
    
    def _load_model(self):
        """Lazy load TRM model."""
        if self.model is None:
            try:
                # Import TRM model
                import sys
                sys.path.insert(0, str(self.trm_repo_path))
                
                from models.recursive_reasoning.trm import TRMModel
                
                self.model = TRMModel.load_from_checkpoint(str(self.model_path))
                print(f"✅ Loaded TRM model from {self.model_path}")
            except Exception as e:
                print(f"⚠️  Could not load TRM model: {e}")
                print("   Falling back to rule-based selection")
                self.model = None
    
    def select_remediation(
        self,
        alert_data: Dict[str, Any],
        max_iterations: int = 10
    ) -> Dict[str, Any]:
        """
        Use TRM to recursively reason about remediation selection.
        
        Args:
            alert_data: Prometheus alert data
            max_iterations: Max recursive reasoning iterations
            
        Returns:
            Dict with lambda_function, parameters, reasoning, confidence
        """
        self._load_model()
        
        if self.model is None:
            # Fallback to rule-based
            return self._rule_based_selection(alert_data)
        
        # Create problem prompt
        problem = self._create_problem_prompt(alert_data)
        
        # Initial empty answer
        initial_answer = ""
        
        # Run TRM recursive reasoning
        try:
            result = self.model.reason(
                problem=problem,
                initial_answer=initial_answer,
                max_iterations=max_iterations
            )
            
            # Parse structured output from TRM
            parsed = self._parse_trm_output(result)
            
            return {
                "lambda_function": parsed.get("lambda_function"),
                "parameters": parsed.get("parameters", {}),
                "reasoning": parsed.get("reasoning", ""),
                "confidence": parsed.get("confidence", 0.7),
                "method": "trm_recursive_reasoning",
                "iterations": result.iterations if hasattr(result, 'iterations') else max_iterations
            }
        except Exception as e:
            print(f"⚠️  TRM reasoning failed: {e}")
            return self._rule_based_selection(alert_data)
    
    def _create_problem_prompt(self, alert_data: Dict[str, Any]) -> str:
        """Create problem prompt for TRM."""
        labels = alert_data.get("labels", {})
        annotations = alert_data.get("annotations", {})
        alertname = labels.get("alertname", "unknown")
        
        prompt = f"""Analyze this Prometheus alert and select the appropriate remediation Lambda function.

Alert Name: {alertname}
Labels: {json.dumps(labels, indent=2)}
Annotations: {json.dumps(annotations, indent=2)}

Available Lambda Functions:
1. flux-reconcile-kustomization: Reconcile Flux Kustomization
   Parameters: name (required), namespace (default: flux-system)
   Use when: Kustomization reconciliation failure

2. flux-reconcile-gitrepository: Reconcile Flux GitRepository
   Parameters: name (required), namespace (default: flux-system)
   Use when: GitRepository sync failure

3. flux-reconcile-helmrelease: Reconcile Flux HelmRelease
   Parameters: name (required), namespace (default: flux-system)
   Use when: HelmRelease reconciliation failure

4. pod-restart: Restart a pod or deployment
   Parameters: name (required), namespace (required), type (pod|deployment, default: pod)
   Use when: Pod is crashing, stuck, or needs restart

5. pod-check-status: Check pod status
   Parameters: name (required) OR selector (required), namespace (required)
   Use when: Need to verify pod health before remediation

6. scale-deployment: Scale deployment to specific replicas
   Parameters: name (required), namespace (required), replicas (required)
   Use when: Need to scale up/down a deployment

7. check-pvc-status: Check PVC status and usage
   Parameters: name (required), namespace (required)
   Use when: Storage issues suspected

Output your decision as JSON:
{{
  "lambda_function": "<function-name>",
  "parameters": {{
    "name": "<resource-name>",
    "namespace": "<namespace>",
    ...
  }},
  "reasoning": "<explanation>"
}}
"""
        return prompt
    
    def _parse_trm_output(self, result: Any) -> Dict[str, Any]:
        """Parse TRM output to extract structured decision."""
        # TRM returns text, we need to extract JSON from it
        output_text = str(result) if hasattr(result, '__str__') else result
        
        # Try to extract JSON from output
        json_match = re.search(r'\{[^{}]*"lambda_function"[^{}]*\}', output_text, re.DOTALL)
        if json_match:
            try:
                parsed = json.loads(json_match.group(0))
                return parsed
            except:
                pass
        
        # Try to find JSON in code blocks
        json_block = re.search(r'```json\s*(\{.*?\})\s*```', output_text, re.DOTALL)
        if json_block:
            try:
                parsed = json.loads(json_block.group(1))
                return parsed
            except:
                pass
        
        # Fallback: Try to extract function name from text
        function_match = re.search(r'lambda_function["\']?\s*[:=]\s*["\']?([^"\'\s]+)', output_text)
        if function_match:
            return {
                "lambda_function": function_match.group(1),
                "parameters": {},
                "reasoning": "Extracted from TRM output",
                "confidence": 0.5
            }
        
        # Last resort: return empty
        return {
            "lambda_function": None,
            "parameters": {},
            "reasoning": "Could not parse TRM output",
            "confidence": 0.0
        }
    
    def _rule_based_selection(self, alert_data: Dict[str, Any]) -> Dict[str, Any]:
        """Fallback rule-based selection based on runbook mappings."""
        labels = alert_data.get("labels", {})
        alertname = labels.get("alertname", "")
        
        # Flux alerts
        if "FluxReconciliationFailure" in alertname or (labels.get("kind") == "Kustomization"):
            return {
                "lambda_function": "flux-reconcile-kustomization",
                "parameters": {
                    "name": labels.get("name", ""),
                    "namespace": labels.get("namespace", "flux-system")
                },
                "reasoning": "Rule-based: Flux Kustomization failure",
                "confidence": 0.6,
                "method": "rule_based"
            }
        elif "FluxGitRepositoryOutOfSync" in alertname or "GitRepository" in alertname:
            return {
                "lambda_function": "flux-reconcile-gitrepository",
                "parameters": {
                    "name": labels.get("name", ""),
                    "namespace": labels.get("namespace", "flux-system")
                },
                "reasoning": "Rule-based: GitRepository out of sync",
                "confidence": 0.6,
                "method": "rule_based"
            }
        elif "FluxHelmReleaseFailing" in alertname or "HelmRelease" in alertname:
            return {
                "lambda_function": "flux-reconcile-helmrelease",
                "parameters": {
                    "name": labels.get("name", ""),
                    "namespace": labels.get("namespace", "flux-system")
                },
                "reasoning": "Rule-based: HelmRelease failing",
                "confidence": 0.6,
                "method": "rule_based"
            }
        # Pod alerts
        elif "PodCrashLoop" in alertname or "CrashLoopBackOff" in alertname:
            return {
                "lambda_function": "pod-restart",
                "parameters": {
                    "name": labels.get("pod", labels.get("name", "")),
                    "namespace": labels.get("namespace", ""),
                    "type": "pod"
                },
                "reasoning": "Rule-based: Pod crash loop",
                "confidence": 0.6,
                "method": "rule_based"
            }
        elif "ServiceDown" in alertname or "ServiceUnavailable" in alertname:
            # Service down - first check status
            namespace = labels.get("namespace", "")
            selector = labels.get("selector", "")
            if not selector and namespace:
                # Try to infer selector from namespace
                if namespace == "prometheus":
                    selector = "app.kubernetes.io/name=prometheus"
                elif namespace == "postgres":
                    selector = "app=postgres"
            
            return {
                "lambda_function": "pod-check-status",
                "parameters": {
                    "namespace": namespace,
                    "selector": selector or labels.get("selector", "")
                },
                "reasoning": "Rule-based: Service down, check pod status first",
                "confidence": 0.6,
                "method": "rule_based"
            }
        # Storage alerts
        elif "PersistentVolume" in alertname or "PVC" in alertname or labels.get("pvc"):
            return {
                "lambda_function": "check-pvc-status",
                "parameters": {
                    "name": labels.get("pvc", labels.get("name", "")),
                    "namespace": labels.get("namespace", "")
                },
                "reasoning": "Rule-based: PVC/storage issue",
                "confidence": 0.6,
                "method": "rule_based"
            }
        # Database alerts
        elif "Postgres" in alertname or "Database" in alertname:
            namespace = labels.get("namespace", "postgres")
            return {
                "lambda_function": "pod-check-status",
                "parameters": {
                    "namespace": namespace,
                    "selector": labels.get("selector", "app=postgres")
                },
                "reasoning": "Rule-based: Database issue, check status first",
                "confidence": 0.6,
                "method": "rule_based"
            }
        
        return {
            "lambda_function": None,
            "parameters": {},
            "reasoning": "No rule-based match found",
            "confidence": 0.0,
            "method": "rule_based"
        }


def main():
    """Test TRM remediation selector."""
    import sys
    
    if len(sys.argv) < 2:
        print("Usage: python trm_remediation_selector.py <alert-json>")
        sys.exit(1)
    
    alert_json = sys.argv[1]
    alert_data = json.loads(alert_json)
    
    model_path = os.getenv("TRM_MODEL_PATH", "./models/trm-finetuned/export")
    selector = TRMRemediationSelector(model_path)
    
    result = selector.select_remediation(alert_data)
    
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()

