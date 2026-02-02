# Deployment Guide

Complete guide to deploying the Vaultwarden password manager in your homelab.

## Prerequisites

- Kubernetes cluster
- kubectl configured
- Flux CD installed
- HashiCorp Vault (we'll deploy this)
- Cert-manager for TLS certificates
- Ingress controller (NGINX)

## Step 1: Deploy HashiCorp Vault

First, deploy Vault to your cluster:

```bash
# Apply Vault manifests
kubectl apply -k flux/infrastructure/vault

# Wait for Vault to be ready
kubectl wait --for=condition=ready pod -l app=vault -n vault-system --timeout=300s

# Initialize Vault (dev mode - for production, use proper initialization)
kubectl exec -it deployment/vault -n vault-system -- vault auth -methods

# Enable KV secrets engine
kubectl exec -it deployment/vault -n vault-system -- vault secrets enable -path=secret kv-v2
```

### Production Vault Setup

For production, configure Vault properly:

1. Initialize with unseal keys
2. Enable Kubernetes authentication
3. Create policies for vaultwarden service
4. Configure auto-unseal (optional)

## Step 2: Configure Vault Secrets

Create secrets for the vaultwarden API:

```bash
# Generate JWT secret
JWT_SECRET=$(openssl rand -hex 32)

# Create secret (use Sealed Secrets for GitOps)
kubectl create secret generic vaultwarden-secrets \
  --from-literal=jwt-secret="$JWT_SECRET" \
  --namespace vaultwarden \
  --dry-run=client -o yaml | kubeseal -o yaml > flux/services/vaultwarden/k8s/base/secret-sealed.yaml
```

## Step 3: Deploy Vaultwarden API

```bash
# Build and push Docker image (if using local registry)
cd flux/services/vaultwarden/backend
docker build -t localhost:5001/vaultwarden-api:latest .
docker push localhost:5001/vaultwarden-api:latest

# Apply Kubernetes manifests
kubectl apply -k flux/services/vaultwarden/k8s/base

# Or with Flux
flux create kustomization vaultwarden \
  --source=flux-system \
  --path=./flux/services/vaultwarden/k8s/base \
  --prune=true \
  --interval=5m
```

## Step 4: Verify Deployment

```bash
# Check pods
kubectl get pods -n vaultwarden

# Check service
kubectl get svc -n vaultwarden

# Check ingress
kubectl get ingress -n vaultwarden

# Test API
curl https://vaultwarden.lucena.cloud/health
```

## Step 5: Create First User

The first user needs to be created manually or via a setup script:

```bash
# Login to Vault
export VAULT_ADDR="http://vault.vault-system.svc.cluster.local:8200"
export VAULT_TOKEN="root"  # Use proper token in production

# Create user authentication data
vault kv put secret/vaultwarden/auth/user@example.com \
  user_id="user-123" \
  password_hash="$(bcrypt_hash 'your-password')" \
  email="user@example.com"

# Create user profile
vault kv put secret/vaultwarden/users/user-123/profile \
  email="user@example.com" \
  name="User Name"
```

## Step 6: Build and Install Clients

### iOS App

```bash
cd flux/services/vaultwarden/ios
open Vaultwarden.xcodeproj

# In Xcode:
# 1. Configure signing
# 2. Set API URL in APIService.swift
# 3. Build and run on device/simulator
```

### Browser Extensions

#### Chrome

1. Open Chrome → Extensions
2. Enable Developer mode
3. Click "Load unpacked"
4. Select `flux/services/vaultwarden/browser-extension/chrome/`
5. Configure server URL in extension options

#### Firefox

1. Open Firefox → about:debugging
2. Click "This Firefox"
3. Click "Load Temporary Add-on"
4. Select `manifest.json` in `flux/services/vaultwarden/browser-extension/firefox/`

#### Safari

1. Open `flux/services/vaultwarden/browser-extension/safari/VaultwardenExtension.xcodeproj`
2. Build in Xcode
3. Enable extension in Safari → Preferences → Extensions

## Security Considerations

### Production Checklist

- [ ] Use proper Vault initialization (not dev mode)
- [ ] Configure Vault auto-unseal
- [ ] Use Kubernetes service accounts for Vault auth
- [ ] Enable TLS for Vault
- [ ] Rotate JWT secrets regularly
- [ ] Enable audit logging in Vault
- [ ] Use Sealed Secrets for all Kubernetes secrets
- [ ] Enable network policies
- [ ] Configure proper RBAC
- [ ] Set up monitoring and alerts

### Encryption

- All passwords are encrypted client-side before being sent to the server
- Server only stores encrypted blobs
- Vault provides additional encryption at rest
- TLS encryption in transit

## Troubleshooting

### Vault Connection Issues

```bash
# Check Vault is accessible
kubectl exec -it deployment/vaultwarden-api -n vaultwarden -- wget -O- http://vault.vault-system.svc.cluster.local:8200/v1/sys/health

# Check Vault logs
kubectl logs -n vault-system deployment/vault
```

### API Issues

```bash
# Check API logs
kubectl logs -n vaultwarden deployment/vaultwarden-api

# Test API endpoint
kubectl port-forward -n vaultwarden svc/vaultwarden-api 8080:80
curl http://localhost:8080/health
```

### Authentication Issues

- Verify JWT secret is set correctly
- Check token expiration
- Verify user exists in Vault
- Check password hash format

## Monitoring

Set up monitoring for:

- API response times
- Vault access patterns
- Failed authentication attempts
- Disk usage (Vault storage)
- Certificate expiration

## Backup

Regularly backup:

- Vault data (if using persistent storage)
- Vault unseal keys
- Vault configuration
- Kubernetes secrets (via Sealed Secrets backup)

## Updates

To update the deployment:

```bash
# Update image
docker build -t localhost:5001/vaultwarden-api:v1.1.0 .
docker push localhost:5001/vaultwarden-api:v1.1.0

# Update deployment
kubectl set image deployment/vaultwarden-api vaultwarden-api=localhost:5001/vaultwarden-api:v1.1.0 -n vaultwarden
```

Or let Flux handle it automatically if using image automation.
