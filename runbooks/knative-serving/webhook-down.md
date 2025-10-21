# 🚨 Runbook: Knative Webhook Down

## Alert Information

**Alert Name:** `KnativeWebhookDown`  
**Severity:** Critical  
**Component:** knative-serving / webhook  
**Impact:** Cannot create or update Knative resources

## Symptom

The Knative Webhook component is unavailable or not responding. Webhook provides admission control (validation and mutation) for Knative resources. When down, Knative services cannot be created or updated.

## Impact

- **User Impact:** CRITICAL - Cannot create/update Knative services
- **Business Impact:** CRITICAL - Deployment pipeline blocked
- **Data Impact:** LOW - No data loss, existing services unaffected

## Diagnosis

### 1. Check Webhook Pod Status

```bash
kubectl get pods -n knative-serving -l app=webhook
kubectl get pods -n knative-serving -l app=webhook -o wide
```

**Expected Output:**
```
NAME                       READY   STATUS    RESTARTS   AGE
webhook-xxxxxxxxxx-xxxxx   1/1     Running   0          24h
```

### 2. Check Webhook Deployment

```bash
kubectl describe deployment -n knative-serving webhook
kubectl get deployment -n knative-serving webhook -o yaml
```

### 3. Check Webhook Logs

```bash
# Recent logs
kubectl logs -n knative-serving -l app=webhook --tail=100

# Previous container logs (if crashed)
kubectl logs -n knative-serving -l app=webhook --tail=100 --previous
```

### 4. Check Webhook Service

```bash
kubectl get svc -n knative-serving webhook
kubectl describe svc -n knative-serving webhook
kubectl get endpoints -n knative-serving webhook
```

### 5. Check Webhook Configurations

```bash
# Validating webhook configurations
kubectl get validatingwebhookconfiguration | grep knative

# Mutating webhook configurations
kubectl get mutatingwebhookconfiguration | grep knative

# Describe specific configurations
kubectl describe validatingwebhookconfiguration config.webhook.serving.knative.dev
kubectl describe mutatingwebhookconfiguration webhook.serving.knative.dev
```

### 6. Check Webhook Certificates

```bash
# Check webhook TLS secret
kubectl get secret -n knative-serving | grep webhook

# Describe webhook certificate
kubectl describe secret -n knative-serving webhook-certs
```

### 7. Check Resource Usage

```bash
kubectl top pods -n knative-serving -l app=webhook
```

### 8. Check Recent Events

```bash
kubectl get events -n knative-serving --field-selector involvedObject.name=webhook --sort-by='.lastTimestamp'
```

## Resolution Steps

### Step 1: Identify Root Cause

Check pod status for common issues:

```bash
# Get pod details
kubectl describe pod -n knative-serving -l app=webhook

# Common indicators:
# - CrashLoopBackOff: Certificate or configuration error
# - ImagePullBackOff: Image issue
# - Running but not Ready: Service/endpoints issue
```

### Step 2: Common Issues and Fixes

#### Issue: Pod CrashLoopBackOff
**Cause:** Certificate error or configuration issue  
**Fix:**
```bash
# Check logs for specific errors
kubectl logs -n knative-serving -l app=webhook --tail=200

# Common errors:
# - Certificate errors
# - Port binding issues
# - API server connectivity

# Check webhook certificates
kubectl get secret -n knative-serving webhook-certs

# Restart webhook
kubectl rollout restart deployment -n knative-serving webhook
```

#### Issue: Certificate Expired or Invalid
**Cause:** Webhook TLS certificates expired  
**Fix:**
```bash
# Check certificate expiration
kubectl get secret -n knative-serving webhook-certs -o yaml

# Delete certificate secret (will be auto-recreated)
kubectl delete secret -n knative-serving webhook-certs

# Restart webhook to generate new certs
kubectl rollout restart deployment -n knative-serving webhook

# Wait for new certificate
sleep 30
kubectl get secret -n knative-serving webhook-certs
```

#### Issue: Webhook Configuration Missing/Broken
**Cause:** WebhookConfiguration resources deleted or corrupted  
**Fix:**
```bash
# Check if configurations exist
kubectl get validatingwebhookconfiguration | grep knative
kubectl get mutatingwebhookconfiguration | grep knative

# If missing, delete and recreate via operator
kubectl annotate knativeserving knative-serving -n knative-serving reconcile=$(date +%s) --overwrite

# Or manually delete (will be auto-recreated)
kubectl delete validatingwebhookconfiguration config.webhook.serving.knative.dev
kubectl delete mutatingwebhookconfiguration webhook.serving.knative.dev

# Wait for recreation
sleep 30
kubectl get validatingwebhookconfiguration | grep knative
kubectl get mutatingwebhookconfiguration | grep knative
```

#### Issue: Service Not Routing to Pod
**Cause:** Service endpoints not updated  
**Fix:**
```bash
# Check service endpoints
kubectl get endpoints -n knative-serving webhook

# Should show pod IPs, if empty:
kubectl describe svc -n knative-serving webhook

# Check if pod has correct labels
kubectl get pods -n knative-serving -l app=webhook --show-labels

# Restart webhook
kubectl rollout restart deployment -n knative-serving webhook
```

#### Issue: Pod OOMKilled
**Cause:** Insufficient memory allocation  
**Fix:**
```bash
# Check current resource limits
kubectl get deployment -n knative-serving webhook -o yaml | grep -A 10 resources

# Increase memory limits
kubectl patch deployment -n knative-serving webhook -p '{"spec":{"template":{"spec":{"containers":[{"name":"webhook","resources":{"limits":{"memory":"512Mi"},"requests":{"memory":"256Mi"}}}]}}}}'

# Wait for rollout
kubectl rollout status deployment -n knative-serving webhook
```

#### Issue: API Server Cannot Reach Webhook
**Cause:** Network or firewall blocking API server → webhook  
**Fix:**
```bash
# Check webhook service
kubectl get svc -n knative-serving webhook

# Verify service is accessible
kubectl run test-webhook --rm -it --image=curlimages/curl --restart=Never -- \
  curl -k https://webhook.knative-serving.svc.cluster.local:443

# Check network policies
kubectl get networkpolicy -n knative-serving

# Check if API server can reach webhook
kubectl logs -n kube-system -l component=kube-apiserver | grep -i webhook
```

### Step 3: Restart Webhook

If no specific issue identified:

```bash
# Restart webhook deployment
kubectl rollout restart deployment -n knative-serving webhook

# Watch rollout progress
kubectl rollout status deployment -n knative-serving webhook

# Verify new pod is running
kubectl get pods -n knative-serving -l app=webhook
```

### Step 4: Recreate Webhook Configurations

```bash
# Delete webhook configurations (will be auto-recreated)
kubectl delete validatingwebhookconfiguration config.webhook.serving.knative.dev
kubectl delete mutatingwebhookconfiguration webhook.serving.knative.dev

# Wait for webhook to recreate them
sleep 30

# Verify recreation
kubectl get validatingwebhookconfiguration | grep knative
kubectl get mutatingwebhookconfiguration | grep knative
```

### Step 5: Force Reconciliation via Operator

```bash
# Trigger operator to reconcile
kubectl annotate knativeserving knative-serving -n knative-serving reconcile=$(date +%s) --overwrite

# Watch operator logs
kubectl logs -n knative-operator -l app=knative-operator --tail=100
```

## Verification

### 1. Check Webhook is Running

```bash
kubectl get pods -n knative-serving -l app=webhook
# Should show Running status with 1/1 READY
```

### 2. Check Webhook Logs

```bash
kubectl logs -n knative-serving -l app=webhook --tail=50
# Should show no errors
```

### 3. Check Webhook Service

```bash
kubectl get svc -n knative-serving webhook
kubectl get endpoints -n knative-serving webhook
# Should show pod IPs in endpoints
```

### 4. Check Webhook Configurations

```bash
# Verify configurations exist
kubectl get validatingwebhookconfiguration | grep knative
kubectl get mutatingwebhookconfiguration | grep knative

# Check configuration details
kubectl get validatingwebhookconfiguration config.webhook.serving.knative.dev -o yaml | grep -A 5 webhooks
```

### 5. Test Service Creation (Webhook Validation)

```bash
# Create test service (will be validated by webhook)
cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: webhook-test
  namespace: default
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
        env:
        - name: TARGET
          value: "Webhook Test"
EOF

# Should succeed without errors
# Wait for service to be ready
kubectl wait --for=condition=Ready ksvc/webhook-test -n default --timeout=60s

# Check service was mutated (default values added)
kubectl get ksvc webhook-test -n default -o yaml | grep -A 20 spec

# Test invalid service (webhook should reject)
cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: invalid-test
  namespace: default
spec:
  template:
    spec:
      containers:
      - image: "invalid-image-name"
        resources:
          requests:
            cpu: "invalid-cpu"
EOF
# Should fail with validation error from webhook

# Cleanup
kubectl delete ksvc webhook-test -n default
```

### 6. Test Webhook Mutation

```bash
# Create service without all fields (webhook will add defaults)
kubectl apply -f - <<EOF
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: mutation-test
  namespace: default
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
EOF

# Check mutated fields were added
kubectl get ksvc mutation-test -n default -o yaml | grep -A 30 spec

# Should see default values added by webhook:
# - containerConcurrency
# - timeoutSeconds
# - etc.

# Cleanup
kubectl delete ksvc mutation-test -n default
```

## Prevention

### 1. Resource Management

Ensure adequate resources:

```yaml
# In KnativeServing CR or deployment
resources:
  requests:
    cpu: 20m
    memory: 64Mi
  limits:
    cpu: 200m
    memory: 512Mi
```

### 2. Certificate Management

Monitor certificate expiration:

```bash
# Check certificate validity
kubectl get secret -n knative-serving webhook-certs -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates
```

### 3. Monitoring Setup

Key metrics to monitor:
- Webhook pod availability
- Admission request latency
- Admission request failures
- Certificate expiration
- Service endpoint availability

### 4. High Availability

For production, run multiple replicas:

```bash
# Scale webhook for HA
kubectl scale deployment -n knative-serving webhook --replicas=2
```

### 5. Pod Disruption Budget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: webhook-pdb
  namespace: knative-serving
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: webhook
```

## Performance Tips

1. **Multiple Replicas**: Run 2+ replicas in production for HA
2. **Resource Allocation**: Provide adequate resources for high request rate
3. **Certificate Management**: Automate certificate rotation
4. **Timeout Configuration**: Set appropriate webhook timeout (default 10s)
5. **Failure Policy**: Understand webhook failure policy (Fail vs Ignore)

## Related Alerts

- `KnativeServingDown`
- `KnativeControllerDown`
- `APIServerDown`

## Escalation

If webhook cannot be restored within 10 minutes:

1. ✅ Verify all resolution steps completed
2. 🔍 Check API server can reach webhook service
3. 📊 Review certificate validity
4. 🔄 Consider temporary webhook bypass (emergency only)
5. 📞 Escalate to platform team
6. 🆘 Page on-call engineer if deployments are blocked

## Additional Resources

- [Kubernetes Admission Webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
- [Knative Webhook Documentation](https://knative.dev/docs/serving/webhook/)
- [Knative Troubleshooting](https://knative.dev/docs/serving/troubleshooting/)

## Quick Commands Reference

```bash
# Check webhook status
kubectl get pods -n knative-serving -l app=webhook

# View webhook logs
kubectl logs -n knative-serving -l app=webhook --tail=100

# Restart webhook
kubectl rollout restart deployment -n knative-serving webhook

# Check webhook service
kubectl get svc -n knative-serving webhook
kubectl get endpoints -n knative-serving webhook

# Check webhook configurations
kubectl get validatingwebhookconfiguration | grep knative
kubectl get mutatingwebhookconfiguration | grep knative

# Check webhook certificates
kubectl get secret -n knative-serving webhook-certs

# Delete and recreate webhook configs
kubectl delete validatingwebhookconfiguration config.webhook.serving.knative.dev
kubectl delete mutatingwebhookconfiguration webhook.serving.knative.dev

# Check resource usage
kubectl top pods -n knative-serving -l app=webhook
```

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Platform Team

