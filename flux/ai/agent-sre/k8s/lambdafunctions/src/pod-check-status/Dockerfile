# LambdaFunction: pod-check-status
# Pre-built Docker image for CI/CD pipeline
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt ./
RUN if [ -s requirements.txt ]; then pip install --no-cache-dir -r requirements.txt; fi

# Copy function code
COPY main.py ./

# Expose port for Knative
ENV PORT=8080
EXPOSE 8080

# Simple HTTP server wrapper for Lambda handler
COPY <<'EOF' /app/server.py
from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import os
from main import handler

class LambdaHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(content_length)
        
        try:
            event = json.loads(body.decode('utf-8'))
            result = handler(event)
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(result).encode('utf-8'))
        except Exception as e:
            self.send_response(500)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({'status': 'error', 'message': str(e)}).encode('utf-8'))
    
    def do_GET(self):
        if self.path == '/health' or self.path == '/healthz':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({'status': 'healthy'}).encode('utf-8'))
        else:
            self.send_response(404)
            self.end_headers()
    
    def log_message(self, format, *args):
        pass

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8080))
    server = HTTPServer(('', port), LambdaHandler)
    print(f'Lambda function server listening on port {port}')
    server.serve_forever()
EOF

CMD ["python", "/app/server.py"]
