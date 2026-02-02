"""
Kubernetes assertions for testing.

Provides semantic assertions for validating Kubernetes operations
in agent tests.
"""

from typing import Any, Optional
from unittest.mock import AsyncMock


def assert_k8s_resource_created(
    mock_client: Any,
    resource_kind: str,
    resource_name: str,
    namespace: Optional[str] = None,
):
    """
    Assert that a Kubernetes resource was created.
    
    Args:
        mock_client: Mock Kubernetes client
        resource_kind: Kind of resource (e.g., "Pod", "Deployment")
        resource_name: Name of the resource
        namespace: Optional namespace
    
    Raises:
        AssertionError: If resource was not created
    """
    # Check if apply_manifest or create was called
    if hasattr(mock_client, "apply_manifest"):
        apply_calls = mock_client.apply_manifest.call_args_list
        found = False
        
        for call in apply_calls:
            args, kwargs = call
            manifest = args[0] if args else kwargs.get("manifest", {})
            
            if isinstance(manifest, dict):
                if (manifest.get("kind") == resource_kind and
                    manifest.get("metadata", {}).get("name") == resource_name):
                    if namespace is None or manifest.get("metadata", {}).get("namespace") == namespace:
                        found = True
                        break
            elif isinstance(manifest, str):
                # YAML string - check if it contains the resource
                if resource_kind in manifest and resource_name in manifest:
                    found = True
                    break
        
        assert found, (
            f"Resource {resource_kind}/{resource_name} was not created. "
            f"apply_manifest calls: {apply_calls}"
        )
    else:
        raise AssertionError(
            "Mock client does not have apply_manifest method"
        )


def assert_k8s_resource_deleted(
    mock_client: Any,
    resource_kind: str,
    resource_name: str,
    namespace: Optional[str] = None,
):
    """
    Assert that a Kubernetes resource was deleted.
    
    Args:
        mock_client: Mock Kubernetes client
        resource_kind: Kind of resource
        resource_name: Name of the resource
        namespace: Optional namespace
    
    Raises:
        AssertionError: If resource was not deleted
    """
    if hasattr(mock_client, "delete_resource"):
        delete_calls = mock_client.delete_resource.call_args_list
        found = False
        
        for call in delete_calls:
            args, kwargs = call
            
            # Check positional args
            if len(args) >= 2:
                if args[0] == resource_kind and args[1] == resource_name:
                    if namespace is None or (len(args) > 2 and args[2] == namespace):
                        found = True
                        break
            
            # Check keyword args
            if (kwargs.get("kind") == resource_kind and
                kwargs.get("name") == resource_name):
                if namespace is None or kwargs.get("namespace") == namespace:
                    found = True
                    break
        
        assert found, (
            f"Resource {resource_kind}/{resource_name} was not deleted. "
            f"delete_resource calls: {delete_calls}"
        )
    else:
        raise AssertionError(
            "Mock client does not have delete_resource method"
        )


def assert_k8s_resource_exists(
    mock_client: Any,
    resource_kind: str,
    resource_name: str,
    namespace: Optional[str] = None,
):
    """
    Assert that a Kubernetes resource exists (was queried).
    
    Args:
        mock_client: Mock Kubernetes client
        resource_kind: Kind of resource
        resource_name: Name of the resource
        namespace: Optional namespace
    
    Raises:
        AssertionError: If resource was not queried
    """
    if hasattr(mock_client, "get_resource"):
        get_calls = mock_client.get_resource.call_args_list
        found = False
        
        for call in get_calls:
            args, kwargs = call
            
            if len(args) >= 2:
                if args[0] == resource_kind and args[1] == resource_name:
                    found = True
                    break
            
            if (kwargs.get("kind") == resource_kind and
                kwargs.get("name") == resource_name):
                found = True
                break
        
        assert found, (
            f"Resource {resource_kind}/{resource_name} was not queried. "
            f"get_resource calls: {get_calls}"
        )


def assert_k8s_resource_patched(
    mock_client: Any,
    resource_kind: str,
    resource_name: str,
    expected_patch: Optional[dict] = None,
):
    """
    Assert that a Kubernetes resource was patched.
    
    Args:
        mock_client: Mock Kubernetes client
        resource_kind: Kind of resource
        resource_name: Name of the resource
        expected_patch: Optional expected patch content
    
    Raises:
        AssertionError: If resource was not patched
    """
    if hasattr(mock_client, "patch_resource"):
        patch_calls = mock_client.patch_resource.call_args_list
        found = False
        
        for call in patch_calls:
            args, kwargs = call
            
            kind_match = (
                (len(args) >= 1 and args[0] == resource_kind) or
                kwargs.get("kind") == resource_kind
            )
            name_match = (
                (len(args) >= 2 and args[1] == resource_name) or
                kwargs.get("name") == resource_name
            )
            
            if kind_match and name_match:
                if expected_patch is None:
                    found = True
                    break
                
                # Check if patch content matches
                patch = kwargs.get("patch") or (args[2] if len(args) > 2 else None)
                if patch and all(
                    patch.get(k) == v for k, v in expected_patch.items()
                ):
                    found = True
                    break
        
        assert found, (
            f"Resource {resource_kind}/{resource_name} was not patched "
            f"with expected content. patch_resource calls: {patch_calls}"
        )


def assert_k8s_namespace_created(
    mock_client: Any,
    namespace_name: str,
):
    """
    Assert that a Kubernetes namespace was created.
    
    Args:
        mock_client: Mock Kubernetes client
        namespace_name: Name of the namespace
    
    Raises:
        AssertionError: If namespace was not created
    """
    if hasattr(mock_client, "create_namespace"):
        mock_client.create_namespace.assert_called()
        
        calls = mock_client.create_namespace.call_args_list
        found = any(
            namespace_name in str(call) for call in calls
        )
        
        assert found, (
            f"Namespace '{namespace_name}' was not created. "
            f"create_namespace calls: {calls}"
        )
    else:
        # Check if applied as manifest
        assert_k8s_resource_created(
            mock_client, "Namespace", namespace_name
        )


def assert_k8s_logs_retrieved(
    mock_client: Any,
    pod_name: str,
    container: Optional[str] = None,
):
    """
    Assert that pod logs were retrieved.
    
    Args:
        mock_client: Mock Kubernetes client
        pod_name: Name of the pod
        container: Optional container name
    
    Raises:
        AssertionError: If logs were not retrieved
    """
    if hasattr(mock_client, "get_logs"):
        get_logs_calls = mock_client.get_logs.call_args_list
        found = False
        
        for call in get_logs_calls:
            args, kwargs = call
            
            if pod_name in str(args) or pod_name in str(kwargs):
                if container is None or container in str(args) or container in str(kwargs):
                    found = True
                    break
        
        assert found, (
            f"Logs for pod '{pod_name}' were not retrieved. "
            f"get_logs calls: {get_logs_calls}"
        )


def assert_k8s_no_mutations(
    mock_client: Any,
):
    """
    Assert that no mutation operations were performed.
    
    Args:
        mock_client: Mock Kubernetes client
    
    Raises:
        AssertionError: If any mutations occurred
    """
    mutation_methods = [
        "apply_manifest",
        "delete_resource",
        "patch_resource",
        "create_namespace",
        "delete_namespace",
    ]
    
    for method_name in mutation_methods:
        if hasattr(mock_client, method_name):
            method = getattr(mock_client, method_name)
            if hasattr(method, "call_count") and method.call_count > 0:
                raise AssertionError(
                    f"Unexpected K8s mutation: {method_name} "
                    f"called {method.call_count} times"
                )


def assert_k8s_resource_labels(
    resource: dict,
    expected_labels: dict,
):
    """
    Assert that a K8s resource has expected labels.
    
    Args:
        resource: Kubernetes resource dictionary
        expected_labels: Expected label key-value pairs
    
    Raises:
        AssertionError: If labels don't match
    """
    actual_labels = resource.get("metadata", {}).get("labels", {})
    
    for key, value in expected_labels.items():
        assert key in actual_labels, (
            f"Resource missing label: '{key}'"
        )
        assert actual_labels[key] == value, (
            f"Label '{key}' value mismatch: "
            f"expected '{value}', got '{actual_labels[key]}'"
        )


def assert_k8s_resource_annotations(
    resource: dict,
    expected_annotations: dict,
):
    """
    Assert that a K8s resource has expected annotations.
    
    Args:
        resource: Kubernetes resource dictionary
        expected_annotations: Expected annotation key-value pairs
    
    Raises:
        AssertionError: If annotations don't match
    """
    actual_annotations = resource.get("metadata", {}).get("annotations", {})
    
    for key, value in expected_annotations.items():
        assert key in actual_annotations, (
            f"Resource missing annotation: '{key}'"
        )
        assert actual_annotations[key] == value, (
            f"Annotation '{key}' value mismatch: "
            f"expected '{value}', got '{actual_annotations[key]}'"
        )
