"""
Prometheus metrics for Agent Store MultiBrands.

Provides observability into sales, conversations, and system performance.
"""
import os
from prometheus_client import Counter, Gauge, Histogram, Info

# =============================================================================
# Build Information
# =============================================================================

BUILD_INFO = Info(
    "agent_store_multibrands_build",
    "Build information for the store agent"
)


def init_metrics(version: str, commit: str = "unknown", agent_name: str = "unknown"):
    """Initialize build info metrics."""
    BUILD_INFO.info({
        "version": version,
        "commit": commit,
        "agent_name": agent_name,
    })


def init_build_info(version: str, commit: str = "unknown"):
    """Initialize build info metric (alias for init_metrics)."""
    BUILD_INFO.info({
        "version": version,
        "commit": commit,
    })


# =============================================================================
# Message Metrics
# =============================================================================

MESSAGES_RECEIVED = Counter(
    "store_messages_received_total",
    "Total messages received from customers",
    ["brand", "channel", "type"]  # channel: whatsapp, type: text/image/etc
)

MESSAGES_SENT = Counter(
    "store_messages_sent_total",
    "Total messages sent to customers",
    ["brand", "sender_type", "type"]  # sender_type: ai/human
)

MESSAGE_LATENCY = Histogram(
    "store_message_latency_seconds",
    "Time from customer message to response",
    ["brand", "sender_type"],
    buckets=[0.5, 1, 2, 5, 10, 30, 60, 120]
)


# =============================================================================
# Conversation Metrics
# =============================================================================

CONVERSATIONS_ACTIVE = Gauge(
    "store_conversations_active",
    "Number of active conversations",
    ["brand", "state"]  # state: browsing, checkout, escalated, etc
)

CONVERSATIONS_TOTAL = Counter(
    "store_conversations_total",
    "Total conversations started",
    ["brand"]
)

CONVERSATION_DURATION = Histogram(
    "store_conversation_duration_seconds",
    "Duration of conversations",
    ["brand", "outcome"],  # outcome: purchase, abandoned, escalated
    buckets=[60, 300, 600, 1800, 3600, 7200]
)


# =============================================================================
# Sales Metrics
# =============================================================================

ORDERS_CREATED = Counter(
    "store_orders_created_total",
    "Total orders created",
    ["brand", "seller_type"]  # seller_type: ai/human
)

ORDERS_COMPLETED = Counter(
    "store_orders_completed_total",
    "Total orders successfully completed",
    ["brand", "seller_type"]
)

ORDERS_CANCELLED = Counter(
    "store_orders_cancelled_total",
    "Total orders cancelled",
    ["brand", "reason"]
)

SALES_AMOUNT = Counter(
    "store_sales_amount_total",
    "Total sales amount",
    ["brand", "currency", "seller_type"]
)

AVERAGE_ORDER_VALUE = Gauge(
    "store_average_order_value",
    "Average order value",
    ["brand", "currency"]
)

CART_ABANDONMENT = Counter(
    "store_cart_abandonment_total",
    "Abandoned carts",
    ["brand"]
)


# =============================================================================
# AI Performance Metrics
# =============================================================================

AI_RESPONSE_TIME = Histogram(
    "store_ai_response_time_seconds",
    "Time for AI to generate response",
    ["brand", "model"],
    buckets=[0.1, 0.5, 1, 2, 5, 10, 30]
)

AI_TOKENS_USED = Counter(
    "store_ai_tokens_used_total",
    "Total tokens used for AI inference",
    ["brand", "model", "type"]  # type: input/output
)

AI_RECOMMENDATIONS_MADE = Counter(
    "store_ai_recommendations_made_total",
    "Product recommendations made by AI",
    ["brand", "type"]  # type: upsell/cross_sell/alternative
)

AI_RECOMMENDATIONS_ACCEPTED = Counter(
    "store_ai_recommendations_accepted_total",
    "Product recommendations accepted by customers",
    ["brand", "type"]
)


# =============================================================================
# Escalation Metrics
# =============================================================================

ESCALATIONS_TOTAL = Counter(
    "store_escalations_total",
    "Total escalations to human sellers",
    ["brand", "reason", "priority"]
)

ESCALATION_WAIT_TIME = Histogram(
    "store_escalation_wait_time_seconds",
    "Time customer waits for human seller",
    ["brand", "priority"],
    buckets=[30, 60, 120, 300, 600, 1800]
)

ESCALATION_RESOLUTION_TIME = Histogram(
    "store_escalation_resolution_time_seconds",
    "Time to resolve escalated conversations",
    ["brand", "outcome"],
    buckets=[60, 300, 600, 1800, 3600]
)


# =============================================================================
# Product Metrics
# =============================================================================

PRODUCT_QUERIES = Counter(
    "store_product_queries_total",
    "Product search queries",
    ["brand", "query_type"]  # query_type: search/recommend/view
)

PRODUCT_VIEWS = Counter(
    "store_product_views_total",
    "Product detail views",
    ["brand", "category"]
)

PRODUCT_STOCK_ALERTS = Counter(
    "store_product_stock_alerts_total",
    "Low stock alerts",
    ["brand", "category"]
)


# =============================================================================
# WhatsApp Metrics
# =============================================================================

WHATSAPP_MESSAGES_RECEIVED = Counter(
    "store_whatsapp_messages_received_total",
    "WhatsApp messages received",
    ["message_type"]  # text, image, audio, etc
)

WHATSAPP_MESSAGES_SENT = Counter(
    "store_whatsapp_messages_sent_total",
    "WhatsApp messages sent",
    ["message_type"]
)

WHATSAPP_API_LATENCY = Histogram(
    "store_whatsapp_api_latency_seconds",
    "WhatsApp API call latency",
    ["operation"],  # send, webhook
    buckets=[0.1, 0.5, 1, 2, 5, 10]
)

WHATSAPP_API_ERRORS = Counter(
    "store_whatsapp_api_errors_total",
    "WhatsApp API errors",
    ["operation", "error_type"]
)


# =============================================================================
# Sentiment Metrics
# =============================================================================

CUSTOMER_SENTIMENT = Histogram(
    "store_customer_sentiment",
    "Customer sentiment score distribution",
    ["brand"],
    buckets=[0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0]
)

NEGATIVE_SENTIMENT_ALERTS = Counter(
    "store_negative_sentiment_alerts_total",
    "Negative sentiment detected",
    ["brand"]
)
