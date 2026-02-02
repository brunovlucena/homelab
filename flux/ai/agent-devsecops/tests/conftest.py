"""Pytest configuration and fixtures for agent-devsecops tests."""
import pytest
from unittest.mock import MagicMock, patch


@pytest.fixture
def mock_k8s_config():
    """Mock Kubernetes config loading."""
    with patch('kubernetes.config.load_incluster_config') as mock_in, \
         patch('kubernetes.config.load_kube_config') as mock_out:
        mock_in.side_effect = Exception("Not in cluster")
        yield mock_out


@pytest.fixture
def sample_lambdafunction():
    """Sample LambdaFunction resource."""
    return {
        "apiVersion": "lambda.knative.io/v1alpha1",
        "kind": "LambdaFunction",
        "metadata": {
            "name": "test-function",
            "namespace": "default",
        },
        "spec": {
            "image": "ghcr.io/brunovlucena/test-function:v1.0.0",
            "runtime": "python3.12",
        },
        "status": {
            "ready": True,
        }
    }


@pytest.fixture
def sample_image_info():
    """Sample parsed image info."""
    return {
        "registry": "ghcr.io",
        "owner": "brunovlucena",
        "name": "test-function",
        "tag": "v1.0.0",
        "version": "1.0.0",
    }


@pytest.fixture
def sample_cloudevent():
    """Sample CloudEvent for testing."""
    return {
        "type": "io.homelab.scan.lambdafunctions",
        "source": "test",
        "data": {
            "namespace": "default",
        }
    }
