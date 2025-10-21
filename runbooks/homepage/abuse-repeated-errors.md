# 🚨 Runbook: Abuse - Repeated 4xx Errors

## Alert Information
**Alert Name:** `BrunoSiteAbuseRepeatedErrors`  
**Severity:** Critical  
**Type:** Security / Abuse

## Symptom
High rate of 4xx errors (>2/sec), indicating scanning, brute force attempts, or malicious probing.

## Diagnosis

```bash
# Check 4xx error patterns
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=1000 | grep ' 4[0-9][0-9] ' | head -50

# Check which endpoints are being targeted
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=1000 | grep ' 4[0-9][0-9] ' | awk '{print $7}' | sort | uniq -c | sort -rn
```

## Resolution Steps

### 1. Identify Attack Type

- **401/403:** Authentication/authorization attempts
- **404:** Path scanning/enumeration
- **400:** Malformed requests/injection attempts

### 2. Block Malicious IPs

```bash
# Identify attacking IPs
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=5000 | grep ' 4[0-9][0-9] ' | awk '{print $1}' | sort | uniq -c | sort -rn | head -20

# Block via Cloudflare or firewall
```

### 3. Enable Cloudflare Challenge Mode

Temporarily require JavaScript challenges for all visitors.

## Prevention
- Enable Cloudflare WAF rules
- Implement fail2ban-like blocking
- Use CAPTCHA for suspicious behavior
- Monitor failed authentication attempts
