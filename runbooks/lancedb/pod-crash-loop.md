# LanceDB Pod Crash Loop

## 🚨 Alert

**Alert Name:** LanceDBPodCrashLooping  
**Severity:** Critical  
**Component:** LanceDB

## 📊 Description

LanceDB pod is repeatedly crashing and restarting. This indicates a critical issue preventing the service from running.

## 🔍 Investigation Steps

### 1. Check Pod Status

\`\`\`bash
kubectl get pods -n lancedb
kubectl describe pod -n lancedb -l app=lancedb
\`\`\`

### 2. Check Current and Previous Logs

\`\`\`bash
# Current logs
kubectl logs -n lancedb deployment/lancedb --tail=100

# Previous container logs (before crash)
kubectl logs -n lancedb deployment/lancedb --previous --tail=100
\`\`\`

### 3. Check Events

\`\`\`bash
kubectl get events -n lancedb --sort-by='.lastTimestamp' | head -20
\`\`\`

### 4. Check Resource Limits

\`\`\`bash
kubectl describe pod -n lancedb -l app=lancedb | grep -A10 "Limits\\|Requests"
\`\`\`

## 🔧 Common Causes and Solutions

### Cause 1: OOMKilled (Out of Memory)

**Symptoms:** Pod status shows `OOMKilled`

**Solution:**
\`\`\`bash
# Increase memory limits
kubectl edit deployment lancedb -n lancedb

# Update:
# resources:
#   limits:
#     memory: 4Gi
#   requests:
#     memory: 1Gi
\`\`\`

### Cause 2: Configuration Error

**Symptoms:** Logs show configuration errors or invalid parameters

**Solution:**
\`\`\`bash
# Check deployment configuration
kubectl get deployment lancedb -n lancedb -o yaml

# Verify environment variables
kubectl describe deployment lancedb -n lancedb | grep -A10 "Environment"

# Fix configuration
kubectl edit deployment lancedb -n lancedb
\`\`\`

### Cause 3: Data Corruption

**Symptoms:** Logs show database corruption or read/write errors

**Solution:**
\`\`\`bash
# Check PVC status
kubectl get pvc -n lancedb
kubectl describe pvc lancedb-pvc -n lancedb

# If data is corrupted, you may need to:
# 1. Backup data if possible
# 2. Delete and recreate the PVC (DATA LOSS!)
# 3. Restore from backup

# Scale down deployment
kubectl scale deployment lancedb -n lancedb --replicas=0

# Delete PVC (CAUTION: This deletes all data!)
# kubectl delete pvc lancedb-pvc -n lancedb

# Recreate PVC
# kubectl apply -f flux/clusters/homelab/infrastructure/lancedb/pvc.yaml

# Scale up deployment
kubectl scale deployment lancedb -n lancedb --replicas=1
\`\`\`

### Cause 4: Failed Health Checks

**Symptoms:** Pod is killed due to failed liveness/readiness probes

**Solution:**
\`\`\`bash
# Temporarily disable or adjust probes
kubectl edit deployment lancedb -n lancedb

# Adjust initialDelaySeconds or timeoutSeconds:
# livenessProbe:
#   initialDelaySeconds: 60  # Increase from 30
#   timeoutSeconds: 10       # Increase from 5
\`\`\`

### Cause 5: Image Issues

**Symptoms:** Image pull errors or incompatible image

**Solution:**
\`\`\`bash
# Check image details
kubectl describe pod -n lancedb -l app=lancedb | grep -A5 "Image"

# If lancedb/lancedb:latest doesn't exist or is incompatible:
# You may need to build your own image or use a different deployment strategy
\`\`\`

## ✅ Resolution

Once resolved:

1. Verify pod is running and stable:
   \`\`\`bash
   kubectl get pods -n lancedb -w
   \`\`\`

2. Check logs for any warnings:
   \`\`\`bash
   kubectl logs -n lancedb deployment/lancedb --tail=50
   \`\`\`

3. Test service:
   \`\`\`bash
   kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \\
     curl http://lancedb.lancedb.svc.cluster.local:8000/health
   \`\`\`

4. Monitor for 1 hour to ensure stability

## 📝 Post-Incident

- Document root cause
- Update deployment configuration if needed
- Consider implementing:
  - Better health checks
  - Resource monitoring
  - Backup strategies
  - Graceful degradation

## 🔗 Related Runbooks

- [LanceDB Down](./lancedb-down.md)
- [High Memory Usage](./high-memory-usage.md)
- [PVC Almost Full](./pvc-almost-full.md)

