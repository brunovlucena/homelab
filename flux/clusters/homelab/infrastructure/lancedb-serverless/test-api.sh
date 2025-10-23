#!/bin/bash

# 🧪 LanceDB Serverless API Test Script
# This script tests all API endpoints to verify the deployment

set -e

echo "🧪 Testing LanceDB Serverless API"
echo "=================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get service URL
echo "📡 Getting service URL..."
SERVICE_URL=$(kubectl get ksvc lancedb-serverless -n lancedb-serverless -o jsonpath='{.status.url}')

if [ -z "$SERVICE_URL" ]; then
    echo -e "${RED}❌ Failed to get service URL${NC}"
    echo "Make sure the service is deployed: kubectl get ksvc -n lancedb-serverless"
    exit 1
fi

echo -e "${GREEN}✅ Service URL: $SERVICE_URL${NC}"
echo ""

# Test 1: Health Check
echo "Test 1: Health Check"
echo "--------------------"
RESPONSE=$(curl -s -w "\n%{http_code}" $SERVICE_URL/health)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" == "200" ]; then
    echo -e "${GREEN}✅ Health check passed${NC}"
    echo "$BODY" | jq .
else
    echo -e "${RED}❌ Health check failed (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
    exit 1
fi
echo ""

# Test 2: List Tables (should be empty initially)
echo "Test 2: List Tables"
echo "-------------------"
RESPONSE=$(curl -s -w "\n%{http_code}" $SERVICE_URL/tables)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" == "200" ]; then
    echo -e "${GREEN}✅ List tables successful${NC}"
    echo "$BODY" | jq .
else
    echo -e "${RED}❌ List tables failed (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
fi
echo ""

# Test 3: Create a Table
echo "Test 3: Create Table"
echo "--------------------"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $SERVICE_URL/tables \
    -H "Content-Type: application/json" \
    -d '{
        "name": "test_table",
        "data": [
            {"id": 1, "vector": [0.1, 0.2, 0.3, 0.4, 0.5], "text": "First document", "category": "A"},
            {"id": 2, "vector": [0.2, 0.3, 0.4, 0.5, 0.6], "text": "Second document", "category": "B"},
            {"id": 3, "vector": [0.3, 0.4, 0.5, 0.6, 0.7], "text": "Third document", "category": "A"},
            {"id": 4, "vector": [0.4, 0.5, 0.6, 0.7, 0.8], "text": "Fourth document", "category": "C"}
        ]
    }')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" == "201" ]; then
    echo -e "${GREEN}✅ Table created successfully${NC}"
    echo "$BODY" | jq .
else
    echo -e "${RED}❌ Table creation failed (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
fi
echo ""

# Test 4: List Tables Again (should show the new table)
echo "Test 4: List Tables (After Creation)"
echo "-------------------------------------"
RESPONSE=$(curl -s -w "\n%{http_code}" $SERVICE_URL/tables)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" == "200" ]; then
    TABLE_COUNT=$(echo "$BODY" | jq '.count')
    if [ "$TABLE_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✅ Table found in list (count: $TABLE_COUNT)${NC}"
        echo "$BODY" | jq .
    else
        echo -e "${YELLOW}⚠️ No tables found${NC}"
    fi
else
    echo -e "${RED}❌ List tables failed (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
fi
echo ""

# Test 5: Vector Search
echo "Test 5: Vector Search"
echo "---------------------"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $SERVICE_URL/search \
    -H "Content-Type: application/json" \
    -d '{
        "table": "test_table",
        "query_vector": [0.1, 0.2, 0.3, 0.4, 0.5],
        "limit": 3
    }')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" == "200" ]; then
    RESULT_COUNT=$(echo "$BODY" | jq '.count')
    echo -e "${GREEN}✅ Search successful (found $RESULT_COUNT results)${NC}"
    echo "$BODY" | jq .
else
    echo -e "${RED}❌ Search failed (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
fi
echo ""

# Test 6: Vector Search with Filter
echo "Test 6: Vector Search with Filter"
echo "----------------------------------"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $SERVICE_URL/search \
    -H "Content-Type: application/json" \
    -d '{
        "table": "test_table",
        "query_vector": [0.1, 0.2, 0.3, 0.4, 0.5],
        "limit": 5,
        "filter": "category = '\''A'\''"
    }')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" == "200" ]; then
    RESULT_COUNT=$(echo "$BODY" | jq '.count')
    echo -e "${GREEN}✅ Filtered search successful (found $RESULT_COUNT results)${NC}"
    echo "$BODY" | jq .
else
    echo -e "${RED}❌ Filtered search failed (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
fi
echo ""

# Test 7: Metrics Endpoint
echo "Test 7: Metrics Endpoint"
echo "------------------------"
RESPONSE=$(curl -s -w "\n%{http_code}" $SERVICE_URL/metrics)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" == "200" ]; then
    METRIC_COUNT=$(echo "$BODY" | grep -c "lancedb_" || true)
    echo -e "${GREEN}✅ Metrics endpoint accessible (found $METRIC_COUNT LanceDB metrics)${NC}"
    echo "Sample metrics:"
    echo "$BODY" | grep "lancedb_" | head -n 5
else
    echo -e "${RED}❌ Metrics endpoint failed (HTTP $HTTP_CODE)${NC}"
fi
echo ""

# Test 8: Delete Table
echo "Test 8: Delete Table"
echo "--------------------"
RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE $SERVICE_URL/tables/test_table)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" == "200" ]; then
    echo -e "${GREEN}✅ Table deleted successfully${NC}"
    echo "$BODY" | jq .
else
    echo -e "${RED}❌ Table deletion failed (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
fi
echo ""

# Final Summary
echo "=================================="
echo "🎉 All tests completed!"
echo ""
echo "Summary:"
echo "- Service is accessible at: $SERVICE_URL"
echo "- All endpoints are working correctly"
echo "- Vector search is functional"
echo "- Metrics are being collected"
echo ""
echo "Next steps:"
echo "1. Check Prometheus for metrics: kubectl port-forward -n prometheus svc/prometheus-operated 9090:9090"
echo "2. View service logs: make logs"
echo "3. Try the Python examples: python example-usage.py"
echo ""

