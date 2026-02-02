# Agent Troubleshooting Guide

## Common Pod Failure Issues

### 1. ImagePullBackOff / ErrImagePull

**Symptoms:**
```
Failed to pull image "ghcr.io/brunovlucena/contract-fetcher:v1.3.0": 
rpc error: code = NotFound desc = failed to pull and unpack image: 
ghcr.io/brunovlucena/contract-fetcher:v1.3.0: not found
```

**Root Cause:**
- Studio overlays configure agents to use GHCR images (`ghcr.io/brunovlucena/...`)
- These images don't exist in GHCR yet (not built/pushed)
- Base configurations use `localhost:5001` for local development

**Solutions:**

#### Option A: Use Local Images (Recommended for Studio)
Update studio kustomization overlays to use `localhost:5001` instead of `ghcr.io`:

```yaml
patches:
  - target:
      kind: LambdaAgent
      name: contract-fetcher
    patch: |-
      - op: replace
        path: /spec/image/repository
        value: localhost:5001/agent-contracts/contract-fetcher
      - op: replace
        path: /spec/image/tag
        value: "v1.1.0"  # Use version that exists locally
```

#### Option B: Build and Push Images to GHCR
1. Build agent images locally
2. Tag with GHCR format: `ghcr.io/brunovlucena/<agent-name>:<version>`
3. Push to GHCR: `docker push ghcr.io/brunovlucena/<agent-name>:<version>`
4. Ensure `ghcr-secret` has valid credentials

#### Option C: Make Images Optional
Configure agents to scale to zero and only deploy when images are available.

### 2. CreateContainerConfigError - Missing Secrets

**Symptoms:**
```
Error: secret "restaurant-redis" not found
Error: secret "restaurant-postgres" not found
```

**Root Cause:**
- Agents require secrets that don't exist yet
- Secret creation jobs haven't run or failed

**Solutions:**

1. **Check if secret job exists:**
   ```bash
   kubectl get jobs -n <agent-namespace> | grep secret
   ```

2. **Manually create secrets:**
   ```bash
   # For restaurant-redis
   kubectl create secret generic restaurant-redis \
     --from-literal=url="redis://redis.redis.svc.cluster.local:6379" \
     -n agent-restaurant
   
   # For restaurant-postgres
   kubectl create secret generic restaurant-postgres \
     --from-literal=url="postgresql://postgres:password@postgres.postgres.svc.cluster.local:5432/homepage" \
     -n agent-restaurant
   ```

3. **Check secret job logs:**
   ```bash
   kubectl logs -n <agent-namespace> job/<secret-job-name>
   ```

### 3. FailedToRetrieveImagePullSecret

**Symptoms:**
```
Warning  FailedToRetrieveImagePullSecret  Unable to retrieve some image pull secrets (ghcr-secret)
```

**Root Cause:**
- `ghcr-secret` doesn't exist in the namespace
- Secret exists but has invalid credentials
- Secret exists but doesn't have proper permissions

**Solutions:**

1. **Check if secret exists:**
   ```bash
   kubectl get secret ghcr-secret -n <agent-namespace>
   ```

2. **Verify secret format:**
   ```bash
   kubectl get secret ghcr-secret -n <agent-namespace> -o jsonpath='{.type}'
   # Should be: kubernetes.io/dockerconfigjson
   ```

3. **Recreate secret if needed:**
   - The knative-lambda-operator should auto-create this for GHCR images
   - Check operator logs if secret creation fails

### 4. InternalError - Trigger Name Too Long

**Symptoms:**
```
Queue.rabbitmq.com "t.agent-bruno.agent-bruno-fwd-c30da233a8db8cc2448aaf374be9dabfa" is invalid: 
metadata.labels: Invalid value: "agent-bruno-fwd-contract-fetcher-io-homelab-chat-intent-projects": 
must be no more than 63 characters
```

**Root Cause:**
- Trigger names generated from event forwarding are too long
- Kubernetes labels have a 63 character limit

**Solutions:**
- Shorten event type names
- Use abbreviations in trigger names
- Update operator to truncate long names

## Quick Fixes

### Check All Failing Pods
```bash
kubectl get pods -A | grep -E "Error|CrashLoop|ImagePull|ErrImagePull|Pending"
```

### Check Specific Agent
```bash
kubectl get pods -n <agent-namespace>
kubectl describe pod <pod-name> -n <agent-namespace>
kubectl logs <pod-name> -n <agent-namespace>
```

### Force Reconcile
```bash
flux reconcile kustomization studio-07-apps -n flux-system
```

## Prevention

1. **Use localhost:5001 for studio**: Studio is local development, use local registry
2. **Create secrets jobs**: Ensure all required secrets are created before agents deploy
3. **Build images first**: Before deploying agents, ensure images exist in the registry
4. **Monitor operator logs**: Check knative-lambda-operator logs for secret creation issues
