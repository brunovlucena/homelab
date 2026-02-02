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
