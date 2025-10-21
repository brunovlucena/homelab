# 🚨 Runbook: Prometheus Rule Evaluation Failures

## Alert Information

**Alert Name:** `PrometheusRuleEvaluationFailures`  
**Severity:** Warning  
**Component:** prometheus  
**Service:** rule-evaluation

## Symptom

Prometheus is experiencing failures when evaluating recording or alerting rules. This may result in missing alerts or metrics.

## Impact

- **User Impact:** MEDIUM - Some alerts may not fire
- **Business Impact:** HIGH - Potential to miss critical alerts
- **Data Impact:** MEDIUM - Recording rules not generating metrics

## Diagnosis

### 1. Check Rule Evaluation Errors

```promql
# Rate of rule evaluation failures
rate(prometheus_rule_evaluation_failures_total[5m])

# Total failures by rule group
sum by (rule_group) (prometheus_rule_evaluation_failures_total)

# Failures by type (recording vs alerting)
sum by (rule_group, rule_type) (prometheus_rule_evaluation_failures_total)
```

### 2. Check Prometheus Logs

```bash
# Look for rule evaluation errors
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus --tail=200 | grep -i "rule\|evaluation"

# Look for specific error messages
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus --tail=500 | grep -i "error\|failed"
```

### 3. Check Rule Groups Status

```bash
# Access Prometheus UI
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-prometheus 9090:9090

# Open http://localhost:9090/rules
# Look for rules with "error" status
```

### 4. List All PrometheusRules

```bash
# List all PrometheusRule CRDs
kubectl get prometheusrule -A

# Get details of specific rule
kubectl describe prometheusrule -n <namespace> <rule-name>
```

### 5. Check Rule Evaluation Duration

```promql
# Rule evaluation duration
prometheus_rule_group_last_duration_seconds

# Slow rule groups (taking > 30s)
prometheus_rule_group_last_duration_seconds > 30

# Rule evaluation iterations
rate(prometheus_rule_group_iterations_total[5m])
```

## Resolution Steps

### Step 1: Identify failing rules

```bash
# Check Prometheus logs for failed rules
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus --tail=500 \
  | grep -i "evaluation failed" -A 5 -B 5

# Or use Prometheus UI -> Status -> Rules
# Look for rules marked with errors
```

### Step 2: Common Issues and Fixes

#### Issue: Invalid PromQL Syntax
**Cause:** Rule query has syntax errors  
**Fix:**
```bash
# Identify the problematic rule from logs
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus | grep "invalid query"

# Test the query in Prometheus UI
# Fix the query in the PrometheusRule CRD
kubectl edit prometheusrule -n <namespace> <rule-name>

# Validate PromQL syntax before applying
# Use Prometheus UI -> Graph to test queries
```

#### Issue: Missing Metrics
**Cause:** Rule queries reference metrics that don't exist  
**Fix:**
```bash
# Check if the metric exists
# In Prometheus UI, search for the metric name

# If metric doesn't exist:
# 1. Check if the exporter is running
# 2. Check if ServiceMonitor is correctly configured
# 3. Check if metric was renamed

# Update rule to use correct metric name
kubectl edit prometheusrule -n <namespace> <rule-name>
```

#### Issue: Query Timeout
**Cause:** Rule query is too complex or slow  
**Fix:**
```bash
# Identify slow rules
# Query: prometheus_rule_group_last_duration_seconds > 30

# Optimize the query:
# 1. Use recording rules for complex calculations
# 2. Reduce time range (use shorter lookback)
# 3. Use more specific label matchers
# 4. Avoid high-cardinality operations

# Increase query timeout (if necessary)
kubectl edit prometheus -n prometheus prometheus-kube-prometheus-prometheus
# Add: ruleQueryOffset: 30s (default is 5s)
```

#### Issue: Division by Zero
**Cause:** Rule performs division without checking for zero  
**Fix:**
```bash
# Add zero check to rule
# Bad:  metric1 / metric2
# Good: metric1 / (metric2 > 0)
# Or:   metric1 / (metric2 or vector(1))

kubectl edit prometheusrule -n <namespace> <rule-name>
```

#### Issue: Label Mismatch
**Cause:** Rule tries to operate on metrics with different label sets  
**Fix:**
```bash
# Use 'ignoring' or 'on' to handle label mismatches
# Example: metric1 / ignoring(pod) metric2
# Or:      metric1 / on(job, instance) metric2

kubectl edit prometheusrule -n <namespace> <rule-name>
```

#### Issue: Circular Dependency
**Cause:** Recording rule references another rule in the same group  
**Fix:**
```bash
# Move dependent rules to different groups with proper ordering
# Groups are evaluated in alphabetical order

# Check rule group intervals
kubectl get prometheusrule -A -o yaml | grep -A 2 "interval:"

# Reorganize rules to avoid circular dependencies
kubectl edit prometheusrule -n <namespace> <rule-name>
```

#### Issue: High Cardinality Result
**Cause:** Rule produces too many time series  
**Fix:**
```bash
# Add aggregation to reduce cardinality
# Example: sum by (job, instance) (metric) instead of just metric

# Or add label filtering
# Example: metric{job="important-job"}

kubectl edit prometheusrule -n <namespace> <rule-name>
```

#### Issue: PrometheusRule Not Loaded
**Cause:** PrometheusRule CRD not picked up by Prometheus Operator  
**Fix:**
```bash
# Check if PrometheusRule has correct labels
kubectl get prometheusrule -n <namespace> <rule-name> -o yaml | grep -A 5 "labels:"

# Prometheus selects rules based on ruleSelector
kubectl get prometheus -n prometheus prometheus-kube-prometheus-prometheus \
  -o jsonpath='{.spec.ruleSelector}'

# Add required labels to PrometheusRule
kubectl edit prometheusrule -n <namespace> <rule-name>
# Ensure labels match ruleSelector (usually: release: kube-prometheus-stack)
```

### Step 3: Validate rule syntax

```bash
# Get the PrometheusRule
kubectl get prometheusrule -n <namespace> <rule-name> -o yaml > /tmp/rule.yaml

# Use promtool to validate (if available locally)
promtool check rules /tmp/rule.yaml

# Or test each query in Prometheus UI
```

### Step 4: Apply fixes and monitor

```bash
# After editing, monitor for errors
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus --follow | grep -i "error\|rule"

# Check rule status in UI
# Prometheus UI -> Status -> Rules
```

### Step 5: Force rule reload (if needed)

```bash
# Prometheus should auto-reload, but can force restart
kubectl delete pod -n prometheus -l app.kubernetes.io/name=prometheus

# Or use reload endpoint
kubectl exec -n prometheus prometheus-prometheus-kube-prometheus-prometheus-0 -- \
  kill -HUP 1
```

## Verification

1. Check rule evaluation is succeeding:
```promql
# Should be 0 or decreasing
rate(prometheus_rule_evaluation_failures_total[5m])
```

2. Verify rules are loaded:
```bash
# In Prometheus UI -> Status -> Rules
# All rules should show "OK" status
```

3. Check recording rule metrics:
```promql
# If it was a recording rule, verify the new metric exists
<recording_rule_metric_name>
```

4. Check alert state:
```bash
# If it was an alerting rule, check alert status
# Prometheus UI -> Alerts
```

5. Monitor evaluation duration:
```promql
prometheus_rule_group_last_duration_seconds{rule_group="<group-name>"}
```

## Prevention

1. Test PromQL queries before adding to rules
2. Use promtool to validate rule files
3. Implement CI/CD validation for PrometheusRules
4. Monitor rule evaluation duration
5. Set up alerts for rule evaluation failures
6. Document complex rules with comments
7. Use recording rules for complex calculations
8. Regular audits of rule performance
9. Keep rules simple and maintainable
10. Test rules in staging before production

## Related Alerts

- `PrometheusDown`
- `PrometheusRuleEvaluationSlow`
- `PrometheusRuleDuplicateName`
- `PrometheusAlertmanagerDown`
- `PrometheusConfigReloadFailed`

## Escalation

If the issue persists after following these steps:
1. Review all recent PrometheusRule changes
2. Check for Prometheus version-specific bugs
3. Review Prometheus resource usage (CPU/Memory)
4. Check for underlying infrastructure issues
5. Contact Prometheus expert or on-call engineer

## Additional Resources

- [Prometheus Recording Rules](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/)
- [Prometheus Alerting Rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/)
- [PromQL Operators](https://prometheus.io/docs/prometheus/latest/querying/operators/)
- [PrometheusRule CRD](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#prometheusrule)
- [Rule Best Practices](https://prometheus.io/docs/practices/rules/)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

