"""
📝 Loki query client
Provides log querying capabilities for Grafana Loki
"""

import logging
from datetime import datetime, timedelta
from typing import Any, Dict, List, Optional
from urllib.parse import urlencode

import aiohttp

logger = logging.getLogger(__name__)


class LokiClient:
    """Client for querying Grafana Loki"""

    def __init__(self, url: str, timeout: int = 30):
        """Initialize Loki client"""
        self.url = url.rstrip("/")
        self.timeout = timeout
        logger.info(f"📝 Initialized Loki client: {self.url}")

    async def query_range(
        self,
        query: str,
        start: Optional[datetime] = None,
        end: Optional[datetime] = None,
        limit: int = 100,
        direction: str = "backward",
    ) -> Dict[str, Any]:
        """
        Query Loki logs over a time range

        Args:
            query: LogQL query string (e.g., '{namespace="default"}')
            start: Start time (defaults to 30 minutes ago)
            end: End time (defaults to now)
            limit: Maximum number of log lines to return
            direction: 'forward' or 'backward'

        Returns:
            Dictionary with query results
        """
        if not start:
            start = datetime.utcnow() - timedelta(minutes=30)
        if not end:
            end = datetime.utcnow()

        # Convert to nanosecond timestamps
        start_ns = int(start.timestamp() * 1e9)
        end_ns = int(end.timestamp() * 1e9)

        params = {
            "query": query,
            "start": start_ns,
            "end": end_ns,
            "limit": limit,
            "direction": direction,
        }

        logger.info(f"📝 Querying Loki: {query} from {start} to {end}")

        async with aiohttp.ClientSession() as session:
            try:
                async with session.get(
                    f"{self.url}/loki/api/v1/query_range",
                    params=params,
                    timeout=aiohttp.ClientTimeout(total=self.timeout),
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return {
                            "status": "success",
                            "query": query,
                            "start": start.isoformat(),
                            "end": end.isoformat(),
                            "result": data.get("data", {}),
                            "timestamp": datetime.utcnow().isoformat(),
                        }
                    else:
                        error_text = await response.text()
                        logger.error(f"❌ Loki error: {response.status} - {error_text}")
                        return {
                            "status": "error",
                            "query": query,
                            "error": f"HTTP {response.status}: {error_text}",
                            "timestamp": datetime.utcnow().isoformat(),
                        }
            except Exception as e:
                logger.error(f"❌ Loki query error: {e}", exc_info=True)
                return {
                    "status": "error",
                    "query": query,
                    "error": str(e),
                    "timestamp": datetime.utcnow().isoformat(),
                }

    async def query_labels(self, start: Optional[datetime] = None, end: Optional[datetime] = None) -> List[str]:
        """Get available label names"""
        if not start:
            start = datetime.utcnow() - timedelta(hours=1)
        if not end:
            end = datetime.utcnow()

        start_ns = int(start.timestamp() * 1e9)
        end_ns = int(end.timestamp() * 1e9)

        params = {"start": start_ns, "end": end_ns}

        async with aiohttp.ClientSession() as session:
            try:
                async with session.get(
                    f"{self.url}/loki/api/v1/labels",
                    params=params,
                    timeout=aiohttp.ClientTimeout(total=self.timeout),
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return data.get("data", [])
                    return []
            except Exception as e:
                logger.error(f"❌ Error getting Loki labels: {e}")
                return []

    async def query_label_values(
        self, label: str, start: Optional[datetime] = None, end: Optional[datetime] = None
    ) -> List[str]:
        """Get values for a specific label"""
        if not start:
            start = datetime.utcnow() - timedelta(hours=1)
        if not end:
            end = datetime.utcnow()

        start_ns = int(start.timestamp() * 1e9)
        end_ns = int(end.timestamp() * 1e9)

        params = {"start": start_ns, "end": end_ns}

        async with aiohttp.ClientSession() as session:
            try:
                async with session.get(
                    f"{self.url}/loki/api/v1/label/{label}/values",
                    params=params,
                    timeout=aiohttp.ClientTimeout(total=self.timeout),
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return data.get("data", [])
                    return []
            except Exception as e:
                logger.error(f"❌ Error getting Loki label values: {e}")
                return []

    async def query_stats(self, query: str, start: Optional[datetime] = None, end: Optional[datetime] = None) -> Dict[str, Any]:
        """Get statistics about log streams matching a selector"""
        if not start:
            start = datetime.utcnow() - timedelta(minutes=30)
        if not end:
            end = datetime.utcnow()

        start_ns = int(start.timestamp() * 1e9)
        end_ns = int(end.timestamp() * 1e9)

        params = {"query": query, "start": start_ns, "end": end_ns}

        async with aiohttp.ClientSession() as session:
            try:
                async with session.get(
                    f"{self.url}/loki/api/v1/index/stats",
                    params=params,
                    timeout=aiohttp.ClientTimeout(total=self.timeout),
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return data.get("data", {})
                    else:
                        logger.warning(f"Stats query returned {response.status}")
                        return {}
            except Exception as e:
                logger.error(f"❌ Error getting Loki stats: {e}")
                return {}

