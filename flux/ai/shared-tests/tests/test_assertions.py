"""Tests for shared assertions."""
import pytest
from shared_tests.assertions.cloudevent import (
    assert_cloudevent_valid,
    assert_cloudevent_type,
    assert_cloudevent_source,
    assert_cloudevent_data_contains,
)
from shared_tests.assertions.health import (
    assert_health_check_passes,
    assert_health_check_fails,
    assert_dependency_healthy,
)
from shared_tests.assertions.kubernetes import (
    assert_k8s_resource_labels,
    assert_k8s_resource_annotations,
)


class TestCloudEventAssertions:
    """Tests for CloudEvent assertions."""
    
    def test_assert_cloudevent_valid_passes(self):
        """Test valid CloudEvent passes validation."""
        event = {
            "specversion": "1.0",
            "type": "io.homelab.test.event",
            "source": "/test",
            "id": "123",
            "time": "2024-01-01T00:00:00Z",
            "datacontenttype": "application/json",
            "data": {}
        }
        
        assert_cloudevent_valid(event)  # Should not raise
    
    def test_assert_cloudevent_valid_fails_missing_field(self):
        """Test invalid CloudEvent fails validation."""
        event = {
            "specversion": "1.0",
            "type": "io.homelab.test.event",
            # Missing "source" and "id"
        }
        
        with pytest.raises(AssertionError, match="missing required field"):
            assert_cloudevent_valid(event)
    
    def test_assert_cloudevent_valid_fails_wrong_specversion(self):
        """Test wrong specversion fails validation."""
        event = {
            "specversion": "0.3",  # Wrong version
            "type": "io.homelab.test.event",
            "source": "/test",
            "id": "123",
        }
        
        with pytest.raises(AssertionError, match="Invalid specversion"):
            assert_cloudevent_valid(event)
    
    def test_assert_cloudevent_type_passes(self):
        """Test type assertion passes with correct type."""
        event = {"type": "io.homelab.chat.message"}
        
        assert_cloudevent_type(event, "io.homelab.chat.message")
    
    def test_assert_cloudevent_type_with_shorthand(self):
        """Test type assertion with shorthand type."""
        event = {"type": "io.homelab.chat.message"}
        
        assert_cloudevent_type(event, "chat.message")
    
    def test_assert_cloudevent_type_fails(self):
        """Test type assertion fails with wrong type."""
        event = {"type": "io.homelab.chat.response"}
        
        with pytest.raises(AssertionError, match="type mismatch"):
            assert_cloudevent_type(event, "io.homelab.chat.message")
    
    def test_assert_cloudevent_source_passes(self):
        """Test source assertion passes."""
        event = {"source": "/agent-bruno/chatbot"}
        
        assert_cloudevent_source(event, "/agent-bruno/chatbot")
    
    def test_assert_cloudevent_source_partial_match(self):
        """Test source assertion with partial match."""
        event = {"source": "/agent-bruno/chatbot"}
        
        assert_cloudevent_source(event, "agent-bruno", partial_match=True)
    
    def test_assert_cloudevent_data_contains_passes(self):
        """Test data contains assertion passes."""
        event = {
            "data": {
                "message": "Hello",
                "user_id": "123",
                "extra": "field"
            }
        }
        
        assert_cloudevent_data_contains(event, {
            "message": "Hello",
            "user_id": "123"
        })
    
    def test_assert_cloudevent_data_contains_fails_missing_key(self):
        """Test data contains assertion fails with missing key."""
        event = {"data": {"message": "Hello"}}
        
        with pytest.raises(AssertionError, match="missing key"):
            assert_cloudevent_data_contains(event, {"user_id": "123"})


class TestHealthAssertions:
    """Tests for health check assertions."""
    
    def test_assert_health_check_passes_dict(self):
        """Test health check passes with healthy dict."""
        response = {"status": "healthy"}
        
        assert_health_check_passes(response)
    
    def test_assert_health_check_passes_ok_status(self):
        """Test health check passes with 'ok' status."""
        response = {"status": "ok"}
        
        assert_health_check_passes(response)
    
    def test_assert_health_check_passes_bool(self):
        """Test health check passes with True."""
        assert_health_check_passes(True)
    
    def test_assert_health_check_fails_dict(self):
        """Test health check fails with unhealthy dict."""
        response = {"status": "unhealthy", "reason": "Database down"}
        
        assert_health_check_fails(response)
    
    def test_assert_health_check_fails_with_reason(self):
        """Test health check fails with expected reason."""
        response = {"status": "unhealthy", "reason": "Connection refused"}
        
        assert_health_check_fails(response, expected_reason="connection")
    
    def test_assert_dependency_healthy_passes(self):
        """Test dependency healthy assertion passes."""
        response = {
            "status": "healthy",
            "checks": {
                "database": {"status": "healthy"},
                "redis": {"status": "ok"}
            }
        }
        
        assert_dependency_healthy(response, "database")
        assert_dependency_healthy(response, "redis")
    
    def test_assert_dependency_healthy_fails_not_found(self):
        """Test dependency healthy assertion fails when not found."""
        response = {
            "checks": {
                "database": {"status": "healthy"}
            }
        }
        
        with pytest.raises(AssertionError, match="not found"):
            assert_dependency_healthy(response, "redis")


class TestKubernetesAssertions:
    """Tests for Kubernetes assertions."""
    
    def test_assert_k8s_resource_labels_passes(self):
        """Test resource labels assertion passes."""
        resource = {
            "metadata": {
                "labels": {
                    "app": "test-app",
                    "version": "1.0.0",
                    "env": "test"
                }
            }
        }
        
        assert_k8s_resource_labels(resource, {
            "app": "test-app",
            "version": "1.0.0"
        })
    
    def test_assert_k8s_resource_labels_fails_missing(self):
        """Test resource labels assertion fails with missing label."""
        resource = {
            "metadata": {
                "labels": {
                    "app": "test-app"
                }
            }
        }
        
        with pytest.raises(AssertionError, match="missing label"):
            assert_k8s_resource_labels(resource, {"version": "1.0.0"})
    
    def test_assert_k8s_resource_labels_fails_wrong_value(self):
        """Test resource labels assertion fails with wrong value."""
        resource = {
            "metadata": {
                "labels": {
                    "app": "wrong-app"
                }
            }
        }
        
        with pytest.raises(AssertionError, match="value mismatch"):
            assert_k8s_resource_labels(resource, {"app": "test-app"})
    
    def test_assert_k8s_resource_annotations_passes(self):
        """Test resource annotations assertion passes."""
        resource = {
            "metadata": {
                "annotations": {
                    "description": "Test resource",
                    "owner": "test-team"
                }
            }
        }
        
        assert_k8s_resource_annotations(resource, {
            "description": "Test resource"
        })
