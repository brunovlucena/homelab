#!/usr/bin/env python3
"""
Build TRM-compatible dataset from runbook JSONL.

Converts text-based runbook examples to TRM's puzzle dataset format.
"""

import json
import os
import numpy as np
from pathlib import Path
from typing import List, Dict, Any
import sys

# Add TRM to path
sys.path.insert(0, str(Path(__file__).parent.parent.parent / "trm"))
from dataset.common import PuzzleDatasetMetadata


def text_to_token_ids(text: str, char2id: Dict[str, int], max_len: int) -> np.ndarray:
    """Convert text to token IDs."""
    tokens = [char2id.get(c, 0) for c in text[:max_len]]
    # Pad to max_len
    tokens = tokens + [0] * (max_len - len(tokens))
    return np.array(tokens, dtype=np.uint8)


def build_char_vocab(examples: List[Dict[str, Any]]) -> Dict[str, int]:
    """Build character vocabulary from examples."""
    all_chars = set()
    for ex in examples:
        problem = ex.get("problem", "")
        solution = ex.get("solution", "")
        all_chars.update(problem + solution)
    
    # Create char2id mapping (0 is PAD, 1+ are chars)
    chars = sorted(all_chars)
    char2id = {chr(0): 0}  # PAD
    for i, char in enumerate(chars, start=1):
        char2id[char] = i
    
    return char2id


def convert_runbook_dataset(input_jsonl: str, output_dir: str):
    """Convert runbook JSONL to TRM dataset format."""
    print(f"📚 Reading examples from {input_jsonl}...")
    
    # Read examples
    examples = []
    with open(input_jsonl, 'r') as f:
        for line in f:
            if line.strip():
                examples.append(json.loads(line))
    
    print(f"✅ Loaded {len(examples)} examples")
    
    # Build vocabulary
    print("🔤 Building vocabulary...")
    char2id = build_char_vocab(examples)
    vocab_size = len(char2id)
    print(f"✅ Vocabulary size: {vocab_size}")
    
    # Determine max sequence length
    max_problem_len = max(len(ex.get("problem", "")) for ex in examples)
    max_solution_len = max(len(ex.get("solution", "")) for ex in examples)
    seq_len = max(max_problem_len, max_solution_len, 512)  # At least 512
    print(f"📏 Max sequence length: {seq_len}")
    
    # Convert to numpy arrays
    print("🔄 Converting to TRM format...")
    results = {
        "inputs": [],
        "labels": [],
        "puzzle_indices": [],
        "puzzle_identifiers": [],
        "group_indices": []
    }
    
    example_id = 0
    puzzle_id = 0
    
    # group_indices starts at 0
    results["group_indices"].append(0)
    
    for ex in examples:
        problem = ex.get("problem", "")
        solution = ex.get("solution", "")
        
        # Convert to token IDs
        input_ids = text_to_token_ids(problem, char2id, seq_len)
        label_ids = text_to_token_ids(solution, char2id, seq_len)
        
        results["inputs"].append(input_ids)
        results["labels"].append(label_ids)
        
        example_id += 1
        puzzle_id += 1
        
        results["puzzle_indices"].append(example_id)
        results["puzzle_identifiers"].append(0)
        
        # group_indices is cumulative - add after each puzzle
        results["group_indices"].append(puzzle_id)
    
    # puzzle_indices needs one extra element (total count)
    results["puzzle_indices"].append(example_id)
    
    # Convert to numpy
    results = {
        "inputs": np.vstack(results["inputs"]),
        "labels": np.vstack(results["labels"]),
        "group_indices": np.array(results["group_indices"], dtype=np.int32),
        "puzzle_indices": np.array(results["puzzle_indices"], dtype=np.int32),
        "puzzle_identifiers": np.array(results["puzzle_identifiers"], dtype=np.int32),
    }
    
    # Create metadata
    metadata = PuzzleDatasetMetadata(
        seq_len=seq_len,
        vocab_size=vocab_size,
        pad_id=0,
        ignore_label_id=0,
        blank_identifier_id=0,
        num_puzzle_identifiers=1,
        total_groups=len(results["group_indices"]) - 1,
        mean_puzzle_examples=1.0,
        total_puzzles=len(results["group_indices"]) - 1,
        sets=["all"]
    )
    
    # Save
    save_dir = Path(output_dir) / "train"
    save_dir.mkdir(parents=True, exist_ok=True)
    
    print(f"💾 Saving dataset to {save_dir}...")
    
    # Save metadata
    with open(save_dir / "dataset.json", "w") as f:
        json.dump(metadata.model_dump(), f, indent=2)
    
    # Save data as numpy arrays with set name prefix (TRM expects "all__inputs.npy" format)
    set_name = "all"  # From metadata.sets
    for k, v in results.items():
        np.save(save_dir / f"{set_name}__{k}.npy", v)
    
    print(f"✅ Dataset saved!")
    print(f"   Examples: {len(examples)}")
    print(f"   Vocab size: {vocab_size}")
    print(f"   Sequence length: {seq_len}")
    print(f"   Output: {save_dir}")


def main():
    import argparse
    parser = argparse.ArgumentParser(description="Convert runbook JSONL to TRM dataset")
    parser.add_argument("--input", type=str, required=True, help="Input JSONL file")
    parser.add_argument("--output", type=str, required=True, help="Output directory")
    
    args = parser.parse_args()
    convert_runbook_dataset(args.input, args.output)


if __name__ == "__main__":
    main()
