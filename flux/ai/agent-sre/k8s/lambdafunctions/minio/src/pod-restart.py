import os
import subprocess
import json
from typing import Dict, Any

def handler(event: Dict[str, Any]) -> Dict[str, Any]:
    """
    Restart a pod or deployment.
    
    Parameters:
    - name: Pod or deployment name (required)
    - namespace: Namespace (required)
    - type: 'pod' or 'deployment' (default: pod)
    """
    name = event.get('name')
    namespace = event.get('namespace')
    resource_type = event.get('type', 'pod')
    
    if not name or not namespace:
        return {
            'status': 'error',
            'message': 'name and namespace parameters are required'
        }
    
    try:
        if resource_type == 'deployment':
            # Restart deployment
            result = subprocess.run(
                ['kubectl', 'rollout', 'restart', f'deployment/{name}', '-n', namespace],
                capture_output=True,
                text=True,
                timeout=30
            )
        else:
            # Delete pod (will be recreated)
            result = subprocess.run(
                ['kubectl', 'delete', 'pod', name, '-n', namespace],
                capture_output=True,
                text=True,
                timeout=30
            )
        
        if result.returncode == 0:
            return {
                'status': 'success',
                'message': f'{resource_type} {name} restarted successfully',
                'output': result.stdout
            }
        else:
            return {
                'status': 'error',
                'message': f'Restart failed: {result.stderr}',
                'output': result.stdout
            }
    except subprocess.TimeoutExpired:
        return {
            'status': 'error',
            'message': 'Restart operation timed out'
        }
    except Exception as e:
        return {
            'status': 'error',
            'message': f'Unexpected error: {str(e)}'
        }

