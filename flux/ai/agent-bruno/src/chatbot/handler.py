"""
Chatbot handler with TRM (Tiny Recursive Model) integration, CloudEvents support, and Domain Memory.

Uses TRM for built-in reflection and self-refinement capabilities.
Following Nate B. Jones's Domain Memory Factory pattern:
- Persistent, structured representation of work
- Multi-tiered memory (short-term, working, user, long-term)
- Agents that remember and learn across sessions

Emits CloudEvents for:
- Chat messages (for analytics)
- Security-related questions (triggers cross-agent awareness)
- Status inquiries

Receives CloudEvents from:
- agent-contracts (vulnerability findings, exploit validations)
- alertmanager (system alerts)
"""
import os
import time
from typing import Optional
from uuid import uuid4

import httpx
import structlog

from shared.types import (
    Message, 
    MessageRole, 
    Conversation, 
    ChatResponse,
)
from shared.metrics import (
    MESSAGES_PROCESSED,
    CONVERSATIONS_ACTIVE,
    CONVERSATION_LENGTH,
    RESPONSE_DURATION,
    LLM_INFERENCE_DURATION,
    TOKENS_USED,
    API_CALLS,
    MEMORY_CONTEXT_BUILD_DURATION,
    MEMORY_CONTEXT_SIZE,
    USER_FACTS_RECORDED,
    USER_PREFERENCES_UPDATED,
    LEARNINGS_RECORDED,
    EVENT_PROCESSING_DURATION,
    record_memory_operation,
    record_event_published,
    record_event_received,
    set_memory_store_connected,
    set_ollama_available,
)

# Import TRM client
try:
    from agent_trm import TRMClient, TRMRequest, ReflectionMode
    TRM_AVAILABLE = True
except ImportError:
    TRM_AVAILABLE = False
    logger.warning("TRM client not available, falling back to Ollama")
from shared.events import (
    EventPublisher,
    EventSubscriber,
    EventType,
    ChatMessageEvent,
    ChatIntentEvent,
    detect_intent,
)

# Import Domain Memory components
try:
    from agent_memory import (
        DomainMemoryManager,
        ConversationMemory,
        UserMemory,
        ChatAgentSchema,
    )
    MEMORY_AVAILABLE = True
except ImportError:
    MEMORY_AVAILABLE = False

logger = structlog.get_logger()

# Default system prompt
DEFAULT_SYSTEM_PROMPT = """You are Agent-Bruno, a helpful and friendly AI assistant for a homelab infrastructure.

Your responsibilities:
- Answer questions about the homelab services and infrastructure
- Help users understand how to use various services
- Provide technical guidance when asked
- Be conversational and approachable

Important guidelines:
- Be concise but helpful
- If you don't know something, say so honestly
- Focus on being practical and actionable
- Use markdown formatting when helpful

Available homelab services include:
- Homepage (main dashboard)
- Prometheus & Grafana (monitoring)
- Various AI agents (agent-contracts, etc.)
- Kubernetes infrastructure
- Storage services (MinIO)
- Message queues (RabbitMQ)

Keep responses focused and helpful. You're here to assist!"""


class ConversationManager:
    """
    Simple in-memory conversation manager.
    
    DEPRECATED: Use DomainMemoryManager for persistent memory.
    Kept for backward compatibility when memory is disabled.
    """
    
    def __init__(self, max_conversations: int = 1000):
        self._conversations: dict[str, Conversation] = {}
        self._max_conversations = max_conversations
    
    def get_or_create(self, conversation_id: Optional[str] = None) -> Conversation:
        """Get existing conversation or create new one."""
        if conversation_id and conversation_id in self._conversations:
            return self._conversations[conversation_id]
        
        # Create new conversation
        new_id = conversation_id or str(uuid4())
        conv = Conversation(id=new_id)
        
        # Cleanup old conversations if limit reached
        if len(self._conversations) >= self._max_conversations:
            self._cleanup_oldest()
        
        self._conversations[new_id] = conv
        CONVERSATIONS_ACTIVE.set(len(self._conversations))
        return conv
    
    def _cleanup_oldest(self):
        """Remove oldest conversations."""
        if not self._conversations:
            return
        
        # Sort by updated_at and remove oldest 10%
        sorted_convs = sorted(
            self._conversations.items(),
            key=lambda x: x[1].updated_at
        )
        
        to_remove = max(1, len(sorted_convs) // 10)
        for conv_id, _ in sorted_convs[:to_remove]:
            del self._conversations[conv_id]
        
        logger.info("conversations_cleaned", removed=to_remove)


class ChatBot:
    """
    Main chatbot handler with TRM (Tiny Recursive Model) integration, CloudEvents, and Domain Memory.
    
    Uses TRM for built-in reflection and self-refinement capabilities.
    Now supports persistent, stateful memory following Nate B. Jones's patterns:
    - Conversation memory: Remembers chat context across sessions
    - User memory: Stores user preferences and learned facts
    - Working memory: Tracks goals, requirements, and task progress
    - Long-term memory: Accumulates learnings and patterns
    """
    
    def __init__(
        self,
        ollama_url: str = None,  # Deprecated: kept for backward compatibility
        model: str = None,  # Deprecated: kept for backward compatibility
        system_prompt: str = None,
        max_context_messages: int = 10,
        event_publisher: Optional[EventPublisher] = None,
        event_subscriber: Optional[EventSubscriber] = None,
        # Domain Memory configuration
        memory_enabled: bool = None,
        redis_url: str = None,
        postgres_url: str = None,
        # TRM configuration
        trm_model_name: str = None,
        trm_use_hf_api: bool = None,
    ):
        # TRM configuration (primary)
        self.trm_model_name = trm_model_name or os.getenv(
            "TRM_MODEL_NAME",
            "ainz/tiny-recursive-model"
        )
        self.trm_use_hf_api = trm_use_hf_api
        if self.trm_use_hf_api is None:
            self.trm_use_hf_api = os.getenv("TRM_USE_HF_API", "false").lower() == "true"
        
        # Legacy Ollama config (fallback only)
        self.ollama_url = ollama_url or os.getenv(
            "OLLAMA_URL", 
            "http://ollama.ai-inference.svc.cluster.local:11434"
        )
        self.model = model or os.getenv("OLLAMA_MODEL", "llama3.2:3b")
        
        self.system_prompt = system_prompt or os.getenv("SYSTEM_PROMPT", DEFAULT_SYSTEM_PROMPT)
        self.max_context_messages = max_context_messages
        
        # CloudEvents integration
        self.event_publisher = event_publisher
        self.event_subscriber = event_subscriber
        self._emit_events = os.getenv("EMIT_EVENTS", "true").lower() == "true"
        
        # Domain Memory configuration
        self._memory_enabled = memory_enabled
        if self._memory_enabled is None:
            self._memory_enabled = os.getenv("MEMORY_ENABLED", "true").lower() == "true"
        
        self._redis_url = redis_url or os.getenv("REDIS_URL")
        self._postgres_url = postgres_url or os.getenv("POSTGRES_URL")
        
        # Initialize memory manager or fallback to legacy
        self.memory_manager: Optional[DomainMemoryManager] = None
        self.conversations = ConversationManager()  # Fallback
        
        # Initialize TRM client
        self.trm_client: Optional[TRMClient] = None
        if TRM_AVAILABLE:
            try:
                self.trm_client = TRMClient(
                    model_name=self.trm_model_name,
                    use_hf_api=self.trm_use_hf_api,
                )
                logger.info("trm_client_initialized", model_name=self.trm_model_name)
            except Exception as e:
                logger.error("trm_client_init_failed", error=str(e))
                self.trm_client = None
        
        logger.info(
            "chatbot_initialized",
            trm_enabled=TRM_AVAILABLE and self.trm_client is not None,
            trm_model=self.trm_model_name,
            ollama_fallback=self.ollama_url,
            model=self.model,
            events_enabled=self._emit_events,
            memory_enabled=self._memory_enabled and MEMORY_AVAILABLE,
            memory_available=MEMORY_AVAILABLE,
        )
    
    async def initialize_memory(self):
        """Initialize domain memory manager."""
        if not self._memory_enabled or not MEMORY_AVAILABLE:
            logger.info("domain_memory_disabled")
            return
        
        try:
            self.memory_manager = DomainMemoryManager(
                agent_id="agent-bruno",
                agent_type="chat",
                domain="conversation",
                redis_url=self._redis_url,
                postgres_url=self._postgres_url,
                use_redis=bool(self._redis_url),
                use_postgres=bool(self._postgres_url),
                default_constraints=[
                    {
                        "description": "Maintain helpful, accurate, and respectful communication",
                        "hard": True,
                        "category": "behavior",
                    },
                    {
                        "description": "Protect user privacy - don't share sensitive information",
                        "hard": True,
                        "category": "privacy",
                    },
                ],
            )
            await self.memory_manager.connect()
            logger.info(
                "domain_memory_initialized",
                redis_enabled=bool(self._redis_url),
                postgres_enabled=bool(self._postgres_url),
            )
        except Exception as e:
            logger.error("domain_memory_init_failed", error=str(e))
            self.memory_manager = None
    
    async def shutdown_memory(self):
        """Shutdown domain memory manager."""
        if self.memory_manager:
            try:
                await self.memory_manager.disconnect()
                logger.info("domain_memory_shutdown")
            except Exception as e:
                logger.error("domain_memory_shutdown_failed", error=str(e))
    
    async def chat(
        self,
        message: str,
        conversation_id: Optional[str] = None,
        user_id: Optional[str] = None,
    ) -> ChatResponse:
        """
        Process a chat message and return response.
        
        Uses Domain Memory when available for:
        - Persistent conversation context
        - User preferences and history
        - Learning and improvement over time
        
        Args:
            message: User's message
            conversation_id: Optional conversation ID for context
            user_id: Optional user ID for personalization
            
        Returns:
            ChatResponse with the assistant's reply
        """
        start_time = time.time()
        log = logger.bind(conversation_id=conversation_id, user_id=user_id)
        
        try:
            # Use domain memory if available
            if self.memory_manager:
                return await self._chat_with_memory(
                    message, conversation_id, user_id, start_time, log
                )
            else:
                return await self._chat_legacy(
                    message, conversation_id, start_time, log
                )
            
        except Exception as e:
            MESSAGES_PROCESSED.labels(status="error").inc()
            log.error("chat_failed", error=str(e))
            
            # Return error response
            return ChatResponse(
                response="I apologize, but I encountered an error processing your message. Please try again.",
                conversation_id=conversation_id or str(uuid4()),
                tokens_used=0,
                model=self.model,
                duration_ms=(time.time() - start_time) * 1000,
            )
    
    async def _chat_with_memory(
        self,
        message: str,
        conversation_id: Optional[str],
        user_id: Optional[str],
        start_time: float,
        log,
    ) -> ChatResponse:
        """
        Chat with Domain Memory enabled.
        
        This provides:
        - Persistent conversation history
        - User-specific context and preferences
        - Learning from interactions
        - Full observability with metrics and tracing
        """
        # Get or create conversation memory
        record_memory_operation("read", "conversation")
        conv_memory = await self.memory_manager.start_conversation(
            user_id=user_id,
            conversation_id=conversation_id,
            initial_message=message if not conversation_id else None,
        )
        
        # If this is a continuing conversation, add the new message
        if conversation_id:
            record_memory_operation("write", "conversation")
            await self.memory_manager.add_message(conv_memory, "user", message)
        
        # Record conversation length
        CONVERSATION_LENGTH.observe(conv_memory.message_count)
        
        log.info(
            "message_received",
            message_length=len(message),
            memory_enabled=True,
            message_count=conv_memory.message_count,
        )
        
        # Detect intent for event emission
        intent, matched_keywords = detect_intent(message)
        
        # Build comprehensive context from memory with timing
        context_start = time.time()
        context = await self.memory_manager.build_context(
            user_id=user_id,
            conversation_id=conv_memory.conversation_id,
            include_user_memory=True,
            include_domain_knowledge=True,
            conversation_limit=self.max_context_messages,
        )
        context_duration = time.time() - context_start
        
        # Build prompt with memory context
        enhanced_prompt = self._build_prompt_with_context(message, context)
        
        # Record context metrics
        MEMORY_CONTEXT_BUILD_DURATION.observe(context_duration)
        MEMORY_CONTEXT_SIZE.observe(len(enhanced_prompt))
        
        # Inject recent notifications if relevant
        if intent == EventType.CHAT_INTENT_SECURITY and self.event_subscriber:
            notifications = self.event_subscriber.get_recent_notifications(limit=3)
            if notifications:
                enhanced_prompt = self._inject_notifications(enhanced_prompt, notifications)
        
        # Query TRM (or Ollama fallback) with timing
        with LLM_INFERENCE_DURATION.labels(model=self.model).time():
            if self.trm_client:
                response_text, tokens, input_tokens, output_tokens = await self._query_trm(enhanced_prompt, conv_memory.conversation_id)
            else:
                response_text, tokens, input_tokens, output_tokens = await self._query_ollama(enhanced_prompt)
        
        if not response_text:
            raise ValueError("Empty response from LLM")
        
        # Add assistant response to conversation memory
        record_memory_operation("write", "conversation")
        await self.memory_manager.add_message(conv_memory, "assistant", response_text)
        
        # Update user interaction stats
        if user_id:
            record_memory_operation("write", "user")
            await self.memory_manager.record_user_interaction(user_id, "chat")
        
        # Calculate duration
        duration_ms = (time.time() - start_time) * 1000
        
        # Record metrics
        MESSAGES_PROCESSED.labels(status="success").inc()
        RESPONSE_DURATION.labels(model=self.model).observe(duration_ms / 1000)
        self._record_token_metrics(tokens, input_tokens, output_tokens)
        
        log.info(
            "response_generated",
            tokens=tokens,
            duration_ms=duration_ms,
            memory_enabled=True,
        )
        
        # Emit CloudEvents
        if self._emit_events and self.event_publisher:
            await self._emit_chat_events(
                conv_memory.conversation_id, message, response_text, 
                tokens, duration_ms, intent, matched_keywords
            )
        
        # Record learnings from the interaction (async, non-blocking)
        await self._maybe_record_learning(message, response_text, intent)
        
        return ChatResponse(
            response=response_text,
            conversation_id=conv_memory.conversation_id,
            tokens_used=tokens,
            model=self.model,
            duration_ms=duration_ms,
        )
    
    async def _chat_legacy(
        self,
        message: str,
        conversation_id: Optional[str],
        start_time: float,
        log,
    ) -> ChatResponse:
        """
        Legacy chat without Domain Memory (fallback).
        
        Uses in-memory conversation manager.
        """
        # Get or create conversation
        conv = self.conversations.get_or_create(conversation_id)
        
        # Add user message
        conv.add_message(MessageRole.USER, message)
        log.info("message_received", message_length=len(message), memory_enabled=False)
        
        # Detect intent for event emission
        intent, matched_keywords = detect_intent(message)
        
        # Inject recent notifications into context if relevant
        enhanced_prompt = self._build_prompt(conv)
        if intent == EventType.CHAT_INTENT_SECURITY and self.event_subscriber:
            notifications = self.event_subscriber.get_recent_notifications(limit=3)
            if notifications:
                enhanced_prompt = self._inject_notifications(enhanced_prompt, notifications)
        
        # Query TRM (or Ollama fallback)
        with LLM_INFERENCE_DURATION.labels(model=self.model).time():
            if self.trm_client:
                response_text, tokens, input_tokens, output_tokens = await self._query_trm(enhanced_prompt, conv.id)
            else:
                response_text, tokens, input_tokens, output_tokens = await self._query_ollama(enhanced_prompt)
        
        if not response_text:
            raise ValueError("Empty response from LLM")
        
        # Add assistant response to conversation
        conv.add_message(MessageRole.ASSISTANT, response_text)
        
        # Calculate duration
        duration_ms = (time.time() - start_time) * 1000
        
        # Record metrics
        MESSAGES_PROCESSED.labels(status="success").inc()
        RESPONSE_DURATION.labels(model=self.model).observe(duration_ms / 1000)
        self._record_token_metrics(tokens, input_tokens, output_tokens)
        
        log.info(
            "response_generated",
            tokens=tokens,
            duration_ms=duration_ms,
            memory_enabled=False,
        )
        
        # Emit CloudEvents
        if self._emit_events and self.event_publisher:
            await self._emit_chat_events(
                conv.id, message, response_text, tokens, duration_ms, intent, matched_keywords
            )
        
        return ChatResponse(
            response=response_text,
            conversation_id=conv.id,
            tokens_used=tokens,
            model=self.model,
            duration_ms=duration_ms,
        )
    
    def _build_prompt_with_context(self, message: str, context: dict) -> str:
        """Build prompt with domain memory context."""
        prompt_parts = [f"System: {self.system_prompt}\n"]
        
        # Add user context if available
        if "user" in context:
            user = context["user"]
            user_context = []
            if user.get("preferences"):
                user_context.append(f"User preferences: {user['preferences']}")
            if user.get("facts"):
                user_context.append(f"Known about user: {', '.join(user['facts'])}")
            if user.get("custom_instructions"):
                user_context.append(f"User instructions: {user['custom_instructions']}")
            if user_context:
                prompt_parts.append("[USER CONTEXT]\n" + "\n".join(user_context) + "\n[END USER CONTEXT]\n")
        
        # Add conversation history
        if "conversation" in context and context["conversation"].get("messages"):
            prompt_parts.append("[CONVERSATION HISTORY]")
            for msg in context["conversation"]["messages"]:
                role = msg["role"].capitalize()
                content = msg["content"]
                prompt_parts.append(f"{role}: {content}")
            prompt_parts.append("[END HISTORY]\n")
        
        # Add current message
        prompt_parts.append(f"User: {message}")
        prompt_parts.append("Assistant:")
        
        return "\n\n".join(prompt_parts)
    
    def _build_prompt(self, conv: Conversation) -> str:
        """Build prompt with system message and conversation context."""
        # Get recent messages for context
        context_messages = conv.to_prompt_format(self.max_context_messages)
        
        # Build full prompt
        prompt_parts = [f"System: {self.system_prompt}\n"]
        
        for msg in context_messages:
            role = msg["role"].capitalize()
            content = msg["content"]
            prompt_parts.append(f"{role}: {content}")
        
        prompt_parts.append("Assistant:")
        
        return "\n\n".join(prompt_parts)
    
    def _inject_notifications(self, prompt: str, notifications: list[dict]) -> str:
        """Inject recent security notifications into the prompt context."""
        if not notifications:
            return prompt
        
        notification_text = "\n\n[RECENT SECURITY NOTIFICATIONS - Reference if relevant to user's question]\n"
        for notif in notifications:
            notification_text += f"- {notif.get('message', 'Unknown notification')}\n"
        notification_text += "[END NOTIFICATIONS]\n"
        
        # Insert after system prompt
        parts = prompt.split("\n\n", 1)
        if len(parts) == 2:
            return parts[0] + notification_text + "\n\n" + parts[1]
        return notification_text + prompt
    
    async def _emit_chat_events(
        self,
        conversation_id: str,
        message: str,
        response: str,
        tokens: int,
        duration_ms: float,
        intent: Optional[EventType],
        keywords: list[str],
    ):
        """Emit CloudEvents for the chat interaction (fire-and-forget)."""
        try:
            # Always emit chat message event for analytics
            await self.event_publisher.emit_chat_message(
                ChatMessageEvent(
                    conversation_id=conversation_id,
                    message_length=len(message),
                    response_length=len(response),
                    tokens_used=tokens,
                    model=self.model,
                    duration_ms=duration_ms,
                )
            )
            
            # Emit intent-specific events
            if intent and keywords:
                await self.event_publisher.emit_intent(
                    intent,
                    ChatIntentEvent(
                        conversation_id=conversation_id,
                        intent=intent.value,
                        query=message[:200],  # Truncate for privacy
                        keywords_matched=keywords,
                    )
                )
                logger.debug(
                    "intent_event_emitted",
                    intent=intent.value,
                    keywords=keywords,
                )
        except Exception as e:
            # Don't fail the chat if event emission fails
            logger.warning("event_emission_failed", error=str(e))
    
    async def _maybe_record_learning(
        self,
        message: str,
        response: str,
        intent: Optional[EventType],
    ):
        """
        Record learnings from the interaction to long-term memory.
        
        This enables the agent to improve over time.
        """
        if not self.memory_manager:
            return
        
        try:
            # Record topic patterns
            if intent:
                await self.memory_manager.record_learning(
                    domain="conversation",
                    content=f"User asked about {intent.value}: {message[:100]}",
                    source="chat_interaction",
                    category="topic_pattern",
                    confidence=0.7,
                )
        except Exception as e:
            # Non-critical, don't fail
            logger.debug("learning_record_failed", error=str(e))
    
    async def _query_trm(self, prompt: str, conversation_id: Optional[str] = None) -> tuple[str, int, int, int]:
        """
        Query TRM (Tiny Recursive Model) with built-in reflection.
        
        Returns:
            Tuple of (response_text, total_tokens, input_tokens, output_tokens)
        """
        if not self.trm_client:
            raise RuntimeError("TRM client not initialized")
        
        try:
            request = TRMRequest(
                prompt=prompt,
                max_reflection_steps=3,
                reflection_mode=ReflectionMode.AUTO,
                max_tokens=2048,
                conversation_id=conversation_id,
            )
            
            response = await self.trm_client.generate(request)
            
            API_CALLS.labels(service="trm", status="success").inc()
            
            # Estimate input/output tokens (TRM doesn't provide exact counts)
            # Rough approximation: prompt length vs answer length
            input_tokens = len(prompt.split())
            output_tokens = len(response.answer.split())
            total_tokens = response.tokens_used
            
            logger.info(
                "trm_generation_completed",
                reflection_steps=response.reflection_steps,
                confidence=response.confidence,
                duration_ms=response.duration_ms,
            )
            
            return (
                response.answer.strip(),
                total_tokens,
                input_tokens,
                output_tokens,
            )
            
        except Exception as e:
            API_CALLS.labels(service="trm", status="error").inc()
            logger.error("trm_query_failed", error=str(e))
            # Fallback to Ollama if TRM fails
            logger.warning("falling_back_to_ollama")
            return await self._query_ollama(prompt)
    
    async def _query_ollama(self, prompt: str) -> tuple[str, int, int, int]:
        """
        Query Ollama API for response (fallback only).
        
        Returns:
            Tuple of (response_text, total_tokens, input_tokens, output_tokens)
        """
        try:
            async with httpx.AsyncClient(timeout=120.0) as client:
                response = await client.post(
                    f"{self.ollama_url}/api/generate",
                    json={
                        "model": self.model,
                        "prompt": prompt,
                        "stream": False,
                        "options": {
                            "temperature": 0.7,
                            "top_p": 0.9,
                            "num_predict": 1024,
                        }
                    }
                )
                response.raise_for_status()
                
                API_CALLS.labels(service="ollama", status="success").inc()
                
                result = response.json()
                # Ollama returns prompt_eval_count (input) and eval_count (output)
                input_tokens = result.get("prompt_eval_count", 0)
                output_tokens = result.get("eval_count", 0)
                total_tokens = input_tokens + output_tokens
                
                return (
                    result.get("response", "").strip(),
                    total_tokens,
                    input_tokens,
                    output_tokens,
                )
                
        except httpx.TimeoutException:
            API_CALLS.labels(service="ollama", status="timeout").inc()
            logger.error("ollama_timeout")
            raise
        except httpx.HTTPStatusError as e:
            API_CALLS.labels(service="ollama", status="error").inc()
            logger.error("ollama_http_error", status=e.response.status_code)
            raise
        except Exception as e:
            API_CALLS.labels(service="ollama", status="error").inc()
            logger.error("ollama_query_failed", error=str(e))
            raise
    
    def _record_token_metrics(self, total: int, input_tokens: int, output_tokens: int):
        """Record token usage metrics with proper input/output breakdown."""
        TOKENS_USED.labels(model=self.model, type="total").inc(total)
        TOKENS_USED.labels(model=self.model, type="input").inc(input_tokens)
        TOKENS_USED.labels(model=self.model, type="output").inc(output_tokens)
    
    async def health_check(self) -> bool:
        """Check if TRM (or Ollama fallback) is accessible."""
        if self.trm_client:
            try:
                return await self.trm_client.health_check()
            except Exception:
                pass
        
        # Fallback to Ollama check
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                response = await client.get(f"{self.ollama_url}/api/tags")
                return response.status_code == 200
        except Exception:
            return False
    
    # =========================================================================
    # User Memory API (exposed for external access)
    # =========================================================================
    
    async def get_user_preferences(self, user_id: str) -> dict:
        """Get user preferences from memory."""
        if not self.memory_manager:
            return {}
        
        user_mem = await self.memory_manager.get_user_memory(user_id)
        if user_mem:
            return user_mem.preferences
        return {}
    
    async def set_user_preference(self, user_id: str, key: str, value):
        """Set a user preference."""
        if self.memory_manager:
            await self.memory_manager.update_user_preference(user_id, key, value)
    
    async def set_user_instructions(self, user_id: str, instructions: str):
        """Set custom instructions for a user."""
        if self.memory_manager:
            user_mem = await self.memory_manager.get_or_create_user_memory(user_id)
            user_mem.custom_instructions = instructions
            await self.memory_manager.long_term_store.save(user_mem)
