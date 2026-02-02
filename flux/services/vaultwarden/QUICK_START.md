# Quick Start Guide

This guide will help you get started with your self-hosted password manager.

## What Was Created

### Backend (Go)
- ✅ API server with Bitwarden-compatible endpoints
- ✅ HashiCorp Vault integration for secure storage
- ✅ JWT authentication
- ✅ Kubernetes deployment manifests

### Infrastructure
- ✅ HashiCorp Vault deployment (dev mode)
- ✅ Kubernetes manifests for API server
- ✅ Service, Ingress, and ServiceAccount configurations

### Clients
- ✅ iOS app structure (SwiftUI)
- ✅ Chrome browser extension
- ✅ Foundation for Firefox and Safari extensions

## Next Steps

### 1. Complete Backend Implementation

The backend is mostly complete, but you may want to:

- Add user registration endpoint
- Implement proper UUID generation (currently using simple IDs)
- Add password strength validation
- Implement rate limiting
- Add audit logging

### 2. Deploy Infrastructure

```bash
# Deploy Vault
kubectl apply -k flux/infrastructure/vault

# Wait for Vault to be ready
kubectl wait --for=condition=ready pod -l app=vault -n vault-system --timeout=300s

# Enable KV secrets engine
kubectl exec -it deployment/vault -n vault-system -- vault secrets enable -path=secret kv-v2
```

### 3. Build and Deploy Backend

```bash
cd flux/services/vaultwarden/backend

# Install dependencies
go mod tidy

# Build Docker image
docker build -t localhost:5001/vaultwarden-api:latest .

# Push to registry
docker push localhost:5001/vaultwarden-api:latest

# Deploy to Kubernetes
kubectl apply -k k8s/base
```

### 4. Create Secrets

```bash
# Generate JWT secret
JWT_SECRET=$(openssl rand -hex 32)

# Create secret (adjust namespace as needed)
kubectl create secret generic vaultwarden-secrets \
  --from-literal=jwt-secret="$JWT_SECRET" \
  --namespace vaultwarden
```

### 5. Create First User

You'll need to create the first user manually in Vault:

```bash
# Get Vault pod
VAULT_POD=$(kubectl get pod -n vault-system -l app=vault -o jsonpath='{.items[0].metadata.name}')

# Hash password (in production, do this properly)
# For now, we'll store it as plaintext (NOT RECOMMENDED FOR PRODUCTION)
# You should hash it properly before storing

# Store user auth
kubectl exec -it $VAULT_POD -n vault-system -- vault kv put secret/vaultwarden/auth/user@example.com \
  user_id="user-123" \
  email="user@example.com" \
  password_hash="<bcrypt-hashed-password>"

# Store user profile
kubectl exec -it $VAULT_POD -n vault-system -- vault kv put secret/vaultwarden/users/user-123/profile \
  email="user@example.com" \
  name="User Name"
```

### 6. Complete iOS App

The iOS app structure is created, but you need to:

1. Create Xcode project:
   ```bash
   cd ios
   # Create new Xcode project in Xcode, or use command line tools
   ```

2. Implement password encryption/decryption in `CryptoService.swift`

3. Add biometric authentication

4. Implement auto-fill credential provider extension

### 7. Complete Browser Extensions

The Chrome extension has basic structure. You need to:

1. Create extension icons (16x16, 48x48, 128x128)
2. Implement proper password encryption/decryption
3. Add password generator
4. Improve auto-fill detection
5. Copy Firefox and Safari extensions from Chrome base

## Security Notes

⚠️ **IMPORTANT**: This is a development setup. For production:

1. **Vault**: Don't use dev mode. Initialize properly with unseal keys
2. **Authentication**: Implement proper password hashing before storage
3. **Encryption**: Ensure client-side encryption is properly implemented
4. **TLS**: Verify certificates are properly configured
5. **Secrets**: Use Sealed Secrets for all Kubernetes secrets
6. **Network Policies**: Add network policies for security
7. **Audit Logging**: Enable Vault audit logging
8. **Backups**: Set up regular backups of Vault data

## Testing

### Test API

```bash
# Port forward to API
kubectl port-forward -n vaultwarden svc/vaultwarden-api 8080:80

# Test health
curl http://localhost:8080/health

# Test login (after creating user)
curl -X POST http://localhost:8080/api/identity/connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password&username=user@example.com&password=your-password"
```

## Troubleshooting

### Vault Connection Issues

```bash
# Check Vault pod
kubectl get pods -n vault-system

# Check Vault logs
kubectl logs -n vault-system deployment/vault

# Test connection from API pod
kubectl exec -it deployment/vaultwarden-api -n vaultwarden -- wget -O- http://vault.vault-system.svc.cluster.local:8200/v1/sys/health
```

### API Issues

```bash
# Check API logs
kubectl logs -n vaultwarden deployment/vaultwarden-api

# Check service
kubectl get svc -n vaultwarden

# Check ingress
kubectl get ingress -n vaultwarden
```

## Next Enhancements

Consider adding:

- [ ] Password generator API endpoint
- [ ] Two-factor authentication
- [ ] Password sharing between users
- [ ] Password history/versioning
- [ ] Secure file attachments
- [ ] Organization support
- [ ] Email notifications
- [ ] Password breach detection
- [ ] Web interface (admin panel)
- [ ] Mobile apps for Android
- [ ] CLI client

## Resources

- [Bitwarden API Documentation](https://bitwarden.com/help/api/)
- [HashiCorp Vault Documentation](https://www.vaultproject.io/docs)
- [SwiftUI Documentation](https://developer.apple.com/documentation/swiftui/)
- [Chrome Extension Documentation](https://developer.chrome.com/docs/extensions/)
