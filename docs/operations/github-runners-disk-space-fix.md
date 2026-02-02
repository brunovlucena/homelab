# GitHub Runners Disk Space Fix

## Problem Summary

GitHub runner pods were crashing during initialization with the error:
```
cp: error writing '/home/runner/tmpDir/./node20/bin/node': No space left on device
```

## Root Cause

The `studio-worker2` node had only **3.7GB free** (98% disk usage), causing the init container to fail when copying externals (Node.js binaries, etc.) to EmptyDir volumes.

## Solutions Implemented

### 1. ✅ Cleanup Script

Created `scripts/cleanup-github-runners.sh` to automatically clean up:
- Failed pods
- Old succeeded pods (older than 1 hour)
- Old EphemeralRunner resources
- Provides disk usage status

**Usage:**
```bash
# Dry run (see what would be deleted)
./scripts/cleanup-github-runners.sh --dry-run

# Actually clean up
./scripts/cleanup-github-runners.sh

# Specify namespace
./scripts/cleanup-github-runners.sh --namespace github-runners
```

### 2. ✅ EmptyDir Size Limits

Added size limits to EmptyDir volumes in both runner HelmReleases:
- `dind-sock`: 1Gi limit
- `dind-externals`: 5Gi limit (Node.js and other externals)
- `work`: 10Gi limit (GitHub Actions work directory)

**Files updated:**
- `flux/infrastructure/github-runners/studio/runner-helmrelease.yaml`
- `flux/infrastructure/github-runners/pro/runner-helmrelease.yaml`

This prevents individual pods from consuming all available disk space.

### 3. ✅ Prometheus Alerts

Added disk space monitoring alerts:

**GitHub Runners Alert** (`github-runners.yaml`):
- `RunnerPodDiskSpaceExhausted`: Alerts when runner pods fail due to disk space exhaustion

**Node Disk Alerts** (`node-disk.yaml`):
- `NodeDiskSpaceCritical`: Fires when node has <5% disk space remaining
- `NodeDiskSpaceWarning`: Fires when node has <10% disk space remaining
- `NodeDiskWillFillIn4Hours`: Predictive alert based on growth rate
- `NodeDiskPressure`: Alerts when Kubernetes reports disk pressure

## Immediate Actions

1. **Clean up failed pods:**
   ```bash
   kubectl delete pod -n github-runners --field-selector=status.phase=Failed
   ```

2. **Clean up old succeeded pods:**
   ```bash
   kubectl delete pod -n github-runners --field-selector=status.phase=Succeeded
   ```

3. **Clean up unused container images on nodes:**
   ```bash
   # SSH to affected node and run:
   docker system prune -a --volumes -f
   # OR if using containerd:
   crictl rmi --prune
   ```

4. **Run the cleanup script:**
   ```bash
   ./scripts/cleanup-github-runners.sh
   ```

## Long-term Prevention

1. **Monitor disk usage:**
   - Check Grafana dashboards for node disk metrics
   - Set up alerting (already done)
   - Run cleanup script regularly (consider CronJob)

2. **Automatic cleanup:**
   Consider creating a CronJob to run the cleanup script periodically:
   ```yaml
   apiVersion: batch/v1
   kind: CronJob
   metadata:
     name: github-runners-cleanup
     namespace: github-runners
   spec:
     schedule: "0 */6 * * *"  # Every 6 hours
     jobTemplate:
       spec:
         template:
           spec:
             containers:
             - name: cleanup
               image: bitnami/kubectl:latest
               command:
               - /bin/sh
               - -c
               - |
                 kubectl delete pod --field-selector=status.phase=Failed -n github-runners --ignore-not-found=true
                 kubectl delete pod --field-selector=status.phase=Succeeded -n github-runners --ignore-not-found=true
             restartPolicy: OnFailure
   ```

3. **Resource quotas:**
   Consider adding ResourceQuota to limit ephemeral storage per namespace.

## Verification

After applying changes:

1. **Check runner pods:**
   ```bash
   kubectl get pods -n github-runners
   ```

2. **Check disk usage:**
   ```bash
   kubectl top nodes
   kubectl get nodes -o json | jq -r '.items[] | "\(.metadata.name): \(.status.conditions[] | select(.type=="DiskPressure") | .status)"'
   ```

3. **Check alerts:**
   ```bash
   kubectl get prometheusrule -n prometheus
   ```

4. **Verify EmptyDir limits:**
   ```bash
   kubectl get pod <runner-pod> -n github-runners -o yaml | grep -A 5 "emptyDir"
   ```

## Related Files

- `scripts/cleanup-github-runners.sh` - Cleanup script
- `flux/infrastructure/github-runners/studio/runner-helmrelease.yaml` - Studio runner config
- `flux/infrastructure/github-runners/pro/runner-helmrelease.yaml` - Pro runner config
- `flux/infrastructure/prometheus-operator/k8s/prometheusrules/github-runners.yaml` - Runner alerts
- `flux/infrastructure/prometheus-operator/k8s/prometheusrules/node-disk.yaml` - Node disk alerts
