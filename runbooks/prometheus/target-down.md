# 🚨 Runbook: Prometheus Target Down

## Alert Information

**Alert Name:** `PrometheusTargetDown`  
**Severity:** Warning  
**Component:** prometheus  
**Service:** scrape-targets

## Symptom

One or more Prometheus scrape targets have been down for more than 5 minutes. Metrics from these targets are not being collected.

## Impact

- **User Impact:** MEDIUM - Specific service metrics unavailable
- **Business Impact:** MEDIUM - Partial observability blind spot
- **Data Impact:** MEDIUM - Metrics gap for affected targets

## Diagnosis

### 1. Identify Down Targets

```bash
# Via Prometheus UI
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-prometheus 9090:9090
# Open http://localhost:9090/targets
# Look for targets with "DOWN" status
```

### 2. Query Down Targets via PromQL

```promql
# Show all down targets
up == 0

# Count down targets by job
count by (job) (up == 0)

# Show targets that were recently up but now down
(up == 0) and (up offset 10m == 1)
```

### 3. Check Target Pod Status

```bash
# Replace <namespace> and <pod-label> with actual values
kubectl get pods -n <namespace> -l <pod-label>
kubectl describe pod -n <namespace> -l <pod-label>
```

### 4. Check Service and ServiceMonitor

```bash
# List all ServiceMonitors
kubectl get servicemonitor -A

# Describe specific ServiceMonitor
kubectl describe servicemonitor -n <namespace> <servicemonitor-name>

# Check if service exists and has endpoints
kubectl get svc -n <namespace>
kubectl get endpoints -n <namespace> <service-name>
```

### 5. Check Network Policies

```bash
# Check if network policies are blocking scraping
kubectl get networkpolicies -n <namespace>
kubectl describe networkpolicy -n <namespace>
```

## Resolution Steps

### Step 1: Identify the affected target

```bash
# Get target details from Prometheus
# Open Prometheus UI -> Status -> Targets
# Note the job name, namespace, and pod name
```

### Step 2: Check if target pods are running

```bash
TARGET_JOB="<job-name>"
TARGET_NAMESPACE="<namespace>"

kubectl get pods -n $TARGET_NAMESPACE
```

### Step 3: Common Issues and Fixes

#### Issue: Target Pod Not Running
**Cause:** Pod crashed, evicted, or not scheduled  
**Fix:**
```bash
# Check pod status
kubectl get pods -n $TARGET_NAMESPACE

# If pod is missing, check deployment/statefulset
kubectl get deployment,statefulset -n $TARGET_NAMESPACE

# Check for recent events
kubectl get events -n $TARGET_NAMESPACE --sort-by='.lastTimestamp' | head -20

# Restart if needed
kubectl rollout restart deployment/<deployment-name> -n $TARGET_NAMESPACE
```

#### Issue: Metrics Endpoint Not Responding
**Cause:** Application not exposing metrics, port mismatch, or application crashed  
**Fix:**
```bash
# Find the pod
POD_NAME=$(kubectl get pods -n $TARGET_NAMESPACE -l app=<app-label> -o jsonpath='{.items[0].metadata.name}')

# Check if metrics endpoint is accessible from within the pod
kubectl exec -n $TARGET_NAMESPACE $POD_NAME -- wget -O- http://localhost:<metrics-port>/metrics

# Port forward and test locally
kubectl port-forward -n $TARGET_NAMESPACE pod/$POD_NAME <metrics-port>:<metrics-port>
curl http://localhost:<metrics-port>/metrics
```

#### Issue: ServiceMonitor Misconfigured
**Cause:** Wrong selector, port, or path in ServiceMonitor  
**Fix:**
```bash
# Check ServiceMonitor configuration
kubectl get servicemonitor -n $TARGET_NAMESPACE <servicemonitor-name> -o yaml

# Verify selector matches service labels
kubectl get svc -n $TARGET_NAMESPACE <service-name> --show-labels

# Check port name matches
kubectl get svc -n $TARGET_NAMESPACE <service-name> -o yaml | grep -A 5 "ports:"

# Edit ServiceMonitor if needed
kubectl edit servicemonitor -n $TARGET_NAMESPACE <servicemonitor-name>
```

#### Issue: Service Missing or No Endpoints
**Cause:** Service deleted, no pods matching selector  
**Fix:**
```bash
# Check if service exists
kubectl get svc -n $TARGET_NAMESPACE

# Check service endpoints
kubectl get endpoints -n $TARGET_NAMESPACE <service-name>

# If no endpoints, check pod labels
kubectl get pods -n $TARGET_NAMESPACE --show-labels

# Verify service selector matches pod labels
kubectl get svc -n $TARGET_NAMESPACE <service-name> -o yaml | grep -A 3 "selector:"
```

#### Issue: Network Policy Blocking
**Cause:** Network policy preventing Prometheus from scraping  
**Fix:**
```bash
# Check network policies
kubectl get networkpolicies -n $TARGET_NAMESPACE

# Verify Prometheus can reach the target
kubectl get pods -n prometheus -l app.kubernetes.io/name=prometheus -o name | \
  xargs -I {} kubectl exec -n prometheus {} -- wget -O- --timeout=5 \
  http://<service-name>.<namespace>.svc.cluster.local:<port>/metrics

# If blocked, update network policy to allow Prometheus namespace
```

#### Issue: Authentication/TLS Issues
**Cause:** Metrics endpoint requires auth or TLS  
**Fix:**
```bash
# Check ServiceMonitor for TLS/auth config
kubectl get servicemonitor -n $TARGET_NAMESPACE <servicemonitor-name> -o yaml

# If TLS is required, ensure certificates are configured
# If bearer token is required, ensure secret exists
kubectl get secrets -n $TARGET_NAMESPACE
```

#### Issue: High Cardinality/Timeout
**Cause:** Too many metrics causing scrape timeout  
**Fix:**
```bash
# Check scrape duration
# In Prometheus UI, query: scrape_duration_seconds{job="<job-name>"}

# Check sample count
# Query: scrape_samples_scraped{job="<job-name>"}

# If high, increase scrape timeout in ServiceMonitor
kubectl edit servicemonitor -n $TARGET_NAMESPACE <servicemonitor-name>
# Add: scrapeTimeout: 30s (default is 10s)

# Or reduce metric cardinality in the application
```

### Step 4: Force Prometheus config reload

```bash
# Prometheus should auto-reload, but you can force it
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  kill -HUP 1

# Or restart Prometheus
kubectl delete pod -n prometheus -l app.kubernetes.io/name=prometheus
```

### Step 5: Verify ServiceMonitor is picked up

```bash
# Check Prometheus configuration
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  cat /etc/prometheus/config_out/prometheus.env.yaml | grep -A 20 "<job-name>"
```

## Verification

1. Check target status in Prometheus UI:
```bash
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-prometheus 9090:9090
# Open http://localhost:9090/targets
# Verify target is UP
```

2. Verify metrics are being scraped:
```promql
# Check up metric
up{job="<job-name>"}

# Check scrape samples
scrape_samples_scraped{job="<job-name>"}

# Verify actual metrics from target
<any_metric_from_target>
```

3. Check scrape health:
```promql
# Scrape duration
scrape_duration_seconds{job="<job-name>"}

# Scrape samples
scrape_samples_scraped{job="<job-name>"}

# Scrape series added
scrape_series_added{job="<job-name>"}
```

## Prevention

1. Set up alerts for target down conditions
2. Implement health checks for metrics endpoints
3. Monitor scrape duration and sample counts
4. Use pod disruption budgets for critical targets
5. Document required network policies for scraping
6. Regular testing of ServiceMonitor configurations
7. Monitor application health that exposes metrics
8. Set appropriate scrape intervals and timeouts

## Related Alerts

- `PrometheusDown`
- `PrometheusScrapeFailed`
- `PrometheusScrapeTimeout`
- `PrometheusHighCardinality`
- `PrometheusTargetMissing`

## Escalation

If the issue persists after following these steps:
1. Check if this is a known issue with the target application
2. Review recent changes to ServiceMonitor or target deployment
3. Check for cluster-wide networking issues
4. Verify Prometheus operator is working correctly
5. Contact application team responsible for the target

## Additional Resources

- [Prometheus Scraping Documentation](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config)
- [ServiceMonitor CRD](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#servicemonitor)
- [Prometheus Operator Troubleshooting](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/troubleshooting.md)
- [Network Policies Guide](https://kubernetes.io/docs/concepts/services-networking/network-policies/)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

