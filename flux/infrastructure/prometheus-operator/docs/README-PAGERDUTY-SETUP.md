# PagerDuty Setup for Alertmanager

This document explains how to set up PagerDuty integration for Alertmanager using the same pattern as GHCR secrets.

## Overview

The PagerDuty secret is created in two steps:
1. **Pulumi** creates the `prometheus` secret in `flux-system` namespace with `pagerduty-service-key`
2. **Kubernetes Job** (`secrets-job.yaml`) creates the `alertmanager-pagerduty` secret in `prometheus` namespace

## Setup Steps

### 1. Set Environment Variable

Add your PagerDuty service key to your `~/.zshrc`:

```bash
export PAGERDUTY_SERVICE_KEY='your-pagerduty-service-key-here'
```

Then reload your shell:
```bash
source ~/.zshrc
```

### 2. Create Secret via Pulumi

The Pulumi code (`pulumi/main.go`) already includes `pagerdutyServiceKey` in the prometheus secret configuration. Run:

```bash
cd pulumi
pulumi up
```

This will create the `prometheus` secret in the `flux-system` namespace with the key `pagerduty-service-key` (converted from camelCase `pagerdutyServiceKey`).

### 3. Secrets Job Creates Alertmanager Secret

The `secrets-job.yaml` will automatically:
1. Sync the `prometheus` secret from `flux-system` to `prometheus` namespace
2. Extract `pagerduty-service-key` from the synced secret
3. Create `alertmanager-pagerduty` secret with key `service-key`

To trigger the job manually:
```bash
kubectl delete job -n prometheus prometheus-operator-secret-sync
kubectl apply -f flux/infrastructure/prometheus-operator/secrets-job.yaml
```

Or wait for Flux to reconcile it automatically.

### 4. Verify Setup

Check that the secret was created:
```bash
kubectl get secret -n prometheus alertmanager-pagerduty
```

Check AlertmanagerConfig is using it:
```bash
kubectl get alertmanagerconfig -n prometheus pagerduty-config -o yaml | grep -A 5 serviceKey
```

## How It Works

### Pulumi Secret Creation

In `pulumi/main.go`, the `getSecretsConfig()` function maps:
- Secret name: `prometheus`
- Config keys: `["grafanaPassword", "grafanaApiKey", "pagerdutyUrl", "pagerdutyServiceKey", "slackWebhookUrl"]`

The `camelToKebab()` function converts `pagerdutyServiceKey` → `pagerduty-service-key` in the secret.

### Secrets Job Pattern

The `secrets-job.yaml` follows the same pattern as GHCR secret creation:
1. Reads from `flux-system` namespace (where Pulumi creates secrets)
2. Syncs to target namespace (`prometheus`)
3. Creates derived secrets (like `alertmanager-pagerduty` from `prometheus`)

### AlertmanagerConfig

The `alertmanagerconfig-pagerduty.yaml` references:
```yaml
serviceKey:
  name: alertmanager-pagerduty
  key: service-key
```

This matches the secret created by the job.

## Troubleshooting

### Secret Not Created

1. Check if Pulumi secret exists:
   ```bash
   kubectl get secret -n flux-system prometheus
   ```

2. Check if it has the key:
   ```bash
   kubectl get secret -n flux-system prometheus -o jsonpath='{.data}' | jq 'keys'
   ```

3. Check job logs:
   ```bash
   kubectl logs -n prometheus -l app=prometheus-operator,component=secret-sync
   ```

### Alertmanager Not Using Secret

1. Verify AlertmanagerConfig:
   ```bash
   kubectl get alertmanagerconfig -n prometheus pagerduty-config
   ```

2. Check Alertmanager configuration:
   ```bash
   kubectl get secret -n prometheus alertmanager-kube-prometheus-stack-alertmanager-generated -o jsonpath='{.data.alertmanager\.yaml}' | base64 -d | grep -A 10 pagerduty
   ```

3. Restart Alertmanager if needed:
   ```bash
   kubectl rollout restart statefulset -n prometheus kube-prometheus-stack-alertmanager
   ```

## Related Files

- `pulumi/main.go` - Pulumi secret creation
- `flux/infrastructure/prometheus-operator/secrets-job.yaml` - Kubernetes job that creates alertmanager-pagerduty secret
- `flux/infrastructure/prometheus-operator/k8s/alertmanagerconfig-pagerduty.yaml` - AlertmanagerConfig CRD
- `flux/infrastructure/prometheus-operator/k8s/prometheusrules/grafana.yaml` - Grafana health monitoring alerts
