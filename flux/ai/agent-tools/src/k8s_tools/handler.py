"""
CloudEvent handler for K8s API operations.

Handles ALL Kubernetes API operations via CloudEvents.
"""

import json
import logging
import time
from datetime import datetime
from typing import Any, Dict, Optional, Tuple

from kubernetes import client, config
from kubernetes.client.rest import ApiException
from kubernetes.dynamic import DynamicClient

logger = logging.getLogger(__name__)

# API group mappings
API_GROUPS = {
    # Core (v1)
    "pods": ("", "v1", "Pod"),
    "services": ("", "v1", "Service"),
    "configmaps": ("", "v1", "ConfigMap"),
    "secrets": ("", "v1", "Secret"),
    "namespaces": ("", "v1", "Namespace"),
    "serviceaccounts": ("", "v1", "ServiceAccount"),
    "persistentvolumeclaims": ("", "v1", "PersistentVolumeClaim"),
    "pvcs": ("", "v1", "PersistentVolumeClaim"),
    "events": ("", "v1", "Event"),
    "endpoints": ("", "v1", "Endpoints"),
    # Apps
    "deployments": ("apps", "v1", "Deployment"),
    "statefulsets": ("apps", "v1", "StatefulSet"),
    "daemonsets": ("apps", "v1", "DaemonSet"),
    "replicasets": ("apps", "v1", "ReplicaSet"),
    # Batch
    "jobs": ("batch", "v1", "Job"),
    "cronjobs": ("batch", "v1", "CronJob"),
    # Autoscaling
    "hpa": ("autoscaling", "v2", "HorizontalPodAutoscaler"),
    "horizontalpodautoscalers": ("autoscaling", "v2", "HorizontalPodAutoscaler"),
    # Networking
    "ingresses": ("networking.k8s.io", "v1", "Ingress"),
    "networkpolicies": ("networking.k8s.io", "v1", "NetworkPolicy"),
    # Knative Serving
    "ksvc": ("serving.knative.dev", "v1", "Service"),
    "revisions": ("serving.knative.dev", "v1", "Revision"),
    "routes": ("serving.knative.dev", "v1", "Route"),
    "configurations": ("serving.knative.dev", "v1", "Configuration"),
    # Knative Eventing
    "brokers": ("eventing.knative.dev", "v1", "Broker"),
    "triggers": ("eventing.knative.dev", "v1", "Trigger"),
    "eventtypes": ("eventing.knative.dev", "v1beta2", "EventType"),
    # Lambda
    "lambdafunctions": ("lambda.knative.io", "v1alpha1", "LambdaFunction"),
    "functions": ("lambda.knative.io", "v1alpha1", "LambdaFunction"),
    "lambdaagents": ("lambda.knative.io", "v1alpha1", "LambdaAgent"),
    "agents": ("lambda.knative.io", "v1alpha1", "LambdaAgent"),
    # Flux
    "kustomizations": ("kustomize.toolkit.fluxcd.io", "v1", "Kustomization"),
    "helmreleases": ("helm.toolkit.fluxcd.io", "v2", "HelmRelease"),
    "gitrepositories": ("source.toolkit.fluxcd.io", "v1", "GitRepository"),
    "helmrepositories": ("source.toolkit.fluxcd.io", "v1", "HelmRepository"),
    "ocirepositories": ("source.toolkit.fluxcd.io", "v1", "OCIRepository"),
    # Cert-manager
    "certificates": ("cert-manager.io", "v1", "Certificate"),
    "certificaterequests": ("cert-manager.io", "v1", "CertificateRequest"),
    "issuers": ("cert-manager.io", "v1", "Issuer"),
    "clusterissuers": ("cert-manager.io", "v1", "ClusterIssuer"),
    # Monitoring
    "prometheusrules": ("monitoring.coreos.com", "v1", "PrometheusRule"),
    "servicemonitors": ("monitoring.coreos.com", "v1", "ServiceMonitor"),
    "podmonitors": ("monitoring.coreos.com", "v1", "PodMonitor"),
    "alertmanagerconfigs": ("monitoring.coreos.com", "v1alpha1", "AlertmanagerConfig"),
    # RabbitMQ
    "rabbitmqclusters": ("rabbitmq.com", "v1beta1", "RabbitmqCluster"),
    "clusters": ("rabbitmq.com", "v1beta1", "RabbitmqCluster"),
    "queues": ("rabbitmq.com", "v1beta1", "Queue"),
    "exchanges": ("rabbitmq.com", "v1beta1", "Exchange"),
    "bindings": ("rabbitmq.com", "v1beta1", "Binding"),
    "policies": ("rabbitmq.com", "v1beta1", "Policy"),
    "users": ("rabbitmq.com", "v1beta1", "User"),
    "vhosts": ("rabbitmq.com", "v1beta1", "Vhost"),
    "permissions": ("rabbitmq.com", "v1beta1", "Permission"),
    # External Secrets
    "externalsecrets": ("external-secrets.io", "v1beta1", "ExternalSecret"),
    "secretstores": ("external-secrets.io", "v1beta1", "SecretStore"),
    "clustersecretstores": ("external-secrets.io", "v1beta1", "ClusterSecretStore"),
    # Sealed Secrets
    "sealedsecrets": ("bitnami.com", "v1alpha1", "SealedSecret"),
    # K6
    "testruns": ("k6.io", "v1alpha1", "TestRun"),
    "privateloadzones": ("k6.io", "v1alpha1", "PrivateLoadZone"),
}


class K8sHandler:
    """Handler for K8s API operations."""

    def __init__(self):
        try:
            config.load_incluster_config()
        except config.ConfigException:
            config.load_kube_config()
        
        self.api_client = client.ApiClient()
        self.dynamic = DynamicClient(self.api_client)
        self.core_v1 = client.CoreV1Api()
        self.apps_v1 = client.AppsV1Api()

    def _get_resource(self, resource_type: str):
        """Get dynamic resource client."""
        if resource_type not in API_GROUPS:
            raise ValueError(f"Unknown resource type: {resource_type}")
        
        group, version, kind = API_GROUPS[resource_type]
        api_version = f"{group}/{version}" if group else version
        
        return self.dynamic.resources.get(api_version=api_version, kind=kind)

    def list(self, resource_type: str, namespace: Optional[str] = None,
             labels: Optional[Dict[str, str]] = None, limit: int = 100) -> Dict[str, Any]:
        """List resources."""
        resource = self._get_resource(resource_type)
        kwargs = {"limit": limit}
        
        if labels:
            kwargs["label_selector"] = ",".join(f"{k}={v}" for k, v in labels.items())
        
        if namespace:
            result = resource.get(namespace=namespace, **kwargs)
        else:
            result = resource.get(**kwargs)
        
        return {
            "items": [item.to_dict() for item in result.items],
            "count": len(result.items),
        }

    def get(self, resource_type: str, name: str, namespace: Optional[str] = None) -> Dict[str, Any]:
        """Get a specific resource."""
        resource = self._get_resource(resource_type)
        
        if namespace:
            result = resource.get(name=name, namespace=namespace)
        else:
            result = resource.get(name=name)
        
        return result.to_dict()

    def create(self, resource_type: str, manifest: Dict[str, Any],
               namespace: Optional[str] = None) -> Dict[str, Any]:
        """Create a resource."""
        resource = self._get_resource(resource_type)
        
        if namespace:
            result = resource.create(body=manifest, namespace=namespace)
        else:
            result = resource.create(body=manifest)
        
        return result.to_dict()

    def update(self, resource_type: str, name: str, manifest: Dict[str, Any],
               namespace: Optional[str] = None) -> Dict[str, Any]:
        """Update a resource."""
        resource = self._get_resource(resource_type)
        
        if namespace:
            result = resource.replace(body=manifest, name=name, namespace=namespace)
        else:
            result = resource.replace(body=manifest, name=name)
        
        return result.to_dict()

    def patch(self, resource_type: str, name: str, patch: Dict[str, Any],
              namespace: Optional[str] = None) -> Dict[str, Any]:
        """Patch a resource."""
        resource = self._get_resource(resource_type)
        
        if namespace:
            result = resource.patch(body=patch, name=name, namespace=namespace,
                                   content_type="application/strategic-merge-patch+json")
        else:
            result = resource.patch(body=patch, name=name,
                                   content_type="application/strategic-merge-patch+json")
        
        return result.to_dict()

    def delete(self, resource_type: str, name: str, namespace: Optional[str] = None) -> bool:
        """Delete a resource."""
        resource = self._get_resource(resource_type)
        
        if namespace:
            resource.delete(name=name, namespace=namespace)
        else:
            resource.delete(name=name)
        
        return True

    def scale(self, resource_type: str, name: str, replicas: int, namespace: str) -> Dict[str, Any]:
        """Scale a workload."""
        return self.patch(resource_type, name, {"spec": {"replicas": replicas}}, namespace)

    def restart(self, resource_type: str, name: str, namespace: str) -> Dict[str, Any]:
        """Restart a deployment."""
        patch = {
            "spec": {
                "template": {
                    "metadata": {
                        "annotations": {
                            "kubectl.kubernetes.io/restartedAt": datetime.utcnow().isoformat()
                        }
                    }
                }
            }
        }
        return self.patch(resource_type, name, patch, namespace)

    def logs(self, name: str, namespace: str, container: Optional[str] = None,
             tail_lines: int = 100) -> str:
        """Get pod logs."""
        kwargs = {"name": name, "namespace": namespace, "tail_lines": tail_lines}
        if container:
            kwargs["container"] = container
        return self.core_v1.read_namespaced_pod_log(**kwargs)

    def reconcile(self, resource_type: str, name: str, namespace: str) -> Dict[str, Any]:
        """Trigger Flux reconciliation."""
        patch = {
            "metadata": {
                "annotations": {
                    "reconcile.fluxcd.io/requestedAt": datetime.utcnow().isoformat()
                }
            }
        }
        return self.patch(resource_type, name, patch, namespace)

    def suspend(self, resource_type: str, name: str, namespace: str) -> Dict[str, Any]:
        """Suspend a Flux resource."""
        return self.patch(resource_type, name, {"spec": {"suspend": True}}, namespace)

    def resume(self, resource_type: str, name: str, namespace: str) -> Dict[str, Any]:
        """Resume a Flux resource."""
        return self.patch(resource_type, name, {"spec": {"suspend": False}}, namespace)


# Global handler instance
_handler: Optional[K8sHandler] = None


def get_handler() -> K8sHandler:
    """Get or create handler instance."""
    global _handler
    if _handler is None:
        _handler = K8sHandler()
    return _handler


def parse_event_type(event_type: str) -> Tuple[str, str, str]:
    """
    Parse event type to extract domain, resource, and operation.
    
    Examples:
        io.homelab.k8s.pods.list -> (k8s, pods, list)
        io.homelab.knative.services.create -> (knative, services, create)
        io.homelab.flux.kustomizations.reconcile -> (flux, kustomizations, reconcile)
    """
    parts = event_type.replace("io.homelab.", "").split(".")
    if len(parts) >= 3:
        return parts[0], parts[1], parts[2]
    elif len(parts) == 2:
        return parts[0], parts[1], "list"
    else:
        raise ValueError(f"Invalid event type: {event_type}")


def handle(event: Dict[str, Any]) -> Dict[str, Any]:
    """
    Handle CloudEvent for K8s operation.
    
    Args:
        event: CloudEvent dict with type, source, data
        
    Returns:
        Response dict with success, data/error
    """
    event_type = event.get("type", "")
    data = event.get("data", {})
    start_time = time.time()
    
    logger.info(f"Handling event: {event_type}")
    
    try:
        domain, resource, operation = parse_event_type(event_type)
        handler = get_handler()
        
        # Map domain prefixes to resource types
        resource_map = {
            "k8s": resource,
            "knative": resource if resource != "services" else "ksvc",
            "lambda": resource,
            "flux": resource,
            "certmanager": resource,
            "monitoring": resource,
            "rabbitmq": resource,
            "secrets": resource,
            "k6": resource,
        }
        
        resource_type = resource_map.get(domain, resource)
        namespace = data.get("namespace")
        name = data.get("name")
        manifest = data.get("manifest")
        labels = data.get("labels")
        
        result = None
        
        # Execute operation
        if operation == "list":
            result = handler.list(resource_type, namespace, labels, data.get("limit", 100))
        
        elif operation == "get":
            result = handler.get(resource_type, name, namespace)
        
        elif operation == "create":
            result = handler.create(resource_type, manifest, namespace)
        
        elif operation == "update":
            result = handler.update(resource_type, name, manifest, namespace)
        
        elif operation == "patch":
            result = handler.patch(resource_type, name, data.get("patch", {}), namespace)
        
        elif operation == "delete":
            handler.delete(resource_type, name, namespace)
            result = {"deleted": True, "name": name}
        
        elif operation == "scale":
            result = handler.scale(resource_type, name, data.get("replicas", 1), namespace)
        
        elif operation == "restart":
            result = handler.restart(resource_type, name, namespace)
        
        elif operation == "logs":
            logs = handler.logs(name, namespace, data.get("container"), data.get("tailLines", 100))
            result = {"logs": logs}
        
        elif operation == "reconcile":
            result = handler.reconcile(resource_type, name, namespace)
        
        elif operation == "suspend":
            result = handler.suspend(resource_type, name, namespace)
        
        elif operation == "resume":
            result = handler.resume(resource_type, name, namespace)
        
        elif operation == "trigger":
            # Trigger CronJob
            job_manifest = {
                "apiVersion": "batch/v1",
                "kind": "Job",
                "metadata": {
                    "name": f"{name}-manual-{int(time.time())}",
                    "namespace": namespace,
                },
                "spec": handler.get("cronjobs", name, namespace).get("spec", {}).get("jobTemplate", {}).get("spec", {}),
            }
            result = handler.create("jobs", job_manifest, namespace)
        
        elif operation == "refresh":
            # Refresh ExternalSecret
            patch = {"metadata": {"annotations": {"force-sync": datetime.utcnow().isoformat()}}}
            result = handler.patch(resource_type, name, patch, namespace)
        
        elif operation == "stop":
            # Stop K6 TestRun
            result = handler.delete(resource_type, name, namespace)
        
        else:
            raise ValueError(f"Unknown operation: {operation}")
        
        duration_ms = (time.time() - start_time) * 1000
        
        return {
            "success": True,
            "operation": operation,
            "resource": resource_type,
            "namespace": namespace,
            "name": name,
            "result": result,
            "durationMs": duration_ms,
        }
    
    except ApiException as e:
        logger.error(f"K8s API error: {e.reason}")
        return {
            "success": False,
            "error": str(e.reason),
            "errorCode": e.status,
            "durationMs": (time.time() - start_time) * 1000,
        }
    
    except Exception as e:
        logger.exception(f"Error handling event: {e}")
        return {
            "success": False,
            "error": str(e),
            "durationMs": (time.time() - start_time) * 1000,
        }
