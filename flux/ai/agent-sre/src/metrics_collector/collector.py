"""
Metrics Collector - Collects metrics from Prometheus record rules.
"""
from typing import Dict, Any, Optional
import httpx
import structlog

logger = structlog.get_logger()


class MetricsCollector:
    """Collects metrics from Prometheus for SRE reports."""
    
    def __init__(self, prometheus_url: str, timeout: int = 30):
        self.prometheus_url = prometheus_url.rstrip("/")
        self.timeout = timeout
        self.client = httpx.AsyncClient(timeout=timeout)
    
    async def query(self, query: str) -> Dict[str, Any]:
        """Query Prometheus for metrics."""
        url = f"{self.prometheus_url}/api/v1/query"
        params = {"query": query}
        
        try:
            response = await self.client.get(url, params=params)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            logger.error("Prometheus query failed", query=query, error=str(e))
            raise
    
    async def collect_loki_metrics(self) -> Dict[str, Any]:
        """Collect Loki health metrics from record rules."""
        metrics = {}
        
        # Query record rules
        record_rules = [
            "loki:health:availability:ratio",
            "loki:health:score",
            "loki:health:error_rate:ingestion",
            "loki:health:error_rate:query",
            "loki:health:query_latency:p95",
            "loki:health:dropped_entries:rate",
            "loki:health:dropped_lines:rate",
            "loki:health:ingestion_rate:bytes_per_sec",
            "loki:health:memory_streams:total",
        ]
        
        for rule in record_rules:
            try:
                result = await self.query(rule)
                if result.get("status") == "success":
                    data = result.get("data", {})
                    metrics[rule] = self._extract_value(data)
            except Exception as e:
                logger.warning("Failed to query metric", metric=rule, error=str(e))
                metrics[rule] = None
        
        return metrics
    
    async def collect_prometheus_metrics(self) -> Dict[str, Any]:
        """Collect Prometheus health metrics."""
        metrics = {}
        
        record_rules = [
            "prometheus:health:availability:ratio",
            "prometheus:health:score",
            "prometheus:health:scrape_failures:rate",
            "prometheus:health:query_latency:p95",
            "prometheus:health:storage:series_total",
        ]
        
        for rule in record_rules:
            try:
                result = await self.query(rule)
                if result.get("status") == "success":
                    data = result.get("data", {})
                    metrics[rule] = self._extract_value(data)
            except Exception as e:
                logger.warning("Failed to query metric", metric=rule, error=str(e))
                metrics[rule] = None
        
        return metrics
    
    async def collect_infrastructure_metrics(self) -> Dict[str, Any]:
        """Collect infrastructure health metrics."""
        metrics = {}
        
        record_rules = [
            "k8s:health:node_ready:ratio",
            "k8s:health:pod_ready:ratio",
            "k8s:health:cpu_utilization:ratio",
            "k8s:health:memory_utilization:ratio",
            "infrastructure:health:score",
        ]
        
        for rule in record_rules:
            try:
                result = await self.query(rule)
                if result.get("status") == "success":
                    data = result.get("data", {})
                    metrics[rule] = self._extract_value(data)
            except Exception as e:
                logger.warning("Failed to query metric", metric=rule, error=str(e))
                metrics[rule] = None
        
        return metrics
    
    async def collect_observability_metrics(self) -> Dict[str, Any]:
        """Collect overall observability health metrics."""
        metrics = {}
        
        record_rules = [
            "observability:health:score",
            "observability:health:components:ratio",
        ]
        
        for rule in record_rules:
            try:
                result = await self.query(rule)
                if result.get("status") == "success":
                    data = result.get("data", {})
                    metrics[rule] = self._extract_value(data)
            except Exception as e:
                logger.warning("Failed to query metric", metric=rule, error=str(e))
                metrics[rule] = None
        
        return metrics
    
    def _extract_value(self, data: Dict[str, Any]) -> Optional[float]:
        """Extract value from Prometheus query result."""
        result = data.get("result", [])
        if result and len(result) > 0:
            value = result[0].get("value", [None, None])[1]
            try:
                return float(value)
            except (ValueError, TypeError):
                return None
        return None
    
    async def close(self):
        """Close HTTP client."""
        await self.client.aclose()

