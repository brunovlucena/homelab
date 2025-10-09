# Loki MinIO Secret

## Current Status
✅ **Loki is running!** The secret `loki-minio-secret` has been created directly in the cluster.

## Create SealedSecret for GitOps

To create the SealedSecret and manage it via GitOps, run:

```bash
# From the homelab repo root
./scripts/create-loki-secret.sh
```

If the script fails due to connectivity issues with sealed-secrets, you can use this manual approach:

```bash
# 1. Port-forward to sealed-secrets
kubectl port-forward -n flux-system svc/sealed-secrets 8080:8080 &
PF_PID=$!

# 2. Wait for port-forward to be ready
sleep 2

# 3. Create and seal the secret
echo -n "supersecretpassword" | kubectl create secret generic loki-minio-secret \
  --dry-run=client \
  --from-file=root-password=/dev/stdin \
  --namespace=loki \
  -o yaml | \
  kubeseal --format=yaml --cert=<(curl -s http://localhost:8080/v1/cert.pem) > \
  flux/clusters/homelab/infrastructure/sealed-secrets/loki-minio-secret.yaml

# 4. Kill port-forward
kill $PF_PID

# 5. Add to kustomization
# Edit flux/clusters/homelab/infrastructure/sealed-secrets/kustomization.yaml
# and add: - loki-minio-secret.yaml
```

## Alternative: Use existing certificate

If you have the sealed-secrets certificate already saved:

```bash
kubectl create secret generic loki-minio-secret \
  --from-literal=root-password=supersecretpassword \
  --namespace=loki \
  --dry-run=client -o yaml | \
  kubeseal --format=yaml --cert=/path/to/cert.pem > \
  flux/clusters/homelab/infrastructure/sealed-secrets/loki-minio-secret.yaml
```

## Current Secret Details
- **Namespace**: `loki`
- **Name**: `loki-minio-secret`
- **Key**: `root-password`
- **Value**: `supersecretpassword` (default homelab password)

