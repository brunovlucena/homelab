"""
Ollama LLM fixtures for testing AI agent interactions.

Provides mocked Ollama API responses for testing agents
that use local LLM inference.
"""

import pytest
from unittest.mock import AsyncMock, MagicMock
from dataclasses import dataclass, field
from typing import Optional


@dataclass
class OllamaResponse:
    """Mock Ollama API response."""
    
    model: str = "llama3.2:3b"
    response: str = ""
    done: bool = True
    context: list = field(default_factory=list)
    total_duration: int = 1000000000  # nanoseconds
    load_duration: int = 100000000
    prompt_eval_count: int = 20
    prompt_eval_duration: int = 200000000
    eval_count: int = 50
    eval_duration: int = 500000000
    
    def to_dict(self) -> dict:
        """Convert to API response format."""
        return {
            "model": self.model,
            "response": self.response,
            "done": self.done,
            "context": self.context,
            "total_duration": self.total_duration,
            "load_duration": self.load_duration,
            "prompt_eval_count": self.prompt_eval_count,
            "prompt_eval_duration": self.prompt_eval_duration,
            "eval_count": self.eval_count,
            "eval_duration": self.eval_duration,
        }


@dataclass
class OllamaChatResponse:
    """Mock Ollama Chat API response."""
    
    model: str = "llama3.2:3b"
    message: dict = field(default_factory=lambda: {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
    })
    done: bool = True
    total_duration: int = 1000000000
    load_duration: int = 100000000
    prompt_eval_count: int = 20
    eval_count: int = 50
    
    def to_dict(self) -> dict:
        return {
            "model": self.model,
            "message": self.message,
            "done": self.done,
            "total_duration": self.total_duration,
            "load_duration": self.load_duration,
            "prompt_eval_count": self.prompt_eval_count,
            "eval_count": self.eval_count,
        }


@dataclass
class OllamaEmbeddingResponse:
    """Mock Ollama Embedding API response."""
    
    model: str = "nomic-embed-text"
    embedding: list = field(default_factory=lambda: [0.1] * 768)
    
    def to_dict(self) -> dict:
        return {
            "model": self.model,
            "embedding": self.embedding,
        }


class OllamaResponseFactory:
    """Factory for creating Ollama API responses."""
    
    SAMPLE_RESPONSES = {
        "greeting": "Hello! I'm an AI assistant powered by Ollama. How can I help you today?",
        "homelab_info": "The homelab is running well. All services are healthy and there are no critical alerts.",
        "security_analysis": "Based on my analysis, there are no immediate security concerns. However, I recommend regular security audits.",
        "code_review": "The code looks good overall. I have a few suggestions for improvement regarding error handling.",
        "exploit_code": '''```python
# Exploit proof-of-concept
def exploit():
    payload = "malicious_input"
    return payload
```''',
        "contract_analysis": '''I found the following vulnerabilities:
1. Reentrancy in withdraw() function
2. Integer overflow in transfer()
3. Missing access controls on sensitive functions''',
    }
    
    def __init__(self, model: str = "llama3.2:3b"):
        self.model = model
    
    def create_generate_response(
        self,
        response: Optional[str] = None,
        eval_count: int = 50,
    ) -> OllamaResponse:
        """Create a generate API response."""
        return OllamaResponse(
            model=self.model,
            response=response or self.SAMPLE_RESPONSES["greeting"],
            eval_count=eval_count,
        )
    
    def create_chat_response(
        self,
        content: Optional[str] = None,
        role: str = "assistant",
    ) -> OllamaChatResponse:
        """Create a chat API response."""
        return OllamaChatResponse(
            model=self.model,
            message={
                "role": role,
                "content": content or self.SAMPLE_RESPONSES["greeting"],
            },
        )
    
    def create_embedding_response(
        self,
        dimension: int = 768,
    ) -> OllamaEmbeddingResponse:
        """Create an embedding API response."""
        return OllamaEmbeddingResponse(
            embedding=[0.1] * dimension,
        )
    
    def create_error_response(
        self,
        error: str = "model not found",
    ) -> dict:
        """Create an error response."""
        return {"error": error}


class MockOllamaClient:
    """Mock Ollama client for testing."""
    
    def __init__(
        self,
        model: str = "llama3.2:3b",
        default_response: Optional[str] = None,
    ):
        self.model = model
        self.factory = OllamaResponseFactory(model)
        self._default_response = default_response or OllamaResponseFactory.SAMPLE_RESPONSES["greeting"]
        
        # Setup async mock methods
        self.generate = AsyncMock(
            return_value=self.factory.create_generate_response(self._default_response)
        )
        self.chat = AsyncMock(
            return_value=self.factory.create_chat_response(self._default_response)
        )
        self.embeddings = AsyncMock(
            return_value=self.factory.create_embedding_response()
        )
        self.pull = AsyncMock(return_value={"status": "success"})
        self.list = AsyncMock(return_value={"models": [{"name": model}]})
        self.show = AsyncMock(return_value={"modelfile": "", "parameters": ""})
        
        # Health check
        self.health = AsyncMock(return_value=True)
    
    def set_response(self, response: str):
        """Set the response for subsequent calls."""
        self.generate.return_value = self.factory.create_generate_response(response)
        self.chat.return_value = self.factory.create_chat_response(response)
    
    def set_error(self, error: str = "model not found"):
        """Set an error response."""
        error_response = self.factory.create_error_response(error)
        self.generate.side_effect = Exception(error)
        self.chat.side_effect = Exception(error)


@pytest.fixture
def ollama_response_factory():
    """Factory for creating Ollama responses."""
    return OllamaResponseFactory()


@pytest.fixture
def mock_ollama_response(ollama_response_factory):
    """Standard Ollama generate response."""
    return ollama_response_factory.create_generate_response().to_dict()


@pytest.fixture
def mock_ollama_chat_response(ollama_response_factory):
    """Standard Ollama chat response."""
    return ollama_response_factory.create_chat_response().to_dict()


@pytest.fixture
def mock_ollama_embedding_response(ollama_response_factory):
    """Standard Ollama embedding response."""
    return ollama_response_factory.create_embedding_response().to_dict()


@pytest.fixture
def mock_ollama_client():
    """Mock Ollama client for testing."""
    return MockOllamaClient()


# Specialized response fixtures for different agent use cases
@pytest.fixture
def mock_security_analysis_response(ollama_response_factory):
    """Security analysis response for agent-blueteam/redteam."""
    return ollama_response_factory.create_generate_response(
        response=OllamaResponseFactory.SAMPLE_RESPONSES["security_analysis"]
    ).to_dict()


@pytest.fixture
def mock_code_review_response(ollama_response_factory):
    """Code review response for agent-devsecops."""
    return ollama_response_factory.create_generate_response(
        response=OllamaResponseFactory.SAMPLE_RESPONSES["code_review"]
    ).to_dict()


@pytest.fixture
def mock_exploit_generation_response(ollama_response_factory):
    """Exploit generation response for agent-contracts."""
    return ollama_response_factory.create_generate_response(
        response=OllamaResponseFactory.SAMPLE_RESPONSES["exploit_code"]
    ).to_dict()


@pytest.fixture
def mock_contract_analysis_response(ollama_response_factory):
    """Contract analysis response for agent-contracts."""
    return ollama_response_factory.create_generate_response(
        response=OllamaResponseFactory.SAMPLE_RESPONSES["contract_analysis"]
    ).to_dict()
