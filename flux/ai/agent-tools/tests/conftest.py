"""Pytest configuration and fixtures for agent-tools tests."""
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
def mock_dynamic_client():
    """Mock Kubernetes dynamic client."""
    mock_client = MagicMock()
    mock_resource = MagicMock()
    
    # Mock get() to return items
    mock_result = MagicMock()
    mock_result.items = []
    mock_result.to_dict.return_value = {"metadata": {"name": "test"}}
    mock_resource.get.return_value = mock_result
    mock_resource.create.return_value = mock_result
    mock_resource.replace.return_value = mock_result
    mock_resource.patch.return_value = mock_result
    mock_resource.delete.return_value = None
    
    mock_client.resources.get.return_value = mock_resource
    
    return mock_client


@pytest.fixture
def sample_pod():
    """Sample pod manifest."""
    return {
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {
            "name": "test-pod",
            "namespace": "default",
        },
        "spec": {
            "containers": [
                {
                    "name": "test",
                    "image": "nginx:latest",
                }
            ]
        }
    }


@pytest.fixture
def sample_deployment():
    """Sample deployment manifest."""
    return {
        "apiVersion": "apps/v1",
        "kind": "Deployment",
        "metadata": {
            "name": "test-deployment",
            "namespace": "default",
        },
        "spec": {
            "replicas": 1,
            "selector": {
                "matchLabels": {"app": "test"}
            },
            "template": {
                "metadata": {"labels": {"app": "test"}},
                "spec": {
                    "containers": [
                        {"name": "test", "image": "nginx:latest"}
                    ]
                }
            }
        }
    }


@pytest.fixture
def sample_cloudevent():
    """Sample CloudEvent for testing."""
    return {
        "type": "io.homelab.k8s.pods.list",
        "source": "test",
        "data": {
            "namespace": "default",
        }
    }
