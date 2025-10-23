"""
🍃 MongoDB Persistent Store

Manages long-term conversation memory in MongoDB.
"""

import logging
from datetime import datetime
from typing import Any, Dict, List, Optional

from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorDatabase

logger = logging.getLogger(__name__)


class MongoStore:
    """MongoDB-based persistent storage for conversations"""

    def __init__(self, mongodb_url: str, db_name: str = "agent_bruno"):
        """
        Initialize MongoDB store

        Args:
            mongodb_url: MongoDB connection URL
            db_name: Database name
        """
        self.mongodb_url = mongodb_url
        self.db_name = db_name
        self.client: Optional[AsyncIOMotorClient] = None
        self.db: Optional[AsyncIOMotorDatabase] = None
        self.collection_name = "conversations"
        self._connected = False

    async def connect(self):
        """Connect to MongoDB (non-blocking if unavailable)"""
        try:
            self.client = AsyncIOMotorClient(self.mongodb_url)
            self.db = self.client[self.db_name]

            # Create indexes
            await self._create_indexes()

            # Test connection
            await self.client.admin.command("ping")

            self._connected = True
            logger.info(f"✅ Connected to MongoDB: {self.db_name}")
        except Exception as e:
            self._connected = False
            logger.warning(
                f"⚠️  MongoDB unavailable: {e} - service will continue with degraded functionality"
            )
            self.client = None
            self.db = None

    async def _create_indexes(self):
        """Create database indexes"""
        collection = self.db[self.collection_name]

        # Index on IP address
        await collection.create_index("ip")

        # Index on timestamp
        await collection.create_index("timestamp")

        # Compound index on IP + timestamp
        await collection.create_index([("ip", 1), ("timestamp", -1)])

        logger.info("✅ Created MongoDB indexes")

    async def disconnect(self):
        """Disconnect from MongoDB"""
        if self.client:
            self.client.close()
            logger.info("🔌 Disconnected from MongoDB")

    async def save_conversation(
        self, ip: str, message: str, response: str, context: Dict[str, Any] = None
    ):
        """
        Save a conversation to persistent storage

        Args:
            ip: User IP address
            message: User message
            response: Agent response
            context: Additional context
        """
        if not self.db or not self._connected:
            logger.debug(f"⚠️  MongoDB unavailable, skipping save for IP: {ip}")
            return

        collection = self.db[self.collection_name]

        document = {
            "ip": ip,
            "timestamp": datetime.utcnow(),
            "message": message,
            "response": response,
            "context": context or {},
            "created_at": datetime.utcnow(),
        }

        try:
            await collection.insert_one(document)
            logger.debug(f"💾 Saved conversation for IP: {ip}")
        except Exception as e:
            logger.warning(f"⚠️  MongoDB operation failed: {e}")
            self._connected = False

    async def get_conversation_history(
        self, ip: str, limit: int = 50, skip: int = 0
    ) -> List[Dict[str, Any]]:
        """
        Get conversation history for IP

        Args:
            ip: User IP address
            limit: Maximum number of conversations to return
            skip: Number of conversations to skip

        Returns:
            List of conversation dictionaries
        """
        if not self.db or not self._connected:
            logger.debug(
                f"⚠️  MongoDB unavailable, returning empty history for IP: {ip}"
            )
            return []

        collection = self.db[self.collection_name]

        try:
            cursor = (
                collection.find({"ip": ip})
                .sort("timestamp", -1)
                .skip(skip)
                .limit(limit)
            )

            conversations = await cursor.to_list(length=limit)

            # Convert ObjectId to string for JSON serialization
            for conv in conversations:
                conv["_id"] = str(conv["_id"])

            return conversations
        except Exception as e:
            logger.warning(f"⚠️  MongoDB operation failed: {e}")
            self._connected = False
            return []

    async def delete_conversation_history(self, ip: str):
        """
        Delete all conversations for IP

        Args:
            ip: User IP address
        """
        if not self.db or not self._connected:
            logger.debug(f"⚠️  MongoDB unavailable, cannot delete history for IP: {ip}")
            return

        collection = self.db[self.collection_name]

        try:
            result = await collection.delete_many({"ip": ip})
            logger.info(f"🗑️ Deleted {result.deleted_count} conversations for IP: {ip}")
        except Exception as e:
            logger.warning(f"⚠️  MongoDB operation failed: {e}")
            self._connected = False

    async def get_total_conversations(self, ip: Optional[str] = None) -> int:
        """
        Get total number of conversations

        Args:
            ip: Optional IP to filter by

        Returns:
            Number of conversations
        """
        if not self.db or not self._connected:
            logger.debug("⚠️  MongoDB unavailable, returning 0 conversations")
            return 0

        collection = self.db[self.collection_name]

        try:
            query = {"ip": ip} if ip else {}
            count = await collection.count_documents(query)
            return count
        except Exception as e:
            logger.warning(f"⚠️  MongoDB operation failed: {e}")
            self._connected = False
            return 0

    async def get_unique_ips(self) -> List[str]:
        """
        Get list of unique IP addresses

        Returns:
            List of IP addresses
        """
        if not self.db or not self._connected:
            logger.debug("⚠️  MongoDB unavailable, returning empty IP list")
            return []

        collection = self.db[self.collection_name]

        try:
            ips = await collection.distinct("ip")
            return ips
        except Exception as e:
            logger.warning(f"⚠️  MongoDB operation failed: {e}")
            self._connected = False
            return []

    async def health_check(self) -> bool:
        """
        Check MongoDB health (and attempt reconnection if needed)

        Returns:
            True if healthy, False otherwise
        """
        try:
            if not self.client:
                # Attempt reconnection
                await self.connect()
                return self._connected

            await self.client.admin.command("ping")
            self._connected = True
            return True
        except Exception as e:
            logger.debug(f"⚠️  MongoDB health check failed: {e}")
            self._connected = False
            return False
