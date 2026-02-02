#!/usr/bin/env python3
"""
✅ Validate TRM Remediation Selector Against Test Dataset

Tests the selector against all examples in the test dataset.
"""

import json
import sys
from pathlib import Path
from typing import Dict, Any, List
from trm_remediation_selector import TRMRemediationSelector


def load_test_dataset(dataset_path: str) -> List[Dict[str, Any]]:
    """Load test dataset from JSONL file."""
    examples = []
    with open(dataset_path, 'r') as f:
        for line in f:
            if line.strip():
                examples.append(json.loads(line))
    return examples


def extract_alert_from_example(example: Dict[str, Any]) -> Dict[str, Any]:
    """Extract alert data from TRM example format."""
    problem = example.get("problem", "")
    
    # Parse alert name from problem
    alert_name = None
    labels = {}
    annotations = {}
    
    # Extract alert name
    if "Alert Name:" in problem:
        alert_name = problem.split("Alert Name:")[1].split("\n")[0].strip()
    
    # Extract labels
    if "Labels:" in problem:
        labels_section = problem.split("Labels:")[1].split("Annotations:")[0].strip()
        try:
            labels = json.loads(labels_section)
        except:
            # Fallback: try to parse manually
            pass
    
    # Extract annotations
    if "Annotations:" in problem:
        annotations_section = problem.split("Annotations:")[1].split("Runbook Section:")[0].strip()
        try:
            annotations = json.loads(annotations_section)
        except:
            pass
    
    return {
        "labels": labels,
        "annotations": annotations
    }


def validate_selector(dataset_path: str):
    """Validate selector against test dataset."""
    print("📚 Loading test dataset...")
    examples = load_test_dataset(dataset_path)
    print(f"✅ Loaded {len(examples)} test examples\n")
    
    # Initialize selector
    model_path = "./models/trm-finetuned/export"  # Will fallback to rule-based
    selector = TRMRemediationSelector(model_path)
    
    # Track results
    results = {
        "total": len(examples),
        "correct": 0,
        "incorrect": 0,
        "details": []
    }
    
    print("🧪 Testing selector against test dataset...\n")
    
    for i, example in enumerate(examples, 1):
        # Get expected result
        solution = json.loads(example.get("solution", "{}"))
        expected_lambda = solution.get("lambda_function")
        expected_params = solution.get("parameters", {})
        expected_reasoning = solution.get("reasoning", "")
        
        # Extract alert data
        alert_data = extract_alert_from_example(example)
        
        # Get actual result
        actual = selector.select_remediation(alert_data)
        actual_lambda = actual.get("lambda_function")
        actual_params = actual.get("parameters", {})
        
        # Compare
        lambda_match = actual_lambda == expected_lambda
        params_match = actual_params == expected_params
        
        is_correct = lambda_match and params_match
        
        if is_correct:
            results["correct"] += 1
            status = "✅"
        else:
            results["incorrect"] += 1
            status = "❌"
        
        # Store details
        result_detail = {
            "example": i,
            "alert_name": alert_data.get("labels", {}).get("alertname", "unknown"),
            "expected": {
                "lambda": expected_lambda,
                "parameters": expected_params
            },
            "actual": {
                "lambda": actual_lambda,
                "parameters": actual_params,
                "method": actual.get("method", "unknown")
            },
            "correct": is_correct
        }
        results["details"].append(result_detail)
        
        # Print result
        print(f"{status} Example {i}: {alert_data.get('labels', {}).get('alertname', 'unknown')}")
        if not lambda_match:
            print(f"   Expected lambda: {expected_lambda}")
            print(f"   Actual lambda:   {actual_lambda}")
        if not params_match:
            print(f"   Expected params: {expected_params}")
            print(f"   Actual params:   {actual_params}")
        print()
    
    # Print summary
    print("=" * 60)
    print("📊 Validation Summary")
    print("=" * 60)
    print(f"Total examples:  {results['total']}")
    print(f"✅ Correct:      {results['correct']} ({100 * results['correct'] / results['total']:.1f}%)")
    print(f"❌ Incorrect:    {results['incorrect']} ({100 * results['incorrect'] / results['total']:.1f}%)")
    print()
    
    # Show incorrect examples
    if results['incorrect'] > 0:
        print("❌ Incorrect Examples:")
        for detail in results['details']:
            if not detail['correct']:
                print(f"  - {detail['alert_name']}")
                print(f"    Expected: {detail['expected']['lambda']}")
                print(f"    Actual:   {detail['actual']['lambda']} ({detail['actual']['method']})")
        print()
    
    return results


def main():
    """Main entry point."""
    dataset_path = sys.argv[1] if len(sys.argv) > 1 else "./data/runbook_test_dataset.jsonl"
    
    if not Path(dataset_path).exists():
        print(f"❌ Dataset not found: {dataset_path}")
        print("   Run: python src/test_trm_runbook.py first")
        sys.exit(1)
    
    results = validate_selector(dataset_path)
    
    # Exit with error if any failures
    if results['incorrect'] > 0:
        sys.exit(1)
    else:
        print("🎉 All tests passed!")


if __name__ == "__main__":
    main()
