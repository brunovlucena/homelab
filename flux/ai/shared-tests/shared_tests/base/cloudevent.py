"""
Base test class for CloudEvent handler testing.

Provides specialized utilities for testing event-driven agents
that process CloudEvents.
"""

import pytest
from typing import Any, Optional
from datetime import datetime, timezone
from uuid import uuid4


class BaseCloudEventTest:
    """
    Base class for testing CloudEvent handlers.
    
    Provides utilities for:
    - Creating and validating CloudEvents
    - Testing event routing
    - Testing event acknowledgment patterns
    
    Usage:
        class TestMyEventHandler(BaseCloudEventTest):
            async def test_handles_event(self, cloudevent_factory):
                event = cloudevent_factory.create(
                    type="io.homelab.test.action",
                    data={"key": "value"}
                )
                result = await self.handler.process(event)
                self.assert_event_processed(result)
    """
    
    # Common CloudEvent types
    EVENT_TYPES = {
        # Lifecycle events
        "health": "io.homelab.health.check",
        "ready": "io.homelab.ready",
        "shutdown": "io.homelab.shutdown",
        
        # Chat events
        "chat_message": "io.homelab.chat.message",
        "chat_response": "io.homelab.chat.response",
        
        # Security events
        "exploit_started": "io.homelab.exploit.started",
        "exploit_success": "io.homelab.exploit.success",
        "exploit_blocked": "io.homelab.exploit.blocked",
        "defense_activated": "io.homelab.defense.activated",
        
        # Contract events
        "contract_created": "io.homelab.contract.created",
        "contract_scanned": "io.homelab.contract.scanned",
        "vulnerability_found": "io.homelab.vulnerability.found",
    }
    
    def create_cloudevent(
        self,
        event_type: str,
        data: Optional[dict] = None,
        source: str = "/test",
        subject: Optional[str] = None,
    ) -> dict:
        """Create a CloudEvent dictionary."""
        # Resolve shorthand type names
        resolved_type = self.EVENT_TYPES.get(event_type, event_type)
        
        event = {
            "specversion": "1.0",
            "type": resolved_type,
            "source": source,
            "id": str(uuid4()),
            "time": datetime.now(timezone.utc).isoformat(),
            "datacontenttype": "application/json",
            "data": data or {},
        }
        
        if subject:
            event["subject"] = subject
        
        return event
    
    def create_cloudevent_batch(
        self,
        event_type: str,
        count: int,
        data_factory: Optional[callable] = None,
    ) -> list[dict]:
        """Create a batch of CloudEvents."""
        events = []
        for i in range(count):
            data = data_factory(i) if data_factory else {"index": i}
            events.append(self.create_cloudevent(event_type, data))
        return events
    
    def assert_event_valid(self, event: dict):
        """Assert CloudEvent has valid required fields."""
        required_fields = ["specversion", "type", "source", "id"]
        for field in required_fields:
            assert field in event, f"Missing required field: {field}"
        
        assert event["specversion"] == "1.0"
        assert event["type"].startswith("io.homelab.")
    
    def assert_event_processed(
        self,
        result: Any,
        expected_status: str = "success",
    ):
        """Assert event was processed successfully."""
        if isinstance(result, dict):
            status = (
                result.get("status") or
                result.get("data", {}).get("status") or
                "success"
            )
            assert status == expected_status, f"Expected {expected_status}, got {status}"
        elif hasattr(result, "status"):
            assert result.status == expected_status
        elif hasattr(result, "success"):
            expected_bool = expected_status in ("success", "ok", "completed")
            assert result.success == expected_bool
    
    def assert_event_rejected(
        self,
        result: Any,
        reason: Optional[str] = None,
    ):
        """Assert event was rejected."""
        if isinstance(result, dict):
            status = result.get("status") or result.get("data", {}).get("status")
            assert status in ("rejected", "error", "failed")
            
            if reason:
                error_msg = (
                    result.get("error") or
                    result.get("reason") or
                    result.get("data", {}).get("error")
                )
                assert reason.lower() in str(error_msg).lower()
    
    def assert_event_forwarded(
        self,
        result: Any,
        target: Optional[str] = None,
    ):
        """Assert event was forwarded to another handler."""
        if isinstance(result, dict):
            status = result.get("status") or result.get("data", {}).get("status")
            assert status == "forwarded"
            
            if target:
                forwarded_to = (
                    result.get("target") or
                    result.get("forwarded_to") or
                    result.get("data", {}).get("target")
                )
                assert forwarded_to == target
    
    def assert_event_response_type(
        self,
        response: dict,
        expected_type: str,
    ):
        """Assert response event has expected type."""
        response_type = response.get("type")
        expected = self.EVENT_TYPES.get(expected_type, expected_type)
        assert response_type == expected
    
    def create_event_sequence(
        self,
        types: list[str],
        data: Optional[dict] = None,
    ) -> list[dict]:
        """Create a sequence of related events."""
        sequence_id = str(uuid4())
        events = []
        
        for i, event_type in enumerate(types):
            event_data = {**(data or {}), "sequence_id": sequence_id, "sequence_index": i}
            events.append(self.create_cloudevent(event_type, event_data))
        
        return events


class BaseEventDrivenTest(BaseCloudEventTest):
    """
    Extended base class for event-driven agent testing.
    
    Adds utilities for testing event routing, filtering,
    and complex event processing patterns.
    """
    
    def create_event_with_tracing(
        self,
        event_type: str,
        data: Optional[dict] = None,
        parent_id: Optional[str] = None,
    ) -> dict:
        """Create a CloudEvent with tracing context."""
        event = self.create_cloudevent(event_type, data)
        event["traceparent"] = f"00-{uuid4().hex}-{uuid4().hex[:16]}-01"
        
        if parent_id:
            event["tracestate"] = f"parent_id={parent_id}"
        
        return event
    
    def assert_event_traced(self, event: dict):
        """Assert event has valid tracing information."""
        assert "traceparent" in event or "trace_id" in event.get("data", {})
    
    def create_retry_event(
        self,
        original_event: dict,
        retry_count: int = 1,
    ) -> dict:
        """Create a retry event based on original event."""
        retry_event = original_event.copy()
        retry_event["id"] = str(uuid4())  # New ID for retry
        retry_event.setdefault("extensions", {})
        retry_event["extensions"]["retrycount"] = retry_count
        return retry_event
    
    def assert_idempotent(
        self,
        results: list[Any],
    ):
        """Assert multiple event processing is idempotent."""
        # All results should be equivalent for idempotent processing
        first_result = results[0]
        for result in results[1:]:
            assert self._compare_results(first_result, result), (
                "Event processing is not idempotent"
            )
    
    def _compare_results(self, r1: Any, r2: Any) -> bool:
        """Compare two results for idempotency checking."""
        if isinstance(r1, dict) and isinstance(r2, dict):
            # Compare ignoring timestamps and IDs
            ignore_keys = {"id", "timestamp", "time", "created_at", "updated_at"}
            d1 = {k: v for k, v in r1.items() if k not in ignore_keys}
            d2 = {k: v for k, v in r2.items() if k not in ignore_keys}
            return d1 == d2
        return r1 == r2
