# 🔐 Sealed Secrets

Bitnami Sealed Secrets controller for encrypting Kubernetes secrets in Git.

## What This Does

The sealed-secrets controller:
- Runs in the `sealed-secrets` namespace
- Generates a 4096-bit RSA key pair on first startup
- Decrypts `SealedSecret` resources into regular Kubernetes `Secret` resources
- Allows safe storage of encrypted secrets in Git

## Directory Structure

```
flux/infrastructure/bootstrap/sealed-secrets/
├── secrets/                          # All sealed secrets (centralized)
│   ├── kustomization.yaml           
│   ├── cloudflare-secrets-sealed.yaml
│   ├── grafana-secrets-sealed.yaml
│   └── README.md
├── kustomization.yaml               # References secrets/ (commented during bootstrap)
├── helmrelease.yaml                 # Controller deployment
└── namespace.yaml
```

## Installation (Bootstrap Workflow)

### Phase 1: Deploy Controller (Secrets Commented Out)

Ensure `kustomization.yaml` has secrets commented out:

```yaml
resources:
  - namespace.yaml
  - helmrelease.yaml
  # - secrets  # ← Keep commented during initial bootstrap
```

Then deploy:

```bash
# 1. Commit and push the sealed-secrets infrastructure
git add flux/infrastructure/bootstrap/sealed-secrets/
git commit -m "feat: add sealed-secrets controller"
git push

# 2. Wait for Flux to reconcile (or force it)
flux reconcile kustomization pro-01-core --with-source

# 3. Verify controller is running
kubectl get pods -n sealed-secrets
```

### Phase 2: Enable Secrets (After Controller is Ready)

## Creating Sealed Secrets

### Step 1: Create Plain Secrets

Plain secret files should be created in each infrastructure directory (they're gitignored):
- `flux/infrastructure/kube-prometheus-stack/grafana-secrets.yaml` 
- `flux/infrastructure/cloudflare-warp/cloudflare-secrets.yaml` 

These files contain the actual secret values from your `.zshrc`.

### Step 2: Run the Sealing Script

```bash
cd ~/workspace/bruno/repos/homelab
./scripts/mac/sealed-secrets-seal.sh
# OR simply:
make seal-secrets
```

This script will:
1. Fetch the public key from the controller
2. Encrypt each plain secret file
3. **Consolidate** all sealed secrets in `flux/infrastructure/bootstrap/sealed-secrets/secrets/`
4. Update the centralized kustomization
5. Delete the plain secret files
6. Clean up the public key

### Step 3: Uncomment the Secrets Directory

Edit `flux/infrastructure/bootstrap/sealed-secrets/kustomization.yaml`:

```yaml
resources:
  - namespace.yaml
  - helmrelease.yaml
  - secrets  # ← Uncomment this line
```

This single change enables **all** sealed secrets at once! 🎉

### Step 4: Commit and Push

```bash
git add flux/infrastructure/bootstrap/sealed-secrets/
git commit -m "feat: enable sealed secrets"
git push
```

### Benefits of Centralized Approach

✅ **Easy Bootstrap**: Comment/uncomment one line to control all secrets  
✅ **Single Location**: All sealed secrets in one directory  
✅ **Clean Infrastructure**: No secret files scattered across infrastructure dirs  
✅ **Simple Management**: Easy to see all secrets at a glance

## Secrets Included

### Prometheus Namespace (Grafana)
- `PAGERDUTY_SERVICE_KEY` - PagerDuty integration
- `GRAFANA_PASSWORD` - Admin password
- `GRAFANA_API_KEY` - API token
- `SLACK_WEBHOOK_URL` - Alert webhook

### Cloudflare-WARP Namespace
- `CLOUDFLARE_WARP_TOKEN` - WARP connector token
- `CLOUDFLARE_API_KEY` - API key
- `CLOUDFLARE_EMAIL` - Account email
- `CLOUDFLARE_PRO_TUNNEL_TOKEN` - Tunnel token

## Verification

```bash
# Check SealedSecrets
kubectl get sealedsecrets -A

# Check that Secrets were created
kubectl get secrets -n prometheus grafana-secrets
kubectl get secrets -n cloudflare-warp cloudflare-secrets

# View secret keys (not values)
kubectl describe secret -n prometheus grafana-secrets
```

## Backup and Restore

**IMPORTANT:** Back up the controller's private key!

```bash
# Backup
./scripts/mac/sealed-secrets-backup.sh
# OR: make backup-sealed-secrets ENV=pro

# Restore (if cluster is recreated)
./scripts/mac/sealed-secrets-restore.sh
# OR: make restore-sealed-secrets ENV=pro
```

## How It Works

```
Developer (kubeseal)
    │
    ├─► Encrypts with PUBLIC key
    │
    ▼
SealedSecret (in Git) ✅ Safe to commit
    │
    ├─► Applied to cluster
    │
    ▼
sealed-secrets-controller
    │
    ├─► Decrypts with PRIVATE key
    │
    ▼
Kubernetes Secret (in cluster only)
```

## Security

✅ **Safe to commit:**
- Sealed secrets (`*-sealed.yaml`)
- Public key (pub-sealed-secrets.pem)

❌ **NEVER commit:**
- Plain secrets (`*-secrets.yaml`)
- Controller's private key

## Troubleshooting

### Controller not starting

```bash
kubectl logs -n flux-system -l app.kubernetes.io/name=sealed-secrets
kubectl describe helmrelease -n flux-system sealed-secrets
```

### Cannot decrypt error

The controller may not have started yet or the key was lost. Check:
```bash
kubectl get secrets -n flux-system | grep sealed-secrets-key
```

If missing, you need to restore from backup or re-seal all secrets with a new key.

## References

- [Flux Guide](https://fluxcd.io/flux/guides/sealed-secrets/)
- [GitHub](https://github.com/bitnami-labs/sealed-secrets)
- [Docs](https://sealed-secrets.netlify.app/)

