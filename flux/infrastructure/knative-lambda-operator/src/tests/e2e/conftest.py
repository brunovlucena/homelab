"""
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🧪 E2E Test Configuration and Fixtures

Purpose: Provide shared fixtures and configuration for E2E tests
User Story: QA-001 - E2E CloudEvents Testing

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
"""

import os
import pytest
import time
from kubernetes import client, config
from typing import Dict, Any

@pytest.fixture(scope="session")
def test_environment() -> Dict[str, str]:
    """Get test environment configuration"""
    env = os.getenv("ENV", "dev")
    return {
        'env': env,
        'namespace': f"knative-lambda-{env}",
        'rabbitmq_namespace': f"rabbitmq-{env}",
        'registry': "339954290315.dkr.ecr.us-west-2.amazonaws.com",
    }

@pytest.fixture(scope="session")
def kubernetes_clients():
    """Provide Kubernetes clients for tests"""
    try:
        config.load_kube_config()
    except Exception:
        config.load_incluster_config()
    
    return {
        'core': client.CoreV1Api(),
        'batch': client.BatchV1Api(),
        'apps': client.AppsV1Api(),
        'custom': client.CustomObjectsApi(),
    }

@pytest.fixture(scope="session")
def broker_url(test_environment) -> str:
    """Get broker URL - assumes port-forward is active"""
    # Check if running in cluster
    if os.path.exists('/var/run/secrets/kubernetes.io'):
        env = test_environment['env']
        return f"http://knative-lambda-broker-{env}-broker-ingress.{test_environment['namespace']}.svc.cluster.local"
    else:
        return "http://0.0.0.0:8081"

@pytest.fixture(scope="session")
def ecr_client():
    """Provide AWS ECR client"""
    import boto3
    return boto3.client('ecr', region_name='us-west-2')

@pytest.fixture(autouse=True)
def test_marker(request):
    """Print test marker before each test"""
    print(f"\n{'='*80}")
    print(f"🧪 Running: {request.node.nodeid}")
    print(f"{'='*80}\n")
    yield
    print(f"\n{'='*80}")
    print(f"✅ Completed: {request.node.nodeid}")
    print(f"{'='*80}\n")

def pytest_configure(config):
    """Register custom markers"""
    config.addinivalue_line("markers", "e2e: End-to-end integration tests")
    config.addinivalue_line("markers", "build: Build event tests")
    config.addinivalue_line("markers", "parser: Parser event tests")
    config.addinivalue_line("markers", "delete: Service deletion tests")
    config.addinivalue_line("markers", "lifecycle: Full lifecycle tests")
    config.addinivalue_line("markers", "slow: Slow tests (> 5 minutes)")

def pytest_collection_modifyitems(config, items):
    """Automatically mark e2e tests as slow"""
    for item in items:
        if "e2e" in item.keywords:
            item.add_marker(pytest.mark.slow)

