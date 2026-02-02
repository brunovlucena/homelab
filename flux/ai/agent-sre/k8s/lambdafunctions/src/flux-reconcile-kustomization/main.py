import os
import subprocess
import json
from typing import Dict, Any

def handler(event: Dict[str, Any]) -> Dict[str, Any]:
    """
    Reconcile Flux Kustomization resource.
    
    Parameters:
    - name: Kustomization name (required)
    - namespace: Namespace (default: flux-system)
    """
    name = event.get('name')
    namespace = event.get('namespace', 'flux-system')
    
    if not name:
        return {
            'status': 'error',
            'message': 'name parameter is required'
        }
    
    try:
        # Execute flux reconcile command
        result = subprocess.run(
            ['flux', 'reconcile', 'kustomization', name, '-n', namespace],
            capture_output=True,
            text=True,
            timeout=60
        )
        
        if result.returncode == 0:
            return {
                'status': 'success',
                'message': f'Kustomization {name} reconciled successfully',
                'output': result.stdout
            }
        else:
            return {
                'status': 'error',
                'message': f'Reconciliation failed: {result.stderr}',
                'output': result.stdout
            }
    except subprocess.TimeoutExpired:
        return {
            'status': 'error',
            'message': 'Reconciliation timed out after 60 seconds'
        }
    except Exception as e:
        return {
            'status': 'error',
            'message': f'Unexpected error: {str(e)}'
        }
