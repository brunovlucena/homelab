# FutBoss AI - Match Routes
# Author: Bruno Lucena (bruno@lucena.cloud)

from fastapi import APIRouter, HTTPException, Depends, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from typing import List, Optional

from app.models.match import Match, MatchCreate, MatchStatus
from app.services.auth import AuthService
from app.services.match import MatchService
from app.services.team import TeamService

router = APIRouter()
security = HTTPBearer()
auth_service = AuthService()
match_service = MatchService()
team_service = TeamService()


async def get_current_user_id(credentials: HTTPAuthorizationCredentials = Depends(security)) -> str:
    user = await auth_service.get_current_user(credentials.credentials)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid token")
    return user.id


@router.post("/", response_model=Match, status_code=status.HTTP_201_CREATED)
async def create_match(match_data: MatchCreate, user_id: str = Depends(get_current_user_id)):
    """Create a new match"""
    team = await team_service.get_user_team(user_id)
    if not team:
        raise HTTPException(status_code=400, detail="Create a team first")
    
    if team.id != match_data.home_team_id and team.id != match_data.away_team_id:
        raise HTTPException(status_code=403, detail="You must be part of the match")
    
    match = await match_service.create_match(match_data)
    return match


@router.get("/{match_id}", response_model=Match)
async def get_match(match_id: str):
    """Get match by ID"""
    match = await match_service.get_match(match_id)
    if not match:
        raise HTTPException(status_code=404, detail="Match not found")
    return match


@router.post("/{match_id}/start", response_model=Match)
async def start_match(match_id: str, user_id: str = Depends(get_current_user_id)):
    """Start a match"""
    match = await match_service.start_match(match_id)
    if not match:
        raise HTTPException(status_code=400, detail="Cannot start match")
    return match


@router.get("/team/{team_id}", response_model=List[Match])
async def get_team_matches(team_id: str, status: Optional[MatchStatus] = None, skip: int = 0, limit: int = 20):
    """Get matches for a team"""
    return await match_service.get_team_matches(team_id, status, skip, limit)


@router.get("/", response_model=List[Match])
async def list_matches(status: Optional[MatchStatus] = None, skip: int = 0, limit: int = 20):
    """List all matches"""
    return await match_service.list_matches(status, skip, limit)

