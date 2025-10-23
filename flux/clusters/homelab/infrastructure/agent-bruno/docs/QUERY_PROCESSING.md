# Query Processing - Agent Bruno

**[← Back to Architecture](ARCHITECTURE.md)** | **[Main README](../README.md)**

---

## Table of Contents
1. [Overview](#overview)
2. [Query Analysis Pipeline](#query-analysis-pipeline)
3. [Intent Classification](#intent-classification)
4. [Entity Extraction](#entity-extraction)
5. [Query Expansion](#query-expansion)
6. [Query Transformation](#query-transformation)
7. [Performance & Optimization](#performance--optimization)
8. [Observability](#observability)

---

## Overview

Query Processing is the first critical stage in Agent Bruno's RAG pipeline. It transforms raw user input into optimized queries that can be effectively used for both semantic and keyword-based retrieval.

### Goals
- 🎯 **Understand user intent** - Classify what the user is trying to accomplish
- 🔍 **Extract entities** - Identify key components, services, and concepts
- 📈 **Expand queries** - Add synonyms and related terms to improve recall
- ⚡ **Optimize for retrieval** - Transform queries for maximum retrieval effectiveness

### Architecture Position

```
User Query
    ↓
┌─────────────────────────────────────────┐
│      Query Analysis & Processing         │  ← YOU ARE HERE
│  • Intent Classification                │
│  • Entity Extraction                    │
│  • Query Expansion                      │
│  • Query Transformation                 │
└─────────────────────────────────────────┘
    ↓
Semantic Search Path  +  Keyword Search Path
    ↓
Fusion & Re-ranking
    ↓
Context Assembly
```

---

## Query Analysis Pipeline

### 1. Input Normalization

**Purpose**: Clean and standardize the user query before processing.

```python
from typing import Dict, List
import re
from pydantic import BaseModel

class QueryInput(BaseModel):
    text: str
    user_id: str
    session_id: str
    context: Dict = {}

class NormalizedQuery(BaseModel):
    original_text: str
    normalized_text: str
    language: str = "en"
    has_code: bool = False
    code_blocks: List[str] = []
    
def normalize_query(query_input: QueryInput) -> NormalizedQuery:
    """
    Normalize user query for processing.
    
    Steps:
    1. Trim whitespace
    2. Lowercase (preserve for NER)
    3. Extract code blocks
    4. Remove excessive punctuation
    5. Detect language
    """
    text = query_input.text.strip()
    
    # Extract code blocks (preserve original case)
    code_pattern = r'```[\w]*\n(.*?)```'
    code_blocks = re.findall(code_pattern, text, re.DOTALL)
    
    # Remove code blocks for text processing
    text_without_code = re.sub(code_pattern, ' [CODE_BLOCK] ', text, flags=re.DOTALL)
    
    # Normalize whitespace
    normalized = ' '.join(text_without_code.split())
    
    # Detect language (simple heuristic)
    # In production, use langdetect or similar
    language = "en"  # Default
    
    return NormalizedQuery(
        original_text=query_input.text,
        normalized_text=normalized,
        language=language,
        has_code=len(code_blocks) > 0,
        code_blocks=code_blocks
    )
```

**Example**:
```python
Input:  "   How do I fix   Loki crashes?  \n\n  "
Output: "How do I fix Loki crashes?"
```

### 2. Query Understanding

**Purpose**: Analyze the semantic structure of the query.

```python
from enum import Enum

class QueryType(str, Enum):
    QUESTION = "question"           # "How do I...?", "What is...?"
    COMMAND = "command"             # "Show me...", "List all..."
    TROUBLESHOOTING = "troubleshooting"  # "Fix...", "Debug..."
    EXPLANATION = "explanation"     # "Explain...", "Why does..."
    COMPARISON = "comparison"       # "Difference between..."
    PROCEDURAL = "procedural"       # "Steps to...", "How to..."

class QueryComplexity(str, Enum):
    SIMPLE = "simple"      # Single concept, direct lookup
    MEDIUM = "medium"      # Multiple concepts, requires reasoning
    COMPLEX = "complex"    # Multi-step, requires deep analysis

class QueryAnalysis(BaseModel):
    query_type: QueryType
    complexity: QueryComplexity
    requires_context: bool
    requires_code: bool
    requires_realtime_data: bool
    confidence: float

def analyze_query_structure(normalized_query: NormalizedQuery) -> QueryAnalysis:
    """
    Analyze query structure to determine type and complexity.
    
    This helps:
    - Route to appropriate retrieval strategy
    - Determine context window requirements
    - Adjust re-ranking weights
    """
    text = normalized_query.normalized_text.lower()
    
    # Detect query type
    query_type = QueryType.QUESTION
    if text.startswith(("how ", "how do ", "how to ")):
        query_type = QueryType.PROCEDURAL
    elif text.startswith(("what ", "what is ", "what are ")):
        query_type = QueryType.EXPLANATION
    elif text.startswith(("fix ", "debug ", "troubleshoot ")):
        query_type = QueryType.TROUBLESHOOTING
    elif " vs " in text or "difference between" in text:
        query_type = QueryType.COMPARISON
    elif text.startswith(("show ", "list ", "get ")):
        query_type = QueryType.COMMAND
    
    # Detect complexity (simple heuristic)
    word_count = len(text.split())
    has_multiple_entities = len(re.findall(r'\b[A-Z][a-z]+\b', normalized_query.normalized_text)) > 2
    
    if word_count <= 5 and not has_multiple_entities:
        complexity = QueryComplexity.SIMPLE
    elif word_count <= 15 or has_multiple_entities:
        complexity = QueryComplexity.MEDIUM
    else:
        complexity = QueryComplexity.COMPLEX
    
    # Determine requirements
    requires_context = query_type in [QueryType.TROUBLESHOOTING, QueryType.PROCEDURAL]
    requires_code = normalized_query.has_code or "code" in text or "snippet" in text
    requires_realtime_data = any(kw in text for kw in ["current", "now", "latest", "today"])
    
    return QueryAnalysis(
        query_type=query_type,
        complexity=complexity,
        requires_context=requires_context,
        requires_code=requires_code,
        requires_realtime_data=requires_realtime_data,
        confidence=0.85  # In production, use ML model confidence
    )
```

---

## Intent Classification

### Purpose

Intent classification determines **what the user wants to accomplish**, which influences:
- Which knowledge sources to prioritize
- What type of response to generate
- Whether to invoke external tools (MCP servers)

### Classification Categories

```python
class UserIntent(str, Enum):
    # Knowledge Retrieval
    LOOKUP_RUNBOOK = "lookup_runbook"           # "How to fix X?"
    SEARCH_DOCUMENTATION = "search_documentation"  # "Find docs for X"
    EXPLAIN_CONCEPT = "explain_concept"         # "What is X?"
    
    # Operational
    CHECK_STATUS = "check_status"               # "Is X running?"
    QUERY_METRICS = "query_metrics"             # "Show metrics for X"
    ANALYZE_LOGS = "analyze_logs"               # "Check logs for errors"
    
    # Troubleshooting
    DEBUG_ISSUE = "debug_issue"                 # "X is not working"
    ROOT_CAUSE_ANALYSIS = "root_cause_analysis" # "Why did X fail?"
    
    # Code & Configuration
    GENERATE_CODE = "generate_code"             # "Create a Kubernetes manifest"
    REVIEW_CODE = "review_code"                 # "Review this YAML"
    
    # Learning
    LEARN_TOPIC = "learn_topic"                 # "Teach me about X"
    COMPARE = "compare"                         # "X vs Y"
    
    # Conversational
    FOLLOWUP = "followup"                       # Continues previous conversation
    CLARIFICATION = "clarification"             # User asking for clarification
    GENERAL_CHAT = "general_chat"               # Non-technical chat

class IntentClassifier:
    """
    Classify user intent using a combination of:
    1. Rule-based patterns (fast, interpretable)
    2. ML model (accurate, context-aware)
    3. Conversation context (for follow-ups)
    """
    
    def __init__(self, ollama_url: str):
        self.ollama_url = ollama_url
        self.intent_patterns = self._build_patterns()
    
    def _build_patterns(self) -> Dict[UserIntent, List[str]]:
        """Rule-based patterns for fast classification"""
        return {
            UserIntent.LOOKUP_RUNBOOK: [
                r"how (?:do i|to|can i) (?:fix|solve|resolve)",
                r"(?:fix|debug|troubleshoot) .+",
                r"runbook for .+"
            ],
            UserIntent.CHECK_STATUS: [
                r"(?:is|are) .+ (?:running|up|down|healthy)",
                r"status of .+",
                r"check if .+ is .+"
            ],
            UserIntent.QUERY_METRICS: [
                r"(?:show|get|display) metrics (?:for|of)",
                r"what (?:is|are) the .+ (?:metrics|stats)",
                r"cpu|memory|latency|throughput .+ usage"
            ],
            UserIntent.EXPLAIN_CONCEPT: [
                r"what is .+",
                r"explain .+",
                r"define .+"
            ],
            # ... more patterns
        }
    
    def classify_fast(self, query: str) -> Optional[UserIntent]:
        """Fast rule-based classification"""
        query_lower = query.lower()
        
        for intent, patterns in self.intent_patterns.items():
            for pattern in patterns:
                if re.search(pattern, query_lower):
                    return intent
        return None
    
    async def classify_ml(self, query: str, context: Dict) -> UserIntent:
        """
        ML-based classification using Ollama.
        Falls back to rule-based if ML fails.
        """
        # Try fast classification first
        fast_result = self.classify_fast(query)
        if fast_result and len(query.split()) < 10:
            # For simple queries, trust rule-based
            return fast_result
        
        # Use LLM for complex queries
        prompt = f"""Classify the user's intent from the following query.

Query: {query}

Context: {context.get('recent_queries', [])}

Intent Categories:
- lookup_runbook: User wants troubleshooting steps
- search_documentation: User wants to find docs
- explain_concept: User wants to understand something
- check_status: User wants to check system status
- query_metrics: User wants metrics/stats
- analyze_logs: User wants to check logs
- debug_issue: User is reporting a problem
- generate_code: User wants code/config generated
- compare: User wants to compare options
- followup: Continuing previous conversation
- general_chat: Casual conversation

Respond with ONLY the intent category, nothing else."""

        try:
            # Call Ollama for classification
            # In production, use a fine-tuned classifier
            response = await self._call_ollama(prompt)
            intent_str = response.strip().lower()
            
            # Map to enum
            try:
                return UserIntent(intent_str)
            except ValueError:
                # Fall back to rule-based
                return fast_result or UserIntent.GENERAL_CHAT
        except Exception as e:
            logger.error(f"ML classification failed: {e}")
            return fast_result or UserIntent.GENERAL_CHAT
    
    async def _call_ollama(self, prompt: str) -> str:
        """Call Ollama for classification"""
        # Implementation details
        pass
```

### Intent-Based Routing

```python
def get_retrieval_strategy(intent: UserIntent) -> Dict:
    """
    Determine retrieval strategy based on intent.
    
    Returns configuration for:
    - Which knowledge tables to query
    - Relative weights for semantic vs keyword
    - Whether to invoke real-time tools
    """
    strategies = {
        UserIntent.LOOKUP_RUNBOOK: {
            "tables": ["runbooks", "knowledge_base"],
            "semantic_weight": 0.7,
            "keyword_weight": 0.3,
            "top_k": 10,
            "invoke_tools": False
        },
        UserIntent.CHECK_STATUS: {
            "tables": [],
            "semantic_weight": 0.0,
            "keyword_weight": 0.0,
            "top_k": 0,
            "invoke_tools": True,
            "tools": ["grafana_mcp.query_metrics", "prometheus_query"]
        },
        UserIntent.QUERY_METRICS: {
            "tables": ["knowledge_base"],
            "semantic_weight": 0.3,
            "keyword_weight": 0.1,
            "top_k": 3,
            "invoke_tools": True,
            "tools": ["grafana_mcp.query_metrics"]
        },
        # ... more strategies
    }
    
    return strategies.get(intent, {
        "tables": ["knowledge_base"],
        "semantic_weight": 0.6,
        "keyword_weight": 0.4,
        "top_k": 5,
        "invoke_tools": False
    })
```

---

## Entity Extraction

### Purpose

Extract key entities from the query to:
- **Filter search results** by relevant systems/components
- **Enrich context** with entity metadata
- **Route to specialists** (e.g., Loki expert)

### Entity Types

```python
class EntityType(str, Enum):
    # Infrastructure
    SERVICE = "service"           # "loki", "prometheus", "grafana"
    POD = "pod"                   # Pod names
    NAMESPACE = "namespace"       # "monitoring", "agent-bruno"
    NODE = "node"                 # Kubernetes nodes
    
    # Observability
    METRIC = "metric"             # "cpu_usage", "memory_usage"
    LOG_SOURCE = "log_source"     # Log streams
    TRACE = "trace"               # Trace IDs
    
    # Concepts
    TECHNOLOGY = "technology"     # "kubernetes", "docker", "helm"
    PATTERN = "pattern"           # "circuit breaker", "retry"
    
    # Errors
    ERROR_TYPE = "error_type"     # "OOMKilled", "CrashLoopBackOff"
    STATUS_CODE = "status_code"   # "500", "404"

class Entity(BaseModel):
    text: str
    type: EntityType
    confidence: float
    start_idx: int
    end_idx: int
    metadata: Dict = {}

class EntityExtractor:
    """
    Multi-strategy entity extraction:
    1. Dictionary lookup (fast, high precision)
    2. Regex patterns (medium speed, good recall)
    3. NER model (slow, best accuracy)
    """
    
    def __init__(self):
        self.entity_dict = self._load_entity_dictionary()
        self.patterns = self._build_patterns()
    
    def _load_entity_dictionary(self) -> Dict[str, EntityType]:
        """
        Load known entities from knowledge base.
        Updated automatically as new services are discovered.
        """
        return {
            # Services
            "loki": EntityType.SERVICE,
            "prometheus": EntityType.SERVICE,
            "grafana": EntityType.SERVICE,
            "tempo": EntityType.SERVICE,
            "alloy": EntityType.SERVICE,
            "minio": EntityType.SERVICE,
            "mongodb": EntityType.SERVICE,
            "redis": EntityType.SERVICE,
            "postgres": EntityType.SERVICE,
            "knative": EntityType.SERVICE,
            
            # Namespaces
            "agent-bruno": EntityType.NAMESPACE,
            "monitoring": EntityType.NAMESPACE,
            "loki": EntityType.NAMESPACE,
            "tempo": EntityType.NAMESPACE,
            
            # Technologies
            "kubernetes": EntityType.TECHNOLOGY,
            "k8s": EntityType.TECHNOLOGY,
            "helm": EntityType.TECHNOLOGY,
            "docker": EntityType.TECHNOLOGY,
            "flux": EntityType.TECHNOLOGY,
            
            # Error types
            "oomkilled": EntityType.ERROR_TYPE,
            "crashloopbackoff": EntityType.ERROR_TYPE,
            "imagepullbackoff": EntityType.ERROR_TYPE,
            
            # ... more entities
        }
    
    def _build_patterns(self) -> List[tuple]:
        """Regex patterns for entity extraction"""
        return [
            # Pod names: app-name-abc123-xyz
            (r'\b[\w-]+-[a-f0-9]{8,10}-[a-z0-9]{5}\b', EntityType.POD),
            
            # Metrics: metric_name{labels}
            (r'\b[\w_]+(?:\{[^}]+\})?\b', EntityType.METRIC),
            
            # Status codes
            (r'\b[45]\d{2}\b', EntityType.STATUS_CODE),
            
            # Trace IDs (hex)
            (r'\b[a-f0-9]{32}\b', EntityType.TRACE),
        ]
    
    def extract(self, query: str) -> List[Entity]:
        """
        Extract entities using multi-strategy approach.
        """
        entities = []
        query_lower = query.lower()
        
        # Strategy 1: Dictionary lookup (O(n) with trie)
        for entity_text, entity_type in self.entity_dict.items():
            if entity_text in query_lower:
                start_idx = query_lower.index(entity_text)
                entities.append(Entity(
                    text=entity_text,
                    type=entity_type,
                    confidence=0.95,  # High confidence for known entities
                    start_idx=start_idx,
                    end_idx=start_idx + len(entity_text)
                ))
        
        # Strategy 2: Regex patterns
        for pattern, entity_type in self.patterns:
            for match in re.finditer(pattern, query, re.IGNORECASE):
                entities.append(Entity(
                    text=match.group(0),
                    type=entity_type,
                    confidence=0.75,  # Medium confidence for pattern matches
                    start_idx=match.start(),
                    end_idx=match.end()
                ))
        
        # Strategy 3: NER model (optional, for complex queries)
        # if query_complexity == QueryComplexity.COMPLEX:
        #     entities.extend(self._extract_with_ner(query))
        
        # Deduplicate and sort by confidence
        entities = self._deduplicate_entities(entities)
        entities.sort(key=lambda e: e.confidence, reverse=True)
        
        return entities
    
    def _deduplicate_entities(self, entities: List[Entity]) -> List[Entity]:
        """Remove duplicate/overlapping entities"""
        seen = set()
        unique = []
        
        for entity in sorted(entities, key=lambda e: e.confidence, reverse=True):
            # Create a key for deduplication
            key = (entity.text.lower(), entity.type)
            if key not in seen:
                seen.add(key)
                unique.append(entity)
        
        return unique
```

### Entity Enrichment

```python
async def enrich_entities(entities: List[Entity]) -> List[Entity]:
    """
    Enrich extracted entities with metadata from knowledge base.
    
    For example:
    - Service "loki" → namespace, port, recent alerts
    - Error "OOMKilled" → common causes, runbooks
    """
    for entity in entities:
        if entity.type == EntityType.SERVICE:
            # Query knowledge base for service metadata
            metadata = await get_service_metadata(entity.text)
            entity.metadata = metadata
        
        elif entity.type == EntityType.ERROR_TYPE:
            # Link to runbooks
            runbooks = await find_runbooks_for_error(entity.text)
            entity.metadata = {"runbooks": runbooks}
    
    return entities
```

---

## Query Expansion

### Purpose

Expand the original query with **synonyms**, **related terms**, and **domain-specific variations** to improve recall.

### Expansion Strategies

```python
class QueryExpansion(BaseModel):
    original_query: str
    expanded_terms: List[str]
    synonyms: Dict[str, List[str]]
    related_concepts: List[str]
    acronyms: Dict[str, str]

class QueryExpander:
    """
    Query expansion using:
    1. Synonym dictionaries (WordNet, custom)
    2. Domain-specific expansions
    3. Acronym expansion
    4. Embedding-based similarity
    """
    
    def __init__(self):
        self.synonyms = self._load_synonyms()
        self.acronyms = self._load_acronyms()
        self.domain_expansions = self._load_domain_expansions()
    
    def _load_synonyms(self) -> Dict[str, List[str]]:
        """Load synonym mappings"""
        return {
            "fix": ["repair", "resolve", "solve", "troubleshoot"],
            "crash": ["failure", "error", "down", "not working"],
            "slow": ["latency", "performance issue", "high response time"],
            "memory": ["ram", "heap", "oom"],
            # ... more synonyms
        }
    
    def _load_acronyms(self) -> Dict[str, str]:
        """Load acronym expansions"""
        return {
            "k8s": "kubernetes",
            "oom": "out of memory",
            "cpu": "central processing unit",
            "ram": "random access memory",
            "api": "application programming interface",
            "mcp": "model context protocol",
            # ... more acronyms
        }
    
    def _load_domain_expansions(self) -> Dict[str, List[str]]:
        """Domain-specific term expansions"""
        return {
            "loki": ["grafana loki", "loki logs", "log aggregation"],
            "prometheus": ["prometheus metrics", "promql", "metric collection"],
            "knative": ["knative serving", "knative eventing", "serverless"],
            # ... more expansions
        }
    
    def expand(self, query: str, entities: List[Entity]) -> QueryExpansion:
        """
        Expand query with synonyms and related terms.
        """
        words = query.lower().split()
        expanded_terms = []
        synonyms_used = {}
        related_concepts = []
        acronyms_expanded = {}
        
        # 1. Synonym expansion
        for word in words:
            if word in self.synonyms:
                synonyms_used[word] = self.synonyms[word]
                expanded_terms.extend(self.synonyms[word])
        
        # 2. Acronym expansion
        for word in words:
            if word in self.acronyms:
                acronyms_expanded[word] = self.acronyms[word]
                expanded_terms.append(self.acronyms[word])
        
        # 3. Domain-specific expansion
        for entity in entities:
            entity_text = entity.text.lower()
            if entity_text in self.domain_expansions:
                related_concepts.extend(self.domain_expansions[entity_text])
                expanded_terms.extend(self.domain_expansions[entity_text])
        
        # 4. Remove duplicates
        expanded_terms = list(set(expanded_terms))
        related_concepts = list(set(related_concepts))
        
        return QueryExpansion(
            original_query=query,
            expanded_terms=expanded_terms,
            synonyms=synonyms_used,
            related_concepts=related_concepts,
            acronyms=acronyms_expanded
        )
```

**Example**:
```python
Query: "How to fix k8s OOM crashes?"

Expansion:
{
    "original_query": "How to fix k8s OOM crashes?",
    "expanded_terms": [
        "repair", "resolve", "troubleshoot",  # fix synonyms
        "kubernetes",                         # k8s acronym
        "out of memory",                      # OOM acronym
        "failure", "error"                    # crash synonyms
    ],
    "synonyms": {
        "fix": ["repair", "resolve", "troubleshoot"],
        "crash": ["failure", "error"]
    },
    "acronyms": {
        "k8s": "kubernetes",
        "oom": "out of memory"
    },
    "related_concepts": [
        "memory limits", "resource quotas", "pod eviction"
    ]
}
```

---

## Query Transformation

### Purpose

Transform the query into optimal formats for different retrieval strategies:
- **Semantic search**: Natural language embedding
- **Keyword search**: Tokenized terms with weights
- **Metadata filters**: Structured filters for LanceDB

### Transformation Pipeline

```python
class TransformedQuery(BaseModel):
    semantic_query: str
    keyword_terms: List[str]
    keyword_weights: Dict[str, float]
    metadata_filters: Dict[str, Any]
    boosted_fields: List[str]

class QueryTransformer:
    """
    Transform processed query for retrieval.
    """
    
    def transform(
        self,
        normalized_query: NormalizedQuery,
        analysis: QueryAnalysis,
        intent: UserIntent,
        entities: List[Entity],
        expansion: QueryExpansion
    ) -> TransformedQuery:
        """
        Create optimized query representations.
        """
        
        # 1. Semantic query (for dense retrieval)
        semantic_query = self._build_semantic_query(
            normalized_query, expansion, entities
        )
        
        # 2. Keyword terms (for BM25)
        keyword_terms = self._extract_keywords(
            normalized_query, entities, expansion
        )
        
        # 3. Keyword weights (TF-IDF inspired)
        keyword_weights = self._calculate_term_weights(
            keyword_terms, entities
        )
        
        # 4. Metadata filters
        metadata_filters = self._build_metadata_filters(
            entities, intent, analysis
        )
        
        # 5. Field boosting
        boosted_fields = self._determine_boosted_fields(intent)
        
        return TransformedQuery(
            semantic_query=semantic_query,
            keyword_terms=keyword_terms,
            keyword_weights=keyword_weights,
            metadata_filters=metadata_filters,
            boosted_fields=boosted_fields
        )
    
    def _build_semantic_query(
        self,
        normalized_query: NormalizedQuery,
        expansion: QueryExpansion,
        entities: List[Entity]
    ) -> str:
        """
        Build semantic query by enriching with expansion terms.
        """
        parts = [normalized_query.normalized_text]
        
        # Add expanded terms (limited to top 5 to avoid noise)
        if expansion.expanded_terms:
            parts.append(" ".join(expansion.expanded_terms[:5]))
        
        # Add entity context
        for entity in entities[:3]:  # Top 3 entities
            if entity.type in [EntityType.SERVICE, EntityType.TECHNOLOGY]:
                parts.append(entity.text)
        
        return " ".join(parts)
    
    def _extract_keywords(
        self,
        normalized_query: NormalizedQuery,
        entities: List[Entity],
        expansion: QueryExpansion
    ) -> List[str]:
        """
        Extract keyword terms for BM25 search.
        """
        # Start with query words
        words = set(normalized_query.normalized_text.lower().split())
        
        # Add entity texts
        for entity in entities:
            words.add(entity.text.lower())
        
        # Add top expanded terms
        words.update(expansion.expanded_terms[:10])
        
        # Remove stop words
        stop_words = {"a", "an", "the", "is", "are", "to", "of", "in", "for"}
        keywords = [w for w in words if w not in stop_words and len(w) > 2]
        
        return keywords
    
    def _calculate_term_weights(
        self,
        terms: List[str],
        entities: List[Entity]
    ) -> Dict[str, float]:
        """
        Assign weights to terms based on importance.
        """
        weights = {}
        
        # Base weight for all terms
        for term in terms:
            weights[term] = 1.0
        
        # Boost entity terms
        for entity in entities:
            entity_text = entity.text.lower()
            if entity_text in weights:
                # Boost based on entity type and confidence
                boost = 1.0
                if entity.type == EntityType.SERVICE:
                    boost = 2.0
                elif entity.type == EntityType.ERROR_TYPE:
                    boost = 1.5
                
                weights[entity_text] *= boost * entity.confidence
        
        return weights
    
    def _build_metadata_filters(
        self,
        entities: List[Entity],
        intent: UserIntent,
        analysis: QueryAnalysis
    ) -> Dict[str, Any]:
        """
        Build metadata filters for LanceDB.
        """
        filters = {}
        
        # Filter by source type based on intent
        if intent == UserIntent.LOOKUP_RUNBOOK:
            filters["source_type"] = "runbook"
        elif intent == UserIntent.SEARCH_DOCUMENTATION:
            filters["source_type"] = "documentation"
        
        # Filter by entities (service, namespace, etc.)
        services = [e.text for e in entities if e.type == EntityType.SERVICE]
        if services:
            filters["tags"] = {"$in": services}
        
        # Recency filter for real-time queries
        if analysis.requires_realtime_data:
            from datetime import datetime, timedelta
            recent_date = datetime.utcnow() - timedelta(days=30)
            filters["last_updated"] = {"$gte": recent_date}
        
        return filters
    
    def _determine_boosted_fields(self, intent: UserIntent) -> List[str]:
        """
        Determine which fields to boost in search.
        """
        if intent == UserIntent.LOOKUP_RUNBOOK:
            return ["title", "symptoms", "solution"]
        elif intent == UserIntent.EXPLAIN_CONCEPT:
            return ["title", "description", "examples"]
        else:
            return ["title", "content"]
```

---

## Performance & Optimization

### Caching Strategy

```python
from functools import lru_cache
import hashlib

class QueryCache:
    """
    Cache processed queries to avoid redundant work.
    """
    
    def __init__(self, max_size: int = 1000):
        self.cache = {}
        self.max_size = max_size
    
    def get_cache_key(self, query: str, user_id: str) -> str:
        """Generate cache key"""
        key_str = f"{query}:{user_id}"
        return hashlib.sha256(key_str.encode()).hexdigest()
    
    @lru_cache(maxsize=1000)
    def get_processed_query(self, query: str, user_id: str):
        """Cache processed query results"""
        # Implementation
        pass
```

### Batch Processing

For multiple queries (e.g., query variations), process in batches:

```python
async def process_queries_batch(queries: List[str]) -> List[TransformedQuery]:
    """Process multiple queries in parallel"""
    tasks = [process_single_query(q) for q in queries]
    return await asyncio.gather(*tasks)
```

### Performance Metrics

```python
import time
from prometheus_client import Histogram

query_processing_duration = Histogram(
    'query_processing_duration_seconds',
    'Time spent processing queries',
    ['stage']
)

@query_processing_duration.labels(stage='normalization').time()
def normalize_query(...):
    # Implementation
    pass
```

---

## Observability

### Structured Logging

```python
import structlog

logger = structlog.get_logger()

async def process_query(query_input: QueryInput):
    logger.info(
        "query_processing_started",
        query_id=query_input.query_id,
        user_id=query_input.user_id,
        query_length=len(query_input.text),
        has_code=query_input.has_code
    )
    
    # ... processing ...
    
    logger.info(
        "query_processing_completed",
        query_id=query_input.query_id,
        intent=intent.value,
        entities_found=len(entities),
        expansion_terms=len(expansion.expanded_terms),
        processing_time_ms=elapsed_ms
    )
```

### Tracing

```python
from opentelemetry import trace

tracer = trace.get_tracer(__name__)

async def process_query(query_input: QueryInput):
    with tracer.start_as_current_span("query_processing") as span:
        span.set_attribute("query.length", len(query_input.text))
        span.set_attribute("query.user_id", query_input.user_id)
        
        with tracer.start_as_current_span("intent_classification"):
            intent = await classify_intent(query_input)
            span.set_attribute("intent", intent.value)
        
        with tracer.start_as_current_span("entity_extraction"):
            entities = extract_entities(query_input)
            span.set_attribute("entities.count", len(entities))
        
        # ... more spans ...
```

---

## Best Practices

### 1. Query Quality Checks

```python
def validate_query(query: str) -> bool:
    """Validate query before processing"""
    if not query or not query.strip():
        return False
    if len(query) > 1000:  # Too long
        return False
    if len(query) < 3:  # Too short
        return False
    return True
```

### 2. Graceful Degradation

```python
async def process_query_safe(query_input: QueryInput) -> TransformedQuery:
    """Process query with fallbacks"""
    try:
        # Try full processing
        return await process_query_full(query_input)
    except Exception as e:
        logger.warning("Full processing failed, using fallback", error=str(e))
        # Fall back to simple processing
        return process_query_simple(query_input)
```

### 3. A/B Testing

```python
async def process_query_with_experiment(query_input: QueryInput):
    """A/B test different processing strategies"""
    variant = get_experiment_variant(query_input.user_id)
    
    if variant == "control":
        return await process_query_v1(query_input)
    else:
        return await process_query_v2(query_input)
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

