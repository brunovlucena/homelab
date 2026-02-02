"""
Memory Store Backends - Persistence layer for domain memory.

Supports:
- Redis: Short-term and working memory (fast, TTL support)
- PostgreSQL: Long-term, entity, and user memory (durable, queryable)
- InMemory: Development and testing

Following the pattern: "The real competitive advantage lies not in the AI
models themselves but in the surrounding framework that enables durable progress."

Observability Features:
- Prometheus metrics for all store operations
- Structured logging with trace context
- Operation timing and error tracking
"""

import json
import os
import time
from abc import ABC, abstractmethod
from datetime import datetime, timezone, timedelta
from typing import Any, Optional, Type, TypeVar

from .types import (
    MemoryType,
    MemoryEntry,
    ConversationMemory,
    WorkingMemory,
    EntityMemory,
    UserMemory,
    DomainMemory,
)
from .schema import DomainMemorySchema
from .observability import (
    MemoryLogger,
    record_store_connected,
    record_store_error,
    set_store_entries,
)

T = TypeVar("T", bound=MemoryEntry)


class MemoryStore(ABC):
    """Abstract base class for memory storage backends."""
    
    @abstractmethod
    async def connect(self):
        """Initialize connection to the store."""
        pass
    
    @abstractmethod
    async def disconnect(self):
        """Close connection to the store."""
        pass
    
    @abstractmethod
    async def save(self, entry: MemoryEntry) -> str:
        """Save a memory entry. Returns the entry ID."""
        pass
    
    @abstractmethod
    async def get(self, entry_id: str, entry_type: Type[T]) -> Optional[T]:
        """Retrieve a memory entry by ID."""
        pass
    
    @abstractmethod
    async def delete(self, entry_id: str) -> bool:
        """Delete a memory entry. Returns success status."""
        pass
    
    @abstractmethod
    async def query(
        self,
        memory_type: MemoryType,
        agent_id: str,
        filters: dict[str, Any] = None,
        limit: int = 100,
    ) -> list[MemoryEntry]:
        """Query memory entries."""
        pass
    
    @abstractmethod
    async def save_schema(self, schema: DomainMemorySchema) -> str:
        """Save a domain memory schema."""
        pass
    
    @abstractmethod
    async def get_schema(self, schema_id: str) -> Optional[DomainMemorySchema]:
        """Retrieve a domain memory schema."""
        pass
    
    @abstractmethod
    async def get_schema_by_agent(
        self,
        agent_id: str,
        session_id: str = None,
    ) -> Optional[DomainMemorySchema]:
        """Get the current schema for an agent."""
        pass


class InMemoryStore(MemoryStore):
    """
    In-memory store for development and testing.
    
    Data is lost on restart - use only for dev/test!
    """
    
    def __init__(self, agent_id: str = "unknown"):
        self._entries: dict[str, dict] = {}
        self._schemas: dict[str, dict] = {}
        self._connected = False
        self._agent_id = agent_id
        self.log = MemoryLogger(agent_id=agent_id, component="store.in_memory")
    
    async def connect(self):
        self._connected = True
        record_store_connected(self._agent_id, "in_memory", True)
        self.log.info("memory_store_connected", store="in_memory")
    
    async def disconnect(self):
        self._connected = False
        entry_count = len(self._entries)
        self._entries.clear()
        self._schemas.clear()
        record_store_connected(self._agent_id, "in_memory", False)
        self.log.info("memory_store_disconnected", store="in_memory", cleared_entries=entry_count)
    
    async def save(self, entry: MemoryEntry) -> str:
        entry.updated_at = datetime.now(timezone.utc)
        self._entries[entry.id] = entry.model_dump(mode="json")
        return entry.id
    
    async def get(self, entry_id: str, entry_type: Type[T]) -> Optional[T]:
        data = self._entries.get(entry_id)
        if data is None:
            return None
        return entry_type.model_validate(data)
    
    async def delete(self, entry_id: str) -> bool:
        if entry_id in self._entries:
            del self._entries[entry_id]
            return True
        return False
    
    async def query(
        self,
        memory_type: MemoryType,
        agent_id: str,
        filters: dict[str, Any] = None,
        limit: int = 100,
    ) -> list[MemoryEntry]:
        results = []
        filters = filters or {}
        
        for entry_data in self._entries.values():
            if entry_data.get("memory_type") != memory_type.value:
                continue
            if entry_data.get("agent_id") != agent_id:
                continue
            
            # Apply additional filters
            match = True
            for key, value in filters.items():
                if entry_data.get(key) != value:
                    match = False
                    break
            
            if match:
                # Determine the correct type class
                type_class = self._get_type_class(memory_type)
                results.append(type_class.model_validate(entry_data))
            
            if len(results) >= limit:
                break
        
        return results
    
    async def save_schema(self, schema: DomainMemorySchema) -> str:
        schema.updated_at = datetime.now(timezone.utc)
        self._schemas[schema.schema_id] = schema.model_dump(mode="json")
        return schema.schema_id
    
    async def get_schema(self, schema_id: str) -> Optional[DomainMemorySchema]:
        data = self._schemas.get(schema_id)
        if data is None:
            return None
        
        from .schema import get_schema_class
        schema_class = get_schema_class(data.get("agent_type", "default"))
        return schema_class.model_validate(data)
    
    async def get_schema_by_agent(
        self,
        agent_id: str,
        session_id: str = None,
    ) -> Optional[DomainMemorySchema]:
        for schema_data in self._schemas.values():
            if schema_data.get("agent_id") != agent_id:
                continue
            if session_id and schema_data.get("session_id") != session_id:
                continue
            
            from .schema import get_schema_class
            schema_class = get_schema_class(schema_data.get("agent_type", "default"))
            return schema_class.model_validate(schema_data)
        
        return None
    
    def _get_type_class(self, memory_type: MemoryType) -> Type[MemoryEntry]:
        """Get the appropriate class for a memory type."""
        mapping = {
            MemoryType.SHORT_TERM: ConversationMemory,
            MemoryType.WORKING: WorkingMemory,
            MemoryType.ENTITY: EntityMemory,
            MemoryType.USER: UserMemory,
            MemoryType.LONG_TERM: DomainMemory,
        }
        return mapping.get(memory_type, MemoryEntry)


class RedisMemoryStore(MemoryStore):
    """
    Redis-based store for short-term and working memory.
    
    Features:
    - Fast read/write with latency tracking
    - TTL support for automatic expiration
    - Good for conversation context and session state
    - Full observability with metrics and tracing
    """
    
    def __init__(
        self,
        url: str = None,
        default_ttl: int = 86400,  # 24 hours
        prefix: str = "agent_memory:",
        agent_id: str = "unknown",
    ):
        self.url = url or os.getenv("REDIS_URL", "redis://localhost:6379/0")
        self.default_ttl = default_ttl
        self.prefix = prefix
        self._client = None
        self._agent_id = agent_id
        self.log = MemoryLogger(agent_id=agent_id, component="store.redis")
    
    async def connect(self):
        start_time = time.perf_counter()
        try:
            import redis.asyncio as redis
            self._client = redis.from_url(self.url, decode_responses=True)
            await self._client.ping()
            
            duration_ms = (time.perf_counter() - start_time) * 1000
            record_store_connected(self._agent_id, "redis", True)
            self.log.info(
                "memory_store_connected",
                store="redis",
                duration_ms=round(duration_ms, 2),
            )
        except ImportError:
            record_store_error(self._agent_id, "redis", "import_error")
            self.log.error("redis_not_installed", hint="pip install redis")
            raise
        except Exception as e:
            record_store_connected(self._agent_id, "redis", False)
            record_store_error(self._agent_id, "redis", "connection_error")
            self.log.error("redis_connection_failed", error=str(e))
            raise
    
    async def disconnect(self):
        if self._client:
            await self._client.close()
            self._client = None
            record_store_connected(self._agent_id, "redis", False)
            self.log.info("memory_store_disconnected", store="redis")
    
    def _key(self, entry_id: str) -> str:
        """Generate Redis key for an entry."""
        return f"{self.prefix}{entry_id}"
    
    def _schema_key(self, schema_id: str) -> str:
        """Generate Redis key for a schema."""
        return f"{self.prefix}schema:{schema_id}"
    
    def _agent_schema_key(self, agent_id: str, session_id: str = None) -> str:
        """Generate Redis key for agent's current schema."""
        if session_id:
            return f"{self.prefix}agent:{agent_id}:session:{session_id}"
        return f"{self.prefix}agent:{agent_id}:current"
    
    async def save(self, entry: MemoryEntry) -> str:
        if not self._client:
            raise RuntimeError("Redis not connected")
        
        entry.updated_at = datetime.now(timezone.utc)
        data = entry.model_dump(mode="json")
        
        key = self._key(entry.id)
        await self._client.set(key, json.dumps(data))
        
        # Set TTL based on memory type
        ttl = self._get_ttl(entry)
        if ttl:
            await self._client.expire(key, ttl)
        
        # Index by agent_id and memory_type
        index_key = f"{self.prefix}index:{entry.agent_id}:{entry.memory_type.value}"
        await self._client.sadd(index_key, entry.id)
        
        return entry.id
    
    async def get(self, entry_id: str, entry_type: Type[T]) -> Optional[T]:
        if not self._client:
            raise RuntimeError("Redis not connected")
        
        data = await self._client.get(self._key(entry_id))
        if data is None:
            return None
        
        return entry_type.model_validate(json.loads(data))
    
    async def delete(self, entry_id: str) -> bool:
        if not self._client:
            raise RuntimeError("Redis not connected")
        
        key = self._key(entry_id)
        result = await self._client.delete(key)
        return result > 0
    
    async def query(
        self,
        memory_type: MemoryType,
        agent_id: str,
        filters: dict[str, Any] = None,
        limit: int = 100,
    ) -> list[MemoryEntry]:
        if not self._client:
            raise RuntimeError("Redis not connected")
        
        index_key = f"{self.prefix}index:{agent_id}:{memory_type.value}"
        entry_ids = await self._client.smembers(index_key)
        
        results = []
        type_class = self._get_type_class(memory_type)
        
        for entry_id in list(entry_ids)[:limit]:
            data = await self._client.get(self._key(entry_id))
            if data:
                entry = type_class.model_validate(json.loads(data))
                
                # Apply filters
                if filters:
                    match = all(
                        getattr(entry, k, None) == v
                        for k, v in filters.items()
                    )
                    if not match:
                        continue
                
                results.append(entry)
        
        return results
    
    async def save_schema(self, schema: DomainMemorySchema) -> str:
        if not self._client:
            raise RuntimeError("Redis not connected")
        
        schema.updated_at = datetime.now(timezone.utc)
        data = schema.model_dump(mode="json")
        
        # Save schema data
        await self._client.set(self._schema_key(schema.schema_id), json.dumps(data))
        
        # Update agent's current schema pointer
        agent_key = self._agent_schema_key(schema.agent_id, schema.session_id)
        await self._client.set(agent_key, schema.schema_id)
        
        return schema.schema_id
    
    async def get_schema(self, schema_id: str) -> Optional[DomainMemorySchema]:
        if not self._client:
            raise RuntimeError("Redis not connected")
        
        data = await self._client.get(self._schema_key(schema_id))
        if data is None:
            return None
        
        schema_data = json.loads(data)
        from .schema import get_schema_class
        schema_class = get_schema_class(schema_data.get("agent_type", "default"))
        return schema_class.model_validate(schema_data)
    
    async def get_schema_by_agent(
        self,
        agent_id: str,
        session_id: str = None,
    ) -> Optional[DomainMemorySchema]:
        if not self._client:
            raise RuntimeError("Redis not connected")
        
        agent_key = self._agent_schema_key(agent_id, session_id)
        schema_id = await self._client.get(agent_key)
        
        if schema_id is None:
            return None
        
        return await self.get_schema(schema_id)
    
    def _get_ttl(self, entry: MemoryEntry) -> Optional[int]:
        """Get TTL based on memory type."""
        ttl_map = {
            MemoryType.SHORT_TERM: 3600,  # 1 hour
            MemoryType.WORKING: 86400,    # 24 hours
            MemoryType.EPISODIC: 604800,  # 7 days
        }
        return ttl_map.get(entry.memory_type, self.default_ttl)
    
    def _get_type_class(self, memory_type: MemoryType) -> Type[MemoryEntry]:
        """Get the appropriate class for a memory type."""
        mapping = {
            MemoryType.SHORT_TERM: ConversationMemory,
            MemoryType.WORKING: WorkingMemory,
            MemoryType.ENTITY: EntityMemory,
            MemoryType.USER: UserMemory,
            MemoryType.LONG_TERM: DomainMemory,
        }
        return mapping.get(memory_type, MemoryEntry)


class PostgresMemoryStore(MemoryStore):
    """
    PostgreSQL-based store for long-term, entity, and user memory.
    
    Features:
    - Durable storage with connection pool
    - Complex queries with SQL
    - JSON support for flexible schemas
    - Good for accumulated knowledge and user data
    - Full observability with metrics and tracing
    """
    
    def __init__(
        self,
        url: str = None,
        pool_size: int = 5,
        agent_id: str = "unknown",
    ):
        self.url = url or os.getenv(
            "POSTGRES_URL",
            "postgresql://postgres:postgres@localhost:5432/agent_memory"
        )
        self.pool_size = pool_size
        self._pool = None
        self._agent_id = agent_id
        self.log = MemoryLogger(agent_id=agent_id, component="store.postgres")
    
    async def connect(self):
        start_time = time.perf_counter()
        try:
            import asyncpg
            self._pool = await asyncpg.create_pool(self.url, min_size=1, max_size=self.pool_size)
            
            # Create tables if they don't exist
            await self._init_schema()
            
            duration_ms = (time.perf_counter() - start_time) * 1000
            record_store_connected(self._agent_id, "postgres", True)
            self.log.info(
                "memory_store_connected",
                store="postgres",
                pool_size=self.pool_size,
                duration_ms=round(duration_ms, 2),
            )
        except ImportError:
            record_store_error(self._agent_id, "postgres", "import_error")
            self.log.error("asyncpg_not_installed", hint="pip install asyncpg")
            raise
        except Exception as e:
            record_store_connected(self._agent_id, "postgres", False)
            record_store_error(self._agent_id, "postgres", "connection_error")
            self.log.error("postgres_connection_failed", error=str(e))
            raise
    
    async def _init_schema(self):
        """Initialize database schema."""
        async with self._pool.acquire() as conn:
            await conn.execute("""
                CREATE TABLE IF NOT EXISTS memory_entries (
                    id TEXT PRIMARY KEY,
                    memory_type TEXT NOT NULL,
                    agent_id TEXT NOT NULL,
                    data JSONB NOT NULL,
                    created_at TIMESTAMPTZ DEFAULT NOW(),
                    updated_at TIMESTAMPTZ DEFAULT NOW(),
                    expires_at TIMESTAMPTZ
                );
                
                CREATE INDEX IF NOT EXISTS idx_memory_agent_type 
                ON memory_entries(agent_id, memory_type);
                
                CREATE TABLE IF NOT EXISTS domain_schemas (
                    id TEXT PRIMARY KEY,
                    agent_id TEXT NOT NULL,
                    agent_type TEXT NOT NULL,
                    session_id TEXT,
                    data JSONB NOT NULL,
                    created_at TIMESTAMPTZ DEFAULT NOW(),
                    updated_at TIMESTAMPTZ DEFAULT NOW()
                );
                
                CREATE INDEX IF NOT EXISTS idx_schema_agent 
                ON domain_schemas(agent_id);
                
                CREATE INDEX IF NOT EXISTS idx_schema_agent_session 
                ON domain_schemas(agent_id, session_id);
            """)
    
    async def disconnect(self):
        if self._pool:
            await self._pool.close()
            self._pool = None
            record_store_connected(self._agent_id, "postgres", False)
            self.log.info("memory_store_disconnected", store="postgres")
    
    async def save(self, entry: MemoryEntry) -> str:
        if not self._pool:
            raise RuntimeError("PostgreSQL not connected")
        
        entry.updated_at = datetime.now(timezone.utc)
        data = entry.model_dump(mode="json")
        
        async with self._pool.acquire() as conn:
            await conn.execute("""
                INSERT INTO memory_entries (id, memory_type, agent_id, data, created_at, updated_at, expires_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7)
                ON CONFLICT (id) DO UPDATE SET
                    data = $4,
                    updated_at = $6
            """,
                entry.id,
                entry.memory_type.value,
                entry.agent_id,
                json.dumps(data),
                entry.created_at,
                entry.updated_at,
                entry.expires_at,
            )
        
        return entry.id
    
    async def get(self, entry_id: str, entry_type: Type[T]) -> Optional[T]:
        if not self._pool:
            raise RuntimeError("PostgreSQL not connected")
        
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(
                "SELECT data FROM memory_entries WHERE id = $1",
                entry_id
            )
        
        if row is None:
            return None
        
        return entry_type.model_validate(json.loads(row["data"]))
    
    async def delete(self, entry_id: str) -> bool:
        if not self._pool:
            raise RuntimeError("PostgreSQL not connected")
        
        async with self._pool.acquire() as conn:
            result = await conn.execute(
                "DELETE FROM memory_entries WHERE id = $1",
                entry_id
            )
        
        return result == "DELETE 1"
    
    async def query(
        self,
        memory_type: MemoryType,
        agent_id: str,
        filters: dict[str, Any] = None,
        limit: int = 100,
    ) -> list[MemoryEntry]:
        if not self._pool:
            raise RuntimeError("PostgreSQL not connected")
        
        query = """
            SELECT data FROM memory_entries
            WHERE agent_id = $1 AND memory_type = $2
            ORDER BY updated_at DESC
            LIMIT $3
        """
        
        async with self._pool.acquire() as conn:
            rows = await conn.fetch(query, agent_id, memory_type.value, limit)
        
        type_class = self._get_type_class(memory_type)
        results = []
        
        for row in rows:
            entry = type_class.model_validate(json.loads(row["data"]))
            
            # Apply additional filters
            if filters:
                match = all(
                    getattr(entry, k, None) == v
                    for k, v in filters.items()
                )
                if not match:
                    continue
            
            results.append(entry)
        
        return results
    
    async def save_schema(self, schema: DomainMemorySchema) -> str:
        if not self._pool:
            raise RuntimeError("PostgreSQL not connected")
        
        schema.updated_at = datetime.now(timezone.utc)
        data = schema.model_dump(mode="json")
        
        async with self._pool.acquire() as conn:
            await conn.execute("""
                INSERT INTO domain_schemas (id, agent_id, agent_type, session_id, data, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7)
                ON CONFLICT (id) DO UPDATE SET
                    data = $5,
                    updated_at = $7
            """,
                schema.schema_id,
                schema.agent_id,
                schema.agent_type,
                schema.session_id,
                json.dumps(data),
                schema.created_at,
                schema.updated_at,
            )
        
        return schema.schema_id
    
    async def get_schema(self, schema_id: str) -> Optional[DomainMemorySchema]:
        if not self._pool:
            raise RuntimeError("PostgreSQL not connected")
        
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(
                "SELECT data, agent_type FROM domain_schemas WHERE id = $1",
                schema_id
            )
        
        if row is None:
            return None
        
        from .schema import get_schema_class
        schema_class = get_schema_class(row["agent_type"])
        return schema_class.model_validate(json.loads(row["data"]))
    
    async def get_schema_by_agent(
        self,
        agent_id: str,
        session_id: str = None,
    ) -> Optional[DomainMemorySchema]:
        if not self._pool:
            raise RuntimeError("PostgreSQL not connected")
        
        if session_id:
            query = """
                SELECT data, agent_type FROM domain_schemas
                WHERE agent_id = $1 AND session_id = $2
                ORDER BY updated_at DESC
                LIMIT 1
            """
            params = (agent_id, session_id)
        else:
            query = """
                SELECT data, agent_type FROM domain_schemas
                WHERE agent_id = $1
                ORDER BY updated_at DESC
                LIMIT 1
            """
            params = (agent_id,)
        
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(query, *params)
        
        if row is None:
            return None
        
        from .schema import get_schema_class
        schema_class = get_schema_class(row["agent_type"])
        return schema_class.model_validate(json.loads(row["data"]))
    
    def _get_type_class(self, memory_type: MemoryType) -> Type[MemoryEntry]:
        """Get the appropriate class for a memory type."""
        mapping = {
            MemoryType.SHORT_TERM: ConversationMemory,
            MemoryType.WORKING: WorkingMemory,
            MemoryType.ENTITY: EntityMemory,
            MemoryType.USER: UserMemory,
            MemoryType.LONG_TERM: DomainMemory,
        }
        return mapping.get(memory_type, MemoryEntry)
