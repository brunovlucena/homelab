#!/usr/bin/env python3
"""
🧪 Test TRM Model on Agent-SRE Runbook

Tests if TRM can learn to follow the runbook and select appropriate Lambda functions.
"""

import os
import json
from pathlib import Path
from typing import Dict, Any, List
from dataclasses import dataclass, asdict


@dataclass
class RunbookTestExample:
    """Test example from runbook."""
    alert_name: str
    alert_labels: Dict[str, str]
    alert_annotations: Dict[str, str]
    expected_lambda: str
    expected_parameters: Dict[str, Any]
    runbook_section: str
    reasoning: str


class RunbookDatasetGenerator:
    """Generates training/test data from agent-sre runbook."""
    
    def __init__(self, runbook_path: str):
        self.runbook_path = Path(runbook_path)
        self.examples: List[RunbookTestExample] = []
    
    def parse_runbook(self) -> List[RunbookTestExample]:
        """Parse runbook and extract alert → Lambda function mappings."""
        runbook_content = self.runbook_path.read_text()
        
        examples = []
        
        # Parse Flux reconciliation alerts
        examples.extend(self._parse_flux_alerts(runbook_content))
        
        # Parse pod/service alerts
        examples.extend(self._parse_pod_alerts(runbook_content))
        
        # Parse storage alerts
        examples.extend(self._parse_storage_alerts(runbook_content))
        
        # Parse database alerts
        examples.extend(self._parse_database_alerts(runbook_content))
        
        return examples
    
    def _parse_flux_alerts(self, content: str) -> List[RunbookTestExample]:
        """Parse Flux-related alerts."""
        examples = []
        
        # FluxReconciliationFailure → flux-reconcile-kustomization
        examples.append(RunbookTestExample(
            alert_name="FluxReconciliationFailure",
            alert_labels={
                "alertname": "FluxReconciliationFailure",
                "name": "homepage",
                "namespace": "flux-system",
                "kind": "Kustomization"
            },
            alert_annotations={},
            expected_lambda="flux-reconcile-kustomization",
            expected_parameters={"name": "homepage", "namespace": "flux-system"},
            runbook_section="Flux CD (GitOps)",
            reasoning="Alert indicates Kustomization reconciliation failure, so reconcile it."
        ))
        
        # FluxGitRepositoryOutOfSync → flux-reconcile-gitrepository
        examples.append(RunbookTestExample(
            alert_name="FluxGitRepositoryOutOfSync",
            alert_labels={
                "alertname": "FluxGitRepositoryOutOfSync",
                "name": "homelab",
                "namespace": "flux-system"
            },
            alert_annotations={},
            expected_lambda="flux-reconcile-gitrepository",
            expected_parameters={"name": "homelab", "namespace": "flux-system"},
            runbook_section="Agent-SRE Flux Reconciliation Triggers",
            reasoning="GitRepository is out of sync, trigger reconciliation."
        ))
        
        # FluxHelmReleaseFailing → flux-reconcile-helmrelease
        examples.append(RunbookTestExample(
            alert_name="FluxHelmReleaseFailing",
            alert_labels={
                "alertname": "FluxHelmReleaseFailing",
                "name": "prometheus",
                "namespace": "prometheus"
            },
            alert_annotations={},
            expected_lambda="flux-reconcile-helmrelease",
            expected_parameters={"name": "prometheus", "namespace": "prometheus"},
            runbook_section="Agent-SRE Flux Reconciliation Triggers",
            reasoning="HelmRelease is failing, trigger reconciliation."
        ))
        
        return examples
    
    def _parse_pod_alerts(self, content: str) -> List[RunbookTestExample]:
        """Parse pod/service alerts."""
        examples = []
        
        # PodCrashLoopBackOff → pod-restart
        examples.append(RunbookTestExample(
            alert_name="PodCrashLoopBackOff",
            alert_labels={
                "alertname": "PodCrashLoopBackOff",
                "pod": "agent-sre-abc123",
                "namespace": "agent-sre"
            },
            alert_annotations={},
            expected_lambda="pod-restart",
            expected_parameters={"name": "agent-sre-abc123", "namespace": "agent-sre", "type": "pod"},
            runbook_section="Kubernetes Infrastructure",
            reasoning="Pod is in crash loop, restarting it may resolve transient issues."
        ))
        
        # PrometheusServiceDown → pod-check-status then pod-restart
        examples.append(RunbookTestExample(
            alert_name="PrometheusServiceDown",
            alert_labels={
                "alertname": "PrometheusServiceDown",
                "namespace": "prometheus"
            },
            alert_annotations={},
            expected_lambda="pod-check-status",
            expected_parameters={"namespace": "prometheus", "selector": "app.kubernetes.io/name=prometheus"},
            runbook_section="Prometheus & Alertmanager",
            reasoning="Service is down, first check pod status to understand the issue."
        ))
        
        return examples
    
    def _parse_storage_alerts(self, content: str) -> List[RunbookTestExample]:
        """Parse storage alerts."""
        examples = []
        
        # PersistentVolumeFillingUpCritical → check-pvc-status
        examples.append(RunbookTestExample(
            alert_name="PersistentVolumeFillingUpCritical",
            alert_labels={
                "alertname": "PersistentVolumeFillingUpCritical",
                "namespace": "prometheus",
                "pvc": "prometheus-storage"
            },
            alert_annotations={},
            expected_lambda="check-pvc-status",
            expected_parameters={"name": "prometheus-storage", "namespace": "prometheus"},
            runbook_section="Storage & Persistent Volumes",
            reasoning="PVC is filling up, check status to understand usage patterns."
        ))
        
        return examples
    
    def _parse_database_alerts(self, content: str) -> List[RunbookTestExample]:
        """Parse database alerts."""
        examples = []
        
        # PostgresHighConnectionCount → scale-deployment (if needed) or check status
        examples.append(RunbookTestExample(
            alert_name="PostgresHighConnectionCount",
            alert_labels={
                "alertname": "PostgresHighConnectionCount",
                "namespace": "postgres"
            },
            alert_annotations={},
            expected_lambda="pod-check-status",
            expected_parameters={"namespace": "postgres", "selector": "app=postgres"},
            runbook_section="PostgreSQL (Database)",
            reasoning="High connection count, first check pod status to understand the issue."
        ))
        
        return examples
    
    def format_for_trm(self, examples: List[RunbookTestExample]) -> List[Dict[str, Any]]:
        """Format examples for TRM training."""
        trm_examples = []
        
        for ex in examples:
            # Create problem: Alert analysis
            problem = f"""Analyze this Prometheus alert and select the appropriate remediation Lambda function.

Alert Name: {ex.alert_name}
Labels: {json.dumps(ex.alert_labels, indent=2)}
Annotations: {json.dumps(ex.alert_annotations, indent=2)}

Runbook Section: {ex.runbook_section}

Available Lambda Functions:
- flux-reconcile-kustomization: Reconcile Flux Kustomization (params: name, namespace)
- flux-reconcile-gitrepository: Reconcile Flux GitRepository (params: name, namespace)
- flux-reconcile-helmrelease: Reconcile Flux HelmRelease (params: name, namespace)
- pod-restart: Restart a pod or deployment (params: name, namespace, type)
- pod-check-status: Check pod status (params: name/selector, namespace)
- scale-deployment: Scale deployment (params: name, namespace, replicas)
- check-pvc-status: Check PVC status (params: name, namespace)
"""
            
            # Initial answer (empty - model will reason)
            initial_answer = ""
            
            # Solution: Structured JSON output
            solution = json.dumps({
                "lambda_function": ex.expected_lambda,
                "parameters": ex.expected_parameters,
                "reasoning": ex.reasoning
            }, indent=2)
            
            # Recursive reasoning steps
            reasoning_steps = [
                f"Step 1: Identify alert type: {ex.alert_name}",
                f"Step 2: Analyze alert labels and context",
                f"Step 3: Match to runbook section: {ex.runbook_section}",
                f"Step 4: Select appropriate Lambda function: {ex.expected_lambda}",
                f"Step 5: Extract parameters from labels: {json.dumps(ex.expected_parameters)}",
                f"Step 6: Generate reasoning: {ex.reasoning}"
            ]
            
            trm_examples.append({
                "problem": problem,
                "initial_answer": initial_answer,
                "solution": solution,
                "reasoning_steps": reasoning_steps,
                "metadata": {
                    "source": "runbook",
                    "alert_name": ex.alert_name,
                    "runbook_section": ex.runbook_section,
                    "expected_lambda": ex.expected_lambda
                }
            })
        
        return trm_examples
    
    def save_test_dataset(self, examples: List[Dict[str, Any]], output_path: str):
        """Save test dataset to JSONL."""
        output_path = Path(output_path)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        
        with open(output_path, 'w') as f:
            for ex in examples:
                f.write(json.dumps(ex) + '\n')
        
        print(f"💾 Saved {len(examples)} test examples to {output_path}")


def main():
    """Generate test dataset from runbook."""
    # Default to runbook in sibling agent-sre directory
    script_dir = Path(__file__).parent
    default_runbook_path = script_dir.parent.parent / "agent-sre" / "docs" / "RUNBOOK.md"
    
    runbook_path = os.getenv(
        "RUNBOOK_PATH",
        str(default_runbook_path)
    )
    output_path = os.getenv(
        "OUTPUT_PATH",
        "./data/runbook_test_dataset.jsonl"
    )
    
    generator = RunbookDatasetGenerator(runbook_path)
    
    print("📚 Parsing runbook...")
    examples = generator.parse_runbook()
    print(f"✅ Extracted {len(examples)} examples from runbook")
    
    print("📊 Formatting for TRM...")
    trm_examples = generator.format_for_trm(examples)
    print(f"✅ Formatted {len(trm_examples)} TRM examples")
    
    print("💾 Saving dataset...")
    generator.save_test_dataset(trm_examples, output_path)
    
    print(f"\n🎉 Test dataset ready!")
    print(f"   Total examples: {len(trm_examples)}")
    print(f"   Output: {output_path}")
    print(f"\n📋 Example alerts covered:")
    for ex in examples[:5]:
        print(f"   - {ex.alert_name} → {ex.expected_lambda}")


if __name__ == "__main__":
    main()

