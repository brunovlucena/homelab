"""
Prometheus Metrics Exporter for Image Scanning

Exposes metrics about LambdaFunction image versions and outdated status.
"""

import logging
import os
from typing import Any, Optional
from prometheus_client import Gauge, Counter, Info, start_http_server, REGISTRY

logger = logging.getLogger(__name__)

# =============================================================================
# BUILD INFO (for Agent Versions Dashboard)
# =============================================================================

BUILD_INFO = Info(
    "agent_devsecops_build",
    "Build information"
)


def init_build_info(version: str, commit: str = "unknown"):
    """Initialize build info metric."""
    BUILD_INFO.info({
        "version": version,
        "commit": commit,
    })


class MetricsExporter:
    """Exports Prometheus metrics for LambdaFunction image scanning."""
    
    def __init__(self, port: int = 9090):
        self.port = int(os.getenv("METRICS_PORT", port))
        
        # LambdaFunction image version info metric
        self.lambda_image_info = Gauge(
            "devsecops_lambdafunction_image_info",
            "LambdaFunction image version information",
            ["function", "namespace", "image_uri", "tag", "version", "registry"]
        )
        
        # Outdated image counter
        self.outdated_images = Gauge(
            "devsecops_lambdafunction_outdated_total",
            "Number of LambdaFunctions with outdated images",
            ["namespace", "registry"]
        )
        
        # Total LambdaFunctions scanned
        self.total_scanned = Counter(
            "devsecops_lambdafunction_scans_total",
            "Total number of LambdaFunction scans performed",
            ["namespace", "status"]
        )
        
        # Scan errors
        self.scan_errors = Counter(
            "devsecops_lambdafunction_scan_errors_total",
            "Total number of scan errors",
            ["namespace", "error_type"]
        )
        
        logger.info(f"Metrics exporter initialized on port {self.port}")
    
    def start(self):
        """Start the Prometheus metrics HTTP server."""
        try:
            start_http_server(self.port)
            logger.info(f"Prometheus metrics server started on port {self.port}")
        except Exception as e:
            logger.error(f"Failed to start metrics server: {e}")
    
    def update_lambda_image_info(self, scan_result: dict[str, Any]):
        """Update metrics for a LambdaFunction image scan result."""
        name = scan_result.get("name", "unknown")
        namespace = scan_result.get("namespace", "default")
        image_uri = scan_result.get("image_uri", "")
        tag = scan_result.get("tag", "unknown")
        version = scan_result.get("version", "unknown")
        is_outdated = scan_result.get("is_outdated", False)
        
        # Extract registry from image URI
        registry = "unknown"
        if image_uri:
            parts = image_uri.split("/")
            if len(parts) > 0:
                registry = parts[0].split(":")[0]  # Remove port if present
        
        # Set metric (1 if image exists, 0 otherwise)
        value = 1 if image_uri else 0
        self.lambda_image_info.labels(
            function=name,
            namespace=namespace,
            image_uri=image_uri or "none",
            tag=tag or "none",
            version=version or "none",
            registry=registry
        ).set(value)
        
        # Note: outdated_images is a Gauge that will be recalculated on each scan
        # We don't increment here, but rather set it based on the current scan results
    
    def record_scan(self, namespace: str, status: str = "success"):
        """Record a scan operation."""
        self.total_scanned.labels(namespace=namespace, status=status).inc()
    
    def record_error(self, namespace: str, error_type: str):
        """Record a scan error."""
        self.scan_errors.labels(namespace=namespace, error_type=error_type).inc()
    
    def reset_outdated_counters(self):
        """Reset outdated image counters (call before a new scan)."""
        # Note: Gauges don't need reset, but we can set them to 0
        # In practice, we'll recalculate on each scan
        pass


# Global metrics exporter instance
_metrics_exporter: Optional[MetricsExporter] = None


def get_metrics_exporter(port: int = 9090) -> MetricsExporter:
    """Get or create the global metrics exporter instance."""
    global _metrics_exporter
    if _metrics_exporter is None:
        _metrics_exporter = MetricsExporter(port)
    return _metrics_exporter
