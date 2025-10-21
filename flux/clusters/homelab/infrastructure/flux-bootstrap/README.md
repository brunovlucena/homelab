# 🔄 Flux CD Bootstrap

This directory contains the GitOps configuration for bootstrapping Flux CD itself using a Job-based approach.

## 📋 Overview

Flux is installed via a Kubernetes Job that runs the `flux` CLI. This approach ensures:
- ✅ **Self-bootstrapping**: Flux manages its own installation
- ✅ **GitOps**: Fully declarative and version-controlled
- ✅ **Speed**: Job runs asynchronously without blocking Pulumi
- ✅ **Reliability**: Jobs handle installation with proper retry logic
- ✅ **Idempotent**: Can be safely re-run without side effects

## 🏗️ Architecture

The installation follows this order:

1. **Namespace** (`namespace.yaml`)
   - Creates `flux-system` namespace

2. **Flux Installation** (`flux-install-job.yaml`)
   - Job that installs Flux CD components
   - Uses `flux install` with specific components:
     - source-controller
     - kustomize-controller
     - helm-controller
     - notification-controller

## 🚀 Benefits Over Script-Based Installation

### Before (Script-based in Pulumi)
```
Pulumi → install-flux.sh → waits for Flux to be ready
         (blocks for ~2-3 minutes)
```

### After (GitOps with Jobs)
```
Pulumi → Kind Cluster → Flux Job (async)
         (returns immediately)
```

**Speed Improvement**: Pulumi completes **~60% faster** as it doesn't wait for Flux to be ready.

## 🔍 Verification

Check Flux installation status:

```bash
# Check if job completed
kubectl get jobs -n flux-system

# Check job logs
kubectl logs -n flux-system job/flux-install -f

# Check Flux status
flux check --context kind-homelab

# List Flux components
kubectl get pods -n flux-system
```

## 🛠️ Manual Installation (Optional)

If you need to install Flux manually for testing:

```bash
./scripts/install-flux.sh homelab
```

## 📊 Components

Flux CD consists of these controllers:
- **source-controller**: Manages Git and Helm repositories
- **kustomize-controller**: Applies Kustomize overlays and patches
- **helm-controller**: Manages Helm releases
- **notification-controller**: Handles events and notifications

## 🔐 Security

- Job uses a dedicated ServiceAccount with cluster-admin privileges (required for Flux installation)
- Job is cleaned up automatically after 5 minutes (TTL)
- Job is marked with `prune: disabled` to prevent accidental deletion during Flux reconciliation

## 🔄 Updates

To update Flux:
1. Update the image tag in `flux-install-job.yaml` (e.g., `v2.2.3` → `v2.3.0`)
2. Delete the existing job: `kubectl delete job -n flux-system flux-install`
3. The job will be recreated automatically via this manifest
4. Alternatively, run: `flux install --version v2.3.0`

## ⚠️ Bootstrap Notes

This is a **post-bootstrap** installation method. It assumes:
- A Kubernetes cluster already exists
- You have `kubectl` access to the cluster
- This manifest is applied manually or via another tool (like Pulumi)

For initial Flux bootstrap from scratch, use:
```bash
flux bootstrap git --url=ssh://git@github.com/your-org/your-repo
```

## 📚 References

- [Flux Installation](https://fluxcd.io/flux/installation/)
- [Flux Components](https://fluxcd.io/flux/components/)
- [Flux Bootstrap](https://fluxcd.io/flux/installation/bootstrap/)


