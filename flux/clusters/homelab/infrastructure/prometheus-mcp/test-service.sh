#!/bin/bash
# Test script for Prometheus MCP Knative Service

set -e

echo "🔍 Testing Prometheus MCP Knative Service..."
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if service exists
echo "1️⃣ Checking Knative Service status..."
if kubectl get ksvc prometheus-mcp -n prometheus-mcp &>/dev/null; then
    echo -e "${GREEN}✓ Knative Service exists${NC}"
    kubectl get ksvc prometheus-mcp -n prometheus-mcp
else
    echo -e "${RED}✗ Knative Service not found${NC}"
    exit 1
fi

echo ""

# Check if pods are running
echo "2️⃣ Checking pods..."
POD_COUNT=$(kubectl get pods -n prometheus-mcp -l app=prometheus-mcp --field-selector=status.phase=Running --no-headers 2>/dev/null | wc -l)
if [ "$POD_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Found $POD_COUNT running pod(s)${NC}"
    kubectl get pods -n prometheus-mcp -l app=prometheus-mcp
else
    echo -e "${YELLOW}⚠ No running pods found (service may be scaled to zero)${NC}"
fi

echo ""

# Get service URL
echo "3️⃣ Getting Knative Service URL..."
SERVICE_URL=$(kubectl get ksvc prometheus-mcp -n prometheus-mcp -o jsonpath='{.status.url}')
if [ -n "$SERVICE_URL" ]; then
    echo -e "${GREEN}✓ Service URL: $SERVICE_URL${NC}"
else
    echo -e "${RED}✗ Could not get service URL${NC}"
    exit 1
fi

echo ""

# Port forward to test
echo "4️⃣ Setting up port forward for testing..."
echo -e "${YELLOW}Starting port forward on localhost:8000...${NC}"
kubectl port-forward -n prometheus-mcp svc/prometheus-mcp 8000:80 &
PF_PID=$!

# Wait for port forward to be ready
sleep 3

echo ""

# Test health endpoint
echo "5️⃣ Testing health endpoint..."
if curl -s http://localhost:8000/health &>/dev/null; then
    echo -e "${GREEN}✓ Health endpoint is responding${NC}"
    curl -s http://localhost:8000/health | jq . 2>/dev/null || curl -s http://localhost:8000/health
else
    echo -e "${RED}✗ Health endpoint is not responding${NC}"
fi

echo ""

# Test ready endpoint
echo "6️⃣ Testing ready endpoint..."
if curl -s http://localhost:8000/ready &>/dev/null; then
    echo -e "${GREEN}✓ Ready endpoint is responding${NC}"
    curl -s http://localhost:8000/ready | jq . 2>/dev/null || curl -s http://localhost:8000/ready
else
    echo -e "${RED}✗ Ready endpoint is not responding${NC}"
fi

echo ""

# Check Prometheus connectivity
echo "7️⃣ Testing Prometheus connectivity..."
POD_NAME=$(kubectl get pods -n prometheus-mcp -l app=prometheus-mcp --field-selector=status.phase=Running -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
if [ -n "$POD_NAME" ]; then
    echo "Testing from pod: $POD_NAME"
    if kubectl exec -n prometheus-mcp "$POD_NAME" -- curl -s http://prometheus-kube-prometheus-prometheus.prometheus.svc.cluster.local:9090/api/v1/status/config &>/dev/null; then
        echo -e "${GREEN}✓ Can connect to Prometheus${NC}"
    else
        echo -e "${RED}✗ Cannot connect to Prometheus${NC}"
    fi
else
    echo -e "${YELLOW}⚠ No pod available for testing${NC}"
fi

echo ""

# Cleanup
echo "8️⃣ Cleaning up..."
kill $PF_PID 2>/dev/null || true
echo -e "${GREEN}✓ Port forward stopped${NC}"

echo ""
echo "🎉 Testing complete!"
echo ""
echo "To access the service:"
echo "  1. Port forward: kubectl port-forward -n prometheus-mcp svc/prometheus-mcp 8000:80"
echo "  2. Or access via cluster URL: $SERVICE_URL"
echo ""

