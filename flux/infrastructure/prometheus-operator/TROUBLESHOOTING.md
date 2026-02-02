# Prometheus Operator Troubleshooting

## GitHub 504 Gateway Timeout Issues

### Problem

When deploying `kube-prometheus-stack`, you may encounter errors like:

```
chart pull error: failed to download chart for remote reference: 
failed to fetch https://github.com/prometheus-community/helm-charts/releases/download/kube-prometheus-stack-79.5.0/kube-prometheus-stack-79.5.0.tgz : 
504 Gateway Time-out
```

### Root Cause

GitHub's CDN (Content Delivery Network) can experience transient failures when serving large Helm chart tarballs. This is a known issue with GitHub releases, especially for large files.

### Solutions

#### 1. Manual Retry (Quick Fix)

Force a retry by annotating the HelmChart:

```bash
kubectl annotate helmchart prometheus-kube-prometheus-stack -n flux-system \
  reconcile.fluxcd.io/requestedAt="$(date +%s)" --overwrite
```

#### 2. Increase HelmRelease Timeout

The HelmRelease has a default timeout. Increase it for large charts:

```yaml
spec:
  timeout: 15m  # Increase from default 5m
  chart:
    spec:
      interval: 12h  # Already configured
```

#### 3. Monitor and Wait

Flux's source-controller automatically retries with exponential backoff. The chart will eventually download successfully. Monitor with:

```bash
kubectl get helmchart prometheus-kube-prometheus-stack -n flux-system -w
kubectl logs -n flux-system -l app=source-controller --tail=50 | grep prometheus
```

#### 4. Use OCI Registry (Future Improvement)

Consider migrating to OCI-based Helm repositories which are more reliable than GitHub releases:

```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: prometheus-community
  namespace: flux-system
spec:
  type: oci
  url: oci://ghcr.io/prometheus-community
  interval: 1h0m0s
```

**Note:** This requires the chart maintainers to publish to OCI registry.

### Prevention

1. **Monitor HelmChart Status**: Set up alerts for HelmChart failures
2. **Use Retry Annotations**: Document the manual retry command for operators
3. **Consider Chart Mirroring**: For production, consider mirroring charts to a local registry
4. **Increase Timeouts**: For large charts, always configure appropriate timeouts

### Related Issues

- Similar issues can occur with other GitHub-hosted charts (Grafana, Tempo, etc.)
- The same retry mechanism applies to all HelmCharts

### References

- [Flux HelmRepository Documentation](https://fluxcd.io/flux/components/source/helmrepositories/)
- [Flux Troubleshooting Guide](https://fluxcd.io/flux/guides/troubleshooting/)

---

## Grafana Session/Logout Issues (BVL-301)

### Problem

Users are repeatedly logged out of Grafana, especially after pod restarts, deployments, or cluster operations.

### Root Cause

The primary root cause is the **missing `secret_key` configuration**. Grafana uses a secret key to:

1. Sign session cookies
2. Encrypt sensitive data in the database
3. Generate CSRF tokens

**Without a persistent secret_key, Grafana generates a random key at each startup**, which immediately invalidates all existing sessions and cookies.

### The Fix (Applied)

The fix involves three components working together:

#### 1. Persistent Secret Key (Critical)

```yaml
# In helmrelease.yaml
grafana:
  envValueFrom:
    GF_SECURITY_SECRET_KEY:
      secretKeyRef:
        name: prometheus
        key: grafana-secret-key
```

The `secrets-job.yaml` automatically generates and persists this key in the `prometheus` secret on first run.

#### 2. Persistent Storage (PVC)

```yaml
grafana:
  persistence:
    enabled: true
    type: pvc
    size: 2Gi
```

Ensures SQLite database and session files survive pod restarts.

#### 3. File-based Session Storage

```yaml
grafana:
  ini:
    session:
      provider: file
      provider_config: /var/lib/grafana/sessions
```

Stores sessions on the persistent volume instead of memory.

#### 4. Extended Session Timeouts

```yaml
grafana:
  ini:
    auth:
      token_rotation_interval_minutes: 60
      login_maximum_inactive_lifetime_duration: 30d
      login_maximum_lifetime_duration: 90d
    session:
      session_life_time: 604800  # 7 days
    security:
      login_remember_days: 30
```

### Verification

1. Check secret key exists:
   ```bash
   kubectl get secret prometheus -n prometheus -o jsonpath='{.data.grafana-secret-key}' | base64 -d | head -c 20
   ```

2. Check Grafana is using the secret key:
   ```bash
   kubectl exec -n prometheus -it deploy/kube-prometheus-stack-grafana -- env | grep GF_SECURITY
   ```

3. Check PVC is bound:
   ```bash
   kubectl get pvc -n prometheus | grep grafana
   ```

### Manual Secret Key Generation (If Needed)

If the secrets-job hasn't run yet, you can manually add the key:

```bash
# Generate a secure key
GRAFANA_SECRET_KEY=$(openssl rand -base64 32)

# Add to prometheus secret
kubectl patch secret prometheus -n prometheus --type='json' \
  -p="[{\"op\": \"add\", \"path\": \"/data/grafana-secret-key\", \"value\": \"$(echo -n "${GRAFANA_SECRET_KEY}" | base64 -w0)\"}]"

# Restart Grafana to pick up the new key
kubectl rollout restart deployment/kube-prometheus-stack-grafana -n prometheus
```

**Important:** After adding/changing the secret_key, all users will need to log in again (one time only).

### References

- [Grafana Configuration - Security](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#security)
- [Grafana Configuration - Session](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#session)
