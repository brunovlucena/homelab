#!/bin/bash
# Quick local test script for agent-medical

set -e

echo "🏥 Testing agent-medical locally..."

# Set environment variables
export OLLAMA_URL="http://localhost:11434"
export MONGODB_URL="mongodb://localhost:27017/test_db"
export HIPAA_MODE="false"
export VAULT_ADDR="http://localhost:8200"

# Install dependencies if needed
if [ ! -d "venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv venv
fi

source venv/bin/activate
pip install -q -r src/requirements.txt
pip install -q -r tests/requirements.txt

# Run tests
echo "Running unit tests..."
python3 -m pytest tests/unit/ -v

echo "Running health check tests..."
python3 -m pytest tests/test_health.py -v

echo "✅ Tests completed!"
