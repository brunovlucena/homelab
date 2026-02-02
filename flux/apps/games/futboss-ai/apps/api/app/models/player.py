# FutBoss AI - Player Model
# Author: Bruno Lucena (bruno@lucena.cloud)

from pydantic import BaseModel, Field
from typing import Optional, Literal
from datetime import datetime
from enum import Enum


class Position(str, Enum):
    GK = "GK"  # Goalkeeper
    CB = "CB"  # Center Back
    LB = "LB"  # Left Back
    RB = "RB"  # Right Back
    CDM = "CDM"  # Defensive Midfielder
    CM = "CM"  # Central Midfielder
    CAM = "CAM"  # Attacking Midfielder
    LW = "LW"  # Left Winger
    RW = "RW"  # Right Winger
    ST = "ST"  # Striker


class Temperament(str, Enum):
    CALM = "calm"
    EXPLOSIVE = "explosive"
    CALCULATED = "calculated"


class PlayStyle(str, Enum):
    OFFENSIVE = "offensive"
    DEFENSIVE = "defensive"
    BALANCED = "balanced"


class PlayerAttributes(BaseModel):
    """Physical, technical and mental attributes (1-100)"""
    # Physical
    speed: int = Field(ge=1, le=100, default=50)
    strength: int = Field(ge=1, le=100, default=50)
    stamina: int = Field(ge=1, le=100, default=50)

    # Technical
    finishing: int = Field(ge=1, le=100, default=50)
    passing: int = Field(ge=1, le=100, default=50)
    dribbling: int = Field(ge=1, le=100, default=50)
    defense: int = Field(ge=1, le=100, default=50)

    # Mental
    intelligence: int = Field(ge=1, le=100, default=50)
    aggression: int = Field(ge=1, le=100, default=50)
    leadership: int = Field(ge=1, le=100, default=50)
    creativity: int = Field(ge=1, le=100, default=50)

    def overall(self) -> int:
        """Calculate overall rating"""
        attrs = [
            self.speed, self.strength, self.stamina,
            self.finishing, self.passing, self.dribbling, self.defense,
            self.intelligence, self.aggression, self.leadership, self.creativity
        ]
        return round(sum(attrs) / len(attrs))


class PlayerPersonality(BaseModel):
    """Personality traits that influence AI agent behavior"""
    temperament: Temperament = Temperament.CALM
    play_style: PlayStyle = PlayStyle.BALANCED


class PlayerCreate(BaseModel):
    name: str = Field(..., min_length=2, max_length=50)
    position: Position
    nationality: str = Field(..., min_length=2, max_length=50)
    age: int = Field(ge=16, le=45)
    attributes: PlayerAttributes = Field(default_factory=PlayerAttributes)
    personality: PlayerPersonality = Field(default_factory=PlayerPersonality)


class Player(BaseModel):
    id: Optional[str] = Field(default=None, alias="_id")
    name: str
    position: Position
    nationality: str
    age: int
    attributes: PlayerAttributes
    personality: PlayerPersonality
    team_id: Optional[str] = None
    price: int = 100
    is_listed: bool = False
    created_at: datetime = Field(default_factory=datetime.utcnow)

    class Config:
        populate_by_name = True

    def get_market_value(self) -> int:
        """Calculate market value based on attributes and age"""
        base_value = self.attributes.overall() * 10
        age_modifier = 1.0
        if self.age < 23:
            age_modifier = 1.3
        elif self.age > 32:
            age_modifier = 0.7
        return int(base_value * age_modifier)

