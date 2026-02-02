# FutBoss AI - Match Model
# Author: Bruno Lucena (bruno@lucena.cloud)

from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime
from enum import Enum


class MatchStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    FINISHED = "finished"
    CANCELLED = "cancelled"


class EventType(str, Enum):
    GOAL = "goal"
    ASSIST = "assist"
    YELLOW_CARD = "yellow_card"
    RED_CARD = "red_card"
    SUBSTITUTION = "substitution"
    INJURY = "injury"
    PENALTY = "penalty"
    SAVE = "save"
    FOUL = "foul"


class MatchEvent(BaseModel):
    minute: int = Field(ge=0, le=120)
    event_type: EventType
    player_id: str
    team_id: str
    description: str
    ai_narration: Optional[str] = None


class MatchState(BaseModel):
    """Current state of a match for real-time updates"""
    minute: int = 0
    home_score: int = 0
    away_score: int = 0
    possession_home: int = 50
    possession_away: int = 50
    shots_home: int = 0
    shots_away: int = 0
    is_paused: bool = False


class MatchCreate(BaseModel):
    home_team_id: str
    away_team_id: str


class Match(BaseModel):
    id: Optional[str] = Field(default=None, alias="_id")
    home_team_id: str
    away_team_id: str
    home_score: int = 0
    away_score: int = 0
    status: MatchStatus = MatchStatus.PENDING
    events: List[MatchEvent] = Field(default_factory=list)
    state: MatchState = Field(default_factory=MatchState)
    started_at: Optional[datetime] = None
    finished_at: Optional[datetime] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)

    class Config:
        populate_by_name = True

    def is_home_winner(self) -> Optional[bool]:
        if self.status != MatchStatus.FINISHED:
            return None
        if self.home_score > self.away_score:
            return True
        elif self.home_score < self.away_score:
            return False
        return None  # Draw

