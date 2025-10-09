# 🔐 Secrets Management Guide

This document explains how to manage secrets securely in the homelab infrastructure using **Sealed Secrets**.

## 🎯 Overview

We use [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets) to encrypt secrets that can be safely committed to Git. The secrets are encrypted using the cluster's public key and can only be decrypted by the Sealed Secrets controller running in the cluster.

## ⚠️ Security Policy

**NEVER commit plaintext secrets to Git!**

All secrets must be:
1. ✅ Encrypted using Sealed Secrets
2. ✅ Stored as environment variables
3. ✅ Referenced from Kubernetes secrets

## 📋 Available Secret Scripts

### 1. Cloudflare Tunnel API Token
**Script**: `create-cloudflare-tunnel-sealed-secret.sh`

Creates a sealed secret for Cloudflare Tunnel API credentials.

```bash
./scripts/create-cloudflare-tunnel-sealed-secret.sh
```

**Required values**:
- Cloudflare API Token
- Cloudflare Account ID

**Used by**: `scripts/update-tunnel.py`

**Environment variables**:
```bash
export CLOUDFLARE_API_TOKEN="your-token"
export CLOUDFLARE_ACCOUNT_ID="your-account-id"
export CLOUDFLARE_TUNNEL_NAME="homelab"  # optional, defaults to homelab
```

---

### 2. Loki MinIO Credentials
**Script**: `create-loki-minio-sealed-secret.sh`

Creates a sealed secret for Loki's MinIO storage backend.

```bash
./scripts/create-loki-minio-sealed-secret.sh
```

**Required values**:
- MinIO Root User (default: root-user)
- MinIO Root Password

**Used by**: `flux/clusters/homelab/infrastructure/loki/helmrelease.yaml`

**Output**: `flux/clusters/homelab/infrastructure/loki/loki-minio-secret-sealed.yaml`

---

### 3. SRE Agent API Key
**Script**: `create-sre-agent-sealed-secret.sh`

Creates a sealed secret for the SRE Agent API authentication.

```bash
./scripts/create-sre-agent-sealed-secret.sh
```

**Required values**:
- SRE API Key

**Used by**: `flux/clusters/homelab/infrastructure/agent-sre/`

**Output**: `flux/clusters/homelab/infrastructure/agent-sre/agent-sre-secret-sealed.yaml`

**Environment variable**: `SRE_API_KEY`

---

### 4. Grafana MCP API Key
**Script**: `create-grafana-mcp-sealed-secret.sh`

Creates a sealed secret for the Grafana MCP server API key.

```bash
./scripts/create-grafana-mcp-sealed-secret.sh
```

**Required values**:
- Grafana API Key

**Used by**: `flux/clusters/homelab/infrastructure/grafana-mcp/k8s-all.yaml`

**Output**: `flux/clusters/homelab/infrastructure/grafana-mcp/grafana-mcp-secret-sealed.yaml`

---

## 🚀 Quick Start

### Prerequisites

1. **Install kubeseal**:
   ```bash
   brew install kubeseal
   ```

2. **Install Sealed Secrets controller** (if not already installed):
   ```bash
   kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.5/controller.yaml
   ```

3. **Verify controller is running**:
   ```bash
   kubectl get pods -n kube-system | grep sealed-secrets
   ```

### Creating Secrets

1. Run the appropriate script from the `scripts/` directory
2. Follow the prompts to enter secret values
3. Review the generated sealed secret file
4. Commit the sealed secret to Git
5. Apply or let Flux sync it to the cluster

### Example Workflow

```bash
# 1. Create a sealed secret
cd /Users/brunolucena/workspace/bruno/repos/homelab
./scripts/create-loki-minio-sealed-secret.sh

# 2. Review the sealed secret
cat flux/clusters/homelab/infrastructure/loki/loki-minio-secret-sealed.yaml

# 3. Commit to Git
git add flux/clusters/homelab/infrastructure/loki/loki-minio-secret-sealed.yaml
git commit -m "🔐 Add Loki MinIO sealed secret"
git push

# 4. Verify the secret was created in the cluster
kubectl get secret loki-minio-secret -n loki
```

## 🔍 Verifying Secrets

### Check if a secret exists:
```bash
kubectl get secret <secret-name> -n <namespace>
```

### View secret contents (base64 encoded):
```bash
kubectl get secret <secret-name> -n <namespace> -o yaml
```

### Decode a specific key:
```bash
kubectl get secret <secret-name> -n <namespace> -o jsonpath='{.data.<key>}' | base64 -d
```

## 🔄 Rotating Secrets

To rotate a secret:

1. Delete the old sealed secret:
   ```bash
   kubectl delete sealedsecret <secret-name> -n <namespace>
   ```

2. Run the creation script again with new values

3. Apply the new sealed secret

4. Restart pods that use the secret:
   ```bash
   kubectl rollout restart deployment <deployment-name> -n <namespace>
   ```

## 📚 Additional Resources

- [Sealed Secrets Documentation](https://github.com/bitnami-labs/sealed-secrets)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
- [Security Best Practices](https://kubernetes.io/docs/concepts/security/secrets-good-practices/)

## 🛡️ Security Checklist

Before committing any changes:

- [ ] No plaintext secrets in YAML files
- [ ] No hardcoded API keys in scripts
- [ ] All secrets use `existingSecret` or environment variables
- [ ] Sealed secret files are committed (they're safe!)
- [ ] Original secret values are NOT committed
- [ ] README/docs updated if needed

## ❗ Emergency: Leaked Secret

If a secret is accidentally committed:

1. **Immediately rotate the leaked credential** (change the password/key at the source)
2. Remove from Git history:
   ```bash
   git filter-branch --force --index-filter \
     "git rm --cached --ignore-unmatch path/to/secret" \
     --prune-empty --tag-name-filter cat -- --all
   ```
3. Force push: `git push origin --force --all`
4. Create new sealed secret with rotated credentials
5. Notify team members to rebase their branches

## 📞 Support

For questions or issues with secrets management, check:
- This documentation
- Sealed Secrets logs: `kubectl logs -n kube-system -l name=sealed-secrets-controller`
- Kubernetes events: `kubectl get events -n <namespace>`

