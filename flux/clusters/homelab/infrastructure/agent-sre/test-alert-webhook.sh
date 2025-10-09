#!/bin/bash
# 🧪 Test Alert Webhook Integration
# This script simulates Alertmanager sending alerts to agent-sre

set -e

echo "🧪 Testing Alertmanager → Agent-SRE Webhook Integration"
echo "========================================================"

# Configuration
SERVICE_URL="${SERVICE_URL:-http://192.168.0.16:31081}"
WEBHOOK_PATH="/webhook/alert"
FULL_URL="${SERVICE_URL}${WEBHOOK_PATH}"

echo ""
echo "🎯 Target: ${FULL_URL}"
echo ""

# Test 1: Firing Alert (High Memory Usage)
echo "📤 Test 1: Sending FIRING alert (High Memory Usage)"
echo "---------------------------------------------------"

ALERT_PAYLOAD_FIRING=$(cat <<EOF
{
  "version": "4",
  "groupKey": "{}:{alertname=\"HighMemoryUsage\"}",
  "status": "firing",
  "receiver": "ai-agent-webhook",
  "groupLabels": {
    "alertname": "HighMemoryUsage"
  },
  "commonLabels": {
    "alertname": "HighMemoryUsage",
    "severity": "warning",
    "namespace": "homepage",
    "pod": "bruno-site-api-xyz123"
  },
  "commonAnnotations": {
    "summary": "High memory usage detected on bruno-site API",
    "description": "Pod bruno-site-api-xyz123 is using 85% memory (1.7Gi of 2Gi)",
    "runbook": "https://runbooks.bruno.dev/memory-pressure"
  },
  "externalURL": "http://alertmanager.homelab.local:9093",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "HighMemoryUsage",
        "severity": "warning",
        "namespace": "homepage",
        "pod": "bruno-site-api-xyz123",
        "container": "api",
        "job": "bruno-site-api"
      },
      "annotations": {
        "summary": "High memory usage detected on bruno-site API",
        "description": "Pod bruno-site-api-xyz123 is using 85% memory (1.7Gi of 2Gi). This may lead to OOM kills.",
        "runbook": "https://runbooks.bruno.dev/memory-pressure",
        "dashboard": "http://grafana.homelab.local/d/bruno-site"
      },
      "startsAt": "$(date -u +%Y-%m-%dT%H:%M:%S.000Z)",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://prometheus.homelab.local:9090/graph?g0.expr=container_memory_usage_bytes%7Bpod%3D%22bruno-site-api-xyz123%22%7D+%2F+container_spec_memory_limit_bytes%7Bpod%3D%22bruno-site-api-xyz123%22%7D+%3E+0.85",
      "fingerprint": "abc123def456"
    }
  ]
}
EOF
)

echo ""
echo "📋 Payload:"
echo "$ALERT_PAYLOAD_FIRING" | jq '.'
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${FULL_URL}" \
  -H "Content-Type: application/json" \
  -d "$ALERT_PAYLOAD_FIRING")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "📥 Response (HTTP ${HTTP_CODE}):"
echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
echo ""

if [ "$HTTP_CODE" == "200" ]; then
  echo "✅ Test 1 PASSED: Alert webhook accepted and processed"
else
  echo "❌ Test 1 FAILED: Expected HTTP 200, got ${HTTP_CODE}"
  exit 1
fi

echo ""
echo "=================================================="
echo ""

# Test 2: Critical Alert (Pod Crashlooping)
echo "📤 Test 2: Sending CRITICAL alert (Pod CrashLooping)"
echo "----------------------------------------------------"

ALERT_PAYLOAD_CRITICAL=$(cat <<EOF
{
  "version": "4",
  "groupKey": "{}:{alertname=\"PodCrashLooping\"}",
  "status": "firing",
  "receiver": "ai-agent-webhook",
  "groupLabels": {
    "alertname": "PodCrashLooping"
  },
  "commonLabels": {
    "alertname": "PodCrashLooping",
    "severity": "critical",
    "namespace": "homepage",
    "pod": "bruno-site-postgres-0"
  },
  "commonAnnotations": {
    "summary": "Pod is crash looping",
    "description": "Pod bruno-site-postgres-0 has restarted 5 times in the last 10 minutes"
  },
  "externalURL": "http://alertmanager.homelab.local:9093",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "PodCrashLooping",
        "severity": "critical",
        "namespace": "homepage",
        "pod": "bruno-site-postgres-0",
        "container": "postgres",
        "job": "bruno-site-postgres"
      },
      "annotations": {
        "summary": "Pod is crash looping",
        "description": "Pod bruno-site-postgres-0 has restarted 5 times in the last 10 minutes. Check logs immediately.",
        "impact": "Database unavailable, API requests failing",
        "runbook": "https://runbooks.bruno.dev/crashloop"
      },
      "startsAt": "$(date -u +%Y-%m-%dT%H:%M:%S.000Z)",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://prometheus.homelab.local:9090/graph?g0.expr=rate%28kube_pod_container_status_restarts_total%7Bpod%3D%22bruno-site-postgres-0%22%7D%5B10m%5D%29+%3E+0",
      "fingerprint": "xyz789abc321"
    }
  ]
}
EOF
)

echo ""
echo "📋 Payload:"
echo "$ALERT_PAYLOAD_CRITICAL" | jq '.'
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${FULL_URL}" \
  -H "Content-Type: application/json" \
  -d "$ALERT_PAYLOAD_CRITICAL")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "📥 Response (HTTP ${HTTP_CODE}):"
echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
echo ""

if [ "$HTTP_CODE" == "200" ]; then
  echo "✅ Test 2 PASSED: Critical alert webhook accepted and processed"
else
  echo "❌ Test 2 FAILED: Expected HTTP 200, got ${HTTP_CODE}"
  exit 1
fi

echo ""
echo "=================================================="
echo ""

# Test 3: Resolved Alert
echo "📤 Test 3: Sending RESOLVED alert"
echo "----------------------------------"

ALERT_PAYLOAD_RESOLVED=$(cat <<EOF
{
  "version": "4",
  "groupKey": "{}:{alertname=\"HighMemoryUsage\"}",
  "status": "resolved",
  "receiver": "ai-agent-webhook",
  "groupLabels": {
    "alertname": "HighMemoryUsage"
  },
  "commonLabels": {
    "alertname": "HighMemoryUsage",
    "severity": "warning",
    "namespace": "homepage",
    "pod": "bruno-site-api-xyz123"
  },
  "commonAnnotations": {
    "summary": "High memory usage resolved",
    "description": "Memory usage has returned to normal levels"
  },
  "externalURL": "http://alertmanager.homelab.local:9093",
  "alerts": [
    {
      "status": "resolved",
      "labels": {
        "alertname": "HighMemoryUsage",
        "severity": "warning",
        "namespace": "homepage",
        "pod": "bruno-site-api-xyz123"
      },
      "annotations": {
        "summary": "High memory usage resolved",
        "description": "Memory usage has returned to normal levels (65% of 2Gi)"
      },
      "startsAt": "$(date -u -d '30 minutes ago' +%Y-%m-%dT%H:%M:%S.000Z)",
      "endsAt": "$(date -u +%Y-%m-%dT%H:%M:%S.000Z)",
      "generatorURL": "http://prometheus.homelab.local:9090/graph",
      "fingerprint": "abc123def456"
    }
  ]
}
EOF
)

echo ""
echo "📋 Payload:"
echo "$ALERT_PAYLOAD_RESOLVED" | jq '.'
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${FULL_URL}" \
  -H "Content-Type: application/json" \
  -d "$ALERT_PAYLOAD_RESOLVED")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "📥 Response (HTTP ${HTTP_CODE}):"
echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
echo ""

if [ "$HTTP_CODE" == "200" ]; then
  echo "✅ Test 3 PASSED: Resolved alert webhook accepted"
else
  echo "❌ Test 3 FAILED: Expected HTTP 200, got ${HTTP_CODE}"
fi

echo ""
echo "=================================================="
echo ""
echo "🎉 All webhook tests completed!"
echo ""
echo "📊 Next Steps:"
echo "1. Check agent-sre logs: kubectl logs -n agent-sre -l app=sre-agent -f"
echo "2. Verify LLM analysis was generated"
echo "3. Check for recommendations in the response"
echo ""
echo "💡 To test with real Alertmanager:"
echo "   kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-alertmanager 9093:9093"
echo "   Then trigger a real alert via Prometheus rules"

