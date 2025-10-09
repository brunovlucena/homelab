# 🚀 Quick Reference - Sealed Secrets

## 🎯 TL;DR

All secrets are now managed with **Sealed Secrets**. Run the scripts, commit the encrypted files, done!

## 📋 One-Time Setup

```bash
# 1. Install kubeseal (if not already installed)
brew install kubeseal

# 2. Verify Sealed Secrets controller is running
kubectl get pods -n kube-system | grep sealed-secrets
```

## 🔐 Creating Secrets

### All Secrets at Once
```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab

# Run all scripts
./scripts/create-cloudflare-tunnel-sealed-secret.sh
./scripts/create-loki-minio-sealed-secret.sh
./scripts/create-sre-agent-sealed-secret.sh
./scripts/create-grafana-mcp-sealed-secret.sh
```

### Individual Secrets

```bash
# Cloudflare Tunnel
./scripts/create-cloudflare-tunnel-sealed-secret.sh

# Loki MinIO
./scripts/create-loki-minio-sealed-secret.sh

# SRE Agent
./scripts/create-sre-agent-sealed-secret.sh

# Grafana MCP
./scripts/create-grafana-mcp-sealed-secret.sh
```

## ✅ Verification

```bash
# Check sealed secret files were created
find flux/clusters/homelab/infrastructure -name "*-sealed.yaml"

# Verify in cluster
kubectl get secret loki-minio-secret -n loki
kubectl get secret agent-sre-secret -n agent-sre
kubectl get secret grafana-mcp-secrets -n grafana-mcp
kubectl get secret cloudflare-tunnel-secret -n default
```

## 🔄 Using Cloudflare Tunnel Script

```bash
# Export credentials from sealed secret
export CLOUDFLARE_API_TOKEN=$(kubectl get secret cloudflare-tunnel-secret -n default -o jsonpath='{.data.api-token}' | base64 -d)
export CLOUDFLARE_ACCOUNT_ID=$(kubectl get secret cloudflare-tunnel-secret -n default -o jsonpath='{.data.account-id}' | base64 -d)

# Run the script
python scripts/update-tunnel.py
```

## 📚 Full Documentation

- **How-to Guide**: `scripts/SECRETS_MANAGEMENT.md`
- **Security Policy**: `.github/SECURITY.md`
- **Audit Report**: `scripts/SECURITY_AUDIT.md`
- **Summary**: `SECURITY_CHANGES_SUMMARY.md`

## 🆘 Common Issues

### "kubeseal: command not found"
```bash
brew install kubeseal
```

### "Sealed Secrets controller not found"
```bash
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.5/controller.yaml
```

### "Cannot connect to cluster"
```bash
kubectl cluster-info
# Fix your kubeconfig if needed
```

### Need to rotate a secret?
1. Delete old sealed secret: `kubectl delete sealedsecret <name> -n <namespace>`
2. Run the script again with new values
3. Restart affected pods: `kubectl rollout restart deployment <name> -n <namespace>`

## 💡 Tips

- ✅ **Safe to commit**: Sealed secret YAML files
- ❌ **NEVER commit**: The values you enter into the scripts
- 🔄 **Rotate regularly**: Especially after team changes
- 📝 **Document custom secrets**: Add to this guide if you create new ones

## 🎓 Learning Resources

- [Sealed Secrets GitHub](https://github.com/bitnami-labs/sealed-secrets)
- [Kubernetes Secrets Docs](https://kubernetes.io/docs/concepts/configuration/secret/)

