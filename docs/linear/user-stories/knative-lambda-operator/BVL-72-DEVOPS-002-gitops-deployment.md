# 🔄 DEVOPS-002: GitOps Deployment

**Priority**: P1 | **Status**: ✅ Implemented K  | **Story Points**: 8
**Linear URL**: https://linear.app/bvlucena/issue/BVL-234/devops-002-gitops-deployment


---

## 📋 User Story

**As a** DevOps Engineer  
**I want to** deploy all infrastructure via GitOps  
**So that** deployments are auditable, repeatable, and self-healing

---

## 🎯 Acceptance Criteria

- [ ] All infrastructure defined in Git
- [ ] Flux automatically syncs changes
- [ ] Rollback via Git revert
- [ ] Environment promotion (dev → staging → prod)
- [ ] Drift detection and auto-remediation
- [ ] Deployment notifications to Slack

---

## 🔄 GitOps Workflow

```
┌────────────────────────────────────────────────────────────────┐
│                    GITOPS WORKFLOW                             │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  1. DEVELOPER COMMITS                                          │
│     Developer → git commit → git push origin main              │
│                                                                │
│  2. FLUX DETECTS CHANGE                                        │
│     Flux (5min interval) → Poll Git repository                 │
│     └─ Detect new commit → SHA abc123                          │
│                                                                │
│  3. FLUX APPLIES MANIFESTS                                     │
│     Flux → kubectl apply -f manifests/                         │
│     ├─ Namespace                                               │
│     ├─ ConfigMap                                               │
│     ├─ Secret (Sealed)                                         │
│     ├─ Deployment                                              │
│     ├─ Service                                                 │
│     └─ Ingress                                                 │
│                                                                │
│  4. KUBERNETES RECONCILES                                      │
│     Deployment Controller → Rolling update                     │
│     ├─ Create new ReplicaSet                                   │
│     ├─ Scale up new pods                                       │
│     ├─ Wait for readiness                                      │
│     ├─ Scale down old pods                                     │
│     └─ Terminate old ReplicaSet                                │
│                                                                │
│  5. FLUX HEALTH CHECK                                          │
│     Flux → Check resource health                               │
│     ├─ Deployment: Available replicas = desired                │
│     ├─ Service: Endpoints exist                                │
│     └─ Ingress: Ready                                          │
│                                                                │
│  6. NOTIFICATION                                               │
│     Flux → Slack webhook → #deployments                        │
│     Message: "✅ knative-lambda deployed (v1.2.3)"             │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## 📁 Repository Structure

```
flux/clusters/homelab/infrastructure/knative-lambda/
├── base/                           # Shared resources
│   ├── namespace.yaml
│   ├── serviceaccount.yaml
│   └── rbac.yaml
│
├── overlays/
│   ├── dev/
│   │   ├── kustomization.yaml     # Dev overrides
│   │   └── values-dev.yaml        # Dev Helm values
│   │
│   ├── staging/
│   │   ├── kustomization.yaml
│   │   └── values-staging.yaml
│   │
│   └── prd/
│       ├── kustomization.yaml
│       └── values-prd.yaml        # Prod Helm values
│
└── flux-system/
    ├── kustomization-dev.yaml     # Flux Kustomization for dev
    ├── kustomization-staging.yaml
    └── kustomization-prd.yaml     # Flux Kustomization for prod
```

---

## 🚀 Deployment Process

### 1. Initial Setup

```bash
# Bootstrap Flux on cluster
flux bootstrap github \
  --owner=brunolucena \
  --repository=homelab \
  --branch=main \
  --path=flux/clusters/homelab \
  --personal

# Verify Flux installation
flux check
kubectl get kustomization -n flux-system
```

### 2. Deploy Application

```bash
# Commit infrastructure changes
git add flux/clusters/homelab/infrastructure/knative-lambda/
git commit -m "feat: deploy knative-lambda v1.2.3"
git push origin main

# Trigger immediate reconciliation (optional)
flux reconcile kustomization knative-lambda
```

### 3. Monitor Deployment

```bash
# Watch Flux sync status
flux get kustomizations --watch

# Check deployment status
kubectl rollout status deployment/knative-lambda-builder -n knative-lambda

# View Flux logs
flux logs --level=info
```

### 4. Rollback (if needed)

```bash
# Git revert
git revert HEAD
git push origin main

# Or: Flux suspend + manual rollback
flux suspend kustomization knative-lambda
kubectl rollout undo deployment/knative-lambda-builder -n knative-lambda
flux resume kustomization knative-lambda
```

---

## 🔐 Secrets Management

### Using Sealed Secrets

```bash
# Create secret
kubectl create secret generic rabbitmq-credentials \
  --from-literal=username=admin \
  --from-literal=password=supersecret \
  --dry-run=client -o yaml > secret.yaml

# Seal the secret
kubeseal --format=yaml < secret.yaml > sealed-secret.yaml

# Commit sealed secret (safe for Git)
git add sealed-secret.yaml
git commit -m "chore: add RabbitMQ credentials"
git push
```

---

## 💡 Pro Tips

- **Sync interval**: 5min is good balance (responsiveness vs. API load)
- **Health checks**: Always define for critical resources
- **Drift detection**: Flux auto-corrects manual `kubectl` changes
- **Notifications**: Send to Slack for visibility
- **Image automation**: Use Flux image automation for auto-updates

---

## 📈 Performance Requirements

- **Git Sync Interval**: 5 minutes
- **Deployment Time**: < 5 minutes
- **Rollback Time**: < 2 minutes
- **Drift Detection**: < 1 minute
- **Health Check Duration**: < 30 seconds

---

## 📚 Related Documentation

- [DEVOPS-001: Observability Setup](DEVOPS-001-observability-setup.md)
- [DEVOPS-003: Multi-Environment Management](DEVOPS-003-multi-environment.md)
- [DEVOPS-005: Infrastructure as Code](DEVOPS-005-infrastructure-as-code.md)
- Flux CD Documentation: https://fluxcd.io/flux/
- GitOps Principles: https://opengitops.dev/

---

**Last Updated**: October 29, 2025  
**Owner**: DevOps Team  
**Status**: ✅ Implemented K
