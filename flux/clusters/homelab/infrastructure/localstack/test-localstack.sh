#!/bin/bash

# 🧪 LocalStack Test Script
# Tests LocalStack deployment and basic functionality

set -e

echo "🔍 Testing LocalStack Deployment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl not found${NC}"
    exit 1
fi

# Check if aws CLI is available
if ! command -v aws &> /dev/null; then
    echo -e "${YELLOW}⚠️  AWS CLI not found. Install it to run DynamoDB tests.${NC}"
    AWS_AVAILABLE=false
else
    AWS_AVAILABLE=true
fi

echo ""
echo "1️⃣  Checking LocalStack namespace..."
if kubectl get namespace localstack &> /dev/null; then
    echo -e "${GREEN}✅ Namespace 'localstack' exists${NC}"
else
    echo -e "${RED}❌ Namespace 'localstack' not found${NC}"
    exit 1
fi

echo ""
echo "2️⃣  Checking LocalStack pods..."
POD_STATUS=$(kubectl get pods -n localstack -l app.kubernetes.io/name=localstack -o jsonpath='{.items[0].status.phase}' 2>/dev/null || echo "NotFound")

if [ "$POD_STATUS" == "Running" ]; then
    echo -e "${GREEN}✅ LocalStack pod is running${NC}"
    kubectl get pods -n localstack -l app.kubernetes.io/name=localstack
else
    echo -e "${RED}❌ LocalStack pod is not running (Status: $POD_STATUS)${NC}"
    echo "Pod details:"
    kubectl get pods -n localstack
    echo ""
    echo "Recent logs:"
    kubectl logs -n localstack -l app.kubernetes.io/name=localstack --tail=20
    exit 1
fi

echo ""
echo "3️⃣  Checking LocalStack service..."
if kubectl get svc -n localstack localstack &> /dev/null; then
    echo -e "${GREEN}✅ LocalStack service exists${NC}"
    kubectl get svc -n localstack localstack
else
    echo -e "${RED}❌ LocalStack service not found${NC}"
    exit 1
fi

echo ""
echo "4️⃣  Checking LocalStack health endpoint..."
echo "Starting port-forward in background..."
kubectl port-forward -n localstack svc/localstack 4566:4566 &> /dev/null &
PORT_FORWARD_PID=$!

# Wait for port-forward to be ready
sleep 3

# Check health endpoint
if curl -s http://localhost:4566/_localstack/health &> /dev/null; then
    echo -e "${GREEN}✅ LocalStack health endpoint is responding${NC}"
    echo ""
    echo "Health status:"
    curl -s http://localhost:4566/_localstack/health | jq '.' || curl -s http://localhost:4566/_localstack/health
else
    echo -e "${RED}❌ LocalStack health endpoint is not responding${NC}"
    kill $PORT_FORWARD_PID 2>/dev/null
    exit 1
fi

echo ""
echo "5️⃣  Testing DynamoDB..."
if [ "$AWS_AVAILABLE" = true ]; then
    export AWS_ENDPOINT_URL=http://localhost:4566
    export AWS_ACCESS_KEY_ID=test
    export AWS_SECRET_ACCESS_KEY=test
    export AWS_DEFAULT_REGION=us-east-1
    
    # Create test table
    echo "Creating test table..."
    TABLE_NAME="test-table-$(date +%s)"
    
    if aws dynamodb create-table \
        --table-name "$TABLE_NAME" \
        --attribute-definitions AttributeName=id,AttributeType=S \
        --key-schema AttributeName=id,KeyType=HASH \
        --billing-mode PAY_PER_REQUEST \
        --endpoint-url $AWS_ENDPOINT_URL &> /dev/null; then
        echo -e "${GREEN}✅ Table created successfully${NC}"
    else
        echo -e "${RED}❌ Failed to create table${NC}"
        kill $PORT_FORWARD_PID 2>/dev/null
        exit 1
    fi
    
    # Put item
    echo "Inserting test item..."
    if aws dynamodb put-item \
        --table-name "$TABLE_NAME" \
        --item '{"id": {"S": "test-id"}, "name": {"S": "Test User"}}' \
        --endpoint-url $AWS_ENDPOINT_URL &> /dev/null; then
        echo -e "${GREEN}✅ Item inserted successfully${NC}"
    else
        echo -e "${RED}❌ Failed to insert item${NC}"
        kill $PORT_FORWARD_PID 2>/dev/null
        exit 1
    fi
    
    # Get item
    echo "Retrieving test item..."
    ITEM=$(aws dynamodb get-item \
        --table-name "$TABLE_NAME" \
        --key '{"id": {"S": "test-id"}}' \
        --endpoint-url $AWS_ENDPOINT_URL \
        --output json 2>/dev/null || echo "")
    
    if [ -n "$ITEM" ] && echo "$ITEM" | jq -e '.Item.name.S == "Test User"' &> /dev/null; then
        echo -e "${GREEN}✅ Item retrieved successfully${NC}"
    else
        echo -e "${RED}❌ Failed to retrieve item${NC}"
        kill $PORT_FORWARD_PID 2>/dev/null
        exit 1
    fi
    
    # List tables
    echo "Listing tables..."
    if aws dynamodb list-tables --endpoint-url $AWS_ENDPOINT_URL &> /dev/null; then
        TABLES=$(aws dynamodb list-tables --endpoint-url $AWS_ENDPOINT_URL --output json | jq -r '.TableNames[]' | wc -l)
        echo -e "${GREEN}✅ Found $TABLES table(s)${NC}"
    fi
    
    # Cleanup test table
    echo "Cleaning up test table..."
    aws dynamodb delete-table --table-name "$TABLE_NAME" --endpoint-url $AWS_ENDPOINT_URL &> /dev/null || true
    
else
    echo -e "${YELLOW}⚠️  Skipping DynamoDB tests (AWS CLI not available)${NC}"
fi

echo ""
echo "6️⃣  Checking persistence..."
if kubectl get pvc -n localstack &> /dev/null; then
    echo -e "${GREEN}✅ PVC exists${NC}"
    kubectl get pvc -n localstack
else
    echo -e "${YELLOW}⚠️  No PVC found (data may not persist)${NC}"
fi

# Cleanup port-forward
kill $PORT_FORWARD_PID 2>/dev/null

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ All tests passed!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "LocalStack is ready to use! 🚀"
echo ""
echo "To access LocalStack:"
echo "  kubectl port-forward -n localstack svc/localstack 4566:4566"
echo ""
echo "Then use: http://localhost:4566"
echo ""
echo "AWS credentials (use with AWS CLI/SDK):"
echo "  AWS_ENDPOINT_URL=http://localhost:4566"
echo "  AWS_ACCESS_KEY_ID=test"
echo "  AWS_SECRET_ACCESS_KEY=test"
echo "  AWS_DEFAULT_REGION=us-east-1"

