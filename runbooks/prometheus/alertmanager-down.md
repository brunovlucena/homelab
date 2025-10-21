# 🚨 Runbook: AlertManager Down

## Alert Information

**Alert Name:** `AlertManagerDown`  
**Severity:** Critical  
**Component:** alertmanager  
**Service:** alertmanager

## Symptom

AlertManager has been down for more than 1 minute. No alerts are being sent to notification channels.

## Impact

- **User Impact:** HIGH - No alert notifications
- **Business Impact:** CRITICAL - Cannot respond to incidents
- **Data Impact:** LOW - Alerts are queued in Prometheus

## Diagnosis

### 1. Check Pod Status

```bash
kubectl get pods -n prometheus -l app.kubernetes.io/name=alertmanager
kubectl describe pod -n prometheus -l app.kubernetes.io/name=alertmanager
```

### 2. Check Pod Logs

```bash
kubectl logs -n prometheus -l app.kubernetes.io/name=alertmanager --tail=100
kubectl logs -n prometheus -l app.kubernetes.io/name=alertmanager --previous  # If pod restarted
```

### 3. Check Service and Endpoints

```bash
kubectl get svc -n prometheus prometheus-kube-prometheus-alertmanager
kubectl get endpoints -n prometheus prometheus-kube-prometheus-alertmanager
```

### 4. Check Events

```bash
kubectl get events -n prometheus --sort-by='.lastTimestamp' | grep -i alertmanager | head -20
```

### 5. Check Resource Limits

```bash
kubectl top pods -n prometheus -l app.kubernetes.io/name=alertmanager
```

### 6. Check StatefulSet Status

```bash
kubectl get statefulset -n prometheus
kubectl describe statefulset -n prometheus alertmanager-prometheus-kube-prometheus-alertmanager
```

## Resolution Steps

### Step 1: Check if pods are running

```bash
POD_STATUS=$(kubectl get pods -n prometheus -l app.kubernetes.io/name=alertmanager -o jsonpath='{.items[0].status.phase}')
echo "Pod Status: $POD_STATUS"

# Check all replicas
kubectl get pods -n prometheus -l app.kubernetes.io/name=alertmanager
```

### Step 2: If pods are not running, check why

```bash
# Check for ImagePullBackOff
kubectl describe pod -n prometheus -l app.kubernetes.io/name=alertmanager | grep -A 10 "Events:"

# Check for CrashLoopBackOff
kubectl logs -n prometheus -l app.kubernetes.io/name=alertmanager --previous
```

### Step 3: Common Issues and Fixes

#### Issue: ImagePullBackOff
**Cause:** Cannot pull container image  
**Fix:**
```bash
# Verify image exists
kubectl get statefulset -n prometheus alertmanager-prometheus-kube-prometheus-alertmanager \
  -o jsonpath='{.spec.template.spec.containers[0].image}'

# Check image pull secrets
kubectl get secrets -n prometheus
```

#### Issue: CrashLoopBackOff
**Cause:** Application failing to start  
**Fix:**
```bash
# Check application logs for startup errors
kubectl logs -n prometheus -l app.kubernetes.io/name=alertmanager --tail=100

# Common causes:
# - Configuration errors
# - Invalid receiver config
# - Missing secrets (e.g., webhook URLs, API keys)
# - Template errors
```

#### Issue: Configuration Error
**Cause:** Invalid AlertManager configuration  
**Fix:**
```bash
# Check AlertManager config
kubectl get secret -n prometheus alertmanager-prometheus-kube-prometheus-alertmanager -o yaml

# Decode and check config
kubectl get secret -n prometheus alertmanager-prometheus-kube-prometheus-alertmanager \
  -o jsonpath='{.data.alertmanager\.yaml}' | base64 -d

# Validate config syntax
kubectl exec -n prometheus alertmanager-prometheus-kube-prometheus-alertmanager-0 -- \
  amtool check-config /etc/alertmanager/config/alertmanager.yaml

# If config is invalid, fix via Helm values
# alertmanager:
#   config:
#     route:
#       receiver: 'default'
#     receivers:
#       - name: 'default'
```

#### Issue: Missing Secrets
**Cause:** Secrets referenced in config don't exist  
**Fix:**
```bash
# Check for secret references in config
kubectl get secret -n prometheus alertmanager-prometheus-kube-prometheus-alertmanager \
  -o jsonpath='{.data.alertmanager\.yaml}' | base64 -d | grep -i secret

# Create missing secrets
kubectl create secret generic <secret-name> -n prometheus \
  --from-literal=<key>=<value>

# Or use scripts/create-secrets.sh
```

#### Issue: OOMKilled
**Cause:** Out of memory  
**Fix:**
```bash
# Check memory usage patterns
kubectl top pods -n prometheus -l app.kubernetes.io/name=alertmanager

# Increase memory limits via Helm values
# alertmanager:
#   alertmanagerSpec:
#     resources:
#       limits:
#         memory: 512Mi  # Increase from current
#       requests:
#         memory: 256Mi
```

#### Issue: Storage Full
**Cause:** PVC out of space  
**Fix:**
```bash
# Check PVC usage
kubectl exec -n prometheus alertmanager-prometheus-kube-prometheus-alertmanager-0 -- \
  df -h /alertmanager

# If full, expand PVC
kubectl edit pvc -n prometheus alertmanager-prometheus-kube-prometheus-alertmanager-db-alertmanager-prometheus-kube-prometheus-alertmanager-0

# Or reduce retention
# alertmanager:
#   alertmanagerSpec:
#     retention: 120h  # Reduce from current
```

#### Issue: Cluster Communication Failure
**Cause:** AlertManager instances can't communicate (in HA setup)  
**Fix:**
```bash
# Check cluster status
kubectl exec -n prometheus alertmanager-prometheus-kube-prometheus-alertmanager-0 -- \
  amtool cluster show

# Check network policies
kubectl get networkpolicies -n prometheus

# Check service mesh (if using Linkerd/Istio)
kubectl get pods -n prometheus -l app.kubernetes.io/name=alertmanager \
  -o jsonpath='{.items[*].metadata.annotations}'
```

### Step 4: Restart AlertManager if needed

```bash
kubectl rollout restart statefulset/alertmanager-prometheus-kube-prometheus-alertmanager -n prometheus
kubectl rollout status statefulset/alertmanager-prometheus-kube-prometheus-alertmanager -n prometheus
```

### Step 5: Force reconcile Flux/Helm if needed

```bash
flux reconcile helmrelease kube-prometheus-stack -n prometheus
```

## Verification

1. Check that pods are running:
```bash
kubectl get pods -n prometheus -l app.kubernetes.io/name=alertmanager
```

2. Verify AlertManager UI is accessible:
```bash
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-alertmanager 9093:9093
# Open http://localhost:9093
```

3. Check AlertManager status:
```bash
kubectl exec -n prometheus alertmanager-prometheus-kube-prometheus-alertmanager-0 -- \
  amtool check-config /etc/alertmanager/config/alertmanager.yaml

# Check cluster status (if HA)
kubectl exec -n prometheus alertmanager-prometheus-kube-prometheus-alertmanager-0 -- \
  amtool cluster show
```

4. Verify Prometheus can reach AlertManager:
```bash
# In Prometheus UI -> Status -> Runtime & Build Information
# Check "Alertmanagers" section
```

5. Test alert routing:
```bash
# Send a test alert
kubectl exec -n prometheus alertmanager-prometheus-kube-prometheus-alertmanager-0 -- \
  amtool alert add test severity=critical instance=test alertname=TestAlert

# Check in AlertManager UI if alert appears
# Check if notification was sent to configured receivers
```

6. Verify alert notifications:
```bash
# Check logs for notification attempts
kubectl logs -n prometheus -l app.kubernetes.io/name=alertmanager | grep -i "notify\|sent"
```

## Prevention

1. Set appropriate resource requests and limits
2. Validate AlertManager config before applying
3. Monitor AlertManager health
4. Implement HA setup (multiple replicas)
5. Regular testing of notification channels
6. Monitor storage usage
7. Use secrets management for sensitive data
8. Set up alerts for AlertManager down
9. Document and test all receiver configurations

## Related Alerts

- `PrometheusDown`
- `AlertManagerClusterDown`
- `AlertManagerClusterFailedToSendAlerts`
- `AlertManagerConfigInconsistent`
- `AlertManagerFailedToSendAlerts`
- `AlertManagerMembersInconsistent`

## Escalation

If the issue persists after following these steps:
1. Check notification receiver endpoints (Slack, email, etc.)
2. Review recent configuration changes
3. Check for network issues blocking outbound connections
4. Verify secrets and credentials are valid
5. Contact on-call engineer

## Additional Resources

- [AlertManager Documentation](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [AlertManager Configuration](https://prometheus.io/docs/alerting/latest/configuration/)
- [amtool CLI](https://github.com/prometheus/alertmanager#examples)
- [Notification Template Reference](https://prometheus.io/docs/alerting/latest/notification_examples/)
- [AlertManager Clustering](https://prometheus.io/docs/alerting/latest/alertmanager/#high-availability)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

