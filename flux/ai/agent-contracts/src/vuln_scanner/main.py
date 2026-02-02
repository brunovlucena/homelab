"""
Vulnerability Scanner - FastAPI entry point.

Handles CloudEvents for contract vulnerability scanning.
"""
import os
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request, Response, HTTPException
from fastapi.responses import JSONResponse
from cloudevents.http import from_http, to_structured, CloudEvent
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
import structlog
from opentelemetry import trace

from .handler import VulnerabilityScanner, create_vuln_found_event, Severity
# Import metrics to ensure they're registered with Prometheus
from shared import metrics

logger = structlog.get_logger()

# Global scanner
scanner: VulnerabilityScanner = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Initialize and cleanup resources."""
    global scanner
    
    scanner = VulnerabilityScanner()
    logger.info("vulnerability_scanner_initialized")
    
    yield
    
    logger.info("vulnerability_scanner_shutdown")


app = FastAPI(
    title="Vulnerability Scanner",
    description="Static and LLM-based smart contract vulnerability analysis",
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
    if scanner is None:
        raise HTTPException(status_code=503, detail="Scanner not initialized")
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
    - io.homelab.contract.created: New contract to scan
    
    Uses the same shared business logic as API endpoint for consistency and tracing.
    """
    import time
    start_time = time.monotonic()
    
    tracer = trace.get_tracer(__name__)
    
    # Parse CloudEvent
    try:
        event = from_http(dict(request.headers), await request.body())
    except Exception as e:
        logger.error("cloudevent_parse_failed", error=str(e))
        raise HTTPException(status_code=400, detail=f"Invalid CloudEvent: {e}")
    
    event_type = event["type"]
    event_id = event.get("id", "unknown")
    event_source = event.get("source", "unknown")
    
    # Create tracing span for CloudEvent processing
    with tracer.start_as_current_span(
        f"cloudevent.{event_type.replace('.', '_').replace(':', '_')}",
        attributes={
            "cloudevent.type": event_type,
            "cloudevent.source": event_source,
            "cloudevent.id": event_id,
        }
    ) as span:
        log = logger.bind(
            event_type=event_type,
            event_id=event_id,
        )
        log.info("event_received", source_type="cloudevent")
        
        if event_type != "io.homelab.contract.created":
            log.warning("unknown_event_type")
            span.set_attribute("cloudevent.error", "unknown_event_type")
            span.set_attribute("cloudevent.duration_ms", (time.monotonic() - start_time) * 1000)
            return JSONResponse(
                status_code=200,
                content={"status": "ignored", "reason": "Unknown event type"}
            )
        
        # Use shared business logic (same as API endpoint)
        result = await handle_contract_scan(event, span)
        
        span.set_attribute("cloudevent.duration_ms", (time.monotonic() - start_time) * 1000)
        return result


async def handle_contract_scan(event: CloudEvent, span=None) -> Response:
    """
    Handle contract scanning request.
    
    Shared business logic used by both API endpoint and CloudEvent handler.
    """
    data = event.data
    chain = data.get("chain", "ethereum")
    address = data.get("address")
    source_code = data.get("source_code")
    contract_name = data.get("contract_name", "Contract")
    
    if span:
        span.set_attribute("contract.chain", chain)
        span.set_attribute("contract.address", address or "unknown")
        span.set_attribute("contract.name", contract_name)
    
    if not address:
        if span:
            span.set_attribute("cloudevent.error", "missing_address")
        raise HTTPException(status_code=400, detail="Missing 'address' in event data")
    
    if not source_code:
        logger.warning("no_source_code", address=address)
        if span:
            span.set_attribute("scan.skipped", True)
            span.set_attribute("scan.reason", "no_source_code")
        return JSONResponse(
            status_code=200,
            content={"status": "skipped", "reason": "No source code available"}
        )
    
    log = logger.bind(chain=chain, address=address)
    log.info("scanning_contract")
    
    try:
        # Use shared business logic (same as API endpoint)
        result = await scanner.scan(chain, address, source_code, contract_name)
        
        if span:
            span.set_attribute("scan.vulnerabilities_found", len(result.vulnerabilities))
            if result.max_severity:
                span.set_attribute("scan.max_severity", result.max_severity.value)
            span.set_attribute("scan.duration_seconds", result.scan_duration_seconds)
            span.set_attribute("scan.analyzers", ",".join(result.analyzers_used))
        
        # Create output CloudEvent
        output_event = create_vuln_found_event(result)
        headers, body = to_structured(output_event)
        
        # Only send to sink if vulnerabilities found
        if result.vulnerabilities:
            k_sink = os.getenv("K_SINK")
            if k_sink:
                import httpx
                async with httpx.AsyncClient() as client:
                    await client.post(k_sink, headers=headers, content=body)
                    log.info(
                        "event_sent_to_sink",
                        sink=k_sink,
                        num_vulns=len(result.vulnerabilities)
                    )
                    if span:
                        span.set_attribute("cloudevent.sent_to_sink", True)
                        span.set_attribute("cloudevent.sink", k_sink)
        
        return Response(
            content=body,
            media_type="application/cloudevents+json",
            headers=dict(headers)
        )
        
    except Exception as e:
        log.error("scan_failed", error=str(e))
        if span:
            span.set_attribute("cloudevent.error", str(e))
        raise HTTPException(status_code=500, detail=str(e))


# Direct API endpoint for testing
@app.post("/scan")
async def scan_contract(
    chain: str = "ethereum",
    address: str = None,
    source_code: str = None,
):
    """
    Direct API to scan a contract (for testing).
    
    Uses the same shared business logic as CloudEvent handler.
    """
    if not address or not source_code:
        raise HTTPException(status_code=400, detail="Missing address or source_code")
    
    logger.info(
        "scan_requested",
        chain=chain,
        address=address,
        source="api",
    )
    
    # Use shared business logic (same as CloudEvent handler)
    result = await scanner.scan(chain, address, source_code)
    
    return {
        "chain": result.chain,
        "address": result.address,
        "vulnerabilities": [v.to_dict() for v in result.vulnerabilities],
        "max_severity": result.max_severity.value if result.max_severity else None,
        "scan_duration_seconds": result.scan_duration_seconds,
        "analyzers_used": result.analyzers_used,
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)

