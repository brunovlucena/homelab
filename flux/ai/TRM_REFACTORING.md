# TRM Refactoring Summary

## Overview

All homelab agents have been refactored to use **TRM (Tiny Recursive Model)** with built-in reflection capabilities instead of Ollama. This provides:

- **Built-in reflection**: Models self-reflect and refine their answers automatically
- **Better reasoning**: Recursive self-improvement mechanism
- **Smaller models**: ~7M parameters vs larger Ollama models
- **Unified interface**: All agents use the same TRM client from shared-lib

## What Was Changed

### 1. Created Shared TRM Client Library

**Location**: `flux/ai/shared-lib/agent_trm/`

- `client.py`: TRMClient with built-in reflection
- `types.py`: TRMRequest, TRMResponse, ReflectionStep types
- `__init__.py`: Public API

**Features**:
- Supports Hugging Face models (e.g., `ainz/tiny-recursive-model`)
- Can use Hugging Face Inference API or local models
- Automatic reflection with configurable steps
- Fallback to Ollama if TRM unavailable

### 2. Refactored Agents

#### ✅ agent-bruno
- **File**: `src/chatbot/handler.py`
- **Changes**: 
  - Added TRMClient initialization
  - Replaced `_query_ollama` with `_query_trm` (with Ollama fallback)
  - Updated health checks

#### ✅ agent-sre
- **File**: `src/report_generator/generator.py`
- **Changes**:
  - Added TRMClient to ReportGenerator
  - Added `_generate_with_trm` method
  - Falls back to legacy MLX/Ollama/Anthropic if TRM fails

#### ✅ agent-store-multibrands
- **File**: `src/ai_seller/handler.py`
- **Changes**:
  - Added TRMClient to AISeller
  - Added `_query_trm` method
  - Falls back to Ollama if TRM fails

### 3. Updated Shared Library

**File**: `shared-lib/pyproject.toml`
- Added `transformers>=4.35.0` and `torch>=2.0.0` dependencies
- Added `agent_trm*` to package includes

## Configuration

### Environment Variables

All agents support these TRM configuration variables:

```bash
# TRM Model Configuration
TRM_MODEL_NAME=ainz/tiny-recursive-model  # Hugging Face model name
TRM_USE_HF_API=false                      # Use HF Inference API vs local model
HF_API_TOKEN=your_token                   # Required if using HF API

# Legacy Ollama (fallback only)
OLLAMA_URL=http://ollama-native.ollama.svc.cluster.local:11434
OLLAMA_MODEL=llama3.2:3b
```

### Model Options

Available TRM models on Hugging Face:
- `ainz/tiny-recursive-model` (default) - 7M params, 3 layers, 8 recursive loops
- `wtfmahe/Samsung-TRM` - Samsung implementation
- `ordlibrary/x402` - Fine-tuned for Solana Q&A (3.5M params)

## How It Works

### Reflection Mechanism

1. **Initial Answer**: TRM generates first response
2. **Reflection Loop** (up to `max_reflection_steps`):
   - Reflect on current answer (identify issues, improvements)
   - Refine answer based on reflection
   - Calculate confidence and improvement score
3. **Early Stopping**: Stops if confidence > 0.85 or no improvement

### Example Flow

```python
from agent_trm import TRMClient, TRMRequest, ReflectionMode

client = TRMClient(
    model_name="ainz/tiny-recursive-model",
    use_hf_api=False,
)

request = TRMRequest(
    prompt="What is the health status of Loki?",
    max_reflection_steps=3,
    reflection_mode=ReflectionMode.AUTO,
)

response = await client.generate(request)
# response.answer: Final refined answer
# response.reflection_steps: Number of reflection cycles
# response.confidence: Final confidence score
# response.reflection_trace: Step-by-step reflection process
```

## Remaining Work

### Pending Refactoring

- [ ] agent-restaurant
- [ ] agent-screenshot
- [ ] agent-reasoning (update to use real TRM implementation)

### Dependencies

All agents need to ensure `shared-lib` is installed with TRM support:

```bash
# In agent requirements.txt or Dockerfile
pip install -e /app/shared-lib
```

The shared-lib includes:
- `transformers>=4.35.0`
- `torch>=2.0.0`
- `httpx>=0.25.0` (for HF API)

## Benefits

1. **Better Reasoning**: Built-in reflection improves answer quality
2. **Smaller Models**: 7M params vs 3B+ for Ollama models
3. **Unified Interface**: All agents use same TRM client
4. **Graceful Fallback**: Falls back to Ollama if TRM unavailable
5. **Observability**: Full metrics and tracing support

## Testing

To test TRM in an agent:

```bash
# Set environment variables
export TRM_MODEL_NAME=ainz/tiny-recursive-model
export TRM_USE_HF_API=false  # or true for HF API

# Run agent
python -m chatbot.main
```

Check logs for:
- `trm_client_initialized`
- `trm_generation_completed`
- `trm_reflection_steps`

## Migration Notes

- **Backward Compatible**: All agents maintain Ollama fallback
- **No Breaking Changes**: Existing Ollama config still works
- **Gradual Migration**: Can enable TRM per-agent via env vars
- **Zero Downtime**: Fallback ensures agents keep working
