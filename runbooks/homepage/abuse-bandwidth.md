# 🚨 Runbook: Abuse - Excessive Bandwidth

## Alert Information
**Alert Name:** `BrunoSiteAbuseExcessiveBandwidth`  
**Severity:** Critical  
**Type:** Security / Abuse

## Symptom
Bruno Site is serving more than 10MB/s of traffic, indicating potential asset downloading/scraping.

## Diagnosis

```bash
# Check which assets are being requested
kubectl logs -n homepage -l app.kubernetes.io/component=frontend --tail=2000 | awk '{print $7, $10}' | sort -k2 -rn | head -20

# Check traffic by source IP
kubectl logs -n homepage -l app.kubernetes.io/component=frontend --tail=5000 | awk '{print $1}' | sort | uniq -c | sort -rn | head -20
```

## Resolution Steps

### 1. Identify what's being downloaded

Check logs for:
- Large file downloads
- Repeated asset requests
- Video/image scraping

### 2. Enable Cloudflare Caching

Ensure all static assets are cached at CDN level.

### 3. Block excessive downloaders

```bash
# Block IPs downloading excessive data
```

### 4. Implement Download Rate Limiting

## Prevention
- Enable Cloudflare caching for all static assets
- Implement download rate limiting
- Use signed URLs for sensitive content
- Monitor bandwidth usage
- Consider adding watermarks to images
