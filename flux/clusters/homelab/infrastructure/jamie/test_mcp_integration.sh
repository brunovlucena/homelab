#!/bin/bash
# 🧪 Test Jamie MCP Integration with agent-sre

set -e

echo "🧪 Testing Jamie MCP Integration with agent-sre"
echo "================================================"
echo ""

# Get Jamie pod name
JAMIE_POD=$(kubectl get pods -n jamie -l app=jamie-slack-bot -o jsonpath='{.items[0].metadata.name}')
echo "📦 Jamie Pod: $JAMIE_POD"
echo ""

# Test 1: Check connectivity to MCP server
echo "1️⃣  Testing connectivity to MCP server..."
kubectl exec -n jamie $JAMIE_POD -- python3 -c "
import urllib.request
print(urllib.request.urlopen('http://sre-agent-mcp-server-service.agent-sre:30120/health', timeout=5).read().decode())
"
echo "✅ Connectivity test passed"
echo ""

# Test 2: List available MCP tools
echo "2️⃣  Listing available MCP tools..."
kubectl exec -n jamie $JAMIE_POD -- python3 -c "
import urllib.request
import json

payload = json.dumps({
    'jsonrpc': '2.0',
    'id': 1,
    'method': 'tools/list',
    'params': {}
})

req = urllib.request.Request(
    'http://sre-agent-mcp-server-service.agent-sre:30120/mcp',
    data=payload.encode(),
    headers={'Content-Type': 'application/json'}
)

response = json.loads(urllib.request.urlopen(req).read().decode())
tools = response['result']['tools']

print(f'Found {len(tools)} tools:')
for tool in tools:
    print(f'  - {tool[\"name\"]}: {tool[\"description\"]}')
"
echo ""

# Test 3: Call sre_chat tool
echo "3️⃣  Testing sre_chat tool..."
kubectl exec -n jamie $JAMIE_POD -- python3 -c "
import urllib.request
import json

payload = json.dumps({
    'jsonrpc': '2.0',
    'id': 1,
    'method': 'tools/call',
    'params': {
        'name': 'sre_chat',
        'arguments': {
            'message': 'What are the four golden signals in SRE?'
        }
    }
})

req = urllib.request.Request(
    'http://sre-agent-mcp-server-service.agent-sre:30120/mcp',
    data=payload.encode(),
    headers={'Content-Type': 'application/json'}
)

response = json.loads(urllib.request.urlopen(req).read().decode())
result = response['result']['content'][0]['text']

print('Response from agent-sre:')
print(result[:500])  # Print first 500 chars
if len(result) > 500:
    print('... (truncated)')
"
echo ""

# Test 4: Test health_check tool
echo "4️⃣  Testing health_check tool..."
kubectl exec -n jamie $JAMIE_POD -- python3 -c "
import urllib.request
import json

payload = json.dumps({
    'jsonrpc': '2.0',
    'id': 1,
    'method': 'tools/call',
    'params': {
        'name': 'health_check',
        'arguments': {}
    }
})

req = urllib.request.Request(
    'http://sre-agent-mcp-server-service.agent-sre:30120/mcp',
    data=payload.encode(),
    headers={'Content-Type': 'application/json'}
)

response = json.loads(urllib.request.urlopen(req).read().decode())
result = response['result']['content'][0]['text']

print('Agent health status:')
print(result)
"
echo ""

echo "================================================"
echo "✅ All MCP integration tests passed!"
echo ""
echo "🚀 Jamie can now communicate with agent-sre via MCP"
echo ""
echo "📝 Try asking Jamie in Slack:"
echo "   - 'Can you check the golden signals?'"
echo "   - 'Use agent-sre to analyze these logs: [logs]'"
echo "   - 'What monitoring advice do you have?'"
echo ""

