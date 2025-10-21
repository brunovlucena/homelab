# 🚨 Runbook: PostgreSQL Replication Lag

## Alert Information

**Alert Name:** `PostgreSQLReplicationLag`  
**Severity:** Warning  
**Component:** postgres  
**Service:** postgres-postgresql  
**Threshold:** Replication lag > 10 seconds

## Symptom

PostgreSQL replica (standby) falling behind the primary server, causing stale reads and potential failover issues.

## Impact

- **User Impact:** MEDIUM - Stale data reads from replicas, potential inconsistency
- **Business Impact:** MEDIUM - Delayed reporting, slow replica queries
- **Data Impact:** LOW - Data eventually consistent, but lag can cause issues

## Important Note

⚠️ **Current Configuration:** The homelab PostgreSQL setup runs in **standalone mode** (single instance, no replication). This runbook is provided for **future reference** when replication is configured.

If you're seeing this alert, it may be misconfigured. Check if replication is actually enabled.

## Prerequisites for Replication

This runbook applies when PostgreSQL is configured with:
- Streaming replication (primary + replica)
- Logical replication
- PostgreSQL HA setup (Patroni, Stolon, etc.)

## Diagnosis

### 1. Check Replication Status

```bash
# On primary: Check replication connections
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT client_addr, state, 
   pg_wal_lsn_diff(pg_current_wal_lsn(), sent_lsn) AS pending_bytes,
   pg_wal_lsn_diff(pg_current_wal_lsn(), write_lsn) AS write_lag_bytes,
   pg_wal_lsn_diff(pg_current_wal_lsn(), flush_lsn) AS flush_lag_bytes,
   pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) AS replay_lag_bytes
   FROM pg_stat_replication;"

# Check replication lag in time
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT client_addr, state, 
   write_lag, flush_lag, replay_lag,
   sync_state, sync_priority
   FROM pg_stat_replication;"
```

### 2. Check on Replica

```bash
# On replica: Check if recovery is running
kubectl exec -n postgres postgres-postgresql-replica-0 -- psql -U postgres -c \
  "SELECT pg_is_in_recovery();"

# Check last received LSN
kubectl exec -n postgres postgres-postgresql-replica-0 -- psql -U postgres -c \
  "SELECT pg_last_wal_receive_lsn(), 
   pg_last_wal_replay_lsn(),
   pg_wal_lsn_diff(pg_last_wal_receive_lsn(), pg_last_wal_replay_lsn()) AS replay_lag_bytes;"

# Check replay lag time
kubectl exec -n postgres postgres-postgresql-replica-0 -- psql -U postgres -c \
  "SELECT now() - pg_last_xact_replay_timestamp() AS replication_lag;"
```

### 3. Check Network Connectivity

```bash
# Test network from replica to primary
kubectl exec -n postgres postgres-postgresql-replica-0 -- nc -zv postgres-postgresql-0.postgres-postgresql.postgres.svc.cluster.local 5432

# Check network latency
kubectl exec -n postgres postgres-postgresql-replica-0 -- ping -c 5 postgres-postgresql-0.postgres-postgresql.postgres.svc.cluster.local
```

### 4. Check Primary Load

```bash
# Check primary CPU and memory
kubectl top pod -n postgres postgres-postgresql-0

# Check active connections on primary
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT count(*), state FROM pg_stat_activity GROUP BY state;"

# Check WAL generation rate
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT 
   pg_current_wal_lsn(),
   pg_wal_lsn_diff(pg_current_wal_lsn(), '0/0') / 1024 / 1024 AS wal_mb;"
```

### 5. Check Replica Resources

```bash
# Check replica CPU and memory
kubectl top pod -n postgres postgres-postgresql-replica-0

# Check if replica is overwhelmed with queries
kubectl exec -n postgres postgres-postgresql-replica-0 -- psql -U postgres -c \
  "SELECT count(*), state FROM pg_stat_activity WHERE state = 'active' GROUP BY state;"

# Check I/O wait
kubectl exec -n postgres postgres-postgresql-replica-0 -- iostat -x 1 5
```

### 6. Check Replication Slots

```bash
# On primary: Check replication slot status
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT slot_name, slot_type, active, 
   pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn) AS retained_bytes,
   pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn)) AS retained_size
   FROM pg_replication_slots;"

# Check if WAL is accumulating
kubectl exec -n postgres postgres-postgresql-0 -- du -sh /var/lib/postgresql/data/pg_wal
```

## Resolution Steps

### Step 1: Immediate Actions

#### Check for Connection Issues

```bash
# Restart replica if connection lost
kubectl delete pod -n postgres postgres-postgresql-replica-0

# Wait for replica to reconnect
kubectl wait --for=condition=ready pod -n postgres postgres-postgresql-replica-0 --timeout=300s

# Verify replication resumed
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT client_addr, state FROM pg_stat_replication;"
```

### Step 2: Reduce Primary Load

```bash
# Identify and terminate expensive queries on primary
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pid, now() - query_start AS duration, query
   FROM pg_stat_activity
   WHERE state = 'active'
   AND now() - query_start > interval '5 minutes'
   ORDER BY duration DESC;"

# Terminate long-running queries
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT pg_terminate_backend(<pid>);"
```

### Step 3: Increase Replication Resources

#### Issue: Replica Underpowered
**Cause:** Replica cannot keep up with primary write rate  
**Fix:**
```bash
# Increase replica resources
kubectl edit statefulset -n postgres postgres-postgresql-replica

# Update:
# resources:
#   limits:
#     memory: "4Gi"  # Match or exceed primary
#     cpu: "2000m"
#   requests:
#     memory: "2Gi"
#     cpu: "1000m"

# Wait for rolling update
kubectl rollout status statefulset -n postgres postgres-postgresql-replica
```

### Step 4: Optimize Replication Configuration

```bash
# On primary: Tune WAL settings
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET max_wal_senders = 10;"  # Default: 10

kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET wal_keep_size = '1GB';"  # Keep more WAL for slow replicas

kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET max_replication_slots = 10;"

# Restart primary to apply
kubectl delete pod -n postgres postgres-postgresql-0
```

```bash
# On replica: Tune recovery settings
kubectl exec -n postgres postgres-postgresql-replica-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET hot_standby_feedback = on;"  # Prevent query conflicts

kubectl exec -n postgres postgres-postgresql-replica-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET max_standby_streaming_delay = '30s';"  # Allow queries to delay replay

# Restart replica
kubectl delete pod -n postgres postgres-postgresql-replica-0
```

### Step 5: Reduce Replica Query Load

#### Issue: Read Queries Blocking Replay
**Cause:** Long queries on replica blocking WAL replay  
**Fix:**
```bash
# Check for long queries on replica
kubectl exec -n postgres postgres-postgresql-replica-0 -- psql -U postgres -c \
  "SELECT pid, now() - query_start AS duration, query
   FROM pg_stat_activity
   WHERE state = 'active'
   ORDER BY duration DESC
   LIMIT 10;"

# Terminate blocking queries
kubectl exec -n postgres postgres-postgresql-replica-0 -- psql -U postgres -c \
  "SELECT pg_terminate_backend(<pid>);"

# Or route read queries to additional replicas
```

### Step 6: Check and Fix Replication Conflicts

```bash
# On replica: Check for replication conflicts
kubectl exec -n postgres postgres-postgresql-replica-0 -- psql -U postgres -c \
  "SELECT * FROM pg_stat_database_conflicts WHERE datname = 'bruno_site';"

# Types of conflicts:
# - confl_tablespace: Tablespace deleted on primary while in use on replica
# - confl_lock: Lock conflicts
# - confl_snapshot: Snapshot conflicts
# - confl_bufferpin: Buffer pin conflicts
# - confl_deadlock: Deadlock conflicts

# Increase delay to reduce conflicts
kubectl exec -n postgres postgres-postgresql-replica-0 -- psql -U postgres -c \
  "ALTER SYSTEM SET max_standby_streaming_delay = '60s';"
```

### Step 7: Rebuild Replica (Last Resort)

```bash
# If lag is too large or replica is corrupted

# Step 1: Stop replica
kubectl scale statefulset -n postgres postgres-postgresql-replica --replicas=0

# Step 2: Delete replica PVC
kubectl delete pvc -n postgres data-postgres-postgresql-replica-0

# Step 3: Recreate replica PVC
# (Will be auto-created when pod starts)

# Step 4: Start replica - it will re-sync from primary
kubectl scale statefulset -n postgres postgres-postgresql-replica --replicas=1

# Step 5: Monitor replication catch-up
watch -n 5 'kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -t -c "SELECT client_addr, replay_lag FROM pg_stat_replication;"'
```

## Verification

### 1. Check Replication Lag Decreased

```bash
# Should show lag < 10 seconds
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT client_addr, replay_lag FROM pg_stat_replication;"
```

### 2. Verify Replica is Catching Up

```bash
# Monitor LSN diff over time
watch -n 2 'kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -t -c "SELECT pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) FROM pg_stat_replication;"'

# Should show decreasing bytes
```

### 3. Check Replication State

```bash
# Should show 'streaming'
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT client_addr, state FROM pg_stat_replication;"
```

### 4. Test Read from Replica

```bash
# Test query on replica
kubectl exec -n postgres postgres-postgresql-replica-0 -- psql -U postgres -c \
  "SELECT count(*) FROM users;"

# Compare with primary
kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c \
  "SELECT count(*) FROM users;"

# Should be same or very close
```

### 5. Monitor Stability

```bash
# Watch replication over time
watch -n 10 'kubectl exec -n postgres postgres-postgresql-0 -- psql -U postgres -c "SELECT client_addr, state, replay_lag FROM pg_stat_replication;"'

# Should remain stable
```

## Prevention

### 1. Configure Streaming Replication Properly

```yaml
# Primary HelmRelease configuration
primary:
  configuration: |
    # Replication settings
    wal_level = replica
    max_wal_senders = 10
    max_replication_slots = 10
    wal_keep_size = 1GB
    hot_standby = on
    
    # Performance tuning
    synchronous_commit = off  # For better performance (risk of data loss)
    # OR
    synchronous_commit = local  # Wait for local flush only
    
    # WAL archiving (optional but recommended)
    archive_mode = on
    archive_command = 'cp %p /archive/%f'

# Replica configuration
replica:
  configuration: |
    hot_standby = on
    hot_standby_feedback = on
    max_standby_streaming_delay = 30s
    primary_conninfo = 'host=postgres-postgresql-0.postgres-postgresql user=replicator'
```

### 2. Set Up Monitoring

```yaml
# Prometheus alert rules
- alert: PostgreSQLReplicationLag
  expr: |
    pg_replication_lag_seconds > 10
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "PostgreSQL replication lag high"
    description: "Replica is {{ $value }}s behind primary"

- alert: PostgreSQLReplicationLagCritical
  expr: |
    pg_replication_lag_seconds > 60
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "PostgreSQL replication lag critical"

- alert: PostgreSQLReplicationBroken
  expr: |
    pg_replication_is_replica == 1 AND pg_replication_lag_seconds > 300
  labels:
    severity: critical
  annotations:
    summary: "PostgreSQL replication may be broken"
```

### 3. Size Replicas Appropriately

```yaml
# Replica should have same or more resources than primary
replica:
  resources:
    limits:
      memory: "4Gi"
      cpu: "2000m"
    requests:
      memory: "2Gi"
      cpu: "1000m"
```

### 4. Use Connection Pooling for Replica Reads

```yaml
# PgBouncer for replica reads
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pgbouncer-replica
  namespace: postgres
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: pgbouncer
        image: edoburu/pgbouncer:1.21.0
        env:
        - name: DATABASE_URL
          value: "postgres://postgres@postgres-postgresql-replica.postgres.svc.cluster.local:5432/bruno_site"
        - name: POOL_MODE
          value: "transaction"
        - name: MAX_CLIENT_CONN
          value: "1000"
        - name: DEFAULT_POOL_SIZE
          value: "25"
```

### 5. Implement Read-Write Splitting

```python
# Application code: Route reads to replica
from psycopg2 import pool

# Primary connection pool (for writes)
primary_pool = pool.SimpleConnectionPool(
    2, 10,
    host='postgres-postgresql.postgres.svc.cluster.local',
    port=5432,
    database='bruno_site',
    user='postgres'
)

# Replica connection pool (for reads)
replica_pool = pool.SimpleConnectionPool(
    2, 20,  # More connections for reads
    host='postgres-postgresql-replica.postgres.svc.cluster.local',
    port=5432,
    database='bruno_site',
    user='postgres'
)

def read_data(query):
    """Use replica for read queries"""
    conn = replica_pool.getconn()
    try:
        cur = conn.cursor()
        cur.execute(query)
        return cur.fetchall()
    finally:
        replica_pool.putconn(conn)

def write_data(query):
    """Use primary for write queries"""
    conn = primary_pool.getconn()
    try:
        cur = conn.cursor()
        cur.execute(query)
        conn.commit()
    finally:
        primary_pool.putconn(conn)
```

### 6. Regular Replication Health Checks

```bash
# Create monitoring CronJob
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-replication-check
  namespace: postgres
spec:
  schedule: "*/5 * * * *"  # Every 5 minutes
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: check
            image: postgres:16
            command:
            - /bin/sh
            - -c
            - |
              # Check replication status
              LAG=$(psql -h postgres-postgresql -U postgres -t -c \
                "SELECT EXTRACT(EPOCH FROM replay_lag) FROM pg_stat_replication;" | tr -d ' ')
              
              if [ -z "$LAG" ]; then
                echo "ERROR: No replication connection found"
                exit 1
              elif [ "$LAG" -gt 60 ]; then
                echo "WARNING: Replication lag ${LAG}s is too high"
                exit 1
              else
                echo "OK: Replication lag ${LAG}s"
              fi
          restartPolicy: OnFailure
```

## Setting Up Replication (For Future Reference)

### Enable Streaming Replication

```yaml
# Update HelmRelease for HA mode
architecture: replication
replication:
  enabled: true
  numSynchronousReplicas: 1
  synchronousCommit: "on"

primary:
  configuration: |
    wal_level = replica
    max_wal_senders = 10
    max_replication_slots = 10
    hot_standby = on

replica:
  replicaCount: 2
  resources:
    limits:
      memory: "4Gi"
      cpu: "2000m"
```

## Replication Monitoring Dashboard

```promql
# Useful Prometheus queries

# Replication lag in seconds
pg_replication_lag{instance="postgres-postgresql-0"}

# Bytes behind primary
pg_stat_replication_pg_wal_lsn_diff{application_name="replica"}

# Replication state (1 = streaming)
pg_replication_is_replica

# Number of active replicas
count(pg_stat_replication_state{state="streaming"})
```

## Related Alerts

- `PostgreSQLDown`
- `PostgreSQLHighMemory`
- `PostgreSQLStorageFull`

## Escalation

If replication lag persists:

1. ✅ Check if replica has sufficient resources
2. 📊 Analyze primary write workload patterns
3. 🔍 Review replica query load
4. 💾 Consider adding more replicas to distribute read load
5. 🔄 Evaluate network bandwidth between primary and replica
6. 📞 Contact database team for advanced troubleshooting
7. 🆘 Consider Patroni/Stolon for automatic failover management

## Additional Resources

- [PostgreSQL Streaming Replication](https://www.postgresql.org/docs/current/warm-standby.html)
- [Replication Configuration](https://www.postgresql.org/docs/current/runtime-config-replication.html)
- [pg_stat_replication View](https://www.postgresql.org/docs/current/monitoring-stats.html#MONITORING-PG-STAT-REPLICATION-VIEW)
- [Bitnami PostgreSQL HA Chart](https://github.com/bitnami/charts/tree/main/bitnami/postgresql-ha)

---

**Last Updated:** 2025-10-15  
**Version:** 1.0  
**Owner:** Infrastructure Team

