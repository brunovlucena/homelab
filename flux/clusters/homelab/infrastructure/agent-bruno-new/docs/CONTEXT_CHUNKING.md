# Context Assembly & Chunking - Agent Bruno

**[← Back to Architecture](ARCHITECTURE.md)** | **[Main README](../README.md)**

---

## Table of Contents
1. [Overview](#overview)
2. [Context Selection](#context-selection)
3. [Chunk Boundary Optimization](#chunk-boundary-optimization)
4. [Context Window Management](#context-window-management)
5. [Metadata Enrichment](#metadata-enrichment)
6. [Context Formatting](#context-formatting)
7. [Advanced Strategies](#advanced-strategies)
8. [Performance & Optimization](#performance--optimization)
9. [Observability](#observability)

---

## Overview

Context Assembly & Chunking is the stage that prepares retrieved information for LLM consumption by selecting, optimizing, and formatting chunks to fit within the context window while maximizing information density.

### Goals
- 📦 **Select optimal chunks** - Choose the most relevant information
- ✂️ **Optimize boundaries** - Ensure chunks are complete and coherent
- 📏 **Manage context window** - Fit within LLM token limits
- 🏷️ **Enrich metadata** - Add source citations and relevance scores
- 📝 **Format for LLM** - Structure context for optimal LLM comprehension

### Architecture Position

```
Query Processing
    ↓
Semantic + Keyword Search
    ↓
Fusion & Re-ranking
    ↓
┌─────────────────────────────────────────┐
│    Context Assembly & Chunking          │  ← YOU ARE HERE
│  • Select Top-N Results                 │
│  • Optimize Chunk Boundaries            │
│  • Manage Context Window                │
│  • Add Metadata & Citations             │
│  • Format for LLM                       │
└─────────────────────────────────────────┘
    ↓
LLM Generation (Ollama)
    ↓
Response Post-processing
```

---

## Context Selection

### Top-N Selection

```python
from typing import List, Dict, Optional
from pydantic import BaseModel, Field, field_validator
from enum import Enum
from datetime import datetime

# ✅ Use Pydantic for validation
class ContextChunk(BaseModel):
    """
    Single chunk of context with automatic validation.
    
    Pydantic ensures:
    - All required fields are present
    - Types are correct
    - Constraints are enforced
    - Easy serialization/deserialization
    """
    chunk_id: str = Field(..., min_length=1, description="Unique chunk identifier")
    content: str = Field(..., min_length=10, description="Chunk content text")
    metadata: Dict = Field(default_factory=dict, description="Chunk metadata")
    relevance_score: float = Field(..., ge=0.0, le=1.0, description="Relevance score (0-1)")
    token_count: int = Field(..., ge=0, description="Estimated token count")
    source: str = Field(..., min_length=1, description="Source document path")
    timestamp: datetime = Field(default_factory=datetime.utcnow, description="Retrieval timestamp")
    
    @field_validator('content')
    @classmethod
    def validate_content_not_empty(cls, v: str) -> str:
        """Ensure content is not just whitespace."""
        if not v.strip():
            raise ValueError('Content cannot be empty or whitespace only')
        return v
    
    @field_validator('metadata')
    @classmethod
    def validate_metadata_has_source_info(cls, v: Dict) -> Dict:
        """Ensure metadata contains essential source information."""
        if 'source_type' not in v and 'source_name' not in v:
            raise ValueError('Metadata must contain source_type or source_name')
        return v
    
    class Config:
        # Allow assignment after initialization
        validate_assignment = True
        # Serialize datetime as ISO string
        json_encoders = {
            datetime: lambda v: v.isoformat()
        }

class ContextSelector:
    """
    Select optimal chunks for LLM context.
    """
    
    def __init__(self, max_context_tokens: int = 4000):
        """
        Args:
            max_context_tokens: Maximum tokens to use for retrieved context
                               (leaves room for system prompt + query + response)
        """
        self.max_context_tokens = max_context_tokens
    
    def select_top_n(
        self,
        ranked_results: List['FusedResult'],
        target_count: int = 5,
        min_relevance: float = 0.1
    ) -> List[ContextChunk]:
        """
        Select top N chunks based on relevance and token budget.
        
        Strategy:
        1. Take top N results by score
        2. Filter by minimum relevance threshold
        3. Ensure total tokens <= max_context_tokens
        4. Prioritize diverse sources
        """
        selected_chunks = []
        total_tokens = 0
        seen_sources = set()
        
        for result in ranked_results:
            # Skip if below relevance threshold
            if result.rrf_score < min_relevance:
                continue
            
            # Calculate token count
            token_count = self._estimate_tokens(result.content)
            
            # Check if adding this chunk would exceed budget
            if total_tokens + token_count > self.max_context_tokens:
                # Try to fit smaller chunks
                continue
            
            # Prefer diverse sources
            source = result.metadata.get("source_path", "")
            diversity_bonus = 0.0 if source in seen_sources else 0.05
            adjusted_score = result.rrf_score + diversity_bonus
            
            # Create context chunk (auto-validated by Pydantic)
            try:
                chunk = ContextChunk(
                    chunk_id=result.chunk_id,
                    content=result.content,
                    metadata=result.metadata,
                    relevance_score=adjusted_score,
                    token_count=token_count,
                    source=source,
                    timestamp=datetime.fromisoformat(result.metadata.get("last_updated", datetime.utcnow().isoformat()))
                )
            except ValidationError as e:
                # Log validation error
                logger.error(
                    "Chunk validation failed",
                    extra={"chunk_id": result.chunk_id, "error": str(e)}
                )
                continue  # Skip invalid chunks
            
            selected_chunks.append(chunk)
            total_tokens += token_count
            seen_sources.add(source)
            
            # Stop if we have enough chunks
            if len(selected_chunks) >= target_count:
                break
        
        return selected_chunks
    
    def _estimate_tokens(self, text: str) -> int:
        """
        Estimate token count (fast approximation).
        
        More accurate: use tiktoken library
        Approximation: ~4 chars per token for English
        """
        return len(text) // 4
    
    def select_adaptive(
        self,
        ranked_results: List[FusedResult],
        query_complexity: str,
        user_intent: str
    ) -> List[ContextChunk]:
        """
        Adaptive selection based on query characteristics.
        
        Simple queries: fewer chunks (1-3)
        Complex queries: more chunks (5-10)
        Troubleshooting: prioritize runbooks
        """
        if query_complexity == "simple":
            target_count = 3
            max_tokens = 2000
        elif query_complexity == "medium":
            target_count = 5
            max_tokens = 4000
        else:  # complex
            target_count = 10
            max_tokens = 6000
        
        # Filter by intent
        if user_intent == "lookup_runbook":
            # Only select runbooks
            ranked_results = [
                r for r in ranked_results
                if r.metadata.get("source_type") == "runbook"
            ]
        
        # Use adaptive max tokens
        original_max = self.max_context_tokens
        self.max_context_tokens = max_tokens
        
        selected = self.select_top_n(ranked_results, target_count=target_count)
        
        # Restore original
        self.max_context_tokens = original_max
        
        return selected
```

---

## Chunk Boundary Optimization

### Why Optimize Boundaries?

Retrieved chunks might have poor boundaries that:
- Cut off mid-sentence
- Split code blocks
- Separate related concepts
- Lose critical context

### Boundary Optimization Strategies

```python
import re
from typing import List, Tuple

class ChunkBoundaryOptimizer:
    """
    Optimize chunk boundaries for coherence.
    """
    
    def __init__(self):
        # Sentence endings
        self.sentence_endings = r'[.!?]\s+'
        
        # Code block markers
        self.code_markers = ['```', '~~~']
        
        # Section headers (Markdown)
        self.header_pattern = r'^#{1,6}\s+'
    
    def optimize_chunk(
        self,
        chunk: ContextChunk,
        full_document: str = None
    ) -> ContextChunk:
        """
        Optimize a single chunk's boundaries.
        
        Strategies:
        1. Extend to complete sentences
        2. Include full code blocks
        3. Respect section boundaries
        4. Add surrounding context
        """
        content = chunk.content
        
        # 1. Complete sentences
        content = self._complete_sentences(content)
        
        # 2. Complete code blocks
        content = self._complete_code_blocks(content)
        
        # 3. Add section headers if truncated
        if full_document:
            content = self._add_section_header(content, full_document)
        
        # Update chunk
        chunk.content = content
        chunk.token_count = len(content) // 4
        
        return chunk
    
    def _complete_sentences(self, text: str) -> str:
        """
        Ensure chunk ends at sentence boundary.
        """
        # If text doesn't end with sentence terminator, find the last one
        if not re.search(r'[.!?]\s*$', text):
            # Find last sentence ending
            matches = list(re.finditer(self.sentence_endings, text))
            if matches:
                last_match = matches[-1]
                # Truncate to last complete sentence
                text = text[:last_match.end()].strip()
        
        # Similarly, ensure it starts at sentence beginning
        if not re.match(r'^[A-Z]', text.strip()):
            # Find first sentence start
            match = re.search(r'\.\s+([A-Z])', text)
            if match:
                text = text[match.start() + 2:].strip()
        
        return text
    
    def _complete_code_blocks(self, text: str) -> str:
        """
        Ensure code blocks are complete (not split).
        """
        for marker in self.code_markers:
            count = text.count(marker)
            
            # If odd number, code block is incomplete
            if count % 2 != 0:
                # Try to find closing marker in context
                # For now, remove incomplete code block
                last_marker_pos = text.rfind(marker)
                text = text[:last_marker_pos].strip()
        
        return text
    
    def _add_section_header(self, chunk_text: str, full_doc: str) -> str:
        """
        Add section header if chunk is from middle of section.
        """
        # Find where this chunk appears in full doc
        chunk_pos = full_doc.find(chunk_text[:100])  # Use first 100 chars
        
        if chunk_pos > 0:
            # Look backwards for section header
            preceding_text = full_doc[:chunk_pos]
            
            # Find last header
            headers = list(re.finditer(self.header_pattern, preceding_text, re.MULTILINE))
            if headers:
                last_header = headers[-1]
                header_line_start = preceding_text.rfind('\n', 0, last_header.start()) + 1
                header_line_end = preceding_text.find('\n', last_header.end())
                header_text = preceding_text[header_line_start:header_line_end]
                
                # Prepend header to chunk
                chunk_text = f"{header_text}\n\n{chunk_text}"
        
        return chunk_text
    
    def merge_overlapping_chunks(
        self,
        chunks: List[ContextChunk]
    ) -> List[ContextChunk]:
        """
        Merge chunks that have significant overlap.
        
        This can happen when:
        - Both semantic and keyword search return same document
        - Adjacent chunks from same document
        """
        if not chunks:
            return chunks
        
        merged = []
        current = chunks[0]
        
        for next_chunk in chunks[1:]:
            # Check if chunks are from same document
            if current.source == next_chunk.source:
                # Check for overlap
                overlap = self._calculate_overlap(current.content, next_chunk.content)
                
                if overlap > 0.5:  # >50% overlap
                    # Merge chunks
                    current = self._merge_chunks(current, next_chunk)
                else:
                    # No overlap, save current and move to next
                    merged.append(current)
                    current = next_chunk
            else:
                # Different source, save current and move to next
                merged.append(current)
                current = next_chunk
        
        # Add last chunk
        merged.append(current)
        
        return merged
    
    def _calculate_overlap(self, text1: str, text2: str) -> float:
        """Calculate text overlap ratio"""
        # Simple word-based overlap
        words1 = set(text1.lower().split())
        words2 = set(text2.lower().split())
        
        intersection = words1 & words2
        union = words1 | words2
        
        if not union:
            return 0.0
        
        return len(intersection) / len(union)
    
    def _merge_chunks(
        self,
        chunk1: ContextChunk,
        chunk2: ContextChunk
    ) -> ContextChunk:
        """Merge two overlapping chunks"""
        # Combine content (deduplicate)
        combined = chunk1.content
        
        # Find where chunk2 starts in chunk1
        overlap_start = chunk1.content.find(chunk2.content[:50])
        
        if overlap_start == -1:
            # No direct overlap, concatenate
            combined = f"{chunk1.content}\n\n{chunk2.content}"
        else:
            # Merge at overlap point
            combined = chunk1.content + chunk2.content[50:]
        
        # Combine metadata
        merged_metadata = {**chunk1.metadata, **chunk2.metadata}
        
        return ContextChunk(
            chunk_id=f"{chunk1.chunk_id}_merged_{chunk2.chunk_id}",
            content=combined,
            metadata=merged_metadata,
            relevance_score=max(chunk1.relevance_score, chunk2.relevance_score),
            token_count=len(combined) // 4,
            source=chunk1.source,
            timestamp=chunk1.timestamp
        )
```

### Pydantic AI Integration for Context Assembly

When using Pydantic AI agents, return validated context as structured output:

```python
from pydantic_ai import Agent, RunContext
from pydantic import BaseModel

class AssembledContext(BaseModel):
    """Validated context assembly result."""
    chunks: List[ContextChunk] = Field(..., min_length=1, max_length=10)
    total_tokens: int = Field(..., ge=0, le=8000)
    sources: List[str] = Field(..., description="Unique source list")
    formatted_context: str = Field(..., description="LLM-ready formatted context")
    
    @field_validator('chunks')
    @classmethod
    def validate_chunks_within_budget(cls, v: List[ContextChunk]) -> List[ContextChunk]:
        """Ensure total tokens don't exceed budget."""
        total = sum(chunk.token_count for chunk in v)
        if total > 8000:
            raise ValueError(f'Total tokens {total} exceeds budget of 8000')
        return v
    
    @field_validator('sources')
    @classmethod  
    def validate_unique_sources(cls, v: List[str]) -> List[str]:
        """Ensure sources are unique."""
        if len(v) != len(set(v)):
            raise ValueError('Sources must be unique')
        return v

# Use as agent tool return type
@agent.tool
async def assemble_context_for_query(
    ctx: RunContext[AgentDependencies],
    query: str,
    top_k: int = 5
) -> AssembledContext:
    """
    Assemble and validate context for query.
    
    Returns validated AssembledContext that guarantees:
    - Chunks are valid and complete
    - Token budget is respected
    - Sources are properly cited
    """
    # Retrieve results
    results = await retrieve_hybrid(ctx, query, top_k=20)
    
    # Select and optimize chunks
    selector = ContextSelector(max_context_tokens=4000)
    chunks = selector.select_top_n(results, target_count=top_k)
    
    # Optimize boundaries
    optimizer = ChunkBoundaryOptimizer()
    optimized_chunks = [optimizer.optimize_chunk(c) for c in chunks]
    
    # Format for LLM
    formatter = ContextFormatter()
    formatted = formatter.format_for_llm(optimized_chunks, format_style="structured")
    
    # Extract unique sources
    sources = list(set(chunk.source for chunk in optimized_chunks))
    
    # Return validated context (Pydantic auto-validates)
    return AssembledContext(
        chunks=optimized_chunks,
        total_tokens=sum(c.token_count for c in optimized_chunks),
        sources=sources,
        formatted_context=formatted
    )
```

**Benefits of Pydantic Validation**:
- ✅ Invalid contexts rejected before reaching LLM
- ✅ Token budget violations caught early
- ✅ Type safety across the pipeline
- ✅ Automatic serialization for logging/caching
- ✅ Clear error messages when validation fails

---

## Context Window Management

### Token Budget Allocation

```python
class ContextWindowManager:
    """
    Manage LLM context window allocation.
    
    Typical breakdown for 8K context window:
    - System prompt: 500 tokens
    - User query: 200 tokens
    - Retrieved context: 4000 tokens
    - Conversation history: 1000 tokens
    - Response: 2000 tokens (reserved)
    - Buffer: 300 tokens
    """
    
    def __init__(
        self,
        total_context_window: int = 8192,
        system_prompt_tokens: int = 500,
        max_response_tokens: int = 2000,
        buffer_tokens: int = 300
    ):
        self.total_context_window = total_context_window
        self.system_prompt_tokens = system_prompt_tokens
        self.max_response_tokens = max_response_tokens
        self.buffer_tokens = buffer_tokens
        
        # Calculate available tokens for context
        self.available_for_context = (
            total_context_window
            - system_prompt_tokens
            - max_response_tokens
            - buffer_tokens
        )
    
    def allocate_budget(
        self,
        query_tokens: int,
        conversation_history_tokens: int = 0,
        episodic_memory_tokens: int = 0
    ) -> Dict[str, int]:
        """
        Allocate token budget across components.
        
        Returns:
            Dict with token allocation for each component
        """
        remaining = self.available_for_context
        
        # 1. Query (required)
        remaining -= query_tokens
        
        # 2. Conversation history (if any)
        history_allocation = min(conversation_history_tokens, 1000)
        remaining -= history_allocation
        
        # 3. Episodic memory (if any)
        memory_allocation = min(episodic_memory_tokens, 500)
        remaining -= memory_allocation
        
        # 4. Remaining goes to retrieved context
        context_allocation = max(0, remaining)
        
        return {
            "system_prompt": self.system_prompt_tokens,
            "query": query_tokens,
            "conversation_history": history_allocation,
            "episodic_memory": memory_allocation,
            "retrieved_context": context_allocation,
            "response": self.max_response_tokens,
            "buffer": self.buffer_tokens,
        }
    
    def fit_chunks_to_budget(
        self,
        chunks: List[ContextChunk],
        token_budget: int
    ) -> List[ContextChunk]:
        """
        Fit chunks within token budget using greedy algorithm.
        
        Strategy:
        1. Sort by relevance score
        2. Add chunks until budget exhausted
        3. Optionally truncate last chunk
        """
        sorted_chunks = sorted(chunks, key=lambda c: c.relevance_score, reverse=True)
        
        selected = []
        total_tokens = 0
        
        for chunk in sorted_chunks:
            if total_tokens + chunk.token_count <= token_budget:
                selected.append(chunk)
                total_tokens += chunk.token_count
            else:
                # Check if we can fit a truncated version
                remaining_budget = token_budget - total_tokens
                if remaining_budget > 100:  # At least 100 tokens
                    truncated = self._truncate_chunk(chunk, remaining_budget)
                    selected.append(truncated)
                break
        
        return selected
    
    def _truncate_chunk(
        self,
        chunk: ContextChunk,
        max_tokens: int
    ) -> ContextChunk:
        """
        Truncate chunk to fit within token budget.
        """
        # Approximate characters from tokens
        max_chars = max_tokens * 4
        
        # Truncate content
        truncated_content = chunk.content[:max_chars]
        
        # Try to end at sentence boundary
        last_period = truncated_content.rfind('.')
        if last_period > max_chars * 0.8:  # At least 80% of content
            truncated_content = truncated_content[:last_period + 1]
        
        # Add ellipsis
        truncated_content += "..."
        
        # Create new chunk
        return ContextChunk(
            chunk_id=chunk.chunk_id + "_truncated",
            content=truncated_content,
            metadata={**chunk.metadata, "truncated": True},
            relevance_score=chunk.relevance_score * 0.9,  # Penalize truncation
            token_count=max_tokens,
            source=chunk.source,
            timestamp=chunk.timestamp
        )
```

---

## Metadata Enrichment

### Adding Context Metadata

```python
from datetime import datetime

class MetadataEnricher:
    """
    Enrich chunks with metadata for citations and traceability.
    """
    
    def enrich(self, chunks: List[ContextChunk]) -> List[ContextChunk]:
        """
        Add metadata to chunks:
        - Source citations
        - Relevance scores
        - Timestamps
        - Provenance
        """
        for i, chunk in enumerate(chunks):
            # Add position in context
            chunk.metadata["context_position"] = i + 1
            chunk.metadata["total_chunks"] = len(chunks)
            
            # Add retrieval timestamp
            chunk.metadata["retrieved_at"] = datetime.utcnow().isoformat()
            
            # Format citation
            chunk.metadata["citation"] = self._format_citation(chunk)
            
            # Add confidence level
            chunk.metadata["confidence"] = self._calculate_confidence(chunk)
        
        return chunks
    
    def _format_citation(self, chunk: ContextChunk) -> str:
        """
        Format citation string for LLM.
        
        Example: [1] Runbook: Loki Crashes (updated: 2025-10-15)
        """
        source_type = chunk.metadata.get("source_type", "document")
        source_name = chunk.metadata.get("source_name", "Unknown")
        position = chunk.metadata.get("context_position", 0)
        
        # Get last updated date
        last_updated = chunk.metadata.get("last_updated", "")
        if last_updated:
            date_str = f" (updated: {last_updated[:10]})"
        else:
            date_str = ""
        
        return f"[{position}] {source_type.title()}: {source_name}{date_str}"
    
    def _calculate_confidence(self, chunk: ContextChunk) -> str:
        """
        Calculate confidence level: high, medium, low
        """
        score = chunk.relevance_score
        
        if score >= 0.8:
            return "high"
        elif score >= 0.5:
            return "medium"
        else:
            return "low"
```

---

## Context Formatting

### LLM-Optimized Formatting

```python
class ContextFormatter:
    """
    Format context for optimal LLM comprehension.
    """
    
    def format_for_llm(
        self,
        chunks: List[ContextChunk],
        format_style: str = "structured"
    ) -> str:
        """
        Format chunks for LLM consumption.
        
        Styles:
        - structured: Clear sections with citations
        - compact: Minimal formatting, max density
        - narrative: Natural language flow
        """
        if format_style == "structured":
            return self._format_structured(chunks)
        elif format_style == "compact":
            return self._format_compact(chunks)
        else:
            return self._format_narrative(chunks)
    
    def _format_structured(self, chunks: List[ContextChunk]) -> str:
        """
        Structured format with clear sections and citations.
        
        Example:
        ```
        ## Retrieved Context
        
        ### [1] Runbook: Loki Crashes (Relevance: High)
        Source: /runbooks/loki/crash-loop.md
        Last Updated: 2025-10-15
        
        [Content here...]
        
        ---
        
        ### [2] Documentation: Loki Configuration
        ...
        ```
        """
        parts = ["## Retrieved Context\n"]
        
        for chunk in chunks:
            # Header with citation
            citation = chunk.metadata.get("citation", "")
            confidence = chunk.metadata.get("confidence", "").title()
            
            parts.append(f"### {citation}")
            parts.append(f"**Relevance**: {confidence}")
            parts.append(f"**Source**: {chunk.source}")
            parts.append("")
            
            # Content
            parts.append(chunk.content)
            parts.append("\n---\n")
        
        return "\n".join(parts)
    
    def _format_compact(self, chunks: List[ContextChunk]) -> str:
        """
        Compact format for maximum token efficiency.
        
        Example:
        ```
        [1] Content from source 1...
        [2] Content from source 2...
        ```
        """
        parts = []
        
        for chunk in chunks:
            position = chunk.metadata.get("context_position", 0)
            parts.append(f"[{position}] {chunk.content}")
        
        return "\n\n".join(parts)
    
    def _format_narrative(self, chunks: List[ContextChunk]) -> str:
        """
        Narrative format that flows naturally.
        
        Example:
        ```
        Based on available documentation:
        
        Regarding Loki crashes, [1] indicates that...
        
        Additionally, [2] mentions that...
        ```
        """
        parts = ["Based on available documentation:\n"]
        
        for i, chunk in enumerate(chunks):
            position = chunk.metadata.get("context_position", 0)
            
            # Add transition
            if i == 0:
                transition = "According to"
            elif i == len(chunks) - 1:
                transition = "Finally,"
            else:
                transition = "Additionally,"
            
            # Format with citation reference
            parts.append(f"{transition} [{position}] {chunk.content}")
        
        return "\n\n".join(parts)
    
    def add_instructions(
        self,
        formatted_context: str,
        query_type: str
    ) -> str:
        """
        Add LLM instructions based on context.
        """
        instructions = {
            "troubleshooting": (
                "Use the following runbooks and documentation to help diagnose and resolve the issue. "
                "Cite sources using [N] notation."
            ),
            "explanation": (
                "Use the following documentation to provide a clear explanation. "
                "Include examples where relevant and cite sources using [N] notation."
            ),
            "question": (
                "Answer the question using the following context. "
                "Be concise and cite sources using [N] notation."
            )
        }
        
        instruction = instructions.get(query_type, "Use the following context to respond.")
        
        return f"{instruction}\n\n{formatted_context}"
```

---

## Advanced Strategies

### 1. Hierarchical Context Assembly

```python
class HierarchicalContextAssembler:
    """
    Assemble context in hierarchical fashion:
    - High-level summary first
    - Detailed sections below
    - Code examples last
    """
    
    def assemble(self, chunks: List[ContextChunk]) -> str:
        """
        Organize chunks hierarchically.
        """
        # Categorize chunks
        summaries = []
        details = []
        code_examples = []
        
        for chunk in chunks:
            if "summary" in chunk.metadata.get("tags", []):
                summaries.append(chunk)
            elif "```" in chunk.content:
                code_examples.append(chunk)
            else:
                details.append(chunk)
        
        # Assemble in order
        context_parts = []
        
        if summaries:
            context_parts.append("## Overview")
            for chunk in summaries:
                context_parts.append(chunk.content)
        
        if details:
            context_parts.append("\n## Detailed Information")
            for chunk in details:
                context_parts.append(chunk.content)
        
        if code_examples:
            context_parts.append("\n## Code Examples")
            for chunk in code_examples:
                context_parts.append(chunk.content)
        
        return "\n\n".join(context_parts)
```

### 2. Progressive Context Loading

```python
class ProgressiveContextLoader:
    """
    Load context progressively for long-running queries.
    
    Start with high-confidence chunks, add more if needed.
    """
    
    async def load_progressive(
        self,
        chunks: List[ContextChunk],
        llm_callback,
        max_iterations: int = 3
    ) -> str:
        """
        Progressively load context until LLM has enough info.
        """
        # Start with top 3 chunks
        current_context = chunks[:3]
        
        for iteration in range(max_iterations):
            # Format current context
            formatted = self.format_context(current_context)
            
            # Ask LLM if it has enough information
            response = await llm_callback(formatted)
            
            if self._has_sufficient_info(response):
                return formatted
            
            # Add more context
            if len(current_context) < len(chunks):
                current_context.append(chunks[len(current_context)])
            else:
                break
        
        # Return best effort
        return self.format_context(current_context)
    
    def _has_sufficient_info(self, response: str) -> bool:
        """Check if LLM indicates sufficient information"""
        insufficient_indicators = [
            "I don't have enough information",
            "Based on the limited context",
            "I would need more details"
        ]
        
        return not any(ind in response for ind in insufficient_indicators)
```

---

## Performance & Optimization

### Caching Formatted Context

```python
from functools import lru_cache
import hashlib

class ContextCache:
    """Cache formatted context to avoid redundant work"""
    
    @lru_cache(maxsize=500)
    def get_formatted_context(
        self,
        chunk_ids_hash: str,
        format_style: str
    ) -> Optional[str]:
        """Get cached formatted context"""
        # Implementation
        pass
    
    def cache_key(self, chunks: List[ContextChunk]) -> str:
        """Generate cache key from chunks"""
        chunk_ids = sorted([c.chunk_id for c in chunks])
        key_str = "|".join(chunk_ids)
        return hashlib.sha256(key_str.encode()).hexdigest()
```

### Parallel Processing

```python
import asyncio

async def assemble_context_parallel(chunks: List[ContextChunk]) -> str:
    """Process chunks in parallel"""
    tasks = [
        optimize_chunk(chunk),
        enrich_metadata(chunk),
        format_chunk(chunk)
    for chunk in chunks]
    
    processed_chunks = await asyncio.gather(*tasks)
    
    return combine_chunks(processed_chunks)
```

---

## Observability

### Metrics

```python
from prometheus_client import Histogram, Gauge

context_assembly_duration = Histogram(
    'context_assembly_duration_seconds',
    'Time spent assembling context'
)

context_token_count = Histogram(
    'context_token_count',
    'Number of tokens in assembled context'
)

context_chunk_count = Histogram(
    'context_chunk_count',
    'Number of chunks in assembled context'
)
```

### Logging

```python
import structlog

logger = structlog.get_logger()

def assemble_context(chunks):
    logger.info(
        "context_assembly_started",
        chunk_count=len(chunks),
        total_tokens=sum(c.token_count for c in chunks)
    )
    
    # Assembly logic...
    
    logger.info(
        "context_assembly_completed",
        final_chunk_count=len(selected_chunks),
        final_tokens=total_tokens,
        format_style=format_style
    )
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
- ✅ **AI ML Engineer (COMPLETE)** - Added Pydantic validation for context chunks
- [Bruno (Pending)]

**Review Date**: October 22, 2025  
**Document Status**: Under Review  
**Next Review**: TBD

---

