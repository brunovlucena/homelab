# 🚨 Runbook: Knative Operator Crash Loop

## Alert Information

**Alert Name:** `KnativeOperatorCrashLoop`  
**Severity:** Critical  
**Component:** knative-operator  
**Service:** knative-operator

## Symptom

Knative Operator pod is continuously crashing and restarting (CrashLoopBackOff state).

## Impact

- **User Impact:** MEDIUM - Existing Knative services continue running
- **Business Impact:** HIGH - Cannot manage Knative components or deploy updates
- **Data Impact:** LOW - No data loss, management functionality impaired

## Diagnosis

### 1. Check Pod Status

```bash
# Check operator pod status
kubectl get pods -n knative-operator -l app=knative-operator

# Check detailed pod status
kubectl get pods -n knative-operator -l app=knative-operator -o wide

# Check restart count
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].restartCount}'
```

**Problematic Output:**
```
NAME                                READY   STATUS             RESTARTS   AGE
knative-operator-xxxxxxxxxx-xxxxx   0/1     CrashLoopBackOff   15         45m
```

### 2. Check Pod Events

```bash
# Get pod events
kubectl describe pod -n knative-operator -l app=knative-operator | grep -A 20 "Events:"

# Get recent events in namespace
kubectl get events -n knative-operator --sort-by='.lastTimestamp' | tail -20
```

### 3. Check Current Logs

```bash
# View current logs
kubectl logs -n knative-operator -l app=knative-operator --tail=100

# View previous logs (from before crash)
kubectl logs -n knative-operator -l app=knative-operator --previous --tail=200

# Follow logs in real-time
kubectl logs -n knative-operator -l app=knative-operator -f
```

### 4. Check Container Exit Code

```bash
# Get exit code
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].lastState.terminated.exitCode}'

# Get termination reason
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].lastState.terminated.reason}'
```

**Common Exit Codes:**
- `1` - Application error
- `137` - SIGKILL (OOMKilled)
- `139` - Segmentation fault
- `143` - SIGTERM (graceful shutdown)

### 5. Check Resource Constraints

```bash
# Check resource usage
kubectl top pod -n knative-operator -l app=knative-operator

# Check resource limits
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].spec.containers[0].resources}'

# Check for OOMKills
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].lastState.terminated.reason}'
```

### 6. Check Configuration

```bash
# Check deployment configuration
kubectl get deployment -n knative-operator knative-operator -o yaml

# Check HelmRelease
kubectl get helmrelease -n knative-operator knative-operator -o yaml

# Check for config errors
flux get helmreleases -n knative-operator
```

## Resolution Steps

### Step 1: Identify Crash Cause

#### Check Logs for Panic or Fatal Errors

```bash
# Check for panic messages
kubectl logs -n knative-operator -l app=knative-operator --previous | grep -i "panic\|fatal\|error" | tail -20

# Check startup sequence
kubectl logs -n knative-operator -l app=knative-operator --previous | head -50

# Check for specific error patterns
kubectl logs -n knative-operator -l app=knative-operator --previous | grep -E "failed to|cannot|unable to"
```

#### Analyze Exit Code and Reason

```bash
# Get full termination details
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].lastState.terminated}' | jq .
```

### Step 2: Common Issues and Fixes

#### Issue 1: OOMKilled (Exit Code 137)

**Cause:** Operator exceeded memory limits  
**Fix:**

```bash
# Confirm OOMKilled
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].lastState.terminated.reason}'

# Check current memory limit
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].spec.containers[0].resources.limits.memory}'

# Increase memory limits
kubectl edit helmrelease knative-operator -n knative-operator

# Update resources:
# values:
#   resources:
#     limits:
#       memory: 2Gi  # Increase from 512Mi/1Gi
#     requests:
#       memory: 512Mi

# Reconcile to apply
flux reconcile helmrelease knative-operator -n knative-operator
```

#### Issue 2: CRD Installation Failure

**Cause:** Cannot install or validate CRDs  
**Fix:**

```bash
# Check if CRDs exist
kubectl get crd | grep knative

# Check CRD installation errors
kubectl logs -n knative-operator -l app=knative-operator --previous | grep -i "crd"

# Get existing CRD status
kubectl get crd knativeservings.operator.knative.dev -o yaml
kubectl get crd knativeeventings.operator.knative.dev -o yaml

# Delete and recreate CRDs (CAUTION: Don't do if KnativeServing resources exist!)
kubectl get knativeserving -A  # Check if any exist
kubectl get knativeeventing -A

# If safe (no resources), delete CRDs
kubectl delete crd knativeservings.operator.knative.dev
kubectl delete crd knativeeventings.operator.knative.dev

# Reconcile to reinstall
flux reconcile helmrelease knative-operator -n knative-operator --force
```

#### Issue 3: RBAC Permission Denied

**Cause:** Missing or incorrect permissions  
**Fix:**

```bash
# Check for permission errors in logs
kubectl logs -n knative-operator -l app=knative-operator --previous | grep -i "forbidden\|unauthorized\|permission denied"

# Verify ServiceAccount exists
kubectl get serviceaccount -n knative-operator knative-operator-webhook

# Check ClusterRoleBindings
kubectl get clusterrolebinding | grep knative-operator

# Check operator's ClusterRole
kubectl get clusterrole knative-operator-webhook-cluster-admin -o yaml

# Reconcile to restore RBAC
flux reconcile helmrelease knative-operator -n knative-operator --force
```

#### Issue 4: Configuration Error

**Cause:** Invalid configuration in HelmRelease or values  
**Fix:**

```bash
# Check HelmRelease status
flux get helmreleases -n knative-operator

# Check for validation errors
kubectl describe helmrelease -n knative-operator knative-operator | grep -A 20 "Status:"

# Get HelmRelease configuration
kubectl get helmrelease -n knative-operator knative-operator -o yaml > /tmp/knative-operator-hr.yaml

# Review configuration
cat /tmp/knative-operator-hr.yaml

# Check Helm values
helm get values knative-operator -n knative-operator

# If bad configuration, edit to fix
kubectl edit helmrelease knative-operator -n knative-operator

# Or suspend and fix in Git
flux suspend helmrelease knative-operator -n knative-operator
# Fix configuration in Git repo
# Commit and push
flux resume helmrelease knative-operator -n knative-operator
```

#### Issue 5: Webhook Certificate Issues

**Cause:** Webhook TLS certificate problems  
**Fix:**

```bash
# Check for certificate errors
kubectl logs -n knative-operator -l app=knative-operator --previous | grep -i "certificate\|tls\|x509"

# Check webhook configuration
kubectl get validatingwebhookconfigurations | grep knative
kubectl get mutatingwebhookconfigurations | grep knative

# Describe webhook
kubectl describe validatingwebhookconfigurations operator.knative.dev

# Check webhook secret
kubectl get secret -n knative-operator | grep webhook
kubectl describe secret -n knative-operator knative-operator-webhook-certs

# Delete webhook configurations to force recreation
kubectl delete validatingwebhookconfigurations operator.knative.dev
kubectl delete mutatingwebhookconfigurations operator.knative.dev

# Restart operator to recreate
kubectl rollout restart deployment -n knative-operator knative-operator
```

#### Issue 6: Port Conflict or Network Issue

**Cause:** Cannot bind to required ports  
**Fix:**

```bash
# Check for port binding errors
kubectl logs -n knative-operator -l app=knative-operator --previous | grep -i "port\|bind\|address already in use"

# Check ports configuration
kubectl get deployment -n knative-operator knative-operator -o jsonpath='{.spec.template.spec.containers[0].ports}' | jq .

# Check for port conflicts
kubectl get pods -n knative-operator -o wide

# Check service configuration
kubectl get service -n knative-operator
kubectl describe service -n knative-operator knative-operator-webhook
```

#### Issue 7: API Server Connection Issues

**Cause:** Cannot connect to Kubernetes API server  
**Fix:**

```bash
# Check for API connection errors
kubectl logs -n knative-operator -l app=knative-operator --previous | grep -i "connection refused\|timeout\|dial tcp"

# Check if API server is accessible from pod network
kubectl run test-api --image=curlimages/curl:latest --rm -it --restart=Never -- sh -c "curl -k https://kubernetes.default.svc"

# Check network policies
kubectl get networkpolicies -n knative-operator

# Verify DNS resolution
kubectl run test-dns --image=busybox:latest --rm -it --restart=Never -- nslookup kubernetes.default.svc
```

#### Issue 8: Image or Binary Corruption

**Cause:** Corrupted operator image or binary  
**Fix:**

```bash
# Check image
kubectl get deployment -n knative-operator knative-operator -o jsonpath='{.spec.template.spec.containers[0].image}'

# Check image pull errors
kubectl describe pod -n knative-operator -l app=knative-operator | grep -A 10 "Events:"

# Force image re-pull
kubectl patch deployment -n knative-operator knative-operator -p '{"spec":{"template":{"spec":{"containers":[{"name":"knative-operator","imagePullPolicy":"Always"}]}}}}'

# Delete pod to force recreation with fresh image
kubectl delete pod -n knative-operator -l app=knative-operator

# Or rollout restart
kubectl rollout restart deployment -n knative-operator knative-operator
```

### Step 3: Emergency Recovery

If standard fixes don't work:

#### Option A: Rollback to Previous Version

```bash
# Check HelmRelease history
flux get helmreleases -n knative-operator

# Edit HelmRelease to use previous version
kubectl edit helmrelease knative-operator -n knative-operator
# Change spec.chart.spec.version to previous working version

# Reconcile
flux reconcile helmrelease knative-operator -n knative-operator

# Monitor rollback
kubectl get pods -n knative-operator -l app=knative-operator -w
```

#### Option B: Temporarily Suspend and Debug

```bash
# Suspend HelmRelease
flux suspend helmrelease knative-operator -n knative-operator

# Delete crashlooping pod
kubectl delete deployment -n knative-operator knative-operator

# Manually inspect what would be deployed
helm template knative-operator <chart-repo>/knative-operator --version <version> --namespace knative-operator > /tmp/operator-template.yaml

# Review for issues
cat /tmp/operator-template.yaml

# Fix issues in Git repository
# Commit and push

# Resume
flux resume helmrelease knative-operator -n knative-operator
flux reconcile helmrelease knative-operator -n knative-operator
```

#### Option C: Complete Reinstall

⚠️ **Warning:** This will disrupt operator management temporarily!

```bash
# Backup current state
kubectl get helmrelease -n knative-operator knative-operator -o yaml > /tmp/knative-operator-backup.yaml
kubectl get knativeserving -A -o yaml > /tmp/knativeserving-backup.yaml

# Suspend HelmRelease
flux suspend helmrelease knative-operator -n knative-operator

# Delete all operator resources
kubectl delete deployment -n knative-operator --all
kubectl delete service -n knative-operator --all
kubectl delete configmap -n knative-operator --all
kubectl delete secret -n knative-operator --all

# Do NOT delete CRDs if Knative resources exist!
kubectl get knativeserving -A  # Check first

# Resume and reconcile
flux resume helmrelease knative-operator -n knative-operator
flux reconcile helmrelease knative-operator -n knative-operator --force

# Monitor recreation
watch kubectl get pods -n knative-operator
```

### Step 4: Enable Debug Logging

If cause still unclear:

```bash
# Edit deployment to add verbose logging
kubectl edit deployment -n knative-operator knative-operator

# Add or modify args:
# args:
#   - --zap-log-level=debug
#   - --zap-encoder=console

# Or set via environment variable
# env:
# - name: LOG_LEVEL
#   value: "debug"

# Apply and watch logs
kubectl logs -n knative-operator -l app=knative-operator -f
```

## Verification

### 1. Check Pod is Running

```bash
# Pod should be Running and READY
kubectl get pod -n knative-operator -l app=knative-operator

# Expected:
# NAME                                READY   STATUS    RESTARTS   AGE
# knative-operator-xxxxxxxxxx-xxxxx   1/1     Running   0          2m
```

### 2. Verify No Recent Restarts

```bash
# Check restart count stopped increasing
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].restartCount}'

# Monitor for new restarts
watch -n 5 'kubectl get pod -n knative-operator -l app=knative-operator'
```

### 3. Check Logs Show Normal Operation

```bash
# View startup logs
kubectl logs -n knative-operator -l app=knative-operator --tail=50

# Should show:
# - Successful startup
# - No panic or fatal errors
# - Reconciliation activity
# - No repeated error messages
```

### 4. Verify Operator Functionality

```bash
# Check managed resources
kubectl get knativeserving -A
kubectl describe knativeserving knative-serving -n knative-serving

# Trigger reconciliation
kubectl annotate knativeserving knative-serving -n knative-serving test=$(date +%s) --overwrite

# Watch operator processes it
kubectl logs -n knative-operator -l app=knative-operator -f
```

### 5. Check Webhooks Working

```bash
# Test webhook by trying to create KnativeServing
cat <<EOF | kubectl apply --dry-run=server -f -
apiVersion: operator.knative.dev/v1beta1
kind: KnativeServing
metadata:
  name: test-webhook
  namespace: default
EOF

# Should validate successfully or show expected validation errors
```

### 6. Monitor Stability

```bash
# Monitor for 10 minutes
watch -n 30 'kubectl get pod -n knative-operator -l app=knative-operator'

# Check metrics if available
kubectl top pod -n knative-operator -l app=knative-operator
```

## Prevention

### 1. Set Appropriate Resource Limits

```yaml
# Recommended configuration
resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "2Gi"    # Generous limit to prevent OOMKills
    cpu: "1000m"
```

### 2. Implement Liveness and Readiness Probes

```yaml
# Ensure proper health checks
livenessProbe:
  httpGet:
    path: /healthz
    port: 8081
  initialDelaySeconds: 15
  periodSeconds: 20
readinessProbe:
  httpGet:
    path: /readyz
    port: 8081
  initialDelaySeconds: 5
  periodSeconds: 10
```

### 3. Monitor Crash Patterns

Create alerts for crash loops:

```yaml
- alert: KnativeOperatorCrashLoop
  expr: |
    rate(kube_pod_container_status_restarts_total{namespace="knative-operator", pod=~"knative-operator.*"}[15m]) > 0
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Knative Operator is crash looping"
```

### 4. Version Management

```yaml
# Pin to stable versions
spec:
  chart:
    spec:
      version: "1.16.3"  # Use tested versions
```

### 5. Pre-deployment Testing

```bash
# Test configuration before applying
helm template knative-operator <chart-repo>/knative-operator \
  --version <version> \
  --namespace knative-operator \
  --values values.yaml | kubectl apply --dry-run=server -f -
```

### 6. Regular Health Checks

```bash
# Schedule regular monitoring
kubectl get pods -n knative-operator -l app=knative-operator
kubectl top pod -n knative-operator -l app=knative-operator
kubectl get events -n knative-operator --sort-by='.lastTimestamp' | tail -10
```

## Performance Tips

1. **Startup Time**: Operator should start within 30-60 seconds
2. **Memory Usage**: Typical usage 150-300Mi, shouldn't exceed 512Mi regularly
3. **CPU Usage**: Low CPU usage < 100m, spikes to 200-300m during reconciliation
4. **Restart Tolerance**: Zero restarts is ideal, investigate any restart
5. **Log Volume**: Reduce log level to info in production

## Related Alerts

- `KnativeOperatorDown` - May fire simultaneously with crash loop
- `KnativeOperatorHighMemory` - May precede crash loop if OOMKills
- `KnativeOperatorReconciliationFailed` - May occur due to crashes
- `KnativeServingDown` - May occur if operator can't manage components

## Escalation

If crash loop cannot be resolved within 15 minutes:

1. ✅ Verify all resolution steps attempted
2. 📋 Collect complete diagnostics:
   ```bash
   kubectl logs -n knative-operator -l app=knative-operator --previous > /tmp/operator-crash-logs.txt
   kubectl describe pod -n knative-operator -l app=knative-operator > /tmp/operator-pod-describe.txt
   kubectl get events -n knative-operator --sort-by='.lastTimestamp' > /tmp/operator-events.txt
   kubectl get deployment -n knative-operator knative-operator -o yaml > /tmp/operator-deployment.yaml
   ```
3. 🔍 Search for known issues in Knative releases
4. 📞 Contact platform team with diagnostics
5. 🐛 File bug report with Knative project if suspected bug
6. 🆘 Page on-call engineer for critical production impact

**Escalation Criteria:**
- Crash loop persisting > 15 minutes
- Unable to identify root cause from logs
- Production Knative services impacted
- Suspected operator bug or regression

## Additional Resources

- [Knative Operator Troubleshooting](https://knative.dev/docs/install/operator/knative-with-operators/#troubleshooting)
- [Kubernetes Debugging Pods](https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/)
- [Container Exit Codes](https://komodor.com/learn/exit-codes-in-containers-and-kubernetes-the-complete-guide/)
- [Knative Operator GitHub Issues](https://github.com/knative/operator/issues)

## Quick Commands Reference

```bash
# Check pod status
kubectl get pod -n knative-operator -l app=knative-operator

# View crash logs
kubectl logs -n knative-operator -l app=knative-operator --previous

# Check restart count
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].restartCount}'

# Check exit code
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].lastState.terminated.exitCode}'

# Restart operator
kubectl rollout restart deployment -n knative-operator knative-operator

# Force HelmRelease reconcile
flux reconcile helmrelease knative-operator -n knative-operator --force

# Delete pod to force recreation
kubectl delete pod -n knative-operator -l app=knative-operator
```

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Platform Team

