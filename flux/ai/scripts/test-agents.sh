#!/bin/bash
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Agent Functionality Test Script
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
# Tests agent functionality:
# 1. API endpoints (health, ready, metrics)
# 2. Prometheus metrics
# 3. CloudEvents via broker
# 4. Cross-agent communication
#
# Usage: ./scripts/test-agents.sh [agent-name]
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

set -euo pipefail

AGENT_NAME="${1:-agent-medical}"
NAMESPACE="$AGENT_NAME"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Testing Agent: ${AGENT_NAME}${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Check if agent exists
if ! kubectl get lambdaagent -n "$NAMESPACE" 2>/dev/null | grep -q "$AGENT_NAME"; then
    echo -e "${RED}❌ Agent not found: $AGENT_NAME${NC}"
    exit 1
fi

# Get service URL
SVC_URL=$(kubectl get lambdaagent -n "$NAMESPACE" "$AGENT_NAME" -o jsonpath='{.status.url}' 2>/dev/null || echo "")
if [ -z "$SVC_URL" ]; then
    echo -e "${YELLOW}⚠️  Service URL not available, using port-forward${NC}"
    PORT=8080
    kubectl port-forward -n "$NAMESPACE" svc/"$AGENT_NAME" $PORT:80 > /dev/null 2>&1 &
    PF_PID=$!
    sleep 2
    SVC_URL="http://localhost:$PORT"
    USE_PF=true
else
    USE_PF=false
fi

echo -e "${BLUE}Service URL: ${SVC_URL}${NC}"
echo ""

# Test 1: Health endpoint
echo -e "${BLUE}1. Testing /health endpoint...${NC}"
HEALTH=$(curl -s "${SVC_URL}/health" 2>&1 || echo "FAILED")
if echo "$HEALTH" | grep -q "healthy\|status"; then
    echo -e "${GREEN}✅ Health check passed${NC}"
    echo "$HEALTH" | python3 -m json.tool 2>/dev/null || echo "$HEALTH"
else
    echo -e "${RED}❌ Health check failed${NC}"
    echo "$HEALTH"
fi
echo ""

# Test 2: Ready endpoint
echo -e "${BLUE}2. Testing /ready endpoint...${NC}"
READY=$(curl -s "${SVC_URL}/ready" 2>&1 || echo "FAILED")
if echo "$READY" | grep -q "ready\|true"; then
    echo -e "${GREEN}✅ Readiness check passed${NC}"
    echo "$READY" | python3 -m json.tool 2>/dev/null || echo "$READY"
else
    echo -e "${RED}❌ Readiness check failed${NC}"
    echo "$READY"
fi
echo ""

# Test 3: Metrics endpoint
echo -e "${BLUE}3. Testing /metrics endpoint...${NC}"
METRICS=$(curl -s "${SVC_URL}/metrics" 2>&1 || echo "FAILED")
if echo "$METRICS" | grep -q "# HELP\|# TYPE"; then
    METRIC_COUNT=$(echo "$METRICS" | grep -c "^[^#]" || echo "0")
    echo -e "${GREEN}✅ Metrics endpoint working (${METRIC_COUNT} metrics)${NC}"
    echo "$METRICS" | grep -E "^[^#]" | head -5
else
    echo -e "${YELLOW}⚠️  Metrics endpoint may not be working${NC}"
    echo "$METRICS" | head -5
fi
echo ""

# Test 4: CloudEvents endpoint
echo -e "${BLUE}4. Testing CloudEvents endpoint (POST /)...${NC}"
EVENT_ID="test-$(date +%s)"
CE_RESPONSE=$(curl -s -X POST "${SVC_URL}/" \
    -H "Content-Type: application/cloudevents+json" \
    -H "Ce-Specversion: 1.0" \
    -H "Ce-Type: io.homelab.agent.query" \
    -H "Ce-Source: /test/agent-tester" \
    -H "Ce-Id: $EVENT_ID" \
    -d '{"query":"Test query","test":true}' 2>&1 || echo "FAILED")

if echo "$CE_RESPONSE" | grep -q "error\|status\|response"; then
    echo -e "${GREEN}✅ CloudEvents endpoint responded${NC}"
    echo "$CE_RESPONSE" | python3 -m json.tool 2>/dev/null | head -10 || echo "$CE_RESPONSE" | head -10
else
    echo -e "${YELLOW}⚠️  CloudEvents endpoint response unclear${NC}"
    echo "$CE_RESPONSE" | head -10
fi
echo ""

# Test 5: Check Prometheus metrics
echo -e "${BLUE}5. Checking Prometheus metrics...${NC}"
if command -v kubectl port-forward >/dev/null 2>&1; then
    # Try to query Prometheus
    PROM_POD=$(kubectl get pods -n prometheus -l app.kubernetes.io/name=prometheus -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    if [ -n "$PROM_POD" ]; then
        kubectl port-forward -n prometheus "$PROM_POD" 9091:9090 > /dev/null 2>&1 &
        PROM_PF_PID=$!
        sleep 2
        
        METRIC_NAME=$(echo "$AGENT_NAME" | tr '-' '_')
        QUERY="up{job=~\"$AGENT_NAME.*\"}"
        PROM_RESULT=$(curl -s "http://localhost:9091/api/v1/query?query=$(echo "$QUERY" | sed 's/ /%20/g')" 2>&1 || echo "FAILED")
        
        if echo "$PROM_RESULT" | grep -q "\"status\":\"success\""; then
            echo -e "${GREEN}✅ Prometheus query successful${NC}"
            echo "$PROM_RESULT" | python3 -c "import sys, json; d=json.load(sys.stdin); [print(f\"  {r['metric']}: {r['value'][1]}\") for r in d.get('data', {}).get('result', [])[:3]]" 2>/dev/null || echo "  (parsing failed)"
        else
            echo -e "${YELLOW}⚠️  Prometheus query failed or no data${NC}"
        fi
        
        kill $PROM_PF_PID 2>/dev/null || true
    else
        echo -e "${YELLOW}⚠️  Prometheus pod not found${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Cannot test Prometheus (kubectl not available)${NC}"
fi
echo ""

# Test 6: Check broker and triggers
echo -e "${BLUE}6. Checking CloudEvents infrastructure...${NC}"
BROKER_COUNT=$(kubectl get broker -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l | tr -d ' ')
TRIGGER_COUNT=$(kubectl get trigger -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l | tr -d ' ')
READY_TRIGGERS=$(kubectl get trigger -n "$NAMESPACE" --no-headers 2>/dev/null | grep -c "True" || echo "0")

echo "  Brokers: $BROKER_COUNT"
echo "  Triggers: $TRIGGER_COUNT (${READY_TRIGGERS} ready)"

if [ "$BROKER_COUNT" -gt 0 ] && [ "$TRIGGER_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✅ CloudEvents infrastructure configured${NC}"
else
    echo -e "${YELLOW}⚠️  CloudEvents infrastructure may be incomplete${NC}"
fi
echo ""

# Cleanup
if [ "$USE_PF" = true ]; then
    kill $PF_PID 2>/dev/null || true
fi

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ Testing complete for ${AGENT_NAME}${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
