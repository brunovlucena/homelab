"""Database connection and operations"""
import os
from typing import Optional, List, Dict, Any
from datetime import datetime
import structlog
from motor.motor_asyncio import AsyncIOMotorClient
from pymongo.errors import ConnectionFailure

from .types import User, Exercise, GameSession, Progress, GameStatus

logger = structlog.get_logger()


class Database:
    """MongoDB database connection"""
    
    def __init__(self):
        self.client: Optional[AsyncIOMotorClient] = None
        self.db = None
        self.mongodb_url = os.getenv("MONGODB_URL", os.getenv("MONGODB_URI", ""))
        self.mongodb_database = os.getenv("MONGODB_DATABASE", "speech_coach_db")
    
    async def connect(self):
        """Connect to MongoDB"""
        if not self.mongodb_url:
            logger.warning("mongodb_url_not_configured")
            return
        
        try:
            self.client = AsyncIOMotorClient(self.mongodb_url)
            self.db = self.client[self.mongodb_database]
            # Test connection
            await self.client.admin.command('ping')
            logger.info("database_connected", database=self.mongodb_database)
        except ConnectionFailure as e:
            logger.error("database_connection_failed", error=str(e))
            raise
    
    async def disconnect(self):
        """Disconnect from MongoDB"""
        if self.client:
            self.client.close()
            logger.info("database_disconnected")
    
    async def get_user(self, user_id: str) -> Optional[User]:
        """Get user by ID"""
        if not self.db:
            return None
        
        user_doc = await self.db.users.find_one({"id": user_id})
        if user_doc:
            return User(**user_doc)
        return None
    
    async def save_user(self, user: User):
        """Save or update user"""
        if not self.db:
            return
        
        await self.db.users.update_one(
            {"id": user.id},
            {"$set": user.model_dump()},
            upsert=True
        )
    
    async def get_exercise(self, exercise_id: str) -> Optional[Exercise]:
        """Get exercise by ID"""
        if not self.db:
            return None
        
        exercise_doc = await self.db.exercises.find_one({"id": exercise_id})
        if exercise_doc:
            return Exercise(**exercise_doc)
        return None
    
    async def get_exercises_by_type(self, exercise_type: str) -> List[Exercise]:
        """Get exercises by type"""
        if not self.db:
            return []
        
        cursor = self.db.exercises.find({"type": exercise_type})
        exercises = []
        async for doc in cursor:
            exercises.append(Exercise(**doc))
        return exercises
    
    async def save_session(self, session: GameSession):
        """Save or update game session"""
        if not self.db:
            return
        
        await self.db.sessions.update_one(
            {"id": session.id},
            {"$set": session.model_dump(mode="json")},
            upsert=True
        )
    
    async def get_user_sessions(self, user_id: str, limit: int = 20) -> List[GameSession]:
        """Get recent sessions for a user"""
        if not self.db:
            return []
        
        cursor = self.db.sessions.find(
            {"user_id": user_id}
        ).sort("started_at", -1).limit(limit)
        
        sessions = []
        async for doc in cursor:
            sessions.append(GameSession(**doc))
        return sessions
    
    async def get_progress(self, user_id: str) -> Optional[Progress]:
        """Get user progress"""
        if not self.db:
            return None
        
        progress_doc = await self.db.progress.find_one({"user_id": user_id})
        if progress_doc:
            return Progress(**progress_doc)
        return None
    
    async def update_progress(self, progress: Progress):
        """Update user progress"""
        if not self.db:
            return
        
        await self.db.progress.update_one(
            {"user_id": progress.user_id},
            {"$set": progress.model_dump(mode="json")},
            upsert=True
        )
