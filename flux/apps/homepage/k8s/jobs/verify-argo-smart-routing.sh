#!/bin/bash
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#  🔍 VERIFY ARGO SMART ROUTING - Check Cloudflare Argo Status
#
#  Purpose: Verify that Argo Smart Routing is enabled for lucena.cloud
#  Related: BVL-23 - Enable Cloudflare Argo Smart Routing
#
#  Usage: ./verify-argo-smart-routing.sh
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

set -e

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔍 Verifying Cloudflare Argo Smart Routing for lucena.cloud"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Get API token from Kubernetes secret (if running in cluster)
if command -v kubectl &>/dev/null; then
  API_TOKEN=$(kubectl get secret cloudflare-api-token -n homepage -o jsonpath='{.data.api-token}' 2>/dev/null | base64 -d 2>/dev/null || echo "")
  
  if [ -z "${API_TOKEN}" ]; then
    # Try getting from cloudflare-tunnel namespace
    API_TOKEN=$(kubectl get secret cloudflare -n cloudflare-tunnel -o jsonpath='{.data.cloudflare-api-token}' 2>/dev/null | base64 -d 2>/dev/null || echo "")
  fi
fi

# Fallback to environment variable
if [ -z "${API_TOKEN}" ]; then
  API_TOKEN="${CLOUDFLARE_API_TOKEN:-}"
fi

if [ -z "${API_TOKEN}" ]; then
  echo "❌ Could not get API token from:"
  echo "   - Kubernetes secret cloudflare-api-token in homepage namespace"
  echo "   - Kubernetes secret cloudflare in cloudflare-tunnel namespace"
  echo "   - CLOUDFLARE_API_TOKEN environment variable"
  echo ""
  echo "💡 Set CLOUDFLARE_API_TOKEN environment variable and try again"
  exit 1
fi

echo "✅ Retrieved Cloudflare API token"
echo ""

DOMAIN="lucena.cloud"
API_BASE="https://api.cloudflare.com/client/v4"

# Get Zone ID
echo "🔍 Getting Zone ID for ${DOMAIN}..."
ZONE_RESPONSE=$(curl -s -X GET "${API_BASE}/zones?name=${DOMAIN}" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json")

ZONE_ID=$(echo "${ZONE_RESPONSE}" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "${ZONE_ID}" ]; then
  echo "❌ Error: Could not find zone ID for ${DOMAIN}"
  echo "Response: ${ZONE_RESPONSE}"
  exit 1
fi

echo "✅ Found Zone ID: ${ZONE_ID}"
echo ""

# Check Argo Smart Routing status
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Argo Smart Routing Status"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

ARGO_RESPONSE=$(curl -s -X GET "${API_BASE}/zones/${ZONE_ID}/argo/smart_routing" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json")

# Check if the API call was successful
SUCCESS=$(echo "${ARGO_RESPONSE}" | grep -o '"success":[^,}]*' | cut -d':' -f2 | tr -d ' ')

if [ "${SUCCESS}" != "true" ]; then
  echo "❌ API call failed"
  echo "Response: ${ARGO_RESPONSE}"
  exit 1
fi

# Extract Argo status
ARGO_STATUS=$(echo "${ARGO_RESPONSE}" | grep -o '"value":"[^"]*"' | cut -d'"' -f4)
ARGO_EDITABLE=$(echo "${ARGO_RESPONSE}" | grep -o '"editable":[^,}]*' | cut -d':' -f2 | tr -d ' ')

echo "   Argo Smart Routing: ${ARGO_STATUS^^}"
echo "   Editable: ${ARGO_EDITABLE}"
echo ""

if [ "${ARGO_STATUS}" = "on" ]; then
  echo "✅ Argo Smart Routing is ENABLED!"
  echo ""
  echo "🚀 Expected Performance Improvements:"
  echo "   - USA users: 50-100ms faster on cache misses"
  echo "   - Brazil users: 20-50ms faster"
  echo "   - Overall latency: 20-40% reduction"
  echo ""
  echo "💰 Cost:"
  echo "   - FREE tier: 1GB/month included"
  echo "   - After free tier: \$5/month per 10GB"
  echo ""
  EXIT_CODE=0
else
  echo "❌ Argo Smart Routing is DISABLED"
  echo ""
  echo "💡 To enable Argo Smart Routing:"
  echo "   1. Go to Cloudflare Dashboard → Network → Argo Smart Routing"
  echo "   2. Toggle 'Argo Smart Routing' to ON"
  echo "   3. Or re-run the cloudflare-setup-job in Kubernetes"
  echo ""
  EXIT_CODE=1
fi

# Additional verification: Test latency with Argo headers
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🌐 Testing Connection to ${DOMAIN}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test the actual domain and check for Argo-related headers
echo "Testing https://${DOMAIN}..."
CURL_OUTPUT=$(curl -s -w "\nHTTP_CODE:%{http_code}\nTIME_TOTAL:%{time_total}\nTIME_CONNECT:%{time_connect}\nTIME_STARTTRANSFER:%{time_starttransfer}" \
  -o /dev/null \
  "https://${DOMAIN}" 2>&1)

HTTP_CODE=$(echo "${CURL_OUTPUT}" | grep "HTTP_CODE:" | cut -d':' -f2)
TIME_TOTAL=$(echo "${CURL_OUTPUT}" | grep "TIME_TOTAL:" | cut -d':' -f2)
TIME_CONNECT=$(echo "${CURL_OUTPUT}" | grep "TIME_CONNECT:" | cut -d':' -f2)
TIME_STARTTRANSFER=$(echo "${CURL_OUTPUT}" | grep "TIME_STARTTRANSFER:" | cut -d':' -f2)

echo "   HTTP Status: ${HTTP_CODE}"
echo "   Connect Time: ${TIME_CONNECT}s"
echo "   Time to First Byte: ${TIME_STARTTRANSFER}s"
echo "   Total Time: ${TIME_TOTAL}s"
echo ""

# Check CF-Cache-Status header
echo "Checking cache headers..."
HEADERS=$(curl -s -I "https://${DOMAIN}" 2>&1)
CF_CACHE=$(echo "${HEADERS}" | grep -i "cf-cache-status" | cut -d':' -f2 | tr -d ' \r\n')
CF_RAY=$(echo "${HEADERS}" | grep -i "cf-ray" | cut -d':' -f2 | tr -d ' \r\n')

echo "   CF-Cache-Status: ${CF_CACHE:-N/A}"
echo "   CF-Ray: ${CF_RAY:-N/A}"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📚 Monitoring Argo Performance"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "To monitor Argo Smart Routing performance:"
echo "   1. Cloudflare Dashboard → Analytics → Performance"
echo "   2. Look for 'Argo' section showing latency improvements"
echo "   3. Check 'Origin Response Time' graphs"
echo ""
echo "Or test from different regions using:"
echo "   curl -w '%{time_total}\n' -o /dev/null -s https://${DOMAIN}"
echo ""

exit ${EXIT_CODE}
