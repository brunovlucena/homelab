"""
Order Processor Agent - FastAPI entry point.

Handles order creation, payment, and fulfillment.
"""
import os
from contextlib import asynccontextmanager
from typing import Optional, List

from fastapi import FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import Response
from pydantic import BaseModel, Field
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
from cloudevents.http import from_http
import structlog

from .handler import OrderProcessor
from shared.types import OrderStatus
from shared.metrics import init_metrics
from shared.events import init_events, shutdown_events

logger = structlog.get_logger()

# Global processor instance
processor: OrderProcessor = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global processor
    
    # Initialize CloudEvents
    pub, sub = init_events(source="/agent-store-multibrands/order-processor")
    
    # Initialize processor
    processor = OrderProcessor(
        event_publisher=pub,
        event_subscriber=sub,
    )
    processor.setup_event_handlers()
    
    # Initialize metrics
    version = os.getenv("VERSION", "0.1.0")
    commit = os.getenv("GIT_COMMIT", "unknown")
    init_metrics(version, commit, "order-processor")
    
    logger.info("order_processor_started", version=version)
    
    yield
    
    # Cleanup
    await shutdown_events()
    logger.info("order_processor_shutdown")


app = FastAPI(
    title="Order Processor",
    description="Order management and fulfillment service",
    version="0.1.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# =============================================================================
# Request/Response Models
# =============================================================================

class OrderItemRequest(BaseModel):
    """Order item in creation request."""
    product_id: str
    product_name: str
    quantity: int = Field(default=1, ge=1)
    unit_price: float = Field(ge=0)
    brand: str = "fashion"
    attributes: dict = {}


class CreateOrderRequest(BaseModel):
    """Order creation request."""
    customer_id: str
    customer_phone: str
    items: List[OrderItemRequest]
    shipping_address: Optional[dict] = None
    ai_seller_id: str = ""
    human_seller_id: str = ""


class UpdateStatusRequest(BaseModel):
    """Order status update request."""
    order_id: str
    status: str  # pending, confirmed, processing, shipped, delivered, cancelled
    tracking_code: Optional[str] = None
    notes: Optional[str] = None


class ConfirmOrderRequest(BaseModel):
    """Order confirmation request."""
    order_id: str
    payment_id: str = ""


class ShipOrderRequest(BaseModel):
    """Ship order request."""
    order_id: str
    tracking_code: str


# =============================================================================
# Health Endpoints
# =============================================================================

@app.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "healthy"}


@app.get("/ready")
async def ready():
    """Readiness check endpoint."""
    if processor is None:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    return {"status": "ready"}


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST
    )


# =============================================================================
# Order Endpoints
# =============================================================================

@app.post("/orders")
async def create_order(request: CreateOrderRequest):
    """Create a new order."""
    if processor is None:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    items = [
        {
            "product_id": item.product_id,
            "product_name": item.product_name,
            "quantity": item.quantity,
            "unit_price": item.unit_price,
            "brand": item.brand,
            "attributes": item.attributes,
        }
        for item in request.items
    ]
    
    order = processor.create_order(
        customer_id=request.customer_id,
        customer_phone=request.customer_phone,
        items=items,
        shipping_address=request.shipping_address,
        ai_seller_id=request.ai_seller_id,
        human_seller_id=request.human_seller_id,
    )
    
    return order.to_dict()


@app.get("/orders/{order_id}")
async def get_order(order_id: str):
    """Get order by ID."""
    if processor is None:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    order = processor.get_order(order_id)
    
    if order is None:
        raise HTTPException(status_code=404, detail="Order not found")
    
    return order.to_dict()


@app.get("/orders/customer/{customer_id}")
async def get_customer_orders(customer_id: str, status: Optional[str] = None, limit: int = 10):
    """Get orders for a customer."""
    if processor is None:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    status_enum = OrderStatus(status) if status else None
    
    orders = processor.get_customer_orders(
        customer_id=customer_id,
        status=status_enum,
        limit=limit,
    )
    
    return {
        "count": len(orders),
        "orders": [o.to_dict() for o in orders],
    }


@app.get("/orders/pending")
async def get_pending_orders(limit: int = 50):
    """Get all pending orders."""
    if processor is None:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    orders = processor.get_pending_orders(limit=limit)
    
    return {
        "count": len(orders),
        "orders": [o.to_dict() for o in orders],
    }


# =============================================================================
# Status Update Endpoints
# =============================================================================

@app.post("/orders/confirm")
async def confirm_order(request: ConfirmOrderRequest):
    """Confirm order after payment."""
    if processor is None:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    order = processor.confirm_order(
        order_id=request.order_id,
        payment_id=request.payment_id,
    )
    
    if order is None:
        raise HTTPException(status_code=404, detail="Order not found")
    
    return order.to_dict()


@app.post("/orders/ship")
async def ship_order(request: ShipOrderRequest):
    """Mark order as shipped."""
    if processor is None:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    order = processor.ship_order(
        order_id=request.order_id,
        tracking_code=request.tracking_code,
    )
    
    if order is None:
        raise HTTPException(status_code=404, detail="Order not found")
    
    return order.to_dict()


@app.post("/orders/{order_id}/deliver")
async def deliver_order(order_id: str):
    """Mark order as delivered."""
    if processor is None:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    order = processor.deliver_order(order_id)
    
    if order is None:
        raise HTTPException(status_code=404, detail="Order not found")
    
    return order.to_dict()


@app.post("/orders/{order_id}/cancel")
async def cancel_order(order_id: str, reason: str = ""):
    """Cancel an order."""
    if processor is None:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    order = processor.cancel_order(order_id, reason=reason)
    
    if order is None:
        raise HTTPException(status_code=404, detail="Order not found")
    
    return order.to_dict()


@app.post("/orders/status")
async def update_status(request: UpdateStatusRequest):
    """Update order status."""
    if processor is None:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    try:
        status = OrderStatus(request.status)
    except ValueError:
        raise HTTPException(status_code=400, detail=f"Invalid status: {request.status}")
    
    order = processor.update_status(
        order_id=request.order_id,
        new_status=status,
        tracking_code=request.tracking_code,
        notes=request.notes,
    )
    
    if order is None:
        raise HTTPException(status_code=404, detail="Order not found")
    
    return order.to_dict()


# =============================================================================
# Stats Endpoint
# =============================================================================

@app.get("/stats")
async def get_stats():
    """Get order statistics."""
    if processor is None:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    return processor.get_stats()


# =============================================================================
# CloudEvents Endpoint
# =============================================================================

@app.post("/events")
async def receive_cloudevent(request: Request):
    """
    Receive CloudEvents from Knative triggers.
    
    Subscribed events:
    - store.order.create: Create new order
    """
    try:
        headers = dict(request.headers)
        body = await request.body()
        event = from_http(headers, body)
        
        logger.info(
            "cloudevent_received",
            event_type=event["type"],
            source=event["source"],
        )
        
        if processor and processor.event_subscriber:
            handled = await processor.event_subscriber.handle(event)
            return {"status": "handled" if handled else "no_handler"}
        
        return {"status": "subscriber_not_initialized"}
        
    except Exception as e:
        logger.error("cloudevent_processing_failed", error=str(e))
        raise HTTPException(status_code=400, detail=f"Invalid CloudEvent: {str(e)}")


# =============================================================================
# Root Endpoint
# =============================================================================

@app.get("/")
async def root():
    """Root endpoint with service info."""
    return {
        "service": "order-processor",
        "description": "Order management and fulfillment service",
        "version": os.getenv("VERSION", "0.1.0"),
        "endpoints": {
            "create": "POST /orders",
            "get": "GET /orders/{order_id}",
            "customer_orders": "GET /orders/customer/{customer_id}",
            "pending": "GET /orders/pending",
            "confirm": "POST /orders/confirm",
            "ship": "POST /orders/ship",
            "deliver": "POST /orders/{order_id}/deliver",
            "cancel": "POST /orders/{order_id}/cancel",
            "status": "POST /orders/status",
            "stats": "GET /stats",
            "events": "POST /events (CloudEvents)",
            "health": "GET /health",
            "ready": "GET /ready",
            "metrics": "GET /metrics",
        },
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
