# 🚨 Runbook: Redis Replication Issues

## Alert Information

**Alert Name:** `RedisReplicationLag` / `RedisReplicationBroken`  
**Severity:** High  
**Component:** redis  
**Service:** redis-master, redis-replica

## Symptom

Redis replication is experiencing issues: replicas out of sync, replication lag, or broken replication links.

## Impact

- **User Impact:** LOW (immediate) - Reads still work from replicas
- **Business Impact:** MEDIUM - High availability compromised
- **Data Impact:** MEDIUM - Risk of data loss on master failure

## Diagnosis

### 1. Check Replication Status

```bash
# Check master replication info
kubectl exec -n redis redis-master-0 -- redis-cli info replication

# Key metrics:
# - role: master/slave
# - connected_slaves: Number of connected replicas
# - master_repl_offset: Current replication offset
# - repl_backlog_size: Replication backlog buffer size
```

### 2. Check Replica Status

```bash
# If using replicas
kubectl exec -n redis redis-replica-0 -- redis-cli info replication

# Key metrics:
# - role: Should be "slave"
# - master_host: Should point to master
# - master_link_status: Should be "up"
# - master_sync_in_progress: Should be 0
# - master_last_io_seconds_ago: Should be < 10
# - slave_repl_offset: Should be close to master_repl_offset
```

### 3. Calculate Replication Lag

```bash
# Get master offset
MASTER_OFFSET=$(kubectl exec -n redis redis-master-0 -- redis-cli info replication | grep master_repl_offset | cut -d: -f2 | tr -d '\r')

# Get replica offset
REPLICA_OFFSET=$(kubectl exec -n redis redis-replica-0 -- redis-cli info replication | grep slave_repl_offset | cut -d: -f2 | tr -d '\r')

# Calculate lag
LAG=$((MASTER_OFFSET - REPLICA_OFFSET))
echo "Replication lag: $LAG bytes"
```

### 4. Check Network Connectivity

```bash
# Test connectivity from replica to master
kubectl exec -n redis redis-replica-0 -- nc -zv redis-master.redis.svc.cluster.local 6379

# Check DNS resolution
kubectl exec -n redis redis-replica-0 -- nslookup redis-master.redis.svc.cluster.local
```

### 5. Check Redis Logs

```bash
# Master logs
kubectl logs -n redis redis-master-0 --tail=100 | grep -i "repl\|sync\|slave"

# Replica logs
kubectl logs -n redis redis-replica-0 --tail=100 | grep -i "repl\|sync\|master"

# Look for:
# - Connection errors
# - Timeout errors
# - Sync errors
# - PSYNC failures
```

### 6. Check Resource Usage

```bash
# Check if resources are constrained
kubectl top pod -n redis
kubectl describe node $(kubectl get pod -n redis redis-master-0 -o jsonpath='{.spec.nodeName}')
```

## Resolution Steps

### Step 1: Identify the Problem

Common replication issues:

1. **Replication Link Down** - Connection broken between master and replica
2. **High Replication Lag** - Replica falling behind master
3. **Full Resync Loop** - Continuously doing full sync
4. **Partial Resync Failures** - Cannot do partial resync
5. **Network Issues** - DNS, firewall, or network problems

### Step 2: Fix Specific Issues

#### Issue: Replication Link Down (master_link_status: down)
**Cause:** Network issues or replica cannot connect to master  
**Fix:**
```bash
# Check replica configuration
kubectl exec -n redis redis-replica-0 -- redis-cli config get replicaof
# Should show: redis-master.redis.svc.cluster.local 6379

# If wrong, reconfigure
kubectl exec -n redis redis-replica-0 -- redis-cli replicaof redis-master.redis.svc.cluster.local 6379

# Check connectivity
kubectl exec -n redis redis-replica-0 -- ping redis-master.redis.svc.cluster.local
kubectl exec -n redis redis-replica-0 -- nc -zv redis-master.redis.svc.cluster.local 6379

# Restart replica if needed
kubectl delete pod -n redis redis-replica-0
```

#### Issue: High Replication Lag
**Cause:** Master too busy, network slow, or replica underpowered  
**Fix:**
```bash
# Check replication backlog size
kubectl exec -n redis redis-master-0 -- redis-cli config get repl-backlog-size
# Should be large enough: 16mb minimum, 64mb+ recommended

# Increase backlog size
kubectl exec -n redis redis-master-0 -- redis-cli config set repl-backlog-size 67108864  # 64MB

# Check master load
kubectl exec -n redis redis-master-0 -- redis-cli info stats | grep instantaneous_ops_per_sec

# If master is overloaded:
# 1. Scale up resources
# 2. Optimize queries
# 3. Add more replicas to distribute reads

# Increase replica resources
kubectl edit helmrelease redis -n redis
# Update replica resources:
#   replica:
#     resources:
#       limits:
#         cpu: "1000m"
#         memory: "2Gi"
```

#### Issue: Full Resync Loop (Continuous SYNC)
**Cause:** Backlog too small or replication timing out  
**Fix:**
```bash
# Increase replication backlog
kubectl exec -n redis redis-master-0 -- redis-cli config set repl-backlog-size 134217728  # 128MB
kubectl exec -n redis redis-master-0 -- redis-cli config set repl-backlog-ttl 3600  # 1 hour

# Increase replication timeout
kubectl exec -n redis redis-master-0 -- redis-cli config set repl-timeout 300  # 5 minutes

# Make permanent in HelmRelease
kubectl edit helmrelease redis -n redis
# Add:
#   master:
#     configuration: |
#       repl-backlog-size 134217728
#       repl-backlog-ttl 3600
#       repl-timeout 300
```

#### Issue: Partial Resync Failing
**Cause:** Replication ID mismatch or backlog expired  
**Fix:**
```bash
# Check replication ID
kubectl exec -n redis redis-master-0 -- redis-cli info replication | grep master_replid

# Force full resync (will cause temporary lag)
kubectl exec -n redis redis-replica-0 -- redis-cli replicaof no one
sleep 5
kubectl exec -n redis redis-replica-0 -- redis-cli replicaof redis-master.redis.svc.cluster.local 6379

# Monitor sync progress
watch -n 2 "kubectl exec -n redis redis-replica-0 -- redis-cli info replication | grep -E 'master_link_status|master_sync_in_progress|slave_repl_offset'"
```

#### Issue: Diskless Replication Issues
**Cause:** Diskless replication misconfigured  
**Fix:**
```bash
# Check diskless replication settings
kubectl exec -n redis redis-master-0 -- redis-cli config get repl-diskless-sync

# Disable if causing issues
kubectl exec -n redis redis-master-0 -- redis-cli config set repl-diskless-sync no

# Or tune delay
kubectl exec -n redis redis-master-0 -- redis-cli config set repl-diskless-sync-delay 5
```

#### Issue: Network Bandwidth Limitation
**Cause:** Network between master and replica is saturated  
**Fix:**
```bash
# Limit replication bandwidth (bytes per second)
kubectl exec -n redis redis-master-0 -- redis-cli config set repl-diskless-sync-max-replicas 1

# Add configuration
kubectl edit helmrelease redis -n redis
# Add:
#   master:
#     configuration: |
#       # Limit sync bandwidth
#       client-output-buffer-limit slave 256mb 64mb 60
```

#### Issue: Replica Promoted by Accident
**Cause:** Replica was promoted but should be replica  
**Fix:**
```bash
# Check role
kubectl exec -n redis redis-replica-0 -- redis-cli role

# If showing "master", demote back to replica
kubectl exec -n redis redis-replica-0 -- redis-cli replicaof redis-master.redis.svc.cluster.local 6379

# Verify
kubectl exec -n redis redis-replica-0 -- redis-cli info replication | grep role
# Should show: slave
```

### Step 3: Restart Replication

If replication is completely broken:

```bash
# Stop replication on replica
kubectl exec -n redis redis-replica-0 -- redis-cli replicaof no one

# Clear replica data (if acceptable)
kubectl exec -n redis redis-replica-0 -- redis-cli flushall

# Restart replication
kubectl exec -n redis redis-replica-0 -- redis-cli replicaof redis-master.redis.svc.cluster.local 6379

# Monitor sync
kubectl logs -n redis redis-replica-0 -f
```

### Step 4: Configure High Availability with Sentinel

For production, use Redis Sentinel:

```yaml
# In HelmRelease
sentinel:
  enabled: true
  replicas: 3
  quorum: 2
  downAfterMilliseconds: 5000
  failoverTimeout: 10000
  parallelSyncs: 1

replica:
  replicaCount: 2
  resources:
    limits:
      cpu: "1000m"
      memory: "1Gi"
    requests:
      cpu: "100m"
      memory: "512Mi"

master:
  configuration: |
    # Replication settings
    repl-backlog-size 134217728  # 128MB
    repl-backlog-ttl 3600
    repl-timeout 300
    repl-diskless-sync yes
    repl-diskless-sync-delay 5
    min-replicas-to-write 1
    min-replicas-max-lag 10
```

## Verification

### 1. Check Replication Status

```bash
# Master status
kubectl exec -n redis redis-master-0 -- redis-cli info replication

# Should show:
# role:master
# connected_slaves:2 (or your replica count)
```

### 2. Verify Replication Link

```bash
# Replica status
kubectl exec -n redis redis-replica-0 -- redis-cli info replication

# Should show:
# role:slave
# master_link_status:up
# master_last_io_seconds_ago:<10
```

### 3. Test Data Replication

```bash
# Write to master
kubectl exec -n redis redis-master-0 -- redis-cli set test:repl "$(date)"

# Read from replica (with small delay)
sleep 2
kubectl exec -n redis redis-replica-0 -- redis-cli get test:repl

# Should return the same value
```

### 4. Check Replication Lag

```bash
MASTER_OFFSET=$(kubectl exec -n redis redis-master-0 -- redis-cli info replication | grep master_repl_offset | cut -d: -f2 | tr -d '\r')
REPLICA_OFFSET=$(kubectl exec -n redis redis-replica-0 -- redis-cli info replication | grep slave_repl_offset | cut -d: -f2 | tr -d '\r')
LAG=$((MASTER_OFFSET - REPLICA_OFFSET))

echo "Master offset:  $MASTER_OFFSET"
echo "Replica offset: $REPLICA_OFFSET"
echo "Lag:            $LAG bytes"

# Lag should be < 1000 bytes in normal operation
```

### 5. Verify Sentinel (if enabled)

```bash
# Check sentinel status
kubectl exec -n redis redis-sentinel-0 -- redis-cli -p 26379 sentinel masters

# Check sentinel's view of replicas
kubectl exec -n redis redis-sentinel-0 -- redis-cli -p 26379 sentinel replicas redis-master
```

## Prevention

### 1. Proper Configuration

```yaml
master:
  configuration: |
    # Replication
    repl-backlog-size 134217728  # 128MB
    repl-backlog-ttl 3600        # 1 hour
    repl-timeout 300              # 5 minutes
    repl-diskless-sync yes
    repl-diskless-sync-delay 5
    
    # High availability
    min-replicas-to-write 1      # Require at least 1 replica
    min-replicas-max-lag 10      # Max 10s lag
    
    # Output buffer for replicas
    client-output-buffer-limit slave 256mb 64mb 60
```

### 2. Enable Sentinel

```yaml
sentinel:
  enabled: true
  replicas: 3          # Always odd number (3, 5, 7)
  quorum: 2            # Majority for failover
  downAfterMilliseconds: 5000
  failoverTimeout: 60000
  parallelSyncs: 1     # One replica at a time
```

### 3. Resource Allocation

```yaml
replica:
  replicaCount: 2      # At least 2 for HA
  resources:
    limits:
      cpu: "1000m"
      memory: "2Gi"    # Same as master
    requests:
      cpu: "200m"
      memory: "1Gi"
```

### 4. Monitoring

```yaml
# Prometheus alerts
- alert: RedisReplicationBroken
  expr: redis_connected_slaves < 1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Redis has no connected replicas"

- alert: RedisReplicationLag
  expr: |
    (redis_master_repl_offset - 
     on(instance) group_right redis_slave_repl_offset) > 1000000
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Redis replication lag is high"

- alert: RedisReplicaDown
  expr: redis_up{role="slave"} == 0
  for: 5m
  labels:
    severity: high
  annotations:
    summary: "Redis replica is down"
```

### 5. Network Optimization

- Use same availability zone for master and replicas
- Ensure adequate network bandwidth
- Configure proper network policies
- Use persistent connections

### 6. Regular Testing

```bash
#!/bin/bash
# test-redis-replication.sh

echo "Testing Redis replication..."

# Write test data to master
TEST_VALUE="test-$(date +%s)"
kubectl exec -n redis redis-master-0 -- redis-cli set test:repl "$TEST_VALUE"

# Wait for replication
sleep 2

# Read from all replicas
for i in 0 1; do
  VALUE=$(kubectl exec -n redis redis-replica-$i -- redis-cli get test:repl)
  if [ "$VALUE" = "$TEST_VALUE" ]; then
    echo "✅ Replica $i: Replication OK"
  else
    echo "❌ Replica $i: Replication FAILED (expected: $TEST_VALUE, got: $VALUE)"
  fi
done

# Check replication lag
MASTER_OFFSET=$(kubectl exec -n redis redis-master-0 -- redis-cli info replication | grep master_repl_offset | cut -d: -f2 | tr -d '\r')
for i in 0 1; do
  REPLICA_OFFSET=$(kubectl exec -n redis redis-replica-$i -- redis-cli info replication | grep slave_repl_offset | cut -d: -f2 | tr -d '\r')
  LAG=$((MASTER_OFFSET - REPLICA_OFFSET))
  echo "Replica $i lag: $LAG bytes"
done
```

## Failover Procedures

### Manual Failover

```bash
# 1. Choose a replica to promote
# Check which replica is most up-to-date
kubectl exec -n redis redis-replica-0 -- redis-cli info replication | grep slave_repl_offset
kubectl exec -n redis redis-replica-1 -- redis-cli info replication | grep slave_repl_offset

# 2. Promote chosen replica
kubectl exec -n redis redis-replica-0 -- redis-cli replicaof no one

# 3. Point other replicas to new master
kubectl exec -n redis redis-replica-1 -- redis-cli replicaof redis-replica-0.redis.svc.cluster.local 6379

# 4. Update application configuration to point to new master

# 5. Fix old master when it comes back
kubectl exec -n redis redis-master-0 -- redis-cli replicaof redis-replica-0.redis.svc.cluster.local 6379
```

### Sentinel Automatic Failover

```bash
# Check sentinel status
kubectl exec -n redis redis-sentinel-0 -- redis-cli -p 26379 sentinel masters

# Force failover (if needed)
kubectl exec -n redis redis-sentinel-0 -- redis-cli -p 26379 sentinel failover redis-master

# Monitor failover
kubectl logs -n redis redis-sentinel-0 -f
```

## Replication Architecture

### Simple Replication (No Sentinel)

```
┌─────────────┐
│   Master    │ ──┐
└─────────────┘   │
                  │ Replication
┌─────────────┐   │
│  Replica 0  │ ◄─┤
└─────────────┘   │
                  │
┌─────────────┐   │
│  Replica 1  │ ◄─┘
└─────────────┘
```

### Redis Sentinel (Recommended)

```
┌─────────────┐
│   Master    │ ──┐
└─────────────┘   │
      ▲           │ Replication
      │           │
      │ Monitor   │
      │           │
┌─────────────┐   │
│ Sentinel 0  │   │
│ Sentinel 1  │   │    ┌─────────────┐
│ Sentinel 2  │   ├──► │  Replica 0  │
└─────────────┘   │    └─────────────┘
                  │
                  │    ┌─────────────┐
                  └──► │  Replica 1  │
                       └─────────────┘
```

## Related Alerts

- `RedisReplicationBroken`
- `RedisReplicationLag`
- `RedisReplicaDown`
- `RedisSentinelDown`
- `RedisFailoverFailed`

## Escalation

If replication issues persist:

1. ✅ Verify all resolution steps
2. 📊 Analyze network latency between pods
3. 🔍 Check for resource constraints
4. 💾 Review replication configuration
5. 🔄 Consider Redis Cluster for better HA
6. 📞 Contact infrastructure team
7. 🆘 Prepare for manual failover if critical

## Additional Resources

- [Redis Replication](https://redis.io/docs/management/replication/)
- [Redis Sentinel](https://redis.io/docs/management/sentinel/)
- [Redis High Availability](https://redis.io/docs/management/scaling/)
- [Troubleshooting Replication](https://redis.io/docs/management/replication/#faq)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

