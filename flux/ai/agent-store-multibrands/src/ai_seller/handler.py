"""
AI Seller handler with TRM (Tiny Recursive Model) integration.

Uses TRM with built-in reflection for intelligent sales conversations.
Each brand has customized prompts and personality.
"""
import os
import json
import time
import re
from typing import Optional, Any
from uuid import uuid4
from decimal import Decimal

import httpx
import structlog
from cloudevents.http import CloudEvent

from shared.types import (
    Brand,
    Product,
    Order,
    OrderItem,
    OrderStatus,
    Message,
    MessageRole,
    Conversation,
    ConversationState,
    SalesRecommendation,
    EscalationReason,
)
from shared.events import (
    EventType,
    EventPublisher,
    EventSubscriber,
    ChatMessageEvent,
    ChatResponseEvent,
    ProductQueryEvent,
    OrderEvent,
    EscalationEvent,
    SalesInsightEvent,
)
from shared.metrics import (
    MESSAGES_SENT,
    AI_RESPONSE_TIME,
    AI_TOKENS_USED,
    AI_RECOMMENDATIONS_MADE,
    ESCALATIONS_TOTAL,
    ORDERS_CREATED,
    CONVERSATIONS_ACTIVE,
    CUSTOMER_SENTIMENT,
)

# Import TRM client
try:
    from agent_trm import TRMClient, TRMRequest, ReflectionMode
    TRM_AVAILABLE = True
except ImportError:
    TRM_AVAILABLE = False

logger = structlog.get_logger()


# =============================================================================
# Brand Personality Prompts
# =============================================================================

BRAND_PROMPTS = {
    Brand.FASHION: """Você é LUNA, uma consultora de moda sofisticada e estilosa da loja MultiBrands.

PERSONALIDADE:
- Elegante e conhecedora das últimas tendências
- Entusiasta de moda sustentável
- Fala com confiança sobre estilo e combinações
- Usa emojis de moda com moderação (👗✨💃)

ESTILO DE COMUNICAÇÃO:
- Sugira looks completos, não apenas peças isoladas
- Pergunte sobre ocasiões e preferências de estilo
- Ofereça dicas de como combinar peças
- Mencione tendências atuais quando relevante

ESPECIALIDADES:
- Roupas femininas e masculinas
- Acessórios e bolsas
- Calçados
- Moda sustentável""",

    Brand.TECH: """Você é MAX, um especialista em tecnologia apaixonado da loja MultiBrands.

PERSONALIDADE:
- Conhecedor profundo de especificações técnicas
- Entusiasta de novidades tecnológicas
- Objetivo e prático nas recomendações
- Usa emojis tech com moderação (📱💻🎧)

ESTILO DE COMUNICAÇÃO:
- Compare especificações de forma clara
- Pergunte sobre o uso pretendido
- Explique benefícios de forma simples
- Mencione compatibilidade entre produtos

ESPECIALIDADES:
- Smartphones e tablets
- Notebooks e computadores
- Fones e acessórios
- Smart home""",

    Brand.HOME: """Você é SOFIA, uma designer de interiores calorosa da loja MultiBrands.

PERSONALIDADE:
- Acolhedora e atenciosa
- Apaixonada por criar ambientes harmoniosos
- Práctica e criativa
- Usa emojis de casa com moderação (🏠✨🛋️)

ESTILO DE COMUNICAÇÃO:
- Pergunte sobre o espaço e estilo desejado
- Sugira combinações de cores e materiais
- Ofereça dicas de decoração
- Mencione tendências de design

ESPECIALIDADES:
- Móveis e decoração
- Artigos para cozinha
- Organização de ambientes
- Iluminação""",

    Brand.BEAUTY: """Você é BELLA, uma especialista em beleza glamourosa da loja MultiBrands.

PERSONALIDADE:
- Carinhosa e empática
- Conhecedora de skincare e maquiagem
- Atenta às necessidades individuais de cada pele
- Usa emojis de beleza com moderação (💄✨💖)

ESTILO DE COMUNICAÇÃO:
- Pergunte sobre tipo de pele e preferências
- Sugira rotinas de cuidados completas
- Explique ingredientes importantes
- Recomende produtos para cada necessidade

ESPECIALIDADES:
- Skincare e tratamentos
- Maquiagem
- Perfumaria
- Cuidados com cabelo""",

    Brand.GAMING: """Você é PIXEL, um gamer entusiasmado da loja MultiBrands.

PERSONALIDADE:
- Energético e divertido
- Expert em games de todas as plataformas
- Conhecedor de setups e periféricos
- Usa emojis de games com moderação (🎮🕹️🔥)

ESTILO DE COMUNICAÇÃO:
- Pergunte sobre plataformas e gêneros favoritos
- Compare specs de hardware
- Sugira jogos baseado em preferências
- Mencione lançamentos e promoções

ESPECIALIDADES:
- Consoles e jogos
- Periféricos gamer
- PC gaming
- Acessórios e merchandise""",
}


# =============================================================================
# Conversation Manager
# =============================================================================

class ConversationManager:
    """Manages customer conversations."""
    
    def __init__(self, max_conversations: int = 1000):
        self._conversations: dict[str, Conversation] = {}
        self._max_conversations = max_conversations
    
    def get_or_create(
        self,
        conversation_id: str,
        customer_id: str,
        customer_phone: str,
        brand: Brand,
    ) -> Conversation:
        """Get existing or create new conversation."""
        if conversation_id in self._conversations:
            conv = self._conversations[conversation_id]
            conv.customer_phone = customer_phone  # Update in case changed
            return conv
        
        # Cleanup if limit reached
        if len(self._conversations) >= self._max_conversations:
            self._cleanup_oldest()
        
        conv = Conversation(
            id=conversation_id,
            customer_id=customer_id,
            customer_phone=customer_phone,
            brand=brand,
        )
        self._conversations[conversation_id] = conv
        
        CONVERSATIONS_ACTIVE.labels(
            brand=brand.value,
            state=ConversationState.NEW.value
        ).inc()
        
        return conv
    
    def _cleanup_oldest(self):
        """Remove oldest conversations."""
        if not self._conversations:
            return
        
        sorted_convs = sorted(
            self._conversations.items(),
            key=lambda x: x[1].updated_at
        )
        
        to_remove = max(1, len(sorted_convs) // 10)
        for conv_id, conv in sorted_convs[:to_remove]:
            CONVERSATIONS_ACTIVE.labels(
                brand=conv.brand.value,
                state=conv.state.value
            ).dec()
            del self._conversations[conv_id]


# =============================================================================
# AI Seller Agent
# =============================================================================

class AISeller:
    """
    AI-powered sales agent with TRM (Tiny Recursive Model).
    
    Uses TRM with built-in reflection to handle customer conversations and drive sales.
    """
    
    def __init__(
        self,
        brand: Brand,
        ollama_url: str = None,  # Deprecated: kept for backward compatibility
        model: str = None,  # Deprecated: kept for backward compatibility
        event_publisher: Optional[EventPublisher] = None,
        event_subscriber: Optional[EventSubscriber] = None,
        # TRM configuration
        trm_model_name: str = None,
        trm_use_hf_api: bool = None,
    ):
        self.brand = brand
        
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
            "http://ollama-native.ollama.svc.cluster.local:11434"
        )
        self.model = model or os.getenv("OLLAMA_MODEL", "llama3.2:3b")
        
        self.event_publisher = event_publisher
        self.event_subscriber = event_subscriber
        
        self.conversations = ConversationManager()
        
        # Get brand-specific prompt
        self.system_prompt = self._build_system_prompt()
        
        # Product cache (would be populated via events)
        self._products: dict[str, Product] = {}
        
        # Initialize TRM client
        self.trm_client: Optional[TRMClient] = None
        if TRM_AVAILABLE:
            try:
                self.trm_client = TRMClient(
                    model_name=self.trm_model_name,
                    use_hf_api=self.trm_use_hf_api,
                )
                logger.info("trm_client_initialized_for_seller", model_name=self.trm_model_name)
            except Exception as e:
                logger.error("trm_client_init_failed", error=str(e))
                self.trm_client = None
        
        logger.info(
            "ai_seller_initialized",
            brand=brand.value,
            trm_enabled=TRM_AVAILABLE and self.trm_client is not None,
            trm_model=self.trm_model_name,
            ollama_fallback=self.model,
        )
    
    def _build_system_prompt(self) -> str:
        """Build the complete system prompt."""
        base_prompt = BRAND_PROMPTS.get(self.brand, BRAND_PROMPTS[Brand.FASHION])
        
        return f"""{base_prompt}

REGRAS:
1. Seja educado e profissional
2. Responda em português do Brasil
3. Seja útil e amigável
4. Ofereça ajuda e sugestões de produtos
5. Mantenha respostas concisas (2-4 frases)
6. Use emojis com moderação

Responda diretamente ao cliente, sem formatação especial."""
    
    async def handle_message(
        self,
        message: str,
        customer_id: str,
        customer_phone: str,
        conversation_id: Optional[str] = None,
    ) -> dict:
        """
        Handle an incoming customer message.
        
        Returns a response dict with message and metadata.
        """
        start_time = time.time()
        log = logger.bind(
            customer_phone=customer_phone,
            brand=self.brand.value,
        )
        
        try:
            # Get or create conversation
            conv_id = conversation_id or f"conv-{customer_phone}-{self.brand.value}"
            conv = self.conversations.get_or_create(
                conv_id,
                customer_id,
                customer_phone,
                self.brand,
            )
            
            # Add customer message
            conv.add_message(MessageRole.CUSTOMER, message)
            
            # Build prompt with context
            prompt = self._build_prompt(conv, message)
            
            # Query TRM (or Ollama fallback)
            with AI_RESPONSE_TIME.labels(brand=self.brand.value, model=self.model).time():
                if self.trm_client:
                    response_text, tokens = await self._query_trm(prompt, conv_id)
                else:
                    response_text, tokens = await self._query_llm(prompt)
            
            logger.info("raw_llm_response", 
                        response_length=len(response_text) if response_text else 0,
                        response_preview=response_text[:200] if response_text else "EMPTY")
            
            # Parse response
            response_data = self._parse_response(response_text)
            
            # Update conversation state
            conv.sentiment_score = response_data.get("sentiment", 0.5)
            conv.intent = response_data.get("intent", "")
            
            # Add AI response to conversation
            conv.add_message(MessageRole.AI_SELLER, response_data.get("message", ""))
            
            # Calculate metrics
            duration_ms = (time.time() - start_time) * 1000
            
            AI_TOKENS_USED.labels(
                brand=self.brand.value,
                model=self.model,
                type="total"
            ).inc(tokens)
            
            MESSAGES_SENT.labels(
                brand=self.brand.value,
                sender_type="ai",
                type="text"
            ).inc()
            
            CUSTOMER_SENTIMENT.labels(brand=self.brand.value).observe(conv.sentiment_score)
            
            log.info(
                "message_handled",
                tokens=tokens,
                duration_ms=duration_ms,
                intent=response_data.get("intent"),
            )
            
            # Check for escalation
            if response_data.get("escalate"):
                await self._handle_escalation(conv, response_data)
            
            # Emit events
            if self.event_publisher:
                await self._emit_response_event(
                    conv,
                    response_data,
                    tokens,
                    duration_ms,
                )
            
            # Track recommendations
            for rec in response_data.get("recommendations", []):
                AI_RECOMMENDATIONS_MADE.labels(
                    brand=self.brand.value,
                    type=rec.get("type", "general")
                ).inc()
            
            return {
                "message": response_data.get("message", ""),
                "conversation_id": conv.id,
                "products_mentioned": response_data.get("products_mentioned", []),
                "recommendations": response_data.get("recommendations", []),
                "actions": response_data.get("actions", []),
                "tokens_used": tokens,
                "duration_ms": duration_ms,
                "escalated": response_data.get("escalate", False),
            }
            
        except Exception as e:
            log.error("message_handling_failed", error=str(e))
            return {
                "message": "Desculpe, tive um problema técnico. Pode repetir sua mensagem? 🙏",
                "conversation_id": conv_id if 'conv_id' in locals() else "",
                "products_mentioned": [],
                "recommendations": [],
                "actions": [],
                "tokens_used": 0,
                "duration_ms": 0.0,
                "escalated": False,
                "error": str(e),
            }
    
    def _build_prompt(self, conv: Conversation, current_message: str) -> str:
        """Build the full prompt with conversation context."""
        # Get conversation history
        context = conv.get_context(max_messages=6)
        
        prompt_parts = [
            f"Sistema: {self.system_prompt}\n",
        ]
        
        # Add some context if there's history
        if len(context) > 1:
            prompt_parts.append("\n--- Conversa ---\n")
            for msg in context[:-1]:  # Exclude the current message we're about to add
                role = msg["role"]
                content = msg["content"]
                if role == "customer":
                    prompt_parts.append(f"Cliente: {content}\n")
                elif role == "ai_seller":
                    prompt_parts.append(f"Você: {content}\n")
        
        prompt_parts.append(f"\nCliente: {current_message}\n")
        prompt_parts.append("Você:")
        
        return "\n".join(prompt_parts)
    
    async def _query_trm(self, prompt: str, conversation_id: Optional[str] = None) -> tuple[str, int]:
        """Query TRM (Tiny Recursive Model) with built-in reflection."""
        if not self.trm_client:
            raise RuntimeError("TRM client not initialized")
        
        try:
            request = TRMRequest(
                prompt=prompt,
                max_reflection_steps=2,  # Fewer steps for sales conversations
                reflection_mode=ReflectionMode.AUTO,
                max_tokens=512,
                conversation_id=conversation_id,
            )
            
            response = await self.trm_client.generate(request)
            
            logger.info(
                "trm_sales_response_generated",
                reflection_steps=response.reflection_steps,
                confidence=response.confidence,
                duration_ms=response.duration_ms,
            )
            
            return (response.answer.strip(), response.tokens_used)
            
        except Exception as e:
            logger.error("trm_query_failed", error=str(e))
            # Fallback to Ollama
            logger.warning("falling_back_to_ollama")
            return await self._query_llm(prompt)
    
    async def _query_llm(self, prompt: str) -> tuple[str, int]:
        """Query the LLM for a response (Ollama fallback)."""
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
                            "num_predict": 256,
                        },
                    }
                )
                response.raise_for_status()
                
                result = response.json()
                response_text = str(result.get("response", "")).strip()
                tokens = int(result.get("eval_count", 0))
                
                # Debug logging
                logger.info("llm_raw_result", 
                           has_response="response" in result,
                           response_type=type(result.get("response")).__name__,
                           response_len=len(response_text),
                           tokens=tokens)
                
                return (response_text, tokens)
                
        except httpx.TimeoutException:
            logger.error("llm_timeout")
            raise
        except Exception as e:
            logger.error("llm_query_failed", error=str(e), error_type=type(e).__name__)
            raise
    
    def _parse_response(self, response_text: str) -> dict:
        """Parse response from LLM - just use the text directly."""
        logger.info("parse_input", 
                   input_type=type(response_text).__name__,
                   input_len=len(response_text) if response_text else 0,
                   input_repr=repr(response_text[:100]) if response_text else "None")
        
        # Simply use the response text as the message
        message = str(response_text) if response_text else ""
        
        # Clean up common artifacts
        message = message.strip()
        
        # Remove any leading/trailing quotes if present
        if len(message) >= 2 and message.startswith('"') and message.endswith('"'):
            message = message[1:-1]
        
        # Remove common prefixes that LLMs sometimes add
        prefixes_to_remove = ["Você:", "Assistant:", "AI:", "Luna:", "Max:", "Sofia:", "Bella:", "Pixel:"]
        for prefix in prefixes_to_remove:
            if message.startswith(prefix):
                message = message[len(prefix):].strip()
                break
        
        # Final strip and check
        message = message.strip()
        
        # If empty for some reason, provide default based on brand
        if not message:
            default_messages = {
                "fashion": "Olá! Sou Luna, sua consultora de moda. Como posso ajudá-la hoje? 👗",
                "tech": "Olá! Sou Max, especialista em tecnologia. Como posso ajudá-lo? 📱",
                "home": "Olá! Sou Sofia, sua designer de interiores. Como posso ajudá-la? 🏠",
                "beauty": "Olá! Sou Bella, sua especialista em beleza. Como posso ajudá-la? 💄",
                "gaming": "Olá! Sou Pixel, seu especialista em games. Como posso ajudá-lo? 🎮",
            }
            message = default_messages.get(self.brand.value, "Olá! Como posso ajudá-lo hoje? 👋")
            logger.warning("empty_response_using_default", brand=self.brand.value)
        
        logger.info("parse_output", message_len=len(message), message_preview=message[:80])
        
        return {
            "message": message,
            "intent": "browse",
            "products_mentioned": [],
            "recommendations": [],
            "actions": [],
            "escalate": False,
            "sentiment": 0.5,
        }
    
    async def _handle_escalation(self, conv: Conversation, response_data: dict):
        """Handle escalation to human seller."""
        reason = response_data.get("escalation_reason", "customer_request")
        
        conv.state = ConversationState.ESCALATED
        conv.escalation_reason = reason
        
        ESCALATIONS_TOTAL.labels(
            brand=self.brand.value,
            reason=reason,
            priority="medium"
        ).inc()
        
        if self.event_publisher:
            await self.event_publisher.emit_escalation(
                EscalationEvent(
                    conversation_id=conv.id,
                    customer_id=conv.customer_id,
                    customer_phone=conv.customer_phone,
                    reason=reason,
                    brand=self.brand.value,
                    ai_seller_id=f"ai-seller-{self.brand.value}",
                    context_summary=response_data.get("message", ""),
                    customer_sentiment=conv.sentiment_score,
                )
            )
        
        logger.info(
            "conversation_escalated",
            conversation_id=conv.id,
            reason=reason,
        )
    
    async def _emit_response_event(
        self,
        conv: Conversation,
        response_data: dict,
        tokens: int,
        duration_ms: float,
    ):
        """Emit chat response event."""
        await self.event_publisher.emit_chat_response(
            ChatResponseEvent(
                conversation_id=conv.id,
                customer_phone=conv.customer_phone,
                response=response_data.get("message", ""),
                brand=self.brand.value,
                ai_seller_id=f"ai-seller-{self.brand.value}",
                products_mentioned=response_data.get("products_mentioned", []),
                recommendations=response_data.get("recommendations", []),
                suggested_actions=response_data.get("actions", []),
                tokens_used=tokens,
                duration_ms=duration_ms,
            )
        )
    
    # =========================================================================
    # Event Handlers
    # =========================================================================
    
    async def handle_chat_message_event(self, event: CloudEvent):
        """Handle incoming chat message event."""
        data = event.data or {}
        
        # Only process messages for our brand
        message_brand = data.get("brand", "")
        if message_brand != self.brand.value:
            return
        
        message = data.get("message", "")
        customer_id = data.get("customer_id", "")
        customer_phone = data.get("customer_phone", "")
        conversation_id = data.get("conversation_id")
        
        await self.handle_message(
            message=message,
            customer_id=customer_id,
            customer_phone=customer_phone,
            conversation_id=conversation_id,
        )
    
    async def handle_product_query_result(self, event: CloudEvent):
        """Handle product query result event."""
        data = event.data or {}
        products = data.get("products", [])
        
        # Cache products
        for p in products:
            self._products[p.get("id")] = Product(
                id=p.get("id"),
                name=p.get("name"),
                brand=Brand(p.get("brand", "fashion")),
                description=p.get("description", ""),
                price=Decimal(str(p.get("price", 0))),
            )
    
    def setup_event_handlers(self):
        """Register event handlers."""
        if self.event_subscriber:
            self.event_subscriber.register(
                EventType.CHAT_MESSAGE_NEW,
                self.handle_chat_message_event
            )
            self.event_subscriber.register(
                EventType.PRODUCT_QUERY_RESULT,
                self.handle_product_query_result
            )
            logger.info("ai_seller_event_handlers_registered", brand=self.brand.value)
    
    async def health_check(self) -> bool:
        """Check if LLM is accessible."""
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                response = await client.get(f"{self.ollama_url}/api/tags")
                return response.status_code == 200
        except Exception:
            return False
