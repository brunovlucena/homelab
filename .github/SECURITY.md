# 🔒 Security Policy

## 🎯 Reporting a Security Vulnerability

If you discover a security vulnerability in this repository, please report it by:

1. **DO NOT** create a public GitHub issue
2. Contact the maintainers privately
3. Provide detailed information about the vulnerability

## 🛡️ Security Best Practices

This repository follows these security practices:

### ✅ Secrets Management

- **All secrets are encrypted** using [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)
- **No plaintext secrets** are committed to version control
- **Environment variables** are used for runtime configuration
- **Kubernetes native secrets** are referenced in deployments

### 📋 Sealed Secrets

We use Sealed Secrets to encrypt sensitive data before committing to Git:

- Public key encryption ensures secrets can only be decrypted in the cluster
- Sealed secrets are safe to commit to Git
- Controller running in cluster handles decryption
- Helper scripts provided for secret generation

See [scripts/SECRETS_MANAGEMENT.md](../scripts/SECRETS_MANAGEMENT.md) for detailed documentation.

### 🔍 Security Audit

Regular security audits are performed to ensure:

- No hardcoded credentials
- No plaintext API keys
- No embedded tokens or passwords
- Proper use of Kubernetes secrets
- Environment variables used appropriately

Latest audit: [scripts/SECURITY_AUDIT.md](../scripts/SECURITY_AUDIT.md)

### 🚫 What NOT to Commit

**NEVER commit these to Git:**

- ❌ API keys or tokens
- ❌ Passwords or passphrases
- ❌ Private keys or certificates
- ❌ Database connection strings with credentials
- ❌ OAuth client secrets
- ❌ Encryption keys
- ❌ Cloud provider credentials (AWS keys, GCP service accounts, etc.)

### ✅ What IS Safe to Commit

These are safe to commit:

- ✅ Sealed Secret YAML files (encrypted)
- ✅ Public keys
- ✅ Public certificates
- ✅ Configuration templates
- ✅ Environment variable names (without values)
- ✅ Documentation

## 🔄 Secret Rotation

To rotate a secret:

1. Revoke the old credential at the source
2. Delete the old sealed secret from the cluster
3. Run the appropriate sealed secret creation script
4. Commit the new sealed secret
5. Restart affected pods

See [scripts/SECRETS_MANAGEMENT.md](../scripts/SECRETS_MANAGEMENT.md) for detailed instructions.

## 🚨 Leaked Secret Response

If a secret is accidentally committed:

1. **IMMEDIATELY** revoke/rotate the credential at the source
2. Remove it from Git history
3. Create a new sealed secret with the rotated credential
4. Force push to remove from all branches
5. Notify all team members
6. Document the incident

See emergency procedures in [scripts/SECRETS_MANAGEMENT.md](../scripts/SECRETS_MANAGEMENT.md).

## 📚 Resources

- [Kubernetes Secrets Documentation](https://kubernetes.io/docs/concepts/configuration/secret/)
- [Sealed Secrets GitHub](https://github.com/bitnami-labs/sealed-secrets)
- [OWASP Secrets Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)

## 🔐 Security Checklist

Before committing code, ensure:

- [ ] No plaintext secrets in YAML files
- [ ] No hardcoded API keys in scripts
- [ ] All secrets use sealed secrets or environment variables
- [ ] No credentials in configuration files
- [ ] Scripts have proper error handling for missing secrets
- [ ] Documentation updated if secrets management changed

## 📞 Contact

For security-related questions:
- Review documentation in `scripts/SECRETS_MANAGEMENT.md`
- Check sealed secrets status: `kubectl get sealedsecrets --all-namespaces`
- View controller logs: `kubectl logs -n kube-system -l name=sealed-secrets-controller`

---

**Last Updated**: October 9, 2025  
**Security Audit**: See [scripts/SECURITY_AUDIT.md](../scripts/SECURITY_AUDIT.md)

