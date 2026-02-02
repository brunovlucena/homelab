"""
Domain Memory Schemas - Structured memory representations for different agent types.

Following Nate B. Jones's Domain Memory Factory pattern:
- Each agent type has a specific schema
- Schema defines goals, requirements, constraints, and state structure
- Schema enables disciplined, accountable agent execution

The competitive advantage is in the meticulously designed domain-specific
memory schemas, not the AI models themselves.
"""

from datetime import datetime, timezone
from enum import Enum
from typing import Any, Optional, TypeVar
from uuid import uuid4

from pydantic import BaseModel, Field


class TaskStatus(str, Enum):
    """Task execution status."""
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    BLOCKED = "blocked"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


class GoalPriority(int, Enum):
    """Goal priority levels."""
    CRITICAL = 1
    HIGH = 2
    MEDIUM = 3
    LOW = 4


class AgentGoal(BaseModel):
    """
    A goal the agent is working towards.
    
    Goals are explicit, measurable objectives that guide agent behavior.
    """
    id: str = Field(default_factory=lambda: str(uuid4()))
    description: str
    priority: GoalPriority = GoalPriority.MEDIUM
    status: TaskStatus = TaskStatus.PENDING
    
    # Success criteria - how do we know the goal is achieved?
    success_criteria: list[str] = Field(default_factory=list)
    
    # Dependencies - other goals that must be completed first
    depends_on: list[str] = Field(default_factory=list)
    
    # Progress tracking
    progress_percentage: int = 0
    
    # Timestamps
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    
    def start(self):
        """Mark goal as started."""
        self.status = TaskStatus.IN_PROGRESS
        self.started_at = datetime.now(timezone.utc)
    
    def complete(self):
        """Mark goal as completed."""
        self.status = TaskStatus.COMPLETED
        self.progress_percentage = 100
        self.completed_at = datetime.now(timezone.utc)
    
    def fail(self, reason: str = None):
        """Mark goal as failed."""
        self.status = TaskStatus.FAILED
        self.completed_at = datetime.now(timezone.utc)


class AgentRequirement(BaseModel):
    """
    A requirement that must be satisfied.
    
    Requirements define what MUST be true for successful completion.
    """
    id: str = Field(default_factory=lambda: str(uuid4()))
    description: str
    mandatory: bool = True
    satisfied: bool = False
    
    # Verification - how to check if requirement is met
    verification_method: Optional[str] = None
    
    # Evidence - proof that requirement is satisfied
    evidence: Optional[str] = None
    
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    satisfied_at: Optional[datetime] = None
    
    def satisfy(self, evidence: str = None):
        """Mark requirement as satisfied."""
        self.satisfied = True
        self.evidence = evidence
        self.satisfied_at = datetime.now(timezone.utc)


class AgentConstraint(BaseModel):
    """
    A constraint the agent must respect.
    
    Constraints define boundaries that CANNOT be violated.
    """
    id: str = Field(default_factory=lambda: str(uuid4()))
    description: str
    
    # Hard constraints cannot be violated under any circumstances
    # Soft constraints can be violated with justification
    hard: bool = True
    
    # Category of constraint
    category: str = "general"  # e.g., "security", "privacy", "resource", "time"
    
    # Has this constraint been violated?
    violated: bool = False
    violation_reason: Optional[str] = None
    
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))


class AgentState(BaseModel):
    """
    Current state of the agent's execution.
    
    State captures WHERE the agent is in its task.
    """
    current_step: str = "initialized"
    previous_step: Optional[str] = None
    
    # Context data relevant to current step
    context: dict[str, Any] = Field(default_factory=dict)
    
    # Pending actions to take
    pending_actions: list[str] = Field(default_factory=list)
    
    # Blockers preventing progress
    blockers: list[str] = Field(default_factory=list)
    
    # Last error encountered
    last_error: Optional[str] = None
    
    updated_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    
    def transition(self, new_step: str, context_updates: dict = None):
        """Transition to a new step."""
        self.previous_step = self.current_step
        self.current_step = new_step
        if context_updates:
            self.context.update(context_updates)
        self.updated_at = datetime.now(timezone.utc)


class TaskProgress(BaseModel):
    """Progress tracking for a task."""
    steps_total: int = 0
    steps_completed: int = 0
    current_step_name: Optional[str] = None
    current_step_started_at: Optional[datetime] = None
    
    @property
    def percentage(self) -> int:
        if self.steps_total == 0:
            return 0
        return int((self.steps_completed / self.steps_total) * 100)
    
    def complete_step(self, next_step: str = None):
        """Complete current step and move to next."""
        self.steps_completed += 1
        if next_step:
            self.current_step_name = next_step
            self.current_step_started_at = datetime.now(timezone.utc)


class DomainMemorySchema(BaseModel):
    """
    Base schema for domain-specific memory.
    
    This is the structured representation that transforms agents
    from "forgetful entities into disciplined workers."
    
    Each agent type should extend this with domain-specific fields.
    """
    
    # Identity
    schema_id: str = Field(default_factory=lambda: str(uuid4()))
    schema_version: str = "1.0.0"
    agent_id: str
    agent_type: str
    domain: str
    
    # Core components (Nate B. Jones pattern)
    goals: list[AgentGoal] = Field(default_factory=list)
    requirements: list[AgentRequirement] = Field(default_factory=list)
    constraints: list[AgentConstraint] = Field(default_factory=list)
    state: AgentState = Field(default_factory=AgentState)
    progress: TaskProgress = Field(default_factory=TaskProgress)
    
    # Session tracking
    session_id: Optional[str] = None
    task_id: Optional[str] = None
    user_id: Optional[str] = None
    
    # Decision log - why did the agent do what it did?
    decisions: list[dict[str, Any]] = Field(default_factory=list)
    
    # Artifacts produced
    artifacts: list[dict[str, Any]] = Field(default_factory=list)
    
    # Timestamps
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    updated_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    
    def add_goal(self, description: str, priority: GoalPriority = GoalPriority.MEDIUM) -> AgentGoal:
        """Add a new goal."""
        goal = AgentGoal(description=description, priority=priority)
        self.goals.append(goal)
        self.updated_at = datetime.now(timezone.utc)
        return goal
    
    def add_requirement(self, description: str, mandatory: bool = True) -> AgentRequirement:
        """Add a new requirement."""
        req = AgentRequirement(description=description, mandatory=mandatory)
        self.requirements.append(req)
        self.updated_at = datetime.now(timezone.utc)
        return req
    
    def add_constraint(self, description: str, hard: bool = True, category: str = "general") -> AgentConstraint:
        """Add a new constraint."""
        constraint = AgentConstraint(description=description, hard=hard, category=category)
        self.constraints.append(constraint)
        self.updated_at = datetime.now(timezone.utc)
        return constraint
    
    def record_decision(self, decision: str, reasoning: str, alternatives: list[str] = None):
        """Record a decision with reasoning."""
        self.decisions.append({
            "id": str(uuid4()),
            "decision": decision,
            "reasoning": reasoning,
            "alternatives": alternatives or [],
            "state_at_decision": self.state.current_step,
            "timestamp": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)
    
    def add_artifact(self, name: str, artifact_type: str, content: Any, metadata: dict = None):
        """Record an artifact produced during execution."""
        self.artifacts.append({
            "id": str(uuid4()),
            "name": name,
            "type": artifact_type,
            "content": content,
            "metadata": metadata or {},
            "created_at": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)
    
    def get_active_goals(self) -> list[AgentGoal]:
        """Get goals that are pending or in progress."""
        return [g for g in self.goals if g.status in (TaskStatus.PENDING, TaskStatus.IN_PROGRESS)]
    
    def get_unsatisfied_requirements(self) -> list[AgentRequirement]:
        """Get requirements that haven't been satisfied yet."""
        return [r for r in self.requirements if not r.satisfied and r.mandatory]
    
    def check_constraints(self) -> list[AgentConstraint]:
        """Get any violated constraints."""
        return [c for c in self.constraints if c.violated]
    
    def to_context_prompt(self) -> str:
        """Generate a context prompt for the LLM."""
        lines = [
            f"## Current Task Context",
            f"Agent: {self.agent_type} ({self.domain})",
            f"Session: {self.session_id or 'N/A'}",
            f"Current Step: {self.state.current_step}",
            f"Progress: {self.progress.percentage}%",
            "",
            "### Goals",
        ]
        
        for goal in self.get_active_goals():
            status_icon = "🎯" if goal.status == TaskStatus.IN_PROGRESS else "⏳"
            lines.append(f"  {status_icon} [{goal.priority.name}] {goal.description}")
        
        unsatisfied = self.get_unsatisfied_requirements()
        if unsatisfied:
            lines.append("")
            lines.append("### Pending Requirements")
            for req in unsatisfied:
                lines.append(f"  ❗ {req.description}")
        
        if self.constraints:
            lines.append("")
            lines.append("### Constraints")
            for c in self.constraints:
                icon = "🔒" if c.hard else "⚠️"
                lines.append(f"  {icon} {c.description}")
        
        if self.state.blockers:
            lines.append("")
            lines.append("### Blockers")
            for blocker in self.state.blockers:
                lines.append(f"  🚫 {blocker}")
        
        recent_decisions = self.decisions[-3:] if self.decisions else []
        if recent_decisions:
            lines.append("")
            lines.append("### Recent Decisions")
            for d in recent_decisions:
                lines.append(f"  → {d['decision']}: {d['reasoning']}")
        
        return "\n".join(lines)


# =============================================================================
# Domain-Specific Schemas
# =============================================================================

class ChatAgentSchema(DomainMemorySchema):
    """Schema for conversational chat agents."""
    
    agent_type: str = "chat"
    domain: str = "conversation"
    
    # Chat-specific fields
    conversation_history_summary: Optional[str] = None
    user_preferences: dict[str, Any] = Field(default_factory=dict)
    topics_discussed: list[str] = Field(default_factory=list)
    sentiment_trend: Optional[str] = None  # positive, neutral, negative
    
    def add_topic(self, topic: str):
        """Track a topic that was discussed."""
        if topic not in self.topics_discussed:
            self.topics_discussed.append(topic)
            self.updated_at = datetime.now(timezone.utc)


class RestaurantAgentSchema(DomainMemorySchema):
    """Schema for restaurant service agents (host, waiter, chef, sommelier)."""
    
    agent_type: str = "restaurant"
    domain: str = "hospitality"
    
    # Restaurant-specific fields
    role: str = "waiter"  # host, waiter, chef, sommelier
    active_tables: list[dict[str, Any]] = Field(default_factory=list)
    current_orders: list[dict[str, Any]] = Field(default_factory=list)
    guest_preferences: dict[str, dict[str, Any]] = Field(default_factory=dict)
    service_events: list[dict[str, Any]] = Field(default_factory=list)
    
    def add_table(self, table_id: str, guests: int, status: str = "seated"):
        """Track an active table."""
        self.active_tables.append({
            "id": str(uuid4()),
            "table_id": table_id,
            "guests": guests,
            "status": status,
            "seated_at": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)
    
    def add_order(self, table_id: str, items: list[str], special_requests: str = None):
        """Track an order."""
        self.current_orders.append({
            "id": str(uuid4()),
            "table_id": table_id,
            "items": items,
            "special_requests": special_requests,
            "status": "pending",
            "ordered_at": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)


class MedicalAgentSchema(DomainMemorySchema):
    """Schema for medical agents (HIPAA-compliant)."""
    
    agent_type: str = "medical"
    domain: str = "healthcare"
    
    # Medical-specific fields (PHI is hashed/encrypted)
    current_patient_context: Optional[str] = None  # Hashed patient ID
    query_history: list[dict[str, Any]] = Field(default_factory=list)  # Sanitized
    clinical_context: dict[str, Any] = Field(default_factory=dict)
    
    # Compliance tracking
    hipaa_audit_trail: list[dict[str, Any]] = Field(default_factory=list)
    access_justifications: list[dict[str, Any]] = Field(default_factory=list)
    
    def record_access(self, patient_id_hash: str, data_type: str, justification: str):
        """Record data access for HIPAA compliance."""
        self.hipaa_audit_trail.append({
            "id": str(uuid4()),
            "patient_id_hash": patient_id_hash,
            "data_type": data_type,
            "justification": justification,
            "timestamp": datetime.now(timezone.utc).isoformat(),
        })
        self.updated_at = datetime.now(timezone.utc)


class SecurityAgentSchema(DomainMemorySchema):
    """Schema for security agents (red team, blue team)."""
    
    agent_type: str = "security"
    domain: str = "cybersecurity"
    
    # Security-specific fields
    team: str = "blue"  # red or blue
    active_threats: list[dict[str, Any]] = Field(default_factory=list)
    defenses_deployed: list[dict[str, Any]] = Field(default_factory=list)
    attack_history: list[dict[str, Any]] = Field(default_factory=list)
    vulnerabilities_found: list[dict[str, Any]] = Field(default_factory=list)
    
    # Metrics
    security_score: float = 100.0
    threats_blocked: int = 0
    incidents_active: int = 0


class POSAgentSchema(DomainMemorySchema):
    """Schema for POS/retail agents."""
    
    agent_type: str = "pos"
    domain: str = "retail"
    
    # POS-specific fields
    role: str = "cashier"  # cashier, inventory, manager
    active_transactions: list[dict[str, Any]] = Field(default_factory=list)
    inventory_alerts: list[dict[str, Any]] = Field(default_factory=list)
    customer_queue: list[dict[str, Any]] = Field(default_factory=list)
    daily_stats: dict[str, Any] = Field(default_factory=lambda: {
        "transactions_count": 0,
        "total_sales": 0.0,
        "avg_transaction_value": 0.0,
    })


# Schema registry for dynamic instantiation
SCHEMA_REGISTRY: dict[str, type[DomainMemorySchema]] = {
    "chat": ChatAgentSchema,
    "restaurant": RestaurantAgentSchema,
    "medical": MedicalAgentSchema,
    "security": SecurityAgentSchema,
    "pos": POSAgentSchema,
    "default": DomainMemorySchema,
}


def get_schema_class(agent_type: str) -> type[DomainMemorySchema]:
    """Get the appropriate schema class for an agent type."""
    return SCHEMA_REGISTRY.get(agent_type, DomainMemorySchema)


def create_schema(agent_id: str, agent_type: str, **kwargs) -> DomainMemorySchema:
    """Create a new schema instance for an agent."""
    schema_class = get_schema_class(agent_type)
    return schema_class(agent_id=agent_id, agent_type=agent_type, **kwargs)
