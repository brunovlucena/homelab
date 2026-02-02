"""Unit tests for chatbot handler."""
import pytest
from unittest.mock import patch, AsyncMock, MagicMock

from shared.types import MessageRole, Conversation, Message


class TestConversation:
    """Tests for Conversation class."""
    
    def test_create_conversation(self):
        """Test creating a new conversation."""
        conv = Conversation(id="test-123")
        assert conv.id == "test-123"
        assert len(conv.messages) == 0
        assert conv.created_at is not None
    
    def test_add_message(self):
        """Test adding messages to conversation."""
        conv = Conversation(id="test-123")
        
        msg = conv.add_message(MessageRole.USER, "Hello!")
        
        assert len(conv.messages) == 1
        assert msg.role == MessageRole.USER
        assert msg.content == "Hello!"
    
    def test_to_prompt_format(self):
        """Test converting conversation to prompt format."""
        conv = Conversation(id="test-123")
        conv.add_message(MessageRole.USER, "Hello!")
        conv.add_message(MessageRole.ASSISTANT, "Hi there!")
        
        prompt = conv.to_prompt_format()
        
        assert len(prompt) == 2
        assert prompt[0]["role"] == "user"
        assert prompt[1]["role"] == "assistant"
    
    def test_to_prompt_format_max_messages(self):
        """Test prompt format respects max messages."""
        conv = Conversation(id="test-123")
        for i in range(20):
            conv.add_message(MessageRole.USER, f"Message {i}")
        
        prompt = conv.to_prompt_format(max_messages=5)
        
        assert len(prompt) == 5
        # Should be the last 5 messages
        assert "Message 15" in prompt[0]["content"]


class TestMessage:
    """Tests for Message class."""
    
    def test_create_message(self):
        """Test creating a message."""
        msg = Message(role=MessageRole.USER, content="Hello!")
        
        assert msg.role == MessageRole.USER
        assert msg.content == "Hello!"
        assert msg.timestamp is not None
    
    def test_message_to_dict(self):
        """Test message serialization."""
        msg = Message(role=MessageRole.ASSISTANT, content="Hi!")
        
        data = msg.to_dict()
        
        assert data["role"] == "assistant"
        assert data["content"] == "Hi!"
    
    def test_message_from_dict(self):
        """Test message deserialization."""
        data = {"role": "user", "content": "Hello!"}
        
        msg = Message.from_dict(data)
        
        assert msg.role == MessageRole.USER
        assert msg.content == "Hello!"


@pytest.mark.asyncio
class TestChatBot:
    """Tests for ChatBot class."""
    
    async def test_chat_success(self, mock_httpx_client, mock_ollama_response):
        """Test successful chat interaction."""
        from chatbot.handler import ChatBot
        
        with patch('httpx.AsyncClient', return_value=mock_httpx_client):
            bot = ChatBot(
                ollama_url="http://test:11434",
                model="test-model"
            )
            
            response = await bot.chat("Hello!")
            
            assert response.response == mock_ollama_response["response"]
            assert response.conversation_id is not None
            # Total tokens = input (prompt_eval_count) + output (eval_count)
            expected_tokens = mock_ollama_response["prompt_eval_count"] + mock_ollama_response["eval_count"]
            assert response.tokens_used == expected_tokens
    
    async def test_chat_maintains_conversation(self, mock_httpx_client):
        """Test that conversation context is maintained."""
        from chatbot.handler import ChatBot
        
        with patch('httpx.AsyncClient', return_value=mock_httpx_client):
            bot = ChatBot(
                ollama_url="http://test:11434",
                model="test-model"
            )
            
            response1 = await bot.chat("Hello!")
            conv_id = response1.conversation_id
            
            response2 = await bot.chat("How are you?", conversation_id=conv_id)
            
            assert response2.conversation_id == conv_id
            
            # Check conversation has both messages
            conv = bot.conversations.get_or_create(conv_id)
            assert len(conv.messages) >= 2
    
    async def test_health_check(self, mock_httpx_client):
        """Test health check endpoint."""
        from chatbot.handler import ChatBot
        
        with patch('httpx.AsyncClient', return_value=mock_httpx_client):
            bot = ChatBot(ollama_url="http://test:11434")
            
            result = await bot.health_check()
            
            assert result is True
