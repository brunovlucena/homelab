# FutBoss AI - User Model
# Author: Bruno Lucena (bruno@lucena.cloud)

from pydantic import BaseModel, EmailStr, Field
from typing import Optional
from datetime import datetime
from bson import ObjectId


class UserCreate(BaseModel):
    username: str = Field(..., min_length=3, max_length=30)
    email: EmailStr
    password: str = Field(..., min_length=6)


class UserLogin(BaseModel):
    email: EmailStr
    password: str


class User(BaseModel):
    id: Optional[str] = Field(default=None, alias="_id")
    username: str
    email: EmailStr
    hashed_password: str
    tokens: int = 1000
    created_at: datetime = Field(default_factory=datetime.utcnow)
    is_active: bool = True

    class Config:
        populate_by_name = True
        json_encoders = {ObjectId: str}


class UserResponse(BaseModel):
    id: str
    username: str
    email: EmailStr
    tokens: int
    created_at: datetime

    class Config:
        from_attributes = True

