# LanceDB High Memory Usage

## 🚨 Alert

**Alert Name:** LanceDBHighMemoryUsage  
**Severity:** Warning  
**Component:** LanceDB

## 📊 Description

LanceDB memory usage is above 85% of the configured limit. This could lead to OOM kills and service disruption.

## 🔍 Investigation Steps

### 1. Check Current Memory Usage

\`\`\`bash
kubectl top pods -n lancedb
kubectl describe pod -n lancedb -l app=lancedb | grep -A5 "Limits\\|Requests"
\`\`\`

### 2. Check Memory Metrics in Grafana

Navigate to the LanceDB dashboard in Grafana and review:
- Memory usage trends
- Memory growth rate
- Peak memory usage times

### 3. Check Database Size

\`\`\`bash
# Check PVC usage
kubectl exec -it -n lancedb deployment/lancedb -- df -h /data

# List database size
kubectl exec -it -n lancedb deployment/lancedb -- du -sh /data/lancedb/*
\`\`\`

### 4. Review Application Logs

\`\`\`bash
kubectl logs -n lancedb deployment/lancedb --tail=200 | grep -i "memory\\|oom\\|killed"
\`\`\`

## 🔧 Solutions

### Solution 1: Increase Memory Limits

If memory usage is legitimate and growing:

\`\`\`bash
kubectl edit deployment lancedb -n lancedb

# Update memory limits:
# resources:
#   limits:
#     memory: 4Gi  # Increase from 2Gi
#   requests:
#     memory: 1Gi  # Increase from 512Mi
\`\`\`

### Solution 2: Optimize Database

If there's data bloat or unnecessary data:

\`\`\`bash
# Connect to LanceDB and run cleanup/optimization commands
# This depends on LanceDB's API and your use case
\`\`\`

### Solution 3: Implement Data Retention

Set up data retention policies to automatically clean old data:

\`\`\`python
# Example: Delete old tables or compact data
import lancedb

db = lancedb.connect("http://lancedb.lancedb.svc.cluster.local:8000")
# Implement retention logic based on your use case
\`\`\`

### Solution 4: Scale Horizontally (Future)

Consider implementing horizontal scaling with multiple LanceDB instances and a load balancer.

## 📊 Monitoring

Set up alerts for:
- Memory usage trends
- Database size growth
- Query performance degradation

## ✅ Resolution Verification

1. Check memory usage has decreased:
   \`\`\`bash
   kubectl top pods -n lancedb
   \`\`\`

2. Monitor for 24 hours to ensure stability

3. Update alert thresholds if necessary

## 📝 Prevention

- Implement data retention policies
- Monitor database growth trends
- Right-size memory limits based on actual usage
- Consider data archival strategies

## 🔗 Related Runbooks

- [LanceDB Down](./lancedb-down.md)
- [High CPU Usage](./high-cpu-usage.md)
- [PVC Almost Full](./pvc-almost-full.md)

