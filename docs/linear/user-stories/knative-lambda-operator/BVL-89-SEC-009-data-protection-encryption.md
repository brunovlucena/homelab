# SEC-009: Data Protection & Encryption Testing

**Priority**: P0 | **Status**: 📋 Backlog K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-249/sec-009-data-protection-and-encryption-testing

**Priority:** P1 | **Story Points:** 5

## 📋 User Story

**As a** Principal Pentester  
**I want to** validate that data is properly encrypted at rest and in transit  
**So that** sensitive data cannot be intercepted or accessed by unauthorized parties

## 🎯 Acceptance Criteria

### AC1: Encryption at Rest
**Given** sensitive data is stored  
**When** examining storage encryption  
**Then** all data should be encrypted

**Security Tests:**
- ✅ etcd encrypted at rest
- ✅ S3 buckets use server-side encryption (SSE)
- ✅ EBS volumes encrypted
- ✅ RDS/database encrypted
- ✅ Secrets encrypted in Kubernetes
- ✅ Backup data encrypted

**Attack Scenarios:**
```bash
# Attempt to read etcd data directly
etcdctl get / --prefix
# Expected: Encrypted data or access denied

# Attempt to read EBS snapshot
aws ec2 describe-snapshots --snapshot-ids <id>
# Expected: Encrypted: true
```

### AC2: Encryption in Transit (TLS)
**Given** data is transmitted over networks  
**When** examining network traffic  
**Then** all sensitive traffic should use TLS

**Security Tests:**
- ✅ All external APIs use HTTPS
- ✅ Internal service communication encrypted (mTLS)
- ✅ Database connections use TLS
- ✅ RabbitMQ uses TLS
- ✅ Metrics endpoints use TLS
- ✅ No SSLv3, TLS 1.0, TLS 1.1 (only TLS 1.2+)

**Test Commands:**
```bash
# Test TLS version
nmap --script ssl-enum-ciphers -p 443 api.domain.com

# Check certificate validity
openssl s_client -connect api.domain.com:443 -tls1_2

# Test for weak ciphers
testssl api.domain.com

# Capture traffic to verify encryption
tcpdump -i any -w capture.pcap port 443
tshark -r capture.pcap -Y "http"  # Should be empty
```

### AC3: Certificate Management
**Given** TLS certificates are used  
**When** examining certificate configuration  
**Then** certificates should be properly managed

**Security Tests:**
- ✅ Certificates not self-signed (in production)
- ✅ Certificate expiration > 30 days
- ✅ Strong key length (≥2048 bits RSA, ≥256 bits ECC)
- ✅ Subject Alternative Names (SAN) configured
- ✅ Certificate chain complete
- ✅ OCSP stapling enabled
- ✅ Certificate transparency logged

**Vulnerable Configurations:**
- ❌ Self-signed certificates in production
- ❌ Expired certificates
- ❌ Weak key sizes (1024 bits)
- ❌ Missing intermediate certificates

### AC4: Key Management
**Given** encryption keys are used  
**When** testing key security  
**Then** keys should be properly protected

**Security Tests:**
- ✅ Keys stored in AWS KMS or similar
- ✅ Key rotation enabled
- ✅ Keys not hardcoded in code
- ✅ Keys not in environment variables
- ✅ Key access logged (CloudTrail)
- ✅ Separate keys per environment

**Attack Scenarios:**
```bash
# Search for hardcoded keys
grep -r "-----BEGIN PRIVATE KEY-----" .
grep -r "-----BEGIN RSA PRIVATE KEY-----" .

# Check environment for keys
env | grep -i "key\ | secret"
```

### AC5: Database Encryption
**Given** databases store sensitive data  
**When** examining database security  
**Then** encryption should be enforced

**Security Tests:**
- ✅ Transparent Data Encryption (TDE) enabled
- ✅ Connection encryption (TLS/SSL)
- ✅ Encrypted backups
- ✅ Column-level encryption for PII
- ✅ Key rotation supported

### AC6: Message Queue Encryption
**Given** events flow through RabbitMQ  
**When** testing message security  
**Then** messages should be protected

**Security Tests:**
- ✅ TLS enforced for connections
- ✅ Messages encrypted in transit
- ✅ Queue storage encrypted at rest
- ✅ Authentication required
- ✅ No plaintext credentials

### AC7: Backup Encryption
**Given** backups contain sensitive data  
**When** testing backup security  
**Then** backups should be encrypted

**Security Tests:**
- ✅ Velero backups encrypted
- ✅ S3 backup buckets encrypted
- ✅ Database backup files encrypted
- ✅ Encryption keys rotated
- ✅ Backup access audited

### AC8: Data Masking and Tokenization
**Given** PII may be processed  
**When** testing data handling  
**Then** sensitive data should be masked/tokenized

**Security Tests:**
- ✅ PII masked in logs
- ✅ Credit card numbers tokenized
- ✅ API responses redact sensitive fields
- ✅ Test environments use synthetic data
- ✅ Data retention policies enforced

## 🔴 Attack Surface Analysis

### Encryption Points

1. **Data at Rest**
   - Kubernetes etcd
   - S3 buckets (parsers, artifacts)
   - EBS volumes
   - RDS databases
   - Secrets

2. **Data in Transit**
   - API endpoints (HTTPS)
   - Internal services (mTLS)
   - Database connections (TLS)
   - RabbitMQ (TLS)
   - AWS service calls (TLS)

3. **Key Storage**
   - AWS KMS
   - Kubernetes Secrets
   - Certificate Manager

## 🛠️ Testing Tools

### TLS/SSL Testing
```bash
# Test SSL/TLS configuration
testssl https://api.domain.com

# Check certificate details
openssl s_client -connect api.domain.com:443 -showcerts

# Test cipher suites
nmap --script ssl-enum-ciphers -p 443 api.domain.com

# Check for weak protocols
sslscan api.domain.com
```

### Encryption Verification
```bash
# Check S3 bucket encryption
aws s3api get-bucket-encryption --bucket bucket-name

# Check EBS encryption
aws ec2 describe-volumes --volume-ids vol-xxx \
  --query 'Volumes[*].{ID:VolumeId,Encrypted:Encrypted}'

# Check RDS encryption
aws rds describe-db-instances \
  --query 'DBInstances[*].[DBInstanceIdentifier,StorageEncrypted]'

# Check etcd encryption
kubectl get secret -n kube-system encryption-config -o yaml
```

### Certificate Validation
```bash
# Check certificate expiration
echo | openssl s_client -connect api.domain.com:443 2>/dev/null | \
  openssl x509 -noout -dates

# Validate certificate chain
openssl s_client -connect api.domain.com:443 -showcerts < /dev/null

# Check for weak key size
echo | openssl s_client -connect api.domain.com:443 2>/dev/null | \
  openssl x509 -noout -text | grep "Public-Key"
```

## 📊 Success Metrics

- **100%** data encrypted at rest
- **100%** sensitive traffic uses TLS 1.2+
- **Zero** weak cipher suites
- **Zero** expired certificates
- **100%** key rotation implemented

## 🚨 Incident Response

If encryption failure is detected:

1. **Immediate** (< 15 min)
   - Block access to unencrypted data
   - Enable encryption
   - Rotate potentially compromised keys

2. **Short-term** (< 1 hour)
   - Audit all encryption settings
   - Review access logs
   - Identify exposed data

3. **Long-term** (< 24 hours)
   - Implement encryption everywhere
   - Update security policies
   - Conduct security training

## 📚 Related Stories

- **SEC-006:** Secrets Exposure & Credential Leakage
- **SEC-007:** Network Segmentation & Data Exfiltration
- **SRE-008:** Certificate Lifecycle Management
- **SRE-014:** Security Incident Response

## 🔗 References

- [OWASP Cryptographic Storage Cheatsheet](https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html)
- [NIST Encryption Standards](https://csrc.nist.gov/projects/cryptographic-standards-and-guidelines)
- [AWS KMS Best Practices](https://docs.aws.amazon.com/kms/latest/developerguide/best-practices.html)
- [Kubernetes Encryption at Rest](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/)

---

**Test File:** `internal/security/security_009_data_protection_test.go`  
**Owner:** Security Team  
**Last Updated:** October 29, 2025

