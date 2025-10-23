# StatefulSet Migration Guide - Fix EmptyDir Data Loss

**Priority**: 🔴 P0 - IMMEDIATE  
**Current Issue**: EmptyDir causes complete data loss on pod restart  
**Estimated Time**: 5 days

> **Source**: AI Senior SRE Review - Critical Data Loss Issue

---

## The Problem

```yaml
# ❌ CURRENT (BROKEN) - DATA LOSS ON EVERY RESTART
volumes:
- name: lancedb-data
  emptyDir: {}
```

**Impact**: Every pod restart deletes all conversations, memory, and knowledge.

---

## The Solution

```yaml
# ✅ REQUIRED - StatefulSet with PersistentVolumeClaims
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: agent-bruno
spec:
  volumeClaimTemplates:
  - metadata:
      name: lancedb-data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 100Gi
```

---

## Quick Migration

```bash
# 1. Run migration script
./scripts/migrate-to-statefulset.sh

# 2. Verify
kubectl get statefulset -n agent-bruno
kubectl get pvc -n agent-bruno

# 3. Test persistence
kubectl exec agent-bruno-0 -n agent-bruno -- \
  bash -c "echo 'test' > /data/lancedb/test.txt"
kubectl delete pod agent-bruno-0 -n agent-bruno
# Wait for restart
kubectl exec agent-bruno-0 -n agent-bruno -- cat /data/lancedb/test.txt
# Should output: test ✅
```

---

**Full migration details**: See [ARCHITECTURE.md](./ARCHITECTURE.md#-ai-senior-sre-review) SRE Review section.

