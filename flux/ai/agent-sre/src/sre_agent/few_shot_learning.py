"""
Few-Shot Learning - Phase 2: Example Database and Retrieval

Stores and retrieves successful remediation examples to enhance AI prompts.
"""
from typing import Dict, Any, List, Optional
import json
import os
from pathlib import Path
from datetime import datetime, timedelta
import structlog

logger = structlog.get_logger()


class RemediationExample:
    """Represents a remediation example."""
    
    def __init__(
        self,
        alertname: str,
        labels: Dict[str, Any],
        lambda_function: str,
        parameters: Dict[str, Any],
        success: bool,
        timestamp: Optional[datetime] = None,
        reasoning: Optional[str] = None
    ):
        self.alertname = alertname
        self.labels = labels
        self.lambda_function = lambda_function
        self.parameters = parameters
        self.success = success
        self.timestamp = timestamp or datetime.utcnow()
        self.reasoning = reasoning
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for storage."""
        return {
            "alertname": self.alertname,
            "labels": self.labels,
            "lambda_function": self.lambda_function,
            "parameters": self.parameters,
            "success": self.success,
            "timestamp": self.timestamp.isoformat(),
            "reasoning": self.reasoning
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "RemediationExample":
        """Create from dictionary."""
        return cls(
            alertname=data["alertname"],
            labels=data["labels"],
            lambda_function=data["lambda_function"],
            parameters=data["parameters"],
            success=data["success"],
            timestamp=datetime.fromisoformat(data.get("timestamp", datetime.utcnow().isoformat())),
            reasoning=data.get("reasoning")
        )
    
    def similarity_score(self, other_labels: Dict[str, Any]) -> float:
        """Calculate similarity score with another alert's labels."""
        score = 0.0
        total_keys = 0
        
        # Check for matching keys
        for key in set(list(self.labels.keys()) + list(other_labels.keys())):
            total_keys += 1
            if key in self.labels and key in other_labels:
                if self.labels[key] == other_labels[key]:
                    score += 1.0
                elif key in ["alertname", "namespace", "kind"]:
                    # Important keys get partial match
                    score += 0.5
        
        return score / total_keys if total_keys > 0 else 0.0


class ExampleDatabase:
    """Stores and retrieves remediation examples."""
    
    def __init__(self, storage_path: Optional[str] = None):
        """
        Initialize example database.
        
        Args:
            storage_path: Path to JSON file for persistence (default: in-memory only)
        """
        self.storage_path = storage_path
        self.examples: List[RemediationExample] = []
        
        if storage_path:
            self._load_examples()
    
    def _load_examples(self):
        """Load examples from storage."""
        if not self.storage_path:
            return
        
        path = Path(self.storage_path)
        if path.exists():
            try:
                with open(path, "r") as f:
                    data = json.load(f)
                    self.examples = [
                        RemediationExample.from_dict(item)
                        for item in data.get("examples", [])
                    ]
                logger.info(
                    "examples_loaded",
                    count=len(self.examples),
                    path=str(path)
                )
            except Exception as e:
                logger.warning(
                    "failed_to_load_examples",
                    error=str(e),
                    path=str(path)
                )
    
    def _save_examples(self):
        """Save examples to storage."""
        if not self.storage_path:
            return
        
        try:
            path = Path(self.storage_path)
            path.parent.mkdir(parents=True, exist_ok=True)
            
            data = {
                "examples": [ex.to_dict() for ex in self.examples],
                "updated_at": datetime.utcnow().isoformat()
            }
            
            with open(path, "w") as f:
                json.dump(data, f, indent=2)
            
            logger.debug("examples_saved", count=len(self.examples), path=str(path))
        except Exception as e:
            logger.warning("failed_to_save_examples", error=str(e))
    
    def add_example(
        self,
        alertname: str,
        labels: Dict[str, Any],
        lambda_function: str,
        parameters: Dict[str, Any],
        success: bool,
        reasoning: Optional[str] = None
    ):
        """Add a new remediation example."""
        example = RemediationExample(
            alertname=alertname,
            labels=labels,
            lambda_function=lambda_function,
            parameters=parameters,
            success=success,
            reasoning=reasoning
        )
        
        self.examples.append(example)
        
        # Keep only recent examples (last 1000)
        if len(self.examples) > 1000:
            self.examples = sorted(
                self.examples,
                key=lambda x: x.timestamp,
                reverse=True
            )[:1000]
        
        self._save_examples()
        
        logger.info(
            "example_added",
            alertname=alertname,
            lambda_function=lambda_function,
            success=success
        )
    
    def find_similar_examples(
        self,
        alertname: str,
        labels: Dict[str, Any],
        top_k: int = 5,
        min_similarity: float = 0.3,
        only_successful: bool = True
    ) -> List[RemediationExample]:
        """
        Find similar examples based on alert name and labels.
        
        Args:
            alertname: Name of the alert
            labels: Alert labels
            top_k: Number of examples to return
            min_similarity: Minimum similarity score (0-1)
            only_successful: Only return successful remediations
            
        Returns:
            List of similar examples, sorted by similarity
        """
        candidates = []
        
        for example in self.examples:
            # Filter by success if requested
            if only_successful and not example.success:
                continue
            
            # Check alertname match (exact match gets bonus)
            alertname_match = 1.0 if example.alertname == alertname else 0.0
            
            # Calculate label similarity
            label_similarity = example.similarity_score(labels)
            
            # Combined score (alertname match is important)
            similarity = (alertname_match * 0.6) + (label_similarity * 0.4)
            
            if similarity >= min_similarity:
                candidates.append((similarity, example))
        
        # Sort by similarity (descending)
        candidates.sort(key=lambda x: x[0], reverse=True)
        
        # Return top K
        return [ex for _, ex in candidates[:top_k]]
    
    def get_examples_for_prompt(
        self,
        alertname: str,
        labels: Dict[str, Any],
        top_k: int = 5
    ) -> List[Dict[str, Any]]:
        """Get examples formatted for prompt inclusion."""
        examples = self.find_similar_examples(alertname, labels, top_k=top_k)
        
        return [
            {
                "alert": ex.alertname,
                "labels": ex.labels,
                "lambda_function": ex.lambda_function,
                "parameters": ex.parameters,
                "success": ex.success,
                "reasoning": ex.reasoning
            }
            for ex in examples
        ]


def create_few_shot_prompt(
    alert_data: Dict[str, Any],
    examples: List[Dict[str, Any]]
) -> str:
    """Create prompt with few-shot examples."""
    alertname = alert_data.get("labels", {}).get("alertname", "unknown")
    labels = alert_data.get("labels", {})
    
    prompt = "Here are examples of successful remediation selections:\n\n"
    
    for i, example in enumerate(examples[:5], 1):  # Use top 5 similar examples
        prompt += f"Example {i}:\n"
        prompt += f"Alert: {example['alert']}\n"
        prompt += f"Labels: {json.dumps(example['labels'], indent=2)}\n"
        prompt += f"Selected: {example['lambda_function']}\n"
        prompt += f"Parameters: {json.dumps(example['parameters'], indent=2)}\n"
        if example.get('reasoning'):
            prompt += f"Reasoning: {example['reasoning']}\n"
        prompt += f"Result: {'Success' if example['success'] else 'Failed'}\n\n"
    
    prompt += f"\nCurrent Alert:\n"
    prompt += f"Alert: {alertname}\n"
    prompt += f"Labels: {json.dumps(labels, indent=2)}\n"
    prompt += "Select the appropriate Lambda function and parameters based on the examples above.\n"
    
    return prompt


# Global example database instance
_example_db: Optional[ExampleDatabase] = None


def get_example_database(storage_path: Optional[str] = None) -> ExampleDatabase:
    """Get or create global example database instance."""
    global _example_db
    
    if _example_db is None:
        if storage_path is None:
            # Default to data directory in project root
            default_path = os.path.join(
                os.path.dirname(os.path.dirname(os.path.dirname(__file__))),
                "data",
                "remediation_examples.json"
            )
            storage_path = default_path
        
        _example_db = ExampleDatabase(storage_path=storage_path)
    
    return _example_db

