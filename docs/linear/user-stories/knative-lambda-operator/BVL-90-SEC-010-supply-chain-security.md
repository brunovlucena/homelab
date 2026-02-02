# SEC-010: Supply Chain & Dependency Security Testing

**Priority**: P0 | **Status**: 📋 Backlog K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-245/sec-010-supply-chain-and-dependency-security-testing

**Priority:** P1 | **Story Points:** 8

## 📋 User Story

**As a** Principal Pentester  
**I want to** validate the security of the software supply chain and dependencies  
**So that** compromised or vulnerable dependencies cannot be introduced into the system

## 🎯 Acceptance Criteria

### AC1: Container Image Vulnerability Scanning
**Given** container images are built and deployed  
**When** scanning images for vulnerabilities  
**Then** no critical/high severity vulnerabilities should exist

**Security Tests:**
- ✅ All images scanned with Trivy before deployment
- ✅ Critical vulnerabilities blocked from deployment
- ✅ Base images from trusted sources only
- ✅ Image signatures verified
- ✅ Image provenance tracked (SLSA)

**Attack Scenarios:**
```bash
# Scan image for vulnerabilities
trivy image --severity HIGH,CRITICAL \
  knative-lambdas/builder:latest

# Check for known malicious packages
trivy image --list-all-pkgs \
  knative-lambdas/builder:latest | \
  grep -i "malicious\ | backdoor"

# Verify image signature
cosign verify --key cosign.pub \
  knative-lambdas/builder:latest
```

### AC2: Dependency Version Pinning
**Given** application dependencies are specified  
**When** reviewing dependency manifests  
**Then** versions should be pinned (not using `latest`)

**Security Tests:**
- ✅ Dockerfile: `FROM golang:1.24.4` not `FROM golang:latest`
- ✅ go.mod: exact versions specified
- ✅ Helm charts: chart versions pinned
- ✅ Python: requirements.txt with versions
- ✅ npm: package-lock.json committed

**Vulnerable Patterns:**
```dockerfile
# BAD: Unpinned versions
FROM golang:latest
FROM alpine:latest
RUN apt-get install -y package

# GOOD: Pinned versions
FROM golang:1.24.4-alpine3.19
FROM alpine:3.19.1
RUN apt-get install -y package=1.2.3-1
```

### AC3: Software Bill of Materials (SBOM)
**Given** applications have many dependencies  
**When** generating SBOM  
**Then** all components should be documented

**Security Tests:**
- ✅ SBOM generated for all images
- ✅ SBOM includes direct and transitive dependencies
- ✅ SBOM format: SPDX or CycloneDX
- ✅ SBOM signed and verifiable
- ✅ SBOM stored with images

**Generate SBOM:**
```bash
# Generate SBOM with Syft
syft knative-lambdas/builder:latest -o spdx-json > sbom.json

# Generate SBOM with Trivy
trivy image --format cyclonedx \
  knative-lambdas/builder:latest > sbom.cdx.json

# Sign SBOM
cosign sign-blob --key cosign.key sbom.json > sbom.json.sig
```

### AC4: Dependency Vulnerability Scanning
**Given** dependencies may have vulnerabilities  
**When** scanning dependency manifests  
**Then** vulnerable dependencies should be flagged

**Security Tests:**
- ✅ Go modules scanned (`go list -m -json all | nancy`)
- ✅ Python packages scanned (`safety check`)
- ✅ npm packages scanned (`npm audit`)
- ✅ Helm charts scanned (`helm lint --strict`)
- ✅ Automated scanning in CI/CD

**Scanning Commands:**
```bash
# Scan Go dependencies
go list -m -json all | nancy sleuth

# Scan Python dependencies
safety check -r requirements.txt

# Scan container for CVEs
grype knative-lambdas/builder:latest
```

### AC5: Build Process Security
**Given** images are built in CI/CD  
**When** testing build security  
**Then** builds should be reproducible and verifiable

**Security Tests:**
- ✅ Multi-stage builds used
- ✅ No secrets in build args
- ✅ Build cache not shared across builds
- ✅ Build provenance recorded
- ✅ Builds run in isolated environment
- ✅ Build attestations signed (SLSA Level 2+)

**Secure Build Example:**
```dockerfile
# Stage 1: Build
FROM golang:1.24.4 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 go build -o app

# Stage 2: Runtime
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /build/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
```

### AC6: Third-Party Service Security
**Given** external services are integrated  
**When** reviewing third-party integrations  
**Then** services should be vetted and monitored

**Security Tests:**
- ✅ Third-party services vetted before integration
- ✅ API keys rotated regularly
- ✅ Service access logged
- ✅ Fallback mechanisms for service failures
- ✅ Data shared with third-parties minimized

**Third-Party Services:**
- AWS services (S3, ECR, CloudWatch)
- GitHub (CI/CD)
- Docker Hub (if used)
- Helm chart repositories

### AC7: Parser Upload Security
**Given** users upload parser code  
**When** accepting parser uploads  
**Then** uploads should be scanned and validated

**Security Tests:**
- ✅ Parser code scanned for malicious content
- ✅ Static analysis performed (Bandit for Python)
- ✅ Dependency check (no known malicious packages)
- ✅ Sandbox execution validated
- ✅ File type validation (must be valid Python)
- ✅ Size limits enforced (<10MB)

**Malicious Pattern Detection:**
```python
# Detect dangerous imports
DANGEROUS_IMPORTS = [
    'os',           # File system access
    'subprocess',   # Command execution
    'socket',       # Network access
    'requests',     # HTTP requests
    '__import__',   # Dynamic imports
    'eval',         # Code execution
    'exec',         # Code execution
]

# Detect obfuscated code
if 'base64' in code and 'decode' in code:
    flag_as_suspicious()
```

### AC8: License Compliance
**Given** open source dependencies are used  
**When** checking license compliance  
**Then** licenses should be compatible

**Security Tests:**
- ✅ License scanning automated
- ✅ Copyleft licenses identified (GPL, AGPL)
- ✅ License compatibility verified
- ✅ Attribution requirements met
- ✅ License violations blocked

**Scanning Commands:**
```bash
# Scan Go module licenses
go-licenses check ./...

# Scan npm licenses
license-checker --production

# Generate license report
trivy image --format json \
  knative-lambdas/builder:latest | \
  jq '.Results[].Licenses'
```

## 🔴 Attack Surface Analysis

### Supply Chain Attack Vectors

1. **Compromised Base Images**
   - Malicious Dockerfile `FROM` images
   - Backdoored official images
   - Typosquatting (alpine vs alpne)

2. **Dependency Confusion**
   - Internal package names matching public packages
   - Higher version numbers in malicious packages
   - Private registry misconfiguration

3. **Malicious Dependencies**
   - npm packages with malicious postinstall scripts
   - Python packages with backdoors
   - Go modules with malicious code

4. **Build Process Compromise**
   - Compromised CI/CD pipeline
   - Stolen signing keys
   - Unauthorized image push

5. **User-Uploaded Parsers**
   - Malicious Python code
   - Obfuscated exploits
   - Resource exhaustion attacks

## 🛠️ Testing Tools

### Vulnerability Scanning
```bash
# Comprehensive image scan
trivy image --severity HIGH,CRITICAL \
  --ignore-unfixed \
  knative-lambdas/builder:latest

# Grype scanner
grype knative-lambdas/builder:latest

# Snyk container scan
snyk container test knative-lambdas/builder:latest
```

### SBOM Generation
```bash
# Syft SBOM
syft knative-lambdas/builder:latest \
  -o spdx-json=sbom.spdx.json

# Trivy SBOM
trivy image --format cyclonedx \
  knative-lambdas/builder:latest

# Docker SBOM
docker sbom knative-lambdas/builder:latest
```

### Dependency Scanning
```bash
# Go dependencies
go list -m -json all | nancy sleuth
govulncheck ./...

# Python dependencies
pip-audit -r requirements.txt
safety check --json

# Node dependencies
npm audit --audit-level=high
yarn audit --level high
```

### Signing and Verification
```bash
# Sign container image
cosign sign --key cosign.key \
  knative-lambdas/builder:latest

# Verify image signature
cosign verify --key cosign.pub \
  knative-lambdas/builder:latest

# Generate and sign provenance
cosign attest --predicate provenance.json \
  --key cosign.key \
  knative-lambdas/builder:latest
```

## 📊 Success Metrics

- **Zero** critical vulnerabilities in production images
- **100%** dependencies scanned before deployment
- **100%** images signed and verified
- **100%** SBOM coverage
- **Zero** malicious packages detected

## 🚨 Incident Response

If supply chain compromise is detected:

1. **Immediate** (< 5 min)
   - Stop deployment pipeline
   - Quarantine affected images
   - Block compromised dependencies

2. **Short-term** (< 30 min)
   - Identify scope of compromise
   - Roll back to known good versions
   - Rotate signing keys

3. **Long-term** (< 24 hours)
   - Audit entire supply chain
   - Implement additional controls
   - Update security policies

## 📚 Related Stories

- **SEC-002:** Input Validation & Injection Attacks
- **SEC-006:** Secrets Exposure & Credential Leakage
- **SRE-014:** Security Incident Response
- **DEVOPS-004:** CI/CD Pipeline

## 🔗 References

- [SLSA Framework](https://slsa.dev/)
- [Supply Chain Levels for Software Artifacts](https://slsa.dev/spec/v1.0/)
- [Sigstore (Cosign, Rekor)](https://www.sigstore.dev/)
- [OWASP Dependency-Check](https://owasp.org/www-project-dependency-check/)
- [in-toto Supply Chain Security](https://in-toto.io/)

---

**Test File:** `internal/security/security_010_supply_chain_test.go`  
**Owner:** Security Team  
**Last Updated:** October 29, 2025

