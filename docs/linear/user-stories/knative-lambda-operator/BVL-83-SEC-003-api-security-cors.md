# SEC-003: API Security & CORS Misconfiguration Testing

**Priority**: P0 | **Status**: 📋 Backlog K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-247/sec-003-api-security-and-cors-misconfiguration-testing

**Priority:** P0 | **Story Points:** 8

## 📋 User Story

**As a** Principal Pentester  
**I want to** validate API security controls and CORS configurations  
**So that** APIs cannot be abused and cross-origin attacks are prevented

## 🎯 Acceptance Criteria

### AC1: CORS Configuration Security
**Given** API endpoints may be accessed from browsers  
**When** attempting cross-origin requests from unauthorized origins  
**Then** requests should be blocked by CORS policy

**Security Tests:**
- ✅ `Access-Control-Allow-Origin: *` is NOT used on sensitive endpoints
- ✅ Only whitelisted origins are allowed
- ✅ Credentials are not allowed with wildcard origins
- ✅ CORS preflight requests properly validated
- ✅ `Access-Control-Allow-Methods` restricted to needed methods
- ✅ `Access-Control-Allow-Headers` whitelisted only

**Attack Scenarios:**
- ❌ CORS bypass via null origin
- ❌ CORS bypass via origin reflection
- ❌ CORS bypass via subdomain wildcards
- ❌ Cross-site request forgery (CSRF) via misconfigured CORS

### AC2: HTTP Security Headers
**Given** API responds to HTTP requests  
**When** examining security headers  
**Then** all security headers should be properly configured

**Required Headers:**
```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

**Security Tests:**
- ✅ All security headers present
- ✅ No sensitive information in headers
- ✅ Server header removed or generic
- ✅ X-Powered-By header removed

### AC3: Rate Limiting Enforcement
**Given** APIs may be subject to abuse  
**When** sending excessive requests  
**Then** rate limits should be enforced

**Security Tests:**
- ✅ Rate limiting active on all endpoints
- ✅ 429 Too Many Requests returned when exceeded
- ✅ Rate limit headers present (`X-RateLimit-*`)
- ✅ Rate limits vary by endpoint criticality
- ✅ Distributed rate limiting (not per-instance)
- ✅ IP-based and token-based limiting

**Expected Limits:**
- Authentication endpoints: 10 req/min
- Build endpoints: 100 req/min
- Read endpoints: 1000 req/min
- Health/metrics: No limit

### AC4: API Versioning Security
**Given** API may support multiple versions  
**When** attempting to access deprecated versions  
**Then** security controls should be consistent

**Security Tests:**
- ✅ Old API versions properly deprecated
- ✅ Security controls apply to all versions
- ✅ Version negotiation cannot be bypassed
- ✅ Deprecated versions return clear warnings
- ✅ Breaking changes documented

### AC5: HTTP Method Security
**Given** different HTTP methods have different security implications  
**When** testing HTTP method handling  
**Then** methods should be properly restricted

**Security Tests:**
- ✅ OPTIONS method returns only allowed methods
- ✅ Dangerous methods (TRACE, CONNECT) disabled
- ✅ HEAD method doesn't leak sensitive data
- ✅ HTTP method override headers blocked (`X-HTTP-Method-Override`)
- ✅ Verb tampering detection active

**Attack Scenarios:**
- ❌ GET endpoint called with POST to bypass restrictions
- ❌ DELETE via PUT with method override
- ❌ OPTIONS revealing admin endpoints

### AC6: API Authentication Enforcement
**Given** API endpoints require authentication  
**When** accessing endpoints without credentials  
**Then** authentication should be enforced

**Security Tests:**
- ✅ All non-public endpoints require authentication
- ✅ 401 Unauthorized for missing credentials
- ✅ Bearer token validation enforced
- ✅ Token expiration checked
- ✅ API keys properly validated

**Public Endpoints (no auth required):**
- `/healthz`
- `/metrics` (internal network only)
- `/readyz`

**Protected Endpoints (auth required):**
- `/api/v1/build`
- `/api/v1/lambda`
- `/api/v1/service`

### AC7: Content Type Validation
**Given** APIs accept various content types  
**When** sending unexpected content types  
**Then** requests should be validated and rejected

**Security Tests:**
- ✅ Content-Type header required
- ✅ Only expected content types accepted
- ✅ Content-Type vs actual content validated
- ✅ Large payloads rejected (e.g., >10MB)
- ✅ Malformed JSON/YAML rejected gracefully

**Attack Scenarios:**
- ❌ Content-Type mismatch (declare JSON, send XML)
- ❌ XXE attack via unexpected XML
- ❌ Polyglot file upload
- ❌ Compression bomb (gzip bomb)

### AC8: Error Handling Security
**Given** API may encounter errors  
**When** triggering error conditions  
**Then** error messages should not leak sensitive information

**Security Tests:**
- ✅ Stack traces not exposed in production
- ✅ Database errors sanitized
- ✅ File paths not revealed
- ✅ Internal IP addresses not exposed
- ✅ Version numbers minimally disclosed
- ✅ Generic error messages for security failures

**Sensitive Information to Avoid:**
- Stack traces
- Database schema
- Internal paths
- Credentials
- API versions
- Server software

## 🔴 Attack Surface Analysis

### API Endpoints

1. **Builder Service**
   ```
   POST   /api/v1/build         - Create build job
   GET    /api/v1/build/:id     - Get build status
   DELETE /api/v1/build/:id     - Cancel build
   ```

2. **Lambda Service**
   ```
   POST   /api/v1/lambda        - Create lambda
   GET    /api/v1/lambda/:id    - Get lambda
   DELETE /api/v1/lambda/:id    - Delete lambda
   ```

3. **Health/Metrics**
   ```
   GET    /healthz              - Health check
   GET    /readyz               - Readiness check
   GET    /metrics              - Prometheus metrics
   ```

## 🛠️ Testing Tools

### CORS Testing
```bash
# Test wildcard origin
curl -H "Origin: https://evil.com" \
  -H "Access-Control-Request-Method: POST" \
  -X OPTIONS http://api/endpoint

# Test null origin
curl -H "Origin: null" \
  http://api/endpoint

# Test origin reflection
curl -H "Origin: https://attacker.com" \
  http://api/endpoint
```

### Rate Limit Testing
```bash
# Burst test
for i in {1..100}; do
  curl http://api/endpoint &
done
wait

# Check rate limit headers
curl -I http://api/endpoint | grep -i ratelimit
```

### HTTP Method Testing
```bash
# Test method override
curl -X POST -H "X-HTTP-Method-Override: DELETE" \
  http://api/endpoint

# Test TRACE method
curl -X TRACE http://api/endpoint

# Test method confusion
curl -X GET -d "data=value" http://api/endpoint
```

## 📊 Success Metrics

- **Zero** CORS misconfiguration vulnerabilities
- **100%** security headers present
- **100%** rate limiting enforced
- **Zero** information leakage in errors
- **Zero** unauthenticated access to protected endpoints

## 🚨 Incident Response

If API security issue is discovered:

1. **Immediate** (< 5 min)
   - Enable WAF if available
   - Restrict API access
   - Block malicious IPs

2. **Short-term** (< 1 hour)
   - Patch configuration
   - Review access logs
   - Update security headers

3. **Long-term** (< 24 hours)
   - API security audit
   - Implement API gateway
   - Add automated security tests

## 📚 Related Stories

- **SEC-001:** Authentication & Authorization Bypass
- **SEC-002:** Input Validation & Injection Attacks
- **SEC-008:** Denial of Service & Resource Exhaustion
- **BACKEND-009:** API Management

## 🔗 References

- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
- [OWASP CORS](https://owasp.org/www-community/attacks/CORS_OriginHeaderScrutiny)
- [Mozilla HTTP Security Headers](https://infosec.mozilla.org/guidelines/web_security)
- [API Security Checklist](https://github.com/shieldfy/API-Security-Checklist)

---

**Test File:** `internal/security/security_003_api_security_test.go`  
**Owner:** Security Team  
**Last Updated:** October 29, 2025

