"""
Shared modules for Agent Store MultiBrands.

This package contains common types, events, and metrics used by all agents.
"""
from .types import (
    Brand,
    Product,
    Customer,
    Order,
    OrderItem,
    OrderStatus,
    Message,
    MessageRole,
    Conversation,
    ConversationState,
    WhatsAppMessage,
    WhatsAppMessageType,
    SalesRecommendation,
    EscalationReason,
)
from .events import (
    EventType,
    EventPublisher,
    EventSubscriber,
    init_events,
    shutdown_events,
)
from .metrics import (
    init_metrics,
    MESSAGES_RECEIVED,
    MESSAGES_SENT,
    ORDERS_CREATED,
    SALES_AMOUNT,
    AI_RESPONSE_TIME,
    CONVERSATIONS_ACTIVE,
    ESCALATIONS_TOTAL,
)

__all__ = [
    # Types
    "Brand",
    "Product",
    "Customer",
    "Order",
    "OrderItem",
    "OrderStatus",
    "Message",
    "MessageRole",
    "Conversation",
    "ConversationState",
    "WhatsAppMessage",
    "WhatsAppMessageType",
    "SalesRecommendation",
    "EscalationReason",
    # Events
    "EventType",
    "EventPublisher",
    "EventSubscriber",
    "init_events",
    "shutdown_events",
    # Metrics
    "init_metrics",
    "MESSAGES_RECEIVED",
    "MESSAGES_SENT",
    "ORDERS_CREATED",
    "SALES_AMOUNT",
    "AI_RESPONSE_TIME",
    "CONVERSATIONS_ACTIVE",
    "ESCALATIONS_TOTAL",
]
