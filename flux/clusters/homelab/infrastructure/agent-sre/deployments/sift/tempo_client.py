"""
🔍 Tempo query client
Provides trace querying capabilities for Grafana Tempo
"""

import logging
from datetime import datetime, timedelta
from typing import Any, Dict, List, Optional

import aiohttp

logger = logging.getLogger(__name__)


class TempoClient:
    """Client for querying Grafana Tempo"""

    def __init__(self, url: str, timeout: int = 30):
        """Initialize Tempo client"""
        self.url = url.rstrip("/")
        self.timeout = timeout
        logger.info(f"🔍 Initialized Tempo client: {self.url}")

    async def search_traces(
        self,
        tags: Optional[Dict[str, str]] = None,
        start: Optional[datetime] = None,
        end: Optional[datetime] = None,
        min_duration: Optional[str] = None,
        max_duration: Optional[str] = None,
        limit: int = 20,
    ) -> Dict[str, Any]:
        """
        Search for traces in Tempo

        Args:
            tags: Key-value pairs to filter traces (e.g., {"service.name": "api", "http.status_code": "500"})
            start: Start time (defaults to 30 minutes ago)
            end: End time (defaults to now)
            min_duration: Minimum duration (e.g., "100ms", "1s")
            max_duration: Maximum duration
            limit: Maximum number of traces to return

        Returns:
            Dictionary with search results
        """
        if not start:
            start = datetime.utcnow() - timedelta(minutes=30)
        if not end:
            end = datetime.utcnow()

        # Convert to Unix timestamps (seconds)
        start_unix = int(start.timestamp())
        end_unix = int(end.timestamp())

        params = {
            "start": start_unix,
            "end": end_unix,
            "limit": limit,
        }

        # Add tag filters
        if tags:
            for key, value in tags.items():
                params["tags"] = f'{key}="{value}"'

        if min_duration:
            params["minDuration"] = min_duration
        if max_duration:
            params["maxDuration"] = max_duration

        logger.info(f"🔍 Searching Tempo traces: {tags} from {start} to {end}")

        async with aiohttp.ClientSession() as session:
            try:
                async with session.get(
                    f"{self.url}/api/search",
                    params=params,
                    timeout=aiohttp.ClientTimeout(total=self.timeout),
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return {
                            "status": "success",
                            "tags": tags or {},
                            "start": start.isoformat(),
                            "end": end.isoformat(),
                            "result": data,
                            "timestamp": datetime.utcnow().isoformat(),
                        }
                    else:
                        error_text = await response.text()
                        logger.error(f"❌ Tempo error: {response.status} - {error_text}")
                        return {
                            "status": "error",
                            "tags": tags or {},
                            "error": f"HTTP {response.status}: {error_text}",
                            "timestamp": datetime.utcnow().isoformat(),
                        }
            except Exception as e:
                logger.error(f"❌ Tempo search error: {e}", exc_info=True)
                return {
                    "status": "error",
                    "tags": tags or {},
                    "error": str(e),
                    "timestamp": datetime.utcnow().isoformat(),
                }

    async def get_trace(self, trace_id: str) -> Dict[str, Any]:
        """Get a specific trace by ID"""
        logger.info(f"🔍 Getting trace: {trace_id}")

        async with aiohttp.ClientSession() as session:
            try:
                async with session.get(
                    f"{self.url}/api/traces/{trace_id}",
                    timeout=aiohttp.ClientTimeout(total=self.timeout),
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return {
                            "status": "success",
                            "trace_id": trace_id,
                            "result": data,
                            "timestamp": datetime.utcnow().isoformat(),
                        }
                    else:
                        error_text = await response.text()
                        return {
                            "status": "error",
                            "trace_id": trace_id,
                            "error": f"HTTP {response.status}: {error_text}",
                            "timestamp": datetime.utcnow().isoformat(),
                        }
            except Exception as e:
                logger.error(f"❌ Error getting trace: {e}")
                return {
                    "status": "error",
                    "trace_id": trace_id,
                    "error": str(e),
                    "timestamp": datetime.utcnow().isoformat(),
                }

    async def search_tags(self) -> List[str]:
        """Get available tag names"""
        async with aiohttp.ClientSession() as session:
            try:
                async with session.get(
                    f"{self.url}/api/search/tags",
                    timeout=aiohttp.ClientTimeout(total=self.timeout),
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return data.get("tagNames", [])
                    return []
            except Exception as e:
                logger.error(f"❌ Error getting Tempo tags: {e}")
                return []

    async def search_tag_values(self, tag: str) -> List[str]:
        """Get values for a specific tag"""
        async with aiohttp.ClientSession() as session:
            try:
                async with session.get(
                    f"{self.url}/api/search/tag/{tag}/values",
                    timeout=aiohttp.ClientTimeout(total=self.timeout),
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return data.get("tagValues", [])
                    return []
            except Exception as e:
                logger.error(f"❌ Error getting Tempo tag values: {e}")
                return []
