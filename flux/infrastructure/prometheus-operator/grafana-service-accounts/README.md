# 🔐 Grafana Service Account Automation

Automated Grafana service account creation through GitOps workflows for consistent, version-controlled, and auditable service account management.

**Related Issue:** BVL-319 - Automate Service Account Creation

## 📋 Overview

This module provides a standardized way to create and manage Grafana ServiceAccounts through GitOps. All service accounts defined here are:

- **Defined as code** - YAML configuration in Git
- **Version controlled** - Full history of changes
- **Automatically applied** - Job runs on Flux reconciliation
- **Consistent** - Follows naming conventions and patterns
- **Auditable** - Tokens stored in labeled K8s secrets

## 🏗️ Directory Structure

```
grafana-service-accounts/
├── README.md                    # This file
├── kustomization.yaml           # Kustomization config
├── rbac.yaml                    # RBAC for the sync job
├── configmap.yaml               # Service account definitions
└── service-account-job.yaml     # Job that creates accounts
```

## 🚀 Quick Start

### Adding a New Service Account

1. Edit `configmap.yaml` and add your service account:

```yaml
serviceAccounts:
  - name: my-new-service
    role: Editor           # Admin | Editor | Viewer
    description: "Description of what this SA is for"
    secret_name: grafana-sa-my-new-service-token
```

2. Commit and push to Git

3. Flux will reconcile and the job will create the service account

4. Token will be available in secret: `grafana-sa-my-new-service-token`

### Using a Service Account Token

```bash
# Get the token
kubectl get secret grafana-sa-my-new-service-token -n prometheus -o jsonpath='{.data.token}' | base64 -d

# Use with Grafana API
curl -H "Authorization: Bearer <TOKEN>" \
  http://grafana.example.com/api/dashboards/home
```

### Using in Applications

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
    - name: app
      env:
        - name: GRAFANA_TOKEN
          valueFrom:
            secretKeyRef:
              name: grafana-sa-my-new-service-token
              key: token
        - name: GRAFANA_URL
          valueFrom:
            secretKeyRef:
              name: grafana-sa-my-new-service-token
              key: grafana-url
```

## 📝 Service Account Roles

| Role | Permissions |
|------|-------------|
| **Viewer** | Read-only access to dashboards and data sources |
| **Editor** | Create/edit dashboards, alerts, and data sources |
| **Admin** | Full access including user management and settings |

## 🏷️ Naming Conventions

| Component | Pattern | Example |
|-----------|---------|---------|
| Service Account | `{purpose}` | `ci-cd-deployer` |
| K8s Secret | `grafana-sa-{name}-token` | `grafana-sa-ci-cd-deployer-token` |

## 🔄 Lifecycle Management

### Create
1. Add entry to `configmap.yaml`
2. Commit to Git
3. Job creates SA and stores token

### Update Role
1. Modify entry in `configmap.yaml`
2. Manually delete the existing SA in Grafana UI
3. Re-run job (delete existing job pod to trigger new run)

### Delete
1. Remove entry from `configmap.yaml`
2. Manually delete SA in Grafana UI
3. Delete the K8s secret:
   ```bash
   kubectl delete secret grafana-sa-{name}-token -n prometheus
   ```

### Rotate Token
```bash
# Delete the secret to trigger token regeneration
kubectl delete secret grafana-sa-{name}-token -n prometheus

# Re-run the job
kubectl delete job grafana-service-accounts-sync -n prometheus
# Flux will recreate the job
```

## 📊 Monitoring

Service account creation is monitored via PrometheusRule alerts:

| Alert | Severity | Description |
|-------|----------|-------------|
| `GrafanaServiceAccountSyncFailed` | warning | Sync job failed |
| `GrafanaServiceAccountTokenExpiring` | warning | Token expiring soon |

## 🔒 Security Considerations

1. **Token Storage**: Tokens are stored in K8s secrets with labels for tracking
2. **Least Privilege**: Use the minimum required role for each service account
3. **Token Rotation**: Rotate tokens periodically by deleting the secret
4. **Audit Trail**: All changes tracked in Git history

## 🆘 Troubleshooting

### Job Failed

```bash
# Check job status
kubectl get job grafana-service-accounts-sync -n prometheus

# View job logs
kubectl logs -l app=grafana-service-accounts -n prometheus

# Re-run job
kubectl delete job grafana-service-accounts-sync -n prometheus
```

### Token Not Working

```bash
# Verify token exists
kubectl get secret grafana-sa-{name}-token -n prometheus

# Test token
TOKEN=$(kubectl get secret grafana-sa-{name}-token -n prometheus -o jsonpath='{.data.token}' | base64 -d)
curl -H "Authorization: Bearer $TOKEN" http://grafana-url/api/user
```

### Service Account Not Created

1. Check Grafana is accessible:
   ```bash
   kubectl exec -it deploy/kube-prometheus-stack-grafana -n prometheus -- curl -s localhost:3000/api/health
   ```

2. Verify admin credentials:
   ```bash
   kubectl get secret prometheus -n prometheus -o jsonpath='{.data.grafana-password}' | base64 -d
   ```

## 📚 Default Service Accounts

| Name | Role | Purpose |
|------|------|---------|
| `ci-cd-deployer` | Editor | CI/CD pipeline dashboard deployments |
| `monitoring-reader` | Viewer | External monitoring system access |
| `agent-sre` | Editor | Agent SRE automated operations |
| `alerting-integration` | Viewer | External alerting integration |
| `api-automation` | Admin | Full API access for automation |

## 🔗 Related Documentation

- [Grafana Service Accounts](https://grafana.com/docs/grafana/latest/administration/service-accounts/)
- [Grafana HTTP API](https://grafana.com/docs/grafana/latest/developers/http_api/)
- [Flux GitOps](https://fluxcd.io/docs/)
