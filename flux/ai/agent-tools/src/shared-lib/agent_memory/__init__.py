"""
🧠 Agent Domain Memory Library

Implements Nate B. Jones's Domain Memory Factory pattern for stateful AI agents.

This library provides:
- Multi-tiered memory system (short-term, long-term, entity, user, working)
- Domain memory schemas for structured state representation
- Persistence backends (Redis, PostgreSQL)
- Memory federation for multi-agent systems

Key Concepts:
- Domain Memory: Structured, persistent representation of an agent's work
- Memory Factory: Two-agent pattern (Initializer + Worker) for disciplined execution
- Stateful Intelligence: Agents that remember and learn across sessions

References:
- https://www.texttube.ai/read/xNcEgqzlPqs/agent-memory-the-key-to-effective-ai-agents
"""

from .types import (
    MemoryType,
    MemoryEntry,
    DomainMemory,
    WorkingMemory,
    EntityMemory,
    UserMemory,
    ConversationMemory,
)
from .schema import (
    DomainMemorySchema,
    AgentGoal,
    AgentRequirement,
    AgentConstraint,
    AgentState,
    TaskProgress,
    # Domain-specific schemas
    ChatAgentSchema,
    RestaurantAgentSchema,
    MedicalAgentSchema,
    SecurityAgentSchema,
    POSAgentSchema,
)
from .store import (
    MemoryStore,
    RedisMemoryStore,
    PostgresMemoryStore,
    InMemoryStore,
)
from .factory import (
    DomainMemoryFactory,
    MemoryInitializer,
)
from .manager import (
    DomainMemoryManager,
)
from .observability import (
    MemoryLogger,
    trace_memory_operation,
    trace_async,
    get_tracer,
    get_current_trace_context,
    # Metric helpers
    record_store_connected,
    record_conversation_message,
    record_user_fact,
    record_learning,
    record_task_created,
    record_task_completed,
    record_context_build,
    set_conversations_active,
    init_memory_build_info,
)

__all__ = [
    # Types
    "MemoryType",
    "MemoryEntry",
    "DomainMemory",
    "WorkingMemory",
    "EntityMemory",
    "UserMemory",
    "ConversationMemory",
    # Schema - Base
    "DomainMemorySchema",
    "AgentGoal",
    "AgentRequirement",
    "AgentConstraint",
    "AgentState",
    "TaskProgress",
    # Schema - Domain-specific
    "ChatAgentSchema",
    "RestaurantAgentSchema",
    "MedicalAgentSchema",
    "SecurityAgentSchema",
    "POSAgentSchema",
    # Store
    "MemoryStore",
    "RedisMemoryStore",
    "PostgresMemoryStore",
    "InMemoryStore",
    # Factory
    "DomainMemoryFactory",
    "MemoryInitializer",
    # Manager
    "DomainMemoryManager",
    # Observability
    "MemoryLogger",
    "trace_memory_operation",
    "trace_async",
    "get_tracer",
    "get_current_trace_context",
    "record_store_connected",
    "record_conversation_message",
    "record_user_fact",
    "record_learning",
    "record_task_created",
    "record_task_completed",
    "record_context_build",
    "set_conversations_active",
    "init_memory_build_info",
]

__version__ = "1.0.0"
