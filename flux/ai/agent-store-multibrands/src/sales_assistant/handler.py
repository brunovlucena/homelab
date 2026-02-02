"""
Sales Assistant handler.

Provides AI-powered assistance to human sales representatives.
Features:
- Receives escalated conversations from AI sellers
- Provides product recommendations and upsell suggestions
- Generates sales insights and reports
- Handles complex customer queries
"""
import os
import time
from typing import Optional, Any
from datetime import datetime, timezone
from collections import defaultdict

import httpx
import structlog
from cloudevents.http import CloudEvent

from shared.types import (
    Brand,
    Conversation,
    ConversationState,
    Message,
    MessageRole,
    SalesRecommendation,
    SalesInsight,
    EscalationReason,
)
from shared.events import (
    EventType,
    EventPublisher,
    EventSubscriber,
    EscalationEvent,
    ChatMessageEvent,
    ChatResponseEvent,
    SalesInsightEvent,
)
from shared.metrics import (
    ESCALATIONS_TOTAL,
    ESCALATION_WAIT_TIME,
    AI_RESPONSE_TIME,
)

logger = structlog.get_logger()


class EscalationQueue:
    """Manages escalated conversations waiting for human sellers."""
    
    def __init__(self):
        self._queue: dict[str, dict] = {}  # conversation_id -> escalation data
        self._by_brand: dict[str, list[str]] = defaultdict(list)  # brand -> [conv_ids]
        self._by_priority: dict[str, list[str]] = defaultdict(list)  # priority -> [conv_ids]
    
    def add(self, escalation: dict):
        """Add an escalation to the queue."""
        conv_id = escalation.get("conversation_id")
        if conv_id in self._queue:
            return  # Already in queue
        
        escalation["queued_at"] = datetime.now(timezone.utc).isoformat()
        self._queue[conv_id] = escalation
        
        brand = escalation.get("brand", "unknown")
        priority = escalation.get("priority", "medium")
        
        self._by_brand[brand].append(conv_id)
        self._by_priority[priority].append(conv_id)
        
        logger.info(
            "escalation_queued",
            conversation_id=conv_id,
            brand=brand,
            priority=priority,
            queue_size=len(self._queue),
        )
    
    def accept(self, conv_id: str, seller_id: str) -> Optional[dict]:
        """Accept an escalation for handling."""
        if conv_id not in self._queue:
            return None
        
        escalation = self._queue.pop(conv_id)
        
        brand = escalation.get("brand", "unknown")
        priority = escalation.get("priority", "medium")
        
        if conv_id in self._by_brand[brand]:
            self._by_brand[brand].remove(conv_id)
        if conv_id in self._by_priority[priority]:
            self._by_priority[priority].remove(conv_id)
        
        escalation["accepted_by"] = seller_id
        escalation["accepted_at"] = datetime.now(timezone.utc).isoformat()
        
        # Calculate wait time
        queued_at = datetime.fromisoformat(escalation.get("queued_at", ""))
        wait_seconds = (datetime.now(timezone.utc) - queued_at).total_seconds()
        
        ESCALATION_WAIT_TIME.labels(
            brand=brand,
            priority=priority
        ).observe(wait_seconds)
        
        logger.info(
            "escalation_accepted",
            conversation_id=conv_id,
            seller_id=seller_id,
            wait_seconds=wait_seconds,
        )
        
        return escalation
    
    def get_pending(
        self,
        brand: Optional[str] = None,
        priority: Optional[str] = None,
        limit: int = 10,
    ) -> list[dict]:
        """Get pending escalations."""
        if priority:
            conv_ids = self._by_priority.get(priority, [])
        elif brand:
            conv_ids = self._by_brand.get(brand, [])
        else:
            conv_ids = list(self._queue.keys())
        
        result = []
        for conv_id in conv_ids[:limit]:
            if conv_id in self._queue:
                result.append(self._queue[conv_id])
        
        return result
    
    @property
    def queue_size(self) -> int:
        return len(self._queue)


class SalesAssistant:
    """
    AI-powered assistant for human sales representatives.
    
    Features:
    - Escalation handling and routing
    - Real-time sales suggestions
    - Customer sentiment analysis
    - Product knowledge base queries
    - Sales performance insights
    """
    
    def __init__(
        self,
        ollama_url: str = None,
        model: str = None,
        event_publisher: Optional[EventPublisher] = None,
        event_subscriber: Optional[EventSubscriber] = None,
    ):
        self.ollama_url = ollama_url or os.getenv(
            "OLLAMA_URL",
            "http://ollama-native.ollama.svc.cluster.local:11434"
        )
        self.model = model or os.getenv("OLLAMA_MODEL", "llama3.2:3b")
        
        self.event_publisher = event_publisher
        self.event_subscriber = event_subscriber
        
        # Escalation queue
        self.escalation_queue = EscalationQueue()
        
        # Active conversations being handled by human sellers
        self._active_conversations: dict[str, dict] = {}
        
        # Sales insights cache
        self._insights: list[dict] = []
        self._max_insights = 100
        
        logger.info(
            "sales_assistant_initialized",
            model=self.model,
        )
    
    # =========================================================================
    # Escalation Handling
    # =========================================================================
    
    async def handle_escalation(self, escalation: dict) -> dict:
        """
        Handle an incoming escalation from AI seller.
        
        Returns escalation status and any immediate suggestions.
        """
        conv_id = escalation.get("conversation_id")
        brand = escalation.get("brand", "unknown")
        reason = escalation.get("reason", "unknown")
        
        log = logger.bind(conversation_id=conv_id, brand=brand)
        
        # Add to queue
        self.escalation_queue.add(escalation)
        
        # Generate initial insights
        insights = await self._generate_escalation_insights(escalation)
        
        # Emit insight event
        if insights and self.event_publisher:
            await self.event_publisher.emit_sales_insight(
                SalesInsightEvent(
                    conversation_id=conv_id,
                    customer_phone=escalation.get("customer_phone", ""),
                    insight_type="escalation",
                    summary=insights.get("summary", ""),
                    suggested_action=insights.get("suggested_action", ""),
                    priority=escalation.get("priority", "medium"),
                    brand=brand,
                    data=insights,
                )
            )
        
        log.info("escalation_processed", reason=reason)
        
        return {
            "status": "queued",
            "queue_position": self.escalation_queue.queue_size,
            "insights": insights,
        }
    
    async def _generate_escalation_insights(self, escalation: dict) -> dict:
        """Generate AI insights for an escalation."""
        try:
            context = escalation.get("context_summary", "")
            reason = escalation.get("reason", "")
            sentiment = escalation.get("customer_sentiment", 0.5)
            
            prompt = f"""Você é um assistente de vendas experiente.

Analise esta escalação de conversa:
- Motivo: {reason}
- Sentimento do cliente: {sentiment:.1%} positivo
- Contexto: {context}

Forneça uma análise breve em JSON:
{{
    "summary": "Resumo da situação em 1-2 frases",
    "customer_mood": "calmo|frustrado|irritado|ansioso|neutro",
    "urgency": "baixa|média|alta",
    "suggested_action": "Ação recomendada para o vendedor",
    "talking_points": ["ponto 1", "ponto 2"],
    "avoid": ["coisa a evitar"]
}}"""
            
            with AI_RESPONSE_TIME.labels(brand="assistant", model=self.model).time():
                response, _ = await self._query_llm(prompt)
            
            import json
            import re
            
            json_match = re.search(r'\{[^{}]*\}', response, re.DOTALL)
            if json_match:
                return json.loads(json_match.group())
            
            return {
                "summary": "Escalação recebida",
                "suggested_action": "Revisar histórico da conversa",
            }
            
        except Exception as e:
            logger.error("insight_generation_failed", error=str(e))
            return {
                "summary": "Erro ao gerar insights",
                "suggested_action": "Verificar manualmente",
            }
    
    # =========================================================================
    # Human Seller Assistance
    # =========================================================================
    
    async def get_suggestions(
        self,
        conversation_context: str,
        customer_message: str,
        brand: str,
    ) -> dict:
        """
        Get AI suggestions for human seller's response.
        
        Returns suggested responses and product recommendations.
        """
        start_time = time.time()
        
        prompt = f"""Você é um assistente de vendas experiente na marca {brand}.

Contexto da conversa:
{conversation_context}

Última mensagem do cliente:
{customer_message}

Ajude o vendedor humano sugerindo:
1. Uma resposta amigável e profissional
2. Produtos que podem interessar o cliente
3. Técnicas de venda apropriadas

Responda em JSON:
{{
    "suggested_response": "Resposta sugerida para o vendedor usar/adaptar",
    "alternative_responses": ["outra opção 1", "outra opção 2"],
    "product_suggestions": [
        {{"reason": "motivo", "approach": "como apresentar"}}
    ],
    "sales_technique": "técnica recomendada",
    "sentiment_tip": "dica baseada no sentimento do cliente"
}}"""
        
        try:
            with AI_RESPONSE_TIME.labels(brand="assistant", model=self.model).time():
                response, tokens = await self._query_llm(prompt)
            
            import json
            import re
            
            json_match = re.search(r'\{[^{}]*\}', response, re.DOTALL)
            if json_match:
                data = json.loads(json_match.group())
                data["duration_ms"] = (time.time() - start_time) * 1000
                data["tokens_used"] = tokens
                return data
            
            return {
                "suggested_response": response,
                "duration_ms": (time.time() - start_time) * 1000,
            }
            
        except Exception as e:
            logger.error("suggestions_failed", error=str(e))
            return {"error": str(e)}
    
    async def analyze_objection(
        self,
        objection: str,
        product: str,
        brand: str,
    ) -> dict:
        """
        Analyze customer objection and suggest responses.
        """
        prompt = f"""Você é um especialista em vendas da marca {brand}.

O cliente levantou esta objeção sobre o produto {product}:
"{objection}"

Forneça análise e técnicas de resposta em JSON:
{{
    "objection_type": "preço|qualidade|necessidade|timing|confiança|outro",
    "root_cause": "Causa raiz provável da objeção",
    "response_options": [
        {{"approach": "nome da técnica", "script": "exemplo de resposta"}}
    ],
    "questions_to_ask": ["pergunta para entender melhor"],
    "avoid_saying": ["frase a evitar"]
}}"""
        
        try:
            response, _ = await self._query_llm(prompt)
            
            import json
            import re
            
            json_match = re.search(r'\{[^{}]*\}', response, re.DOTALL)
            if json_match:
                return json.loads(json_match.group())
            
            return {"response": response}
            
        except Exception as e:
            logger.error("objection_analysis_failed", error=str(e))
            return {"error": str(e)}
    
    # =========================================================================
    # Event Handlers
    # =========================================================================
    
    async def handle_escalation_event(self, event: CloudEvent):
        """Handle escalation event from AI sellers."""
        data = event.data or {}
        
        await self.handle_escalation({
            "conversation_id": data.get("conversation_id"),
            "customer_id": data.get("customer_id"),
            "customer_phone": data.get("customer_phone"),
            "reason": data.get("reason"),
            "priority": data.get("priority", "medium"),
            "brand": data.get("brand"),
            "ai_seller_id": data.get("ai_seller_id"),
            "context_summary": data.get("context_summary"),
            "cart_value": data.get("cart_value", 0),
            "customer_sentiment": data.get("customer_sentiment", 0.5),
        })
    
    async def handle_chat_message_event(self, event: CloudEvent):
        """
        Handle chat message events for active escalated conversations.
        
        When a human seller is handling a conversation, provide real-time suggestions.
        """
        data = event.data or {}
        conv_id = data.get("conversation_id")
        
        # Only process if this conversation is actively being handled
        if conv_id not in self._active_conversations:
            return
        
        # Generate suggestions for the human seller
        active_conv = self._active_conversations[conv_id]
        
        suggestions = await self.get_suggestions(
            conversation_context=active_conv.get("context", ""),
            customer_message=data.get("message", ""),
            brand=data.get("brand", ""),
        )
        
        # Emit suggestions as insight
        if self.event_publisher:
            await self.event_publisher.emit_sales_insight(
                SalesInsightEvent(
                    conversation_id=conv_id,
                    customer_phone=data.get("customer_phone", ""),
                    insight_type="suggestion",
                    summary="Sugestões de resposta disponíveis",
                    suggested_action=suggestions.get("suggested_response", ""),
                    priority="medium",
                    brand=data.get("brand", ""),
                    data=suggestions,
                )
            )
    
    def setup_event_handlers(self):
        """Register event handlers."""
        if self.event_subscriber:
            self.event_subscriber.register(
                EventType.SALES_ESCALATE,
                self.handle_escalation_event
            )
            self.event_subscriber.register(
                EventType.CHAT_MESSAGE_NEW,
                self.handle_chat_message_event
            )
            logger.info("sales_assistant_event_handlers_registered")
    
    # =========================================================================
    # Helper Methods
    # =========================================================================
    
    async def _query_llm(self, prompt: str) -> tuple[str, int]:
        """Query the LLM."""
        try:
            async with httpx.AsyncClient(timeout=60.0) as client:
                response = await client.post(
                    f"{self.ollama_url}/api/generate",
                    json={
                        "model": self.model,
                        "prompt": prompt,
                        "stream": False,
                        "options": {
                            "temperature": 0.7,
                            "num_predict": 512,
                        },
                    }
                )
                response.raise_for_status()
                
                result = response.json()
                return (
                    result.get("response", "").strip(),
                    result.get("eval_count", 0)
                )
                
        except Exception as e:
            logger.error("llm_query_failed", error=str(e))
            raise
    
    async def health_check(self) -> bool:
        """Check if LLM is accessible."""
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                response = await client.get(f"{self.ollama_url}/api/tags")
                return response.status_code == 200
        except Exception:
            return False
    
    # =========================================================================
    # Queue Management
    # =========================================================================
    
    def accept_escalation(self, conv_id: str, seller_id: str) -> Optional[dict]:
        """Accept an escalation for handling."""
        escalation = self.escalation_queue.accept(conv_id, seller_id)
        
        if escalation:
            self._active_conversations[conv_id] = {
                "seller_id": seller_id,
                "escalation": escalation,
                "started_at": datetime.now(timezone.utc).isoformat(),
            }
        
        return escalation
    
    def resolve_escalation(self, conv_id: str, outcome: str) -> bool:
        """Mark an escalation as resolved."""
        if conv_id not in self._active_conversations:
            return False
        
        conv_data = self._active_conversations.pop(conv_id)
        
        logger.info(
            "escalation_resolved",
            conversation_id=conv_id,
            seller_id=conv_data.get("seller_id"),
            outcome=outcome,
        )
        
        return True
    
    def get_queue_status(self) -> dict:
        """Get current queue status."""
        return {
            "total_pending": self.escalation_queue.queue_size,
            "active_conversations": len(self._active_conversations),
            "by_priority": {
                "urgent": len(self.escalation_queue._by_priority.get("urgent", [])),
                "high": len(self.escalation_queue._by_priority.get("high", [])),
                "medium": len(self.escalation_queue._by_priority.get("medium", [])),
                "low": len(self.escalation_queue._by_priority.get("low", [])),
            },
        }
