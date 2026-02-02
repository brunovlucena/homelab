# 🔄 DEVOPS-003: Multi-Environment Management

**Priority**: P0 | **Status**: ✅ Implemented  | **Story Points**: 8
**Linear URL**: https://linear.app/bvlucena/issue/BVL-235/devops-003-multi-environment-management

---

## 📋 User Story

**As a** DevOps Engineer  
**I want to** manage multiple isolated environments (dev, staging, prod)  
**So that** we can safely test changes before production deployment and maintain environment parity

---

## 🎯 Acceptance Criteria

### ✅ Environment Isolation
- [ ] Separate Kubernetes namespaces per environment
- [ ] Isolated AWS resources (ECR repos, S3 buckets, IAM roles)
- [ ] Network policies preventing cross-environment traffic
- [ ] Resource quotas per environment
- [ ] Separate RabbitMQ virtual hosts per environment

### ✅ Configuration Management
- [ ] Environment-specific Helm values files
- [ ] Kustomize overlays for environment customization
- [ ] ConfigMap/Secret separation per environment
- [ ] Feature flags for gradual rollout
- [ ] Environment variable inheritance

### ✅ Promotion Strategy
- [ ] Automated promotion: dev → staging
- [ ] Manual approval: staging → prod
- [ ] Smoke tests before promotion
- [ ] Rollback procedures for each environment
- [ ] Version tracking across environments

### ✅ Access Control
- [ ] RBAC policies per environment
- [ ] Separate service accounts
- [ ] Audit logging per environment
- [ ] Developer access limited to dev/staging
- [ ] Production changes require approval

### ✅ Monitoring & Observability
- [ ] Per-environment Grafana dashboards
- [ ] Separate Prometheus instances or federation
- [ ] Environment labels on all metrics
- [ ] Environment-specific alert routing
- [ ] Cost tracking per environment

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                  MULTI-ENVIRONMENT ARCHITECTURE                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  KUBERNETES CLUSTER                                             │
│  ├─ Namespace: knative-lambda                                   │
│  │  ├─ Builder Deployment (replicas: 1)                         │
│  │  ├─ RabbitMQ (dev vhost)                                     │
│  │  ├─ Resource Quota: 4 CPU, 8Gi RAM                           │
│  │  └─ ECR: 339954290315.dkr.ecr.us-west-2/knative-lambdas-dev  │
│  │                                                              │
│  ├─ Namespace: knative-lambda                                   │
│  │  ├─ Builder Deployment (replicas: 2, HPA enabled)            │
│  │  ├─ RabbitMQ (staging vhost, HA mode)                        │
│  │  ├─ Resource Quota: 8 CPU, 16Gi RAM                          │
│  │  └─ ECR: 339954290315.dkr.ecr.us-west-2/knative-lambdas-stg  │
│  │                                                              │
│  └─ Namespace: knative-lambda                                   │
│     ├─ Builder Deployment (replicas: 3, HPA + PDB)              │
│     ├─ RabbitMQ (prod vhost, HA + mirrored queues)              │
│     ├─ Resource Quota: 16 CPU, 32Gi RAM                         │
│     └─ ECR: 339954290315.dkr.ecr.us-west-2/knative-lambdas-prd  │
│                                                                 │
│  AWS RESOURCES                                                  │
│  ├─ S3 Buckets                                                  │
│  │  ├─ knative-lambda-fusion-code                               │
│  │  ├─ knative-lambda-fusion-code                               │
│  │  └─ knative-lambda-fusion-code                               │
│  │                                                              │
│  ├─ IAM Roles                                                   │
│  │  ├─ knative-lambda-builder-dev                               │
│  │  ├─ knative-lambda-builder-staging                           │
│  │  └─ knative-lambda-builder-prd                               │
│  │                                                              │
│  └─ ECR Repositories                                            │
│     ├─ knative-lambdas-dev/* (lifecycle: 7 days)                │
│     ├─ knative-lambdas-staging/* (lifecycle: 30 days)           │
│     └─ knative-lambdas-prd/* (lifecycle: 90 days)               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Technical Implementation

### Repository Structure

```
flux/clusters/homelab/infrastructure/knative-lambda/
├── base/                              # Shared base resources
│   ├── namespace.yaml
│   ├── serviceaccount.yaml
│   ├── rbac.yaml
│   ├── deployment.yaml               # Base deployment template
│   └── service.yaml
│
├── overlays/
│   ├── dev/
│   │   ├── kustomization.yaml        # Dev customizations
│   │   ├── values-dev.yaml           # Dev Helm values
│   │   ├── resource-quota.yaml       # Dev quota: 4 CPU, 8Gi
│   │   └── network-policy.yaml       # Dev network restrictions
│   │
│   ├── staging/
│   │   ├── kustomization.yaml
│   │   ├── values-staging.yaml
│   │   ├── resource-quota.yaml       # Staging quota: 8 CPU, 16Gi
│   │   ├── hpa.yaml                  # HorizontalPodAutoscaler
│   │   └── network-policy.yaml
│   │
│   └── prd/
│       ├── kustomization.yaml
│       ├── values-prd.yaml
│       ├── resource-quota.yaml       # Prod quota: 16 CPU, 32Gi
│       ├── hpa.yaml
│       ├── pdb.yaml                  # PodDisruptionBudget
│       └── network-policy.yaml
│
└── environments/
    ├── dev-config.yaml               # Dev environment config
    ├── staging-config.yaml           # Staging environment config
    └── prd-config.yaml               # Prod environment config
```

### Environment-Specific Kustomization

**File**: `overlays/dev/kustomization.yaml`
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: knative-lambda

# Base resources
resources:
- ../../base
- namespace.yaml
- resource-quota.yaml
- network-policy.yaml

# Common labels
commonLabels:
  environment: dev
  managed-by: flux

# Image transformations
images:
- name: knative-lambda-builder
  newName: 339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-builder
  newTag: dev-latest

# ConfigMap generator
configMapGenerator:
- name: knative-lambda-config
  envs:
  - config.env

# Secret generator (sealed secrets)
secretGenerator:
- name: aws-credentials
  files:
  - credentials=aws-credentials.txt

# Patches
patches:
- path: deployment-patch.yaml
  target:
    kind: Deployment
    name: knative-lambda-builder

# Replicas override
replicas:
- name: knative-lambda-builder
  count: 1
```

**File**: `overlays/dev/deployment-patch.yaml`
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
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        env:
        - name: ENVIRONMENT
          value: "dev"
        - name: LOG_LEVEL
          value: "debug"
        - name: MAX_CONCURRENT_JOBS
          value: "5"
```

### Environment-Specific Helm Values

**File**: `overlays/dev/values-dev.yaml`
```yaml
environment: dev

replicaCount: 1

image:
  repository: 339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-builder
  tag: dev-latest
  pullPolicy: Always

resources:
  requests:
    memory: 512Mi
    cpu: 250m
  limits:
    memory: 1Gi
    cpu: 500m

autoscaling:
  enabled: false

serviceAccount:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::339954290315:role/knative-lambda-builder-dev

config:
  environment: dev
  logLevel: debug
  maxConcurrentJobs: 5
  s3Bucket: knative-lambda-fusion-code
  ecrRegistry: 339954290315.dkr.ecr.us-west-2.amazonaws.com
  ecrRepository: knative-lambdas-dev

rabbitmq:
  host: rabbitmq.rabbitmq-system.svc.cluster.local
  vhost: /knative-lambda
  queue: build-events-dev

monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    namespace: prometheus
```

**File**: `overlays/staging/values-staging.yaml`
```yaml
environment: staging

replicaCount: 2

image:
  repository: 339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-builder
  tag: staging-v1.2.3  # Promoted from dev
  pullPolicy: IfNotPresent

resources:
  requests:
    memory: 1Gi
    cpu: 500m
  limits:
    memory: 2Gi
    cpu: 1000m

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5
  targetCPUUtilizationPercentage: 70

serviceAccount:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::339954290315:role/knative-lambda-builder-staging

config:
  environment: staging
  logLevel: info
  maxConcurrentJobs: 10
  s3Bucket: knative-lambda-fusion-code
  ecrRegistry: 339954290315.dkr.ecr.us-west-2.amazonaws.com
  ecrRepository: knative-lambdas-staging

rabbitmq:
  host: rabbitmq.rabbitmq-system.svc.cluster.local
  vhost: /knative-lambda
  queue: build-events-staging

monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    namespace: prometheus
```

**File**: `overlays/prd/values-prd.yaml`
```yaml
environment: prd

replicaCount: 3

image:
  repository: 339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambda-builder
  tag: prd-v1.2.3  # Promoted from staging
  pullPolicy: IfNotPresent

resources:
  requests:
    memory: 2Gi
    cpu: 1000m
  limits:
    memory: 4Gi
    cpu: 2000m

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 60

podDisruptionBudget:
  enabled: true
  minAvailable: 2

serviceAccount:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::339954290315:role/knative-lambda-builder-prd

config:
  environment: prd
  logLevel: warn
  maxConcurrentJobs: 20
  s3Bucket: knative-lambda-fusion-code
  ecrRegistry: 339954290315.dkr.ecr.us-west-2.amazonaws.com
  ecrRepository: knative-lambdas-prd

rabbitmq:
  host: rabbitmq.rabbitmq-system.svc.cluster.local
  vhost: /knative-lambda
  queue: build-events-prd

monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    namespace: prometheus
  alerts:
    enabled: true
    severity: critical
```

### Resource Quotas

**File**: `overlays/dev/resource-quota.yaml`
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: knative-lambda-quota
  namespace: knative-lambda
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "8"
    limits.memory: 16Gi
    pods: "20"
    services: "5"
    persistentvolumeclaims: "2"
```

**File**: `overlays/prd/resource-quota.yaml`
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: knative-lambda-quota
  namespace: knative-lambda
spec:
  hard:
    requests.cpu: "16"
    requests.memory: 32Gi
    limits.cpu: "32"
    limits.memory: 64Gi
    pods: "100"
    services: "20"
    persistentvolumeclaims: "10"
```

### Network Policies

**File**: `overlays/prd/network-policy.yaml`
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: knative-lambda-network-policy
  namespace: knative-lambda
spec:
  podSelector:
    matchLabels:
      app: knative-lambda-builder
  policyTypes:
  - Ingress
  - Egress
  
  ingress:
  # Allow traffic from Knative Serving
  - from:
    - namespaceSelector:
        matchLabels:
          name: knative-serving
    ports:
    - protocol: TCP
      port: 8080
  
  # Allow Prometheus scraping
  - from:
    - namespaceSelector:
        matchLabels:
          name: prometheus
    ports:
    - protocol: TCP
      port: 8080
  
  egress:
  # Allow DNS
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: UDP
      port: 53
  
  # Allow RabbitMQ
  - to:
    - namespaceSelector:
        matchLabels:
          name: rabbitmq-system
    ports:
    - protocol: TCP
      port: 5672
  
  # Allow Kubernetes API
  - to:
    - namespaceSelector: {}
      podSelector:
        matchLabels:
          component: apiserver
    ports:
    - protocol: TCP
      port: 443
  
  # Allow AWS services (S3, ECR)
  - to:
    - podSelector: {}
    ports:
    - protocol: TCP
      port: 443
```

---

## 🚀 Deployment Commands

### Deploy to Specific Environment

```bash
# Deploy to dev
export ENV=dev
flux reconcile kustomization knative-lambda --with-source

# Deploy to staging
export ENV=staging
flux reconcile kustomization knative-lambda --with-source

# Deploy to prod (requires approval)
export ENV=prd
flux reconcile kustomization knative-lambda --with-source
```

### Environment Promotion Workflow

```bash
# 1. Deploy to dev and test
make deploy-dev
make test-integration ENV=dev

# 2. Promote to staging (automated)
./scripts/promote-to-staging.sh

# 3. Run staging tests
make test-e2e ENV=staging
make test-load ENV=staging

# 4. Promote to prod (manual approval required)
./scripts/promote-to-prod.sh --require-approval
```

### Promotion Script

**File**: `scripts/promote-to-staging.sh`
```bash
#!/bin/bash
set -euo pipefail

# Get current dev version
DEV_TAG=$(kubectl get deployment knative-lambda-builder \
  -n knative-lambda \
  -o jsonpath='{.spec.template.spec.containers[0].image}' | cut -d: -f2)

echo "📦 Promoting dev tag: ${DEV_TAG}"

# Update staging values
yq eval ".image.tag = \"${DEV_TAG}\"" \
  -i overlays/staging/values-staging.yaml

# Commit and push
git add overlays/staging/values-staging.yaml
git commit -m "chore: promote ${DEV_TAG} to staging"
git push origin main

echo "✅ Promotion to staging initiated"
echo "🔄 Flux will sync in ~5 minutes"
```

---

## 🧪 Testing Multi-Environment Setup

### Test 1: Environment Isolation

```bash
# Deploy to all environments
make deploy-all-envs

# Verify namespace isolation
kubectl get all -n knative-lambda
kubectl get all -n knative-lambda
kubectl get all -n knative-lambda

# Test network policy (should fail)
kubectl run test-pod --image=curlimages/curl -it --rm \
  -n knative-lambda -- \
  curl http://knative-lambda-builder.knative-lambda:8080/health
```

**Expected**: Network policy blocks cross-environment traffic

### Test 2: Resource Quotas

```bash
# Check quota usage
kubectl describe resourcequota -n knative-lambda

# Try to exceed quota (should fail)
kubectl scale deployment knative-lambda-builder \
  --replicas=100 -n knative-lambda
```

**Expected**: Quota exceeded error

### Test 3: Configuration Differences

```bash
# Compare environment configs
diff <(kubectl get deployment knative-lambda-builder \
  -n knative-lambda -o yaml) \
  <(kubectl get deployment knative-lambda-builder \
  -n knative-lambda -o yaml)
```

**Expected**: Different resource limits, replicas, log levels

---

## 📊 Monitoring & Observability

### Environment-Specific Dashboards

```json
{
  "dashboard": {
    "title": "Knative Lambda - Multi-Environment Overview",
    "panels": [
      {
        "title": "Deployments by Environment",
        "targets": [
          {
            "expr": "up{job=\"knative-lambda-builder\"}",
            "legendFormat": "{{environment}}"
          }
        ]
      },
      {
        "title": "Build Success Rate by Environment",
        "targets": [
          {
            "expr": "sum by (environment) (rate(builds_total{status=\"success\"}[5m])) / sum by (environment) (rate(builds_total[5m]))"
          }
        ]
      },
      {
        "title": "Resource Usage by Environment",
        "targets": [
          {
            "expr": "sum by (environment) (container_memory_usage_bytes{namespace=~\"knative-lambda-.*\"})"
          }
        ]
      }
    ]
  }
}
```

### Metrics with Environment Labels

```promql
# Builds per environment
sum by (environment) (rate(builds_total[5m]))

# P95 build duration by environment
histogram_quantile(0.95, 
  sum by (environment, le) (rate(build_duration_seconds_bucket[5m]))
)

# Cost per environment (estimated)
sum by (environment) (
  rate(container_cpu_usage_seconds_total{namespace=~"knative-lambda-.*"}[1h]) * 0.04
)
```

---

## 💰 Cost Tracking

### Monthly Cost by Environment | Environment | Compute | Storage | Data Transfer | Total | |------------- | --------- | --------- | --------------- | ------- | | **Dev** | $80 | $10 | $5 | $95 | | **Staging** | $150 | $20 | $10 | $180 | | **Prod** | $320 | $40 | $15 | $375 | | **Total** | $550 | $70 | $30 | **$650** | ---

## 📈 Performance Requirements

- **Environment Deployment Time**: < 5 minutes
- **Promotion Time (dev → staging)**: < 3 minutes
- **Promotion Time (staging → prod)**: < 5 minutes
- **Rollback Time**: < 2 minutes per environment
- **Configuration Drift Detection**: < 1 minute

---

## 🔐 RBAC Policies

### Developer Access (Dev/Staging Only)

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: developers
  namespace: knative-lambda
subjects:
- kind: Group
  name: developers
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: edit
  apiGroup: rbac.authorization.k8s.io
```

### Production Access (SRE Only)

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: sre-team
  namespace: knative-lambda
subjects:
- kind: Group
  name: sre-team
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: admin
  apiGroup: rbac.authorization.k8s.io
```

---

## 📚 Related Documentation

- [DEVOPS-002: GitOps Deployment](DEVOPS-002-gitops-deployment.md)
- [DEVOPS-005: Infrastructure as Code](DEVOPS-005-infrastructure-as-code.md)
- [DEVOPS-006: Secret Management](DEVOPS-006-secret-management.md)
- Kustomize Documentation: https://kustomize.io/
- Kubernetes Network Policies: https://kubernetes.io/docs/concepts/services-networking/network-policies/

---

**Last Updated**: October 29, 2025  
**Owner**: DevOps Team  
**Status**: Production Ready

