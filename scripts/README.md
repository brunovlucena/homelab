# 📜 Scripts Directory

This directory contains automation scripts and security tools for managing the homelab infrastructure.

## 🔐 Security & Secrets Management

### Sealed Secret Generation Script

All secrets are now managed using **Sealed Secrets** for secure GitOps workflows.

| Script | Purpose | Documentation |
|--------|---------|---------------|
| **`create-all-secrets.sh`** | **🔐 ONE SCRIPT TO RULE THEM ALL - Generates ALL secrets interactively with beautiful menu** | [SECRETS_MANAGEMENT.md](SECRETS_MANAGEMENT.md) |

This master script replaces all previous individual secret generation scripts. It handles:
- 🔐 Loki MinIO credentials (`loki` namespace)
- 🤖 SRE Agent API keys (`agent-sre` namespace)
- 📊 Grafana MCP API keys (`grafana-mcp` namespace)
- ☁️ Cloudflare Tunnel credentials (`default` namespace)
- 🌐 Homepage Cloudflare settings (`bruno` namespace)
- 🗄️ Homepage MinIO credentials (`bruno` namespace)

### Security Documentation

| Document | Description |
|----------|-------------|
| [QUICK_REFERENCE.md](QUICK_REFERENCE.md) | 🚀 Quick start guide - start here! |
| [SECRETS_MANAGEMENT.md](SECRETS_MANAGEMENT.md) | 📚 Complete sealed secrets guide |
| [SECURITY_AUDIT.md](SECURITY_AUDIT.md) | 🔍 Security audit report & findings |
| [../SECURITY.md](../.github/SECURITY.md) | 🛡️ Repository security policy |
| [../SECURITY_CHANGES_SUMMARY.md](../SECURITY_CHANGES_SUMMARY.md) | 📊 Summary of security changes |

## 🚀 Quick Start

### 🔥 ONE SCRIPT TO RULE THEM ALL 🔥
```bash
# 1. Install kubeseal
brew install kubeseal

# 2. Run the master script (generates ALL secrets)
./create-all-secrets.sh

# 3. Follow the interactive menu:
#    - Option 1: Generate ALL secrets at once (recommended)
#    - Options 2-7: Generate individual secrets
#    - Option 8: Exit

# 4. Verify
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
**Scripts**: 7 total (1 MASTER SECRET GENERATOR + 6 infrastructure/utility scripts)  
**Documentation**: 5 comprehensive guides

## 🎯 TL;DR - Just Get Started

```bash
# Install prerequisites
brew install kubeseal

# Run ONE script to generate ALL secrets
cd scripts
./create-all-secrets.sh

# Follow the interactive prompts
# Select option 1 to generate all secrets at once
# Or pick individual secrets from the menu

# Done! 🎉
```

