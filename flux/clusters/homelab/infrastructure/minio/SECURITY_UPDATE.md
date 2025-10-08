# MinIO Security Update: Sealed Secrets Implementation

## Overview
Replaced insecure hardcoded MinIO credentials with encrypted SealedSecrets for improved security.

## Changes Made

### 1. Created Sealed Secret ✅
- **File**: `sealed-secret.yaml`
- **Password**: Strong random 32-byte password (encrypted)
- **User**: `minioadmin` (encrypted)
- **Encryption**: Using sealed-secrets controller in `flux-system` namespace

### 2. Updated Setup Job ✅
- **File**: `setup-job.yaml`
- **Change**: References secret via environment variables instead of hardcoding
- **Environment Variables**:
  - `MINIO_ROOT_USER` (from secret)
  - `MINIO_ROOT_PASSWORD` (from secret)

### 3. Updated Kustomization ✅
- **File**: `kustomization.yaml`
- **Change**: References `sealed-secret.yaml` instead of `secret.yaml`

### 4. Removed Insecure Files ✅
- **Deleted**: `secret.yaml` (contained plaintext base64 encoded credentials)
- **Updated**: Documentation to reflect sealed secret usage

## Security Improvements

### Before
- Password: `minioadmin!123` (weak, hardcoded)
- Storage: Base64 encoded in Git (easily decodable)
- Visibility: Anyone with repo access could read credentials

### After
- Password: `fd5p7k9Lr+vz8OXdv35pHgVA3UCWaoe1gbKbR1dbXQU=` (strong, random)
- Storage: Encrypted with sealed-secrets public key
- Visibility: Only sealed-secrets controller can decrypt

## Verification

```bash
# Check sealed secret status
kubectl get sealedsecret minio-secret -n minio

# Verify secret was created by sealed-secrets controller
kubectl get secret minio-secret -n minio -o yaml

# Check MinIO deployment is running
kubectl get pods -n minio -l app=minio

# Verify setup job completed
kubectl logs -n minio job/minio-setup
```

## Expected Output
```
✅ MinIO pod: Running
✅ Setup job: Completed successfully
✅ Bucket: homepage-assets created
✅ Policy: Public read access set
```

## Deployment Status
- **Applied**: 2025-10-08
- **Status**: ✅ Successful
- **MinIO**: Restarted with new credentials
- **Setup Job**: Completed successfully

## Important Notes

1. **Sealed Secret Certificate**: 
   - The sealed secret was encrypted using the cluster's sealed-secrets controller
   - If you rebuild the cluster, you must backup/restore the sealed-secrets private key

2. **Password Recovery**:
   - The password is only accessible within the cluster
   - To retrieve: `kubectl get secret minio-secret -n minio -o jsonpath='{.data.MINIO_ROOT_PASSWORD}' | base64 -d`

3. **Flux Integration**:
   - Flux will automatically apply the sealed secret on cluster bootstrap
   - The sealed-secrets controller will decrypt it into a regular Kubernetes secret
   - MinIO deployment will automatically pick up the credentials

## Files Modified
```
flux/clusters/homelab/infrastructure/minio/
├── sealed-secret.yaml         (NEW - encrypted credentials)
├── setup-job.yaml            (MODIFIED - uses secret env vars)
├── kustomization.yaml        (MODIFIED - references sealed-secret.yaml)
└── secret.yaml               (DELETED - insecure)
```

## Related Documentation
- Homepage deployment: `../homepage/FINAL_DEPLOYMENT_STATUS.md`
- Sealed secrets setup: `../sealed-secrets/`

