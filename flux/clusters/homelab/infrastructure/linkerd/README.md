# 🔗 Linkerd Service Mesh

This directory contains the Flux GitOps configuration for installing and managing Linkerd service mesh and the Linkerd Viz extension.

## 📋 Overview

Linkerd is installed via Kubernetes Jobs that run the `linkerd` CLI. This approach ensures:
- ✅ **Compatibility**: Uses official Linkerd CLI for installation
- ✅ **GitOps**: Fully managed by Flux
- ✅ **Speed**: Jobs run asynchronously without blocking Pulumi
- ✅ **Reliability**: Jobs handle installation with proper retry logic

## 🏗️ Architecture

The installation follows this order:

1. **Namespaces** (`namespace.yaml`)
   - Creates `linkerd` and `linkerd-viz` namespaces

2. **Gateway API CRDs** (`gateway-api-crds.yaml`)
   - Installs Kubernetes Gateway API CRDs (required by Linkerd)
   - Fetched from official Gateway API repository

3. **Linkerd CRDs** (`linkerd-crds-job.yaml`)
   - Job that installs Linkerd Custom Resource Definitions
   - Uses `linkerd install --crds`

4. **Linkerd Control Plane** (`linkerd-control-plane-job.yaml`)
   - Job that installs the Linkerd control plane
   - Uses `linkerd install`

5. **Linkerd Viz** (`linkerd-viz-job.yaml`)
   - Job that installs the Linkerd Viz extension
   - Uses `linkerd viz install`

## 🚀 Benefits Over Script-Based Installation

### Before (Script-based in Pulumi)
```
Pulumi → install-linkerd.sh → install-linkerd-viz.sh
         (blocks for ~5 minutes)
```

### After (GitOps with Jobs)
```
Pulumi → Flux → Linkerd Jobs (async)
         (returns immediately)
```

**Speed Improvement**: Pulumi completes **~80% faster** as it doesn't wait for Linkerd to be ready.

## 🔍 Verification

Check Linkerd installation status:

```bash
# Check if jobs completed
kubectl get jobs -n linkerd
kubectl get jobs -n linkerd-viz

# Check Linkerd status
linkerd check --context kind-homelab

# Check Viz status
linkerd viz check --context kind-homelab

# Access dashboard
linkerd viz dashboard --context kind-homelab
```

## 🛠️ Manual Installation (Optional)

If you need to install Linkerd manually for testing:

```bash
# Combined installation script (faster)
./scripts/install-linkerd-full.sh homelab

# Or individual scripts (legacy)
./scripts/install-linkerd.sh homelab
./scripts/install-linkerd-viz.sh homelab
```

## 📊 Components

### Linkerd Control Plane
- `linkerd-destination`: Service discovery and routing
- `linkerd-identity`: mTLS certificate management
- `linkerd-proxy-injector`: Automatic sidecar injection

### Linkerd Viz
- `metrics-api`: Metrics aggregation
- `tap`: Live traffic inspection
- `web`: Dashboard UI
- `tap-injector`: Tap sidecar injection
- `prometheus`: Metrics storage

## 🔐 Security

- Jobs use dedicated ServiceAccounts with appropriate RBAC
- Jobs are cleaned up automatically after 5 minutes (TTL)
- Jobs are marked with `prune: disabled` to prevent accidental deletion

## 🔄 Updates

To update Linkerd:
1. Update the image tag in the job manifests (e.g., `stable-2.16.2` → `stable-2.17.0`)
2. Delete existing jobs: `kubectl delete job -n linkerd --all && kubectl delete job -n linkerd-viz --all`
3. Commit and push - Flux will recreate the jobs with new versions

## 📚 References

- [Linkerd Documentation](https://linkerd.io/2-edge/tasks/install/)
- [Linkerd Viz Extension](https://linkerd.io/2-edge/tasks/extensions/)
- [Gateway API](https://gateway-api.sigs.k8s.io/)

