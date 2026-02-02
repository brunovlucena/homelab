"""AgentChat Command Center - Platform orchestration and admin"""
import os, json, logging
from flask import Flask, request, jsonify
from cloudevents.http import from_http

logging.basicConfig(level=os.getenv('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)
app = Flask(__name__)

@app.route('/', methods=['POST'])
def handle_cloudevent():
    try:
        event = from_http(request.headers, request.get_data())
        logger.info(f"Command center request: {event.get('type')}")
        return jsonify({'status': 'processed', 'agent': 'command-center'}), 200
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/health', methods=['GET'])
def health():
    return jsonify({'status': 'healthy', 'service': 'command-center'}), 200

@app.route('/ready', methods=['GET'])
def ready():
    return jsonify({'status': 'ready', 'service': 'command-center'}), 200

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=int(os.getenv('PORT', 8080)))
