"""
Shared types for agent-bruno chatbot.
"""
from enum import Enum
from dataclasses import dataclass, field
from typing import Optional
from datetime import datetime, timezone


class MessageRole(str, Enum):
    """Message role in conversation."""
    USER = "user"
    ASSISTANT = "assistant"
    SYSTEM = "system"


class NotificationType(str, Enum):
    """Types of notifications from other agents."""
    VULNERABILITY = "vulnerability"
    EXPLOIT = "exploit"
    ALERT = "alert"


@dataclass
class Message:
    """A single message in conversation."""
    role: MessageRole
    content: str
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()
    
    def to_dict(self) -> dict:
        return {
            "role": self.role.value,
            "content": self.content,
            "timestamp": self.timestamp,
        }
    
    @classmethod
    def from_dict(cls, data: dict) -> "Message":
        return cls(
            role=MessageRole(data.get("role", "user")),
            content=data.get("content", ""),
            timestamp=data.get("timestamp", ""),
        )


@dataclass
class Conversation:
    """A conversation with context."""
    id: str
    messages: list[Message] = field(default_factory=list)
    created_at: str = ""
    updated_at: str = ""
    
    def __post_init__(self):
        now = datetime.now(timezone.utc).isoformat()
        if not self.created_at:
            self.created_at = now
        self.updated_at = now
    
    def add_message(self, role: MessageRole, content: str) -> Message:
        """Add a message to the conversation."""
        msg = Message(role=role, content=content)
        self.messages.append(msg)
        self.updated_at = datetime.now(timezone.utc).isoformat()
        return msg
    
    def to_prompt_format(self, max_messages: int = 10) -> list[dict]:
        """Convert conversation to LLM prompt format."""
        # Get last N messages for context
        recent = self.messages[-max_messages:] if len(self.messages) > max_messages else self.messages
        return [{"role": m.role.value, "content": m.content} for m in recent]
    
    def to_dict(self) -> dict:
        return {
            "id": self.id,
            "messages": [m.to_dict() for m in self.messages],
            "created_at": self.created_at,
            "updated_at": self.updated_at,
        }


@dataclass
class ChatResponse:
    """Response from chatbot."""
    response: str
    conversation_id: str
    tokens_used: int = 0
    model: str = ""
    duration_ms: float = 0.0
    
    def to_dict(self) -> dict:
        return {
            "response": self.response,
            "conversation_id": self.conversation_id,
            "tokens_used": self.tokens_used,
            "model": self.model,
            "duration_ms": self.duration_ms,
        }
