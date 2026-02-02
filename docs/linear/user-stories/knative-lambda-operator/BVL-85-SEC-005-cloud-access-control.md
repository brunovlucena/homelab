# SEC-005: Cloud Resource Access Control Testing

**Priority**: P0 | **Status**: 📋 Backlog K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-246/sec-005-cloud-resource-access-control-testing

**Priority:** P0 | **Story Points:** 8

## 📋 User Story

**As a** Principal Pentester  
**I want to** validate that cloud resources (AWS) are properly secured with least-privilege access  
**So that** unauthorized access to S3, ECR, and other AWS services is prevented

## 🎯 Acceptance Criteria

### AC1: IAM Roles for Service Accounts (IRSA) Security
**Given** pods use IRSA for AWS authentication  
**When** attempting to assume unauthorized roles  
**Then** role assumption should be blocked

**Security Tests:**
- ✅ Service accounts cannot assume arbitrary roles
- ✅ Role trust policy validates service account
- ✅ Role session tags enforced
- ✅ Temporary credentials expire (<1 hour)
- ✅ Cross-account access properly scoped

**Attack Scenarios:**
- ❌ Assume role from wrong namespace
- ❌ Assume role from different service account
- ❌ Role chaining to escalate privileges
- ❌ Steal STS credentials from pod metadata

### AC2: S3 Bucket Access Control
**Given** parsers are stored in S3 buckets  
**When** attempting unauthorized S3 access  
**Then** access should be denied

**Security Tests:**
- ✅ Bucket policies enforce least privilege
- ✅ Public access blocked (`aws:BlockPublicAccess`)
- ✅ Cross-account access denied (except explicit)
- ✅ MFA required for destructive operations
- ✅ Object versioning enabled
- ✅ Server-side encryption enforced

**Attack Scenarios:**
```bash
# Attempt to list other environments' buckets
aws s3 ls s3://knative-lambda-parsers/

# Attempt to write to production bucket
aws s3 cp malicious.py s3://knative-lambda-parsers/

# Attempt to delete objects
aws s3 rm s3://knative-lambda-parsers/parser-123
```

### AC3: ECR Repository Access Control
**Given** container images stored in ECR  
**When** attempting unauthorized ECR operations  
**Then** operations should be blocked

**Security Tests:**
- ✅ Image push restricted to CI/CD only
- ✅ Image pull authenticated
- ✅ Repository policies enforced
- ✅ Cross-account pull denied (unless explicit)
- ✅ Image scanning enforced
- ✅ Lifecycle policies active

**Attack Scenarios:**
- ❌ Push malicious image
- ❌ Pull images from other accounts
- ❌ Delete production images
- ❌ Disable image scanning

### AC4: AWS Metadata Service (IMDS) Protection
**Given** pods can access EC2 instance metadata  
**When** attempting to access IMDS from containers  
**Then** access should be restricted

**Security Tests:**
- ✅ IMDSv2 enforced (session token required)
- ✅ Hop limit = 1 (prevents container access)
- ✅ Metadata access logged
- ✅ Credentials cannot be stolen from IMDS
- ✅ Network policy blocks 169.254.169.254

**Attack Scenarios:**
```bash
# Attempt to access IMDSv1
curl http://169.254.169.254/latest/meta-data/

# Attempt to get node IAM credentials
curl http://169.254.169.254/latest/meta-data/iam/security-credentials/

# Attempt IMDSv2 without session token
TOKEN=$(curl -X PUT "http://169.254.169.254/latest/api/token" \
  -H "X-aws-ec2-metadata-token-ttl-seconds: 21600")
# Should fail from container
```

### AC5: Resource Policy Validation
**Given** AWS resources have resource-based policies  
**When** reviewing resource policies  
**Then** policies should follow least-privilege

**Security Tests:**
- ✅ No wildcard principals (`*`)
- ✅ Conditions enforce MFA where appropriate
- ✅ Source IP restrictions where applicable
- ✅ Encryption in transit enforced
- ✅ Secure transport (`aws:SecureTransport`)

**Vulnerable Policy Example:**
```json
{
  "Effect": "Allow",
  "Principal": "*",
  "Action": "s3:GetObject",
  "Resource": "arn:aws:s3:::bucket/*"
}
```

### AC6: Secrets Manager / Parameter Store Access
**Given** secrets stored in AWS Secrets Manager  
**When** attempting unauthorized secret access  
**Then** access should be denied

**Security Tests:**
- ✅ Secrets access requires specific IAM permissions
- ✅ Secret rotation enabled
- ✅ Least-privilege access per service
- ✅ Audit logging enabled
- ✅ Cross-account access denied

**Attack Scenarios:**
```bash
# Attempt to list all secrets
aws secretsmanager list-secrets

# Attempt to read secrets from other namespaces
aws secretsmanager get-secret-value \
  --secret-id knative-lambda/database-password

# Attempt to create malicious secret
aws secretsmanager create-secret \
  --name backdoor --secret-string "malicious"
```

### AC7: CloudWatch Logs Access Control
**Given** logs are sent to CloudWatch  
**When** attempting to access or manipulate logs  
**Then** access should be properly controlled

**Security Tests:**
- ✅ Log groups encrypted
- ✅ Read access restricted by role
- ✅ Write access restricted by role
- ✅ Log retention enforced
- ✅ Cannot delete logs (unless admin)
- ✅ Cross-account access denied

### AC8: VPC Endpoint Security
**Given** VPC endpoints provide private AWS service access  
**When** testing VPC endpoint policies  
**Then** policies should restrict access

**Security Tests:**
- ✅ VPC endpoint policies enforced
- ✅ Only required services accessible
- ✅ Source VPC validation
- ✅ DNS resolution restricted
- ✅ No internet gateway bypass

## 🔴 Attack Surface Analysis

### AWS Resources in Scope

1. **S3 Buckets**
   - `knative-lambda-fusion-modules-tmp`
   - `knative-lambda-fusion-modules-tmp`
   - Parser storage and build artifacts

2. **ECR Repositories**
   - `knative-lambdas/knative-lambda-builder`
   - `knative-lambdas/knative-lambda-sidecar`
   - Parser container images

3. **IAM Roles**
   - `knative-lambda-builder-role`
   - `knative-lambda-parser-role`
   - Service account mapping

4. **Secrets Manager**
   - Database credentials
   - API keys
   - Third-party tokens

5. **CloudWatch Logs**
   - `/aws/eks/knative-lambda`
   - Application logs
   - Audit trails

## 🛠️ Testing Tools

### AWS IAM Testing
```bash
# Test current permissions
aws sts get-caller-identity
aws iam list-attached-user-policies --user-name test-user

# Enumerate permissions
enumerate-iam.py --access-key <key> --secret-key <secret>

# Test privilege escalation
pmapper graph --account 123456789012
pmapper query "preset privesc" --account 123456789012

# Policy Validator
aws accessanalyzer validate-policy --policy-document file://policy.json
```

### S3 Bucket Testing
```bash
# Test bucket permissions
aws s3api get-bucket-acl --bucket knative-lambda-parsers

# Test public access
aws s3api get-public-access-block --bucket knative-lambda-parsers

# Attempt unauthorized access
aws s3 ls s3://knative-lambda-parsers/ \
  --profile unauthorized-user
```

### ECR Testing
```bash
# Test repository permissions
aws ecr describe-repositories

# Test image pull
aws ecr get-login-password --region us-west-2 | \
  docker login --username AWS --password-stdin <ecr-url>

# Test image push (should fail)
docker push <ecr-url>/unauthorized-image:latest
```

### Metadata Service Testing
```bash
# From inside pod
curl http://169.254.169.254/latest/meta-data/
# Expected: Connection timeout or denied

# Test IMDSv2
TOKEN=$(curl -X PUT "http://169.254.169.254/latest/api/token" \
  -H "X-aws-ec2-metadata-token-ttl-seconds: 21600" 2>/dev/null)
# Expected: No token received from container
```

## 📊 Success Metrics

- **Zero** overly permissive IAM policies
- **Zero** public S3 buckets
- **Zero** unauthenticated ECR access
- **100%** IMDS access blocked from containers
- **100%** encryption enforced on sensitive resources

## 🚨 Incident Response

If cloud access control breach is detected:

1. **Immediate** (< 5 min)
   - Revoke compromised credentials
   - Rotate AWS access keys
   - Enable CloudTrail logging

2. **Short-term** (< 30 min)
   - Review CloudTrail for unauthorized access
   - Audit all IAM policies
   - Check S3 bucket access logs

3. **Long-term** (< 24 hours)
   - Implement least-privilege policies
   - Enable AWS GuardDuty
   - Conduct IAM audit

## 📚 Related Stories

- **SEC-001:** Authentication & Authorization Bypass
- **SEC-004:** Container Escape & Privilege Escalation
- **SEC-006:** Secrets Exposure & Credential Leakage
- **SRE-014:** Security Incident Response

## 🔗 References

- [AWS IAM Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
- [IRSA Documentation](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [S3 Security Best Practices](https://docs.aws.amazon.com/AmazonS3/latest/userguide/security-best-practices.html)
- [AWS Security Hub](https://docs.aws.amazon.com/securityhub/)
- [Prowler (AWS Security Tool)](https://github.com/prowler-cloud/prowler)

---

**Test File:** `internal/security/security_005_cloud_access_control_test.go`  
**Owner:** Security Team  
**Last Updated:** October 29, 2025

