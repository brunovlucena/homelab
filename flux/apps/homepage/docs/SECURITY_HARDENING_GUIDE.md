# 🛡️ Security Hardening Guide for lucena.cloud
## Senior Security Engineer Review & Recommendations

**Reviewer**: Senior Security Engineer  
**Date**: 2025-01-XX  
**Scope**: Homepage security hardening (Frontend + API)  
**Current Setup**: Cloudflare Tunnel + nginx + Go API

---

## 🔍 Current Security Assessment

### ✅ What's Already Good

1. **API Security** (Go Backend):
   - ✅ Rate limiting implemented (100 req/min, 20 req/min for chat)
   - ✅ Security headers middleware
   - ✅ CORS configured (restrictive)
   - ✅ Request ID tracking
   - ✅ Error handling with recovery
   - ✅ Input validation patterns

2. **Infrastructure**:
   - ✅ Cloudflare Tunnel (DDoS protection)
   - ✅ HTTPS enforced (Cloudflare)
   - ✅ Kubernetes security contexts

### ❌ Critical Security Gaps

1. **nginx Security Headers Missing** 🔴 CRITICAL
   - No HSTS header
   - No X-Frame-Options
   - No Content-Security-Policy
   - No X-Content-Type-Options
   - No Referrer-Policy

2. **CORS Too Permissive** 🔴 CRITICAL
   - `Access-Control-Allow-Origin: *` in nginx
   - Allows any origin to access resources

3. **No Request Size Limits** ⚠️ HIGH
   - No `client_max_body_size` in nginx
   - Vulnerable to DoS via large uploads

4. **No Rate Limiting at nginx Level** ⚠️ HIGH
   - Only API has rate limiting
   - Static assets can be abused

5. **No Cloudflare WAF** ⚠️ MEDIUM
   - Free tier doesn't include WAF
   - No SQL injection/XSS protection at edge

6. **Missing Security Headers in Frontend** ⚠️ MEDIUM
   - No CSP in HTML
   - No security meta tags

7. **No IP Blocking/Whitelisting** ⚠️ MEDIUM
   - No geo-blocking
   - No IP reputation filtering

---

## 🚀 Security Hardening Implementation

### Phase 1: Critical Fixes (Do Today - 1 Hour)

#### 1.1 Add Security Headers to nginx

**File**: `k8s/kustomize/base/frontend-nginx-configmap.yaml`

```nginx
server {
    listen       8080;
    listen  [::]:8080;
    server_name  localhost;

    # ============================================================================
    # 🔒 SECURITY HEADERS
    # ============================================================================
    
    # Prevent clickjacking
    add_header X-Frame-Options "DENY" always;
    
    # Prevent MIME type sniffing
    add_header X-Content-Type-Options "nosniff" always;
    
    # Enable XSS protection (legacy browsers)
    add_header X-XSS-Protection "1; mode=block" always;
    
    # Referrer policy
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    
    # Permissions policy (restrict browser features)
    add_header Permissions-Policy "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()" always;
    
    # Content Security Policy (adjust based on your needs)
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://www.googletagmanager.com https://www.google-analytics.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com data:; img-src 'self' data: https: blob:; connect-src 'self' https://www.google-analytics.com https://storage.googleapis.com https://*.aliyuncs.com; frame-ancestors 'none'; base-uri 'self'; form-action 'self';" always;
    
    # HSTS (if using HTTPS directly, Cloudflare handles this but good to have)
    # Note: Cloudflare already adds HSTS, but adding here for defense in depth
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    
    # Remove server version disclosure
    server_tokens off;
    
    # Hide nginx version
    more_clear_headers Server;
    
    # ============================================================================
    # 🔒 REQUEST LIMITS
    # ============================================================================
    
    # Limit request body size (prevent DoS)
    client_max_body_size 10M;
    
    # Limit buffer sizes
    client_body_buffer_size 128k;
    client_header_buffer_size 1k;
    large_client_header_buffers 4 16k;
    
    # Timeouts
    client_body_timeout 10s;
    client_header_timeout 10s;
    send_timeout 10s;
    keepalive_timeout 65s;
    
    # ============================================================================
    # 🔒 RATE LIMITING
    # ============================================================================
    
    # Define rate limit zones
    limit_req_zone $binary_remote_addr zone=general:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=api:10m rate=5r/s;
    limit_req_zone $binary_remote_addr zone=static:10m rate=50r/s;
    
    # Connection limiting
    limit_conn_zone $binary_remote_addr zone=conn_limit_per_ip:10m;
    limit_conn conn_limit_per_ip 20;
    
    # Health check endpoint
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }

    location / {
        root   /usr/share/nginx/html;
        index  index.html index.htm;
        try_files $uri $uri/ /index.html;
        
        # Apply rate limiting
        limit_req zone=general burst=20 nodelay;
        limit_conn conn_limit_per_ip 10;
        
        # Cache static assets
        location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg|webp|woff|woff2|ttf|eot)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
            
            # Higher rate limit for static assets
            limit_req zone=static burst=100 nodelay;
        }
    }

    # Handle API requests
    location /api {
        # Stricter rate limiting for API
        limit_req zone=api burst=10 nodelay;
        limit_conn conn_limit_per_ip 5;
        
        proxy_pass http://homepage-api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Security: Don't pass client body to upstream if too large
        proxy_request_buffering on;
        proxy_max_temp_file_size 0;
        
        # Proxy buffer limits
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
        proxy_busy_buffers_size 8k;
        
        # Timeouts
        proxy_connect_timeout 5s;
        proxy_send_timeout 10s;
        proxy_read_timeout 10s;
    }

    # Proxy MinIO storage requests
    location /storage/ {
        # Rate limiting
        limit_req zone=static burst=50 nodelay;
        
        proxy_pass http://minio.minio.svc.cluster.local:9000/;
        proxy_set_header Host minio.minio.svc.cluster.local;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 🔒 FIX CORS - Restrict to your domain
        add_header Access-Control-Allow-Origin "https://lucena.cloud" always;
        add_header Access-Control-Allow-Methods "GET, OPTIONS" always;
        add_header Access-Control-Allow-Headers "Origin, X-Requested-With, Content-Type, Accept" always;
        add_header Access-Control-Max-Age "3600" always;
        
        # Handle OPTIONS preflight
        if ($request_method = 'OPTIONS') {
            add_header Access-Control-Allow-Origin "https://lucena.cloud" always;
            add_header Access-Control-Allow-Methods "GET, OPTIONS" always;
            add_header Access-Control-Allow-Headers "Origin, X-Requested-With, Content-Type, Accept" always;
            add_header Access-Control-Max-Age "3600" always;
            add_header Content-Length 0;
            add_header Content-Type text/plain;
            return 204;
        }
        
        # Buffer settings
        proxy_buffering on;
        proxy_buffer_size 128k;
        proxy_buffers 8 128k;
        proxy_busy_buffers_size 256k;
        proxy_max_temp_file_size 0;
        
        # Timeouts
        proxy_connect_timeout 300s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
        
        # Cache control
        proxy_cache off;
        add_header Cache-Control "public, max-age=300, must-revalidate";
        add_header X-Cache-Status "bypass";
    }
    
    # ============================================================================
    # 🔒 BLOCK COMMON ATTACK PATTERNS
    # ============================================================================
    
    # Block access to hidden files
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }
    
    # Block access to backup files
    location ~ ~$ {
        deny all;
        access_log off;
        log_not_found off;
    }
    
    # Block common exploit attempts
    location ~* (\.env|\.git|\.svn|\.htaccess|\.htpasswd|wp-config\.php|\.DS_Store) {
        deny all;
        access_log off;
        log_not_found off;
    }
    
    # Block SQL injection attempts in query strings
    if ($query_string ~* "union.*select|insert.*into|delete.*from|drop.*table|exec\(|information_schema") {
        return 403;
    }
    
    # Block XSS attempts
    if ($query_string ~* "<script|javascript:|onload=|onerror=") {
        return 403;
    }
    
    # Block path traversal
    if ($uri ~* "\.\./|\.\.\\|\.\.%2f|\.\.%5c") {
        return 403;
    }
    
    # Error pages
    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }
}
```

---

#### 1.2 Fix CORS in nginx

**Change from**:
```nginx
add_header Access-Control-Allow-Origin * always;
```

**Change to**:
```nginx
add_header Access-Control-Allow-Origin "https://lucena.cloud" always;
```

---

#### 1.3 Add CSP to Frontend HTML

**File**: `src/frontend/index.html`

Add to `<head>`:
```html
<!-- Content Security Policy -->
<meta http-equiv="Content-Security-Policy" content="default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://www.googletagmanager.com https://www.google-analytics.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com data:; img-src 'self' data: https: blob:; connect-src 'self' https://www.google-analytics.com https://storage.googleapis.com https://*.aliyuncs.com; frame-ancestors 'none'; base-uri 'self'; form-action 'self';">
```

---

### Phase 2: Cloudflare Security (This Week)

#### 2.1 Enable Cloudflare Security Features (FREE)

**Go to**: Cloudflare Dashboard → **Security** → **Settings**

1. **Security Level**: Set to **"Medium"** or **"High"**
   - Blocks known bad IPs
   - Challenges suspicious traffic

2. **Bot Fight Mode**: Enable (FREE)
   - Blocks simple bots
   - Reduces automated attacks

3. **Challenge Passage**: Set to **"17 minutes"**
   - Balance between security and UX

4. **Browser Integrity Check**: Enable
   - Validates browser headers
   - Blocks automated tools

---

#### 2.2 Configure Firewall Rules (FREE Tier)

**Go to**: Cloudflare Dashboard → **Security** → **WAF** → **Tools** → **Firewall Rules**

**Rule 1: Block Known Bad IPs**
```
(http.request.uri.path contains "/api" and ip.geoip.country in {"CN" "RU" "KP"}) 
or 
(cf.threat_score gt 50)
```
**Action**: Block

**Rule 2: Rate Limit API Endpoints**
```
http.request.uri.path contains "/api/chat"
```
**Action**: Challenge after 5 requests per minute

**Rule 3: Block Common Attack Patterns**
```
(http.request.uri.query contains "union select") 
or 
(http.request.uri.query contains "<script") 
or 
(http.request.uri.query contains "../")
```
**Action**: Block

---

#### 2.3 Enable Always Use HTTPS

**Go to**: **SSL/TLS** → **Edge Certificates**
- ✅ **Always Use HTTPS**: ON
- ✅ **Automatic HTTPS Rewrites**: ON
- ✅ **Minimum TLS Version**: 1.2
- ✅ **Opportunistic Encryption**: ON

---

#### 2.4 Configure Page Rules for Security

**Go to**: **Rules** → **Page Rules**

**Rule 1: Security Headers for API**
```
URL: lucena.cloud/api/*
Settings:
  - Security Level: High
  - Browser Integrity Check: ON
  - Cache Level: Bypass
```

**Rule 2: Security Headers for Static Assets**
```
URL: lucena.cloud/assets/*
Settings:
  - Security Level: Medium
  - Browser Integrity Check: ON
  - Cache Level: Cache Everything
```

---

### Phase 3: Advanced Security (This Month)

#### 3.1 Implement IP Reputation Filtering

**Option A: Cloudflare Access Rules (FREE)**
- Go to: **Security** → **WAF** → **Tools** → **IP Access Rules**
- Block known bad IP ranges
- Allow only trusted IPs for admin endpoints

**Option B: nginx GeoIP Module**
```nginx
# Block specific countries (if needed)
geoip_country /usr/share/GeoIP/GeoIP.dat;
map $geoip_country_code $blocked_country {
    default 0;
    CN 1;  # Example: Block China (adjust as needed)
    RU 1;
}

# Use in location block
if ($blocked_country) {
    return 403;
}
```

---

#### 3.2 Add Request Logging for Security Events

**Update nginx config**:
```nginx
# Log security events
log_format security '$remote_addr - $remote_user [$time_local] '
                   '"$request" $status $body_bytes_sent '
                   '"$http_referer" "$http_user_agent" '
                   '$request_time $upstream_response_time '
                   'threat_score=$http_cf_threat_score';

# Log blocked requests
access_log /var/log/nginx/security.log security if=$blocked;
```

---

#### 3.3 Implement Fail2Ban (Optional)

Monitor nginx logs and automatically block IPs with too many 403/401 errors.

---

## 📊 Security Checklist

### ✅ Immediate (Today)
- [ ] Add security headers to nginx
- [ ] Fix CORS (remove wildcard)
- [ ] Add request size limits
- [ ] Add rate limiting to nginx
- [ ] Add CSP to HTML
- [ ] Block common attack patterns

### ✅ This Week
- [ ] Enable Cloudflare security features
- [ ] Configure firewall rules
- [ ] Enable Always Use HTTPS
- [ ] Configure Page Rules for security
- [ ] Test security headers

### ✅ This Month
- [ ] Implement IP reputation filtering
- [ ] Add security event logging
- [ ] Set up monitoring/alerts
- [ ] Review and tighten CSP
- [ ] Consider Cloudflare WAF (paid)

---

## 🧪 Testing Security

### Test Security Headers

```bash
# Test headers
curl -I https://lucena.cloud | grep -i "x-frame\|x-content\|strict-transport\|content-security"

# Expected output:
# X-Frame-Options: DENY
# X-Content-Type-Options: nosniff
# Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
# Content-Security-Policy: default-src 'self'; ...
```

### Test Rate Limiting

```bash
# Should be rate limited after 10 requests
for i in {1..15}; do
  curl -I https://lucena.cloud/api/projects
  sleep 0.1
done
# Should see 429 Too Many Requests
```

### Test CORS

```bash
# Should fail (CORS error)
curl -H "Origin: https://evil.com" https://lucena.cloud/storage/test.jpg

# Should succeed
curl -H "Origin: https://lucena.cloud" https://lucena.cloud/storage/test.jpg
```

### Test Attack Patterns

```bash
# Should be blocked (403)
curl "https://lucena.cloud/?q=union%20select"
curl "https://lucena.cloud/?q=<script>alert(1)</script>"
curl "https://lucena.cloud/../../../etc/passwd"
```

---

## 📈 Security Metrics to Monitor

1. **Rate Limit Hits**: Track 429 responses
2. **Blocked Requests**: Track 403 responses
3. **Threat Score**: Monitor Cloudflare threat scores
4. **Failed Login Attempts**: Track API auth failures
5. **Large Requests**: Monitor request sizes
6. **CORS Violations**: Track CORS errors

---

## 🚨 Incident Response

### If Under Attack

1. **Immediate**:
   - Enable "I'm Under Attack" mode in Cloudflare
   - Block attacking IPs in Cloudflare firewall
   - Increase rate limits temporarily

2. **Short-term**:
   - Review Cloudflare logs
   - Identify attack pattern
   - Update firewall rules

3. **Long-term**:
   - Consider Cloudflare WAF (paid)
   - Implement additional rate limiting
   - Add DDoS protection

---

## 💰 Cost Analysis

| Security Feature | Cost | Priority |
|------------------|------|----------|
| **nginx Security Headers** | FREE | 🔴 Critical |
| **Cloudflare Security (Free)** | FREE | 🔴 Critical |
| **Cloudflare Firewall Rules** | FREE | 🔴 Critical |
| **Rate Limiting (nginx)** | FREE | 🔴 Critical |
| **Cloudflare WAF** | $20+/month | ⚠️ Optional |
| **IP Reputation Service** | $0-50/month | ⚠️ Optional |

**Total Cost for Basic Security**: **$0/month** ✅

---

## 📚 References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Cloudflare Security Best Practices](https://developers.cloudflare.com/fundamentals/get-started/concepts/cloudflare-challenges/)
- [nginx Security Headers](https://nginx.org/en/docs/http/ngx_http_headers_module.html)
- [Content Security Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)

---

**Next Steps**: Implement Phase 1 (Critical Fixes) today, then proceed with Phase 2 this week.
