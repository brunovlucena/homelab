"""
Contract Fetcher - FastAPI entry point.

Handles CloudEvents for contract fetching requests.

🔬 OBSERVABILITY: Uses AgentCommunicationLogger for PROOF OF COMMUNICATION
"""
import os
import time
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request, Response, HTTPException
from fastapi.responses import JSONResponse
from cloudevents.http import from_http, to_structured, CloudEvent
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
import structlog
import redis.asyncio as redis

from .handler import ContractFetcher, create_contract_created_event
# Import metrics to ensure they're registered with Prometheus
from shared import metrics

# 🔬 Enhanced observability for communication proof
try:
    from agent_communication import (
        AgentCommunicationLogger,
        record_event_received,
        record_event_sent,
        record_event_processed,
        init_agent_build_info,
    )
    OBSERVABILITY_AVAILABLE = True
except ImportError:
    OBSERVABILITY_AVAILABLE = False

logger = structlog.get_logger()

# Initialize communication logger for PROOF OF COMMUNICATION
AGENT_ID = "contract-fetcher"
comm_logger = AgentCommunicationLogger(AGENT_ID) if OBSERVABILITY_AVAILABLE else None

# Initialize build info for dashboards
if OBSERVABILITY_AVAILABLE:
    init_agent_build_info(
        agent_id=AGENT_ID,
        version=os.getenv("VERSION", "1.0.0"),
        commit=os.getenv("GIT_COMMIT", "unknown"),
    )

# Global clients
fetcher: ContractFetcher = None
redis_client: redis.Redis = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global fetcher, redis_client
    
    # Initialize Redis
    redis_url = os.getenv("REDIS_URL", "redis://localhost:6379")
    try:
        redis_client = redis.from_url(redis_url)
        await redis_client.ping()
        logger.info("redis_connected", url=redis_url)
    except Exception as e:
        logger.warning("redis_connection_failed", error=str(e))
        redis_client = None
    
    # Initialize fetcher
    fetcher = ContractFetcher(redis_client=redis_client)
    logger.info("contract_fetcher_initialized")
    
    yield
    
    # Cleanup
    if fetcher:
        await fetcher.close()
    if redis_client:
        await redis_client.close()


app = FastAPI(
    title="Contract Fetcher",
    description="Fetch smart contract source and metadata from block explorers",
    version="1.0.0",
    lifespan=lifespan,
)


@app.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "healthy"}


@app.get("/ready")
async def ready():
    """Readiness check endpoint."""
    if fetcher is None:
        raise HTTPException(status_code=503, detail="Fetcher not initialized")
    return {"status": "ready"}


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST
    )


@app.post("/")
async def handle_event(request: Request):
    """
    Handle incoming CloudEvents.
    
    Accepts:
    - io.homelab.scan.request: Manual scan request
    - io.homelab.block.new: New block notification (for monitoring)
    
    🔬 PROOF OF COMMUNICATION: All events are logged with trace context
    """
    start_time = time.perf_counter()
    
    # Parse CloudEvent
    try:
        event = from_http(dict(request.headers), await request.body())
    except Exception as e:
        logger.error("cloudevent_parse_failed", error=str(e))
        raise HTTPException(status_code=400, detail=f"Invalid CloudEvent: {e}")
    
    # 🔬 PROOF: Log event receipt with full context
    if comm_logger:
        comm_logger.event_received(event)
        record_event_received(AGENT_ID, event["type"], event.get("source", "unknown"))
    
    log = logger.bind(
        event_type=event["type"],
        event_id=event["id"],
        source=event.get("source", "unknown"),
    )
    log.info("cloudevent_received")
    
    # Handle based on event type
    try:
        if event["type"] == "io.homelab.scan.request":
            result = await handle_scan_request(event)
        elif event["type"] == "io.homelab.block.new":
            result = await handle_new_block(event)
        else:
            log.warning("unknown_event_type")
            result = JSONResponse(
                status_code=200,
                content={"status": "ignored", "reason": "Unknown event type"}
            )
        
        # 🔬 PROOF: Log successful processing
        duration_ms = (time.perf_counter() - start_time) * 1000
        if comm_logger:
            comm_logger.event_processed(event, success=True, duration_ms=duration_ms)
            record_event_processed(AGENT_ID, event["type"], duration_ms / 1000, success=True)
        
        return result
        
    except Exception as e:
        duration_ms = (time.perf_counter() - start_time) * 1000
        if comm_logger:
            comm_logger.event_error(event, str(e), duration_ms=duration_ms)
        raise


async def handle_scan_request(event: CloudEvent) -> Response:
    """Handle manual scan request."""
    data = event.data
    chain = data.get("chain", "ethereum")
    address = data.get("address")
    
    if not address:
        raise HTTPException(status_code=400, detail="Missing 'address' in event data")
    
    log = logger.bind(chain=chain, address=address)
    log.info("fetching_contract")
    
    try:
        contract = await fetcher.fetch(chain, address)
        
        # Create output CloudEvent
        output_event = create_contract_created_event(contract)
        headers, body = to_structured(output_event)
        
        # Send to K_SINK if configured
        k_sink = os.getenv("K_SINK")
        if k_sink:
            import httpx
            async with httpx.AsyncClient() as client:
                await client.post(k_sink, headers=headers, content=body)
                log.info("event_sent_to_sink", sink=k_sink)
        
        return Response(
            content=body,
            media_type="application/cloudevents+json",
            headers=dict(headers)
        )
        
    except Exception as e:
        log.error("fetch_failed", error=str(e))
        raise HTTPException(status_code=500, detail=str(e))


async def handle_new_block(event: CloudEvent) -> Response:
    """Handle new block notification - scan for new contracts."""
    data = event.data
    chain = data.get("chain", "ethereum")
    block_number = data.get("block_number")
    new_contracts = data.get("new_contracts", [])
    
    log = logger.bind(chain=chain, block=block_number, num_contracts=len(new_contracts))
    log.info("processing_new_block")
    
    # Fetch each new contract
    results = []
    for address in new_contracts[:10]:  # Limit to 10 per block
        try:
            contract = await fetcher.fetch(chain, address)
            results.append({
                "address": address,
                "status": "fetched",
                "verified": contract.is_verified,
            })
        except Exception as e:
            results.append({
                "address": address,
                "status": "failed",
                "error": str(e),
            })
    
    return JSONResponse(content={"processed": len(results), "results": results})


# Direct API endpoint for testing
@app.post("/fetch")
async def fetch_contract(chain: str = "ethereum", address: str = None):
    """Direct API to fetch a contract (for testing)."""
    if not address:
        raise HTTPException(status_code=400, detail="Missing address parameter")
    
    contract = await fetcher.fetch(chain, address)
    return contract.to_dict()


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)

