"""
Shared types for Agent Store MultiBrands.

Defines data models for products, orders, customers, and conversations.
"""
from enum import Enum
from dataclasses import dataclass, field
from typing import Optional, Any
from datetime import datetime, timezone
from decimal import Decimal


# =============================================================================
# Brand Definitions
# =============================================================================

class Brand(str, Enum):
    """Available brands in the store."""
    FASHION = "fashion"
    TECH = "tech"
    HOME = "home"
    BEAUTY = "beauty"
    GAMING = "gaming"
    
    @property
    def display_name(self) -> str:
        """Human-readable brand name."""
        return {
            Brand.FASHION: "Fashion & Style",
            Brand.TECH: "Tech & Electronics",
            Brand.HOME: "Home & Living",
            Brand.BEAUTY: "Beauty & Care",
            Brand.GAMING: "Gaming & Entertainment",
        }[self]
    
    @property
    def emoji(self) -> str:
        """Brand emoji for messages."""
        return {
            Brand.FASHION: "👗",
            Brand.TECH: "📱",
            Brand.HOME: "🏠",
            Brand.BEAUTY: "💄",
            Brand.GAMING: "🎮",
        }[self]


# =============================================================================
# Product Types
# =============================================================================

@dataclass
class Product:
    """Product in the catalog."""
    id: str
    name: str
    brand: Brand
    description: str
    price: Decimal
    currency: str = "BRL"
    images: list[str] = field(default_factory=list)
    category: str = ""
    subcategory: str = ""
    tags: list[str] = field(default_factory=list)
    stock: int = 0
    sku: str = ""
    attributes: dict[str, Any] = field(default_factory=dict)
    created_at: str = ""
    updated_at: str = ""
    
    def __post_init__(self):
        now = datetime.now(timezone.utc).isoformat()
        if not self.created_at:
            self.created_at = now
        if not self.updated_at:
            self.updated_at = now
    
    def to_dict(self) -> dict:
        return {
            "id": self.id,
            "name": self.name,
            "brand": self.brand.value,
            "description": self.description,
            "price": float(self.price),
            "currency": self.currency,
            "images": self.images,
            "category": self.category,
            "subcategory": self.subcategory,
            "tags": self.tags,
            "stock": self.stock,
            "sku": self.sku,
            "attributes": self.attributes,
        }
    
    def format_price(self) -> str:
        """Format price for display."""
        if self.currency == "BRL":
            return f"R$ {self.price:,.2f}"
        elif self.currency == "USD":
            return f"$ {self.price:,.2f}"
        return f"{self.currency} {self.price:,.2f}"


# =============================================================================
# Customer Types
# =============================================================================

@dataclass
class Customer:
    """Customer information."""
    id: str
    phone: str  # WhatsApp phone number
    name: str = ""
    email: str = ""
    preferred_brand: Optional[Brand] = None
    preferred_language: str = "pt-BR"
    tags: list[str] = field(default_factory=list)
    purchase_history: list[str] = field(default_factory=list)  # Order IDs
    total_spent: Decimal = Decimal("0.00")
    created_at: str = ""
    last_interaction: str = ""
    metadata: dict[str, Any] = field(default_factory=dict)
    
    def __post_init__(self):
        now = datetime.now(timezone.utc).isoformat()
        if not self.created_at:
            self.created_at = now
        if not self.last_interaction:
            self.last_interaction = now
    
    def to_dict(self) -> dict:
        return {
            "id": self.id,
            "phone": self.phone,
            "name": self.name,
            "email": self.email,
            "preferred_brand": self.preferred_brand.value if self.preferred_brand else None,
            "preferred_language": self.preferred_language,
            "tags": self.tags,
            "total_spent": float(self.total_spent),
        }


# =============================================================================
# Order Types
# =============================================================================

class OrderStatus(str, Enum):
    """Order status states."""
    PENDING = "pending"           # Just created, awaiting payment
    CONFIRMED = "confirmed"       # Payment confirmed
    PROCESSING = "processing"     # Being prepared
    SHIPPED = "shipped"           # Shipped to customer
    DELIVERED = "delivered"       # Delivered successfully
    CANCELLED = "cancelled"       # Cancelled
    REFUNDED = "refunded"         # Money refunded


@dataclass
class OrderItem:
    """Single item in an order."""
    product_id: str
    product_name: str
    quantity: int
    unit_price: Decimal
    total_price: Decimal
    brand: Brand
    attributes: dict[str, Any] = field(default_factory=dict)
    
    def to_dict(self) -> dict:
        return {
            "product_id": self.product_id,
            "product_name": self.product_name,
            "quantity": self.quantity,
            "unit_price": float(self.unit_price),
            "total_price": float(self.total_price),
            "brand": self.brand.value,
            "attributes": self.attributes,
        }


@dataclass
class Order:
    """Customer order."""
    id: str
    customer_id: str
    customer_phone: str
    items: list[OrderItem] = field(default_factory=list)
    subtotal: Decimal = Decimal("0.00")
    shipping: Decimal = Decimal("0.00")
    discount: Decimal = Decimal("0.00")
    total: Decimal = Decimal("0.00")
    currency: str = "BRL"
    status: OrderStatus = OrderStatus.PENDING
    shipping_address: dict[str, str] = field(default_factory=dict)
    payment_method: str = ""
    payment_id: str = ""
    tracking_code: str = ""
    notes: str = ""
    created_at: str = ""
    updated_at: str = ""
    ai_seller_id: str = ""  # Which AI seller created this order
    human_seller_id: str = ""  # Human seller if escalated
    
    def __post_init__(self):
        now = datetime.now(timezone.utc).isoformat()
        if not self.created_at:
            self.created_at = now
        if not self.updated_at:
            self.updated_at = now
        # Calculate total if not set
        if self.total == Decimal("0.00") and self.items:
            self.recalculate_totals()
    
    def recalculate_totals(self):
        """Recalculate order totals."""
        self.subtotal = sum(item.total_price for item in self.items)
        self.total = self.subtotal + self.shipping - self.discount
    
    def add_item(self, item: OrderItem):
        """Add an item to the order."""
        self.items.append(item)
        self.recalculate_totals()
        self.updated_at = datetime.now(timezone.utc).isoformat()
    
    def to_dict(self) -> dict:
        return {
            "id": self.id,
            "customer_id": self.customer_id,
            "customer_phone": self.customer_phone,
            "items": [item.to_dict() for item in self.items],
            "subtotal": float(self.subtotal),
            "shipping": float(self.shipping),
            "discount": float(self.discount),
            "total": float(self.total),
            "currency": self.currency,
            "status": self.status.value,
            "shipping_address": self.shipping_address,
            "tracking_code": self.tracking_code,
            "created_at": self.created_at,
        }


# =============================================================================
# Conversation Types
# =============================================================================

class MessageRole(str, Enum):
    """Message role in conversation."""
    CUSTOMER = "customer"
    AI_SELLER = "ai_seller"
    HUMAN_SELLER = "human_seller"
    SYSTEM = "system"


class ConversationState(str, Enum):
    """Conversation state machine."""
    NEW = "new"                     # New conversation
    BROWSING = "browsing"           # Customer browsing products
    INQUIRING = "inquiring"         # Asking about products
    SELECTING = "selecting"         # Selecting products to buy
    CHECKOUT = "checkout"           # In checkout process
    PAYMENT = "payment"             # Awaiting payment
    POST_SALE = "post_sale"         # After purchase (support)
    ESCALATED = "escalated"         # Escalated to human seller
    CLOSED = "closed"               # Conversation ended


@dataclass
class Message:
    """A single message in conversation."""
    id: str
    role: MessageRole
    content: str
    timestamp: str = ""
    metadata: dict[str, Any] = field(default_factory=dict)
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()
    
    def to_dict(self) -> dict:
        return {
            "id": self.id,
            "role": self.role.value,
            "content": self.content,
            "timestamp": self.timestamp,
            "metadata": self.metadata,
        }


@dataclass
class Conversation:
    """A conversation with a customer."""
    id: str
    customer_id: str
    customer_phone: str
    brand: Brand
    state: ConversationState = ConversationState.NEW
    messages: list[Message] = field(default_factory=list)
    current_order_id: Optional[str] = None
    cart_items: list[dict] = field(default_factory=list)
    ai_seller_id: str = ""
    human_seller_id: str = ""
    escalation_reason: str = ""
    sentiment_score: float = 0.5  # 0-1, neutral = 0.5
    intent: str = ""  # Current detected intent
    created_at: str = ""
    updated_at: str = ""
    closed_at: str = ""
    
    def __post_init__(self):
        now = datetime.now(timezone.utc).isoformat()
        if not self.created_at:
            self.created_at = now
        if not self.updated_at:
            self.updated_at = now
    
    def add_message(self, role: MessageRole, content: str, msg_id: str = "") -> Message:
        """Add a message to the conversation."""
        from uuid import uuid4
        msg = Message(
            id=msg_id or str(uuid4()),
            role=role,
            content=content,
        )
        self.messages.append(msg)
        self.updated_at = datetime.now(timezone.utc).isoformat()
        return msg
    
    def get_context(self, max_messages: int = 10) -> list[dict]:
        """Get conversation context for AI."""
        recent = self.messages[-max_messages:] if len(self.messages) > max_messages else self.messages
        return [m.to_dict() for m in recent]
    
    def to_dict(self) -> dict:
        return {
            "id": self.id,
            "customer_id": self.customer_id,
            "customer_phone": self.customer_phone,
            "brand": self.brand.value,
            "state": self.state.value,
            "message_count": len(self.messages),
            "cart_items": len(self.cart_items),
            "sentiment_score": self.sentiment_score,
            "created_at": self.created_at,
            "updated_at": self.updated_at,
        }


# =============================================================================
# WhatsApp Types
# =============================================================================

class WhatsAppMessageType(str, Enum):
    """WhatsApp message types."""
    TEXT = "text"
    IMAGE = "image"
    AUDIO = "audio"
    VIDEO = "video"
    DOCUMENT = "document"
    LOCATION = "location"
    STICKER = "sticker"
    INTERACTIVE = "interactive"  # Buttons, lists
    TEMPLATE = "template"        # Template messages


@dataclass
class WhatsAppMessage:
    """Incoming or outgoing WhatsApp message."""
    id: str
    phone_from: str
    phone_to: str
    type: WhatsAppMessageType
    content: str
    media_url: Optional[str] = None
    media_mime_type: Optional[str] = None
    timestamp: str = ""
    status: str = "received"  # received, sent, delivered, read
    context_message_id: Optional[str] = None  # For replies
    interactive_data: Optional[dict] = None  # Button/list selection
    
    def __post_init__(self):
        if not self.timestamp:
            self.timestamp = datetime.now(timezone.utc).isoformat()
    
    def to_dict(self) -> dict:
        return {
            "id": self.id,
            "phone_from": self.phone_from,
            "phone_to": self.phone_to,
            "type": self.type.value,
            "content": self.content,
            "media_url": self.media_url,
            "timestamp": self.timestamp,
            "status": self.status,
        }


# =============================================================================
# Sales & Escalation Types
# =============================================================================

class EscalationReason(str, Enum):
    """Reasons for escalating to human seller."""
    COMPLEX_QUERY = "complex_query"         # AI can't handle the query
    CUSTOMER_REQUEST = "customer_request"   # Customer asked for human
    HIGH_VALUE = "high_value"               # High value order
    COMPLAINT = "complaint"                 # Customer complaint
    TECHNICAL_ISSUE = "technical_issue"     # Product technical question
    PAYMENT_ISSUE = "payment_issue"         # Payment problems
    NEGATIVE_SENTIMENT = "negative_sentiment"  # Customer seems upset
    CUSTOM = "custom"                       # Other reasons


@dataclass
class SalesRecommendation:
    """AI-generated sales recommendation."""
    product_id: str
    product_name: str
    reason: str
    confidence: float  # 0-1
    recommendation_type: str  # upsell, cross_sell, alternative
    price_info: str = ""
    
    def to_dict(self) -> dict:
        return {
            "product_id": self.product_id,
            "product_name": self.product_name,
            "reason": self.reason,
            "confidence": self.confidence,
            "recommendation_type": self.recommendation_type,
            "price_info": self.price_info,
        }


@dataclass
class SalesInsight:
    """Insight for human sellers."""
    conversation_id: str
    customer_phone: str
    insight_type: str  # sentiment, purchase_intent, objection, opportunity
    summary: str
    suggested_action: str
    priority: str = "medium"  # low, medium, high, urgent
    created_at: str = ""
    
    def __post_init__(self):
        if not self.created_at:
            self.created_at = datetime.now(timezone.utc).isoformat()
    
    def to_dict(self) -> dict:
        return {
            "conversation_id": self.conversation_id,
            "customer_phone": self.customer_phone,
            "insight_type": self.insight_type,
            "summary": self.summary,
            "suggested_action": self.suggested_action,
            "priority": self.priority,
            "created_at": self.created_at,
        }
