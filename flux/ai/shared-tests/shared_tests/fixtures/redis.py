"""
Redis fixtures for testing agent caching and messaging.

Provides mocked Redis clients for testing agents that use
Redis for caching, pub/sub, or task queues.
"""

import pytest
from unittest.mock import AsyncMock, MagicMock
from dataclasses import dataclass, field
from typing import Any, Optional
from datetime import timedelta
import json


class MockRedisClient:
    """Mock async Redis client for testing."""
    
    def __init__(self):
        self._data: dict[str, Any] = {}
        self._expires: dict[str, float] = {}
        self._pubsub_channels: dict[str, list] = {}
        self._streams: dict[str, list] = {}
        
        # Setup async mock methods
        self.get = AsyncMock(side_effect=self._get)
        self.set = AsyncMock(side_effect=self._set)
        self.setex = AsyncMock(side_effect=self._setex)
        self.delete = AsyncMock(side_effect=self._delete)
        self.exists = AsyncMock(side_effect=self._exists)
        self.expire = AsyncMock(side_effect=self._expire)
        self.ttl = AsyncMock(side_effect=self._ttl)
        self.incr = AsyncMock(side_effect=self._incr)
        self.decr = AsyncMock(side_effect=self._decr)
        self.hget = AsyncMock(side_effect=self._hget)
        self.hset = AsyncMock(side_effect=self._hset)
        self.hgetall = AsyncMock(side_effect=self._hgetall)
        self.lpush = AsyncMock(side_effect=self._lpush)
        self.rpush = AsyncMock(side_effect=self._rpush)
        self.lpop = AsyncMock(side_effect=self._lpop)
        self.rpop = AsyncMock(side_effect=self._rpop)
        self.lrange = AsyncMock(side_effect=self._lrange)
        self.llen = AsyncMock(side_effect=self._llen)
        self.sadd = AsyncMock(side_effect=self._sadd)
        self.smembers = AsyncMock(side_effect=self._smembers)
        self.sismember = AsyncMock(side_effect=self._sismember)
        self.publish = AsyncMock(side_effect=self._publish)
        self.subscribe = AsyncMock(side_effect=self._subscribe)
        self.xadd = AsyncMock(side_effect=self._xadd)
        self.xread = AsyncMock(side_effect=self._xread)
        self.ping = AsyncMock(return_value=True)
        self.close = AsyncMock()
        self.flushdb = AsyncMock(side_effect=self._flushdb)
    
    async def _get(self, key: str) -> Optional[str]:
        return self._data.get(key)
    
    async def _set(self, key: str, value: Any, ex: Optional[int] = None) -> bool:
        self._data[key] = value
        if ex:
            self._expires[key] = ex
        return True
    
    async def _setex(self, key: str, seconds: int, value: Any) -> bool:
        self._data[key] = value
        self._expires[key] = seconds
        return True
    
    async def _delete(self, *keys: str) -> int:
        deleted = 0
        for key in keys:
            if key in self._data:
                del self._data[key]
                deleted += 1
                if key in self._expires:
                    del self._expires[key]
        return deleted
    
    async def _exists(self, *keys: str) -> int:
        return sum(1 for key in keys if key in self._data)
    
    async def _expire(self, key: str, seconds: int) -> bool:
        if key in self._data:
            self._expires[key] = seconds
            return True
        return False
    
    async def _ttl(self, key: str) -> int:
        return self._expires.get(key, -1)
    
    async def _incr(self, key: str) -> int:
        value = int(self._data.get(key, 0)) + 1
        self._data[key] = str(value)
        return value
    
    async def _decr(self, key: str) -> int:
        value = int(self._data.get(key, 0)) - 1
        self._data[key] = str(value)
        return value
    
    async def _hget(self, name: str, key: str) -> Optional[str]:
        hash_data = self._data.get(name, {})
        if isinstance(hash_data, dict):
            return hash_data.get(key)
        return None
    
    async def _hset(self, name: str, key: str, value: Any) -> int:
        if name not in self._data:
            self._data[name] = {}
        created = key not in self._data[name]
        self._data[name][key] = value
        return 1 if created else 0
    
    async def _hgetall(self, name: str) -> dict:
        return self._data.get(name, {})
    
    async def _lpush(self, name: str, *values: Any) -> int:
        if name not in self._data:
            self._data[name] = []
        for value in reversed(values):
            self._data[name].insert(0, value)
        return len(self._data[name])
    
    async def _rpush(self, name: str, *values: Any) -> int:
        if name not in self._data:
            self._data[name] = []
        self._data[name].extend(values)
        return len(self._data[name])
    
    async def _lpop(self, name: str) -> Optional[str]:
        lst = self._data.get(name, [])
        if lst:
            return lst.pop(0)
        return None
    
    async def _rpop(self, name: str) -> Optional[str]:
        lst = self._data.get(name, [])
        if lst:
            return lst.pop()
        return None
    
    async def _lrange(self, name: str, start: int, end: int) -> list:
        lst = self._data.get(name, [])
        if end == -1:
            return lst[start:]
        return lst[start:end + 1]
    
    async def _llen(self, name: str) -> int:
        return len(self._data.get(name, []))
    
    async def _sadd(self, name: str, *values: Any) -> int:
        if name not in self._data:
            self._data[name] = set()
        initial_size = len(self._data[name])
        self._data[name].update(values)
        return len(self._data[name]) - initial_size
    
    async def _smembers(self, name: str) -> set:
        return self._data.get(name, set())
    
    async def _sismember(self, name: str, value: Any) -> bool:
        return value in self._data.get(name, set())
    
    async def _publish(self, channel: str, message: Any) -> int:
        if channel not in self._pubsub_channels:
            self._pubsub_channels[channel] = []
        self._pubsub_channels[channel].append(message)
        return len(self._pubsub_channels.get(channel, []))
    
    async def _subscribe(self, *channels: str):
        for channel in channels:
            if channel not in self._pubsub_channels:
                self._pubsub_channels[channel] = []
        return True
    
    async def _xadd(
        self,
        name: str,
        fields: dict,
        id: str = "*",
    ) -> str:
        if name not in self._streams:
            self._streams[name] = []
        entry_id = f"{len(self._streams[name])}-0"
        self._streams[name].append((entry_id, fields))
        return entry_id
    
    async def _xread(
        self,
        streams: dict,
        count: Optional[int] = None,
        block: Optional[int] = None,
    ) -> list:
        result = []
        for stream_name, last_id in streams.items():
            entries = self._streams.get(stream_name, [])
            if entries:
                if count:
                    entries = entries[:count]
                result.append((stream_name, entries))
        return result
    
    async def _flushdb(self) -> bool:
        self._data.clear()
        self._expires.clear()
        self._pubsub_channels.clear()
        self._streams.clear()
        return True
    
    # Helper methods for testing
    def set_data(self, key: str, value: Any):
        """Directly set data for testing."""
        self._data[key] = value
    
    def get_data(self) -> dict:
        """Get all stored data."""
        return self._data.copy()


@pytest.fixture
def mock_redis_client():
    """Mock Redis client fixture."""
    return MockRedisClient()


@pytest.fixture
def mock_redis_with_data(mock_redis_client):
    """Mock Redis client with pre-populated test data."""
    mock_redis_client.set_data("test:key", "test_value")
    mock_redis_client.set_data("test:counter", "42")
    mock_redis_client.set_data("test:hash", {
        "field1": "value1",
        "field2": "value2",
    })
    mock_redis_client.set_data("test:list", ["item1", "item2", "item3"])
    mock_redis_client.set_data("test:set", {"member1", "member2"})
    return mock_redis_client


@pytest.fixture
def redis_cache_factory():
    """Factory for creating mock Redis cache instances."""
    def create(initial_data: Optional[dict] = None) -> MockRedisClient:
        client = MockRedisClient()
        if initial_data:
            for key, value in initial_data.items():
                client.set_data(key, value)
        return client
    return create
