# ⚙️ DevOps Engineer - Knative Lambda

**Infrastructure automation and deployment excellence**

---

## 🎯 Overview

As a DevOps engineer working with Knative Lambda, you're responsible for infrastructure automation, CI/CD pipelines, GitOps workflows, and ensuring zero-downtime deployments. This guide covers deployment strategies, infrastructure as code, and operational best practices.

---

## 🚀 Quick Start

### 1. Deploy to Environment

```bash
# Set environment
export ENV=dev  # or prd

# Deploy via Flux (GitOps)
flux reconcile kustomization knative-lambda

# Verify deployment
kubectl get deployment knative-lambda-builder -n knative-lambda
kubectl get ksvc -n knative-lambda
```

### 2. Common DevOps Tasks

```bash
# Build and push all images
make build-and-push-all-${ENV}

# Update Helm values
vim deploy/overlays/${ENV}/values-${ENV}.yaml
git commit -am "chore: update ${ENV} config"
git push

# Monitor deployment
kubectl rollout status deployment/knative-lambda-builder -n knative-lambda

# Check deployment health
kubectl get pods -n knative-lambda
make rabbitmq-status ENV=${ENV}
```

### 3. Infrastructure Commands

```bash
# Check infrastructure status
kubectl get nodes
kubectl top nodes
kubectl get namespaces | grep knative-lambda

# View ArgoCD/Flux sync status
kubectl get kustomization -n flux-system
kubectl get helmrelease -n flux-system

# Access dashboards
make pf-argocd        # Port-forward ArgoCD
make pf-prometheus    # Port-forward Prometheus
```

---

## 📊 Infrastructure Overview

### Components

| Component | Purpose | Technology | Replicas |
|-----------|---------|------------|----------|
| **Builder Service** | CloudEvent processor | Go + Chi | 2-10 (HPA) |
| **RabbitMQ** | Event queue | RabbitMQ Operator | 3 (HA) |
| **Knative Serving** | Serverless platform | Knative | Cluster-wide |
| **Kaniko Jobs** | Container builds | Kaniko | 0-100 (dynamic) |
| **Prometheus** | Metrics collection | Prometheus Operator | 2 (HA) |
| **Flux CD** | GitOps sync | Flux v2 | Cluster-wide |

### Network Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    NETWORK ARCHITECTURE                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  External                                                       │
│     │                                                           │
│     ├─→ ALB/Ingress (TLS)                                       │
│     │      ├─→ Builder API (:8080)                              │
│     │      └─→ Knative Functions (*.knative-lambda.homelab)     │
│     │                                                           │
│  Internal                                                       │
│     │                                                           │
│     ├─→ RabbitMQ (5672, 15672)                                  │
│     │      └─→ Queues: build-events, service-events, results   │
│     │                                                           │
│     ├─→ Knative Broker (HTTP)                                   │
│     │      └─→ Event routing to functions                       │
│     │                                                           │
│     ├─→ ECR (Pull/Push)                                         │
│     │      └─→ Image registry: knative-lambdas/*                │
│     │                                                           │
│     └─→ S3 (HTTPS)                                              │
│            └─→ Parser files: knative-lambda-fusion-*/    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Deployment Strategies

### 1. GitOps with Flux

**How it works**:
1. Developer commits code → GitHub
2. CI builds Docker image → ECR
3. Update Helm chart version → Git
4. Flux detects change → Applies to cluster
5. Knative performs rolling update

**Flux Kustomization**:
```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: knative-lambda
  namespace: flux-system
spec:
  interval: 5m
  path: ./flux/clusters/homelab/infrastructure/knative-lambda
  prune: true
  sourceRef:
    kind: GitRepository
    name: homelab
  healthChecks:
  - apiVersion: apps/v1
    kind: Deployment
    name: knative-lambda-builder
    namespace: knative-lambda
```

### 2. Canary Deployment (with Flagger)

**Progressive rollout**:
```yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: knative-lambda-builder
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: knative-lambda-builder
  progressDeadlineSeconds: 600
  service:
    port: 8080
  analysis:
    interval: 1m
    threshold: 5
    maxWeight: 50
    stepWeight: 10
    metrics:
    - name: request-success-rate
      thresholdRange:
        min: 99
    - name: request-duration
      thresholdRange:
        max: 500
```

### 3. Blue-Green Deployment

**Manual approval workflow**:
```bash
# 1. Deploy new version as "green"
kubectl apply -f deploy/green-deployment.yaml

# 2. Test green deployment
curl https://green.knative-lambda.homelab/health

# 3. Switch traffic (update Service selector)
kubectl patch svc knative-lambda-builder \
  -p '{"spec":{"selector":{"version":"green"}}}'

# 4. Verify traffic switch
kubectl get endpoints knative-lambda-builder

# 5. Delete old "blue" deployment
kubectl delete deployment knative-lambda-builder-blue
```

---

## 📚 User Stories

| Story ID | Title | Priority | Status | Story Points |
|----------|-------|----------|--------|--------------|
| **DEVOPS-001** | [Observability Setup](user-stories/DEVOPS-001-observability-setup.md) | P0 | ✅ | 8 |
| **DEVOPS-002** | [GitOps Deployment](user-stories/DEVOPS-002-gitops-deployment.md) | P0 | ✅ | 13 |
| **DEVOPS-003** | [Multi-Environment Management](user-stories/DEVOPS-003-multi-environment.md) | P0 | ✅ | 8 |
| **DEVOPS-004** | [CI/CD Pipeline](user-stories/DEVOPS-004-cicd-pipeline.md) | P1 | ✅ | 13 |
| **DEVOPS-005** | [Infrastructure as Code](user-stories/DEVOPS-005-infrastructure-as-code.md) | P1 | ✅ | 8 |
| **DEVOPS-006** | [Secret Management](user-stories/DEVOPS-006-secret-management.md) | P0 | ✅ | 8 |
| **DEVOPS-007** | [Cost Optimization](user-stories/DEVOPS-007-cost-optimization.md) | P1 | ✅ | 5 |
| **DEVOPS-008** | [Disaster Recovery Automation](user-stories/DEVOPS-008-dr-automation.md) | P0 | ✅ | 13 |
| **DEVOPS-009** | [Multi-Registry Support](user-stories/DEVOPS-009-multi-registry.md) | P1 | 📋 | 8 |
| **DEVOPS-010** | [Multi-Storage Backend Support](user-stories/DEVOPS-010-multi-storage.md) | P1 | 📋 | 8 |

→ **[View All User Stories](user-stories/README.md)**

---

## 🔐 Security & Compliance

### RBAC Configuration

```yaml
# ServiceAccount for builder
apiVersion: v1
kind: ServiceAccount
metadata:
  name: knative-lambda-builder
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::339954290315:role/knative-lambda-builder-prd

---
# Role for job management
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: knative-lambda-builder
rules:
- apiGroups: ["batch"]
  resources: ["jobs"]
  verbs: ["create", "get", "list", "delete"]
- apiGroups: ["serving.knative.dev"]
  resources: ["services"]
  verbs: ["create", "get", "list", "update", "delete"]
```

### Secret Management

```bash
# Sealed Secrets (GitOps-friendly)
kubeseal --format=yaml < secret.yaml > sealed-secret.yaml
kubectl apply -f sealed-secret.yaml

# External Secrets Operator
kubectl create secret generic rabbitmq-credentials \
  --from-literal=username=admin \
  --from-literal=password=${RABBITMQ_PASSWORD}
```

---

## 💰 Cost Tracking

### Current Monthly Costs

| Resource | Cost | Optimization Opportunity |
|----------|------|--------------------------|
| EC2 Nodes (build) | $450 | Use Spot (save 60%) |
| ECR Storage | $50 | Lifecycle policy (save $20) |
| RabbitMQ | $35 | Right-size (save $18) |
| Data Transfer | $25 | Use VPC endpoints (save $10) |
| **Total** | **$560** | **Potential: $412/month** |

---

## 📈 Performance Metrics

### Deployment Metrics

| Metric | Target | Current |
|--------|--------|---------|
| Deployment Duration | <5min | 3.5min ✅ |
| Rollback Time | <2min | 1.5min ✅ |
| Zero Downtime | 100% | 100% ✅ |
| Build Success Rate | >95% | 97% ✅ |

---

## 🎓 Learning Resources

### Internal Docs
- [Makefile Reference](../../../Makefile) - All automation commands
- [Helm Chart](../../../deploy/) - Infrastructure as code
- [Architecture](../../04-architecture/) - System design

### External Resources
- [Knative Docs](https://knative.dev/docs/)
- [Flux CD Guide](https://fluxcd.io/flux/)
- [Helm Best Practices](https://helm.sh/docs/chart_best_practices/)

---

**Need help?** Join `#knative-lambda` on Slack or file a GitHub issue.

