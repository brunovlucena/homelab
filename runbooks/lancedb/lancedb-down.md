# LanceDB Down

## 🚨 Alert

**Alert Name:** LanceDBDown  
**Severity:** Critical  
**Component:** LanceDB

## 📊 Description

LanceDB service is down and unavailable. This affects all AI applications that depend on vector storage and retrieval.

## 🔍 Investigation Steps

### 1. Check Pod Status

\`\`\`bash
kubectl get pods -n lancedb
kubectl describe pod -n lancedb -l app=lancedb
\`\`\`

### 2. Check Pod Logs

\`\`\`bash
kubectl logs -n lancedb deployment/lancedb --tail=100
kubectl logs -n lancedb deployment/lancedb --previous  # If pod is restarting
\`\`\`

### 3. Check Service Status

\`\`\`bash
kubectl get svc -n lancedb
kubectl describe svc lancedb -n lancedb
\`\`\`

### 4. Check Events

\`\`\`bash
kubectl get events -n lancedb --sort-by='.lastTimestamp'
\`\`\`

### 5. Check Resource Availability

\`\`\`bash
kubectl top pods -n lancedb
kubectl describe node  # Check node resources
\`\`\`

## 🔧 Common Causes and Solutions

### Cause 1: Image Pull Error

**Symptoms:** Pod status shows `ImagePullBackOff` or `ErrImagePull`

**Solution:**
\`\`\`bash
# Check image availability
kubectl describe pod -n lancedb -l app=lancedb | grep -A5 "Events"

# If the lancedb/lancedb:latest image doesn't exist, you may need to:
# 1. Build a custom LanceDB server image
# 2. Use LanceDB Cloud
# 3. Deploy LanceDB as a library in your application
\`\`\`

### Cause 2: PVC Mount Failure

**Symptoms:** Pod cannot start due to volume mount issues

**Solution:**
\`\`\`bash
# Check PVC status
kubectl get pvc -n lancedb
kubectl describe pvc lancedb-pvc -n lancedb

# Check PV
kubectl get pv

# If PVC is pending, check storage class
kubectl get storageclass
\`\`\`

### Cause 3: Resource Constraints

**Symptoms:** Pod is in `Pending` state

**Solution:**
\`\`\`bash
# Check node resources
kubectl describe nodes | grep -A5 "Allocated resources"

# Reduce resource requests if necessary
kubectl edit deployment lancedb -n lancedb
\`\`\`

### Cause 4: Configuration Error

**Symptoms:** Pod is crashing immediately after start

**Solution:**
\`\`\`bash
# Check deployment configuration
kubectl get deployment lancedb -n lancedb -o yaml

# Check environment variables
kubectl exec -it -n lancedb deployment/lancedb -- env

# Verify the LANCE_DB_URI is correctly set
\`\`\`

## ✅ Resolution

Once the issue is resolved:

1. Verify pod is running:
   \`\`\`bash
   kubectl get pods -n lancedb
   \`\`\`

2. Test health endpoint:
   \`\`\`bash
   kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \\
     curl http://lancedb.lancedb.svc.cluster.local:8000/health
   \`\`\`

3. Check Prometheus metrics:
   \`\`\`bash
   kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \\
     curl http://lancedb.lancedb.svc.cluster.local:8000/metrics
   \`\`\`

## 📝 Post-Incident

- Document root cause
- Update configuration if necessary
- Consider implementing backup/replica if this is a critical service
- Review resource limits and requests

## 🔗 Related Runbooks

- [Pod Crash Loop](./pod-crash-loop.md)
- [High Memory Usage](./high-memory-usage.md)
- [PVC Almost Full](./pvc-almost-full.md)

