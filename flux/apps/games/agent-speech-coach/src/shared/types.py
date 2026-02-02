"""Type definitions for speech coach agent"""
from enum import Enum
from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field
from datetime import datetime


class ExerciseType(str, Enum):
    """Types of speech exercises"""
    WORD_REPETITION = "word_repetition"
    PHRASE_COMPLETION = "phrase_completion"
    STORY_TELLING = "story_telling"
    CONVERSATION = "conversation"
    IMITATION = "imitation"
    QUESTION_ANSWER = "question_answer"


class DifficultyLevel(str, Enum):
    """Difficulty levels for exercises"""
    BEGINNER = "beginner"
    INTERMEDIATE = "intermediate"
    ADVANCED = "advanced"


class GameStatus(str, Enum):
    """Status of a game session"""
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    ABANDONED = "abandoned"


class User(BaseModel):
    """User model"""
    id: str
    name: str
    age: Optional[int] = None
    preferences: Dict[str, Any] = Field(default_factory=dict)


class Exercise(BaseModel):
    """Exercise model"""
    id: str
    type: ExerciseType
    title: str
    description: str
    difficulty: DifficultyLevel
    instructions: str
    target_words: List[str] = Field(default_factory=list)
    expected_duration_minutes: int = 5
    points: int = 10


class GameSession(BaseModel):
    """Game session model"""
    id: str
    user_id: str
    exercise_id: str
    status: GameStatus
    started_at: datetime
    completed_at: Optional[datetime] = None
    score: Optional[int] = None
    attempts: int = 0
    feedback: Optional[str] = None
    metadata: Dict[str, Any] = Field(default_factory=dict)


class Progress(BaseModel):
    """Progress tracking model"""
    user_id: str
    total_sessions: int = 0
    completed_exercises: int = 0
    total_points: int = 0
    current_streak: int = 0
    longest_streak: int = 0
    achievements: List[str] = Field(default_factory=list)
    last_activity: Optional[datetime] = None
    exercises_completed: Dict[str, int] = Field(default_factory=dict)  # exercise_type -> count


class SpeechRequest(BaseModel):
    """Speech coaching request"""
    user_id: str
    query: str
    exercise_type: Optional[ExerciseType] = None
    session_id: Optional[str] = None
    audio_data: Optional[str] = None  # Base64 encoded audio
    face_data: Optional[Dict[str, Any]] = None  # Face recognition data
    metadata: Dict[str, Any] = Field(default_factory=dict)


class SpeechResponse(BaseModel):
    """Speech coaching response"""
    response: str
    exercise: Optional[Exercise] = None
    session: Optional[GameSession] = None
    progress: Optional[Progress] = None
    suggestions: List[str] = Field(default_factory=list)
    encouragement: Optional[str] = None
    model: str = ""
    tokens_used: int = 0
    duration_ms: float = 0.0
