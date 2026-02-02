"""
Health check assertions for testing.

Provides semantic assertions for validating agent health checks
and dependency status.
"""

from typing import Any, Optional


def assert_health_check_passes(
    response: Any,
    expected_status: str = "healthy",
):
    """
    Assert that a health check response indicates healthy status.
    
    Args:
        response: Health check response (dict or object)
        expected_status: Expected status string
    
    Raises:
        AssertionError: If health check doesn't indicate healthy
    """
    healthy_statuses = {"healthy", "ok", "up", "pass", "success"}
    
    if isinstance(response, dict):
        status = response.get("status", "").lower()
        assert status in healthy_statuses or status == expected_status.lower(), (
            f"Health check failed: expected healthy status, got '{status}'"
        )
    elif hasattr(response, "status_code"):
        assert response.status_code == 200, (
            f"Health check failed: expected 200, got {response.status_code}"
        )
    elif hasattr(response, "status"):
        assert response.status.lower() in healthy_statuses, (
            f"Health check failed: got status '{response.status}'"
        )
    elif isinstance(response, bool):
        assert response is True, "Health check failed: returned False"
    else:
        # Assume string or other truthy value
        assert response, f"Health check failed: {response}"


def assert_health_check_fails(
    response: Any,
    expected_reason: Optional[str] = None,
):
    """
    Assert that a health check response indicates unhealthy status.
    
    Args:
        response: Health check response (dict or object)
        expected_reason: Optional expected reason string
    
    Raises:
        AssertionError: If health check doesn't indicate unhealthy
    """
    unhealthy_statuses = {"unhealthy", "error", "down", "fail", "failure"}
    
    if isinstance(response, dict):
        status = response.get("status", "").lower()
        assert status in unhealthy_statuses, (
            f"Expected unhealthy status, got '{status}'"
        )
        
        if expected_reason:
            reason = (
                response.get("reason") or
                response.get("message") or
                response.get("error") or
                ""
            )
            assert expected_reason.lower() in reason.lower(), (
                f"Expected reason containing '{expected_reason}', got '{reason}'"
            )
    elif hasattr(response, "status_code"):
        assert response.status_code >= 400, (
            f"Expected error status code, got {response.status_code}"
        )
    elif isinstance(response, bool):
        assert response is False, "Expected health check to fail"


def assert_dependency_healthy(
    response: dict,
    dependency_name: str,
):
    """
    Assert that a specific dependency is healthy.
    
    Args:
        response: Health check response containing dependency checks
        dependency_name: Name of the dependency to check
    
    Raises:
        AssertionError: If dependency is not healthy
    """
    checks = response.get("checks", response.get("dependencies", {}))
    
    assert dependency_name in checks, (
        f"Dependency '{dependency_name}' not found in health checks. "
        f"Available: {list(checks.keys())}"
    )
    
    dep_check = checks[dependency_name]
    
    if isinstance(dep_check, dict):
        status = dep_check.get("status", "").lower()
        healthy_statuses = {"healthy", "ok", "up", "pass"}
        assert status in healthy_statuses, (
            f"Dependency '{dependency_name}' is unhealthy: {dep_check}"
        )
    elif isinstance(dep_check, bool):
        assert dep_check is True, (
            f"Dependency '{dependency_name}' is unhealthy"
        )


def assert_dependency_unhealthy(
    response: dict,
    dependency_name: str,
    expected_error: Optional[str] = None,
):
    """
    Assert that a specific dependency is unhealthy.
    
    Args:
        response: Health check response containing dependency checks
        dependency_name: Name of the dependency to check
        expected_error: Optional expected error message
    
    Raises:
        AssertionError: If dependency is not unhealthy as expected
    """
    checks = response.get("checks", response.get("dependencies", {}))
    
    assert dependency_name in checks, (
        f"Dependency '{dependency_name}' not found in health checks"
    )
    
    dep_check = checks[dependency_name]
    
    if isinstance(dep_check, dict):
        status = dep_check.get("status", "").lower()
        unhealthy_statuses = {"unhealthy", "error", "down", "fail"}
        assert status in unhealthy_statuses, (
            f"Expected '{dependency_name}' to be unhealthy, got: {dep_check}"
        )
        
        if expected_error:
            error = dep_check.get("error", dep_check.get("message", ""))
            assert expected_error.lower() in error.lower(), (
                f"Expected error containing '{expected_error}', got: {error}"
            )


def assert_all_dependencies_healthy(
    response: dict,
):
    """
    Assert that all dependencies in a health check are healthy.
    
    Args:
        response: Health check response containing dependency checks
    
    Raises:
        AssertionError: If any dependency is unhealthy
    """
    checks = response.get("checks", response.get("dependencies", {}))
    
    unhealthy = []
    for name, check in checks.items():
        if isinstance(check, dict):
            status = check.get("status", "").lower()
            if status not in {"healthy", "ok", "up", "pass"}:
                unhealthy.append(f"{name}: {check}")
        elif check is False:
            unhealthy.append(name)
    
    assert not unhealthy, (
        f"Unhealthy dependencies found: {unhealthy}"
    )


def assert_health_response_format(
    response: dict,
    required_fields: Optional[list[str]] = None,
):
    """
    Assert health response follows expected format.
    
    Args:
        response: Health check response dictionary
        required_fields: List of required fields (defaults to ["status"])
    
    Raises:
        AssertionError: If response format is invalid
    """
    required_fields = required_fields or ["status"]
    
    for field in required_fields:
        assert field in response, (
            f"Health response missing required field: '{field}'"
        )
    
    # Validate status field value
    if "status" in response:
        valid_statuses = {
            "healthy", "unhealthy", "degraded",
            "ok", "error", "up", "down",
            "pass", "fail", "warn"
        }
        status = response["status"].lower()
        assert status in valid_statuses, (
            f"Invalid health status: '{response['status']}'"
        )


def assert_readiness_includes_checks(
    response: dict,
    expected_checks: list[str],
):
    """
    Assert readiness response includes all expected checks.
    
    Args:
        response: Readiness check response
        expected_checks: List of expected check names
    
    Raises:
        AssertionError: If any expected check is missing
    """
    checks = response.get("checks", response.get("dependencies", {}))
    
    missing = set(expected_checks) - set(checks.keys())
    
    assert not missing, (
        f"Readiness response missing checks: {missing}"
    )


def assert_startup_complete(
    response: dict,
):
    """
    Assert startup probe indicates initialization is complete.
    
    Args:
        response: Startup probe response
    
    Raises:
        AssertionError: If startup is not complete
    """
    status = response.get("status", "").lower()
    initialized = response.get("initialized", False)
    
    assert status in {"healthy", "ok", "ready"} or initialized is True, (
        f"Startup not complete: {response}"
    )
