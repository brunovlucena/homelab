# 🛡️ Security Quick Start (30 Minutes)
## Immediate Security Hardening for lucena.cloud

This guide implements critical security fixes in 30 minutes.

---

## ✅ Step 1: Update nginx Config (10 minutes)

### Option A: Replace ConfigMap (Recommended)

```bash
cd flux/apps/homepage/k8s/kustomize/base

# Backup current config
cp frontend-nginx-configmap.yaml frontend-nginx-configmap.yaml.backup

# Replace with secure config
# Copy the secure config from the secure template
# The secure config is in: flux/apps/homepage/k8s/kustomize/base/frontend-nginx-configmap-secure.yaml
cp frontend-nginx-configmap-secure.yaml frontend-nginx-configmap.yaml

# Apply changes
kubectl apply -f frontend-nginx-configmap.yaml
kubectl rollout restart deployment/homepage-frontend -n homepage
```

### Option B: Manual Edit

Edit `k8s/kustomize/base/frontend-nginx-configmap.yaml` and add security headers (see `SECURITY_HARDENING_GUIDE.md` for full config).

---

## ✅ Step 2: Enable Cloudflare Security (5 minutes)

1. **Go to**: [Cloudflare Dashboard](https://dash.cloudflare.com) → Select `lucena.cloud`

2. **Security Level**:
   - Go to: **Security** → **Settings**
   - Set **Security Level** to **"Medium"** or **"High"**

3. **Bot Fight Mode**:
   - Go to: **Security** → **Bots**
   - Enable **"Bot Fight Mode"** (FREE)

4. **Always Use HTTPS**:
   - Go to: **SSL/TLS** → **Edge Certificates**
   - Enable **"Always Use HTTPS"**
   - Set **"Minimum TLS Version"** to **1.2**

---

## ✅ Step 3: Add Firewall Rules (10 minutes)

**Go to**: **Security** → **WAF** → **Tools** → **Firewall Rules**

### Rule 1: Block SQL Injection Attempts

**Name**: `Block SQL Injection`
**Expression**:
```
(http.request.uri.query contains "union select") or (http.request.uri.query contains "drop table") or (http.request.uri.query contains "information_schema")
```
**Action**: **Block**

### Rule 2: Block XSS Attempts

**Name**: `Block XSS`
**Expression**:
```
(http.request.uri.query contains "<script") or (http.request.uri.query contains "javascript:") or (http.request.uri.query contains "onload=")
```
**Action**: **Block**

### Rule 3: Rate Limit API Chat Endpoint

**Name**: `Rate Limit Chat API`
**Expression**:
```
http.request.uri.path contains "/api/chat"
```
**Action**: **Challenge** → **After 5 requests per minute**

---

## ✅ Step 4: Add CSP to HTML (5 minutes)

Edit `src/frontend/index.html` and add to `<head>`:

```html
<!-- Content Security Policy -->
<meta http-equiv="Content-Security-Policy" content="default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://www.googletagmanager.com https://www.google-analytics.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com data:; img-src 'self' data: https: blob:; connect-src 'self' https://www.google-analytics.com https://storage.googleapis.com https://*.aliyuncs.com; frame-ancestors 'none'; base-uri 'self'; form-action 'self';">
```

Then rebuild and deploy:
```bash
cd flux/apps/homepage/src/frontend
npm run build
# Deploy as usual
```

---

## 🧪 Step 5: Test Security (5 minutes)

### Test Security Headers

```bash
curl -I https://lucena.cloud | grep -i "x-frame\|x-content\|strict-transport\|content-security"
```

**Expected**:
```
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
Content-Security-Policy: default-src 'self'; ...
```

### Test Rate Limiting

```bash
# Should see 429 after 10 requests
for i in {1..15}; do
  curl -I https://lucena.cloud/api/projects
  sleep 0.1
done
```

### Test Attack Blocking

```bash
# Should be blocked (403)
curl "https://lucena.cloud/?q=union%20select"
curl "https://lucena.cloud/?q=<script>alert(1)</script>"
```

---

## ✅ Verification Checklist

- [ ] Security headers present in response
- [ ] Rate limiting working (429 after limit)
- [ ] SQL injection attempts blocked (403)
- [ ] XSS attempts blocked (403)
- [ ] CORS restricted to lucena.cloud
- [ ] Cloudflare security features enabled
- [ ] Firewall rules active

---

## 🚨 Troubleshooting

### Headers Not Appearing

1. Check nginx config is applied:
   ```bash
   kubectl get configmap homepage-frontend-nginx -n homepage -o yaml
   ```

2. Restart nginx pod:
   ```bash
   kubectl rollout restart deployment/homepage-frontend -n homepage
   ```

3. Check nginx logs:
   ```bash
   kubectl logs -n homepage -l app=homepage-frontend --tail=50
   ```

### Rate Limiting Too Aggressive

Adjust limits in nginx config:
```nginx
limit_req_zone $binary_remote_addr zone=general:10m rate=10r/s;  # Change rate
```

### CSP Breaking Site

Adjust CSP in nginx config or HTML meta tag. Start with report-only mode:
```nginx
add_header Content-Security-Policy-Report-Only "..." always;
```

---

## 📊 What You've Secured

✅ **Clickjacking Protection** (X-Frame-Options)  
✅ **MIME Sniffing Protection** (X-Content-Type-Options)  
✅ **XSS Protection** (X-XSS-Protection + CSP)  
✅ **HTTPS Enforcement** (HSTS)  
✅ **Rate Limiting** (DoS protection)  
✅ **SQL Injection Blocking** (nginx + Cloudflare)  
✅ **XSS Blocking** (nginx + Cloudflare)  
✅ **CORS Restriction** (no wildcard)  
✅ **Request Size Limits** (DoS protection)  
✅ **Bot Protection** (Cloudflare)  

---

**Total Time**: 30 minutes  
**Cost**: $0 (all FREE features)  
**Security Improvement**: 🔴 Critical gaps → ✅ Hardened

---

**Next Steps**: See `SECURITY_HARDENING_GUIDE.md` for advanced security features.
