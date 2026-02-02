# FutBoss AI - Team Model
# Author: Bruno Lucena (bruno@lucena.cloud)

from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime
from enum import Enum


class Formation(str, Enum):
    F442 = "4-4-2"
    F433 = "4-3-3"
    F352 = "3-5-2"
    F451 = "4-5-1"
    F343 = "3-4-3"
    F532 = "5-3-2"


class TeamCreate(BaseModel):
    name: str = Field(..., min_length=3, max_length=50)
    formation: Formation = Formation.F442


class Team(BaseModel):
    id: Optional[str] = Field(default=None, alias="_id")
    name: str
    owner_id: str
    formation: Formation = Formation.F442
    player_ids: List[str] = Field(default_factory=list)
    wins: int = 0
    draws: int = 0
    losses: int = 0
    goals_for: int = 0
    goals_against: int = 0
    created_at: datetime = Field(default_factory=datetime.utcnow)

    class Config:
        populate_by_name = True

    def points(self) -> int:
        return (self.wins * 3) + self.draws

    def matches_played(self) -> int:
        return self.wins + self.draws + self.losses

    def goal_difference(self) -> int:
        return self.goals_for - self.goals_against


class TeamResponse(BaseModel):
    id: str
    name: str
    owner_id: str
    formation: Formation
    player_ids: List[str]
    wins: int
    draws: int
    losses: int
    points: int
    goal_difference: int

    class Config:
        from_attributes = True

