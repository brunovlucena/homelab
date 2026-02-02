"""
Base test class for agent testing.

Provides common test setup, fixtures, and helper methods
for testing any agent in the homelab infrastructure.
"""

import pytest
from typing import Any, Optional
from unittest.mock import AsyncMock, MagicMock, patch
from datetime import datetime, timezone
from uuid import uuid4


class BaseAgentTest:
    """
    Base class for testing homelab agents.
    
    Provides common utilities for:
    - Setting up mock dependencies
    - Creating test events
    - Asserting common patterns
    - Cleaning up test resources
    
    Usage:
        class TestMyAgent(BaseAgentTest):
            @pytest.fixture(autouse=True)
            def setup(self, mock_k8s_client):
                from myagent.handler import MyHandler
                self.handler = MyHandler(k8s_client=mock_k8s_client)
            
            async def test_processes_event(self):
                event = self.create_event("test.event", {"key": "value"})
                result = await self.handler.process(event)
                assert result.success
    """
    
    # Override in subclass to set agent name
    agent_name: str = "test-agent"
    agent_namespace: str = "test-namespace"
    
    def create_event(
        self,
        event_type: str,
        data: Optional[dict] = None,
        source: Optional[str] = None,
        event_id: Optional[str] = None,
    ) -> dict:
        """Create a CloudEvent for testing."""
        return {
            "specversion": "1.0",
            "type": event_type,
            "source": source or f"/{self.agent_name}",
            "id": event_id or str(uuid4()),
            "time": datetime.now(timezone.utc).isoformat(),
            "datacontenttype": "application/json",
            "data": data or {},
        }
    
    def create_http_headers(
        self,
        event_type: str,
        source: Optional[str] = None,
        event_id: Optional[str] = None,
    ) -> dict:
        """Create CloudEvent HTTP headers."""
        return {
            "ce-specversion": "1.0",
            "ce-type": event_type,
            "ce-source": source or f"/{self.agent_name}",
            "ce-id": event_id or str(uuid4()),
            "ce-time": datetime.now(timezone.utc).isoformat(),
            "content-type": "application/json",
        }
    
    def assert_event_response(
        self,
        response: dict,
        expected_type: str,
        success: bool = True,
    ):
        """Assert event response follows expected format."""
        assert "type" in response
        assert response["type"] == expected_type
        if success:
            assert response.get("data", {}).get("status") != "error"
    
    def assert_success_response(self, response: dict):
        """Assert response indicates success."""
        status = response.get("status") or response.get("data", {}).get("status")
        assert status in ("success", "ok", "completed", "processed")
    
    def assert_error_response(
        self,
        response: dict,
        expected_error: Optional[str] = None,
    ):
        """Assert response indicates error."""
        status = response.get("status") or response.get("data", {}).get("status")
        error = response.get("error") or response.get("data", {}).get("error")
        
        assert status in ("error", "failed", "failure")
        if expected_error:
            assert expected_error.lower() in str(error).lower()
    
    def mock_env(self, **env_vars) -> "patch":
        """Create a context manager to mock environment variables."""
        import os
        return patch.dict(os.environ, env_vars)
    
    @staticmethod
    def create_mock_handler(
        async_methods: Optional[list[str]] = None,
        sync_methods: Optional[list[str]] = None,
    ) -> MagicMock:
        """Create a mock handler with specified methods."""
        handler = MagicMock()
        
        for method in async_methods or []:
            setattr(handler, method, AsyncMock())
        
        for method in sync_methods or []:
            setattr(handler, method, MagicMock())
        
        return handler


class BaseAsyncAgentTest(BaseAgentTest):
    """
    Base class for testing async agents.
    
    Extends BaseAgentTest with async-specific utilities.
    """
    
    @pytest.fixture(autouse=True)
    def setup_async(self, event_loop):
        """Setup async test environment."""
        self.loop = event_loop
    
    async def run_handler_with_timeout(
        self,
        handler_coro,
        timeout: float = 5.0,
    ) -> Any:
        """Run handler with timeout to prevent hanging tests."""
        import asyncio
        
        try:
            return await asyncio.wait_for(handler_coro, timeout=timeout)
        except asyncio.TimeoutError:
            pytest.fail(f"Handler timed out after {timeout} seconds")
    
    async def assert_raises_timeout(
        self,
        handler_coro,
        timeout: float = 1.0,
    ):
        """Assert that handler times out."""
        import asyncio
        
        with pytest.raises(asyncio.TimeoutError):
            await asyncio.wait_for(handler_coro, timeout=timeout)


class BaseKnativeAgentTest(BaseAgentTest):
    """
    Base class for testing Knative Lambda agents.
    
    Provides utilities specific to Knative serverless functions.
    """
    
    function_name: str = "test-function"
    function_version: str = "1.0.0"
    
    def create_knative_request(
        self,
        data: dict,
        method: str = "POST",
        path: str = "/",
    ) -> dict:
        """Create a mock Knative HTTP request."""
        return {
            "method": method,
            "path": path,
            "headers": self.create_http_headers(
                event_type=f"io.homelab.{self.agent_name}.request"
            ),
            "body": data,
        }
    
    def assert_knative_response(
        self,
        response: dict,
        expected_status: int = 200,
    ):
        """Assert Knative function response format."""
        assert "statusCode" in response or response.get("status_code") == expected_status
        if "body" in response:
            assert response["body"] is not None


class BaseSecurityAgentTest(BaseAgentTest):
    """
    Base class for testing security-focused agents (redteam, blueteam).
    
    Provides utilities for security testing scenarios.
    """
    
    def create_exploit_event(
        self,
        exploit_id: str,
        status: str = "success",
        severity: str = "high",
        **kwargs,
    ) -> dict:
        """Create an exploit event."""
        return self.create_event(
            event_type=f"io.homelab.exploit.{status}",
            data={
                "exploit_id": exploit_id,
                "status": status,
                "severity": severity,
                "namespace": kwargs.get("namespace", self.agent_namespace),
                **kwargs,
            },
            source="/agent-redteam/exploit-runner",
        )
    
    def create_defense_event(
        self,
        threat_type: str,
        action: str = "blocked",
        **kwargs,
    ) -> dict:
        """Create a defense event."""
        return self.create_event(
            event_type="io.homelab.defense.activated",
            data={
                "threat_type": threat_type,
                "action": action,
                **kwargs,
            },
            source="/agent-blueteam/defense-runner",
        )
    
    def assert_exploit_blocked(self, result: dict):
        """Assert an exploit was blocked."""
        status = result.get("status") or result.get("data", {}).get("status")
        assert status == "blocked"
    
    def assert_threat_detected(self, result: dict, threat_type: str):
        """Assert a threat was detected."""
        detected = result.get("threat_type") or result.get("data", {}).get("threat_type")
        assert detected == threat_type


class BaseChatAgentTest(BaseAgentTest):
    """
    Base class for testing chat/conversational agents.
    
    Provides utilities for testing LLM-based chat agents.
    """
    
    default_model: str = "llama3.2:3b"
    
    def create_chat_event(
        self,
        message: str,
        conversation_id: Optional[str] = None,
        user_id: str = "test-user",
    ) -> dict:
        """Create a chat message event."""
        return self.create_event(
            event_type="io.homelab.chat.message",
            data={
                "message": message,
                "conversation_id": conversation_id or str(uuid4()),
                "user_id": user_id,
            },
        )
    
    def assert_chat_response(
        self,
        response: dict,
        min_length: int = 1,
    ):
        """Assert chat response is valid."""
        content = (
            response.get("response") or
            response.get("message") or
            response.get("data", {}).get("response")
        )
        assert content is not None
        assert len(content) >= min_length
    
    def assert_conversation_maintained(
        self,
        response: dict,
        expected_conversation_id: str,
    ):
        """Assert conversation context is maintained."""
        conv_id = (
            response.get("conversation_id") or
            response.get("data", {}).get("conversation_id")
        )
        assert conv_id == expected_conversation_id
