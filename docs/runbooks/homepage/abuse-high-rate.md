# 🚨 Runbook: Abuse - High Request Rate

## Alert Information
**Alert Name:** `BrunoSiteAbuseHighRequestRate`  
**Severity:** Critical  
**Type:** Security / Abuse

## Symptom
Bruno Site is receiving more than 10 requests per second, indicating potential DDoS attack, scraping bot, or abuse.

## Impact
- **User Impact:** POTENTIAL - May degrade performance
- **Business Impact:** HIGH - Increased costs, potential service degradation
- **Security Impact:** HIGH - Potential attack

## Diagnosis

```bash
# Check request patterns in logs
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=500 | awk '{print $1}' | sort | uniq -c | sort -rn | head -20

# Check Prometheus for traffic sources
# Query: rate(http_requests_total{namespace="homepage"}[1m])

# Check Grafana dashboards for traffic patterns
```

## Resolution Steps

### Step 1: Identify Source IPs

```bash
# Extract top IPs from logs
kubectl logs -n homepage -l app.kubernetes.io/component=frontend --tail=5000 | grep -oE '[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | sort | uniq -c | sort -rn | head -20
```

### Step 2: Verify if it's legitimate traffic

Check if:
- Known search engine bots (Googlebot, Bingbot)
- Legitimate monitoring services
- Legitimate user traffic spike

### Step 3: Implement Rate Limiting

If malicious:
1. Enable Cloudflare rate limiting
2. Enable Cloudflare "Under Attack" mode
3. Block specific IPs at firewall level
4. Implement application-level rate limiting

```bash
# Example: Block IP in nginx (if using)
kubectl exec -it -n homepage <frontend-pod> -- sh -c 'echo "deny 1.2.3.4;" >> /etc/nginx/conf.d/blocked-ips.conf && nginx -s reload'
```

### Step 4: Scale if needed

```bash
kubectl scale deployment -n homepage homepage-api --replicas=5
kubectl scale deployment -n homepage homepage-frontend --replicas=5
```

## Verification

Monitor traffic returns to normal levels (<2 req/sec for homepage).

## Prevention

1. Implement Cloudflare protection
2. Enable rate limiting at CDN level
3. Use Web Application Firewall (WAF)
4. Implement CAPTCHA for suspicious traffic
5. Monitor traffic patterns regularly
6. Set up automated blocking rules

## Related Alerts
- `BrunoSiteAbuseRepeatedErrors`
- `BrunoSiteAbuseSuspiciousPattern`
- `BrunoSiteAbuseExcessiveBandwidth`

## Escalation

If attack continues:
1. Enable full DDoS protection
2. Contact Cloudflare support
3. Consider temporarily blocking all non-essential traffic
4. Review firewall rules with network team

## Additional Resources
- [Cloudflare DDoS Protection](https://www.cloudflare.com/ddos/)
- [OWASP DDoS Prevention](https://owasp.org/www-community/attacks/Denial_of_Service)
