#!/usr/bin/env python3
"""
📊 Prepare Training Data for TRM Fine-Tuning

Validates and prepares the runbook dataset for TRM training.
"""

import json
import sys
from pathlib import Path
from typing import List, Dict, Any
from dataclasses import dataclass


@dataclass
class TrainingStats:
    """Training dataset statistics."""
    total_examples: int
    alert_types: Dict[str, int]
    lambda_functions: Dict[str, int]
    avg_problem_length: float
    avg_solution_length: float
    reasoning_steps_count: Dict[int, int]


def load_dataset(dataset_path: str) -> List[Dict[str, Any]]:
    """Load dataset from JSONL file."""
    examples = []
    with open(dataset_path, 'r') as f:
        for line in f:
            if line.strip():
                examples.append(json.loads(line))
    return examples


def analyze_dataset(examples: List[Dict[str, Any]]) -> TrainingStats:
    """Analyze dataset and compute statistics."""
    alert_types = {}
    lambda_functions = {}
    problem_lengths = []
    solution_lengths = []
    reasoning_steps_counts = {}
    
    for ex in examples:
        # Extract alert name from metadata
        metadata = ex.get("metadata", {})
        alert_name = metadata.get("alert_name", "unknown")
        alert_types[alert_name] = alert_types.get(alert_name, 0) + 1
        
        # Extract lambda function from solution
        solution = json.loads(ex.get("solution", "{}"))
        lambda_func = solution.get("lambda_function", "unknown")
        lambda_functions[lambda_func] = lambda_functions.get(lambda_func, 0) + 1
        
        # Calculate lengths
        problem_lengths.append(len(ex.get("problem", "")))
        solution_lengths.append(len(ex.get("solution", "")))
        
        # Count reasoning steps
        steps = ex.get("reasoning_steps", [])
        num_steps = len(steps)
        reasoning_steps_counts[num_steps] = reasoning_steps_counts.get(num_steps, 0) + 1
    
    return TrainingStats(
        total_examples=len(examples),
        alert_types=alert_types,
        lambda_functions=lambda_functions,
        avg_problem_length=sum(problem_lengths) / len(problem_lengths) if problem_lengths else 0,
        avg_solution_length=sum(solution_lengths) / len(solution_lengths) if solution_lengths else 0,
        reasoning_steps_count=reasoning_steps_counts
    )


def validate_dataset(examples: List[Dict[str, Any]]) -> tuple[bool, List[str]]:
    """Validate dataset format."""
    errors = []
    
    required_fields = ["problem", "initial_answer", "solution", "reasoning_steps", "metadata"]
    
    for i, ex in enumerate(examples, 1):
        # Check required fields
        for field in required_fields:
            if field not in ex:
                errors.append(f"Example {i}: Missing required field '{field}'")
        
        # Validate solution is valid JSON
        try:
            solution = json.loads(ex.get("solution", "{}"))
            if "lambda_function" not in solution:
                errors.append(f"Example {i}: Solution missing 'lambda_function'")
            if "parameters" not in solution:
                errors.append(f"Example {i}: Solution missing 'parameters'")
        except json.JSONDecodeError as e:
            errors.append(f"Example {i}: Invalid JSON in solution: {e}")
        
        # Validate reasoning_steps is a list
        if not isinstance(ex.get("reasoning_steps", []), list):
            errors.append(f"Example {i}: 'reasoning_steps' must be a list")
        
        # Validate metadata
        metadata = ex.get("metadata", {})
        if "alert_name" not in metadata:
            errors.append(f"Example {i}: Metadata missing 'alert_name'")
        if "expected_lambda" not in metadata:
            errors.append(f"Example {i}: Metadata missing 'expected_lambda'")
    
    return len(errors) == 0, errors


def prepare_trm_format(examples: List[Dict[str, Any]], output_dir: Path) -> Path:
    """Prepare dataset in TRM format."""
    trm_data_dir = output_dir / "trm_data"
    trm_data_dir.mkdir(parents=True, exist_ok=True)
    
    trm_examples = []
    for ex in examples:
        trm_examples.append({
            "problem": ex["problem"],
            "initial_answer": ex["initial_answer"],
            "solution": ex["solution"],
            "reasoning_steps": ex["reasoning_steps"],
            "metadata": ex["metadata"]
        })
    
    # Save as JSONL
    trm_data_path = trm_data_dir / "train.jsonl"
    with open(trm_data_path, 'w') as f:
        for ex in trm_examples:
            f.write(json.dumps(ex) + '\n')
    
    return trm_data_path


def print_stats(stats: TrainingStats):
    """Print dataset statistics."""
    print("=" * 60)
    print("📊 Dataset Statistics")
    print("=" * 60)
    print(f"Total Examples: {stats.total_examples}")
    print()
    
    print("Alert Types:")
    for alert_name, count in sorted(stats.alert_types.items()):
        print(f"  - {alert_name}: {count}")
    print()
    
    print("Lambda Functions:")
    for lambda_func, count in sorted(stats.lambda_functions.items()):
        print(f"  - {lambda_func}: {count}")
    print()
    
    print(f"Average Problem Length: {stats.avg_problem_length:.0f} characters")
    print(f"Average Solution Length: {stats.avg_solution_length:.0f} characters")
    print()
    
    print("Reasoning Steps Distribution:")
    for num_steps, count in sorted(stats.reasoning_steps_count.items()):
        print(f"  - {num_steps} steps: {count} examples")
    print()


def main():
    """Main entry point."""
    dataset_path = sys.argv[1] if len(sys.argv) > 1 else "./data/runbook_test_dataset.jsonl"
    output_dir = Path(sys.argv[2]) if len(sys.argv) > 2 else Path("./models/trm-runbook-only")
    
    if not Path(dataset_path).exists():
        print(f"❌ Dataset not found: {dataset_path}")
        print("   Run: python src/test_trm_runbook.py first")
        sys.exit(1)
    
    print("📚 Loading dataset...")
    examples = load_dataset(dataset_path)
    print(f"✅ Loaded {len(examples)} examples\n")
    
    print("🔍 Validating dataset...")
    is_valid, errors = validate_dataset(examples)
    if not is_valid:
        print("❌ Dataset validation failed:")
        for error in errors:
            print(f"   {error}")
        sys.exit(1)
    print("✅ Dataset validation passed\n")
    
    print("📊 Analyzing dataset...")
    stats = analyze_dataset(examples)
    print_stats(stats)
    
    print("📦 Preparing TRM format...")
    trm_data_path = prepare_trm_format(examples, output_dir)
    print(f"✅ Prepared TRM dataset at: {trm_data_path}\n")
    
    print("=" * 60)
    print("✅ Training Data Preparation Complete!")
    print("=" * 60)
    print(f"Dataset: {dataset_path}")
    print(f"TRM Format: {trm_data_path}")
    print(f"Output Dir: {output_dir}")
    print()
    print("Next steps:")
    print("1. Ensure TRM repository is cloned:")
    print("   git clone https://github.com/SamsungSAILMontreal/TinyRecursiveModels.git")
    print()
    print("2. Run training:")
    print(f"   python src/trm_trainer.py --training-data {dataset_path} --output-dir {output_dir}")
    print()
    print("⚠️  Note: With only 7 examples, consider:")
    print("   - Data augmentation (generate variations)")
    print("   - Using pre-trained TRM model and fine-tuning")
    print("   - Collecting more examples from actual alerts")


if __name__ == "__main__":
    main()
