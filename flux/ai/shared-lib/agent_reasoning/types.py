"""
Types for Agent-Reasoning client library.
"""
from typing import Optional, Dict, Any, List
from pydantic import BaseModel, Field
from enum import Enum


class TaskType(str, Enum):
    """Types of reasoning tasks."""
    PLANNING = "planning"
    TROUBLESHOOTING = "troubleshooting"
    OPTIMIZATION = "optimization"
    LOGIC = "logic"
    GENERAL = "general"


class ReasoningRequest(BaseModel):
    """Request for reasoning task."""
    question: str = Field(..., min_length=1, max_length=4096)
    context: Optional[Dict[str, Any]] = Field(default_factory=dict)
    max_steps: int = Field(default=6, ge=1, le=20)
    task_type: TaskType = Field(default=TaskType.GENERAL)
    conversation_id: Optional[str] = None


class ReasoningStep(BaseModel):
    """A single step in the reasoning process."""
    step: int
    latent_state: Optional[Dict[str, Any]] = None
    intermediate_answer: Optional[str] = None
    confidence: float = Field(ge=0.0, le=1.0)


class ReasoningResponse(BaseModel):
    """Response from reasoning task."""
    answer: str
    steps: int
    confidence: float = Field(ge=0.0, le=1.0)
    reasoning_trace: List[ReasoningStep] = Field(default_factory=list)
    duration_ms: float
    task_type: TaskType
    conversation_id: Optional[str] = None

