"""Tests for base test classes."""
import pytest
from shared_tests.base.agent import (
    BaseAgentTest,
    BaseSecurityAgentTest,
    BaseChatAgentTest,
)
from shared_tests.base.cloudevent import (
    BaseCloudEventTest,
    BaseEventDrivenTest,
)
from shared_tests.base.health import (
    BaseHealthCheckTest,
    BaseK8sProbeTest,
)
from shared_tests.base.metrics import (
    BaseMetricsTest,
    BaseAgentMetricsTest,
)


class TestBaseAgentTest:
    """Tests for BaseAgentTest class."""
    
    def test_create_event(self):
        """Test creating an event."""
        base = BaseAgentTest()
        event = base.create_event(
            event_type="io.homelab.test.event",
            data={"key": "value"}
        )
        
        assert event["specversion"] == "1.0"
        assert event["type"] == "io.homelab.test.event"
        assert event["data"] == {"key": "value"}
        assert "id" in event
    
    def test_create_http_headers(self):
        """Test creating HTTP headers."""
        base = BaseAgentTest()
        headers = base.create_http_headers(
            event_type="io.homelab.test.event"
        )
        
        assert "ce-specversion" in headers
        assert "ce-type" in headers
        assert headers["ce-type"] == "io.homelab.test.event"
    
    def test_assert_success_response(self):
        """Test success response assertion."""
        base = BaseAgentTest()
        
        # These should pass
        base.assert_success_response({"status": "success"})
        base.assert_success_response({"status": "ok"})
        base.assert_success_response({"data": {"status": "completed"}})
    
    def test_assert_error_response(self):
        """Test error response assertion."""
        base = BaseAgentTest()
        
        base.assert_error_response({"status": "error"})
        base.assert_error_response(
            {"status": "error", "error": "Connection failed"},
            expected_error="connection"
        )


class TestBaseSecurityAgentTest:
    """Tests for BaseSecurityAgentTest class."""
    
    def test_create_exploit_event(self):
        """Test creating an exploit event."""
        base = BaseSecurityAgentTest()
        event = base.create_exploit_event(
            exploit_id="vuln-001",
            status="success",
            severity="critical"
        )
        
        assert event["type"] == "io.homelab.exploit.success"
        assert event["data"]["exploit_id"] == "vuln-001"
        assert event["data"]["severity"] == "critical"
    
    def test_create_defense_event(self):
        """Test creating a defense event."""
        base = BaseSecurityAgentTest()
        event = base.create_defense_event(
            threat_type="ssrf",
            action="blocked"
        )
        
        assert event["type"] == "io.homelab.defense.activated"
        assert event["data"]["threat_type"] == "ssrf"
    
    def test_assert_exploit_blocked(self):
        """Test exploit blocked assertion."""
        base = BaseSecurityAgentTest()
        
        result = {"status": "blocked"}
        base.assert_exploit_blocked(result)
    
    def test_assert_threat_detected(self):
        """Test threat detected assertion."""
        base = BaseSecurityAgentTest()
        
        result = {"threat_type": "ssrf"}
        base.assert_threat_detected(result, "ssrf")


class TestBaseChatAgentTest:
    """Tests for BaseChatAgentTest class."""
    
    def test_create_chat_event(self):
        """Test creating a chat event."""
        base = BaseChatAgentTest()
        event = base.create_chat_event(
            message="Hello!",
            user_id="user-123"
        )
        
        assert event["type"] == "io.homelab.chat.message"
        assert event["data"]["message"] == "Hello!"
        assert event["data"]["user_id"] == "user-123"
    
    def test_assert_chat_response(self):
        """Test chat response assertion."""
        base = BaseChatAgentTest()
        
        response = {"response": "Hello back!"}
        base.assert_chat_response(response)
        
        response = {"data": {"response": "Hi there!"}}
        base.assert_chat_response(response)
    
    def test_assert_conversation_maintained(self):
        """Test conversation maintained assertion."""
        base = BaseChatAgentTest()
        
        response = {"conversation_id": "conv-123"}
        base.assert_conversation_maintained(response, "conv-123")


class TestBaseCloudEventTest:
    """Tests for BaseCloudEventTest class."""
    
    def test_create_cloudevent(self):
        """Test creating a CloudEvent."""
        base = BaseCloudEventTest()
        event = base.create_cloudevent(
            event_type="health",  # Shorthand
            data={"status": "ok"}
        )
        
        assert event["type"] == "io.homelab.health.check"
    
    def test_create_cloudevent_batch(self):
        """Test creating a batch of CloudEvents."""
        base = BaseCloudEventTest()
        
        events = base.create_cloudevent_batch(
            event_type="health",
            count=5,
            data_factory=lambda i: {"index": i}
        )
        
        assert len(events) == 5
        for i, event in enumerate(events):
            assert event["data"]["index"] == i
    
    def test_assert_event_valid(self):
        """Test event validation."""
        base = BaseCloudEventTest()
        
        event = base.create_cloudevent("health")
        base.assert_event_valid(event)  # Should not raise
    
    def test_assert_event_processed(self):
        """Test event processed assertion."""
        base = BaseCloudEventTest()
        
        base.assert_event_processed({"status": "success"})
        base.assert_event_processed({"data": {"status": "success"}})


class TestBaseEventDrivenTest:
    """Tests for BaseEventDrivenTest class."""
    
    def test_create_event_with_tracing(self):
        """Test creating event with tracing."""
        base = BaseEventDrivenTest()
        event = base.create_event_with_tracing(
            event_type="health",
            data={}
        )
        
        assert "traceparent" in event
    
    def test_create_retry_event(self):
        """Test creating a retry event."""
        base = BaseEventDrivenTest()
        
        original = base.create_cloudevent("health")
        retry = base.create_retry_event(original, retry_count=2)
        
        assert retry["extensions"]["retrycount"] == 2
        assert retry["id"] != original["id"]
    
    def test_create_event_sequence(self):
        """Test creating an event sequence."""
        base = BaseEventDrivenTest()
        
        sequence = base.create_event_sequence(
            types=["exploit_started", "exploit_success"],
            data={"target": "test"}
        )
        
        assert len(sequence) == 2
        # All events should share the same sequence_id
        assert sequence[0]["data"]["sequence_id"] == sequence[1]["data"]["sequence_id"]


class TestBaseHealthCheckTest:
    """Tests for BaseHealthCheckTest class."""
    
    def test_create_health_response(self):
        """Test creating a health response."""
        base = BaseHealthCheckTest()
        
        response = base.create_health_response(healthy=True)
        assert response["status"] == "healthy"
        
        response = base.create_health_response(healthy=False, message="DB down")
        assert response["status"] == "unhealthy"
        assert response["message"] == "DB down"
    
    def test_assert_healthy(self):
        """Test healthy assertion."""
        base = BaseHealthCheckTest()
        
        base.assert_healthy({"status": "healthy"})
        base.assert_healthy({"status": "ok"})
        base.assert_healthy(True)
    
    def test_assert_unhealthy(self):
        """Test unhealthy assertion."""
        base = BaseHealthCheckTest()
        
        base.assert_unhealthy({"status": "unhealthy"})
        base.assert_unhealthy(
            {"status": "error", "reason": "Connection refused"},
            expected_reason="connection"
        )
    
    def test_assert_readiness_check(self):
        """Test readiness check assertion."""
        base = BaseHealthCheckTest()
        
        response = {
            "status": "healthy",
            "checks": {
                "database": {"status": "healthy"},
                "redis": {"status": "healthy"}
            }
        }
        
        base.assert_readiness_check(response, expected_checks=["database", "redis"])


class TestBaseMetricsTest:
    """Tests for BaseMetricsTest class."""
    
    def test_snapshot_and_assert_incremented(self):
        """Test snapshot and increment assertion."""
        base = BaseMetricsTest()
        
        # This tests the snapshot mechanism
        base.snapshot_metric("test_metric")
        
        # Can't really test increment without actual prometheus_client
        # but we can test the snapshot storage
        assert "test_metric:None" in base._snapshots
