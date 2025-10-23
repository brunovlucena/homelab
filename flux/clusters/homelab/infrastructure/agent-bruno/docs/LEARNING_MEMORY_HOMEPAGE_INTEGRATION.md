# Continuous Learning, Long-term Memory & Homepage Integration

**[← Back to README](../README.md)** | **[Architecture](ARCHITECTURE.md)** | **[Session Management](SESSION_MANAGEMENT.md)** | **[Learning](LEARNING.md)** | **[Memory](MEMORY.md)**

---

## Table of Contents
1. [Overview](#overview)
2. [The Closed-Loop Learning System](#the-closed-loop-learning-system)
3. [Complete Integration Flow](#complete-integration-flow)
   - [3.1 Detailed RLHF Training Data Creation](#31-detailed-rlhf-training-data-creation)
4. [Key Integration Points](#key-integration-points)
5. [The Virtuous Cycle](#the-virtuous-cycle)
6. [Week-by-Week Improvement](#week-by-week-improvement)
7. [Implementation Examples](#implementation-examples)
8. [Monitoring & Metrics](#monitoring--metrics)

---

## Overview

Agent Bruno implements a **closed-loop learning system** where three components work together to continuously improve the AI:

1. **Homepage** (bruno.dev) - User interface where conversations happen
2. **Long-term Memory** (LanceDB) - Storage of all conversations, facts, and preferences
3. **Continuous Learning** (Fine-tuning pipeline) - Weekly process that improves the model

These components form a virtuous cycle:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                  THE CLOSED-LOOP LEARNING SYSTEM                        │
│                                                                         │
│  Homepage → Memory → Learning → Better Model → Better Responses → ...   │
└─────────────────────────────────────────────────────────────────────────┘
```

**Key Principle**: Users don't know they're training the AI - they just use the homepage, and every interaction makes Agent Bruno smarter! 🚀

---

## The Closed-Loop Learning System

### System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        USER INTERACTIONS                                │
│                                                                         │
│  👤 User on Homepage → 💬 Chat with Agent Bruno → ✅ Provide Feedback   │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             │ Dual Storage Path
                             │
                 ┌───────────┴────────────┐
                 ▼                        ▼
┌────────────────────────────┐  ┌────────────────────────────────────────┐
│   Long-term Memory         │  │   Feedback Storage                     │
│   (LanceDB)                │  │   (Postgres)                           │
│                            │  │                                        │
│ • Episodic Memory          │  │ • Explicit Feedback (👍👎)             │
│   (conversations)          │  │ • Implicit Signals (clicks, copy)      │
│ • Semantic Memory          │  │ • Quality Scores                       │
│   (learned facts)          │  │ • RLHF Training Data                   │
│ • Procedural Memory        │  │                                        │
│   (user preferences)       │  │                                        │
└────────────────────────────┘  └────────────────────────────────────────┘
                 │                        │
                 │                        │
                 └────────────┬───────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│              WEEKLY DATA CURATION (Every Sunday 2 AM)                   │
│                                                                         │
│  • JOIN episodic memory + feedback                                      │
│  • Calculate quality scores                                             │
│  • Filter high-quality interactions (score > 0.5)                       │
│  • Format for training (SFT + RLHF)                                     │
│  • Result: ~5K quality examples from ~50K total                         │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│              FINE-TUNING PIPELINE (Mac Studio - 6 hours)                │
│                                                                         │
│  • Load base model: llama3.1:8b                                         │
│  • Apply LoRA (only 0.1% parameters trained)                            │
│  • Train on curated data (3 epochs)                                     │
│  • Track with Weights & Biases                                          │
│  • Export to Ollama format                                              │
│  • Tag: llama3.1-agent-bruno:week-{N}                                   │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    A/B TESTING (24 hours minimum)                       │
│                                                                         │
│  • Control (90%): Current production model                              │
│  • Treatment (10%): New fine-tuned model                                │
│  • Measure: User satisfaction, quality, latency                         │
│  • Guardrails: Error rate, hallucinations                               │
│  • Decision: Promote if statistically significant improvement           │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│              GRADUAL ROLLOUT (10% → 25% → 50% → 100%)                   │
│                                                                         │
│  • 4-day process with monitoring at each stage                          │
│  • Automatic rollback on guardrail violations                           │
│  • Final: New model becomes production default                          │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                   IMPROVED HOMEPAGE EXPERIENCE                          │
│                                                                         │
│  ✅ More accurate responses                                             │
│  ✅ Better context awareness                                            │
│  ✅ Improved tone matching                                              │
│  ✅ Fewer hallucinations                                                │
│                                                                         │
│  → Higher user satisfaction → More positive feedback → Better data!     │
└─────────────────────────────────────────────────────────────────────────┘
                             │
                             │ ⟲ LOOP CONTINUES
                             └────────────────────────────────────────────┐
                                                                          │
                             ┌────────────────────────────────────────────┘
                             ▼
                    NEXT WEEK: EVEN BETTER MODEL
```

---

## Complete Integration Flow

### Step-by-Step Process

#### 1. User Interaction on Homepage

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Homepage (lucene.cloud)                          │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  💬 User: "How do I fix Loki crashes?"                          │    │
│  │                                                                 │    │
│  │  🤖 Agent Bruno: "Loki crashes are typically caused by:         │    │
│  │     1. Out of memory issues                                     │    │
│  │     2. Disk space exhaustion                                    │    │
│  │     3. Network connectivity problems                            │    │
│  │     [detailed explanation with code examples]"                  │    │
│  │                                                                 │    │
│  │  User Actions:                                                  │    │
│  │  ✅ Clicks 👍 Thumbs Up (Explicit Feedback)                     │    │
│  │  📋 Copies response to clipboard (Implicit Signal)              │    │
│  │  🔗 Clicks 2 citation links (Implicit Signal)                   │    │
│  │  📝 Asks follow-up: "How do I check disk space?" (Engagement)   │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  Data Collected:                                                        │
│  ───────────────                                                        │
│  {                                                                      │
│    "interaction_id": "int-789",                                         │
│    "user_id": "user-456",                                               │
│    "session_id": "session-abc123",                                      │
│    "query": "How do I fix Loki crashes?",                               │
│    "response": "Loki crashes are typically...",                         │
│    "trace_id": "trace-xyz789",                                          │
│    "model_version": "llama3.1-agent-bruno:week-41",                     │
│    "timestamp": "2025-10-22T10:30:00Z",                                 │
│    "explicit_feedback": {                                               │
│      "type": "thumbs_up",                                               │
│      "value": 1.0                                                       │
│    },                                                                   │
│    "implicit_signals": {                                                │
│      "copy_event": true,                                                │
│      "citation_clicks": 2,                                              │
│      "follow_up_asked": true,                                           │
│      "read_time_seconds": 45                                            │
│    }                                                                    │
│  }                                                                      │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 2A. Storage in Long-term Memory (LanceDB)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     LanceDB - Episodic Memory                           │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  INSERT INTO episodic_memory:                                   │    │
│  │  {                                                              │    │
│  │    "vector": embed("How do I fix Loki crashes? Loki..."),       │    │
│  │    "user_id": "user-456",                                       │    │
│  │    "session_id": "session-abc123",                              │    │
│  │    "timestamp": "2025-10-22T10:30:00Z",                         │    │
│  │    "query": "How do I fix Loki crashes?",                       │    │
│  │    "response": "Loki crashes are typically caused...",          │    │
│  │    "trace_id": "trace-xyz789",                                  │    │
│  │    "sentiment": 0.85,  # Derived from positive feedback         │    │
│  │    "topic": "kubernetes-troubleshooting",                       │    │
│  │    "model_version": "llama3.1-agent-bruno:week-41"              │    │
│  │  }                                                              │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  Purpose:                                                               │
│  ────────                                                               │
│  • Remember user's conversation history                                 │
│  • Provide context for future requests                                  │
│  • Enable personalized responses                                        │
│  • Track conversation flow and topics                                   │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                     LanceDB - Semantic Memory                           │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Extracted Facts (background process):                          │    │
│  │                                                                 │    │
│  │  1. {                                                           │    │
│  │      "vector": embed("Loki crashes caused by memory"),          │    │
│  │      "user_id": "user-456",                                     │    │
│  │      "entity_type": "troubleshooting_fact",                     │    │
│  │      "fact": "Loki crashes are often caused by OOM issues",     │    │
│  │      "confidence": 0.95,                                        │    │
│  │      "source": "conversation:trace-xyz789",                     │    │
│  │      "extracted_at": "2025-10-22T10:30:15Z"                     │    │
│  │    }                                                            │    │
│  │                                                                 │    │
│  │  2. {                                                           │    │
│  │      "vector": embed("user interested in Kubernetes"),          │    │
│  │      "user_id": "user-456",                                     │    │
│  │      "entity_type": "user_interest",                            │    │
│  │      "fact": "User frequently asks about Kubernetes issues",    │    │
│  │      "confidence": 0.88,                                        │    │
│  │      "frequency": 15  # 15th Kubernetes question                │    │
│  │    }                                                            │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  Purpose:                                                               │
│  ────────                                                               │
│  • Build knowledge graph about user                                     │
│  • Store learned facts from conversations                               │
│  • Enable smarter, context-aware responses                              │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                    LanceDB - Procedural Memory                          │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Updated Patterns (background process):                         │    │
│  │                                                                 │    │
│  │  UPSERT procedural_memory:                                      │    │
│  │  {                                                              │    │
│  │    "vector": embed("user prefers detailed technical answers"),  │    │
│  │    "user_id": "user-456",                                       │    │
│  │    "preference_type": "response_style",                         │    │
│  │    "preference_value": {                                        │    │
│  │      "detail_level": "high",                                    │    │
│  │      "include_code_examples": true,                             │    │
│  │      "include_citations": true,                                 │    │
│  │      "tone": "technical"                                        │    │
│  │    },                                                           │    │
│  │    "frequency": 23,  # Increment (22 → 23)                      │    │
│  │    "confidence": 0.92,                                          │    │
│  │    "last_observed": "2025-10-22T10:30:00Z",                     │    │
│  │    "first_observed": "2025-09-15T14:20:00Z"                     │    │
│  │  }                                                              │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  Purpose:                                                               │
│  ────────                                                               │
│  • Learn user's communication preferences                               │
│  • Adapt response style automatically                                   │
│  • Track behavioral patterns                                            │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 2B. Storage in Feedback Database (Postgres)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                  Postgres - feedback_events Table                       │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  INSERT INTO feedback_events:                                   │    │
│  │  {                                                              │    │
│  │    "event_id": "evt-123",                                       │    │
│  │    "interaction_id": "int-789",  # Links to LanceDB             │    │
│  │    "user_id": "user-456",                                       │    │
│  │    "session_id": "session-abc123",                              │    │
│  │    "feedback_type": "thumbs_up",                                │    │
│  │    "feedback_value": 1.0,  # Normalized -1 to +1                │    │
│  │    "timestamp": "2025-10-22T10:30:05Z",                         │    │
│  │    "model_version": "llama3.1-agent-bruno:week-41",             │    │
│  │    "metadata": {                                                │    │
│  │      "implicit_signals": {                                      │    │
│  │        "copy_event": true,                                      │    │
│  │        "citation_clicks": 2,                                    │    │
│  │        "follow_up_asked": true,                                 │    │
│  │        "read_time_seconds": 45,                                 │    │
│  │        "time_to_feedback_seconds": 5                            │    │
│  │      },                                                         │    │
│  │      "quality_signals": {                                       │    │
│  │        "response_length_tokens": 256,                           │    │
│  │        "context_used": true,                                    │    │
│  │        "rag_sources_count": 3,                                  │    │
│  │        "memory_context_used": true                              │    │
│  │      }                                                          │    │
│  │    },                                                           │    │
│  │    "corrected_response": null  # User didn't correct            │    │
│  │  }                                                              │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  Purpose:                                                               │
│  ────────                                                               │
│  • Track user satisfaction per interaction                              │
│  • Identify high-quality vs low-quality responses                       │
│  • Collect RLHF training data (preference pairs)                        │
│  • Enable A/B testing and quality analysis                              │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 3. Weekly Data Curation (Every Sunday 2 AM)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Data Curation Pipeline                               │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Step 1: JOIN Data Sources                                      │    │
│  │  ──────────────────────────                                     │    │
│  │                                                                 │    │
│  │  SELECT                                                         │    │
│  │    e.interaction_id,                                            │    │
│  │    e.query,                                                     │    │
│  │    e.response,                                                  │    │
│  │    e.user_id,                                                   │    │
│  │    e.model_version,                                             │    │
│  │    f.feedback_value,                                            │    │
│  │    f.metadata AS implicit_signals                               │    │
│  │  FROM lancedb.episodic_memory e                                 │    │
│  │  LEFT JOIN postgres.feedback_events f                           │    │
│  │    ON e.trace_id = f.interaction_id                             │    │
│  │  WHERE                                                          │    │
│  │    e.timestamp >= NOW() - INTERVAL '7 days'                     │    │
│  │                                                                 │    │
│  │  Result: ~50,000 interactions from past week                    │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Step 2: Calculate Quality Score                                │    │
│  │  ───────────────────────────────                                │    │
│  │                                                                 │    │
│  │  For each interaction:                                          │    │
│  │                                                                 │    │
│  │  quality_score = (                                              │    │
│  │    0.40 * explicit_feedback_score +      # Thumbs up/down       │    │
│  │    0.30 * implicit_feedback_score +      # Copy, clicks, etc.   │    │
│  │    0.15 * response_completeness_score +  # Token count 50-500   │    │
│  │    0.15 * context_usage_score            # Used RAG/memory      │    │
│  │  )                                                              │    │
│  │                                                                 │    │
│  │  Example Calculation:                                           │    │
│  │  ──────────────────                                             │    │
│  │  explicit_feedback_score = 1.0  (thumbs up)                     │    │
│  │  implicit_feedback_score = 0.8  (copy + citations + follow-up)  │    │
│  │  response_completeness = 1.0   (256 tokens, good length)        │    │
│  │  context_usage = 1.0           (used RAG + memory)              │    │
│  │                                                                 │    │
│  │  quality_score = 0.40*1.0 + 0.30*0.8 + 0.15*1.0 + 0.15*1.0      │    │
│  │                = 0.40 + 0.24 + 0.15 + 0.15                      │    │
│  │                = 0.94  ✅ HIGH QUALITY                          │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Step 3: Filter High Quality (score > 0.5)                      │    │
│  │  ─────────────────────────────────────────                      │    │
│  │                                                                 │    │
│  │  Additional Filters:                                            │    │
│  │  - No PII detected                                              │    │
│  │  - Complete conversation (not abandoned)                        │    │
│  │  - Response length 10-1000 tokens                               │    │
│  │  - Valid query-response pair                                    │    │
│  │  - No toxic language                                            │    │
│  │                                                                 │    │
│  │  Result: ~5,000 high-quality interactions (10% of total)        │    │
│  │                                                                 │    │
│  │  Quality Distribution:                                          │    │
│  │  - Excellent (0.8-1.0): 1,200 interactions (24%)                │    │
│  │  - Good (0.6-0.8):      2,300 interactions (46%)                │    │
│  │  - Acceptable (0.5-0.6): 1,500 interactions (30%)               │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Step 4: Format for Training                                    │    │
│  │  ──────────────────────────                                     │    │
│  │                                                                 │    │
│  │  A. Supervised Fine-Tuning (SFT) Format:                        │    │
│  │  {                                                              │    │
│  │    "prompt": "<|begin_of_text|>                                 │    │
│  │      <|start_header_id|>system<|end_header_id|>                 │    │
│  │      You are Agent Bruno, a helpful SRE assistant...<|eot_id|>  │    │
│  │      <|start_header_id|>context<|end_header_id|>                │    │
│  │      [RAG context + user memory]<|eot_id|>                      │    │
│  │      <|start_header_id|>user<|end_header_id|>                   │    │
│  │      How do I fix Loki crashes?<|eot_id|>                       │    │
│  │      <|start_header_id|>assistant<|end_header_id|>",            │    │
│  │    "completion": "Loki crashes are typically caused by...",     │    │
│  │    "metadata": {                                                │    │
│  │      "quality_score": 0.94,                                     │    │
│  │      "topic": "kubernetes-troubleshooting",                     │    │
│  │      "model_version": "llama3.1-agent-bruno:week-41"            │    │
│  │    }                                                            │    │
│  │  }                                                              │    │
│  │                                                                 │    │
│  │  B. RLHF Preference Pairs (see detailed explanation below):     │    │
│  │  {                                                              │    │
│  │    "prompt": "How do I fix Loki crashes?",                      │    │
│  │    "response_chosen": "Loki crashes are typically... [good]",   │    │
│  │    "response_rejected": "Check the logs [bad response]",        │    │
│  │    "metadata": {                                                │    │
│  │      "chosen_score": 0.94,                                      │    │
│  │      "rejected_score": 0.15,                                    │    │
│  │      "margin": 0.79  # High confidence preference               │    │
│  │    }                                                            │    │
│  │  }                                                              │    │
│  │                                                                 │    │
│  │  Save to: s3://agent-bruno/training-data/week-42.jsonl          │    │
│  │  Size: ~50MB (5K examples)                                      │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 3.1. Detailed RLHF Training Data Creation

```
┌─────────────────────────────────────────────────────────────────────────┐
│           HOW RLHF PREFERENCE PAIRS ARE CREATED FROM FEEDBACK           │
│                                                                         │
│  RLHF (Reinforcement Learning from Human Feedback) requires preference │
│  pairs: for the same prompt, which response is better?                 │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                     Method 1: Direct Preference Pairs                   │
│                     (Same Query, Different Responses)                   │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Scenario: Two users ask the same/similar question              │    │
│  │  ───────────────────────────────────────────────────────────    │    │
│  │                                                                 │    │
│  │  User A at 10:00 AM:                                            │    │
│  │  Query: "How do I fix Loki crashes?"                            │    │
│  │  Response A: "Check the logs and restart Loki."                 │    │
│  │  Feedback: 👎 Thumbs down (score: -1.0)                         │    │
│  │  Implicit signals: No copy, no clicks, abandoned chat           │    │
│  │  Quality score: 0.15 ❌                                          │    │
│  │                                                                 │    │
│  │  User B at 14:30 PM:                                            │    │
│  │  Query: "How to troubleshoot Loki crashes?"                     │    │
│  │  Response B: "Loki crashes are typically caused by three..."    │    │
│  │  Feedback: 👍 Thumbs up (score: +1.0)                           │    │
│  │  Implicit signals: Copied response, clicked 2 citations         │    │
│  │  Quality score: 0.94 ✅                                          │    │
│  │                                                                 │    │
│  │  RLHF Pair Creation:                                            │    │
│  │  ──────────────────                                             │    │
│  │  1. Detect semantic similarity (embedding cosine > 0.85)        │    │
│  │  2. Quality score difference > 0.3 (margin threshold)           │    │
│  │  3. Create preference pair:                                     │    │
│  │     {                                                           │    │
│  │       "prompt": "How do I fix Loki crashes?",                   │    │
│  │       "response_chosen": Response B (score: 0.94),              │    │
│  │       "response_rejected": Response A (score: 0.15),            │    │
│  │       "margin": 0.79,  # Strong preference signal               │    │
│  │       "source": "direct_comparison",                            │    │
│  │       "confidence": 0.95  # High confidence                     │    │
│  │     }                                                           │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                Method 2: Corrected Response Pairs                       │
│                (User Explicitly Corrects Bad Response)                  │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Scenario: User corrects Agent Bruno's response                 │    │
│  │  ────────────────────────────────────────────────────────       │    │
│  │                                                                 │    │
│  │  User interaction:                                              │    │
│  │  Query: "What's the default port for Prometheus?"               │    │
│  │  Response (AI): "Prometheus runs on port 8080 by default."      │    │
│  │  User feedback: 👎 + Correction provided                        │    │
│  │  User correction: "Actually, it's port 9090, not 8080."         │    │
│  │                                                                 │    │
│  │  Storage:                                                       │    │
│  │  ────────                                                       │    │
│  │  feedback_events table:                                         │    │
│  │  {                                                              │    │
│  │    "interaction_id": "int-456",                                 │    │
│  │    "feedback_type": "thumbs_down",                              │    │
│  │    "feedback_value": -1.0,                                      │    │
│  │    "corrected_response": "Prometheus runs on port 9090.",       │    │
│  │    "correction_type": "factual_error"                           │    │
│  │  }                                                              │    │
│  │                                                                 │    │
│  │  RLHF Pair Creation:                                            │    │
│  │  ──────────────────                                             │    │
│  │  {                                                              │    │
│  │    "prompt": "What's the default port for Prometheus?",         │    │
│  │    "response_rejected": "Prometheus runs on port 8080...",      │    │
│  │    "response_chosen": "Prometheus runs on port 9090...",        │    │
│  │    "margin": 1.0,  # Maximum margin (correction = strong signal)│    │
│  │    "source": "user_correction",                                 │    │
│  │    "confidence": 1.0,  # Highest confidence                     │    │
│  │    "correction_type": "factual_error"                           │    │
│  │  }                                                              │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│              Method 3: Model Version Comparison Pairs                   │
│              (A/B Test Results Create Preference Data)                  │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Scenario: During A/B test, track which model performs better   │    │
│  │  ───────────────────────────────────────────────────────────    │    │
│  │                                                                 │    │
│  │  A/B Test Setup:                                                │    │
│  │  - Model A (week-41): 90% traffic                               │    │
│  │  - Model B (week-42): 10% traffic                               │    │
│  │                                                                 │    │
│  │  Same Query, Different Models:                                  │    │
│  │  ─────────────────────────────                                  │    │
│  │  Query: "How to debug pod crash loops?"                         │    │
│  │                                                                 │    │
│  │  User 1 (gets Model A):                                         │    │
│  │  Response A: "Check pod logs using kubectl logs..."             │    │
│  │  Feedback: 😐 No feedback                                       │    │
│  │  Quality score: 0.42 (neutral)                                  │    │
│  │                                                                 │    │
│  │  User 2 (gets Model B):                                         │    │
│  │  Response B: "Pod crash loops are usually caused by: 1. OOM..." │    │
│  │  Feedback: 👍 Thumbs up + copied response                       │    │
│  │  Quality score: 0.87 (high)                                     │    │
│  │                                                                 │    │
│  │  RLHF Pair Creation:                                            │    │
│  │  ──────────────────                                             │    │
│  │  {                                                              │    │
│  │    "prompt": "How to debug pod crash loops?",                   │    │
│  │    "response_chosen": Response B (model: week-42),              │    │
│  │    "response_rejected": Response A (model: week-41),            │    │
│  │    "margin": 0.45,                                              │    │
│  │    "source": "ab_test_comparison",                              │    │
│  │    "confidence": 0.80,                                          │    │
│  │    "model_chosen": "week-42",                                   │    │
│  │    "model_rejected": "week-41"                                  │    │
│  │  }                                                              │    │
│  │                                                                 │    │
│  │  Note: This creates pairs across 100s of similar queries!       │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│           Method 4: Synthetic Negative Generation                       │
│           (Create Rejected Responses for High-Quality Responses)        │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │  Scenario: We have great responses but need contrasting pairs     │  │
│  │  ───────────────────────────────────────────────────────────      │  │
│  │                                                                   │  │
│  │  High-quality interaction:                                        │  │
│  │  Query: "How to scale Knative services?"                          │  │
│  │  Response (Good): "Knative autoscaling is controlled by three..." │  │
│  │  Feedback: 👍 Thumbs up (score: 0.92)                             │  │
│  │                                                                   │  │
│  │  Problem: No bad response to compare against!                     │  │
│  │  Solution: Generate synthetic negative examples                   │  │
│  │                                                                   │  │
│  │  Synthetic Negative Types:                                        │  │
│  │  ─────────────────────────                                        │  │
│  │  1. Too Vague:                                                    │  │
│  │     "You can scale Knative services using kubectl."               │  │
│  │                                                                   │  │
│  │  2. Incomplete:                                                   │  │
│  │     "Set minScale and maxScale annotations."                      │  │
│  │                                                                   │  │
│  │  3. Off-topic:                                                    │  │
│  │     "Knative is a Kubernetes extension for serverless..."         │  │
│  │                                                                   │  │
│  │  4. Factually Incorrect:                                          │  │
│  │     "Knative doesn't support autoscaling by default."             │  │
│  │                                                                   │  │
│  │  Generation Process:                                              │  │
│  │  ──────────────────                                               │  │
│  │  1. Take high-quality response (score > 0.8)                      │  │
│  │  2. Use LLM to generate degraded version:                         │  │
│  │     - Remove details                                              │  │
│  │     - Remove code examples                                        │  │
│  │     - Make response generic                                       │  │
│  │     - Introduce subtle errors                                     │  │
│  │  3. Human review (sample 10% for quality)                         │  │
│  │  4. Label as synthetic_negative                                   │  │
│  │                                                                   │  │
│  │  RLHF Pair Creation:                                              │  │
│  │  ──────────────────                                               │  │
│  │  {                                                                │  │
│  │    "prompt": "How to scale Knative services?",                    │  │
│  │    "response_chosen": "Knative autoscaling is controlled...",     │  │
│  │    "response_rejected": "You can scale using kubectl.",           │  │
│  │    "margin": 0.60,  # Estimated margin                            │  │
│  │    "source": "synthetic_negative",                                │  │
│  │    "confidence": 0.70,  # Lower confidence (synthetic)            │  │
│  │    "negative_type": "too_vague"                                   │  │
│  │  }                                                                │  │
│  │                                                                   │  │
│  │  Caution: Limit synthetic negatives to 20% of RLHF dataset!       │  │
│  └───────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                  Complete RLHF Data Curation Pipeline                   │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Step 1: Collect All Feedback Data (Past Week)                  │    │
│  │  ─────────────────────────────────────────────                  │    │
│  │                                                                 │    │
│  │  FROM postgres.feedback_events:                                 │    │
│  │  - 50,000 total interactions                                    │    │
│  │  - 12,000 with explicit feedback (👍/👎)                        │    │
│  │  - 4,500 with thumbs up (positive)                              │    │
│  │  - 2,800 with thumbs down (negative)                            │    │
│  │  - 320 with corrections                                         │    │
│  │  - 4,700 with implicit signals only                             │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Step 2: Create Preference Pairs by Method                      │    │
│  │  ──────────────────────────────────────────                     │    │
│  │                                                                 │    │
│  │  Method 1 - Direct Comparisons:                                 │    │
│  │  • Find similar queries (embedding similarity > 0.85)           │    │
│  │  • Group by semantic clusters                                   │    │
│  │  • Pairs with score delta > 0.3: ~800 pairs                     │    │
│  │  • Example cluster:                                             │    │
│  │    - "Fix Loki crashes" (5 variations)                          │    │
│  │    - Scores: [0.92, 0.88, 0.45, 0.23, 0.18]                     │    │
│  │    - Creates: 3 high-confidence pairs                           │    │
│  │      (0.92 vs 0.23, 0.88 vs 0.18, 0.92 vs 0.45)                 │    │
│  │                                                                 │    │
│  │  Method 2 - User Corrections:                                   │    │
│  │  • Explicit corrections: 320 interactions                       │    │
│  │  • All become pairs (highest confidence)                        │    │
│  │  • Results: ~320 pairs                                          │    │
│  │                                                                 │    │
│  │  Method 3 - A/B Test Comparisons:                               │    │
│  │  • Compare Model A vs Model B responses                         │    │
│  │  • Same queries answered by both models: ~2,100 pairs           │    │
│  │  • Filter by significant score delta (>0.2): ~450 pairs         │    │
│  │                                                                 │    │
│  │  Method 4 - Synthetic Negatives:                                │    │
│  │  • Take top 500 responses (score > 0.85)                        │    │
│  │  • Generate degraded versions                                   │    │
│  │  • Human review 50 samples (rejection rate: 12%)                │    │
│  │  • Results: ~440 synthetic pairs                                │    │
│  │                                                                 │    │
│  │  Total RLHF Pairs: 2,010 preference pairs                       │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Step 3: Quality Filtering                                      │    │
│  │  ──────────────────────────                                     │    │
│  │                                                                 │    │
│  │  Filters Applied:                                               │    │
│  │  1. Minimum margin > 0.2 (clear preference)                     │    │
│  │  2. No PII in either response                                   │    │
│  │  3. Both responses complete (not truncated)                     │    │
│  │  4. No toxic content                                            │    │
│  │  5. Response length: 10-1000 tokens each                        │    │
│  │  6. Valid query-response pairs                                  │    │
│  │                                                                 │    │
│  │  Results:                                                       │    │
│  │  - Before filtering: 2,010 pairs                                │    │
│  │  - After filtering: 1,720 pairs (85.6% pass rate)               │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Step 4: Confidence-Based Weighting                             │    │
│  │  ─────────────────────────────────                              │    │
│  │                                                                 │    │
│  │  Assign confidence scores based on source:                      │    │
│  │                                                                 │    │
│  │  Source                    Confidence  Count   Weight           │    │
│  │  ─────────────────────────────────────────────────────────      │    │
│  │  User corrections          1.00        315     1.0x             │    │
│  │  Direct comparison (>0.5)  0.95        520     1.0x             │    │
│  │  Direct comparison (0.3-0.5) 0.80      280     0.8x             │    │
│  │  A/B test comparison       0.75        385     0.7x             │    │
│  │  Synthetic negatives       0.60        220     0.5x             │    │
│  │                                                                 │    │
│  │  During training, higher confidence pairs are sampled more      │    │
│  │  frequently using weighted sampling.                            │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Step 5: Final RLHF Dataset Format                              │    │
│  │  ─────────────────────────────────                              │    │
│  │                                                                 │    │
│  │  Each pair saved as JSONL with metadata:                        │    │
│  │                                                                 │    │
│  │  {                                                              │    │
│  │    "pair_id": "pair-week42-001",                                │    │
│  │    "prompt": "How do I fix Loki crashes?",                      │    │
│  │    "context": "[RAG context if available]",                     │    │
│  │    "response_chosen": {                                         │    │
│  │      "text": "Loki crashes are typically caused by...",         │    │
│  │      "quality_score": 0.94,                                     │    │
│  │      "feedback_value": 1.0,                                     │    │
│  │      "implicit_signals": {                                      │    │
│  │        "copy_event": true,                                      │    │
│  │        "citation_clicks": 2                                     │    │
│  │      }                                                          │    │
│  │    },                                                           │    │
│  │    "response_rejected": {                                       │    │
│  │      "text": "Check the logs and restart Loki.",                │    │
│  │      "quality_score": 0.15,                                     │    │
│  │      "feedback_value": -1.0,                                    │    │
│  │      "implicit_signals": {                                      │    │
│  │        "copy_event": false,                                     │    │
│  │        "citation_clicks": 0                                     │    │
│  │      }                                                          │    │
│  │    },                                                           │    │
│  │    "margin": 0.79,  # chosen_score - rejected_score             │    │
│  │    "source": "direct_comparison",                               │    │
│  │    "confidence": 0.95,                                          │    │
│  │    "topic": "kubernetes-troubleshooting",                       │    │
│  │    "created_at": "2025-10-22T02:15:00Z",                        │    │
│  │    "week": 42                                                   │    │
│  │  }                                                              │    │
│  │                                                                 │    │
│  │  Save to: s3://agent-bruno/rlhf-data/week-42.jsonl              │    │
│  │  Size: ~28MB (1,720 pairs)                                      │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                        RLHF Training Algorithm                          │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Using TRL (Transformer Reinforcement Learning) Library         │    │
│  │  ──────────────────────────────────────────────────────         │    │
│  │                                                                 │    │
│  │  from trl import PPOTrainer, PPOConfig                          │    │
│  │  from transformers import AutoModelForCausalLM                  │    │
│  │                                                                 │    │
│  │  # 1. Load SFT fine-tuned model as starting point               │    │
│  │  model = AutoModelForCausalLM.from_pretrained(                  │    │
│  │      "./llama3.1-agent-bruno-sft"                               │    │
│  │  )                                                              │    │
│  │                                                                 │    │
│  │  # 2. Train reward model on preference pairs                    │    │
│  │  reward_model = train_reward_model(                             │    │
│  │      pairs=rlhf_dataset,                                        │    │
│  │      base_model=model,                                          │    │
│  │      epochs=1                                                   │    │
│  │  )                                                              │    │
│  │  # Reward model learns: chosen response → high score            │    │
│  │  #                      rejected response → low score           │    │
│  │                                                                 │    │
│  │  # 3. PPO training loop                                         │    │
│  │  ppo_config = PPOConfig(                                        │    │
│  │      learning_rate=1e-5,                                        │    │
│  │      batch_size=4,                                              │    │
│  │      mini_batch_size=1,                                         │    │
│  │      gradient_accumulation_steps=4                              │    │
│  │  )                                                              │    │
│  │                                                                 │    │
│  │  ppo_trainer = PPOTrainer(                                      │    │
│  │      model=model,                                               │    │
│  │      config=ppo_config,                                         │    │
│  │      reward_model=reward_model                                  │    │
│  │  )                                                              │    │
│  │                                                                 │    │
│  │  # Train for 1 epoch                                            │    │
│  │  for batch in rlhf_dataset:                                     │    │
│  │      # Generate responses                                       │    │
│  │      responses = model.generate(batch['prompts'])               │    │
│  │                                                                 │    │
│  │      # Get rewards from reward model                            │    │
│  │      rewards = reward_model(responses)                          │    │
│  │                                                                 │    │
│  │      # Update model to maximize reward                          │    │
│  │      ppo_trainer.step(responses, rewards)                       │    │
│  │                                                                 │    │
│  │  Result: Model learns to generate responses similar to          │    │
│  │          high-rated responses and avoid low-rated ones          │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                   Quality Metrics for RLHF Dataset                      │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Tracked Metrics:                                               │    │
│  │  ───────────────                                                │    │
│  │                                                                 │    │
│  │  1. Margin Distribution:                                        │    │
│  │     - High margin (>0.6): 680 pairs (39.5%)  ✅                 │    │
│  │     - Medium margin (0.3-0.6): 820 pairs (47.7%)  ✅            │    │
│  │     - Low margin (0.2-0.3): 220 pairs (12.8%)  ⚠️               │    │
│  │     - Average margin: 0.52                                      │    │
│  │                                                                 │    │
│  │  2. Source Diversity:                                           │    │
│  │     - User corrections: 18.3% (strong signal)                   │    │
│  │     - Direct comparisons: 46.5% (medium signal)                 │    │
│  │     - A/B test: 22.4% (medium signal)                           │    │
│  │     - Synthetic: 12.8% (weak signal)                            │    │
│  │                                                                 │    │
│  │  3. Topic Coverage:                                             │    │
│  │     - kubernetes-troubleshooting: 32%                           │    │
│  │     - observability: 18%                                        │    │
│  │     - gitops: 14%                                               │    │
│  │     - infrastructure: 22%                                       │    │
│  │     - general: 14%                                              │    │
│  │                                                                 │    │
│  │  4. Quality Assurance:                                          │    │
│  │     - Human review sample: 10% (172 pairs)                      │    │
│  │     - Agreement rate: 94.2% ✅ (reviewer agrees with labeling)  │    │
│  │     - Errors found: 10 pairs (corrected)                        │    │
│  │     - Ambiguous: 12 pairs (removed)                             │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 4. Fine-Tuning Pipeline (Mac Studio)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                  Fine-Tuning with LoRA (6 hours)                        │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Environment:                                                   │    │
│  │  ────────────                                                   │    │
│  │  Hardware: Mac Studio (M2 Ultra, 128GB RAM)                     │    │
│  │  Location: 192.168.0.16                                         │    │
│  │  Framework: PyTorch + HuggingFace + PEFT                        │    │
│  │  Tracking: Weights & Biases (wandb)                             │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Load Base Model:                                               │    │
│  │  ────────────────                                               │    │
│  │  model = AutoModelForCausalLM.from_pretrained(                  │    │
│  │      "meta-llama/Llama-3.1-8B",                                 │    │
│  │      torch_dtype=torch.float16                                  │    │
│  │  )                                                              │    │
│  │                                                                 │    │
│  │  Total Parameters: 8,030,261,248 (8B)                           │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Apply LoRA (Low-Rank Adaptation):                              │    │
│  │  ─────────────────────────────────                              │    │
│  │  lora_config = LoraConfig(                                      │    │
│  │      r=16,              # Rank (size of low-rank matrices)      │    │
│  │      lora_alpha=32,     # Scaling factor                        │    │
│  │      lora_dropout=0.05, # Dropout for regularization            │    │
│  │      target_modules=[   # Which layers to adapt                 │    │
│  │          "q_proj",      # Query projection                      │    │
│  │          "v_proj",      # Value projection                      │    │
│  │          "k_proj",      # Key projection                        │    │
│  │          "o_proj"       # Output projection                     │    │
│  │      ]                                                          │    │
│  │  )                                                              │    │
│  │                                                                 │    │
│  │  model = get_peft_model(model, lora_config)                     │    │
│  │                                                                 │    │
│  │  Trainable Parameters: 8,388,608 (8.4M)                         │    │
│  │  All Parameters: 8,030,261,248 (8B)                             │    │
│  │  Trainable %: 0.104% ✅ (only training 0.1%!)                   │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Training Configuration:                                        │    │
│  │  ───────────────────────                                        │    │
│  │  Learning rate: 2e-4                                            │    │
│  │  Epochs: 3                                                      │    │
│  │  Batch size: 4                                                  │    │
│  │  Gradient accumulation: 4 (effective batch = 16)                │    │
│  │  Warmup steps: 100                                              │    │
│  │  Weight decay: 0.01                                             │    │
│  │  Mixed precision: FP16                                          │    │
│  │  Optimizer: AdamW                                               │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Training Progress (tracked in wandb):                          │    │
│  │  ────────────────────────────────────────                       │    │
│  │                                                                 │    │
│  │  Epoch 1/3:                                                     │    │
│  │  ──────────                                                     │    │
│  │  Step    Loss    Perplexity  Learning Rate  GPU Mem             │    │
│  │  ────────────────────────────────────────────────────           │    │
│  │    0     2.134     8.45       0.0           45GB                │    │
│  │  100     1.842     6.31       2e-4          52GB                │    │
│  │  200     1.623     5.07       2e-4          52GB                │    │
│  │  300     1.489     4.43       2e-4          52GB                │    │
│  │                                                                 │    │
│  │  Epoch 2/3:                                                     │    │
│  │  ──────────                                                     │    │
│  │  400     1.234     3.44       1.8e-4        52GB                │    │
│  │  500     1.087     2.96       1.6e-4        52GB                │    │
│  │  600     0.982     2.67       1.4e-4        52GB                │    │
│  │                                                                 │    │
│  │  Epoch 3/3:                                                     │    │
│  │  ──────────                                                     │    │
│  │  700     0.876     2.40       1.2e-4        52GB                │    │
│  │  800     0.812     2.25       1.0e-4        52GB                │    │
│  │  900     0.764     2.15       0.5e-4        52GB                │    │
│  │                                                                 │    │
│  │  Final Training Loss: 0.764                                     │    │
│  │  Final Validation Loss: 0.821                                   │    │
│  │  Duration: 5 hours 47 minutes                                   │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Evaluation Metrics:                                            │    │
│  │  ───────────────────                                            │    │
│  │                                        Base    Fine-tuned       │    │
│  │  Perplexity (↓ better):               12.34      8.21  ✅       │    │
│  │  BLEU Score (↑ better):                0.42      0.58  ✅       │    │
│  │  ROUGE-L (↑ better):                   0.38      0.52  ✅       │    │
│  │  Factual Consistency (↑ better):       0.76      0.89  ✅       │    │
│  │  Hallucination Rate (↓ better):        2.3%      1.1%  ✅       │    │
│  │                                                                 │    │
│  │  ✅ All metrics improved!                                       │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Export to Ollama:                                              │    │
│  │  ────────────────                                               │    │
│  │  1. Merge LoRA weights with base model                          │    │
│  │  2. Convert to GGUF format (llama.cpp)                          │    │
│  │  3. Create Modelfile with chat template                         │    │
│  │  4. Push to Ollama registry                                     │    │
│  │                                                                 │    │
│  │  $ ollama create llama3.1-agent-bruno:week-42 \                 │    │
│  │      -f ./Modelfile                                             │    │
│  │                                                                 │    │
│  │  ✅ Model ready: llama3.1-agent-bruno:week-42                   │    │
│  │  Size: 4.7GB (quantized)                                        │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 5. A/B Testing (Canary Deployment)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    A/B Test Experiment Setup                            │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Experiment Configuration:                                      │    │
│  │  ─────────────────────────                                      │    │
│  │  Name: "week-42-finetuning"                                     │    │
│  │  Control (A): llama3.1-agent-bruno:week-41 (current prod)       │    │
│  │  Treatment (B): llama3.1-agent-bruno:week-42 (new)              │    │
│  │  Traffic Split: 90% A / 10% B                                   │    │
│  │  Duration: 24 hours minimum                                     │    │
│  │  Min Sample Size: 1,000 users per variant                       │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  User Assignment (Consistent Hashing):                          │    │
│  │  ────────────────────────────────────────                       │    │
│  │  def assign_model(user_id: str) -> str:                         │    │
│  │      hash_value = int(                                          │    │
│  │          hashlib.md5(user_id.encode()).hexdigest(), 16          │    │
│  │      )                                                          │    │
│  │                                                                 │    │
│  │      if (hash_value % 100) < 10:                                │    │
│  │          return "llama3.1-agent-bruno:week-42"  # Treatment     │    │
│  │      else:                                                      │    │
│  │          return "llama3.1-agent-bruno:week-41"  # Control       │    │
│  │                                                                 │    │
│  │  Properties:                                                    │    │
│  │  - Same user always gets same model (consistent)                │    │
│  │  - 10% of users see new model                                   │    │
│  │  - Random but deterministic distribution                        │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Metrics Collected (24 hours):                                  │    │
│  │  ────────────────────────────────                               │    │
│  │                                                                 │    │
│  │  PRIMARY METRICS:                                               │    │
│  │  ┌──────────────────┬─────────┬──────────┬────────┬─────────┐   │    │
│  │  │ Metric           │ Model A │ Model B  │ Delta  │ p-value │   │    │
│  │  ├──────────────────┼─────────┼──────────┼────────┼─────────┤   │    │
│  │  │ Thumbs up rate   │ 72.3%   │ 78.1%    │ +5.8%  │ 0.003   │   │    │
│  │  │ User satisfaction│ 4.2/5   │ 4.5/5    │ +0.3   │ 0.012   │   │    │
│  │  │ Response quality │ 8.1/10  │ 8.6/10   │ +0.5   │ 0.008   │   │    │
│  │  │ Follow-up rate   │ 45.2%   │ 51.8%    │ +6.6%  │ 0.015   │   │    │
│  │  └──────────────────┴─────────┴──────────┴────────┴─────────┘   │    │
│  │                                                                 │    │
│  │  ✅ All primary metrics significantly improved (p < 0.05)       │    │
│  │                                                                 │    │
│  │  GUARDRAIL METRICS:                                             │    │
│  │  ┌──────────────────┬─────────┬──────────┬────────┬────────┐    │    │
│  │  │ Metric           │ Model A │ Model B  │ Delta  │ Status │    │    │
│  │  ├──────────────────┼─────────┼──────────┼────────┼────────┤    │    │
│  │  │ P95 Latency      │ 1.82s   │ 1.91s    │ +0.09s │ ✅ OK  │    │    │
│  │  │ P99 Latency      │ 3.45s   │ 3.52s    │ +0.07s │ ✅ OK  │    │    │
│  │  │ Error rate       │ 0.84%   │ 0.61%    │ -0.23% │ ✅ OK  │    │    │
│  │  │ Hallucination    │ 2.12%   │ 1.83%    │ -0.29% │ ✅ OK  │    │    │
│  │  │ Timeout rate     │ 0.43%   │ 0.38%    │ -0.05% │ ✅ OK  │    │    │
│  │  └──────────────────┴─────────┴──────────┴────────┴────────┘    │    │
│  │                                                                 │    │
│  │  ✅ All guardrails within acceptable thresholds (<5% regression)│    │
│  │                                                                 │    │
│  │  Sample Sizes:                                                  │    │
│  │  - Model A: 9,123 users (45,615 requests)                       │    │
│  │  - Model B: 1,034 users (5,170 requests)                        │    │
│  │  - Total: 10,157 users (50,785 requests)                        │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Statistical Analysis:                                          │    │
│  │  ─────────────────────                                          │    │
│  │  Test: Two-sample t-test (independent samples)                  │    │
│  │  Confidence level: 95%                                          │    │
│  │  Significance threshold: p < 0.05                               │    │
│  │                                                                 │    │
│  │  Primary Metric Analysis (Thumbs up rate):                      │    │
│  │  ─────────────────────────────────────────                      │    │
│  │  H₀: μ_A = μ_B (null hypothesis: no difference)                 │    │
│  │  H₁: μ_A ≠ μ_B (alternative: there is a difference)             │    │
│  │                                                                 │    │
│  │  t-statistic: 3.124                                             │    │
│  │  p-value: 0.003                                                 │    │
│  │  Conclusion: REJECT H₀ (statistically significant)              │    │
│  │                                                                 │    │
│  │  Effect Size (Cohen's d): 0.42 (medium effect)                  │    │
│  │  95% CI for delta: [+2.1%, +9.5%]                               │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Decision: ✅ PROMOTE MODEL B                                   │    │
│  │  ─────────────────────────────                                  │    │
│  │  Justification:                                                 │    │
│  │  1. Statistically significant improvement in all primary metrics│    │
│  │  2. No guardrail violations                                     │    │
│  │  3. Error rate actually improved                                │    │
│  │  4. Hallucination rate decreased                                │    │
│  │  5. Latency increase negligible (<100ms)                        │    │
│  │                                                                 │    │
│  │  Next Step: Gradual rollout (10% → 25% → 50% → 100%)            │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 6. Gradual Rollout (4-Day Process)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Gradual Rollout Schedule                         │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Day 1: 10% → 25% Traffic                                       │    │
│  │  ───────────────────────────                                    │    │
│  │  Time: Monday 10:00 AM                                          │    │
│  │  Action: Update traffic split to 25%                            │    │
│  │                                                                 │    │
│  │  Monitoring (24 hours):                                         │    │
│  │  - Thumbs up rate: 77.8% ✅ (stable)                            │    │
│  │  - P95 latency: 1.89s ✅ (within threshold)                     │    │
│  │  - Error rate: 0.59% ✅ (improved)                              │    │
│  │  - No incidents, no rollback triggers                           │    │
│  │                                                                 │    │
│  │  Check: No regressions ✅ PASS                                  │    │
│  │  Decision: Proceed to next stage                                │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Day 2: 25% → 50% Traffic                                       │    │
│  │  ───────────────────────────                                    │    │
│  │  Time: Tuesday 10:00 AM                                         │    │
│  │  Action: Update traffic split to 50%                            │    │
│  │                                                                 │    │
│  │  Monitoring (24 hours):                                         │    │
│  │  - Thumbs up rate: 78.3% ✅ (improving)                         │    │
│  │  - User satisfaction: 4.5/5 ✅ (stable)                         │    │
│  │  - P95 latency: 1.90s ✅ (within threshold)                     │    │
│  │  - Error rate: 0.57% ✅ (continuing to improve)                 │    │
│  │                                                                 │    │
│  │  Check: Positive trend continues ✅ PASS                        │    │
│  │  Decision: Proceed to next stage                                │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Day 3: 50% → 100% Traffic                                      │    │
│  │  ────────────────────────────                                   │    │
│  │  Time: Wednesday 10:00 AM                                       │    │
│  │  Action: Update traffic split to 100%                           │    │
│  │                                                                 │    │
│  │  Monitoring (24 hours):                                         │    │
│  │  - Thumbs up rate: 78.6% ✅ (stable at higher level)            │    │
│  │  - User satisfaction: 4.6/5 ✅ (slight improvement)             │    │
│  │  - P95 latency: 1.91s ✅ (stable)                               │    │
│  │  - Error rate: 0.55% ✅ (best so far)                           │    │
│  │  - Hallucination rate: 1.79% ✅ (continuing to drop)            │    │
│  │                                                                 │    │
│  │  Check: Significant improvement confirmed ✅ PASS               │    │
│  │  Decision: Complete rollout                                     │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Day 4: Full Production (100%)                                  │    │
│  │  ─────────────────────────────                                  │    │
│  │  Time: Thursday 10:00 AM                                        │    │
│  │  Action: Make week-42 model the default                         │    │
│  │                                                                 │    │
│  │  Commands executed:                                             │    │
│  │  $ ollama tag llama3.1-agent-bruno:week-42 \                    │    │
│  │        llama3.1-agent-bruno:latest                              │    │
│  │                                                                 │    │
│  │  $ kubectl -n agent-bruno set env deployment/agent-core \       │    │
│  │        OLLAMA_MODEL=llama3.1-agent-bruno:latest                 │    │
│  │                                                                 │    │
│  │  🎉 Rollout Complete!                                           │    │
│  │  New fine-tuned model is now 100% in production.                │    │
│  │                                                                 │    │
│  │  Continuous Monitoring:                                         │    │
│  │  - Dashboard: https://grafana.bruno.dev/d/agent-bruno           │    │
│  │  - Alerts configured for any degradation                        │    │
│  │  - Automatic rollback if error rate > 1%                        │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Rollback Triggers (Automatic):                                 │    │
│  │  ──────────────────────────────                                 │    │
│  │  If ANY of these occur at ANY stage:                            │    │
│  │  - Error rate increases by > 5% (e.g., 0.8% → 1.2%)             │    │
│  │  - P95 latency increases by > 10% (e.g., 1.8s → 2.0s)           │    │
│  │  - Thumbs up rate decreases by > 5% (e.g., 72% → 68%)           │    │
│  │  - Hallucination rate increases by > 5% (e.g., 2.1% → 2.2%)     │    │
│  │                                                                 │    │
│  │  Action: Instant rollback to previous model                     │    │
│  │  Notification: PagerDuty alert to ML team                       │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 7. Improved Homepage Experience

```
┌─────────────────────────────────────────────────────────────────────────┐
│               Homepage Users Experience Better Responses                │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Improvements from Fine-tuning:                                 │    │
│  │  ─────────────────────────────                                  │    │
│  │                                                                 │    │
│  │  ✅ More Accurate Responses                                     │    │
│  │     - Learned from past corrections                             │    │
│  │     - Better understanding of technical terms                   │    │
│  │     - More precise troubleshooting steps                        │    │
│  │                                                                 │    │
│  │  ✅ Better Context Awareness                                    │    │
│  │     - Learned from high-quality conversations                   │    │
│  │     - Remembers user's technical level                          │    │
│  │     - Adapts to conversation flow                               │    │
│  │                                                                 │    │
│  │  ✅ Improved Tone Matching                                      │    │
│  │     - Learned from highly-rated responses                       │    │
│  │     - Matches user's communication style                        │    │
│  │     - Appropriate level of detail                               │    │
│  │                                                                 │    │
│  │  ✅ Fewer Hallucinations                                        │    │
│  │     - RLHF from negative feedback                               │    │
│  │     - Better fact-checking                                      │    │
│  │     - More conservative when uncertain                          │    │
│  │                                                                 │    │
│  │  ✅ More Relevant Citations                                     │    │
│  │     - Learned from citation clicks                              │    │
│  │     - Better source selection                                   │    │
│  │     - Higher quality references                                 │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Example Improvement:                                           │    │
│  │  ───────────────────                                            │    │
│  │                                                                 │    │
│  │  Query: "How do I fix Loki crashes?"                            │    │
│  │                                                                 │    │
│  │  Before (Week 41):                                              │    │
│  │  "Check the logs and restart Loki. Make sure you have           │    │
│  │   enough memory and disk space."                                │    │
│  │  👎 Generic, not helpful, no specifics                          │    │
│  │                                                                 │    │
│  │  After (Week 42):                                               │    │
│  │  "Loki crashes are typically caused by three main issues:       │    │
│  │                                                                 │    │
│  │   1. **Out of Memory (OOM)**: Check if Loki is hitting memory   │    │
│  │      limits. Run: `kubectl top pod -n loki` to verify.          │    │
│  │      Increase memory in your HelmRelease if needed.             │    │
│  │                                                                 │    │
│  │   2. **Disk Space**: Loki stores chunks locally before          │    │
│  │      shipping to object storage. Check PVC usage:               │    │
│  │      `kubectl get pvc -n loki`                                  │    │
│  │                                                                 │    │
│  │   3. **Object Storage Connection**: Verify MinIO/S3             │    │
│  │      connectivity. Check logs for 'NoSuchBucket' errors:        │    │
│  │      `kubectl logs -n loki deployment/loki-write -f`            │    │
│  │                                                                 │    │
│  │   [📚 Related runbooks: Loki Troubleshooting Guide]"            │    │
│  │  👍 Specific, actionable, includes commands and context         │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  User Satisfaction Impact:                                      │    │
│  │  ──────────────────────────                                     │    │
│  │  - Thumbs up rate: 72% → 78% (+6%)                              │    │
│  │  - Average rating: 4.2 → 4.6 (+0.4)                             │    │
│  │  - Follow-up questions: 45% → 52% (users more engaged)          │    │
│  │  - Session duration: 3.2min → 4.1min (users stay longer)        │    │
│  │  - Return rate (weekly): 38% → 47% (users come back)            │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Key Integration Points

### 1. Homepage ↔ Long-term Memory

**Every user interaction on homepage triggers memory storage:**

```python
# homepage/api/handlers/agent_bruno.py
async def chat_handler(request: ChatRequest) -> ChatResponse:
    """Handle chat request from homepage frontend."""
    
    # === STEP 1: Fetch context from LanceDB ===
    memory_context = await memory_retriever.get_context_for_request(
        user_id=request.user_id,
        session_id=request.session_id,
        current_query=request.message,
        lancedb=lancedb,
        redis=redis
    )
    # Returns:
    # - recent_history: Last 10 messages from Redis or LanceDB
    # - similar_episodes: 5 semantically similar past conversations
    # - relevant_facts: ~10 facts about user and topic
    # - user_preferences: Top 10 behavioral patterns
    
    # === STEP 2: Process request with full context ===
    response = await agent.process(
        message=request.message,
        context=memory_context
    )
    
    # === STEP 3: Store to Redis (sync, fast) ===
    await redis.setex(
        f"session:{request.session_id}",
        3600,  # 1 hour TTL
        json.dumps({
            "conversation_history": memory_context.recent_history + [
                {"role": "user", "content": request.message},
                {"role": "assistant", "content": response}
            ],
            "last_updated": datetime.utcnow().isoformat()
        })
    )
    
    # === STEP 4: Store to LanceDB (async, non-blocking) ===
    asyncio.create_task(
        update_long_term_memory(
            user_id=request.user_id,
            session_id=request.session_id,
            message=request.message,
            response=response,
            trace_id=request.trace_id
        )
    )
    
    return ChatResponse(
        response=response,
        session_id=request.session_id,
        trace_id=request.trace_id
    )
```

### 2. Homepage ↔ Feedback Collection

**Feedback widget on homepage captures user satisfaction:**

```typescript
// homepage/frontend/components/ChatFeedback.tsx
export function ChatFeedback({ interactionId }: Props) {
  const handleFeedback = async (feedbackType: string, value: number) => {
    // Send to backend
    await fetch('/api/feedback', {
      method: 'POST',
      body: JSON.stringify({
        interaction_id: interactionId,
        feedback_type: feedbackType,  // 'thumbs_up', 'thumbs_down', 'rating'
        value: value,  // -1 to +1
        timestamp: new Date().toISOString()
      })
    });
    
    // Track implicit signals
    trackImplicitSignals({
      copy_event: false,
      citation_clicks: citationClickCount,
      read_time_seconds: calculateReadTime(),
      follow_up_asked: hasFollowUp
    });
  };
  
  return (
    <div className="feedback-widget">
      <button onClick={() => handleFeedback('thumbs_up', 1.0)}>
        👍 Helpful
      </button>
      <button onClick={() => handleFeedback('thumbs_down', -1.0)}>
        👎 Not helpful
      </button>
      <StarRating onChange={(rating) => handleFeedback('rating', rating)} />
    </div>
  );
}
```

```python
# homepage/api/handlers/feedback.py
@app.post("/api/feedback")
async def collect_feedback(feedback: FeedbackRequest) -> dict:
    """Store feedback in Postgres for training."""
    
    # Store in database
    await db.execute("""
        INSERT INTO feedback_events (
            event_id,
            interaction_id,
            user_id,
            feedback_type,
            feedback_value,
            timestamp,
            metadata
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
    """,
        str(uuid.uuid4()),
        feedback.interaction_id,
        feedback.user_id,
        feedback.feedback_type,
        feedback.value,
        datetime.utcnow(),
        json.dumps({
            "implicit_signals": feedback.implicit_signals,
            "model_version": feedback.model_version,
            "platform": "homepage"
        })
    )
    
    # Update real-time metrics
    metrics.feedback_events_total.labels(
        feedback_type=feedback.feedback_type
    ).inc()
    
    metrics.feedback_score_distribution.observe(feedback.value)
    
    return {"status": "recorded"}
```

### 3. Long-term Memory + Feedback → Training Data

**Weekly curation job joins both data sources:**

```python
# curation/weekly_job.py
def curate_training_data() -> TrainingDataset:
    """
    Join episodic memory (LanceDB) with feedback (Postgres)
    to create high-quality training dataset.
    """
    
    # === STEP 1: Fetch data from both sources ===
    
    # Get all interactions from past week (LanceDB)
    episodic_records = lancedb.query(
        table="episodic_memory",
        filters=f"timestamp >= '{one_week_ago}'",
        limit=100000
    )
    
    # Get all feedback from past week (Postgres)
    feedback_records = await db.fetch("""
        SELECT 
            interaction_id,
            feedback_type,
            feedback_value,
            metadata
        FROM feedback_events
        WHERE timestamp >= $1
    """, one_week_ago)
    
    # === STEP 2: JOIN on trace_id / interaction_id ===
    
    interactions_with_feedback = []
    
    for episode in episodic_records:
        # Find matching feedback
        feedback = next(
            (f for f in feedback_records 
             if f['interaction_id'] == episode.trace_id),
            None
        )
        
        if feedback:
            interactions_with_feedback.append({
                'query': episode.query,
                'response': episode.response,
                'user_id': episode.user_id,
                'explicit_feedback': feedback['feedback_value'],
                'implicit_signals': feedback['metadata']['implicit_signals'],
                'context_used': episode.metadata.get('rag_results', []),
                'model_version': episode.model_version
            })
    
    # === STEP 3: Calculate quality scores ===
    
    quality_interactions = []
    
    for interaction in interactions_with_feedback:
        score = calculate_quality_score(interaction)
        
        if score > 0.5:  # High quality threshold
            quality_interactions.append({
                **interaction,
                'quality_score': score
            })
    
    # === STEP 4: Format for training ===
    
    dataset = {
        'sft': format_for_sft(quality_interactions),
        'rlhf': format_for_rlhf(quality_interactions),
        'metadata': {
            'total_interactions': len(interactions_with_feedback),
            'quality_interactions': len(quality_interactions),
            'filter_rate': len(quality_interactions) / len(interactions_with_feedback),
            'created_at': datetime.utcnow().isoformat()
        }
    }
    
    # === STEP 5: Save to S3 ===
    
    s3_client.upload_file(
        json.dumps(dataset),
        bucket='agent-bruno-training',
        key=f'datasets/week-{week_number}.jsonl'
    )
    
    return dataset


def calculate_quality_score(interaction: dict) -> float:
    """Calculate quality score from multiple signals."""
    
    score = 0.0
    
    # Factor 1: Explicit feedback (40%)
    if interaction['explicit_feedback']:
        score += 0.4 * interaction['explicit_feedback']
    
    # Factor 2: Implicit feedback (30%)
    implicit = interaction['implicit_signals']
    implicit_score = (
        0.4 if implicit.get('copy_event') else 0.0 +
        0.3 * min(implicit.get('citation_clicks', 0) / 2, 1.0) +
        0.3 if implicit.get('follow_up_asked') else 0.0
    )
    score += 0.3 * implicit_score
    
    # Factor 3: Response completeness (15%)
    token_count = len(interaction['response'].split())
    if 50 <= token_count <= 500:
        score += 0.15
    
    # Factor 4: Context usage (15%)
    if interaction['context_used']:
        score += 0.15
    
    return max(0.0, min(1.0, score))
```

### 4. Continuous Learning → Better Homepage Experience

**After fine-tuning and rollout, homepage automatically uses new model:**

```python
# agent-bruno/core/agent.py
class AgentBruno:
    """Main agent that homepage interacts with."""
    
    def __init__(self):
        # Ollama client points to latest model
        self.ollama = OllamaClient(
            host="http://192.168.0.16:11434",
            model="llama3.1-agent-bruno:latest"  # Points to week-42 after rollout
        )
        
    async def process(
        self,
        message: str,
        context: MemoryContext
    ) -> str:
        """Process user message with context."""
        
        # Build prompt with context
        prompt = self._build_prompt(message, context)
        
        # Generate response using latest fine-tuned model
        response = await self.ollama.generate(
            prompt=prompt,
            max_tokens=2048,
            temperature=0.7
        )
        
        # Users automatically benefit from:
        # - More accurate answers (learned from corrections)
        # - Better tone (learned from highly-rated responses)
        # - Fewer errors (RLHF from negative feedback)
        # - Better citations (learned from citation clicks)
        
        return response
```

---

## The Virtuous Cycle

### How Quality Improves Over Time

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      THE VIRTUOUS CYCLE                                 │
│                                                                         │
│  Better Model → Better Responses → Happier Users → More Positive        │
│  Feedback → Higher Quality Training Data → Better Model → ...           │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Week 1 (Baseline):                                             │   │
│  │  ──────────────────                                             │   │
│  │  - Model: llama3.1-base (no fine-tuning)                        │   │
│  │  - User satisfaction: 72%                                       │   │
│  │  - Quality interaction rate: 10%                                │   │
│  │  - From 50K interactions → 5K quality examples                  │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Week 2 (First Fine-tune):                                      │   │
│  │  ──────────────────────────                                     │   │
│  │  - Model: llama3.1-agent-bruno:week-1                           │   │
│  │  - User satisfaction: 75% (+3%)                                 │   │
│  │  - Quality interaction rate: 12.5% (+2.5%)                      │   │
│  │  - From 52K interactions → 6.5K quality examples                │   │
│  │  - Improvement: Better model → fewer poor responses             │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Week 3 (Compounding Improvements):                             │   │
│  │  ─────────────────────────────────                              │   │
│  │  - Model: llama3.1-agent-bruno:week-2                           │   │
│  │  - User satisfaction: 78% (+3%)                                 │   │
│  │  - Quality interaction rate: 14.5% (+2%)                        │   │
│  │  - From 55K interactions → 8K quality examples                  │   │
│  │  - Improvement: Higher baseline → more quality data             │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Week 4 (Accelerating):                                         │   │
│  │  ──────────────────────                                         │   │
│  │  - Model: llama3.1-agent-bruno:week-3                           │   │
│  │  - User satisfaction: 82% (+4%)                                 │   │
│  │  - Quality interaction rate: 17% (+2.5%)                        │   │
│  │  - From 58K interactions → 10K quality examples                 │   │
│  │  - Improvement: Users give better feedback to better model      │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Week 5 (Plateau Approaching):                                  │   │
│  │  ──────────────────────────                                     │   │
│  │  - Model: llama3.1-agent-bruno:week-4                           │   │
│  │  - User satisfaction: 85% (+3%)                                 │   │
│  │  - Quality interaction rate: 21% (+4%)                          │   │
│  │  - From 62K interactions → 13K quality examples                 │   │
│  │  - Improvement: Model approaching human-level quality           │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  Key Insight:                                                           │
│  ───────────                                                            │
│  Quality % increases because better model → fewer poor responses →      │
│  less negative feedback → higher quality data ratio                     │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Week-by-Week Improvement

### Metrics Tracking

| Week | Total Interactions | Quality Examples | Quality % | Model Version | User Satisfaction | Hallucination Rate | Thumbs Up Rate |
|------|-------------------|------------------|-----------|---------------|-------------------|--------------------|----------------|
| 1    | 50,000            | 5,000            | 10.0%     | base          | 72%               | 2.3%               | 68%            |
| 2    | 52,000            | 6,500            | 12.5%     | week-1-ft     | 75%               | 2.0%               | 72%            |
| 3    | 55,000            | 8,000            | 14.5%     | week-2-ft     | 78%               | 1.7%               | 76%            |
| 4    | 58,000            | 10,000           | 17.2%     | week-3-ft     | 82%               | 1.3%               | 81%            |
| 5    | 62,000            | 13,000           | 21.0%     | week-4-ft     | 85%               | 1.0%               | 85%            |
| 10   | 78,000            | 23,000           | 29.5%     | week-9-ft     | 91%               | 0.5%               | 92%            |
| 20   | 95,000            | 38,000           | 40.0%     | week-19-ft    | 94%               | 0.3%               | 95%            |

### Why Quality Percentage Increases

```
┌─────────────────────────────────────────────────────────────────────────┐
│              Understanding the Quality Percentage Growth                │
│                                                                         │
│  Week 1 (Baseline):                                                     │
│  ──────────────────                                                     │
│  50,000 interactions                                                    │
│  ├─ 👍 Good (score > 0.5): 5,000 (10%)                                  │
│  ├─ 😐 Neutral (score 0-0.5): 20,000 (40%)                              │
│  └─ 👎 Poor (score < 0): 25,000 (50%)                                   │
│                                                                         │
│  Problem: Base model gives many poor responses                          │
│                                                                         │
│  Week 2 (After First Fine-tune):                                        │
│  ───────────────────────────────                                        │
│  52,000 interactions                                                    │
│  ├─ 👍 Good (score > 0.5): 6,500 (12.5%)  ← Increased                   │
│  ├─ 😐 Neutral (score 0-0.5): 28,000 (54%)  ← Grew                      │
│  └─ 👎 Poor (score < 0): 17,500 (33.5%)  ← Decreased!                   │
│                                                                         │
│  Improvement: Fine-tuned model converts "Poor" → "Neutral" → "Good"     │
│                                                                         │
│  Week 5 (Compounding Effect):                                           │
│  ─────────────────────────────                                          │
│  62,000 interactions                                                    │
│  ├─ 👍 Good (score > 0.5): 13,000 (21%)  ← Doubled from Week 1!         │
│  ├─ 😐 Neutral (score 0-0.5): 42,000 (68%)                              │
│  └─ 👎 Poor (score < 0): 7,000 (11%)  ← 50% → 11% reduction!            │
│                                                                         │
│  Key: Better model → Fewer poor responses → More quality data           │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Implementation Examples

### Homepage Integration Example

```typescript
// homepage/frontend/pages/chat.tsx
import { useState, useEffect } from 'react';
import { ChatWidget } from '@/components/ChatWidget';
import { FeedbackWidget } from '@/components/FeedbackWidget';

export default function ChatPage() {
  const [sessionId, setSessionId] = useState<string>();
  const [messages, setMessages] = useState<Message[]>([]);
  
  useEffect(() => {
    // Initialize session
    const initSession = async () => {
      const session = await fetch('/api/chat/session', {
        method: 'POST',
        body: JSON.stringify({ user_id: getUserId() })
      });
      setSessionId(session.id);
    };
    
    initSession();
  }, []);
  
  const handleSendMessage = async (message: string) => {
    // Send to Agent Bruno
    const response = await fetch('/api/chat', {
      method: 'POST',
      body: JSON.stringify({
        session_id: sessionId,
        user_id: getUserId(),
        message: message
      })
    });
    
    const data = await response.json();
    
    // Update UI
    setMessages([
      ...messages,
      { role: 'user', content: message },
      { role: 'assistant', content: data.response, id: data.interaction_id }
    ]);
    
    // Track implicit signals
    trackReadTime(data.interaction_id);
    trackCitationClicks(data.interaction_id);
  };
  
  return (
    <div className="chat-page">
      <ChatWidget 
        messages={messages}
        onSendMessage={handleSendMessage}
      />
      <FeedbackWidget 
        interactionId={messages[messages.length - 1]?.id}
      />
    </div>
  );
}
```

### Memory Storage Example

```python
# agent-bruno/core/memory.py
async def update_long_term_memory(
    user_id: str,
    session_id: str,
    message: str,
    response: str,
    trace_id: str
):
    """
    Background task to update all three memory types.
    Called asynchronously after response is sent to user.
    """
    
    try:
        # === 1. Episodic Memory (Conversation History) ===
        await lancedb.insert(
            table="episodic_memory",
            records=[{
                "vector": embed(f"{message} {response}"),
                "user_id": user_id,
                "session_id": session_id,
                "timestamp": datetime.utcnow(),
                "query": message,
                "response": response,
                "trace_id": trace_id,
                "sentiment": analyze_sentiment(response),
                "topic": classify_topic(message),
                "model_version": get_current_model_version()
            }]
        )
        
        # === 2. Semantic Memory (Extract Facts) ===
        facts = await extract_facts_with_llm(message, response)
        
        if facts:
            await lancedb.insert(
                table="semantic_memory",
                records=[{
                    "vector": embed(fact.text),
                    "user_id": user_id,
                    "entity_type": fact.entity_type,
                    "fact": fact.text,
                    "confidence": fact.confidence,
                    "source": f"conversation:{trace_id}",
                    "extracted_at": datetime.utcnow()
                } for fact in facts]
            )
        
        # === 3. Procedural Memory (Update Patterns) ===
        patterns = await analyze_interaction_patterns(user_id, message, response)
        
        for pattern in patterns:
            # Upsert (increment frequency if exists)
            await lancedb.upsert(
                table="procedural_memory",
                records=[{
                    "vector": embed(pattern.description),
                    "user_id": user_id,
                    "preference_type": pattern.type,
                    "preference_value": pattern.value,
                    "frequency": pattern.frequency + 1,
                    "confidence": pattern.confidence,
                    "last_observed": datetime.utcnow()
                }],
                on_conflict="user_id, preference_type"
            )
        
        logger.info(
            "Long-term memory updated",
            user_id=user_id,
            trace_id=trace_id,
            facts_extracted=len(facts),
            patterns_updated=len(patterns)
        )
        
    except Exception as e:
        logger.error(
            "Failed to update long-term memory",
            error=str(e),
            trace_id=trace_id
        )
        # Don't fail the user request - this is background processing
```

---

## Monitoring & Metrics

### Learning Loop Dashboard

```yaml
# grafana/dashboards/learning-loop.json
panels:
  - title: "Feedback Collection"
    metrics:
      - feedback_events_total (by type)
      - feedback_score_distribution
      - feedback_capture_rate (% of interactions with feedback)
    
  - title: "Training Pipeline"
    metrics:
      - training_jobs_total (success/failed)
      - training_duration_seconds
      - curated_data_quality_score
      - training_examples_count
    
  - title: "Model Performance"
    metrics:
      - model_quality_score (perplexity, BLEU, ROUGE)
      - user_satisfaction_by_model_version
      - hallucination_rate_trend
      - response_quality_score
    
  - title: "A/B Test Results"
    metrics:
      - active_experiments_count
      - experiment_traffic_split
      - primary_metric_delta
      - guardrail_status
    
  - title: "Virtuous Cycle"
    metrics:
      - quality_interaction_percentage_trend
      - user_satisfaction_trend
      - model_improvement_rate
```

### Alerts

```yaml
# alertmanager/rules/learning-loop.yaml
groups:
  - name: learning_loop
    rules:
      - alert: FineTuningJobFailed
        expr: training_jobs_total{status="failed"} > 0
        for: 5m
        annotations:
          summary: "Fine-tuning job failed"
          runbook: /runbooks/agent-bruno/finetuning-failure.md
      
      - alert: ModelQualityRegression
        expr: |
          (
            model_quality_score{version="latest"} 
            < 
            model_quality_score{version="previous"} * 0.95
          )
        for: 1h
        annotations:
          summary: "Model quality regression detected"
          action: "Consider rolling back to previous model"
      
      - alert: UserSatisfactionDrop
        expr: |
          rate(feedback_events_total{feedback_type="thumbs_down"}[1h])
          /
          rate(feedback_events_total[1h])
          > 0.3
        for: 2h
        annotations:
          summary: "User satisfaction dropping (>30% thumbs down)"
          action: "Investigate recent model changes"
```

---

## Summary

### The Complete Picture

1. **Homepage** = User interface where conversations happen
2. **Long-term Memory** = Storage layer (Redis + LanceDB) for conversations, facts, and preferences
3. **Continuous Learning** = Weekly fine-tuning pipeline that improves the model

### The Integration Flow

```
Homepage User → Conversation → Memory Storage (LanceDB + Postgres)
                                        ↓
                              Weekly Curation Job
                                        ↓
                              Training Data (5K quality examples)
                                        ↓
                              Fine-tuning (Mac Studio, 6 hours)
                                        ↓
                              A/B Testing (24 hours)
                                        ↓
                              Gradual Rollout (4 days)
                                        ↓
                              Better Model in Production
                                        ↓
                              Homepage Users Benefit
                                        ↓
                              More Positive Feedback
                                        ↓
                              Higher Quality Training Data
                                        ↓
                              (Loop continues...)
```

### Key Benefits

1. **Users don't know they're training the AI** - They just use the homepage naturally
2. **Automatic improvement** - No manual intervention needed
3. **Safe rollouts** - A/B testing and gradual deployment prevent regressions
4. **Measurable progress** - Track satisfaction, quality, and hallucinations over time
5. **Virtuous cycle** - Better model → happier users → better data → even better model

### The Magic

**Every conversation on homepage makes Agent Bruno smarter for everyone!** 🚀

---

**Document Version**: 1.0  
**Last Updated**: 2025-10-22  
**Owner**: AI/ML Team / Bruno

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

