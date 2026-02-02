"""
HTTP client fixtures for testing API interactions.

Provides mocked httpx clients and response factories for
testing agents that make HTTP requests.
"""

import pytest
from unittest.mock import AsyncMock, MagicMock
from dataclasses import dataclass, field
from typing import Any, Optional
import json


@dataclass
class MockHTTPResponse:
    """Mock HTTP response for testing."""
    
    status_code: int = 200
    content: bytes = b""
    text: str = ""
    json_data: Optional[dict] = None
    headers: dict = field(default_factory=dict)
    
    def __post_init__(self):
        if self.json_data and not self.content:
            self.content = json.dumps(self.json_data).encode()
            self.text = json.dumps(self.json_data)
        if self.content and not self.text:
            self.text = self.content.decode()
    
    def json(self) -> dict:
        """Return JSON data."""
        if self.json_data:
            return self.json_data
        return json.loads(self.content)
    
    def raise_for_status(self):
        """Raise exception for error status codes."""
        if self.status_code >= 400:
            raise Exception(f"HTTP {self.status_code}: {self.text}")


class MockHTTPClientBuilder:
    """Builder for creating mock HTTP clients with configured responses."""
    
    def __init__(self):
        self._responses: dict[str, MockHTTPResponse] = {}
        self._default_response = MockHTTPResponse(
            status_code=200,
            json_data={"status": "ok"}
        )
    
    def with_response(
        self,
        method: str,
        url: str,
        status_code: int = 200,
        json_data: Optional[dict] = None,
        text: str = "",
    ) -> "MockHTTPClientBuilder":
        """Add a response for a specific method/url combination."""
        key = f"{method.upper()}:{url}"
        self._responses[key] = MockHTTPResponse(
            status_code=status_code,
            json_data=json_data,
            text=text,
        )
        return self
    
    def with_error(
        self,
        method: str,
        url: str,
        status_code: int = 500,
        error_message: str = "Internal Server Error",
    ) -> "MockHTTPClientBuilder":
        """Add an error response."""
        return self.with_response(
            method=method,
            url=url,
            status_code=status_code,
            json_data={"error": error_message},
        )
    
    def build(self) -> AsyncMock:
        """Build the mock HTTP client."""
        client = AsyncMock()
        
        async def request_handler(method: str, url: str, **kwargs):
            key = f"{method.upper()}:{url}"
            # Try exact match first
            if key in self._responses:
                return self._responses[key]
            # Try pattern matching (check if URL starts with registered patterns)
            for registered_key, response in self._responses.items():
                reg_method, reg_url = registered_key.split(":", 1)
                if method.upper() == reg_method and url.startswith(reg_url):
                    return response
            return self._default_response
        
        async def get_handler(url: str, **kwargs):
            return await request_handler("GET", url, **kwargs)
        
        async def post_handler(url: str, **kwargs):
            return await request_handler("POST", url, **kwargs)
        
        async def put_handler(url: str, **kwargs):
            return await request_handler("PUT", url, **kwargs)
        
        async def delete_handler(url: str, **kwargs):
            return await request_handler("DELETE", url, **kwargs)
        
        async def patch_handler(url: str, **kwargs):
            return await request_handler("PATCH", url, **kwargs)
        
        client.get = AsyncMock(side_effect=get_handler)
        client.post = AsyncMock(side_effect=post_handler)
        client.put = AsyncMock(side_effect=put_handler)
        client.delete = AsyncMock(side_effect=delete_handler)
        client.patch = AsyncMock(side_effect=patch_handler)
        client.request = AsyncMock(side_effect=request_handler)
        
        # Context manager support
        client.__aenter__ = AsyncMock(return_value=client)
        client.__aexit__ = AsyncMock(return_value=None)
        
        return client


@pytest.fixture
def mock_httpx_response():
    """Factory for creating mock HTTP responses."""
    def create(
        status_code: int = 200,
        json_data: Optional[dict] = None,
        text: str = "",
    ) -> MockHTTPResponse:
        return MockHTTPResponse(
            status_code=status_code,
            json_data=json_data,
            text=text,
        )
    return create


@pytest.fixture
def mock_httpx_client():
    """Basic mock httpx AsyncClient."""
    mock_response = MockHTTPResponse(
        status_code=200,
        json_data={"status": "ok"}
    )
    
    client = AsyncMock()
    client.get = AsyncMock(return_value=mock_response)
    client.post = AsyncMock(return_value=mock_response)
    client.put = AsyncMock(return_value=mock_response)
    client.delete = AsyncMock(return_value=mock_response)
    client.patch = AsyncMock(return_value=mock_response)
    
    client.__aenter__ = AsyncMock(return_value=client)
    client.__aexit__ = AsyncMock(return_value=None)
    
    return client


@pytest.fixture
def mock_httpx_client_builder():
    """Builder for creating configured mock HTTP clients."""
    return MockHTTPClientBuilder()


@pytest.fixture
def respx_mock():
    """
    Fixture for respx HTTP mocking.
    
    Usage:
        async def test_api_call(respx_mock):
            respx_mock.get("https://api.example.com/data").respond(
                json={"key": "value"}
            )
            # Make actual httpx request - will be intercepted
    """
    try:
        import respx
        with respx.mock(assert_all_called=False) as respx_mock:
            yield respx_mock
    except ImportError:
        # If respx is not installed, return a mock that raises helpful error
        mock = MagicMock()
        mock.get = MagicMock(
            side_effect=ImportError("respx not installed. Run: pip install respx")
        )
        mock.post = mock.get
        mock.put = mock.get
        mock.delete = mock.get
        yield mock


# Common HTTP response fixtures
@pytest.fixture
def http_success_response(mock_httpx_response):
    """Standard success response."""
    return mock_httpx_response(
        status_code=200,
        json_data={"status": "success", "message": "Operation completed"}
    )


@pytest.fixture
def http_error_response(mock_httpx_response):
    """Standard error response."""
    return mock_httpx_response(
        status_code=500,
        json_data={"error": "Internal Server Error", "code": "INTERNAL_ERROR"}
    )


@pytest.fixture
def http_not_found_response(mock_httpx_response):
    """404 Not Found response."""
    return mock_httpx_response(
        status_code=404,
        json_data={"error": "Not Found", "code": "NOT_FOUND"}
    )


@pytest.fixture
def http_unauthorized_response(mock_httpx_response):
    """401 Unauthorized response."""
    return mock_httpx_response(
        status_code=401,
        json_data={"error": "Unauthorized", "code": "UNAUTHORIZED"}
    )


@pytest.fixture  
def http_rate_limit_response(mock_httpx_response):
    """429 Rate Limit response."""
    return mock_httpx_response(
        status_code=429,
        json_data={"error": "Rate limit exceeded", "code": "RATE_LIMITED"}
    )
