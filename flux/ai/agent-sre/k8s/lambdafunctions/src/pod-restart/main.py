import os
import json
from typing import Dict, Any
from kubernetes import client, config
from kubernetes.client.rest import ApiException

def handler(event: Dict[str, Any]) -> Dict[str, Any]:
    """
    Restart a pod or deployment using Kubernetes API.
    
    Parameters:
    - name: Pod or deployment name (required)
    - namespace: Namespace (required)
    - type: 'pod' or 'deployment' (default: pod)
    """
    name = event.get('name')
    namespace = event.get('namespace', 'ai')
    resource_type = event.get('type', 'pod')
    
    if not name:
        return {
            'status': 'error',
            'message': 'name parameter is required'
        }
    
    try:
        # Load in-cluster config (runs in Kubernetes)
        config.load_incluster_config()
        
        if resource_type == 'deployment':
            # Restart deployment by patching annotation
            apps_v1 = client.AppsV1Api()
            deployment = apps_v1.read_namespaced_deployment(name, namespace)
            
            # Add restart annotation to trigger rollout
            from datetime import datetime
            if deployment.spec.template.metadata.annotations is None:
                deployment.spec.template.metadata.annotations = {}
            deployment.spec.template.metadata.annotations['kubectl.kubernetes.io/restartedAt'] = datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%SZ')
            
            apps_v1.patch_namespaced_deployment(
                name=name,
                namespace=namespace,
                body=deployment
            )
            
            return {
                'status': 'success',
                'message': f'Deployment {name} in namespace {namespace} restarted successfully'
            }
        else:
            # Delete pod (will be recreated by controller)
            core_v1 = client.CoreV1Api()
            core_v1.delete_namespaced_pod(
                name=name,
                namespace=namespace,
                body=client.V1DeleteOptions(grace_period_seconds=0)
            )
            
            return {
                'status': 'success',
                'message': f'Pod {name} in namespace {namespace} deleted successfully (will be recreated)'
            }
            
    except ApiException as e:
        return {
            'status': 'error',
            'message': f'Kubernetes API error: {e.reason}',
            'details': json.loads(e.body) if e.body else {}
        }
    except Exception as e:
        return {
            'status': 'error',
            'message': f'Unexpected error: {str(e)}'
        }

