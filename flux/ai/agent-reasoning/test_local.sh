#!/bin/bash
# Local testing script for Agent-Reasoning

set -e

CONTAINER_NAME="agent-reasoning-local"
PORT=8080

echo "🧪 Testing Agent-Reasoning Locally"
echo ""

# Check if container is already running
if docker ps | grep -q $CONTAINER_NAME; then
    echo "⚠️  Container $CONTAINER_NAME is already running"
    echo "   Stopping existing container..."
    docker stop $CONTAINER_NAME > /dev/null 2>&1 || true
    docker rm $CONTAINER_NAME > /dev/null 2>&1 || true
fi

# Start the container
echo "🚀 Starting container..."
docker run -d \
    --name $CONTAINER_NAME \
    -p $PORT:8080 \
    -e MODEL_PATH=/models/trm-checkpoint.pth \
    -e DEVICE=cpu \
    -e H_CYCLES=3 \
    -e L_CYCLES=6 \
    agent-reasoning:latest

echo "⏳ Waiting for service to start..."
sleep 5

# Test health endpoint
echo ""
echo "📊 Testing Health Endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:$PORT/health)
echo "$HEALTH_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$HEALTH_RESPONSE"

# Test root endpoint
echo ""
echo "📋 Testing Root Endpoint..."
ROOT_RESPONSE=$(curl -s http://localhost:$PORT/)
echo "$ROOT_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$ROOT_RESPONSE"

# Test reasoning endpoint
echo ""
echo "🧠 Testing Reasoning Endpoint..."
REASONING_RESPONSE=$(curl -s -X POST http://localhost:$PORT/reason \
    -H "Content-Type: application/json" \
    -d '{
        "question": "How should I optimize my Kubernetes cluster?",
        "context": {
            "nodes": 10,
            "workloads": ["app1", "app2"]
        },
        "max_steps": 6,
        "task_type": "optimization"
    }')

echo "$REASONING_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$REASONING_RESPONSE"

# Test metrics endpoint
echo ""
echo "📈 Testing Metrics Endpoint..."
METRICS_RESPONSE=$(curl -s http://localhost:$PORT/metrics | head -20)
echo "$METRICS_RESPONSE"

echo ""
echo "✅ Testing complete!"
echo ""
echo "Container is running. To stop it, run:"
echo "  docker stop $CONTAINER_NAME"
echo ""
echo "To view logs:"
echo "  docker logs -f $CONTAINER_NAME"
echo ""
echo "To test manually:"
echo "  curl http://localhost:$PORT/health"
echo "  curl -X POST http://localhost:$PORT/reason -H 'Content-Type: application/json' -d '{\"question\": \"test\", \"max_steps\": 3}'"


