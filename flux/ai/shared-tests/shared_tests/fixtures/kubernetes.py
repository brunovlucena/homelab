"""
Kubernetes fixtures for testing agent interactions with K8s APIs.

Provides mocked Kubernetes clients and resource factories for
testing agents that interact with Kubernetes.
"""

import pytest
from unittest.mock import AsyncMock, MagicMock
from dataclasses import dataclass, field
from typing import Any, Optional
from datetime import datetime, timezone


@dataclass
class K8sResource:
    """Base Kubernetes resource for testing."""
    
    api_version: str = "v1"
    kind: str = "Resource"
    name: str = ""
    namespace: str = "default"
    labels: dict = field(default_factory=dict)
    annotations: dict = field(default_factory=dict)
    uid: str = ""
    created_at: str = field(
        default_factory=lambda: datetime.now(timezone.utc).isoformat()
    )
    
    def to_dict(self) -> dict:
        """Convert to Kubernetes resource format."""
        return {
            "apiVersion": self.api_version,
            "kind": self.kind,
            "metadata": {
                "name": self.name,
                "namespace": self.namespace,
                "labels": self.labels,
                "annotations": self.annotations,
                "uid": self.uid or f"{self.name}-uid",
                "creationTimestamp": self.created_at,
            },
        }


@dataclass
class K8sPod(K8sResource):
    """Kubernetes Pod resource for testing."""
    
    kind: str = "Pod"
    api_version: str = "v1"
    phase: str = "Running"
    container_image: str = "test-image:latest"
    container_name: str = "main"
    restart_count: int = 0
    
    def to_dict(self) -> dict:
        base = super().to_dict()
        base["spec"] = {
            "containers": [
                {
                    "name": self.container_name,
                    "image": self.container_image,
                }
            ],
        }
        base["status"] = {
            "phase": self.phase,
            "containerStatuses": [
                {
                    "name": self.container_name,
                    "ready": self.phase == "Running",
                    "restartCount": self.restart_count,
                    "state": {
                        "running" if self.phase == "Running" else "waiting": {}
                    },
                }
            ],
        }
        return base


@dataclass
class K8sDeployment(K8sResource):
    """Kubernetes Deployment resource for testing."""
    
    kind: str = "Deployment"
    api_version: str = "apps/v1"
    replicas: int = 1
    available_replicas: int = 1
    container_image: str = "test-image:latest"
    
    def to_dict(self) -> dict:
        base = super().to_dict()
        base["spec"] = {
            "replicas": self.replicas,
            "selector": {
                "matchLabels": {"app": self.name},
            },
            "template": {
                "metadata": {
                    "labels": {"app": self.name},
                },
                "spec": {
                    "containers": [
                        {
                            "name": self.name,
                            "image": self.container_image,
                        }
                    ],
                },
            },
        }
        base["status"] = {
            "replicas": self.replicas,
            "availableReplicas": self.available_replicas,
            "readyReplicas": self.available_replicas,
            "updatedReplicas": self.replicas,
        }
        return base


@dataclass
class K8sNamespace(K8sResource):
    """Kubernetes Namespace resource for testing."""
    
    kind: str = "Namespace"
    api_version: str = "v1"
    phase: str = "Active"
    
    def to_dict(self) -> dict:
        base = super().to_dict()
        # Namespace doesn't have a namespace in metadata
        del base["metadata"]["namespace"]
        base["status"] = {
            "phase": self.phase,
        }
        return base


@dataclass
class K8sService(K8sResource):
    """Kubernetes Service resource for testing."""
    
    kind: str = "Service"
    api_version: str = "v1"
    port: int = 80
    target_port: int = 8080
    service_type: str = "ClusterIP"
    cluster_ip: str = "10.96.0.1"
    
    def to_dict(self) -> dict:
        base = super().to_dict()
        base["spec"] = {
            "type": self.service_type,
            "clusterIP": self.cluster_ip,
            "ports": [
                {
                    "port": self.port,
                    "targetPort": self.target_port,
                    "protocol": "TCP",
                }
            ],
            "selector": {"app": self.name},
        }
        return base


@dataclass
class K8sConfigMap(K8sResource):
    """Kubernetes ConfigMap resource for testing."""
    
    kind: str = "ConfigMap"
    api_version: str = "v1"
    data: dict = field(default_factory=dict)
    
    def to_dict(self) -> dict:
        base = super().to_dict()
        base["data"] = self.data
        return base


@dataclass
class K8sSecret(K8sResource):
    """Kubernetes Secret resource for testing."""
    
    kind: str = "Secret"
    api_version: str = "v1"
    secret_type: str = "Opaque"
    string_data: dict = field(default_factory=dict)
    
    def to_dict(self) -> dict:
        base = super().to_dict()
        base["type"] = self.secret_type
        base["stringData"] = self.string_data
        return base


class K8sResourceFactory:
    """Factory for creating Kubernetes resources."""
    
    def __init__(self, namespace: str = "test-namespace"):
        self.namespace = namespace
    
    def create_pod(
        self,
        name: str,
        image: str = "test-image:latest",
        phase: str = "Running",
        **kwargs,
    ) -> K8sPod:
        """Create a Pod resource."""
        return K8sPod(
            name=name,
            namespace=kwargs.get("namespace", self.namespace),
            container_image=image,
            phase=phase,
            labels=kwargs.get("labels", {"app": name}),
            annotations=kwargs.get("annotations", {}),
        )
    
    def create_deployment(
        self,
        name: str,
        replicas: int = 1,
        image: str = "test-image:latest",
        **kwargs,
    ) -> K8sDeployment:
        """Create a Deployment resource."""
        return K8sDeployment(
            name=name,
            namespace=kwargs.get("namespace", self.namespace),
            replicas=replicas,
            available_replicas=kwargs.get("available_replicas", replicas),
            container_image=image,
            labels=kwargs.get("labels", {"app": name}),
        )
    
    def create_namespace(
        self,
        name: str,
        phase: str = "Active",
        **kwargs,
    ) -> K8sNamespace:
        """Create a Namespace resource."""
        return K8sNamespace(
            name=name,
            phase=phase,
            labels=kwargs.get("labels", {}),
        )
    
    def create_service(
        self,
        name: str,
        port: int = 80,
        target_port: int = 8080,
        **kwargs,
    ) -> K8sService:
        """Create a Service resource."""
        return K8sService(
            name=name,
            namespace=kwargs.get("namespace", self.namespace),
            port=port,
            target_port=target_port,
            labels=kwargs.get("labels", {"app": name}),
        )
    
    def create_configmap(
        self,
        name: str,
        data: Optional[dict] = None,
        **kwargs,
    ) -> K8sConfigMap:
        """Create a ConfigMap resource."""
        return K8sConfigMap(
            name=name,
            namespace=kwargs.get("namespace", self.namespace),
            data=data or {},
            labels=kwargs.get("labels", {}),
        )
    
    def create_secret(
        self,
        name: str,
        string_data: Optional[dict] = None,
        **kwargs,
    ) -> K8sSecret:
        """Create a Secret resource."""
        return K8sSecret(
            name=name,
            namespace=kwargs.get("namespace", self.namespace),
            string_data=string_data or {},
            labels=kwargs.get("labels", {}),
        )


class MockKubernetesClient:
    """Mock Kubernetes client for testing."""
    
    def __init__(self, namespace: str = "test-namespace"):
        self.namespace = namespace
        self.context = "test-context"
        self.timeout = 60
        self._resources: dict[str, list] = {}
        
        # Async mock methods
        self.apply_manifest = AsyncMock(return_value=(True, "resource created"))
        self.delete_resource = AsyncMock(return_value=(True, "resource deleted"))
        self.get_resource = AsyncMock(return_value=(True, {"status": "success"}))
        self.get_logs = AsyncMock(return_value=(True, "log output"))
        self.list_resources = AsyncMock(return_value=(True, []))
        self.patch_resource = AsyncMock(return_value=(True, "resource patched"))
        self.create_namespace = AsyncMock(return_value=(True, "namespace created"))
        self.delete_namespace = AsyncMock(return_value=(True, "namespace deleted"))
        
        # Sync mock methods
        self.get_current_context = MagicMock(return_value=self.context)
        self.get_namespaces = MagicMock(return_value=["default", namespace])
    
    def add_resource(self, resource: K8sResource):
        """Add a resource to the mock client."""
        kind = resource.kind.lower()
        if kind not in self._resources:
            self._resources[kind] = []
        self._resources[kind].append(resource)
        
        # Update get_resource to return this resource
        def get_resource_side_effect(kind, name, namespace=None):
            for r in self._resources.get(kind.lower(), []):
                if r.name == name:
                    return (True, r.to_dict())
            return (False, "resource not found")
        
        self.get_resource.side_effect = get_resource_side_effect


@pytest.fixture
def mock_k8s_client():
    """Mock Kubernetes client fixture."""
    return MockKubernetesClient()


@pytest.fixture
def k8s_resource_factory():
    """Factory for creating K8s resources."""
    return K8sResourceFactory()


@pytest.fixture
def mock_k8s_pod(k8s_resource_factory):
    """Sample Pod resource."""
    return k8s_resource_factory.create_pod(
        name="test-pod",
        image="test-image:latest",
        phase="Running",
    )


@pytest.fixture
def mock_k8s_deployment(k8s_resource_factory):
    """Sample Deployment resource."""
    return k8s_resource_factory.create_deployment(
        name="test-deployment",
        replicas=3,
        image="test-image:latest",
    )


@pytest.fixture
def mock_k8s_namespace(k8s_resource_factory):
    """Sample Namespace resource."""
    return k8s_resource_factory.create_namespace(
        name="test-namespace",
        phase="Active",
    )


@pytest.fixture
def mock_k8s_service(k8s_resource_factory):
    """Sample Service resource."""
    return k8s_resource_factory.create_service(
        name="test-service",
        port=80,
        target_port=8080,
    )
