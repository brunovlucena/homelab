# SEC-006: Secrets Exposure & Credential Leakage Testing

**Priority**: P0 | **Status**: 📋 Backlog K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-250/sec-006-secrets-exposure-and-credential-leakage-testing

**Priority:** P0 | **Story Points:** 13

## 📋 User Story

**As a** Principal Pentester  
**I want to** validate that secrets and credentials cannot be exposed or leaked  
**So that** sensitive information remains protected from unauthorized access

## 🎯 Acceptance Criteria

### AC1: Kubernetes Secrets Protection
**Given** secrets are stored in Kubernetes  
**When** attempting to access secrets  
**Then** access should be properly controlled

**Security Tests:**
- ✅ Secrets not accessible from unauthorized namespaces
- ✅ Secrets encrypted at rest (etcd encryption)
- ✅ Secrets not exposed in logs
- ✅ Secrets not exposed in environment variable listings
- ✅ RBAC prevents secret enumeration
- ✅ Secret rotation supported

**Attack Scenarios:**
```bash
# Attempt to list secrets across namespaces
kubectl get secrets --all-namespaces

# Attempt to read secret from different namespace
kubectl get secret -n prod db-password -o yaml

# Attempt to extract secrets from pod
kubectl exec -it <pod> -- env | grep -i secret
```

### AC2: Environment Variable Exposure
**Given** secrets may be passed as environment variables  
**When** inspecting running processes or logs  
**Then** secrets should not be visible

**Security Tests:**
- ✅ Secrets not in plain text environment variables
- ✅ `/proc/<pid>/environ` not readable
- ✅ `ps aux` doesn't expose secrets
- ✅ Debug endpoints don't leak environment
- ✅ Error messages don't contain secrets

**Vulnerable Patterns:**
```yaml
# BAD: Secret in plain env var
env:
- name: DB_PASSWORD
  value: "SuperSecret123!"  # ❌ Exposed

# GOOD: Secret from SecretRef
env:
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: db-credentials
      key: password  # ✅ Protected
```

### AC3: Secrets in Logs Prevention
**Given** application logging occurs  
**When** reviewing logs for sensitive data  
**Then** secrets should never appear in logs

**Security Tests:**
- ✅ No credentials in application logs
- ✅ No API keys in logs
- ✅ No tokens in logs
- ✅ Structured logging with field redaction
- ✅ Log aggregation sanitizes secrets

**Patterns to Detect:**
```regex
# Patterns that should NOT appear in logs
password=.*
api[_-]?key=.*
token=.*
secret=.*
aws[_-]?access[_-]?key[_-]?id=AKIA[A-Z0-9]{16}
aws[_-]?secret[_-]?access[_-]?key=.{40}
Bearer\s+[A-Za-z0-9\-._~+/]+=*
```

### AC4: Version Control Secrets Scanning
**Given** code is stored in Git repositories  
**When** scanning commit history  
**Then** no secrets should be found

**Security Tests:**
- ✅ No secrets in current codebase
- ✅ No secrets in Git history
- ✅ Pre-commit hooks detect secrets
- ✅ CI/CD pipeline scans for secrets
- ✅ `.gitignore` blocks credential files

**Tools to Use:**
```bash
# gitleaks
gitleaks detect --source . --verbose

# trufflehog
trufflehog git file://. --only-verified

# git-secrets
git secrets --scan-history
```

### AC5: API Response Secret Leakage
**Given** APIs return data to clients  
**When** examining API responses  
**Then** sensitive data should be redacted

**Security Tests:**
- ✅ Passwords never returned in responses
- ✅ API keys redacted (show last 4 chars only)
- ✅ Tokens not included in GET responses
- ✅ Database connection strings sanitized
- ✅ Internal IPs/hostnames not exposed

**Example Redaction:**
```json
{
  "api_key": "sk-**********************xyz123",
  "database_url": "postgres://*****:*****@internal-db:5432/db"
}
```

### AC6: Container Image Secrets
**Given** container images may be built with secrets  
**When** inspecting image layers  
**Then** no secrets should be embedded

**Security Tests:**
- ✅ No secrets in Dockerfile
- ✅ No secrets in image environment variables
- ✅ No secrets in image layers
- ✅ Multi-stage builds used to remove secrets
- ✅ `.dockerignore` excludes credential files

**Attack Scenarios:**
```bash
# Inspect image for secrets
docker history <image>
docker inspect <image> | grep -i "password\ | key\ | secret"

# Extract filesystem
docker save <image> -o image.tar
tar xf image.tar
grep -r "password\ | api_key" .
```

### AC7: CloudEvent Data Sanitization
**Given** CloudEvents may contain sensitive data  
**When** events are logged or queued  
**Then** sensitive fields should be redacted

**Security Tests:**
- ✅ PII redacted in event logs
- ✅ Credentials not logged in event data
- ✅ RabbitMQ queue data encrypted
- ✅ Event replay doesn't expose secrets
- ✅ Dead letter queue sanitized

**Sensitive Fields to Redact:**
- `credentials`
- `api_key`
- `token`
- `password`
- `secret`

### AC8: AWS Credentials Protection
**Given** AWS credentials used for service access  
**When** attempting to steal credentials  
**Then** credentials should be protected

**Security Tests:**
- ✅ No hardcoded AWS keys in code
- ✅ IRSA used instead of static credentials
- ✅ Temporary credentials expire (<1 hour)
- ✅ Credentials not in environment variables
- ✅ CloudTrail logs credential usage
- ✅ Access key rotation enforced (90 days)

**Attack Scenarios:**
```bash
# Attempt to steal credentials from metadata
curl http://169.254.169.254/latest/meta-data/iam/security-credentials/

# Search for hardcoded keys
grep -r "AKIA[A-Z0-9]{16}" .

# Check environment
env | grep -i "AWS_ACCESS_KEY\ | AWS_SECRET"
```

### AC9: Build-Time Secrets Handling
**Given** secrets needed during image builds  
**When** building container images  
**Then** secrets should not persist in images

**Security Tests:**
- ✅ BuildKit secrets used for build-time secrets
- ✅ Secrets not in final image layers
- ✅ Multi-stage builds remove intermediate secrets
- ✅ `.dockerignore` configured
- ✅ Kaniko secrets properly handled

**Secure Build Example:**
```dockerfile
# BAD: Secret persists in layer
ARG API_TOKEN
RUN echo "API_TOKEN=${API_TOKEN}" > /app/config  # ❌

# GOOD: BuildKit secret (doesn't persist)
RUN --mount=type=secret,id=api_token \
    API_TOKEN=$(cat /run/secrets/api_token) && \
    app-setup && \
    rm /run/secrets/api_token  # ✅
```

## 🔴 Attack Surface Analysis

### Secret Storage Locations

1. **Kubernetes Secrets**
   - Database credentials
   - API keys
   - TLS certificates

2. **Environment Variables**
   - AWS credentials
   - Service tokens
   - Feature flags

3. **ConfigMaps**
   - Application config (should not contain secrets)
   - Connection strings (sanitized)

4. **AWS Secrets Manager**
   - Database passwords
   - Third-party API keys
   - Encryption keys

5. **Container Images**
   - Application code
   - Configuration files
   - Build artifacts

6. **Application Logs**
   - CloudWatch Logs
   - Loki
   - Stdout/stderr

## 🛠️ Testing Tools

### Secret Scanning
```bash
# Scan Git repository
gitleaks detect --source . --report-format json --report-path gitleaks-report.json

# Scan Docker images
docker scan <image>
trivy image --severity HIGH,CRITICAL <image>

# Scan Kubernetes manifests
kubesec scan deployment.yaml

# Scan code for secrets
trufflehog filesystem --directory . --json
```

### Manual Testing
```bash
# Check Kubernetes secrets
kubectl get secrets --all-namespaces -o json | \
  jq '.items[] | {name: .metadata.name, data: .data | keys}'

# Check environment variables in pods
kubectl exec <pod> -- env

# Check logs for secrets
kubectl logs <pod> | grep -E "password | api[_-]?key | token | secret"

# Inspect container images
docker history --no-trunc <image>
dive <image>  # Interactive image explorer
```

### Automated Pipeline Integration
```yaml
# .github/workflows/secret-scan.yml
name: Secret Scanning
on: [push, pull_request]
jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Full history for gitleaks
      
      - name: Run gitleaks
        uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Scan Docker images
        run: | trivy image --exit-code 1 --severity HIGH,CRITICAL \
            <registry>/knative-lambda-builder:latest
```

## 📊 Success Metrics

- **Zero** secrets in version control
- **Zero** secrets in logs
- **Zero** secrets in container images
- **100%** secrets encrypted at rest
- **100%** secret rotation capability

## 🚨 Incident Response

If secret exposure is detected:

1. **Immediate** (< 5 min)
   - Rotate exposed credentials immediately
   - Revoke API keys/tokens
   - Block compromised accounts

2. **Short-term** (< 30 min)
   - Review access logs for unauthorized use
   - Remove secrets from exposed locations
   - Update code/configs

3. **Long-term** (< 24 hours)
   - Conduct full secret audit
   - Implement secret scanning in CI/CD
   - Train team on secure secret handling

## 📚 Related Stories

- **SEC-001:** Authentication & Authorization Bypass
- **SEC-005:** Cloud Resource Access Control
- **SEC-009:** Data Protection & Encryption
- **SRE-014:** Security Incident Response

## 🔗 References

- [OWASP Secrets Management Cheatsheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [Kubernetes Secrets Best Practices](https://kubernetes.io/docs/concepts/configuration/secret/)
- [AWS Secrets Manager Best Practices](https://docs.aws.amazon.com/secretsmanager/latest/userguide/best-practices.html)
- [gitleaks](https://github.com/gitleaks/gitleaks)
- [trufflehog](https://github.com/trufflesecurity/trufflehog)

---

**Test File:** `internal/security/security_006_secrets_exposure_test.go`  
**Owner:** Security Team  
**Last Updated:** October 29, 2025

