"""
Types for TRM (Tiny Recursive Model) client library.
"""
from typing import Optional, Dict, Any, List
from pydantic import BaseModel, Field
from enum import Enum


class ReflectionMode(str, Enum):
    """Reflection modes for TRM."""
    AUTO = "auto"  # Automatic reflection based on confidence
    ALWAYS = "always"  # Always reflect and refine
    NEVER = "never"  # Single pass, no reflection


class TRMRequest(BaseModel):
    """Request for TRM inference with reflection."""
    prompt: str = Field(..., min_length=1, max_length=8192)
    context: Optional[Dict[str, Any]] = Field(default_factory=dict)
    max_reflection_steps: int = Field(default=3, ge=1, le=10)
    reflection_mode: ReflectionMode = Field(default=ReflectionMode.AUTO)
    temperature: float = Field(default=0.7, ge=0.0, le=2.0)
    max_tokens: int = Field(default=2048, ge=1, le=8192)
    conversation_id: Optional[str] = None


class ReflectionStep(BaseModel):
    """A single reflection step in TRM reasoning."""
    step: int
    initial_answer: str
    reflection: Optional[str] = None
    refined_answer: Optional[str] = None
    confidence: float = Field(ge=0.0, le=1.0)
    improvement_score: Optional[float] = None


class TRMResponse(BaseModel):
    """Response from TRM inference."""
    answer: str
    reflection_steps: int
    confidence: float = Field(ge=0.0, le=1.0)
    reflection_trace: List[ReflectionStep] = Field(default_factory=list)
    duration_ms: float
    tokens_used: int
    model_name: str
    conversation_id: Optional[str] = None
