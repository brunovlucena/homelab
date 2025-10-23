# 🎓 Continuous Learning Loop

**[← Back to README](../README.md)** | **[RAG](RAG.md)** | **[Memory](MEMORY.md)** | **[Roadmap](ROADMAP.md)**

---

## 🎯 Executive Summary

**What is this?** A continuous learning system that improves Agent Bruno through user feedback, automated fine-tuning, and A/B testing.

**Documentation Status**: ✅ **100% Complete** - Fully designed architecture with code examples  
**Implementation Status**: 🔴 **20% Complete** - Only 3 of 8 components working

| What's Working ✅ | What's Missing 🔴 |
|-------------------|-------------------|
| • Feedback schema designed | • No production feedback collection |
| • WandB configured locally | • No automated data curation |
| • LoRA tested on Mac Studio | • No training pipeline (Flyte) |
| | • No model registry/versioning |
| | • No A/B testing infrastructure |
| | • No automated deployment |

**Time to Production**: 12 weeks (3 months) with 1 engineer full-time

**Quick Wins Available** (1-3 days each):
- Feedback collection API
- Model versioning in Minio
- Manual A/B testing
- WandB dashboard

👉 **Start with**: [Implementation Gap Analysis](#-implementation-gap-analysis) to understand what needs to be built.

---

## 📋 Implementation Status

> **⚠️ IMPORTANT**: This document describes the **TARGET ARCHITECTURE** (100% documented).  
> **Current Implementation**: ~20% complete  
> See [Implementation Gap Analysis](#-implementation-gap-analysis) below.

### Current State (v1.0)

| Component | Status | Details |
|-----------|--------|---------|
| **Feedback Collection Schema** | ✅ **Designed** | Schema defined, not implemented in production |
| **WandB Integration** | ✅ **Configured** | Working on Mac Studio for local training |
| **LoRA Fine-tuning Strategy** | ✅ **Designed** | Strategy documented, tested locally |
| **Automated Data Curation** | 🔴 **Missing** | No pipeline exists |
| **Training Pipeline (Flyte)** | 🔴 **Missing** | No orchestration |
| **Model Registry** | 🔴 **Missing** | No versioning system |
| **A/B Testing Infrastructure** | 🔴 **Missing** | No traffic splitting |
| **Automated Model Deployment** | 🔴 **Missing** | Manual Ollama push only |

**Overall**: 20% complete (3 of 8 components implemented)

---

## 📚 Table of Contents

1. [Implementation Status](#-implementation-status) - **Start here to understand what's real vs. design**
2. [Architecture](#️-architecture) - Complete learning loop design
3. [Implementation Details](#-implementation-details) - Code examples (design only)
   - Feedback Collection
   - Training Data Curation
   - LoRA Fine-Tuning
   - RLHF
   - A/B Testing Framework
4. [Metrics & Monitoring](#-metrics--monitoring) - Planned metrics
5. [Automation & Scheduling](#-automation--scheduling) - Workflow automation design
6. [Best Practices](#-best-practices) - Guidelines for implementation
7. [Configuration](#-configuration) - Configuration schema
8. [Monitoring Dashboard](#-monitoring-dashboard) - Dashboard design
9. [Implementation Gap Analysis](#-implementation-gap-analysis) - **Critical: What's missing and how to build it**
   - What's Working (20%)
   - What's Missing (80%)
   - Implementation Roadmap (12 weeks)
   - Quick Wins
10. [References](#-references) - Papers, tools, and related docs

---

## Overview

Agent Bruno implements a continuous learning system that improves over time through user feedback, fine-tuning, and reinforcement learning from human feedback (RLHF). The system collects both explicit and implicit feedback signals, curates training data, and automatically fine-tunes models in a closed-loop process.

**This document describes the complete design**. For actual implementation status, see [ASSESSMENT.md](ASSESSMENT.md) Section 20.

---

## 🏗️ Architecture

> **📋 DESIGN SPECIFICATION**: The architecture below is fully documented (100%) but only 20% implemented.  
> Components marked 🔴 are not yet implemented. See [Implementation Gap Analysis](#-implementation-gap-analysis) for details.

### Continuous Learning Pipeline

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    Production Interactions                                  │
│                                                                             │
│  User Query → Agent Response → User Feedback → Model Improvement            │
│                                                                             │
│  [Thousands of interactions daily]                                          │
└────────────────────────────────┬────────────────────────────────────────────┘
                                 │
                                 ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                      Feedback Collection System                            │
│                                                                            │
│  ┌────────────────────────────────┐  ┌──────────────────────────────────┐  │
│  │  Explicit Feedback             │  │  Implicit Feedback               │  │
│  │  ┌──────────────────────────┐  │  │  ┌────────────────────────────┐  │  │
│  │  │ 👍 Thumbs Up: +1.0       │  │  │  │ Query Reformulation: -0.5  │  │  │
│  │  │ 👎 Thumbs Down: -1.0     │  │  │  │ Session Abandon: -0.8      │  │  │
│  │  │ ⭐ Rating (1-5): score   │  │  │  │ Quick Accept: +0.6         │  │  │
│  │  │ 📝 Correction: RLHF data │  │  │  │ Citation Click: +0.3       │  │  │
│  │  │ 💬 Follow-up: context    │  │  │  │ Copy Response: +0.4        │  │  │
│  │  └──────────────────────────┘  │  │  │ Long Read Time: +0.5       │  │  │
│  └────────────────────────────────┘  │  └────────────────────────────┘  │  │
│                                      └──────────────────────────────────┘  │
│  Storage: Postgres + Minio/S3                                              │
│  Schema: feedback_events table with JSONB data column                      │
└────────────────────────────────┬───────────────────────────────────────────┘
                                 │
                                 ▼
┌───────────────────────────────────────────────────────────────────────────┐
│                      Weekly Data Curation Job                             │
│  ┌───────────────────────────────────────────────────────────────────┐    │
│  │  Step 1: Filter Quality Interactions                              │    │
│  │  ┌────────────────────────────────────────────────────────────┐   │    │
│  │  │  Criteria:                                                 │   │    │
│  │  │  - Overall feedback score > 0.5                            │   │    │
│  │  │  - Complete conversation (not abandoned)                   │   │    │
│  │  │  - Response length 10-1000 tokens                          │   │    │
│  │  │  - No PII detected                                         │   │    │
│  │  │  - Valid query-response pair                               │   │    │
│  │  │                                                            │   │    │
│  │  │  Result: ~5K high-quality interactions/week                │   │    │
│  │  └────────────────────────────────────────────────────────────┘   │    │
│  │                                                                   │    │
│  │  Step 2: Format for Fine-tuning                                   │    │
│  │  ┌────────────────────────────────────────────────────────────┐   │    │
│  │  │  SFT (Supervised Fine-Tuning) Format:                      │   │    │
│  │  │  {                                                         │   │    │
│  │  │    "prompt": "<system>You are Agent Bruno...</system>      │   │    │
│  │  │               <context>...</context>                       │   │    │
│  │  │               <user>How to fix Loki?</user>",              │   │    │
│  │  │    "completion": "Loki crashes are caused by..."           │   │    │
│  │  │  }                                                         │   │    │
│  │  │                                                            │   │    │
│  │  │  RLHF Preference Pairs:                                    │   │    │
│  │  │  {                                                         │   │    │
│  │  │    "prompt": "...",                                        │   │    │
│  │  │    "response_good": "... (thumbs up)",                     │   │    │
│  │  │    "response_bad": "... (thumbs down)"                     │   │    │
│  │  │  }                                                         │   │    │
│  │  └────────────────────────────────────────────────────────────┘   │    │
│  │                                                                   │    │
│  │  Step 3: Data Augmentation                                        │    │
│  │  - Paraphrase queries for robustness                              │    │
│  │  - Add negative examples (low-score responses)                    │    │
│  │  - Balance dataset across topics                                  │    │
│  │                                                                   │    │
│  │  Step 4: Train/Val/Test Split                                     │    │
│  │  - Train: 80% (4K examples)                                       │    │
│  │  - Validation: 10% (500 examples)                                 │    │
│  │  - Test: 10% (500 examples)                                       │    │
│  └───────────────────────────────────────────────────────────────────┘    │
└────────────────────────────────┬──────────────────────────────────────────┘
                                 │
                                 ▼
┌───────────────────────────────────────────────────────────────────────────┐
│                      Fine-Tuning Pipeline (Flyte/Airflow)                 │
│  ┌───────────────────────────────────────────────────────────────────┐    │
│  │  Task 1: Data Preparation                                         │    │
│  │    - Load curated dataset from Minio/S3                           │    │
│  │    - Tokenize with base model tokenizer                           │    │
│  │    - Create data loaders (batch size: 4)                          │    │
│  │    - Validate data quality                                        │    │
│  │                                                                   │    │
│  │  Task 2: Model Training (LoRA)                                    │    │
│  │    ┌────────────────────────────────────────────────────────┐     │    │
│  │    │  Base Model: llama3.1:8b from Ollama                   │     │    │
│  │    │  Method: LoRA (Low-Rank Adaptation)                    │     │    │
│  │    │                                                        │     │    │
│  │    │  LoRA Config:                                          │     │    │
│  │    │  - rank (r): 16                                        │     │    │
│  │    │  - alpha: 32                                           │     │    │
│  │    │  - target_modules: [q_proj, v_proj, k_proj, o_proj]    │     │    │
│  │    │  - dropout: 0.05                                       │     │    │
│  │    │                                                        │     │    │
│  │    │  Training Config:                                      │     │    │
│  │    │  - learning_rate: 2e-4                                 │     │    │
│  │    │  - epochs: 3                                           │     │    │
│  │    │  - warmup_steps: 100                                   │     │    │
│  │    │  - weight_decay: 0.01                                  │     │    │
│  │    │  - gradient_accumulation: 4                            │     │    │
│  │    │  - mixed_precision: fp16                               │     │    │
│  │    │                                                        │     │    │
│  │    │  Hardware: Mac Studio (M2 Ultra, 128GB RAM)            │     │    │
│  │    │  Duration: ~6 hours for 4K examples                    │     │    │
│  │    └────────────────────────────────────────────────────────┘     │    │
│  │                                                                   │    │
│  │  Task 3: Validation & Evaluation                                  │    │
│  │    - Perplexity on validation set                                 │    │
│  │    - BLEU/ROUGE scores                                            │    │
│  │    - Human evaluation on 100 hold-out examples                    │    │
│  │    - Regression testing (ensure no performance drop)              │    │
│  │    - Factual consistency check                                    │    │
│  │                                                                   │    │
│  │  Task 4: Model Export                                             │    │
│  │    - Merge LoRA weights with base model                           │    │
│  │    - Convert to Ollama format (GGUF)                              │    │
│  │    - Quantize (optional, for faster inference)                    │    │
│  │    - Tag version: v1.{week}.{iteration}                           │    │
│  │    - Push to model registry (Ollama + wandb)                      │    │
│  └───────────────────────────────────────────────────────────────────┘    │
└────────────────────────────────┬──────────────────────────────────────────┘
                                 │
                                 ▼
┌───────────────────────────────────────────────────────────────────────────┐
│                  Weights & Biases (wandb) Tracking                        │
│  ┌────────────────────────────────────────────────────────────────────┐   │
│  │  Logged Metrics:                                                   │   │
│  │  - Training loss (per step)                                        │   │
│  │  - Validation loss (per epoch)                                     │   │
│  │  - Learning rate schedule                                          │   │
│  │  - Gradient norms                                                  │   │
│  │  - Perplexity                                                      │   │
│  │  - BLEU/ROUGE scores                                               │   │
│  │                                                                    │   │
│  │  Logged Artifacts:                                                 │   │
│  │  - Model checkpoints (every epoch)                                 │   │
│  │  - Training dataset (versioned)                                    │   │
│  │  - Evaluation results                                              │   │
│  │  - Configuration files                                             │   │
│  │                                                                    │   │
│  │  Experiment Comparison:                                            │   │
│  │  - Side-by-side metric comparison                                  │   │
│  │  - Hyperparameter parallel coordinates                             │   │
│  │  - Model version lineage                                           │   │
│  └────────────────────────────────────────────────────────────────────┘   │
└────────────────────────────────┬──────────────────────────────────────────┘
                                 │
                                 ▼
┌───────────────────────────────────────────────────────────────────────────┐
│                           A/B Testing                                     │
│  ┌────────────────────────────────────────────────────────────────────┐   │
│  │  Experiment: llama3.1-base vs llama3.1-ft-week42                   │   │
│  │                                                                    │   │
│  │  Traffic Split:                                                    │   │
│  │  ┌───────────────────┐              ┌──────────────────────┐       │   │
│  │  │  Model A: 90%     │              │  Model B: 10%        │       │   │
│  │  │  (Current Prod)   │              │  (New Fine-tuned)    │       │   │
│  │  │  llama3.1-base    │              │  llama3.1-ft-week42  │       │   │
│  │  └───────────────────┘              └──────────────────────┘       │   │
│  │                                                                    │   │
│  │  Primary Metrics:                                                  │   │
│  │  ┌────────────────────────────────────────────────────────────┐    │   │
│  │  │ Metric           │ Model A  │ Model B  │ Delta    │ p-value│    │   │
│  │  ├──────────────────┼──────────┼──────────┼──────────┼────────┤    │   │
│  │  │ Thumbs up rate   │ 72%      │ 78%      │ +6%  ✅  │ 0.003  │    │   │
│  │  │ User satisfaction│ 4.2/5    │ 4.5/5    │ +0.3 ✅  │ 0.012  │    │   │
│  │  │ Response quality │ 8.1/10   │ 8.6/10   │ +0.5 ✅  │ 0.008  │    │   │
│  │  └────────────────────────────────────────────────────────────┘    │   │
│  │                                                                    │   │
│  │  Guardrail Metrics:                                                │   │
│  │  ┌────────────────────────────────────────────────────────────┐    │   │
│  │  │ Metric           │ Model A  │ Model B  │ Delta    │ Status │    │   │
│  │  ├──────────────────┼──────────┼──────────┼──────────┼────────┤    │   │
│  │  │ P95 Latency      │ 1.8s     │ 1.9s     │ +0.1s ✅ │ OK     │    │   │
│  │  │ Error rate       │ 0.8%     │ 0.6%     │ -0.2% ✅ │ OK     │    │   │
│  │  │ Hallucination    │ 2.1%     │ 1.8%     │ -0.3% ✅ │ OK     │    │   │
│  │  └────────────────────────────────────────────────────────────┘    │   │
│  │                                                                    │   │
│  │  Decision: Model B shows statistically significant improvement     │   │
│  │  Action: Promote to 25% → 50% → 100% over 3 days                   │   │
│  └────────────────────────────────────────────────────────────────────┘   │
└────────────────────────────────┬──────────────────────────────────────────┘
                                 │
                                 ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                      Production Rollout (Gradual)                          │
│                                                                            │
│  Day 1: 10% → Monitor for 24h → ✅ No regressions                          │
│  Day 2: 25% → Monitor for 24h → ✅ Metrics improving                       │
│  Day 3: 50% → Monitor for 24h → ✅ Stable performance                      │
│  Day 4: 100% → Full rollout    → 🎉 New model in production                │
│                                                                            │
│  Rollback trigger: Any guardrail metric degrades by >5%                    │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## 📊 Implementation Details

> **⚠️ DESIGN ONLY**: All code examples below are design specifications, not implemented code.  
> Refer to [Implementation Status](#-implementation-status) for what's actually working.

### 1. Feedback Collection (🔴 NOT IMPLEMENTED)

#### Feedback Schema

```python
from enum import Enum
from dataclasses import dataclass
from datetime import datetime
from typing import Optional, Dict

class FeedbackType(Enum):
    THUMBS_UP = "thumbs_up"
    THUMBS_DOWN = "thumbs_down"
    RATING = "rating"  # 1-5 stars
    CORRECTION = "correction"  # User corrected response
    FOLLOW_UP = "follow_up"  # User asked follow-up
    ABANDONMENT = "abandonment"  # Session abandoned

@dataclass
class FeedbackEvent:
    """Represents a feedback event."""
    event_id: str
    interaction_id: str  # Links to the query/response
    user_id: str
    feedback_type: FeedbackType
    feedback_value: float  # Normalized -1 to +1
    timestamp: datetime
    metadata: Dict
    
    # For corrections (RLHF)
    corrected_response: Optional[str] = None
    correction_type: Optional[str] = None  # factual, style, tone

class FeedbackCollector:
    """Collects and stores user feedback."""
    
    def __init__(self, db_connection):
        self.db = db_connection
    
    def record_feedback(self, feedback: FeedbackEvent):
        """Record a feedback event."""
        # 1. Normalize feedback value
        normalized_value = self._normalize_feedback(feedback)
        
        # 2. Store in database
        self.db.execute("""
            INSERT INTO feedback_events (
                event_id, interaction_id, user_id, feedback_type,
                feedback_value, timestamp, metadata, corrected_response
            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
        """, (
            feedback.event_id,
            feedback.interaction_id,
            feedback.user_id,
            feedback.feedback_type.value,
            normalized_value,
            feedback.timestamp,
            json.dumps(feedback.metadata),
            feedback.corrected_response
        ))
        
        # 3. Update real-time metrics
        self._update_metrics(feedback)
    
    def _normalize_feedback(self, feedback: FeedbackEvent) -> float:
        """Normalize different feedback types to -1 to +1 scale."""
        if feedback.feedback_type == FeedbackType.THUMBS_UP:
            return 1.0
        elif feedback.feedback_type == FeedbackType.THUMBS_DOWN:
            return -1.0
        elif feedback.feedback_type == FeedbackType.RATING:
            # Rating is 1-5, normalize to -1 to +1
            return (feedback.feedback_value - 3) / 2
        elif feedback.feedback_type == FeedbackType.CORRECTION:
            return -0.5  # Indicates room for improvement
        elif feedback.feedback_type == FeedbackType.ABANDONMENT:
            return -0.8
        else:
            return 0.0
```

#### Implicit Feedback Signals

```python
class ImplicitFeedbackTracker:
    """Track implicit feedback signals."""
    
    def __init__(self):
        self.signals = []
    
    def track_user_behavior(self, interaction_id: str, events: List[Dict]):
        """Analyze user behavior for implicit feedback."""
        feedback_signals = []
        
        # Signal 1: Query reformulation (user rephrased query)
        if self._detect_reformulation(events):
            feedback_signals.append({
                "type": "reformulation",
                "score": -0.5,  # Original response wasn't good
                "reason": "User reformulated query"
            })
        
        # Signal 2: Quick acceptance (user engaged immediately)
        time_to_action = self._calculate_time_to_action(events)
        if time_to_action < 3:  # seconds
            feedback_signals.append({
                "type": "quick_accept",
                "score": 0.6,
                "reason": "User engaged quickly"
            })
        
        # Signal 3: Response copied (user found it useful)
        if self._detect_copy_event(events):
            feedback_signals.append({
                "type": "copy_response",
                "score": 0.4,
                "reason": "User copied response"
            })
        
        # Signal 4: Citation clicked (user verified source)
        citation_clicks = self._count_citation_clicks(events)
        if citation_clicks > 0:
            feedback_signals.append({
                "type": "citation_click",
                "score": 0.3 * citation_clicks,
                "reason": f"User clicked {citation_clicks} citations"
            })
        
        # Signal 5: Long read time (user thoroughly read response)
        read_time = self._calculate_read_time(events)
        expected_read_time = self._estimate_read_time(interaction_id)
        if read_time >= expected_read_time * 0.8:
            feedback_signals.append({
                "type": "thorough_read",
                "score": 0.5,
                "reason": "User read response thoroughly"
            })
        
        # Aggregate signals
        total_score = sum(s["score"] for s in feedback_signals)
        
        return {
            "interaction_id": interaction_id,
            "implicit_score": max(-1, min(1, total_score)),
            "signals": feedback_signals
        }
```

### 2. Training Data Curation (🔴 NOT IMPLEMENTED)

```python
class TrainingDataCurator:
    """Curate high-quality training data from feedback."""
    
    def __init__(self, db_connection, min_quality_score: float = 0.5):
        self.db = db_connection
        self.min_quality_score = min_quality_score
    
    def curate_weekly_dataset(self) -> Dict:
        """Curate training dataset from past week's interactions."""
        # 1. Fetch interactions with feedback
        interactions = self._fetch_feedback_interactions(days=7)
        
        # 2. Filter by quality
        quality_interactions = self._filter_by_quality(interactions)
        
        # 3. Format for different training methods
        dataset = {
            "sft": self._format_for_sft(quality_interactions),
            "rlhf": self._format_for_rlhf(quality_interactions),
            "metadata": {
                "total_interactions": len(interactions),
                "quality_interactions": len(quality_interactions),
                "filter_rate": len(quality_interactions) / len(interactions),
                "created_at": datetime.utcnow().isoformat()
            }
        }
        
        # 4. Save to Minio/S3
        self._save_dataset(dataset)
        
        return dataset
    
    def _filter_by_quality(self, interactions: List[Dict]) -> List[Dict]:
        """Filter interactions by quality criteria."""
        quality = []
        
        for interaction in interactions:
            # Calculate overall quality score
            score = self._calculate_quality_score(interaction)
            
            if score >= self.min_quality_score:
                quality.append({
                    **interaction,
                    "quality_score": score
                })
        
        return quality
    
    def _calculate_quality_score(self, interaction: Dict) -> float:
        """Calculate quality score for an interaction."""
        score = 0.0
        
        # Factor 1: Explicit feedback (40%)
        if interaction["explicit_feedback"]:
            score += 0.4 * interaction["explicit_feedback"]["normalized_value"]
        
        # Factor 2: Implicit feedback (30%)
        if interaction["implicit_feedback"]:
            score += 0.3 * interaction["implicit_feedback"]["score"]
        
        # Factor 3: Response completeness (15%)
        token_count = len(interaction["response"].split())
        if 50 <= token_count <= 500:
            score += 0.15
        
        # Factor 4: Context usage (15%)
        if interaction["context_used"]:
            score += 0.15
        
        return max(0, min(1, score))
    
    def _format_for_sft(self, interactions: List[Dict]) -> List[Dict]:
        """Format data for supervised fine-tuning."""
        sft_data = []
        
        for interaction in interactions:
            sft_data.append({
                "prompt": self._construct_prompt(interaction),
                "completion": interaction["response"],
                "metadata": {
                    "interaction_id": interaction["id"],
                    "quality_score": interaction["quality_score"],
                    "topic": interaction["topic"]
                }
            })
        
        return sft_data
    
    def _format_for_rlhf(self, interactions: List[Dict]) -> List[Dict]:
        """Format data for RLHF (preference pairs)."""
        rlhf_data = []
        
        # Find interactions with both good and bad responses
        for good_interaction in interactions:
            if good_interaction["quality_score"] < 0.7:
                continue
            
            # Find a bad response for the same or similar query
            bad_interaction = self._find_similar_bad_response(
                query=good_interaction["query"],
                exclude_id=good_interaction["id"]
            )
            
            if bad_interaction:
                rlhf_data.append({
                    "prompt": self._construct_prompt(good_interaction),
                    "response_chosen": good_interaction["response"],
                    "response_rejected": bad_interaction["response"],
                    "metadata": {
                        "chosen_score": good_interaction["quality_score"],
                        "rejected_score": bad_interaction["quality_score"],
                        "margin": good_interaction["quality_score"] - bad_interaction["quality_score"]
                    }
                })
        
        return rlhf_data
```

### 3. Fine-Tuning with LoRA (✅ TESTED LOCALLY, 🔴 NOT IN PRODUCTION)

```python
import torch
from transformers import AutoModelForCausalLM, AutoTokenizer
from peft import LoraConfig, get_peft_model, TaskType
import wandb

class LoRAFineTuner:
    """Fine-tune models using LoRA."""
    
    def __init__(
        self,
        base_model: str = "meta-llama/Llama-3.1-8B",
        lora_rank: int = 16,
        lora_alpha: int = 32
    ):
        self.base_model_name = base_model
        self.model = None
        self.tokenizer = None
        self.lora_rank = lora_rank
        self.lora_alpha = lora_alpha
        
        # Initialize wandb
        wandb.init(
            project="agent-bruno-finetuning",
            config={
                "base_model": base_model,
                "lora_rank": lora_rank,
                "lora_alpha": lora_alpha,
            }
        )
    
    def setup_model(self):
        """Load base model and apply LoRA."""
        # 1. Load base model
        self.model = AutoModelForCausalLM.from_pretrained(
            self.base_model_name,
            torch_dtype=torch.float16,
            device_map="auto"
        )
        
        self.tokenizer = AutoTokenizer.from_pretrained(self.base_model_name)
        
        # 2. Configure LoRA
        lora_config = LoraConfig(
            task_type=TaskType.CAUSAL_LM,
            r=self.lora_rank,
            lora_alpha=self.lora_alpha,
            lora_dropout=0.05,
            target_modules=["q_proj", "v_proj", "k_proj", "o_proj"],
            bias="none",
        )
        
        # 3. Apply LoRA to model
        self.model = get_peft_model(self.model, lora_config)
        
        # Print trainable parameters
        self.model.print_trainable_parameters()
        # Output: trainable params: 8.4M || all params: 8B || trainable: 0.1%
    
    def train(self, train_dataset, val_dataset, epochs: int = 3):
        """Train the model with LoRA."""
        from transformers import Trainer, TrainingArguments
        
        # Training arguments
        training_args = TrainingArguments(
            output_dir="./models/checkpoints",
            num_train_epochs=epochs,
            per_device_train_batch_size=4,
            per_device_eval_batch_size=4,
            gradient_accumulation_steps=4,
            learning_rate=2e-4,
            warmup_steps=100,
            weight_decay=0.01,
            logging_steps=10,
            eval_steps=100,
            save_steps=200,
            eval_strategy="steps",
            save_strategy="steps",
            load_best_model_at_end=True,
            fp16=True,
            report_to="wandb",
        )
        
        # Trainer
        trainer = Trainer(
            model=self.model,
            args=training_args,
            train_dataset=train_dataset,
            eval_dataset=val_dataset,
            tokenizer=self.tokenizer,
        )
        
        # Train
        trainer.train()
        
        # Save final model
        self.model.save_pretrained("./models/final")
        
        return trainer.state.log_history
    
    def export_to_ollama(self, output_path: str, model_name: str):
        """Export fine-tuned model to Ollama format."""
        # 1. Merge LoRA weights
        merged_model = self.model.merge_and_unload()
        
        # 2. Save in HuggingFace format
        merged_model.save_pretrained(output_path)
        self.tokenizer.save_pretrained(output_path)
        
        # 3. Convert to GGUF for Ollama
        # (using llama.cpp conversion tools)
        import subprocess
        subprocess.run([
            "python", "convert.py",
            output_path,
            "--outfile", f"{output_path}/model.gguf",
            "--outtype", "f16"
        ])
        
        # 4. Create Modelfile
        modelfile = f"""
FROM {output_path}/model.gguf

TEMPLATE \"\"\"
<|begin_of_text|><|start_header_id|>system<|end_header_id|>
{{ .System }}<|eot_id|>
<|start_header_id|>user<|end_header_id|>
{{ .Prompt }}<|eot_id|>
<|start_header_id|>assistant<|end_header_id|>
\"\"\"

PARAMETER temperature 0.7
PARAMETER top_p 0.9
PARAMETER stop <|eot_id|>
"""
        
        with open(f"{output_path}/Modelfile", "w") as f:
            f.write(modelfile)
        
        # 5. Push to Ollama
        subprocess.run([
            "ollama", "create", model_name,
            "-f", f"{output_path}/Modelfile"
        ])
        
        print(f"Model exported to Ollama as: {model_name}")
```

### 4. RLHF (Reinforcement Learning from Human Feedback) (🔴 NOT IMPLEMENTED)

```python
class RLHFTrainer:
    """Implement RLHF with DPO (Direct Preference Optimization)."""
    
    def __init__(self, base_model: str):
        self.base_model = base_model
        self.reward_model = None
    
    def train_reward_model(self, preference_dataset):
        """Train a reward model from preference pairs."""
        from transformers import AutoModelForSequenceClassification
        
        # 1. Load base model for reward modeling
        self.reward_model = AutoModelForSequenceClassification.from_pretrained(
            self.base_model,
            num_labels=1,  # Single reward score
            torch_dtype=torch.float16
        )
        
        # 2. Prepare preference pairs
        # Each example: (prompt, response_chosen, response_rejected)
        
        # 3. Train using Bradley-Terry preference model
        # Loss: -log(σ(r_chosen - r_rejected))
        # where σ is sigmoid and r is reward
        
        # Training code here...
        
        return self.reward_model
    
    def train_with_dpo(self, base_model, preference_dataset, beta: float = 0.1):
        """Train using Direct Preference Optimization."""
        # DPO simplifies RLHF by directly optimizing for preferences
        # without needing a separate reward model
        
        from trl import DPOTrainer, DPOConfig
        
        # 1. Load models
        model = AutoModelForCausalLM.from_pretrained(base_model)
        ref_model = AutoModelForCausalLM.from_pretrained(base_model)
        
        # 2. Configure DPO
        dpo_config = DPOConfig(
            beta=beta,  # KL penalty coefficient
            learning_rate=5e-7,
            num_train_epochs=1,
            gradient_accumulation_steps=4,
            per_device_train_batch_size=2,
            max_length=2048,
            max_prompt_length=1024,
        )
        
        # 3. Train
        trainer = DPOTrainer(
            model=model,
            ref_model=ref_model,
            args=dpo_config,
            train_dataset=preference_dataset,
            tokenizer=self.tokenizer,
        )
        
        trainer.train()
        
        return model
```

### 5. A/B Testing Framework (🔴 NOT IMPLEMENTED)

```python
class ABTestFramework:
    """Manage A/B tests for model comparison."""
    
    def __init__(self, db_connection):
        self.db = db_connection
        self.active_experiments = {}
    
    def create_experiment(
        self,
        name: str,
        model_a: str,
        model_b: str,
        traffic_split: float = 0.1,  # 10% to model B
        primary_metric: str = "user_satisfaction",
        guardrail_metrics: List[str] = ["latency_p95", "error_rate"]
    ):
        """Create a new A/B test experiment."""
        experiment = {
            "id": str(uuid.uuid4()),
            "name": name,
            "model_a": model_a,
            "model_b": model_b,
            "traffic_split": traffic_split,
            "primary_metric": primary_metric,
            "guardrail_metrics": guardrail_metrics,
            "status": "active",
            "start_time": datetime.utcnow(),
            "min_sample_size": 1000,  # Per variant
        }
        
        self.active_experiments[experiment["id"]] = experiment
        
        return experiment
    
    def assign_variant(self, user_id: str, experiment_id: str) -> str:
        """Assign user to experiment variant (consistent hashing)."""
        experiment = self.active_experiments[experiment_id]
        
        # Consistent hash based on user_id
        import hashlib
        hash_value = int(hashlib.md5(user_id.encode()).hexdigest(), 16)
        
        # Assign based on traffic split
        if (hash_value % 100) < (experiment["traffic_split"] * 100):
            return "model_b"
        else:
            return "model_a"
    
    def collect_metric(
        self,
        experiment_id: str,
        variant: str,
        metric_name: str,
        value: float
    ):
        """Collect a metric value for an experiment variant."""
        self.db.execute("""
            INSERT INTO ab_test_metrics (
                experiment_id, variant, metric_name, value, timestamp
            ) VALUES (%s, %s, %s, %s, %s)
        """, (experiment_id, variant, metric_name, value, datetime.utcnow()))
    
    def analyze_experiment(self, experiment_id: str) -> Dict:
        """Analyze experiment results."""
        from scipy import stats
        
        experiment = self.active_experiments[experiment_id]
        
        # Fetch metrics for both variants
        metrics_a = self._get_variant_metrics(experiment_id, "model_a")
        metrics_b = self._get_variant_metrics(experiment_id, "model_b")
        
        results = {
            "experiment_id": experiment_id,
            "sample_sizes": {
                "model_a": len(metrics_a),
                "model_b": len(metrics_b)
            },
            "metrics": {}
        }
        
        # Analyze primary metric
        primary = experiment["primary_metric"]
        a_values = [m["value"] for m in metrics_a if m["metric"] == primary]
        b_values = [m["value"] for m in metrics_b if m["metric"] == primary]
        
        # Statistical test (t-test)
        t_stat, p_value = stats.ttest_ind(a_values, b_values)
        
        results["metrics"][primary] = {
            "model_a_mean": np.mean(a_values),
            "model_b_mean": np.mean(b_values),
            "delta": np.mean(b_values) - np.mean(a_values),
            "delta_percent": ((np.mean(b_values) - np.mean(a_values)) / np.mean(a_values)) * 100,
            "p_value": p_value,
            "is_significant": p_value < 0.05,
            "winner": "model_b" if (np.mean(b_values) > np.mean(a_values) and p_value < 0.05) else "model_a"
        }
        
        # Check guardrail metrics
        for metric in experiment["guardrail_metrics"]:
            a_vals = [m["value"] for m in metrics_a if m["metric"] == metric]
            b_vals = [m["value"] for m in metrics_b if m["metric"] == metric]
            
            results["metrics"][metric] = {
                "model_a_mean": np.mean(a_vals),
                "model_b_mean": np.mean(b_vals),
                "regression_threshold": 0.05,  # 5% regression allowed
                "has_regression": (np.mean(b_vals) - np.mean(a_vals)) / np.mean(a_vals) > 0.05
            }
        
        # Overall decision
        results["decision"] = self._make_decision(results)
        
        return results
    
    def _make_decision(self, results: Dict) -> str:
        """Make decision on experiment."""
        primary_result = results["metrics"][list(results["metrics"].keys())[0]]
        
        # Check if winner
        if not primary_result["is_significant"]:
            return "inconclusive"
        
        # Check guardrails
        for metric_name, metric_data in results["metrics"].items():
            if metric_name != list(results["metrics"].keys())[0]:  # Skip primary
                if metric_data.get("has_regression", False):
                    return "reject_due_to_guardrail"
        
        if primary_result["winner"] == "model_b":
            return "promote_model_b"
        else:
            return "keep_model_a"
```

---

## 📈 Metrics & Monitoring

> **🔴 NOT IMPLEMENTED**: Metrics are defined but not collecting in production.

### Learning Loop Metrics

```python
# Feedback collection metrics
feedback_events_total = Counter(
    'feedback_events_total',
    'Total feedback events collected',
    ['feedback_type', 'user_tier']
)

feedback_score_distribution = Histogram(
    'feedback_score_distribution',
    'Distribution of feedback scores',
    buckets=[-1.0, -0.5, 0, 0.5, 1.0]
)

# Training metrics
training_jobs_total = Counter(
    'training_jobs_total',
    'Total training jobs run',
    ['status']  # success, failed, skipped
)

training_duration_seconds = Histogram(
    'training_duration_seconds',
    'Training job duration',
    buckets=[3600, 7200, 14400, 21600, 43200]  # 1h to 12h
)

# Model performance
model_quality_score = Gauge(
    'model_quality_score',
    'Model quality score from evaluation',
    ['model_version', 'metric']
)

# A/B test metrics
ab_test_conversions = Counter(
    'ab_test_conversions',
    'A/B test conversion events',
    ['experiment_id', 'variant', 'converted']
)
```

---

## 🔄 Automation & Scheduling

> **🔴 NOT IMPLEMENTED**: No automation exists. Training is manual on Mac Studio.

### Weekly Fine-Tuning Schedule

```yaml
# Flyte/Airflow DAG configuration
name: weekly_finetuning
schedule: "0 2 * * 0"  # Sunday 2 AM

tasks:
  - name: curate_data
    type: python
    script: curate_training_data.py
    outputs: [training_data.jsonl, metadata.json]
    
  - name: validate_data
    type: python
    script: validate_dataset.py
    inputs: [training_data.jsonl]
    outputs: [validation_report.json]
    
  - name: train_model
    type: python
    script: train_lora.py
    inputs: [training_data.jsonl]
    resources:
      gpu: 1
      memory: 64Gi
      cpu: 8
    timeout: 12h
    outputs: [model_checkpoint/]
    
  - name: evaluate_model
    type: python
    script: evaluate_model.py
    inputs: [model_checkpoint/]
    outputs: [eval_results.json]
    
  - name: export_to_ollama
    type: python
    script: export_ollama.py
    inputs: [model_checkpoint/]
    outputs: [model_version]
    
  - name: create_ab_test
    type: python
    script: create_experiment.py
    inputs: [model_version, eval_results.json]
    outputs: [experiment_id]

notifications:
  on_success:
    - slack: "#agent-bruno-ml"
  on_failure:
    - pagerduty: ml-oncall
    - slack: "#agent-bruno-ml"
```

### Gradual Rollout Automation (🔴 NOT IMPLEMENTED)

```python
class GradualRollout:
    """Automate gradual model rollout."""
    
    def __init__(self, ab_framework: ABTestFramework):
        self.ab = ab_framework
        self.rollout_schedule = [
            {"traffic": 0.10, "duration_hours": 24, "check": "no_regressions"},
            {"traffic": 0.25, "duration_hours": 24, "check": "positive_trend"},
            {"traffic": 0.50, "duration_hours": 24, "check": "significant_improvement"},
            {"traffic": 1.00, "duration_hours": 0, "check": "final_validation"},
        ]
    
    async def execute_rollout(self, experiment_id: str):
        """Execute gradual rollout with monitoring."""
        for stage in self.rollout_schedule:
            print(f"Rolling out to {stage['traffic'] * 100}% traffic...")
            
            # Update traffic split
            self.ab.update_traffic_split(experiment_id, stage["traffic"])
            
            # Wait and monitor
            if stage["duration_hours"] > 0:
                await asyncio.sleep(stage["duration_hours"] * 3600)
                
                # Check metrics
                results = self.ab.analyze_experiment(experiment_id)
                
                if not self._stage_check_passes(results, stage["check"]):
                    print(f"Rollout failed at {stage['traffic'] * 100}%")
                    self._rollback(experiment_id)
                    return False
            
            print(f"Stage {stage['traffic'] * 100}% successful")
        
        # Complete rollout
        print("Rollout complete! New model is now 100% in production.")
        return True
    
    def _stage_check_passes(self, results: Dict, check_type: str) -> bool:
        """Check if stage validation passes."""
        if check_type == "no_regressions":
            # No guardrail regressions
            return not any(
                m.get("has_regression", False)
                for m in results["metrics"].values()
            )
        elif check_type == "positive_trend":
            # Primary metric shows positive trend
            primary = list(results["metrics"].values())[0]
            return primary["delta"] > 0
        elif check_type == "significant_improvement":
            # Statistically significant improvement
            primary = list(results["metrics"].values())[0]
            return primary["is_significant"] and primary["delta"] > 0
        
        return True
```

---

## 🎯 Best Practices

### 1. Data Quality

- **Minimum feedback**: Require at least 1K interactions before fine-tuning
- **Balanced dataset**: Ensure diverse topics and query types
- **PII removal**: Automatic redaction before training
- **Deduplication**: Remove near-duplicate examples
- **Quality threshold**: Only use interactions with score > 0.5

### 2. Training Safety

- **Validation set**: Hold out 10% for unbiased evaluation
- **Regression testing**: Run comprehensive test suite on fine-tuned model
- **Hallucination check**: Verify factual consistency
- **Toxicity filtering**: Screen all training data
- **Version control**: Tag every model with git commit + dataset version

### 3. Rollout Safety

- **Gradual rollout**: 10% → 25% → 50% → 100%
- **Monitor guardrails**: Latency, error rate, hallucination rate
- **Automatic rollback**: Trigger on any guardrail violation
- **Canary analysis**: Statistical significance before promotion
- **Feature flags**: Ability to instantly switch models

---

## 🔧 Configuration

```yaml
learning:
  # Feedback Collection
  feedback:
    enable_explicit: true
    enable_implicit: true
    min_session_duration: 10  # seconds
    implicit_tracking_events:
      - reformulation
      - copy
      - citation_click
      - read_time
  
  # Data Curation
  curation:
    schedule: "weekly"
    min_quality_score: 0.5
    min_interactions: 1000
    max_interactions: 10000
    balance_topics: true
    augmentation_factor: 1.5
  
  # Fine-Tuning
  training:
    method: "lora"  # lora, full, qlora
    lora_rank: 16
    lora_alpha: 32
    learning_rate: 2e-4
    epochs: 3
    batch_size: 4
    gradient_accumulation: 4
    warmup_ratio: 0.1
    
  # RLHF
  rlhf:
    enable: true
    method: "dpo"  # dpo, ppo
    beta: 0.1  # KL penalty
    min_preference_pairs: 500
    
  # A/B Testing
  ab_testing:
    initial_traffic_split: 0.1
    min_sample_size: 1000
    confidence_level: 0.95
    rollout_stages: [0.1, 0.25, 0.5, 1.0]
    stage_duration_hours: 24
    
  # Automation
  automation:
    auto_finetune: true
    auto_experiment: true
    auto_rollout: false  # Require manual approval
    rollback_on_regression: true
```

---

## 📊 Monitoring Dashboard

> **🔴 NOT IMPLEMENTED**: No dedicated learning loop dashboard exists.

### Learning Loop Dashboard Panels

1. **Feedback Collection**
   - Feedback events per day (by type)
   - Average feedback score
   - Feedback capture rate

2. **Training Pipeline**
   - Training job status (success/failed)
   - Training duration trend
   - Model performance over time

3. **A/B Test Results**
   - Active experiments count
   - Win rate per model version
   - Rollout progress

4. **Model Quality**
   - Perplexity trend
   - User satisfaction by model version
   - Hallucination rate trend

---

## 🚧 Implementation Gap Analysis

### What's Working (20%)

#### ✅ 1. Feedback Collection Schema
```python
# Schema is defined in design
# Location: (design only, not implemented)
# Status: Ready for implementation
```

#### ✅ 2. WandB Integration
```bash
# Configured on Mac Studio
wandb login
wandb init --project agent-bruno-finetuning

# Status: Working for local training runs
# Location: ~/workspace/bruno/repos/flyte-test/_vault/sre-chatbot-finetune/
```

#### ✅ 3. LoRA Fine-Tuning Strategy
```python
# Tested locally with small datasets
# Location: repos/flyte-test/_vault/sre-chatbot-finetune/
# Config: lora_rank=16, lora_alpha=32
# Status: Proven on Mac Studio M2 Ultra
```

---

### What's Missing (80%)

#### 🔴 1. Automated Data Curation Pipeline

**Current State**: No automated curation  
**Required**:
- Weekly CronJob to extract feedback from Postgres
- Quality filtering (score > 0.5)
- PII removal
- Format conversion (SFT + RLHF)
- Upload to Minio/S3

**Implementation Effort**: 1 week  
**Priority**: P1 (blocks fine-tuning automation)

**Next Steps**:
```yaml
# 1. Create CronJob manifest
apiVersion: batch/v1
kind: CronJob
metadata:
  name: data-curation
  namespace: agent-bruno
spec:
  schedule: "0 2 * * 0"  # Sunday 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: curator
            image: agent-bruno/data-curator:v1.0.0
            env:
            - name: POSTGRES_URI
              valueFrom:
                secretKeyRef:
                  name: postgres-credentials
                  key: uri
            - name: MINIO_ENDPOINT
              value: "minio.minio.svc.cluster.local:9000"
```

#### 🔴 2. Training Pipeline (Flyte/Airflow)

**Current State**: Manual training on Mac Studio  
**Required**:
- Flyte workflow orchestration
- GPU resource allocation
- Checkpoint management
- Failure recovery
- Progress tracking

**Implementation Effort**: 2-3 weeks  
**Priority**: P1 (critical for automation)

**Next Steps**:
```python
# 1. Install Flyte on Kubernetes
# 2. Create workflow: repos/flyte-test/_vault/sre-chatbot-finetune/flyte_pipeline.py
# 3. Test end-to-end pipeline
# 4. Schedule weekly runs

# Example workflow (already exists, needs K8s deployment):
@workflow
def weekly_finetuning_workflow(dataset_path: str) -> str:
    # Task 1: Load data
    data = load_training_data(dataset_path=dataset_path)
    
    # Task 2: Train model
    model_path = train_lora_model(training_data=data)
    
    # Task 3: Evaluate
    metrics = evaluate_model(model_path=model_path)
    
    # Task 4: Export to Ollama
    model_version = export_to_ollama(model_path=model_path)
    
    return model_version
```

**Blockers**:
- No Flyte deployment in homelab cluster
- No GPU nodes in cluster (training on Mac Studio)
- No remote execution setup

#### 🔴 3. Model Registry and Versioning

**Current State**: Models in `~/sre-model-ollama/`, no versioning  
**Required**:
- Centralized model storage (Minio)
- Version tagging (v1.week.iteration)
- Metadata tracking (dataset, metrics, hyperparameters)
- Model lineage
- Rollback capability

**Implementation Effort**: 1 week  
**Priority**: P2 (needed before A/B testing)

**Next Steps**:
```python
# 1. Create model registry service
class ModelRegistry:
    def __init__(self, minio_client):
        self.minio = minio_client
        self.bucket = "agent-bruno-models"
    
    def register_model(
        self,
        model_path: str,
        version: str,
        metadata: Dict
    ):
        """Register a new model version."""
        # Upload model files
        self.minio.fput_object(
            bucket_name=self.bucket,
            object_name=f"{version}/model.gguf",
            file_path=f"{model_path}/model.gguf"
        )
        
        # Upload metadata
        self.minio.put_object(
            bucket_name=self.bucket,
            object_name=f"{version}/metadata.json",
            data=json.dumps(metadata)
        )
    
    def get_model(self, version: str) -> str:
        """Download model by version."""
        # Implementation here...
        pass

# 2. Add to Ollama service
# 3. Implement version tracking in Postgres
```

#### 🔴 4. A/B Testing Infrastructure

**Current State**: Single Ollama endpoint, no traffic splitting  
**Required**:
- Model router service
- User assignment (consistent hashing)
- Traffic splitting (10% → 100%)
- Metrics collection per variant
- Statistical analysis
- Automatic rollback

**Implementation Effort**: 2-3 weeks  
**Priority**: P1 (critical for safe rollouts)

**Next Steps**:
```yaml
# 1. Create model router service
apiVersion: apps/v1
kind: Deployment
metadata:
  name: model-router
  namespace: agent-bruno
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: router
        image: agent-bruno/model-router:v1.0.0
        env:
        - name: MODEL_A_ENDPOINT
          value: "http://ollama.ollama.svc.cluster.local:11434"
        - name: MODEL_B_ENDPOINT
          value: "http://ollama-canary.ollama.svc.cluster.local:11434"
        - name: TRAFFIC_SPLIT
          value: "0.1"  # 10% to model B

# 2. Update API service to route through model-router
# 3. Add metrics collection
# 4. Build analysis dashboard
```

**Design**:
```python
# Model Router Logic
class ModelRouter:
    def __init__(self, experiments: ABTestFramework):
        self.experiments = experiments
        self.model_endpoints = {
            "production": "http://ollama:11434",
            "canary": "http://ollama-canary:11434"
        }
    
    async def route_request(self, user_id: str, request: Dict) -> Dict:
        """Route request to appropriate model."""
        # Get active experiment
        experiment = self.experiments.get_active_experiment()
        
        if experiment:
            # Assign variant
            variant = self.experiments.assign_variant(user_id, experiment["id"])
            endpoint = self.model_endpoints["canary"] if variant == "model_b" else self.model_endpoints["production"]
        else:
            endpoint = self.model_endpoints["production"]
        
        # Forward request
        response = await self._call_model(endpoint, request)
        
        # Record metrics
        self._record_metrics(experiment, variant, response)
        
        return response
```

#### 🔴 5. Automated Model Deployment

**Current State**: Manual `ollama create` on Mac Studio  
**Required**:
- CI/CD pipeline for model deployment
- Kubernetes rollout (Deployment update)
- Health checks
- Gradual rollout (10% → 100%)
- Automatic rollback on failure

**Implementation Effort**: 1-2 weeks  
**Priority**: P1 (enables full automation)

**Next Steps**:
```yaml
# 1. Create GitHub Actions workflow
name: Deploy Fine-tuned Model

on:
  workflow_dispatch:
    inputs:
      model_version:
        description: 'Model version to deploy'
        required: true
      traffic_split:
        description: 'Initial traffic split (0.1 = 10%)'
        required: true
        default: '0.1'

jobs:
  deploy:
    runs-on: self-hosted  # Mac Studio
    steps:
      - name: Download model from registry
        run: |
          mc cp minio/agent-bruno-models/${{ inputs.model_version }}/model.gguf ./
      
      - name: Push to Ollama
        run: |
          ollama create agent-bruno:${{ inputs.model_version }} -f Modelfile
          ollama push agent-bruno:${{ inputs.model_version }}
      
      - name: Update Kubernetes deployment
        run: |
          kubectl set env deployment/ollama-canary \
            -n ollama \
            MODEL_VERSION=${{ inputs.model_version }}
          kubectl rollout status deployment/ollama-canary -n ollama
      
      - name: Create A/B experiment
        run: |
          python create_experiment.py \
            --model-a production \
            --model-b ${{ inputs.model_version }} \
            --traffic-split ${{ inputs.traffic_split }}
```

---

### Implementation Roadmap

#### Phase 1: Foundation (Weeks 1-2)
- [ ] Deploy Postgres feedback_events table
- [ ] Implement feedback collection API endpoints
- [ ] Create model registry in Minio
- [ ] Set up model versioning schema

#### Phase 2: Data Pipeline (Weeks 3-4)
- [ ] Build data curation script
- [ ] Create CronJob for weekly curation
- [ ] Implement PII removal
- [ ] Test end-to-end data pipeline

#### Phase 3: Training Automation (Weeks 5-7)
- [ ] Deploy Flyte on homelab cluster
- [ ] Port existing training code to Flyte workflow
- [ ] Configure GPU access (Mac Studio remote)
- [ ] Test automated training runs
- [ ] Integrate with wandb

#### Phase 4: A/B Testing (Weeks 8-10)
- [ ] Build model router service
- [ ] Implement user assignment logic
- [ ] Create metrics collection
- [ ] Build analysis dashboard
- [ ] Test traffic splitting

#### Phase 5: Deployment Automation (Weeks 11-12)
- [ ] Create GitHub Actions workflow
- [ ] Implement gradual rollout logic
- [ ] Add automatic rollback
- [ ] End-to-end integration test
- [ ] Production deployment

**Total Duration**: 12 weeks (3 months)  
**Effort**: 1 engineer full-time

---

### Quick Wins (Low-hanging fruit)

1. **Feedback Collection API** (2 days)
   - Add POST /api/feedback endpoint
   - Store in Postgres
   - Enable basic analytics

2. **Model Versioning** (3 days)
   - Tag models with v1.{week}.{iteration}
   - Upload to Minio
   - Track in metadata table

3. **Manual A/B Test** (3 days)
   - Deploy second Ollama instance (canary)
   - Use Kubernetes Service weights for traffic split
   - Manual metric comparison

4. **WandB Dashboard** (1 day)
   - Create project dashboard
   - Add key metrics
   - Share with team

---

## 📚 References

### Academic Papers
- [Fine-Tuning Language Models from Human Preferences](https://arxiv.org/abs/1909.08593)
- [Direct Preference Optimization (DPO)](https://arxiv.org/abs/2305.18290)
- [LoRA: Low-Rank Adaptation](https://arxiv.org/abs/2106.09685)
- [RLHF: Reinforcement Learning from Human Feedback](https://arxiv.org/abs/2203.02155)

### Tools & Frameworks
- [Weights & Biases Documentation](https://docs.wandb.ai/)
- [Flyte Documentation](https://docs.flyte.org/)
- [PEFT (Parameter-Efficient Fine-Tuning)](https://github.com/huggingface/peft)
- [TRL (Transformer Reinforcement Learning)](https://github.com/huggingface/trl)

### Related Documentation
- [ASSESSMENT.md](ASSESSMENT.md) - Section 20: ML Engineering Gaps
- [ROADMAP.md](ROADMAP.md) - Learning Loop Roadmap
- [flyte-test/_vault/sre-chatbot-finetune/](../../flyte-test/_vault/sre-chatbot-finetune/) - Existing training code

---

## 🎬 Summary & Next Steps

### What This Document Provides

✅ **Complete Architecture Design** (100%)
- End-to-end learning loop from feedback → training → deployment
- All components designed with code examples
- Metrics, monitoring, and automation fully specified
- Best practices and configuration documented

✅ **Honest Implementation Status** (20%)
- Clear distinction between design and reality
- 3 of 8 components working (feedback schema, WandB, LoRA)
- 5 critical components missing (curation, pipeline, registry, A/B, deployment)

✅ **Actionable Roadmap** (12 weeks)
- 5 implementation phases with clear tasks
- Effort estimates and priorities
- Blockers and dependencies identified
- Quick wins for immediate progress

### What You Need to Know

**If you're evaluating Agent Bruno:**
- Learning loop is **designed** but not operational
- Currently: Manual training on Mac Studio only
- No automated fine-tuning or A/B testing in production
- Plan requires 3 months of engineering effort

**If you're implementing this:**
1. Start with [Quick Wins](#quick-wins-low-hanging-fruit) (1 week)
2. Follow [Implementation Roadmap](#implementation-roadmap) (12 weeks)
3. Review [Blockers](#🔴-2-training-pipeline-flyteairflow) before starting
4. Reference existing code in `flyte-test/_vault/sre-chatbot-finetune/`

**If you're using this in production:**
- ⚠️ **DO NOT** assume automated fine-tuning is working
- ⚠️ **DO NOT** expect A/B testing infrastructure
- ✅ **DO** use for local model experimentation
- ✅ **DO** use as implementation blueprint

### Critical Gaps to Address First

**Priority 0 (Blockers):**
1. **Feedback Collection** - No data being collected from users
2. **Model Registry** - No centralized model versioning

**Priority 1 (Enablers):**
3. **Data Curation Pipeline** - Blocks training automation
4. **A/B Testing** - Blocks safe deployments
5. **Training Pipeline** - Enables weekly fine-tuning

### Success Metrics

When learning loop is fully implemented, expect:
- **Feedback**: 1K+ interactions/week with quality scores
- **Training**: Weekly fine-tuning runs (Sundays 2 AM)
- **A/B Tests**: 2-3 active experiments at any time
- **Deployment**: 10% → 100% gradual rollout over 3 days
- **Quality**: +5-10% user satisfaction per iteration

### Resources Required

**Engineering**:
- 1 ML engineer (full-time, 12 weeks)
- 0.5 backend engineer (API endpoints, 2 weeks)
- 0.25 SRE (Flyte deployment, 1 week)

**Infrastructure**:
- Flyte deployment on Kubernetes
- GPU access (Mac Studio or cloud)
- Model storage (100GB Minio)
- Postgres for feedback (10GB)

**Costs**:
- Storage: ~$5/month (Minio)
- Compute: Free (Mac Studio) or $50-100/month (cloud GPU)
- WandB: Free tier sufficient
- Total: **$5-100/month**

---

**Document Status**: 100% Designed, 20% Implemented  
**Last Updated**: October 22, 2025  
**Next Review**: November 22, 2025 (monthly during implementation)  
**Owner**: AI/ML Team  
**Implementation Lead**: Bruno

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

