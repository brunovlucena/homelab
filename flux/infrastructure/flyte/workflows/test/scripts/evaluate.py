#!/usr/bin/env python3
"""
Evaluate fine-tuned FunctionGemma model on test dataset.
"""
import argparse
import json
from pathlib import Path
from typing import List, Dict, Any
import subprocess
import sys

from mlx_lm import load, generate


def load_test_data(test_data_path: Path) -> List[Dict[str, Any]]:
    """Load test dataset."""
    examples = []
    with open(test_data_path, 'r') as f:
        for line in f:
            examples.append(json.loads(line))
    return examples


def evaluate_model(
    model_path: str,
    adapter_path: str,
    test_data: List[Dict[str, Any]],
    max_tokens: int = 512
):
    """Evaluate model on test dataset."""
    print(f"Loading model: {model_path}")
    model, tokenizer = load(model_path, adapter_path=adapter_path)
    
    print(f"Evaluating on {len(test_data)} test examples...\n")
    
    correct = 0
    total = 0
    
    for i, example in enumerate(test_data[:10]):  # Evaluate first 10 for quick test
        if "messages" in example:
            # Function calling format
            user_msg = next(
                (msg["content"] for msg in example["messages"] if msg["role"] == "user"),
                ""
            )
            expected_tool = next(
                (msg.get("tool_calls", [{}])[0].get("function", {}).get("name", "")
                 for msg in example["messages"] if msg["role"] == "assistant"),
                ""
            )
        elif "text" in example:
            # Instruction format - extract prompt
            text = example["text"]
            if "<|im_start|>user\n" in text:
                user_msg = text.split("<|im_start|>user\n")[1].split("<|im_end|>")[0]
                expected = text.split("<|im_start|>assistant\n")[1].split("<|im_end|>")[0] if "<|im_start|>assistant\n" in text else ""
            else:
                continue
        else:
            continue
        
        # Generate response
        prompt = user_msg
        response = generate(
            model,
            tokenizer,
            prompt=prompt,
            max_tokens=max_tokens,
            temp=0.1  # Low temperature for deterministic evaluation
        )
        
        # Simple evaluation: check if response contains expected command
        # In production, use more sophisticated metrics
        if expected_tool:
            correct += 1 if expected_tool.lower() in response.lower() else 0
        elif expected:
            correct += 1 if expected.lower() in response.lower() else 0
        
        total += 1
        
        if i < 3:  # Show first 3 examples
            print(f"\nExample {i+1}:")
            print(f"  Prompt: {prompt[:100]}...")
            print(f"  Response: {response[:200]}...")
    
    accuracy = correct / total if total > 0 else 0
    print(f"\n📊 Evaluation Results:")
    print(f"  Accuracy: {accuracy:.2%} ({correct}/{total})")
    
    return accuracy


def main():
    parser = argparse.ArgumentParser(description="Evaluate fine-tuned FunctionGemma model")
    parser.add_argument(
        "--model",
        type=Path,
        required=True,
        help="Path to base model (MLX format)"
    )
    parser.add_argument(
        "--adapter",
        type=Path,
        required=True,
        help="Path to LoRA adapters"
    )
    parser.add_argument(
        "--test-data",
        type=Path,
        default=Path(__file__).parent / "data" / "test.jsonl",
        help="Test dataset JSONL file"
    )
    parser.add_argument(
        "--max-tokens",
        type=int,
        default=512,
        help="Maximum tokens to generate"
    )
    
    args = parser.parse_args()
    
    # Load test data
    test_data = load_test_data(args.test_data)
    print(f"Loaded {len(test_data)} test examples")
    
    # Evaluate
    accuracy = evaluate_model(
        model_path=str(args.model),
        adapter_path=str(args.adapter),
        test_data=test_data,
        max_tokens=args.max_tokens
    )
    
    print(f"\n✅ Evaluation complete!")


if __name__ == "__main__":
    main()

