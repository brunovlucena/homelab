# 🔒 Homepage Security Monitoring Guide

**Reference**: BVL-193 (SEC-013)  
**Last Updated**: January 2026  
**Owner**: SRE Team

---

## Overview

This document describes the comprehensive security monitoring and alerting system implemented for the Homepage service (lucena.cloud). The system provides:

- Real-time security event detection
- Multi-layer attack detection (nginx, Cloudflare edge)
- Security-focused Grafana dashboard
- Automated alerting for suspicious activity
- Security event correlation

---

## Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                          Security Monitoring Stack                        │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                  │
│  │  Cloudflare │───▶│    nginx    │───▶│  Homepage   │                  │
│  │    Edge     │    │  Frontend   │    │    API      │                  │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘                  │
│         │                  │                  │                          │
│         ▼                  ▼                  ▼                          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                  │
│  │ CF Exporter │    │   nginx     │    │   API       │                  │
│  │  Metrics    │    │  Metrics    │    │  Metrics    │                  │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘                  │
│         │                  │                  │                          │
│         └────────────┬─────┴──────────────────┘                          │
│                      ▼                                                   │
│              ┌─────────────┐                                             │
│              │ Prometheus  │                                             │
│              │   Server    │                                             │
│              └──────┬──────┘                                             │
│                     │                                                    │
│         ┌───────────┼───────────┐                                        │
│         ▼           ▼           ▼                                        │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐                        │
│  │ Prometheus  │ │   Grafana   │ │ Alertmanager│                        │
│  │   Rules     │ │  Security   │ │   Alerts    │                        │
│  │             │ │  Dashboard  │ │             │                        │
│  └─────────────┘ └─────────────┘ └─────────────┘                        │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Components

### 1. PrometheusRules (`homepage-security.yaml`)

Location: `flux/infrastructure/prometheus-operator/k8s/prometheusrules/homepage-security.yaml`

#### Alert Categories

| Category | Severity | Description |
|----------|----------|-------------|
| `homepage.security.critical` | Critical | Immediate response required - active attacks |
| `homepage.security.high` | Warning | Urgent attention needed - suspicious patterns |
| `homepage.security.medium` | Info | Monitoring required - anomalies detected |
| `homepage.security.cloudflare` | Warning | Edge security events |
| `homepage.security.correlation` | Critical/Warning | Multi-signal attack detection |

#### Critical Alerts

| Alert | Trigger | Response |
|-------|---------|----------|
| `HomepageBruteForceAttack` | >1 req/s 401 responses | Enable Cloudflare "I'm Under Attack" mode |
| `HomepageDDoSAttackDetected` | >100 req/s total | Review rate limits, enable edge protection |
| `HomepageMassRateLimitViolations` | >5 req/s 429 responses | Identify source IPs, block malicious actors |
| `HomepageSQLInjectionAttack` | >10 blocked injection attempts | Review WAF rules, report to security |
| `HomepagePathTraversalAttack` | >5 path traversal attempts | Verify file access rules, check for exfiltration |
| `HomepageCoordinatedAttack` | Multiple attack vectors | Full incident response |

### 2. Grafana Dashboard (`homepage-security`)

Location: `flux/infrastructure/prometheus-operator/k8s/dashboards/homepage-security-dashboard-configmap.yaml`

#### Dashboard Sections

1. **Security Overview**
   - Security status indicator (SECURE/ALERT)
   - Active alerts count
   - Blocked requests rate
   - Rate limited requests rate
   - Auth failures rate
   - Error rate percentage

2. **Security Events Timeline**
   - 403 Forbidden responses
   - 429 Rate Limited responses
   - 401 Unauthorized responses
   - 400 Bad Request responses

3. **Attack Detection**
   - WordPress/PHP probes
   - Config file probes (.env, .git)
   - SQL Injection attempts
   - XSS attempts
   - Path traversal attempts

4. **Rate Limiting & Access Control**
   - Request rate vs rate limited comparison
   - API endpoint request rates

5. **Cloudflare Edge Security**
   - Edge vs Origin traffic comparison
   - Edge protection rate (cache/filter percentage)
   - WAF blocked requests
   - Bot blocked requests

6. **Active Alerts Table**
   - Real-time view of firing alerts
   - Severity color coding

---

## Security Metrics

### nginx Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `nginx_http_requests_total{status="403"}` | Blocked requests | >0.5/s sustained |
| `nginx_http_requests_total{status="429"}` | Rate limited requests | >0.5/s sustained |
| `nginx_http_requests_total{status="401"}` | Auth failures | >0.1/s sustained |
| `nginx_http_requests_total{status="400"}` | Bad requests | >1/s sustained |

### Cloudflare Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `cloudflare_zone_requests_total{waf_action="block"}` | WAF blocks | >0.5/s sustained |
| `cloudflare_zone_requests_total{bot_management_decision="block"}` | Bot blocks | >1/s sustained |
| `cloudflare_zone_requests_total{threat_score="high"}` | High threat score | >0.1/s sustained |

### Recording Rules

Pre-computed metrics for efficient querying:

```promql
# Blocked requests rate
homepage:security:blocked_requests:rate5m

# Rate limited requests rate  
homepage:security:rate_limited_requests:rate5m

# Auth failures rate
homepage:security:auth_failures:rate5m

# Total security events rate
homepage:security:events_total:rate5m

# API attack surface
homepage:security:api_attack_surface:rate5m
```

---

## Incident Response

### Level 1: Automated Response

These actions are triggered automatically by nginx and Cloudflare:

1. **Rate Limiting** - nginx blocks IPs exceeding rate limits
2. **Attack Pattern Blocking** - nginx returns 403 for SQL injection/XSS attempts
3. **Bot Fight Mode** - Cloudflare challenges suspicious bots
4. **WAF Rules** - Cloudflare blocks known attack patterns

### Level 2: Manual Response (Warning Alerts)

When a warning alert fires:

1. Check the Security Dashboard for patterns
2. Review logs in Loki: `{namespace="homepage"} |= "403" or |= "429"`
3. Identify source IPs
4. If legitimate traffic: adjust rate limits
5. If attack: block IPs in Cloudflare

### Level 3: Emergency Response (Critical Alerts)

When a critical alert fires:

1. **Immediate**: Enable "I'm Under Attack" mode in Cloudflare
   ```bash
   # Via Cloudflare Dashboard or API
   curl -X PATCH "https://api.cloudflare.com/client/v4/zones/{zone_id}/settings/security_level" \
     -H "Authorization: Bearer $CF_API_TOKEN" \
     -d '{"value":"under_attack"}'
   ```

2. **Short-term**: Review and block attacking IPs
   ```bash
   # View top attacking IPs
   kubectl logs -n homepage -l app=homepage-frontend --tail=1000 | \
     grep "403\|429" | awk '{print $1}' | sort | uniq -c | sort -rn | head -20
   ```

3. **Long-term**: Review and update security rules

---

## Runbooks

### Brute Force Attack

**File**: `docs/runbooks/brute-force-attack.md`

**Symptoms**:
- High rate of 401 responses
- `HomepageBruteForceAttack` alert firing

**Response**:
1. Enable Cloudflare challenge mode for `/api/` endpoints
2. Review authentication logs
3. Consider implementing account lockout
4. Block offending IPs

### DDoS Attack

**File**: `docs/runbooks/ddos-attack.md`

**Symptoms**:
- Request rate >100 req/s
- `HomepageDDoSAttackDetected` alert firing

**Response**:
1. Enable "I'm Under Attack" mode
2. Review Cloudflare analytics for attack pattern
3. Create firewall rules for attack signature
4. Scale up origin if needed

### Coordinated Attack

**File**: `docs/runbooks/coordinated-attack.md`

**Symptoms**:
- Multiple attack types simultaneously
- `HomepageCoordinatedAttack` alert firing

**Response**:
1. Immediately enable "I'm Under Attack" mode
2. Notify security team
3. Gather evidence from all logs
4. Consider temporary service restriction

---

## Configuration

### Alert Routing

Security alerts are routed via Alertmanager:

```yaml
# Critical alerts -> PagerDuty immediate
# Warning alerts -> Slack #security-alerts
# Info alerts -> Email digest
```

### Dashboard Access

The Security Dashboard is accessible at:
- URL: `/d/homepage-security`
- Folder: Homepage
- Tags: `homepage`, `security`, `monitoring`, `sec-013`

---

## Testing

### Test Rate Limiting

```bash
# Should trigger rate limit after 10 requests
for i in {1..15}; do
  curl -I https://lucena.cloud/
  sleep 0.05
done
# Should see 429 Too Many Requests
```

### Test Attack Pattern Blocking

```bash
# SQL Injection (should return 403)
curl -I "https://lucena.cloud/?q=union%20select"

# XSS (should return 403)
curl -I "https://lucena.cloud/?q=<script>alert(1)</script>"

# Path Traversal (should return 403)
curl -I "https://lucena.cloud/../../../etc/passwd"
```

### Verify Alerts

```bash
# Check PrometheusRule is loaded
kubectl get prometheusrules -n prometheus homepage-security-rules

# Check active alerts
kubectl exec -n prometheus prometheus-0 -- wget -qO- localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.labels.alertname | startswith("Homepage"))'
```

---

## Maintenance

### Regular Security Review

**Weekly**:
- Review security dashboard for patterns
- Check for new attack signatures
- Review blocked IPs

**Monthly**:
- Review and tune alert thresholds
- Update WAF rules based on new threats
- Test incident response procedures

**Quarterly**:
- Full security audit
- Penetration testing
- Update runbooks

---

## References

- [SECURITY_HARDENING_GUIDE.md](./SECURITY_HARDENING_GUIDE.md)
- [SECURITY_QUICK_START.md](./SECURITY_QUICK_START.md)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Cloudflare Security Best Practices](https://developers.cloudflare.com/fundamentals/)
- [nginx Security Headers](https://nginx.org/en/docs/http/ngx_http_headers_module.html)

---

## Changelog

| Date | Change | Author |
|------|--------|--------|
| 2026-01 | Initial implementation (SEC-013) | SRE Team |
