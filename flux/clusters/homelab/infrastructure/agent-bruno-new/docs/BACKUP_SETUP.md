# Automated Backup Setup Guide

**Priority**: 🔴 P0 - HIGH  
**Current State**: No backups, no disaster recovery  
**Estimated Time**: 1 week  
**RTO Target**: < 15 minutes | **RPO Target**: < 1 hour

> **Source**: AI Senior SRE Review

---

## Quick Start with Velero

```bash
# 1. Install Velero
helm install velero vmware-tanzu/velero \
  --namespace velero \
  --create-namespace \
  --set configuration.provider=aws \
  --set configuration.backupStorageLocation.bucket=agent-bruno-backups

# 2. Create hourly backup schedule
velero schedule create lancedb-hourly \
  --schedule="0 * * * *" \
  --include-namespaces agent-bruno \
  --ttl 168h

# 3. Verify
velero schedule get
velero backup get
```

---

## Restore Procedure

```bash
# Full disaster recovery
LATEST=$(velero backup get -o json | jq -r '.items[-1].metadata.name')
velero restore create --from-backup $LATEST --wait

# Verify
kubectl get all -n agent-bruno
```

---

**Full backup strategy**: See [ARCHITECTURE.md](./ARCHITECTURE.md#disaster-recovery--high-availability) DR section.

