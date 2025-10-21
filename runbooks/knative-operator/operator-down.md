# 🚨 Runbook: Knative Operator Down

## Alert Information

**Alert Name:** `KnativeOperatorDown`  
**Severity:** Critical  
**Component:** knative-operator  
**Service:** knative-operator

## Symptom

Knative Operator is completely unavailable. Cannot manage Knative Serving or Knative Eventing components.

## Impact

- **User Impact:** MEDIUM - Existing Knative services continue running but cannot be updated
- **Business Impact:** HIGH - Cannot deploy new services or update configurations
- **Data Impact:** LOW - No data loss, operational impact only

## Diagnosis

### 1. Check Operator Pod Status

```bash
kubectl get pods -n knative-operator
kubectl get pods -n knative-operator -l app=knative-operator -o wide
```

**Expected Output:**
```
NAME                                READY   STATUS    RESTARTS   AGE
knative-operator-xxxxxxxxxx-xxxxx   1/1     Running   0          24h
```

### 2. Check Operator Deployment

```bash
kubectl get deployment -n knative-operator
kubectl describe deployment -n knative-operator knative-operator
```

### 3. Check Recent Events

```bash
kubectl get events -n knative-operator --sort-by='.lastTimestamp' | tail -20
```

### 4. Check Operator Logs

```bash
kubectl logs -n knative-operator -l app=knative-operator --tail=100
kubectl logs -n knative-operator -l app=knative-operator --previous  # If pod restarted
```

### 5. Check Resource Usage

```bash
kubectl top pod -n knative-operator
kubectl describe node $(kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.spec.nodeName}' | head -1)
```

### 6. Check KnativeServing Status

```bash
kubectl get knativeserving -A
kubectl describe knativeserving knative-serving -n knative-serving
```

## Resolution Steps

### Step 1: Identify the Root Cause

Check the pod status and logs to understand why the operator is down:

```bash
# Get detailed pod status
kubectl describe pod -n knative-operator -l app=knative-operator

# Check for OOMKilled
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].lastState.terminated.reason}'

# Check for CrashLoopBackOff
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].status.containerStatuses[0].state.waiting.reason}'
```

### Step 2: Common Issues and Fixes

#### Issue: Pod OOMKilled (Out of Memory)
**Cause:** Operator exceeded memory limits  
**Fix:**
```bash
# Check current memory limits
kubectl get pod -n knative-operator -l app=knative-operator -o jsonpath='{.items[0].spec.containers[0].resources.limits.memory}'

# Check operator logs for memory issues
kubectl logs -n knative-operator -l app=knative-operator --tail=200 | grep -i "memory\|oom"

# Increase memory limits in HelmRelease
kubectl edit helmrelease knative-operator -n knative-operator
# Update:
#   resources:
#     limits:
#       memory: 1Gi  # Increase from 512Mi
#     requests:
#       memory: 256Mi

# Or use Flux to update
flux reconcile helmrelease knative-operator -n knative-operator
```

#### Issue: CRD Installation Failed
**Cause:** Missing or corrupt Custom Resource Definitions  
**Fix:**
```bash
# Check if CRDs exist
kubectl get crd | grep knative

# Check operator logs for CRD errors
kubectl logs -n knative-operator -l app=knative-operator --tail=200 | grep -i "crd\|custom resource"

# Reconcile HelmRelease to reinstall CRDs
flux reconcile helmrelease knative-operator -n knative-operator --force
```

#### Issue: RBAC Permission Denied
**Cause:** Operator lacks necessary permissions  
**Fix:**
```bash
# Check ServiceAccount
kubectl get serviceaccount -n knative-operator

# Check RoleBindings and ClusterRoleBindings
kubectl get rolebinding,clusterrolebinding -A | grep knative-operator

# Check operator logs for permission errors
kubectl logs -n knative-operator -l app=knative-operator --tail=200 | grep -i "forbidden\|unauthorized\|permission"

# Reconcile to restore RBAC
flux reconcile helmrelease knative-operator -n knative-operator
```

#### Issue: Image Pull Error
**Cause:** Cannot pull operator image  
**Fix:**
```bash
# Check image pull status
kubectl describe pod -n knative-operator -l app=knative-operator | grep -A 10 "Events:"

# Verify image exists
kubectl get deployment -n knative-operator knative-operator -o jsonpath='{.spec.template.spec.containers[0].image}'

# Check image pull secrets
kubectl get secrets -n knative-operator | grep docker

# Manual pull test (if using private registry)
kubectl run test --image=<operator-image> --rm -it --restart=Never -n knative-operator -- version
```

#### Issue: Configuration Error
**Cause:** Invalid operator configuration  
**Fix:**
```bash
# Check HelmRelease configuration
kubectl get helmrelease -n knative-operator knative-operator -o yaml

# Check for validation errors
flux get helmreleases -n knative-operator

# Rollback to previous working version
flux suspend helmrelease knative-operator -n knative-operator
flux resume helmrelease knative-operator -n knative-operator
```

### Step 3: Force Operator Restart

If no clear issue found, force restart:

```bash
# Rollout restart deployment
kubectl rollout restart deployment -n knative-operator knative-operator

# Wait for deployment to be ready
kubectl rollout status deployment -n knative-operator knative-operator --timeout=5m

# Check pod status
kubectl get pods -n knative-operator -l app=knative-operator
```

### Step 4: Force Flux Reconciliation

```bash
# Reconcile Flux HelmRelease
flux reconcile helmrelease knative-operator -n knative-operator --force

# Check reconciliation status
flux get helmreleases -n knative-operator

# Check Flux logs
flux logs -n flux-system
```

### Step 5: Emergency Recovery - Reinstall Operator

⚠️ **Warning:** This will temporarily disrupt operator functionality but existing services will continue running!

```bash
# Suspend Flux HelmRelease
flux suspend helmrelease knative-operator -n knative-operator

# Delete deployment (keeps CRDs and managed resources)
kubectl delete deployment -n knative-operator knative-operator

# Resume Flux to recreate
flux resume helmrelease knative-operator -n knative-operator
flux reconcile helmrelease knative-operator -n knative-operator

# Monitor recreation
watch kubectl get pods -n knative-operator
```

## Verification

### 1. Check Operator Pod is Running

```bash
kubectl get pod -n knative-operator -l app=knative-operator
# Should show: Running and 1/1 READY
```

### 2. Check Operator Logs

```bash
kubectl logs -n knative-operator -l app=knative-operator --tail=50
# Should show successful reconciliation messages
```

### 3. Check Managed Resources

```bash
# Check KnativeServing
kubectl get knativeserving -A
kubectl describe knativeserving knative-serving -n knative-serving

# Verify Ready status
kubectl get knativeserving knative-serving -n knative-serving -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
# Should return: True
```

### 4. Test Operator Functionality

```bash
# Annotate KnativeServing to trigger reconciliation
kubectl annotate knativeserving knative-serving -n knative-serving test-reconcile=$(date +%s) --overwrite

# Watch operator logs for reconciliation
kubectl logs -n knative-operator -l app=knative-operator -f
```

### 5. Verify Knative Services Still Working

```bash
# List Knative services
kubectl get ksvc -A

# Check a sample service
kubectl describe ksvc <service-name> -n <namespace>
```

## Prevention

### 1. Resource Management

```yaml
# In Knative Operator HelmRelease, configure proper resources
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

### 2. Set Up Monitoring

- Monitor operator pod availability
- Alert on operator restarts
- Monitor resource usage
- Track reconciliation failures

### 3. Pod Disruption Budget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: knative-operator-pdb
  namespace: knative-operator
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: knative-operator
```

### 4. Regular Health Checks

```bash
# Schedule regular operator health checks
kubectl get pods -n knative-operator -l app=knative-operator
kubectl get knativeserving -A
```

## Performance Tips

1. **Resource Allocation**: Ensure adequate CPU and memory for large clusters
2. **CRD Management**: Keep CRDs up to date with operator version
3. **RBAC**: Regularly audit operator permissions
4. **Version Compatibility**: Ensure operator version matches Knative versions

## Related Alerts

- `KnativeOperatorCrashLoop`
- `KnativeOperatorHighMemory`
- `KnativeOperatorReconciliationFailed`
- `KnativeServingDown`
- `KnativeEventingDown`

## Escalation

If operator cannot be restored within 15 minutes:

1. ✅ Check all resolution steps above
2. 🔍 Review HelmRelease and deployment configuration
3. 📊 Analyze node health and cluster capacity
4. 🔄 Consider temporary workaround: manual CRD updates
5. 📞 Contact platform team
6. 🆘 Page on-call engineer for critical production impact

## Additional Resources

- [Knative Operator Documentation](https://knative.dev/docs/install/operator/knative-with-operators/)
- [Knative Operator GitHub](https://github.com/knative/operator)
- [Kubernetes Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
- [Flux HelmRelease Documentation](https://fluxcd.io/docs/components/helm/)

## Quick Commands Reference

```bash
# Health check
kubectl get pods -n knative-operator

# Get operator version
kubectl get deployment -n knative-operator knative-operator -o jsonpath='{.spec.template.spec.containers[0].image}'

# View logs
kubectl logs -n knative-operator -l app=knative-operator --tail=100

# Check managed resources
kubectl get knativeserving -A

# Restart operator
kubectl rollout restart deployment -n knative-operator knative-operator

# Force reconcile
flux reconcile helmrelease knative-operator -n knative-operator
```

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Platform Team

