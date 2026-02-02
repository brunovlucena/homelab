import os
import subprocess
import json
from typing import Dict, Any

def handler(event: Dict[str, Any]) -> Dict[str, Any]:
    """
    Check pod status and return details.
    
    Parameters:
    - name: Pod name or label selector (required)
    - namespace: Namespace (required)
    - selector: Label selector (optional, if name not provided)
    """
    name = event.get('name')
    namespace = event.get('namespace')
    selector = event.get('selector')
    
    if not namespace:
        return {'status': 'error', 'message': 'namespace parameter is required'}
    
    try:
        if selector:
            cmd = ['kubectl', 'get', 'pods', '-n', namespace, '-l', selector, '-o', 'json']
        elif name:
            cmd = ['kubectl', 'get', 'pod', name, '-n', namespace, '-o', 'json']
        else:
            return {'status': 'error', 'message': 'name or selector parameter is required'}
        
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        
        if result.returncode == 0:
            pod_data = json.loads(result.stdout)
            return {
                'status': 'success',
                'pods': pod_data,
                'message': 'Pod status retrieved successfully'
            }
        else:
            return {
                'status': 'error',
                'message': f'Failed to get pod status: {result.stderr}'
            }
    except Exception as e:
        return {'status': 'error', 'message': f'Error: {str(e)}'}

