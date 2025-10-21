# 🚨 Runbook: Redis Persistence Issues

## Alert Information

**Alert Name:** `RedisPersistenceFailure`  
**Severity:** High  
**Component:** redis  
**Service:** redis-master  

## Symptom

Redis persistence (RDB snapshots or AOF) is failing, risking data loss on restart.

## Impact

- **User Impact:** LOW (immediate) - No impact while Redis is running
- **Business Impact:** HIGH (potential) - Risk of complete data loss on crash/restart
- **Data Impact:** CRITICAL - All in-memory data at risk

## Diagnosis

### 1. Check Persistence Configuration

```bash
# Check RDB configuration
kubectl exec -n redis redis-master-0 -- redis-cli config get save
kubectl exec -n redis redis-master-0 -- redis-cli config get dir
kubectl exec -n redis redis-master-0 -- redis-cli config get dbfilename

# Check AOF configuration
kubectl exec -n redis redis-master-0 -- redis-cli config get appendonly
kubectl exec -n redis redis-master-0 -- redis-cli config get appendfilename
kubectl exec -n redis redis-master-0 -- redis-cli config get appendfsync
```

### 2. Check Persistence Status

```bash
# Get persistence info
kubectl exec -n redis redis-master-0 -- redis-cli info persistence

# Key metrics:
# - rdb_last_save_time: Unix timestamp of last save
# - rdb_last_bgsave_status: Success/failure of last BGSAVE
# - rdb_last_bgsave_time_sec: Duration of last save
# - aof_enabled: AOF status
# - aof_last_rewrite_time_sec: Duration of last rewrite
# - aof_last_bgrewrite_status: Success/failure
# - aof_last_write_status: Success/failure
```

### 3. Check Persistence Files

```bash
# List persistence files
kubectl exec -n redis redis-master-0 -- ls -lh /data/

# Check RDB file
kubectl exec -n redis redis-master-0 -- ls -lh /data/dump.rdb

# Check AOF file
kubectl exec -n redis redis-master-0 -- ls -lh /data/appendonly.aof

# Check disk space
kubectl exec -n redis redis-master-0 -- df -h /data/
```

### 4. Check PVC Status

```bash
# Check Persistent Volume Claim
kubectl get pvc -n redis
kubectl describe pvc -n redis redis-data-redis-master-0

# Check Persistent Volume
kubectl get pv | grep redis
```

### 5. Check Redis Logs

```bash
# Look for persistence errors
kubectl logs -n redis redis-master-0 --tail=200 | grep -i "bgsave\|aof\|persist\|write"

# Common errors:
# - "Can't save in background: fork: Cannot allocate memory"
# - "Background saving error"
# - "AOF write error"
# - "Disk full"
```

### 6. Check Resource Limits

```bash
# Check memory limits
kubectl get pod -n redis redis-master-0 -o jsonpath='{.spec.containers[0].resources}'

# Check actual usage
kubectl top pod -n redis redis-master-0
```

## Resolution Steps

### Step 1: Identify the Issue Type

Common persistence failures:

1. **Disk Full** - No space for RDB/AOF
2. **Fork Failure** - Cannot fork for BGSAVE (memory)
3. **Permission Issues** - Cannot write to /data
4. **PVC Issues** - Volume not mounted or corrupted
5. **AOF Corruption** - AOF file corrupted

### Step 2: Fix Specific Issues

#### Issue: Disk Full
**Cause:** PVC size too small or filled with old backups  
**Fix:**
```bash
# Check disk usage
kubectl exec -n redis redis-master-0 -- df -h /data/

# List files
kubectl exec -n redis redis-master-0 -- ls -lh /data/

# Remove old backup files if any
kubectl exec -n redis redis-master-0 -- rm /data/dump.rdb.old
kubectl exec -n redis redis-master-0 -- rm /data/appendonly.aof.old

# Resize PVC (if supported by storage class)
kubectl edit pvc -n redis redis-data-redis-master-0
# Update spec.resources.requests.storage: 16Gi

# Or update HelmRelease
kubectl edit helmrelease redis -n redis
# Update:
#   master:
#     persistence:
#       size: 16Gi
```

#### Issue: Fork Failure (Cannot Allocate Memory)
**Cause:** Not enough memory for fork() during BGSAVE  
**Fix:**
```bash
# Redis needs 2x memory for fork
# Check current memory
kubectl exec -n redis redis-master-0 -- redis-cli info memory | grep used_memory_human

# Increase pod memory limits
kubectl edit helmrelease redis -n redis
# Update:
#   master:
#     resources:
#       limits:
#         memory: "2Gi"  # Double of used memory
#       requests:
#         memory: "1Gi"

# Temporary: Disable persistence (careful!)
kubectl exec -n redis redis-master-0 -- redis-cli config set save ""

# Or reduce save frequency
kubectl exec -n redis redis-master-0 -- redis-cli config set save "3600 1 1800 10"
```

#### Issue: Permission Denied
**Cause:** Redis cannot write to /data directory  
**Fix:**
```bash
# Check permissions
kubectl exec -n redis redis-master-0 -- ls -ld /data/
kubectl exec -n redis redis-master-0 -- ls -l /data/

# Check file ownership
kubectl exec -n redis redis-master-0 -- id

# Fix permissions
kubectl exec -n redis redis-master-0 -- chown -R redis:redis /data/
kubectl exec -n redis redis-master-0 -- chmod 755 /data/

# Or fix in HelmRelease with init container
kubectl edit helmrelease redis -n redis
# Add security context:
#   master:
#     podSecurityContext:
#       fsGroup: 1000
#     containerSecurityContext:
#       runAsUser: 1000
```

#### Issue: PVC Not Bound
**Cause:** PVC in Pending state  
**Fix:**
```bash
# Check PVC status
kubectl get pvc -n redis redis-data-redis-master-0

# Check events
kubectl describe pvc -n redis redis-data-redis-master-0

# Check storage class
kubectl get storageclass

# If no storage class, install one
# For kind/local: use hostpath-provisioner
# For cloud: use cloud provider's storage class

# Delete and recreate pod to retry binding
kubectl delete pod -n redis redis-master-0
```

#### Issue: AOF Corruption
**Cause:** AOF file corrupted, Redis won't start  
**Fix:**
```bash
# Check AOF status
kubectl exec -n redis redis-master-0 -- redis-cli info persistence | grep aof

# Check for AOF errors in logs
kubectl logs -n redis redis-master-0 | grep -i "aof"

# Try to fix AOF
kubectl exec -n redis redis-master-0 -- redis-check-aof --fix /data/appendonly.aof

# If cannot fix, backup and remove AOF
kubectl exec -n redis redis-master-0 -- cp /data/appendonly.aof /data/appendonly.aof.corrupt
kubectl exec -n redis redis-master-0 -- rm /data/appendonly.aof

# Restart Redis
kubectl delete pod -n redis redis-master-0

# Redis will start with RDB backup if available
```

#### Issue: RDB Corruption
**Cause:** RDB file corrupted  
**Fix:**
```bash
# Check RDB file
kubectl exec -n redis redis-master-0 -- redis-check-rdb /data/dump.rdb

# If corrupted, remove and restart
kubectl exec -n redis redis-master-0 -- mv /data/dump.rdb /data/dump.rdb.corrupt

# Restart Redis (will start empty!)
kubectl delete pod -n redis redis-master-0

# Consider restoring from backup
```

### Step 3: Test Persistence

```bash
# Force a save
kubectl exec -n redis redis-master-0 -- redis-cli bgsave

# Check if successful
kubectl exec -n redis redis-master-0 -- redis-cli lastsave

# Check for errors
kubectl logs -n redis redis-master-0 --tail=50 | grep -i bgsave

# Verify file created
kubectl exec -n redis redis-master-0 -- ls -lh /data/dump.rdb
```

### Step 4: Enable Proper Persistence

Configure both RDB and AOF for best durability:

```bash
# Edit HelmRelease
kubectl edit helmrelease redis -n redis
```

Add configuration:

```yaml
master:
  persistence:
    enabled: true
    size: 16Gi
    storageClass: "standard"
  
  configuration: |
    # RDB Snapshots (point-in-time backups)
    save 900 1      # Save after 15min if 1 key changed
    save 300 10     # Save after 5min if 10 keys changed
    save 60 10000   # Save after 1min if 10000 keys changed
    
    # AOF (Append Only File - more durable)
    appendonly yes
    appendfilename "appendonly.aof"
    appendfsync everysec  # Good balance of performance and durability
    
    # AOF rewrite settings
    auto-aof-rewrite-percentage 100
    auto-aof-rewrite-min-size 64mb
    
    # Directories
    dir /data
    dbfilename dump.rdb
```

Apply changes:

```bash
flux reconcile helmrelease redis -n redis
```

## Verification

### 1. Verify Persistence Enabled

```bash
kubectl exec -n redis redis-master-0 -- redis-cli config get save
kubectl exec -n redis redis-master-0 -- redis-cli config get appendonly
# Should show: yes
```

### 2. Test Manual Save

```bash
# Trigger BGSAVE
kubectl exec -n redis redis-master-0 -- redis-cli bgsave
sleep 5

# Check status
kubectl exec -n redis redis-master-0 -- redis-cli info persistence | grep rdb_last_bgsave_status
# Should show: ok

# Check file exists
kubectl exec -n redis redis-master-0 -- ls -lh /data/dump.rdb
```

### 3. Test Data Recovery

```bash
# Add test data
kubectl exec -n redis redis-master-0 -- redis-cli set test:persistence "test-value-$(date +%s)"

# Force save
kubectl exec -n redis redis-master-0 -- redis-cli bgsave
sleep 5

# Restart Redis
kubectl delete pod -n redis redis-master-0
kubectl wait --for=condition=ready pod -n redis redis-master-0 --timeout=5m

# Verify data persisted
kubectl exec -n redis redis-master-0 -- redis-cli get test:persistence
# Should return the test value
```

### 4. Verify PVC Mounted

```bash
kubectl exec -n redis redis-master-0 -- df -h /data/
kubectl exec -n redis redis-master-0 -- mount | grep /data
```

### 5. Monitor Persistence Operations

```bash
# Watch persistence info
watch -n 5 "kubectl exec -n redis redis-master-0 -- redis-cli info persistence"
```

## Prevention

### 1. Enable Both RDB and AOF

Use both for maximum durability:
- **RDB:** Fast restarts, compact, point-in-time snapshots
- **AOF:** Better durability, every write logged

```yaml
master:
  configuration: |
    # Enable both
    save 900 1
    appendonly yes
    appendfsync everysec
```

### 2. Proper Resource Allocation

```yaml
master:
  resources:
    limits:
      memory: "2Gi"  # 2x of data size for fork
    requests:
      memory: "1Gi"
  
  persistence:
    size: 16Gi  # 2-3x of expected data size
```

### 3. Set Up Monitoring

```yaml
# Prometheus alerts
- alert: RedisPersistenceFailure
  expr: |
    redis_rdb_last_save_timestamp_seconds < (time() - 3600)
  for: 10m
  labels:
    severity: critical
  annotations:
    summary: "Redis hasn't saved to disk in over 1 hour"

- alert: RedisAOFFailure
  expr: redis_aof_last_rewrite_status == 0
  labels:
    severity: high
  annotations:
    summary: "Redis AOF rewrite failed"
```

### 4. Regular Backup Jobs

```yaml
# CronJob for Redis backups
apiVersion: batch/v1
kind: CronJob
metadata:
  name: redis-backup
  namespace: redis
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: redis:7.2
            command:
            - /bin/sh
            - -c
            - |
              # Trigger save
              redis-cli -h redis-master bgsave
              
              # Wait for save to complete
              while [ "$(redis-cli -h redis-master info persistence | grep rdb_bgsave_in_progress | cut -d: -f2 | tr -d '\r')" = "1" ]; do
                sleep 5
              done
              
              # Copy to backup location
              DATE=$(date +%Y%m%d-%H%M%S)
              cp /data/dump.rdb /backup/dump-${DATE}.rdb
              
              # Keep only last 7 days
              find /backup -name "dump-*.rdb" -mtime +7 -delete
            volumeMounts:
            - name: redis-data
              mountPath: /data
            - name: backup-storage
              mountPath: /backup
          volumes:
          - name: redis-data
            persistentVolumeClaim:
              claimName: redis-data-redis-master-0
          - name: backup-storage
            persistentVolumeClaim:
              claimName: redis-backup-pvc
          restartPolicy: OnFailure
```

### 5. Test Recovery Regularly

```bash
#!/bin/bash
# test-redis-recovery.sh

echo "Testing Redis persistence and recovery..."

# Add test data
kubectl exec -n redis redis-master-0 -- redis-cli set test:recovery "$(date)"

# Force save
kubectl exec -n redis redis-master-0 -- redis-cli bgsave

# Wait for save to complete
sleep 10

# Restart Redis
kubectl delete pod -n redis redis-master-0
kubectl wait --for=condition=ready pod -n redis redis-master-0 --timeout=5m

# Verify data
VALUE=$(kubectl exec -n redis redis-master-0 -- redis-cli get test:recovery)

if [ -n "$VALUE" ]; then
  echo "✅ Recovery test passed: $VALUE"
else
  echo "❌ Recovery test failed: Data lost!"
  exit 1
fi
```

### 6. Use Volume Snapshots

```yaml
# VolumeSnapshot for backup
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: redis-snapshot-$(date +%Y%m%d)
  namespace: redis
spec:
  volumeSnapshotClassName: csi-snapshots
  source:
    persistentVolumeClaimName: redis-data-redis-master-0
```

## Backup and Restore Procedures

### Manual Backup

```bash
# Trigger save
kubectl exec -n redis redis-master-0 -- redis-cli bgsave

# Wait for completion
kubectl exec -n redis redis-master-0 -- redis-cli lastsave

# Copy from pod
kubectl cp redis/redis-master-0:/data/dump.rdb ./redis-backup-$(date +%Y%m%d).rdb
```

### Manual Restore

```bash
# Stop Redis
kubectl scale statefulset -n redis redis-master --replicas=0

# Copy backup to pod (use init container or manually)
kubectl cp ./redis-backup.rdb redis/redis-master-0:/data/dump.rdb

# Start Redis
kubectl scale statefulset -n redis redis-master --replicas=1

# Verify data
kubectl exec -n redis redis-master-0 -- redis-cli dbsize
```

## Configuration Best Practices

### For Session Storage (Ephemeral)

```yaml
master:
  configuration: |
    # Minimal persistence
    save 3600 1  # Save once per hour if any changes
    appendonly no
    maxmemory-policy allkeys-lru
```

### For Cache (Semi-Persistent)

```yaml
master:
  configuration: |
    # Moderate persistence
    save 900 1
    save 300 10
    appendonly yes
    appendfsync everysec
    maxmemory-policy allkeys-lfu
```

### For Critical Data (High Durability)

```yaml
master:
  configuration: |
    # Maximum durability
    save 300 1    # Save every 5min
    save 60 100
    appendonly yes
    appendfsync always  # Slower but most durable
    
    # Enable replication
    replica:
      replicaCount: 2
```

## Related Alerts

- `RedisDown`
- `RedisDiskFull`
- `RedisMemoryHigh`
- `RedisBGSaveFailure`
- `RedisAOFRewriteFailure`

## Escalation

If persistence issues cannot be resolved:

1. ✅ Verify all resolution steps
2. 💾 Check underlying storage health
3. 🔍 Review Kubernetes storage class configuration
4. 📊 Analyze resource limits and usage
5. 🔄 Consider migrating to managed Redis
6. 📞 Contact storage/infrastructure team
7. 🆘 Prepare for emergency data recovery

## Additional Resources

- [Redis Persistence](https://redis.io/docs/management/persistence/)
- [Redis Backup Guide](https://redis.io/docs/management/backup/)
- [AOF vs RDB](https://redis.io/docs/management/persistence/#aof-vs-rdb)
- [Redis Disaster Recovery](https://redis.io/docs/management/backup/#disaster-recovery)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

