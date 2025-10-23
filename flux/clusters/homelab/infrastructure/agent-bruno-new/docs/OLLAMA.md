# LLM Generation (Ollama) - Agent Bruno

**[← Back to Architecture](ARCHITECTURE.md)** | **[Main README](../README.md)**

---

## Table of Contents
1. [Overview](#overview)
2. [Ollama Integration](#ollama-integration)
3. [Model Selection](#model-selection)
4. [Prompt Engineering](#prompt-engineering)
5. [Generation Parameters](#generation-parameters)
6. [Streaming & Responses](#streaming--responses)
7. [Error Handling & Retries](#error-handling--retries)
8. [Performance Optimization](#performance-optimization)
9. [Observability](#observability)

---

## Overview

LLM Generation is the stage where Agent Bruno sends the assembled context and query to Ollama for response generation. This is the core reasoning and response generation component.

### Goals
- 🤖 **Generate accurate responses** - Leverage LLM capabilities for high-quality answers
- ⚡ **Minimize latency** - Optimize for fast response times
- 🎯 **Ensure relevance** - Stay grounded in retrieved context
- 💰 **Manage costs** - Efficient token usage (though Ollama is self-hosted)
- 🔄 **Handle errors gracefully** - Robust error handling and retries

### Architecture Position

```
Query Processing
    ↓
Retrieval (Semantic + Keyword)
    ↓
Fusion & Re-ranking
    ↓
Context Assembly
    ↓
┌─────────────────────────────────────────┐
│      LLM Generation (Ollama)             │  ← YOU ARE HERE
│  • Model Selection                      │
│  • Prompt Construction                  │
│  • Generation Parameters                │
│  • Streaming Response                   │
│  • Error Handling                       │
└─────────────────────────────────────────┘
    ↓
Response Post-processing
    ↓
Return to User
```

---

## Ollama Integration

### Ollama Server Configuration

Agent Bruno connects to Ollama running on the Mac Studio at `192.168.0.16:11434`.

```python
from typing import Optional, Dict, List, AsyncIterator
import aiohttp
import asyncio
from pydantic import BaseModel

class OllamaConfig(BaseModel):
    """Ollama server configuration"""
    base_url: str = "http://192.168.0.16:11434"
    timeout: int = 120  # seconds
    max_retries: int = 3
    retry_delay: float = 1.0  # seconds
    
class OllamaClient:
    """
    Client for Ollama API.
    
    Supports:
    - Synchronous and streaming generation
    - Model management
    - Health checks
    - Connection pooling
    """
    
    def __init__(self, config: OllamaConfig = None):
        self.config = config or OllamaConfig()
        self.session: Optional[aiohttp.ClientSession] = None
    
    async def __aenter__(self):
        """Create aiohttp session with connection pooling"""
        connector = aiohttp.TCPConnector(
            limit=10,  # Max 10 concurrent connections
            limit_per_host=5,
            ttl_dns_cache=300
        )
        
        timeout = aiohttp.ClientTimeout(total=self.config.timeout)
        
        self.session = aiohttp.ClientSession(
            connector=connector,
            timeout=timeout
        )
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Close session"""
        if self.session:
            await self.session.close()
    
    async def health_check(self) -> bool:
        """
        Check if Ollama server is healthy.
        """
        try:
            async with self.session.get(f"{self.config.base_url}/api/tags") as response:
                return response.status == 200
        except Exception as e:
            logger.error("ollama_health_check_failed", error=str(e))
            return False
    
    async def list_models(self) -> List[str]:
        """
        List available models on Ollama server.
        """
        async with self.session.get(f"{self.config.base_url}/api/tags") as response:
            if response.status == 200:
                data = await response.json()
                return [model["name"] for model in data.get("models", [])]
            return []
```

---

## Model Selection

### Available Models

```python
from enum import Enum

class OllamaModel(str, Enum):
    """Available Ollama models"""
    # General purpose
    LLAMA3_2_3B = "llama3.2:3b"           # Fast, good for simple queries
    LLAMA3_2_8B = "llama3.2:8b"           # Balanced performance/quality
    LLAMA3_1_70B = "llama3.1:70b"         # High quality, slower
    
    # Code-focused
    DEEPSEEK_CODER_33B = "deepseek-coder:33b"  # Best for code
    CODELLAMA_34B = "codellama:34b"       # Code generation
    
    # Fine-tuned (custom)
    SRE_AGENT_V1 = "sre-agent:v1.0"       # Fine-tuned for SRE tasks
    
    # Embeddings
    NOMIC_EMBED = "nomic-embed-text"      # For embeddings

class ModelSelector:
    """
    Select optimal model based on query characteristics.
    """
    
    def __init__(self):
        self.model_capabilities = {
            OllamaModel.LLAMA3_2_3B: {
                "speed": "fast",
                "quality": "good",
                "max_context": 8192,
                "use_cases": ["simple_queries", "fast_response"]
            },
            OllamaModel.LLAMA3_2_8B: {
                "speed": "medium",
                "quality": "excellent",
                "max_context": 8192,
                "use_cases": ["general", "default"]
            },
            OllamaModel.DEEPSEEK_CODER_33B: {
                "speed": "slow",
                "quality": "excellent",
                "max_context": 16384,
                "use_cases": ["code_generation", "code_review"]
            },
            OllamaModel.SRE_AGENT_V1: {
                "speed": "medium",
                "quality": "excellent",
                "max_context": 8192,
                "use_cases": ["sre_tasks", "troubleshooting", "runbooks"]
            }
        }
    
    def select_model(
        self,
        query_type: str,
        query_complexity: str,
        requires_code: bool = False,
        max_latency_ms: int = 5000
    ) -> OllamaModel:
        """
        Select optimal model based on query characteristics.
        
        Decision tree:
        1. If requires code → use code-focused model
        2. If SRE/troubleshooting → use fine-tuned model
        3. If simple + fast required → use small model
        4. Default → use balanced model
        """
        # Code queries
        if requires_code:
            return OllamaModel.DEEPSEEK_CODER_33B
        
        # SRE-specific queries
        if query_type in ["troubleshooting", "lookup_runbook", "debug_issue"]:
            return OllamaModel.SRE_AGENT_V1
        
        # Fast response required
        if max_latency_ms < 2000 and query_complexity == "simple":
            return OllamaModel.LLAMA3_2_3B
        
        # Default: balanced model
        return OllamaModel.LLAMA3_2_8B
```

---

## Prompt Engineering

### System Prompt Construction

```python
class PromptBuilder:
    """
    Build prompts for Ollama LLM.
    """
    
    def __init__(self):
        self.base_system_prompt = """You are Agent Bruno, an expert SRE AI assistant with deep knowledge of:
- Kubernetes and cloud-native technologies
- Observability (Prometheus, Grafana, Loki, Tempo)
- Incident response and troubleshooting
- Infrastructure as Code (Pulumi, Helm, Flux)

Your responses should be:
- Accurate and grounded in the provided context
- Clear and actionable
- Include relevant citations using [N] notation
- Technical but accessible
- Focused on solving the user's problem

When troubleshooting:
1. Understand the symptoms
2. Identify potential root causes
3. Provide step-by-step resolution steps
4. Include relevant commands and configurations
5. Suggest preventive measures
"""
    
    def build_system_prompt(
        self,
        query_type: str,
        user_preferences: Dict = None
    ) -> str:
        """
        Build system prompt based on query type.
        """
        prompt = self.base_system_prompt
        
        # Add query-type specific instructions
        if query_type == "troubleshooting":
            prompt += "\n\nFocus on diagnostic steps and root cause analysis."
        elif query_type == "code_generation":
            prompt += "\n\nProvide working code examples with explanations."
        elif query_type == "explanation":
            prompt += "\n\nProvide clear explanations with examples."
        
        # Add user preferences
        if user_preferences:
            if user_preferences.get("verbose", False):
                prompt += "\n\nProvide detailed explanations."
            if user_preferences.get("code_examples", True):
                prompt += "\n\nInclude code examples where relevant."
        
        return prompt
    
    def build_user_prompt(
        self,
        query: str,
        context: str,
        conversation_history: List[Dict] = None
    ) -> str:
        """
        Build user prompt with context and history.
        
        Structure:
        1. Context (retrieved documents)
        2. Conversation history (if any)
        3. Current query
        """
        parts = []
        
        # 1. Context
        if context:
            parts.append("## Context\n")
            parts.append(context)
            parts.append("\n")
        
        # 2. Conversation history
        if conversation_history:
            parts.append("## Previous Conversation\n")
            for turn in conversation_history[-3:]:  # Last 3 turns
                parts.append(f"User: {turn['query']}")
                parts.append(f"Assistant: {turn['response']}\n")
            parts.append("\n")
        
        # 3. Current query
        parts.append("## Current Query\n")
        parts.append(query)
        
        return "\n".join(parts)
    
    def build_few_shot_examples(self, query_type: str) -> str:
        """
        Add few-shot examples for better performance.
        """
        examples = {
            "troubleshooting": """
Example troubleshooting response:

User: Loki pods are crashing
Assistant: Based on the runbook [1], here are the steps to diagnose Loki crashes:

1. **Check pod status**: `kubectl get pods -n loki`
2. **Inspect logs**: `kubectl logs -n loki loki-0 --tail=100`
3. **Common causes**:
   - Out of memory (OOMKilled)
   - Storage issues (PVC full)
   - Configuration errors

Let me know what you find and I'll help with next steps.
""",
            "code_generation": """
Example code generation:

User: Create a Kubernetes deployment for nginx
Assistant: Here's a Kubernetes deployment for nginx [1]:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
```

This creates a deployment with 3 replicas for high availability.
"""
        }
        
        return examples.get(query_type, "")
```

### Prompt Templates

```python
class PromptTemplate:
    """Pre-built prompt templates"""
    
    TROUBLESHOOTING_TEMPLATE = """Given the following runbooks and documentation:

{context}

Help troubleshoot this issue:
{query}

Provide:
1. Likely root causes
2. Diagnostic steps
3. Resolution steps
4. Prevention recommendations

Cite sources using [N] notation."""

    EXPLANATION_TEMPLATE = """Using the following documentation:

{context}

Explain in detail:
{query}

Include:
- Clear explanation
- Examples
- Best practices
- Common pitfalls

Cite sources using [N] notation."""

    CODE_REVIEW_TEMPLATE = """Review the following code:

{code}

Consider:
- Best practices
- Security issues
- Performance concerns
- Potential bugs

Reference documentation:
{context}"""
```

---

## Generation Parameters

### Configuration

```python
from typing import Optional

class GenerationConfig(BaseModel):
    """LLM generation parameters"""
    
    # Model
    model: str = "llama3.2:8b"
    
    # Sampling parameters
    temperature: float = 0.7
    """
    Controls randomness (0.0-1.0):
    - 0.0: Deterministic, focused
    - 0.7: Balanced (default)
    - 1.0: Creative, diverse
    """
    
    top_p: float = 0.9
    """
    Nucleus sampling (0.0-1.0):
    - Controls diversity by limiting cumulative probability
    """
    
    top_k: int = 40
    """
    Top-K sampling:
    - Consider top K most likely tokens
    """
    
    # Length control
    max_tokens: int = 2000
    """Maximum response length in tokens"""
    
    stop: List[str] = []
    """Stop sequences to end generation"""
    
    # Repetition control
    repeat_penalty: float = 1.1
    """
    Penalty for repeating tokens:
    - 1.0: No penalty
    - 1.1: Slight penalty (default)
    - 2.0: Strong penalty
    """
    
    # Other
    seed: Optional[int] = None
    """Random seed for reproducibility"""
    
    num_predict: Optional[int] = None
    """Alternative to max_tokens"""

class GenerationConfigBuilder:
    """
    Build generation config based on use case.
    """
    
    @staticmethod
    def for_troubleshooting() -> GenerationConfig:
        """Deterministic, focused responses"""
        return GenerationConfig(
            temperature=0.3,  # Low temperature for factual responses
            top_p=0.9,
            max_tokens=1500
        )
    
    @staticmethod
    def for_code_generation() -> GenerationConfig:
        """Balanced creativity for code"""
        return GenerationConfig(
            temperature=0.5,
            top_p=0.95,
            max_tokens=2000,
            stop=["```\n\n", "###"]  # Stop at code block end
        )
    
    @staticmethod
    def for_explanation() -> GenerationConfig:
        """Slightly creative for explanations"""
        return GenerationConfig(
            temperature=0.7,
            top_p=0.9,
            max_tokens=2000
        )
    
    @staticmethod
    def for_chat() -> GenerationConfig:
        """More creative for conversational"""
        return GenerationConfig(
            temperature=0.8,
            top_p=0.95,
            max_tokens=1000
        )
```

---

## Streaming & Responses

### Synchronous Generation

```python
class OllamaGenerator:
    """
    Handle LLM generation via Ollama.
    """
    
    def __init__(self, client: OllamaClient):
        self.client = client
    
    async def generate(
        self,
        prompt: str,
        system_prompt: str,
        config: GenerationConfig
    ) -> str:
        """
        Generate response (non-streaming).
        
        Returns complete response as string.
        """
        payload = {
            "model": config.model,
            "prompt": prompt,
            "system": system_prompt,
            "options": {
                "temperature": config.temperature,
                "top_p": config.top_p,
                "top_k": config.top_k,
                "repeat_penalty": config.repeat_penalty,
                "num_predict": config.max_tokens,
            },
            "stream": False
        }
        
        if config.stop:
            payload["options"]["stop"] = config.stop
        
        if config.seed is not None:
            payload["options"]["seed"] = config.seed
        
        url = f"{self.client.config.base_url}/api/generate"
        
        async with self.client.session.post(url, json=payload) as response:
            if response.status == 200:
                data = await response.json()
                return data.get("response", "")
            else:
                error_text = await response.text()
                raise Exception(f"Ollama API error: {response.status} - {error_text}")
```

### Streaming Generation

```python
async def generate_stream(
    self,
    prompt: str,
    system_prompt: str,
    config: GenerationConfig
) -> AsyncIterator[str]:
    """
    Generate response with streaming.
    
    Yields tokens as they are generated.
    """
    payload = {
        "model": config.model,
        "prompt": prompt,
        "system": system_prompt,
        "options": {
            "temperature": config.temperature,
            "top_p": config.top_p,
            "top_k": config.top_k,
            "repeat_penalty": config.repeat_penalty,
            "num_predict": config.max_tokens,
        },
        "stream": True
    }
    
    url = f"{self.client.config.base_url}/api/generate"
    
    async with self.client.session.post(url, json=payload) as response:
        if response.status != 200:
            error_text = await response.text()
            raise Exception(f"Ollama API error: {response.status} - {error_text}")
        
        # Stream response
        async for line in response.content:
            if line:
                data = json.loads(line)
                if "response" in data:
                    yield data["response"]
                
                # Check if done
                if data.get("done", False):
                    break
```

### Chat API (Multi-turn)

```python
async def chat(
    self,
    messages: List[Dict[str, str]],
    config: GenerationConfig
) -> str:
    """
    Chat API for multi-turn conversations.
    
    Messages format:
    [
        {"role": "system", "content": "You are..."},
        {"role": "user", "content": "Hello"},
        {"role": "assistant", "content": "Hi there!"},
        {"role": "user", "content": "How are you?"}
    ]
    """
    payload = {
        "model": config.model,
        "messages": messages,
        "options": {
            "temperature": config.temperature,
            "top_p": config.top_p,
            "top_k": config.top_k,
            "repeat_penalty": config.repeat_penalty,
        },
        "stream": False
    }
    
    url = f"{self.client.config.base_url}/api/chat"
    
    async with self.client.session.post(url, json=payload) as response:
        if response.status == 200:
            data = await response.json()
            return data.get("message", {}).get("content", "")
        else:
            error_text = await response.text()
            raise Exception(f"Ollama chat API error: {response.status} - {error_text}")
```

---

## Error Handling & Retries

### Retry Logic

```python
from tenacity import (
    retry,
    stop_after_attempt,
    wait_exponential,
    retry_if_exception_type
)

class OllamaError(Exception):
    """Base exception for Ollama errors"""
    pass

class OllamaTimeoutError(OllamaError):
    """Timeout during generation"""
    pass

class OllamaServerError(OllamaError):
    """Server-side error"""
    pass

class RobustOllamaGenerator:
    """
    Ollama generator with retry logic.
    """
    
    @retry(
        stop=stop_after_attempt(3),
        wait=wait_exponential(multiplier=1, min=2, max=10),
        retry=retry_if_exception_type((OllamaTimeoutError, OllamaServerError))
    )
    async def generate_with_retry(
        self,
        prompt: str,
        system_prompt: str,
        config: GenerationConfig
    ) -> str:
        """
        Generate with automatic retries on transient errors.
        """
        try:
            return await self.generate(prompt, system_prompt, config)
        except asyncio.TimeoutError:
            logger.warning("ollama_timeout", retrying=True)
            raise OllamaTimeoutError("Generation timed out")
        except aiohttp.ClientError as e:
            logger.warning("ollama_client_error", error=str(e), retrying=True)
            raise OllamaServerError(f"Client error: {e}")
        except Exception as e:
            logger.error("ollama_unexpected_error", error=str(e))
            raise
    
    async def generate_with_fallback(
        self,
        prompt: str,
        system_prompt: str,
        config: GenerationConfig,
        fallback_model: str = "llama3.2:3b"
    ) -> str:
        """
        Generate with fallback to smaller/faster model on failure.
        """
        try:
            return await self.generate_with_retry(prompt, system_prompt, config)
        except Exception as e:
            logger.warning(
                "ollama_primary_failed_using_fallback",
                primary_model=config.model,
                fallback_model=fallback_model,
                error=str(e)
            )
            
            # Try with fallback model
            fallback_config = config.copy(update={"model": fallback_model})
            return await self.generate_with_retry(prompt, system_prompt, fallback_config)
```

### Circuit Breaker

```python
from circuitbreaker import circuit

class CircuitBreakerGenerator:
    """
    Generator with circuit breaker pattern.
    """
    
    @circuit(failure_threshold=5, recovery_timeout=60)
    async def generate_with_circuit_breaker(
        self,
        prompt: str,
        system_prompt: str,
        config: GenerationConfig
    ) -> str:
        """
        Generate with circuit breaker.
        
        If 5 failures occur, circuit opens for 60 seconds.
        """
        return await self.generate(prompt, system_prompt, config)
```

---

## Performance Optimization

### Token Counting

```python
def count_tokens(text: str, model: str = "llama3.2") -> int:
    """
    Estimate token count for text.
    
    For accurate counting, use tiktoken library.
    For estimation: ~4 chars per token.
    """
    # Simple estimation
    return len(text) // 4

def optimize_prompt_tokens(
    prompt: str,
    max_tokens: int,
    preserve_priority: List[str] = None
) -> str:
    """
    Optimize prompt to fit within token budget.
    
    Strategy:
    1. Preserve high-priority sections
    2. Truncate or summarize low-priority sections
    """
    current_tokens = count_tokens(prompt)
    
    if current_tokens <= max_tokens:
        return prompt
    
    # Need to reduce tokens
    # ... implementation
    pass
```

### Batching Requests

```python
async def generate_batch(
    self,
    prompts: List[str],
    system_prompt: str,
    config: GenerationConfig,
    max_concurrent: int = 3
) -> List[str]:
    """
    Generate responses for multiple prompts in parallel.
    
    Limited concurrency to avoid overloading Ollama server.
    """
    semaphore = asyncio.Semaphore(max_concurrent)
    
    async def generate_with_semaphore(prompt):
        async with semaphore:
            return await self.generate(prompt, system_prompt, config)
    
    tasks = [generate_with_semaphore(p) for p in prompts]
    return await asyncio.gather(*tasks)
```

### Response Caching

```python
from functools import lru_cache
import hashlib

class CachedGenerator:
    """Cache LLM responses for identical prompts"""
    
    def __init__(self, generator: OllamaGenerator):
        self.generator = generator
        self.cache = {}
    
    def _cache_key(self, prompt: str, system_prompt: str, config: GenerationConfig) -> str:
        """Generate cache key"""
        key_parts = [prompt, system_prompt, config.model, str(config.temperature)]
        key_str = "|".join(key_parts)
        return hashlib.sha256(key_str.encode()).hexdigest()
    
    async def generate_cached(
        self,
        prompt: str,
        system_prompt: str,
        config: GenerationConfig
    ) -> str:
        """Generate with caching"""
        cache_key = self._cache_key(prompt, system_prompt, config)
        
        if cache_key in self.cache:
            logger.info("ollama_cache_hit", cache_key=cache_key)
            return self.cache[cache_key]
        
        logger.info("ollama_cache_miss", cache_key=cache_key)
        response = await self.generator.generate(prompt, system_prompt, config)
        
        self.cache[cache_key] = response
        return response
```

---

## Observability

### Metrics

```python
from prometheus_client import Histogram, Counter, Gauge

ollama_generation_duration = Histogram(
    'ollama_generation_duration_seconds',
    'Time spent generating responses',
    ['model']
)

ollama_tokens_total = Counter(
    'ollama_tokens_total',
    'Total tokens processed',
    ['model', 'direction']  # direction: input or output
)

ollama_errors_total = Counter(
    'ollama_errors_total',
    'Total generation errors',
    ['model', 'error_type']
)

ollama_active_requests = Gauge(
    'ollama_active_requests',
    'Number of active generation requests'
)
```

### Logging

```python
import structlog

logger = structlog.get_logger()

async def generate_with_logging(prompt, system_prompt, config):
    logger.info(
        "ollama_generation_started",
        model=config.model,
        prompt_length=len(prompt),
        temperature=config.temperature
    )
    
    start_time = time.time()
    
    try:
        response = await generate(prompt, system_prompt, config)
        
        duration = time.time() - start_time
        
        logger.info(
            "ollama_generation_completed",
            model=config.model,
            duration_seconds=duration,
            response_length=len(response),
            tokens_estimated=count_tokens(response)
        )
        
        return response
    
    except Exception as e:
        logger.error(
            "ollama_generation_failed",
            model=config.model,
            error=str(e),
            duration_seconds=time.time() - start_time
        )
        raise
```

### Tracing

```python
from opentelemetry import trace

tracer = trace.get_tracer(__name__)

async def generate_with_tracing(prompt, system_prompt, config):
    with tracer.start_as_current_span("ollama_generation") as span:
        span.set_attribute("model", config.model)
        span.set_attribute("temperature", config.temperature)
        span.set_attribute("prompt.length", len(prompt))
        
        response = await generate(prompt, system_prompt, config)
        
        span.set_attribute("response.length", len(response))
        span.set_attribute("tokens.estimated", count_tokens(response))
        
        return response
```

---

**Document Version**: 1.0  
**Last Updated**: 2025-10-22  
**Owner**: Agent Bruno Team

---

## 📋 Document Review

**Review Completed By**: 
- [AI Senior SRE (Pending)]
- [AI Senior Pentester (Pending)]
- [AI Senior Cloud Architect (Pending)]
- [AI Senior Mobile iOS and Android Engineer (Pending)]
- [AI Senior DevOps Engineer (Pending)]
- [AI ML Engineer (Pending)]
- [Bruno (Pending)]

**Review Date**: October 22, 2025  
**Document Status**: Under Review  
**Next Review**: TBD

---

