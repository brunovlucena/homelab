"""
Main entry point for agent-blueteam defense runner.

🛡️ Blue Team - Protecting the realm from evil exploits and MAG7!
"""
import os
import json
import asyncio
from contextlib import asynccontextmanager

import structlog
import uvicorn
from fastapi import FastAPI, Request, Response, HTTPException
from fastapi.responses import JSONResponse
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST, REGISTRY
from opentelemetry import trace

from handler import DefenseRunner
from shared.types import ThreatLevel
from shared.metrics import (
    REQUEST_COUNT,
    REQUEST_LATENCY,
    CLOUDEVENTS_RECEIVED,
    CLOUDEVENTS_PROCESSED,
    init_build_info,
)

# Configure structured logging
structlog.configure(
    processors=[
        structlog.stdlib.add_log_level,
        structlog.stdlib.add_logger_name,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.JSONRenderer(),
    ],
    wrapper_class=structlog.stdlib.BoundLogger,
    context_class=dict,
    logger_factory=structlog.stdlib.LoggerFactory(),
)

logger = structlog.get_logger()

# Global defense runner instance
defense_runner: DefenseRunner = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan handler."""
    global defense_runner
    
    logger.info("agent_blueteam_starting")
    
    # Initialize build info for Agent Versions dashboard
    version = os.getenv("VERSION", "1.0.0")
    commit = os.getenv("GIT_COMMIT", "unknown")
    init_build_info(version, commit)
    
    # Initialize defense runner
    defense_runner = DefenseRunner()
    
    logger.info(
        "agent_blueteam_ready",
        version=version,
        defense_mode=defense_runner.defense_mode,
        mag7_health=defense_runner.mag7.health,
    )
    
    yield
    
    logger.info("agent_blueteam_shutting_down")


app = FastAPI(
    title="Agent Blueteam - Defense Runner",
    description="🛡️ Defensive security agent for threat mitigation and MAG7 battle",
    version="1.0.0",
    lifespan=lifespan,
)


@app.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "healthy", "agent": "blueteam"}


@app.get("/ready")
async def ready():
    """Readiness check endpoint."""
    if defense_runner is None:
        raise HTTPException(status_code=503, detail="Defense runner not initialized")
    return {"status": "ready", "defense_mode": defense_runner.defense_mode}


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(REGISTRY),
        media_type=CONTENT_TYPE_LATEST,
    )


@app.get("/mag7/status")
async def get_mag7_status():
    """Get current MAG7 boss status."""
    if defense_runner is None:
        raise HTTPException(status_code=503, detail="Defense runner not initialized")
    return defense_runner.get_mag7_status()


@app.post("/mag7/attack")
async def attack_mag7(request: Request):
    """
    Deal damage to MAG7 boss.
    
    Body: {"damage": int, "attack_type": str}
    """
    if defense_runner is None:
        raise HTTPException(status_code=503, detail="Defense runner not initialized")
    
    data = await request.json()
    damage = data.get("damage", 10)
    attack_type = data.get("attack_type", "exploit_blocked")
    
    logger.info(
        "mag7_attack_requested",
        damage=damage,
        attack_type=attack_type,
        source="api",
    )
    
    # Use shared business logic
    result = await defense_runner.attack_mag7(damage, attack_type)
    return result


@app.post("/mag7/reset")
async def reset_mag7():
    """Reset MAG7 boss for a new game."""
    if defense_runner is None:
        raise HTTPException(status_code=503, detail="Defense runner not initialized")
    
    from shared.types import MAG7Boss
    defense_runner.mag7 = MAG7Boss()
    
    from shared.metrics import MAG7_HEALTH
    MAG7_HEALTH.set(defense_runner.mag7.health)
    
    return {"status": "reset", "mag7_status": defense_runner.get_mag7_status()}


@app.post("/")
async def handle_cloudevent(request: Request):
    """
    Handle incoming CloudEvents.
    
    This is the main entry point for events from:
    - agent-redteam (exploit events)
    - knative-lambda-operator (lifecycle events)
    - MAG7 battle demo (game events)
    
    Uses the same shared business logic as API endpoints for consistency and tracing.
    """
    import time
    start_time = time.monotonic()
    
    tracer = trace.get_tracer(__name__)
    
    if defense_runner is None:
        raise HTTPException(status_code=503, detail="Defense runner not initialized")
    
    # Parse CloudEvent headers
    ce_type = request.headers.get("ce-type", "unknown")
    ce_source = request.headers.get("ce-source", "unknown")
    ce_id = request.headers.get("ce-id", "unknown")
    
    # Create tracing span for CloudEvent processing
    with tracer.start_as_current_span(
        f"cloudevent.{ce_type.replace('.', '_').replace(':', '_')}",
        attributes={
            "cloudevent.type": ce_type,
            "cloudevent.source": ce_source,
            "cloudevent.id": ce_id,
        }
    ) as span:
        log = logger.bind(
            ce_type=ce_type,
            ce_source=ce_source,
            ce_id=ce_id,
        )
        
        CLOUDEVENTS_RECEIVED.labels(event_type=ce_type, source=ce_source).inc()
        
        try:
            # Parse event body
            body = await request.json()
            
            log.info("cloudevent_received", body_keys=list(body.keys()), source_type="cloudevent")
            
            # Handle game events - use shared business logic
            if ce_type.startswith("io.homelab.demo.game"):
                span.set_attribute("event.category", "game")
                result = await defense_runner.handle_game_event({
                    "type": ce_type,
                    "data": body,
                })
                span.set_attribute("cloudevent.duration_ms", (time.monotonic() - start_time) * 1000)
                CLOUDEVENTS_PROCESSED.labels(event_type=ce_type, status="success").inc()
                return JSONResponse(content=result)
            
            # Handle MAG7 events - use shared business logic
            if ce_type.startswith("io.homelab.mag7"):
                span.set_attribute("event.category", "mag7")
                result = await defense_runner.handle_game_event({
                    "type": ce_type,
                    "data": body,
                })
                span.set_attribute("cloudevent.duration_ms", (time.monotonic() - start_time) * 1000)
                CLOUDEVENTS_PROCESSED.labels(event_type=ce_type, status="success").inc()
                return JSONResponse(content=result)
            
            # Handle exploit events - analyze threat using shared business logic
            if ce_type.startswith("io.homelab.exploit"):
                span.set_attribute("event.category", "exploit")
                event_data = {
                    "type": ce_type,
                    "exploit_id": body.get("exploit_id"),
                    "payload": body,
                    "namespace": body.get("namespace", "unknown"),
                }
                
                # Use shared business logic (same as would be used by API endpoint)
                threat_report = await defense_runner.analyze_threat(event_data)
                
                span.set_attribute("threat.level", threat_report.threat_level.value)
                span.set_attribute("threat.confidence", threat_report.confidence)
                if threat_report.matched_signature:
                    span.set_attribute("threat.signature", threat_report.matched_signature)
                
                log.info(
                    "threat_analyzed",
                    threat_level=threat_report.threat_level.value,
                    confidence=threat_report.confidence,
                    matched_signature=threat_report.matched_signature,
                )
                
                # Execute defense if needed - use shared business logic
                if threat_report.threat_level in [ThreatLevel.HIGH, ThreatLevel.CRITICAL]:
                    defense_result = await defense_runner.execute_defense(threat_report)
                    
                    span.set_attribute("defense.action", defense_result.action_taken.value)
                    span.set_attribute("defense.success", defense_result.success)
                    
                    # If defense was successful, deal damage to MAG7! - use shared business logic
                    if defense_result.success and threat_report.threat_level == ThreatLevel.CRITICAL:
                        damage = 100  # Critical threats deal more damage
                        await defense_runner.attack_mag7(damage, "critical_blocked")
                        span.set_attribute("mag7.damage", damage)
                    elif defense_result.success:
                        damage = 50
                        await defense_runner.attack_mag7(damage, "threat_blocked")
                        span.set_attribute("mag7.damage", damage)
                    
                    span.set_attribute("cloudevent.duration_ms", (time.monotonic() - start_time) * 1000)
                    CLOUDEVENTS_PROCESSED.labels(event_type=ce_type, status="defended").inc()
                    
                    return JSONResponse(content={
                        "status": "defended",
                        "threat_report": {
                            "id": threat_report.id,
                            "threat_level": threat_report.threat_level.value,
                            "confidence": threat_report.confidence,
                        },
                        "defense_result": {
                            "action": defense_result.action_taken.value,
                            "success": defense_result.success,
                            "message": defense_result.message,
                        },
                        "mag7_status": defense_runner.get_mag7_status(),
                        "event_id": ce_id,
                        "duration_ms": (time.monotonic() - start_time) * 1000,
                    })
                
                span.set_attribute("cloudevent.duration_ms", (time.monotonic() - start_time) * 1000)
                CLOUDEVENTS_PROCESSED.labels(event_type=ce_type, status="monitored").inc()
                
                return JSONResponse(content={
                    "status": "monitored",
                    "threat_report": {
                        "id": threat_report.id,
                        "threat_level": threat_report.threat_level.value,
                        "confidence": threat_report.confidence,
                    },
                    "event_id": ce_id,
                    "duration_ms": (time.monotonic() - start_time) * 1000,
                })
            
            # Default: acknowledge event
            span.set_attribute("cloudevent.duration_ms", (time.monotonic() - start_time) * 1000)
            CLOUDEVENTS_PROCESSED.labels(event_type=ce_type, status="acknowledged").inc()
            return JSONResponse(content={
                "status": "acknowledged",
                "event_type": ce_type,
                "event_id": ce_id,
                "duration_ms": (time.monotonic() - start_time) * 1000,
            })
            
        except Exception as e:
            log.error("cloudevent_processing_failed", error=str(e))
            span.set_attribute("cloudevent.error", str(e))
            span.set_attribute("cloudevent.duration_ms", (time.monotonic() - start_time) * 1000)
            CLOUDEVENTS_PROCESSED.labels(event_type=ce_type, status="error").inc()
            raise HTTPException(status_code=500, detail=str(e))
        finally:
            duration = time.monotonic() - start_time
            REQUEST_LATENCY.labels(endpoint="/", method="POST").observe(duration)
            REQUEST_COUNT.labels(endpoint="/", method="POST", status="200").inc()


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    uvicorn.run(app, host="0.0.0.0", port=port)
