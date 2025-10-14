# 🔐 Security - Homepage System

Comprehensive security implementation and best practices for the homepage system.

## 📋 Overview

The homepage implements multiple layers of security following defense-in-depth principles, with automated scanning, secure secrets management, and production-ready practices.

## 🛡️ Security Architecture

```
Internet
    ↓ HTTPS/TLS
Cloudflare (DDoS, WAF)
    ↓ Encrypted
Kubernetes Ingress
    ↓ Network Policies
Services (Internal)
    ↓ Sealed Secrets
Sensitive Data
```

## 🔒 Security Layers

### 1. Network Security

**Cloudflare CDN (Optional):**
- DDoS protection
- Web Application Firewall (WAF)
- Bot management
- SSL/TLS termination
- Rate limiting

**Kubernetes Network:**
- Internal ClusterIP services
- Network policies (future)
- Service mesh (future)

**Configuration:**
```yaml
# No direct external exposure
services:
  - name: api
    type: ClusterIP  # Internal only
  - name: frontend
    type: ClusterIP  # Internal only
```

### 2. API Security

**Proxy Pattern:**
- Agent-SRE not directly exposed
- All traffic through homepage API
- Request validation
- Header forwarding control

**Implementation:**
```go
// api/handlers/agent_sre.go
- Validates requests
- Controls headers
- Timeouts (30s)
- Error sanitization
```

**CORS Configuration:**
```go
AllowOrigins: ["https://lucena.cloud"]
AllowMethods: ["GET", "POST", "PUT", "DELETE"]
AllowCredentials: true
```

**Compression:**
- Gzip enabled
- Reduces bandwidth
- Mitigates compression attacks

### 3. Secrets Management

**Method:** Sealed Secrets

**Sealed Secrets:**
```yaml
# sealed-secrets/bruno-site-secret.yaml
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: bruno-site-secret
  namespace: homepage
spec:
  encryptedData:
    password: <encrypted>
```

**Secrets:**
- Database passwords
- Redis passwords
- MinIO credentials
- Cloudflare API tokens

**Best Practices:**
- ✅ Never commit plain secrets
- ✅ Use sealed-secrets controller
- ✅ Rotate regularly
- ✅ Audit access

### 4. Container Security

**Image Security:**
```dockerfile
# Non-root user
USER nonroot:nonroot

# Read-only filesystem
securityContext:
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 65532
```

**Resource Limits:**
```yaml
resources:
  limits:
    cpu: 2000m
    memory: 2Gi
  requests:
    cpu: 500m
    memory: 512Mi
```

**Security Context:**
```yaml
securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop: ["ALL"]
  seccompProfile:
    type: RuntimeDefault
```

### 5. Application Security

**Input Validation:**
```go
// Validate all inputs
if err := validator.Validate(input); err != nil {
    return BadRequest
}
```

**SQL Injection Prevention:**
```go
// Use parameterized queries
db.Where("id = ?", id).First(&project)
```

**XSS Prevention:**
```typescript
// Sanitize user input
import DOMPurify from 'dompurify'
const clean = DOMPurify.sanitize(userInput)
```

**Error Handling:**
```go
// Never expose internal errors
if err != nil {
    log.Error(err)
    return "Internal server error"
}
```

## 🔍 Automated Security Scanning

### GitHub Actions Workflows

**1. Trivy Vulnerability Scanner**
```yaml
# Scans code, dependencies, containers
- uses: aquasecurity/trivy-action@master
  with:
    scan-type: 'fs'
    format: 'sarif'
```

**2. Go Vulnerability Check**
```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

**3. NPM Audit**
```bash
npm audit --production
```

**4. Security Code Scanning**
```yaml
# CodeQL analysis
- uses: github/codeql-action/upload-sarif@v3
```

### CI/CD Security

**Workflows:**
- `homepage-tests.yml` - Basic security checks
- `homepage-pr-check.yml` - Comprehensive scanning
- `homepage-nightly-tests.yml` - Deep security audit

**Checks:**
- ✅ Dependency vulnerabilities
- ✅ Container vulnerabilities
- ✅ Code security issues
- ✅ Secret detection
- ✅ License compliance

## 🔐 Access Control

### Current State

**Public Access:**
- Homepage (public site)
- Chatbot (public feature)

**Internal Only:**
- Database (PostgreSQL)
- Cache (Redis)
- Admin API (future)

### Future Implementation

**Authentication:**
```yaml
# JWT tokens for admin API
auth:
  method: JWT
  issuer: homepage-api
  audience: homepage-admin
```

**Authorization:**
```yaml
# Role-based access
roles:
  - admin: full access
  - editor: content management
  - viewer: read only
```

## 🛡️ Security Headers

**Implemented:**
```yaml
headers:
  X-Content-Type-Options: nosniff
  X-Frame-Options: DENY
  X-XSS-Protection: 1; mode=block
  Strict-Transport-Security: max-age=31536000
  Content-Security-Policy: default-src 'self'
```

**Configuration:** See `frontend/nginx.conf`

## 📊 Security Monitoring

### Logging

**Security Events:**
```go
log.Info("Authentication attempt", 
    "user", user,
    "ip", clientIP,
    "result", result)
```

**Audit Trail:**
- All API requests logged
- Failed authentication attempts
- Configuration changes
- Secret access

### Metrics

**Security Metrics:**
```
# Failed requests
http_requests_failed_total{reason="auth"}

# Rate limit hits
rate_limit_exceeded_total

# Suspicious activity
suspicious_requests_total
```

## 🚨 Incident Response

### Detection

**Automated:**
- Log analysis
- Metric anomalies
- Security alerts

**Manual:**
- Regular security audits
- Penetration testing
- Code reviews

### Response Plan

1. **Detect** - Identify security incident
2. **Contain** - Isolate affected systems
3. **Investigate** - Analyze root cause
4. **Remediate** - Fix vulnerability
5. **Document** - Update procedures

## 🔄 Security Updates

### Dependency Updates

**Automated:**
```yaml
# Dependabot
dependabot:
  package-ecosystem: "gomod"
  schedule:
    interval: "weekly"
```

**Manual:**
```bash
# Update Go dependencies
go get -u ./...
go mod tidy

# Update npm dependencies
npm update
npm audit fix
```

### Image Updates

**Strategy:**
- Use specific versions (not `latest`)
- Scan before deployment
- Test in staging
- Roll out gradually

## 🧪 Security Testing

### Unit Tests

```bash
# Test authentication
go test -v ./auth/

# Test authorization
go test -v ./middleware/
```

### Integration Tests

```bash
# Test security headers
./tests/security/test-headers.sh

# Test CORS
./tests/security/test-cors.sh
```

### Penetration Testing

**Scope:**
- OWASP Top 10
- API security
- Authentication bypass
- Injection attacks
- XSS/CSRF

**Frequency:** Quarterly

## 📋 Security Checklist

### Development

- [x] Use parameterized queries
- [x] Validate all inputs
- [x] Sanitize outputs
- [x] Use HTTPS everywhere
- [x] Implement CORS properly
- [x] Log security events
- [x] Handle errors securely
- [x] Use secure dependencies

### Deployment

- [x] Secrets in Sealed Secrets
- [x] Non-root containers
- [x] Resource limits set
- [x] Network policies (planned)
- [x] Security scanning enabled
- [x] Monitoring configured
- [x] Backup strategy
- [x] Incident response plan

### Operations

- [x] Regular updates
- [x] Security audits
- [x] Log monitoring
- [x] Access reviews
- [x] Backup testing
- [x] DR drills
- [x] Documentation updated

## 🎯 Security Best Practices

### Code Security

```go
// ✅ Good
db.Where("id = ?", userID).First(&user)

// ❌ Bad
db.Where(fmt.Sprintf("id = %s", userID)).First(&user)
```

### Secret Management

```yaml
# ✅ Good
env:
  - name: DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: db-secret
        key: password

# ❌ Bad
env:
  - name: DB_PASSWORD
    value: "hardcoded-password"
```

### Error Handling

```go
// ✅ Good
if err != nil {
    log.Error("Database error", "error", err)
    return errors.New("internal server error")
}

// ❌ Bad
if err != nil {
    return err // Exposes internal details
}
```

## 📚 Security Resources

### Documentation

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Kubernetes Security](https://kubernetes.io/docs/concepts/security/)
- [Go Security](https://golang.org/doc/security)
- [NPM Security](https://docs.npmjs.com/auditing-package-dependencies-for-security-vulnerabilities)

### Tools

- **Trivy** - Vulnerability scanner
- **govulncheck** - Go vulnerability checker
- **npm audit** - npm security audit
- **CodeQL** - Static analysis
- **sealed-secrets** - Secret encryption

## 🔒 Compliance

### Standards

- **OWASP** - Security best practices
- **CIS** - Configuration benchmarks
- **NIST** - Security framework

### Certifications

- Container security
- Kubernetes hardening
- Cloud security

## 🎉 Security Status

| Component | Status | Last Scan |
|-----------|--------|-----------|
| Dependencies | ✅ Clean | 2025-10-08 |
| Containers | ✅ No CVEs | 2025-10-08 |
| Code | ✅ No issues | 2025-10-08 |
| Secrets | ✅ Encrypted | 2025-10-08 |
| Network | ✅ Isolated | 2025-10-08 |

---

**Security Version:** 1.0.0  
**Last Audit:** 2025-10-08  
**Status:** ✅ Production Ready  
**Contact:** Security issues → [GitHub Security](https://github.com/brunovlucena/homelab/security)

