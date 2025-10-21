# 🚨 Runbook: Knative Controller Down

## Alert Information

**Alert Name:** `KnativeControllerDown`  
**Severity:** Critical  
**Component:** knative-serving / controller  
**Impact:** Cannot create or update Knative services

## Symptom

The Knative Controller component is unavailable or not responding. New Knative services cannot be created, and existing services cannot be updated. Revisions, routes, and configurations are not being reconciled.

## Impact

- **User Impact:** CRITICAL - Cannot deploy new services or update existing ones
- **Business Impact:** CRITICAL - Deployment pipeline blocked
- **Data Impact:** LOW - No data loss, existing services continue to run

## Diagnosis

### 1. Check Controller Pod Status

```bash
kubectl get pods -n knative-serving -l app=controller
kubectl get pods -n knative-serving -l app=controller -o wide
```

**Expected Output:**
```
NAME                          READY   STATUS    RESTARTS   AGE
controller-xxxxxxxxxx-xxxxx   1/1     Running   0          24h
```

### 2. Check Controller Deployment

```bash
kubectl describe deployment -n knative-serving controller
kubectl get deployment -n knative-serving controller -o yaml
```

### 3. Check Controller Logs

```bash
# Recent logs
kubectl logs -n knative-serving -l app=controller --tail=100

# Previous container logs (if crashed)
kubectl logs -n knative-serving -l app=controller --tail=100 --previous
```

### 4. Check Resource Status

```bash
# Check Knative services
kubectl get ksvc -A

# Check revisions
kubectl get revision -A

# Check routes
kubectl get route -A

# Check configurations
kubectl get configuration -A
```

### 5. Check Resource Usage

```bash
kubectl top pods -n knative-serving -l app=controller
```

### 6. Check Recent Events

```bash
kubectl get events -n knative-serving --field-selector involvedObject.name=controller --sort-by='.lastTimestamp'
```

### 7. Check Leader Election

```bash
# Controller uses leader election
kubectl get lease -n knative-serving | grep controller
kubectl describe lease -n knative-serving <controller-lease-name>
```

## Resolution Steps

### Step 1: Identify Root Cause

Check pod status for common issues:

```bash
# Get pod details
kubectl describe pod -n knative-serving -l app=controller

# Common indicators:
# - CrashLoopBackOff: Application or API server connectivity error
# - OOMKilled: Out of memory
# - Error logs: Permission or RBAC issues
```

### Step 2: Common Issues and Fixes

#### Issue: Pod CrashLoopBackOff
**Cause:** Application error or API server connectivity  
**Fix:**
```bash
# Check logs for specific errors
kubectl logs -n knative-serving -l app=controller --tail=200

# Common errors to look for:
# - "connection refused" -> API server issues
# - "permission denied" -> RBAC issues
# - "timeout" -> Resource constraints

# Check API server connectivity
kubectl cluster-info

# Restart controller
kubectl rollout restart deployment -n knative-serving controller
```

#### Issue: RBAC/Permission Errors
**Cause:** Missing or incorrect RBAC permissions  
**Fix:**
```bash
# Check controller service account
kubectl get sa -n knative-serving controller -o yaml

# Check controller role and rolebinding
kubectl get clusterrole | grep knative-serving-controller
kubectl describe clusterrole knative-serving-controller-admin

# Check rolebindings
kubectl get clusterrolebinding | grep knative-serving-controller
kubectl describe clusterrolebinding knative-serving-controller-admin

# If RBAC is missing, reconcile via operator
kubectl annotate knativeserving knative-serving -n knative-serving reconcile=$(date +%s) --overwrite
```

#### Issue: Pod OOMKilled
**Cause:** Insufficient memory allocation  
**Fix:**
```bash
# Check current resource limits
kubectl get deployment -n knative-serving controller -o yaml | grep -A 10 resources

# Increase memory limits
kubectl patch deployment -n knative-serving controller -p '{"spec":{"template":{"spec":{"containers":[{"name":"controller","resources":{"limits":{"memory":"1Gi"},"requests":{"memory":"512Mi"}}}]}}}}'

# Wait for rollout
kubectl rollout status deployment -n knative-serving controller
```

#### Issue: Leader Election Failure
**Cause:** Cannot acquire leader election lease  
**Fix:**
```bash
# Check leader election lease
kubectl get lease -n knative-serving | grep controller
kubectl describe lease -n knative-serving <controller-lease>

# Delete stale lease if needed (will be recreated)
kubectl delete lease -n knative-serving <controller-lease>

# Restart controller
kubectl rollout restart deployment -n knative-serving controller
```

#### Issue: Webhook Not Responding
**Cause:** Controller needs webhook for validation  
**Fix:**
```bash
# Check webhook is running
kubectl get pods -n knative-serving -l app=webhook

# Check webhook service
kubectl get svc -n knative-serving webhook

# Restart webhook
kubectl rollout restart deployment -n knative-serving webhook

# Wait for webhook to be ready
kubectl rollout status deployment -n knative-serving webhook
```

#### Issue: Reconciliation Stuck
**Cause:** Resources stuck in pending state  
**Fix:**
```bash
# Check for stuck resources
kubectl get ksvc -A | grep -v Running
kubectl get revision -A | grep -v Ready

# Check specific resource status
kubectl describe ksvc <service-name> -n <namespace>

# Force reconciliation
kubectl annotate ksvc <service-name> -n <namespace> reconcile=$(date +%s) --overwrite

# Or restart controller to trigger reconciliation
kubectl rollout restart deployment -n knative-serving controller
```

### Step 3: Restart Controller

If no specific issue identified:

```bash
# Restart controller deployment
kubectl rollout restart deployment -n knative-serving controller

# Watch rollout progress
kubectl rollout status deployment -n knative-serving controller

# Verify new pod is running
kubectl get pods -n knative-serving -l app=controller
```

### Step 4: Verify CRD Installation

```bash
# Check Knative Serving CRDs are installed
kubectl get crd | grep knative.dev

# Expected CRDs:
# - services.serving.knative.dev
# - configurations.serving.knative.dev
# - revisions.serving.knative.dev
# - routes.serving.knative.dev

# If CRDs are missing, reinstall via operator
kubectl annotate knativeserving knative-serving -n knative-serving reconcile=$(date +%s) --overwrite
```

### Step 5: Check API Server Health

```bash
# Controller needs to communicate with API server
kubectl cluster-info
kubectl get --raw /healthz
kubectl get --raw /readyz

# Check controller can list resources
kubectl auth can-i list services.serving.knative.dev --as=system:serviceaccount:knative-serving:controller
```

## Verification

### 1. Check Controller is Running

```bash
kubectl get pods -n knative-serving -l app=controller
# Should show Running status with 1/1 READY
```

### 2. Check Controller Logs

```bash
kubectl logs -n knative-serving -l app=controller --tail=50
# Should show reconciliation events, no errors
```

### 3. Test Service Creation

```bash
# Create test service
cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: controller-test
  namespace: default
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
        env:
        - name: TARGET
          value: "Controller Test"
EOF

# Wait for service to be ready
kubectl wait --for=condition=Ready ksvc/controller-test -n default --timeout=60s

# Check service status
kubectl get ksvc controller-test -n default

# Verify revision created
kubectl get revision -n default -l serving.knative.dev/service=controller-test

# Verify route created
kubectl get route -n default -l serving.knative.dev/service=controller-test

# Verify configuration created
kubectl get configuration -n default -l serving.knative.dev/service=controller-test

# Test service
URL=$(kubectl get ksvc controller-test -n default -o jsonpath='{.status.url}')
curl $URL

# Cleanup
kubectl delete ksvc controller-test -n default
```

### 4. Test Service Update

```bash
# Create service
kubectl apply -f - <<EOF
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: update-test
  namespace: default
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
        env:
        - name: TARGET
          value: "v1"
EOF

# Wait for ready
kubectl wait --for=condition=Ready ksvc/update-test -n default --timeout=60s

# Update service
kubectl apply -f - <<EOF
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: update-test
  namespace: default
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
        env:
        - name: TARGET
          value: "v2"
EOF

# Wait for update
kubectl wait --for=condition=Ready ksvc/update-test -n default --timeout=60s

# Verify new revision created
kubectl get revision -n default -l serving.knative.dev/service=update-test

# Should show 2 revisions
# Cleanup
kubectl delete ksvc update-test -n default
```

### 5. Check Leader Election

```bash
kubectl get lease -n knative-serving | grep controller
# Should show active lease holder
```

## Prevention

### 1. Resource Management

Ensure adequate resources:

```yaml
# In KnativeServing CR or deployment
resources:
  requests:
    cpu: 100m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi
```

### 2. High Availability

For production, run multiple replicas:

```bash
# Scale controller for HA (with leader election)
kubectl scale deployment -n knative-serving controller --replicas=2
```

### 3. Monitoring Setup

Key metrics to monitor:
- Controller pod availability
- Reconciliation success/failure rate
- Reconciliation latency
- Resource creation/update events
- Leader election status

### 4. RBAC Health

Regularly verify RBAC permissions:

```bash
# Verify controller permissions
kubectl auth can-i list services.serving.knative.dev --as=system:serviceaccount:knative-serving:controller
kubectl auth can-i update services.serving.knative.dev --as=system:serviceaccount:knative-serving:controller
```

### 5. Pod Disruption Budget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: controller-pdb
  namespace: knative-serving
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: controller
```

## Performance Tips

1. **Multiple Replicas**: Run 2+ replicas in production for HA (leader election handles coordination)
2. **Resource Allocation**: Provide adequate CPU/memory for large clusters
3. **API Server**: Ensure controller can efficiently communicate with API server
4. **Watch Caching**: Controller uses watch caching for efficiency
5. **Rate Limiting**: Configure appropriate API rate limits

## Related Alerts

- `KnativeServingDown`
- `KnativeWebhookDown`
- `KnativeServingScalingIssues`

## Escalation

If controller cannot be restored within 15 minutes:

1. ✅ Verify all resolution steps completed
2. 🔍 Check API server health and connectivity
3. 📊 Review RBAC and permissions
4. 🔄 Check cluster control plane health
5. 📞 Escalate to platform team
6. 🆘 Page on-call engineer if deployments are blocked

## Additional Resources

- [Knative Controller Documentation](https://knative.dev/docs/serving/architecture/#controller)
- [Knative Serving API](https://knative.dev/docs/serving/spec/knative-api-specification-1.0/)
- [Knative Troubleshooting](https://knative.dev/docs/serving/troubleshooting/)

## Quick Commands Reference

```bash
# Check controller status
kubectl get pods -n knative-serving -l app=controller

# View controller logs
kubectl logs -n knative-serving -l app=controller --tail=100

# Restart controller
kubectl rollout restart deployment -n knative-serving controller

# Check Knative resources
kubectl get ksvc -A
kubectl get revision -A
kubectl get route -A

# Check leader election
kubectl get lease -n knative-serving | grep controller

# Force reconciliation
kubectl annotate ksvc <service-name> -n <namespace> reconcile=$(date +%s) --overwrite

# Check resource usage
kubectl top pods -n knative-serving -l app=controller
```

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Platform Team

