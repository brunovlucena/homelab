# ⚠️ Runbook: Projects Load Failed

## Alert Information
**Alert Name:** `ProjectsLoadFailed`  
**Severity:** Warning/Critical  

## Symptom
Projects Load Failed detected in Bruno Site.

## Diagnosis
```bash
# Check logs for related errors
kubectl logs -n homepage -l app.kubernetes.io/component=api --tail=200 | grep -i "projects"

# Check pod status
kubectl get pods -n homepage

# Check recent events
kubectl get events -n homepage --sort-by='.lastTimestamp' | head -20
```

## Resolution Steps

### Step 1: Identify Root Cause
Review application logs and metrics to identify the specific issue.

### Step 2: Apply Fix
- Check configuration
- Verify dependencies
- Restart pods if needed
- Scale resources if required

### Step 3: Verify Resolution
Monitor metrics to confirm the issue is resolved.

## Prevention
1. Monitor metrics regularly
2. Implement proper error handling
3. Set up automated testing
4. Review code for issues
5. Implement rate limiting where appropriate

## Related Alerts
Check related alerts for cascading issues.

## Additional Resources
- [Homepage Documentation](../../../flux/clusters/homelab/infrastructure/homepage/README.md)
- [Kubernetes Troubleshooting](https://kubernetes.io/docs/tasks/debug/)
