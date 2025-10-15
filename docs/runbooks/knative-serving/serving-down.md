# 🚨 Runbook: Knative Serving Down

## Alert Information

**Alert Name:** `KnativeServingDown`  
**Severity:** Critical  
**Component:** knative-serving  
**Services:** controller, autoscaler, activator, webhook, kourier-gateway

## Symptom

Knative Serving is completely unavailable. All serverless workloads are down or cannot be accessed.

## Impact

- **User Impact:** CRITICAL - All Knative services unavailable
- **Business Impact:** CRITICAL - Complete serverless platform outage
- **Data Impact:** LOW - No data loss, but service disruption

## Diagnosis

### 1. Check Knative Serving Pods

```bash
kubectl get pods -n knative-serving
kubectl get pods -n knative-serving -o wide
```

**Expected Output:**
```
NAME                                  READY   STATUS    RESTARTS   AGE
activator-xxxxxxxxxx-xxxxx            1/1     Running   0          24h
autoscaler-xxxxxxxxxx-xxxxx           1/1     Running   0          24h
controller-xxxxxxxxxx-xxxxx           1/1     Running   0          24h
webhook-xxxxxxxxxx-xxxxx              1/1     Running   0          24h
3scale-kourier-gateway-xxxxx-xxxxx    1/1     Running   0          24h
```

### 2. Check KnativeServing CR Status

```bash
kubectl get knativeserving -n knative-serving
kubectl describe knativeserving knative-serving -n knative-serving
```

### 3. Check Recent Events

```bash
kubectl get events -n knative-serving --sort-by='.lastTimestamp' | tail -30
```

### 4. Check Component Logs

```bash
# Controller logs
kubectl logs -n knative-serving -l app=controller --tail=100

# Autoscaler logs
kubectl logs -n knative-serving -l app=autoscaler --tail=100

# Activator logs
kubectl logs -n knative-serving -l app=activator --tail=100

# Webhook logs
kubectl logs -n knative-serving -l app=webhook --tail=100

# Kourier logs
kubectl logs -n knative-serving -l app=3scale-kourier-gateway --tail=100
```

### 5. Check Resource Usage

```bash
kubectl top pods -n knative-serving
```

### 6. Check Knative Services

```bash
kubectl get ksvc -A
kubectl get revision -A
kubectl get route -A
```

## Resolution Steps

### Step 1: Identify Which Component is Down

Check each critical component:

```bash
# Check all components
for component in controller autoscaler activator webhook 3scale-kourier-gateway; do
  echo "=== $component ==="
  kubectl get pods -n knative-serving -l app=$component
  kubectl logs -n knative-serving -l app=$component --tail=20
done
```

### Step 2: Common Issues and Fixes

#### Issue: All Pods Down - KnativeServing CR Issues
**Cause:** KnativeServing custom resource not properly configured  
**Fix:**
```bash
# Check KnativeServing status
kubectl describe knativeserving knative-serving -n knative-serving

# Check operator logs
kubectl logs -n knative-operator -l app=knative-operator --tail=100

# Force reconciliation
kubectl annotate knativeserving knative-serving -n knative-serving reconcile=$(date +%s) --overwrite

# If operator is down, restart it first
kubectl rollout restart deployment -n knative-operator knative-operator
```

#### Issue: Controller Pod Down
**Cause:** Controller crashed or OOMKilled  
**Fix:**
```bash
# Check controller status
kubectl describe pod -n knative-serving -l app=controller

# Check logs
kubectl logs -n knative-serving -l app=controller --tail=200 --previous

# Restart controller
kubectl rollout restart deployment -n knative-serving controller

# Wait for ready
kubectl rollout status deployment -n knative-serving controller
```

#### Issue: Autoscaler Pod Down
**Cause:** Autoscaler crashed or misconfigured  
**Fix:**
```bash
# Check autoscaler status
kubectl describe pod -n knative-serving -l app=autoscaler

# Check autoscaler config
kubectl get configmap -n knative-serving config-autoscaler -o yaml

# Restart autoscaler
kubectl rollout restart deployment -n knative-serving autoscaler
```

#### Issue: Activator Pod Down
**Cause:** Activator crashed, preventing scale-from-zero  
**Fix:**
```bash
# Check activator status
kubectl describe pod -n knative-serving -l app=activator

# Check logs for errors
kubectl logs -n knative-serving -l app=activator --tail=200

# Restart activator
kubectl rollout restart deployment -n knative-serving activator

# Verify it's running
kubectl get pods -n knative-serving -l app=activator
```

#### Issue: Webhook Pod Down
**Cause:** Webhook crashed, preventing service creation/updates  
**Fix:**
```bash
# Check webhook status
kubectl describe pod -n knative-serving -l app=webhook

# Check webhook certificates
kubectl get secret -n knative-serving | grep webhook

# Restart webhook
kubectl rollout restart deployment -n knative-serving webhook
```

#### Issue: Kourier Ingress Down
**Cause:** Kourier gateway crashed, no external access  
**Fix:**
```bash
# Check Kourier status
kubectl describe pod -n knative-serving -l app=3scale-kourier-gateway

# Check Kourier service
kubectl get svc -n knative-serving kourier

# Restart Kourier
kubectl rollout restart deployment -n knative-serving 3scale-kourier-gateway
```

#### Issue: Webhook Configuration Broken
**Cause:** Webhook configuration corrupted  
**Fix:**
```bash
# Check webhook configurations
kubectl get validatingwebhookconfiguration | grep knative
kubectl get mutatingwebhookconfiguration | grep knative

# Delete and recreate (will be auto-recreated)
kubectl delete validatingwebhookconfiguration config.webhook.serving.knative.dev
kubectl delete mutatingwebhookconfiguration webhook.serving.knative.dev

# Wait for recreation
sleep 30
kubectl get validatingwebhookconfiguration | grep knative
```

### Step 3: Restart All Components

If multiple components are down:

```bash
# Restart all Knative Serving deployments
kubectl rollout restart deployment -n knative-serving

# Wait for all to be ready
kubectl rollout status deployment -n knative-serving controller
kubectl rollout status deployment -n knative-serving autoscaler
kubectl rollout status deployment -n knative-serving activator
kubectl rollout status deployment -n knative-serving webhook
kubectl rollout status deployment -n knative-serving 3scale-kourier-gateway
```

### Step 4: Force KnativeServing Reconciliation

```bash
# Annotate to trigger reconciliation
kubectl annotate knativeserving knative-serving -n knative-serving reconcile=$(date +%s) --overwrite

# Watch operator reconcile
kubectl logs -n knative-operator -l app=knative-operator -f
```

### Step 5: Emergency Recovery - Reinstall KnativeServing

⚠️ **Warning:** This may cause temporary service disruption!

```bash
# Delete KnativeServing CR (this will delete all Knative Serving components)
kubectl delete knativeserving knative-serving -n knative-serving

# Wait for cleanup
sleep 30

# Recreate KnativeServing CR
kubectl apply -f /path/to/knativeserving.yaml

# Or let Flux reconcile
flux reconcile kustomization infrastructure
```

## Verification

### 1. Check All Pods are Running

```bash
kubectl get pods -n knative-serving
# All pods should be Running and READY
```

### 2. Check KnativeServing Status

```bash
kubectl get knativeserving knative-serving -n knative-serving
# Status should be Ready

kubectl get knativeserving knative-serving -n knative-serving -o jsonpath='{.status.conditions[?(@.type=="Ready")]}'
```

### 3. Test Service Creation

```bash
# Deploy test service
cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: hello-test
  namespace: default
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
        env:
        - name: TARGET
          value: "Test"
EOF

# Check service status
kubectl get ksvc hello-test -n default -w

# Get URL
kubectl get ksvc hello-test -n default -o jsonpath='{.status.url}'

# Test service (replace URL)
curl $(kubectl get ksvc hello-test -n default -o jsonpath='{.status.url}')

# Cleanup
kubectl delete ksvc hello-test -n default
```

### 4. Check Existing Services

```bash
# List all Knative services
kubectl get ksvc -A

# Check a sample service status
kubectl describe ksvc <service-name> -n <namespace>

# Verify service is accessible
kubectl get route <service-name> -n <namespace> -o jsonpath='{.status.url}'
```

### 5. Verify Autoscaling

```bash
# Check PodAutoscalers
kubectl get pa -A

# Verify autoscaler is making decisions
kubectl logs -n knative-serving -l app=autoscaler --tail=50
```

## Prevention

### 1. Resource Management

Ensure adequate resources for all components:

```yaml
# Example resource configuration
controller:
  resources:
    requests:
      cpu: 100m
      memory: 100Mi
    limits:
      cpu: 1000m
      memory: 1Gi
```

### 2. Set Up Monitoring

- Monitor all component pod availability
- Alert on pod restarts
- Monitor resource usage (CPU/memory)
- Track service creation/update failures
- Monitor autoscaler metrics

### 3. Pod Disruption Budgets

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: knative-serving-pdb
  namespace: knative-serving
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: knative-serving
```

### 4. High Availability Configuration

For production, consider:
- Multiple replicas for critical components
- Anti-affinity rules to spread pods across nodes
- Resource quotas and limits
- Regular health checks

## Performance Tips

1. **Component Scaling**: Scale critical components (activator, autoscaler) for high load
2. **Resource Allocation**: Provide adequate resources for controller and webhook
3. **Monitoring**: Use Prometheus metrics for observability
4. **Configuration**: Tune autoscaler parameters for your workload

## Related Alerts

- `KnativeControllerDown`
- `KnativeAutoscalerDown`
- `KnativeActivatorDown`
- `KnativeWebhookDown`
- `KnativeKourierDown`
- `KnativeServingScalingIssues`

## Escalation

If Knative Serving cannot be restored within 15 minutes:

1. ✅ Check all resolution steps above
2. 🔍 Review KnativeServing CR and operator logs
3. 📊 Analyze node health and cluster capacity
4. 🔄 Consider emergency failover or reinstallation
5. 📞 Contact platform team immediately
6. 🆘 Page on-call engineer for critical production impact

## Additional Resources

- [Knative Serving Documentation](https://knative.dev/docs/serving/)
- [Knative Troubleshooting Guide](https://knative.dev/docs/serving/troubleshooting/)
- [Knative GitHub Issues](https://github.com/knative/serving/issues)
- [Kourier Documentation](https://github.com/knative-extensions/net-kourier)

## Quick Commands Reference

```bash
# Health check all components
kubectl get pods -n knative-serving

# Check KnativeServing CR
kubectl get knativeserving -n knative-serving

# View controller logs
kubectl logs -n knative-serving -l app=controller --tail=100

# Restart all components
kubectl rollout restart deployment -n knative-serving

# Force reconciliation
kubectl annotate knativeserving knative-serving -n knative-serving reconcile=$(date +%s) --overwrite

# List Knative services
kubectl get ksvc -A

# Test connectivity
curl $(kubectl get ksvc <service-name> -n <namespace> -o jsonpath='{.status.url}')
```

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Platform Team

