# FutBoss AI - Authentication Routes
# Author: Bruno Lucena (bruno@lucena.cloud)

from fastapi import APIRouter, HTTPException, Depends, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials

from app.models.user import UserCreate, UserLogin, UserResponse
from app.services.auth import AuthService

router = APIRouter()
security = HTTPBearer()
auth_service = AuthService()


@router.post("/register", response_model=dict, status_code=status.HTTP_201_CREATED)
async def register(user_data: UserCreate):
    """Register a new user"""
    user = await auth_service.register(user_data)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Email or username already exists"
        )
    token = auth_service.create_token(user.id)
    return {"access_token": token, "token_type": "bearer", "user": UserResponse(**user.model_dump())}


@router.post("/login", response_model=dict)
async def login(credentials: UserLogin):
    """Login with email and password"""
    user = await auth_service.authenticate(credentials.email, credentials.password)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid credentials"
        )
    token = auth_service.create_token(user.id)
    return {"access_token": token, "token_type": "bearer", "user": UserResponse(**user.model_dump())}


@router.get("/me", response_model=UserResponse)
async def get_current_user(credentials: HTTPAuthorizationCredentials = Depends(security)):
    """Get current authenticated user"""
    user = await auth_service.get_current_user(credentials.credentials)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid or expired token"
        )
    return UserResponse(**user.model_dump())

