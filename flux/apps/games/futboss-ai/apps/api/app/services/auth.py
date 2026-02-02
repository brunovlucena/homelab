# FutBoss AI - Authentication Service
# Author: Bruno Lucena (bruno@lucena.cloud)

from datetime import datetime, timedelta
from typing import Optional
from jose import JWTError, jwt
from passlib.context import CryptContext
from bson import ObjectId

from app.config import get_settings
from app.models.user import User, UserCreate
from app.services.database import users_collection

settings = get_settings()
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")


class AuthService:
    def hash_password(self, password: str) -> str:
        return pwd_context.hash(password)

    def verify_password(self, plain: str, hashed: str) -> bool:
        return pwd_context.verify(plain, hashed)

    def create_token(self, user_id: str) -> str:
        expire = datetime.utcnow() + timedelta(hours=settings.jwt_expiration_hours)
        payload = {"sub": user_id, "exp": expire}
        return jwt.encode(payload, settings.jwt_secret, algorithm=settings.jwt_algorithm)

    def decode_token(self, token: str) -> Optional[str]:
        try:
            payload = jwt.decode(token, settings.jwt_secret, algorithms=[settings.jwt_algorithm])
            return payload.get("sub")
        except JWTError:
            return None

    async def register(self, user_data: UserCreate) -> Optional[User]:
        # Check if user exists
        existing = await users_collection.find_one({
            "$or": [
                {"email": user_data.email},
                {"username": user_data.username}
            ]
        })
        if existing:
            return None

        # Create user
        user = User(
            username=user_data.username,
            email=user_data.email,
            hashed_password=self.hash_password(user_data.password),
            tokens=settings.initial_tokens,
        )
        result = await users_collection.insert_one(user.model_dump(exclude={"id"}))
        user.id = str(result.inserted_id)
        return user

    async def authenticate(self, email: str, password: str) -> Optional[User]:
        user_data = await users_collection.find_one({"email": email})
        if not user_data:
            return None
        if not self.verify_password(password, user_data["hashed_password"]):
            return None
        user_data["_id"] = str(user_data["_id"])
        return User(**user_data)

    async def get_current_user(self, token: str) -> Optional[User]:
        user_id = self.decode_token(token)
        if not user_id:
            return None
        user_data = await users_collection.find_one({"_id": ObjectId(user_id)})
        if not user_data:
            return None
        user_data["_id"] = str(user_data["_id"])
        return User(**user_data)

    async def get_user_by_id(self, user_id: str) -> Optional[User]:
        user_data = await users_collection.find_one({"_id": ObjectId(user_id)})
        if not user_data:
            return None
        user_data["_id"] = str(user_data["_id"])
        return User(**user_data)

