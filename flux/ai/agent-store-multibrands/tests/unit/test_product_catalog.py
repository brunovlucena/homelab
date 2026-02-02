"""
Unit tests for Product Catalog handler.
"""
import pytest
from decimal import Decimal

import sys
sys.path.insert(0, 'src')

from shared.types import Brand


class TestProductCatalog:
    """Test product catalog functionality."""
    
    def test_brand_enum_values(self):
        """Test all brand enum values."""
        assert Brand.FASHION.value == "fashion"
        assert Brand.TECH.value == "tech"
        assert Brand.HOME.value == "home"
        assert Brand.BEAUTY.value == "beauty"
        assert Brand.GAMING.value == "gaming"
    
    def test_brand_display_names(self):
        """Test brand display names."""
        assert Brand.FASHION.display_name == "Fashion & Style"
        assert Brand.TECH.display_name == "Tech & Electronics"
        assert Brand.GAMING.display_name == "Gaming & Entertainment"
    
    def test_brand_emojis(self):
        """Test brand emojis."""
        assert Brand.FASHION.emoji == "👗"
        assert Brand.TECH.emoji == "📱"
        assert Brand.GAMING.emoji == "🎮"
        assert Brand.BEAUTY.emoji == "💄"
        assert Brand.HOME.emoji == "🏠"


class TestProduct:
    """Test Product model."""
    
    def test_product_creation(self, sample_product):
        """Test creating a product."""
        from shared.types import Product
        
        product = Product(
            id=sample_product["id"],
            name=sample_product["name"],
            brand=Brand(sample_product["brand"]),
            description=sample_product["description"],
            price=Decimal(str(sample_product["price"])),
            category=sample_product["category"],
            tags=sample_product["tags"],
            stock=sample_product["stock"],
        )
        
        assert product.id == "test-001"
        assert product.name == "Test Product"
        assert product.brand == Brand.FASHION
        assert product.price == Decimal("99.90")
    
    def test_product_format_price_brl(self, sample_product):
        """Test price formatting for BRL."""
        from shared.types import Product
        
        product = Product(
            id=sample_product["id"],
            name=sample_product["name"],
            brand=Brand.FASHION,
            description="",
            price=Decimal("199.90"),
            currency="BRL",
        )
        
        assert product.format_price() == "R$ 199.90"
    
    def test_product_to_dict(self, sample_product):
        """Test product serialization."""
        from shared.types import Product
        
        product = Product(
            id=sample_product["id"],
            name=sample_product["name"],
            brand=Brand.FASHION,
            description="Test",
            price=Decimal("99.90"),
        )
        
        data = product.to_dict()
        assert data["id"] == "test-001"
        assert data["brand"] == "fashion"
        assert data["price"] == 99.90
