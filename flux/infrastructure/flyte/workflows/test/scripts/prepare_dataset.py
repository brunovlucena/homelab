#!/usr/bin/env python3
"""
Prepare training dataset from RUNBOOK.md for FunctionGemma fine-tuning.

Converts runbook incidents into function-calling format suitable for FunctionGemma.
Each training example represents an alert -> remediation action pattern.
"""
import json
import re
import argparse
from pathlib import Path
from typing import List, Dict, Any
from dataclasses import dataclass


@dataclass
class AlertExample:
    """Represents a single alert-to-action training example."""
    alert_name: str
    symptoms: str
    investigation_steps: List[str]
    resolution_commands: List[str]
    auto_resolvable: bool
    section: str


def parse_runbook_md(runbook_path: Path) -> List[AlertExample]:
    """Parse RUNBOOK.md and extract alert examples."""
    content = runbook_path.read_text()
    examples = []
    
    # Pattern to match alert sections
    alert_pattern = r'### Alert: (\w+)\s*\n\n\*\*Symptoms:\*\*\s*\n(.*?)\n\n\*\*Investigation:\*\*|### Alert: (\w+)\s*\n\n\*\*Symptoms:\*\*\s*\n(.*?)\n\n\*\*Resolution:'
    
    # Split by major sections
    sections = re.split(r'^## ', content, flags=re.MULTILINE)
    
    current_section = ""
    for section in sections:
        if not section.strip():
            continue
            
        # Extract section name
        section_match = re.match(r'^([^\n]+)', section)
        if section_match:
            current_section = section_match.group(1).strip()
        
        # Find all alerts in this section
        alert_matches = re.finditer(
            r'### Alert: ([^\n]+)\s*\n\n\*\*Symptoms:\*\*\s*\n(.*?)(?=\n\n\*\*Investigation:|\n\n\*\*Resolution:)',
            section,
            re.DOTALL
        )
        
        for match in alert_matches:
            alert_name = match.group(1).strip()
            symptoms = match.group(2).strip()
            
            # Extract investigation steps
            investigation_match = re.search(
                r'\*\*Investigation:\*\*\s*\n```bash\n(.*?)\n```',
                section[match.end():],
                re.DOTALL
            )
            investigation_steps = []
            if investigation_match:
                investigation_steps = [
                    line.strip()
                    for line in investigation_match.group(1).split('\n')
                    if line.strip() and not line.strip().startswith('#')
                ]
            
            # Extract resolution commands
            resolution_match = re.search(
                r'\*\*Resolution:\*\*\s*\n(\d+\.\s*\*\*[^\n]+\*\*:.*?)(?=\n\n###|\n\n---|\Z)',
                section[match.end():],
                re.DOTALL
            )
            resolution_commands = []
            if resolution_match:
                # Extract all code blocks with commands
                code_blocks = re.findall(
                    r'```bash\n(.*?)\n```',
                    resolution_match.group(1),
                    re.DOTALL
                )
                for block in code_blocks:
                    commands = [
                        line.strip()
                        for line in block.split('\n')
                        if line.strip() and not line.strip().startswith('#')
                    ]
                    resolution_commands.extend(commands)
            
            # Determine if auto-resolvable
            auto_resolvable = (
                "flux reconcile" in " ".join(resolution_commands).lower() or
                "kubectl delete pod" in " ".join(resolution_commands).lower() or
                "kubectl rollout restart" in " ".join(resolution_commands).lower()
            ) and "security" not in current_section.lower()
            
            examples.append(AlertExample(
                alert_name=alert_name,
                symptoms=symptoms,
                investigation_steps=investigation_steps,
                resolution_commands=resolution_commands,
                auto_resolvable=auto_resolvable,
                section=current_section
            ))
    
    return examples


def create_functiongemma_example(example: AlertExample) -> Dict[str, Any]:
    """Convert alert example to FunctionGemma function-calling format."""
    
    # Create user message (the alert/symptoms)
    user_content = f"""Alert: {example.alert_name}

Symptoms:
{example.symptoms}

Investigation Steps:
{chr(10).join(f"- {step}" for step in example.investigation_steps[:5])}

This alert requires immediate remediation. Follow the runbook procedures to resolve the incident."""
    
    # Create function call (the remediation action)
    # FunctionGemma expects function calls in a specific format
    if example.resolution_commands:
        # Primary remediation command
        primary_command = example.resolution_commands[0] if example.resolution_commands else ""
        
        # Determine function name based on command type
        if "flux reconcile" in primary_command:
            tool_name = "reconcile_flux_resource"
            # Extract resource details
            parts = primary_command.split()
            resource_type = ""
            resource_name = ""
            namespace = "flux-system"
            for i, part in enumerate(parts):
                if part in ["kustomization", "source", "helmrelease"]:
                    resource_type = parts[i+1] if i+1 < len(parts) else ""
                if part == "-n" and i+1 < len(parts):
                    namespace = parts[i+1]
            tool_arguments = json.dumps({
                "resource_type": resource_type,
                "resource_name": resource_name,
                "namespace": namespace,
                "command": primary_command
            })
        elif "kubectl" in primary_command:
            tool_name = "execute_kubectl_command"
            tool_arguments = json.dumps({
                "command": primary_command,
                "description": f"Execute remediation command for {example.alert_name}"
            })
        else:
            tool_name = "execute_remediation"
            tool_arguments = json.dumps({
                "command": primary_command,
                "alert": example.alert_name,
                "section": example.section
            })
    else:
        tool_name = "escalate_to_human"
        tool_arguments = json.dumps({
            "reason": "No automated remediation available",
            "alert": example.alert_name,
            "severity": "requires_manual_intervention"
        })
    
    # FunctionGemma format
    return {
        "messages": [
            {
                "role": "user",
                "content": user_content
            },
            {
                "role": "assistant",
                "content": "",
                "tool_calls": [
                    {
                        "function": {
                            "name": tool_name,
                            "arguments": tool_arguments
                        }
                    }
                ]
            }
        ]
    }


def create_instruction_example(example: AlertExample) -> Dict[str, Any]:
    """Create instruction-following format (alternative to function calling)."""
    
    prompt = f"""You are an SRE agent responsible for incident response. When an alert fires, you must follow the runbook to resolve it.

Alert: {example.alert_name}
Section: {example.section}
Symptoms: {example.symptoms}

Investigation Steps:
{chr(10).join(f"{i+1}. {step}" for i, step in enumerate(example.investigation_steps[:5]))}

Resolution Actions:
{chr(10).join(f"{i+1}. {cmd}" for i, cmd in enumerate(example.resolution_commands[:3]))}

Based on the runbook, provide the exact remediation command to execute."""
    
    completion = example.resolution_commands[0] if example.resolution_commands else "escalate_to_human"
    
    return {
        "text": f"<|im_start|>user\n{prompt}<|im_end|>\n<|im_start|>assistant\n{completion}<|im_end|>"
    }


def main():
    parser = argparse.ArgumentParser(description="Prepare training dataset from RUNBOOK.md")
    parser.add_argument(
        "--runbook",
        type=Path,
        default=Path(__file__).parent.parent / "docs" / "RUNBOOK.md",
        help="Path to RUNBOOK.md"
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path(__file__).parent / "data",
        help="Output directory for training data"
    )
    parser.add_argument(
        "--format",
        choices=["function_calling", "instruction"],
        default="instruction",
        help="Training data format"
    )
    parser.add_argument(
        "--train-split",
        type=float,
        default=0.8,
        help="Training set split ratio"
    )
    parser.add_argument(
        "--val-split",
        type=float,
        default=0.1,
        help="Validation set split ratio"
    )
    
    args = parser.parse_args()
    
    # Create output directory
    args.output_dir.mkdir(parents=True, exist_ok=True)
    
    # Parse runbook
    print(f"Parsing runbook: {args.runbook}")
    examples = parse_runbook_md(args.runbook)
    print(f"Found {len(examples)} alert examples")
    
    # Convert to training format
    training_data = []
    for example in examples:
        if args.format == "function_calling":
            training_data.append(create_functiongemma_example(example))
        else:
            training_data.append(create_instruction_example(example))
    
    # Split dataset
    total = len(training_data)
    train_size = int(total * args.train_split)
    val_size = int(total * args.val_split)
    
    train_data = training_data[:train_size]
    val_data = training_data[train_size:train_size + val_size]
    test_data = training_data[train_size + val_size:]
    
    # Write JSONL files
    def write_jsonl(path: Path, data: List[Dict]):
        with open(path, 'w') as f:
            for item in data:
                f.write(json.dumps(item) + '\n')
    
    train_path = args.output_dir / "train.jsonl"
    val_path = args.output_dir / "val.jsonl"
    test_path = args.output_dir / "test.jsonl"
    
    write_jsonl(train_path, train_data)
    write_jsonl(val_path, val_data)
    write_jsonl(test_path, test_data)
    
    print(f"\nDataset prepared:")
    print(f"  Training: {len(train_data)} examples -> {train_path}")
    print(f"  Validation: {len(val_data)} examples -> {val_path}")
    print(f"  Test: {len(test_data)} examples -> {test_path}")
    print(f"\nTotal examples: {total}")


if __name__ == "__main__":
    main()

