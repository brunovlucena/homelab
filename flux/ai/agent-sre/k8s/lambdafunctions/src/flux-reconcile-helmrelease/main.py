import os
import subprocess
from typing import Dict, Any

def handler(event: Dict[str, Any]) -> Dict[str, Any]:
    name = event.get('name')
    namespace = event.get('namespace', 'flux-system')
    
    if not name:
        return {'status': 'error', 'message': 'name parameter is required'}
    
    try:
        result = subprocess.run(
            ['flux', 'reconcile', 'helmrelease', name, '-n', namespace],
            capture_output=True,
            text=True,
            timeout=60
        )
        
        if result.returncode == 0:
            return {
                'status': 'success',
                'message': f'HelmRelease {name} reconciled successfully',
                'output': result.stdout
            }
        else:
            return {
                'status': 'error',
                'message': f'Reconciliation failed: {result.stderr}',
                'output': result.stdout
            }
    except Exception as e:
        return {'status': 'error', 'message': f'Error: {str(e)}'}
