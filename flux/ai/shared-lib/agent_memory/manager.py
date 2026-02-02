"""
Domain Memory Manager - Unified interface for agent memory operations.

This is the main entry point for agents to interact with the memory system.
It combines:
- Domain Memory Factory (task execution)
- Multi-tiered memory (conversation, working, entity, user, long-term)
- Memory queries and retrieval
- Full observability: metrics, tracing, structured logging

Usage:
    manager = DomainMemoryManager(
        agent_id="agent-bruno",
        agent_type="chat",
        redis_url="redis://localhost:6379",
        postgres_url="postgresql://...",
    )
    
    await manager.connect()
    
    # Start a new conversation with memory
    memory = await manager.start_conversation(
        user_id="user-123",
        message="Hello, what can you do?",
    )
    
    # Access conversation context
    context = memory.get_conversation_context(limit=10)
    
    # Add a message
    await manager.add_message(memory, "user", "Tell me about the weather")
    
    # Get user preferences
    user_prefs = await manager.get_user_memory(user_id="user-123")
    
    # Save learnings to long-term memory
    await manager.record_learning(
        domain="weather",
        content="User prefers detailed forecasts",
        source="conversation",
    )
"""

import os
import time
from datetime import datetime, timezone
from typing import Any, Optional, Callable, Awaitable
from uuid import uuid4

from .types import (
    MemoryType,
    MemoryEntry,
    ConversationMemory,
    WorkingMemory,
    EntityMemory,
    UserMemory,
    DomainMemory,
)
from .schema import (
    DomainMemorySchema,
    create_schema,
    get_schema_class,
)
from .store import (
    MemoryStore,
    InMemoryStore,
    RedisMemoryStore,
    PostgresMemoryStore,
)
from .factory import DomainMemoryFactory
from .observability import (
    MemoryLogger,
    trace_memory_operation,
    trace_async,
    record_store_connected,
    record_conversation_message,
    record_conversation_length,
    record_user_fact,
    record_user_preference,
    record_entity_relationship,
    record_learning,
    record_error_pattern,
    record_task_created,
    record_task_completed,
    record_cache_hit,
    record_cache_miss,
    record_context_build,
    set_conversations_active,
    set_entities_count,
    set_patterns_count,
)


class DomainMemoryManager:
    """
    Unified interface for all agent memory operations.
    
    Provides:
    - Conversation memory management (short-term)
    - Working memory for task execution
    - User memory for personalization
    - Entity memory for domain objects
    - Long-term memory for accumulated knowledge
    - Domain memory factory for structured execution
    - Full observability: metrics, tracing, structured logging
    """
    
    def __init__(
        self,
        agent_id: str,
        agent_type: str,
        domain: str = None,
        redis_url: str = None,
        postgres_url: str = None,
        use_redis: bool = True,
        use_postgres: bool = True,
        llm_analyzer: Callable[[str], Awaitable[dict]] = None,
        default_constraints: list[dict] = None,
    ):
        self.agent_id = agent_id
        self.agent_type = agent_type
        self.domain = domain or agent_type
        
        # Initialize structured logger with trace context
        self.log = MemoryLogger(agent_id=agent_id, component="memory_manager")
        
        # Configure stores
        self.redis_url = redis_url or os.getenv("REDIS_URL", "redis://localhost:6379/0")
        self.postgres_url = postgres_url or os.getenv("POSTGRES_URL")
        
        # Short-term store (Redis or InMemory)
        self._short_term_type = "redis" if use_redis else "in_memory"
        if use_redis:
            self.short_term_store = RedisMemoryStore(url=self.redis_url)
        else:
            self.short_term_store = InMemoryStore()
        
        # Long-term store (PostgreSQL or InMemory)
        self._long_term_type = "postgres" if (use_postgres and self.postgres_url) else "in_memory"
        if use_postgres and self.postgres_url:
            self.long_term_store = PostgresMemoryStore(url=self.postgres_url)
        else:
            self.long_term_store = InMemoryStore()
        
        # Domain memory factory
        self.factory = DomainMemoryFactory(
            agent_id=agent_id,
            agent_type=agent_type,
            domain=self.domain,
            store=self.short_term_store,  # Use Redis for schemas (with TTL)
            llm_analyzer=llm_analyzer,
            default_constraints=default_constraints,
        )
        
        self._connected = False
        self._active_conversations = 0
    
    async def connect(self):
        """Connect to all memory stores."""
        if self._connected:
            return
        
        start_time = time.perf_counter()
        
        try:
            await self.short_term_store.connect()
            record_store_connected(self.agent_id, self._short_term_type, True)
        except Exception as e:
            record_store_connected(self.agent_id, self._short_term_type, False)
            self.log.error("memory_store_connect_failed", 
                          store_type=self._short_term_type, error=str(e))
            raise
        
        try:
            await self.long_term_store.connect()
            record_store_connected(self.agent_id, self._long_term_type, True)
        except Exception as e:
            record_store_connected(self.agent_id, self._long_term_type, False)
            self.log.error("memory_store_connect_failed",
                          store_type=self._long_term_type, error=str(e))
            raise
        
        self._connected = True
        duration_ms = (time.perf_counter() - start_time) * 1000
        
        self.log.info(
            "memory_manager_connected",
            short_term=type(self.short_term_store).__name__,
            long_term=type(self.long_term_store).__name__,
            duration_ms=round(duration_ms, 2),
        )
    
    async def disconnect(self):
        """Disconnect from all memory stores."""
        if not self._connected:
            return
        
        await self.short_term_store.disconnect()
        record_store_connected(self.agent_id, self._short_term_type, False)
        
        await self.long_term_store.disconnect()
        record_store_connected(self.agent_id, self._long_term_type, False)
        
        self._connected = False
        self.log.info("memory_manager_disconnected")
    
    # =========================================================================
    # Conversation Memory (Short-term)
    # =========================================================================
    
    @trace_async("memory.conversation.start", record_args=["user_id", "conversation_id"])
    async def start_conversation(
        self,
        user_id: str = None,
        conversation_id: str = None,
        initial_message: str = None,
        context: dict = None,
    ) -> ConversationMemory:
        """
        Start a new conversation or resume an existing one.
        
        Returns ConversationMemory for tracking the interaction.
        """
        start_time = time.perf_counter()
        conv_id = conversation_id or str(uuid4())
        
        # Check for existing conversation
        if conversation_id:
            existing = await self.get_conversation(conversation_id)
            if existing:
                record_cache_hit(self.agent_id, "conversation")
                self.log.conversation_event(
                    "resumed",
                    conversation_id=conv_id,
                    user_id=user_id,
                    message_count=existing.message_count,
                )
                return existing
            else:
                record_cache_miss(self.agent_id, "conversation")
        
        # Create new conversation
        conv = ConversationMemory(
            agent_id=self.agent_id,
            conversation_id=conv_id,
            user_id=user_id,
            context=context or {},
        )
        
        if initial_message:
            conv.add_message("user", initial_message)
            record_conversation_message(self.agent_id, "user")
        
        async with trace_memory_operation("save", self.agent_id, self._short_term_type):
            await self.short_term_store.save(conv)
        
        self._active_conversations += 1
        set_conversations_active(self.agent_id, self._active_conversations)
        
        duration_ms = (time.perf_counter() - start_time) * 1000
        self.log.conversation_event(
            "started",
            conversation_id=conv_id,
            user_id=user_id,
            message_count=1 if initial_message else 0,
            duration_ms=round(duration_ms, 2),
        )
        
        return conv
    
    @trace_async("memory.conversation.get", record_args=["conversation_id"])
    async def get_conversation(self, conversation_id: str) -> Optional[ConversationMemory]:
        """Get an existing conversation by ID."""
        async with trace_memory_operation("query", self.agent_id, self._short_term_type):
            entries = await self.short_term_store.query(
                memory_type=MemoryType.SHORT_TERM,
                agent_id=self.agent_id,
                filters={"conversation_id": conversation_id},
                limit=1,
            )
        
        if entries:
            record_cache_hit(self.agent_id, "conversation")
            return entries[0]
        
        record_cache_miss(self.agent_id, "conversation")
        return None
    
    @trace_async("memory.conversation.add_message", record_args=["role"])
    async def add_message(
        self,
        conversation: ConversationMemory,
        role: str,
        content: str,
        metadata: dict = None,
    ) -> ConversationMemory:
        """Add a message to a conversation."""
        conversation.add_message(role, content, metadata)
        record_conversation_message(self.agent_id, role)
        record_conversation_length(self.agent_id, conversation.message_count)
        
        async with trace_memory_operation("save", self.agent_id, self._short_term_type):
            await self.short_term_store.save(conversation)
        
        self.log.conversation_event(
            "message_added",
            conversation_id=conversation.conversation_id,
            user_id=conversation.user_id,
            message_count=conversation.message_count,
            role=role,
            content_length=len(content),
        )
        
        return conversation
    
    async def summarize_conversation(
        self,
        conversation: ConversationMemory,
        summarizer: Callable[[list[dict]], Awaitable[str]] = None,
    ) -> str:
        """
        Generate a summary of the conversation.
        
        Useful for long conversations to maintain context without
        exceeding token limits.
        """
        if summarizer:
            summary = await summarizer(conversation.messages)
        else:
            # Simple extractive summary
            messages = conversation.messages
            if len(messages) <= 5:
                summary = " | ".join([m["content"][:100] for m in messages])
            else:
                # Take first 2, last 2, and middle
                key_messages = messages[:2] + [messages[len(messages)//2]] + messages[-2:]
                summary = " | ".join([m["content"][:50] for m in key_messages])
        
        conversation.summary = summary
        await self.short_term_store.save(conversation)
        
        return summary
    
    # =========================================================================
    # Working Memory (Task Execution)
    # =========================================================================
    
    async def create_task(
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
        Create a new task with domain memory.
        
        This uses the Domain Memory Factory pattern.
        """
        return await self.factory.initialize(
            request=request,
            user_id=user_id,
            session_id=session_id,
            context=context,
            goals=goals,
            requirements=requirements,
            constraints=constraints,
        )
    
    async def get_task(
        self,
        session_id: str = None,
        schema_id: str = None,
    ) -> Optional[DomainMemorySchema]:
        """Get an existing task/schema."""
        if schema_id:
            return await self.short_term_store.get_schema(schema_id)
        if session_id:
            return await self.short_term_store.get_schema_by_agent(
                agent_id=self.agent_id,
                session_id=session_id,
            )
        return None
    
    async def update_task(self, schema: DomainMemorySchema):
        """Update a task's domain memory."""
        await self.factory.update(schema)
    
    async def complete_task(
        self,
        schema: DomainMemorySchema,
        summary: str,
        success: bool = True,
        learnings: list[str] = None,
    ):
        """Complete a task and record learnings."""
        await self.factory.complete(schema, summary, success, learnings)
        
        # Also record learnings to long-term memory
        if learnings:
            for learning in learnings:
                await self.record_learning(
                    domain=self.domain,
                    content=learning,
                    source=f"task:{schema.task_id}",
                )
    
    # =========================================================================
    # User Memory (Personalization)
    # =========================================================================
    
    @trace_async("memory.user.get", record_args=["user_id"])
    async def get_user_memory(self, user_id: str) -> Optional[UserMemory]:
        """Get memory for a specific user."""
        async with trace_memory_operation("query", self.agent_id, self._long_term_type):
            entries = await self.long_term_store.query(
                memory_type=MemoryType.USER,
                agent_id=self.agent_id,
                filters={"user_id": user_id},
                limit=1,
            )
        
        if entries:
            record_cache_hit(self.agent_id, "user")
            return entries[0]
        
        record_cache_miss(self.agent_id, "user")
        return None
    
    @trace_async("memory.user.get_or_create", record_args=["user_id"])
    async def get_or_create_user_memory(self, user_id: str) -> UserMemory:
        """Get existing user memory or create new one."""
        existing = await self.get_user_memory(user_id)
        if existing:
            return existing
        
        user_mem = UserMemory(
            agent_id=self.agent_id,
            user_id=user_id,
        )
        
        async with trace_memory_operation("save", self.agent_id, self._long_term_type):
            await self.long_term_store.save(user_mem)
        
        self.log.user_memory_event("created", user_id=user_id)
        return user_mem
    
    @trace_async("memory.user.update_preference", record_args=["user_id", "key"])
    async def update_user_preference(
        self,
        user_id: str,
        key: str,
        value: Any,
        explicit: bool = True,
    ):
        """Update a user preference."""
        user_mem = await self.get_or_create_user_memory(user_id)
        user_mem.set_preference(key, value, explicit)
        
        async with trace_memory_operation("save", self.agent_id, self._long_term_type):
            await self.long_term_store.save(user_mem)
        
        record_user_preference(self.agent_id, explicit)
        self.log.user_memory_event(
            "preference_updated",
            user_id=user_id,
            key=key,
            explicit=explicit,
        )
    
    @trace_async("memory.user.add_fact", record_args=["user_id", "source"])
    async def add_user_fact(
        self,
        user_id: str,
        fact: str,
        source: str = "conversation",
        confidence: float = 0.8,
    ):
        """Add a fact about a user."""
        user_mem = await self.get_or_create_user_memory(user_id)
        user_mem.add_fact(fact, source, confidence)
        
        async with trace_memory_operation("save", self.agent_id, self._long_term_type):
            await self.long_term_store.save(user_mem)
        
        record_user_fact(self.agent_id, source)
        self.log.user_memory_event(
            "fact_added",
            user_id=user_id,
            source=source,
            confidence=confidence,
            fact_preview=fact[:50],
        )
    
    @trace_async("memory.user.record_interaction", record_args=["user_id", "interaction_type"])
    async def record_user_interaction(self, user_id: str, interaction_type: str):
        """Record a user interaction for stats."""
        user_mem = await self.get_or_create_user_memory(user_id)
        user_mem.record_interaction(interaction_type)
        
        async with trace_memory_operation("save", self.agent_id, self._long_term_type):
            await self.long_term_store.save(user_mem)
    
    # =========================================================================
    # Entity Memory (Domain Objects)
    # =========================================================================
    
    @trace_async("memory.entity.get", record_args=["entity_type", "entity_id"])
    async def get_entity(
        self,
        entity_type: str,
        entity_id: str,
    ) -> Optional[EntityMemory]:
        """Get memory for a specific entity."""
        async with trace_memory_operation("query", self.agent_id, self._long_term_type):
            entries = await self.long_term_store.query(
                memory_type=MemoryType.ENTITY,
                agent_id=self.agent_id,
                filters={"entity_type": entity_type, "entity_id": entity_id},
                limit=1,
            )
        
        if entries:
            record_cache_hit(self.agent_id, "entity")
            return entries[0]
        
        record_cache_miss(self.agent_id, "entity")
        return None
    
    @trace_async("memory.entity.create_or_update", record_args=["entity_type", "entity_id"])
    async def create_or_update_entity(
        self,
        entity_type: str,
        entity_id: str,
        entity_name: str = None,
        attributes: dict = None,
        tags: list[str] = None,
    ) -> EntityMemory:
        """Create or update an entity in memory."""
        existing = await self.get_entity(entity_type, entity_id)
        
        if existing:
            if attributes:
                for key, value in attributes.items():
                    existing.update_attribute(key, value)
            if tags:
                existing.tags = list(set(existing.tags + tags))
            if entity_name:
                existing.entity_name = entity_name
            
            async with trace_memory_operation("save", self.agent_id, self._long_term_type):
                await self.long_term_store.save(existing)
            
            self.log.entity_event(
                "updated",
                entity_type=entity_type,
                entity_id=entity_id,
                attribute_count=len(existing.attributes),
            )
            return existing
        
        entity = EntityMemory(
            agent_id=self.agent_id,
            entity_type=entity_type,
            entity_id=entity_id,
            entity_name=entity_name,
            attributes={k: {"value": v, "confidence": 1.0} for k, v in (attributes or {}).items()},
            tags=tags or [],
        )
        
        async with trace_memory_operation("save", self.agent_id, self._long_term_type):
            await self.long_term_store.save(entity)
        
        self.log.entity_event(
            "created",
            entity_type=entity_type,
            entity_id=entity_id,
            entity_name=entity_name,
        )
        
        return entity
    
    @trace_async("memory.entity.add_relationship", record_args=["entity_type", "relation_type"])
    async def add_entity_relationship(
        self,
        entity_type: str,
        entity_id: str,
        relation_type: str,
        target_entity_id: str,
        metadata: dict = None,
    ):
        """Add a relationship between entities."""
        entity = await self.get_entity(entity_type, entity_id)
        if entity:
            entity.add_relationship(relation_type, target_entity_id, metadata)
            
            async with trace_memory_operation("save", self.agent_id, self._long_term_type):
                await self.long_term_store.save(entity)
            
            record_entity_relationship(self.agent_id, relation_type)
            self.log.entity_event(
                "relationship_added",
                entity_type=entity_type,
                entity_id=entity_id,
                relation_type=relation_type,
                target_entity_id=target_entity_id,
            )
    
    # =========================================================================
    # Long-term Memory (Knowledge)
    # =========================================================================
    
    @trace_async("memory.domain.get")
    async def get_domain_memory(self) -> Optional[DomainMemory]:
        """Get the agent's long-term domain memory."""
        async with trace_memory_operation("query", self.agent_id, self._long_term_type):
            entries = await self.long_term_store.query(
                memory_type=MemoryType.LONG_TERM,
                agent_id=self.agent_id,
                filters={"domain": self.domain},
                limit=1,
            )
        
        if entries:
            record_cache_hit(self.agent_id, "domain")
            return entries[0]
        
        record_cache_miss(self.agent_id, "domain")
        return None
    
    @trace_async("memory.domain.get_or_create")
    async def get_or_create_domain_memory(self) -> DomainMemory:
        """Get existing domain memory or create new one."""
        existing = await self.get_domain_memory()
        if existing:
            return existing
        
        domain_mem = DomainMemory(
            agent_id=self.agent_id,
            domain=self.domain,
        )
        
        async with trace_memory_operation("save", self.agent_id, self._long_term_type):
            await self.long_term_store.save(domain_mem)
        
        self.log.info("domain_memory_created", domain=self.domain)
        return domain_mem
    
    @trace_async("memory.learning.record", record_args=["category", "source"])
    async def record_learning(
        self,
        domain: str,
        content: str,
        source: str,
        category: str = "general",
        confidence: float = 1.0,
    ):
        """Record a learning to long-term memory."""
        domain_mem = await self.get_or_create_domain_memory()
        domain_mem.add_knowledge(content, category, source, confidence)
        
        async with trace_memory_operation("save", self.agent_id, self._long_term_type):
            await self.long_term_store.save(domain_mem)
        
        record_learning(self.agent_id, category)
        self.log.learning_event(
            category=category,
            content_preview=content,
            source=source,
        )
    
    @trace_async("memory.pattern.record", record_args=["name"])
    async def record_pattern(
        self,
        name: str,
        description: str,
        when_to_use: str,
        example: str = None,
    ):
        """Record a discovered pattern or best practice."""
        domain_mem = await self.get_or_create_domain_memory()
        domain_mem.add_pattern(name, description, when_to_use, example)
        
        async with trace_memory_operation("save", self.agent_id, self._long_term_type):
            await self.long_term_store.save(domain_mem)
        
        set_patterns_count(self.agent_id, len(domain_mem.patterns))
        self.log.info(
            "memory_pattern_recorded",
            pattern_name=name,
            total_patterns=len(domain_mem.patterns),
        )
    
    @trace_async("memory.error_pattern.record", record_args=["error_type", "severity"])
    async def record_error_pattern(
        self,
        error_type: str,
        description: str,
        prevention: str,
        severity: str = "medium",
    ):
        """Record an error pattern for future avoidance."""
        domain_mem = await self.get_or_create_domain_memory()
        domain_mem.record_error(error_type, description, prevention, severity)
        
        async with trace_memory_operation("save", self.agent_id, self._long_term_type):
            await self.long_term_store.save(domain_mem)
        
        record_error_pattern(self.agent_id, severity)
        self.log.info(
            "memory_error_pattern_recorded",
            error_type=error_type,
            severity=severity,
        )
    
    @trace_async("memory.task_completion.record", record_args=["task_id", "success"])
    async def record_task_completion(
        self,
        task_id: str,
        summary: str,
        success: bool,
        duration_ms: float,
        learnings: list[str] = None,
    ):
        """Record a completed task to long-term memory."""
        domain_mem = await self.get_or_create_domain_memory()
        domain_mem.record_task_completion(
            task_id, summary, success, duration_ms, learnings
        )
        
        async with trace_memory_operation("save", self.agent_id, self._long_term_type):
            await self.long_term_store.save(domain_mem)
        
        record_task_completed(self.agent_id, success, duration_ms / 1000.0)
        self.log.task_event(
            "completed",
            task_id=task_id,
            success=success,
            duration_ms=duration_ms,
            learnings_count=len(learnings) if learnings else 0,
        )
    
    # =========================================================================
    # Context Building
    # =========================================================================
    
    @trace_async("memory.context.build", record_args=["user_id", "conversation_id"])
    async def build_context(
        self,
        user_id: str = None,
        conversation_id: str = None,
        session_id: str = None,
        include_user_memory: bool = True,
        include_domain_knowledge: bool = True,
        include_recent_tasks: bool = False,
        conversation_limit: int = 10,
    ) -> dict:
        """
        Build a comprehensive context object for LLM prompts.
        
        This aggregates relevant information from all memory tiers.
        """
        start_time = time.perf_counter()
        
        context = {
            "agent_id": self.agent_id,
            "agent_type": self.agent_type,
            "domain": self.domain,
        }
        
        # Conversation context
        if conversation_id:
            conv = await self.get_conversation(conversation_id)
            if conv:
                context["conversation"] = {
                    "id": conv.conversation_id,
                    "messages": conv.get_recent_messages(conversation_limit),
                    "summary": conv.summary,
                    "message_count": conv.message_count,
                }
        
        # Task/schema context
        if session_id:
            schema = await self.get_task(session_id=session_id)
            if schema:
                context["task"] = schema.to_context_prompt()
        
        # User context
        if user_id and include_user_memory:
            user_mem = await self.get_user_memory(user_id)
            if user_mem:
                context["user"] = {
                    "id": user_id,
                    "preferences": user_mem.preferences,
                    "facts": [f["fact"] for f in user_mem.facts[-5:]],
                    "custom_instructions": user_mem.custom_instructions,
                }
        
        # Domain knowledge
        if include_domain_knowledge:
            domain_mem = await self.get_domain_memory()
            if domain_mem:
                context["domain_knowledge"] = {
                    "patterns": [p["name"] for p in domain_mem.patterns[-5:]],
                    "performance": domain_mem.performance,
                }
        
        # Calculate metrics
        duration_ms = (time.perf_counter() - start_time) * 1000
        context_str = self.format_context_for_prompt(context)
        context_size = len(context_str)
        
        record_context_build(self.agent_id, duration_ms / 1000.0, context_size)
        self.log.context_built(
            user_id=user_id,
            conversation_id=conversation_id,
            context_size_chars=context_size,
            duration_ms=round(duration_ms, 2),
            has_user_context=bool(context.get("user")),
            has_conversation=bool(context.get("conversation")),
            has_domain_knowledge=bool(context.get("domain_knowledge")),
        )
        
        return context
    
    def format_context_for_prompt(self, context: dict) -> str:
        """Format context dictionary as a string for LLM prompt."""
        lines = ["## Context"]
        
        if "user" in context:
            user = context["user"]
            lines.append(f"\n### User: {user['id']}")
            if user.get("preferences"):
                lines.append(f"Preferences: {user['preferences']}")
            if user.get("facts"):
                lines.append(f"Known facts: {', '.join(user['facts'])}")
            if user.get("custom_instructions"):
                lines.append(f"Instructions: {user['custom_instructions']}")
        
        if "task" in context:
            lines.append(f"\n{context['task']}")
        
        if "conversation" in context:
            conv = context["conversation"]
            if conv.get("summary"):
                lines.append(f"\n### Conversation Summary\n{conv['summary']}")
        
        if "domain_knowledge" in context:
            dk = context["domain_knowledge"]
            if dk.get("patterns"):
                lines.append(f"\n### Available Patterns: {', '.join(dk['patterns'])}")
        
        return "\n".join(lines)
