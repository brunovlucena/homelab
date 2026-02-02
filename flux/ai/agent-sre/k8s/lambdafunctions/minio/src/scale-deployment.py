import os
import subprocess
from typing import Dict, Any

def handler(event: Dict[str, Any]) -> Dict[str, Any]:
    """
    Scale a deployment.
    
    Parameters:
    - name: Deployment name (required)
    - namespace: Namespace (required)
    - replicas: Number of replicas (required)
    """
    name = event.get('name')
    namespace = event.get('namespace')
    replicas = event.get('replicas')
    
    if not name or not namespace or replicas is None:
        return {
            'status': 'error',
            'message': 'name, namespace, and replicas parameters are required'
        }
    
    try:
        result = subprocess.run(
            ['kubectl', 'scale', 'deployment', name, '-n', namespace, '--replicas', str(replicas)],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0:
            return {
                'status': 'success',
                'message': f'Deployment {name} scaled to {replicas} replicas',
                'output': result.stdout
            }
        else:
            return {
                'status': 'error',
                'message': f'Scaling failed: {result.stderr}',
                'output': result.stdout
            }
    except Exception as e:
        return {'status': 'error', 'message': f'Error: {str(e)}'}

