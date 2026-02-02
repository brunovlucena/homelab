"""
Order Processor handler.

Handles order lifecycle: creation, payment, fulfillment, and tracking.
"""
import os
from typing import Optional, List
from decimal import Decimal
from datetime import datetime, timezone
from uuid import uuid4

import structlog
from cloudevents.http import CloudEvent

from shared.types import Order, OrderItem, OrderStatus, Brand
from shared.events import (
    EventType,
    EventPublisher,
    EventSubscriber,
    OrderEvent,
)
from shared.metrics import (
    ORDERS_CREATED,
    ORDERS_COMPLETED,
    ORDERS_CANCELLED,
    SALES_AMOUNT,
    AVERAGE_ORDER_VALUE,
)

logger = structlog.get_logger()


class OrderProcessor:
    """
    Order processing and management.
    
    Handles:
    - Order creation and validation
    - Payment processing integration
    - Order status updates
    - Shipping and tracking
    """
    
    def __init__(
        self,
        event_publisher: Optional[EventPublisher] = None,
        event_subscriber: Optional[EventSubscriber] = None,
    ):
        self.event_publisher = event_publisher
        self.event_subscriber = event_subscriber
        
        # Orders storage (use database in production)
        self._orders: dict[str, Order] = {}
        self._orders_by_customer: dict[str, List[str]] = {}
        
        # Stats
        self._total_sales: dict[str, Decimal] = {}  # by brand
        self._order_count: dict[str, int] = {}  # by brand
        
        logger.info("order_processor_initialized")
    
    # =========================================================================
    # Order Creation
    # =========================================================================
    
    def create_order(
        self,
        customer_id: str,
        customer_phone: str,
        items: List[dict],
        shipping_address: Optional[dict] = None,
        ai_seller_id: str = "",
        human_seller_id: str = "",
    ) -> Order:
        """
        Create a new order.
        
        Args:
            items: List of {"product_id", "product_name", "quantity", "unit_price", "brand"}
        """
        order_id = f"ORD-{datetime.now().strftime('%Y%m%d')}-{uuid4().hex[:8].upper()}"
        
        # Convert items to OrderItem objects
        order_items = []
        primary_brand = None
        
        for item in items:
            brand = Brand(item.get("brand", "fashion"))
            if primary_brand is None:
                primary_brand = brand
            
            unit_price = Decimal(str(item.get("unit_price", 0)))
            quantity = item.get("quantity", 1)
            
            order_items.append(OrderItem(
                product_id=item.get("product_id", ""),
                product_name=item.get("product_name", ""),
                quantity=quantity,
                unit_price=unit_price,
                total_price=unit_price * quantity,
                brand=brand,
                attributes=item.get("attributes", {}),
            ))
        
        # Calculate shipping (simplified)
        subtotal = sum(item.total_price for item in order_items)
        shipping = Decimal("15.00") if subtotal < Decimal("200.00") else Decimal("0.00")
        
        order = Order(
            id=order_id,
            customer_id=customer_id,
            customer_phone=customer_phone,
            items=order_items,
            subtotal=subtotal,
            shipping=shipping,
            total=subtotal + shipping,
            shipping_address=shipping_address or {},
            ai_seller_id=ai_seller_id,
            human_seller_id=human_seller_id,
        )
        
        # Store order
        self._orders[order_id] = order
        
        if customer_id not in self._orders_by_customer:
            self._orders_by_customer[customer_id] = []
        self._orders_by_customer[customer_id].append(order_id)
        
        # Update metrics
        seller_type = "human" if human_seller_id else "ai"
        brand_str = primary_brand.value if primary_brand else "unknown"
        
        ORDERS_CREATED.labels(brand=brand_str, seller_type=seller_type).inc()
        
        logger.info(
            "order_created",
            order_id=order_id,
            customer_id=customer_id,
            item_count=len(order_items),
            total=float(order.total),
        )
        
        return order
    
    # =========================================================================
    # Order Status Management
    # =========================================================================
    
    def update_status(
        self,
        order_id: str,
        new_status: OrderStatus,
        tracking_code: Optional[str] = None,
        notes: Optional[str] = None,
    ) -> Optional[Order]:
        """Update order status."""
        if order_id not in self._orders:
            return None
        
        order = self._orders[order_id]
        old_status = order.status
        order.status = new_status
        order.updated_at = datetime.now(timezone.utc).isoformat()
        
        if tracking_code:
            order.tracking_code = tracking_code
        if notes:
            order.notes = notes
        
        # Track metrics based on status change
        brand = order.items[0].brand.value if order.items else "unknown"
        seller_type = "human" if order.human_seller_id else "ai"
        
        if new_status == OrderStatus.DELIVERED:
            ORDERS_COMPLETED.labels(brand=brand, seller_type=seller_type).inc()
            SALES_AMOUNT.labels(
                brand=brand,
                currency=order.currency,
                seller_type=seller_type
            ).inc(float(order.total))
            
            # Update brand stats
            if brand not in self._total_sales:
                self._total_sales[brand] = Decimal("0")
                self._order_count[brand] = 0
            
            self._total_sales[brand] += order.total
            self._order_count[brand] += 1
            
            avg_value = self._total_sales[brand] / self._order_count[brand]
            AVERAGE_ORDER_VALUE.labels(brand=brand, currency=order.currency).set(float(avg_value))
            
        elif new_status == OrderStatus.CANCELLED:
            ORDERS_CANCELLED.labels(brand=brand, reason="customer").inc()
        
        logger.info(
            "order_status_updated",
            order_id=order_id,
            old_status=old_status.value,
            new_status=new_status.value,
        )
        
        return order
    
    def confirm_order(self, order_id: str, payment_id: str = "") -> Optional[Order]:
        """Confirm order after payment."""
        order = self._orders.get(order_id)
        if order:
            order.payment_id = payment_id
            order.payment_method = "pix"  # Default for Brazil
        return self.update_status(order_id, OrderStatus.CONFIRMED)
    
    def ship_order(self, order_id: str, tracking_code: str) -> Optional[Order]:
        """Mark order as shipped with tracking code."""
        return self.update_status(
            order_id,
            OrderStatus.SHIPPED,
            tracking_code=tracking_code,
        )
    
    def deliver_order(self, order_id: str) -> Optional[Order]:
        """Mark order as delivered."""
        return self.update_status(order_id, OrderStatus.DELIVERED)
    
    def cancel_order(self, order_id: str, reason: str = "") -> Optional[Order]:
        """Cancel an order."""
        return self.update_status(order_id, OrderStatus.CANCELLED, notes=reason)
    
    # =========================================================================
    # Query Methods
    # =========================================================================
    
    def get_order(self, order_id: str) -> Optional[Order]:
        """Get order by ID."""
        return self._orders.get(order_id)
    
    def get_customer_orders(
        self,
        customer_id: str,
        status: Optional[OrderStatus] = None,
        limit: int = 10,
    ) -> List[Order]:
        """Get orders for a customer."""
        order_ids = self._orders_by_customer.get(customer_id, [])
        orders = []
        
        for oid in reversed(order_ids):  # Most recent first
            order = self._orders.get(oid)
            if order:
                if status is None or order.status == status:
                    orders.append(order)
                    if len(orders) >= limit:
                        break
        
        return orders
    
    def get_pending_orders(self, limit: int = 50) -> List[Order]:
        """Get all pending orders (awaiting processing)."""
        pending = []
        for order in self._orders.values():
            if order.status in (OrderStatus.PENDING, OrderStatus.CONFIRMED):
                pending.append(order)
        
        return sorted(pending, key=lambda o: o.created_at)[:limit]
    
    # =========================================================================
    # Event Handlers
    # =========================================================================
    
    async def handle_order_create(self, event: CloudEvent):
        """Handle order creation event."""
        data = event.data or {}
        
        order = self.create_order(
            customer_id=data.get("customer_id", ""),
            customer_phone=data.get("customer_phone", ""),
            items=data.get("items", []),
            shipping_address=data.get("shipping_address"),
            ai_seller_id=data.get("ai_seller_id", ""),
            human_seller_id=data.get("human_seller_id", ""),
        )
        
        # Emit order created event
        if self.event_publisher:
            primary_brand = order.items[0].brand.value if order.items else "unknown"
            
            await self.event_publisher.emit_order_created(
                OrderEvent(
                    order_id=order.id,
                    customer_id=order.customer_id,
                    customer_phone=order.customer_phone,
                    items=[item.to_dict() for item in order.items],
                    total=float(order.total),
                    currency=order.currency,
                    status=order.status.value,
                    brand=primary_brand,
                    ai_seller_id=order.ai_seller_id,
                    human_seller_id=order.human_seller_id,
                )
            )
        
        return order
    
    def setup_event_handlers(self):
        """Register event handlers."""
        if self.event_subscriber:
            self.event_subscriber.register(
                EventType.ORDER_CREATE,
                self.handle_order_create
            )
            logger.info("order_processor_event_handlers_registered")
    
    # =========================================================================
    # Stats
    # =========================================================================
    
    def get_stats(self) -> dict:
        """Get order statistics."""
        total_orders = len(self._orders)
        
        by_status = {}
        for order in self._orders.values():
            status = order.status.value
            by_status[status] = by_status.get(status, 0) + 1
        
        return {
            "total_orders": total_orders,
            "by_status": by_status,
            "total_sales_by_brand": {k: float(v) for k, v in self._total_sales.items()},
            "order_count_by_brand": self._order_count,
        }
