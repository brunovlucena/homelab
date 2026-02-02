# FutBoss AI - Team Routes
# Author: Bruno Lucena (bruno@lucena.cloud)

from fastapi import APIRouter, HTTPException, Depends, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from typing import List

from app.models.team import Team, TeamCreate, TeamResponse, Formation
from app.services.auth import AuthService
from app.services.team import TeamService

router = APIRouter()
security = HTTPBearer()
auth_service = AuthService()
team_service = TeamService()


async def get_current_user_id(credentials: HTTPAuthorizationCredentials = Depends(security)) -> str:
    user = await auth_service.get_current_user(credentials.credentials)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid token")
    return user.id


@router.post("/", response_model=TeamResponse, status_code=status.HTTP_201_CREATED)
async def create_team(team_data: TeamCreate, user_id: str = Depends(get_current_user_id)):
    """Create a new team"""
    existing = await team_service.get_user_team(user_id)
    if existing:
        raise HTTPException(status_code=400, detail="User already has a team")
    team = await team_service.create_team(team_data, user_id)
    return TeamResponse(**team.model_dump(), points=team.points(), goal_difference=team.goal_difference())


@router.get("/me", response_model=TeamResponse)
async def get_my_team(user_id: str = Depends(get_current_user_id)):
    """Get current user's team"""
    team = await team_service.get_user_team(user_id)
    if not team:
        raise HTTPException(status_code=404, detail="Team not found")
    return TeamResponse(**team.model_dump(), points=team.points(), goal_difference=team.goal_difference())


@router.get("/{team_id}", response_model=TeamResponse)
async def get_team(team_id: str):
    """Get team by ID"""
    team = await team_service.get_team(team_id)
    if not team:
        raise HTTPException(status_code=404, detail="Team not found")
    return TeamResponse(**team.model_dump(), points=team.points(), goal_difference=team.goal_difference())


@router.put("/{team_id}/formation", response_model=TeamResponse)
async def update_formation(team_id: str, formation: Formation, user_id: str = Depends(get_current_user_id)):
    """Update team formation"""
    team = await team_service.get_team(team_id)
    if not team or team.owner_id != user_id:
        raise HTTPException(status_code=403, detail="Not authorized")
    team = await team_service.update_formation(team_id, formation)
    return TeamResponse(**team.model_dump(), points=team.points(), goal_difference=team.goal_difference())


@router.get("/", response_model=List[TeamResponse])
async def list_teams(skip: int = 0, limit: int = 20):
    """List all teams"""
    teams = await team_service.list_teams(skip, limit)
    return [TeamResponse(**t.model_dump(), points=t.points(), goal_difference=t.goal_difference()) for t in teams]

