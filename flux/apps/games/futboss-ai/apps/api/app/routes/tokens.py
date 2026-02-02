# FutBoss AI - Token Routes
# Author: Bruno Lucena (bruno@lucena.cloud)

from fastapi import APIRouter, HTTPException, Depends
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from typing import List

from app.models.token import TokenWallet, TokenTransaction
from app.services.auth import AuthService
from app.services.token import TokenService

router = APIRouter()
security = HTTPBearer()
auth_service = AuthService()
token_service = TokenService()


async def get_current_user_id(credentials: HTTPAuthorizationCredentials = Depends(security)) -> str:
    user = await auth_service.get_current_user(credentials.credentials)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid token")
    return user.id


@router.get("/balance", response_model=dict)
async def get_balance(user_id: str = Depends(get_current_user_id)):
    """Get user's token balance"""
    wallet = await token_service.get_wallet(user_id)
    return {
        "balance": wallet.balance,
        "total_earned": wallet.total_earned,
        "total_spent": wallet.total_spent
    }


@router.get("/transactions", response_model=List[TokenTransaction])
async def get_transactions(user_id: str = Depends(get_current_user_id), skip: int = 0, limit: int = 50):
    """Get user's transaction history"""
    return await token_service.get_transactions(user_id, skip, limit)


@router.post("/transfer", response_model=TokenTransaction)
async def transfer_tokens(to_user_id: str, amount: int, user_id: str = Depends(get_current_user_id)):
    """Transfer tokens to another user"""
    if amount <= 0:
        raise HTTPException(status_code=400, detail="Amount must be positive")
    
    tx = await token_service.transfer(user_id, to_user_id, amount)
    if not tx:
        raise HTTPException(status_code=400, detail="Insufficient balance")
    return tx

