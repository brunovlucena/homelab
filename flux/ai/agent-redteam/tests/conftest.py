"""
Pytest configuration and fixtures for agent-redteam tests.
"""
import os
import sys
import pytest
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock

# Add src to path for imports
src_path = Path(__file__).parent.parent / "src"
sys.path.insert(0, str(src_path))

# Set test environment variables
os.environ.setdefault("DRY_RUN", "true")
os.environ.setdefault("TARGET_NAMESPACE", "test-namespace")
os.environ.setdefault("EXPLOITS_PATH", "/tmp/exploits")


@pytest.fixture
def mock_k8s_client():
    """Mock Kubernetes client for testing."""
    from exploit_runner.handler import KubernetesClient
    
    client = MagicMock(spec=KubernetesClient)
    client.apply_manifest = AsyncMock(return_value=(True, "resource created"))
    client.delete_resource = AsyncMock(return_value=(True, "resource deleted"))
    client.get_resource = AsyncMock(return_value=(True, {"status": "success"}))
    client.get_logs = AsyncMock(return_value=(True, "log output"))
    client.context = "test-context"
    client.timeout = 60
    
    return client


@pytest.fixture
def exploit_runner(mock_k8s_client):
    """Create ExploitRunner with mocked K8s client."""
    from exploit_runner.handler import ExploitRunner
    
    runner = ExploitRunner(k8s_client=mock_k8s_client)
    runner.dry_run = True  # Always dry-run in tests
    
    return runner


@pytest.fixture
def sample_exploit_result():
    """Create a sample exploit result for testing."""
    from shared.types import ExploitResult, ExploitStatus
    
    return ExploitResult(
        exploit_id="vuln-001",
        status=ExploitStatus.SUCCESS,
        started_at="2024-01-01T00:00:00Z",
        completed_at="2024-01-01T00:00:10Z",
        duration_seconds=10.0,
        output="Exploit executed successfully",
    )


@pytest.fixture
def sample_test_run():
    """Create a sample test run for testing."""
    from shared.types import TestRun, ExploitResult, ExploitStatus
    
    return TestRun(
        id="test-123",
        name="test-run",
        target_cluster="test-cluster",
        target_namespace="test-namespace",
        results=[
            ExploitResult(
                exploit_id="vuln-001",
                status=ExploitStatus.SUCCESS,
                duration_seconds=5.0,
            ),
            ExploitResult(
                exploit_id="vuln-002",
                status=ExploitStatus.BLOCKED,
                duration_seconds=3.0,
                mitigated_by="admission_webhook",
            ),
        ]
    )
