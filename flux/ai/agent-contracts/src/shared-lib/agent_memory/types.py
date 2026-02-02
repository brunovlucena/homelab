"""
Memory type definitions following Nate B. Jones's multi-tiered memory pattern.

Memory Tiers:
1. Short-term Memory (Conversation) - Current interaction context
2. Working Memory - Current task state, goals, progress
3. Entity Memory - Structured data about named entities
4. User Memory - User-specific preferences and history
5. Long-term Memory - Accumulated knowledge and learnings
"""

from datetime import datetime, timezone
from enum import Enum
from typing import Any, Optional
from uuid import uuid4

from pydantic import BaseModel, Field


class MemoryType(str, Enum):
    """Types of memory in the multi-tiered system."""
    
    # Current conversation context (ephemeral, Redis)
    SHORT_TERM = "short_term"
    
    # Current task state, goals, requirements (session, Redis)
    WORKING = "working"
    
    # Structured data about entities (persistent, PostgreSQL)
    ENTITY = "entity"
    
    # User-specific preferences and history (persistent, PostgreSQL)
    USER = "user"
    
    # Accumulated knowledge and learnings (persistent, PostgreSQL)
    LONG_TERM = "long_term"
    
    # Episodic memory - specific events/interactions (persistent, PostgreSQL)
    EPISODIC = "episodic"


class MemoryEntry(BaseModel):
    """Base class for all memory entries."""
    
    id: str = Field(default_factory=lambda: str(uuid4()))
    memory_type: MemoryType
    agent_id: str
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    updated_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    expires_at: Optional[datetime] = None
    metadata: dict[str, Any] = Field(default_factory=dict)
    
    def is_expired(self) -> bool:
        """Check if this memory entry has expired."""
        if self.expires_at is None:
            return False
        return datetime.now(timezone.utc) > self.expires_at


class ConversationMemory(MemoryEntry):
    """
    Short-term conversation memory.
    
    Stores recent messages and context for the current conversation.
    TTL: Session-based (typically 1-24 hours)
    """
    
    memory_type: MemoryType = MemoryType.SHORT_TERM
    conversation_id: str
    user_id: Optional[str] = None
    messages: list[dict[str, Any]] = Field(default_factory=list)
    context: dict[str, Any] = Field(default_factory=dict)
    summary: Optional[str] = None  # LLM-generated summary for long conversations
    message_count: int = 0
    
    def add_message(self, role: str, content: str, metadata: Optional[dict] = None):
        """Add a message to the conversation."""
        self.messages.append({
            "role": role,
            "content": content,
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "metadata": metadata or {},
        })
        self.message_count = len(self.messages)
        self.updated_at = datetime.now(timezone.utc)
    
    def get_recent_messages(self, limit: int = 10) -> list[dict]:
        """Get the most recent messages."""
        return self.messages[-limit:] if self.messages else []
    
    def to_prompt_context(self, limit: int = 10) -> str:
        """Format messages for LLM prompt context."""
        recent = self.get_recent_messages(limit)
        lines = []
        for msg in recent:
            role = msg["role"].capitalize()
            content = msg["content"]
            lines.append(f"{role}: {content}")
        return "\n\n".join(lines)


class WorkingMemory(MemoryEntry):
    """
    Working memory for current task execution.
    
    Following Nate B. Jones's pattern:
    - Explicit goals
    - Requirements tracking
    - Constraints
    - State progression
    
    TTL: Task-based (until task completion)
    """
    
    memory_type: MemoryType = MemoryType.WORKING
    task_id: str
    session_id: Optional[str] = None
    
    # Goals - what the agent is trying to achieve
    goals: list[dict[str, Any]] = Field(default_factory=list)
    
    # Requirements - what must be satisfied
    requirements: list[dict[str, Any]] = Field(default_factory=list)
    
    # Constraints - boundaries the agent must respect
    constraints: list[dict[str, Any]] = Field(default_factory=list)
    
    # Current state - where we are in the task
    state: dict[str, Any] = Field(default_factory=dict)
    
    # Progress tracking
    progress: dict[str, Any] = Field(default_factory=lambda: {
        "steps_completed": 0,
        "steps_total": 0,
        "percentage": 0,
        "current_step": None,
        "blockers": [],
    })
    
    # Decision history - why we made certain choices
    decisions: list[dict[str, Any]] = Field(default_factory=list)
    
    # Artifacts produced during task
    artifacts: list[dict[str, Any]] = Field(default_factory=list)
    
    def add_goal(self, description: str, priority: int = 1, status: str = "pending"):
        """Add a goal to the working memory."""
        self.goals.append({
            "id": str(uuid4()),
            "description": description,
            "priority": priority,
            "status": status,
            "created_at": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)
    
    def add_requirement(self, description: str, mandatory: bool = True):
        """Add a requirement."""
        self.requirements.append({
            "id": str(uuid4()),
            "description": description,
            "mandatory": mandatory,
            "satisfied": False,
            "created_at": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)
    
    def add_constraint(self, description: str, hard: bool = True):
        """Add a constraint."""
        self.constraints.append({
            "id": str(uuid4()),
            "description": description,
            "hard": hard,  # Hard constraints cannot be violated
            "created_at": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)
    
    def record_decision(self, decision: str, reasoning: str, alternatives: list[str] = None):
        """Record a decision with reasoning (for explainability)."""
        self.decisions.append({
            "id": str(uuid4()),
            "decision": decision,
            "reasoning": reasoning,
            "alternatives": alternatives or [],
            "timestamp": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)
    
    def update_progress(self, step: str, completed: bool = False):
        """Update task progress."""
        self.progress["current_step"] = step
        if completed:
            self.progress["steps_completed"] += 1
        if self.progress["steps_total"] > 0:
            self.progress["percentage"] = int(
                (self.progress["steps_completed"] / self.progress["steps_total"]) * 100
            )
        self.updated_at = datetime.now(timezone.utc)
    
    def to_context(self) -> dict:
        """Export working memory as context for LLM."""
        return {
            "task_id": self.task_id,
            "goals": [g["description"] for g in self.goals if g["status"] == "pending"],
            "requirements": [r["description"] for r in self.requirements if not r["satisfied"]],
            "constraints": [c["description"] for c in self.constraints],
            "current_state": self.state,
            "progress": self.progress,
            "recent_decisions": self.decisions[-3:] if self.decisions else [],
        }


class EntityMemory(MemoryEntry):
    """
    Memory for structured data about named entities.
    
    Stores information about:
    - Users, customers, patients
    - Products, services
    - Locations, organizations
    - Custom domain entities
    
    TTL: Persistent (with optional refresh)
    """
    
    memory_type: MemoryType = MemoryType.ENTITY
    entity_type: str  # e.g., "user", "product", "patient", "restaurant_table"
    entity_id: str
    entity_name: Optional[str] = None
    
    # Core attributes of the entity
    attributes: dict[str, Any] = Field(default_factory=dict)
    
    # Relationships to other entities
    relationships: list[dict[str, Any]] = Field(default_factory=list)
    
    # Tags for categorization
    tags: list[str] = Field(default_factory=list)
    
    # Confidence score (0-1) for inferred data
    confidence: float = 1.0
    
    # Source of this information
    source: Optional[str] = None
    
    def add_relationship(self, relation_type: str, target_entity_id: str, metadata: dict = None):
        """Add a relationship to another entity."""
        self.relationships.append({
            "type": relation_type,
            "target_id": target_entity_id,
            "metadata": metadata or {},
            "created_at": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)
    
    def update_attribute(self, key: str, value: Any, confidence: float = 1.0):
        """Update an entity attribute."""
        self.attributes[key] = {
            "value": value,
            "confidence": confidence,
            "updated_at": datetime.now(timezone.utc).isoformat(),
        }
        self.updated_at = datetime.now(timezone.utc)


class UserMemory(MemoryEntry):
    """
    User-specific memory for preferences, history, and personalization.
    
    Stores:
    - Communication preferences
    - Interaction history
    - Learned preferences (from behavior)
    - Explicit settings
    
    TTL: Persistent (user-controlled)
    """
    
    memory_type: MemoryType = MemoryType.USER
    user_id: str
    
    # Explicit preferences set by user
    preferences: dict[str, Any] = Field(default_factory=dict)
    
    # Inferred preferences from behavior
    inferred_preferences: dict[str, Any] = Field(default_factory=dict)
    
    # Interaction statistics
    interaction_stats: dict[str, Any] = Field(default_factory=lambda: {
        "total_interactions": 0,
        "first_interaction": None,
        "last_interaction": None,
        "favorite_topics": [],
        "communication_style": None,
    })
    
    # Key facts about the user
    facts: list[dict[str, Any]] = Field(default_factory=list)
    
    # User's custom instructions for the agent
    custom_instructions: Optional[str] = None
    
    def record_interaction(self, interaction_type: str, metadata: dict = None):
        """Record a user interaction."""
        now = datetime.now(timezone.utc).isoformat()
        stats = self.interaction_stats
        stats["total_interactions"] += 1
        stats["last_interaction"] = now
        if stats["first_interaction"] is None:
            stats["first_interaction"] = now
        self.updated_at = datetime.now(timezone.utc)
    
    def add_fact(self, fact: str, source: str = "conversation", confidence: float = 0.8):
        """Add a fact about the user."""
        self.facts.append({
            "id": str(uuid4()),
            "fact": fact,
            "source": source,
            "confidence": confidence,
            "created_at": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)
    
    def set_preference(self, key: str, value: Any, explicit: bool = True):
        """Set a user preference."""
        if explicit:
            self.preferences[key] = value
        else:
            self.inferred_preferences[key] = {
                "value": value,
                "confidence": 0.7,
                "updated_at": datetime.now(timezone.utc).isoformat(),
            }
        self.updated_at = datetime.now(timezone.utc)


class DomainMemory(MemoryEntry):
    """
    Long-term domain knowledge memory.
    
    Stores:
    - Accumulated learnings
    - Domain-specific knowledge
    - Past task summaries
    - Best practices discovered
    
    TTL: Persistent (agent lifecycle)
    """
    
    memory_type: MemoryType = MemoryType.LONG_TERM
    domain: str  # e.g., "restaurant", "medical", "security"
    
    # Knowledge entries (facts, rules, patterns)
    knowledge: list[dict[str, Any]] = Field(default_factory=list)
    
    # Task history summaries
    task_history: list[dict[str, Any]] = Field(default_factory=list)
    
    # Learned patterns and best practices
    patterns: list[dict[str, Any]] = Field(default_factory=list)
    
    # Common errors and how to avoid them
    error_patterns: list[dict[str, Any]] = Field(default_factory=list)
    
    # Performance metrics over time
    performance: dict[str, Any] = Field(default_factory=lambda: {
        "tasks_completed": 0,
        "success_rate": 0.0,
        "avg_duration_ms": 0.0,
        "user_satisfaction": None,
    })
    
    def add_knowledge(self, content: str, category: str, source: str, confidence: float = 1.0):
        """Add a knowledge entry."""
        self.knowledge.append({
            "id": str(uuid4()),
            "content": content,
            "category": category,
            "source": source,
            "confidence": confidence,
            "created_at": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)
    
    def record_task_completion(
        self,
        task_id: str,
        summary: str,
        success: bool,
        duration_ms: float,
        learnings: list[str] = None,
    ):
        """Record a completed task."""
        self.task_history.append({
            "task_id": task_id,
            "summary": summary,
            "success": success,
            "duration_ms": duration_ms,
            "learnings": learnings or [],
            "completed_at": datetime.now(timezone.utc).isoformat(),
        })
        
        # Update performance metrics
        total = len(self.task_history)
        successes = sum(1 for t in self.task_history if t["success"])
        self.performance["tasks_completed"] = total
        self.performance["success_rate"] = successes / total if total > 0 else 0.0
        
        # Rolling average duration
        durations = [t["duration_ms"] for t in self.task_history[-100:]]
        self.performance["avg_duration_ms"] = sum(durations) / len(durations) if durations else 0.0
        
        self.updated_at = datetime.now(timezone.utc)
    
    def add_pattern(self, name: str, description: str, when_to_use: str, example: str = None):
        """Record a discovered pattern or best practice."""
        self.patterns.append({
            "id": str(uuid4()),
            "name": name,
            "description": description,
            "when_to_use": when_to_use,
            "example": example,
            "times_applied": 0,
            "created_at": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)
    
    def record_error(self, error_type: str, description: str, prevention: str, severity: str = "medium"):
        """Record an error pattern for future avoidance."""
        self.error_patterns.append({
            "id": str(uuid4()),
            "type": error_type,
            "description": description,
            "prevention": prevention,
            "severity": severity,
            "occurrences": 1,
            "created_at": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)
