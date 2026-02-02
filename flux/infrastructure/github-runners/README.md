# 🏃 GitHub Actions Self-Hosted Runners (ARC)

Self-hosted GitHub Actions runners using the **official GitHub-supported [Actions Runner Controller (ARC)](https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners-with-actions-runner-controller/about-actions-runner-controller)** with runner scale sets, deployed via Flux GitOps.

## 📋 Overview

This setup deploys self-hosted GitHub Actions runners on your Kubernetes homelab cluster using the **new GitHub-supported ARC architecture**. The runners use **runner scale sets** that automatically scale based on workflow demand.

> **Note**: This uses the new GitHub-supported ARC (runner scale sets), not the old community-supported summerwind/actions-runner-controller.

## 🏗️ Architecture

```
GitHub Actions Workflow
    ↓
Runner Scale Set Listener (in cluster)
    ↓
Scale Set Controller
    ↓
Ephemeral Runner Pods (auto-scaled)
```

**Key Components:**
- **Controller** (`arc-controller`): Manages CRDs and coordinates runner scale sets
- **Runner Scale Set** (`arc-runners`): Manages the actual runner pods with auto-scaling
- **Ephemeral Runners**: Pods that run jobs and are destroyed after completion

## 🔧 Components

### 1. Controller HelmRelease (`controller-helmrelease.yaml`)
- **Chart**: `gha-runner-scale-set-controller`
- **Version**: `0.9.3`
- **Namespace**: `github-runners`
- **Features**:
  - ✅ Automatically installs CRDs
  - ✅ Prometheus metrics enabled
  - ✅ ServiceMonitor for monitoring

### 2. Runner Scale Set HelmRelease (`runner-helmrelease.yaml`)
- **Chart**: `gha-runner-scale-set`
- **Version**: `0.9.3`
- **Repository**: `brunovlucena/homelab`
- **Scaling**: 1-5 runners (auto-scaled)
- **Mode**: Kubernetes mode (more secure than Docker-in-Docker)
- **Labels**: `self-hosted`, `linux`, `x64`, `homelab`

## 🚀 Deployment (Fully Automated via GitOps)

### Prerequisites

The secret `github-token` must exist in the `github-runners` namespace. All secrets in this homelab are managed via **External Secrets Operator** (ESO).

The GitHub token needs the following permissions:
- `repo` scope (for repository-level runners)
- `admin:org` scope (for organization-level runners)

Store the token in GitHub repository secrets and create an ExternalSecret resource to pull it into the cluster.

### Automatic Deployment

Everything is deployed automatically by Flux:

1. **Flux reconciles** the `phase4-apps` kustomization
2. **Controller deploys** and installs CRDs automatically
3. **Runner scale set deploys** (waits for controller via `dependsOn`)
4. **Runners register** to GitHub automatically
5. **Auto-scaling** begins based on job demand

### Manual Reconciliation (Optional)

Force reconciliation if needed:

```bash
flux reconcile kustomization phase4-apps -n flux-system
```

## 📊 Monitoring

### Check Deployment Status

```bash
# Check HelmReleases
kubectl get helmrelease -n github-runners

# Check controller pods
kubectl get pods -n github-runners -l app.kubernetes.io/name=gha-runner-scale-set-controller

# Check runner pods
kubectl get pods -n github-runners -l actions.github.com/scale-set-name=homelab-runners

# Check runner scale set status
kubectl get runners -n github-runners
```

### Prometheus Metrics

Metrics are automatically exposed on port `:8080` at `/metrics`:
- Controller metrics: `gha_controller_*`
- Listener metrics: `gha_assigned_jobs`, `gha_running_jobs`, `gha_registered_runners`, etc.

See Grafana dashboards for visualization.

## 📝 Usage in Workflows

Use the self-hosted runners in your GitHub Actions workflows:

```yaml
name: CI/CD
on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: [self-hosted, linux, x64, homelab]
    steps:
      - uses: actions/checkout@v4
      - name: Run tests
        run: make test
```

## ⚙️ Configuration

### Adjust Scaling Limits

Edit `runner-helmrelease.yaml`:

```yaml
values:
  minRunners: 2      # Change minimum
  maxRunners: 10     # Change maximum
```

### Change Resource Limits

Edit `runner-helmrelease.yaml`:

```yaml
template:
  spec:
    resources:
      limits:
        cpu: "8"
        memory: 16Gi
      requests:
        cpu: "2"
        memory: 4Gi
```

### Use Organization-Level Runners

Edit `runner-helmrelease.yaml`:

```yaml
values:
  githubConfigUrl: https://github.com/your-org  # Change from repo to org
```

## 🔍 Troubleshooting

### Runners Not Appearing in GitHub

```bash
# Check controller logs
kubectl logs -n github-runners -l app.kubernetes.io/name=gha-runner-scale-set-controller

# Check listener logs
kubectl logs -n github-runners -l actions.github.com/scale-set-name=homelab-runners

# Verify secret exists
kubectl get secret github-token -n github-runners

# Check secret data (base64 encoded)
kubectl get secret github-token -n github-runners -o jsonpath='{.data}'
```

### Phase4-apps Failing to Reconcile

```bash
# Check kustomization status
kubectl get kustomization phase4-apps -n flux-system -o yaml

# Check for CRD issues
kubectl get crds | grep actions.github.com

# Force reconcile
flux reconcile kustomization phase4-apps -n flux-system
```

### Runners Not Scaling

```bash
# Check runner scale set status
kubectl describe runners -n github-runners

# Check listener metrics
kubectl port-forward -n github-runners svc/arc-runners-listener 8080:8080
curl http://localhost:8080/metrics | grep gha_
```

## 🔄 Upgrading

Upgrades are handled automatically by Flux when you update the chart version in the HelmRelease files.

To upgrade manually:
1. Update `version:` in `controller-helmrelease.yaml`
2. Update `version:` in `runner-helmrelease.yaml`
3. Commit and push
4. Flux will automatically upgrade

## 📚 Resources

- [GitHub ARC Documentation](https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners-with-actions-runner-controller)
- [ARC Quickstart](https://docs.github.com/en/actions/tutorials/use-actions-runner-controller/quickstart)
- [Deploy Runner Scale Sets](https://docs.github.com/en/actions/tutorials/use-actions-runner-controller/deploy-runner-scale-sets)
- [ARC GitHub Repository](https://github.com/actions/actions-runner-controller)

## 🔒 Security Notes

⚠️ **Important Security Considerations:**

1. **Kubernetes Mode vs Docker-in-Docker**: This setup uses Kubernetes mode (more secure)
2. **Ephemeral Runners**: Each runner is destroyed after job completion
3. **Private Repositories**: Only use with private repos to prevent malicious code
4. **Token Security**: Rotate GitHub PAT regularly
5. **Namespace Isolation**: Runners run in dedicated `github-runners` namespace
6. **Resource Limits**: Always set resource limits to prevent exhaustion

## 📊 Resource Usage

- **Controller**: ~100m CPU, 128Mi memory
- **Per Runner**: ~1 CPU, 2Gi memory (configurable)
- **Total (5 runners max)**: ~5.1 CPUs, ~10.1Gi memory

## 🆕 What's New in GitHub-Supported ARC

**vs Old Community-Supported Version:**

| Feature | Old (summerwind) | New (GitHub) |
|---------|------------------|--------------|
| CRD Management | Manual | ✅ Automatic via Helm |
| Scaling | RunnerDeployment | ✅ Runner Scale Sets |
| Architecture | Polling | ✅ Listener-based |
| Support | Community | ✅ Official GitHub |
| Metrics | Limited | ✅ Comprehensive |
| Container Mode | Docker-in-Docker | ✅ Kubernetes mode |
| Ephemeral | Optional | ✅ Default |
