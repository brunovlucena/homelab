"""
Domain Memory Factory - Two-Agent Pattern for Disciplined Execution.

Following Nate B. Jones's Domain Memory Factory pattern:

1. INITIALIZER AGENT: Sets up the structured memory
   - Analyzes the task/request
   - Defines explicit goals
   - Identifies requirements and constraints
   - Establishes success criteria
   - Creates the domain memory schema

2. WORKER AGENT: Acts upon the structured memory
   - Works towards goals in discrete steps
   - Updates progress and state
   - Records decisions with reasoning
   - Produces artifacts
   - Maintains accountability through the schema

This transforms agents from "forgetful entities into disciplined workers."
"""

import os
from datetime import datetime, timezone
from typing import Any, Optional, Callable, Awaitable
from uuid import uuid4

import structlog

from .types import (
    MemoryType,
    WorkingMemory,
    ConversationMemory,
    UserMemory,
    EntityMemory,
    DomainMemory,
)
from .schema import (
    DomainMemorySchema,
    AgentGoal,
    AgentRequirement,
    AgentConstraint,
    TaskStatus,
    GoalPriority,
    create_schema,
    get_schema_class,
)
from .store import MemoryStore, InMemoryStore

logger = structlog.get_logger()


class MemoryInitializer:
    """
    Initializer component of the Domain Memory Factory.
    
    Responsible for:
    - Analyzing incoming requests
    - Setting up structured memory with goals, requirements, constraints
    - Creating the domain memory schema
    
    This can be implemented as:
    - LLM-based (uses AI to analyze and structure)
    - Rule-based (uses predefined patterns)
    - Hybrid (combines both)
    """
    
    def __init__(
        self,
        agent_id: str,
        agent_type: str,
        domain: str,
        store: MemoryStore = None,
        llm_analyzer: Callable[[str], Awaitable[dict]] = None,
    ):
        self.agent_id = agent_id
        self.agent_type = agent_type
        self.domain = domain
        self.store = store or InMemoryStore()
        self.llm_analyzer = llm_analyzer
    
    async def initialize_memory(
        self,
        request: str,
        user_id: str = None,
        session_id: str = None,
        context: dict = None,
        predefined_goals: list[dict] = None,
        predefined_requirements: list[dict] = None,
        predefined_constraints: list[dict] = None,
    ) -> DomainMemorySchema:
        """
        Initialize domain memory for a new task/request.
        
        This is the core function that transforms an unstructured request
        into a structured, accountable execution plan.
        
        Args:
            request: The incoming request/task description
            user_id: Optional user identifier
            session_id: Optional session identifier
            context: Additional context data
            predefined_goals: Pre-defined goals to use (skips LLM analysis)
            predefined_requirements: Pre-defined requirements
            predefined_constraints: Pre-defined constraints
        
        Returns:
            Initialized DomainMemorySchema ready for the worker agent
        """
        log = logger.bind(
            agent_id=self.agent_id,
            agent_type=self.agent_type,
            session_id=session_id,
        )
        
        log.info("initializing_domain_memory", request_length=len(request))
        
        # Create the schema
        schema_class = get_schema_class(self.agent_type)
        schema = schema_class(
            agent_id=self.agent_id,
            agent_type=self.agent_type,
            domain=self.domain,
            session_id=session_id or str(uuid4()),
            task_id=str(uuid4()),
            user_id=user_id,
        )
        
        # If predefined goals/requirements/constraints provided, use them
        if predefined_goals:
            for goal_data in predefined_goals:
                schema.add_goal(
                    description=goal_data["description"],
                    priority=GoalPriority(goal_data.get("priority", 3)),
                )
        
        if predefined_requirements:
            for req_data in predefined_requirements:
                schema.add_requirement(
                    description=req_data["description"],
                    mandatory=req_data.get("mandatory", True),
                )
        
        if predefined_constraints:
            for con_data in predefined_constraints:
                schema.add_constraint(
                    description=con_data["description"],
                    hard=con_data.get("hard", True),
                    category=con_data.get("category", "general"),
                )
        
        # If no predefined structure, use LLM or rule-based analysis
        if not predefined_goals:
            if self.llm_analyzer:
                analysis = await self._llm_analyze(request, context)
            else:
                analysis = self._rule_based_analyze(request, context)
            
            # Apply analysis results to schema
            for goal in analysis.get("goals", []):
                schema.add_goal(
                    description=goal["description"],
                    priority=GoalPriority(goal.get("priority", 3)),
                )
            
            for req in analysis.get("requirements", []):
                schema.add_requirement(
                    description=req["description"],
                    mandatory=req.get("mandatory", True),
                )
            
            for con in analysis.get("constraints", []):
                schema.add_constraint(
                    description=con["description"],
                    hard=con.get("hard", True),
                    category=con.get("category", "general"),
                )
            
            # Set progress steps if identified
            if "steps" in analysis:
                schema.progress.steps_total = len(analysis["steps"])
                schema.state.context["planned_steps"] = analysis["steps"]
        
        # Initialize state
        schema.state.current_step = "initialized"
        schema.state.context["original_request"] = request
        schema.state.context["initialized_at"] = datetime.now(timezone.utc).isoformat()
        if context:
            schema.state.context.update(context)
        
        # Persist the schema
        await self.store.save_schema(schema)
        
        log.info(
            "domain_memory_initialized",
            schema_id=schema.schema_id,
            goals=len(schema.goals),
            requirements=len(schema.requirements),
            constraints=len(schema.constraints),
        )
        
        return schema
    
    async def _llm_analyze(self, request: str, context: dict = None) -> dict:
        """Use LLM to analyze request and extract structure."""
        if not self.llm_analyzer:
            return self._rule_based_analyze(request, context)
        
        try:
            analysis = await self.llm_analyzer(request)
            return analysis
        except Exception as e:
            logger.warning("llm_analysis_failed", error=str(e))
            return self._rule_based_analyze(request, context)
    
    def _rule_based_analyze(self, request: str, context: dict = None) -> dict:
        """
        Rule-based analysis for common patterns.
        
        This provides a fallback when LLM is not available and
        handles domain-specific patterns efficiently.
        """
        request_lower = request.lower()
        
        goals = []
        requirements = []
        constraints = []
        steps = []
        
        # Domain-specific patterns
        if self.agent_type == "chat":
            goals.append({
                "description": "Provide helpful, accurate response to user query",
                "priority": 2,
            })
            constraints.append({
                "description": "Maintain conversation context and coherence",
                "hard": True,
                "category": "quality",
            })
        
        elif self.agent_type == "restaurant":
            if "order" in request_lower:
                goals.append({
                    "description": "Process and confirm the order accurately",
                    "priority": 1,
                })
                requirements.append({
                    "description": "Verify all order items are available",
                    "mandatory": True,
                })
                steps = ["confirm_items", "check_availability", "process_order", "confirm"]
            
            elif "recommend" in request_lower or "suggest" in request_lower:
                goals.append({
                    "description": "Provide personalized recommendation",
                    "priority": 2,
                })
            
            constraints.append({
                "description": "Maintain professional hospitality demeanor",
                "hard": True,
                "category": "behavior",
            })
        
        elif self.agent_type == "medical":
            goals.append({
                "description": "Provide accurate medical information while maintaining privacy",
                "priority": 1,
            })
            constraints.append({
                "description": "HIPAA compliance - protect patient data",
                "hard": True,
                "category": "privacy",
            })
            constraints.append({
                "description": "Verify authorization before accessing records",
                "hard": True,
                "category": "security",
            })
            requirements.append({
                "description": "Log all data access for audit trail",
                "mandatory": True,
            })
        
        elif self.agent_type == "security":
            if "attack" in request_lower or "exploit" in request_lower:
                goals.append({
                    "description": "Execute security test safely and document findings",
                    "priority": 1,
                })
            else:
                goals.append({
                    "description": "Monitor and respond to security events",
                    "priority": 1,
                })
            
            constraints.append({
                "description": "Stay within authorized scope",
                "hard": True,
                "category": "authorization",
            })
        
        elif self.agent_type == "pos":
            if "transaction" in request_lower or "sale" in request_lower:
                goals.append({
                    "description": "Complete transaction accurately",
                    "priority": 1,
                })
                steps = ["scan_items", "calculate_total", "process_payment", "receipt"]
            
            constraints.append({
                "description": "Ensure accurate pricing and inventory tracking",
                "hard": True,
                "category": "accuracy",
            })
        
        # Default goal if none matched
        if not goals:
            goals.append({
                "description": f"Process request: {request[:100]}...",
                "priority": 3,
            })
        
        return {
            "goals": goals,
            "requirements": requirements,
            "constraints": constraints,
            "steps": steps,
        }


class DomainMemoryFactory:
    """
    Domain Memory Factory - Orchestrates the two-agent pattern.
    
    Usage:
        factory = DomainMemoryFactory(
            agent_id="agent-restaurant-waiter",
            agent_type="restaurant",
            domain="hospitality",
            store=redis_store,
        )
        
        # Initialize memory for a new request
        schema = await factory.initialize(
            request="I'd like to order the salmon",
            user_id="customer-123",
        )
        
        # Worker agent processes with the schema
        schema.state.transition("processing_order")
        schema.record_decision(
            decision="Recommend wine pairing",
            reasoning="Customer ordered fish, suggest white wine",
        )
        
        # Update and persist progress
        await factory.update(schema)
        
        # Complete the task
        await factory.complete(schema, summary="Order processed successfully")
    """
    
    def __init__(
        self,
        agent_id: str,
        agent_type: str,
        domain: str,
        store: MemoryStore = None,
        llm_analyzer: Callable[[str], Awaitable[dict]] = None,
        default_constraints: list[dict] = None,
    ):
        self.agent_id = agent_id
        self.agent_type = agent_type
        self.domain = domain
        self.store = store or InMemoryStore()
        self.default_constraints = default_constraints or []
        
        self.initializer = MemoryInitializer(
            agent_id=agent_id,
            agent_type=agent_type,
            domain=domain,
            store=self.store,
            llm_analyzer=llm_analyzer,
        )
    
    async def connect(self):
        """Connect to the memory store."""
        await self.store.connect()
    
    async def disconnect(self):
        """Disconnect from the memory store."""
        await self.store.disconnect()
    
    async def initialize(
        self,
        request: str,
        user_id: str = None,
        session_id: str = None,
        context: dict = None,
        goals: list[dict] = None,
        requirements: list[dict] = None,
        constraints: list[dict] = None,
    ) -> DomainMemorySchema:
        """
        Initialize domain memory for a new task.
        
        This is the entry point for the Initializer Agent role.
        """
        # Merge default constraints with provided ones
        all_constraints = self.default_constraints.copy()
        if constraints:
            all_constraints.extend(constraints)
        
        schema = await self.initializer.initialize_memory(
            request=request,
            user_id=user_id,
            session_id=session_id,
            context=context,
            predefined_goals=goals,
            predefined_requirements=requirements,
            predefined_constraints=all_constraints if all_constraints else None,
        )
        
        return schema
    
    async def get_or_create(
        self,
        session_id: str,
        request: str = None,
        user_id: str = None,
        context: dict = None,
    ) -> DomainMemorySchema:
        """
        Get existing schema for session or create new one.
        
        This supports resuming work on an existing task.
        """
        # Try to get existing schema
        existing = await self.store.get_schema_by_agent(
            agent_id=self.agent_id,
            session_id=session_id,
        )
        
        if existing:
            logger.info(
                "resuming_domain_memory",
                schema_id=existing.schema_id,
                session_id=session_id,
            )
            return existing
        
        # Create new if not found
        if not request:
            request = "New session"
        
        return await self.initialize(
            request=request,
            user_id=user_id,
            session_id=session_id,
            context=context,
        )
    
    async def update(self, schema: DomainMemorySchema):
        """
        Persist updates to the domain memory schema.
        
        Call this after making changes to the schema during task execution.
        """
        schema.updated_at = datetime.now(timezone.utc)
        await self.store.save_schema(schema)
    
    async def complete(
        self,
        schema: DomainMemorySchema,
        summary: str,
        success: bool = True,
        learnings: list[str] = None,
    ):
        """
        Mark a task as complete and record learnings.
        
        This captures the outcome for long-term memory.
        """
        # Update all active goals
        for goal in schema.goals:
            if goal.status in (TaskStatus.PENDING, TaskStatus.IN_PROGRESS):
                if success:
                    goal.complete()
                else:
                    goal.fail()
        
        # Update state
        schema.state.transition("completed" if success else "failed")
        schema.progress.steps_completed = schema.progress.steps_total
        
        # Record completion
        schema.add_artifact(
            name="completion_summary",
            artifact_type="summary",
            content=summary,
            metadata={
                "success": success,
                "learnings": learnings or [],
                "completed_at": datetime.now(timezone.utc).isoformat(),
            },
        )
        
        # Persist final state
        await self.update(schema)
        
        # TODO: Also update long-term memory with learnings
        
        logger.info(
            "task_completed",
            schema_id=schema.schema_id,
            success=success,
            summary=summary[:100],
        )
    
    async def fail(
        self,
        schema: DomainMemorySchema,
        error: str,
        recoverable: bool = False,
    ):
        """
        Mark a task as failed and record the error.
        """
        schema.state.last_error = error
        schema.state.transition("failed")
        
        for goal in schema.goals:
            if goal.status in (TaskStatus.PENDING, TaskStatus.IN_PROGRESS):
                goal.fail(error)
        
        schema.add_artifact(
            name="failure_record",
            artifact_type="error",
            content=error,
            metadata={
                "recoverable": recoverable,
                "failed_at": datetime.now(timezone.utc).isoformat(),
            },
        )
        
        await self.update(schema)
        
        logger.error(
            "task_failed",
            schema_id=schema.schema_id,
            error=error,
            recoverable=recoverable,
        )
