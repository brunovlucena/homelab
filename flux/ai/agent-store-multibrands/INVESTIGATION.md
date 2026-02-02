# 🔍 ImagePullBackOff Investigation - agent-store-multibrands

**Date**: $(date)  
**Issue**: Multiple pods in `ImagePullBackOff` state in `agent-store-multibrands` namespace

## 📊 Symptoms

- **14 pods** in `ImagePullBackOff` state
- **7 pods** in `Terminating` state  
- **0 ready pods** for all affected services
- Affected services:
  - `whatsapp-gateway`
  - `ai-seller-*` (beauty, fashion, gaming, home, tech)
  - `sales-assistant`
  - `product-catalog`
  - `order-processor`

## 🔍 Root Cause

The studio overlay (`k8s/kustomize/studio/kustomization.yaml`) was configured to use **GHCR images** (`ghcr.io/brunovlucena/*:0.3.0`), but these images **don't exist in GHCR** yet.

### Error Message
```
Failed to pull image "ghcr.io/brunovlucena/whatsapp_gateway:0.3.0": 
rpc error: code = NotFound desc = failed to pull and unpack image: 
ghcr.io/brunovlucena/whatsapp_gateway:0.3.0: not found
```

### Configuration Issue
- **Base configs**: Use `ghcr.io/brunovlucena/*` with tag `v0.1.0`
- **Studio overlay**: Patched to use `ghcr.io/brunovlucena/*` with tag `0.3.0`
- **Problem**: Images with tag `0.3.0` were never built/pushed to GHCR
- **Expected**: Studio environment should use `localhost:5001` for local development

## ✅ Solution Applied

Updated `k8s/kustomize/studio/kustomization.yaml` to use **localhost:5001** instead of **ghcr.io**:

### Changes
- Changed all image repositories from `ghcr.io/brunovlucena/{agent}` to `localhost:5001/agent-store-multibrands/{agent}`
- Kept tag `0.3.0` for consistency
- Added documentation comments explaining the change

### Image Paths Updated
- `whatsapp_gateway` → `localhost:5001/agent-store-multibrands/whatsapp_gateway:0.3.0`
- `ai_seller` → `localhost:5001/agent-store-multibrands/ai_seller:0.3.0`
- `sales_assistant` → `localhost:5001/agent-store-multibrands/sales_assistant:0.3.0`
- `product_catalog` → `localhost:5001/agent-store-multibrands/product_catalog:0.3.0`
- `order_processor` → `localhost:5001/agent-store-multibrands/order_processor:0.3.0`

## 📋 Next Steps

### 1. Build and Push Images to Local Registry
```bash
# Build images with local registry
cd /path/to/agent-store-multibrands
make build REGISTRY=localhost:5001/agent-store-multibrands

# Or manually for each agent
docker build -t localhost:5001/agent-store-multibrands/whatsapp_gateway:0.3.0 \
  -f src/whatsapp_gateway/Dockerfile src/
docker build -t localhost:5001/agent-store-multibrands/ai_seller:0.3.0 \
  -f src/ai_seller/Dockerfile src/
# ... repeat for other agents

# Push to local registry
docker push localhost:5001/agent-store-multibrands/whatsapp_gateway:0.3.0
docker push localhost:5001/agent-store-multibrands/ai_seller:0.3.0
# ... repeat for other agents
```

### 2. Apply Updated Configuration
```bash
# Apply the updated studio overlay
kubectl apply -k k8s/kustomize/studio/

# Or if using Flux
flux reconcile kustomization studio-07-apps -n flux-system
```

### 3. Verify Pods Start Successfully
```bash
# Watch pods come up
kubectl get pods -n agent-store-multibrands -w

# Check for any remaining issues
kubectl get pods -n agent-store-multibrands | grep -E "Error|ImagePull|Pending"
```

## 🔄 Alternative: Use Existing Images

If images with different tags exist locally, update the tag in the studio overlay:
```yaml
- op: replace
  path: /spec/image/tag
  value: "v0.1.0"  # or whatever version exists locally
```

## 📚 References

- [TROUBLESHOOTING.md](../../TROUBLESHOOTING.md) - Common pod failure issues
- [Local Registry Guide](../../../docs/operations/local-registry.md) - Local registry setup
- Similar fixes in other agents:
  - `agent-contracts/k8s/kustomize/studio/kustomization.yaml`
  - `agent-tools/k8s/kustomize/studio/kustomization.yaml`

## ✅ Verification Checklist

- [ ] Local registry is running (`localhost:5001`)
- [ ] Images are built and pushed to local registry
- [ ] Studio overlay updated to use `localhost:5001`
- [ ] Configuration applied to cluster
- [ ] Pods successfully pull images
- [ ] All services are running
