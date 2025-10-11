"""
🧠 Memory Manager

Coordinates session (Redis) and persistent (MongoDB) memory stores.
"""

import logging
from typing import List, Dict, Any
from .redis_store import RedisStore
from .mongo_store import MongoStore

logger = logging.getLogger(__name__)


class MemoryManager:
    """Manages conversation memory across Redis and MongoDB"""

    def __init__(
        self,
        redis_url: str,
        mongodb_url: str,
        mongodb_db: str = "agent_bruno",
        session_ttl: int = 86400
    ):
        """
        Initialize memory manager
        
        Args:
            redis_url: Redis connection URL
            mongodb_url: MongoDB connection URL
            mongodb_db: MongoDB database name
            session_ttl: Session TTL in seconds (default: 24 hours)
        """
        self.redis_store = RedisStore(redis_url, ttl=session_ttl)
        self.mongo_store = MongoStore(mongodb_url, db_name=mongodb_db)

    async def connect(self):
        """Connect to both stores"""
        await self.redis_store.connect()
        await self.mongo_store.connect()
        logger.info("✅ Memory manager connected")

    async def disconnect(self):
        """Disconnect from both stores"""
        await self.redis_store.disconnect()
        await self.mongo_store.disconnect()
        logger.info("🔌 Memory manager disconnected")

    async def save(
        self,
        ip: str,
        message: str,
        response: str,
        context: Dict[str, Any] = None
    ):
        """
        Save conversation to both session and persistent stores
        
        Args:
            ip: User IP address
            message: User message
            response: Agent response
            context: Additional context
        """
        # Save to both stores in parallel
        await self.redis_store.save_message(ip, message, response, context)
        await self.mongo_store.save_conversation(ip, message, response, context)
        logger.debug(f"💾 Saved conversation for IP: {ip}")

    async def get_recent_context(
        self,
        ip: str,
        limit: int = 5
    ) -> List[Dict[str, Any]]:
        """
        Get recent conversation context from session store
        
        Args:
            ip: User IP address
            limit: Maximum number of messages to return
            
        Returns:
            List of recent messages
        """
        return await self.redis_store.get_session(ip, limit)

    async def get_full_history(
        self,
        ip: str,
        limit: int = 50,
        skip: int = 0
    ) -> List[Dict[str, Any]]:
        """
        Get full conversation history from persistent store
        
        Args:
            ip: User IP address
            limit: Maximum number of conversations to return
            skip: Number of conversations to skip
            
        Returns:
            List of conversations
        """
        return await self.mongo_store.get_conversation_history(ip, limit, skip)

    async def clear_memory(self, ip: str):
        """
        Clear all memory for IP
        
        Args:
            ip: User IP address
        """
        await self.redis_store.clear_session(ip)
        await self.mongo_store.delete_conversation_history(ip)
        logger.info(f"🗑️ Cleared all memory for IP: {ip}")

    async def get_stats(self) -> Dict[str, Any]:
        """
        Get memory statistics
        
        Returns:
            Statistics dictionary
        """
        active_sessions = await self.redis_store.get_active_sessions()
        total_conversations = await self.mongo_store.get_total_conversations()
        unique_ips = await self.mongo_store.get_unique_ips()

        return {
            "active_sessions": active_sessions,
            "total_conversations": total_conversations,
            "unique_ips": len(unique_ips),
            "unique_ip_list": unique_ips
        }

    async def health_check(self) -> Dict[str, bool]:
        """
        Check health of both stores
        
        Returns:
            Health status dictionary
        """
        redis_healthy = await self.redis_store.health_check()
        mongo_healthy = await self.mongo_store.health_check()

        return {
            "redis": redis_healthy,
            "mongodb": mongo_healthy,
            "overall": redis_healthy and mongo_healthy
        }

    def format_context_for_prompt(
        self,
        recent_messages: List[Dict[str, Any]]
    ) -> str:
        """
        Format recent messages for LLM prompt
        
        Args:
            recent_messages: List of recent messages
            
        Returns:
            Formatted context string
        """
        if not recent_messages:
            return "No previous conversation context."

        context_parts = ["Previous conversation:"]
        
        for msg in recent_messages:
            timestamp = msg.get("timestamp", "")
            user_msg = msg.get("message", "")
            agent_resp = msg.get("response", "")
            
            context_parts.append(f"\nUser: {user_msg}")
            context_parts.append(f"Agent: {agent_resp}")

        return "\n".join(context_parts)

