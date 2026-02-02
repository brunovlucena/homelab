"""
Unit tests for Order processing.
"""
import pytest
from decimal import Decimal

import sys
sys.path.insert(0, 'src')

from shared.types import Order, OrderItem, OrderStatus, Brand


class TestOrderItem:
    """Test OrderItem model."""
    
    def test_order_item_creation(self):
        """Test creating an order item."""
        item = OrderItem(
            product_id="fashion-001",
            product_name="Vestido Floral",
            quantity=2,
            unit_price=Decimal("199.90"),
            total_price=Decimal("399.80"),
            brand=Brand.FASHION,
        )
        
        assert item.product_id == "fashion-001"
        assert item.quantity == 2
        assert item.total_price == Decimal("399.80")
    
    def test_order_item_to_dict(self):
        """Test order item serialization."""
        item = OrderItem(
            product_id="tech-001",
            product_name="Fone Bluetooth",
            quantity=1,
            unit_price=Decimal("599.90"),
            total_price=Decimal("599.90"),
            brand=Brand.TECH,
        )
        
        data = item.to_dict()
        assert data["product_id"] == "tech-001"
        assert data["brand"] == "tech"
        assert data["total_price"] == 599.90


class TestOrder:
    """Test Order model."""
    
    def test_order_creation(self, sample_order_items):
        """Test creating an order."""
        items = [
            OrderItem(
                product_id=item["product_id"],
                product_name=item["product_name"],
                quantity=item["quantity"],
                unit_price=Decimal(str(item["unit_price"])),
                total_price=Decimal(str(item["unit_price"])) * item["quantity"],
                brand=Brand(item["brand"]),
            )
            for item in sample_order_items
        ]
        
        order = Order(
            id="ORD-20241210-ABC123",
            customer_id="cust-001",
            customer_phone="5511999999999",
            items=items,
        )
        
        assert order.id == "ORD-20241210-ABC123"
        assert len(order.items) == 2
        assert order.status == OrderStatus.PENDING
    
    def test_order_recalculate_totals(self, sample_order_items):
        """Test order total recalculation."""
        items = [
            OrderItem(
                product_id=item["product_id"],
                product_name=item["product_name"],
                quantity=item["quantity"],
                unit_price=Decimal(str(item["unit_price"])),
                total_price=Decimal(str(item["unit_price"])) * item["quantity"],
                brand=Brand(item["brand"]),
            )
            for item in sample_order_items
        ]
        
        order = Order(
            id="ORD-TEST",
            customer_id="cust-001",
            customer_phone="5511999999999",
            items=items,
        )
        
        # 199.90 + (289.90 * 2) = 779.70
        expected_subtotal = Decimal("199.90") + (Decimal("289.90") * 2)
        assert order.subtotal == expected_subtotal
    
    def test_order_status_transitions(self):
        """Test order status values."""
        assert OrderStatus.PENDING.value == "pending"
        assert OrderStatus.CONFIRMED.value == "confirmed"
        assert OrderStatus.SHIPPED.value == "shipped"
        assert OrderStatus.DELIVERED.value == "delivered"
        assert OrderStatus.CANCELLED.value == "cancelled"
    
    def test_order_add_item(self):
        """Test adding item to order."""
        order = Order(
            id="ORD-TEST",
            customer_id="cust-001",
            customer_phone="5511999999999",
        )
        
        item = OrderItem(
            product_id="gaming-001",
            product_name="Mouse Gamer",
            quantity=1,
            unit_price=Decimal("249.90"),
            total_price=Decimal("249.90"),
            brand=Brand.GAMING,
        )
        
        order.add_item(item)
        
        assert len(order.items) == 1
        assert order.subtotal == Decimal("249.90")
