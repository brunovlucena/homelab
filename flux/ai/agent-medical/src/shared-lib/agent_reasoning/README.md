# Agent-Reasoning Client Library

Shared library for integrating Agent-Reasoning (TRM) into homelab agents.

## Usage

### Basic Usage

```python
from agent_reasoning import ReasoningClient, TaskType

# Initialize client
client = ReasoningClient(
    base_url="http://agent-reasoning.ai-agents.svc.cluster.local:8080"
)

# Perform reasoning
result = await client.reason(
    question="How should I optimize my Kubernetes cluster?",
    context={
        "current_nodes": 10,
        "workloads": [...],
    },
    max_steps=6,
    task_type=TaskType.OPTIMIZATION,
)

print(f"Answer: {result.answer}")
print(f"Confidence: {result.confidence}")
print(f"Steps: {result.steps}")
```

### Integration Example

```python
# In agent-bruno/src/chatbot/handler.py

from agent_reasoning import ReasoningClient, TaskType

class ChatBot:
    def __init__(self, ...):
        # ... existing init ...
        self.reasoning_client = ReasoningClient()
    
    async def chat(self, message: str, ...):
        # Detect if message requires complex reasoning
        if self._needs_reasoning(message):
            # Use TRM for complex reasoning
            reasoning_result = await self.reasoning_client.reason(
                question=message,
                context=self._build_context(),
                max_steps=6,
                task_type=self._detect_task_type(message),
            )
            return reasoning_result.answer
        
        # Otherwise use normal LLM path
        return await self._query_ollama(...)
    
    def _needs_reasoning(self, message: str) -> bool:
        """Detect if message requires complex reasoning."""
        reasoning_keywords = [
            "solve", "plan", "optimize", "figure out", 
            "calculate", "determine", "analyze"
        ]
        return any(kw in message.lower() for kw in reasoning_keywords)
    
    def _detect_task_type(self, message: str) -> TaskType:
        """Detect task type from message."""
        message_lower = message.lower()
        
        if any(kw in message_lower for kw in ["plan", "deploy", "setup"]):
            return TaskType.PLANNING
        elif any(kw in message_lower for kw in ["optimize", "improve", "efficient"]):
            return TaskType.OPTIMIZATION
        elif any(kw in message_lower for kw in ["fix", "debug", "troubleshoot", "why"]):
            return TaskType.TROUBLESHOOTING
        elif any(kw in message_lower for kw in ["solve", "logic", "puzzle"]):
            return TaskType.LOGIC
        else:
            return TaskType.GENERAL
```

## Installation

Add to your agent's `requirements.txt`:

```
# Agent-Reasoning client (shared library)
# Path: ../../shared-lib/agent_reasoning
```

Or install as package:

```bash
pip install -e ../../shared-lib/agent_reasoning
```

