# Response Post-processing - Agent Bruno

**[← Back to Architecture](ARCHITECTURE.md)** | **[Main README](../README.md)**

---

## Table of Contents
1. [Overview](#overview)
2. [Source Citation](#source-citation)
3. [Hallucination Detection](#hallucination-detection)
4. [Response Formatting](#response-formatting)
5. [Quality Validation](#quality-validation)
6. [Memory Storage](#memory-storage)
7. [Feedback Collection](#feedback-collection)
8. [Performance Optimization](#performance-optimization)
9. [Observability](#observability)

---

## Overview

Response Post-processing is the final stage that transforms the raw LLM output into a polished, validated, and properly cited response ready for user consumption.

### Goals
- 📚 **Add source citations** - Provide traceability to source documents
- 🔍 **Detect hallucinations** - Validate response against retrieved context
- 📝 **Format response** - Convert to user-friendly Markdown
- ✅ **Validate quality** - Ensure response meets quality standards
- 💾 **Store in memory** - Save conversation for future reference
- 📊 **Collect feedback** - Enable continuous improvement

### Architecture Position

```
Query Processing
    ↓
Retrieval
    ↓
Fusion & Re-ranking
    ↓
Context Assembly
    ↓
LLM Generation (Ollama)
    ↓
┌─────────────────────────────────────────┐
│    Response Post-processing             │  ← YOU ARE HERE
│  • Source Citations                     │
│  • Hallucination Detection              │
│  • Response Formatting                  │
│  • Quality Validation                   │
│  • Memory Storage                       │
│  • Feedback Collection                  │
└─────────────────────────────────────────┘
    ↓
Return to User
```

---

## Source Citation

### Why Citations?

Citations provide:
- **Traceability** - Users can verify information
- **Trust** - Increases confidence in responses
- **Accountability** - Clear attribution of sources
- **Learning** - Users can explore source material

### Citation Extraction

```python
from typing import List, Dict, Optional
from dataclasses import dataclass
import re

@dataclass
class Citation:
    """Single citation reference"""
    number: int
    source_type: str
    source_name: str
    source_path: str
    excerpt: str
    last_updated: str
    confidence: str

class CitationExtractor:
    """
    Extract and validate citations from LLM response.
    """
    
    def __init__(self):
        # Pattern to match citation markers: [1], [2], etc.
        self.citation_pattern = r'\[(\d+)\]'
    
    def extract_citations(
        self,
        response: str,
        context_chunks: List[ContextChunk]
    ) -> List[Citation]:
        """
        Extract citation references from response.
        
        Args:
            response: LLM-generated response
            context_chunks: Original context chunks provided to LLM
        
        Returns:
            List of citations with metadata
        """
        # Find all citation markers in response
        citation_numbers = set()
        for match in re.finditer(self.citation_pattern, response):
            citation_numbers.add(int(match.group(1)))
        
        # Build citation list
        citations = []
        
        for num in sorted(citation_numbers):
            # Find corresponding context chunk
            if num <= len(context_chunks):
                chunk = context_chunks[num - 1]  # 0-indexed
                
                citation = Citation(
                    number=num,
                    source_type=chunk.metadata.get("source_type", "document"),
                    source_name=chunk.metadata.get("source_name", "Unknown"),
                    source_path=chunk.metadata.get("source_path", ""),
                    excerpt=self._extract_excerpt(chunk.content),
                    last_updated=chunk.metadata.get("last_updated", ""),
                    confidence=chunk.metadata.get("confidence", "medium")
                )
                
                citations.append(citation)
        
        return citations
    
    def _extract_excerpt(self, content: str, max_length: int = 200) -> str:
        """Extract a brief excerpt from content"""
        if len(content) <= max_length:
            return content
        
        # Find first complete sentence
        excerpt = content[:max_length]
        last_period = excerpt.rfind('.')
        
        if last_period > max_length * 0.5:
            return excerpt[:last_period + 1]
        else:
            return excerpt + "..."
    
    def validate_citations(
        self,
        response: str,
        citations: List[Citation]
    ) -> Dict:
        """
        Validate that all citations are used and valid.
        
        Returns:
            {
                "valid": True/False,
                "unused_citations": [],
                "invalid_references": [],
                "coverage": 0.85
            }
        """
        # Find citation references in response
        referenced = set()
        for match in re.finditer(self.citation_pattern, response):
            referenced.add(int(match.group(1)))
        
        # Check for unused citations
        available = {c.number for c in citations}
        unused = available - referenced
        
        # Check for invalid references (cited but not available)
        invalid = referenced - available
        
        # Calculate coverage (% of response backed by citations)
        coverage = self._calculate_citation_coverage(response, citations)
        
        return {
            "valid": len(invalid) == 0,
            "unused_citations": list(unused),
            "invalid_references": list(invalid),
            "coverage": coverage
        }
    
    def _calculate_citation_coverage(
        self,
        response: str,
        citations: List[Citation]
    ) -> float:
        """
        Estimate what percentage of response is backed by citations.
        
        Simple heuristic: count sentences with citations.
        """
        sentences = response.split('. ')
        cited_sentences = 0
        
        for sentence in sentences:
            if re.search(self.citation_pattern, sentence):
                cited_sentences += 1
        
        if not sentences:
            return 0.0
        
        return cited_sentences / len(sentences)
```

### Citation Formatting

```python
class CitationFormatter:
    """
    Format citations for display.
    """
    
    def format_inline(self, citations: List[Citation]) -> str:
        """
        Format citations as inline footnotes.
        
        Example:
        According to the Loki runbook [1], the most common cause...
        """
        # Citations are already inline in the response
        return ""
    
    def format_references(self, citations: List[Citation]) -> str:
        """
        Format citations as reference list.
        
        Example:
        ## References
        
        [1] **Runbook**: Loki Crashes
            Source: /runbooks/loki/crash-loop.md
            Updated: 2025-10-15
            Confidence: High
        
        [2] **Documentation**: Loki Configuration
            ...
        """
        if not citations:
            return ""
        
        lines = ["\n## References\n"]
        
        for citation in citations:
            lines.append(f"[{citation.number}] **{citation.source_type.title()}**: {citation.source_name}")
            lines.append(f"    Source: {citation.source_path}")
            
            if citation.last_updated:
                lines.append(f"    Updated: {citation.last_updated[:10]}")
            
            lines.append(f"    Confidence: {citation.confidence.title()}")
            
            if citation.excerpt:
                lines.append(f"    > {citation.excerpt}")
            
            lines.append("")
        
        return "\n".join(lines)
    
    def format_interactive(self, citations: List[Citation]) -> List[Dict]:
        """
        Format citations for interactive UI (with clickable links).
        
        Returns:
            List of citation objects for frontend rendering
        """
        return [
            {
                "id": f"citation-{c.number}",
                "number": c.number,
                "type": c.source_type,
                "title": c.source_name,
                "url": self._generate_url(c),
                "excerpt": c.excerpt,
                "metadata": {
                    "last_updated": c.last_updated,
                    "confidence": c.confidence
                }
            }
            for c in citations
        ]
    
    def _generate_url(self, citation: Citation) -> str:
        """Generate URL to source document"""
        # For runbooks/docs in Git
        if citation.source_path.startswith("/runbooks/"):
            return f"https://github.com/brunolucena/homelab/tree/main{citation.source_path}"
        
        # For Grafana dashboards
        elif "grafana" in citation.source_path:
            return f"https://grafana.bruno.dev{citation.source_path}"
        
        return ""
```

---

## Hallucination Detection

### What are Hallucinations?

LLM hallucinations occur when the model generates:
- **Fabricated facts** - Information not in the source material
- **Incorrect details** - Misrepresenting source information
- **Unsupported claims** - Assertions without backing evidence
- **Contradictions** - Conflicting with source documents

### Detection Strategies

```python
from typing import List, Tuple
import numpy as np
from sentence_transformers import util

class HallucinationDetector:
    """
    Detect potential hallucinations in LLM responses.
    """
    
    def __init__(self, embedding_model):
        self.model = embedding_model
        self.hallucination_indicators = [
            "I think",
            "probably",
            "maybe",
            "might be",
            "I'm not sure",
            "Based on my knowledge",  # Not from context
        ]
    
    async def detect(
        self,
        response: str,
        context_chunks: List[ContextChunk],
        threshold: float = 0.7
    ) -> Dict:
        """
        Detect potential hallucinations.
        
        Strategies:
        1. Semantic similarity check (response vs context)
        2. Fact extraction and verification
        3. Confidence language detection
        4. Citation coverage analysis
        
        Returns:
            {
                "hallucination_risk": "low" | "medium" | "high",
                "confidence_score": 0.85,
                "issues": [],
                "suggestions": []
            }
        """
        issues = []
        
        # 1. Check for uncertainty language
        uncertainty_count = sum(
            1 for indicator in self.hallucination_indicators
            if indicator.lower() in response.lower()
        )
        
        if uncertainty_count > 2:
            issues.append({
                "type": "uncertainty_language",
                "severity": "medium",
                "message": f"Response contains {uncertainty_count} uncertainty markers"
            })
        
        # 2. Semantic similarity check
        similarity = await self._check_semantic_similarity(response, context_chunks)
        
        if similarity < threshold:
            issues.append({
                "type": "low_context_similarity",
                "severity": "high",
                "message": f"Response similarity to context: {similarity:.2f} (threshold: {threshold})",
                "similarity": similarity
            })
        
        # 3. Fact verification
        facts_verified = await self._verify_facts(response, context_chunks)
        
        if facts_verified < 0.6:
            issues.append({
                "type": "unverified_facts",
                "severity": "high",
                "message": f"Only {facts_verified*100:.0f}% of facts verified in context"
            })
        
        # 4. Citation coverage
        citation_coverage = self._get_citation_coverage(response)
        
        if citation_coverage < 0.5:
            issues.append({
                "type": "low_citation_coverage",
                "severity": "medium",
                "message": f"Only {citation_coverage*100:.0f}% of response is cited"
            })
        
        # Calculate overall risk
        risk_score = self._calculate_risk_score(issues)
        
        if risk_score < 0.3:
            risk_level = "low"
        elif risk_score < 0.7:
            risk_level = "medium"
        else:
            risk_level = "high"
        
        return {
            "hallucination_risk": risk_level,
            "confidence_score": 1.0 - risk_score,
            "issues": issues,
            "suggestions": self._generate_suggestions(issues)
        }
    
    async def _check_semantic_similarity(
        self,
        response: str,
        context_chunks: List[ContextChunk]
    ) -> float:
        """
        Check semantic similarity between response and context.
        
        High similarity = response is grounded in context
        Low similarity = potential hallucination
        """
        # Get embeddings
        response_embedding = await self.model.encode(response)
        context_embeddings = await self.model.encode_batch(
            [chunk.content for chunk in context_chunks]
        )
        
        # Calculate max similarity to any context chunk
        similarities = util.cos_sim(response_embedding, context_embeddings)[0]
        max_similarity = max(similarities).item()
        
        return max_similarity
    
    async def _verify_facts(
        self,
        response: str,
        context_chunks: List[ContextChunk]
    ) -> float:
        """
        Verify factual claims in response against context.
        
        Extract claims from response and check if they appear in context.
        """
        # Extract claims (simplified)
        claims = self._extract_claims(response)
        
        if not claims:
            return 1.0  # No claims to verify
        
        # Check each claim against context
        verified_count = 0
        context_text = " ".join([c.content for c in context_chunks])
        
        for claim in claims:
            # Simple substring check (in production, use semantic matching)
            if claim.lower() in context_text.lower():
                verified_count += 1
            else:
                # Try semantic matching
                claim_embedding = await self.model.encode(claim)
                context_embeddings = await self.model.encode_batch(
                    [c.content for c in context_chunks]
                )
                
                similarities = util.cos_sim(claim_embedding, context_embeddings)[0]
                if max(similarities) > 0.8:  # High similarity threshold
                    verified_count += 1
        
        return verified_count / len(claims)
    
    def _extract_claims(self, text: str) -> List[str]:
        """
        Extract factual claims from text.
        
        Simplified: split by sentences and filter.
        In production, use NLP to extract claims.
        """
        sentences = text.split('. ')
        
        # Filter out questions, commands, and subjective statements
        claims = []
        for sentence in sentences:
            sentence = sentence.strip()
            if not sentence:
                continue
            
            # Skip questions
            if sentence.endswith('?'):
                continue
            
            # Skip commands (imperative)
            if sentence.split()[0].lower() in ['run', 'check', 'ensure', 'verify']:
                continue
            
            claims.append(sentence)
        
        return claims
    
    def _get_citation_coverage(self, response: str) -> float:
        """Get percentage of response with citations"""
        sentences = response.split('. ')
        if not sentences:
            return 0.0
        
        cited = sum(1 for s in sentences if re.search(r'\[\d+\]', s))
        return cited / len(sentences)
    
    def _calculate_risk_score(self, issues: List[Dict]) -> float:
        """Calculate overall hallucination risk score (0-1)"""
        if not issues:
            return 0.0
        
        severity_weights = {
            "low": 0.2,
            "medium": 0.5,
            "high": 0.9
        }
        
        total_risk = sum(severity_weights[issue["severity"]] for issue in issues)
        return min(1.0, total_risk / len(issues))
    
    def _generate_suggestions(self, issues: List[Dict]) -> List[str]:
        """Generate suggestions to reduce hallucination risk"""
        suggestions = []
        
        for issue in issues:
            if issue["type"] == "low_context_similarity":
                suggestions.append("Consider regenerating with more relevant context")
            elif issue["type"] == "unverified_facts":
                suggestions.append("Review factual claims against source material")
            elif issue["type"] == "low_citation_coverage":
                suggestions.append("Add more citations to support claims")
            elif issue["type"] == "uncertainty_language":
                suggestions.append("Reduce uncertainty markers or add supporting evidence")
        
        return suggestions
```

### Hallucination Mitigation

```python
class HallucinationMitigator:
    """
    Mitigate detected hallucinations.
    """
    
    async def mitigate(
        self,
        response: str,
        detection_result: Dict,
        context_chunks: List[ContextChunk]
    ) -> str:
        """
        Attempt to mitigate hallucinations.
        
        Strategies:
        1. Add disclaimer for high-risk responses
        2. Regenerate with different parameters
        3. Filter out unsupported claims
        """
        risk = detection_result["hallucination_risk"]
        
        if risk == "low":
            return response  # No mitigation needed
        
        if risk == "medium":
            # Add disclaimer
            disclaimer = "\n\n> **Note**: This response may contain information not fully verified against the source material. Please verify critical details.\n"
            return response + disclaimer
        
        else:  # high risk
            # Suggest regeneration
            suggestion = "\n\n> **Warning**: This response has high hallucination risk. Consider rephrasing your query or requesting more specific information.\n"
            return response + suggestion
```

---

## Response Formatting

### Markdown Formatting

```python
import re
from typing import List

class ResponseFormatter:
    """
    Format LLM response for user consumption.
    """
    
    def format(
        self,
        response: str,
        citations: List[Citation],
        include_references: bool = True,
        style: str = "standard"
    ) -> str:
        """
        Format response with proper Markdown.
        
        Args:
            response: Raw LLM response
            citations: List of citations
            include_references: Add reference section
            style: Formatting style (standard, compact, verbose)
        
        Returns:
            Formatted Markdown response
        """
        # 1. Clean up response
        formatted = self._clean_response(response)
        
        # 2. Format code blocks
        formatted = self._format_code_blocks(formatted)
        
        # 3. Format lists
        formatted = self._format_lists(formatted)
        
        # 4. Format emphasis
        formatted = self._format_emphasis(formatted)
        
        # 5. Add citations
        if include_references and citations:
            citation_formatter = CitationFormatter()
            references = citation_formatter.format_references(citations)
            formatted += references
        
        # 6. Apply style
        if style == "compact":
            formatted = self._apply_compact_style(formatted)
        elif style == "verbose":
            formatted = self._apply_verbose_style(formatted)
        
        return formatted
    
    def _clean_response(self, text: str) -> str:
        """Clean up common issues in LLM output"""
        # Remove excessive whitespace
        text = re.sub(r'\n{3,}', '\n\n', text)
        
        # Remove leading/trailing whitespace
        text = text.strip()
        
        # Ensure proper spacing after periods
        text = re.sub(r'\.([A-Z])', r'. \1', text)
        
        return text
    
    def _format_code_blocks(self, text: str) -> str:
        """
        Ensure code blocks are properly formatted.
        
        - Add language hints if missing
        - Ensure proper fence markers
        """
        # Find code blocks
        code_pattern = r'```(\w*)\n(.*?)```'
        
        def fix_code_block(match):
            lang = match.group(1) or 'bash'  # Default to bash
            code = match.group(2)
            return f'```{lang}\n{code}```'
        
        text = re.sub(code_pattern, fix_code_block, text, flags=re.DOTALL)
        
        return text
    
    def _format_lists(self, text: str) -> str:
        """
        Ensure lists are properly formatted.
        
        - Consistent bullet points
        - Proper indentation
        """
        # This is complex; simplified version
        return text
    
    def _format_emphasis(self, text: str) -> str:
        """Add emphasis to important terms"""
        # Bold important keywords
        keywords = ['Error', 'Warning', 'Critical', 'Important', 'Note']
        
        for keyword in keywords:
            # Only bold if not already formatted
            pattern = f'(?<!\\*\\*){keyword}(?!\\*\\*)'
            text = re.sub(pattern, f'**{keyword}**', text)
        
        return text
    
    def _apply_compact_style(self, text: str) -> str:
        """Apply compact formatting (minimal whitespace)"""
        # Remove extra newlines
        text = re.sub(r'\n\n+', '\n\n', text)
        return text
    
    def _apply_verbose_style(self, text: str) -> str:
        """Apply verbose formatting (more explanatory)"""
        # Add section headers if missing
        if not text.startswith('#'):
            text = "## Response\n\n" + text
        
        return text
```

### Structured Output

```python
from pydantic import BaseModel
from typing import Optional

class FormattedResponse(BaseModel):
    """Structured response format"""
    
    content: str
    """Main response content (Markdown)"""
    
    citations: List[Citation]
    """List of source citations"""
    
    metadata: Dict
    """Response metadata"""
    
    quality_score: float
    """Quality validation score (0-1)"""
    
    hallucination_risk: str
    """Hallucination risk level: low, medium, high"""
    
    suggestions: List[str] = []
    """Suggestions for improving response"""
    
    interactive_elements: Optional[Dict] = None
    """Interactive UI elements (buttons, links, etc.)"""

def create_formatted_response(
    raw_response: str,
    context_chunks: List[ContextChunk],
    detection_result: Dict,
    quality_result: Dict
) -> FormattedResponse:
    """
    Create structured formatted response.
    """
    # Extract citations
    citation_extractor = CitationExtractor()
    citations = citation_extractor.extract_citations(raw_response, context_chunks)
    
    # Format response
    formatter = ResponseFormatter()
    formatted_content = formatter.format(raw_response, citations)
    
    return FormattedResponse(
        content=formatted_content,
        citations=citations,
        metadata={
            "timestamp": datetime.utcnow().isoformat(),
            "model": "llama3.2:8b",
            "context_chunks_used": len(context_chunks)
        },
        quality_score=quality_result["overall_score"],
        hallucination_risk=detection_result["hallucination_risk"],
        suggestions=detection_result.get("suggestions", [])
    )
```

---

## Quality Validation

### Response Quality Metrics

```python
class QualityValidator:
    """
    Validate response quality.
    """
    
    def validate(self, response: str, query: str, context: List[ContextChunk]) -> Dict:
        """
        Validate response quality across multiple dimensions.
        
        Returns:
            {
                "overall_score": 0.85,
                "metrics": {
                    "relevance": 0.9,
                    "completeness": 0.8,
                    "clarity": 0.85,
                    "actionability": 0.9
                },
                "issues": [],
                "passes_validation": True
            }
        """
        metrics = {}
        issues = []
        
        # 1. Relevance - Does response answer the query?
        metrics["relevance"] = self._check_relevance(response, query)
        if metrics["relevance"] < 0.7:
            issues.append("Response may not fully address the query")
        
        # 2. Completeness - Is the answer complete?
        metrics["completeness"] = self._check_completeness(response, query)
        if metrics["completeness"] < 0.6:
            issues.append("Response appears incomplete")
        
        # 3. Clarity - Is the response clear and well-structured?
        metrics["clarity"] = self._check_clarity(response)
        if metrics["clarity"] < 0.7:
            issues.append("Response could be clearer")
        
        # 4. Actionability - Can the user act on this information?
        metrics["actionability"] = self._check_actionability(response)
        if metrics["actionability"] < 0.5:
            issues.append("Response lacks actionable steps")
        
        # Calculate overall score
        overall_score = sum(metrics.values()) / len(metrics)
        
        # Check minimum threshold
        passes_validation = overall_score >= 0.7 and len(issues) < 3
        
        return {
            "overall_score": overall_score,
            "metrics": metrics,
            "issues": issues,
            "passes_validation": passes_validation
        }
    
    def _check_relevance(self, response: str, query: str) -> float:
        """Check if response is relevant to query"""
        # Simple keyword overlap (in production, use semantic similarity)
        query_words = set(query.lower().split())
        response_words = set(response.lower().split())
        
        overlap = len(query_words & response_words)
        return min(1.0, overlap / len(query_words))
    
    def _check_completeness(self, response: str, query: str) -> float:
        """Check if response is complete"""
        # Heuristics:
        # - Minimum length
        # - Has conclusion
        # - Answers main question
        
        score = 0.0
        
        # Length check
        if len(response) > 100:
            score += 0.3
        elif len(response) > 300:
            score += 0.5
        
        # Has multiple paragraphs
        paragraphs = response.split('\n\n')
        if len(paragraphs) >= 2:
            score += 0.3
        
        # Doesn't end abruptly
        if response[-1] in '.!':
            score += 0.2
        
        return min(1.0, score)
    
    def _check_clarity(self, response: str) -> float:
        """Check response clarity"""
        score = 1.0
        
        # Check for overly long sentences
        sentences = response.split('. ')
        avg_sentence_length = sum(len(s.split()) for s in sentences) / len(sentences)
        
        if avg_sentence_length > 30:
            score -= 0.2  # Too complex
        
        # Check for structure (headers, lists)
        if '##' in response or '\n-' in response:
            score += 0.1  # Well-structured
        
        # Check for code examples if technical
        if '```' in response:
            score += 0.1  # Has examples
        
        return min(1.0, max(0.0, score))
    
    def _check_actionability(self, response: str) -> float:
        """Check if response provides actionable information"""
        score = 0.0
        
        # Has step-by-step instructions
        if re.search(r'\b\d+\.\s+', response):
            score += 0.4
        
        # Has commands/code
        if '```' in response:
            score += 0.3
        
        # Has action verbs
        action_verbs = ['run', 'execute', 'check', 'verify', 'configure', 'set', 'create']
        if any(verb in response.lower() for verb in action_verbs):
            score += 0.3
        
        return min(1.0, score)
```

---

## Memory Storage

### Store Conversation

```python
from datetime import datetime

class ConversationMemory:
    """
    Store conversation turns in long-term memory.
    """
    
    def __init__(self, lancedb_client):
        self.db = lancedb_client
    
    async def store_conversation_turn(
        self,
        user_query: str,
        agent_response: str,
        context_used: List[ContextChunk],
        metadata: Dict
    ) -> str:
        """
        Store conversation turn in LanceDB.
        
        Returns:
            Conversation turn ID
        """
        turn_data = {
            "turn_id": self._generate_turn_id(),
            "timestamp": datetime.utcnow().isoformat(),
            "user_id": metadata.get("user_id"),
            "session_id": metadata.get("session_id"),
            "query": user_query,
            "response": agent_response,
            "context_chunks": [c.chunk_id for c in context_used],
            "query_type": metadata.get("query_type"),
            "quality_score": metadata.get("quality_score"),
            "hallucination_risk": metadata.get("hallucination_risk"),
            "user_feedback": None,  # To be updated later
        }
        
        # Generate embedding for retrieval
        turn_text = f"{user_query} {agent_response}"
        embedding = await self._generate_embedding(turn_text)
        turn_data["vector"] = embedding
        
        # Store in LanceDB
        await self.db.table("episodic_memory").add([turn_data])
        
        return turn_data["turn_id"]
    
    def _generate_turn_id(self) -> str:
        """Generate unique turn ID"""
        import uuid
        return f"turn_{uuid.uuid4().hex[:12]}"
    
    async def _generate_embedding(self, text: str):
        """Generate embedding for conversation turn"""
        # Use embedding model
        pass
```

---

## Feedback Collection

### Implicit Feedback

```python
class FeedbackCollector:
    """
    Collect user feedback on responses.
    """
    
    async def collect_implicit_feedback(
        self,
        turn_id: str,
        user_actions: Dict
    ):
        """
        Collect implicit feedback from user actions.
        
        Implicit signals:
        - Did user ask follow-up question?
        - Did user rephrase query?
        - Time spent reading response
        - Click-through on citations
        """
        feedback = {
            "turn_id": turn_id,
            "timestamp": datetime.utcnow().isoformat(),
            "feedback_type": "implicit",
            "signals": {
                "followup_query": user_actions.get("followup_query"),
                "query_rephrased": user_actions.get("query_rephrased", False),
                "time_spent_seconds": user_actions.get("time_spent_seconds", 0),
                "citations_clicked": user_actions.get("citations_clicked", []),
                "response_copied": user_actions.get("response_copied", False)
            }
        }
        
        # Calculate implicit satisfaction score
        feedback["satisfaction_score"] = self._calculate_implicit_satisfaction(
            feedback["signals"]
        )
        
        # Store feedback
        await self._store_feedback(feedback)
    
    def _calculate_implicit_satisfaction(self, signals: Dict) -> float:
        """
        Calculate satisfaction score from implicit signals.
        
        Positive signals:
        - Long time spent reading (engaged)
        - Citations clicked (validating)
        - Response copied (useful)
        
        Negative signals:
        - Immediate follow-up (didn't answer)
        - Query rephrased (unclear response)
        """
        score = 0.5  # Neutral baseline
        
        # Time spent
        time_spent = signals.get("time_spent_seconds", 0)
        if time_spent > 30:
            score += 0.2
        elif time_spent < 5:
            score -= 0.2
        
        # Citations clicked
        if len(signals.get("citations_clicked", [])) > 0:
            score += 0.1
        
        # Response copied
        if signals.get("response_copied"):
            score += 0.2
        
        # Negative signals
        if signals.get("query_rephrased"):
            score -= 0.3
        
        return max(0.0, min(1.0, score))
```

### Explicit Feedback

```python
async def collect_explicit_feedback(
    self,
    turn_id: str,
    rating: int,  # 1-5 stars
    comment: str = None
):
    """
    Collect explicit user feedback.
    
    Args:
        turn_id: Conversation turn ID
        rating: 1-5 stars
        comment: Optional text feedback
    """
    feedback = {
        "turn_id": turn_id,
        "timestamp": datetime.utcnow().isoformat(),
        "feedback_type": "explicit",
        "rating": rating,
        "comment": comment,
        "satisfaction_score": rating / 5.0
    }
    
    await self._store_feedback(feedback)
```

---

## Performance Optimization

### Caching

```python
from functools import lru_cache

class ResponseCache:
    """Cache formatted responses"""
    
    @lru_cache(maxsize=100)
    def get_formatted_response(self, response_hash: str) -> Optional[str]:
        """Get cached formatted response"""
        pass
```

---

## Observability

### Metrics

```python
from prometheus_client import Histogram, Counter

response_processing_duration = Histogram(
    'response_processing_duration_seconds',
    'Time spent post-processing responses',
    ['stage']
)

response_quality_score = Histogram(
    'response_quality_score',
    'Response quality validation scores'
)

hallucination_risk_total = Counter(
    'hallucination_risk_total',
    'Hallucination risk detections',
    ['risk_level']
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
- [AI ML Engineer (Pending)]
- [Bruno (Pending)]

**Review Date**: October 22, 2025  
**Document Status**: Under Review  
**Next Review**: TBD

---

