"""
CloudEvents integration for Agent Store MultiBrands.

Enables event-driven communication between all store agents via Knative Eventing.

Event Naming Convention:
    store.{domain}.{action}[.{detail}]

Examples:
    store.whatsapp.message.received
    store.chat.message.new
    store.order.created
    store.sales.escalate
"""
import os
import json
from typing import Optional, Callable, Awaitable, Any
from dataclasses import dataclass, asdict
from datetime import datetime, timezone
from enum import Enum

import httpx
import structlog
from cloudevents.http import CloudEvent, to_structured, from_http
from prometheus_client import Counter

logger = structlog.get_logger()


# =============================================================================
# Metrics
# =============================================================================

EVENTS_EMITTED = Counter(
    "store_events_emitted_total",
    "Total CloudEvents emitted",
    ["event_type", "source", "status"]
)

EVENTS_RECEIVED = Counter(
    "store_events_received_total",
    "Total CloudEvents received",
    ["event_type", "status"]
)


# =============================================================================
# Event Type Definitions
# =============================================================================

class EventType(str, Enum):
    """CloudEvent types for the multi-brand store."""
    
    # ─────────────────────────────────────────────────────────────────────────
    # WhatsApp Events (Gateway)
    # ─────────────────────────────────────────────────────────────────────────
    WHATSAPP_MESSAGE_RECEIVED = "store.whatsapp.message.received"
    WHATSAPP_MESSAGE_SEND = "store.whatsapp.message.send"
    WHATSAPP_MESSAGE_STATUS = "store.whatsapp.message.status"
    WHATSAPP_MEDIA_RECEIVED = "store.whatsapp.media.received"
    
    # ─────────────────────────────────────────────────────────────────────────
    # Chat Events (AI Seller Communication)
    # ─────────────────────────────────────────────────────────────────────────
    CHAT_MESSAGE_NEW = "store.chat.message.new"
    CHAT_RESPONSE = "store.chat.response"
    CHAT_INTENT_DETECTED = "store.chat.intent.detected"
    CHAT_SENTIMENT_CHANGED = "store.chat.sentiment.changed"
    
    # ─────────────────────────────────────────────────────────────────────────
    # Product Events (Catalog)
    # ─────────────────────────────────────────────────────────────────────────
    PRODUCT_QUERY = "store.product.query"
    PRODUCT_QUERY_RESULT = "store.product.query.result"
    PRODUCT_RECOMMEND = "store.product.recommend"
    PRODUCT_RECOMMEND_RESULT = "store.product.recommend.result"
    PRODUCT_STOCK_LOW = "store.product.stock.low"
    PRODUCT_VIEWED = "store.product.viewed"
    
    # ─────────────────────────────────────────────────────────────────────────
    # Cart Events
    # ─────────────────────────────────────────────────────────────────────────
    CART_ITEM_ADDED = "store.cart.item.added"
    CART_ITEM_REMOVED = "store.cart.item.removed"
    CART_UPDATED = "store.cart.updated"
    CART_ABANDONED = "store.cart.abandoned"
    
    # ─────────────────────────────────────────────────────────────────────────
    # Order Events
    # ─────────────────────────────────────────────────────────────────────────
    ORDER_CREATE = "store.order.create"
    ORDER_CREATED = "store.order.created"
    ORDER_CONFIRMED = "store.order.confirmed"
    ORDER_SHIPPED = "store.order.shipped"
    ORDER_DELIVERED = "store.order.delivered"
    ORDER_CANCELLED = "store.order.cancelled"
    ORDER_STATUS_UPDATE = "store.order.status.update"
    
    # ─────────────────────────────────────────────────────────────────────────
    # Sales Events (Human Seller Assistance)
    # ─────────────────────────────────────────────────────────────────────────
    SALES_ESCALATE = "store.sales.escalate"
    SALES_ESCALATION_ACCEPTED = "store.sales.escalation.accepted"
    SALES_ESCALATION_RESOLVED = "store.sales.escalation.resolved"
    SALES_INSIGHT = "store.sales.insight"
    SALES_RECOMMENDATION = "store.sales.recommendation"
    SALES_HANDOFF = "store.sales.handoff"
    
    # ─────────────────────────────────────────────────────────────────────────
    # Customer Events
    # ─────────────────────────────────────────────────────────────────────────
    CUSTOMER_NEW = "store.customer.new"
    CUSTOMER_UPDATED = "store.customer.updated"
    CUSTOMER_VIP_DETECTED = "store.customer.vip.detected"
    
    # ─────────────────────────────────────────────────────────────────────────
    # Analytics Events
    # ─────────────────────────────────────────────────────────────────────────
    ANALYTICS_CONVERSION = "store.analytics.conversion"
    ANALYTICS_SESSION = "store.analytics.session"


# =============================================================================
# Event Data Structures
# =============================================================================

@dataclass
class WhatsAppMessageEvent:
    """Data for WhatsApp message events."""
    message_id: str
    phone_from: str
    phone_to: str
    message_type: str
    content: str
    media_url: Optional[str] = None
    context_message_id: Optional[str] = None
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()


@dataclass
class ChatMessageEvent:
    """Data for chat message events."""
    conversation_id: str
    customer_id: str
    customer_phone: str
    brand: str
    message: str
    message_role: str = "customer"
    intent: Optional[str] = None
    sentiment: float = 0.5
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()


@dataclass
class ChatResponseEvent:
    """Data for AI seller response events."""
    conversation_id: str
    customer_phone: str
    response: str
    brand: str
    ai_seller_id: str
    products_mentioned: list[str] = None
    recommendations: list[dict] = None
    suggested_actions: list[str] = None
    tokens_used: int = 0
    duration_ms: float = 0.0
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()
        if self.products_mentioned is None:
            self.products_mentioned = []
        if self.recommendations is None:
            self.recommendations = []
        if self.suggested_actions is None:
            self.suggested_actions = []


@dataclass
class ProductQueryEvent:
    """Data for product query events."""
    conversation_id: str
    query: str
    brand: Optional[str] = None
    category: Optional[str] = None
    price_min: Optional[float] = None
    price_max: Optional[float] = None
    limit: int = 5
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()


@dataclass
class ProductRecommendEvent:
    """Data for product recommendation request."""
    conversation_id: str
    customer_id: str
    brand: str
    context: str  # Conversation context
    current_cart: list[str] = None  # Product IDs in cart
    purchase_history: list[str] = None  # Previous purchase IDs
    recommendation_type: str = "general"  # upsell, cross_sell, alternative
    limit: int = 3
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()
        if self.current_cart is None:
            self.current_cart = []
        if self.purchase_history is None:
            self.purchase_history = []


@dataclass
class OrderEvent:
    """Data for order events."""
    order_id: str
    customer_id: str
    customer_phone: str
    items: list[dict]
    total: float
    currency: str = "BRL"
    status: str = "pending"
    brand: Optional[str] = None  # Primary brand
    ai_seller_id: str = ""
    human_seller_id: str = ""
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()


@dataclass
class EscalationEvent:
    """Data for sales escalation events."""
    conversation_id: str
    customer_id: str
    customer_phone: str
    reason: str
    priority: str = "medium"  # low, medium, high, urgent
    brand: str = ""
    ai_seller_id: str = ""
    context_summary: str = ""
    cart_value: float = 0.0
    customer_sentiment: float = 0.5
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()


@dataclass
class SalesInsightEvent:
    """Data for sales insight events."""
    conversation_id: str
    customer_phone: str
    insight_type: str  # sentiment, purchase_intent, objection, opportunity
    summary: str
    suggested_action: str
    priority: str = "medium"
    brand: str = ""
    data: dict = None
    timestamp: str = ""
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()
        if self.data is None:
            self.data = {}


# =============================================================================
# Event Publisher
# =============================================================================

class EventPublisher:
    """
    Publishes CloudEvents to Knative broker.
    
    Uses HTTP POST to the broker ingress endpoint.
    """
    
    def __init__(
        self,
        broker_url: str = None,
        source: str = "/agent-store-multibrands/unknown"
    ):
        self.broker_url = broker_url or os.getenv(
            "K_SINK",
            os.getenv(
                "KNATIVE_BROKER_URL",
                "http://store-broker-broker-ingress.agent-store-multibrands.svc.cluster.local"
            )
        )
        self.source = source
        self._client: Optional[httpx.AsyncClient] = None
        
        logger.info(
            "event_publisher_initialized",
            broker_url=self.broker_url,
            source=self.source,
        )
    
    async def _get_client(self) -> httpx.AsyncClient:
        """Get or create HTTP client."""
        if self._client is None or self._client.is_closed:
            self._client = httpx.AsyncClient(timeout=10.0)
        return self._client
    
    async def publish(
        self,
        event_type: EventType,
        data: Any,
        subject: Optional[str] = None,
    ) -> bool:
        """
        Publish a CloudEvent to the broker.
        
        Args:
            event_type: The event type to publish
            data: Event data (dataclass or dict)
            subject: Optional subject for routing
            
        Returns:
            True if successful, False otherwise.
        """
        try:
            # Convert dataclass to dict if needed
            if hasattr(data, '__dataclass_fields__'):
                event_data = asdict(data)
            else:
                event_data = data
            
            event = CloudEvent({
                "type": event_type.value,
                "source": self.source,
                "subject": subject,
            }, event_data)
            
            client = await self._get_client()
            headers, body = to_structured(event)
            
            response = await client.post(
                self.broker_url,
                headers=headers,
                content=body,
            )
            
            if response.status_code in (200, 202, 204):
                EVENTS_EMITTED.labels(
                    event_type=event_type.value,
                    source=self.source,
                    status="success"
                ).inc()
                logger.debug(
                    "event_published",
                    event_type=event_type.value,
                    subject=subject,
                )
                return True
            else:
                EVENTS_EMITTED.labels(
                    event_type=event_type.value,
                    source=self.source,
                    status="error"
                ).inc()
                logger.warning(
                    "event_publish_failed",
                    event_type=event_type.value,
                    status_code=response.status_code,
                )
                return False
                
        except Exception as e:
            EVENTS_EMITTED.labels(
                event_type=event_type.value,
                source=self.source,
                status="error"
            ).inc()
            logger.error("event_publish_error", error=str(e))
            return False
    
    # ─────────────────────────────────────────────────────────────────────────
    # Convenience Methods
    # ─────────────────────────────────────────────────────────────────────────
    
    async def emit_whatsapp_received(self, data: WhatsAppMessageEvent) -> bool:
        """Emit WhatsApp message received event."""
        return await self.publish(
            EventType.WHATSAPP_MESSAGE_RECEIVED,
            data,
            subject=data.phone_from,
        )
    
    async def emit_whatsapp_send(self, data: WhatsAppMessageEvent) -> bool:
        """Emit WhatsApp send message event."""
        return await self.publish(
            EventType.WHATSAPP_MESSAGE_SEND,
            data,
            subject=data.phone_to,
        )
    
    async def emit_chat_message(self, data: ChatMessageEvent) -> bool:
        """Emit new chat message event."""
        return await self.publish(
            EventType.CHAT_MESSAGE_NEW,
            data,
            subject=f"{data.brand}/{data.customer_phone}",
        )
    
    async def emit_chat_response(self, data: ChatResponseEvent) -> bool:
        """Emit AI seller response event."""
        return await self.publish(
            EventType.CHAT_RESPONSE,
            data,
            subject=data.customer_phone,
        )
    
    async def emit_product_query(self, data: ProductQueryEvent) -> bool:
        """Emit product query event."""
        return await self.publish(
            EventType.PRODUCT_QUERY,
            data,
            subject=data.conversation_id,
        )
    
    async def emit_order_created(self, data: OrderEvent) -> bool:
        """Emit order created event."""
        return await self.publish(
            EventType.ORDER_CREATED,
            data,
            subject=data.order_id,
        )
    
    async def emit_escalation(self, data: EscalationEvent) -> bool:
        """Emit sales escalation event."""
        return await self.publish(
            EventType.SALES_ESCALATE,
            data,
            subject=f"{data.brand}/{data.priority}",
        )
    
    async def emit_sales_insight(self, data: SalesInsightEvent) -> bool:
        """Emit sales insight event."""
        return await self.publish(
            EventType.SALES_INSIGHT,
            data,
            subject=data.conversation_id,
        )
    
    async def close(self):
        """Close HTTP client."""
        if self._client and not self._client.is_closed:
            await self._client.aclose()


# =============================================================================
# Event Subscriber / Handler
# =============================================================================

EventHandler = Callable[[CloudEvent], Awaitable[None]]


class EventSubscriber:
    """
    Handles incoming CloudEvents from Knative triggers.
    """
    
    def __init__(self):
        self._handlers: dict[str, list[EventHandler]] = {}
    
    def register(self, event_type: str | EventType, handler: EventHandler):
        """Register a handler for an event type."""
        event_type_str = event_type.value if isinstance(event_type, EventType) else event_type
        if event_type_str not in self._handlers:
            self._handlers[event_type_str] = []
        self._handlers[event_type_str].append(handler)
        logger.info("event_handler_registered", event_type=event_type_str)
    
    async def handle(self, event: CloudEvent) -> bool:
        """
        Handle an incoming CloudEvent.
        
        Returns True if handled successfully.
        """
        event_type = event["type"]
        handlers = self._handlers.get(event_type, [])
        
        if not handlers:
            EVENTS_RECEIVED.labels(event_type=event_type, status="no_handler").inc()
            logger.debug("event_no_handler", event_type=event_type)
            return False
        
        try:
            for handler in handlers:
                await handler(event)
            
            EVENTS_RECEIVED.labels(event_type=event_type, status="success").inc()
            logger.info("event_handled", event_type=event_type)
            return True
            
        except Exception as e:
            EVENTS_RECEIVED.labels(event_type=event_type, status="error").inc()
            logger.error("event_handler_error", event_type=event_type, error=str(e))
            return False


# =============================================================================
# Global Instances & Initialization
# =============================================================================

publisher: Optional[EventPublisher] = None
subscriber: Optional[EventSubscriber] = None


def init_events(source: str = "/agent-store-multibrands/unknown") -> tuple[EventPublisher, EventSubscriber]:
    """Initialize event publisher and subscriber."""
    global publisher, subscriber
    
    publisher = EventPublisher(source=source)
    subscriber = EventSubscriber()
    
    logger.info(
        "events_initialized",
        broker_url=publisher.broker_url,
        source=source,
    )
    
    return publisher, subscriber


async def shutdown_events():
    """Shutdown event resources."""
    global publisher
    if publisher:
        await publisher.close()
        publisher = None
