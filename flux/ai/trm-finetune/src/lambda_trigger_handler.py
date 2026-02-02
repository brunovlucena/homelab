#!/usr/bin/env python3
"""
🔧 Lambda Function CloudEvent Handler Template

Template for Lambda functions that receive CloudEvents from agent-sre.
Each Lambda function implements this handler to execute remediation.
"""

from typing import Dict, Any
import json
from cloudevents.http import from_http


def handle_cloudevent(headers: Dict[str, str], body: bytes) -> Dict[str, Any]:
    """
    Handle CloudEvent from agent-sre.
    
    Event Type: io.homelab.agent-sre.lambda.trigger
    
    Args:
        headers: HTTP headers (including CloudEvent headers)
        body: Request body (CloudEvent data)
        
    Returns:
        Dict with execution result
    """
    # Parse CloudEvent
    event = from_http(headers, body)
    event_data = event.get("data", {})
    
    # Extract Lambda function info
    lambda_function = event_data.get("lambda_function")
    parameters = event_data.get("parameters", {})
    alert = event_data.get("alert", {})
    correlation_id = event_data.get("correlation_id")
    
    print(f"🔧 Executing Lambda function: {lambda_function}")
    print(f"   Parameters: {json.dumps(parameters, indent=2)}")
    print(f"   Correlation ID: {correlation_id}")
    
    # Route to appropriate handler
    if lambda_function == "flux-reconcile-kustomization":
        return handle_flux_reconcile_kustomization(parameters, alert)
    elif lambda_function == "flux-reconcile-gitrepository":
        return handle_flux_reconcile_gitrepository(parameters, alert)
    elif lambda_function == "flux-reconcile-helmrelease":
        return handle_flux_reconcile_helmrelease(parameters, alert)
    elif lambda_function == "pod-restart":
        return handle_pod_restart(parameters, alert)
    elif lambda_function == "pod-check-status":
        return handle_pod_check_status(parameters, alert)
    elif lambda_function == "scale-deployment":
        return handle_scale_deployment(parameters, alert)
    elif lambda_function == "check-pvc-status":
        return handle_check_pvc_status(parameters, alert)
    else:
        return {
            "success": False,
            "error": f"Unknown Lambda function: {lambda_function}"
        }


def handle_flux_reconcile_kustomization(parameters: Dict[str, Any], alert: Dict[str, Any]) -> Dict[str, Any]:
    """Reconcile Flux Kustomization."""
    from kubernetes import client, config
    
    try:
        config.load_incluster_config()
        api = client.CustomObjectsApi()
        
        name = parameters["name"]
        namespace = parameters.get("namespace", "flux-system")
        
        # Trigger reconciliation via annotation
        api.patch_namespaced_custom_object(
            group="kustomize.toolkit.fluxcd.io",
            version="v1",
            namespace=namespace,
            plural="kustomizations",
            name=name,
            body={
                "metadata": {
                    "annotations": {
                        "reconcile.fluxcd.io/requestedAt": str(int(time.time()))
                    }
                }
            }
        )
        
        return {
            "success": True,
            "action": "flux-reconcile-kustomization",
            "name": name,
            "namespace": namespace
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


def handle_flux_reconcile_gitrepository(parameters: Dict[str, Any], alert: Dict[str, Any]) -> Dict[str, Any]:
    """Reconcile Flux GitRepository."""
    # Similar to kustomization
    pass


def handle_flux_reconcile_helmrelease(parameters: Dict[str, Any], alert: Dict[str, Any]) -> Dict[str, Any]:
    """Reconcile Flux HelmRelease."""
    # Similar to kustomization
    pass


def handle_pod_restart(parameters: Dict[str, Any], alert: Dict[str, Any]) -> Dict[str, Any]:
    """Restart a pod or deployment."""
    from kubernetes import client, config
    
    try:
        config.load_incluster_config()
        core_api = client.CoreV1Api()
        apps_api = client.AppsV1Api()
        
        name = parameters["name"]
        namespace = parameters["namespace"]
        resource_type = parameters.get("type", "pod")
        
        if resource_type == "pod":
            # Delete pod (will be recreated)
            core_api.delete_namespaced_pod(name=name, namespace=namespace)
        elif resource_type == "deployment":
            # Restart deployment
            apps_api.patch_namespaced_deployment(
                name=name,
                namespace=namespace,
                body={
                    "spec": {
                        "template": {
                            "metadata": {
                                "annotations": {
                                    "kubectl.kubernetes.io/restartedAt": str(int(time.time()))
                                }
                            }
                        }
                    }
                }
            )
        
        return {
            "success": True,
            "action": "pod-restart",
            "name": name,
            "namespace": namespace,
            "type": resource_type
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


def handle_pod_check_status(parameters: Dict[str, Any], alert: Dict[str, Any]) -> Dict[str, Any]:
    """Check pod status."""
    from kubernetes import client, config
    
    try:
        config.load_incluster_config()
        core_api = client.CoreV1Api()
        
        namespace = parameters["namespace"]
        selector = parameters.get("selector")
        name = parameters.get("name")
        
        if selector:
            pods = core_api.list_namespaced_pod(
                namespace=namespace,
                label_selector=selector
            )
        elif name:
            pod = core_api.read_namespaced_pod(name=name, namespace=namespace)
            pods = [pod]
        else:
            return {"success": False, "error": "Need either 'name' or 'selector'"}
        
        statuses = []
        for pod in pods.items:
            statuses.append({
                "name": pod.metadata.name,
                "phase": pod.status.phase,
                "ready": pod.status.conditions[0].status if pod.status.conditions else "Unknown"
            })
        
        return {
            "success": True,
            "action": "pod-check-status",
            "pods": statuses
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


def handle_scale_deployment(parameters: Dict[str, Any], alert: Dict[str, Any]) -> Dict[str, Any]:
    """Scale deployment."""
    from kubernetes import client, config
    
    try:
        config.load_incluster_config()
        apps_api = client.AppsV1Api()
        
        name = parameters["name"]
        namespace = parameters["namespace"]
        replicas = parameters["replicas"]
        
        apps_api.patch_namespaced_deployment_scale(
            name=name,
            namespace=namespace,
            body={"spec": {"replicas": replicas}}
        )
        
        return {
            "success": True,
            "action": "scale-deployment",
            "name": name,
            "namespace": namespace,
            "replicas": replicas
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


def handle_check_pvc_status(parameters: Dict[str, Any], alert: Dict[str, Any]) -> Dict[str, Any]:
    """Check PVC status."""
    from kubernetes import client, config
    
    try:
        config.load_incluster_config()
        core_api = client.CoreV1Api()
        
        name = parameters["name"]
        namespace = parameters["namespace"]
        
        pvc = core_api.read_namespaced_persistent_volume_claim(
            name=name,
            namespace=namespace
        )
        
        return {
            "success": True,
            "action": "check-pvc-status",
            "name": name,
            "namespace": namespace,
            "status": pvc.status.phase,
            "capacity": pvc.status.capacity.get("storage", "Unknown") if pvc.status.capacity else "Unknown"
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }

