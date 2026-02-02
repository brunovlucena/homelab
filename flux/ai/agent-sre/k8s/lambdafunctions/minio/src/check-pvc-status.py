import os
import subprocess
import json
from typing import Dict, Any

def handler(event: Dict[str, Any]) -> Dict[str, Any]:
    """
    Check PVC status and usage.
    
    Parameters:
    - name: PVC name (optional, if not provided checks all in namespace)
    - namespace: Namespace (required)
    """
    name = event.get('name')
    namespace = event.get('namespace')
    
    if not namespace:
        return {'status': 'error', 'message': 'namespace parameter is required'}
    
    try:
        if name:
            cmd = ['kubectl', 'get', 'pvc', name, '-n', namespace, '-o', 'json']
        else:
            cmd = ['kubectl', 'get', 'pvc', '-n', namespace, '-o', 'json']
        
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        
        if result.returncode == 0:
            pvc_data = json.loads(result.stdout)
            return {
                'status': 'success',
                'pvc': pvc_data,
                'message': 'PVC status retrieved successfully'
            }
        else:
            return {
                'status': 'error',
                'message': f'Failed to get PVC status: {result.stderr}'
            }
    except Exception as e:
        return {'status': 'error', 'message': f'Error: {str(e)}'}

