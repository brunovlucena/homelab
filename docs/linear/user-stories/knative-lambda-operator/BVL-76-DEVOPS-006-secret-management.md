# 🔐 DEVOPS-006: Secret Management

**Priority**: P0 | **Status**: ✅ Implemented  | **Story Points**: 8
**Linear URL**: https://linear.app/bvlucena/issue/BVL-238/devops-006-secret-management

---

## 📋 User Story

**As a** DevOps Engineer  
**I want to** securely manage secrets using Sealed Secrets and External Secrets Operator  
**So that** sensitive data is never committed to Git in plaintext and follows security best practices

---

## 🎯 Acceptance Criteria

### ✅ Sealed Secrets Integration
- [ ] Sealed Secrets controller deployed
- [ ] Public/private key pair generated
- [ ] Secrets encrypted before Git commit
- [ ] Automatic decryption in cluster
- [ ] Key rotation procedures documented
- [ ] Backup and recovery strategy

### ✅ External Secrets Operator
- [ ] ESO deployed and configured
- [ ] AWS Secrets Manager integration
- [ ] Automatic secret synchronization
- [ ] Secret rotation support
- [ ] Multi-environment secret separation
- [ ] Access control via IAM roles

### ✅ Secret Types
- [ ] AWS credentials (ECR, S3)
- [ ] RabbitMQ credentials
- [ ] Database passwords
- [ ] API keys and tokens
- [ ] TLS certificates
- [ ] SSH keys

### ✅ Security Controls
- [ ] Secrets encrypted at rest
- [ ] RBAC for secret access
- [ ] Audit logging enabled
- [ ] No secrets in environment variables (use mounted files)
- [ ] Automatic secret expiration
- [ ] Secret scanning in CI/CD

### ✅ Monitoring & Alerts
- [ ] Secret sync status monitoring
- [ ] Failed decryption alerts
- [ ] Expired secret alerts
- [ ] Unauthorized access alerts
- [ ] Secret rotation tracking

---

## 🏗️ Secret Management Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                  SECRET MANAGEMENT FLOW                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. SECRET CREATION                                             │
│     Developer → kubectl create secret → secret.yaml             │
│                                                                 │
│  2. ENCRYPTION (Sealed Secrets)                                 │
│     kubeseal → Encrypt with public key → sealed-secret.yaml     │
│                                                                 │
│  3. GIT COMMIT (Safe)                                           │
│     git add sealed-secret.yaml                                  │
│     git commit -m "Add encrypted secret"                        │
│     git push origin main                                        │
│                                                                 │
│  4. GITOPS SYNC (Flux)                                          │
│     Flux → Detect change → Apply to cluster                     │
│                                                                 │
│  5. DECRYPTION (Sealed Secrets Controller)                      │
│     Sealed Secrets Controller → Decrypt with private key        │
│     → Create Kubernetes Secret                                  │
│                                                                 │
│  6. CONSUMPTION                                                 │
│     Pod → Mount secret as volume → Application reads secret     │
│                                                                 │
│  ═══════════════════════════════════════════════════════════════│
│                                                                 │
│  EXTERNAL SECRETS FLOW                                          │
│                                                                 │
│  1. SECRET STORED IN AWS                                        │
│     AWS Secrets Manager → Store sensitive data                  │
│                                                                 │
│  2. ESO CONFIGURATION                                           │
│     ExternalSecret CRD → Define secret mapping                  │
│                                                                 │
│  3. SYNC TO KUBERNETES                                          │
│     ESO → Fetch from AWS → Create K8s Secret                    │
│                                                                 │
│  4. AUTO-ROTATION                                               │
│     AWS rotates secret → ESO detects → Updates K8s Secret       │
│     → Reloads pods (if configured)                              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Technical Implementation

### 1. Install Sealed Secrets

```bash
# Install Sealed Secrets controller
helm repo add sealed-secrets https://bitnami-labs.github.io/sealed-secrets
helm install sealed-secrets sealed-secrets/sealed-secrets \
  --namespace kube-system \
  --set-string fullnameOverride=sealed-secrets-controller

# Install kubeseal CLI
brew install kubeseal

# Verify installation
kubectl get pods -n kube-system | grep sealed-secrets
```

### 2. Create and Seal Secrets

**Create plaintext secret** (never commit this!)
```bash
# Create secret file
kubectl create secret generic rabbitmq-credentials \
  --from-literal=username=admin \
  --from-literal=password=supersecret123 \
  --dry-run=client -o yaml > secret.yaml
```

**Seal the secret**
```bash
# Encrypt using cluster's public key
kubeseal --format=yaml < secret.yaml > sealed-secret.yaml

# Now safe to commit!
git add sealed-secret.yaml
git commit -m "chore: add RabbitMQ credentials (encrypted)"
git push
```

**Result**: `sealed-secret.yaml`
```yaml
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: rabbitmq-credentials
  namespace: knative-lambda
spec:
  encryptedData:
    username: AgBHj8KN4ZF...encrypted...
    password: AgCY3mK9Xp1...encrypted...
  template:
    metadata:
      name: rabbitmq-credentials
      namespace: knative-lambda
    type: Opaque
```

### 3. External Secrets Operator

**Install ESO**
```bash
# Install ESO
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets \
  external-secrets/external-secrets \
  --namespace external-secrets-system \
  --create-namespace
```

**SecretStore Configuration**
```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets-manager
  namespace: knative-lambda
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-west-2
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets-sa
```

**ExternalSecret Definition**
```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: rabbitmq-credentials
  namespace: knative-lambda
spec:
  refreshInterval: 1h  # Sync every hour
  
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  
  target:
    name: rabbitmq-credentials
    creationPolicy: Owner
  
  data:
  - secretKey: username
    remoteRef:
      key: knative-lambda/prd/rabbitmq
      property: username
  
  - secretKey: password
    remoteRef:
      key: knative-lambda/prd/rabbitmq
      property: password
```

### 4. AWS Secrets Manager Setup

```bash
# Create secret in AWS
aws secretsmanager create-secret \
  --name knative-lambda/prd/rabbitmq \
  --description "RabbitMQ credentials for production" \
  --secret-string '{
    "username": "admin",
    "password": "supersecret123"
  }' \
  --region us-west-2

# Update secret
aws secretsmanager update-secret \
  --secret-id knative-lambda/prd/rabbitmq \
  --secret-string '{
    "username": "admin",
    "password": "newsupersecret456"
  }'

# Enable automatic rotation
aws secretsmanager rotate-secret \
  --secret-id knative-lambda/prd/rabbitmq \
  --rotation-lambda-arn arn:aws:lambda:us-west-2:339954290315:function:rotate-rabbitmq \
  --rotation-rules AutomaticallyAfterDays=30
```

### 5. IAM Role for ESO

**IAM Policy** (`external-secrets-policy.json`)
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue",
        "secretsmanager:DescribeSecret"
      ],
      "Resource": [
        "arn:aws:secretsmanager:us-west-2:339954290315:secret:knative-lambda/*"
      ]
    }
  ]
}
```

**Create IAM Role**
```bash
# Create policy
aws iam create-policy \
  --policy-name ExternalSecretsPolicy \
  --policy-document file://external-secrets-policy.json

# Create IRSA (IAM Role for Service Account)
eksctl create iamserviceaccount \
  --name external-secrets-sa \
  --namespace external-secrets-system \
  --cluster homelab \
  --attach-policy-arn arn:aws:iam::339954290315:policy/ExternalSecretsPolicy \
  --approve \
  --override-existing-serviceaccounts
```

### 6. Secret Consumption in Pods

**Mount as Volume (Recommended)**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: knative-lambda-builder
spec:
  template:
    spec:
      containers:
      - name: builder
        image: knative-lambda-builder:latest
        volumeMounts:
        - name: rabbitmq-credentials
          mountPath: /etc/secrets/rabbitmq
          readOnly: true
        env:
        - name: RABBITMQ_USERNAME_FILE
          value: /etc/secrets/rabbitmq/username
        - name: RABBITMQ_PASSWORD_FILE
          value: /etc/secrets/rabbitmq/password
      
      volumes:
      - name: rabbitmq-credentials
        secret:
          secretName: rabbitmq-credentials
          defaultMode: 0400  # Read-only for owner
```

**Read in Application** (Go example)
```go
// Read secret from file
func readSecret(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(data)), nil
}

// Usage
username, err := readSecret("/etc/secrets/rabbitmq/username")
password, err := readSecret("/etc/secrets/rabbitmq/password")
```

---

## 🔑 Secret Types and Management

### 1. AWS Credentials (ECR/S3 Access)

**Using IRSA (Recommended)**
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: knative-lambda-builder
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::339954290315:role/knative-lambda-builder-prd
```

### 2. RabbitMQ Credentials

**Sealed Secret**
```bash
kubectl create secret generic rabbitmq-credentials \
  --from-literal=url="amqp://admin:password@rabbitmq:5672/" \
  --dry-run=client -o yaml | \
kubeseal --format=yaml > rabbitmq-credentials-sealed.yaml
```

### 3. Database Passwords

**External Secret from AWS**
```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: postgres-credentials
spec:
  secretStoreRef:
    name: aws-secrets-manager
  data:
  - secretKey: DATABASE_URL
    remoteRef:
      key: knative-lambda/prd/postgres
      property: connection_string
```

### 4. TLS Certificates

**Using cert-manager**
```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: knative-lambda-tls
spec:
  secretName: knative-lambda-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
  - knative-lambda.homelab
  - "*.knative-lambda.homelab"
```

---

## 🔄 Secret Rotation

### Automatic Rotation with AWS

```bash
# Lambda function for rotation (Python)
def rotate_rabbitmq_secret(event):
    secret_arn = event['SecretId']
    token = event['Token']
    step = event['Step']
    
    if step == "createSecret":
        # Generate new password
        new_password = generate_password()
        # Store pending version
        secrets_manager.put_secret_value(
            SecretId=secret_arn,
            ClientRequestToken=token,
            SecretString=json.dumps({"password": new_password}),
            VersionStages=['AWSPENDING']
        )
    
    elif step == "setSecret":
        # Update RabbitMQ with new password
        update_rabbitmq_password(new_password)
    
    elif step == "testSecret":
        # Test new credentials
        test_rabbitmq_connection(new_password)
    
    elif step == "finishSecret":
        # Mark new version as current
        secrets_manager.update_secret_version_stage(
            SecretId=secret_arn,
            VersionStage='AWSCURRENT',
            MoveToVersionId=token
        )
```

### Manual Secret Rotation

```bash
# 1. Create new secret version
kubectl create secret generic rabbitmq-credentials-new \
  --from-literal=password=newsecret \
  --dry-run=client -o yaml | \
kubeseal --format=yaml > rabbitmq-credentials-new-sealed.yaml

# 2. Apply to cluster
kubectl apply -f rabbitmq-credentials-new-sealed.yaml

# 3. Update deployment to use new secret
kubectl patch deployment knative-lambda-builder \
  --patch '{"spec":{"template":{"spec":{"volumes":[{"name":"rabbitmq-credentials","secret":{"secretName":"rabbitmq-credentials-new"}}]}}}}'

# 4. Verify rollout
kubectl rollout status deployment knative-lambda-builder

# 5. Delete old secret
kubectl delete sealedsecret rabbitmq-credentials
```

---

## 🧪 Testing Secret Management

### Test 1: Seal and Unseal

```bash
# Create test secret
echo -n "supersecret" | kubectl create secret generic test-secret \
  --from-file=password=/dev/stdin \
  --dry-run=client -o yaml > test-secret.yaml

# Seal it
kubeseal < test-secret.yaml > test-sealed.yaml

# Apply sealed secret
kubectl apply -f test-sealed.yaml

# Verify decryption
kubectl get secret test-secret -o jsonpath='{.data.password}' | base64 -d
```

**Expected**: "supersecret"

### Test 2: External Secret Sync

```bash
# Create secret in AWS
aws secretsmanager create-secret \
  --name test/secret \
  --secret-string '{"key":"value"}'

# Create ExternalSecret
kubectl apply -f - <<EOF
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: test-external-secret
spec:
  refreshInterval: 1m
  secretStoreRef:
    name: aws-secrets-manager
  data:
  - secretKey: key
    remoteRef:
      key: test/secret
      property: key
EOF

# Wait for sync
sleep 60

# Verify
kubectl get secret test-external-secret -o jsonpath='{.data.key}' | base64 -d
```

**Expected**: "value"

### Test 3: Secret Rotation

```bash
# Update AWS secret
aws secretsmanager update-secret \
  --secret-id test/secret \
  --secret-string '{"key":"newvalue"}'

# Wait for ESO to sync (refreshInterval: 1m)
sleep 70

# Verify new value
kubectl get secret test-external-secret -o jsonpath='{.data.key}' | base64 -d
```

**Expected**: "newvalue"

---

## 📊 Monitoring & Alerts

### Metrics

```promql
# External Secret sync status
external_secrets_sync_calls_total{status="success"}
external_secrets_sync_calls_total{status="error"}

# Secret age (custom metric)
time() - kube_secret_created{namespace="knative-lambda"}

# Secret access audit
kube_audit_event_count{objectRef_resource="secrets",verb="get"}
```

### Alerts

```yaml
groups:
- name: secrets
  rules:
  - alert: ExternalSecretSyncFailed
    expr: | increase(external_secrets_sync_calls_total{status="error"}[5m]) > 0
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "External Secret sync failing"
      description: "ExternalSecret {{ $labels.name }} failing to sync"
  
  - alert: SecretExpiringSoon
    expr: | (time() - kube_secret_created) > 7776000  # 90 days
    labels:
      severity: warning
    annotations:
      summary: "Secret expiring soon (> 90 days old)"
  
  - alert: SealedSecretDecryptionFailed
    expr: | increase(sealed_secrets_controller_unseal_errors_total[5m]) > 0
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "Sealed Secret decryption failed"
```

---

## 🔒 Security Best Practices

### 1. Never Commit Plaintext Secrets

```bash
# Add to .gitignore
echo "secret.yaml" >> .gitignore
echo "*.secret" >> .gitignore
echo "*-credentials.yaml" >> .gitignore
```

### 2. Use Secret Scanning

**GitHub Actions**
```yaml
- name: GitGuardian scan
  uses: GitGuardian/ggshield-action@v1
  env:
    GITHUB_PUSH_BEFORE_SHA: ${{ github.event.before }}
    GITHUB_PUSH_BASE_SHA: ${{ github.event.base }}
    GITHUB_DEFAULT_BRANCH: ${{ github.event.repository.default_branch }}
    GITGUARDIAN_API_KEY: ${{ secrets.GITGUARDIAN_API_KEY }}
```

### 3. RBAC for Secrets

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: secret-reader
  namespace: knative-lambda
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "list"]
  resourceNames: ["rabbitmq-credentials"]  # Specific secret only
```

### 4. Audit Logging

```bash
# Enable audit logging for secrets
kubectl get events --field-selector involvedObject.kind=Secret -w
```

---

## 💡 Pro Tips

1. **Key Rotation**: Rotate Sealed Secrets keys annually
2. **Backup Keys**: Store private keys in encrypted backup
3. **Least Privilege**: Grant minimal secret access via RBAC
4. **Immutable Secrets**: Use `immutable: true` for static secrets
5. **Secret Expiration**: Set TTL on secrets in AWS Secrets Manager

---

## 📈 Performance Requirements

- **Secret Decryption**: < 100ms
- **ESO Sync Interval**: 1-5 minutes
- **Secret Rotation**: < 2 minutes (zero downtime)
- **Sealed Secret Controller CPU**: < 50m
- **ESO Memory**: < 128Mi

---

## 📚 Related Documentation

- [DEVOPS-002: GitOps Deployment](DEVOPS-002-gitops-deployment.md)
- [DEVOPS-005: Infrastructure as Code](DEVOPS-005-infrastructure-as-code.md)
- Sealed Secrets: https://github.com/bitnami-labs/sealed-secrets
- External Secrets Operator: https://external-secrets.io/
- AWS Secrets Manager: https://docs.aws.amazon.com/secretsmanager/

---

**Last Updated**: October 29, 2025  
**Owner**: DevOps Team  
**Status**: Production Ready

