"""
Prometheus Events - CloudEvents converter for Alertmanager alerts.

Receives Alertmanager webhook payloads and converts them to CloudEvents format,
then publishes to Knative Eventing.
"""
import os
import json
import uuid
from datetime import datetime, timezone
from typing import Dict, Any, List
from fastapi import FastAPI, Request, HTTPException
from fastapi.responses import JSONResponse
import httpx
import structlog
from cloudevents.http import CloudEvent, to_structured, from_http

logger = structlog.get_logger()

app = FastAPI(title="Prometheus Events", version="0.1.0")

# Knative Eventing sink URL
SINK_URL = os.getenv("K_SINK", "http://default-broker.rabbitmq.svc.cluster.local")


def convert_alert_to_cloudevent(
    alert: Dict[str, Any],
    status: str,
    receiver: str,
    common_labels: Dict[str, Any],
    common_annotations: Dict[str, Any],
    version: str
) -> CloudEvent:
    """Convert Alertmanager alert to CloudEvent."""
    labels = alert.get("labels", {})
    annotations = alert.get("annotations", {})
    
    # Extract key information
    alertname = labels.get("alertname", "unknown")
    severity = labels.get("severity", "warning")
    
    # Determine event type based on status
    if status == "firing":
        event_type = "io.homelab.prometheus.alert.fired"
    elif status == "resolved":
        event_type = "io.homelab.prometheus.alert.resolved"
    else:
        event_type = f"io.homelab.prometheus.alert.{status}"
    
    # Create CloudEvent data
    event_data = {
        "alertname": alertname,
        "status": alert.get("status", status),
        "severity": severity,
        "labels": labels,
        "annotations": annotations,
        "commonLabels": common_labels,
        "commonAnnotations": common_annotations,
        "startsAt": alert.get("startsAt"),
        "endsAt": alert.get("endsAt"),
        "generatorURL": alert.get("generatorURL", ""),
        "receiver": receiver,
        "version": version
    }
    
    # Add PrometheusRule reference if available
    prometheus_rule = labels.get("prometheus_rule") or common_labels.get("prometheus_rule")
    if prometheus_rule:
        event_data["prometheusRule"] = prometheus_rule
    
    # Create CloudEvent
    event = CloudEvent(
        {
            "type": event_type,
            "source": f"prometheus/alertmanager/{receiver}",
            "id": str(uuid.uuid4()),
            "time": datetime.now(timezone.utc).isoformat(),
            "subject": alertname,
        },
        event_data
    )
    
    return event


@app.post("/webhook")
async def webhook(request: Request):
    """Receive Alertmanager webhook and convert to CloudEvents."""
    try:
        payload = await request.json()
        
        version = payload.get("version", "4")
        status = payload.get("status", "unknown")
        receiver = payload.get("receiver", "unknown")
        alerts = payload.get("alerts", [])
        common_labels = payload.get("commonLabels", {})
        common_annotations = payload.get("commonAnnotations", {})
        
        logger.info(
            "alertmanager_webhook_received",
            version=version,
            status=status,
            receiver=receiver,
            alert_count=len(alerts)
        )
        
        # Convert each alert to CloudEvent
        events = []
        for alert in alerts:
            event = convert_alert_to_cloudevent(
                alert=alert,
                status=status,
                receiver=receiver,
                common_labels=common_labels,
                common_annotations=common_annotations,
                version=version
            )
            events.append(event)
        
        # Publish events to Knative Eventing
        async with httpx.AsyncClient(timeout=30.0) as client:
            for event in events:
                headers, body = to_structured(event)
                
                try:
                    response = await client.post(
                        SINK_URL,
                        headers=dict(headers),
                        content=body
                    )
                    response.raise_for_status()
                    
                    logger.info(
                        "cloudevent_published",
                        event_type=event["type"],
                        event_id=event["id"],
                        status_code=response.status_code
                    )
                except Exception as e:
                    logger.error(
                        "cloudevent_publish_failed",
                        event_type=event["type"],
                        event_id=event["id"],
                        error=str(e)
                    )
                    # Continue with other events even if one fails
                    continue
        
        return JSONResponse(
            status_code=200,
            content={
                "status": "success",
                "events_published": len(events),
                "events": [
                    {
                        "type": event["type"],
                        "id": event["id"]
                    }
                    for event in events
                ]
            }
        )
        
    except json.JSONDecodeError as e:
        logger.error("invalid_json_payload", error=str(e))
        raise HTTPException(status_code=400, detail="Invalid JSON payload")
    except Exception as e:
        logger.error("webhook_processing_failed", error=str(e))
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "healthy", "sink": SINK_URL}


@app.get("/")
async def root():
    """Root endpoint."""
    return {
        "service": "prometheus-events",
        "version": "0.1.0",
        "endpoints": {
            "webhook": "/webhook",
            "health": "/health"
        }
    }

