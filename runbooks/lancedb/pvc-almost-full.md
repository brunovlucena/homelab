# LanceDB PVC Almost Full

## 🚨 Alert

**Alert Name:** LanceDBPVCAlmostFull  
**Severity:** Warning  
**Component:** LanceDB Storage

## 📊 Description

LanceDB persistent volume is more than 85% full. This could lead to write failures and service disruption.

## 🔍 Investigation Steps

### 1. Check PVC Usage

\`\`\`bash
kubectl get pvc -n lancedb
kubectl describe pvc lancedb-pvc -n lancedb

# Check actual disk usage inside the pod
kubectl exec -it -n lancedb deployment/lancedb -- df -h /data
\`\`\`

### 2. Analyze Disk Usage

\`\`\`bash
# Check directory sizes
kubectl exec -it -n lancedb deployment/lancedb -- du -sh /data/*
kubectl exec -it -n lancedb deployment/lancedb -- du -sh /data/lancedb/*

# Find large files
kubectl exec -it -n lancedb deployment/lancedb -- find /data -type f -size +100M -exec ls -lh {} \\;
\`\`\`

### 3. Check Database Growth Trends

Review metrics in Grafana:
- Disk usage over time
- Data growth rate
- Number of tables/records

## 🔧 Solutions

### Solution 1: Expand PVC (Recommended)

Expand the existing PVC if the storage class supports it:

\`\`\`bash
# Check if storage class supports volume expansion
kubectl get storageclass standard -o yaml | grep allowVolumeExpansion

# Edit PVC to increase size
kubectl edit pvc lancedb-pvc -n lancedb

# Update spec.resources.requests.storage:
# storage: 100Gi  # Increase from 50Gi

# Wait for expansion to complete
kubectl get pvc -n lancedb -w
\`\`\`

### Solution 2: Clean Up Old Data

If data can be safely deleted:

\`\`\`bash
# List tables
kubectl exec -it -n lancedb deployment/lancedb -- ls -lah /data/lancedb/

# Remove old/unused tables (be careful!)
# kubectl exec -it -n lancedb deployment/lancedb -- rm -rf /data/lancedb/old_table_name
\`\`\`

### Solution 3: Implement Data Retention

Set up automated data retention:

\`\`\`python
import lancedb
from datetime import datetime, timedelta

db = lancedb.connect("http://lancedb.lancedb.svc.cluster.local:8000")

# Example: Delete data older than 30 days
retention_days = 30
cutoff_date = datetime.now() - timedelta(days=retention_days)

# Implement deletion logic based on your schema
# table = db.open_table("your_table")
# table.delete("date < '{}'".format(cutoff_date.isoformat()))
\`\`\`

### Solution 4: Archive to S3/MinIO

Move old data to object storage:

1. Set up MinIO or S3 bucket
2. Export old LanceDB tables
3. Store in object storage
4. Delete from local PVC

### Solution 5: Migrate to Larger PVC

If expansion is not possible:

\`\`\`bash
# 1. Scale down deployment
kubectl scale deployment lancedb -n lancedb --replicas=0

# 2. Backup data
kubectl exec -it -n lancedb deployment/lancedb -- tar -czf /data/backup.tar.gz /data/lancedb

# 3. Create new larger PVC
# Edit pvc.yaml and change size to 100Gi
kubectl apply -f flux/clusters/homelab/infrastructure/lancedb/pvc.yaml

# 4. Restore data to new PVC

# 5. Scale up deployment
kubectl scale deployment lancedb -n lancedb --replicas=1
\`\`\`

## 📊 Monitoring

Set up monitoring for:
- Disk usage trends
- Growth rate (GB/day)
- Largest tables
- Data age distribution

## ✅ Resolution Verification

1. Check disk usage:
   \`\`\`bash
   kubectl exec -it -n lancedb deployment/lancedb -- df -h /data
   \`\`\`

2. Verify PVC size:
   \`\`\`bash
   kubectl get pvc -n lancedb
   \`\`\`

3. Monitor for 24 hours to ensure issue doesn't recur

## 📝 Prevention

- Implement automated data retention
- Set up capacity planning alerts
- Regular data archival
- Monitor data growth trends
- Document expected growth rates

## 🔗 Related Runbooks

- [LanceDB Down](./lancedb-down.md)
- [High Memory Usage](./high-memory-usage.md)
- [Pod Crash Loop](./pod-crash-loop.md)

