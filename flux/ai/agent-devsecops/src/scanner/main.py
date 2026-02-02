"""
Agent DevSecOps - Main Entry Point

Starts the CloudEvent server for handling security events.

🔬 OBSERVABILITY: Uses structured logging for PROOF OF COMMUNICATION
"""

import logging
import os
import sys
import time
import threading
from datetime import datetime, timezone

from flask import Flask, request, jsonify
from cloudevents.http import from_http, CloudEvent
from handler import handle
from metrics_exporter import get_metrics_exporter, init_build_info

# Configure logging with JSON format for observability
logging.basicConfig(
    level=getattr(logging, os.getenv("LOG_LEVEL", "INFO").upper()),
    format='{"timestamp": "%(asctime)s", "level": "%(levelname)s", "logger": "%(name)s", "message": "%(message)s"}',
    stream=sys.stdout
)
logger = logging.getLogger(__name__)

# Agent identification for observability
AGENT_ID = "agent-devsecops"

app = Flask(__name__)


@app.route("/", methods=["POST"])
def receive_event():
    """
    Receive and process CloudEvents.
    
    🔬 PROOF OF COMMUNICATION: All events logged with source, type, and timing
    """
    start_time = time.perf_counter()
    
    try:
        event = from_http(request.headers, request.get_data())
        
        # 🔬 PROOF: Log event receipt with full context
        logger.info(f"Received event: {event['type']} from {event['source']}")
        logger.info(
            f'{{"event": "cloudevent_received", "agent_id": "{AGENT_ID}", '
            f'"event_type": "{event["type"]}", "event_id": "{event["id"]}", '
            f'"source": "{event["source"]}"}}'
        )
        
        result = handle(event)
        
        # 🔬 PROOF: Log successful processing
        duration_ms = (time.perf_counter() - start_time) * 1000
        logger.info(
            f'{{"event": "cloudevent_processed", "agent_id": "{AGENT_ID}", '
            f'"event_type": "{event["type"]}", "event_id": "{event["id"]}", '
            f'"success": true, "duration_ms": {duration_ms:.3f}}}'
        )
        
        return result, 200, {"Content-Type": "application/json"}
    except Exception as e:
        duration_ms = (time.perf_counter() - start_time) * 1000
        logger.exception(f"Error processing event: {e}")
        logger.error(
            f'{{"event": "cloudevent_processing_failed", "agent_id": "{AGENT_ID}", '
            f'"error": "{str(e)}", "duration_ms": {duration_ms:.3f}}}'
        )
        return jsonify({"error": str(e)}), 500


@app.route("/health", methods=["GET"])
def health():
    """Health check endpoint."""
    return jsonify({
        "status": "healthy",
        "agent": "devsecops",
        "version": os.getenv("VERSION", "0.1.0"),
        "rbac_level": os.getenv("RBAC_LEVEL", "readonly")
    })


@app.route("/ready", methods=["GET"])
def ready():
    """Readiness check endpoint."""
    return jsonify({"ready": True})


@app.route("/metrics", methods=["GET"])
def metrics():
    """Prometheus metrics endpoint - serves metrics from prometheus_client registry."""
    try:
        from prometheus_client import generate_latest, CONTENT_TYPE_LATEST, REGISTRY
        # Generate metrics from the global registry
        # This will include all metrics registered via prometheus_client
        metrics_output = generate_latest(REGISTRY)
        # Always return valid Prometheus format (even if empty)
        if not metrics_output or len(metrics_output.strip()) == 0:
            metrics_output = b"# No metrics available yet\n"
        return metrics_output, 200, {"Content-Type": CONTENT_TYPE_LATEST}
    except ImportError as e:
        logger.error(f"prometheus_client not available: {e}")
        return "# Metrics endpoint not available\n", 503, {"Content-Type": "text/plain"}
    except Exception as e:
        logger.exception(f"Error generating metrics: {e}")
        return f"# Error generating metrics: {str(e)}\n", 500, {"Content-Type": "text/plain"}


@app.route("/scan/lambdafunctions", methods=["POST"])
def scan_lambdafunctions():
    """Trigger a manual scan of LambdaFunctions for outdated images."""
    from handler import scanner
    from cloudevents.http import CloudEvent
    
    # Create a CloudEvent to trigger the scan
    event = CloudEvent({
        "type": "io.homelab.scan.lambdafunctions",
        "source": "agent-devsecops/api",
        "id": f"scan-{datetime.now(timezone.utc).isoformat()}",
    }, data=request.get_json() or {})
    
    try:
        result = scanner.handle_scan_lambdafunctions(event)
        return jsonify(result), 200
    except Exception as e:
        logger.exception(f"Error triggering scan: {e}")
        return jsonify({"error": str(e)}), 500


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    metrics_port = int(os.getenv("METRICS_PORT", "9090"))
    version = os.getenv("VERSION", "0.1.0")
    commit = os.getenv("GIT_COMMIT", "unknown")
    
    # Initialize build info for Agent Versions dashboard
    init_build_info(version, commit)
    
    logger.info(f"Starting Agent DevSecOps v{version} on port {port}")
    logger.info(f"RBAC Level: {os.getenv('RBAC_LEVEL', 'readonly')}")
    logger.info(f"Environment: {os.getenv('ENVIRONMENT', 'pro')}")
    
    # Start Prometheus metrics server in background thread
    metrics = get_metrics_exporter(metrics_port)
    metrics_thread = threading.Thread(target=metrics.start, daemon=True)
    metrics_thread.start()
    logger.info(f"Prometheus metrics server starting on port {metrics_port}")
    
    app.run(host="0.0.0.0", port=port)
