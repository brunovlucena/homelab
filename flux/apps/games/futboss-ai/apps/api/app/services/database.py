# FutBoss AI - Database Service
# Author: Bruno Lucena (bruno@lucena.cloud)

from motor.motor_asyncio import AsyncIOMotorClient
from app.config import get_settings

settings = get_settings()

client = AsyncIOMotorClient(settings.mongodb_url)
db = client[settings.mongodb_database]

# Collections
users_collection = db["users"]
teams_collection = db["teams"]
players_collection = db["players"]
matches_collection = db["matches"]
transactions_collection = db["transactions"]
payments_collection = db["payments"]

