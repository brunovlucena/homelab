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

## 🌐 Multi-Cluster Configuration

### Studio ↔ Pro Cluster Linking

The Linkerd multicluster extension enables service discovery and secure mTLS communication between `studio` and `pro` clusters.

**Network Requirements**:
- ✅ Both clusters are on the same VPN network
- ✅ Non-overlapping IP ranges:
  - Studio: Pods `10.246.0.0/16`, Services `10.98.0.0/16`
  - Pro: Pods `10.247.0.0/16`, Services `10.99.0.0/16`
- ✅ Linkerd multicluster extension installed on both clusters
- ✅ Shared trust anchor (see below)

**Linking Clusters**:

The link job (`linkerd-multicluster-link-job.yaml`) automatically detects which cluster it's running on and attempts to link to the other. However, since it runs in-cluster, it may not have access to the target cluster's kubeconfig.

To complete the link manually from a machine with access to both clusters:

```bash
# Link studio → pro
linkerd multicluster link \
  --context kind-studio \
  --cluster-name pro \
  --target-context kind-pro

# Link pro → studio (bidirectional)
linkerd multicluster link \
  --context kind-pro \
  --cluster-name studio \
  --target-context kind-studio
```

**Verification**:

```bash
# Check gateway pods are running on both clusters
kubectl get pods -n linkerd-multicluster -l component=gateway --context kind-studio
kubectl get pods -n linkerd-multicluster -l component=gateway --context kind-pro

# Check ServiceMirror resources
kubectl get servicemirror -n linkerd-multicluster --context kind-studio
kubectl get servicemirror -n linkerd-multicluster --context kind-pro

# Check multicluster status
linkerd multicluster check --context kind-studio
linkerd multicluster check --context kind-pro
```

**Exporting Services for Cross-Cluster Access**:

To make a service accessible from another cluster, add the export annotation:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service
  namespace: default
  annotations:
    multicluster.linkerd.io/export: "true"  # Enable cross-cluster access
spec:
  # ... service spec
```

Services exported from the target cluster will appear in the source cluster with the `.remote` suffix:
- `my-service.default.svc.cluster.local` (local)
- `my-service.default.svc.pro.remote` (from pro cluster)
- `my-service.default.svc.studio.remote` (from studio cluster)

## 🔐 Security & Trust Anchor Management

### Shared Trust Anchor (Multi-Cluster)

All clusters share the **same Linkerd trust anchor** (root CA certificate) stored in GitHub repository secrets:
- ✅ Enables secure mTLS between clusters
- ✅ Required for Linkerd multi-cluster service discovery
- ✅ Automatically synced by External Secrets Operator (ESO)

**Architecture**: GitHub Repository Secrets → ESO → `linkerd-trust-anchor` K8s Secret → Linkerd Installation

**Management**:
```bash
# Extract trust anchor from cluster with Linkerd installed
./scripts/mac/manage-linkerd-trust-anchors.sh extract homelab

# Store in GitHub repository secrets
# Add LINKERD_TRUST_ANCHOR_CA_CRT to repository settings → Secrets
# ESO will automatically sync to all clusters

# Or use GitHub CLI
gh secret set LINKERD_TRUST_ANCHOR_CA_CRT < ca.crt
```

**Verification**:
```bash
# Verify trust anchor exists in cluster
kubectl get secret linkerd-trust-anchor -n linkerd

# Verify trust anchors match across clusters
./scripts/mac/manage-linkerd-trust-anchors.sh verify-all
```

📚 **Full Documentation**: [/docs/linkerd-trust-anchor-distribution.md](../../../../docs/linkerd-trust-anchor-distribution.md)

### Job Security

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

