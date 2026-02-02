"""
Product Catalog handler.

Manages product inventory, search, and AI-powered recommendations.
"""
import os
from typing import Optional, List
from decimal import Decimal
from datetime import datetime, timezone

import structlog
from cloudevents.http import CloudEvent

from shared.types import Brand, Product, SalesRecommendation
from shared.events import (
    EventType,
    EventPublisher,
    EventSubscriber,
    ProductQueryEvent,
)
from shared.metrics import (
    PRODUCT_QUERIES,
    PRODUCT_VIEWS,
    PRODUCT_STOCK_ALERTS,
)

logger = structlog.get_logger()


# Sample product catalog (in production, use database)
SAMPLE_PRODUCTS = {
    # Fashion
    "fashion-001": Product(
        id="fashion-001",
        name="Vestido Floral Verão",
        brand=Brand.FASHION,
        description="Vestido leve com estampa floral, perfeito para dias quentes",
        price=Decimal("199.90"),
        category="vestidos",
        subcategory="casual",
        tags=["verão", "floral", "leve"],
        stock=25,
        images=["https://store.example.com/images/vestido-floral.jpg"],
    ),
    "fashion-002": Product(
        id="fashion-002",
        name="Jaqueta Jeans Clássica",
        brand=Brand.FASHION,
        description="Jaqueta jeans atemporal que combina com tudo",
        price=Decimal("289.90"),
        category="jaquetas",
        subcategory="jeans",
        tags=["casual", "versátil", "clássico"],
        stock=15,
    ),
    "fashion-003": Product(
        id="fashion-003",
        name="Bolsa Tote Couro Sintético",
        brand=Brand.FASHION,
        description="Bolsa espaçosa em couro sintético de alta qualidade",
        price=Decimal("159.90"),
        category="bolsas",
        subcategory="tote",
        tags=["trabalho", "casual", "versátil"],
        stock=30,
    ),
    
    # Tech
    "tech-001": Product(
        id="tech-001",
        name="Fone Bluetooth Pro Max",
        brand=Brand.TECH,
        description="Fone com cancelamento de ruído ativo e 40h de bateria",
        price=Decimal("599.90"),
        category="audio",
        subcategory="fones",
        tags=["bluetooth", "cancelamento ruído", "premium"],
        stock=20,
    ),
    "tech-002": Product(
        id="tech-002",
        name="Carregador Sem Fio 15W",
        brand=Brand.TECH,
        description="Carregador wireless compatível com todos os smartphones modernos",
        price=Decimal("129.90"),
        category="acessórios",
        subcategory="carregadores",
        tags=["wireless", "rápido", "universal"],
        stock=50,
    ),
    "tech-003": Product(
        id="tech-003",
        name="Smartwatch Fitness Pro",
        brand=Brand.TECH,
        description="Relógio inteligente com GPS, monitor cardíaco e 100+ modos de exercício",
        price=Decimal("899.90"),
        category="wearables",
        subcategory="smartwatch",
        tags=["fitness", "saúde", "gps"],
        stock=12,
    ),
    
    # Home
    "home-001": Product(
        id="home-001",
        name="Luminária LED Mesa Articulada",
        brand=Brand.HOME,
        description="Luminária com 3 intensidades de luz e braço articulado",
        price=Decimal("179.90"),
        category="iluminação",
        subcategory="mesa",
        tags=["led", "home office", "ajustável"],
        stock=35,
    ),
    "home-002": Product(
        id="home-002",
        name="Organizador Multiuso Bambu",
        brand=Brand.HOME,
        description="Organizador sustentável em bambu para escritório ou banheiro",
        price=Decimal("89.90"),
        category="organização",
        subcategory="decorativo",
        tags=["sustentável", "bambu", "decoração"],
        stock=45,
    ),
    
    # Beauty
    "beauty-001": Product(
        id="beauty-001",
        name="Sérum Vitamina C 30ml",
        brand=Brand.BEAUTY,
        description="Sérum antioxidante com vitamina C pura para pele radiante",
        price=Decimal("149.90"),
        category="skincare",
        subcategory="sérum",
        tags=["vitamina c", "antioxidante", "clareador"],
        stock=40,
    ),
    "beauty-002": Product(
        id="beauty-002",
        name="Kit Maquiagem Básico",
        brand=Brand.BEAUTY,
        description="Kit completo com base, pó, blush e batom em tons neutros",
        price=Decimal("199.90"),
        category="maquiagem",
        subcategory="kit",
        tags=["básico", "neutro", "iniciante"],
        stock=25,
    ),
    
    # Gaming
    "gaming-001": Product(
        id="gaming-001",
        name="Mouse Gamer RGB 16000 DPI",
        brand=Brand.GAMING,
        description="Mouse gamer com sensor óptico de alta precisão e iluminação RGB",
        price=Decimal("249.90"),
        category="periféricos",
        subcategory="mouse",
        tags=["rgb", "precisão", "ergonômico"],
        stock=30,
    ),
    "gaming-002": Product(
        id="gaming-002",
        name="Teclado Mecânico Switch Blue",
        brand=Brand.GAMING,
        description="Teclado mecânico com switches blue clicky e retroiluminação",
        price=Decimal("399.90"),
        category="periféricos",
        subcategory="teclado",
        tags=["mecânico", "blue switch", "rgb"],
        stock=18,
    ),
    "gaming-003": Product(
        id="gaming-003",
        name="Headset 7.1 Surround",
        brand=Brand.GAMING,
        description="Headset gamer com som surround virtual 7.1 e microfone retrátil",
        price=Decimal("349.90"),
        category="audio",
        subcategory="headset",
        tags=["7.1", "surround", "microfone"],
        stock=22,
    ),
}


class ProductCatalog:
    """
    Product catalog manager.
    
    Handles:
    - Product search and filtering
    - Stock management
    - Product recommendations
    - Price calculations
    """
    
    def __init__(
        self,
        event_publisher: Optional[EventPublisher] = None,
        event_subscriber: Optional[EventSubscriber] = None,
    ):
        self.event_publisher = event_publisher
        self.event_subscriber = event_subscriber
        
        # In production, load from database
        self._products = SAMPLE_PRODUCTS.copy()
        
        logger.info(
            "product_catalog_initialized",
            product_count=len(self._products),
        )
    
    # =========================================================================
    # Search & Query
    # =========================================================================
    
    def search(
        self,
        query: Optional[str] = None,
        brand: Optional[Brand] = None,
        category: Optional[str] = None,
        price_min: Optional[Decimal] = None,
        price_max: Optional[Decimal] = None,
        tags: Optional[List[str]] = None,
        in_stock_only: bool = True,
        limit: int = 10,
    ) -> List[Product]:
        """
        Search products with filters.
        """
        results = []
        
        for product in self._products.values():
            # Filter by brand
            if brand and product.brand != brand:
                continue
            
            # Filter by category
            if category and product.category.lower() != category.lower():
                continue
            
            # Filter by price range
            if price_min and product.price < price_min:
                continue
            if price_max and product.price > price_max:
                continue
            
            # Filter by stock
            if in_stock_only and product.stock <= 0:
                continue
            
            # Filter by tags
            if tags:
                product_tags_lower = [t.lower() for t in product.tags]
                if not any(t.lower() in product_tags_lower for t in tags):
                    continue
            
            # Text search in name and description
            if query:
                query_lower = query.lower()
                if (query_lower not in product.name.lower() and 
                    query_lower not in product.description.lower() and
                    query_lower not in product.category.lower()):
                    continue
            
            results.append(product)
        
        # Track metric
        PRODUCT_QUERIES.labels(
            brand=brand.value if brand else "all",
            query_type="search"
        ).inc()
        
        return results[:limit]
    
    def get_by_id(self, product_id: str) -> Optional[Product]:
        """Get a single product by ID."""
        product = self._products.get(product_id)
        
        if product:
            PRODUCT_VIEWS.labels(
                brand=product.brand.value,
                category=product.category
            ).inc()
        
        return product
    
    def get_by_brand(self, brand: Brand, limit: int = 20) -> List[Product]:
        """Get all products for a brand."""
        return [
            p for p in self._products.values()
            if p.brand == brand and p.stock > 0
        ][:limit]
    
    # =========================================================================
    # Recommendations
    # =========================================================================
    
    def get_recommendations(
        self,
        brand: Brand,
        current_product_ids: Optional[List[str]] = None,
        customer_tags: Optional[List[str]] = None,
        recommendation_type: str = "general",
        limit: int = 3,
    ) -> List[SalesRecommendation]:
        """
        Get product recommendations.
        
        Types:
        - general: General suggestions based on popularity
        - cross_sell: Complementary products
        - upsell: Higher value alternatives
        - alternative: Similar products
        """
        recommendations = []
        current_ids = set(current_product_ids or [])
        
        # Get products in the same brand
        brand_products = [
            p for p in self._products.values()
            if p.brand == brand and p.id not in current_ids and p.stock > 0
        ]
        
        if recommendation_type == "cross_sell" and current_ids:
            # Find complementary products
            current_categories = {
                self._products[pid].category
                for pid in current_ids
                if pid in self._products
            }
            
            for product in brand_products:
                if product.category not in current_categories:
                    recommendations.append(SalesRecommendation(
                        product_id=product.id,
                        product_name=product.name,
                        reason="Complementa sua seleção",
                        confidence=0.8,
                        recommendation_type="cross_sell",
                        price_info=product.format_price(),
                    ))
                    
        elif recommendation_type == "upsell" and current_ids:
            # Find higher value alternatives
            current_prices = [
                self._products[pid].price
                for pid in current_ids
                if pid in self._products
            ]
            avg_price = sum(current_prices) / len(current_prices) if current_prices else Decimal("0")
            
            for product in brand_products:
                if product.price > avg_price * Decimal("1.2"):
                    recommendations.append(SalesRecommendation(
                        product_id=product.id,
                        product_name=product.name,
                        reason="Versão premium disponível",
                        confidence=0.7,
                        recommendation_type="upsell",
                        price_info=product.format_price(),
                    ))
                    
        else:
            # General recommendations (most popular / in stock)
            sorted_products = sorted(brand_products, key=lambda p: p.stock, reverse=True)
            
            for product in sorted_products:
                recommendations.append(SalesRecommendation(
                    product_id=product.id,
                    product_name=product.name,
                    reason="Produto popular",
                    confidence=0.6,
                    recommendation_type="general",
                    price_info=product.format_price(),
                ))
        
        PRODUCT_QUERIES.labels(
            brand=brand.value,
            query_type="recommend"
        ).inc()
        
        return recommendations[:limit]
    
    # =========================================================================
    # Stock Management
    # =========================================================================
    
    def update_stock(self, product_id: str, quantity_change: int) -> bool:
        """
        Update product stock.
        
        Positive quantity_change adds stock, negative removes.
        """
        if product_id not in self._products:
            return False
        
        product = self._products[product_id]
        new_stock = product.stock + quantity_change
        
        if new_stock < 0:
            return False
        
        product.stock = new_stock
        product.updated_at = datetime.now(timezone.utc).isoformat()
        
        # Check for low stock alert
        if new_stock <= 5 and new_stock > 0:
            PRODUCT_STOCK_ALERTS.labels(
                brand=product.brand.value,
                category=product.category
            ).inc()
            
            logger.warning(
                "low_stock_alert",
                product_id=product_id,
                stock=new_stock,
            )
        
        return True
    
    def reserve_stock(self, product_id: str, quantity: int) -> bool:
        """Reserve stock for an order (decrements stock)."""
        return self.update_stock(product_id, -quantity)
    
    def release_stock(self, product_id: str, quantity: int) -> bool:
        """Release reserved stock (increments stock)."""
        return self.update_stock(product_id, quantity)
    
    # =========================================================================
    # Event Handlers
    # =========================================================================
    
    async def handle_product_query(self, event: CloudEvent):
        """Handle product query event from AI sellers."""
        data = event.data or {}
        
        query = data.get("query", "")
        brand_str = data.get("brand")
        category = data.get("category")
        price_min = data.get("price_min")
        price_max = data.get("price_max")
        limit = data.get("limit", 5)
        
        brand = Brand(brand_str) if brand_str else None
        
        results = self.search(
            query=query,
            brand=brand,
            category=category,
            price_min=Decimal(str(price_min)) if price_min else None,
            price_max=Decimal(str(price_max)) if price_max else None,
            limit=limit,
        )
        
        # Emit result event
        if self.event_publisher:
            await self.event_publisher.publish(
                EventType.PRODUCT_QUERY_RESULT,
                {
                    "conversation_id": data.get("conversation_id"),
                    "products": [p.to_dict() for p in results],
                    "query": query,
                    "result_count": len(results),
                },
                subject=data.get("conversation_id"),
            )
        
        return results
    
    async def handle_product_recommend(self, event: CloudEvent):
        """Handle product recommendation request."""
        data = event.data or {}
        
        brand_str = data.get("brand", "fashion")
        current_cart = data.get("current_cart", [])
        rec_type = data.get("recommendation_type", "general")
        limit = data.get("limit", 3)
        
        brand = Brand(brand_str)
        
        recommendations = self.get_recommendations(
            brand=brand,
            current_product_ids=current_cart,
            recommendation_type=rec_type,
            limit=limit,
        )
        
        # Emit result event
        if self.event_publisher:
            await self.event_publisher.publish(
                EventType.PRODUCT_RECOMMEND_RESULT,
                {
                    "conversation_id": data.get("conversation_id"),
                    "recommendations": [r.to_dict() for r in recommendations],
                    "recommendation_type": rec_type,
                },
                subject=data.get("conversation_id"),
            )
        
        return recommendations
    
    def setup_event_handlers(self):
        """Register event handlers."""
        if self.event_subscriber:
            self.event_subscriber.register(
                EventType.PRODUCT_QUERY,
                self.handle_product_query
            )
            self.event_subscriber.register(
                EventType.PRODUCT_RECOMMEND,
                self.handle_product_recommend
            )
            logger.info("product_catalog_event_handlers_registered")
