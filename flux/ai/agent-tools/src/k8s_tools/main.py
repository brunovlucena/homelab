"""
Main entrypoint for Agent Tools - K8s API operations.
"""

import json
import logging
import os
import sys

from flask import Flask, request, jsonify
from cloudevents.http import from_http, to_structured, CloudEvent
from prometheus_client import make_wsgi_app, Counter, Histogram, Info
from werkzeug.middleware.dispatcher import DispatcherMiddleware
from opentelemetry import trace

from handler import handle
from metrics import init_build_info, BUILD_INFO

# Logging
logging.basicConfig(
    level=os.getenv("LOG_LEVEL", "INFO").upper(),
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

# Metrics
EVENTS_TOTAL = Counter(
    "agent_tools_events_total",
    "Total CloudEvents processed",
    ["event_type", "status"]
)
EVENT_DURATION = Histogram(
    "agent_tools_event_duration_seconds",
    "Event processing duration",
    ["event_type"]
)

# Flask app
app = Flask(__name__)
app.wsgi_app = DispatcherMiddleware(app.wsgi_app, {"/metrics": make_wsgi_app()})


@app.route("/", methods=["POST"])
def handle_event():
    """
    Handle incoming CloudEvent.
    
    Uses the same shared business logic as API endpoint for consistency and tracing.
    """
    import time
    start_time = time.time()
    
    tracer = trace.get_tracer(__name__)
    
    try:
        event = from_http(request.headers, request.get_data())
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
            logger.info(
                f"Received event: {event_type}",
                extra={
                    "event_type": event_type,
                    "event_id": event_id,
                    "source": event_source,
                    "source_type": "cloudevent",
                }
            )
            
            # Use shared business logic (same as API endpoint)
            with EVENT_DURATION.labels(event_type=event_type).time():
                result = handle({
                    "type": event_type,
                    "source": event_source,
                    "id": event_id,
                    "data": event.data or {},
                })
            
            status = "success" if result.get("success") else "error"
            
            if span:
                span.set_attribute("operation.status", status)
                span.set_attribute("operation.name", result.get("operation", "unknown"))
                span.set_attribute("operation.resource", result.get("resource", "unknown"))
                if result.get("durationMs"):
                    span.set_attribute("operation.duration_ms", result.get("durationMs"))
                span.set_attribute("cloudevent.duration_ms", (time.time() - start_time) * 1000)
            
            EVENTS_TOTAL.labels(event_type=event_type, status=status).inc()
            
            # Create response CloudEvent
            response_type = f"{event_type}.{status}"
            response_event = CloudEvent(
                attributes={
                    "type": response_type,
                    "source": "/agent-tools/k8s",
                },
                data=result,
            )
            
            headers, body = to_structured(response_event)
            return body, 200, headers
    
    except Exception as e:
        logger.exception(f"Error: {e}")
        EVENTS_TOTAL.labels(event_type="unknown", status="error").inc()
        return jsonify({"error": str(e)}), 500


@app.route("/health", methods=["GET"])
def health():
    return {"status": "healthy"}, 200


@app.route("/ready", methods=["GET"])
def ready():
    return {"status": "ready"}, 200


if __name__ == "__main__":
    # Initialize build info for Agent Versions dashboard
    version = os.getenv("VERSION", "0.1.0")
    commit = os.getenv("GIT_COMMIT", "unknown")
    init_build_info(version, commit)
    logger.info(f"Starting agent-tools v{version}")
    
    port = int(os.getenv("PORT", "8080"))
    app.run(host="0.0.0.0", port=port)
