"""
Base test class for health check testing.

Provides utilities for testing agent health endpoints,
readiness probes, and liveness probes.
"""

import pytest
from typing import Any, Optional
from unittest.mock import AsyncMock


class BaseHealthCheckTest:
    """
    Base class for testing agent health checks.
    
    Provides utilities for:
    - Testing liveness probes
    - Testing readiness probes
    - Testing startup probes
    - Validating health response formats
    
    Usage:
        class TestMyAgentHealth(BaseHealthCheckTest):
            @pytest.fixture(autouse=True)
            def setup(self):
                from myagent.main import app
                self.app = app
            
            async def test_liveness(self):
                response = await self.check_liveness()
                self.assert_healthy(response)
    """
    
    # Standard health check paths
    LIVENESS_PATH = "/healthz"
    READINESS_PATH = "/readyz"
    STARTUP_PATH = "/startupz"
    METRICS_PATH = "/metrics"
    
    # Standard response format
    HEALTHY_RESPONSE = {"status": "healthy"}
    UNHEALTHY_RESPONSE = {"status": "unhealthy"}
    
    def create_health_response(
        self,
        healthy: bool = True,
        message: Optional[str] = None,
        checks: Optional[dict] = None,
    ) -> dict:
        """Create a standard health response."""
        response = {
            "status": "healthy" if healthy else "unhealthy",
        }
        
        if message:
            response["message"] = message
        
        if checks:
            response["checks"] = checks
        
        return response
    
    def assert_healthy(self, response: Any):
        """Assert health response indicates healthy status."""
        if isinstance(response, dict):
            status = response.get("status", "").lower()
            assert status in ("healthy", "ok", "up", "pass"), (
                f"Expected healthy status, got: {status}"
            )
        elif hasattr(response, "status_code"):
            assert response.status_code == 200, (
                f"Expected 200 status code, got: {response.status_code}"
            )
        else:
            assert response is True or response == "healthy"
    
    def assert_unhealthy(
        self,
        response: Any,
        expected_reason: Optional[str] = None,
    ):
        """Assert health response indicates unhealthy status."""
        if isinstance(response, dict):
            status = response.get("status", "").lower()
            assert status in ("unhealthy", "error", "down", "fail"), (
                f"Expected unhealthy status, got: {status}"
            )
            
            if expected_reason:
                reason = response.get("reason") or response.get("message", "")
                assert expected_reason.lower() in reason.lower(), (
                    f"Expected reason containing '{expected_reason}', got: {reason}"
                )
        elif hasattr(response, "status_code"):
            assert response.status_code >= 400, (
                f"Expected error status code, got: {response.status_code}"
            )
    
    def assert_readiness_check(
        self,
        response: dict,
        expected_checks: Optional[list[str]] = None,
    ):
        """Assert readiness response contains expected checks."""
        if expected_checks:
            checks = response.get("checks", {})
            for check_name in expected_checks:
                assert check_name in checks, (
                    f"Missing readiness check: {check_name}"
                )
    
    def assert_all_checks_pass(self, response: dict):
        """Assert all health checks in response passed."""
        checks = response.get("checks", {})
        for check_name, check_result in checks.items():
            if isinstance(check_result, dict):
                status = check_result.get("status", "").lower()
                assert status in ("healthy", "ok", "pass"), (
                    f"Health check '{check_name}' failed: {check_result}"
                )
            elif isinstance(check_result, bool):
                assert check_result is True, (
                    f"Health check '{check_name}' failed"
                )
    
    def create_mock_dependency_checks(
        self,
        dependencies: dict[str, bool],
    ) -> dict:
        """Create mock dependency health checks."""
        checks = {}
        for dep_name, is_healthy in dependencies.items():
            checks[dep_name] = {
                "status": "healthy" if is_healthy else "unhealthy",
                "latency_ms": 10 if is_healthy else 0,
            }
        return checks


class BaseK8sProbeTest(BaseHealthCheckTest):
    """
    Specialized test class for Kubernetes probe testing.
    
    Tests health checks in the context of Kubernetes
    liveness, readiness, and startup probes.
    """
    
    def assert_liveness_probe(
        self,
        response: Any,
        max_latency_ms: float = 100.0,
    ):
        """Assert liveness probe response is valid for K8s."""
        self.assert_healthy(response)
        
        if isinstance(response, dict) and "latency_ms" in response:
            assert response["latency_ms"] <= max_latency_ms, (
                f"Liveness probe too slow: {response['latency_ms']}ms"
            )
    
    def assert_readiness_probe(
        self,
        response: Any,
        required_dependencies: Optional[list[str]] = None,
    ):
        """Assert readiness probe response is valid for K8s."""
        self.assert_healthy(response)
        
        if required_dependencies and isinstance(response, dict):
            checks = response.get("checks", {})
            for dep in required_dependencies:
                assert dep in checks, (
                    f"Readiness probe missing dependency check: {dep}"
                )
                self.assert_healthy(checks[dep])
    
    def assert_startup_probe(
        self,
        response: Any,
        initialization_complete: bool = True,
    ):
        """Assert startup probe response is valid for K8s."""
        if initialization_complete:
            self.assert_healthy(response)
        else:
            # During startup, unhealthy is expected
            if isinstance(response, dict):
                assert "status" in response


class BaseServiceHealthTest(BaseHealthCheckTest):
    """
    Specialized test class for service health testing.
    
    Tests health of services with external dependencies
    like databases, message queues, and external APIs.
    """
    
    # Common dependencies across agents
    COMMON_DEPENDENCIES = [
        "database",
        "redis",
        "rabbitmq",
        "ollama",
        "kubernetes",
    ]
    
    def create_dependency_health_checks(
        self,
        healthy_deps: Optional[list[str]] = None,
        unhealthy_deps: Optional[list[str]] = None,
    ) -> dict:
        """Create dependency health check results."""
        healthy_deps = healthy_deps or []
        unhealthy_deps = unhealthy_deps or []
        
        checks = {}
        
        for dep in healthy_deps:
            checks[dep] = {
                "status": "healthy",
                "latency_ms": 5.0,
                "last_check": "2024-01-01T00:00:00Z",
            }
        
        for dep in unhealthy_deps:
            checks[dep] = {
                "status": "unhealthy",
                "error": f"Connection to {dep} failed",
                "last_check": "2024-01-01T00:00:00Z",
            }
        
        return checks
    
    def assert_degraded_but_available(
        self,
        response: dict,
        failed_deps: list[str],
    ):
        """Assert service is degraded but still available."""
        status = response.get("status", "").lower()
        assert status in ("degraded", "healthy"), (
            f"Expected degraded or healthy status, got: {status}"
        )
        
        checks = response.get("checks", {})
        for dep in failed_deps:
            if dep in checks:
                dep_status = checks[dep].get("status", "")
                assert dep_status in ("unhealthy", "error"), (
                    f"Expected {dep} to be unhealthy"
                )
    
    async def test_graceful_degradation(
        self,
        handler,
        mock_failing_dependency: str,
    ):
        """Test that service handles dependency failure gracefully."""
        # This is a template method - implement in subclass
        raise NotImplementedError(
            "Implement test_graceful_degradation in subclass"
        )
