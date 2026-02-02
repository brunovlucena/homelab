# Integration Example: Using TRM in Agent-Bruno

This example shows how to integrate Agent-Reasoning (TRM) into an existing agent.

## Step 1: Add Reasoning Client to Agent

```python
# agent-bruno/src/chatbot/handler.py

import os
from agent_reasoning import ReasoningClient, TaskType

class ChatBot:
    def __init__(
        self,
        # ... existing parameters ...
        reasoning_enabled: bool = True,
    ):
        # ... existing init ...
        
        # Initialize reasoning client
        self.reasoning_enabled = reasoning_enabled
        if self.reasoning_enabled:
            self.reasoning_client = ReasoningClient(
                base_url=os.getenv(
                    "REASONING_SERVICE_URL",
                    "http://agent-reasoning.ai-agents.svc.cluster.local:8080"
                )
            )
        else:
            self.reasoning_client = None
```

## Step 2: Detect When Reasoning is Needed

```python
def _needs_reasoning(self, message: str) -> bool:
    """Detect if message requires complex reasoning."""
    reasoning_keywords = [
        "solve", "plan", "optimize", "figure out", 
        "calculate", "determine", "analyze", "how should",
        "what's the best way", "troubleshoot", "debug"
    ]
    return any(kw in message.lower() for kw in reasoning_keywords)

def _detect_task_type(self, message: str) -> TaskType:
    """Detect task type from message."""
    message_lower = message.lower()
    
    if any(kw in message_lower for kw in ["plan", "deploy", "setup", "design"]):
        return TaskType.PLANNING
    elif any(kw in message_lower for kw in ["optimize", "improve", "efficient", "cost"]):
        return TaskType.OPTIMIZATION
    elif any(kw in message_lower for kw in ["fix", "debug", "troubleshoot", "why", "error"]):
        return TaskType.TROUBLESHOOTING
    elif any(kw in message_lower for kw in ["solve", "logic", "puzzle", "calculate"]):
        return TaskType.LOGIC
    else:
        return TaskType.GENERAL
```

## Step 3: Integrate Reasoning into Chat Flow

```python
async def _chat_with_memory(
    self,
    message: str,
    conversation_id: Optional[str],
    user_id: Optional[str],
    start_time: float,
    log,
) -> ChatResponse:
    """Chat with Domain Memory enabled."""
    
    # ... existing memory setup ...
    
    # Check if reasoning is needed
    use_reasoning = (
        self.reasoning_enabled and
        self.reasoning_client and
        self._needs_reasoning(message)
    )
    
    if use_reasoning:
        try:
            # Build context for reasoning
            reasoning_context = {
                "conversation_history": context.get("conversation", {}).get("messages", [])[-5:],
                "user_context": context.get("user", {}),
                "system_state": self._get_system_state(),  # Your method to get current state
            }
            
            # Perform reasoning
            reasoning_result = await self.reasoning_client.reason(
                question=message,
                context=reasoning_context,
                max_steps=6,
                task_type=self._detect_task_type(message),
                conversation_id=conversation_id,
            )
            
            # Use reasoning result
            response_text = reasoning_result.answer
            
            log.info(
                "reasoning_used",
                steps=reasoning_result.steps,
                confidence=reasoning_result.confidence,
                task_type=reasoning_result.task_type.value,
            )
            
        except Exception as e:
            # Fallback to normal LLM if reasoning fails
            log.warning("reasoning_failed_fallback", error=str(e))
            use_reasoning = False
    
    if not use_reasoning:
        # Normal LLM path
        enhanced_prompt = self._build_prompt_with_context(message, context)
        response_text, tokens, input_tokens, output_tokens = await self._query_ollama(enhanced_prompt)
    
    # ... rest of the method ...
```

## Step 4: Add Configuration

```python
# In main.py lifespan

chatbot = ChatBot(
    # ... existing parameters ...
    reasoning_enabled=os.getenv("REASONING_ENABLED", "true").lower() == "true",
)
```

## Step 5: Update Environment Variables

```yaml
# k8s/kustomize/base/configmap.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: agent-bruno-config
data:
  REASONING_ENABLED: "true"
  REASONING_SERVICE_URL: "http://agent-reasoning.ai-agents.svc.cluster.local:8080"
```

## Example Use Cases

### Use Case 1: Infrastructure Planning

**User**: "How should I deploy this application across my clusters?"

**Flow**:
1. Agent-Bruno detects planning task
2. Calls Agent-Reasoning with:
   - Question: Deployment planning
   - Context: Cluster state, app requirements
   - Task type: PLANNING
3. TRM recursively reasons through:
   - Resource requirements
   - Network topology
   - High availability
   - Cost optimization
4. Returns structured deployment plan

### Use Case 2: Troubleshooting

**User**: "Why is my service slow?"

**Flow**:
1. Agent-Bruno detects troubleshooting task
2. Calls Agent-Reasoning with:
   - Question: Root cause analysis
   - Context: Metrics, logs, topology
   - Task type: TROUBLESHOOTING
3. TRM reasons through:
   - Dependency chains
   - Bottleneck identification
   - Solution prioritization
4. Returns diagnosis and recommendations

### Use Case 3: Optimization

**User**: "Optimize my Kubernetes resource allocation"

**Flow**:
1. Agent-Bruno detects optimization task
2. Calls Agent-Reasoning with:
   - Question: Resource optimization
   - Context: Current allocations, workloads
   - Task type: OPTIMIZATION
3. TRM iteratively improves solution
4. Returns optimized configuration

## Testing

```python
# tests/unit/test_reasoning_integration.py

import pytest
from unittest.mock import AsyncMock, patch

@pytest.mark.asyncio
async def test_reasoning_integration():
    """Test reasoning integration in chatbot."""
    chatbot = ChatBot(reasoning_enabled=True)
    
    # Mock reasoning client
    with patch.object(chatbot.reasoning_client, 'reason') as mock_reason:
        mock_reason.return_value = ReasoningResponse(
            answer="Optimized solution...",
            steps=6,
            confidence=0.9,
            reasoning_trace=[],
            duration_ms=1000.0,
            task_type=TaskType.OPTIMIZATION,
        )
        
        # Test reasoning path
        result = await chatbot.chat(
            message="How should I optimize my cluster?",
            conversation_id="test-123",
        )
        
        assert "optimized" in result.response.lower()
        mock_reason.assert_called_once()
```

## Monitoring

Add metrics for reasoning usage:

```python
# In handler.py

REASONING_REQUESTS = Counter(
    "agent_bruno_reasoning_requests_total",
    "Total reasoning requests",
    ["task_type", "status"]
)

# In chat method
if use_reasoning:
    REASONING_REQUESTS.labels(
        task_type=reasoning_result.task_type.value,
        status="success"
    ).inc()
```

## Best Practices

1. **Always have a fallback**: If reasoning fails, fall back to normal LLM
2. **Set appropriate max_steps**: Simple tasks (3-4), complex (6-8)
3. **Provide rich context**: Include relevant system state, history
4. **Monitor usage**: Track when reasoning is used vs. LLM
5. **Cache results**: For similar questions, cache reasoning results

