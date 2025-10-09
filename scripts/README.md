# 📜 Scripts Directory

This directory contains automation scripts and security tools for managing the homelab infrastructure.

## 🔐 Security & Secrets Management

### Sealed Secret Generation Scripts

All secrets are now managed using **Sealed Secrets** for secure GitOps workflows. The following scripts help generate encrypted secrets:

| Script | Purpose | Namespace | Documentation |
|--------|---------|-----------|---------------|
| `create-cloudflare-tunnel-sealed-secret.sh` | Cloudflare API credentials | `default` | [SECRETS_MANAGEMENT.md](SECRETS_MANAGEMENT.md) |
| `create-loki-minio-sealed-secret.sh` | Loki MinIO storage | `loki` | [SECRETS_MANAGEMENT.md](SECRETS_MANAGEMENT.md) |
| `create-sre-agent-sealed-secret.sh` | SRE Agent API | `agent-sre` | [SECRETS_MANAGEMENT.md](SECRETS_MANAGEMENT.md) |
| `create-grafana-mcp-sealed-secret.sh` | Grafana MCP API | `grafana-mcp` | [SECRETS_MANAGEMENT.md](SECRETS_MANAGEMENT.md) |

### Security Documentation

| Document | Description |
|----------|-------------|
| [QUICK_REFERENCE.md](QUICK_REFERENCE.md) | 🚀 Quick start guide - start here! |
| [SECRETS_MANAGEMENT.md](SECRETS_MANAGEMENT.md) | 📚 Complete sealed secrets guide |
| [SECURITY_AUDIT.md](SECURITY_AUDIT.md) | 🔍 Security audit report & findings |
| [../SECURITY.md](../.github/SECURITY.md) | 🛡️ Repository security policy |
| [../SECURITY_CHANGES_SUMMARY.md](../SECURITY_CHANGES_SUMMARY.md) | 📊 Summary of security changes |

## 🚀 Quick Start

### First Time Setup
```bash
# 1. Install kubeseal
brew install kubeseal

# 2. Create all sealed secrets
./create-cloudflare-tunnel-sealed-secret.sh
./create-loki-minio-sealed-secret.sh
./create-sre-agent-sealed-secret.sh
./create-grafana-mcp-sealed-secret.sh

# 3. Verify
find ../flux/clusters/homelab/infrastructure -name "*-sealed.yaml"
```

### Daily Usage
```bash
# Export Cloudflare credentials for tunnel script
export CLOUDFLARE_API_TOKEN=$(kubectl get secret cloudflare-tunnel-secret -n default -o jsonpath='{.data.api-token}' | base64 -d)
export CLOUDFLARE_ACCOUNT_ID=$(kubectl get secret cloudflare-tunnel-secret -n default -o jsonpath='{.data.account-id}' | base64 -d)

# Run tunnel update
python update-tunnel.py
```

## 🔧 Other Scripts

| Script | Purpose |
|--------|---------|
| `update-tunnel.py` | Update Cloudflare tunnel routes with current service IPs |
| `install-linkerd.sh` | Install Linkerd service mesh |
| `install-linkerd-viz.sh` | Install Linkerd visualization dashboard |
| `setup-registry.sh` | Set up local container registry |

## 📚 Learning Path

If you're new to sealed secrets, read in this order:

1. **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - Quick commands and common tasks
2. **[SECRETS_MANAGEMENT.md](SECRETS_MANAGEMENT.md)** - Detailed how-to guide
3. **[SECURITY_AUDIT.md](SECURITY_AUDIT.md)** - What was changed and why
4. **[../.github/SECURITY.md](../.github/SECURITY.md)** - Overall security policy

## ⚠️ Security Reminders

### ✅ DO:
- ✅ Run sealed secret scripts to generate encrypted secrets
- ✅ Commit sealed secret YAML files to Git (they're encrypted!)
- ✅ Keep kubeseal and kubectl up to date
- ✅ Rotate secrets regularly

### ❌ DON'T:
- ❌ Commit plaintext secrets to Git
- ❌ Share the values you enter into the scripts
- ❌ Hardcode credentials in configuration files
- ❌ Skip the security documentation

## 🆘 Troubleshooting

### Script won't run?
```bash
# Make sure it's executable
chmod +x create-*-sealed-secret.sh

# Check prerequisites
command -v kubectl
command -v kubeseal
```

### Can't connect to cluster?
```bash
# Verify cluster access
kubectl cluster-info

# Check sealed-secrets controller
kubectl get pods -n kube-system | grep sealed-secrets
```

### Need help?
1. Check [QUICK_REFERENCE.md](QUICK_REFERENCE.md) for common issues
2. Review [SECRETS_MANAGEMENT.md](SECRETS_MANAGEMENT.md) for detailed guides
3. Check Sealed Secrets logs: `kubectl logs -n kube-system -l name=sealed-secrets-controller`

## 📞 Resources

- **Sealed Secrets**: https://github.com/bitnami-labs/sealed-secrets
- **Kubernetes Secrets**: https://kubernetes.io/docs/concepts/configuration/secret/
- **Security Best Practices**: https://kubernetes.io/docs/concepts/security/secrets-good-practices/

## 📝 Contributing

When adding new secrets:

1. Create a new sealed secret generation script
2. Update this README
3. Document in SECRETS_MANAGEMENT.md
4. Add to QUICK_REFERENCE.md
5. Test thoroughly before committing

---

**Last Updated**: October 9, 2025  
**Scripts**: 8 total (4 sealed secret generators + 4 infrastructure scripts)  
**Documentation**: 5 comprehensive guides

