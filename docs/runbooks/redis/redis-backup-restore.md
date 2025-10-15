# 📦 Runbook: Redis Backup and Restore

## Overview

This runbook covers procedures for backing up and restoring Redis data in Kubernetes.

## Backup Types

### 1. RDB (Redis Database) Snapshots
- **Format:** Binary point-in-time snapshot
- **Speed:** Fast to create and restore
- **Size:** Compact
- **Use Case:** Regular backups, disaster recovery

### 2. AOF (Append Only File)
- **Format:** Log of all write operations
- **Speed:** Slower but more durable
- **Size:** Larger than RDB
- **Use Case:** Continuous backup, minimal data loss

### 3. Volume Snapshots
- **Format:** Storage-level snapshot
- **Speed:** Very fast
- **Size:** Depends on storage backend
- **Use Case:** Quick backups, cloning environments

## Manual Backup Procedures

### RDB Backup

#### Create RDB Snapshot

```bash
# Trigger background save
kubectl exec -n redis redis-master-0 -- redis-cli bgsave

# Check save status
kubectl exec -n redis redis-master-0 -- redis-cli lastsave
kubectl exec -n redis redis-master-0 -- redis-cli info persistence | grep rdb_last_bgsave_status

# Wait for save to complete
while [ "$(kubectl exec -n redis redis-master-0 -- redis-cli info persistence | grep rdb_bgsave_in_progress | cut -d: -f2 | tr -d '\r')" = "1" ]; do
  echo "Waiting for BGSAVE to complete..."
  sleep 5
done

echo "BGSAVE completed"
```

#### Download RDB File

```bash
# Copy RDB file from pod
DATE=$(date +%Y%m%d-%H%M%S)
kubectl cp redis/redis-master-0:/data/dump.rdb ./backups/redis-backup-${DATE}.rdb

# Verify backup
ls -lh ./backups/redis-backup-${DATE}.rdb

# Compress backup (optional)
gzip ./backups/redis-backup-${DATE}.rdb
```

#### Verify RDB File

```bash
# Check RDB file integrity
kubectl exec -n redis redis-master-0 -- redis-check-rdb /data/dump.rdb
```

### AOF Backup

```bash
# Trigger AOF rewrite (compacts AOF)
kubectl exec -n redis redis-master-0 -- redis-cli bgrewriteaof

# Wait for rewrite to complete
kubectl exec -n redis redis-master-0 -- redis-cli info persistence | grep aof_rewrite_in_progress

# Copy AOF file
DATE=$(date +%Y%m%d-%H%M%S)
kubectl cp redis/redis-master-0:/data/appendonly.aof ./backups/redis-aof-${DATE}.aof

# Compress
gzip ./backups/redis-aof-${DATE}.aof
```

## Automated Backup

### Kubernetes CronJob for RDB Backups

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-backup-pvc
  namespace: redis
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: redis-backup
  namespace: redis
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  successfulJobsHistoryLimit: 7
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            app: redis-backup
        spec:
          containers:
          - name: backup
            image: redis:7.2
            command:
            - /bin/sh
            - -c
            - |
              set -e
              
              echo "Starting Redis backup at $(date)"
              
              # Trigger BGSAVE
              redis-cli -h redis-master bgsave
              
              # Wait for BGSAVE to complete
              while [ "$(redis-cli -h redis-master info persistence | grep rdb_bgsave_in_progress | cut -d: -f2 | tr -d '\r')" = "1" ]; do
                echo "Waiting for BGSAVE..."
                sleep 5
              done
              
              # Check if BGSAVE was successful
              STATUS=$(redis-cli -h redis-master info persistence | grep rdb_last_bgsave_status | cut -d: -f2 | tr -d '\r')
              if [ "$STATUS" != "ok" ]; then
                echo "ERROR: BGSAVE failed"
                exit 1
              fi
              
              # Create dated backup
              DATE=$(date +%Y%m%d-%H%M%S)
              cp /data/dump.rdb /backup/dump-${DATE}.rdb
              
              # Create latest symlink
              ln -sf dump-${DATE}.rdb /backup/dump-latest.rdb
              
              # Compress old backups
              find /backup -name "dump-*.rdb" -mtime +1 -exec gzip {} \;
              
              # Keep only last 30 days
              find /backup -name "dump-*.rdb.gz" -mtime +30 -delete
              
              # Log backup info
              echo "Backup completed: /backup/dump-${DATE}.rdb"
              ls -lh /backup/ | tail -10
              
            volumeMounts:
            - name: redis-data
              mountPath: /data
              readOnly: true
            - name: backup-storage
              mountPath: /backup
            env:
            - name: REDIS_CLI_CONNECT_TIMEOUT
              value: "5"
          volumes:
          - name: redis-data
            persistentVolumeClaim:
              claimName: redis-data-redis-master-0
          - name: backup-storage
            persistentVolumeClaim:
              claimName: redis-backup-pvc
          restartPolicy: OnFailure
```

### Backup to S3/MinIO

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: redis-backup-s3
  namespace: redis
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup-to-s3
            image: amazon/aws-cli:latest
            command:
            - /bin/sh
            - -c
            - |
              set -e
              
              DATE=$(date +%Y%m%d-%H%M%S)
              BACKUP_FILE="redis-backup-${DATE}.rdb"
              
              # Wait for Redis backup to complete
              while [ ! -f /backup/dump-latest.rdb ]; do
                echo "Waiting for backup file..."
                sleep 10
              done
              
              # Copy and compress
              cp /backup/dump-latest.rdb /tmp/${BACKUP_FILE}
              gzip /tmp/${BACKUP_FILE}
              
              # Upload to S3
              aws s3 cp /tmp/${BACKUP_FILE}.gz s3://${S3_BUCKET}/redis/${BACKUP_FILE}.gz
              
              # Cleanup old backups (keep last 30 days)
              aws s3 ls s3://${S3_BUCKET}/redis/ | \
                awk '{print $4}' | \
                head -n -30 | \
                xargs -I {} aws s3 rm s3://${S3_BUCKET}/redis/{}
              
              echo "Backup uploaded to S3: ${BACKUP_FILE}.gz"
              
            volumeMounts:
            - name: backup-storage
              mountPath: /backup
              readOnly: true
            env:
            - name: AWS_ACCESS_KEY_ID
              valueFrom:
                secretKeyRef:
                  name: s3-credentials
                  key: access-key-id
            - name: AWS_SECRET_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: s3-credentials
                  key: secret-access-key
            - name: AWS_DEFAULT_REGION
              value: "us-east-1"
            - name: S3_BUCKET
              value: "my-redis-backups"
          volumes:
          - name: backup-storage
            persistentVolumeClaim:
              claimName: redis-backup-pvc
          restartPolicy: OnFailure
```

## Restore Procedures

### Restore from RDB Backup

#### Method 1: Direct File Replacement

```bash
# 1. Scale down Redis
kubectl scale statefulset -n redis redis-master --replicas=0

# Wait for pod to terminate
kubectl wait --for=delete pod/redis-master-0 -n redis --timeout=60s

# 2. Create a restore job
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: redis-restore
  namespace: redis
spec:
  containers:
  - name: restore
    image: busybox
    command: ['sh', '-c', 'sleep 3600']
    volumeMounts:
    - name: redis-data
      mountPath: /data
  volumes:
  - name: redis-data
    persistentVolumeClaim:
      claimName: redis-data-redis-master-0
EOF

# Wait for restore pod
kubectl wait --for=condition=ready pod/redis-restore -n redis --timeout=60s

# 3. Copy backup to pod
kubectl cp ./backups/redis-backup-20251015.rdb redis/redis-restore:/data/dump.rdb

# 4. Set correct permissions
kubectl exec -n redis redis-restore -- chown 1000:1000 /data/dump.rdb
kubectl exec -n redis redis-restore -- chmod 644 /data/dump.rdb

# 5. Delete restore pod
kubectl delete pod redis-restore -n redis

# 6. Scale up Redis
kubectl scale statefulset -n redis redis-master --replicas=1

# 7. Wait for Redis to start
kubectl wait --for=condition=ready pod/redis-master-0 -n redis --timeout=5m

# 8. Verify data restored
kubectl exec -n redis redis-master-0 -- redis-cli dbsize
kubectl exec -n redis redis-master-0 -- redis-cli info keyspace
```

#### Method 2: Init Container (Permanent Setup)

Add init container to Redis HelmRelease:

```yaml
master:
  initContainers:
  - name: restore-from-backup
    image: busybox
    command:
    - sh
    - -c
    - |
      if [ -f /backup/restore-trigger ]; then
        echo "Restore triggered, copying backup..."
        if [ -f /backup/dump-latest.rdb ]; then
          cp /backup/dump-latest.rdb /data/dump.rdb
          chown 1000:1000 /data/dump.rdb
          rm /backup/restore-trigger
          echo "Restore completed"
        else
          echo "ERROR: Backup file not found"
          exit 1
        fi
      else
        echo "No restore trigger found, skipping"
      fi
    volumeMounts:
    - name: redis-data
      mountPath: /data
    - name: backup-storage
      mountPath: /backup
```

To trigger restore:

```bash
# 1. Copy backup to backup volume
kubectl cp ./backups/redis-backup-20251015.rdb redis/redis-backup-pod:/backup/dump-latest.rdb

# 2. Create restore trigger
kubectl exec -n redis redis-backup-pod -- touch /backup/restore-trigger

# 3. Restart Redis
kubectl rollout restart statefulset redis-master -n redis
```

### Restore from AOF

```bash
# 1. Scale down Redis
kubectl scale statefulset -n redis redis-master --replicas=0

# 2. Copy AOF file
kubectl cp ./backups/redis-aof-20251015.aof redis/redis-restore:/data/appendonly.aof

# 3. Enable AOF in config
kubectl exec -n redis redis-restore -- sh -c 'echo "appendonly yes" > /data/redis.conf'

# 4. Scale up Redis
kubectl scale statefulset -n redis redis-master --replicas=1

# Redis will replay AOF on startup
```

### Restore from Volume Snapshot

```bash
# 1. List available snapshots
kubectl get volumesnapshot -n redis

# 2. Create new PVC from snapshot
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-data-restored
  namespace: redis
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 8Gi
  dataSource:
    name: redis-snapshot-20251015
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
EOF

# 3. Update StatefulSet to use new PVC
kubectl edit statefulset redis-master -n redis
# Change volumeClaimTemplates to reference redis-data-restored

# 4. Restart Redis
kubectl rollout restart statefulset redis-master -n redis
```

## Verification After Restore

```bash
# 1. Check Redis is running
kubectl get pod -n redis redis-master-0

# 2. Check database size
kubectl exec -n redis redis-master-0 -- redis-cli dbsize

# 3. Check keyspace info
kubectl exec -n redis redis-master-0 -- redis-cli info keyspace

# 4. Sample some keys
kubectl exec -n redis redis-master-0 -- redis-cli --scan | head -10

# 5. Verify specific data
kubectl exec -n redis redis-master-0 -- redis-cli get "some:known:key"

# 6. Check persistence status
kubectl exec -n redis redis-master-0 -- redis-cli info persistence

# 7. Test write operations
kubectl exec -n redis redis-master-0 -- redis-cli set test:restore "$(date)"
kubectl exec -n redis redis-master-0 -- redis-cli get test:restore
```

## Backup Best Practices

### 1. Multiple Backup Strategies

Use combination of:
- ✅ RDB for regular snapshots (every 6-12 hours)
- ✅ AOF for continuous backup (if durability critical)
- ✅ Volume snapshots for quick recovery
- ✅ Off-cluster backups (S3/MinIO) for disaster recovery

### 2. Backup Retention

```
Hourly:  Keep last 24 hours
Daily:   Keep last 7 days  
Weekly:  Keep last 4 weeks
Monthly: Keep last 12 months
```

### 3. Test Restores Regularly

```bash
#!/bin/bash
# test-redis-restore.sh

echo "Testing Redis restore from backup..."

# Create test namespace
kubectl create namespace redis-test

# Restore Redis from backup to test namespace
# ... (restore steps) ...

# Verify data
TEST_DBSIZE=$(kubectl exec -n redis-test redis-master-0 -- redis-cli dbsize)
PROD_DBSIZE=$(kubectl exec -n redis redis-master-0 -- redis-cli dbsize)

if [ "$TEST_DBSIZE" -eq "$PROD_DBSIZE" ]; then
  echo "✅ Restore test passed: $TEST_DBSIZE keys"
else
  echo "❌ Restore test failed: Expected $PROD_DBSIZE, got $TEST_DBSIZE"
fi

# Cleanup
kubectl delete namespace redis-test
```

### 4. Monitor Backups

```yaml
# Prometheus alerts
- alert: RedisBackupFailed
  expr: time() - redis_last_backup_timestamp_seconds > 86400
  labels:
    severity: critical
  annotations:
    summary: "Redis backup hasn't succeeded in 24 hours"

- alert: RedisBackupJobFailed
  expr: kube_job_status_failed{job_name=~"redis-backup.*"} > 0
  labels:
    severity: high
  annotations:
    summary: "Redis backup job failed"
```

## Recovery Time Objectives (RTO)

| Backup Method | RTO | Data Loss (RPO) |
|---------------|-----|-----------------|
| RDB Restore | 5-15 min | Last snapshot |
| AOF Restore | 10-30 min | Minimal (<1s) |
| Volume Snapshot | 2-5 min | Last snapshot |
| S3 Restore | 15-60 min | Last backup |

## Disaster Recovery Scenarios

### Scenario 1: Redis Pod Deleted

**RTO:** 2 minutes  
**RPO:** None (data persisted)

```bash
# StatefulSet automatically recreates pod
kubectl get pod -n redis -w

# Data loaded from PVC automatically
```

### Scenario 2: Data Corruption

**RTO:** 10-30 minutes  
**RPO:** Last backup (6-12 hours)

```bash
# Restore from latest backup
./restore-from-backup.sh latest
```

### Scenario 3: Complete Cluster Loss

**RTO:** 1-2 hours  
**RPO:** Last off-cluster backup

```bash
# 1. Rebuild cluster
# 2. Install Redis
# 3. Restore from S3
aws s3 cp s3://my-backups/redis/dump-latest.rdb.gz ./
gunzip dump-latest.rdb.gz
# 4. Restore to Redis
# (follow restore procedures above)
```

### Scenario 4: Accidental Data Deletion

**RTO:** 15-30 minutes  
**RPO:** Point-in-time of last backup

```bash
# Restore from specific backup
./restore-from-backup.sh 20251015-140000
```

## Backup Scripts

### Complete Backup Script

```bash
#!/bin/bash
# redis-backup.sh

set -e

NAMESPACE="redis"
POD="redis-master-0"
BACKUP_DIR="./backups"
S3_BUCKET="${S3_BUCKET:-my-redis-backups}"
DATE=$(date +%Y%m%d-%H%M%S)

echo "Starting Redis backup at $(date)"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Trigger BGSAVE
echo "Triggering BGSAVE..."
kubectl exec -n "$NAMESPACE" "$POD" -- redis-cli bgsave

# Wait for completion
echo "Waiting for BGSAVE to complete..."
while [ "$(kubectl exec -n $NAMESPACE $POD -- redis-cli info persistence | grep rdb_bgsave_in_progress | cut -d: -f2 | tr -d '\r')" = "1" ]; do
  sleep 5
done

# Check status
STATUS=$(kubectl exec -n $NAMESPACE $POD -- redis-cli info persistence | grep rdb_last_bgsave_status | cut -d: -f2 | tr -d '\r')
if [ "$STATUS" != "ok" ]; then
  echo "ERROR: BGSAVE failed"
  exit 1
fi

# Copy backup
echo "Copying backup..."
kubectl cp "$NAMESPACE/$POD:/data/dump.rdb" "$BACKUP_DIR/redis-backup-${DATE}.rdb"

# Compress
echo "Compressing backup..."
gzip "$BACKUP_DIR/redis-backup-${DATE}.rdb"

# Upload to S3 (if configured)
if [ -n "$AWS_ACCESS_KEY_ID" ]; then
  echo "Uploading to S3..."
  aws s3 cp "$BACKUP_DIR/redis-backup-${DATE}.rdb.gz" "s3://${S3_BUCKET}/redis/"
fi

# Cleanup old local backups (keep last 7 days)
find "$BACKUP_DIR" -name "redis-backup-*.rdb.gz" -mtime +7 -delete

echo "✅ Backup completed: redis-backup-${DATE}.rdb.gz"
ls -lh "$BACKUP_DIR/" | tail -5
```

## Related Alerts

- `RedisBackupFailed`
- `RedisBackupOld`
- `RedisPersistenceFailure`
- `RedisDown`

## Additional Resources

- [Redis Persistence Documentation](https://redis.io/docs/management/persistence/)
- [Redis Backup Strategies](https://redis.io/docs/management/backup/)
- [Kubernetes Volume Snapshots](https://kubernetes.io/docs/concepts/storage/volume-snapshots/)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

