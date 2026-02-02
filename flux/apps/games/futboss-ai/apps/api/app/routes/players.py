# FutBoss AI - Player Routes
# Author: Bruno Lucena (bruno@lucena.cloud)

from fastapi import APIRouter, HTTPException, Depends, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from typing import List, Optional

from app.models.player import Player, PlayerCreate, Position
from app.services.auth import AuthService
from app.services.player import PlayerService
from app.services.team import TeamService

router = APIRouter()
security = HTTPBearer()
auth_service = AuthService()
player_service = PlayerService()
team_service = TeamService()


async def get_current_user_id(credentials: HTTPAuthorizationCredentials = Depends(security)) -> str:
    user = await auth_service.get_current_user(credentials.credentials)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid token")
    return user.id


@router.get("/market", response_model=List[Player])
async def get_market_players(
    position: Optional[Position] = None,
    min_overall: int = 0,
    max_price: Optional[int] = None,
    skip: int = 0,
    limit: int = 20
):
    """Get players available in the market"""
    return await player_service.get_market_players(position, min_overall, max_price, skip, limit)


@router.get("/team/{team_id}", response_model=List[Player])
async def get_team_players(team_id: str):
    """Get all players in a team"""
    return await player_service.get_team_players(team_id)


@router.get("/{player_id}", response_model=Player)
async def get_player(player_id: str):
    """Get player by ID"""
    player = await player_service.get_player(player_id)
    if not player:
        raise HTTPException(status_code=404, detail="Player not found")
    return player


@router.post("/{player_id}/buy", response_model=Player)
async def buy_player(player_id: str, user_id: str = Depends(get_current_user_id)):
    """Buy a player from the market"""
    team = await team_service.get_user_team(user_id)
    if not team:
        raise HTTPException(status_code=400, detail="Create a team first")
    
    player = await player_service.buy_player(player_id, team.id, user_id)
    if not player:
        raise HTTPException(status_code=400, detail="Cannot buy player - insufficient funds or player not available")
    return player


@router.post("/{player_id}/sell", response_model=Player)
async def list_player_for_sale(player_id: str, price: int, user_id: str = Depends(get_current_user_id)):
    """List a player for sale in the market"""
    team = await team_service.get_user_team(user_id)
    if not team:
        raise HTTPException(status_code=400, detail="No team found")
    
    player = await player_service.list_for_sale(player_id, team.id, price)
    if not player:
        raise HTTPException(status_code=400, detail="Cannot list player")
    return player


@router.post("/generate", response_model=Player, status_code=status.HTTP_201_CREATED)
async def generate_player(position: Optional[Position] = None):
    """Generate a random player (for market/testing)"""
    return await player_service.generate_random_player(position)

