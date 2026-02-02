"""
Product Catalog Agent - FastAPI entry point.

Manages product inventory, search, and recommendations.
"""
import os
from contextlib import asynccontextmanager
from typing import Optional, List
from decimal import Decimal

from fastapi import FastAPI, HTTPException, Request, Query
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import Response
from pydantic import BaseModel, Field
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
from cloudevents.http import from_http
import structlog

from .handler import ProductCatalog
from shared.types import Brand
from shared.metrics import init_metrics
from shared.events import init_events, shutdown_events

logger = structlog.get_logger()

# Global catalog instance
catalog: ProductCatalog = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global catalog
    
    # Initialize CloudEvents
    pub, sub = init_events(source="/agent-store-multibrands/product-catalog")
    
    # Initialize catalog
    catalog = ProductCatalog(
        event_publisher=pub,
        event_subscriber=sub,
    )
    catalog.setup_event_handlers()
    
    # Initialize metrics
    version = os.getenv("VERSION", "0.1.0")
    commit = os.getenv("GIT_COMMIT", "unknown")
    init_metrics(version, commit, "product-catalog")
    
    logger.info("product_catalog_started", version=version)
    
    yield
    
    # Cleanup
    await shutdown_events()
    logger.info("product_catalog_shutdown")


app = FastAPI(
    title="Product Catalog",
    description="Product inventory and recommendation service",
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

class SearchRequest(BaseModel):
    """Product search request."""
    query: Optional[str] = None
    brand: Optional[str] = None
    category: Optional[str] = None
    price_min: Optional[float] = None
    price_max: Optional[float] = None
    tags: Optional[List[str]] = None
    in_stock_only: bool = True
    limit: int = Field(default=10, le=50)


class RecommendRequest(BaseModel):
    """Recommendation request."""
    brand: str
    current_product_ids: Optional[List[str]] = None
    recommendation_type: str = Field(default="general")  # general, cross_sell, upsell
    limit: int = Field(default=3, le=10)


class StockUpdateRequest(BaseModel):
    """Stock update request."""
    product_id: str
    quantity_change: int


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
    if catalog is None:
        raise HTTPException(status_code=503, detail="Catalog not initialized")
    return {"status": "ready", "product_count": len(catalog._products)}


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST
    )


# =============================================================================
# Product Endpoints
# =============================================================================

@app.post("/search")
async def search_products(request: SearchRequest):
    """
    Search products with filters.
    """
    if catalog is None:
        raise HTTPException(status_code=503, detail="Catalog not initialized")
    
    brand = Brand(request.brand) if request.brand else None
    
    results = catalog.search(
        query=request.query,
        brand=brand,
        category=request.category,
        price_min=Decimal(str(request.price_min)) if request.price_min else None,
        price_max=Decimal(str(request.price_max)) if request.price_max else None,
        tags=request.tags,
        in_stock_only=request.in_stock_only,
        limit=request.limit,
    )
    
    return {
        "count": len(results),
        "products": [p.to_dict() for p in results],
    }


@app.get("/products/{product_id}")
async def get_product(product_id: str):
    """Get a single product by ID."""
    if catalog is None:
        raise HTTPException(status_code=503, detail="Catalog not initialized")
    
    product = catalog.get_by_id(product_id)
    
    if product is None:
        raise HTTPException(status_code=404, detail="Product not found")
    
    return product.to_dict()


@app.get("/products/brand/{brand}")
async def get_products_by_brand(brand: str, limit: int = Query(default=20, le=100)):
    """Get all products for a brand."""
    if catalog is None:
        raise HTTPException(status_code=503, detail="Catalog not initialized")
    
    try:
        brand_enum = Brand(brand.lower())
    except ValueError:
        raise HTTPException(status_code=400, detail=f"Invalid brand: {brand}")
    
    results = catalog.get_by_brand(brand_enum, limit=limit)
    
    return {
        "brand": brand_enum.value,
        "brand_display": brand_enum.display_name,
        "count": len(results),
        "products": [p.to_dict() for p in results],
    }


@app.post("/recommendations")
async def get_recommendations(request: RecommendRequest):
    """
    Get product recommendations.
    """
    if catalog is None:
        raise HTTPException(status_code=503, detail="Catalog not initialized")
    
    try:
        brand = Brand(request.brand.lower())
    except ValueError:
        raise HTTPException(status_code=400, detail=f"Invalid brand: {request.brand}")
    
    recommendations = catalog.get_recommendations(
        brand=brand,
        current_product_ids=request.current_product_ids,
        recommendation_type=request.recommendation_type,
        limit=request.limit,
    )
    
    return {
        "count": len(recommendations),
        "recommendation_type": request.recommendation_type,
        "recommendations": [r.to_dict() for r in recommendations],
    }


# =============================================================================
# Stock Management Endpoints
# =============================================================================

@app.post("/stock/update")
async def update_stock(request: StockUpdateRequest):
    """
    Update product stock.
    
    Positive quantity_change adds stock, negative removes.
    """
    if catalog is None:
        raise HTTPException(status_code=503, detail="Catalog not initialized")
    
    success = catalog.update_stock(
        product_id=request.product_id,
        quantity_change=request.quantity_change,
    )
    
    if not success:
        raise HTTPException(
            status_code=400,
            detail="Failed to update stock (product not found or insufficient stock)"
        )
    
    product = catalog.get_by_id(request.product_id)
    
    return {
        "status": "updated",
        "product_id": request.product_id,
        "new_stock": product.stock if product else 0,
    }


# =============================================================================
# CloudEvents Endpoint
# =============================================================================

@app.post("/events")
async def receive_cloudevent(request: Request):
    """
    Receive CloudEvents from Knative triggers.
    
    Subscribed events:
    - store.product.query: Product search request
    - store.product.recommend: Recommendation request
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
        
        if catalog and catalog.event_subscriber:
            handled = await catalog.event_subscriber.handle(event)
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
        "service": "product-catalog",
        "description": "Product inventory and recommendation service",
        "version": os.getenv("VERSION", "0.1.0"),
        "endpoints": {
            "search": "POST /search",
            "get_product": "GET /products/{id}",
            "by_brand": "GET /products/brand/{brand}",
            "recommendations": "POST /recommendations",
            "stock_update": "POST /stock/update",
            "events": "POST /events (CloudEvents)",
            "health": "GET /health",
            "ready": "GET /ready",
            "metrics": "GET /metrics",
        },
        "brands": [b.value for b in Brand],
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
