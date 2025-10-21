# 🚨 Runbook: MinIO Authentication Issues

## Alert Information

**Alert Name:** `MinIOAuthenticationFailure`  
**Severity:** High  
**Component:** minio  
**Service:** object-storage

## Symptom

Users or applications cannot authenticate to MinIO. Seeing 403 Forbidden or 401 Unauthorized errors.

## Impact

- **User Impact:** HIGH - Cannot access object storage
- **Business Impact:** HIGH - Applications fail to read/write data
- **Data Impact:** NONE - Data is safe but inaccessible

## Diagnosis

### 1. Check Authentication Errors in Logs

```bash
kubectl logs -n minio -l app=minio --tail=200 | grep -i "auth\|forbidden\|unauthorized\|403\|401"
```

**Common patterns:**
- "Access Denied"
- "Invalid access key"
- "Signature does not match"
- "Request has expired"

### 2. Verify Root Credentials

```bash
# Check root user credentials
kubectl get secret -n minio minio -o jsonpath='{.data.rootUser}' | base64 -d
kubectl get secret -n minio minio -o jsonpath='{.data.rootPassword}' | base64 -d
```

### 3. Test Root Authentication

```bash
# Port forward to MinIO
kubectl port-forward -n minio svc/minio 9000:9000

# Test with root credentials
mc alias set local http://localhost:9000 <root-user> <root-password>
mc admin info local
```

### 4. Check User/Service Account

```bash
# List users
mc admin user list local

# Check specific user
mc admin user info local <username>

# List service accounts
mc admin user svcacct list local <username>
```

### 5. Check IAM Policies

```bash
# List policies
mc admin policy list local

# Check user's policy
mc admin policy info local <policy-name>

# Check which policy is attached to user
mc admin user info local <username> | grep Policy
```

### 6. Check Bucket Policies

```bash
# Check bucket policy
mc policy get local/<bucket-name>

# Check bucket versioning (affects access)
mc version info local/<bucket-name>
```

## Resolution

### Scenario A: Root Credentials Incorrect

**Likely Cause:** Password changed or secret corrupted

**Steps:**
1. Check if secret exists:
   ```bash
   kubectl get secret -n minio minio
   ```

2. If secret missing, check Flux source:
   ```bash
   # Check if secret is in sealed-secrets or external-secrets
   kubectl get sealedsecrets -n minio
   kubectl get externalsecrets -n minio
   ```

3. Recreate secret if needed:
   ```bash
   # Generate new credentials
   ROOT_USER="admin"
   ROOT_PASSWORD=$(openssl rand -base64 32)
   
   kubectl create secret generic minio -n minio \
     --from-literal=rootUser=$ROOT_USER \
     --from-literal=rootPassword=$ROOT_PASSWORD \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

4. Restart MinIO to pick up new credentials:
   ```bash
   kubectl rollout restart deployment/minio -n minio
   ```

5. Update client configurations with new credentials

### Scenario B: Service Account Credentials Invalid

**Likely Cause:** Service account deleted or credentials rotated

**Steps:**
1. Check if service account exists:
   ```bash
   mc admin user svcacct list local <parent-user>
   ```

2. If missing, create new service account:
   ```bash
   mc admin user svcacct add local <parent-user> \
     --access-key <access-key> \
     --secret-key <secret-key>
   ```

3. Or let MinIO generate credentials:
   ```bash
   mc admin user svcacct add local <parent-user>
   # Note the generated access-key and secret-key
   ```

4. Update application configuration:
   ```bash
   kubectl create secret generic app-minio-creds -n <app-namespace> \
     --from-literal=access-key=<access-key> \
     --from-literal=secret-key=<secret-key> \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

5. Restart application:
   ```bash
   kubectl rollout restart deployment/<app-name> -n <app-namespace>
   ```

### Scenario C: Insufficient Permissions

**Likely Cause:** User/service account lacks required IAM policy

**Steps:**
1. Check current policy:
   ```bash
   mc admin user info local <username>
   ```

2. Create appropriate policy if missing:
   ```bash
   # Create read-only policy
   cat > readonly-policy.json <<EOF
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "s3:GetObject",
           "s3:ListBucket"
         ],
         "Resource": [
           "arn:aws:s3:::<bucket-name>/*",
           "arn:aws:s3:::<bucket-name>"
         ]
       }
     ]
   }
   EOF
   
   mc admin policy create local readonly-policy readonly-policy.json
   ```

3. Or use built-in policies:
   ```bash
   # Available built-in policies:
   # - readonly: read-only access to all buckets
   # - readwrite: read-write access to all buckets
   # - writeonly: write-only access
   # - diagnostics: admin diagnostics
   # - consoleAdmin: full admin access
   
   mc admin policy attach local readwrite --user <username>
   ```

4. Verify policy attached:
   ```bash
   mc admin user info local <username>
   ```

### Scenario D: Bucket Policy Blocking Access

**Likely Cause:** Bucket policy is too restrictive

**Steps:**
1. Check current bucket policy:
   ```bash
   mc policy get local/<bucket-name>
   ```

2. Set appropriate bucket policy:
   ```bash
   # Make bucket private
   mc policy set private local/<bucket-name>
   
   # Make bucket public for downloads
   mc policy set download local/<bucket-name>
   
   # Make bucket public for uploads
   mc policy set upload local/<bucket-name>
   
   # Make bucket fully public
   mc policy set public local/<bucket-name>
   ```

3. Or set custom bucket policy:
   ```bash
   cat > bucket-policy.json <<EOF
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Principal": {"AWS": ["arn:aws:iam::<account-id>:user/<username>"]},
         "Action": [
           "s3:GetObject",
           "s3:PutObject"
         ],
         "Resource": ["arn:aws:s3:::<bucket-name>/*"]
       }
     ]
   }
   EOF
   
   mc policy set-json bucket-policy.json local/<bucket-name>
   ```

### Scenario E: Signature Mismatch

**Likely Cause:** Clock skew or incorrect signature calculation

**Steps:**
1. Check time synchronization on client:
   ```bash
   # On client pod/machine
   date
   # Compare with MinIO server time
   kubectl exec -n minio $(kubectl get pod -n minio -l app=minio -o jsonpath='{.items[0].metadata.name}') -- date
   ```

2. If time difference >15 minutes, fix NTP:
   ```bash
   # On client machine/pod
   ntpdate -s time.nist.gov
   # Or
   timedatectl set-ntp true
   ```

3. Check client SDK configuration:
   - Verify access key and secret key are correct
   - Ensure no extra spaces in credentials
   - Check URL format (http vs https)

4. Test with curl to isolate issue:
   ```bash
   # Simple GET request
   curl -v \
     --aws-sigv4 "aws:amz:us-east-1:s3" \
     --user "$ACCESS_KEY:$SECRET_KEY" \
     "http://minio.minio.svc.cluster.local:9000/<bucket-name>/"
   ```

### Scenario F: Request Expired

**Likely Cause:** Pre-signed URL expired or clock skew

**Steps:**
1. Check if using pre-signed URLs:
   ```python
   # Generate new pre-signed URL with longer expiry
   from minio import Minio
   
   client = Minio(
       "minio.minio.svc.cluster.local:9000",
       access_key="access-key",
       secret_key="secret-key",
       secure=False
   )
   
   # Generate URL valid for 7 days
   url = client.presigned_get_object(
       "bucket-name",
       "object-name",
       expires=timedelta(days=7)
   )
   ```

2. Fix clock skew (see Scenario E)

3. Regenerate expired URLs

### Scenario G: LDAP/AD Integration Issues

**Likely Cause:** External authentication provider unavailable

**Steps:**
1. Check LDAP configuration:
   ```bash
   mc admin config get local identity_ldap
   ```

2. Test LDAP connectivity:
   ```bash
   kubectl run -it --rm ldaptest --image=nicolaka/netshoot --restart=Never -- sh
   # Inside pod:
   ldapsearch -x -H ldap://<ldap-server> -b "<base-dn>"
   ```

3. Check LDAP credentials:
   ```bash
   mc admin config set local identity_ldap \
     server_addr=<ldap-server>:389 \
     lookup_bind_dn=<bind-dn> \
     lookup_bind_password=<password> \
     user_dn_search_base_dn=<base-dn> \
     user_dn_search_filter="(uid=%s)"
   
   mc admin service restart local
   ```

4. Check logs for LDAP errors:
   ```bash
   kubectl logs -n minio -l app=minio --tail=100 | grep -i ldap
   ```

## Verification

### 1. Test Root Access

```bash
mc alias set local http://localhost:9000 <root-user> <root-password>
mc admin info local
```

Should return MinIO server information.

### 2. Test User/Service Account Access

```bash
mc alias set test http://localhost:9000 <access-key> <secret-key>
mc ls test/
```

Should list accessible buckets.

### 3. Test Bucket Operations

```bash
# Upload test
echo "test" > test.txt
mc cp test.txt test/<bucket-name>/test.txt

# Download test
mc cp test/<bucket-name>/test.txt test-download.txt

# Cleanup
mc rm test/<bucket-name>/test.txt
rm test.txt test-download.txt
```

### 4. Check Application Logs

```bash
kubectl logs -n <app-namespace> -l app=<app-name> --tail=50
# Should not show auth errors
```

## Prevention

1. **Use Kubernetes secrets for credentials:**
   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: minio-creds
     namespace: app-namespace
   stringData:
     access-key: "access-key-here"
     secret-key: "secret-key-here"
   ```

2. **Implement proper IAM policies:**
   - Principle of least privilege
   - Separate service accounts per application
   - Regular policy audits

3. **Set up monitoring:**
   ```yaml
   - alert: MinIOAuthenticationFailures
     expr: |
       rate(minio_s3_requests_4xx_errors_total{code="403"}[5m]) > 10
     for: 5m
     annotations:
       summary: "High rate of authentication failures"
   ```

4. **Enable audit logging:**
   ```yaml
   environment:
     MINIO_AUDIT_LOGGER_ENABLED: "on"
     MINIO_AUDIT_WEBHOOK_ENABLE: "on"
     MINIO_AUDIT_WEBHOOK_ENDPOINT: "http://logging-service:9000/audit"
   ```

5. **Use service accounts instead of user credentials:**
   - Better security
   - Easier credential rotation
   - Scoped permissions

6. **Implement credential rotation:**
   ```bash
   # Rotate service account credentials regularly
   mc admin user svcacct edit local <access-key> --secret-key <new-secret-key>
   ```

7. **Document IAM structure:**
   - Which users/service accounts exist
   - What permissions they have
   - Which applications use which credentials

8. **Time synchronization:**
   - Ensure NTP is configured on all nodes
   - Monitor clock skew

## Related Alerts

- `MinIODown`
- `MinIOHighErrorRate`
- `MinIOBucketAccessDenied`
- `MinIOLDAPAuthFailure`

## Escalation

**When to escalate:**
- Unable to access with root credentials
- LDAP/AD integration completely broken
- Suspected security breach
- Credential rotation causing widespread failures

**Escalation Path:**
1. Senior SRE Team
2. Security Team (if breach suspected)
3. Identity Management Team (for LDAP/AD issues)
4. MinIO Vendor Support

## Additional Resources

- [MinIO Identity and Access Management](https://min.io/docs/minio/linux/administration/identity-access-management.html)
- [MinIO Policy Management](https://min.io/docs/minio/linux/administration/identity-access-management/policy-based-access-control.html)
- [MinIO Service Accounts](https://min.io/docs/minio/linux/administration/identity-access-management/minio-user-management.html#service-accounts)
- [MinIO LDAP Integration](https://min.io/docs/minio/linux/operations/external-iam/configure-ad-ldap-external-identity-management.html)
- Internal Wiki: IAM Guidelines
- Slack: #sre-security

