#!/bin/bash
# Quick script to verify Cloudflare Page Rules were created

API_TOKEN=$(kubectl get secret cloudflare-api-token -n homepage -o jsonpath='{.data.api-token}' 2>/dev/null | base64 -d)

if [ -z "${API_TOKEN}" ]; then
  echo "❌ Could not get API token from secret"
  exit 1
fi

ZONE_ID="83a5da98dae8e39a4cde8bb26ca41bed"
API_BASE="https://api.cloudflare.com/client/v4"

echo "🔍 Fetching all Page Rules for lucena.cloud..."
echo ""

RESPONSE=$(curl -s -X GET "${API_BASE}/zones/${ZONE_ID}/pagerules?match=all" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json")

echo "Raw API Response:"
echo "${RESPONSE}" | head -50
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Parse and display rules
SUCCESS=$(echo "${RESPONSE}" | grep -o '"success":[^,}]*' | cut -d':' -f2 | tr -d ' ')

if [ "${SUCCESS}" != "true" ]; then
  echo "❌ API call failed"
  exit 1
fi

RULE_COUNT=$(echo "${RESPONSE}" | grep -o '"id":"[^"]*"' | wc -l | tr -d ' ')
echo "📊 Found ${RULE_COUNT} Page Rule(s)"
echo ""

if [ "${RULE_COUNT}" -eq 0 ]; then
  echo "❌ NO PAGE RULES FOUND!"
  exit 1
fi

# Extract each rule
echo "Page Rules:"
echo "${RESPONSE}" | grep -o '"id":"[^"]*"' | cut -d'"' -f4 | while read RULE_ID; do
  if [ -n "${RULE_ID}" ]; then
    echo ""
    echo "Rule ID: ${RULE_ID}"
    RULE_DETAILS=$(curl -s -X GET "${API_BASE}/zones/${ZONE_ID}/pagerules/${RULE_ID}" \
      -H "Authorization: Bearer ${API_TOKEN}" \
      -H "Content-Type: application/json")
    
    PATTERN=$(echo "${RULE_DETAILS}" | grep -o '"value":"[^"]*"' | head -1 | cut -d'"' -f4)
    STATUS=$(echo "${RULE_DETAILS}" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
    
    echo "  Pattern: ${PATTERN}"
    echo "  Status: ${STATUS}"
    
    # Get cache level
    CACHE_LEVEL=$(echo "${RULE_DETAILS}" | grep -o '"cache_level":"[^"]*"' | cut -d'"' -f4 || echo "N/A")
    EDGE_TTL=$(echo "${RULE_DETAILS}" | grep -o '"edge_cache_ttl":[0-9]*' | cut -d':' -f2 || echo "N/A")
    
    echo "  Cache Level: ${CACHE_LEVEL}"
    echo "  Edge TTL: ${EDGE_TTL} seconds"
  fi
done
