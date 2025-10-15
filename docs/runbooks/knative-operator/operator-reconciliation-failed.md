# ⚠️ Runbook: Knative Operator Reconciliation Failed

## Alert Information

**Alert Name:** `KnativeOperatorReconciliationFailed`  
**Severity:** Warning  
**Component:** knative-operator  
**Service:** knative-operator

## Symptom

Knative Operator is unable to reconcile KnativeServing or KnativeEventing resources. Changes to Knative components are not being applied.

## Impact

- **User Impact:** MEDIUM - Existing services run normally but updates don't apply
- **Business Impact:** HIGH - Cannot deploy updates or changes to Knative components
- **Data Impact:** LOW - No data loss, deployment/update impact only

## Diagnosis

### 1. Check KnativeServing Status

```bash
# List all KnativeServing resources
kubectl get knativeserving -A

# Check detailed status
kubectl describe knativeserving knative-serving -n knative-serving

# Get status conditions
kubectl get knativeserving knative-serving -n knative-serving -o jsonpath='{.status.conditions}' | jq .
```

**Expected Output:**
```json
[
  {
    "type": "Ready",
    "status": "True",
    "reason": "InstallSucceeded"
  }
]
```

### 2. Check KnativeEventing Status

```bash
# List all KnativeEventing resources
kubectl get knativeeventing -A

# Check detailed status
kubectl describe knativeeventing knative-eventing -n knative-eventing

# Get status conditions
kubectl get knativeeventing knative-eventing -n knative-eventing -o jsonpath='{.status.conditions}' | jq .
```

### 3. Check Operator Logs

```bash
# View operator logs for reconciliation errors
kubectl logs -n knative-operator -l app=knative-operator --tail=200

# Search for specific errors
kubectl logs -n knative-operator -l app=knative-operator --tail=500 | grep -i "error\|failed\|reconcile"

# Check for CRD errors
kubectl logs -n knative-operator -l app=knative-operator --tail=500 | grep -i "crd\|custom resource"
```

### 4. Check Operator Events

```bash
# Get recent events in operator namespace
kubectl get events -n knative-operator --sort-by='.lastTimestamp' | tail -20

# Get events for KnativeServing namespace
kubectl get events -n knative-serving --sort-by='.lastTimestamp' | tail -20

# Get events for KnativeEventing namespace (if applicable)
kubectl get events -n knative-eventing --sort-by='.lastTimestamp' | tail -20
```

### 5. Check CRD Status

```bash
# Verify CRDs exist
kubectl get crd | grep knative

# Check specific CRDs
kubectl get crd knativeservings.operator.knative.dev
kubectl get crd knativeeventings.operator.knative.dev

# Describe CRD for issues
kubectl describe crd knativeservings.operator.knative.dev
```

### 6. Check Component Deployments

```bash
# Check Knative Serving components
kubectl get deployments -n knative-serving

# Check pod status
kubectl get pods -n knative-serving

# Check for failed pods
kubectl get pods -n knative-serving --field-selector=status.phase!=Running
```

## Resolution Steps

### Step 1: Identify Reconciliation Failure Type

#### Check Status Conditions

```bash
# Get failure reason
kubectl get knativeserving knative-serving -n knative-serving -o jsonpath='{.status.conditions[?(@.type=="Ready")]}' | jq .

# Common failure reasons:
# - "InstallFailed" - Installation or update failed
# - "ComponentNotReady" - One or more components not ready
# - "VersionMigrationFailed" - Failed to upgrade version
# - "DependenciesInstalling" - Dependencies not ready
```

#### Check Operator Logs for Specific Error

```bash
# Search for reconciliation errors
kubectl logs -n knative-operator -l app=knative-operator --tail=500 | grep -A 10 "reconcile.*failed"

# Look for resource creation failures
kubectl logs -n knative-operator -l app=knative-operator --tail=500 | grep -i "failed to create\|failed to update"
```

### Step 2: Common Issues and Fixes

#### Issue 1: CRD Version Mismatch

**Cause:** Knative CRDs don't match operator version  
**Fix:**

```bash
# Check operator version
kubectl get deployment -n knative-operator knative-operator -o jsonpath='{.spec.template.spec.containers[0].image}'

# Check CRD versions
kubectl get crd knativeservings.operator.knative.dev -o yaml | grep "version:"

# Update CRDs by reconciling HelmRelease
flux reconcile helmrelease knative-operator -n knative-operator --force

# Verify CRDs updated
kubectl get crd knativeservings.operator.knative.dev -o yaml | grep -A 5 "versions:"
```

#### Issue 2: Invalid Configuration

**Cause:** Invalid spec in KnativeServing/KnativeEventing  
**Fix:**

```bash
# Get current configuration
kubectl get knativeserving knative-serving -n knative-serving -o yaml > /tmp/knative-serving.yaml

# Check for validation errors
kubectl apply --dry-run=server -f /tmp/knative-serving.yaml

# Review and fix configuration
kubectl edit knativeserving knative-serving -n knative-serving

# Common issues to check:
# - Invalid resource names
# - Incorrect version specifications
# - Invalid configuration values
# - Missing required fields
```

#### Issue 3: Resource Conflicts

**Cause:** Conflicting resources or ownership  
**Fix:**

```bash
# Check for conflicting resources
kubectl get all -n knative-serving -o yaml | grep -i "owner\|managed"

# Look for resources not managed by operator
kubectl get deployments -n knative-serving -o json | jq -r '.items[] | select(.metadata.ownerReferences == null) | .metadata.name'

# Remove conflicting resources if safe
kubectl delete deployment <conflicting-deployment> -n knative-serving

# Trigger reconciliation
kubectl annotate knativeserving knative-serving -n knative-serving reconcile=$(date +%s) --overwrite
```

#### Issue 4: Component Image Pull Failures

**Cause:** Cannot pull Knative component images  
**Fix:**

```bash
# Check for ImagePullBackOff
kubectl get pods -n knative-serving | grep -i "imagepull"

# Describe failing pods
kubectl describe pod -n knative-serving <pod-name>

# Check image pull secrets
kubectl get secrets -n knative-serving | grep docker

# Verify image exists and is accessible
kubectl run test --image=<failing-image> --rm -it --restart=Never -n knative-serving -- sh -c "echo success"
```

#### Issue 5: RBAC Permission Issues

**Cause:** Operator lacks permissions to create resources  
**Fix:**

```bash
# Check operator permissions
kubectl auth can-i create deployments --namespace=knative-serving --as=system:serviceaccount:knative-operator:knative-operator

# Check ClusterRoleBindings
kubectl get clusterrolebinding | grep knative-operator

# Check RoleBindings
kubectl get rolebinding -n knative-serving | grep knative

# Reconcile to restore RBAC
flux reconcile helmrelease knative-operator -n knative-operator
```

#### Issue 6: Resource Quota Exceeded

**Cause:** Namespace resource quotas preventing resource creation  
**Fix:**

```bash
# Check resource quotas
kubectl get resourcequota -n knative-serving
kubectl describe resourcequota -n knative-serving

# Check limit ranges
kubectl get limitrange -n knative-serving
kubectl describe limitrange -n knative-serving

# Adjust quotas if needed
kubectl edit resourcequota -n knative-serving
```

### Step 3: Force Reconciliation

```bash
# Method 1: Annotate resource to trigger reconciliation
kubectl annotate knativeserving knative-serving -n knative-serving reconcile=$(date +%s) --overwrite

# Method 2: Restart operator
kubectl rollout restart deployment -n knative-operator knative-operator
kubectl rollout status deployment -n knative-operator knative-operator

# Method 3: Suspend and resume HelmRelease
flux suspend helmrelease knative-operator -n knative-operator
flux resume helmrelease knative-operator -n knative-operator
flux reconcile helmrelease knative-operator -n knative-operator

# Monitor reconciliation
kubectl logs -n knative-operator -l app=knative-operator -f
```

### Step 4: Reset to Known Good State

⚠️ **Warning:** This will briefly interrupt operator management!

```bash
# Backup current configuration
kubectl get knativeserving knative-serving -n knative-serving -o yaml > /tmp/knativeserving-backup.yaml
kubectl get knativeeventing knative-eventing -n knative-eventing -o yaml > /tmp/knativeeventing-backup.yaml

# Delete and recreate (if safe and non-production)
kubectl delete knativeserving knative-serving -n knative-serving
kubectl delete knativeeventing knative-eventing -n knative-eventing

# Reconcile HelmRelease to recreate
flux reconcile helmrelease knative-operator -n knative-operator

# Wait for resources to be recreated
watch kubectl get knativeserving -A
watch kubectl get knativeeventing -A

# Verify status
kubectl get knativeserving knative-serving -n knative-serving -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
```

### Step 5: Upgrade/Downgrade Operator Version

If specific version has issues:

```bash
# Check current version
kubectl get helmrelease -n knative-operator knative-operator -o jsonpath='{.spec.chart.spec.version}'

# List available versions
helm search repo knative-operator --versions

# Edit HelmRelease to change version
kubectl edit helmrelease knative-operator -n knative-operator
# Update spec.chart.spec.version

# Reconcile to apply
flux reconcile helmrelease knative-operator -n knative-operator

# Monitor upgrade
kubectl get pods -n knative-operator -w
```

## Verification

### 1. Check KnativeServing Status is Ready

```bash
# Check Ready condition
kubectl get knativeserving knative-serving -n knative-serving -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
# Expected: True

# Full status check
kubectl get knativeserving -A
# All should show Ready: True

# Detailed status
kubectl describe knativeserving knative-serving -n knative-serving
```

### 2. Check All Components Running

```bash
# Check Knative Serving components
kubectl get deployments -n knative-serving
kubectl get pods -n knative-serving

# All pods should be Running and READY
# Expected components:
# - activator
# - autoscaler
# - controller
# - webhook
# - net-kourier-controller (if using Kourier)
```

### 3. Check Operator Logs Show Success

```bash
# View recent operator logs
kubectl logs -n knative-operator -l app=knative-operator --tail=100

# Should show successful reconciliation messages like:
# "Reconciled KnativeServing"
# "Installation complete"
# "Ready: True"
```

### 4. Test Component Updates

```bash
# Make a test configuration change
kubectl annotate knativeserving knative-serving -n knative-serving test-update=$(date +%s) --overwrite

# Monitor operator processes the change
kubectl logs -n knative-operator -l app=knative-operator -f

# Verify change applied
kubectl get knativeserving knative-serving -n knative-serving -o yaml | grep "test-update"
```

### 5. Test Knative Service Deployment

```bash
# Deploy a test Knative service
cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: test-reconciliation
  namespace: default
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
        env:
        - name: TARGET
          value: "Reconciliation Test"
EOF

# Check service becomes ready
kubectl get ksvc test-reconciliation -n default
kubectl describe ksvc test-reconciliation -n default

# Cleanup test service
kubectl delete ksvc test-reconciliation -n default
```

### 6. Verify No Events Show Errors

```bash
# Check events for success
kubectl get events -n knative-operator --sort-by='.lastTimestamp' | tail -10
kubectl get events -n knative-serving --sort-by='.lastTimestamp' | tail -10

# Should not show reconciliation errors
```

## Prevention

### 1. Version Management

```yaml
# Pin operator and Knative versions in HelmRelease
spec:
  chart:
    spec:
      version: "1.16.3"  # Specific version, not "latest"
  values:
    knative:
      version: "1.16.0"  # Compatible Knative version
```

### 2. Pre-deployment Validation

```bash
# Always dry-run configuration changes
kubectl apply --dry-run=server -f knativeserving-config.yaml

# Validate CRD versions match operator
kubectl get crd knativeservings.operator.knative.dev -o yaml | grep version

# Test in non-production first
```

### 3. Monitoring Reconciliation

Create Prometheus alerts:

```yaml
- alert: KnativeOperatorReconciliationFailed
  expr: |
    kube_customresource_status_condition{
      customresource_group="operator.knative.dev",
      customresource_kind="KnativeServing",
      condition="Ready",
      status="false"
    } == 1
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Knative reconciliation failing for 10m"

- alert: KnativeComponentNotReady
  expr: |
    kube_deployment_status_replicas_available{namespace="knative-serving"} 
    != 
    kube_deployment_spec_replicas{namespace="knative-serving"}
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Knative component not ready"
```

### 4. Regular Health Checks

```bash
# Schedule regular checks
kubectl get knativeserving -A
kubectl get knativeeventing -A
kubectl get pods -n knative-serving
kubectl logs -n knative-operator -l app=knative-operator --tail=50 | grep -i error
```

### 5. Configuration Backup

```bash
# Regularly backup working configurations
kubectl get knativeserving knative-serving -n knative-serving -o yaml > knativeserving-backup-$(date +%Y%m%d).yaml
kubectl get knativeeventing knative-eventing -n knative-eventing -o yaml > knativeeventing-backup-$(date +%Y%m%d).yaml
```

### 6. Document Custom Configurations

Keep documentation of any custom configurations or overrides applied to Knative components.

## Performance Tips

1. **Reconciliation Frequency**: Operator typically reconciles within seconds
2. **Component Startup**: Full component deployment may take 2-5 minutes
3. **Resource Allocation**: Ensure adequate cluster resources for components
4. **Image Caching**: Pre-pull images to reduce deployment time
5. **Version Compatibility**: Always use compatible operator and Knative versions

## Troubleshooting Checklist

- [ ] Operator pod is running and healthy
- [ ] CRD versions match operator version
- [ ] KnativeServing/KnativeEventing configuration is valid
- [ ] No RBAC permission errors in operator logs
- [ ] No image pull errors for component images
- [ ] No resource quota or limit issues
- [ ] No conflicting resources in Knative namespaces
- [ ] Network connectivity to image registries
- [ ] Adequate cluster resources available
- [ ] Compatible Kubernetes version

## Related Alerts

- `KnativeOperatorDown` - Operator not running
- `KnativeOperatorCrashLoop` - Operator crashing repeatedly
- `KnativeServingDown` - Knative Serving components down
- `KnativeComponentNotReady` - Individual components not ready

## Escalation

If reconciliation cannot be restored within 30 minutes:

1. ✅ Verify all resolution steps attempted
2. 📋 Collect diagnostics:
   - Operator logs: `kubectl logs -n knative-operator -l app=knative-operator`
   - KnativeServing status: `kubectl describe knativeserving -A`
   - Component pod status: `kubectl get pods -n knative-serving`
   - Recent events: `kubectl get events -n knative-serving --sort-by='.lastTimestamp'`
3. 🔍 Check for known issues in Knative releases
4. 📞 Contact platform team with diagnostics
5. 🐛 Consider filing bug report if suspected operator issue

**Escalation Criteria:**
- Reconciliation failing for > 30 minutes
- Multiple reconciliation attempts failed
- Impact on production Knative services
- Suspected operator or CRD bug

## Additional Resources

- [Knative Operator Documentation](https://knative.dev/docs/install/operator/knative-with-operators/)
- [KnativeServing CRD Reference](https://knative.dev/docs/install/operator/knative-with-operators/#installing-knative-serving)
- [Knative Troubleshooting Guide](https://knative.dev/docs/serving/troubleshooting/)
- [Kubernetes Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

## Quick Commands Reference

```bash
# Check reconciliation status
kubectl get knativeserving -A
kubectl describe knativeserving knative-serving -n knative-serving

# View operator logs
kubectl logs -n knative-operator -l app=knative-operator --tail=100

# Force reconciliation
kubectl annotate knativeserving knative-serving -n knative-serving reconcile=$(date +%s) --overwrite

# Restart operator
kubectl rollout restart deployment -n knative-operator knative-operator

# Check component health
kubectl get pods -n knative-serving
kubectl get deployments -n knative-serving

# Reconcile HelmRelease
flux reconcile helmrelease knative-operator -n knative-operator
```

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Platform Team

