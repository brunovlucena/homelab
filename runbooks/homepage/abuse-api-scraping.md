# 🚨 Runbook: Abuse - API Scraping Pattern

## Alert Information
**Alert Name:** `BrunoSiteAbuseSuspiciousPattern`  
**Severity:** Critical  
**Type:** Security / Abuse

## Symptom
Ratio of API requests to homepage views exceeds 50:1, suggesting automated scraping rather than normal user behavior.

## Diagnosis

```bash
# Check API vs homepage traffic
# In Prometheus:
# sum(rate(http_requests_total{namespace="homepage", path=~"/api/.*"}[1m])) / sum(rate(http_requests_total{namespace="homepage", path="/"}[1m]))

# Identify API endpoints being scraped
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=2000 | grep '/api/' | awk '{print $7}' | sort | uniq -c | sort -rn
```

## Resolution Steps

### 1. Identify Scraper Source

```bash
# Find IPs making excessive API calls
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=5000 | grep '/api/' | awk '{print $1}' | sort | uniq -c | sort -rn | head -10
```

### 2. Implement API Rate Limiting

```bash
# Add rate limiting to API routes
# Example: Limit to 100 req/min per IP
```

### 3. Require API Authentication

Consider implementing API keys or OAuth for API endpoints.

### 4. Block Scraper IPs

## Prevention
- Implement API authentication
- Add rate limiting per endpoint
- Use API keys for public APIs
- Monitor API usage patterns
- Implement CAPTCHA for excessive requests
