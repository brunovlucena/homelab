"""
🔴 Redis Session Store

Manages short-term conversation memory in Redis with TTL.
"""

import json
import logging
from datetime import datetime, timedelta
from typing import List, Dict, Any, Optional
import redis.asyncio as redis

logger = logging.getLogger(__name__)


class RedisStore:
    """Redis-based session storage for conversations"""

    def __init__(self, redis_url: str, ttl: int = 86400):
        """
        Initialize Redis store
        
        Args:
            redis_url: Redis connection URL
            ttl: Session TTL in seconds (default: 24 hours)
        """
        self.redis_url = redis_url
        self.ttl = ttl
        self.client: Optional[redis.Redis] = None
        self.key_prefix = "bruno:session"
        self._connected = False

    async def connect(self):
        """Connect to Redis (non-blocking if unavailable)"""
        try:
            self.client = await redis.from_url(
                self.redis_url,
                encoding="utf-8",
                decode_responses=True
            )
            await self.client.ping()
            self._connected = True
            logger.info("✅ Connected to Redis")
        except Exception as e:
            self._connected = False
            logger.warning(f"⚠️  Redis unavailable: {e} - service will continue with degraded functionality")
            self.client = None

    async def disconnect(self):
        """Disconnect from Redis"""
        if self.client:
            await self.client.close()
            logger.info("🔌 Disconnected from Redis")

    def _make_key(self, ip: str) -> str:
        """Make Redis key for IP"""
        return f"{self.key_prefix}:{ip}"

    async def save_message(
        self,
        ip: str,
        message: str,
        response: str,
        context: Dict[str, Any] = None
    ):
        """
        Save a message to session
        
        Args:
            ip: User IP address
            message: User message
            response: Agent response
            context: Additional context
        """
        if not self.client or not self._connected:
            logger.debug(f"⚠️  Redis unavailable, skipping save for IP: {ip}")
            return

        key = self._make_key(ip)

        entry = {
            "timestamp": datetime.utcnow().isoformat(),
            "message": message,
            "response": response,
            "context": context or {}
        }

        try:
            # Add to list
            await self.client.rpush(key, json.dumps(entry))
            
            # Set TTL
            await self.client.expire(key, self.ttl)
            
            logger.debug(f"💾 Saved message for IP: {ip}")
        except Exception as e:
            logger.warning(f"⚠️  Redis operation failed: {e}")
            self._connected = False

    async def get_session(self, ip: str, limit: int = 10) -> List[Dict[str, Any]]:
        """
        Get recent session messages for IP
        
        Args:
            ip: User IP address
            limit: Maximum number of messages to return
            
        Returns:
            List of message dictionaries
        """
        if not self.client or not self._connected:
            logger.debug(f"⚠️  Redis unavailable, returning empty session for IP: {ip}")
            return []

        key = self._make_key(ip)

        try:
            # Get last N messages
            messages = await self.client.lrange(key, -limit, -1)
            
            return [json.loads(msg) for msg in messages]
        except Exception as e:
            logger.warning(f"⚠️  Redis operation failed: {e}")
            self._connected = False
            return []

    async def clear_session(self, ip: str):
        """
        Clear session for IP
        
        Args:
            ip: User IP address
        """
        if not self.client or not self._connected:
            logger.debug(f"⚠️  Redis unavailable, cannot clear session for IP: {ip}")
            return

        key = self._make_key(ip)

        try:
            await self.client.delete(key)
            logger.info(f"🗑️ Cleared session for IP: {ip}")
        except Exception as e:
            logger.warning(f"⚠️  Redis operation failed: {e}")
            self._connected = False

    async def get_active_sessions(self) -> int:
        """
        Get count of active sessions
        
        Returns:
            Number of active sessions
        """
        if not self.client or not self._connected:
            logger.debug("⚠️  Redis unavailable, returning 0 active sessions")
            return 0

        try:
            keys = await self.client.keys(f"{self.key_prefix}:*")
            return len(keys)
        except Exception as e:
            logger.warning(f"⚠️  Redis operation failed: {e}")
            self._connected = False
            return 0

    async def health_check(self) -> bool:
        """
        Check Redis health (and attempt reconnection if needed)
        
        Returns:
            True if healthy, False otherwise
        """
        try:
            if not self.client:
                # Attempt reconnection
                await self.connect()
                return self._connected
            
            await self.client.ping()
            self._connected = True
            return True
        except Exception as e:
            logger.debug(f"⚠️  Redis health check failed: {e}")
            self._connected = False
            return False

